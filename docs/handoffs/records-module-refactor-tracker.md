# records Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/modules/records` |
| Normalized target label | `records` |
| Output path | `docs/handoffs/records-module-refactor-tracker.md` |
| Repository baseline | Branch `main`; initial planning commit `bf9febefd400764e576ae4504d3e0171dbbd2d87` |
| Document class | Planning tracker and implementation handoff; not an owner NLSpec |
| Allowed change in this session | This tracker file only |
| Forbidden changes in this session | Production code, tests, owner documents, contracts, generated artifacts, package configuration, migrations, and harness inputs |
| Implementation gate | A later explicitly authorized task is required before any implementation or owner-document change |

The label `records` is lowercase kebab case and contains no spaces, path
separators, shell metacharacters, or unsafe filename characters.

This tracker uses **MUST**, **MUST NOT**, **SHALL**, **SHOULD**, and **MAY** to
make planning gates, sequencing, and acceptance criteria unambiguous. Those
terms do not make this tracker a product-behavior authority. Adopted owner
documents remain authoritative. Proposed requirements in this tracker become
implementation-conformance requirements only after adoption by their named
owners.

### Source hierarchy

1. Adopted subsystem NLSpecs, within their named subsystems.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests for current implementation state.
6. Prior trackers, handoffs, research notes, and the planning framework as
   evidence only.

No adopted-owner contradiction was found. A later contradiction MUST be
recorded as `BLOCKED: owner contradiction`; an implementer MUST NOT choose a
side. Core 05 is inapplicable because this work publishes no claim and changes
no timed or fixture-sensitive publication.

### Owner documents and planning doctrine inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/extension-subsystem-nlspec.md`, for adjacent provider-boundary evidence
- `docs/domain.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/research/nlspec-spec.md`, as writing doctrine rather than product
  authority
- `temp/analysis-notes.md`, as decision evidence rather than product authority

Core 01 owns Incident Portability and already requires the Records source
family, the three declared Records invariants, one final import transaction,
and the safe failure envelope. Core 02 owns the first-class record envelope.
Core 03 owns collaboration and concurrency consequences. Core 04 owns
authorization outcomes and conformance acceptance. Reporting and Testing
Harness NLSpecs own their named subsystem behavior. Supporting appendices,
guides, catalogs, and this tracker MUST NOT redefine those owners.

### Proposed owner-document adoption

The identifiers below were unused when inspected. They are reserved by this
plan but remain non-authoritative until adopted.

| Proposed ID | Intended owner text | Adoption state |
| --- | --- | --- |
| `REQ-00-067` | Establish precedence between authoritative current envelope state and Revisions-owned history state. | adoption pending |
| `REQ-01-649` | Adopt `module.records` as the current-envelope owner; define relation ownership, transaction boundaries, and SQL access. | adoption pending |
| `REQ-01-650` | Define the Revisions-owned `DeleteRestoreSource` consumer port, catalog, source-adapter construction, and assembly rules. | adoption pending |
| `REQ-01-651` | Define the contract-major-1 portable Records row, validation rules, subtype registry, and safe rejection behavior. | adoption pending |
| `REQ-02-262` | Distinguish current envelope state from history and close the current first-class record-type registry. | adoption pending |
| `AC-509` through `AC-516` | Add binary acceptance criteria mapped in Section 12. | adoption pending |

### Repository evidence inspected

All 16 files under `internal/modules/records` were opened and are inventoried in
Section 2. Supporting evidence included exact application-composition callers,
Workbook and Revisions routes and stores, Reporting and Incident Bundle
consumers, source catalogs, OpenAPI owner fragments, recovery fixtures, schema
ownership, backend boundaries, verification owners, test families,
test-support inventory, generated-artifact policy, and related current
handoffs. Searches enumerated 69 Go files importing a Records package; exact
source files were opened before their behavior was used as evidence.

### Repository and authority mismatches

| Finding | Current evidence | Planning disposition |
| --- | --- | --- |
| The modular planning framework does not list Records. | Live verification and Incident Bundle catalogs contain `module.records`; production modules import it. | Treat Records as a valid live internal owner refinement; adopt that refinement in Core 01. |
| An earlier Timeline handoff says Records lacked an owner catalog. | The live repository now contains the catalog with two portability rows. | Treat the older statement as historical evidence only. |
| Records test support is registered as owner-local. | Its eight files and callers span application startup, Auth, Incidents, Workbook, Revisions, Projections, Timeline, Entities, Indicators, Links, and other owners. | Split by semantic owner after a source-derived symbol/caller matrix. |
| Incident Portability assigns logical Records ownership, while schema and recovery projections assign `records` to Revisions. | Core separates current source state from history; live metadata conflates them. | Planning decision: Records owns current envelope state and the physical `records` relation; Revisions owns history and coordination. Adoption is pending. |
| The source port declares three invariants but local validation visibly proves only a weak incident-scope check. | Core 01 already requires all three invariants and fail-closed rejection. | Treat current admission as nonconforming; runtime correction still requires later authorization. |

Implementation requires a separate authorized task. This tracker revision does
not adopt owner text, change runtime behavior, or perform the refactor.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Required target owner after adoption | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/records/deleterestore/provider.go` | Defines the shared source-provider interface and generic table-backed snapshot, tombstone, view-schema, and delete-precondition implementation consumed by Revisions. | `SourceProvider`, `TableProvider`, and four methods. | Revisions catalog; Artifacts, Assessments, Entities, Evidence, Indicators, Parties, Tasks/Decisions, and Timeline providers. | `pgx`, UUID, JSON, dynamic owner-table SQL selected from exported fields. | Revisions catalog, route, rollback, recovery, and source-owner suites. | Revisions delete/restore/history contracts and owner accounting; no generated file is authored here. | Consumer interface and catalog move to Revisions; concrete behavior moves to each source owner; exported `TableProvider` retires. | high | Method signatures and errors remain frozen for the first migration; no compatibility package. |
| `internal/modules/records/destructive_lock.go` | Sorts, deduplicates, and locks envelopes with fail-fast `FOR UPDATE NOWAIT`; distinguishes contention and missing rows. | `ErrRecordEnvelopeNotFound`, `ErrDestructiveOperationRecordLocked`, typed lock error, lock function. | Revisions delete/restore/rollback; Entities merge adapters. | `pgx`, SQLSTATE `55P03`, UUID sorting, direct `records` read lock. | Revisions and Entities destructive-flow tests. | Revisions public error and destructive-contention contracts. | Records envelope facade. | high | Caller owns the transaction; empty input succeeds; partial success is not returned. |
| `internal/modules/records/incident_bundle_portability.go` | Exports ordered `data/records.ndjson` and imports rows with attribution remapping through the registered relation. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`. | Records source port and Incident Bundles assembly. | Incident portability helpers, `pgx`, direct `records` SQL. | Direct Records source-port tests and broader Incident Bundles suites. | Incident Bundle source catalog and generated topology downstream of authored inputs. | Records portability adapter. | high | Strict parsing and subtype completeness are required but not implemented locally. |
| `internal/modules/records/incident_bundle_source_port.go` | Declares the Records source-family descriptor, prepares/applies files, and runs post-import validation. | `NewIncidentBundleSourcePort`. | `internal/app/incidentportabilityassembly/catalog.go`. | Incident Bundles `sourceport`, portability functions, validation SQL. | Direct Records duplicate/rollback tests and Incident Bundles integration. | `data/records.ndjson`, contract major 1, current `record-revisions` relation identity, and three Records invariants. | Records portability adapter. | high | Current relation identity must become `record-envelope`; no runtime alias is planned. |
| `internal/modules/records/incident_bundle_source_port_test.go` | Characterizes duplicate stable identity and transactional apply/rollback. | Two Go tests. | Current `module.records` unit and integration rows. | Testcontainers/Postgres support, source port, NDJSON fixtures. | It is the complete direct Records package test surface. | Records verification owner and test-family manifests. | Records verification evidence. | medium | Direct envelope, resolver, lock, Reporting, and invariant evidence remains to be authored. |
| `internal/modules/records/reportingprovider/provider.go` | Emits nondeleted envelope facts to Reporting through a caller-supplied transaction. | `CollectFieldsTx`, `CollectFactsTx`. | Reporting export materializer. | Reporting provider interface, `pgx`, direct fixed Records read. | Reporting boundary, export-model, and integration tests. | Reporting provider contract and boundary/read-shape rules. | Records source-owner Reporting adapter. | medium | Reporting retains composition, validation, rendering, redaction, and publication. |
| `internal/modules/records/route_target.go` | Resolves incident, type, deletion state, and row version for trusted route dispatch and access scoping. | `RouteTarget`, resolver, constructor, `Resolve`, `ResolveTx`, `ResolveIncident`, `RecordIncident`, `RecordRouteTarget`. | Workbook assembly/handlers and Timeline adapters. | Platform Postgres interface, `pgx`, UUID, direct Records query. | Workbook route/mutation/conflict tests, Timeline integration, route inventory. | Workbook, Revisions, and Timeline OpenAPI/protocol surfaces downstream. | Records envelope query facade. | high | The view-schema argument to `RecordIncident` is currently ignored and MUST NOT acquire security semantics. |
| `internal/modules/records/store.go` | Inserts, loads, batch-loads, optionally locks, and advances first-class envelope versions in caller transactions. | `Store`, `Envelope`, `InsertParams`, `ErrEnvelopeNotFound`, constructor, and five transaction methods. | Timeline, Indicators, Parties, Artifacts, Assessments, Tasks/Decisions, Evidence, Entities, and app adapters. | `pgx`, UUID, direct Records SQL, UTC normalization. | Broad indirect source-owner suites; no direct Store test. | Core 02 envelope semantics, ownership/recovery metadata, route and portability surfaces. | Records envelope facade. | high | This is the defensible core; it owns no subtype rule or peer side effect. |
| `internal/modules/records/testsupport/assertx/records.go` | Provides 19 assertions spanning envelope, history, projections, links, merge, mentions, raw preservation, and indicators. | Nineteen exported assertion helpers. | Entities and Timeline tests plus cross-domain fixtures. | Records fixtures, UUID, test framework. | Cross-owner module tests. | Test evidence and future catalog-selector changes. | Split among Records `envelopetest`, Revisions, Projections, Links, Entities, Indicators, and Timeline. | medium | Every exported symbol requires exactly one source-derived disposition row. |
| `internal/modules/records/testsupport/fixtures/records.go` | Defines incident, user, membership, view, Timeline, entity, link, assessment, indicator, interval, and mutation fixtures. | Fourteen fixture types and many builders. | Entities, Timeline, Workbook support, and application test support. | UUID and cross-domain wire/persistence shapes. | Broad unit, integration, conformance, and route suites. | Fixture and selector accounting only. | Split by semantic module; cross-domain scenario composition may move to application support. | medium | Historical Records naming does not establish ownership. |
| `internal/modules/records/testsupport/golden/records.go` | Holds deterministic cross-domain IDs, values, and link/indicator expectation types. | Exported constants/variables and two expectation types. | Entities, Indicators, Timeline, and Workbook support. | UUID and test-only values. | Cross-owner fixture and integration suites. | Fixture-sensitive evidence only; Core 05 is not activated. | Split among Links, Indicators, semantic owners, and application support. | medium | No generic Records golden package remains. |
| `internal/modules/records/testsupport/routetest/inventory.go` | Combines patch, mark-reviewed, and supersede route-control inventory. | `ControlMutations`. | Incidents HTTP conformance. | Route inventory helpers and Timeline view-schema constant. | Incidents route/conformance tests. | Workbook and Timeline OpenAPI/harness accounting. | Delete combined inventory; Workbook owns patch/supersede, Timeline owns mark-reviewed. | medium | Test accounting does not establish runtime ownership. |
| `internal/modules/records/testsupport/storetest/contracts.go` | Compares view-schema field keys and row keys. | `AllowedFieldKeys`, `SortedRowFieldKeys`. | Store/service-backed tests across modules. | Platform view-schema support. | Projections, Workbook, and source-owner tests. | View-schema and generated UI/view projections are consumed, not authored. | Neutral platform view-schema test support. | low | No envelope behavior is asserted. |
| `internal/modules/records/testsupport/storetest/harness.go` | Starts full server/runtime harnesses and checks migrations, schema, views, and route surfaces. | Harness types, start helpers, and assertion helpers. | Broad service-backed module tests. | Application support, Postgres/S3 testcontainers, migrations, HTTP, views. | Broad service-backed owner suites. | Test-support inventory and route/view/migration evidence accounting. | Startup to `internal/testutil/appsupport`; other assertions to database, view, and route owners. | high | Only application support may remain service-starting. |
| `internal/modules/records/testsupport/storetest/runtime.go` | Provides auth/TOTP, HTTP, incident/membership setup, cross-domain seed/query helpers, and response/serialization assertions. | `StoreHarness`, `LoginResult`, and numerous helpers/types. | Nearly every record-type owner, Revisions, Projections, Workbook, Timeline, and shared support. | Auth, Incidents, normalization, Postgres, HTTP, fixtures, testcontainers. | Broad unit and integration suites. | Test-support inventory and runtime/support security scans. | Application support, Auth, Incidents, and each semantic source owner. | high | Rename `StoreHarness` to the existing application-harness name during the all-caller move. |
| `internal/modules/records/testsupport/storetest/views.go` | Queries view envelopes/rows and inspects collection items and mention references. | Eight exported query/assertion helpers. | Workbook, Timeline, Entities, and collection tests. | HTTP server, view envelopes, Records runtime login type. | Cross-owner service-backed suites. | Saved-view, view-schema, collection, and wire contracts. | Workbook/view-query support, with mention helpers under Entities. | medium | Exact row/item/reference behavior remains frozen during relocation. |

Every target file is in scope. No file is excluded.

## 3. Module Boundary Diagnosis

The target is a **mixed-responsibility package** around a legitimate thin
record-envelope facade. It also contains persistence-adjacent adapters,
destructive-operation locking, intentional source-owner Reporting and
portability adapters, a misplaced delete/restore consumer contract, and an
accidental cross-domain test framework.

It is not a frontend shell/controller, grid-vendor adapter, projection engine,
saved-view implementation, transport route owner, authorization coordinator,
or WebSocket implementation. Those concerns MUST remain with their current
owners.

### Responsibility disposition

| Responsibility found | Current location | Correct owner | Keep / move / split / defer | Evidence | Normative planning result |
| --- | --- | --- | --- | --- | --- |
| Current envelope identity, incident scope, type, attribution, version, delete tuple, persistence, and standalone lookup | Records root | `module.records` | keep | Core 02 envelope plus broad callers | Keep the permanent facade at `internal/modules/records`. |
| Record-target resolution | `route_target.go` | `module.records` | keep | Workbook/Timeline ports and routes | Return only the trusted internal tuple; route owners authorize. |
| Deterministic destructive locking | `destructive_lock.go` | `module.records` | keep | Revisions/Entities callers and SQLSTATE mapping | Revisions or the initiating owner coordinates the operation. |
| Incident Bundle Records family | `incident_bundle_*` | `module.records` | keep | Core 01 and source catalog | Enforce exact v1/v2 invariants after authorization. |
| Record-envelope Reporting facts | `reportingprovider` | `module.records` | keep | Reporting provider model | Keep source adaptation only. |
| Delete/restore consumer interface and catalog | `records/deleterestore` and Revisions | `module.revisions` | move | Revisions owns destructive coordination | Rename to `DeleteRestoreSource`; source owners return concrete adapters. |
| Generic cross-owner delete/restore SQL | `TableProvider` | Each source owner | split and retire | Dynamic SQL crosses subtype boundaries | Use fixed owner-controlled SQL or SQLC; no compatibility package. |
| Cross-domain test support | `records/testsupport` | Semantic owners and application support | split | Direct symbol and caller inspection | Retain only `records/testsupport/envelopetest`. |
| History, rollback interpretation, change sets, idempotency, and consequence ordering | Revisions | `module.revisions` | keep | Core and live Revisions code | Revisions does not own current envelope storage. |
| Projection refresh, collaboration publication, and authorization | Outside Records | Existing owners | keep | Live call ordering and owners | Records MUST NOT acquire these side effects. |

### Object ownership map

This table is a planning decision pending adoption of `REQ-00-067`,
`REQ-01-649`, and `REQ-02-262`.

| Object or capability | Required owner after adoption | Rule |
| --- | --- | --- |
| `records` table | `module.records` | Owns current authoritative envelope state. |
| Indexes, checks, triggers, and constraints local to `records` | `module.records` | Ownership follows the relation on which the object is defined. |
| Foreign key defined on `records` | `module.records` | Constraint owner follows its defining table. |
| Foreign key defined on a subtype table and referencing `records` | Subtype owner | Constraint owner follows the subtype table. |
| `record_revisions` and local objects | `module.revisions` | Retained history only. |
| `change_sets` and `change_set_mutations` | `module.revisions` | Coordination and operation history. |
| Subtype/source relations | Record-type source owner | Owns source reconstruction and source-specific invariants. |
| Projection tables | Projection provider/storage owner | Derived state does not transfer envelope ownership. |
| `data/records.ndjson` | `module.records` under Incident Portability | Portable current envelope family. |
| Record-envelope Reporting facts | `module.records` provider; Reporting consumer | Records supplies facts; Reporting materializes and publishes. |
| Delete/restore coordination and consumer port | `module.revisions` | Source owners implement concrete adapters. |
| Envelope-specific test support | `module.records` | Final package is `internal/modules/records/testsupport/envelopetest`. |

Historical migration `00006_record_revision_substrate.sql` MUST NOT be
rewritten, moved, or renamed merely to correct ownership metadata. The
ownership correction requires no SQL table rename, foreign-key rewrite, new
deployable, or service split.

### Dependency and transaction rules

```text
record-type source modules ──► Records envelope facade
Revisions coordinator       ──► Records envelope facade
Revisions coordinator       ──► Revisions-owned DeleteRestoreSource
application assembly        ──► Revisions + concrete source adapters

Records ─X─► Revisions coordination
Records ─X─► route or authorization packages
Records ─X─► Collaboration publication
Records ─X─► projection refresh
Records ─X─► Reporting orchestration
```

Every Records mutating operation MUST receive a caller-supplied transaction.
Records MUST NOT begin, commit, roll back, or nest a transaction. It MUST NOT
append history, refresh projections, publish collaboration events, perform
authorization, map HTTP errors, or call an object store or network peer.

### SQL access policy

| Access shape | Post-adoption rule | Enforcement posture |
| --- | --- | --- |
| Write to `records` | MUST use a Records-owned transaction port. | Static boundary rule plus direct Records tests. |
| Standalone current-envelope lookup | MUST use the Records facade. | Revisions and peer direct SQL are removed. |
| Records-internal query | MAY use fixed Records-owned SQL or SQLC. | No runtime relation metadata. |
| Owner source/projection/reporting/portability query that necessarily joins envelope state | MAY use a machine-declared, fixed, read-only SQL or SQLC shape. | Exact owner/read-shape allowlist; no write permission. |
| Descriptor-generated or arbitrary SQL | MUST NOT be used. | Boundary check fails closed. |
| Dynamic table, column, relation, or SQL-fragment parameters | MUST NOT cross a provider interface. | Retire `TableProvider`; use concrete source adapters. |

The read-join exception does not grant relation ownership. A query qualifies
only when it necessarily combines owner-controlled source, projection,
Reporting, or portability data with envelope state. Revisions standalone reads
and all Revisions writes to `records` MUST migrate to the Records facade.

### Relation accounting

| Relation identity | Relation membership after adoption | Compatibility rule |
| --- | --- | --- |
| Planned `record-envelope` | `records` and its table-local object family; Records portability family | Allocate as a new authored machine identity; do not add a runtime alias. |
| Retained `record-revisions` | `record_revisions`, `change_sets`, `change_set_mutations`, and history references | Remove `records` from this identity without renaming history relations. |

If a machine identity is immutable, the implementation task MUST add an
authored migration crosswalk. It MUST NOT preserve an indefinite alias that
continues to imply Revisions owns the envelope.

## 4. Public Contract and Behavior Freeze Map

No public interface change is authorized by this tracker. HTTP paths, request
and response envelopes, WebSocket behavior, OpenAPI/protocol/view contracts,
authorization outcomes, storage semantics, and browser behavior remain frozen.

### Observable contract map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Envelope insert, load, batch load, locking, version, attribution, and delete state | Core 02 and Records facade | Store, lock, relation DDL, callers | Broad indirect owner suites | Direct Store and lock rows in this section | critical | Records performs no peer side effect. |
| Workbook generic patch, conflict resolution, supersede, and linked-note targeting | Workbook | Workbook routes/OpenAPI/mutation ports | Workbook route, conflict, mutation, and integration | Preserve owner rows; add Records only as collaborator | critical | Wire shapes, idempotency, versions, and errors are frozen. |
| Timeline mark-reviewed and other Timeline record routes | Timeline | Timeline OpenAPI/routes | Timeline and route-conformance suites | Preserve owner rows and target seam | high | Records owns neither admission nor mutation semantics. |
| Delete, restore, rollback, and history routes | Revisions | Revisions OpenAPI/routes/stores | Revisions route, recovery, rollback, lock, and integration | Direct Records lock plus Revisions source-port/catalog rows | critical | Preserve role checks, hidden outcomes, idempotency, revisions, and ordering. |
| `record_changed` authorization, event path, payload, ordering, replay, and publication | Collaboration and mutation owners | WS contracts and callers | Collaboration, owner, frontend, and browser suites | No Records-owned WS row; preserve collaborator evidence | critical | Records MUST NOT publish. |
| Projection refresh and rebuild | Projections with mutation owners | Projection ports and callers | Projections, Revisions, Workbook, Timeline | Preserve exact collaborator rows | critical | Records MUST NOT refresh projections. |
| Saved-view and view-schema query behavior | Saved Views, Projections, Workbook/view-query | Current test support and consumers | Workbook, Projections, Saved Views | Preserve field keys, row envelopes, item references, and collections | high | No production saved-view code exists in Records. |
| Authorization and hidden incident outcomes | Core 04 and route owners | Exact handlers/access services | Route guards, conformance, integration, browser | Preserve owner tests; add no Records authorization | critical | Records returns trusted internal data only. |
| `data/records.ndjson`, stable identity, ordering, attribution remap, v1/v2, and three invariants | Core 01 and Records source owner | Source port/catalog and Core | Two direct Records tests plus Incident Bundles | Exact positive/negative invariant row | critical | Required behavior is adopted at high level; exact row rules await adoption. |
| Reporting provider key, path prefix, source family, filter, content, and support reference | Reporting model with Records provider | Provider/materializer | Reporting boundary/export/integration | Direct deterministic Records provider row | high | Reporting owns orchestration and publication. |
| Generated OpenAPI, protocol, view, and topology surfaces | Authored owner fragments and catalogs | Generated policy and owner inputs | Drift and owner tests | Regenerate only if authored inputs change | high | Never hand-edit generated files. |
| Harness/test accounting | Testing Harness NLSpec | Verification owners, families, support inventory | Harness contract and owner explanation | Eight planned rows and exact selector/collaborator mapping | high | Phase maps are evidence accounting, not runtime architecture. |

No direct grid-adapter or UI-selector contract was found in the target.

### Current envelope operation contract

This table freezes current behavior for characterization. It is not permission
to broaden the API. Any later normalization of errors or removal of parameters
requires a separately authorized change.

| Interface | Inputs and caller ownership | Output | Current default or edge behavior | Current error and side-effect boundary |
| --- | --- | --- | --- | --- |
| `Store.InsertTx` | Caller supplies transaction and `InsertParams`. | Record UUID. | Nil `RecordID` generates a database UUID; supplied ID is preserved; `RowVersion == 0` defaults to 1; timestamps are stored in UTC. | SQL/constraint errors are wrapped; no history, projection, publication, or authorization. |
| `Store.AdvanceVersionTx` | Caller transaction, record ID, actor ID, time. | Incremented row version. | Time is normalized to UTC. | Missing rows currently surface as wrapped query-row error, not `ErrEnvelopeNotFound`. |
| `Store.LoadRowVersionTx` | Caller transaction and record ID. | Current row version. | No implicit locking. | Missing rows currently surface as wrapped query-row error. |
| `Store.LoadEnvelopeTx` | Caller transaction, record ID, `lock` flag. | One `Envelope`. | `lock=false` is the nonlocking default chosen by the caller; `lock=true` adds `FOR UPDATE`. Returned timestamps are UTC. | Missing row returns `ErrEnvelopeNotFound`. |
| `Store.LoadEnvelopesTx` | Caller transaction, list of IDs, `lock` flag. | Map keyed by record UUID. | Empty input returns a non-nil empty map; duplicates collapse; missing IDs are omitted; map iteration order is unspecified; SQL ordering is only an internal lock/read mechanic. | Query/scan errors are wrapped; it does not convert missing members to errors. |
| `RouteTargetResolver.Resolve` / `ResolveTx` | Database handle or caller transaction and record ID. | Incident ID, record type, deleted boolean, row version. | Deleted records remain resolvable to trusted internal callers. Transactional and nontransactional variants SHALL remain behaviorally equivalent. | Missing rows currently return a wrapped query-row error. Route owners authorize and map public errors. |
| `ResolveIncident`, `RecordIncident`, `RecordRouteTarget` | Record ID; `RecordIncident` also receives an ignored view-schema string. | Incident ID or route tuple. | The view-schema string has no authorization, dispatch, or filtering semantics. | Errors pass through the resolver boundary. |
| `LockDestructiveOperationRecordsNowaitTx` | Caller transaction and record IDs. | Success or typed error. | IDs are copied, sorted lexically, and deduplicated; empty input succeeds; locks use `FOR UPDATE NOWAIT`. | Missing row returns `ErrRecordEnvelopeNotFound`; SQLSTATE `55P03` returns typed lock error with the contended ID; no partial protected set is returned. |

The initial characterization MUST preserve the two distinct envelope
not-found sentinels and the wrapped `pgx.ErrNoRows` paths. A later owner
decision MAY normalize them only through an explicit interface change with
updated callers and acceptance evidence.

### Planned `DeleteRestoreSource` interface

The first migration SHALL preserve the current four method signatures, result
types, and error semantics while moving the consumer interface to Revisions.

| Responsibility | Interface behavior | Owner |
| --- | --- | --- |
| Snapshot source state | `SnapshotTx` returns the authoritative envelope/source snapshot from the supplied transaction. | Source adapter; consumed by Revisions |
| Apply or clear subtype tombstone | `UpdateSourceDeleteStateTx` applies current source-owner delete/restore behavior in the supplied transaction. | Source adapter |
| Determine default view consequence | `ViewSchemaID` returns the source-defined view-schema consequence. | Source adapter |
| Enforce source delete blocker | `ValidateDeletePreconditionsTx` returns the current typed blocker tuple. | Source adapter |

The Revisions catalog MUST fail application assembly when an admitted record
type has zero or multiple adapters, an adapter declares an unknown type, a
view-schema consequence is unknown, construction fails, or catalog ordering is
nondeterministic. An adapter MUST use only the supplied transaction and MUST
NOT commit, authorize, publish, refresh, map HTTP errors, call a network/object
store, or expose raw SQL/relation names. Future optional behavior MUST use a
separate narrow capability interface.

### Planned portable Records row

For Incident Bundle source contract major 1, every nonblank row in
`data/records.ndjson` SHALL contain exactly these members:

| Member | Required form |
| --- | --- |
| `record_id` | Canonical valid UUID string; stable and unique in the admitted family |
| `incident_id` | Canonical valid UUID string equal to the immutable import-context and manifest incident |
| `record_type` | One exact value from the closed mapping below |
| `created_at` | Canonical UTC RFC3339Nano with `Z` |
| `created_by_user_id` | Canonical UUID resolving exactly once through the admitted actor catalog |
| `updated_at` | Canonical UTC RFC3339Nano with `Z`; not earlier than `created_at` |
| `updated_by_user_id` | Canonical UUID resolving exactly once through the admitted actor catalog |
| `row_version` | Integer greater than or equal to 1 |
| `deleted_at` | `null` for active rows; otherwise canonical UTC RFC3339Nano with `Z` and `created_at <= deleted_at <= updated_at` |
| `deleted_by_user_id` | `null` exactly when `deleted_at` is `null`; otherwise canonical UUID resolving exactly once |

Unknown, missing, duplicate, or aliased members SHALL fail admission. Admission
MUST NOT repair, skip, canonicalize, or partially import an invalid row.

| `record_type` | Primary source family |
| --- | --- |
| `timeline_event` | `timeline` |
| `host` | `entities` |
| `identity` | `entities` |
| `party` | `parties` |
| `indicator` | `indicators` |
| `artifact` | `artifacts` |
| `task_request` | `tasks_decisions` |
| `decision` | `tasks_decisions` |
| `evidence` | `evidence` |
| `assessment` | `assessments` |

The application-composed subtype-presence registry MUST contain exactly one
contributor for every admitted type. Records owns the type-to-family mapping
and aggregate exactly-once comparison. Each source owner owns a concrete
presence adapter and fixed SQL or SQLC. The registry MUST NOT contain table
names, SQL fragments, or executable owner metadata. Versions 1 and 2 use this
same contract-major-1 mapping; version 1 has no lenient path, version 2 has no
additive-field tolerance, and no `allow_legacy_invalid_records` switch exists.

| Invariant | Exact planned acceptance rule |
| --- | --- |
| `records.incident_scope` | Every row incident equals the immutable import-context and manifest incident. |
| `records.envelope_legal` | Shape, identities, types, versions, timestamps, actor mappings, deletion tuple, and stable identity satisfy the rules above. |
| `records.subtype_complete` | Every envelope has exactly one compatible primary source-owner row, and no primary source-owner row binds to a missing or incompatible envelope. |

The public failure remains:

```text
error.code = incident_bundle_import_rejected
reason_code = source_family_invalid
source_family_id = records
invariant_id = one exact closed Records invariant ID
retryable = false
```

Raw rows, SQL, relation names, actor hints, and hostile member values MUST NOT
appear in the public error. Validation MUST occur in the one final import
transaction; failure leaves no partially visible incident.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Declared portability invariants are not visibly enforced. | Port declares three invariants; current validation does not prove legal tuples or subtype completeness. | Nonconforming bundles may be admitted. | `must_fix` | Records under Core 01 | Adopt exact rules, authorize runtime correction, implement shared parser/registry, and add positive/negative evidence. |
| Cross-domain test support is registered under Records. | Eight files, many exported symbols, broad callers, and service startup. | Misleading ownership and fragile selector/security accounting. | `must_fix` | Semantic owners plus application support | Generate the mandatory matrix; move each helper family with callers and accounting. |
| Envelope and provider behavior lacks direct owner rows. | Current Records owner selects two portability tests only. | Structural moves may change defaults, SQL, locks, or errors undetected. | `should_fix` | Records and Revisions with exact collaborators | Author the eight rows in Section 7 before movement. |
| Provider and SQL boundaries are not fully machine-enforced. | Reporting has a guard; other imports/read shapes remain conventional or broad. | Facades may regrow or generic SQL may cross owners. | `should_fix` | Records, Revisions, source owners, coordinators | Add fixed import, write, and read-shape rules before moving implementation. |
| Logical and physical envelope ownership differ in projections. | Portability says Records; schema/recovery say Revisions. | Recovery, SQL, and dependency accounting disagree. | `defer` | Core 00/01/02 adoption, then Records/Revisions projections | Decision is complete; implementation remains blocked until adoption. |
| Delete/restore interface is under the wrong namespace. | Revisions consumes it and owns coordination; source owners construct behavior. | Current location implies false Records ownership and enables dynamic cross-owner SQL. | `defer` | Revisions consumer port and concrete source owners | Decision is complete; migrate after characterization and adoption. |
| Fixed SQL is hidden behind narrow Records/source adapters. | Store, resolver, lock, portability, and Reporting own bounded shapes. | Moving SQL could alter transactions or ordering. | `intentional/no_action` | Records and source-owner adapters | Retain fixed SQL where admitted by Section 3; forbid arbitrary SQL. |
| Authorization, projection, and collaboration remain outside Records. | Route owners authorize; callers coordinate projection/events after envelope work. | Moving them inward would hide peer side effects and transport policy. | `intentional/no_action` | Existing owners | Preserve side-effect-free Records boundary. |
| Reporting and Incident Bundle providers are owner contributions. | Coordinators collect source-owner providers. | Removing them would force peer-owned source queries. | `intentional/no_action` | Records provider subpackages | Retain, characterize, and boundary-guard. |
| No frontend, grid, or production saved-view responsibility exists here. | Target source/import inspection. | Invented movement would expand scope without evidence. | `intentional/no_action` | Existing frontend/view owners | Exclude from implementation except conditional regression verification. |
| Generated outputs are downstream projections. | Generated artifact policy and Testing Harness NLSpec. | Hand edits create drift and false authority. | `intentional/no_action` | Authored owner inputs and Make generators | Use `make generate`; never hand-edit generated roots. |

The two `defer` rows defer execution, not architectural choice. Their decisions
are fixed by this plan and await owner adoption or migration prerequisites.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Fix authority, scope, baseline, and write constraints. | Tracker only in this task. | Status and exact source reads. | Twelve-section tracker is current and no product file changed. |
| WF-01 | Authority adoption | chain | WF-00 | WF-02, WF-03, WF-04 | Adopt the ownership, interface, portability, and acceptance decisions before code movement. | Core 00/01/02/04 in a later authorized task. | Owner review; downstream drift only after adoption. | Proposed requirement and AC text is adopted without contradiction. |
| WF-02 | Contract-owner and verification mapping | parallel | WF-01 | WF-05 | Author exact verification identities, immutable rows, collaborators, and relation crosswalk. | Verification owners, test families, relation/ownership inputs. | Owner explanation and JSON/harness gates. | Every contract has one owner and exact evidence row. |
| WF-03 | Characterization test implementation | parallel | WF-01 | WF-05 | Prove current envelope, resolver, lock, provider, and Reporting behavior before movement. | Records/Revisions tests and exact collaborator selectors. | Focused owner and service-backed slices. | Section 7 selector scopes pass and current errors/defaults are frozen. |
| WF-04 | Boundary and source-matrix implementation | parallel | WF-01 | WF-05 | Add machine SQL/import boundaries and derive every test-support symbol/caller disposition. | Backend boundaries and test-support source/callers. | Static boundary and matrix completeness review. | Boundary rules fail closed and each exported helper appears once. |
| WF-05 | Ownership and facade migration | chain | WF-02, WF-03, WF-04 | WF-06 | Align relation metadata and route all writes/standalone lookups through Records. | Records/Revisions, assembly, ownership/recovery/boundary inputs. | Focused owners, migration drift, static boundary, server build. | Current envelope has one owner and Revisions has no direct standalone/write SQL. |
| WF-06 | Delete/restore and test-support migration | chain | WF-05 | WF-07 | Install Revisions consumer port, concrete adapters, and semantic test-support roots. | Revisions, source owners, app assembly, test support/inventory. | Owner slices, harness, static boundary. | Generic provider and Records catch-all are absent; no compatibility package remains. |
| WF-07 | Strict portability and projection regeneration | chain | WF-06 | WF-08 | Enforce all Records invariants and update authored projections/accounting. | Records portability, source adapters, owner inputs, generated outputs via Make. | Incident Bundle/Records slices, generation, JSON, harness gates. | Legal v1/v2 pass; every negative invariant fails safely and rolls back. |
| WF-08 | Validation and final handoff | chain | WF-02, WF-03, WF-04, WF-05, WF-06, WF-07 | none | Run focused-to-broad verification and leave restartable evidence. | All authorized changes and tracker. | Finalization, affected owners, drift, build, broad checks. | Passing roots or exact failures and all handoffs are recorded. |

WF-02, WF-03, and WF-04 MAY run in parallel only after WF-01 adoption. WF-05
MUST NOT begin until all three complete. There is no frontend implementation
workstream.

## 7. Proposed Refactor Slice Plan

### Planned verification identities

If an equivalent active row exists, its immutable ID MUST be retained and its
selector, verification mapping, or collaborators updated instead of creating a
duplicate.

| Owner | Verification identity | Semantic row ID | Evidence and exact selector scope |
| --- | --- | --- | --- |
| Records | `module.records.verification.envelope_store` | `module.records.envelope.store` | Integration: insert defaults/supplied ID; single load/not-found; batch dedupe/missing/unordered-map behavior; version advance; optional row locking. |
| Records | `module.records.verification.route_target` | `module.records.envelope.route_target` | Integration: `Resolve`/`ResolveTx` parity; found/not-found; deleted/version tuple; ignored view-schema has no auth semantics. |
| Records | `module.records.verification.destructive_lock` | `module.records.envelope.destructive_lock` | Integration: sort/dedupe; missing envelope; `55P03` contention; caller transaction; no partial set. |
| Records | `module.records.verification.reporting_envelope_facts` | `module.records.reporting.envelope_facts` | Integration: deterministic facts; deletion filter; key/family/path/value/support reference; caller snapshot. |
| Records | `module.records.verification.portability_envelope_invariants` | `module.records.portability.envelope_invariants` | Integration: legal v1/v2; one negative per invariant; missing/duplicate subtype; safe failure and rollback. |
| Records | `module.records.verification.architecture_boundary` | `module.records.architecture.boundary` | Static: Records owns writes/standalone lookups; admitted fixed read joins only; no route/auth/projection/collaboration/Reporting-orchestration imports. |
| Revisions | `module.revisions.verification.delete_restore_source_port` | `module.revisions.delete_restore.source_port_catalog` | Unit: exactly one adapter per admitted type; missing/duplicate/unknown/nondeterministic catalogs fail closed. |
| Revisions | `module.revisions.verification.delete_restore_source_adapter_matrix` | `module.revisions.delete_restore.source_adapter_matrix` | Integration: every source snapshot, blocker, tombstone/restore, view result, typed failure, and caller transaction. |

Existing Workbook, Timeline, Revisions, Entities, Reporting, Incident Bundles,
and only the exact affected Projections/Collaboration rows SHALL add
`module.records` as collaborator. Passing a broad owner slice is not a
substitute for these direct rows.

### Mandatory test-support source matrix

Before a helper moves, a source-derived artifact MUST contain one row per
exported symbol with exactly these columns:

```text
current_package
symbol
symbol_kind
direct_callers
transitive_shared_callers
asserted_postcondition
normative_owner
destination_package
disposition = move | split | inline | delete
shared_or_owner_local
service_starting
runtime_security_scan
support_security_scan
affected_catalog_row_ids
move_slice
compatibility_alias_allowed = false
```

Every symbol and caller MUST appear exactly once. A symbol spanning multiple
postconditions MUST split. Only application support may be service-starting.
Every new support root MUST be registered exactly once and remain in the
applicable security scan. A helper family, callers, selector updates, and
inventory changes MUST move in the same slice.

### Slice plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | WF-00 | Adopt proposed Core ownership, interface, portable-row, registry, and acceptance text. No production move. | Core 00/01/02/04 in a later authorized task. | Conflicting owner text or premature projection authority. | Owner review only. | Applicable Markdown/owner checks discovered in that task. | Revert the coherent owner-text change if adopted sections contradict. | `REQ-00-067`, `REQ-01-649..651`, `REQ-02-262`, and `AC-509..516` are adopted and internally consistent. |
| SL-01 | SL-00 | Author the eight direct verification rows and characterize current behavior without changing production behavior. | Records/Revisions tests, verification owners, families, exact collaborators. | Defaults, errors, locks, transaction ownership, Reporting shape. | All selector scopes in the verification table; preserve the existing two portability rows. | `make test-slice OWNER=module.records`; `make service-backed-test-slice OWNER=module.records`; affected owner slices. | Revert each test row and authored accounting together. | Every production Records surface has direct or exact collaborator evidence. |
| SL-02 | SL-01 | Reassign `records` ownership and planned relation identity; migrate all writes and standalone Revisions envelope lookups to the Records port. | Records, Revisions, app assembly, schema/recovery/source/boundary inputs. | Transaction scope, row versions, delete tuple, recovery and portability identity. | Direct Records rows plus Revisions/Workbook/Timeline/source-owner collaborators. | Focused owners; `make backend-module-boundary-check`; `make migration-drift`; `make build-server`. | Move one coherent caller family with its boundary entries; historical migrations remain untouched. | `record-envelope` owns `records`; Revisions retains history identity; no direct Revisions write/standalone lookup remains. |
| SL-03 | SL-01, SL-02 | Define Revisions-owned `DeleteRestoreSource`; move all consumers; construct concrete source adapters in app assembly; remove `TableProvider`. | Revisions, source owners, app assembly, Records deleterestore package. | Snapshot shape, blocker semantics, subtype delete state, default view, errors. | Revisions catalog and adapter-matrix rows plus source-owner rows. | Affected unit/service slices; static boundary; server build. | Move interface, implementations, assembly, and callers as one coherent slice. | Interface/catalog are Revisions-owned, every type resolves once, dynamic provider SQL and compatibility package are absent. |
| SL-04 | SL-01 | Generate the symbol/caller matrix and split test support by semantic owner; retain only `records/testsupport/envelopetest`. | Records test support, semantic owner support, `internal/testutil/appsupport`, inventories and selectors. | Fixture identity, auth, startup, queries, views, route inventory, selection count. | Preserve every selector and behavior; add no product test semantics. | Affected owner/service slices; `make harness-contract`; `make backend-module-boundary-check`. | Move one semantic family with callers and accounting; revert that family on failure. | Matrix is exhaustive, every helper has one owner, only app support starts services, and no compatibility root remains. |
| SL-05 | SL-00, SL-01, SL-02 | Implement exact v1/v2 parser, legal tuple validation, actor resolution, and application-composed subtype-presence registry. **Requires later authorization.** | Records portability/source port, source-owner adapters, Incident Bundle assembly, fixtures and tests. | Runtime now rejects malformed inputs previously admitted; safe error and rollback must not drift. | Legal v1/v2, each invalid member/type/time/actor/delete case, every subtype mapping, each invariant failure, rollback. | Records and Incident Bundles focused/service-backed slices. | Revert parser, adapters, registry, and acceptance fixtures together; do not add lenient fallback. | Both versions enforce the same exact contract; failures use one closed invariant ID and leave no state. |
| SL-06 | SL-02, SL-03, SL-04, SL-05 | Update authored verification, family, support, ownership, recovery, source, and boundary inputs; regenerate downstream artifacts only through Make. | Authored contracts/tools inputs and generated roots through generators. | Duplicate/missing rows, relation drift, support drift, hand-edited output. | Catalog, shape, boundary, harness, and generation policy tests. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make harness-contract`. | Revert authored input and its generated projections together. | Every test selects once, projections agree with owners, and generated outputs match authored inputs. |
| SL-07 | SL-01 through SL-06 | Run final focused-to-broad verification and append implementation handoffs. | Whole affected tree and tracker. | False completion from partial or stale evidence. | All new and preserved exact rows. | `make agent-finalize`; affected owner slices; boundary/drift/build gates; risk-appropriate `make test-fast` and `make check`; browser targets only if a public route/view/WS surface changed. | Revert only the first failing implementation slice and retain failure artifacts. | Required checks pass or exact failures/run roots are recorded; all binary criteria and handoffs are current. |

SL-00 through SL-07 require later authorization. SL-05 changes runtime
admission and therefore requires explicit implementation authorization even
though Core 01 already requires fail-closed invariants.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| tracker documentation | `make lint-markdown` | This tracker and authored Markdown | yes | Passed for this revision at `.cartulary/test-results/20260730T003022Z-p1451107`; no product validation is implied. |
| unit | `make test-slice OWNER=module.records` | Active Records owner rows | yes | Initial baseline passed with two selected tests at `.cartulary/test-results/20260729T225732Z-p1421309`; rerun after new rows. |
| integration | `make service-backed-test-slice OWNER=module.records` | Service-backed Records rows | yes | Initial baseline passed with one selected test at `.cartulary/test-results/20260729T225744Z-p1423135`. |
| affected owners | Make-owned owner slices discovered by `make task-guide` | Revisions, Reporting, Incident Bundles, Workbook, Timeline, Entities, and exact collaborators | yes | Select the narrowest affected rows; do not substitute broad passing suites for direct rows. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Routes, views, history, restore, collaboration | no | Required only if an implementation changes a public route, view, or WebSocket surface; not run for this tracker. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Authored projections and generated policy | no | Required after authored contract, verification, harness, or generated inputs change. |
| migration drift | `make migration-drift` | Append-only migration and schema projection posture | no | Required for ownership/schema work even though historical migrations remain unchanged. |
| import-boundary/static | `make backend-module-boundary-check` | Imports, SQL writes/lookups, provider and read-shape rules | yes | Initial baseline passed at `.cartulary/test-results/20260729T225732Z-p1421389`; rerun after boundary changes. |
| harness accounting | `make harness-contract` | Owner/family rows, support roots, execution topology | no | Required after verification or test-support changes. |
| build | `make build-server` | Production composition after port/assembly moves | no | Required after SL-02 or SL-03. |
| full implementation gate | `make agent-finalize`; risk-appropriate `make test-fast`; `make check` | Final affected repository | no | Not run in this planning-only session. |

Canonical command discovery used `make help-all`,
`make task-guide ROLE=module-author OWNER=module.records`, and
`make explain-test-owner OWNER=module.records`, plus affected-owner task
guides. Raw `go`, `pnpm`, Vitest, and Playwright commands are not canonical
evidence.

No browser, generation-drift, migration-drift, harness-contract, build,
`agent-finalize`, `test-fast`, or `check` result is claimed for this tracker
revision.

## 9. Top-Level Work Tracker

Only `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED` are
valid statuses in this table.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Fix target, authority, planning-only posture, and NLSpec writing rules. | WF-00 | DONE | none | Section 1 | Scope and authority are explicit; only the tracker changed. |
| T-002 | Inventory all 16 target files and their current surfaces. | WF-00 | DONE | T-001 | Section 2 | Every file has responsibility, callers, dependencies, tests, contracts, owner, and risk. |
| T-003 | Decide logical/physical ownership, SQL policy, relation identities, and transaction direction. | WF-00 | DONE | T-002 | Section 3; RB-001 | One decision exists for every object and access shape. |
| T-004 | Decide strict portability compatibility and exact contract-major-1 rules. | WF-00 | DONE | T-002 | Section 4; RB-002 | Shape, mappings, defaults, invariants, versions, and failure envelope are complete. |
| T-005 | Decide direct verification identities, rows, collaborators, and selector scopes. | WF-00 | DONE | T-002 | Section 7; RB-003 | Eight rows and collaborator updates are named without duplicate evidence. |
| T-006 | Decide semantic test-support destinations and matrix gate. | WF-00 | DONE | T-002 | Sections 2 and 7; RB-004 | Every file family has a destination and the matrix has binary completeness rules. |
| T-007 | Decide delete/restore consumer ownership and migration posture. | WF-00 | DONE | T-002 | Sections 3 and 4; RB-005 | Revisions port, concrete adapters, assembly join, and no-alias rule are explicit. |
| T-008 | Adopt proposed Core requirements and acceptance criteria. | WF-01 | BLOCKED | T-003 through T-007 | SL-00 | Later authorization results in adopted, internally consistent owner text. |
| T-009 | Author direct characterization and verification accounting. | WF-02, WF-03 | TODO | T-008 | SL-01 | All eight direct rows select and pass with exact collaborators. |
| T-010 | Author static SQL/import/read-shape boundaries and source matrix. | WF-04 | TODO | T-008 | SL-01, SL-04 | Boundaries fail closed and every helper symbol/caller appears once. |
| T-011 | Align relation ownership and narrow the Records facade. | WF-05 | BLOCKED | T-009, T-010 | SL-02 | `records` has one owner and all writes/standalone lookups use Records. |
| T-012 | Migrate delete/restore port and concrete adapters. | WF-06 | BLOCKED | T-011 | SL-03 | Revisions owns the port/catalog and generic provider SQL is absent. |
| T-013 | Split Records test support by semantic owner. | WF-06 | TODO | T-009, T-010 | SL-04 | Only envelope-specific support remains and accounting is exact. |
| T-014 | Implement strict v1/v2 Records portability validation. | WF-07 | BLOCKED | T-008, T-009, T-011 | SL-05 | Authorized exact validation passes every positive/negative case and rolls back failures. |
| T-015 | Update authored projections/accounting and regenerate through Make. | WF-07 | TODO | T-011 through T-014 | SL-06 | Ownership, recovery, source, boundary, verification, and support projections agree. |
| T-016 | Execute final validation and implementation handoff. | WF-08 | TODO | T-009 through T-015 | SL-07 | Required commands and retained roots are recorded; all implementation criteria pass. |

Planning decisions are complete. Adoption and implementation are not complete.

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Planning scope, authority, repository baseline, and safe label are fixed. | Inspected framework, domain, Core 00-04, Reporting/Testing Harness/Extension NLSpecs, current handoffs; touched this tracker only. | `git status --short --branch`; `git rev-parse HEAD`; exact `sed`/`rg` reads. | Target exists with 16 files; tracker was absent; no adopted-owner contradiction found. | RB-001 through RB-005 affect later implementation, not tracker creation. | Authorize SL-00 characterization before structural movement. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | Planning decisions are complete; proposed owner additions remain adoption-pending. | Inspected NLSpec doctrine, analysis notes, current tracker, exact Records store/resolver/lock/provider source, owners and projections; touched this tracker only. | Exact `sed`/`rg` reads; repository status/diff; `make lint-markdown`. | Tracker rewritten with normative planning gates, explicit defaults, mappings, and binary acceptance. Markdown lint passed at `.cartulary/test-results/20260730T003022Z-p1451107`. | Owner adoption and implementation authorization remain. | Begin SL-00 only in a later owner-document task. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Legitimate envelope kernel and source-owner providers are mixed with broad test support; no production refactor performed. | Inspected all `internal/modules/records/**`, exact app assembly, Workbook/Revisions callers, Reporting and portability consumers; touched this tracker only. | `find internal/modules/records`; import/symbol `rg`; exact source reads; `make backend-module-boundary-check`. | All files inventoried; boundary baseline passed at `.cartulary/test-results/20260729T225732Z-p1421389`. | RB-001 and RB-005 block permanent placement decisions. | Add SL-00 direct characterization, then resolve owner decisions. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | Records is the planned logical and physical envelope owner; Revisions owns history and destructive coordination. | Inspected current envelope SQL, Revisions relation use, schema/recovery/source/boundary inputs; touched this tracker only. | Exact source and projection reads; targeted `rg`. | Object, SQL access, relation identity, dependency, and transaction rules are decision-complete. | Adoption of proposed Core requirements blocks movement. | Adopt owner text, then characterize before changing relation metadata or callers. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | No frontend shell, controller, grid-vendor, or UI-selector implementation exists in the target. | Inspected backend target/callers and searched `apps/web`, packages, OpenAPI, and WS surfaces; touched this tracker only. | Targeted `rg` for record routes, `record_changed`, view-schema, selectors, and grid dependencies. | Frontend/browser behavior is a downstream regression surface, not a Records ownership candidate. | None for planning; browser scope depends on later route/view/WS changes. | Run owner browser targets only when an authorized slice affects those contracts. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | The absence of frontend ownership is retained as an explicit boundary. | Reused direct target/import evidence and current contract map; touched this tracker only. | Tracker/source review. | No frontend workstream or ownership movement is planned; browser checks remain conditional. | None for planning. | Preserve route/view/WS contracts during backend work. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Authored Records portability, OpenAPI owners, schema/recovery ownership, generated policy, and provider boundaries are mapped. | Inspected source catalog, three OpenAPI owner fragments, recovery catalog, schema/boundary/generated policies, and verification inputs; touched this tracker only. | Targeted `rg` and exact JSON reads; `make help-all`. | Logical `module.records` portability ownership differs from physical Revisions relation ownership; no generated file was touched. | RB-001 and RB-002. | Resolve ownership and behavior authorization before changing authored inputs; regenerate only with `make generate`. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | Proposed owner requirements, relation identities, portable row, subtype mapping, invariant rules, and safe failure are exact but non-authoritative until adoption. | Inspected Core owner sections, source/recovery/schema/boundary projections, portability source, and NLSpec doctrine; touched this tracker only. | Exact reads and identifier searches. | Planned IDs `REQ-00-067`, `REQ-01-649..651`, `REQ-02-262`, and `AC-509..516` were unused; no contract or generated file changed. | Owner adoption and later runtime authorization. | Adopt once in Core owners, then update projections and regenerate through Make. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Current Records owner evidence covers only portability duplicate identity and apply/rollback; test support is broadly cross-owner. | Inspected Records tests, owner verification/family inputs, test-support inventory, importing tests, and Testing Harness NLSpec; touched this tracker only. | `make task-guide ROLE=module-author OWNER=module.records`; `make explain-test-owner OWNER=module.records`; task guides for affected owners; Records unit/service slices; `make lint-markdown`. | Unit passed with two tests at `.cartulary/test-results/20260729T225732Z-p1421309`; service-backed passed with one test at `.cartulary/test-results/20260729T225744Z-p1423135`; Markdown passed at `.cartulary/test-results/20260729T231913Z-p1430019`. | RB-003 and RB-004. | Author exact SL-00 characterization and helper disposition before moving code. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | Eight direct verification rows, exact collaborator posture, semantic test-support destinations, and the source-matrix schema are fixed. | Inspected current tests, verification/family inputs, all eight support files and caller evidence; touched this tracker only. | Exact reads/searches; `make lint-markdown`. | Planning gaps are closed; no product, harness, owner-slice, or generated validation was run in this session. Markdown lint passed. | Authoring waits for owner adoption and implementation authorization. | Implement SL-01 and produce the matrix before any code or helper move. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Records resolves envelope targets but does not authenticate, authorize, publish WS events, or refresh projections. | Inspected Workbook and Revisions route handlers, Core 04, destructive locking, and downstream collaboration searches; touched this tracker only. | Exact source reads and targeted `rg` for route roles, lock errors, and `record_changed`. | Authorization remains correctly route-owned; destructive lock outcomes and hidden incident behavior are freeze surfaces. | Direct Records lock/target characterization is missing under RB-003. | Preserve route-owner checks and add seam-level characterization without moving security policy inward. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | Security separation is specified as a negative interface contract. | Inspected resolver, lock behavior, route-owner evidence, and Core 04 acceptance sequence; touched this tracker only. | Exact source reads and tracker review. | Records returns typed internal outcomes only; route owners retain authorization, hidden outcomes, and public error mapping. | Direct characterization and AC adoption remain. | Adopt AC-511/512, then prove the seam without moving policy inward. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-29T19:13:53-04:00 | Codex / initial Records planning | Inventory, diagnosis, contract map, workflows, slices, validation, and blockers are ready for handoff. | Inspected all evidence named in Sections 1-8; touched this tracker only. | Repository searches, owner discovery, three baseline product/static validation targets, and `make lint-markdown`. | Planning criteria and Markdown validation are complete; no browser, generated-drift, harness-contract, `test-fast`, or `check` success is claimed. | RB-001 through RB-005. | Next authorized session starts with SL-00; do not rename or move the root facade first. |
| 2026-07-29T20:24:27-04:00 | Codex / NLSpec tracker hardening | No architectural question remains open; five former blockers are decisions with adoption, authoring, matrix, or migration gates. | Inspected all sources named in Sections 1-8; touched this tracker only. | Repository searches, exact source reads, status/diff, and `make lint-markdown`. | Tracker is decision-complete; the refactor and owner adoption are not implemented. | RB-001 through RB-005 retain the exact pending states in Section 11. | Next authorized session adopts SL-00; no production move precedes adoption and characterization. |

## 11. Open Questions and Blockers

No architectural question remains open. The stable IDs are retained because
they identify gates that a later session must close. A planning decision does
not satisfy its authority or implementation gate.

| ID | Decision and remaining gate | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Records owns the logical envelope and physical `records` relation; Revisions owns history and coordination. Adopt the decision before movement. | Ownership controls SQL, recovery, portability, relation identity, and dependency direction. | Adopt `REQ-00-067`, `REQ-01-649`, and `REQ-02-262`; align projections in SL-02. | `DECIDED—ADOPTION PENDING` |
| RB-002 | Both admitted bundle versions reject illegal envelopes and incomplete subtype bindings; there is no lenient compatibility mode. Authorize the runtime correction. | Current behavior may admit nonconforming bundles; failure must remain safe and transactional. | Adopt exact `REQ-01-651` details and authorize SL-05 implementation/evidence. | `DECIDED—IMPLEMENTATION PENDING` |
| RB-003 | Five direct Records behavior rows, two Revisions rows, one Records static row, and exact collaborator updates are defined. | High-risk SQL, default, lock, error, and provider behavior must be directly evidenced. | Author the exact identities/selectors in Section 7; preserve equivalent immutable rows. | `READY FOR AUTHORING` |
| RB-004 | Cross-domain Records test support splits by semantic owner; only `envelopetest` remains. | Helper movement can alter fixtures, startup, security scans, and selector accounting. | Generate the mandatory exported-symbol/caller matrix before SL-04. | `DECIDED—SOURCE MATRIX REQUIRED` |
| RB-005 | Revisions owns `DeleteRestoreSource`; source owners return concrete adapters; app assembly joins them; `TableProvider` retires. | The current namespace and dynamic SQL imply false ownership and unsafe coupling. | Characterize current semantics, adopt `REQ-01-650`, then execute SL-03 without a compatibility package. | `DECIDED—MIGRATION PENDING` |

## 12. Binary Completion Criteria

### Tracker acceptance

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/records` is inventoried. | PASS | Section 2 contains exactly 16 file rows and excludes none. |
| Every discovered contract has one current owner and test posture. | PASS | Section 4 covers envelope, routes, WebSocket, revisions, projections, views, authorization, portability, Reporting, generated surfaces, and harness accounting. |
| Every planned object, relation identity, dependency, transaction, and SQL access shape has one disposition. | PASS | Section 3 mapping tables are exhaustive for discovered Records concerns. |
| Interface inputs, outputs, errors, side effects, defaults, and edge behavior are explicit. | PASS | Section 4 defines current envelope operations, `DeleteRestoreSource`, subtype registry, portable row, invariant, and safe failure contracts. |
| Every mapping is tabular and closed where incompatible implementations would be observable. | PASS | Object, SQL, relation, record type, requirement, verification, blocker, and AC mappings are explicit. |
| Every workflow and slice has dependencies, rollback posture, validation, and a binary exit condition. | PASS | Sections 6 and 7 define WF-00 through WF-08 and SL-00 through SL-07. |
| Behavior changes are identified and authorization-gated. | PASS | SL-05 and RB-002 distinguish adopted fail-closed intent from pending runtime authorization. |
| Validation commands are repository-owned and unrun checks are not claimed. | PASS | Section 8 uses Make targets and lists every skipped layer. |
| Owner contradictions use the required marker. | PASS | No contradiction was found; Section 1 mandates `BLOCKED: owner contradiction` if one appears. |
| Repository/framework mismatches and their dispositions are explicit. | PASS | Section 1 records five mismatches without treating evidence as authority. |
| Prior handoff history is preserved and a new session is appended to all seven tables. | PASS | Section 10 contains both initial and NLSpec-hardening rows. |
| No normative product requirement is asserted only by this tracker. | PASS | Proposed IDs are adoption-pending and owner authority is stated in Section 1. |

### Planned Core acceptance mapping

These criteria remain proposed until Core 04 adopts them.

| Proposed AC | Requirement trace | Verification identity | Binary acceptance |
| --- | --- | --- | --- |
| `AC-509` | `REQ-00-067`, `REQ-01-649`, `REQ-02-262` | `module.records.verification.architecture_boundary` | `records` and its local objects have exactly one Records owner; history relations have exactly one Revisions owner; recovery, portability, and relation identities agree. |
| `AC-510` | `REQ-01-649`, `REQ-02-262` | `module.records.verification.envelope_store` | Direct evidence proves ID/default/version/attribution/time/load/batch/lock behavior and absence of history, authorization, projection, or publication side effects. |
| `AC-511` | `REQ-01-649` | `module.records.verification.route_target` | Resolver variants return only the exact target tuple; deleted rows remain internally resolvable; route owners retain authorization and hidden outcomes; view schema is not a predicate. |
| `AC-512` | `REQ-01-649` | `module.records.verification.destructive_lock` | IDs sort/dedupe; missing and `55P03` contention are distinct typed outcomes; caller owns transaction; no partial protected set is returned. |
| `AC-513` | `REQ-01-651` | `module.records.verification.portability_envelope_invariants` | Legal v1/v2 import; one negative per exact invariant; safe error is exact; failure leaves no state. |
| `AC-514` | `REQ-01-650` | Revisions delete/restore port and adapter-matrix verifications | Revisions owns interface/catalog; source owners construct concrete adapters; assembly is the join point; every type resolves once; adapters perform no peer orchestration. |
| `AC-515` | `REQ-01-649` and Reporting owner contract | `module.records.verification.reporting_envelope_facts` | Facts are deterministic, caller-transaction-bound, source-only, and exclude soft-deleted envelopes; Reporting retains materialization/publication. |
| `AC-516` | Testing Harness owner requirements | Exact affected owner rows and `make harness-contract` | Every helper has one semantic owner; every support root is registered once; startup/security classifications are exact; no Records compatibility root remains. |

The tracker is planning-complete. Owner adoption, characterization,
implementation, projection updates, generation, and implementation validation
are not complete. No production refactor was performed.
