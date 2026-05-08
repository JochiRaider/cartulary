---
doc_id: THR-050
title: Testing Harness Recovery Agent Handoff Template
status: draft
role: handoff-template
---

# Testing Harness Recovery Agent Handoff Template

## Document role

This template preserves continuity across multiple agent sessions without relying on hidden transcript memory.

Use one handoff file per session or append one dated section to a shared handoff log.

## Handoff filename convention

Default filename:

```text
TODO: recovery_doc_root/handoffs/YYYY-MM-DD-session-<ordinal>.md
```

Use the target repository's existing convention if one exists.

## Session metadata

| Field | Value |
|---|---|
| Handoff ID | `TODO:` |
| Session date | `TODO:` |
| Agent/session identifier | `TODO:` |
| Repository revision | `TODO:` |
| Branch | `TODO:` |
| Dirty state at start | `TODO:` |
| Dirty state at end | `TODO:` |
| Runtime platform | `TODO:` |
| Current sprint | `TODO:` |
| Sprint status after session | `TODO:` |
| Recovery-doc paths changed | `TODO:` |
| Implementation files changed | `MUST be none unless separately authorized: TODO:` |

## Work completed this session

List concrete work completed. Use exact artifact names or paths.

- `TODO:`

## Files and surfaces inspected

| Surface | Why inspected | Evidence status | Notes |
|---|---|---|---|
| `TODO:` | `TODO:` | `observed/runtime_observed/inferred/source_limit` | `TODO:` |

## Commands run

| Command | Purpose | Result | Exit code | Artifacts produced | Notes |
|---|---|---|---:|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

If no commands were run, record:

```text
No runtime commands were run in this session.
```

## Recovery artifacts updated

| Artifact | Update summary | Remaining gaps |
|---|---|---|
| `TODO:` | `TODO:` | `TODO:` |

## Key findings

Use evidence labels. Do not present guesses as facts.

| Finding ID | Finding | Evidence status | Evidence reference | Impact |
|---|---|---|---|---|
| FIND-0001 | `TODO:` | `observed/runtime_observed/inferred/assumed/contradiction/maintainer_decision_required/source_limit` | `TODO:` | `TODO:` |

## New or updated ambiguities

| Ambiguity ID | Surface | Decision required | Proposed owner | Blocking sprint |
|---|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## New or updated hazards

| Hazard ID | Surface | Trigger | Severity | Next action |
|---|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `low/medium/high/critical/unknown` | `TODO:` |

## New or updated failure modes

| Failure ID | Failure class | Trigger | Retryable | Next action |
|---|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `yes/no/conditional/unknown` | `TODO:` |

## Source limits and inaccessible material

| Source-limit ID | Surface | Limit | Impact | Follow-up |
|---|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Decisions made

Record only decisions actually made by a maintainer or governing source. Do not invent authority.

| Decision ID | Decision | Authority/source | Affected docs | Follow-up |
|---|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` | `TODO:` |

If no authority decisions were made, record:

```text
No maintainer or governing-source decisions were made in this session.
```

## Pending owner decisions

| Decision prompt | Why it matters | Suggested owner | Blocking effect |
|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Current blockers

| Blocker | Affected sprint | What is needed to unblock | Owner |
|---|---|---|---|
| `TODO:` | `TODO:` | `TODO:` | `TODO:` |

## Suggested next actions

List the next actions in exact order. Each item must be actionable by a future agent without reading the previous transcript.

1. `TODO:`
2. `TODO:`
3. `TODO:`

## Do not repeat or redo

List work the next agent should not repeat unless evidence has changed.

- `TODO:`

## Implementation-change audit

The recovery process must not modify implementation behavior. Record the audit result.

| Check | Result |
|---|---|
| Harness implementation files modified | `yes/no/TODO` |
| Test logic modified | `yes/no/TODO` |
| CI behavior modified | `yes/no/TODO` |
| Fixture contents modified | `yes/no/TODO` |
| Cleanup scripts modified | `yes/no/TODO` |
| Only recovery docs changed | `yes/no/TODO` |

If any implementation file changed without separate authorization, mark the handoff as blocked and record remediation instructions.

## Final handoff summary

Write a concise paragraph for the next agent.

`TODO:`
