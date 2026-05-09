---
doc_id: THR-S5-PARTIAL-COMPLETION-STATES
title: Testing Harness Recovery Partial Completion State List
status: active
role: partial-completion-state-list
---

# Testing Harness Recovery Partial Completion State List

## Partial-completion states

| State ID | Surface | Partial state | Observable output | Cleanup expectation | Linked gaps | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| PCS-0001 | Service suite | Preflight failed before services start. | failure summary, no child target. | no owned service cleanup beyond summary. | `AUD-S4-FU-0001` | `tools/testservices/main.go`; tests cited by S4 | `observed` | Runtime Docker failure shape is not live-observed. |
| PCS-0002 | Service suite | MinIO started, Postgres/template failed. | startup failure summary. | MinIO cleanup path in source. | `AUD-S4-FU-0002` | `SVC-0001`, `SVC-0002`, `SVC-0005` | `observed/source_limit` | Cleanup completion on timeout/interrupt remains unknown. |
| PCS-0003 | Service suite | Services started, child command failed. | child status plus cleanup summary. | cleanup/reaper path scheduled. | `AUD-S4-FU-0002` | `SVC-0001`, `RES-0012` | `observed/source_limit` | Reaper scheduling is not completion proof. |
| PCS-0004 | Browser stack | Backend ready, frontend startup failed or timed out. | backend/web logs and startup failure. | process group stop and port wait in source. | `AUD-S4-FU-0001`, `AUD-S4-FU-0002` | `scripts/start-web-e2e.sh`; `SVC-0009`, `SVC-0010` | `observed/source_limit` | Runtime port release unobserved. |
| PCS-0005 | Browser fixture | Active testservices fixture retired but DB connection leak detected. | cleanup failed summary. | destructive cleanup blocked by leak handling in source. | `AUD-S4-FU-0002`, `AUD-S4-FU-0007` | `SVC-0011`, `HAZ-S4-0004` | `observed/source_limit` | Active-connection behavior needs S6 classification. |
| PCS-0006 | Reset boundary | Reset route returns non-200 or invalid JSON. | response/status files retained; script exits non-zero. | state dir clear may not run if validation fails. | `AUD-S4-FU-0003` | `scripts/reset-web-e2e-stack.sh`; `RES-0023` | `observed/source_limit` | Route call not executed. |
| PCS-0007 | Compose/local dev | Compose services persist after standalone browser or `db-up`. | local services/volumes remain. | no compose down target observed. | `AUD-S4-FU-0008` | `SVC-0008`, `RES-0019` | `observed/source_limit` | S8 must decide local-dev persistence authority. |
| PCS-0008 | Package script | Tool produced artifacts outside Make result root. | tool-default output only. | cleanup coverage mixed. | `AUD-S4-FU-0004` | `ART-0030`, `HAZ-S4-0009` | `observed` | Keep outside Make contract pending S8. |
| PCS-0009 | Scheduler | Work unit fails while sibling jobs drain. | scheduler summary/events and possible Make drain message. | resource claims released by scheduler completion. | `AUD-S4-FU-0007` | `entrypoint-command-map.md`; scheduler scripts | `observed/source_limit` | Runtime event order validation belongs to S6. |

