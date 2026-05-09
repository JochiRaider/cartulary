---
doc_id: THR-030
title: Testing Harness Specification Recovery Sprint Plan
status: draft
role: sprint-plan
---

# Testing Harness Specification Recovery Sprint Plan

## Document role

This document breaks the recovery effort into practical, agent-sized sprints. Use it as the progress board across multiple sessions.

Each sprint has explicit status, blocker, issues, and handoff fields. Agents must update those fields before stopping work.

## Status vocabulary

| Status        | Meaning                                                                                                  |
| ------------- | -------------------------------------------------------------------------------------------------------- |
| `not_started` | No material work has begun.                                                                              |
| `in_progress` | Work has begun and outputs are incomplete.                                                               |
| `blocked`     | Work cannot proceed without a decision, missing dependency, inaccessible source, or unavailable runtime. |
| `complete`    | Outputs exist and validation criteria pass.                                                              |
| `superseded`  | Sprint was replaced by a revised plan.                                                                   |

## Sprint progress board

| Sprint                                       | Status        | Blocker | Primary output                                                                                                                                                                                                                                                                           |
| -------------------------------------------- | ------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| S0. Charter and setup                        | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/recovery-charter.md`                                                                                                                                                                                                                            |
| S1. Inventory and boundary                   | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/harness-inventory.md`                                                                                                                                                                                                                           |
| S2. Entrypoints and commands                 | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md`                                                                                                                                                                                                                      |
| S3. Fixtures, artifacts, and cleanup         | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/artifact-ownership-matrix.md`                                                                                                                                                                                                                   |
| S4. Services, environments, and resources    | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/service-lifecycle-map.md`                                                                                                                                                                                                                       |
| S5. Lifecycle, interfaces, and failures      | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/observable-interface-map.md` and `docs/testing-harness-spec-recovery-docs/failure-mode-register.md`                                                                                                                                             |
| S6. Hazards, resources, and timing           | `complete`    | `none`  | `docs/testing-harness-spec-recovery-docs/race-timing-resource-register.md`                                                                                                                                                                                                               |
| S7. NLSpec, acceptance, roadmap, and handoff | `in_progress` | `none`  | `docs/testing-harness-spec-recovery-docs/harness-nlspec.md`, `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`, `docs/testing-harness-spec-recovery-docs/harness-implementation-roadmap.md`, and `docs/testing-harness-spec-recovery-docs/harness-review-packet.md` |
| S8. Authority and preservation follow-up     | `in_progress` | `none`  | `docs/testing-harness-spec-recovery-docs/preservation-matrix.md` and `docs/testing-harness-spec-recovery-docs/harness-authority-map.md`                                                                                                                                                  |

## S0: Charter and setup

### Sprint objective

Establish the recovery boundary, allowed writes, evidence labels, repository state, and implementation-change prohibition.

### Required inputs

- Main project specification.
- Repository root.
- Existing harness documentation.
- CI configuration.
- Build and test configuration.
- Operator-provided recovery scope.

### Concrete tasks

- [x] Record repository revision, branch, dirty state, runtime platform, package manager, and current date.
- [x] Create `docs/testing-harness-spec-recovery-docs/recovery-charter.md`.
- [x] Define candidate harness boundary.
- [x] Define evidence labels.
- [x] Record permitted recovery-doc paths.
- [x] Record prohibited implementation edits.
- [x] Initialize `docs/testing-harness-spec-recovery-docs/source-limit-log.md`.
- [x] Initialize this sprint plan status fields.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/recovery-charter.md`
- `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/recovery-charter.md#provisional-harness-boundary-candidate-list`
- Updated `03-sprint-plan.md`

### Validation criteria

- Charter states what the agent may write.
- Charter states what the agent must not modify.
- Candidate harness boundary is explicit and provisional.
- Source-limit log exists before deeper recovery begins.

### Exit criteria

- [x] Charter exists.
- [x] Source limits initialized.
- [x] Recovery write locations established.
- [x] Harness implementation rewrite prohibition recorded.
- [x] S1 can begin without relying on transcript memory.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

CI workflow configuration under `.github/**` was unavailable because `.github/` was absent in the working tree. This is recorded as source limit `SL-0001` and does not block S1 inventory.

### Findings or handoff notes for future sprints

S0 created `docs/testing-harness-spec-recovery-docs/recovery-charter.md`, initialized `docs/testing-harness-spec-recovery-docs/source-limit-log.md`, and added handoff `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s0-charter-and-setup.md`. S1 should use the charter boundary list as a discovery seed, keep writes under `docs/testing-harness-spec-recovery-docs/**`, and avoid implementation, fixture, CI, cleanup, generated-code, or lockfile edits.

## S1: Inventory and boundary

### Sprint objective

Produce a complete candidate inventory of harness-related files, directories, configs, generated paths, logs, cleanup surfaces, and embedded harness logic.

### Required inputs

- S0 charter.
- Repository tree.
- Test directories.
- CI files.
- Package and build manifests.
- Task-runner files.
- Ignore files.
- Existing generated artifacts.

### Concrete tasks

- [x] Inventory test directories and test runner configuration.
- [x] Inventory package scripts and task-runner targets.
- [x] Inventory CI workflows and local automation files.
- [x] Inventory fixtures, snapshots, goldens, seeds, mocks, and sample projects.
- [x] Inventory reports, coverage outputs, logs, screenshots, traces, and failure bundles.
- [x] Inventory cleanup scripts and cleanup hooks.
- [x] Classify each inventory item by role.
- [x] Record read/write/generated/ignored/committed/external status.
- [x] List harness-like files not invoked by discovered entrypoints.
- [x] List harness behavior embedded in ordinary tests.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/harness-inventory.md`
- `docs/testing-harness-spec-recovery-docs/uninvoked-surface-list.md`
- `docs/testing-harness-spec-recovery-docs/embedded-harness-logic-list.md`
- `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s1-inventory-and-boundary.md`

### Validation criteria

- Every discovered harness path has a role classification.
- Every generated or temp path has owner hypothesis or `TODO: owner_unknown`.
- Uninvoked surfaces are listed separately.
- Embedded harness logic is visible for later contract extraction.

### Exit criteria

- [x] Inventory covers every discovered candidate surface.
- [x] Missing or inaccessible areas are recorded as source limits.
- [x] S2 can trace entrypoints using the inventory.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

Open ambiguity/source-limit items remain for later sprints: absent `.github/**` CI workflow source (`SL-0001`, `AMB-0001`), intentionally unrun broad verification commands (`SL-0002`), static-only service lifecycle evidence (`SL-0003`, `AMB-0007`), retained artifact provenance gaps (`SL-0004`, `AMB-0002`, `AMB-0010`), planned phase7/phase8 missing files (`SL-0005`, `AMB-0004`), embedded harness/product-test ownership (`AMB-0005`), test runtime reset owner boundary (`AMB-0006`), cleanup owner/idempotency (`AMB-0008`), and local-dev versus verification-contract boundary (`AMB-0009`).

### Findings or handoff notes for future sprints

S1 created `docs/testing-harness-spec-recovery-docs/harness-inventory.md`, `docs/testing-harness-spec-recovery-docs/uninvoked-surface-list.md`, `docs/testing-harness-spec-recovery-docs/embedded-harness-logic-list.md`, `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`, and handoff `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s1-inventory-and-boundary.md`. The harness boundary is Make/task-surface centered and spans generated task/scheduler manifests, shell and Node orchestration scripts, Go test utilities, `tools/testservices`, Playwright/Vitest harness code, committed fixtures/goldens/snapshots, ignored runtime artifacts, and selected test-only product-package hooks. S2 should begin with `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `tools/task_surface_manifest.json`, and `tools/task_surface.generated.mk`, then trace each entrypoint to exact command contracts while preserving S1's distinction between harness mechanics and product assertions.

## S2: Entrypoints and commands

### Sprint objective

Recover every harness invocation path and its caller-visible contract.

### Required inputs

- S1 inventory.
- Package scripts.
- CI job steps.
- CLI definitions.
- Shell or task scripts.
- Pre-commit or release validation scripts.
- Developer docs.

### Concrete tasks

- [x] Identify all local validation commands.
- [x] Identify all CI validation commands.
- [x] Identify watch and debug commands.
- [x] Identify fixture-update and snapshot-update commands.
- [x] Identify cleanup and service commands.
- [x] Record exact command strings.
- [x] Trace each command to first implementation file or external tool.
- [x] Record inputs, outputs, side effects, defaults, env vars, and failure behavior.
- [x] Record sequencing assumptions and parallel invocation constraints.
- [x] Map CI steps to harness entrypoints or non-harness commands.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md`
- `docs/testing-harness-spec-recovery-docs/sequencing-assumption-list.md`
- Updated `docs/testing-harness-spec-recovery-docs/harness-inventory.md`
- Updated `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s2-entrypoints-and-commands.md`

### Validation criteria

- Every CI test step is mapped.
- Every package script that runs tests or validation is classified.
- Every fixture-update command is distinguished from normal validation.
- Every entrypoint has a success and failure contract or `TODO:` gap.

### Exit criteria

- [x] Entrypoint map is row-complete for discovered commands.
- [x] Unknown command behavior is recorded, not guessed.
- [x] S3 and S5 can link artifacts and outputs to entrypoints.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S2 did not execute broad runtime gates or mutating commands. Runtime success and failure behavior remains source-limited for `make test-fast`, `make test`, `make check`, `make ci`, `make release-check`, service-backed targets, browser E2E targets, cleanup, format, generate, baseline refresh/update targets, and failure scenarios. `.github/**` remains absent, so provider CI workflow steps are still unresolved. Package scripts are mapped as alternate entrypoints that can bypass Make output and scheduler policy.

### Findings or handoff notes for future sprints

S2 created `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md`, `docs/testing-harness-spec-recovery-docs/sequencing-assumption-list.md`, and handoff `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s2-entrypoints-and-commands.md`. The current command surface reconciles to 122 Make targets: 77 public, 17 check-internal, and 28 helper-only. S2 maps those targets to command families, maps root and `apps/web` package scripts separately, records `47` logical harness smoke checks, and carries forward source limits for absent `.github/**`, broad runtime behavior, static-only script usage, retained artifact provenance, and planned phase7/phase8 files. S3 should use `entrypoint-command-map.md` and `sequencing-assumption-list.md` to link artifacts, cleanup, generated outputs, and fixture update behavior back to stable entrypoint IDs.

## S3: Fixtures, artifacts, and cleanup

### Sprint objective

Recover authority, ownership, lifecycle, mutation, persistence, and cleanup rules for fixtures and generated artifacts.

### Required inputs

- S1 inventory.
- S2 entrypoint map.
- Fixture directories.
- Snapshot and golden files.
- Fixture update commands.
- Test setup and teardown hooks.
- Ignore files.
- Cleanup scripts.
- Runtime-generated paths.

### Concrete tasks

- [x] Classify fixtures and golden or snapshot files.
- [x] Classify generated artifacts, logs, reports, caches, and temp files.
- [x] Identify canonical fixtures versus derived reports.
- [x] Identify artifacts reused across runs.
- [x] Identify artifacts created only on failure.
- [x] Record mutation authority and update commands for canonical fixtures.
- [x] Record cleanup trigger, cleanup owner, and idempotence expectations.
- [x] Record external state ownership for databases, volumes, buckets, or emulator state.
- [x] Identify cleanup gaps and artifact ownership ambiguities.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/artifact-ownership-matrix.md`
- `docs/testing-harness-spec-recovery-docs/cleanup-lifecycle-matrix.md`
- `docs/testing-harness-spec-recovery-docs/shared-state-hazard-list.md`
- Updated `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s3-fixtures-artifacts-and-cleanup.md`

### Validation criteria

- Every generated artifact has owner and cleanup rule or `TODO: owner_unknown`.
- Every canonical fixture has update rules or `TODO: update_rule_unknown`.
- Every derived report is marked non-authoritative unless consumed by execution.
- Every external state dependency has reset or isolation behavior recorded.

### Exit criteria

- [x] Artifact ownership matrix is usable by spec drafting.
- [x] Cleanup behavior is linked to entrypoints and lifecycle phases.
- [x] S4 can focus on runtime services without losing artifact context.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S3 did not execute mutating writer or cleanup commands. Fixture, golden, and visual snapshot update authority remains unresolved (`AMB-0015`, `AMB-0022`) and is recorded as `TODO: update_rule_unknown` where no supported update command was found. Runtime cleanup on timeout/interrupt, failure-only bundle schemas, retained artifact provenance, destructive stale fixture janitor boundaries, and live Postgres/MinIO/browser behavior remain source-limited (`SL-0008` through `SL-0011`).

### Findings or handoff notes for future sprints

S3 created `docs/testing-harness-spec-recovery-docs/artifact-ownership-matrix.md`, `docs/testing-harness-spec-recovery-docs/cleanup-lifecycle-matrix.md`, `docs/testing-harness-spec-recovery-docs/shared-state-hazard-list.md`, and handoff `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s3-fixtures-artifacts-and-cleanup.md`. The key S4 inputs are external-state rows for Postgres databases/templates/transactions, MinIO buckets/prefixes, browser runtime roots, Playwright shared state, process groups, ports, and test runtime reset behavior. S4 should recover provision/start/ready/reset/stop/reaper details without rediscovering artifact owners or cleanup surfaces.

S3 audit follow-ups are mapped below without changing S4 status.

| Follow-up ID     | Recovery artifact update                                                                                                                                                                                            | Target sprint or owner path                               | Blocking status                                                               |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `AUD-S3-FU-0001` | `artifact-ownership-matrix.md` replaces premature future hazard references in `ART-0025` and `ART-0026` with existing `HAZ-S3-*` rows; future sprint IDs must exist before citation.                                | S3 follow-up complete; S4 may create later resource rows. | Blocking before S4/S6 consume those rows.                                     |
| `AUD-S3-FU-0002` | `ambiguity-register.md` keeps `AMB-0015` and `AMB-0022` open with owner-decision prompts and default no-refresh authority.                                                                                          | S8 or explicit harness/browser owner decision.            | Non-blocking for S4.                                                          |
| `AUD-S3-FU-0003` | `artifact-ownership-matrix.md`, `ambiguity-register.md`, and `source-limit-log.md` define retained-run freshness as explicit run selection, with newest-run fallback only for human investigation until S5 decides. | S5, with S8 authority if policy becomes normative.        | Blocking before observable-interface, drift, or baseline rules are finalized. |
| `AUD-S3-FU-0004` | `source-limit-log.md` assigns runner logs, watchdog JSON, Playwright traces/screenshots/videos, and reports to S5 schema recovery.                                                                                  | S5.                                                       | Blocking before failure-mode schema claims.                                   |
| `AUD-S3-FU-0005` | `cleanup-lifecycle-matrix.md`, `ambiguity-register.md`, and `source-limit-log.md` assign live cleanup, stale janitor bounds, active connections, port release, timeout, and interrupt behavior to S4/S6.            | S4 service lifecycle and S6 resource/timing hazards.      | Blocking before S6 hazard classification.                                     |
| `AUD-S3-FU-0006` | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, and `ambiguity-register.md` keep direct package-script artifacts and cleanup authority separate from Make behavior.                                  | S8 authority classification.                              | Non-blocking for S4.                                                          |
| `AUD-S3-FU-0007` | `artifact-ownership-matrix.md` and `ambiguity-register.md` classify external Go cache cleanup as an S8 authority decision, with proposed tool-managed external default.                                             | S8 authority classification.                              | Non-blocking for S4.                                                          |

## S4: Services, environments, and resources

### Sprint objective

Recover service lifecycle, environment assumptions, resource allocation, isolation, readiness, and reset behavior.

### Required inputs

- S2 entrypoint map.
- S3 artifact ownership matrix.
- Service startup scripts.
- Container, emulator, browser, or database setup.
- Port and temp-path allocation code.
- Health checks.
- Retry and wait loops.
- Secrets and environment variable handling.

### Concrete tasks

- [x] Inventory service and environment dependencies.
- [x] Record provision, configure, start, ready-check, use, reset, stop, and cleanup phases.
- [x] Record readiness checks and timeout behavior.
- [x] Record isolation scope.
- [x] Record port, socket, lock, DB, browser profile, process, worker, and container allocation rules.
- [x] Record secret and network assumptions.
- [x] Record service failure behavior.
- [x] Record unsupported platform behavior.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/service-lifecycle-map.md`
- `docs/testing-harness-spec-recovery-docs/environment-contract-observations.md`
- `docs/testing-harness-spec-recovery-docs/resource-allocation-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/shared-state-hazard-list.md`
- Updated `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s4-services-environments-resources.md`

### Validation criteria

- Every service has a ready condition or `TODO: readiness_unknown`.
- Every shared resource has an allocation rule or conflict warning.
- Environment assumptions are tied to entrypoints.
- Cleanup paths are documented.

### Exit criteria

- [x] Service lifecycle map is row-complete for discovered services.
- [x] Resource allocation rules are known or explicitly unknown.
- [x] S6 can analyze hazards using concrete resource rows.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S4 intentionally did not run live service-backed targets, browser E2E targets, Docker/testcontainers, Docker Compose, reset routes, cleanup paths, timeout/interrupt scenarios, formatters, generators, or broad verification gates. Runtime readiness, timeout, interrupt, detached reaper completion, stale fixture janitor execution, active DB connection cleanup, package-script bypass behavior, and env override precedence remain source-limited under `SL-0012` through `SL-0015` and open ambiguity rows `AMB-0023` through `AMB-0028`.

### Findings or handoff notes for future sprints

S4 created `docs/testing-harness-spec-recovery-docs/service-lifecycle-map.md`, `docs/testing-harness-spec-recovery-docs/environment-contract-observations.md`, `docs/testing-harness-spec-recovery-docs/resource-allocation-register.md`, and handoff `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s4-services-environments-resources.md`. S4 added S4 hazard rows `HAZ-S4-0001` through `HAZ-S4-0009`, ambiguity rows `AMB-0023` through `AMB-0028`, and source-limit rows `SL-0012` through `SL-0015`.

S5 may consume S4 service, env, and resource row IDs to map observable interfaces and failure outputs, but must not treat S4 static lifecycle evidence as settled stdout/stderr, schema, runtime timing, failure taxonomy, or cleanup guarantee evidence. S6 should use `RES-*`, `SVC-*`, and `HAZ-S4-*` rows to recover race, timing, cancellation, platform, and resource hazards. S8 should settle authority for package scripts, local-dev services, environment contracts, external cache cleanup, stale janitors, and the app test runtime reset route.

## S5: Lifecycle, interfaces, and failures

### Sprint objective

Recover the testing harness's caller-visible behavior by mapping stdout,
stderr, exit codes, structured reports, logs, CI-adjacent output, failure
artifacts, lifecycle phases, terminal states, partial-completion states, output
consumers, and current failure modes back to S2 entrypoints, S3 artifacts, and
S4 services/resources.

### Required inputs

- S2 entrypoint map, sequencing assumptions, S2 handoff, and S2 audit follow-ups.
- S3 artifact ownership matrix, cleanup lifecycle matrix, shared-state hazards,
  and S3 ambiguity/source-limit rows.
- S4 service lifecycle map, environment observations, resource allocation
  register, hazards, handoff, and audit follow-ups.
- Recovery charter, registers/checklists, source-limit log, and ambiguity
  register.
- Static evidence from CLI/script sources, structured report helpers, logs,
  retained-artifact readers, package manifests, CI helpers, and failure-artifact
  writers.

### Concrete tasks

- [x] Map stdout, stderr, and exit codes for major S2 entrypoint families.
- [x] Map structured reports and schemas for machine-consumed outputs.
- [x] Map failure bundles, logs, dashboards, retained reports, and diagnostics
  tied to entrypoints, artifacts, services, and resources.
- [x] Distinguish machine-consumed outputs from human diagnostics.
- [x] Reconstruct lifecycle phases for each major entrypoint family.
- [x] Record terminal states and partial-completion states.
- [x] Build failure-mode taxonomy.
- [x] Populate failure-mode register with observed and plausible recurring
  failures from S2/S3/S4/S5 evidence.
- [x] Separate product assertion failures from harness operational failures.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/observable-interface-map.md`
- `docs/testing-harness-spec-recovery-docs/structured-output-schema-notes.md`
- `docs/testing-harness-spec-recovery-docs/output-consumer-map.md`
- `docs/testing-harness-spec-recovery-docs/harness-lifecycle-map.md`
- `docs/testing-harness-spec-recovery-docs/phase-transition-table.md`
- `docs/testing-harness-spec-recovery-docs/partial-completion-state-list.md`
- `docs/testing-harness-spec-recovery-docs/failure-mode-register.md`
- `docs/testing-harness-spec-recovery-docs/failure-class-taxonomy.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s5-lifecycle-interfaces-failures.md`

### Validation criteria

- Every machine-consumed output has a schema or `TODO: schema_unknown`.
- Every major entrypoint has lifecycle phases and terminal states.
- Every recurring failure is registered.
- Retryability is explicit for every failure class.

### Exit criteria

- [x] Observable interface map and failure register are usable by NLSpec drafting.
- [x] Lifecycle, transition, and partial-state docs cover major entrypoint
  families.
- [x] Structured output schema notes have no unclassified machine-consumed output.
- [x] S6 can connect failures to race, timing, resource, cleanup, timeout, and
  interrupt issues.
- [x] S8 authority decisions are routed without being normalized by S5.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S5 was refreshed for the full lifecycle/interface/failure recovery scope. It
did not run service-backed targets, browser E2E targets, Docker/testcontainers,
Docker Compose, reset routes, cleanup paths, formatters, generators, baseline
refreshes, broad gates, or controlled failure commands. Runtime readiness,
cleanup strength, platform behavior, failure-only bundle schemas, CI provider
annotations, and retained-run freshness remain source-limited unless a later
authorized runtime or artifact-selection pass provides evidence.

### Findings or handoff notes for future sprints

S5 refreshed `observable-interface-map.md`,
`structured-output-schema-notes.md`, `output-consumer-map.md`,
`harness-lifecycle-map.md`, `phase-transition-table.md`,
`partial-completion-state-list.md`, `failure-mode-register.md`,
`failure-class-taxonomy.md`, S5 handoff
`handoffs/2026-05-08-s5-lifecycle-interfaces-failures.md`, and S5 audit
`audits/2026-05-08-s5-lifecycle-interfaces-failures-audit.md`.

S5 routes live readiness, cleanup guarantees, resource/capacity, timing,
interrupt, and stale-state behavior to S6. It routes reset-route, package-script,
local-dev, public env contract, supported platform, destructive janitor, and
external Go cache authority decisions to S8.

S5 audit follow-up `AUD-S5-BLOCK-0001` is resolved by correcting the affected
`failure-mode-register.md` linked output/schema references. S6 and S7 may
consume the corrected failure-mode register while preserving all existing
source-limit and authority-routing notes.

## S6: Hazards, resources, and timing

### Sprint objective

Identify hazards and classify behavior as intentional, accidental, compatibility-only, deprecated, redesign-required, or authority-required.

### Required inputs

- S3 artifact ownership matrix.
- S4 service lifecycle map.
- S5 lifecycle and failure registers.
- Parallelism configuration.
- Sleeps, polls, retries, timeouts, and sharding configuration.
- Main project specification.
- Existing harness docs and comments.
- Operator notes.

### Concrete tasks

- [x] Identify concurrency points for S4 follow-up surfaces.
- [x] Identify shared mutable resources for S4 follow-up surfaces.
- [x] Record sleeps, poll loops, timeouts, retries, debounces, and watch triggers.
- [x] Populate race/timing/resource register.
- [x] Connect recurring S4 follow-up failures to hazards.
- [x] Route preservation and authority questions to S8.
- [x] Preserve main-spec and harness authority conflicts for S8.
- [x] Update ambiguity/source-limit routing without closing source limits.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/race-timing-resource-register.md`
- `docs/testing-harness-spec-recovery-docs/concurrency-model-notes.md`
- `docs/testing-harness-spec-recovery-docs/timeout-retry-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/shared-state-hazard-list.md`
- Updated `docs/testing-harness-spec-recovery-docs/preservation-matrix.md`
- Updated `docs/testing-harness-spec-recovery-docs/harness-authority-map.md`
- Updated `docs/testing-harness-spec-recovery-docs/main-spec-conflict-list.md`
- Updated `docs/testing-harness-spec-recovery-docs/ambiguity-register.md`
- Updated `docs/testing-harness-spec-recovery-docs/source-limit-log.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s6-hazards-resources-timing.md`
- `docs/testing-harness-spec-recovery-docs/audits/2026-05-08-s6-hazards-resources-timing-audit.md`

### Validation criteria

- Every shared mutable resource appears in hazard analysis.
- Every fixed sleep, poll, retry, timeout, debounce/watch trigger, readiness check, signal wait, lock wait, and cleanup wait appears in timing analysis or an explicit source-limit row.
- Every major subsystem has a preservation classification.
- Every authority-required behavior has a specific owner question.
- Scheduler lanes are not described as concrete capacity guarantees.
- Runtime-sensitive claims remain `source_limit` unless actual runtime evidence exists.

### Exit criteria

- [x] Hazard register can drive spec resource and timing rules for S4 follow-up surfaces.
- [x] Preservation and authority questions are routed to S8 rather than silently closed.
- [x] Open runtime evidence gaps remain source-limited.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S6 was completed for S4 follow-up hazard routing. It did not execute runtime
readiness, service-backed, browser, Docker, Compose, reset, cleanup, timeout,
interrupt, or parent-death scenarios. Scheduler lanes are treated as scheduler
accounting only, not concrete host or service capacity guarantees.

### Findings or handoff notes for future sprints

S6 created or refreshed `race-timing-resource-register.md`,
`concurrency-model-notes.md`, `timeout-retry-register.md`,
`shared-state-hazard-list.md`, `preservation-matrix.md`,
`harness-authority-map.md`, `main-spec-conflict-list.md`,
`ambiguity-register.md`, `source-limit-log.md`, S6 handoff
`handoffs/2026-05-08-s6-hazards-resources-timing.md`, and S6 audit
`audits/2026-05-08-s6-hazards-resources-timing-audit.md`.

S6 keeps `SL-0012`, `SL-0013`, `SL-0014`, and `SL-0015` open. S8 must settle
authority for reset route, package scripts, local-dev services, stale janitor
destructive bounds, public environment contracts, supported platform/tool
profile, and external Go cache cleanup.

## S8: Authority and preservation follow-up

### Sprint objective

Classify S4 follow-up authority and preservation decisions without changing
harness behavior or inventing maintainer decisions.

### Required inputs

- S4 audit follow-up rows.
- S5 observable interface and failure outputs.
- S6 race/timing/resource outputs.
- Core 00 through Core 04 authority posture.
- `docs/domain.md` vocabulary context.
- Open ambiguity rows for reset route, package scripts, local-dev services,
  env precedence, platform support, stale janitors, and external caches.

### Concrete tasks

- [ ] Classify preservation posture for S4 follow-up surfaces.
- [ ] Record harness authority map and owner questions.
- [ ] Record main-spec conflict risks without resolving them by inference.
- [ ] Preserve maintainer-decision-required status where owner input is absent.
- [ ] Keep generated artifacts as downstream execution inputs, not behavior owners.

### Expected outputs

- `docs/testing-harness-spec-recovery-docs/preservation-matrix.md`
- `docs/testing-harness-spec-recovery-docs/harness-authority-map.md`
- `docs/testing-harness-spec-recovery-docs/main-spec-conflict-list.md`
- `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s8-authority-preservation.md`

### Validation criteria

- Every S4 authority follow-up has a proposed owner and required decision.
- Product behavior owned by Core 00 through Core 04 is not redefined.
- Owner-required decisions remain open unless an explicit maintainer decision exists.

### Exit criteria

- [ ] Preservation matrix can drive roadmap and normative decisions.
- [ ] Authority map is ready to include in the harness NLSpec.
- [ ] Main-spec conflict risks are recorded.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S8 did not make maintainer decisions. Reset route, package scripts, local-dev
services, stale janitor destructive bounds, public environment contracts,
supported platform/tool profile, and external Go cache cleanup remain
`maintainer_decision_required`.

### Findings or handoff notes for future sprints

S8 created `preservation-matrix.md`, `harness-authority-map.md`,
`main-spec-conflict-list.md`, handoff
`handoffs/2026-05-08-s8-authority-preservation.md`, and audit
`audits/2026-05-08-s8-authority-preservation-audit.md`. NLSpec drafting must
not normalize owner-required items without recorded maintainer decisions.

## S7: NLSpec, acceptance, roadmap, and handoff

### Sprint objective

Produce the harness NLSpec draft, acceptance matrix, roadmap, review packet, and final handoff.

### Required inputs

- S0 through S6 outputs plus S8 authority/preservation follow-up outputs.
- `templates/harness-nlspec-template.md`.
- Main project specification.
- Ambiguity register.
- Preservation matrix.
- Authority map.
- `s0-s6-gap-closure-plan.md`.
- `s7-s6-audit-gap-follow-up.md`.
- Existing tests and CI.

### Concrete tasks

- [ ] Create `docs/testing-harness-spec-recovery-docs/harness-nlspec.md`.
- [ ] Write document status, purpose, scope, non-goals, and authority relationship.
- [ ] Define terms, actors, and harness-owned surfaces.
- [ ] Draft entrypoint contracts.
- [ ] Draft run lifecycle and phase transitions at the contract level while preserving source limits.
- [ ] Draft fixture, artifact, service, environment, resource, timing, retry, timeout, cancellation, failure, cleanup, and diagnostic rules.
- [ ] Add tables required for cleanup tiers, acceptance routing, and source-limit carry-forward.
- [ ] Extract normative candidates into `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`.
- [ ] Assign verification method and current coverage to every criterion.
- [ ] Identify missing tests, missing fixtures, missing golden files, and future CI gates.
- [ ] Create `docs/testing-harness-spec-recovery-docs/harness-implementation-roadmap.md`.
- [ ] Produce `docs/testing-harness-spec-recovery-docs/harness-review-packet.md`.

### Expected outputs

- Final `docs/testing-harness-spec-recovery-docs/harness-nlspec.md`
- `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`
- Missing-test and future-gate list in `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md`
- `docs/testing-harness-spec-recovery-docs/harness-implementation-roadmap.md`
- `docs/testing-harness-spec-recovery-docs/harness-review-packet.md`
- `docs/testing-harness-spec-recovery-docs/harness-review-packet.md` final handoff section

### Validation criteria

- Draft specifies observable behavior over mechanism unless mechanism is contract-bearing.
- Draft separates observed behavior from intended normative behavior.
- Draft does not redefine product behavior owned by main spec.
- Every normative requirement has binary acceptance criterion or owner decision.
- Every roadmap item is future work, not implied completed work.
- Implementation changes preserve generated artifact ownership and do not redefine product behavior.

### Exit criteria

- [ ] Harness NLSpec draft is complete enough for maintainer review.
- [ ] Acceptance matrix is complete.
- [ ] Roadmap is complete.
- [ ] Review packet is complete.
- [ ] Source limits are summarized.
- [ ] Open owner decisions are consolidated.
- [ ] Verification commands have completed and outcomes are recorded.

### Status field

`complete`

### Blocker field

`none`

### Issues or concerns field

S7 implementation completed from the completed S0 through S6 and S8 recovery
outputs plus `s0-s6-gap-closure-plan.md`,
`s7-s6-audit-gap-follow-up.md`, and the 2026-05-09 maintainer decisions.
Runtime readiness, cleanup strength beyond selected evidence, retained failure
bundle schemas, environment precedence, full platform support, visual snapshot
refresh authority, active DB cleanup, parent-death cleanup, detached reaper hard
completion, and CI provider workflow behavior remain source-limited or
maintainer-decision-required.

Verification completed on 2026-05-09:
`make agent-finalize`, targeted reset and stale-janitor tests,
`make generated-artifact-policy-check`, `make json-shape-check`,
`make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `make test-fast`,
`make ci`, and `make release-check` all passed. `agent-finalize` skipped duration
baseline refresh because `RESULTS_DIR` was unset. Standalone
`scripts/test-print-target-plan.sh` still fails with the known stale Phase 5
extended-smoke phase-map accounting issue, which remains diagnostic and
non-blocking.

### Findings or handoff notes for future sprints

`s0-s6-gap-closure-plan.md` is the S7 preflight closure artifact for all S0
through S6 recovery gaps. `s7-s6-audit-gap-follow-up.md` remains the focused S6
audit carry-forward track. S7 must not draft final normative `MUST` language for
live readiness, cleanup completion beyond selected evidence, signal/timeout
behavior, detached reaper completion, active DB cleanup, port release, visual
snapshot refresh bounds, env precedence, CI provider annotations, Playwright
tool report internals, or concrete scheduler capacity guarantees unless later
authorized runtime evidence or explicit owner decisions exist. Decided S7 items
are recorded in `maintainer-decision-summary-2026-05-09.md`.
