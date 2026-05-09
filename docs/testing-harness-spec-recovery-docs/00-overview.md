---
doc_id: THR-000
title: Testing Harness Reverse-Specification Recovery Overview
status: draft
role: overview
---

# Testing Harness Reverse-Specification Recovery Overview

## Document role

This document explains the purpose, scope, authority posture, and expected outputs of a reverse-specification recovery effort for an existing testing harness.

Agents should read this document before inspecting repository files or drafting the harness specification.

## Purpose

The project already has a testing harness whose behavior evolved through implementation work rather than through a dedicated specification. The harness is now its own subsystem. It has commands, services, fixtures, generated artifacts, logs, cleanup behavior, resource policy, sequencing assumptions, and failure modes that need a first-class specification.

This recovery effort must produce a dedicated NLSpec-style testing harness specification grounded in current implementation behavior. The output must help maintainers understand what the harness currently does, what behavior is intentional, what behavior is accidental, what must be preserved for compatibility, and what should be refactored or redesigned later.

## Scope

The recovery scope includes all repository surfaces that affect test orchestration or validation behavior.

| Surface | In scope |
|---|---|
| Entrypoints | Local commands, CI commands, package scripts, task-runner targets, pre-commit hooks, release validation commands. |
| Orchestration | Test runners, setup hooks, teardown hooks, service launchers, retry logic, sharding, watch mode, cleanup scripts. |
| Fixtures | Test data, golden files, snapshots, seeds, sample projects, mocks, and generated fixture updates. |
| Services | Databases, containers, browsers, emulators, mock servers, queues, object stores, and local app instances. |
| Artifacts | Reports, coverage, screenshots, traces, logs, temp files, caches, failure bundles, and uploaded CI artifacts. |
| Runtime policy | Timeouts, retries, parallelism, resource allocation, port selection, temp path selection, locks, leases, environment variables, secrets, and platform assumptions. |
| Failure behavior | Setup failures, service failures, fixture failures, test assertion failures, resource conflicts, timeouts, cancellation, cleanup failures, and missing-secret failures. |

## Non-goals

This recovery effort must not rewrite the harness.

It must not change test logic, fixture contents, CI behavior, cleanup scripts, service lifecycle behavior, command semantics, or generated artifact formats. Any future implementation remediation must be tracked as roadmap work outside the recovery process.

## Required posture

The agent must treat current implementation as evidence of behavior, not automatic proof of intended design. The recovered specification must distinguish:

| Classification | Meaning |
|---|---|
| `observed` | Directly inspected current behavior. |
| `inferred` | Behavior derived from multiple observed facts. |
| `assumed` | Temporary assumption pending evidence. |
| `intentional` | Behavior confirmed by governing spec, maintainer decision, or reliable project documentation. |
| `compatibility_only` | Behavior that appears accidental but external callers may depend on. |
| `accidental` | Behavior that exists without clear intent and should not become contract by default. |
| `contradiction` | Sources disagree. Do not pick a side without owner decision. |
| `maintainer_decision_required` | Normative closure requires owner authority. |
| `source_limit` | The agent could not inspect enough to decide. |

## Authority posture

The new harness specification must not redefine product behavior owned by the main project specification. It may define how the harness validates that behavior.

| Surface | Owns | Does not own |
|---|---|---|
| Main project specification | Product behavior, public interfaces, release-level acceptance meaning. | Harness mechanics unless they affect product conformance. |
| Harness specification | Test orchestration, fixture lifecycle, service lifecycle, command behavior, generated artifacts, cleanup, resource policy, and harness failure taxonomy. | Product behavior that belongs to the main project specification. |
| Implementation | Evidence of current behavior. | Future normative truth by itself. |
| Tests and fixtures | Evidence of current validation expectations. | Normative authority when they conflict with active specifications. |
| CI configuration | Operational gate behavior. | Product or harness semantics not explicitly specified. |
| Logs, reports, and dashboards | Diagnostic or derived views by default. | Canonical pass/fail or state authority unless explicitly designated. |
| Local policy | Local execution constraints and secrets. | Portable harness contract unless adopted into the specification. |

## Recovery artifact set

The recovery must produce or update these artifacts in the target repository.

| Artifact | Purpose |
|---|---|
| `TODO: recovery_charter_path` | Scope, allowed writes, prohibited implementation changes, repository revision, and source limits. |
| `TODO: harness_inventory` | File, directory, config, command, fixture, service, artifact, log, and cleanup inventory. |
| `TODO: entrypoint_command_map` | Exact command contracts and invocation paths. |
| `TODO: artifact_ownership_matrix` | Fixture, generated artifact, log, temp file, cache, and external state ownership. |
| `TODO: service_lifecycle_map` | Service provisioning, readiness, reset, stop, and cleanup behavior. |
| `TODO: observable_interface_map` | Exit codes, stdout/stderr, machine reports, logs, CI annotations, and output consumers. |
| `TODO: harness_lifecycle_map` | Lifecycle phases, legal transitions, terminal states, and partial-completion states. |
| `TODO: race_timing_resource_register` | Race, timing, concurrency, allocation, and ordering hazards. |
| `TODO: failure_mode_register` | Failure classes, triggers, outputs, side effects, cleanup, retryability, and ownership. |
| `TODO: ambiguity_register` | Open gaps, contradictions, missing defaults, and owner decisions. |
| `TODO: preservation_matrix` | Preserve, clarify, refactor, deprecate, redesign, remove, and authority-decision classifications. |
| `TODO: harness_authority_map` | Relationship between main spec, harness spec, implementation, tests, fixtures, CI, reports, and local policy. |
| `s0-s6-gap-closure-plan.md` | S0 through S6 gap closure, source-limit carry-forward, authority questions, minimal execution order, and S7 readiness criteria. |
| `TODO: harness_nlspec_draft_path` | Dedicated testing harness NLSpec draft. |
| `TODO: acceptance_matrix` | Binary acceptance criteria for every normative harness requirement. |
| `TODO: harness_roadmap` | Future implementation remediation plan. |
| `TODO: harness_recovery_review_packet` | Maintainer review packet summarizing findings, risks, decisions, and readiness. |

## Definition of done

The recovery is complete only when all conditions below are true.

1. The harness structure has been inventoried.
2. Every harness entrypoint has a command contract or an explicit gap.
3. Every fixture, generated artifact, log, temp file, and cleanup path has an ownership classification or recorded ambiguity.
4. Every service and environment dependency has lifecycle, readiness, reset, and cleanup behavior recorded or marked unknown.
5. Race conditions, timing issues, resource conflicts, ordering assumptions, and brittle coupling points are recorded.
6. Current failure modes have trigger, observable result, retryability, cleanup behavior, and desired spec treatment.
7. Intentional behavior is separated from accidental, emergent, and compatibility-only behavior.
8. The authority relationship between the main project specification and harness specification is explicit.
9. The harness NLSpec draft contains contracts, interfaces, defaults, bounds, failure behavior, lifecycle rules, and acceptance criteria.
10. Every normative requirement maps to a binary acceptance criterion or owner decision.
11. Future work is classified separately from specification recovery.
12. No harness implementation rewrite occurred during recovery.
