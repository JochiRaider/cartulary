# Frontend Phase 8 Implementation Plan

## Summary

FE-P8 covers saved views, sorting, filtering, grouping, layout state, active chips, startup and default surface UI, group rows, and query-control persistence for the Cartulary frontend. The live repository status at plan creation is `planned` with `row_rollup_state` `no_rows_implemented`; the phase is blocked by `FE-P8-ACTIVATION-BLOCKER-01` in `tools/frontend_phase_registry.json`.

This plan does not claim FE-P8 completion. Current product-conformance closure is blocked because the live FE-P8 map has all six rows at `claim_status` `blocked`, the latest retained full-check row-accounting artifacts inspected from `.cartulary/test-results/20260611T141819Z-p21786` contain no FE-P8 rows, and retained visual/accessibility artifacts either use stale digests or remain design-direction preflight only. Existing FE-P8-adjacent implementation and test surfaces exist under `apps/web`, `packages/grid-adapter`, `packages/view-contracts`, `packages/protocol-ts`, `packages/ui-contracts`, and `packages/test-utils`, but those files are candidate implementation context until current row-owned FE-P8 evidence closes through the live map.

FE-P8 must be treated as planned and blocked until direct current evidence, row accounting, freshness validation, registry promotion, and ledger regeneration agree. Broad `make check`, generated ledgers, retained old artifacts, visual snapshot filenames, fixture-registry status, and support-only checks must not close FE-P8 rows by themselves.

## Authority Model

The authority order for this plan is:

| Order | Source | Use in FE-P8 |
| --- | --- | --- |
| 1 | Adopted Cartulary NLSpecs | Product and harness authority only inside their stated scope. `docs/testing-harness-nlspec.md` owns harness mechanics only. |
| 2 | Core 00 through Core 04 under `docs/spec/` | Product-conformance behavior authority. |
| 3 | Core 05 | Inactive for FE-P8 unless explicit claim-bearing timed, benchmark, fixture-sensitive, visual, or publication metadata exists and satisfies Core 05. |
| 4 | `docs/domain.md` | Vocabulary and concept-boundary interpretation only. |
| 5 | `docs/guides/cartulary_frontend_implementation_testing_guide.md` | FE-P8 planning mechanics, row mapping, evidence classes, and row-owner mapping. |
| 6 | `docs/guides/cartulary_implementation_testing_guide.md` | Sequencing, test-row discipline, shared harnesses, completion rules, and coverage-ledger shape. |
| 7 | `docs/guides/cartulary-dev-guide.md` | Frontend package boundaries, generated-artifact policy, Make targets, workspace shape, and implementation baseline. |
| 8 | UI, design, and visual guides | Design-direction, visual, and accessibility readiness only. |
| 9 | Previous frontend implementation plans | Examples and handoff context only. |
| 10 | Research reports | Rationale only. |

Core 00 through Core 04 own product conformance. Core 05 must remain inactive for FE-P8 unless explicit claim-bearing metadata exists. The testing harness NLSpec owns command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, verification gates, registry/map/ledger mechanics, row accounting, freshness, and target explanations; it must not redefine product behavior. The frontend implementation testing guide owns planning and verification mechanics only. Generated ledgers are downstream artifacts and must not be hand-edited or treated as row-closure authority. Visual and accessibility evidence remain design-direction evidence unless owner mappings explicitly promote a narrower claim. Conflicts between guide text, live maps, generated ledgers, registries, retained artifacts, code, or command output must be recorded as blockers instead of silently reconciled.

## Current Repo Status

Live inspection occurred from repository root `/home/jochi/code/cartulary` on 2026-06-11. The worktree was clean before this plan file was created.

### Inspected Inputs

| Source | Live status at inspection | Digest or command anchor |
| --- | --- | --- |
| `tools/frontend_phase_registry.json` | Present. FE-P8 is `planned`, `no_rows_implemented`, blocked by `FE-P8-ACTIVATION-BLOCKER-01`. | File SHA256 `46d57e203433102aa9fe6ffbe58ce83338890306de417decd0b56b2694094bcc`. FE-P8 manifest digest `c2de1d13c71bfe08cac8c2ffc4b33c68b45fb697d6a34df2da791752e621d4ce`; FE-P8 ledger digest `d18408e2b4dec27fe959f4338077c61cc5473e83f99eb5e1d6c2612404fbe8e1`; FE-P8 evidence freshness digest `d93df1d73b4d31cbbb926dc939e9c45dfd7f18684c682ea904bfb5dc49faa39a`. |
| `tools/frontend_phase_maps/fe_p8_test_map.json` | Present. Six FE-P8 rows, all `claim_status` `blocked`. | SHA256 `c2de1d13c71bfe08cac8c2ffc4b33c68b45fb697d6a34df2da791752e621d4ce`. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` | Present. Generated downstream ledger reports FE-P8 `planned`, `no_rows_implemented`, six blocked rows. | SHA256 `d18408e2b4dec27fe959f4338077c61cc5473e83f99eb5e1d6c2612404fbe8e1`. |
| `tools/frontend_visual_fixture_registry.json` | Present. FE-P8 owner fixtures are current for grouped result, group/tree row, and empty successful query; registry status alone cannot close FE-V-P8-01. | SHA256 `293c7f3b5211ee6a1ef5461f83d4c119240d16d2837eecb358b1e836a067ab3d`. |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | Present. FE-P8 guide scope matches saved views, sorting, filtering, grouping, layout state, active chips, startup/default-surface UI, group rows, and query-control persistence. | SHA256 `3750847aac8805eb59a1f93430de24a036665241aa7ad3eb6dfb13dcb349b0b5`. |
| `docs/testing-harness-nlspec.md` | Present. Harness mechanics only; frontend registry/maps/ledgers/row-accounting are implementation-readiness artifacts and do not define product behavior. | SHA256 `92cfaebb5bcac06565f0cc6b246edfb3446fa2d3ae1e8662f7baf3a1794c7064`. |
| `docs/opentelemetry-instrumentation-nlspec.md` | Present. Telemetry-only NLSpec; no FE-P8 product behavior authority. | SHA256 `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451`. |
| Core 00 through Core 05 | Present under `docs/spec/`. Core 00 through Core 04 own product behavior; Core 05 remains publication-only. | Core 00 `e3b2e5e9ed4f47d29694612d571f3255437a9a1acbceb31fc38d9229756a682f`; Core 01 `0aa290e6dc4c92e68470a5da35372df8461435a34069144494a473601c62efaa`; Core 02 `bb92665e26804b8c465d961fdef39b78e3f07c389a26a6478e9a210ce393d3fa`; Core 03 `fb561f66e61cf75e777a8c1c4d618d1064ca3b36e8d02435e85131ce631f5b10`; Core 04 `ab4d03966850879625141165d7902f108cfe989914e2b01ed42e2ff7968f6da1`; Core 05 `ee2f572430b75b41ccd20d4dede9c72251b3a4432db2ccf525bec9415da7ef89`. |
| `docs/domain.md` | Present. Vocabulary and concept-boundary reference only. | SHA256 `c461f865e5a0524865e691661f9049ceb9031226124aec60e4014b00847d0e21`. |
| `docs/guides/cartulary_implementation_testing_guide.md` | Present. Shared row/test/harness discipline and browser helper mechanics. | SHA256 `0f37b87228685269cef75e299be4d03a54ad4ee5a51af91ee4f74e3e25d1ae3a`. |
| `docs/guides/cartulary-dev-guide.md` | Present. Frontend package and generated artifact boundaries. | SHA256 `0c35b7c6bf206e22a2f2ae0b4a00a14dd16d43dfefbdbe65b4887f17b7c4d83d`. |
| `docs/guides/cartulary-ui-ux-design-guide.md` | Present. Design direction only. | SHA256 `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26`. |
| `docs/design.md` | Present. Design token and visual/accessibility direction only. | SHA256 `728da652a0c8d233145c980bde1f32aad66957eea9f9bd32081ed73a13a038bf`. |
| `docs/guides/cartulary_visual_golden_maintenance.md` | Present. Visual golden maintenance only; fixture status cannot close product rows. | SHA256 `21faf12489ef6a9ee93e10197ad8197ceb1a1da8a7303a5fc693ac71db5ce5c1`. |
| `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` through `FRONTEND_PHASE7_IMPLEMENTATION_PLAN.md` | Phase 5, 6, and 7 exist at repository root. Phase 0 through Phase 4 exist under `docs/archive/`. | `rg --files` inspection. |

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
| FE-P8 | `planned` | `no_rows_implemented` | `FE-P8-ACTIVATION-BLOCKER-01`: frontend phase is not active until direct row evidence and freshness validation are promoted together. |

### FE-P8 Map Inventory

| Row ID | Layer | Evidence class | Live targets | Current claim status | Current blocker summary |
| --- | --- | --- | --- | --- | --- |
| FE-U-P8-01 | unit | product_conformance | `frontend-unit` | `blocked` | `FE-U-P8-01-BLOCKER-01`, `frontend_phase_row_not_implemented`. |
| FE-I-P8-01 | integration | product_conformance | `frontend-unit`, `browser-e2e-webserver-backed` | `blocked` | `FE-I-P8-01-BLOCKER-01`, `frontend_phase_row_not_implemented`. |
| FE-B-P8-01 | browser_integration | product_conformance | `browser-e2e-webserver-backed`, `browser-e2e-support` | `blocked` | `FE-B-P8-01-BLOCKER-01`, `frontend_phase_row_not_implemented`. |
| FE-E-P8-01 | e2e | product_conformance | `browser-e2e-webserver-backed`, `browser-e2e-stateful` | `blocked` | `FE-E-P8-01-BLOCKER-01`, `frontend_phase_row_not_implemented`. |
| FE-V-P8-01 | visual | design_direction | `browser-e2e-visual` | `blocked` | `FE-V-P8-01-BLOCKER-01`, visual fixture not recaptured for frontend row. |
| FE-A11Y-P8-01 | accessibility | design_direction | `browser-e2e-a11y-preflight` | `blocked` | `FE-A11Y-P8-01-BLOCKER-01`, accessibility preflight not recaptured for frontend row. |

The live map marks every FE-P8 target with `frontend_row_accounting_required: true` and `required_for_closure: false`. Product rows must not close until the row map is promoted to closure-eligible status and current row-owned evidence exists. Design-direction rows must not be promoted into product conformance.

### Ledger Status

`docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` is generated downstream of the live map. It records FE-P8 as `planned`, `no_rows_implemented`, and blocked for all six rows. It must not be hand-edited and must not be used as independent closure evidence.

### Fixture Registry Status

| Fixture ID | FE-P8 row ownership | Status | Current limitation |
| --- | --- | --- | --- |
| FE-VFIX-06 | `FE-V-P8-01` | `current` | Covers grouped result. It cannot close FE-V-P8-01 without current row-owned visual accounting and map authority. |
| FE-VFIX-13 | `FE-V-P8-01` with FE-P3 and FE-P10 owners | `current` | Covers tree/group row. It is shared and cannot close FE-V-P8-01 by fixture status alone. |
| FE-VFIX-15 | `FE-V-P8-01` with FE-P3 and FE-P4 owners | `current` | Covers empty successful query. It is shared and cannot close FE-V-P8-01 by fixture status alone. |

No inspected fixture-registry entry with `owner_row_ids` containing `FE-V-P8-01` was found for saved-view selector, active chips, or default/startup state indicator as separate named fixture IDs. FE-V-P8-01 closure must therefore require either row-owned visual accounting that captures the full row scenario or a registry/map update that names the missing fixture ownership.

### Relevant App And Package Status

| Path | Inspected FE-P8 relevance | Current status |
| --- | --- | --- |
| `apps/web/src/workbookQuery.ts` | Query-state model and compilation for sort, filters, `group_by`, and active chip labeling. | Candidate FE-P8 implementation surface. Closure still blocked. |
| `apps/web/src/WorkbookGridControls.tsx` | Sort, group, filter, active chip UI controls. | Candidate FE-P8 implementation surface. Closure still blocked. |
| `apps/web/src/WorkbookShell.tsx` | Saved-view selector, startup selection, query execution, view bar integration, saved-view route reads, and public query route calls. | Candidate FE-P8 implementation surface. Closure still blocked. |
| `apps/web/src/workbookStartup.ts` | Startup/default sheet reference normalization and URL startup query handling. | Candidate FE-P8 implementation surface. Closure still blocked. |
| `apps/web/src/workbookSurfaceRegistry.ts` | Surface registry and view contract grouping. | Candidate FE-P8 implementation surface. Closure still blocked. |
| `apps/web/src/WorkbookShell.phase8.query.test.tsx` | FE-P8-adjacent unit tests for query controls and group rows. | Test names use legacy `U-8`/`E-8` forms, not live FE row IDs. No FE-P8 row closure. |
| `apps/web/e2e/phase8.workbook.spec.ts` | FE-P8-adjacent browser tests for saved views, startup fallback, query controls, full-text/prefix queries, and public routes. | Existing tests are candidate context. Live latest row-accounting artifacts inspected did not contain FE-P8 product rows. |
| `packages/grid-adapter/src/core.ts` | Grid row identity, group presentation rows, paste/edit target blocking for presentation rows. | Candidate support for non-writable group rows. Closure still blocked. |
| `packages/grid-adapter/src/index.tsx` | Owns the direct `react-data-grid` import. | `rg "react-data-grid"` found direct runtime import in `packages/grid-adapter`; app code must continue to use adapter APIs. |
| `packages/view-contracts/src/index.ts` | Contract-derived view schema, sortable/filterable/groupable field maps, default sort, and synthetic predicates. | Candidate contract boundary for `field_key` authority. Closure still blocked. |
| `packages/protocol-ts/src/index.ts` and `packages/protocol-ts/src/generated/index.ts` | Generated protocol contract exports. | Generated roots must not be hand-edited. |
| `packages/ui-contracts/src/index.ts` | Stable selectors/test-id builders for grid shell, sort headers, filter chips, grouping control, group rows, saved-view selector, saved-view options, cells, and workbook readiness. | Candidate selector authority. Closure still blocked. |
| `packages/test-utils/src/index.ts` | Browser helpers for sort, filter chips, grouping, group expand/collapse, paste/fill-down, anchors, and scrolling. | Candidate helper surface. Closure still blocked. |

### Retained Evidence Status

| Artifact root | Inspected finding | FE-P8 use |
| --- | --- | --- |
| `.cartulary/test-results/20260611T141819Z-p21786/frontend-unit/frontend-row-accounting.json` | `target_status` `pass`, current guide digest and registry digest, but `row_results` contains no FE-P8 rows. | Must not close FE-U-P8-01 or FE-I-P8-01. |
| `.cartulary/test-results/20260611T141819Z-p21786/browser-e2e-webserver-backed/frontend-row-accounting.json` | `target_status` `pass`, current guide digest and registry digest, but `row_results` contains no FE-P8 rows. | Must not close FE-I-P8-01, FE-B-P8-01, or FE-E-P8-01. |
| `.cartulary/test-results/20260611T141819Z-p21786/browser-e2e-stateful/frontend-row-accounting.json` | `target_status` `pass`, current guide digest and registry digest, but `row_results` contains no FE-P8 rows. | Must not close FE-E-P8-01. |
| `.cartulary/test-results/20260610T191949Z-p14292/browser-e2e-visual/frontend-row-accounting.json` | `target_status` `pass`, but guide digest `7bfa510b010f7696ab79d174ea23489e28045b7924743d16d9990e2ef8e22339` and registry digest `4ad7f3f5fa9dd7ecd4865375e2c47940e089d17b5d4bf00297621842948782e8` are stale relative to current inspected digests; row results omit FE-V-P8-01. | Must not close FE-V-P8-01. |
| `.cartulary/test-results/20260610T171433Z-p14284/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json` | `status` `pass` for FE-A11Y-P8-01 scenario, but phase row still records `claim_status` `blocked`, `required_for_closure` `false`, and the target root has no `frontend-row-accounting.json`. | May inform readiness; must not close FE-A11Y-P8-01. |
| `.cartulary/test-results/20260611T141819Z-p21786/check/target-summary.json` | `status` `pass`. It is a broad retained check run. | Must not close FE-P8 rows because current FE-P8 row-owned accounting is absent. |

### Target Explanation Status

All required explanation commands were run and exited with status 0 before plan creation.

| Command | Result | Live finding |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | FE-P8 is `planned`; all six rows are blocked; planned frontend phases are explainable but not executable as a whole. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | pass | Public fast target; phase coverage includes `phase8`; latest artifact `.cartulary/test-results/20260611T141819Z-p21786/frontend-unit/frontend-unit/phase-summary.json`. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | pass | Public full-gate target; phase coverage includes `phase8`; latest artifact `.cartulary/test-results/20260611T141819Z-p21786/browser-e2e-webserver-backed/target-summary.json`. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | pass | Public full-gate target; phase coverage lists `phase1,phase6`, while the live FE-P8 map names it for FE-E-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-support DETAIL=summary` | pass | `internal_helper`; phase coverage lists `phase1,phase2,phase3`, while the live FE-P8 map names it for FE-B-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | pass | Public full-gate target; phase coverage lists `phase3,phase4,phase5,phase6`, while the live FE-P8 map names it for FE-V-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` | pass | Public full-gate target; phase coverage `none`; not the live mapped FE-P8 accessibility target. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | pass | Public full-gate target; phase coverage `none`; live FE-P8 map names it for FE-A11Y-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=check DETAIL=summary` | pass | Public full-gate target; latest artifact `.cartulary/test-results/20260611T141819Z-p21786/check/target-summary.json`. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | pass | Public helper target requiring `RESULTS_DIR` when retained-run maintenance is intended. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | pass | Public fast target. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | pass | Public fast target. |

## Source Limits

1. FE-P8 is not active. `tools/frontend_phase_registry.json` marks FE-P8 as `planned` with `row_rollup_state` `no_rows_implemented` and activation blocker `FE-P8-ACTIVATION-BLOCKER-01`.
2. All six FE-P8 rows in `tools/frontend_phase_maps/fe_p8_test_map.json` are blocked. The plan must not mark any row complete.
3. The live FE-P8 guide section uses broader REQ ranges than the exact REQ ID lists in the live map. Closure must use the live map exact IDs. TODO: `FE-P8-SOURCE-DRIFT-01` reconcile guide range language and map exact IDs before expanding any FE-P8 row claim.
4. The generated FE-P8 ledger is downstream. It records the current planned/blocked state but cannot close rows.
5. The latest current-digest product row-accounting artifacts inspected under `.cartulary/test-results/20260611T141819Z-p21786` contain no FE-P8 rows. TODO: `FE-P8-EVIDENCE-01` produce current FE-P8 row-owned accounting in the mapped targets.
6. The retained visual row-accounting artifact from `.cartulary/test-results/20260610T191949Z-p14292` has stale guide and registry digests and omits FE-V-P8-01. TODO: `FE-P8-VISUAL-01` recapture FE-V-P8-01 against current registry, guide, and map digests.
7. The retained accessibility preflight summary from `.cartulary/test-results/20260610T171433Z-p14284` has a passing FE-A11Y-P8-01 scenario but still records the row as blocked and has no root `frontend-row-accounting.json`. TODO: `FE-P8-A11Y-01` produce accepted current row accounting or map-authorized accessibility evidence.
8. `make explain-target` reports FE-P8 coverage for `frontend-unit` and `browser-e2e-webserver-backed`; it does not currently report FE-P8 coverage for `browser-e2e-stateful`, `browser-e2e-support`, `browser-e2e-visual`, or `browser-e2e-a11y-preflight`, even though the live FE-P8 map names those targets. TODO: `FE-P8-TARGET-SURFACE-01` confirm target coverage after FE-P8 activation or update the task surface if the live map and target explanation remain inconsistent.
9. The fixture registry has current FE-P8 ownership for grouped result, group/tree row, and empty successful query. It does not, by itself, prove saved-view selector, active chips, or default/startup indicator capture for FE-V-P8-01. TODO: `FE-P8-VISUAL-02` map or recapture the missing visual states under FE-V-P8-01.
10. Existing FE-P8-adjacent tests in `apps/web/src/WorkbookShell.phase8.query.test.tsx` and `apps/web/e2e/phase8.workbook.spec.ts` are implementation context only until their scenarios are row-owned by the live FE-P8 map and current accounting artifacts.
11. The plan cannot claim public route behavior from frontend mocks. Saved-view persistence, startup/default persistence, and query replay must be proven through `/api/v1/` route evidence in mapped browser/E2E targets.
12. The plan cannot claim Core 05 publication, benchmark, visual-publication, fixture-sensitive publication, or claim-publication readiness. No explicit claim-bearing FE-P8 metadata was inspected.

## FE-P7 Handoff Inputs

`FRONTEND_PHASE7_IMPLEMENTATION_PLAN.md` is handoff context only. It must not be imported as FE-P8 evidence.

| FE-P7 handoff item | Recorded status in FE-P7 plan | FE-P8 use |
| --- | --- | --- |
| Registry status | FE-P7 `active`, `active_green`; all six FE-P7 rows implemented. | Dependency context only. |
| Row inventory | FE-U-P7-01, FE-U-P7-02, FE-I-P7-01, FE-E-P7-01, FE-V-P7-01, FE-A11Y-P7-01 closed in FE-P7 handoff. | Must not close FE-P8 rows. |
| FE-P7 product evidence roots | FE-P7 plan lists current row-accounting roots for `frontend-unit`, `browser-e2e-webserver-backed`, and `browser-e2e-stateful`. | Earlier-phase green context only. |
| FE-P7 visual evidence root | `.cartulary/test-results/20260610T183855Z-p48644/browser-e2e-visual/frontend-row-accounting.json`. | Visual handoff context only. |
| FE-P7 accessibility evidence root | `.cartulary/test-results/20260610T183422Z-p29533/browser-e2e-a11y/frontend-row-accounting.json`. | Accessibility handoff context only. |
| FE-P7 broad check | `.cartulary/test-results/20260610T184536Z-p20470` passed per FE-P7 plan. | Historical health context only. |
| FE-P7 finalization | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260610T184536Z-p20470` passed per FE-P7 plan. | Historical finalization context only. |
| FE-P7 strict non-claims | FE-P7 plan explicitly excludes saved-view persistence, rollback readiness, full coordination-surface readiness, product conformance from visual/a11y, and Core 05 publication. | These non-claims remain active for FE-P8 unless FE-P8 produces direct evidence. |

FE-P7 evidence may show that earlier frontend dependencies were green at the time of the FE-P7 handoff. FE-P8 must still produce direct row-owned evidence for saved views, query controls, layout/query persistence, grouping, active chips, default/startup controls, and non-writable group rows.

## Phase Objective

FE-P8 must deliver and verify saved-view selector behavior, sort/filter/group controls, active filter chips, layout and query persistence, default and startup surface controls, grouped-result presentation, non-writable group rows, and query replay through public route contracts using stable `view_schema_id`, `saved_view_id`, `field_key`, `record_id`, `row_version`, and route identifiers rather than visible labels, DOM order, row index, or vendor coordinates.

## Implementation Scope

FE-P8 may change these implementation surfaces:

| Surface | FE-P8 scope |
| --- | --- |
| `/apps/web` | Workbook shell, saved-view selector, active surface selector integration, query controls, active chips, startup/default UI, query-state serialization, public route calls, browser/E2E scenarios, and visual/a11y specs. |
| `/packages/grid-adapter` | Group presentation rows, identity handling, sort/group state adapters, focus/selection behavior around group rows, and edit/paste/mutation blocking for presentation rows. |
| `/packages/view-contracts` | Contract-derived sortable/filterable/groupable field metadata and stable field-key lookup. |
| `/packages/protocol-ts` | Generated public protocol contract dependencies only. Generated files must not be hand-edited. |
| `/packages/ui-contracts` | Stable selector and test-id builders for saved views, query controls, chips, group rows, cells, and surfaces. |
| `/packages/test-utils` | Browser command helpers for sort, filter, group, active chips, group expand/collapse, layout persistence, startup/default controls, and stable anchors. |
| `/apps/web/e2e` | FE-P8 browser, stateful, visual, and accessibility scenarios with live row IDs and scenario titles from the FE-P8 map. |

FE-P8 behavior must be keyed by stable `view_schema_id`, `saved_view_id`, `field_key`, `record_id`, `row_version`, and public route contracts where applicable. It must not be keyed by visible labels, DOM order, row index, CSS coordinates, `react-data-grid` vendor internals, or mutable presentation-row positions.

Sort changes must compile to view-query `sort[]` entries keyed by sortable `field_key`. Filters and groups must use schema fields and public query contracts. Group rows must be presentation-only and must not emit mutation-capable edit, paste, export, history, preview, or row-mutation events. Saved views must appear under the active surface selector. Startup and default choices must persist through owner contracts.

## Out of Scope

FE-P8 must not include:

1. Timeline create or patch semantics beyond query state.
2. Evidence handle redemption, preview, download, or object-store behavior.
3. Same-field conflict resolver internals except regression anchoring and non-claim boundaries.
4. Inspector rollback, destructive confirmation, relationship tabs, or FE-P9 inspector behavior.
5. Publication, benchmark, fixture-sensitive publication, or Core 05 claim readiness.
6. Raw object-store behavior.
7. Direct `react-data-grid` imports outside `/packages/grid-adapter`.
8. Hand edits to generated ledgers or generated artifacts.
9. FE-P9 or later implementation work except dependency and blocker context.
10. Workbook collaboration WebSocket behavior except regression checks when live updates affect query/group/saved-view state.
11. Evidence handle readiness, rollback readiness, or claim-publication readiness.

## Row Inventory

| Row ID | Layer | Evidence class | Owner sections | Exact REQ IDs | Exact AC IDs | Repository target or TODO | Current `claim_status` | Expected closure evidence | Blockers | Non-claims |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| FE-U-P8-01 | unit | product_conformance | Core 01 Section 3.3.4; Core 03 Section 14 | REQ-01-035, REQ-01-038, REQ-01-047, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-243, AC-359, AC-361, AC-363, AC-364, AC-372, AC-375, AC-387 | `make frontend-unit`; TODO: add/promote exact FE-U-P8-01 unit evidence and row accounting. | blocked | Current `frontend-row-accounting.json` from `frontend-unit` with FE-U-P8-01 row, current map/registry/guide digests, passing unit identity, and closure eligibility. | `FE-U-P8-01-BLOCKER-01`; no current FE-P8 unit row in inspected latest accounting. | Does not claim saved-view persistence, browser behavior, visual readiness, a11y readiness, or Core 05 publication. |
| FE-I-P8-01 | integration | product_conformance | Core 01 Section 3.3.5.2; Core 03 Section 3 | REQ-01-138, REQ-01-151, REQ-03-012, REQ-03-032 | AC-146, AC-153, AC-231, AC-233, AC-360 | `make frontend-unit`; `make browser-e2e-webserver-backed`; TODO: add/promote exact FE-I-P8-01 integration evidence and row accounting. | blocked | Current row accounting from mapped targets proving saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts. | `FE-I-P8-01-BLOCKER-01`; no current FE-P8 row in inspected latest accounting. | Does not claim E2E reload persistence, visual readiness, a11y readiness, or Core 05 publication. |
| FE-B-P8-01 | browser_integration | product_conformance | Core 03 Section 3; Core 03 Section 14; implementation guide Section 14.9A | REQ-03-012, REQ-03-032, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-146, AC-153, AC-231, AC-233, AC-360, AC-363, AC-364 | `make browser-e2e-webserver-backed`; `make browser-e2e-support`; TODO: add/promote exact FE-B-P8-01 browser helper evidence and row accounting. | blocked | Current row accounting with scenario title `FE-B-P8-01 Verify browser command helpers for sort, filter, group, active chips, layout persistence, group expand-collapse, and startup/default surface UI.` | `FE-B-P8-01-BLOCKER-01`; `browser-e2e-support` is an internal helper and target explanation does not currently list phase8 coverage. | Does not claim route persistence after reload, visual readiness, a11y readiness, or Core 05 publication. |
| FE-E-P8-01 | e2e | product_conformance | Core 01 Section 3.3.4; Core 01 Section 3.3.5.2; Core 03 Section 3; Core 03 Section 14 | REQ-01-035, REQ-01-038, REQ-01-047, REQ-01-138, REQ-01-151, REQ-03-012, REQ-03-032, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-124, AC-127, AC-146, AC-153, AC-184, AC-185, AC-231, AC-233, AC-238, AC-243, AC-359, AC-361, AC-363, AC-364, AC-372, AC-375, AC-387 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; TODO: add/promote exact FE-E-P8-01 public-route reload/replay evidence and row accounting. | blocked | Current row accounting with scenario title `FE-E-P8-01 Verify saved-view persistence, default/startup surface persistence, and query replay through /api/v1/ after reload.` plus public route request/response artifacts. | `FE-E-P8-01-BLOCKER-01`; latest current-digest accounting has no FE-P8 row. | Does not claim visual readiness, a11y readiness, rollback readiness, evidence handles, or Core 05 publication. |
| FE-V-P8-01 | visual | design_direction | UI/UX guide Sections 7, 10.5, 13; visual golden guide Sections 2, 3, 5 | None in live map | R2-AC-027, R2-AC-032, R2-AC-051, R2-AC-054, R2-AC-073, R2-AC-079, R2-RDG-AC-001, R2-RDG-AC-010 | `make browser-e2e-visual`; TODO: recapture exact FE-V-P8-01 visual readiness evidence against current map/registry/guide digests. | blocked | Current visual row accounting and fixture artifacts for saved-view selector, active chips, grouped result, group row, default/startup state indicator, and empty successful query fixtures. | `FE-V-P8-01-BLOCKER-01`; retained visual accounting is stale and omits FE-V-P8-01; fixture registry alone is insufficient. | Does not claim product conformance, Core 05 visual publication, or row closure for product rows. |
| FE-A11Y-P8-01 | accessibility | design_direction | UI/UX guide Sections 7, 10.5, 14 | None in live map | R2-AC-027, R2-AC-032, R2-AC-051, R2-AC-054, R2-AC-080, R2-AC-086, D-AC-009, D-AC-012 | `make browser-e2e-a11y-preflight`; TODO: recapture exact FE-A11Y-P8-01 accessibility preflight evidence with accepted row accounting. | blocked | Current accessibility artifact proving keyboard reachability and announced state for sort, filter, group, saved-view menu, active chips, group expand-collapse, and default/startup controls, plus map-authorized row accounting. | `FE-A11Y-P8-01-BLOCKER-01`; retained preflight passes but remains blocked and lacks root frontend row accounting. | Does not claim product conformance, full a11y suite closure, or Core 05 publication. |

## Evidence Layer Matrix

| Row ID | Target(s) | Evidence class | Row-accounting artifact family | Closure predicate | Non-claim boundary |
| --- | --- | --- | --- | --- | --- |
| FE-U-P8-01 | `frontend-unit` | product_conformance | `.cartulary/test-results/<run-id>/frontend-unit/frontend-row-accounting.json` and target/phase summaries | Row appears with current digests, mapped command ID, passing unit identity, `claim_status` implemented or map-promoted equivalent, no blockers, and closure eligibility. | Unit evidence alone does not prove browser persistence, visual readiness, accessibility readiness, or publication. |
| FE-I-P8-01 | `frontend-unit`, `browser-e2e-webserver-backed` | product_conformance | Mapped frontend-unit and webserver-backed row accounting plus public route/browser artifacts when required by scenario | Row appears in current mapped artifacts with saved-view create/update/select/default UI evidence through active surface scope and public saved-view/workbook-preference contracts. | Integration evidence does not prove reload replay unless FE-E-P8-01 closes. |
| FE-B-P8-01 | `browser-e2e-webserver-backed`, `browser-e2e-support` | product_conformance | Webserver-backed row accounting and support-helper accounting/summaries | Row appears with exact scenario title and browser command helper coverage for sort, filter, group, chips, layout persistence, group expand/collapse, and startup/default controls. | Support-helper evidence cannot independently close product conformance without mapped primary row accounting. |
| FE-E-P8-01 | `browser-e2e-webserver-backed`, `browser-e2e-stateful` | product_conformance | Browser row accounting, route trace artifacts, target summaries, and retained run root | Row appears with exact scenario title and public `/api/v1/` route evidence proving saved-view persistence, default/startup persistence, and query replay after reload. | E2E product evidence does not promote visual/a11y design-direction rows. |
| FE-V-P8-01 | `browser-e2e-visual` | design_direction | Visual frontend row accounting, fixture registry linkage, screenshot/golden artifacts, target summaries | Row appears with current digests and captured fixtures for every mapped visual state; fixture registry is current; no stale digest or missing fixture blockers remain. | Visual evidence must remain design-direction and cannot prove product conformance or Core 05 publication. |
| FE-A11Y-P8-01 | `browser-e2e-a11y-preflight` | design_direction | Accessibility preflight summary, row accounting if required by map, target summaries | Row appears with current accepted accessibility evidence for keyboard reachability and announced states; no blocked-row or missing-accounting condition remains. | Accessibility preflight must not be used as product conformance or full accessibility closure without map authority. |

## Dependencies And Prerequisites

| Dependency | Required FE-P8 posture | Current inspected status |
| --- | --- | --- |
| FE-P0 through FE-P7 | Must remain green because FE-P8 depends on all earlier frontend phases. | Registry reports FE-P0 through FE-P7 `active_green`. FE-P7 plan records closed rows, but that is handoff context only. |
| Generated contracts | Must be current before FE-P8 row closure; generated roots must not be edited by hand. | `packages/protocol-ts/src/generated/index.ts` is generated. No generated file was edited for this plan. |
| View contracts | Must expose stable `view_schema_id`, sortable/filterable/groupable `field_key`, capability metadata, default sort, and synthetic predicates. | `packages/view-contracts/src/index.ts` inspected as contract-derived metadata surface. |
| Public query routes | Must support view-query `sort[]`, `filters[]`, and `group_by` through `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`. | Route behavior must be proven by mapped browser/E2E artifacts, not plan text. |
| Saved-view routes | Must support list/create/update/select/default evidence under incident and active-surface scope. | `apps/web/e2e/phase8.workbook.spec.ts` contains route-context tests, but current FE-P8 row accounting is absent. |
| Workbook preference routes | Must support user home sheet and incident default sheet persistence through public contracts. | Candidate code and E2E surfaces exist; closure is blocked pending row-owned evidence. |
| Grid adapter | Must keep direct `react-data-grid` dependency inside `/packages/grid-adapter`, protect identity anchors, and make group rows presentation-only. | `packages/grid-adapter/src/core.ts` and `src/index.tsx` inspected. |
| Selectors/test IDs | Must use stable generated builders for saved-view selector, chips, group rows, cells, and controls. | `packages/ui-contracts/src/index.ts` inspected. |
| Browser helpers | Must cover sort, filter chips, group change, expand/collapse, anchors, paste/fill-down, and stable selection. | `packages/test-utils/src/index.ts` inspected. |
| Visual fixtures | Must capture all FE-V-P8-01 mapped states with current digests and row accounting. | Registry has current shared fixtures, but FE-V-P8-01 remains blocked. |
| Accessibility harness | Must prove keyboard reachability and announced state through mapped FE-A11Y-P8-01 evidence. | Preflight artifact passes but remains blocked and lacks root row accounting. |
| Registry/map/ledger freshness | Must align before promotion. | Current file digests align for registry map/ledger references; retained FE-P8 evidence does not close rows. |

## Shared Harness Analysis

| Harness or shared surface | FE-P8 rows affected | Required analysis and closure posture |
| --- | --- | --- |
| Contract-derived view-schema and field-key mapping | FE-U-P8-01, FE-I-P8-01, FE-E-P8-01 | Tests must prove query state uses schema `field_key` and `view_schema_id`, not labels, DOM order, storage names, or table names. |
| Grid-adapter identity and capability invariants | FE-U-P8-01, FE-B-P8-01, FE-E-P8-01, FE-V-P8-01 | Group rows must be presentation rows without writable `record_id`; record rows must remain anchored by `record_id`, `field_key`, and `row_version` where mutation is possible. |
| Renderer/editor lifecycle cleanup | FE-B-P8-01, FE-E-P8-01, FE-A11Y-P8-01 | Sort/filter/group changes must not leave stale editors, focus traps, stuck popovers, or editable controls attached to group rows. |
| Sync-engine pending queue and replay | FE-E-P8-01 | Query state and saved-view selection must not corrupt pending mutation replay. FE-P8 must only claim saved-view/query replay through owner routes. |
| WebSocket reducer behavior | FE-E-P8-01 | Live updates may be regression context when query/group/saved-view state intersects refreshed rows. FE-P8 must not claim FE-P7 collaboration behavior unless row-owned FE-P8 evidence maps it. |
| Same-field conflict anchoring | FE-E-P8-01 | Same-field conflict handling is regression/non-claim context only. FE-P8 must not modify resolver internals unless a blocker proves a necessary query-state regression fix. |
| Presence anchoring under sort/group/scroll | FE-B-P8-01, FE-E-P8-01 | Presence markers must stay anchored to stable record/cell identities when sort/group/scroll state changes. FE-P8 may verify regressions but must not claim full collaboration closure. |
| Save-state presentation | FE-B-P8-01, FE-E-P8-01, FE-V-P8-01 | Query and saved-view changes must not misrepresent pending/saved/error state. Visual evidence remains design-direction unless mapped otherwise. |
| Browser command helpers | FE-B-P8-01 | Helpers must interact through stable test IDs and public UI controls for sort, filter, group, active chips, group expand/collapse, layout persistence, and startup/default controls. |
| Visual-regression fixtures | FE-V-P8-01 | Fixtures must cover saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query with current fixture registry and row accounting. |
| Keyboard/focus traversal | FE-A11Y-P8-01 | Sort/filter/group controls, saved-view selector, active chips, group rows, expand/collapse, and default/startup controls must be reachable by keyboard. |
| Accessibility names and ARIA | FE-A11Y-P8-01 | Controls and group rows must expose announced state, expanded/collapsed state, and non-color-only affordances. Preflight must not become product conformance. |
| Stable selector/test-id contracts | FE-B-P8-01, FE-E-P8-01, FE-V-P8-01, FE-A11Y-P8-01 | Tests must use builders from `packages/ui-contracts` rather than hand-rolled selectors keyed to visible labels or vendor DOM coordinates. |
| Frontend route/API boundary conformance | FE-I-P8-01, FE-E-P8-01 | Saved-view and preference persistence must be proven through public `/api/v1/` routes. Frontend mocks may support unit tests but cannot close route persistence claims. |

## Public Interfaces And Deliverables

FE-P8 deliverables must include:

| Deliverable | Required public surface | Evidence boundary |
| --- | --- | --- |
| Saved-view selector | Active surface selector UI with saved views scoped by `view_schema_id` and selected by `saved_view_id`. | Product rows FE-I-P8-01 and FE-E-P8-01. Visual readiness FE-V-P8-01. Accessibility readiness FE-A11Y-P8-01. |
| Sort controls | Sort toggles compiled to public view-query `sort[]` with sortable `field_key` and direction. | FE-U-P8-01, FE-B-P8-01, FE-E-P8-01. |
| Filter controls and active chips | Filter operations using schema fields, supported operators, active chips, and chip removal keyed by `field_key`. | FE-U-P8-01, FE-B-P8-01, FE-E-P8-01. |
| Grouping controls and group rows | Group by schema groupable fields; group rows are presentation-only and non-writable. | FE-U-P8-01, FE-B-P8-01, FE-E-P8-01, FE-V-P8-01, FE-A11Y-P8-01. |
| Layout/query persistence | Saved-view `query_json` and `layout_json` through owner contracts, without selection, scroll, focus, popover, inspector, preview, or presence state. | FE-I-P8-01 and FE-E-P8-01. |
| Default/startup surface controls | User home sheet and incident default sheet choices through public workbook-preference/startup contracts. | FE-I-P8-01, FE-E-P8-01, FE-V-P8-01, FE-A11Y-P8-01. |
| Selector/test-id surfaces | `gridShellTestId`, `gridSortHeaderTestId`, `gridFilterChipTestId`, `gridGroupingSelectTestId`, `gridGroupRowTestId`, `savedViewSelectorTestId`, `savedViewOptionTestId`, and row/cell selectors. | Selector stability support for browser, visual, and accessibility evidence. |
| Generated contract dependencies | Protocol/view contract outputs consumed through packages. | Generated outputs must be refreshed by Make generators only. |
| Fixture and visual artifacts | Current screenshots/goldens and frontend visual row accounting for FE-V-P8-01. | Design-direction only. |
| Accessibility artifacts | Preflight or mapped a11y artifacts for FE-A11Y-P8-01. | Design-direction only unless map authority changes. |
| Map/registry/ledger updates | Owner map promotion, registry status update, and regenerated ledger after direct evidence exists. | Must be produced from owner inputs and generators; generated ledgers must not be edited by hand. |
| Retained artifact families | Target summaries, phase summaries, row accounting, route traces, visual outputs, a11y summaries, and finalization artifacts. | Must be current, row-owned, and mapped to FE-P8 rows before closure. |

## Sprint Checklist

- [x] Inspect live registry, FE-P8 map, FE-P8 ledger, visual fixture registry, governing docs, prior frontend plans, relevant app/package code, relevant tests, and target explanations.
- [x] Record current FE-P8 status as planned/blocked with no accepted FE-P8 row closure.
- [ ] Reconcile guide range language and live map exact REQ IDs before expanding any row claim.
- [ ] Activate or promote FE-P8 map/registry rows only after direct current evidence exists.
- [ ] Implement or validate FE-U-P8-01 query-state compilation using stable `field_key` and capability metadata.
- [ ] Implement or validate FE-I-P8-01 saved-view create/update/select/default UI through public contracts.
- [ ] Implement or validate FE-B-P8-01 browser command helpers and group-row behavior.
- [ ] Implement or validate FE-E-P8-01 saved-view persistence, default/startup persistence, reload behavior, and query replay through `/api/v1/`.
- [ ] Implement or validate FE-V-P8-01 visual fixtures without promoting design-direction evidence to product conformance.
- [ ] Implement or validate FE-A11Y-P8-01 keyboard and announcement evidence without promoting preflight evidence to product conformance.
- [ ] Run mapped unit, integration, browser, E2E, visual, and accessibility validation commands after FE-P8 activation.
- [ ] Run `make phase-ledger-drift`, `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check`.
- [ ] Run `make check` only when environment readiness and phase work justify the full gate.
- [ ] Run `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when retained successful full-check evidence is used.
- [ ] Update the blocker register with every unclosed row and exact source condition.
- [ ] Prepare FE-P9 handoff with final FE-P8 registry status, evidence roots, blockers, drift/finalization outcomes, and strict non-claims.

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness And Source Alignment

| Field | Plan |
| --- | --- |
| Objective | Establish the live FE-P8 source baseline before implementation or evidence promotion. |
| Implementation targets | `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p8_test_map.json`, `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md`, `tools/frontend_visual_fixture_registry.json`, governing docs, previous plans, retained artifacts, and Make explanations. |
| Validation commands | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8`; all `make explain-target ... DETAIL=summary` commands listed in this plan; `git status --short`. |
| Expected evidence | Source inventory with file digests, registry status, row inventory, target explanation results, and blocker register. |
| Blockers | Any missing owner ref, stale digest, guide/map/ledger mismatch, absent target, or retained artifact that conflicts with current map state. |
| Strict non-claims | This sprint must not claim product behavior, row closure, generated-ledger authority, or FE-P8 activation. |

### Sprint 2: Unit And Query-State Foundation

| Field | Plan |
| --- | --- |
| Objective | Implement or validate deterministic query-state compilation for sort, filters, groups, layout state, and active chips. |
| Implementation targets | `apps/web/src/workbookQuery.ts`, `apps/web/src/WorkbookGridControls.tsx`, `packages/view-contracts/src/index.ts`, `packages/ui-contracts/src/index.ts`, `apps/web/src/WorkbookShell.phase8.query.test.tsx`. |
| Validation commands | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make explain-target TARGET=frontend-unit DETAIL=summary`. |
| Expected evidence | FE-U-P8-01 row-owned `frontend-unit` accounting with current digests; unit names or scenario identities proving `sort[]`, `filters[]`, `group_by`, active chips, layout state serialization, and capability filtering use stable `field_key`. |
| Blockers | Visible-label keys, DOM-order keys, row-index keys, unsupported field operations, non-sortable field sorting, group rows with writable IDs, stale map digests, or missing FE-U-P8-01 accounting. |
| Strict non-claims | Unit closure must not claim saved-view route persistence, reload behavior, visual readiness, or accessibility readiness. |

### Sprint 3: Saved-View Integration

| Field | Plan |
| --- | --- |
| Objective | Implement or validate saved-view create/update/select/default UI using active surface scope and public saved-view/workbook-preference contracts. |
| Implementation targets | `apps/web/src/WorkbookShell.tsx`, `apps/web/src/workbookStartup.ts`, `apps/web/src/workbookSurfaceRegistry.ts`, `apps/web/e2e/phase8.workbook.spec.ts`, `packages/protocol-ts`, `packages/view-contracts`, `packages/ui-contracts`. |
| Validation commands | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make frontend-typecheck`; `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`. |
| Expected evidence | FE-I-P8-01 row accounting from mapped targets with public route evidence for active-surface saved-view scope, saved-view selection, owner query/layout persistence, user home sheet, and incident default sheet where authorized. |
| Blockers | Saved-view state proved only through mocks, saved views not scoped by active `view_schema_id`, query/layout JSON containing visible labels or row IDs, default/startup persistence without public route evidence, or missing FE-I-P8-01 accounting. |
| Strict non-claims | This sprint must not claim reload replay unless FE-E-P8-01 closes and must not claim visual/a11y readiness. |

### Sprint 4: Browser Commands And Group-Row Behavior

| Field | Plan |
| --- | --- |
| Objective | Verify browser command helpers, group expand/collapse, active chips, layout persistence, default/startup controls, and non-writable group rows. |
| Implementation targets | `packages/test-utils/src/index.ts`, `packages/grid-adapter/src/core.ts`, `packages/grid-adapter/src/index.tsx`, `packages/ui-contracts/src/index.ts`, `apps/web/e2e/phase8.workbook.spec.ts`. |
| Validation commands | `make browser-e2e-webserver-backed`; `make browser-e2e-support`; `make frontend-import-boundary-check`; `make explain-target TARGET=browser-e2e-support DETAIL=summary`. |
| Expected evidence | FE-B-P8-01 row accounting with exact scenario title, browser command helper execution, stable selector use, group expand/collapse proof, active chip add/remove proof, and group-row non-mutation proof. |
| Blockers | Direct `react-data-grid` imports outside adapter, helper selectors based on visible labels or vendor coordinates, group rows emitting mutation-capable events, support target evidence without primary accounting, or missing FE-B-P8-01 accounting. |
| Strict non-claims | Browser helper evidence must not claim route persistence after reload unless FE-E-P8-01 closes. |

### Sprint 5: Stateful And E2E Persistence

| Field | Plan |
| --- | --- |
| Objective | Verify saved-view persistence, default/startup surface persistence, reload behavior, and query replay through `/api/v1/`. |
| Implementation targets | `apps/web/src/WorkbookShell.tsx`, `apps/web/src/workbookStartup.ts`, `apps/web/e2e/phase8.workbook.spec.ts`, public saved-view/preference/query routes, generated protocol contracts. |
| Validation commands | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make explain-target TARGET=browser-e2e-stateful DETAIL=summary`. |
| Expected evidence | FE-E-P8-01 row accounting with exact scenario title, public route traces, reload proof, persisted saved-view selection, default/startup surface persistence, and query replay using public `/api/v1/` contracts. |
| Blockers | Persistence asserted without route evidence, state stored only in browser memory, query replay using labels or DOM order, group rows treated as records, stale row-accounting digests, or missing FE-E-P8-01 accounting. |
| Strict non-claims | E2E product evidence must not promote FE-V-P8-01 or FE-A11Y-P8-01 to product conformance. |

### Sprint 6: Visual And Accessibility Readiness

| Field | Plan |
| --- | --- |
| Objective | Verify visual and accessibility readiness for FE-P8 while keeping design-direction evidence separate from product conformance. |
| Implementation targets | `apps/web/e2e/workbook.visual.spec.ts`, `apps/web/e2e/workbook.a11y-preflight.spec.ts`, `tools/frontend_visual_fixture_registry.json`, `docs/guides/cartulary_visual_golden_maintenance.md`, `packages/ui-contracts/src/index.ts`. |
| Validation commands | `make browser-e2e-visual`; `make browser-e2e-a11y-preflight`; `make explain-target TARGET=browser-e2e-visual DETAIL=summary`; `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary`. |
| Expected evidence | FE-V-P8-01 visual row accounting and fixture artifacts for saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query; FE-A11Y-P8-01 accessibility evidence for keyboard reachability and announced state. |
| Blockers | Stale visual digests, missing visual states, fixture registry status without row accounting, preflight pass with blocked claim status, no root row accounting when map requires it, or accessibility evidence represented as product conformance. |
| Strict non-claims | Visual and accessibility evidence remain design-direction and must not claim Core 05 publication readiness. |

### Sprint 7: Drift, Full Check, Finalization, Closeout, And FE-P9 Handoff

| Field | Plan |
| --- | --- |
| Objective | Validate generated and harness state, run the required gate sequence for the final implementation posture, close blockers, and produce FE-P9 handoff. |
| Implementation targets | Owner maps, registry owner inputs, generated ledgers via Make targets, retained artifact roots, final blocker register, FE-P9 handoff section. |
| Validation commands | `make phase-ledger-drift`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make check`; `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when retained full-check evidence is used. |
| Expected evidence | Current map/registry/ledger digests, generated-artifact drift pass, JSON shape pass, full check pass when run, finalization pass when using retained evidence, and FE-P9 handoff with direct FE-P8 evidence roots. |
| Blockers | Generated-ledger drift, stale freshness digest, old retained artifacts, broad `make check` without row-owned evidence, `agent-finalize` without `RESULTS_DIR` when retained evidence is used, or any unclosed FE-P8 row. |
| Strict non-claims | Full `make check` must not close a FE-P8 row unless the mapped row-owned accounting and current map/registry/guide digests are present. |

## Validation Commands

### Explanation Commands

These commands must be run before depending on any validation target in FE-P8:

```sh
make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8
make explain-target TARGET=frontend-unit DETAIL=summary
make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary
make explain-target TARGET=browser-e2e-stateful DETAIL=summary
make explain-target TARGET=browser-e2e-support DETAIL=summary
make explain-target TARGET=browser-e2e-visual DETAIL=summary
make explain-target TARGET=browser-e2e-a11y DETAIL=summary
make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary
make explain-target TARGET=phase-ledger-drift DETAIL=summary
make explain-target TARGET=generate-drift DETAIL=summary
make explain-target TARGET=generated-artifact-policy-check DETAIL=summary
make explain-target TARGET=json-shape-check DETAIL=summary
make explain-target TARGET=check DETAIL=summary
make explain-target TARGET=agent-finalize DETAIL=summary
make explain-target TARGET=frontend-typecheck DETAIL=summary
make explain-target TARGET=frontend-import-boundary-check DETAIL=summary
```

All explanation commands above were run for plan creation and existed in the public Make surface or explainable target surface. No compatibility aliases for legacy probe names must be added.

### Execution Commands

The following execution commands are canonical public Make targets, except `browser-e2e-support`, which is explainable as an internal helper named by the live FE-P8 map. Absent or internally scoped target semantics must be recorded as blockers if they prevent row closure.

```sh
make frontend-unit
make frontend-typecheck
make frontend-import-boundary-check
make browser-e2e-webserver-backed
make browser-e2e-stateful
make browser-e2e-support
make browser-e2e-visual
make browser-e2e-a11y-preflight
make phase-ledger-drift
make generate-drift
make generated-artifact-policy-check
make json-shape-check
make check
make agent-finalize RESULTS_DIR=<successful-full-check-run-root>
```

`make browser-e2e-a11y` exists and was explained, but the live FE-P8 map names `browser-e2e-a11y-preflight` for FE-A11Y-P8-01. `make browser-e2e-a11y` may be used as support evidence only when the map or owner guidance explicitly requires it; it must not replace the mapped preflight row evidence.

`make agent-finalize` without `RESULTS_DIR` may be run only when retained-run maintenance is intentionally skipped. The final report must record that reason. When a retained successful full-check run is used, `RESULTS_DIR=<successful-full-check-run-root>` is required.

## Evidence Requirements

Every FE-P8 evidence payload must include:

1. Row-accounting schema ID and row ID.
2. Target identity and command ID.
3. Accounting scope and phase namespace.
4. Evidence class.
5. Scenario, unit, or artifact identity.
6. Owner-map digest or freshness evidence.
7. Registry digest.
8. Guide digest.
9. Pass/fail status.
10. Closure status.
11. Artifact root.
12. Exact blocker when missing or stale.
13. Per-row artifact family.

| Row ID | Required artifact family | Minimum evidence payload |
| --- | --- | --- |
| FE-U-P8-01 | `frontend-unit` row accounting, target summary, phase summary, unit test output | `cartulary.frontend_row_accounting.v3`, row FE-U-P8-01, command `cartulary.harness.command.frontend_unit.v1`, target `frontend-unit`, current guide digest `3750847aac8805eb59a1f93430de24a036665241aa7ad3eb6dfb13dcb349b0b5`, current registry digest, current FE-P8 map digest, passing unit identity for sort/filter/group/layout/chip compilation, closure status closed, artifact root. |
| FE-I-P8-01 | `frontend-unit` and `browser-e2e-webserver-backed` row accounting, route/browser artifacts | Row FE-I-P8-01, command IDs for mapped targets, exact scenario title, public saved-view/workbook-preference contract evidence, active surface scope evidence, current digests, pass status, closure status, artifact roots. |
| FE-B-P8-01 | `browser-e2e-webserver-backed` and `browser-e2e-support` row accounting/summaries | Row FE-B-P8-01, exact scenario title, command helper proof for sort/filter/group/chips/layout/group expand-collapse/startup-default controls, stable selector evidence, current digests, pass status, closure status, artifact roots. |
| FE-E-P8-01 | `browser-e2e-webserver-backed` and `browser-e2e-stateful` row accounting, route traces, reload artifacts | Row FE-E-P8-01, exact scenario title, public `/api/v1/` saved-view/preference/query route evidence, reload replay evidence, current digests, pass status, closure status, artifact roots. |
| FE-V-P8-01 | `browser-e2e-visual` row accounting, fixture registry references, screenshots/goldens, target summary | Row FE-V-P8-01, exact scenario title, current guide/registry/map digests, captured states for saved-view selector, active chips, grouped result, group row, default/startup indicator, empty successful query, pass status, design-direction closure status, artifact roots, no stale fixture blockers. |
| FE-A11Y-P8-01 | `browser-e2e-a11y-preflight` row accounting or map-authorized preflight summary, target summary | Row FE-A11Y-P8-01, exact scenario title, current digests, keyboard reachability and announced state evidence for all mapped controls, pass status, design-direction closure status, artifact root, no blocked/preflight-only closure mismatch. |

## Blocker Rules

The following conditions must create blockers:

1. Unresolved owner refs, missing owner sections, or owner REQ/AC IDs that cannot be found in authoritative sources.
2. Guide/map/ledger mismatch, including broader guide ranges that are not represented in exact live map IDs.
3. Stale guide, registry, phase-map, ledger, or evidence freshness digest.
4. Generated-ledger drift or hand-edited generated ledgers.
5. Retained artifacts older than the current map/registry/guide digests.
6. Product rows relying only on frontend mocks or unit-only evidence for public route behavior.
7. Public-route behavior not proven through public `/api/v1/` routes for saved-view persistence, default/startup persistence, and query replay.
8. Saved-view persistence asserted without saved-view route and workbook-preference route evidence.
9. Sort/filter/group behavior keyed by visible label, DOM order, row index, storage table name, or vendor coordinates.
10. Group rows emitting edit, paste, export, history, preview, or mutation-capable events.
11. Direct `react-data-grid` imports outside `/packages/grid-adapter`.
12. Visual readiness represented as product conformance.
13. Accessibility preflight used as row closure without map authority and current row accounting when required.
14. Core 05 publication readiness implied without claim-bearing metadata.
15. Broad `make check` used as row closure when FE-P8 row-owned accounting is absent.
16. `make agent-finalize` used without `RESULTS_DIR` while retaining a full-check run.

| Affected row | Affected source | Exact observed condition | Why closure is blocked | Minimum follow-up | Prevents |
| --- | --- | --- | --- | --- | --- |
| All FE-P8 | `tools/frontend_phase_registry.json` | FE-P8 status `planned`, rollup `no_rows_implemented`, activation blocker `FE-P8-ACTIVATION-BLOCKER-01`. | Planned phases are explainable but not executable as whole-phase closure. | Promote direct row evidence and freshness validation together through owner inputs. | FE-P8 activation and full phase completion. |
| All FE-P8 | FE-P8 guide vs live map | Guide uses broader REQ ranges than exact live map IDs. | Row closure cannot expand owner claims beyond exact live map IDs. | Reconcile guide/map or update owner map and regenerate ledger. | Expanded guide-range claims. |
| FE-U-P8-01 | Latest `frontend-unit` row accounting | `.cartulary/test-results/20260611T141819Z-p21786/frontend-unit/frontend-row-accounting.json` has no FE-P8 rows. | No current row-owned unit evidence. | Add/promote FE-U-P8-01 unit evidence and rerun mapped target. | FE-U-P8-01 closure. |
| FE-I-P8-01 | Latest `frontend-unit` and `browser-e2e-webserver-backed` row accounting | Latest current-digest artifacts have no FE-P8 rows. | No current row-owned saved-view integration evidence. | Add/promote FE-I-P8-01 evidence and rerun mapped targets. | FE-I-P8-01 closure. |
| FE-B-P8-01 | Live map and target explanation | Map names `browser-e2e-support`; target explanation marks it `internal_helper` and phase coverage does not list phase8. | Helper evidence cannot silently become closure evidence. | Confirm support target accounting path after activation or update map/task surface. | FE-B-P8-01 closure. |
| FE-E-P8-01 | Latest `browser-e2e-stateful` row accounting | Current-digest stateful accounting has no FE-P8 rows; target explanation phase coverage does not list phase8. | No current row-owned E2E persistence evidence. | Add/promote FE-E-P8-01 scenario and rerun mapped targets with public route traces. | FE-E-P8-01 closure. |
| FE-V-P8-01 | Retained visual row accounting | Visual accounting uses stale guide/registry digests and omits FE-V-P8-01. | Stale visual artifact cannot close current FE-P8 readiness. | Recapture FE-V-P8-01 visual evidence with current digests. | FE-V-P8-01 readiness. |
| FE-V-P8-01 | Fixture registry | Current FE-P8-owned fixtures cover grouped result, group row, and empty successful query, but no separate FE-P8-owned fixture IDs were found for saved-view selector, active chips, or default/startup indicator. | Fixture coverage is incomplete for the live visual row wording unless row-owned visual accounting captures the complete scenario. | Add fixture ownership or capture full FE-V-P8-01 scenario under current accounting. | FE-V-P8-01 readiness. |
| FE-A11Y-P8-01 | Retained preflight summary | Scenario passes, but phase row still has `claim_status` `blocked`, `required_for_closure` `false`, and no root `frontend-row-accounting.json`. | Preflight pass is not accepted row closure under current map state. | Produce accepted row accounting or update map authority, then rerun preflight. | FE-A11Y-P8-01 readiness. |
| All product rows | Existing FE-P8-adjacent tests | Existing app tests use legacy `U-8`/`E-8` naming and current row accounting omits FE-P8 rows. | Tests are implementation context, not current row-owned evidence. | Retitle/map scenarios to exact FE-P8 row IDs and produce current accounting. | Product row closure. |
| All FE-P8 | Core 05 boundary | No explicit FE-P8 claim-bearing metadata inspected. | Publication readiness cannot be inferred. | Add Core 05-compliant claim metadata only if publication evidence is in scope. | Core 05 claims. |

## Strict Non-Claims

This plan does not claim:

1. FE-P8 completion by document creation.
2. Any FE-P8 row closure from generated ledgers.
3. Any FE-P8 row closure from prior phase plans.
4. Any FE-P8 row closure from old retained artifacts unless current row-owned evidence and map authority permit it.
5. Product conformance from visual evidence.
6. Product conformance from accessibility evidence.
7. Claim-publication readiness or Core 05 readiness.
8. Evidence handle readiness.
9. Inspector rollback readiness.
10. Full FE-P9 readiness.
11. Saved-view persistence from frontend mocks alone.
12. Default/startup persistence without public route evidence.
13. Query-control correctness keyed by visible labels, DOM order, row index, or vendor coordinates.
14. Group-row writability or mutation support.
15. `react-data-grid` import permission outside `/packages/grid-adapter`.
16. Any behavior not mapped to FE-P8 rows.

## Binary Exit Criteria

| Criterion | Pass condition | Fail condition |
| --- | --- | --- |
| 1. Initial plan creation | `FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` exists at repository root, records live inspected inputs, lists blockers, and passes post-write diff/available drift checks. | File missing, live claims unanchored, blockers omitted, generated artifacts edited, or post-write checks fail without blocker recording. |
| 2. Product-conformance row closure | FE-U-P8-01, FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 each have current mapped product-conformance row accounting, current digests, pass status, closure status, exact scenario/unit identity, and no blockers. | Any product row lacks current row-owned evidence, uses stale digests, remains blocked, or depends only on broad `make check`. |
| 3. Saved-view integration closure | FE-I-P8-01 has current mapped evidence for create/update/select/default UI through active surface scope and public saved-view/workbook-preference contracts. | Saved-view behavior is only unit mocked, not active-surface scoped, not public-route proven, or not row-owned. |
| 4. Query-control and grouping closure | FE-U-P8-01 and FE-B-P8-01 prove sort/filter/group/layout/chip compilation and browser commands use stable `field_key` and capability metadata, with non-writable group rows. | Any query key uses visible labels/DOM order, unsupported fields are accepted, or group rows can mutate/export/history/preview as records. |
| 5. Stateful reload and replay closure | FE-E-P8-01 has current public `/api/v1/` route evidence for saved-view persistence, default/startup persistence, reload behavior, and query replay. | Persistence is not route-proven, reload is absent, state exists only in memory, or row accounting is missing/stale. |
| 6. Visual readiness | FE-V-P8-01 has current design-direction row accounting and fixture artifacts for saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query. | Visual evidence is stale, fixture coverage is incomplete, row accounting is absent, or visual evidence is represented as product conformance. |
| 7. Accessibility readiness | FE-A11Y-P8-01 has current map-authorized accessibility evidence and row accounting for keyboard reachability and announced state across all mapped controls. | Preflight remains blocked, row accounting is absent when required, controls are not keyboard reachable, announced state is missing, or evidence is promoted to product conformance. |
| 8. Full FE-P8 phase completion | Registry status is promoted from `planned` to the registry-owner-authorized active/green state, all six rows are closed under current map authority, ledgers are regenerated by Make, drift checks pass, and finalization records retained evidence correctly. | Any row remains blocked, registry/map/ledger freshness fails, generated-ledger drift exists, or retained evidence is stale/missing. |
| 9. FE-P9 handoff readiness | Handoff lists final FE-P8 registry status, row inventory, direct evidence roots, drift/finalization outcomes, strict non-claims, and blockers for unclosed rows. | Handoff treats FE-P8 plan text or earlier phase evidence as FE-P8 closure, omits blockers, or lacks direct evidence roots for closed rows. |

## FE-P9 Handoff

At plan creation, FE-P8 is not complete. FE-P9 may use this plan as dependency context for source limits, implementation surfaces, and blockers. FE-P9 must not treat this plan, FE-P7 evidence, generated FE-P8 ledger text, current fixture-registry status, broad retained `make check`, or retained stale visual/a11y artifacts as FE-P8 row closure.

| Handoff field | Current FE-P8 value |
| --- | --- |
| Registry status | `planned`. |
| Row rollup | `no_rows_implemented`. |
| Activation blocker | `FE-P8-ACTIVATION-BLOCKER-01`. |
| Product rows | FE-U-P8-01, FE-I-P8-01, FE-B-P8-01, FE-E-P8-01 all blocked. |
| Design-direction rows | FE-V-P8-01 and FE-A11Y-P8-01 blocked. |
| Direct accepted FE-P8 evidence roots | None. TODO: `FE-P8-HANDOFF-01` produce current row-owned evidence roots before FE-P9 may depend on FE-P8 behavior. |
| Retained context roots | `.cartulary/test-results/20260611T141819Z-p21786` broad check context; `.cartulary/test-results/20260610T191949Z-p14292` stale visual context; `.cartulary/test-results/20260610T171433Z-p14284` a11y preflight context. These are context only and do not close FE-P8. |
| Drift/finalization at plan creation | Target explanations passed. Post-write `git diff --check` passed with no output. `make phase-ledger-drift` passed at `.cartulary/test-results/20260611T152507Z-p92384`. `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260611T152420Z-p91128`. `make json-shape-check` passed at `.cartulary/test-results/20260611T152420Z-p91151`. `make agent-finalize` was not run because no retained successful full-check evidence was used for FE-P8 closure. |
| Strict non-claims | No product row closure, no visual/product promotion, no accessibility/product promotion, no Core 05 publication, no evidence handle readiness, no rollback readiness, no full FE-P9 readiness. |
| Remaining source limits | Guide/map exact-ID drift, target explanation phase-coverage gaps for mapped planned targets, stale retained visual evidence, preflight-only a11y evidence, and absent current FE-P8 row accounting. |

FE-P9 handoff after FE-P8 implementation must replace this baseline with final registry status, row inventory, direct evidence roots for every closed row, exact blockers for every unclosed row, drift/finalization outcomes, strict non-claims, and any remaining source limits.
