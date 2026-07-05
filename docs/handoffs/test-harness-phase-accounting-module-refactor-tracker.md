# test-harness-phase-accounting Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `tools/harness/phase-accounting`
- Target label: `test-harness-phase-accounting`
- Output path: `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md`
- Status: tracker created in a planning-only session, then advanced by an authorized harness remediation implementation session on 2026-07-05.
- Allowed changes in the remediation session: harness specification text, harness helper implementation, harness smoke/contract tests, and generated topology metadata refreshed through Make-owned generation.
- Non-goals: no product route, workbook, saved-view, projection, authorization, storage, revision, generated product contract, package/dependency, or migration change.
- Implementation posture: additional product behavior or public harness contract changes require a later owner-backed authorization task.

Source hierarchy used:

1. Adopted subsystem NLSpecs for their named subsystem only. `docs/testing-harness-nlspec.md` owns the relevant harness mechanics here.
2. Core 00 through Core 04 for product implementation-conformance behavior.
3. Core 05 only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
4. `docs/domain.md` and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and framework files as evidence only.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `AGENTS.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`

Repository files inspected:

- Every tracked file under `tools/harness/phase-accounting`.
- `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`, `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`, and `tools/execution_topology_render_index.json` references.
- `tools/harness/static-analysis/harness-import-boundary.mjs`.
- `tools/harness/tests/test-harness-contracts.mjs`.
- Selected adjacent harness callers in backend, browser, diagnostics, execution, generated-artifacts, output/test-output, and release-evidence helpers.

Architectural finding:

`tools/harness/phase-accounting` is a harness evidence/accounting helper area, not a product workbook orchestration module. It currently owns phase manifests, phase registries, evidence naming, row accounting, fixture policy, frontend retained-evidence audit, and phase-slice planning. Live inspection found no current ownership of product HTTP routes, WebSocket behavior, workbook row mutations, saved views, view-schema runtime behavior, projection refresh, authorization, revision/change-set semantics, storage semantics, or generated product contracts.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tools/harness/phase-accounting/check-phase-maps.sh` | Runs phase registry, frontend phase artifacts, phase-policy exceptions, and per-phase map checks. | Executable shell wrapper used by Make target `phase-map-check`. | `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`. | Node, `phase-registry.mjs`, `frontend-phase-manifest.mjs`, `phase-manifest.mjs`, `phase-map-check-cli.mjs`. | Indirect via harness smoke and task-surface/topology tests. | `phase-map-check` command metadata and phase-map validation. | Phase-accounting command wrapper. | medium | Harness mechanics only; not a product route or module. |
| `tools/harness/phase-accounting/frontend-evidence-audit-cli.mjs` | Audits retained frontend row-accounting roots and writes `frontend-evidence-audit-summary.json`. | CLI for `frontend-evidence-audit`; no module exports. | Make/task-surface/topology command bindings. | `frontend-phase-manifest.mjs`, contract schema validation, filesystem/hash helpers. | `tests/test-frontend-evidence-audit.sh`; task-surface/topology smoke. | `cartulary.frontend_evidence_audit_summary.v1`, `cartulary.frontend_row_accounting.v3`, retained root inputs. | `frontend_evidence_audit`. | high | Requires retained `CHECK_RESULTS_DIR`, support, visual, and a11y roots for full direct validation. |
| `tools/harness/phase-accounting/frontend-phase-manifest.mjs` | Validates frontend phase registry, phase maps, visual fixture registry, guide restatements, freshness digests, ledgers, and Playwright grep helpers. | Exports FE namespace/schema IDs, registry/map loaders, validators, ledger renderer, scenario-title and grep helpers. | Browser a11y/stateful/visual helpers, frontend-unit runner, generated-artifact checks/renderers, output indexes, release evidence, harness-contract tests, task-surface/topology metadata. | Node crypto/fs/path, `contract/json-shape.mjs`, Core/doc path reads, frontend phase maps/registry/guide/visual fixtures. | `tools/harness/generated-artifacts/tests/test-json-shapes.sh`, `tools/harness/tests/test-harness-contracts.mjs`, phase-accounting smoke. | `cartulary.frontend_phase_registry.v2`, `cartulary.frontend_phase_test_map.v3`, `cartulary.frontend_visual_fixture_registry.v2`, frontend ledgers. | `frontend_phase_accounting`. | high | Current owner facade candidate for frontend phase-map behavior. |
| `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs` | Builds frontend namespace phase-slice work-unit plans and selected-row accounting scopes. | Exports `FrontendPhaseNotExecutableError`, `buildFrontendPhaseSlicePlan`, `printableFrontendPlan`, schema alias and repo root. | `phase-slice-cli.mjs` through `frontend-readiness.mjs`; task-surface/topology metadata; scheduler handoffs. | Frontend phase manifest, task-surface loader, scheduler resource policy, phase-slice plan contract. | `tests/test-run-phase-slice.sh`; harness-contract/import-boundary fixtures. | `cartulary.phase_slice_plan.v1`; `phase-slice` and `service-backed-slice` plan output. | `frontend_phase_accounting` plus `phase_slice_planning`. | high | Mixed frontend accounting and scheduler plan construction; keep behavior frozen before splitting. |
| `tools/harness/phase-accounting/frontend-readiness.mjs` | Re-export facade for frontend phase-slice planning. | `export *` from `frontend-phase-slice-plan.mjs`. | Harness import-boundary allowlist; `phase-slice-cli.mjs`. | `frontend-phase-slice-plan.mjs`. | Import-boundary contract tests. | None directly beyond facade use. | `frontend_phase_accounting` facade. | low | Facade should stay stable if internals move. |
| `tools/harness/phase-accounting/frontend-row-accounting.mjs` | Builds `frontend-row-accounting.json` from target artifacts and target status. | Exports row-accounting scope modes, default/normalization helpers, accounting builder, failure helpers. | `tools/harness/output/test-output/target-summary.mjs`, task-surface/topology metadata, browser tracker evidence. | Frontend phase manifest, `tools/harness/output/test-output/frontend-row-evidence.mjs`, `contract/test-output-context.mjs`, filesystem/hash helpers. | Browser harness smoke and row-accounting schema tests; indirect through FE audit smoke. | `cartulary.frontend_row_accounting.v3`, retained `<target>/frontend-row-accounting.json`. | `frontend_row_accounting`. | high | Test-output artifact parsing now sits behind the test-output helper boundary; row closure semantics remain here. |
| `tools/harness/phase-accounting/frontend/index.mjs` | Frontend phase-accounting facade. | Re-exports frontend phase manifest and frontend row accounting. | Harness import-boundary allowlist; diagnostics/browser/generated-artifact callers migrated to direct owner modules where appropriate. | `frontend-phase-manifest.mjs`, `frontend-row-accounting.mjs`. | Import-boundary contract tests. | Facade over FE registry/map and row-accounting contracts. | Frontend owner facade. | medium | Test-output index re-export was removed in the remediation session. |
| `tools/harness/phase-accounting/phase-entry-evidence.mjs` | Extracts Go symbols, Vitest titles, Playwright titles, evidence names, claim status, and executable status from phase entries. | Exports symbol/title/evidence and collection helpers. | `phase-manifest.mjs`, `phase-fixture-policy.mjs`, `phase-selection.mjs`, `phase-manifest-validation.mjs`. | `phase-manifest-constants.mjs`. | Covered through `phase-map-check`, `phase-test-name-check`, phase-slice smoke. | Phase-map row/evidence naming contract. | Base phase-accounting. | medium | Pure harness evidence parsing. |
| `tools/harness/phase-accounting/phase-fixture-policy.mjs` | Validates and derives Postgres fixture policy, budgets, reset assignments, and clone/reason rules. | Exports policy constants and fixture policy/budget validators. | `phase-manifest.mjs`, `phase-manifest-validation.mjs`, backend target/shard planning through facade. | `phase-manifest-shape.mjs`, `phase-entry-evidence.mjs`. | `tests/test-print-target-plan.sh`, `phase-map-check`, shard-plan smoke. | Phase-map fixture policy fields; target/shard plan DTOs. | Base phase-accounting. | medium | Harness fixture lifecycle mechanics, not storage/domain ownership. |
| `tools/harness/phase-accounting/phase-frontend-fixtures.mjs` | Validates base visual fixture refs, frontend visual fixture refs, and live grid-adapter usage for authoritative grid rows. | Exports fixture-ref validators and grid-adapter guard. | `phase-manifest-validation.mjs`. | `frontend-phase-manifest.mjs`, `contract/json-shape.mjs`, test source files. | `phase-map-check`, visual fixture metadata tests. | FE visual fixture registry refs. | Frontend phase-accounting. | medium | Reads frontend test source only for harness guardrails. |
| `tools/harness/phase-accounting/phase-manifest-constants.mjs` | Defines shared sections and closed constants for base phase maps. | Exports section definitions, coverage/Go section/claim/runtime constants, support target section map. | Phase evidence, selection, validation, policy exceptions, manifest shape. | None. | Covered indirectly by all phase-map and phase-slice tests. | Base phase-map closed values. | Base phase-accounting. | low | Low-risk constant module. |
| `tools/harness/phase-accounting/phase-manifest-loader.mjs` | Loads registry-backed phase manifests and validates filename/identity/shape. | Exports `phaseNumberFromPhase`, `loadManifest`, `phaseManifestNames`. | `phase-manifest.mjs`, validation, selection, policy exceptions, diagnostics/generated-artifact callers through facade. | `contract/json-shape.mjs`, `phase-registry.mjs`, `phase-manifest-shape.mjs`. | `tests/test-print-target-plan.sh`, `phase-map-check`. | `cartulary.phase_registry.v1`, `cartulary.phase_test_map.v2`. | Base phase-accounting. | medium | Registry status controls active/planned/retired behavior. |
| `tools/harness/phase-accounting/phase-manifest-shape.mjs` | Validates raw base phase-map JSON shape and closed values. | Exports schema ID, key sets, closed value sets, and shape validators. | `phase-manifest-loader.mjs`, `phase-manifest-validation.mjs`, JSON-shape checks. | `contract/json-shape.mjs`. | `tools/harness/generated-artifacts/tests/test-json-shapes.sh`, phase-map smoke. | `cartulary.phase_test_map.v2`. | Base phase-accounting. | high | Shape drift can invalidate phase-map and JSON-shape gates. |
| `tools/harness/phase-accounting/phase-manifest-validation.mjs` | Performs semantic phase-map validation against source tests, guide expected IDs, fixtures, profile claims, and support targets. | Exports `validateManifest`. | `phase-manifest.mjs`, `phase-map-check-cli.mjs`, generated ledger drift. | Execution dependency catalog, JSON shape helpers, entry evidence, fixture policy, frontend fixtures, loader. | `tests/test-print-target-plan.sh`, `tests/test-check-phase-test-names.sh`, `phase-map-check`. | Phase ledgers/maps, profile claims, fixture policy, row evidence closure. | Base phase-accounting. | high | Central semantic gate; characterize before edits. |
| `tools/harness/phase-accounting/phase-manifest.mjs` | Public phase-manifest facade and CLI for phase listing, regex/count selection, run verification, and policy exception validation. | Re-exports many phase accounting helpers; CLI commands include `list-phases`, `go-regex`, `playwright-grep`, `vitest-verify-run`, and related selectors. | Backend target plan/run scripts, browser run scripts, diagnostics, generated-artifacts, output adapters, harness-contract tests, task-surface/topology metadata. | Phase evidence, fixture policy, loader, validation, policy exceptions, selection, run verification. | Browser/backend/generated-artifact/harness-contract tests; phase-accounting smoke. | `cartulary.playwright_manifest_selection.v1`, phase-map row selection behavior. | Base phase-accounting facade. | high | Current broad facade; avoid fragmenting without import-boundary plan. |
| `tools/harness/phase-accounting/phase-map-check-cli.mjs` | Validates one phase manifest with planned phases allowed. | CLI only. | `check-phase-maps.sh`, backend run-go-target tests. | `phase-manifest.mjs`. | `tools/harness/backend/tests/test-run-go-target.sh`, `tests/test-print-target-plan.sh`. | Phase-map validation result. | Phase-accounting command. | low | Thin wrapper. |
| `tools/harness/phase-accounting/phase-policy-exceptions.mjs` | Loads and validates temporary phase policy exceptions and allowed empty Go selections. | Exports `loadPhasePolicyExceptions`, `emptyGoManifestSelectionAllowed`. | `phase-manifest.mjs`, phase selection callers. | Execution dependency catalog, JSON shape helpers, phase registry/loader/constants. | `phase-map-check`, `test-print-target-plan.sh`. | `tools/phase_policy_exceptions.json`, task-surface/topology dependency metadata. | Base phase-accounting. | medium | Exception expiry is harness policy, not product behavior. |
| `tools/harness/phase-accounting/phase-registry.mjs` | Loads and validates base phase registry, active/planned/retired status, manifest and ledger paths. | Exports registry schema/status constants, loaders, filters, validators, CLI `validate/list/list-active`. | `check-phase-maps.sh`, diagnostics, generated-artifacts, phase-slice planning, task-surface/topology metadata. | JSON shape helpers, filesystem path validation. | `tests/test-print-target-plan.sh`, harness-contract tests. | `cartulary.phase_registry.v1`, phase ledger/map registry. | Base phase-accounting. | medium | Registry status is execution eligibility only. |
| `tools/harness/phase-accounting/phase-run-verification.mjs` | Verifies selected Go, Vitest, and Playwright runs against expected manifest symbols/titles. | Exports Go log, Vitest report, Playwright report verification helpers. | `phase-manifest.mjs`, browser duration/shard planning by facade. | `output/test-output/playwright-artifacts.mjs`, JSON reports/logs. | Browser run smoke, Vitest/browser phase wrappers. | Playwright/Vitest/Go retained run evidence. | Phase-accounting plus test-output boundary. | high | Coupled to runner report shapes; preserve artifact interpretation. |
| `tools/harness/phase-accounting/phase-selection.mjs` | Selects base and frontend Go/Vitest/Playwright entries from phase manifests and FE maps. | Exports package matching, row selection, path normalization, and phase lists. | `phase-manifest.mjs`, browser/backend runner wrappers. | Execution dependency catalog, FE facade, constants, entry evidence, manifest loader, filesystem source scans. | Browser/backend/execution smoke via manifest commands. | Selected test lists and phase-scoped execution. | Phase-accounting selection. | high | Bridges base phase maps and FE row maps for Playwright. |
| `tools/harness/phase-accounting/phase-slice-cli.mjs` | CLI for `phase-slice` and `service-backed-slice`; parses user inputs, selects base/FE planners, prints JSON plans, and delegates runtime execution. | CLI only. | `tools/harness/execution/make-node-tools.mjs`, task-surface/topology metadata. | Base/FE phase-slice planners and `tools/harness/scheduler/phase-slice-execution.mjs`. | `tests/test-run-phase-slice.sh`, harness-contract tests, Make node tools smoke. | `cartulary.phase_slice_scheduler_summary.v4`, `cartulary.scheduler_event.v6`, target/tool summaries. | `phase_slice_planning` plus scheduler execution adapter. | high | Scheduler setup/reexec/runtime attachment moved to the scheduler-owned adapter while public CLI behavior remains frozen. |
| `tools/harness/phase-accounting/phase-slice-plan-contract.mjs` | Serializes and semantically validates stable phase-slice plan fields. | Exports `phaseSlicePlanSchemaID`, `resourceLimitObject`, `serializePhaseSliceWorkUnit`, `validatePhaseSlicePlanContract`. | Base and frontend phase-slice planners, phase-slice smoke tests. | Scheduler resource registry helpers. | `tests/test-run-phase-slice.sh`, JSON-shape checks. | `cartulary.phase_slice_plan.v1`. | `phase_slice_planning` contract. | high | Contract guard for stable emitted JSON plan fields. |
| `tools/harness/phase-accounting/phase-slice-plan.mjs` | Builds base phase-slice plans and selected work-unit accounting. | Exports `buildPhaseSlicePlan`, `validateAllPhaseSlicePlans`, `printablePlan`, schema ID, repo root. | `phase-slice-cli.mjs`, generated topology renderer, import-boundary checks, harness-contract tests, task-surface/topology metadata. | Browser scheduler adapter, execution dependencies, execution topology, phase manifest/registry, backend shard/target plan, scheduler resource policy, diagnostics guidance, plan contract. | `tests/test-run-phase-slice.sh`, harness-contract/import-boundary tests. | `cartulary.phase_slice_plan.v1`, task-surface/topology plan dependencies. | `phase_slice_planning`. | high | Declared owner facade in harness NLSpec; key boundary file. |
| `tools/harness/phase-accounting/phase-test-name-check-cli.mjs` | Enforces phase test-name evidence fragments and authoritative evidence names. | CLI only. | Make/task-surface/topology command bindings. | Filesystem scans of Go tests, `phase-manifest.mjs`. | `tests/test-check-phase-test-names.sh`, harness smoke. | `phase-test-name-check` check-internal target. | Phase-accounting command. | medium | Guards naming contract between Go tests and phase maps. |
| `tools/harness/phase-accounting/tests/test-check-phase-test-names.sh` | Shell smoke coverage for phase test-name checker. | Executable test script. | Harness smoke/task-surface/topology metadata. | Synthetic phase registry/maps, `phase-test-name-check-cli.mjs`, temp scratch. | It is direct coverage. | Harness smoke row for phase-test-name behavior. | Harness smoke. | medium | Test-only fixture generation under scratch/tmp. |
| `tools/harness/phase-accounting/tests/test-frontend-evidence-audit.sh` | Shell smoke coverage for FE retained evidence audit pass, missing root, and stale digest failures. | Executable test script. | Harness smoke/task-surface/topology metadata. | Live FE registry/map, synthetic retained roots, `frontend-evidence-audit-cli.mjs`. | It is direct coverage. | `cartulary.frontend_evidence_audit_summary.v1`, `cartulary.frontend_row_accounting.v3`. | Harness smoke. | medium | Does not replace direct audit over real retained roots. |
| `tools/harness/phase-accounting/tests/test-print-target-plan.sh` | Smoke coverage for target/shard plan determinism, phase registry/map identity, fixture policy validation, ledger rendering, and diagnostics. | Executable test script. | Harness smoke/task-surface/topology metadata. | Diagnostics target-plan, backend shard plan, phase registry/maps, phase-map check, generated ledgers. | It is direct coverage. | Target-plan JSON, phase registry/map validation, phase ledger drift. | Harness smoke plus diagnostics support. | high | Mixed ownership test; consider eventual relocation or split if diagnostics ownership is refined. |
| `tools/harness/phase-accounting/tests/test-run-phase-slice.sh` | Smoke coverage for base and frontend phase-slice plans, service-backed mode, wrapper reexec, scheduler failure summaries, and finalizer log naming. | Executable test script. | Harness smoke/task-surface/topology metadata. | `phase-slice-cli.mjs`, scheduler runner, plan contract, backend target plan, synthetic phase maps. | It is direct coverage. | `cartulary.phase_slice_plan.v1`, `cartulary.phase_slice_scheduler_summary.v4`, scheduler events/logs. | Harness smoke. | high | Primary characterization target before phase-slice refactors. |

## 3. Module Boundary Diagnosis

Current target classification:

- legitimate harness owner facade for phase-accounting behavior;
- mixed-responsibility package where phase accounting, frontend row accounting, retained evidence audit, test-output indexing, and phase-slice scheduler orchestration meet;
- view/projection orchestration layer: no, not for product projections; only harness evidence projections and row-accounting artifacts;
- transport-adjacent adapter: no product transport; Make/CLI command wrapper only;
- persistence-adjacent adapter: no product persistence; fixture policy only;
- mutation coordinator: no product mutation; scheduler and retained-artifact summary coordination only;
- frontend shell/controller surface: no runtime app shell ownership; frontend readiness evidence only;
- grid-vendor integration layer: no vendor integration; one guard verifies grid-adapter test usage;
- misplaced home for product logic: no current evidence;
- mixed-responsibility package: yes, for harness helper boundaries.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Base phase manifest shape, registry semantics, evidence naming, fixture policy, row selection, and run verification | `phase-*.mjs`, `check-phase-maps.sh` | Base phase-accounting helpers | keep | Testing Harness NLSpec Section 4.1 says these are phase-accounting behavior. | Current path is a valid harness helper area. |
| Frontend phase registry/maps, FE row IDs, fixture references, guide restatements, grep helpers, and ledger rendering | `frontend-phase-manifest.mjs`, `frontend/index.mjs` | `frontend_phase_accounting` | keep | Harness helper ownership registry lists this facade key. | Preserve FE schema IDs and freshness behavior; direct manifest callers may import the owner module. |
| Frontend row-accounting artifact generation and closure failure injection | `frontend-row-accounting.mjs` | `frontend_row_accounting` | keep | Harness NLSpec names `cartulary.frontend_row_accounting.v3` and retained artifact path. | Coupled to target-summary/test-output; characterize before moving. |
| Frontend retained-evidence audit | `frontend-evidence-audit-cli.mjs`, `tests/test-frontend-evidence-audit.sh` | `frontend_evidence_audit` | keep | Harness NLSpec names retained-root inputs and audit summary schema. | Direct validation needs retained roots. |
| Phase-slice plan construction and selected work-unit accounting | `phase-slice-plan.mjs`, `frontend-phase-slice-plan.mjs`, `phase-slice-plan-contract.mjs` | `phase_slice_planning` | keep | Harness NLSpec declares current facade `phase-slice-plan.mjs`. | Public JSON plan contract must stay stable. |
| Phase-slice CLI setup, service-wrapper reexec, scheduler runtime attachment, and summary emission | `phase-slice-cli.mjs`, `tools/harness/scheduler/phase-slice-execution.mjs` | Split between `phase_slice_planning` and `scheduler_execution_core` | split | Remediation moved runtime execution into a scheduler-owned adapter while the CLI still composes planning and runtime. | Future work can further narrow the CLI/env resolver, but the largest scheduler coupling has moved. |
| Test-output indexing used by FE row accounting | `tools/harness/output/test-output/frontend-row-evidence.mjs`, `frontend-row-accounting.mjs` | Test-output helpers plus FE row-accounting facade | split | Remediation moved Vitest/Playwright title observation parsing behind test-output ownership. | Retained `frontend-row-accounting.json` schema and path remain stable. |
| Target/shard plan diagnostics smoke | `tests/test-print-target-plan.sh` | Diagnostics/backend harness tests | defer | Test drives `target-plan`, backend shard plan, registry validation, ledgers. | Mixed smoke may remain until diagnostics test ownership is revisited. |
| Product workbook orchestration | None found in target path | Core 01/Core 03/product modules if ever touched | intentional/no_action | No HTTP handlers, WebSocket hubs, workbook mutation packages, storage adapters, or generated product contracts were found under the target path. | Do not infer product ownership from phase names or FE row labels. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `phase-map-check` Make target | Testing Harness NLSpec and task-surface metadata | Make/task-surface/topology invoke `check-phase-maps.sh`; NLSpec owns Make command surface. | Phase-map smoke, target-plan smoke, JSON-shape checks. | Run `make phase-map-check` before moving phase-map validators. | high | Command identity and output/failure mapping are frozen. |
| `phase-test-name-check` Make target | Testing Harness NLSpec and task-surface metadata | Task-surface/topology bind `phase-test-name-check-cli.mjs`. | `tests/test-check-phase-test-names.sh`. | Run `make phase-test-name-check` before naming-check edits. | medium | Check-internal target participates in default check. |
| `frontend-evidence-audit` Make target | Testing Harness NLSpec | NLSpec row for `cartulary.harness.command.frontend_evidence_audit.v1`; CLI writes audit summary. | `tests/test-frontend-evidence-audit.sh`. | Direct audit with retained roots when available; otherwise retain TODO. | high | Broad `check` root is not proof of support/visual/a11y roots. |
| `phase-slice` and `service-backed-slice` Make targets | Testing Harness NLSpec and scheduler contracts | NLSpec command rows; CLI builds plans and scheduler summaries. | `tests/test-run-phase-slice.sh`, tool-output real target smoke. | Run phase-slice plan JSON and scheduler smoke before CLI/planner changes. | high | Public target names, command IDs, summaries, service wrapper behavior frozen. |
| `cartulary.phase_registry.v1` | Phase registry helper under harness NLSpec | `phase-registry.mjs`; schema owner in harness docs. | `test-print-target-plan.sh`, JSON-shape checks. | Validate active/planned/retired fixtures after registry changes. | medium | Registry status is execution eligibility, not product phase behavior. |
| `cartulary.phase_test_map.v2` | Phase manifest shape/validation | `phase-manifest-shape.mjs`, `phase-manifest-validation.mjs`. | `phase-map-check`, JSON-shape fixtures. | Run JSON-shape and phase-map check before schema/shape changes. | high | Phase maps are evidence accounting only. |
| `cartulary.phase_slice_plan.v1` | Phase-slice plan contract | `phase-slice-plan-contract.mjs`, NLSpec schema attachment table. | `test-run-phase-slice.sh`, JSON-shape checks. | Plan JSON fixtures for base and frontend slices. | high | Stable emitted fields need semantic validation. |
| `cartulary.frontend_phase_registry.v2` | Frontend phase accounting | `frontend-phase-manifest.mjs`, schema file, NLSpec. | JSON-shape, FE phase validation. | Validate registry freshness and row rollups. | high | Status and row rollup state are distinct. |
| `cartulary.frontend_phase_test_map.v3` | Frontend phase accounting | `frontend-phase-manifest.mjs`, frontend maps, NLSpec. | JSON-shape, FE ledger generation, FE audit smoke. | Validate implemented/stale/blocked row behavior. | high | FE rows do not create product behavior. |
| `cartulary.frontend_row_accounting.v3` | Frontend row accounting | `frontend-row-accounting.mjs`, schema, NLSpec retained artifact table. | FE audit smoke, browser target-summary tests. | Target-summary closure tests for active, selected, disabled scopes. | high | Artifact path is `<target>/frontend-row-accounting.json`. |
| `cartulary.frontend_evidence_audit_summary.v1` | Frontend retained evidence audit | `frontend-evidence-audit-cli.mjs`, schema, NLSpec attachment table. | `test-frontend-evidence-audit.sh`. | Real retained-root audit when roots are available. | high | Must verify digests before accepting row closure. |
| Retained scheduler summaries and logs | Scheduler execution core | `phase-slice-cli.mjs` attaches scheduler runtime and summary schema IDs. | `test-run-phase-slice.sh`. | Run scheduler smoke before CLI/scheduler interaction changes. | high | Summary/event schema ownership remains scheduler. |
| Generated task-surface/topology metadata | Generated-artifact surface | `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`. | Generated-artifact and harness-contract tests. | Run phase-schedules/drift only if owner inputs change. | high | Do not hand-edit generated outputs. |
| HTTP route shapes | Core 01/Core 04 if touched | No HTTP route handlers found under target path. | Not applicable. | None for tracker; required only if later product code changes. | low | Not owned by this target. |
| WebSocket paths and events | Core 01/Core 03/Core 04 if touched | No WebSocket handlers found under target path. | Not applicable. | None for tracker; required only if later product code changes. | low | Not owned by this target. |
| Workbook row/query/mutation behavior | Core 01/Core 03/product modules | Target handles harness row accounting only, not workbook rows. | Product tests elsewhere. | None for tracker; required for product changes. | low | Do not infer workbook behavior from FE row IDs. |
| Saved-view or view-schema behavior | Core 01/Core 03/product modules | Target validates FE phase maps, not view-schema runtime. | Product/generated tests elsewhere. | None for tracker; required for product changes. | low | `view_schema_id` product contracts remain outside this path. |
| Projection refresh behavior | Core 01/projections owner | No projection refresh code found under target path. | Product tests elsewhere. | None for tracker. | low | Phase ledgers are not runtime projections. |
| Authorization checks | Core 04/product modules | No auth checks or product session handling under target path. | Product/security tests elsewhere. | None for tracker. | low | Make wrapper env sanitization is harness, not authorization. |
| Revision/change-set behavior | Core 02/Core 03/product modules | No revision/change-set code found under target path. | Product tests elsewhere. | None for tracker. | low | Not owned by target. |
| Grid-adapter or UI selector contracts | `/packages/grid-adapter`, `/packages/ui-contracts`, frontend tests | `phase-frontend-fixtures.mjs` only guards that authoritative grid rows use live adapter path. | Phase-map validation. | Re-run phase-map check if guard changes. | medium | Guard is test policy, not grid-vendor integration ownership. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Target path is a valid harness phase-accounting area, not a permanent product module boundary. | Harness NLSpec Section 4.1 says helper paths are implementation details and phase-accounting owns evidence mechanics. | medium | intentional/no_action | Testing harness phase-accounting | Record product non-ownership and preserve harness contracts. |
| `phase-slice-cli.mjs` mixed planning, setup, service-wrapper reexec, scheduler runtime attachment, and summary emission. | Remediation moved scheduler setup, wrapper reexec, runtime attachment, and scheduler execution into `tools/harness/scheduler/phase-slice-execution.mjs`. | medium | should_fix | Split between `phase_slice_planning` and `scheduler_execution_core`. | Keep future CLI changes limited to argument/env resolution and planner selection. |
| `phase-slice-plan.mjs` is a declared owner facade but imports backend, browser, diagnostics, generated-artifact, scheduler resource, and topology helpers. | Live imports include backend shard/target plan, browser adapter, execution topology, task guidance, scheduler resource policy. | high | defer | `phase_slice_planning` facade with owner facades for dependencies. | Keep facade stable; avoid moving internals until import-boundary tests are planned. |
| `frontend-row-accounting.mjs` coupled FE row closure to test-output artifact parsing. | Remediation introduced `tools/harness/output/test-output/frontend-row-evidence.mjs`; row accounting now consumes normalized title observations. | medium | should_fix | `frontend_row_accounting` plus `test_output_frontend_indexing`. | Preserve fixture output and retained artifact schema while further hardening DTO tests. |
| `frontend/index.mjs` re-exported test-output frontend indexes. | Remediation removed the test-output re-export and migrated direct manifest callers. | low | intentional/no_action | FE facade and test-output facade. | Keep contract test coverage that prevents the re-export from returning. |
| `tests/test-print-target-plan.sh` is located under phase-accounting but covers diagnostics/backend target and shard planning. | Test invokes target-plan and shard-plan scripts, registry/map fixtures, ledger generation. | medium | should_fix | Diagnostics/backend harness tests plus phase-accounting tests. | Split only if smoke tier and topology owner inputs are updated. |
| Generated task-surface/topology manifests reference phase-accounting helper paths. | Live refs in `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`. | high | must_fix | `generated_artifact_surface` | Any path move must update owner inputs and regenerate through Make; never hand-edit generated outputs. |
| Visual/accessibility FE evidence could be misread as product conformance. | Harness NLSpec says visual/accessibility rows are design/readiness unless explicit Core 05 claim boundary. | medium | must_fix | Testing harness and Core 05 only when claimed. | Tracker must preserve evidence class boundaries. |
| Product HTTP, WebSocket, workbook, saved-view, projection, storage, revision, and authorization code are absent from target path. | Live inventory found only harness scripts/modules and tests. | low | intentional/no_action | Product owners remain Core 01 through Core 04 modules. | Do not propose product implementation patches from this tracker. |
| Direct SQL/storage coupling is absent. | No SQL queries, storage adapters, migrations, or DB packages found under target. | low | intentional/no_action | Not applicable. | Keep out of scope. |
| Direct grid-vendor integration is absent; only a guard checks grid-adapter test usage. | `phase-frontend-fixtures.mjs` reads test source for `@cartulary/grid-adapter` guard. | medium | defer | FE phase-accounting plus grid-adapter test policy. | Preserve guard unless a frontend testing owner replaces it. |
| Test-only assumptions remain under harness files, not production code. | Target is under `tools/harness` and local tests generate synthetic fixtures. | low | intentional/no_action | Testing harness. | Ensure future moves do not leak harness assumptions into product modules. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish scope, authority, write constraints, and tracker. | This tracker only. | Read-only inspection; markdown validation later if permitted. | Tracker exists and session log appended. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03 | Inventory every tracked file under target path. | All `tools/harness/phase-accounting/**` files. | `git ls-files`, source scan, optional `make phase-map-check`. | Section 2 has one row per file. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-04, WF-05 | Map harness public surfaces, schemas, and non-applicable product surfaces. | Harness NLSpec, task-surface/topology metadata, schema refs. | `make json-shape-check`, `make harness-contract` when implementation is authorized. | Section 4 freezes contracts. |
| WF-03 | Characterization test gap analysis | chain | WF-01 | WF-04, WF-08 | Map current smoke tests and missing retained-root coverage. | Local phase-accounting tests plus retained run roots. | `make run-harness-smoke-extended`; direct FE audit if roots exist. | Section 8 records commands and TODO roots. |
| WF-04 | Boundary/coupling scan | chain | WF-02, WF-03 | WF-05, WF-06 | Classify mixed responsibilities and import-boundary risks. | `phase-slice-cli.mjs`, `phase-slice-plan.mjs`, `frontend-row-accounting.mjs`, import-boundary checks. | `make harness-contract`, `make lint-scripts`. | Section 5 findings classified. |
| WF-05 | Facade or ownership redesign plan | parallel | WF-04 | WF-06 | Plan future facade splits without behavior drift. | Phase-slice, FE row-accounting, test-output, scheduler helpers. | TODO: run characterization first. | Deferred unless implementation is authorized. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Sequence behavior-preserving future slices. | Tracker and future implementation files. | Slice-specific Make targets. | Section 7 has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Plan task-surface/topology/generation updates for any path move. | Owner inputs, generated mirrors, smoke tests. | `make phase-schedules`, drift checks. | Generation plan names owner inputs before generated outputs. |
| WF-08 | Validation and final handoff | chain | WF-03, WF-06, WF-07 | none | Record commands, results, skipped checks, blockers, and next action. | Tracker and validation artifacts. | See Section 8. | Session log current enough to resume. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/test-harness-phase-accounting-module-refactor-tracker.md` | Documentation accuracy only. | Preserve no code/test changes. | `make lint-markdown` when writes beyond tracker are permitted; skipped in this session. | Delete tracker if wrong. | Tracker exists with all required sections. |
| S-01 | S-00 | Run characterization baseline for phase maps, FE audit fixtures, phase-slice plans, and import-boundary checks. | No source edits expected. | Existing harness failures may block implementation. | Preserve phase-accounting smoke and harness-contract coverage. | `make phase-map-check`, `make phase-test-name-check`, `make run-harness-smoke-extended`, `make harness-contract`. | No rollback; read-only evidence. | Baseline results and run roots are recorded. |
| S-02 | S-01 | Clarify facades without moving behavior; keep `phase-manifest.mjs`, `frontend/index.mjs`, `frontend-readiness.mjs`, and `phase-slice-plan.mjs` as facade candidates. | Phase-accounting facade files and import-boundary metadata if needed. | Facade exports and importer paths may drift. | Preserve harness-contract import-boundary tests. | `make harness-contract`, `make lint-scripts`. | Revert facade metadata/import-only edits. | Callers still import declared facades and public output is unchanged. |
| S-03 | S-01, S-02 | Split test-output indexing from FE row-accounting only if characterization proves no contract drift. | `frontend-row-accounting.mjs`, `frontend/index.mjs`, `tools/harness/output/test-output/*`. | `frontend-row-accounting.json` closure semantics and artifact refs. | Preserve FE audit smoke and browser target-summary tests. | `make run-harness-smoke-extended`, `make json-shape-check`. | Restore original FE facade and row-accounting imports. | `cartulary.frontend_row_accounting.v3` output unchanged. |
| S-04 | S-01, S-02 | Reduce `phase-slice-cli.mjs` scheduler/runtime coupling through a later authorized harness refactor. | `phase-slice-cli.mjs`, scheduler runtime helpers, scheduler runner/reporting. | `phase-slice`, `service-backed-slice`, scheduler summaries/logs, service wrapper behavior. | Preserve `tests/test-run-phase-slice.sh` and tool-output real target smoke. | `make run-harness-smoke-extended`, targeted `make phase-slice PHASE=phase4`, targeted `make service-backed-slice PHASE=phase4` if services are ready. | Restore old CLI/runtime attachment path. | Public target behavior and summaries are byte/shape compatible where required. |
| S-05 | S-02, S-03, S-04 | If any helper path referenced by task-surface/topology changes, update owner inputs and regenerate through Make. | Task-surface/topology owner inputs and generated mirrors. | Generated drift, command invocation, helper path metadata. | Preserve generated-artifact tests and harness-contract. | `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`. | Restore owner input and regenerated outputs together. | Drift checks pass and no generated file was hand-edited. |
| S-06 | S-05 | Run final drift/static/full validation according to touched surface. | No additional planned edits. | Broad harness or product regression. | Preserve all affected smoke and drift tests. | `make agent-finalize` before broad verification; then `make check` if scope warrants. | Revert last slice or document unrelated failure. | Handoff includes exact passing/failing commands and artifacts. |

S-02 through S-05 were partially implemented in the 2026-07-05 authorized harness remediation session where supported by characterization and import-boundary evidence. Any further behavior change remains out of scope unless separately authorized.

## 8. Validation Plan

Validation run in the 2026-07-05 remediation session:

- `node --check tools/harness/output/test-output/frontend-row-evidence.mjs` passed.
- `node --check tools/harness/scheduler/phase-slice-execution.mjs` passed.
- `node --check tools/harness/phase-accounting/frontend-row-accounting.mjs` passed.
- `node --check tools/harness/phase-accounting/phase-slice-cli.mjs` passed.
- `make phase-map-check` passed.
- `make phase-test-name-check` passed.
- `make harness-contract` passed.
- `make json-shape-check` passed after `make phase-schedules` regenerated scheduler/topology metadata.
- `make generated-artifact-policy-check` passed.
- `make generate-drift` passed.
- `make phase-schedule-drift` passed.
- `make lint-scripts` passed.
- `make lint-shell` passed.
- `make lint-markdown` passed.
- `make run-harness-smoke-extended` passed after fixing the scheduler smoke fixture event-order race.
- `make phase-slice PHASE=phase4 JSON=1` passed.
- `make service-backed-slice PHASE=phase4 JSON=1` passed.
- `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` passed.
- `make agent-finalize` passed; retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Command discovery note: `make harness-smoke-check-scheduler` and `make harness-smoke-phase-slice` were tried as direct public targets and failed with "No rule to make target"; the Make-owned validation surface is `make run-harness-smoke-extended` plus the public `phase-slice`/`service-backed-slice` targets.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make phase-map-check` and `make phase-test-name-check` | Base phase-map validation and phase test-name evidence naming. | yes | Passed in remediation session. |
| integration | `make run-harness-smoke-extended` | Harness smoke for FE audit, phase-slice, phase-test-name, target-plan, and related helpers. | yes | Passed in remediation session. |
| e2e/browser | `make frontend-evidence-audit CHECK_RESULTS_DIR=<root> BROWSER_SUPPORT_RESULTS_DIR=<root> BROWSER_VISUAL_RESULTS_DIR=<root> BROWSER_A11Y_RESULTS_DIR=<root> PHASE_NAMESPACE=frontend PHASE=<FE-PN>` | Retained FE row-accounting audit over broad/support/visual/a11y roots. | no for tracker; yes before FE audit behavior changes | TODO: retained roots required. Do not infer closure from phase names or ledgers. |
| generated drift | `make json-shape-check`, `make generated-artifact-policy-check`, `make phase-schedule-drift`, `make generate-drift` | Schema attachments, generated policy, task-surface/topology/schedule drift. | yes if schemas, owner inputs, task-surface, topology, or generated paths change | Passed in remediation session; `make phase-schedules` was run before drift checks when JSON shape detected stale schedule inputs. |
| import-boundary/static | `make harness-contract`, `make lint-scripts`, `make lint-shell` | Harness helper ownership/import-boundary plus JS/shell script checks. | yes for implementation slices | Passed in remediation session. |
| full check | `make agent-finalize` then `make check` | Broad developer verification gate. | no for tracker; yes when implementation breadth warrants it | `make agent-finalize` passed; `make check` not run because harness-specific broad smoke and drift/static gates passed. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Normalize target label and scope. | WF-00 | DONE | none | Target label `test-harness-phase-accounting`; this tracker. | Label is lowercase kebab case and path/output are explicit. |
| T-002 | Confirm target path exists. | WF-01 | DONE | T-001 | `git ls-files tools/harness/phase-accounting`; source inventory. | Existing tracked target files listed. |
| T-003 | Inventory every tracked target file. | WF-01 | DONE | T-002 | Section 2 table. | All 28 tracked files have rows. |
| T-004 | Record architectural non-product ownership finding. | WF-02 | DONE | T-003 | Sections 1, 3, and 4. | Product contracts are marked not owned by this target. |
| T-005 | Map harness public contract freeze. | WF-02 | DONE | T-003 | Section 4 table. | Make targets, schema IDs, retained artifacts, and generated metadata are frozen. |
| T-006 | Classify coupling and boundary findings. | WF-04 | DONE | T-004, T-005 | Section 5 table. | Each finding has classification and planning action. |
| T-007 | Record existing characterization tests and gaps. | WF-03 | DONE | T-003 | Sections 4, 8, and 11. | Existing smoke tests and retained-root FE audit gap are named. |
| T-008 | Define future behavior-preserving slices. | WF-06 | DONE | T-006, T-007 | Section 7 table. | Slices include dependencies, risks, validation, rollback, and completion criteria. |
| T-009 | Plan generated/task-surface handling for path moves. | WF-07 | DONE | T-008 | `make phase-schedules`; `tools/execution_topology_render_index.json`. | Owner implementation changed first, generated metadata refreshed through Make, and drift checks passed. |
| T-010 | Run validation targets. | WF-08 | DONE | T-008 | Section 8 validation list. | Harness-specific syntax, smoke, contract, drift, static, and targeted phase-slice commands passed. |
| T-011 | Direct FE audit retained-root validation. | WF-03 | BLOCKED | T-007 | RB-001. | Retained broad/support/visual/a11y roots are available or task records skip reason. |
| T-012 | Later product/public contract authorization. | WF-05 | DEFERRED | T-008 | This tracker. | User authorizes a behavior change, public contract change, or product-surface task beyond harness internals. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Tracker created from live inspection and prior proposed plan. | Inspected planning framework, harness NLSpec, domain doc, AGENTS, Core 00-04 excerpts, target files, task-surface/topology refs; touched this tracker only. | `sed`, `find`, `wc`, `rg`, `git status`, `git rev-parse`, `git ls-files`, read-only `node` source scan, `make help`, `make help-all`, `date`. | Scope and authority recorded; no production refactor. | Make validation skipped due one-file write constraint. | If implementation is authorized, start with S-01 characterization. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | Authorized harness remediation implemented; no product behavior changed. | Touched tracker, `docs/testing-harness-nlspec.md`, phase-accounting helpers, scheduler adapter/test fixture, test-output helper, harness-contract tests, and generated topology render index. | `make phase-map-check`, `make phase-test-name-check`, `make run-harness-smoke-extended`, `make harness-contract`, drift/static/lint commands listed in Section 8. | Harness-specific validation passed. | RB-001 direct retained-root FE audit remains open. | Run direct FE audit when retained broad/support/visual/a11y roots exist. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Backend coupling is harness target/shard planning only, not product backend module ownership. | Inspected target phase files plus backend target-plan references; touched tracker only. | `rg`, source scan, `sed`. | Backend owner candidates remain `backend_target_plan` and `backend_shard_plan` facades when phase-slice planning consumes backend rows. | None for planning. | Preserve backend facade imports in any future phase-slice refactor. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | No backend product code changed; backend references remain planner inputs. | Touched phase-slice CLI/runtime adapter and generated topology metadata only for harness execution. | `make phase-slice PHASE=phase4 JSON=1`, `make service-backed-slice PHASE=phase4 JSON=1`, `make run-harness-smoke-extended`. | Public phase-slice plan targets passed. | None. | Keep backend target-plan ownership separate from phase-slice CLI runtime execution. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Target owns FE readiness/evidence accounting only, not web app shell or grid-vendor runtime behavior. | Inspected FE phase manifest, FE row accounting, FE phase-slice planning, FE audit CLI, FE facade; touched tracker only. | `sed`, source scan, `rg`. | FE owner candidates recorded as `frontend_phase_accounting`, `frontend_row_accounting`, and `frontend_evidence_audit`. | Direct retained-root audit blocked without roots. | Characterize FE row accounting before any FE facade split. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | FE row closure remains in phase accounting; runner title/artifact parsing moved to test-output helper ownership. | Touched `frontend-row-accounting.mjs`, `frontend/index.mjs`, `frontend-phase-manifest` callers, `tools/harness/output/test-output/frontend-row-evidence.mjs`, harness-contract tests. | `make run-harness-smoke-extended`, `make harness-contract`, `make json-shape-check`. | FE audit fixture smoke and contract gates passed; FE facade no longer re-exports test-output indexes. | RB-001 retained real-root audit remains open. | Run direct `make frontend-evidence-audit ...` with all four retained roots when available. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Public Make targets, schema IDs, retained artifact paths, and generated task-surface/topology refs are frozen. | Inspected task-surface/topology refs and schema/NLSpec mentions; touched tracker only. | `rg`, `sed`. | Generated outputs marked no-hand-edit; path moves require owner inputs plus Make regeneration. | No generated drift validation run. | If helper paths move, plan owner input edits and `make phase-schedules` before drift checks. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | Schema IDs, Make target names, command IDs, retained paths, and output contracts remain stable. | Touched `render-service-backed-schedule-manifest.mjs` import, `tools/execution_topology_render_index.json`, import-boundary owner registry, harness-contract tests. | `make phase-schedules`, `make json-shape-check`, `make phase-schedule-drift`, `make generate-drift`, `make generated-artifact-policy-check`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`. | Generated metadata refreshed through Make and drift checks passed. | None for current generated state. | Future path moves must follow owner-input-first regeneration again. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Existing harness smoke covers phase-test-name, FE audit fixtures, phase-slice, and target/shard plan behavior. | Inspected all phase-accounting test files; touched tracker only. | `sed`, `rg`, source scan. | Validation plan lists Make-owned targets; none run. | One-file write constraint prevented artifact-emitting validation. | Run S-01 baseline in later implementation/validation session. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | Harness smoke and contract tests now cover new owner boundaries and scheduler fixture race. | Touched `tools/harness/tests/test-harness-contracts.mjs` and `tools/harness/scheduler/tests/test-check-scheduler.sh`. | `make run-harness-smoke-extended`, `make harness-contract`, `make lint-shell`, `make lint-scripts`. | Extended harness smoke passed after fixture assertion fix. | Direct single smoke targets are not public Make targets. | Use `make run-harness-smoke-extended` for smoke validation unless a public narrow target is added. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | No product auth/session/authorization code found in target; Make env sanitization remains harness mechanics. | Inspected Core 04 excerpt, phase-slice CLI, task-surface bindings; touched tracker only. | `sed`, `rg`. | Authorization outcomes marked not applicable unless later product code changes. | None for planning. | Keep Core 04 product security separate from harness command behavior. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | No product auth/session/authorization code changed. | Touched harness-only scheduler/test-output/phase-accounting files. | `make harness-contract`, `make run-harness-smoke-extended`. | No product security boundary changed. | None. | Any product authorization work needs separate Core-backed plan. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T22:17:37-04:00 | Codex GPT-5 tracker creation | Tracker is ready for continuation; implementation remains deferred. | Touched tracker only. | `date`, `git status`, plus inspection commands listed above. | Open blockers recorded in Section 11. | Retained roots and generated-path move guard. | Start S-01 or authorize a specific later implementation slice. |
| 2026-07-05T03:01:02Z | Codex GPT-5 remediation | Main remediation slices are implemented and validated; remaining open item is direct real-root FE audit. | Touched files listed in git diff; no product files. | Section 8 commands. | Validation passed except direct FE audit was not run because retained roots were not available. | RB-001. | Provide retained broad/support/visual/a11y roots for final FE audit proof. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Direct `frontend-evidence-audit` validation needs retained broad/support/visual/a11y roots. | FE audit cannot be proven complete from phase names, generated ledgers, or a broad check root alone. | Retained roots for `CHECK_RESULTS_DIR`, `BROWSER_SUPPORT_RESULTS_DIR`, `BROWSER_VISUAL_RESULTS_DIR`, and `BROWSER_A11Y_RESULTS_DIR`. | Open for direct audit validation. |
| RB-002 | Any future helper path move touching task-surface/topology requires owner-input edits and Make regeneration. | Generated outputs must not be hand-edited and public command metadata must not drift. | Owner input diff plus `make phase-schedules`, drift checks, and harness-contract results. | Applied in this remediation for current path/import metadata; remains an open planning guard for future moves. |

## 12. Binary Completion Criteria

| Criterion | Tracker status |
| --- | --- |
| Every file in `tools/harness/phase-accounting` is inventoried or explicitly out of scope. | Met in Section 2. |
| Every discovered public contract risk has an owner and test posture. | Met in Section 4; retained-root FE audit remains RB-001. |
| Every proposed workflow has dependencies and exit criteria. | Met in Section 6. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | Met in Section 7. |
| Validation commands are discovered or marked `TODO` with a reason. | Met in Section 8; harness-specific commands were run in the remediation session. |
| Contradictions are marked `BLOCKED: owner contradiction`. | No owner contradiction found during inspection. |
| Repository/framework mismatches are recorded as planning findings. | Met: framework module catalog does not make this target a product module; Sections 1 and 3 record the mismatch. |
| Handoff sections are current enough for another agent to continue without rediscovery. | Met in Section 10, subject to RB-001 retained-root evidence for FE audit validation. |
