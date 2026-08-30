# Graph Projection Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-076.

This decision owns only the repository-internal Go package topology, import
direction, constructor posture, restore placement, narrow port allocation,
application composition, and compatibility removal for Graph Projection. It
does not create a domain bounded context or redefine Graph Projection protocol
bytes, result identity, Network Flow saved-graph behavior, Reporting
consumption, Recovery orchestration, storage schemas, authorization, jobs, or
public routes.

The adopted Graph Projection NLSpec owns deterministic projection semantics and
the Graph restore participant. Network Flow owns saved graph declarations,
graph-view identity, materialization, and public resources. Recovery owns
selection, binding admission, target lifecycle, and orchestration. Reporting
owns report jobs and exact leased consumption. Core 04 owns security,
deployment-local Recovery authority, and conformance. `docs/domain.md` owns
vocabulary and owner navigation.

If this decision conflicts with an adopted behavioral owner, that owner
governs and this decision must be repaired. Runtime code, tests, generators,
conformance, and release evidence do not read this file. Typed contracts and
authored verification policies project the adopted facts without superseding
their owners.

## Decision

Graph Projection is one pure deterministic engine plus cohesive operational
adapters. Its root is `internal/modules/graphprojection`. The root owns the
closed v2 invocation and completed-result model and exposes exactly one engine
entrypoint:

```go
func ProjectV2(
    ctx context.Context,
    invocation InvocationContextV2,
    rawInput []byte,
) (ProjectionResultV2, error)
```

The root imports only the Go standard library and its own private
implementation. Parsing, normalization, semantic validation, derivation,
canonicalization, limit enforcement, cancellation cadence, and result-identity
helpers remain private. There is no engine instance, engine constructor,
mutable option set, graph-view identity helper, restore contract, database
facade, HTTP or authorization surface, job facade, or cross-owner alias in the
root.

Graph-view identity is a Network Flow concern because it names a saved
declaration, not a projected result. Network Flow keeps its derivation helper
private and passes the trusted value through `InvocationContextV2`. Trusted
invocation values never enter semantic configuration or configuration digests.

## Restore ownership

`internal/modules/graphprojection/restore` owns the current Graph restore
contracts, v4 source-registry construction, rebuild orchestration, and the pure
Graph recovery-state contribution. It may depend on the Graph root through the
single projection entrypoint. The root does not import it.

`internal/modules/graphprojection/postgresrestore` owns PostgreSQL realization
of the restore transaction and writer ports. Its production constructor is
fallible and requires both a database and a reconciler. A writer that skips
reconciliation is unconstructable. Clear, rebuild, immutable publication,
Network Flow nonterminal-job reconciliation, Reporting job and lease
reconciliation, postcondition verification, and ready publication occur in one
transaction.

Recovery may expose or alias only the minimal cross-owner port required to
invoke the current Graph participant. Recovery owns selection and rejects a
Graph v2 or v3 binding before opening the Graph mutation transaction. Network
Flow supplies only current semantic-query v2 restore candidates. Application
and revision assembly bind these owners; they do not move their behavior into a
composition package.

## Ports and exports

Consumers own narrow ports when substitution is required. Graph adapters may
retain concrete publication, exact-read, traversal, lease, cleanup, and restore
capabilities that have production consumers. Compile-assertion-only aggregate
interfaces, unused constructors, candidate-validity hooks, clock or generation
options, optional reconcilers, and exports with no production consumer are not
supported extension points and are removed.

Positive allowlists define the stable internal architecture:

- the root import set is the standard library plus its private implementation;
- the root export set is the v2 invocation/result model, closed errors, and the
  `ProjectV2` entrypoint required by production callers;
- restore exports are only the current v4 cross-owner contracts and narrow
  ports used by production composition;
- adapter capabilities are publication, exact read, traversal, lease, cleanup,
  and mandatory-reconciler restore; and
- application composition has one current v4 source registry, binding, writer,
  reconciler, and participant path.

An extra import, export, constructor, adapter capability, or composition path
fails the boundary evidence. Permanent executable tests do not enumerate old
package or symbol names. Old-name searches are final cleanup evidence only;
the sole allowed historical matches are immutable authored SQL or frozen
handoff records and the owner-required inert migration-ledger verification
facts.

## Compatibility and transition

The repository is pre-production. The cutover is atomic by workstream and has
no deprecation period. No compatibility alias, dual reader or writer, feature
flag, Graph backup translator, fallback dispatcher, or inventory gate is
introduced.

Graph Projection retains `graph_projection.v2`,
`graph_projection_result.v2`, current valid-v2 canonical output, every existing
result/object identity algorithm, persistence shape, traversal, leases, and
cleanup behavior. Network Flow advances to public major 5 and state 4 while
retaining route paths and semantic-query v2. Recovery advances to Graph v4 and
supports no earlier Graph artifact. Database migrations 00032, 00033, and
00034 and their hashes remain unchanged.

Valid state-3 Network Flow installations may contain committed 1→2 and 2→3
migration ledger entries. Extensions therefore retains their identity,
lineage, from/to version, and definition digest as inert ledger-verification
facts. Those facts never register or invoke an apply or pending-state
validation algorithm. Only state 3→4 is executable. State 1 or 2 is rejected
before execution, and a v1 declaration makes 3→4 fail atomically without
rewriting declarations, digests, selected-result references, ledger, or state
metadata.

Once state 4 commits, production downgrade to a major-4 binary is unsupported.
An incompatible declaration or backup requires a separately adopted
remediation; this decision authorizes no silent delete or rewrite.

## Verification

Conformance requires positive root and restore import/export allowlists,
constructor-negative tests, current-only composition evidence, exact identity
fixtures, exhaustive semantic matrix and boundary tests, mandatory
reconciliation rollback evidence, and pre-mutation rejection of Graph v2/v3
artifacts. Verification is routed by machine-owned catalogs and consumes typed
contracts rather than this decision or another Markdown file.
