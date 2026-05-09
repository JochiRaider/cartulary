---
doc_id: THR-S7-MAINTAINER-DECISIONS-2026-05-09
title: Testing Harness Recovery Maintainer Decision Summary
status: active
role: maintainer-decision-summary
---

# Testing Harness Recovery Maintainer Decision Summary

## Document role

This file records the maintainer decisions supplied on 2026-05-09 for S7
planning and implementation. It is an input to the harness NLSpec, acceptance
matrix, roadmap, and review packet. It does not override Core 00 through Core
04 product behavior.

## Decisions recorded

| Decision ID | Topic | Decision | Affected recovery rows | Evidence status for S7 |
|---|---|---|---|---|
| MD-S7-0001 | Command authority | `make` is the sole canonical harness command surface. Direct package scripts remain developer conveniences unless they re-enter Make-owned wrappers. | `AUTH-0002`, `AUTH-0003`, `PRES-0001`, `PRES-0011`, `AMB-0011`, `AMB-0020` | `maintainer_decision` |
| MD-S7-0002 | Planned future phases | Phase 7 and Phase 8 files are in scope only as planned future work, not active current coverage. | `AMB-0004`, `SL-0005` | `maintainer_decision` |
| MD-S7-0003 | Generated artifacts | Generated task, schedule, and code artifacts are downstream execution inputs only; they do not own behavior. | `AUTH-0010`, `PRES-0002`, `AMB-0003`, `AMB-0016`, `MSC-0005` | `maintainer_decision` |
| MD-S7-0004 | Reset route ownership | The former `internal/app/test_runtime_reset.go` behavior is harness-owned, must live outside production application ownership, and must keep test behavior. Current implementation path: `internal/testutil/testruntime/reset.go`. | `AUTH-0004`, `PRES-0010`, `AMB-0006`, `MSC-0001` | `maintainer_decision` |
| MD-S7-0005 | Local development | Local-dev Compose and `make dev` are normative local verification behavior, not deployment conformance behavior. | `AUTH-0005`, `PRES-0012`, `AMB-0009`, `MSC-0003` | `maintainer_decision` |
| MD-S7-0006 | Scheduler lanes | Scheduler lanes are logical scheduling constraints only, not physical host, service, Docker, Postgres, MinIO, or browser capacity guarantees. | `AUTH-0011`, `PRES-0003`, `RES-*` | `maintainer_decision` |
| MD-S7-0007 | Stale janitors | Stale janitors may delete DBs, buckets, or containers only with generated names plus harness metadata or lease evidence and conservative age or completed-summary checks. | `AUTH-0007`, `PRES-0013`, `AMB-0019`, `AMB-0027`, `MSC-0008` | `maintainer_decision` |
| MD-S7-0008 | Cleanup strength | Cleanup is best-effort unless selected evidence proves a specific path is guaranteed. Parent-death cleanup and active DB cleanup remain source-limited. | `AUTH-0012`, `PRES-0005`, `PRES-0008`, `PRES-0009`, `SL-0014` | `maintainer_decision/source_limit` |
| MD-S7-0009 | External Go caches | External Go caches under `/tmp/cartulary-go-*` are outside default `clean` and `distclean` cleanup scope. | `AUTH-0009`, `PRES-0016`, `AMB-0021`, `ART-0024` | `maintainer_decision` |
| MD-S7-0010 | Environment contracts | Document only source-observed environment variables and defaults for now. Do not specify precedence until a future owner decision. | `AUTH-0006`, `PRES-0014`, `AMB-0012`, `AMB-0026`, `SL-0015` | `maintainer_decision/source_limit` |
| MD-S7-0011 | Platform statement | WSL/Linux is the primary observed environment. Support should remain portable across Linux environments where practical. Do not draft a full support matrix. | `AUTH-0008`, `PRES-0015`, `AMB-0025`, `AMB-0028`, `SL-0013` | `maintainer_decision/source_limit` |
| MD-S7-0012 | Retained artifacts | Durable retained-artifact claims require explicit run identity. Newest-artifact fallback is human-investigation only. | `AUTH-0013`, `PRES-0017`, `AMB-0017`, `ART-0013` through `ART-0018` | `maintainer_decision` |
| MD-S7-0013 | Fixtures and snapshots | Define controlled fixture/golden/snapshot refresh workflows. Visual snapshots remain validation-only until an owner supplies exact OS, browser, and version bounds. | `AUTH-0014`, `PRES-0018`, `AMB-0015`, `AMB-0022`, `ART-0001` through `ART-0005` | `maintainer_decision/source_limit` |
| MD-S7-0014 | Machine schemas | Stabilize currently known harness machine schemas, while preserving unknown external/tool schemas as `schema_unknown`, `partial`, or `authority_unknown`. | `SCHEMA-*`, `OI-*` | `maintainer_decision` |
| MD-S7-0015 | CI | Keep CI provider-neutral. Do not invent `.github/**` behavior. Release readiness is not required before S7 implementation proceeds. | `AUTH-0015`, `PRES-0019`, `AMB-0001`, `SL-0001`, `MSC-0010` | `maintainer_decision/source_limit` |
| MD-S7-0016 | Stale extended smoke | The stale `run-harness-smoke-extended` failure must not block phase advancement. Remove or demote it from blocking advancement paths after implementation analysis. | `FAIL-0028`, `runtime-evidence-register.md`, `03-sprint-plan.md` | `maintainer_decision/selected_runtime_observed` |
| MD-S7-0017 | Final MUST language | Final S7 `MUST` language uses the audit register evidence split: `selected_runtime_observed`, `source_observed`, and `source_limit` or `maintainer_decision_required`. | S7 deliverables | `maintainer_decision` |

## S7 deliverable paths

| Deliverable | Path |
|---|---|
| Harness NLSpec draft | `docs/testing-harness-spec-recovery-docs/harness-nlspec.md` |
| Harness acceptance matrix | `docs/testing-harness-spec-recovery-docs/harness-acceptance-matrix.md` |
| Harness implementation roadmap | `docs/testing-harness-spec-recovery-docs/harness-implementation-roadmap.md` |
| Harness review packet | `docs/testing-harness-spec-recovery-docs/harness-review-packet.md` |

## Preserved unresolved items

These rows remain open unless a later owner decision or selected evidence closes
them: env-var precedence, visual snapshot refresh OS/browser/version, parent
death cleanup, active DB cleanup, CI provider annotations and uploads,
Playwright report internals, and any final normative `MUST` statement that lacks
selected evidence or an explicit maintainer decision.
