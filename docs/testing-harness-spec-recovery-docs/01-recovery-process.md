---
doc_id: THR-010
title: Testing Harness Reverse-Specification Recovery Process
status: draft
role: process
---

# Testing Harness Reverse-Specification Recovery Process

## Document role

This document defines the staged recovery workflow. Agents should follow these stages in order, but may loop back when later evidence invalidates an earlier finding.

Each stage must update the sprint plan, source-limit log, and handoff notes when work stops.

## Process rules

1. Inspect exact files before making claims about repository behavior.
2. Use search for discovery, not as the sole source for final claims.
3. Do not modify harness implementation, tests, fixtures, CI, cleanup scripts, or service behavior during recovery.
4. Label every behavior as `observed`, `runtime_observed`, `inferred`, `assumed`, `contradiction`, `maintainer_decision_required`, or `source_limit`.
5. Record missing context with `TODO:`. Do not invent filenames, owners, defaults, or compatibility rules.
6. Distinguish canonical execution-driving surfaces from derived diagnostic surfaces.
7. Do not treat passing tests as proof of intended behavior.
8. Do not silently reconcile conflicts between main project spec, implementation, tests, fixtures, CI, and documentation.
9. Keep recovery artifacts small, updateable, and table-driven.
10. Move future implementation fixes into the roadmap instead of performing them during recovery.

## Stage map

| Stage | Name | Primary output |
|---:|---|---|
| 0 | Charter and scope freeze | `TODO: recovery_charter_path` |
| 1 | Harness inventory | `TODO: harness_inventory` |
| 2 | Entrypoint and command map | `TODO: entrypoint_command_map` |
| 3 | Fixtures, artifacts, and cleanup | `TODO: artifact_ownership_matrix` |
| 4 | Services, environments, and resources | `TODO: service_lifecycle_map` |
| 5 | Observable behavior and lifecycle | `TODO: observable_interface_map` and `TODO: harness_lifecycle_map` |
| 6 | Race, timing, and resource hazards | `TODO: race_timing_resource_register` |
| 7 | Failure-mode recovery | `TODO: failure_mode_register` |
| 8 | Intent and authority classification | `TODO: preservation_matrix` and `TODO: harness_authority_map` |
| 9 | Harness NLSpec contract pass | `harness-nlspec.md` |
| 10 | Mechanics pass | Updated `harness-nlspec.md` |
| 11 | Verification pass | `harness-acceptance-matrix.md` |
| 12 | Roadmap and review packet | `harness-implementation-roadmap.md` and `harness-review-packet.md` |

## Stage 0: Charter and scope freeze

### Objective

Define what counts as the testing harness, what may be inspected, what may be written, and what must not be changed.

### Inputs to inspect

- Main project specification.
- Existing harness documentation.
- CI workflows.
- Build and test configuration.
- Test directories.
- Operator-provided scope notes.

### Concrete agent actions

- Record repository revision, branch, dirty state, runtime platform, package manager, and current date.
- Define a provisional harness boundary.
- Record allowed recovery-doc write paths.
- Record prohibited implementation changes.
- Define evidence labels.
- Initialize source-limit log.

### Expected outputs

- `TODO: recovery_charter_path`
- `TODO: source_limit_log`
- `TODO: harness_boundary_candidate_list`

### Validation criteria

- The charter states what the agent may write.
- The charter states what the agent must not modify.
- The harness boundary is provisional and explicit.
- Source limits are initialized before deeper recovery begins.

### Completion checklist

- [ ] Repository revision recorded.
- [ ] Dirty state recorded.
- [ ] Runtime platform recorded.
- [ ] Recovery write paths recorded.
- [ ] Implementation rewrite prohibition recorded.
- [ ] Candidate harness boundary recorded.
- [ ] Evidence labels defined.
- [ ] Source limits initialized.

### Risks or ambiguity to record

- Harness code may be interleaved with product tests.
- CI may exercise behavior unavailable locally.
- Some behavior may depend on secrets, containers, ports, external services, or platform-specific defaults.

## Stage 1: Harness inventory

### Objective

Inventory every file, directory, configuration surface, generated path, and external state surface that belongs to or affects the harness.

### Inputs to inspect

- Repository tree.
- Test directories.
- Package and build manifests.
- CI workflows.
- Task-runner files.
- Ignore files.
- Fixture and generated-output directories.
- Cleanup scripts.

### Concrete agent actions

- Build a candidate inventory of harness-related paths and external surfaces.
- Classify each item as `entrypoint`, `orchestration`, `fixture`, `service`, `generated_artifact`, `temporary_artifact`, `log`, `cleanup`, `policy`, `adapter`, or `derived_view`.
- Record read/write/generated/ignored/committed/external status.
- Identify harness-like files not invoked by discovered entrypoints.
- Identify harness behavior embedded in ordinary tests.

### Expected outputs

- `TODO: harness_inventory`
- `TODO: uninvoked_surface_list`
- `TODO: embedded_harness_logic_list`

### Validation criteria

- Every discovered harness path has a role classification.
- Every generated or temp path has an owner hypothesis or `TODO: owner_unknown`.
- Inventory distinguishes committed fixtures from generated outputs.
- Uninvoked surfaces are listed separately.

### Completion checklist

- [ ] Test directories inventoried.
- [ ] Runner and config files inventoried.
- [ ] CI files inventoried.
- [ ] Fixture directories inventoried.
- [ ] Service definitions inventoried.
- [ ] Generated paths inventoried.
- [ ] Logs and reports inventoried.
- [ ] Cleanup paths inventoried.
- [ ] Uninvoked surfaces listed.
- [ ] Embedded harness logic listed.

### Risks or ambiguity to record

- Ignore rules may hide important state.
- Generated artifacts may be committed intentionally or accidentally.
- Some temp paths may be created only on failure.

## Stage 2: Entrypoint and command map

### Objective

Recover every way the harness can be invoked and what each invocation promises or assumes.

### Inputs to inspect

- Package scripts.
- CLI definitions.
- CI job steps.
- Developer docs.
- Shell scripts.
- Test runner configuration.
- Watch mode configuration.
- Pre-commit or release scripts.

### Concrete agent actions

- Identify all harness entrypoints.
- Record each command exactly as declared.
- Trace each command to the first implementation file it invokes.
- Record caller, mode, flags, environment variables, defaults, outputs, side effects, and failure behavior.
- Record whether commands may run concurrently, nest, or reuse shared state.
- Record implicit sequencing assumptions.

### Expected outputs

- `TODO: entrypoint_command_map`
- `TODO: sequencing_assumption_list`

### Validation criteria

- Every CI test step maps to a harness entrypoint or explicit non-harness command.
- Every package script that runs tests or validation is classified.
- Fixture-update commands are distinguished from ordinary validation.
- Every entrypoint has success and failure behavior or a `TODO:` gap.

### Completion checklist

- [ ] Local test commands mapped.
- [ ] CI test commands mapped.
- [ ] Watch/debug commands mapped.
- [ ] Fixture-update commands mapped.
- [ ] Cleanup commands mapped.
- [ ] Service commands mapped.
- [ ] Exit-code behavior recorded.
- [ ] Environment-variable behavior recorded.
- [ ] Command ordering assumptions recorded.
- [ ] Parallel invocation assumptions recorded.

### Risks or ambiguity to record

- Scripts may wrap other scripts recursively.
- CI may set environment variables unavailable locally.
- Developer docs may list obsolete commands.

## Stage 3: Fixtures, artifacts, and cleanup

### Objective

Define ownership, authority, lifecycle, mutation, persistence, and cleanup behavior for fixtures and generated artifacts.

### Inputs to inspect

- Fixture directories.
- Snapshot and golden files.
- Fixture update scripts.
- Test setup and teardown hooks.
- Temp-path creation code.
- Cache directories.
- Coverage and report directories.
- Ignore files.
- Cleanup scripts.
- External service state.

### Concrete agent actions

- List every fixture and generated-artifact class.
- Classify each artifact as `canonical_fixture`, `canonical_runtime_state`, `derived_report`, `ephemeral_scratch`, `diagnostic_log`, `external_state`, or `unknown_authority`.
- Record create, read, mutate, persist, commit, ignore, upload, and cleanup behavior.
- Record cleanup owner and trigger.
- Identify artifacts shared across test runs.
- Identify artifacts whose absence or staleness changes behavior.

### Expected outputs

- `TODO: artifact_ownership_matrix`
- `TODO: cleanup_lifecycle_matrix`
- `TODO: shared_state_hazard_list`

### Validation criteria

- Every generated artifact has owner and cleanup behavior or `TODO: owner_unknown`.
- Every canonical fixture has update rules or `TODO: update_rule_unknown`.
- Every derived report is marked non-authoritative unless execution consumes it.
- Every external state dependency has reset or isolation behavior recorded.

### Completion checklist

- [ ] Fixtures classified.
- [ ] Golden and snapshot files classified.
- [ ] Generated artifacts classified.
- [ ] Logs classified.
- [ ] Coverage and report outputs classified.
- [ ] Caches classified.
- [ ] Temp paths classified.
- [ ] External service state classified.
- [ ] Cleanup owners recorded.
- [ ] Shared-state hazards recorded.

### Risks or ambiguity to record

- Cleanup may be partial after failure.
- Snapshot update commands may mutate canonical truth.
- Logs may be parsed by later steps despite being described as diagnostic.

## Stage 4: Services, environments, and resources

### Objective

Recover how the harness starts, detects, uses, shares, resets, and stops runtime services and execution environments.

### Inputs to inspect

- Service startup scripts.
- Container or compose files.
- Browser automation setup.
- Database setup and teardown.
- Mock server setup.
- Port allocation logic.
- Health checks.
- Retry and wait loops.
- Secrets and environment variable handling.

### Concrete agent actions

- Inventory service and environment dependencies.
- Record lifecycle phases: provision, configure, start, ready-check, use, reset, stop, cleanup.
- Identify readiness conditions and whether they are deterministic.
- Record port, socket, lock, database, browser profile, process, container, and worker-pool allocation rules.
- Record isolation scope.
- Record secret exposure and network access assumptions.
- Record failure behavior when services are unavailable, slow, already running, or partially stopped.

### Expected outputs

- `TODO: service_lifecycle_map`
- `TODO: environment_contract_observations`
- `TODO: resource_allocation_register`

### Validation criteria

- Every service has a ready condition or `TODO: readiness_unknown`.
- Every shared resource has an allocation rule or conflict warning.
- Environment assumptions are tied to entrypoints.
- Service cleanup paths are documented.

### Completion checklist

- [ ] Services inventoried.
- [ ] Containers inventoried.
- [ ] Databases inventoried.
- [ ] Browser or emulator dependencies inventoried.
- [ ] Port allocation recorded.
- [ ] Worker allocation recorded.
- [ ] Readiness checks recorded.
- [ ] Reset rules recorded.
- [ ] Stop rules recorded.
- [ ] Secret and network assumptions recorded.

### Risks or ambiguity to record

- Service readiness may rely on fixed sleeps.
- Port allocation may collide across parallel runs.
- Service state may leak between tests.
- Local and CI service lifecycles may differ.

## Stage 5: Observable behavior and lifecycle

### Objective

Map what the harness exposes to callers and reconstruct major entrypoint lifecycles from invocation through teardown.

### Inputs to inspect

- CLI output.
- Exit codes.
- Structured reports.
- Coverage outputs.
- CI annotations.
- Logs.
- Failure artifacts.
- Global setup and teardown hooks.
- Retry and report-generation code.

### Concrete agent actions

- Identify all caller-visible outputs.
- Distinguish machine interfaces from human diagnostics.
- Record schemas for structured outputs.
- Record ordering guarantees or nondeterminism.
- Record exit-code semantics.
- Draw lifecycle phases for every major entrypoint.
- Identify terminal states, skipped phases, retry loops, and partial-completion states.
- Record cleanup behavior on success, failure, timeout, interrupt, and crash where observable.

### Expected outputs

- `TODO: observable_interface_map`
- `TODO: structured_output_schema_notes`
- `TODO: output_consumer_map`
- `TODO: harness_lifecycle_map`
- `TODO: phase_transition_table`
- `TODO: partial_completion_state_list`

### Validation criteria

- Every machine-consumed output has a schema or `TODO: schema_unknown`.
- Every exit code used by callers has a meaning or `TODO:` gap.
- Human-only diagnostics are marked non-authoritative.
- Every major entrypoint has lifecycle phases and terminal states.

### Completion checklist

- [ ] Exit codes mapped.
- [ ] Stdout and stderr behavior mapped.
- [ ] Structured reports mapped.
- [ ] CI annotations mapped.
- [ ] Failure artifacts mapped.
- [ ] Output consumers mapped.
- [ ] Lifecycle phases recorded.
- [ ] Terminal states recorded.
- [ ] Retry and rerun behavior recorded.
- [ ] Cleanup conditions recorded.

### Risks or ambiguity to record

- Human logs may be scraped by automation.
- Output ordering may vary under parallelism.
- Report paths may differ locally and in CI.
- Interrupt handling may differ from ordinary failure.

## Stage 6: Race, timing, and resource hazards

### Objective

Identify concurrency, timing, resource, and ordering hazards that produce or may produce brittle harness behavior.

### Inputs to inspect

- Parallel test configuration.
- Worker pool settings.
- Shared fixtures and services.
- Temp path and port allocation.
- Locks, semaphores, queues, leases, or absence of them.
- Sleeps, polls, retries, timeouts, debounces, and watch triggers.
- Flaky-test suppressions.
- CI resource limits and sharding configuration.

### Concrete agent actions

- Identify all concurrency points.
- Identify all shared mutable resources.
- Record all fixed sleeps, readiness polls, retries, and timeouts.
- Record defaults, units, bounds, and failure behavior for timeouts and retries.
- Record allocation strategies for ports, temp directories, databases, files, browser profiles, containers, and worker IDs.
- Classify each hazard as confirmed failure, plausible latent failure, accepted nondeterminism, main-spec conflict, or unknown.

### Expected outputs

- `TODO: race_timing_resource_register`
- `TODO: concurrency_model_notes`

### Validation criteria

- Every shared mutable resource appears in hazard analysis.
- Every fixed sleep is recorded.
- Every timeout has default, unit, bound, and result or `TODO:`.
- Every recurring failure maps to at least one hazard or ambiguity.

### Completion checklist

- [ ] Parallelism points recorded.
- [ ] Shared mutable resources recorded.
- [ ] Fixed sleeps recorded.
- [ ] Poll loops recorded.
- [ ] Timeouts recorded.
- [ ] Retry and backoff rules recorded.
- [ ] Port allocation recorded.
- [ ] Temp path allocation recorded.
- [ ] Locking or lease behavior recorded.
- [ ] Race hazards classified.

### Risks or ambiguity to record

- Rare races may not reproduce during recovery.
- Test framework defaults may hide scheduling behavior.
- CI machine differences may change timing.
- External services may introduce nondeterminism outside harness control.

## Stage 7: Failure-mode recovery

### Objective

Document current and plausible harness failures so the future spec can make them deterministic, diagnosable, and recoverable.

### Inputs to inspect

- Failing CI logs.
- Local failure logs.
- Issue reports.
- Flaky-test suppressions.
- Retry and rerun logic.
- Known-failure lists.
- Error handling code.
- Cleanup-after-failure code.
- Timeout and cancellation behavior.
- Partial artifact behavior.

### Concrete agent actions

- Build `TODO: failure_mode_register`.
- Separate product assertion failures from harness operational failures.
- Separate setup failures, service failures, fixture failures, cleanup failures, missing-secret failures, unsupported-platform failures, timeout, deadlock, and resource exhaustion.
- Record trigger, observable result, exit/report behavior, side effects, cleanup behavior, retryability, and owner.
- Record whether failure artifacts are sufficient for reproduction or diagnosis.

### Expected outputs

- `TODO: failure_mode_register`
- `TODO: failure_class_taxonomy`

### Validation criteria

- Every recurring problem named by the operator appears in the register.
- Every failure class has deterministic desired behavior or `TODO: maintainer_decision_required`.
- Operational failures are not conflated with product validation failures.
- Retryability is explicit.

### Completion checklist

- [ ] Race failures recorded.
- [ ] Timing failures recorded.
- [ ] Resource conflicts recorded.
- [ ] Service startup failures recorded.
- [ ] Fixture errors recorded.
- [ ] Cleanup failures recorded.
- [ ] Timeout and cancellation behavior recorded.
- [ ] Missing secret and platform failures recorded.
- [ ] Assertion failures separated from harness failures.
- [ ] Retryability recorded.

### Risks or ambiguity to record

- Current logs may not identify the true failure phase.
- Retried failures may erase original evidence.
- Cleanup failures may mask primary failures.
- Some failures may belong to the product under test.

## Stage 8: Intent and authority classification

### Objective

Separate behavior that should become a harness contract from behavior that exists because of implementation history.

### Inputs to inspect

- Main project specification.
- Existing harness docs.
- Test comments.
- Commit history if available.
- CI configuration.
- Operator notes.
- Implementation evidence.
- Failure and hazard registers.
- Fixture history.

### Concrete agent actions

- Classify each significant behavior as `preserve`, `preserve_with_clarification`, `preserve_compatibility_only`, `refactor_preserving_behavior`, `deprecate`, `redesign_required`, `remove_if_unused`, `exclude_from_contract`, or `authority_decision_required`.
- Identify accidental complexity such as duplicated setup, hidden sequencing, implicit globals, brittle sleeps, shared temp paths, non-isolated services, and implicit fixture mutation.
- Identify behavior externally depended on despite appearing accidental.
- Record conflicts between main spec, harness behavior, tests, fixtures, and CI.
- Draft the harness authority map.

### Expected outputs

- `TODO: preservation_matrix`
- `TODO: ambiguity_register`
- `TODO: harness_authority_map`
- `TODO: main_spec_conflict_list`

### Validation criteria

- Every major subsystem has a classification.
- Every preserved behavior has evidence or pending owner decision.
- Every compatibility-only behavior has a named compatibility risk.
- Every authority-required behavior has a specific owner question.

### Completion checklist

- [ ] Entrypoint behavior classified.
- [ ] Fixture behavior classified.
- [ ] Service lifecycle classified.
- [ ] Temp artifact behavior classified.
- [ ] Cleanup behavior classified.
- [ ] Logging and report behavior classified.
- [ ] Parallelism behavior classified.
- [ ] Retry behavior classified.
- [ ] CI behavior classified.
- [ ] Authority conflicts recorded.

### Risks or ambiguity to record

- Maintainers may disagree about whether current behavior is intended.
- Tests may encode accidental behavior.
- Compatibility obligations may exist outside the repository.
- The main spec may intentionally omit harness mechanics.

## Stage 9: Harness NLSpec contract pass

### Objective

Create the first complete contract-level draft of the dedicated harness specification.

### Inputs to inspect

- All recovery artifacts from stages 0 through 8.
- Main project specification.
- Current harness implementation evidence.
- Ambiguity, hazard, and failure registers.
- Preservation matrix.

### Concrete agent actions

- Copy `templates/harness-nlspec-template.md` to `harness-nlspec.md` when starting from the template; for S7 the draft now exists at that path.
- Write purpose, scope, non-goals, terminology, authority relationship, and harness-owned surfaces.
- Define entrypoint contracts, lifecycle states, artifact ownership, service lifecycle, resource policy, timing behavior, failure taxonomy, cleanup behavior, and diagnostics.
- Mark observed behavior, intended behavior, and open decisions separately.
- Use `TODO:` for unresolved paths, owners, schemas, defaults, bounds, or compatibility obligations.

### Expected outputs

- `harness-nlspec.md`
- Updated `TODO: ambiguity_register`

### Validation criteria

- The draft specifies behavior over mechanism unless mechanism is contract-bearing.
- The draft distinguishes observed behavior from normative desired behavior.
- The draft avoids vague wording such as “handle appropriately.”
- Every unresolved area is marked `TODO:` or `maintainer_decision_required`.

### Completion checklist

- [ ] Purpose written.
- [ ] Scope written.
- [ ] Non-goals written.
- [ ] Authority relationship written.
- [ ] Terms defined.
- [ ] Harness surfaces defined.
- [ ] Entrypoints specified.
- [ ] Lifecycle specified.
- [ ] Artifact ownership specified.
- [ ] Service lifecycle specified.
- [ ] Resource policy specified.
- [ ] Failure taxonomy specified.
- [ ] Cleanup specified.
- [ ] Acceptance criteria stubbed.

### Risks or ambiguity to record

- The first draft may overfit current implementation.
- Desired behavior may require maintainer authority.
- Some current behavior may need main-spec changes.
- Compatibility may require preserving accidental behavior temporarily.

## Stage 10: Mechanics pass

### Objective

Make the harness spec implementable by defining schemas, algorithms, defaults, bounds, errors, ordering, and omitted cases.

### Inputs to inspect

- Harness NLSpec draft.
- Entrypoint map.
- Artifact ownership matrix.
- Service lifecycle map.
- Race and resource register.
- Failure-mode register.

### Concrete agent actions

- Add explicit defaults, bounds, units, omitted-case behavior, and failure behavior.
- Add deterministic algorithms for resource allocation, run workspace creation, service readiness, cleanup, retry, and report ordering where needed.
- Add schemas for machine-consumed outputs and harness-owned records.
- Add mapping tables for commands, artifacts, failures, services, cleanup triggers, resource types, fixture types, and CI jobs.
- Mark current-risk behavior where the implementation is nondeterministic or unsafe.

### Expected outputs

- Updated `harness-nlspec.md`
- `TODO: mechanics_gap_list`

### Validation criteria

- Every optional input has omitted-case behavior.
- Every timeout has unit, default, maximum, and timeout result.
- Every resource allocation rule is conflict-safe or marked as current risk.
- Every error class maps to caller-visible behavior.

### Completion checklist

- [ ] Defaults specified.
- [ ] Bounds specified.
- [ ] Omitted cases specified.
- [ ] Command schemas specified.
- [ ] Output schemas specified.
- [ ] Resource allocation specified.
- [ ] Timeout behavior specified.
- [ ] Retry behavior specified.
- [ ] Cleanup algorithms specified.
- [ ] Failure mappings specified.
- [ ] Ordering rules specified.

### Risks or ambiguity to record

- Current harness may not have deterministic algorithms.
- Some third-party framework defaults may be difficult to specify.
- Strictly specifying fragile behavior may make later repairs harder.

## Stage 11: Verification pass

### Objective

Bind every normative harness requirement to binary acceptance criteria.

### Inputs to inspect

- Mechanics-pass harness spec.
- Existing tests.
- CI workflows.
- Fixture directories.
- Failure registers.
- Main project acceptance criteria.

### Concrete agent actions

- Extract every `must` and `must not` from the draft.
- Define how each requirement can be observed.
- Create acceptance criteria with stable IDs.
- Mark criteria as `existing_test`, `new_test_needed`, `manual_review`, `golden_fixture`, `static_inspection`, or `owner_decision`.
- Identify missing fixtures and golden files.
- Identify future CI gates.
- Rewrite untestable requirements or mark them owner-decision-required.

### Expected outputs

- `harness-acceptance-matrix.md`
- `TODO: missing_fixture_list`
- `TODO: future_ci_gate_list`
- Updated `harness-nlspec.md`

### Validation criteria

- Every normative requirement has an acceptance criterion or owner decision.
- Every acceptance criterion is binary.
- Every acceptance criterion has pass and fail conditions.
- Every missing fixture is named.

### Completion checklist

- [ ] Requirements extracted.
- [ ] Acceptance IDs assigned.
- [ ] Verification methods assigned.
- [ ] Fixtures mapped.
- [ ] Existing tests mapped.
- [ ] Missing tests listed.
- [ ] CI gaps listed.
- [ ] Golden files listed.
- [ ] Untestable requirements removed or rewritten.

### Risks or ambiguity to record

- Existing tests may validate product behavior but not harness behavior.
- CI may not expose enough artifacts to verify lifecycle.
- Race checks may need repeated-run evidence.

## Stage 12: Roadmap and review packet

### Objective

Prepare the recovered harness specification and artifacts for maintainer review, adoption, and later implementation remediation.

### Inputs to inspect

- Harness NLSpec draft.
- Recovery artifacts.
- Acceptance matrix.
- Ambiguity register.
- Main-spec conflict list.
- Preservation matrix.

### Concrete agent actions

- Classify future work as `preserve`, `preserve_with_clarification`, `refactor_preserving_behavior`, `deprecate`, `redesign_required`, `remove_if_unused`, or `authority_decision_required`.
- Produce `harness-implementation-roadmap.md`.
- Perform consistency pass across recovery artifacts.
- Remove duplicate definitions and drift-prone repetition.
- Verify every table covers its declared scope.
- Consolidate open decisions for maintainers.
- Produce a final recovery review packet.

### Expected outputs

- `harness-implementation-roadmap.md`
- `harness-review-packet.md`
- Final `harness-nlspec.md`
- Final `harness-acceptance-matrix.md`

### Validation criteria

- No implementation rewrite is performed.
- Every roadmap item references recovered evidence.
- Every authority decision has a specific decision prompt.
- The review packet can be read without raw recovery notes.
- Open issues are explicit.

### Completion checklist

- [ ] Preserve items listed.
- [ ] Refactor candidates listed.
- [ ] Deprecation candidates listed.
- [ ] Redesign candidates listed.
- [ ] Removal candidates listed.
- [ ] Authority decisions listed.
- [ ] Main-spec impact recorded.
- [ ] Duplicate definitions removed.
- [ ] Tables checked for declared scope.
- [ ] Review packet produced.
- [ ] No implementation changes made.

### Risks or ambiguity to record

- A complete spec may be blocked until maintainers make authority decisions.
- Some current behavior may be too inconsistent to specify without redesign.
- Deprecating accidental behavior may break external workflows.
