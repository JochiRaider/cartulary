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
Current FE-P10 final posture after Sprint 8 is `status=active`,
`row_rollup_state=active_green`, activation blockers `[]`, registry digest
`ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`,
map digest
`42cfca5547fda580aaf3388826fd0bf81d71bff33976480ebcf0a91d51067a5e`,
guide digest
`459ffe40579a33702160d51d5033145e15e48573dfc4e5c97c481e4781062462`,
generated ledger digest
`83768f981e59a603ef3619c4ab69afc7179aaf3c8978c78d277ead365ee34af2`,
and visual fixture registry digest
`4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c`.

All six FE-P10 rows are closed from current mapped row-owned evidence under
the final registry, map, guide, and ledger digests. `FE-U-P10-01`,
`FE-B-P10-01`, `FE-B-P10-02`, and `FE-E-P10-01` close only as
product-conformance rows from mapped product evidence. `FE-V-P10-01` and
`FE-A11Y-P10-01` close only as design-direction readiness; visual and
accessibility evidence is not product conformance, Core 05 evidence, benchmark
evidence, or publication evidence.

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
the six required FE-P10 rows. The `required_for_closure` flags are not a bypass;
row promotion requires current owner input, current row-owned evidence, and no
blockers.

| Row ID | Layer | Evidence class | Mapped targets | Owner refs | Evidence expectation | Closure and non-closure |
| --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P10-01` | unit | product_conformance | `frontend-unit` | Core 01 Section 8.5; Core 02 Sections 18-19; Core 03 Sections 2.2, 16.3, 16.4, 18, 19. REQs `REQ-01-296`, `REQ-01-302`, `REQ-01-499`, `REQ-01-509`, `REQ-02-222`, `REQ-02-232`, `REQ-03-005`, `REQ-03-011`, `REQ-03-250`, `REQ-03-260`, `REQ-03-265`, `REQ-03-274`; ACs `AC-076`, `AC-090`, `AC-116`, `AC-122`, `AC-137`, `AC-145`, `AC-231`, `AC-252`, `AC-253`, `AC-277`, `AC-287`, `AC-300`, `AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-410`, `AC-411`. | Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata. | Closed by current `frontend-unit` row accounting at `.cartulary/test-results/20260612T231550Z-p58785/frontend-unit/frontend-row-accounting.json`; registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `evidence_class=product_conformance`, `closure_status=closed`, and the row-owned scenario title passed. |
| `FE-B-P10-01` | browser_integration | product_conformance | `browser-e2e-webserver-backed` | Core 03 Sections 2.2, 16.3, 16.4; UI/UX Sections 6, 11. REQs `REQ-03-005`, `REQ-03-011`, `REQ-03-250`, `REQ-03-260`, `REQ-03-273`; ACs `AC-078`, `AC-080`, `AC-090`, `AC-121`, `AC-122`, `AC-137`, `AC-145`, `AC-231`, `AC-277`, `AC-287`, `AC-315`, `AC-319`, `AC-410`, `AC-411`, `AC-055`, `AC-058`. | Exact scenario title: `FE-B-P10-01 Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.` | Closed by current `browser-e2e-webserver-backed` row accounting at `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `evidence_class=product_conformance`, `closure_status=closed`, and the row-owned scenario title passed. |
| `FE-B-P10-02` | browser_integration | product_conformance | `browser-e2e-support`, `browser-e2e-webserver-backed` | Core 03 Sections 4.1, 14, 18, 19; UI/UX Sections 10, 14. REQs `REQ-03-217`, `REQ-03-235`, `REQ-03-263`, `REQ-03-265`; ACs `AC-003`, `AC-005`, `AC-013`, `AC-014`, `AC-024`, `AC-026`, `AC-040`, `AC-043`, `AC-044`, `AC-047`, `AC-231`, `AC-354`, `AC-360`, `AC-363`, `AC-364`, `AC-394`, `AC-396`, `AC-033`, `AC-039`, `AC-080`, `AC-086`. | Exact scenario title: `FE-B-P10-02 Verify full keyboard/clipboard contract: copy, paste, fill-down, frozen columns, virtual scroll, group rows, focus restoration, and Esc priority ladder.` | Closed by current row accounting in both mapped targets: `.cartulary/test-results/20260612T231610Z-p66811/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; both record registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `evidence_class=product_conformance`, `closure_status=closed`, and the row-owned scenario title passed. |
| `FE-E-P10-01` | e2e | product_conformance | `browser-e2e-webserver-backed`, `browser-e2e-stateful` | Core 01 Sections 3.3.4, 3.3.6, 8.5; Core 03 Sections 16.3-16.4; Core 04 Section 3. REQs `REQ-01-034`, `REQ-01-036`, `REQ-01-057`, `REQ-01-070`, `REQ-01-296`, `REQ-01-302`, `REQ-01-499`, `REQ-01-506`, `REQ-03-250`, `REQ-03-260`, `REQ-03-273`, `REQ-04-021`, `REQ-04-030`; ACs `AC-078`, `AC-080`, `AC-090`, `AC-121`, `AC-122`, `AC-124`, `AC-127`, `AC-137`, `AC-145`, `AC-149`, `AC-178`, `AC-180`, `AC-181`, `AC-183`, `AC-188`, `AC-190`, `AC-200`, `AC-218`, `AC-221`, `AC-225`, `AC-231`, `AC-238`, `AC-243`, `AC-277`, `AC-284`, `AC-299`, `AC-300`, `AC-303`, `AC-315`, `AC-318`, `AC-319`, `AC-340`, `AC-342`, `AC-352`, `AC-370`, `AC-371`, `AC-402`. | Exact scenario title: `FE-E-P10-01 Verify coordination rows can be queried and edited through public view/row mutation contracts with current-role authorization.` | Closed by current row accounting in both mapped targets: `.cartulary/test-results/20260612T231648Z-p70255/browser-e2e-stateful/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; both record registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `evidence_class=product_conformance`, `closure_status=closed`, and the row-owned scenario title passed. |
| `FE-V-P10-01` | visual | design_direction | `browser-e2e-visual` | UI/UX Sections 11, 13, 14; visual golden guide Sections 2, 3, 5. Design/support IDs `R2-AC-055`, `R2-AC-058`, `R2-AC-073`, `R2-AC-086`, `R2-RDG-AC-001`, `R2-RDG-AC-010`. | Exact scenario title: `FE-V-P10-01 Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and fill-down fixtures.` Fixtures `FE-VFIX-07`, `FE-VFIX-09`, `FE-VFIX-10`, `FE-VFIX-11`, `FE-VFIX-12`, `FE-VFIX-13` are current context only. | Closed by current `browser-e2e-visual` row accounting at `.cartulary/test-results/20260612T232112Z-p95497/browser-e2e-visual/frontend-row-accounting.json`; registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `evidence_class=design_direction`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`. This is design-direction readiness only and not product conformance. |
| `FE-A11Y-P10-01` | accessibility | design_direction | `browser-e2e-a11y` | UI/UX Sections 10, 11, 14; `docs/design.md` Accessibility Direction. Design IDs `R2-AC-033`, `R2-AC-039`, `R2-AC-055`, `R2-AC-058`, `R2-AC-080`, `R2-AC-086`, `D-AC-009`, `D-AC-012`. | Exact scenario title: `FE-A11Y-P10-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.` | Closed by current `browser-e2e-a11y` row accounting at `.cartulary/test-results/20260612T232241Z-p7636/browser-e2e-a11y/frontend-row-accounting.json`; registry digest `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477`, `evidence_class=design_direction`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`. This is design-direction readiness only and not product conformance. `browser-e2e-a11y-preflight` passed as non-closing smoke at `.cartulary/test-results/20260612T232350Z-p17497`. |

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
| [x] | 1. Readiness and source alignment | Verify guide, map, registry, ledger, fixture registry, prior handoff, Core authority, harness mechanics, target explanations, and blockers. | Completed 2026-06-12; live source facts match recorded digests, FE-P10 remains planned with row implementation/promotion blockers. |
| [x] | 2. Unit/system-view registrations and field mappings | Verify stable `view_schema_id`, `field_key`, closed vocabulary, and contract metadata. | Final current closure: `FE-U-P10-01` closed from `frontend-unit` row accounting at `.cartulary/test-results/20260612T231550Z-p58785/frontend-unit/frontend-row-accounting.json`. |
| [x] | 3. Browser coordination surfaces inside workbook shell | Open Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson inside the same workbook shell and retain controls. | Final current closure: `FE-B-P10-01` closed from `browser-e2e-webserver-backed` row accounting at `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`. |
| [x] | 4. Keyboard/clipboard/fill/frozen/virtual-scroll/group/focus/`Esc` | Complete full keyboard contract with stable anchors and deterministic priority ladder. | Final current closure: `FE-B-P10-02` closed from both mapped targets at `.cartulary/test-results/20260612T231610Z-p66811/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`. |
| [x] | 5. E2E public-route query/edit authorization | Prove coordination rows query and edit through public contracts with current-role authorization. | Final current closure: `FE-E-P10-01` closed from both mapped targets at `.cartulary/test-results/20260612T231648Z-p70255/browser-e2e-stateful/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`. |
| [x] | 6. Visual readiness | Recapture mapped visual row-owned fixtures only through visual target and registry rules. | Final current closure: `FE-V-P10-01` closed as design-direction readiness only from `browser-e2e-visual` row accounting at `.cartulary/test-results/20260612T232112Z-p95497/browser-e2e-visual/frontend-row-accounting.json`. |
| [x] | 7. Accessibility readiness | Run mapped accessibility readiness for keyboard reachability, focus visibility, names, ARIA, and non-color-only state. | Final current closure: `FE-A11Y-P10-01` closed as design-direction readiness only from `browser-e2e-a11y` row accounting at `.cartulary/test-results/20260612T232241Z-p7636/browser-e2e-a11y/frontend-row-accounting.json`; `browser-e2e-a11y-preflight` passed as non-closing smoke at `.cartulary/test-results/20260612T232350Z-p17497`. |
| [x] | 8. Drift, full check, finalization, and FE-P11 handoff | Run drift gates, broad health, retained-run finalization, and handoff. | Complete: post-finalize drift sweep passed, `make check` passed at `.cartulary/test-results/20260612T232539Z-p28882`, and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T232539Z-p28882` passed at `.cartulary/test-results/20260612T232735Z-p84216`. |

## Live Execution Tracker

### Sprint 1 Update: Readiness And Source Alignment

Status: complete with explicit blockers retained.

Commands run on 2026-06-12:

| Command | Result | Notes |
| --- | --- | --- |
| `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P10` | pass | Reports `status=planned`; recommended inspection targets are `explain-phase` and `phase-ledger-drift`. |
| `make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P10` | pass | Same planned-phase guidance as phase-author. |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P10` | pass | FE-P10 is planned and non-executable as a whole phase; all six rows are `claim_status=blocked`. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | pass | Public target; latest artifact `.cartulary/test-results/20260612T170650Z-p49228/frontend-unit/frontend-unit/phase-summary.json`; phase coverage does not include FE-P10. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | pass | Public target; latest artifact `.cartulary/test-results/20260612T170650Z-p49228/browser-e2e-webserver-backed/target-summary.json`; base phase coverage includes `phase10`, not frontend FE-P10 row closure. |
| `make explain-target TARGET=browser-e2e-support DETAIL=summary` | pass | Internal helper target; latest artifact `.cartulary/test-results/20260612T124750Z-p45347/browser-e2e-support/target-summary.json`; no default inclusion. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | pass | Public target; latest artifact `.cartulary/test-results/20260612T170650Z-p49228/browser-e2e-stateful/target-summary.json`; phase coverage does not include FE-P10. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | pass | Public target; latest artifact `.cartulary/test-results/20260612T170358Z-p28074/browser-e2e-visual/target-summary.json`; phase coverage does not include FE-P10. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | pass | Public target; latest artifact `.cartulary/test-results/20260612T045159Z-p12967/browser-e2e-a11y-preflight/target-summary.json`; reports `phase_coverage: none`. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T170650Z-p49228/generate-drift/generate-drift/phase-summary.json`. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T204004Z-p67861/generated-artifact-policy-check/generated-artifact-policy-check/phase-summary.json`. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T204004Z-p67892/json-shape-check/json-shape-check/phase-summary.json`. |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T204018Z-p68781/phase-ledger-drift/phase-ledger-drift/phase-summary.json`. |
| `make explain-target TARGET=phase-schedule-drift DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T185540Z-p37656/phase-schedule-drift/phase-schedule-drift/phase-summary.json`. |
| `make explain-target TARGET=check DETAIL=summary` | pass | Broad health target; latest artifact `.cartulary/test-results/20260612T170650Z-p49228/check/target-summary.json`; does not close FE-P10 rows. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | pass | Maintenance target; latest artifact `.cartulary/test-results/20260612T170847Z-p4303/agent-finalize/agent-finalize/phase-summary.json`; requires `RESULTS_DIR` for retained-run maintenance. |

Live source facts:

- Current digests match the plan baseline for the guide, registry, FE-P10 map, generated FE-P10 ledger, visual fixture registry, and harness NLSpec.
- FE-P10 registry remains `status=planned`, `row_rollup_state=no_rows_implemented`, with activation blocker `FE-P10-ACTIVATION-BLOCKER-01`.
- `tools/frontend_phase_maps/fe_p10_test_map.json` contains exactly six FE-P10 rows and all are blocked. Under the harness rules, standalone frontend-aware targets exclude blocked rows from active-target row accounting, so target passes alone cannot close FE-P10.
- Visual fixtures `FE-VFIX-07`, `FE-VFIX-09`, `FE-VFIX-10`, `FE-VFIX-11`, `FE-VFIX-12`, and `FE-VFIX-13` are present and `status=current`, but fixture registry status alone does not close `FE-V-P10-01`.

Sprint 1 blockers and non-claims:

- `FE-P10-ACTIVATION-BLOCKER-01` remains until row evidence and frontend freshness are promoted together.
- Each FE-P10 row retains its live map blocker (`frontend_phase_row_not_implemented` or `visual_fixture_not_recaptured_for_frontend_row`) until current mapped row-owned evidence exists.
- No product row is closed from plan text, generated ledger text, target explanation, fixture registry state, screenshots, old FE-P9 evidence, or broad `make check`.
- Visual and accessibility rows remain design-direction readiness only.

### Sprint 2 Update: Unit/System-View Registrations And Field Mappings

Status: complete for `FE-U-P10-01`.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/src/workbookSurfaceRegistry.test.ts` | Added the row-owned `FE-U-P10-01` unit scenario covering Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson registration, stable `view_schema_id`, stable `field_key`, closed enum vocabularies, and contract-backed writable metadata. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted only `FE-U-P10-01` to `claim_status=implemented`, `closure_scope=scenario`, `required_for_closure=true`, with the exact unit scenario title. |
| `tools/frontend_phase_registry.json` | Updated FE-P10 `row_rollup_state=partially_implemented` and refreshed FE-P10 map, ledger, and freshness digest mirrors after generated maintenance. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make frontend-unit` | pass | Preliminary root `.cartulary/test-results/20260612T205803Z-p76696`; closed `FE-U-P10-01` before generated-maintenance refresh. |
| `make explain-target TARGET=phase-ledgers DETAIL=summary` | pass | Confirmed `phase-ledgers` is a public helper-only maintenance target. |
| `make phase-ledgers` | fail, then resolved | First run root `.cartulary/test-results/20260612T205845Z-p78423` failed because `FE-P10.row_rollup_state` still said `no_rows_implemented`; registry was corrected to `partially_implemented`. |
| `make phase-ledgers` | pass | Regenerated ledger root `.cartulary/test-results/20260612T205855Z-p78843`. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T205931Z-p80918`; generated ledger matched authored inputs. |
| `make json-shape-check` | fail, then resolved | First run root `.cartulary/test-results/20260612T205931Z-p80920` failed because FE-P10 registry digest mirrors were stale; map, ledger, and freshness digests were refreshed. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T210011Z-p81926`; registry/map/ledger shape and freshness passed. |
| `make frontend-unit` | pass | Final row-owned evidence root `.cartulary/test-results/20260612T210011Z-p81949`. |

Final Sprint 2 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-U-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T210011Z-p81949/frontend-unit/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=frontend-unit`, `target_status=pass`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-U-P10-01 Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata.` |

Current source digests after Sprint 2:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `a0348ea388c192665be695a9d3f63b6310829155a54cfb1e8e91534446fcbaf6` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `caaa0c0f883e208a0740ca39c1b703ce3517f48c85c3ece4848fdc62faadf3f2` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `4dc625fec57bce9fc180e042c44f6bbc405fc6d895da01a34619c417d206e2d6` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map and registry changes.
- `make phase-ledger-drift` and `make json-shape-check` passed after the refresh and digest mirror update.
- No generated protocol artifacts, generated schedules, lockfiles, visual goldens, or tool-managed dependency artifacts were hand-edited.

Sprint 2 blockers and non-claims:

- `FE-B-P10-01`, `FE-B-P10-02`, `FE-E-P10-01`, `FE-V-P10-01`, and `FE-A11Y-P10-01` remain blocked in the live map.
- `FE-P10-ACTIVATION-BLOCKER-01` remains because the phase is only partially implemented.
- `FE-U-P10-01` closure comes only from current `frontend-unit` row accounting, not from the generated ledger, target explanation, plan text, broad check, visual evidence, or accessibility evidence.

### Sprint 3 Update: Browser Coordination Surfaces Inside Workbook Shell

Status: complete for `FE-B-P10-01`.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/e2e/phase9.sentinel.spec.ts` | Added the row-owned `FE-B-P10-01` Playwright scenario. The scenario opens Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson through the system-view switcher, keeps the same workbook shell ID, verifies active `view_schema_id`, grid shell, switcher options, view-bar saved-view selector, query controls, and status strip. |
| `apps/web/src/WorkbookShell.tsx` | Added the stable status-strip shell slot to generic workbook surfaces so FE-P10 coordination and review surfaces preserve the same status-strip contract as other workbook surfaces. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted only `FE-B-P10-01` to `claim_status=implemented`, `closure_scope=scenario`, `required_for_closure=true`, with the exact browser scenario title. |
| `tools/frontend_phase_registry.json` | Refreshed FE-P10 map, ledger, and freshness digest mirrors after generated maintenance. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make browser-e2e-webserver-backed` | fail, then resolved | Root `.cartulary/test-results/20260612T210656Z-p86272`; `FE-B-P10-01` failed because the test expected a `workbook-shell-slot-primary-grid` wrapper that generic surfaces did not render. Row accounting kept `closure_status=not_closed`. |
| `make browser-e2e-webserver-backed` | fail, then resolved | Root `.cartulary/test-results/20260612T210901Z-p552`; `FE-B-P10-01` failed because generic surfaces lacked the status-strip shell slot. Row accounting kept `closure_status=failed`. |
| `make browser-e2e-webserver-backed` | pass | Root `.cartulary/test-results/20260612T211128Z-p14763`; `FE-B-P10-01` closed before registry digest mirror refresh. |
| `make phase-ledgers` | pass | Root `.cartulary/test-results/20260612T211341Z-p29454`; regenerated the FE-P10 coverage ledger from authored inputs. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T211426Z-p30117`; registry, map, and freshness mirrors passed after FE-P10 digest update. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T211426Z-p30140`; generated ledger matched owner inputs. |
| `make frontend-unit` | pass | Root `.cartulary/test-results/20260612T211431Z-p30774`; refreshed `FE-U-P10-01` row-owned evidence under the current registry digest. |
| `make browser-e2e-webserver-backed` | pass | Final row-owned browser evidence root `.cartulary/test-results/20260612T211455Z-p32397`. |

Final Sprint 3 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-U-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T211431Z-p30774/frontend-unit/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=frontend-unit`, registry digest `6cfc98c711b6405e936f1f444d56369af13a86bf3543587946ea269222b6cf47`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the row-owned scenario title passed. |
| `FE-B-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T211455Z-p32397/browser-e2e-webserver-backed/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=browser-e2e-webserver-backed`, registry digest `6cfc98c711b6405e936f1f444d56369af13a86bf3543587946ea269222b6cf47`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-B-P10-01 Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.` |

Current source digests after Sprint 3:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `6cfc98c711b6405e936f1f444d56369af13a86bf3543587946ea269222b6cf47` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `9fd721e36ac7e4d05754fd13505f0a35628dd6a8317bb2f03f0ec339ba533ec8` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `f1bd78a81a24fc2149e521647a59892e78bab76a5ca60cd2319df1622ada2a7f` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map and registry changes.
- `make json-shape-check` and `make phase-ledger-drift` passed after the refresh and digest mirror update.
- No generated protocol artifacts, generated schedules, lockfiles, visual goldens, or tool-managed dependency artifacts were hand-edited.

Sprint 3 blockers and non-claims:

- `FE-B-P10-02`, `FE-E-P10-01`, `FE-V-P10-01`, and `FE-A11Y-P10-01` remain blocked in the live map.
- `FE-P10-ACTIVATION-BLOCKER-01` remains because the phase is only partially implemented.
- `FE-B-P10-01` closure comes only from current `browser-e2e-webserver-backed` row accounting, not from target explanation, generated ledger, screenshot state, fixture registry state, plan text, broad check, visual evidence, or accessibility evidence.
- The status-strip implementation change is product-surface support for current browser row evidence; it is not a Core 05 publication claim, benchmark claim, visual claim, or accessibility claim.

### Sprint 4 Update: Keyboard, Clipboard, Fill, Frozen, Virtual Scroll, Group, Focus, And `Esc`

Status: complete for `FE-B-P10-02`.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/e2e/phase9.keyboard.spec.ts` | Added the row-owned `FE-B-P10-02` Playwright scenario. The scenario proves copy, paste, fill-down, keyboard navigation, editor draft handling, inspector focus restoration, group-row presentation-only state, virtual scroll, frozen row gutter behavior, and the `Esc` priority ladder using stable `record_id`, `field_key`, `view_schema_id`, `base_row_version`, `row_version`, and `client_txn_id` contracts. |
| `apps/web/src/WorkbookShell.tsx` | Added editor copy handling, cell copy handling, and dirty-draft `Esc` cancellation before broader grid/inspector handling so copy and `Esc` behavior are anchored to the active editor/cell. |
| `packages/grid-adapter/src/index.tsx` | Added a stable gutter field key and sticky positioning for the row gutter so frozen-column behavior can be asserted from DOM/CSS contract state rather than screenshots or pointer geometry. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted `FE-B-P10-02` to `claim_status=implemented`, required both mapped targets for closure, and retained the exact scenario title from the live map. |
| `tools/phase3_test_map.json` | Added an explicit-only supplemental support selector row so `browser-e2e-support` runs the FE-P10 exact-title scenario. This shim is implementation-support selection only and does not create legacy Phase 3 product closure. |
| `scripts/lib/test-output/cli.mjs` | Adjusted Playwright classification precedence so frontend authoritative row ownership wins for authoritative phase runs while the legacy supplemental shim remains support-only. |
| `tools/frontend_phase_registry.json` | Refreshed FE-P10 map, ledger, and freshness digest mirrors after generated maintenance. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |
| `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` | Regenerated through `make phase-schedules` after JSON shape validation detected stale schedule inputs. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make browser-e2e-support` | fail, then resolved | Root `.cartulary/test-results/20260612T212442Z-p56286`; the exact FE-P10 scenario was not selected by the support target yet. |
| `make browser-e2e-support` | fail, then resolved | Root `.cartulary/test-results/20260612T213040Z-p69166`; first assertion failure while tightening the stable-anchor keyboard scenario. |
| `make browser-e2e-support` | fail, then resolved | Root `.cartulary/test-results/20260612T213327Z-p80674`; mark-reviewed setup was still tied to a brittle browser interaction and timed out. |
| `make browser-e2e-support` | fail, then resolved | Root `.cartulary/test-results/20260612T213643Z-p92823`; direct public-route setup used a stale base row version before the test re-read the row. |
| `make browser-e2e-support` | pass | Root `.cartulary/test-results/20260612T213751Z-p3681`; closed `FE-B-P10-02` before registry digest mirror refresh. |
| `make browser-e2e-webserver-backed` | fail, then resolved | Roots `.cartulary/test-results/20260612T213853Z-p7300` and `.cartulary/test-results/20260612T214202Z-p21479`; tests passed but the Phase 10 artifact check reported missing `FE-B-P10-02` because the support selector shim collided with authoritative selection. |
| `make browser-e2e-webserver-backed` | fail, then resolved | Root `.cartulary/test-results/20260612T214443Z-p35312`; a first classifier fix was too broad and produced unexpected IDs in other phase selections. |
| `make browser-e2e-webserver-backed` | pass | Root `.cartulary/test-results/20260612T214638Z-p48606`; closed `FE-B-P10-02` and retained `FE-B-P10-01` before registry digest mirror refresh. |
| `make phase-ledgers` | pass | Root `.cartulary/test-results/20260612T214833Z-p61925`; regenerated the FE-P10 coverage ledger from authored inputs. |
| `make json-shape-check` | fail, then resolved | Root `.cartulary/test-results/20260612T215046Z-p62769`; reported stale phase schedule inputs because `scripts/lib/test-output/cli.mjs` and `tools/phase3_test_map.json` changed. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T215046Z-p62792`; generated ledger matched owner inputs. |
| `make phase-schedules` | pass | Root `.cartulary/test-results/20260612T215053Z-p63441`; regenerated schedule outputs through the Make-owned generator. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T215103Z-p63665`; registry, map, ledger, and schedule freshness passed. |
| `make phase-schedule-drift` | pass | Root `.cartulary/test-results/20260612T215103Z-p63667`; regenerated schedule outputs matched authored inputs. |
| `make frontend-unit` | pass | Root `.cartulary/test-results/20260612T215111Z-p64179`; refreshed `FE-U-P10-01` row-owned evidence under registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`. |
| `make browser-e2e-support` | pass | Final support evidence root `.cartulary/test-results/20260612T215131Z-p71986`; `FE-B-P10-02` closed from mapped support row accounting under registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`. |
| `make browser-e2e-webserver-backed` | pass | Final browser evidence root `.cartulary/test-results/20260612T215225Z-p75465`; 81 tests, zero failed, zero missing, and `FE-B-P10-01` plus `FE-B-P10-02` closed from mapped row accounting under registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`. |

Final Sprint 4 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-U-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T215111Z-p64179/frontend-unit/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=frontend-unit`, registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-U-P10-01 Verify coordination and review system-view registrations, field mappings, and closed vocabulary options use stable IDs and contract metadata.` |
| `FE-B-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T215225Z-p75465/browser-e2e-webserver-backed/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=browser-e2e-webserver-backed`, registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-B-P10-01 Verify Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson open inside the same workbook shell and retain view controls.` |
| `FE-B-P10-02` | closed as product conformance | `.cartulary/test-results/20260612T215131Z-p71986/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260612T215225Z-p75465/browser-e2e-webserver-backed/frontend-row-accounting.json` both record `schema_id=cartulary.frontend_row_accounting.v3`, registry digest `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-B-P10-02 Verify full keyboard/clipboard contract: copy, paste, fill-down, frozen columns, virtual scroll, group rows, focus restoration, and Esc priority ladder.` |

Current source digests after Sprint 4:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `c27ba28719c96032032125343bfe1829cb07927d8a5fe6772db3c3fe5303bbf2` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `31ead66ed464d156bf0d8a1bc52e5bcde6db17ab0ad08cabed2687df71428d32` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `923dddaaab6cf77a116a864ee9d54af4eb46591c4da42dbf807e418438c50c88` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |
| `tools/scheduler_manifest.json` | `02b1b21597798697b56132d2387a242702a8d3d219899331fd5f0ec067924d34` |
| `tools/execution_topology_render_index.json` | `6b6b57af9643e06d1a4a1b9a45a644e35ddfdf8d413d264af436341880a8a573` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map and registry changes.
- `make phase-schedules` refreshed generated schedule/topology outputs after `json-shape-check` detected stale schedule inputs.
- `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift` passed after generated maintenance.
- No generated protocol artifacts, lockfiles, visual goldens, or tool-managed dependency artifacts were hand-edited.

Sprint 4 blockers and non-claims:

- `FE-E-P10-01`, `FE-V-P10-01`, and `FE-A11Y-P10-01` remain blocked in the live map.
- `FE-P10-ACTIVATION-BLOCKER-01` remains because the phase is only partially implemented.
- The Phase 3 support selector shim is only implementation-support target selection; it does not create or replace product closure for any legacy Phase 3 row.
- `FE-B-P10-02` closure comes only from current mapped `browser-e2e-support` and `browser-e2e-webserver-backed` frontend row accounting, not from support target explanation, generated ledger text, screenshot geometry, plan text, broad check, visual evidence, accessibility evidence, or old artifacts.

### Sprint 5 Update: E2E Public-Route Query/Edit Authorization

Status: complete for `FE-E-P10-01`.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/e2e/frontend.phase10.public-route.spec.ts` | Added the row-owned `FE-E-P10-01` Playwright scenario. The scenario seeds Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson records through public create routes, queries each coordination/review view through `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`, patches each row through `/api/v1/records/{record_id}` with `view_schema_id`, `base_row_version`, field-keyed `changes`, and `client_txn_id`, re-queries committed `row_version` and cell values, then demotes the same logged-in editor session to viewer and verifies a current-role `authorization_denied` public error envelope. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted `FE-E-P10-01` to `claim_status=implemented`, required both mapped targets for closure, and retained the exact scenario title from the live map. |
| `tools/frontend_phase_registry.json` | Refreshed FE-P10 map, ledger, and freshness digest mirrors after generated maintenance. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make frontend-typecheck` | pass | Root `.cartulary/test-results/20260612T220117Z-p91469`; new Playwright spec typechecked through the Make-owned frontend target. |
| `make phase-ledgers` | pass | Root `.cartulary/test-results/20260612T220131Z-p91904`; regenerated the FE-P10 coverage ledger from authored inputs. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T220155Z-p92350`; registry, map, ledger, and freshness mirrors passed. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T220155Z-p92372`; generated ledger matched owner inputs. |
| `make browser-e2e-stateful` | pass | Final stateful evidence root `.cartulary/test-results/20260612T220201Z-p93017`; 18 tests, zero failed, zero missing, and `FE-E-P10-01` closed from mapped row accounting under registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`. |
| `make browser-e2e-webserver-backed` | pass | Final browser evidence root `.cartulary/test-results/20260612T220355Z-p6609`; 82 tests, zero failed, zero missing, and `FE-B-P10-01`, `FE-B-P10-02`, and `FE-E-P10-01` closed from mapped row accounting under registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`. |
| `make frontend-unit` | pass | Root `.cartulary/test-results/20260612T220617Z-p20122`; refreshed `FE-U-P10-01` row-owned evidence under registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`. |
| `make browser-e2e-support` | pass | Root `.cartulary/test-results/20260612T220623Z-p27138`; refreshed the support-target half of `FE-B-P10-02` row-owned evidence under registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`. |

Final Sprint 5 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-U-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T220617Z-p20122/frontend-unit/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=frontend-unit`, registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the row-owned scenario title. |
| `FE-B-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T220355Z-p6609/browser-e2e-webserver-backed/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=browser-e2e-webserver-backed`, registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the row-owned scenario title. |
| `FE-B-P10-02` | closed as product conformance | `.cartulary/test-results/20260612T220623Z-p27138/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260612T220355Z-p6609/browser-e2e-webserver-backed/frontend-row-accounting.json` both record `schema_id=cartulary.frontend_row_accounting.v3`, registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and the row-owned scenario title. |
| `FE-E-P10-01` | closed as product conformance | `.cartulary/test-results/20260612T220201Z-p93017/browser-e2e-stateful/frontend-row-accounting.json` and `.cartulary/test-results/20260612T220355Z-p6609/browser-e2e-webserver-backed/frontend-row-accounting.json` both record `schema_id=cartulary.frontend_row_accounting.v3`, registry digest `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-E-P10-01 Verify coordination rows can be queried and edited through public view/row mutation contracts with current-role authorization.` |

Current source digests after Sprint 5:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `8baa2dae0014f3ec92610842027d767fed3a08b90d53caafc6260396023e5a8a` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `8800e389121372ab0410d89f06832a7668b067f8b511cfaa7b73653ed37dc1ac` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `4eed5bcd94eea122229342bd21a0210534affa2ccbd2b38e07162668399585d6` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map and registry changes.
- `make json-shape-check` and `make phase-ledger-drift` passed after the refresh and digest mirror update.
- No generated schedules were refreshed in Sprint 5 because shape validation did not report stale schedule inputs.
- No generated protocol artifacts, lockfiles, visual goldens, or tool-managed dependency artifacts were hand-edited.

Sprint 5 blockers and non-claims:

- `FE-V-P10-01` and `FE-A11Y-P10-01` remain blocked in the live map.
- `FE-P10-ACTIVATION-BLOCKER-01` remains because visual and accessibility readiness are still not complete.
- `FE-E-P10-01` closure comes only from current mapped `browser-e2e-stateful` and `browser-e2e-webserver-backed` frontend row accounting, not from frontend mocks, target explanation, generated ledger text, plan text, broad check, visual evidence, accessibility evidence, or old artifacts.
- The current-role authorization denial is product-route evidence for `/api/v1/records/{record_id}` only; it is not a new security policy, Core 05 claim, publication claim, or benchmark claim.

### Sprint 6 Update: Visual Readiness

Status: complete for `FE-V-P10-01` as design-direction readiness only.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/e2e/workbook.visual.spec.ts` | Added the row-owned `FE-V-P10-01` visual scenario. The scenario sets a deterministic `1440x900` viewport and 100% zoom, seeds Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson records, verifies party-link state, normalizes grid scroll/focus state, reuses existing golden scopes for `FE-VFIX-07` and `FE-VFIX-09` through `FE-VFIX-13`, and attaches a fixture matrix documenting seed, masks, focus/editor state, and screenshot scope. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted `FE-V-P10-01` to `claim_status=implemented`, required `browser-e2e-visual` for closure, and retained the exact visual scenario title from the live map. |
| `tools/frontend_phase_registry.json` | Refreshed FE-P10 map, ledger, and freshness digest mirrors after generated maintenance. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make frontend-typecheck` | pass | Root `.cartulary/test-results/20260612T221312Z-p33217`; new visual spec code typechecked through the Make-owned frontend target. |
| `make phase-ledgers` | pass | Root `.cartulary/test-results/20260612T221320Z-p33607`; regenerated the FE-P10 coverage ledger from authored inputs. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T221344Z-p34058`; registry, map, ledger, and freshness mirrors passed. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T221344Z-p34081`; generated ledger matched owner inputs. |
| `make browser-e2e-visual` | pass | Final visual evidence root `.cartulary/test-results/20260612T221349Z-p34715`; 25 tests, zero failed, zero missing, and `FE-V-P10-01` closed from mapped visual row accounting under registry digest `f04b86b22d0fcc1b3121a77ea667b6f5f2784bca250825fa160881ca2732ce37`. |

Final Sprint 6 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-V-P10-01` | closed as design-direction readiness only | `.cartulary/test-results/20260612T221349Z-p34715/browser-e2e-visual/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=browser-e2e-visual`, registry digest `f04b86b22d0fcc1b3121a77ea667b6f5f2784bca250825fa160881ca2732ce37`, `evidence_class=design_direction`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`, and closing scenario `FE-V-P10-01 Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and fill-down fixtures.` |

Current source digests after Sprint 6:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `f04b86b22d0fcc1b3121a77ea667b6f5f2784bca250825fa160881ca2732ce37` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `0d8bf8d4bc2c8d5451088112e5d5f30cf33f7b5c15bb4371209faf087d489879` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `b3c3c3457a3b4156fd11355e90b4b06a7db46250c79d4f0cd9e03ed7b93c3230` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `a6137698882ce7dab5209b4ee8b7d0a99bdb13bcede7f7c0ce0cf3d1d6bed9c7` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map and registry changes.
- `make json-shape-check` and `make phase-ledger-drift` passed after the refresh and digest mirror update.
- No visual golden files were hand-edited.
- No generated schedules, generated protocol artifacts, lockfiles, or tool-managed dependency artifacts were hand-edited.

Sprint 6 blockers and non-claims:

- `FE-A11Y-P10-01` remains blocked in the live map.
- `FE-P10-ACTIVATION-BLOCKER-01` remains because accessibility readiness is still not complete.
- `FE-V-P10-01` closure is design-direction readiness only. It does not close any product row, does not satisfy product conformance, and is not Core 05 publication or benchmark evidence.
- Product rows from Sprint 5 will be refreshed again after the final Sprint 7 registry change before FE-P10 final handoff.

### Sprint 7 Update: Accessibility Readiness

Sprint 7 found and resolved a live source conflict: the handoff sprint text
named `browser-e2e-a11y-preflight` as the closing target, but the current
frontend guide and manifest validator require implemented accessibility rows
to close from `browser-e2e-a11y`; preflight is blocked-row smoke only. The
authored FE-P10 guide row and live FE-P10 map were updated to
`browser-e2e-a11y`, and preflight was retained only as non-closing smoke.

Files changed:

| File | Purpose |
| --- | --- |
| `apps/web/e2e/workbook.a11y.spec.ts` | Added the row-owned `FE-A11Y-P10-01` scenario. The scenario seeds Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, Lesson, and a Timeline clipboard row; opens all FE-P10 coordination/review surfaces through the system-view switcher; verifies shell controls, focus visibility, accessible names, system-view menu ARIA, status/saved-view live regions, non-color-only filter chip state, and clipboard copy/paste readiness; and attaches a readiness matrix. |
| `packages/grid-adapter/src/index.tsx` | Added focused styling for sortable grid header buttons so generic grid headers expose a visible focus treatment during keyboard navigation. |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | Promoted `FE-A11Y-P10-01` to `claim_status=implemented`, required `browser-e2e-a11y` for closure, retained `evidence_class=design_direction`, and removed the row blocker. |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | Corrected the FE-P10 accessibility row target restatement from preflight smoke to implemented-row `make browser-e2e-a11y`. |
| `tools/frontend_phase_registry.json` | Moved FE-P10 to `row_rollup_state=activation_ready` while leaving `status=planned` and `FE-P10-ACTIVATION-BLOCKER-01` in place until final activation evidence is refreshed; refreshed guide, map, ledger, and freshness digest mirrors after guide/map changes. |
| `tools/frontend_phase_maps/fe_p0_test_map.json` through `tools/frontend_phase_maps/fe_p11_test_map.json` | Mechanically refreshed `guide_digest` mirrors after the frontend guide target correction. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | Regenerated through `make phase-ledgers`; not hand-edited. |

Commands run on 2026-06-12:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make frontend-typecheck` | fail, then resolved | Roots `.cartulary/test-results/20260612T222505Z-p49986` and `.cartulary/test-results/20260612T222542Z-p50668` caught the stale preflight target and a switcher group-token type issue; final passing root after the a11y scenario edits was `.cartulary/test-results/20260612T224639Z-p57278`. |
| `make phase-ledgers` | fail, then resolved | Root `.cartulary/test-results/20260612T222632Z-p51827` failed because the guide still restated `browser-e2e-a11y-preflight` while the live FE-P10 map used `browser-e2e-a11y`; root `.cartulary/test-results/20260612T222725Z-p52591` passed after the guide correction and digest refresh. |
| `make json-shape-check` | pass | Root `.cartulary/test-results/20260612T222745Z-p53014`; registry, maps, and freshness mirrors passed after guide/map/ledger maintenance. |
| `make phase-ledger-drift` | pass | Root `.cartulary/test-results/20260612T222749Z-p53352`; generated ledgers matched owner inputs. |
| `make phase-schedule-drift` | pass | Root `.cartulary/test-results/20260612T222754Z-p53690`; target schedule metadata remained current after the a11y target correction. |
| `make browser-e2e-a11y` | fail, then resolved | Iteration roots `.cartulary/test-results/20260612T222801Z-p53881`, `.cartulary/test-results/20260612T223015Z-p65013`, `.cartulary/test-results/20260612T223149Z-p75911`, `.cartulary/test-results/20260612T223405Z-p87237`, `.cartulary/test-results/20260612T223546Z-p98930`, `.cartulary/test-results/20260612T223737Z-p10922`, `.cartulary/test-results/20260612T223917Z-p22179`, `.cartulary/test-results/20260612T224116Z-p33422`, and `.cartulary/test-results/20260612T224413Z-p45619` exposed and resolved a11y scenario issues around sentinel tab traversal, generic grid focus visibility, generic cell semantics, unsupported generic grouping, and filtered-row inspect availability. Final passing root `.cartulary/test-results/20260612T224647Z-p57668`; 18 tests, zero failed, zero missing. |
| `make browser-e2e-a11y-preflight` | pass, non-closing smoke only | Root `.cartulary/test-results/20260612T224809Z-p68161`; 2 tests, zero failed, zero missing. This target is not FE-A11Y-P10-01 closure evidence under current guide/map rules. |

Final Sprint 7 row evidence:

| Row | Status | Direct evidence |
| --- | --- | --- |
| `FE-A11Y-P10-01` | closed as design-direction readiness only | `.cartulary/test-results/20260612T224647Z-p57668/browser-e2e-a11y/frontend-row-accounting.json` records `schema_id=cartulary.frontend_row_accounting.v3`, `target_name=browser-e2e-a11y`, registry digest `8370088a2ac98c534c07e1d7ab6e1d3b48cc8c915857964643a14fc891680078`, `evidence_class=design_direction`, `claim_status=implemented`, `target_status=pass`, `closure_status=closed`, and closing scenario `FE-A11Y-P10-01 Verify coordination surfaces and full keyboard/clipboard controls meet keyboard reachability, focus visibility, accessible-name, ARIA, and non-color-only state expectations.` |

Current source digests after Sprint 7:

| Source | SHA-256 |
| --- | --- |
| `tools/frontend_phase_registry.json` | `8370088a2ac98c534c07e1d7ab6e1d3b48cc8c915857964643a14fc891680078` |
| `tools/frontend_phase_maps/fe_p10_test_map.json` | `42cfca5547fda580aaf3388826fd0bf81d71bff33976480ebcf0a91d51067a5e` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p10_coverage_ledger.md` | `e340e6ccfa5707263ecb76bf971a515cd576ded281b640b512ec5d3a4f40388d` |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `459ffe40579a33702160d51d5033145e15e48573dfc4e5c97c481e4781062462` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Generated maintenance:

- `make phase-ledgers` refreshed the generated FE-P10 coverage ledger after authored map, guide, and registry changes.
- `make json-shape-check`, `make phase-ledger-drift`, and `make phase-schedule-drift` passed after digest and ledger maintenance.
- Frontend guide digest mirrors in all frontend phase maps were mechanically refreshed after the authored guide correction.
- No generated schedules, generated protocol artifacts, lockfiles, visual goldens, or tool-managed dependency artifacts were hand-edited.

Sprint 7 blockers and non-claims:

- `FE-P10-ACTIVATION-BLOCKER-01` remains because Sprint 8 still must promote final activation metadata and refresh all six rows under the final registry digest.
- `FE-A11Y-P10-01` closure is design-direction readiness only. It does not close any product row, does not satisfy product conformance, and is not Core 05 publication or benchmark evidence.
- `browser-e2e-a11y-preflight` is retained only as blocked-row smoke and is not FE-A11Y-P10-01 closure evidence.
- Product rows and the visual row will be refreshed again after Sprint 8 activation metadata changes before FE-P10 final handoff.

### Sprint 8 Update: Drift, Full Check, Finalization, And FE-P11 Handoff

Status: complete. FE-P10 is active, active-green, and has no remaining
activation blockers.

Final live source posture:

| Source | Final value |
| --- | --- |
| FE-P10 registry status | `active` |
| FE-P10 row rollup | `active_green` |
| FE-P10 activation blockers | `[]` |
| Registry digest | `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477` |
| FE-P10 map digest | `42cfca5547fda580aaf3388826fd0bf81d71bff33976480ebcf0a91d51067a5e` |
| FE-P10 guide digest | `459ffe40579a33702160d51d5033145e15e48573dfc4e5c97c481e4781062462` |
| FE-P10 ledger digest | `83768f981e59a603ef3619c4ab69afc7179aaf3c8978c78d277ead365ee34af2` |
| FE-P10 evidence freshness digest | `23cd90b9d276b65223da638e515e4ac562973b2d85f12439aca4e387102c849f` |
| Visual fixture registry digest | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |

Sprint 8 source changes:

| File | Change |
| --- | --- |
| `tools/frontend_phase_registry.json` | Promoted FE-P10 from `planned`/`activation_ready` to `active`/`active_green` and cleared activation blockers. |
| `apps/web/src/WorkbookShell.tsx` | Fixed timeline focus restoration so collection-cell navigation restores by stable `record_id + field_key` and attempts immediate focus before deferred fallback. |
| `apps/web/e2e/workbook.visual.spec.ts` | Stabilized the existing Phase 6 blocked-conflict visual fixture by explicitly scrolling the conflict resolver into the intended viewport before capture; no visual goldens were edited. |
| `tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/service_backed_make_target_duration_baselines.json` | Refreshed by `make agent-finalize` from retained successful `make check` run `.cartulary/test-results/20260612T232539Z-p28882`. |

Final row evidence:

| Row | Final status | Direct evidence |
| --- | --- | --- |
| `FE-U-P10-01` | closed, product_conformance | `.cartulary/test-results/20260612T231550Z-p58785/frontend-unit/frontend-row-accounting.json`; target `frontend-unit`, `target_status=pass`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`. |
| `FE-B-P10-01` | closed, product_conformance | `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; target `browser-e2e-webserver-backed`, `target_status=pass`, `claim_status_at_run=implemented`, `target_mapping_status=mapped`, `closure_status=closed`. |
| `FE-B-P10-02` | closed, product_conformance | `.cartulary/test-results/20260612T231610Z-p66811/browser-e2e-support/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; both mapped targets pass and close the exact row-owned scenario. |
| `FE-E-P10-01` | closed, product_conformance | `.cartulary/test-results/20260612T231648Z-p70255/browser-e2e-stateful/frontend-row-accounting.json` and `.cartulary/test-results/20260612T231357Z-p44274/browser-e2e-webserver-backed/frontend-row-accounting.json`; both mapped targets pass and close the exact public-route row-owned scenario. |
| `FE-V-P10-01` | closed, design_direction only | `.cartulary/test-results/20260612T232112Z-p95497/browser-e2e-visual/frontend-row-accounting.json`; `evidence_class=design_direction`, `closure_status=closed`. |
| `FE-A11Y-P10-01` | closed, design_direction only | `.cartulary/test-results/20260612T232241Z-p7636/browser-e2e-a11y/frontend-row-accounting.json`; `evidence_class=design_direction`, `closure_status=closed`. Non-closing smoke: `.cartulary/test-results/20260612T232350Z-p17497/browser-e2e-a11y-preflight/tool-run-summary.json`. |

Sprint 8 command results:

| Command | Result | Evidence or outcome |
| --- | --- | --- |
| `make phase-ledgers` | pass | Activation refresh root `.cartulary/test-results/20260612T225101Z-p78011`; Make-owned generated ledger refresh only. |
| `make frontend-typecheck` | pass | Post-focus-fix roots include `.cartulary/test-results/20260612T231332Z-p39485` and `.cartulary/test-results/20260612T232105Z-p94540`. |
| `make lint-biome` | pass | Post-format/post-fix roots include `.cartulary/test-results/20260612T231008Z-p17586`, `.cartulary/test-results/20260612T231332Z-p39486`, and `.cartulary/test-results/20260612T232105Z-p94539`. |
| `make lint-markdown` | pass | Final tracker validation root `.cartulary/test-results/20260612T233320Z-p89583`. |
| `make format` | pass | Root `.cartulary/test-results/20260612T231002Z-p16205`; applied Biome formatting/import organization to touched frontend files after the first `make check` failed on `lint-biome`. |
| `make frontend-unit` | pass | Final direct row evidence root `.cartulary/test-results/20260612T231550Z-p58785`. |
| `make browser-e2e-support` | pass | Final direct support evidence root `.cartulary/test-results/20260612T231610Z-p66811`; wrapper completed without summary stdout but row accounting records `target_status=pass`. |
| `make browser-e2e-webserver-backed` | pass | Final direct browser evidence root `.cartulary/test-results/20260612T231357Z-p44274`; 82 tests, 0 failed, 0 missing. |
| `make browser-e2e-stateful` | pass | Final direct stateful evidence root `.cartulary/test-results/20260612T231648Z-p70255`; 18 tests, 0 failed, 0 missing. |
| `make browser-e2e-visual` | pass | Final direct visual evidence root `.cartulary/test-results/20260612T232112Z-p95497`; 25 tests, 0 failed, 0 missing. Earlier post-fix visual run `.cartulary/test-results/20260612T231840Z-p83030` failed only `V-6-GRID-03` due viewport positioning and was resolved by explicit fixture scrolling without editing goldens. |
| `make browser-e2e-a11y` | pass | Final direct accessibility evidence root `.cartulary/test-results/20260612T232241Z-p7636`; 18 tests, 0 failed, 0 missing. |
| `make browser-e2e-a11y-preflight` | pass, non-closing | Final smoke root `.cartulary/test-results/20260612T232350Z-p17497`; 2 tests, 0 failed, 0 missing. |
| `make generated-artifact-policy-check` | pass | Final post-finalize root `.cartulary/test-results/20260612T232820Z-p86443`. |
| `make json-shape-check` | pass | Final post-finalize root `.cartulary/test-results/20260612T232823Z-p86628`. |
| `make generate-drift` | pass | Final post-finalize root `.cartulary/test-results/20260612T232831Z-p86972`. |
| `make phase-ledger-drift` | pass | Final post-finalize root `.cartulary/test-results/20260612T232838Z-p87835`. |
| `make phase-schedule-drift` | pass | Final post-finalize root `.cartulary/test-results/20260612T232843Z-p88173`. |
| `make check` | pass | Retained successful run `.cartulary/test-results/20260612T232539Z-p28882`; 149/149 work units, 804 tests, 0 failed, 0 missing. Earlier root `.cartulary/test-results/20260612T230908Z-p90537` failed `lint-biome` before formatting/import cleanup. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T232539Z-p28882` | pass | Finalizer root `.cartulary/test-results/20260612T232735Z-p84216`; retained run accepted as latest, run checks passed, four duration-baseline files refreshed, scheduler drift/json/ledger/schedule/duration-baseline validations passed. |

Sprint 8 blockers and non-claims:

- No FE-P10 activation blockers remain.
- No row blocker remains for the six live FE-P10 rows.
- Product closure is limited to `FE-U-P10-01`, `FE-B-P10-01`,
  `FE-B-P10-02`, and `FE-E-P10-01` from current mapped product row accounting.
- `FE-V-P10-01` and `FE-A11Y-P10-01` are design-direction readiness only and
  are not product conformance, Core 05 publication evidence, benchmark
  evidence, fixture-sensitive publication evidence, or visual-publication
  evidence.
- `browser-e2e-a11y-preflight`, visual fixture status, screenshots, broad
  `make check`, generated ledgers, target explanations, retained artifacts,
  and this plan are not row-closure evidence by themselves.
- No generated protocol artifacts, lockfiles, visual goldens, or tool-managed
  dependency artifacts were hand-edited.

Generated maintenance and finalization:

- `make phase-ledgers` and `make phase-schedules` were invoked only through
  Make-owned targets, including the `agent-finalize` structure refresh.
- `make agent-finalize` refreshed four advisory duration-baseline files from
  retained successful check run `.cartulary/test-results/20260612T232539Z-p28882`.
- Retained-run maintenance was not skipped; `RESULTS_DIR` was set and accepted
  as the latest retained successful check run.

FE-P11 handoff:

- FE-P11 receives FE-P10 as `active` and `active_green` with no blockers.
- FE-P11 receives the final row evidence roots listed above, but must treat
  them as dependency context only and use its own live map, evidence classes,
  row accounting, and finalization rules.
- FE-P11 must preserve the strict FE-P10 non-claims: no visual/product
  promotion, no accessibility/product promotion, no Core 05 publication claim,
  no benchmark claim, no fixture-sensitive publication claim, no route/security
  expansion claim, and no closure from plan text, screenshots, generated
  ledgers, target explanations, broad check, fixture registry status alone, or
  retained old artifacts.

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
3. Run `make browser-e2e-a11y` directly and retain row-owned accounting.
4. Run `make browser-e2e-a11y-preflight` only as non-closing smoke.

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
make explain-target TARGET=browser-e2e-a11y DETAIL=summary
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
make browser-e2e-a11y
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

1. Current mapped accessibility row accounting from `browser-e2e-a11y`.
2. Keyboard reachability, focus visibility, accessible-name, ARIA, and
   non-color-only state coverage.
3. Explicit statement that the evidence is design-direction only.
4. `browser-e2e-a11y-preflight` may be retained only as non-closing smoke.

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

This plan records the final FE-P10 closure evidence, but the plan text itself
is not closure evidence. The final handoff does not claim:

1. Product conformance for visual evidence.
2. Product conformance for accessibility evidence.
3. Core 05 publication readiness.
4. Benchmark readiness.
5. Fixture-sensitive publication readiness.
6. Visual-publication readiness.
7. Extension-profile behavior.
8. New route semantics.
9. New security policy.
10. FE-P11 behavior.
11. Generated ledger correctness beyond inspected generated state.
12. Closure from old FE-P4, FE-P8, FE-P9, or broad `make check` artifacts.
13. Closure from fixture registry `current` status alone.
14. Closure from screenshots alone.
15. Closure from target explanations alone.
16. Closure from plan text.
17. Closure from generated ledgers.
18. Closure from accessibility preflight smoke alone.

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
