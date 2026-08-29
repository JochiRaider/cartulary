# Collaboration Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-075.

This decision owns only the repository-internal Go topology, application
composition, lifecycle boundary, source-owner port allocation, and
compatibility removal for the Collaboration implementation capability. It
does not create a domain bounded context. Workbook Interaction owns presence,
save/conflict, and live-update language. Core 01 owns the public WebSocket,
view-row, storage, replay, job-progress, extension-event, and portability
contracts. Core 02 owns source records, history, and conflict-fact meaning.
Core 03 owns Collaboration behavior and client consequences. Core 04 owns
security, deployment-local recovery authority, and conformance.

If this decision conflicts with an adopted behavioral owner, that owner
governs and this decision must be repaired. Runtime code, tests, generators,
conformance, and release evidence do not read this file. Authored policies and
machine projections under `contracts/**` and `tools/**` enforce adopted facts
without superseding their owners.

## Decision

Collaboration is a cohesive application and supporting capability that owns
semantic WebSocket admission, public event construction, durable incident
streaming, live session effects, and deployment-local recovery for that
stream. Its implementation root is `internal/modules/collaboration`.

The root exposes one application runtime facade and the sole server-facing
`RegisterRoutes` registrar. The facade owns construction, readiness, start,
shutdown, and reverse-order closure for Collaboration-private components.
Server assembly supplies borrowed platform dependencies and narrow consumer
capabilities; it does not construct or coordinate a concrete hub, dispatcher,
replay store, stream store, or recovery adapter independently.

The final boundary has these properties:

- semantic envelope, payload, socket, and close/error contracts live behind a
  narrow Collaboration-owned `protocol` boundary;
- concrete WebSocket libraries remain in Platform and enter Collaboration only
  through the semantic socket capability;
- route admission and session orchestration, the live hub, durable stream
  storage, dispatcher process mechanics, and recovery persistence are private
  owner implementation details;
- Collaboration owns the four Collaboration table meanings and authored SQL
  consequences, while PostgreSQL owns connectivity and its narrow `DB` port;
- the stream store owns append, sequence, intent-key, replay-token, replay,
  quarantine-state, and retention persistence;
- the dispatcher owns scheduling, notification listening, ordered tailing,
  retry/backoff, quarantine progression, retention cadence, and bounded
  lifecycle behavior;
- live session effects for Auth and Incidents are exposed through narrow
  capabilities that do not expose hub identity, storage, or process control;
- recovery exposes one narrow deployment-local capability to Operator; SQL,
  locks, proof checks, administrative-journal persistence, and transaction
  outcomes remain private recovery adapters; and
- the source-owned recovery-state contribution is pure data and exposes no
  live runtime or recovery authority.

No concrete hub, dispatcher, replay store, stream store, recovery service, or
free lifecycle constructor is a root-package coupling point in the final
topology. Tests observe semantic capabilities through the application facade
or test-only support rather than production concrete types.

## Independent source-owner fact ports

Private Revisions conflict facts and public Collaboration effects have
different consumers, disclosure rules, and evolution pressure. A source owner
derives both independently from its admitted semantic operation. Revisions
and Collaboration MUST NOT import one another, call one another, translate one
another's values, or reconstruct facts from the other representation.

### Revisions live input

Revisions owns `LiveRevisionInput`, `RevisionConflictFact`, and
`AppendLiveRevisionTx`. The live input contains the exact revision identity,
attribution, authoritative history snapshot, and explicit conflict facts
required by the Revisions owner. A conflict fact contains:

- one stable source-owned field key;
- explicit before-value presence and the before value when present; and
- explicit after-value presence and the after value when present.

Facts are unique by exact field key and ordered canonically before persistence.
Presence is distinct from JSON `null`; omission, deletion, and an explicit
null value MUST NOT be collapsed. Values are source-semantic admitted values,
not a whole-row JSON diff, public view cells, projection rows, portable
snapshots, or Collaboration payloads.

Revisions also exposes a distinct historical-revision operation. Historical
restore/import reconstruction records the owner-required history state but
does not accept live conflict facts and cannot trigger public Collaboration
publication. There is no Boolean historical switch on the live operation.

### Collaboration publication input

Collaboration owns `RecordChangeIntentInput`, `AffectedViewChange`, and
`AppendRecordChangedTx`. The input contains exactly:

- incident, record, change-set, and actor identities;
- authoritative row version;
- client transaction identity and mutation ordinal for the deterministic
  intent key;
- owner-supplied creation time;
- sorted unique public field keys; and
- sorted unique affected-view effects, each carrying its exact view schema,
  row identity, row version, and either a catalog-valid sparse public patch or
  explicit invalidation.

The operation validates and canonicalizes the public input, constructs exactly
one canonical `record_changed` intent, and appends it through the borrowed
source transaction. It does not read a source row, compare arbitrary JSON,
load a projection as source truth, append a revision, or infer whether a
private field is public.

### Publication catalog

Application assembly compiles one immutable publication catalog from
source-owner contributions. A contribution identifies one source owner and its
affected views; each affected-view declaration carries the exact
`view_schema_id`, admitted record types, public field keys, and patchable public
cell fields for that view. There is no contribution-wide record-type default.
Application assembly obtains canonical owner/view relationships from the
active projection descriptors and independently obtains canonical record types
and public fields from the view-schema resources. It joins those sources before
constructing Collaboration and fails when they disagree.

Catalog construction compares the complete contribution set with those
canonical views. Missing, extra, duplicate, unknown, typed-nil, cross-owner,
record/view-mismatched, or contradictory declarations fail before serving.
Construction defensively copies every admitted input. The resulting runtime
catalog retains only the public-field and patch-field policy needed for append
admission; it does not retain a second owner, record-type, or view registry.

The catalog is validation authority for internal publication inputs only. It
is not a view-schema registry replacement, a runtime callback registry, a
public discovery contract, or a mechanism for Revisions conflict facts.
Source owners keep source semantics and exact derivation. Collaboration keeps
public disclosure validation and wire construction.

### Family-specific publication capabilities

Collaboration exposes three distinct append capabilities: record change, job
progress, and extension-resource change. Each capability has its own narrow
consumer port, constructor, unexported implementation, and Runtime accessor.
There is no aggregate publication appender, aggregate constructor, or aggregate
Runtime accessor. A value implementing one narrow capability MUST NOT be
widenable through a type assertion to either other capability.

App Server supplies each source owner, Jobs, Network Flow, and test composition
only the capability that consumer uses. Adding a future replayable family adds
a capability for its actual consumers; it does not widen existing ports.

### Closed intents and replay validation

A production event intent has no exported mutable fields and cannot be created
with a caller-selected family string. Collaboration provides one named
constructor for each admitted replayable family. Each constructor supplies the
complete family identity atomically; record-change identity includes change-set
ID, record ID, and row version at construction. Intent keys, persisted family
values, source identity, mutation ordinals, and canonical payload bytes remain
unchanged.

One exhaustive protocol validator owns replayable payload admission, and one
validator owns the complete sequenced replayable message. The validation path
covers record change, job progress, and extension-resource change at append,
sequencing, recovery proof, durable tail reads, authenticated reconnect replay
reads, and hub delivery. Unknown families fail closed. Known families continue
to accept allowed additive members. A malformed durable row is not returned in
replay, fanned out, skipped, or allowed to advance the durable tail cursor;
later sequences remain blocked until the invalid state is repaired through the
adopted recovery or Operator boundary.

### Transaction ownership and sequence

The source application owns one transaction for a live semantic operation and
passes that already-open transaction to all participants. Revisions and
Collaboration borrow it: neither begins, commits, rolls back, nests, retries,
or closes it.

After owner authorization, admission, locking, and semantic validation, the
transactional effects occur in the owner-required dependency order and include
the source mutation, current envelope, live revision/private facts, projection
effects, exactly one public record-change intent, and idempotency result.
Every pre-commit failure leaves all of those effects absent. A response or
dispatcher failure after proven commit does not repeat the source mutation.
Exact idempotency replay returns the original result without another revision,
projection effect, or intent; divergent replay commits nothing.

Historical reconstruction uses the historical Revisions operation and omits
the Collaboration operation. Delete, restore, rollback, and explicit conflict
resolution use the same independent fact-port rule whenever their adopted
behavior requires a live public effect.

## Import and construction rules

Production import direction is acyclic:

- source owners may depend on consumer-owned narrow Revisions and
  Collaboration ports supplied by application assembly;
- Revisions and Collaboration do not import each other;
- neither capability imports source-owner production packages;
- application assembly may import source owners and both capabilities to bind
  ports and compile immutable catalogs;
- Platform supplies transport, PostgreSQL, authentication primitives,
  telemetry plumbing, and administrative-audit substrate through narrow
  dependencies but owns none of the Collaboration semantics; and
- Operator consumes only the narrow recovery capability and generated result
  contract.

Runtime construction is a closed, fallible operation. App Server supplies the
borrowed database, transport admission, publication catalog, resolved service
version, behavior clock where time is read, and required unexpected-dispatcher-
loss callback before construction. Collaboration constructs hub and dispatcher
telemetry with that resolved version and exposes no post-construction dependency
or telemetry configurator. A Runtime cannot escape with a missing or partially
installed mandatory dependency.

App Server owns translation of unexpected dispatcher loss into the adopted
mandatory publication component `websocket`. Dispatcher startup failure,
unexpected loop return, or loop panic reports that loss once per Runtime
lifetime and terminalizes the dispatcher. Graceful cancellation and retryable
run, listener, notification, or polling-fallback failures remain nonfatal and
do not trigger an in-process restart. No Collaboration-specific component
identifier is introduced.

Forwarding aliases, deprecated constructors, compatibility packages, dynamic
global registries, source-owner switches, generic row-diff helpers, generic
event-intent inputs outside Collaboration, and direct cross-owner Collaboration
SQL are forbidden in the final topology. Direct Collaboration-table SQL is
limited to owner storage/recovery tests and PostgreSQL role or physical-schema
evidence where physical realization is the test subject.

Reusable socket and semantic intent test support, message and payload builders,
semantic record-change decoding, narrow fakes, and persisted-corruption SQL
fixtures belong under `internal/testutil/collaborationsupport`. Production
protocol APIs expose behavior used in production, not fixture-only constructors
or semantic carrier types. Reusable application test composition belongs under
`internal/testutil/appsupport`. A module-local wrapper around the generic
application scenario runtime is not retained.

## Transition and compatibility

The repository is pre-production. The cutover is atomic by workstream and has
no internal deprecation period. Old and replacement surfaces may coexist only
inside one actively executing workstream and must be removed before that
workstream is complete. No alias, adapter preserving a retired API, dual
reader, dual writer, feature flag, backfill, or historical data translator is
introduced.

No public HTTP route, WebSocket message family, CLI grammar, database schema,
Incident Bundle version, authorization precedence, or telemetry vocabulary
changes under this decision. Known valid WebSocket messages remain
semantically equivalent. The separately owner-required decoder correction may
admit valid additive members, reject incomplete known messages, and return a
fresh known-member projection; that is a downstream conformance repair, not a
compatibility exception created by this decision.

An implementation discovery that requires an incompatible private persisted
state change must stop the applicable workstream and obtain owner adoption for
the reset-required diagnostic and reset-only path. It must not infer a
production migration or compatibility layer from this decision.

## Acceptance

The decision is implemented only when:

- one Collaboration runtime facade and one `RegisterRoutes` registrar own the
  full route and process lifecycle;
- concrete transport, hub, stream-store, replay-store, dispatcher, and recovery
  implementation types are private and server assembly receives only narrow
  capabilities;
- durable stream persistence and dispatcher orchestration are separate private
  components with unchanged storage and observable behavior;
- every live source-owner and Revisions-owned destructive path supplies
  independent explicit Revisions facts and Collaboration public effects in one
  borrowed transaction;
- historical revision reconstruction has no live conflict or publication
  effect;
- the immutable publication catalog exactly matches active projection
  owner/view descriptors and independent view-schema record/public-field
  resources, has exact per-view record facts, defensively copies inputs,
  rejects every missing, extra, duplicate, unknown, cross-owner, mismatched, or
  contradictory case, and retains only append-admission key policy;
- record-change, job-progress, and extension-resource publication use distinct
  least-capability ports and implementations, with no aggregate appender or
  widenable narrow value;
- event intents have private state, one named constructor per admitted family,
  and atomic record identity, with no caller-selected family string;
- the exhaustive payload and sequenced-message validators protect append,
  sequencing, recovery, durable tailing, authenticated replay, and hub
  delivery for all families without rejecting allowed additive members;
- malformed durable messages fail closed without replay inclusion, fan-out,
  cursor advancement, skipping, or reordering;
- Runtime construction is complete, immutable, fallible, and consistently
  versioned before escape; unexpected dispatcher startup or loop loss reports
  the adopted `websocket` component once, while graceful and retryable failures
  remain nonfatal;
- Revisions and Collaboration have no production import or representation
  dependency in either direction;
- no whole-row diff, projection-as-source inference, broad publication
  capability, broad event intent, combined revision/publication operation,
  retired constructor, mutable telemetry installation, invalid component ID,
  fixture-only protocol surface, forwarding alias, old test-support path, or
  unauthorized cross-owner SQL remains;
- recovery remains deployment-local and preserves its CLI, result, state,
  audit, concurrency, cancellation, timeout, and commit-unknown contracts;
- source-owner, application, socket, browser, stream, recovery, boundary,
  generated, harness, security, and broad checks pass; and
- the controlling tracker records every workstream checkpoint, compatibility
  outcome, rollback point, residual risk, exact validation root, and final
  handoff closure.
