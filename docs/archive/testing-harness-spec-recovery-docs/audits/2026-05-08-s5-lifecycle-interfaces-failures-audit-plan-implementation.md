---
doc_id: THR-S5-AUDIT-PLAN-IMPLEMENTATION-2026-05-08
title: S5 Lifecycle Interfaces and Failures Audit Plan Implementation
status: complete
role: recovery-audit
---

# S5 Lifecycle, Interfaces, and Failures Audit Plan Implementation

## Audit Verdict

`pass_after_s5_link_correction`

Original audit verdict before this documentation correction: `block_s6`.

S5 has the required output artifacts, complete row families, schema coverage for
machine-consumed outputs, lifecycle coverage, transition coverage, partial-state
coverage, failure classes, and preserved source-limit routing. The follow-up
correction fixed the material linked-output/schema inconsistencies in the S5
failure-mode register, so this audit packet no longer blocks S6/S7 consumers
from using the corrected rows.

This is a documentation-readiness blocker only. The audit did not identify or
perform any required harness implementation change.

The current sprint plan already marks S6 complete in this working tree. The
original `block_s6` verdict here applied to the requested S5 readiness gate
semantics: downstream S6/S7 consumers should not rely on the affected S5
failure-register links until they are corrected or explicitly accepted. That
condition is now resolved by the link correction.

## Objective, Scope, and Non-Scope

Audit objective: verify whether S5 gives a complete, accurate,
evidence-grounded account of observable behavior, machine-consumed interfaces,
lifecycle phases, terminal states, partial-completion states, and failure modes.

Scope: documentation-only review of S5 outputs, S0-S4 control inputs, S5
handoff claims, source citations, row coverage, schema links, lifecycle links,
failure classifications, source limits, and ambiguity routing.

Non-scope: no harness rewrite, command behavior change, exit-code change,
report/schema update, CI change, cleanup behavior change, fixture change,
generated-output change, lockfile change, broad runtime probing, service start,
browser run, Docker/Compose run, reset call, mutating maintenance command, or
S6 work.

## Session Metadata

| Field | Value |
|---|---|
| Repository root | `/home/askahn/code/cartulary` |
| HEAD revision | `3cfde09450b3217f0cb9613a78a708e866ad37f5` |
| Audit timestamp command output | `2026-05-08T23:14:54-04:00` |
| Runtime platform | `Linux DeskRip 6.6.114.1-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Mon Dec 1 20:46:23 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| Working tree at audit start | Dirty; existing modified S5 recovery docs and existing S5 audit were present before this audit file was added. |
| Audit write path | `docs/testing-harness-spec-recovery-docs/audits/2026-05-08-s5-lifecycle-interfaces-failures-audit-plan-implementation.md` |

## Reviewed Artifacts

| Artifact group | Files reviewed | Result |
|---|---|---|
| Required S5 outputs | `observable-interface-map.md`, `structured-output-schema-notes.md`, `output-consumer-map.md`, `harness-lifecycle-map.md`, `phase-transition-table.md`, `partial-completion-state-list.md`, `failure-mode-register.md`, `failure-class-taxonomy.md` | present |
| Control documents | `recovery-charter.md`, `03-sprint-plan.md`, `04-registers-and-checklists.md`, `source-limit-log.md`, `ambiguity-register.md` | present |
| Prior audits and handoffs | S2-S4 audits and handoffs; S5 handoff | present |
| Authority/vocabulary references | `docs/spec/00_document_set_status_and_precedence.md`, `docs/domain.md` | reviewed for authority discipline only |
| Source evidence areas | `Makefile`, `tools/task_surface.generated.mk`, `tools/*manifest*.json`, scheduler/resource registries, `scripts/**`, `scripts/lib/test-output/**`, `scripts/lib/tool-output.mjs`, `scripts/lib/failure-taxonomy.mjs`, `tools/testservices/**`, `internal/testutil/**`, `apps/web/e2e/**`, Playwright/Vitest configs, package manifests, `scripts/ci/**` | spot-checked |
| CI provider source | `.github/**` | absent; source limit preserved |

## Evidence and Source Summary

Representative source checks confirmed that source files cited by S5 exist:
Make/task-surface files, scheduler manifests, browser batch manifests, scheduler
resource registry, test-output helpers, tool-output helpers, failure-taxonomy
helpers, check/service-backed scheduler scripts, browser stack/reset scripts,
CI helper scripts, `tools/testservices`, suite-service diagnostics, Playwright
and Vite config, and root/app package manifests.

Representative schema markers were observed in source or generated manifests,
including `cartulary.test_phase_summary.v3`,
`cartulary.test_target_summary.v4`, `cartulary.test_run_summary.v6`,
`cartulary.tool_run_summary.v2`, check and service-backed scheduler summary
and event schemas, `cartulary.web_e2e_stack.v1`,
`cartulary.web_e2e_session_lease.v1`, `cartulary.test_services.lease.v1`,
`cartulary.test.runtime_reset.v1`, `cartulary.fixture_report.v1`, and
`cartulary.playwright_manifest_selection.v1`.

The authority references confirm that recovery docs may describe harness
evidence but must not redefine product behavior or current-profile runtime
conformance.

## Row-Coverage Matrix

| Row family | Expected | Observed | Field-shape check | Result |
|---|---:|---:|---|---|
| `OI-*` | `OI-0001..OI-0017` | 17 | no malformed table rows | pass |
| `SCHEMA-*` | `SCHEMA-0001..SCHEMA-0020` | 20 | no malformed table rows | pass |
| `CONS-*` | `CONS-0001..CONS-0014` | 14 | no malformed table rows | pass |
| `LIFE-*` | `LIFE-0001..LIFE-0015` | 15 | no malformed table rows | pass |
| `PT-*` | `PT-0001..PT-0035` | 35 | no malformed table rows | pass |
| `PCS-*` | `PCS-0001..PCS-0019` | 19 | no malformed table rows | pass |
| `FAIL-*` | `FAIL-0001..FAIL-0029` | 29 | no malformed table rows | pass for shape; blocker for semantic links |
| Failure classes | required taxonomy classes | 15 | retryability present for all classes | pass |

The ID cross-reference pass found no orphaned S5 references for `OI-*`,
`SCHEMA-*`, `CONS-*`, `LIFE-*`, `PT-*`, `PCS-*`, or `FAIL-*`. `SL-*` and
`AMB-*` references used by S5 are also defined in their registers.

## Machine-Output Schema Cross-Check

Machine-consumed rows in `observable-interface-map.md` all link to a
`SCHEMA-*` note or an explicit `TODO: schema_unknown`. Human diagnostics are
distinguished from machine outputs in the classification rules and row notes.

Provider CI annotations and uploaded artifacts are correctly marked
`schema_unknown`/`source_limit` because `.github/**` is absent and repository
source only exposes provider-neutral `scripts/ci/**` helpers.

Playwright failure-only reports, traces, screenshots, and videos remain
`TODO: schema_unknown`; this is correctly preserved as source-limited evidence,
not a stable contract.

## Lifecycle and Partial-State Findings

S5 lifecycle coverage includes Make phase wrappers, aggregate sequences, check
scheduler, service-backed scheduler, Go runner, frontend/Vitest, browser E2E
owned stack, reset boundary, `tools/testservices`, local-dev flows, direct
package scripts, CI-adjacent wrapper, investigation commands, maintenance and
cleanup commands, and uninvoked/source-limited entrypoints.

Every `LIFE-*` row has transition coverage through `PT-*` rows and, where
applicable, partial-completion coverage through `PCS-*` rows. Partial states
cover setup, child work, reporting, cleanup/finalizer, retained-artifact,
package-script bypass, maintenance, Go accounting, Vitest watchdog, browser
stack, reset, and testservices cases.

Runtime-only cleanup, readiness, timeout, interrupt, parent-death, port release,
detached reaper, and failure-bundle claims remain source-limited and routed to
S6/S8 rather than promoted as guarantees.

## Failure Taxonomy and Register Findings

The failure class taxonomy separates source helper classes from S5 classes and
explicitly separates product assertion failures from harness operational
failures. Every taxonomy row includes detection, reporting location,
exit/report/artifact surface, retryability, ownership, cleanup follow-up,
artifact follow-up, resource follow-up, and evidence status.

The failure-mode register has complete table shape and explicit retryability
for all `FAIL-*` rows. It also includes product-versus-harness ownership,
cleanup/artifact/resource follow-up, evidence, evidence status, and handoff
notes for every row.

The material issue is that several `FAIL-*` rows point to semantically wrong
`OI-*` or `SCHEMA-*` rows. These are not orphaned IDs, but they are inaccurate
links:

| Finding ID | Severity | Rows | Issue | Disposition |
|---|---|---|---|---|
| AUD-S5-PLAN-0001 | resolved | `FAIL-0005`, `FAIL-0006` | Testservices preflight/start failures linked `OI-0009` and `SCHEMA-0013`, which are reset/Playwright surfaces, instead of the testservices output/schema surfaces. | Resolved: rows now link `OI-0005`, `OI-0010`, and `SCHEMA-0014`. |
| AUD-S5-PLAN-0002 | resolved | `FAIL-0010` | Browser reset failure linked `SCHEMA-0012`, a Vitest manifest summary surface, instead of the reset/session schema surface. | Resolved: row now links `OI-0009` and `SCHEMA-0010`. |
| AUD-S5-PLAN-0003 | resolved | `FAIL-0016` | Maintenance/cleanup failure linked `OI-0013`, the direct package-script output surface, while the maintenance/cleanup surface is `OI-0015`. | Resolved: row now links `OI-0015`. |
| AUD-S5-PLAN-0004 | resolved | `FAIL-0018` | Direct package-script authority row linked `OI-0011`, the investigation/node-tool output surface, while direct package scripts are mapped by `OI-0013`. | Resolved: row now links `OI-0013` and `SCHEMA-0018`. |
| AUD-S5-PLAN-0005 | resolved | `FAIL-0027` | Retained-artifact discovery/investigation failure linked `OI-0014`, the CI-adjacent output surface, instead of investigation/retained-artifact output surfaces. | Resolved: row now links `OI-0011`, `SCHEMA-0016`, `SCHEMA-0017`, `CONS-0009`, and `CONS-0011`. |

These inconsistencies did not require harness code changes, and the corrected
links make the S5 failure-mode register accurate enough for downstream use.

## Source-Limit and Ambiguity Findings

Required source limits and ambiguities are preserved. The audit verified
coverage for absent `.github/**`, unexecuted broad gates, unexecuted
service-backed/browser/Docker/Compose/reset/cleanup behavior, stale retained
artifacts, failure-only bundles, unsupported platform behavior, environment
precedence, reset-route authority, package-script authority, local-dev
persistence, stale janitor authority, detached reaper completion, timeout and
interrupt cleanup, and external cache cleanup policy.

These are correctly routed as follow-up work to S6, S7, or S8 and are not
blockers by themselves.

## Blockers and Follow-Ups

### S6-Blocking Issues

| Blocker ID | Issue | Required resolution |
|---|---|---|
| AUD-S5-BLOCK-0001 | Failure-mode register contained semantically incorrect linked output/schema references in multiple rows. | Resolved by corrected `failure-mode-register.md` links; S6/S7 may consume the affected rows. |

### Follow-Up Work

| Follow-up | Route | Reason non-blocking after link correction |
|---|---|---|
| Controlled failure examples | S6/S7 | S5 preserves lack of controlled failures as source-limited. |
| Playwright failure bundle schemas | S7 with selected artifacts | Marked `TODO: schema_unknown`; not promoted. |
| CI provider workflows, annotations, dashboards, and uploads | S8/CI owner | `.github/**` absent; provider source unavailable. |
| Direct package-script contract status | S8 | Authority row exists; Make behavior not normalized into package behavior. |
| Public environment variable precedence | S8 | Source-limited precedence rows preserved. |
| Supported platform/tool profile | S8/S6 | Static platform assumptions are not runtime guarantees. |
| Reset-route authority and partial side effects | S8/S6 | Reset mapped as interface only; authority remains open. |
| Stale janitor destructive authority | S8/S6 | Source-limited cleanup boundaries preserved. |
| Cleanup/reaper completion, timeout, interrupt, and parent death | S6 | Source-only cleanup paths are not guaranteed. |
| Retained-run freshness and external cache cleanup | S8/S6 | Retained artifacts remain selected-run/source-limited evidence. |

## Validation Evidence

Inspection commands used: `git status --short --branch`, `git rev-parse HEAD`,
`date -Is`, `uname -a`, `ls`, `rg`, `test -d .github`, and read-only `node`
scripts that parsed markdown row IDs and table shape.

No broad verification command, service-backed command, browser command, Docker
or Compose command, reset command, cleanup command, formatter, generator,
baseline refresh, release build, or controlled failure command was run.

No implementation, CI, generated, fixture, lockfile, report, schema, exit-code,
cleanup, or command-behavior file was changed by this audit. The only intended
write is this audit document.

## Final Audit Checklist

| Checklist item | Result | Notes |
|---|---|---|
| Audit objective, scope, and non-scope recorded. | pass | Recorded above. |
| All eight required S5 outputs reviewed. | pass | All present. |
| S0-S4 inputs and S5 handoff checked. | pass | Control docs, audits, and handoffs inspected. |
| Repository source evidence spot-checked against representative S5 rows. | pass | Source files and schema markers spot-checked. |
| stdout, stderr, exit codes, structured reports, logs, CI annotations, failure artifacts, and dashboards classified correctly. | pass_with_source_limits | CI/provider and failure-only artifacts remain source-limited. |
| Machine-consumed outputs linked to `SCHEMA-*` or `TODO: schema_unknown`. | pass | No missing machine/schema links found. |
| Human diagnostics kept separate from stable machine interfaces. | pass | Classification rules and schema notes preserve the split. |
| Every major entrypoint has lifecycle phases and terminal states. | pass | `LIFE-0001..LIFE-0015`. |
| Partial-completion states exist where applicable. | pass | `PCS-0001..PCS-0019`. |
| Failure modes are specific, evidence-grounded, and consistently classified. | pass | Register shape is complete and linked-output/schema references are corrected. |
| Product assertion failures are separated from harness operational failures. | pass | Taxonomy and `FAIL-0011` preserve the split. |
| Retryability is explicit for every class and mode. | pass | Taxonomy and failure-mode rows include retryability. |
| Unsupported assumptions, ambiguity, missing evidence, and source limits are preserved. | pass | Preserved in S5 rows and registers. |
| Blockers and follow-ups are classified with S6/S7/S8 routing. | pass | Recorded above. |
| Final verdict states whether S6 may proceed. | pass | Verdict is `pass_after_s5_link_correction`; original verdict was `block_s6`. |
| No prohibited implementation or behavior change occurred. | pass | Audit wrote this recovery-doc file only. |
