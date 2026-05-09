---
doc_id: THR-S6-TIMEOUT-RETRY
title: Testing Harness Recovery Timeout Retry Register
status: active
role: timeout-retry-register
---

# Testing Harness Recovery Timeout Retry Register

## Document role

This register records source-observed waits, retries, timeouts, and cleanup
windows relevant to S4 follow-ups. It is static evidence unless a row says
otherwise.

## Timeout and retry register

| Rule ID | Surface | Timeout/retry type | Unit/default | Eligible failure classes | Algorithm | Exhaustion behavior | Linked rows | Evidence | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|---|---|
| TMR-0001 | `tools/testservices` suite preflight | timeout | `3s` | Docker/preflight errors | bounded context around preflight | startup failure before child | `SVC-0001`, `SVC-0007` | `tools/testservices/main.go` | `observed/source_limit` | Live Docker not exercised. |
| TMR-0002 | Suite Postgres startup wrapper | timeout | wrapper `2m`; attempt `35s` | Postgres start/readiness | service-specific attempt timeout | startup failure summary | `SVC-0002` | S4 service map; testservices source | `observed/source_limit` | Exact live timing not proven. |
| TMR-0003 | Postgres testcontainer readiness | wait/retry | port mapping `15s`; client readiness `15s`; poll `250ms`; attempt `5s`; max attempts `3`; backoff `500ms` | transient Docker/readiness failures except logical auth/db errors | retry via `testcontainersx.StartWithRetry` | structured start failure | `SVC-0002` | `pgtest.go`; `testcontainersx.go` | `observed/source_limit` | Static retry taxonomy only. |
| TMR-0004 | Suite MinIO startup wrapper | timeout | wrapper `5m`; attempt `2m` | MinIO start/readiness | service-specific attempt timeout | startup failure summary | `SVC-0005` | S4 service map; testservices source | `observed/source_limit` | Live MinIO not exercised. |
| TMR-0005 | MinIO testcontainer readiness | wait/retry | port mapping `30s`; client readiness `60s`; poll `500ms`; attempt `5s`; default attempts `2` | readiness deadline and configured retryable errors | authenticated `ListBuckets` loop | structured start failure | `SVC-0005` | `s3test.go`; `testcontainersx.go` | `observed/source_limit` | Static only. |
| TMR-0006 | Postgres template preparation | timeout | `2m` | migration/template setup | context deadline | startup failure stage `postgres-template` | `SVC-0003` | `tools/testservices/main.go` | `observed/source_limit` | Template freshness bound to suite startup. |
| TMR-0007 | Stale browser fixture janitor | timeout | `10s` source-observed in S4 | stale cleanup/janitor failures | bounded janitor context and worker count | startup or cleanup failure | `SVC-0001`, `SVC-0011`, `ENV-0011` | `tools/testservices/main.go` | `observed/source_limit` | Destructive bounds require S8. |
| TMR-0008 | Browser backend/frontend readiness | wait | 180 attempts at 1s; Playwright webServer `180000ms` | readiness timeout/early process exit | `curl -fsS` loop with process checks | startup failure and logs | `SVC-0009`, `SVC-0010` | `scripts/start-web-e2e.sh`; Playwright shared config | `observed/source_limit` | Browser stack not run. |
| TMR-0009 | Browser process group stop | cleanup wait | TERM then 50 x 0.2s; KILL then wait again | process cleanup | process-group poll/kill | cleanup status/non-zero possible | `RES-0020`, `CLN-0015` | `scripts/lib/process-lifecycle.sh` | `observed/source_limit` | Signal behavior unobserved. |
| TMR-0010 | Browser port release wait | cleanup wait | 50 x 0.2s | port release conflicts | `ss` check if available | release diagnostic or source-limited no-op when `ss` absent | `RES-0017`, `HAZ-S4-0005` | `scripts/start-web-e2e.sh` | `observed/source_limit` | `ss` absence routes to S8 platform support. |
| TMR-0011 | Playwright global setup lock | lock wait | `120s` | stale lock/contention | `mkdir` lock acquisition loop | setup throws on timeout | `SVC-0013`, `RES-0022` | `global-setup.ts`; `harnessState.ts` | `observed/source_limit` | Abrupt-exit stale lock behavior unobserved. |
| TMR-0012 | Playwright test timeout | test timeout | `60s` | test or fixture timeout | Playwright runner default/config | Playwright failure artifacts if retained | `SVC-0013`, `ART-0018` | Playwright shared config | `observed/source_limit` | Failure artifacts not generated. |
| TMR-0013 | Compose Postgres wait | wait | `180s`; healthcheck interval/timeout `5s`, retries `10` | local-dev readiness | `pg_isready` loop and compose healthcheck | fail with compose logs | `SVC-0008` | `scripts/dev-services.sh`; `docker-compose.dev.yml` | `observed/source_limit` | Compose not run. |
| TMR-0014 | Compose MinIO wait | wait | `120s`; healthcheck interval/timeout `5s`, retries `20` | local-dev readiness | `mc alias set` and `mc ready` | fail with compose logs | `SVC-0008` | `scripts/dev-services.sh`; `docker-compose.dev.yml` | `observed/source_limit` | Compose not run. |
| TMR-0015 | Dev stack readiness | wait | default `180s` | local backend/frontend readiness | `curl` loop with process checks | local dev startup failure | `SVC-0015` | `scripts/dev-stack.sh` | `observed/source_limit` | Local-dev contract status is S8. |
| TMR-0016 | Reset route call | request context/curl | no route-specific timeout observed | reset route failure/partial mutation | HTTP POST then JSON validation | reset script non-zero | `SVC-0014`, `RES-0026` | reset script and route source | `observed/source_limit` | `TODO: reset_route_timeout_unknown` remains. |

## Source-limit summary

Timeouts listed here are source-observed declarations only. They do not prove
host-specific timing, cleanup completion, retry exhaustion behavior, signal
handling, or parent-death behavior.

