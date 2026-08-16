# Network Flow GP3 Rollout and Rollback Guide

This guide supports the adopted Network Flow, Graph Projection, Jobs,
Extensions, Recovery, Reporting, and application-server owners. It does not
define product behavior or publish a hardware-independent performance claim.

## Deployment profile selection

The GP3 reference deployment selects the conservative configuration-major-2
defaults. A deployment may select a raised value only after retaining a
successful environment-local workload run that binds the source commit,
source digest, toolchain digest, system digest, runtime and hardware profile,
complete effective limits, wall time, allocations, peak resident memory,
database work, output bytes, and outcome.

The release-tier Network Flow capacity row retains default skew, default dense,
raised dense, default temporal, maximum contributing-row, and maximum bucket
observations. Its raised and semantic-maximum observations are diagnostic: a
passing observation does not select that profile. Cleanup, cancellation,
timeout, mixed-queue, crash recovery, browser response, restore, and shutdown
remain separate routed rows so their failures are attributable. If any
deployment-local raised-profile gate fails, remove that override and repeat the
drill at the highest fully proven lower profile; never change stable Graph
semantics to make a workload pass.

## Pre-rollout evidence

Before applying GP3, quiesce mutating workers and capture an exact verified
pre-GP3 backup. Retain these facts together:

- backup set ID and consistency point;
- old binary commit and deployable artifact digest;
- Extension state version and Network Flow contract/configuration majors;
- database schema head and migration-history digest;
- Recovery catalog, codec registry, Graph restore registry, and implementation
  binding digests;
- database and object-store restore anchors and target-generation identity;
- inventory of installed persisted graph semantic generations; and
- inventory and retention expiry of backup sets that still require the
  historical `graphprojection.restore_rebuild.v2` dispatcher.

Verify the backup in a fresh isolated replacement target before continuing.
The target must pass exact authoritative-row, selected Graph result, vertex,
edge, job, lease, and readiness reconciliation. Do not use a backup whose
integrity, compatibility, or isolated-restore result is incomplete.

## Rollout order

1. Apply additive migration `00034_graph_projection_cleanup_indexes.sql`.
   Migrations `00032` and `00033` remain immutable.
2. Deploy the GP3 binary with a complete claimed Network Flow
   configuration-major-2 value. Configuration admission must finish before any
   listener or worker starts.
3. Confirm the generated worker runtime contract admits the existing workers
   at eight active attempts per process and the Network Flow graph worker at
   one. Confirm the cleanup dispatcher starts only after serving readiness.
4. Advance Network Flow extension state `2 -> 3`. Verify that installed v1
   declarations remain byte-preserved and readable while the major-4 API
   rejects new v1 writes.
5. Deploy the major-4 browser and verify discovery, default and temporal graph
   queries, saved lifecycle, contributor pivots, bucket navigation, bounded
   rendering, and fail-closed external Reporting release.
6. Enable new semantic-query-v2 writes. Verify the first v2 saved result and a
   mixed v1/v2 rebuild before opening normal traffic.
7. Resume workers and confirm queue wait, materialization phases and volume,
   cleanup health, eligible backlog, and oldest eligible publication-age
   signals contain only closed safe attributes.

## Rollback boundary

Before extension state advances and before any v2 declaration is written, the
new binary may be removed only when the prior binary has been verified against
the still-supported schema and configuration. Migration `00034` is additive
and may remain when that compatibility has been demonstrated.

Once extension state 3 or any v2 declaration exists, do not point the prior
binary at the upgraded target. Stop admission, quiesce workers, preserve the
failed target for diagnosis, restore the exact pre-GP3 database and object-store
backup into a fresh replacement target, start the exact recorded prior binary,
verify Recovery and serving readiness, and then switch traffic. Do not
translate v2 declarations, reverse-map v2 digests or identities, downgrade
state in place, edit migration history, or reconstruct Graph rows manually.

## Required drills and retained handoff

The rollout is certified only when the retained evidence identifies successful
results for:

- default, raised diagnostic, semantic-maximum diagnostic, skew, dense, and
  temporal capacity workloads;
- cleanup batch boundaries, destructive races, paced continuation, and
  shutdown;
- cancellation, timeout, crash recovery, mixed-job queue pressure, and graph
  worker isolation;
- a fresh state-3 backup containing persisted v1 and v2 declarations, followed
  by exact isolated restore and mixed-generation Graph rebuild;
- major-4 browser response and bounded rendering; and
- the replacement-target rollback procedure using the exact pre-GP3 backup.

Record every run root and artifact path in the controlling Graph Projection
refactor tracker. Retain persisted-v1 execution until the installed declaration
inventory is zero. Retain the exact Recovery v2 dispatcher until the supported
backup inventory is also zero. Removing either path requires a later adopted
owner decision; the GP3 rollout does not authorize it.
