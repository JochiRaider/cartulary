# FE-P10: Remaining Workbook Surfaces And Keyboard Completion

## Summary

FE-P10 covers the remaining workbook coordination and review surfaces plus
completion of the workbook keyboard, clipboard, fill-down, frozen-column,
virtual-scroll, group-row, focus-restoration, and deterministic `Esc`
contracts. This file is an implementation plan and handoff artifact only. It
does not implement FE-P10 behavior, close FE-P10 rows, update generated
ledgers, or edit generated protocol artifacts.

The live FE-P10 contract is `docs/guides/cartulary_frontend_implementation_testing_guide.md`
Section 4.10 plus `tools/frontend_phase_maps/fe_p10_test_map.json`.
Current inspected FE-P10 registry posture is `status=planned`,
`row_rollup_state=no_rows_implemented`, activation blocker
`FE-P10-ACTIVATION-BLOCKER-01`, map digest
`76249689c3bce1183cd7297f4f9b10e360eb73f17a9dd1f0d88f3cda8df8735c`,
and generated ledger digest
`de0ee90808402ff463b5dbdcbc34f0f04fb2e3d2230f215886956c3a20d5ecb5`.

All six FE-P10 rows are currently blocked. Product rows can close only from
current mapped row-owned evidence under the current registry, map, guide, and
ledger digests. Visual and accessibility rows remain design-direction evidence
unless a later owner map explicitly changes that evidence class.

## Authority Model

Core 00 through Core 04 own product behavior. FE-P10 product behavior must be
read from the normative core first, with the frontend guide and phase map used
as derived implementation and verification contracts.

Core 05 is inactive for FE-P10 unless explicit claim-bearing timed, benchmark,
fixture-sensitive, or publication predicates exist. No inspected FE-P10 row has
claim-publication intent.

`docs/testing-harness-nlspec.md` owns harness mechanics only: command
invocation, target selection, scheduling, fixture lifecycle, artifact emission,
cleanup, verification gates, and frontend row-accounting mechanics. It does not
own FE-P10 product behavior.

Design, visual, accessibility, fixture, guide, and retained-artifact evidence
is design-direction or implementation-support evidence unless the live row map
explicitly says otherwise. Visual screenshots and accessibility summaries must
not be promoted to product-conformance evidence.

`docs/domain.md` is the vocabulary and concept-boundary reference. It supports
terms such as `party`, `task_request`, `decision`, `view_schema_id`,
`field_key`, and `record_id`, but it is not row closure evidence.

`docs/opentelemetry-instrumentation-nlspec.md` is telemetry-subsystem
authority only. It does not change FE-P10 workbook behavior, FE-P10 row
ownership, or frontend closure rules.

## Sprint 1 Repo Baseline

Before this file was authored, `FRONTEND_PHASE10_IMPLEMENTATION_PLAN.md` was
absent and `git status --short` produced no output.

Prior frontend phase plans were inspected as style and handoff inputs only:

| Plan | SHA-256 |
| --- | --- |
| `FRONTEND_PHASE5_IMPLEMENTATION_PLAN.md` | `1dc922aab59ba22d41636cfe5933a698d932fa19126250cedf885bd45a5f2b2a` |
| `FRONTEND_PHASE6_IMPLEMENTATION_PLAN.md` | `c1dad81966e36373c126dca731da5553feba789674386a8e14d3c6bff6c5c8ee` |
| `FRONTEND_PHASE7_IMPLEMENTATION_PLAN.md` | `83b8786f0ea37129f65d7f7153ec153202ef42acec46f1bd6aa3bd5fbf1e5bc3` |
| `FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` | `8bf970f55443ba7a9fe2fb2e7c9c364df80f9d9d6f380eb6e79f0cee3e4a6d9f` |
| `FRONTEND_PHASE9_IMPLEMENTATION_PLAN.md` | `4d2b997a954aecb6626d3e237c8ff453152caf384517bd48867379ff7d4e4a7b` |

Primary FE-P10 planning sources were inspected:

| Source | SHA-256 or inspected status |
| --- | --- |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_phase_registry.json` | `f14cad6f5de41796a120aeb51a71626e8bef7a21d04e140204def07591e49d53` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `76249689c3bce1183cd7297f4f9b10e360eb73f17a9dd1f0d88f3cda8df8735c` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `de0ee90808402ff463b5dbdcbc34f0f04fb2e3d2230f215886956c3a20d5ecb5` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Normative and support sources inspected for authority boundaries:

| Source | SHA-256 |
| --- | --- |
| Core 00 `docs/spec/00_document_set_status_and_precedence.md` | `e3b2e5e9ed4f47d29694612d571f3255437a9a1acbceb31fc38d9229756a682f` |
| Core 01 `docs/spec/01_architecture_storage_and_view_contracts.md` | `1c55b261681c59e948356d8f80e2d3f5ab8936d33db5742d18a31f701a81bac9` |
| Core 02 `docs/spec/02_domain_model_schema_and_history.md` | `bb92665e26804b8c465d961fdef39b78e3f07c389a26a6478e9a210ce393d3fa` |
| Core 03 `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | `fb561f66e61cf75e777a8c1c4d618d1064ca3b36e8d02435e85131ce631f5b10` |
| Core 04 `docs/spec/04_security_deployment_and_conformance.md` | `ab4d03966850879625141165d7902f108cfe989914e2b01ed42e2ff7968f6da1` |
| Core 05 `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` | `ee2f572430b75b41ccd20d4dede9c72251b3a4432db2ccf525bec9415da7ef89` |
| `docs/domain.md` | `c461f865e5a0524865e691661f9049ceb9031226124aec60e4014b00847d0e21` |
| `docs/testing-harness-nlspec.md` | `f8857f2d67316ba43ac9c7da71040b26fb0f66250c991508b6570d3cf367af83` |
| `docs/opentelemetry-instrumentation-nlspec.md` | `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451` |
| `docs/guides/cartulary-dev-guide.md` | `a4b8fb4b9e3b03c905ed276d19a692559ddf1e70396f224f8f8b2a3f68e58776` |
| `docs/guides/cartulary-ui-ux-design-guide.md` | `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26` |
| `docs/guides/cartulary_visual_golden_maintenance.md` | `b9a372d62fb890e72de140e29da2319fd8e202083e96819675e1edf1f64e07c3` |
| `docs/guides/cartulary_implementation_testing_guide.md` | `050ac1da2a1fa9139a2b721a590405e510c82868bb8b9f358709ffde8a8caf8c` |
| `docs/design.md` | `e28345fac8ba22fc58264454237af209360a84af0c714ff4e1c94c6028d8cd05` |

Relevant implementation surfaces inspected for current substrate were
`/apps/web`, `/packages/grid-adapter`, `/packages/view-contracts`,
`/packages/protocol-ts`, `/packages/ui-contracts`, and `/packages/test-utils`.
Existing code already exposes several coordination-surface and grid-helper
building blocks, but those existing facts do not close FE-P10 rows until the
live FE-P10 map is promoted with direct row-owned evidence.

## Source Limits

Plan text, generated ledgers, fixture registry `current` status, screenshots,
retained old artifacts, historical handoffs, target explanations, and broad
`make check` do not close rows.

Product rows close only from current mapped row-owned evidence under the
current registry, map, guide, and ledger digests. A target pass without
frontend row-accounting v3 for the FE-P10 row is not closure evidence.

Generated ledgers and generated protocol artifacts must not be hand-edited.
If owner inputs change, update authored inputs first and refresh generated
outputs only through Make-owned generator or drift targets.

If live sources conflict, record a blocker and stop promotion. Do not silently
reconcile guide, registry, map, ledger, fixture registry, harness, or core
authority conflicts.

## FE-P9 Handoff Inputs

`FRONTEND_PHASE9_IMPLEMENTATION_PLAN.md` is dependency and style context only.
FE-P9 may show that earlier frontend dependencies were green at the time of
handoff, but FE-P10 must produce its own owner-map-aligned row accounting.

The FE-P9 handoff records FE-P9 as implemented and closed for its mapped rows,
with finalization passing after a retained full-check run. FE-P9 row evidence,
generated ledgers, screenshots, broad health artifacts, and finalization
artifacts must not be imported as FE-P10 closure evidence.

FE-P9 strict non-claims continue into FE-P10 planning until FE-P10 produces
direct evidence: no FE-P10 implementation claim, no Core 05 publication, no
visual/product promotion, no accessibility/product promotion, no benchmark
claim, no fixture-sensitive publication claim, and no visual-publication
claim.

## Phase Objective

FE-P10 must complete remaining workbook surfaces and keyboard completion for
Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review,
Lesson, the full keyboard contract, clipboard behavior, fill-down, frozen
columns, virtual scroll, group rows, focus restoration, and deterministic
`Esc` behavior.

All interactions that can address a writable row or cell must preserve
`record_id + field_key` identity. FE-P10 must not rely on row index, DOM order,
visible label, grouped presentation position, vendor coordinate, or screenshot
state as mutation identity.

## Implementation Scope

| Surface | FE-P10 scope |
| --- | --- |
| `/apps/web` | Workbook shell routing and switcher behavior, coordination surfaces, generic surface behavior, focus, keyboard, clipboard, browser specs, visual specs, and accessibility specs. |
| `/packages/grid-adapter` | Stable row/cell anchors, group-row presentation, frozen-column behavior, virtual scroll behavior, paste/fill targeting, focus restoration hooks, and cleanup. |
| `/packages/view-contracts` | Field metadata, closed vocabulary interpretation, grouping/filter/sort capability interpretation, and stable `field_key` lookup. |
| `/packages/protocol-ts` | Facade consumption of generated protocol contracts only. Generated protocol roots are not hand-edited. |
| `/packages/ui-contracts` | Stable selector and test-id builders for surfaces, grid cells, group rows, keyboard controls, fixture states, and accessibility anchors. Generated roots are not hand-edited. |
| `/packages/test-utils` | Browser helpers for paste, fill-down, group rows, virtual scroll, stable anchors, visual setup, and accessibility setup. |

## Out of Scope

FE-P10 must not include:

1. New route semantics.
2. New security policy.
3. Package ownership changes.
4. Core 05 publication evidence.
5. Benchmark claims.
6. Claim publication.
7. Extension-profile behavior.
8. FE-P11 implementation work.
9. Product-conformance claims from visual evidence.
10. Product-conformance claims from accessibility evidence.
11. Direct `react-data-grid` imports outside `/packages/grid-adapter`.
12. Hand edits to generated protocol artifacts.
13. Hand edits to generated ledgers.
14. Visual golden refreshes outside `docs/guides/cartulary_visual_golden_maintenance.md`.

## Row Inventory

The row inventory below is derived from the live FE-P10 map. It contains exactly
the six required FE-P10 rows, all currently blocked. The `required_for_closure`
flags in the current blocked map are not a bypass; row promotion requires
current owner input, current row-owned evidence, and no blockers.

| Row ID | Layer | Evidence class | Mapped targets | Owner refs | Evidence expectation | Closure and non-closure |
| --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P10-01` | unit | product_conformance | `frontend-unit` | Core 01 Section 8.5; Core 02 Sections 18-19; Core 03 Sections 2.2, 16.3, 16.4, 18, 19. REQs `REQ-01-296`, `REQ-01-302`, `REQ-01-499`, `REQ-01-509`, `REQ-02-222`, `REQ-02-232`, `REQ-03-005`, `REQ-03-011`, `REQ-03-250`, `REQ-03-260`, `REQ-03-265`, `REQ-03-274`; ACs `AC-076`, `AC-090`, `AC-116`, `AC-122`, `AC-137`, `AC-145`, `AC-231`, `AC-252`, `AC-253`, `AC-277`, `AC-287`, `AC-300`, `AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-410`, `AC-411`. | Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata. | Currently blocked by `frontend_phase_row_not_implemented`; only current `frontend-unit` row accounting can close after map and registry promotion. |
| `FE-B-P10-01` | browser_integration | product_conformance | `browser-e2e-webserver-backed` | Core 03 Sections 2.2, 16.3, 16.4; UI/UX Sections 6, 11. REQs `REQ-03-005`, `REQ-03-011`, `REQ-03-250`, `REQ-03-260`, `REQ-03-273`; ACs `AC-078`, `AC-080`, `AC-090`, `AC-121`, `AC-122`, `AC-137`, `AC-145`, `AC-231`, `AC-277`, `AC-287`, `AC-315`, `AC-319`, `AC-410`, `AC-411`, `AC-055`, `AC-058`. | Exact scenario title: `FE-B-P10-01 Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.` | Currently blocked; cannot close from Phase 4 or Phase 9 smoke, selector presence, generated ledgers, or plan text. |
| `FE-B-P10-02` | browser_integration | product_conformance | `browser-e2e-support`, `browser-e2e-webserver-backed` | Core 03 Sections 4.1, 14, 18, 19; UI/UX Sections 10, 14. REQs `REQ-03-217`, `REQ-03-235`, `REQ-03-263`, `REQ-03-265`; ACs `AC-003`, `AC-005`, `AC-013`, `AC-014`, `AC-024`, `AC-026`, `AC-040`, `AC-043`, `AC-044`, `AC-047`, `AC-231`, `AC-354`, `AC-360`, `AC-363`, `AC-364`, `AC-394`, `AC-396`, `AC-033`, `AC-039`, `AC-080`, `AC-086`. | Exact scenario title: `FE-B-P10-02 Verify full keyboard/clipboard contract: copy, paste, fill-down, frozen columns, virtual scroll, group rows, focus restoration, and Esc priority ladder.` | Currently blocked; support helper evidence must be current row-owned and cannot replace mapped product browser evidence. |
| `FE-E-P10-01` | e2e | product_conformance | `browser-e2e-webserver-backed`, `browser-e2e-stateful` | Core 01 Sections 3.3.4, 3.3.6, 8.5; Core 03 Sections 16.3-16.4; Core 04 Section 3. REQs `REQ-01-034`, `REQ-01-036`, `REQ-01-057`, `REQ-01-070`, `REQ-01-296`, `REQ-01-302`, `REQ-01-499`, `REQ-01-506`, `REQ-03-250`, `REQ-03-260`, `REQ-03-273`, `REQ-04-021`, `REQ-04-030`; ACs `AC-078`, `AC-080`, `AC-090`, `AC-121`, `AC-122`, `AC-124`, `AC-127`, `AC-137`, `AC-145`, `AC-149`, `AC-178`, `AC-180`, `AC-181`, `AC-183`, `AC-188`, `AC-190`, `AC-200`, `AC-218`, `AC-221`, `AC-225`, `AC-231`, `AC-238`, `AC-243`, `AC-277`, `AC-284`, `AC-299`, `AC-300`, `AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-340`, `AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402`. | Exact scenario title: `FE-E-P10-01 Verify coordination rows can be queried and edited through public view/row mutation contracts with current-role authorization.` | Currently blocked; must prove public `/api/v1/` query and edit plus action-time authorization, not frontend mocks. |
| `FE-V-P10-01` | visual | design_direction | `browser-e2e-visual` | UI/UX Sections 11, 13, 14; visual golden guide Sections 2, 3, 5. Design/support IDs `R2-AC-055`, `R2-AC-058`, `R2-AC-073`, `R2-AC-086`, `R2-RDG-AC-001`, `R2-RDG-AC-010`. | Exact scenario title: `FE-V-P10-01 Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and fill-down fixtures.` Fixtures `FE-VFIX-07`, `FE-VFIX-09`, `FE-VFIX-10`, `FE-VFIX-11`, `FE-VFIX-12`, `FE-VFIX-13` are current context only. | Currently blocked by `visual_fixture_not_recaptured_for_frontend_row`; visual readiness remains design-direction only. |
| `FE-A11Y-P10-01` | accessibility | design_direction | `browser-e2e-a11y-preflight` | UI/UX Sections 10, 11, 14; `docs/design.md` Accessibility Direction. Design IDs `R2-AC-033`, `R2-AC-039`, `R2-AC-055`, `R2-AC-058`, `R2-AC-080`, `R2-AC-086`, `D-AC-009`, `D-AC-012`. | Exact scenario title: `FE-A11Y-P10-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.` | Currently blocked; preflight must remain design-direction readiness and cannot promote product conformance. |

## Evidence Layer Matrix

| Evidence layer | Rows | Closure source | Cannot close from |
| --- | --- | --- | --- |
| Product conformance | `FE-U-P10-01`, `FE-B-P10-01`, `FE-B-P10-02`, `FE-E-P10-01` | Direct current row-owned frontend-row-accounting artifacts from mapped targets, exact scenario titles where required, matching registry/map/guide/ledger digests, target pass status, and no blockers. | This plan, generated ledgers, fixture registry state, screenshots, broad `make check`, old FE-P4/FE-P9 smoke, frontend mocks for public-route behavior, visual evidence, accessibility evidence. |
| Design direction | `FE-V-P10-01`, `FE-A11Y-P10-01` | Direct current row-owned visual or accessibility readiness evidence in mapped targets and fixture/accessibility artifacts where required. | Product-conformance claims, Core 05 claims, fixture registry `current` status alone, screenshots alone, preflight smoke without row-accounting authority. |
| Support evidence | Typecheck, import-boundary, helpers, target explanations, fixture registry, drift checks, command-surface inspection. | Supports implementation confidence and artifact hygiene. | Product or design row closure unless current mapped row-owned evidence also exists. |
| Claim publication | None for FE-P10. | Explicit Core 05 claim-bearing metadata and required Core 05 evidence, if a future owner decision adds them. | FE-P10 implementation, visual readiness, accessibility readiness, retained artifacts, broad checks, fixture status, or plan text. |

## Dependencies And Prerequisites

Before FE-P10 row promotion, verify:

1. FE-P0 through FE-P9 remain acceptable dependency context in the live registry.
2. The frontend guide Section 4.10 still matches the live FE-P10 map.
3. The FE-P10 map still contains exactly the six required row IDs and no
   invented rows.
4. The generated FE-P10 ledger agrees with the owner map after any safe
   Make-owned ledger refresh.
5. Visual fixture registry rows for FE-P10-owned fixtures remain current or
   have explicit blockers.
6. `browser-e2e-support` and `browser-e2e-a11y-preflight` are run directly
   when their mapped rows require evidence because they are not ordinary broad
   check closure substitutes.
7. Generated protocol consumption uses facade exports and no generated roots
   are hand-edited.
8. Stable selectors route through authored UI-contract surfaces, not ad hoc
   string construction where a local selector API exists.
9. Public-route product evidence uses `/api/v1/` routes, current public
   envelopes, `view_schema_id`, `record_id`, `field_key`, `row_version`, and
   `client_txn_id` where applicable.
10. Authorization denial evidence uses current server-managed sessions and
    action-time authorization derivation.
11. Focus restoration uses stable row/cell anchors and deterministic fallbacks.
12. Direct `react-data-grid` imports remain confined to `/packages/grid-adapter`.

## Shared Harness Analysis

FE-P10 spans all frontend evidence layers. It must use the frontend phase map
as row authority and the testing harness NLSpec for mechanics.

Harness mechanics:

1. `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P10` is the phase
   discovery source. It reports FE-P10 as planned and not executable as a whole
   phase.
2. `make explain-target TARGET=<target> DETAIL=summary` must be run before
   relying on target behavior.
3. Product row accounting must be produced by each mapped target needed for the
   implemented row. Old retained artifacts are diagnostic only.
4. Scenario title expectations from the FE-P10 map are required for rows that
   declare them, but title text alone is never evidence.
5. `browser-e2e-support` is an internal helper target and not a default
   `check` inclusion. Run it directly for the keyboard/clipboard support row.
6. `browser-e2e-a11y-preflight` is not a default `check` inclusion and reports
   no ordinary phase coverage in target explanation. Run it directly for the
   mapped accessibility readiness row.
7. Visual fixture lifecycle mechanics are owned by the harness and visual
   golden guide. Fixture registry status cannot close the visual row without
   current row-owned visual accounting.
8. Drift targets detect stale generated outputs and schedules; they do not
   create row evidence.
9. `make check` is end-of-run repository health. It does not close FE-P10 rows
   without current row-owned evidence.
10. `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` is
    required when retaining a successful full-check run for handoff.

Public-route product rows:

1. Browser coordination-surface evidence must use the same workbook shell and
   public view contracts, not detached route or mock-only assertions.
2. Keyboard, clipboard, fill-down, frozen-column, virtual-scroll, group-row,
   focus, and `Esc` evidence must prove stable anchors and deterministic
   behavior in the mapped browser targets.
3. E2E query/edit authorization evidence must prove public `/api/v1/` query and
   mutation behavior with current-role authorization.

## Public Interfaces And Deliverables

Expected FE-P10 implementation deliverables for future product work:

1. Contract-backed surface registrations for Task Requests, Decisions, Parties,
   Communications Log, Handoff, Status Review, and Lesson in the workbook shell.
2. Field mappings and closed vocabulary options keyed by stable `field_key` and
   contract metadata.
3. Same-shell navigation and view controls for all FE-P10 coordination and
   review surfaces.
4. Public query and row mutation flows for coordination rows using stable
   `view_schema_id`, `record_id`, `field_key`, `row_version`, and
   `client_txn_id`.
5. Current-role authorization evidence for query and edit operations.
6. Full keyboard behavior for editable cells, navigation, copy, paste,
   fill-down, frozen columns, virtual scroll, group rows, and focus return.
7. Deterministic `Esc` priority ladder: editor-local popup, editor draft,
   inspector, then no-op when none apply.
8. Presentation-only group rows that do not expose ordinary writable row
   affordances or mutation targets.
9. Visual readiness fixtures for the mapped visual states with deterministic
   seed data, viewport, scroll normalization, masks, and row-owned accounting.
10. Accessibility preflight readiness for keyboard reachability, focus
    visibility, accessible names, ARIA state, and non-color-only state.
11. Stable selector and test-helper updates needed for row-owned scenarios.
12. FE-P11 handoff notes with row evidence status, direct artifact roots,
    blockers, strict non-claims, and finalization outcome.

Generated protocol files, generated ledgers, generated schedules, lockfiles,
and visual goldens are not hand-edited deliverables. If implementation later
requires generated or contract-surface changes, update authored owner inputs
and run the repository-supported generator or drift targets.

## Sprint Checklist

| Status | Sprint | Focus | Exit posture |
| --- | --- | --- | --- |
| [ ] | 1. Readiness and source alignment | Verify guide, map, registry, ledger, fixture registry, prior handoff, Core authority, harness mechanics, target explanations, and blockers. | All live source facts are current; conflicts are recorded as blockers. |
| [ ] | 2. Unit/system-view registrations and field mappings | Verify stable `view_schema_id`, `field_key`, closed vocabulary, and contract metadata. | Unit row has direct mapped evidence or remains blocked with exact reason. |
| [ ] | 3. Browser coordination surfaces inside workbook shell | Open Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson inside the same workbook shell and retain controls. | Browser row has direct mapped evidence or remains blocked with exact reason. |
| [ ] | 4. Keyboard/clipboard/fill/frozen/virtual-scroll/group/focus/`Esc` | Complete full keyboard contract with stable anchors and deterministic priority ladder. | Keyboard row has direct mapped evidence in both mapped targets or remains blocked. |
| [ ] | 5. E2E public-route query/edit authorization | Prove coordination rows query and edit through public contracts with current-role authorization. | E2E row has direct mapped evidence in both mapped targets or remains blocked. |
| [ ] | 6. Visual readiness | Recapture mapped visual row-owned fixtures only through visual target and registry rules. | Visual row closes as design-direction readiness only or remains blocked. |
| [ ] | 7. Accessibility readiness | Run mapped preflight readiness for keyboard reachability, focus visibility, names, ARIA, and non-color-only state. | Accessibility row closes as design-direction readiness only or remains blocked. |
| [ ] | 8. Drift, full check, finalization, and FE-P11 handoff | Run drift gates, broad health, retained-run finalization, and handoff. | All row statuses, evidence roots, blockers, strict non-claims, and finalization outcome are recorded. |

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness And Source Alignment

Objective: lock live authority and make conflicts visible before implementation.

Actions:

1. Re-run `make task-guide` and `make explain-phase` for FE-P10.
2. Re-run target explanations for all mapped targets and maintenance targets.
3. Re-check guide, registry, map, ledger, fixture registry, Core 00 through Core
   05, domain, harness, design, UI/UX, dev, visual, and implementation/testing
   guides.
4. Verify prior FE-P9 handoff as dependency context only.
5. Record any stale digest, source mismatch, missing target, or ownership
   ambiguity as a blocker.

Exit condition: all FE-P10 owner inputs are current or blockers are explicit.

### Sprint 2: Unit/System-View Registrations And Field Mappings

Objective: prove registrations and field metadata use stable contract IDs.

Actions:

1. Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status
   Review, and Lesson surface registrations.
2. Verify field mappings use stable `field_key`, `view_schema_id`, and contract
   metadata.
3. Verify closed vocabulary options come from contract metadata or owner-backed
   mappings.
4. Add only row-owned unit scenarios required by the live map.

Exit condition: mapped unit evidence closes the unit row or the row remains
blocked with a precise missing-evidence reason.

### Sprint 3: Browser Coordination Surfaces Inside The Workbook Shell

Objective: prove the remaining coordination and review surfaces open inside the
same workbook shell and retain view controls.

Actions:

1. Exercise Task Requests, Decisions, Parties, Communications Log, Handoff,
   Status Review, and Lesson through workbook shell navigation.
2. Preserve active surface identity, system-view switcher context, saved-view
   controls, query controls, grid shell, inspector context where applicable, and
   status strip.
3. Use exact FE-P10 scenario titles from the live map.

Exit condition: mapped browser evidence closes the same-shell row or the row
remains blocked with a precise missing-evidence reason.

### Sprint 4: Keyboard, Clipboard, Fill, Frozen, Virtual Scroll, Group, Focus, And `Esc`

Objective: complete the full workbook interaction contract without row-index or
vendor-coordinate identity.

Actions:

1. Prove copy, paste, fill-down, keyboard navigation, and editor behavior bind
   to `record_id + field_key`.
2. Prove frozen columns and virtual scroll keep the intended cell/row anchors.
3. Prove group rows are presentation-only and do not expose ordinary mutation
   targets.
4. Prove focus restoration after inspector, editor, scroll, grouping, and
   route-refresh conditions.
5. Prove deterministic `Esc` priority: editor-local popup, editor draft,
   inspector, then no-op.
6. Run both mapped targets for the keyboard/clipboard row because support
   helper evidence cannot replace product browser evidence.

Exit condition: mapped browser evidence closes the keyboard row in both mapped
targets or the row remains blocked with a precise missing-evidence reason.

### Sprint 5: E2E Public-Route Query/Edit Authorization

Objective: prove coordination rows can be queried and edited through public
contracts with current-role authorization.

Actions:

1. Query coordination rows through public `/api/v1/` view routes.
2. Edit allowed fields through public row mutation contracts with
   `client_txn_id`, `base_row_version`, and field-keyed changes.
3. Verify denied operations use current-role authorization and public error
   envelopes.
4. Re-query after mutation to prove committed state and focus continuity.

Exit condition: mapped E2E evidence closes the public-route row in both mapped
targets or the row remains blocked with a precise missing-evidence reason.

### Sprint 6: Visual Readiness

Objective: capture design-direction fixtures for mapped FE-P10 visual states.

Actions:

1. Confirm fixture registry status and owners for `FE-VFIX-07`, `FE-VFIX-09`,
   `FE-VFIX-10`, `FE-VFIX-11`, `FE-VFIX-12`, and `FE-VFIX-13`.
2. Encode deterministic seed data, viewport, zoom, scroll normalization, focus
   state, dynamic masks, and screenshot scope.
3. Run `make browser-e2e-visual` and retain row-owned visual accounting.

Exit condition: mapped visual row closes as design-direction readiness only or
remains blocked with precise fixture/accounting reasons.

### Sprint 7: Accessibility Readiness

Objective: prove the mapped accessibility readiness scenario without promoting
it to product conformance.

Actions:

1. Verify keyboard reachability across coordination surfaces and full keyboard
   controls.
2. Verify focus visibility, accessible names, ARIA state, and non-color-only
   states.
3. Run `make browser-e2e-a11y-preflight` directly and retain row-owned
   accounting.

Exit condition: mapped accessibility row closes as design-direction readiness
only or remains blocked with precise missing-evidence reasons.

### Sprint 8: Drift, Full Check, Finalization, And FE-P11 Handoff

Objective: validate generated and harness state, run repository health gates,
and produce the next-phase handoff.

Actions:

1. Run generated-artifact, JSON-shape, generation, ledger, and schedule drift
   checks.
2. Run broad repository health after row-owned evidence is current.
3. Run `make agent-finalize RESULTS_DIR=<retained-check-run>` when retaining a
   successful full-check run.
4. If `RESULTS_DIR` is unset, explicitly report retained-run maintenance was
   skipped.
5. Record FE-P11 handoff with row statuses, evidence roots, blockers, strict
   non-claims, and finalization outcome.

Exit condition: FE-P10 has either binary closure or explicit blockers, and
FE-P11 receives a complete handoff.

## Validation Commands

Explanation commands:

```sh
make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P10
make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P10
make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P10
make explain-target TARGET=frontend-unit DETAIL=summary
make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary
make explain-target TARGET=browser-e2e-support DETAIL=summary
make explain-target TARGET=browser-e2e-stateful DETAIL=summary
make explain-target TARGET=browser-e2e-visual DETAIL=summary
make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary
make explain-target TARGET=generate-drift DETAIL=summary
make explain-target TARGET=generated-artifact-policy-check DETAIL=summary
make explain-target TARGET=json-shape-check DETAIL=summary
make explain-target TARGET=phase-ledger-drift DETAIL=summary
make explain-target TARGET=phase-schedule-drift DETAIL=summary
make explain-target TARGET=check DETAIL=summary
make explain-target TARGET=agent-finalize DETAIL=summary
```

Execution commands:

```sh
make frontend-unit
make browser-e2e-webserver-backed
make browser-e2e-support
make browser-e2e-stateful
make browser-e2e-visual
make browser-e2e-a11y-preflight
make frontend-typecheck
make frontend-import-boundary-check
make lint-markdown
make generated-artifact-policy-check
make json-shape-check
make generate-drift
make phase-ledger-drift
make phase-schedule-drift
make phase-ledgers
make check
make agent-finalize RESULTS_DIR=<retained-check-run>
```

`make phase-ledgers` is execution-only when authored owner inputs require a
safe generated-ledger refresh. Do not run it as a substitute for row evidence
or hand-edit the generated ledger.

## Evidence Requirements

Every closing FE-P10 artifact must be current to the live registry, map, guide,
and ledger digests and must use frontend row-accounting v3 where the harness
requires it.

Product row evidence must include:

1. The mapped FE-P10 row ID.
2. The mapped command ID and target.
3. Exact scenario title when `scenario_title_required=true`.
4. Passing target and row status.
5. Current guide, map, registry, and ledger digests.
6. No unresolved blocker for the row.
7. Public-route evidence when the row asserts query, edit, authorization, or
   mutation behavior through `/api/v1/`.

Visual readiness evidence must include:

1. Current visual row accounting from `browser-e2e-visual`.
2. Fixture registry linkage for FE-P10-owned fixture IDs.
3. Deterministic seed data, viewport, zoom, scroll normalization, focus/editor
   state, masks, and screenshot scope.
4. Explicit statement that the evidence is design-direction only.

Accessibility readiness evidence must include:

1. Current mapped preflight row accounting from `browser-e2e-a11y-preflight`.
2. Keyboard reachability, focus visibility, accessible-name, ARIA, and
   non-color-only state coverage.
3. Explicit statement that the evidence is design-direction only.

Broad check, retained artifacts, target explanations, and generated ledgers may
support handoff and repository health, but they do not close rows without the
current mapped row-owned evidence above.

## Blocker Rules

Record a blocker when any of these conditions occur:

1. Guide, registry, map, ledger, fixture registry, core, harness, or support
   guide sources conflict.
2. A digest is stale relative to the owner input being claimed.
3. A mapped target is missing or its target explanation conflicts with the map.
4. A required frontend-row-accounting artifact is missing, stale, v1/v2-only, or
   not tied to the mapped row.
5. A required exact scenario title is missing.
6. Generated protocol files, generated ledgers, generated schedules, lockfiles,
   or tool-managed artifacts are hand-edited.
7. Fixture registry ownership is missing, duplicated, stale, or contradicted by
   visual row evidence.
8. Visual or accessibility evidence is used as product conformance.
9. FE-P9 evidence, old retained artifacts, screenshots, or broad check are used
   as FE-P10 row closure.
10. A direct `react-data-grid` import appears outside `/packages/grid-adapter`.
11. Core 05, telemetry, design, or guide text attempts to redefine FE-P10
    product behavior.
12. Public-route behavior is claimed from frontend mocks only.
13. Writable row or cell behavior relies on row index, DOM order, visible label,
    group presentation position, vendor coordinate, or screenshot geometry.

## Strict Non-Claims

This plan does not claim:

1. FE-P10 implementation.
2. FE-P10 row closure.
3. FE-P10 product conformance.
4. Visual evidence as product conformance.
5. Accessibility evidence as product conformance.
6. Core 05 publication readiness.
7. Benchmark readiness.
8. Fixture-sensitive publication readiness.
9. Visual-publication readiness.
10. Extension-profile behavior.
11. New route semantics.
12. New security policy.
13. FE-P11 behavior.
14. Generated ledger correctness beyond inspected generated state.
15. Closure from old FE-P4, FE-P8, FE-P9, or broad `make check` artifacts.
16. Closure from fixture registry `current` status alone.
17. Closure from screenshots alone.
18. Closure from target explanations alone.
19. Closure from plan text.
20. Closure from generated ledgers.

## Binary Exit Criteria

FE-P10 exits only when all of the following are true:

1. All six live mapped FE-P10 rows have current row-owned evidence under the
   current registry, map, guide, and ledger digests.
2. Product rows have direct product-conformance evidence from their mapped
   targets.
3. Visual and accessibility rows close only as design-direction readiness.
4. All blockers are resolved or explicitly retained with owner disposition.
5. Generated ledgers, generated schedules, JSON shape, generated-artifact
   policy, and generation drift checks pass.
6. Broad `make check` passes after row-owned evidence is current.
7. `make agent-finalize RESULTS_DIR=<retained-check-run>` passes when a
   successful full-check run is retained, or retained-run maintenance is
   explicitly reported as skipped because `RESULTS_DIR` is unset.
8. FE-P11 receives a handoff with row status, direct evidence roots, blockers,
   strict non-claims, validation commands and results, generated maintenance
   outcome, and finalization outcome.

If any condition is false, FE-P10 remains blocked or incomplete.

## FE-P11 Handoff

The FE-P11 handoff must include:

1. Final FE-P10 registry status and row rollup state.
2. Final FE-P10 map, guide, registry, ledger, and fixture registry digests.
3. Row-by-row status for all six FE-P10 rows.
4. Direct evidence roots for each closed row.
5. Remaining blockers, if any, with reason codes and resolution owners.
6. Validation command results, including drift checks and broad health.
7. `agent-finalize` result and retained check root, or an explicit statement
   that retained-run maintenance was skipped because `RESULTS_DIR` was unset.
8. Strict non-claims that still apply after FE-P10.
9. Any safe generated-ledger refresh performed through Make-owned targets.
10. Confirmation that no generated protocol artifacts or generated ledgers were
    hand-edited.

FE-P11 must treat this plan and FE-P10 handoff as dependency context only. FE-P11
must still use its own owner map, direct row accounting, evidence classes, and
finalization.
