---
doc_id: THR-S5-PHASE-TRANSITIONS
title: Testing Harness Recovery Phase Transition Table
status: active
role: phase-transition-table
---

# Testing Harness Recovery Phase Transition Table

## Phase transition table

| Transition ID | Lifecycle ID | From phase | To phase | Gate condition | Failure transition | Cleanup transition | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|
| PT-0001 | `LIFE-0001` | suite preflight | service startup | Docker preflight and artifact path validation pass. | startup failure summary; no child command. | cleanup status `startup_failed`. | `SVC-0001`, `SVC-0007` | `observed/source_limit` | Live Docker unavailable behavior not executed. |
| PT-0002 | `LIFE-0001` | service startup | child execution | Postgres/template and MinIO ready in source-observed order. | service or template failure summary. | started services cleaned or reaper scheduled. | `SVC-0002`, `SVC-0003`, `SVC-0005` | `observed/source_limit` | Live readiness remains `SL-0012`. |
| PT-0003 | `LIFE-0001` | child execution | cleanup/reaper | child process exits or child start fails. | child-start or child exit failure propagated. | `cleanupOwnedServices` path. | `tools/testservices/main.go`; `CLN-0011` | `observed/source_limit` | Parent death and interrupt behavior unobserved. |
| PT-0004 | `LIFE-0002` | manifest load | work scheduling | manifest and resource limits validate. | usage/config failure before child work. | no service cleanup unless session started. | scheduler scripts; `RES-0001` through `RES-0010` | `observed` | Concrete resource capacity is separate. |
| PT-0005 | `LIFE-0003` | port/runtime allocation | backend startup | runtime root and ports selected; service fixture ready. | setup failure with logs/summary. | cleanup already-started resources. | `SVC-0009`, `RES-0017`, `RES-0021` | `observed/source_limit` | Dynamic port race remains S6 hazard. |
| PT-0006 | `LIFE-0003` | backend ready | frontend startup | `/readyz` succeeds before frontend readiness wait completes. | backend readiness timeout/early exit. | process group cleanup. | `scripts/start-web-e2e.sh`; `SVC-0009` | `observed/source_limit` | No browser runtime pass was run. |
| PT-0007 | `LIFE-0003` | frontend ready | browser child/session | frontend origin responds. | frontend timeout/early exit. | process group cleanup. | `scripts/start-web-e2e.sh`; `SVC-0010` | `observed/source_limit` | Port release not runtime-observed. |
| PT-0008 | `LIFE-0004` | reset request | reset validation | HTTP status and JSON satisfy reset script checks. | reset script exits non-zero and leaves artifacts. | Playwright state clear only after validation path in source. | `scripts/reset-web-e2e-stack.sh`; `SVC-0014` | `observed/source_limit` | Reset route call not run; route authority open. |
| PT-0009 | `LIFE-0005` | compose startup | local readiness | compose service health/wait loops pass. | recent compose logs printed on missing/exited/dead/timeout. | no compose-down cleanup observed. | `scripts/dev-services.sh`; `SVC-0008` | `observed/source_limit` | Local-dev lifecycle authority open. |
| PT-0010 | `LIFE-0006` | package script start | tool output | pnpm child command starts. | tool-specific non-zero exit. | tool-specific cleanup only. | package manifests; `HAZ-S4-0009` | `observed` | Make guarantees do not automatically apply. |

