# Graph Projection Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-076, with current behavioral owners
superseding earlier restore and lifetime descriptions.

This decision owns repository-internal Go package topology, import direction,
constructor posture, narrow port allocation and application composition. It
creates no separate domain bounded context. The Graph Projection NLSpec owns
projection semantics, completed-result integrity and derived reconstruction.
Network Flow owns saved declarations, graph-view identity, materialization and
public resources. Recovery owns authenticated selection, structural restore and
target lifecycle. Reporting owns admission and execution of frozen report
inputs. Core 04 owns security, deployment-local Recovery authority and
conformance; `docs/domain.md` owns vocabulary and owner navigation.

If this decision conflicts with an adopted behavioral owner, that owner
governs and this decision must be repaired. Runtime code, tests, generators,
conformance and release evidence do not read this file. Typed contracts and
verification policies project adopted facts without superseding their owners.

## Decision

Graph Projection is one pure deterministic engine with cohesive operational
adapters. Its root is `internal/modules/graphprojection`. It owns the closed v2
invocation and completed-result model and exposes one engine entrypoint:

```go
func ProjectV2(
    ctx context.Context,
    invocation InvocationContextV2,
    rawInput []byte,
) (ProjectionResultV2, error)
```

The root imports only the Go standard library and its private implementation.
Parsing, normalization, semantic validation, derivation, canonical encoding,
limits and cancellation remain Graph-owned. Identity derivation helpers remain
private. Narrow validation methods on the binding and completed-result models
let adapters prove identity, canonical content and exact row fidelity through
the same owner implementation. They are value validation, not another engine
entrypoint or a generic validation framework.

There is no engine instance, mutable option set, graph-view identity helper,
restore contract, database facade, HTTP surface, authorization or job facade in
the root. Network Flow derives graph-view identity privately and supplies it
through `InvocationContextV2`; trusted invocation values never enter semantic
configuration or its digest.

## Reconstruction ownership

`internal/modules/graphprojection/restore` owns current reconstruction
contracts, source-registry v5 construction, rebuild orchestration and the pure
Graph recovery-state contribution. It depends on the Graph root; the root does
not import it. Its current algorithm/result are v5 and implementation binding
is v7. The binding contains only the eight logical members assigned by Graph
Projection §9; build attestation and Core database generation remain separate.

`internal/modules/graphprojection/postgresrestore` realizes the transaction and
writer ports. Its fallible constructor requires the caller-owned transaction,
a committed-state reader and a reconciler. In one transaction it clears the
three derived tables, rebuilds and publishes exact immutable results,
reconciles Network Flow nonterminal materialization jobs and verifies the
required postconditions. Reconciliation cannot be omitted by construction.
Reporting jobs and retained inputs are outside this transaction.

Recovery admits the current frozen catalog and identities before mutation,
restores authoritative declarations structurally and leaves the three Graph
derived tables empty. It never invokes an inactive source adapter. Application
composition runs reconstruction only after successful source-profile claim
admission. Dependent Network Flow operations and new live Graph-input
admissions remain gated until reconstruction succeeds; already admitted
Reporting work validates and uses its retained input independently. Structural
Recovery completion and Graph availability are separate outcomes.

## Ports and exports

Consumers own narrow ports where substitution is needed. Production adapters
provide publication, exact read, traversal, cleanup and reconstruction. Every
content-returning read and publication validates completed-result integrity.
Metadata-only envelope locking preserves the result-before-declaration lock
order and the same-transaction selected-binding recheck. These locks do not
establish a persistent Graph lease or extend consumer lifetime.

Positive architecture allowlists cover:

- the standard-library/private-implementation root imports;
- v2 invocation/result models, their narrow validation methods, closed errors
  and the single `ProjectV2` engine entrypoint;
- current reconstruction contracts and ports used by composition;
- the concrete storage capabilities and mandatory reconciler; and
- one current source registry, binding and reconstruction path.

Unused aggregate interfaces, constructors, candidate-validity hooks, clock or
generation options, optional reconcilers and compatibility aliases are not
extension points. Production lifetime belongs to each admitted consumer's
immutable input. Graph lease persistence, acquisition, renewal, expiry,
release, configuration and telemetry have been removed.

## Compatibility and transition

The project is pre-production and current-only. Graph input/result v2 and their
valid canonical identities remain current. Network Flow is public major 8,
state/minimum 5 and semantic-query v2. No older Network Flow state migrates into
this generation. Existing incompatible state remains intact and unavailable;
this decision authorizes no conversion, deletion or environment reset.

Current Graph reconstruction uses algorithm/result v5, source registry v5 and
implementation binding v7. Earlier generations, aliases, dual readers,
translators and fallback dispatchers are unsupported. Immutable authored SQL
remains history. The Core-owned retirement migration removes the Graph lease
table only when empty under its transaction lock; it cannot expire or delete
rows to make admission pass. Current database membership and generation are
owned by the authored migration inputs and frozen Core catalog, not duplicated
version lists in this decision.

An unpublished rollback reverts the coherent source/projection change. Older
code must never reinterpret backups or completion evidence written by the new
generation. Release qualification and deployment remain separate work.

## Verification

Required evidence includes root/import/export boundaries, constructor-negative
cases, exact byte/identity vectors, semantic and capacity cases, publication
and stored-row tamper rejection, current source admission and cleanup races,
mandatory reconciliation rollback, and current capture/restore followed by
inactive and claimed process startup. Failed claim and reconstruction must
preserve the owner's availability and target-state contracts. Reporting
execution must remain independent of later source cleanup once input is frozen.
Machine-owned catalogs select these checks; no check consumes this document.
