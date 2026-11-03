# Frozen Graph input lifetime amendment

Status: explicitly adopted by the repository maintainer on 2026-11-02.
Owner application, implementation and validation are recorded separately in the
controlling tracker.

The [Extensions tracker](extensions-module-refactor-tracker.md) records F-051 and
separates source adoption from implementation. This document is review material,
never an executable requirements input.

## Problem and recommendation

Reporting already freezes the complete Graph result, its binding and typed label
candidates in its admission transaction, with an independent admission digest.
Nevertheless, its worker requires an unexpired live lease and compares the live
result with that frozen copy. This creates an unnecessary lifetime dependency:

- A valid queued report can outlive the five-minute lease even while its exact
  admitted input remains intact.
- Retiring a Network Flow view clears its selected binding. Current Graph
  reconstruction enumerates only active selected views; a pending report can
  still refer to the retired result, causing reconciliation to fail and keeping
  all reconstructed Graph operations unavailable.
- Current reconciliation selects `release_create` only. Composition previews
  retain the same inputs and lease dependency but receive no restored lease.

Production consumer inspection found lease acquisition, renewal and lookup only
in Network Flow's Reporting contribution, with Reporting reconciliation and
Graph cleanup supporting that use. No independent continuing consumer was found.

Use the retained admission as the execution boundary. Validate and freeze the
selected result once under the caller's transaction; thereafter render, retry,
replay and validate from the exact retained input. Retire Graph leases end to end.
Do not add historical result reconstruction, another source registration, lease
extension heuristics or a fallback reader.

## Concrete Reporting amendment

In Reporting's Graph input retention requirement, retain the existing complete
input, private codec, digest, canonical model, disclosure and artifact-membership
requirements. Replace its worker-renewal and pending-job gating paragraphs with:

> Admission of a release or composition preview naming Graph results MUST require
> the source profile's successful current claim and the source's current derived
> Graph readiness. Through the admission transaction, the source owner MUST
> validate the exact selected binding, lock the immutable result against cleanup,
> read its complete ordered result and produce typed label candidates. Reporting
> MUST validate and retain that complete contribution and its independent
> admission digest before committing the job. A missing or unavailable source
> rejects admission through existing owner errors without partial job/input
> publication.
>
> After admission, execution, retry, completion, replay and Recovery MUST validate
> and use only that retained contribution. They MUST NOT acquire, renew, inspect
> or release Graph leases, re-read a source owner's current declaration or
> projection, or invoke source-owner callbacks. They MUST recompute canonical
> diagram models and Mermaid source from the retained input and verify its exact
> binding, canonical bytes, admission digest, model membership and requested
> artifact membership. They MUST NOT substitute a newly selected result or
> reconstruct missing historical input.
>
> An admitted release or preview therefore does not depend on subsequent source
> profile availability, derived reconstruction, declaration retirement or result
> cleanup. It remains subject to Reporting's own claim, current actor/incident
> authorization, immutable input validation and ordinary Core Jobs lifecycle.
> Invalid retained input fails closed under the existing Reporting contracts;
> source unavailability alone does not park or fail an already admitted job.
> No new public job status, reason code or historical reader is introduced.

Remove lease identities and timestamps from the input-exclusion sentence because
those concepts no longer exist in the current contract; publication timestamps
remain operational and excluded from render-input identity. The current private
`cartulary.reporting_graph_input.v1` shape and digest framing remain unchanged.
Apply this rule equally to releases and composition previews and reconcile the
Reporting status/scope paragraph and any Report Composition references to it.

## Concrete Graph Projection amendment

In §8, retain immutable publication, exact result reads, ordered object reads,
traversal and bounded result cleanup. Remove the lease model, persistence and all
acquire, renew, release, lookup, expiry-deletion and lease-presence capabilities.
Replace §8.0's lease rules with:

> A consumer that needs a result beyond its admission transaction MUST retain the
> complete required immutable contribution under its own current admission
> contract. The result adapter MUST support an exact binding read through a
> borrowed transaction that locks the result envelope against cleanup. All
> ordered objects and canonical bytes MUST belong to that same admitted result.
> Result locking precedes source declaration locking. No consumer may infer
> continued availability from an earlier read or an unretained result ID.
>
> Cleanup locks at most one candidate using the existing owner, age, ordering and
> skip-locked rules, then rechecks the source owner's selected binding in that
> same transaction. It MUST NOT delete a selected result or bypass an in-flight
> admission's result lock. Unselected results remain subject to the existing
> result retention policy; private retained consumer inputs are not reachable
> Graph results and are never deleted by Graph cleanup.

In §8.1, remove expired-lease cleanup and lease rechecks. Preserve the borrowed
transaction boundary, result-before-declaration lock order, bounded result
cleanup, pure engine boundary and positive capability/import/export allowlists.
Remove lease-only configuration, counters and telemetry; retain operational
controls still used by result cleanup.

Replace §9's current reconstruction generation with:

- algorithm `graphprojection.restore_rebuild.v5`;
- result `cartulary.graph_projection_restore_rebuild_result.v5`, retaining the
  current fields except that `reconciled_lease_count` is removed entirely;
- source registry v5, retaining the current active Network Flow semantic-query
  v2 source definition;
- implementation binding v6, binding the new algorithm, registry, current catalog
  and exactly the three result, vertex and edge tables.

The three derived tables are `graph_projection_results`,
`graph_projection_result_vertices` and `graph_projection_result_edges`. Structural Recovery
leaves them empty. After successful source claim, reconstruction still validates
current frozen identities, uses the original durable operation/generation,
clears these tables, enumerates active selected declarations, rebuilds and
publishes exact results, reconciles Network Flow nonterminal materialization
jobs and verifies exact postconditions in one caller-controlled transaction.
Reporting jobs and inputs are outside this derived reconstruction transaction;
their ordinary Core Jobs recovery and Reporting validation remain mandatory.

Pending or failed reconstruction gates Network Flow's dependent operations and
new Reporting admissions requiring live Graph contributions. It does not gate
already admitted Reporting work. Preserve one attempt per process activation,
idempotent restart, separate structural completion, safe determinate failure and
Core fatal shutdown for indeterminate commit. Do not manufacture successful
results for unavailable Graph operations.

Retire prior rebuild/registry/binding generations from current selection and
support; retire the lease capabilities entirely. Pure projection input/result v2 and their
canonical identities remain current and unchanged.

## Coordinated Core and Network Flow amendments

Core 01 REQ-01-625A and Graph §9 must agree on the new generation, three derived
tables, Network Flow-only job reconciliation, independent retained Reporting
execution and the gates above. Core 01's current frozen catalog removes the
lease table; repeated catalog facts are regenerated from authored machine inputs.
Core 00 records the newly adopted current companion behavior without workflow
IDs. Core 04's structural journal v7 and fatal-integrity rules remain unchanged.

Update Network Flow NF-REQ-201/203 and the current shared-Graph consumer section:
source admission validates the current selected tuple, locks/reads the complete
result and derives labels within the caller transaction. It has no lease API.
Retirement still changes only the declaration and does not delete private
Reporting inputs. Result cleanup continues to protect current selections and
in-flight admissions. Network Flow's own dependency gate and materialization
job reconciliation remain required. No profile-state migration is introduced.

Report Composition's admission, local-validation and fixture tables must remove
lease side effects while preserving their no-partial-publication guarantees.
Graph's security/acceptance sections must replace lease checks with exact
admission-lock and retained-input checks.

Remove current normative claims that Reporting needs lease renewal or Graph
reconstruction after its input is frozen. Preserve historical decision evidence
in handoffs, not as selectable contracts or compatibility promises.

## Compatibility and migration

This is one coherent unpublished current-only cutover. Snapshot v3, control v1,
journal v7, marker v5 and verification v5 retain their shapes. The frozen catalog,
Graph definitions, generation membership and dependent digests change together;
old backup catalog/codec pairs and retired Graph identities reject before
restore target mutation. Old generation-bound completion cannot authorize current
startup, and its control tuple MUST NOT be rewritten to claim compatibility.
No reader, converter, alias, implicit claim or automatic environment reset is
added. Existing private Reporting input bytes are neither rewritten nor deleted.

Preserve all existing migration files. A new Core-owned migration may remove the
retired lease table only after verifying it is empty in the same transaction
under the required table lock. A nonempty table MUST fail migration before any
mutation; the migration MUST NOT delete or expire lease rows to make itself pass.
Existing incompatible deployments require an operator-controlled preparation or
fresh deployment outside this effort. No deployment or data cleanup is authorized
by adopting this source amendment.

## Validation and adoption boundary

After adoption, implement the complete projection/runtime retirement and validate:

1. Atomic selected-result freeze versus concurrent refresh, retirement and
   cleanup, with the preserved lock order and no partial job/input publication.
2. Current release and preview execution from frozen inputs after lease-duration
   time has elapsed, source unclaim, view retirement, derived-table removal and
   a current structural restore. Authorization loss and corrupt retained input
   still fail under their own contracts.
3. Current Graph reconstruction with pending retained release and preview jobs;
   it neither requires their old derived results nor mutates their immutable
   inputs. Ordinary Core Jobs recovery preserves required attempt semantics.
4. New graph-dependent admission remains unavailable until successful claim and
   exact reconstruction; existing admitted non-graph work continues. Unknown
   Graph commit still closes admission and triggers fatal shutdown.
5. Current capture/restore schema and catalog parity, retired identity rejection,
   empty-table migration success and nonempty-table rejection without mutation,
   generated drift, API/boundary checks and absence of production lease paths.

Replace still-required lease-era assertions with current admission-lock,
retained-input fidelity and cleanup evidence before deleting obsolete positive
lease fixtures. Keep small static retired-schema fixtures only for rejection.
The maintainer adopted this concrete amendment on 2026-11-02 and authorized
future required amendments without another approval request. Concrete source
review and adoption records remain required; tests and lint are validation
evidence, not adoption authority.
