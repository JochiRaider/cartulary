---
doc_id: TODO-HARNESS-SPEC
title: TODO Testing Harness Specification
status: draft
role: recovered-nlspec-template
---

# TODO Testing Harness Specification

## Document status

This document is a draft recovered specification for the testing harness. It is not binding until adopted by the project maintainer under the project’s normal specification governance.

## Document role

This document defines the observable behavior, interfaces, defaults, bounds, lifecycle rules, failure behavior, resource policy, cleanup behavior, diagnostics, and acceptance criteria for the testing harness.

Use this document as the implementation-grounded harness contract after recovery evidence has been collected. Do not use this template as evidence of current behavior.

## Purpose and scope

TODO: Define the testing harness purpose.

The harness specification owns TODO: commands, orchestration, fixtures, generated artifacts, services, environment lifecycle, resource allocation, timing, retries, cleanup, diagnostics, and harness acceptance criteria.

## Non-goals

The harness specification does not define TODO: product behavior owned by the main project specification.

The harness specification does not require implementation refactors unless those refactors are explicitly listed as future work.

## Authority relationship

| Surface | Owns | Does not own | Conflict behavior |
|---|---|---|---|
| Main project specification | TODO: product behavior and release-level acceptance meaning | TODO: harness mechanics | TODO: |
| Harness specification | TODO: harness orchestration and validation behavior | TODO: product semantics | TODO: |
| Harness implementation | Evidence of current behavior | Normative truth by itself | TODO: |
| Tests and fixtures | Evidence of current validation expectations | Normative truth when conflicting with active specs | TODO: |
| CI configuration | Operational gate behavior | Product or harness semantics not specified | TODO: |
| Logs, reports, dashboards | Diagnostic or derived views by default | Canonical pass/fail unless explicitly specified | TODO: |
| Local policy | Local execution constraints | Portable harness contract unless adopted | TODO: |

## Terminology

| Term | Meaning |
|---|---|
| Harness | TODO: |
| Harness run | TODO: |
| Entrypoint | TODO: |
| Fixture | TODO: |
| Canonical fixture | TODO: |
| Generated artifact | TODO: |
| Temporary artifact | TODO: |
| Diagnostic log | TODO: |
| Service dependency | TODO: |
| Run workspace | TODO: |
| Cleanup owner | TODO: |
| Resource lease | TODO: |
| Failure class | TODO: |

## Actors and responsibilities

| Actor | Responsibilities | Must not do |
|---|---|---|
| Developer | TODO: | TODO: |
| CI | TODO: | TODO: |
| Release gate | TODO: | TODO: |
| Local coding agent | TODO: | TODO: |
| Test runner | TODO: | TODO: |
| Service broker | TODO: | TODO: |
| Cleanup owner | TODO: | TODO: |
| Report consumer | TODO: | TODO: |

## Harness-owned surfaces

| Surface class | Required meaning | Authority class | Notes |
|---|---|---|---|
| Entrypoints | TODO: | TODO: | TODO: |
| Orchestration scripts | TODO: | TODO: | TODO: |
| Fixtures | TODO: | TODO: | TODO: |
| Generated artifacts | TODO: | TODO: | TODO: |
| Temporary artifacts | TODO: | TODO: | TODO: |
| Logs | TODO: | TODO: | TODO: |
| Reports | TODO: | TODO: | TODO: |
| Services | TODO: | TODO: | TODO: |
| Resource policy | TODO: | TODO: | TODO: |
| Cleanup | TODO: | TODO: | TODO: |

## Entrypoints and command contracts

### Entrypoint contract table

| Entrypoint ID | Command | Caller | Mode | Inputs | Defaults | Outputs | Side effects | Failure behavior | Acceptance criteria |
|---|---|---|---|---|---|---|---|---|---|
| EP-0001 | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Command contract rules

TODO: Define command-level invariants.

## Run lifecycle

### Lifecycle states

| State | Meaning | Allowed next states | Required artifacts | Notes |
|---|---|---|---|---|
| `TODO` | TODO | TODO | TODO | TODO |

### Lifecycle phases

| Phase order | Phase | Owner | Inputs | Outputs | Side effects | Failure behavior | Cleanup behavior |
|---:|---|---|---|---|---|---|---|
| 1 | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Lifecycle algorithm

```text
Algorithm: TODO_harness_run_lifecycle

Input:
  TODO

Preconditions:
  TODO

Steps:
  1. TODO

Terminal states:
  - TODO

Errors:
  - TODO
```

## Fixture and artifact lifecycle

### Artifact authority classes

| Authority class | Required meaning | May drive execution | Update authority |
|---|---|---|---|
| `canonical_fixture` | TODO | TODO | TODO |
| `canonical_runtime_state` | TODO | TODO | TODO |
| `derived_report` | TODO | TODO | TODO |
| `ephemeral_scratch` | TODO | TODO | TODO |
| `diagnostic_log` | TODO | TODO | TODO |
| `external_state` | TODO | TODO | TODO |

### Artifact lifecycle table

| Artifact class | Created by | Read by | Mutated by | Persisted where | Cleanup owner | Cleanup trigger | Absence/staleness behavior |
|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

## Service and environment lifecycle

| Service/environment | Provision owner | Ready condition | Scope | Shared resources | Reset rule | Stop rule | Failure behavior |
|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

## Resource allocation and concurrency

### Resource allocation table

| Resource type | Allocation rule | Collision detection | Collision behavior | Release rule | Parallel-safe |
|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO |

### Concurrency model

TODO: Define parallelism, worker ownership, shared mutable resources, and legal concurrent invocations.

## Timing, retry, timeout, and cancellation

| Timing surface | Unit | Default | Minimum | Maximum | Failure result | Notes |
|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO |

### Retry algorithm

```text
Algorithm: TODO_retry_policy

Input:
  failure_class
  attempt_number
  entrypoint_id

Steps:
  1. TODO

Exhaustion result:
  TODO
```

### Cancellation behavior

TODO: Define cancellation on caller interrupt, CI cancellation, timeout, process crash, and service teardown failure.

## Failure taxonomy and error behavior

| Failure class | Trigger | Caller-visible result | Exit/report behavior | Cleanup behavior | Retryable | Owner |
|---|---|---|---|---|---|---|
| `usage_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `configuration_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `preflight_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `service_start_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `service_readiness_timeout` | TODO | TODO | TODO | TODO | TODO | TODO |
| `fixture_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `resource_conflict` | TODO | TODO | TODO | TODO | TODO | TODO |
| `test_assertion_failure` | TODO | TODO | TODO | TODO | TODO | TODO |
| `harness_internal_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `timeout` | TODO | TODO | TODO | TODO | TODO | TODO |
| `cancelled` | TODO | TODO | TODO | TODO | TODO | TODO |
| `cleanup_error` | TODO | TODO | TODO | TODO | TODO | TODO |
| `unsupported_platform` | TODO | TODO | TODO | TODO | TODO | TODO |
| `missing_secret` | TODO | TODO | TODO | TODO | TODO | TODO |

## Cleanup and recovery behavior

| Cleanup trigger | Cleanup action | Required order | Idempotence | Failure behavior | Acceptance criteria |
|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO |

## Logs, reports, and diagnostic surfaces

| Diagnostic surface | Produced by | Consumer | Machine-consumed | Authority class | Schema or format | Retention | Notes |
|---|---|---|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO | TODO | TODO | TODO |

Derived diagnostic surfaces must not drive execution unless this section explicitly designates them as canonical for a named purpose.

## Main-spec validation relationship

TODO: State which product requirements the harness validates and which main-spec surfaces own those requirements.

| Product requirement or surface | Main-spec owner | Harness validation surface | Harness-owned behavior | Authority gap |
|---|---|---|---|---|
| TODO | TODO | TODO | TODO | TODO |

## Intent classification and roadmap

| Behavior | Classification | Current treatment | Future action | Owner decision required |
|---|---|---|---|---|
| TODO | `preserve/preserve_with_clarification/refactor_preserving_behavior/deprecate/redesign_required/remove_if_unused/authority_decision_required/exclude_from_contract` | TODO | TODO | TODO |

## Acceptance criteria

| Criterion ID | Requirement | Verification method | Pass condition | Fail condition | Current coverage |
|---|---|---|---|---|---|
| HARNESS-AC-001 | TODO | TODO | TODO | TODO | TODO |

## Open ambiguities and owner decisions

| ID | Surface | Open issue | Required owner | Blocking effect | Notes |
|---|---|---|---|---|---|
| AMB-0001 | TODO | TODO | TODO | TODO | TODO |

## Source limits

| ID | Surface | Limit | Impact | Follow-up |
|---|---|---|---|---|
| SL-0001 | TODO | TODO | TODO | TODO |
