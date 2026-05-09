---
doc_id: THR-S6-CONCURRENCY-MODEL
title: Testing Harness Recovery Concurrency Model Notes
status: active
role: concurrency-model-notes
---

# Testing Harness Recovery Concurrency Model Notes

## Document role

These notes summarize the concurrency model observed while routing S4 follow-up
gaps. They are not a capacity guarantee.

## Model summary

| Model area | Source-observed behavior | Concrete guarantee status | Linked rows | Evidence status | Notes |
|---|---|---|---|---|---|
| Make sequencing | Generated recipes and sequence manifests define serial/parallel child targets. | source-observed only for declarations; broad runtime not executed. | `EP-0002`, `EP-0008` | `observed/source_limit` | Runtime drain behavior belongs to S5/S6 failure evidence. |
| Check scheduler | Resource registry declares host, service, process, browser-stage, and scratch Postgres lanes. | scheduler-accounting guarantee only. | `RES-0001` through `RES-0004`, `RES-0008`, `RES-0009` | `observed` | No OS-level CPU/IO reservation is implied. |
| Service-backed scheduler | Schedules Go shards, Make targets, and browser stages under `postgres`, `minio`, Go, reset, process, and browser lanes. | scheduler-accounting guarantee only. | `RES-0005` through `RES-0010` | `observed` | Concrete service capacity remains `RTR-0001`. |
| Suite service stack | `suite_service_stack=1` serializes suite stack ownership in check profile; child tests attach to suite services. | source-observed wrapper behavior; live readiness not executed. | `SVC-0001` through `SVC-0007` | `observed/source_limit` | Docker/testcontainers runtime remains `SL-0012`. |
| DB fixture isolation | DB names include suite/process/counter; package/group maps use mutexes; transactions rollback. | source-observed collision mitigation. | `RES-0014`, `RES-0015` | `observed/source_limit` | Active-connection cleanup is not guaranteed. |
| Bucket isolation | Bucket names include suite/process/counter; package bucket reuse is mutex/key scoped. | source-observed collision mitigation. | `RES-0016` | `observed/source_limit` | Prefix cleanup timeout is caller-context only. |
| Browser stacks | Scheduler uses browser stack/stage/process lanes; shell allocates ports and starts process groups. | source-observed; live port/process cleanup not executed. | `RES-0008`, `RES-0017`, `RES-0020` | `observed/source_limit` | `ss` support profile remains open. |
| Playwright workers | Worker count and index offset come from Make/scheduler/env/Playwright config. | source-observed; direct package defaults differ. | `ENV-0016`, `RES-0022` | `observed/source_limit` | Package-script differences route to S8. |
| Local dev services | Compose services use fixed ports and persistent volumes. | local-dev behavior source-observed; contract status open. | `SVC-0008`, `SVC-0015`, `RES-0018`, `RES-0019` | `observed/source_limit` | No compose down target observed. |

## Concurrency rules for later NLSpec drafting

- Treat scheduler lanes as harness scheduling limits, not concrete host or
  service capacity proof.
- Treat unique names, locks, and counters as collision mitigations, not proof
  that live resource creation cannot fail.
- Treat shell traps and process groups as cleanup paths, not cleanup guarantees
  under timeout, interrupt, or parent death.
- Treat direct package scripts as outside Make concurrency policy until S8
  decides otherwise.

