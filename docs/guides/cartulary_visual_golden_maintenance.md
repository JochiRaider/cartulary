# Cartulary Visual Golden Maintenance Guide

**Status**: Implementation-support guide
**Authority**: Core 00-04 own product behavior. The adopted testing harness NLSpec owns harness conformance. This guide does not promote visual snapshot refresh into current harness conformance.

## Purpose

Visual goldens are committed validation inputs for browser-rendered workbook states. They help detect UI drift in `browser-e2e-visual`, but they are not product behavior owners and they are not claim-bearing evidence by themselves.

For pre-MVP browser inspection and design-discovery review, use
`docs/guides/cartulary_browser_design_readiness_workflow.md` before accepting or
refreshing goldens. That workflow treats current goldens as coverage hints and
regression inputs until the live layout has been reviewed directly in a browser.

## Canonical Surface

- Use `make browser-e2e-visual` for the canonical validation target.
- Use `make browser-e2e-visual-update` for the canonical helper-only refresh flow when committed Playwright visual goldens are intentionally updated.
- Visual workbook tests live in `apps/web/e2e/workbook.visual.spec.ts`.
- Committed Playwright goldens live beside the spec under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- Retained actual/diff artifacts from failed runs live under the run root reported by the harness, usually `.cartulary/test-results/<run-id>/.../playwright-output/`.
- Concept images and external bitmaps are design inputs only. The only committed golden source is Playwright output captured from the running app.

The authoritative current-profile visual rows are the active `visual` evidence
rows in `tools/test_families/*.json`. A golden refresh MUST cite the affected
owner row IDs and stable fixture IDs. Do not infer row closure from a Playwright
title, filename, or fixture identifier alone.

## Visual Fixture Matrix

The fixture matrix is implementation-support ownership for visual readiness rows. It does not create product behavior, does not replace the owner catalog, and does not make visual screenshots claim-bearing evidence.

Fixture status is closed to `current`, `missing`, and `retired`. `current` means the fixture has an owned Playwright scenario and committed golden. `missing` means an active owner row requires the fixture but cannot satisfy support validation until it is added or the row is removed through owner review. `retired` means the fixture is intentionally no longer required and MUST name the replacement row or owner-approved removal reason.

| Fixture ID | Fixture title | Intended phase | Required surface state | Required scroll normalization | Status |
| --- | --- | --- | --- | --- | --- |
| `visual.fixture.default_timeline_workbook_shell` | Default Timeline workbook shell | `browser.workbook-shell`, `browser.inspector-history` | App-owned workbook shell root with top bar, compact view bar with query controls after saved-view actions, compact Timeline grid, row gutter, header affordances, row-context inspector, status strip, Core 01 default Timeline fields, selected row context, focused Summary cell, and no admin/control card stack above the active grid. browser.inspector-history supporting artifacts extend the row-context inspector state to history, rollback preview, destructive confirmation, and public error fixtures. | Top-left outer and grid scroll; viewport `1440x900` unless the fixture row states otherwise. | `current` |
| `visual.fixture.mention_chip_state_matrix` | Mention chip state matrix | `browser.entity-linking` | Timeline relationship chips and inspector mention state render unresolved, resolved, auto-resolved, dismissed, and resolved chips with manual-resolution metadata together. | Top-left grid scroll; incident and generated record identifiers masked. | `current` |
| `visual.fixture.same_field_conflict` | Same-field conflict | `browser.collaboration` | Conflict strip or resolver shows same-field conflict state and recovery affordance. | Top-left grid scroll or conflict row anchor. | `current` |
| `visual.fixture.row_gutter_presence` | Row-gutter presence | `browser.collaboration` | Row gutter or presence marker is visible and anchored to the intended row. | Presence row anchor; dynamic actor labels masked where needed. | `current` |
| `visual.fixture.evidence_affordance` | Evidence affordance | `browser.evidence-workflow` | Evidence grid count, access-control affordance, available, requested, pending, blocked, failed-handle, inconsistent-handle, unsupported-preview, download-handle states, and Timeline inspector evidence count are visible. | Evidence grid right-edge actions column; Timeline row anchor with inspector evidence section open; incident and generated record identifiers masked. | `current` |
| `visual.fixture.saved_view_query_controls_and_grouped_result` | Grouped result | `browser.saved-view-query` | Grouped grid result or group header row is visible. | Group row anchor; scroll top normalized before capture. | `current` |
| `visual.fixture.task_requests_or_decisions` | Task Requests or Decisions | `browser.coordination-review` | Task Requests or Decisions view state is visible with representative rows. | Top-left or named task row anchor. | `current` |
| `visual.fixture.save_state_strip` | Save-state and recovery | `browser.mutation-lifecycle`, `browser.collaboration` | Saved, syncing, conflict, recovered save-state, and actionable blocked-edit recovery presentations are isolated for comparison. | Strip-level or full-viewport capture; grid scroll is normalized when row context is included. | `current` |
| `visual.fixture.frozen_column` | Frozen column | `browser.grid-interaction`, `browser.coordination-review` | Frozen column remains visible while horizontal scroll exposes far-right state. | Horizontal scroll right; frozen-column edge visible. | `current` |
| `visual.fixture.resize_handle` | Resize handle | `browser.grid-interaction`, `browser.coordination-review` | Column resize affordance is visible in a deterministic hover/focus state. | Top-left grid scroll; pointer/hover state declared. | `current` |
| `visual.fixture.drag_fill_handle` | Fill-down handle | `browser.grid-interaction`, `browser.coordination-review` | Fill-down affordance is visible in a deterministic focus/editor state. | Top-left grid scroll; active cell declared. | `current` |
| `visual.fixture.edit_cell` | Edit cell | `browser.grid-interaction`, `browser.mutation-lifecycle`, `browser.coordination-review` | Active edit cell renders editor state and save-state relationship. | Top-left grid scroll; active cell declared. | `current` |
| `visual.fixture.tree_group_row` | Group outline row | `browser.grid-interaction`, `browser.saved-view-query`, `browser.coordination-review` | Group row state is expanded/collapsed as declared by the fixture. | Group row anchor. | `current` |
| `visual.fixture.exposed_theme_states` | Exposed theme states | `browser.design-readiness` | Exposed theme, density, color, and token states render in representative controls. | Viewport `1280x720`; `capture_scope.kind="selector"`; no workbook-grid scroll state. | `current` |
| `visual.fixture.empty_successful_query` | Empty successful query | `browser.grid-interaction`, `browser.mutation-lifecycle`, `browser.saved-view-query` | Successful empty result state renders without error or loading affordance. | Top-left grid scroll; empty-state container anchored. | `current` |
| `visual.fixture.host_identity_merge_inspector` | Host or Identity merge inspector | `browser.inspector-history` | Selected host or identity row shows merge plan, reviewer-gated action state, and stale-confirmation invalidation affordance. Covers AC-454, AC-457, and AC-459. | Selected entity row anchor; grid and inspector scroll normalized independently. | `missing` |
| `visual.fixture.evidence_blocked_preview_inspector` | Evidence blocked preview inspector | `browser.inspector-history` | Evidence inspector shows blocked or unsupported preview state with download affordance where allowed and no third-party egress affordance. Covers AC-460. | Evidence row anchor; inspector Evidence panel visible. | `missing` |
| `visual.fixture.history_rollback_confirmation` | History rollback/destructive confirmation | `browser.inspector-history` | History panel shows rollback preview or destructive confirmation bound to the selected row and invalidated when row context changes. Covers AC-457 and AC-458. | Timeline selected row anchor; inspector History panel visible. | `missing` |
| `visual.fixture.handoff_acknowledgement_inspector` | Handoff acknowledgement inspector | `browser.coordination-review` | Handoff inspector shows acknowledgement state, open task/decision/risk refs, and read-only blocked state when not authorized. Covers AC-458 and AC-459. | Handoff selected row anchor. | `missing` |
| `visual.fixture.status_review_blocked_work_inspector` | Status Review blocked-work inspector | `browser.coordination-review` | Status Review inspector shows blocked tasks, pending evidence, open decisions, active risks, and next-report state. Covers AC-458 and AC-459. | Status Review selected row anchor. | `missing` |
| `visual.fixture.timeline_create_related_workflow_inspector` | Timeline create-related workflow inspector | `browser.inspector-history`, `browser.coordination-review` | Timeline selected row with Workflow panel showing create-related actions for Task Request, Decision, Evidence, Communications Log, Handoff, Status Review, and Lesson; ordinary Timeline grid remains visible and editable when inspector is closed. | Timeline row anchor; grid and inspector scroll normalized independently. | `missing` |

### Visual Support Acceptance

| ID | Requirement |
| --- | --- |
| `VG-AC-001` | The matrix MUST contain exactly one row for each fixture identifier currently declared by `tools/frontend_visual_fixture_registry.json` and MUST NOT contain duplicate fixture IDs. |
| `VG-AC-002` | Every `current` fixture MUST declare deterministic seed data, viewport, browser zoom, fixture order, capture scope, dynamic masks or an explicit no-dynamic-regions declaration, artifact owner rows, a primary golden filename, and every supporting golden artifact owned by the fixture before golden refresh is accepted. |
| `VG-AC-003` | Every workbook-grid fixture MUST declare scroll normalization using `GridVisualScrollState` or an equivalent named anchor before capture. |
| `VG-AC-004` | Every golden refresh MUST cite an accepted refresh trigger, the affected owner row IDs and stable fixture IDs, and whether dynamic masks, viewport state, or screenshot scope changed. |
| `VG-AC-005` | A `missing` fixture MUST remain explicit in `tools/frontend_visual_fixture_registry.json` with a precise `blocked_reason`, and it MUST NOT be cited as closing visual evidence until a row-owned Playwright scenario and committed golden make the fixture `current`. |

### Current shell/status refresh citation map

The workbook shell/status accessibility refresh uses this implementation-support
map when regenerating stale current-profile goldens. The trigger is the accepted
shell/status contract change: named workbook slot regions, the save-state
`role="status"`, the visually hidden workbook focus anchor, and hardened
screenshot scopes. The current accepted shell refresh additionally includes the
width-only shell chrome contract, the vertical-only resize regression fix for
the top bar, and reactivated Playwright screenshot comparisons.

| Affected row | Current authority | Related fixture IDs | Screenshot scope |
| --- | --- | --- | --- |
| `visual.workbook-shell.row-01` | `tools/test_families/*.json` | `visual.fixture.default_timeline_workbook_shell` | App-owned shell root, snapshot `visual.workbook-shell.row-01-default-timeline-workbook-shell`. |
| `V-3-GRID-01` | `tools/test_families/*.json` | `visual.fixture.default_timeline_workbook_shell`, `visual.fixture.save_state_strip` | Fixed viewport shell. |
| `V-3-GRID-02` | `tools/test_families/*.json` | `visual.fixture.save_state_strip`, `visual.fixture.edit_cell` | Grid shell and named status-strip slot. |
| `V-3-GRID-03` | `tools/test_families/*.json` | `visual.fixture.saved_view_query_controls_and_grouped_result`, `visual.fixture.tree_group_row` | Grid shell. |
| `V-4-GRID-01` | `tools/test_families/*.json` | None currently claimed. | Grid shell. |
| `V-4-GRID-03` | `tools/test_families/*.json` | `visual.fixture.task_requests_or_decisions` | Grid shell. |
| `V-5-GRID-02` | `tools/test_families/*.json` | `visual.fixture.evidence_affordance` | Grid shell. |
| `V-6-GRID-01` | `tools/test_families/*.json` | `visual.fixture.row_gutter_presence` | Grid shell. |
| `V-6-GRID-02` | `tools/test_families/*.json` | `visual.fixture.same_field_conflict`, `visual.fixture.save_state_strip` | Fixed viewport shell. |
| `V-6-GRID-03` | `tools/test_families/*.json` | `visual.fixture.same_field_conflict`, `visual.fixture.save_state_strip` | Fixed viewport shell and named status-strip slot. |

### Current Evidence managed-object refresh citation map

The Evidence managed-object refresh accepts the current object-blob storage
label shown after an attached Evidence row reaches the available lifecycle
state. This map is current-profile visual maintenance for the authoritative
`V-*` row only; browser.evidence-workflow readiness is owned separately by `visual.fixture.evidence_affordance` and
`visual.evidence-workflow.row-01`.

| Affected row | Current authority | Related fixture IDs | Screenshot scope |
| --- | --- | --- | --- |
| `V-5-GRID-01` | `tools/test_families/*.json` | None currently claimed. | Grid shell, snapshot `v-5-grid-01-available-evidence`. |

### Current browser.grid-interaction grid-adapter fixture citation map

The browser.grid-interaction grid-adapter fixture uses deterministic test-only DOM inside an
ordinary workbook page so adapter visual states can be owned without pulling
later row-gutter or grouped-result query behavior into browser.grid-interaction.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.grid-interaction.row-01` | `tools/test_families/*.json` | `visual.fixture.frozen_column`, `visual.fixture.resize_handle`, `visual.fixture.drag_fill_handle`, `visual.fixture.edit_cell`, `visual.fixture.tree_group_row`, `visual.fixture.empty_successful_query` | Deterministic workbook incident plus test-only `section[data-design-fixture='grid-interaction-grid-adapter']` grid-adapter specimen. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; specimen text is static and contains no generated IDs, actor names, timestamps, or cursors. | Selector crop, snapshot `visual.grid-interaction.row-01-grid-adapter-fixtures`. |

### Current browser.mutation-lifecycle visual readiness fixture citation map

The browser.mutation-lifecycle visual readiness fixture uses deterministic app-owned Timeline
workbook state. It captures the real status strip, pending queue notice,
transaction-conflict recovery panel, active inline edit cell, and successful
empty Timeline query state. This map is
design-direction evidence only; it does not create product conformance or Core
05 publication evidence.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.mutation-lifecycle.row-01` | `tools/test_families/*.json` | `visual.fixture.save_state_strip`, `visual.fixture.edit_cell`, `visual.fixture.empty_successful_query` | Deterministic workbook incident with one Timeline row, held transport failure for pending replay, one injected transaction conflict for actionable recovery, and a fresh zero-row Timeline incident for the empty query state. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; generated record identifiers and clock-derived labels masked by visual preparation; transaction identifiers are not rendered. | Grid/status captures, snapshots `visual.mutation-lifecycle.row-01-active-edit-cell`, `visual.mutation-lifecycle.row-01-pending-replay-status`, `visual.mutation-lifecycle.row-01-transaction-recovery-panel`, and `visual.mutation-lifecycle.row-01-empty-timeline-query`. |

### Current browser.entity-linking visual readiness fixture citation map

The browser.entity-linking visual readiness fixture uses deterministic app-owned Timeline
workbook state. It captures the unresolved token chip, direct resolved chip,
resolved chip with manual-resolution metadata, auto-resolved chip, and dismissed
mention inspector state in one first-viewport fixture. This map is
design-direction evidence
only; it does not create product conformance or Core 05 publication evidence.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.entity-linking.row-01` | `tools/test_families/*.json` | `visual.fixture.mention_chip_state_matrix` | Deterministic workbook incident with Timeline rows for unresolved, resolved, resolved-with-manual-metadata, auto-resolved, and dismissed mention states. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; generated record identifiers and clock-derived labels masked by visual preparation. | Fixed viewport capture, snapshot `visual.entity-linking.row-01-mention-chip-states`. |

### Current browser.evidence-workflow visual readiness fixture citation map

The browser.evidence-workflow visual readiness fixture uses deterministic app-owned Evidence and
Timeline workbook state. It is design-direction evidence only and keeps Core 05
claim-publication evidence out of scope.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.evidence-workflow.row-01` | `tools/test_families/*.json` | `visual.fixture.evidence_affordance` | Deterministic workbook incident with requested, pending, quarantined, available-preview, available-download, unsupported-preview, failed-handle, and inconsistent-handle Evidence rows, plus one Timeline row with an attached PNG evidence count. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; generated record identifiers and clock-derived labels masked by visual preparation; native file-input dimensions are normalized only for the browser.evidence-workflow actions-column screenshot crop. | Evidence grid right-edge actions-column capture, snapshot `visual.evidence-workflow.row-01-evidence-affordance-states`; Timeline row evidence-actions anchor capture, snapshot `visual.evidence-workflow.row-01-timeline-evidence-count`. |

### Current browser.collaboration visual readiness fixture citation map

The browser.collaboration visual readiness fixture uses deterministic app-owned Timeline
workbook state and public collaboration traffic. It is design-direction
evidence only and keeps Core 05 claim-publication evidence out of scope. Fresh
browser.collaboration closure requires current `browser-e2e-visual` row accounting; retained
passing visual evidence is context only after a newer visual audit fails.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.collaboration.row-01` | `tools/test_families/*.json` | `visual.fixture.same_field_conflict`, `visual.fixture.row_gutter_presence`, `visual.fixture.save_state_strip` | Deterministic workbook incidents for a remote Timeline presence row, same-field conflict resolver state with the exact rejected grid draft retained, conflict and recovered save-state strips, and a delete/restore invalidate refresh strip. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; generated record identifiers and dynamic actor connection data masked by visual preparation; remote actor initials remain deterministic. | Grid capture `visual.collaboration.row-01-presence-markers`; viewport capture `visual.collaboration.row-01-conflict-resolver`; status-strip captures `visual.collaboration.row-01-conflict-strip`, `visual.collaboration.row-01-recovered-saved-strip`, and `visual.collaboration.row-01-reset-invalidate-notice`. |

### Current browser.inspector-history visual readiness fixture citation map

The browser.inspector-history visual readiness fixture uses deterministic app-owned Timeline and
Evidence workbook state. It is design-direction evidence only and does not
create product conformance or Core 05 claim-publication evidence.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.inspector-history.row-01` | `tools/test_families/*.json` | `visual.fixture.default_timeline_workbook_shell`, `visual.fixture.host_identity_merge_inspector`, `visual.fixture.evidence_blocked_preview_inspector`, `visual.fixture.history_rollback_confirmation` | Deterministic workbook incident with one Timeline row linked to one Evidence row, an unresolved host relationship token, retained row history, rollback preview, destructive confirmation, public rollback error envelope, host/identity merge state, and blocked Evidence preview state. | `1440x900`, browser default zoom matching `{layout.zoomDefault}`. | Incident identity masked by the ordinary workbook visual helper; generated record identifiers, history references, and clock-derived labels masked by visual preparation. | Fixed viewport captures `visual.inspector-history.row-01-inspector-relationships`, `visual.inspector-history.row-01-inspector-history`, `visual.inspector-history.row-01-rollback-preview`, `visual.inspector-history.row-01-destructive-confirmation`, `visual.inspector-history.row-01-public-error`, plus added selector captures for merge and blocked-preview fixtures when implemented. |

### Current exposed-theme fixture citation map

The exposed-theme fixture uses test-only DOM inside the ordinary workbook page
so generated theme CSS is injected through the same runtime path as the
workbook shell. It does not expose a theme switcher and does not claim support
for light or high-contrast themes.

| Affected row | Current authority | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `visual.design-readiness.row-03` | `tools/test_families/*.json` | `visual.fixture.exposed_theme_states` | Deterministic workbook incident plus test-only `section[data-design-fixture='exposed-theme']` token specimen. | `1280x720`, browser default zoom matching `{layout.zoomDefault}`. | None; specimen text is static and contains no generated IDs, actor names, timestamps, or cursors. | `capture_scope.kind="selector"` with `[data-design-fixture='exposed-theme']`, snapshot `visual.design-readiness.row-03-exposed-theme-states`. |

## Accepted Refresh Triggers

Refresh a golden only when at least one of these is true:

- the UI contract intentionally changed;
- the visual harness intentionally changed viewport, masking, scroll normalization, or screenshot scope;
- Playwright visual comparison behavior intentionally changed, including reactivation of committed golden comparisons;
- a dependency, browser, or platform pin changed and the rendered output change is accepted;
- the previous golden is stale relative to already-validated functional behavior.

Do not refresh a golden to hide an unexplained product regression, broken functional assertion, unstable dynamic text, missing data, or a browser/runtime mismatch.

`visual.fixture.default_timeline_workbook_shell` refreshes MUST be fixed first-viewport captures. Full-page captures, locator captures taller than the viewport, browser.grid-interaction grid-adapter support specimens, browser.design-readiness token/theme specimens, concept images, or screenshots containing the default-shell admin/demo card stack cannot satisfy the default Timeline workbook-shell fixture.

## Workbook Grid Scroll Contract

Every workbook-grid visual snapshot must declare its scroll normalization through the test helper. For `GridVisualScrollState.left`:

- `"left"` means `scrollLeft = 0`.
- `"right"` means the live computed maximum, `scrollWidth - clientWidth`, clamped at zero.
- a number means that exact requested offset, clamped into the live scroll range.

Use `"right"` for snapshots whose purpose is to validate far-right columns, badges, or status fields. Do not encode historical pixel offsets unless the specific pixel position is the behavior under test.

## Review Expectations

When visual normalization helpers change, the handoff or pull request must state:

- which helper behavior changed;
- which snapshots were affected;
- which goldens were regenerated, or why none were affected;
- whether the visual diff is explained by the intended contract change.

When a frontend fixture is added, the handoff or pull request must state:

- the semantic `visual.fixture.*` IDs claimed;
- the surface state, seed data, scroll state, focus/editor state, inspector state, dynamic masks, browser viewport, and target golden names;
- the matching owner catalog rows that own the fixture;
- whether the fixture is `current`, remains `missing`, or retires a prior fixture row with replacement rationale.

When the vendored font bundle changes, treat the change as an intentional visual-refresh trigger only after reviewing the diff. The visual harness waits for `document.fonts.ready`, requires active Inter and JetBrains Mono faces to load, and retains the `FONT_MANIFEST.json` SHA-256 with screenshot artifacts so font metric changes can be traced to the bundle version.

Before refreshing any font- or glyph-sensitive golden, inspect the retained
`*-render-diagnostics` and `*-grid-diagnostics` attachments. The render
diagnostics record browser version, user agent, Playwright viewport metadata,
`devicePixelRatio`, font-face status, computed grid/chip typography, grid token
CSS variables, and representative text/chip bounds. Grid diagnostics additionally
record density ID, computed row height, computed line height, focused element,
shell and scrollport state, visible field keys, and chip bounds when chips are
present.

Skipped work units after a `browser-e2e-visual/visual` failure should be treated as cascade skips unless their own summaries show independent failures. Scheduler `resource:host_cpu`, `resource:host_io`, and dependency blocker counts are scheduling metadata, not root-cause failures.

## Validation Checklist

Before handoff:

- run `make browser-e2e-visual-update` when committed Playwright goldens are intentionally refreshed;
- validate the visual fixture matrix before `make browser-e2e-visual`; any missing active owner fixture fails support validation until restored or removed through owner review;
- inspect the generated diff and confirm it matches the intended visual contract;
- run `make frontend-unit` or `tools/harness/frontend/font-bundle-check-cli.mjs` when font files, font CSS, manifests, or generated report templates changed;
- run `make browser-e2e-visual`;
- run `make agent-finalize`, passing `RESULTS_DIR=<successful full warm check run root>` when a successful retained run should refresh and validate timing maintenance inputs;
- report whether `agent-finalize` ran unchanged, updated generated artifacts, skipped retained-run maintenance because `RESULTS_DIR` was unset, or failed.

`RESULTS_DIR` must identify a successful, uncontaminated full warm `make check` retained run. Failed runs, owner-slice runs, service-backed-only runs, browser-only runs, and `make check` runs whose summaries report `failure_class=product` are invalid finalizer input; running `make agent-finalize` without `RESULTS_DIR` validates only the non-retained-run maintenance path.
