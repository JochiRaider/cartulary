---
doc_id: THR-S5-OUTPUT-CONSUMERS
title: Testing Harness Recovery Output Consumer Map
status: active
role: output-consumer-map
---

# Testing Harness Recovery Output Consumer Map

## Document role

This S5 artifact identifies which outputs from S4-related harness surfaces are
machine-consumed, diagnostic-only, or authority-ambiguous.

## Consumer map

| Consumer ID | Consumer | Reads | Purpose | Stable input requirement | Linked outputs | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| CONS-0001 | `make explain-run` and artifact discovery helpers. | Retained run summaries, target summaries, scheduler events/logs. | Human investigation and selected-run diagnostics. | Must report selected `RESULTS_DIR`/run identity before claims. | `OI-0001`, `OI-0007`, `OI-0008` | `ART-0013` through `ART-0016`; `entrypoint-command-map.md` | `observed/source_limit` | Newest fallback is investigation-only until S8 decides otherwise. |
| CONS-0002 | Drift and baseline checks. | Duration baselines, retained successful run timing, scheduler events. | Validate committed planning inputs. | Explicit `RESULTS_DIR` or active run context; no ambient stale run truth. | `OI-0001`, `OI-0007` | `ART-0011`, `ART-0013`, `ART-0015` | `observed/source_limit` | Successful-run provenance remains S5/S8-sensitive. |
| CONS-0003 | `make fixture-report`. | Suite service events and fixture summaries. | Summarize fixture/resource use. | Explicit retained run selection should be preferred. | `OI-0002`, `OI-0008` | `ART-0017`, `ART-0029` | `observed/source_limit` | Live fixture report was not run here. |
| CONS-0004 | Service-backed and check schedulers. | Committed schedule manifests and resource registry. | Drive work-unit graph and resource claims. | Generated schedules must be drift-current. | `OI-0007` | `ART-0010`, `RES-0001` through `RES-0010` | `observed` | Generated outputs drive execution but are not behavioral owners. |
| CONS-0005 | Browser reset wrapper. | Reset route HTTP response, status file, state marker. | Validate reset boundary and clear Playwright state dir. | Test-enabled backend only; route authority open. | `OI-0004` | `scripts/reset-web-e2e-stack.sh`; `SVC-0014` | `observed/source_limit` | Wrong-backend risk remains S8 authority/hazard work. |
| CONS-0006 | Browser/session scripts. | Stack env/json files and session lease. | Reuse or stop an owned browser stack. | Lease/env file must match active stack; runtime not observed here. | `OI-0003` | `scripts/start-web-e2e.sh`; `RES-0021` | `observed/source_limit` | Parent death and port release behavior remain S6. |
| CONS-0007 | Direct package-script users. | Tool stdout/stderr and default report paths. | Local developer feedback. | No Make result-root guarantee unless wrapper re-enters harness. | `OI-0006` | package manifests; `HAZ-S4-0009` | `observed` | S8 decides first-class contract status. |
| CONS-0008 | Maintainer/spec reviewer. | Recovery docs, authority map, source limits, ambiguity rows. | Decide preservation and future normative contracts. | Evidence labels must remain intact. | all S5/S6/S8 docs | recovery process and charter | `observed` | Owner decisions must be recorded before closing authority gaps. |

