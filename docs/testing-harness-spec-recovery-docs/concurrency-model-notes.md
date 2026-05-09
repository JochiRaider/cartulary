---
doc_id: THR-S6-CONCURRENCY-MODEL
title: Testing Harness Recovery Concurrency Model Notes
status: active
role: concurrency-model-notes
---

# Testing Harness Recovery Concurrency Model Notes

## Document role

These S6 notes summarize the source-observed concurrency model across Make,
schedulers, service stacks, fixtures, browser execution, Playwright workers,
package scripts, and local-dev flows. They are recovery evidence, not capacity
guarantees.

## Concurrency model summary

| Concurrency ID | Model area | Source-observed behavior | Concrete guarantee status | Linked rows | Evidence status | Notes |
|---|---|---|---|---|---|---|
| CONC-0001 | Generated Make sequencing | `Makefile` includes the generated task surface; sequence recipes declare serial and parallel child target groups. | Source-observed declarations only; broad runtime drain behavior was not executed in S6. | `SEQ-0001` through `SEQ-0007`, `RTR-0002` | `observed/source_limit` | Make parallelism is caller-controlled except where recipes delegate to sequence or scheduler scripts. |
| CONC-0002 | Check scheduler resource model | `check` consumes `tools/check_schedule_manifest.json` and `tools/scheduler_resource_registry.json` to schedule work by `needs`, retained claims, and resource claims. | Scheduler-accounting guarantee only. | `RES-0001` through `RES-0004`, `RES-0008`, `RES-0009`, `RTR-0001` | `observed` | No OS-level CPU/IO, Docker, DB, MinIO, or browser capacity reservation is implied. |
| CONC-0003 | Service-backed scheduler resource model | Service-backed schedules expand Go shards, Make targets, and browser stages under `postgres`, `minio`, `go_cpu`, `go_io`, `postgres_reset`, `process`, `browser_stack`, and dynamic browser-stage lanes. | Scheduler-accounting guarantee only. | `RES-0005` through `RES-0010`, `RTR-0001` | `observed` | Concrete service capacity and fixture isolation are separate hazards. |
| CONC-0004 | Go shard and shared-report model | Go target runner captures shard reports, uses shared report directories and lock directories, and finalizes shard accounting. | Source-observed lock/capture behavior; lock timeout and stale-lock behavior are not runtime-observed here. | `HAZ-S3-0006`, `RTR-0018`, `TMR-0018`, `TMR-0019` | `observed/source_limit` | Locks mitigate duplicate capture but do not prove all process interruption paths release cleanly. |
| CONC-0005 | Suite service stack ownership | `tools/testservices run` starts one owned suite unless already active; children attach via suite env; check profile serializes `suite_service_stack=1`. | Source-observed wrapper and lane behavior; live readiness and cleanup were not executed. | `SVC-0001` through `SVC-0007`, `RES-0012`, `RTR-0004`, `RTR-0007` | `observed/source_limit` | Nested active suites pass through rather than starting a second owned suite. |
| CONC-0006 | Postgres fixture isolation | DB names include suite/process/counter/suffix; package and group fixtures use keyed maps and mutexes; transactions roll back. | Source-observed collision mitigation, not proof that live DB create/drop cannot fail. | `RES-0014`, `RES-0015`, `RTR-0005`, `RTR-0010` | `observed/source_limit` | Active connections and ordinary DB operation timeouts remain source-limited. |
| CONC-0007 | MinIO bucket isolation | Bucket names include suite/process/counter/suffix; package bucket reuse is keyed and mutex-protected; prefix cleanup occurs before reuse. | Source-observed collision mitigation, not proof that object cleanup cannot fail. | `RES-0016`, `RTR-0006`, `RTR-0010` | `observed/source_limit` | Prefix cleanup timeout is caller-context only. |
| CONC-0008 | Browser stack concurrency | Scheduler lanes bound browser stack/stage/process concurrency; `start-web-e2e.sh` allocates ports, runtime root, backend/frontend process groups, DB/bucket fixture, and stack metadata. | Source-observed; live port/process cleanup and readiness were not executed. | `RES-0008`, `RES-0017`, `RES-0020`, `RES-0021`, `RTR-0011`, `RTR-0012` | `observed/source_limit` | Dynamic port allocation has a bind race before backend/Vite start. |
| CONC-0009 | Browser batch and reset sequencing | Browser batch manifests declare stage groups and shards; resettable/stateful/measurement/visual stages use worker `1`; reset boundaries call the app test reset route. | Source-observed ordering; reset route behavior was not invoked. | `tools/browser_e2e_batch_manifest.json`, `SVC-0014`, `RTR-0014` | `observed/source_limit` | Reset route authority is product/harness boundary-sensitive. |
| CONC-0010 | Playwright workers and shared state | Worker count comes from Make/env/Playwright config; global setup writes worker admin manifest; shared setup uses a lock directory. | Source-observed lock/manifest behavior; abrupt-exit stale state remains source-limited. | `ENV-0016`, `RES-0022`, `RTR-0013`, `TMR-0011` | `observed/source_limit` | Direct package scripts can use different worker defaults than Make. |
| CONC-0011 | Direct package scripts | Root/app package scripts invoke pnpm/Vite/Vitest/Biome/Playwright directly unless a child script re-enters a harness wrapper. | Tool-native concurrency and artifacts only unless owner adopts them. | `EP-0016`, `HAZ-S4-0009`, `RTR-0015`, `AUTH-0003` | `observed` | Treat as outside Make scheduler/result-root policy until S8 decides otherwise. |
| CONC-0012 | Local dev and Compose services | Local dev uses fixed ports, persistent Compose volumes, default local credentials, and no observed compose-down target. | Local-dev behavior source-observed; verification-contract status open. | `SVC-0008`, `SVC-0012`, `SVC-0015`, `RES-0018`, `RES-0019`, `RTR-0019` | `observed/source_limit` | `db-reset` explicitly does not reset MinIO/object storage. |
| CONC-0013 | Retained artifacts and generated planning inputs | Retained result roots, generated schedules, ledgers, and baselines are reused by investigation, drift, scheduling, and baseline-refresh tools. | Source-observed reuse; freshness/provenance must be selected explicitly for durable claims. | `RTR-0002`, `RTR-0003`, `RTR-0016` | `observed/source_limit` | Derived artifacts may be execution inputs without becoming behavior owners. |
| CONC-0014 | External caches and scratch dirs | Go caches default outside repo under `/tmp`; shell harness scratch roots are outside repo by default and guarded. | Tool-managed/cache behavior or caller-trap behavior only. | `RES-0024`, `RES-0025`, `RTR-0017`, `RTR-0020` | `observed/source_limit` | Cleanup scope remains authority-bound for external Go caches. |

## Later NLSpec drafting rules

- Treat scheduler lanes as harness scheduling constraints, not concrete capacity
  guarantees.
- Treat generated Make and schedule files as downstream execution inputs, not
  owner text.
- Treat unique names, counters, hashes, locks, and mutexes as collision
  mitigations, not proof that live resource creation or cleanup cannot fail.
- Treat shell traps, process groups, Go `t.Cleanup`, Playwright teardown, and
  detached reapers as cleanup paths, not cleanup guarantees under timeout,
  interrupt, or parent death.
- Treat package scripts as outside Make-owned output, cleanup, and resource
  policy unless S8 adopts them as first-class harness contracts.
- Treat local-dev Compose state as local-dev persistence unless S8 adopts a
  verification contract for it.
- Treat runtime-sensitive timing and cleanup claims as `source_limit` unless a
  later authorized run records `runtime_observed` evidence.
