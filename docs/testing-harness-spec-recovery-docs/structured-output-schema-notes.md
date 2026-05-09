---
doc_id: THR-S5-STRUCTURED-OUTPUT-SCHEMAS
title: Testing Harness Recovery Structured Output Schema Notes
status: active
role: structured-output-schema-notes
---

# Testing Harness Recovery Structured Output Schema Notes

## Document role

This S5 artifact records schema-bearing outputs that matter to the S4 follow-up
gaps. It does not claim schema completeness for runtime-only, failure-only, or
retained local artifacts that were not selected from an authorized run.

## Schema notes

| Schema note ID | Surface | Producer | Consumer | Known schema indicator | S4-linked rows | Completeness | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| SCHEMA-0001 | Target, phase, run, timing, and tool summaries. | Make wrappers, runner context, schedulers, test-output helpers. | `explain-run`, drift tools, summaries, humans. | JSON summary paths under `ART-0014`; exact fields remain source-observed from helpers. | `ENV-0001`, `RES-0011` | `partial` | `artifact-ownership-matrix.md`; `entrypoint-command-map.md`; runner helper sources cited by S3/S4 | `observed/source_limit` | Needs selected retained run or source schema extraction before NLSpec. |
| SCHEMA-0002 | Scheduler summaries and event streams. | `scripts/lib/scheduler/engine.mjs`, check/service-backed runners. | scheduler drift, explain tools, target summaries. | `scheduler-summary.json`, `scheduler-events.jsonl`, `progress-summary.log`. | `RES-0001` through `RES-0010` | `partial` | `ART-0015`; scheduler manifests and runner scripts | `observed/source_limit` | Event ordering is S6 timing scope. |
| SCHEMA-0003 | Suite service scope and cleanup summaries. | `tools/testservices`, `internal/testutil/suiteservices`. | fixture reports, target summaries, cleanup decisions. | Service scope JSON and event files under suite artifact dir. | `SVC-0001`, `RES-0011`, `RES-0012` | `partial` | `tools/testservices/main.go`; `internal/testutil/suiteservices/diagnostics.go` | `observed/source_limit` | Runtime cleanup summaries were not produced during this pass. |
| SCHEMA-0004 | Fixture report output. | `make fixture-report`, fixture-reporting helpers. | human and machine investigation. | `cartulary.fixture_report.v1`. | `ART-0017` | `partial` | `scripts/lib/fixture-reporting.mjs`; `internal/testutil/suiteservices/diagnostics.go` | `observed/source_limit` | Retained artifact selection remains source-limited. |
| SCHEMA-0005 | Browser owned stack metadata and session lease. | `scripts/start-web-e2e.sh`. | browser scheduler/session stop, diagnostics. | `cartulary.web_e2e_session_lease.v1` for lease JSON; env/json stack files. | `SVC-0009`, `SVC-0010`, `RES-0021` | `source_observed` | `scripts/start-web-e2e.sh` | `observed` | Live stack files were not generated in this pass. |
| SCHEMA-0006 | Reset boundary response. | `internal/app/test_runtime_reset.go`; `scripts/reset-web-e2e-stack.sh`. | reset wrapper validation and diagnostics. | `cartulary.test.runtime_reset.v1`. | `SVC-0014`, `RES-0023`, `RES-0026` | `source_observed` | `internal/app/test_runtime_reset.go`; `scripts/reset-web-e2e-stack.sh` | `observed/source_limit` | Authority and route-specific readiness remain open. |
| SCHEMA-0007 | Playwright reports, traces, screenshots, and videos. | Playwright runner/configs. | human triage and future failure recovery. | Playwright tool-defined report and trace formats. | `SVC-0013`, `RES-0022` | `schema_unknown` | Playwright config and `ART-0018` | `observed/source_limit` | No browser failure artifact was generated or selected. |
| SCHEMA-0008 | Runner logs, watchdog JSON, stdout/stderr logs. | Go target runner, harness smoke, test output helpers. | failure triage and explain tools. | JSONL/watchdog/log paths under `ART-0016`. | `SVC-0001`, `SVC-0004` | `schema_unknown` | `ART-0016`; S3 source limits | `observed/source_limit` | Needs selected failure evidence before stable schema claims. |
| SCHEMA-0009 | Package-script tool outputs. | pnpm, Vite, Vitest, Biome, Playwright. | direct tool users. | tool-defined stdout/stderr and report roots. | `HAZ-S4-0009`, `AMB-0020` | `authority_unknown` | package manifests; S2 package-script table | `observed` | S8 must decide whether these are harness interfaces. |

## Schema closure rule

Future NLSpec drafting must not promote `partial`, `schema_unknown`, or
`authority_unknown` schema notes into stable contracts until either source schema
extraction is complete, selected retained artifacts are inspected, controlled
runtime evidence is authorized, or a maintainer decision is recorded.

