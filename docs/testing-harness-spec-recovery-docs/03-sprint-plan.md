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

| Status | Meaning |
|---|---|
| `not_started` | No material work has begun. |
| `in_progress` | Work has begun and outputs are incomplete. |
| `blocked` | Work cannot proceed without a decision, missing dependency, inaccessible source, or unavailable runtime. |
| `complete` | Outputs exist and validation criteria pass. |
| `superseded` | Sprint was replaced by a revised plan. |

## Sprint progress board

| Sprint | Status | Blocker | Primary output |
|---|---|---|---|
| S0. Charter and setup | `complete` | `none` | `docs/testing-harness-spec-recovery-docs/recovery-charter.md` |
| S1. Inventory and boundary | `complete` | `none` | `docs/testing-harness-spec-recovery-docs/harness-inventory.md` |
| S2. Entrypoints and commands | `complete` | `none` | `docs/testing-harness-spec-recovery-docs/entrypoint-command-map.md` |
| S3. Fixtures, artifacts, and cleanup | `complete` | `none` | `docs/testing-harness-spec-recovery-docs/artifact-ownership-matrix.md` |
| S4. Services, environments, and resources | `not_started` | `TODO:` | `TODO: service_lifecycle_map` |
| S5. Lifecycle, interfaces, and failures | `not_started` | `TODO:` | `TODO: observable_interface_map` and `TODO: failure_mode_register` |
| S6. Hazards, intent, and authority | `not_started` | `TODO:` | `TODO: race_timing_resource_register` and `TODO: preservation_matrix` |
| S7. NLSpec, acceptance, roadmap, and handoff | `not_started` | `TODO:` | `TODO: harness_nlspec_draft_path` and `TODO: acceptance_matrix` |

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

| Follow-up ID | Recovery artifact update | Target sprint or owner path | Blocking status |
|---|---|---|---|
| `AUD-S3-FU-0001` | `artifact-ownership-matrix.md` replaces premature future hazard references in `ART-0025` and `ART-0026` with existing `HAZ-S3-*` rows; future sprint IDs must exist before citation. | S3 follow-up complete; S4 may create later resource rows. | Blocking before S4/S6 consume those rows. |
| `AUD-S3-FU-0002` | `ambiguity-register.md` keeps `AMB-0015` and `AMB-0022` open with owner-decision prompts and default no-refresh authority. | S8 or explicit harness/browser owner decision. | Non-blocking for S4. |
| `AUD-S3-FU-0003` | `artifact-ownership-matrix.md`, `ambiguity-register.md`, and `source-limit-log.md` define retained-run freshness as explicit run selection, with newest-run fallback only for human investigation until S5 decides. | S5, with S8 authority if policy becomes normative. | Blocking before observable-interface, drift, or baseline rules are finalized. |
| `AUD-S3-FU-0004` | `source-limit-log.md` assigns runner logs, watchdog JSON, Playwright traces/screenshots/videos, and reports to S5 schema recovery. | S5. | Blocking before failure-mode schema claims. |
| `AUD-S3-FU-0005` | `cleanup-lifecycle-matrix.md`, `ambiguity-register.md`, and `source-limit-log.md` assign live cleanup, stale janitor bounds, active connections, port release, timeout, and interrupt behavior to S4/S6. | S4 service lifecycle and S6 resource/timing hazards. | Blocking before S6 hazard classification. |
| `AUD-S3-FU-0006` | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, and `ambiguity-register.md` keep direct package-script artifacts and cleanup authority separate from Make behavior. | S8 authority classification. | Non-blocking for S4. |
| `AUD-S3-FU-0007` | `artifact-ownership-matrix.md` and `ambiguity-register.md` classify external Go cache cleanup as an S8 authority decision, with proposed tool-managed external default. | S8 authority classification. | Non-blocking for S4. |

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

- [ ] Inventory service and environment dependencies.
- [ ] Record provision, configure, start, ready-check, use, reset, stop, and cleanup phases.
- [ ] Record readiness checks and timeout behavior.
- [ ] Record isolation scope.
- [ ] Record port, socket, lock, DB, browser profile, process, worker, and container allocation rules.
- [ ] Record secret and network assumptions.
- [ ] Record service failure behavior.
- [ ] Record unsupported platform behavior.

### Expected outputs

- `TODO: service_lifecycle_map`
- `TODO: environment_contract_observations`
- `TODO: resource_allocation_register`
- Updated `TODO: shared_state_hazard_list`

### Validation criteria

- Every service has a ready condition or `TODO: readiness_unknown`.
- Every shared resource has an allocation rule or conflict warning.
- Environment assumptions are tied to entrypoints.
- Cleanup paths are documented.

### Exit criteria

- [ ] Service lifecycle map is row-complete for discovered services.
- [ ] Resource allocation rules are known or explicitly unknown.
- [ ] S6 can analyze hazards using concrete resource rows.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

## S5: Lifecycle, interfaces, and failures

### Sprint objective

Map caller-visible outputs, machine interfaces, lifecycle phases, terminal states, partial states, and current failure modes.

### Required inputs

- S2 entrypoint map.
- S3 artifact ownership matrix.
- S4 service lifecycle map.
- CLI output.
- Exit codes.
- Structured reports.
- Logs.
- CI annotations.
- Failure artifacts.
- Failing CI or local logs.

### Concrete tasks

- [ ] Map stdout, stderr, and exit codes.
- [ ] Map structured reports and schemas.
- [ ] Map coverage, CI annotations, failure bundles, logs, and dashboards.
- [ ] Distinguish machine-consumed outputs from human diagnostics.
- [ ] Reconstruct lifecycle phases for each major entrypoint.
- [ ] Record terminal states and partial-completion states.
- [ ] Build failure-mode taxonomy.
- [ ] Populate failure-mode register with observed and plausible failures.
- [ ] Separate product assertion failures from harness operational failures.

### Expected outputs

- `TODO: observable_interface_map`
- `TODO: structured_output_schema_notes`
- `TODO: output_consumer_map`
- `TODO: harness_lifecycle_map`
- `TODO: phase_transition_table`
- `TODO: partial_completion_state_list`
- `TODO: failure_mode_register`
- `TODO: failure_class_taxonomy`

### Validation criteria

- Every machine-consumed output has a schema or `TODO: schema_unknown`.
- Every major entrypoint has lifecycle phases and terminal states.
- Every recurring failure is registered.
- Retryability is explicit for every failure class.

### Exit criteria

- [ ] Observable interface map is usable by the NLSpec draft.
- [ ] Failure modes are specific enough for failure taxonomy writing.
- [ ] S6 can connect failures to race, timing, resource, and authority issues.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

## S6: Hazards, intent, and authority

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

- [ ] Identify concurrency points.
- [ ] Identify shared mutable resources.
- [ ] Record sleeps, poll loops, timeouts, retries, debounces, and watch triggers.
- [ ] Populate race/timing/resource register.
- [ ] Connect recurring failures to hazards.
- [ ] Classify significant behavior in preservation matrix.
- [ ] Record main-spec and harness conflicts.
- [ ] Draft authority map and owner questions.
- [ ] Update ambiguity register.

### Expected outputs

- `TODO: race_timing_resource_register`
- `TODO: concurrency_model_notes`
- `TODO: preservation_matrix`
- `TODO: ambiguity_register`
- `TODO: harness_authority_map`
- `TODO: main_spec_conflict_list`

### Validation criteria

- Every shared mutable resource appears in hazard analysis.
- Every fixed sleep and timeout appears in timing analysis.
- Every major subsystem has a preservation classification.
- Every authority-required behavior has a specific owner question.

### Exit criteria

- [ ] Hazard register can drive spec resource and timing rules.
- [ ] Preservation matrix can drive roadmap and normative decisions.
- [ ] Authority map is ready to include in the harness NLSpec.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

## S7: NLSpec, acceptance, roadmap, and handoff

### Sprint objective

Produce the harness NLSpec draft, acceptance matrix, roadmap, review packet, and final handoff.

### Required inputs

- S0 through S6 outputs.
- `templates/harness-nlspec-template.md`.
- Main project specification.
- Ambiguity register.
- Preservation matrix.
- Authority map.
- Existing tests and CI.

### Concrete tasks

- [ ] Copy template to `TODO: harness_nlspec_draft_path`.
- [ ] Write document status, purpose, scope, non-goals, and authority relationship.
- [ ] Define terms, actors, and harness-owned surfaces.
- [ ] Draft entrypoint contracts.
- [ ] Draft run lifecycle and phase transitions.
- [ ] Draft fixture, artifact, service, environment, resource, timing, retry, timeout, cancellation, failure, cleanup, and diagnostic rules.
- [ ] Add deterministic algorithms and tables required to close mechanics gaps.
- [ ] Extract every `must` and `must not` into `TODO: acceptance_matrix`.
- [ ] Assign verification method and current coverage to every criterion.
- [ ] Identify missing tests, missing fixtures, missing golden files, and future CI gates.
- [ ] Create roadmap with preserve/refactor/deprecate/redesign/remove/authority classifications.
- [ ] Produce maintainer review packet and final handoff.

### Expected outputs

- Final `TODO: harness_nlspec_draft_path`
- `TODO: acceptance_matrix`
- `TODO: missing_fixture_list`
- `TODO: future_ci_gate_list`
- `TODO: harness_roadmap`
- `TODO: harness_recovery_review_packet`
- Final handoff note

### Validation criteria

- Draft specifies observable behavior over mechanism unless mechanism is contract-bearing.
- Draft separates observed behavior from intended normative behavior.
- Draft does not redefine product behavior owned by main spec.
- Every normative requirement has binary acceptance criterion or owner decision.
- Every roadmap item is future work, not implied completed work.
- No implementation changes were made.

### Exit criteria

- [ ] Harness NLSpec draft is complete enough for maintainer review.
- [ ] Acceptance matrix is complete.
- [ ] Roadmap is complete.
- [ ] Review packet is complete.
- [ ] Source limits are summarized.
- [ ] Open owner decisions are consolidated.
- [ ] No implementation changes were made.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`
