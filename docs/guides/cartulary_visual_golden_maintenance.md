# Cartulary Visual Golden Maintenance Guide

**Status**: Implementation-support guide
**Authority**: Core 00-04 own product behavior. The adopted testing harness NLSpec owns harness conformance. This guide does not promote visual snapshot refresh into current harness conformance.

## Purpose

Visual goldens are committed validation inputs for browser-rendered workbook states. They help detect UI drift in `browser-e2e-visual`, but they are not product behavior owners and they are not claim-bearing evidence by themselves.

## Canonical Surface

- Use `make browser-e2e-visual` for the canonical validation target.
- Visual workbook tests live in `apps/web/e2e/workbook.visual.spec.ts`.
- Committed Playwright goldens live beside the spec under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- Retained actual/diff artifacts from failed runs live under the run root reported by the harness, usually `.cartulary/test-results/<run-id>/.../playwright-output/`.

The authoritative current-profile visual rows are the `V-*` rows in
`tools/phase*_test_map.json` whose `execution_dependency` is `browser_visual`.
Those rows are separate from frontend `FE-*` readiness rows. A golden refresh
for an existing `V-*` row MUST cite the affected `V-*` row IDs. When the same
golden also serves a frontend visual fixture, the refresh MUST additionally cite
the applicable `FE-VFIX-*` IDs and frontend phase-map rows. Do not infer
frontend row closure from a Playwright title, filename, or base phase manifest
entry alone.

## Visual Fixture Matrix

The fixture matrix is implementation-support ownership for visual readiness rows. It does not create product behavior, does not replace frontend phase maps, and does not make visual screenshots claim-bearing evidence.

Fixture status is closed to `current`, `missing`, and `retired`. `current` means the fixture has an owned Playwright scenario and committed golden. `missing` means the fixture is required by frontend readiness planning but cannot satisfy support validation until added or explicitly blocked with a reason in the frontend phase map. `retired` means the fixture is intentionally no longer required and MUST name the replacement row or removal reason before the phase row can be complete.

| Fixture ID | Fixture title | Intended phase | Required surface state | Required scroll normalization | Status |
| --- | --- | --- | --- | --- | --- |
| `FE-VFIX-01` | Default grid shell | `FE-P2` | Main workbook grid with stable row version and save-state strip visible. | Top-left grid scroll; viewport `1440x900` unless the fixture row states otherwise. | `current` |
| `FE-VFIX-02` | Unresolved and resolved entity state | `FE-P5` | Entity evidence or request state renders both unresolved and resolved variants. | Named row anchor; dynamic identifiers masked. | `missing` |
| `FE-VFIX-03` | Same-field conflict | `FE-P7` | Conflict strip or resolver shows same-field conflict state and recovery affordance. | Top-left grid scroll or conflict row anchor. | `current` |
| `FE-VFIX-04` | Row-gutter presence | `FE-P7` | Row gutter or presence marker is visible and anchored to the intended row. | Presence row anchor; dynamic actor labels masked where needed. | `current` |
| `FE-VFIX-05` | Evidence affordance | `FE-P6` | Evidence badge, access control state, or blocked preview affordance is visible. | Evidence row anchor; preview dynamic values masked. | `current` |
| `FE-VFIX-06` | Grouped result | `FE-P8` | Grouped grid result or group header row is visible. | Group row anchor; scroll top normalized before capture. | `current` |
| `FE-VFIX-07` | Task Requests or Decisions | `FE-P10` | Task Requests or Decisions view state is visible with representative rows. | Top-left or named task row anchor. | `current` |
| `FE-VFIX-08` | Save-state strip | `FE-P4`, `FE-P7` | Saved, syncing, conflict, or recovered save-state strip is isolated for comparison. | Strip-level crop; grid scroll irrelevant unless row context is included. | `current` |
| `FE-VFIX-09` | Frozen column | `FE-P3`, `FE-P10` | Frozen column remains visible while horizontal scroll exposes far-right state. | Horizontal scroll right; frozen-column edge visible. | `missing` |
| `FE-VFIX-10` | Resize handle | `FE-P3`, `FE-P10` | Column resize affordance is visible in a deterministic hover/focus state. | Top-left grid scroll; pointer/hover state declared. | `missing` |
| `FE-VFIX-11` | Drag-fill handle | `FE-P3`, `FE-P10` | Drag-fill affordance is visible in a deterministic focus/editor state. | Top-left grid scroll; active cell declared. | `missing` |
| `FE-VFIX-12` | Edit cell | `FE-P3`, `FE-P4`, `FE-P10` | Active edit cell renders editor state and save-state relationship. | Top-left grid scroll; active cell declared. | `current` |
| `FE-VFIX-13` | Tree/group row | `FE-P3`, `FE-P8`, `FE-P10` | Tree or group row state is expanded/collapsed as declared by the fixture. | Group/tree row anchor. | `current` |
| `FE-VFIX-14` | Exposed theme states | `FE-P11` | Exposed theme, density, color, and token states render in representative controls. | Viewport `1280x720`; selector screenshot scope; no workbook-grid scroll state. | `current` |
| `FE-VFIX-15` | Empty successful query | `FE-P3`, `FE-P4`, `FE-P8` | Successful empty result state renders without error or loading affordance. | Top-left grid scroll; empty-state container anchored. | `missing` |

### Visual Support Acceptance

| ID | Requirement |
| --- | --- |
| `VG-AC-001` | The matrix MUST contain exactly one row for each required `FE-VFIX-01` through `FE-VFIX-15` identifier and MUST NOT contain duplicate fixture IDs. |
| `VG-AC-002` | Every `current` fixture MUST declare deterministic seed data, viewport, browser zoom, fixture order, dynamic masks, and artifact owner rows before golden refresh is accepted. |
| `VG-AC-003` | Every workbook-grid fixture MUST declare scroll normalization using `GridVisualScrollState` or an equivalent named anchor before capture. |
| `VG-AC-004` | Every golden refresh MUST cite an accepted refresh trigger, the affected authoritative `V-*` phase rows, any affected `FE-VFIX-*` fixture IDs and corresponding frontend phase-map rows when applicable, and whether dynamic masks, viewport state, or screenshot scope changed. |
| `VG-AC-005` | A `missing` fixture MUST fail support validation unless the corresponding frontend phase-map row is `claim_status="blocked"` with a precise missing-fixture reason. |

### Current shell/status refresh citation map

The workbook shell/status accessibility refresh uses this implementation-support
map when regenerating stale current-profile goldens. The trigger is the accepted
shell/status contract change: named workbook slot regions, the save-state
`role="status"`, the visually hidden workbook focus anchor, and hardened
screenshot scopes.

| Affected row | Owner map | Related fixture IDs | Screenshot scope |
| --- | --- | --- | --- |
| `FE-V-P2-01` | `tools/frontend_phase_maps/fe_p2_test_map.json` | `FE-VFIX-01` | Fixed viewport shell, snapshot `fe-v-p2-01-default-workbook-shell`. |
| `V-3-GRID-01` | `tools/phase3_test_map.json` | `FE-VFIX-01`, `FE-VFIX-08` | Fixed viewport shell. |
| `V-3-GRID-02` | `tools/phase3_test_map.json` | `FE-VFIX-08`, `FE-VFIX-12` | Grid shell and named status-strip slot. |
| `V-3-GRID-03` | `tools/phase3_test_map.json` | `FE-VFIX-06`, `FE-VFIX-13` | Grid shell. |
| `V-4-GRID-01` | `tools/phase4_test_map.json` | None currently claimed. | Grid shell. |
| `V-4-GRID-03` | `tools/phase4_test_map.json` | `FE-VFIX-07` | Grid shell. |
| `V-5-GRID-02` | `tools/phase5_test_map.json` | `FE-VFIX-05` | Grid shell. |
| `V-6-GRID-01` | `tools/phase6_test_map.json` | `FE-VFIX-04` | Grid shell. |
| `V-6-GRID-02` | `tools/phase6_test_map.json` | `FE-VFIX-03`, `FE-VFIX-08` | Fixed viewport shell. |
| `V-6-GRID-03` | `tools/phase6_test_map.json` | `FE-VFIX-03`, `FE-VFIX-08` | Fixed viewport shell and named status-strip slot. |

### Current exposed-theme fixture citation map

The exposed-theme fixture uses test-only DOM inside the ordinary workbook page
so generated theme CSS is injected through the same runtime path as the
workbook shell. It does not expose a theme switcher and does not claim support
for light or high-contrast themes.

| Affected row | Owner map | Related fixture IDs | Fixture seed | Viewport and zoom | Dynamic masks | Screenshot scope |
| --- | --- | --- | --- | --- | --- | --- |
| `FE-V-P11-03` | `tools/frontend_phase_maps/fe_p11_test_map.json` | `FE-VFIX-14` | Deterministic workbook incident plus test-only `section[data-design-fixture='exposed-theme']` token specimen. | `1280x720`, browser default zoom matching `{layout.zoomDefault}`. | None; specimen text is static and contains no generated IDs, actor names, timestamps, or cursors. | Selector crop, snapshot `fe-v-p11-03-exposed-theme-states`. |

## Accepted Refresh Triggers

Refresh a golden only when at least one of these is true:

- the UI contract intentionally changed;
- the visual harness intentionally changed viewport, masking, scroll normalization, or screenshot scope;
- a dependency, browser, or platform pin changed and the rendered output change is accepted;
- the previous golden is stale relative to already-validated functional behavior.

Do not refresh a golden to hide an unexplained product regression, broken functional assertion, unstable dynamic text, missing data, or a browser/runtime mismatch.

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

- the `FE-VFIX-*` row IDs claimed;
- the surface state, seed data, scroll state, focus/editor state, inspector state, dynamic masks, browser viewport, and target golden names;
- the matching frontend phase-map rows that own the fixture;
- whether the fixture is `current`, remains `missing`, or retires a prior fixture row with replacement rationale.

When the vendored font bundle changes, treat the change as an intentional visual-refresh trigger only after reviewing the diff. The visual harness waits for `document.fonts.ready`, requires active Inter and JetBrains Mono faces to load, and retains the `FONT_MANIFEST.json` SHA-256 with screenshot artifacts so font metric changes can be traced to the bundle version.

Skipped work units after a `browser-e2e-visual/visual` failure should be treated as cascade skips unless their own summaries show independent failures. Scheduler `resource:host_cpu`, `resource:host_io`, and dependency blocker counts are scheduling metadata, not root-cause failures.

## Validation Checklist

Before handoff:

- run the narrow Playwright update flow for the affected visual test;
- validate the visual fixture matrix before `make browser-e2e-visual`; any missing fixture row fails support validation unless it is marked blocked with a precise reason in the frontend phase map;
- inspect the generated diff and confirm it matches the intended visual contract;
- run `make frontend-unit` or `scripts/check-font-bundle.mjs` when font files, font CSS, manifests, or generated report templates changed;
- run `make browser-e2e-visual`;
- run `make agent-finalize`, passing `RESULTS_DIR=<successful full warm check run root>` when a successful retained run should refresh and validate timing maintenance inputs;
- report whether `agent-finalize` ran unchanged, updated generated artifacts, skipped retained-run maintenance because `RESULTS_DIR` was unset, or failed.

`RESULTS_DIR` must identify a successful, uncontaminated full warm `make check` retained run. Failed runs, service-backed-only runs, browser-only runs, phase-slice runs, and `make check` runs whose summaries report `failure_class=product` are invalid finalizer input; running `make agent-finalize` without `RESULTS_DIR` validates only the non-retained-run maintenance path.
