# Cartulary Normative Core 03: Workbook Interaction, Collaboration, and Workflows

## 1. Interaction model

**REQ-03-001**
Cartulary MUST be **grid first and forms second**.
Profiles: base
Verified by: AC-001, AC-002, AC-005, AC-043, AC-231

**REQ-03-002**
The primary interaction surface MUST be a workbook-like grid with:

- inline editing,
- keyboard navigation,
- paste,
- low-friction row creation,
- saved or system views over projections,
- a collapsible inspector for enrichment, relationships, history, and destructive actions.
Profiles: base
Verified by: AC-001, AC-002, AC-005, AC-043, AC-231

**REQ-03-003**
The implementation MUST preserve the spreadsheet mental model at the view layer while keeping source data relational and auditable underneath.
Profiles: base
Verified by: AC-001, AC-002, AC-005, AC-043, AC-231

**REQ-03-286**
Every workbook surface MUST render inside a shell-owned work area between the active surface view bar and the status strip. That work area MUST fill the available workbook shell block size. The primary grid and, when open, the inspector MUST fill that same work area while allowing grid content and inspector content to scroll independently. Workbook layout MUST NOT introduce document-level vertical scrolling. Work-area and inspector block geometry MUST be independent of rendered row count and MUST remain stable for zero, one, three, and many rendered rows, including empty, loading, error, and draft-row states. Implementations MUST NOT satisfy this requirement by inserting synthetic rows, calculating surface height from row count, applying fixed `100vh - Npx` offsets, or adding surface-specific minimum-height workarounds.
Profiles: base
Verified by: AC-005, AC-043, AC-231

## 2. Workbook surface

### 2.1 Built-in tabs

**REQ-03-004**
The workbook MUST expose these built-in tabs in the base profile:

- Timeline,
- Hosts,
- Identities,
- Evidence,
- Notes.

The Notes sheet is the built-in-sheet member of the tagged-variant family defined by Core 02 §10.4.4A.
Profiles: base
Verified by: AC-112, AC-116, AC-231, AC-410

### 2.2 System views

**REQ-03-005**
The workbook MUST support additional contract-backed system views, including indicator, compromise-assessment, task-request, decision, and party surfaces.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231, AC-277

**REQ-03-006**
The Indicators system view MUST surface canonical indicator rows and support pivots to source-bound observations and lifecycle history without leaving the workbook interaction model.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-03-007**
The Compromise Assessments system view MUST surface incident-scoped assessment rows and support pivots to the assessed host or identity and that subject's prior assessment history without leaving the workbook interaction model.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-03-008**
The Task Requests system view MUST surface `task_request` rows and support queue-oriented filtering and sorting by `status`, `owner_user_id`, `priority`, `task_kind`, `workstream`, `due_at`, `requester_party_text`, `blocked_reason`, `completed_at`, `external_ticket_ref`, and `updated_at` without leaving the workbook interaction model.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-03-009**
The Decisions system view MUST surface `decision` rows and support review-oriented filtering and sorting by `status`, `owner_user_id`, `decision_type`, and `decided_at` without leaving the workbook interaction model.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231

**REQ-03-010**
Structured coordination artifacts such as `comm_log`, `handoff`, `status_review`, and `lesson` MUST be available as workbook-native base-profile surfaces identified canonically by `cartulary.view.comm_log.v1`, `cartulary.view.handoff.v1`, `cartulary.view.status_review.v1`, and `cartulary.view.lesson.v1`. These surfaces MUST NOT require additional built-in tabs in the base profile.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231, AC-281, AC-282, AC-283, AC-284

**REQ-03-011**
Such surfaces MUST remain workbook surfaces rather than separate application modules. Their canonical public identity MUST be the `view_schema` form of `sheet_ref` using that surface's standardized `view_schema_id`. In the current profile, only standardized base-profile `view_schema_id` values and the explicitly standardized optional artifact-backed `view_schema_id` values are valid `sheet_ref.kind='view_schema'` workbook-surface identities. Pack-dependent framework overlays such as ATT&CK, D3FEND, and VERIS MUST NOT be exposed or referenced as workbook-native `sheet_ref.kind='view_schema'` targets in the current profile. A saved view over the same `view_schema_id` MAY exist as an additional workbook surface, but it is a distinct saved-view object and MUST NOT replace the canonical identity of the required base surface. Variant membership, durable-discriminator semantics, and the no-separate-hypothesis rule for the artifact-backed note/coordination/finding family remain owned by Core 02 §10.4.4A and §10.4.5.
For the authoritative cross-layer workbook-surface mapping, including `source_record_types`, canonical source discriminator or filter, `surface_status`, and `required_reference_pack_keys`, see Core 01 Table 7.4-A.
Profiles: base
Verified by: AC-078, AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-121, AC-122, AC-231, AC-281, AC-282, AC-283, AC-284, AC-410, AC-411

### 2.3 Saved views

**REQ-03-012**
A saved view MUST be an incident-bound workbook configuration over exactly one `view_schema_id`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-03-013**
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

`layout_json` carries only shared portable layout state. Selection, scroll position, focused cell, local popover state, open inspector state, active inspector panel, preview state, local inspector forms, rollback previews, merge plans, stale confirmations, and presence remain client-local and MUST NOT be persisted as part of a saved view.

Persisted saved-view `query_json` MUST use the same stable field-key grammar as workbook view queries. Omitted `sort` and `filters` members MUST persist as empty arrays. Inactive grouping MUST persist by omitting `query_json.group_by`; explicit JSON `null` for `group_by` is invalid. Omitted create-time `layout_json` and create-time `layout_json={}` MUST normalize to the schema-derived `cartulary.layout.v1` default before persistence.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-03-014**
A saved view created from another saved view MUST persist a normalized canonical copy of that source view's `view_schema_id`, `query_json`, and `layout_json`. After creation, the new saved view MUST NOT inherit runtime behavior from the source `saved_view_id`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-03-291**
Workbook inspector behavior MUST be derived from the active immutable `view_schema_id` inspector config. A saved view inherits inspector behavior from its immutable `view_schema_id`; switching between saved views over the same `view_schema_id` MUST NOT change inspector configuration, and switching to a saved view over another `view_schema_id` MUST select that other schema's config.

The inspector MUST be closed by default on workbook open, surface switch, saved-view switch, and hard refresh. Opening the inspector MUST be an explicit user action and MUST NOT be required for ordinary row creation, inline editing, paste, correction, or rough capture.

When the inspector is open and no saved row is selected, it MUST render the exact no-row state token `no_row_selected` and MUST NOT display details, confirmations, previews, merge plans, rollback plans, local forms, or workflow state from a prior row.

When the selected `record_id` changes, every inspector panel MUST retarget to the new row or clear before showing new row data. Pending destructive confirmations, rollback previews, merge plans, and workflow forms MUST invalidate on selected-row change, row-version change, incident closure, authorization loss, record deletion, or merge. Mutating inspector actions that refresh the same surface MUST preserve selected row, grid scroll, and focus continuity where the route response and current authorization still allow it.
Profiles: base
Verified by: AC-453

### 2.3A Inspector workflow interaction semantics

**REQ-03-292**
The workbook client MUST execute inspector feature groups using the deterministic algorithms in this subsection. These algorithms govern client-visible interaction only. Server-side authorization, route validation, mutation legality, and source-state changes remain owned by the route and domain owners.
Profiles: base
Verified by: AC-456, AC-457, AC-458

Algorithm: `open_inspector(origin, selected_record_id)`.

1. If the inspector is already open for the same selected `record_id`, preserve the active panel unless the invoking origin names another declared panel.
2. If no saved row is selected, open the inspector in `no_row_selected` state.
3. If a saved row is selected, bind the inspector subject to `(view_schema_id, record_id, row_version)`.
4. Opening the inspector MUST never be required before ordinary row creation, inline edit, paste, correction, or rough capture.

Algorithm: `retarget_inspector(next_record_id, next_row_version)`.

1. If `next_record_id` differs from the active inspector subject, synchronously invalidate local forms, previews, destructive confirmations, rollback previews, merge plans, supersede forms, and stale errors.
2. Clear panel content or show panel-local loading state before painting new-row data.
3. Fetch or derive the new panel data using the active `view_schema_id` config.
4. If no saved row remains selected, enter `no_row_selected`.

Algorithm: `execute_feature_group(feature_group_key, current_subject)`.

1. Verify that `feature_group_key` exists in the active `view_schema_id` config.
2. Verify that a saved row is selected unless the feature explicitly supports `no_row_selected`.
3. Verify that current local state does not contain a matching `disabled_when[]` token.
4. If `requires_confirmation=true`, show a row-bound confirmation that names the affected record or records by stable identifiers and relevant display labels.
5. Submit only the existing route family declared by `route_binding.owner`.
6. Do not infer routes, write targets, or permissions from labels, component names, menu text, row order, or grid coordinates.
7. Treat any server authorization or validation failure as authoritative even if the control appeared enabled.

Algorithm: `complete_mutating_action(result)`.

1. If the result refreshes the same row and the caller still has access, preserve selected row, grid scroll, and focus where possible.
2. If the result retargets to a survivor, replacement, or created record, select the server-returned target when it is visible under the current surface or pivot to the declared target surface when `success_result_behavior='surface_pivot'`.
3. If the selected row is deleted, merged away, no longer visible, or no longer authorized, enter `no_row_selected`.
4. Do not navigate away from the workbook shell unless the feature group is an explicit `surface_pivot`.

Algorithm: `invalidate_confirmation(reason)`.

Pending confirmations, rollback previews, merge plans, supersede forms, and workflow forms MUST invalidate on selected-row change, row-version change, incident closure, authorization loss, record deletion, record merge, hard refresh, and active `view_schema_id` change.

Algorithm: `surface_pivot(binding)`.

1. Construct target `sheet_ref` using stable `view_schema_id` or `saved_view_id`.
2. Construct target filters only from declared `field_key` seed bindings.
3. Execute navigation inside the same workbook shell.
4. Preserve browser session, incident context, and status strip.
5. Do not use visible labels, storage tables, row indexes, or grid vendor coordinates as pivot identity.

**REQ-03-015**
`owner_user_id` MUST be present for `private` and `shared` saved views. It MAY be null only for `system` saved views.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

**REQ-03-016**
`incident_id`, `saved_view_id`, and `view_schema_id` MUST be immutable after creation.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-152, AC-231

A `system` saved view is a saved-view configuration object with `scope='system'`. It is not the same object as a contract-backed system view identified by `view_schema_id`.

#### 2.3.1 Scope and discoverability

**REQ-03-017**
Saved-view scope MUST use exactly these three values:

- `private`,
- `shared`,
- `system`.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-231

**REQ-03-018**
`private` means the saved-view object is visible only to its owner and incident admins. Any incident member MUST be able to create a `private` saved view for their own use.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-231

**REQ-03-019**
`shared` means the saved-view object is visible to all incident members. It MUST still record one owner for accountability. The owner and incident admins MUST be able to update or delete it in place. Other incident members MAY open it and duplicate it, but MUST NOT update or delete it in place through the ordinary saved-view routes.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-231

**REQ-03-020**
`system` means an implementation-owned or admin-seeded saved-view configuration still bound to one incident. It MUST be visible to all incident members, but it MUST be immutable through the ordinary saved-view write path. Users MAY duplicate a visible `system` saved view, but they MUST NOT edit or delete it in place through the ordinary saved-view routes.
Profiles: base
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-231

**REQ-03-021**
Saved-view scope controls discoverability and mutability of the saved-view object only. It MUST NOT widen or narrow access to underlying incident rows, fields, search results, export redaction behavior, or evidence visibility.
Profiles: base, snapshot_reporting
Verified by: AC-146, AC-147, AC-148, AC-149, AC-151, AC-231, AC-233

#### 2.3.2 Ordinary lifecycle semantics

**REQ-03-022**
Ordinary saved-view listing MUST return only the saved views visible to the caller.
Profiles: base
Verified by: AC-152, AC-231

**REQ-03-023**
Ordinary saved-view creation MUST default `scope` to `private` when the request omits it. The ordinary public create path MUST reject `scope='system'`.
Profiles: base
Verified by: AC-152, AC-231

**REQ-03-024**
Ordinary saved-view update MUST allow mutation only of `display_name`, `query_json`, `layout_json`, and, when permitted by scope rules, `scope` between `private` and `shared`. It MUST reject attempted mutation of `incident_id`, `saved_view_id`, or `view_schema_id`. A structurally valid no-op update MUST return the current saved-view configuration without advancing `saved_view_version` or `updated_at`. Any materially changed successful in-place saved-view mutation MUST advance `saved_view_version` and refresh `updated_at`.
Profiles: base
Verified by: AC-152, AC-231

**REQ-03-025**
Ordinary saved-view delete MUST delete only the saved-view configuration object. It MUST NOT delete or mutate underlying incident rows, links, tags, evidence records, or canonical entities.
Profiles: base
Verified by: AC-152, AC-231

**REQ-03-026**
A user who can open a visible saved view MUST be able to duplicate it into a new saved view by persisting a normalized copy of its current `view_schema_id`, `query_json`, and `layout_json`.
Profiles: base
Verified by: AC-152, AC-231

### 2.4 Startup and default surface selection

**REQ-03-027**
Saved-view scope and startup/default surface selection MUST remain separate concerns.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-03-028**
The per-user startup pointer MUST be `user_workbook_preferences.home_sheet_ref`. The incident-wide fallback pointer MUST be `incident_workbook_preferences.default_sheet_ref`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-03-029**
Both pointers MUST use the stable `sheet_ref` shape defined by Core 01 §3.3.10.1. When the pointed surface is any pack-independent base-profile registry surface listed in Core 01 REQ-01-307, the stored `sheet_ref` MUST use `{ "kind": "view_schema", "id": <view_schema_id> }` for the required base surface itself; the `saved_view` form remains valid only for a distinct saved-view object over that schema.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-03-030**
Workbook open MUST select the starting surface in this order:

1. an explicit launch `sheet_ref`, if present and still valid for the caller,
2. the caller's `home_sheet_ref`, if present, still valid, and visible to the caller,
3. the incident-wide `default_sheet_ref`, if present, still valid, and visible to the caller,
4. `cartulary.view.timeline.v2`.
The startup selection resource defined by Core 01 MUST return both the selected `sheet_ref` identity and the base `view_schema_id` used for workbook queries so a selected saved view is not collapsed into the base surface identity.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-03-031**
If a persisted referenced saved view or view schema is missing, deleted, no longer visible to the caller, invalid because a required optional pack is unavailable, or invalid because the referenced `view_schema_id` is not standardized for the current profile, the implementation MUST clear the invalid pointer and continue to the next step in the ordered fallback chain rather than failing workbook open. Invalid explicit launch pointers are not persisted and therefore MUST be skipped without reporting a cleared pointer. This fallback logic MUST NOT depend on the existence of a saved-view object for any pack-independent base-profile registry surface listed in Core 01 REQ-01-307, because those surfaces remain directly addressable by standardized `view_schema_id`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

Startup request-validation failures and persisted-pointer clearing reasons MUST use the Core 01 REQ-01-151.1 reason-code registries. The current profile represents hard-deleted and never-existing saved-view references with `saved_view_not_found`; it does not expose a distinct deleted-vs-never-existed public state. Pack-unavailability fallback MUST use `required_reference_pack_unavailable` only when the addressed view contract declares unavailable `required_reference_pack_keys`.

For authenticated root landing flows that open a workbook without an explicit valid launch `sheet_ref`, Core 01 §3.3.2.1A reuses this same ordered fallback chain rather than defining a separate workbook-startup order.

**REQ-03-290**
Deployment administration is not a workbook startup surface. The `/deployment-administration` browser context defined by Core 01 §3.3.2.1B MUST NOT be represented as a `sheet_ref`, saved view, system view, built-in tab, `home_sheet_ref`, `default_sheet_ref`, startup fallback candidate, or workbook-surface registry entry.

When a successful incident-bundle import exposes an `Open imported incident` action, activating that action MUST open the imported incident without an explicit launch `sheet_ref` and MUST use the ordinary startup chain in REQ-03-030. The action MUST preserve the imported incident's actual lifecycle state; if the imported incident is `closed`, ordinary closed/read-only behavior applies after open.
Profiles: base, incident_portability
Verified by: AC-441, AC-442

**REQ-03-032**
Any incident member MUST be able to set or clear their own `home_sheet_ref`. Only incident admins MUST be able to set or clear `incident_workbook_preferences.default_sheet_ref`.
Profiles: base
Verified by: AC-150, AC-153, AC-231

**REQ-03-289**
Workbook rendering MUST apply the effective density computed by the account-preference contract in Core 01 §3.3.2.3. When the caller's `account_preferences.density_mode` is `null`, the Timeline surface uses `compact` and every other workbook surface uses `default`; when it is non-null, the exact token `compact`, `default`, or `comfortable` applies as the user's density override. Implementations MUST NOT invent custom density tokens, custom row-height persistence, or per-surface density overrides in the current profile.

Changing account density is a presentation preference only. It MUST NOT alter `view_schema` definitions, saved-view objects, saved-view `query_json`, saved-view `layout_json`, query request JSON, row data, row versions, collaboration ordering, `presence_snapshot` or `presence_delta` ordering, per-incident `user_workbook_preferences.home_sheet_ref`, `incident_workbook_preferences.default_sheet_ref`, workbook-startup selection, or incident portability content.
Profiles: base
Verified by: AC-431, AC-432

## 3. Collaboration and concurrency model

### 3.1 Concurrency strategy

**REQ-03-033**
The implementation MUST use **field-level optimistic concurrency on top of row versioning**.
Profiles: base
Verified by: AC-009, AC-013, AC-047, AC-231

**REQ-03-034**
Each visible row MUST be bound to the `record_id` and `row_version` emitted by its projection row.
Profiles: base
Verified by: AC-009, AC-013, AC-047, AC-231

**REQ-03-035**
Every grid write MUST include:

- `record_id`,
- the client’s `base_row_version`,
- changed fields only.

For client-driven dependent writes to the same record, including autosave-originated collection actions and non-grid record actions such as review, supersede, delete, restore, rollback, mention-resolution, and evidence-attachment actions, the client MUST derive `base_row_version` from the latest committed `row_version` it has accepted for that record from a prior successful mutation response, authoritative live row patch, or full row returned by the workbook query route. A same-record pending write MUST NOT be considered safe for dependent dispatch until that committed version has been accepted into the client's local committed-version source. If that latest committed version is not available at dispatch time, the dependent write MUST remain queued or the client MUST refresh before dispatching it. The client MUST NOT intentionally dispatch a dependent same-record write from an older rendered row snapshot when a newer committed row version is already known locally. For every active workbook surface, the client MUST treat that local committed-version source as a per-`record_id` high-water mark. An incoming full query row, row-refreshing mutation response, action response, or live sparse row patch whose `row_version` is lower than that high-water mark MUST NOT replace the rendered committed row state for that `record_id` or lower the version used for later optimistic writes.

For editable collection controls, an Enter or Tab commit and the blur event caused by that same keyboard commit are one logical collection commit. The client MUST NOT emit a second autosave mutation for the same control value solely because focus moved after the keyboard commit.
Profiles: base
Verified by: AC-009, AC-013, AC-047, AC-231

**REQ-03-283**
After a successful same-surface row mutation that refreshes or replaces the rendered row, the workbook client MUST preserve the selected `record_id`, restore the owned grid scroll position, and restore focus to a deterministic same-surface continuity target. If the original editor remains active and visible, that editor MAY be the continuity target. If the original editor is removed, hidden, inert, compacted behind an overflow presentation, or otherwise unsuitable for visible focus, the continuity target MUST be a visible row-local fallback such as the row action or inspector affordance for the same `record_id`. This requirement applies to autosave-originated collection actions, inline scalar edits, reviewer lifecycle actions, mention-resolution actions, and other row-local inspector actions that complete without leaving the current workbook surface.

When a row-local mutation requires follow-up same-surface refresh or render work before the workbook can show the mutation's canonical consequences, the mutation's continuity boundary includes that follow-up work. The client MUST NOT treat continuity as settled solely because the mutation response rendered. It MUST preserve the row-local continuity target until the required follow-up has rendered, or until that follow-up reaches a terminal failure state and the client restores a deterministic row-local fallback.
Profiles: base
Verified by: AC-005, AC-006, AC-188, AC-190, AC-205, AC-231

### 3.2 Server-side conflict behavior

**REQ-03-036**
If the current row version matches the supplied `base_row_version`, the patch MUST apply normally.
Profiles: base
Verified by: AC-009, AC-126, AC-231

**REQ-03-037**
If another user has changed the row:

- when the other change touched a different field, the server MUST auto-rebase and accept the write,
- when the other change touched the same field, the server MUST reject the write with a conflict payload.
Profiles: base
Verified by: AC-009, AC-126, AC-231

**REQ-03-038**
For `conflict_resolution_class='text_compare_merge'`, same-field detection MUST be keyed by `field_key`, not by textual subrange. The server MUST reject the original patch with `error.code='same_field_conflict'` whenever another committed write changed that same writable `field_key` after `base_row_version`, even if a deterministic clean text merge suggestion would be possible.
Profiles: base
Verified by: AC-009, AC-126, AC-231

**REQ-03-039**
At minimum, the conflict payload MUST preserve:

- the stable `field_key`,
- the client-submitted value,
- the persisted server value,
- the client `base_row_version`,
- the current `row_version`.
Profiles: base
Verified by: AC-009, AC-126, AC-231

**REQ-03-040**
A same-field conflict MUST remain unresolved until an analyst explicitly chooses a resolution. Same-field edits MUST NOT silently overwrite each other.
Profiles: base
Verified by: AC-009, AC-126, AC-231

### 3.3 Analyst-facing same-field conflict resolution

#### 3.3.1 Resolver surface and conflict state

**REQ-03-041**
When a same-field conflict occurs, only the affected cell MUST enter conflict state.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

**REQ-03-042**
The conflicted cell MUST continue to display the current saved server value plus a visible conflict marker. The client MUST retain the analyst's unsaved local value separately and MUST NOT render that local value as though it were already saved.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

**REQ-03-043**
The analyst-facing resolver MUST open from the conflicted cell and MUST keep the main grid visible while the resolver is open. The implementation MAY realize this as a compare drawer or another same-surface panel, provided the row containing the conflicted cell remains in view.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

**REQ-03-044**
Closing the resolver without selecting a resolution MUST leave the cell in conflict state.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

**REQ-03-045**
The workbook save-state presentation MUST remain `Conflict` until every unresolved same-field conflict local to that client has been cleared or resolved.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

When same-cell presence is available before the conflict occurs, the UI SHOULD show a hint equivalent to another analyst actively editing the field.

**REQ-03-046**
The analyst MUST be able to continue editing other rows or cells while a same-field conflict on a different cell remains unresolved.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

**REQ-03-047**
After an explicit same-field conflict resolution or clear action, focus MUST return to the same cell and scroll position MUST be preserved.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-042, AC-231

#### 3.3.2 Resolver contents and safety rules

**REQ-03-048**
The resolver MUST display:

- row context sufficient to identify the record without leaving the sheet,
- the field display label and stable `field_key`,
- the saved value plus the actor and timestamp of that saved value,
- the analyst's unsaved local value,
- for merge-capable fields, a diff summary against the common base,
- direct resolution actions in plain language.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

The resolver SHOULD open with a message equivalent to: `This field changed before your edit was saved. Review the saved value and your unsaved value.`

**REQ-03-049**
The resolver MUST NOT default focus to a destructive action such as `Use my unsaved value`.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-050**
Initial keyboard focus MUST land on the conflict summary or an equivalent non-destructive element.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-051**
Pressing Enter while the resolver first opens MUST NOT resolve the conflict.
Profiles: base
Verified by: AC-037, AC-038, AC-039, AC-040, AC-041, AC-042, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

#### 3.3.3 Contract-declared resolution classes

**REQ-03-052**
For each write-back-capable `field_key`, `view_schemas.writeback_contract` MUST declare `conflict_resolution_class`.
Profiles: base
Verified by: AC-118, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

The closed vocabulary is:

| `conflict_resolution_class` | Required use | Required resolver behavior |
| --- | --- | --- |
| `atomic_replace` | scalar fields such as timestamps, enums, numbers, single-value identifiers, and state fields | present explicit `Keep saved value` and `Use my unsaved value` actions |
| `text_compare_merge` | analyst-authored free text such as Timeline v2 visible operational text, note body, and description | present side-by-side comparison with change highlighting and an optional `Edit merged value` path |
| `collection_review` | multi-value chip or set fields such as tags, support refs, and hidden inspector-side Timeline mention actions | present base, saved, and local deltas plus a final preview before commit |

**REQ-03-053**
Unknown or omitted classes MUST behave as `atomic_replace`.
Profiles: base
Verified by: AC-118, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

#### 3.3.3.1 Operational semantics for `text_compare_merge`

**REQ-03-054**
`text_compare_merge` is a plain-text merge-capable conflict class. It MUST treat the field value as a scalar text value and MUST NOT parse or interpret Markdown structure, HTML structure, entity chips, links, or any other rendered representation.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-055**
For merge computation only:

- `null` MUST be treated as the empty string,
- line endings MUST be normalized from `CRLF` or `CR` to `LF`,
- all other code points MUST be preserved exactly.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-056**
The authoritative merge unit is the line. The server MAY compute a deterministic three-way merge suggestion from `base_value`, `server_value`, and `client_value`, but the presence or absence of a clean suggestion MUST NOT change same-field conflict detection under §3.2.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-057**
For this class, a changed line hunk is a maximal contiguous changed line range relative to the normalized `base_value`. A clean merge suggestion exists only when the client-side and server-side changed line hunks are disjoint in base-line coordinates and the two sides do not both insert at the same base boundary. When a clean suggestion exists, the server MUST materialize `suggested_merged_value` by applying the non-overlapping hunks to the normalized `base_value` in ascending base order.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-058**
Word-level or character-level highlighting MAY be used for presentation inside changed lines, but that highlighting is non-authoritative and MUST NOT control conflict detection or merge validity.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-059**
The server MUST NOT silently auto-commit, auto-accept, or auto-retry a same-field text edit. If the deterministic line merge is clean, the conflict object MAY include `suggested_merged_value`. If the merge detects overlapping changed line hunks, or competing insertions at the same base position, the conflict object MUST omit `suggested_merged_value`.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-060**
The merge editor MUST start from the current saved value and MUST keep the analyst's unsaved local draft visible as reference. If `suggested_merged_value` is present, the UI MAY offer `Apply suggested merge` or equivalent, but MUST NOT apply it implicitly.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

For short one-line fields such as `task.title`, this line-based rule means a clean suggestion will often be absent. In that case the analyst resolves the conflict explicitly using the saved value and local draft already shown in the resolver.

**REQ-03-061**
A later profile that needs structured Markdown, rich-text, or semantic entity-aware merge MUST define a new `conflict_resolution_class` rather than widening `text_compare_merge`.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-062**
For relationship cells that mix unresolved mention tokens and canonical entity links, the resolver MUST preserve those as different object types. It MUST NOT coerce unresolved tokens and canonical chips into plain delimited text solely for conflict presentation.
Profiles: base
Verified by: AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

#### 3.3.4 Same-field conflict transport contract

**REQ-03-063**
A same-field conflict response MUST use `409 Conflict` or an equivalent explicit concurrency-conflict status in deployments that do not expose HTTP directly.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-064**
A same-field conflict response MUST use the generic error envelope defined in Core 01 §3.3.6 rather than the normal transient retry path.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-065**
The response MUST set `error.code` to `same_field_conflict`, set `error.status` to `409`, and include an `error.conflict` object.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-066**
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
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-067**
When `conflict_resolution_class='text_compare_merge'`, the conflict object MUST include `base_value`; `base_revision_ref` alone is insufficient for the base profile. In that case `error.conflict.client_value`, `error.conflict.server_value`, and `error.conflict.base_value` MUST each be the raw text scalar for the field or `null`. The conflict object MAY additionally include `suggested_merged_value`, which, when present, MUST also be the raw text scalar for the field or `null`. The presence of `suggested_merged_value` means only that the server found a deterministic clean line merge suggestion and MUST NOT imply that the rejected write has been accepted.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-068**
When `resolution_kind='merged_value'` targets `conflict_resolution_class='text_compare_merge'`, `resolved_value` MUST be the final plain-text scalar for the field or `null`. `resolved_value` MUST NOT be a diff script, markup object, token list, AST, or field-specific merge action object. The server MUST persist that value through the field's ordinary write target, subject only to ordinary field validation and the same newline canonicalization used by ordinary writes.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-069**
When `conflict_resolution_class='collection_review'`, the conflict object MUST include `base_value`. In that case `error.conflict.client_value`, `error.conflict.server_value`, and `error.conflict.base_value` MUST each use `collection_value_v1`. The server MUST preserve distinct item kinds in those values and MUST NOT collapse unresolved mentions, resolved entity refs, tags, aliases, or linked-record references into raw string arrays or plain delimited text for conflict transport.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-070**
When `resolution_kind='merged_value'` targets `conflict_resolution_class='collection_review'`, `resolved_value` MUST use `collection_actions_v1` evaluated against the `server_value` collection carried in the same conflict object.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-071**
The client MUST place same-field conflicts into a client-local conflict queue keyed by the canonical composite `record_id:field_key`.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-072**
The client MUST keep this conflict queue separate from the transient local pending queue used for retryable transport failures.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231, AC-381

**REQ-03-073**
A same-field conflict MUST NOT auto-retry.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-074**
When exposed over the public HTTP surface, the explicit resolution request MUST use `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-075**
An explicit resolution request MUST include:

- `conflict_token`,
- `resolution_kind` with one of `keep_saved`, `use_unsaved`, or `merged_value`,
- `resolved_value` when required by the chosen `resolution_kind`.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

**REQ-03-076**
If the same field changes again before the analyst resolves the conflict, the server MUST reject the stale `conflict_token` and return a fresh same-field conflict payload in the same generic error envelope. The client MUST preserve the analyst's unsaved local draft and refresh the compare surface against the newest saved value.
Profiles: base
Verified by: AC-126, AC-203, AC-204, AC-226, AC-227, AC-228, AC-229, AC-230, AC-231

#### 3.3.5 Local draft, history, and analytics boundary

**REQ-03-077**
Unresolved same-field conflict drafts MUST be treated as client-local unsaved work rather than authoritative incident state.
Profiles: base
Verified by: AC-041, AC-163, AC-231

**REQ-03-078**
Until explicitly committed, they:

- MUST NOT be broadcast to other analysts,
- MUST NOT appear in history, search, exports, or snapshots,
- MUST remain memory-local in the base profile.
Profiles: base, snapshot_reporting
Verified by: AC-041, AC-163, AC-231, AC-233

**REQ-03-079**
If a deployment later adds durable local draft persistence, that persistence MUST be explicitly configured and session-scoped.
Profiles: base
Verified by: AC-041, AC-163, AC-231

**REQ-03-080**
Selecting `Use my unsaved value` or `Edit merged value` MUST create a new attributed `change_set` and a new row-field mutation against the current row version.
Profiles: base
Verified by: AC-041, AC-163, AC-231

**REQ-03-081**
Selecting `Keep saved value` MUST clear the local conflict without creating a new source revision.
Profiles: base
Verified by: AC-041, AC-163, AC-231

**REQ-03-082**
The implementation MAY record same-field conflict analytics, but it MUST NOT persist discarded field content solely for analytics. Conflict analytics SHOULD be limited to field identifier, timing, actors, and chosen outcome.
Profiles: base
Verified by: AC-041, AC-163, AC-231

#### 3.3.6 Paste-time same-field conflicts

**REQ-03-083**
During multi-cell paste:

- non-conflicting cells MUST commit normally,
- same-field conflicts MUST move into the same-field conflict queue,
- conflict entries MUST be ordered by source row ordinal and then source column ordinal for the paste operation,
- the UI MUST present one grouped same-surface conflict navigator for the paste operation rather than one blocking modal per conflicted cell. The implementation MAY realize this as a grouped conflict tray or another same-surface grouped navigator.
Profiles: base
Verified by: AC-040, AC-231

**REQ-03-084**
When a paste produces at least one same-field conflict, the committed non-conflicting portion of the paste MUST remain one visible `change_set`. If a valid paste produces only same-field conflicts, the batch result MUST contain an empty `rows[]` and no `change_set_id`. Each later per-cell conflict resolution MUST create its own attributed `change_set`.
Profiles: base
Verified by: AC-040, AC-231

**REQ-03-085**
The base profile MUST NOT offer a blanket action such as `Use my value for all conflicts`.
Profiles: base
Verified by: AC-040, AC-231

### 3.4 Client addressing rules

**REQ-03-086**
The client MUST NOT address or mutate a row by:

- visible row number,
- sort position,
- filtered position,
- grouped position,
- displayed cell values.
Profiles: base
Verified by: AC-013, AC-047, AC-125, AC-231

## 4. Save behavior and presence

### 4.1 Autosave

**REQ-03-087**
The grid MUST autosave on:

- Enter,
- Tab,
- blur,
- paste completion.
Profiles: base
Verified by: AC-043, AC-231

**REQ-03-088**
The normal workflow MUST NOT require an explicit Save button.
Profiles: base
Verified by: AC-043, AC-231

### 4.2 Save-state presentation

**REQ-03-089**
The UI MUST present a compact save state using exactly these user-visible labels:

- `Syncing`,
- `Saved`,
- `Conflict`.

For the base profile, the label mapping is:

- `Syncing`: at least one workbook mutation is in flight or the local pending queue is non-empty, including while replay is paused waiting for connectivity recovery, re-authentication, or an HTTP re-query required by `REQ-03-096`.
- `Saved`: no workbook mutation is in flight, the local pending queue is empty, and the client has no unresolved same-field local drafts for that workbook state.
- `Conflict`: at least one unresolved same-field conflict exists, or queue overflow has refused admission of a new replay unit, or replay is halted on a non-retryable failure that requires analyst action.

When the save-state label is `Conflict` because unresolved same-field local drafts exist, the status strip MUST render exactly one primary label, `Conflict`, and MAY render one user-facing same-surface secondary message that summarizes the affected conflict count. That ordinary visible status-strip summary MUST NOT use `record_id`, `field_key`, `conflict_token`, route names, or raw public error text as its primary copy. The client MAY retain those anchors for resolver routing and MAY expose them as technical metadata in an appropriate secondary detail surface.

Ambient collaboration state MUST NOT change this label mapping.
Profiles: base
Verified by: AC-043, AC-231, AC-376

**REQ-03-282**
When a workbook-dispatched record action requires `base_row_version`, the client MUST NOT dispatch that action while same-record local workbook mutation work is queued, in flight, or blocked by a refresh boundary. The client MUST first wait until the row has a known latest committed `row_version` for that record, then use that committed version as the action's `base_row_version`. If the pending queue is halted, authentication-paused, overflowing, or blocked by an unresolved same-field conflict, the client MUST fail the action locally through the workbook conflict state rather than guessing a stale base version.
Profiles: base
Verified by: AC-043, AC-125, AC-126, AC-181, AC-183, AC-200, AC-231

### 4.3 Presence

**REQ-03-090**
The UI MUST provide all of the following presence indicators:

- workbook-header presence avatars for users on the same sheet,
- row-gutter indicators when another analyst is focused on a row,
- same-cell indicators when another analyst is actively editing the same field and such a signal is available.
Profiles: base
Verified by: AC-008, AC-132, AC-231

**REQ-03-091**
Presence and live row updates MUST be driven by the bounded WebSocket message families defined in Core 01 §3.3.10 rather than by a second mutation API.
Profiles: base
Verified by: AC-008, AC-132, AC-231

#### 4.3.1 Collaboration message application

**REQ-03-092**
The client MUST include its initial workbook presence in `hello` or `resume` and MUST send `presence_update` whenever any of the following changes:

- the active workbook surface changes to a different `sheet_ref`,
- the focused `record_id` changes,
- same-cell editing starts or stops for a writable `field_key`,
- the client becomes `idle` or returns from `idle`.

When the active workbook surface is any pack-independent base-profile registry surface listed in Core 01 REQ-01-307 and the user is on the required base surface itself rather than a distinct saved-view object over the same schema, the transmitted `sheet_ref` MUST use `kind = view_schema` with the standardized `view_schema_id`; opening a distinct saved view over that schema MUST instead transmit `kind = saved_view` with that saved view's `saved_view_id`.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231

**REQ-03-093**
The client MAY coalesce rapid local cursor motion, but the server-visible presence state for any stable user-visible change MUST settle within 1 second.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231

**REQ-03-094**
Workbook-header presence avatars MUST be derived from `presence_snapshot` and `presence_delta` records whose `sheet_ref` exactly matches the active workbook surface. Row-gutter indicators MUST be derived from matching `record_id`. Same-cell indicators MUST be derived from matching `record_id` plus `field_key` with `mode = editing`. The client MUST key these indicators from `sheet_ref`, `record_id`, and `field_key`; it MUST NOT infer collaboration state from visible tab labels, row numbers, or column headers. The client MUST treat `presence_snapshot.payload.presences[]` as a keyed collection by exact `connection_id` and MUST NOT infer recency, tie-break, or presentation order from array position. The client MUST also preserve the distinction between a direct base coordination surface addressed as `sheet_ref.kind="view_schema"` and a distinct saved view over the same schema addressed as `sheet_ref.kind="saved_view"`. Any avatar order shown in the UI MAY use a separate deterministic local presentation rule, but that presentation order is non-authoritative and MUST NOT be fed back into diffing, cache keys, or resume-state checks.
Profiles: base, snapshot_reporting
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231, AC-233

**REQ-03-095**
Presence updates are ambient last-write-wins state. They MUST NOT change the local save state, conflict queue, or local pending queue.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231, AC-376

**REQ-03-096**
When a replayable `record_changed` message arrives, the client MUST de-duplicate it by `(incident_id, stream_seq)`. If the client detects a gap in replayable `stream_seq` or receives `resume_ack.status = reset_required`, it MUST stop incremental apply and re-query the current view through the HTTP query route before presenting the sheet as synchronized again. A replayed or live `record_changed` message whose `row_version` is older than the client's accepted high-water mark for the same `record_id` MUST be treated as stale and MUST NOT trigger a row regression or a compensating refresh solely for that stale message.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231

**REQ-03-097**
When the active surface's underlying `view_schema_id` appears in `record_changed.payload.affected_views[]`, the client MUST locate the relevant entry by exact `view_schema_id` rather than by array position, MUST treat `changed_field_keys[]` as a canonical set of exact `field_key` identifiers rather than mutation order, and MUST apply the matching entry as follows:

- `patch`: treat `patch_cells` as `view_row_patch_v1`; update only the `patch_cells.cells` members present in the event, apply only the `patch_cells.group_values` members present in the event, replace the row's `row_version` with the authoritative value from `patch_cells.row_version` only when that value is not older than the client's accepted high-water mark for the same `record_id`, treat an included cell `{ "value": null }` as authoritative null when that field contract admits null, treat omitted cells and omitted grouping scalars as unchanged, preserve selection and edit anchoring by `record_id`, and clear any matching local pending-queue entry only when the authoritative event covers that row and row version or a later committed row version,
- `invalidate`: mark the row or affected visible block dirty and refresh it through the existing HTTP view-query route rather than inventing a separate WebSocket read path or synthesizing a sparse patch the server did not send,
- `remove`: remove the row from the current materialized grid or mark it absent on the next synchronized query, and if the removed row was selected, clear the selection or move it according to the normal row-removal behavior without silently rebinding the old selection to a different `record_id`.

The client MAY use `changed_field_keys[]` for cache invalidation or reconciliation hints, but MUST NOT attach semantics to the order of those entries.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231, AC-368

**REQ-03-098**
A client that originated the mutation MUST reconcile against the same echoed `record_changed` message family as every other subscriber. Incoming collaboration messages MUST NOT surface unresolved same-field local drafts as saved state or overwrite the client-local conflict queue defined in §3.3.4 and §3.3.5. When a terminal polled job resource or a terminal `job_progress` message includes `result_summary.resource_refs[]`, the client MUST treat those refs as a compact navigation summary rather than a deep result payload. The client MUST surface known current-profile `kind` values as non-modal result chips or links on the current surface, MUST degrade unknown `kind` values to `result_summary.message`-only rendering without failing job rendering, MUST treat `route` as an opaque same-origin path rather than a UI-local route or preview/download handle, and MUST NOT auto-follow `route` or automatically change the active workbook surface, selection, or scroll position when the terminal result arrives. If the client already has a stronger local navigation affordance for a known `kind`, it MAY use that affordance, but it MUST NOT require UI-route strings inside the job resource.
Profiles: base
Verified by: AC-129, AC-131, AC-132, AC-133, AC-134, AC-135, AC-136, AC-231

#### 4.3.2 Closed incident workbook mode

**REQ-03-287**
When the incident resource exposes `status='closed'`, the workbook MUST remain visible and readable to current incident members, render a persistent incident-lifecycle state labeled exactly `Closed, read-only`, and treat that lifecycle state as separate from the compact save-state labels defined in §4.2. The client MUST disable or hide authoritative source-state write affordances, including row creation, cell editing, destructive actions, same-field conflict resolution, mention resolution, evidence attachment, and incident metadata patch controls. The client MUST retain read affordances and allowed derived operations, including incident list/get navigation, workbook queries, record history, evidence preview/download, saved views, workbook preferences, and extension-authorized snapshot, report, release, or incident-export actions.
Profiles: base, snapshot_reporting, incident_portability
Verified by: AC-424

**REQ-03-288**
When an in-flight or queued workbook source-state mutation receives `incident_closed` from HTTP replay or the collaboration stream terminates with `error.payload.code = incident_closed`, the client MUST treat the condition as terminal for that queued source-state work, not as a retryable transport outage. Queued or unsent source-state mutations for the closed incident MUST become non-authoritative rejected drafts that MAY remain locally visible and copyable, but they MUST NOT be auto-replayed while the incident is closed and MUST NOT auto-replay after a later successful reopen. After reopen, committing any retained draft requires a fresh user action that dispatches a new current-profile mutation request subject to ordinary authorization, lifecycle, version, and conflict checks.
Profiles: base
Verified by: AC-424, AC-426

**REQ-03-275**
When the client continues a pageable workbook surface with `cursor_token`, it MUST treat every page obtained from that cursor chain as an ordered live-authorized continuation, not as an immutable content snapshot. The client MUST NOT parse, alter, or manufacture cursor tokens. On explicit refresh, or on `400` with `error.code='invalid_view_query'` and a pagination `reason_code`, the client MUST discard the old chain and restart from page 1 without `cursor_token`. Where practical, that restart SHOULD preserve the same workbook surface, focus target, and viewport anchor rather than forcing full-page navigation.
Profiles: base
Verified by: AC-231

### 4.4 Local pending queue

**REQ-03-099**
The client MUST maintain a local pending queue for autosave-originated workbook hot-path mutations so that transient network interruptions do not lose typed data. A replay unit is exactly one autosave-originated workbook hot-path mutation: one row-create intent or one row-patch intent.

In the base profile, the local pending queue is client-memory-local, incident-scoped, and client-instance-scoped by `(incident_id, client_instance_id)`. It MUST NOT be shared across tabs or client instances.

In the base profile, the local pending queue MUST survive:

- transient transport failure,
- an HTTP auth failure on a queued write,
- `session_revoked` within the same browser runtime.

In the base profile, the local pending queue MUST NOT be relied on to survive:

- full page reload,
- tab close,
- browser restart,
- cross-tab transfer,
- tab crash.

The client MUST support a capacity of exactly `64` replay units per `(incident_id, client_instance_id)`. Replay MUST be FIFO by original enqueue order. The client MUST NOT reorder queued writes by visible row order, sort order, record type, or other presentation-derived state.

Coalescing is allowed only as follows:

- For a still-uncommitted local row, one queued create plus later unsent edits to that same local row MUST fold into one queued create unit until the first authoritative create succeeds.
- For an existing authoritative row, unsent patch units for the same `record_id` MAY coalesce only within one contiguous same-record run in the queue. The coalesced unit MUST preserve the final direct-write value for each `field_key` and the declared order of any `collection_actions_v1.actions[]`.

The client MUST NOT coalesce:

- across record boundaries,
- across intervening queued units for a different record,
- across destructive actions,
- across conflict-resolution actions,
- across non-hot-path operations.

When admitting a new replay unit would exceed capacity, the client MUST:

- keep every already queued unit,
- refuse admission of the new replay unit,
- preserve the current visible edit as unsaved local work,
- set save state to `Conflict`,
- show a same-surface non-modal overflow message.

On overflow the client MUST NOT silently evict oldest or newest queued units, and MUST NOT reorder queued units to make room.

When the platform permits, the client SHOULD warn before unload while the local pending queue is non-empty or unresolved same-field local drafts exist.
Profiles: base
Verified by: AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-376, AC-377, AC-378, AC-379, AC-380, AC-381, AC-382

**REQ-03-100**
An authentication failure on a queued write, or a `session_revoked` event on the collaboration stream, MUST NOT discard unresolved same-field local drafts or queued unsent writes. This requirement applies when `session_revoked` is caused by self-service password change, self-service TOTP replacement, administrator password reset, administrator TOTP reset, or explicit session revoke-all. The client MUST preserve that client-local unsaved work and prompt for re-authentication when required.

Before replay begins, the client MUST:

- establish a new authenticated session when required,
- re-derive current incident authorization,
- complete any HTTP re-query required by `REQ-03-096`.

Replay MUST proceed in FIFO order.

Replay MUST stop at the first non-retryable failure that requires analyst action. If that blocking failure is a same-field conflict, the blocked replay unit MUST leave the local pending queue, MUST enter the existing client-local same-field conflict queue keyed by `record_id:field_key`, and later queued units MUST remain queued behind it without being applied out of order. If the blocking failure is another terminal failure, later queued units MUST remain queued, save state MUST remain `Conflict`, and the blocking failure MUST be surfaced on the same workbook surface.

Replayed writes MUST still satisfy ordinary `base_row_version`, authorization, and same-field conflict checks before they become authoritative incident state. A replayed row-patch unit MUST materialize its `base_row_version` from the latest committed row version known at dispatch time rather than from a stale version captured when the unit was admitted. No additional workbook tab, saved view, or inspector workflow is required for this credential-lifecycle behavior.
Profiles: base
Verified by: AC-156, AC-157, AC-158, AC-159, AC-160, AC-161, AC-162, AC-163, AC-231, AC-376, AC-377, AC-378, AC-379, AC-380, AC-381, AC-382

## 5. Locking policy

**REQ-03-101**
Routine inline edits MUST NOT use hard record locks. Any workbook-surface lock-backed contention behavior MUST follow the destructive-operation concurrency contract owned by Core 01 §3.3.5.0. This section defines only workbook-surface consequences of that owner contract.
Profiles: base
Verified by: AC-182, AC-187, AC-218, AC-231, AC-353

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

**REQ-03-102**
A newly created Timeline row MUST persist with `capture_state='rough'`. This applies to blank-row creation, paste-created rows, and rows created with only an attached screenshot.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

**REQ-03-103**
The first committed `capture-state-material` mutation to a row whose current `capture_state` is `rough` MUST set `capture_state='enriched'` in the same committed `change_set`.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

**REQ-03-104**
The explicit review action MUST:

- be exposed only as a non-grid action from the inspector, history surface, or another reviewer-only surface that preserves the same non-grid invocation semantics,
- require current incident role `reviewer` or `admin`,
- be available only when the current `capture_state` is `rough` or `enriched`,
- set `capture_state='reviewed'` in the same committed `change_set`,
- create an attributed change that is visible through ordinary row history.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

**REQ-03-105**
Any committed `capture-state-material` mutation to a row whose current `capture_state` is `reviewed` MUST set `capture_state='enriched'` in the same committed `change_set`. A tag-only or other non-material change MUST leave `capture_state='reviewed'` unchanged.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

**REQ-03-106**
The explicit supersede action MUST:

- be exposed only as a non-grid action from the inspector, history surface, or another reviewer-only surface that preserves the same non-grid invocation semantics,
- MAY accept an optional replacement-row selection from that same inspector, history surface, or equivalent reviewer-only surface,
- require current incident role `reviewer` or `admin`,
- require a non-empty reason captured in the resulting attributed change,
- be available only when the current `capture_state` is `rough`, `enriched`, or `reviewed`,
- when a replacement row is selected and the action succeeds, surface that replacement nearby in the resulting inspector or equivalent reviewer surface,
- set `capture_state='superseded'` in the same committed `change_set`,
- when a replacement row is selected, persist the authoritative replacement relation in that same committed `change_set`,
- create an attributed change that is visible through ordinary row history, including both the `capture_state` transition and the replacement-link unit when present.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231, AC-329, AC-331

**REQ-03-107**
`superseded` is terminal for ordinary Timeline workflow. Ordinary Timeline patch, enrichment, and review actions MUST NOT mutate a row whose current `capture_state` is `superseded`. Leaving `superseded` MUST require reviewer rollback of the superseding change through row history. If the wrong replacement row was chosen, correction MUST require rollback of that superseding change and a new supersede action. The base profile defines no ordinary direct transition out of `superseded` and no ordinary direct edit path for a hidden Timeline replacement reference on a superseded row.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231, AC-331

**REQ-03-108**
`linked` is a derived milestone meaning the record has acquired one or more typed links, resolved mentions, evidence associations, or other typed relational structure. It MUST NOT be stored as a separate `capture_state` value in the current profile.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

`rolled back` is a history or reviewer action outcome, not a persisted `capture_state` value.

**REQ-03-109**
The implementation MUST NOT re-derive `capture_state` from current link counts, evidence counts, tags, or current unresolved-mention state alone. Removing all current links or evidence from an already `enriched` row MUST NOT revert it to `rough`. `timeline.has_unresolved_mentions` remains a separate derived field from current unresolved mention rows and MUST NOT override `capture_state`.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

**REQ-03-110**
Normalization MUST add structure without erasing the original observed input.
Profiles: base
Verified by: AC-107, AC-108, AC-109, AC-110, AC-111, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-191, AC-192, AC-193, AC-194, AC-195, AC-196, AC-197, AC-198, AC-199, AC-231

## 7. Timeline creation workflow

**REQ-03-111**
Typing into the blank trailing timeline row MUST create a real record as soon as one non-empty user-entered value exists.
Profiles: base
Verified by: AC-001, AC-002, AC-125, AC-191, AC-193, AC-231

**REQ-03-112**
A timeline row MUST be persistable with:

- only one non-empty user-entered value, or
- only an attached screenshot.
Profiles: base
Verified by: AC-001, AC-002, AC-125, AC-191, AC-193, AC-231

**REQ-03-113**
At row creation time, the system MUST support:

- nullable `occurred_at`,
- summary text,
- raw mention tokens for host and identity references,
- raw indicator-bearing text in summary, details, source text, or other contract-declared fields without requiring dedicated IOC columns.
Profiles: base
Verified by: AC-001, AC-002, AC-125, AC-191, AC-193, AC-231

**REQ-03-114**
The system MUST NOT block row creation on missing canonical host, identity, or indicator records.
Profiles: base
Verified by: AC-001, AC-002, AC-125, AC-191, AC-193, AC-231

**REQ-03-115**
Indicator capture from supported source fields is enrichment. Row creation MUST preserve raw field text and MUST NOT require immediate indicator extraction, resolution, or canonical indicator selection.
Profiles: base
Verified by: AC-001, AC-002, AC-125, AC-191, AC-193, AC-231

## 8. Evidence attachment workflow

### 8.1 Two-step upload

**REQ-03-116**
Binary evidence attachment MUST use a two-step flow:

1. create a pending blob slot for one intended upload by sending `incident_id`, `client_txn_id`, `byte_size`, and optional `filename_hint`, `content_type_hint`, and `sha256_hex`, and receive an opaque upload target plus `accepted_contract`,
2. finalize the evidence attachment after the upload completes.
Profiles: base
Verified by: AC-004, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-03-117**
When exposed over the public HTTP surface, step 1 MUST use `POST /api/v1/object-blobs`. In the base profile, the returned `upload_target.href` MUST be an opaque same-origin `PUT /api/v1/object-uploads/{upload_token}` capability; clients MUST treat it as an opaque URL, MUST honor returned upload target `method` and `headers`, and MUST NOT parse, replace, persist, or infer bucket names, storage keys, presigned-object-store query parameters, or object-store hostnames from it. Step 2 MUST use `POST /api/v1/evidence-records/{record_id}/attach-blob` or the normal record-creation path that binds the returned `object_blob_id` during evidence-row creation. The client MUST treat returned `accepted_contract` as the server-approved upload contract and MUST NOT reconstruct that contract from stale local state after an uncertain network boundary. When step 2 targets an existing evidence record through `POST /api/v1/evidence-records/{record_id}/attach-blob`, the client MUST send `object_blob_id`, `base_row_version`, and `client_txn_id`, MUST treat the action as record-scoped optimistic finalization rather than as a blind blob mutation, and, if it cannot tell whether attach succeeded, MUST replay the same normalized attach request with the same `client_txn_id`.

A successful selected-row evidence attachment or screenshot-only Timeline create MUST keep the active Timeline workbook surface mounted. The refreshed Timeline row MUST render through ordinary workbook mutation application, replay, or refresh behavior, including projection-backed evidence fields such as `timeline.evidence_count` and `timeline.has_evidence`; clients MUST NOT satisfy this workflow by navigating away from the workbook surface or by replacing it with an evidence-only surface.
Profiles: base
Verified by: AC-004, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-03-118**
If the client does not know whether blob-slot creation succeeded, it MUST replay the same normalized create request with the same `client_txn_id`. In the base profile, the upload target expires 60 minutes after issuance and the pending blob slot expires 24 hours after creation. The pending slot is a single-upload lease. If replay returns the original slot after target expiry, the client MUST request a fresh slot with a new `client_txn_id` rather than attempt same-slot upload-target refresh. The base profile MUST NOT require same-slot upload-target refresh or resumable upload semantics.
Profiles: base
Verified by: AC-004, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

**REQ-03-119**
The system MUST NOT leave fake attached evidence rows for incomplete uploads. The UI MUST surface expired-slot replay and terminal upload-contract failure as states that require a fresh slot rather than silently attaching evidence anyway.
Profiles: base
Verified by: AC-004, AC-102, AC-103, AC-128, AC-154, AC-155, AC-231

### 8.2 Pending evidence without blob

**REQ-03-120**
The evidence model MUST also support requested or pending evidence records with no blob yet attached, followed later by receipt, availability, custody, and optional blob attachment.
Profiles: base
Verified by: AC-015, AC-102, AC-154, AC-155, AC-231

### 8.3 Blob and evidence lifecycle bridge

Cartulary defines two linked but separate lifecycle machines for evidence attachment:

- blob upload, authoritative on `object_blobs.upload_state`,
- evidence custody and availability, authoritative on `evidence_records.lifecycle_state` plus custody events.

The blob-upload machine uses the exact `object_blobs.upload_state` tokens `pending`, `available`, `failed`, and `quarantined`.

The evidence machine uses the exact `evidence_records.lifecycle_state` tokens `requested`, `pending_receipt`, `received`, `available`, `quarantined`, and `released`.

Core 02 §13 owns the lifecycle state sets, legal transitions, bridge-derived outcomes, quarantine-entry semantics, and recovery rules for these machines. This subsection owns workbook-surface consequences only.

**REQ-03-121**
The required workbook-surface bridge behavior is:

- a `pending` blob slot MUST NOT by itself create or imply an attached evidence row,
- an evidence record whose linked blob or evidence lifecycle state makes preview or download unavailable under Core 02 §13 MUST render inline as blocked rather than navigating away,
- if Core 02 §13.1 requires a linked-blob quarantine to drive the evidence row to `lifecycle_state='quarantined'`, the workbook surface MUST show that exact state rather than a paraphrased non-available fallback,
- requested or pending evidence with no blob remains valid and MUST continue to support later receipt, custody, and optional blob linkage.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

**REQ-03-122**
Abandoned pending uploads and terminal upload-contract mismatches MUST fail closed. A blob slot left in `pending` without successful finalization, or failed because declared size or expected SHA-256 did not match observed bytes, MUST NOT increment row evidence counts, MUST NOT show as attached, and MUST remain eligible only for retry, timeout handling, or administrative cleanup. A content-type mismatch against advisory `content_type_hint` alone MUST update observed metadata and preview policy and MUST NOT by itself fail the slot. Filename mismatch alone MUST NOT be a terminal finalization failure in the base profile.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

**REQ-03-123**
A pending blob slot that reaches `pending_expires_at` without successful finalization MUST transition to `upload_state='failed'` with `terminal_reason='pending_timeout'`. Timeout MUST NOT create a distinct blob lifecycle state.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

**REQ-03-124**
The implementation MUST allow 3 failed explicit finalization attempts per pending blob slot for non-terminal explicit finalization failures. Size or expected-hash mismatch is terminal on first detection and MUST NOT consume or extend this retry budget. Only other unsuccessful explicit finalization attempts count toward the limit, and an idempotent replay after already-committed success MUST NOT consume retry budget. On the 4th such failed finalization attempt, the slot MUST transition to `upload_state='failed'` with `terminal_reason='finalize_retry_exhausted'`. Later retry then requires a fresh blob slot.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

**REQ-03-125**
Pending-blob timeout handling and orphaned unattached-blob cleanup MUST run as background cleanup work at least every 15 minutes. Any blob slot still in `pending` after `pending_expires_at` MUST be marked `failed` by the next cleanup sweep. Orphaned object bytes for a failed unattached blob slot MUST be deleted within 1 hour of terminal failure. Failed unattached slot metadata MUST remain queryable for at least 7 days, after which it MAY be hard-deleted automatically or by administrative action.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

**REQ-03-126**
If the system detects an inconsistent blob-versus-evidence state or a create-contract-versus-observed-byte mismatch state that has not yet been repaired, it MUST block preview and download and surface the row as inconsistent until explicit repair or re-finalization completes.
Profiles: base
Verified by: AC-015, AC-016, AC-103, AC-107, AC-108, AC-109, AC-110, AC-111, AC-128, AC-154, AC-155, AC-231, AC-313

### 8.4 Evidence access

Core 01 §16 owns the evidence-access handle wire contract. This subsection owns workbook-surface invocation, layout, and blocked-state behavior only.

**REQ-03-127**
Evidence preview MUST open without forcing a full-page navigation away from the grid. A bottom or side preview is acceptable. When preview is unavailable because the server returns `unsupported_preview` or another blocked evidence-access state, the grid or inspector MUST remain in place and surface that state inline rather than silently falling back to download or navigating away.
Profiles: base
Verified by: AC-053, AC-054, AC-103, AC-128, AC-231, AC-252, AC-255

**REQ-03-128**
Workbook preview and download affordances MUST invoke the Core 01 evidence-access handle contract. Clients MUST treat returned `href` values as opaque same-origin URLs and MUST NOT synthesize object-store URLs or parse handle tokens. Handle issuance intentionally does not send `client_txn_id`, and each successful issuance call mints a fresh handle rather than using issuance idempotency.
Profiles: base
Verified by: AC-053, AC-054, AC-103, AC-128, AC-231, AC-252, AC-253, AC-254

## 9. Mention resolution workflow

**REQ-03-129**
The inspector MUST support:

- resolving a selected unresolved mention to an existing entity,
- dismissing a selected mention,
- restoring a dismissed mention to unresolved state.
The current profile MAY expose a separate explicit host-or-identity create flow seeded from a selected mention, but that flow MUST use the ordinary entity-create or create-or-upsert contract and then resolve the mention to the resulting existing `record_id`; it MUST NOT be implemented as omitted-target `resolve_item` behavior.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-03-130**
Resolving a selected mention, and any explicit create flow seeded from a selected mention, MUST preserve the raw mention.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-03-131**
If an explicit create flow is seeded from a selected mention, it MUST create or upsert exactly one entity by default and resolve only the selected mention unless the user later invokes an explicit bulk action.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-03-132**
Dismissing a selected mention MUST create a new attributed change, preserve the raw mention and source-bound mention identity, set `resolution_status='dismissed'`, clear any current resolution metadata, and remove or tombstone any corresponding active resolved `record_link` in the same `change_set`.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-03-133**
Dismissed mentions MUST NOT remain in the default unresolved resolution queue, MUST NOT contribute to `timeline.has_unresolved_mentions`, and MUST NOT appear in the active relationship collection value for the row. They MAY remain visible through history and a secondary dismissed-mentions section or toggle in the inspector.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

**REQ-03-134**
Restoring a dismissed mention MUST create a new attributed change that returns the mention to `unresolved`, preserves the raw mention unchanged, and leaves resolution metadata empty. Ordinary restore MUST NOT silently relink to any prior resolved entity; exact pre-dismiss state recovery belongs to reviewer rollback.
Profiles: base
Verified by: AC-006, AC-019, AC-020, AC-021, AC-188, AC-189, AC-190, AC-221, AC-222, AC-223, AC-224, AC-225, AC-231

### 9.1 Source-bound indicator workflow

**REQ-03-135**
The implementation MUST provide a same-surface enrichment flow, ordinarily the inspector, that supports:

- linking a selected raw source value or text span in a supported field to an existing indicator,
- creating a canonical indicator from a selected source-bound observation,
- dismissing or marking a selected source-bound observation as non-indicator,
- viewing observation provenance and current lifecycle intervals for the selected indicator.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-231

**REQ-03-136**
Indicator capture from Timeline, Notes, Evidence, and other supported source fields MUST preserve the raw source field unchanged. It MUST NOT require dedicated IOC columns on non-indicator sheets.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-231

**REQ-03-137**
Direct row creation on the Indicators system view MUST create canonical indicator records. Grid edits to an existing indicator row MUST be limited to the fields that the active `view_schema` declares writable for existing indicator rows. In the base profile, `cartulary.view.indicators.v1` declares no writable fields for existing indicator rows. The exact identity-defining immutable field set for an existing indicator row is `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, `indicator.normalized_value`, and, when populated and used by the canonical dedupe key, `indicator.hash_algorithm` plus `indicator.hash_value`. `indicator.stix_pattern` and `indicator.defanged_value` remain create-only in this view but MUST NOT be treated as identity-defining. Observation creation from a source field MUST remain distinct from canonical indicator creation even when both happen in one analyst action.
Profiles: base
Verified by: AC-017, AC-072, AC-073, AC-074, AC-075, AC-076, AC-077, AC-078, AC-079, AC-231

## 10. Reviewer history and rollback workflow

### 10.1 Reviewer lens

**REQ-03-138**
The history experience MUST remain row-centric.
Profiles: base
Verified by: AC-007, AC-231

**REQ-03-285**
When a row history surface is rendered inside the inspector, the inspector's Details, Relationships, Evidence, History, Workflow, and row-local action sections MUST share one active `record_id` subject. If the active saved row changes while the inspector is open, every panel MUST retarget to the new active `record_id` or clear before showing new content. The History section MUST render a loading or error state for the new target until its history response is accepted and MUST NOT continue displaying history items, rollback previews, delete or restore confirmations, or row-local history actions from the previously active row. A deleted-row restore surface MAY remain bound to the deleted row only while no different live saved row is active.
Profiles: base
Verified by: AC-007, AC-215, AC-231

### 10.2 Minimum history presentation

**REQ-03-139**
In the base profile, the history panel for a selected row MUST show:

- actor,
- timestamp,
- operation,
- diff summary expanded to field, link, mention, tag, evidence-entry, and capture-state-transition units.
Profiles: base
Verified by: AC-007, AC-215, AC-231

**REQ-03-140**
The history surface MUST receive enough structured metadata from `GET /api/v1/records/{record_id}/history` to render only legal reviewer actions without client-side inference from visible text. At minimum, each displayed logical history item MUST expose `change_set_id`, `reversible`, `available_rollback_actions[]`, optional `history_entry_ref` when the item maps to exactly one reversible mutation target, and `revision_no` when whole-row restore is legal.
Profiles: base
Verified by: AC-007, AC-215, AC-231

### 10.3 Rollback granularity

**REQ-03-141**
The reviewer UI MUST allow rollback of a single logical history entry when that entry maps to one reversible mutation target, including:

- one scalar field edit,
- one link add or remove,
- one tag add or remove,
- one mention resolve, dismiss, or restore,
- one auto-resolution or auto-match,
- one evidence attach or detach association.
Profiles: base
Verified by: AC-010, AC-011, AC-012, AC-215, AC-216, AC-217, AC-218, AC-231

**REQ-03-142**
The UI MUST also expose:

- whole-row restore,
- whole-change-set rollback,
Profiles: base
Verified by: AC-010, AC-011, AC-012, AC-215, AC-216, AC-217, AC-218, AC-231

as secondary actions for destructive or multi-target changes.

Arbitrary user-selected subsets of fields from historical snapshots are not required in the base profile.

### 10.4 Rollback semantics

**REQ-03-143**
Rollback MUST create a new attributed revision. It MUST NOT rewrite prior history in place.
Profiles: base
Verified by: AC-216, AC-217, AC-218, AC-231

**REQ-03-144**
The reviewer client MUST choose rollback scope only from the selectors and action metadata returned by `GET /api/v1/records/{record_id}/history`. It MUST NOT infer legal rollback scope from visible labels, diff text, or storage-specific identifiers.
Profiles: base
Verified by: AC-216, AC-217, AC-218, AC-231

## 11. Clipboard paste and import workflows

### 11.1 Clipboard paste

**REQ-03-145**
Clipboard paste is base-profile functionality and MUST remain part of the base workbook interaction surface.
Profiles: base
Verified by: AC-003, AC-231

**REQ-03-146**
Clipboard paste alone does not satisfy the Import Extension Profile and MUST NOT be used to claim file-based structured import support.
Profiles: base, import
Verified by: AC-003, AC-231, AC-232

**REQ-03-147**
Pasting TSV or CSV into the timeline sheet MUST create or update multiple rows starting from the selected cell.
Profiles: base, import
Verified by: AC-003, AC-231, AC-232

Default interactive Ctrl+V tabular dispatch MUST require an unambiguous tabular signal: tab, newline, carriage return, or a future explicit paste-as-table command. A single-line comma-only `text/plain` payload such as `Hello, world` MUST be treated as scalar text by default even though explicit API-level CSV ingest remains supported.

When clipboard paste updates existing rows through `record` targets, every target MUST belong to the addressed incident and the active addressed workbook surface. Target ownership and visibility MUST be validated before any row-version comparison, conflict construction, batch commit, or response row serialization. A paste containing a missing, foreign-incident, wrong-surface, wrong-type, or deleted record target MUST fail closed as one rejected batch rather than partially committing other targets.

**REQ-03-148**
Known columns MUST map directly.
Profiles: base
Verified by: AC-003, AC-231

**REQ-03-149**
Unknown columns MUST be preserved in `raw_capture.import_columns` or equivalent structured raw-capture storage.
Profiles: base
Verified by: AC-003, AC-231

**REQ-03-150**
Host and identity text from pasted cells MUST follow the same `entity_binding_mode` contract as interactive edits.
Profiles: base
Verified by: AC-003, AC-231

When the active workbook sheet is Hosts or Identities, tabular clipboard paste MUST use `entity_binding_mode='entity_origin'` through the shared tabular-ingest contract. The server remains responsible for exact-match reuse and stub creation for those entity-origin rows.

**REQ-03-151**
Successful non-conflicting writes from one paste action MUST appear as:

- one visible `change_set`,
- ordered mutation entries,
- one row revision per affected record.
Profiles: base
Verified by: AC-003, AC-231

**REQ-03-152**
Any same-field conflicts arising from that paste MUST remain outside that committed batch until explicitly resolved. Each later same-field conflict resolution MUST create its own attributed `change_set`.
Profiles: base
Verified by: AC-003, AC-231

### 11.2 Import Extension Profile

File-based structured import beyond clipboard paste belongs to the **Import Extension Profile**.

#### 11.2.1 Assistant boundary

**REQ-03-153**
If implemented, the file-based import path MUST be exposed as a **Phase 2 Workbook Import Assistant** for structured onboarding of CSV and XLSX sources.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

Clipboard paste remains the base hot-path ingest surface. File-based workbook import remains an extension-profile capability.

**REQ-03-154**
The assistant MUST be the only application surface that knows about workbook discovery, sheet heuristics, Excel tables, named ranges, parser quirks, previewing, and workbook downgrade warnings.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-155**
The file-based import path MUST be isolated behind the dedicated `imports` module defined by Core 01.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-156**
The file-based import path MUST use the same stable tabular-ingest contract and shared mapping engine as clipboard-driven structured ingest, including the same `entity_binding_mode` contract.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-157**
Each candidate unit MUST support operator preview, mapping, select, and skip before any apply step.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-158**
The assistant MUST prioritize mappings for timeline, systems or hosts, accounts or identities, indicators, evidence tracker, and VERIS-like summaries when present.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-159**
The import path MUST preserve unknown columns in `raw_capture` or `custom_attrs`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-160**
The import path MUST determine `mention_origin` versus `entity_origin` from contracts rather than sheet labels alone.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

**REQ-03-161**
The import path MUST avoid auto-resolution of imported host or account aliases during ingest.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-067, AC-232

#### 11.2.2 `import_session`

**REQ-03-162**
An `import_session` MUST be the durable, audit-visible coordinator for one uploaded source file and one operator-driven import workflow.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-163**
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

`selected_unit_ids[]` MUST be a persisted deduplicated session-local list. When serialized it MUST default to `[]` when empty and MUST be ordered in the deterministic apply order defined by REQ-03-186.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-164**
`session_status` MUST use the closed vocabulary `created`, `discovered`, `mapped`, `ready_to_apply`, `applying`, `partially_applied`, `applied`, `failed`, and `canceled`, with these meanings:

- `created`: the upload was accepted and discovery is not yet complete;
- `discovered`: discovery completed and no apply is active;
- `mapped`: at least one unit has an approved mapping, but the session is not yet `ready_to_apply`;
- `ready_to_apply`: the persisted selected set is non-empty and every selected unit is `ready`;
- `applying`, `partially_applied`, `applied`, `failed`, and `canceled`: durable apply-lifecycle states.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-165**
`source_content_sha256` MUST be computed from the exact uploaded file bytes, not from a parsed or normalized representation.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-166**
`parser_version` MUST identify the stable behavior version of the import adapter profile. It MUST NOT be limited to a raw parser-library semantic version.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-167**
`nonblocking_warning_codes[]` MUST use the same closed vocabulary as `import_unit.warning_codes[]` and MAY aggregate source-level warnings that do not block preview or apply.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

**REQ-03-168**
Discovery and preview MUST be read-only against incident state. Preview MUST expose inert display data only and MUST NOT advance lifecycle state by itself. Apply MUST execute as a background job and MUST NOT block ordinary workbook editing.
Profiles: import
Verified by: AC-027, AC-063, AC-064, AC-232

#### 11.2.3 `import_unit`

**REQ-03-169**
An `import_unit` MUST be the explicit unit of structured source material selected from an `import_session`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-170**
An `import_unit` MUST persist, at minimum:

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
- optional `mapping_fingerprint`,
- an approved mapping plan, or an equivalent structured realization sufficient to reconstruct the exact Core 01 §17.2 `approved_mapping` object,
- `unit_status`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-171**
`locator_kind` MUST use the closed vocabulary `csv_file`, `xlsx_used_range`, `xlsx_table`, `xlsx_named_range`, and `xlsx_region`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-172**
`unit_status` MUST use the closed vocabulary `discovered`, `selected`, `mapped`, `ready`, `applying`, `applied`, `skipped`, `rejected`, and `failed`, with these meanings:

- `discovered`: no approved mapping exists and the unit is not selected;
- `selected`: the unit is in `selected_unit_ids[]` and no approved mapping exists;
- `mapped`: an approved mapping exists but readiness checks still fail;
- `ready`: an approved mapping exists and all readiness checks pass;
- `skipped`: the unit is explicitly excluded from persisted selection and retains any prior approved mapping;
- `applying`, `applied`, `rejected`, and `failed`: durable processing states.

Re-selecting a skipped unit MUST recompute `unit_status` from current durable mapping state back to `selected`, `mapped`, or `ready` rather than creating a new mapping plan.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-173**
The semantic identity of an `import_unit` MUST be the tuple `source_content_sha256 + canonical_locator + parser_version`. If the implementation stores one derived key for that identity, it MUST compute that key as the SHA-256 of the canonical JSON serialization of `{source_content_sha256, locator, parser_version}`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-174**
The current profile MUST use the following canonical locator shapes:

- `csv_file`: one trivial single-unit locator bound to the uploaded file,
- `xlsx_used_range`: `{sheet_name, rect_a1}`,
- `xlsx_table`: `{sheet_name, table_name, rect_a1}`,
- `xlsx_named_range`: `{defined_name, sheet_name, rect_a1}`,
- `xlsx_region`: `{sheet_name, rect_a1}`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-175**
The locator MUST be a canonical contract object. It MUST NOT rely on display labels or UI-only wording.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-176**
`source_rect_a1` MUST identify the deterministic rectangular extent of the imported unit. For `csv_file`, it MUST cover the full parsed rectangle of the file.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-177**
`header_row_ref` MUST be a deterministic 1-based row reference within `source_rect_a1`.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

**REQ-03-178**
`mapping_fingerprint` MAY be absent while `unit_status` is `discovered` or `selected`. The approved mapping plan or equivalent structured realization MAY be absent until mapping approval. Once a mapping is approved, both the approved mapping plan and `mapping_fingerprint` MUST be persisted and used for duplicate-apply detection.
Profiles: import
Verified by: AC-027, AC-028, AC-029, AC-063, AC-064, AC-065, AC-066, AC-232

#### 11.2.4 Discovery and batch semantics

**REQ-03-179**
Whole-workbook import MUST mean an orchestrated batch of explicit `import_unit` objects discovered from one uploaded workbook. It MUST follow this flow:

1. upload one XLSX file and create one `import_session`,
2. discover candidate `import_unit` objects,
3. allow operator preview, approve mapping, and select or skip per unit,
4. validate duplicate-apply and overlap constraints,
5. apply the persisted selected set in deterministic order,
6. record one session outcome and one apply outcome per unit.

Preview rows MUST be returned in source order and capped to 50 data rows. Select and skip actions MUST recompute `session_status` immediately from the persisted selected set and current unit states.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-180**
Whole-workbook import MUST NOT mean preserving workbook object identity, formula behavior, chart behavior, pivot behavior, protection behavior, merged-cell layout behavior, or general XLSX semantics outside the assistant.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-181**
The current profile MUST support discovery of:

- parser-resolved non-empty used ranges,
- Excel tables using their current rectangular extent only,
- named ranges only when they are static, single-sheet, and single-rectangle,
- operator-selected rectangular regions from a sheet preview.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-182**
Dynamic named-range formulas, external references, and multi-area named ranges MUST be rejected as unsupported.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-183**
Table filters, sorts, styles, hidden states, and other presentation metadata MUST NOT change imported semantics.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-184**
A whole-workbook batch MUST remain unit-atomic rather than session-atomic. Each applied unit MUST yield its own `change_set`. The session MAY end in `partially_applied`.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-185**
The persisted selected set for one batch MUST have disjoint source-cell coverage. Overlapping units MAY be previewed, but they MUST NOT be jointly applied until the operator reduces the persisted selection to a disjoint set.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

**REQ-03-186**
The default apply order for the persisted selected set MUST be deterministic: workbook sheet order, then top-left rectangle position, with operator-added regions ordered by explicit session sequence when their rectangles would otherwise compare equal.
Profiles: import
Verified by: AC-027, AC-064, AC-065, AC-066, AC-232

#### 11.2.5 `mapping_fingerprint` and duplicate-apply detection

**REQ-03-187**
`mapping_fingerprint` MUST be the deterministic identity of the operator-approved header-to-field plan for one `import_unit`. It MUST be derived from the same fully closed approved mapping plan that backs the exact Core 01 §17.2 read-side `approved_mapping` object rather than from a lossy summary.
Profiles: import
Verified by: AC-065, AC-232

**REQ-03-188**
The current profile closes the mapping-plan registries and entry semantics as follows:

- `unknown_column_policy` MUST use exactly `preserve_raw_capture`, `preserve_custom_attrs`, or `reject_if_unmapped`;
- `transform_id` MUST use exactly `null`, `trim_v1`, `collapse_whitespace_v1`, `lowercase_v1`, or `split_delimited_v1`;
- `empty_value_policy` MUST use exactly `omit_field` or `write_null`.

The fingerprint input MUST include, at minimum:

- `mapping_contract_version`,
- `target_view_schema_id`,
- `header_row_ref`,
- `data_start_row_ref`,
- `unknown_column_policy`,
- one exhaustive ordered entry per discovered source column containing `source_column_ordinal`, `source_header_text`, `field_key`, `entity_binding_mode`, and any declared `transform_id`, `transform_options`, or `empty_value_policy`.

For the current profile:

- `source_columns[]` MUST be exhaustive over discovered source columns;
- `source_header_text` MUST be the raw imported header text or `null`;
- `field_key = null` means intentionally unmapped;
- `entity_binding_mode = null` is required for unmapped columns and non-entity targets;
- duplicate non-null target `field_key` values are invalid;
- transform execution order is parser extraction, then optional mapping transform, then target-field normalization;
- `split_delimited_v1` is the only transform that MAY use non-empty options, and those options are limited to `delimiter`, `trim_items`, and `drop_empty_items`;
- `delimiter` for `split_delimited_v1` MUST be one of `,`, `;`, `|`, `\n`, or `\t`;
- all current-profile transforms other than `split_delimited_v1` MUST use `transform_options = {}`;
- `preserve_raw_capture` is legal only when the target view's authoritative record model exposes `raw_capture`;
- `preserve_custom_attrs` is legal only when the target view allows `custom_attrs`;
- `reject_if_unmapped` blocks `ready` while any source column is intentionally unmapped;
- `write_null` blocks `ready` when the target field is not clearable;
- `formula_cached_value_missing` remains a `ready` blocker under the closed warning vocabulary in §11.2.6.

When `target_view_schema_id='cartulary.view.timeline.v2'` and `unknown_column_policy='preserve_raw_capture'`, every intentionally unmapped source cell applied to a Timeline row MUST be copied onto that target Timeline row under `timeline_events.raw_capture.import_columns[]`. Each preserved entry MUST carry at least `import_session_id`, `import_unit_id`, `mapping_fingerprint`, `source_file_kind`, `source_content_sha256`, `parser_profile_id`, `parser_version`, `locator_kind`, `locator`, `source_rect_a1`, `source_row_ordinal`, `source_column_ordinal`, `source_header_text`, `raw_value`, and `cell_kind`. Preview-only import tables are not sufficient provenance for applied Timeline rows.
Profiles: import
Verified by: AC-065, AC-232

**REQ-03-189**
`header_row_ref` and `data_start_row_ref` MUST use the same 1-based row coordinate system within `source_rect_a1`.
Profiles: import
Verified by: AC-065, AC-232

**REQ-03-190**
The implementation MUST serialize that fully closed mapping plan canonically, with lexicographically sorted object keys, normalized option objects, and source columns ordered by `source_column_ordinal`, then hash the result as lower-hex SHA-256. The persisted approved mapping plan or equivalent structured realization MUST be sufficient to reconstruct the exact Core 01 §17.2 `approved_mapping` object and recompute the same `mapping_fingerprint` deterministically, without lossy normalization, inference from display labels, or dependence on non-authoritative UI state.
Profiles: import
Verified by: AC-065, AC-232

**REQ-03-191**
Visible labels MUST NOT affect `mapping_fingerprint` except through the raw imported `source_header_text`.
Profiles: import
Verified by: AC-065, AC-232

**REQ-03-192**
The assistant MUST compare the fully closed mapping plan when evaluating the `(import_unit_id, mapping_fingerprint, incident_id)` tuple for duplicate-apply detection. It MUST warn on re-applying that same tuple and MUST default to blocking the apply until the operator explicitly chooses re-import.
Profiles: import
Verified by: AC-065, AC-232

#### 11.2.6 Closed warning vocabulary and workbook downgrade semantics

**REQ-03-193**
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
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-194**
Hard failures MUST be reported through normal job diagnostics rather than `warning_code[]`.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-195**
Formula cells and other spreadsheet active content MUST be treated as inert input. The implementation MUST NOT execute workbook formulas, macros or VBA, workbook automation, or external links as live logic during preview or apply.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-196**
Encrypted or password-protected files that cannot be parsed MUST be rejected before discovery.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-197**
If a stored cached value exists for a formula cell, preview and import MAY use that inert cached value and MUST emit `formula_inert`. If no cached value exists for a mapped formula cell, the unit MUST NOT enter `ready` until the operator excludes or remaps the affected column, and `formula_cached_value_missing` MUST be emitted.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-198**
Merged cells MUST use only the upper-left anchor value for deterministic rectangular import. Covered cells MUST otherwise be treated as empty, and the unit MUST emit `merged_layout_downgraded`.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-199**
Comments or notes MUST be preserved only as source metadata and MUST emit `comments_metadata_only` when encountered.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-200**
Pivot tables, charts, and external links MUST NOT be interpreted as executable or live workbook logic and MUST emit `pivot_ignored`, `chart_ignored`, or `external_links_ignored` when encountered.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-201**
Workbook or sheet protection that does not block parsing MUST be treated only as source metadata and MUST emit `workbook_protection_metadata_only` or `sheet_protection_metadata_only` as applicable.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-202**
Hidden rows, hidden columns, and active filters MUST NOT change imported semantics in the current profile and MUST emit `filtered_or_hidden_state_ignored` when encountered.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-203**
Unsupported workbook features MUST be downgraded with one or more declared `warning_code[]` values, preserved only as raw source metadata, or rejected as unsupported.
Profiles: import
Verified by: AC-063, AC-067, AC-232

**REQ-03-204**
XLSX inputs MUST be staged and parsed as hostile archive content consistent with Core 04.
Profiles: import
Verified by: AC-063, AC-067, AC-232

## 12. Auto-resolution policy

### 12.1 Allowed scope

**REQ-03-276**
Auto-resolution eligibility is owned by this section. In the current profile, visible Timeline v2 cells remain source-preserving strings. The system MAY offer or commit host, identity, indicator, MITRE, or data-source suggestions only through inspector-side actions or hidden action fields whose stable `field_key` is declared by Core 01. Such suggestions or resolutions MUST NOT rewrite the source visible cell string. No workflow is eligible for auto-resolution based on visible labels.
Profiles: base
Verified by: AC-205, AC-231, AC-388, AC-392, AC-393

### 12.2 Required confidence conditions

`auto_resolution_confidence = 100` applies only when all of the requirements in this subsection are satisfied.

**REQ-03-277**
For auto-resolution comparison, the system MUST derive one exact comparison value named `auto_resolution_candidate_text`. This value is derived only and MUST NOT by itself require persisted source-model state in the current profile. The system MUST derive it from the JSON-decoded submitted `raw_text` by applying `mention_token_text_v1` normalization and then locale-independent Unicode case folding for comparison only. Exact alias equality for auto-resolution MUST compare `auto_resolution_candidate_text` to candidate alias text after applying that same `mention_token_text_v1` normalization and the same locale-independent Unicode case folding for comparison only. This comparison substrate MUST NOT alter authoritative alias storage or alias dedupe semantics.
Profiles: base
Verified by: AC-205, AC-231, AC-388, AC-389, AC-390, AC-391

**REQ-03-278**
The current-profile uncertainty suppressor grammar is closed. Auto-resolution MUST NOT occur when either of the following is true after deriving `auto_resolution_candidate_text` under REQ-03-277:

- the submitted token contains ASCII `?` or ASCII `~` anywhere,
- the submitted token contains one or more whole whitespace-delimited suppressor tokens from exactly this lexical set: `maybe`, `prob`, `probably`, `approx`, `approximately`.
Profiles: base
Verified by: AC-231, AC-389, AC-390

**REQ-03-279**
Auto-resolution MUST NOT depend on punctuation stripping, parenthetical stripping, duplicate-punctuation collapse, token deletion, transliteration, stemming, fuzzy matching, or any locale- or language-specific uncertainty lexicon in the current profile. If exact alias equality would require any such rewrite, the token is not eligible for auto-resolution.
Profiles: base
Verified by: AC-231, AC-391

**REQ-03-205**
Anything below 100 MAY drive ranking or suggestions. The system MUST take the ordinary unresolved path whenever any required condition in this subsection is not satisfied, any suppressor match occurs, or exact alias equality would require a forbidden rewrite. In that case the system MUST create or preserve the corresponding `entity_mention` with `resolution_status='unresolved'`, MUST leave `resolved_record_id=null`, MUST create no active resolved `record_link`, MUST emit no `provenance='auto_match'`, and MUST show no auto-resolution disclosure. Suggestions or ranking MAY still be shown, but they MUST remain non-mutating.
Profiles: base
Verified by: AC-205, AC-231, AC-389, AC-390, AC-391

### 12.3 Required write effects

**REQ-03-206**
When auto-resolution occurs, the system MUST still insert an `entity_mentions` row for the typed token with:

- `resolution_status='resolved'`,
- `resolved_record_id` set to the chosen record.
Profiles: base
Verified by: AC-205, AC-231

**REQ-03-207**
The corresponding `record_links` row MUST use:

- `provenance='auto_match'`,
- `confidence=100`.
Profiles: base
Verified by: AC-205, AC-231

### 12.4 Allowed and forbidden workflows

The current profile's allowed workflows are fixed by REQ-03-276.

**REQ-03-208**
Auto-resolution MUST NOT occur in:

- the inspector’s explicit resolve flow,
- Hosts or Identities alias-edit cells,
- merge or dedupe workflows,
- file-based import through the Import Extension Profile,
- background jobs,
- asynchronous enrichment or cleanup,
- any workflow that would create a new canonical host or identity or edit alias rows without explicit confirmation.
Profiles: base, import
Verified by: AC-205, AC-231, AC-232, AC-392, AC-393

### 12.5 Disclosure and undo

**REQ-03-209**
The UI MUST NOT silently auto-resolve.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-210**
When auto-resolution occurs, the current sheet MUST show an immediate non-modal disclosure on the same surface that includes:

- the raw token,
- the canonical target,
- the matched alias text,
- direct `Undo`,
- direct `Review`.

The disclosure lifecycle is bound to the current auto-resolved mention state. Activating `Review` MUST NOT by itself dismiss the disclosure. The disclosure MUST remain visible until the user successfully corrects the target, successfully reverts the mention to unresolved state, or a refreshed row proves that the same source-bound mention is no longer an auto-resolved resolved reference.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-211**
For batch paste, the disclosure MUST also include the number of tokens auto-resolved in that visible change set.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-212**
The resolved chip or cell MUST remain inspectably marked as auto-resolved.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-213**
Row history MUST preserve:

- raw token,
- matched alias text,
- confidence,
- mutation source.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-214**
`Undo` from the immediate disclosure MUST:

- restore the raw unresolved token,
- remove the auto-created link,
- preserve focus,
- preserve scroll position.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-215**
After the immediate disclosure expires, the user MUST still be able to choose `Revert to unresolved` from the chip context or row history in no more than two actions.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

**REQ-03-216**
That later correction MUST be a new attributed revision.
Profiles: base
Verified by: AC-006, AC-188, AC-189, AC-190, AC-205, AC-231

## 13. Keyboard and editing contract

### 13.1 Grid editing

**REQ-03-217**
Selecting a cell and typing MUST edit it immediately.
Profiles: base
Verified by: AC-005, AC-043, AC-231

**REQ-03-218**
Enter MUST commit and move vertically. Tab MUST commit and move horizontally.
Profiles: base
Verified by: AC-005, AC-043, AC-231

**REQ-03-219**
Relationship cells MUST accept raw typing and MUST NOT require picker-first interaction.
Profiles: base
Verified by: AC-005, AC-043, AC-231

**REQ-03-281**
When a visible edit to a field bound to `direct_scalar_contract_id=timestamp_instant_v1` fails local validation or authoritative save, the client MUST keep the analyst's typed text as unsaved local state in the active editor, MUST NOT render that typed text as authoritative committed row state, MUST NOT silently coerce it to another persisted value, MUST NOT treat the empty string as an authoritative clear, and MUST require explicit correction, discard, or explicit JSON `null` clear when the bound field declares `clearable=true`. For a bound field that declares `clearable=false`, an attempted clear MUST fail closed and the visible local state MUST remain unsaved until corrected or discarded.
Profiles: base
Verified by: AC-231, AC-354

This requirement defines workbook-surface editing consequences only. Core 01 §18A continues to own lexical, normalization, equality, and authoritative clear semantics for `timestamp_instant_v1`.

**REQ-03-280**
Current-profile manual link actions are scoreless. The workbook MUST NOT require analyst entry of numeric link confidence for ordinary manual linking, and any unsupported attempt to supply one is a validation error rather than a silent downgrade.
Profiles: base
Verified by: AC-394, AC-395, AC-396, AC-231

### 13.2 Required keyboard actions

**REQ-03-220**
The base profile MUST support:

- Arrow keys to move selection,
- Enter and Shift+Enter row navigation,
- `Ctrl+V` for paste,
- `Ctrl+K` for quick link or resolve on the current cell,
- `Space` to preview linked evidence for the selected row,
- `Alt+H` to open history for the selected row,
- `Esc` to close the inspector and return focus to the prior cell.
Profiles: base
Verified by: AC-005, AC-231

### 13.3 Bulk editing

**REQ-03-221**
The implementation MUST support:

- multi-cell paste,
- fill-down,
- multi-row tag assignment.
Profiles: base
Verified by: AC-003, AC-040, AC-231

**REQ-03-222**
Bulk edits are mutation batches. They MUST NOT rely on hidden macro semantics.
Profiles: base
Verified by: AC-003, AC-040, AC-231

The current base-profile explicit bulk command vocabulary is:

- `fill_down_v1`: copies one submitted source value for one writable non-collection field into explicit stable row targets identified by `record_id` and `base_row_version`;
- `multi_row_tag_assignment_v1`: applies one submitted tag label to explicit stable Timeline row targets through the `timeline.tags` collection contract.

Each command MUST commit as one attributable batch when all accepted target mutations commit, MUST record one visible `change_set` for the committed non-conflicting portion, and MUST reject presentation-only, group-row, vendor-coordinate, row-index-only, missing, deleted, wrong-surface, wrong-type, or foreign-incident targets. Bulk target ownership and visibility MUST be validated before row-version comparison, conflict construction, batch commit, or response row serialization. Later expansion MAY add additional explicit command kinds, but MUST NOT reinterpret these command identifiers.

## 14. Sorting, filtering, and grouping

### 14.1 Sort and filter behavior

**REQ-03-223**
Column-header sort and inline filter chips MUST apply without leaving the sheet. A visible column header MAY initiate sorting only when the active `view_schema` declares that field sortable. Header sort MUST use `header_sort_field_key` when the field registry declares it and otherwise the visible field's own `field_key`. Visible collection or relationship columns not present in the active schema's `sort_fields` MUST NOT initiate a sort. The client MUST treat discovery `fields[]`, `sort_fields`, `filter_fields`, `synthetic_filter_predicates[]`, and `grouping_fields` as the authoritative source of field identity and capability.
Profiles: base
Verified by: AC-013, AC-014, AC-044, AC-047, AC-124, AC-184, AC-185, AC-231, AC-363

**REQ-03-224**
Sorting and filtering MUST apply to underlying rows before grouping is computed. Clearing a user sort override MUST return the surface to schema default sort only, represented by omitted `sort` on `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` and by canonical persisted `query_json.sort=[]` when the state is saved in a saved view. Clearing grouping MUST persist omitted `query_json.group_by`; the client MUST NOT persist JSON `null` for that state. When the server returns canonicalized `meta.query`, the client MUST treat that canonical query contract as authoritative for rendered filter chips, restored saved-view state, and cursor continuation. For workbook surface queries triggered by sort, filter, group, surface, cursor, or refresh state, the client MUST render only the latest applicable query result for the active incident, surface, and query identity. Older in-flight query results and query errors that have been superseded by a newer applicable query MUST NOT replace rendered rows, clear rendered rows, change visible query errors, or drive access-loss handling. Even when a query result is still latest for the active query identity, it MUST NOT overwrite a locally accepted newer committed row for the same `record_id`; if a full query response is older than the client's accepted row-version high-water mark for any returned row, the client MUST preserve the newer local committed row state and refresh through the same HTTP query route to recover canonical sort, filter, and grouping placement. The client MUST NOT preserve caller-entered order for set-like `values[]`, MUST NOT preserve case-only variants of `prefix.value` as authoritative state, and MUST NOT re-tokenize `full_text` independently of the server contract.
Profiles: base
Verified by: AC-013, AC-014, AC-044, AC-047, AC-124, AC-184, AC-185, AC-231, AC-360, AC-363

### 14.2 Grouping boundary

Grouping is a presentation-only transform over the current filtered result set for any workbook surface whose active `view_schema` declares `grouping_fields`.

**REQ-03-225**
Grouping MUST NOT create, delete, or mutate:

- source records,
- projection rows,
- links,
- tags.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231, AC-364

### 14.3 Allowed grouping keys

**REQ-03-226**
A workbook surface whose active `view_schema` declares `grouping_fields` MUST support `Group: None` plus exactly one active grouping key in the base profile. The grouping control MUST offer exactly `Group: None` plus the active surface's declared `grouping_fields` in the canonical order returned by the active `view_schema`. It MUST NOT offer undeclared keys, visible-label aliases, or implementation-defined extras.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231, AC-364

**REQ-03-227**
For any workbook surface whose active `view_schema` declares non-empty `grouping_fields`, the active grouping key MUST be stored as a stable contract value in `query_json.group_by`, not as a visible label. `Group: None` is represented by omitted `group_by` on the public query route and omitted `query_json.group_by` in saved views. The UI MUST NOT send explicit `null` for that state.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231, AC-360

For Timeline sheets, the allowed grouping keys are exactly:

- `timeline.date_entered_sort_day`,
- `timeline.activity_time_pair_state`,
- `timeline.capture_state`,
- `timeline.has_evidence`,
- `timeline.has_unresolved_mentions`.

The base-profile Timeline whitelist is frozen at these five keys for the current profile. No additional Timeline grouping key is permitted in the base profile.

**REQ-03-228**
`timeline.event_type` MUST NOT be exposed as a base-profile grouping key.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231

Grouping by arbitrary custom columns, formulas, ad hoc expressions, or visible labels is out of scope for current conformance. For Timeline sheets, grouping by visible source-text fields such as Activity Synopsis, MITRE, Device/Object, IP Address, RAW Activity, or Data Source is also out of scope.

### 14.4 Grouping value rules

**REQ-03-229**
Grouping keys MUST be scalar contract-backed values. Unless the active `view_schema` explicitly overrides group-order behavior, grouped workbook surfaces sort groups by the grouping key's normal sort comparator in ascending order with null buckets last. In the current profile, no grouped surface other than Timeline defines a group-order override. The current row sort applies unchanged within each group.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231, AC-364

**REQ-03-230**
Timeline group order overrides the generic rule in REQ-03-229 and MUST be deterministic:

- `timeline.date_entered_sort_day` sorts by bucket value descending, with null buckets last,
- `timeline.activity_time_pair_state` sorts `paired_generated`, `paired_user_preserved`, `paired_mismatch`, `conversion_unavailable`, `disabled`, `empty`,
- `timeline.capture_state` sorts `rough`, `enriched`, `reviewed`, `superseded`,
- `timeline.has_evidence` and `timeline.has_unresolved_mentions` sort `true` then `false`.
Profiles: base
Verified by: AC-024, AC-025, AC-026, AC-231

### 14.5 Group-header behavior

**REQ-03-231**
A grouped workbook surface whose active `view_schema` declares `grouping_fields` MUST expose exactly one derived group-header level. For the latest applicable query result rendered in one workbook surface, each visible grouping bucket derived from committed rows MUST have at most one visible group header.
Profiles: base
Verified by: AC-025, AC-026, AC-231, AC-364

**REQ-03-232**
Group headers are derived UI rows only. They MUST be derived solely from committed rows in the latest applicable query result, the active `group_by` key, and the corresponding `group_values[group_by]` bucket for those grouped rows. Local draft rows, create affordances, loading rows, and other recordless UI affordances MUST NOT create grouping buckets, join grouping buckets, or generate group headers. Group headers MUST NOT use manual labels, client heuristics that invent new buckets, subtotal rows, summary rows, spacer rows, or other synthetic rows. They:

- MUST NOT have a `record_id`,
- MUST NOT accept inline edits,
- MUST NOT accept paste targets,
- MUST NOT appear in exports,
- MUST NOT appear in revision history,
- MUST NOT become mutation targets.
Profiles: base
Verified by: AC-025, AC-026, AC-231, AC-364

The allowed group operations are:

- expand group,
- collapse group,
- expand all,
- collapse all,
- `Group: None`.

### 14.6 Edit movement across groups

A row MAY move between visible groups only when an edit changes the grouped field value.

**REQ-03-233**
Dragging a row between groups MUST NOT be a write path.
Profiles: base
Verified by: AC-026, AC-047, AC-231, AC-364

### 14.7 Collaborative state boundary

**REQ-03-234**
Transient expand and collapse state SHOULD remain client-local and MUST NOT be broadcast as collaborative state.
Profiles: base
Verified by: AC-047, AC-231, AC-364

**REQ-03-235**
Saved views MAY persist the default grouping key in `query_json.group_by` when the active `view_schema` declares `grouping_fields`. They MUST NOT persist another user’s live open or closed state.
Profiles: base
Verified by: AC-047, AC-231, AC-364

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

**REQ-03-236**
The Timeline sheet MUST read from `timeline_grid_projection` or an equivalent projection.
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

**REQ-03-237**
When exposed over the public HTTP surface, Timeline reads MUST use the view-shaped query route `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`. New Timeline rows MUST use `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows`. Updates to existing Timeline rows MUST use `PATCH /api/v1/records/{record_id}` with `view_schema_id`, `base_row_version`, `client_txn_id`, and `changes[]` keyed by `field_key`. Group headers remain client-local presentation state and MUST NOT appear as writable API rows.
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

**REQ-03-238**
The implementation MUST preserve the following Timeline v2 write-back semantics:
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

| Timeline field or column | Read model | Required write-back behavior |
| --- | --- | --- |
| Date Entered | `date_entered_text` | update `timeline_events.date_entered_text` only |
| Analyst | `analyst_text` | update `timeline_events.analyst_text` only |
| MITRE | `mitre_stage_text` | update `timeline_events.mitre_stage_text` only |
| Device/Object | `device_object_text` | update `timeline_events.device_object_text` only |
| IP Address | `ip_address_text` | update `timeline_events.ip_address_text` only |
| Activity Date (UTC) | `activity_utc_text` | update `timeline_events.activity_utc_text`; fixed-offset conversion MAY generate `activity_local_text` only under Core 01 §7.4.1 |
| Activity Date (Local Time) | `activity_local_text` | update `timeline_events.activity_local_text`; fixed-offset conversion MAY generate `activity_utc_text` only under Core 01 §7.4.1 |
| RAW Activity | `raw_activity_text` | update `timeline_events.raw_activity_text` only and render as inert escaped text |
| Activity Synopsis | `activity_synopsis_text` | update `timeline_events.activity_synopsis_text` only |
| Data Source | `data_source_text` | update `timeline_events.data_source_text` only |
| Inspector mentions, tags, evidence, MITRE/entity/indicator/link suggestions | hidden action fields or derived/link records | create or update inspector-side state without rewriting any visible Timeline v2 cell string |


For the lifecycle machine in §6, the current Timeline write surfaces and row-anchored mutation types are classified as follows:

- `capture-state-material`: any committed change to one of the ten visible Timeline v2 operational fields, Timeline-row evidence attach or detach, and row-anchored source-bound MITRE/entity/indicator observation create, link, dismiss, or equivalent typed-link mutation initiated from the row or its inspector.
- not `capture-state-material`: hidden tag edits, generated conversion-pair updates that only fill a server-generated paired Activity Date value, explicit `mark-reviewed` or `supersede` lifecycle actions, selection, focus, presence, sort, filter, grouping, projection rebuild, and idempotent no-op retry.

**REQ-03-239**
Any future Timeline writable `field_key`, dedicated Timeline action route, or row-anchored mutation surface MUST declare whether it is `capture-state-material` before it can claim base-profile conformance.
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

Core 01 §7.4.1 fixes the authoritative Timeline `view_schema_id`, ordered field set, hidden technical fields, default sort tuple, filter whitelist, and grouping whitelist for the base profile.

**REQ-03-240**
Reads MAY be denormalized. Write-back MUST remain intent-aware and contract-driven.
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

**REQ-03-241**
Timeline v2 visible fields remain raw source-preserving text fields. When the implementation supports MITRE, entity, data-source, or indicator capture from such a field, write-back MUST preserve the raw cell text and create or update source-bound derived records, suggestions, links, or `indicator_observation` rows separately. It MUST NOT rewrite the raw field to a canonical chip, indicator label, MITRE object, URL, formula result, or entity reference.
Profiles: base
Verified by: AC-119, AC-120, AC-124, AC-125, AC-188, AC-189, AC-190, AC-191, AC-192, AC-193, AC-231

## 16. Entity, evidence, and inspector surfaces

### 16.1 Entity and evidence sheets

**REQ-03-242**
Hosts, Identities, and Evidence sheets MUST remain workbook-shaped peer sheets backed by their own projections.
Profiles: base
Verified by: AC-015, AC-045, AC-097, AC-098, AC-099, AC-100, AC-112, AC-116, AC-117, AC-118, AC-231

**REQ-03-243**
Canonical fields such as `display_name`, `hostname`, `upn`, host `location`, `os_platform`, `business_owner`, `criticality`, `containment_status`, identity `privilege_level`, `mfa_state`, `reset_status`, and evidence `title`, `requested_at`, `received_at`, `storage_ref`, `collector_party_text`, and `source_party_text` MUST be inline-editable where appropriate.
Profiles: base
Verified by: AC-015, AC-045, AC-097, AC-098, AC-099, AC-100, AC-112, AC-116, AC-117, AC-118, AC-231

**REQ-03-244**
Alias cells MUST behave like chip editors.
Profiles: base
Verified by: AC-015, AC-045, AC-097, AC-098, AC-099, AC-100, AC-112, AC-116, AC-117, AC-118, AC-231

**REQ-03-245**
Relationship-derived or machine-derived columns such as linked-event counts, evidence counts, blob-upload state, preview handles, and last-updated markers MUST be read-only and navigable.
Profiles: base
Verified by: AC-015, AC-045, AC-097, AC-098, AC-099, AC-100, AC-112, AC-116, AC-117, AC-118, AC-231

**REQ-03-246**
Entity normalization-state fields with exact token values `stub` and `canonical` MUST be projection-backed and read-only in-grid in the base profile. A normalization transition such as `stub -> canonical` MAY be exposed through the inspector or another explicit flow, but it MUST NOT be an ordinary cell edit on the Hosts or Identities sheet.
Profiles: base
Verified by: AC-015, AC-045, AC-097, AC-098, AC-099, AC-100, AC-112, AC-116, AC-117, AC-118, AC-231

### 16.2 Inspector

**REQ-03-247**
The inspector MUST:

- remain a non-blocking drawer rather than the default primary capture surface,
- expose details, relationships, evidence, and history views,
- keep the main grid visible while open,
- support mention resolution, dismissal, and restore, indicator observation linking, entity and indicator creation, party create or link or clear flows for requester, collector, source, audience, or attendee fields, indicator lifecycle inspection or editing, evidence inspection, and rollback,
- for view schemas that declare hidden writable fields, operate against the same canonical full-row state already delivered by query or row-refresh success and MUST NOT depend on an unstated generic record-read fetch to expose or clear those fields,
- preserve the distinction between raw mention lineage and current canonical links,
- preserve the distinction between raw indicator observation lineage and current canonical indicator links and lifecycle windows.
Profiles: base
Verified by: AC-006, AC-020, AC-023, AC-072, AC-073, AC-074, AC-075, AC-186, AC-187, AC-209, AC-210, AC-231, AC-278, AC-279, AC-366

**REQ-03-248**
The inspector is the enrichment surface. It MUST NOT be required for the common path of timeline row creation and editing.
Profiles: base
Verified by: AC-006, AC-020, AC-023, AC-072, AC-073, AC-074, AC-075, AC-186, AC-187, AC-209, AC-210, AC-231

**REQ-03-284**
When a compact Timeline relationship cell cannot show every active relationship item inline, the compact overflow indicator MUST be keyboard and pointer reachable and MUST provide a same-surface path to inspect the hidden relationship items in the row's Relationships inspector context. The compact overflow indicator MUST NOT become relationship authority; authoritative relationship and mention state remains the source-backed collection value and inspector actions defined by Core 01 and Core 02.
Profiles: base
Verified by: AC-006, AC-188, AC-190, AC-205, AC-231

**REQ-03-249**
For authorized users, the inspector MUST support explicit host and identity merge initiation. Merge MUST NOT be a bulk grid action in the base profile. The merge UI MUST require explicit survivor and loser selection and a final destructive-action confirmation that identifies both `record_id` values before submission. Before submission, that confirmation MUST show a deterministic plan derived from the currently loaded survivor and loser state, including exact-match identifiers that will be promoted into empty survivor canonical fields, exact-match identifiers that will be preserved on the survivor as additional active `exact_match_reuse` values, suggestion-only aliases that will be copied, and a statement that `provenance_only` values remain preserved on the historical loser and do not affect future reuse. When the selected host or identity has active secondary `exact_match_reuse` values, the inspector MUST show them in a read-only reusable-identifier section that is visually distinct from ordinary aliases. If merge submission fails with `error.code='merge_precondition_failed'` and `error.details.reason_code='carry_forward_identifier_collision'`, the same inspector flow MUST display the returned `identifier_class`, `normalized_value`, and `blocking_record_id`.
Profiles: base
Verified by: AC-006, AC-020, AC-023, AC-072, AC-073, AC-074, AC-075, AC-186, AC-187, AC-209, AC-210, AC-231

**REQ-03-272**
When the inspector clears a writable direct-reference scalar, it MUST use the ordinary record patch route `PATCH /api/v1/records/{record_id}` with a direct-write change entry keyed by the affected `field_key` and `value=null`. It MUST NOT use a bespoke action route for that clear, and it MUST NOT implicitly rewrite sibling source-preserving text fields.
Profiles: base
Verified by: AC-315, AC-318

### 16.3 Compromise-assessment surfaces

**REQ-03-250**
Interactive assessment-entry surfaces MUST keep `assessment_state` separate from operational-response actions such as containment, isolation, disablement, credential reset, or monitoring.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-03-251**
Interactive assessment-entry surfaces MUST expose confidence by default as `unset`, `low`, `medium`, or `high`. They MAY additionally expose an exact integer `confidence_score` in the range `0..100` as a secondary control.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-03-252**
When the band-first path is used, the implementation MUST persist canonical default scores of `25` for `low`, `55` for `medium`, and `85` for `high`. `unset` MUST persist `confidence_score=NULL`. A band-first control MUST write `confidence_score`; `confidence_band` remains derived.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-03-253**
Workbook filtering on compromise-assessment surfaces MUST treat `assessment_state` and `confidence_band` as separate fields.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

**REQ-03-254**
The base-profile Assessments view is append-only for semantic assessment fields. Creating a row or equivalent entry MUST append a new assessment record. In-place grid edits to an existing assessment row MUST NOT overwrite `subject_ref`, `subject_type`, `assessment_state`, `confidence_score`, `rationale`, `assessor`, `assessed_at`, or supporting-link semantics; correction or superseding flows MUST append a new assessment record instead.
Profiles: base
Verified by: AC-018, AC-080, AC-081, AC-082, AC-083, AC-084, AC-121, AC-231

### 16.4 Analyst-work coordination surfaces

**REQ-03-255**
Task requests and decisions MUST remain workbook surfaces rather than separate application modules. In the base profile, the required `Task Requests` and `Decisions` surfaces MUST be the contract-backed system views identified canonically by `cartulary.view.task_requests.v1` and `cartulary.view.decisions.v1`. An implementation MAY additionally expose one or more implementation-owned `scope='system'` saved views over either schema as additive convenience presets. Any such saved view is a distinct saved-view object, MUST NOT be required for conformance, MUST NOT replace the canonical public identity of the required base surface, and MUST NOT be the only discoverable, openable, or startup-targetable surface for that schema.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-03-256**
From a selected Timeline, Host, Identity, Evidence, Notes, Task Requests, or Decisions row, the analyst MUST be able to create or link a `task_request`, `decision`, structured coordination artifact, or incident-scoped party reference without leaving the workbook flow when the active surface exposes requester, collector, source, audience, or attendee semantics. When that flow pre-seeds linked-record, decision, support, or party-reference context from the selected record, those preseeded references MUST remain editable context and MUST NOT by themselves satisfy the target surface's minimum create signal.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-278, AC-279

**REQ-03-273**
If the active task surface exposes decision-link clearing, that flow MUST use the ordinary record patch route `PATCH /api/v1/records/{record_id}` with one direct-write change for `task.decision_record_id` set by `value=null`. It MUST remain same-surface and MUST NOT require a bespoke decision-link action route.
Profiles: base
Verified by: AC-315, AC-319

**REQ-03-257**
Routine timeline row creation and editing MUST NOT require task, decision, owner, approver, challenge, or checklist fields on the timeline sheet itself.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-03-258**
Communications logs, handoffs, status reviews, and lesson artifacts MUST remain explicit workbook-native actions. The implementation MAY assist creation from the inspector or a coordination view, but it MUST NOT require creation of those artifacts during ordinary row capture or ordinary grid editing, and it MUST NOT interrupt ordinary grid editing to solicit them.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

**REQ-03-259**
Saved views over task requests, decisions, and coordination artifacts, including implementation-owned `scope='system'` saved views, MUST support owner queues, blocked-work views, due or next-checkpoint views, workstream views, requester, source-party, or audience views where applicable, external-ticket lookup, and no-owner gap detection where applicable. When the implementation exposes findings, investigative queries, or forensic keywords as workbook surfaces, those surfaces MUST use the same workbook-native queue, filter, and grouping model rather than a separate application module. Findings and hypotheses share `cartulary.view.findings.v1`; the current profile defines no separate hypothesis module or `cartulary.view.hypotheses.v1` workbook surface.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231, AC-279, AC-281, AC-282, AC-283, AC-284, AC-285, AC-286, AC-287

**REQ-03-260**
Promoting recurrent coordination or evidence-management fields MUST NOT add them as mandatory Timeline columns or move them into inspector-only JSON payloads.
Profiles: base
Verified by: AC-085, AC-086, AC-087, AC-088, AC-089, AC-090, AC-137, AC-138, AC-139, AC-140, AC-141, AC-142, AC-143, AC-144, AC-145, AC-231

## 17. Authorship and attribution in the UI

**REQ-03-261**
The workbook MUST surface authorship with low friction.
Profiles: base
Verified by: AC-007, AC-231

**REQ-03-262**
At minimum:

- the row must expose last editor and relative update time,
- history must be reachable in one click or one shortcut,
- the system MUST preserve end-to-end attribution across current records, change sets, revisions, links, tags, blob metadata, and evidence metadata.
Profiles: base
Verified by: AC-007, AC-231

## 18. Excel-like feel versus intentional differences

### 18.1 Excel-like behaviors

**REQ-03-263**
The implementation MUST preserve the following spreadsheet-like behaviors:

- tabular grid,
- direct typing,
- paste,
- fill-down,
- keyboard navigation,
- flexible sorting and filtering.
Profiles: base
Verified by: AC-005, AC-043, AC-231

### 18.2 Intentional differences

**REQ-03-264**
The implementation MUST intentionally differ from a raw spreadsheet in the following ways:

- relationship cells render canonical or unresolved chips rather than remaining raw delimited strings forever,
- evidence is attached object state rather than path text inside a cell,
- history and rollback are built in,
- formulas, macros, and merged cells are not part of the model.
Profiles: base
Verified by: AC-090, AC-231

## 19. Interaction invariants

**REQ-03-265**
A conformant implementation MUST preserve all of the following:

1. the primary capture path remains in the grid,
2. same-field concurrent edits never silently overwrite,
3. low-friction row creation survives incomplete information, but zero-field create is reserved to surfaces whose active `view_schema` explicitly allows it,
4. non-Timeline surfaces commit only after their active `view_schema` minimum create signal is satisfied after create-time normalization,
5. a rejected non-Timeline create leaves no partial record row, no misleading projection row, and no misleading live-update event,
6. destructive operations stay explicit and attributed,
7. grouping remains presentation-only,
8. grouped surfaces still mutate underlying rows by `record_id` and `row_version`,
9. auto-resolution stays tightly bounded and reversible,
10. the inspector enriches work but does not replace the grid,
11. unresolved same-field conflict drafts remain client-local until explicit resolution,
12. analyst-work coordination remains workbook-native without adding required coordination fields to the routine timeline hot path.
Profiles: base
Verified by: AC-231, AC-281, AC-282, AC-283, AC-284, AC-285, AC-286, AC-287


## 20. Parties system-view and linking flows

**REQ-03-266**
The Parties surface MUST remain a workbook-native contract-backed system view. It MUST NOT require a new built-in tab in the base profile.
Profiles: base
Verified by: AC-231, AC-277

**REQ-03-267**
Ordinary text entry into requester, collector, source, or audience text fields MUST NOT auto-create or auto-link a `party` record. The base profile does not define a separate Communications Log attendee text field; attendee semantics on that surface are represented by supplemental party-reference collections.
Profiles: base
Verified by: AC-231, AC-279

**REQ-03-268**
Where the active surface exposes requester, collector, source, or audience text/ref pair semantics, the inspector or Parties view MUST support explicit `Create party from text`, `Link existing party`, `Clear party link`, `Clear party text`, and `Clear both` actions. Where the active surface exposes audience or attendee supplemental party-reference collection semantics, the implementation MUST support explicit same-surface link and clear actions for those references without replacing required source-preserving audience text. The implementation MAY expose those actions through any same-surface command set that preserves their distinct behavior and non-destructive defaults.
Profiles: base
Verified by: AC-231, AC-278, AC-279

**REQ-03-269**
`Clear party text` MUST preserve any linked `party_id`. `Clear party link` MUST preserve any source-preserving party text. `Clear both` MUST clear both.
Profiles: base
Verified by: AC-231, AC-278

**REQ-03-274**
`Clear party link` MUST send `PATCH /api/v1/records/{record_id}` with only the corresponding hidden `*_party_id` field changed through direct-write `value=null`. `Clear party text` MUST clear only the text field. `Clear both` MUST send one patch containing both field changes so the resulting round-trip state is explicit and atomic.
Profiles: base
Verified by: AC-315, AC-318

**REQ-03-270**
Task Requests and Evidence MUST remain text-first on the grid. Party linking is same-surface enrichment from the inspector or Parties view and MUST NOT block row creation, row commit, or later text editing.
Profiles: base
Verified by: AC-231, AC-278, AC-279

**REQ-03-271**
Party create, link, unlink, and clear actions MUST keep the main grid or Parties view visible and preserve the originating row context, focus target, and scroll position when the action completes.
Profiles: base
Verified by: AC-231, AC-278, AC-279
