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

Every production export from the root belongs to exactly one of these
responsibilities:

- register the two Entities-owned HTTP operations and preserve their existing
  authentication, authorization, CSRF, decoding, and error precedence;
- publish Entities-owned source contributions to Revisions, Recovery, and
  Incident Portability;
- implement entity bundle export, import, and subtype declarations;
- expose entity-owned merge and mention application operations; or
- compose typed owner-local providers from direct child packages for one of
  those responsibilities.

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
- `hostidentity` is consumed by Workbook, Imports, Timeline, and Assessment
  composition through typed APIs;
- `mentions` is consumed by Timeline and Workbook composition through
  caller-transaction commands and explicit effect ports;
- `merge` is consumed by Timeline/entity-merge composition and the root route
  facade, with every cross-owner effect injected;
- `timelinefacts` is consumed only by Timeline projection composition and
  returns Entities-owned facts rather than Timeline presentation policy;
- `workbookprojection` is consumed by Projections, Workbook, Imports, and
  Timeline composition and does not own physical projection storage; and
- the root itself is consumed directly only by Server, Revision, Recovery, and
  Incident Portability assembly.

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

Dependencies are supplied through constructors, options, or application
composition. Package `init` registration, mutable global registries, service
locators, fallback dependency lookup, and runtime plugin discovery are
prohibited. Package-local immutable descriptor values are permitted when they
are complete, deterministic, and have no registration API.

There is no internal deprecation window, forwarding package, alias, dual
dispatch, fallback, or compatibility shim for this decision. No public HTTP or
WebSocket operation, OpenAPI identity, schema ID, field key, database schema,
Incident Bundle shape, generated contract, authorization rule, or source
behavior changes merely because this boundary is adopted.

## Acceptance

The decision is implemented only when:

- every root export belongs to the closed responsibility list;
- the direct-consumer and import topology is enforced by authored machine
  policy and owner-local tests;
- no child imports the root or application assembly to discover dependencies;
- every caller transaction is borrowed and every cross-owner effect remains
  behind a typed port;
- Timeline retains automatic-resolution policy and transaction ownership;
- no generic facade, private cross-owner SQL/event construction, package-init
  registration, mutable global registry, or service locator exists;
- every active Entities test has exactly one compatible exact authored harness
  selector; and
- focused, service-backed, boundary, generated, frontend/browser when
  affected, and full repository gates pass with evidence recorded in the
  controlling handoff tracker.
