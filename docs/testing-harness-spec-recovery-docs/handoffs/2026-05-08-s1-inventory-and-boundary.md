---
doc_id: THR-S1-HANDOFF-2026-05-08
title: Testing Harness Recovery S1 Inventory and Boundary Handoff
status: active
role: handoff
---

# Testing Harness Recovery S1 Inventory and Boundary Handoff

## Session metadata

| Field | Value |
|---|---|
| Handoff ID | `2026-05-08-s1-inventory-and-boundary` |
| Session date | `2026-05-08` |
| Agent/session identifier | `Codex S1 inventory implementation` |
| Repository revision | `7b8ca029684ed70867325a5066ef2a1714f25f4d` |
| Branch | `main...origin/main` |
| Dirty state at start | Clean; `git status --short --branch` printed only `## main...origin/main` |
| Dirty state at end | Only recovery docs changed; final `git status --short` listed modified/new files under `docs/testing-harness-spec-recovery-docs/**` only |
| Runtime platform | `Linux DeskRip 6.6.87.2-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Thu Jun 5 18:30:46 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| Current sprint | `S1: Inventory and boundary` |
| Sprint status after session | `complete` if final git audit shows only recovery docs changed |
| Recovery-doc paths changed | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md`, `ambiguity-register.md`, `source-limit-log.md`, `03-sprint-plan.md`, this handoff |
| Implementation files changed | Must be none; final audit required |

## Work completed this session

- Created the S1 harness inventory with boundary classifications.
- Created the uninvoked/orphan-like surface list.
- Created the embedded harness logic list.
- Created the ambiguity register.
- Updated the source-limit log with S1 source limits.
- Updated S1 sprint plan fields for status, blockers, concerns, and handoff.

## Files and surfaces inspected

| Surface | Why inspected | Evidence status | Notes |
|---|---|---|---|
| `docs/testing-harness-spec-recovery-docs/recovery-charter.md` | S0 boundary and permitted writes | `observed` | Confirmed writes must stay under recovery docs. |
| `docs/testing-harness-spec-recovery-docs/01-recovery-process.md` | Stage 1 expectations | `observed` | S1 outputs match process templates. |
| `docs/testing-harness-spec-recovery-docs/04-registers-and-checklists.md` | Inventory/source-limit/ambiguity table shapes | `observed` | Used existing labels and register fields. |
| `docs/domain.md` and Core spec file list | Authority boundary | `observed` | Used only for owner boundaries, not behavior recovery. |
| `Makefile`, `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json` | Main task surface | `observed/runtime_observed` | Task-surface report found 77 public, 17 check-internal, 28 helper-only targets. |
| `package.json`, `apps/web/package.json`, `packages/*/package.json` | pnpm entrypoints | `observed` | Root exposes build/test/typecheck; `apps/web` exposes Vite/Vitest/Playwright aliases. |
| `apps/web/playwright*.config.ts`, `apps/web/vite.config.ts` | Browser/frontend harness config | `observed` | Recorded Playwright global setup/teardown, traces, worker defaults, Vitest projects. |
| `scripts/**` | Shell/Node harness orchestration and self-tests | `observed` | `find scripts` listed orchestration, scheduler, reporting, cleanup, smoke, and self-test surfaces. |
| `tools/*.json`, `tools/testservices/**`, `tools/gotest*` | Manifests, schedulers, service wrapper | `observed` | Active phase maps are phase0 through phase6; phase7/8 are planned in registry but files absent. |
| `internal/testutil/**` | Go harness utilities | `observed` | Recorded pgtest, s3test, testcontainersx, suiteservices, httptestx, process, WebSocket, and phase harnesses. |
| `cmd/**/*_test.go`, `internal/**/*_test.go`, `tools/**/*_test.go` | Product tests with embedded harness logic | `observed/runtime_observed` | 91 Go test files counted. |
| `apps/web/e2e/**`, `apps/web/src/**/*.test.*`, `packages/*/src/**/*.test.*` | Browser/frontend/product tests and support | `observed` | Browser specs, Vitest harness-node tests, visual snapshots, and frontend setup recorded. |
| `internal/testutil/fixtures/**`, `internal/testutil/golden/**`, visual snapshots | Fixtures/goldens/snapshots | `observed` | Authority/update rules deferred to S3. |
| `.gitignore`, ignored artifact paths, `.cartulary/test-results/**` | Runtime/generated/temp/log classification | `observed` | `.cartulary/test-results` is ignored by unanchored `test-results/`. |
| `.github/**` | CI workflow source | `source_limit` | Directory absent. |

## Commands run

| Command | Purpose | Result | Exit code | Artifacts produced | Notes |
|---|---|---|---:|---|---|
| `git status --short --branch` | Record dirty state | Clean at start | 0 | none | Printed `## main...origin/main`. |
| `git rev-parse HEAD` | Record revision | Printed `7b8ca029684ed70867325a5066ef2a1714f25f4d` | 0 | none | Metadata evidence. |
| `uname -a` | Record platform | Printed WSL2 Linux platform | 0 | none | Metadata evidence. |
| `date -Is` | Record timestamp | Printed `2026-05-08T19:54:28-04:00` | 0 | none | Metadata evidence. |
| `test -d .github && find .github -maxdepth 3 -type f \| sort \|\| true` | Check CI workflow directory | No files printed | 0 | none | Source limit retained. |
| `make help` | Compact public target surface | Printed compact help | 0 | none | Runtime evidence. |
| `make help-all` | Exhaustive public target surface | Printed public task surface | 0 | none | Runtime evidence. |
| `make task-guide` | Role/phase guidance | Printed role guidance and latest artifact references | 0 | none | Runtime evidence; did not run tests. |
| `make target-plan` | Backend target family summary | Printed backend target summaries | 0 | none | Runtime evidence. |
| `make target-plan-json` | Backend plan shape | Emitted target plan JSON | 0 | none | Output was large; used as row-shape evidence only. |
| `make explain-phase PHASE=phase0` through `phase6` | Phase evidence and target coverage | Printed phase maps for active phases | 0 | none | Runtime evidence; did not run phase tests. |
| `make explain-target TARGET=<target> DETAIL=summary` | Target classifications for major test/check/browser/release targets | Printed target summaries | 0 | none | Targets included `test-fast`, `test`, `check`, `ci`, `release-check`, backend targets, frontend targets, browser targets, `agent-finalize`, `generate`, `phase-slice`, and `service-backed-slice`. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | Public/private/helper classification | Printed task classifications and logical harness checks | 0 | none | Key S1 evidence. |
| `rg --files`, `git ls-files`, `find ...` inventory commands | Enumerate files, tests, fixtures, generated outputs, docs, scripts | Printed inventories | 0 | none | Static evidence. |
| `git check-ignore -v ...`, `git status --ignored --short ...` | Classify ignored temp/report/build paths | Printed ignore matches | 0 | none | `.cartulary/test-results` ignored by `test-results/`; `tmp`, `dist`, and `*.tsbuildinfo` ignored. |
| `jq` manifest inspection commands | Inspect manifest keys, targets, resources, registry | Mostly printed JSON summaries | 0 or 5 for exploratory malformed filters | none | Malformed exploratory filters were retried with corrected jq; no files changed. |

No broad verification commands such as `make test-fast`, `make test`, `make check`, `make ci`, or browser E2E runs were executed.

## Recovery artifacts updated

| Artifact | Update summary | Remaining gaps |
|---|---|---|
| `docs/testing-harness-spec-recovery-docs/harness-inventory.md` | Added S1 inventory and boundary answers. | S2 must trace command contracts; S3/S4 must recover artifact/service lifecycles. |
| `docs/testing-harness-spec-recovery-docs/uninvoked-surface-list.md` | Added uninvoked/orphan-like candidate surfaces. | S2 should confirm `scripts/test-run-go-target-fast.sh` status. |
| `docs/testing-harness-spec-recovery-docs/embedded-harness-logic-list.md` | Added representative embedded harness mechanics. | Later sprints must decide what becomes harness contract. |
| `docs/testing-harness-spec-recovery-docs/ambiguity-register.md` | Added S1 ambiguity and owner-decision rows. | All rows remain open. |
| `docs/testing-harness-spec-recovery-docs/source-limit-log.md` | Added S1 source limits and confirmed `.github/**` absence. | Runtime behavior remains intentionally unobserved. |
| `docs/testing-harness-spec-recovery-docs/03-sprint-plan.md` | Updated S1 status/checklist/handoff fields. | None for S1 if final audit passes. |

## Key findings

| Finding ID | Finding | Evidence status | Evidence reference | Impact |
|---|---|---|---|---|
| FIND-S1-0001 | The harness entrypoint surface is Make-centered and generated from `tools/task_surface_manifest.json` into `tools/task_surface.generated.mk`. | `observed/runtime_observed` | `Makefile`, `tools/task_surface.generated.mk`, `task-surface-report --all` | S2 should start from task-surface report and generated Make surface. |
| FIND-S1-0002 | Active phase coverage is phase0 through phase6; phase7/phase8 are planned in `tools/phase_registry.json` but files are absent. | `observed/source_limit` | `make explain-phase PHASE=phase0..phase6`; `tools/phase_registry.json` | Prevents treating planned phases as active missing coverage. |
| FIND-S1-0003 | Service-backed behavior spans Make, service-backed scheduler, `tools/testservices`, `pgtest`, `s3test`, and testcontainers helpers. | `observed` | `make explain-target`, `tools/testservices/main.go`, `pgtest.go`, `s3test.go` | S4 should own service lifecycle recovery. |
| FIND-S1-0004 | Browser E2E behavior spans Playwright configs, browser batch manifests/scripts, `start-web-e2e.sh`, and shared Playwright state helpers. | `observed` | `apps/web/playwright*.config.ts`, `apps/web/e2e/**`, browser `explain-target` output | S2/S4 must trace browser entrypoints and runtime state. |
| FIND-S1-0005 | Retained `.cartulary/test-results/**` artifacts exist and are ignored by `test-results/`; Make defaults results root to `.cartulary/test-results`. | `observed` | `.gitignore`, `Makefile`, `git check-ignore -v`, `find .cartulary` | S3/S5 must recover artifact schema and authority. |
| FIND-S1-0006 | Test-only runtime reset behavior lives in `internal/app/test_runtime_reset.go`, outside `internal/testutil/**`. | `observed` | `sed -n '1,220p' internal/app/test_runtime_reset.go` | Owner decision needed before NLSpec classification. |

## New or updated ambiguities

| Ambiguity ID | Surface | Decision required | Proposed owner | Blocking sprint |
|---|---|---|---|---|
| `AMB-0001` | `.github/**` versus `scripts/ci/**` | Decide CI authority/location. | Maintainer/repository owner | S2 for CI mapping completeness |
| `AMB-0004` | Planned phase7/phase8 registry entries | Decide whether planned absent phase files belong in recovered harness scope now. | Phase/harness owner | Not blocking S2 |
| `AMB-0006` | `internal/app/test_runtime_reset.go` | Decide whether this test-only route belongs in harness spec. | Maintainer/harness owner | S3/S4/S8 |
| `AMB-0008` | Cleanup owner and idempotency | Decide canonical cleanup behavior after S3 evidence. | Harness owner | S3 |

## New or updated hazards

| Hazard ID | Surface | Trigger | Severity | Next action |
|---|---|---|---|---|
| `HAZ-S4-0001` | Postgres, MinIO, browser stack, process slots, ports, Playwright state | Parallel service-backed/browser execution or stale retained service fixtures | `unknown` | Recover in S4/S6 using scheduler resource registry, service wrapper, and runtime artifacts. |
| `HAZ-S3-0001` | Generated/ignored artifacts and committed generated outputs | Treating generated outputs or retained reports as owner specs | `unknown` | Recover in S3 and authority pass before NLSpec drafting. |

## New or updated failure modes

| Failure ID | Failure class | Trigger | Retryable | Next action |
|---|---|---|---|---|
| `FAIL-S5-0001` | `service_start_error/service_readiness_timeout` | Postgres, MinIO, or browser stack readiness failure | `unknown` | Recover from `testcontainersx`, `pgtest`, `s3test`, `tools/testservices`, and `start-web-e2e.sh`. |
| `FAIL-S5-0002` | `harness_internal_error` | Missing or malformed run summaries, scheduler summaries, phase summaries, or fixture reports | `unknown` | Recover from `scripts/lib/test-output.*`, scheduler runners, and explain-run/fixture-report tools. |

## Source limits and inaccessible material

| Source-limit ID | Surface | Limit | Impact | Follow-up |
|---|---|---|---|---|
| `SL-0001` | `.github/**` | Directory absent. | CI workflow mapping incomplete. | S2 should classify external/absent CI. |
| `SL-0002` | Broad verification commands | `make test-fast`, `make test`, `make check`, `make ci`, and browser E2E were intentionally not run. | Runtime behavior and failure bundles not observed in S1. | S2/S3/S5 may inspect retained artifacts or run targeted commands if authorized. |
| `SL-0003` | Service-backed live runtime | Docker/Postgres/MinIO/browser lifecycle not exercised during S1. | Readiness, cleanup, retry, and resource collision behavior remain static-only. | S4/S6. |
| `SL-0004` | Retained `.cartulary/test-results/**` artifacts | Provenance of existing local retained runs was not validated. | Artifact authority and staleness cannot be inferred from presence. | S3/S5. |
| `SL-0005` | Planned phase7/phase8 files | Registry names planned files that are absent. | Planned phases cannot be inventoried as active files. | Phase owner decision or future phase creation. |

## Decisions made

No maintainer or governing-source decisions were made in this session.

## Pending owner decisions

| Decision prompt | Why it matters | Suggested owner | Blocking effect |
|---|---|---|---|
| Is CI intentionally external to this checkout, absent, or represented only by `scripts/ci/**`? | S2 must map CI entrypoints or record their absence. | Maintainer/repository owner | Blocks complete CI entrypoint mapping. |
| Should `internal/app/test_runtime_reset.go` be part of the recovered harness spec? | It is test-only but lives in product assembly code and mutates runtime state. | Maintainer/harness owner | Blocks reset-route authority classification. |
| Should planned phase7/phase8 registry entries be specified now? | Registry lists absent files. | Phase/harness owner | Not blocking S2; affects later phase coverage recovery. |

## Current blockers

| Blocker | Affected sprint | What is needed to unblock | Owner |
|---|---|---|---|
| `.github/**` absent | S2 CI command mapping | Maintainer decision or external CI source | Maintainer/repository owner |
| Runtime behavior intentionally unobserved | S3/S4/S5 detail recovery | Later targeted runtime inspection or retained artifact analysis | Future recovery agent |

## Suggested next actions

1. In S2, start with `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` and `tools/task_surface_manifest.json`.
2. Trace public, check-internal, helper-only, and package-script commands to first implementation files.
3. Split CI/provider-neutral command mapping between `make ci`, `scripts/ci/**`, and absent `.github/**`.
4. Preserve S1 boundary classifications while adding exact entrypoint inputs, outputs, env vars, ordering, parallel-safety, and failure behavior.

## Do not repeat or redo

- Do not rerun broad verification gates just to inventory paths.
- Do not recreate the S1 static file lists unless files changed.
- Do not hand-edit generated outputs or runtime artifacts.
- Do not collapse product assertions into harness contract during S2.

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Test logic modified | `no` |
| CI behavior modified | `no` |
| Fixture contents modified | `no` |
| Cleanup scripts modified | `no` |
| Only recovery docs changed | `yes` |

## Final handoff summary

S1 established the inventory and boundary evidence needed for S2. The harness is
Make/task-surface centered, with substantial behavior in generated manifests,
scheduler scripts, Go test utilities, service wrappers, Playwright/Vitest
helpers, and retained ignored artifact roots. S2 should trace commands from the
task surface outward, while keeping service lifecycle, artifact ownership,
cleanup, failure modes, and authority decisions deferred to the designated later
sprints.
