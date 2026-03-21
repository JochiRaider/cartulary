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

The workbook MUST support additional contract-backed system views, including indicator and compromise-assessment surfaces.

Such views MUST remain workbook surfaces rather than separate application modules.

### 2.3 Saved views

Views MUST be saveable and shareable within the incident boundary according to the saved-view scope model.

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

The conflict payload MUST preserve:

- the field key,
- the client-submitted value,
- the persisted server value,
- the client `base_row_version`,
- the current `row_version`.

A same-field conflict MUST remain unresolved until an analyst explicitly chooses a resolution. Same-field edits MUST NOT silently overwrite each other.

### 3.3 Client addressing rules

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

### 4.4 Local pending queue

The client MUST maintain a small local pending-patch queue so that transient network interruptions do not lose typed data.

## 5. Locking policy

Routine inline edits MUST NOT use hard record locks.

Short-lived server-side record locks MAY be used only for rare destructive operations such as:

- merging host or identity records,
- rolling back an existing change,
- restoring a soft-deleted record.

## 6. Record lifecycle

The base record lifecycle MUST support this progression:

`rough capture -> enriched -> linked -> reviewed -> superseded or rolled back`

Normalization MUST add structure without erasing the original observed input.

## 7. Timeline creation workflow

Typing into the blank trailing timeline row MUST create a real record as soon as one non-empty user-entered value exists.

A timeline row MUST be persistable with:

- only one non-empty user-entered value, or
- only an attached screenshot.

At row creation time, the system MUST support:

- nullable `occurred_at`,
- summary text,
- raw mention tokens for host and identity references.

The system MUST NOT block row creation on missing canonical host or identity records.

## 8. Evidence attachment workflow

### 8.1 Two-step upload

Binary evidence attachment MUST use a two-step flow:

1. create a pending blob slot and return an upload target,
2. finalize the evidence attachment after the upload completes.

The system MUST NOT leave fake attached evidence rows for incomplete uploads.

### 8.2 Pending evidence without blob

The evidence model MUST also support requested or pending evidence records with no blob yet attached, followed later by receipt, availability, custody, and optional blob attachment.

### 8.3 Evidence access

Evidence preview MUST open without forcing a full-page navigation away from the grid. A bottom or side preview is acceptable.

## 9. Mention resolution workflow

The inspector MUST support:

- resolving a selected unresolved mention to an existing entity,
- creating a host or identity from a selected mention,
- dismissing a selected mention.

Resolving or creating from a selected mention MUST preserve the raw mention.

Creating from a selected mention MUST create exactly one stub entity by default and resolve only the selected mention unless the user later invokes an explicit bulk action.

## 10. Reviewer history and rollback workflow

### 10.1 Reviewer lens

The history experience MUST remain row-centric.

### 10.2 Minimum history presentation

In the base profile, the history panel for a selected row MUST show:

- actor,
- timestamp,
- operation,
- diff summary expanded to field, link, mention, tag, and evidence-entry units.

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

## 11. Clipboard paste and import workflows

### 11.1 Clipboard paste

Clipboard paste is base-profile functionality.

Pasting TSV or CSV into the timeline sheet MUST create or update multiple rows starting from the selected cell.

Known columns MUST map directly.

Unknown columns MUST be preserved in `raw_capture.import_columns` or equivalent structured raw-capture storage.

Host and identity text from pasted cells MUST follow the same `entity_binding_mode` contract as interactive edits.

A single paste action MUST appear as:

- one visible `change_set`,
- ordered mutation entries,
- one row revision per affected record.

### 11.2 Import Extension Profile

Full XLSX import belongs to the **Import Extension Profile**.

If implemented, the import assistant MUST:

- use the same mapping engine as clipboard and other structured import paths,
- prioritize mappings for timeline, systems or hosts, accounts or identities, indicators, evidence tracker, and VERIS-like summaries when present,
- preserve unknown columns in `raw_capture` or `custom_attrs`,
- determine `mention_origin` versus `entity_origin` from contracts rather than sheet labels alone,
- avoid auto-resolution of imported host or account aliases during ingest.

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
- full XLSX import,
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

The active grouping key MUST be stored as a stable contract value in `saved_views.layout_json.group_by_key`, not as a visible label.

The allowed grouping keys are exactly:

- `timeline.occurred_day`,
- `timeline.recorded_day`,
- `timeline.capture_state`,
- `timeline.has_evidence`,
- `timeline.has_unresolved_mentions`.

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

Saved views MAY persist the default grouping key. They MUST NOT persist another user’s live open or closed state.

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

The implementation MUST preserve the following write-back semantics:

| Timeline column | Read model | Required write-back behavior |
| --- | --- | --- |
| Time | `occurred_at` | update `timeline_events.occurred_at` |
| Summary | `summary` | update `timeline_events.summary` and, when applicable, `details` |
| Hosts | host labels plus unresolved host tokens | if a unique exact normalized alias match qualifies for auto-resolution, insert resolved `entity_mentions` plus `record_links` with `provenance='auto_match'` and `confidence=100`; otherwise insert unresolved `entity_mentions` |
| Identities | identity labels plus unresolved identity tokens | if a unique exact normalized alias match qualifies for auto-resolution, insert resolved `entity_mentions` plus `record_links` with `provenance='auto_match'` and `confidence=100`; otherwise insert unresolved `entity_mentions` |
| Evidence | `evidence_count` | create `object_blob`, `evidence_record`, and `record_link` |
| Tags | `tag_names` | upsert tags and record-tag bindings |

Reads MAY be denormalized. Write-back MUST remain intent-aware and contract-driven.

## 16. Entity, evidence, and inspector surfaces

### 16.1 Entity and evidence sheets

Hosts, Identities, and Evidence sheets MUST remain workbook-shaped peer sheets backed by their own projections.

Canonical fields such as `display_name`, `hostname`, `upn`, and evidence `title` MUST be inline-editable where appropriate.

Alias cells MUST behave like chip editors.

Relationship-derived columns such as linked-event counts MUST be read-only and navigable.

### 16.2 Inspector

The inspector MUST:

- remain a non-blocking drawer rather than the default primary capture surface,
- expose details, relationships, evidence, and history views,
- keep the main grid visible while open,
- support mention resolution, entity creation, evidence inspection, and rollback,
- preserve the distinction between raw mention lineage and current canonical links.

The inspector is the enrichment surface. It MUST NOT be required for the common path of timeline row creation and editing.

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
8. the inspector enriches work but does not replace the grid.
