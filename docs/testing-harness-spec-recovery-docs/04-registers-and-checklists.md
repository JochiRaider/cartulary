---
doc_id: THR-040
title: Testing Harness Recovery Registers and Checklists
status: draft
role: registers-and-checklists
---

# Testing Harness Recovery Registers and Checklists

## Document role

This document summarizes the current state of the testing-harness recovery package and preserves reusable register templates for later targeted follow-up work.

The populated current-state registers live in their individual files, such as `harness-inventory.md`, `entrypoint-command-map.md`, `runtime-evidence-register.md`, and `harness-review-packet.md`. Those files are the current evidence sources. The templates below are not the current state by themselves.

## Use rules

- Use the source-of-truth index below before opening or updating a populated register.
- Do not close preserved source limits by inference. Keep them `source_limit`, `owner_required`, or `maintainer_decision_required` until later selected evidence or owner decision exists.
- Add evidence references to exact files, sections, commands, run IDs, or retained-artifact paths.
- Keep one row per observable behavior, artifact class, service, failure mode, hazard, or decision.
- Do not merge contradictory evidence into one reconciled statement. Add or preserve a contradiction row.
- Use the templates only for new follow-up rows or new recovery sessions; do not treat placeholder template rows as active evidence.

## Evidence status values

| Status | Meaning |
|---|---|
| `observed` | Legacy/general direct inspection of repository source, config, docs, tests, fixtures, CI, or committed artifacts. |
| `source_observed` | Direct source-level evidence from authored source, manifests, docs, tests, fixtures, schemas, or committed artifacts. It does not prove live runtime success. |
| `selected_runtime_observed` | Runtime evidence from an explicitly selected command, result directory, run ID, platform/tool profile, exit status, and artifact set. |
| `runtime_observed` | Runtime evidence from a recovery command. Prefer `selected_runtime_observed` for durable S7 claims. |
| `maintainer_decision` | A maintainer or governing owner decision has settled authority or intent for the scoped claim. |
| `owner_required` | Owner input is required before the claim can become final. |
| `maintainer_decision_required` | Behavior requires human authority; retained for older register rows and equivalent to `owner_required` for S7 routing. |
| `inferred` | Derived from multiple observed sources. |
| `assumed` | Temporary assumption pending evidence. |
| `contradiction` | Sources disagree and owner decision is required. |
| `source_limit` | Agent could not inspect enough to decide. |

## Current recovery state

As of 2026-05-09, the recovery package is in a reviewable recovered-specification state.

| Area | Current state | Primary evidence |
|---|---|---|
| S0 through S6 recovery and register phases | Complete. Charter, inventory, command mapping, artifact ownership, service lifecycle, observable interface, lifecycle, failure, race, timing, resource, and gap-closure artifacts exist. | `00-overview.md`, `01-recovery-process.md`, `03-sprint-plan.md` |
| S8 authority and preservation follow-up | Complete. Authority and preservation routing exists without converting owner-required questions into inferred decisions. | `preservation-matrix.md`, `harness-authority-map.md`, `main-spec-conflict-list.md` |
| S7 NLSpec package | Complete and reviewable. NLSpec, acceptance matrix, roadmap, review packet, final handoff material, runtime evidence registers, cleanup evidence register, and maintainer decision summary exist. | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `harness-implementation-roadmap.md`, `harness-review-packet.md`, `maintainer-decision-summary-2026-05-09.md` |
| Source limits and owner decisions | Preserved. Remaining gaps are bounded source limits or owner-required items, not blockers to documenting current source-bounded behavior. | `source-limit-log.md`, `ambiguity-register.md`, `harness-review-packet.md` |

The canonical harness command surface is `make`, per `MD-S7-0001`. Direct package scripts remain developer conveniences unless they re-enter Make-owned wrappers.

Generated task, schedule, Go, and TypeScript artifacts are downstream execution inputs. They may drive execution when fresh, but they do not own behavior.

## Recovery progress checklist

| Area | Complete | Notes |
|---|---:|---|
| Recovery charter exists | [x] | `recovery-charter.md`; S0 is complete in `03-sprint-plan.md`. |
| Repository revision and dirty state recorded | [x] | `recovery-charter.md`; selected runtime evidence also records git state in `runtime-evidence-register.md`. |
| Harness boundary candidate list created | [x] | `recovery-charter.md`; current boundary summarized by `harness-inventory.md`. |
| Harness inventory complete | [x] | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md`. |
| Entrypoint map complete | [x] | `entrypoint-command-map.md`, `sequencing-assumption-list.md`; Make authority settled by `MD-S7-0001`. |
| Artifact ownership matrix complete | [x] | `artifact-ownership-matrix.md`; generated artifacts settled as downstream inputs by `MD-S7-0003`. |
| Cleanup lifecycle matrix complete | [x] | `cleanup-lifecycle-matrix.md`; cleanup strength remains bounded by `cleanup-signal-evidence-register.md` and `MD-S7-0008`. |
| Service lifecycle map complete | [x] | `service-lifecycle-map.md`, `environment-contract-observations.md`. |
| Resource allocation register complete | [x] | `resource-allocation-register.md`; scheduler lanes remain logical constraints per `MD-S7-0006`. |
| Observable interface map complete | [x] | `observable-interface-map.md`, `output-consumer-map.md`, `structured-output-schema-notes.md`. |
| Lifecycle map complete | [x] | `harness-lifecycle-map.md`, `phase-transition-table.md`, `partial-completion-state-list.md`. |
| Race/timing/resource register complete | [x] | `race-timing-resource-register.md`, `timeout-retry-register.md`, `concurrency-model-notes.md`. |
| Failure-mode register complete | [x] | `failure-mode-register.md`, `failure-class-taxonomy.md`. |
| Ambiguity register complete | [x] | `ambiguity-register.md`; unresolved rows remain routed instead of inferred closed. |
| Authority map complete | [x] | `harness-authority-map.md`, `maintainer-decision-summary-2026-05-09.md`. |
| Preservation matrix complete | [x] | `preservation-matrix.md`, `main-spec-conflict-list.md`. |
| Harness NLSpec draft created | [x] | `harness-nlspec.md`; evidence-gated final language rule recorded by `MD-S7-0017`. |
| Acceptance matrix complete | [x] | `harness-acceptance-matrix.md`; `HAC-GAP-*` rows preserve unresolved blockers. |
| Roadmap complete | [x] | `harness-implementation-roadmap.md`; future implementation remediation is separated from recovery claims. |
| Review packet complete | [x] | `harness-review-packet.md`; final handoff section and historical implementation-change audit are present. |
| No unauthorized or unaccounted implementation changes remain | [x] | Historical S7 implementation changes are accounted for in `harness-review-packet.md` and `MD-S7-0004`, `MD-S7-0015`, `MD-S7-0016`, `MD-S7-0018`. This row does not claim that no implementation files ever changed during S7. |

## Register and packet source-of-truth index

| Area | Current populated source |
|---|---|
| Recovery status and stage map | `00-overview.md`, `01-recovery-process.md`, `03-sprint-plan.md` |
| Charter, repository state, and source limits | `recovery-charter.md`, `source-limit-log.md` |
| Inventory and boundary | `harness-inventory.md`, `uninvoked-surface-list.md`, `embedded-harness-logic-list.md` |
| Entrypoints and sequencing | `entrypoint-command-map.md`, `sequencing-assumption-list.md` |
| Artifacts, fixtures, generated outputs, and cleanup | `artifact-ownership-matrix.md`, `cleanup-lifecycle-matrix.md`, `shared-state-hazard-list.md` |
| Services, environments, and resources | `service-lifecycle-map.md`, `environment-contract-observations.md`, `resource-allocation-register.md` |
| Observable outputs and schemas | `observable-interface-map.md`, `output-consumer-map.md`, `structured-output-schema-notes.md` |
| Lifecycle and partial completion | `harness-lifecycle-map.md`, `phase-transition-table.md`, `partial-completion-state-list.md` |
| Races, timing, retries, and concurrency | `race-timing-resource-register.md`, `timeout-retry-register.md`, `concurrency-model-notes.md` |
| Failure classes and failure modes | `failure-mode-register.md`, `failure-class-taxonomy.md` |
| Ambiguities, gaps, and source limits | `ambiguity-register.md`, `source-limit-log.md`, `s0-s6-gap-closure-plan.md`, `s7-s6-audit-gap-follow-up.md` |
| Authority, preservation, and main-spec conflict routing | `harness-authority-map.md`, `preservation-matrix.md`, `main-spec-conflict-list.md`, `maintainer-decision-summary-2026-05-09.md` |
| Selected runtime and cleanup evidence | `runtime-evidence-register.md`, `cleanup-signal-evidence-register.md` |
| NLSpec package and review | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `harness-implementation-roadmap.md`, `harness-review-packet.md` |
| Audits and handoffs | `audits/**`, `handoffs/**`, `harness-review-packet.md#final-handoff` |

## Preserved open items

These boundaries remain open unless later selected evidence or maintainer decision closes them.

| Item | Current status | Evidence and routing |
|---|---|---|
| Environment-variable precedence | `source_limit/owner_required` | `SL-0015`, `HAC-GAP-0001`, `MD-S7-0010`; document only source-observed variables and defaults. |
| Visual snapshot refresh OS, browser, version, and command | `source_limit/owner_required` | `AMB-0022`, `AUTH-0014`, `PRES-0018`, `HAC-GAP-0002`, `MD-S7-0013`; visual snapshots remain validation-only. |
| Parent-death cleanup | `source_limit` | `cleanup-signal-evidence-register.md`; no live SIGKILL-parent scenario was executed. |
| Active DB cleanup | `source_limit` | `cleanup-signal-evidence-register.md`; unit-level leak evidence exists, but live active-connection cleanup is not proven. |
| Detached reaper hard completion | `source_limit` | `cleanup-signal-evidence-register.md`; delayed after-state evidence is recorded, but hard completion is not guaranteed. |
| Provider-specific CI behavior while `.github/**` is absent | `source_limit` | `SL-0001`, `HAC-GAP-0005`, `MD-S7-0015`; keep CI provider-neutral and do not invent annotations, uploads, or workflow behavior. |
| Playwright report, trace, video, and screenshot internals | `schema_unknown/source_limit` | `structured-output-schema-notes.md`, `HAC-GAP-0006`, `MD-S7-0014`; treat as tool-owned unless later adopted as stable harness schemas. |
| Release readiness beyond recorded evidence | `not_claimed` | `harness-review-packet.md`, `HAC-GAP-0007`; readiness claims remain separate from stale-smoke demotion and recorded verification outcomes. |

## Reusable follow-up templates

The following tables are templates for new follow-up rows or new recovery sessions. Use the source-of-truth index above for current populated registers.

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

Current maintainer-review checklist:

- [x] Recovery charter is complete.
- [x] Source limits are summarized.
- [x] Harness inventory is complete or explicitly bounded.
- [x] Entrypoint map covers local and CI surfaces.
- [x] Artifact ownership matrix covers fixtures, generated artifacts, temp files, logs, and reports.
- [x] Service lifecycle map covers all service dependencies.
- [x] Resource allocation register covers shared resources.
- [x] Failure-mode register covers recurring and plausible harness failures.
- [x] Ambiguity register has owner questions for unresolved items.
- [x] Preservation matrix classifies every major subsystem.
- [x] Harness authority map states main-spec versus harness-spec ownership.
- [x] Harness NLSpec draft preserves source limits and maps harness-owned behavior to evidence.
- [x] Acceptance matrix maps every normative requirement or routes it to an owner/source-limit blocker.
- [x] Roadmap separates specification recovery from future implementation remediation.
- [x] Handoff material names next recommended actions.
- [x] No unauthorized or unaccounted implementation changes remain after S7 accounting in `harness-review-packet.md` and `maintainer-decision-summary-2026-05-09.md`.
