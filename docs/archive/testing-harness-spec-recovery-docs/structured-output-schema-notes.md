---
doc_id: THR-S5-STRUCTURED-OUTPUT-SCHEMAS
title: Testing Harness Recovery Structured Output Schema Notes
status: active
role: structured-output-schema-notes
---

# Testing Harness Recovery Structured Output Schema Notes

## Document role

This S5 artifact records schema-bearing outputs that are caller-visible or
machine-consumed by the testing harness. It is a schema recovery note, not a
schema change request. Where the source exposes a stable `schema_id`, that
identifier is recorded. Where the output is tool-defined, failure-only,
retained-but-unselected, or authority-ambiguous, the completeness field remains
`partial`, `schema_unknown`, `authority_unknown`, or `source_limit`.

## Schema completeness values

| Completeness | Meaning |
|---|---|
| `complete_source_marker` | Source declares a `schema_id` and the producing/consuming surface is identified. Field-level validation may still belong to later NLSpec work. |
| `source_observed` | Source writes/reads a structured shape but full field contract was not exhaustively extracted. |
| `partial` | Shape is partly recovered from source and/or retained artifacts, but S5 cannot claim full schema. |
| `schema_unknown` | Machine-consumed or potentially machine-consumed output exists, but stable schema was not recovered. |
| `authority_unknown` | Output exists, but whether it is a first-class harness interface requires owner decision. |

## Schema notes

| Schema note ID | Surface | Producer | Consumer | Known schema indicator | Linked outputs | Completeness | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| SCHEMA-0001 | Phase summary. | `scripts/lib/test-output/cli.mjs` via `shell-phase`, `go-phase`, `vitest-phase`, `playwright-phase`, and manifest-aware variants. | Target summaries, run summaries, explain tools, humans. | `cartulary.test_phase_summary.v3`; `phase-summary.json`; companion `meta.json`. | `OI-0001`, `OI-0006`, `OI-0007`, `OI-0008` | `complete_source_marker` | `scripts/lib/test-output/context.mjs`; `scripts/lib/test-output/cli.mjs`; `scripts/lib/run-phase-common.sh` | `observed` | Includes label, target, runner, status, command, timing, counts, artifacts, owners, inventory, dossiers, and failure fields. |
| SCHEMA-0002 | Shell phase stdout/stderr logs. | `run_phase_command` and `finalizeShellPhase`. | Humans, failure records, target/tool summaries through raw artifact refs. | Log paths only; `TODO: schema_unknown` for log contents. | `OI-0001`, `OI-0015` | `schema_unknown` for contents; path schema via phase summary. | `scripts/lib/run-phase-common.sh`; `scripts/lib/test-output/cli.mjs` | `observed` | Empty logs may be removed; failure headline is derived from known tool diagnostics or first actionable line. |
| SCHEMA-0003 | Target summary. | `test-output.mjs target-summary`; summary-target wrappers. | Run summaries, check/CI/release gates, explain tools, drift tools, humans. | `cartulary.test_target_summary.v4`; `target-summary.json`. | `OI-0002`, `OI-0003`, `OI-0004`, `OI-0005`, `OI-0012` | `complete_source_marker` | `scripts/lib/test-output/context.mjs`; `scripts/lib/test-output/cli.mjs`; S2 `EP-0015` | `observed` | Includes own/children/totals rollups, artifacts, failures, accounting, scheduler timing, and status. |
| SCHEMA-0004 | Run summary. | `test-output.mjs run-summary`; aggregate sequence/check scheduler. | Aggregate callers, CI/release gates, explain tools, baseline/drift consumers. | `cartulary.test_run_summary.v6`; `run-summary.json`. | `OI-0003`, `OI-0004`, `OI-0014` | `complete_source_marker` | `scripts/lib/test-output/context.mjs`; `scripts/lib/test-output/cli.mjs`; `scripts/run-make-sequence.sh` | `observed` | Records work units, evidence targets, summary groups, helper units, shared execution groups, failures, fixture summary, and selected artifact dir. |
| SCHEMA-0005 | Scheduler summary and events. | Check and service-backed scheduler runners plus scheduler engine/reporting. | Scheduler drift, explain-run, target summaries, duration/timing drift, humans. | `cartulary.check_scheduler_summary.v10`, `cartulary.scheduler_event.v6`, `cartulary.service_backed_scheduler_summary.v10`, `cartulary.scheduler_event.v6`. | `OI-0004`, `OI-0005` | `complete_source_marker` | `scripts/run-check-schedule.mjs`; `scripts/run-service-backed-schedule.mjs`; `scripts/lib/scheduler/**`; scheduler smoke tests | `observed/source_limit` | Event order/timing validity is S6; S5 records the structured interface only. |
| SCHEMA-0006 | Tool run summary and terminal lines. | `scripts/lib/tool-output.mjs`; `test-output` phase/target/run summary handlers. | Humans, machine output mode callers, explain tools. | `cartulary.tool_run_summary.v2`; `tool-run-summary.json`; `[RESULT]`, `[ARTIFACTS]`, `[FAIL]`, `[RERUN]`, `[INVESTIGATE]`. | `OI-0002`, `OI-0003`, `OI-0004`, `OI-0011`, `OI-0012`, `OI-0014` | `complete_source_marker` | `scripts/lib/tool-output.mjs`; `scripts/lib/test-output/cli.mjs` | `observed` | Terminal text is human diagnostic; JSON tool summary is machine-consumed. |
| SCHEMA-0007 | Go runner JSONL and shard metadata. | `tools/harness/backend/go-target-runner.mjs`; Go `-json` stream capture. | Go phase summarizer, shard finalizer, humans. | `runner.jsonl`, shard `exit_status.txt`, metadata dirs; no single top-level schema ID. | `OI-0006` | `partial` | `tools/harness/backend/go-target-runner.mjs`; `scripts/lib/test-output/cli.mjs` go summarizer | `observed/source_limit` | Go `-json` lines are Go tool-defined; harness metadata shape needs later extraction before NLSpec. |
| SCHEMA-0008 | Vitest runner JSON and watchdog JSON. | `scripts/run-frontend-unit.sh`; `run_vitest_command_with_watchdog`; Vitest JSON reporter. | Vitest phase summarizer, humans. | Watchdog `cartulary.vitest_watchdog.v1`; Vitest `runner.json` is tool-defined. | `OI-0007` | `partial` | `scripts/lib/run-phase-common.sh`; `scripts/run-frontend-unit.sh`; `scripts/lib/test-output/cli.mjs` | `observed/source_limit` | Watchdog field shape is source-written; Vitest report schema is external/tool-defined. |
| SCHEMA-0009 | Browser E2E stack metadata. | `scripts/start-web-e2e.sh`. | Browser wrappers, session scripts, humans. | `cartulary.web_e2e_stack.v2`; `stack.json`; `stack.env`. | `OI-0008` | `complete_source_marker` | `scripts/start-web-e2e.sh` | `observed` | Contains origins, ports, runtime root, process group IDs, readiness proof, server log, and web log paths. |
| SCHEMA-0010 | Browser session lease and reset boundary response. | `scripts/start-web-e2e.sh`, `scripts/reset-web-e2e-stack.sh`, `internal/testutil/testruntime/reset.go`. | Session stop, reset wrapper validation, browser batch, humans. | `cartulary.web_e2e_session_lease.v1`; reset data `cartulary.test.runtime_reset.v1`; response/status files. | `OI-0008`, `OI-0009` | `source_observed` | `scripts/start-web-e2e.sh`; `scripts/reset-web-e2e-stack.sh`; `internal/testutil/testruntime/reset.go`; `MD-S7-0004` | `observed/source_limit` | Reset route ownership is harness-owned; partial mutation behavior remains source-limited. |
| SCHEMA-0011 | Go test phase summaries and manifest mismatch artifacts. | `test-output.mjs go-phase` / `go-manifest-phase`. | Target summaries, phase ledger/check tooling, humans. | Phase summary schema plus `manifest-summary.json`, `manifest-mismatch.json`. | `OI-0006` | `source_observed` | `scripts/lib/test-output/cli.mjs`; phase manifest code | `observed` | Manifest mismatch is classified as artifact failure and exits non-zero through helper status. |
| SCHEMA-0012 | Vitest manifest phase summaries. | `test-output.mjs vitest-phase` / `vitest-manifest-phase`. | Target summaries, phase slices, humans. | Phase summary schema plus Vitest runner JSON refs and optional watchdog JSON. | `OI-0007` | `source_observed` | `scripts/run-frontend-unit.sh`; `scripts/lib/test-output/cli.mjs` | `observed/source_limit` | Zero-selection and manifest mismatch behavior is source-observed; runtime timeout unobserved. |
| SCHEMA-0013 | Playwright reports, traces, screenshots, videos, and manifest selections. | Playwright runner/configs; harness Playwright wrappers. | Browser phase summarizer, humans, Playwright report UI. | `TODO: schema_unknown` for Playwright tool reports/traces/screenshots/videos; selected-test schema `cartulary.playwright_manifest_selection.v1` appears in smoke tests. | `OI-0008` | `schema_unknown` | `apps/web/playwright*.config.ts`; `scripts/lib/run-playwright-*.sh`; Playwright smoke tests; `ART-0018` | `observed/source_limit` | No selected failing browser bundle was inspected during S5. |
| SCHEMA-0014 | Suite service event files and service scope summary. | `internal/testutil/suiteservices.RecordEvent` and `RefreshSummary`. | Fixture report, target summaries, cleanup diagnostics, humans. | Event JSON files have typed fields but no `schema_id`; summary file `service-scope.json` is Go struct-shaped. | `OI-0010` | `source_observed` | `internal/testutil/suiteservices/diagnostics.go` | `observed/source_limit` | Field-level structs are visible; no embedded schema ID is written. |
| SCHEMA-0015 | Test-services service lease and reaper log. | `tools/testservices`. | `terminate-suite`, detached reaper, humans. | Lease JSON has `schema_id: cartulary.test_services.lease.v1`; reaper log is text. | `OI-0010` | `partial` | `tools/testservices/main.go`; `SVC-0001`; `RES-0012` | `observed/source_limit` | Reaper scheduling is source-observed; detached completion remains source-limited. |
| SCHEMA-0016 | Fixture report output. | `scripts/print-fixture-report.mjs` and fixture-reporting helpers. | Humans and machine investigation. | `cartulary.fixture_report.v1` per S3; exact field closure not extracted here. | `OI-0010`, `OI-0011` | `partial` | `scripts/lib/fixture-reporting.mjs`; `scripts/print-fixture-report.mjs`; `ART-0017` | `observed/source_limit` | Retained artifact selection must be explicit for claims. |
| SCHEMA-0017 | Node-tool JSON and explanation outputs. | `scripts/run-make-node-tool.mjs` and per-tool scripts. | Developers, agents, drift/baseline tooling. | Tool-dependent; `target-plan-json` and machine modes have structured JSON but not one shared schema. | `OI-0011`, `OI-0016` | `partial` | `scripts/lib/make-node-tools.mjs`; `scripts/print-target-plan.mjs`; explain/fixture scripts | `observed/source_limit` | S5 does not close every node-tool schema; machine-consumed tools need per-tool notes in NLSpec. |
| SCHEMA-0018 | Direct package-script tool outputs. | pnpm, Vite, Vitest, Biome, Playwright, and direct wrapper scripts. | Direct package-script users. | Tool-defined stdout/stderr and report roots; no Make summary guarantee. | `OI-0013` | `authority_unknown` | root `package.json`; `apps/web/package.json`; `ART-0030`; `AMB-0020` | `observed` | S8 must decide whether direct scripts are first-class harness interfaces. |
| SCHEMA-0019 | CI provider annotations and uploaded artifacts. | Unknown provider workflow; `scripts/ci/verify.sh` delegates to `make ci`. | External CI provider and maintainers. | `TODO: schema_unknown`; `.github/**` absent. | `OI-0014` | `schema_unknown` | `.github absent`; `scripts/ci/**`; `SL-0001` | `source_limit` | No repository source proves provider annotation format, upload paths, or CI dashboard behavior. |
| SCHEMA-0020 | Generated schedule/task manifests and resource registry. | Generation scripts and committed generated JSON files. | Make include, schedulers, explain tools, drift checks. | Manifest `schema_id` values such as `cartulary.scheduler_resource_registry.v4`, `cartulary.check_schedule.v12`, `cartulary.service_backed_schedule.v11`. | `OI-0004`, `OI-0005`, `OI-0016` | `complete_source_marker` | `tools/*manifest*.json`; scheduler runners; `ART-0010` | `observed` | Generated outputs drive execution but are downstream of owner manifests/specs. |

## Schema closure rule

Future NLSpec drafting must not promote `partial`, `schema_unknown`, or
`authority_unknown` rows into stable contracts until source schema extraction is
complete, selected retained artifacts are inspected, controlled runtime evidence
is authorized, or a maintainer decision is recorded.

S7 maintainer decision `MD-S7-0014` stabilizes known harness schemas only where
source declares a `schema_id`. Playwright report internals, CI provider
annotations, shell log contents, and tool-defined reports remain `partial`,
`schema_unknown`, `authority_unknown`, or diagnostic unless a later owner
decision adopts them.

## Checklist Support

| Field | Value |
|---|---|
| Status | `complete` |
| Blockers | Playwright failure bundle schemas, CI provider annotations, retained failure-only artifacts, and some service event field contracts remain unavailable or incomplete. |
| Findings | `SCHEMA-0001` through `SCHEMA-0020` classify every S5 machine-consumed output as `complete_source_marker`, `source_observed`, `partial`, `schema_unknown`, or `authority_unknown`. |
| Source limits | Tool-defined Go/Vitest/Playwright schemas, provider annotations, fixture report field closure, retained-artifact provenance, and detached reaper completion remain source-limited. |
| Ambiguities | Direct package-script outputs and CI provider annotations are authority/source-limited; reset response authority is unresolved. |
| Handoff notes | S7 should not turn `partial`, `schema_unknown`, or `authority_unknown` rows into stable schemas without additional evidence or owner decision. |
| Evidence status | Rows use `observed`, `observed/source_limit`, or `source_limit`. |
| No-change confirmation | This file records recovery documentation only and does not alter harness behavior or schemas. |
