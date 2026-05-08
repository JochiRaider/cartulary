---
doc_id: THR-README
title: Testing Harness Specification Recovery Document Set
status: draft
role: document-index
---

# Testing Harness Specification Recovery Document Set

## Document role

This file is the index for a repository-ready Markdown pack that guides agents through reverse-specification recovery for an existing testing harness.

## How to use this set

Start with `00-overview.md`, then use `03-sprint-plan.md` as the working progress board while following `01-recovery-process.md`. Use `02-nlspec-writing-guide.md` when converting recovered behavior into specification text. Use `04-registers-and-checklists.md` and `05-agent-handoff.md` to preserve continuity across multiple agent sessions. Copy `templates/harness-nlspec-template.md` when drafting the recovered harness specification.

## Documents

| File | Role |
|---|---|
| `00-overview.md` | Explains purpose, boundary, authority posture, and definition of done. |
| `01-recovery-process.md` | Defines the staged reverse-specification recovery workflow. |
| `02-nlspec-writing-guide.md` | Shows how to convert recovered behavior into precise NLSpec material. |
| `03-sprint-plan.md` | Breaks the work into practical, agent-sized sprints with progress fields. |
| `04-registers-and-checklists.md` | Provides working checklists and register templates. |
| `05-agent-handoff.md` | Provides a multi-session handoff template. |
| `templates/harness-nlspec-template.md` | Provides a scaffold for the recovered testing harness NLSpec. |

## Implementation-change boundary

This pack does not authorize implementation changes. Harness code, tests, fixtures, CI behavior, generated artifacts, and cleanup scripts must not be rewritten during recovery unless the maintainer gives a separate implementation-change instruction.
