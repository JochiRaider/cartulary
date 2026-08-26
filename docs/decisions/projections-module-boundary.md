# Workbook Projections Module Boundary Decision

## Status and scope

Status: Adopted through Core 00 REQ-00-070; Tasks/Decisions port topology
amended 2026-08-25.

This decision owns only the repository-internal Go topology and transition for
the workbook-grid Projections module. Core 01 owns projection behavior and Core
04 owns conformance. The Graph Projection NLSpec governs a different subsystem.
If this decision conflicts with an adopted behavioral owner, the owner governs
and this decision must be repaired.

Runtime code, tests, generators, conformance, and release evidence do not read
this file. Machine enforcement lives in authored contracts and policies under
`contracts/` and `tools/`.

## Decision

The final topology is:

```text
internal/modules/projections/
├── adapters/
├── providercontract/
├── testsupport/
└── internal/
    ├── runtime/
    ├── storage/
    └── queryengine/
```

- `internal/modules/projections/adapters` is the sole production construction
  boundary. Only `internal/app/projectionassembly` imports it in production.
- `providercontract` contains immutable provider descriptors, descriptor-set
  snapshots, enums, and semantic query intent. It exposes no Store, SQL,
  callback, executable provider, mutable catalog, transaction runner, table
  operation, or query-engine type.
- `testsupport` exposes named typed test capabilities only. Test permissions do
  not create production permissions.
- Nested `internal` packages contain catalog validation and coordination,
  projection-table persistence, and query compilation/materialization.
- The root `internal/modules/projections` directory contains no production Go
  package after the transition.

`adapters.New(Dependencies) (Ports, error)` is the only production constructor.
Dependencies contain `postgres.DB` plus exactly one required typed owner
contribution from Timeline, Entities, Indicators, Assessments, Artifacts,
Evidence, Parties, and Tasks/Decisions. Construction fails closed for a nil
database or required contribution, duplicate provider/table/view ownership,
missing active provider or dependency, unsupported descriptor version,
unresolved semantic query intent, invalid source/storage/facade ownership, or
incomplete/cyclic restore order. Failure returns a zero unusable `Ports`; every
required successful port is non-nil.

`Ports` exposes only:

- an immutable deep-copying descriptor set;
- Workbook's consumer-owned query provider;
- Recovery's consumer-owned projection rebuilder;
- Revisions' consumer-owned projection services;
- the immutable Projections recovery-state contribution;
- typed Timeline, Entities, Indicators, Assessments, Artifacts, Evidence,
  and Parties writer/rebuild/reader ports;
- a four-method Tasks/Decisions mutation-row port and a separate two-method
  Tasks/Decisions Reporting reader; and
- typed artifact, host/identity, and task/decision derived-fact readers used by
  source-owner Reporting providers.

Tasks/Decisions exposes no source-owner-specific rebuild port. Its task-request
and decision providers remain required, ordered members of the same immutable
provider catalog as every other workbook projection provider. Generic
catalog-driven Projections coordination exclusively performs their incident,
import, Revisions, and restore rebuilds. The mutation-row port contains only
typed refresh and load operations for task requests and decisions. The
Reporting reader contains only the two derived-fact collection operations.
Neither port can begin or commit a transaction, select rebuild scope, or invoke
provider-catalog coordination.

Projections does not import Reporting. Source owners retain
`exportprovider.FieldProvider` implementations, content class, source family,
fact selection, support-reference, and ordering semantics. Application
composition injects the typed derived-fact readers.

Owner contributions expose typed authoritative source access and semantic
intent. Their query intent contains stable view/field identities, kinds,
collection ordering, grouping eligibility, and semantic discriminator tokens;
it contains no projection table, SQL join, expression, predicate, or alias.
Private Projections plans bind that intent to physical SQL and validate it
against view schemas and descriptors.

Projection writers accept `context.Context`, caller-owned `pgx.Tx`, and typed
owner inputs. Generic `viewSchemaID`/`any` mutation dispatch and generic deletion
are removed. Only Timeline mutation deletion and typed host, identity, and
indicator deletion remain. No new deletion behavior is created.

## Transition and compatibility

The sequence is characterization, adapter/contracts, application consumers,
eight source-owner facades, ten physical providers in rebuild order, query-seam
closure, test capability migration, policy reconciliation, and root removal.
The Tasks/Decisions amendment additionally separates its source contribution,
mutation-row consumption, and Reporting consumption, then deletes its obsolete
rebuild facade after every generic caller is characterized. Every slice has an
independently green validation boundary. Temporary delegation may exist only
inside an active slice and is deleted when its last caller migrates. There is
no deprecation or release compatibility window for the repository-internal
root API, and no forwarding package or compatibility alias is retained.

Descriptor schema v3 and validation manifest v4 remain current unless their
serialized shapes change. Host and identity become query-capable projection
providers after their physical query paths move. Existing migrations remain in
place. Authored/generated Timeline query inputs are removed only through their
owner input and Make-owned generation when proved unused.

No public HTTP, WebSocket, cursor, authorization, saved-view, `view_row_v1`,
view-schema, error, telemetry, restore-result, or database-schema change is
authorized by this decision.

## Acceptance

The decision is implemented only when:

- root production imports and exports are empty;
- exact ten-table SQL ownership and four-way set equality pass;
- Tasks/Decisions has separate source-contribution, mutation-row, and Reporting
  contracts, with no owner-specific rebuild interface or adapter/runtime method;
- generic catalog-driven incident, import, Revisions, and restore rebuild
  evidence includes the task-request and decision providers in descriptor
  order;
- constructor, transaction, deletion, query, Reporting, restore, rebuild,
  telemetry, and test-capability matrices pass before and after migration;
- generated and migration drift checks pass; and
- explanatory documentation matches the adopted owners without becoming a
  runtime or verification input.
