---
doc_id: THR-S2-HANDOFF-2026-05-08
title: Testing Harness Recovery S2 Entrypoints and Commands Handoff
status: active
role: handoff
---

# Testing Harness Recovery S2 Entrypoints and Commands Handoff

## Session metadata

| Field | Value |
|---|---|
| Handoff ID | `2026-05-08-s2-entrypoints-and-commands` |
| Session date | `2026-05-08` |
| Agent/session identifier | `Codex S2 entrypoint implementation` |
| Repository revision | `9e523d9b110a7433ed08d4c35474f63f8c6c8080` |
| Branch | `main...origin/main` |
| Dirty state at start | Clean; `git status --short --branch` printed only `## main...origin/main` |
| Dirty state at end | Only recovery docs changed; final `git status --short --branch` listed modified/new files under `docs/testing-harness-spec-recovery-docs/**` only |
| Runtime platform | `Linux DeskRip 6.6.87.2-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Thu Jun 5 18:30:46 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| Current sprint | `S2: Entrypoints and commands` |
| Sprint status after session | `complete` if final git audit shows only recovery docs changed |
| Recovery-doc paths changed | `entrypoint-command-map.md`, `sequencing-assumption-list.md`, `03-sprint-plan.md`, `ambiguity-register.md`, `source-limit-log.md`, `harness-inventory.md`, this handoff |
| Implementation files changed | Must be none; final audit required |

## Work completed this session

- Created the S2 entrypoint command map.
- Created the S2 sequencing assumption list.
- Reconciled all `122` task-surface targets to command families.
- Mapped root and `apps/web` package scripts as alternate entrypoints.
- Mapped aggregate command flows for `test-fast`, `test`, `check`, `ci`, `release-check`, service-backed scheduler targets, and browser stages.
- Updated source limits and ambiguities for static-only runtime behavior, package-script authority, env precedence, mutating commands, and script CLI usage recovery.
- Updated S2 status, blockers, findings, and handoff fields in the sprint plan.

## Files and surfaces inspected

| Surface | Why inspected | Evidence status | Notes |
|---|---|---|---|
| `docs/testing-harness-spec-recovery-docs/**` S1 outputs | S2 inputs and register templates | `observed` | Used S1 inventory, uninvoked list, embedded logic list, source limits, ambiguity register, and handoff. |
| `Makefile`, `tools/task_surface.generated.mk`, `tools/task_surface_manifest.json` | Primary command declaration surface | `observed/runtime_observed` | Current surface is 77 public, 17 check-internal, 28 helper-only targets. |
| `tools/check_schedule_manifest.json`, `scripts/run-check-schedule.mjs` | `make check` scheduler flow | `observed` | `check` has 96 work units and resource limits for host/service/browser lanes. |
| `tools/service_backed_schedule_manifest.json`, `scripts/run-service-backed-schedule.mjs` | Service-backed scheduler flow | `observed` | `test-service-backed` and `check-service-backed` have 8 work-unit sources; `test-fast-service-backed` has 4 backend sources. |
| `tools/browser_e2e_batch_manifest.json`, browser scripts, Playwright configs | Browser stage and package browser aliases | `observed` | Browser stages are webserver-backed, functional, support, stateful, measurement, visual, resettable, and isolated. |
| `scripts/cartulary-runner.mjs`, `scripts/run-make-sequence.sh`, `scripts/run-phase-slice.mjs`, `scripts/run-harness-smoke.mjs`, `scripts/run-make-node-tool.mjs` | Core orchestration CLI contracts | `observed` | Usage and failure behavior recovered statically; scripts were not executed beyond Make explain/report surfaces. |
| `tools/testservices/main.go`, `scripts/start-web-e2e.sh` | Service wrapper and browser stack CLI contracts | `observed` | Lifecycle details deferred to S4/S5. |
| `package.json`, `apps/web/package.json`, `packages/*/package.json` | Package script entrypoints | `observed` | Root and app scripts exist; `packages/*` manifests declare no scripts. |
| `.github/**`, `scripts/ci/**` | CI surface | `observed/source_limit` | `.github/**` remains absent; provider-neutral CI scripts and `make ci` are recoverable. |

## Commands run

| Command | Purpose | Result | Exit code | Artifacts produced | Notes |
|---|---|---|---:|---|---|
| `git status --short --branch` | Record dirty state | Clean at start | 0 | none | Printed `## main...origin/main`. |
| `git rev-parse HEAD` | Record revision | Printed `9e523d9b110a7433ed08d4c35474f63f8c6c8080` | 0 | none | Metadata evidence. |
| `uname -a` | Record platform | Printed WSL2 Linux platform | 0 | none | Metadata evidence. |
| `date -Is` | Record timestamp | Printed `2026-05-08T20:23:28-04:00` | 0 | none | Metadata evidence. |
| `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | Enumerate task surface and harness smoke checks | Printed counts, targets, scripts, checks, phase dependency counts | 0 | none | Primary S2 runtime-observed discovery. |
| `make help-all` | Confirm public help surface | Printed public task tiers | 0 | none | Non-mutating discovery. |
| `make target-plan` | Confirm backend target families | Printed backend summaries | 0 | none | Non-mutating discovery. |
| `make target-plan-json` | Confirm backend target row shape | Printed large JSON | 0 | none | Used as static row-shape evidence only. |
| `make explain-phase PHASE=phase0` through `phase6` | Confirm phase execution dependencies and artifact references | Printed phase maps | 0 | none | Did not run phase tests. |
| `make explain-target TARGET=<target> DETAIL=summary` | Confirm representative target execution maps | Printed target summaries | 0 | none | Targets included backend, frontend, browser, scheduler, sequence, release, and maintenance examples. |
| `jq` manifest inspection commands | Count and summarize task, sequence, scheduler, service-backed, and browser batch manifests | Printed summaries | 0 | none | No files changed. |
| `sed`/`rg`/`find` inspection commands | Read CLI usage, env/defaults, package scripts, and script inventories | Printed source excerpts/lists | 0 | none | Static inspection only. |

No broad verification commands such as `make test-fast`, `make test`, `make check`, `make ci`, `make release-check`, browser E2E targets, service-backed targets, format, generate, cleanup, or baseline refresh/update targets were executed.

## Recovery artifacts updated

| Artifact | Update summary | Remaining gaps |
|---|---|---|
| `entrypoint-command-map.md` | Added S2 command map, target reconciliation, package scripts, aggregate flows, and harness smoke check list. | Runtime failure behavior remains static/source-limited. |
| `sequencing-assumption-list.md` | Added command ordering, nested scheduler, browser, service-wrapper, package-script, and mutating-command sequencing rows. | S3/S4/S5/S6 must expand artifact, lifecycle, failure, cleanup, and resource details. |
| `03-sprint-plan.md` | Marked S2 tasks and exit criteria complete; recorded concerns and S3 handoff. | Later sprint rows still placeholders. |
| `ambiguity-register.md` | Added S2 ambiguities `AMB-0011` through `AMB-0014`. | All S1/S2 ambiguities remain open. |
| `source-limit-log.md` | Added S2 source limits `SL-0006` and `SL-0007`. | Provider CI, runtime behavior, and retained artifacts remain unresolved. |
| `harness-inventory.md` | Added S2 invocation addendum rows. | Inventory roles unchanged. |

## Key findings

| Finding ID | Finding | Evidence status | Evidence reference | Impact |
|---|---|---|---|---|
| FIND-S2-0001 | The current Make task surface has `122` targets: `77` public, `17` check-internal, `28` helper-only. | `runtime_observed` | `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` | S2 map is complete against the current task-surface rows. |
| FIND-S2-0002 | Task-surface targets reconcile to generated recipe families: `print_help`, `phase_command`, `alias`, `go_target`, `service_backed_target`, `service_backed_schedule`, `check_schedule`, `sequence`, `browser_batch`, `node_tool`, `summary_target`, and `cleanup`. | `observed` | `tools/task_surface_manifest.json`; `entrypoint-command-map.md` | S3/S5 can link outputs and lifecycle to stable entrypoint rows. |
| FIND-S2-0003 | `test-fast`, `test`, `ci`, and `release-check` are Make sequences implemented by `scripts/run-make-sequence.sh`. | `observed` | `tools/task_surface_manifest.json`; `scripts/run-make-sequence.sh` | Aggregate flow ordering is now explicit. |
| FIND-S2-0004 | `check` is a check scheduler target with `96` work units and nested service-backed/browser resources. | `observed` | `tools/check_schedule_manifest.json`; `scripts/run-check-schedule.mjs` | S4/S6 should recover lifecycle and hazard behavior from this scheduler surface. |
| FIND-S2-0005 | Service-backed schedules expand from work-unit sources, not direct static work-unit arrays: full schedules include backend service-backed targets and browser stages, while fast service-backed is backend-only. | `observed` | `tools/service_backed_schedule_manifest.json` | S3/S4 can attach artifacts/services to scheduler source targets. |
| FIND-S2-0006 | Root and `apps/web` package scripts are real alternate entrypoints and can bypass Make result summaries/output policy. | `observed` | `package.json`; `apps/web/package.json` | Authority decision needed before treating them as harness normative commands. |
| FIND-S2-0007 | Provider CI workflow files remain absent. The recoverable CI surface is `make ci` plus `scripts/ci/**`. | `observed/source_limit` | `.github/**` absence from S0/S1; S2 command map | Provider workflow mapping remains unresolved. |

## New or updated ambiguities

| Ambiguity ID | Surface | Decision required | Blocking sprint |
|---|---|---|---|
| `AMB-0011` | Package scripts versus Make task surface | Decide whether package scripts are supported harness contracts or developer convenience aliases. | S8/S9 |
| `AMB-0012` | Env precedence | Decide public env override contract and precedence across Make, scripts, schedulers, Playwright, and Vitest. | S4/S5/S9 |
| `AMB-0013` | Mutating maintenance commands | Decide which writer commands are normative harness update workflows. | S3/S8 |
| `AMB-0014` | Static command contracts versus runtime-observed failure behavior | Decide which declared failures become stable contracts after S5 evidence. | S5/S9 |

## Source limits

| Source-limit ID | Surface | Limit | Impact | Follow-up |
|---|---|---|---|---|
| `SL-0001` | `.github/**` | Directory absent. | Provider CI workflow mapping incomplete. | Maintainer decision or external CI source. |
| `SL-0002` | Broad verification commands | S1/S2 intentionally did not run broad gates. | Runtime behavior not proven. | Later targeted runtime evidence if authorized. |
| `SL-0003` | Live service-backed runtime | Services were not started by S2. | Readiness/cleanup/retry behavior static-only. | S4/S6. |
| `SL-0004` | Retained `.cartulary/test-results/**` | Provenance/freshness not validated. | Artifact authority unknown. | S3/S5. |
| `SL-0005` | Planned phase7/phase8 files | Planned files absent. | Cannot map active phase7/phase8 commands. | Phase owner decision/future files. |
| `SL-0006` | S2 broad runtime behavior | Runtime gates, browser/service targets, cleanup, format, generate, baseline refreshes, and failures unexecuted. | Failure and lifecycle behavior remain static-only. | S3-S6. |
| `SL-0007` | Script CLI usage | CLI usage recovered statically, not by executing every usage branch. | Exit-code detail for usage paths remains static-only. | S5 targeted checks if needed. |

## Suggested next actions

1. In S3, use `entrypoint-command-map.md` as the stable entrypoint ID source for artifact and cleanup ownership.
2. Link each artifact class to entrypoint IDs rather than only to paths.
3. Distinguish writer/update commands from validation commands before defining fixture or generated-artifact authority.
4. Preserve source limits for `.github/**`, retained artifacts, and broad runtime behavior until later evidence resolves them.

## Do not repeat or redo

- Do not rerun broad gates solely to improve S2 command mapping.
- Do not hand-edit generated Make/task/scheduler manifests.
- Do not collapse package scripts into Make behavior without an authority decision.
- Do not treat retained `.cartulary/test-results/**` artifacts as fresh execution evidence.

## Implementation-change audit

| Check | Result |
|---|---|
| Harness implementation files modified | `no` |
| Test logic modified | `no` |
| CI behavior modified | `no` |
| Fixture contents modified | `no` |
| Cleanup scripts modified | `no` |
| Generated code or generated manifests modified | `no` |
| Only recovery docs changed | `yes` |

## Final handoff summary

S2 established a command-level map of the current harness. The primary command
surface is Make/task-surface driven, with package scripts and script CLIs
recorded separately as alternate or lower-level entrypoints. Aggregate command
flows, scheduler entrypoints, service-backed sources, browser stages, and
mutating maintenance commands are now explicit enough for S3 to attach
artifacts, cleanup behavior, fixture update rules, and generated-output
ownership to stable command IDs.
