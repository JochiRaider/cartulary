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

The following facts were verified by local inspection, non-mutating command execution during FE-P1 guide creation, Sprint 1 readiness validation, Sprint 2 validation, Sprint 3 remediation, Sprint 3 audit follow-up, Sprint 4 browser E2E execution, Sprint 5 accessibility execution, and Sprint 5 accessibility audit refresh:

- `tools/frontend_phase_registry.json` exists and includes `FE-P1` in the `frontend` namespace with `status="planned"`, manifest path `tools/frontend_phase_maps/fe_p1_test_map.json`, ledger path `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`, and dependency on `FE-P0`.
- `tools/frontend_phase_maps/fe_p1_test_map.json` exists with schema `cartulary.frontend_phase_test_map.v1`, `phase_namespace="frontend"`, and `phase_id="FE-P1"`.
- The FE-P1 phase map contains exactly these row IDs once each: `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01`.
- After Sprint 5 execution, `FE-U-P1-01`, `FE-I-P1-01`, `FE-E-P1-01`, and `FE-A11Y-P1-01` have `claim_status="implemented"` in the phase map and generated ledger. `FE-S-P1-01` remains `claim_status="blocked"`, and the FE-P1 registry `status` remains `planned`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md` exists and states that it is generated from `tools/frontend_phase_maps/fe_p1_test_map.json`; it must not be hand-edited for FE-P1 closure.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1` passed during Sprint 1 readiness validation and reported FE-P1 as planned, explainable, and non-executable.
- Sprint 1 registry and map invariant checks passed with `jq -e`; the registry tuple is exact and the FE-P1 map contains no missing or duplicate row IDs.
- `make phase-ledger-drift` passed during Sprint 2 validation with final summary `.cartulary/test-results/20260529T222417Z-p2204673/phase-ledger-drift/tool-run-summary.json`.
- FE-P0 handoff regressions were rerun during Sprint 1 readiness validation: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make phase-ledger-drift` all passed.
- Target dry-run discovery found these candidate targets available: `make explain-phase`, `make frontend-typecheck`, `make frontend-unit`, `make browser-e2e-webserver-backed`, `make browser-e2e-a11y`, `make browser-e2e-support`, `make phase-ledgers`, `make phase-ledger-drift`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make check`.
- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` exists and records FE-P0 handoff state. FE-P1 may consume that handoff, but FE-P0 closure claims are not FE-P1 completion evidence.
- `/apps/web` contains existing Phase 1 app-shell surfaces by inspection: `AppRoot.tsx`, `App.tsx`, `Phase1Surface.tsx`, `Phase1Harness.tsx`, `phase1Client.ts`, `browserApi.ts`, `App.phase1.test.tsx`, `App.phase1.support.test.tsx`, `e2e/phase1.spec.ts`, `e2e/phase1.clock.spec.ts`, and `e2e/phase1Page.ts`.
- Existing source inspection found `/api/v1/` requests for auth session, credential state, login, logout, MFA TOTP begin/complete, password change, deployment-user admin, session revoke-all, and incidents. This is implementation-shape inspection only, not FE-P1 row completion evidence.
- `/packages/protocol-ts`, `/packages/ui-contracts`, and `/packages/test-utils` exist. `@cartulary/ui-contracts` already contains some app-shell and incident selector builders, while `apps/web/src/selectorContractPolicy.test.ts` still treats `auth-*`, `account-*`, `admin-*`, and some landing shell selectors as app-local until FE-P1 promotes them.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` lists FE-P1 scope as app bootstrap, server-managed session state, login/MFA/admin/incidents entry points, public error envelopes, and initial incident selection through `/api/v1/`.

Source limits:

- No newer root-level frontend phase-plan naming convention was found during inspection; this guide therefore uses the requested `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md` name, matching the existing root-level `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md` pattern.
- FE-P1 row-owned closure evidence was not collected during Sprint 1. Sprint 4 later collected direct product-conformance evidence for `FE-E-P1-01` from `make browser-e2e-webserver-backed`; Sprint 5 later collected and audit-refreshed direct design-direction evidence for `FE-A11Y-P1-01` from `make browser-e2e-a11y`. `make browser-e2e-support` remains planned FE-P1 validation, not cited completion evidence.
- `make check` was not run for Sprint 4 because the row-owned product-conformance target passed and broad developer-gate evidence remains Sprint 7 closure scope unless repository completion rules require it earlier.
- Initial test-name and existing Phase 1 file inspection identified implementation surfaces only. Sprint 4 completion evidence is the later `make browser-e2e-webserver-backed` run, not inspection.
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
| [x] | 2. App bootstrap state model and public error-envelope rendering; owns `FE-U-P1-01` | `make frontend-unit`; `make frontend-typecheck`; `make phase-ledgers`; `make phase-ledger-drift` | None for Sprint 2 unit scope | `FE-U-P1-01` promoted only; FE-I/E/A11Y/S rows remain blocked |
| [x] | 3. API client and route-boundary integration baseline; owns `FE-I-P1-01` | `make frontend-unit`; supporting `make frontend-typecheck` | None for Sprint 3 row-owned route-boundary scope | Public-route behavior stays under `/api/v1/`; FE-I owns client serialization and rendering evidence, not backend rejection conformance |
| [x] | 4. Login, session bootstrap, incident entry, authorization, and revocation E2E flow; owns `FE-E-P1-01` | `make browser-e2e-webserver-backed` | None for Sprint 4 browser E2E scope | `FE-E-P1-01` promoted from direct browser-observed product-conformance evidence through public `/api/v1/` routes |
| [x] | 5. Accessibility coverage for session, MFA, incident, forbidden, loading, and error states; owns `FE-A11Y-P1-01` | `make browser-e2e-a11y` | None for Sprint 5 accessibility scope | `FE-A11Y-P1-01` promoted from direct design-direction accessibility evidence only; no Core product-conformance or Core 05 evidence claimed |
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

Status: Complete for `FE-U-P1-01` unit implementation and row-owned unit evidence. FE-P1 overall remains planned because `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01` remain blocked.
Relevant IDs: owns `FE-U-P1-01` only.

### 1. Objective and non-goals

- Objective: implement and verify the app bootstrap state model for loading, anonymous, MFA-required, authenticated, forbidden, revoked, and public error-envelope states.
- Objective: make row-owned Sprint 2 unit evidence precise enough that `FE-U-P1-01` can later be claimed only from direct `make frontend-unit` evidence or a precise blocker.
- Non-goals: no FE-I/E/A11Y/S row closure, no generated-ledger hand edits, no Core behavior changes, no Core 05 claim publication, no workbook-shell layout work, no grid interaction, no mutation replay, and no live-collaboration work.
- This Sprint 2 plan is an execution roadmap only. It is not behavior authority, product evidence, or completion evidence.

### 2. Source and authority constraints

- Core 00 through Core 04 own implementation-conformance behavior.
- Core 05 applies only to claim-bearing timed or fixture-sensitive publication boundaries.
- `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns FE-P1 row planning and maps `FE-U-P1-01` to `make frontend-unit`.
- `docs/testing-harness-nlspec.md` owns harness mechanics only and does not define Sprint 2 product behavior.
- Preserve a server-managed session model: browser state follows `/api/v1/auth/session` and public auth routes; the frontend must not parse, mint, or treat any identity/session token as client authority.
- No `FE-U-P1-01` completion claim is allowed from metadata inspection, FE-P0 historical evidence, generated ledger consistency, support-only tests, accessibility checks, visual goldens, retained artifacts, test names, or this plan.

### 3. File-by-file inspection checklist

- `apps/web/src/App.tsx`: inspect `AuthShellState`, `LandingRefreshState`, `refreshShell`, incident loading, route reset, and 401/403 handling before edits.
- `apps/web/src/AppRoot.tsx`: confirm StrictMode/root wiring needs no behavior beyond exercising `App` consistently in tests.
- `apps/web/src/Phase1Surface.tsx`: inspect auth, MFA, account, and admin error rendering; ensure MFA token use is not conflated with session state.
- `apps/web/src/browserApi.ts`: inspect `APIError`, `APIResult`, `fetchJSON`, and `extractError`; plan any public-error view model or sanitizer here.
- `apps/web/src/appShellTestSupport.ts`: extend fixtures for deferred loading, public error envelopes, forbidden, revoked, MFA-required, and private-diagnostic leak probes.
- `apps/web/src/App.phase1.test.tsx`: add row-owned `FE-U-P1-01` unit coverage here.
- `apps/web/src/App.phase1.support.test.tsx`: support-only coverage may verify helpers but must not be cited as `FE-U-P1-01` closure evidence by itself.

### 4. Test-first sequence

- Add failing unit tests before implementation changes.
- Cover initial loading with a deferred `/api/v1/auth/session`; loading must be visibly distinct from anonymous and ready-to-sign-in states.
- Cover anonymous from an initial `401 session_required` response.
- Cover `mfa_required` login response with required TOTP details and no session creation.
- Cover `mfa_setup_required` as an MFA-required setup variant; store the bootstrap token only for follow-up requests and do not render it as safe public detail.
- Cover authenticated bootstrap after session, credential state, and incident list success.
- Cover forbidden `403 authorization_denied` while retaining the authenticated session and rendering only public envelope fields.
- Cover revoked/session-loss after a previously authenticated state by clearing session and showing a re-auth or revoked state, not a generic anonymous first-load state.
- Cover public error-envelope rendering with safe details plus private diagnostic fields that must not appear.

### 5. Implementation tasks

- Refactor app bootstrap into an explicit deterministic state model if the current split state cannot represent all required states without ambiguity.
- Keep loading separate from anonymous; a pending session probe must not be treated as no session.
- Keep MFA state separate from authenticated state; `mfa_required` and `mfa_setup_required` must not create or imply a server session.
- Treat `session_required` after prior authentication as revoked or session-ended state; treat initial `session_required` as anonymous.
- Treat 403 public envelopes as forbidden while preserving the current session unless `/api/v1/auth/session` later says otherwise.
- Centralize public error rendering so all Sprint 2 surfaces use the same allowlist.
- Do not depend on private server diagnostics, stack text, internal paths, SQL messages, raw transport exception messages, or unknown `error.details` members.

### 6. Public error-envelope rendering contract

- Render only public `error.code`, public status text, and safe public details.
- Public status text means envelope `error.message` when present and display-safe, otherwise a deterministic fallback from `error.status` or the response status.
- Safe Sprint 2 detail allowlist: `reason_code`, `field`, `required_role`, `required_second_factor_kinds`, `required_setup_kinds`, and `bootstrap_expires_at`.
- Never render `bootstrap_token`, `secret_base32`, `otpauth_uri`, password values, provider assertions, `request_id`, stack traces, diagnostic objects, internal file paths, SQL text, or unknown detail keys.
- Add tests proving unknown/private detail keys are ignored even when the envelope otherwise renders.

### 7. State transition coverage

- `loading`: app starts or refreshes while the session probe is pending.
- `anonymous`: initial session probe returns `session_required`; login surface is ready.
- `mfa_required`: login returns `mfa_required` or `mfa_setup_required`; no session is established.
- `authenticated`: session probe succeeds and app loads credential state plus visible incidents.
- `forbidden`: authenticated public route returns 403; show forbidden/error state with public envelope only.
- `revoked`: previously authenticated session later returns `session_required` or a protected route returns 401; clear session and require re-auth.
- `public_error_envelope`: any non-success JSON envelope displays through the shared public rendering contract.

### 8. Validation commands

- Required row-owned Sprint 2 evidence: `make frontend-unit`.
- Required when TypeScript state, API, or error type surfaces change: `make frontend-typecheck`.
- Run `git diff --check`; run `git diff --cached --check` if changes are staged.
- Do not run or edit generated ledgers unless row metadata changes; if metadata changes, regenerate through the generator, never by hand.

### 9. Deliverables

- Row-owned unit coverage for loading, anonymous, MFA-required, authenticated, forbidden, revoked, and public error-envelope states.
- Any app-shell state helper and public-error rendering helper needed for deterministic tests.
- Recorded `make frontend-unit` evidence or a precise blocker.
- Recorded `make frontend-typecheck` evidence when type surfaces changed, or a precise blocker.
- No generated-ledger edit unless row metadata changes and the ledger is regenerated through the normal generator.

### 10. Blockers and source limits

- Existing Phase 1 unit/support test names are not completion evidence until the row-owned command passes after Sprint 2 changes.
- If a public error detail is not in the allowlist, do not render it; record `TODO: owner lookup required` only when Core or contract ownership is genuinely unresolved.
- If forbidden-state code ownership is unclear for a route, block that scenario rather than inventing a frontend-only error code.
- If `make frontend-unit` cannot run, record command, failure point, artifact paths when available, owner class, row ID, and minimum follow-up.

### 11. Binary acceptance criteria

- Sprint 2 remains scoped to `FE-U-P1-01`.
- The implementation plan anchors only to `apps/web/src/App.tsx`, `apps/web/src/AppRoot.tsx`, `apps/web/src/Phase1Surface.tsx`, `apps/web/src/browserApi.ts`, `apps/web/src/appShellTestSupport.ts`, `apps/web/src/App.phase1.test.tsx`, and `apps/web/src/App.phase1.support.test.tsx`.
- Loading is testably separate from anonymous.
- Server-managed session authority is preserved.
- Loading, anonymous, MFA-required, authenticated, forbidden, revoked, and public error-envelope states all have row-owned unit coverage.
- Public errors render only `error.code`, public status text, and safe public details.
- Private diagnostics are not rendered or required by tests.
- `make frontend-unit` passes as direct row-owned evidence, or a precise blocker is recorded.
- `make frontend-typecheck` passes when type surfaces change, or a precise blocker is recorded.
- No `FE-U-P1-01` completion claim is made from prohibited evidence sources or from this plan.

### 12. Sprint 2 completion evidence

At Sprint 2 closure, implementation completed only the `FE-U-P1-01` unit row. The phase registry remained `status="planned"`, and no `FE-I-P1-01`, `FE-E-P1-01`, `FE-A11Y-P1-01`, `FE-S-P1-01`, or Core 05 publication evidence was claimed.

Validation artifacts:

| Command | Status | Artifact |
| ------- | ------ | -------- |
| `make frontend-unit` | Passed | `.cartulary/test-results/20260529T222359Z-p2202982/frontend-unit/tool-run-summary.json` |
| `make frontend-typecheck` | Passed | `.cartulary/test-results/20260529T222359Z-p2202985/frontend-typecheck/tool-run-summary.json` |
| `make phase-ledgers` | Passed | `.cartulary/test-results/20260529T221959Z-p2192852/phase-ledgers/tool-run-summary.json` |
| `make phase-ledger-drift` | Passed | `.cartulary/test-results/20260529T222417Z-p2204673/phase-ledger-drift/tool-run-summary.json` |

Generated artifacts:

- `tools/frontend_phase_maps/fe_p1_test_map.json` now marks only `FE-U-P1-01` as `claim_status="implemented"` and records the exact existing `App.phase1.test.tsx` Vitest titles used as `scenario_titles`.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md` was regenerated through `make phase-ledgers`; it was not hand-edited.

## Sprint 3: API Client And Route-Boundary Integration Baseline

- Objective: Ensure the frontend API client and route-boundary tests preserve public `/api/v1/` routing, server-managed sessions, and public error-envelope rendering.
- Status: Complete for `FE-I-P1-01` route-boundary evidence after audit follow-up remediation. FE-P1 overall remains planned because `FE-E-P1-01`, `FE-A11Y-P1-01`, and `FE-S-P1-01` remain blocked.
- Relevant IDs: Owns `FE-I-P1-01`.
- Files and areas to inspect or edit: `apps/web/src/phase1Client.ts`, `apps/web/src/browserApi.ts`, `apps/web/src/fetchMockTestSupport.ts`, `apps/web/src/App.phase1.test.tsx`, `/packages/protocol-ts`, `/packages/ui-contracts`, and route-boundary fixtures.
- Test-first sequence: Add or tighten integration-style unit tests for every FE-P1 client route family before implementation changes; assert method, path, credentials, CSRF behavior, body shape, and error rendering.
- Implementation tasks: Keep all app client requests under `/api/v1/`; preserve default `credentials="include"` for server-managed sessions; use bootstrap-token credentials exceptions only for bootstrap MFA routes; include CSRF headers only when cookies and request method require them; reject private error details in rendered UI; test unknown closed request members where route owners require closure.
- Validation commands: `make frontend-unit`; supporting `make frontend-typecheck`, `make generated-artifact-policy-check`, `make generate-drift`, and `make frontend-import-boundary-check` when implementation touches contracts, generated inputs, or package boundaries.
- Deliverables: API route-boundary tests and client behavior that can be traced to Core-owned public routes and public error envelopes.
- Risks and assumptions: FE-P1 must not infer new route closure policy from frontend convenience; closure expectations must follow route owners.
- Blockers and follow-up notes: No Core 00-04 owner text, OpenAPI, generated protocol, or public API behavior changes were required. Server-side rejection of unknown request members remains backend/API owner evidence; `FE-I-P1-01` asserts that the frontend client does not send unknown members on exercised owner-closed Phase 1 routes. The audit-identified gaps for route-specific CSRF evidence, login second-factor serialization, logout/TOTP-complete/load-user/revoke-all public-error rendering, FE-P1 map titles, and generated ledger traceability are remediated by the validation artifacts below. A later audit found no remaining FE-I route-boundary product gap, but `make frontend-unit` was blocked by an unrelated Phase 9 jsdom grouped-paste conflict test; the audit blocker and follow-up remediation are recorded below.

### Sprint 3 remediation checklist

- Route-boundary request evidence asserts method, `/api/v1/` confinement, `credentials="include"`, absent `Authorization`, and route-specific cookie-backed CSRF behavior for the exercised Phase 1 auth, account, deployment-admin, and incident route families, including read-route no-CSRF behavior.
- Closed request-body evidence asserts exact serialized keys for primary and TOTP-assisted login, logout `{}`, session-mode TOTP complete, admin password reset, admin TOTP reset, revoke-all sessions, and the previously covered owner-closed client helpers.
- Bootstrap-token evidence remains limited to `POST /api/v1/auth/mfa/totp/begin` and `POST /api/v1/auth/mfa/totp/complete` in bootstrap mode, with no cookie credentials or CSRF on those bearer-token calls.
- Incident landing evidence directly asserts both `GET /api/v1/incidents` and `POST /api/v1/incidents` use cookie-backed sessions, omit `Authorization`, and send CSRF on the mutating request when the CSRF cookie exists.
- Public error-envelope evidence is row-owned for auth login, credential-state/account bootstrap, logout, account password/TOTP begin and complete, deployment-admin load/create/patch/password-reset/TOTP-reset/revoke-all actions, and incident create. Tests assert public `code`, `message`, `reason_code`, `field`, and `required_role` where allowed, and reject private probes such as request IDs, bootstrap tokens, secrets, paths, SQL, stacks, and unknown private details.
- Scenario titles in `tools/frontend_phase_maps/fe_p1_test_map.json` match the FE-I-labeled tests in `apps/web/src/phase1Client.routeBoundary.test.ts` and `apps/web/src/App.phase1.test.tsx`.

Validation artifacts:

| Command | Status | Artifact |
| ------- | ------ | -------- |
| `make frontend-unit` | Passed; `cartulary.frontend_row_accounting` closes `FE-I-P1-01` with 11/11 scenarios passed | `.cartulary/test-results/20260530T004132Z-p2464041/frontend-unit/tool-run-summary.json` |
| `make frontend-typecheck` | Passed | `.cartulary/test-results/20260530T004157Z-p2465698/frontend-typecheck/tool-run-summary.json` |
| `make phase-ledgers` | Passed | `.cartulary/test-results/20260530T004206Z-p2466008/phase-ledgers/tool-run-summary.json` |
| `make phase-ledger-drift` | Passed | `.cartulary/test-results/20260530T004209Z-p2466250/phase-ledger-drift/tool-run-summary.json` |
| `make generated-artifact-policy-check` | Passed | `.cartulary/test-results/20260530T004213Z-p2466491/generated-artifact-policy-check/tool-run-summary.json` |
| `make generate-drift` | Passed | `.cartulary/test-results/20260530T004722Z-p2480005/generate-drift/tool-run-summary.json` |
| `make frontend-import-boundary-check` | Passed | `.cartulary/test-results/20260530T004221Z-p2467353/frontend-import-boundary-check/tool-run-summary.json` |
| `make json-shape-check` | Passed | `.cartulary/test-results/20260530T004228Z-p2467671/json-shape-check/tool-run-summary.json` |
| `bash scripts/test-run-frontend-unit.sh` | Passed; covers frontend row pass/fail/missing/target-blocked cases and manifest supplemental Vitest support failure accounting | Console output only |
| `make lint-scripts` | Passed | `.cartulary/test-results/20260530T004630Z-p2478313/lint-scripts/tool-run-summary.json` |
| `make lint-shell` | Passed | `.cartulary/test-results/20260530T004634Z-p2478586/lint-shell/tool-run-summary.json` |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P1` | Passed; reports `FE-I-P1-01 claim_status=implemented` and FE-P1 overall `planned` | Console output only |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260530T003729Z-p2453010` | Failed in retained-run preflight because `agent-finalize` currently requires a successful full warm `make check` retained root, not a standalone `frontend-unit` run root | `.cartulary/test-results/20260530T003937Z-p2460587/agent-finalize/tool-run-summary.json` |
| `make agent-finalize` | Passed; updated `tools/execution_topology_render_index.json`; retained-run maintenance skipped because `RESULTS_DIR` was unset | `.cartulary/test-results/20260530T004012Z-p2461540/agent-finalize/tool-run-summary.json` |

Validation notes:

- The audit-only `make frontend-unit` run at `.cartulary/test-results/20260530T001459Z-p2412280` failed in unrelated Phase 9 test `Phase 9 E-9-02 registers grouped paste conflicts without losing selection continuity`. The runner JSON reported all 11 `FE-I-P1-01` scenario titles as passed, but the target failed overall, so that run is not FE-I closure evidence.
- The Phase 9 grouped-paste jsdom failure was remediated without quarantine by adding bounded async waits around paste conflict registration, conflict panel update, and focus-anchor restoration in `apps/web/src/WorkbookShell.phase9.sentinel.test.tsx`.
- The same Phase 9 jsdom test is now explicit supplemental frontend-unit evidence through `I-9-SUPPORT-01` in `tools/phase9_test_map.json`; browser row `E-9-CONFLICT-02` remains the authoritative conflict evidence owner. `make phase-ledgers` regenerated `docs/testing/phase9_coverage_ledger.md`.
- `frontend-unit` summaries now include additive `cartulary.frontend_row_accounting` extensions derived from frontend phase maps. Row closure remains strict: `FE-I-P1-01` is closed only when all mapped scenarios pass and the canonical target exits 0; unrelated target failures are reported as `blocked_by_target`.

## Sprint 4: Login, Session Bootstrap, Incident Entry, Authorization, And Revocation E2E Flow

### 1. Sprint Objective And Non-Goals

Objective: complete `FE-E-P1-01` with browser-observed E2E evidence for login, server-managed session bootstrap, incident entry, authorization, revocation, and re-authentication through public `/api/v1/` routes.

Non-goals: workbook-shell layout, mutation replay, accessibility closure, selector-builder promotion, visual fixtures, FE-P2 work, generated-ledger hand edits, and Core 05 publication claims. Touch those only if they block the `FE-E-P1-01` browser flow.

### 2. Source Limits And Assumptions

Core 00-04 own product behavior. `docs/guides/cartulary_frontend_implementation_testing_guide.md` owns the FE-P1 row shape. `docs/testing-harness-nlspec.md` owns harness mechanics only.

Product-conformance evidence for Sprint 4 must be direct browser E2E evidence from `make browser-e2e-webserver-backed`. Unit tests, support tests, generated ledgers, retained old artifacts, and inferred backend evidence are supporting evidence only.

Assume existing `E-1-01`, `E-1-02`, `E-1-03`, `E-1-05`, `E-1-06`, `E-1-07`, and `E-1-08` can be reused only after the webserver-backed target passes and their exact titles are listed under `FE-E-P1-01`.

### 3. Files And Areas To Inspect Before Editing

Inspect first:

- `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md`
- `tools/frontend_phase_maps/fe_p1_test_map.json`
- `apps/web/e2e/phase1.spec.ts`
- `apps/web/e2e/phase1Page.ts`
- `apps/web/e2e/authRuntime.ts`
- `apps/web/e2e/helpers.ts`
- `apps/web/src/App.tsx`
- `apps/web/src/Phase1Surface.tsx`
- `apps/web/src/IncidentAdminPanel.tsx`
- `apps/web/src/WorkbookShell.tsx`
- `packages/ui-contracts/src/index.ts`

Do not edit generated ledgers directly. If phase-map metadata changes, regenerate with `make phase-ledgers`.

### 4. Test-First Scenarios And Expected Browser-Visible Outcomes

Add or tighten Playwright scenarios before implementation changes.

- Local login success: user signs in on `/`, browser observes `/api/v1/auth/login`, `/api/v1/auth/session`, `/api/v1/auth/credential-state`, and `/api/v1/incidents`; session panel shows `provider_type=local`, user ID, MFA state, expiry fields, deployment-admin flag, and `memberships[]`.
- Invalid credentials: login renders `auth-error-code=invalid_credentials`, no session cookie is issued, and the login surface remains visible.
- MFA required/setup: omitted TOTP renders `mfa_required`; wrong TOTP renders `invalid_second_factor`; no session cookie exists until valid TOTP succeeds.
- Bootstrap TOTP: `mfa_setup_required` exposes only safe setup UI, TOTP begin/complete use bootstrap-token routes, completion alone issues no session, and a later normal login with the generated TOTP succeeds.
- Password change revocation: wrong current password and missing TOTP render public errors; successful password change returns the browser to login/re-auth state, old password fails, new password succeeds.
- Explicit revoke-all and re-auth: a deployment admin revokes all sessions for a logged-in target through `/api/v1/users/{user_id}/sessions/revoke-all`; the target browser's next public session/incident action returns to login, then fresh login succeeds.
- Deployment-admin controls: deployment-admin session sees and can use user-admin controls; incident-admin-only session sees the denial note and no deployment-user action controls; any direct public-route denial uses the public error envelope only.
- Incident list/create/open: authenticated user creates an incident from landing through `POST /api/v1/incidents`; the created incident appears in the visible list; opening it navigates to `/?incident_id=<id>` and renders `workbook-shell-ready` with current role `admin`.
- Stale incident selection: after membership removal for the selected incident, reload or refresh returns to landing, removes that incident from the list, shows the stale-selection notice, preserves the still-valid session, and allows opening another visible incident.
- Forbidden incident observation: start as reviewer so edit controls render, demote the same user to editor/admin-lower role through public membership mutation, then submit the stale edit control; browser observes `403` with `error.code=authorization_denied`, `incident-admin-error-code` renders that public code, and private diagnostics are not rendered.
- Current-role/current-membership effects: role changes affect the next public request; removed membership fails closed for that incident without revoking the account session for other visible incidents.

### 5. Implementation Work Packages

1. Browser tests: extend `apps/web/e2e/phase1.spec.ts` with the missing incident create/open, stale selection, forbidden role-change, and explicit revoke-all re-auth scenarios. Keep assertions on visible UI plus public route responses.
2. Page helpers: extend `Phase1Page` only for repeated Sprint 4 actions such as landing create/open, refresh, patch incident fields, and account refresh. Prefer existing `@cartulary/ui-contracts` selectors where already available.
3. Minimal app fixes: adjust `App.tsx`, `IncidentAdminPanel.tsx`, or `Phase1Surface.tsx` only if the tests expose a real FE-P1 flow gap. Preserve public error-envelope rendering via `publicErrorView`; do not render private fields.
4. Phase map: replace the generic `FE-E-P1-01` `scenario_titles` with exact Playwright titles that cover the scenarios above. Set `claim_status="implemented"` only after the primary validation passes.
5. Ledger: run `make phase-ledgers` after map metadata changes; never hand-edit `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md`.

### 6. Validation Commands And Pass/Fail Interpretation

Primary validation path: `make browser-e2e-webserver-backed`, which invokes `scripts/run-browser-e2e-webserver-backed.sh`, `apps/web/playwright.webserver-backed.config.ts`, and the selected `apps/web/e2e/phase1.spec.ts` browser-functional scenarios.

Required after map edits: `make phase-ledgers` and `make phase-ledger-drift`.

Run `make frontend-typecheck` for TypeScript changes and `make frontend-unit` only when app-state/client logic changes. Run `git diff --check`.

Pass means the primary target exits 0 and every exact `FE-E-P1-01` scenario title is present and passed. If the target is nonzero, scenarios are skipped/missing, or evidence comes only from unit/support/ledger artifacts, `FE-E-P1-01` stays blocked.

### 7. Evidence And Artifact Recording

Record in `FRONTEND_PHASE1_IMPLEMENTATION_PLAN.md`:

- command, status, run root, and summary path for `make browser-e2e-webserver-backed`;
- Playwright scenario titles mapped to `FE-E-P1-01`;
- whether evidence is product-conformance or supporting evidence;
- any `frontend-unit`, `frontend-typecheck`, `phase-ledgers`, and `phase-ledger-drift` artifacts;
- blockers using the existing blocker format.

Use `make explain-run RESULTS_DIR=<run-root> TARGET=browser-e2e-webserver-backed DETAIL=logs` if artifact paths need confirmation.

Sprint 4 execution evidence:

- Product-conformance evidence: `make browser-e2e-webserver-backed` passed with 61 tests, 0 failed, and 0 missing. Run root: `.cartulary/test-results/20260530T011915Z-p2549929`; summary: `.cartulary/test-results/20260530T011915Z-p2549929/browser-e2e-webserver-backed/tool-run-summary.json`.
- `FE-E-P1-01` mapped Playwright titles: `E-1-01 signs in as a local user and inspects the ordinary session surface`; `E-1-02 requires MFA on the ordinary login surface, rejects wrong codes, and accepts a valid TOTP code`; `E-1-03 rejects invalid credentials without issuing a session cookie`; `E-1-06 follows the bootstrap-token enrollment sequence on the ordinary login shell and proves completion alone issues no session`; `E-1-07 requires the current password and current TOTP code, revokes the session immediately, and requires re-login with the new password`; `E-1-08 keeps deployment-user administration on deployment-admin sessions and hides it from an incident admin on the ordinary shell`; `E-1-09 creates an incident from the landing screen, lists it, and opens the workbook as incident admin`; `E-1-10 clears a stale selected incident after membership removal while preserving the account session`; `E-1-11 observes current-role authorization on a stale reviewer edit through the public incident error envelope`; `E-1-12 returns a revoked target browser to login and allows re-authentication with unchanged incident membership`.
- Supporting evidence: `make frontend-typecheck` passed at `.cartulary/test-results/20260530T012107Z-p2558859/frontend-typecheck/tool-run-summary.json`; `make phase-ledgers` passed at `.cartulary/test-results/20260530T012107Z-p2558851/phase-ledgers/tool-run-summary.json`; `make phase-ledger-drift` passed at `.cartulary/test-results/20260530T012123Z-p2559684/phase-ledger-drift/tool-run-summary.json`; `git diff --check` passed with no output.
- Maintenance evidence: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260530T011915Z-p2549929` failed retained-run preflight because the finalizer requires a successful full warm `make check` retained root, not a browser-only run root. `make agent-finalize` then passed at `.cartulary/test-results/20260530T012135Z-p2560466/agent-finalize/tool-run-summary.json`, refreshed generated structure artifacts, and skipped retained-run duration/run checks because `RESULTS_DIR` was unset.
- Resolved blockers during execution: the first `make phase-ledgers` run failed because the base Phase 1 guide did not list newly added `E-1-09` through `E-1-12` rows; the guide row table was updated and the rerun passed. The first `make browser-e2e-webserver-backed` run failed on a test wait in `E-1-09`; the wait was corrected and the primary rerun passed. No Sprint 4 blocker remains for `FE-E-P1-01`.

### 8. Promotion/Blocker Rules For `FE-E-P1-01`

Promote `FE-E-P1-01` only when all are true:

- `make browser-e2e-webserver-backed` passes;
- all mapped Sprint 4 Playwright scenario titles pass;
- scenarios exercise public `/api/v1/` routes and server-managed sessions;
- public error envelopes remain visible and private diagnostics are not rendered;
- incident create/open, stale selection, forbidden role observation, current membership/role effects, revocation, and re-auth are covered;
- phase map and generated ledger drift pass.

Leave `FE-E-P1-01` blocked if any required scenario is absent, skipped, failed, not run, not public-route-backed, or only supported by unit/support/generated evidence.

### 9. Handoff Notes For Sprints 5 Through 7

Sprint 5 may reuse Sprint 4 states for accessibility, but `FE-A11Y-P1-01` remains design-direction evidence, not product conformance.

Sprint 6 promotes selector builders only where selectors cross stable package/test boundaries. Do not fold selector-builder promotion into Sprint 4 unless raw selectors block the E2E flow.

Sprint 7 must rerun row-owned commands, verify phase-ledger drift, record P0 regression status, and keep FE-P2 handoff limited to status and blockers for session, incident entry, authorization, public errors, accessibility, and selector ownership.

## Sprint 5: Accessibility Coverage For Session, MFA, Incident, Forbidden, Loading, And Error States

- Status: Complete for `FE-A11Y-P1-01`.
- Relevant IDs: Owns `FE-A11Y-P1-01`.
- Primary validation: `make browser-e2e-a11y`.
- Evidence class: `design_direction`; gates FE-P1 readiness only.

### 1. Objective And Non-Goals

Close `FE-A11Y-P1-01` by proving FE-P1 session, MFA, incident, forbidden, loading, revoked, and public-error states are keyboard reachable, visibly focused, screen-reader safe, and communicated with non-color-only cues.

Non-goals: Core product-conformance promotion, Core 05 claim-publication evidence, visual-golden acceptance, selector-builder promotion, new public API behavior, generated-ledger hand edits, or treating retained artifacts/test names/the plan itself as completion evidence.

### 2. Source And Authority Constraints

- Core 00 through Core 04 remain behavior authority. `docs/domain.md` supplies vocabulary for incident, user, session, membership, and evidence terms.
- `docs/guides/cartulary-ui-ux-design-guide.md` Sections 10.5 and 14 plus `docs/design.md` accessibility direction are design-direction inputs only.
- `docs/testing-harness-nlspec.md` owns `make browser-e2e-a11y` mechanics and `cartulary.frontend_accessibility_summary.v2`; the current retained summary is implemented-row evidence only. Blocked future-row smoke belongs to `make browser-e2e-a11y-preflight` and `cartulary.frontend_accessibility_preflight_summary.v1`.
- `FE-A11Y-P1-01` remains `design_direction`, claims no Core REQ/AC ownership, and does not activate Core 05.
- Row metadata was updated first in `tools/frontend_phase_maps/fe_p1_test_map.json`; `docs/testing/frontend_phase_coverage_ledgers/fe_p1_coverage_ledger.md` was regenerated by `make phase-ledgers`, not hand-edited.

### 3. File-By-File Inspection Checklist

- `apps/web/e2e/workbook.a11y.spec.ts`: replaced the generic FE-P1 smoke with exact FE-P1 state scenarios. Generic non-P1 smoke now belongs to the explicit preflight spec/target and does not complete blocked rows.
- `apps/web/e2e/phase1Page.ts`, `apps/web/e2e/helpers.ts`, and auth runtime helpers: reused public-route setup for users, MFA, incidents, membership changes, and session revocation.
- `apps/web/src/App.tsx`: inspected and fixed loading, authenticated landing, stale incident, forbidden, revoked, and public-error rendering for busy state, status/alert exposure, and focus recovery.
- `apps/web/src/Phase1Surface.tsx`: inspected and fixed login, `mfa_required`, `mfa_setup_required`, account, admin, retry, re-authentication, status, and public-error controls.
- `apps/web/src/IncidentAdminPanel.tsx`: inspected and fixed incident selected/error controls, read-only/denied states, membership controls, and incident error status.
- `apps/web/src/browserApi.ts`: verified public error rendering remains allowlisted and does not render private diagnostics.
- `apps/web/src/ReferencePackAdminPanel.tsx`: named the always-rendered deployment-admin file input so the FE-P1 all-controls accessibility check is not invalidated by an adjacent admin panel.
- `scripts/run-browser-e2e-a11y.sh` and `scripts/write-frontend-accessibility-summary.mjs`: extended retained summary evidence so executed FE-P1 scenarios map to `FE-A11Y-P1-01` under the strict v2 implemented-row schema.
- `tools/frontend_phase_maps/fe_p1_test_map.json`: replaced the generic FE-A11Y title with exact Sprint 5 titles and promoted `claim_status` only after direct passing evidence.

### 4. Test-First Accessibility Scenarios

Sprint 5 added these exact `FE-A11Y-P1-01` Playwright scenarios:

- `FE-A11Y-P1-01 deferred session loading exposes progress and keeps recovery controls keyboard reachable`
- `FE-A11Y-P1-01 anonymous login after initial session_required reaches login controls and authenticated landing`
- `FE-A11Y-P1-01 mfa_required challenge is keyboard reachable, visibly focused, named, and safely announced`
- `FE-A11Y-P1-01 mfa_setup_required enrollment is keyboard reachable and public errors hide private setup diagnostics`
- `FE-A11Y-P1-01 authenticated landing exposes account, admin, incident, retry, and visible incident controls`
- `FE-A11Y-P1-01 incident empty, list, selected, stale-selection, and incident-error states expose keyboard recovery`
- `FE-A11Y-P1-01 forbidden access-denied public envelope is announced and exposes recovery without private diagnostics`
- `FE-A11Y-P1-01 revoked session after prior authentication announces session end and supports re-authentication`
- `FE-A11Y-P1-01 generic public error envelope renders safe diagnostics and keyboard error recovery`

Each scenario asserts accessible names for interactive controls, keyboard reachability by role/name or stable test id, visible focus, no incorrect focus trap, required live/status/alert exposure, non-color-only state text, and absence of private diagnostics in public-error containers.

### 5. Implementation Tasks

- Added shared a11y test helpers for named controls, Tab/Shift+Tab traversal, visible-focus detection, live/status/alert checks, non-color-only cue checks, and private-diagnostic redaction checks.
- Extended the a11y summary path so per-scenario results populate `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, measured `contrast_checks[]`, `violations[]`, and `artifact_refs[]` for `FE-A11Y-P1-01`.
- Added roles/ARIA only where state changes needed programmatic exposure: loading/retry, forbidden, revoked/session-ended, incident errors, and public error summaries.
- Fixed app surfaces minimally: added accessible names, `role="status"`/`aria-live`/`aria-busy`/`role="alert"` where needed, preserved visible focus, and added state text where required.
- Kept public error display allowlisted. Public error rendering does not show stack traces, SQL text, internal paths, raw request IDs, secret material, bootstrap tokens, TOTP secrets in error details, or unknown private detail keys.
- Kept ordinary public `/api/v1/` session and incident flows. Deterministic private-diagnostic rendering fixtures intercept same-origin public-route responses only as frontend rendering evidence, not backend conformance.

### 6. Accessibility-State Coverage Matrix

| State | Required controls | Required checks | Sprint 5 result |
| --- | --- | --- | --- |
| Deferred session loading | login/retry path after load resolves | status/live or busy exposure, visible focus, no trap | Pass |
| Anonymous after `session_required` | email, password, TOTP, sign in | names, keyboard order, re-auth recovery | Pass |
| `mfa_required` | TOTP input, sign in retry | announced challenge, error not color-only, no session-only leakage | Pass |
| `mfa_setup_required` | begin enrollment, setup code, complete enrollment | setup state named, error details safe, focus remains recoverable | Pass |
| Authenticated landing | incident refresh/create/open, account refresh/logout, admin controls | all controls named and keyboard reachable | Pass |
| Incident empty/list/selected | empty message, incident cards, open workbook, return landing | non-color state cues, stable focus movement | Pass |
| Stale-selection/incident-error | stale notice, refresh/open other incident, incident retry/recovery | programmatic status/error exposure | Pass |
| Forbidden/access denied | public error code/message, retry or landing recovery | `authorization_denied` safe display, no private diagnostics | Pass |
| Revoked/session ended | sign-in controls after prior auth | announced session end, re-auth keyboard path | Pass |
| Generic public error | public code/message/details and retry/recovery | allowlisted details only, no private keys or secret text | Pass |

### 7. Validation Commands And Artifact Recording

Primary validation passed:

- `make browser-e2e-a11y` passed with run root `.cartulary/test-results/20260530T022308Z-p2711173`, summary `.cartulary/test-results/20260530T022308Z-p2711173/browser-e2e-a11y/tool-run-summary.json`, and accessibility summary `.cartulary/test-results/20260530T022308Z-p2711173/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`.
- Audit refresh: `make browser-e2e-a11y` passed with run root `.cartulary/test-results/20260530T144809Z-p3621062`, summary `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/tool-run-summary.json`, target summary `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/target-summary.json`, phase summary `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/browser-e2e-a11y-accessibility/phase-summary.json`, runner JSON `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/browser-e2e-a11y-accessibility/runner.json`, and accessibility summary `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`.
- Audit refresh retained summary result: `cartulary.frontend_accessibility_summary.v2` status `pass`; `phase_rows[]` contains only `FE-A11Y-P1-01` with `evidence_class="design_direction"` and target `make browser-e2e-a11y`; `scenarios[]`, `keyboard_matrix[]`, `state_communication_checks[]`, and `contrast_checks[]` each contain 9 passing entries; `violations[]` is empty; contrast records were retained under `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/accessibility/contrast-checks/`.

Supporting commands run because Sprint 5 changed TypeScript, app logic, summary tooling, and phase metadata:

- `make frontend-typecheck` passed with `.cartulary/test-results/20260530T022243Z-p2709340/frontend-typecheck/tool-run-summary.json`.
- `make frontend-unit` passed with `.cartulary/test-results/20260530T022251Z-p2709707/frontend-unit/tool-run-summary.json`.
- `make lint-scripts` passed with `.cartulary/test-results/20260530T022130Z-p2706276/lint-scripts/tool-run-summary.json`.
- `make phase-ledgers` passed with `.cartulary/test-results/20260530T022032Z-p2701945/phase-ledgers/tool-run-summary.json`.
- `make phase-ledger-drift` passed with `.cartulary/test-results/20260530T022035Z-p2702181/phase-ledger-drift/tool-run-summary.json`.
- `git diff --check` passed after implementation; no staged changes existed, so `git diff --cached --check` was not required.
- End-of-run maintenance: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260530T022308Z-p2711173` failed at retained-run preflight because the browser-only run root was not accepted as a retained-run maintenance input; rerun `make agent-finalize` passed with `.cartulary/test-results/20260530T022657Z-p2719401/agent-finalize/tool-run-summary.json`, skipped retained-run maintenance because `RESULTS_DIR` was unset, and updated `tools/execution_topology_render_index.json`.
- Audit refresh supporting checks passed: `make frontend-typecheck` with `.cartulary/test-results/20260530T144908Z-p3625733/frontend-typecheck/tool-run-summary.json`; `make frontend-unit` with `.cartulary/test-results/20260530T144918Z-p3626660/frontend-unit/tool-run-summary.json`; `make lint-scripts` with `.cartulary/test-results/20260530T144908Z-p3625734/lint-scripts/tool-run-summary.json`; `make phase-ledger-drift` with `.cartulary/test-results/20260530T144908Z-p3625786/phase-ledger-drift/tool-run-summary.json`; `git diff --check`; `git diff --cached --check`; and `make agent-finalize` with `.cartulary/test-results/20260530T144940Z-p3628324/agent-finalize/tool-run-summary.json` and finalize summary `.cartulary/test-results/20260530T144940Z-p3628324/agent-finalize/finalize-summary.json`. The audit-refresh finalizer reported generated artifacts unchanged and skipped retained-run maintenance because `RESULTS_DIR` was unset.

### 8. Promotion And Blocker Rules For `FE-A11Y-P1-01`

Promotion was allowed only after `make browser-e2e-a11y` passed, every exact FE-P1 scenario title was present and passed, the normalized accessibility summary mapped those scenarios to `FE-A11Y-P1-01`, all required matrix entries passed, private diagnostics were absent, and phase-ledger drift passed after metadata changed.

Do not mark `FE-A11Y-P1-01` complete from test names, visual goldens, retained artifacts alone, generated ledger consistency, support-only evidence, FE-P0 historical evidence, or this plan.

Blocker format, if Sprint 5 is reopened: `Blocked: FE-A11Y-P1-01. Command/scenario: <command or title>. Failure point: <assertion/target/tooling step>. Artifact: <path or unavailable>. Evidence class: design_direction. Owner class: <FE-P1-owned|harness-owned|infra-owned|outside-FE-P1>. Minimum follow-up: <smallest required action>.`

Final blocker state: none. The Sprint 5 audit refresh found no mismatch between implementation, phase metadata, generated ledger, normalized retained summary, and Sprint 5 acceptance criteria.

### 9. Deliverables

- Exact FE-P1 a11y Playwright scenarios and shared assertions in `apps/web/e2e/workbook.a11y.spec.ts`.
- Minimal FE-P1 app-surface accessibility fixes in `apps/web/src/App.tsx`, `apps/web/src/AppRoot.tsx`, `apps/web/src/Phase1Surface.tsx`, `apps/web/src/IncidentAdminPanel.tsx`, and `apps/web/src/ReferencePackAdminPanel.tsx`.
- Normalized accessibility summary evidence mapped to `FE-A11Y-P1-01`.
- Phase-map and regenerated ledger updates for exact scenario metadata and implemented claim status.
- Sprint 5 execution evidence recorded in this section.

### 10. Binary Acceptance Criteria

| Criterion | Required result | Sprint 5 result |
| --- | --- | --- |
| `make browser-e2e-a11y` | Passes, or precise blocker recorded | Pass |
| Scenario coverage | All required FE-P1 states executed directly | Pass: 9/9 mapped scenarios |
| Keyboard/focus | Required controls reachable, visibly focused, not trapped | Pass |
| Screen-reader safety | Loading, forbidden, revoked, and error changes programmatically exposed where needed | Pass |
| State cues | No required state relies on color alone | Pass |
| Error privacy | No private diagnostics in public-error rendering | Pass |
| Summary evidence | `frontend-accessibility-summary.json` maps executed evidence to `FE-A11Y-P1-01` | Pass |
| Classification | Evidence remains `design_direction`, not Core conformance or Core 05 | Pass |
| Ledgers | Generated only through `make phase-ledgers` when metadata changes | Pass |

### 11. Sprint 5 Completion-Evidence Recording Template

- Design-direction evidence: `make browser-e2e-a11y` passed; run root `.cartulary/test-results/20260530T022308Z-p2711173`; summary `.cartulary/test-results/20260530T022308Z-p2711173/browser-e2e-a11y/tool-run-summary.json`.
- Normalized accessibility summary: `.cartulary/test-results/20260530T022308Z-p2711173/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`; `phase_rows[]` contains `FE-A11Y-P1-01` with `evidence_class="design_direction"` and `claim_status="implemented"`; `scenarios[]` maps 9/9 FE-P1 scenarios to pass.
- Audit-refresh design-direction evidence: `make browser-e2e-a11y` passed; run root `.cartulary/test-results/20260530T144809Z-p3621062`; summary `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/tool-run-summary.json`.
- Audit-refresh normalized accessibility summary: `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/accessibility/frontend-accessibility-summary.json`; `phase_rows[]` contains only `FE-A11Y-P1-01` with `evidence_class="design_direction"` and `claim_status="implemented"`; `scenarios[]` maps 9/9 FE-P1 scenarios to pass; `keyboard_matrix[]`, `state_communication_checks[]`, and `contrast_checks[]` each contain 9 passing entries; `violations[]` is empty.
- Mapped scenario titles: the nine exact titles listed in Section 4.
- Required checks: keyboard matrix pass; state communication pass; contrast/non-color cues pass; private diagnostics pass; violations none.
- Supporting evidence: `make frontend-typecheck` pass `.cartulary/test-results/20260530T022243Z-p2709340/frontend-typecheck/tool-run-summary.json`; `make frontend-unit` pass `.cartulary/test-results/20260530T022251Z-p2709707/frontend-unit/tool-run-summary.json`; `make lint-scripts` pass `.cartulary/test-results/20260530T022130Z-p2706276/lint-scripts/tool-run-summary.json`; `make phase-ledgers` pass `.cartulary/test-results/20260530T022032Z-p2701945/phase-ledgers/tool-run-summary.json`; `make phase-ledger-drift` pass `.cartulary/test-results/20260530T022035Z-p2702181/phase-ledger-drift/tool-run-summary.json`; `make agent-finalize` pass `.cartulary/test-results/20260530T022657Z-p2719401/agent-finalize/tool-run-summary.json` after retained-run maintenance was skipped because `RESULTS_DIR` was unset; `git diff --check` pass; `git diff --cached --check` not run because no staged changes existed.
- Audit-refresh supporting evidence: `make frontend-typecheck` pass `.cartulary/test-results/20260530T144908Z-p3625733/frontend-typecheck/tool-run-summary.json`; `make frontend-unit` pass `.cartulary/test-results/20260530T144918Z-p3626660/frontend-unit/tool-run-summary.json`; `make lint-scripts` pass `.cartulary/test-results/20260530T144908Z-p3625734/lint-scripts/tool-run-summary.json`; `make phase-ledger-drift` pass `.cartulary/test-results/20260530T144908Z-p3625786/phase-ledger-drift/tool-run-summary.json`; `make agent-finalize` pass `.cartulary/test-results/20260530T144940Z-p3628324/agent-finalize/tool-run-summary.json`; `git diff --check` pass; `git diff --cached --check` pass.
- Phase metadata: `FE-A11Y-P1-01` implemented; `make phase-ledgers` and `make phase-ledger-drift` passed after metadata changes.
- Blockers: none.

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
| Generated ledgers are produced by generator, not hand-edited | Satisfied through Sprint 5 metadata: `make phase-ledgers` passed with `.cartulary/test-results/20260530T022032Z-p2701945/phase-ledgers/tool-run-summary.json` and `make phase-ledger-drift` passed with `.cartulary/test-results/20260530T022035Z-p2702181/phase-ledger-drift/tool-run-summary.json` |
| `FE-U-P1-01` has direct unit evidence or a blocker | Satisfied for Sprint 2 unit scope: `make frontend-unit` passed with `.cartulary/test-results/20260529T222359Z-p2202982/frontend-unit/tool-run-summary.json` and only `FE-U-P1-01` was promoted |
| `FE-I-P1-01` has direct integration/unit evidence or a blocker | Satisfied for Sprint 3 route-boundary scope: `make frontend-unit` passed with `.cartulary/test-results/20260530T004132Z-p2464041/frontend-unit/tool-run-summary.json`, the summary extension closes `FE-I-P1-01` with 11/11 scenarios passed, and the FE-P1 map lists row-owned FE-I scenario titles |
| `FE-E-P1-01` has browser E2E evidence or a blocker | Satisfied for Sprint 4 browser E2E scope: `make browser-e2e-webserver-backed` passed with `.cartulary/test-results/20260530T011915Z-p2549929/browser-e2e-webserver-backed/tool-run-summary.json`, all 10 mapped `FE-E-P1-01` scenario titles passed, and the row was promoted |
| `FE-A11Y-P1-01` has accessibility evidence or a blocker | Satisfied for Sprint 5 accessibility scope and audit refresh: `make browser-e2e-a11y` passed with `.cartulary/test-results/20260530T144809Z-p3621062/browser-e2e-a11y/tool-run-summary.json`, all 9 mapped scenarios passed, the retained accessibility summary maps evidence to `FE-A11Y-P1-01`, and the row remains `design_direction` |
| `FE-S-P1-01` has selector-builder support evidence or a blocker | Blocked for FE-P1 closure: no row-owned FE-P1 selector-builder evidence was promoted; Sprint 2 selector accounting remains app-local and `make browser-e2e-support` was not run |
| Public-route behavior stays under `/api/v1/` | Satisfied for `FE-I-P1-01` route-boundary evidence and Sprint 4 `FE-E-P1-01` browser-observed `/api/v1/` auth, session, incident, membership, and revoke-all flows |
| Frontend error rendering uses public error envelopes and not private server details | Satisfied for `FE-U-P1-01` unit scope, `FE-I-P1-01` route-boundary rendering evidence, and Sprint 4 `FE-E-P1-01` forbidden incident observation through `authorization_denied` without private diagnostics rendered |
| Server-managed session state is respected | Satisfied for `FE-U-P1-01` unit scope, `FE-I-P1-01` route-boundary client evidence, and Sprint 4 `FE-E-P1-01` browser evidence for login bootstrap, password-change revocation, explicit revoke-all, and re-authentication |
| Unknown closed request members are tested where route owners require closure | Satisfied for `FE-I-P1-01` client serialization evidence on exercised owner-closed Phase 1 request bodies; backend/API rows remain responsible for server-side rejection conformance |
| P0 regression checks remain green or exact blockers are recorded | Satisfied for Sprint 1 readiness: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, `make generate-drift`, `make frontend-import-boundary-check`, and `make phase-ledger-drift` passed |
| Support-only and design-direction evidence are not represented as product-conformance evidence | Satisfied through Sprint 5 metadata and audit refresh; current FE-P1 map separates product conformance, design direction, and implementation support, and `FE-A11Y-P1-01` remains `design_direction` |
| No Core 05 claim-publication review is activated unless explicitly requested | Satisfied for Sprint 1; this guide does not activate Core 05 publication review |
| Whitespace checks pass | Satisfied for Sprint 5 audit refresh: `git diff --check` and `git diff --cached --check` passed |

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
