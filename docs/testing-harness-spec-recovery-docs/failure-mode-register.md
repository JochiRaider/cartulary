---
doc_id: THR-S5-FAILURE-MODES
title: Testing Harness Recovery Failure Mode Register
status: active
role: failure-mode-register
---

# Testing Harness Recovery Failure Mode Register

## Document role

This register captures S4 follow-up failure modes at the observable-interface
level. Runtime-only outcomes remain source-limited.

## Failure modes

| Failure ID | Failure class | Phase | Trigger | Observable result | Side effects | Cleanup behavior | Retryable | Owner | Linked gaps | Evidence | Evidence status | Spec treatment |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| FAIL-0001 | `usage_error` | CLI parse | Missing required command args or invalid option. | stderr usage, exit `2` in source-observed scripts. | none expected. | none. | no | harness CLI | S5 interface | S2 command map; CLI sources | `observed` | preserve usage shape only after schema review. |
| FAIL-0002 | `configuration_error` | setup | Missing `TEST_SERVICES_BIN` or browser stack binary/tool env. | non-zero startup/config failure. | no child work. | partial setup cleanup if started. | no until fixed | harness config | `AUD-S4-FU-0005` | `ENV-0005`, `ENV-0022` | `observed/source_limit` | keep source-limited for exact stderr. |
| FAIL-0003 | `preflight_error` | suite preflight | Docker client/ping or managed-container preflight fails. | service failure summary, no child execution. | no child tests. | startup cleanup path. | conditional | S6/S8 | `AUD-S4-FU-0001`, `AUD-S4-FU-0006` | `SVC-0007`, `HAZ-S4-0002` | `observed/source_limit` | runtime evidence required. |
| FAIL-0004 | `service_start_error` | service startup | Postgres or MinIO testcontainer fails after attempts. | startup attempt summary and non-zero suite. | possible partially-started container. | attempted container terminate/direct cleanup/reaper. | conditional | S6 | `AUD-S4-FU-0001`, `AUD-S4-FU-0002` | `SVC-0002`, `SVC-0005` | `observed/source_limit` | retryability from source only until runtime pass. |
| FAIL-0005 | `service_start_error` | template setup | Postgres template migration/preparation fails. | startup failure class for template stage. | Postgres/MinIO may be running. | suite cleanup. | conditional | S6 | `AUD-S4-FU-0001`, `AUD-S4-FU-0002` | `SVC-0003`, `PCS-0002` | `observed/source_limit` | classify cleanup strength in S6. |
| FAIL-0006 | `fixture_error` | stale janitor | Stale browser fixture janitor fails or encounters unsafe candidate. | suite startup or cleanup failure summary. | stale DB/bucket may remain. | source shows bounded workers and generated-name guards. | unknown | S6/S8 | `AUD-S4-FU-0002` | `SVC-0011`, `HAZ-S4-0004` | `observed/source_limit` | destructive bounds require S8. |
| FAIL-0007 | `resource_conflict` | browser setup | Configured backend/frontend port in use or dynamic port race. | diagnostic stderr, startup failure. | no or partial stack. | process cleanup if any process started. | conditional | S6/S8 | `AUD-S4-FU-0006`, `AUD-S4-FU-0007` | `RES-0017`, `HAZ-S4-0005` | `observed/source_limit` | decide `ss` support profile in S8. |
| FAIL-0008 | `service_readiness_timeout` | browser setup | Backend `/readyz` or frontend origin never responds. | startup failure and logs. | process may remain until cleanup. | process group stop/port wait in source. | conditional | S6 | `AUD-S4-FU-0001`, `AUD-S4-FU-0002` | `SVC-0009`, `SVC-0010` | `observed/source_limit` | runtime timing needed. |
| FAIL-0009 | `fixture_error` | reset boundary | Reset route non-200 or response validation failure. | response/status artifacts, script non-zero. | DB/object state may be partially changed. | Playwright state clear may not happen. | unknown | S8/S6 | `AUD-S4-FU-0003` | `SVC-0014`, `RES-0023` | `observed/source_limit` | authority required before normative reset contract. |
| FAIL-0010 | `cleanup_error` | service cleanup | Detached reaper scheduling fails or cleanup returns error. | cleanup status `cleanup_failed` or direct cleanup fallback. | containers/fixtures may remain. | source-limited completion. | conditional | S6 | `AUD-S4-FU-0002` | `RES-0012`, `HAZ-S4-0003` | `observed/source_limit` | classify best-effort versus guaranteed. |
| FAIL-0011 | `cleanup_error` | browser fixture cleanup | Active DB connection leak blocks fixture reclaim. | cleanup failed summary. | fixture remains pending/stale. | stale janitor later may clean if authorized/safe. | conditional | S6/S8 | `AUD-S4-FU-0002` | `HAZ-S4-0004` | `observed/source_limit` | live active-connection behavior unobserved. |
| FAIL-0012 | `unsupported_platform` | platform/tool check | Docker, Compose, `ss`, `curl`, `setsid`, `realpath`, shell, Node/pnpm, or browser runtime missing. | mixed source-observed fail/skip/no-op behavior. | command-specific. | command-specific. | no until environment changes | S8 | `AUD-S4-FU-0006` | `ENV-0024`, `AMB-0028` | `source_limit` | support profile required. |
| FAIL-0013 | `authority_required` | command selection | Direct package scripts bypass Make guarantees. | tool-default output and weaker summaries/cleanup. | tool-defined artifacts. | mixed. | n/a | S8 | `AUD-S4-FU-0004` | `HAZ-S4-0009`, `AMB-0020` | `observed` | decide contract or developer convenience. |
| FAIL-0014 | `authority_required` | env contract | Multiple layers set/read same env without proven public precedence. | caller-visible behavior can differ by invocation path. | command-specific. | n/a | n/a | S8 | `AUD-S4-FU-0005` | `AMB-0012`, `AMB-0026`, `SL-0015` | `observed/source_limit` | public env matrix required. |
| FAIL-0015 | `authority_required` | local dev | Compose fixed ports/persistent volumes and `db-reset` object-store boundary affect local state. | local service state persists or conflicts. | DB/object data may remain. | no compose down observed. | n/a | S8/S6 | `AUD-S4-FU-0008` | `SVC-0008`, `HAZ-S4-0008` | `observed/source_limit` | local-dev contract decision required. |
| FAIL-0016 | `authority_required` | cleanup | External Go cache stale/corrupt or unremoved. | Go tooling failure or slow rebuild. | external `/tmp` cache remains. | not repo-cleaned. | conditional | S8 | `AUD-S4-FU-0009` | `RES-0025`, `AMB-0021` | `observed/source_limit` | default route outside cleanup contract pending decision. |

