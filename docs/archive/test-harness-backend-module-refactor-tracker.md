# test-harness-backend> Module Refactoring Tracker and Handoff

Last updated: 2026-07-04T04:49:14Z.

## 1. Scope and Authority

- Target path: `tools/harness/backend`.
- Target label: `test-harness-backend>`.
- Normalized target ID: `test-harness-backend`.
- Controlling plan artifact: this tracker.
- Canonical harness authority: `docs/testing-harness-nlspec.md`.
- Domain-vocabulary reference: `docs/domain.md`; this target currently owns no
  product-domain, workbook, frontend, route, or WebSocket behavior.
- Public compatibility surface: public Make targets, stable `command_id` values,
  schema IDs, retained artifact paths, output shape, failure mapping, cleanup
  behavior, and declared public input contracts.
- Private helper policy: private helper compatibility is not retained unless the
  Testing Harness NLSpec declares an owner facade.

## 2. Current Backend File Inventory

Current files under `tools/harness/backend` after the backend target execution
split:

- `backend-target-execution.mjs`
- `backend-duration-accounting.mjs`
- `backend-shard-plan.mjs`
- `backend-target-plan.mjs`
- `build-go-artifact.sh`
- `go-duration-artifacts.mjs`
- `go-duration-baselines.mjs`
- `go-shard-plan-cli.mjs`
- `go-shard-plan.mjs`
- `go-target-aggregate.mjs`
- `go-target-plan-coverage-cli.mjs`
- `go-target-runner.mjs`
- `go-test-duration-baseline-coverage-cli.mjs`
- `go-test-duration-baseline-drift-cli.mjs`
- `go-test-duration-baselines-cli.mjs`
- `postgres-fixture-budget-cli.mjs`
- `run-go-manifest-phase.sh`
- `run-go-phase.sh`
- `target-plan.mjs`
- `target-execution/capture.mjs`
- `target-execution/cli.mjs`
- `target-execution/command.mjs`
- `target-execution/context.mjs`
- `target-execution/planning.mjs`
- `target-execution/reports.mjs`
- `target-execution/runtime-binaries.mjs`
- `target-execution/summary-emission.mjs`
- `target-execution/targets.mjs`
- `target-execution/util.mjs`
- `tests/test-check-migrations.sh`
- `tests/test-go-test-duration-baselines.sh`
- `tests/test-run-go-target.sh`

The previously tracked backend legacy helper paths for govulncheck findings,
migration/schema drift validators, backend drift compatibility exports,
duration compatibility exports, and runner shard compatibility exports are not
present in the live backend tree. Their owner moves are complete.

## 3. Resolved Gaps

| ID | Gap | Current status | Evidence |
| --- | --- | --- | --- |
| RB-001 | Govulncheck findings ownership | DONE | Helper lives at `tools/harness/static-analysis/govulncheck-findings.mjs`; old backend path is unsupported private. |
| RB-002 | Migration/schema drift ownership | DONE | Helpers live under `tools/harness/generated-artifacts/database-contract-drift/`. |
| RB-003 | Unsupported compatibility re-exports | DONE | `drift/manifests.mjs`, `duration/baselines.mjs`, and `runner/go-shards.mjs` are absent from backend. |
| RB-004 | Import-boundary guardrails for backend facades | DONE | `harness-import-boundary.mjs` rejects unsupported private helper paths and non-owner backend internals. |
| RB-005 | Duration retained-run eligibility | DONE AS RULE | NLSpec 4.1A and 4.1 duration requirements define retained-run eligibility; no baseline update target is part of this run. |

## 4. Remediation Outcomes

| ID | Gap | Remediation | Area | Status | Validation |
| --- | --- | --- | --- | --- | --- |
| RG-001 | Tracker current-state drift | Tracker updated before code movement and after each workstream. | Documentation | DONE | `make lint-markdown` passed. |
| RG-002 | Missing backend target execution facade | NLSpec 4.1A declares `backend_target_execution`; import-boundary allowlist includes the facade. | Specification, tests | DONE | `make lint-markdown`; `make harness-contract`. |
| RG-003 | Mixed-concern `go-target-runner.mjs` | Runner is now a CLI shim; implementation lives behind `backend-target-execution.mjs` and `target-execution/*`. | Implementation, tests | DONE | `make lint-scripts`; `make check-harness-smoke`; backend target runs; `make check`. |
| RG-004 | Direct runner helper imports | Diagnostics and smoke helper imports use `backend-target-execution.mjs`; CLI path remains executable only. | Implementation, tests | DONE | `test-run-go-target.sh`; coverage CLI assertions; import-boundary fixtures. |
| RG-005 | Final validation and handoff | Narrow, backend public, finalizer, and post-finalizer broad checks were run and logged. | Tests, documentation | DONE | Section 9 records commands and run roots. |

## 5. Workstreams

| Workstream | Status | Depends on | Exit criteria |
| --- | --- | --- | --- |
| WS-00 Tracker reconciliation | DONE | none | Current inventory, resolved gaps, remaining gaps, and slice plan are recorded; `make lint-markdown` passes. |
| WS-01 Spec and guardrail ownership | DONE | WS-00 | NLSpec declares `backend_target_execution`; import-boundary fixtures cover facade and private runner denial. |
| WS-02 Execution facade and caller migration | DONE | WS-01 | `backend-target-execution.mjs` exists; diagnostics/tests import it; `go-target-runner.mjs` still works as CLI. |
| WS-03 Low-risk runner extraction | DONE | WS-02 | Context, command, planning, util, and runtime-binary helpers are split under `target-execution/`. |
| WS-04 Capture, reporting, and finalization extraction | DONE | WS-03 | Capture, reports, summary emission, target orchestration, and CLI dispatch are split; runner file is thin. |
| WS-05 Cleanup and boundary tightening | DONE | WS-04 | No in-repo helper imports depend on `go-target-runner.mjs`; guardrails allow only declared facades for non-backend callers. |
| WS-06 Validation and handoff completion | DONE | WS-05 | Required validation is run or explicitly skipped with reason; final handoff row is complete. |

## 6. Implementation Slice Plan

| Slice | Workstream | Change | Contract posture | Required validation |
| --- | --- | --- | --- | --- |
| S-00 | WS-00 | Reconcile this tracker. | Documentation only. | `make lint-markdown`. |
| S-01 | WS-01 | Add `backend_target_execution` owner facade to NLSpec and static import-boundary checks. | No public harness behavior change. | `make lint-markdown`; `make lint-scripts`; `make harness-contract`. |
| S-02 | WS-02 | Add `backend-target-execution.mjs` facade and migrate diagnostics/tests to it. | CLI script path remains `go-target-runner.mjs`. | `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`. |
| S-03 | WS-03 | Extract context, command, planning, utility, and runtime-binary modules. | Preserve command text, artifact names, target names, shard names, and exit behavior. | `make lint-scripts`; `make check-harness-smoke`. |
| S-04 | WS-04 | Extract capture, reports, summary emission, targets, and CLI modules. | Preserve phase summaries, target summaries, timing spans, locks, reuse semantics, and failure classes. | `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`. |
| S-05 | WS-05 | Tighten cleanup and guardrails after all callers move. | Generated task-surface/topology files are not hand-edited. | `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract`. |
| S-06 | WS-06 | Run final validation and update handoff. | No public contract change. | Narrow checks, backend target checks, `make check`, and `make agent-finalize` unless skipped with reason. |

## 7. Public Contract Freeze

The following must remain unchanged unless the Testing Harness NLSpec is updated
first:

- Public Make targets `backend-unit`, `backend-store`,
  `backend-integration`, and `backend-process`.
- Check-internal helper target `backend-integration-support`.
- Stable `command_id` values for backend public targets.
- Schema IDs including `cartulary.go_shard_plan.v3`,
  `cartulary.test_phase_summary.v3`, and `cartulary.tool_run_summary.v3`.
- Retained artifact paths for target summaries, phase summaries, timing spans,
  shared Go reports, runtime-binary metadata, and build-artifact cache records.
- Output budgets, failure classes, failure reasons, redaction behavior, lock
  timeout diagnostics, run-root identity protection, and cleanup behavior.

## 8. Validation Plan

Run the narrowest commands that cover each slice, then broaden after the runner
split:

- `make lint-markdown`
- `make lint-scripts`
- `make lint-shell`
- `make check-harness-smoke`
- `make harness-contract`
- `make backend-module-boundary-check`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make backend-unit`
- `make backend-store`
- `make backend-integration`
- `make backend-process`
- `make check`
- `make agent-finalize`

Do not run `make go-test-duration-baselines` unless a retained successful
`RESULTS_DIR` is explicitly selected because that target mutates baseline JSON.

## 9. Execution Log

| Time | Workstream | Files changed | Commands | Result | Next action |
| --- | --- | --- | --- | --- | --- |
| 2026-07-04T04:29:48Z | WS-00 | This tracker. | `git status --short`; backend inventory and import scans. | Tracker reconciled before code movement. | Run `make lint-markdown`, then start WS-01. |
| 2026-07-04T04:35:03Z | WS-00 | This tracker. | `make lint-markdown`. | PASS. | Start WS-01. |
| 2026-07-04T04:37:02Z | WS-01 | `docs/testing-harness-nlspec.md`; `tools/harness/static-analysis/harness-import-boundary.mjs`; `tools/harness/tests/test-harness-contracts.mjs`; this tracker. | `make lint-markdown`; `make lint-scripts`; `make harness-contract`. | PASS. | Start WS-02. |
| 2026-07-04T04:39:20Z | WS-02 | `tools/harness/backend/backend-target-execution.mjs`; `tools/harness/backend/go-target-plan-coverage-cli.mjs`; `tools/harness/backend/tests/test-run-go-target.sh`; this tracker. | `make lint-scripts`; `make harness-contract`; `make check-harness-smoke`. | PASS. | Start WS-03. |
| 2026-07-04T04:46:12Z | WS-03/WS-04 | `tools/harness/backend/target-execution/*`; `tools/harness/backend/backend-target-execution.mjs`; `tools/harness/backend/go-target-runner.mjs`; this tracker. | `node --check` for new modules and runner/facade; `make lint-scripts`; `make check-harness-smoke`; `make harness-contract`. | PASS. | Start WS-05 cleanup and boundary checks. |
| 2026-07-04T04:49:20Z | WS-05 | Static import-boundary and generated-artifact posture; this tracker. | `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract`. | PASS. Run roots: `.cartulary/test-results/20260704T043936Z-p3199705`, `.cartulary/test-results/20260704T043936Z-p3199756`, `.cartulary/test-results/20260704T043936Z-p3199746`. | Start WS-06 final validation. |
| 2026-07-04T04:49:14Z | WS-06 | All refactor files plus finalizer-updated baseline/generated files. | `make lint-markdown`; `make lint-scripts`; `make lint-shell`; `make check-harness-smoke`; `make harness-contract`. | PASS. | Run backend public targets. |
| 2026-07-04T04:49:14Z | WS-06 | Backend target execution split. | `make backend-unit`; `make backend-store`; `make backend-integration`; `make backend-process`. | PASS. Run roots: `.cartulary/test-results/20260704T044024Z-p3203504`, `.cartulary/test-results/20260704T044038Z-p3205775`, `.cartulary/test-results/20260704T044118Z-p3214990`, `.cartulary/test-results/20260704T044210Z-p3225138`. | Run broad check. |
| 2026-07-04T04:49:14Z | WS-06 | Full harness worktree before finalizer refresh. | `make check`. | PASS. Run root: `.cartulary/test-results/20260704T044308Z-p3236748`; 255/255 work units; 981 tests; 0 failed. | Run `agent-finalize` with retained check root. |
| 2026-07-04T04:49:14Z | WS-06 | `tools/browser_e2e_duration_baselines.json`; `tools/execution_topology_render_index.json`; `tools/go_test_duration_baselines.json`; `tools/harness_smoke_duration_baselines.json`; `tools/scheduler_manifest.json`; `tools/service_backed_make_target_duration_baselines.json`. | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260704T044308Z-p3236748`. | PASS. Run root: `.cartulary/test-results/20260704T044535Z-p3314661`; generated updated file count: 6. | Rerun broad check after finalizer mutations. |
| 2026-07-04T04:49:14Z | WS-06 | Final worktree after finalizer refresh. | `make check`. | PASS. Run root: `.cartulary/test-results/20260704T044622Z-p3316924`; 259/259 work units; 981 tests; 0 failed. | Final markdown lint and handoff. |
| 2026-07-04T04:49:14Z | WS-06 | This tracker. | `make lint-markdown`. | PASS. | Handoff complete. |

## 10. Final Handoff Checklist

- DONE: tracker has no stale open ownership blockers.
- DONE: NLSpec facade registry includes backend target execution.
- DONE: non-backend callers use owner facades, not runner internals.
- DONE: `go-target-runner.mjs` is a thin executable CLI shim.
- DONE: validation commands and outcomes are logged in Section 9.
- DONE: no required checks were skipped.
- DONE: retained-run maintenance used `.cartulary/test-results/20260704T044308Z-p3236748`.
