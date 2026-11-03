# Graph Projection v2 Rollout and Recovery Guide

This guide supports current adopted Graph Projection, Network Flow, Reporting,
Extensions and Recovery owners. It defines no product behavior or deployment
authorization. Graph input/result protocol and canonical identities remain v2.
Network Flow is public major 8, durable state/minimum 5 and semantic-query v2.
Graph reconstruction uses algorithm/result v5, source registry v5 and
implementation binding v7.

Database generation admission precedes extension-profile admission. Use the
current authored migrations and frozen Core catalog; this guide is not another
migration-version or table-count registry. Structural Recovery completion is
separate from later Graph availability. There is no reader, converter or
compatibility fallback for incompatible backup generations or profile states.

## Preconditions

- Qualify one coherent server, operator, worker, browser, authored-contract,
  generated-artifact and migration set. Generated identities must match the
  actual typed consumers. No mixed public major or format generation is valid.
- Confirm current configuration, secret/key availability, authorization,
  process fencing and required service operation through the owner-selected
  Make portfolios. Module readiness alone is not release qualification.
- Preserve existing data and control evidence. Unsupported database lineage,
  profile state or backup identity is a closed stop, not permission to delete,
  rewrite or initialize over existing data.
- Retain the exact backup-set and consistency-point identities, integrity and
  binding proofs, current Recovery catalog/generation, source-registry and
  implementation-binding digests, binary identity and object-store evidence.
- Qualify capture and restore against a separately admitted stopped-empty
  target. Exercise failure, interruption, quarantine, restart and later claim;
  passing an in-memory projection test is insufficient.

## Qualification and serving order

1. Verify database generation and process ownership before serving or starting
   workers. The retired Graph lease table may be removed only when empty under
   the Core migration's transaction lock. Nonempty state stops admission.
2. Authenticate and admit a current backup's identities and exact catalog
   before mutation. Restore source declarations and binding proofs through the
   current fenced structural path. Leave the three derived Graph tables empty;
   inactive source semantics must not run.
3. Establish generation-bound structural completion through Core's durable
   target control and completion evidence. A failed or interrupted target stays
   non-serving until the current Recovery lifecycle permits otherwise.
4. Claim Network Flow explicitly. Current state 5 undergoes exact final
   validation; older state or semantic-query v1 declarations reject without
   repair. Unclaimed Network Flow state remains preserved without active routes,
   operational adapters, workers or source semantics.
5. After successful claim, reconstruct current Graph results from admitted
   authoritative state in one caller-controlled transaction. Publish immutable
   results, reconcile Network Flow nonterminal materialization jobs and verify
   exact result/row/digest and readiness postconditions. New Graph-dependent
   operations remain gated until that transaction succeeds.
6. Verify actual process and browser behavior: default and temporal Graph
   creation/refresh, selected-result reads, contributor queries, authorization
   loss, failed claims and reconstruction failure. A determinate Graph failure
   does not invalidate completed structural Recovery or unrelated Base serving.
7. Verify Reporting freezes complete contributions during admission and later
   executes from that immutable input. Source retirement, Graph cleanup or a
   later unavailable source must not create a Graph lease dependency or change
   the admitted report input.
8. Capture a fresh current backup and repeat isolated restore plus subsequent
   claim/reconstruction. Retain the terminal outcomes and exact source/evidence
   identities before considering release or deployment separately.

## Compatibility and rollback

Current-only admission has no inventory bypass, feature flag, dual reader,
translator, old Graph dispatch path or implicit profile claim. Existing
incompatible state stays intact and unavailable. Any operator remediation is a
separate authorized action; this guide grants no automatic reset or deletion.

Before publication, rollback means reverting the coherent unpublished change.
Never point an older binary at newly written backups, current completion
evidence or state whose contracts it cannot interpret. Preserve failed targets
for diagnosis. A future deployment rollback needs its own qualified data and
release procedure; editing migration history, replacing stored digests or
reverse-mapping Graph identities is not such a procedure.

## Required retained evidence

- exact source, current generated identities and database/catalog admission;
- immutable backup, integrity and per-binding proof identities;
- stopped-empty admission, fenced mutation, structural completion and
  interruption/quarantine/restart outcomes;
- inactive preservation, current state-5 claim, rejected older/invalid state,
  and post-claim Graph reconstruction with atomic reconciliation;
- canonical content and row fidelity at publication and every content read;
- default and temporal Graph identity, cleanup/locking and authorization cases;
- actual browser and process witnesses, including availability loss;
- Reporting frozen-input survival through source cleanup and restart; and
- distinct conclusions for module readiness, release qualification and actual
  deployment, with unsupported or unexecuted obligations stated explicitly.
