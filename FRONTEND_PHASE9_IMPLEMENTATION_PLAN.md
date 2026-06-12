# Frontend Phase 9 Implementation Plan

## Summary

FE-P9 covers Inspector and Row-Local Actions for the Cartulary frontend. The planned user-observable scope is row-local inspector tabs for Details, Relationships, Evidence, and History, plus rollback preview/action and destructive-action confirmation/error behavior. This file is a planning artifact only. It does not implement FE-P9 product code and does not close any FE-P9 row.

Live repository inspection for this plan found FE-P9 in `tools/frontend_phase_registry.json` as `planned` with `row_rollup_state` `no_rows_implemented`. The phase is blocked by `FE-P9-ACTIVATION-BLOCKER-01`: frontend phase is not active until direct row evidence and freshness validation are promoted together. The live FE-P9 row map contains the five expected rows from `docs/guides/cartulary_frontend_implementation_testing_guide.md` Section 4.9: `FE-U-P9-01`, `FE-I-P9-01`, `FE-E-P9-01`, `FE-V-P9-01`, and `FE-A11Y-P9-01`. All five are currently `blocked`.

Current FE-P9 digests at plan creation:

| Source | Digest |
| --- | --- |
| `tools/frontend_phase_registry.json` | `864efbe415dc1853dc5d294c9eaa18b3f6463b8d1c1f3087cd62291ed25f84ba` |
| `tools/frontend_phase_maps/fe_p9_test_map.json` | `bf71d99e4500100ec7f484e1ee8c951e3eb5a5502767500dd0fd8fb33812f3e9` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p9_coverage_ledger.md` | `ccab8974ebfff0e80eacf106b6bf9d4ef10afc7daae824b2d73c520fae116e6e` |
| FE-P9 registry `evidence_freshness_digest` | `a5e01f742ddc4edb6a541e5c0dc01b1f2e0ce53881f4eeaa1186d81f849418a3` |
| Frontend guide digest in the FE-P9 map | `882c8fdd6c2127ee91e6f4713924d76d3ac1953e2078e469e468015ba34beeac` |

Broad `make check`, generated ledgers, retained old artifacts, fixture-registry status, screenshots, plan text, and target explanations do not close FE-P9 rows by themselves. Rows close only from current row-owned evidence under the current map, registry, and guide digests.

## Authority Model

The authority order for FE-P9 is:

| Order | Source | FE-P9 use |
| --- | --- | --- |
| 1 | Adopted Cartulary NLSpecs | Product and harness authority only inside their stated scope. |
| 2 | Core 00 through Core 04 under `docs/spec/` | Product-conformance behavior authority. |
| 3 | Core 05 | Inactive unless explicit claim-bearing timed, benchmark, fixture-sensitive, visual-publication, or publication metadata exists and satisfies Core 05. |
| 4 | `docs/domain.md` | Vocabulary and concept-boundary reference for records, evidence, relationships, history, rollback, and row-local actions. |
| 5 | `docs/guides/cartulary_frontend_implementation_testing_guide.md` | FE-P9 planning mechanics, row mapping, evidence classes, and row-owner mapping. |
| 6 | `docs/testing-harness-nlspec.md` | Command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, row accounting, and verification gates. |
| 7 | `docs/guides/cartulary_implementation_testing_guide.md` | Shared row/test/harness discipline and coverage-ledger shape. |
| 8 | `docs/guides/cartulary-dev-guide.md` | Frontend package boundaries, generated-artifact policy, Make targets, workspace shape, and implementation baseline. |
| 9 | `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` | Design-direction, visual, and accessibility readiness only. |
| 10 | Previous frontend implementation plans | Structural model and handoff context only. |
| 11 | Research reports | Rationale only. |

Core 00 through Core 04 own FE-P9 product conformance. Core 05 is inactive for FE-P9 absent explicit claim-bearing metadata. The testing harness owns harness mechanics and must not redefine product behavior. The frontend guide owns frontend planning mechanics but does not close rows. Generated ledgers are downstream rendered artifacts and must not be hand-edited or used as row-closure authority.

Conflicts between owner specs, guide text, live phase maps, generated ledgers, registries, retained artifacts, code, or command output must be recorded as blockers instead of silently reconciled. Visual and accessibility evidence remain design-direction evidence unless an owner map explicitly promotes a narrower claim, and no such FE-P9 promotion was inspected.

## Sprint 1 Repo Baseline

Initial live inspection occurred from repository root `/home/jochi/code/cartulary` on 2026-06-12. `FRONTEND_PHASE9_IMPLEMENTATION_PLAN.md` was absent before creation, and the worktree was clean before this authored Markdown file was added.

### Inspected Inputs

| Source | Live status at inspection | Digest or command anchor |
| --- | --- | --- |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` Section 4.9 | Present. Defines FE-P9 as Inspector and Row-Local Actions with five expected rows. | SHA256 `882c8fdd6c2127ee91e6f4713924d76d3ac1953e2078e469e468015ba34beeac`. |
| `FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` FE-P9 handoff | Present. Records FE-P8 complete and says FE-P9 must produce its own evidence for inspector behavior, rollback, evidence handles, Core 05, benchmark, and visual-publication readiness. | SHA256 `8bf970f55443ba7a9fe2fb2e7c9c364df80f9d9d6f380eb6e79f0cee3e4a6d9f`. |
| `tools/frontend_phase_registry.json` | Present. FE-P9 is `planned`, `no_rows_implemented`, and blocked by `FE-P9-ACTIVATION-BLOCKER-01`. | SHA256 `864efbe415dc1853dc5d294c9eaa18b3f6463b8d1c1f3087cd62291ed25f84ba`. |
| `tools/frontend_phase_maps/fe_p9_test_map.json` | Present. Five FE-P9 rows; all `claim_status` `blocked`. | SHA256 `bf71d99e4500100ec7f484e1ee8c951e3eb5a5502767500dd0fd8fb33812f3e9`. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p9_coverage_ledger.md` | Present. Generated downstream ledger reports FE-P9 `planned`, `no_rows_implemented`, and all five rows blocked. | SHA256 `ccab8974ebfff0e80eacf106b6bf9d4ef10afc7daae824b2d73c520fae116e6e`. |
| `tools/frontend_visual_fixture_registry.json` | Present. No FE-P9-owned visual fixture entries were found. | SHA256 `e03d4c105a1557059574c9b5ad980723ba312a9d2c4e3ff9beee2c0056394421`. |
| `docs/testing-harness-nlspec.md` | Present. Harness mechanics only. | SHA256 `f8857f2d67316ba43ac9c7da71040b26fb0f66250c991508b6570d3cf367af83`. |
| `docs/opentelemetry-instrumentation-nlspec.md` | Present. Telemetry-only NLSpec; no FE-P9 product behavior authority. | SHA256 `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451`. |
| Core 00 through Core 05 | Present under `docs/spec/`. Core 00 through Core 04 own product behavior; Core 05 remains publication-only. | Core 00 `e3b2e5e9ed4f47d29694612d571f3255437a9a1acbceb31fc38d9229756a682f`; Core 01 `1c55b261681c59e948356d8f80e2d3f5ab8936d33db5742d18a31f701a81bac9`; Core 02 `bb92665e26804b8c465d961fdef39b78e3f07c389a26a6478e9a210ce393d3fa`; Core 03 `fb561f66e61cf75e777a8c1c4d618d1064ca3b36e8d02435e85131ce631f5b10`; Core 04 `ab4d03966850879625141165d7902f108cfe989914e2b01ed42e2ff7968f6da1`; Core 05 `ee2f572430b75b41ccd20d4dede9c72251b3a4432db2ccf525bec9415da7ef89`. |
| `docs/domain.md` | Present. Domain vocabulary and concept-boundary reference only. | SHA256 `c461f865e5a0524865e691661f9049ceb9031226124aec60e4014b00847d0e21`. |
| `docs/guides/cartulary_implementation_testing_guide.md` | Present. Shared row/test/harness discipline. | SHA256 `050ac1da2a1fa9139a2b721a590405e510c82868bb8b9f358709ffde8a8caf8c`. |
| `docs/guides/cartulary-dev-guide.md` | Present. Frontend package and generated-artifact boundaries. | SHA256 `a4b8fb4b9e3b03c905ed276d19a692559ddf1e70396f224f8f8b2a3f68e58776`. |
| `docs/guides/cartulary-ui-ux-design-guide.md` | Present. Design direction only. | SHA256 `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26`. |
| `docs/design.md` | Present. Design token and visual/accessibility direction only. | SHA256 `e28345fac8ba22fc58264454237af209360a84af0c714ff4e1c94c6028d8cd05`. |
| `docs/guides/cartulary_visual_golden_maintenance.md` | Present. Visual golden maintenance only. | SHA256 `21faf12489ef6a9ee93e10197ad8197ceb1a1da8a7303a5fc693ac71db5ce5c1`. |

### Registry Rollup

| Phase | Registry status | Row rollup | Blockers |
| --- | --- | --- | --- |
| FE-P0 | `active` | `active_green` | None listed. |
| FE-P1 | `active` | `active_green` | None listed. |
| FE-P2 | `active` | `active_green` | None listed. |
| FE-P3 | `active` | `active_green` | None listed. |
| FE-P4 | `active` | `active_green` | None listed. |
| FE-P5 | `active` | `active_green` | None listed. |
| FE-P6 | `active` | `active_green` | None listed. |
| FE-P7 | `active` | `active_green` | None listed. |
| FE-P8 | `active` | `active_green` | None listed. |
| FE-P9 | `planned` | `no_rows_implemented` | `FE-P9-ACTIVATION-BLOCKER-01`. |

### Target Explanation Status

All explanation commands listed below were run for planning and exited with status 0.

| Command | Live finding |
| --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P9` | FE-P9 is `planned`; all five rows are blocked; planned frontend phases are explainable but not executable. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | Public fast target. Phase coverage includes phase9. Latest artifact `.cartulary/test-results/20260612T124559Z-p85431/frontend-unit/frontend-unit/phase-summary.json`. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | Public full-gate target. Phase coverage includes phase9. Latest artifact `.cartulary/test-results/20260612T124559Z-p85431/browser-e2e-webserver-backed/target-summary.json`. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | Public full-gate target. Phase coverage lists phase1, phase6, and phase8, while the FE-P9 map names this target for `FE-E-P9-01`; future closure must rely on current FE-P9 row accounting. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | Public full-gate visual target. Phase coverage omits phase9, while the FE-P9 map names this target for `FE-V-P9-01`; future closure must rely on current FE-P9 row accounting. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | Public preflight target. Phase coverage `none`; the FE-P9 map names it for `FE-A11Y-P9-01`. |
| `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` | Public accessibility target. Not the current mapped FE-P9 accessibility target. |
| `make explain-target TARGET=browser-e2e-support DETAIL=summary` | Internal helper; no FE-P9 rows currently map to it. |
| `make explain-target TARGET=phase-ledgers DETAIL=summary` | Public phase-maintenance helper. |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | Public phase-maintenance target. |
| `make explain-target TARGET=phase-schedule-drift DETAIL=summary` | Public phase-maintenance target. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | Public generated-artifact drift target. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | Public generated-artifact policy target. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | Public JSON and manifest shape target. |
| `make explain-target TARGET=phase-map-check DETAIL=summary` | Internal helper explainable through Make. |
| `make explain-target TARGET=check DETAIL=summary` | Public full-gate target. Latest artifact `.cartulary/test-results/20260612T124559Z-p85431/check/target-summary.json`. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | Public helper target requiring `RESULTS_DIR` when retained-run maintenance is intended. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | Public fast-verification target. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | Public fast-verification target. |
| `make explain-target TARGET=lint-markdown DETAIL=summary` | Public fast-verification target for authored Markdown changes. |
| `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P9` | For planned FE-P9, narrows source validation to `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P9` and `make phase-ledger-drift`. |
| `make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P9` | Same planned-phase guidance as phase-author. |

### Retained Artifact Context

| Artifact root | Inspected finding | FE-P9 use |
| --- | --- | --- |
| `.cartulary/test-results/20260612T124559Z-p85431` | `make check` passed with 145/145 work units, 791 tests, 0 failed. The `frontend-unit`, `browser-e2e-webserver-backed`, and `browser-e2e-stateful` row-accounting files pass but contain no FE-P9 `row_results`. | Health context only. Must not close FE-P9 rows. |
| `.cartulary/test-results/20260612T124828Z-p48784` | `browser-e2e-visual` passed; row accounting contains no FE-P9 `row_results`. | Visual context only. Must not close `FE-V-P9-01`. |
| `.cartulary/test-results/20260612T124950Z-p59875` | `browser-e2e-a11y` passed; row accounting and accessibility summary contain no FE-P9 rows. | Accessibility context only. Not the current mapped FE-P9 accessibility target. |
| `.cartulary/test-results/20260612T045159Z-p12967` | `browser-e2e-a11y-preflight` passed and the preflight summary includes `FE-A11Y-P9-01` scenario `pass`, but the row remains `claim_status` `blocked` and no `frontend-row-accounting.json` exists for this target. | Readiness smoke only. Must not close `FE-A11Y-P9-01`. |
| `.cartulary/test-results/20260612T125604Z-p81938` | `make agent-finalize` passed for `RESULTS_DIR=.cartulary/test-results/20260612T124559Z-p85431`. Finalizer refreshed generated maintenance artifacts from that retained run. | Retained-run maintenance context only. Must not close FE-P9 rows. |

### Relevant Current Code Context

The repo already contains FE-P9-adjacent support code, selectors, and tests. These are implementation context only unless future FE-P9 row accounting maps them to current rows.

| Path | Inspected relevance | Current limitation |
| --- | --- | --- |
| `apps/web/src/WorkbookShell.tsx` | Existing inspector, relationship chip rendering, row history fetch through `/api/v1/records/{record_id}/history`, rollback through `/api/v1/records/{record_id}/rollback`, soft delete through `DELETE /api/v1/records/{record_id}`, restore through `POST /api/v1/records/{record_id}/restore`, public error parsing, and viewport continuity. | Existing code is not FE-P9 closure without current row-owned evidence. |
| `packages/ui-contracts/src/index.ts` | Stable selector builders exist for inspector sections, row-history actions, relationship items, evidence affordances, and inspector-surface row cells. | Selector existence alone does not close FE-P9 rows. |
| `apps/web/src/WorkbookShell.phase9.sentinel.test.tsx` | Phase 9-adjacent unit/support tests exist for keyboard and grid anchors. | Scenario names do not match FE-P9 row IDs and current row accounting contains no FE-P9 rows. |
| `apps/web/e2e/phase9.keyboard.spec.ts` and `apps/web/e2e/phase9.sentinel.spec.ts` | Phase 9-adjacent browser tests exist for keyboard, paste, and workbook-native workflows. | These are not the FE-P9 Inspector and Row-Local Actions rows in the live FE-P9 map. |
| `apps/web/e2e/phase7.history.spec.ts` | Existing row history and rollback browser coverage through public routes. | FE-P7 evidence is dependency context only. FE-P9 must produce its own row-owned evidence. |
| `apps/web/e2e/workbook.visual.spec.ts` and `apps/web/e2e/workbook.a11y-preflight.spec.ts` | Existing visual and accessibility readiness surfaces include inspector-adjacent checks. | No FE-P9-owned visual fixture entries were found, and FE-P9 visual/a11y rows remain blocked. |

## Source Limits

1. `tools/frontend_visual_fixture_registry.json` is present but has no FE-P9-owned fixture entries. `FE-V-P9-01` is blocked until fixture ownership and current visual row accounting exist.
2. Retained row-accounting artifacts from `.cartulary/test-results/20260612T124559Z-p85431`, `.cartulary/test-results/20260612T124828Z-p48784`, and `.cartulary/test-results/20260612T124950Z-p59875` pass but contain no FE-P9 `row_results`. They are health context only.
3. `.cartulary/test-results/20260612T045159Z-p12967` preflight includes `FE-A11Y-P9-01` scenario `pass`, but the row remains blocked and the target has no `frontend-row-accounting.json`.
4. `browser-e2e-stateful`, `browser-e2e-visual`, and `browser-e2e-a11y-preflight` target explanations do not advertise phase9 coverage even though the FE-P9 map names them. Closure must come from current map-authorized row accounting.
5. Generated FE-P9 ledger exists and matches the planned/blocked posture, but it is downstream only and must not be hand-edited or used as closure authority.
6. Existing FE-P9-adjacent support tests do not close FE-P9 rows unless future current row accounting maps them to the live FE-P9 row IDs.
7. FE-P8 is complete and active/green, but FE-P8 evidence is dependency context only and does not prove FE-P9 inspector behavior.
8. FE-P7 history/rollback browser evidence is dependency context only and does not prove FE-P9 inspector integration, destructive-action confirmation, or row-local focus behavior.
9. Public-route FE-P9 behavior cannot close from frontend mocks alone. Browser-facing product evidence must use public `/api/v1/` routes and public envelopes where mapped.
10. Visual and accessibility evidence remain design-direction evidence only. They must not be represented as Core product conformance.
11. Core 05 claim-publication evidence is absent. FE-P9 must not claim Core 05 publication, benchmark, fixture-sensitive publication, or visual-publication readiness.
12. Broad `make check`, target explanations, screenshots, plan text, fixture status, old retained artifacts, and generated ledgers do not close FE-P9 rows.

## FE-P8 Handoff Inputs

`FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` is handoff context only. FE-P9 may rely on FE-P8 as dependency health but must produce direct FE-P9 evidence for Inspector and Row-Local Actions.

| FE-P8 handoff item | Recorded FE-P8 value | FE-P9 use |
| --- | --- | --- |
| Registry status | `active`. | Dependency context only. |
| Row rollup | `active_green`. | Dependency context only. |
| Activation blocker | Cleared; `FE-P8-ACTIVATION-BLOCKER-01` removed. | Dependency context only. |
| Product evidence roots | FE-U-P8-01 and FE-I-P8-01 in `.cartulary/test-results/20260612T044428Z-p24246/frontend-unit/frontend-row-accounting.json`; FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 in `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-webserver-backed/frontend-row-accounting.json`; FE-E-P8-01 in `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-stateful/frontend-row-accounting.json`; FE-B-P8-01 in `.cartulary/test-results/20260612T044835Z-p88067/browser-e2e-support/frontend-row-accounting.json`. | Shows saved-view/query-control dependency health only. |
| Design-direction evidence roots | FE-V-P8-01 in `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json`; FE-A11Y-P8-01 in `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json`. | Design context only. |
| Final digest set | Registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`; frontend guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`; FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`. | FE-P8 closure context only. |
| Full check and finalization | `make check` passed at `.cartulary/test-results/20260612T044428Z-p24246`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246` passed at `.cartulary/test-results/20260612T045247Z-p22120`. | Historical retained context only; FE-P9 must finalize its own retained run later. |
| Strict non-claims | No visual/product promotion, no accessibility/product promotion, no Core 05 publication, no evidence handle readiness, no rollback readiness, no benchmark/fixture-sensitive publication claim, no FE-P9 implementation claim, and no direct `react-data-grid` import permission outside `packages/grid-adapter`. | These non-claims remain active until FE-P9 produces direct row-owned evidence where applicable. |

## Phase Objective

FE-P9 must deliver and verify Inspector and Row-Local Actions through current row-owned evidence. The objective is a row-local inspector anchored by `record_id`, with Details, Relationships, Evidence, and History tabs or sections, rollback preview/action, and destructive-action confirmation/errors that use stable identifiers and public route contracts.

Completion requires:

1. Inspector selection, tab state, details, relationships, evidence, and history anchors are based on `record_id`, not visible row index, DOM order, labels, or grid vendor coordinates.
2. Relationship items and evidence items use stable relationship/evidence IDs in selectors and action payloads.
3. History, rollback, delete, and restore behavior use public route contracts, public success envelopes, and public error envelopes.
4. Destructive actions re-check current authorization and base row version at action time.
5. Focus returns to the originating row/cell when appropriate after closing, rollback, delete/restore, error, or refresh.
6. FE-P0 through FE-P8 remain green or any dependency regression is recorded as a blocker.
7. Product-conformance rows close only from current FE-P9 row-owned evidence in the mapped targets.

## Implementation Scope

FE-P9 may change these implementation surfaces during future product work:

| Surface | FE-P9 scope |
| --- | --- |
| `/apps/web` | Timeline workbook inspector, row-local tab/section state, details rendering, relationship links, evidence panel integration, history timeline, rollback preview/action, destructive confirmation/errors, focus return, public route calls, browser/E2E scenarios, visual and accessibility specs. |
| `/packages/protocol-ts` | Public protocol facade consumption for row history, rollback, evidence, delete/restore, and public envelope types. Generated protocol files must not be hand-edited. |
| `/packages/ui-contracts` | Stable selectors and test IDs for inspector sections, relationship/evidence controls, history items, rollback actions, destructive actions, public errors, and focus anchors. |
| `/packages/test-utils` | Browser helpers for opening row-local inspector sections, asserting `record_id` anchoring, exercising rollback/destructive actions, verifying focus return, and checking stable selector identity. |
| `/packages/ui` | Reusable presentational components only where an existing shared UI component is appropriate. |
| `apps/web/e2e` | FE-P9 browser, stateful, visual, and accessibility scenarios with exact live row IDs and scenario titles from the FE-P9 map. |

FE-P9 behavior must use stable `record_id`, `field_key`, `history_item_ref`, relationship item refs, evidence IDs, `row_version`, `client_txn_id`, and public route identifiers where applicable. It must not depend on visible labels, visible row index, DOM order, CSS coordinates, or direct `react-data-grid` internals.

## Out of Scope

FE-P9 must not include:

1. Coordination workbook surfaces beyond relationships.
2. WebSocket live updates.
3. Core 05 publication evidence.
4. Benchmark claims.
5. Fixture-sensitive publication claims.
6. Visual-publication claims.
7. Product-conformance claims from visual evidence.
8. Product-conformance claims from accessibility evidence.
9. FE-P10 remaining workbook surfaces or keyboard-completion work except handoff notes.
10. Generated protocol file hand edits.
11. Generated ledger hand edits.
12. Visual golden updates without following `docs/guides/cartulary_visual_golden_maintenance.md`.
13. Direct `react-data-grid` imports outside `/packages/grid-adapter`.

## Row Inventory

The row inventory below is derived from the live FE-P9 map. It matches the expected FE-P9 row IDs in the frontend guide. All rows are currently blocked.

| Row ID | Layer | Evidence class | Owner sections | Exact REQ IDs | Exact AC IDs | Mapped target or TODO | Current `claim_status` | Current blocker summary |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P9-01 | unit | product_conformance | Core 01 Sections 3.3.4, 3.3.7, 3.3.8; Core 03 Sections 16.1 and 16.2 | REQ-01-034, REQ-01-036, REQ-01-048, REQ-01-056, REQ-01-243, REQ-01-247, REQ-03-242, REQ-03-249, REQ-03-272 | AC-006, AC-015, AC-020, AC-023, AC-045, AC-072, AC-075, AC-097, AC-100, AC-102, AC-103, AC-107, AC-111, AC-124, AC-127, AC-154, AC-155, AC-184, AC-185, AC-209, AC-210, AC-231, AC-278, AC-280, AC-313, AC-315, AC-318, AC-321, AC-366, AC-367, AC-372, AC-374 | `make frontend-unit` | blocked | `FE-U-P9-01-BLOCKER-01`: direct implementation evidence has not been added. |
| FE-I-P9-01 | integration | product_conformance | Core 01 Section 3.3.7; Core 02 Section 15; Core 03 Section 16.2 | REQ-01-048, REQ-01-056, REQ-01-089, REQ-01-108, REQ-02-205, REQ-02-218, REQ-02-238, REQ-02-242 | AC-107, AC-111, AC-124, AC-128, AC-154, AC-155, AC-215, AC-218, AC-231, AC-383, AC-386, AC-412 | `make frontend-unit`; `make browser-e2e-webserver-backed` | blocked | `FE-I-P9-01-BLOCKER-01`: direct implementation evidence has not been added. |
| FE-E-P9-01 | e2e | product_conformance | Core 01 Sections 3.3.6, 3.3.7, 3.3.8; Core 02 Sections 13 and 15; Core 04 Sections 3 and 5 | REQ-01-057, REQ-01-070, REQ-01-089, REQ-01-108, REQ-01-243, REQ-01-247, REQ-02-186, REQ-02-201, REQ-02-205, REQ-02-218, REQ-04-021, REQ-04-030, REQ-04-052, REQ-04-053 | AC-049, AC-055, AC-102, AC-103, AC-107, AC-111, AC-124, AC-128, AC-149, AC-154, AC-155, AC-178, AC-180, AC-181, AC-183, AC-188, AC-190, AC-200, AC-218, AC-221, AC-225, AC-231, AC-232, AC-234, AC-251, AC-257, AC-260, AC-261, AC-278, AC-280, AC-299, AC-313, AC-321, AC-340, AC-342, AC-352, AC-370, AC-371, AC-402 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | blocked | `FE-E-P9-01-BLOCKER-01`: direct implementation evidence has not been added. |
| FE-V-P9-01 | visual | design_direction | UI/UX guide Sections 9, 12, 13; visual golden guide Sections 2, 3, 5 | N/A: design direction | R2-AC-023, R2-AC-026, R2-AC-059, R2-AC-062, R2-AC-073, R2-AC-079 | `make browser-e2e-visual`; fixture IDs from `tools/frontend_visual_fixture_registry.json` | blocked | `FE-V-P9-01-BLOCKER-01`: visual fixture planned, but direct frontend row-accounting evidence has not been recaptured under a closed fixture registry; no FE-P9 fixture owner entry was inspected. |
| FE-A11Y-P9-01 | accessibility | design_direction | UI/UX guide Sections 9, 12, 14 | N/A: design direction | R2-AC-080, R2-AC-086, D-AC-009, D-AC-012 | `make browser-e2e-a11y-preflight` | blocked | `FE-A11Y-P9-01-BLOCKER-01`: direct implementation evidence has not been added. |

## Evidence Layer Matrix

| Evidence layer | FE-P9 rows | Closure source | Cannot close from |
| --- | --- | --- | --- |
| Product conformance | FE-U-P9-01, FE-I-P9-01, FE-E-P9-01 | Direct current row-owned evidence in mapped targets, with current map/registry/guide digests, public route evidence where mapped, pass status, and closure status. | This plan, generated ledgers, retained artifacts alone, target explanations, frontend-only mocks for public routes, visual evidence, accessibility evidence, broad `make check`. |
| Design direction | FE-V-P9-01, FE-A11Y-P9-01 | Direct current row-owned visual or accessibility readiness evidence in mapped targets and fixture/accessibility artifacts where required. | Product-conformance claims, Core 05 claims, fixture registry `current` status alone, screenshots alone, preflight smoke without row-accounting authority. |
| Support evidence | Selectors, helpers, shared harness utilities, import-boundary checks, typecheck, linting. | Support targets and drift checks when shared surfaces change. | Product row closure unless current mapped product evidence also exists. |
| Claim publication | None current. | Explicit Core 05 claim-bearing metadata and required Core 05 evidence, if introduced by a future owner decision. | FE-P9 implementation, visual readiness, accessibility readiness, retained artifacts, broad checks, fixture status, or plan text. |

Evidence classes must never be collapsed. If product, design, support, or claim-publication evidence is counted across classes, record a blocker.

## Dependencies And Prerequisites

Before FE-P9 implementation work promotes any row, verify:

1. `tools/frontend_phase_registry.json` still marks FE-P0 through FE-P8 as green or records owner-accepted blockers that do not invalidate FE-P9.
2. `tools/frontend_phase_maps/fe_p9_test_map.json` still contains exactly `FE-U-P9-01`, `FE-I-P9-01`, `FE-E-P9-01`, `FE-V-P9-01`, and `FE-A11Y-P9-01`.
3. The frontend guide Section 4.9 still matches the live FE-P9 row inventory. If it differs, record a blocker instead of reconciling silently.
4. The generated FE-P9 ledger has not been hand-edited and agrees with the owner map after `make phase-ledgers` and drift checks.
5. Visual fixture ownership for `FE-V-P9-01` is explicit before visual readiness is claimed.
6. Accessibility target ownership remains `browser-e2e-a11y-preflight` unless the live FE-P9 map changes through owner input.
7. Generated protocol consumption uses `packages/protocol-ts` facade exports and no generated protocol roots are hand-edited.
8. Selector/test-id additions route through authored `packages/ui-contracts` inputs and preserve stable ID encoding.
9. Browser product evidence uses public `/api/v1/` routes and public envelopes.
10. Authorization denial evidence uses current server-managed sessions and action-time authorization derivation.
11. Focus-return evidence uses stable row/cell anchors, not vendor coordinates.
12. Direct `react-data-grid` imports remain confined to `/packages/grid-adapter`.

## Shared Harness Analysis

FE-P9 touches shared harnesses because Inspector and Row-Local Actions span unit state modeling, public-route integration, server-backed browser behavior, stateful authorization/focus behavior, visual readiness, and accessibility readiness.

Harness mechanics:

1. `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P9` is the phase discovery source. It reports FE-P9 as planned and not executable as a whole phase.
2. `make explain-target TARGET=<target> DETAIL=summary` must be run before depending on target behavior.
3. Product row accounting must be produced by the mapped target for the implemented row. Old retained artifacts are diagnostic only.
4. Scenario title expectations from the FE-P9 map are required for rows that declare them, but title text alone is never evidence.
5. `browser-e2e-a11y-preflight` currently maps to `FE-A11Y-P9-01`, but the retained preflight summary remains smoke until row accounting and map authority permit closure.
6. Visual fixture lifecycle mechanics are owned by the harness and visual golden guide. Fixture registry status cannot close `FE-V-P9-01` without current row-owned visual accounting.
7. Generated ledgers must be regenerated through `make phase-ledgers` only after authored phase-map or registry owner input changes. The ledger itself must not be hand-edited.
8. Drift targets detect stale generated outputs and schedules; they do not create row evidence.
9. `make check` may be required for end-of-run repository health, but it does not close FE-P9 rows without current row-owned evidence.
10. `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` is required when retaining a successful full-check run for handoff.

Public-route product rows:

1. FE-I-P9-01 requires public route evidence for history and rollback preview/action when behavior crosses the frontend/server boundary.
2. FE-E-P9-01 requires browser evidence through public routes for Details, Relationships, Evidence, History, rollback, destructive action authorization, public error envelopes, and focus return.
3. Frontend mocks can support unit state and selectors, but cannot close browser-facing public-route product conformance.

## Public Interfaces And Deliverables

Expected FE-P9 implementation deliverables for future product work:

1. Row-local inspector state keyed by `record_id`, with stable section/tab state for Details, Relationships, Evidence, and History.
2. Details rendering and row-local edit affordances that use `record_id`, `field_key`, and `row_version`.
3. Relationship panel rendering that uses stable relationship item refs and stable selector builders.
4. Evidence panel rendering that uses stable evidence IDs and existing public evidence affordance contracts without exposing raw storage details.
5. History timeline rendering from public `/api/v1/records/{record_id}/history` envelopes, preserving retained history and stable `history_item_ref` selectors.
6. Rollback preview/action through public `/api/v1/records/{record_id}/rollback` with `base_row_version`, `client_txn_id`, public success envelopes, and public error envelopes.
7. Destructive-action confirmation and errors for soft delete and restore through public `/api/v1/records/{record_id}` and `/api/v1/records/{record_id}/restore` contracts, with current authorization re-checks.
8. Focus return to the originating row/cell or a stable fallback when the originating row is deleted, restored, filtered, or refreshed.
9. Stable selectors and test IDs for inspector sections, relationship links, evidence controls, history items, rollback actions, destructive confirmations, public errors, and focus anchors.
10. Row-owned unit, integration, browser, visual, and accessibility scenarios using exact FE-P9 row IDs and live map scenario titles.
11. FE-P10 handoff notes that preserve row evidence status, direct artifact roots, blockers, strict non-claims, and retained-run finalization outcome.

Generated protocol files, generated ledgers, visual goldens, phase maps, registries, and lockfiles are not deliverables for this plan-only task. If implementation later requires generated or contract-surface changes, update authored owner inputs and run the repository-supported generator targets.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers |
| --- | --- | --- | --- |
| [x] | 1. Readiness and source alignment | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P9`; source digest inspection; retained artifact inspection; `git status --short` | FE-P9 remains planned and blocked; no product rows are closed. |
| [ ] | 2. Unit inspector state for FE-U-P9-01 | `make frontend-unit`; `make frontend-typecheck` | Blocked until row-owned unit scenarios and row accounting exist. |
| [ ] | 3. History and rollback integration for FE-I-P9-01 | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make frontend-import-boundary-check` | Blocked until public route contract evidence and row accounting exist. |
| [ ] | 4. Browser E2E for inspector, evidence, history, rollback, and destructive actions for FE-E-P9-01 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Blocked until current public-route browser evidence and focus-return evidence exist. |
| [ ] | 5. Visual readiness for FE-V-P9-01 | `make browser-e2e-visual` | Blocked because no FE-P9-owned visual fixture registry entry was inspected. |
| [ ] | 6. Accessibility readiness for FE-A11Y-P9-01 | `make browser-e2e-a11y-preflight` under current map, or the mapped target if owner input changes | Blocked because preflight smoke is not row closure and no row accounting exists. |
| [ ] | 7. Drift, full check, finalization, and FE-P10 handoff | Drift targets, `make check`, `make agent-finalize RESULTS_DIR=<retained-check-run>` | Blocked until all FE-P9 rows close under current row-owned evidence. |

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness And Source Alignment

Objective: create a planning baseline without implementing FE-P9 behavior.

Owned rows: none.

Actions:

1. Inspect frontend guide Section 4.9, FE-P8 handoff, FE-P9 registry, FE-P9 map, generated FE-P9 ledger, visual fixture registry, harness guide, frontend guide, dev guide, UI/UX guide, visual golden guide, relevant retained artifacts, and target explanations.
2. Record FE-P9 current state as planned, no rows implemented, all rows blocked.
3. Record source limits and blockers for missing visual fixture ownership, absent FE-P9 row-owned evidence, preflight-only accessibility smoke, and target explanation coverage mismatches.
4. Do not implement product code.
5. Do not edit generated protocol files, generated ledgers, visual goldens, phase maps, registry files, lockfiles, or generated artifacts.

Exit condition: this plan exists at repository root, lists inspected sources and blockers, and post-write diff verification shows only intended authored planning text.

### Sprint 2: Unit Inspector State

Objective: close `FE-U-P9-01` only after deterministic unit evidence proves inspector state identity.

Owned row: `FE-U-P9-01`.

Implementation requirements:

1. Model inspector selection by `record_id`.
2. Preserve selected row and inspector section/tab state across row refresh and reordering.
3. Keep Details, Relationships, Evidence, and History anchors stable under row refresh.
4. Keep relationship and evidence selectors based on stable IDs, not labels or order.
5. Prove missing, stale, deleted, or inaccessible selected rows fail closed and return focus appropriately.

Validation:

```sh
make frontend-unit
make frontend-typecheck
```

Closure requires current `frontend-unit/frontend-row-accounting.json` with row `FE-U-P9-01`, current registry/guide/map digests, pass status, `closure_status` `closed`, and no row blocker.

### Sprint 3: History And Rollback Integration

Objective: close `FE-I-P9-01` only after history and rollback preview/action use public route contracts and public envelopes.

Owned row: `FE-I-P9-01`.

Implementation requirements:

1. Fetch row history through public route contracts for the selected `record_id`.
2. Preserve retained history identity and stable `history_item_ref` selectors.
3. Build rollback preview/action targets from stable history identity.
4. Submit rollback through public route contracts with `base_row_version` and `client_txn_id`.
5. Render public error envelopes for conflict, stale row, denied authorization, invalid target, missing history, or unavailable rollback.

Validation:

```sh
make frontend-unit
make browser-e2e-webserver-backed
make frontend-import-boundary-check
```

Closure requires current row accounting from mapped targets with row `FE-I-P9-01`, current digests, pass status, exact scenario title, public route evidence, and no row blocker.

### Sprint 4: Browser E2E For Inspector, Evidence, History, Rollback, And Destructive Actions

Objective: close `FE-E-P9-01` only after browser evidence proves row-local inspector behavior through public routes.

Owned row: `FE-E-P9-01`.

Implementation requirements:

1. Verify Details, Relationships, Evidence, and History in the row-local inspector through public browser flows.
2. Verify rollback preview/action through public browser routes and retained history.
3. Verify soft delete and restore confirmation/error behavior through public browser routes.
4. Re-check current authorization at destructive-action time and render public error envelopes.
5. Prove focus return to the originating row/cell or a stable fallback after action, failure, refresh, delete, restore, and rollback.
6. Ensure row actions do not target group rows, presentation rows, draft rows without stable IDs, or stale visible positions.

Validation:

```sh
make browser-e2e-webserver-backed
make browser-e2e-stateful
```

Closure requires current row accounting from both mapped targets with row `FE-E-P9-01`, exact scenario title, public route traces, current digests, pass status, `closure_status` `closed`, and no row blocker.

### Sprint 5: Visual Readiness

Objective: close `FE-V-P9-01` as design-direction evidence only after visual fixture ownership and current visual row accounting exist.

Owned row: `FE-V-P9-01`.

Implementation requirements:

1. Add or confirm authored visual fixture ownership for FE-P9 before claiming visual readiness.
2. Capture inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures.
3. Keep visual readiness design-direction only.
4. Follow `docs/guides/cartulary_visual_golden_maintenance.md` for any golden changes.

Validation:

```sh
make browser-e2e-visual
```

Closure requires current visual row accounting for `FE-V-P9-01`, exact scenario title, current guide/registry/map digests, fixture ownership, pass status, design-direction closure status, and no stale fixture blocker.

### Sprint 6: Accessibility Readiness

Objective: close `FE-A11Y-P9-01` as design-direction evidence only after map-authorized accessibility evidence exists.

Owned row: `FE-A11Y-P9-01`.

Implementation requirements:

1. Verify inspector tabs or sections, relationship links, evidence controls, history controls, rollback actions, destructive actions, and errors are keyboard reachable and announced.
2. Verify visible focus, accessible names, ARIA state where applicable, and non-color-only distinctions.
3. Keep accessibility readiness design-direction only.
4. Treat preflight smoke as support unless the current map and row accounting permit closure.

Validation under the current map:

```sh
make browser-e2e-a11y-preflight
```

If the FE-P9 map changes to a full accessibility target, use that mapped target and record the owner input. Closure requires current row-owned accessibility evidence with row `FE-A11Y-P9-01`, exact scenario title, current digests, pass status, and no row blocker.

### Sprint 7: Drift, Full Check, Finalization, And FE-P10 Handoff

Objective: validate generated and harness state, run required gates for the final implementation posture, close blockers, and produce FE-P10 handoff.

Owned rows: all FE-P9 rows after their owning sprints close.

Validation:

```sh
make phase-ledgers
make phase-ledger-drift
make phase-schedule-drift
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make frontend-typecheck
make frontend-import-boundary-check
make check
make agent-finalize RESULTS_DIR=<successful-full-check-run-root>
```

Exit condition: all five FE-P9 rows have current row-owned evidence; FE-P9 registry owner inputs are promoted by the appropriate workflow; generated ledgers agree with owner maps; drift checks pass; full check passes; retained-run finalization passes; FE-P10 handoff lists evidence roots, blockers, and strict non-claims.

## Validation Commands

### Explanation Commands

These commands were run for plan creation and existed in the public Make surface or explainable target surface:

```sh
make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P9
make explain-target TARGET=frontend-unit DETAIL=summary
make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary
make explain-target TARGET=browser-e2e-stateful DETAIL=summary
make explain-target TARGET=browser-e2e-visual DETAIL=summary
make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary
make explain-target TARGET=browser-e2e-a11y DETAIL=summary
make explain-target TARGET=browser-e2e-support DETAIL=summary
make explain-target TARGET=phase-ledgers DETAIL=summary
make explain-target TARGET=phase-ledger-drift DETAIL=summary
make explain-target TARGET=phase-schedule-drift DETAIL=summary
make explain-target TARGET=generate-drift DETAIL=summary
make explain-target TARGET=generated-artifact-policy-check DETAIL=summary
make explain-target TARGET=json-shape-check DETAIL=summary
make explain-target TARGET=phase-map-check DETAIL=summary
make explain-target TARGET=check DETAIL=summary
make explain-target TARGET=agent-finalize DETAIL=summary
make explain-target TARGET=frontend-typecheck DETAIL=summary
make explain-target TARGET=frontend-import-boundary-check DETAIL=summary
make explain-target TARGET=lint-markdown DETAIL=summary
make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P9
make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P9
make target-plan-json
make explain-run RESULTS_DIR=.cartulary/test-results/20260612T124559Z-p85431
make explain-run RESULTS_DIR=.cartulary/test-results/20260612T125604Z-p81938
```

### Execution Commands

The following commands are canonical future validation commands. This plan does not claim they passed for FE-P9 unless current run evidence exists.

```sh
make phase-ledgers
make frontend-unit
make browser-e2e-webserver-backed
make browser-e2e-stateful
make browser-e2e-visual
make browser-e2e-a11y-preflight
make frontend-typecheck
make frontend-import-boundary-check
make phase-ledger-drift
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make check
make agent-finalize RESULTS_DIR=<retained-check-run>
make lint-markdown
git diff --check
```

`make browser-e2e-a11y` exists but is not the current FE-P9 mapped accessibility target. If the FE-P9 map changes, use the current mapped target and record the owner input.

`make agent-finalize` without `RESULTS_DIR` may be run only when retained-run maintenance is intentionally skipped. When a retained successful full-check run is used, `RESULTS_DIR=<successful-full-check-run-root>` is required.

## Evidence Requirements

Every FE-P9 evidence payload must include:

1. Row-accounting schema ID and FE-P9 row ID.
2. Target identity and command ID.
3. Accounting scope and phase namespace.
4. Evidence class.
5. Scenario, unit, route, or artifact identity.
6. Current FE-P9 map digest.
7. Current registry digest.
8. Current guide digest.
9. Pass/fail status.
10. Closure status.
11. Artifact root.
12. Exact blocker when missing or stale.
13. Per-row artifact family.

| Row ID | Required artifact family | Minimum evidence payload |
| --- | --- | --- |
| FE-U-P9-01 | `frontend-unit` row accounting, target summary, phase summary, unit output | Row `FE-U-P9-01`; command `cartulary.harness.command.frontend_unit.v1`; inspector state anchored by `record_id`; Details, Relationships, Evidence, and History anchors survive refresh; current digests; pass status; closure status. |
| FE-I-P9-01 | `frontend-unit` and `browser-e2e-webserver-backed` row accounting, public route/browser artifacts | Row `FE-I-P9-01`; exact scenario title; public history and rollback route evidence; retained history preservation; public error envelopes; current digests; pass status; closure status. |
| FE-E-P9-01 | `browser-e2e-webserver-backed` and `browser-e2e-stateful` row accounting, route traces, focus artifacts | Row `FE-E-P9-01`; exact scenario title; Details, Relationships, Evidence, History, rollback, delete, restore, authorization denial, public error envelopes, and focus return through public browser routes; current digests; pass status; closure status. |
| FE-V-P9-01 | `browser-e2e-visual` row accounting, fixture registry references, screenshots/goldens, target summary | Row `FE-V-P9-01`; exact scenario title; fixture ownership; inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures; current digests; pass status; design-direction closure status. |
| FE-A11Y-P9-01 | `browser-e2e-a11y-preflight` row accounting under current map, preflight/accessibility summary, target summary | Row `FE-A11Y-P9-01`; exact scenario title; keyboard reachability and announced state for inspector tabs, relationship links, evidence controls, history controls, rollback, destructive actions, and errors; current digests; pass status; design-direction closure status. |

## Blocker Rules

The following conditions must create blockers:

1. FE-P9 guide/map row inventory mismatch.
2. Unresolved owner refs, missing owner sections, or owner REQ/AC IDs that cannot be found in authoritative sources.
3. Stale guide, registry, phase-map, ledger, or evidence freshness digest.
4. Generated-ledger drift or hand-edited generated ledgers.
5. Retained artifacts older than the current map/registry/guide digests.
6. Any FE-P9 row relying on broad `make check`, target explanations, plan text, fixture status, screenshots, or generated ledgers for closure.
7. Product rows relying only on frontend mocks for public route behavior.
8. Inspector identity keyed by visible label, visible row index, DOM order, CSS coordinates, or grid vendor coordinates.
9. Relationship or evidence controls keyed by mutable labels instead of stable IDs.
10. History or rollback selectors keyed by render order instead of `history_item_ref` or stable action identity.
11. Rollback, delete, or restore without `base_row_version`, current authorization re-checks, public success envelopes, and public error envelopes.
12. Destructive actions without confirmation/error behavior.
13. Focus not returning to the originating row/cell or stable fallback after close, rollback, delete/restore, error, or refresh.
14. Group rows, presentation rows, or draft-only rows receiving mutation-capable inspector actions.
15. Direct `react-data-grid` imports outside `/packages/grid-adapter`.
16. Visual readiness represented as product conformance.
17. Accessibility readiness represented as product conformance.
18. Accessibility preflight used as row closure without map authority and current row accounting.
19. Core 05 publication readiness implied without claim-bearing metadata.
20. `make agent-finalize` used without `RESULTS_DIR` while retaining a full-check run.

Current known blockers:

| Affected row | Affected source | Observed condition | Minimum follow-up |
| --- | --- | --- | --- |
| All FE-P9 | `tools/frontend_phase_registry.json` | FE-P9 is `planned`, `no_rows_implemented`, with `FE-P9-ACTIVATION-BLOCKER-01`. | Add direct row evidence and promote freshness/registry state through owner workflow. |
| FE-U-P9-01 | Live FE-P9 map and retained row accounting | Row is blocked; retained `frontend-unit` row accounting contains no FE-P9 rows. | Add FE-U-P9-01 unit scenarios and rerun `make frontend-unit`. |
| FE-I-P9-01 | Live FE-P9 map and retained row accounting | Row is blocked; retained `frontend-unit` and `browser-e2e-webserver-backed` row accounting contain no FE-P9 rows. | Add row-owned integration/browser evidence and rerun mapped targets. |
| FE-E-P9-01 | Live FE-P9 map and target explanation | Row is blocked; retained browser accounting contains no FE-P9 rows; `browser-e2e-stateful` explanation omits phase9 coverage. | Add row-owned webserver-backed and stateful evidence; rely on current FE-P9 row accounting. |
| FE-V-P9-01 | `tools/frontend_visual_fixture_registry.json` | No FE-P9-owned fixture entry inspected. | Add or confirm fixture ownership, recapture visual evidence, and rerun `make browser-e2e-visual`. |
| FE-A11Y-P9-01 | `.cartulary/test-results/20260612T045159Z-p12967` | Preflight scenario passed, but row remains blocked and target has no `frontend-row-accounting.json`. | Produce current map-authorized row accounting or update owner map through the appropriate workflow. |
| All FE-P9 | Core 05 boundary | No explicit FE-P9 claim-bearing metadata inspected. | Add Core 05-compliant claim metadata only if publication evidence becomes in scope. |

## Strict Non-Claims

This plan does not claim:

1. FE-P9 completion by document creation.
2. Any FE-P9 row closure.
3. Any FE-P9 row closure from generated ledgers.
4. Any FE-P9 row closure from prior phase plans.
5. Any FE-P9 row closure from old retained artifacts.
6. Any FE-P9 row closure from broad `make check`.
7. Any FE-P9 row closure from target explanations.
8. Any FE-P9 row closure from screenshots or fixture status alone.
9. Product conformance from visual evidence.
10. Product conformance from accessibility evidence.
11. Claim-publication readiness or Core 05 readiness.
12. Benchmark readiness.
13. Fixture-sensitive publication readiness.
14. Visual-publication readiness.
15. Coordination workbook surfaces beyond relationships.
16. WebSocket live updates.
17. Inspector behavior keyed by visible row index, DOM order, labels, or vendor coordinates.
18. Rollback readiness without current public route evidence.
19. Evidence readiness without stable evidence IDs and public contract evidence.
20. Destructive action readiness without confirmation, authorization re-check, base row version, and public error envelopes.
21. `react-data-grid` import permission outside `/packages/grid-adapter`.
22. Any behavior not mapped to FE-P9 rows.

## Binary Exit Criteria

| Criterion | Pass condition | Fail condition |
| --- | --- | --- |
| 1. Plan creation | `FRONTEND_PHASE9_IMPLEMENTATION_PLAN.md` exists at repository root, records live inspected inputs, lists blockers, and diff verification shows only intended planning text. | File missing, live claims unanchored, blockers omitted, generated artifacts edited, or post-write checks fail without blocker recording. |
| 2. Product row closure | FE-U-P9-01, FE-I-P9-01, and FE-E-P9-01 each have current mapped product-conformance row accounting with current digests, pass status, exact scenario/unit identity, closure status, and no blockers. | Any product row lacks current row-owned evidence, uses stale digests, remains blocked, or depends only on broad `make check`. |
| 3. Inspector identity | Inspector selection, section/tab state, Details, Relationships, Evidence, and History are anchored by `record_id` and stable IDs across refresh/reorder. | Any inspector action depends on visible row index, DOM order, labels, CSS coordinates, or vendor coordinates. |
| 4. History and rollback | History and rollback preview/action use public route contracts, stable `history_item_ref` identity, retained history, current authorization, public success envelopes, and public error envelopes. | Rollback is unit-only, mock-only, stale-version unsafe, not public-route proven, or not row-owned. |
| 5. Destructive actions | Delete/restore confirmation and errors re-check current authorization, include base row version and client transaction identity, render public envelopes, and preserve focus continuity. | Destructive action skips confirmation/error behavior, lacks current auth re-check, targets stale rows, or loses focus without fallback. |
| 6. Evidence and relationships | Relationship and evidence controls use stable relationship/evidence IDs and do not expose raw storage details or mutable display labels as action identity. | Controls use labels/order, expose raw object-store details, or lack stable selectors. |
| 7. Visual readiness | FE-V-P9-01 has current design-direction row accounting and fixture ownership for inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public errors. | Visual fixture ownership is missing, row accounting is absent/stale, coverage is incomplete, or visual evidence is represented as product conformance. |
| 8. Accessibility readiness | FE-A11Y-P9-01 has current map-authorized accessibility evidence and row accounting for keyboard reachability, focus visibility, names, announcements, and non-color-only states. | Preflight remains smoke without row accounting, controls are not keyboard reachable, announcements are missing, or evidence is promoted to product conformance. |
| 9. Full FE-P9 phase completion | Registry owner workflow promotes FE-P9 from `planned` to active/green, all five rows close under current map authority, ledgers are regenerated by Make, drift checks pass, full check passes, and retained-run finalization passes. | Any row remains blocked, registry/map/ledger freshness fails, generated-ledger drift exists, or retained evidence is stale/missing. |
| 10. FE-P10 handoff readiness | Handoff lists final FE-P9 registry status, row inventory, direct evidence roots, drift/finalization outcomes, strict non-claims, and blockers for any unclosed rows. | Handoff treats FE-P9 plan text or earlier phase evidence as FE-P9 closure, omits blockers, or lacks direct evidence roots. |

Current status at plan creation:

| Criterion | Current status |
| --- | --- |
| 1. Plan creation | In progress for this authored Markdown task. |
| 2. Product row closure | Blocked. No FE-P9 product row-owned evidence exists in retained row accounting. |
| 3. Inspector identity | Blocked until FE-U-P9-01 evidence exists. |
| 4. History and rollback | Blocked until FE-I-P9-01 and FE-E-P9-01 evidence exists. |
| 5. Destructive actions | Blocked until FE-E-P9-01 evidence exists. |
| 6. Evidence and relationships | Blocked until FE-U-P9-01 and FE-E-P9-01 evidence exists. |
| 7. Visual readiness | Blocked. No FE-P9-owned fixture entry was inspected. |
| 8. Accessibility readiness | Blocked. Preflight smoke exists but no row accounting exists. |
| 9. Full FE-P9 phase completion | Blocked. FE-P9 is `planned`, `no_rows_implemented`. |
| 10. FE-P10 handoff readiness | Blocked until FE-P9 rows close or blockers are owner-accepted. |

## FE-P10 Handoff

FE-P10 must receive FE-P9 status as dependency context only. Until FE-P9 is implemented and closed, the FE-P10 handoff is:

| Handoff field | Current FE-P9 value |
| --- | --- |
| Registry status | `planned`. |
| Row rollup | `no_rows_implemented`. |
| Activation blocker | `FE-P9-ACTIVATION-BLOCKER-01` remains active. |
| Product rows | Blocked: `FE-U-P9-01`, `FE-I-P9-01`, `FE-E-P9-01`. |
| Design-direction rows | Blocked: `FE-V-P9-01`, `FE-A11Y-P9-01`. |
| Direct FE-P9 evidence roots | None. Retained artifacts inspected contain no FE-P9 row-owned closure evidence. |
| Visual fixture status | No FE-P9-owned fixture entry inspected. |
| Accessibility status | Preflight smoke exists at `.cartulary/test-results/20260612T045159Z-p12967`, but it does not close `FE-A11Y-P9-01`. |
| Full check context | `.cartulary/test-results/20260612T124559Z-p85431` passed, but it does not close FE-P9 rows. |
| Finalization context | `.cartulary/test-results/20260612T125604Z-p81938` passed for retained run `.cartulary/test-results/20260612T124559Z-p85431`, but it does not close FE-P9 rows. |
| Strict non-claims | No visual/product promotion, no accessibility/product promotion, no Core 05 publication, no benchmark claim, no fixture-sensitive publication claim, no visual-publication claim, no WebSocket live-update claim, and no FE-P10 implementation claim. |
| Minimum FE-P10 handoff requirement after FE-P9 implementation | Record final FE-P9 registry state, direct row evidence roots, final digest set, drift/finalization results, visual/accessibility design-only boundaries, unresolved blockers, and strict non-claims. |

FE-P10 must use its own owner map, row accounting, and retained-run finalization. FE-P9 plan text, generated ledgers, and retained health artifacts must not be imported as FE-P10 closure evidence.
