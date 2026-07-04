# test-harness-scheduler Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `tools/harness/scheduler`
- Target label: `test-harness-scheduler`
- Normalized label: `test-harness-scheduler`
- Output path: `docs/handoffs/test-harness-scheduler-module-refactor-tracker.md`
- Current status: implementation refactor authorized and performed after the original planning-only tracker session; this file now records both the original plan and the implementation handoff.
- Allowed changes for the implementation session: harness scheduler refactor, owner-facade/spec/test updates, and generated task-surface/topology refresh through Make-owned generation.
- Non-goals: no product HTTP/WebSocket/workbook behavior changes, no product Core/subsystem behavior changes, no dependency/package configuration edits, no migrations, and no duration baseline content refresh without retained-run validation.
- Later authorization required: duration baseline value updates or public product behavior changes still require a separate owner-backed task.
- Target existence: `tools/harness/scheduler` exists and now contains the scheduler core, scheduler public CLIs/facades, browser scheduler adapter, and scheduler shell tests; phase-slice planning, service-backed planning, duration-accounting, execution-topology, process execution, and diagnostics helpers were moved to clearer harness owner paths.
- Tracker existence before this session: not present.

Source hierarchy used:

1. Adopted subsystem NLSpecs, for their named subsystem only.
2. Core 00 through Core 04, for implementation-conformance behavior.
3. Core 05, only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides, for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests, for current implementation state.
6. Prior plans, handoffs, and framework files, as evidence only.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`

Repository files inspected:

- `Makefile`
- `tools/task_surface_manifest.json`
- Existing related handoffs under `docs/handoffs/`, including the browser and backend harness refactor trackers.
- All files under `tools/harness/scheduler`, listed in the inventory below.

Source posture findings:

- `docs/testing-harness-nlspec.md` owns harness mechanics for command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, and verification gates.
- `docs/domain.md` treats test harness terms as implementation-support terms, not product-domain module terms.
- Core 00 through Core 04 are product behavior authority but do not make `tools/harness/scheduler` a product runtime module.
- The planning framework is doctrine for this tracker, not proof of current repository state.
- Implementation authorization from the follow-up request resolves the original planning-only implementation blocker (`RB-002`).
- Phase maps, task rows, and scheduler evidence are verification/accounting surfaces, not runtime product architecture.
- No owner contradiction was found during this planning session.

Implementation remediation summary:

- Kept public harness contracts stable: Make target names, `command_id`s, schema IDs, retained scheduler artifact paths, failure mapping, cleanup behavior, and public inputs.
- Removed scheduler-local shallow backend and schedule-context adapters; scheduler callers now import the owning backend/execution facades directly.
- Moved phase-slice planning to `tools/harness/phase-accounting`.
- Moved service-backed schedule expansion, manifest validation, and topology planning to `tools/harness/execution/service-backed`.
- Moved shared execution dependency metadata access to `tools/harness/execution/execution-dependencies.mjs`.
- Moved duration baseline/drift helpers and CLIs to `tools/harness/duration-accounting`.
- Moved retained scheduler drift CLIs to `tools/harness/diagnostics`.
- Moved child process execution/redaction/log handling to the scheduler process facade `tools/harness/scheduler/process-executor.mjs`.
- Promoted `cartulary.scheduler_pressure_summary.v1` to a present diagnostic schema attachment and validates pressure summaries before scheduler success.

## 2. Current-State Repository Inventory

Implementation note: the table below was the planning baseline used to drive the remediation. Current path ownership after the implementation is:

- Scheduler core/facades remain under `tools/harness/scheduler`: `check-schedule-cli.mjs`, `service-backed-schedule-cli.mjs`, `scheduler-cli.mjs`, `scheduler-browser-runtime.mjs`, `scheduler-{manifest,resources,reporting,runner}.mjs`, `check-schedule-manifest.mjs`, `process-executor.mjs`, `adapters/browser.mjs`, `scheduler/*.mjs`, and `tests/*.sh`.
- Removed scheduler-local re-export adapters: `tools/harness/scheduler/adapters/backend.mjs` and `tools/harness/scheduler/adapters/schedule-context.mjs`.
- New owner paths: `tools/harness/phase-accounting/phase-slice-{cli,plan}.mjs`, `tools/harness/execution/execution-dependencies.mjs`, `tools/harness/execution/service-backed/{schedule-expansion,schedule-manifest,schedule-planning,schedule-topology}.mjs`, `tools/harness/duration-accounting/*`, and `tools/harness/diagnostics/scheduler-*-drift-cli.mjs`.
- Intentional old scheduler-local helper paths are now `unsupported_private` in `docs/testing-harness-nlspec.md` and `tools/harness/static-analysis/harness-import-boundary.mjs`.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tools/harness/scheduler/adapters/backend.mjs` | Thin backend shard-plan facade for scheduler callers. | Re-exports `collectGoShardPlan`, `collectGoShardsForTarget`. | `phase-slice-plan.mjs`, `service-backed-schedule-cli.mjs`, same-package tests through CLIs. | `tools/harness/backend/backend-shard-plan.mjs`. | `test-run-phase-slice.sh`, `test-service-backed-scheduler.sh`. | Backend target-plan and scheduler summary accounting. | Backend shard-plan facade or scheduler adapter. | medium | Keep as facade unless owner docs move backend planning access elsewhere. |
| `tools/harness/scheduler/adapters/browser.mjs` | Declared browser scheduler adapter facade. | Re-exports browser command, dependency, worker-slot, session, and env helpers. | Scheduler CLIs, expansion/topology helpers, tests, related browser handoff. | `tools/harness/browser/browser-batch-manifest.mjs`, `tools/harness/browser/browser-scheduler-dependencies.mjs`. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | Browser stage topology, worker-slot env, service-backed summaries. | Browser scheduler adapter facade. | high | Harness NLSpec names this as the current `browser_scheduler_adapter` owner facade. |
| `tools/harness/scheduler/adapters/schedule-context.mjs` | Schedule-context facade for summary topology and target descriptors. | Re-exports `buildSummaryTopology`, `createRunnerContext`, `findTargetDescriptor`. | Scheduler CLIs and expansion helpers. | `tools/harness/generated-artifacts/execution-topology.mjs`, `tools/harness/backend/backend-target-plan.mjs`. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | Execution topology and target-summary context. | Schedule-context adapter facade. | medium | Candidate to keep as explicit dependency boundary. |
| `tools/harness/scheduler/check-schedule-cli.mjs` | Check scheduler CLI for Make-owned `check` execution. | Executable CLI surface; schema constants for scheduler event and check summary. | `Makefile`, `tools/task_surface_manifest.json`, harness contract context, tests. | Scheduler manifest/resources/reporting/engine, runner context, browser adapter, test output, service lifecycle. | `test-check-scheduler.sh`. | `cartulary.scheduler_manifest.v1`, `cartulary.scheduler_event.v6`, `cartulary.check_scheduler_summary.v10`, retained scheduler artifacts. | Harness scheduler CLI facade. | high | Public behavior includes exit codes, dry-run, stop-on-first-failure, child summary propagation, service/browser lifecycle. |
| `tools/harness/scheduler/check-schedule-manifest.mjs` | Validates check scheduler manifest shape. | `validateCheckScheduleManifestShape`, `checkScheduleSchemaID`. | Generated JSON shape checks and scheduler tests. | Generated artifact contract helpers. | `test-check-scheduler.sh`. | Check schedule schema, generated JSON shape validation. | Harness scheduler manifest validation. | medium | Validation compatibility surface; do not hand-edit generated inputs. |
| `tools/harness/execution/service-backed/schedule-planning.mjs` | Expands service-backed schedule sources into normalized scheduler work units for `check` and direct service-backed schedules. | `mapServiceBackedClaimsToCheckClaims`, `expandServiceBackedScheduleForCheck`, `expandServiceBackedSchedule`. | `check-schedule-cli.mjs`, `service-backed-schedule-cli.mjs`, topology tests. | Browser adapter, schedule-context adapter, service-backed manifest, scheduler resources and manifest. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`. | Scheduler manifest work-unit shape, service-backed schedule sources, browser group/session artifacts. | Service-backed expansion planner. | high | Mixed scheduler, topology, browser, backend, and generated-artifact responsibilities. |
| `tools/harness/scheduler/duration-baseline-cli.mjs` | Shared CLI argument/context helper for duration baseline commands. | `durationBaselineCliContext`, `parseDurationBaselineResultsArgs`. | Duration baseline CLIs. | Node path/url/process helpers. | Duration baseline shell tests through CLIs. | Duration baseline command arguments only. | Duration accounting CLI support. | medium | Small shared helper with duration ownership. |
| `tools/harness/duration-accounting/duration-baseline-drift-suite.sh` | Shell wrapper for duration baseline drift suite. | Make-invoked shell entrypoint. | `tools/task_surface_manifest.json`, Make target `duration-baseline-drift-suite`. | Make targets for Go, browser, service-backed, and harness-smoke duration drift. | Task-surface and harness smoke coverage. | Duration baseline drift evidence. | Duration accounting harness wrapper. | medium | Requires `RESULTS_DIR` for retained-run use. |
| `tools/harness/scheduler/duration-drift.mjs` | Shared duration drift thresholding and service timing contamination reporting. | `defaultDurationDriftThresholds`, `formatRatio`, `formatSignedMs`, `durationDriftKind`, `durationDriftDescription`, `collectServiceTimingContamination`, `formatContaminationReasons`, `printContaminationReasons`. | Duration baseline CLIs. | Filesystem JSON summaries, retained result trees. | Duration baseline shell tests. | Retained-run duration drift and contamination diagnostics. | Duration accounting. | high | Core retained-run safety logic for timed evidence. |
| `tools/harness/scheduler/execution-dependencies.mjs` | Facade over generated execution dependency metadata. | `executionDependencyMetadata`, `validExecutionDependencies`, `validSupportTargets`, `serviceBackedGoExecutionDependencies`, `serviceBackedSupportTargets`, `executionDependencyInfo`, `compareExecutionDependencies`, `targetForExecutionDependency`. | Phase and service-backed planning helpers. | Generated execution topology helpers. | Scheduler shell tests through CLIs. | Execution topology dependency metadata. | Execution-topology facade. | medium | Boundary should remain a facade over generated data, not hand-edit generated data. |
| `tools/harness/duration-accounting/harness-smoke-durations-cli.mjs` | Maintains and checks harness-smoke duration baselines. | Executable CLI surface. | `tools/task_surface_manifest.json`, Make targets for harness-smoke duration baseline update/drift. | Duration CLI helper, target-duration baselines, task-surface manifest, retained target summaries. | `test-harness-smoke-duration-baselines.sh`. | `cartulary.harness_smoke_duration_baselines.v1`, harness smoke fast-tier accounting. | Duration accounting. | high | `update` mode mutates baseline inputs and requires later authorization. |
| `tools/harness/phase-accounting/phase-slice-cli.mjs` | CLI for `phase-slice` and `service-backed-slice` scheduler execution. | Executable CLI surface; phase-slice summary schema constants. | `tools/task_surface_manifest.json`, phase-accounting tests, Make targets. | `phase-slice-plan.mjs`, scheduler engine, test output, runtime command helpers, service wrapper. | `test-run-phase-slice.sh`. | `cartulary.phase_slice_plan.v1`, `cartulary.phase_slice_scheduler_summary.v4`, scheduler events. | Phase-slice scheduler facade. | high | Mixes phase accounting with scheduler execution and service-wrapper orchestration. |
| `tools/harness/phase-accounting/phase-slice-plan.mjs` | Builds phase-slice plans from phase registry, task-surface, backend target plan, frontend phase plan, browser stages, and execution topology. | `repoRoot`, `phaseSlicePlanSchemaID`, `buildPhaseSlicePlan`, `validateAllPhaseSlicePlans`, `printablePlan`. | `phase-slice-cli.mjs`, task-surface manifest, tests. | Phase registry, phase accounting, frontend phase plan, backend target plan, browser adapter, execution topology. | `test-run-phase-slice.sh`. | `cartulary.phase_slice_plan.v1`, phase ledgers/maps as evidence accounting. | Phase-accounting or phase-slice planning facade. | high | Architectural finding: this is not product runtime architecture. |
| `tools/harness/scheduler/scheduler-browser-runtime.mjs` | Builds runtime command for browser stage completion and optional session stop. | `browserStageCompleteRuntimeCommand`. | Scheduler CLIs. | Shell command construction, browser session env files. | Browser lifecycle assertions in scheduler tests. | Browser stage complete command shape and target-summary behavior. | Browser scheduler runtime adapter. | high | Shell command shape is observable through retained artifacts and failure logs. |
| `tools/harness/scheduler/scheduler-cli.mjs` | Parses common scheduler runner CLI options. | `parseSchedulerRunnerArgs`. | `check-schedule-cli.mjs`, `service-backed-schedule-cli.mjs`, other scheduler runners. | Node process args only. | Scheduler CLI tests through runners. | Public/private CLI option compatibility. | Harness scheduler CLI support. | medium | Keep stable for Make-owned wrappers. |
| `tools/harness/diagnostics/scheduler-event-order-drift-cli.mjs` | CLI for retained scheduler event ordering drift checks. | Executable CLI surface. | Make target `scheduler-event-order-drift`, task surface. | `scheduler/event-order.mjs`. | Event-order cases in `test-check-scheduler.sh`. | `cartulary.scheduler_event.v6` ordering validation. | Scheduler event drift checker. | medium | Requires retained run root for meaningful end-to-end use. |
| `tools/harness/scheduler/scheduler-manifest.mjs` | Validates and normalizes scheduler manifest, commands, resources, dependencies, finalizers, and browser worker-slot ranges. | `schedulerManifestSchemaID`, `schedulerCommandTypeValues`, `schedulerCommandShapes`, `validateSchedulerCommandShape`, `validateSchedulerManifestShape`, `parseResourceLimitOverride`, `loadSchedulerManifest`, `loadScheduleManifest`, `selectSingleSchedule`, `normalizeSchedulerSchedule`, `normalizeStringList`, `normalizeNeeds`, `validateTargetDependencyGraph`. | Scheduler CLIs, JSON shape checks, task-surface topology checks, tests. | Scheduler resources, generated JSON shape validation, runtime command shapes. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | `cartulary.scheduler_manifest.v1`, command-shape closure, worker-slot ranges. | Harness scheduler manifest core. | high | Major public compatibility surface. |
| `tools/harness/scheduler/scheduler-reporting.mjs` | Scheduler output formatting, retained paths, progress and summary lines. | Many reporting helpers including `schedulerOutputMode`, `schedulerTargetDir`, `schedulerLogDir`, progress/summary/dry-run line builders. | Scheduler engine and CLIs. | Resource formatting and retained artifact path conventions. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | stdout/stderr contract, retained scheduler directories, progress summary. | Harness scheduler reporting core. | high | Output text and artifact paths are harness compatibility surfaces. |
| `tools/harness/scheduler/scheduler-resources.mjs` | Scheduler resource registry validation, resource limits, auto estimators, claims, forwarding profiles. | Registry/resource APIs including `validateSchedulerResourceRegistryShape`, `schedulerResourceRegistry`, `normalizeResourceLimits`, `resolveAutoResourceLimits`, auto estimators, claim normalizers. | Scheduler CLIs, manifest normalization, generated artifact checks, task-surface generation. | `tools/scheduler_resource_registry.json`, OS CPU info, browser stage resource naming. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | `cartulary.scheduler_resource_registry.v4`, task-surface resource env variables, execution topology. | Harness scheduler resource core. | high | Auto capacity and resource claim behavior are scheduler-owned contracts. |
| `tools/harness/scheduler/scheduler-runner.mjs` | Public facade over scheduler engine. | Re-exports engine runner and dry-run APIs. | `phase-slice-cli.mjs`, `test-run-phase-slice.sh`, possible external harness callers. | `scheduler/engine.mjs`. | `test-run-phase-slice.sh`. | Scheduler run behavior and dry-run output. | Harness scheduler facade. | medium | Useful stable facade before any internal file movement. |
| `tools/harness/diagnostics/scheduler-summary-timing-drift-cli.mjs` | CLI for scheduler summary timing drift and warm-check health checks. | Executable CLI surface. | Make target `scheduler-summary-timing-drift`, task surface. | `scheduler/summary-timing-drift.mjs`, retained summaries, scheduler accounting extensions. | Timing drift cases in `test-check-scheduler.sh`. | Scheduler summaries, target summaries, warm-check budgets, accounting extensions. | Scheduler timing drift checker. | high | Core retained-run safety and CI/finalize evidence gate. |
| `tools/harness/scheduler/scheduler/blockers.mjs` | Blocked resource/dependency diagnostics and top-blocker summaries. | `schedulerBlockedDiagnostics`, `addTopBlockerObservations`, `topBlockerRecords`, `formatTopBlockers`. | Scheduler engine/reporting. | Scheduler state snapshots. | Scheduler artifact tests. | Pressure summary and human progress diagnostics. | Harness scheduler diagnostics core. | medium | Diagnostic posture must stay non-product-conformance evidence. |
| `tools/harness/scheduler/scheduler/clock.mjs` | Monotonic scheduler clock and wall timestamp helper. | `SchedulerClock`. | Scheduler engine. | Node performance clock and Date. | Event-order and timing drift tests. | Scheduler event timing fields and clock-skew marker behavior. | Harness scheduler timing core. | medium | Event ordering depends on monotonic and wall timestamps. |
| `tools/harness/scheduler/scheduler/engine.mjs` | Runs normalized schedules, emits events, writes retained artifacts, manages dependency/resource execution and summaries. | Re-exports process executor helpers; `finalizerRunningDisplayUnits`, `countVisibleCompletedUnit`, `replayFailedAggregateLogsBeforeFinalizer`, `runNormalizedSchedule`, `writeSchedulerDryRun`. | `scheduler-runner.mjs`, scheduler CLIs, tests. | Scheduler state/reporting/resources, process executor, progress recorder, blockers, summary timing validation, test-output lifecycle. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | Scheduler events, summaries, pressure summary, progress logs, failure mapping. | Harness scheduler execution core. | high | Highest-risk behavior-preserving refactor area. |
| `tools/harness/scheduler/scheduler/event-order.mjs` | Finds retained event files and validates scheduler event ordering. | `schedulerEventFiles`, `validateSchedulerEventOrderFile`, `validateSchedulerEventOrder`. | Event-order drift CLI, timing tests. | Filesystem retained scheduler event JSONL. | `test-check-scheduler.sh`. | `cartulary.scheduler_event.v6`. | Scheduler event drift checker. | medium | Must preserve clock-skew exception semantics. |
| `tools/harness/scheduler/scheduler/process-executor.mjs` | Child process execution, dry-run detection, env redaction, log replay. | `isDryRunFromMakeFlags`, `sanitizeMakeFlags`, `makeChildEnv`, `sanitizeLogName`, `runLifecycle`, `runCommand`, `replayLog`. | Scheduler engine and CLIs. | Node child process, logs, environment variables. | Scheduler failure/log tests. | Child log files, redaction and failure output. | Harness scheduler process adapter. | high | Security-sensitive for env/log leakage. |
| `tools/harness/scheduler/scheduler/progress-recorder.mjs` | Records progress snapshots and slowest-running observations. | `SchedulerProgressRecorder`. | Scheduler engine. | Progress summary file and in-memory observations. | Scheduler artifact tests. | `progress-summary.log`, scheduler summary slowest units. | Harness scheduler diagnostics core. | medium | Progress bounds and summary shape are observable. |
| `tools/harness/scheduler/scheduler/runtime-command-helpers.mjs` | Runtime command helpers for browser sessions, test output, env files, and runner manifests. | `browserSessionKeyFor`, `readStringEnvFile`, `loadSchedulerRunnerManifest`, `browserSessionFilesFor`, `testOutputRuntimeCommand`, `browserSessionStartCommand`, `browserStageCompleteCommand`, `browserSessionFinalizerCommand`. | Scheduler CLIs and expansion helpers. | Browser session env files, runner manifests, test-output command surface. | Browser/session scheduler tests. | Browser session artifacts, lifecycle command shape. | Scheduler runtime command adapter. | high | Runtime binary injection and browser lifecycle are NLSpec-owned behavior. |
| `tools/harness/scheduler/scheduler/state.mjs` | Scheduler dependency, priority, resource, skip propagation, and blocked-state calculations. | State helpers including `readyPendingUnits`, `hasResourceCapacity`, `priorityAdmissiblePendingUnitIndex`, `skipFailedDependencyUnits`, `blockedSnapshot`. | Scheduler engine and diagnostics. | Resource claims and dependency graph inputs. | Scheduler semantic tests. | Event/progress state, pressure diagnostics, skipped dependency behavior. | Harness scheduler execution core. | high | Changing this risks scheduler determinism and deadlock behavior. |
| `tools/harness/scheduler/scheduler/summary-timing-drift.mjs` | Validates timing consistency across retained scheduler summaries, events, target summaries, and nested scheduler records. | `validateSchedulerSummaryTiming`. | Timing drift CLI. | Retained result tree summaries and events. | Timing drift cases in `test-check-scheduler.sh`. | Scheduler summary timing drift, nested scheduler accounting. | Scheduler timing drift checker. | high | Retained-run safety and warm-check health gate. |
| `tools/harness/duration-accounting/service-backed-make-target-durations-cli.mjs` | Maintains and checks service-backed scheduler work-unit duration baselines. | Executable CLI surface. | Make targets `service-backed-make-target-duration-baselines` and drift target. | Duration CLI helper, duration drift helper, target baselines, scheduler summaries/events, execution topology. | `test-service-backed-make-target-duration-baselines.sh`. | `cartulary.scheduler_work_unit_duration_baselines.v2`, service-backed scheduler summaries. | Duration accounting. | high | `update` mode mutates baseline inputs and requires retained evidence plus later authorization. |
| `tools/harness/scheduler/service-backed-schedule-cli.mjs` | Direct service-backed scheduler CLI. | Executable CLI surface; service-backed scheduler summary schema constants. | `Makefile`, task surface, harness contract context, tests. | Scheduler manifest/resources/engine, browser adapter, backend adapter, topology, test-output, service lifecycle. | `test-service-backed-scheduler.sh`. | `cartulary.scheduler_manifest.v1`, `cartulary.scheduler_event.v6`, `cartulary.service_backed_scheduler_summary.v10`. | Service-backed scheduler CLI facade. | high | Public behavior includes fixture budget, browser sessions, Go shards, finalizers, and target summaries. |
| `tools/harness/scheduler/service-backed-schedule-manifest.mjs` | Validates service-backed schedule source manifest shape. | `validateServiceBackedScheduleManifestShape`, `serviceBackedScheduleSchemaID`. | Service-backed topology and expansion helpers. | Generated artifact validation helpers. | `test-service-backed-scheduler.sh`. | Service-backed schedule source schema and generated JSON shape checks. | Service-backed schedule manifest validation. | medium | Compatibility layer for schedule source manifests. |
| `tools/harness/scheduler/service-backed-schedule-topology.mjs` | Validates service-backed schedule topology against execution topology and browser batch manifest. | `validateServiceBackedScheduleTopology`. | Service-backed schedule CLI and tests. | Browser adapter, execution topology, service-backed manifest. | `test-service-backed-scheduler.sh`. | Execution topology, browser stage tags/dependencies, service-backed source manifests. | Generated-artifact execution topology validation. | high | Mixed scheduler, browser, and generated topology validation. |
| `tools/harness/scheduler/target-duration-baselines.mjs` | Reads and validates positive target duration baseline JSON. | `readJSON`, `sortedObjectByKey`, `assertPositiveTargetWeights`, `readPositiveTargetBaseline`. | Duration baseline CLIs. | Filesystem baseline JSON. | Duration baseline shell tests. | Duration baseline schema files. | Duration accounting support. | medium | Shared validation helper for baseline maintenance. |
| `tools/harness/scheduler/tests/test-check-scheduler.sh` | Shell test suite for check scheduler semantics, artifacts, schemas, event order, timing drift, resource behavior, and CLI output. | Executable test script. | Harness smoke and task surface. | Scheduler CLIs, temp fixtures, Node assertions. | It is the direct test. | Scheduler event, summary, pressure, target-summary, tool-run-summary schemas. | Harness scheduler test coverage. | high | Broad characterization evidence for check scheduler. |
| `tools/harness/scheduler/tests/test-harness-smoke-duration-baselines.sh` | Shell test for harness-smoke duration baseline update and drift behavior. | Executable test script. | Harness smoke and task surface. | Harness-smoke duration CLI, temp task surface, retained target summaries. | It is the direct test. | `cartulary.harness_smoke_duration_baselines.v1`. | Duration accounting test coverage. | medium | Covers baseline mutation in temp files only. |
| `tools/harness/scheduler/tests/test-run-phase-slice.sh` | Node-backed shell test for phase-slice plans and scheduler execution. | Executable test script. | Harness smoke and task surface. | `phase-slice-cli.mjs`, `scheduler-runner.mjs`, backend target plan, synthetic phase registry/maps. | It is the direct test. | `cartulary.phase_slice_plan.v1`, `cartulary.phase_slice_scheduler_summary.v4`, scheduler events. | Phase-slice planning test coverage. | high | Directly exercises mixed phase-accounting and scheduler behavior. |
| `tools/harness/scheduler/tests/test-service-backed-make-target-duration-baselines.sh` | Shell test for service-backed make-target duration baseline update/drift and contamination handling. | Executable test script. | Harness smoke and task surface. | Duration CLI, temp scheduler summaries/events, copied topology/baselines. | It is the direct test. | `cartulary.scheduler_work_unit_duration_baselines.v2`, scheduler summaries/events. | Duration accounting test coverage. | high | Covers retained-run safety behavior in temporary fixtures. |
| `tools/harness/scheduler/tests/test-service-backed-scheduler.sh` | Shell test suite for service-backed scheduler semantics, artifacts, browser/session lifecycle, resources, and legacy manifest rejection. | Executable test script. | Harness smoke and task surface. | Service-backed scheduler CLI, test-output, temp manifests, browser session fixtures. | It is the direct test. | Service-backed scheduler summaries, scheduler events, tool-run summaries, web E2E session lease. | Service-backed scheduler test coverage. | high | Broad characterization evidence for service-backed scheduling. |

## 3. Module Boundary Diagnosis

Architectural diagnosis:

- `tools/harness/scheduler` is a mixed-responsibility harness package.
- The package contains legitimate scheduler execution core logic and stable facades, but it also contains phase-slice planning, service-backed topology expansion, generated-topology validation, browser/runtime adapters, backend shard-plan adapters, and duration baseline accounting.
- The target must not be treated as a permanent product module boundary merely because the path exists.
- No evidence was found that this target owns workbook runtime behavior for timeline, projections, revisions, collaboration, imports/tabular ingest, entities/indicators, evidence, links, saved views, view contracts, production transport/platform adapters, frontend shell/controller state, or grid-adapter/vendor integration.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Scheduler execution engine | `scheduler/engine.mjs`, `scheduler/state.mjs`, `scheduler/clock.mjs`, `scheduler/process-executor.mjs` | Harness scheduler core | keep | Scheduler NLSpec ACs and direct tests assert scheduler artifacts and semantics. | Candidate internal core behind `scheduler-runner.mjs`. |
| Scheduler manifest and command-shape validation | `scheduler-manifest.mjs` | Harness scheduler manifest core | keep | `cartulary.scheduler_manifest.v1` and command-shape closure. | Public compatibility surface. |
| Resource registry, claims, and auto limits | `scheduler-resources.mjs` | Harness scheduler resource core | keep | NLSpec resource registry and auto-capacity rules. | Generated task-surface consumers import this file. |
| Scheduler reporting and diagnostics | `scheduler-reporting.mjs`, `scheduler/blockers.mjs`, `scheduler/progress-recorder.mjs` | Harness scheduler diagnostics/reporting | keep | Retained artifacts and stdout/progress tests. | Pressure summary remains diagnostic evidence. |
| Browser scheduler integration | `adapters/browser.mjs`, `scheduler-browser-runtime.mjs`, runtime command helpers | Browser scheduler adapter facade | keep | Harness NLSpec identifies browser scheduler adapter owner facade. | Avoid private browser imports outside the facade. |
| Backend shard-plan access | `adapters/backend.mjs` | Backend shard-plan facade | keep | Adapter only re-exports backend shard plan functions. | Keep thin unless backend owner changes. |
| Schedule context and execution topology access | `adapters/schedule-context.mjs`, `execution-dependencies.mjs` | Generated-artifact execution topology facade | keep | Generated topology helpers are imported through facades. | Do not hand-edit generated outputs. |
| Phase-slice planning | `phase-slice-plan.mjs`, `phase-slice-cli.mjs` | Phase-accounting or phase-slice planning facade | split | Imports phase registry, frontend phase plan, backend target plan, browser stages, execution topology. | Verification/accounting layer, not product runtime module. |
| Service-backed expansion and topology validation | `check-service-backed-expansion.mjs`, `service-backed-schedule-topology.mjs` | Service-backed schedule planning plus execution-topology validation | split | Builds work units from schedule sources and validates browser/generated topology. | Candidate facade before any movement. |
| Duration baseline maintenance | `duration-*.mjs`, `*-durations-cli.mjs`, drift suite shell wrapper | Duration accounting and retained-run maintenance | split | Baseline schemas and retained-run drift commands. | `update` commands require later authorization. |
| Drift check CLIs | `scheduler-event-order-drift-cli.mjs`, `scheduler-summary-timing-drift-cli.mjs`, `scheduler/event-order.mjs`, `scheduler/summary-timing-drift.mjs` | Scheduler evidence drift checker | keep or split | Retained scheduler summaries/events are drift evidence. | Could be separate evidence-validation facade. |
| Shell tests | `tests/*.sh` | Harness scheduler test coverage | keep | Task-surface harness smoke references these tests. | Test-only assumptions should not leak into implementation. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Make-owned scheduler target names | `Makefile`, `tools/task_surface_manifest.json` | Targets include `check`, `check-service-backed`, `phase-slice`, `service-backed-slice`, scheduler drift, duration baseline, and harness-smoke targets. | Scheduler shell tests through harness smoke. | Preserve target names and wrapper invocation in any move. | high | Do not hand-edit `tools/task_surface.generated.mk`. |
| Scheduler manifest schema | `scheduler-manifest.mjs` | `cartulary.scheduler_manifest.v1`; command type closure and command shapes. | `test-check-scheduler.sh`, `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | Add focused characterization before moving command validation. | high | Generated JSON-shape checks consume this. |
| Scheduler event schema and ordering | Scheduler CLIs, `scheduler/event-order.mjs` | `cartulary.scheduler_event.v6`; seq, monotonic, wall time, clock-skew handling. | Event-order cases in `test-check-scheduler.sh`. | Preserve retained JSONL ordering cases. | high | Used by drift commands and timing summary validation. |
| Scheduler summaries | Scheduler CLIs and engine | `cartulary.check_scheduler_summary.v10`, `cartulary.service_backed_scheduler_summary.v10`, `cartulary.phase_slice_scheduler_summary.v4`. | All scheduler shell tests. | Characterize summary fields touched by a slice. | high | Includes nested scheduler and accounting extensions. |
| Retained artifacts | `scheduler-reporting.mjs`, `scheduler/engine.mjs` | `scheduler-summary.json`, `scheduler-events.jsonl`, `pressure-summary.json`, `progress-summary.log`, log directories. | Shell tests assert files and fields. | Preserve path and relative/absolute path safety assertions. | high | Artifact paths are harness compatibility surface. |
| Resource registry and auto limits | `scheduler-resources.mjs`, `tools/scheduler_resource_registry.json` | Registry schema `cartulary.scheduler_resource_registry.v4`; auto capacity rules. | Scheduler tests and JSON-shape checks. | Add targeted tests before resource ownership movement. | high | Logical resources are scheduling constraints, not physical guarantees. |
| Browser sessions and worker slots | Browser adapter and runtime helpers | Browser session keys, worker-slot env, finalizers, stage completion. | `test-service-backed-scheduler.sh`, `test-run-phase-slice.sh`. | Preserve session lifecycle and worker-slot range behavior. | high | Adapter facade is explicitly owned by harness NLSpec. |
| Runtime binary injection | Scheduler CLIs and runtime helpers | Test-services and browser runtime binaries injected into commands. | Phase-slice and scheduler shell tests. | Preserve command env and missing-binary failure behavior. | high | NLSpec AC-037. |
| Phase-slice plan schema and behavior | `phase-slice-plan.mjs`, `phase-slice-cli.mjs` | `cartulary.phase_slice_plan.v1`; phase namespace/mode and work-unit selection. | `test-run-phase-slice.sh`. | Characterize plan JSON before splitting phase planning. | high | Phase maps are evidence accounting only. |
| Service-backed schedule expansion | `check-service-backed-expansion.mjs`, `service-backed-schedule-topology.mjs` | Schedule source manifests, generated topology, browser stage dependencies. | `test-service-backed-scheduler.sh`, `test-check-scheduler.sh`. | Preserve expansion parity and topology rejection cases. | high | Candidate split after facade stabilization. |
| Duration baseline drift | Duration CLIs and helpers | `cartulary.scheduler_work_unit_duration_baselines.v2`, `cartulary.harness_smoke_duration_baselines.v1`. | Two duration baseline shell tests. | Preserve update/check-drift temp-fixture behavior. | high | Retained-run commands need `RESULTS_DIR`. |
| Authorization checks | Not owned by target | No production auth implementation found in target. | Indirect through harness child commands only. | TODO: verify only if future slice touches service/browser target commands. | defer | Product auth remains outside scheduler. |
| HTTP routes and WebSocket events | Not owned by target | No HTTP route or WebSocket handler implementation found in target. | None in scheduler tests. | No scheduler characterization required unless future slice touches browser test server code. | low | Explicitly no direct product HTTP/WebSocket behavior. |
| Workbook row/query/mutation/save/conflict/inspector behavior | Not owned by target | No workbook runtime mutation code found in target. | None in scheduler tests. | No scheduler characterization required. | low | Product workbook modules remain out of scope. |
| Saved-view or view-schema behavior | Not owned by target | No view-schema implementation found in target. | None in scheduler tests. | No scheduler characterization required. | low | Generated harness schemas are distinct from product view contracts. |
| Grid adapter or UI selector contracts | Not owned by target | No direct grid-vendor integration found in target. | None in scheduler tests. | Defer if future evidence finds UI-selector contract usage. | defer | Browser harness commands may run UI tests but do not own UI selectors. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Scheduler target is mixed responsibility rather than one permanent module. | Path contains engine, CLIs, adapters, phase planning, service-backed topology, duration accounting, and tests. | high | should_fix | Split across harness scheduler core, phase planning, service-backed planning, duration accounting, adapters. | Plan slices around stable facades before movement. |
| Phase-slice planning is coupled to scheduler execution. | `phase-slice-plan.mjs` imports phase registry/accounting, frontend phase plan, backend target plan, browser adapter, and execution topology. | high | should_fix | Phase-accounting or phase-slice planning facade. | Characterize plan JSON, then isolate planning API from execution runner. |
| Service-backed expansion mixes schedule planning, browser topology, resource mapping, and generated execution topology. | `check-service-backed-expansion.mjs` and `service-backed-schedule-topology.mjs`. | high | should_fix | Service-backed schedule planning plus execution-topology validation. | Add/retain expansion parity tests before any split. |
| Browser dependency is intentionally routed through adapter facade. | `adapters/browser.mjs`; prior browser handoff and harness NLSpec identify the facade. | high | intentional/no_action | Browser scheduler adapter facade. | Keep public adapter and avoid private browser imports. |
| Backend shard-plan dependency is routed through a thin adapter. | `adapters/backend.mjs` re-exports backend shard-plan helpers. | medium | intentional/no_action | Backend shard-plan facade. | Preserve facade unless backend owner changes. |
| Generated topology and task-surface consumers import scheduler helpers. | `tools/harness/generated-artifacts/*` and `tools/task_surface_manifest.json` reference scheduler resources and paths. | high | should_fix | Generated-artifact helper facades. | Do not move files without updating owner inputs and drift checks. |
| Duration baseline maintenance lives beside scheduler execution. | Duration CLIs and helpers under scheduler path. | high | should_fix | Duration accounting and retained-run maintenance. | Keep behavior frozen; any baseline update requires later authorization. |
| Process executor handles environment/log redaction in scheduler core. | `scheduler/process-executor.mjs` sanitizes Make flags, child env, and log names. | high | must_fix | Harness scheduler process adapter. | Treat as security-sensitive; add characterization before movement. |
| Test-only shell fixtures exercise implementation details. | Tests construct temporary manifests, summaries, leases, and topology inputs. | medium | defer | Harness scheduler tests. | Avoid promoting test fixture assumptions into production behavior. |
| No direct product module ownership found. | Searches and file inspection found harness-only code, not workbook modules. | medium | intentional/no_action | Product modules remain outside scope. | Record no HTTP/WebSocket/workbook contract ownership in tracker. |
| Retained-run validation requires external result root. | Drift commands consume `RESULTS_DIR`. | medium | defer | Scheduler evidence drift checker. | Mark retained-run checks TODO until a successful run root exists. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Record target, source hierarchy, constraints, branch/commit posture, and tracker status. | This tracker, owner docs. | Inspection commands only. | Scope and authority log has current session entry. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every file under `tools/harness/scheduler`. | All target files. | `rg --files tools/harness/scheduler`, targeted source opens. | Inventory table complete. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map public scheduler, drift, duration, artifact, and Make contracts to owners. | Scheduler CLIs, manifests, resources, reporting, task surface. | `rg` over Make/task-surface/schema IDs. | Freeze map complete. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-06 | Identify existing shell tests and missing pre-move characterization. | `tests/*.sh`, scheduler CLIs/helpers. | Harness smoke targets recorded but not run. | Test posture in Section 4 and Section 8. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05 | Classify mixed responsibilities, adapters, generated-artifact coupling, duration accounting. | Adapters, phase/service-backed/duration files. | Source inspection. | Findings table complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06, WF-07 | Define stable facade-first movement strategy. | `scheduler-runner.mjs`, adapter facades, phase/service-backed/duration helpers. | No implementation validation yet. | Slice plan lists behavior-preserving moves. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Sequence smallest safe slices and dependencies. | Tracker plus future affected packages. | Per-slice commands in Section 7. | Slice plan complete. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Record required harness smoke, JSON shape, drift, and retained-run validations. | Task surface, scheduler tests, drift CLIs. | Make-owned commands only. | Validation plan complete. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Provide next-session handoff and blockers. | Tracker file. | No validation run in planning-only session. | Handoff log and binary criteria complete. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Preserve scheduler public facades before any movement. | `scheduler-runner.mjs`, `adapters/*.mjs`, current CLI entrypoints. | Import compatibility, Make wrappers, generated task-surface references. | Preserve all existing scheduler shell tests. | `make run-harness-smoke-fast` | Revert facade/import changes only. | Existing import paths still work and smoke tests pass. |
| S-01 | S-00 | Add or retain characterization around scheduler artifacts, event order, summaries, resources, and failure mapping before core movement. | `tests/test-check-scheduler.sh`, `tests/test-service-backed-scheduler.sh`, `scheduler/engine.mjs`, `scheduler/state.mjs`. | Scheduler schemas, retained artifacts, exit codes. | Preserve check and service-backed scheduler assertions; add targeted cases only if gaps are found. | `make run-harness-smoke-fast` then `make run-harness-smoke-extended` if risk warrants. | Revert added characterization tests only. | Behavior coverage is sufficient before code movement. |
| S-02 | S-01 | Isolate phase-slice planning from scheduler execution behind a stable phase-slice planning facade. | `phase-slice-plan.mjs`, `phase-slice-cli.mjs`, phase-accounting helpers. | `cartulary.phase_slice_plan.v1`, child target selection, service-wrapper behavior. | Preserve `test-run-phase-slice.sh`; add plan JSON snapshots only if needed. | `make run-harness-smoke-fast` | Restore original plan import paths. | Phase-slice JSON and execution behavior remain unchanged. |
| S-03 | S-01 | Isolate service-backed expansion/topology into clearer planning facade while preserving CLI behavior. | `check-service-backed-expansion.mjs`, `service-backed-schedule-topology.mjs`, `service-backed-schedule-cli.mjs`, `check-schedule-cli.mjs`. | Browser stage grouping, generated topology validation, resource mapping. | Preserve `test-service-backed-scheduler.sh` and check scheduler service-backed cases. | `make run-harness-smoke-fast` and `make harness-contract` | Restore original expansion/topology files and imports. | Service-backed work units and topology errors remain unchanged. |
| S-04 | S-01 | Keep browser and backend adapter facades as public harness boundaries; move only private implementation if owner docs later require it. | `adapters/browser.mjs`, `adapters/backend.mjs`, browser/backend harness packages. | Browser session lifecycle, worker slots, backend shard selection. | Preserve scheduler/browser tests and existing browser handoff expectations. | `make frontend-import-boundary-check`, `make run-harness-smoke-fast` | Restore adapter exports. | No scheduler caller imports private browser/backend internals. |
| S-05 | S-01 | Split duration baseline maintenance from scheduler execution into duration accounting ownership. | `duration-*.mjs`, `*-durations-cli.mjs`, `target-duration-baselines.mjs`. | Baseline schemas, retained-run drift, contamination reporting. | Preserve duration baseline shell tests. | `make run-harness-smoke-extended`; retained checks require `RESULTS_DIR`. | Restore CLI paths and baseline helpers. | Drift/update behavior in temp tests remains unchanged. |
| S-06 | S-03, S-05 | Update generated-artifact owner inputs and task-surface references only if file paths change. | Owner inputs for task surface/topology; generated outputs downstream only. | Generated drift and command surfaces. | Preserve harness contract and JSON shape checks. | `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-schedule-drift`, `make generate-drift` | Revert owner-input changes and regenerate if authorized. | No generated drift remains. |
| S-07 | S-02, S-03, S-04, S-05, S-06 | Final retained-run validation and handoff. | Drift CLIs and retained result root. | Timed evidence and retained-run interpretation. | Preserve event-order and summary timing drift checks. | `make scheduler-event-order-drift RESULTS_DIR=<successful-run-root>` and related duration drift commands. | Revert final slice if retained-run checks fail due to refactor. | Handoff includes passing commands or documented unrelated failures. |

Slices S-02 through S-07 are behavior-preserving by default. Any slice that changes public command behavior, schema shape, artifact path, timing budget, fixture lifecycle, browser worker-slot semantics, or baseline contents requires later authorization.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make run-harness-smoke-fast` | Fast harness smoke, including public Make/wrapper, check scheduler semantic, and service-backed scheduler semantic coverage. | yes | Narrowest Make-owned gate for scheduler refactor planning. |
| integration | `make run-harness-smoke-extended` | Broader harness smoke, including phase-slice and duration-baseline smoke targets. | yes | Use after slices touching phase-slice, duration, service-backed topology, or wrappers. |
| e2e/browser | `make browser-e2e-webserver-backed` or phase-selected browser targets via `make task-guide ROLE=feature-dev PHASE=phaseN` | Browser stack behavior indirectly affected by scheduler session and worker-slot changes. | no | Required only if a slice touches browser lifecycle semantics beyond harness smoke coverage. |
| generated drift | `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-schedule-drift`, `make generate-drift` | Generated policy, schema shape, phase schedule, and generated artifact drift. | yes if paths or owner inputs change | Do not hand-edit generated outputs. |
| import-boundary/static | `make frontend-import-boundary-check`, `make lint-scripts`, `make lint-shell` | Import boundaries and script/shell style. | yes for import/path/shell changes | Make-owned commands discovered in task-surface manifest. |
| full check | `make harness-contract` then broader `make check` only if risk requires | Harness contract first; full repository check only for cross-cutting changes. | no for planning-only, yes for broad implementation | Choose narrowest sufficient target first. |
| retained-run drift and finalizer | `make scheduler-event-order-drift RESULTS_DIR=<successful-run-root>`, `make scheduler-summary-timing-drift RESULTS_DIR=<successful-run-root>`, `make service-backed-make-target-duration-baseline-drift RESULTS_DIR=<successful-run-root>`, `make harness-smoke-duration-baseline-drift RESULTS_DIR=<successful-run-root>`, `make duration-baseline-drift-suite RESULTS_DIR=<successful-run-root>`, `make agent-finalize RESULTS_DIR=<successful-run-root>` | Retained scheduler event/timing, duration baseline evidence, and finalizer maintenance. | yes for duration/timing slices | Completed with `RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`. |

Commands were discovered through `Makefile`, `tools/task_surface_manifest.json`, owner docs, and scheduler source/tests.

Implementation validation results from the remediation session:

- `make phase-schedules`: pass; refreshed generated task-surface/topology outputs from owner inputs.
- `make json-shape-check`: pass; validates `cartulary.scheduler_pressure_summary.v1` positive and negative fixtures.
- `make frontend-import-boundary-check`: pass; validates new owner facade classifications and unsupported old scheduler paths.
- `make generated-artifact-policy-check`: pass.
- `make phase-schedule-drift`: pass.
- `make harness-contract`: pass.
- `make lint-scripts`: pass.
- `make lint-shell`: pass.
- `make lint-markdown`: initially failed in `docs/network-flow-activity-nlspec.md`; fixed Markdown table pipe escaping and unused source references in a follow-up lint remediation session, then passed.
- `make generate-drift`: pass.
- `make run-harness-smoke-fast`: pass.
- `make run-harness-smoke-extended`: first run failed in `harness-smoke-phase-slice` because a direct test fixture used non-canonical scheduler kind `phase-slice`; fixed the fixture to `phase_slice` and reran successfully. Passing rerun root: `.cartulary/test-results/20260704T152118Z-p3920991`.
- `make check`: pass after lint remediation; final run root `.cartulary/test-results/20260704T153646Z-p4080542`.
- `make scheduler-event-order-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass; run root `.cartulary/test-results/20260704T154013Z-p4159620`.
- `make scheduler-summary-timing-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass; run root `.cartulary/test-results/20260704T154013Z-p4159645`.
- `make service-backed-make-target-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass; run root `.cartulary/test-results/20260704T154020Z-p4159774`.
- `make harness-smoke-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass; run root `.cartulary/test-results/20260704T154020Z-p4159783`.
- `make duration-baseline-drift-suite RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: pass; run root `.cartulary/test-results/20260704T154026Z-p4160135`; refreshed `tools/browser_e2e_duration_baselines.json`, `tools/execution_topology_render_index.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, and `tools/service_backed_make_target_duration_baselines.json`.
- Post-finalizer checks after tracker update: `make lint-markdown`, `make json-shape-check`, `make generate-drift`, and `git diff --check` passed.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TSS-001 | Normalize target label and initialize tracker | WF-00 | DONE | none | This file; label `test-harness-scheduler`. | Tracker exists at requested output path. |
| TSS-002 | Inspect target path and enumerate files | WF-01 | DONE | TSS-001 | `rg --files tools/harness/scheduler`; inventory table. | Every target file is inventoried. |
| TSS-003 | Record source hierarchy and owner docs | WF-00 | DONE | TSS-001 | Section 1. | Authority posture is explicit. |
| TSS-004 | Diagnose module boundary | WF-04 | DONE | TSS-002 | Section 3. | Mixed-responsibility finding recorded. |
| TSS-005 | Map public contracts and behavior freeze | WF-02 | DONE | TSS-002 | Section 4. | Contract risks have owners and test posture. |
| TSS-006 | Classify coupling findings | WF-04 | DONE | TSS-004 | Section 5. | Findings use required classifications. |
| TSS-007 | Define workstreams and dependencies | WF-06 | DONE | TSS-005, TSS-006 | Section 6. | Workstream dependency chain is explicit. |
| TSS-008 | Define behavior-preserving slice plan | WF-06 | DONE | TSS-007 | Section 7. | Every slice has risk, validation, rollback, and completion criteria. |
| TSS-009 | Record validation command plan | WF-07 | DONE | TSS-008 | Section 8. | Commands are Make-owned or marked TODO with reason. |
| TSS-010 | Execute implementation refactor | WF-05 | DONE | TSS-008, implementation authorization | Owner-facade file moves, NLSpec updates, pressure-summary schema, generated refresh, and validation commands. | Public harness contracts preserved and targeted validation passes. |
| TSS-011 | Run retained-run drift validation | WF-08 | DONE | TSS-010, retained run root | `RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; drift commands and `agent-finalize` results in Section 8. | Retained drift and finalizer commands pass. |
| TSS-012 | Final implementation handoff | WF-08 | DONE | TSS-010, TSS-011 where applicable | This tracker and final report. | Implementation results, validation, and retained-run blocker are recorded. |
| TSS-013 | Promote scheduler pressure summary schema | WF-02 | DONE | TSS-005 | `tools/schemas/cartulary.scheduler_pressure_summary.v1.schema.json`, NLSpec Section 8, runtime validation, JSON-shape fixtures. | `pressure-summary.json` is schema-owned diagnostic evidence before scheduler success. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | Tracker created for planning-only refactor target. | Inspected owner docs and scheduler target; touched only this tracker file. | `sed`, `rg --files`, `ls`, `date`, `git status`. | Scope and authority recorded. | None for tracker creation. | Use this tracker as source for a later implementation task. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Implementation authorization resolved RB-002; product behavior remains Core/subsystem-owned. | Touched tracker, NLSpec, scheduler/phase/execution/duration/diagnostics harness files, generated owner inputs, generated outputs, and schema/test fixtures. | `git mv`, `git rm`, `rg`, `sed`, `perl`, `make phase-schedules`, validation commands listed in Section 8. | Scheduler remediation implemented with targeted validation passing. | Retained-run drift still needs a warm full-check `RESULTS_DIR`. | Supply a warm `make check` run root for retained drift/finalizer validation if duration/timing maintenance is required. |
| 2026-07-04T11:40:44-04:00 | Codex GPT-5 retained-run cleanup session | Retained-run and finalizer blockers cleared using successful full-check root. | Touched tracker and generated/baseline outputs refreshed by `agent-finalize`. | `make scheduler-event-order-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`, `make scheduler-summary-timing-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`, duration drift targets, `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`. | Retained drift and `agent-finalize` pass; finalizer refreshed six generated/baseline files. | None for retained-run validation. | Keep `.cartulary/test-results/20260704T153646Z-p4080542` as the evidence root for this handoff. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | Backend coupling is through shard-plan and target-plan facades, not product backend modules. | Inspected `adapters/backend.mjs`, `adapters/schedule-context.mjs`, `phase-slice-plan.mjs`, `service-backed-schedule-cli.mjs`; touched tracker only. | `rg`, `sed`. | Backend ownership candidate recorded as adapter/facade, not scheduler module ownership. | None for planning. | Preserve backend adapter facade in any future movement. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Scheduler-local backend/schedule-context re-export adapters were removed. | Removed `tools/harness/scheduler/adapters/backend.mjs` and `tools/harness/scheduler/adapters/schedule-context.mjs`; updated scheduler callers to backend/execution owner facades. | `git rm`, `rg`, validation commands in Section 8. | No scheduler caller imports the removed shallow adapters; import-boundary check passes. | None for backend boundary. | Keep backend shard-plan and target-plan imports behind their owner facades. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | No frontend shell/controller or grid-vendor ownership found; browser harness integration is routed through scheduler/browser adapter. | Inspected `adapters/browser.mjs`, `scheduler-browser-runtime.mjs`, `runtime-command-helpers.mjs`, phase/browser references; touched tracker only. | `rg`, `sed`. | Browser harness behavior recorded as adapter facade and contract risk. | Defer UI-selector/grid conclusions unless future evidence finds direct usage. | Keep browser scheduler adapter stable. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Browser scheduler adapter remains the declared facade; no frontend shell/controller or grid-vendor behavior moved. | Kept `tools/harness/scheduler/adapters/browser.mjs`; updated phase/service-backed callers around it. | `rg`, `make frontend-import-boundary-check`, smoke tiers. | Browser adapter semantics preserved by passing fast and extended smoke. | None for frontend boundary. | Route any product/UI behavior question back to Core/subsystem owners. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | Scheduler schemas, task-surface references, generated artifact consumers, and drift commands mapped. | Inspected scheduler schema declarations, `tools/task_surface_manifest.json`, `Makefile`; touched tracker only. | `rg` for schema IDs, scheduler paths, and Make/task-surface targets. | Contract freeze map populated. | Generated outputs must not be hand-edited. | If paths move later, update owner inputs and run generated drift checks. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Contract/schema updates are implemented and generated outputs were refreshed through Make. | Touched `docs/testing-harness-nlspec.md`, `tools/schemas/cartulary.scheduler_pressure_summary.v1.schema.json`, task-surface/topology owner inputs, generated outputs, and JSON-shape fixtures. | `make phase-schedules`, `make json-shape-check`, `make generated-artifact-policy-check`, `make phase-schedule-drift`, `make generate-drift`. | Pressure summary is schema-owned diagnostic evidence; generated drift checks pass. | Retained-run drift not run without warm full-check root. | Do not hand-edit generated outputs; rerun Make generators after owner input path changes. |
| 2026-07-04T11:40:44-04:00 | Codex GPT-5 retained-run cleanup session | Finalizer refreshed generated/baseline artifacts from the retained full-check root. | `tools/browser_e2e_duration_baselines.json`, `tools/execution_topology_render_index.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, `tools/service_backed_make_target_duration_baselines.json`, tracker. | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`. | `agent-finalize` passed with `generated=updated`, `duration=refreshed`, `run_checks=pass`. | None. | Preserve these generated/baseline updates with the refactor. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | Existing shell tests cover check scheduler, service-backed scheduler, phase-slice, and duration baselines. | Inspected `tools/harness/scheduler/tests/*.sh`; touched tracker only. | `sed`, `rg`. | Test posture recorded in inventory and contract map. | No validation run for planning-only tracker creation. | Run Make-owned harness smoke targets during implementation. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Test fixtures updated for moved paths and canonical scheduler kind. | Touched scheduler shell tests, generated-artifact shell tests, and harness contract tests. | `make harness-contract`, `make run-harness-smoke-fast`, `make run-harness-smoke-extended`, `make lint-shell`. | Fast and extended smoke pass; first extended run exposed and fixed a non-canonical `phase-slice` fixture value. | Retained-run drift commands skipped without full-check `RESULTS_DIR`. | Run retained drift commands when a warm full-check root is available. |
| 2026-07-04T11:40:44-04:00 | Codex GPT-5 retained-run cleanup session | Retained scheduler and duration drift checks were executed against the full-check root. | Tracker only, plus finalizer-updated generated/baseline files. | `make scheduler-event-order-drift`, `make scheduler-summary-timing-drift`, `make service-backed-make-target-duration-baseline-drift`, `make harness-smoke-duration-baseline-drift`, `make duration-baseline-drift-suite` with `RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`. | All retained drift commands passed. | None. | No retained-run validation blocker remains. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | No production auth implementation found; process executor redaction and child env handling are security-sensitive harness behavior. | Inspected `scheduler/process-executor.mjs`, scheduler CLIs; touched tracker only. | `sed`, `rg`. | Security-sensitive coupling recorded. | Product auth conclusions deferred unless future slice touches service/browser runtime code. | Characterize env/log redaction before moving process execution. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Process execution/redaction moved behind `tools/harness/scheduler/process-executor.mjs`; product auth untouched. | Moved process executor and updated scheduler engine import. | Import checks, `make run-harness-smoke-fast`, `make run-harness-smoke-extended`. | Scheduler child execution still passes smoke coverage; no product auth files changed. | No retained log audit without retained full-check root. | Keep future process changes focused on env stripping, log-name sanitization, dry-run detection, and replay behavior. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-04T10:51:05-04:00 | Codex GPT-5 tracker session | Refactor plan is ready; implementation is deferred. | Touched only tracker file. | Inspection commands only. | Future slices and validations are listed. | Retained-run drift validation needs `RESULTS_DIR`; implementation needs later authorization. | Start with S-00 and S-01 in a separate authorized implementation task. |
| 2026-07-04T11:24:29-04:00 | Codex GPT-5 implementation session | Refactor implementation is complete for scheduler harness ownership, with targeted validation passing. | Files listed in final report; tracker updated. | Validation commands listed in Section 8. | Public harness contracts preserved; old private scheduler paths are unsupported and absent from live metadata. | Warm retained full-check root unavailable for retained drift/finalizer maintenance. | Optional next step: run retained drift/finalizer commands with `RESULTS_DIR=<successful warm check root>`. |
| 2026-07-04T11:40:44-04:00 | Codex GPT-5 retained-run cleanup session | Retained-run and finalizer blockers cleared. | Tracker and finalizer-updated generated/baseline files. | Retained drift suite and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`. | All requested blocker-clearing commands passed. | None recorded. | Continue with normal review/commit flow. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Retained-run drift and finalizer validation. | Event-order, summary timing, duration baseline drift, and finalizer maintenance require retained artifacts from an accepted full `make check` profile. | `.cartulary/test-results/20260704T153646Z-p4080542`; retained drift commands and `agent-finalize` passed. | RESOLVED. |
| RB-002 | Implementation refactor authorization. | The original tracker session was planning-only, but the follow-up task authorized implementation. | User follow-up authorization and this implementation handoff. | RESOLVED; no longer a blocker. |
| RB-003 | Generated task-surface or schedule owner-input changes. | Scheduler paths were referenced by task-surface manifest and generated outputs. | Owner input updates plus Make-owned regeneration and drift checks. | RESOLVED for this refactor; `make phase-schedules`, `make phase-schedule-drift`, and `make generate-drift` passed. |
| RB-004 | Product HTTP/WebSocket/workbook behavior ownership is out of scope unless new evidence appears. | Scheduler does not directly implement those product contracts, but browser harness commands can exercise them indirectly. | Future slice evidence showing direct coupling; otherwise Core/subsystem owner docs. | DEFERRED; product behavior remains owned by Core and subsystem specifications. |

## 12. Binary Completion Criteria

- Every file in `tools/harness/scheduler` is inventoried or explicitly out of scope: complete.
- Every discovered public contract risk has an owner and test posture: complete.
- Every proposed workflow has dependencies and exit criteria: complete.
- Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`: complete.
- Validation commands are discovered or marked `TODO` with a reason: complete.
- Contradictions are marked `BLOCKED: owner contradiction`: no contradiction found.
- Repository/framework mismatches are recorded as planning findings: complete; the framework is doctrine only and the live repository shows a mixed-responsibility harness package rather than a permanent product module boundary.
- Handoff sections are current enough for another agent to continue without rediscovery: complete.
- Retained-run drift and finalizer blocker rows are resolved with `RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`: complete.

## 13. Remediation Execution Tracker

This section tracks the follow-up remediation plan authorized on 2026-07-04. The scope remains harness-only: no product HTTP, WebSocket, workbook, view-schema, or Core behavior changes are in scope unless a later slice finds direct coupling evidence and records the owner authority before changing behavior. Public harness contracts remain stable unless `docs/testing-harness-nlspec.md` is updated first.

Required execution rule: update this section after each remediation workstream is complete and before starting the next workstream.

| ID | Workstream | Status | Dependencies | Exit criteria | Evidence |
| --- | --- | --- | --- | --- | --- |
| REM-00 | Tracker baseline | DONE | none | Remediation section records scope, sequencing, dependencies, and tracker-update rule. | 2026-07-04T12:16:16-04:00 baseline entry. |
| REM-01 | Phase-slice service-wrapper reexec fix | DONE | REM-00 | Service-wrapper reexec derives the phase-accounting CLI path structurally and targeted regression coverage fails on the old scheduler path. | `tools/harness/phase-accounting/phase-slice-cli.mjs` now derives the reexec script from `scriptDir`; `test-run-phase-slice.sh` captures fake service-wrapper args; `make run-harness-smoke-extended` passed. |
| REM-02 | Compatibility boundary closure | DONE | REM-01 | Unsupported legacy scheduler helper paths remain unsupported, import-boundary coverage passes, and no executable live caller references old paths. | `harness-import-boundary.mjs` reports unsupported helper paths; contract tests assert the full legacy scheduler catch-all set; `make frontend-import-boundary-check` and `make harness-contract` passed; live `rg` scan found no executable references outside registry/tests. |
| REM-03 | Runtime binary provenance coverage | DONE | REM-02 | Operator runtime-binary public/internal input rules, executable checks, retained build-artifact provenance, and digest mismatch behavior are covered. | Backend target-execution fixture covers missing env, missing artifact, digest mismatch, non-executable, symlink, and valid `runtime-binaries.json`; contract tests reject command-line `OPERATOR_BIN`/`CARTULARY_OPERATOR_BIN` on unrelated targets; `make run-harness-smoke-fast` and `make harness-contract` passed. |
| REM-04 | Scheduler contract hardening | DONE | REM-03 | Resource/default policy and primary failure determinism fixtures cover the closed NLSpec behavior. | Failure taxonomy tests now pin lifecycle, event, child-order, artifact, reason, and input-order tie breakers; scheduler resource tests now assert fixed/default resource sources for check and service-backed profiles; `make harness-contract` and `make run-harness-smoke-fast` passed. |
| REM-05 | Lifecycle, reset, cleanup, and redaction closure | DONE | REM-04 | Lifecycle accounting, runtime reset/test-route guard, cleanup guard, and redaction closure have explicit passing evidence. | Lifecycle tests now cover duplicate child start and terminal mutation rejection; cleanup/redaction contract tests cover no-op missing cleanup-owned paths, structural `lease_file` fields, browser stage sessions, and CLI secret args; `make harness-contract`, `make run-harness-smoke-lifecycle`, and `make test-fast` passed. |
| REM-06 | Static-analysis, generated-artifact, and duration ownership closure | DONE | REM-05 | Govulncheck, migration/schema validator, generated-artifact policy, duration accounting, and retained-run finalizer ownership checks pass or have recorded blockers. | Verification-only slice. Retained root `.cartulary/test-results/20260704T153646Z-p4080542` exists; old ownership paths have no live executable references outside registry/tests; `make generated-artifact-policy-check`, `make json-shape-check`, `make migration-drift`, `make duration-baseline-drift-suite RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`, and `make go-vulncheck` passed. |
| REM-07 | Final validation and handoff | DONE | REM-06 | Tracker records files changed, substantive edits, validation commands/results, skipped checks with reasons, and residual risks. | Final validation passed, including retained drift/finalizer and broad `make check`; no checks skipped. Broad check run root: `.cartulary/test-results/20260704T163339Z-p152515`. |

### Remediation Command Log

| Time | Workstream | Files inspected or changed | Commands run | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-07-04T12:16:16-04:00 | REM-00 | Inspected this tracker, `tools/harness/phase-accounting/phase-slice-cli.mjs`, `tools/harness/scheduler/tests/test-run-phase-slice.sh`, `docs/testing-harness-nlspec.md`, `docs/domain.md`. Changed this tracker. | `sed`, `rg`, `git status`, `date`. | Baseline tracker section added; product behavior remains out of scope. | Start REM-01. |
| 2026-07-04T12:19:27-04:00 | REM-01 | Changed `tools/harness/phase-accounting/phase-slice-cli.mjs`, `tools/harness/scheduler/tests/test-run-phase-slice.sh`, and this tracker. | `make run-harness-smoke-extended`. | Pass. The service-wrapper reexec now targets `tools/harness/phase-accounting/phase-slice-cli.mjs`; the regression fixture proves the wrapper args and setup sequencing through fake Make/test-services binaries. | Start REM-02. |
| 2026-07-04T12:20:59-04:00 | REM-02 | Changed `tools/harness/static-analysis/harness-import-boundary.mjs`, `tools/harness/tests/test-harness-contracts.mjs`, and this tracker. | `make frontend-import-boundary-check`; `make harness-contract`; `rg` scan for legacy scheduler helper references under `tools`. | Pass. The `rg` scan exited `1` with no matches, which is the expected clean result after excluding the static-analysis registry and its contract test. | Start REM-03. |
| 2026-07-04T12:23:32-04:00 | REM-03 | Changed `tools/harness/backend/tests/test-run-go-target.sh`, `tools/harness/tests/test-harness-contracts.mjs`, and this tracker. | `make run-harness-smoke-fast`; `make harness-contract`. | Pass. Runtime-binary provenance now has focused owner-boundary coverage and public Make preflight coverage. | Start REM-04. |
| 2026-07-04T12:25:29-04:00 | REM-04 | Changed `tools/harness/tests/test-harness-contracts.mjs`, `tools/harness/scheduler/tests/test-check-scheduler.sh`, and this tracker. | `make harness-contract`; `make run-harness-smoke-fast`. | Pass. Primary failure determinism and resource/default source coverage are pinned at the shared contract and scheduler registry layers. | Start REM-05. |
| 2026-07-04T12:30:18-04:00 | REM-05 | Changed `internal/testutil/suiteservices/diagnostics_test.go`, `tools/harness/tests/test-harness-contracts.mjs`, and this tracker. | `make harness-contract`; `make run-harness-smoke-lifecycle`; `make test-fast`. | Pass. `make test-fast` run root: `.cartulary/test-results/20260704T162727Z-p75992`; reset/test-route coverage remains in the passing Go gate. | Start REM-06. |
| 2026-07-04T12:31:30-04:00 | REM-06 | Inspected owner references and changed this tracker only. | `rg` ownership-path scans; `make generated-artifact-policy-check`; `make json-shape-check`; `make migration-drift`; `make duration-baseline-drift-suite RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; `make go-vulncheck`. | Pass. Static-analysis/security, generated-artifact, migration/schema, and duration retained-run safety checks all pass with no new code edits. | Start REM-07. |
| 2026-07-04T12:36:34-04:00 | REM-07 | Changed this tracker; final changed set also includes harness tests/code and Make-managed baseline/generated files listed below. | `make lint-scripts`; `make lint-shell`; `make scheduler-event-order-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; `make scheduler-summary-timing-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; `make service-backed-make-target-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; `make harness-smoke-duration-baseline-drift RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T153646Z-p4080542`; post-finalizer `make generate-drift`; `make json-shape-check`; `make lint-markdown`; `make generated-artifact-policy-check`; `make phase-schedule-drift`; `git diff --check`; broad `make check`. | Pass. `agent-finalize` run root: `.cartulary/test-results/20260704T163230Z-p148113`; broad check run root: `.cartulary/test-results/20260704T163339Z-p152515`. No skipped checks or unresolved blockers. | Ready for review/commit. |

### Remediation Final Handoff

- Authored implementation/test edits: phase-slice service-wrapper reexec path, phase-slice wrapper regression, unsupported-private helper registry reporting, runtime-binary provenance tests, public input rejection tests, scheduler resource/default source assertions, primary failure tie-breaker tests, lifecycle duplicate/terminal transition tests, and cleanup/redaction closure assertions.
- Documentation/control artifact edits: this tracker section records all remediation workstreams, command evidence, retained root usage, and completion status.
- Make-managed refreshes from `agent-finalize`: `tools/browser_e2e_duration_baselines.json`, `tools/execution_topology_render_index.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, and `tools/service_backed_make_target_duration_baselines.json`.
- Final validation status: all required targeted gates, retained-run drift/finalizer gates, post-finalizer drift checks, `git diff --check`, and broad `make check` passed.
- Skipped checks: none.
- Residual risks: none recorded for this remediation; product HTTP/WebSocket/workbook behavior remains out of scope and unchanged.

## 14. 2026-07-04 Scheduler Remediation and Evolution Iteration

This iteration is the controlling handoff for the next scheduler cleanup and evolution pass. It records live implementation facts separately from Sections 1 through 13, which remain historical evidence for the earlier scheduler refactor and remediation. This section does not perform or authorize product HTTP, WebSocket, workbook, view-schema, or Core product behavior changes.

### 14.1 Scope and Authority

Owning authority:

- `docs/testing-harness-nlspec.md` owns harness command invocation, target selection, scheduling, fixture lifecycle, service ownership, artifact emission, summary emission, cleanup, verification gates, helper ownership, unsupported-private compatibility status, scheduler resources, retained-run drift, and finalizer requirements.
- `docs/domain.md` owns vocabulary and boundary interpretation. It classifies harness and scheduler terms as implementation-support language, not product-domain behavior.
- Core 00 through Core 04 own product behavior. Core 05 applies only to claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
- `tools/execution_topology_manifest.json` and `tools/task_surface_manifest.json` are owner inputs or generated mirrors as defined by the harness NLSpec. Generated outputs must be refreshed through Make targets, not hand-edited.

In scope for the test harness scheduler:

- scheduler DAG execution, dependency handling, priority handling, resource admission, retained-resource release, finalizers, skipped work, and primary failure propagation;
- scheduler manifest, resource registry, resource-limit resolution, reporting, scheduler events, scheduler summaries, pressure summaries, and retained scheduler artifact paths;
- scheduler process execution, child environment construction, log redaction, log naming, log replay, and child exit propagation;
- browser scheduler adapter behavior needed for scheduler work-unit command projection, session groups, worker slots, and scheduler expansion semantics;
- `check` and service-backed scheduler CLIs, plus scheduler execution used by phase-accounting facades;
- retained scheduler event-order, summary-timing, duration, and `agent-finalize` validation only where scheduler timing, duration, finalizer, or retained-run interpretation is touched.

Out of scope:

- product HTTP routes, WebSocket behavior, workbook mutation/query behavior, saved views, view schemas, frontend shell state, grid-vendor behavior, product auth, deployment behavior, migrations, and Core product conformance;
- duration baseline value refreshes unless retained-run validation is explicitly part of the workstream;
- compatibility shims for old private scheduler paths unless a workstream records a clear future-facing value and a removal condition.

Direct coupling finding: inspected scheduler, phase-accounting, duration-accounting, execution/service-backed, diagnostics, import-boundary, and current smoke/contract tests show no scheduler ownership of product HTTP, WebSocket, workbook, view-schema, or Core product behavior. Browser and service-backed harness work may execute product tests, but the scheduler owns only harness orchestration around that child work.

Inspection basis for this iteration:

- `docs/handoffs/test-harness-scheduler-module-refactor-tracker.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `tools/harness/scheduler/**`
- `tools/harness/phase-accounting/phase-slice-*.mjs`
- `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs`
- `tools/harness/execution/service-backed/*.mjs`
- `tools/harness/duration-accounting/*.mjs`
- `tools/harness/diagnostics/scheduler-*-drift-cli.mjs`
- `tools/harness/static-analysis/harness-import-boundary.mjs`
- `tools/harness/tests/test-harness-contracts.mjs`
- `tools/scheduler_resource_registry.json`
- `tools/execution_topology_manifest.json`
- `tools/task_surface_manifest.json`
- current scheduler, phase-slice, duration, and harness smoke tests under `tools/harness/**/tests`

### 14.2 Current-State Findings

Live implementation facts:

- `tools/harness/scheduler` currently owns the scheduler execution core, public scheduler facades, check scheduler CLI, service-backed scheduler CLI, scheduler manifest/resource/reporting helpers, process executor, browser scheduler adapter, runtime command helpers, events, summary timing validation, pressure summary emission, and scheduler runner tests.
- Phase-slice planning, the phase-slice CLI, and phase-slice smoke coverage now live under `tools/harness/phase-accounting`.
- Service-backed schedule expansion and topology planning live under `tools/harness/execution/service-backed`.
- Duration baseline and drift helpers, service-backed make-target duration smoke coverage, and harness-smoke duration smoke coverage now live under `tools/harness/duration-accounting`.
- Retained scheduler event-order and summary-timing CLIs live under `tools/harness/diagnostics`.
- Unsupported legacy scheduler helper paths are recorded in the NLSpec and import-boundary tests. They must remain unsupported private paths, not recreated as stable compatibility shims.
- `phase_slice` is now a first-class scheduler resource family in the closed NLSpec registry and `tools/scheduler_resource_registry.json`, with a `phase_slice_default` capacity profile. Phase-slice runtime schedules now advertise `resourceScheduler: "phase_slice"` and validate generated resource limits through registry helpers.
- Go shard scheduler-profile resource claims now flow through `tools/harness/scheduler/scheduler-resource-policy.mjs` for `check`, `service_backed`, and `phase_slice`. The owner-backed current mapping is `reset_heavy={cpu:1, io:2, postgres_reset:1}` for service-backed-style schedulers; historical phase-slice `go_io:3` is no longer current conformance.
- Runtime command/session binding for browser session start, browser group execution, browser stage completion, browser session finalization, Go shard and finalizer commands, Make target commands, browser session lease cleanup, and shared child environment sanitization now flows through `tools/harness/scheduler/scheduler/runtime-command-helpers.mjs`.

Remaining coupling and brittle behavior:

- `check-schedule-cli.mjs`, `service-backed-schedule-cli.mjs`, and `phase-slice-cli.mjs` still own schedule-family lifecycle, summary, and wrapper policy. This remaining local code is intentional family-specific orchestration, not duplicated runtime command binding.
- The Section 10.2 resource-table conflict was closed in SE-02; the older summary table now delegates to the closed registry table instead of restating bounds.
- Current-profile scheduler-owned `work_units[].timeout_seconds` support was deferred in SE-02; the key remains rejected until a later adopted watchdog contract covers schema, runner cancellation, finalizer interaction, events, failure mapping, and fixtures.
- `make_prerequisite_policy` omission for scheduler `make_target` work units was closed in SE-02; generated scheduler manifests and check-schedule owner metadata now declare explicit `run` or `skip`, and scheduler manifest validation rejects omission.

Historical notes only:

- Sections 1 through 13 record completed refactor/remediation evidence, including removed scheduler-local adapters, owner-facade movement, retained-run validation, and previous broad validation. They must not be read as proof that current ownership gaps are already resolved.
- Prior retained run roots remain evidence for earlier work, not automatic validation for future duration, timing, scheduler drift, or finalizer changes.

### 14.3 Gap Matrix

| Gap | Remediation | Area | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if left unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Test ownership still follows historical scheduler paths. | Completed in SE-01: moved duration tests to `tools/harness/duration-accounting/tests`, moved phase-slice tests to `tools/harness/phase-accounting/tests`, kept pure scheduler runner assertions in scheduler tests, updated topology owner inputs, and regenerated mirrors. | multiple | Tests should live with the semantic owner they protect. | Reduces scheduler catch-all gravity and makes future owner changes easier to find. | Make target names and smoke tier membership were preserved; no legacy test-path shims were added. | Future maintainers edit or skip the wrong tests and continue treating scheduler as a mixed owner. | Completed: `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make harness-contract`, `make run-harness-smoke-extended`, and `make lint-shell` passed. |
| `phase_slice` is not first-class in the resource registry. | Completed in SE-03: added `phase_slice` to the closed resource registry and NLSpec resource table, added `phase_slice_default`, changed phase-slice runtime schedules to `resourceScheduler: "phase_slice"`, and validated phase-slice plan resource limits through registry helpers. | specification, implementation, tests | Phase growth should use the same resource policy surface as other scheduler kinds. | New phase rows can be added without duplicating capacity policy. | Public target names and schema IDs stayed stable; phase-slice summary `resource_limit_sources` now use normalized `manifest` sources instead of the old bespoke `phase_slice_plan` label. | Resource drift between `check`, `service_backed`, and phase slices; future phases may overload shared lanes. | Completed: `make phase-schedules`, `make json-shape-check`, `make phase-schedule-drift`, `make generate-drift`, `make harness-contract`, `make run-harness-smoke-fast`, and `make run-harness-smoke-extended` passed. |
| Scheduler-profile claim mapping is duplicated and diverges. | Completed in SE-03: created `tools/harness/scheduler/scheduler-resource-policy.mjs` for Go shard profile claims and service-backed-to-check resource translation; updated service-backed expansion and phase-slice planning to call it; added parity coverage for default, `cpu_heavy`, `io_heavy`, `reset_heavy`, `clone_heavy`, and `transaction_heavy`. | implementation, tests | Current profile mappings are policy, not per-caller convenience code. | One policy point for `cpu_heavy`, `io_heavy`, `reset_heavy`, `clone_heavy`, `transaction_heavy`, and future profiles. | The NLSpec authorizes phase-slice `reset_heavy` correction from historical `go_io:3` to `go_io:2`; no compatibility shim was added. | New profiles or fixes land in only one scheduler path. | Completed: profile parity fixture coverage plus `make run-harness-smoke-fast` and `make run-harness-smoke-extended` passed. |
| Runtime command/session binding is duplicated. | Completed in SE-04: extended `runtime-command-helpers.mjs` with scheduler-owned helpers for browser session start/group/complete/finalizer, Go shard/finalizer/target commands, Make target commands, browser lease cleanup, default runtime paths, and shared child env sanitization; migrated check, service-backed, and phase-slice CLIs to use them. | implementation, tests | Duplicated binding code is hard to audit for cleanup, redaction, and env propagation. | New scheduler modes or session types can reuse one reviewed implementation. | Public CLI names, artifact paths, failure mapping, cleanup behavior, and required env fields were preserved. | Cleanup, redaction, session-stop, or runtime-binary fixes remain inconsistent by CLI. | Completed: `node --check` on migrated CLIs/helpers, `make run-harness-smoke-fast`, `make run-harness-smoke-lifecycle`, `make harness-contract`, and `make run-harness-smoke-extended` passed. |
| NLSpec resource tables conflict. | Completed in SE-02: removed the older duplicated bounds table and delegated resource bounds to the closed registry table. | specification, documentation | Implementers should not need to choose between contradictory resource bounds. | Clearer resource contract and lower chance of accidental capacity regressions. | Documentation/spec clarification only; closed registry values were not changed. | Future capacity work may cite the wrong table. | Completed: `make lint-markdown`, `make harness-contract`, and `make json-shape-check` passed. |
| `timeout_seconds` is specified but unsupported. | Completed in SE-02: deferred current-profile scheduler watchdog support and documented future adoption requirements while keeping `timeout_seconds` rejected. | specification, tests | A specified but rejected field is worse than no feature because future work may trust it. | Prevents false current-conformance claims and keeps timeout ownership explicit. | Current rejection remains expected behavior; any future implementation needs a full owner-backed watchdog contract. | Callers may depend on a field that validation rejects or that execution ignores. | Completed: negative shape fixture rejects `timeout_seconds`; `make lint-markdown`, `make json-shape-check`, `make harness-contract`, and `make run-harness-smoke-fast` passed. |
| `make_prerequisite_policy` defaults to legacy `skip`. | Completed in SE-02: made policy explicit for every generated `make_target` work unit and check-schedule owner row; scheduler validation now rejects omission. | specification, implementation, tests | Hidden prerequisite skipping makes future schedule authoring unsafe. | Future schedule rows must declare whether prerequisites are scheduler-modeled or direct Make-owned. | Existing omitted metadata was made explicit as `skip`; runtime-binary producer units use explicit `run`; no public compatibility shim was added. | New manual work units may silently skip correctness setup. | Completed: omitted and invalid policy fixtures reject; `make phase-schedules`, `make json-shape-check`, `make harness-contract`, `make run-harness-smoke-fast`, `make phase-schedule-drift`, and `make generate-drift` passed. |

### 14.4 Sequenced Workstreams

Required execution rule: update this Section 14 after each completed workstream and before starting the next workstream. Do not proceed to a later workstream with stale tracker status.

| ID | Workstream | Depends on | Planned changes | Risks | Exit criteria |
| --- | --- | --- | --- | --- | --- |
| SE-00 | Tracker iteration baseline | none | Append this dated section with scope, findings, gap matrix, workstreams, and validation rules. | Markdown-only change may not exercise implementation assumptions. | Tracker section exists; `make lint-markdown` passes or a concrete skip reason is recorded. |
| SE-01 | Test ownership cleanup | SE-00 | Move/split duration and phase-slice smoke tests into owner-aligned test directories; update execution topology owner inputs; regenerate generated task-surface/topology outputs through Make. | Generated metadata can drift if owner inputs are not updated first. | Public smoke target names and tiers are unchanged; generated drift checks pass; tracker records changed paths and validation results. |
| SE-02 | Spec contract closure | SE-01 | Resolve resource-table conflict; decide and close `timeout_seconds`; make prerequisite-policy omission unsupported after fixtures and generated outputs are explicit. | Over-broad spec edits can accidentally redefine public target behavior. | NLSpec, fixtures, and generated metadata agree; tracker records any public summary/source wording changes. |
| SE-03 | Resource policy facade | SE-02 | Add a single resource-claim policy facade; update phase-slice and service-backed planning to use it; make `phase_slice` capacity registry-backed. | Existing phase-slice summaries may change resource source labels. | Behavior is preserved or differences are explicitly owner-backed; parity fixtures and smoke checks pass. |
| SE-04 | Runtime attachment facade | SE-03 | Extract shared runtime binding helpers for browser sessions, browser groups, Go shards/finalizers, cleanup, and environment construction. | Cleanup or redaction regressions can hide in successful child exits. | No public CLI, artifact path, failure mapping, cleanup, or env-contract drift; lifecycle smoke runs if browser/session cleanup is touched. |
| SE-05 | Final validation and handoff | SE-04 | Run final targeted and broadened validation; record retained roots, skipped checks, residual risks, and ready-for-review status. | Retained-run checks need a valid warm root when duration/timing/finalizer behavior changed. | Tracker has files changed, substantive edits, validation commands/results, skipped checks with reasons, retained-run evidence when required, and next action. |

Implementation preferences for all workstreams:

- Prefer owner facades and structural ownership fixes over shims or compatibility patches.
- Do not preserve old private scheduler paths for historical compatibility.
- Carry behavior forward only when it improves future maintainability or preserves a public harness contract.
- Do not hand-edit generated files.
- Keep product HTTP, WebSocket, workbook, view-schema, and Core product behavior out of scope unless direct coupling evidence is recorded first.

### 14.5 Validation and Handoff Requirements

Validation ladder:

1. For this tracker-only baseline, run `make lint-markdown`.
2. For ownership/test moves, run `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make harness-contract`, `make run-harness-smoke-extended`, and `make lint-shell`.
3. For scheduler contract/resource changes, start with `make json-shape-check`, `make harness-contract`, and `make run-harness-smoke-fast`; add `make run-harness-smoke-extended` when phase-slice, duration, service-backed, generated topology, or public wrapper behavior is touched.
4. For browser/session lifecycle, cleanup, reset, or service ownership changes, add `make run-harness-smoke-lifecycle`.
5. For generated-artifact policy or task-surface changes, add `make generated-artifact-policy-check`, `make phase-schedule-drift`, and `make generate-drift`.
6. For duration, timing, scheduler drift, retained-run interpretation, or finalizer behavior changes, require a successful warm retained `RESULTS_DIR` and run:
   - `make scheduler-event-order-drift RESULTS_DIR=<successful-warm-check-root>`
   - `make scheduler-summary-timing-drift RESULTS_DIR=<successful-warm-check-root>`
   - `make duration-baseline-drift-suite RESULTS_DIR=<successful-warm-check-root>`
   - `make agent-finalize RESULTS_DIR=<successful-warm-check-root>`

Skipped checks must name the exact skipped target and the concrete reason, such as "no successful warm retained `RESULTS_DIR` was available and this workstream did not touch duration, timing, scheduler drift, retained-run interpretation, or finalizer behavior."

Final handoff for each completed workstream must record:

- files inspected and changed;
- substantive edits;
- generated files refreshed through Make, if any;
- validation commands and results;
- retained run root used, if applicable;
- skipped checks with reasons;
- residual risks and the next workstream.

### 14.6 Iteration Tracker

| ID | Workstream | Status | Dependencies | Exit criteria | Evidence |
| --- | --- | --- | --- | --- | --- |
| SE-00 | Tracker iteration baseline | DONE | none | Section 14 exists and `make lint-markdown` passes. | This section appended on 2026-07-04; `make lint-markdown` passed. Broader checks skipped because this workstream changed only the tracker handoff. |
| SE-01 | Test ownership cleanup | DONE | SE-00 | Owner-aligned test paths, regenerated task-surface/topology outputs, and passing owner validation. | Moved `test-run-phase-slice.sh` to `tools/harness/phase-accounting/tests`; moved `test-service-backed-make-target-duration-baselines.sh` and `test-harness-smoke-duration-baselines.sh` to `tools/harness/duration-accounting/tests`; updated `tools/execution_topology_manifest.json`; refreshed `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`, `tools/browser_e2e_batch_manifest.json`, and `tools/execution_topology_render_index.json` through `make phase-schedules`. Validation passed: `make phase-schedules`, `make phase-schedule-drift`, `make generate-drift`, `make harness-contract`, `make run-harness-smoke-extended`, `make lint-shell`. Retained-run checks skipped because this workstream did not touch duration values, timing drift interpretation, scheduler drift interpretation, or finalizer behavior. Next: SE-02. |
| SE-02 | Spec contract closure | DONE | SE-01 | Resource-table, timeout, and prerequisite-policy contracts are closed and validated. | Removed the conflicting resource summary bounds from `docs/testing-harness-nlspec.md`; deferred current-profile `timeout_seconds`; required explicit `make_prerequisite_policy` on scheduler `make_target` work units; made topology owner rows and generated service-backed/phase-slice make units explicit; added negative fixtures for omitted policy and unsupported timeout. Validation passed: `make phase-schedules`, `make lint-markdown`, `make json-shape-check`, `make harness-contract` after one unrelated SeaweedFS lifecycle flake rerun, `make run-harness-smoke-fast` after fixing scheduler fixture fallout, `make phase-schedule-drift`, and `make generate-drift`. Retained-run checks skipped because this workstream did not touch duration values, timing drift interpretation, scheduler drift interpretation, or finalizer behavior. Next: SE-03. |
| SE-03 | Resource policy facade | DONE | SE-02 | Phase-slice and service-backed planning share one resource claim/profile policy facade. | Added `tools/harness/scheduler/scheduler-resource-policy.mjs`; updated `tools/harness/execution/service-backed/schedule-expansion.mjs`, `tools/harness/phase-accounting/phase-slice-plan.mjs`, and `tools/harness/phase-accounting/phase-slice-cli.mjs`; updated `docs/testing-harness-nlspec.md` and `tools/scheduler_resource_registry.json`; refreshed generated schedule artifacts through `make phase-schedules`; added registry/profile parity checks in scheduler smoke fixtures and corrected lingering explicit-policy fixtures. Validation passed: `make phase-schedules`, `make json-shape-check` after the expected stale-generated-input failure, `make phase-schedule-drift`, `make generate-drift`, `make harness-contract`, `make run-harness-smoke-fast`, and `make run-harness-smoke-extended` after fixing stale smoke fixtures for explicit `make_prerequisite_policy`. Retained-run checks skipped because this workstream did not touch duration values, timing drift interpretation, scheduler drift interpretation, or finalizer behavior. Next: SE-04. |
| SE-04 | Runtime attachment facade | DONE | SE-03 | Runtime/session attachment duplication is removed without public harness contract drift. | Extended `tools/harness/scheduler/scheduler/runtime-command-helpers.mjs`; migrated `tools/harness/scheduler/check-schedule-cli.mjs`, `tools/harness/scheduler/service-backed-schedule-cli.mjs`, and `tools/harness/phase-accounting/phase-slice-cli.mjs` to use shared runtime builders for browser sessions/groups/completion/finalizers, Go shard/finalizer/target commands, Make target commands, browser lease cleanup, and child env sanitization. Validation passed: `node --check` for the migrated CLIs and helper, `make run-harness-smoke-fast`, `make run-harness-smoke-lifecycle`, `make harness-contract`, and `make run-harness-smoke-extended`. Retained-run drift checks skipped because command binding was structurally refactored without changing duration values, timing drift interpretation, scheduler drift interpretation, retained-run interpretation, or finalizer semantics. Next: SE-05. |
| SE-05 | Final validation and handoff | DONE | SE-04 | Final validation, retained-run requirements, skipped checks, and residual risks are recorded. | Final validation passed on 2026-07-04: `make lint-markdown`, `make lint-shell`, `make json-shape-check` (`.cartulary/test-results/20260704T173255Z-p479011`), `make generated-artifact-policy-check` (`.cartulary/test-results/20260704T173256Z-p479052`), `make phase-schedule-drift` (`.cartulary/test-results/20260704T173317Z-p480214`), `make generate-drift` (`.cartulary/test-results/20260704T173317Z-p480259`), `make harness-contract`, `make run-harness-smoke-fast` (`.cartulary/test-results/20260704T173317Z-p480293`), `make run-harness-smoke-lifecycle` (`.cartulary/test-results/20260704T173330Z-p482762`), `make run-harness-smoke-extended` (`.cartulary/test-results/20260704T173330Z-p483204`), `make agent-finalize` without `RESULTS_DIR` (`.cartulary/test-results/20260704T173330Z-p482797`; retained run checks skipped by target), and `git diff --check`. No successful warm retained `RESULTS_DIR` was used because this iteration did not change duration values, scheduler timing/drift interpretation, retained-run interpretation, or finalizer semantics. Ready for review. |

### 14.7 Final Handoff

Files inspected and changed:

- Specification and handoff: `docs/testing-harness-nlspec.md`, this tracker.
- Owner inputs and generated mirrors: `tools/execution_topology_manifest.json`, `tools/task_surface_manifest.json`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, `tools/scheduler_resource_registry.json`, and `tools/schemas/cartulary.scheduler_manifest.v1.schema.json`.
- Scheduler/runtime implementation: `tools/harness/scheduler/scheduler-manifest.mjs`, `tools/harness/scheduler/scheduler/runtime-command-helpers.mjs`, `tools/harness/scheduler/check-schedule-cli.mjs`, `tools/harness/scheduler/service-backed-schedule-cli.mjs`, and new `tools/harness/scheduler/scheduler-resource-policy.mjs`.
- Service-backed and phase-accounting implementation: `tools/harness/execution/service-backed/schedule-expansion.mjs`, `tools/harness/phase-accounting/phase-slice-plan.mjs`, `tools/harness/phase-accounting/phase-slice-cli.mjs`, and `tools/harness/phase-accounting/frontend-phase-slice-plan.mjs`.
- Tests and fixtures: moved phase-slice smoke coverage to `tools/harness/phase-accounting/tests/test-run-phase-slice.sh`; moved duration smoke coverage to `tools/harness/duration-accounting/tests/`; updated `tools/harness/scheduler/tests/test-check-scheduler.sh`, `tools/harness/generated-artifacts/tests/test-execution-topology.sh`, and `tools/harness/generated-artifacts/tests/test-json-shapes.sh`.

Substantive edits:

- Closed the NLSpec resource-table conflict, deferred current-profile `timeout_seconds`, and made `make_prerequisite_policy` explicit and required for scheduler `make_target` work units.
- Made `phase_slice` registry-backed with `phase_slice_default`, changed phase-slice runtime resource scheduling to `phase_slice`, and normalized phase-slice resource-limit source reporting through registry validation.
- Centralized Go shard scheduler-profile resource claims and service-backed-to-check claim translation in `scheduler-resource-policy.mjs`; current conformance intentionally maps phase-slice `reset_heavy` to `go_io:2`.
- Centralized runtime command binding, child env sanitization, and browser lease cleanup helpers in `runtime-command-helpers.mjs`.
- Removed old-path test ownership from scheduler test directories without adding compatibility shims.

Generated refresh:

- `make phase-schedules` refreshed scheduler/task-surface/topology mirrors after owner-input and registry changes. Final drift checks passed.

Skipped checks and residual risk:

- Retained-run drift commands with a warm `RESULTS_DIR` were not run because no duration baseline values, scheduler timing/drift interpretation, retained-run interpretation, or finalizer semantics changed. `make agent-finalize` was run without `RESULTS_DIR` and reported retained run checks skipped.
- No known product-scope changes were made. No remaining Section 14 remediation gap is open.
