# Frontend Phase 1 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for `FE-P1: App Shell And Session Bootstrap`.

`docs/guides/cartulary_frontend_implementation_testing_guide.md` is the controlling FE-P1 planning row set and frontend phase-planning source. Core 00 through Core 04 remain the implementation-conformance behavior authority. Core 05 applies only to claim-bearing timed or fixture-sensitive publication boundaries.

This guide does not implement FE-P1 behavior. It is not behavior authority and must not be used to mark any FE-P1 row complete without direct local validation evidence.

## Authority Model

- Adopted Cartulary NLSpecs apply only where explicitly adopted and applicable. `docs/testing-harness-nlspec.md` is used here for frontend harness mechanics only and does not define FE-P1 product behavior.
- Core 00 through Core 04 own implementation-conformance behavior.
- Core 05 owns claim-bearing timed or fixture-sensitive publication boundaries only.
- `docs/domain.md` is vocabulary and concept support for domain-facing terms such as incident, membership, session, and evidence; it is not a replacement for owner specs.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns the FE-P1 planning row set and frontend phase shape.
- `docs/guides/cartulary_implementation_testing_guide.md` informs shared harness sequencing, completion rules, and coverage-ledger shape.
- `docs/guides/cartulary-dev-guide.md` informs repo-local frontend package boundaries, generated-artifact policy, Make targets, workspace shape, and implementation baseline.
- `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md` are design-direction, visual, accessibility, and reviewer-discipline context only.
- Generated artifacts, generated ledgers, generated schedules, retained run artifacts, support-only tests, visual goldens, test names, previous summaries, and this guide are not behavior authorities.

## Current Repo Status

The following facts were verified by local inspection, non-mutating command execution during FE-P1 guide creation, and Sprint 1 readiness validation:

- `tools/frontend_phase_registry.json` exists and includes `FE-P1` in the `frontend` namespace with `status="planned"`, manifest path `tools/frontend_phase_maps/fe_p1_test_map.json`, ledger path `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`, and dependency on `FE-P0`.
- `tools/frontend_phase_maps/fe_p1_test_map.json` exists with schema `cartulary.frontend_phase_test_map.v1`, `phase_namespace="frontend"`, and `phase_id="FE-P1"`.
- The FE-P1 phase map contains exactly these row IDs once each: `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01`.
- All FE-P1 rows currently have `claim_status="blocked"` in the phase map and generated ledger.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md` exists and states that it is generated from `tools/frontend_phase_maps/fe_p1_test_map.json`; it must not be hand-edited for FE-P1 closure.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1` passed during Sprint 1 readiness validation and reported FE-P1 as planned, explainable, and non-executable.
- Sprint 1 registry and map invariant checks passed with `jq -e`; the registry tuple is exact and the FE-P1 map contains no missing or duplicate row IDs.
- `make phase-ledger-drift` passed during Sprint 1 readiness validation with final summary `.cartulary/test-results/20260529T211122Z-p2095159/phase-ledger-drift/tool-run-summary.json`.
- FE-P0 handoff regressions were rerun during Sprint 1 readiness validation: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make phase-ledger-drift` all passed.
- Target dry-run discovery found these candidate targets available: `make explain-phase`, `make frontend-typecheck`, `make frontend-unit`, `make browser-e2e-webserver-backed`, `make browser-e2e-a11y`, `make browser-e2e-support`, `make phase-ledgers`, `make phase-ledger-drift`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make check`.
- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` exists and records FE-P0 handoff state. FE-P1 may consume that handoff, but FE-P0 closure claims are not FE-P1 completion evidence.
- `/apps/web` contains existing Phase 1 app-shell surfaces by inspection: `AppRoot.tsx`, `App.tsx`, `Phase1Surface.tsx`, `Phase1Harness.tsx`, `phase1Client.ts`, `browserApi.ts`, `App.phase1.test.tsx`, `App.phase1.support.test.tsx`, `e2e/phase1.spec.ts`, `e2e/phase1.clock.spec.ts`, and `e2e/phase1Page.ts`.
- Existing source inspection found `/api/v1/` requests for auth session, credential state, login, logout, MFA TOTP begin/complete, password change, deployment-user admin, session revoke-all, and incidents. This is implementation-shape inspection only, not FE-P1 row completion evidence.
- `/packages/protocol-ts`, `/packages/ui-contracts`, and `/packages/test-utils` exist. `@cartulary/ui-contracts` already contains some app-shell and incident selector builders, while `apps/web/src/selectorContractPolicy.test.ts` still treats `auth-*`, `account-*`, `admin-*`, and some landing shell selectors as app-local until FE-P1 promotes them.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` lists FE-P1 scope as app bootstrap, server-managed session state, login/MFA/admin/incidents entry points, public error envelopes, and initial incident selection through `/api/v1/`.

Source limits:

- No newer root-level frontend phase-plan naming convention was found during inspection; this guide therefore uses the requested `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md` name, matching the existing root-level `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` pattern.
- FE-P1 row-owned closure evidence was not collected during Sprint 1. `make frontend-unit` was rerun only as FE-P0 handoff regression evidence; `make browser-e2e-webserver-backed`, `make browser-e2e-a11y`, and `make browser-e2e-support` remain planned FE-P1 validation, not cited completion evidence.
- `make check` was not run for guide creation because FE-P1 guide creation does not require broad developer-gate evidence and the requested task does not implement FE-P1 product behavior.
- Test names and existing Phase 1 files were inspected only to identify implementation surfaces; they do not prove FE-P1 completion.
- FE-P0 handoff state was inspected through the FE-P0 guide and FE-P0 map/ledger artifacts, and the Sprint 1 FE-P0 regression set passed. Later FE-P1 closure after product changes must rerun those checks again or precisely block them.

## Phase Objective

By FE-P1 exit, the frontend must have an app shell and session bootstrap baseline that uses public server routes and keeps frontend-visible state aligned with Core-owned session and authorization behavior:

- `/apps/web` starts from a real app bootstrap surface rather than a product-invisible harness path.
- Server-managed session state drives anonymous, loading, authenticated, MFA-required, forbidden, revoked, and error states.
- Login, MFA, account security, deployment-admin, incident landing, and initial incident selection enter through public `/api/v1/` routes.
- Frontend error rendering uses public error envelopes and does not depend on private server details.
- Route/API boundaries stay explicit and testable.
- Stable selector and test-id contracts exist for bootstrap and error-state surfaces that cross package, unit, browser, or support-test boundaries.
- Accessibility evidence covers authentication and entry states without promoting design-direction evidence to product-conformance evidence.
- FE-P0 package, generated-artifact, selector, and import-boundary guarantees remain preserved.

FE-P1 must not treat existing FE-P0 closure claims, generated ledgers, retained artifacts, support-only tests, accessibility checks, visual goldens, or this guide as FE-P1 product-conformance proof.

## Implementation Scope

### In scope

- `/apps/web` app bootstrap and initial route handling.
- Server-managed session state and revocation handling.
- Login, MFA enrollment/challenge, account security, deployment-admin, and incident entry points.
- Initial incident listing, creation, selection, stale selection, and authorization-observation behavior through `/api/v1/`.
- Public error-envelope parsing and frontend error-state rendering.
- Route/API boundary conformance, including `/api/v1/` confinement and closure behavior where route owners require it.
- Stable selector and test-id builders for bootstrap, landing, auth, account, admin, session, and error-state surfaces when those selectors cross test/runtime boundaries.
- Accessibility names, keyboard reachability, visible focus, and non-color-only state communication for FE-P1 states.
- P0 dependency preservation for generated protocol facades, selector stability, generated-artifact policy, generated drift, and import boundaries.

### Out of scope

- Workbook shell layout.
- Grid interaction.
- Mutation replay.
- Live collaboration.
- Visual fixture matrix.
- Claim publication.
- New Core behavior, new public wire shapes, or behavior-owner text changes unless FE-P1 implementation later discovers a true owner-spec gap.
- Hand edits to generated artifacts, generated ledgers, `pnpm-lock.yaml`, or `go.sum`.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| ---- | ------ | ------------------ | -------- | --------------- |
| [x] | 1. Frontend phase manifest, FE-P1 map, ledger, and P0 handoff validation | `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`; `make phase-ledger-drift`; FE-P0 handoff regression set | None for Sprint 1 metadata readiness | Planned frontend phases are explainable but not executable; no FE-P1 row is complete |
| [ ] | 2. App bootstrap state model and public error-envelope rendering; owns `FE-U-P1-01` | `make frontend-unit` | Unit evidence not run during guide creation | Keep public error envelopes separate from private server diagnostics |
| [ ] | 3. API client and route-boundary integration baseline; owns `FE-I-P1-01` | `make frontend-unit`; supporting `make frontend-typecheck` | Integration/unit evidence not run during guide creation | Public-route behavior stays under `/api/v1/` |
| [ ] | 4. Login, session bootstrap, incident entry, authorization, and revocation E2E flow; owns `FE-E-P1-01` | `make browser-e2e-webserver-backed` | Browser E2E evidence not run during guide creation | Browser-observed authorization is product-conformance only when Core-owned |
| [ ] | 5. Accessibility coverage for session, MFA, incident, forbidden, loading, and error states; owns `FE-A11Y-P1-01` | `make browser-e2e-a11y` | Accessibility evidence not run during guide creation | Design-direction evidence must not be counted as product conformance |
| [ ] | 6. Stable bootstrap and error-state selector builders; owns `FE-S-P1-01` | `make frontend-unit`; `make browser-e2e-support` | Support evidence not run during guide creation | Promote cross-boundary selectors to shared builders where needed |
| [ ] | 7. Closure pass, ledger drift, P0 regression check, and FE-P2 handoff | Row-owned commands plus `make phase-ledger-drift`; whitespace checks | Broad `make check` not run unless closure rules require it | Record exact artifact paths and unresolved owner-lookup TODOs |

## Global References

- Controlling FE-P1 guide row set: `docs/guides/cartulary_frontend_implementation_testing_guide.md`, Phase `FE-P1`.
- FE-P1 rows: `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01`.
- Core owner documents: `docs/spec/00_document_set_status_and_precedence.md`, `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/02_domain_model_schema_and_history.md`, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, `docs/spec/04_security_deployment_and_conformance.md`, and `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`.
- Harness and planning references: `docs/testing-harness-nlspec.md`, `docs/guides/cartulary_implementation_testing_guide.md`, and `docs/guides/cartulary-dev-guide.md`.
- Vocabulary and design context: `docs/domain.md`, `docs/guides/cartulary-ui-ux-design-guide.md`, `docs/design.md`, and `docs/guides/cartulary_visual_golden_maintenance.md`.
- FE-P1 package and app surfaces: `/apps/web`, `/packages/protocol-ts`, `/packages/ui-contracts`, and `/packages/test-utils`.
- FE-P1 phase artifacts: `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p1_test_map.json`, and `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`.
- Command surface: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`, `make frontend-unit`, `make frontend-typecheck`, `make browser-e2e-webserver-backed`, `make browser-e2e-a11y`, `make browser-e2e-support`, `make phase-ledgers`, `make phase-ledger-drift`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, `make check`, `git diff --check`, and `git diff --cached --check` when staged changes exist.

## Evidence Layer Matrix

| Row | Evidence layer | Evidence class | Claim intent |
| --- | -------------- | -------------- | ------------ |
| `FE-U-P1-01` | Unit | Product conformance | App bootstrap state distinguishes unauthenticated, MFA-required, authenticated, forbidden, revoked, loading, and public error-envelope states |
| `FE-I-P1-01` | Integration | Product conformance | API client requests stay under `/api/v1/`, preserve server-managed session behavior, and avoid private server-detail dependencies in rendered errors |
| `FE-E-P1-01` | E2E | Product conformance | Browser-observed login, session bootstrap, incident entry, and current-role/current-membership authorization effects work through public routes |
| `FE-A11Y-P1-01` | Accessibility | Design direction | Session, MFA, incident, forbidden, loading, and error states are keyboard-reachable, visibly focused, and screen-reader safe |
| `FE-S-P1-01` | Support | Implementation support | Bootstrap route selectors and error-state selectors use stable test-id builders |

## Dependencies and Prerequisites

- FE-P0 handoff must remain valid: generated protocol facades, view-contract adapters, stable selector/test-id foundations, generated-artifact policy, generated drift, frontend import boundaries, frontend phase maps, and generated frontend ledgers remain prerequisites.
- FE-P1 uses the frontend namespace. Use `PHASE_NAMESPACE=frontend PHASE=FE-P1`; do not append FE-P1 rows to base `PHASE=phaseN` maps.
- Public route and error-envelope behavior is owned by Core specs and derived contracts. FE-P1 implementation must not invent private frontend-only route semantics.
- Server-managed session state means the browser observes and reacts to server session state; the frontend must not create an independent session authority.
- Accessibility rows and support rows gate implementation readiness but must stay visibly separate from product-conformance evidence.
- Generated frontend ledgers are downstream renderings of frontend maps. Update maps first and regenerate ledgers through the generator when row metadata changes.

## Public Interfaces and Deliverables

FE-P1 should not introduce new public API shapes through this guide. Later FE-P1 implementation may consume existing public `/api/v1/` routes and generated contracts, but behavior changes must follow owner specs first.

Expected deliverables for FE-P1 implementation are:

- A maintained `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md` guide.
- FE-P1 row evidence recorded against `tools/frontend_phase_maps/fe_p1_test_map.json` and generated companion ledger updates produced by the ledger generator.
- Unit evidence for app bootstrap states and public error envelopes.
- Unit/integration evidence for `/api/v1/` client confinement, server-managed session behavior, CSRF/cookie behavior, and public error rendering.
- Browser evidence for login, MFA, session bootstrap, incident entry, authorization observation, and session revocation.
- Accessibility evidence for FE-P1 state surfaces.
- Shared selector/test-id builders where FE-P1 selectors cross stable boundaries.
- P0 regression evidence or precise blockers.

## Sprint 1: Frontend Phase Manifest, FE-P1 Map, Ledger, And P0 Handoff Validation

Status: Complete for Sprint 1 readiness and traceability. This sprint does not complete any FE-P1 row.

### 1. Objective And Non-Goals

- Objective: prove `FE-P1` is explainable in the `frontend` namespace, has the expected row inventory, has a generated ledger consistent with its map, and is gated by current FE-P0 handoff validation before any FE-P1 completion claim.
- Non-goals: no FE-P1 product behavior, no row completion, no hand edits to generated ledgers, no Core behavior changes, and no Core 05 claim-publication activation.
- Relevant IDs: supports `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01`.

### 2. Source And Authority Order

1. Core 00 through Core 04 own implementation-conformance behavior.
2. Core 05 applies only to claim-bearing timed or fixture-sensitive publication.
3. `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns FE-P1 row planning and frontend phase shape.
4. `docs/testing-harness-nlspec.md` owns harness mechanics only.
5. `docs/guides/cartulary_implementation_testing_guide.md` informs shared completion and ledger rules.
6. `docs/guides/cartulary-dev-guide.md` informs repo/package/Make/generated-artifact boundaries.
7. `docs/domain.md` is vocabulary support only.
8. Generated ledgers, retained artifacts, test names, previous summaries, FE-P0 closure claims, and this plan are not FE-P1 product-conformance proof.

### 3. File-By-File Inspection Checklist

- `tools/frontend_phase_registry.json`: schema `cartulary.frontend_phase_registry.v1`, namespace `frontend`, exactly one `FE-P1` entry, `status="planned"`, expected map and ledger paths, owner ref to the frontend guide, and `depends_on=["FE-P0"]` were verified.
- `tools/frontend_phase_maps/fe_p1_test_map.json`: schema `cartulary.frontend_phase_test_map.v1`, namespace `frontend`, `phase_id="FE-P1"`, and the five expected row IDs exactly once were verified.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`: states it is generated from `tools/frontend_phase_maps/fe_p1_test_map.json` and mirrors FE-P1 status, dependency, evidence classes, rows, targets, claims, and out-of-scope text.
- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`: FE-P0 handoff requirements and historical evidence were inspected; historical FE-P0 closure claims remain prerequisite context only.
- `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md`: this Sprint 1 section records readiness procedure, command results, source limits, blocker language, and binary criteria.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md`: FE-P1 scope, five-row inventory, evidence-class rules, planned/non-executable phase rule, and completion checklist were inspected.
- `scripts/lib/frontend-phase-manifest.mjs`, `scripts/render-phase-ledger.mjs`, `scripts/check-phase-ledger-drift.mjs`, and `scripts/print-explain-phase.mjs`: schema validation, ledger rendering, drift comparison, and planned frontend explain behavior were inspected.
- `Makefile`, `tools/task_surface_manifest.json`, and `scripts/lib/make-node-tools.mjs`: `make explain-phase`, `make phase-ledger-drift`, `make phase-ledgers`, and accepted `PHASE_NAMESPACE` wiring were inspected.

### 4. Registry Invariants

- `FE-P1` resolves through `PHASE_NAMESPACE=frontend PHASE=FE-P1`; FE rows were not added to base `tools/phase_registry.json`.
- Registry entry points to `tools/frontend_phase_maps/fe_p1_test_map.json` and `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`.
- `depends_on` remains exactly `["FE-P0"]`.
- `status` remains `planned`; planned frontend phases are explainable but non-executable.

### 5. FE-P1 Map Invariants

- Row inventory is exactly `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01`; no missing or duplicate IDs.
- Product-conformance rows are exactly `FE-U-P1-01`, `FE-I-P1-01`, and `FE-E-P1-01`, each with non-empty Core REQ and Core AC IDs.
- `FE-A11Y-P1-01` remains `design_direction`, has support/design ACs, and has no Core REQ/AC claim.
- `FE-S-P1-01` remains `implementation_support`, has support ACs, and has no Core REQ/AC claim.
- `claim_publication_boundary` remains a distinct supported evidence class, but FE-P1 has no claim-publication row; Sprint 1 activates no Core 05 claim-publication review.
- All five FE-P1 rows remain `claim_status="blocked"` because no direct row-owned FE-P1 evidence was promoted.

### 6. Generated-Ledger Invariants

- The FE-P1 ledger is generated from `tools/frontend_phase_maps/fe_p1_test_map.json`; it was inspected but not hand-edited.
- Ledger mirrors namespace `frontend`, status `planned`, dependency `FE-P0`, and all five row records.
- `make phase-ledger-drift` passed, proving generated ledger consistency with the map.
- `make phase-ledgers` was not run because Sprint 1 made no registry or map metadata changes.

### 7. FE-P0 Handoff Validation Procedure

- Rerun before any FE-P1 closure claim: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make phase-ledger-drift`.
- Sprint 1 reran that FE-P0 handoff regression set and all commands passed.
- `make phase-slice PHASE=phase7` and `make service-backed-slice PHASE=phase7` were not rerun because Sprint 1 touched only this readiness plan and did not change shared selector, row-history, `history_item_ref`, or view-schema identity surfaces.
- `make check` was not run because Sprint 1 is metadata/readiness-only; it remains a closure or broad-gate command when repo completion rules require it.
- If a FE-P0 handoff check cannot run later, record exactly: `BLOCKER: FE-P1 closure blocked because FE-P0 handoff regression <command> was not rerun: <reason>. Last FE-P0 evidence in FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md is historical only and is not current FE-P1 closure evidence. owner=P0-regression-owned minimum_follow_up=<next command/action>.`

### 8. Validation Commands And Outcomes

- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1`: passed; output reported namespace `frontend`, status `planned`, expected map and ledger paths, dependency `FE-P0`, all five rows, and planned/non-executable diagnostic.
- Registry invariant check with `jq -e`: passed.
- Map invariant check with `jq -e`: passed.
- `make phase-ledger-drift`: passed with final post-edit summary `.cartulary/test-results/20260529T211122Z-p2095159/phase-ledger-drift/tool-run-summary.json`.
- `make frontend-typecheck`: passed with `.cartulary/test-results/20260529T210821Z-p2088837/frontend-typecheck/tool-run-summary.json`.
- `make frontend-unit`: passed as FE-P0 handoff regression with `.cartulary/test-results/20260529T210821Z-p2088882/frontend-unit/tool-run-summary.json`.
- `make generated-artifact-policy-check`: passed with `.cartulary/test-results/20260529T210821Z-p2088832/generated-artifact-policy-check/tool-run-summary.json`.
- `make frontend-import-boundary-check`: passed with `.cartulary/test-results/20260529T210821Z-p2088856/frontend-import-boundary-check/tool-run-summary.json`.
- `make generate-drift`: passed with `.cartulary/test-results/20260529T210843Z-p2090950/generate-drift/tool-run-summary.json`.

### 9. Failure Modes And Blocker Language

- Missing or duplicate FE-P1 row: `BLOCKER: FE-P1 map row inventory invalid; expected exactly FE-U/I/E/A11Y/S-P1-01 once each; actual=<ids/counts>.`
- Wrong namespace or paths: `BLOCKER: FE-P1 registry is not frontend-namespace traceable; expected namespace/path/dependency tuple does not match registry.`
- Ledger drift: `BLOCKER: FE-P1 generated ledger is stale relative to map; rerun generator only after confirming map is the intended source.`
- Explain failure: `BLOCKER: FE-P1 is not explainable in frontend namespace; command=<command> output=<diagnostic>.`
- P0 rerun skipped or failed: use the FE-P0 blocker template in this sprint section.
- Evidence-class collapse: `BLOCKER: FE-P1 evidence classes collapsed; design/support/claim-publication-boundary evidence cannot be counted as product_conformance.`

### 10. Deliverables

- FE-P1 metadata inventory is recorded in this section: registry tuple, row inventory, evidence classes, row statuses, target mappings, and generated-ledger relationship.
- FE-P0 handoff validation record is recorded with current passing command artifact paths.
- Source limits and explicit non-claims are recorded so later agents do not overclaim Sprint 1 evidence.
- No FE-P1 row is marked complete from metadata inspection alone.

### 11. Binary Acceptance Criteria

| Criterion | Sprint 1 status |
| --------- | --------------- |
| `FE-P1` is explainable with `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1` | Satisfied |
| Registry tuple is exact: namespace `frontend`, expected map path, expected ledger path, dependency `FE-P0`, status `planned` | Satisfied |
| FE-P1 row inventory has no missing or duplicate row IDs | Satisfied |
| Evidence classes remain separate and are not collapsed into product conformance | Satisfied |
| FE-P1 ledger is generated-consistent | Satisfied by `make phase-ledger-drift` |
| FE-P0 handoff regressions are rerun or precisely blocked | Satisfied; all required commands passed |
| Source limits are recorded | Satisfied |
| No FE-P1 row is marked complete from metadata inspection, generated ledger text, retained artifacts, support-only checks, accessibility checks, visual goldens, test names, FE-P0 closure claims, or this plan | Satisfied |

### 12. Explicit Non-Claims

- Sprint 1 does not prove app bootstrap, session, MFA, authorization, incident entry, public error-envelope, accessibility, or selector-builder behavior.
- Sprint 1 does not activate Core 05 publication review.
- Sprint 1 `make frontend-unit` evidence is FE-P0 handoff regression evidence only; it does not close `FE-U-P1-01`, `FE-I-P1-01`, or `FE-S-P1-01`.
- Generated ledger consistency proves map/ledger agreement, not FE-P1 product conformance.

## Sprint 2: App Bootstrap State Model And Public Error-Envelope Rendering

- Objective: Implement and verify the app bootstrap state model for loading, anonymous, MFA-required, authenticated, forbidden, revoked, and public error-envelope states.
- Status: Planned.
- Relevant IDs: Owns `FE-U-P1-01`.
- Files and areas to inspect or edit: `apps/web/src/App.tsx`, `apps/web/src/AppRoot.tsx`, `apps/web/src/Phase1Surface.tsx`, `apps/web/src/browserApi.ts`, `apps/web/src/appShellTestSupport.ts`, `apps/web/src/App.phase1.test.tsx`, and `apps/web/src/App.phase1.support.test.tsx`.
- Test-first sequence: Add or tighten unit tests for each bootstrap state, including public error envelopes with allowed public fields only; verify rendered text and state transitions before changing implementation.
- Implementation tasks: Keep bootstrap state explicit; separate anonymous from loading; surface MFA setup/challenge states; treat `session_required` and session revocation as server state; render public `error.code`, public status text, and safe public details only; avoid private diagnostic dependencies.
- Validation commands: `make frontend-unit`; supporting `make frontend-typecheck` when types change.
- Deliverables: Unit coverage for FE-P1 bootstrap states and public error rendering; any app-shell state helpers needed to keep transitions deterministic.
- Risks and assumptions: Existing Phase 1 unit test names are not sufficient evidence unless the row-owned target is run and passes.
- Blockers and follow-up notes: No FE-U-P1-01 completion claim is allowed until `make frontend-unit` passes with direct evidence or a precise blocker is recorded.

## Sprint 3: API Client And Route-Boundary Integration Baseline

- Objective: Ensure the frontend API client and route-boundary tests preserve public `/api/v1/` routing, server-managed sessions, and public error-envelope rendering.
- Status: Planned.
- Relevant IDs: Owns `FE-I-P1-01`.
- Files and areas to inspect or edit: `apps/web/src/phase1Client.ts`, `apps/web/src/browserApi.ts`, `apps/web/src/fetchMockTestSupport.ts`, `apps/web/src/App.phase1.test.tsx`, `/packages/protocol-ts`, `/packages/ui-contracts`, and route-boundary fixtures.
- Test-first sequence: Add or tighten integration-style unit tests for every FE-P1 client route family before implementation changes; assert method, path, credentials, CSRF behavior, body shape, and error rendering.
- Implementation tasks: Keep all app client requests under `/api/v1/`; preserve default `credentials="include"` for server-managed sessions; use bootstrap-token credentials exceptions only for bootstrap MFA routes; include CSRF headers only when cookies and request method require them; reject private error details in rendered UI; test unknown closed request members where route owners require closure.
- Validation commands: `make frontend-unit`; supporting `make frontend-typecheck`, `make generated-artifact-policy-check`, `make generate-drift`, and `make frontend-import-boundary-check` when implementation touches contracts, generated inputs, or package boundaries.
- Deliverables: API route-boundary tests and client behavior that can be traced to Core-owned public routes and public error envelopes.
- Risks and assumptions: FE-P1 must not infer new route closure policy from frontend convenience; closure expectations must follow route owners.
- Blockers and follow-up notes: No FE-I-P1-01 completion claim is allowed until direct `make frontend-unit` evidence exists or a precise blocker records why it cannot run.

## Sprint 4: Login, Session Bootstrap, Incident Entry, Authorization, And Revocation E2E Flow

- Objective: Verify the browser-visible FE-P1 flow through login, MFA, server session bootstrap, incident entry, deployment-admin controls, current-role/current-membership authorization effects, and revocation handling.
- Status: Planned.
- Relevant IDs: Owns `FE-E-P1-01`.
- Files and areas to inspect or edit: `apps/web/e2e/phase1.spec.ts`, `apps/web/e2e/phase1.clock.spec.ts`, `apps/web/e2e/phase1Page.ts`, `apps/web/e2e/authRuntime.ts`, `apps/web/e2e/sessionSupport.ts`, `apps/web/e2e/helpers.ts`, `apps/web/src/App.tsx`, and `apps/web/src/Phase1Surface.tsx`.
- Test-first sequence: Add or tighten Playwright scenarios for local login, invalid credentials, MFA required/setup, bootstrap TOTP, password-change revocation, deployment-admin actions, incident-admin denial, incident create/list/open, stale incident selection, forbidden incident observation, and re-auth after revocation.
- Implementation tasks: Preserve public-route browser flows; ensure incident entry starts from public incident list/create routes; return to anonymous/login state on session loss; surface authorization failures through public error envelopes; keep workbook-shell layout and mutation replay out of FE-P1 scope.
- Validation commands: `make browser-e2e-webserver-backed`; supporting `make frontend-unit` when browser failures indicate app-state defects.
- Deliverables: Browser E2E evidence for FE-P1 entry and session flows with retained run artifacts.
- Risks and assumptions: Browser E2E is service-backed and may be expensive; skipped or failed runs must be recorded as exact blockers rather than inferred from unit evidence.
- Blockers and follow-up notes: No FE-E-P1-01 completion claim is allowed until direct browser E2E evidence exists or a precise blocker is recorded.

## Sprint 5: Accessibility Coverage For Session, MFA, Incident, Forbidden, Loading, And Error States

- Objective: Verify FE-P1 state surfaces are keyboard reachable, visibly focused, and communicated with screen-reader-safe names and non-color-only state cues.
- Status: Planned.
- Relevant IDs: Owns `FE-A11Y-P1-01`.
- Files and areas to inspect or edit: `apps/web/e2e/workbook.a11y.spec.ts`, FE-P1 app surfaces under `apps/web/src`, accessibility summary tooling, `docs/guides/cartulary-ui-ux-design-guide.md`, and `docs/design.md`.
- Test-first sequence: Add or tighten accessibility scenarios for session loading, anonymous login, MFA required/setup, authenticated landing, incident empty/list/error, forbidden/access-denied, revoked/session-required, and public error states.
- Implementation tasks: Ensure form controls have accessible names; keep focus visible; make status and error changes programmatically discoverable where needed; avoid color-only state; preserve keyboard reachability for login, MFA, incident, account, and admin entry controls.
- Validation commands: `make browser-e2e-a11y`.
- Deliverables: Accessibility run evidence and any required app-surface accessibility fixes.
- Risks and assumptions: FE-A11Y-P1-01 is `design_direction`, not product conformance. Its evidence gates readiness but must not be represented as Core-owned product behavior.
- Blockers and follow-up notes: No FE-A11Y-P1-01 completion claim is allowed until direct accessibility evidence exists or a precise blocker is recorded.

## Sprint 6: Stable Bootstrap And Error-State Selector Builders

- Objective: Promote FE-P1 bootstrap, route, and error-state selectors to stable builders where they cross runtime, unit, browser, or support-test boundaries.
- Status: Planned.
- Relevant IDs: Owns `FE-S-P1-01`.
- Files and areas to inspect or edit: `packages/ui-contracts/src/index.ts`, `packages/ui-contracts/src/index.test.ts`, `packages/test-utils/src/index.ts`, `apps/web/src/selectorContractPolicy.test.ts`, `apps/web/src/Phase1Surface.tsx`, `apps/web/src/App.tsx`, `apps/web/src/App.phase1.test.tsx`, and `apps/web/e2e/phase1Page.ts`.
- Test-first sequence: Add failing shared-builder tests for FE-P1 auth, account, admin, landing, bootstrap, session, and error selectors before replacing app-local literals; add policy coverage that rejects new raw cross-boundary FE-P1 selectors.
- Implementation tasks: Add selector builders for FE-P1 surfaces where needed; migrate app, unit, E2E, and support helpers to consume builders; keep stable IDs derived from stable route/session/incident/user identifiers rather than visible labels or render order; leave purely local test fixture literals app-local only when ownership metadata is explicit.
- Validation commands: `make frontend-unit`; `make browser-e2e-support`; supporting `make frontend-typecheck`.
- Deliverables: Shared selector/test-id builders, migrated FE-P1 selector consumers, and policy coverage that keeps FE-P1 selector ownership explicit.
- Risks and assumptions: Selector/test-id strings are internal test contracts, not public API compatibility commitments.
- Blockers and follow-up notes: No FE-S-P1-01 completion claim is allowed until direct support evidence exists or a precise blocker is recorded.

## Sprint 7: Closure Pass, Ledger Drift, P0 Regression Check, And FE-P2 Handoff

- Objective: Collect FE-P1 evidence, refresh generated ledgers through the generator when metadata changes, verify P0 regression status, and prepare FE-P2 handoff without turning support/design evidence into product-conformance evidence.
- Status: Planned.
- Relevant IDs: Supports all FE-P1 rows.
- Files and areas to inspect or edit: FE-P1 plan, frontend phase registry/map/ledger, validation artifacts, FE-P0 artifacts, app-shell files, selector packages, and FE-P2 handoff notes.
- Test-first sequence: Rerun row-owned validation commands; run `make phase-ledger-drift`; run P0 regression checks relevant to changed surfaces; run whitespace checks.
- Implementation tasks: Record exact row evidence or blockers; regenerate ledgers through `make phase-ledgers` only if phase-map metadata changes; keep `FE-A11Y-P1-01` as design-direction evidence and `FE-S-P1-01` as implementation-support evidence; avoid Core 05 claim-publication activation.
- Validation commands: `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-a11y`; `make browser-e2e-support`; `make phase-ledger-drift`; `make generated-artifact-policy-check`; `make generate-drift`; `make frontend-import-boundary-check`; `git diff --check`; `git diff --cached --check` when staged; `make check` only when repository completion rules require broad closure.
- Deliverables: FE-P1 closure evidence, precise blockers, generated-ledger consistency, P0 regression status, and FE-P2 handoff state.
- Risks and assumptions: Broad `make check` is expensive and should be reserved for closure or repository rules, not guide creation.
- Blockers and follow-up notes: FE-P1 remains incomplete until every row has direct evidence or an explicit blocker and P0 regression status is recorded.

## Blocker Recording Rules

For every failed validation command or missing prerequisite, record:

- Exact command or missing artifact.
- Exact failing target, scheduler unit, or test when available.
- Result root, run ID, run root, summary JSON, stdout log, and stderr log paths when available.
- FE-P1 row ID when applicable.
- Failure class and failure reason when exposed.
- Ownership: FE-P1-owned, P0-regression-owned, support-tooling-owned, harness-owned, infra-owned, or outside FE-P1.
- Minimum follow-up needed.
- Whether the blocker prevents FE-P1 completion or is support/design-only.
- Whether the blocker affects FE-P2 handoff.

Do not replace blockers with guesses. If a command is not run, state that it was not run and do not cite it as evidence.

## Binary Exit Criteria

| Criterion | Current status |
| --------- | -------------- |
| The guide file exists and states that it is not behavior authority | Satisfied by this file once committed or otherwise accepted |
| FE-P1 phase registry, phase map, and generated ledger are present or a precise blocker is recorded | Satisfied by Sprint 1 readiness inspection: registry, map, and ledger exist |
| Every FE-P1 row is represented exactly once in the phase map | Satisfied by Sprint 1 duplicate-row inspection: all five requested rows appear once |
| Generated ledgers are produced by generator, not hand-edited | Satisfied for Sprint 1 metadata: ledger states it is generated from the map and `make phase-ledger-drift` passed |
| `FE-U-P1-01` has direct unit evidence or a blocker | Blocked for FE-P1 closure: no row-owned FE-P1 unit evidence was promoted; `make frontend-unit` passed only as FE-P0 handoff regression |
| `FE-I-P1-01` has direct integration/unit evidence or a blocker | Blocked for FE-P1 closure: no row-owned FE-P1 integration evidence was promoted; `make frontend-unit` passed only as FE-P0 handoff regression |
| `FE-E-P1-01` has browser E2E evidence or a blocker | Blocked: `make browser-e2e-webserver-backed` was not run for guide creation |
| `FE-A11Y-P1-01` has accessibility evidence or a blocker | Blocked: `make browser-e2e-a11y` was not run for guide creation |
| `FE-S-P1-01` has selector-builder support evidence or a blocker | Blocked for FE-P1 closure: no row-owned FE-P1 selector-builder evidence was promoted; `make frontend-unit` passed only as FE-P0 handoff regression and `make browser-e2e-support` was not run |
| Public-route behavior stays under `/api/v1/` | Required for closure; source inspection found existing `/api/v1/` usage but row-owned evidence is still blocked |
| Frontend error rendering uses public error envelopes and not private server details | Required for closure; direct row-owned evidence is still blocked |
| Server-managed session state is respected | Required for closure; direct row-owned evidence is still blocked |
| Unknown closed request members are tested where route owners require closure | Required for closure; exact route-owner coverage must be verified during implementation |
| P0 regression checks remain green or exact blockers are recorded | Satisfied for Sprint 1 readiness: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make phase-ledger-drift` passed |
| Support-only and design-direction evidence are not represented as product-conformance evidence | Satisfied for Sprint 1 metadata; current FE-P1 map separates product conformance, design direction, and implementation support |
| No Core 05 claim-publication review is activated unless explicitly requested | Satisfied for Sprint 1; this guide does not activate Core 05 publication review |
| Whitespace checks pass | Satisfied for Sprint 1: `git diff --check` and `git diff --cached --check` passed after the plan edit |

FE-P1 must not be marked complete until every criterion is satisfied with direct evidence or an explicit blocker.

## Handoff Requirements For FE-P2

FE-P2 must receive:

- App bootstrap state status, including loading, anonymous, MFA-required, authenticated, forbidden, revoked, and error-state handling.
- Public API client and session status, including `/api/v1/` route confinement, server-managed cookie/CSRF behavior, bootstrap-token exceptions, and public error-envelope rendering.
- Login, MFA, and session-revocation status, including direct browser evidence or exact blockers.
- Incident entry and authorization-observation status, including visible incident list/create/open behavior, stale incident handling, forbidden/access-denied rendering, and current-role/current-membership effects.
- Public error-envelope rendering status, including any known private-detail leakage blockers.
- Accessibility status for authentication, MFA, incident, forbidden, loading, and error states.
- Selector/test-id builder status for bootstrap, route, landing, auth, account, admin, session, and error surfaces.
- P0 regression status for generated protocol facades, view-contract adapters, stable selector foundations, generated-artifact policy, generated drift, and frontend import boundaries.
- Command-surface blockers and validation artifact locations for every row-owned command and support command.
- Generated ledger/map drift status and any generated-artifact policy blockers.
- Unresolved owner-lookup TODOs, especially for route closure, public error detail boundaries, and any selector ownership that remains app-local.

Support-only status and design-direction status must stay visibly separate from product-conformance status in the FE-P2 handoff.
