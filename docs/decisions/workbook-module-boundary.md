# Workbook Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-072.

This decision owns only the repository-internal Go topology, application
composition boundary, and compatibility removal for Workbook. Core 01 owns
public routes, request and response contracts, startup, query, mutation,
restore-probe, projection, and storage-boundary behavior. Core 02 owns source
state, record envelopes, history, and rollback meaning. Core 03 owns Workbook
interaction and Collaboration consequences. Core 04 owns security and
conformance. If this decision conflicts with an adopted behavioral owner, that
owner governs and this decision must be repaired.

Runtime code, tests, generators, conformance, and release evidence do not read
this file. Authored policy under `tools/` and owner-routed executable tests
enforce the adopted facts.

## Decision

Workbook is the generic interaction/application facade for the thirteen
Workbook-owned HTTP operations. It owns route registration, common
authentication and authorization coordination, public request and response
mapping, query and mutation dispatch, startup, bounded telemetry,
Workbook-owned recovery-state contribution, and the restore-probe registry and
executor assigned by Core 01 REQ-01-578.

The final boundary has these properties:

- `RegisterRoutes` is the sole server-facing Workbook entry point. Its private
  route families retain the exact adopted operation inventory and security
  precedence.
- One immutable contribution catalog indexes query, create, patch, conflict,
  clipboard, bulk, linked-note, and supersede capabilities by stable view or
  record identity. Construction rejects nil, typed-nil, duplicate, unknown,
  cross-surface, and incomplete contributions before serving.
- Workbook owns consumer-side provider interfaces, neutral route targets,
  conflict claims, common mutation results, and a closed safe mutation-failure
  vocabulary. It owns the final mapping to public HTTP status, code, and safe
  details.
- Source owners retain admission, normalization, defaults, canonical hashes,
  source mutation, history inputs, projection inputs, and Collaboration
  consequences. Workbook generic code contains no source field registry,
  source command or result type, source error switch, or opaque `any`
  admission channel.
- `internal/app/workbookassembly` is the sole concrete cross-owner adapter and
  catalog-construction boundary. It adapts module-native owner capabilities to
  Workbook-owned provider interfaces and validates the complete catalog.
- Workbook production code imports no concrete source-owner implementation.
  Its cross-module imports are limited to the narrow incident-admission
  capability and the immutable Projections provider contract, in addition to
  Workbook subpackages and platform packages.
- The adopted Projections boundary may continue to implement Workbook-owned
  query and restore-probe consumer interfaces at its existing application
  composition boundary. Incidents may continue to consume only the narrow
  Workbook startup bootstrap port. Neither exception transfers Workbook
  implementation or source semantics.
- Timeline declares the Base restore-probe registration, Workbook validates
  and executes the selected registration, Recovery selects the incident and
  publishes the result, and Projections executes the query. Graph Projection
  remains a distinct restore participant.
- Generic test inventory and assertions live under `internal/testutil`;
  Workbook test support describes only Workbook routes, startup, preferences,
  and cross-owner behavior visible through the Workbook facade. Source fixtures
  remain with their semantic owners.

The Workbook startup application and its PostgreSQL persistence adapter remain
cohesive Workbook-owned subpackages. Their narrow saved-view and bootstrap
ports do not transfer Saved Views or Incidents semantics to Workbook.

## Transition and compatibility

The transition is tracker rebaseline, owner adoption, verification-accounting
repair, owner-derived characterization, restore-probe closure, private route
decomposition, neutral contribution contracts, one source-owner adapter at a
time, test-support ownership, and final cleanup. Each workstream is validated
and recorded before its successor begins.

Temporary old and candidate Go surfaces may coexist only inside the active
workstream and are deleted when its callers cut over. There is no internal
deprecation window, alias, forwarding package, dual dispatch, fallback, or
runtime discovery mechanism.

No public HTTP or WebSocket method, path, operation ID, request or response
shape, error precedence, cursor, view schema, startup result, OpenAPI contract,
frontend contract, database schema, Incident Bundle shape, restore artifact,
or Collaboration event changes under this decision. Existing behavior is
retained only where an adopted owner requires it.

## Acceptance

The decision is implemented only when:

- the exact thirteen Workbook operations enter through one registrar and use
  only the immutable contribution catalog and neutral route ports;
- all active query, mutation, batch, and record-action capabilities have one
  complete application-composed contribution;
- source-owner admission, result, error, hash, and effect semantics are absent
  from generic Workbook production code;
- concrete cross-owner adapter construction exists only under
  `internal/app/workbookassembly`;
- Workbook's production cross-module import allowlist contains only the adopted
  incident-admission and Projections provider-contract entries;
- the old Store mutation facade, concrete provider constructors, opaque
  admissions, redundant owner aliases, and transitional tests are absent;
- restore-probe, startup, query, mutation, replay, authorization, effect-order,
  browser, boundary, generation, security, broad, and release checks pass; and
- the controlling handoff tracker records exact verification accounting, run
  roots, skipped checks, residual risks, and a completed final workstream.
