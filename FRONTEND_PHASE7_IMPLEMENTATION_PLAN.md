# Frontend Phase 7 Implementation Plan

## Summary

FE-P7 covers live collaboration and conflicts in the frontend: WebSocket stream handling, event reducers, presence anchoring, live row updates, reset and invalidate handling, stale-row requery, same-field conflicts, and resolver behavior.

Current repository status: FE-P7 is `planned`, `row_rollup_state=no_rows_implemented`, and blocked for row and phase closure. The frontend registry reports `FE-P7-ACTIVATION-BLOCKER-01` with reason `frontend_phase_not_active`; the FE-P7 map reports every FE-P7 row as `claim_status=blocked`. The generated FE-P7 coverage ledger is current to the map and `make phase-ledger-drift` passed during this inspection.

Execution eligibility: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P7` says planned frontend phases are explainable but not executable as phases. Individual targets exist, but current rows are not closure-eligible because the map marks them blocked and their targets are not required for closure yet.

Phase state: FE-P7 is planned and blocked for closure. It is not active. It is not treated as stale after the successful phase-ledger drift check, but it has no current direct FE-P7 row-owned evidence roots.

This document is an execution roadmap, progress marker, validation guide, blocker register, and FE-P8 handoff aid. It is not product behavior authority, row-closure authority, registry authority, or evidence authority. Row promotion and closure come only from authored maps, mapped target evidence, current row-accounting artifacts, and applicable owner references.

## Authority Model

Authoritative and supporting sources are applied in this order:

1. Adopted Cartulary NLSpecs, with `docs/testing-harness-nlspec.md` limited to harness mechanics.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for explicit claim-bearing timed, benchmark, visual, fixture-sensitive, or publication evidence boundaries.
4. `docs/domain.md` for vocabulary and concept-boundary interpretation only.
5. `docs/guides/cartulary_frontend_implementation_testing_guide.md` for FE-P7 planning mechanics, row mapping, evidence classes, and row-owner mapping.
6. `docs/guides/cartulary_implementation_testing_guide.md` for sequencing, test-row discipline, shared harnesses, completion rules, and coverage-ledger shape.
7. `docs/guides/cartulary-dev-guide.md` for repo-local frontend package boundaries, generated-artifact policy, Make targets, workspace shape, and frontend implementation baseline.
8. `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` for design-direction, visual, and accessibility readiness only.
9. Previous frontend implementation plans as examples only.
10. Research reports as rationale only.

Core 00 through Core 04 own product implementation-conformance behavior. Core 05 is inactive unless explicit claim-bearing timed, benchmark, visual, fixture-sensitive, or publication metadata exists and satisfies Core 05.

`docs/testing-harness-nlspec.md` owns harness mechanics only. `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns frontend planning and verification mechanics only.

Generated phase ledgers are downstream generated artifacts and are never row owners or closure evidence by themselves. Design and visual guides are design-direction inputs only. Research reports are rationale only. Conflicts between live sources must be recorded as blockers instead of silently reconciled.

## Current Repo Status

Exact live inputs inspected:

- `tools/frontend_phase_registry.json`
- `tools/frontend_phase_maps/fe_p7_test_map.json`
- `docs/testing/frontend_phase_coverage_ledgers/fe_p7_coverage_ledger.md`
- `tools/frontend_visual_fixture_registry.json`
- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` through `FRONTEND_PHASE6_IMPLEMENTATION_PLAN.md`
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`
- `docs/testing-harness-nlspec.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/domain.md`
- `docs/guides/cartulary_implementation_testing_guide.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary-ui-ux-design-guide.md`
- `docs/design.md`
- `docs/guides/cartulary_visual_golden_maintenance.md`
- Relevant FE-P7 app code under `apps/web`
- Relevant grid-adapter code under `packages/grid-adapter`
- Generated protocol contract facade under `packages/protocol-ts`
- Selector and test-id builders under `packages/ui-contracts`
- Reusable browser helpers under `packages/test-utils` and `apps/web/e2e`
- Make target explanations listed below.

Frontend phase registry status for FE-P0 through FE-P7:

| Phase | Status | Row rollup | Blocker status |
| --- | --- | --- | --- |
| FE-P0 | `active` | `active_green` | none |
| FE-P1 | `active` | `active_green` | none |
| FE-P2 | `active` | `active_green` | none |
| FE-P3 | `active` | `active_green` | none |
| FE-P4 | `active` | `active_green` | none |
| FE-P5 | `active` | `active_green` | none |
| FE-P6 | `active` | `active_green` | none |
| FE-P7 | `planned` | `no_rows_implemented` | `FE-P7-ACTIVATION-BLOCKER-01`, reason `frontend_phase_not_active` |

FE-P7 registry detail:

- `status`: `planned`
- `row_rollup_state`: `no_rows_implemented`
- `evidence_freshness_digest`: `63e18f46b12e86912786867dc3afbd81133261ae13e7db40f3f4a50d794a93a6`
- `manifest_digest`: `ae9224fbbb8af587f1a73bb8e1b0a612f77de43407a1feb98b9c39e3879cd6cd`
- `ledger_digest`: `3a8afdfc9dc119088ba3a04888b04d2feb68641267926a4565e8d03efb8aa4d3`
- Activation blocker: `FE-P7-ACTIVATION-BLOCKER-01`, `frontend_phase_not_active`, "Frontend phase is not active until direct row evidence and freshness validation are promoted together."
- Dependencies: FE-P0 through FE-P6.

FE-P7 map inventory and current `claim_status`:

| Row | Layer | Evidence class | Targets | Current `claim_status` |
| --- | --- | --- | --- | --- |
| `FE-U-P7-01` | unit | `product_conformance` | `frontend-unit` | `blocked` |
| `FE-U-P7-02` | unit | `product_conformance` | `frontend-unit` | `blocked` |
| `FE-I-P7-01` | integration | `product_conformance` | `frontend-unit`, `browser-e2e-webserver-backed` | `blocked` |
| `FE-E-P7-01` | e2e | `product_conformance` | `browser-e2e-stateful`, `browser-e2e-webserver-backed` | `blocked` |
| `FE-V-P7-01` | visual | `design_direction` | `browser-e2e-visual` | `blocked` |
| `FE-A11Y-P7-01` | accessibility | `design_direction` | `browser-e2e-a11y-preflight` | `blocked` |

Generated FE-P7 coverage ledger status:

- Ledger path: `docs/testing/frontend_phase_coverage_ledgers/fe_p7_coverage_ledger.md`
- Generated status: `planned`
- Generated rollup: `no_rows_implemented`
- Rows: same six FE-P7 rows as the live map, all `blocked`
- Drift status: `make phase-ledger-drift` passed with run root `.cartulary/test-results/20260607T215728Z-p3866530`
- Closure limit: the ledger is downstream and cannot close rows by itself.

FE-P7 visual fixture IDs from the live visual fixture registry:

| Fixture ID | Registry title | Status | Owner rows | Notes |
| --- | --- | --- | --- | --- |
| `FE-VFIX-03` | Same-field conflict | `current` | `FE-V-P7-01` | Uses `V-6-GRID-02` and golden `v-6-grid-02-conflict-resolver-linux.png`; context only until FE-P7 row accounting is current. |
| `FE-VFIX-04` | Row-gutter presence | `current` | `FE-V-P7-01` | Uses `V-6-GRID-01` and golden `v-6-grid-01-presence-markers-linux.png`; context only until FE-P7 row accounting is current. |
| `FE-VFIX-08` | Save-state strip | `current` | `FE-V-P4-01`, `FE-V-P7-01` | Includes `v-6-grid-03-*` and FE-P4 save-state goldens; shared visual context only. |

Retained current evidence roots that may be relevant:

- No direct FE-P7 row-owned evidence roots were found by searching `.cartulary/test-results` for FE-P7 row IDs.
- Latest target artifacts reported by `make explain-target` include `.cartulary/test-results/20260607T212136Z-p3783644` for `frontend-unit`, `browser-e2e-stateful`, `browser-e2e-webserver-backed`, `check`, `generate-drift`, `generated-artifact-policy-check`, `json-shape-check`, `phase-ledger-drift`, `frontend-typecheck`, and `frontend-import-boundary-check`. These are retained context only unless row ownership, freshness, target mapping, and scenario expectations are current and directly applicable.
- `browser-e2e-a11y` and `browser-e2e-a11y-preflight` reported no latest artifact.
- FE-P6 handoff roots are listed in the FE-P6 handoff section as dependency context only.

Relevant existing code, selectors, helpers, and tests:

- `apps/web/src/WorkbookShell.tsx` builds `changeSocketURL` as `/ws/v1/incidents/${incidentId}`, handles `hello`, `resume`, `hello_ack`, `resume_ack`, `reset_required`, presence snapshots/deltas, `session_revoked`, stream sequence gaps, live row patching, stale requery, same-field conflict registration, conflict resolver UI, conflict resolution `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve`, save-state status, and presence status.
- `apps/web/src/workbookPendingQueue.ts` models pending replay, same-field conflict parsing, conflict anchors, save-state conflict anchors, auth pause, retryable failures, queue ordering, and `record_id + field_key + base_row_version` identity.
- `apps/web/src/workbookShellPhase4.ts` normalizes and applies `record_changed` payloads and ignores self-originated socket transactions.
- `apps/web/src/timelineWorkbookTestSupport.ts` provides jsdom WebSocket mocks, live row helpers, visible row helpers, and `/api/v1/records/` patch helpers for unit tests.
- `apps/web/src/WorkbookShell.phase7.test.tsx` exists, but currently covers Phase 7 workbook history support, including remove/invalidate socket continuity. It is context only for FE-P7 until mapped row accounting exists.
- `apps/web/e2e/phase6.collaboration.spec.ts` and `apps/web/e2e/phase6Harness.ts` already exercise several collaboration behaviors through `/ws/v1/` and `/api/v1/`, including presence, same-field conflict resolver UI, pending replay, revocation, and stable marker anchoring. These are FE-P6-era tests and cannot close FE-P7 without current FE-P7 row mapping and accounting.
- `apps/web/e2e/workbook.visual.spec.ts` contains V-6 visual scenarios for row-gutter presence, same-field conflict resolver, and pending save-state transitions. These are visual readiness context, not FE-P7 closure by filename or fixture title.
- `apps/web/e2e/workbook.a11y-preflight.spec.ts` runs blocked later-phase accessibility smoke by reading blocked accessibility scenario titles from `apps/web/e2e/a11yPhaseMap.ts`; it is not row closure unless the live map and guide explicitly allow it.
- `packages/grid-adapter/src/core.ts` and `packages/grid-adapter/src/index.tsx` preserve record identity through `recordId`, `fieldKey`, presentation rows, navigation, paste targets, and row rendering.
- `packages/protocol-ts/src/index.ts` exposes generated contract artifacts and public envelope types from generated roots.
- `packages/ui-contracts/src/index.ts` defines stable selectors such as `conflictMarkerTestId(recordId, fieldKey)`, `rowPresenceMarkerTestId(recordId)`, `cellPresenceMarkerTestId(recordId, fieldKey)`, `saveStateTestId()`, `pendingQueueNoticeTestId()`, and `rowCellTestId(recordId, fieldKey)`.
- `packages/test-utils/src/index.ts` provides browser command helpers and `assertMarkerAnchoredToGridTarget`, which verifies marker and target share `record_id` and `field_key` in the grid.

Make target explanation results:

| Command | Result |
| --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P7` | Passed. FE-P7 is planned; rows are blocked; planned frontend phases are explainable but not executable. |
| `make explain-target TARGET=frontend-unit DETAIL=summary` | Passed. Public fast-verification target; latest artifact `.cartulary/test-results/20260607T212136Z-p3783644/frontend-unit/frontend-unit/phase-summary.json`; phase coverage excludes phase7. |
| `make explain-target TARGET=browser-e2e-stateful DETAIL=summary` | Passed. Public full-gate target with Postgres, object store, browser stack; phase coverage includes phase1 and phase6. |
| `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary` | Passed. Public full-gate target with webserver-backed browser batch; phase coverage includes phase7. |
| `make explain-target TARGET=browser-e2e-visual DETAIL=summary` | Passed. Public visual browser batch; phase coverage currently phase3 through phase6. |
| `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` | Passed. Public a11y browser batch; phase coverage none; no latest artifact. |
| `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary` | Passed. Public blocked-row accessibility preflight smoke; phase coverage none; no latest artifact. |
| `make explain-target TARGET=phase-ledger-drift DETAIL=summary` | Passed. Public phase-maintenance target. |
| `make explain-target TARGET=generate-drift DETAIL=summary` | Passed. Public generated artifact drift target. |
| `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary` | Passed. Public generated artifact policy drift target. |
| `make explain-target TARGET=json-shape-check DETAIL=summary` | Passed. Public JSON and manifest shape target. |
| `make explain-target TARGET=check DETAIL=summary` | Passed. Public full-gate scheduler target. |
| `make explain-target TARGET=agent-finalize DETAIL=summary` | Passed. Public phase-maintenance target with optional `RESULTS_DIR`. |
| `make explain-target TARGET=frontend-typecheck DETAIL=summary` | Passed. Public fast-verification target. |
| `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary` | Passed. Public fast-verification target. |
| `make explain-target TARGET=frontend-phase-drift DETAIL=summary` | Failed as expected: target not declared. Record as `TODO: command not present`. |
| `make explain-target TARGET=generated-drift DETAIL=summary` | Failed as expected: target not declared. Use `make generate-drift`. |
| `make explain-target TARGET=lint-docs DETAIL=summary` | Failed as expected: target not declared. No public Markdown lint target found. |
| `make explain-target TARGET=format-docs DETAIL=summary` | Failed as expected: target not declared. No public Markdown format target found. |

Do not infer row closure from previous phase plans, previous retained runs, generated ledgers, fixture registry status, visual filenames, scenario titles, broad `make check`, or stale artifacts.

## Source Limits

- Previous phase plans are examples only.
- FE-P6 handoff can supply dependency context only, not FE-P7 row closure.
- Generated ledgers cannot close rows.
- Retained old artifacts can be context only unless row ownership, freshness, target mapping, and scenario expectations are current and directly applicable.
- Visual and accessibility artifacts remain design-direction evidence unless the row map says otherwise and owner rules allow it.
- Frontend-only mocks cannot close public-route product-conformance behavior.
- Public-route product-conformance rows must close through public `/api/v1/` and `/ws/v1/` evidence where mapped.
- Core 05 claim-publication evidence remains absent unless explicit claim-bearing metadata exists and satisfies Core 05.
- The FE-P7 guide section gives scope and row-intent ranges, but the live map owns exact current REQ IDs, AC IDs, targets, blockers, and `claim_status`.
- Existing V-6 visual scenarios and Phase 6 collaboration tests overlap FE-P7 behavior but are not FE-P7 row closure without current FE-P7 row-owned accounting.
- `browser-e2e-a11y-preflight` is a blocked-row smoke target; it is not accessibility row closure unless the map and guide explicitly allow it.
- `required_for_closure=false` on current FE-P7 targets means the map must be promoted before passing targets can close the rows.

## FE-P6 Handoff Inputs

FE-P7 may inherit FE-P6 as dependency context only.

- Final FE-P6 registry status: `active`, `active_green`.
- FE-P0 through FE-P6 dependency state: all `active`, `active_green` in the frontend registry.
- FE-P6 direct evidence roots recorded by `FRONTEND_PHASE6_IMPLEMENTATION_PLAN.md`:
  - `FE-U-P6-01`: `.cartulary/test-results/20260607T183324Z-p3150835/frontend-unit/frontend-row-accounting.json`
  - `FE-I-P6-01`: `.cartulary/test-results/20260607T183324Z-p3150835/frontend-unit/frontend-row-accounting.json` and `.cartulary/test-results/20260607T183324Z-p3150835/browser-e2e-webserver-backed/frontend-row-accounting.json`
  - `FE-E-P6-01`: `.cartulary/test-results/20260607T183324Z-p3150835/browser-e2e-webserver-backed/frontend-row-accounting.json` and `.cartulary/test-results/20260607T183324Z-p3150835/browser-e2e-stateful/frontend-row-accounting.json`
  - `FE-V-P6-01`: `.cartulary/test-results/20260607T175905Z-p2973141/browser-e2e-visual/frontend-row-accounting.json`
  - `FE-A11Y-P6-01`: `.cartulary/test-results/20260607T180443Z-p2986830/browser-e2e-a11y/frontend-row-accounting.json`
- FE-P6 public evidence-handle status is dependency context only.
- FE-P6 visual and accessibility status is design-direction context only.
- FE-P6 retained full check context: `.cartulary/test-results/20260607T183324Z-p3150835`; finalized context `.cartulary/test-results/20260607T184052Z-p3212967`.
- FE-P6 preflight context: `.cartulary/test-results/20260607T181024Z-p3000172`; diagnostic only for later phases.
- Unresolved FE-P6 blockers affecting FE-P7: none found in the current frontend registry. FE-P7-specific blockers remain in the FE-P7 registry and map.

If any FE-P6 handoff root is reused, the minimum follow-up is `make explain-run RESULTS_DIR=<run-root>` plus row-accounting inspection proving current map digests and direct FE-P7 row IDs. Without that proof, FE-P6 artifacts remain context only.

## Phase Objective

FE-P7 is implementation and verification of live collaboration and conflicts in the frontend through current row-owned evidence. The objective includes:

- WebSocket stream handling under `/ws/v1/`.
- Event reducer behavior for row update, reset, invalidate, stale-row requery, authorization close, and session revocation states.
- Presence anchoring by stable row, cell, and surface identity.
- Live row refresh without visible-index identity leakage.
- Same-field conflict display and resolver behavior.
- Conflict anchors based on stable identifiers, not visible grid position.
- Public mutation contracts for resolver actions.
- Focus and pending queue preservation.

## Implementation Scope

In-scope implementation surfaces:

- `/apps/web`
- `/packages/grid-adapter`
- `/packages/protocol-ts`
- `/packages/ui-contracts`
- `/packages/test-utils`
- `apps/web/e2e`
- Current FE-P7-relevant files discovered during inspection:
  - `apps/web/src/WorkbookShell.tsx`
  - `apps/web/src/workbookPendingQueue.ts`
  - `apps/web/src/workbookShellPhase4.ts`
  - `apps/web/src/timelineWorkbookTestSupport.ts`
  - `apps/web/src/WorkbookShell.phase7.test.tsx`
  - `apps/web/e2e/phase6.collaboration.spec.ts`
  - `apps/web/e2e/phase6Harness.ts`
  - `apps/web/e2e/workbook.visual.spec.ts`
  - `apps/web/e2e/workbook.a11y-preflight.spec.ts`
  - `apps/web/e2e/a11yPhaseMap.ts`
  - `packages/grid-adapter/src/core.ts`
  - `packages/grid-adapter/src/index.tsx`
  - `packages/protocol-ts/src/index.ts`
  - `packages/ui-contracts/src/index.ts`
  - `packages/test-utils/src/index.ts`

Expected implementation domains:

- WebSocket connection and lifecycle handling.
- Reducer state model.
- Presence state model.
- Stale-row and invalidate handling.
- Conflict state model.
- Resolver actions.
- Save-state presentation.
- Selector/test-id coverage.
- Browser helper coverage.
- Visual and accessibility fixture readiness.

## Out of Scope

- Saved-view persistence.
- Inspector rollback.
- Full coordination surfaces.
- FE-P8 sorting/filtering/grouping/layout persistence.
- FE-P10 coordination surfaces.
- Core 05 claim publication.
- Any new backend route semantics not already owned by Core 00 through Core 04 or an adopted NLSpec.
- Broad redesign of the workbook shell unless directly required by FE-P7 evidence.

## Row Inventory

| Row ID | Layer | Evidence class | Owner sections | Exact REQ IDs | Exact AC IDs | Repository target or TODO | Current `claim_status` | Expected closure evidence | Blockers | Non-claims |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `FE-U-P7-01` | unit | product conformance | `docs/spec/01_architecture_storage_and_view_contracts.md#Core 01 Section 3.3.10`; `docs/spec/04_security_deployment_and_conformance.md#Core 04 Sections 2 and 5` | `REQ-01-250`, `REQ-01-253`, `REQ-04-001`, `REQ-04-017`, `REQ-04-052`, `REQ-04-053`, `REQ-04-110` | `AC-129`, `AC-131`, `AC-136`, `AC-156`, `AC-163`, `AC-231`, `AC-232`, `AC-233`, `AC-234`, `AC-298`, `AC-334`, `AC-342` | `make frontend-unit`; TODO add/promote row-owned unit evidence | `blocked` | `frontend-row-accounting.json` for `frontend-unit` with `FE-U-P7-01`, current map/registry digests, WebSocket reducer cases for row update, reset, invalidate, stale-row requery, authorization close, and session revocation | `FE-U-P7-01-BLOCKER-01: frontend_phase_row_not_implemented` | Does not claim public route behavior, design readiness, or Core 05 publication. |
| `FE-U-P7-02` | unit | product conformance | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md#Core 03 Sections 3.1, 3.3.5, 3.3.6, 4.2, and 4.4` | `REQ-03-033`, `REQ-03-084` | `AC-009`, `AC-013`, `AC-037`, `AC-042`, `AC-047`, `AC-126`, `AC-163`, `AC-203`, `AC-204`, `AC-226`, `AC-231`, `AC-381` | `make frontend-unit`; TODO add/promote row-owned unit evidence | `blocked` | `frontend-row-accounting.json` for `frontend-unit` with `FE-U-P7-02`, current digests, same-field conflict anchors, queue, resolver state, and `record_id + field_key + base_row_version` identity assertions | `FE-U-P7-02-BLOCKER-01: frontend_phase_row_not_implemented` | Does not claim public resolver integration, visual readiness, or Core 05 publication. |
| `FE-I-P7-01` | integration | product conformance | `docs/spec/01_architecture_storage_and_view_contracts.md#Core 01 Section 3.3.6`; `docs/spec/03_workbook_interaction_collaboration_and_workflows.md#Core 03 Sections 3.1, 3.3.5, 3.3.6, 4.2, and 4.4` | `REQ-01-057`, `REQ-01-070`, `REQ-03-033`, `REQ-03-084` | `AC-009`, `AC-013`, `AC-037`, `AC-042`, `AC-047`, `AC-124`, `AC-127`, `AC-181`, `AC-183`, `AC-188`, `AC-190`, `AC-200`, `AC-218`, `AC-221`, `AC-231`, `AC-299`, `AC-381` | `make frontend-unit`; `make browser-e2e-webserver-backed`; TODO add/promote row-owned integration scenario | `blocked` | Row accounting for `FE-I-P7-01` proving conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering | `FE-I-P7-01-BLOCKER-01: frontend_phase_row_not_implemented` | Does not claim multi-client E2E closure, visual readiness, or Core 05 publication. |
| `FE-E-P7-01` | e2e | product conformance | `docs/spec/01_architecture_storage_and_view_contracts.md#Core 01 Sections 3.3.6 and 3.3.10`; `docs/spec/03_workbook_interaction_collaboration_and_workflows.md#Core 03 Sections 3.1, 3.3.5, 3.3.6, 4.2, and 4.4` | `REQ-01-057`, `REQ-01-070`, `REQ-01-250`, `REQ-01-253`, `REQ-03-033`, `REQ-03-084` | `AC-009`, `AC-013`, `AC-037`, `AC-042`, `AC-047`, `AC-124`, `AC-129`, `AC-131`, `AC-136`, `AC-156`, `AC-163`, `AC-181`, `AC-183`, `AC-188`, `AC-190`, `AC-200`, `AC-218`, `AC-221`, `AC-233`, `AC-299`, `AC-381` | `make browser-e2e-stateful`; `make browser-e2e-webserver-backed`; TODO add/promote row-owned multi-client scenario | `blocked` | Row accounting for `FE-E-P7-01` proving multi-client live row update, presence anchoring, reset/invalidate, stale-row requery, and same-field resolver through `/ws/v1/` and `/api/v1/` | `FE-E-P7-01-BLOCKER-01: frontend_phase_row_not_implemented` | Does not claim visual/a11y readiness or Core 05 publication. |
| `FE-V-P7-01` | visual | design direction | `docs/guides/cartulary-ui-ux-design-guide.md#UI/UX guide Sections 8, 10.4, 10.5, 13`; `docs/guides/cartulary_visual_golden_maintenance.md#visual golden guide Sections 2, 3, 5` | none | `R2-AC-023`, `R2-AC-026`, `R2-AC-045`, `R2-AC-050`, `R2-AC-073`, `R2-AC-079` | `make browser-e2e-visual`; fixture IDs `FE-VFIX-03`, `FE-VFIX-04`, `FE-VFIX-08`; TODO recapture/promote FE-P7 row-owned visual accounting | `blocked` | Row accounting for `FE-V-P7-01` plus visual artifacts for same-field conflict, row-gutter presence, presence hints, conflict resolver, reset/invalidate notice, and save-state conflict fixtures | `FE-V-P7-01-BLOCKER-01: visual_fixture_not_recaptured_for_frontend_row` | Design-direction only; does not claim product conformance or Core 05 publication. |
| `FE-A11Y-P7-01` | accessibility | design direction | `docs/guides/cartulary-ui-ux-design-guide.md#UI/UX guide Sections 10.4, 10.5, 14` | none | `R2-AC-045`, `R2-AC-050`, `R2-AC-080`, `R2-AC-086`, `D-AC-009`, `D-AC-012` | Live target `make browser-e2e-a11y-preflight`; TODO resolve whether closure target must be `browser-e2e-a11y` or preflight can remain smoke-only | `blocked` | Row accounting for `FE-A11Y-P7-01` plus accessibility summary/preflight artifacts proving accessible name/state communication for conflict state, resolver controls, presence hint, stale-row notice, and save-state conflict | `FE-A11Y-P7-01-BLOCKER-01: frontend_phase_row_not_implemented`; preflight target has no phase coverage | Design-direction only; does not claim product conformance or Core 05 publication. |

## Evidence Layer Matrix

| Evidence class | Allowed evidence sources | Forbidden evidence substitutions | Targets | Closure rule | Failure/blocker rule |
| --- | --- | --- | --- | --- | --- |
| Product conformance | Core 00-04 owner refs, live FE-P7 map, current mapped target artifacts, `frontend-row-accounting.json`, public `/api/v1/` and `/ws/v1/` evidence where mapped | Visual snapshots, accessibility smoke, generated ledgers, prior plans, old retained runs without current row ownership, frontend-only mocks for public route rows | `frontend-unit`, `browser-e2e-webserver-backed`, `browser-e2e-stateful` | Row passes only when current map marks the row implemented/closure-eligible and row accounting names the row, target, scenario/unit, owner digests, and pass status | Block when public-route behavior is mock-only, map status is blocked, row accounting is missing/stale, or stable identity assertions are absent. |
| Design direction | UI/UX guide, `docs/design.md`, visual golden guide, row map design rows, design ACs | Product conformance claims, Core 05 publication claims, generated ledger closure | `browser-e2e-visual`, `browser-e2e-a11y`, `browser-e2e-a11y-preflight` when mapped | Passes only as design readiness for the row's mapped evidence class | Block when design evidence is represented as product conformance or fixture status is treated as closure. |
| Implementation support | Dev guide, implementation/testing guide, package tests, helper tests, Make explanations, code inspection | Product behavior authority, row closure by helper-only tests, route semantics not owned by Core/NLSpec | `frontend-typecheck`, `frontend-import-boundary-check`, helper/package tests when selected by target | Supports implementation readiness but does not close product rows unless mapped as primary evidence | Block when support-only evidence is substituted for row-owned primary evidence. |
| Visual readiness | Visual fixture registry, visual row map, `browser-e2e-visual`, visual snapshots, visual row accounting | Product conformance, claim publication, broad `make check`, snapshot filename existence alone | `browser-e2e-visual` | `FE-V-P7-01` passes only with current row-owned visual accounting and required fixture artifacts | Block when fixtures are current in registry but not recaptured/accounted for FE-P7. |
| Accessibility readiness | UI/UX guide, design ACs, accessibility row map, `browser-e2e-a11y` summary or preflight target if live map explicitly owns it | Product conformance, color-only assertions, visual snapshots, preflight smoke as closure without map authority | Live map currently names `browser-e2e-a11y-preflight`; guide says implemented a11y rows should close through `browser-e2e-a11y` summary | `FE-A11Y-P7-01` passes only when map, guide, target, and row accounting agree on closure evidence | Block current ambiguity between blocked-row preflight smoke and implemented accessibility closure. |
| Claim-publication boundary | Core 05 benchmark/profile/manifest metadata and publication criteria | Any FE-P7 plan, row map, visual snapshot, accessibility artifact, or broad check without Core 05 metadata | `benchmark-claim-check` only when explicit Core 05 claim exists | Inactive for FE-P7 unless explicit claim-bearing metadata exists and satisfies Core 05 | Block when publication readiness is implied without claim metadata. |

## Dependencies And Prerequisites

- FE-P0 through FE-P6 must remain `active_green` in `tools/frontend_phase_registry.json`.
- FE-P7 map, registry, generated ledger, and guide freshness digests must remain consistent.
- `make phase-ledger-drift` must pass after any row map or ledger owner input changes.
- `make generate-drift` and `make generated-artifact-policy-check` must pass when generated protocol or selector roots are affected.
- Service-backed browser readiness is required before `browser-e2e-stateful`, `browser-e2e-webserver-backed`, `browser-e2e-visual`, `browser-e2e-a11y`, or `browser-e2e-a11y-preflight` can supply evidence.
- WebSocket harness readiness requires live `/ws/v1/incidents/{incident_id}` observation and frame parsing for hello/resume, presence, row changes, reset, revoke, close, and sequence behavior.
- Selector/test-id builder readiness requires stable selectors from `packages/ui-contracts`, especially `record_id` and `field_key` based selectors.
- Protocol generation drift status must stay clean because `/api/v1/` envelopes and `/ws/v1/` contracts flow through `packages/protocol-ts` generated artifacts.
- Visual fixture registry status for `FE-VFIX-03`, `FE-VFIX-04`, and `FE-VFIX-08` must be checked before visual recapture.
- Accessibility summary or preflight behavior must match the live row map; current `browser-e2e-a11y-preflight` target has no phase coverage and cannot be assumed to close implemented accessibility rows.

## Shared Harness Analysis

| Harness | Owner | Target | Relevant FE-P7 rows | Artifact expectation | Blocker rule |
| --- | --- | --- | --- | --- | --- |
| WebSocket event reducer behavior | `docs/testing-harness-nlspec.md` for mechanics; Core 01/Core 04 for behavior | `frontend-unit`; `browser-e2e-stateful`; `browser-e2e-webserver-backed` | `FE-U-P7-01`, `FE-E-P7-01` | Unit reducer results and browser row accounting with `/ws/v1/` message evidence | Block if tested outside `/ws/v1/` without owner permission or if reset/invalidate/stale/revocation cases are absent. |
| Same-field conflict anchoring | Core 03, UI contracts, grid adapter | `frontend-unit`; `browser-e2e-webserver-backed` | `FE-U-P7-02`, `FE-I-P7-01`, `FE-E-P7-01`, `FE-V-P7-01` | Row accounting plus assertions for `record_id + field_key + base_row_version`; no visible-index reliance | Block if identity relies on visible row index, DOM order, labels, or vendor coordinates. |
| Presence anchoring | Core 01 WebSocket stream, UI contracts, test-utils anchor helper | `browser-e2e-stateful`; `browser-e2e-webserver-backed`; `browser-e2e-visual` | `FE-E-P7-01`, `FE-V-P7-01`, `FE-A11Y-P7-01` | Browser artifacts proving row and cell presence markers share stable row/cell identity | Block if presence only appears visually or lacks stable selector/identity assertions. |
| Save-state presentation | Core 03 save state, UI/UX guide, pending queue model | `frontend-unit`; `browser-e2e-webserver-backed`; `browser-e2e-visual`; accessibility target | `FE-U-P7-02`, `FE-I-P7-01`, `FE-V-P7-01`, `FE-A11Y-P7-01` | Row accounting and artifacts for `Syncing`, `Saved`, `Conflict`, pending notice, and conflict anchors | Block if visual/accessibility state is treated as product conformance. |
| Frontend route/API boundary conformance | Core 01 public envelopes, Core 03 conflict behavior, Core 04 auth/session | `browser-e2e-webserver-backed`; `browser-e2e-stateful` | `FE-I-P7-01`, `FE-E-P7-01` | Public route traffic through `/api/v1/` and `/ws/v1/`, success/error envelope checks, row refresh | Block if public-route behavior is proven only by mocks or private helpers. |
| Browser command helpers | `packages/test-utils`, `apps/web/e2e/phase6Harness.ts` | Browser targets that include FE-P7 scenarios | `FE-I-P7-01`, `FE-E-P7-01`, `FE-V-P7-01`, `FE-A11Y-P7-01` | Helper evidence in target artifacts and stable helper APIs | Block if helper-only tests are substituted for mapped row accounting. |
| Visual-regression fixtures | Visual fixture registry and visual golden guide | `browser-e2e-visual` | `FE-V-P7-01` | Visual artifact set for fixture IDs `FE-VFIX-03`, `FE-VFIX-04`, `FE-VFIX-08` plus row accounting | Block if registry `current` status or snapshot filenames are claimed as closure without FE-P7 row accounting. |
| Accessibility state names and ARIA | UI/UX guide, `docs/design.md`, accessibility map | Live map currently `browser-e2e-a11y-preflight`; implemented rows normally require `browser-e2e-a11y` summary | `FE-A11Y-P7-01` | Accessibility artifact proving names/states for conflict resolver, presence hints, stale notices, save-state conflict | Block if color-only evidence or blocked-row preflight smoke is used as closure without map authority. |

## Public Interfaces And Deliverables

FE-P7 must use or verify these public interfaces and deliverables:

- `/ws/v1/` WebSocket stream, currently `/ws/v1/incidents/{incident_id}` in the app.
- `/api/v1/` public mutation/query surfaces used by resolver actions, including record patching and conflict resolution.
- Generated protocol types and contract artifact access through `packages/protocol-ts`.
- Public success and error envelopes from Core 01.
- Stable identifiers: `record_id`, `field_key`, `base_row_version`, `row_version`, `view_schema_id`, `client_txn_id`, `client_instance_id`, stream sequence, and connection identity.
- Selector/test-id contracts from `packages/ui-contracts`.
- Row-accounting artifacts: `frontend-row-accounting.json`, target summaries, phase summaries, and run roots.
- Accessibility summary/preflight artifacts where mapped.
- Visual artifacts and fixture registry entries for FE-P7 fixture IDs.

This plan does not specify new route behavior unless already owned by Core 00 through Core 04.

## Sprint Checklist

- [ ] Readiness: registry, map, ledger, fixture registry, owner refs, generated policy, and target explanations are current.
- [ ] Implementation: WebSocket reducer, conflict reducer, resolver integration, E2E route proof, visual fixtures, and accessibility scenarios are mapped to FE-P7 rows.
- [ ] Evidence: each row has direct row-accounting artifacts with current digests and target identity.
- [ ] Validation: row targets pass, drift targets pass, finalization is handled with or without `RESULTS_DIR`.
- [ ] Blockers: every blocked row has exact observed condition and minimum follow-up.
- [ ] Handoff: FE-P8 receives only current FE-P7 evidence roots and explicit non-claims.

## Sprint-by-Sprint Execution Plan

### Sprint 1: Readiness And Drift Baseline

- Goal: Establish the live baseline for FE-P7 before implementation.
- Source inputs: frontend registry, FE-P7 map, generated ledger, visual fixture registry, owner specs, frontend guide, harness NLSpec, dev guide, existing code, target explanations.
- Implementation tasks: none beyond inspection; do not edit generated artifacts.
- Validation commands: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P7`; `make phase-ledger-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift`.
- Expected artifacts: target summaries and phase summaries under `.cartulary/test-results/<run-id>/`.
- Row-accounting expectations: no row should be claimed closed during readiness; current row accounting is expected to be absent for FE-P7.
- Blockers: current registry activation blocker and all six row blockers remain until implementation evidence and map promotion are added.
- Non-claims: readiness does not implement behavior, close rows, or activate Core 05.
- Binary completion criteria: pass if live status, digests, blockers, fixture IDs, code surfaces, target explanations, and drift results are recorded; fail if any owner ref is missing or contradictory without blocker entry.

#### Sprint 1 Closeout: Readiness And Drift Baseline

Sprint 1 completed as a baseline-and-correction sprint. It did not implement product behavior, promote any FE-P7 row, change public APIs, change selectors, update visual goldens, edit generated artifacts, or claim accessibility/product/Core 05 closure.

Current live baseline recorded during the Sprint 1 run:

- `FE-P7` remains `planned`, with `row_rollup_state=no_rows_implemented`.
- `FE-P7-ACTIVATION-BLOCKER-01` remains active with reason `frontend_phase_not_active`.
- All six FE-P7 rows remain `claim_status=blocked`.
- Product-conformance rows are `FE-U-P7-01`, `FE-U-P7-02`, `FE-I-P7-01`, and `FE-E-P7-01`; they require current mapped row-owned evidence and `frontend-row-accounting.json`.
- Design-direction rows are `FE-V-P7-01` and `FE-A11Y-P7-01`; visual and accessibility evidence must not be counted as product conformance or Core 05 publication evidence.
- The generated FE-P7 coverage ledger is current downstream status only and is not row, target, phase, or sprint closure evidence.
- Existing FE-P6 collaboration, visual, accessibility, selector, and helper artifacts overlap FE-P7 behavior but remain dependency context unless current FE-P7 row ownership and freshness are proven.

Inspected live inputs:

- Governing sources: Core 00 through Core 05, `docs/testing-harness-nlspec.md`, `docs/domain.md`, `docs/design.md`, the frontend implementation/testing guide, and the visual golden maintenance guide.
- FE-P7 sources: `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p7_test_map.json`, `docs/testing/frontend_phase_coverage_ledgers/fe_p7_coverage_ledger.md`, and `tools/frontend_visual_fixture_registry.json`.
- Examples only: prior frontend phase implementation plans.
- Implementation-support sources: relevant app, E2E, selector, helper, grid-adapter, and protocol facade code under `apps/web` and `packages`.
- Retained artifacts under `.cartulary/test-results`, including frontend row-accounting files. No direct FE-P7 row-owned evidence root was found.

Fresh Sprint 1 command evidence:

| Command | Result | Evidence |
| --- | --- | --- |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P7` | pass | FE-P7 is planned, blocked, explainable, and not executable as a phase. |
| `make task-guide ROLE=feature-dev PHASE=FE-P7` | fail | Frontend phases require `PHASE_NAMESPACE=frontend`; this is an invocation requirement, not product drift. |
| `make task-guide ROLE=feature-dev PHASE_NAMESPACE=frontend PHASE=FE-P7` | pass | Recommended Sprint 1 path is `explain-phase` and `phase-ledger-drift`. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260610T000156Z-p69566` |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260610T000218Z-p69964` |
| `make json-shape-check` | pass | `.cartulary/test-results/20260610T000227Z-p70160` |
| `make generate-drift` | pass | `.cartulary/test-results/20260610T000234Z-p70503` |
| `make agent-finalize` | pass | `.cartulary/test-results/20260610T000247Z-p71387`; generated unchanged; `RESULTS_DIR` unset, so retained-run maintenance was skipped. |

Sprint 1 blockers and unresolved evidence gaps:

- Row, target, phase, and sprint implementation closure remain blocked because FE-P7 is not active, all rows are blocked, and current FE-P7 targets have `required_for_closure=false`.
- No direct FE-P7 row-owned `frontend-row-accounting.json` evidence exists in retained local artifacts.
- Retained FE-P6 row accounting is stale or unowned for FE-P7 closure even when it exercises overlapping collaboration behavior.
- `FE-VFIX-03`, `FE-VFIX-04`, and `FE-VFIX-08` are current fixture-registry entries, but fixture status and V-6/FE-P4 golden names cannot close `FE-V-P7-01`.
- `FE-A11Y-P7-01` maps to `browser-e2e-a11y-preflight`, while both `browser-e2e-a11y` and `browser-e2e-a11y-preflight` currently explain with no phase coverage and no latest artifact. Preflight smoke cannot be used as implemented-row closure without map and guide authority.
- Core 05 claim-publication evidence remains inactive because no explicit FE-P7 claim-bearing metadata was found.

Tracked-file correction status:

- Baseline command execution changed no tracked files.
- This Sprint 1 closeout section is the only tracked correction for the sprint.
- No implementation, tests, maps, registries, generated artifacts, docs/spec authority sources, visual goldens, or lockfiles were changed.

Post-correction validation after this planning-artifact edit:

| Command | Result | Evidence |
| --- | --- | --- |
| `git diff --check` | pass | No whitespace errors. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260610T001206Z-p74028` |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260610T001214Z-p74384` |
| `make json-shape-check` | pass | `.cartulary/test-results/20260610T001218Z-p74568` |
| `make generate-drift` | pass | `.cartulary/test-results/20260610T001222Z-p74899` |
| `make agent-finalize` | pass | `.cartulary/test-results/20260610T001230Z-p75769`; generated unchanged; `RESULTS_DIR` unset, so retained-run maintenance was skipped. |

Binary Sprint 1 outcome: `pass with blockers`. Readiness and drift baseline are established; FE-P7 row, target, phase, and implementation closure remain blocked.

### Sprint 2: WebSocket Reducer And Lifecycle

- Goal: Implement and validate `FE-U-P7-01`.
- Source inputs: Core 01 Section 3.3.10, Core 04 Sections 2 and 5, `apps/web/src/WorkbookShell.tsx`, `apps/web/src/workbookShellPhase4.ts`, `apps/web/src/timelineWorkbookTestSupport.ts`.
- Implementation tasks: isolate reducer state cases for row update, reset, invalidate, stale-row requery request, authorization close, session revocation, duplicate sequence, and sequence gap. Add FE-P7 unit scenario identity and row accounting.
- Validation commands: `make explain-target TARGET=frontend-unit DETAIL=summary`; `make frontend-unit`.
- Expected artifacts: `frontend-unit/target-summary.json`, `frontend-unit/<phase-label>/phase-summary.json`, `frontend-unit/frontend-row-accounting.json`.
- Row-accounting expectations: `FE-U-P7-01` appears with `product_conformance`, target `frontend-unit`, current map digest, current registry digest, current guide digest, and pass status.
- Blockers: block if tests only simulate a private socket model without owner-approved mapping to `/ws/v1/` semantics.
- Non-claims: unit evidence alone does not close public route E2E behavior.
- Binary completion criteria: pass if the row is implemented in the live map and `frontend-row-accounting.json` directly closes `FE-U-P7-01`; fail otherwise.

### Sprint 3: Conflict Anchoring And Resolver State

- Goal: Implement and validate `FE-U-P7-02`.
- Source inputs: Core 03 Sections 3.1, 3.3.5, 3.3.6, 4.2, 4.4; `apps/web/src/workbookPendingQueue.ts`; `apps/web/src/WorkbookShell.tsx`; `packages/ui-contracts/src/index.ts`; `packages/grid-adapter/src/core.ts`.
- Implementation tasks: assert same-field conflict anchors, queue state, resolver state, save-state conflict anchors, and focus restoration use `record_id + field_key + base_row_version`, not visible indexes or DOM order.
- Validation commands: `make explain-target TARGET=frontend-unit DETAIL=summary`; `make frontend-unit`.
- Expected artifacts: `frontend-unit/frontend-row-accounting.json` with row-level evidence and unit identifiers.
- Row-accounting expectations: `FE-U-P7-02` appears with exact scenario/unit identity and current digests.
- Blockers: block if anchor identity can be satisfied by visible row index, column label, DOM order, or grid vendor coordinates.
- Non-claims: unit state coverage does not prove public mutation submission.
- Binary completion criteria: pass if row-owned unit evidence proves stable conflict identity and the live map promotes the row; fail otherwise.

### Sprint 4: Resolver Integration

- Goal: Implement and validate `FE-I-P7-01`.
- Source inputs: Core 01 Section 3.3.6, Core 03 conflict sections, public envelope contracts, `apps/web/src/WorkbookShell.tsx`, `apps/web/e2e/phase6Harness.ts`.
- Implementation tasks: verify resolver actions submit public `/api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` mutations, handle success/error envelopes, refresh row state, preserve focus, and preserve pending queue ordering.
- Validation commands: `make explain-target TARGET=frontend-unit DETAIL=summary`; `make frontend-unit`; `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`; `make browser-e2e-webserver-backed`.
- Expected artifacts: row accounting from both mapped targets if both remain required by the promoted map.
- Row-accounting expectations: scenario title exactly matches the live map when `scenario_title_required=true`: `FE-I-P7-01 Verify conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering.`
- Blockers: block if resolver evidence is from frontend mocks only or if row refresh/focus/pending queue ordering are not asserted.
- Non-claims: resolver integration does not prove multi-client WebSocket behavior.
- Binary completion criteria: pass if public mutation behavior and row refresh are proven by mapped row-owned artifacts; fail otherwise.

### Sprint 5: Multi-Client Browser E2E

- Goal: Implement and validate `FE-E-P7-01`.
- Source inputs: Core 01 Sections 3.3.6 and 3.3.10, Core 03 conflict sections, Core 04 auth/session behavior, `apps/web/e2e/phase6.collaboration.spec.ts`, `apps/web/e2e/phase6Harness.ts`, `packages/test-utils`.
- Implementation tasks: create or promote FE-P7 multi-client scenarios covering live row update, presence anchoring, reset/invalidate handling, stale-row requery, same-field resolver, and auth/session revocation through `/ws/v1/` and `/api/v1/`.
- Validation commands: `make explain-target TARGET=browser-e2e-stateful DETAIL=summary`; `make browser-e2e-stateful`; `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`; `make browser-e2e-webserver-backed`.
- Expected artifacts: browser target summaries, Playwright artifacts, WebSocket frame evidence, public route evidence, `frontend-row-accounting.json`.
- Row-accounting expectations: scenario title exactly matches the live map when required: `FE-E-P7-01 Verify multi-client live row update, presence anchoring, reset/invalidate handling, stale-row requery, and same-field conflict resolver through /ws/v1/ and /api/v1/.`
- Blockers: block if WebSocket evidence is not under `/ws/v1/`, if public mutations are not under `/api/v1/`, or if marker anchoring lacks stable identity checks.
- Non-claims: E2E product evidence does not claim visual or accessibility readiness.
- Binary completion criteria: pass if both mapped browser targets have current row-owned accounting and no route/identity blocker remains; fail otherwise.

### Sprint 6: Visual Readiness

- Goal: Implement and validate `FE-V-P7-01`.
- Source inputs: UI/UX guide Sections 8, 10.4, 10.5, 13; visual golden guide Sections 2, 3, 5; live fixture registry; `apps/web/e2e/workbook.visual.spec.ts`.
- Implementation tasks: verify or recapture FE-P7 visual fixtures for same-field conflict, row-gutter presence, presence hint, conflict resolver, reset/invalidate notice, and save-state conflict. Live fixture candidates are `FE-VFIX-03`, `FE-VFIX-04`, and `FE-VFIX-08`.
- Validation commands: `make explain-target TARGET=browser-e2e-visual DETAIL=summary`; `make browser-e2e-visual`.
- Expected artifacts: `browser-e2e-visual/target-summary.json`, visual phase summaries, snapshot artifacts, and `browser-e2e-visual/frontend-row-accounting.json`.
- Row-accounting expectations: `FE-V-P7-01` appears as `design_direction`, not product conformance, with current fixture IDs and map digests.
- Blockers: current blocker `visual_fixture_not_recaptured_for_frontend_row`; block if V-6 filenames or registry `current` status are treated as closure without FE-P7 row accounting.
- Non-claims: visual readiness does not claim product conformance or Core 05 publication.
- Binary completion criteria: pass if row-owned visual accounting and expected fixture artifacts exist for FE-P7; fail otherwise.

### Sprint 7: Accessibility Readiness

- Goal: Implement and validate `FE-A11Y-P7-01`.
- Source inputs: UI/UX guide Sections 10.4, 10.5, 14; `docs/design.md`; live FE-P7 map; `apps/web/e2e/workbook.a11y-preflight.spec.ts`; `apps/web/e2e/a11yPhaseMap.ts`.
- Implementation tasks: verify conflict state, resolver controls, presence hints, stale-row notice, and save-state conflict communication by accessible name/state, not color alone. Resolve target ambiguity before claiming closure.
- Validation commands: live map target is `make browser-e2e-a11y-preflight`; also inspect `make explain-target TARGET=browser-e2e-a11y DETAIL=summary` before promoting implemented accessibility closure. Run `make browser-e2e-a11y-preflight` only as mapped smoke unless the map/guide promote closure.
- Expected artifacts: preflight target artifacts if smoke remains mapped; for implemented closure, accessibility row accounting and `cartulary.frontend_accessibility_summary.v2` from `browser-e2e-a11y` if guide requirements still apply.
- Row-accounting expectations: no closure from blocked-row smoke unless current map and guide explicitly allow it. Current `browser-e2e-a11y-preflight` has no phase coverage.
- Blockers: block if preflight smoke is used as closure without map authority, or if state is conveyed by color only.
- Non-claims: accessibility readiness is design-direction only and not product conformance.
- Binary completion criteria: pass if the row map, guide, target, accessibility artifact, and row accounting agree; fail otherwise.

### Sprint 8: Closure, Final Validation, And FE-P8 Handoff

- Goal: Audit FE-P7 closure, refresh drift checks, run final validation, and prepare FE-P8 handoff.
- Source inputs: all row accounting roots, frontend registry, FE-P7 map, generated ledger, visual fixture registry, target summaries, finalization artifacts.
- Implementation tasks: verify every row status, remove or retain blockers with exact evidence, update owner inputs only if required, regenerate ledgers through Make if maps changed, and record FE-P8 strict non-claims.
- Validation commands: `make phase-ledger-drift`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; row targets; `make check`; `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when retained-run maintenance is used.
- Expected artifacts: final row accounting for each closed row, drift summaries, check scheduler summaries, optional finalization summary.
- Row-accounting expectations: all closed rows have current direct row accounting; blocked rows have exact blocker entries.
- Blockers: block full phase completion if any row remains blocked, if FE-P7 is not `active_green`, or if drift/finalization fails.
- Non-claims: FE-P8 cannot inherit FE-P7 claims without current direct evidence roots.
- Binary completion criteria: pass if every required row is closed or explicitly owner-accepted as blocked and registry/ledger/drift/finalization outcomes are recorded; fail otherwise.

## Validation Commands

Explanation commands:

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P7`
- `make explain-target TARGET=frontend-unit DETAIL=summary`
- `make explain-target TARGET=browser-e2e-stateful DETAIL=summary`
- `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`
- `make explain-target TARGET=browser-e2e-visual DETAIL=summary`
- `make explain-target TARGET=browser-e2e-a11y DETAIL=summary`
- `make explain-target TARGET=browser-e2e-a11y-preflight DETAIL=summary`
- `make explain-target TARGET=phase-ledger-drift DETAIL=summary`
- `make explain-target TARGET=generate-drift DETAIL=summary`
- `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary`
- `make explain-target TARGET=json-shape-check DETAIL=summary`
- `make explain-target TARGET=check DETAIL=summary`
- `make explain-target TARGET=agent-finalize DETAIL=summary`
- `make explain-target TARGET=frontend-typecheck DETAIL=summary`
- `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`

Execution commands for implementation and closure:

- `make frontend-unit`
- `make browser-e2e-stateful`
- `make browser-e2e-webserver-backed`
- `make browser-e2e-visual`
- `make browser-e2e-a11y-preflight` for the current live `FE-A11Y-P7-01` mapping, with the preflight closure caveat above
- `make browser-e2e-a11y` only if the live map or guide promotes implemented accessibility closure to the accessibility batch
- `make phase-ledger-drift`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make check`
- `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when using retained successful full warm check evidence
- `make agent-finalize` without `RESULTS_DIR` only when retained-run maintenance is intentionally skipped and the final report records that reason

Absent command records:

- `TODO: command not present`: `make frontend-phase-drift`
- `TODO: command not present`: `make generated-drift`; use `make generate-drift`
- `TODO: command not present`: `make lint-docs`
- `TODO: command not present`: `make format-docs`

## Evidence Requirements

For every FE-P7 row, the minimum evidence payload is:

- Row accounting schema and row ID.
- Target identity and command ID.
- Accounting scope.
- Evidence class.
- Scenario or unit identity.
- Owner-map digest or freshness evidence.
- Registry digest and guide digest.
- Pass/fail status.
- Closure status.
- Artifact root.
- Exact blocker when missing.

Per-row artifact requirements:

| Row | Required artifact family | Minimum evidence payload |
| --- | --- | --- |
| `FE-U-P7-01` | `frontend-unit` row accounting and unit artifacts | Reducer cases for row update, reset, invalidate, stale requery, auth close, session revocation, duplicate/gap stream handling; current map/registry/guide digests; pass status. |
| `FE-U-P7-02` | `frontend-unit` row accounting and unit artifacts | Conflict anchor identity, conflict queue state, save-state conflict anchors, resolver state, focus preservation; explicit rejection of visible index/DOM order identity; pass status. |
| `FE-I-P7-01` | `frontend-unit` plus `browser-e2e-webserver-backed` row accounting if both remain mapped | Public resolver mutation, success/error envelopes, row refresh, focus return, pending queue ordering, current scenario title, pass status. |
| `FE-E-P7-01` | `browser-e2e-stateful` and `browser-e2e-webserver-backed` row accounting | Multi-client `/ws/v1/` frames, `/api/v1/` mutations/queries, presence anchoring, reset/invalidate, stale requery, same-field resolver, session revocation behavior, pass status. |
| `FE-V-P7-01` | `browser-e2e-visual` row accounting, visual snapshots, fixture registry IDs | Fixture IDs `FE-VFIX-03`, `FE-VFIX-04`, `FE-VFIX-08` or updated live IDs; same-field conflict, row-gutter presence, presence hints, resolver, reset/invalidate notice, save-state conflict; design-direction pass status. |
| `FE-A11Y-P7-01` | Current mapped accessibility/preflight row accounting and accessibility summary when applicable | Conflict state, resolver controls, presence hint, stale-row notice, save-state conflict accessible names/states; no color-only reliance; explicit closure rule for preflight versus `browser-e2e-a11y`. |

## Blocker Rules

A blocker must be recorded when:

- Owner refs are missing, unresolved, stale, or contradictory.
- Registry, map, ledger, or row-accounting freshness fails.
- Generated ledgers drift from maps.
- Prior retained artifacts are stale or not row-owned.
- Product rows rely only on frontend mocks.
- Public-route behavior is not proven through public routes.
- WebSocket behavior is tested outside `/ws/v1/` without owner permission.
- Conflict state relies on visible row index, column label, DOM order, or grid vendor coordinates instead of stable identifiers.
- Visual readiness is represented as product conformance.
- Accessibility smoke is used as row closure without map authority.
- Core 05 publication readiness is implied without claim-bearing metadata.

Current blocker register:

| Affected row | Affected source | Exact observed condition | Why closure is blocked | Minimum follow-up | Prevents |
| --- | --- | --- | --- | --- | --- |
| Phase `FE-P7` | `tools/frontend_phase_registry.json` | `status=planned`, `row_rollup_state=no_rows_implemented`, `FE-P7-ACTIVATION-BLOCKER-01` | Phase is not active and no direct row evidence has been promoted | Implement rows, update owner map inputs, run row targets, run drift/finalization checks | Phase closure |
| `FE-U-P7-01` | `tools/frontend_phase_maps/fe_p7_test_map.json` | `claim_status=blocked`, `frontend_phase_row_not_implemented` | Direct unit reducer evidence is absent | Add/promote row-owned `frontend-unit` evidence and row accounting | Row closure |
| `FE-U-P7-02` | `tools/frontend_phase_maps/fe_p7_test_map.json` | `claim_status=blocked`, `frontend_phase_row_not_implemented` | Direct conflict anchor/reducer evidence is absent | Add/promote row-owned `frontend-unit` evidence and row accounting | Row closure |
| `FE-I-P7-01` | `tools/frontend_phase_maps/fe_p7_test_map.json` | `claim_status=blocked`, `frontend_phase_row_not_implemented` | Resolver integration evidence is absent | Add/promote row-owned `frontend-unit` and `browser-e2e-webserver-backed` evidence | Row closure |
| `FE-E-P7-01` | `tools/frontend_phase_maps/fe_p7_test_map.json` | `claim_status=blocked`, `frontend_phase_row_not_implemented` | Multi-client public-route evidence is absent | Add/promote row-owned stateful and webserver-backed evidence through `/ws/v1/` and `/api/v1/` | Row closure and phase closure |
| `FE-V-P7-01` | `tools/frontend_phase_maps/fe_p7_test_map.json`; visual fixture registry | `claim_status=blocked`, `visual_fixture_not_recaptured_for_frontend_row`; registry has current FE-P7 fixture IDs but V-6/FE-P4 scenario names | Visual fixtures/goldens are context only without FE-P7 row accounting | Recapture or promote FE-P7 visual artifacts, run `browser-e2e-visual`, verify row accounting | Visual row closure |
| `FE-A11Y-P7-01` | `tools/frontend_phase_maps/fe_p7_test_map.json`; `apps/web/e2e/workbook.a11y-preflight.spec.ts` | `claim_status=blocked`; live target is `browser-e2e-a11y-preflight`; target explanation says phase coverage none | Blocked-row smoke cannot close implemented accessibility row unless map/guide allow it | Resolve map/guide target, run mapped accessibility target, capture row accounting and summary/preflight artifacts | Accessibility row closure |
| All rows | `.cartulary/test-results` | No direct FE-P7 row-owned evidence roots found | No current row accounting proves FE-P7 behavior | Run mapped targets after implementation and inspect `frontend-row-accounting.json` | Row closure |
| Claim-publication boundary | Core 05 | No explicit Core 05 claim-bearing metadata found | Publication evidence is inactive | Add Core 05 metadata only if claim publication is intentionally in scope | Publication claims only |

## Strict Non-Claims

This plan does not claim:

- FE-P7 completion by document creation.
- Row closure from generated ledgers.
- Row closure from prior phase plans.
- Row closure from previous retained artifacts unless current row-owned evidence permits it.
- Product conformance from visual or accessibility evidence.
- Claim-publication readiness.
- Saved-view persistence.
- Rollback readiness.
- Full coordination-surface readiness.
- FE-P8 readiness except as handoff context.

## Binary Exit Criteria

1. Initial plan creation: pass if `FRONTEND_PHASE7_IMPLEMENTATION_PLAN.md` exists at the repository root, covers all required sections and six live FE-P7 rows, records current blockers, does not edit generated artifacts, `make phase-ledger-drift` passes, and `git diff --check` passes. Fail otherwise.
2. Product-conformance row closure: pass only when `FE-U-P7-01`, `FE-U-P7-02`, `FE-I-P7-01`, and `FE-E-P7-01` are implemented/closure-eligible in the live map and current row-owned accounting from mapped targets proves the exact row behavior. Fail if any row remains blocked or evidence is mock-only where public routes are required.
3. Visual readiness: pass only when `FE-V-P7-01` has current `browser-e2e-visual` row accounting and required FE-P7 fixture artifacts. Fail if relying on fixture registry status, snapshot filenames, or previous V-6/FE-P4 runs alone.
4. Accessibility readiness: pass only when `FE-A11Y-P7-01` has map-authorized accessibility evidence and row accounting proving accessible names/states. Fail if blocked-row preflight smoke is used as closure without map authority.
5. Full FE-P7 phase completion: pass only when all FE-P7 rows are closed or owner-accepted blockers are recorded, FE-P7 registry status is updated accordingly, generated ledgers are drift-free, required targets pass, and finalization is recorded. Fail otherwise.
6. FE-P8 handoff readiness: pass only when FE-P8 receives final FE-P7 registry status, final row inventory, direct evidence roots, blockers for every unclosed row, drift/finalization outcomes, and strict non-claims. Fail if FE-P8 would need to infer evidence.

## FE-P8 Handoff

Current FE-P8 handoff baseline at plan creation:

- Final FE-P7 registry status: not final. Current status is `planned`, `no_rows_implemented`.
- Final FE-P7 row inventory: not final. Current inventory is the six rows listed above, all `blocked`.
- Direct evidence roots for closed FE-P7 rows: none found.
- Exact blockers for unclosed FE-P7 rows: see blocker register.
- WebSocket and public-route evidence status: existing Phase 6-era helpers and scenarios provide context, but no current FE-P7 row-owned evidence root exists.
- Conflict resolver evidence status: existing app code and Phase 6-era scenarios provide context, but no current FE-P7 row-owned evidence root exists.
- Presence anchoring status: existing selectors/helpers and V-6 scenarios provide context, but no FE-P7 closure.
- Visual readiness status and fixture IDs: live registry has `FE-VFIX-03`, `FE-VFIX-04`, and `FE-VFIX-08` as current FE-P7-relevant fixtures; row `FE-V-P7-01` remains blocked until FE-P7 row-owned visual accounting exists.
- Accessibility readiness status: live target is `browser-e2e-a11y-preflight`; `FE-A11Y-P7-01` remains blocked and preflight closure authority must be resolved before any claim.
- FE-P0 through FE-P7 dependency state: FE-P0 through FE-P6 are `active_green`; FE-P7 is planned/blocked.
- Drift and finalization command outcomes: `make phase-ledger-drift` passed for plan creation with `.cartulary/test-results/20260607T215728Z-p3866530`; full FE-P7 finalization has not run.
- Owner-accepted blockers FE-P8 must respect: none accepted as final; all current FE-P7 blockers prevent row closure.
- Strict non-claims FE-P8 cannot inherit without its own row-owned evidence: FE-P7 completion, saved-view persistence, rollback readiness, full coordination-surface readiness, product conformance from visual/a11y artifacts, and Core 05 publication readiness.
