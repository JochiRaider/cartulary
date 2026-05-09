---
doc_id: THR-S5-OBSERVABLE-INTERFACES
title: Testing Harness Recovery Observable Interface Map
status: active
role: observable-interface-map
---

# Testing Harness Recovery Observable Interface Map

## Document role

This S5 artifact maps caller-visible and machine-consumed interfaces for the
testing harness. It links observable outputs back to S2 entrypoints, S3
artifacts, and S4 service/resource rows. It is recovery evidence only and does
not change command behavior, exit codes, reports, schemas, CI behavior, or
cleanup policy.

Evidence was gathered by static source inspection and non-mutating discovery on
2026-05-08. Existing retained files under `.cartulary/test-results/**` were
listed only as retained artifact examples; they were not promoted to current-run
truth. No service-backed, browser, Docker, Compose, reset, cleanup, formatter,
generator, baseline refresh, broad gate, or controlled failure command was run.

## Session metadata

| Field | Value |
|---|---|
| Repository root | `/home/askahn/code/cartulary` |
| Branch state at S5 refresh start | `## main...origin/main` |
| HEAD revision | `3cfde09450b3217f0cb9613a78a708e866ad37f5` |
| Runtime platform | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Mon Dec 1 20:46:23 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| S5 refresh timestamp | `2026-05-08T22:48:54-04:00` |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` only |

## Classification rules used

| Class | S5 meaning |
|---|---|
| `human_diagnostic` | Human-facing stdout, stderr, recent logs, progress text, report pages, or investigation text. Stable wording must not be assumed unless parsed by a consumer. |
| `machine_consumed` | JSON, JSONL, generated manifests, status files, or report files read by harness code, schedulers, drift/baseline tools, explain tools, fixture reports, wrappers, or CI/release gates. Requires a schema note or `TODO: schema_unknown`. |
| `durable_report` | Retained run/report artifact or committed generated report reused after the producing command exits. Requires run identity or freshness rule. |
| `temporary_log` | Scratch/runtime-root stdout, stderr, process, progress, or tool log that is diagnostic unless separately parsed. |
| `failure_only_artifact` | Output normally relevant only on failure, verbose mode, timeout, or trace-on-failure. Requires schema note or `TODO: schema_unknown` when machine-consumed. |
| `authority_ambiguous` | Output from direct package scripts, local-dev commands, reset routes, CI-provider behavior, or cleanup surfaces whose first-class contract status is unresolved. |

## Controlling input rows

| Source | Rows consumed by S5 |
|---|---|
| S2 entrypoints | `EP-0002` through `EP-0021`, with Make/task-surface rows treated as primary and package-script rows kept separate. |
| S3 artifacts | `ART-0013` through `ART-0030`, plus fixture/golden rows where test assertion failures depend on them. |
| S3 cleanup | `CLN-0001`, `CLN-0002`, `CLN-0011` through `CLN-0020`. |
| S4 services/resources/env | `SVC-0001` through `SVC-0015`, `RES-0001` through `RES-0026`, `ENV-*` rows, and `HAZ-S4-*` routing rows. |
| Source limits and ambiguities | `SL-0001`, `SL-0004`, `SL-0006`, `SL-0007`, `SL-0009` through `SL-0015`; `AMB-0001`, `AMB-0010` through `AMB-0028`. |

## Observable interface map

| Interface ID | Entrypoint IDs | Output surface | Consumer | Machine-consumed | Schema/path | Ordering guarantee | Stable across CI/local | Authority class | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| OI-0001 | `EP-0002` | Phase wrapper stdout/stderr policy, optional banner, per-phase `stdout.log`, `stderr.log`, `phase-summary.json`, `meta.json`, `tool-run-summary.json`, and timing spans. | Make callers, target summaries, explain tools, humans. | yes for JSON summaries; logs are diagnostic. | `SCHEMA-0001`, `SCHEMA-0002`; `.cartulary/test-results/<run-id>/<target>/<phase>/`. | `run-phase.sh` captures child output, writes phase artifacts, then emits helper summary; child non-zero is returned after summary attempt. | stable through Make wrappers; direct scripts may bypass. | `machine_consumed` plus `temporary_log`. | `scripts/lib/run-phase.sh`; `scripts/lib/run-phase-common.sh`; `scripts/lib/test-output/cli.mjs`; `ART-0014`, `ART-0016` | `observed/source_limit` | Quiet mode stores logs and prints bounded summaries; verbose/debug streams through `tee`. |
| OI-0002 | `EP-0003`, `EP-0015` | Target summary projection from child targets: `target-summary.json`, `tool-run-summary.json`, `[RESULT]`, `[ARTIFACTS]`, `[FAIL]`, `[RERUN]`, `[INVESTIGATE]`. | Aggregates, check wrappers, release gates, humans, explain tools. | yes | `SCHEMA-0003`, `SCHEMA-0006`. | Target summary is emitted after child target or summary projection completes; missing children become artifact/helper failures in summary logic. | stable when invoked through `TEST_OUTPUT_SCRIPT`; package tools do not guarantee it. | `machine_consumed` and `durable_report`. | `scripts/cartulary-runner.mjs`; `scripts/lib/test-output/cli.mjs`; `scripts/lib/tool-output.mjs`; `ART-0014` | `observed` | Failure terminal lines are stderr in human modes; machine mode emits JSON on stdout. |
| OI-0003 | `EP-0008` | Aggregate sequence run output: run-start/step-start human lines, child Make output, `run-summary.json`, aggregate `target-summary.json`, `tool-run-summary.json`, skipped-after-failure fields. | `make test-fast`, `make test`, `make ci`, `make release-check`, humans, CI, explain tools. | yes | `SCHEMA-0004`, `SCHEMA-0006`. | Sequence manifest order controls serial/parallel steps; first failing step emits fail run/target summary and exits with child status. | stable through Make sequence surface; provider CI workflow source absent. | `machine_consumed` and `durable_report`. | `scripts/run-make-sequence.sh`; `tools/task_surface_manifest.json`; `scripts/lib/test-output/cli.mjs`; `ART-0013`, `ART-0014` | `observed/source_limit` | Broad sequence runtime was not rerun during S5. |
| OI-0004 | `EP-0007` | Check scheduler output: `[CHECK-SCHEDULER]`, `[PROGRESS]`, child logs on failure/verbose, `scheduler-summary.json`, `scheduler-events.jsonl`, `progress-summary.log`, scheduler logs dir, target/run summaries. | `make check`, `make ci`, `make release-check`, scheduler drift, explain tools, humans. | yes | `SCHEMA-0005`, `SCHEMA-0006`; schema IDs `cartulary.check_scheduler_summary.v9` and `cartulary.check_scheduler_event.v5`. | Manifest validation precedes scheduling; `stopOnFirstFailure=true`; failed units cause dependent skips while siblings may drain. | stable when generated schedules are current; concrete host capacity remains S6. | `machine_consumed` and `durable_report`. | `scripts/run-check-schedule.mjs`; `scripts/lib/scheduler/**`; `tools/check_schedule_manifest.json`; `ART-0015`, `RES-0001..RES-0010` | `observed/source_limit` | Check scheduler cleanup/session stop behavior is source-observed only. |
| OI-0005 | `EP-0006`, `EP-0012` | Service-backed and phase-slice scheduler output: scheduler start/progress/summary lines, events, summaries, logs, child target summaries, browser/session cleanup summaries. | Service-backed Make targets, phase slices, check nested scheduler, drift tools, explain tools. | yes | `SCHEMA-0005`, `SCHEMA-0006`; schema IDs `cartulary.service_backed_scheduler_summary.v9`, `cartulary.service_backed_scheduler_event.v5`. | Work-unit sources expand before scheduling; `stopOnFirstFailure=false` for service-backed scheduler, with finalizers and deferred summaries where configured. | stable through Make/scheduler; package bypass unknown. | `machine_consumed` and `durable_report`. | `scripts/run-service-backed-schedule.mjs`; `tools/service_backed_schedule_manifest.json`; `tools/browser_e2e_batch_manifest.json`; `ART-0015` | `observed/source_limit` | Runtime service readiness and finalizer cleanup remain source-limited. |
| OI-0006 | `EP-0004`, `EP-0005`, `EP-0006` | Go target runner artifacts: shard metadata, `runner.jsonl`, `stderr.log`, `exit_status.txt`, phase summaries, target summaries, shard/finalizer outputs. | Go target aggregates, service-backed scheduler, target/run summaries, humans. | yes | `SCHEMA-0001`, `SCHEMA-0007`, `SCHEMA-0011`. | Go runner writes raw artifacts, then test-output summarizes phases; shard finalizer reads captured shard metadata before aggregate summary. | stable through Make runner; direct uninvoked script is source-limited. | `machine_consumed` plus `failure_only_artifact`. | `scripts/lib/go-target-runner.mjs`; `scripts/lib/run-go-phase.sh`; `scripts/lib/test-output/cli.mjs`; `ART-0016` | `observed/source_limit` | Controlled Go failures were not generated during S5. |
| OI-0007 | `EP-0019`, `EP-0002` | Frontend unit/Vitest artifacts: `runner.json`, stdout/stderr logs, watchdog JSON on timeout, derived manifest phase summaries, target summaries. | `make frontend-unit`, phase slices, humans, explain tools. | yes | `SCHEMA-0001`, `SCHEMA-0008`, `SCHEMA-0012`; watchdog schema `cartulary.vitest_watchdog.v1`. | Vitest raw run executes once, then manifest-aware summaries are derived per phase and residual slice. | stable through Make wrapper; direct Vitest package scripts bypass result-root policy. | `machine_consumed` and `failure_only_artifact`. | `scripts/run-frontend-unit.sh`; `scripts/lib/run-phase-common.sh`; `scripts/lib/test-output/cli.mjs`; `apps/web/vite.config.ts` | `observed/source_limit` | Watchdog timeout behavior is source-observed, not runtime-observed in this pass. |
| OI-0008 | `EP-0009`, `EP-0010`, `EP-0018` | Browser E2E target output: stack env/json, server/web logs, Playwright stdout/stderr, Playwright report/test-results/traces/screenshots/videos, child target summaries. | Browser wrappers, service-backed scheduler, Playwright humans, explain tools. | mixed | `SCHEMA-0009`, `SCHEMA-0010`, `SCHEMA-0013`; Playwright tool report/trace schema is `TODO: schema_unknown`. | Owned stack prepares services before child tests; browser batch can reset before groups and emits child summaries after group execution. | differs between Make-owned stack and direct package scripts. | `machine_consumed`, `temporary_log`, `failure_only_artifact`. | `scripts/start-web-e2e.sh`; `scripts/run-browser-e2e-target.sh`; `scripts/run-browser-e2e-batch.sh`; Playwright configs; `ART-0018`, `ART-0027` | `observed/source_limit` | Browser runtime and failure bundle completeness were not executed in S5. |
| OI-0009 | `EP-0009`, `EP-0010`, `EP-0020` | Reset boundary output: response JSON, HTTP status file, state-reset marker, stderr on non-200 or invalid JSON. | Browser batch reset step, humans, future failure analysis. | yes for reset wrapper validation | `SCHEMA-0010`; response data schema marker `cartulary.test.runtime_reset.v1`. | Reset script writes response and status before validation; Playwright state marker is written only after valid reset and configured state dir cleanup. | only when test route is enabled by owned backend. | `machine_consumed` with `authority_ambiguous`. | `scripts/reset-web-e2e-stack.sh`; `internal/app/test_runtime_reset.go`; `SVC-0014`; `AMB-0006` | `observed/source_limit` | Route authority, route-specific readiness, and partial mutation behavior remain open. |
| OI-0010 | `EP-0011`, `EP-0005`, `EP-0006`, `EP-0007` | `tools/testservices` output: stderr usage/failure lines, suite event JSON files, `service-scope.json`, service lease JSON, reaper log, fixture summaries. | Service-backed wrappers, fixture report, target summaries, cleanup decisions, humans. | yes | `SCHEMA-0014`, `SCHEMA-0015`, `SCHEMA-0016`; lease schema `cartulary.test_services.lease.v1`. | Events are appended during preflight/start/use/cleanup; `RefreshSummary` rewrites `service-scope.json`; child exit propagates after cleanup scheduling. | stable when service wrapper is used; live Docker behavior source-limited. | `machine_consumed`, `durable_report`, `temporary_log`. | `tools/testservices/main.go`; `internal/testutil/suiteservices/diagnostics.go`; `ART-0017`, `ART-0029`; `SVC-0001` | `observed/source_limit` | Reaper scheduling is observable; detached reaper completion is not. |
| OI-0011 | `EP-0014` | Investigation/node-tool output: human reports or `tool-run-summary.json`, optional machine JSON, selected artifact/run paths. | Agents, developers, explain/drift/baseline tooling, future NLSpec drafting. | mixed | `SCHEMA-0006`, `SCHEMA-0017`; tool-specific schemas vary. | Tool wrapper validates args before reading selected results; some tools require `RESULTS_DIR`, others inspect current/default roots. | stable for Make node-tool targets; retained artifact freshness remains source-limited. | `machine_consumed` for JSON and `human_diagnostic` for text. | `scripts/run-make-node-tool.mjs`; `scripts/lib/make-node-tools.mjs`; `scripts/lib/artifact-discovery.mjs`; `ART-0013..ART-0017` | `observed/source_limit` | Ambient newest-run behavior must remain investigation-only unless owner decides otherwise. |
| OI-0012 | `EP-0013` | Harness smoke output: logical check stdout/stderr, target summary, tool summary, skipped checks after first failure. | Harness self-tests, check/CI/release gates, humans. | yes | `SCHEMA-0003`, `SCHEMA-0006`. | Smoke runner runs checks with bounded jobs and stops scheduling after first failure; summary/target summary emitted after checks. | stable through Make smoke targets. | `machine_consumed` and `human_diagnostic`. | `scripts/run-harness-smoke.mjs`; `tools/task_surface_manifest.json` harness checks; S2 `47` checks | `observed/source_limit` | S5 did not execute smoke tiers. |
| OI-0013 | `EP-0016`, `EP-0018`, `EP-0019` | Direct package-script output: pnpm/Vite/Vitest/Biome/Playwright stdout/stderr and tool-default artifacts. | Direct developer users and package tooling. | mixed/unknown | `SCHEMA-0018`; package tool defaults, `ART-0030`. | Package manager runs the declared script; Make run ID/result-root/scheduler policy is not guaranteed unless child re-enters a harness wrapper. | not stable against Make behavior until S8 decision. | `authority_ambiguous`. | root `package.json`; `apps/web/package.json`; S2 package table; `AMB-0011`, `AMB-0020` | `observed` | Default S5 treatment is developer convenience except wrapper re-entry. |
| OI-0014 | `EP-0017`, `EP-0008` | CI-adjacent output: `scripts/ci/verify.sh` sets `CARTULARY_OUTPUT_MODE=ci` if unset and execs `make ci`; deployable-shape script writes success/failure text. | External CI provider if configured, humans, release/check gates. | yes only through Make summaries; provider annotations unknown. | Make CI summaries use `SCHEMA-0004`, `SCHEMA-0006`; provider CI annotation schema is `TODO: schema_unknown`. | `verify.sh` delegates completely to `make ci`; deployable-shape exits `1` on shape failures and prints to stderr, success to stdout. | provider workflow unknown because `.github/**` is absent. | `authority_ambiguous` plus `machine_consumed` through Make. | `scripts/ci/**`; `.github absent` check; `SL-0001`, `AMB-0001` | `observed/source_limit` | No `::error`/provider annotation source was found in repository CI files because workflow files are absent. |
| OI-0015 | `EP-0020` | Maintenance/cleanup output: generated/format/baseline/release/clean/distclean command stdout/stderr, target summaries where wrapped, mutated/deleted files or ignored paths. | Developers, agents, release/check maintenance. | mixed | Generated output schemas are downstream of owning tools; cleanup reports are mostly human text unless wrapped by Make summaries. | Command-specific; writer/cleanup targets were not run during S5. | Make surface stable, command effects source-limited. | `authority_ambiguous` and `human_diagnostic`. | `Makefile`; generated recipes; `ART-0006..ART-0012`, `ART-0020..ART-0024`; `AMB-0013` | `observed/source_limit` | S5 records output surfaces only; no writer or cleanup command was executed. |
| OI-0016 | `EP-0001`, `EP-0014` | Help, task-guide, explain-target, explain-phase, target-plan, target-plan-json output. | Developers, agents, CI investigators. | mixed | Human text by default; JSON variants use tool-specific schema or `TODO: schema_unknown`. | No child execution; commands inspect manifests/plans/artifacts. | stable for local discovery; not CI-provider specific. | `human_diagnostic` and `machine_consumed` for JSON variants. | `make help-all`; `scripts/print-task-guide.mjs`; `scripts/print-target-plan.mjs`; S2 discovery | `observed/runtime_observed` | These were used as non-mutating recovery inputs in S1/S2. |
| OI-0017 | `EP-0021` | Uninvoked candidate `scripts/test-run-go-target-fast.sh`. | Unknown/manual only. | unknown | `TODO: schema_unknown`; no active task-surface row. | No discovered active caller. | unknown. | `authority_ambiguous`. | `uninvoked-surface-list.md`; S2 `EP-0021` | `observed/source_limit` | Keep out of active lifecycle/failure contracts unless owner revives it. |

## Coverage notes

- Every S2 entrypoint row from `EP-0002` through `EP-0021` is represented by
  at least one `OI-*` row.
- Every row with `Machine-consumed=yes` links to a `SCHEMA-*` note or an
  explicit `TODO: schema_unknown`.
- CI annotations remain source-limited because `.github/**` is absent and no
  provider annotation source was discovered in `scripts/ci/**`.
- Retained run artifacts require explicit run identity for drift/baseline use.
  Newest-run fallback remains human investigation only until an owner decision.

## Checklist Support

| Field | Value |
|---|---|
| Status | `complete` |
| Blockers | CI provider annotations, controlled failure bundles, broad-gate runtime output, and service/browser/reset/cleanup runtime examples were unavailable. |
| Findings | `OI-0001` through `OI-0017` cover major S2 entrypoint families, output surfaces, consumers, schema links, ordering notes, local/CI stability, authority class, and evidence. |
| Source limits | `.github/**` absence, unexecuted service/browser/Docker/Compose/reset/cleanup behavior, failure-only artifact schemas, stale retained artifacts, and uninvoked `EP-0021`. |
| Ambiguities | Reset route, direct package scripts, local-dev commands, CI provider annotations, maintenance/cleanup authority, and package-script artifacts outside Make result roots. |
| Handoff notes | S6 can consume source-limited runtime rows as probe candidates; S7 can consume stable output inventory; S8 must resolve authority-ambiguous rows. |
| Evidence status | Rows use `observed`, `runtime_observed`, `observed/source_limit`, or `source_limit`. |
| No-change confirmation | This file records recovery documentation only and does not alter harness behavior. |
