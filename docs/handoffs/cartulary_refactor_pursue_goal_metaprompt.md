# Cartulary Refactor Pursue Goal Metaprompt

## Purpose

Use this metaprompt to generate one specific Codex `/goal` prompt for a module-level Cartulary refactor. It is designed for Goal mode, where the goal text must carry the durable objective, stopping condition, validation loop, and constraints. The generated `/goal` prompt is capped at **3,900 characters** to leave margin under a 4,000-character prompt limit.

## Required inputs

Fill these values before running the metaprompt:

| Input | Required value |
| --- | --- |
| `TARGET_MODULE` | One Cartulary module, frontend package, or bounded cross-cutting seam. |
| `WORKSTREAM` | `backend`, `frontend`, `contracts`, `harness`, `docs`, or `cross-cutting`. |
| `PRIOR_ANALYSIS_PATH` | Local Markdown file with prior findings, or `none`. |
| `FRAMEWORK_PATH` | Path to the local refactor planning framework. |
| `HARD_SCOPE` | Files, packages, modules, and explicit exclusions. |
| `VALIDATION_TARGETS` | Canonical Make targets or `TODO` when discovery must determine them. |
| `KNOWN_BLOCKERS` | Current failing tests, stale artifacts, missing docs, or `none`. |

## Copy-ready metaprompt

```text
Act as a Cartulary refactor-goal prompt compiler. Create exactly one Codex `/goal` prompt for the current repository. The generated goal prompt must be <=3,900 characters including `/goal`; output only the prompt, no notes, no code fence.

Inputs:
TARGET_MODULE=<one Cartulary module or frontend package>
WORKSTREAM=<backend|frontend|contracts|harness|docs|cross-cutting>
PRIOR_ANALYSIS_PATH=<local markdown file or none>
FRAMEWORK_PATH=<local planning framework path>
HARD_SCOPE=<files/modules in scope; exclusions>
VALIDATION_TARGETS=<Make targets or TODO if unknown>
KNOWN_BLOCKERS=<current failing tests, missing docs, or none>

Prompt contract:
1. Name one durable objective and one verifiable stopping condition.
2. Tell Codex to read FRAMEWORK_PATH, PRIOR_ANALYSIS_PATH when present, AGENTS.md when present, owner docs, package manifests, and the target source before editing.
3. Require current-state discovery before implementation: module inventory, public contract surfaces, dependency imports, tests, generated files, and behavior seams.
4. Preserve observable behavior unless an explicitly cited owner contract requires a behavior change. Do not change public routes, wire envelopes, WebSocket events, storage semantics, auth/authorization, workbook hot-path behavior, generated contracts, or harness accounting as incidental refactor work.
5. Use module boundaries, not phase boundaries. Keep generated files unedited; update owner docs/contracts before generated outputs; keep `/packages/grid-adapter` as the only direct `react-data-grid` importer; keep domain logic out of platform transport/storage plumbing.
6. Work in checkpoints: discovery plan, characterization tests or test-gap record, smallest behavior-preserving move, validation, cleanup, handoff. After each checkpoint, run the cheapest relevant validation and record results.
7. Require tests before risky movement: use characterization tests for current behavior, then unit/integration/E2E only as needed by the touched contract. Do not delete or rewrite tests unless the plan states the owner and replacement evidence.
8. Pause and report instead of guessing if owner docs conflict, repo state contradicts the plan, a generated artifact is stale, validation fails for an unclear reason, or the next step would widen scope.
9. Maintain `<FRAMEWORK_PATH>#Session Handoff`: update top tracker, workstream notes, decisions, commands run, validation results, changed files, unresolved risks, and next workflow.
10. Stop only when the scoped refactor is complete, validation targets pass or are explicitly blocked with evidence, no out-of-scope diffs remain, and handoff notes are current.

Compress the final `/goal` prompt aggressively: use terse imperatives, no rationale, no tables, no citations, no examples, no repeated constraints. If over 3,900 characters, remove explanatory text before removing constraints.
```

## Output acceptance criteria

A generated Pursue Goal prompt is acceptable only when:

| ID | Criterion |
| --- | --- |
| MPG-AC-001 | The output is exactly one `/goal` prompt and contains no commentary outside the prompt. |
| MPG-AC-002 | The output is no more than 3,900 characters including `/goal`. |
| MPG-AC-003 | The prompt names one objective and one verifiable stopping condition. |
| MPG-AC-004 | The prompt requires source reading before editing. |
| MPG-AC-005 | The prompt preserves observable contracts unless the scoped task explicitly authorizes a behavior change. |
| MPG-AC-006 | The prompt requires checkpoints, validation, and handoff updates. |
| MPG-AC-007 | The prompt contains pause conditions for contradictions, unclear validation failures, stale generated artifacts, or scope expansion. |

## Source posture

This artifact uses OpenAI Goal-mode guidance only for the generic Goal prompt structure: durable objective, measurable completion condition, validation loop, checkpoints, and progress/handoff expectations. Cartulary-specific constraints come from the local planning framework and repository owner documents.

## Sources

[^openai-goals]: OpenAI Developers, “Follow a goal” and “Prompting: Goal mode,” accessed 2026-06-27. The official docs describe `/goal` as a persistent objective for long-running work with a verifiable stopping condition, validation loop, files to read first, checkpoints, and compact progress reporting.
