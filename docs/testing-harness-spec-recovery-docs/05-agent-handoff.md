---
doc_id: THR-050
title: Testing Harness Recovery Agent Handoff Contract and Template
status: draft
role: handoff-template
---

# Testing Harness Recovery Agent Handoff Contract and Template

## Document role

This document defines the required interface for testing-harness recovery handoff
artifacts and provides the reusable handoff template.

A handoff must preserve continuity across agent sessions without relying on
hidden transcript memory. A future agent must be able to determine what changed,
what evidence supports the claims, what limits remain open, and what work must
or must not be repeated from the handoff alone.

This document is a template and contract. It is not a filled handoff for the
current session, not runtime evidence, and not authority to modify harness
implementation behavior.

## Current recovery state

As of 2026-05-09, the recovery package is in a reviewable recovered-specification
state. Later handoffs are maintenance or targeted follow-up handoffs, not a
restart of the completed recovery stages.

| Source | Handoff use |
|---|---|
| `01-recovery-process.md` | Current S0 through S12 process state, preserved limits, and rule that completed stages must not be rerun unless new evidence invalidates them or a maintainer requests targeted follow-up. |
| `02-nlspec-writing-guide.md` | NLSpec writing gates, evidence-to-spec rules, acceptance-mapping expectations, and required source-limit preservation. |
| `04-registers-and-checklists.md` | Current register status, evidence vocabulary, source-of-truth index, and reusable row schemas. |
| `harness-review-packet.md` | Current review status, verification summary, final S7 handoff material, and historical implementation-change accounting. |
| `maintainer-decision-summary-2026-05-09.md` | Settled S7 maintainer decisions that must be cited by `MD-S7-*` ID and must not be re-decided in handoffs. |

The handoff must preserve these standing decisions:

| Topic | Required treatment |
|---|---|
| Command authority | `make` remains the sole canonical harness command surface per `MD-S7-0001`. |
| Direct package scripts | Direct package scripts remain developer conveniences unless they re-enter Make-owned wrappers. |
| Generated artifacts | Generated task, schedule, Go, and TypeScript artifacts are downstream execution inputs only and must not be hand-edited. |
| Source limits | Environment precedence, visual snapshot refresh bounds, parent-death cleanup, active DB cleanup, detached reaper completion, provider-specific CI behavior, and Playwright report internals remain `source_limit` unless later selected evidence or owner decision closes them. |
| Implementation behavior | Recovery maintenance handoffs must not authorize harness implementation, test, fixture, CI, cleanup, generated artifact, lockfile, package-manager, runtime-service, or product-behavior changes. |

## Handoff artifact interface

Use one handoff file per recovery session or append one dated section to a
shared handoff log when the repository already uses that convention. New files
must live under:

```text
docs/testing-harness-spec-recovery-docs/handoffs/YYYY-MM-DD-<scope-slug>.md
```

The `<scope-slug>` must be lowercase ASCII with words separated by hyphens. If
more than one handoff for the same scope and date is required, append
`-<ordinal>` where `<ordinal>` starts at `2`.

Each handoff file must start with front matter:

```yaml
---
doc_id: THR-HANDOFF-YYYY-MM-DD-<SCOPE>
title: <scope title> Handoff
status: draft
role: recovery-handoff
---
```

| Front-matter field | Requirement |
|---|---|
| `doc_id` | must be stable, unique within the recovery package, and must not reuse `THR-050`. |
| `title` | must name the session scope and include `Handoff`. |
| `status` | must be `draft`, `complete`, `blocked`, or `superseded`. Use `draft` while the handoff is incomplete. |
| `role` | must be `recovery-handoff`, `sprint-handoff`, or `maintenance-handoff`. |

Unknown values must be written as `TODO: <specific_unknown>`. Empty table cells
are invalid. A handoff must not use plain `TODO:` without a specific unknown
name.

## Normative handoff rules

- The handoff must use lowercase `must`, `must not`, `may`, and `default` for
  contract-bearing statements.
- The handoff must state exact paths, commands, run IDs, artifact paths, commit
  revisions, decision IDs, and register row IDs when they exist.
- The handoff must not present guesses as facts. Guesses must be recorded as
  `assumed`, `inferred`, `source_limit`, `owner_required`, or
  `maintainer_decision_required`, according to the evidence.
- The handoff must not close source limits or `owner_required` decisions by
  inference.
- The handoff must not restart completed recovery stages unless the handoff
  cites new evidence that invalidates an earlier finding or a maintainer request
  for targeted follow-up.
- The handoff must record the initial and final dirty state. Pre-existing dirty
  files outside the session scope must be identified as pre-existing and must
  not be claimed as session work.
- The handoff must record every command run by the session. Read-only commands,
  static inspections, and validation commands all count as commands.
- The handoff must record `make agent-finalize` status. If it was not run, the
  handoff must state the exact reason it was skipped.
- The handoff must record an implementation-change audit even when no
  implementation files changed.

## Allowed vocabularies

### Evidence status values

| Evidence status | Allowed claim strength in a handoff |
|---|---|
| `source_observed` | may support claims about inspected source, manifests, docs, tests, fixtures, schemas, or committed artifacts. Does not prove runtime success. |
| `observed` | Legacy/general direct inspection status. Prefer `source_observed` for new source-level rows. |
| `selected_runtime_observed` | may support a durable runtime claim only for the exact command, environment, inputs, artifacts, exit status, and run identity recorded. |
| `runtime_observed` | may record runtime evidence from a command. Convert durable S7-style claims to `selected_runtime_observed` when run identity and artifact details are present. |
| `maintainer_decision` | may close the scoped authority or intent gap when the decision ID is cited and no primary owner conflicts. |
| `owner_required` | must remain open until a governing owner decides. |
| `maintainer_decision_required` | must remain open until a maintainer or governing owner decides. Equivalent to `owner_required` for routing. |
| `inferred` | may explain a hypothesis derived from observed facts but must not be sole support for final normative behavior. |
| `assumed` | must not support normative behavior. Replace with evidence, source limit, or owner decision when possible. |
| `contradiction` | must not be reconciled silently. Record the conflicting sources, owner, blocked claim, and current treatment. |
| `source_limit` | must remain non-final. Record what was inspected, what was not inspected, impact, and follow-up needed. |

### Status and severity values

| Field | Allowed values |
|---|---|
| Handoff front-matter `status` | `draft`, `complete`, `blocked`, `superseded` |
| Sprint or process status | `not_started`, `in_progress`, `blocked`, `complete`, `superseded` |
| Severity | `low`, `medium`, `high`, `critical`, `unknown` |
| Retryable | `yes`, `no`, `conditional`, `unknown` |
| Audit result | `yes`, `no`, `not_applicable`, or `TODO: <specific_unknown>` |

## Section requirements

The template sections below are mandatory for a complete handoff unless this
table provides an explicit omission default.

| Section | Required content | Exact default when no rows exist |
|---|---|---|
| Session metadata | Repository revision, branch, dirty state, platform, scope, changed paths, and implementation-change status. | No default; all rows are required. |
| Work completed this session | Concrete completed work with paths or artifact IDs. | `No work was completed in this session.` |
| Files and surfaces inspected | Inspected source, docs, registers, artifacts, and why each was inspected. | `No files or surfaces were inspected in this session.` |
| Commands run | Every command, purpose, result, exit code, artifacts, and notes. | `No commands were run in this session.` |
| Recovery artifacts updated | Recovery docs or handoff logs updated, with remaining gaps. | `No recovery artifacts were updated in this session.` |
| Key findings | Evidence-labeled findings that matter to future agents. | `No new findings were recorded in this session.` |
| New or updated ambiguities | New or changed ambiguity rows and owner routing. | `No new or updated ambiguities were recorded in this session.` |
| New or updated hazards | New or changed hazard rows. | `No new or updated hazards were recorded in this session.` |
| New or updated failure modes | New or changed failure-mode rows. | `No new or updated failure modes were recorded in this session.` |
| Source limits and inaccessible material | Source limits discovered or preserved by the session. | `No new source limits were recorded in this session.` |
| Decisions made | Maintainer or governing-source decisions only. | `No maintainer or governing-source decisions were made in this session.` |
| Pending owner decisions | Open `owner_required` decisions that affect later work. | `No pending owner decisions were added or changed in this session.` |
| Current blockers | Active blockers and unblock requirements. | `No current blockers were added or changed in this session.` |
| Suggested next actions | Ordered actions a future agent can take without transcript memory. | `No next actions are proposed from this session.` |
| Do not repeat or redo | Work that must not be repeated unless evidence changes. | `No repeat-avoidance notes were added in this session.` |
| Implementation-change audit | Audit table and unauthorized-change handling. | No default; all rows are required. |
| Final handoff summary | Concise paragraph for the next agent. | No default; a summary is required. |

## Command and artifact field contracts

| Field | Required format |
|---|---|
| `Command` | Exact command as run, including environment prefixes that affect behavior. |
| `Purpose` | One sentence naming why the command was run. |
| `Result` | `pass`, `fail`, `not_run`, or a concise observable result. |
| `Exit code` | Numeric exit code, or `not_applicable` only when no process was started. |
| `Artifacts produced` | Exact paths, run IDs, or `None`. |
| `Notes` | Source limits, partial results, retry details, or `None`. |
| `Artifact` | Exact recovery artifact path or stable artifact ID. |
| `Update summary` | Behaviorally relevant change; avoid cosmetic descriptions. |
| `Remaining gaps` | Open source limits, owner decisions, or `None`. |

## Unauthorized implementation-change handling

| Condition | Handoff requirement |
|---|---|
| No implementation, test, fixture, CI, cleanup, generated, lockfile, package-manager, runtime-service, or product-behavior files changed. | Mark all audit rows `no` except `Only recovery docs changed`, which must be `yes`. |
| A non-recovery file was already dirty before the session and was not touched. | Record it in dirty-state metadata as pre-existing and exclude it from session work. |
| A non-recovery file changed with separate authorization. | Record the authorizing request, affected paths, verification, and follow-up disposition. |
| A non-recovery file changed without separate authorization. | Set handoff `status: blocked`, mark the relevant audit row `yes`, record exact paths, record remediation instructions, and do not describe the session as complete. |
| A generated file changed as a result of required maintenance. | Record the generating command and owner input; do not hand-edit generated paths. |

## Handoff template

## Session metadata

| Field | Value |
|---|---|
| Handoff ID | `TODO: handoff_id_unknown` |
| Session date | `TODO: session_date_unknown` |
| Agent/session identifier | `TODO: agent_session_unknown` |
| Repository revision | `TODO: repository_revision_unknown` |
| Branch | `TODO: branch_unknown` |
| Dirty state at start | `TODO: dirty_state_start_unknown` |
| Dirty state at end | `TODO: dirty_state_end_unknown` |
| Runtime platform | `TODO: runtime_platform_unknown` |
| Recovery scope | `TODO: recovery_scope_unknown` |
| Process or sprint status after session | `TODO: process_status_unknown` |
| Recovery-doc paths changed | `TODO: recovery_doc_paths_changed_unknown` |
| Non-recovery paths changed | `TODO: non_recovery_paths_changed_unknown` |
| `make agent-finalize` status | `TODO: agent_finalize_status_unknown` |

## Work completed this session

List concrete work completed. Use exact artifact names or paths. If no work was
completed, use the exact default from the section-requirements table.

- `TODO: completed_work_unknown`

## Files and surfaces inspected

| Surface | Why inspected | Evidence status | Notes |
|---|---|---|---|
| `TODO: inspected_surface_unknown` | `TODO: inspection_reason_unknown` | `TODO: evidence_status_unknown` | `TODO: inspection_notes_unknown` |

## Commands run

| Command | Purpose | Result | Exit code | Artifacts produced | Notes |
|---|---|---|---:|---|---|
| `TODO: command_unknown` | `TODO: command_purpose_unknown` | `TODO: command_result_unknown` | `TODO: exit_code_unknown` | `TODO: command_artifacts_unknown` | `TODO: command_notes_unknown` |

If no commands were run, record:

```text
No commands were run in this session.
```

## Recovery artifacts updated

| Artifact | Update summary | Remaining gaps |
|---|---|---|
| `TODO: artifact_unknown` | `TODO: update_summary_unknown` | `TODO: remaining_gaps_unknown` |

## Key findings

Use evidence labels. Do not present guesses as facts.

| Finding ID | Finding | Evidence status | Evidence reference | Impact |
|---|---|---|---|---|
| `TODO: finding_id_unknown` | `TODO: finding_unknown` | `TODO: evidence_status_unknown` | `TODO: evidence_reference_unknown` | `TODO: finding_impact_unknown` |

## New or updated ambiguities

| Ambiguity ID | Surface | Decision required | Proposed owner | Blocking sprint |
|---|---|---|---|---|
| `TODO: ambiguity_id_unknown` | `TODO: ambiguity_surface_unknown` | `TODO: ambiguity_decision_unknown` | `TODO: ambiguity_owner_unknown` | `TODO: ambiguity_blocking_scope_unknown` |

## New or updated hazards

| Hazard ID | Surface | Trigger | Severity | Next action |
|---|---|---|---|---|
| `TODO: hazard_id_unknown` | `TODO: hazard_surface_unknown` | `TODO: hazard_trigger_unknown` | `TODO: hazard_severity_unknown` | `TODO: hazard_next_action_unknown` |

## New or updated failure modes

| Failure ID | Failure class | Trigger | Retryable | Next action |
|---|---|---|---|---|
| `TODO: failure_id_unknown` | `TODO: failure_class_unknown` | `TODO: failure_trigger_unknown` | `TODO: failure_retryable_unknown` | `TODO: failure_next_action_unknown` |

## Source limits and inaccessible material

| Source-limit ID | Surface | Limit | Impact | Follow-up |
|---|---|---|---|---|
| `TODO: source_limit_id_unknown` | `TODO: source_limit_surface_unknown` | `TODO: source_limit_unknown` | `TODO: source_limit_impact_unknown` | `TODO: source_limit_follow_up_unknown` |

## Decisions made

Record only decisions actually made by a maintainer or governing source. Do not
invent authority.

| Decision ID | Decision | Authority/source | Affected docs | Follow-up |
|---|---|---|---|---|
| `TODO: decision_id_unknown` | `TODO: decision_unknown` | `TODO: decision_authority_source_unknown` | `TODO: decision_affected_docs_unknown` | `TODO: decision_follow_up_unknown` |

If no authority decisions were made, record:

```text
No maintainer or governing-source decisions were made in this session.
```

## Pending owner decisions

| Decision prompt | Why it matters | Suggested owner | Blocking effect |
|---|---|---|---|
| `TODO: decision_prompt_unknown` | `TODO: decision_importance_unknown` | `TODO: suggested_owner_unknown` | `TODO: blocking_effect_unknown` |

## Current blockers

| Blocker | Affected sprint | What is needed to unblock | Owner |
|---|---|---|---|
| `TODO: blocker_unknown` | `TODO: blocker_scope_unknown` | `TODO: unblock_requirement_unknown` | `TODO: blocker_owner_unknown` |

## Suggested next actions

List the next actions in exact order. Each item must be actionable by a future
agent without reading the previous transcript.

1. `TODO: next_action_unknown`

## Do not repeat or redo

List work the next agent should not repeat unless evidence has changed.

- `TODO: repeat_avoidance_unknown`

## Implementation-change audit

The recovery process must not modify implementation behavior unless separately
authorized. Record the audit result.

| Check | Result |
|---|---|
| Harness implementation files modified | `TODO: harness_implementation_audit_unknown` |
| Test logic modified | `TODO: test_logic_audit_unknown` |
| CI behavior modified | `TODO: ci_behavior_audit_unknown` |
| Fixture, golden, or snapshot contents modified | `TODO: fixture_audit_unknown` |
| Cleanup scripts modified | `TODO: cleanup_script_audit_unknown` |
| Generated artifacts or lockfiles modified | `TODO: generated_or_lockfile_audit_unknown` |
| Package-manager or runtime-service files modified | `TODO: package_runtime_audit_unknown` |
| Product behavior modified | `TODO: product_behavior_audit_unknown` |
| Only recovery docs changed | `TODO: recovery_docs_only_audit_unknown` |

If any non-recovery file changed without separate authorization, set handoff
front-matter `status: blocked` and record remediation instructions here:

`TODO: unauthorized_change_remediation_unknown`

## Final handoff summary

Write a concise paragraph for the next agent.

`TODO: final_handoff_summary_unknown`

## Handoff acceptance criteria

A handoff produced from this template is valid only when all criteria below are
true.

| Criterion ID | Acceptance criterion |
|---|---|
| THR-050-AC-001 | The handoff has unique front matter with allowed `status` and `role` values. |
| THR-050-AC-002 | Every required section is present in the section order defined by this template. |
| THR-050-AC-003 | Required empty sections use the exact default sentence from the section-requirements table. |
| THR-050-AC-004 | No table cell is empty, and every unknown value uses `TODO: <specific_unknown>`. |
| THR-050-AC-005 | Every command run by the session is recorded with purpose, result, exit code, artifacts, and notes. |
| THR-050-AC-006 | Evidence statuses are limited to the allowed evidence vocabulary in this document. |
| THR-050-AC-007 | Source limits and `owner_required` decisions are preserved unless selected evidence or a cited owner decision closes them. |
| THR-050-AC-008 | Settled S7 decisions are cited by `MD-S7-*` ID when they affect the handoff. |
| THR-050-AC-009 | The implementation-change audit accounts for recovery and non-recovery paths changed during the session. |
| THR-050-AC-010 | Unauthorized non-recovery changes force `status: blocked` and include remediation instructions. |
| THR-050-AC-011 | `make agent-finalize` status is recorded as passed, failed with subtarget, skipped with reason, or not run with reason. |
| THR-050-AC-012 | The final summary is sufficient for a future agent to resume without reading the previous transcript. |
