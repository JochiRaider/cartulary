# Test harness scripts Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `tools/harness` |
| Target label | `Test harness scripts` |
| Normalized target key | `test-harness` |
| Output path | `docs/handoffs/test-harness-module-refactor-tracker.md` |
| Planning status | Planning and documentation only. |
| Allowed changes in this session | Create or update this tracker file only. Supporting-plan closure is authorized only inside this tracker. |
| Non-goals | No production refactor, no test edits, no contract edits, no generated-artifact edits, no package configuration edits, no migration edits, and no harness implementation edits. |
| Later authorization | Code implementation requires a later authorized task that names one selected slice ID and starts from the RB-002 characterization baseline. This tracker is not implementation approval. |
| Target existence | `tools/harness` exists. `git ls-files tools/harness` reported 224 tracked files. |
| Existing tracker state | This tracker did not exist before this session. No prior handoff history existed in this output file to preserve. |

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and framework files as evidence only.

Owner documents inspected:

| Document | Role in this tracker | Notes |
| --- | --- | --- |
| `AGENTS.md` | Repository procedure and command/write constraints. | Confirms `tools` contains repo-local automation and that Make targets are canonical for repository commands. |
| `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Planning template and doctrine. | Treated as planning doctrine, not repository-state proof. |
| `docs/testing-harness-nlspec.md` | Adopted harness mechanics owner. | Owns command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, verification gates, public Make surface, output classes, schemas, scheduler, service lifecycle, test-only routes, security, and generated-artifact handling. |
| `docs/domain.md` | Domain vocabulary and concept-boundary reference. | Confirms harness terms are implementation-support unless owner docs promote narrower behavior. |
| `docs/spec/00_document_set_status_and_precedence.md` | Normative authority and contract-owner matrix. | Confirms Core 00 through Core 04 own implementation conformance; Core 05 is not Base Profile implementation conformance. |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Product module boundaries and architecture. | Confirms product module concerns such as auth, incidents, timeline, entities, evidence, imports, links, revisions, projections, collaboration, reporting. |
| `docs/spec/02_domain_model_schema_and_history.md` | Domain model, source state, history boundaries. | Confirms source-state and revision behavior are product-domain concerns, not harness behavior. |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook and collaboration behavior. | Confirms workbook surface, saved-view, inspector, and collaboration behavior are product contracts, not harness ownership. |
| `docs/spec/04_security_deployment_and_conformance.md` | Security, authorization, deployment, conformance. | Confirms product auth/authorization behavior remains Core-owned. Harness test-route security is harness-owned only inside the adopted harness scope. |

Repository files and surfaces inspected:

| Surface | Evidence gathered |
| --- | --- |
| `tools/harness/**` | `git ls-files tools/harness` reported 224 tracked files. All files are inventoried in Section 2. |
| `Makefile` | Public Make target wiring and harness owner script paths. |
| `tools/task_surface_manifest.json` | Current task-surface public target metadata and backing scripts. |
| `tools/execution_topology_manifest.json` | Scheduler/topology reachability and harness self-test ownership. |
| `tools/task_surface.generated.mk` | Generated downstream Make output referenced for generated-output risk only; not edited. |
| `docs/handoffs/scripts-module-refactor-tracker.md` | Prior evidence for migration from root `scripts/` into owner-local `tools/harness/**`; not treated as current-state authority. |
| `make help` and `make help-all` output | Current public validation/discovery target names. |

Architectural source posture:

- `tools/harness` is a broad implementation-support subsystem for harness mechanics and evidence accounting.
- It is not assumed to be a valid permanent product module boundary merely because the path exists.
- Live inspection found no current ownership of legitimate workbook orchestration behavior by `tools/harness`.
- Product behavior remains owned by Core 00 through Core 04 and the named modules/packages under `internal/modules`, `internal/platform`, `apps/web`, `packages/*`, `contracts/*`, `db/*`, and generated roots as applicable.
- Phase maps, test rows, generated schedules, and ledgers are verification/evidence accounting only. They are not runtime product architecture.
- No owner contradiction was found during this planning session.
- Core 00 through Core 04 do not need revision for the current tracker/supporting-plan closure because no product behavior is changed. Core 05 does not need revision unless a later slice changes claim-bearing timed, fixture-sensitive, or publication evidence behavior.
- `docs/testing-harness-nlspec.md` does not need revision for behavior-preserving internal harness refactors. It must be reviewed and revised before any later change to public target membership, command IDs, output classes, accepted inputs, artifact policies, scheduler command shapes, cache behavior, test-route security, failure mapping, or other adopted harness mechanics.
- Public Make command semantics, output schemas, accepted inputs, route shapes, authorization outcomes, generated contract surfaces, and multiple bundled implementation slices require separate owner-spec or user authorization before code work.
- Harness test-route tokens, origin/host checks, redaction, and reset behavior are harness mechanics only; product authorization remains Core 04-owned.
- Default runtime must not expose `/api/v1/test/*` or `/ws/v1/test/*`. No harness refactor may move workbook row/query/mutation, grid-vendor integration, product WebSocket, saved-view, projection, or product authorization ownership into `tools/harness`.

## 2. Current-State Repository Inventory

No tracked file under `tools/harness` is out of scope for this tracker. The table below inventories all 224 tracked files by current owner seam. The explicit file coverage list after the table names every file included in each row.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tools/harness/backend/**` (23 files) | Go target planning/running, Go shard and duration accounting, migration history checks, schema-object ownership validation, Govulncheck finding normalization, backend build wrapper. | ESM helpers such as `collectGoShardPlan`, `collectGoShardsForTarget`, `runGoTargetCLI`, duration baseline helpers, migration/schema validation helpers; shell entrypoints for Go phase and migration checks. | `Makefile`, `tools/task_surface_manifest.json`, `tools/execution_topology_manifest.json`, backend self-tests, scheduler service-backed expansion, planning target-plan code. | Node core modules, Go toolchain, `go test -json`, `../core/*`, `../planning/*`, `../scheduler/*`, migration SQL, schema manifests. | `tools/harness/backend/tests/*`, scheduler and planning tests that import Go plan helpers. | `cartulary.go_test_duration_baselines.v4`, `cartulary.go_shard_plan.v3`, `cartulary.test_phase_summary.v3`, `cartulary.govulncheck_findings.v1`, `cartulary.migration_history_manifest.v1`, `cartulary.schema_object_ownership_manifest.v1`. | `harness/backend-runner` plus `harness/backend-drift`. | high | Backend runner behavior affects public Go test summaries, sharding, fixture accounting, and security scan artifacts. |
| `tools/harness/browser/**` (30 files) | Playwright batch/manifest selection, browser shard planning, browser stack lifecycle, runtime reset, web E2E startup, visual/a11y/measurement wrappers, Playwright install. | ESM helpers such as browser batch manifest loading, browser scheduler dependency helpers, `createPlan`, Playwright report flattening; shell entrypoints for browser public targets. | `Makefile`, task-surface/topology manifests, browser public Make targets, scheduler browser runtime, frontend accounting, browser self-tests. | Node core modules, Playwright, Vite preview, Go server/migrate binaries, service-backed runtime, `../core/run-phase-common.sh`, `../frontend/*`, `../planning/*`, readiness process lifecycle. | `tools/harness/browser/tests/*`, scheduler tests for browser stages, frontend row-accounting consumers. | `cartulary.browser_e2e_batch_manifest.v5`, `cartulary.browser_e2e_shard_plan.v2`, `cartulary.playwright_manifest_selection.v1`, `cartulary.web_e2e_stack.v3`, `cartulary.test.runtime_identity.v1`, `cartulary.browser_startup_diagnostics.v1`, `cartulary.web_e2e_session_lease.v1`, `cartulary.test.runtime_reset.v1`. | `harness/browser-runtime` and `harness/playwright-runner`. | high | Transport/runtime-adjacent harness logic; must not become product WebSocket/HTTP behavior owner. |
| `tools/harness/core/**` (35 files) | Harness contract implementation, result roots/run IDs, output summaries, artifact discovery/writing, failure taxonomy, finalization, Make wrapper plumbing, phase runner common logic, tool-output summaries. | ESM exports for failure taxonomy, harness contract functions, artifact helpers, runner context, Make-node tool dispatch, fixture reporting, tool output builders, test-output schemas; shell wrappers for test-output and run-phase. | `Makefile`, generated task-surface output, every harness subtree, public Make wrappers, scheduler runners, browser/backend/frontend scripts, harness self-tests. | Node core modules, JSON schemas, `../generated-artifacts/*`, `../planning/*`, child Make targets, shell utilities. | `tools/harness/core/tests/*` plus many subtree tests using `test-output.sh`, `run-phase-common.sh`, and harness-contract helpers. | `cartulary.tool_run_summary.v3`, `cartulary.test_phase_summary.v3`, `cartulary.test_target_summary.v4`, `cartulary.test_run_summary.v6`, `cartulary.fixture_report.v1`, `cartulary.agent_finalize_summary.v3`, `cartulary.agent_finalize_action_cache_record.v1`, additional summary/artifact schemas. | `harness/core-contract`. | high | This is the likely deepest harness facade; public Make behavior must remain stable while internals move. |
| `tools/harness/frontend/**` (20 files) | Vitest phase wrappers, frontend unit aggregation, frontend phase/row accounting, frontend evidence audit, accessibility summaries, design-token and font-bundle checks, web build/install/embed wrappers. | ESM exports for design-token parsing/rendering, frontend phase manifest validation, frontend slice plans, row accounting helpers, font bundle checks; shell entrypoints for Vitest/frontend public targets. | `Makefile`, frontend public Make targets, generated-artifact checks, browser wrappers, scheduler/check topology, frontend self-tests. | Node core modules, Vitest reports, Playwright reports, `../browser/playwright-report.mjs`, `../core/harness-contract.mjs`, `../core/test-output/context.mjs`, design docs. | `tools/harness/frontend/tests/*`, browser tests and generated-artifact tests consuming frontend map/accounting helpers. | `cartulary.frontend_phase_registry.v2`, `cartulary.frontend_phase_test_map.v3`, `cartulary.frontend_visual_fixture_registry.v2`, `cartulary.frontend_row_accounting.v3`, `cartulary.frontend_accessibility_summary.v2`, `cartulary.frontend_accessibility_preflight_summary.v1`, `cartulary.design_tokens.v1`, `cartulary.font_manifest.v1`, `cartulary.vitest_failure_details.v1`. | `harness/frontend-readiness`. | medium | Frontend evidence is implementation-readiness/design-direction evidence unless a product owner contract says otherwise. |
| `tools/harness/generated-artifacts/**` (18 files) | Generators and drift/shape checkers for task surface, execution topology, schedules, ledgers, JSON schemas, generated artifact policy, and generation drift. | ESM exports for execution topology, task-surface rendering, JSON shape helpers, render index, phase ledger rendering, service-backed schedule rendering; shell entrypoints for generation/drift. | `Makefile`, public drift/generation targets, task-surface/topology manifests, scheduler/planning tests, generated-artifact self-tests. | Node core modules, `../planning/*`, `../browser/*`, `../scheduler/*`, `../core/harness-contract.mjs`, tools manifests/schemas. | `tools/harness/generated-artifacts/tests/*`, core and planning task-surface tests. | `cartulary.execution_topology.v3`, `cartulary.task_surface_manifest.v15`, `cartulary.scheduler_manifest.v1`, `cartulary.check_schedule_sources.v1`, `cartulary.service_backed_schedule_sources.v1`, `cartulary.service_backed_schedule.v11`, `cartulary.execution_topology_render_index.v1`, generated ledgers/schedules/task surface. | `harness/generated-artifact-tools`. | high | These files are authored generator/checker inputs, not generated outputs. Downstream generated files must be refreshed through Make, never hand-edited. |
| `tools/harness/planning/**` (23 files) | Phase manifest validation, phase registry, phase slices, target plans, task guides, task execution maps, task-surface reporting, explain commands. | ESM exports for phase manifest/registry, phase slice plans, target plans, task guidance, execution maps, summary topology; shell check helpers. | Public `task-guide`, `target-plan`, `explain-*`, `phase-slice`, `service-backed-slice`, generated-artifact renderers, backend/browser/frontend runners, planning self-tests. | Node core modules, `../generated-artifacts/*`, `../backend/*`, `../browser/*`, `../scheduler/*`, `../core/*`, phase maps/manifests. | `tools/harness/planning/tests/*`, backend/browser/frontend tests that use phase/target planning. | `cartulary.phase_registry.v1`, `cartulary.phase_test_map.v2`, `cartulary.phase_slice_plan.v1`, `cartulary.phase_slice_scheduler_summary.v4`, `cartulary.task_execution_map.v1`, `cartulary.task_guide.v1`, `cartulary.task_surface_report.v3`, guidance schemas. | `harness/planning-evidence`. | medium | Phase maps and task rows are evidence accounting only; they must not drive product runtime architecture. |
| `tools/harness/readiness/**` (18 files) | Toolchain bootstrap, Node/pnpm/ShellCheck/Go helper provisioning, cache helper, local dev services/stack, process lifecycle, inotify diagnostics, build-input discovery, doctor checks. | Shell entrypoints and helper functions; ESM inotify/toolchain pin CLIs. | `Makefile` readiness/build rules, public `doctor`, `bootstrap`, `frontend-toolchain`, `frontend-install`, `playwright-install`, dev-service targets, browser stack, readiness self-tests. | Shell, Docker/Compose, curl, Node/pnpm, Go toolchain, process lifecycle helper, cache profiles. | `tools/harness/readiness/tests/*`, browser lifecycle tests, core harness-contract tests. | `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.inotify_diagnostics.v1`, `cartulary.toolchain_pins.v1`. | `harness/readiness`. | medium | Readiness/cache behavior is local acceleration and setup mechanics, not product conformance. |
| `tools/harness/scheduler/**` (34 files) | Check and service-backed scheduler CLIs, manifest validation, resource registry, DAG engine, process executor, progress/events, timing drift, duration baselines, service-backed schedule topology/expansion. | ESM exports for scheduler manifest/resource/reporting/engine/state/event/timing helpers, service-backed topology, duration helpers, browser runtime command helpers; scheduler CLI entrypoints. | `Makefile`, public `check`, `service-backed-slice`, scheduler drift targets, generated topology, backend/browser/planning code, scheduler self-tests. | Node core modules, `../backend/go-shard-plan.mjs`, `../browser/*`, `../core/*`, `../planning/*`, generated schedules/manifests, child Make targets. | `tools/harness/scheduler/tests/*`, core harness-contract tests, planning phase-slice tests. | `cartulary.scheduler_manifest.v1`, `cartulary.scheduler_event.v6`, `cartulary.check_scheduler_summary.v10`, `cartulary.service_backed_scheduler_summary.v10`, `cartulary.scheduler_resource_registry.v4`, `cartulary.scheduler_pressure_summary.v1`, `cartulary.scheduler_work_unit_duration_baselines.v2`. | `harness/scheduler-core` plus `harness/scheduler-adapters`. | high | Scheduler has coherent core logic but also adapter coupling to backend/browser planning; split only after characterization. |
| `tools/harness/static-analysis/**` (20 files) | Backend module boundary check, frontend import boundary check, Biome wrappers, Go format/vet/staticcheck/security wrappers, Markdown/script/shell lint, Fallow static wrapper. | CLI entrypoints and wrappers; no public product API. | Public lint/static/security Make targets, task-surface/topology manifests, generated-artifact policy checks, static-analysis self-tests. | Node core modules, TypeScript compiler, Fallow, Biome, ShellCheck, Go tooling, generated-artifacts helper. | `tools/harness/static-analysis/tests/*`. | `cartulary.backend_module_boundaries.v1`, `cartulary.backend_module_boundary_summary.v1`, `cartulary.frontend_import_boundaries.v2`, `cartulary.fallow_static_summary.v1`, `cartulary.govulncheck_findings.v1`. | `harness/static-analysis`. | medium | Test fixtures intentionally contain forbidden imports and generated-path examples; do not misclassify them as production coupling. |
| `tools/harness/test-support/**` (3 files) | Harness self-test support for scratch directories, JSON assertions, and artifact assertions. | Helper CLIs and shell functions for tests only. | Harness self-tests across backend/browser/core/generated/planning/readiness/scheduler/static-analysis; release-evidence tests. | Node core modules, shell temp directories, retained artifacts. | Many `tools/harness/**/tests/*` files source or invoke these helpers. | None directly; validates retained summary/artifact shapes. | `harness/test-support`. | low | Test-only support should stay out of production harness public contracts except where NLSpec names scratch behavior. |

Per-file inventory coverage:

```text
tools/harness/backend/build-go-artifact.sh
tools/harness/backend/check-migrations.sh
tools/harness/backend/go-duration-artifacts.mjs
tools/harness/backend/go-duration-baselines.mjs
tools/harness/backend/go-shard-plan-cli.mjs
tools/harness/backend/go-shard-plan.mjs
tools/harness/backend/go-target-aggregate.mjs
tools/harness/backend/go-target-plan-coverage-cli.mjs
tools/harness/backend/go-target-runner.mjs
tools/harness/backend/go-test-duration-baseline-coverage-cli.mjs
tools/harness/backend/go-test-duration-baseline-drift-cli.mjs
tools/harness/backend/go-test-duration-baselines-cli.mjs
tools/harness/backend/govulncheck-findings.mjs
tools/harness/backend/migration-history-cli.mjs
tools/harness/backend/migration-history.mjs
tools/harness/backend/postgres-fixture-budget-cli.mjs
tools/harness/backend/run-go-manifest-phase.sh
tools/harness/backend/run-go-phase.sh
tools/harness/backend/schema-object-ownership-cli.mjs
tools/harness/backend/schema-object-ownership.mjs
tools/harness/backend/tests/test-check-migrations.sh
tools/harness/backend/tests/test-go-test-duration-baselines.sh
tools/harness/backend/tests/test-run-go-target.sh
tools/harness/browser/browser-batch-manifest.mjs
tools/harness/browser/browser-scheduler-dependencies.mjs
tools/harness/browser/browser-shard-plan.mjs
tools/harness/browser/playwright-install.sh
tools/harness/browser/playwright-owned-stack.sh
tools/harness/browser/playwright-report.mjs
tools/harness/browser/reset-web-e2e-stack.sh
tools/harness/browser/run-browser-e2e-a11y-preflight.sh
tools/harness/browser/run-browser-e2e-a11y.sh
tools/harness/browser/run-browser-e2e-batch.sh
tools/harness/browser/run-browser-e2e-functional.sh
tools/harness/browser/run-browser-e2e-manifest-dependency.sh
tools/harness/browser/run-browser-e2e-measurement.sh
tools/harness/browser/run-browser-e2e-owned-stack.sh
tools/harness/browser/run-browser-e2e-resettable.sh
tools/harness/browser/run-browser-e2e-stateful.sh
tools/harness/browser/run-browser-e2e-target.sh
tools/harness/browser/run-browser-e2e-visual-update.sh
tools/harness/browser/run-browser-e2e-visual.sh
tools/harness/browser/run-browser-e2e-webserver-backed.sh
tools/harness/browser/run-playwright-manifest-phase.sh
tools/harness/browser/run-playwright-phase.sh
tools/harness/browser/run-playwright-webserver-batch.sh
tools/harness/browser/start-web-e2e.sh
tools/harness/browser/tests/test-browser-shard-plan.sh
tools/harness/browser/tests/test-run-playwright-manifest-phase.sh
tools/harness/browser/tests/test-run-playwright-phase.sh
tools/harness/browser/tests/test-run-playwright-webserver-batch.sh
tools/harness/browser/tests/test-web-e2e-lifecycle.sh
tools/harness/browser/web-e2e-lifecycle.sh
tools/harness/core/agent-finalize-action-cache.mjs
tools/harness/core/agent-finalize-cli.mjs
tools/harness/core/artifact-discovery.mjs
tools/harness/core/artifact-writer.mjs
tools/harness/core/cartulary-runner-cli.mjs
tools/harness/core/explain-run-cli.mjs
tools/harness/core/failure-taxonomy.mjs
tools/harness/core/fixture-report-cli.mjs
tools/harness/core/fixture-reporting.mjs
tools/harness/core/harness-contract-cli.mjs
tools/harness/core/harness-contract.mjs
tools/harness/core/make-node-tools.mjs
tools/harness/core/repo-paths.mjs
tools/harness/core/result-artifacts.mjs
tools/harness/core/run-harness-smoke-cli.mjs
tools/harness/core/run-make-node-tool-cli.mjs
tools/harness/core/run-make-node-tool.sh
tools/harness/core/run-make-sequence.sh
tools/harness/core/run-phase-common.sh
tools/harness/core/run-phase.sh
tools/harness/core/runner-context.mjs
tools/harness/core/test-output.mjs
tools/harness/core/test-output.sh
tools/harness/core/test-output/cli.mjs
tools/harness/core/test-output/context.mjs
tools/harness/core/tests/test-agent-finalize.sh
tools/harness/core/tests/test-cartulary-runner-service-backed-target.sh
tools/harness/core/tests/test-harness-contracts.mjs
tools/harness/core/tests/test-make-node-tools.sh
tools/harness/core/tests/test-public-make-wrapper-smoke.sh
tools/harness/core/tests/test-run-make-sequence-fast.sh
tools/harness/core/tests/test-run-make-sequence.sh
tools/harness/core/tests/test-run-phase.sh
tools/harness/core/tests/test-tool-output-real-targets.sh
tools/harness/core/tool-output.mjs
tools/harness/frontend/accessibility-summary-cli.mjs
tools/harness/frontend/build-web-artifact.sh
tools/harness/frontend/design-token-cli.mjs
tools/harness/frontend/design-tokens.mjs
tools/harness/frontend/embed-web-assets.sh
tools/harness/frontend/font-bundle-check-cli.mjs
tools/harness/frontend/frontend-evidence-audit-cli.mjs
tools/harness/frontend/frontend-install.sh
tools/harness/frontend/frontend-phase-manifest.mjs
tools/harness/frontend/frontend-phase-slice-plan.mjs
tools/harness/frontend/frontend-row-accounting.mjs
tools/harness/frontend/frontend-toolchain.sh
tools/harness/frontend/run-frontend-unit.sh
tools/harness/frontend/run-vitest-manifest-phase.sh
tools/harness/frontend/run-vitest-phase.sh
tools/harness/frontend/tests/test-frontend-evidence-audit.sh
tools/harness/frontend/tests/test-run-frontend-unit.sh
tools/harness/frontend/tests/test-run-vitest-manifest-phase.sh
tools/harness/frontend/tests/test-run-vitest-phase.sh
tools/harness/frontend/vitest-failure-details.mjs
tools/harness/generated-artifacts/check-generate-drift.sh
tools/harness/generated-artifacts/check-generated-artifact-policy.mjs
tools/harness/generated-artifacts/check-json-shapes.mjs
tools/harness/generated-artifacts/check-phase-ledger-drift.mjs
tools/harness/generated-artifacts/execution-topology.mjs
tools/harness/generated-artifacts/generate-artifacts.sh
tools/harness/generated-artifacts/generated-artifacts.sh
tools/harness/generated-artifacts/json-shape.mjs
tools/harness/generated-artifacts/render-execution-topology-artifacts.mjs
tools/harness/generated-artifacts/render-phase-ledger.mjs
tools/harness/generated-artifacts/render-phase-ledgers.mjs
tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs
tools/harness/generated-artifacts/render-task-surface-make.mjs
tools/harness/generated-artifacts/task-surface.mjs
tools/harness/generated-artifacts/tests/test-execution-topology.sh
tools/harness/generated-artifacts/tests/test-generate-drift.sh
tools/harness/generated-artifacts/tests/test-generated-artifact-policy.sh
tools/harness/generated-artifacts/tests/test-json-shapes.sh
tools/harness/planning/check-phase-maps.sh
tools/harness/planning/explain-phase-cli.mjs
tools/harness/planning/explain-target-cli.mjs
tools/harness/planning/phase-manifest-shape.mjs
tools/harness/planning/phase-manifest.mjs
tools/harness/planning/phase-map-check-cli.mjs
tools/harness/planning/phase-registry.mjs
tools/harness/planning/phase-slice-cli.mjs
tools/harness/planning/phase-slice-plan.mjs
tools/harness/planning/phase-test-name-check-cli.mjs
tools/harness/planning/summary-topology.mjs
tools/harness/planning/target-plan-cli.mjs
tools/harness/planning/target-plan.mjs
tools/harness/planning/task-execution-map.mjs
tools/harness/planning/task-guidance.mjs
tools/harness/planning/task-guide-cli.mjs
tools/harness/planning/task-surface-check-common.sh
tools/harness/planning/task-surface-report-cli.mjs
tools/harness/planning/tests/test-check-phase-test-names.sh
tools/harness/planning/tests/test-print-target-plan.sh
tools/harness/planning/tests/test-run-phase-slice.sh
tools/harness/planning/tests/test-task-guidance.mjs
tools/harness/planning/tests/test-task-surface-report.sh
tools/harness/readiness/bootstrap-go-tool.sh
tools/harness/readiness/bootstrap-node-runtime.sh
tools/harness/readiness/bootstrap-shellcheck.sh
tools/harness/readiness/cache-artifact.sh
tools/harness/readiness/check-doctor.sh
tools/harness/readiness/dev-services.sh
tools/harness/readiness/dev-stack.sh
tools/harness/readiness/diagnose-inotify.mjs
tools/harness/readiness/list-build-inputs.sh
tools/harness/readiness/process-lifecycle.sh
tools/harness/readiness/tests/test-bootstrap-node-runtime.sh
tools/harness/readiness/tests/test-bootstrap-shellcheck.sh
tools/harness/readiness/tests/test-build-input-discovery.sh
tools/harness/readiness/tests/test-cache-artifact.sh
tools/harness/readiness/tests/test-check-toolchain-pins.sh
tools/harness/readiness/tests/test-dev-services-lifecycle.sh
tools/harness/readiness/tests/test-dev-stack-lifecycle.sh
tools/harness/readiness/toolchain-pin-check-cli.mjs
tools/harness/scheduler/check-schedule-cli.mjs
tools/harness/scheduler/check-schedule-manifest.mjs
tools/harness/scheduler/check-service-backed-expansion.mjs
tools/harness/scheduler/duration-baseline-cli.mjs
tools/harness/scheduler/duration-baseline-drift-suite.sh
tools/harness/scheduler/duration-drift.mjs
tools/harness/scheduler/execution-dependencies.mjs
tools/harness/scheduler/harness-smoke-durations-cli.mjs
tools/harness/scheduler/scheduler-browser-runtime.mjs
tools/harness/scheduler/scheduler-cli.mjs
tools/harness/scheduler/scheduler-event-order-drift-cli.mjs
tools/harness/scheduler/scheduler-manifest.mjs
tools/harness/scheduler/scheduler-reporting.mjs
tools/harness/scheduler/scheduler-resources.mjs
tools/harness/scheduler/scheduler-runner.mjs
tools/harness/scheduler/scheduler-summary-timing-drift-cli.mjs
tools/harness/scheduler/scheduler/blockers.mjs
tools/harness/scheduler/scheduler/clock.mjs
tools/harness/scheduler/scheduler/engine.mjs
tools/harness/scheduler/scheduler/event-order.mjs
tools/harness/scheduler/scheduler/process-executor.mjs
tools/harness/scheduler/scheduler/progress-recorder.mjs
tools/harness/scheduler/scheduler/runtime-command-helpers.mjs
tools/harness/scheduler/scheduler/state.mjs
tools/harness/scheduler/scheduler/summary-timing-drift.mjs
tools/harness/scheduler/service-backed-make-target-durations-cli.mjs
tools/harness/scheduler/service-backed-schedule-cli.mjs
tools/harness/scheduler/service-backed-schedule-manifest.mjs
tools/harness/scheduler/service-backed-schedule-topology.mjs
tools/harness/scheduler/target-duration-baselines.mjs
tools/harness/scheduler/tests/test-check-scheduler.sh
tools/harness/scheduler/tests/test-harness-smoke-duration-baselines.sh
tools/harness/scheduler/tests/test-service-backed-make-target-duration-baselines.sh
tools/harness/scheduler/tests/test-service-backed-scheduler.sh
tools/harness/static-analysis/backend-module-boundary-check-cli.mjs
tools/harness/static-analysis/fallow-static-cli.mjs
tools/harness/static-analysis/frontend-biome.sh
tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs
tools/harness/static-analysis/go-format.sh
tools/harness/static-analysis/go-gosec-audit.sh
tools/harness/static-analysis/go-gosec-targeted.sh
tools/harness/static-analysis/go-govulncheck.sh
tools/harness/static-analysis/go-staticcheck.sh
tools/harness/static-analysis/go-vet.sh
tools/harness/static-analysis/markdownlint.sh
tools/harness/static-analysis/scripts-biome.sh
tools/harness/static-analysis/shellcheck.sh
tools/harness/static-analysis/tests/test-fallow-static.sh
tools/harness/static-analysis/tests/test-frontend-import-boundaries.sh
tools/harness/static-analysis/tests/test-lint-shell.sh
tools/harness/static-analysis/tests/test-run-go-gosec-audit.sh
tools/harness/static-analysis/tests/test-run-go-gosec-targeted.sh
tools/harness/static-analysis/tests/test-run-go-govulncheck.sh
tools/harness/static-analysis/tests/test-run-go-staticcheck.sh
tools/harness/test-support/harness-artifact-assert.mjs
tools/harness/test-support/harness-scratch.sh
tools/harness/test-support/json-test-helper.mjs
```

## 3. Module Boundary Diagnosis

Current diagnosis: `tools/harness` is a mixed-responsibility harness subsystem with several legitimate internal harness seams. It is not a permanent product module boundary and should not own workbook orchestration behavior. Live inspection supports the following architectural finding: `tools/harness` owns command/evidence mechanics, target planning, scheduling, readiness, test runtime fixtures, static checks, generated-artifact verification, and self-tests. It does not own timeline, projections, revisions, collaboration, imports/tabular ingest, entities/indicators, evidence records, links, saved views, view contracts, transport/platform adapters for production runtime, frontend shell/controller state, or grid-adapter/vendor integration.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Generic harness public contract plumbing | `tools/harness/core` | `harness/core-contract` | keep/split | `harness-contract.mjs`, `test-output/*`, `failure-taxonomy.mjs`, `runner-context.mjs` | Likely a real facade. Split only to reduce public/private surface after characterization. |
| Go test execution and backend harness accounting | `tools/harness/backend` | `harness/backend-runner` | keep/split | Go runner, shard plan, duration baseline, migration/schema check files. | Backend harness logic, not `internal/modules/*` backend product logic. |
| Browser E2E stack and Playwright orchestration | `tools/harness/browser` | `harness/browser-runtime` and `harness/playwright-runner` | split | `start-web-e2e.sh`, Playwright phase/batch wrappers, browser shard plan. | Browser lifecycle is runtime-adjacent harness logic; Playwright selection can be a separate runner seam. |
| Scheduler engine and schedule adapters | `tools/harness/scheduler` | `harness/scheduler-core` plus `harness/scheduler-adapters` | split | `scheduler/engine.mjs`, `scheduler/state.mjs`, service-backed/check schedule CLIs. | Engine is coherent; service/browser/backend expansion imports create adapter coupling. |
| Frontend readiness/evidence accounting | `tools/harness/frontend` | `harness/frontend-readiness` | split/defer | Frontend phase manifest, row accounting, evidence audit, accessibility summary, design token parsing. | Frontend evidence is implementation support and design-direction support, not product conformance by itself. |
| Generated-artifact and JSON-shape verification | `tools/harness/generated-artifacts` | `harness/generated-artifact-tools` | keep | Execution topology, task surface, schedule, ledger, shape, drift tools. | Authored tooling, not generated output. |
| Phase and target planning | `tools/harness/planning` | `harness/planning-evidence` | keep/split | Phase registry/manifest, task guidance, target plan, explain commands. | Phase identity must remain evidence accounting, not production architecture. |
| Readiness and local tool/service provisioning | `tools/harness/readiness` | `harness/readiness` | keep | Bootstrap, cache, doctor, process lifecycle, dev services. | Local acceleration/setup mechanics only. |
| Static and import-boundary checks | `tools/harness/static-analysis` | `harness/static-analysis` | keep | Backend module boundary, frontend import boundary, lint/security wrappers. | Enforces boundaries; does not own the domain behavior under check. |
| Harness self-test support | `tools/harness/test-support` | `harness/test-support` | keep | Scratch helper, artifact assertion, JSON helper. | Test-only support; scratch behavior is harness-owned when NLSpec names it. |
| Workbook row/query/mutation orchestration | Not observed in `tools/harness` | Product owners under Core 01 through Core 03 and implementation modules/packages | defer/no_action | No source file under target implements product HTTP route handlers, workbook mutation owners, saved-view persistence, or projection materialization. | Record as architectural finding; do not create moves without later evidence. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Public Make target names and command IDs | `docs/testing-harness-nlspec.md` Section 4 and `tools/task_surface_manifest.json` mirror | `make help-all`, task-surface manifest, Makefile | Harness contract tests, task-surface report tests, generated-artifact tests | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`, `make harness-contract` before any public-surface move | high | Raw helper scripts remain child commands unless Make adopts them. |
| Output modes and machine output | Harness NLSpec Sections 7 and 8; `tools/harness/core/test-output/*`; `tool-output.mjs` | Output class registry, schemas, summary builders | `tools/harness/core/tests/test-run-phase.sh`, `test-tool-output-real-targets.sh`, `test-harness-contracts.mjs` | Focused output-mode tests plus `harness-contract` | high | Must preserve one-JSON-object machine output where required. |
| Result roots, run IDs, retained artifact identity | Harness NLSpec Section 6; `harness-contract.mjs`; `artifact-discovery.mjs`; `artifact-writer.mjs` | Result-root normalization and secure writes | Core harness tests, agent-finalize tests | Retained artifact identity and permission fixtures | high | Custom result roots and owner-only permissions are observable. |
| Scheduler DAG, events, resources, summaries | Harness NLSpec Section 10; `tools/harness/scheduler/**` | `scheduler_manifest`, `scheduler_event`, summary schemas | Scheduler self-tests and service-backed scheduler tests | Scheduler determinism, resource, event-order, timing drift tests | high | Engine/adapters split is high risk. |
| Service and fixture lifecycle | Harness NLSpec Section 11; readiness/browser/scheduler service files | Service-backed suite mode, fixture policies, lifecycle records | Service-backed scheduler tests, dev service lifecycle tests, browser lifecycle tests | Owned/attach mode and cleanup lifecycle fixtures | high | Must preserve failure mapping and cleanup evidence. |
| Browser stack/reset behavior | Harness NLSpec Sections 11-13; browser scripts; Core 04 for product security outside test routes | Browser E2E wrappers, runtime reset, test route token, origin/host checks | Browser lifecycle tests, Playwright phase/batch tests | Browser lifecycle and runtime reset fixtures | high | Test-only routes must not become product route authority. |
| HTTP routes | Product route shapes are Core 01/Core 04; test-only runtime routes are harness Section 12 | No product route handlers under target; browser stack invokes product server and test reset route | Browser E2E and lifecycle tests | Product route tests remain outside this refactor; test-route fixtures required if browser lifecycle moves | medium | Refactor must not alter production HTTP route shapes. |
| WebSocket paths/events | Core 01/Core 03/Core 04 product owners | Target only starts browser/server fixtures; no product WebSocket implementation found | Browser E2E may observe WebSocket behavior | Existing product/browser tests must continue to pass | medium | No harness file should redefine event semantics. |
| Workbook row/query/mutation behavior | Core 01 through Core 03 and product modules/packages | No target source implements workbook mutation/query logic | Product/backend/frontend tests outside target | Preserve all existing product characterization tests | medium | Harness refactor should be invisible to workbook behavior. |
| Saved-view and view-schema behavior | Core 01/Core 03, contracts, generated packages | Target validates shapes and frontend row accounting only | JSON shape, frontend accounting tests | Shape/drift and frontend row-accounting tests | medium | Do not represent frontend evidence as behavior owner. |
| Projection refresh behavior | Core 01 and adopted projection docs for named scope | Target may schedule/check projection tests but does not own runtime projection code | Product/backend tests outside target | Existing projection tests plus boundary checks if scheduler paths move | medium | No production projection ownership observed. |
| Authorization checks | Core 04 for product auth; harness Section 12/15 for test-route token/redaction | Browser stack token/origin behavior, redaction helpers | Harness contract, browser lifecycle, static/security tests | Token/origin/redaction fixtures before moving core/browser/security files | high | Product authorization outcomes must not change. |
| Revision/change-set behavior | Core 02/Core 03 product owners | No revision implementation under target | Product tests outside target | Existing product revision tests only | low | Harness may run tests but does not own behavior. |
| Generated protocol/view contracts | Owner specs/contracts/generators; harness generated-artifact checks | `check-json-shapes`, `generate-drift`, generated-artifact policy | Generated-artifact self-tests | `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift` | high | Generated roots must not be hand-edited. |
| Grid-adapter or UI-selector contracts | `/packages/grid-adapter`, `/packages/ui-contracts`, frontend tests | Frontend import-boundary checker and test fixtures | Static-analysis frontend boundary tests | `make frontend-import-boundary-check` | medium | Test fixtures intentionally contain forbidden imports. |
| Harness/test accounting | Harness NLSpec, phase maps, task-surface/topology manifests | `test-output`, frontend row accounting, phase manifests, task execution map | Harness smoke, phase-slice, frontend unit, scheduler tests | `make check-harness-smoke`, `make harness-contract`, targeted self-tests | high | Accounting drift changes are observable harness behavior. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `tools/harness` is too broad to be treated as one permanent module. | Ten subtrees with distinct responsibilities and 224 tracked files. | A single refactor could mix scheduler, browser, readiness, static-analysis, and generated-artifact behavior. | `must_fix` | Multiple harness seams listed in Section 3. | Future implementation tasks must pick one seam or dependency chain at a time. |
| Public Make target behavior is the observable contract, not raw helper paths. | Harness NLSpec Sections 4.1 and 4.3; Makefile and task-surface manifests. | Moving raw scripts can still break public targets if generated metadata and wrappers drift. | `must_fix` | `harness/core-contract` and task-surface owner inputs. | Freeze `make help-all`, task-surface report, harness-contract, and generated drift before moves. |
| Scheduler core imports adapter-facing backend/browser/planning concepts. | `service-backed-schedule-cli.mjs`, `check-schedule-cli.mjs`, `check-service-backed-expansion.mjs`, runtime command helpers. | Engine changes can affect browser/service-backed scheduling semantics. | `should_fix` | `harness/scheduler-core` plus adapter modules. | Characterize scheduler events/resources before any split. |
| Browser stack lifecycle combines process/runtime service ownership with Playwright selection. | `start-web-e2e.sh`, `web-e2e-lifecycle.sh`, Playwright wrappers, browser shard plan. | Refactor can change ports, origins, test-route tokens, reset behavior, or cleanup. | `should_fix` | `harness/browser-runtime` and `harness/playwright-runner`. | Split only after lifecycle tests and browser target smoke are green. |
| Generated-artifact tooling is authored source but controls generated outputs. | `tools/harness/generated-artifacts/*` plus `tools/task_surface.generated.mk` and schedule/ledger outputs. | Hand-editing generated files or stale generated outputs can invalidate command surface. | `must_fix` | `harness/generated-artifact-tools`. | Plan owner-input edits first, then Make generators/drift checks. |
| Static-analysis test fixtures contain deliberately invalid imports. | `tools/harness/static-analysis/tests/test-frontend-import-boundaries.sh` embeds `react-data-grid` and generated import examples. | Automated coupling scans may report false production violations. | `intentional/no_action` | `harness/static-analysis`. | Classify fixtures as test-only when scanning. |
| Prior recovery/handoff docs contain historical harness paths. | `docs/testing-harness-spec-recovery-docs/**` and scripts refactor tracker mention old `scripts/*` paths. | Treating historical docs as live callers can create false blockers. | `intentional/no_action` | Documentation/archive owners. | Use active manifests and current code for caller truth; update historical docs only if republished. |
| Domain behavior is absent from target path. | No inspected file under `tools/harness` implements product route handlers, workbook mutations, saved views, projections, or record models. | A refactor plan might invent product-module moves from labels alone. | `intentional/no_action` | Product modules/packages outside `tools/harness`. | Record absence and keep product behavior preservation as a test/contract concern. |
| Harness cache/readiness artifacts are acceleration/support evidence only. | Harness NLSpec cache family requirements; readiness subtree. | Cache semantics could be overclaimed as product or release evidence. | `must_fix` | `harness/readiness`. | Keep validation language scoped to harness readiness; do not cite cache hits as product proof. |
| Authorization in target is test-route/redaction/security-harness only. | Browser runtime reset/test token and redaction/static security wrappers. | Could be confused with product auth if moved into product modules. | `should_fix` | Harness Section 12/15 owners plus Core 04 for product auth. | Keep test-only route authorization separate from production auth behavior. |

### Governance gap remediation matrix

| Gap | Remediation | Belongs in | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| RB-001 implementation authority ambiguity | Resolve tracker/supporting-plan authority, but keep code gated by one later selected slice ID plus RB-002 baseline. Explicitly require separate authorization for product behavior, public harness semantic, generated contract, route/auth, or bundled multi-slice changes. | Documentation now; implementation only later. | Prevents a planning closure from becoming implicit code authorization. | Smaller auditable slices and clearer future agent handoff. | No runtime impact. Tracker status changes from open-ended blocker to resolved-for-planning and gated-for-code. | Future work may modify behavior under vague scope. | Section 11 states RB-001 is resolved for tracker/supporting updates and gated for code. |
| RB-002 missing characterization baseline | Make S-01 mandatory before any movement and record public-surface, smoke, harness-contract, and seam-specific evidence with exit status, run/result root, failures, and blocker classification. | Documentation and test procedure. | Behavior-preserving refactors need pre-change evidence to detect harness drift. | Stable public Make/accounting behavior across future seam splits. | No immediate impact. A failing baseline blocks implementation until classified. | Public command IDs, output schemas, artifacts, scheduler events, or accounting can drift silently. | Section 7 lists baseline evidence; Section 8 lists Make commands; Section 11 marks RB-002 blocking for implementation. |
| RB-003 finalizer evidence open-ended | Make `make agent-finalize RESULTS_DIR=<dir>` conditional on affected surfaces: finalizer, cache/readiness reuse, scheduler timing, duration baselines, default `make check` topology, warm check behavior, public target summary emission, or final handoff maintenance evidence. | Documentation and test procedure. | Full warm finalizer evidence is important only for slices that touch those surfaces. | Precise evidence requirements without unnecessary broad runs. | Ordinary seam refactors may record N/A with reason. Affected slices need a successful full warm `make check` retained run root. | Teams may either skip necessary finalizer evidence or spend time on irrelevant broad evidence. | Sections 8 and 11 define required/N/A conditions. |
| RB-004 scheduler split boundary undefined | Add a scheduler split design contract defining `scheduler-core`, `scheduler-adapters`, `scheduler-contracts`, import rules, normalized plan input, unchanged event/summary outputs, failure mapping, and acceptance posture. | Documentation now; implementation/tests later. | Scheduler touches the broadest harness behavior and needs a durable seam before code movement. | Future phase expansion without turning scheduler core into a backend/browser/planning catch-all. | No immediate impact. S-03 later must preserve public scheduler contracts. | Core/adapters can become more coupled and accidentally alter public harness behavior. | Section 7 contains the design contract; Section 11 marks RB-004 design-resolved and S-03 code-gated. |
| Spec authority placement | Keep Core 00-04 unchanged unless product behavior changes; keep Core 05 unchanged unless claim-bearing timed or fixture-sensitive publication behavior changes; revise Testing Harness NLSpec only for public harness mechanics changes. | Specification and documentation gate. | Preserves owner precedence and prevents harness plans from redefining product behavior. | Cleaner spec evolution with less churn and fewer authority contradictions. | No tracker-only migration impact. Later public harness semantic changes become spec-first. | Future refactors may place behavior in the wrong owner document. | Section 1 records authority rules; each future slice must record whether a spec change is required. |
| Security/product boundary leakage | Record hard constraints that harness test-route tokens/origin/redaction/reset are not product auth, default runtime exposes no test routes, and harness must not own workbook, grid, projection, saved-view, product WebSocket, or product authorization behavior. | Documentation now; tests if later code touches these surfaces. | Keeps test support from becoming accidental product surface. | Safer harness expansion and clearer Core 04 product-auth boundary. | No immediate impact. Route/security-affecting slices need focused characterization and owner review. | Test-only assumptions can leak into production route/auth/reset/redaction behavior. | Sections 1, 4, 10, and 11 record constraints; affected future slices validate default route absence and token/origin/redaction behavior. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source bootstrap and tracker initialization | root | none | WF-01 | Create and maintain the tracker; record authority, live inventory, constraints, and output path. | This tracker only. | None for tracker creation; command-surface discovery only. | Tracker exists and states later implementation authorization. |
| WF-01 | Complete 224-file inventory | chain | WF-00 | WF-02, WF-04 | Keep every `tools/harness` file mapped to a seam and risk level. | `tools/harness/**`; manifests for callers. | Re-run `git ls-files tools/harness | sort` and owner-specific scans. | Inventory count and coverage list current. |
| WF-02 | Harness contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map public Make contracts, schemas, and owner docs to seams. | Harness NLSpec, Makefile, task-surface/topology manifests. | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`, `make harness-contract`. | Public contract freeze map current. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-06 | Decide which self-tests or public targets must pass before moving each seam. | Owner-local tests under `tools/harness/**/tests`. | `make check-harness-smoke` plus focused public targets. | Missing characterization rows marked `TODO:` or `BLOCKED:`. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Identify cross-seam imports, generated-file risks, raw-script assumptions, and test-only fixtures. | ESM import graph, shell `source` graph, manifests. | `make lint-scripts`, `make lint-shell`, static boundary targets as applicable. | Findings classified in Section 5. |
| WF-05 | Internal harness ownership redesign plan | chain | WF-02, WF-04 | WF-06 | Decide behavior-preserving seam boundaries inside harness without changing public contracts. | Core, scheduler, browser, backend, frontend, generated-artifacts subtrees. | Characterization commands selected by WF-03. | Proposed slices are one seam at a time. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07 | Define smallest safe behavior-preserving implementation slices. | Tracker and later implementation files. | Per-slice Make commands from Section 8. | Each slice has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Plan manifest/test/accounting updates, generated refreshes, and drift checks for later authorized implementation. | Task-surface/topology owner inputs, generated-artifacts tools, tests. | `make generated-artifact-policy-check`, `make json-shape-check`, drift targets. | Generated-output handling is Make-owned. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Record executed validation, skipped checks, blockers, and next slice. | Tracker and retained run artifacts. | `make agent-finalize RESULTS_DIR=<dir>` when a suitable full warm check run exists; broader gates as risk requires. | Another agent can continue without rediscovery. |

## 7. Proposed Refactor Slice Plan

### Baseline evidence required before edits

S-01 is mandatory, not advisory. No harness file movement, symbol split, import rewrite, generated-owner input edit, or public wrapper adjustment may begin until the following baseline is recorded:

| Evidence item | Required command or source | Must record | Blocks |
| --- | --- | --- | --- |
| Public command surface parity | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | Exit status, result root or artifact path when emitted, failures, and whether failures are pre-existing. | Any public Make/task-surface change. |
| Harness smoke | `make check-harness-smoke` | Exit status, result root or run root when emitted, failures, and affected seam. | Any harness implementation movement. |
| Extended harness contract | `make harness-contract` | Exit status, result root or run root when emitted, failures, and affected contract class. | Any change to output modes, result roots, summaries, artifacts, failure taxonomy, or accounting. |
| Seam-specific characterization | Selected command rows from Section 8 and the slice row below. | Command, exit status, retained evidence root if applicable, and blocker classification. | The selected implementation slice. |

If any required baseline command fails, implementation stops. The failure must be recorded as pre-existing, environment-related, or slice-blocking before files are moved.

### Scheduler split design contract for S-03

S-03 is design-closed for planning only; code remains gated by S-01/RB-002. The future split must preserve public Make targets, command IDs, output classes, accepted inputs, scheduler event order, summary schemas, resource names, retained artifact paths, result-root semantics, failure classes, and cleanup behavior unless a later owner-authorized spec change says otherwise.

| Component | Owns | Must not own | Interface contract |
| --- | --- | --- | --- |
| `scheduler-core` | DAG validation, work-unit state transitions, dependency ordering, logical resource arbitration, event emission inputs, summary assembly inputs, cancellation/interruption, and existing failure propagation. | Backend Go shard planning, browser stack startup, Playwright selection, frontend row accounting, product HTTP/WebSocket behavior, product authorization, generated-artifact generation, or phase-map interpretation beyond closed scheduler inputs. | Accept a normalized scheduler plan and emit the existing scheduler event and summary artifacts unchanged. |
| `scheduler-adapters` | Conversion from current topology, phase maps, browser helpers, backend helpers, frontend helpers, readiness helpers, and generated schedules into normalized scheduler plans. | Core scheduler state mutation after admission, ad hoc resource creation, public command reclassification, product behavior, schema changes, or failure remapping. | Produce closed scheduler plan objects and fail before execution when expansion is incomplete or ambiguous. |
| `scheduler-contracts` | Closed command descriptor shapes, resource claim names, event tokens, failure reason mappings, and validation helpers. | Product route schemas, WebSocket messages, view schemas, generated protocol contracts, Core 05 benchmark/publication manifests, or product authorization policy. | Provide the descriptor grammar shared by core and adapters only. |

Binary import rule for S-03:

```text
scheduler-core MUST NOT import from:
- tools/harness/backend/**
- tools/harness/browser/**
- tools/harness/frontend/**
- tools/harness/planning/**
- apps/**
- internal/modules/**
- packages/grid-adapter/**
- product route, auth, WebSocket, view-schema, or projection implementation paths

scheduler-adapters MAY import backend/browser/frontend/planning helpers only to compile a normalized scheduler plan.
```

Scheduler resources remain logical execution constraints. Cache/readiness hits must not skip summary emission, failure classification, drift, security, service readiness, cleanup, reset, or aggregate verdicts. Current profile rejects digest-only input stamp reuse for required correctness work.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/test-harness-module-refactor-tracker.md` | Documentation drift only. | Preserve tracker structure and all 224-file coverage. | Not run in this session; optional `make lint-markdown` if allowed by later verification scope. | Revert only this tracker file. | Tracker exists with all required sections. |
| S-01 | S-00 | Characterize public Make/task-surface behavior before moving any harness file. This slice is mandatory before implementation. | `Makefile`, task-surface/topology manifests, harness contract tests. | Public target names, command IDs, output classes, side effects. | Existing harness contract, harness smoke, and task-surface parity tests. | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`; `make check-harness-smoke`; `make harness-contract` | No code movement before baseline. | Baseline commands pass or blockers are recorded with result roots/run roots and failure classification. |
| S-02 | S-01 | Isolate `core` public harness contract utilities behind a smaller internal facade without changing public Make behavior. | `tools/harness/core/**` | Result roots, run IDs, output summaries, failure taxonomy, redaction, finalization. | Core harness self-tests and public wrapper smoke. | `make check-harness-smoke`; `make harness-contract` | Restore prior core files and manifests. | Public summaries and self-tests remain unchanged. |
| S-03 | S-01 | Separate scheduler engine from service/browser/backend expansion adapters. | `tools/harness/scheduler/**`, selected `backend`, `browser`, `planning` helpers | Scheduler event order, resources, priority, service-backed summary. | Scheduler and service-backed scheduler tests. | `make check-harness-smoke`; `make scheduler-event-order-drift RESULTS_DIR=<dir>` when retained evidence exists; `make scheduler-summary-timing-drift RESULTS_DIR=<dir>` when retained evidence exists | Revert adapter split as one slice. | Scheduler summaries/events match baseline. |
| S-04 | S-01 | Separate browser stack lifecycle from Playwright selection wrappers. | `tools/harness/browser/**`, related scheduler browser runtime helpers | Ports, origins, test-route token, runtime reset, cleanup, visual/a11y selection. | Browser lifecycle tests and Playwright wrapper tests. | `make browser-e2e-webserver-backed` or narrower public browser target chosen by WF-03; `make check-harness-smoke` | Revert lifecycle wrapper changes together. | Browser stack artifacts and target summaries preserve shape. |
| S-05 | S-01 | Split backend Go runner, duration baseline, and migration/schema drift responsibilities. | `tools/harness/backend/**` | Go sharding, shared report locks, fixture budgets, migration drift, security findings. | Backend harness self-tests and relevant drift checks. | `make check-harness-smoke`; `make migration-drift`; duration drift targets when retained evidence exists | Revert one backend seam at a time. | Go target summaries and drift diagnostics preserve behavior. |
| S-06 | S-01 | Split frontend evidence/readiness/design-token responsibilities. | `tools/harness/frontend/**` | Frontend row accounting, phase registry/map validation, accessibility summaries, generated design tokens. | Frontend harness self-tests, frontend unit target, JSON shape checks. | `make frontend-unit`; `make json-shape-check`; `make check-harness-smoke` | Revert frontend seam changes and regenerated outputs together. | Frontend accounting and readiness artifacts preserve schemas. |
| S-07 | S-01 | Preserve generator owner inputs and regenerate downstream artifacts only through Make. | `tools/harness/generated-artifacts/**`, owner manifests, generated outputs | Generated task surface, schedules, ledgers, JSON schema drift. | Generated-artifact self-tests and drift checks. | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift` | Revert owner input changes plus generated outputs as a coherent slice. | No generated drift remains and no generated file was hand-edited. |
| S-08 | S-01 | Preserve static-analysis, readiness, and test-support seams while reducing cross-path assumptions. | `tools/harness/static-analysis/**`, `readiness/**`, `test-support/**` | Boundary checks, toolchain cache/readiness, scratch outside repo, security scan summaries. | Static-analysis/readiness self-tests and public lint/security targets. | `make lint-shell`; `make lint-scripts`; `make backend-module-boundary-check`; `make frontend-import-boundary-check`; `make generated-artifact-policy-check` | Revert seam-specific changes; do not change product code. | Boundary/readiness summaries preserve public behavior. |

Any slice that changes product behavior, public Make command semantics, accepted inputs, output schemas, route shapes, authorization outcomes, generated contract surfaces, or more than one bundled implementation seam requires later authorization before implementation.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| public surface baseline | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | Public Make/task-surface parity and command metadata. | yes | Required before any harness movement or public command-surface edit. Not run in this tracker-only session. |
| unit | `make check-harness-smoke`; `make harness-contract` | Harness self-test smoke plus extended harness contract checks. | yes | Required before broad harness movement. Not run in this tracker-only session. |
| integration | `make test-fast` | Narrower local verification loop, including relevant service-backed/browser projections selected by current topology. | no | Required when implementation affects scheduler, service, or browser runtime. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make browser-e2e-visual`; `make browser-e2e-a11y` as applicable | Browser public target behavior. | no | Choose only affected browser layer; visual/a11y full targets are explicit and not default local check evidence. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift`; `make phase-ledger-drift`; `make phase-schedule-drift` | Generated artifact policy, JSON shapes, generated outputs, phase ledgers, schedules. | yes when generated-artifact owner inputs change | Use Make generators/drift targets only; do not hand-edit generated outputs. |
| import-boundary/static | `make lint-shell`; `make lint-scripts`; `make backend-module-boundary-check`; `make frontend-import-boundary-check` | Shell/script lint and backend/frontend boundary guardrails. | yes when affected files move or imports change | Test fixtures with intentional forbidden imports must remain classified as test-only. |
| full check | `make check` | Developer verification gate. | no | Broaden to `make check` when risk spans multiple seams or public command behavior. |
| handoff maintenance | `make agent-finalize RESULTS_DIR=<successful full warm check run root>` | Retained-run maintenance and finalizer cache evidence. | conditional | Required only when a slice affects `agent-finalize`, cache/readiness reuse, scheduler timing, duration-baseline drift, default `make check` topology, `check-service-backed` warm-run behavior, public target summary emission, or final handoff maintenance evidence. Otherwise record `RB-003=N/A; reason=<why finalizer evidence is not relevant to this slice>`. |

Commands are discovered from `AGENTS.md`, `make help`, `make help-all`, and `docs/testing-harness-nlspec.md`. No Make validation target was run in this tracker-only session.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TH-001 | Create target-specific tracker file | WF-00 | DONE | none | `docs/handoffs/test-harness-module-refactor-tracker.md` | Tracker exists and records planning-only status. |
| TH-002 | Confirm target key and target existence | WF-00 | DONE | none | `test-harness`; `git ls-files tools/harness` count 224 | Target key normalized and path exists. |
| TH-003 | Inventory current harness files by seam | WF-01 | DONE | TH-002 | Section 2 coverage list | Every tracked file under `tools/harness` is listed. |
| TH-004 | Map harness contracts and behavior freeze | WF-02 | DONE | TH-003 | Section 4 freeze map | Observable harness contracts have owners and test posture. |
| TH-005 | Record boundary diagnosis | WF-04 | DONE | TH-003 | Sections 3 and 5 | Target classified as mixed harness subsystem, not product module. |
| TH-013 | Close RB-001 through RB-004 governance gaps | WF-00 | DONE | TH-001 | Sections 1, 7, 8, 10, 11, and 12 | Authorization, baseline, conditional finalizer, and scheduler split rules are binary. |
| TH-006 | Establish characterization baseline | WF-03 | TODO | TH-004 | Planned Make commands | Required commands run or blockers recorded before implementation. |
| TH-007 | Plan core contract facade split | WF-05 | TODO | TH-006 | S-02 | Behavior-preserving edit sequence defined in later task. |
| TH-008 | Plan scheduler core/adapter split | WF-05 | TODO | TH-006 | S-03 | Scheduler characterization complete. |
| TH-009 | Plan browser lifecycle/Playwright split | WF-05 | TODO | TH-006 | S-04 | Browser lifecycle characterization complete. |
| TH-010 | Plan backend/frontend/generated/static/readiness slices | WF-06 | TODO | TH-006 | S-05 through S-08 | Each seam has validation and rollback criteria. |
| TH-011 | Run generated/static/harness validation for future implementation | WF-07 | TODO | TH-010 | Make target artifacts | Commands pass or failures are triaged. |
| TH-012 | Finalize later implementation handoff | WF-08 | TODO | TH-011 | Updated session log | Next agent can continue without rediscovery. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Tracker created from live target inventory and owner docs. | Touched: `docs/handoffs/test-harness-module-refactor-tracker.md`. Inspected: `AGENTS.md`, planning framework, harness NLSpec, domain doc, Core 00-04, Makefile, manifests, `tools/harness/**`. | `git status --short --branch`; `git ls-files tools/harness`; `make help`; `make help-all`; targeted `sed`/`rg` inspections. | Target exists with 224 tracked files; output tracker did not previously exist. | Implementation not authorized in this task. | Use this tracker to pick one future harness seam and run characterization first. |
| 2026-07-03T19:47:24Z | Codex / RB closure | RB-001 is resolved for tracker/supporting-plan updates; code remains gated by a selected slice ID and RB-002 baseline. | Touched: `docs/handoffs/test-harness-module-refactor-tracker.md`. Inspected: current tracker sections for authority, slices, validation, blockers, and handoff. | `rg -n ... docs/handoffs/test-harness-module-refactor-tracker.md`; `sed -n ...`; `git status --short -- docs/handoffs/test-harness-module-refactor-tracker.md`; `date -u +%Y-%m-%dT%H:%M:%SZ`. | Added binary authority, spec-gate, security/product-boundary, baseline, conditional finalizer, and scheduler split rules. | RB-002 still blocks implementation until characterization runs. | Run S-01 before any implementation slice. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Backend harness files are inventory-only; no production backend module movement planned. | Inspected: `tools/harness/backend/**`; touched tracker only. | `git ls-files tools/harness`; targeted import/export inventory. | Backend harness owns Go runner/duration/migration/schema checks, not product domain logic. | Need later characterization before moving Go runner files. | Start with S-05 only after S-01 baseline. |
| 2026-07-03T19:47:24Z | Codex / RB closure | Backend implementation remains unchanged; S-05 remains a later behavior-preserving slice gated by S-01 and the spec gate. | Touched tracker only. No backend source files inspected or edited in this closure pass. | Tracker `rg`/`sed` inspections only. | No backend module boundary decision changed. | RB-002 baseline required before S-05. | Keep backend Go runner/duration/migration changes isolated to one authorized slice. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Frontend harness files are readiness/evidence accounting; no app shell/controller refactor planned. | Inspected: `tools/harness/frontend/**`, browser/frontend test references; touched tracker only. | Targeted import/export inventory and Make help discovery. | No ownership of `/apps/web` shell state or grid vendor integration observed. | Frontend row-accounting/a11y/design-token schemas need characterization if moved. | Use S-06 after S-01 baseline. |
| 2026-07-03T19:47:24Z | Codex / RB closure | Frontend implementation remains unchanged; S-06 remains a later behavior-preserving slice gated by S-01 and the spec gate. | Touched tracker only. No frontend source files inspected or edited in this closure pass. | Tracker `rg`/`sed` inspections only. | No frontend shell/controller or grid-adapter ownership was added to harness. | RB-002 baseline required before S-06. | Keep frontend evidence/readiness/design-token work separate from product UI behavior. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Generated-artifacts subtree is authored tooling; generated outputs remain downstream. | Inspected: `tools/harness/generated-artifacts/**`, `tools/task_surface.generated.mk`, task-surface/topology manifests; touched tracker only. | `rg` searches for harness/generated references; `make help-all` discovery. | Must use Make generators/drift targets for any future generated-output refresh. | No generated file edits authorized. | Use S-07 only in a later implementation task. |
| 2026-07-03T19:47:24Z | Codex / RB closure | Core 00-04, Core 05, and Testing Harness NLSpec remain unchanged for tracker-only closure; future public harness semantic changes require owner review first. | Touched tracker only. No generated files, specs, contracts, or manifests edited. | Tracker `rg`/`sed` inspections only. | Added explicit spec-gate and no-hand-edit generated-output rules. | Any future public harness semantic change must revise owner inputs/specs before implementation. | For S-07, edit owner inputs only and regenerate downstream artifacts through Make. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Harness self-tests are owner-local under `tools/harness/**/tests`; no tests changed. | Inspected all test paths in inventory; touched tracker only. | `git ls-files tools/harness`; prior task-surface/topology reference scans. | Existing tests provide characterization candidates; no validation run claimed. | Characterization baseline not executed in this session. | Run S-01 commands before implementation. |
| 2026-07-03T19:47:24Z | Codex / RB closure | S-01 is now a hard pre-implementation gate with required public-surface, smoke, and harness-contract commands. | Touched tracker only. No tests or harness source files edited. | Tracker `rg`/`sed` inspections only. | Added required baseline evidence fields for exit status, result/run roots, failures, and seam-specific commands. | Baseline commands were not run in this tracker-only closure pass. | Run and record `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`, `make check-harness-smoke`, and `make harness-contract`. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Product authorization remains Core 04-owned; target has harness redaction/security and test-route mechanics only. | Inspected Core 04, harness NLSpec Sections 12/15, browser/static-analysis paths; touched tracker only. | Targeted `sed` and `rg` inspections. | No product auth implementation under target. | Moving browser/static security files needs token/origin/redaction characterization. | Keep test-only route security separate from product auth in future slices. |
| 2026-07-03T19:47:24Z | Codex / RB closure | Security boundary constraints are explicit: harness test-route mechanics are not product auth, and default runtime must not expose test routes. | Touched tracker only. No product routes, auth code, browser runtime, or tests edited. | Tracker `rg`/`sed` inspections only. | Added route/auth/product-boundary constraints and future validation posture. | Any route/security-affecting slice needs focused owner review and characterization. | Validate test-route absence/default behavior only in later authorized implementation work. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T19:28:09Z | Codex / tracker creation | Tracker is current enough for later implementation planning. | Touched tracker only. | Discovery commands listed above; no Make validation target run. | No production refactor performed. | Later work needs explicit implementation authorization and characterization baseline. | Begin with S-01, then pick one seam such as S-02 or S-03. |
| 2026-07-03T19:47:24Z | Codex / RB closure | RB-001 and RB-004 are closed for planning; RB-002 remains the hard implementation baseline; RB-003 is conditional by slice. | Touched tracker only. | Tracker `rg`/`sed` inspections only; no Make validation target run. | Remaining blockers are now binary and slice-specific. | Implementation remains blocked until S-01 passes; finalizer evidence is required only when RB-003 applies. | Choose one later slice, run S-01, then apply the relevant validation plan. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Code implementation authority is gated. | Tracker/supporting-plan updates are authorized, but code work must not proceed under ambiguous scope. | Later user authorization naming exactly one behavior-preserving implementation slice ID, plus passing RB-002 baseline. | RESOLVED for tracker/supporting update; code implementation requires selected slice ID and RB-002 baseline. |
| RB-002 | Characterization baseline has not been run in this session. | Refactoring harness internals without baseline risks public Make/output/scheduler drift. | `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'`, `make check-harness-smoke`, `make harness-contract`, and seam-specific commands from Section 8. | BLOCKING for implementation; not blocking tracker revision. |
| RB-003 | Finalizer evidence is conditional by slice. | Retained-run finalizer evidence validates timing/maintenance artifacts only when the selected slice affects those surfaces. | A successful full warm `make check` retained run root and `make agent-finalize RESULTS_DIR=<dir>` only when the slice affects `agent-finalize`, cache/readiness reuse, scheduler timing, duration baselines, default `make check` topology, `check-service-backed` warm-run behavior, public target summary emission, or final handoff maintenance evidence. | CONDITIONAL; otherwise record `RB-003=N/A; reason=<why finalizer evidence is not relevant to this slice>`. |
| RB-004 | Scheduler engine/adapter split boundary needed focused design before edits. | Scheduler imports backend/browser/planning helpers and has public event/resource contracts. | Section 7 scheduler split design contract plus scheduler characterization tests and import graph for selected S-03 work. | RESOLVED for design in Section 7; S-03 code remains gated by RB-002. |

## 12. Binary Completion Criteria

| Criterion | Current status |
| --- | --- |
| Every file in `tools/harness` is inventoried or explicitly out of scope. | DONE. All 224 tracked files are listed in Section 2. No tracked file is out of scope. |
| Every discovered public contract risk has an owner and test posture. | DONE. Section 4 maps owners, evidence, existing tests, and characterization posture. |
| Every proposed workflow has dependencies and exit criteria. | DONE. Section 6 defines dependencies and handoff checkpoints. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | DONE. Section 7 marks behavior changes as requiring later authorization. |
| Validation commands are discovered or marked `TODO` with a reason. | DONE. Section 8 lists Make-owned commands and states they were not run in this tracker-only session. |
| Contradictions are marked `BLOCKED: owner contradiction`. | DONE. No owner contradiction was found. |
| Repository/framework mismatches are recorded as planning findings. | DONE. The framework's module catalog is treated as doctrine; live `tools/harness` state is recorded as a mixed harness subsystem, not a product module. |
| Handoff sections are current enough for another agent to continue without rediscovery. | DONE. Section 10 records scope, authority, inspected files, commands, results, blockers, and next action. |
| RB-001 states tracker/supporting update authorization exists while code remains gated. | DONE. Section 11 resolves RB-001 for tracker/supporting updates and requires a selected slice ID plus RB-002 baseline for code. |
| RB-002 requires S-01 baseline evidence before code movement. | DONE for tracker gate; BLOCKING for implementation until required commands and result roots/failures are recorded. |
| RB-003 is conditional rather than open-ended. | DONE. Sections 8 and 11 state when `make agent-finalize RESULTS_DIR=<dir>` is required and how to record N/A. |
| RB-004 has a scheduler split design contract. | DONE. Section 7 defines `scheduler-core`, `scheduler-adapters`, `scheduler-contracts`, import rules, preservation requirements, and acceptance posture. |
| Core 00-04 and Core 05 remain unchanged unless their owner scopes are affected. | DONE. Section 1 records product and publication authority boundaries. |
| Testing Harness NLSpec changes are required only for public harness semantic changes. | DONE. Section 1 records this as a pre-implementation spec gate. |
| No generated artifact is hand-edited. | DONE for this tracker task. Generated outputs remain downstream of owner inputs and Make generators. |
| Public Make target names, command IDs, output/artifact schemas, failure mapping, scheduler schemas, route/auth behavior, and product contracts are unchanged. | DONE for this tracker task. Any future change requires owner-spec or user authorization. |

No production refactor was performed by this tracker task.
