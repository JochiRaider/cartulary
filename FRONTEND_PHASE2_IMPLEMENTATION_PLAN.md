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
- All FE-P2 rows are currently `claim_status="blocked"`.
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
- `/packages/ui` exists only as `packages/ui/.gitkeep`; it is inactive until a
  manifest-backed package is introduced.
- `WorkbookShell` currently has built-in tab selectors, a native `select` for
  system views, backend `/workbook-startup` consumption, and selected saved-view
  `sheet_ref` preservation.
- Current shell composition is not yet FE-P2-complete: startup and tabs exist,
  but the shell still has hero/card-shaped sections, a native select instead of
  roving-focus switcher behavior, no dedicated saved-view selector under the
  active surface, and status/presence/inspector regions are partly
  timeline-local.

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

Missing evidence:
- No direct FE-P2 row-owned evidence has been collected.
- No FE-P2-specific visual golden named for a phase-2 shell capture was found;
  visual goldens currently start with `v-3-*` files even though `FE-VFIX-01` is
  marked current.

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
- FE-P2 phase map/ledger exist but all FE-P2 rows remain blocked.
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
| [ ] | 1. Readiness, map, ledger, FE-P1 handoff | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P2`; `make phase-ledger-drift`; `git diff --check` | FE-P2 metadata separation defects must be fixed or blocked |
| [ ] | 2. Startup model and shell registry | `make frontend-unit`; `make frontend-typecheck` | Backend `/workbook-startup` contract mismatch blocks `FE-U-P2-01` |
| [ ] | 3. Continuous shell composition | `make frontend-unit`; `make browser-e2e-webserver-backed` | Missing shell slots block `FE-B-P2-01` |
| [ ] | 4. `System views` keyboard/focus | `make frontend-unit`; `make browser-e2e-webserver-backed` | Native select is insufficient for roving-focus evidence |
| [ ] | 5. Saved-view placement and same-shell E2E | `make browser-e2e-webserver-backed` | Missing saved-view selector blocks `FE-E-P2-01` |
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

Inspect: `WorkbookShell.tsx`, `WorkbookShell.surfaces.test.tsx`,
`packages/ui-contracts`, `packages/view-contracts`.

Test-first sequence:
- add unit tests for explicit launch, home preference, incident default, Timeline
  fallback;
- include invalid, missing, invisible, deleted, or unsupported pointers
  clearing/falling through;
- assert selected `sheet_ref` and base `view_schema_id` remain distinct.

Implementation tasks:
- extract startup model and shell registry helpers;
- define required built-in and required system-view registries by
  `view_schema_id`;
- keep optional standardized surfaces additive and after required groups.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make phase-ledger-drift` if metadata changed

Deliverables: `FE-U-P2-01` and `FE-U-P2-02` unit evidence or blockers.

Blocker rules: backend route/contract mismatch is `outside FE-P2` unless
frontend adapter behavior is wrong.

Binary acceptance: startup and registry tests pass through `make frontend-unit`;
no label/order/CSS selectors are used as behavior anchors.

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

Test-first sequence:
- add browser scenario for `FE-B-P2-01`;
- assert all shell regions remain within one shell root and same URL/workbook
  route;
- assert built-in tabs are ordered by registry IDs.

Implementation tasks:
- introduce shell slot component in `/apps/web`;
- hoist or bridge current view bar/status/presence/inspector outputs into shell
  slots;
- keep `IncidentAdminPanel` from becoming FE-P2 shell evidence unless moved into
  an explicit non-primary/support region.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed`

Deliverables: browser scenario and app shell composition changes.

Blocker rules: missing slot or fragmented route/shell blocks `FE-B-P2-01`.

Binary acceptance: `FE-B-P2-01` scenario is present, mapped, and passes only
through `make browser-e2e-webserver-backed`.

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

Deliverables: `FE-B-P2-02` row-owned browser evidence or blocker.

Blocker rules: browser focus instability is `FE-P2-owned` unless caused by
harness/service failure.

Binary acceptance: switcher scenario passes and existing later-phase system-view
tests are updated without weakening stable-ID assertions.

## Sprint 5: Saved-View Placement And Same-Shell E2E

Objective: show saved views only under the active surface's view selector and
keep system/saved views inside the same workbook shell.

Non-goals: no saved-view create/update/delete UX, no FE-P8 query/layout
persistence behavior.

Source constraints: Core 03 saved-view model and Core 01 saved-view/workbook-startup
routes own behavior.

Inspect: saved-view route helpers in phase8 e2e,
`WorkbookShell.surfaces.test.tsx`, app URL update logic.

Test-first sequence:
- browser-test visible saved views grouped by active `view_schema_id`;
- assert saved views do not become primary tabs;
- assert selecting system views does not leave the shell.

Implementation tasks:
- fetch visible saved views through existing public route;
- render active-surface saved-view selector in the view bar;
- keep selected saved-view identity as `sheet_ref.kind="saved_view"` and
  `saved_view_id`, while querying/rendering by base `view_schema_id`.

Validation commands:
- `make frontend-unit`
- `make frontend-typecheck`
- `make browser-e2e-webserver-backed`

Deliverables: `FE-E-P2-01` mapped scenario and app code.

Blocker rules: missing saved-view API route behavior is `outside FE-P2`; missing
placement/rendering is `FE-P2-owned`.

Binary acceptance: saved-view selector is under active surface only, keyed by
`saved_view_id`, and row passes through browser-backed E2E.

## Sprint 6: Visual And Accessibility Readiness

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
