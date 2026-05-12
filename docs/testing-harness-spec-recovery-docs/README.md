---
doc_id: THR-README
title: Testing Harness Specification Recovery Document Set
status: draft
role: document-index
---

# Testing Harness Specification Recovery Document Set

## Document role

This file is the index and entrypoint for the testing harness recovery document
set. It identifies the current owner documents, evidence registers, preserved
limits, and acceptance criteria for the recovered testing harness NLSpec
package.

This README is not a behavior owner. It must route readers to the document that
owns the relevant claim, gap, or acceptance row.

## Authority and scope

Core 00 through Core 04 own Cartulary product behavior. The testing harness
recovery package may define validation mechanics, harness command behavior,
fixture and artifact handling, cleanup bounds, service lifecycle behavior,
resource policy, diagnostics, and acceptance mapping. It must not redefine
product behavior owned by the normative core.

The adopted harness NLSpec package owns harness mechanics only after maintainer
adoption. Until then, `harness-nlspec.md` is the reviewable recovered harness
NLSpec draft, and `harness-acceptance-matrix.md` is the acceptance mapping for
that draft.

Implementation, tests, fixtures, CI configuration, logs, reports, retained
local artifacts, and local policy are evidence surfaces. They are not automatic
future normative authority when they conflict with Core 00 through Core 04, an
adopted harness NLSpec, or a maintainer decision.

## How to use this set

Use this set according to the task:

| Task | Start with | Then use |
|---|---|---|
| Understand package status | `00-overview.md` | `01-recovery-process.md`, `04-registers-and-checklists.md`, `harness-review-packet.md` |
| Review or revise harness requirements | `harness-nlspec.md` | `02-nlspec-writing-guide.md`, `harness-acceptance-matrix.md`, `maintainer-decision-summary-2026-05-09.md` |
| Check a behavior claim | Relevant register from the source-of-truth map below | `source-limit-log.md`, `ambiguity-register.md`, `harness-authority-map.md` |
| Prepare maintainer review | `harness-review-packet.md` | `harness-implementation-roadmap.md`, `harness-acceptance-matrix.md` |
| Continue recovery work | `05-agent-handoff.md` | latest file under `handoffs/`, `03-sprint-plan.md`, `04-registers-and-checklists.md` |
| Draft a future replacement NLSpec | `templates/harness-nlspec-template.md` | `02-nlspec-writing-guide.md`, current `harness-nlspec.md` |

## Current recovery state

As of 2026-05-09, the recovery package is in a reviewable recovered-specification
state.

Historical schema references in this recovery package preserve dated audit
evidence. Mentions of `cartulary.tool_run_summary.v2`,
`cartulary.web_e2e_stack.v1`, `cartulary.agent_finalize_summary.v1`, or
`cartulary.service_backed_schedule.v11` are not current supported public
contracts unless an adopted owner spec or current `tools/schemas` attachment
also declares them.

| Area | State | Controlling reference |
|---|---|
| S0 through S6 recovery and register phases | Complete; inventory, command, artifact, service, lifecycle, observable-interface, failure, race, timing, resource, and gap-closure records exist. | `00-overview.md`, `01-recovery-process.md`, `04-registers-and-checklists.md` |
| S8 authority and preservation follow-up | Complete; authority and preservation routing exists without inferring owner-required decisions. | `harness-authority-map.md`, `preservation-matrix.md`, `main-spec-conflict-list.md` |
| S7 NLSpec package | Complete and reviewable; NLSpec draft, acceptance matrix, roadmap, review packet, runtime evidence, cleanup evidence, and maintainer decisions exist. | `harness-nlspec.md`, `harness-acceptance-matrix.md`, `harness-review-packet.md` |
| Settled S7 maintainer decisions | Recorded as S7 inputs and must not be re-decided without a later owner decision. | `maintainer-decision-summary-2026-05-09.md` |
| Remaining open items | Preserved as `source_limit`, `owner_required`, or `maintainer_decision_required`; they are not implicit blockers to documenting source-bounded current behavior. | `source-limit-log.md`, `ambiguity-register.md`, `harness-acceptance-matrix.md` |

## Defaults and boundaries

| Surface | Default or boundary |
|---|---|
| Command authority | The default canonical harness command surface is `make`, per `MD-S7-0001`. |
| Direct package scripts | Direct package scripts are developer conveniences unless they re-enter Make-owned wrappers. |
| Generated artifacts | Generated task, schedule, Go, and TypeScript artifacts are downstream execution inputs only. They do not own behavior and must not be hand-edited. |
| Cleanup strength | Cleanup is `best_effort_cleanup` unless selected evidence proves a stronger cleanup tier for a specific path. |
| CI scope | CI is provider-neutral while `.github/**` is absent. This package must not invent provider workflow, annotation, upload, or dashboard behavior. |
| Retained artifacts | Durable claims from retained artifacts require explicit `RESULTS_DIR`, `RUN_ID`, command or target, platform/tool profile, exit status, and artifact paths. Newest-run fallback is human-investigation only. |
| Visual snapshots | Visual snapshots are validation-only. No snapshot refresh OS, browser, version, or update command is adopted. |
| Environment variables | Only source-observed variables and defaults may be documented. Cross-layer precedence remains source-limited. |
| Product behavior | Harness documents may define validation mechanics only. Core 00 through Core 04 remain product-behavior authority. |

## Source-of-truth map

### Orientation and process

| File | Owns in this package |
|---|---|
| `00-overview.md` | Recovery purpose, scope, authority posture, package state, artifact set, preserved limits, and definition of done. |
| `01-recovery-process.md` | Stage workflow, current maintenance posture, stage completion state, and process rules for follow-up recovery. |
| `02-nlspec-writing-guide.md` | NLSpec quality gates, evidence-to-spec rules, normative language rules, required mappings, mechanics, failure taxonomy, and acceptance-matrix rules. |
| `03-sprint-plan.md` | Historical and current sprint progress board. |
| `04-registers-and-checklists.md` | Current recovery checklist, evidence status vocabulary, register source index, preserved open items, and reusable row templates. |
| `05-agent-handoff.md` | Handoff template for future recovery sessions. |

### Active NLSpec and review package

| File | Owns in this package |
|---|---|
| `harness-nlspec.md` | Reviewable recovered testing harness NLSpec draft. |
| `harness-acceptance-matrix.md` | `HAC-*` binary criteria and `HAC-GAP-*` blockers for harness requirements. |
| `harness-implementation-roadmap.md` | Future remediation plan separated from recovered-specification claims. |
| `harness-review-packet.md` | Maintainer review summary, verification record, deferred limits, and final handoff material. |
| `maintainer-decision-summary-2026-05-09.md` | Settled S7 maintainer decisions used by the NLSpec package. |
| `templates/harness-nlspec-template.md` | Scaffold for a future replacement or derived harness NLSpec, not the current recovered draft. |

### Evidence and recovery registers

| File | Owns in this package |
|---|---|
| `recovery-charter.md` | Recovery boundary, allowed write paths, prohibited edits, evidence labels, and initial repository state. |
| `source-limit-log.md` | Unavailable, incomplete, or intentionally uninspected sources and follow-up routing. |
| `ambiguity-register.md` | Missing defaults, contradictions, authority gaps, schema gaps, and owner-required questions. |
| `harness-inventory.md` | Harness file, directory, config, command, fixture, service, artifact, log, cleanup, and policy inventory. |
| `artifact-ownership-matrix.md` | Fixture, generated artifact, report, cache, temp path, external state, and cleanup ownership. |
| `runtime-evidence-register.md` | Selected runtime evidence and durable evidence-selection rules. |
| `cleanup-signal-evidence-register.md` | Cleanup, signal, process, reaper, and port-release evidence from selected S7 runs. |
| `s0-s6-gap-closure-plan.md` | S0 through S6 gap closure, source limits, authority questions, and S7 readiness criteria. |
| `s7-s6-audit-gap-follow-up.md` | S7 carry-forward track for remaining S6 source limits and later authorized evidence collection. |

### Commands, interfaces, schemas, and lifecycle

| File | Owns in this package |
|---|---|
| `entrypoint-command-map.md` | Exact harness entrypoints, command modes, inputs, defaults, outputs, side effects, ordering, parallel-safety, and failure behavior. |
| `sequencing-assumption-list.md` | Command ordering and sequencing assumptions. |
| `observable-interface-map.md` | Exit codes, stdout/stderr, reports, logs, CI outputs, machine outputs, and consumers. |
| `output-consumer-map.md` | Human and machine consumers for harness outputs. |
| `structured-output-schema-notes.md` | Schema-bearing outputs, known `schema_id` markers, partial schemas, `schema_unknown`, and `authority_unknown` rows. |
| `harness-lifecycle-map.md` | Lifecycle phases, terminal states, partial completion, cleanup, and failure routing. |
| `phase-transition-table.md` | Legal and observed phase transitions. |
| `partial-completion-state-list.md` | Partial-completion states and caller-visible consequences. |
| `environment-contract-observations.md` | Source-observed environment variables, defaults, validation behavior, and precedence limits. |

### Services, resources, hazards, and failures

| File | Owns in this package |
|---|---|
| `service-lifecycle-map.md` | Service provisioning, readiness, reset, stop, cleanup, and failure behavior. |
| `resource-allocation-register.md` | Ports, temp dirs, databases, browser profiles, workers, processes, containers, caches, locks, sockets, and release behavior. |
| `cleanup-lifecycle-matrix.md` | Cleanup triggers, scope, order, idempotence, success/failure/timeout/interrupt behavior, and cleanup failure behavior. |
| `shared-state-hazard-list.md` | Shared-state hazards discovered during recovery. |
| `race-timing-resource-register.md` | Race, timing, concurrency, allocation, and ordering hazards. |
| `timeout-retry-register.md` | Timeouts, retries, backoff, polling, debounce/watch behavior, and exhaustion behavior. |
| `concurrency-model-notes.md` | Scheduler and test concurrency model notes. |
| `failure-class-taxonomy.md` | Harness failure class vocabulary and meanings. |
| `failure-mode-register.md` | Failure triggers, outputs, side effects, cleanup, retryability, and ownership. |

### Authority, preservation, audits, and handoffs

| Surface | Owns in this package |
|---|---|
| `harness-authority-map.md` | Main-spec, harness-spec, implementation, tests, fixtures, CI, reports, and local-policy authority relationships. |
| `preservation-matrix.md` | Preserve, clarify, refactor, deprecate, redesign, remove, and authority-decision classifications. |
| `main-spec-conflict-list.md` | Product-spec and harness-spec conflict routing. |
| `embedded-harness-logic-list.md` | Harness behavior embedded in ordinary tests or support code. |
| `uninvoked-surface-list.md` | Harness-like files or scripts not invoked by discovered entrypoints. |
| `audits/` | Dated audit records for recovery stages and follow-up checks. |
| `handoffs/` | Dated recovery-session handoffs. |

## Preserved limits

The README must preserve these limits exactly as limits unless a later selected
evidence pass or owner decision closes them.

| Limit | Current status | Claim that remains blocked |
|---|---|---|
| Environment-variable precedence | `source_limit/owner_required` | Any final precedence matrix across Make, scripts, schedulers, config files, and Playwright. |
| Visual snapshot refresh OS, browser, version, and command | `source_limit/owner_required` | Any adopted visual snapshot update workflow or platform bound. |
| Parent-death cleanup | `source_limit` | Any final claim of abrupt-exit cleanup. |
| Active DB cleanup | `source_limit` | Any final claim of cleanup while live DB connections remain active. |
| Detached reaper hard completion | `source_limit` | Any final claim of detached cleanup completion. |
| Provider-specific CI behavior while `.github/**` is absent | `source_limit` | Provider workflow, annotation, upload, and dashboard contracts. |
| Playwright report, trace, video, and screenshot internals | `schema_unknown/source_limit` | Stable harness schemas for tool-owned Playwright artifacts. |
| Release readiness beyond recorded evidence | `not_claimed` | A broad release-ready claim not backed by the recorded gate result or later selected evidence. |

## README acceptance criteria

| Criterion ID | Requirement | Verification method | Pass condition |
|---|---|---|---|
| README-AC-001 | The README must identify itself as an index and guardrail, not a behavior owner. | Manual review | `Document role` and `Authority and scope` route behavior ownership to the normative core or package owner documents. |
| README-AC-002 | The README must preserve current recovery state without restarting completed recovery stages. | Manual review | The current-state table says S0-S6, S8, and S7 package work are complete/reviewable and routes follow-up work to maintenance or targeted evidence. |
| README-AC-003 | The README must state the default command authority. | Manual review | `make` is named as the default canonical harness command surface and direct package scripts are bounded as conveniences. |
| README-AC-004 | The README must preserve unresolved limits. | Manual review | The preserved-limits table includes environment precedence, visual snapshot refresh bounds, parent-death cleanup, active DB cleanup, detached reaper completion, provider CI behavior, Playwright internals, and release readiness. |
| README-AC-005 | The README must not authorize implementation changes. | Manual review | The implementation-change boundary prohibits harness code, tests, fixtures, CI behavior, generated artifacts, cleanup scripts, lockfiles, package-manager files, runtime services, and product behavior changes. |
| README-AC-006 | The README must keep traceability to owner documents. | Static inspection | Every named Markdown file exists in `docs/testing-harness-spec-recovery-docs/**`; directory references point to existing package directories. |
| README-AC-007 | The README must not promote unsupported claims. | Manual review | It does not specify env precedence, hard cleanup guarantees, visual snapshot update bounds, provider-specific CI behavior, Playwright internals, or product behavior changes. |

## Implementation-change boundary

This pack does not authorize implementation changes. Harness code, product code,
test logic, fixtures, goldens, snapshots, CI behavior, generated artifacts,
cleanup scripts, service lifecycle behavior, lockfiles, package-manager files,
runtime-service behavior, and product behavior must not be rewritten from this
README or from recovery-document maintenance unless a maintainer gives a
separate implementation-change instruction.
