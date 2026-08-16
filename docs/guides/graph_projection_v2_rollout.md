# Graph Projection v2 Rollout and Recovery Guide

This guide supports the adopted Graph Projection, Network Flow, Reporting, and
Recovery owners. It does not define product behavior.

## Rollout order

1. Deploy code that understands additive migration `00032` while retaining the
   five v1 tables.
2. Verify Network Flow major 3 discovery, state migration `1 -> 2`, Graph v2
   materialization, Reporting exact-result leases, and Recovery v2 in a
   disposable environment.
3. Capture and durably verify an exact pre-cutover backup. Record its backup set
   ID, manifest digest, binary commit, schema head, object-store generation, and
   Recovery artifact digests.
4. Stop saved-graph materialization and Reporting render workers.
5. Run the `00033` preflight. Any legacy Graph row or persisted Reporting v1
   projection reference blocks the deployment; do not translate or delete it
   manually.
6. Apply `00033`, restart the v2 binary, and verify readiness before workers
   resume.
7. Produce a fresh v2 backup, pass isolated restore, and verify the resulting
   binary accepts only the current Graph v2 registry and binding inventory.

## Rollback boundary

Before `00033`, an old binary may be restarted only against a schema it supports
and only after v2 workers are stopped. Migration `00032` may remain when the old
binary demonstrably ignores the additive tables.

After `00033`, old binaries are not in-place rollback candidates. Restore the
exact pre-cutover database and object-store backup into a replacement target,
start the exact recorded old binary there, verify Recovery/readiness, and then
switch traffic. Never recreate legacy tables by hand, reverse-map v2 result
identity, or point an old binary at the destructively migrated target.

## Required evidence

- zero-state preflight result for all five legacy tables and Reporting v1 refs;
- stopped-worker proof and target-generation identity;
- pre-cutover and fresh-v2 backup identities and integrity results;
- isolated restore run roots and exact Graph result/object identity comparison;
- current supported backup/journal inventory;
- final confirmation that no supported artifact or runtime path retains the
  historical bridge.
