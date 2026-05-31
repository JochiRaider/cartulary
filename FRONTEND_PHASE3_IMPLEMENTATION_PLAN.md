# FE-P3 Implementation And Handoff

## Status

FE-P3 grid-adapter remediation is implemented as of 2026-05-31. The frontend
phase remains `status="planned"` in `tools/frontend_phase_registry.json`, but
all nine FE-P3 rows in `tools/frontend_phase_maps/fe_p3_test_map.json` now have
`claim_status="implemented"` and direct row-owned evidence from their intended
targets.

This file is a progress and handoff record only. It is not behavior authority.
Core 00 through Core 04 remain the current product-conformance authority.
`docs/testing-harness-nlspec.md` owns harness mechanics. Frontend visual and
accessibility evidence remain `design_direction`, not product conformance and
not Core 05 claim-publication evidence. Core 05 stayed inactive for this work.

## Remediation Decisions

| Gap | Resolution | Areas changed | Migration impact |
| --- | --- | --- | --- |
| FE-P3 rows had no direct row-owned implementation evidence. | FE-P3 rows were promoted only after exact scenario titles or support-target row accounting existed in the mapped target artifacts. Scenario-less implemented `implementation_support` rows can close from a passing mapped target while product/design rows remain scenario-backed. | Map, harness row accounting, implementation, tests, guide, generated ledger. | Old retained artifacts cannot close the revised rows. Consumers should use the new target artifacts listed below. |
| `FE-V-P3-01` mixed adapter visuals with later-phase fixtures. | FE-P3 now owns only grid-adapter visual states: frozen column, resize handle, drag-fill handle, edit cell, tree/group row, and empty successful query. Row-gutter presence remains FE-P7. Grouped-result query ownership remains FE-P8. | Frontend guide, visual guide, FE-P3 map, visual test, golden, generated ledger. | Added snapshot `apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p3-01-grid-adapter-fixtures-linux.png`. Historical FE-P3 visual artifacts are no longer closure evidence. |
| `FE-A11Y-P3-01` mapped to preflight smoke. | FE-P3 accessibility now closes under `make browser-e2e-a11y`; preflight remains only for blocked future rows. | Frontend guide, FE-P3 map, a11y test, accessibility summary writer, generated ledger. | CI/check surfaces now have stricter implemented FE-P3 a11y evidence. Preflight summaries no longer include FE-P3. |

## Implementation Summary

FE-P3 behavior is contained in `/packages/grid-adapter` and exposed through the
adapter boundary consumed by `/apps/web`. Direct RDG imports and the RDG
stylesheet remain adapter-owned.

Implemented adapter behavior:

- stable `record_id` row identity with unsafe or duplicate IDs rejected before
  mutation-capable anchoring;
- `field_key` cell identity for edit, paste, fill, focus, selection, and anchor
  translation;
- presentation-row write gating for group, tree, loading, spacer, and other
  non-record rows;
- explicit editor-adapter and contract writeability checks;
- deterministic renderer/editor resolution with cleanup hooks;
- sort, filter, group, resize, paste, drag-fill, scroll-to-cell, tree/group,
  and mutation-anchor browser helper evidence;
- sparse patch reconciliation that preserves unchanged row references and
  replaces changed rows by `record_id`;
- target-owned row-accounting emission for import-boundary and Biome support
  rows.

## Row Evidence

| Row | Target evidence | Closure |
| --- | --- | --- |
| `FE-U-P3-01` | `.cartulary/test-results/20260531T121030Z-p1670716/frontend-unit/frontend-row-accounting.json` | closed |
| `FE-U-P3-02` | `.cartulary/test-results/20260531T121030Z-p1670716/frontend-unit/frontend-row-accounting.json` | closed |
| `FE-U-P3-03` | `.cartulary/test-results/20260531T121030Z-p1670716/frontend-unit/frontend-row-accounting.json` | closed |
| `FE-U-P3-04` | `.cartulary/test-results/20260531T121030Z-p1670716/frontend-unit/frontend-row-accounting.json` | closed |
| `FE-I-P3-01` | `.cartulary/test-results/20260531T121030Z-p1670716/frontend-unit/frontend-row-accounting.json` | closed |
| `FE-B-P3-01` | `.cartulary/test-results/20260531T121352Z-p1677560/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260531T121442Z-p1681140/browser-e2e-webserver-backed/frontend-row-accounting.json` | closed |
| `FE-V-P3-01` | `.cartulary/test-results/20260531T121945Z-p1698018/browser-e2e-visual/frontend-row-accounting.json` | closed |
| `FE-A11Y-P3-01` | `.cartulary/test-results/20260531T122048Z-p1703601/browser-e2e-a11y/frontend-row-accounting.json` | closed |
| `FE-S-P3-01` | `.cartulary/test-results/20260531T121323Z-p1676077/frontend-import-boundary-check/frontend-row-accounting.json` and `.cartulary/test-results/20260531T121324Z-p1676275/lint-biome/frontend-row-accounting.json` | closed |

Accessibility evidence:

- `.cartulary/test-results/20260531T122048Z-p1703601/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`
  reports `FE-A11Y-P3-01` as implemented `design_direction` evidence with a
  passing scenario, keyboard checks, state-communication checks, contrast
  checks, and zero FE-P3 violations.
- `.cartulary/test-results/20260531T122138Z-p1707284/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json`
  passed with zero FE-P3 rows, confirming FE-P3 no longer relies on preflight
  smoke.

Visual evidence:

- `FE-V-P3-01` closes from `make browser-e2e-visual` with snapshot
  `fe-v-p3-01-grid-adapter-fixtures`.
- Claimed fixtures: `FE-VFIX-09`, `FE-VFIX-10`, `FE-VFIX-11`, `FE-VFIX-12`,
  `FE-VFIX-13`, and `FE-VFIX-15`.
- Deferred fixtures: `FE-VFIX-04` row-gutter presence remains FE-P7;
  `FE-VFIX-06` grouped result remains FE-P8.

## Validation Record

| Command | Result |
| --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3` | pass; reports all FE-P3 rows implemented and mapped to intended targets |
| `make frontend-typecheck` | pass, `.cartulary/test-results/20260531T121021Z-p1670349` |
| `make frontend-unit` | pass, `.cartulary/test-results/20260531T121030Z-p1670716` |
| `make frontend-import-boundary-check` | pass, `.cartulary/test-results/20260531T121323Z-p1676077` |
| `make lint-biome` | pass, `.cartulary/test-results/20260531T121324Z-p1676275` |
| `make browser-e2e-support` | pass, `.cartulary/test-results/20260531T121352Z-p1677560` |
| `make browser-e2e-webserver-backed` | pass, `.cartulary/test-results/20260531T121442Z-p1681140` |
| `make browser-e2e-visual` | pass, `.cartulary/test-results/20260531T121945Z-p1698018` |
| `make browser-e2e-a11y` | pass, `.cartulary/test-results/20260531T122048Z-p1703601` |
| `make browser-e2e-a11y-preflight` | pass, `.cartulary/test-results/20260531T122138Z-p1707284` |
| `make phase-ledgers` | pass, `.cartulary/test-results/20260531T121056Z-p1672401` |
| `make generated-artifact-policy-check` | pass, `.cartulary/test-results/20260531T123721Z-p1776984` |
| `make phase-ledger-drift` | pass, `.cartulary/test-results/20260531T123721Z-p1776999` |
| `make phase-schedule-drift` | pass, `.cartulary/test-results/20260531T123721Z-p1776987` |
| `make lint-scripts` | pass, `.cartulary/test-results/20260531T122349Z-p1713874` |
| `make generate-drift` | pass, `.cartulary/test-results/20260531T122357Z-p1714188` |
| `git diff --check` | pass |
| `make agent-finalize` | pass, `.cartulary/test-results/20260531T122524Z-p1716747`; generated maintenance files updated, retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make check` | pass, `.cartulary/test-results/20260531T122740Z-p1721061`, 167/167 work units and 817 tests |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260531T122740Z-p1721061` | pass, `.cartulary/test-results/20260531T123531Z-p1773328`; retained-run maintenance validated, duration baselines refreshed, run checks passed |

`make agent-finalize` refreshed `tools/execution_topology_render_index.json`
and `tools/scheduler_manifest.json`; its embedded `phase-ledger-drift`,
`phase-schedule-drift`, `json-shape-check`, and duration-baseline coverage
substeps passed.

The retained-run `agent-finalize` refreshed
`tools/browser_e2e_duration_baselines.json`,
`tools/go_test_duration_baselines.json`,
`tools/harness_smoke_duration_baselines.json`, and
`tools/service_backed_make_target_duration_baselines.json`. Its retained-run
preflight, duration-baseline drift suite, scheduler event-order drift, and
scheduler summary timing drift checks passed.

## FE-P4 Handoff

FE-P4 can build on a grid foundation with direct evidence for stable row/cell
identity, adapter-only RDG containment, editability gating, command-helper
translation, sparse patch preservation, deterministic visual fixtures, and
implemented accessibility coverage.

Carry-forward constraints:

- Preserve RDG import and stylesheet ownership in `/packages/grid-adapter`.
- Keep product/design/support evidence classes separate in maps, ledgers, and
  retained artifacts.
- Do not reuse FE-P3 visual artifacts to close FE-P7 row-gutter or FE-P8
  grouped-result ownership.
- Continue closing product and design rows only from exact scenario-backed
  target artifacts.
- Keep scenario-less target-level closure limited to implemented
  `implementation_support` rows.

## Unresolved Blockers

No FE-P3-owned blocker remains in the current remediation record.
