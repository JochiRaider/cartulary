# Entities Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-073.

This decision owns only the repository-internal Go topology, composition
boundary, and compatibility posture for Entities. Core 01 owns public routes,
request and response contracts, view schemas, query and mutation behavior,
source storage, imports, recovery, and portability. Core 02 owns canonical
record, identifier, history, and rollback meaning. Core 03 owns Timeline
interaction and Collaboration consequences. Core 04 owns security and
conformance. `docs/domain.md` owns vocabulary. If this decision conflicts with
an adopted behavioral owner, that owner governs and this decision must be
repaired.

Runtime code, tests, generators, conformance, and release evidence do not read
this file. Authored policy under `tools/`, machine projections under
`contracts/`, and owner-routed executable tests enforce the adopted facts.

## Decision

Entities remains one bounded source-owner module rooted at
`internal/modules/entities`. The root is not a generic application facade and
the current package is not split merely because it publishes several typed
source-owner contributions.

Every production export from the root and every production child package
belongs to exactly one of these responsibilities:

- register the two Entities-owned HTTP operations and preserve their existing
  authentication, authorization, CSRF, decoding, and error precedence;
- publish Entities-owned source contributions to Revisions, Recovery, and
  Incident Portability;
- implement entity bundle export, import, and subtype declarations;
- expose entity-owned merge and mention application operations; or
- compose typed owner-local providers from direct child packages for one of
  those responsibilities.

An export is retained only when it has a live production consumer, implements
a required interface method, or publishes a source contribution. Test usage,
capitalization, historical compatibility, and possible future use do not
justify a production export. Runtime-excluded owner-local test-support
packages are outside this production inventory and remain subject to their
separate import boundary. The executable boundary projection records every
production export with an exact `retain`, `privatize`, `remove`, or `replace`
disposition and rejects additions that have not been adopted here.

The bounded root does not own generic Workbook behavior, frontend or grid
behavior, Timeline automatic-resolution policy, the Timeline source-row
transaction, generic Revisions, Recovery, Portability, Projections, or
Reporting coordination, another owner's SQL or event construction, or hidden
dependency lookup.

The production import and composition topology has these properties:

- application assembly may import the Entities root and its typed child
  surfaces; a child package never imports the Entities root or application
  assembly to recover dependencies;
- the root may import owner-local children and the narrow coordinating
  contracts needed to publish its source contributions, but it does not import
  concrete coordinating stores;
- `hostidentity` publishes separate capabilities for its materially distinct
  consumers: Workbook receives separately named mutation and query
  dependencies and never an aggregate projection store;
  Timeline and Assessments receive stateless source facts whose operations
  borrow the caller transaction; Imports receives an import-create facade over
  the owner-private mutation core; and merge receives one immutable
  owner-local merge capability;
- `mentions` is consumed by Timeline and Workbook composition through
  caller-transaction commands and explicit effect ports;
- `merge` is consumed by Timeline/entity-merge composition and the root route
  facade, with every cross-owner effect injected;
- `timelinefacts` is consumed only by Timeline projection composition and
  returns Entities-owned facts rather than Timeline presentation policy;
- `projectioncontract` publishes only the source-facing Host/Identity
  projection inputs, pages, descriptors, contribution, and source reader;
  surface-intent construction remains package-private;
- `projectionports` publishes separate `MutationRows`, `QueryReader`, and
  `ReportingReader` capabilities; application assembly and every downstream
  consumer receive only the capability used by that role, while Projections
  retains physical projection storage and one private implementation;
- `mutationadmission` publishes immutable semantic admission failures with a
  closed reason vocabulary and optional typed detail facts; child decoders do
  not choose HTTP status, wire error code, message, or arbitrary detail-map
  shape, and translation occurs only in the Entities HTTP root or Workbook
  application assembly;
- application assembly may import the narrow owner-local `entitycontract`
  solely for canonical Host and Identity schema IDs; forwarding schema-ID
  constants in feature packages are prohibited; and
- the root itself is consumed directly only by Server, Revision, Recovery, and
  Incident Portability assembly.

A cross-owner consumer owns the language of its injected port. Translation
between that language and another owner's concrete implementation types,
including classified error translation, belongs in application assembly.
Owner production packages do not import another owner's concrete
implementation merely to adapt it to a consumer port.

New direct consumers or imports require an amendment to this adopted decision
and its executable boundary projection. Package convenience is not sufficient
authority.

## Transaction and cross-owner rules

A caller-supplied `pgx.Tx` is borrowed. An Entities operation receiving it does
not begin, commit, roll back, nest, detach, or replace the transaction. It does
not execute private SQL for another source owner or publish that owner's event
shape. Classified errors remain machine-classifiable across the port.

Timeline retains automatic-resolution eligibility, normalization,
suppression, ambiguity, source mutation, transaction lifecycle, Links,
History, Projection and Collaboration coordination, disclosure, and Undo.
Entities supplies incident-scoped alias facts, target validation, mention
persistence, and mention-history facts through typed operations that borrow
Timeline's transaction.

The injected `mentions.TimelineEffectsPort` and
`merge.TimelineEffectsPort` remain command/fact boundaries. They do not
authorize Entities to import a concrete Timeline store, reconstruct Timeline
rows or disclosure policy, or publish Timeline events directly.

Owner-local PostgreSQL stores remain permitted for Entities-owned source rows.
Source contributions may read declared cross-source facts only through their
adopted descriptor and port boundaries; they do not acquire ownership of the
other source or of derived projection storage.

## Construction and compatibility

Dependencies are supplied through dependency structs at application
composition. A constructor that returns an operational capability is fallible
and succeeds only when every dependency required by any operation on that
capability is present. It rejects nil and typed-nil dependencies without
panic, returns a nil capability on failure, and reports missing dependencies
deterministically in declaration order. Required dependencies do not use
variadic options, `Must` constructors, delayed fallback, or partial objects.
Stateless capabilities with no construction-time dependency may use an
infallible zero-argument constructor.

Package `init` registration, mutable global registries, service locators,
fallback dependency lookup, and runtime plugin discovery are prohibited.
Package-local immutable descriptor values are permitted when they are
complete, deterministic, and have no registration API. Dependency validation
helpers remain owner-package-local rather than creating a generic cross-module
construction abstraction.

There is no internal deprecation window, forwarding package, alias, dual
dispatch, fallback, or compatibility shim for this decision. Public Entities
JSON admission accepts exactly one duplicate-free object and rejects a scalar,
top-level `null`, duplicate member at any nesting level, or trailing JSON value
with `400 invalid_mutation_payload` and
`error.details.reason_code=request_not_object`. This is an intentional
tightening for previously ambiguous invalid forms. Valid requests, normalized
request hashes, idempotent replay, successful response bytes, routes, OpenAPI
identities, schema IDs, field keys, database schema, Incident Bundle shape,
authorization and concealment order, and source behavior remain unchanged.

## Acceptance

The decision is implemented only when:

- every root and production-child export has one exact adopted disposition,
  every retained export has a live production consumer, required interface
  method, or source-contribution role, and a synthetic unapproved export is
  rejected;
- the direct-consumer and import topology is enforced by authored machine
  policy and owner-local tests;
- no child imports the root or application assembly to discover dependencies;
- every caller transaction is borrowed and every cross-owner effect remains
  behind a typed port;
- Timeline retains automatic-resolution policy and transaction ownership;
- every operational capability is complete after successful construction,
  nil and typed-nil dependencies fail without panic, and no partial, option,
  `Must`, or fallback construction path exists;
- Workbook, Timeline, Assessments, Imports, and merge consume only their
  declared Host/Identity capability;
- projection source contribution, mutation, query, and reporting compile
  against their separate directional contracts, with no aggregate port or
  forwarding package;
- every Entities JSON decoder returns only semantic admission facts, and the
  HTTP root or Workbook application assembly performs the one applicable wire
  translation;
- consumer-owned ports cross owner boundaries and concrete translation remains
  in application assembly;
- no generic facade, private cross-owner SQL/event construction, package-init
  registration, mutable global registry, or service locator exists;
- every active Entities test has exactly one compatible exact authored harness
  selector; and
- focused, service-backed, boundary, generated, frontend/browser when
  affected, and full repository gates pass with evidence recorded in the
  controlling handoff tracker.
