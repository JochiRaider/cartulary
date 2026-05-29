# Frontend Phase 0 Implementation Plan

## Summary

This file is the execution roadmap, progress marker, and handoff aid for `FE-P0: Contract/Codegen Baseline And Package Boundaries`.

`docs/guides/cartulary_frontend_implementation_testing_guide.md` is the controlling FE-P0 planning row set and frontend phase-planning source. Core 00 through Core 04 remain the implementation-conformance behavior authority. Core 05 applies only to claim-bearing timed or fixture-sensitive publication boundaries.

This plan does not implement FE-P0 behavior. It is not behavior authority and must not be used to mark any FE-P0 row complete without direct local validation evidence.

## Authority Model

- Core 00 through Core 04 own implementation-conformance behavior.
- Core 05 owns claim-bearing publication only.
- Adopted NLSpecs may own bounded subsystem behavior only where explicitly adopted and applicable; `docs/testing-harness-nlspec.md` is harness guidance for FE-P0 mechanics, not product behavior authority.
- The frontend guide owns the FE-P0 planning row set and local frontend phase shape.
- The original Phase 5 through Phase 10 implementation plans are structural examples only.
- Generated artifacts, generated ledgers, generated schedules, support-only tests, visual goldens, retained run artifacts, and this plan are not behavior authorities.
- `docs/domain.md` is vocabulary and concept support only.

## Current Repo Status

The following facts were verified in the local repository during FE-P0 planning, Sprint 0 update inspection, and Sprint 1 audit inspection:

- `tools/frontend_phase_registry.json` exists and owns the frontend phase catalog under `phase_namespace="frontend"`.
- `tools/frontend_phase_maps/fe_p0_test_map.json` exists and lists the six FE-P0 rows exactly once.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md` exists as a generated companion ledger for the FE-P0 phase map.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` passes and reports `FE-P0` as `planned`; planned frontend phases are explainable but not executable.
- All six FE-P0 rows have `claim_status=implemented` in the current phase map.
- Package manifests exist for `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, and `/apps/web`.
- `/packages/ui` is future-reserved and intentionally inactive for FE-P0 closure; no empty package manifest is introduced.
- Target discovery verified these targets: `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, and `make phase-ledger-drift`.
- Source inspection found direct `react-data-grid` imports only under `/packages/grid-adapter/src/**`; `make frontend-import-boundary-check` now also enforces a singleton `react-data-grid/lib/styles.css` import owned by `/packages/grid-adapter/src/**`.
- `packages/protocol-ts/src/generated/**` exists, contains generated-file markers, and had no tracked diff during Sprint 1 audit inspection.
- No root-level frontend phase plan naming convention was found. This file uses the instructed root-level name `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`.
- Root-level `PHASE5_IMPLEMENTATION_PLAN.md` through `PHASE10_IMPLEMENTATION_PLAN.md` were not present; structural examples were found under `docs/archive/`.

Source limits:

- `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, and `make generate-drift` passed during Sprint 1 audit inspection. `FE-S-P0-01` was later closed by row-owned generated-artifact policy and drift validation.
- `make frontend-unit` and `make frontend-typecheck` passed during Sprint 2 remediation for `FE-U-P0-02` after correcting its owner reference to Core 03 Section 14.
- `make frontend-import-boundary-check` passed after adding RDG stylesheet singleton enforcement and closes `FE-S-P0-02`.
- `make phase-ledgers` and `make phase-ledger-drift` passed during Sprint 2 closure after the `FE-U-P0-02` status change; generated ledgers were refreshed by the generator.
- `make agent-finalize` passed during Sprint 2 closure and refreshed `tools/execution_topology_render_index.json` for the changed frontend phase manifest helper.
- `FE-U-P0-03` now cites Core 02 Section 18 for `REQ-02-222` and `REQ-02-223`.
- `FE-U-P0-03` failed audit until row-history display labels were separated from selector identity. Current remediation adds Core 01 `history_item_ref` response identity, keeps Core 02 Section 18 row-history rollback selectors separate, removes `operation` from row-history test IDs, centralizes retained selector families in `/packages/ui-contracts`, migrates retained app/E2E selectors to shared builders, and adds selector-policy coverage.
- Post-remediation `make check` passed with 161/161 work units, 766 tests, and 0 failures at `.cartulary/test-results/20260529T191817Z-p1857456/tool-run-summary.json`.
- The current FE-P0 phase map is generated-ledger-consistent after running the ledger generator and `make phase-ledger-drift`; no FE-P0 row remains blocked.

## Phase Objective

By FE-P0 exit, the frontend must have a stable contract/codegen and package-boundary foundation for later phases:

- Generated protocol types are consumed through facades without hand edits to generated output.
- View-schema adapters address surfaces by stable `view_schema_id` and fields by `field_key`.
- Selector and test-id builders are deterministic and label-independent; row-history selectors use `history_item_ref`, not `operation`, `revision_no`, visible text, or render order.
- `/apps/web` does not import `react-data-grid` directly.
- Generated-artifact policy and generated drift checks are enforceable.
- Frontend phase manifest and ledger mechanics are present.

FE-P0 must introduce no user-observable product behavior by itself.

## Implementation Scope

### In scope

- FE-P0 row manifest and ledger planning.
- Generated protocol facade consumption.
- View-contract adapter boundaries.
- Stable selector and test-id builders.
- Generated-artifact policy checks.
- Generated drift checks.
- Frontend import-boundary enforcement.
- RDG containment through `/packages/grid-adapter`.
- Frontend phase namespace and manifest/ledger scaffolding.
- Validation commands and blocker recording.

### Out of scope

- Workbook shell UI behavior.
- Runtime row mutation UX.
- Data-entry flows.
- Browser visual acceptance.
- Accessibility acceptance.
- Claim-bearing publication.
- Backend product behavior except the row-history response identity required for FE-U-P0-03 frontend contract consumption.
- Any direct edit to generated artifacts or generated ledgers.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| ---- | ------ | ------------------ | -------- | --------------- |
| [x] | 0. Frontend phase manifest and ownership setup; owns `FE-S-P0-03` | `make phase-ledger-drift` | None for `FE-S-P0-03` | Use `PHASE_NAMESPACE=frontend PHASE=FE-P0`; never append FE rows to base `PHASE=phaseN` maps |
| [x] | 1. Generated protocol facade and type-consumption baseline; owns `FE-U-P0-01` | `make frontend-typecheck`; `make frontend-unit` | None for `FE-U-P0-01`; Sprint 1 audit passed | Generated outputs were not hand-edited; generated-artifact checks were supporting evidence only |
| [x] | 2. View-schema and `field_key` adapter contract; owns `FE-U-P0-02` | `make frontend-unit`; `make frontend-typecheck` | None for `FE-U-P0-02`; owner trace corrected to Core 03 Section 14 | Adapters key fields by `field_key`, not labels, indexes, or visible order; duplicate-label coverage is present |
| [x] | 3. Stable selector and test-id builders; owns `FE-U-P0-03` | `make frontend-typecheck`; `make frontend-unit`; `make phase-slice PHASE=phase7`; `make service-backed-slice PHASE=phase7`; `make check` | None for `FE-U-P0-03`; owner trace corrected to Core 02 Section 18 and row-history response identity added under Core 01 | Builders derive from stable IDs, registry-backed `view_schema_id`, `field_key`, `history_item_ref`, and selector-relevant closed vocabularies; app/E2E consumers use shared builders |
| [x] | 4. Generated-artifact policy, drift, and codegen hygiene; owns `FE-S-P0-01` | `make generated-artifact-policy-check`; `make generate-drift` | None for `FE-S-P0-01` | Support-only evidence; generated outputs remain downstream and unedited by hand |
| [x] | 5. RDG import boundary and stylesheet ownership; owns `FE-S-P0-02` | `make frontend-import-boundary-check` | None for `FE-S-P0-02`; `/packages/ui` is future-reserved and inactive for FE-P0 closure | `/apps/web` consumes `/packages/grid-adapter`, direct RDG imports stay adapter-owned, and the RDG stylesheet is imported exactly once |
| [x] | 6. Frontend namespace, ledger drift, and phase-gate handoff; supports all FE-P0 rows | `make phase-ledger-drift`; row-specific commands above | None; all row owners complete | Generated ledgers are produced by generator, not hand-edited |

## Global References

- Controlling FE-P0 guide: `docs/guides/cartulary_frontend_implementation_testing_guide.md`, Phase `FE-P0`.
- FE-P0 rows: `FE-U-P0-01`, `FE-U-P0-02`, `FE-U-P0-03`, `FE-S-P0-01`, `FE-S-P0-02`, `FE-S-P0-03`.
- Core owner documents: `docs/spec/00_document_set_status_and_precedence.md`, `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/02_domain_model_schema_and_history.md`, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, `docs/spec/04_security_deployment_and_conformance.md`, and `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`.
- FE-P0 owner anchors cited by the frontend guide include Core 00 Section 1; Core 01 Sections 3.3.1, 3.3.4, and 7.4; Core 02 Section 18; Core 03 Section 14; development-guide Sections 2, 6.1, 6.2, 6.3, 6.4, 6.6, 6.8, 6.10, and 7.1; and implementation-guide Sections 14.8, 15, and 16.
- Package surfaces: `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, and `/apps/web`; `/packages/ui` is future-reserved and inactive for FE-P0 closure.
- Generated boundaries: `/packages/protocol-ts/src/generated/**`, `/internal/gen/**`, generated frontend ledgers under `/docs/testing/frontend_phase_coverage_ledgers/**`, generated schedules, generated Make includes, `pnpm-lock.yaml`, and `go.sum`.
- Command surface: `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, `make phase-ledger-drift`, and `git diff --check`.

## Evidence Layer Matrix

| Row | Evidence layer | Evidence class | Claim intent |
| --- | -------------- | -------------- | ------------ |
| `FE-U-P0-01` | Unit | Product conformance | Generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code; completion must cite confirmed Core owners and ACs |
| `FE-U-P0-02` | Unit | Product conformance | View-schema adapters key editable and queryable fields by `field_key`, not labels, indexes, or visible column order; completion evidence is recorded under Sprint 2 |
| `FE-U-P0-03` | Unit | Product conformance | Stable selector and test-id builders derive identifiers from stable IDs, row-history `history_item_ref`, registry-backed `view_schema_id` values, and selector-relevant closed vocabularies rather than visible labels; completion evidence is recorded under Sprint 3 |
| `FE-S-P0-01` | Support | Implementation support | Generated protocol policy and generated contract drift are enforced; this must not be represented as product conformance |
| `FE-S-P0-02` | Support | Implementation support | `/apps/web` consumes `/packages/grid-adapter` and does not import `react-data-grid` directly; this must not be represented as product conformance |
| `FE-S-P0-03` | Support | Implementation support | Frontend phase rows are recorded in namespace-specific registry, map, and generated ledger mechanics before any phase-enforced completion claim; this must not be represented as product conformance |

## Dependencies and Prerequisites

- Generated protocol artifacts exist under `/packages/protocol-ts/src/generated/**`; FE-P0 depends on consuming them through handwritten facades and must not hand-edit generated output.
- Frontend workspace package presence is intentional: `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, and `/apps/web` have manifests; `/packages/ui` remains future-reserved until a cohesive reusable presentational component surface exists.
- View-schema contracts are consumed through `/packages/view-contracts`; FE-P0 completion must validate stable `view_schema_id` and `field_key` behavior with unit evidence.
- Development-guide package boundaries govern `/apps/web`, `/packages/grid-adapter`, `/packages/ui-contracts`, `/packages/view-contracts`, `/packages/protocol-ts`, and `/packages/test-utils`; `/packages/ui` remains future-reserved.
- Make target availability was verified by target discovery; passing validation is not claimed for targets that were not run.
- Phase-map and ledger machinery exists for the frontend namespace; all FE-P0 rows are implemented.
- Existing generated-artifact policy tooling exists; latest remediation validation records passing `make generated-artifact-policy-check` and `make generate-drift` evidence for `FE-S-P0-01`.

## Public Interfaces and Deliverables

FE-P0 introduces no user-visible product routes, workflows, or runtime workbook shell behavior.

Deliverables are repository artifacts and validation surfaces:

- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`, this planning artifact.
- `tools/frontend_phase_registry.json`, existing frontend phase registry; any future update must preserve `phase_namespace="frontend"`.
- `tools/frontend_phase_maps/fe_p0_test_map.json`, existing FE-P0 row map; all six FE-P0 rows are implemented.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md`, existing generated companion ledger; must not be hand-edited.
- Package-boundary validation through `make frontend-import-boundary-check`.
- Generated-artifact validation through `make generated-artifact-policy-check` and `make generate-drift`.
- Unit/typecheck validation through `make frontend-typecheck` and `make frontend-unit`.
- Selector/test-id contract implementation, if changed later, must stay under appropriate frontend contract packages such as `/packages/ui-contracts` and must remain label-independent.

## Sprint 0: Frontend Phase Manifest And Ownership Setup

- Objective: Ensure FE-P0 uses the separate frontend phase namespace and has one authoritative row-map inventory plus generated ledger mechanics before any completion claim.
- Status: Implemented for `FE-S-P0-03`; FE-P0 remains `planned` and non-executable as a phase, while all FE-P0 rows now have row-owned evidence.
- Relevant IDs: Owns `FE-S-P0-03`; supports all FE-P0 rows.
- Files and areas: `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p0_test_map.json`, `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md`, phase tooling under `scripts/`, and Make phase-ledger targets.
- Test-first sequence: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` and `make phase-ledger-drift` passed during Sprint 0 audit inspection; inspect row statuses before changing any later claims.
- Implementation tasks: Completed for Sprint 0: namespace-aware FE-P0 registry, one FE-P0 map inventory, generated FE-P0 ledger mechanics, base-only `PHASE=phaseN` separation, and ambiguous frontend phase rejection.
- Validation commands: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`; `make phase-ledger-drift`; `git diff --check`.
- Deliverables: Namespace-aware FE-P0 representation and generated ledger are present; remaining blockers belong to row-owner sprints outside Sprint 0.
- Risks and assumptions: A drift pass only proves ledger/map agreement; it does not prove row completion.
- Blockers and follow-up notes: No blocker remains for `FE-S-P0-03`.

## Sprint 1: Generated Protocol Facade And Type-Consumption Baseline

- Objective: Ensure generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code.
- Status: Implemented for `FE-U-P0-01` by Sprint 1 audit inspection. `FE-S-P0-01` is closed separately by Sprint 4 support-row evidence.
- Relevant IDs: Owns `FE-U-P0-01`.
- Files and areas: `/packages/protocol-ts`, `/packages/protocol-ts/src/generated/**`, `/contracts/**`, generated-artifact policy configuration, and frontend consumers.
- Test-first sequence: Completed during Sprint 1 audit inspection: `make frontend-typecheck`, `make frontend-unit`, `make generated-artifact-policy-check`, and `make generate-drift` all passed.
- Implementation tasks: Completed or verified: generated output is behind handwritten facade exports, generated markers are preserved, consumers avoid direct generated-path imports except through approved facade code, and stable identifiers are covered by unit/typecheck evidence.
- Validation commands: `make frontend-typecheck` passed with summary `.cartulary/test-results/20260528T235835Z-p565709/frontend-typecheck/tool-run-summary.json`; `make frontend-unit` passed with summary `.cartulary/test-results/20260528T235846Z-p566137/frontend-unit/tool-run-summary.json`; `make generated-artifact-policy-check` passed with summary `.cartulary/test-results/20260528T235902Z-p567563/generated-artifact-policy-check/tool-run-summary.json`; `make generate-drift` passed with summary `.cartulary/test-results/20260528T235906Z-p567902/generate-drift/tool-run-summary.json`.
- Deliverables: Facade-backed protocol consumption and validation evidence are recorded for `FE-U-P0-01`; no generated protocol file diff was observed under `/packages/protocol-ts/src/generated/**`.
- Risks and assumptions: The audit started with pre-existing tracked changes in Sprint-related authored files and `pnpm-lock.yaml`; no new tracked mutation was introduced by the audit commands. Retained evidence artifacts were created under ignored `.cartulary/test-results/**`.
- Blockers and follow-up notes: No blocker remains for `FE-U-P0-01`.

## Sprint 2: View-Schema And `field_key` Adapter Contract

- Objective: Ensure view-schema adapters key editable and queryable fields by `field_key`, not labels, indexes, or visible column order.
- Status: Implemented for `FE-U-P0-02` by Sprint 2 remediation.
- Relevant IDs: Owns `FE-U-P0-02`.
- Files and areas: `/packages/view-contracts`, generated contract inputs, view-schema fixtures, and frontend tests that exercise adapter behavior.
- Test-first sequence: Unit coverage validates label/order-independent field addressing, duplicate labels, strict malformed-contract failures, and `view_schema_id` lookup; `make frontend-unit` passed.
- Implementation tasks: Completed: adapter lookup uses `view_schema_id`, editable/queryable field surfaces use `field_key`, malformed contracts fail closed, and tests keep presentation order scoped to rendering rather than semantic identity.
- Validation commands: `make frontend-unit` passed with summary `.cartulary/test-results/20260529T004532Z-p625790/frontend-unit/tool-run-summary.json`; `make frontend-typecheck` passed with summary `.cartulary/test-results/20260529T004546Z-p627115/frontend-typecheck/tool-run-summary.json`.
- Deliverables: Unit evidence for `field_key` identity behavior, strict contract parsing, and duplicate-label display-only behavior is recorded.
- Risks and assumptions: `field_key` is a domain vocabulary term and Core-owned concept; visible labels and column order must remain non-authoritative.
- Blockers and follow-up notes: No blocker remains for `FE-U-P0-02`; the invalid legacy Core 03 owner reference was corrected to Core 03 Section 14.

## Sprint 3: Stable Selector And Test-ID Builders

- Objective: Ensure selector and test-id builders derive identifiers from stable IDs, row-history `history_item_ref`, registry-backed `view_schema_id` values, and selector-relevant closed vocabularies rather than visible labels.
- Status: Implemented for `FE-U-P0-03`.
- Relevant IDs: Owns `FE-U-P0-03`.
- Files and areas: Core 01, OpenAPI and generated contracts, revisions history store/tests, `/packages/ui-contracts`, `/packages/test-utils`, frontend tests, and retained consumer selectors in `/apps/web`.
- Test-first sequence: Unit and slice coverage proves stable IDs, strict validation where owner syntax exists, opaque-ID non-empty/lossless encoding, row-history `history_item_ref` identity, display-only `operation`, selector-relevant closed vocabularies, registry-backed `view_schema_id` acceptance/rejection, CSS-safe `data-testid` selectors, and policy rejection of unowned raw selector literals/templates.
- Implementation tasks: Completed: row-history responses emit stable opaque `history_item_ref`; selector/test-id construction is centralized in `/packages/ui-contracts`; retained static row-history selectors are shared builders; row-history rendering fails closed on malformed or duplicate anchors; workbook runtime, reference-pack controls, app shell surfaces, E2E helpers/specs, and `/packages/test-utils` consume shared builders for retained cross-boundary selectors; selector policy now requires explicit app-local ownership metadata for retained raw literals.
- Validation commands: `make frontend-typecheck` passed with summary `.cartulary/test-results/20260529T200008Z-p1969529/frontend-typecheck/tool-run-summary.json`; `make frontend-unit` passed with summary `.cartulary/test-results/20260529T200008Z-p1969540/frontend-unit/tool-run-summary.json`; `make lint-biome` passed with summary `.cartulary/test-results/20260529T200008Z-p1969558/lint-biome/tool-run-summary.json`; `make phase-slice PHASE=phase7` passed with summary `.cartulary/test-results/20260529T200037Z-p1971697/phase-slice/tool-run-summary.json`; `make service-backed-slice PHASE=phase7` passed with summary `.cartulary/test-results/20260529T200612Z-p1982039/service-backed-slice/tool-run-summary.json`; earlier full-check evidence remains recorded by the original Sprint 3 closure artifacts.
- Deliverables: Core/OpenAPI `history_item_ref` contract, generated contract updates, backend history-item identity, selector/test-id contract evidence, migrated app/E2E/test-utils selector consumers, shared static row-history builders, registry-backed `view_schema_id` validation, opaque-ID contract coverage, and selector-policy ownership guard.
- Risks and assumptions: `data-testid` strings are internal test contracts, not public compatibility commitments.
- Blockers and follow-up notes: No blocker remains for `FE-U-P0-03`.

## Sprint 4: Generated-Artifact Policy, Drift, And Codegen Hygiene

- Objective: Enforce generated protocol policy and generated contract drift without treating support evidence as product conformance.
- Status: Implemented for `FE-S-P0-01`.
- Relevant IDs: Owns `FE-S-P0-01`.
- Files and areas: `tools/generated_artifact_policy.json`, `/packages/protocol-ts/src/generated/**`, `/contracts/**`, generator scripts, generated Make includes, and lockfiles.
- Test-first sequence: Run `make generated-artifact-policy-check`; run `make generate-drift`; inspect exact failing unit if either fails.
- Implementation tasks: Preserve generated markers; keep generated roots excluded from authored-source lint scopes where policy requires; separate codegen drift from migration drift; never hand-edit generated protocol output.
- Validation commands: `make generated-artifact-policy-check` passed with summary `.cartulary/test-results/20260529T191503Z-p1847895/generated-artifact-policy-check/tool-run-summary.json`; `make generate-drift` passed with summary `.cartulary/test-results/20260529T191503Z-p1847911/generate-drift/tool-run-summary.json`; final whitespace validation is recorded separately through `git diff --check`.
- Deliverables: Passing generated policy and drift checks.
- Risks and assumptions: This is implementation-support evidence only.
- Blockers and follow-up notes: No blocker remains for `FE-S-P0-01`.

## Sprint 5: RDG Import Boundary And Stylesheet Ownership

- Objective: Ensure `/apps/web` consumes `/packages/grid-adapter` and does not import `react-data-grid` directly.
- Status: Implemented for `FE-S-P0-02`.
- Relevant IDs: Owns `FE-S-P0-02`.
- Files and areas: `/apps/web`, `/packages/grid-adapter`, `tools/frontend_import_boundaries.json`, frontend package manifests, and stylesheet ownership points.
- Test-first sequence: Run `make frontend-import-boundary-check`; inspect import-boundary diagnostics before implementation.
- Implementation tasks: Keep all RDG imports inside `/packages/grid-adapter`; keep `/apps/web` on adapter exports; enforce exactly one RDG stylesheet import from `/packages/grid-adapter/src/**`; keep `/packages/ui` future-reserved until a manifest-backed reusable UI package is needed.
- Validation commands: `make frontend-import-boundary-check` passed with summary `.cartulary/test-results/20260529T191503Z-p1847875/frontend-import-boundary-check/tool-run-summary.json`.
- Deliverables: Passing import-boundary evidence with RDG vendor and stylesheet containment.
- Risks and assumptions: Source search found no direct `/apps/web` RDG import; selector/test-id strings and RDG stylesheet ownership are internal implementation contracts.
- Blockers and follow-up notes: No blocker remains for `FE-S-P0-02`.

## Sprint 6: Frontend Namespace, Ledger Drift, And Phase-Gate Handoff

- Objective: Collect FE-P0 evidence and handoff state for FE-P1 without duplicating row ownership or turning support-only evidence into product conformance.
- Status: Complete for FE-P0 handoff; no primary row ownership.
- Relevant IDs: Supports `FE-U-P0-01`, `FE-U-P0-02`, `FE-U-P0-03`, `FE-S-P0-01`, `FE-S-P0-02`, and `FE-S-P0-03`.
- Files and areas: FE-P0 plan, frontend phase registry/map/ledger, validation artifacts, and FE-P1 handoff notes.
- Test-first sequence: Re-run row-owned validation commands; run `make phase-ledger-drift`; run `git diff --check`.
- Implementation tasks: Verify every FE-P0 row is complete with evidence; ensure `PHASE_NAMESPACE=frontend PHASE=FE-P0` is the only frontend phase-selection spelling; keep base `PHASE=phaseN` untouched.
- Validation commands: Row-owned commands plus `make phase-ledgers`, `make phase-ledger-drift`, `make agent-finalize`, `make check`, `git diff --check`, and `git diff --cached --check`.
- Deliverables: FE-P1 handoff state with row statuses and validation artifact locations.
- Risks and assumptions: Active frontend phase execution is not implemented for planned phases; completion must not depend on planned-phase execution.
- Blockers and follow-up notes: No FE-P0 handoff blocker remains. Latest remediation artifacts: `.cartulary/test-results/20260529T195832Z-p1963336/phase-ledgers/tool-run-summary.json`, `.cartulary/test-results/20260529T195835Z-p1963617/phase-ledger-drift/tool-run-summary.json`, and `.cartulary/test-results/20260529T201213Z-p1992955/agent-finalize/tool-run-summary.json`; the retained-run maintenance path was skipped because `RESULTS_DIR` was unset for the successful finalizer run.

## Blocker Recording Rules

For every failed validation command or missing prerequisite, record:

- Exact command or missing artifact.
- Exact failing target or scheduler unit when available.
- Result root, run ID, run root, and artifact path when available.
- FE-P0 row ID when applicable.
- Failure class and failure reason when exposed.
- Ownership: FE-P0-owned, support-tooling-owned, harness-owned, infra-owned, or outside FE-P0.
- Minimum follow-up needed.
- Whether the blocker prevents FE-P0 completion or is support-only.

## Binary Exit Criteria

| Criterion | Current status |
| --------- | -------------- |
| The plan file exists and states that it is not behavior authority | Satisfied by this file once committed or otherwise accepted |
| FE-P0 is represented in the frontend phase namespace or a precise blocker prevents that representation | Present: registry, map, and ledger exist; all six rows are implemented |
| `tools/frontend_phase_maps/fe_p0_test_map.json` or equivalent planned manifest covers every FE-P0 row exactly once | Present by local inspection; `FE-U-P0-03` owner refs point to Core 01 Section 7.4, Core 02 Section 18, and development guide Section 6.4 |
| Generated frontend ledger behavior is present or explicitly blocked; any generated ledger is produced by a generator and not hand-edited | Present; `make phase-ledgers` and `make phase-ledger-drift` passed during Sprint 6 handoff |
| `FE-U-P0-01`, `FE-U-P0-02`, and `FE-U-P0-03` have direct unit/typecheck evidence or explicit blockers | Satisfied for all three FE-P0 unit rows |
| `FE-S-P0-01`, `FE-S-P0-02`, and `FE-S-P0-03` have implementation-support validation or explicit blockers | Satisfied: all three support rows have row-owned validation evidence |
| `/apps/web` direct `react-data-grid` imports are prohibited or the absence of an enforcement target is recorded as a blocker | Satisfied: `make frontend-import-boundary-check` enforces adapter-only RDG imports and singleton stylesheet ownership |
| Generated protocol artifacts are not hand-edited | Satisfied for Sprint 1: generated markers observed, no generated protocol diff found, and policy/drift checks passed |
| Generated-artifact policy and generate-drift checks pass or have precise blockers | Satisfied: `make generated-artifact-policy-check` and `make generate-drift` passed as Sprint 4 support-row evidence |
| Frontend phase selection uses `PHASE_NAMESPACE=frontend PHASE=FE-P0` or the absence of namespace-aware selection is recorded as a blocker | Present and verified by `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` |
| No support-only, visual, accessibility, or measurement evidence is represented as product-conformance evidence | Required; current map distinguishes product conformance from implementation support |
| No Core 05 claim-publication review is activated unless a claim-bearing publication intent is explicitly declared | Required; FE-P0 has no claim-bearing publication intent |
| `git diff --check` passes | Passed during Sprint 1 audit inspection; must be rerun after subsequent plan edits |

FE-P0 must not be marked complete until every criterion is satisfied with direct evidence or an explicit blocker.

## Handoff Requirements For FE-P1

FE-P1 must receive:

- Stable protocol facade consumption status.
- Stable view-contract adapter status.
- Selector and test-id builder status.
- Frontend package-boundary status.
- FE-P0 manifest and ledger status.
- Command-surface blockers and validation artifact locations.
- Any unresolved generated-artifact drift, target availability issue, or owner-lookup TODO.

Support-only status must stay visibly separate from product-conformance status in the FE-P1 handoff.
