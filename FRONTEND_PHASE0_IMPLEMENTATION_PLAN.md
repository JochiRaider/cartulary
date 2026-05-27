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

The following facts were verified in the local repository during FE-P0 planning inspection:

- `tools/frontend_phase_registry.json` exists and owns the frontend phase catalog under `phase_namespace="frontend"`.
- `tools/frontend_phase_maps/fe_p0_test_map.json` exists and lists the six FE-P0 rows exactly once.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md` exists as a generated companion ledger for the FE-P0 phase map.
- `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` passes and reports `FE-P0` as `planned`; planned frontend phases are explainable but not executable.
- Every FE-P0 row in the current phase map has `claim_status=blocked`.
- Package manifests exist for `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, and `/apps/web`.
- `/packages/ui` exists only as a directory with `.gitkeep`; no `/packages/ui/package.json` was verified.
- Target discovery verified these targets: `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, and `make phase-ledger-drift`.
- Source inspection found direct `react-data-grid` imports only under `/packages/grid-adapter/src/**`; FE-P0 completion still requires `make frontend-import-boundary-check`.
- `packages/protocol-ts/src/generated/**` exists and contains generated-file markers; FE-P0 completion still requires generated-artifact policy and drift validation.
- No root-level frontend phase plan naming convention was found. This file uses the instructed root-level name `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`.
- Root-level `PHASE5_IMPLEMENTATION_PLAN.md` through `PHASE10_IMPLEMENTATION_PLAN.md` were not present; structural examples were found under `docs/archive/`.

Source limits:

- `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, and `make generate-drift` were not run for this planning artifact.
- `make phase-ledger-drift` passed during planning inspection, but that does not complete `FE-S-P0-03` while the FE-P0 phase-map row remains blocked.
- `TODO: owner lookup required` remains for exact `Core 03 Section 4.8` and `Core 02 Section 5.3` anchors before using affected rows for authoritative product-conformance completion.
- The current FE-P0 phase map is generated-ledger-consistent, but row metadata must be reconciled against the controlling frontend guide before completion where guide ranges or support ACs are broader than the map entries.

## Phase Objective

By FE-P0 exit, the frontend must have a stable contract/codegen and package-boundary foundation for later phases:

- Generated protocol types are consumed through facades without hand edits to generated output.
- View-schema adapters address surfaces by stable `view_schema_id` and fields by `field_key`.
- Selector and test-id builders are deterministic and label-independent.
- `/apps/web` does not import `react-data-grid` directly.
- Generated-artifact policy and generated drift checks are enforceable.
- Frontend phase manifest and ledger mechanics are present or explicitly blocked.

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
- Backend product behavior not needed for FE-P0 frontend contract consumption.
- Any direct edit to generated artifacts or generated ledgers.

## Sprint Checklist

| Done | Sprint | Primary validation | Blockers | Follow-up notes |
| ---- | ------ | ------------------ | -------- | --------------- |
| [ ] | 0. Frontend phase manifest and ownership setup; owns `FE-S-P0-03` | `make phase-ledger-drift` | Existing FE-P0 rows are still blocked; reconcile map metadata with the frontend guide where ranges or support ACs differ | Use `PHASE_NAMESPACE=frontend PHASE=FE-P0`; never append FE rows to base `phaseN` maps |
| [ ] | 1. Generated protocol facade and type-consumption baseline; owns `FE-U-P0-01` | `make frontend-typecheck`; `make frontend-unit` | Row blocked; no current direct unit/typecheck evidence | Generated outputs must not be hand-edited |
| [ ] | 2. View-schema and `field_key` adapter contract; owns `FE-U-P0-02` | `make frontend-unit` | Row blocked; `TODO: owner lookup required` for exact Core 03 Section 4.8 heading before product-conformance completion | Adapters must key fields by `field_key`, not labels, indexes, or visible order |
| [ ] | 3. Stable selector and test-id builders; owns `FE-U-P0-03` | `make frontend-unit` | Row blocked; `TODO: owner lookup required` for exact Core 02 Section 5.3 heading before product-conformance completion | Builders must derive from stable IDs and closed vocabularies |
| [ ] | 4. Generated-artifact policy, drift, and codegen hygiene; owns `FE-S-P0-01` | `make generated-artifact-policy-check`; `make generate-drift` | Row blocked; validation targets were discovered but not run for this plan | Support-only evidence |
| [ ] | 5. RDG import boundary and stylesheet ownership; owns `FE-S-P0-02` | `make frontend-import-boundary-check` | Row blocked; `/packages/ui` package surface is not fully present | `/apps/web` must consume `/packages/grid-adapter` and must not import RDG directly |
| [ ] | 6. Frontend namespace, ledger drift, and phase-gate handoff; supports all FE-P0 rows | `make phase-ledger-drift`; row-specific commands above | No primary row ownership; handoff remains blocked until row owners complete or record precise blockers | Generated ledgers must be produced, not hand-edited |

## Global References

- Controlling FE-P0 guide: `docs/guides/cartulary_frontend_implementation_testing_guide.md`, Phase `FE-P0`.
- FE-P0 rows: `FE-U-P0-01`, `FE-U-P0-02`, `FE-U-P0-03`, `FE-S-P0-01`, `FE-S-P0-02`, `FE-S-P0-03`.
- Core owner documents: `docs/spec/00_document_set_status_and_precedence.md`, `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/02_domain_model_schema_and_history.md`, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, `docs/spec/04_security_deployment_and_conformance.md`, and `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`.
- FE-P0 owner anchors cited by the frontend guide include Core 00 Section 1; Core 01 Sections 3.3.1, 3.3.4, 7.4, and 8.5; Core 02 Section 5.3; Core 03 Sections 4.1 and 4.8; development-guide Sections 2, 6, 6.1, 6.2, 6.3, 6.4, 6.6, 6.8, 6.10, and 7.1; and implementation-guide Sections 7, 14, 15, and 16.
- `TODO: owner lookup required` applies to exact Core 02 Section 5.3 and Core 03 Section 4.8 anchors until those headings are located or corrected by owner text.
- Package surfaces: `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, `/packages/ui`, and `/apps/web`.
- Generated boundaries: `/packages/protocol-ts/src/generated/**`, `/internal/gen/**`, generated frontend ledgers under `/docs/testing/frontend_phase_coverage_ledgers/**`, generated schedules, generated Make includes, `pnpm-lock.yaml`, and `go.sum`.
- Command surface: `make frontend-typecheck`, `make frontend-unit`, `make frontend-import-boundary-check`, `make generated-artifact-policy-check`, `make generate-drift`, `make phase-ledger-drift`, and `git diff --check`.

## Evidence Layer Matrix

| Row | Evidence layer | Evidence class | Claim intent |
| --- | -------------- | -------------- | ------------ |
| `FE-U-P0-01` | Unit | Product conformance | Generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code; completion must cite confirmed Core owners and ACs |
| `FE-U-P0-02` | Unit | Product conformance | View-schema adapters key editable and queryable fields by `field_key`, not labels, indexes, or visible column order; completion is blocked until exact owner lookup is resolved |
| `FE-U-P0-03` | Unit | Product conformance | Stable selector and test-id builders derive identifiers from stable IDs and closed vocabularies rather than visible labels; completion is blocked until exact owner lookup is resolved |
| `FE-S-P0-01` | Support | Implementation support | Generated protocol policy and generated contract drift are enforced; this must not be represented as product conformance |
| `FE-S-P0-02` | Support | Implementation support | `/apps/web` consumes `/packages/grid-adapter` and does not import `react-data-grid` directly; this must not be represented as product conformance |
| `FE-S-P0-03` | Support | Implementation support | Frontend phase rows are recorded in namespace-specific registry, map, and generated ledger mechanics before any phase-enforced completion claim; this must not be represented as product conformance |

## Dependencies and Prerequisites

- Generated protocol artifacts exist under `/packages/protocol-ts/src/generated/**`; FE-P0 depends on consuming them through handwritten facades and must not hand-edit generated output.
- Frontend workspace package presence is partial: `/packages/protocol-ts`, `/packages/view-contracts`, `/packages/ui-contracts`, `/packages/grid-adapter`, `/packages/test-utils`, and `/apps/web` have manifests; `/packages/ui` needs package-surface follow-up.
- View-schema contracts are consumed through `/packages/view-contracts`; FE-P0 completion must validate stable `view_schema_id` and `field_key` behavior with unit evidence.
- Development-guide package boundaries govern `/apps/web`, `/packages/grid-adapter`, `/packages/ui-contracts`, `/packages/view-contracts`, `/packages/protocol-ts`, `/packages/test-utils`, and `/packages/ui`.
- Make target availability was verified by target discovery; passing validation is not claimed for targets that were not run.
- Phase-map and ledger machinery exists for the frontend namespace; FE-P0 remains blocked because the phase map rows are blocked.
- Existing generated-artifact policy tooling exists; FE-P0 completion must run and record `make generated-artifact-policy-check` and `make generate-drift`.

## Public Interfaces and Deliverables

FE-P0 introduces no user-visible product routes, workflows, or runtime workbook shell behavior.

Deliverables are repository artifacts and validation surfaces:

- `FRONTEND_PHASE0_IMPLEMENTATION_PLAN.md`, this planning artifact.
- `tools/frontend_phase_registry.json`, existing frontend phase registry; any future update must preserve `phase_namespace="frontend"`.
- `tools/frontend_phase_maps/fe_p0_test_map.json`, existing FE-P0 row map; TODO reconcile metadata against the controlling frontend guide before completion.
- `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md`, existing generated companion ledger; must not be hand-edited.
- Package-boundary validation through `make frontend-import-boundary-check`.
- Generated-artifact validation through `make generated-artifact-policy-check` and `make generate-drift`.
- Unit/typecheck validation through `make frontend-typecheck` and `make frontend-unit`.
- Selector/test-id contract implementation, if changed later, must stay under appropriate frontend contract packages such as `/packages/ui-contracts` and must remain label-independent.

## Sprint 0: Frontend Phase Manifest And Ownership Setup

- Objective: Ensure FE-P0 uses the separate frontend phase namespace and has one authoritative row-map inventory plus generated ledger mechanics before any completion claim.
- Status: Planned; current registry, map, and ledger exist, but all FE-P0 rows remain blocked.
- Relevant IDs: Owns `FE-S-P0-03`; supports all FE-P0 rows.
- Files and areas: `tools/frontend_phase_registry.json`, `tools/frontend_phase_maps/fe_p0_test_map.json`, `docs/testing/frontend_phase_coverage_ledgers/fe_p0_coverage_ledger.md`, phase tooling under `scripts/`, and Make phase-ledger targets.
- Test-first sequence: Run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`; run `make phase-ledger-drift`; inspect row statuses before changing claims.
- Implementation tasks: Reconcile FE-P0 row metadata against the frontend guide; preserve `PHASE=phaseN` as base-only; reject ambiguous phase identifiers; keep generated ledgers generated.
- Validation commands: `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0`; `make phase-ledger-drift`; `git diff --check`.
- Deliverables: Namespace-aware FE-P0 representation and generated ledger, or precise blockers.
- Risks and assumptions: A drift pass only proves ledger/map agreement; it does not prove row completion.
- Blockers and follow-up notes: Current rows are blocked; map metadata must be checked against frontend-guide owner ranges and support ACs before completion.

## Sprint 1: Generated Protocol Facade And Type-Consumption Baseline

- Objective: Ensure generated protocol exports and frontend contract facades expose stable identifiers without hand-editing generated code.
- Status: Not started; row remains blocked.
- Relevant IDs: Owns `FE-U-P0-01`.
- Files and areas: `/packages/protocol-ts`, `/packages/protocol-ts/src/generated/**`, `/contracts/**`, generated-artifact policy configuration, and frontend consumers.
- Test-first sequence: Run `make frontend-typecheck`; run `make frontend-unit`; inspect generated-artifact policy failures before implementation.
- Implementation tasks: Keep generated output behind handwritten facade exports; preserve generated markers; prevent direct consumer imports from generated paths except through approved facade code; cover stable identifiers in unit/typecheck evidence.
- Validation commands: `make frontend-typecheck`; `make frontend-unit`; `make generated-artifact-policy-check`; `make generate-drift`.
- Deliverables: Facade-backed protocol consumption and validation evidence, or exact blocker records.
- Risks and assumptions: Existing generated markers were observed, but policy and drift checks were not run for this plan.
- Blockers and follow-up notes: Row remains blocked until direct unit/typecheck evidence and generated-artifact validation are recorded.

## Sprint 2: View-Schema And `field_key` Adapter Contract

- Objective: Ensure view-schema adapters key editable and queryable fields by `field_key`, not labels, indexes, or visible column order.
- Status: Not started; row remains blocked.
- Relevant IDs: Owns `FE-U-P0-02`.
- Files and areas: `/packages/view-contracts`, generated contract inputs, view-schema fixtures, and frontend tests that exercise adapter behavior.
- Test-first sequence: Add or identify failing unit coverage for label/order-independent field addressing; run `make frontend-unit`.
- Implementation tasks: Validate adapter lookup by `view_schema_id`; ensure editable/queryable field surfaces use `field_key`; reject tests that assert visible labels or index order as identity.
- Validation commands: `make frontend-unit`.
- Deliverables: Unit evidence for `field_key` identity behavior, or exact blocker records.
- Risks and assumptions: `field_key` is a domain vocabulary term and Core-owned concept; visible labels and column order must remain non-authoritative.
- Blockers and follow-up notes: `TODO: owner lookup required` for exact Core 03 Section 4.8 heading before product-conformance completion.

## Sprint 3: Stable Selector And Test-ID Builders

- Objective: Ensure selector and test-id builders derive identifiers from stable IDs and closed vocabularies rather than visible labels.
- Status: Not started; row remains blocked.
- Relevant IDs: Owns `FE-U-P0-03`.
- Files and areas: `/packages/ui-contracts`, `/packages/test-utils`, frontend tests, and any consumer selectors in `/apps/web`.
- Test-first sequence: Add or identify unit coverage proving stable IDs and closed vocabularies drive selectors; run `make frontend-unit`.
- Implementation tasks: Keep selector/test-id construction deterministic; prohibit visible-label-derived selectors; use stable identifiers such as `record_id`, `field_key`, `view_schema_id`, and closed vocabularies where applicable.
- Validation commands: `make frontend-unit`.
- Deliverables: Selector/test-id contract evidence, or exact blocker records.
- Risks and assumptions: Existing selector helpers were observed, but FE-P0 row completion needs current unit evidence.
- Blockers and follow-up notes: `TODO: owner lookup required` for exact Core 02 Section 5.3 heading before product-conformance completion.

## Sprint 4: Generated-Artifact Policy, Drift, And Codegen Hygiene

- Objective: Enforce generated protocol policy and generated contract drift without treating support evidence as product conformance.
- Status: Not started; row remains blocked.
- Relevant IDs: Owns `FE-S-P0-01`.
- Files and areas: `tools/generated_artifact_policy.json`, `/packages/protocol-ts/src/generated/**`, `/contracts/**`, generator scripts, generated Make includes, and lockfiles.
- Test-first sequence: Run `make generated-artifact-policy-check`; run `make generate-drift`; inspect exact failing unit if either fails.
- Implementation tasks: Preserve generated markers; keep generated roots excluded from authored-source lint scopes where policy requires; separate codegen drift from migration drift; never hand-edit generated protocol output.
- Validation commands: `make generated-artifact-policy-check`; `make generate-drift`; `git diff --check`.
- Deliverables: Passing generated policy and drift checks, or exact blocker records.
- Risks and assumptions: This is implementation-support evidence only.
- Blockers and follow-up notes: Validation targets exist but were not run for this planning artifact.

## Sprint 5: RDG Import Boundary And Stylesheet Ownership

- Objective: Ensure `/apps/web` consumes `/packages/grid-adapter` and does not import `react-data-grid` directly.
- Status: Not started; row remains blocked.
- Relevant IDs: Owns `FE-S-P0-02`.
- Files and areas: `/apps/web`, `/packages/grid-adapter`, `tools/frontend_import_boundaries.json`, frontend package manifests, and stylesheet ownership points.
- Test-first sequence: Run `make frontend-import-boundary-check`; inspect import-boundary diagnostics before implementation.
- Implementation tasks: Keep all RDG imports inside `/packages/grid-adapter`; keep `/apps/web` on adapter exports; ensure RDG stylesheet ownership remains in the adapter boundary; document `/packages/ui` package-surface TODO if no manifest is added in FE-P0.
- Validation commands: `make frontend-import-boundary-check`.
- Deliverables: Passing import-boundary evidence, or exact blocker records.
- Risks and assumptions: Source search found no direct `/apps/web` RDG import, but the enforcement target must pass before completion.
- Blockers and follow-up notes: `/packages/ui` exists without a verified package manifest.

## Sprint 6: Frontend Namespace, Ledger Drift, And Phase-Gate Handoff

- Objective: Collect FE-P0 evidence and handoff state for FE-P1 without duplicating row ownership or turning support-only evidence into product conformance.
- Status: Planned; no primary row ownership.
- Relevant IDs: Supports `FE-U-P0-01`, `FE-U-P0-02`, `FE-U-P0-03`, `FE-S-P0-01`, `FE-S-P0-02`, and `FE-S-P0-03`.
- Files and areas: FE-P0 plan, frontend phase registry/map/ledger, validation artifacts, and FE-P1 handoff notes.
- Test-first sequence: Re-run row-owned validation commands; run `make phase-ledger-drift`; run `git diff --check`.
- Implementation tasks: Verify every FE-P0 row is either complete with evidence or blocked with precise ownership; ensure `PHASE_NAMESPACE=frontend PHASE=FE-P0` is the only frontend phase-selection spelling; keep base `PHASE=phaseN` untouched.
- Validation commands: Row-owned commands plus `make phase-ledger-drift` and `git diff --check`.
- Deliverables: FE-P1 handoff state with row statuses, blockers, and validation artifact locations.
- Risks and assumptions: Active frontend phase execution is not implemented for planned phases; completion must not depend on planned-phase execution.
- Blockers and follow-up notes: Handoff remains blocked until all row-owner sprint blockers are resolved or explicitly recorded.

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
| FE-P0 is represented in the frontend phase namespace or a precise blocker prevents that representation | Present: registry, map, and ledger exist; rows remain blocked |
| `tools/frontend_phase_maps/fe_p0_test_map.json` or equivalent planned manifest covers every FE-P0 row exactly once | Present by local inspection; metadata reconciliation remains TODO |
| Generated frontend ledger behavior is present or explicitly blocked; any generated ledger is produced by a generator and not hand-edited | Present; `make phase-ledger-drift` passed during planning inspection |
| `FE-U-P0-01`, `FE-U-P0-02`, and `FE-U-P0-03` have direct unit/typecheck evidence or explicit blockers | Blocked; unit/typecheck commands were not run for this plan |
| `FE-S-P0-01`, `FE-S-P0-02`, and `FE-S-P0-03` have implementation-support validation or explicit blockers | Blocked; `FE-S-P0-03` has drift evidence but row remains blocked, other support targets were not run |
| `/apps/web` direct `react-data-grid` imports are prohibited or the absence of an enforcement target is recorded as a blocker | Target exists; completion requires `make frontend-import-boundary-check` |
| Generated protocol artifacts are not hand-edited | Generated markers observed; completion requires policy and drift checks |
| Generated-artifact policy and generate-drift checks pass or have precise blockers | Blocked; targets exist but were not run for this plan |
| Frontend phase selection uses `PHASE_NAMESPACE=frontend PHASE=FE-P0` or the absence of namespace-aware selection is recorded as a blocker | Present and verified by `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P0` |
| No support-only, visual, accessibility, or measurement evidence is represented as product-conformance evidence | Required; current map distinguishes product conformance from implementation support |
| No Core 05 claim-publication review is activated unless a claim-bearing publication intent is explicitly declared | Required; FE-P0 has no claim-bearing publication intent |
| `git diff --check` passes | Must be run after plan edits |

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
