---
doc_id: THR-S5-OUTPUT-CONSUMERS
title: Testing Harness Recovery Output Consumer Map
status: active
role: output-consumer-map
---

# Testing Harness Recovery Output Consumer Map

## Document role

This S5 artifact identifies which harness outputs are consumed by humans,
machines, later harness stages, or authority-review workflows. It preserves the
distinction between stable machine inputs, diagnostic logs, retained artifacts,
and authority-ambiguous package/local/CI surfaces.

## Consumer map

| Consumer ID | Consumer | Reads | Purpose | Stable input requirement | Linked outputs | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| CONS-0001 | Make phase/target wrappers. | Child exit statuses, phase summaries, target summaries, stdout/stderr logs. | Propagate child success/failure while creating bounded summaries. | Must run through generated Make recipes or wrapper scripts with `CARTULARY_TEST_TARGET`/result-root context. | `OI-0001`, `OI-0002` | `tools/task_surface.generated.mk`; `scripts/lib/run-phase-common.sh`; `scripts/cartulary-runner.mjs` | `observed` | Direct scripts may bypass wrapper output unless explicitly invoked through Make. |
| CONS-0002 | Aggregate Make sequences. | Child Make target statuses, target summaries, run summary inputs, helper unit status. | Implement `test-fast`, `test`, `ci`, and `release-check` serial/parallel orchestration. | Sequence must resolve from `tools/task_surface_manifest.json`; child targets must emit expected summaries or be reported missing/skipped. | `OI-0003` | `scripts/run-make-sequence.sh`; `tools/task_surface_manifest.json` | `observed` | First failing child aborts later sequence steps after failure summary emission. |
| CONS-0003 | Check scheduler. | `tools/check_schedule_manifest.json`, resource registry, child work outputs, nested scheduler summaries/events. | Drive `make check` and downstream `ci`/`release-check` work. | Generated schedule must match supported schema; resource overrides must validate; nested summaries are selected by current run ID. | `OI-0004`, `OI-0005` | `scripts/run-check-schedule.mjs`; scheduler engine/reporting; `RES-0001..RES-0010` | `observed/source_limit` | Scheduler lanes are logical scheduling resources, not proof of concrete service capacity. |
| CONS-0004 | Service-backed scheduler. | Service-backed schedule manifest, browser batch manifest, Go shard plan, child target summaries, browser session files. | Run backend service-backed and browser stage work under declared resource claims. | Schedule must validate and service wrapper/session env must be available where required. | `OI-0005`, `OI-0006`, `OI-0008`, `OI-0010` | `scripts/run-service-backed-schedule.mjs`; `tools/service_backed_schedule_manifest.json` | `observed/source_limit` | Service runtime behavior remains source-limited. |
| CONS-0005 | `tools/testservices` wrapper and reaper. | Service lease, env files, suite events, `service-scope.json`, Docker labels, child exit status. | Start/attach/stop suite-owned Postgres and object-store, record fixture/service diagnostics, schedule cleanup. | Active suite env and lease files must match current suite; detached reaper completion is not caller-observed in S5. | `OI-0010` | `tools/testservices/main.go`; `internal/testutil/suiteservices/diagnostics.go` | `observed/source_limit` | Cleanup guarantee strength is routed to S6; destructive authority to S8. |
| CONS-0006 | Browser owned-stack scripts. | Stack env/json, session lease/env, backend/frontend process logs, reset artifacts. | Start, reuse, stop, and diagnose browser E2E stack. | Lease/env must be produced by matching `start-web-e2e.sh` session; reset route requires test-enabled backend. | `OI-0008`, `OI-0009` | `scripts/start-web-e2e.sh`; `scripts/run-browser-e2e-batch.sh`; `scripts/reset-web-e2e-stack.sh` | `observed/source_limit` | Parent death, signal, port release, and live readiness remain S6/source-limited. |
| CONS-0007 | Playwright and Vitest phase summarizers. | Tool-defined JSON reports, stdout/stderr logs, watchdog JSON, selected-test manifests. | Convert frontend/browser test results into harness phase summaries and failure dossiers. | Runner JSON must be present; absent runner JSON is classified as artifact failure. | `OI-0007`, `OI-0008` | `scripts/run-frontend-unit.sh`; `scripts/lib/test-output/cli.mjs`; Playwright configs | `observed/source_limit` | Tool report schemas are partially external/tool-defined. |
| CONS-0008 | Go phase summarizer and shard finalizer. | Go `-json` runner logs, shard metadata, exit status files, manifest scope env. | Attribute Go package/test failures, validate manifest coverage, produce aggregate target summaries. | Runner JSONL and shard metadata must be tied to selected run/shard metadata dir. | `OI-0006` | `scripts/lib/go-target-runner.mjs`; `scripts/lib/test-output/cli.mjs` | `observed/source_limit` | Runtime failure examples were not generated during S5. |
| CONS-0009 | Explain and investigation tools. | Retained run summaries, target/phase/tool summaries, scheduler events/progress logs, fixture summaries. | Human investigation and selected-run diagnostics. | Must report selected `RESULTS_DIR`/run ID before claims; newest fallback is investigation-only. | `OI-0001` through `OI-0011`, `OI-0016` | `scripts/lib/artifact-discovery.mjs`; `scripts/lib/make-node-tools.mjs`; `ART-0013..ART-0017` | `observed/source_limit` | Retained artifacts may be stale or incomplete without explicit selection. |
| CONS-0010 | Drift and baseline tools. | Duration baselines, retained successful run timing, scheduler summaries/events, generated schedules. | Validate and refresh execution planning inputs. | Must use explicit `RESULTS_DIR` or active run context; successful-run provenance must be known. | `OI-0003`, `OI-0004`, `OI-0005`, `OI-0011` | S3 `ART-0011`, `ART-0013`, `ART-0015`; duration baseline scripts | `observed/source_limit` | Baseline refresh is mutating and was not run. |
| CONS-0011 | Fixture report tooling. | Suite-service events, `service-scope.json`, fixture activity summaries. | Summarize DB/object fixture costs and cleanup evidence. | Selected retained run should be explicit; missing service events limit report completeness. | `OI-0010`, `OI-0011` | `scripts/lib/fixture-reporting.mjs`; `scripts/print-fixture-report.mjs`; suiteservices diagnostics | `observed/source_limit` | S5 did not run fixture-report against a selected run. |
| CONS-0012 | CI/release consumers. | `make ci`/`make release-check` run summaries, target summaries, deployable-shape output, release artifacts. | Provider-neutral CI and release verification. | Repository-local CI source is `scripts/ci/**` plus Make; provider workflow and annotations are absent/source-limited. | `OI-0003`, `OI-0014`, `OI-0015` | `scripts/ci/verify.sh`; `scripts/ci/check-deployable-shape.sh`; `SL-0001` | `observed/source_limit` | No `.github/**` workflow exists to map uploads or annotations. |
| CONS-0013 | Direct package-script users. | Tool stdout/stderr, Vitest/Playwright/Biome/Vite default reports, wrapper outputs when scripts re-enter harness. | Local developer feedback outside Make. | No Make run ID, result-root, scheduler resource, or target-summary guarantee unless child wrapper adds it. | `OI-0013` | root `package.json`; `apps/web/package.json`; `ART-0030`; `AMB-0020` | `observed` | S8 must decide whether these are first-class contracts. |
| CONS-0014 | Maintainers and NLSpec authors. | S0-S6 recovery docs, ambiguity/source-limit rows, authority map, preservation matrix, S5 maps. | Decide what observed behavior becomes normative and what remains roadmap or owner-required. | Evidence labels and source limits must remain intact; generated outputs cannot become behavior owners by accident. | all `OI-*`; all `SCHEMA-*` | recovery process docs; `docs/domain.md`; Core 00-04 authority posture | `observed` | S5 outputs are direct inputs to S7 NLSpec and S8 authority decisions. |

## Consumer closure notes

- All machine consumers above require either a source-declared schema, a linked
  `SCHEMA-*` note, or explicit `TODO: schema_unknown`.
- Outputs read by humans only remain diagnostic unless a source consumer parses
  them.
- Retained artifact consumers must avoid ambient stale-run truth. Use explicit
  `RESULTS_DIR`, active `CARTULARY_TEST_RUN_ID`, or a reported selected run ID.

## Checklist Support

| Field | Value |
|---|---|
| Status | `complete` |
| Blockers | Provider CI consumers, direct package-script contract status, and selected retained-run consumers remain source/authority-limited. |
| Findings | `CONS-0001` through `CONS-0014` identify readers, purpose, stable input requirements, linked outputs, and evidence status. |
| Source limits | Retained-artifact provenance, CI provider workflows/annotations, service/browser runtime artifacts, and direct package-tool output schemas. |
| Ambiguities | Direct package scripts, provider CI annotations, local-dev services, reset route, and maintenance/cleanup ownership. |
| Handoff notes | S7 can use this as the consumer matrix; S8 should decide authority-ambiguous consumers before normative contract drafting. |
| Evidence status | Rows use `observed` or `observed/source_limit`. |
| No-change confirmation | This file records recovery documentation only and does not alter harness behavior. |
