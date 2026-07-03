# Scripts Facade Elimination Tracker, Iteration 3

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
| `scripts/` files | 113 total |
| Root files | 109 |
| Files under `scripts/ci/` | 4 |
| Root `scripts/test-*` files | 52 |
| `scripts/lib/` | absent in the live tree; 0 files |
| Live paths referenced by `Makefile`, `tools/task_surface_manifest.json`, or `tools/execution_topology_manifest.json` | 111 |
| Live paths not directly referenced by those three | `scripts/ci/verify.sh`, `scripts/harness-contract.sh` |
| Deleted root compatibility facades | 32 |
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
| `scripts/ci/check-deployable-shape.sh` | `keep_operator_or_release_script` | Release/operator deployable-shape check. | Keep until release workflow owner extracts it. |
| `scripts/ci/check-standup-package-smoke.sh` | `keep_operator_or_release_script` | Standup/package release smoke wrapper. | Keep until release workflow owner extracts it. |
| `scripts/ci/check-standup-operational-recovery-smoke.sh` | `keep_operator_or_release_script` | Standup/recovery release smoke wrapper. | Keep until release workflow owner extracts it. |

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
| `scripts/ci/check-deployable-shape.sh` | `keep_operator_or_release_script` | release/operator checks | Retained. |
| `scripts/ci/check-standup-package-smoke.sh` | `keep_operator_or_release_script` | release/operator checks | Retained. |
| `scripts/ci/check-standup-operational-recovery-smoke.sh` | `keep_operator_or_release_script` | release/operator checks | Retained. |
| `scripts/check-backend-module-boundaries.mjs` | `keep_pending_owner_extraction` | backend boundary tooling | Retain pending extraction. |
| `scripts/check-frontend-import-boundaries.mjs` | `keep_pending_owner_extraction` | frontend boundary tooling | Retain pending extraction. |
| `scripts/check-go-test-duration-baseline-coverage.mjs` | `keep_pending_owner_extraction` | backend duration baseline tooling | Retain pending extraction. |
| `scripts/check-go-test-duration-baseline-drift.mjs` | `keep_pending_owner_extraction` | backend duration baseline tooling | Retain pending extraction. |
| `scripts/check-migrations.sh` | `keep_pending_owner_extraction` | backend migration owner CLI | Retain wrapper around multi-check migration flow. |
| `scripts/check-phase-maps.sh` | `keep_pending_owner_extraction` | planning/frontend phase manifests | Retain current generated Make entrypoint. |
| `scripts/check-phase-test-names.mjs` | `keep_pending_owner_extraction` | planning/test-name checks | Retain pending extraction. |
| `scripts/check-scheduler-event-order-drift.mjs` | `keep_pending_owner_extraction` | scheduler drift tooling | Retain pending extraction. |
| `scripts/check-scheduler-summary-timing-drift.mjs` | `keep_pending_owner_extraction` | scheduler drift tooling | Retain pending extraction. |
| `scripts/check-service-backed-unit-tests.sh` | `keep_pending_owner_extraction` | `tools/gotestservicecheck` | Retain; candidate future deletion after topology command revision. |
| `scripts/check-toolchain-pins.mjs` | `keep_pending_owner_extraction` | toolchain pin policy | Retain pending extraction. |
| `scripts/duration-baseline-drift-suite.sh` | `keep_pending_owner_extraction` | duration drift suite | Retain pending extraction. |
| `scripts/embed-web-assets.sh` | `keep_pending_owner_extraction` | build/embed tooling | Retain pending extraction. |
| `scripts/frontend-evidence-audit.mjs` | `keep_pending_owner_extraction` | frontend evidence tooling | Retain pending extraction. |
| `scripts/generate-artifacts.sh` | `keep_pending_owner_extraction` | generated artifact orchestration | Retain aggregate generator wrapper. |
| `scripts/harness-smoke-durations.mjs` | `keep_pending_owner_extraction` | harness duration baselines | Retain pending extraction. |
| `scripts/print-explain-phase.mjs` | `keep_pending_owner_extraction` | planning explain tooling | Retain pending extraction. |
| `scripts/print-explain-run.mjs` | `keep_pending_owner_extraction` | core/planning explain tooling | Retain pending extraction. |
| `scripts/print-explain-target.mjs` | `keep_pending_owner_extraction` | planning explain tooling | Retain pending extraction. |
| `scripts/print-fixture-report.mjs` | `keep_pending_owner_extraction` | backend fixture reporting | Retain pending extraction. |
| `scripts/print-frontend-toolchain.sh` | `keep_pending_owner_extraction` | frontend readiness stamp output | Retain current Make diagnostic helper. |
| `scripts/print-target-plan.mjs` | `keep_pending_owner_extraction` | planning target-plan tooling | Retain pending extraction. |
| `scripts/print-task-guide.mjs` | `keep_pending_owner_extraction` | planning guidance tooling | Retain pending extraction. |
| `scripts/print-task-surface-report.mjs` | `keep_pending_owner_extraction` | task-surface reporting | Retain pending extraction. |
| `scripts/run-browser-e2e-a11y-preflight.sh` | `keep_pending_owner_extraction` | browser/frontend a11y tooling | Retain pending browser extraction. |
| `scripts/run-browser-e2e-a11y.sh` | `keep_pending_owner_extraction` | browser/frontend a11y tooling | Retain pending browser extraction. |
| `scripts/run-browser-e2e-batch.sh` | `keep_pending_owner_extraction` | browser orchestration | Retain as large root orchestration. |
| `scripts/run-browser-e2e-functional.sh` | `keep_pending_owner_extraction` | browser orchestration | Retain pending extraction. |
| `scripts/run-browser-e2e-manifest-dependency.sh` | `keep_pending_owner_extraction` | browser/planning orchestration | Retain pending extraction. |
| `scripts/run-browser-e2e-measurement.sh` | `keep_pending_owner_extraction` | browser measurement | Retain pending extraction. |
| `scripts/run-browser-e2e-resettable.sh` | `keep_pending_owner_extraction` | browser batch/reset wrapper | Retain; candidate future deletion with browser owner extraction. |
| `scripts/run-browser-e2e-stateful.sh` | `keep_pending_owner_extraction` | browser stateful wrapper | Retain pending extraction. |
| `scripts/run-browser-e2e-target.sh` | `keep_pending_owner_extraction` | browser target wrapper | Retain pending extraction. |
| `scripts/run-browser-e2e-visual-update.sh` | `keep_pending_owner_extraction` | browser visual update workflow | Retain pending extraction. |
| `scripts/run-browser-e2e-visual.sh` | `keep_pending_owner_extraction` | browser visual workflow | Retain pending extraction. |
| `scripts/run-browser-e2e-webserver-backed.sh` | `keep_pending_owner_extraction` | browser webserver-backed workflow | Retain pending extraction. |
| `scripts/run-fallow-static.mjs` | `keep_pending_owner_extraction` | frontend static analysis | Retain pending extraction. |
| `scripts/run-frontend-biome.sh` | `keep_pending_owner_extraction` | frontend lint wrapper | Retain pending extraction. |
| `scripts/run-frontend-unit.sh` | `keep_pending_owner_extraction` | frontend unit orchestration | Retain as large root orchestration. |
| `scripts/run-go-format.sh` | `keep_pending_owner_extraction` | Go lint/format wrapper | Retain pending extraction. |
| `scripts/run-go-gosec-audit.sh` | `keep_pending_owner_extraction` | Go security wrapper | Retain pending extraction. |
| `scripts/run-go-gosec-targeted.sh` | `keep_pending_owner_extraction` | Go security wrapper | Retain pending extraction. |
| `scripts/run-go-govulncheck.sh` | `keep_pending_owner_extraction` | Go security wrapper | Retain pending extraction. |
| `scripts/run-go-staticcheck.sh` | `keep_pending_owner_extraction` | Go lint wrapper | Retain pending extraction. |
| `scripts/run-go-vet.sh` | `keep_pending_owner_extraction` | Go lint wrapper | Retain pending extraction. |
| `scripts/run-harness-smoke.mjs` | `keep_pending_owner_extraction` | harness smoke orchestration | Retain pending extraction. |
| `scripts/run-make-node-tool.mjs` | `keep_pending_owner_extraction` | Make node-tool dispatch | Retain pending extraction. |
| `scripts/run-make-node-tool.sh` | `keep_pending_owner_extraction` | Make node-tool shell wrapper | Retain; public generated macro still names it. |
| `scripts/run-make-sequence.sh` | `keep_pending_owner_extraction` | CI/release Make sequence | Retain pending extraction. |
| `scripts/run-markdownlint.sh` | `keep_pending_owner_extraction` | markdown lint wrapper | Retain pending extraction. |
| `scripts/run-phase-slice.mjs` | `keep_pending_owner_extraction` | scheduler/planning phase slice | Retain pending extraction. |
| `scripts/run-scripts-biome.sh` | `keep_pending_owner_extraction` | scripts lint wrapper | Retain pending extraction. |
| `scripts/run-shellcheck.sh` | `keep_pending_owner_extraction` | shell lint wrapper | Retain pending extraction. |
| `scripts/service-backed-make-target-durations.mjs` | `keep_pending_owner_extraction` | scheduler duration baselines | Retain pending extraction. |
| `scripts/start-web-e2e.sh` | `keep_pending_owner_extraction` | browser stack orchestration | Retain as large root orchestration. |
| `scripts/update-go-test-durations.mjs` | `keep_pending_owner_extraction` | backend duration baseline writer | Retain pending extraction. |
| `scripts/test-agent-finalize.sh` | `keep_active_owner_test` | core finalizer owner CLI | Updated to call owner CLI with `node`. |
| `scripts/test-benchmark-claim-check.sh` | `keep_active_owner_test` | benchmark/release evidence | Retained. |
| `scripts/test-bootstrap-node-runtime.sh` | `keep_active_owner_test` | readiness bootstrap owner | Updated to owner path. |
| `scripts/test-bootstrap-shellcheck.sh` | `keep_active_owner_test` | readiness shellcheck owner | Updated to owner path. |
| `scripts/test-browser-shard-plan.sh` | `keep_active_owner_test` | browser shard planner | Retained. |
| `scripts/test-build-input-discovery.sh` | `keep_active_owner_test` | readiness build input owner | Updated to owner path. |
| `scripts/test-cache-artifact.sh` | `keep_active_owner_test` | readiness cache owner | Updated to owner path. |
| `scripts/test-cartulary-runner-service-backed-target.sh` | `keep_active_owner_test` | core runner owner CLI | Updated to owner path. |
| `scripts/test-check-migrations.sh` | `keep_active_owner_test` | backend migration owner CLI | Updated to owner path. |
| `scripts/test-check-phase-test-names.sh` | `keep_active_owner_test` | planning test-name check | Retained. |
| `scripts/test-check-scheduler.sh` | `keep_active_owner_test` | scheduler check owner CLI | Updated to owner path. |
| `scripts/test-check-toolchain-pins.sh` | `keep_active_owner_test` | toolchain/core owner paths | Updated to owner paths. |
| `scripts/test-dev-services-lifecycle.sh` | `keep_active_owner_test` | readiness local-dev owner | Updated to owner path. |
| `scripts/test-dev-stack-lifecycle.sh` | `keep_active_owner_test` | readiness dev-stack owner | Updated to owner path. |
| `scripts/test-execution-topology.sh` | `keep_active_owner_test` | generated topology/task surface | Updated agent-finalize backing script assertion. |
| `scripts/test-fallow-static.sh` | `keep_active_owner_test` | frontend static analysis | Retained. |
| `scripts/test-frontend-evidence-audit.sh` | `keep_active_owner_test` | frontend evidence audit | Retained. |
| `scripts/test-frontend-import-boundaries.sh` | `keep_active_owner_test` | frontend import boundary | Retained. |
| `scripts/test-generate-drift.sh` | `keep_active_owner_test` | generated artifact drift | Updated scratch owner paths. |
| `scripts/test-generated-artifact-policy.sh` | `keep_active_owner_test` | generated artifact policy | Retained. |
| `scripts/test-go-test-duration-baselines.sh` | `keep_active_owner_test` | backend duration baselines | Retained. |
| `scripts/test-harness-contracts.mjs` | `keep_active_owner_test` | harness contract | Retained. |
| `scripts/test-harness-smoke-duration-baselines.sh` | `keep_active_owner_test` | harness smoke duration baselines | Retained. |
| `scripts/test-json-shapes.sh` | `keep_active_owner_test` | generated JSON shape checks | Updated accessibility owner path. |
| `scripts/test-lint-shell.sh` | `keep_active_owner_test` | shell lint wrapper | Retained. |
| `scripts/test-make-node-tools.sh` | `keep_active_owner_test` | Make node-tool dispatch | Retained. |
| `scripts/test-print-target-plan.sh` | `keep_active_owner_test` | planning/backend owner CLIs | Updated owner paths; known heavy-shard invariant still failing. |
| `scripts/test-public-make-wrapper-smoke.sh` | `keep_active_owner_test` | public Make wrappers | Retained. |
| `scripts/test-release-task-surface.sh` | `keep_active_owner_test` | release task surface | Retained. |
| `scripts/test-run-frontend-unit.sh` | `keep_active_owner_test` | frontend unit orchestration | Retained. |
| `scripts/test-run-go-gosec-audit.sh` | `keep_active_owner_test` | Go security wrapper | Retained. |
| `scripts/test-run-go-gosec-targeted.sh` | `keep_active_owner_test` | Go security wrapper | Retained. |
| `scripts/test-run-go-govulncheck.sh` | `keep_active_owner_test` | Go security wrapper | Retained. |
| `scripts/test-run-go-staticcheck.sh` | `keep_active_owner_test` | Go lint wrapper | Retained. |
| `scripts/test-run-go-target.sh` | `keep_active_owner_test` | backend Go target owner runner | Updated to invoke owner runner through `node`. |
| `scripts/test-run-make-sequence-fast.sh` | `keep_active_owner_test` | Make sequence | Retained. |
| `scripts/test-run-make-sequence.sh` | `keep_active_owner_test` | Make sequence/topology | Updated expected owner paths. |
| `scripts/test-run-phase-slice.sh` | `keep_active_owner_test` | phase slice scheduler | Retained. |
| `scripts/test-run-phase.sh` | `keep_active_owner_test` | core run-phase output | Updated owner-path lint fixture. |
| `scripts/test-run-playwright-manifest-phase.sh` | `keep_active_owner_test` | browser manifest phase | Retained. |
| `scripts/test-run-playwright-phase.sh` | `keep_active_owner_test` | browser phase runner | Retained. |
| `scripts/test-run-playwright-webserver-batch.sh` | `keep_active_owner_test` | browser webserver batch | Retained. |
| `scripts/test-run-vitest-manifest-phase.sh` | `keep_active_owner_test` | frontend manifest phase | Retained. |
| `scripts/test-run-vitest-phase.sh` | `keep_active_owner_test` | frontend phase runner | Retained. |
| `scripts/test-sbom-license-evidence.mjs` | `keep_active_owner_test` | release evidence | Retained. |
| `scripts/test-seaweedfs-release-evidence.mjs` | `keep_active_owner_test` | release evidence | Retained. |
| `scripts/test-service-backed-make-target-duration-baselines.sh` | `keep_active_owner_test` | scheduler duration baselines | Retained. |
| `scripts/test-service-backed-scheduler.sh` | `keep_active_owner_test` | scheduler service-backed owner CLI | Updated to owner path. |
| `scripts/test-task-guidance.mjs` | `keep_active_owner_test` | planning guidance | Retained. |
| `scripts/test-task-surface-report.sh` | `keep_active_owner_test` | task-surface report | Retained. |
| `scripts/test-tool-output-real-targets.sh` | `keep_active_owner_test` | core output | Retained. |
| `scripts/test-web-e2e-lifecycle.sh` | `keep_active_owner_test` | browser lifecycle | Updated reset owner path. |

## Candidate Deletion And Caller-Update Opportunities

These are not part of the completed deletion set. They are the next known
places where root `scripts/` may still be acting as an implementation home or
public wrapper.

| Candidate | Remediation | Areas | Rationale | Benefit | Compatibility / migration impact | Risk if unresolved | Validation | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `scripts/check-service-backed-unit-tests.sh` | Move `service-backed-unit-check` topology command to `$(GO) run ./tools/gotestservicecheck`, regenerate task surface, then delete if no direct operator value remains. | implementation, tests, generated owner inputs | Thin wrapper over `tools/gotestservicecheck`. | Removes another root internal helper. | Generated Make recipe changes implementation path only. | Root wrapper remains misleading owner. | `bash scripts/test-execution-topology.sh`, `make phase-schedules`, `make check-harness-smoke`. | Go service-test tool owner. | Wrapper deleted and generated manifests name owner command/tool. |
| `scripts/run-make-node-tool.sh` | Keep until generated `node_tool` macro can own Node resolution directly or move shell wrapper to `tools/harness/core`. | specification, implementation, generated owner inputs | Public generated macro still names it. | Centralizes Make node-tool dispatch. | Requires task-surface generator change. | Root shell wrapper remains a core dispatch path. | `scripts/test-make-node-tools.sh`, `scripts/test-run-make-sequence.sh`, task-surface drift. | Task-surface generator and core runner owner. | Macro invokes owner path or public contract keeps wrapper. |
| `scripts/print-frontend-toolchain.sh` | Decide whether frontend readiness owner should print the stamp directly. | implementation, generated owner inputs | Small Make diagnostic wrapper, not a deleted compatibility facade. | Fewer root readiness helpers. | `frontend-toolchain` output must remain byte-compatible. | Extra root helper persists. | `make frontend-toolchain`, `make check-harness-smoke`. | Frontend/readiness owner. | Owner command prints same output or wrapper is justified. |
| Browser wrapper cluster: `run-browser-e2e-*.sh`, `start-web-e2e.sh` | Separate browser owner-extraction plan before deleting. | implementation, tests, documentation, generated owner inputs | Large browser lifecycle/orchestration remains root-owned. | Browser lifecycle code becomes owner-local. | High migration risk; public browser Make targets must remain stable. | Root browser implementation remains the largest grab-bag area. | Browser lifecycle tests and affected browser Make targets. | Browser owner extraction plan. | Owner scripts exist, task surface regenerated, wrappers deleted or justified. |
| Lint/security wrapper cluster: `run-go-*`, `run-frontend-biome.sh`, `run-scripts-biome.sh`, `run-shellcheck.sh`, `run-markdownlint.sh` | Plan owner extraction by backend/frontend/security/readiness owners. | implementation, tests, documentation, generated owner inputs | Not thin compatibility shims in this slice. | Makes lint/security ownership explicit. | Public lint target behavior must remain unchanged. | Root lint/security code remains broad. | `make lint-shell`, `make lint-scripts`, affected Go security tests. | Owner-specific lint/security plan. | Owner paths referenced directly or wrappers justified. |
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
| WS-Z Final Regeneration And Handoff | Run generators/drift checks; update tracker validation and handoff rows. | Complete after final markdown lint. | `make phase-schedules`, `make generate`, drift/shape/policy checks, final exact caller scan. |

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
| `make agent-finalize RESULTS_DIR=<successful full warm check run>` | SKIPPED | none | No successful full warm `RESULTS_DIR` was supplied or produced in this slice. |
| `make lint-markdown` | PASS | none emitted | Final tracker lint passed after this edit. |

## Handoff Log

Template for future slices:

`Time | Agent/session | Branch/commit | Inventory counts | Facades deleted | Files changed | Specs/manifests/generators touched | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice`

| Time | Agent/session | Branch/commit | Inventory counts | Facades deleted | Files changed | Specs/manifests/generators touched | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T13:21Z | Codex / Iteration 3 | `main` / `54e926e4` base | Pre: 145 total, 141 root, 4 CI, 52 tests, 0 `scripts/lib`; post: 113 total, 109 root, 4 CI, 52 tests, `scripts/lib` absent | 32 root facades across WS-01..WS-05 | Makefile, active docs/tests, owner harness files, generated task-surface/topology/design-token outputs, deleted facades | `tools/execution_topology_manifest.json`, `tools/generate_drift_scratch_inputs.json`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_render_index.json`, `packages/ui-contracts/src/generated/design-tokens.ts` | See validation table; generator roots: `.cartulary/test-results/20260703T132103Z-p585384`, `.cartulary/test-results/20260703T132123Z-p585707`; drift/shape roots listed above | `make agent-finalize RESULTS_DIR=<successful full warm check run>` skipped because no successful full warm run was available | `scripts/test-print-target-plan.sh` still fails known heavy-shard split invariant; remaining root orchestration wrappers need separate extraction plan | Future extraction candidates listed above; no deleted-facade callers remain. |

## Open Questions / Blockers

- External consumers of root `scripts/*` paths are unknown. Default retained external entrypoint is `scripts/ci/verify.sh`.
- `scripts/harness-contract.sh` is frozen by `docs/testing-harness-nlspec.md`; removing it is blocked on spec revision.
- `start-web-e2e.sh` and other large browser/lint/security scripts are not thin facades and remain future owner-extraction work.
- `scripts/test-print-target-plan.sh` still fails the prior heavy-shard split invariant; this slice did not change that planning behavior.
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
