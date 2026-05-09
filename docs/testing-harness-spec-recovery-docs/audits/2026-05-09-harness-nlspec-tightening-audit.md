---
doc_id: THR-HARNESS-NLSPEC-TIGHTENING-AUDIT-2026-05-09
title: Harness NLSpec Tightening Audit
status: complete
role: audit
---

# Harness NLSpec Tightening Audit

## Audit Purpose

This audit records a documentation-only NLSpec-style tightening pass for
`harness-nlspec.md`. The pass uses `docs/research/nlspec-spec.md` and
`02-nlspec-writing-guide.md` as quality references, while preserving the
completed recovery posture recorded by `01-recovery-process.md`,
`03-sprint-plan.md`, `harness-review-packet.md`, and the 2026-05-09 maintainer
decisions.

## Boundary

Scope inspected and cited:

- `docs/research/nlspec-spec.md`
- `docs/testing-harness-spec-recovery-docs/02-nlspec-writing-guide.md`
- `docs/testing-harness-spec-recovery-docs/harness-nlspec.md`
- `entrypoint-command-map.md`, `observable-interface-map.md`,
  `structured-output-schema-notes.md`, `service-lifecycle-map.md`,
  `resource-allocation-register.md`, `timeout-retry-register.md`,
  `cleanup-lifecycle-matrix.md`, and `failure-mode-register.md`
- `harness-acceptance-matrix.md`, `source-limit-log.md`,
  `ambiguity-register.md`, `harness-authority-map.md`, and
  `preservation-matrix.md`

Non-scope:

- No Go, TypeScript, SQL, generated code, fixture, golden, snapshot, lockfile,
  CI, cleanup, package-manager, service, or runtime behavior changes.
- No new command surface, schema ID, or `HAC-*` acceptance family.
- No closed source limits by inference.

## Findings

| Finding ID | Finding | Resolution |
|---|---|---|
| THR-NLSPEC-AUD-0001 | The S7 package remains in the completed recovered-specification state; current work should not restart S0-S12 recovery. | Preserved the existing S7 posture and cited the row-level recovery inputs instead of redoing recovery. |
| THR-NLSPEC-AUD-0002 | The draft had strong authority and acceptance boundaries, but row-level interface ownership was implicit enough that future readers could miss where exact command, output, schema, timing, cleanup, and failure details live. | Added `Audit Basis and Contract Delegation` to adopt the relevant `EP-*`, `OI-*`, `SCHEMA-*`, `SVC-*`, `RES-*`, `TMR-*`, `CLN-*`, `FAIL-*`, `HAC-*`, `SL-*`, and `AMB-*` rows by reference. |
| THR-NLSPEC-AUD-0003 | Command-family contracts were correct at a high level but too aggregated for NLSpec recreatability. | Added a compact command-family-to-controlling-row map without expanding the spec into a register copy. |
| THR-NLSPEC-AUD-0004 | Timing, resource, lock, reset, ordinary DB/object operation, and retained-artifact boundaries needed a caller-visible summary to avoid accidental hard guarantees. | Added a source-observed timing/resource boundary table with preserved source limits. |
| THR-NLSPEC-AUD-0005 | Schema wording needed to distinguish stable schema markers from field-complete contracts and tool-owned reports. | Added schema adoption text that keeps `partial`, `schema_unknown`, `authority_unknown`, and tool-defined outputs bounded. |

## Preserved Limits

The audit keeps these limits open exactly as limits:

| Limit | Preserved treatment |
|---|---|
| Environment-variable precedence | `TODO: precedence_unknown`; source-observed variables and defaults only. |
| Visual snapshot refresh bounds | Validation-only until an owner supplies OS, browser, version, and update command. |
| Parent-death cleanup | No guaranteed abrupt-exit cleanup claim. |
| Active DB cleanup | No guaranteed live active-connection cleanup claim. |
| Detached reaper completion | Scheduling and delayed after-state remain short of hard completion proof. |
| Provider CI behavior | Provider annotations, uploads, dashboards, and workflow behavior remain source-limited while `.github/**` is absent. |
| Playwright artifact internals | Reports, traces, videos, and screenshots remain tool-owned or `schema_unknown`. |
| Release readiness beyond recorded evidence | Kept separate from stale-smoke demotion. |

## Verification Record

Verification is recorded here after the documentation edits are checked. This
audit file itself does not authorize implementation or runtime evidence
collection.

| Check | Status | Notes |
|---|---|---|
| `git diff --check` | `pass` | No whitespace or patch-format issues. |
| Targeted normative-language search | `pass` | `rg` found no `appropriate`, `reasonable`, `robust`, `where needed`, `when available`, `should`, `SHOULD`, `MUST`, or `MAY` matches in `harness-nlspec.md`. |
| Row-ID existence check | `pass` | Every concrete row ID cited by the changed files resolves in existing recovery docs outside the changed files. |
| `make agent-finalize` | `pass` | Ran unchanged; skipped duration baseline refresh because `RESULTS_DIR` was unset. |

## No-Implementation-Change Confirmation

This audit changed recovery documentation only. It did not change harness
implementation, product behavior, fixtures, goldens, snapshots, generated
artifacts, CI behavior, cleanup scripts, lockfiles, package-manager files,
runtime services, or test behavior.
