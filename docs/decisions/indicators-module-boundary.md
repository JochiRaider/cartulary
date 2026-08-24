# Indicators Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-074.

This decision owns only the repository-internal Go topology, composition
boundary, construction posture, and internal compatibility posture for
Indicators. Core 01 owns public routes, request and response contracts, view
schemas, query and mutation behavior, source storage, portability, and
recovery. Core 02 owns canonical Indicator identity, observations, lifecycle,
history, rollback, and closed vocabulary. Core 03 owns Workbook interaction
and Collaboration consequences. Core 04 owns security and conformance.
`docs/domain.md` owns vocabulary navigation only. If this decision conflicts
with an adopted behavioral owner, that owner governs and this decision must be
repaired.

Runtime code, tests, generators, conformance, and release evidence do not read
this file. Authored policy under `tools/`, machine projections under
`contracts/`, and owner-routed executable tests enforce the adopted facts.

## Decision

Indicators remains one bounded source-owner module rooted at
`internal/modules/indicators`. Its root package is the internal Go facade and
owner transaction-coordination boundary. It is not a generic application
facade, a transport owner for other modules, or a reason to move generic
coordinators into Indicators.

Production code belongs to exactly one of these owner-local responsibilities:

- canonical Indicator identity, create, reuse, observation, and lifecycle
  application behavior;
- Indicators-owned HTTP admission and route adaptation;
- Indicators-owned import, revision, rollback, delete/restore, projection,
  Recovery, and Incident Bundle contributions;
- Indicators-owned source text and Records envelope coordination through
  typed capabilities supplied at application composition; or
- immutable owner-local vocabulary and source-state facts used to construct
  those operations and contributions.

The `admission` and `httpapi` child packages remain owner-local adapters.
`internal/identity`, `internal/origin`, `internal/vocabulary`, and
`internal/sourcestate` remain owner-private semantic packages. Providers under
`internal/providers` remain typed source-owner implementations. Runtime-
excluded `testsupport` remains test support, not a production API.

An export is retained only when it has a live production consumer, implements
a required interface method, or publishes a typed source contribution. Test
usage, capitalization, historical compatibility, and possible future use do
not justify a production export. The guarded root baseline at adoption is 55
declarations. The same remediation contracts this to exactly 50 by replacing
the exported `CreateOutcome` type and its four constants with the two
consumer-required `CreateResult.Created` and `CreateResult.Replayed` fields.
Iteration 2 preserves that exact count by replacing the test-convenience
`IndicatorFindOrCreateParticipantV1` export with the production
`RecordEnvelopePort` construction boundary; the participant schema value
remains private and its returned bytes remain unchanged. Test convenience is
never sufficient reason for a root export. Iteration 3 replaces `Store`,
`StoreDependencies`, `NewStore`, and `NewImportCreateFacade` with
`Application`, `ApplicationDependencies`, `NewApplication`, and
`NewImportContribution`, and adds the production `IdempotencyPort`,
`IncidentStatePort`, `RevisionPort`, and `NewProjectionContribution`
declarations. The resulting guarded root surface is exactly 54 declarations.
The owner fixes its import binding identity and publishes its projection
contribution directly; callers do not select either identity.
No alias, forwarding package, dual result, or deprecation period is allowed.
Any other export addition or removal requires an amendment to this decision
and its executable inventory.

The bounded module does not own generic Workbook behavior, frontend or grid
behavior, Records envelope persistence, generic Revisions, Recovery, Incident
Bundle, Projections, Imports, or Collaboration coordination, Network Flow
`binding_only` semantics, another owner's SQL, or generated-contract
authority. It does not turn runtime vocabulary or source-state facts into a
cross-module registry.

## Import, transaction, and composition rules

Application composition constructs the Auth, Incidents, Records, and
Revisions adapters used by Indicators. It injects narrow idempotency,
incident-state, record-envelope, revision, projection, source-text, and clock
capabilities into the root Application. It supplies the same composition-owned
Records, incident-admission, and Auth session capabilities to HTTP routing.
Indicators code constructs no concrete peer or platform store. The
owner-local `admission` and `httpapi` adapters
may import root DTO and application-operation contracts; no production child
imports the root to recover dependencies, and no production child imports
`internal/app`. The root may import owner-local children and narrow
coordinating contracts needed to publish its contributions, but it does not
construct another owner's concrete store.

Records remains authoritative for current record envelopes. Indicators may
consume Records-owned envelope operations in a caller transaction, including
one deterministically ordered locking read for a mutation's complete affected
record set. Indicators must not duplicate that read with private SQL or map a
Records storage failure to a concealed semantic not-found result. Semantic
role validation and owner-classified error translation remain Indicators
application behavior after a successful Records read.

A caller-supplied `pgx.Tx` is borrowed. An Indicators operation receiving it
does not begin, commit, roll back, nest, detach, or replace the transaction.
Owner-local PostgreSQL stores remain permitted for Indicators-owned source
rows. The compound `loadByDedupeTx` query remains an Indicators-specific
coordination query during this remediation; this decision does not authorize
its decomposition or change its lock order.

Recovery and Incident Bundle contributions derive from one immutable,
fallibly constructed owner-local source-state catalog. Generic coordinators
receive typed contributions and do not import Indicators internals. The
catalog models authoritative, rebuildable, and portability descriptors but
does not own portable row shape, Recovery mechanics, or generic coordination.

## Construction and compatibility

Dependencies are supplied explicitly at application composition. A
constructor that returns an operational capability is fallible and succeeds
only when every required dependency and descriptor is valid. It rejects nil
and typed-nil dependencies without panic, returns no partial capability on
failure, and reports deterministic owner-safe errors. Immutable facts are
defensively copied; no mutable registration API is exposed.

Indicators Application construction requires Postgres, `IdempotencyPort`,
`IncidentStatePort`, `RecordEnvelopePort`, `RevisionPort`, projection rows,
source text, and a clock. HTTP construction separately requires the owner
application, a Records envelope reader, incident admission, Auth session
store/slider capabilities, and `DependencySet.Now`; it does not require
Postgres merely to construct those capabilities. Neither boundary constructs
a peer store or falls back to wall-clock time. Create, observation,
transition, lifecycle, and list orchestration lives directly on Application;
self-referential service
objects, empty repository namespace receivers, and Revisions forwarding
adapters are prohibited. Concern-specific SQL remains in named package
functions.

Root mutation operations and transaction-participant commands receive only
the authenticated actor user UUID required for attribution. They reject the
nil UUID before replay lookup, transaction acquisition, or another side
effect. HTTP retains the authenticated principal only at the transport edge;
Workbook, Imports, Network Flow, and HTTP pass the actor UUID into Indicators.
Imports does not synthesize an authentication record.

Indicators Workbook admission uses the transport-neutral strict-JSON
primitive and returns only the existing closed validation error containing
`Field` and `ReasonCode`. Workbook assembly alone translates that error into
its mutation failure and public HTTP error envelope. Indicators production
admission does not import the HTTP transport package or HTTP JSON wrapper.

The root constructs its projection contribution directly from the private
projection provider; the forwarding `projectionprovider` package is absent.
The root import contribution fixes `cartulary.view.indicators.v1` and
`indicators.import_create` internally. The generated Imports catalog remains
an independent startup consistency check rather than the source of owner
identity.

Indicators derives replay identity from the exact validated logical command;
callers do not supply request hashes. The five deployed JSON preimages for
create, observation create, observation resolve, dismiss/restore, and
lifecycle append retain their existing member presence, null, route/scope,
and support-reference ordering semantics and are hashed with SHA-256. Hashes
are derived before normalization or sorting can alter a deployed preimage.
Existing idempotency rows replay without rewriting, alternate digests, or a
version compatibility branch.

Package `init` registration, mutable global registries, service locators,
fallback dependency lookup, and runtime plugin discovery are prohibited.
Package-local immutable vocabulary and descriptor values are permitted only
when their membership is exact, deterministic, defensively copied, and has no
registration API.

No public HTTP or WebSocket path, operation ID, request or success shape,
view-schema field, authorization order, cursor binding, idempotency behavior,
history, projection, Collaboration effect, Incident Bundle identity or valid
bytes, generated protocol, or database schema changes merely because this
boundary is adopted. Owner-conformance corrections separately adopted in Core
01, Core 02, and Core 04 are not compatibility obligations to preserve known
defects.

## Acceptance

The decision is implemented only when:

- every root export has an exact reviewed disposition and role, the surface is
  exactly 54 declarations after the Iteration 3 Application, capability, and
  contribution cutover, and a synthetic unapproved export is rejected;
- production imports match the closed owner-local topology; only `admission`
  and `httpapi` import root contracts, and no child imports the root or
  application assembly to recover dependencies;
- Records supplies one sorted locked envelope snapshot for affected-record
  validation and the two standalone Records SQL validation helpers are absent;
- caller transactions are borrowed and cross-owner effects remain behind
  typed ports supplied at application composition;
- root and HTTP construction reject required nil and typed-nil dependencies,
  use composition-owned Records, incident-admission, and Auth capabilities,
  and use injected clocks without wall-clock fallback;
- orchestration is owned directly by Application, concern SQL uses named package
  functions, and no self-referential service, empty repository namespace, or
  Revisions forwarding adapter remains;
- root operations accept actor user UUIDs, reject the nil UUID before effects,
  and import execution creates no synthetic authentication record;
- Workbook admission returns only its closed semantic validation error,
  imports strict JSON directly, and has exactly one Workbook-owned
  semantic-to-wire translation;
- projection and import contributions are constructed at the root, the
  projection forwarding package is absent, and caller-selected Indicator
  contribution identities are impossible;
- Indicators alone derives the exact five deployed replay hashes before
  preimage-changing normalization, while persisted replay remains compatible;
- one immutable runtime vocabulary owns Indicator type, value-kind,
  observation-status, and lifecycle-state membership while context-specific
  admission behavior remains explicit;
- one validated source-state catalog produces the exact authoritative,
  rebuildable, and portability inventories without coupling generic
  coordinators to Indicators internals;
- no generic facade, frontend/grid behavior, private cross-owner SQL,
  package-init registration, mutable global registry, service locator,
  fallback, alias, or dual-dispatch path exists;
- every active Indicators test has exactly one compatible authored harness
  selector; and
- focused, service-backed, boundary, migration, generated, build, and full
  repository gates pass with evidence recorded in the controlling tracker.
