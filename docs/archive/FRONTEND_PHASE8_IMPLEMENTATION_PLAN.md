# Frontend Phase 8 Implementation Plan

## Summary

FE-P8 covers saved views, sorting, filtering, grouping, layout state, active chips, startup and default surface UI, group rows, and query-control persistence for the Cartulary frontend. The live repository status at Sprint 1 plan creation was `planned` with `row_rollup_state` `no_rows_implemented`; the phase was blocked by `FE-P8-ACTIVATION-BLOCKER-01` in `tools/frontend_phase_registry.json`.

This plan now records FE-P8 final completion on 2026-06-12. Sprint 2 closed FE-U-P8-01, Sprint 3 closed FE-I-P8-01, Sprint 4 closed FE-B-P8-01, Sprint 5 closed FE-E-P8-01, and Sprint 6 closed FE-V-P8-01 plus FE-A11Y-P8-01. Sprint 7 promoted FE-P8 to `active` with `row_rollup_state` `active_green`, regenerated ledgers through Make, reran row-owned targets under the final digest set, passed drift/shape/policy checks, passed full `make check`, and passed retained-run finalization with `RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246`.

FE-P8 closure is anchored to direct current evidence, row accounting, freshness validation, registry promotion, and Make-regenerated ledgers agreeing for every required row. Broad `make check`, generated ledgers, retained old artifacts, visual snapshot filenames, fixture-registry status, and support-only checks still must not close FE-P8 rows by themselves.

### Sprint 2 Executed Status

Sprint 2 implemented the unit/query-state foundation for FE-U-P8-01 on 2026-06-11. The product-conformance scope is deterministic frontend unit behavior for sort/filter/group query compilation, saved-view `query_json` and `layout_json` serialization, active chips, and presentation-only group rows.

Implementation support, design-direction, browser integration, E2E persistence, visual, accessibility, and generated-artifact work remain out of scope for Sprint 2. No generated protocol files were edited.

| Field | Sprint 2 result |
| --- | --- |
| Implemented row | FE-U-P8-01 only. |
| Registry state | FE-P8 remains `planned`; `row_rollup_state` is `partially_implemented`; `FE-P8-ACTIVATION-BLOCKER-01` remains active. |
| Row accounting evidence | `.cartulary/test-results/20260611T204024Z-p40590/frontend-unit/frontend-row-accounting.json` includes FE-U-P8-01 with `closure_status` `closed`. |
| Unit scenario titles | `FE-U-P8-01 compiles query requests with schema field keys and capability metadata`; `FE-U-P8-01 serializes saved-view query_json with canonical empty arrays and omitted inactive grouping`; `FE-U-P8-01 serializes saved-view layout_json as portable schema field-key state`; `FE-U-P8-01 renders active filter chips and grouping controls by field key`; `FE-U-P8-01 keeps grouped presentation rows out of mutation-capable anchors`. |
| Owner update | FE-U-P8-01 owner metadata now includes Core 01 Section 3.3.5.2 for `REQ-01-142` and `REQ-01-143`. |
| Generated ledger | Regenerated through `make phase-ledgers`; `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` was not hand-edited. |
| Non-claims | No saved-view route persistence, reload behavior, browser command closure, visual readiness, accessibility readiness, Core 05 publication, or full FE-P8 completion. |

### Sprint 3 Executed Status

Sprint 3 implemented the saved-view integration boundary for FE-I-P8-01 on 2026-06-12. The product-conformance scope is saved-view loading, active-surface saved-view selection, saved-view identity preservation, normalized query/layout persistence, create/update/duplicate/delete UI, private/shared/system mutability, and user-home/incident-default workbook preference calls through public frontend contracts.

Sprint 3 does not claim stateful reload replay, broad browser command helper closure, visual readiness, accessibility readiness, Core 05 publication evidence, or full FE-P8 activation. No generated protocol files were edited. The generated FE-P8 coverage ledger was regenerated through `make phase-ledgers`.

| Field | Sprint 3 result |
| --- | --- |
| Implemented row | FE-I-P8-01 only. |
| Registry state | FE-P8 remains `planned`; `row_rollup_state` is `partially_implemented`; `FE-P8-ACTIVATION-BLOCKER-01` remains active. |
| Row accounting evidence | `.cartulary/test-results/20260612T021350Z-p95057/frontend-unit/frontend-row-accounting.json` and `.cartulary/test-results/20260612T021406Z-p96577/browser-e2e-webserver-backed/frontend-row-accounting.json` include FE-I-P8-01 with `closure_status` `closed` and current registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`. |
| Scenario title | `FE-I-P8-01 Verify saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts.` |
| Implemented behavior | Frontend saved-view resources preserve scope, owner, version, `query_json`, and `layout_json`; selecting a saved view keeps `sheet_ref_kind=saved_view` and `sheet_ref_id=<saved_view_id>`; normalized saved-query state applies to workbook surfaces; create, update, duplicate, delete, set-home, and set-default controls use stable selectors and public routes; system views are read-only while visible views can be duplicated as private copies. |
| Generated ledger | Regenerated through `make phase-ledgers`; `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` was not hand-edited. |
| Non-claims | No FE-B-P8-01 browser-helper closure, FE-E-P8-01 reload/query-replay closure, visual readiness, accessibility readiness, Core 05 publication, or full FE-P8 completion. |

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

## Sprint 1 Repo Baseline

Initial live inspection occurred from repository root `/home/jochi/code/cartulary` on 2026-06-11. The worktree was clean before this plan file was created. This section is retained as Sprint 1 baseline context; the Sprint 2 executed status above supersedes it for FE-U-P8-01, and the Sprint 3 executed status supersedes it for FE-I-P8-01.

### Inspected Inputs

| Source | Live status at inspection | Digest or command anchor |
| --- | --- | --- |
| `tools/frontend_phase_registry.json` | Present. FE-P8 is `planned`, `no_rows_implemented`, blocked by `FE-P8-ACTIVATION-BLOCKER-01`. | File SHA256 `46d57e203433102aa9fe6ffbe58ce83338890306de417decd0b56b2694094bcc`. FE-P8 manifest digest `c2de1d13c71bfe08cac8c2ffc4b33c68b45fb697d6a34df2da791752e621d4ce`; FE-P8 ledger digest `d18408e2b4dec27fe959f4338077c61cc5473e83f99eb5e1d6c2612404fbe8e1`; FE-P8 evidence freshness digest `d93df1d73b4d31cbbb926dc939e9c45dfd7f18684c682ea904bfb5dc49faa39a`. |
| `tools/frontend_phase_maps/fe_p8_test_map.json` | Present. Six FE-P8 rows, all `claim_status` `blocked`. | SHA256 `c2de1d13c71bfe08cac8c2ffc4b33c68b45fb697d6a34df2da791752e621d4ce`. |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` | Present. Generated downstream ledger reports FE-P8 `planned`, `no_rows_implemented`, six blocked rows. | SHA256 `d18408e2b4dec27fe959f4338077c61cc5473e83f99eb5e1d6c2612404fbe8e1`. |
| `tools/frontend_visual_fixture_registry.json` | Present. FE-P8 owner fixtures are current for grouped result, group/tree row, and empty successful query; registry status alone cannot close FE-V-P8-01. | SHA256 `293c7f3b5211ee6a1ef5461f83d4c119240d16d2837eecb358b1e836a067ab3d`. |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | Present. FE-P8 guide scope matches saved views, sorting, filtering, grouping, layout state, active chips, startup/default-surface UI, group rows, and query-control persistence. | SHA256 `3750847aac8805eb59a1f93430de24a036665241aa7ad3eb6dfb13dcb349b0b5`. |
| `docs/testing-harness-nlspec.md` | Present. Harness mechanics only; frontend registry/maps/ledgers/row-accounting are implementation-readiness artifacts and do not define product behavior. | SHA256 `931aa13edf01869edc9a23ff93cdd58a6ab4ccb63ece76670067727444d96dd2`. |
| `docs/opentelemetry-instrumentation-nlspec.md` | Present. Telemetry-only NLSpec; no FE-P8 product behavior authority. | SHA256 `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451`. |
| Core 00 through Core 05 | Present under `docs/spec/`. Core 00 through Core 04 own product behavior; Core 05 remains publication-only. | Core 00 `e3b2e5e9ed4f47d29694612d571f3255437a9a1acbceb31fc38d9229756a682f`; Core 01 `1c55b261681c59e948356d8f80e2d3f5ab8936d33db5742d18a31f701a81bac9`; Core 02 `bb92665e26804b8c465d961fdef39b78e3f07c389a26a6478e9a210ce393d3fa`; Core 03 `fb561f66e61cf75e777a8c1c4d618d1064ca3b36e8d02435e85131ce631f5b10`; Core 04 `ab4d03966850879625141165d7902f108cfe989914e2b01ed42e2ff7968f6da1`; Core 05 `ee2f572430b75b41ccd20d4dede9c72251b3a4432db2ccf525bec9415da7ef89`. |
| `docs/domain.md` | Present. Vocabulary and concept-boundary reference only. | SHA256 `c461f865e5a0524865e691661f9049ceb9031226124aec60e4014b00847d0e21`. |
| `docs/guides/cartulary_implementation_testing_guide.md` | Present. Shared row/test/harness discipline and browser helper mechanics. | SHA256 `0f37b87228685269cef75e299be4d03a54ad4ee5a51af91ee4f74e3e25d1ae3a`. |
| `docs/guides/cartulary-dev-guide.md` | Present. Frontend package and generated artifact boundaries. | SHA256 `a4b8fb4b9e3b03c905ed276d19a692559ddf1e70396f224f8f8b2a3f68e58776`. |
| `docs/guides/cartulary-ui-ux-design-guide.md` | Present. Design direction only. | SHA256 `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26`. |
| `docs/design.md` | Present. Design token and visual/accessibility direction only. | SHA256 `e28345fac8ba22fc58264454237af209360a84af0c714ff4e1c94c6028d8cd05`. |
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
| FE-V-P8-01 | visual | design_direction | `browser-e2e-visual` | `implemented/closed` | Closed by final `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json` with exact FE-V-P8-01 scenario and current visual goldens. |
| FE-A11Y-P8-01 | accessibility | design_direction | `browser-e2e-a11y` | `implemented/closed` | Closed by final `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json` and `accessibility/frontend-accessibility-summary.json`; preflight remains support-only smoke. |

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
| `.cartulary/test-results/20260611T193506Z-p52331/frontend-unit/frontend-row-accounting.json` | `target_status` `pass`, current guide digest `3750847aac8805eb59a1f93430de24a036665241aa7ad3eb6dfb13dcb349b0b5` and registry digest `46d57e203433102aa9fe6ffbe58ce83338890306de417decd0b56b2694094bcc`, but `row_results` contains no FE-P8 rows. | Must not close FE-U-P8-01 or FE-I-P8-01. |
| `.cartulary/test-results/20260611T193506Z-p52331/browser-e2e-webserver-backed/frontend-row-accounting.json` | `target_status` `pass`, current guide digest and registry digest, but `row_results` contains no FE-P8 rows. | Must not close FE-I-P8-01, FE-B-P8-01, or FE-E-P8-01. |
| `.cartulary/test-results/20260611T193506Z-p52331/browser-e2e-stateful/frontend-row-accounting.json` | `target_status` `pass`, current guide digest and registry digest, but `row_results` contains no FE-P8 rows. | Must not close FE-E-P8-01. |
| `.cartulary/test-results/20260610T191949Z-p14292/browser-e2e-visual/frontend-row-accounting.json` | `target_status` `pass`, but guide digest `7bfa510b010f7696ab79d174ea23489e28045b7924743d16d9990e2ef8e22339` and registry digest `4ad7f3f5fa9dd7ecd4865375e2c47940e089d17b5d4bf00297621842948782e8` are stale relative to current inspected digests; row results omit FE-V-P8-01. | Must not close FE-V-P8-01. |
| `.cartulary/test-results/20260610T171433Z-p14284/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json` | `status` `pass` for FE-A11Y-P8-01 scenario, but phase row still records `claim_status` `blocked`, `required_for_closure` `false`, and the target root has no `frontend-row-accounting.json`. | May inform readiness; must not close FE-A11Y-P8-01. |
| `.cartulary/test-results/20260611T193506Z-p52331/check/target-summary.json` | `status` `pass`. It is a broad retained check run. | Must not close FE-P8 rows because current FE-P8 row-owned accounting is absent. |

### Target Explanation Status

All required explanation commands were run and exited with status 0 before plan creation.

| Command | Result | Live finding |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | FE-P8 is `planned`; all six rows are blocked; planned frontend phases are explainable but not executable as a whole. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | pass | Public fast target; phase coverage includes `phase8`; latest artifact `.cartulary/test-results/20260611T193506Z-p52331/frontend-unit/frontend-unit/phase-summary.json`. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | pass | Public full-gate target; phase coverage includes `phase8`; latest artifact `.cartulary/test-results/20260611T193506Z-p52331/browser-e2e-webserver-backed/target-summary.json`. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | pass | Public full-gate target; phase coverage lists `phase1,phase6`, while the live FE-P8 map names it for FE-E-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-support DETAIL=summary` | pass | `internal_helper`; phase coverage lists `phase1,phase2,phase3`, while the live FE-P8 map names it for FE-B-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | pass | Public full-gate target; phase coverage lists `phase3,phase4,phase5,phase6`, while the live FE-P8 map names it for FE-V-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` | pass | Public full-gate target; phase coverage `none`; not the live mapped FE-P8 accessibility target. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | pass | Public full-gate target; phase coverage `none`; live FE-P8 map names it for FE-A11Y-P8-01. This is a validation-surface finding. |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | pass | Public phase-maintenance target. |
| `make explain-target TARGET=check DETAIL=summary` | pass | Public full-gate target; latest artifact `.cartulary/test-results/20260611T193506Z-p52331/check/target-summary.json`. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | pass | Public helper target requiring `RESULTS_DIR` when retained-run maintenance is intended. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | pass | Public fast target. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | pass | Public fast target. |
| `make explain-target TARGET=phase-ledgers DETAIL=summary` | pass | Public helper target for regenerating generated ledgers only after owner map or registry edits. |
| `make explain-target TARGET=lint-markdown DETAIL=summary` | pass | Public fast target for authored Markdown changes. |
| `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | Current task guide narrows Sprint 1 source validation to `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` and `make phase-ledger-drift`. |

## Sprint 1 Source Limits

This source-limit list records the Sprint 1 audit baseline plus its final Sprint 7 disposition. All FE-P8 implementation blockers below are resolved by the final evidence roots listed in the FE-P9 handoff; retained stale roots remain context only.

1. Resolved. FE-P8 is active, `row_rollup_state` is `active_green`, and `FE-P8-ACTIVATION-BLOCKER-01` is cleared.
2. Resolved. FE-U-P8-01, FE-I-P8-01, FE-B-P8-01, FE-E-P8-01, FE-V-P8-01, and FE-A11Y-P8-01 are implemented and closed under current map authority.
3. Resolved for FE-P8 closure. Row claims use exact live map IDs only; broader guide language is not used to expand product claims.
4. Resolved. Generated FE-P8 ledgers were regenerated through `make phase-ledgers` and validated by drift/shape checks; ledgers remain downstream and are not independent closure evidence.
5. Resolved. Current FE-P8 product row-owned accounting exists for all product rows under the final digest set.
6. Resolved. FE-V-P8-01 was recaptured through current `browser-e2e-visual` row accounting.
7. Resolved. FE-A11Y-P8-01 closes through current `browser-e2e-a11y` row accounting and `cartulary.frontend_accessibility_summary.v2`; preflight remains support-only smoke.
8. Resolved as a target-surface limitation. `browser-e2e-support` and `browser-e2e-stateful` now advertise Phase 8 in `make explain-target`; `browser-e2e-visual` and `browser-e2e-a11y` still report base-phase-oriented `phase_coverage`, but direct frontend row accounting from the FE-P8 map is the closure authority for those frontend readiness rows.
9. Resolved. FE-V-P8-01 visual row accounting captures saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query.
10. Resolved. FE-P8 row-owned scenarios now exist for all six rows; FE-P8-adjacent tests without row ownership remain supplemental context only.
11. The plan cannot claim public route behavior from frontend mocks. Saved-view persistence, startup/default persistence, and query replay must be proven through `/api/v1/` route evidence in mapped browser/E2E targets.
12. The plan cannot claim Core 05 publication, benchmark, visual-publication, fixture-sensitive publication, or claim-publication readiness. No explicit claim-bearing FE-P8 metadata was inspected.
13. Resolved. FE-A11Y-P8-01 support/design IDs were reconciled through FE-P8 owner map and frontend guide updates to current `D-AC-050` and `D-AC-051`; the generated ledgers were regenerated through Make.

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
| FE-U-P8-01 | unit | product_conformance | Core 01 Sections 3.3.4 and 3.3.5.2; Core 03 Section 14 | REQ-01-035, REQ-01-038, REQ-01-047, REQ-01-142, REQ-01-143, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-124, AC-127, AC-184, AC-185, AC-231, AC-238, AC-243, AC-359, AC-361, AC-363, AC-364, AC-372, AC-375, AC-387 | `make frontend-unit`; exact FE-U-P8-01 unit scenarios added and mapped. | implemented | Current `frontend-row-accounting.json` from `frontend-unit` at `.cartulary/test-results/20260611T204024Z-p40590/frontend-unit/frontend-row-accounting.json` has FE-U-P8-01 `closure_status` `closed`, current map/registry/guide digests, passing unit identity, and closure eligibility. | None for FE-U-P8-01 after Sprint 2. | Does not claim saved-view persistence, browser behavior, visual readiness, a11y readiness, or Core 05 publication. |
| FE-I-P8-01 | integration | product_conformance | Core 01 Section 3.3.5.2; Core 03 Section 3 | REQ-01-138, REQ-01-151, REQ-03-012, REQ-03-032 | AC-146, AC-153, AC-231, AC-233, AC-360 | `make frontend-unit`; `make browser-e2e-webserver-backed`; exact FE-I-P8-01 integration scenario added and mapped. | implemented | Current row accounting from mapped targets proves saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts. Unit evidence: `.cartulary/test-results/20260612T021350Z-p95057/frontend-unit/frontend-row-accounting.json`. Browser evidence: `.cartulary/test-results/20260612T021406Z-p96577/browser-e2e-webserver-backed/frontend-row-accounting.json`. | None for FE-I-P8-01 after Sprint 3. | Does not claim E2E reload persistence, browser-helper closure, visual readiness, a11y readiness, or Core 05 publication. |
| FE-B-P8-01 | browser_integration | product_conformance | Core 03 Section 3; Core 03 Section 14; implementation guide Section 14.9A | REQ-03-012, REQ-03-032, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-146, AC-153, AC-231, AC-233, AC-360, AC-363, AC-364 | `make browser-e2e-webserver-backed`; `make browser-e2e-support`; exact FE-B-P8-01 browser helper evidence added and mapped. | implemented | Current row accounting with scenario title `FE-B-P8-01 Verify browser command helpers for sort, filter, group, active chips, layout persistence, group expand-collapse, and startup/default surface UI.` closes in `.cartulary/test-results/20260612T031036Z-p73877/browser-e2e-webserver-backed/frontend-row-accounting.json` and `.cartulary/test-results/20260612T031248Z-p94554/browser-e2e-support/frontend-row-accounting.json`. | None for FE-B-P8-01 after Sprint 4. | Does not claim route persistence after reload, visual readiness, a11y readiness, or Core 05 publication. |
| FE-E-P8-01 | e2e | product_conformance | Core 01 Section 3.3.4; Core 01 Section 3.3.5.2; Core 03 Section 3; Core 03 Section 14 | REQ-01-035, REQ-01-038, REQ-01-047, REQ-01-138, REQ-01-151, REQ-03-012, REQ-03-032, REQ-03-223, REQ-03-235 | AC-013, AC-014, AC-024, AC-026, AC-044, AC-047, AC-124, AC-127, AC-146, AC-153, AC-184, AC-185, AC-231, AC-233, AC-238, AC-243, AC-359, AC-361, AC-363, AC-364, AC-372, AC-375, AC-387 | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; exact FE-E-P8-01 public-route reload/replay evidence added and mapped. | implemented | Current row accounting with scenario title `FE-E-P8-01 Verify saved-view persistence, default/startup surface persistence, and query replay through /api/v1/ after reload.` closes in `.cartulary/test-results/20260612T032712Z-p18527/browser-e2e-webserver-backed/frontend-row-accounting.json` and `.cartulary/test-results/20260612T032959Z-p32522/browser-e2e-stateful/frontend-row-accounting.json`. | None for FE-E-P8-01 after Sprint 5. | Does not claim visual readiness, a11y readiness, rollback readiness, evidence handles, or Core 05 publication. |
| FE-V-P8-01 | visual | design_direction | UI/UX guide Sections 7, 10.5, 13; visual golden guide Sections 2, 3, 5 | None in live map | R2-AC-027, R2-AC-032, R2-AC-051, R2-AC-054, R2-AC-073, R2-AC-079, R2-RDG-AC-001, R2-RDG-AC-010 | `make browser-e2e-visual`; exact FE-V-P8-01 visual readiness scenario added and mapped. | implemented | Current visual row accounting `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json` closes FE-V-P8-01 as design-direction evidence with saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query fixtures. | None for FE-V-P8-01 after Sprint 7 final refresh. | Does not claim product conformance, Core 05 visual publication, or row closure for product rows. |
| FE-A11Y-P8-01 | accessibility | design_direction | UI/UX guide Sections 7, 10.5, 14 | None in live map | R2-AC-027, R2-AC-032, R2-AC-051, R2-AC-054, R2-AC-080, R2-AC-086, D-AC-050, D-AC-051 | `make browser-e2e-a11y`; exact FE-A11Y-P8-01 accessibility scenario added and mapped; preflight remains support-only smoke. | implemented | Current accessibility row accounting `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json` plus `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` closes FE-A11Y-P8-01 as design-direction evidence. | None for FE-A11Y-P8-01 after Sprint 7 final refresh. | Does not claim product conformance, full a11y suite closure beyond mapped readiness, or Core 05 publication. |

## Evidence Layer Matrix

| Row ID | Target(s) | Evidence class | Row-accounting artifact family | Closure predicate | Non-claim boundary |
| --- | --- | --- | --- | --- | --- |
| FE-U-P8-01 | `frontend-unit` | product_conformance | `.cartulary/test-results/<run-id>/frontend-unit/frontend-row-accounting.json` and target/phase summaries | Row appears with current digests, mapped command ID, passing unit identity, `claim_status` implemented or map-promoted equivalent, no blockers, and closure eligibility. | Unit evidence alone does not prove browser persistence, visual readiness, accessibility readiness, or publication. |
| FE-I-P8-01 | `frontend-unit`, `browser-e2e-webserver-backed` | product_conformance | Mapped frontend-unit and webserver-backed row accounting plus public route/browser artifacts when required by scenario | Row appears in current mapped artifacts with saved-view create/update/select/default UI evidence through active surface scope and public saved-view/workbook-preference contracts. | Integration evidence does not prove reload replay unless FE-E-P8-01 closes. |
| FE-B-P8-01 | `browser-e2e-webserver-backed`, `browser-e2e-support` | product_conformance | Webserver-backed row accounting and support-helper accounting/summaries | Row appears with exact scenario title and browser command helper coverage for sort, filter, group, chips, layout persistence, group expand/collapse, and startup/default controls. | Support-helper evidence cannot independently close product conformance without mapped primary row accounting. |
| FE-E-P8-01 | `browser-e2e-webserver-backed`, `browser-e2e-stateful` | product_conformance | Browser row accounting, route trace artifacts, target summaries, and retained run root | Row appears with exact scenario title and public `/api/v1/` route evidence proving saved-view persistence, default/startup persistence, and query replay after reload. | E2E product evidence does not promote visual/a11y design-direction rows. |
| FE-V-P8-01 | `browser-e2e-visual` | design_direction | Visual frontend row accounting, fixture registry linkage, screenshot/golden artifacts, target summaries | Row appears with current digests and captured fixtures for every mapped visual state; fixture registry is current; no stale digest or missing fixture blockers remain. | Visual evidence must remain design-direction and cannot prove product conformance or Core 05 publication. |
| FE-A11Y-P8-01 | `browser-e2e-a11y` | design_direction | Accessibility frontend row accounting, `cartulary.frontend_accessibility_summary.v2`, target summaries | Row appears with current accepted accessibility evidence for keyboard reachability and announced states; no blocked-row or missing-accounting condition remains. | Accessibility evidence must not be used as product conformance; `browser-e2e-a11y-preflight` remains support-only smoke. |

## Dependencies And Prerequisites

| Dependency | Required FE-P8 posture | Current inspected status |
| --- | --- | --- |
| FE-P0 through FE-P7 | Must remain green because FE-P8 depends on all earlier frontend phases. | Registry reports FE-P0 through FE-P7 `active_green`. FE-P7 plan records closed rows, but that is handoff context only. |
| Generated contracts | Must be current before FE-P8 row closure; generated roots must not be edited by hand. | `packages/protocol-ts/src/generated/index.ts` is generated. No generated file was edited for this plan. |
| View contracts | Must expose stable `view_schema_id`, sortable/filterable/groupable `field_key`, capability metadata, default sort, and synthetic predicates. | `packages/view-contracts/src/index.ts` inspected as contract-derived metadata surface. |
| Public query routes | Must support view-query `sort[]`, `filters[]`, and `group_by` through `/api/v1/incidents/{incident_id}/views/{view_schema_id}/query`. | Route behavior must be proven by mapped browser/E2E artifacts, not plan text. |
| Saved-view routes | Must support list/create/update/select/default evidence under incident and active-surface scope. | `apps/web/e2e/phase8.workbook.spec.ts` contains route-context tests, but current FE-P8 row accounting is absent. |
| Workbook preference routes | Must support user home sheet and incident default sheet persistence through public contracts. | Closed by current FE-I-P8-01 and FE-E-P8-01 row-owned browser evidence through public routes. |
| Grid adapter | Must keep direct `react-data-grid` dependency inside `/packages/grid-adapter`, protect identity anchors, and make group rows presentation-only. | `packages/grid-adapter/src/core.ts` and `src/index.tsx` inspected. |
| Selectors/test IDs | Must use stable generated builders for saved-view selector, chips, group rows, cells, and controls. | `packages/ui-contracts/src/index.ts` inspected. |
| Browser helpers | Must cover sort, filter chips, group change, expand/collapse, anchors, paste/fill-down, and stable selection. | `packages/test-utils/src/index.ts` inspected. |
| Visual fixtures | Must capture all FE-V-P8-01 mapped states with current digests and row accounting. | Closed by current FE-V-P8-01 `browser-e2e-visual` row accounting and FE-P8 visual goldens. |
| Accessibility harness | Must prove keyboard reachability and announced state through mapped FE-A11Y-P8-01 evidence. | Closed by current FE-A11Y-P8-01 `browser-e2e-a11y` row accounting and accessibility summary; preflight remains support-only smoke. |
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
- [x] Reconcile guide range language and live map exact REQ IDs before expanding any row claim. Sprint 6 resolved the FE-A11Y design-ID source issue by replacing the missing FE-P8 `D-AC-009`/irrelevant `D-AC-012` references with owner-mapped `D-AC-050` and `D-AC-051`.
- [x] Promote implemented FE-P8 map rows only after direct current evidence exists; all six FE-P8 rows are promoted, FE-P8 registry activation is owner-authorized, and final post-activation evidence refresh is complete.
- [x] Implement or validate FE-U-P8-01 query-state compilation using stable `field_key` and capability metadata.
- [x] Implement or validate FE-I-P8-01 saved-view create/update/select/default UI through public contracts.
- [x] Implement or validate FE-B-P8-01 browser command helpers and group-row behavior. Closed on 2026-06-12 by current `browser-e2e-webserver-backed` and `browser-e2e-support` row accounting after Make-regenerated ledgers and import-boundary validation.
- [x] Implement or validate FE-E-P8-01 saved-view persistence, default/startup persistence, reload behavior, and query replay through `/api/v1/`. Closed on 2026-06-12 by current `browser-e2e-webserver-backed` and `browser-e2e-stateful` row accounting after Make-regenerated ledgers and a replay-shaped public query assertion fix.
- [x] Implement and validate FE-V-P8-01 visual fixtures without promoting design-direction evidence to product conformance.
- [x] Implement and validate FE-A11Y-P8-01 keyboard and announcement evidence without promoting accessibility evidence to product conformance.
- [x] Run mapped unit and integration/browser validation commands for implemented rows FE-U-P8-01 and FE-I-P8-01.
- [x] Run `make phase-ledger-drift`, `make generated-artifact-policy-check`, and `make json-shape-check` for Sprint 3 map/registry/ledger freshness.
- [x] Re-run the required Sprint 4-7 target explanations on 2026-06-12 before relying on validation targets; target-surface gaps are recorded as resolved limitations and are not row-closure evidence.
- [x] Run `make check` only when current post-activation row-owned evidence and drift checks justify the full gate. Passed at `.cartulary/test-results/20260612T044428Z-p24246` after row-owned evidence refresh, drift/shape/policy checks, formatting, and lint remediation.
- [x] Run `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` after a successful full-check run is available. Final pass: `.cartulary/test-results/20260612T045247Z-p22120` with `RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246`.
- [x] Update the blocker register with every unclosed row and exact source condition.
- [x] Prepare FE-P9 handoff with final FE-P8 registry status, evidence roots, blockers, drift/finalization outcomes, and strict non-claims.

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

#### Sprint 1 Scope, Non-Scope, And Assumptions

Sprint 1 is a documentation and implementation-support slice. It closes only the readiness/source-alignment baseline for FE-P8 by recording live authority, live source gaps, deterministic blocker handling, validation commands, and skipped-check rationale in this plan. It does not implement workbook behavior, promote phase rows, activate FE-P8, recapture visual fixtures, close accessibility rows, edit generated outputs, or update harness code.

Assumptions:

1. Core 00 through Core 04 own product conformance; Core 05 remains inactive unless explicit claim-bearing publication metadata is added.
2. `docs/testing-harness-nlspec.md` owns harness mechanics and retained-artifact rules only.
3. `tools/frontend_phase_maps/fe_p8_test_map.json` and `tools/frontend_phase_registry.json` are authored source inputs; `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` is generated downstream.
4. Retained runs are usable only as context unless their row accounting has current guide/registry/map digests, direct FE-P8 row IDs, and map-authorized closure status.
5. Existing FE-P8-adjacent tests remain candidate implementation context until exact FE-P8 row-owned accounting is produced.

Non-scope:

1. No product-conformance changes.
2. No harness behavior changes.
3. No FE-P8 map promotion, registry activation, or row closure.
4. No generated-ledger hand edits.
5. No Core 05 publication, visual-publication, benchmark, rollback, evidence-handle, or FE-P9 readiness claim.

#### Sprint 1 Readiness And Source-Alignment Gaps

| Gap ID | Source | Live condition | Deterministic handling |
| --- | --- | --- | --- |
| `FE-P8-S1-GAP-01` | `tools/frontend_phase_registry.json` | Resolved in Sprint 7. FE-P8 is `active`, row rollup is `active_green`, and `FE-P8-ACTIVATION-BLOCKER-01` is cleared. | No longer blocks FE-P8 closure. |
| `FE-P8-S1-GAP-02` | `tools/frontend_phase_maps/fe_p8_test_map.json` | Resolved in Sprint 7. All six FE-P8 rows are `claim_status="implemented"` with current row-owned closure evidence. | No longer blocks FE-P8 closure. |
| `FE-P8-S1-GAP-03` | Retained product accounting | Latest current-digest product row accounting under `.cartulary/test-results/20260611T193506Z-p52331` contains no FE-P8 rows. | Treat broad target pass and broad `make check` pass as context only. |
| `FE-P8-S1-GAP-04` | Visual retained artifact | `.cartulary/test-results/20260610T191949Z-p14292/browser-e2e-visual/frontend-row-accounting.json` has stale guide/registry digests and no FE-V-P8-01 row. | Do not close visual readiness; recapture through mapped visual row when in scope. |
| `FE-P8-S1-GAP-05` | Accessibility preflight | `.cartulary/test-results/20260610T171433Z-p14284/.../frontend-accessibility-preflight-summary.json` passes blocked-row smoke but records FE-A11Y-P8-01 as blocked and has no root row accounting. | Treat as diagnostic readiness context only. |
| `FE-P8-S1-GAP-06` | Target explanations | Sprint 1 target explanations did not advertise phase8 coverage for all mapped browser targets. Sprint 7 reconfirmed `browser-e2e-support` and `browser-e2e-stateful` now advertise Phase 8; `browser-e2e-visual` and `browser-e2e-a11y` remain base-phase-oriented in `phase_coverage` but emit frontend row accounting from FE-P8 maps. | Resolved as a target-surface limitation; final closure relies on current frontend row accounting, not target explanation text. |
| `FE-P8-S1-GAP-07` | `docs/design.md` and FE-P8 accessibility map | Resolved in Sprint 6/7. FE-P8 accessibility support/design IDs now use current owner-mapped `D-AC-050` and `D-AC-051`. | No longer blocks FE-A11Y-P8-01 readiness. |
| `FE-P8-S1-GAP-08` | Fixture registry | FE-P8-owned fixtures cover grouped result, group/tree row, and empty successful query only. | Do not use fixture status to prove saved-view selector, active chips, or default/startup state indicator readiness. |

#### Sprint 1 Work Breakdown

| Task | Owner source | Expected file changes | Validation target | Acceptance evidence |
| --- | --- | --- | --- | --- |
| Refresh authority and digest inventory | Core 00; frontend guide; harness NLSpec | `FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` only | `sha256sum`; `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | Current source table and FE-P8 planned/blocked status. |
| Record source gaps and blockers | Frontend guide; registry/map; harness NLSpec | Same | `rg`; `jq`; target explanations | Gap table with no inferred row closure. |
| Add `D-AC-009` blocker | `docs/design.md`; FE-P8 map | Same | `rg "D-AC-009"` | Explicit unresolved design-ID blocker for FE-A11Y-P8-01. |
| Record validation evidence | Harness NLSpec; Make target surface | Same | Commands listed in Sprint 1 validation ledger | Pass/fail ledger with run roots. |
| State skipped checks and non-claims | AGENTS procedure; harness NLSpec | Same | `git status --short` | No generated, product, full-check, or finalizer evidence is claimed. |

#### Sprint 1 Deterministic Source Handling

1. Missing owner references, stale digests, contradictory source status, or unverifiable retained artifacts create blockers rather than inferred closure.
2. Generated ledgers are regenerated only through `make phase-ledgers` after an authored map or registry edit; Sprint 1 does not edit those inputs, so no regeneration is expected.
3. Retained artifacts older than current guide/registry/map digests remain diagnostic context only.
4. Broad `make check` success cannot close FE-P8 rows without current FE-P8 row-owned accounting.
5. Design-direction rows cannot be promoted to product conformance; accessibility preflight cannot close an implemented accessibility row without map authority and accepted accounting.
6. FE-P5's `D-AC-009` remediation is historical context only. FE-P8 must resolve its own support/design IDs through the FE-P8 owner map and current design source.

#### Sprint 1 Validation Ledger

| Command | Post-edit result | Evidence |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | FE-P8 reported `planned`; all six rows reported `blocked`; whole planned phase reported non-executable. |
| `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | Task guide selected `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` and `make phase-ledger-drift`. |
| Mapped and maintenance `make explain-target TARGET=<target> DETAIL=summary` commands | pass | Reconfirmed mapped FE-P8 targets, phase-maintenance targets, `lint-markdown`, `check`, and `agent-finalize` are explainable; validation-surface gaps found during Sprint 1 were later reconciled or recorded as resolved target-surface limitations. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260611T195739Z-p16131/phase-ledger-drift/tool-run-summary.json`. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260611T195739Z-p16153/generated-artifact-policy-check/tool-run-summary.json`. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260611T195739Z-p16167/json-shape-check/tool-run-summary.json`. |
| `make lint-markdown` | pass | `.cartulary/test-results/20260611T195739Z-p16115/lint-markdown/tool-run-summary.json`. |
| `git status --short` | pass | Only `FRONTEND_PHASE8_IMPLEMENTATION_PLAN.md` is modified; ignored run artifacts were produced by validation. |

`make phase-ledgers`, `make generate-drift`, `make check`, and `make agent-finalize RESULTS_DIR=...` were intentionally skipped for Sprint 1 because no phase map, registry, generated owner input, product behavior, or retained full-check completion evidence changed.

### Sprint 2: Unit And Query-State Foundation

| Field | Plan |
| --- | --- |
| Objective | Implement or validate deterministic query-state compilation for sort, filters, groups, layout state, and active chips. |
| Implementation targets | `apps/web/src/workbookQuery.ts`, `apps/web/src/WorkbookGridControls.tsx`, `packages/view-contracts/src/index.ts`, `packages/ui-contracts/src/index.ts`, `apps/web/src/WorkbookShell.phase8.query.test.tsx`. |
| Validation commands | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make explain-target TARGET=frontend-unit DETAIL=summary`. |
| Expected evidence | FE-U-P8-01 row-owned `frontend-unit` accounting with current digests; unit names or scenario identities proving `sort[]`, `filters[]`, `group_by`, active chips, layout state serialization, and capability filtering use stable `field_key`. |
| Blockers | Visible-label keys, DOM-order keys, row-index keys, unsupported field operations, non-sortable field sorting, group rows with writable IDs, stale map digests, or missing FE-U-P8-01 accounting. |
| Strict non-claims | Unit closure must not claim saved-view route persistence, reload behavior, visual readiness, or accessibility readiness. |

#### Sprint 2 Execution Ledger

| Command | Result | Evidence |
| --- | --- | --- |
| `make frontend-unit` | pass | `.cartulary/test-results/20260611T204024Z-p40590/frontend-unit/frontend-row-accounting.json`; FE-U-P8-01 `closure_status` is `closed`. |
| Row-accounting inspection | pass | Same artifact contains all five exact FE-U-P8-01 scenario titles and current guide, registry, and FE-P8 map references. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260611T204050Z-p42146`. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260611T204101Z-p42534`. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260611T204106Z-p42887`. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260611T204110Z-p43218`. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260611T204114Z-p43396`. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | pass | Latest artifact `.cartulary/test-results/20260611T204024Z-p40590/frontend-unit/frontend-unit/phase-summary.json`; phase coverage includes `phase8`. |

`make check` was skipped because Sprint 2 did not expand beyond the unit/map/docs slice. `make agent-finalize` was skipped because no retained successful full-check `RESULTS_DIR` was used.

### Sprint 3: Saved-View Integration

| Field | Plan |
| --- | --- |
| Objective | Implement or validate saved-view create/update/select/default UI using active surface scope and public saved-view/workbook-preference contracts. |
| Implementation targets | `apps/web/src/WorkbookShell.tsx`, `apps/web/src/workbookStartup.ts`, `apps/web/src/workbookSurfaceRegistry.ts`, `apps/web/e2e/phase8.workbook.spec.ts`, `packages/protocol-ts`, `packages/view-contracts`, `packages/ui-contracts`. |
| Validation commands | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make frontend-typecheck`; `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`. |
| Expected evidence | FE-I-P8-01 row accounting from mapped targets with public route evidence for active-surface saved-view scope, saved-view selection, owner query/layout persistence, user home sheet, and incident default sheet where authorized. |
| Blockers | Saved-view state proved only through mocks, saved views not scoped by active `view_schema_id`, query/layout JSON containing visible labels or row IDs, default/startup persistence without public route evidence, or missing FE-I-P8-01 accounting. |
| Strict non-claims | This sprint must not claim reload replay unless FE-E-P8-01 closes and must not claim visual/a11y readiness. |

#### Sprint 3 Execution Ledger

| Command | Result | Evidence |
| --- | --- | --- |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T021350Z-p95057/frontend-unit/frontend-row-accounting.json`; FE-I-P8-01 `closure_status` is `closed`. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260612T021406Z-p96577/browser-e2e-webserver-backed/frontend-row-accounting.json`; FE-I-P8-01 `closure_status` is `closed`. |
| Row-accounting inspection | pass | Both mapped artifacts contain the exact FE-I-P8-01 scenario title and current registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T021549Z-p10092`. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260612T021549Z-p10125`. |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260612T020749Z-p63791`; regenerated FE-P8 ledger from the authored phase map. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260612T021549Z-p10164`. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260612T021341Z-p94675`; registry manifest, ledger, and evidence freshness digests match FE-P8 owner inputs. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260612T021549Z-p10126`. |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | FE-P8 reported `planned`; FE-U-P8-01 and FE-I-P8-01 implemented; FE-B-P8-01, FE-E-P8-01, FE-V-P8-01, and FE-A11Y-P8-01 blocked. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | pass | Latest artifact `.cartulary/test-results/20260612T021350Z-p95057/frontend-unit/frontend-unit/phase-summary.json`; phase coverage includes `phase8`. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | pass | Latest artifact `.cartulary/test-results/20260612T021406Z-p96577/browser-e2e-webserver-backed/target-summary.json`; phase coverage includes `phase8`. |
| `make agent-finalize` | pass | `.cartulary/test-results/20260612T021620Z-p11568`; retained-run reuse was skipped because `RESULTS_DIR` was unset. |

`make check` was skipped because Sprint 3 was a targeted FE-I-P8-01 implementation slice and did not attempt full FE-P8 activation or full-repo release gating.

### Sprint 4: Browser Commands And Group-Row Behavior

| Field | Plan |
| --- | --- |
| Objective | Verify browser command helpers, group expand/collapse, active chips, layout persistence, default/startup controls, and non-writable group rows. |
| Implementation targets | `packages/test-utils/src/index.ts`, `packages/grid-adapter/src/core.ts`, `packages/grid-adapter/src/index.tsx`, `packages/ui-contracts/src/index.ts`, `apps/web/e2e/phase8.workbook.spec.ts`. |
| Validation commands | `make browser-e2e-webserver-backed`; `make browser-e2e-support`; `make frontend-import-boundary-check`; `make explain-target TARGET=browser-e2e-support DETAIL=summary`. |
| Expected evidence | FE-B-P8-01 row accounting with exact scenario title, browser command helper execution, stable selector use, group expand/collapse proof, active chip add/remove proof, and group-row non-mutation proof. |
| Blockers | Direct `react-data-grid` imports outside adapter, helper selectors based on visible labels or vendor coordinates, group rows emitting mutation-capable events, support target evidence without primary accounting, or missing FE-B-P8-01 accounting. |
| Strict non-claims | Browser helper evidence must not claim route persistence after reload unless FE-E-P8-01 closes. |

#### Sprint 4 Pre-Implementation Validation Ledger

| Command | Result | Evidence |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P8` | pass | FE-P8 reported `planned`; FE-U-P8-01 and FE-I-P8-01 implemented; FE-B-P8-01, FE-E-P8-01, FE-V-P8-01, and FE-A11Y-P8-01 blocked. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | pass | Public full-gate target; phase coverage includes `phase8`; latest artifact reported as `.cartulary/test-results/20260612T023614Z-p24083/browser-e2e-webserver-backed/target-summary.json`. |
| `make explain-target TARGET=browser-e2e-support DETAIL=summary` | pass, later reconciled | Sprint 4 startup target surface did not yet advertise Phase 8. Final explanation after owner-map/schedule refresh reports `phase1,phase2,phase3,phase8`, and final support row accounting passed. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | pass, later reconciled | Sprint 4 startup target surface did not yet advertise Phase 8. Final explanation after owner-map/schedule refresh reports `phase1,phase6,phase8`, and final stateful row accounting passed. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | pass with resolved target-surface limitation | Public full-gate target; base `phase_coverage` lists `phase3,phase4,phase5,phase6`, not `phase8`. Final FE-V-P8-01 closure relies on frontend row accounting emitted by the direct visual target. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | pass with resolved target-surface limitation | Public full-gate target; base `phase_coverage` is `none`. Final FE-A11Y-P8-01 closure uses `browser-e2e-a11y`; preflight is support-only smoke. |

No row closure is claimed from these explanation commands. They only establish the current validation surface before Sprint 4 implementation.

#### Sprint 4 Implementation Ledger

| Material change | Files | Current status |
| --- | --- | --- |
| Browser helper commands | `packages/test-utils/src/index.ts`; `packages/test-utils/src/index.test.ts` | Added helper-owned commands for saved-view select/read, saved-view name/scope/create/update, current-view home/default controls, active filter-chip visibility, and presentation-only group-row assertions. Validated by final `make check` and final `make browser-e2e-support`. |
| Group-row behavior | `packages/grid-adapter/src/index.tsx`; `packages/grid-adapter/src/index.test.tsx` | Added explicit `data-grid-row-kind="group"` marking on group presentation rows and adapter assertions that group rows omit `data-grid-record-id`, expose no editable controls, and only expose the expand/collapse button. Validated by final `make check`, browser row accounting, and support row accounting. |
| Functional browser evidence | `apps/web/e2e/phase8.workbook.spec.ts` | Added exact FE-B-P8-01 functional browser scenario covering helper-driven saved-view selection, sort/filter/group, active chip, group expand/collapse, layout/query persistence PATCH, user home/default PUTs, and chip removal. Final `browser-e2e-webserver-backed` row accounting passed under `.cartulary/test-results/20260612T044428Z-p24246`. |
| Browser support evidence | `apps/web/e2e/phase8.workbook.support.spec.ts`; `tools/phase8_test_map.json` | Added supplemental `E-8-SUPPORT-01` browser-support owner input with exact FE-B-P8-01 title. Final `browser-e2e-support` row accounting passed under `.cartulary/test-results/20260612T044835Z-p88067`. |
| FE-B map promotion | `tools/frontend_phase_maps/fe_p8_test_map.json` | Promoted FE-B-P8-01 to `claim_status: implemented`, made its two browser targets required for closure, cleared its row blocker, and closed the row with current row-owned accounting. |

Sprint 4 implementation is closed for FE-B-P8-01 by current `browser-e2e-webserver-backed` and `browser-e2e-support` row accounting. FE-B-P8-01 evidence does not close FE-E reload/query replay, visual readiness, accessibility readiness, or activation.

#### Sprint 4 Validation Attempt Ledger

| Command | Result | Evidence and handling |
| --- | --- | --- |
| `make phase-ledgers` | fail, fixed by later rerun | `.cartulary/test-results/20260612T030335Z-p49874`; failure: `phase8 manifest mismatch: missing=E-8-SUPPORT-01 unexpected=none`. Handling: removed supplemental support ID from `expected_ids` while keeping the supplemental support row, matching existing Phase 3 manifest structure; later ledger generation passed. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T030335Z-p49872/frontend-typecheck/target-summary.json`. This does not close FE-B-P8-01; it only verifies the staged TypeScript surface compiles. |
| `make frontend-unit` | fail, fixed by later rerun | `.cartulary/test-results/20260612T030335Z-p49905`; failure: `Expected active filter chip for timeline.capture_state on cartulary.view.timeline.v1 to be visible`. Handling: updated the helper unit mock locator to expose Playwright-like `isVisible()` for the active-chip helper assertion; later unit validation passed. |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260612T030446Z-p52587`; regenerated `docs/testing/phase8_coverage_ledger.md` and `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md` from owner inputs. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T030446Z-p52586/frontend-typecheck/target-summary.json`. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T030446Z-p52619`; helper, adapter, and FE-P8 unit row accounting checks passed. |
| `make browser-e2e-webserver-backed` | fail, fixed by later rerun | `.cartulary/test-results/20260612T030529Z-p54988`; FE-B-P8-01 functional/support scenario timed out while `gridFilterApplyTestId` click was intercepted by saved-view/default/view-bar elements. Handling: updated `apps/web/src/WorkbookShell.tsx` and `apps/web/src/WorkbookGridControls.tsx` so saved-view and query-control rows wrap instead of overlapping; later webserver-backed validation passed. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T031008Z-p71890/frontend-typecheck/target-summary.json` after the view-bar wrapping fix. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T031008Z-p71964` after the view-bar wrapping fix. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260612T031036Z-p73877/browser-e2e-webserver-backed/frontend-row-accounting.json`; FE-B-P8-01 appears with `target_status` `pass`, command ID `cartulary.harness.command.browser_e2e_webserver_backed.v1`, current guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, current registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, current FE-P8 map digest `8fb839c4ab958ce817eb61ad5e67e11be455d021051010fe3c50f8c0a366a2fd`, exact scenario title, and `closure_status` `closed`. Later final support evidence passed at `.cartulary/test-results/20260612T044835Z-p88067`. |
| `make browser-e2e-support` | pass | `.cartulary/test-results/20260612T031248Z-p94554/browser-e2e-support/frontend-row-accounting.json`; FE-B-P8-01 appears with `target_status` `pass`, command ID `cartulary.harness.command.browser_e2e_support.v1`, current guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, current registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, current FE-P8 map digest `8fb839c4ab958ce817eb61ad5e67e11be455d021051010fe3c50f8c0a366a2fd`, exact scenario title, and `closure_status` `closed`. The phase8 supplemental support summary is `.cartulary/test-results/20260612T031248Z-p94554/browser-e2e-support/browser-e2e-support-phase8-supplemental/phase-summary.json`. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260612T031354Z-p98176`; no direct `react-data-grid` import boundary regression after Sprint 4 changes. |

### Sprint 5: Stateful And E2E Persistence

| Field | Plan |
| --- | --- |
| Objective | Verify saved-view persistence, default/startup surface persistence, reload behavior, and query replay through `/api/v1/`. |
| Implementation targets | `apps/web/src/WorkbookShell.tsx`, `apps/web/src/workbookStartup.ts`, `apps/web/e2e/phase8.workbook.spec.ts`, public saved-view/preference/query routes, generated protocol contracts. |
| Validation commands | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make explain-target TARGET=browser-e2e-stateful DETAIL=summary`. |
| Expected evidence | FE-E-P8-01 row accounting with exact scenario title, public route traces, reload proof, persisted saved-view selection, default/startup surface persistence, and query replay using public `/api/v1/` contracts. |
| Blockers | Persistence asserted without route evidence, state stored only in browser memory, query replay using labels or DOM order, group rows treated as records, stale row-accounting digests, or missing FE-E-P8-01 accounting. |
| Strict non-claims | E2E product evidence must not promote FE-V-P8-01 or FE-A11Y-P8-01 to product conformance. |

#### Sprint 5 Implementation Ledger

| Material change | Files | Current status |
| --- | --- | --- |
| Public-route persistence scenario | `apps/web/e2e/phase8.workbook.spec.ts` | Added exact FE-E-P8-01 browser scenario that persists a saved view from the live UI, verifies saved-view list/create route state, persists user home and incident default preferences, checks `/api/v1/.../workbook-startup`, and proves startup/reload query replay through `/api/v1/.../views/.../query`. After the first failed webserver-backed run captured an initial `{}` warm-up query, the scenario was tightened to wait for the replay-shaped public query body. Current webserver-backed and stateful validation now pass. |
| Stateful owner input | `tools/phase8_test_map.json` | Added authoritative `E-8-05` browser_stateful row with the exact FE-E-P8-01 title so `browser-e2e-stateful` advertises Phase 8 coverage. Generated ledgers were refreshed through Make and stateful validation now passes. |
| FE-E map promotion | `tools/frontend_phase_maps/fe_p8_test_map.json` | Promoted FE-E-P8-01 to `claim_status: implemented`, made `browser-e2e-webserver-backed` and `browser-e2e-stateful` required for closure, cleared its row blocker, and closed the row with current row-owned accounting. |

Sprint 5 is closed for FE-E-P8-01 by current `browser-e2e-webserver-backed` and `browser-e2e-stateful` row accounting. FE-E-P8-01 evidence does not close visual readiness, accessibility readiness, rollback readiness, evidence handles, Core 05 claims, or activation.

#### Sprint 5 Validation Attempt Ledger

| Command | Result | Evidence and handling |
| --- | --- | --- |
| `make phase-ledgers` | fail, fixed by later rerun | `.cartulary/test-results/20260612T031827Z-p99870`; failure: authoritative base Phase 8 stateful evidence for `E-8-05` used the frontend FE-E title and did not include `E-8-05` or `E_8_05`. Handling: split the scenario into a shared helper with an exact FE-E frontend title plus an `E-8-05` wrapper title for the base stateful manifest; later ledger generation passed. |
| `make frontend-typecheck` | pass | Typecheck completed during the failed ledger attempt; latest artifact root will be refreshed after the wrapper-title fix. |
| `make phase-ledgers` | fail, fixed by later rerun | `.cartulary/test-results/20260612T031922Z-p1153`; failure: `phase8 guide mismatch: missing=none unexpected=E-8-05`. Handling: added E-8-05 to the Phase 8 row inventory in `docs/guides/cartulary_implementation_testing_guide.md` as the owner guide row for public-route saved-view/default startup persistence and reload query replay; later ledger generation passed. |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260612T032023Z-p2169`; regenerated base and frontend FE-P8 ledgers from reconciled guide/map inputs. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T032023Z-p2192/frontend-typecheck/target-summary.json`. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | pass | Target surface now lists `phase1,phase6,phase8` with `dependencies=browser_stateful`; this resolves the earlier FE-E stateful target-surface mismatch before validation. |
| `make browser-e2e-webserver-backed` | fail, fixed by later rerun | `.cartulary/test-results/20260612T032223Z-p3395`; `frontend-row-accounting.json` schema `cartulary.frontend_row_accounting.v3` contains FE-E-P8-01 with `target_status` `fail`, `closure_status` `not_closed`, current guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, and current map digests. The exact scenario failed at `apps/web/e2e/phase8.workbook.spec.ts:876`: received public `/api/v1/.../views/.../query` body `{}` before the expected saved-view replay query. Handling: updated the scenario to wait for the replay-shaped public query request, then reran webserver-backed and stateful targets successfully. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T032647Z-p18015/frontend-typecheck/target-summary.json` after the replay-shaped query wait helper was added. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260612T032712Z-p18527`; `frontend-row-accounting.json` schema `cartulary.frontend_row_accounting.v3` includes FE-E-P8-01 with command ID `cartulary.harness.command.browser_e2e_webserver_backed.v1`, phase namespace `frontend`, evidence class `product_conformance`, `target_status` `pass`, `claim_status_at_run` `implemented`, `target_mapping_status` `mapped`, `closure_status` `closed`, exact scenario title, guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, and current phase map digests. |
| `make browser-e2e-stateful` | pass | `.cartulary/test-results/20260612T032959Z-p32522`; `frontend-row-accounting.json` schema `cartulary.frontend_row_accounting.v3` includes FE-E-P8-01 with command ID `cartulary.harness.command.browser_e2e_stateful.v1`, phase namespace `frontend`, evidence class `product_conformance`, `target_status` `pass`, `claim_status_at_run` `implemented`, `target_mapping_status` `mapped`, `closure_status` `closed`, exact scenario title, guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, and current FE-P8 map digest list. |

### Sprint 6: Visual And Accessibility Readiness

| Field | Plan |
| --- | --- |
| Objective | Verify visual and accessibility readiness for FE-P8 while keeping design-direction evidence separate from product conformance. |
| Implementation targets | `apps/web/e2e/workbook.visual.spec.ts`, `apps/web/e2e/workbook.a11y-preflight.spec.ts`, `tools/frontend_visual_fixture_registry.json`, `docs/guides/cartulary_visual_golden_maintenance.md`, `packages/ui-contracts/src/index.ts`. |
| Validation commands | `make browser-e2e-visual`; `make browser-e2e-a11y`; `make browser-e2e-a11y-preflight` as supplemental smoke only; `make explain-target TARGET=browser-e2e-visual DETAIL=summary`; `make explain-target TARGET=browser-e2e-a11y DETAIL=summary`; `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary`. |
| Expected evidence | FE-V-P8-01 visual row accounting and fixture artifacts for saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query; FE-A11Y-P8-01 accessibility evidence for keyboard reachability and announced state. |
| Blockers | Stale visual digests, missing visual states, fixture registry status without row accounting, preflight pass with blocked claim status, no root row accounting when map requires it, or accessibility evidence represented as product conformance. |
| Strict non-claims | Visual and accessibility evidence remain design-direction and must not claim Core 05 publication readiness. |

#### Sprint 6 Implementation Ledger

| Material change | Files | Current status |
| --- | --- | --- |
| Visual readiness scenario | `apps/web/e2e/workbook.visual.spec.ts` | Added exact FE-V-P8-01 visual scenario capturing saved-view selector, active chip, grouped result, group row, default/startup status indicator, and empty successful query. Final `browser-e2e-visual` row accounting passed under `.cartulary/test-results/20260612T044934Z-p91545`. |
| Accessibility readiness scenario | `apps/web/e2e/workbook.a11y.spec.ts`; `apps/web/e2e/workbook.a11y-preflight.spec.ts`; `packages/grid-adapter/src/index.tsx` | Added exact FE-A11Y-P8-01 full-a11y scenario for keyboard reachability, announced state, and contrast summary coverage; retained the matching preflight scenario as supplemental smoke only. Sort headers are real buttons so the sort control is keyboard reachable. Final `browser-e2e-a11y` row accounting passed under `.cartulary/test-results/20260612T045057Z-p3086`; final preflight smoke passed under `.cartulary/test-results/20260612T045159Z-p12967`. |
| Design ID reconciliation and row promotion | `tools/frontend_phase_maps/fe_p8_test_map.json`; `docs/guides/cartulary_frontend_implementation_testing_guide.md` | Promoted FE-V-P8-01 and FE-A11Y-P8-01 as `design_direction` implemented rows, made their mapped targets required for closure, cleared row blockers, replaced missing FE-P8 `D-AC-009` plus irrelevant `D-AC-012` with current `D-AC-050` and `D-AC-051`, and corrected FE-A11Y-P8-01 closure authority from preflight to `make browser-e2e-a11y` per the frontend guide. Make regeneration and drift/shape checks passed. |
| Visual fixture registry | `tools/frontend_visual_fixture_registry.json` | Repointed FE-VFIX-06 to the exact FE-V-P8-01 scenario and expected saved-view/query-controls golden, and added the FE-P8 empty-successful-query golden artifact to FE-VFIX-15. The new PNG artifacts were generated and validated by `make browser-e2e-visual`. |
| Visual golden refresh | `apps/web/e2e/workbook.visual.spec.ts-snapshots/*` | Regenerated toolbar-affected P2-P7/V-* workbook goldens and added `fe-v-p8-01-saved-view-query-controls-linux.png` plus `fe-v-p8-01-empty-successful-query-linux.png` after accepted FE-P8 saved-view toolbar expansion. Final acceptance passed in `make browser-e2e-visual` at `.cartulary/test-results/20260612T044934Z-p91545`. |
| FE-P6 visual fixture stabilization | `apps/web/e2e/workbook.visual.spec.ts` | Added a visual-only `record_id` cell mask to the FE-P6 evidence-access fixture after repeated Make-owned visual runs showed run-varying identifier fragments in that crop. This is support-only stabilization for visual target closure and does not change product UI behavior. |

Sprint 6 visual and accessibility readiness are proven for FE-V-P8-01 and FE-A11Y-P8-01 by current row accounting, and Sprint 7 final closeout completed registry activation, Make-regenerated ledgers, drift/shape checks, full check, and finalization.

#### Sprint 6 Validation Attempt Ledger

| Command | Result | Evidence and handling |
| --- | --- | --- |
| `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` | pass with resolved target-surface limitation | Target exists and is the guide-authorized implemented accessibility closure surface. It still reports base-oriented `phase_coverage: none`, but final FE-P8 closure uses the direct frontend row accounting emitted by `browser-e2e-a11y` per the frontend map and harness NLSpec. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | pass with resolved target-surface limitation | Target remains support-only blocked-row smoke and reports `phase_coverage: none`. It does not close FE-A11Y-P8-01; final closure uses `browser-e2e-a11y` row accounting. |
| `make frontend-typecheck` | failed | Harness static validation rejected the first FE-A11Y-P8-01 promotion because implemented accessibility rows must not close from `browser-e2e-a11y-preflight`. This failure is not closure evidence; map and guide are being corrected to `browser-e2e-a11y`. |
| `make frontend-unit` | failed | `.cartulary/test-results/20260612T034030Z-p47651` failed for the same frontend phase-map validation rule. This failure is retained as authority context only. |
| `make frontend-typecheck` | failed | `.cartulary/test-results/20260612T034513Z-p50882` failed after the target correction because `apps/web/e2e/workbook.a11y.spec.ts` was missing `gridFilterValueTestId` and `gridFilterChipTestId` imports. Imports were added; this failed run is not row evidence. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T034513Z-p50905` passed after the FE-A11Y target correction. This supports the staged map/scenario changes but does not close FE-V-P8-01 or FE-A11Y-P8-01. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T034615Z-p53342` passed after the missing UI-contract imports were added. This clears the static TypeScript blocker for the Sprint 6 visual/a11y scenario changes. |
| `make browser-e2e-visual` | failed | `.cartulary/test-results/20260612T034637Z-p53843` failed with existing P2-P7 visual fixtures and the new FE-V-P8-01 scenario. The diffs show the accepted FE-P8 saved-view toolbar expansion in the workbook first viewport and new FE-P8 actual screenshots, but the row accounting records FE-V-P8-01 `closure_status` `failed`; this run is not closure evidence. |
| `make browser-e2e-visual` | failed | `.cartulary/test-results/20260612T035219Z-p71834` passed FE-V-P8-01 scenario execution but failed target closure because the FE-P6 evidence-affordance golden remained stale at pixel level. Row accounting records FE-V-P8-01 `closure_status` `not_closed` with `failure_reason` `target_failed`; this run is not closure evidence. |
| `make browser-e2e-visual` | failed | `.cartulary/test-results/20260612T035442Z-p83814` repeated the FE-P6 evidence-affordance visual failure after FE-V-P8-01 scenario passed. The persistent diff came from run-varying `record_id` fragments in the FE-P6 crop; a visual-only mask was added before the next visual validation. |
| `scripts/start-web-e2e.sh -- ... playwright test ... --update-snapshots` | pass | Targeted FE-P6 snapshot update passed under the token-backed service wrapper and rewrote `fe-v-p6-01-timeline-evidence-count-linux.png`. This is support-only golden maintenance, not FE-P8 closure evidence. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260612T035821Z-p99246` passed with `frontend_row_accounting.v3`; FE-V-P8-01 has exact scenario title, `command_id` `cartulary.harness.command.browser_e2e_visual.v1`, `target_status` `pass`, `evidence_class` `design_direction`, `closure_status` `closed`, `registry_digest` `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, `guide_digest` `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, and FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`. |
| `make browser-e2e-a11y` | pass | `.cartulary/test-results/20260612T040028Z-p11665` passed with `frontend_row_accounting.v3`; FE-A11Y-P8-01 has exact scenario title, `command_id` `cartulary.harness.command.browser_e2e_a11y.v1`, `target_status` `pass`, `evidence_class` `design_direction`, `closure_status` `closed`, `registry_digest` `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, `guide_digest` `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, and FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`. `frontend_accessibility_summary.v2` also records FE-A11Y-P8-01 pass in scenarios, keyboard matrix, state-communication checks, contrast checks, and has no violations. |
| `make browser-e2e-a11y-preflight` | pass | `.cartulary/test-results/20260612T040239Z-p21781` passed as supplemental smoke. Its `frontend_accessibility_preflight_summary.v1` contains no FE-P8 row/scenario entries after FE-A11Y-P8-01 moved to the full implemented-row target, so it is not closure evidence. |
| Sprint 7 reconciliation | complete | Completed final target explanation refresh, `make frontend-typecheck`, `make frontend-unit`, `make phase-ledgers`, final row-target refresh, drift/shape/policy checks, full `make check`, and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246`. |

### Sprint 7: Drift, Full Check, Finalization, Closeout, And FE-P9 Handoff

| Field | Plan |
| --- | --- |
| Objective | Validate generated and harness state, run the required gate sequence for the final implementation posture, close blockers, and produce FE-P9 handoff. |
| Implementation targets | Owner maps, registry owner inputs, generated ledgers via Make targets, retained artifact roots, final blocker register, FE-P9 handoff section. |
| Validation commands | `make phase-ledger-drift`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make check`; `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when retained full-check evidence is used. |
| Expected evidence | Current map/registry/ledger digests, generated-artifact drift pass, JSON shape pass, full check pass when run, finalization pass when using retained evidence, and FE-P9 handoff with direct FE-P8 evidence roots. |
| Blockers | Generated-ledger drift, stale freshness digest, old retained artifacts, broad `make check` without row-owned evidence, `agent-finalize` without `RESULTS_DIR` when retained evidence is used, or any unclosed FE-P8 row. |
| Strict non-claims | Full `make check` must not close a FE-P8 row unless the mapped row-owned accounting and current map/registry/guide digests are present. |

#### Sprint 7 Implementation Ledger

| Material change | Files | Current status |
| --- | --- | --- |
| FE-P8 registry activation | `tools/frontend_phase_registry.json` | Promoted FE-P8 from `planned`/`partially_implemented` with `FE-P8-ACTIVATION-BLOCKER-01` to `active`/`active_green` with no activation blockers after all six row-owned evidence targets closed. Ledger regeneration, drift checks, full check, and retained-run finalization all passed. |
| FE-P8 ledger regeneration | `docs/testing/frontend_phase_coverage_ledgers/fe_p8_coverage_ledger.md`; `docs/testing/phase8_coverage_ledger.md` | `make phase-ledgers` passed at `.cartulary/test-results/20260612T040504Z-p31671` and `.cartulary/test-results/20260612T042754Z-p23231`, regenerating FE-P8 ledgers from owner inputs. Final freshness/digest checks passed after the guide/map/registry refresh and schedule regeneration. |
| Final-digest evidence refresh | Row-owning validation targets | Complete. Final accepted row-owned accounting uses registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, and FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`. |
| System-view focus recovery after sortable headers | `apps/web/src/WorkbookShell.tsx`; `apps/web/src/WorkbookShell.surfaces.test.tsx` | Fixed and validated. The system-view switcher focus recovery now targets focus anchors inside data-row grid cells before row controls, and a unit regression seeds an indicator row and asserts post-switch focus lands on `rowCellTestId("indicator-1", "indicator.indicator_type")`. The corrected implementation passed final `make check`. |

#### Sprint 7 Validation Ledger

| Command | Result | Evidence and handling |
| --- | --- | --- |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260612T040504Z-p31671` regenerated FE-P8 ledgers through Make. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260612T040531Z-p32188` passed after FE-P8 activation and ledger regeneration. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T040611Z-p32750` passed after FE-P8 activation. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T040611Z-p32770` passed with current `frontend_row_accounting.v3`; FE-U-P8-01 and FE-I-P8-01 close with final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, and FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260612T040611Z-p32773` passed after FE-P8 activation. |
| `make browser-e2e-webserver-backed` | fail, fixed by later rerun | `.cartulary/test-results/20260612T040650Z-p35163` failed in FE-B-P2-02, not an FE-P8 scenario. The row accounting has the final registry digest but `target_status` `fail`, so FE-I/FE-B/FE-E rows from this run are not closure evidence. Failure: after selecting the Indicators system view, focus landed on the first sortable header button instead of the first data-cell anchor. Handling: tightened system-view focus recovery to data-row gridcell focus targets and added unit regression coverage; final webserver-backed validation passed. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T041357Z-p50412/frontend-typecheck/target-summary.json` passed after the focus-target fix. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T041409Z-p50868/frontend-unit/frontend-row-accounting.json` passed after the focus-target fix; FE-U-P8-01 and FE-I-P8-01 remain closed with post-activation row accounting. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260612T041442Z-p52545/browser-e2e-webserver-backed/frontend-row-accounting.json` passed after the focus-target fix. `frontend_row_accounting.v3` closes FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 with command ID `cartulary.harness.command.browser_e2e_webserver_backed.v1`, phase namespace `frontend`, final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`, exact scenario titles, and no blockers. |
| `make browser-e2e-support` | pass | `.cartulary/test-results/20260612T041718Z-p73521/browser-e2e-support/frontend-row-accounting.json` passed with `frontend_row_accounting.v3`; FE-B-P8-01 closes with command ID `cartulary.harness.command.browser_e2e_support.v1`, phase namespace `frontend`, final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`, exact scenario title, and no blocker. |
| `make browser-e2e-stateful` | pass | `.cartulary/test-results/20260612T041831Z-p77171/browser-e2e-stateful/frontend-row-accounting.json` passed with `frontend_row_accounting.v3`; FE-E-P8-01 closes with command ID `cartulary.harness.command.browser_e2e_stateful.v1`, phase namespace `frontend`, final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`, exact scenario title, and no blocker. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260612T042041Z-p89288/browser-e2e-visual/frontend-row-accounting.json` passed with `frontend_row_accounting.v3`; FE-V-P8-01 closes as `design_direction` with command ID `cartulary.harness.command.browser_e2e_visual.v1`, phase namespace `frontend`, final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`, exact scenario title, and no blocker. |
| `make browser-e2e-a11y` | pass | `.cartulary/test-results/20260612T042230Z-p1011/browser-e2e-a11y/frontend-row-accounting.json` passed with `frontend_row_accounting.v3`; FE-A11Y-P8-01 closes as `design_direction` with command ID `cartulary.harness.command.browser_e2e_a11y.v1`, phase namespace `frontend`, final registry digest `9b498b63119e49bedd571d4add16f1b18fc2fbbb4b6abdb8ef184497e83f7738`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `004f092569ba76d8920cd93a867c98b00c0691182c6bfd70c937c42889f6ae86`, exact scenario title, and no blocker. `.cartulary/test-results/20260612T042230Z-p1011/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` reports schema `cartulary.frontend_accessibility_summary.v2`, `status` `pass`, FE-A11Y-P8-01 pass entries in scenarios/keyboard/state/contrast checks, and `violations: []`. |
| `make browser-e2e-a11y-preflight` | pass | `.cartulary/test-results/20260612T042410Z-p10930/browser-e2e-a11y-preflight/target-summary.json` passed as supplemental smoke with 4 tests and 0 failures. `.cartulary/test-results/20260612T042410Z-p10930/browser-e2e-a11y-preflight/accessibility-preflight/frontend-accessibility-preflight-summary.json` reports schema `cartulary.frontend_accessibility_preflight_summary.v1`, `status` `pass`, `violations: []`, and no FE-P8 row closure; listed phase rows are later blocked rows only. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260612T042530Z-p20297/frontend-import-boundary-check/target-summary.json` passed after the focus-target fix and FE-P8 target refresh. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260612T042547Z-p20716/phase-ledger-drift/tool-run-summary.json` passed after final post-activation row-target refresh. |
| `make generate-drift` | pass | `.cartulary/test-results/20260612T042602Z-p21114/generate-drift/tool-run-summary.json` passed with no generated drift after FE-P8 activation and target refresh. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260612T042615Z-p22017/generated-artifact-policy-check/tool-run-summary.json` passed; generated roots remain policy-compliant. |
| `make json-shape-check` | fail, fixed by later rerun | `.cartulary/test-results/20260612T042624Z-p22237` failed because `tools/frontend_phase_registry.json.guide_digest` still used the older frontend guide digest after the Sprint 6 FE-A11Y guide correction. Handling: updated frontend phase-map guide digests, regenerated ledgers through Make, refreshed registry manifest/ledger/freshness digests, reran shape/drift checks, then reran row-owning targets because the registry digest changed. |
| `make phase-ledgers` | pass | `.cartulary/test-results/20260612T042754Z-p23231/phase-ledgers/tool-run-summary.json` regenerated frontend ledgers after guide-digest owner metadata refresh. |
| `make json-shape-check` | fail, fixed by later rerun | `.cartulary/test-results/20260612T042812Z-p23651` passed the frontend guide/registry freshness check but failed because phase schedule inputs were stale after `tools/phase8_test_map.json` changed. Handling: ran `make phase-schedules`, then reran shape/drift checks successfully. |
| `make phase-schedules` | pass | `.cartulary/test-results/20260612T042832Z-p24069/phase-schedules/tool-run-summary.json` regenerated phase schedules from the updated Phase 8 owner map. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260612T042841Z-p24267/json-shape-check/tool-run-summary.json` passed after frontend guide/map/registry digest refresh and schedule regeneration. Corrected registry digest is `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`; FE-P8 map digest is `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`. Row-owning targets must be refreshed again under this registry digest. |
| `make frontend-unit` | pass | `.cartulary/test-results/20260612T042912Z-p24795/frontend-unit/frontend-row-accounting.json` passed after registry digest correction; FE-U-P8-01 and FE-I-P8-01 close with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, and no blockers. |
| `make browser-e2e-webserver-backed` | pass | `.cartulary/test-results/20260612T042944Z-p26418/browser-e2e-webserver-backed/frontend-row-accounting.json` passed after registry digest correction; FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 close with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario titles, and no blockers. |
| `make browser-e2e-support` | pass | `.cartulary/test-results/20260612T043153Z-p45863/browser-e2e-support/frontend-row-accounting.json` passed after registry digest correction; FE-B-P8-01 closes with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario title, and no blocker. |
| `make browser-e2e-stateful` | pass | `.cartulary/test-results/20260612T043301Z-p49404/browser-e2e-stateful/frontend-row-accounting.json` passed after registry digest correction; FE-E-P8-01 closes with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario title, and no blocker. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260612T043510Z-p61479/browser-e2e-visual/frontend-row-accounting.json` passed after registry digest correction; FE-V-P8-01 closes as `design_direction` with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario title, and no blocker. |
| `make browser-e2e-a11y` | pass | `.cartulary/test-results/20260612T043651Z-p72677/browser-e2e-a11y/frontend-row-accounting.json` passed after registry digest correction; FE-A11Y-P8-01 closes as `design_direction` with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario title, and no blocker. Accessibility summary `.cartulary/test-results/20260612T043651Z-p72677/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` reports `status` `pass`, `violations: []`, and FE-A11Y-P8-01 pass entries in scenario, keyboard, state-communication, and contrast checks. |
| `make browser-e2e-a11y-preflight` | pass | `.cartulary/test-results/20260612T043822Z-p82505/browser-e2e-a11y-preflight/target-summary.json` passed after registry digest correction as support-only smoke with 4 tests and 0 failures; it remains non-closure evidence for FE-P8. |
| `make frontend-typecheck` | pass | `.cartulary/test-results/20260612T043924Z-p91729/frontend-typecheck/target-summary.json` passed after final guide/map/registry metadata refresh. |
| `make frontend-import-boundary-check` | pass | `.cartulary/test-results/20260612T043938Z-p92184/frontend-import-boundary-check/target-summary.json` passed after final guide/map/registry metadata refresh. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260612T043956Z-p92610/phase-ledger-drift/tool-run-summary.json` passed after final guide/map/registry metadata refresh. |
| `make phase-schedule-drift` | pass | `.cartulary/test-results/20260612T044002Z-p92956/phase-schedule-drift/tool-run-summary.json` passed after `make phase-schedules` regenerated schedule outputs. |
| `make generate-drift` | pass | `.cartulary/test-results/20260612T044017Z-p93196/generate-drift/tool-run-summary.json` passed after final guide/map/registry metadata refresh. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260612T044024Z-p94062/generated-artifact-policy-check/tool-run-summary.json` passed after final guide/map/registry metadata refresh. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260612T044028Z-p94292/json-shape-check/tool-run-summary.json` passed after final guide/map/registry metadata refresh and schedule regeneration. |
| `make check` | fail, fixed by later rerun | `.cartulary/test-results/20260612T044042Z-p94774` failed on child target `lint-biome` before broad service/browser work completed. Biome reported formatting/import organization plus `useExhaustiveDependencies` and optional-chain diagnostics in FE-P8-touched frontend files. Handling: ran `make format`, applied manual hook-dependency and optional-chain fixes, then reran `make format`, `make lint-biome`, and full `make check` successfully. |
| `make format` | pass | `.cartulary/test-results/20260612T044235Z-p22035/format/target-summary.json` passed after manual lint fixes. |
| `make lint-biome` | pass | `.cartulary/test-results/20260612T044241Z-p23410/lint-biome/target-summary.json` passed after formatting and manual lint fixes. |
| `make check` | pass | `.cartulary/test-results/20260612T044428Z-p24246` passed after the lint/format remediation, with 145/145 work units, 791 tests, 0 failed, and no missing tests. This broad gate is retained full-check evidence only; FE-P8 row closure remains anchored to current row-owned accounting roots listed above. |
| `make browser-e2e-support` | pass | `.cartulary/test-results/20260612T044835Z-p88067/browser-e2e-support/frontend-row-accounting.json` passed after formatting/lint remediation; FE-B-P8-01 remains closed with registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`, exact scenario title, and no blocker. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json` passed after formatting/lint remediation with 23 tests and 0 failures; FE-V-P8-01 remains closed as `design_direction` with current registry/guide/FE-P8 map digests, exact scenario title, and no blocker. |
| `make browser-e2e-a11y` | pass | `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json` passed after formatting/lint remediation with 16 tests and 0 failures; FE-A11Y-P8-01 remains closed as `design_direction` with current registry/guide/FE-P8 map digests, exact scenario title, and no blocker. Accessibility summary `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` reports `status` `pass` and `violations: []`. |
| `make browser-e2e-a11y-preflight` | pass | `.cartulary/test-results/20260612T045159Z-p12967/browser-e2e-a11y-preflight/target-summary.json` passed after formatting/lint remediation with 4 tests and 0 failures; it remains support-only smoke and is not FE-P8 closure evidence. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246` | pass | First finalization pass at `.cartulary/test-results/20260612T044704Z-p79312/agent-finalize/finalize-summary.json` refreshed retained-run maintenance and updated 4 duration-baseline files. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246` | pass | Final finalization pass at `.cartulary/test-results/20260612T045247Z-p22120/agent-finalize/finalize-summary.json` reported `generated=unchanged`, `updated_file_count=0`, and `run_checks.status=pass`; this is the final retained-run maintenance evidence. |
| `make lint-markdown` | pass | `.cartulary/test-results/20260612T050551Z-p29069/lint-markdown/tool-run-summary.json` passed after final tracker-only edits to this plan. |

#### Sprint 7 Pre-Implementation Target-Surface Ledger

| Command | Result | Evidence |
| --- | --- | --- |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | pass | Public phase-maintenance target; latest artifact reported as `.cartulary/test-results/20260612T023814Z-p37199/phase-ledger-drift/phase-ledger-drift/phase-summary.json`. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | pass | Public phase-maintenance target; latest artifact reported as `.cartulary/test-results/20260612T023823Z-p38709/generate-drift/generate-drift/phase-summary.json`. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | pass | Public phase-maintenance target; latest artifact reported as `.cartulary/test-results/20260612T023814Z-p37222/generated-artifact-policy-check/generated-artifact-policy-check/phase-summary.json`. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | pass | Public phase-maintenance target; latest artifact reported as `.cartulary/test-results/20260612T023814Z-p37177/json-shape-check/json-shape-check/phase-summary.json`. |
| `make explain-target TARGET=check DETAIL=summary` | pass | Public full-gate target; latest artifact reported as `.cartulary/test-results/20260612T014246Z-p80421/check/target-summary.json`. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | pass | Public helper target requiring `RESULTS_DIR` when retained-run maintenance is intended; latest artifact reported as `.cartulary/test-results/20260612T023830Z-p39564/agent-finalize/agent-finalize/phase-summary.json`. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | pass | Public fast-verification target; latest artifact reported as `.cartulary/test-results/20260612T023814Z-p37257/frontend-typecheck/target-summary.json`. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | pass | Public fast-verification target; latest artifact reported as `.cartulary/test-results/20260612T023814Z-p37246/frontend-import-boundary-check/target-summary.json`. |

This ledger is target-surface context only. It does not replace later Sprint 7 drift, full-check, row-accounting, or finalization evidence.

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

`make browser-e2e-a11y` exists, was explained, and is the final FE-P8 map-authorized closure target for FE-A11Y-P8-01. `make browser-e2e-a11y-preflight` remains support-only smoke and must not replace mapped full-a11y row accounting.

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
| FE-U-P8-01 | `frontend-unit` row accounting, target summary, phase summary, unit test output | `cartulary.frontend_row_accounting.v3`, row FE-U-P8-01, command `cartulary.harness.command.frontend_unit.v1`, target `frontend-unit`, current guide digest `f582d6fc183210a8404b020116b2dd84a7f92b216e06973e5bc53a0e53d85059`, current registry digest, current FE-P8 map digest, passing unit identity for sort/filter/group/layout/chip compilation, closure status closed, artifact root `.cartulary/test-results/20260611T204024Z-p40590`. |
| FE-I-P8-01 | `frontend-unit` and `browser-e2e-webserver-backed` row accounting, route/browser artifacts | `cartulary.frontend_row_accounting.v3`, row FE-I-P8-01, command IDs for mapped targets, exact scenario title, public saved-view/workbook-preference contract evidence, active surface scope evidence, current registry digest `aadfd32a2626a25700ae23874d60cb6622564323200db7e39f8b873f15906751`, pass status, closure status closed, artifact roots `.cartulary/test-results/20260612T021350Z-p95057` and `.cartulary/test-results/20260612T021406Z-p96577`. |
| FE-B-P8-01 | `browser-e2e-webserver-backed` and `browser-e2e-support` row accounting/summaries | Row FE-B-P8-01, exact scenario title, command helper proof for sort/filter/group/chips/layout/group expand-collapse/startup-default controls, stable selector evidence, current digests, pass status, closure status, artifact roots. |
| FE-E-P8-01 | `browser-e2e-webserver-backed` and `browser-e2e-stateful` row accounting, route traces, reload artifacts | Row FE-E-P8-01, exact scenario title, public `/api/v1/` saved-view/preference/query route evidence, reload replay evidence, current digests, pass status, closure status, artifact roots. |
| FE-V-P8-01 | `browser-e2e-visual` row accounting, fixture registry references, screenshots/goldens, target summary | Row FE-V-P8-01, exact scenario title, current guide/registry/map digests, captured states for saved-view selector, active chips, grouped result, group row, default/startup indicator, empty successful query, pass status, design-direction closure status, artifact roots, no stale fixture blockers. |
| FE-A11Y-P8-01 | `browser-e2e-a11y` row accounting, `cartulary.frontend_accessibility_summary.v2`, target summary | Row FE-A11Y-P8-01, exact scenario title, current digests, keyboard reachability and announced state evidence for all mapped controls, pass status, design-direction closure status, artifact root, no blocked/preflight-only closure mismatch. |

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
| All FE-P8 | `tools/frontend_phase_registry.json` | Resolved in Sprint 7. FE-P8 is now `active`, row rollup is `active_green`, and `FE-P8-ACTIVATION-BLOCKER-01` is removed by owner input. | No longer blocks registry activation or final phase closure. | None for FE-P8; keep future phases from reopening activation without owner-map evidence. | None. |
| All FE-P8 | Post-activation row-accounting refresh | Resolved by final row-owned reruns under registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`, guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`, and FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`. | No longer blocks final phase completion. | Retain pre-activation roots as context only. | None. |
| All FE-P8 | `.cartulary/test-results/20260612T042624Z-p22237/json-shape-check/json-shape-check/stderr.log` | Resolved. Frontend phase-map guide digests were updated, ledgers regenerated through Make, registry manifest/ledger/freshness digests refreshed, row-owning targets rerun, and final `make json-shape-check` passed at `.cartulary/test-results/20260612T044028Z-p94292/json-shape-check/tool-run-summary.json`. | No longer blocks guide/map/registry freshness. | None. | None. |
| Phase schedules | `.cartulary/test-results/20260612T042812Z-p23651/json-shape-check/json-shape-check/stderr.log` | Resolved. `make phase-schedules` passed at `.cartulary/test-results/20260612T042832Z-p24069/phase-schedules/tool-run-summary.json`, `make phase-schedule-drift` passed at `.cartulary/test-results/20260612T044002Z-p92956/phase-schedule-drift/tool-run-summary.json`, and final shape check passed. | No longer blocks generated schedule freshness. | None. | None. |
| FE-I-P8-01, FE-B-P8-01, FE-E-P8-01 | `.cartulary/test-results/20260612T040650Z-p35163/browser-e2e-webserver-backed/frontend-row-accounting.json` | Resolved by the final successful webserver-backed row accounting under `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-webserver-backed/frontend-row-accounting.json`. The earlier post-activation run failed FE-B-P2-02 focus restoration after sortable header buttons became focusable. | No longer blocks webserver-backed closure. | Keep the failed root as failure context only. | None. |
| All FE-P8 | `.cartulary/test-results/20260612T044042Z-p94774/check` | Resolved by `.cartulary/test-results/20260612T044428Z-p24246`. The earlier broad check failed at `lint-biome` after FE-P8 source and metadata changes; formatting, hook-dependency, and optional-chain fixes were applied, `make lint-biome` passed, and the full rerun passed. Final `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246` passed at `.cartulary/test-results/20260612T045247Z-p22120`. | No longer blocks the full-check or retained-finalization gates. | None. | None. |
| All FE-P8 | FE-P8 guide vs live map | Resolved for FE-P8 closure by owner-map exact IDs, frontend guide/map digest refresh, and Make-regenerated ledgers. | No longer blocks FE-P8 row closure; row claims remain limited to exact owner-map IDs. | Do not expand FE-P8 claims beyond exact map IDs in future docs. | Expanded guide-range claims only. |
| FE-U-P8-01 | Latest `frontend-unit` row accounting | Resolved in Sprint 2. `.cartulary/test-results/20260611T204024Z-p40590/frontend-unit/frontend-row-accounting.json` includes FE-U-P8-01 with `closure_status` `closed`. | No longer blocks FE-U-P8-01 unit closure. | Continue to exclude FE-U-P8-01 from browser, visual, accessibility, and Core 05 claims. | None for FE-U-P8-01 unit closure. |
| FE-I-P8-01 | Latest `frontend-unit` and `browser-e2e-webserver-backed` row accounting | Resolved in Sprint 3. `.cartulary/test-results/20260612T021350Z-p95057/frontend-unit/frontend-row-accounting.json` and `.cartulary/test-results/20260612T021406Z-p96577/browser-e2e-webserver-backed/frontend-row-accounting.json` include FE-I-P8-01 with `closure_status` `closed`. | No longer blocks FE-I-P8-01 integration closure. | Continue to exclude FE-I-P8-01 from reload replay, browser-helper, visual, accessibility, and Core 05 claims. | None for FE-I-P8-01 integration closure. |
| FE-B-P8-01 | Sprint 4 browser validation | Resolved by final webserver-backed accounting `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-webserver-backed/frontend-row-accounting.json` and final support accounting `.cartulary/test-results/20260612T044835Z-p88067/browser-e2e-support/frontend-row-accounting.json`. Both close FE-B-P8-01 with mapped command IDs, current digests, exact scenario title, and no row blocker. | No longer blocks FE-B-P8-01 browser-helper closure. | Continue to exclude FE-B-P8-01 from FE-E reload/query replay, visual readiness, accessibility readiness, and Core 05 claims. | None for FE-B-P8-01 closure. |
| FE-E-P8-01 | Sprint 5 browser and stateful validation | Resolved by final webserver-backed accounting `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-webserver-backed/frontend-row-accounting.json` and final stateful accounting `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-stateful/frontend-row-accounting.json`. Both close FE-E-P8-01 with mapped command IDs, current digests, exact scenario title, public `/api/v1/` replay evidence, and no row blocker. | No longer blocks FE-E-P8-01 product closure. | Continue to exclude FE-E-P8-01 from visual readiness, accessibility readiness, rollback readiness, evidence handles, and Core 05 claims. | None for FE-E-P8-01 closure. |
| FE-V-P8-01 | Sprint 6 visual scenario, fixture registry update, and visual validation | Resolved by final visual accounting `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json`, which closes FE-V-P8-01 with exact scenario title, current guide/registry/map digests, design-direction evidence class, FE-P8 goldens, and no row blocker. Earlier failed visual runs remain failure context only. | No longer blocks FE-V-P8-01 visual readiness. | Keep visual evidence design-direction only. | None for FE-V-P8-01 closure. |
| FE-A11Y-P8-01 | Sprint 6 a11y scenario, design-ID reconciliation, and accessibility validation | Resolved by final accessibility accounting `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json`, which closes FE-A11Y-P8-01 with exact scenario title, current guide/registry/map digests, design-direction evidence class, and no row blocker. `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json` records `status` `pass` and `violations: []`. Preflight `.cartulary/test-results/20260612T045159Z-p12967` remains support-only smoke. | No longer blocks FE-A11Y-P8-01 accessibility readiness. | Keep accessibility evidence design-direction only. | None for FE-A11Y-P8-01 closure. |
| Remaining product rows | Existing FE-P8-adjacent tests | Resolved. All FE-P8 product rows now have exact mapped scenario identity and current row accounting under final registry/map/guide digests. | No longer blocks FE-P8 product-row closure. | Treat non-row-owned FE-P8-adjacent tests as supplemental context only. | None. |
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
17. Any FE-P8 row closure from target explanations without current row-owned accounting.
18. FE-B-P8-01 closure from staged code or owner-map promotion before current validation artifacts exist.
19. Final FE-P8 completion from pre-activation row-accounting roots after the registry digest changed.
20. FE-P8 product-row closure from the failed `.cartulary/test-results/20260612T040650Z-p35163` webserver-backed run.

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

Current status after final Sprint 7 closeout:

| Criterion | Current status |
| --- | --- |
| 1. Initial plan creation | Pass from Sprint 1. |
| 2. Product-conformance row closure | Pass. FE-U-P8-01, FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 have current mapped product-conformance row accounting with exact scenario/unit identities, pass target status, `closure_status` `closed`, current digests, and no blockers. |
| 3. Saved-view integration closure | Pass. FE-I-P8-01 has current `frontend-unit` and `browser-e2e-webserver-backed` row accounting proving active-surface scope and public saved-view/workbook-preference contracts. |
| 4. Query-control and grouping closure | Pass. FE-U-P8-01 and FE-B-P8-01 prove stable `field_key` query controls, helper commands, active chips, group expand/collapse, layout/default/startup controls, and non-writable group rows. |
| 5. Stateful reload and replay closure | Pass. FE-E-P8-01 has current public `/api/v1/` route evidence for saved-view persistence, default/startup persistence, reload behavior, and query replay. |
| 6. Visual readiness | Pass as design-direction evidence only. FE-V-P8-01 has current visual row accounting and fixture coverage for saved-view selector, active chips, grouped result, group row, default/startup indicator, and empty successful query. |
| 7. Accessibility readiness | Pass as design-direction evidence only. FE-A11Y-P8-01 has current full-a11y row accounting and accessibility summary for keyboard reachability and announced state with `violations: []`; preflight remains support-only smoke. |
| 8. Full FE-P8 phase completion | Pass. Registry is `active`, row rollup is `active_green`, all six rows are closed under current map authority, ledgers were regenerated by Make, drift/shape/policy checks passed, full `make check` passed, and retained-run finalization passed. |
| 9. FE-P9 handoff readiness | Pass. This handoff lists final registry status, row inventory, direct evidence roots, drift/finalization outcomes, strict non-claims, and no unresolved FE-P8 blockers. |

## FE-P9 Handoff

FE-P8 is complete as of 2026-06-12 under current frontend map authority. FE-P9 may use this plan as dependency context for saved-view/query-control behavior, but it must not treat FE-P8 as evidence for FE-P9 inspector behavior, rollback, evidence handles, Core 05 publication, benchmark, or visual-publication readiness.

| Handoff field | Current FE-P8 value |
| --- | --- |
| Registry status | `active`. |
| Row rollup | `active_green`. |
| Activation blocker | Cleared; `FE-P8-ACTIVATION-BLOCKER-01` removed from `tools/frontend_phase_registry.json`. |
| Product rows | Closed. FE-U-P8-01 and FE-I-P8-01 close in `.cartulary/test-results/20260612T044428Z-p24246/frontend-unit/frontend-row-accounting.json`; FE-I-P8-01, FE-B-P8-01, and FE-E-P8-01 close in `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-webserver-backed/frontend-row-accounting.json`; FE-E-P8-01 also closes in `.cartulary/test-results/20260612T044428Z-p24246/browser-e2e-stateful/frontend-row-accounting.json`; FE-B-P8-01 also closes in `.cartulary/test-results/20260612T044835Z-p88067/browser-e2e-support/frontend-row-accounting.json`. |
| Design-direction rows | Closed as design-direction only. FE-V-P8-01 closes in `.cartulary/test-results/20260612T044934Z-p91545/browser-e2e-visual/frontend-row-accounting.json`; FE-A11Y-P8-01 closes in `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/frontend-row-accounting.json` with accessibility summary `.cartulary/test-results/20260612T045057Z-p3086/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`. |
| Final digest set | Registry digest `14bb768467334eb4ca26c6d99d7a4d6c851bb4745dfc3ed9139d9becd0e839f0`; frontend guide digest `4d04187d15d86bd49ec08addb3aa25e625e4fd7740918e4ebe0e2376f56944d1`; FE-P8 map digest `c54d24a8fd60c5c4edd1fd45a217ede48860c58c29d3bd27e959eaf0694daa13`. |
| Supplemental support evidence | `.cartulary/test-results/20260612T045159Z-p12967/browser-e2e-a11y-preflight/target-summary.json` passed with 4 tests and 0 failures. It is support-only smoke and not FE-P8 row closure evidence. |
| Drift and shape outcomes | `make phase-ledger-drift` passed at `.cartulary/test-results/20260612T043956Z-p92610/phase-ledger-drift/tool-run-summary.json`; `make phase-schedule-drift` passed at `.cartulary/test-results/20260612T044002Z-p92956/phase-schedule-drift/tool-run-summary.json`; `make generate-drift` passed at `.cartulary/test-results/20260612T044017Z-p93196/generate-drift/tool-run-summary.json`; `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260612T044024Z-p94062/generated-artifact-policy-check/tool-run-summary.json`; `make json-shape-check` passed at `.cartulary/test-results/20260612T044028Z-p94292/json-shape-check/tool-run-summary.json`. |
| Full check and finalization | `make check` passed at `.cartulary/test-results/20260612T044428Z-p24246` with 145/145 work units, 791 tests, 0 failed, and 0 missing. Final `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260612T044428Z-p24246` passed at `.cartulary/test-results/20260612T045247Z-p22120/agent-finalize/finalize-summary.json` with generated files unchanged and run checks passing. |
| Retained context roots | Failed or stale roots remain context only: `.cartulary/test-results/20260612T040650Z-p35163`, `.cartulary/test-results/20260612T044042Z-p94774`, `.cartulary/test-results/20260611T193506Z-p52331`, `.cartulary/test-results/20260610T191949Z-p14292`, and `.cartulary/test-results/20260610T171433Z-p14284`. They do not close FE-P8. |
| Strict non-claims | No visual/product promotion, no accessibility/product promotion, no Core 05 publication, no evidence handle readiness, no rollback readiness, no benchmark/fixture-sensitive publication claim, no FE-P9 implementation claim, and no direct `react-data-grid` import permission outside `packages/grid-adapter`. |
| Remaining FE-P8 blockers | None. |
| Remaining source limits | FE-P9 implementation remains outside FE-P8 scope. Future phases must use their own owner maps, row accounting, and retained-run finalization. |
