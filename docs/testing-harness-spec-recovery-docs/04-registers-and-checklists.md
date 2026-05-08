---
doc_id: THR-040
title: Testing Harness Recovery Registers and Checklists
status: draft
role: registers-and-checklists
---

# Testing Harness Recovery Registers and Checklists

## Document role

This document provides working checklists and register templates for the recovery effort. Agents should copy or update these tables in the target repository and use them as source material for the final harness NLSpec.

## Use rules

- Do not delete unresolved rows. Mark them `TODO:` or `source_limit`.
- Add evidence references to exact files, sections, commands, or run outputs.
- Keep one row per observable behavior, artifact class, service, failure mode, hazard, or decision.
- Do not merge contradictory evidence into one reconciled statement. Add a contradiction row.
- Mark every row with an evidence status.

## Evidence status values

| Status | Meaning |
|---|---|
| `observed` | Directly inspected in repository source, config, docs, tests, fixtures, CI, or committed artifacts. |
| `runtime_observed` | Observed by running a command during recovery. |
| `inferred` | Derived from multiple observed sources. |
| `assumed` | Temporary assumption pending evidence. |
| `contradiction` | Sources disagree and owner decision is required. |
| `maintainer_decision_required` | Behavior requires human authority. |
| `source_limit` | Agent could not inspect enough to decide. |

## Recovery progress checklist

| Area | Complete | Notes |
|---|---:|---|
| Recovery charter exists | [ ] | `TODO:` |
| Repository revision and dirty state recorded | [ ] | `TODO:` |
| Harness boundary candidate list created | [ ] | `TODO:` |
| Harness inventory complete | [ ] | `TODO:` |
| Entrypoint map complete | [ ] | `TODO:` |
| Artifact ownership matrix complete | [ ] | `TODO:` |
| Cleanup lifecycle matrix complete | [ ] | `TODO:` |
| Service lifecycle map complete | [ ] | `TODO:` |
| Resource allocation register complete | [ ] | `TODO:` |
| Observable interface map complete | [ ] | `TODO:` |
| Lifecycle map complete | [ ] | `TODO:` |
| Race/timing/resource register complete | [ ] | `TODO:` |
| Failure-mode register complete | [ ] | `TODO:` |
| Ambiguity register complete | [ ] | `TODO:` |
| Authority map complete | [ ] | `TODO:` |
| Preservation matrix complete | [ ] | `TODO:` |
| Harness NLSpec draft created | [ ] | `TODO:` |
| Acceptance matrix complete | [ ] | `TODO:` |
| Roadmap complete | [ ] | `TODO:` |
| Review packet complete | [ ] | `TODO:` |
| No implementation changes made | [ ] | `TODO:` |

## Harness inventory template

| Inventory ID | Path or surface | Role | Read/write/generated status | Committed/ignored/external | Owner hypothesis | Invoked by | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| HI-0001 | `TODO:` | `entrypoint/orchestration/fixture/service/generated_artifact/temporary_artifact/log/cleanup/policy/adapter/derived_view` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Entrypoint command map template

| Entrypoint ID | Declared command | Caller | Mode | Inputs | Defaults/omitted cases | Outputs | Side effects | Ordering dependencies | Parallel safety | Failure behavior | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EP-0001 | `TODO:` | `developer/ci/precommit/release/agent/unknown` | `test/watch/fixture_update/service_start/service_stop/cleanup/debug/coverage/unknown` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Artifact ownership matrix template

| Artifact ID | Artifact class/path | Authority class | Created by | Read by | Mutated by | Persisted where | Committed/ignored/uploaded | Cleanup owner | Cleanup trigger | Absence/staleness effect | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-0001 | `TODO:` | `canonical_fixture/canonical_runtime_state/derived_report/ephemeral_scratch/diagnostic_log/external_state/unknown_authority` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Cleanup lifecycle matrix template

| Cleanup ID | Cleanup surface | Trigger | Scope | Order | Idempotent | Runs on success | Runs on failure | Runs on timeout | Runs on interrupt | Failure behavior | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| CLN-0001 | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Service lifecycle map template

| Service ID | Service/environment | Provision owner | Start command/path | Ready condition | Ready timeout | Scope | Shared resources | Reset rule | Stop rule | Cleanup rule | Failure behavior | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| SVC-0001 | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `per_test/per_file/per_worker/per_suite/per_ci_job/global/unknown` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Resource allocation register template

| Resource ID | Resource type | Allocation rule | Scope | Collision detection | Collision behavior | Release rule | Reuse allowed | Parallel-safe | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| RES-0001 | `port/temp_dir/database/browser_profile/worker/process/container/cache/file_lock/socket/other` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Observable interface map template

| Interface ID | Entrypoint ID | Output surface | Consumer | Machine-consumed | Schema/path | Ordering guarantee | Stable across CI/local | Authority class | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| OI-0001 | `TODO:` | `stdout/stderr/exit_code/report/log/coverage/ci_annotation/failure_bundle/dashboard/other` | `TODO:` | `yes/no/unknown` | `TODO:` | `TODO:` | `TODO:` | `canonical/derived/diagnostic/unknown` | `TODO:` | `TODO:` | `TODO:` |

## Lifecycle phase table template

| Lifecycle ID | Entrypoint ID | Phase order | Phase name | Owner | Inputs | Outputs | Side effects | Skipped when | Failure behavior | Cleanup behavior | Evidence | Evidence status | Notes |
|---|---|---:|---|---|---|---|---|---|---|---|---|---|---|
| LIFE-0001 | `TODO:` | 1 | `preflight/setup/service_start/fixture_stage/discovery/execution/report/cleanup/teardown/other` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Race, timing, and resource hazard register template

| Hazard ID | Surface | Trigger | Shared resource | Timing assumption | Observable failure | Current mitigation | Spec gap | Severity | Evidence | Evidence status | Proposed disposition | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| HAZ-0001 | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `low/medium/high/critical/unknown` | `TODO:` | `TODO:` | `confirmed_failure/plausible_latent_failure/accepted_nondeterminism/main_spec_conflict/unknown` | `TODO:` |

## Timeout and retry register template

| Rule ID | Surface | Timeout/retry type | Unit | Default | Minimum | Maximum | Eligible failure classes | Algorithm | Exhaustion behavior | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| TMR-0001 | `TODO:` | `timeout/retry/backoff/poll/debounce/watch` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Failure-mode register template

| Failure ID | Failure class | Phase | Trigger | Observable result | Exit/report behavior | Side effects | Cleanup behavior | Retryable | Owner | Evidence | Evidence status | Spec treatment | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| FAIL-0001 | `usage_error/configuration_error/preflight_error/service_start_error/service_readiness_timeout/fixture_error/resource_conflict/test_assertion_failure/harness_internal_error/timeout/cancelled/cleanup_error/unsupported_platform/missing_secret/unknown_failure` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `yes/no/conditional/unknown` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Ambiguity register template

| Ambiguity ID | Surface | Ambiguity type | Conflicting or missing facts | Impact | Owner required | Current workaround | Decision prompt | Evidence | Evidence status | Resolution status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|
| AMB-0001 | `TODO:` | `missing_default/contradiction/implicit_sequence/resource_owner_unknown/authority_gap/schema_unknown/other` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `open/resolved/deferred` | `TODO:` |

## Authority map template

| Surface | Current source of behavior | Proposed owner | May drive execution | Conflict precedence | Known conflicts | Required decision | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| `TODO:` | `main_spec/harness_spec/implementation/tests/fixtures/ci/logs/reports/local_policy/unknown` | `TODO:` | `yes/no/conditional/unknown` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Preservation matrix template

| Behavior ID | Behavior | Current evidence | Main-spec dependency | External dependency | Failure cost | Classification | Required decision | Roadmap target | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| PRES-0001 | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `low/medium/high/critical/unknown` | `preserve/preserve_with_clarification/refactor_preserving_behavior/deprecate/redesign_required/remove_if_unused/authority_decision_required/exclude_from_contract` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Acceptance matrix template

| Criterion ID | Requirement | Source section | Verification method | Fixture required | Pass condition | Fail condition | Current coverage | Notes |
|---|---|---|---|---|---|---|---|---|
| HARNESS-AC-001 | `TODO:` | `TODO:` | `static_inspection/unit_test/integration_test/ci_run/golden_fixture/manual_review/owner_decision` | `yes/no/TODO` | `TODO:` | `TODO:` | `existing/partial/missing/unknown` | `TODO:` |

## Source-limit log template

| Source-limit ID | Surface | Limit type | What was inspected | What was not inspected | Impact | Follow-up needed | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| SL-0001 | `TODO:` | `inaccessible_file/unavailable_runtime/missing_secret/ci_only/platform_specific/partial_search/unexecuted_command/timeout/other` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `source_limit` | `TODO:` |

## Review packet checklist

Before maintainer review, verify:

- [ ] Recovery charter is complete.
- [ ] Source limits are summarized.
- [ ] Harness inventory is complete or explicitly bounded.
- [ ] Entrypoint map covers local and CI surfaces.
- [ ] Artifact ownership matrix covers fixtures, generated artifacts, temp files, logs, and reports.
- [ ] Service lifecycle map covers all service dependencies.
- [ ] Resource allocation register covers shared resources.
- [ ] Failure-mode register covers recurring and plausible harness failures.
- [ ] Ambiguity register has owner questions for unresolved items.
- [ ] Preservation matrix classifies every major subsystem.
- [ ] Harness authority map states main-spec versus harness-spec ownership.
- [ ] Harness NLSpec draft has no unstated defaults for harness-owned behavior.
- [ ] Acceptance matrix maps every normative requirement.
- [ ] Roadmap separates specification recovery from future implementation remediation.
- [ ] Handoff note names next recommended action.
- [ ] No implementation changes occurred during recovery.
