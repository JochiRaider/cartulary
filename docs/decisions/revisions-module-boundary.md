# Revisions Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-071 and amended with REQ-00-075.

This decision owns only the repository-internal Go topology, composition
boundary, and compatibility removal for Revisions. Core 01 owns application,
route, portability, and storage-boundary behavior. Core 02 owns history,
snapshot, mutation-target, association, and rollback meaning. Core 03 owns
Collaboration consequences. Core 04 owns security and conformance. If this
decision conflicts with an adopted behavioral owner, that owner governs and
this decision must be repaired.

Runtime code, tests, generators, conformance, and release evidence do not read
this file. Machine projections under `contracts/revisions` and authored policy
under `tools/` enforce the adopted facts.

## Decision

Revisions is a bounded context for generic immutable history and destructive
coordination. Its application facade is composed by
`internal/app/revisionassembly`; source owners construct their own providers
and application assembly validates and copies the complete catalogs before
serving.

The final boundary has these properties:

- `Appender` receives a narrow transaction-bound current-envelope reader. It
  never constructs or imports a concrete Records store.
- A source-owned snapshot provider captures the current authoritative row in
  the caller transaction. The captured value is opaque to generic
  coordination and has the canonical envelope
  `{snapshot_schema_id, record, source}`.
- Each ordinary live record revision receives explicit source-semantic
  `RevisionConflictFact` values through `AppendLiveRevisionTx` in the caller
  transaction. A fact names one stable field key and represents before and
  after presence and value independently. Revisions validates and persists the
  facts but does not derive them from snapshots, whole-row JSON, projections,
  view cells, Collaboration effects, or public payloads. These facts support
  revision-window conflict detection for values intentionally excluded from
  canonical snapshots; they never become row-history, rollback, projection,
  publication, or portability authority.
- Historical reconstruction uses a distinct historical-revision operation
  that accepts no live conflict facts and has no Collaboration-publication
  effect. A live operation does not accept a historical/suppression flag.
- One immutable target-semantics catalog is compiled from source-owner
  contributions. Every exact target kind has one pure history facet and one
  rollback dispatch class. Unknown, duplicate, incomplete, typed-nil, or
  cross-owner contributions fail startup.
- A pure history facet supplies sorted unique history associations and
  addressability before persistence. Generic history lookup uses indexed
  association facts and contains no source JSON-key predicates.
- Row targets declare admitted record types and use source-owned row
  providers. Non-row targets adapt source-owned fixed-SQL providers. Generic
  rollback code may branch only between `row` and `non_row`; source vocabulary,
  companion selection, revalidation, and inverse application remain behind
  source facets and providers.
- Revisions retains its history-local transaction order, canonical locking,
  idempotency participation, append-only history persistence, safe error
  mapping, and stable opaque selectors. Source application coordination owns
  the complete mutation sequence; projection refresh and Collaboration intent
  publication are independent transaction participants and are not Revisions
  responsibilities.
- Revision assembly copies immutable source-owned conflict descriptors at
  startup. Revisions contains no global view-schema or Collaboration
  publication lookup.
- Conflict key-ring parsing consumes an explicit environment snapshot or
  resolver. Server assembly owns process-environment capture; an explicit
  empty input never consults the host process.

Providers never own authorization, transport decoding, transaction
completion, idempotency, history append, projection refresh, Collaboration
publication, dynamic SQL, runtime relation names, network calls, or object
storage. Revisions and Collaboration do not import, construct, or translate
one another. Source owners derive their private revision facts and public
Collaboration effects independently and pass them to separate consumer-owned
ports through one borrowed transaction.

## Transition and compatibility

The original transition completed canonical snapshots/events, indexed
history, and provider-driven rollback. The Collaboration separation is one
additional atomic repository-wide cutover to `LiveRevisionInput`,
`RevisionConflictFact`, `AppendLiveRevisionTx`, and the distinct historical
operation. Temporary old and candidate Go surfaces may coexist only inside
that active cutover workstream and are deleted before it completes. There is no
internal deprecation window or runtime fallback.

Existing databases containing Revisions mutation rows from before the
canonical snapshot boundary are pre-production disposable state and must be
reset. The boundary migration fails with an explicit reset-required diagnostic
when such rows exist. It does not backfill or infer history facts.

Schema-less snapshots have no reader, translator, alias, dual-write path, or
shape inference. Incident Bundles containing them are rejected. Bundle
versions 1 and 2 remain supported for canonical snapshots, with their existing
outer row members and ordering; deterministic association facts are recomputed
from the admitted target-semantics version during import.

No public HTTP or WebSocket operation, authorization precedence, opaque
selector, conflict-token v3 wire format, Incident Bundle version, OpenAPI
operation, UI selector, or frontend port changes under this decision.

## Acceptance

The decision is implemented only when:

- all ten record types have one source-owned snapshot schema and all fourteen
  target kinds have exactly one compiled target-semantics entry;
- every persisted non-null snapshot validates its declared schema;
- history association arrays are canonical, complete, indexed, and the sole
  generic mutation-history lookup source;
- live conflict facts are revision-bound, field-keyed, transactionally atomic,
  derived explicitly by source owners, and used only for
  optimistic-concurrency reconstruction;
- the live and historical operations are distinct, and neither Revisions nor
  its fact representation imports, derives, or publishes Collaboration
  effects;
- generic Revisions production code contains no concrete Records
  construction, projection snapshot read, global view-schema lookup, ambient
  environment read, source JSON-key history predicate, or source-kind rollback
  switch;
- source-owner, failure-atomicity, security-precedence, portability, browser,
  migration, boundary, generation, and broad release checks pass; and
- the controlling handoff tracker records the reset instruction, validation
  roots, residual risks, and a completed final slice.
