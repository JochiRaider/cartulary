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
| S1. Inventory and boundary | `not_started` | `TODO:` | `TODO: harness_inventory` |
| S2. Entrypoints and commands | `not_started` | `TODO:` | `TODO: entrypoint_command_map` |
| S3. Fixtures, artifacts, and cleanup | `not_started` | `TODO:` | `TODO: artifact_ownership_matrix` |
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

- [ ] Inventory test directories and test runner configuration.
- [ ] Inventory package scripts and task-runner targets.
- [ ] Inventory CI workflows and local automation files.
- [ ] Inventory fixtures, snapshots, goldens, seeds, mocks, and sample projects.
- [ ] Inventory reports, coverage outputs, logs, screenshots, traces, and failure bundles.
- [ ] Inventory cleanup scripts and cleanup hooks.
- [ ] Classify each inventory item by role.
- [ ] Record read/write/generated/ignored/committed/external status.
- [ ] List harness-like files not invoked by discovered entrypoints.
- [ ] List harness behavior embedded in ordinary tests.

### Expected outputs

- `TODO: harness_inventory`
- `TODO: uninvoked_surface_list`
- `TODO: embedded_harness_logic_list`
- Updated `TODO: source_limit_log`

### Validation criteria

- Every discovered harness path has a role classification.
- Every generated or temp path has owner hypothesis or `TODO: owner_unknown`.
- Uninvoked surfaces are listed separately.
- Embedded harness logic is visible for later contract extraction.

### Exit criteria

- [ ] Inventory covers every discovered candidate surface.
- [ ] Missing or inaccessible areas are recorded as source limits.
- [ ] S2 can trace entrypoints using the inventory.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

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

- [ ] Identify all local validation commands.
- [ ] Identify all CI validation commands.
- [ ] Identify watch and debug commands.
- [ ] Identify fixture-update and snapshot-update commands.
- [ ] Identify cleanup and service commands.
- [ ] Record exact command strings.
- [ ] Trace each command to first implementation file or external tool.
- [ ] Record inputs, outputs, side effects, defaults, env vars, and failure behavior.
- [ ] Record sequencing assumptions and parallel invocation constraints.
- [ ] Map CI steps to harness entrypoints or non-harness commands.

### Expected outputs

- `TODO: entrypoint_command_map`
- `TODO: sequencing_assumption_list`
- Updated `TODO: harness_inventory`

### Validation criteria

- Every CI test step is mapped.
- Every package script that runs tests or validation is classified.
- Every fixture-update command is distinguished from normal validation.
- Every entrypoint has a success and failure contract or `TODO:` gap.

### Exit criteria

- [ ] Entrypoint map is row-complete for discovered commands.
- [ ] Unknown command behavior is recorded, not guessed.
- [ ] S3 and S5 can link artifacts and outputs to entrypoints.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

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

- [ ] Classify fixtures and golden or snapshot files.
- [ ] Classify generated artifacts, logs, reports, caches, and temp files.
- [ ] Identify canonical fixtures versus derived reports.
- [ ] Identify artifacts reused across runs.
- [ ] Identify artifacts created only on failure.
- [ ] Record mutation authority and update commands for canonical fixtures.
- [ ] Record cleanup trigger, cleanup owner, and idempotence expectations.
- [ ] Record external state ownership for databases, volumes, buckets, or emulator state.
- [ ] Identify cleanup gaps and artifact ownership ambiguities.

### Expected outputs

- `TODO: artifact_ownership_matrix`
- `TODO: cleanup_lifecycle_matrix`
- `TODO: shared_state_hazard_list`
- Updated `TODO: ambiguity_register`

### Validation criteria

- Every generated artifact has owner and cleanup rule or `TODO: owner_unknown`.
- Every canonical fixture has update rules or `TODO: update_rule_unknown`.
- Every derived report is marked non-authoritative unless consumed by execution.
- Every external state dependency has reset or isolation behavior recorded.

### Exit criteria

- [ ] Artifact ownership matrix is usable by spec drafting.
- [ ] Cleanup behavior is linked to entrypoints and lifecycle phases.
- [ ] S4 can focus on runtime services without losing artifact context.

### Status field

`not_started`

### Blocker field

`TODO:`

### Issues or concerns field

`TODO:`

### Findings or handoff notes for future sprints

`TODO:`

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
