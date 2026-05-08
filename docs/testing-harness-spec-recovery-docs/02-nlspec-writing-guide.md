---
doc_id: THR-020
title: Testing Harness NLSpec Writing Guide
status: draft
role: writing-guide
---

# Testing Harness NLSpec Writing Guide

## Document role

This guide explains how agents should convert recovered harness behavior into specification material.

Use this guide after the recovery process has produced enough evidence to draft or revise the dedicated harness NLSpec.

## Writing objective

The harness NLSpec must define the observable contract of the testing harness. A future implementer must be able to recreate caller-visible harness behavior from the specification without guessing about commands, defaults, sequencing, resources, artifacts, errors, cleanup, diagnostics, or acceptance criteria.

The harness NLSpec must not merely describe the current implementation. It must separate observed behavior, intended behavior, compatibility-only behavior, accidental behavior, contradictions, missing behavior, and redesign candidates.

## Required writing posture

- Use `must`, `must not`, `may`, and `default` for contract-bearing behavior.
- Avoid vague terms such as “appropriate,” “reasonable,” “best effort,” “robust,” and “as needed” unless they are explicitly defined.
- Specify observable behavior over internal mechanism unless mechanism affects correctness, interoperability, timing, persistence, security, reproducibility, or cleanup.
- Define defaults, bounds, omitted cases, error behavior, ordering, normalization, IDs, serialization, versioning, and compatibility when they affect observable behavior.
- Use exhaustive tables for mappings across commands, services, failures, artifacts, states, outputs, CI jobs, and resource classes.
- Use deterministic pseudocode for nontrivial lifecycle, allocation, retry, cleanup, and report-ordering behavior.
- Define each fact once. Refer to the definition instead of repeating it.
- Mark intentional ambiguity explicitly.
- Mark accidental ambiguity, gaps, contradictions, and owner decisions instead of filling them with guesses.

## Evidence-to-spec conversion table

| Recovery evidence | Specification treatment |
|---|---|
| Exact command declaration | Entrypoint contract row with command, caller, mode, inputs, defaults, outputs, side effects, failure behavior, and acceptance criteria. |
| Script or runner setup code | Lifecycle phase, ordering rule, service dependency, resource policy, or implementation note if not contract-bearing. |
| Test fixture or golden file | Artifact authority row and fixture update rule. |
| Generated report | Derived diagnostic surface unless execution consumes it as authority. |
| CI job | Operational gate row and command mapping. |
| Global setup or teardown | Lifecycle transition and cleanup rule. |
| Fixed sleep | Timing assumption and hazard row. Do not convert to desired readiness rule without owner decision. |
| Retry loop | Retry algorithm with bound, delay, eligible failures, reset behavior, and exhaustion result. |
| Port or temp path allocation | Resource allocation algorithm with collision behavior. |
| Failure log | Failure taxonomy row and observable failure contract. |
| Flaky test rerun | Retry policy and nondeterminism boundary. |
| Maintainer decision | Normative requirement only after decision is recorded. |
| Contradiction between docs and code | Ambiguity register entry. Do not assert either side as intended until resolved. |

## Contract pass

### Objective

Model the harness domain and caller-visible behavior before writing low-level mechanics.

### Required sections

| Section | Required content |
|---|---|
| Purpose and scope | What the harness owns and why it has a dedicated spec. |
| Non-goals | What the harness spec does not define, especially product behavior owned by the main spec. |
| Authority relationship | Precedence and ownership across main spec, harness spec, implementation, tests, fixtures, CI, logs, reports, and local policy. |
| Terminology | Stable names for harness run, entrypoint, fixture, generated artifact, service, worker, run workspace, cleanup, failure class, and resource lease. |
| Actors | Developer, CI, release gate, local agent, test runner, service broker, cleanup owner, report consumer. |
| Harness-owned surfaces | Commands, scripts, config, fixtures, services, generated artifacts, temp paths, logs, reports, cleanup, and resource policy. |
| Entrypoints | Exact command contracts and mode distinctions. |
| Run lifecycle | Phases from preflight to teardown and terminal states. |
| Artifact lifecycle | Authority, ownership, mutation, persistence, and cleanup. |
| Service lifecycle | Provisioning, readiness, use, reset, stop, cleanup, and failure behavior. |
| Resource and concurrency model | Worker, port, temp path, lock, process, container, and shared-state rules. |
| Timing, retry, cancellation | Timeouts, sleeps, polling, retry bounds, cancellation, interrupts, crash behavior. |
| Failure taxonomy | Setup failure, assertion failure, harness failure, service failure, fixture failure, timeout, cancellation, resource conflict, missing secret, unsupported platform. |
| Diagnostics | Logs, reports, dashboards, screenshots, traces, failure bundles, and authority status. |
| Acceptance criteria | Binary verification matrix for every normative requirement. |
| Roadmap | Preserve, refactor, deprecate, redesign, remove-if-unused, and authority-decision-required items. |

### Contract writing pattern

Use this structure for every major behavior.

```markdown
### <Behavior name>

The harness must <observable requirement> when <scope condition>.

Inputs:
- <input 1>
- <input 2>

Outputs:
- <output 1>
- <output 2>

Side effects:
- <side effect 1>

Errors:
- <error class>: <caller-visible behavior>

Acceptance:
- <criterion id>: <pass condition>
```

## Mechanics pass

### Objective

Close every implementation-significant gap in the contract pass.

### Mechanics that must be explicit

| Mechanics surface | Required specification |
|---|---|
| Command inputs | Flags, env vars, positional args, config files, omitted-case behavior, invalid input behavior. |
| Exit behavior | Exit code meanings, stdout/stderr semantics, partial-output behavior, CI annotation behavior. |
| Report schemas | Field names, types, required/default/conditional fields, ordering, null/omission semantics. |
| Artifact paths | Canonical path rules, run-local paths, collision behavior, cleanup ownership, persistence. |
| Resource allocation | Worker IDs, ports, temp dirs, database names, browser profiles, locks, leases, allocation failure behavior. |
| Service readiness | Health checks, poll interval, timeout, failure result, stale service handling. |
| Retry and backoff | Eligible failures, maximum attempts, delay algorithm, reset behavior, exhaustion result. |
| Timeout and cancellation | Units, defaults, minimums, maximums, signal/shutdown sequence, partial artifacts, cleanup. |
| Fixture updates | Who may update, command required, review requirements, generated-vs-canonical distinction. |
| Cleanup | Trigger, order, idempotence, failure behavior, best-effort versus fail-closed semantics. |
| Ordering | Test discovery ordering, report ordering, log ordering, artifact enumeration, parallel fan-in ordering. |
| Compatibility | Behaviors preserved only for existing callers and conditions for future deprecation. |

### Required mapping tables

Include a table whenever more than two cases exist. At minimum, include:

- Entrypoint to lifecycle mapping.
- Entrypoint to generated artifacts.
- Command mode to allowed side effects.
- Artifact authority class to execution effect.
- Fixture type to update authority.
- Service state to recovery action.
- Cleanup trigger to cleanup action.
- Failure class to exit/report behavior.
- Resource type to allocation rule.
- Timeout type to default, maximum, and failure result.
- CI job to harness entrypoint.

### Algorithm format

Use deterministic pseudocode for nontrivial behavior.

```text
Algorithm: <algorithm name>

Input:
  <input name and type>

Preconditions:
  <required state>

Steps:
  1. <deterministic step>
  2. <deterministic step>
  3. <deterministic step>

Output:
  <output value or artifact>

Errors:
  - <error class>: <observable behavior>

Acceptance:
  - <criterion id>: <binary check>
```

Do not use pseudocode when declarative prose fully determines behavior.

## Failure taxonomy writing rules

The harness spec must classify failures by phase and caller-visible behavior.

| Failure class | Required meaning |
|---|---|
| `usage_error` | Caller passed invalid command, flag, path, env, mode, or unsupported combination. |
| `configuration_error` | Required harness configuration is missing, invalid, or contradictory. |
| `preflight_error` | Harness cannot start because dependencies, secrets, platform, or prerequisites are missing. |
| `service_start_error` | Required service cannot be provisioned or started. |
| `service_readiness_timeout` | Service started but never satisfied the ready condition. |
| `fixture_error` | Fixture is missing, malformed, stale, or unauthorized to mutate. |
| `resource_conflict` | Port, temp path, lock, database, browser profile, worker slot, or cache conflicts. |
| `test_assertion_failure` | Product-under-test assertion failed. |
| `harness_internal_error` | Harness code failed independently of product assertion. |
| `timeout` | Configured timeout elapsed. |
| `cancelled` | Caller or CI cancelled execution. |
| `cleanup_error` | Cleanup failed or left ambiguous state. |
| `unsupported_platform` | Current platform cannot satisfy the declared harness contract. |
| `missing_secret` | Required secret or credential is unavailable. |
| `unknown_failure` | Failure observed but not yet classified. This must not remain in the final accepted spec. |

Each failure class must map to exit behavior, report behavior, cleanup behavior, retryability, and owner.

## Verification pass

### Objective

Bind every normative requirement to binary acceptance criteria.

### Acceptance criterion fields

| Field | Required content |
|---|---|
| `criterion_id` | Stable ID such as `HARNESS-AC-001`. |
| `requirement` | One binary requirement. |
| `source_section` | Harness spec section that owns the requirement. |
| `verification_method` | Static inspection, unit test, integration test, CI run, golden fixture, manual review, or owner decision. |
| `fixture_required` | Yes/no and path or `TODO:`. |
| `pass_condition` | Exact condition for pass. |
| `fail_condition` | Exact condition for fail. |
| `current_coverage` | Existing, partial, missing, unknown, or owner decision. |
| `notes` | Source limits or future work. |

### Acceptance quality rules

- One criterion must test one requirement.
- Vague criteria must be rewritten.
- “Works,” “reasonable,” “appropriate,” and “robust” are invalid pass conditions unless defined by a measurable predicate.
- Criteria that require maintainer judgment must be labeled `owner_decision`.
- Existing tests may be cited as current coverage only after the agent inspects them.
- Passing an existing test is not proof that the requirement is sufficient.

## Authority writing rules

The harness specification must not redefine product behavior owned by the main project specification. It may define how the harness validates product behavior.

Use this pattern:

```markdown
The main project specification owns <product behavior>. This harness specification owns <validation behavior>. If a harness fixture or test encodes product behavior not present in the main project specification, the recovery agent must record an authority gap and must not promote that fixture behavior to harness normativity by itself.
```

## Preservation and roadmap writing rules

Future implementation work must be classified separately from recovered specification work.

| Classification | Meaning |
|---|---|
| `preserve` | Make current behavior part of the harness contract. |
| `preserve_with_clarification` | Keep behavior but specify missing defaults, errors, or boundaries. |
| `refactor_preserving_behavior` | Mechanism may change, observable behavior must not. |
| `deprecate` | Keep temporarily with warning, replacement path, and removal gate. |
| `redesign_required` | Current behavior is unsafe, inconsistent, or not specifiable as-is. |
| `remove_if_unused` | Appears unused and non-contractual. Prove no dependency before removal. |
| `authority_decision_required` | Human decision required before normative closure. |
| `exclude_from_contract` | Behavior is incidental and not relied on. State non-guarantee if reliance risk exists. |

Do not write roadmap items as if they have already been implemented.

## Spec closure checklist

Before presenting the harness NLSpec for review, verify:

- [ ] Every command has defaults, inputs, outputs, side effects, and failure behavior.
- [ ] Every generated artifact has ownership, authority, lifecycle, and cleanup behavior.
- [ ] Every service has readiness, reset, stop, and failure behavior.
- [ ] Every timeout has unit, default, maximum, and timeout result.
- [ ] Every retry has eligible failures, bound, delay algorithm, and exhaustion behavior.
- [ ] Every shared resource has allocation and collision behavior.
- [ ] Every machine-consumed output has a schema or explicit source limit.
- [ ] Every normative requirement has a binary acceptance criterion.
- [ ] Every contradiction has a named owner or decision prompt.
- [ ] Every `TODO:` is either intentionally left for owner review or resolved.
- [ ] No derived diagnostic view is silently promoted to authority.
- [ ] No product behavior is redefined under the harness spec.
