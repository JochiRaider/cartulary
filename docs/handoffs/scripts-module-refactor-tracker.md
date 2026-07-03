# scripts Directory Refactor Tracker, Iteration 2

## Current baseline and inventory method

This tracker supersedes the completed first-iteration remediation log. The prior
work remains relevant only as historical baseline: SL-01 through SL-11 moved
generic reusable helpers out of `scripts/lib/` into owner paths such as
`tools/harness/core`, `tools/harness/scheduler`,
`tools/harness/generated-artifacts`, `tools/harness/browser`,
`tools/harness/frontend`, `tools/harness/planning`,
`tools/harness/backend`, `tools/harness/readiness`,
`tools/harness/test-support`, `tools/otel`, and
`tools/release-evidence`.

Iteration 2 starts from the live repository state and treats `scripts/` as a
repository automation surface, not as a product module. It must not introduce
`internal/modules/scripts`, `packages/scripts`, or any equivalent product-facing
module. Public Make targets, command IDs, schemas, artifact paths, failure
taxonomy, generated-output ownership, and cleanup predicates remain frozen
unless an owner spec explicitly changes them.

| Item | Value |
| --- | --- |
| Rebaseline date | 2026-07-03 |
| Branch | `main` |
| Commit | `3c2e9c56` |
| Working tree at rebaseline | clean before tracker edit |
| Tracker path | `docs/handoffs/scripts-module-refactor-tracker.md` |
| Total live files under `scripts/` | 146 |
| Root files under `scripts/` | 142 |
| Files under `scripts/ci/` | 4 |
| Root `scripts/test-*` files | 52 |
| Files under `scripts/lib/` | 0 |
| Direct Make/task-surface/topology `scripts/` references | 130 live paths |
| Live `scripts/` paths not directly referenced by Make/task-surface/topology | 16 paths |
| Pre-edit validation baseline | `make lint-markdown` passed at `.cartulary/test-results/20260703T044945Z-p430930` |
| Post-edit validation recorded by this iteration | `CARTULARY_TEST_RUN_ID=20260703T050120Z-ptracker2-final make lint-markdown` passed at `.cartulary/test-results/20260703T050120Z-ptracker2-final` |

Inventory inputs inspected for this iteration:

- `docs/handoffs/scripts-module-refactor-tracker.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `AGENTS.md`
- `Makefile`
- `tools/task_surface_manifest.json`
- `tools/execution_topology_manifest.json`
- live `scripts/` traversal with `find`
- direct `scripts/` references from Make, task-surface, and topology manifests
- current non-archive references to retired `scripts/lib/*` paths

Inventory commands used:

| Purpose | Command |
| --- | --- |
| Live file list | `find scripts -type f -o -type l \| sort` |
| Directory shape | `find scripts -maxdepth 2 -type d \| sort` |
| Empty helper directory check | `find scripts/lib -mindepth 1 -print` |
| Direct command/reference scan | `rg -o -P "(?<![A-Za-z0-9_./-])(\\./)?scripts/[A-Za-z0-9_./-]+" Makefile tools/task_surface_manifest.json tools/execution_topology_manifest.json` |
| Unreferenced live paths | `comm -23 <(find scripts -type f -o -type l \| sort) <(rg ... \| sort -u)` |
| Stale `scripts/lib` references | `rg -n "scripts/lib" docs/handoffs docs/guides AGENTS.md Makefile tools scripts --glob '!docs/archive/**' --glob '!docs/testing-harness-spec-recovery-docs/**'` |

The 16 live paths not directly referenced by Make/task-surface/topology are:

`scripts/.gitkeep`, `scripts/check-font-bundle.mjs`,
`scripts/check-go-target-plan-coverage.mjs`,
`scripts/check-migration-history.mjs`, `scripts/check-phase-map.mjs`,
`scripts/check-postgres-fixture-budget.mjs`,
`scripts/check-schema-object-ownership.mjs`, `scripts/ci/verify.sh`,
`scripts/diagnose-inotify.mjs`, `scripts/generate-design-tokens.mjs`,
`scripts/harness-contract.mjs`, `scripts/harness-contract.sh`,
`scripts/list-build-inputs.sh`, `scripts/print-go-shard-plan.mjs`,
`scripts/reset-web-e2e-stack.sh`, and
`scripts/run-browser-e2e-owned-stack.sh`.

## Owner-family map

| Owner family | Owner area | Current rule | Durable destination |
| --- | --- | --- | --- |
| Stable Make/public wrapper | multiple | Keep in `scripts/` when Make or the public task surface owns the invocation contract. | `scripts/` wrapper plus reusable logic in `tools/harness/*`, `tools/otel`, or `tools/release-evidence`. |
| Internal Make wrapper | implementation | Keep in `scripts/` only when generated Make or a Make-owned internal helper invokes it. | Thin `scripts/` wrapper or owner-specific `tools/*` CLI. |
| Operator/local-dev/CI wrapper | multiple | Retain only when it provides continuing operator, local-dev, recovery, deployable-shape, or provider-neutral CI value. | `scripts/` wrapper, `scripts/ci/`, or owner-specific `tools/harness/readiness`. |
| Target-local test | tests | Retain when it is a direct task-surface self-test or owner-local characterization test. | `scripts/test-*` for entrypoint tests; reusable fixtures in `tools/harness/test-support`. |
| Harness core implementation | implementation | Shared output, schema, artifact, failure, redaction, cache, or finalizer logic must not live as reusable root script logic. | `tools/harness/core`. |
| Harness scheduler implementation | implementation | Scheduler algorithms, resource logic, durations, event ordering, and summary logic are owner code. | `tools/harness/scheduler`. |
| Harness planning implementation | implementation | Phase, target, task-surface, explain, and guidance logic belongs to planning owner paths. | `tools/harness/planning` or `tools/harness/generated-artifacts`. |
| Harness generated implementation | implementation | Drift/generation and generated artifact checks must remain owner-input driven. | `tools/harness/generated-artifacts`. |
| Harness backend implementation | implementation | Go target, migration, schema ownership, package fixture, and backend static checks belong to backend harness support. | `tools/harness/backend`. |
| Harness frontend implementation | implementation | Frontend package-boundary, design-token, font, accessibility, and frontend evidence support belongs to frontend harness support. | `tools/harness/frontend`. |
| Harness browser implementation | implementation | Browser E2E stack, reset, Playwright stage, visual, accessibility, and shard support belongs to browser harness support. | `tools/harness/browser`. |
| Harness readiness implementation | implementation | Tool install, build artifact, cache, inotify, process lifecycle, and local readiness support belongs to readiness owner paths. | `tools/harness/readiness`, with build helpers under matching backend/frontend owners where useful. |
| Harness test-support implementation | tests | Shared scratch, JSON, artifact assertion, and fixture utilities must not be hidden in root tests. | `tools/harness/test-support`. |
| Release evidence | multiple | Core 05, SBOM, release, and object-store evidence tooling must not be treated as Base Profile behavior. | `tools/release-evidence`; root scripts may keep tests only. |
| Documentation-only/stale reference | documentation | Active docs must not point developers at retired `scripts/lib` paths. Historical archive/recovery docs may remain historical. | Update active docs; leave archived source-limit docs alone unless republished. |
| No-behavior placeholder | documentation | Placeholder files have no runtime contract. | Delete if no longer needed, or explicitly mark out of behavior scope. |

## Public contract and behavior freeze map

| Surface | Frozen behavior | Owner | Validation before any future implementation slice |
| --- | --- | --- | --- |
| Public Make target names | Keep every current public target name and invocation through `make <target>`. | `docs/testing-harness-nlspec.md`, `Makefile`, task-surface manifest | `make harness-contract`; affected `make explain-target TARGET=<target> DETAIL=summary` |
| `command_id` values | Preserve stable `cartulary.harness.command.*.v1` identities. | Harness NLSpec and task-surface manifest | `make harness-contract`; `make json-shape-check` when manifests change |
| Output classes and summary schemas | Preserve output mode behavior, stable summary schemas, and machine-output constraints. | Harness NLSpec Sections 7 and 8 | `make harness-contract`; affected target direct run |
| Artifact paths and retained run identity | Preserve `.cartulary/test-results`, run IDs, target-summary paths, scheduler artifacts, and release artifact paths. | Harness NLSpec and generated manifests | affected owner target plus artifact assertions |
| Failure taxonomy | Preserve failure classes, reasons, exit mapping, and child failure normalization. | `tools/harness/core/failure-taxonomy.mjs` plus Harness NLSpec | scheduler and harness contract tests |
| Cache schema IDs | Preserve `cartulary.cache.readiness.v1`, `cartulary.cache.build_artifact.v1`, `cartulary.agent_finalize_action_cache_record.v1`, and `cartulary.execution_topology_render_cache.v1`. | Harness NLSpec cache requirements | cache direct tests, readiness/build target |
| Generated outputs | Do not hand-edit generated roots or generated harness/topology outputs. | `tools/generated_artifact_policy.json`, execution topology manifest | `make generated-artifact-policy-check`; `make json-shape-check`; drift targets |
| Cleanup/destructive predicates | Preserve confirmation env, dry-run behavior, protected-path checks, and reset proof predicates. | Harness NLSpec and Core 04 boundaries | cleanup tests, browser reset tests, local-dev smoke where touched |
| OTel generator provenance | Preserve canonical OTel owner paths unless the OTel NLSpec changes first. | OTel NLSpec | `make otel-conformance`; `make generate-drift` if touched |
| Core 05/release evidence boundary | Keep claim-publication and release evidence out of Base Profile runtime behavior. | Core 05 and release evidence support | benchmark/release evidence tests and targets |

## Per-file classification matrix

Legend for inbound references: `MTS` means Make/task-surface/topology names the path;
`Make` means Make references the path outside generated target metadata;
`Indirect` means scripts/tests/imports call the path but Make/task-surface/topology
does not name it directly; `None` means no behavior caller was found in this
iteration.

| Path | Owner family | Owner area | Public contract status | Role | Inbound references | Risk | Likely disposition | Dependencies | Validation | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `scripts/.gitkeep` | No-behavior placeholder | documentation | none | placeholder | None | low | delete or retain as out-of-scope | none | `make lint-markdown` if doc-only | no behavior claim |
| `scripts/agent-finalize.mjs` | Stable Make/public wrapper | multiple | public Make input | finalizer CLI and action orchestration | MTS | high | retain wrapper; move only reusable action logic | `tools/harness/core`, scheduler drift | `make agent-finalize RESULTS_DIR=<run>` | public finalizer unchanged |
| `scripts/agent-finalize.sh` | Stable Make/public wrapper | implementation | public Make input | shell entry wrapper | MTS | high | retain thin wrapper | `agent-finalize.mjs` | `make agent-finalize RESULTS_DIR=<run>` | wrapper remains stable |
| `scripts/bootstrap-go-tool.sh` | Harness readiness implementation | implementation | Make readiness input | pinned Go tool install helper | MTS | medium | thin wrapper or move install logic to readiness/backend owner | Make pins, cache helper | bootstrap/toolchain tests | cache and install contracts preserved |
| `scripts/bootstrap-node-runtime.sh` | Harness readiness implementation | implementation | public readiness helper | Node runtime bootstrap | MTS | medium | retain wrapper; extract reusable download/check logic if needed | toolchain pins | `test-bootstrap-node-runtime`; `make frontend-toolchain` | readiness behavior stable |
| `scripts/bootstrap-shellcheck.sh` | Harness readiness implementation | implementation | public readiness helper | ShellCheck bootstrap | MTS | medium | retain wrapper; extract reusable download/check logic if needed | toolchain pins | `test-bootstrap-shellcheck`; `make lint-shell` | pinned tool behavior stable |
| `scripts/build-go-artifact.sh` | Harness backend implementation | implementation | Make build helper | Go binary build wrapper | MTS | medium | retain thin wrapper; owner logic under backend/readiness if expanded | Go build cache, cache helper | build targets | build-artifact cache contract stable |
| `scripts/build-web-artifact.sh` | Harness frontend implementation | implementation | Make build helper | web build wrapper | MTS | medium | retain thin wrapper; owner logic under frontend/readiness if expanded | frontend install, cache helper | `make build-web` if touched | build output stable |
| `scripts/cache-artifact.sh` | Stable Make/public wrapper | implementation | cache schema producer | compatibility wrapper for readiness/build cache implementation | MTS | high | retained thin wrapper; implementation moved to `tools/harness/readiness/cache-artifact.sh` | Harness cache schemas | `test-cache-artifact`; readiness/build targets | cache keys and records stable |
| `scripts/cartulary-runner.mjs` | Stable Make/public wrapper | multiple | public runner wrapper | aggregate runner | MTS | high | retain wrapper; move reusable runner algorithms only | core runner context, topology | runner tests; `make check-harness-smoke` | public runner unchanged |
| `scripts/check-backend-module-boundaries.mjs` | Stable Make/public wrapper | implementation | public Make input | backend module boundary check | MTS | high | retain wrapper; move reusable analyzer if needed | Core 01 boundary support | `make backend-module-boundary-check` | boundary verdict stable |
| `scripts/check-doctor.sh` | Operator/local-dev/CI wrapper | multiple | public Make input | local doctor checks | MTS | medium | retain operator wrapper; move reusable diagnostics | inotify diagnostic, toolchain | `make doctor` | doctor output stable |
| `scripts/check-font-bundle.mjs` | Harness frontend implementation | implementation | indirect/frontend test helper | vendored font bundle validator and fixture creator | Indirect | medium | move reusable font checker to `tools/harness/frontend`; optional wrapper | frontend font assets | frontend unit/font tests | owner path or retention rationale |
| `scripts/check-frontend-import-boundaries.mjs` | Stable Make/public wrapper | implementation | public Make input | frontend import boundary check | MTS | high | retain wrapper; move reusable analyzer if needed | design-token owner helper | `make frontend-import-boundary-check` | boundary behavior stable |
| `scripts/check-go-target-plan-coverage.mjs` | Harness backend implementation | implementation | indirect/backend test helper | Go target plan coverage checker | Indirect | medium | move to `tools/harness/backend` or retain wrapper | backend target plan | go target tests | coverage checker owner-aligned |
| `scripts/check-go-test-duration-baseline-coverage.mjs` | Stable Make/public wrapper | implementation | public duration target input | duration baseline coverage checker | MTS | high | retain wrapper; move reusable logic only | backend duration baselines | `make go-test-duration-baseline-coverage` | duration coverage stable |
| `scripts/check-go-test-duration-baseline-drift.mjs` | Stable Make/public wrapper | implementation | public duration drift input | Go duration drift checker | MTS | high | retain wrapper; move reusable logic only | backend duration artifacts, scheduler drift | duration drift target | drift verdict stable |
| `scripts/check-migration-history.mjs` | Harness backend implementation | implementation | indirect migration checker | migration history manifest checker | Indirect | medium | move CLI to backend owner or retain wrapper | migration history owner helper | migration tests; `make migration-drift` | migration checks owner-aligned |
| `scripts/check-migrations.sh` | Internal Make wrapper | implementation | migration drift input | migration input/scratch wrapper | MTS | high | retain wrapper if Make-owned; move reusable pieces only | migration helper, goose | `make migration-drift` | drift behavior stable |
| `scripts/check-phase-map.mjs` | Harness planning implementation | implementation | indirect planning helper | single phase-map validator wrapper | Indirect | low | simplify or replace with owner CLI | phase-manifest owner | phase-map tests | no duplicate wrapper unless useful |
| `scripts/check-phase-maps.sh` | Internal Make wrapper | implementation | internal Make input | all phase-map validator wrapper | MTS | high | retain wrapper; logic already owner-path based | planning/frontend manifests | phase-map check target | phase validation stable |
| `scripts/check-phase-test-names.mjs` | Internal Make wrapper | tests | check-internal input | phase test-name guard | MTS | medium | retain until owner test moved | phase maps | `test-check-phase-test-names`; check-internal | guard stable |
| `scripts/check-postgres-fixture-budget.mjs` | Harness backend implementation | implementation | indirect scheduler helper | Postgres fixture budget evaluator | Indirect | high | move reusable evaluator to backend owner; wrapper optional | fixture reporting, Go shard plan | service-backed scheduler tests | fixture budget owner-aligned |
| `scripts/check-scheduler-event-order-drift.mjs` | Stable Make/public wrapper | implementation | scheduler drift input | scheduler event-order drift checker | MTS | high | retain wrapper; scheduler logic already owner path | scheduler event-order | scheduler drift tests | event order stable |
| `scripts/check-scheduler-summary-timing-drift.mjs` | Stable Make/public wrapper | implementation | scheduler drift input | scheduler summary timing drift checker | MTS | high | retain wrapper; move reusable logic only | scheduler timing drift | scheduler timing drift target | timing verdict stable |
| `scripts/check-schema-object-ownership.mjs` | Harness backend implementation | implementation | indirect schema checker | schema ownership manifest checker | Indirect | medium | move CLI to backend owner or retain wrapper | schema ownership helper | `make json-shape-check` | schema ownership stable |
| `scripts/check-service-backed-unit-tests.sh` | Internal Make wrapper | tests | check-internal input | service-backed unit self-test runner | MTS | medium | retain target-local wrapper | service-backed tests | check-internal | self-test stable |
| `scripts/check-toolchain-pins.mjs` | Stable Make/public wrapper | implementation | public drift input | toolchain pin drift checker | MTS | medium | retain wrapper; move reusable pin parser if needed | toolchain pins | `make toolchain-drift`; direct test | pin drift stable |
| `scripts/ci/check-deployable-shape.sh` | Operator/local-dev/CI wrapper | multiple | public/internal release input | deployable package shape smoke | MTS | high | retain CI/release wrapper | build artifacts, embedded web archive | `make deployable-shape`; `make release-check` if touched | deployable contract stable |
| `scripts/ci/check-standup-operational-recovery-smoke.sh` | Operator/local-dev/CI wrapper | multiple | public Make input | recovery smoke wrapper | MTS | high | retain if current recovery value remains | deploy recovery scripts | standup recovery target | recovery smoke stable or retired explicitly |
| `scripts/ci/check-standup-package-smoke.sh` | Operator/local-dev/CI wrapper | multiple | public Make input | package smoke wrapper | MTS | high | retain if current package value remains | built artifacts | standup package target | package smoke stable or retired explicitly |
| `scripts/ci/verify.sh` | Operator/local-dev/CI wrapper | multiple | external/manual CI wrapper | provider-neutral CI dispatcher | Indirect | medium | retain if external CI still uses it; document source limit | Make `ci` target | `make ci` only when necessary | CI wrapper value explicit |
| `scripts/dev-services.sh` | Operator/local-dev/CI wrapper | multiple | public local-dev input | Compose service lifecycle wrapper | MTS | high | retain operator wrapper; move reusable lifecycle only | readiness lifecycle, Compose | local-dev service tests; `make services-up` | lifecycle behavior stable |
| `scripts/dev-stack.sh` | Operator/local-dev/CI wrapper | multiple | public local-dev input | dev stack wrapper | MTS | high | retain operator wrapper; move reusable diagnostics only | process lifecycle, inotify | `test-dev-stack-lifecycle`; `make dev` if touched | dev workflow stable |
| `scripts/diagnose-inotify.mjs` | Harness readiness implementation | implementation | indirect doctor/dev helper | Linux inotify diagnostics | Indirect | medium | move to readiness owner with wrapper compatibility | `/proc`, env thresholds | `make doctor`; dev-stack tests | diagnostic owner-aligned |
| `scripts/duration-baseline-drift-suite.sh` | Internal Make wrapper | implementation | check-internal/public support | duration drift suite dispatcher | MTS | high | retain wrapper while duration targets remain Make-owned | duration baseline helpers | duration suite target | suite behavior stable |
| `scripts/embed-web-assets.sh` | Harness frontend implementation | implementation | Make build helper | embedded web asset producer wrapper | MTS | high | retain thin wrapper; reusable embed logic stays in `tools/embedwebassets` | build cache, Go embed | build/deployable targets | archive contract stable |
| `scripts/frontend-evidence-audit.mjs` | Stable Make/public wrapper | implementation | public Make input | frontend evidence audit | MTS | high | retain wrapper; owner logic under frontend if expanded | frontend phase manifest | `make frontend-evidence-audit` | evidence boundary stable |
| `scripts/frontend-install.sh` | Harness frontend implementation | implementation | public readiness helper | pnpm install wrapper | MTS | medium | retain wrapper; move reusable readiness only | pnpm/toolchain/cache | `make frontend-install` | install readiness stable |
| `scripts/frontend-toolchain.sh` | Harness readiness implementation | implementation | public readiness helper | Node/pnpm readiness wrapper | MTS | medium | retain wrapper; move reusable toolchain checks if needed | toolchain pins/cache | `make frontend-toolchain` | readiness stable |
| `scripts/generate-artifacts.sh` | Internal Make wrapper | implementation | internal generated input | aggregate generated-artifact wrapper | MTS | high | retain wrapper; owner inputs stay in tools | sqlc, OTel, design tokens | `make generate`; drift targets | generation stable |
| `scripts/generate-design-tokens.mjs` | Harness frontend implementation | implementation | indirect generated input | design-token generator CLI | Indirect | high | move under frontend/generated owner only with provenance update | `docs/design.md`, UI contracts output | `make generate`; `json-shape-check` | generated token provenance truthful |
| `scripts/harness-contract.mjs` | Stable Make/public wrapper | implementation | compatibility-only direct script | compatibility trampoline for harness contract CLI | Indirect | high | retain temporarily while callers move; canonical implementation is `tools/harness/core/harness-contract-cli.mjs` | `tools/harness/core` | `make harness-contract` | public CLI stable |
| `scripts/harness-contract.sh` | Stable Make/public wrapper | implementation | public Make helper wrapper | shell wrapper for harness contract CLI | Indirect/Make | high | retained stable wrapper; invokes `tools/harness/core/harness-contract-cli.mjs` | core harness CLI | `make harness-contract` | wrapper stable |
| `scripts/harness-smoke-durations.mjs` | Stable Make/public wrapper | implementation | duration baseline input | harness smoke duration checker/updater | MTS | medium | retain wrapper; scheduler logic owner path | scheduler duration helpers | harness smoke duration target | baseline behavior stable |
| `scripts/list-build-inputs.sh` | Harness readiness implementation | implementation | Make build helper | build input discovery | Indirect/Make | medium | move logic to readiness/build owner or retain tiny wrapper | `rg`, Make build vars | `test-build-input-discovery`; build targets | build inputs stable |
| `scripts/playwright-install.sh` | Harness browser implementation | implementation | public readiness helper | Playwright install wrapper | MTS | medium | retain wrapper; move reusable readiness only | frontend install/cache | `make playwright-install` | install stable |
| `scripts/print-explain-phase.mjs` | Stable Make/public wrapper | implementation | public Make input | explain phase CLI | MTS | medium | retain wrapper; planning logic owner path | planning guidance, frontend manifest | `make explain-phase PHASE=phaseN` | output stable |
| `scripts/print-explain-run.mjs` | Stable Make/public wrapper | implementation | public Make input | explain retained run CLI | MTS | medium | retain wrapper; core logic owner path | failure taxonomy, artifacts | `make explain-run RESULTS_DIR=<dir>` | diagnostics stable |
| `scripts/print-explain-target.mjs` | Stable Make/public wrapper | implementation | public Make input | explain target CLI | MTS | medium | retain wrapper; planning logic owner path | target plan | `make explain-target TARGET=<target>` | diagnostics stable |
| `scripts/print-fixture-report.mjs` | Stable Make/public wrapper | implementation | public Make input | fixture report CLI | MTS | medium | retain wrapper; fixture logic owner path | fixture reporting | `make fixture-report` if touched | report stable |
| `scripts/print-frontend-toolchain.sh` | Stable Make/public wrapper | implementation | public Make input | frontend toolchain diagnostic | MTS | low | retain thin wrapper | toolchain env | `make frontend-toolchain` | diagnostic stable |
| `scripts/print-go-shard-plan.mjs` | Harness backend implementation | implementation | indirect backend diagnostic | Go shard plan printer | Indirect | medium | move under backend owner or retain wrapper | go-shard-plan | go target tests | owner-aligned diagnostic |
| `scripts/print-target-plan.mjs` | Stable Make/public wrapper | implementation | public Make input | target-plan printer | MTS | medium | retain wrapper; planning logic owner path | target-plan | `make target-plan` | output stable |
| `scripts/print-task-guide.mjs` | Stable Make/public wrapper | implementation | public Make input | task guide CLI | MTS | medium | retain wrapper; guidance logic owner path | task guidance | `make task-guide ROLE=<role>` | guide stable |
| `scripts/print-task-surface-report.mjs` | Stable Make/public wrapper | implementation | public Make input | task-surface report CLI | MTS | medium | retain wrapper; generated/planning logic owner path | task-surface manifest | `make task-surface-report` | report stable |
| `scripts/reset-web-e2e-stack.sh` | Stable Make/public wrapper | implementation | indirect browser reset helper | compatibility wrapper for runtime reset implementation | Indirect | high | retained wrapper; implementation moved to `tools/harness/browser/reset-web-e2e-stack.sh` | test runtime reset route, harness contract | browser lifecycle tests | reset artifacts stable |
| `scripts/run-browser-e2e-a11y-preflight.sh` | Stable Make/public wrapper | implementation | public browser target input | accessibility preflight wrapper | MTS | high | retain wrapper; browser logic owner path | Playwright stack | browser a11y preflight target | target behavior stable |
| `scripts/run-browser-e2e-a11y.sh` | Stable Make/public wrapper | implementation | public browser target input | accessibility wrapper | MTS | high | retain wrapper; browser logic owner path | Playwright stack | browser a11y target | target behavior stable |
| `scripts/run-browser-e2e-batch.sh` | Harness browser implementation | implementation | browser batch helper | browser batch runner wrapper | MTS | high | retain if task-surface uses it; move reusable batch logic | browser manifest, reset | browser batch tests | batch artifacts stable |
| `scripts/run-browser-e2e-functional.sh` | Stable Make/public wrapper | implementation | public browser target input | functional browser wrapper | MTS | high | retain wrapper | Playwright stack | browser functional target | target behavior stable |
| `scripts/run-browser-e2e-manifest-dependency.sh` | Harness browser implementation | implementation | browser helper input | manifest dependency wrapper | MTS | medium | retain or move under browser owner with wrapper | phase manifest, Playwright | browser manifest tests | dependency behavior stable |
| `scripts/run-browser-e2e-measurement.sh` | Stable Make/public wrapper | implementation | public browser target input | measurement browser wrapper | MTS | high | retain wrapper | Playwright stack | browser measurement target | target behavior stable |
| `scripts/run-browser-e2e-owned-stack.sh` | Harness browser implementation | implementation | indirect browser helper | owned-stack wrapper | Indirect | high | keep only if needed by stateful wrapper; otherwise fold into owner helper | Playwright owned stack | browser lifecycle tests | no duplicate stack wrapper |
| `scripts/run-browser-e2e-resettable.sh` | Stable Make/public wrapper | implementation | public browser target input | resettable browser wrapper | MTS | high | retain wrapper | reset helper, batch runner | browser resettable target | target behavior stable |
| `scripts/run-browser-e2e-stateful.sh` | Stable Make/public wrapper | implementation | public browser target input | stateful browser wrapper | MTS | high | retain wrapper; simplify owned-stack delegation if possible | owned-stack wrapper | browser stateful target | target behavior stable |
| `scripts/run-browser-e2e-target.sh` | Harness browser implementation | implementation | browser helper input | browser target summary wrapper | MTS | high | retain if Make/task-surface-owned; move reusable logic | test-output, batch manifest | browser tests | target summaries stable |
| `scripts/run-browser-e2e-visual-update.sh` | Stable Make/public wrapper | implementation | public browser target input | visual golden update wrapper | MTS | high | retain wrapper; preserve explicit visual-update semantics | visual target, manifest dependency | browser visual update target | update behavior stable |
| `scripts/run-browser-e2e-visual.sh` | Stable Make/public wrapper | implementation | public browser target input | visual browser wrapper | MTS | high | retain wrapper | Playwright stack | browser visual target | visual evidence stable |
| `scripts/run-browser-e2e-webserver-backed.sh` | Stable Make/public wrapper | implementation | public browser target input | webserver-backed browser wrapper | MTS | high | retain wrapper | Playwright stack | browser webserver target | target behavior stable |
| `scripts/run-check-schedule.mjs` | Stable Make/public wrapper | multiple | public scheduler input | check scheduler runner | MTS | high | retain wrapper; move reusable runtime attach logic only | scheduler/core/browser/planning | check scheduler tests | scheduler summaries stable |
| `scripts/run-fallow-static.mjs` | Stable Make/public wrapper | implementation | public static target input | Fallow static wrapper | MTS | medium | retain wrapper; core output owner path | fallow config, tool output | `test-fallow-static`; target | static summary stable |
| `scripts/run-frontend-biome.sh` | Stable Make/public wrapper | implementation | public lint input | frontend Biome wrapper | MTS | medium | retain wrapper | frontend toolchain, test-output | `make lint-biome` | lint output stable |
| `scripts/run-frontend-unit.sh` | Stable Make/public wrapper | implementation | public frontend unit input | Vitest phase wrapper | MTS | high | retain wrapper; frontend logic owner path | frontend phase manifest | `make frontend-unit`; direct test | frontend artifacts stable |
| `scripts/run-go-format.sh` | Internal Make wrapper | implementation | lint/format input | Go format wrapper | MTS | medium | retain wrapper; generated filter owner path | generated-artifacts filters | `make lint-go-format`; `make format` | generated roots excluded |
| `scripts/run-go-gosec-audit.sh` | Stable Make/public wrapper | implementation | public security input | gosec audit wrapper | MTS | high | retain wrapper | pinned gosec | gosec audit tests | security target stable |
| `scripts/run-go-gosec-targeted.sh` | Stable Make/public wrapper | implementation | public security input | gosec targeted wrapper | MTS | high | retain wrapper | pinned gosec | gosec targeted tests | security target stable |
| `scripts/run-go-govulncheck.sh` | Stable Make/public wrapper | implementation | public security input | govulncheck wrapper | MTS | high | retain wrapper | pinned govulncheck | govulncheck tests | security target stable |
| `scripts/run-go-staticcheck.sh` | Internal Make wrapper | implementation | lint input | staticcheck wrapper | MTS | medium | retain wrapper | generated filters, staticcheck | staticcheck tests | lint behavior stable |
| `scripts/run-go-target.mjs` | Stable Make/public wrapper | implementation | public Go target input | Go target runner wrapper | MTS | high | retain wrapper; backend logic owner path | backend go-target-runner | go target tests | Go target summaries stable |
| `scripts/run-go-vet.sh` | Internal Make wrapper | implementation | lint input | Go vet wrapper | MTS | medium | retain wrapper | generated filters | `make lint-go-vet` | lint behavior stable |
| `scripts/run-harness-smoke.mjs` | Stable Make/public wrapper | tests | public harness smoke input | harness smoke dispatcher | MTS | high | retain wrapper; move reusable smoke logic only | core/tool output | `make check-harness-smoke` | smoke tier stable |
| `scripts/run-make-node-tool.mjs` | Stable Make/public wrapper | implementation | internal/public helper input | normalized Node tool wrapper | MTS | medium | retain wrapper; core logic owner path | make-node-tools, tool output | `test-make-node-tools` | tool summaries stable |
| `scripts/run-make-node-tool.sh` | Stable Make/public wrapper | implementation | helper input | shell wrapper for Node tool runner | MTS | medium | retain thin wrapper | run-make-node-tool.mjs | make-node tests | wrapper stable |
| `scripts/run-make-sequence.sh` | Stable Make/public wrapper | implementation | public aggregate helper | Make sequence runner | MTS | high | retain wrapper; planning logic owner path | task-surface sequence definitions | sequence tests | aggregate behavior stable |
| `scripts/run-markdownlint.sh` | Stable Make/public wrapper | implementation | public lint input | Markdown lint wrapper | MTS | low | retain wrapper | frontend install/toolchain | `make lint-markdown` | lint behavior stable |
| `scripts/run-phase-slice.mjs` | Stable Make/public wrapper | multiple | public phase-slice input | phase scheduler wrapper | MTS | high | retain wrapper; move reusable slice algorithms only | planning/frontend/scheduler/core | phase-slice tests | slice summaries stable |
| `scripts/run-scripts-biome.sh` | Stable Make/public wrapper | implementation | public lint input | script Biome wrapper | MTS | medium | retain wrapper | frontend toolchain | `make lint-scripts` | lint behavior stable |
| `scripts/run-service-backed-schedule.mjs` | Stable Make/public wrapper | multiple | public scheduler input | service-backed scheduler runner | MTS | high | retain wrapper; move reusable runtime attach logic only | scheduler/browser/backend/planning | service-backed scheduler tests | scheduler summaries stable |
| `scripts/run-shellcheck.sh` | Stable Make/public wrapper | implementation | public lint input | ShellCheck wrapper | MTS | medium | retain wrapper | generated filters, shellcheck | `make lint-shell`; direct test | shell lint stable |
| `scripts/service-backed-make-target-durations.mjs` | Stable Make/public wrapper | implementation | duration baseline input | service-backed Make duration updater/checker | MTS | medium | retain wrapper; scheduler logic owner path | execution topology, duration helpers | duration baseline tests | baseline behavior stable |
| `scripts/start-web-e2e.sh` | Stable Make/public wrapper | multiple | browser stack public/helper input | browser E2E stack lifecycle | MTS | high | retain wrapper; move reusable lifecycle chunks only | browser lifecycle, dev-services | browser lifecycle tests | session artifacts stable |
| `scripts/test-agent-finalize.sh` | Target-local test | tests | active harness smoke check | finalizer characterization | MTS | medium | retained as `harness-smoke-agent-finalize` in extended/full tiers | agent-finalize | direct test; `make run-harness-smoke-extended` | active owner test |
| `scripts/test-benchmark-claim-check.sh` | Target-local test | tests | check-internal self-test | benchmark claim checker test | MTS | medium | retain test while release helper exists | `tools/release-evidence` | direct test; benchmark target | Core 05 boundary covered |
| `scripts/test-bootstrap-node-runtime.sh` | Target-local test | tests | check-internal self-test | Node bootstrap test | MTS | medium | retain or move under readiness tests | bootstrap node wrapper | direct test | readiness coverage stable |
| `scripts/test-bootstrap-shellcheck.sh` | Target-local test | tests | check-internal self-test | ShellCheck bootstrap test | MTS | medium | retain or move under readiness tests | shellcheck bootstrap | direct test | readiness coverage stable |
| `scripts/test-browser-shard-plan.sh` | Target-local test | tests | check-internal self-test | browser shard planner test | MTS | medium | retain; reusable fixtures stay in test-support | browser shard plan | direct test | browser planning covered |
| `scripts/test-build-input-discovery.sh` | Target-local test | tests | check-internal self-test | build input discovery test | MTS | medium | retain until helper disposition closed | list-build-inputs | direct test | build input coverage stable |
| `scripts/test-cache-artifact.sh` | Target-local test | tests | check-internal self-test | cache helper test | MTS | high | retain before WS-01 | cache-artifact | direct test | cache contract covered |
| `scripts/test-cartulary-runner-service-backed-target.sh` | Target-local test | tests | check-internal self-test | cartulary runner service-backed test | MTS | high | retain | cartulary-runner | direct test | runner coverage stable |
| `scripts/test-check-migrations.sh` | Target-local test | tests | check-internal self-test | migration check test | MTS | medium | retain; backend helper may move | check-migrations/history | direct test | migration coverage stable |
| `scripts/test-check-phase-test-names.sh` | Target-local test | tests | check-internal self-test | phase test-name guard test | MTS | medium | retain | check-phase-test-names | direct test | guard coverage stable |
| `scripts/test-check-scheduler.sh` | Target-local test | tests | check-internal self-test | check scheduler test | MTS | high | retain | run-check-schedule | direct test; harness smoke | scheduler coverage stable |
| `scripts/test-check-toolchain-pins.sh` | Target-local test | tests | check-internal self-test | toolchain pin checker test | MTS | medium | retain | check-toolchain-pins | direct test | pin coverage stable |
| `scripts/test-dev-services-lifecycle.sh` | Target-local test | tests | check-internal self-test | dev services lifecycle test | MTS | high | retain while local-dev wrappers stay | dev-services | direct test | service lifecycle covered |
| `scripts/test-dev-stack-lifecycle.sh` | Target-local test | tests | check-internal self-test | dev stack lifecycle test | MTS | high | retain while dev wrapper stays | dev-stack | direct test | dev lifecycle covered |
| `scripts/test-execution-topology.sh` | Target-local test | tests | check-internal self-test | execution topology test | MTS | high | retain | generated topology tools | direct test; json-shape | topology coverage stable |
| `scripts/test-fallow-static.sh` | Target-local test | tests | check-internal self-test | Fallow static wrapper test | MTS | medium | retain if Fallow wrapper retained | run-fallow-static | direct test | fallow summary covered |
| `scripts/test-frontend-evidence-audit.sh` | Target-local test | tests | active harness smoke check | frontend evidence audit test | MTS | medium | retained as `harness-smoke-frontend-evidence-audit` in extended/full tiers | frontend evidence audit | direct test; `make run-harness-smoke-extended` | active owner test |
| `scripts/test-frontend-import-boundaries.sh` | Target-local test | tests | check-internal self-test | frontend boundary checker test | MTS | high | retain | import boundary checker | direct test; public target | boundary coverage stable |
| `scripts/test-generate-drift.sh` | Target-local test | tests | check-internal self-test | generate drift test | MTS | high | retain | generated/OTel tools | direct test; drift target | drift coverage stable |
| `scripts/test-generated-artifact-policy.sh` | Target-local test | tests | check-internal self-test | generated artifact policy test | MTS | high | retain | generated-artifact policy | direct test; policy check | generated policy covered |
| `scripts/test-go-test-duration-baselines.sh` | Target-local test | tests | check-internal self-test | Go duration baseline test | MTS | medium | retain | duration helpers | direct test | duration coverage stable |
| `scripts/test-harness-contracts.mjs` | Target-local test | tests | check-internal self-test | harness contract test | MTS | high | retain | core harness contract | `make harness-contract` | contract coverage stable |
| `scripts/test-harness-smoke-duration-baselines.sh` | Target-local test | tests | check-internal self-test | harness smoke duration test | MTS | medium | retain | harness-smoke-durations | direct test | smoke duration covered |
| `scripts/test-json-shapes.sh` | Target-local test | tests | check-internal self-test | JSON shape test | MTS | high | retain | generated JSON shape checker | direct test; json-shape | shape coverage stable |
| `scripts/test-lint-shell.sh` | Target-local test | tests | check-internal self-test | shell lint wrapper test | MTS | medium | retain | run-shellcheck | direct test; lint-shell | shell lint covered |
| `scripts/test-make-node-tools.sh` | Target-local test | tests | check-internal self-test | make-node tool test | MTS | medium | retain | make-node tool wrappers | direct test | wrapper covered |
| `scripts/test-print-target-plan.sh` | Target-local test | tests | check-internal self-test | target/phase plan test | MTS | high | retain | planning wrappers | direct test | planning output covered |
| `scripts/test-public-make-wrapper-smoke.sh` | Target-local test | tests | check-internal self-test | public wrapper smoke test | MTS | high | retain | public Make wrappers | direct test; harness-contract | public wrapper coverage stable |
| `scripts/test-release-task-surface.sh` | Target-local test | tests | check-internal self-test | release task-surface test | MTS | high | retain | release evidence/task surface | direct test | release surface covered |
| `scripts/test-run-frontend-unit.sh` | Target-local test | tests | check-internal self-test | frontend unit wrapper test | MTS | high | retain | run-frontend-unit | direct test | frontend unit coverage stable |
| `scripts/test-run-go-gosec-audit.sh` | Target-local test | tests | check-internal self-test | gosec audit wrapper test | MTS | high | retain | run-go-gosec-audit | direct test | audit coverage stable |
| `scripts/test-run-go-gosec-targeted.sh` | Target-local test | tests | check-internal self-test | gosec targeted wrapper test | MTS | high | retain | run-go-gosec-targeted | direct test | targeted scan covered |
| `scripts/test-run-go-govulncheck.sh` | Target-local test | tests | check-internal self-test | govulncheck wrapper test | MTS | high | retain | run-go-govulncheck | direct test | vuln scan covered |
| `scripts/test-run-go-staticcheck.sh` | Target-local test | tests | check-internal self-test | staticcheck wrapper test | MTS | medium | retain | run-go-staticcheck | direct test | staticcheck covered |
| `scripts/test-run-go-target.sh` | Target-local test | tests | check-internal self-test | Go target runner test | MTS | high | retain | run-go-target/backend helpers | direct test | Go runner covered |
| `scripts/test-run-make-sequence-fast.sh` | Target-local test | tests | check-internal self-test | fast Make sequence test | MTS | high | retain | run-make-sequence | direct test | sequence covered |
| `scripts/test-run-make-sequence.sh` | Target-local test | tests | check-internal self-test | Make sequence test | MTS | high | retain | run-make-sequence | direct test | sequence covered |
| `scripts/test-run-phase-slice.sh` | Target-local test | tests | check-internal self-test | phase-slice test | MTS | high | retain | run-phase-slice | direct test | slice coverage stable |
| `scripts/test-run-phase.sh` | Target-local test | tests | check-internal self-test | run-phase core test | MTS | high | retain | tools/harness/core/run-phase | direct test | run-phase covered |
| `scripts/test-run-playwright-manifest-phase.sh` | Target-local test | tests | check-internal self-test | Playwright manifest phase test | MTS | high | retain | browser owner helpers | direct test | browser phase covered |
| `scripts/test-run-playwright-phase.sh` | Target-local test | tests | check-internal self-test | Playwright phase test | MTS | high | retain | browser owner helpers | direct test | browser phase covered |
| `scripts/test-run-playwright-webserver-batch.sh` | Target-local test | tests | check-internal self-test | webserver batch test | MTS | high | retain | browser batch helper | direct test | batch coverage stable |
| `scripts/test-run-vitest-manifest-phase.sh` | Target-local test | tests | check-internal self-test | Vitest manifest phase test | MTS | high | retain | frontend owner helpers | direct test | frontend phase covered |
| `scripts/test-run-vitest-phase.sh` | Target-local test | tests | check-internal self-test | Vitest phase test | MTS | high | retain | frontend owner helpers | direct test | frontend phase covered |
| `scripts/test-sbom-license-evidence.mjs` | Target-local test | tests | check-internal self-test | SBOM/license evidence test | MTS | high | retain | tools/release-evidence | direct test; release targets | release evidence covered |
| `scripts/test-seaweedfs-release-evidence.mjs` | Target-local test | tests | check-internal self-test | SeaweedFS release evidence test | MTS | high | retain | tools/release-evidence | direct test; release gate | release evidence covered |
| `scripts/test-service-backed-make-target-duration-baselines.sh` | Target-local test | tests | check-internal self-test | service-backed duration baseline test | MTS | medium | retain | service-backed durations | direct test | duration coverage stable |
| `scripts/test-service-backed-scheduler.sh` | Target-local test | tests | check-internal self-test | service-backed scheduler test | MTS | high | retain | scheduler/service-backed tools | direct test; scheduler gate | scheduler coverage stable |
| `scripts/test-task-guidance.mjs` | Target-local test | tests | check-internal self-test | task guidance test | MTS | medium | retain | task guidance owner | direct test | guidance coverage stable |
| `scripts/test-task-surface-report.sh` | Target-local test | tests | check-internal self-test | task-surface report test | MTS | medium | retain | task-surface report | direct test | report coverage stable |
| `scripts/test-tool-output-real-targets.sh` | Target-local test | tests | check-internal self-test | tool output real-target test | MTS | high | retain | tool output/core | direct test | output coverage stable |
| `scripts/test-web-e2e-lifecycle.sh` | Target-local test | tests | check-internal self-test | web E2E lifecycle test | MTS | high | retain | browser lifecycle/reset | direct test | lifecycle coverage stable |
| `scripts/update-go-test-durations.mjs` | Stable Make/public wrapper | implementation | duration update input | Go duration baseline updater | MTS | medium | retain wrapper; backend/scheduler logic owner path | duration drift, artifacts | duration update target | update behavior stable |
| `scripts/write-frontend-accessibility-summary.mjs` | Harness frontend implementation | implementation | browser helper input | accessibility summary writer | MTS | high | move reusable summary logic to frontend/browser owner; wrapper optional | harness contract schema | browser a11y targets | summary schema stable |

## Remaining grab-bag findings

| Finding | Remediation | Owner area | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Root `scripts/` still holds reusable implementation logic in several large files. | Move reusable logic behind owner-specific `tools/harness/*` modules while keeping Make-owned wrappers stable. | implementation | Large root scripts such as `check-font-bundle.mjs`, `check-postgres-fixture-budget.mjs`, `cache-artifact.sh`, and scheduler wrappers hide owner logic. | Future changes can target the owning subsystem instead of a grab-bag directory. | Wrapper path compatibility may be needed while Make/task-surface still names root paths. | Refactors keep accumulating local helper APIs in `scripts/`. | Owner tests plus `make lint-scripts`/`make lint-shell`. | Per-file characterization and direct tests. | Reusable logic is owner-path based or explicitly retained. |
| Task-surface checkers were grep-heavy and not direct Make/task-surface entries. | Deleted `check-backend-task-surface.sh`, `check-frontend-task-surface.sh`, and `check-browser-e2e-task-surface.sh`; replacement authority is `harness-contract`, `json-shape-check`, task-surface report tests, execution-topology tests, and browser batch manifest tests. | tests/documentation | These scripts asserted implementation text without owning public behavior. | Reduces brittle test coupling. | No public contract changed; the only active stale classification reference was removed from SeaweedFS occurrence metadata. | If replacement gates regress, task-surface drift could escape. | `make harness-contract`, `make json-shape-check`, task-surface/report and execution-topology tests. | Generated manifest/schema authority. | Closed: deleted with named replacement coverage. |
| Some manual or indirect tests were not active task-surface rows. | Added `test-agent-finalize.sh` and `test-frontend-evidence-audit.sh` as extended/full harness smoke checks; deleted duplicate `test-run-go-target-fast.sh` because `test-run-go-target.sh` remains the active owner test. | tests | Tests outside public inventory can silently rot. | Test inventory matches real gates. | No public target names or command IDs added; generated manifests changed through topology owner input. | Hidden coverage assumptions if active harness smoke is not run. | Direct tests, `make harness-contract`, generated checks. | Owner decision for each test. | Closed: every named test is active or removed with replacement coverage. |
| Readiness/cache/build logic remains partly implemented as root shell. | Moved the cache engine to `tools/harness/readiness/cache-artifact.sh`; retained `scripts/cache-artifact.sh` as a stable wrapper. | implementation/tests | Cache and readiness contracts are harness-owned, not generic script behavior. | Better cache maintenance and narrower validation. | Cache schema IDs, key material, miss/hit behavior, and artifact paths are preserved. | Cache misses/hits become untrustworthy. | `test-cache-artifact`, readiness/build targets, `toolchain-drift`. | Harness cache requirements. | Partially closed for cache; other readiness/build helpers remain future WS-01 work. |
| Browser lifecycle wrappers combine public target behavior with reusable lifecycle details. | Moved reset implementation to `tools/harness/browser/reset-web-e2e-stack.sh`; retained the root reset wrapper. | multiple | Browser stack behavior has artifact and cleanup contracts. | Clearer browser lifecycle ownership. | Reset artifacts, state directories, status/data files, taint markers, and schema validation are preserved. | Browser failures become hard to diagnose or cleanup. | Browser lifecycle tests and relevant browser targets. | Test-route/reset characterization. | Partially closed for reset; broader browser lifecycle helpers remain future WS-02 work. |
| Active docs still mentioned retired `scripts/lib` paths. | Updated active dev guide and fallow handoff references to `tools/harness/browser/browser-shard-plan.mjs` and `tools/harness/core/make-node-tools.mjs`. | documentation | Current guidance should not point at removed helper locations. | Reduces rediscovery and bad follow-up plans. | Docs-only. | Agents follow stale paths. | `make lint-markdown`. | Documentation ownership check. | Closed: active non-archive hits now appear only in this tracker as historical/current-scan text. |
| Release/Core 05 tests remain in `scripts/` while implementation lives under `tools/release-evidence`. | Keep target-local tests, but do not move implementation back into `scripts/`. | tests/multiple | Release evidence is not Base Profile runtime behavior. | Maintains publication boundary clarity. | Test path compatibility only. | Release evidence mistaken for product conformance. | Release evidence direct tests and release targets. | Core 05 boundary. | Tests retained or moved to explicit owner test path with unchanged coverage. |
| Some retained wrappers are intentionally large public orchestration surfaces. | Do not force-move solely because of line count; characterize public behavior first. | multiple | Public scheduler/finalizer/browser wrappers own observable entrypoint behavior. | Avoids churny facade moves. | Wrapper compatibility must remain. | Movement breaks public summaries or failure mapping. | Harness/scheduler/browser gates. | Contract freeze map. | Large wrapper either slimmed safely or documented as stable public entrypoint. |

## Candidate deletion/simplification opportunities

| Candidate | Remediation | Owner area | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `scripts/.gitkeep` | Delete if no directory-preservation need remains. | documentation | `scripts/` is not empty. | Removes non-behavior noise. | None. | Very low; only inventory noise. | `make lint-markdown` if docs mention it. | none | File removed or marked out-of-behavior. |
| `scripts/test-run-go-target-fast.sh` | Deleted; active replacement coverage remains `scripts/test-run-go-target.sh` through `harness-smoke-run-go-target`. | tests | It duplicated active Go target characterization without a caller. | Prevents stale duplicate tests. | No public contract. | Hidden coverage assumptions if the full test loses assertions. | `scripts/test-run-go-target.sh`; `make harness-contract`. | Go target owner test. | Closed: removed. |
| `scripts/test-agent-finalize.sh` | Retained as `harness-smoke-agent-finalize` in extended/full harness smoke tiers. | tests | Manual-only finalizer coverage can rot. | Finalizer coverage is visible. | Internal harness smoke inventory only; no public target added. | Finalizer behavior changes without active test. | direct test plus extended/full harness smoke. | retained run fixture availability. | Closed: active owner check. |
| `scripts/test-frontend-evidence-audit.sh` | Retained as `harness-smoke-frontend-evidence-audit` in extended/full harness smoke tiers. | tests | Manual-only frontend evidence test can drift. | Visible frontend evidence coverage. | Internal harness smoke inventory only; no public target added. | Audit behavior lacks active characterization. | direct test plus extended/full harness smoke. | frontend phase fixtures. | Closed: active owner check. |
| `scripts/check-*-task-surface.sh` | Deleted after mapping replacement authority to generated manifests, task-surface report tests, execution-topology tests, browser batch manifest tests, `json-shape-check`, and `harness-contract`. | tests/documentation | They asserted implementation text rather than stable behavior. | Easier owner refactors. | No public contract; stale SeaweedFS occurrence metadata reference removed. | Coverage gaps if replacement gates are weakened. | task-surface tests, `json-shape-check`, `harness-contract`. | generated/task-surface owner gates. | Closed: removed with replacement authority. |
| `scripts/check-phase-map.mjs` | Inline through `tools/harness/planning` CLI or retain as tiny compatibility wrapper. | implementation | It is a small duplicate wrapper. | Fewer root helper paths. | Test references may need update. | Low duplicate surface. | phase-map tests. | phase-manifest CLI. | Wrapper justified or removed. |
| `scripts/harness-contract.mjs` | Kept as a compatibility trampoline; CLI implementation moved to `tools/harness/core/harness-contract-cli.mjs`, and `harness-contract.sh` invokes the core CLI directly. | implementation | The CLI is core harness logic. | Cleaner core ownership. | Direct `.mjs` callers remain compatible for one iteration. | Core contract logic remains in root scripts if trampoline becomes permanent. | `make harness-contract`; public wrapper smoke. | core harness owner. | Closed for implementation move; future slice may remove trampoline after callers migrate. |
| `scripts/print-go-shard-plan.mjs` and `scripts/check-go-target-plan-coverage.mjs` | Move to backend owner CLI or retain compatibility wrappers. | implementation | They are backend harness diagnostics. | Backend harness support is discoverable. | Test paths may need wrapper compatibility. | Backend planning logic remains split. | Go target tests. | backend target plan. | Owner path or retained rationale. |
| `scripts/diagnose-inotify.mjs` | Move to readiness owner path and keep doctor/dev wrapper compatibility. | implementation | It is local readiness diagnostics, not a public scripts module. | Readiness concerns co-located. | `check-doctor.sh` and `dev-stack.sh` references must stay valid. | Inotify readiness remains hidden. | `make doctor`; dev-stack tests. | local-dev wrappers. | owner path plus stable behavior. |
| `scripts/run-browser-e2e-owned-stack.sh` | Fold into browser owner helper if stateful wrapper no longer needs a root script path. | implementation | It is indirect browser lifecycle glue. | Fewer browser root wrappers. | Stateful browser target must remain stable. | Extra wrapper obscures lifecycle owner. | browser lifecycle/stateful tests. | browser owned-stack helper. | removed or retained with clear value. |

## Workstream matrix

| ID | Remediation | Owner area | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria | Dependencies | Exit criteria |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WS-00 Tracker rebaseline | Maintain this Iteration 2 tracker from live `find`, `rg`, Make, task-surface, and topology scans. | documentation | Prevents stale completed work from masquerading as current guidance. | Future slices start from current truth. | Tracker-only; no runtime impact. | Obsolete counts and stale rows drive bad moves. | Post-edit `make lint-markdown`. | Current repo state. | All required sections present and every file classified. |
| WS-01 Readiness/cache/build helpers | Cache slice closed: `scripts/cache-artifact.sh` is now a wrapper over `tools/harness/readiness/cache-artifact.sh`; future slices still cover `list-build-inputs.sh`, bootstrap/build wrappers, `diagnose-inotify.mjs`, and install/build readiness scripts. | implementation/tests | Reduces root `scripts/` implementation logic while preserving cache/build contracts. | Cache and readiness behavior becomes owner-maintainable. | Preserve cache keys, schema IDs, readiness stamps, and build outputs. | False cache hits/misses or build-input drift. | `test-cache-artifact`, `test-build-input-discovery`, `toolchain-drift`, affected build/readiness targets. | Direct characterization and cache requirements. | Cache wrapper thin; remaining readiness wrappers explicitly future work. |
| WS-02 Frontend/browser wrappers | Reset slice closed: `scripts/reset-web-e2e-stack.sh` is now a wrapper over `tools/harness/browser/reset-web-e2e-stack.sh`; future slices still cover font, design-token, start/session, accessibility summary, and frontend evidence helpers. | implementation/tests | Keeps frontend readiness and browser lifecycle logic owner-aligned. | Browser/frontend changes validate through their owners. | Preserve design-token provenance and browser reset artifacts unless owner inputs change. | Browser artifacts, reset behavior, or boundary checks drift. | `frontend-import-boundary-check`, `frontend-unit`, browser tests, `json-shape-check` if generated references move. | Frontend/browser characterization. | Reset wrapper thin; remaining frontend/browser logic explicitly future work. |
| WS-03 Harness scheduler/public wrappers | Harness-contract CLI slice closed: implementation moved to `tools/harness/core/harness-contract-cli.mjs`; `scripts/harness-contract.sh` remains stable and `scripts/harness-contract.mjs` is a compatibility trampoline. Future slices still cover scheduler/finalizer/explain wrappers. | multiple | Separates stable command entrypoints from deep harness mechanics. | Public wrappers stay small and intentional. | Preserve summaries, scheduler events, timing artifacts, exit codes, and failure classes. | Public harness accounting drift. | `harness-contract`, `check-harness-smoke`, scheduler self-tests, targeted duration checks. | Contract freeze and retained-run fixtures. | Harness-contract boundary closed; remaining wrapper boundaries documented per file. |
| WS-04 Backend/static/migration residuals | Classify and slim backend helper wrappers such as fixture budget, duration baseline, migration/schema ownership, Go target, and module-boundary checks. | implementation/tests | Makes backend harness support discoverable under `tools/harness/backend`. | Backend support evolves with backend harness owner. | Preserve generated-root filtering and Go target selection. | Missed generated files, wrong package selection, fixture-budget regression. | `backend-module-boundary-check`, duration coverage, `migration-drift`, `json-shape-check`, direct tests. | Backend owner helpers and manifests. | Backend reusable logic has owner paths or explicit retention rationale. |
| WS-05 Task-surface checker rationalization | Closed for the three unreferenced grep-heavy checkers: deleted root scripts and recorded generated/task-surface replacement authority. | tests/documentation | Avoids preserving grep-heavy assertions only because they exist. | Task-surface coverage becomes less brittle. | No public contract changed. | Coverage disappears if replacement gates are weakened. | `harness-contract`, `json-shape-check`, task-surface tests. | Coverage map to current generated checks. | Closed: deleted with replacement authority. |
| WS-06 Target-local tests | Closed for the three named OQ tests: `test-agent-finalize.sh` and `test-frontend-evidence-audit.sh` are active extended/full harness smoke checks; `test-run-go-target-fast.sh` was deleted with active `test-run-go-target.sh` replacement coverage. | tests | Prevents test fixture helpers from becoming implicit production APIs. | Test inventory matches active gates. | Internal harness smoke manifest changed through topology owner input. | Silent coverage loss or stale tests. | `harness-contract`, `lint-scripts`, `lint-shell`, selected direct tests. | Owner decision for unreferenced tests. | All 52 current tests classified as retained, moved, or deletion candidates with evidence. |
| WS-07 Local-dev/CI/operator wrappers | Retain or simplify `scripts/ci/*`, `dev-services.sh`, `dev-stack.sh`, `check-doctor.sh`, deployable/standup smoke wrappers, and related operator entrypoints. | multiple | Retention should reflect external/operator value, not historical placement. | Local-dev and CI surfaces stay intentional. | Preserve public Make behavior and provider-neutral CI entry where useful. | Local dev, recovery smoke, or external CI invocation breaks. | `doctor`, `deployable-shape`, relevant standup targets, `lint-shell`. | Operator use and source limits. | Wrappers have clear continuing value or deletion criteria. |
| WS-08 Documentation cleanup | Closed for active non-archive hits: dev guide and fallow-static handoff now point at owner paths; archive/recovery docs remain historical. | documentation | Prevents rediscovery from stale owner paths. | Agents and maintainers follow current owner paths. | Docs-only. | Active guidance points to deleted paths. | `make lint-markdown`. | Documentation owner review. | Closed: current active docs no longer advertise retired helper locations. |

## Slice admission rules

1. One slice may touch only one owner family unless this tracker marks the dependency as inseparable.
2. No slice may create `internal/modules/scripts`, `packages/scripts`, a product-facing `scripts` module, or any equivalent abstraction.
3. Reusable logic moves to existing owner paths; root `scripts/` keeps only public/internal Make wrappers, operator/local-dev/CI wrappers, and target-local tests.
4. Generated outputs must be changed only by owner generators; tracker-only edits must not touch generated files.
5. A slice must freeze observable behavior before edits: Make target, `command_id`, schema IDs, artifact paths, env inputs, exit codes, cleanup behavior, and failure taxonomy.
6. Deletion is allowed only when current Make/task-surface/topology/import/current-doc scans show no active owner value, or when replacement coverage is named.
7. A slice that changes docs only must still run the narrow docs validation target and record the retained run root.
8. A slice that moves a wrapper named by Make/task-surface/topology must update owner inputs and generated outputs through the relevant Make generator, never by hand.
9. A slice that changes release/Core 05 evidence tooling must explicitly state whether claim-publication intent is `none`, `informative_engineering_measurement`, or `claim_bearing_publication`.
10. A slice that changes browser/reset/local-dev wrappers must state cleanup and destructive-operation predicates before editing.

## Validation plan

| Change class | Required validation | Broaden when |
| --- | --- | --- |
| Tracker-only | `make lint-markdown` | Never by default; this iteration is docs-only. |
| Active-doc stale reference cleanup | `make lint-markdown` | Run `json-shape-check` only if docs own schema examples. |
| Shell wrapper movement | `bash -n` for touched scripts, then `make lint-shell` or the owner target. | Wrapper is public, destructive, readiness, or CI-facing. |
| Node wrapper movement | `node --check` for touched modules, then direct owner test or Make target. | Public output, artifact, scheduler, or schema behavior changes. |
| Harness core/scheduler movement | `make harness-contract`; `make check-harness-smoke`; selected scheduler tests. | Public target summaries, failure taxonomy, or scheduler artifacts can drift. |
| Generated/task-surface movement | `make generated-artifact-policy-check`; `make json-shape-check`; relevant drift target. | Any generated output or owner input changes. |
| Frontend/browser movement | `make frontend-import-boundary-check`; `make frontend-unit`; relevant browser target/test. | Browser lifecycle, reset, visual, or accessibility artifacts change. |
| Backend/static movement | `make backend-module-boundary-check`; duration/migration/schema owner checks. | Go package selection, fixture budgets, or generated-root filtering changes. |
| Release evidence movement | Direct release evidence tests plus affected release target. | SBOM, license, benchmark, or object-store evidence changes. |

Validation result for this remediation iteration:

| Command | Result | Run root | Artifacts | Notes |
| --- | --- | --- | --- | --- |
| `make lint-markdown` | PASS before edit | `.cartulary/test-results/20260703T044945Z-p430930` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Baseline observed before replacing the completed tracker. |
| `CARTULARY_TEST_RUN_ID=20260703T045900Z-ptracker2 make lint-markdown` | PASS after edit | `.cartulary/test-results/20260703T045900Z-ptracker2` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Tracker-only validation for Iteration 2. |
| `CARTULARY_TEST_RUN_ID=20260703T045900Z-ptracker2 make lint-markdown` | FAIL after wording patch | `.cartulary/test-results/20260703T045900Z-ptracker2` | existing run root | Harness configuration failure because the fixed run root was already non-empty; unrelated to tracker content. |
| `CARTULARY_TEST_RUN_ID=20260703T050120Z-ptracker2-final make lint-markdown` | PASS final | `.cartulary/test-results/20260703T050120Z-ptracker2-final` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Final validation for the checked-in tracker state. |
| `CARTULARY_TEST_RUN_ID=20260703T-remediation-phase-schedules make phase-schedules` | PASS | `.cartulary/test-results/20260703T-remediation-phase-schedules` | `phase-schedules/tool-run-summary.json` | Regenerated downstream task-surface/topology outputs from owner input. |
| `bash scripts/test-cache-artifact.sh` | PASS | none | none | Direct cache wrapper/owner implementation coverage. |
| `bash scripts/test-agent-finalize.sh` | PASS | none | none | Direct finalizer test promoted to active harness smoke. |
| `bash scripts/test-frontend-evidence-audit.sh` | PASS | none | none | Direct frontend evidence audit test promoted to active harness smoke. |
| `bash scripts/test-web-e2e-lifecycle.sh` | PASS | none | none | Direct browser lifecycle/reset characterization. |
| `bash scripts/test-run-go-target.sh` | PASS | none | none | Replacement owner coverage for deleted fast Go target test. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-json-shape make json-shape-check` | PASS | `.cartulary/test-results/20260703Tremediation-json-shape` | `json-shape-check/tool-run-summary.json` | Generated/task-surface schema validation. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-generated-policy make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260703Tremediation-generated-policy` | `generated-artifact-policy-check/tool-run-summary.json` | Generated artifact policy validation. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-phase-schedule-drift make phase-schedule-drift` | PASS | `.cartulary/test-results/20260703Tremediation-phase-schedule-drift` | `phase-schedule-drift/tool-run-summary.json` | Confirms generated topology outputs are current. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-lint-md make lint-markdown` | PASS | `.cartulary/test-results/20260703Tremediation-lint-md` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Documentation validation after active-doc cleanup and tracker updates. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-harness-contract make harness-contract` | PASS | `.cartulary/test-results/20260703Tremediation-harness-contract` | `harness-contract/target-summary.json`, `harness-contract/tool-run-summary.json` | Harness contract and task-surface manifest validation. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-lint-scripts make lint-scripts` | PASS | `.cartulary/test-results/20260703Tremediation-lint-scripts` | `lint-scripts/target-summary.json`, `lint-scripts/tool-run-summary.json` | Node/script lint validation. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-lint-shell make lint-shell` | PASS | `.cartulary/test-results/20260703Tremediation-lint-shell` | `lint-shell/target-summary.json`, `lint-shell/tool-run-summary.json` | Shell wrapper validation. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-task-surface-report make task-surface-report TASK_SURFACE_REPORT_ARGS='--check --all'` | PASS | none | no retained artifact emitted by helper-only report | Confirms task-surface report and logical harness checks, including the two newly active checks. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-final-lint-md make lint-markdown` | PASS | `.cartulary/test-results/20260703Tremediation-final-lint-md` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Final Markdown validation after tracker closure rows were updated. |
| `CARTULARY_TEST_RUN_ID=20260703Tremediation-final2-lint-md make lint-markdown` | PASS | `.cartulary/test-results/20260703Tremediation-final2-lint-md` | `lint-markdown/target-summary.json`, `lint-markdown/tool-run-summary.json` | Final Markdown validation after recording the previous final lint row. |
| `make agent-finalize` | SKIPPED | none | none | No retained full warm `RESULTS_DIR` was supplied; direct finalizer tests and harness-contract validation were run instead. |

## Handoff log template

Copy this row for each future slice:

| Time | Agent/session | Branch/commit | Inventory counts | Files changed | Sections updated | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `YYYY-MM-DDTHH:MM:SSZ` | TODO | TODO | `scripts=<n>; scripts/lib=<n>; tests=<n>` | TODO | TODO | TODO | TODO | TODO | TODO |

Current Iteration 2 handoff row:

| Time | Agent/session | Branch/commit | Inventory counts | Files changed | Sections updated | Commands/results/run roots | Skipped checks and reason | Risks/blockers | Next slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-03T05:01:20Z | Codex Iteration 2 tracker | `main` / `3c2e9c56` | `scripts=150; scripts/lib=0; scripts/ci=4; tests=53` | `docs/handoffs/scripts-module-refactor-tracker.md` | all required Iteration 2 sections | `make lint-markdown` baseline passed at `.cartulary/test-results/20260703T044945Z-p430930`; post-edit pass at `.cartulary/test-results/20260703T045900Z-ptracker2`; duplicate-ID rerun failed as expected because the run root was non-empty; final pass at `.cartulary/test-results/20260703T050120Z-ptracker2-final` | `make agent-finalize` skipped because this is tracker-only and no retained full warm `RESULTS_DIR` was supplied | Future movement still requires slice-local characterization | Start WS-01 or WS-05; avoid broad scripts cleanup |
| 2026-07-03T05:44:28Z | Codex remediation | `main` / `3c2e9c56` | `scripts=146; scripts/lib=0; scripts/ci=4; tests=52` | docs/spec/tracker updates; wrappers for cache, harness-contract, reset; topology owner input and generated task-surface outputs; deleted unreferenced checkers and duplicate fast Go test | OQ-01 through OQ-07 closure; WS-01, WS-02, WS-03, WS-05, WS-06, WS-08 status; validation table | Direct owner tests passed; `phase-schedules`, `json-shape-check`, `generated-artifact-policy-check`, `phase-schedule-drift`, `lint-markdown`, `harness-contract`, `lint-scripts`, `lint-shell`, and task-surface report passed at run roots listed above | `make agent-finalize` skipped because no retained full warm `RESULTS_DIR` was supplied | `scripts/harness-contract.mjs` remains a compatibility trampoline; `start-web-e2e.sh` remains a source-compatible public browser lifecycle wrapper | Next slice should target remaining readiness/frontend/backend helper wrappers one owner family at a time |

## Open questions and blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| OQ-01 | Are the three unreferenced tests intended to be active coverage? | Manual tests can rot or hide coverage assumptions. | Owner decision for `test-agent-finalize.sh`, `test-frontend-evidence-audit.sh`, and `test-run-go-target-fast.sh`. | Closed: two tests are active extended/full harness smoke checks; duplicate fast Go test deleted with `test-run-go-target.sh` coverage. |
| OQ-02 | Should task-surface checker scripts stay as grep-heavy tests? | They may block behavior-preserving owner moves due to implementation-string assertions. | Coverage map against generated manifests and harness contract tests. | Closed: deleted and replaced by generated/task-surface/harness-contract authorities. |
| OQ-03 | Should `harness-contract.mjs` remain in `scripts/` or move under core with a stable shell wrapper? | It is core harness CLI implementation but is used as a stable entrypoint. | Harness core characterization and public wrapper smoke evidence. | Closed: core CLI moved to `tools/harness/core/harness-contract-cli.mjs`; shell wrapper stable; `.mjs` path is compatibility trampoline. |
| OQ-04 | Which active docs should be updated for retired `scripts/lib` paths? | Current docs should not direct agents to deleted helper locations. | Documentation owner review of non-archive hits. | Closed: active docs now point to owner paths; remaining non-archive hits are tracker self-reference. |
| OQ-05 | Which local-dev/CI wrappers have external consumers outside this checkout? | `scripts/ci/verify.sh` is not directly referenced by current Make/task-surface/topology, but may be provider entrypoint. | Maintainer or CI configuration evidence. | Closed with source limit: no in-repo provider workflow config found; retain `scripts/ci/verify.sh` until maintainer/external-consumer evidence allows deletion. |
| OQ-06 | Can cache implementation move without changing shell-compatible behavior? | Cache records are schema-bearing and used by readiness/build targets. | Direct cache tests and cache schema freeze. | Closed: implementation moved to `tools/harness/readiness/cache-artifact.sh`; wrapper and cache tests passed. |
| OQ-07 | Can browser reset/session logic move without losing cleanup guarantees? | Browser reset touches test routes, state directories, and retained artifacts. | Browser lifecycle characterization and reset artifact checks. | Closed for reset: implementation moved to `tools/harness/browser/reset-web-e2e-stack.sh` and lifecycle tests passed. Session/start lifecycle remains explicitly retained in `scripts/start-web-e2e.sh` as a source-compatible public browser wrapper. |

## Binary completion criteria

| Criterion | Status |
| --- | --- |
| `scripts/lib/` is empty. | PASS: live traversal finds 0 entries. |
| All 146 live `scripts/` files are classified. | PASS: Section "Per-file classification matrix" contains one row for each live path. |
| Every workstream includes remediation, owner area, rationale, long-term benefit, compatibility impact, unresolved risk, validation, dependencies, and exit criteria. | PASS: Section "Workstream matrix" includes those fields for WS-00 through WS-08. |
| Remaining grab-bag findings include remediation and exit criteria. | PASS: Section "Remaining grab-bag findings" includes required fields. |
| Candidate deletion/simplification opportunities justify continuing value or removal. | PASS: Section "Candidate deletion/simplification opportunities" records evidence and validation gates. |
| Public Make targets, command IDs, schemas, artifact paths, failure taxonomy, generated-output ownership, and cleanup predicates are frozen. | PASS: Section "Public contract and behavior freeze map" records the freeze surfaces. |
| No product-facing `scripts` module is introduced. | PASS: Tracker states this prohibition and proposes only owner-specific `tools/*` destinations. |
| Generated outputs are not hand-edited. | PASS: generated task-surface outputs changed through `make phase-schedules`. |
| Post-edit validation is recorded with command result and run root. | PASS: remediation validation table records direct tests, generated gates, lint gates, and retained run roots. |
| Skipped checks are recorded with reasons. | PASS: broad retained-run `make agent-finalize` skipped because no retained full warm `RESULTS_DIR` was supplied. |
