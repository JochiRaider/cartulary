---
doc_id: THR-S5-OBSERVABLE-INTERFACES
title: Testing Harness Recovery Observable Interface Map
status: active
role: observable-interface-map
---

# Testing Harness Recovery Observable Interface Map

## Document role

This S5 artifact maps caller-visible and machine-consumed interfaces needed to
route the S4 `pass_with_followups` gaps. It is recovery evidence only. Static
S4 lifecycle rows remain `observed` or `observed/source_limit`; this document
does not upgrade live readiness, cleanup, timing, or capacity behavior to
`runtime_observed`.

No service-backed targets, browser E2E targets, Docker/testcontainers, Docker
Compose, reset routes, cleanup commands, formatters, generators, broad gates, or
runtime evidence commands were executed for this S5 follow-up pass.

## Observable interface map

| Interface ID | Entrypoint IDs | Output surface | Consumer | Machine-consumed | Schema/path | Ordering guarantee | Stable across CI/local | Authority class | Linked S4 gaps | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| OI-0001 | `EP-0002`, `EP-0008`, `EP-0014`, `EP-0015` | Target, phase, run, timing, and tool summaries under `.cartulary/test-results/**`. | Explain, drift, baseline, investigation, and human summary tools. | yes | `ART-0013`, `ART-0014`; exact schema still `TODO: schema_unknown` per S5 scope. | Source shows writers create or refresh per target/run, but retained-run freshness is source-limited. | mostly, when invoked through Make. | `derived_report` | `AUD-S4-FU-0001`, `AUD-S4-FU-0005` | `artifact-ownership-matrix.md`; `entrypoint-command-map.md`; `scripts/lib/artifact-discovery.mjs` | `observed/source_limit` | Ambient newest-run lookup must not become current-run truth without selected run identity. |
| OI-0002 | `EP-0011`, `EP-0006`, `EP-0007` | `tools/testservices` service scope, cleanup summaries, fixture summaries, event streams, and timing spans. | Fixture report, target summaries, service diagnostics, cleanup decisions. | yes | `ART-0017`, `ART-0029`; service summary schema notes in `structured-output-schema-notes.md`. | Source records setup/teardown events; live completion ordering remains source-limited. | local/check when service wrapper is used. | `derived_report` | `AUD-S4-FU-0001`, `AUD-S4-FU-0002`, `AUD-S4-FU-0007` | `tools/testservices/main.go`; `internal/testutil/suiteservices/diagnostics.go`; `SVC-0001` | `observed/source_limit` | Reaper scheduled summaries are not proof of detached reaper completion. |
| OI-0003 | `EP-0009`, `EP-0010`, `EP-0018` | Browser owned-stack env/json files, server logs, Vite logs, Playwright stdout, reports, traces, screenshots, and videos. | Browser target summaries, manual triage, future failure taxonomy. | mixed | `ART-0018`, `ART-0027`; Playwright report schemas remain `TODO: schema_unknown`. | Stack files are written before child command; Playwright failure artifacts are failure-path only. | differs for Make versus direct package scripts. | `diagnostic` unless consumed by wrapper | `AUD-S4-FU-0001`, `AUD-S4-FU-0002`, `AUD-S4-FU-0006` | `scripts/start-web-e2e.sh`; Playwright configs; `ART-0018`, `ART-0027` | `observed/source_limit` | Browser E2E was not run; failure bundle completeness remains source-limited. |
| OI-0004 | `EP-0009`, `EP-0010`, `EP-0020` | Reset boundary response JSON, status file, and state-marker file. | Browser reset wrapper, later failure analysis, human diagnostics. | yes for reset script validation | `RES-0023`, `RES-0026`; response uses `cartulary.test.runtime_reset.v1`. | Source validates HTTP status and JSON after reset call; route-specific readiness and timeout are unknown. | only when test routes are enabled. | `derived_report` plus authority gap | `AUD-S4-FU-0001`, `AUD-S4-FU-0003` | `scripts/reset-web-e2e-stack.sh`; `internal/app/test_runtime_reset.go`; `SVC-0014` | `observed/source_limit` | S8 must decide reset-route authority before this becomes normative harness behavior. |
| OI-0005 | `EP-0002`, `EP-0010` | Docker Compose/dev-service stdout, stderr, readiness diagnostics, and recent compose logs on failure. | Local developer and standalone browser diagnostics. | no by default | stdout/stderr only; no stable structured schema observed. | Source loops until ready or container missing/exited/dead; live Compose behavior not executed. | local-dev only unless S8 decides otherwise. | `diagnostic` | `AUD-S4-FU-0001`, `AUD-S4-FU-0006`, `AUD-S4-FU-0008` | `scripts/dev-services.sh`; `docker-compose.dev.yml`; `SVC-0008` | `observed/source_limit` | Compose stop/volume cleanup remains outside recovered contract unless S8 expands scope. |
| OI-0006 | `EP-0016`, `EP-0018`, `EP-0019` | Direct package-script stdout/stderr and tool-default artifacts. | Developer and package-tool users. | mixed/unknown | package-tool defaults; `ART-0030`. | Package scripts can bypass Make run ID, result root, output policy, and scheduler limits. | not stable against Make until S8 decides. | `unknown_authority` | `AUD-S4-FU-0004`, `AUD-S4-FU-0005` | root `package.json`; `apps/web/package.json`; `HAZ-S4-0009` | `observed` | Default treatment is developer convenience except where scripts re-enter harness wrappers. |
| OI-0007 | `EP-0006`, `EP-0007`, `EP-0012` | Scheduler usage errors, resource-limit validation failures, scheduler summaries, and events. | Make targets, explain tools, CI/release gates. | yes | `ART-0015`; scheduler manifest/schema notes. | Work-unit ordering follows manifest `needs` and retained resource claims in source. | yes when scheduler manifests are current. | `derived_report` and execution input | `AUD-S4-FU-0007` | scheduler scripts/manifests; `RES-0001` through `RES-0010` | `observed/source_limit` | Logical lanes are visible interfaces, but concrete service capacity is not proved. |
| OI-0008 | `EP-0004`, `EP-0005`, `EP-0006`, `EP-0011`, `EP-0014` | Go target logs, runner JSONL, watchdog JSON, package/test output, and service fixture diagnostics. | Failure triage, explain tools, fixture report, future failure taxonomy. | mixed | `ART-0016`, `ART-0017`; schema still `TODO: schema_unknown` for failure-only bundles. | Source writes by runner and wrapper paths; failure-path freshness requires selected run evidence. | yes through Make wrappers, unknown for package/direct paths. | `diagnostic` unless consumed | `AUD-S4-FU-0001`, `AUD-S4-FU-0002` | `go-target-runner` references in S3; `tools/testservices/main.go` | `observed/source_limit` | No controlled failing run was generated in this pass. |

## Follow-up classification

| S4 follow-up | Interface disposition |
|---|---|
| `AUD-S4-FU-0001` | S5 records observable readiness and failure outputs only as static interfaces; live readiness evidence remains S6 or controlled runtime scope. |
| `AUD-S4-FU-0002` | S5 records cleanup summaries, logs, and failure outputs; cleanup strength belongs to S6. |
| `AUD-S4-FU-0003` | S5 records reset response artifacts; reset route authority belongs to S8. |
| `AUD-S4-FU-0004` | S5 records package-script observable differences; first-class authority belongs to S8. |
| `AUD-S4-FU-0005` | S5 records env-related visible outputs and missing-env failures; public precedence belongs to S8. |
| `AUD-S4-FU-0006` | S5 records missing-tool and platform diagnostics where source-observed; supported profile belongs to S8. |
| `AUD-S4-FU-0007` | S5 records scheduler lane outputs; concrete capacity belongs to S6. |
| `AUD-S4-FU-0008` | S5 records local-dev outputs; lifecycle authority belongs to S8 and hazards belong to S6. |
| `AUD-S4-FU-0009` | S5 records Go-cache-visible paths only; cleanup authority belongs to S8. |

