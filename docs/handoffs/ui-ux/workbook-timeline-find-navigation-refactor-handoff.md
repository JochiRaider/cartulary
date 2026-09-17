# Timeline Find and match navigation

## Execution control

Execution starts on clean `main`, HEAD
`d325b23a7f71aad6674349eddd2c22f67d2b0488`, two commits ahead of
`origin/main`. Root AGENTS.md is the only applicable instruction file. Revalidated
with `git status --short --branch`, `git rev-parse HEAD`, and instruction discovery.
Existing commits are preserved. The user approved the complete implementation
plan and its scope, text, focus, and panel decisions.

| Workstream | Status | Dependency / exit |
| --- | --- | --- |
| TFN-01 Characterization and Find contract | DONE | Core 03 §13.5 and design amendments adopted; Markdown and whitespace checks pass. |
| TFN-02 Bounded matching and semantic navigation | DONE | Focused matching, production presentation and gated navigation tests pass. |
| TFN-03 Presentation, editing, and recovery | DONE | Production composition integrated; focused integration, type and boundary checks pass. |
| TFN-04 Production-browser and spreadsheet evidence | DONE | Production Find, spreadsheet, shared-surface, accessibility, measurement and visual evidence pass. |
| TFN-05 Terminal validation and completed handoff | DONE | Terminal checks, 645/645 frontend units, final Markdown, scope review and completed checklist pass. |

Only the current row may be IN_PROGRESS. Save actual exit evidence and DONE
before beginning its successor. Applicable blocked dependencies prevent dependent
implementation and completion. No digest edits, commits, pushes, deployments,
analyst-data changes, dependencies, new endpoints/operators, persistent search,
Replace, cross-sheet search, or automatic page scanning are included.

## Authority and characterization

Behavior: Core 01 query predicates and stable identities; Core 03 query
continuity, editing, selection, retained work and authorization; Core 04 protected
content; design §§7–8/14 presentation and accessibility. Domain supplies vocabulary
and owner navigation. The digest localized read order and query-continuation,
grid-edit/autosave, range-selection, clipboard, column-sizing and recovery-navigation
handoffs were inspected during planning. They and the NLSpec research essay are
advisory/evidence, not additional behavioral authority or fresh passes.

Source placement belongs to frontend source/import manifests; verification belongs
independently to contracts/verification and authored test families. Executable
consumers must not read Markdown. React 19.2.5 and Adapter-owned RDG 7.0.0-beta.59
are the current authored dependencies.

| Responsibility | Existing owner / reuse |
| --- | --- |
| Accepted query/window | WorkbookQueryBrowser; three pages of 100, explicit continuation, staged admission. |
| Committed values | Timeline rawRow/committedValues and committed-record capability; editor drafts remain separate. |
| Visible columns | Timeline presentation applies Workbook semantic layout, then synchronizes visible columns. |
| Group/collapse/order | Grid Adapter semantic presentation, including offscreen membership. |
| Virtualized focus | Adapter semanticFocusRequest and GridHandle; Workbook continuity supplies semantic anchors. |
| Scalar editor settlement | Adapter ActiveEditorSession.requestCommit and source mutation outcome. |
| Collection/inspector authoring | Timeline draft registry and source-owned save callbacks. |
| Completed selection | Adapter GridCellRange; distinct from inspector and bulk selection. |
| Shortcuts | Pure Workbook application shortcut policy plus focus-scoped producer handlers. |
| Overlay/recovery | Existing external-action focus borrowing and shared recovery attachment coordination. |

## Gap ledger

| Gap / classification | Remediation and affected areas | Rationale / durable benefit | Compatibility / risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 New product behavior: local Find has no adopted boundary | Adopt Core 03/design scope, text, focus, chrome and lifecycle contracts before implementation. | Explicit capability without redefining full_text. | No HTTP/data migration; risk of scope confusion. | All specified acceptance cases map to adopted owners; UI never claims incident completeness. |
| G2 New capability: source-readable text and presented membership | One Workbook Find owner, Timeline text capability, neutral Adapter membership projection. | Avoid DOM scraping, duplicate query/selection stores and vendor leakage. | Internal TS additions only; risk of exposing local/hidden text. | Offscreen matching includes exactly eligible committed cells in presentation order. |
| G3 Confirmed weakness: public requestFocus drops AbortSignal | Forward cancellation options; add editor-gated semantic navigation using existing session. | Obsolete destinations cannot regain focus. | Shared Adapter consumer regression boundary; no change to authoritative settlement. | Abort/supersession before and after delayed mount prevents focus; one commit per departure. |
| G4 New interaction: Find focus borrowing and navigation | Bind existing scalar/collection/inspector retention and settlement to Find; scoped keyboard and panel lifecycle. | Preserve exact drafts and one execution owner. | No draft persistence or new mutation path; risk of blur writes/late focus. | Opening/typing/closing performs no writes; rejection retains exact draft; only current admitted target moves. |
| G5 Hypothesis: long content/chrome may affect responsiveness | Cooperative bounded scans, existing chrome geometry, production browser/measurement evidence. | Stable input and virtualization without speculative indexing. | Existing width/density/measurement obligations remain; hypothesis until measured. | Full retained window stays virtualized, responsive and exact; supported widths and existing budgets pass. |

## Evidence log

- Planning discovery: `make help`, `make help-all`, owner task guides for
  web.workbook, package.grid_adapter, module.workbook, web.design, package.ui,
  web.architecture, web.networkflow and module.savedviews passed. These are routing
  observations, not product acceptance.
- Execution baseline: clean working tree and exact baseline HEAD confirmed.

## Compatibility, rollback and next action

This is a Timeline-only runtime interaction with neutral Adapter mechanics.
Another adopted grid can supply committed readable text, ordered semantic
membership and navigation without another Find state owner. Exhaustive server
search requires separate adoption and cannot be approximated by page scanning.
Rollback must remove the coherent owner/projection/implementation/test changes,
preserving existing commits and accepted editor writes. No data rollback or
saved-layout migration is required.

### TFN-01 exit

Adopted Core 03 §13.5 and the application-shortcut matrix; design §§7.5, 8.3,
8.6 and 14.2 now define the Find entry, scoped panel, orthogonal match cue,
focus borrowing, collapse/close behavior and announcements. Domain §13.11 adds
owner navigation without defining independent behavior. Core 01 full_text and
Core 04 authorization remain unchanged. The approved scope and interaction
choices are explicit acceptance criteria; no unresolved owner contradiction remains.

- `make lint-markdown`: PASS, run root
  `.cartulary/test-results/20260917T211610Z-p730872`, summary
  `adhoc/lint-markdown/tool-run-summary.json`.
- `git diff --check`: PASS.
- Review: only the controlling handoff, Core 03, design and domain changed.
  No executable consumer was changed during contract adoption.

Next action: begin TFN-02 matching and semantic navigation.

### TFN-02 exit

Implemented the bounded, cooperative Workbook Find controller, Timeline committed
readable-text fragments, subscribable Adapter presentation membership and
editor-gated semantic cell navigation. The public focus wrapper now forwards
cancellation. Source rows remain owned by Timeline; Find retains semantic matches
and the current source reference only. Presentation tracks collapsed membership
without requiring mounted cells. Tests cover normalization, ordering, per-cell
counts, collection summaries, committed versus local values, delayed acceptance,
rejection, supersession, and presentation changes.

- `make format`: PASS (`20260917T212836Z-p743456`).
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.find_matching,web.workbook.regression.timeline_find_text`:
  PASS, 3/3 units (`20260917T212714Z-p741036`).
- `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.semantic_cell_navigation,package.grid_adapter.regression.semantic_focus_requests,package.grid_adapter.regression.index_suite_d804ef9789`:
  PASS, 4/4 units (`20260917T212856Z-p747818`).
- `make frontend-typecheck`: PASS, 2/2 units (`20260917T212856Z-p747900`).
- Run roots are under `.cartulary/test-results/`; each includes `run-summary.json`.
  Initial routing preflight rejected unsorted authored rows/titles; sorted and
  reran. A delayed-settlement test exposed mutable capture of the navigation
  authority snapshot; copying its scalar fields fixed the failure. The failed
  run (`20260917T212714Z-p741050`) was inspected with `make explain-run` before rerun.

G2/G3 focused criteria pass. Production browser scrolling/focus and source
integration remain TFN-03/04 obligations. Rollback removes these additive internal
capabilities and routing together; it does not affect authoritative writes.
Next action: begin TFN-03 production controls and editing integration.

### TFN-03 exit

Timeline composition now supplies one Find binding over accepted rows and the
Adapter presentation port. The existing view bar receives a narrow control
projection before Inspector. The floating, labelled panel preserves bar height,
uses compact actions at compact widths and collapses after admitted movement.
Opening borrows focus using the existing external-action boundary; the retained
draft registry supplies editor identity without copying draft text. Scalar grid
sessions retain their gate; collection/inspector departures use source-owned save
callbacks and join the same pending revision. Recordless authoring cannot trigger
creation through Find. Close restores the semantic target/borrowed editor and
eligible row/root fallback. New interactions cancel destinations. Authority
retirement clears runtime Find state synchronously; ordinary stale rows remain
searchable. Recovery activation collapses the panel without restoring focus.

The Find marker is orthogonal to primary cell state, with a distinct accessible
current-match label. Source guides describe execution and extension boundaries.
Authored design presentation JSON/schema and its generator now project scope
copy, live-region policy and successful-navigation collapse. Generated derivatives
were produced through Make. No duplicated navigation executor was introduced;
Find uses the shared editor-gated command, and recovery keeps requestFocus.

- `make test-slice OWNER=web.workbook` with the six Find/composition rows:
  PASS 7/7 (`20260917T214721Z-p779905`).
- `make test-slice OWNER=module.workbook` shortcut row: PASS 2/2
  (`20260917T214350Z-p762325`).
- `make test-slice OWNER=module.timeline` scalar-editor row: PASS 2/2
  (`20260917T214350Z-p762349`), including external focus borrowing.
- Adapter navigation/focus/production binding slices: PASS 4/4
  (`20260917T214722Z-p779970`); the updated production state-composition test:
  PASS 2/2 (`20260917T214903Z-p787715`).
- `package.ui.frontend_unit.design_presentation_projection_8e15b85b40`:
  PASS 2/2 (`20260917T214903Z-p787730`).
- `make frontend-import-boundary-check`: PASS 2/2
  (`20260917T214537Z-p770691`). The prior failed boundary check identified a
  sibling-controller type import; shared source settlement types now live in
  timelineControllerPorts instead.
- `make frontend-typecheck`: PASS 2/2 (`20260917T214956Z-p789798`). Earlier
  checks caught two test-fixture types and a missing public semantic equality
  export; corrected before rerun.
- `make lint-biome`: PASS 2/2 (`20260917T214903Z-p787953`), after correcting
  exhaustive source memo dependencies and fixture assertions.
- `make generate`: PASS (`20260917T214641Z-p772255`); `git diff --check`: PASS.

These runs are under `.cartulary/test-results/`, with `run-summary.json` (generate
uses `generate/tool-run-summary.json`). G4 focused criteria pass. Browser geometry,
virtualization and real-authority races remain TFN-04 obligations, not inferred
from unit handles. Rollback remains a coherent additive interaction rollback.
Next action: begin TFN-04 production browser and spreadsheet evidence.

### TFN-04 investigation and repairs

Production browser evidence exposed two integration weaknesses and several fixture
errors. Hidden Tags and a saved layout retaining its hidden Synopsis required
explicit fixture setup, not wider Find scope. Collection settlement's deferred
row-inspect restoration could override a successfully admitted match; settlement
callbacks now retain scroll continuity and own their explicit destination. Viewer
recovery leaves mutation replay paused by design, so Find now consumes the
collaboration owner's narrow read-authorization projection. Read confirmation
resumes after accepted surface refresh independently of write admission.

| Gap / classification | Remediation / affected areas | Rationale / benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G6 Confirmed integration weakness: write pause used as read authority | Collaboration read-authorization projection; Timeline Find subscription. | Recovered viewers can search without enabling replay. | Internal capability only; requires immediate retirement and viewer browser confirmation. | Session retirement clears protected state and recovered viewer starts empty with reads enabled. |
| G7 Confirmed integration weakness: collection continuity competes with explicit destination | Source save callback uses scroll-only continuity when caller owns settlement navigation. | One destination owner after authoritative save. | Existing Enter/Tab caller also benefits; regression evidence required. | One collection write; late settlement cannot steal focus from admitted match or newer interaction. |

Initial browser failures are retained at `20260917T215415Z-p799047`,
`20260917T215925Z-p839693`, `20260917T220239Z-p877061`,
`20260917T220906Z-p989896`, and `20260917T221426Z-p1125745` under
`.cartulary/test-results/`. Current artifacts, including traces, were inspected.
The reopened-panel redundant scan was removed, pasted multiline input uses a
one-row textarea, and geometry review corrected its border-box sizing. No
assertion, selector, screenshot mask, or tolerance was weakened.

Baseline isolation used a detached worktree at `/tmp/cartulary-tfn-baseline`
with unchanged HEAD and a Make-owned dependency install. Two failures reproduce
exactly there: Network Analysis accepted-inspector differs by 396 focus/range
pixels (`20260917T222117Z-p1363733`), and the ancillary Evidence editor
accessibility case clips at **390px**, below the 768px supported minimum
(`20260917T222118Z-p1363968`, ratio 0.9681817889). Current-worktree equivalents
are `20260917T221833Z-p1251272` and `20260917T221811Z-p1165045`.
The latter is an existing below-minimum diagnostic, outside this interaction's
supported-width acceptance; no unrelated Evidence layout repair or weakened
assertion is included. Supported-width shell, grid, column-control, and Find
accessibility remain applicable.

The first ordinary visual reconciliation at `20260917T220923Z-p1046107`
accounts for 252 active captures/goldens, zero orphans, zero missing goldens,
zero ambiguous mappings, and zero unresolved fixtures. Refresh trigger: adopted
Timeline Find entry changes the exact view bar and its visible background in
existing fixtures. The pre-existing Network Analysis golden is also eligible
for narrow maintenance of the explicitly required shared-adapter regression
boundary: baseline/current render identically, the focus/range state is already
adopted, and its current functional row passes. No viewport, zoom, masks,
normalization, screenshot scope, tolerance or selector changes are made.

The strengthened source-editor browser case uses an offscreen row after a
collection save. It exposed deferred scroll restoration even after focus reached
the correct cell (`20260917T222519Z-p1508071`). The existing Timeline continuity
owner now provides interruption of its current restoration, invoked only after
the editor gate admits the semantic destination. This cancels focus/scroll work,
not writes. Completed-range evidence also verifies that opening/typing/Close
preserve range and bulk selection, while explicit navigation replaces the range
even when its destination was already the endpoint.


Full baseline visual isolation (`/tmp/cartulary-tfn-baseline/.cartulary/test-results/20260917T223038Z-p1623265`)
found seven stale captures: Network accepted inspector, Evidence affordance states,
ordinary reference authoring, record relationship mentions, and inspector History,
rollback preview and public error. The three captures without Timeline chrome
are byte-identical to the current renderer. These are existing focus/range states
on the explicitly required shared Adapter regression surfaces, not new Find
behavior. Their narrow maintenance accompanies that regression verification;
no unrelated product behavior or visual fixture was changed. The four Timeline
captures also include the adopted Find entry. Baseline functional and current
regression evidence distinguish this stale-image issue from a rendering regression.

### Visual refresh review

`make browser-e2e-visual-update` passed 12/12 units at
`20260917T222405Z-p1473106`. Reviewed every one of the 64 changed PNGs individually,
including the seven baseline differences above. Controls remain within one bar;
required query allocation, compact icon actions, focus, pending/conflict cues,
recovery and inspector regions remain legible. Find panel attachments were also
reviewed at 1440, 1280, 1024 and 768 CSS pixels, 200% zoom and increased text
spacing. The compact panel fits and closes before the destination is focused.
No masks, dimensions, fonts, renderer profile, normalization, tolerances or
selectors changed. The manifest was generated by the update target.

The following inventory comes from the run's `frontend-visual-reconciliation.json`,
not filename inference. Paths are under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. A dash means no fixture ID is
declared by the reconciliation owner; catalog row and capture remain explicit.

| Changed golden | Catalog row | Fixture IDs |
| --- | --- | --- |
| account-menu-controls-narrow-linux.png | web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc | — |
| account-menu-workbook-root-linux.png | web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc | — |
| collaboration-conflict-resolver-linux.png | module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1 | visual.fixture.same_field_conflict |
| collaboration-grid-blocked-conflict-linux.png | module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc | — |
| collaboration-grid-conflict-resolver-compact-linux.png | module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c | — |
| collaboration-grid-conflict-resolver-linux.png | module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c | — |
| collaboration-grid-conflict-resolver-narrow-linux.png | module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c | — |
| collaboration-presence-markers-linux.png | module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1 | visual.fixture.presence_overflow |
| coordination-comm-log-authoring-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-handoff-authoring-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-lesson-authoring-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-recovery-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-recovery-narrow-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-source-narrow-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| coordination-status-review-authoring-linux.png | module.workbook.visual.coordination_create_authoring_recovery | visual.fixture.contextual_coordination_creation |
| entity-mention-chip-states-linux.png | module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7 | visual.fixture.mention_chip_state_matrix |
| evidence-affordance-states-linux.png | module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4 | visual.fixture.evidence_affordance |
| incident-directory-compact-desktop-workbook-shell-linux.png | module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0 | visual.fixture.compact_desktop_workbook_shell |
| incident-directory-default-timeline-workbook-shell-linux.png | module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0 | visual.fixture.default_timeline_workbook_shell |
| incident-directory-narrow-desktop-workbook-shell-linux.png | module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0 | visual.fixture.narrow_desktop_workbook_shell |
| indicator-observation-authoring-linux.png | module.workbook.visual.indicator_observations_authoring | visual.fixture.indicator_observations_authoring |
| indicator-observation-authoring-narrow-linux.png | module.workbook.visual.indicator_observations_authoring | visual.fixture.indicator_observations_authoring |
| linked-note-authoring-linux.png | module.workbook.visual.note_create_authoring_recovery | — |
| linked-note-authoring-narrow-linux.png | module.workbook.visual.note_create_authoring_recovery | — |
| linked-note-recovery-linux.png | module.workbook.visual.note_create_authoring_recovery | — |
| linked-note-recovery-narrow-linux.png | module.workbook.visual.note_create_authoring_recovery | — |
| linked-note-source-narrow-linux.png | module.workbook.visual.note_create_authoring_recovery | — |
| network-flow-analysis-accepted-inspector-linux.png | module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6 | visual.fixture.claimed_network_analysis_workspace_states |
| ordinary-reference-authoring-linux.png | module.workbook.visual.ordinary_create_authoring_recovery | — |
| record-relationships-mention-chips-linux.png | module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7 | — |
| timeline-grid-timeline-default-linux.png | module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206 | — |
| timeline-mutation-pending-replay-status-linux.png | module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67 | — |
| timeline-mutation-transaction-recovery-panel-compact-linux.png | module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67 | — |
| timeline-mutation-transaction-recovery-panel-linux.png | module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67 | — |
| timeline-mutation-transaction-recovery-panel-narrow-linux.png | module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67 | — |
| timeline-related-evidence-authoring-linux.png | module.workbook.visual.timeline_related_evidence | — |
| timeline-related-evidence-authoring-narrow-linux.png | module.workbook.visual.timeline_related_evidence | — |
| timeline-related-evidence-partial-linux.png | module.workbook.visual.timeline_related_evidence | — |
| timeline-related-evidence-partial-narrow-linux.png | module.workbook.visual.timeline_related_evidence | — |
| timeline-related-evidence-party-narrow-linux.png | module.workbook.visual.timeline_related_evidence | — |
| timeline-supersession-accepted-linux.png | module.workbook.visual.timeline_capture_actions | — |
| timeline-supersession-authoring-linux.png | module.workbook.visual.timeline_capture_actions | — |
| timeline-supersession-review-linux.png | module.workbook.visual.timeline_capture_actions | — |
| timeline-supersession-review-narrow-linux.png | module.workbook.visual.timeline_capture_actions | — |
| workbook-inspector-compact-actions-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.inspector_compact_actions |
| workbook-inspector-history-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.base_inspector |
| workbook-inspector-narrow-technical-details-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.inspector_narrow_technical_details |
| workbook-inspector-public-error-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.base_inspector |
| workbook-inspector-relationships-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.base_inspector |
| workbook-inspector-rollback-preview-linux.png | module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea | visual.fixture.base_inspector |
| workbook-query-empty-text-spacing-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | visual.fixture.empty_successful_query |
| workbook-query-empty-zoom-200-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | visual.fixture.empty_successful_query |
| workbook-query-saved-view-query-controls-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | visual.fixture.saved_view_query_controls_and_grouped_result |
| workbook-view-bar-filter-editing-overflow-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-long-columns-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-maximum-pressure-base-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-maximum-pressure-compact-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-maximum-pressure-narrow-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-ordered-maximum-sort-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-saved-view-actions-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-saved-view-clean-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-saved-view-modified-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-text-spacing-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |
| workbook-view-bar-zoom-200-linux.png | module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc | — |


### TFN-04 production evidence

The nine owner-routed Find browser cases pass through production Timeline and
Grid Adapter. Network observation asserts no Find-issued query/page requests,
query/layout mutations or creation; only source-owned departure saves are
permitted. The 300-row scenario verifies exact final counts, bounded mounted
rows, input/Close response and obsolete-scan cancellation with long multiline
content. Read-only, hidden/collapsed, summary-only, retained draft, pending value,
creation pin, live update/deletion, query/saved-view replacement, window eviction,
recovery and authority cases pass. Offscreen collection/inspector acceptance
asserts both semantic focus and full viewport visibility. Existing spreadsheet
and shared-surface scenarios remain production browser evidence.

Exact successful selections follow. Run roots are under `.cartulary/test-results/`;
each has `run-summary.json` and `run-manifest.json`. Browser roots also retain
per-group `stdout.log`, `playwright-report.json`, attachments and traces under
`browser-e2e-*/browser-groups/`. Measurement roots retain their measurement artifacts.
Units include harness dependencies; they are not a count of behavioral claims.

- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.find_accessible_geometry,module.workbook.browser.find_authority,module.workbook.browser.find_edit_departure,module.workbook.browser.find_full_window,module.workbook.browser.find_literal,module.workbook.browser.find_query_creation,module.workbook.browser.find_scope_lifecycle,module.workbook.browser.find_source_editors,module.workbook.browser.find_virtualized`: PASS 11/11 units, `20260917T223633Z-p1757957`.

- `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.spreadsheet_keyboard,module.timeline.browser.spreadsheet_pointer,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_rejection,module.timeline.browser.spreadsheet_virtualization,module.timeline.browser.spreadsheet_refresh,module.timeline.browser.clipboard_fidelity,module.timeline.browser.range_gestures,module.timeline.browser.range_departure,module.timeline.browser.range_cancellation,module.timeline.browser.range_scrolling,module.timeline.browser.range_presentation,module.timeline.browser.range_rejection_query,module.timeline.browser.range_live_membership,module.timeline.browser.range_authority,module.timeline.browser.range_accessibility`: PASS 13/13 units, `20260917T223455Z-p1659119`.

- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.column_sizing_gestures_saved,module.workbook.browser.column_sizing_surfaces,module.workbook.browser.column_sizing_viewport,module.workbook.browser.grid_autosave_exact_replay,module.workbook.browser.grid_autosave_surface_matrix,module.workbook.browser.grid_autosave_timestamp_retention,module.workbook.browser.grid_autosave_refresh,module.workbook.browser.grid_autosave_inspector`: PASS 13/13 units, `20260917T220907Z-p990289`.

- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.verify_claimed_network_analysis_discovery_import_f977242343`: PASS 11/11 units, `20260917T220924Z-p1047919`.

- `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6,module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13,module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3,module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95`: PASS 20/20 units, `20260917T221958Z-p1298116`.

- `make service-backed-test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.accessibility.verify_grid_cells_editors_group_rows_active_cell_3156bd379d`: PASS 11/11 units, `20260917T221812Z-p1165282`.

- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.column_sizing,module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159,module.workbook.accessibility.verify_grid_navigation_edit_entry_exit_paste_fee_3de2b0afac`: PASS 13/13 units, `20260917T223456Z-p1659341`.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.find_matching,web.workbook.regression.find_controls,web.workbook.regression.timeline_find_integration,web.workbook.regression.timeline_find_text`: PASS 5/5 units, `20260917T223523Z-p1741861`.

- `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.workbookcollaborationcoordinator_suite_514be174a8`: PASS 2/2 units, `20260917T221716Z-p1163557`.

- `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.semantic_cell_navigation,package.grid_adapter.regression.index_suite_d804ef9789`: PASS 3/3 units, `20260917T222121Z-p1367334`.


Earlier complete Find pass: `20260917T221957Z-p1297875`; strengthened
collection/inspector, range and 768px pass: `20260917T222723Z-p1548385`.
The final nine-case run above includes those stronger assertions. Measurement
passed twice, most recently `20260917T221958Z-p1298116`; the subsequent continuity
interruption executes only for admitted Find movement and the final Find browser
run verifies its focused cell remains unobscured. Other-surface autosave and
column sizing plus Network Analysis regression rows pass.

Limit: the supplemental Evidence 390px diagnostic remains an independently
reproduced baseline failure below the adopted 768px supported minimum. It is
N/A to Find supported-width acceptance; the applicable accessibility rows above
pass without changing its assertion. No historical limitation was assumed.


Additional exit checks:

- `make browser-e2e-visual`: PASS 12/12 twice after golden refresh,
  `20260917T222954Z-p1586634` and `20260917T223510Z-p1712567`; fresh ordinary
  renderer comparisons, no update flag. Each reconciles 252 active captures.
- `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser_support.timeline_keyboard_fill_down_preserves_the_select_66fb142159`:
  PASS 11/11, `20260917T223842Z-p1792202`, including real keyboard fill-down
  and preservation of its selected range/current endpoint.
- Final authority review clears origin/last-admitted semantic anchors as well
  as search text/results on retirement. The integration assertion rejects
  restoration to a retired origin: PASS 2/2, `20260917T224023Z-p1854157`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.find_authority`:
  PASS 11/11, `20260917T224003Z-p1824929`, after that retirement refinement.

- `make service-backed-test-slice OWNER=web.design ROWS=web.design.browser.workbook_responsive_frame_geometry_e78e1d1ac3,web.design.accessibility.verify_global_accessibility_matrix_for_keyboard_90088a3666`:
  PASS 13/13, `20260917T224059Z-p1857591`.

- `make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.accessibility.verify_sort_filter_group_saved_view_menu_active_8b502b579d,module.savedviews.browser.verify_saved_view_create_update_select_default_u_ce0bb1647d,module.savedviews.browser.verify_saved_views_appear_only_under_the_active_69995294db`:
  PASS 15/15, `20260917T224059Z-p1857606`.


### TFN-04 exit

PASS: scoped matching, virtualized semantic reveal/focus, lifecycle and editor
settlement, accessibility/geometry, full-window response, no-extra-read/write
observations, spreadsheet/fill/clipboard/range, saved-view/column fidelity,
other-surface/Network Analysis regressions and both ordinary visual passes.
G5's performance/geometry hypothesis is resolved by current production evidence;
G6/G7 repairs satisfy their binary criteria. No applicable blocker remains.
Rollback remains coherent removal of Find and its projections/tests, including
its authority projection and explicit continuity interruption if unused. Existing
accepted editor writes and baseline commits remain untouched.
Next action: TFN-05 finalizer, terminal verification and completed acceptance.


### TFN-05 execution

Re-ran Make discovery and current task guides for web.workbook,
package.grid_adapter, module.workbook, web.collaboration, package.ui, web.design
and web.architecture. Required narrow owner and service-backed evidence is
recorded above, including saved-view and Network boundaries. No successful full
warm-check root exists for this task: `make agent-finalize` runs with RESULTS_DIR
unset; retained-run maintenance is intentionally skipped. Its output changes
will be reviewed before broader terminal checks.


`make agent-finalize` passed 1/1 at `20260917T224228Z-p1925551`.
Reviewed `unit-artifacts/finalize-summary.json`: generated structure unchanged,
zero updated files, no rollback needed, schema/catalog/tier checks pass.
Retained-run evidence/performance/scheduler maintenance is skipped explicitly
because RESULTS_DIR was not provided. There were no unrelated finalizer edits
to isolate.

### Acceptance and gap closure

| Requirement / gap | Assessment | Current evidence / remaining limitation |
| --- | --- | --- |
| Adopted bounded contract, G1 | PASS | Core 03 §13.5 and design amendments precede implementation; Core 01 full_text unchanged. |
| Readable committed text and eligible semantic membership, G2 | PASS | Find/text unit slices and all nine production Find cases; virtualized rows/columns, collection overflow, creation pins and drafts excluded correctly. |
| Cancellation-forwarding and one editor-departure gate, G3 | PASS | Adapter production/unit tests, delayed scalar/source acceptance, rejection, Close and supersession browser evidence. |
| Focus borrowing, lifecycle and separate selection, G4 | PASS | Opening produces no writes; exact drafts retained; current range replaced only by explicit movement; bulk/inspector targets preserved. |
| Bounded response and unchanged chrome, G5 | PASS | Full 300-row long-content case, two Timeline measurement passes, supported widths/zoom/spacing, both ordinary visual passes. |
| Read authority independent of replay, G6 | PASS | Collaboration unit assertion plus viewer suspend/recover production case; term/results/origin retire and viewer reads resume without replay. |
| One admitted focus/scroll destination, G7 | PASS | Offscreen collection/inspector case asserts exact write count, semantic focus and viewport visibility; spreadsheet/range rerun passes. |
| No query/page/filter/layout mutation or Find creation | PASS | Browser request observation and saved-state snapshots; existing departure writes are source-owned. |
| Spreadsheet and shared Adapter regression boundaries | PASS | Direct typing/single-click, Enter/Tab, native selection/clipboard, range, real fill-down, autosave, sizing/reset, saved views, other surfaces and Network rows. |
| Search text/content retention and logging | PASS | Runtime owner retains only current source references and match anchors; synchronous retirement; no storage/logging path introduced. |
| Source ownership, typed projections and routing | PASS | Authored source/test families and design input/schema updated; derivatives generated through Make; no Markdown consumer introduced. |
| Test Adapter public export parity, G8 | PASS | Existing identity helpers re-exported; representative legacy slice and final 645/645 frontend units pass. |
| Panel layout ownership, G9 | PASS | Existing popover cap reused; layout/import boundary and final production geometry pass. |
| Evidence 390px supplemental diagnostic | N/A | Confirmed original-HEAD failure below adopted 768px minimum, outside Find supported-width scope. Existing assertion remains unchanged. |
| Exhaustive/server/cross-sheet search, Replace and persistence | N/A | Explicitly excluded; Core 01 query/server and persistence owners unchanged. |
| Database/API/dependency/release migration | N/A | Internal TS capabilities and presentation/test projections only; no backend/schema/dependency change. |

G1–G9 binary criteria are satisfied. The initial new-product rows remain product
expansion, not retroactive conformance defects. G3, G6 and G7 are confirmed narrow
weaknesses resolved in their respective owners; G5's hypothesis is resolved by
measurement. No unresolved applicable product risk or blocked dependency remains.
The recorded below-minimum diagnostic and absence of exhaustive incident search
are explicit limits, not concealed acceptance passes.

### Implementation inventory and retired paths

The cohesive owner is `workbook/find/`; Timeline owns source text and settlement;
Grid Adapter owns semantic membership, editor gating, vendor coordinates and DOM
reveal/focus. Shared chrome receives an optional capability and other surfaces
do not enable Find. Existing retained draft/selection/query owners remain in
place. Replaced paths are the cancellation-dropping requestFocus wrapper,
write-pause-as-read-authority assumption and competing post-collection restoration
when a source settlement caller already owns departure focus. Recovery retains
its existing restoration command and does not acquire a commit requirement.

Changed non-image files are listed below. The 64 changed images and their exact
catalog mapping are inventoried in the visual review above. Generated TS/browser/
topology/visual outputs follow their authored inputs, with no hand edits.

- `apps/web/src/workbook/README.md`
- `apps/web/src/workbook/collaboration/README.md`
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts`
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts`
- `apps/web/src/workbook/components/WorkbookViewBar.tsx`
- `apps/web/src/workbook/policies/workbookApplicationShortcuts.test.ts`
- `apps/web/src/workbook/policies/workbookApplicationShortcuts.ts`
- `apps/web/src/workbook/timeline/components/TimelineScalarEditor.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineScalarEditor.tsx`
- `apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts`
- `apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.ts`
- `apps/web/src/workbook/timeline/hooks/README.md`
- `apps/web/src/workbook/timeline/hooks/useTimelineMutationCommands.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineViewportContinuityController.ts`
- `apps/web/src/workbook/timeline/models/timelineControllerPorts.ts`
- `apps/web/src/workbook/timeline/presentation/README.md`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookViewBarRegion.tsx`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `contracts/design/presentation.v1.json`
- `docs/design.md`
- `docs/domain.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `packages/grid-adapter/README.md`
- `packages/grid-adapter/src/SemanticDataGrid.tsx`
- `packages/grid-adapter/src/core.ts`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/grid-adapter/src/index.tsx`
- `packages/grid-adapter/src/semanticState.ts`
- `packages/grid-adapter/src/styles.css`
- `packages/grid-adapter/src/test-support.tsx`
- `packages/ui-contracts/src/design-presentation.test.ts`
- `packages/ui-contracts/src/generated/design-presentation.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/harness/generated-artifacts/design-presentation/design-presentation.mjs`
- `tools/schemas/cartulary.design_presentation.v1.schema.json`
- `tools/test_families/module.workbook.json`
- `tools/test_families/package.grid_adapter.json`
- `tools/test_families/web.workbook.json`
- `apps/web/e2e/timeline-find.spec.ts`
- `apps/web/src/workbook/find/README.md`
- `apps/web/src/workbook/find/WorkbookFindControl.test.tsx`
- `apps/web/src/workbook/find/WorkbookFindControl.tsx`
- `apps/web/src/workbook/find/WorkbookFindController.test.ts`
- `apps/web/src/workbook/find/WorkbookFindController.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineFind.ts`
- `apps/web/src/workbook/timeline/models/timelineFindText.test.ts`
- `apps/web/src/workbook/timeline/models/timelineFindText.ts`
- `apps/web/src/workbook/timeline/useTimelineFind.test.tsx`
- `docs/handoffs/ui-ux/workbook-timeline-find-navigation-refactor-handoff.md`
- `packages/grid-adapter/src/semanticCellNavigation.test.ts`
- `packages/grid-adapter/src/semanticCellNavigation.ts`
- `packages/grid-adapter/src/semanticPresentationPort.ts`


### Terminal verification investigation

The first full frontend unit run (`20260917T224306Z-p1929996`) exposed two
integration gaps not covered by the focused slices: the legacy test Adapter's
module mock did not export the public semantic identity helper, and the new
panel's maximum height duplicated viewport subtraction outside the existing
popover style boundary. Neither failure is classified as unrelated.

| Gap / classification | Remediation / affected areas | Rationale / benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G8 Confirmed integration gap: test-support export parity | Re-export existing identity helpers from Grid Adapter test-support. | Existing Workbook unit composition can consume the public semantic API without a duplicate algorithm. | Test-only additive exports; no fake Find navigation or production evidence substituted. | Legacy Timeline composition slice and complete frontend unit run pass. |
| G9 Confirmed integration gap: panel viewport arithmetic | Consume the existing view-bar popover maximum-height and scrolling style. | Keeps presentation bounds with the established style owner. | No bar/golden layout changes; supported-width panel case must pass again. | Layout boundary, import boundary and production Find geometry pass. |

The production panel now uses the existing popover height cap. The two ordinary
visual passes remain applicable to unchanged closed-panel chrome; the affected
open-panel geometry case is rerun. Test-support gains no Find executor.


G9 validation: layout policy PASS 2/2 (`20260917T224641Z-p2011078`),
import boundary PASS 2/2 (`20260917T224642Z-p2013482`), production Find geometry
PASS 11/11 (`20260917T224620Z-p1993277`). The final compact/zoom panel images were
reviewed again. G8's representative legacy Timeline composition slice passes
2/2 (`20260917T224619Z-p1992595`). Full frontend verification is rerunning after
the two repairs. The failed full run completed 624/645 units; all 21 failed rows
were traced to the missing mock export or layout arithmetic via current
`vitest-failure-details.json` and `make explain-run`.

### Terminal commands and artifacts

| Command | Result | Run root beneath `.cartulary/test-results/` |
| --- | --- | --- |
| `make agent-finalize` (RESULTS_DIR unset) | PASS 1/1; zero generated changes; retained-run maintenance skipped | `20260917T224228Z-p1925551` |
| `make frontend-typecheck` | PASS 2/2 | `20260917T224306Z-p1929940` |
| `make lint-biome` | PASS 2/2 | `20260917T224306Z-p1930103` |
| `make lint-scripts` | PASS 2/2 | `20260917T224306Z-p1930237` |
| `make frontend-import-boundary-check` | PASS 2/2 after final import | `20260917T224642Z-p2013482` |
| `make json-shape-check` | PASS 3/3 | `20260917T224306Z-p1929791` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260917T224306Z-p1929850` |
| `make generate-drift` | PASS 4/4 | `20260917T224313Z-p1933770` |
| `make lint-markdown` | PASS; rerun follows final handoff text | `20260917T224306Z-p1930312` |

Graph targets retain `run-summary.json`; Markdown retains
`adhoc/lint-markdown/tool-run-summary.json`. Browser/accessibility/measurement/
visual owner selections and artifacts are recorded in TFN-04. No broad full
`make check`, release gate, deployment, vulnerability, migration or unrelated
backend suite is claimed: this change's executable boundary is frontend/Adapter
and presentation/routing generation, with real isolated service-backed browser
evidence already run. Generated drift covers affected derivatives. Retained-run
performance maintenance is skipped, not substituted with partial browser runs.

### Compatibility, rollback and extension review

No endpoint, query operator, server full_text reinterpretation, database,
dependency, lockfile, saved-layout or persistent-search migration is introduced.
Internal TypeScript capabilities and authored design/test projections change.
Rollback removes the Find owner, Timeline binding/chrome, semantic navigation and
presentation additions, authority read projection and associated documentation/
projections/tests together; retain the cancellation-forwarding fix if recovery
continues to rely on it. Accepted editor writes are authoritative and are not
rolled back. The existing two commits ahead of origin/main remain intact.

Another authorized grid can provide the same current source reference, readable
committed fragments, ordered semantic surface/record/field membership and gated
semantic navigation to this Find owner. Source object identity is its content
revision; navigationKey identifies the admitted configuration/window/membership.
It must adopt its own surface eligibility/representation contract. It does not
copy Find state, query rows, editor drafts or vendor coordinates. Exhaustive
server search requires separate adoption and a server boundary; no unbounded
client scanning or speculative search framework has been prepared here.

Final review checks the complete tracked/untracked scope, original HEAD, two
preserved commits, no digest edit and no analyst-data changes. All browser writes
were confined to isolated fixture incidents. Work remains uncommitted for review.
The detached baseline worktree and its artifacts remain at
`/tmp/cartulary-tfn-baseline` for the documented independent reproduction.


Final typecheck after the terminal repairs: PASS 2/2,
`20260917T224759Z-p2044308`. Biome then identified only test-support export
ordering (`20260917T224800Z-p2044673`); `make explain-run` and its stdout identified
`assist/source/organizeImports`. `make format` passed 2/2 at
`20260917T224902Z-p2065322`, moving the two test-only exports without semantic
changes. Biome is rerun; no unrelated formatting changes were produced.


Final source navigation adds the Find owner to the Workbook parent guide. The
complete scope is 119 files: 55 non-image files and the 64 reviewed goldens.
The full-window Find browser case establishes responsive interaction and exact
results within the browser scenario timeouts; no new Find latency SLO is claimed.
Existing adopted Timeline measurement budgets passed separately as recorded.


Final Biome: PASS 2/2, `20260917T224943Z-p2080894`. Final Markdown pass before
the parent guide addition: PASS, `20260917T224847Z-p2059183`; the final complete
documentation pass follows that one-line navigation addition. Original HEAD and
branch divergence were rechecked: `d325b23a7f71aad6674349eddd2c22f67d2b0488`,
`main`, `origin/main...HEAD` counts `0 2`. `git diff --check` passes. Scope review
found no digest, backend, migration, dependency or lockfile changes and no missing
file in this handoff's inventory.


The repaired complete `make frontend-unit` run passes **645/645 units** at
`20260917T224758Z-p2044113`. Its `run-summary.json` and per-row diagnostics cover
all frontend owners, including the affected Workbook, collaboration, Adapter,
UI projection, saved-view, architecture and Network rows. G8 and G9 are now
PASS: public test-export parity and shared panel bounds are validated without
weakening selectors or assertions. No applicable blocked row remains.

### Completed-handoff checklist

- [x] TFN-01 through TFN-04 exited sequentially with actual saved evidence.
- [x] Scope, matching, readable values, ordering, focus, lifecycle and authority
  contracts adopted before dependent implementation.
- [x] One Find owner, source-owned committed text and settlement, neutral semantic
  Adapter boundary; no duplicate query, selection, draft or write executor.
- [x] G1–G9 remediation, affected areas, rationale, benefit, compatibility, risk
  and binary validation recorded; all applicable gap criteria PASS.
- [x] Production virtualized Find, bounded long-content scope, editor/recovery,
  spreadsheet, saved-view, other-surface and Network regression evidence passed.
- [x] Applicable accessibility and measurement evidence passed; every changed
  golden reviewed and two fresh ordinary visual passes passed after update.
- [x] Finalizer precedes broader terminal checks; no unrelated finalizer edits;
  RESULTS_DIR intentionally unset and retained-run maintenance recorded skipped.
- [x] Full frontend unit, type, lint, source/import boundary, JSON/generated-policy
  and generated-drift evidence passed; documentation completion recorded below.
- [x] Changed-file inventory, compatibility, excluded behavior, independently
  reproduced limitation, rollback, extension boundary and next action recorded.
- [x] Original branch/HEAD and existing commits preserved; no digest edit,
  commit, push, deployment or analyst-data change.


### TFN-05 exit

Final `make lint-markdown`: PASS at `20260917T225114Z-p2108624`, summary
`adhoc/lint-markdown/tool-run-summary.json`. Full frontend unit PASS 645/645,
final typecheck/Biome/import/layout checks PASS, generated/JSON/script checks PASS,
and the required production browser/accessibility/measurement/visual evidence
PASS as recorded. Final `git diff --check` and complete scope review PASS.
All applicable acceptance rows and G1–G9 criteria PASS; N/A rows have explicit
owner/scope rationale. No applicable BLOCKED row remains. TFN-05 is DONE.

Next action: review the uncommitted cohesive implementation and evidence-backed
handoff. No further implementation or required validation remains in this task.
The independently reproduced below-minimum Evidence diagnostic remains documented;
no exhaustive incident search, release gate or retained full-warm-run maintenance
claim is made.
