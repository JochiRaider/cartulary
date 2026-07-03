# Scripts Facade Elimination Tracker, Iteration 3

## Iteration 4 Remediation Baseline

Iteration 4 implements the end-to-end remediation plan accepted on
2026-07-03. This iteration treats `docs/handoffs/scripts-module-refactor-tracker.md`
as the controlling handoff artifact and updates it after each completed
workstream before starting the next workstream.

Scope boundary:

- Keep `scripts/harness-contract.sh` as the NLSpec-frozen public compatibility
  wrapper unless `docs/testing-harness-nlspec.md` is revised first.
- Keep `scripts/ci/verify.sh` as the external provider-neutral CI entrypoint.
- Retire other root non-test `scripts/*` implementation paths after moving
  callers to owner paths and regenerating task-surface/topology outputs.
- Keep active root `scripts/test-*` files as characterization tests unless a
  later test-layout owner plan moves them.
- Do not update historical recovery/archive docs unless they are republished as
  active guidance.

| Item | Iteration 4 live rebaseline |
| --- | --- |
| Working tree at scan | clean |
| `scripts/` files | 113 total |
| Root files | 109 |
| Files under `scripts/ci/` | 4 |
| Root `scripts/test-*` files | 52 |
| Root non-test implementation/wrapper files | 57 |
| Owner roots used by this iteration | `tools/harness/backend`, `tools/harness/browser`, `tools/harness/core`, `tools/harness/frontend`, `tools/harness/generated-artifacts`, `tools/harness/planning`, `tools/harness/readiness`, `tools/harness/scheduler`, `tools/release-evidence` |

Iteration 4 workstream order:

1. WS-01 Go shard-plan invariant.
2. WS-02 thin wrappers.
3. WS-03 planning/core diagnostics.
4. WS-04 browser owner extraction.
5. WS-05 static analysis, lint, security, and frontend tooling.
6. WS-06 generation, migrations, frontend evidence, and release/operator tooling.
7. WS-07 active test cleanup folded into each implementation slice.
8. WS-Z final regeneration, validation, and handoff.

### Iteration 4 Progress Log

| Workstream | Status | Files / owners touched | Validation | Notes |
| --- | --- | --- | --- | --- |
| WS-00 Rebaseline and spec boundary | Complete | This tracker only. | Live inventory: 113 `scripts/` files, 109 root files, 4 CI files, 52 root `test-*` files. | `scripts/harness-contract.sh` and `scripts/ci/verify.sh` remain intentionally public wrappers. |
| WS-01 Go shard-plan invariant | Complete | `tools/harness/backend/go-shard-plan.mjs`, `tools/go_test_duration_baselines.json`, generated scheduler artifacts. | `bash scripts/test-print-target-plan.sh` PASS; `make target-plan-json` PASS (`rows=313`); `make explain-target TARGET=backend-store DETAIL=rows` PASS; `node --check tools/harness/backend/go-shard-plan.mjs` PASS; `make phase-schedules` PASS (`.cartulary/test-results/20260703T134311Z-p613915`); `make check-harness-smoke` PASS. | Aggregate diagnostic weight now sums underlying work-item weights so overhead does not falsely classify singleton aggregates as heavy; shard scheduling weights still include command/package overhead. Backend integration shard target restored to `18000`; backend store target remains `9000` to avoid unrelated scheduler churn. |
| WS-02 Thin wrappers | Complete | `tools/harness/core/run-make-node-tool-cli.mjs`, `tools/harness/core/run-make-node-tool.sh`, `tools/harness/frontend/frontend-toolchain.sh`, task-surface/topology owner inputs and generated outputs. | `make phase-schedules` PASS (`.cartulary/test-results/20260703T134649Z-p617398`); `bash scripts/test-make-node-tools.sh` PASS; `bash scripts/test-run-make-sequence.sh` PASS; `bash scripts/test-execution-topology.sh` PASS; `make frontend-toolchain` PASS (`.cartulary/test-results/20260703T134738Z-p619481`); `make service-backed-unit-check` PASS; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` PASS; `make check-harness-smoke` PASS; exact non-tracker deleted-wrapper scan PASS. | Deleted `scripts/check-service-backed-unit-tests.sh` and `scripts/print-frontend-toolchain.sh`; moved Make-node launcher implementation from root `scripts/` into core owner paths. |
| WS-03 Planning/core diagnostics | Complete | Planning CLIs under `tools/harness/planning`, core reporting CLIs under `tools/harness/core`, backend duration CLIs under `tools/harness/backend`, scheduler duration/drift CLIs under `tools/harness/scheduler`, task-surface/topology owner inputs and generated outputs. | `node --check` moved CLI files PASS; shell syntax PASS; `make phase-schedules` PASS (`.cartulary/test-results/20260703T135321Z-p623878`); `make target-plan` PASS; `make target-plan-json` PASS; `make task-guide ROLE=feature-dev PHASE=phase1` PASS; `make explain-phase PHASE=phase1` PASS; `make explain-target TARGET=backend-store` PASS; `make fixture-report RESULTS_DIR=.cartulary/test-results/20260703T043250Z-p313996` PASS; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` PASS; `make phase-schedule-drift` PASS (`.cartulary/test-results/20260703T135403Z-p624605`); `bash scripts/test-print-target-plan.sh` PASS; `bash scripts/test-run-phase-slice.sh` PASS; `bash scripts/test-task-surface-report.sh` PASS; `bash scripts/test-check-phase-test-names.sh` PASS; duration/scheduler direct self-tests PASS; `make check-harness-smoke` PASS; exact non-tracker deleted-path scan PASS. | Later WS-07 cleanup fixed the previously noted direct `scripts/test-run-phase.sh` missing-child classification fixture. |
| WS-04 Browser owner extraction | Complete | Browser target/batch/stage/visual/a11y wrappers and `start-web-e2e.sh` under `tools/harness/browser`; scheduler browser session defaults; browser task-surface/topology owner inputs and generated outputs; `apps/web` Playwright/package entrypoints. | Shell syntax PASS; `node --check` browser/scheduler/generator touched files PASS; `make phase-schedules` PASS (`.cartulary/test-results/20260703T140045Z-p640760`); `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` PASS; `bash scripts/test-web-e2e-lifecycle.sh` PASS; `bash scripts/test-run-playwright-webserver-batch.sh` PASS; `bash scripts/test-run-playwright-phase.sh` PASS; `bash scripts/test-run-playwright-manifest-phase.sh` PASS; `make check-harness-smoke` PASS; `make browser-e2e-a11y-preflight` PASS (`.cartulary/test-results/20260703T140154Z-p649799`); `make phase-schedule-drift` PASS (`.cartulary/test-results/20260703T140258Z-p662079`); exact non-tracker deleted-path scan PASS. | Deleted the root browser implementation cluster. Heavier `browser-e2e-webserver-backed` and `browser-e2e-visual` were not run because the public a11y preflight exercised the moved browser stack and wrapper path with lower cost. |
| WS-05 Static analysis, lint, security, and frontend tooling | Complete | Static-analysis wrappers under `tools/harness/static-analysis`; frontend unit wrapper under `tools/harness/frontend`; lint/security/generated-policy owner inputs; `apps/web` package entrypoints; active static-analysis tests. | Shell/Node syntax PASS; `make phase-schedules` PASS (`.cartulary/test-results/20260703T140600Z-p663541`); `make generated-artifact-policy-check` PASS (`.cartulary/test-results/20260703T141222Z-p693758`); `make lint-shell` PASS; `make lint-scripts` PASS; `make lint-biome` PASS; `make frontend-unit` PASS (`.cartulary/test-results/20260703T141222Z-p693706`); `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260703T141222Z-p693747`); `make backend-module-boundary-check` PASS (`.cartulary/test-results/20260703T141035Z-p682332`); `make go-vulncheck` PASS (`.cartulary/test-results/20260703T141035Z-p682280`); `make go-gosec-targeted` PASS (`.cartulary/test-results/20260703T141035Z-p682304`); `make go-gosec-audit` PASS; direct static-analysis self-tests PASS; exact non-tracker deleted-path scan PASS. | `tools/harness/static-analysis/shellcheck.sh` now inventories untracked, non-ignored shell files as well as tracked files so new moved owner scripts are linted before commit. Later WS-07 cleanup fixed the previously noted direct `scripts/test-run-frontend-unit.sh` phase-slice fixture. |
| WS-06 Generation, migrations, frontend evidence, release/operator | Complete | Generation orchestration under `tools/harness/generated-artifacts`; migration check under `tools/harness/backend`; toolchain pin check under `tools/harness/readiness`; embedded web assets and frontend evidence audit under `tools/harness/frontend`; Make sequence and harness smoke under `tools/harness/core`; release/operator checks under `tools/release-evidence`; task-surface/topology owner inputs and generated outputs. | Shell/Node syntax PASS; `make phase-schedules` PASS (`.cartulary/test-results/20260703T141706Z-p706717`); `make generate` PASS (`.cartulary/test-results/20260703T141814Z-p717944`); `make generate-drift` PASS (`.cartulary/test-results/20260703T141911Z-p722972`); `make migration-drift` PASS (`.cartulary/test-results/20260703T141814Z-p717952`); `make toolchain-drift` PASS (`.cartulary/test-results/20260703T141814Z-p717943`); `make deployable-shape` PASS against existing local build artifacts; `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` PASS; `make check-harness-smoke` PASS; direct owner self-tests PASS; exact non-tracker deleted-path scan PASS. | `scripts/ci/verify.sh` remains the only retained `scripts/ci` entrypoint. `make frontend-evidence-audit` was not runnable without required evidence inputs and exited with usage (`PHASE_NAMESPACE is required`); `bash scripts/test-frontend-evidence-audit.sh` covered the moved CLI fixtures. `make release-check` was not run because it is the full release gate (`check`, security audit, SBOM/license, SeaweedFS, build, deployable shape); the moved release helper was covered by `make deployable-shape` and task-surface tests. |
| WS-07 Active test cleanup | Complete | `scripts/test-run-frontend-unit.sh`, `tools/harness/frontend/run-frontend-unit.sh`, `scripts/test-run-phase.sh`, `tools/harness/core/test-output/cli.mjs`, affected tracker rows. | `bash scripts/test-run-frontend-unit.sh` PASS; `make frontend-unit` PASS (`.cartulary/test-results/20260703T142347Z-p738539`); `node --check tools/harness/core/test-output/cli.mjs` PASS; `bash scripts/test-run-phase.sh` PASS; `make check-harness-smoke` PASS. | Fixed the fake Vitest runner so phase-slice `-t` filters are honored, disabled frontend-row accounting by default for base phase slices, and preserved artifact classification for missing child summaries in aggregate rollups. |

## Current Baseline And Inventory Method

Iteration 3 continues the `scripts/` refactor by treating raw root
`scripts/*` paths as implementation details unless the harness NLSpec, public
Make/task-surface contract, operator use, or release workflow keeps them
public. This iteration removes thin compatibility facades whose owner paths are
now clear. It does not create `internal/modules/scripts`, `packages/scripts`,
or any product-facing scripts module.

Controlling historical input: Iteration 2 in this file, rebaselined against
the live repository before edits.

| Item | Pre-slice rebaseline |
| --- | --- |
| Branch / commit | `main` / `54e926e4` |
| Working tree at scan | clean |
| `scripts/` files | 145 total |
| Root files | 141 |
| Files under `scripts/ci/` | 4 |
| Root `scripts/test-*` files | 52 |
| `scripts/lib/` | 0 files at rebaseline |
| Live paths referenced by `Makefile`, `tools/task_surface_manifest.json`, or `tools/execution_topology_manifest.json` | 131 |
| Live root paths not directly referenced by those three | 14: `scripts/check-font-bundle.mjs`, `scripts/check-go-target-plan-coverage.mjs`, `scripts/check-migration-history.mjs`, `scripts/check-phase-map.mjs`, `scripts/check-postgres-fixture-budget.mjs`, `scripts/check-schema-object-ownership.mjs`, `scripts/ci/verify.sh`, `scripts/diagnose-inotify.mjs`, `scripts/generate-design-tokens.mjs`, `scripts/harness-contract.mjs`, `scripts/harness-contract.sh`, `scripts/print-go-shard-plan.mjs`, `scripts/reset-web-e2e-stack.sh`, `scripts/run-browser-e2e-owned-stack.sh` |

| Item | Post-slice state in this worktree |
| --- | --- |
| `scripts/` files | 54 total |
| Root files | 53 |
| Files under `scripts/ci/` | 1 |
| Root `scripts/test-*` files | 52 |
| `scripts/lib/` | absent in the live tree; 0 files |
| Live paths referenced by `Makefile`, `tools/task_surface_manifest.json`, or `tools/execution_topology_manifest.json` | 52 active root test paths in manifests, plus the retained `scripts/harness-contract.sh` Make variable |
| Live paths not directly referenced by those three | `scripts/ci/verify.sh`, `scripts/harness-contract.sh` |
| Deleted root compatibility facades | 91 total since the Iteration 3 baseline; 59 in Iteration 4 |
| Exact deleted-facade caller scan | PASS: no non-archive, non-recovery live references outside this tracker |

Inventory commands used:

| Purpose | Command |
| --- | --- |
| Full file inventory | `find scripts -type f -o -type l \| sort` |
| Root-file inventory | `find scripts -maxdepth 1 -type f` |
| CI inventory | `find scripts/ci -maxdepth 1 -type f` |
| `scripts/lib` inventory | `find scripts/lib -mindepth 1 -print` |
| Size scan | `wc -l scripts/* scripts/ci/*` |
| Root reference scan | `rg -o -P "(?<![A-Za-z0-9_./-])(\\./)?scripts/[A-Za-z0-9_./-]+" Makefile tools/task_surface_manifest.json tools/execution_topology_manifest.json` |
| Active caller scan | `rg` over `Makefile`, `tools/*`, active docs, tests, `apps/*`, and `packages/*`, excluding `docs/archive/**`, `docs/recovery/**`, and `docs/testing-harness-spec-recovery-docs/**` |
| Exact deleted-facade scan | `rg -n -F -f <deleted-facade-list> Makefile scripts apps packages tools docs --glob '!docs/archive/**' --glob '!docs/recovery/**' --glob '!docs/testing-harness-spec-recovery-docs/**'` |

## Facade Elimination Matrix

| Workstream | Deleted root facades | Owner paths now called directly | Status |
| --- | --- | --- | --- |
| WS-01 readiness/build/cache | `scripts/bootstrap-go-tool.sh`, `scripts/bootstrap-node-runtime.sh`, `scripts/bootstrap-shellcheck.sh`, `scripts/build-go-artifact.sh`, `scripts/build-web-artifact.sh`, `scripts/cache-artifact.sh`, `scripts/frontend-install.sh`, `scripts/frontend-toolchain.sh`, `scripts/list-build-inputs.sh`, `scripts/playwright-install.sh` | `tools/harness/readiness/*`, `tools/harness/backend/build-go-artifact.sh`, `tools/harness/frontend/*`, `tools/harness/browser/playwright-install.sh` | Complete |
| WS-02 backend/planning CLIs | `scripts/check-go-target-plan-coverage.mjs`, `scripts/check-migration-history.mjs`, `scripts/check-phase-map.mjs`, `scripts/check-postgres-fixture-budget.mjs`, `scripts/check-schema-object-ownership.mjs`, `scripts/print-go-shard-plan.mjs`, `scripts/run-go-target.mjs` | `tools/harness/backend/*-cli.mjs`, `tools/harness/backend/go-target-runner.mjs`, `tools/harness/planning/phase-map-check-cli.mjs` | Complete |
| WS-03 core/scheduler trampolines | `scripts/agent-finalize.mjs`, `scripts/agent-finalize.sh`, `scripts/cartulary-runner.mjs`, `scripts/run-check-schedule.mjs`, `scripts/run-service-backed-schedule.mjs`, `scripts/harness-contract.mjs` | `tools/harness/core/*-cli.mjs`, `tools/harness/scheduler/*-cli.mjs` | Complete; `scripts/harness-contract.sh` retained |
| WS-04 frontend/browser/readiness facades | `scripts/check-font-bundle.mjs`, `scripts/generate-design-tokens.mjs`, `scripts/write-frontend-accessibility-summary.mjs`, `scripts/diagnose-inotify.mjs`, `scripts/reset-web-e2e-stack.sh`, `scripts/run-browser-e2e-owned-stack.sh` | `tools/harness/frontend/*`, `tools/harness/readiness/diagnose-inotify.mjs`, `tools/harness/browser/*` | Complete |
| WS-05 local-dev root facades | `scripts/check-doctor.sh`, `scripts/dev-services.sh`, `scripts/dev-stack.sh` | `tools/harness/readiness/check-doctor.sh`, `tools/harness/readiness/dev-services.sh`, `tools/harness/readiness/dev-stack.sh` | Complete |

Retained wrappers:

| Path | Disposition | Rationale | Migration impact |
| --- | --- | --- | --- |
| `scripts/harness-contract.sh` | `keep_public_wrapper` | Explicit stable wrapper named by `docs/testing-harness-nlspec.md`; fallback preflight remains useful when Node is unavailable. | Do not remove without first revising the harness NLSpec. |
| `scripts/ci/verify.sh` | `keep_external_operator_entrypoint` | Provider-neutral external CI entrypoint; no in-repo workflow was found, so external use remains source-limited. | External callers may continue using it. |

## Owner-Family Map

| Owner family | Durable path |
| --- | --- |
| Readiness/cache/local-dev | `tools/harness/readiness` |
| Backend harness | `tools/harness/backend` |
| Frontend harness | `tools/harness/frontend` |
| Browser harness | `tools/harness/browser` |
| Core harness/finalizer/output/schema | `tools/harness/core` |
| Scheduler | `tools/harness/scheduler` |
| Planning/task surface | `tools/harness/planning`, `tools/harness/generated-artifacts` |
| Release/Core 05 evidence | `tools/release-evidence` |
| Go service-test support | `tools/gotestservicecheck`, `tools/gotestinventory` |
| OTel support | `tools/otel` |

## Public-Contract And Freeze Map

| Surface | Frozen in this iteration |
| --- | --- |
| Public Make targets | Target names and public behavior preserved. |
| Command IDs | `cartulary.harness.command.*.v1` identities preserved. |
| Output schemas | Summary schemas, event schemas, output classes, and machine-output behavior preserved. |
| Artifact paths | `.cartulary/test-results`, run roots, target summaries, scheduler artifacts, cache artifacts, and generated artifact paths preserved. |
| Failure taxonomy | Failure classes, reasons, usage exits, and child failure normalization preserved. |
| Cleanup predicates | Destructive confirmation, dry-run predicates, cleanup guards, and reset predicates preserved. |
| Cache schema IDs | Readiness/build cache schema IDs and agent-finalize action-cache schema ID preserved. |
| Generated artifacts | `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_render_index.json`, scheduler manifests, and design-token output updated only through generators. |
| Harness NLSpec wrapper | `scripts/harness-contract.sh` remains frozen by NLSpec. |

## Per-File Classification Matrix

The matrix classifies the 145 files present at the pre-slice rebaseline. Row
dispositions expand through the normalized remediation packages below; those
package fields are normative for every row using that disposition key.

| Disposition | Remediation | Areas | Rationale | Long-term benefit | Compatibility / migration impact | Risk if unresolved | Validation | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `delete_after_caller_move` | Completed: move callers/specs/tests/docs/generated owner inputs to owner path, run generators, then delete the root facade. | specification, implementation, tests, documentation, generated owner inputs | Root file only imported or execed durable owner code. | Fewer root shims; cache and topology name owner files directly. | Public Make behavior preserved; direct root-path users must migrate. | Hidden stale compatibility surface and false ownership in manifests. | Exact `rg` scan, direct owner tests, generator/drift checks. | Owner CLI/script exists and accepts old arguments. | File deleted and non-historical callers clean. |
| `keep_public_wrapper` | Keep until the owning public spec is intentionally revised. | specification, implementation, tests | Wrapper is a declared public/operator contract. | Public contract remains stable while internals move. | No migration for users; future deletion requires spec update. | Premature deletion breaks declared harness behavior. | `make harness-contract`, `make check-harness-smoke`. | Harness NLSpec. | Spec either keeps wrapper or explicitly retires it. |
| `keep_external_operator_entrypoint` | Keep as an external-provider/operator entrypoint. | implementation, documentation, release/operator | External consumers are source-limited and cannot be discovered in-repo. | Avoids breaking CI provider wiring outside this repo. | No in-repo caller migration. | External CI could break without visibility. | Shell syntax and release/operator smoke when touched. | External CI/provider usage. | External contract is replaced or explicitly retired. |
| `keep_operator_or_release_script` | Keep pending release/operator extraction. | implementation, release/operator, tests | Script performs release or standup operator work, not a thin compatibility shim. | Preserves release workflows while owner extraction is planned. | No public Make behavior change. | Release flow drift remains rooted in `scripts/`. | Release/standup direct tests when touched. | Release evidence and standup workflow owners. | Release owner moves implementation and callers. |
| `keep_pending_owner_extraction` | Retain for now; future owner extraction must update specs/manifests/tests/docs first when public. | implementation, tests, documentation, generated owner inputs as applicable | Root file is a non-thin implementation/orchestration script or a public/internal Make wrapper not admitted for deletion in this slice. | Prevents scope creep while documenting next extraction seams. | Public Make behavior unchanged. | `scripts/` remains a partial grab bag. | Affected direct test or public Make target. | Future owner extraction plan. | Owner path exists, callers move, generators run, and wrapper is deleted or contract-retained. |
| `keep_active_owner_test` | Keep the root test; assert owner paths or public Make behavior instead of root facade text. | tests | Root `scripts/test-*` files are active harness/task-surface characterization tests. | Keeps behavioral coverage while implementation moves to owners. | No public behavior change. | Tests can encode stale root path assumptions. | Direct test or `make check-harness-smoke`. | Owner implementation under test. | Test passes and remains referenced, or moves with its owner in a later slice. |

| Path | Disposition | Owner/dependency | Status / path-specific exit |
| --- | --- | --- | --- |
| `scripts/agent-finalize.mjs` | `delete_after_caller_move` | `tools/harness/core/agent-finalize-cli.mjs` | Deleted; callers use owner CLI. |
| `scripts/agent-finalize.sh` | `delete_after_caller_move` | `tools/harness/core/agent-finalize-cli.mjs` | Deleted; generated Make recipe invokes `$(NODE_BIN)` owner CLI. |
| `scripts/bootstrap-go-tool.sh` | `delete_after_caller_move` | `tools/harness/readiness/bootstrap-go-tool.sh` | Deleted; toolchain callers moved. |
| `scripts/bootstrap-node-runtime.sh` | `delete_after_caller_move` | `tools/harness/readiness/bootstrap-node-runtime.sh` | Deleted; bootstrap tests updated. |
| `scripts/bootstrap-shellcheck.sh` | `delete_after_caller_move` | `tools/harness/readiness/bootstrap-shellcheck.sh` | Deleted; shellcheck bootstrap tests updated. |
| `scripts/build-go-artifact.sh` | `delete_after_caller_move` | `tools/harness/backend/build-go-artifact.sh` | Deleted; build/cache inputs moved. |
| `scripts/build-web-artifact.sh` | `delete_after_caller_move` | `tools/harness/frontend/build-web-artifact.sh` | Deleted; build/cache inputs moved. |
| `scripts/cache-artifact.sh` | `delete_after_caller_move` | `tools/harness/readiness/cache-artifact.sh` | Deleted; cache schema unchanged. |
| `scripts/cartulary-runner.mjs` | `delete_after_caller_move` | `tools/harness/core/cartulary-runner-cli.mjs` | Deleted; Make runner defaults moved. |
| `scripts/check-doctor.sh` | `delete_after_caller_move` | `tools/harness/readiness/check-doctor.sh` | Deleted; `make doctor` calls owner script. |
| `scripts/check-font-bundle.mjs` | `delete_after_caller_move` | `tools/harness/frontend/font-bundle-check-cli.mjs` | Deleted; frontend font tests moved. |
| `scripts/check-go-target-plan-coverage.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-target-plan-coverage-cli.mjs` | Deleted; Go target tests moved. |
| `scripts/check-migration-history.mjs` | `delete_after_caller_move` | `tools/harness/backend/migration-history-cli.mjs` | Deleted; migration wrapper moved. |
| `scripts/check-phase-map.mjs` | `delete_after_caller_move` | `tools/harness/planning/phase-map-check-cli.mjs` | Deleted; phase-map tests moved. |
| `scripts/check-postgres-fixture-budget.mjs` | `delete_after_caller_move` | `tools/harness/backend/postgres-fixture-budget-cli.mjs` | Deleted; scheduler default moved. |
| `scripts/check-schema-object-ownership.mjs` | `delete_after_caller_move` | `tools/harness/backend/schema-object-ownership-cli.mjs` | Deleted; JSON-shape callers moved. |
| `scripts/dev-services.sh` | `delete_after_caller_move` | `tools/harness/readiness/dev-services.sh` | Deleted; local-dev Make recipes moved. |
| `scripts/dev-stack.sh` | `delete_after_caller_move` | `tools/harness/readiness/dev-stack.sh` | Deleted; local-dev Make recipes moved. |
| `scripts/diagnose-inotify.mjs` | `delete_after_caller_move` | `tools/harness/readiness/diagnose-inotify.mjs` | Deleted; doctor/dev-stack owner calls moved. |
| `scripts/frontend-install.sh` | `delete_after_caller_move` | `tools/harness/frontend/frontend-install.sh` | Deleted; install stamp recipe moved. |
| `scripts/frontend-toolchain.sh` | `delete_after_caller_move` | `tools/harness/frontend/frontend-toolchain.sh` | Deleted; toolchain stamp recipe moved. |
| `scripts/generate-design-tokens.mjs` | `delete_after_caller_move` | `tools/harness/frontend/design-token-cli.mjs` | Deleted; generated marker regenerated. |
| `scripts/harness-contract.mjs` | `delete_after_caller_move` | `tools/harness/core/harness-contract-cli.mjs` | Deleted; shell wrapper remains public. |
| `scripts/list-build-inputs.sh` | `delete_after_caller_move` | `tools/harness/readiness/list-build-inputs.sh` | Deleted; build input discovery moved. |
| `scripts/playwright-install.sh` | `delete_after_caller_move` | `tools/harness/browser/playwright-install.sh` | Deleted; install stamp recipe moved. |
| `scripts/print-go-shard-plan.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-shard-plan-cli.mjs` | Deleted; diagnostic tests moved. |
| `scripts/reset-web-e2e-stack.sh` | `delete_after_caller_move` | `tools/harness/browser/reset-web-e2e-stack.sh` | Deleted; browser lifecycle callers moved. |
| `scripts/run-browser-e2e-owned-stack.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-owned-stack.sh` | Deleted; stateful/browser callers moved. |
| `scripts/run-check-schedule.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/check-schedule-cli.mjs` | Deleted; check scheduler recipe moved. |
| `scripts/run-go-target.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-target-runner.mjs` | Deleted; runner gained direct CLI entry. |
| `scripts/run-service-backed-schedule.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/service-backed-schedule-cli.mjs` | Deleted; service-backed recipes moved. |
| `scripts/write-frontend-accessibility-summary.mjs` | `delete_after_caller_move` | `tools/harness/frontend/accessibility-summary-cli.mjs` | Deleted; browser a11y callers moved. |
| `scripts/harness-contract.sh` | `keep_public_wrapper` | `docs/testing-harness-nlspec.md`, `tools/harness/core/harness-contract-cli.mjs` | Retained by spec. |
| `scripts/ci/verify.sh` | `keep_external_operator_entrypoint` | external CI provider | Retained for source-limited external use. |
| `scripts/ci/check-deployable-shape.sh` | `delete_after_caller_move` | `tools/release-evidence/check-deployable-shape.sh` | Moved and deleted in Iteration 4 WS-06; release/deployable-shape target invokes owner path. |
| `scripts/ci/check-standup-package-smoke.sh` | `delete_after_caller_move` | `tools/release-evidence/check-standup-package-smoke.sh` | Moved and deleted in Iteration 4 WS-06; standup package smoke target invokes owner path. |
| `scripts/ci/check-standup-operational-recovery-smoke.sh` | `delete_after_caller_move` | `tools/release-evidence/check-standup-operational-recovery-smoke.sh` | Moved and deleted in Iteration 4 WS-06; standup recovery smoke target invokes owner path. |
| `scripts/check-backend-module-boundaries.mjs` | `delete_after_caller_move` | `tools/harness/static-analysis/backend-module-boundary-check-cli.mjs` | Moved and deleted in Iteration 4 WS-05; public backend boundary target invokes owner path. |
| `scripts/check-frontend-import-boundaries.mjs` | `delete_after_caller_move` | `tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs` | Moved and deleted in Iteration 4 WS-05; public frontend import-boundary target invokes owner path. |
| `scripts/check-go-test-duration-baseline-coverage.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-test-duration-baseline-coverage-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/check-go-test-duration-baseline-drift.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-test-duration-baseline-drift-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/check-migrations.sh` | `delete_after_caller_move` | `tools/harness/backend/check-migrations.sh` | Moved and deleted in Iteration 4 WS-06; migration drift targets invoke owner path. |
| `scripts/check-phase-maps.sh` | `delete_after_caller_move` | `tools/harness/planning/check-phase-maps.sh` | Moved and deleted in Iteration 4 WS-03; generated phase-map recipe invokes owner path. |
| `scripts/check-phase-test-names.mjs` | `delete_after_caller_move` | `tools/harness/planning/phase-test-name-check-cli.mjs` | Moved and deleted in Iteration 4 WS-03; generated phase-test-name recipe invokes owner path. |
| `scripts/check-scheduler-event-order-drift.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/scheduler-event-order-drift-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/check-scheduler-summary-timing-drift.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/scheduler-summary-timing-drift-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/check-service-backed-unit-tests.sh` | `delete_after_caller_move` | `tools/gotestservicecheck` | Deleted in Iteration 4 WS-02; generated recipes invoke `$(GO) run ./tools/gotestservicecheck`. |
| `scripts/check-toolchain-pins.mjs` | `delete_after_caller_move` | `tools/harness/readiness/toolchain-pin-check-cli.mjs` | Moved and deleted in Iteration 4 WS-06; toolchain drift target invokes owner path. |
| `scripts/duration-baseline-drift-suite.sh` | `delete_after_caller_move` | `tools/harness/scheduler/duration-baseline-drift-suite.sh` | Moved and deleted in Iteration 4 WS-03; generated drift-suite recipe invokes owner path. |
| `scripts/embed-web-assets.sh` | `delete_after_caller_move` | `tools/harness/frontend/embed-web-assets.sh` | Moved and deleted in Iteration 4 WS-06; embedded web asset cache recipe invokes owner path. |
| `scripts/frontend-evidence-audit.mjs` | `delete_after_caller_move` | `tools/harness/frontend/frontend-evidence-audit-cli.mjs` | Moved and deleted in Iteration 4 WS-06; frontend evidence audit target invokes owner path. |
| `scripts/generate-artifacts.sh` | `delete_after_caller_move` | `tools/harness/generated-artifacts/generate-artifacts.sh` | Moved and deleted in Iteration 4 WS-06; generated artifact recipe invokes owner path. |
| `scripts/harness-smoke-durations.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/harness-smoke-durations-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/print-explain-phase.mjs` | `delete_after_caller_move` | `tools/harness/planning/explain-phase-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make explain-phase` is unchanged. |
| `scripts/print-explain-run.mjs` | `delete_after_caller_move` | `tools/harness/core/explain-run-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make explain-run` is unchanged. |
| `scripts/print-explain-target.mjs` | `delete_after_caller_move` | `tools/harness/planning/explain-target-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make explain-target` is unchanged. |
| `scripts/print-fixture-report.mjs` | `delete_after_caller_move` | `tools/harness/core/fixture-report-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make fixture-report` is unchanged. |
| `scripts/print-frontend-toolchain.sh` | `delete_after_caller_move` | frontend readiness stamp output | Deleted in Iteration 4 WS-02; `tools/harness/frontend/frontend-toolchain.sh --print-stamp` owns stamp output. |
| `scripts/print-target-plan.mjs` | `delete_after_caller_move` | `tools/harness/planning/target-plan-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make target-plan` and `make target-plan-json` are unchanged. |
| `scripts/print-task-guide.mjs` | `delete_after_caller_move` | `tools/harness/planning/task-guide-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make task-guide` is unchanged. |
| `scripts/print-task-surface-report.mjs` | `delete_after_caller_move` | `tools/harness/planning/task-surface-report-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make task-surface-report` is unchanged. |
| `scripts/run-browser-e2e-a11y-preflight.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-a11y-preflight.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-a11y.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-a11y.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-batch.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-batch.sh` | Moved and deleted in Iteration 4 WS-04; browser batch ownership is under `tools/harness/browser`. |
| `scripts/run-browser-e2e-functional.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-functional.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-manifest-dependency.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-manifest-dependency.sh` | Moved and deleted in Iteration 4 WS-04; visual/stateful/measurement owners invoke owner path. |
| `scripts/run-browser-e2e-measurement.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-measurement.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-resettable.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-resettable.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-stateful.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-stateful.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-target.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-target.sh` | Moved and deleted in Iteration 4 WS-04; generated browser recipes invoke owner path. |
| `scripts/run-browser-e2e-visual-update.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-visual-update.sh` | Moved and deleted in Iteration 4 WS-04; generated visual-update recipe invokes owner path. |
| `scripts/run-browser-e2e-visual.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-visual.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-browser-e2e-webserver-backed.sh` | `delete_after_caller_move` | `tools/harness/browser/run-browser-e2e-webserver-backed.sh` | Moved and deleted in Iteration 4 WS-04; browser Make targets invoke owner path. |
| `scripts/run-fallow-static.mjs` | `delete_after_caller_move` | `tools/harness/static-analysis/fallow-static-cli.mjs` | Moved and deleted in Iteration 4 WS-05; Make node-tool dispatch invokes owner path. |
| `scripts/run-frontend-biome.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/frontend-biome.sh` | Moved and deleted in Iteration 4 WS-05; frontend package scripts and Make lint targets invoke owner path. |
| `scripts/run-frontend-unit.sh` | `delete_after_caller_move` | `tools/harness/frontend/run-frontend-unit.sh` | Moved and deleted in Iteration 4 WS-05; public `make frontend-unit` invokes owner path. |
| `scripts/run-go-format.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-format.sh` | Moved and deleted in Iteration 4 WS-05; Go format recipes invoke owner path. |
| `scripts/run-go-gosec-audit.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-gosec-audit.sh` | Moved and deleted in Iteration 4 WS-05; Go security audit recipe invokes owner path. |
| `scripts/run-go-gosec-targeted.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-gosec-targeted.sh` | Moved and deleted in Iteration 4 WS-05; targeted Go security recipe invokes owner path. |
| `scripts/run-go-govulncheck.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-govulncheck.sh` | Moved and deleted in Iteration 4 WS-05; vulnerability scan recipe invokes owner path. |
| `scripts/run-go-staticcheck.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-staticcheck.sh` | Moved and deleted in Iteration 4 WS-05; staticcheck recipe invokes owner path. |
| `scripts/run-go-vet.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/go-vet.sh` | Moved and deleted in Iteration 4 WS-05; Go vet recipe invokes owner path. |
| `scripts/run-harness-smoke.mjs` | `delete_after_caller_move` | `tools/harness/core/run-harness-smoke-cli.mjs` | Moved and deleted in Iteration 4 WS-06; public harness-smoke targets invoke owner path. |
| `scripts/run-make-node-tool.mjs` | `delete_after_caller_move` | Make node-tool dispatch | Moved to `tools/harness/core/run-make-node-tool-cli.mjs` in Iteration 4 WS-02. |
| `scripts/run-make-node-tool.sh` | `delete_after_caller_move` | Make node-tool shell wrapper | Moved to `tools/harness/core/run-make-node-tool.sh` in Iteration 4 WS-02; generated macro invokes owner path. |
| `scripts/run-make-sequence.sh` | `delete_after_caller_move` | `tools/harness/core/run-make-sequence.sh` | Moved and deleted in Iteration 4 WS-06; Make sequence variables invoke owner path. |
| `scripts/run-markdownlint.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/markdownlint.sh` | Moved and deleted in Iteration 4 WS-05; markdown lint recipe invokes owner path. |
| `scripts/run-phase-slice.mjs` | `delete_after_caller_move` | `tools/harness/planning/phase-slice-cli.mjs` | Moved and deleted in Iteration 4 WS-03; public `make phase-slice` and `make service-backed-slice` are unchanged. |
| `scripts/run-scripts-biome.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/scripts-biome.sh` | Moved and deleted in Iteration 4 WS-05; scripts Biome recipe invokes owner path. |
| `scripts/run-shellcheck.sh` | `delete_after_caller_move` | `tools/harness/static-analysis/shellcheck.sh` | Moved and deleted in Iteration 4 WS-05; shell lint recipe invokes owner path. |
| `scripts/service-backed-make-target-durations.mjs` | `delete_after_caller_move` | `tools/harness/scheduler/service-backed-make-target-durations-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/start-web-e2e.sh` | `delete_after_caller_move` | `tools/harness/browser/start-web-e2e.sh` | Moved and deleted in Iteration 4 WS-04; scheduler session defaults, Playwright config, and browser Make targets use owner path. |
| `scripts/update-go-test-durations.mjs` | `delete_after_caller_move` | `tools/harness/backend/go-test-duration-baselines-cli.mjs` | Moved and deleted in Iteration 4 WS-03; Make node-tool dispatch invokes owner path. |
| `scripts/test-agent-finalize.sh` | `keep_active_owner_test` | core finalizer owner CLI | Updated to call owner CLI with `node`. |
| `scripts/test-benchmark-claim-check.sh` | `keep_active_owner_test` | benchmark/release evidence | Retained. |
| `scripts/test-bootstrap-node-runtime.sh` | `keep_active_owner_test` | readiness bootstrap owner | Updated to owner path. |
| `scripts/test-bootstrap-shellcheck.sh` | `keep_active_owner_test` | readiness shellcheck owner | Updated to owner path. |
| `scripts/test-browser-shard-plan.sh` | `keep_active_owner_test` | browser shard planner | Retained. |
| `scripts/test-build-input-discovery.sh` | `keep_active_owner_test` | readiness build input owner | Updated to owner path. |
| `scripts/test-cache-artifact.sh` | `keep_active_owner_test` | readiness cache owner | Updated to owner path. |
| `scripts/test-cartulary-runner-service-backed-target.sh` | `keep_active_owner_test` | core runner owner CLI | Updated to owner path. |
| `scripts/test-check-migrations.sh` | `keep_active_owner_test` | backend migration owner CLI | Updated to backend owner path and passed in WS-06. |
| `scripts/test-check-phase-test-names.sh` | `keep_active_owner_test` | planning test-name check | Retained. |
| `scripts/test-check-scheduler.sh` | `keep_active_owner_test` | scheduler check owner CLI | Updated to owner path. |
| `scripts/test-check-toolchain-pins.sh` | `keep_active_owner_test` | toolchain/core owner paths | Updated to readiness owner path and passed in WS-06. |
| `scripts/test-dev-services-lifecycle.sh` | `keep_active_owner_test` | readiness local-dev owner | Updated to owner path. |
| `scripts/test-dev-stack-lifecycle.sh` | `keep_active_owner_test` | readiness dev-stack owner | Updated to owner path. |
| `scripts/test-execution-topology.sh` | `keep_active_owner_test` | generated topology/task surface | Updated agent-finalize backing script assertion. |
| `scripts/test-fallow-static.sh` | `keep_active_owner_test` | frontend static analysis | Updated to invoke static-analysis owner path. |
| `scripts/test-frontend-evidence-audit.sh` | `keep_active_owner_test` | frontend evidence audit | Updated to frontend evidence owner path and passed in WS-06. |
| `scripts/test-frontend-import-boundaries.sh` | `keep_active_owner_test` | frontend import boundary | Updated to invoke static-analysis owner path. |
| `scripts/test-generate-drift.sh` | `keep_active_owner_test` | generated artifact drift | Updated scratch owner paths and fake Node runtime fixture for moved generator owner path. |
| `scripts/test-generated-artifact-policy.sh` | `keep_active_owner_test` | generated artifact policy | Updated lint-scope fixtures for static-analysis owner paths. |
| `scripts/test-go-test-duration-baselines.sh` | `keep_active_owner_test` | backend duration baselines | Updated to invoke owner duration CLIs. |
| `scripts/test-harness-contracts.mjs` | `keep_active_owner_test` | harness contract | Retained. |
| `scripts/test-harness-smoke-duration-baselines.sh` | `keep_active_owner_test` | harness smoke duration baselines | Updated to invoke owner scheduler duration CLI. |
| `scripts/test-json-shapes.sh` | `keep_active_owner_test` | generated JSON shape checks | Updated accessibility owner path. |
| `scripts/test-lint-shell.sh` | `keep_active_owner_test` | shell lint wrapper | Updated to invoke static-analysis owner path and cover untracked shell-file inventory. |
| `scripts/test-make-node-tools.sh` | `keep_active_owner_test` | Make node-tool dispatch | Retained. |
| `scripts/test-print-target-plan.sh` | `keep_active_owner_test` | planning/backend owner CLIs | Updated owner paths; heavy-shard invariant fixed in Iteration 4 WS-01 and test now passes. |
| `scripts/test-public-make-wrapper-smoke.sh` | `keep_active_owner_test` | public Make wrappers | Retained. |
| `scripts/test-release-task-surface.sh` | `keep_active_owner_test` | release task surface | Updated through generated task surface to release-evidence owner paths and passed in WS-06. |
| `scripts/test-run-frontend-unit.sh` | `keep_active_owner_test` | frontend unit orchestration | Updated to invoke frontend owner path; fake Vitest now honors `-t` phase-slice filters and direct self-test passes. |
| `scripts/test-run-go-gosec-audit.sh` | `keep_active_owner_test` | Go security wrapper | Updated to invoke static-analysis owner path. |
| `scripts/test-run-go-gosec-targeted.sh` | `keep_active_owner_test` | Go security wrapper | Updated to invoke static-analysis owner path. |
| `scripts/test-run-go-govulncheck.sh` | `keep_active_owner_test` | Go security wrapper | Updated to invoke static-analysis owner path. |
| `scripts/test-run-go-staticcheck.sh` | `keep_active_owner_test` | Go lint wrapper | Updated to invoke static-analysis owner path. |
| `scripts/test-run-go-target.sh` | `keep_active_owner_test` | backend Go target owner runner | Updated to invoke owner runner through `node`. |
| `scripts/test-run-make-sequence-fast.sh` | `keep_active_owner_test` | Make sequence | Updated to core Make-sequence and harness-smoke owner paths and passed in WS-06. |
| `scripts/test-run-make-sequence.sh` | `keep_active_owner_test` | Make sequence/topology | Updated expected core owner paths and passed in WS-06. |
| `scripts/test-run-phase-slice.sh` | `keep_active_owner_test` | phase slice scheduler | Updated to assert owner phase-slice CLI path. |
| `scripts/test-run-phase.sh` | `keep_active_owner_test` | core run-phase output | Updated owner-path lint fixture; missing-child aggregate classification fixture now passes. |
| `scripts/test-run-playwright-manifest-phase.sh` | `keep_active_owner_test` | browser manifest phase | Retained. |
| `scripts/test-run-playwright-phase.sh` | `keep_active_owner_test` | browser phase runner | Retained. |
| `scripts/test-run-playwright-webserver-batch.sh` | `keep_active_owner_test` | browser webserver batch | Updated to assert owner browser batch wrapper. |
| `scripts/test-run-vitest-manifest-phase.sh` | `keep_active_owner_test` | frontend manifest phase | Retained. |
| `scripts/test-run-vitest-phase.sh` | `keep_active_owner_test` | frontend phase runner | Retained. |
| `scripts/test-sbom-license-evidence.mjs` | `keep_active_owner_test` | release evidence | Retained. |
| `scripts/test-seaweedfs-release-evidence.mjs` | `keep_active_owner_test` | release evidence | Retained. |
| `scripts/test-service-backed-make-target-duration-baselines.sh` | `keep_active_owner_test` | scheduler duration baselines | Updated to invoke owner scheduler duration CLI. |
| `scripts/test-service-backed-scheduler.sh` | `keep_active_owner_test` | scheduler service-backed owner CLI | Updated to owner path. |
| `scripts/test-task-guidance.mjs` | `keep_active_owner_test` | planning guidance | Retained. |
| `scripts/test-task-surface-report.sh` | `keep_active_owner_test` | task-surface report | Retained. |
| `scripts/test-tool-output-real-targets.sh` | `keep_active_owner_test` | core output | Retained. |
| `scripts/test-web-e2e-lifecycle.sh` | `keep_active_owner_test` | browser lifecycle | Updated to source/assert browser lifecycle owner path. |

## Candidate Deletion And Caller-Update Opportunities

These are not part of the completed deletion set. They are the next known
places where root `scripts/` may still be acting as an implementation home or
public wrapper.

| Candidate | Remediation | Areas | Rationale | Benefit | Compatibility / migration impact | Risk if unresolved | Validation | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `scripts/check-service-backed-unit-tests.sh` | Completed in Iteration 4 WS-02: move `service-backed-unit-check` topology command to `$(GO) run ./tools/gotestservicecheck`, regenerate task surface, then delete root wrapper. | implementation, tests, generated owner inputs | Thin wrapper over `tools/gotestservicecheck`. | Removes another root internal helper. | Generated Make recipe changes implementation path only. | Root wrapper no longer exists. | `bash scripts/test-execution-topology.sh`, `make phase-schedules`, `make check-harness-smoke`. | Go service-test tool owner. | Complete. |
| `scripts/run-make-node-tool.sh` | Completed in Iteration 4 WS-02: move shell wrapper and CLI to `tools/harness/core`, regenerate task surface, then delete root paths. | implementation, generated owner inputs | Generated macro no longer names root `scripts/`. | Centralizes Make node-tool dispatch. | Public Make behavior unchanged; raw root callers migrate to owner path. | Root shell wrapper no longer exists. | `scripts/test-make-node-tools.sh`, `scripts/test-run-make-sequence.sh`, task-surface drift. | Task-surface generator and core runner owner. | Complete. |
| `scripts/print-frontend-toolchain.sh` | Completed in Iteration 4 WS-02: frontend readiness owner prints the stamp through `tools/harness/frontend/frontend-toolchain.sh --print-stamp`. | implementation, generated owner inputs | Small Make diagnostic wrapper moved into owner script. | Fewer root readiness helpers. | `frontend-toolchain` output remains byte-compatible. | Extra root helper no longer exists. | `make frontend-toolchain`, `make check-harness-smoke`. | Frontend/readiness owner. | Complete. |
| Browser wrapper cluster: `run-browser-e2e-*.sh`, `start-web-e2e.sh` | Completed in Iteration 4 WS-04: move browser wrappers to `tools/harness/browser`, update callers/tests/generated inputs, then delete root paths. | implementation, tests, documentation, generated owner inputs | Browser lifecycle/orchestration now lives with browser harness owners. | Browser lifecycle code is owner-local and easier to extend. | Public browser Make targets preserved; raw root callers migrate to owner paths or Make targets. | Root browser implementation no longer exists. | Browser lifecycle tests, browser Make targets, exact deleted-path scan. | WS-04 browser owner extraction. | Complete. |
| Lint/security wrapper cluster: `run-go-*`, `run-frontend-biome.sh`, `run-scripts-biome.sh`, `run-shellcheck.sh`, `run-markdownlint.sh` | Completed in Iteration 4 WS-05: move wrappers to `tools/harness/static-analysis` and update callers/tests/generated inputs. | implementation, tests, documentation, generated owner inputs | Static-analysis policy now has one owner family instead of root shell names. | Makes lint/security ownership explicit and easier to extend. | Public lint and security target behavior preserved; raw root-path users migrate to owner paths or Make targets. | Stale root lint/security code would keep generated-path exclusions and scan policy scattered. | `make lint-shell`, `make lint-scripts`, `make lint-biome`, `make frontend-unit`, `make frontend-import-boundary-check`, `make backend-module-boundary-check`, `make go-vulncheck`, `make go-gosec-targeted`, `make go-gosec-audit`, affected direct self-tests. | WS-05 caller migration and generated input updates. | Owner paths referenced directly; root wrappers deleted. |
| Historical recovery docs | Leave unchanged unless republished. | documentation | Recovery docs intentionally preserve observed historical source paths. | Avoids rewriting evidence history. | No active-doc migration. | Active readers could confuse history with current implementation if not scoped. | Active-doc scans exclude `docs/testing-harness-spec-recovery-docs/**`. | Documentation owner. | If republished as active docs, update paths first. |

## Workstream Matrix

| Workstream | Scope | Status | Validation used |
| --- | --- | --- | --- |
| WS-00 Tracker Rebaseline | Re-inventory and rewrite tracker. | Complete in this file. | Final `make lint-markdown` required after this edit. |
| WS-01 Readiness/Build/Cache Facades | Move Make recipes, cache inputs, tests, toolchain pin checks, generate-drift scratch inputs; delete facades. | Complete. | Syntax, direct readiness/cache tests, `make toolchain-drift`, `make lint-shell`. |
| WS-02 Backend/Planning CLI Facades | Move tests/scheduler helpers to owner CLIs; delete backend/planning trampolines. | Complete. | `node --check`, `bash scripts/test-run-go-target.sh`, `bash scripts/test-check-migrations.sh`; `test-print-target-plan` recorded known unrelated failure. |
| WS-03 Core/Scheduler Trampolines | Move Make variables, runner defaults, action-cache inputs, task-surface/topology metadata, tests; delete trampolines except `harness-contract.sh`. | Complete. | Finalizer, runner, scheduler smoke, `make harness-contract`, `make check-harness-smoke`. |
| WS-04 Frontend/Browser/Readiness Facades | Move font/design-token/accessibility/browser reset/owned-stack/inotify/docs; delete facades. | Complete. | `make generate`, `make frontend-unit`, `bash scripts/test-json-shapes.sh`, `bash scripts/test-web-e2e-lifecycle.sh`. |
| WS-05 Local-Dev Root Facades | Move doctor/dev-services/dev-stack callers to readiness owner scripts; delete facades. | Complete. | `bash scripts/test-dev-services-lifecycle.sh`, `bash scripts/test-dev-stack-lifecycle.sh`, `make doctor`, `make lint-shell`. |
| WS-Z Final Regeneration And Handoff | Run generators/drift checks; update tracker validation and handoff rows. | Complete. | `make phase-schedules`, `make generate`, drift/shape/policy checks, final exact caller scan, final markdown lint. |

## Slice Admission Rules

- Start each future slice by rerunning the facade/caller scan for that slice.
- Do not delete a facade until exact non-archive `rg` finds no live caller.
- Do not add new root shims.
- Do not hand-edit generated outputs.
- Update this tracker after each completed slice and before starting the next one.
- Any public behavior change requires owner-spec revision first; otherwise preserve behavior exactly.

## Validation Plan And Results

| Command | Result | Run root | Notes |
| --- | --- | --- | --- |
| Final `make phase-schedules` | PASS | `.cartulary/test-results/20260703T142708Z-p757561` | Regenerated phase/task-surface artifacts after all moves. |
| Final `make generate` | PASS | `.cartulary/test-results/20260703T142708Z-p757608` | Generated contract/design artifacts current. |
| Final `make generate-drift` | PASS | `.cartulary/test-results/20260703T142755Z-p760220` | Scratch drift validation passed after generator owner move. |
| Final `make phase-schedule-drift` | PASS | `.cartulary/test-results/20260703T142727Z-p758592` | Regenerated phase schedule outputs are current. |
| Final `make json-shape-check` | PASS | `.cartulary/test-results/20260703T142727Z-p758600` | JSON shape validation passed. |
| Final `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260703T142727Z-p758610` | Generated artifact policy validation passed. |
| Final `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | PASS | none emitted | Task-surface report check passed. |
| Final `make harness-contract` | PASS | none emitted | Retained public wrapper behavior preserved. |
| Final `make check-harness-smoke` | PASS | none emitted | Harness smoke sweep passed. |
| Final `make lint-shell` | PASS | none emitted | Shell lint passed with untracked owner-script inventory. |
| Final exact deleted-path `rg` scan | PASS | none | No non-archive, non-recovery live references to retired root implementation paths. |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260703T043250Z-p313996` | FAIL | `.cartulary/test-results/20260703T142815Z-p762688` | Retained run was valid/latest, but `scheduler-summary-timing-drift` failed with `duration_baseline_drift`: `backend-integration:backend-integration-app-shard-05` was `26121ms`, above `1.25x` peer median `2088ms`. No files were mutated before failure. |
| `make phase-schedules` | PASS | `.cartulary/test-results/20260703T132103Z-p585384` | Regenerated task-surface/topology outputs from owner input. |
| `make generate` | PASS | `.cartulary/test-results/20260703T132123Z-p585707` | Regenerated design-token marker from owner generator. |
| `git diff --name-only -- '*.sh' \| xargs -r bash -n` | FAIL | none | Initial validation command included deleted shell files; rerun with `--diff-filter=ACMR`. |
| `git diff --name-only --diff-filter=ACMR -- '*.sh' \| xargs -r bash -n` | PASS | none | Existing touched shell scripts parse. |
| `git diff --name-only -- '*.mjs' \| xargs -r -n1 node --check` | FAIL | none | Initial validation command included deleted Node files; rerun with `--diff-filter=ACMR`. |
| `git diff --name-only --diff-filter=ACMR -- '*.mjs' \| xargs -r -n1 node --check` | PASS | none | Existing touched Node files parse. |
| `bash scripts/test-build-input-discovery.sh` | PASS | none | Readiness/build input owner path. |
| `bash scripts/test-bootstrap-node-runtime.sh` | PASS | none | Printed expected fake transient download diagnostics. |
| `bash scripts/test-bootstrap-shellcheck.sh` | PASS | none | ShellCheck bootstrap owner path. |
| `bash scripts/test-cache-artifact.sh` | PASS | none | Cache schema behavior preserved. |
| `bash scripts/test-run-go-target.sh` | FAIL then PASS | none | First failed on direct execution of non-executable owner `.mjs`; fixed test to invoke through `node`. |
| `bash scripts/test-check-migrations.sh` | PASS | none | Migration owner CLI paths. |
| `bash scripts/test-print-target-plan.sh` | FAIL | none | Known unrelated heavy-shard split invariant: `backend-integration go shard plan must be weighted, policy-bearing, split heavy aggregates, and keep authoritative/support shards separate`. |
| `bash scripts/test-agent-finalize.sh` | PASS | none | Owner finalizer CLI and action-cache inputs. |
| `bash scripts/test-cartulary-runner-service-backed-target.sh` | PASS | none | Owner runner CLI. |
| `bash scripts/test-check-scheduler.sh smoke` | PASS | none | Owner check scheduler CLI. |
| `bash scripts/test-service-backed-scheduler.sh smoke` | FAIL then PASS | none | First failed on deleted Postgres fixture-budget default; fixed scheduler default to owner CLI. |
| `node --check tools/harness/scheduler/service-backed-schedule-cli.mjs` | PASS | none | Syntax after scheduler default fix. |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260703T132630Z-p595668` | Shape validation. |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260703T132630Z-p595669` | Generated artifact policy validation. |
| `make phase-schedule-drift` | PASS | `.cartulary/test-results/20260703T132630Z-p595667` | Regenerated outputs are current. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | PASS | none emitted | Report check passed and printed task-surface inventory. |
| `make toolchain-drift` | PASS | `.cartulary/test-results/20260703T132643Z-p596697` | Toolchain pin drift. |
| `make lint-shell` | PASS | none emitted | Shell lint target passed. |
| `make lint-scripts` | PASS | none emitted | Script lint target passed. |
| `bash scripts/test-json-shapes.sh` | PASS | none | Direct JSON-shape harness tests. |
| `bash scripts/test-web-e2e-lifecycle.sh` | PASS | none | Printed expected signal/port-conflict diagnostics while exiting zero. |
| `bash scripts/test-dev-services-lifecycle.sh` | PASS | none | Readiness local-dev owner script. |
| `bash scripts/test-dev-stack-lifecycle.sh` | PASS | none | Readiness dev-stack owner script. |
| `make harness-contract` | PASS | none emitted | Public harness behavior. |
| `make check-harness-smoke` | PASS | none emitted | Harness smoke target. |
| `make frontend-unit` | PASS | `.cartulary/test-results/20260703T132720Z-p601015` | Frontend/font/design-token affected surface. |
| `make doctor` | PASS | `.cartulary/test-results/20260703T132720Z-p601062` | Local readiness public target. |
| Exact deleted-facade `rg -F` scan | PASS | none | No non-archive, non-recovery live references to deleted facades. |
| Tracker matrix coverage sanity check | PASS | none | `baseline=145`, `classified=145`, no missing or extra paths. |
| `make agent-finalize RESULTS_DIR=<successful full warm check run>` | FAIL | `.cartulary/test-results/20260703T142815Z-p762688` | Retained-run maintenance reached `scheduler-summary-timing-drift` and failed duration-baseline drift; no files were mutated before failure. |
| `make lint-markdown` | PASS | none emitted | Final tracker lint passed after this edit. |

## Handoff Log

Template for future slices:

`Time | Agent/session | Branch/commit | Inventory counts | Facades deleted | Files changed | Specs/manifests/generators touched | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice`

| Time | Agent/session | Branch/commit | Inventory counts | Facades deleted | Files changed | Specs/manifests/generators touched | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T13:21Z | Codex / Iteration 3 | `main` / `54e926e4` base | Pre: 145 total, 141 root, 4 CI, 52 tests, 0 `scripts/lib`; post: 113 total, 109 root, 4 CI, 52 tests, `scripts/lib` absent | 32 root facades across WS-01..WS-05 | Makefile, active docs/tests, owner harness files, generated task-surface/topology/design-token outputs, deleted facades | `tools/execution_topology_manifest.json`, `tools/generate_drift_scratch_inputs.json`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_render_index.json`, `packages/ui-contracts/src/generated/design-tokens.ts` | See validation table; generator roots: `.cartulary/test-results/20260703T132103Z-p585384`, `.cartulary/test-results/20260703T132123Z-p585707`; drift/shape roots listed above | `make agent-finalize RESULTS_DIR=<successful full warm check run>` skipped because no successful full warm run was available | `scripts/test-print-target-plan.sh` still fails known heavy-shard split invariant; remaining root orchestration wrappers need separate extraction plan | Future extraction candidates listed above; no deleted-facade callers remain. |
| 2026-07-03T14:29Z | Codex / Iteration 4 | working tree on `main` | Pre: 113 total, 109 root, 4 CI, 52 tests; post: 54 total, 53 root, 1 CI, 52 tests, `scripts/lib` absent | 59 root implementation/wrapper files in Iteration 4; only `scripts/harness-contract.sh` and `scripts/ci/verify.sh` remain as non-test wrappers | Makefile, active docs/tests, owner harness files, `tools/release-evidence`, task-surface/topology/generator manifests and outputs, deleted root implementations | `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_manifest.json`, `tools/execution_topology_render_index.json`, `tools/scheduler_manifest.json`, `tools/generated_artifact_policy.json`, `tools/generate_drift_scratch_inputs.json` | Final validation rows above; final generation roots `.cartulary/test-results/20260703T142708Z-p757561` and `.cartulary/test-results/20260703T142708Z-p757608`; final drift/shape roots listed above | `make frontend-evidence-audit` skipped as a bare target because required evidence inputs were absent; `make release-check` skipped as full release gate in favor of `make deployable-shape`; `make agent-finalize` failed retained-run timing drift as recorded above | No live caller references to retired root implementation paths; retained-run timing drift remains for baseline maintenance | Handoff complete; future work is timing-baseline maintenance, not scripts root implementation extraction. |

## Open Questions / Blockers

- External consumers of root `scripts/*` paths are unknown. Default retained external entrypoint is `scripts/ci/verify.sh`.
- `scripts/harness-contract.sh` is frozen by `docs/testing-harness-nlspec.md`; removing it is blocked on spec revision.
- Retained-run maintenance is blocked by scheduler summary timing drift in `.cartulary/test-results/20260703T043250Z-p313996`.
- Historical `docs/testing-harness-spec-recovery-docs/**` still cite removed root paths by design; do not update unless those docs are republished as active guidance.

## Binary Completion Criteria

Iteration 3 is complete when all of the following are true:

- Tracker baseline records both the pre-slice live counts and post-slice counts.
- All 145 pre-slice `scripts/` files are classified.
- Every listed facade is either deleted or explicitly retained with contract/operator rationale.
- Non-archive, non-recovery callers no longer reference deleted facades.
- Task-surface/topology/generated outputs reflect owner paths through generators.
- Active docs no longer describe removed root facades as implementation homes.
- Public Make target behavior, command IDs, schemas, artifact paths, failure taxonomy, cleanup predicates, and cache schema IDs are unchanged.
- Validation commands and run roots are recorded, including skipped checks and reasons.
- Final `make lint-markdown` passes after this tracker edit.
