# Graph Projection v2 Rollout and Recovery Guide

This operator guide supports the adopted Graph Projection 2.2.0, Network Flow
5.0.0, Reporting, Extensions, and Recovery owners. It does not define product
behavior. Graph protocol and persisted result shape remain v2; Network Flow is
contract major 5 and extension state 4; the only supported Graph restore
generation is v4.

## Preconditions

- Use one release containing the matching server, operator, worker, browser,
  generated contract, and migration artifacts. Do not mix major-4 and major-5
  Network Flow binaries or clients.
- Quiesce Network Flow materialization, Reporting rendering, and cleanup before
  backup or state advancement.
- Capture and verify an exact database and object-store backup. Retain its
  backup-set ID, consistency point, manifest and source-registry digests,
  Recovery catalog and implementation-binding digests, binary commit, schema
  head, and object-store generation.
- Verify migrations `00032`, `00033`, and `00034` and their recorded hashes are
  unchanged. This rollout adds no DDL or data-rewrite migration.
- Test the backup in an isolated replacement target with the current Recovery
  v4 Graph registry and binding before changing the production target.

## Rollout order

1. Stop the prior binary and all Graph-mutating workers.
2. Install the complete Network Flow major-5 release.
3. Run extension admission. Fresh claims initialize at state 4. A state-3
   profile advances only when its historical ledger is either empty or the
   exact verified 1→2 plus 2→3 lineage, and every saved declaration is
   semantic-query v2.
4. Treat state 1 or 2, a malformed historical ledger, or any semantic-query v1
   declaration as a closed stop. Do not delete, translate, or rewrite stored
   bytes to force admission.
5. Verify configuration, current generated contracts, Graph root boundaries,
   Recovery v4 selection, and application readiness before resuming workers.
6. Exercise default and temporal saved-graph creation, refresh, exact-result
   read, contributor query, Reporting lease lifecycle, and bounded cleanup.
7. Produce a fresh state-4 backup and prove isolated Recovery v4 rebuild,
   Network Flow and Reporting job reconciliation, lease reconciliation, exact
   Graph identity, postcondition verification, and readiness.

## Compatibility boundary

This is a hard current-only cut. There is no inventory gate, feature flag,
translator, dual reader or writer, compatibility view, fallback dispatcher, or
old Graph backup path. Saved semantic-query v1 declarations remain stored but
make the profile unavailable pending separately adopted remediation. Graph
Recovery v2 and v3 artifacts are unsupported and must fail selection before a
Graph mutation transaction opens.

## Rollback boundary

Once state 4 commits, a major-4 binary is not a downgrade candidate for that
target. Preserve the failed target for diagnosis. If rollback is required,
restore the exact pre-rollout database and object-store backup into a fresh
replacement target, start the exact recorded prior release there, verify its
own Recovery and serving readiness, and switch traffic. Never downgrade state
in place, edit migration history, recreate removed tables, reverse-map Graph
identities, or point an older binary at the state-4 target.

## Required retained evidence

- quiescence and target-generation identity;
- pre-rollout and state-4 backup identities and integrity results;
- exact state and migration-ledger preflight outcome;
- major-5 server/browser contract and generated-artifact drift results;
- Recovery v4 registry, binding, isolated restore, rollback-injection,
  reconciliation-count, postcondition, and readiness results;
- default and temporal Graph IDs and digests before and after rollout;
- Reporting lease and Network Flow cleanup/race evidence; and
- final confirmation that active code, generated artifacts, browser assets,
  fixtures, and guides expose only current contracts.
