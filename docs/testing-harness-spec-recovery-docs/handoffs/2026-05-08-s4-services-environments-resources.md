---
doc_id: THR-S4-HANDOFF
title: S4 Services Environments Resources Handoff
status: active
role: sprint-handoff
---

# S4 Services, Environments, and Resources Handoff

## Session metadata

| Field | Value |
|---|---|
| Sprint | S4: Services, environments, and resources |
| Status | `complete` |
| Repository root | `/home/askahn/code/cartulary` |
| Branch state at S4 start | `## main...origin/main` |
| HEAD revision at S4 start | `947e6254d6fbc5154ad4691e0485f6e22e3153e1` |
| Timestamp recorded | `2026-05-08T21:48:34-04:00` |
| Platform recorded | `Linux DeskRip 6.6.87.2-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Thu Jun 5 18:30:46 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| Recovery write boundary | `docs/testing-harness-spec-recovery-docs/**` |

## Controlling inputs used

- S2 `entrypoint-command-map.md`: especially `EP-0005` through `EP-0012`, `EP-0016`, `EP-0018`, and `EP-0020`.
- S2 `sequencing-assumption-list.md`.
- S3 `artifact-ownership-matrix.md`: especially `ART-0025` through `ART-0028`.
- S3 `cleanup-lifecycle-matrix.md`: especially `CLN-0011` through `CLN-0019`.
- S3 `shared-state-hazard-list.md`: especially `HAZ-S3-0007` through `HAZ-S3-0014`.
- Open inputs carried forward from earlier registers: `AMB-0006`, `AMB-0007`, `AMB-0012`, `AMB-0018`, `AMB-0019`, `AMB-0020`, `SL-0003`, and `SL-0011`.

## Discovery commands run

Only static or non-mutating discovery commands were used.

- `git status --short --branch`
- `git rev-parse HEAD`
- `date -Is`
- `uname -a`
- `rg` searches across S2/S3 docs, Make recipes, scheduler scripts, service helpers, browser lifecycle scripts, Playwright setup, reset route, and dev-service files
- `sed -n ...` reads of the files listed in the inspection surfaces below
- `jq` reads of scheduler and browser manifest JSON where structure was needed

S4 did not run service-start, reset, cleanup, browser, service-backed, format, generate, broad verification, Docker Compose, or testcontainers commands.

## Source areas inspected

- Command and scheduler layer: `Makefile`, `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json`, `tools/check_schedule_manifest.json`, `tools/service_backed_schedule_manifest.json`, `tools/browser_e2e_batch_manifest.json`, `tools/scheduler_resource_registry.json`, `tools/execution_topology_manifest.json`.
- Scheduler and resource code: `scripts/run-check-schedule.mjs`, `scripts/run-service-backed-schedule.mjs`, `scripts/lib/scheduler-resources.mjs`, `scripts/lib/scheduler/engine.mjs`, `scripts/lib/service-backed-schedule-manifest.mjs`, `scripts/lib/check-schedule-manifest.mjs`.
- Service startup and container setup: `tools/testservices/main.go`, `tools/testservices/*_test.go`, `internal/testutil/testcontainersx/**`, `internal/testutil/suiteservices/**`, `internal/testutil/pgtest/**`, `internal/testutil/s3test/**`.
- Local dev services: `docker-compose.dev.yml`, `scripts/dev-services.sh`, `scripts/dev-stack.sh`, `configs/dev/config.toml`, `configs/dev/bootstrap-admin.json`.
- Browser setup and lifecycle: `scripts/start-web-e2e.sh`, `scripts/reset-web-e2e-stack.sh`, `scripts/lib/process-lifecycle.sh`, `scripts/lib/playwright-owned-stack.sh`, `scripts/lib/web-e2e-lifecycle.sh`, `apps/web/playwright*.config.ts`, `apps/web/e2e/global-setup.ts`, `apps/web/e2e/global-teardown.ts`, `apps/web/e2e/harnessState.ts`, `apps/web/e2e/sessionSupport.ts`, `apps/web/e2e/fixtures.ts`.
- Reset, env, and temp paths: `internal/app/test_runtime_reset.go`, `scripts/lib/harness-scratch.sh`, `scripts/lib/runner-context.mjs`, `.gitignore`.

## Outputs produced

| Output | Purpose |
|---|---|
| `service-lifecycle-map.md` | Maps provision, configuration, start, readiness, use, reset, stop, cleanup, failure behavior, scope, evidence, and source limits for discovered services. |
| `environment-contract-observations.md` | Records env vars, defaults, validation, source-observed overrides, secrets/network assumptions, platform assumptions, and unresolved precedence gaps tied to entrypoints. |
| `resource-allocation-register.md` | Records scheduler lanes and concrete resources such as ports, DBs, buckets, locks, process groups, temp dirs, browser state, caches, and reset state. |
| `shared-state-hazard-list.md` | Adds `HAZ-S4-0001` through `HAZ-S4-0009` for service/resource lifecycle hazards. |
| `ambiguity-register.md` | Adds `AMB-0023` through `AMB-0028` for live readiness, reaper completion, port/platform behavior, env precedence, stale janitor bounds, and unsupported platform behavior. |
| `source-limit-log.md` | Adds `SL-0012` through `SL-0015` for live readiness/runtime, unsupported platform, timeout/interrupt/reaper cleanup, and env precedence limits. |
| `03-sprint-plan.md` | Marks S4 complete, updates output paths, records issues, and names S5/S6/S8 handoff constraints. |

## Key findings

- Service-backed work is centered on `tools/testservices`, with Postgres and MinIO testcontainers, a migrated Postgres template DB, fixture env injection, suite artifact dirs, cleanup/reaper paths, and stale browser fixture janitor logic.
- Browser E2E has two service modes: active testservices mode prepares a suite-owned DB/bucket fixture and later retires it, while standalone mode starts/reuses Docker Compose dev services and owns a temporary browser DB.
- Scheduler resource lanes bound concurrency at the harness level but do not prove concrete Docker, Postgres, MinIO, host process, browser, or port capacity.
- Every S4 service row has a ready condition or an explicit `TODO: readiness_unknown`; the reset route has no independent route-specific readiness probe.
- Every S4 shared resource row has an allocation/release rule or a conflict warning. Several are deliberately source-limited for live timing, cleanup, interruption, platform, or collision behavior.
- Environment assumptions are now tied to entrypoints and evidence files, but precedence remains unresolved where source does not prove it.

## Blockers

`none` for S4 documentation completion.

## Remaining issues and source limits

- `SL-0012`: live service readiness and runtime behavior were not executed.
- `SL-0013`: unsupported platform and missing-tool behavior remain incomplete.
- `SL-0014`: timeout, interrupt, detached reaper, active-connection cleanup, and stale janitor execution were not executed.
- `SL-0015`: cross-layer env override precedence was not executed.
- `AMB-0023` through `AMB-0028` remain open and should not be closed by inference from static source alone.

## Handoff to S5

S5 may use `SVC-*`, `ENV-*`, and `RES-*` row IDs to anchor observable-interface and failure-mode recovery. S5 must not treat S4 rows as final stdout/stderr, exit-code, structured-output, failure-bundle, retryability, or terminal-state contracts. S5 should also keep package-script behavior separate from Make entrypoints unless an S8 authority decision changes that boundary.

## Handoff to S6

S6 should consume `HAZ-S4-0001` through `HAZ-S4-0009`, `RES-*`, and `SVC-*` rows as the concrete resource/timing seed set. Priority hazard areas are scheduler lanes versus concrete capacity, detached reaper completion, Docker/testcontainers cleanup, stale browser fixture janitors, browser port races, Playwright shared state, reset-route ownership, Compose fixed ports/persistent volumes, and package-script bypass.

## Handoff to S8

S8 should settle authority for the app test runtime reset route, package scripts, local-dev service lifecycle, stale janitor destructive bounds, public env-var contracts, unsupported platform behavior, and external Go cache cleanup.

## Implementation-change audit

S4 made recovery-documentation edits only under `docs/testing-harness-spec-recovery-docs/**`. It did not modify harness implementation, service startup code, environment handling, cleanup scripts, generated outputs, fixtures, snapshots, lockfiles, or runtime artifacts.
