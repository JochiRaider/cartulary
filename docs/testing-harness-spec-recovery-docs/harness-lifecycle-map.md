---
doc_id: THR-S5-HARNESS-LIFECYCLE
title: Testing Harness Recovery Harness Lifecycle Map
status: active
role: harness-lifecycle-map
---

# Testing Harness Recovery Harness Lifecycle Map

## Document role

This S5 artifact reconstructs S4-related harness lifecycles from static source
evidence. It records phases and outputs without claiming live timing,
readiness, cleanup, or interrupt guarantees.

## Lifecycle map

| Lifecycle ID | Entrypoint IDs | Lifecycle surface | Phase sequence | Terminal outputs | Cleanup behavior | Linked S4 rows | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| LIFE-0001 | `EP-0005`, `EP-0006`, `EP-0011` | Service-backed suite wrapper. | parse args, pass-through check, suite preflight, start Postgres/MinIO, prepare template, write lease, stale fixture janitor, run child, summarize, cleanup/reaper. | service scope, events, target summaries, child exit status. | direct cleanup or detached reaper scheduling; completion source-limited. | `SVC-0001` through `SVC-0007`, `RES-0011` through `RES-0016` | `tools/testservices/main.go`; S4 maps | `observed/source_limit` | Runtime readiness and reaper completion are S6 follow-ups. |
| LIFE-0002 | `EP-0006`, `EP-0007`, `EP-0012` | Service-backed/check schedulers. | load manifests, normalize resource limits, expand work units, start sessions/groups, run children, finalizers, summaries. | scheduler summary/events/progress, child target summaries. | resource claims released by scheduler completion in source. | `RES-0001` through `RES-0010` | scheduler scripts/manifests; S4 register | `observed` | Concrete resource capacity is not proven by lane accounting. |
| LIFE-0003 | `EP-0009`, `EP-0010`, `EP-0018` | Browser owned stack. | allocate runtime root and ports, prepare DB/bucket, start backend, wait `/readyz`, start Vite, wait frontend, run child/session, stop groups, cleanup DB/fixture/root. | stack files, logs, Playwright outputs, target summaries. | process group stop, port wait, testservices cleanup or standalone DB drop. | `SVC-0009` through `SVC-0013`, `RES-0017` through `RES-0023` | `scripts/start-web-e2e.sh`; Playwright files | `observed/source_limit` | Browser runtime not executed; process/port cleanup strength is S6. |
| LIFE-0004 | `EP-0009`, `EP-0010`, `EP-0020` | Browser reset boundary. | resolve target support dir, call reset route, write response/status, validate JSON, clear Playwright state dir. | reset JSON, status, state marker. | mutates app DB/object state and local Playwright state. | `SVC-0014`, `RES-0023`, `RES-0026` | reset script and reset route source | `observed/source_limit` | S8 must decide reset-route authority. |
| LIFE-0005 | `EP-0002`, `EP-0010` | Local dev and standalone Compose services. | `docker compose up`, wait Postgres/MinIO, optional DB reset/migration, browser standalone DB lifecycle. | stdout/stderr, compose logs on failure, local services. | no compose down target observed; standalone browser drops E2E DB only. | `SVC-0008`, `SVC-0012`, `SVC-0015`, `RES-0018`, `RES-0019` | `scripts/dev-services.sh`; `scripts/dev-stack.sh`; `docker-compose.dev.yml` | `observed/source_limit` | Local-dev authority and persistent volumes are S8. |
| LIFE-0006 | `EP-0016`, `EP-0018`, `EP-0019` | Direct package scripts. | pnpm dispatches tool or wrapper; child may enter browser/Vitest/Biome flows. | tool stdout/stderr, tool-default reports/artifacts. | mixed and authority-ambiguous. | `HAZ-S4-0009`, `ENV-0004`, `ENV-0016`, `ENV-0022` | package manifests; S2/S4 maps | `observed` | Keep separate from Make until S8 decision. |
| LIFE-0007 | `EP-0020` | Cleanup and maintenance entrypoints. | Make target prerequisites, guarded cleanup or writer command, summary emission. | removed ignored paths or updated generated/baseline files. | explicit target action only. | `RES-0025`, S3 `CLN-0001`, `CLN-0002` | `Makefile`; cleanup matrix | `observed/source_limit` | Not executed; external Go caches not cleaned by repo targets. |

