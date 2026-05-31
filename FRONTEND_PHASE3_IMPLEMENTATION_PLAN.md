# FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md

## Summary

Create and maintain this file as the FE-P3 execution roadmap, progress marker,
and FE-P4 handoff aid for `FE-P3: Grid Adapter And View-Schema Rendering`.

This plan is not behavior authority and must not mark FE-P3 rows complete
without direct row-owned evidence from the intended target or an exact blocker.
`docs/guides/cartulary_frontend_implementation_testing_guide.md` controls the
FE-P3 row set, evidence classes, target mapping, and completion rules. Core 00
through Core 04 remain the only current product-conformance authority unless a
future adopted NLSpec changes that boundary. Core 05 stays inactive unless a
claim-bearing timed, benchmark, performance, or fixture-sensitive publication
predicate is explicitly introduced.

Repository reconnaissance found the active root-level naming precedent
`FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`,
`FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md`, and
`FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md`. No conflicting governing phase-plan
naming convention was found, so this plan uses
`FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md`.

This planning task does not implement FE-P3 product behavior.

## Authority Model

- Product-conformance behavior is owned only by Core 00 through Core 04 or a
  future adopted NLSpec.
- `docs/testing-harness-nlspec.md` owns harness mechanics only: command
  invocation, target selection, scheduling, fixture lifecycle, artifact
  emission, summary emission, cleanup, and harness verification gates.
- `docs/domain.md` is vocabulary and concept-boundary support. It is not
  product behavior authority.
- The UI/UX guide, `docs/design.md`, visual-golden guide, development guide,
  and research reports provide design-direction, implementation-support, or
  rationale only.
- Core 05 remains inactive unless a claim-bearing timed, benchmark,
  performance, or fixture-sensitive publication predicate is explicitly
  introduced.
- Product-conformance, design-direction, implementation-support,
  TODO-owner-lookup, and claim-publication-boundary evidence must remain
  separate in plan text, frontend phase maps, generated ledgers, summaries, row
  accounting, and closure judgment.
- Generated frontend ledgers are downstream artifacts rendered from the
  frontend registry and maps. They are not behavior authority and must not be
  hand-edited.

## Current Repo Status

Locally verified facts:

- Before this document was created, root-level frontend phase plans existed for
  FE-P0, FE-P1, and FE-P2 only.
- `tools/frontend_phase_registry.json` contains `FE-P3` in namespace
  `frontend`, `status="planned"`, map
  `tools/frontend_phase_maps/fe_p3_test_map.json`, ledger
  `docs/testing/frontend_phase_coverage_ledgers/fe_p3_coverage_ledger.md`, and
  dependencies `FE-P0`, `FE-P1`, and `FE-P2`.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3` passed and reports
  FE-P3 as planned, explainable, and non-executable.
- `tools/frontend_phase_maps/fe_p3_test_map.json` contains exactly these nine
  rows once each: `FE-U-P3-01`, `FE-U-P3-02`, `FE-U-P3-03`, `FE-U-P3-04`,
  `FE-I-P3-01`, `FE-B-P3-01`, `FE-V-P3-01`, `FE-A11Y-P3-01`, and
  `FE-S-P3-01`.
- All FE-P3 rows currently have `claim_status="blocked"` in the phase map and
  generated ledger.
- The FE-P3 map uses these evidence classes: `product_conformance`,
  `implementation_support`, and `design_direction`.
- The generated FE-P3 ledger exists at
  `docs/testing/frontend_phase_coverage_ledgers/fe_p3_coverage_ledger.md` and
  states that it is generated from `tools/frontend_phase_maps/fe_p3_test_map.json`.
- `make phase-ledger-drift` passed after plan creation with run root
  `.cartulary/test-results/20260531T035102Z-p1032419` and summary
  `.cartulary/test-results/20260531T035102Z-p1032419/phase-ledger-drift/tool-run-summary.json`.
- `make agent-finalize` passed after plan creation with run root
  `.cartulary/test-results/20260531T035106Z-p1032754` and summary
  `.cartulary/test-results/20260531T035106Z-p1032754/agent-finalize/tool-run-summary.json`.
  It reported generated artifacts unchanged and skipped retained-run checks
  because `RESULTS_DIR` was unset.
- FE-P0, FE-P1, and FE-P2 are all still `status="planned"` in the frontend
  registry, but their current maps mark all rows `claim_status="implemented"`.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`,
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`, and
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2` passed and report
  those planned phases as explainable but non-executable.
- `/packages/grid-adapter`, `/packages/view-contracts`,
  `/packages/ui-contracts`, `/packages/test-utils`, and `/apps/web` all have
  package manifests.
- `packages/grid-adapter/package.json` depends on `react-data-grid`
  `7.0.0-beta.59`.
- In the inspected package and app surfaces, direct source imports of
  `react-data-grid` were found in `/packages/grid-adapter/src/index.tsx`; no
  `/apps/web` source import from `react-data-grid` was found.
- The RDG stylesheet import was found once in inspected authored source:
  `/packages/grid-adapter/src/index.tsx` imports
  `react-data-grid/lib/styles.css`.
- `/apps/web/src/WorkbookShell.tsx` consumes exports from
  `@cartulary/grid-adapter`.
- `/packages/grid-adapter/src/core.ts` currently exports helpers named
  `assertGridRows`, `buildGridPresentationRows`, `resolveGridCellAnchor`,
  `navigateGridCellAnchor`, `resolveGridPasteTargets`, and
  `reconcileRecordRows`. This is source-shape context only, not FE-P3 row
  closure evidence.
- Searching inspected app, package, tool, and generated-ledger surfaces found
  FE-P3 row IDs only in `tools/frontend_phase_maps/fe_p3_test_map.json` and
  `docs/testing/frontend_phase_coverage_ledgers/fe_p3_coverage_ledger.md`.
- `apps/web/e2e/workbook.visual.spec.ts` contains existing `v-3-grid-*`
  visual tests, but no `FE-V-P3-01` row-owned test title was found in that
  file during inspection.

Explicit TODOs and blockers:

- TODO: Collect direct row-owned evidence for every FE-P3 row before any
  completion claim. Existing source code, generated ledgers, broad regression
  suites, visual snapshots, and support checks do not close FE-P3 rows by
  themselves.
- TODO: Reconcile FE-P3 visual fixture ownership before closing
  `FE-V-P3-01`. The frontend guide assigns FE-P3 visual coverage for frozen
  column, resize handle, drag-fill handle, edit cell, tree/group row, grouped
  result, row gutter presence, and empty successful query. The visual-golden
  guide currently marks `FE-VFIX-09`, `FE-VFIX-10`, `FE-VFIX-11`, and
  `FE-VFIX-15` as `missing`; marks `FE-VFIX-12` and `FE-VFIX-13` as
  `current`; and lists `FE-VFIX-04` and `FE-VFIX-06` as intended for later
  phases even though the FE-P3 guide row names row-gutter presence and grouped
  result.
- TODO: Treat the current `FE-A11Y-P3-01` target
  `make browser-e2e-a11y-preflight` as blocked-row smoke unless the frontend
  map is intentionally promoted under the guide's implemented-accessibility
  completion rules.
- TODO: Rerun FE-P0, FE-P1, and FE-P2 row-owned regression commands during
  implementation closure. This plan creation only inspected phase metadata and
  did not rerun their full row-owned command sets.

## Source Limits

Inspected sources include:

- `FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md`;
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`;
- `docs/testing-harness-nlspec.md`;
- `docs/domain.md`;
- `docs/guides/cartulary-dev-guide.md`;
- `docs/guides/cartulary-ui-ux-design-guide.md`;
- `docs/guides/cartulary_visual_golden_maintenance.md`;
- `docs/design.md`;
- Core 00 through Core 04 under `docs/spec/`;
- Core 05 under `docs/spec/` for claim-publication boundary separation only;
- `tools/frontend_phase_registry.json`;
- `tools/frontend_phase_maps/fe_p0_test_map.json`,
  `tools/frontend_phase_maps/fe_p1_test_map.json`,
  `tools/frontend_phase_maps/fe_p2_test_map.json`, and
  `tools/frontend_phase_maps/fe_p3_test_map.json`;
- generated frontend ledgers for FE-P0 through FE-P3;
- package manifests and selected source/test files under `/packages/grid-adapter`,
  `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/test-utils`,
  and `/apps/web`;
- `Makefile` and `tools/task_surface.generated.mk` for exact target names.

Source limits and blockers:

- This file is a plan and is not behavior authority.
- The frontend guide and FE-P3 map match on the FE-P3 row IDs, evidence
  classes, and target names inspected here. If later inspection finds a
  row-level mismatch, record it as a blocker rather than reconciling silently.
- Existing frontend maps use compact endpoint-style owner ID lists for some
  guide ranges. This matches the current FE-P0 through FE-P2 map style and
  `make phase-ledger-drift` passes. If owner metadata is expanded or corrected,
  update the authored map first and regenerate ledgers with `make phase-ledgers`.
- Generated ledgers are downstream artifacts. They may confirm map rendering
  and drift status but cannot substitute for direct row-owned evidence.
- Visual fixture status is not fully aligned across the FE-P3 guide row and the
  visual-golden fixture matrix. This blocks `FE-V-P3-01` completion until the
  missing fixtures are added or exact blockers are recorded and the owner
  mapping is reconciled.
- FE-P3 rows currently remain blocked in the repository map. This plan must not
  claim FE-P3 implementation or completion.

## Phase Objective

By FE-P3 exit, the frontend must route workbook grid behavior through the
Cartulary grid adapter and view-schema contracts so that:

- rendered data rows are keyed by `record_id`;
- editable and queryable cells are keyed by `field_key`;
- vendor row/column coordinates translate to `record_id + field_key`;
- group, tree, loading, spacer, and presentation rows cannot produce
  mutation-capable events;
- renderer and editor behavior is explicit, deterministic, and cleaned up;
- sort, paste, drag fill, focus, selection, resize, tree/group, and scroll
  anchors use stable workbook identities;
- sparse patches preserve unchanged row object references where required and
  replace changed rows by `record_id`;
- direct RDG imports and mutation-capable imperative vendor APIs stay inside
  `/packages/grid-adapter`;
- `/apps/web` consumes adapter exports rather than depending directly on RDG.

## Implementation Scope

In scope for FE-P3:

- `/packages/grid-adapter` as the direct `react-data-grid` containment package;
- single RDG stylesheet ownership in `/packages/grid-adapter`;
- wrapper classes, CSS variables, documented stable classes, and accessible
  state attributes for grid styling;
- `record_id` row identity for data rows;
- `field_key` cell identity for editable, queryable, sortable, filterable, and
  groupable cells;
- renderer/editor registry precedence, deterministic fallback, explicit editor
  adapters, and lifecycle cleanup;
- editability derived from explicit editor adapters and contract writeability;
- translation for sort, paste, drag fill, focus, selection, edit, copy, resize,
  scroll-to-cell, tree expand/collapse, and anchor assertions;
- group, tree, loading, spacer, and presentation rows as non-record rows;
- sparse patch application by `record_id` with unchanged-row reference
  preservation where required;
- imperative API containment so mutation-capable vendor APIs are not exported
  outside the adapter boundary;
- `/apps/web` consumption of adapter and contract exports.

Explicitly out of scope for FE-P3:

- Timeline-specific mutation semantics;
- pending replay;
- same-field conflict resolver UI;
- evidence handle redemption;
- saved-view persistence;
- FE-P4 or later sync-engine behavior;
- Core 05 publication work unless a publication predicate is introduced.

## Sprint Checklist

- [x] Sprint 1: Readiness, registry, map, ledger, and FE-P2 handoff intake.
- [ ] Sprint 2: RDG containment, stylesheet ownership, and package boundary
  enforcement.
- [ ] Sprint 3: Row identity, cell identity, and presentation-row write gating.
- [ ] Sprint 4: Renderer/editor registry, writeability, and edit-entry
  contracts.
- [ ] Sprint 5: Coordinate translation, keyboard/focus, sort, paste, fill,
  resize, tree/group, and browser command helpers.
- [ ] Sprint 6: Sparse patch row preservation plus visual and accessibility
  readiness.
- [ ] Sprint 7: Closure, drift, regression, and FE-P4 handoff.

Checklist status after Sprint 1: readiness and FE-P2 intake complete. All FE-P3
map rows remain blocked.

## Global References

Behavior and conformance owners:

- Core 00: document set status, precedence, and implementation-conformance
  boundary.
- Core 01 Section 3.3.4: view-shaped read contract, `view_row_v1`,
  `view_row_patch_v1`, `record_id`, `row_version`, `cells[field_key]`, query
  sort/filter/group shape, and sparse row patch shape.
- Core 01 Section 7.4: view-schema registry, field registries, stable
  `field_key`, capability registries, write targets, and contract-derived
  behavior.
- Core 02: first-class record-envelope identity, source-state boundaries, and
  vocabulary support for `record_id`.
- Core 03 Sections 4.1, 4.13, and 14: autosave, keyboard/focus/paste/fill
  interaction, bulk commands, sorting, filtering, grouping, group-header
  behavior, and presentation-only grouping boundaries.
- Core 04: public boundary, security, conformance, and test-route boundaries.
- Core 05: claim-publication and benchmark reproducibility only when an
  explicit publication predicate exists.

Harness, implementation, and design support:

- `docs/testing-harness-nlspec.md` for frontend phase maps, generated ledgers,
  frontend row accounting, planned-phase explainability, target summaries, and
  harness evidence separation.
- `docs/domain.md` for vocabulary and concept-boundary support around
  `record_id`, `field_key`, `view_schema_id`, saved views, projections, and live
  workbook state.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` for FE-P3
  scope, row set, targets, evidence classes, fixture mapping, and phase
  completion rules.
- `docs/guides/cartulary-dev-guide.md` Sections 6.3, 6.7, 6.8, 6.9, and 6.10
  for implementation-support guidance on grid adapter boundaries, field
  renderer/editor adapters, CSS, row identity, and fragility guardrails.
- `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and
  `docs/guides/cartulary_visual_golden_maintenance.md` for design-direction
  and visual/accessibility fixture support only.
- `/packages/grid-adapter`, `/packages/view-contracts`,
  `/packages/ui-contracts`, `/packages/test-utils`, and `/apps/web` for
  FE-P3 implementation surfaces.

## Evidence Layer Matrix

The row set below is exactly the FE-P3 row set present in the frontend guide and
repository FE-P3 map at plan creation.

| Row | Layer | Evidence class | Controlling claim | Intended targets | Direct row-owned evidence required |
| --- | --- | --- | --- | --- | --- |
| `FE-U-P3-01` | Unit | `product_conformance` | Missing or duplicate `record_id` fails before mutation-capable rendering; group/loading/spacer/presentation rows cannot emit write events. | `make frontend-unit` | Unit scenarios with row ID in evidence accounting, plus failures before any mutation-capable path is reachable. |
| `FE-U-P3-02` | Unit | `product_conformance` | Vendor row/column coordinates translate to `record_id + field_key` for edit, copy, paste, drag fill, selection, focus, and anchor assertions. | `make frontend-unit` | Unit scenarios proving every coordinate translation path returns stable workbook identities or clears intentionally. |
| `FE-U-P3-03` | Unit | `product_conformance` | Editability derives from explicit editor adapters and contract writeability; `editable=true` alone never enters edit mode. | `make frontend-unit` | Unit scenarios against editor resolution, writeability, read-only fields, and RDG editable-only cases. |
| `FE-U-P3-04` | Unit | `implementation_support` | Renderer/editor registry precedence, deterministic fallback, and lifecycle cleanup. | `make frontend-unit` | Unit scenarios proving precedence, fallback, subscription cleanup, portal cleanup, observer cleanup, timer cleanup, and stale-row-reference cleanup. |
| `FE-I-P3-01` | Integration | `product_conformance` | Sparse patches preserve unchanged row object references and replace changed rows by `record_id`. | `make frontend-unit`; `make browser-e2e-support` | Integration-style frontend unit or support evidence with direct row accounting and object-reference assertions. |
| `FE-B-P3-01` | Browser integration | `product_conformance` | Browser command helpers cover sort, filter, group, resize, paste, drag fill, scroll-to-cell, tree expand/collapse, and anchor assertions. | `make browser-e2e-support`; `make browser-e2e-webserver-backed` | Playwright/support scenarios named by the FE-P3 map and retained in `frontend-row-accounting.json`. |
| `FE-V-P3-01` | Visual regression | `design_direction` | Visual fixtures cover frozen column, resize handle, drag-fill handle, edit cell, tree/group row, grouped result, row gutter presence, and empty successful query. | `make browser-e2e-visual` | Visual row accounting plus deterministic fixture/golden status. Missing fixture or owner-map mismatch remains a blocker. |
| `FE-A11Y-P3-01` | Accessibility | `design_direction` | Grid cells, editors, group/tree rows, active cell, edit mode, disabled/read-only state, and blocked actions are keyboard accessible and announced without color-only signals. | `make browser-e2e-a11y-preflight` in the current map | Preflight evidence while blocked, or an intentional map/target promotion with generated ledger refresh before implemented accessibility closure. |
| `FE-S-P3-01` | Support | `implementation_support` | Direct RDG import containment and single stylesheet ownership in `/packages/grid-adapter`. | `make frontend-import-boundary-check`; `make lint-biome` | Support-target evidence that `/apps/web` has no direct RDG dependency, `/packages/grid-adapter` owns RDG imports, and stylesheet ownership is singular. |

Evidence rules:

- Do not mark any row complete from generated ledger text alone.
- Do not mark any row complete from retained artifacts unless the row-owned
  target emitted the row in `frontend-row-accounting.json`.
- Do not substitute broad regression suites, support checks, visual evidence, or
  accessibility evidence for product-conformance rows.
- Do not substitute product-conformance evidence for design-direction or
  implementation-support rows.
- Rows with `claim_status="blocked"` or owner-lookup TODOs are excluded from
  authoritative completion.

## Dependencies and Prerequisites

- FE-P0, FE-P1, and FE-P2 must remain explainable and their regression status
  must be recorded before FE-P3 closure.
- The FE-P3 frontend registry entry, authored phase map, and generated ledger
  must exist and agree, or exact blockers must be recorded.
- Any metadata change to `tools/frontend_phase_maps/fe_p3_test_map.json` or
  `tools/frontend_phase_registry.json` must be followed by `make phase-ledgers`
  and `make phase-ledger-drift`.
- `/packages/grid-adapter` must remain the only direct RDG source import owner.
- `/apps/web` must consume `/packages/grid-adapter` exports and must not import
  `react-data-grid` directly.
- `/packages/view-contracts` must expose the contract fields needed to derive
  visible fields, sortability, filterability, groupability, writeability,
  editor eligibility, and fallback rendering from stable `field_key` values.
- `/packages/ui-contracts` and `/packages/test-utils` must provide stable
  selector/test-helper contracts without relying on visible labels, DOM nesting,
  RDG generated classes, or render order.
- Browser validation requires the repository browser-owned stack and its
  service-backed target prerequisites.
- Visual validation requires deterministic fixture state and visual-golden
  ownership reconciliation for missing FE-P3 fixtures.

## Public Interfaces and Deliverables

Expected authored deliverables during implementation:

- `FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md` as the plan and handoff record.
- Authored package/app changes under `/packages/grid-adapter`,
  `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/test-utils`,
  and `/apps/web` as required by row-owned tests.
- Authored FE-P3 phase-map metadata changes only when needed to add scenario
  titles, fix ownership, record exact blockers, or promote rows.
- Browser helper and selector contracts that target `record_id`, `field_key`,
  `view_schema_id`, `saved_view_id`, and closed tokens instead of labels or
  render order.
- Visual fixture/golden updates only under visual-golden maintenance rules.
- Accessibility summaries or preflight summaries only from the intended
  repository targets.

Generated or tool-managed deliverables:

- Generated frontend ledgers under
  `docs/testing/frontend_phase_coverage_ledgers/` from `make phase-ledgers`
  after metadata changes.
- Target-owned `frontend-row-accounting.json` artifacts from FE-P3 row-owned
  validation commands.
- Tool-run summaries, stdout logs, stderr logs, Playwright reports, visual
  artifacts, and accessibility summaries under retained result roots.
- No hand edits to generated ledgers, generated protocol TypeScript, lockfiles,
  or checksum files.

## Sprint 1: Readiness, Registry, Map, Ledger, And FE-P2 Handoff Intake

Status: complete for readiness/intake on 2026-05-31.

Objective: establish FE-P3's metadata, naming, package, command, and inherited
phase context before implementation.

Non-goals: no grid behavior implementation; no row completion claim from
metadata, generated ledgers, or broad source inspection.

Test-first sequence:

- Verify naming precedent with `rg --files -g 'FRONTEND_PHASE*_IMPLEMENTATION_PLAN.md'`.
- Inspect the FE-P3 guide row set and compare it to
  `tools/frontend_phase_maps/fe_p3_test_map.json`.
- Run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3`.
- Run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`,
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`, and
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2`.
- Run `make phase-ledger-drift`.
- Inspect `/packages/grid-adapter`, `/packages/view-contracts`,
  `/packages/ui-contracts`, `/packages/test-utils`, and `/apps/web` manifests
  and FE-P3 import surfaces.

Implementation planning tasks:

- Record exact map/guide mismatches as blockers.
- Record missing package manifests, missing targets, missing ledgers, or stale
  generated artifacts as blockers.
- Preserve FE-P2 handoff facts only when locally reverified.
- Keep FE-P3 registry `planned` unless a separate implementation change
  promotes row evidence and executable scheduling in the same change.

Validation commands:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3`
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2`
- `make phase-ledger-drift`
- `git diff --check`

Sprint 1 execution record:

- Naming precedent was verified with
  `rg --files -g 'FRONTEND_PHASE*_IMPLEMENTATION_PLAN.md'`; the root contains
  FE-P0 through FE-P3 plan files.
- The guide row inventories match the authored frontend phase maps for FE-P0,
  FE-P1, FE-P2, and FE-P3. No registry or map metadata change was required, so
  `make phase-ledgers` was not run as a standalone Sprint 1 command.
- Active package manifests exist for `apps/web`, `packages/grid-adapter`,
  `packages/view-contracts`, `packages/ui-contracts`, `packages/test-utils`,
  and `packages/protocol-ts`. `packages/ui/package.json` is absent, but
  `packages/ui` is future-reserved via `.gitkeep` and is not an active package
  manifest blocker.
- RDG containment was verified by source inspection: `react-data-grid` and
  `react-data-grid/lib/styles.css` are imported from
  `packages/grid-adapter/src/index.tsx`; `/apps/web` imports
  `@cartulary/grid-adapter`, not RDG directly.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3` passed and reports
  FE-P3 as `planned`, explainable, non-executable, and blocked for all nine
  FE-P3 rows.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`,
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`, and
  `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2` passed and report
  those phases as `planned`, explainable, and non-executable.
- `make phase-ledger-drift` passed with run root
  `.cartulary/test-results/20260531T052900Z-p1219605` and summary
  `.cartulary/test-results/20260531T052900Z-p1219605/phase-ledger-drift/tool-run-summary.json`.
- `git diff --check` passed with no output after the Sprint 1 documentation
  updates.
- `make agent-finalize` passed with run root
  `.cartulary/test-results/20260531T052956Z-p1220952` and summary
  `.cartulary/test-results/20260531T052956Z-p1220952/agent-finalize/tool-run-summary.json`;
  generated maintenance artifacts were unchanged and retained-run maintenance was
  skipped because `RESULTS_DIR` was unset.

FE-P2 handoff intake verified from retained row-owned artifacts:

- FE-P2 registry remains `planned`; map and ledger paths are present and
  explainable.
- FE-P2 has seven rows, all `claim_status="implemented"` in the map and
  generated ledger.
- Unit row accounting was read from
  `.cartulary/test-results/20260531T031120Z-p874196/frontend-unit/frontend-row-accounting.json`;
  `FE-U-P2-01` and `FE-U-P2-02` show `closure_status="closed"`.
- Browser row accounting was read from
  `.cartulary/test-results/20260531T031139Z-p875748/browser-e2e-webserver-backed/frontend-row-accounting.json`;
  `FE-B-P2-01`, `FE-B-P2-02`, and `FE-E-P2-01` show
  `closure_status="closed"`.
- Visual row accounting was read from
  `.cartulary/test-results/20260531T031306Z-p883678/browser-e2e-visual/frontend-row-accounting.json`;
  `FE-V-P2-01` shows `closure_status="closed"` as design-direction evidence.
- Accessibility row accounting was read from
  `.cartulary/test-results/20260531T031357Z-p888634/browser-e2e-a11y/frontend-row-accounting.json`;
  `FE-A11Y-P2-01` shows `closure_status="closed"` as design-direction
  evidence.
- Accessibility summary
  `.cartulary/test-results/20260531T031357Z-p888634/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`
  reports `schema_id="cartulary.frontend_accessibility_summary.v2"` and zero
  violations.
- FE-P3 accessibility preflight summary
  `.cartulary/test-results/20260531T031420Z-p891647/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json`
  lists `FE-A11Y-P3-01` as blocked preflight smoke only.

Sprint 1 blockers and follow-up:

| Item | Ownership | Blocks FE-P3 completion | Affects FE-P4 handoff | Minimum follow-up |
| --- | --- | --- | --- | --- |
| All FE-P3 rows remain blocked with no Sprint 1 row-owned implementation evidence. | `FE-P3-owned` | yes | yes, if unresolved | Later FE-P3 sprints must emit direct row-owned accounting before promotion. |
| `FE-V-P3-01` visual fixture ownership and missing-fixture reconciliation remains open. | `FE-P3-owned/design-direction` | yes | yes, if unresolved | Reconcile the visual fixture matrix and close only from `browser-e2e-visual` frontend row accounting. |
| `FE-A11Y-P3-01` maps to preflight smoke, not implemented accessibility evidence. | `FE-P3-owned/design-direction` | yes | yes, if unresolved | Keep blocked or intentionally promote target/metadata, regenerate ledgers, and validate. |

Sprint 1 does not close any FE-P3 row.

## Sprint 2: RDG Containment, Stylesheet Ownership, And Package Boundary Enforcement

Status: planned.

Primary row: `FE-S-P3-01`.

Objective: make the RDG dependency boundary auditable and enforce that
`/apps/web` consumes Cartulary adapter exports.

Non-goals: no editor semantics, sparse-patch behavior, browser helper
completion, or visual/a11y row closure.

Test-first sequence:

- Add or update boundary tests before changing imports.
- Search all authored frontend sources for `react-data-grid` and
  `react-data-grid/lib/styles.css`.
- Run `make frontend-import-boundary-check`.
- Run `make lint-biome`.

Implementation tasks:

- Keep direct RDG imports inside `/packages/grid-adapter`.
- Keep the RDG stylesheet import singular and owned by the adapter entrypoint
  that precedes workbook grid rendering.
- Ensure `/apps/web` imports only adapter, contract, and test-helper exports.
- Prevent mutation-capable RDG imperative APIs from escaping adapter exports.
- Keep CSS integration on wrapper classes, CSS variables, documented stable
  classes, and accessible state attributes.

Row-owned validation:

- `FE-S-P3-01` closes only from direct support-target evidence in
  `make frontend-import-boundary-check` and `make lint-biome`, with exact
  artifact paths recorded.

Blockers to record:

- Any direct RDG import outside `/packages/grid-adapter`.
- More than one RDG stylesheet import.
- Missing `frontend-import-boundary-check` target.
- Lint failure caused by boundary implementation.

## Sprint 3: Row Identity, Cell Identity, And Presentation-Row Write Gating

Status: planned.

Primary rows: `FE-U-P3-01` and the row/cell-identity subset of
`FE-U-P3-02`.

Objective: fail closed before mutation-capable rendering when row identity is
unsafe and ensure presentation rows cannot produce write events.

Non-goals: no Timeline mutation submission, pending replay, same-field conflict
resolver, evidence handles, or saved-view persistence.

Test-first sequence:

- Add failing unit tests for missing `record_id`, whitespace `record_id`, and
  duplicate `record_id` before mutation-capable rendering.
- Add failing unit tests that group, tree, loading, spacer, and presentation
  rows have no `record_id` mutation target and cannot emit edit, paste, drag
  fill, evidence attach, destructive record, or entity-resolution events.
- Add failing unit tests for `field_key` cell identity on editable and
  queryable cells.
- Run `make frontend-unit`.

Implementation tasks:

- Key data rows by `record_id`, not visible index or row label.
- Key editable/queryable cells by `field_key`, not column label or render order.
- Normalize presentation rows as non-record rows with no mutation-capable
  target.
- Ensure group/tree/loading/spacer rows cannot call mutation-capable callbacks.
- Ensure any write-capable path checks for stable `record_id + field_key`
  before dispatch.

Row-owned validation:

- `FE-U-P3-01` closes only when `make frontend-unit` emits direct row-owned
  evidence or an exact blocker.
- `FE-U-P3-02` remains open unless all coordinate translation paths are also
  covered in Sprint 5.

Blockers to record:

- Missing pre-mutation invariant checks.
- Presentation rows with mutation-capable callbacks.
- Write events keyed by visible row index, label, or DOM order.

## Sprint 4: Renderer/Editor Registry, Writeability, And Edit-Entry Contracts

Status: planned.

Primary rows: `FE-U-P3-03` and `FE-U-P3-04`.

Objective: make renderer/editor resolution explicit, contract-derived,
deterministic, and cleanup-safe.

Non-goals: no FE-P4 sync engine, no pending replay, and no same-field conflict
resolver UI.

Test-first sequence:

- Add failing unit tests for renderer registry precedence by exact `field_key`
  and declared fallback.
- Add failing unit tests for deterministic renderer fallback when no specialized
  renderer is available.
- Add failing unit tests for explicit editor adapters and contract writeability.
- Add failing unit tests proving RDG `editable=true` alone never enters edit
  mode.
- Add failing unit tests for cleanup of subscriptions, portals, observers,
  timers, and stale row references.
- Run `make frontend-unit`.

Implementation tasks:

- Resolve display renderers by `field_key`, view contract, and explicit
  registry rules.
- Resolve editor adapters only from explicit adapter declarations plus contract
  writeability.
- Ensure writable-through-other-surface fields are not automatically grid
  editable.
- Keep editor output mapped to field-keyed mutation payloads without
  dispatching FE-P4 sync behavior.
- Tear down renderer/editor resources on unmount, row replacement, editor close,
  and registry replacement.

Row-owned validation:

- `FE-U-P3-03` closes only from unit evidence for writeability and edit-entry
  rules.
- `FE-U-P3-04` closes only from unit evidence for registry precedence,
  fallback, and cleanup.

Blockers to record:

- Any edit entry path controlled solely by RDG `editable=true`.
- Missing explicit editor adapter for a writable grid field.
- Lifecycle leaks or stale row references after editor/renderer teardown.

## Sprint 5: Coordinate Translation, Keyboard/Focus, Sort, Paste, Fill, Resize, Tree/Group, And Browser Helpers

Status: planned.

Primary rows: `FE-U-P3-02` and `FE-B-P3-01`.

Objective: prove all vendor-coordinate and browser-helper operations translate
through stable workbook identities.

Non-goals: no authoritative local row reordering as sort, no FE-P4 mutation
queue, no saved-view persistence, and no conflict resolver UI.

Test-first sequence:

- Add failing unit tests for edit, copy, paste, drag fill, selection, focus,
  and anchor translation from vendor row/column coordinates to
  `record_id + field_key`.
- Add or update browser command helpers for sort, filter, group, resize, paste,
  drag fill, scroll-to-cell, tree expand/collapse, and anchor assertions.
- Add browser scenarios named by the FE-P3 map for `FE-B-P3-01`.
- Run `make frontend-unit`.
- Run `make browser-e2e-support`.
- Run `make browser-e2e-webserver-backed`.

Implementation tasks:

- Convert header sort commands to view-query `sort[]` entries keyed by
  `field_key`.
- Reject local row reordering as authoritative sort.
- Keep filter and grouping helpers keyed by field contract values.
- Ensure paste and drag fill target explicit stable row and field anchors.
- Keep focus and selection restoration keyed by `record_id + field_key`.
- Ensure scroll-to-cell and anchor assertions use stable selectors from
  `/packages/ui-contracts` and helpers from `/packages/test-utils`.
- Ensure tree/group expand/collapse is client-local presentation state and not
  a record mutation.

Row-owned validation:

- `FE-U-P3-02` closes only when unit evidence covers all coordinate translation
  operations named in the row.
- `FE-B-P3-01` closes only when browser/support targets emit direct row-owned
  `frontend-row-accounting.json` evidence for all named helper operations.

Blockers to record:

- Helper behavior that relies on visible labels, rendered row order, CSS
  classes, or DOM nesting.
- Missing scenario titles for browser-backed row closure.
- Browser target pass without FE-P3 row accounting.

## Sprint 6: Sparse Patch Row Preservation Plus Visual And Accessibility Readiness

Status: planned.

Primary rows: `FE-I-P3-01`, `FE-V-P3-01`, and `FE-A11Y-P3-01`.

Objective: close sparse-patch identity behavior separately from design-direction
visual and accessibility readiness.

Non-goals: visual and accessibility evidence do not become product conformance,
and neither activates Core 05.

Test-first sequence:

- Add failing integration-style tests for sparse patch application that
  preserves unchanged row object references and replaces changed rows by
  `record_id`.
- Add browser support evidence for patch/reference behavior when the row map
  requires it.
- Add or reconcile visual fixtures for frozen column, resize handle, drag-fill
  handle, edit cell, tree/group row, grouped result, row gutter presence, and
  empty successful query.
- Add accessibility/preflight evidence for grid cells, editors, group/tree
  rows, active cell, edit mode, disabled/read-only state, and blocked actions.
- Run `make frontend-unit`.
- Run `make browser-e2e-support`.
- Run `make browser-e2e-visual`.
- Run `make browser-e2e-a11y-preflight` while `FE-A11Y-P3-01` remains blocked
  in the map. If the row is promoted to implemented accessibility evidence,
  update the map target intentionally, regenerate ledgers, and run the promoted
  target.

Implementation tasks:

- Apply sparse patches by `record_id`.
- Preserve row object references where public row data is unchanged and the row
  remains present.
- Replace changed rows with new objects keyed by `record_id`.
- Remove rows by `record_id` and handle invalidation without visible-index
  identity.
- Make visual fixtures deterministic under visual-golden maintenance rules.
- Make accessibility evidence announce active cell, edit mode, blocked actions,
  read-only/disabled state, and group/tree row affordances without color-only
  signals.

Row-owned validation:

- `FE-I-P3-01` closes only from direct integration/support evidence.
- `FE-V-P3-01` closes only from direct visual row evidence and reconciled
  fixture status.
- `FE-A11Y-P3-01` closes only from the intended accessibility target after the
  map and guide completion rules agree; blocked preflight smoke alone does not
  imply implemented accessibility closure.

Blockers to record:

- Missing sparse-patch object-reference assertions.
- Missing or unstable visual fixtures.
- Visual fixture ownership conflict between FE-P3 guide and visual-golden
  matrix.
- Accessibility target mismatch between blocked preflight and implemented-row
  evidence.

## Sprint 7: Closure, Drift, Regression, And FE-P4 Handoff

Status: planned.

Objective: close FE-P3 only with direct row-owned evidence or exact blockers,
then prepare FE-P4 handoff.

Non-goals: no completion claim from generated ledgers, retained artifacts,
aggregate commands, visual/a11y evidence, or support checks outside their own
rows.

Test-first sequence:

- Rerun every FE-P3 row-owned target required by the current FE-P3 map.
- Read each target-owned `frontend-row-accounting.json` artifact.
- Rerun FE-P0, FE-P1, and FE-P2 regression commands needed by their row-owned
  status and record artifact paths.
- Run generated-artifact, generate-drift, import-boundary, ledger-drift, lint,
  typecheck, and whitespace checks.
- Run `make agent-finalize` before broader verification when repository
  procedure requires end-of-run maintenance.
- Run `make check` only when repository completion rules require broad closure.

Validation commands:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3`
- `make frontend-typecheck`
- `make frontend-unit`
- `make browser-e2e-support`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight`
- `make frontend-import-boundary-check`
- `make lint-biome`
- `make generated-artifact-policy-check`
- `make generate-drift`
- `make phase-ledgers` when registry or map metadata changed
- `make phase-ledger-drift`
- `git diff --check`
- `git diff --cached --check` when staged changes exist
- `make check` only when repository completion rules require broad closure

`make check` requirement:

- Required when FE-P3 changes are being presented as broad repository closure,
  when repository completion rules require the developer gate, when scheduler
  metadata/check topology changes, or when a maintainer asks for the full gate.
- Intentionally skipped for plan-only changes and for narrow FE-P3 closure when
  all direct row-owned targets, regression commands, drift checks, lint/type
  checks, import-boundary checks, and whitespace checks have passed and no repo
  rule requires the broad gate.
- Never a substitute for direct FE-P3 row-owned evidence.

Closure records must include:

- every command result;
- run root, summary JSON, stdout log, and stderr log paths when available;
- row-owned evidence paths;
- generated-artifact and drift status;
- FE-P0, FE-P1, and FE-P2 regression status;
- unresolved blockers and owner-lookup TODOs;
- whether Core 05 remained inactive or which explicit publication predicate
  activated it.

## Validation Command Inventory

All target names below exist in the inspected repository task surface.

| Command | FE-P3 use | Substitution rule |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P3` | Registry, map, ledger, evidence-class, and target mapping readiness. | Does not close rows. |
| `make frontend-typecheck` | Type-level regression and broad frontend readiness. | Does not close a row unless row-owned accounting and map target require it. |
| `make frontend-unit` | Unit rows and frontend integration-style rows mapped to unit execution. | Closes only mapped rows with direct accounting. |
| `make browser-e2e-support` | Support/browser helper evidence for mapped FE-P3 rows. | Does not replace product browser evidence unless mapped. |
| `make browser-e2e-webserver-backed` | Browser integration through public browser-facing boundary. | Required for mapped browser product-conformance evidence. |
| `make browser-e2e-visual` | Design-direction visual evidence for `FE-V-P3-01`. | Does not close product-conformance rows. |
| `make browser-e2e-a11y-preflight` | Current blocked-row accessibility smoke for `FE-A11Y-P3-01`. | Does not imply implemented accessibility closure unless the map and guide target rules explicitly allow it. |
| `make frontend-import-boundary-check` | RDG import containment and stylesheet ownership support evidence. | Closes only support rows mapped to it. |
| `make lint-biome` | Authored frontend lint support. | Does not close product-conformance rows. |
| `make generated-artifact-policy-check` | Generated-source policy and generated-path edit guard. | Does not close rows unless mapped by a support row. |
| `make generate-drift` | Generated-output equivalence. | Does not replace row-owned evidence. |
| `make phase-ledgers` | Regenerate generated ledgers after frontend registry/map metadata changes. | Generated output only; never hand-edit ledgers. |
| `make phase-ledger-drift` | Validate generated ledgers match maps. | Does not close rows. |
| `git diff --check` | Whitespace validation for working-tree changes. | Required before handoff. |
| `git diff --cached --check` | Whitespace validation for staged changes. | Required only when staged changes exist. |
| `make check` | Broad developer gate when repository completion rules require it. | Never substitutes for direct row-owned evidence. |

If a future guide revision expects a command not present in the repository task
surface, record the missing target as a blocker with the affected FE-P3 row ID.

## Dependencies and Regression Recording

Before FE-P3 completion judgment, record:

- FE-P0 registry, map, ledger, command, and row-accounting status;
- FE-P1 registry, map, ledger, command, and row-accounting status;
- FE-P2 registry, map, ledger, command, and row-accounting status;
- whether any earlier-phase row regressed and who owns the regression;
- exact artifact paths for every rerun command;
- whether broad `make check` was required, run, skipped, or blocked.

Earlier-phase green status does not close FE-P3 rows. FE-P3 completion also
does not repair earlier-phase blocker ownership unless the same change reruns
and records those row-owned commands.

## Blocker Recording Rules

Every blocker must record:

- exact failed or missing command, file, target, fixture, scenario, artifact,
  package, or owner reference;
- exact failing target, scheduler unit, test, or scenario when available;
- result root, run ID, run root, summary JSON, stdout log, and stderr log paths
  when available;
- affected FE-P3 row ID;
- failure class and failure reason when exposed;
- ownership classification: `FE-P3-owned`, `FE-P0-regression-owned`,
  `FE-P1-regression-owned`, `FE-P2-regression-owned`,
  `support-tooling-owned`, `harness-owned`, `infra-owned`, or
  `outside FE-P3`;
- minimum follow-up required;
- whether the blocker prevents FE-P3 completion;
- whether the blocker affects FE-P4 handoff.

Generated ledgers, retained artifacts, broad regression suites, visual
evidence, accessibility evidence, support checks, and aggregate commands are
not blockers by themselves unless they are the intended row target or expose a
missing or failing prerequisite.

## Binary Exit Criteria

FE-P3 is complete only when all are true:

- `FRONTEND_PHASE3_IMPLEMENTATION_PLAN.md` exists and states that it is not
  behavior authority.
- FE-P3 registry, map, and ledger are present, or exact blockers are recorded.
- Every FE-P3 row appears exactly once in the FE-P3 phase map.
- FE-P3 map metadata preserves evidence-class separation.
- Generated ledgers are regenerated, not hand-edited.
- Each FE-P3 row has direct row-owned evidence or an exact blocker.
- FE-P0, FE-P1, and FE-P2 regression status is recorded.
- `/apps/web` does not import `react-data-grid` directly.
- `/packages/grid-adapter` owns direct RDG imports and stylesheet ownership.
- The RDG stylesheet is imported exactly once before workbook-grid rendering, or
  an exact blocker records why this is not yet true.
- Data rows are keyed by `record_id`.
- Editable and queryable cells are keyed by `field_key`.
- Vendor coordinates translate to `record_id + field_key`.
- Missing or duplicate `record_id` fails before mutation-capable rendering.
- Group, loading, spacer, and presentation rows cannot emit mutation-capable
  events.
- Editability requires explicit editor adapters and contract writeability.
- RDG `editable=true` alone does not enable edit mode.
- Sort, paste, drag fill, selection, focus, resize, tree/group, and scroll
  anchors use stable workbook identities, not visible labels or render order.
- Sparse patches preserve unchanged row object references where required and
  replace changed rows by `record_id`.
- Renderer/editor cleanup removes subscriptions, portals, observers, timers,
  and stale row references.
- Imperative mutation-capable vendor APIs are unavailable outside the adapter
  boundary.
- Visual and accessibility evidence remain design-direction evidence unless
  Core 05 publication predicates are explicitly active.
- Generated-artifact policy, generate-drift, phase-ledger drift,
  import-boundary, lint/typecheck, and whitespace checks pass or are precisely
  blocked.
- Unresolved owner-lookup TODOs are recorded and excluded from completion.

Current plan-creation closure judgment: not complete. All FE-P3 rows are
blocked in the current map.

## Handoff Requirements For FE-P4

FE-P4 must receive a written FE-P3 handoff containing, at minimum:

- grid-adapter package boundary status;
- RDG import and stylesheet ownership status;
- `record_id` row-identity status;
- `field_key` cell-identity status;
- presentation-row write-gating status;
- renderer/editor registry status;
- editor writeability and edit-entry status;
- coordinate translation status;
- browser command helper status;
- keyboard/focus status;
- sparse-patch row-reference status;
- visual fixture status;
- accessibility status;
- selector/test-id status;
- generated-artifact and drift status;
- FE-P0 through FE-P2 regression status;
- command blockers and artifact paths;
- unresolved owner-lookup TODOs.

FE-P4 must not infer Timeline mutation readiness, pending replay readiness,
same-field conflict resolver readiness, evidence handle readiness, or
saved-view persistence readiness from FE-P3 unless those topics are explicitly
covered by later owner rows.
