# FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md

## Summary

Create and maintain this file as the FE-P2 execution roadmap, progress marker,
and FE-P3 handoff aid for `FE-P2: Workbook Shell And Startup Surface`.

This plan is not behavior authority and must not mark FE-P2 rows complete without
direct row-owned evidence. `docs/guides/cartulary_frontend_implementation_testing_guide.md`
controls the FE-P2 row set; Core 00 through Core 04 remain
implementation-conformance authority. Core 05 stays inactive unless a
claim-bearing timed or fixture-sensitive publication predicate is explicitly
introduced.

Reconnaissance verified this natural target file was absent before creation, so
this document records the new root-level FE-P2 implementation plan.

## Authority Model

- Product-conformance rows rely only on Core 00 through Core 04 or a future
  adopted NLSpec.
- `docs/testing-harness-nlspec.md` owns harness mechanics only.
- `docs/domain.md` is vocabulary and concept-boundary support, not behavior
  authority.
- UI/UX guide, `docs/design.md`, and visual-golden guide provide
  design-direction and visual-maintenance inputs only.
- Design-direction, implementation-support, and product-conformance evidence
  must stay separate in phase maps, ledgers, summaries, and plan text.
- Visual and accessibility evidence must not activate Core 05 by default.

## Current Repo Status

Locally verified facts:
- Before this document was created, `FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md` did
  not exist.
- `tools/frontend_phase_registry.json` contains `FE-P2` in namespace `frontend`,
  `status="planned"`, map `tools/frontend_phase_maps/fe_p2_test_map.json`,
  ledger `docs/testing/frontend_phase_coverage_ledgers/fe_p2_coverage_ledger.md`,
  and dependencies `FE-P0`, `FE-P1`.
- `tools/frontend_phase_maps/fe_p2_test_map.json` contains exactly these seven
  rows once each: `FE-U-P2-01`, `FE-U-P2-02`, `FE-B-P2-01`, `FE-B-P2-02`,
  `FE-E-P2-01`, `FE-V-P2-01`, `FE-A11Y-P2-01`.
- At plan creation, all FE-P2 rows were `claim_status="blocked"`.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2` passed after Sprint
  1 metadata correction and reports FE-P2 as planned, explainable, and
  non-executable.
- `make phase-ledgers` regenerated the FE-P2 generated ledger from the authored
  map with run root `.cartulary/test-results/20260530T173411Z-p3939879` and
  summary `.cartulary/test-results/20260530T173411Z-p3939879/phase-ledgers/tool-run-summary.json`.
- `make phase-ledger-drift` passed with run root
  `.cartulary/test-results/20260530T173423Z-p3940471` and summary
  `.cartulary/test-results/20260530T173423Z-p3940471/phase-ledger-drift/tool-run-summary.json`.
- `git diff --check` passed with no output. `git diff --cached --check` was not
  required because no staged changes existed.
- `make agent-finalize` passed with run root
  `.cartulary/test-results/20260530T174430Z-p3953679` and summary
  `.cartulary/test-results/20260530T174430Z-p3953679/agent-finalize/tool-run-summary.json`.
  It reported generated artifacts unchanged and skipped retained-run maintenance
  because `RESULTS_DIR` was unset.
- `/packages/ui` exists only as `packages/ui/.gitkeep`; it is inactive until a
  manifest-backed package is introduced.
- `WorkbookShell` currently has built-in tab selectors, a native `select` for
  system views, backend `/workbook-startup` consumption, and selected saved-view
  `sheet_ref` preservation.
- Current shell composition has Sprint 3 row-owned browser evidence for
  `FE-B-P2-01`: required shell slots remain under one workbook shell root and
  navigation remains on the same workbook route. Sprint 4 row-owned browser
  evidence for `FE-B-P2-02` now verifies the custom `System views` switcher
  keyboard entry, roving focus, selection, dismissal, and focus restoration. Sprint
  5 row-owned browser evidence for `FE-E-P2-01` now verifies active-surface
  saved-view placement, `scope='system'` saved-view non-conflation, same-shell
  contract-backed system-view navigation, and base-`view_schema_id` query
  routing after saved-view selection. FE-P2 is still not complete:
  visual/accessibility readiness remains Sprint 6 work.

Inherited FE-P0 handoff state:
- `FE-P0` remains `planned` but all six FE-P0 rows are `implemented` in the
  current map/ledger.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` passed during Sprint
  1 readiness validation.

Inherited FE-P1 handoff state:
- `FE-P1` remains `planned` but all five FE-P1 rows are `implemented` in the
  current map/ledger, with product-conformance, design-direction, and
  implementation-support rows still separated.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1` passed during Sprint
  1 readiness validation.
- FE-P1 plan Sprint 7 is still unchecked while the binary criteria section
  records satisfied row evidence through Sprint 6. Treat this as stale handoff
  bookkeeping until the FE-P1 plan is reconciled or relevant row-owned checks are
  rerun or explicitly accepted; no FE-P2 row should be promoted on this
  bookkeeping state alone.

Assumptions:
- FE-P2 implementation may edit authored app/test/package files and authored
  frontend phase maps.
- Generated ledgers are regenerated through `make phase-ledgers` and never
  hand-edited.
- Browser validation uses the repository's existing service-backed Playwright
  targets.

Initial missing evidence at plan creation:
- No direct FE-P2 row-owned evidence had been collected.
- No FE-P2-specific visual golden named for a phase-2 shell capture was found;
  visual goldens currently start with `v-3-*` files even though `FE-VFIX-01` is
  marked current.

## Startup And Registry Hardening Addendum

This addendum records post-closure hardening for `FE-U-P2-01` and
`FE-U-P2-02`. It does not reopen browser, visual, or accessibility FE-P2 rows.

Hardening scope:
- `FE-U-P2-01` adds direct unit evidence for unsupported startup
  `sheet_ref.kind` handling, deleted saved-view pointers represented as
  `saved_view_not_found`, `required_reference_pack_unavailable` preservation,
  and backend startup responses with non-standardized `selected_view_schema_id`.
- `FE-U-P2-02` adds direct unit evidence for relabeled registry contracts,
  optional standardized surface absence, missing required-contract failure, and
  `requiredReferencePackKeys` parser/registry exposure.
- Core 01/Core 03 now document startup request-validation reason codes and
  successful cleared-pointer reason codes. `saved_view_deleted` is intentionally
  not introduced in FE-P2.
- Runtime proof that `scope='system'` saved views do not replace canonical
  system views is now covered by Sprint 5 / `FE-E-P2-01`; the system saved-view
  fixture path remains harness-owned and is not product API behavior.

Resolved Sprint 1 metadata defects:
- `FE-B-P2-01` and `FE-B-P2-02` now keep design-direction `R2-AC-*` IDs in
  `support_or_design_ac_ids` rather than collapsed unprefixed Core `AC-*`
  fields.
- `FE-U-P2-01`, `FE-B-P2-02`, and `FE-E-P2-01` owner references now match the
  inspected Core owner sections.
- The generated FE-P2 ledger was refreshed with `make phase-ledgers`; it was not
  hand-edited.

## Source Limits

Inspected sources include the requested phase plans, frontend guide, base
implementation guide, dev guide, UI/UX guide, design doc, visual golden guide,
harness NLSpec, domain doc, frontend registry/maps/ledgers for FE-P0 through
FE-P2, and relevant `/apps/web`, `/packages/ui-contracts`, `/packages/test-utils`,
and `/packages/ui` files.

Source limits and blockers:
- Missing target plan file: resolved by creating this document.
- At Sprint 1, the FE-P2 phase map/ledger existed and all FE-P2 rows remained
  blocked; later sprint sections record row-owned evidence and blockers without
  treating stale ledger or checklist status as behavior authority.
- FE-P2 map metadata separation defects were fixed and regenerated during Sprint
  1, but they are readiness evidence only and do not complete any FE-P2 row.
- Visual fixture status is inconsistent enough to block `FE-V-P2-01` completion
  until an owned FE-P2 shell fixture/golden is added or the fixture claim is
  precisely corrected.
- FE-P1 handoff bookkeeping remains stale until the FE-P1 plan reconciles Sprint
  7 or row-owned checks are rerun or explicitly accepted.
- `/packages/ui` is not an active implementation surface.
- Existing app-shell code is implementation-shape evidence only, not FE-P2
  completion evidence.

## Phase Objective

By FE-P2 exit, the frontend must expose one continuous workbook shell and startup
surface with:

- ordered built-in tabs: Timeline, Hosts, Identities, Evidence, Notes;
- `System views` switcher with grouped required system views;
- current system-view title;
- saved-view selector under the active surface;
- view bar;
- primary grid surface slot;
- inspector regions;
- status strip;
- presence summary slot;
- startup fallback behavior;
- keyboard and focus behavior for the `System views` switcher.

## Implementation Scope

In scope:
- shell registry and stable surface identity model;
- startup-surface request/response adapter and fallback tests;
- shell layout composition across top bar, tabs, switcher, title, view bar, grid,
  inspector, status, and presence slots;
- stable selector/test-id builders for shell, tabs, system views, saved views,
  status strip, inspector, and focus targets;
- FE-P2 unit, browser, visual, and accessibility evidence;
- FE-P2 map metadata fixes and generated ledger regeneration when needed.

Out of scope:
- RDG adapter internals;
- inline edit implementation;
- row mutation replay;
- evidence handle redemption;
- same-field conflict resolver implementation;
- FE-P3 or later grid-adapter behavior;
- FE-P4 or later mutation/sync behavior;
- Core 05 claim-publication work.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [x] | 1. Readiness, map, ledger, FE-P1 handoff | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2`; `make phase-ledger-drift`; `git diff --check` | FE-P1 handoff freshness remains a next-sprint constraint |
| [x] | 2. Startup model and shell registry | `make frontend-unit`; `make frontend-typecheck`; `make generate-drift`; `make phase-ledger-drift`; `git diff --check` | None for `FE-U-P2-01`/`FE-U-P2-02`; saved-view runtime proof is covered by Sprint 5 / `FE-E-P2-01` |
| [x] | 3. Continuous shell composition | `make frontend-unit`; `make browser-e2e-webserver-backed` | None for Sprint 3 audit scope; `FE-B-P2-01` appears eligible for normal row-owned promotion |
| [x] | 4. `System views` keyboard/focus | `make frontend-unit`; `make frontend-typecheck`; `make browser-e2e-webserver-backed`; conditional `make browser-e2e-support`; `git diff --check` | None for Sprint 4 audit scope; `FE-B-P2-02` has retained row-owned browser evidence and is not FE-P2 completion |
| [x] | 5. Saved-view placement and same-shell E2E | `make frontend-unit`; `make frontend-typecheck`; `make backend-store`; `make backend-integration`; `make browser-e2e-webserver-backed`; `make phase-ledger-drift`; `git diff --check` | None for Sprint 5 audit scope; visual/accessibility readiness remains Sprint 6 |
| [ ] | 6. Visual and accessibility readiness | `make browser-e2e-visual`; `make browser-e2e-a11y-preflight` | Missing FE-P2 visual fixture blocks `FE-V-P2-01` |
| [ ] | 7. Closure and FE-P3 handoff | Row-owned commands plus drift and regression checks | `make check` only if repository completion rules require it |

## Global References

Use these as the FE-P2 source set:
- Core 00 through Core 04 under `docs/spec/`.
- Core 05 only for explicit claim-bearing publication boundaries.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`, Phase FE-P2.
- `docs/testing-harness-nlspec.md` for harness mechanics.
- `docs/domain.md`, dev guide, UI/UX guide, `docs/design.md`, visual golden
  guide.
- `tools/frontend_phase_registry.json`, FE-P0/FE-P1/FE-P2 maps, and generated
  frontend ledgers.

## Evidence Layer Matrix

| Row | Evidence class | Intended validation layer | Claim intent |
| --- | --- | --- | --- |
| `FE-U-P2-01` | `product_conformance` | `make frontend-unit` | Core-owned startup selection and fallback behavior only |
| `FE-U-P2-02` | `product_conformance` | `make frontend-unit` | Core-owned built-in/system-view registry by stable IDs |
| `FE-B-P2-01` | `product_conformance` | `make browser-e2e-webserver-backed` | Core-owned continuous shell membership; design details remain separate |
| `FE-B-P2-02` | `product_conformance` | `make browser-e2e-webserver-backed` | Core-owned system-view identity/opening; switcher keyboard/focus remains design-direction evidence |
| `FE-E-P2-01` | `product_conformance` | `make browser-e2e-webserver-backed` | Core-owned saved-view placement and same-shell navigation |
| `FE-V-P2-01` | `design_direction` | `make browser-e2e-visual` | Design-direction visual readiness only; no Core product or Core 05 claim |
| `FE-A11Y-P2-01` | `design_direction` | `make browser-e2e-a11y-preflight` | Design-direction accessibility readiness only; no Core product or Core 05 claim |

## Dependencies and Prerequisites

- FE-P0 and FE-P1 regression status must be rerun or precisely blocked before
  closure.
- Generated outputs under `/internal/gen/**`, `/packages/protocol-ts/src/generated/**`,
  generated ledgers, generated schedules, `pnpm-lock.yaml`, and `go.sum` must
  not be hand-edited.
- Metadata changes to frontend maps require `make phase-ledgers` followed by
  `make phase-ledger-drift`.
- Browser rows require local service-backed browser infrastructure.
- Stable selectors must use `view_schema_id`, `saved_view_id`, `sheet_ref`,
  closed vocabularies, and encoded selector segments, not labels, DOM nesting,
  CSS class names, render order, or component names.

## Public Interfaces and Deliverables

- New root plan: `FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md`.
- Authored selector additions in `/packages/ui-contracts`:
  - shell/top-bar/tab-list/current-title/view-bar/grid-slot/inspector/status-strip/presence
    summary IDs;
  - `System views` trigger/menu/group/option IDs keyed by closed group tokens and
    registry-backed `view_schema_id`;
  - saved-view selector/option IDs keyed by active `view_schema_id` and
    `saved_view_id`.
- App shell changes in `/apps/web`:
  - shell registry;
  - startup adapter;
  - continuous shell slots;
  - custom `System views` switcher;
  - saved-view selector placement.
- Tests:
  - unit coverage for startup and registry rows;
  - browser coverage for shell composition, switcher behavior, and saved-view
    placement;
  - visual and a11y preflight coverage with FE-P2 scenario mapping.
- Phase metadata:
  - corrected FE-P2 map metadata;
  - regenerated FE-P2 ledger through `make phase-ledgers`.

## Sprint 1: Readiness, Map, Ledger, And FE-P1 Handoff

Status: Complete for Sprint 1 readiness and metadata correction. This sprint
does not complete or promote any FE-P2 row.

Objective: make FE-P2 planning metadata trustworthy before behavior work.

Non-goals: no product UI changes, no row promotion, no Core 05 activation.

Source constraints: frontend guide controls row set; maps are source, ledgers
generated.

Inspect: registry, FE-P2 map/ledger, FE-P0/FE-P1 maps/ledgers, FE-P0/FE-P1
plans, visual fixture matrix.

Test-first sequence:
- verify row inventory and duplicate count with `jq`;
- verify FE-B metadata separates `R2-AC-*` design IDs from Core AC IDs;
- verify owner refs match inspected Core sections.

Implementation tasks:
- corrected FE-P2 authored map metadata for owner references and evidence-class
  separation;
- refreshed the generated FE-P2 ledger through `make phase-ledgers`;
- updated the frontend guide FE-P2 row table to match inspected owner sections.

Validation commands:
- `make phase-ledgers`: passed with run root
  `.cartulary/test-results/20260530T173411Z-p3939879`.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2`: passed and reports
  all seven rows as blocked under the planned frontend phase.
- `make phase-ledger-drift`: passed with run root
  `.cartulary/test-results/20260530T173423Z-p3940471`.
- Row inventory `jq` check: passed; exactly seven required FE-P2 rows exist once
  each and all remain blocked.
- Row-specific R2 split `jq` check: passed; FE-B design-direction IDs are no
  longer present as collapsed Core AC fields.
- `git diff --check`: passed with no output.
- `git diff --cached --check`: skipped because no staged changes existed.
- `make agent-finalize`: passed with run root
  `.cartulary/test-results/20260530T174430Z-p3953679`; generated artifacts were
  unchanged and retained-run maintenance was skipped because `RESULTS_DIR` was
  unset.

Deliverables: plan file, corrected authored metadata, regenerated FE-P2 ledger,
and generated ledger consistency.

Blocker rules: FE-P1 handoff freshness and missing FE-P2 row-owned behavior
evidence block FE-P2 row promotion but not later implementation planning.

Binary acceptance: all seven FE-P2 rows appear exactly once; all remain blocked
until row-owned evidence exists; product-conformance, design-direction,
implementation-support, and claim-publication metadata remain separated.

## Sprint 2: Startup Resolution Model And Shell Registry

Objective: implement and test startup-surface resolution and registry identity
without depending on visible labels.

Non-goals: no saved-view CRUD, no grid adapter internals, no mutation replay.

Source constraints: Core 03 Section 2.4 and Core 01 Section 3.3.5.2 own startup
behavior.

Status: complete for `FE-U-P2-01` and `FE-U-P2-02` unit scope as of the Sprint 2
hardening pass. Runtime saved-view placement and `scope='system'` saved-view
non-conflation remain owned by Sprint 5 / `FE-E-P2-01` because the active-surface
saved-view selector is still absent.

Inspect: `WorkbookShell.tsx`, `WorkbookShell.surfaces.test.tsx`,
`packages/ui-contracts`, `packages/view-contracts`.

Completed test-first sequence:
- add unit tests for explicit launch, home preference, incident default, Timeline
  fallback;
- include invalid, missing, invisible, deleted, or unsupported pointers
  clearing/falling through;
- assert selected `sheet_ref` and base `view_schema_id` remain distinct.
- add hardening tests for unsupported explicit/persisted `sheet_ref.kind`,
  deleted saved-view wording as `saved_view_not_found`,
  `required_reference_pack_unavailable`, non-standard backend
  `selected_view_schema_id`, relabeled registry contracts, optional standardized
  surface absence, required surface contract failure, and
  `requiredReferencePackKeys` exposure.

Completed implementation tasks:
- extract startup model and shell registry helpers;
- define required built-in and required system-view registries by
  `view_schema_id`;
- keep optional standardized surfaces additive and after required groups.
- document Core 01/Core 03 startup reason registries and deleted-as-not-found
  semantics;
- expose `requiredReferencePackKeys` from parsed view contracts into workbook
  surface registry entries;
- update the authored FE-P2 map and regenerate the generated FE-P2 ledger.

Validation completed:
- `make generate`: passed,
  `.cartulary/test-results/20260530T184358Z-p4047333`.
- `make phase-ledgers`: passed,
  `.cartulary/test-results/20260530T184407Z-p4047982`.
- `make format`: passed,
  `.cartulary/test-results/20260530T184415Z-p4048374`.
- `make frontend-unit`: passed,
  `.cartulary/test-results/20260530T184437Z-p4050046`.
- `make frontend-typecheck`: passed,
  `.cartulary/test-results/20260530T184454Z-p4051629`.
- `make agent-finalize`: passed unchanged,
  `.cartulary/test-results/20260530T184505Z-p4052081`; retained-run
  maintenance was skipped because `RESULTS_DIR` was unset.
- `make generate-drift`: passed,
  `.cartulary/test-results/20260530T184518Z-p4053267`.
- `make generated-artifact-policy-check`: passed,
  `.cartulary/test-results/20260530T184518Z-p4053275`.
- `make phase-ledger-drift`: passed,
  `.cartulary/test-results/20260530T184518Z-p4053290`.
- `git diff --check`: passed with no output.

Deliverables: `FE-U-P2-01` and `FE-U-P2-02` unit evidence, startup reason-code
spec cleanup, view-contract pack-key parsing, FE-P2 map updates, generated
contract refresh, and regenerated FE-P2 coverage ledger.

Blocker rules: backend route/contract mismatch is `outside FE-P2` unless
frontend adapter behavior is wrong.

Binary acceptance: startup and registry tests pass through `make frontend-unit`;
frontend typecheck passes; generated contract and ledger drift checks pass; no
label/order/CSS selectors are used as behavior anchors.

## Sprint 3: Continuous Workbook Shell Composition

Objective: compose one continuous shell with top bar, tab bar, `System views`,
current title, view bar, primary grid slot, inspector, status strip, and presence
slot.

Non-goals: no inline-edit changes, no conflict resolver implementation, no
evidence redemption.

Source constraints: Core 03 Section 1 through Section 2 own workbook shell
behavior; UI/UX/design docs guide layout only.

Inspect: `WorkbookShell.tsx`, `WorkbookGridControls.tsx`, `App.tsx`, existing
surface tests, Playwright helpers.

Completed test-first sequence:
- add browser scenario for `FE-B-P2-01`;
- assert all shell regions remain within one shell root and same URL/workbook
  route;
- assert built-in tabs are ordered by registry IDs.

Completed implementation tasks:
- introduce shell slot component in `/apps/web`;
- hoist or bridge current view bar/status/presence/inspector outputs into shell
  slots;
- keep `IncidentAdminPanel` from becoming FE-P2 shell evidence unless moved into
  an explicit non-primary/support region.

Audit inspection completed:
- `WorkbookShell.tsx`, `WorkbookGridControls.tsx`, `App.tsx`;
- `WorkbookShell.surfaces.test.tsx`, `workbookSurfaceRegistry.test.ts`, and
  shared selector-contract coverage;
- Playwright helpers and `apps/web/e2e/phase2.spec.ts`;
- `tools/frontend_phase_maps/fe_p2_test_map.json` and generated FE-P2 ledger
  entries relevant to `FE-B-P2-01`.

Assertions verified:
- one `workbook-shell-ready` root is present;
- every required `workbookShellSlots` entry is present exactly once under that
  root and carries the same `data-workbook-shell-id`;
- `System views` is inside the `system-views` slot;
- the workbook remains on `/` with the same `incident_id`;
- built-in tab identity and order come from `requiredBuiltInWorkbookSurfaceIds`
  and shared selector builders;
- `IncidentAdminPanel` is rendered in an explicit support region and is not
  counted as primary-grid shell evidence.

Validation completed:
- `make frontend-unit`: passed,
  `.cartulary/test-results/20260530T200312Z-p13189/frontend-unit/tool-run-summary.json`.
- `make frontend-typecheck`: passed,
  `.cartulary/test-results/20260530T200328Z-p14652/frontend-typecheck/tool-run-summary.json`.
- `make browser-e2e-webserver-backed`: passed, 62 tests, 0 failures,
  `.cartulary/test-results/20260530T200337Z-p15057/browser-e2e-webserver-backed/tool-run-summary.json`.
- Phase 2 authoritative browser evidence selected and passed `E-2-01`,
  `.cartulary/test-results/20260530T200337Z-p15057/browser-e2e-webserver-backed/browser-e2e-functional-phase2-authoritative/phase-summary.json`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260530T200337Z-p15057`:
  failed at retained-run preflight because `agent-finalize` requires a
  successful full warm `make check` retained run root; no files were updated.

Deliverables: browser scenario and app shell composition changes.

Blocker rules: missing slot or fragmented route/shell blocks `FE-B-P2-01`.
No Sprint 3 blocker remains from the audit evidence.

Binary acceptance: satisfied for Sprint 3. `FE-B-P2-01` scenario is present,
mapped, and passes through `make browser-e2e-webserver-backed`. This plan update
does not itself promote the row or replace the normal phase-ledger workflow.

## Sprint 4: `System views` Switcher Keyboard And Focus Behavior

Objective: replace/native-wrap current system-view selection with a
keyboard-testable switcher supporting entry, roving focus, selection, dismissal,
and focus restoration.

Non-goals: no command palette, no new route family, no optional surface behavior
beyond additive ordering.

Source constraints: Core-owned surface identity remains `view_schema_id`;
UI/UX/design own menu interaction direction.

Inspect: current `systemViewSelectorTestId()` consumers, phase9 sentinel tests,
a11y helpers.

Test-first sequence:
- unit-test switcher group/order and selector builders;
- browser-test Tab to trigger, Enter/Space open, Arrow navigation, Enter select,
  Esc dismiss, focus restoration.

Implementation tasks:
- add stable trigger/menu/group/option selectors;
- migrate app/e2e consumers away from native `selectOption`;
- preserve accessible name `System views`;
- restore focus to trigger on dismissal and move to first grid focus target on
  selection.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-support` if shared helpers change

Evidence:
- `make frontend-unit`: passed,
  `.cartulary/test-results/20260530T215324Z-p229213/frontend-unit/tool-run-summary.json`.
- `make frontend-typecheck`: passed,
  `.cartulary/test-results/20260530T215339Z-p230690/frontend-typecheck/tool-run-summary.json`.
- `make browser-e2e-webserver-backed`: passed, 62 tests, 0 failures,
  `.cartulary/test-results/20260530T215349Z-p231033/browser-e2e-webserver-backed/tool-run-summary.json`.
- `FE-B-P2-02` is closed in frontend row accounting and the underlying
  Playwright shard contains the passed browser scenario:
  `.cartulary/test-results/20260530T215349Z-p231033/browser-e2e-webserver-backed/frontend-row-accounting.json`
  and
  `.cartulary/test-results/20260530T215349Z-p231033/browser-e2e-webserver-backed/playwright-webserver-backed-batch/runner.json`.
- `make browser-e2e-support`: passed because shared browser helper and selector
  helper usage was in the Sprint 4 audit surface,
  `.cartulary/test-results/20260530T215520Z-p239038/browser-e2e-support/tool-run-summary.json`.
- `git diff --check`: passed.
- `git diff --cached --check`: skipped because no staged changes existed.

Audit result:
- scope was `FE-B-P2-02` only;
- Core 00 through Core 04 were treated as product-conformance authority;
- `docs/testing-harness-nlspec.md` was treated as harness-mechanics authority;
- the FE-P2 map and generated FE-P2 coverage ledger were treated only as
  generated/support evidence;
- the switcher is keyboard reachable from the shell, Enter/Space open it,
  Arrow navigation uses deterministic roving focus, Enter selects the focused
  system view, Escape dismisses it, dismissal restores focus to the trigger,
  and selection moves focus to the first appropriate grid focus target;
- the accessible name remains `System views`;
- system-view identity and selector contracts remain keyed by stable
  `view_schema_id` values and closed group tokens rather than labels, CSS
  classes, DOM nesting, render order, or component names;
- existing later-phase system-view tests continue to use stable selector
  builders and stable-ID assertions;
- no command palette, new route family, optional surface semantics, or other
  behavior outside Sprint 4 scope was found.

Deliverables: satisfied for `FE-B-P2-02`.

Blocker rules: browser focus instability is `FE-P2-owned` unless caused by
harness/service failure.

Binary acceptance: satisfied for Sprint 4 / `FE-B-P2-02` only. This does not
mark FE-P2 complete, does not promote Sprint 5, Sprint 6, visual,
accessibility, or Core 05 evidence, and does not replace generated-ledger drift
workflow.

## Sprint 5: Saved-View Placement And Same-Shell E2E

Status: complete for `FE-E-P2-01` after the saved-view placement remediation.
This sprint does not complete visual or accessibility readiness and does not
promote FE-P8 query/layout persistence.

Objective: show saved views only under the active surface's view selector and
keep system/saved views inside the same workbook shell.

Non-goals: no saved-view create/update/delete UX, no FE-P8 query/layout
persistence behavior, and no production API path for creating `scope="system"`
saved views.

Source constraints: Core 03 saved-view model and Core 01 saved-view/workbook-startup
routes own product behavior. `docs/testing-harness-nlspec.md` owns only the
harness mechanics for seeding `scope="system"` saved-view fixtures.

Inspect: saved-view route helpers in phase8 e2e,
`WorkbookShell.surfaces.test.tsx`, app URL update logic, saved-view backend
routes/store code, harness test-route guards, and FE-P2/Phase 8 maps.

Completed test-first sequence:
- browser-test visible saved views grouped by active `view_schema_id`;
- assert saved views do not become primary tabs;
- assert selecting contract-backed system views does not leave the shell and
  keeps `sheet_ref.kind="view_schema"`;
- assert visible `scope="system"` saved views remain distinct saved-view
  objects, appear only in the active saved-view selector, and select with
  `sheet_ref.kind="saved_view"` plus `saved_view_id`;
- assert selecting a saved view issues a query through the saved view's base
  `view_schema_id`.

Completed implementation tasks:
- fetch visible saved views through existing public route;
- render active-surface saved-view selector in the view bar;
- keep selected saved-view identity as `sheet_ref.kind="saved_view"` and
  `saved_view_id`, while querying/rendering by base `view_schema_id`;
- add a guarded harness-only `POST
  /api/v1/test/incidents/{incident_id}/saved-views/system` route for browser
  evidence without changing ordinary saved-view create semantics;
- add backend route tests for default unavailability, test-token enforcement,
  and valid system saved-view fixture creation;
- update the authored FE-P2 and Phase 8 maps, regenerate generated ledgers, and
  refresh Go duration baselines for the added backend tests.

Validation completed:
- `make generate`: passed,
  `.cartulary/test-results/20260530T235458Z-p462859`.
- `make phase-ledgers`: passed,
  `.cartulary/test-results/20260530T233751Z-p421140`.
- `make frontend-unit`: passed,
  `.cartulary/test-results/20260530T234314Z-p432338`.
- `make frontend-typecheck`: passed,
  `.cartulary/test-results/20260530T234332Z-p433870`.
- `make backend-store`: passed,
  `.cartulary/test-results/20260530T233755Z-p421377`.
- `make backend-integration`: passed,
  `.cartulary/test-results/20260530T234350Z-p434353`.
- `make browser-e2e-webserver-backed`: passed,
  `.cartulary/test-results/20260530T234930Z-p447422`.
- `make phase-ledger-drift`: passed,
  `.cartulary/test-results/20260530T235514Z-p463663`.
- `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260530T233755Z-p421377`:
  passed, `.cartulary/test-results/20260530T235310Z-p459563`.
- `make go-test-duration-baseline-coverage`: passed,
  `.cartulary/test-results/20260530T235514Z-p463660`.
- `make phase-schedule-drift`: passed,
  `.cartulary/test-results/20260530T235420Z-p461803`.
- `make agent-finalize`: passed,
  `.cartulary/test-results/20260530T235315Z-p459751`; generated artifacts were
  unchanged in the final run and retained-run maintenance was skipped because
  `RESULTS_DIR` was unset.
- `git diff --check`: passed with no output.

Local audit update on 2026-05-31:
- Scope: re-audited `FE-E-P2-01` only. This audit did not assess FE-P2 closure,
  Sprint 6 visual/accessibility readiness, FE-P8 saved-view query/layout
  persistence, Core 05 publication evidence, or saved-view CRUD validation.
- Result: `PASS: Sprint 5 / FE-E-P2-01 remains supported by row-owned evidence.`
- Evidence confirmed: visible saved views remain under only the active surface's
  saved-view selector and are absent from primary workbook tabs; contract-backed
  system views remain in the same workbook shell as `sheet_ref.kind="view_schema"`;
  saved views, including harness-seeded `scope="system"` saved views, remain
  distinct saved-view objects selected as `sheet_ref.kind="saved_view"` with the
  selected `saved_view_id`; saved-view selection still queries and renders through
  the saved view's base `view_schema_id`; ordinary public saved-view create still
  rejects `scope="system"`; and the harness fixture route remains guarded,
  unavailable by default, and outside product API behavior.
- Validation rerun: `make agent-finalize` passed
  `.cartulary/test-results/20260531T001105Z-p488056`; `make frontend-unit`
  passed `.cartulary/test-results/20260531T001113Z-p489022`;
  `make frontend-typecheck` passed
  `.cartulary/test-results/20260531T001131Z-p490550`; `make backend-store`
  passed `.cartulary/test-results/20260531T001141Z-p490938`;
  `make backend-integration` passed
  `.cartulary/test-results/20260531T001707Z-p501562`;
  `make browser-e2e-webserver-backed` passed
  `.cartulary/test-results/20260531T002250Z-p514600`;
  `make phase-ledger-drift` passed
  `.cartulary/test-results/20260531T002420Z-p523377`; and
  `git diff --check` passed with no output. `make phase-ledgers` was not run
  because no FE-P2 metadata changed during the audit.

Deliverables: satisfied for `FE-E-P2-01`: mapped browser scenario, app reload
behavior, harness-only system saved-view fixture route, backend fixture tests,
support-doc updates, regenerated ledgers, schedule drift validation, and duration
baseline coverage.

Blocker rules: none remain for Sprint 5 audit scope. Future saved-view
query/layout persistence remains FE-P8-owned; visual and accessibility readiness
remain Sprint 6-owned.

Binary acceptance: satisfied. Saved-view selector is under active surface only,
saved views never render as primary tabs, contract-backed system views stay
same-shell with `sheet_ref.kind="view_schema"`, saved views including
`scope="system"` select by `sheet_ref.kind="saved_view"` and `saved_view_id`,
selected saved views query/render through the base `view_schema_id`, ordinary
public create still rejects `scope="system"`, and the row passes through
browser-backed E2E.

## Sprint 6: Visual And Accessibility Readiness

Local execution plan, 2026-05-31:
- confirm `FE-V-P2-01` target selection before touching visual fixtures;
- add exact `FE-A11Y-P2-01` preflight coverage first, then make only minimal
  shell accessibility semantics needed by that preflight;
- keep both Sprint 6 rows `design_direction` and leave generated ledgers
  generated-only.

Objective: collect design-direction visual and accessibility readiness without
turning either into product-conformance or Core 05 evidence.

Non-goals: no Core 05 publication, no benchmark/fixture-sensitive public claim.

Source constraints: visual golden guide owns refresh mechanics; UI/UX/design own
direction.

Inspect: `workbook.visual.spec.ts`, visual snapshots,
`workbook.a11y-preflight.spec.ts`, `a11yPhaseMap.ts`, FE-P2 map.

Test-first sequence:
- add/adjust FE-P2 visual scenario title exactly as mapped;
- add/adjust FE-P2 a11y preflight scenario for shell regions, tabs,
  switcher/menu, inspector controls, status strip, visible focus, names.

Implementation tasks:
- create or correct FE-P2 default shell visual fixture/golden;
- mask dynamic incident/user/presence data;
- ensure preflight summary keeps FE-P2 blocked-row evidence separate from
  implemented a11y rows unless promotion rules are met.

Validation commands:
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight`
- `make frontend-typecheck`
- `make phase-ledger-drift`

Deliverables: `FE-V-P2-01` and `FE-A11Y-P2-01` evidence or exact blockers.

Blocker rules: visual fixture mismatch blocks `FE-V-P2-01`; a11y preflight
target failure blocks `FE-A11Y-P2-01`.

Binary acceptance: both rows remain `design_direction`; no Core 05 review runs.

Execution result, 2026-05-31:
- Sprint 6 execution scope is complete with an exact visual blocker and passing
  blocked-row accessibility preflight evidence; FE-P2 is not claimed complete.
- `FE-V-P2-01` remains `claim_status="blocked"` and
  `evidence_class="design_direction"`. Exact blocker:
  `support-tooling-owned`; direct `make browser-e2e-visual` selection is still
  base-manifest driven and selects only `phase3`, `phase4`, `phase5`, and
  `phase6` `browser_visual` rows. The command's retained
  `frontend-row-accounting.json` lists the FE-P2 visual row as missing because
  no exact FE-P2 visual Playwright title is selected. Minimum next action:
  owner-approved support-tooling work to let `browser-e2e-visual` select
  frontend visual rows, or owner-approved remapping to an existing visual
  fixture without product-conformance or Core 05 promotion.
- `FE-A11Y-P2-01` remains `claim_status="blocked"` and
  `evidence_class="design_direction"`. Direct preflight evidence exists from
  `make browser-e2e-a11y-preflight`, but it remains
  `cartulary.frontend_accessibility_preflight_summary.v1` blocked-row smoke
  evidence and is not implemented-row evidence. Minimum next action for
  promotion: owner-approved move to the implemented accessibility evidence path.
- Changed files: `FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md`,
  `apps/web/e2e/workbook.a11y-preflight.spec.ts`, and
  `apps/web/src/WorkbookShell.tsx`.
- App/test changes: the preflight spec now splits out the exact
  `FE-A11Y-P2-01` title, keeps generic blocked-row smoke separate, and asserts
  shell slot names, built-in tab state, System views menu behavior, saved-view
  selector, grid controls, inspector controls, status strip semantics, visible
  focus, accessible names, and a simple no-focus-trap condition. `WorkbookShell`
  adds named shell/slot regions, active tab `aria-current`, a named Timeline
  inspector, and save-state `role="status"` semantics.
- Visual files and snapshots were not edited. Generated frontend ledgers were
  not hand-edited.

Validation and artifacts:
- Visual target-selection inspection passed:
  `tmp/node-runtime/bin/node scripts/lib/phase-manifest.mjs playwright-phases authoritative browser_visual`
  returned only `phase3`, `phase4`, `phase5`, and `phase6`; the matching
  `playwright-grep` checks returned only `V-3-*`, `V-4-*`, `V-5-*`, and
  `V-6-*` titles.
- Test-first `make browser-e2e-a11y-preflight` failed as `FE-P2-owned` before
  shell semantics were added:
  `.cartulary/test-results/20260531T004904Z-p556113/browser-e2e-a11y-preflight/tool-run-summary.json`.
- Final `make browser-e2e-a11y-preflight` passed:
  `.cartulary/test-results/20260531T005016Z-p560728/browser-e2e-a11y-preflight/tool-run-summary.json`;
  normalized preflight summary:
  `.cartulary/test-results/20260531T005016Z-p560728/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json`;
  runner:
  `.cartulary/test-results/20260531T005016Z-p560728/browser-e2e-a11y-preflight/browser-e2e-a11y-preflight-accessibility-preflight/runner.json`.
- `make browser-e2e-visual` failed on existing selected `V-3-*` through `V-6-*`
  screenshot mismatches, outside FE-P2:
  `.cartulary/test-results/20260531T005111Z-p564852/browser-e2e-visual/tool-run-summary.json`;
  row accounting:
  `.cartulary/test-results/20260531T005111Z-p564852/browser-e2e-visual/frontend-row-accounting.json`;
  example retained diff:
  `.cartulary/test-results/20260531T005111Z-p564852/browser-e2e-visual/browser-e2e-visual-phase3-authoritative/playwright-output/workbook.visual-Phase-3-wo-bbd1a-ersion-and-save-state-strip/v-3-grid-01-timeline-default-diff.png`.
- `make frontend-typecheck` passed:
  `.cartulary/test-results/20260531T005254Z-p571141/frontend-typecheck/tool-run-summary.json`.
- Narrow prerequisite `make frontend-unit` passed:
  `.cartulary/test-results/20260531T005345Z-p572841/frontend-unit/tool-run-summary.json`.
- `make phase-ledger-drift` passed:
  `.cartulary/test-results/20260531T005307Z-p571890/phase-ledger-drift/tool-run-summary.json`.
- `git diff --check` passed after recording this execution result.

## Sprint 7: Closure, Drift, Regression, And FE-P3 Handoff

Objective: close FE-P2 only with row-owned evidence or exact blockers, then
prepare FE-P3 handoff.

Non-goals: no broad completion claim from ledgers, retained artifacts,
visual/a11y evidence, or support-only checks.

Source constraints: phase maps are source of truth; ledgers generated.

Inspect: plan file, FE-P2 map/ledger, FE-P0/FE-P1 handoff, app shell selectors,
visual/a11y artifacts.

Test-first sequence:
- rerun all FE-P2 row-owned commands;
- rerun FE-P0/FE-P1 regression commands or record blockers;
- verify generated artifacts and import boundaries.

Validation commands:
- `make frontend-typecheck`
- `make frontend-unit`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight`
- `make browser-e2e-support`
- `make generated-artifact-policy-check`
- `make generate-drift`
- `make frontend-import-boundary-check`
- `make phase-ledgers` when metadata changed
- `make phase-ledger-drift`
- `git diff --check`
- `git diff --cached --check` when staged changes exist
- `make check` only if repository completion rules require broad closure

Deliverables: final FE-P2 evidence table, blockers, artifact paths, regression
status, FE-P3 handoff.

Blocker rules: unresolved row evidence keeps that row blocked and prevents FE-P2
completion.

Binary acceptance: every FE-P2 row has direct row-owned evidence or an exact
blocker, and all evidence classes remain separated.

## Blocker Recording Rules

Every blocker must record:
- exact missing artifact or failed command;
- exact failing target, scheduler unit, or test when available;
- result root, run ID, run root, summary JSON, stdout log, and stderr log paths
  when available;
- affected FE-P2 row ID;
- failure class and failure reason when exposed;
- ownership classification: `FE-P2-owned`, `FE-P0-regression-owned`,
  `FE-P1-regression-owned`, `support-tooling-owned`, `harness-owned`,
  `infra-owned`, or `outside FE-P2`;
- minimum follow-up needed;
- whether the blocker prevents FE-P2 completion;
- whether the blocker affects FE-P3 handoff.

## Binary Exit Criteria

FE-P2 is complete only when all are true:

- `FRONTEND_PHASE2_IMPLEMENTATION_PLAN.md` exists and states it is not behavior
  authority.
- FE-P2 registry, map, and ledger are present or exact blockers are recorded.
- Every FE-P2 row is represented exactly once in the phase map.
- FE-P2 map metadata preserves product-conformance, design-direction,
  implementation-support, and claim-publication separation.
- Generated ledgers are generated, not hand-edited.
- Each FE-P2 row has direct row-owned evidence or an exact blocker.
- FE-P0 and FE-P1 regression status is recorded.
- Visual and accessibility evidence do not activate Core 05 by default.
- Selector/test-id contracts use stable IDs, `view_schema_id`, `saved_view_id`,
  `sheet_ref`, closed vocabularies, and encoded stable segments.
- No FE-P2 behavior depends on labels, translated text, render order, CSS
  classes, DOM nesting, or incidental component names.
- Whitespace checks pass.
- Generated-artifact policy and generate-drift checks pass or are precisely
  blocked.
- Unresolved owner-lookup TODOs are recorded and excluded from completion.
- `make check` is run only if repository completion rules require it; if skipped,
  the reason is recorded.

## Handoff Requirements For FE-P3

FE-P3 must receive:

- shell composition status;
- startup-surface resolution status;
- built-in tab and system-view registry status;
- saved-view placement status;
- shell focus and keyboard behavior status;
- selector/test-id status;
- visual fixture status;
- accessibility status;
- command-surface blockers and artifact paths;
- phase-map and ledger drift status;
- generated-artifact policy status;
- FE-P0 and FE-P1 regression status;
- unresolved owner-lookup TODOs.

Generated artifacts intentionally not edited by hand:
- generated frontend ledgers;
- generated schedules/task-surface files;
- generated protocol files;
- lockfiles and Go checksum files.
