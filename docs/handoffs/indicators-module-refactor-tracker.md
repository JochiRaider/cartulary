# indicators Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/indicators`
- **Target label:** `indicators` (validated lowercase kebab case; no spaces, path separators, shell metacharacters, or unsafe filename characters)
- **Output path:** `docs/handoffs/indicators-module-refactor-tracker.md`
- **Repository baseline:** branch `main`, commit `67d2ce2c98ce9f124c87ca6d980ec9321b4d2ef7`, observed at `2026-08-03T08:32:42-04:00`
- **Status:** implementation authorized and active
- **Allowed change:** the specification, authored contracts, implementation, tests, migrations, harness inputs, generated projections, and handoff evidence required by SL-00 through SL-08
- **Non-goals:** unrelated frontend timestamp inference, draft Reference Pack ownership changes, and the historically Timeline-named runtime projection bundle
- **Implementation authorization:** the user authorized the complete remediation plan on `2026-08-03`, including IND-016 owner closure and SL-00 through SL-08

### Active execution protocol

- Execute `tracker bootstrap -> SL-00 -> IND-015 -> IND-016 -> SL-02 -> SL-01 -> SL-03 -> SL-04 -> SL-05 -> SL-06 -> SL-07 -> SL-08`.
- Treat every named workstream as a separately reviewable and reversible commit.
- After validating a workstream, update this tracker with its status, commands, run roots, risks, and next action; commit that checkpoint before starting the next workstream.
- Mark a workstream `BLOCKED` when an entry or exit gate cannot be met, and do not begin dependent work.
- Update authored inputs and invoke Make-owned generators; never hand-edit generated roots or lockfiles.

### Implementation baseline

- Starting branch and commit: `main` at `67d2ce2c98ce9f124c87ca6d980ec9321b4d2ef7`.
- Starting tracked change: this newly added tracker only; no pre-existing production, contract, test, or generated-artifact changes.
- `make backend-module-boundary-check` passed at `.cartulary/test-results/20260803T135505Z-p3870704`.
- `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260803T135507Z-p3871025`.
- `make json-shape-check` passed at `.cartulary/test-results/20260803T135508Z-p3871422`.
- `make service-backed-test-slice OWNER=module.indicators` passed at `.cartulary/test-results/20260803T135514Z-p3871882`.

The authority order used for this plan is:

1. adopted subsystem NLSpecs within their named scope;
2. Core 00 through Core 04 for implementation-conformance behavior;
3. Core 05 only for claim-bearing timed or fixture-sensitive publication;
4. domain vocabulary and implementation-support guides;
5. current repository code and tests;
6. prior plans, handoffs, and the planning framework as evidence only.

Core 05 is not applicable to this planning artifact because this task publishes no timed, benchmark, or fixture-sensitive claim. No owner contradiction was found. The draft Reference Pack NLSpec is evidence only and does not supersede the current Core 02 assignment of canonical indicator behavior to the Indicators owner.

### Document conventions

This tracker uses four statement classes. A later agent MUST determine the class before treating any sentence as an implementation instruction.

| Class | Meaning | Normative effect |
| --- | --- | --- |
| Current-state finding | A fact observed in the repository baseline. | Descriptive evidence only; it MUST be rechecked if the baseline changes. |
| Adopted-owner requirement | A requirement cited to an adopted Core or subsystem NLSpec. | Binding within that owner's scope. `MUST`, `MUST NOT`, and `MAY` retain their ordinary normative meanings. |
| Authorized implementation requirement | A slice-local requirement that translates adopted behavior into implementation and acceptance gates. | Binding only in a later repository-change task whose scope includes that slice. |
| Proposed owner language | Text that an owner document still needs to adopt. | Non-authoritative until adoption. Code, tests, contracts, and this tracker MUST NOT treat it as current product behavior. |

Unless a requirement explicitly states otherwise, the default is behavior preservation: public routes, envelopes, errors, events, view and saved-view identity, projections, authorization, generated contracts, and frontend selectors MUST remain unchanged. An omitted mechanism is intentionally left to the implementer only when two conforming implementations remain interchangeable at every affected boundary. `TODO` identifies missing evidence; `BLOCKED` identifies a gate that forbids implementation. Requirements are defined once in Sections 4 and 7 and referenced by stable ID elsewhere.

### Owner documents inspected

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` (read first; template and doctrine only)
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/network-flow-activity-nlspec.md` (adopted/current)
- `docs/graph_projection_nlspec.md` (adopted/current; explicitly not the workbook-grid projection owner)
- `docs/testing-harness-nlspec.md` (adopted/current)
- `docs/reference-pack-subsystem-nlspec.md` (draft; non-authoritative evidence only)

### Supporting planning material inspected

- `temp/analysis-notes.md` (gap analysis and proposed closure language; evidence only)
- `docs/research/nlspec-spec.md` (writing and evaluation guidance; non-authoritative)

### Repository files inspected

- Every file under `internal/modules/indicators`, inventoried individually in Section 2.
- Composition and caller paths: `internal/app/workbookassembly/catalog.go`, `internal/app/projectionassembly/catalog.go`, `internal/app/importassembly/owner_registry.go`, `internal/app/incidentportabilityassembly/catalog.go`, `internal/app/recoveryassembly/state_catalog.go`, `internal/app/revisionassembly/revisions.go`, and `internal/app/server/runtime.go`.
- Workbook and Network Flow adapters: `internal/modules/workbook/mutation_contributions.go`, `internal/modules/workbook/mutation_api.go`, `internal/modules/workbook/routes.go`, `internal/modules/networkflow/ports.go`, `internal/modules/networkflow/transaction_participants.go`, `internal/modules/networkflow/indicator_link.go`, and `internal/modules/networkflow/binding_store.go`.
- Projection and revision paths: `internal/modules/projections/provider_factories.go`, `internal/modules/projections/provider_registry.go`, `internal/modules/projections/entity_grids.go`, `internal/modules/revisions/appender.go`, `internal/modules/revisions/indicator_children_test.go`, and relevant portions of projection query, rebuild, and delete/restore tests.
- Authored contracts and support manifests: `contracts/view-schemas/cartulary.view.indicators.v1.json`, `contracts/imports/view-targets.v1.json`, `contracts/imports/index.json`, `contracts/projection-providers/index.json`, `contracts/incident-bundles/source_catalog.json`, `contracts/recovery/fixtures/recovery-state-catalog.v1.json`, `contracts/verification/owners/module.indicators.json`, `tools/test_families/module.indicators.json`, `tools/schema_object_ownership_manifest.json`, and `tools/backend_module_boundaries.json`.
- Storage inputs: `db/migrations/00009_indicator_source.sql`, `db/migrations/00016_workbook_projections.sql`, `db/migrations/00026_indicator_child_rollback_tombstones.sql`, `db/migrations/00030_network_flow_indicator_bindings.sql`, and `db/migrations/00044_collaboration_owner_producers.sql`.
- Generic frontend and view-contract consumers: `packages/view-contracts/src/index.ts`, `packages/view-contracts/src/index.test.ts`, `packages/ui-contracts/src/workbook-shell-grid-selectors.test.ts`, `apps/web/src/workbook/models/workbookQuery.ts`, relevant Indicator cases in `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`, `apps/web/src/workbook/WorkbookShell.query.test.tsx`, `apps/web/e2e/keyboard.spec.ts`, and `apps/web/e2e/incident-administration.spec.ts`.
- Cross-owner test support: `internal/modules/workbook/testsupport/scenariotest/route_inventory.go`, `internal/modules/workbook/testsupport/scenariotest/record_fixtures.go`, and `internal/modules/workbook/notes_indicators_test.go`.

Generated consumers under `internal/gen/**`, `packages/protocol-ts/src/generated/**`, and `packages/ui-contracts/src/generated/**` were identified by search as drift-sensitive downstream artifacts. They must never be hand-edited; any later change must start from the adopted owner and authored contract input and use the Make-owned generator.

## 2. Current-State Repository Inventory

All 28 files currently under `internal/modules/indicators` are in scope and inventoried below. None is excluded.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/indicators/api.go` | Indicator view identity, root `Store`, create and participant DTOs, HTTP-shaped create decoding, request hashing, and row/result assembly. | `ViewSchemaID`, `IndicatorFindOrCreateParticipantV1`, `ErrInvalidCreateRequest`, `Store`, `NewStore`, create/participant/record/projection/result types, `DecodeCreateRequest`, `CreateRequestHash`, `BuildIndicatorRow`, `BuildMutationPayload`. | Workbook create contribution; indicator store, import, query, and tests; Network Flow consumes record/participant types. | Auth, Records, Revisions, view schema, HTTP API, Postgres, UUID. | Route, store, IP, workbook, and Network Flow tests. | Indicator view schema and its generated view/protocol consumers. | Indicators for semantic DTOs; Workbook for transport admission. | high | Transport error mapping and semantic types share one root file. |
| `internal/modules/indicators/boundary_guard_test.go` | Production sibling-import allowlist and guard against entity-source prefix reuse. | `TestIndicatorsProductionImportBoundaries`, `TestIndicatorsDoNotUseEntitiesSourcePrefixes`. | Go test and broad harness execution. | Go parser/filesystem inspection. | Self. | Backend module-boundary policy by alignment. | Indicators test boundary. | medium | Does not currently prove transport, projection-storage, or platform layering seams. |
| `internal/modules/indicators/deleterestore/provider.go` | Indicator source snapshot, tombstone update, view resolution, and delete preconditions for Revisions. | `Source`, `NewSource`, `SnapshotTx`, `UpdateSourceDeleteStateTx`, `ViewSchemaID`, `ValidateDeletePreconditionsTx`. | `RevisionProviderContribution`; Revisions command path through the composed catalog. | PGX, rollback/delete-restore contracts, fixed indicator SQL. | Revisions delete/restore and integration tests. | Revision provider catalog and indicator view contract. | Indicators source-owned provider. | high | Correctly source-owned but directly persistence-adjacent. |
| `internal/modules/indicators/import_create.go` | Indicator-owned Imports facade and caller-transaction row creation. | `ImportCreateCommand`, `NewImportCreateFacade`, `Store.CreateImportRowTx`. | Import assembly owner registry and Imports owner facade. | Imports owner facade, Revisions appender, PGX, indicator store internals. | Imports characterization/registry tests and indicator owner integration indirectly. | Import target/index projections. | Indicators, exposed through Imports facade. | high | Legitimate thin adapter; must keep import transaction and result semantics stable. |
| `internal/modules/indicators/incident_bundle_portability.go` | Deterministic export and fixed-SQL import for the three Indicator source files. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`. | Indicator incident-bundle source port. | Incident portability helper, PGX, raw fixed SQL, UUID. | Incident Bundle round-trip/integration coverage indirectly. | Incident-bundle source catalog and recovery/source layouts. | Indicators portability adapter. | critical | Raw fixed owner SQL is allowed; semantic validation must remain in the owner port. |
| `internal/modules/indicators/incident_bundle_source_port.go` | Constructs the typed portability descriptor and adapts prepare/apply/validate operations. | `NewIncidentBundleSourcePort`. | Incident portability assembly catalog. | Incident Bundles source-port contract, portability functions, PGX. | Incident Bundle source-catalog and route tests indirectly. | `contracts/incident-bundles/source_catalog.json`. | Indicators portability adapter. | critical | Declares ten Core 01 invariants but live validation checks only `indicators.representation_legal`. |
| `internal/modules/indicators/incident_bundle_subtype_presence.go` | Contributes Indicator record subtype presence to portability validation. | `IncidentBundleSubtypeContribution`; source methods. | Incident portability assembly. | Incident Bundle subtype-presence contract, PGX. | Incident Bundle catalog/integration tests. | Incident-bundle subtype/source catalog. | Indicators portability adapter. | low | Narrow, legitimate contribution. |
| `internal/modules/indicators/identity_characterization_test.go` | Locks normalization, representation validation, dedupe inputs, registry coverage, and invalid aliases before identity extraction. | `TestIndicatorIdentityCharacterization`, `TestIndicatorIdentityCharacterizationRejectsAliasesAndIncompleteHashes`. | Focused Indicator identity row. | Target-private identity functions only. | Self. | `module.indicators` identity evidence. | Indicators tests. | high | Covers every supported type and every currently legal type/value-kind combination without blessing aliases. |
| `internal/modules/indicators/indicators_test.go` | Canonical identity, repeated observations, lifecycle derivation, and Network Flow participant rollback evidence. | `TestIndicatorsCanonicalObservationLifecycle_Unit`, `TestNetworkFlowCore02_IndicatorFindOrCreateParticipantRollback`. | Focused owner row for the canonical lifecycle test; broad Go suite for participant coverage. | App test support, auth fixtures, timeline fixtures, revisions support, PGX. | Self. | `module.indicators` test-family evidence. | Indicators tests. | medium | Direct analyst-entry observations now use the exact `manual_entry` owner token. |
| `internal/modules/indicators/ip_identity_test.go` | Exact IPv4/IPv6 normalization, rejection, create validation, and dedupe evidence. | Four exported test functions. | Broad Go suite; not mapped to a focused `module.indicators` row. | Target-private identity functions. | Self. | Core 02 identity behavior; no direct generated artifact. | Indicators tests. | high | Required characterization before identity extraction. |
| `internal/modules/indicators/portability_characterization_test.go` | Locks deterministic export, complete member/null presence, timestamp encoding, stable path order, and shared v1/v2 source support. | `TestIndicatorPortableRowsCharacterization_Integration`. | Focused Indicator portability row. | Indicator facade, source port, portability decoder, Postgres fixtures. | Self. | `module.indicators` portability evidence and test-family topology. | Indicators tests. | critical | Characterizes the valid source-major-1 surface without adopting the still-pending invariant partition. |
| `internal/modules/indicators/network_flow_participant.go` | Active indicator lookup and Core 02 find-or-create participant inside the caller transaction. | `Store.GetActiveIndicatorParticipant`, `GetActiveIndicatorParticipantTx`. | Network Flow ports, binding store, and transaction participants. | PGX, fixed indicator queries, root identity/upsert behavior. | Indicator participant rollback and Network Flow route/behavior tests. | Network Flow generated/request contracts by behavioral dependency only. | Indicators participant port. | high | Must retain no independent commit, observation, lifecycle, link, binding, or audit side effect. |
| `internal/modules/indicators/projectionprovider/provider.go` | Deterministic incident-wide rebuild of `indicator_grid_projection`. | `RebuildIncidentIndicatorsTx`. | Projection provider factory through `projections/entity_grids.go`. | PGX; Records, Indicators, observations, intervals, and Links source tables; projection storage. | Indicator route rebuild equality and projections rebuild tests. | Projection provider descriptor and indicator view schema. | Indicators source provider; Projections owns storage lifecycle. | high | Legitimate provider boundary; row refresh remains elsewhere in the root store. |
| `internal/modules/indicators/projectionprovider/query_surfaces.go` | Declares generic query fields, storage columns, and row assembly metadata. | `QuerySurfaces`. | Projection assembly and generic projections query service. | Projection provider contract. | Projection query-surface and workbook query tests. | Projection provider descriptor and indicator view schema. | Indicators source query descriptor. | high | This is the assembled production query surface. |
| `internal/modules/indicators/query.go` | Indicator-specific direct query and keyset pagination over `indicator_grid_projection`. | `Store.QueryRows`, `Store.QueryRowsPage`; private SQL/scanning helpers. | No live production caller found; target-local paging test covers its SQL builder. | Query-page and view-schema platform contracts, PGX, direct projection SQL. | `query_paging_test.go`; generic route uses a different projections path. | Indicator view/query contract. | Projections query service, with source semantics contributed by Indicators. | high | Duplicates the assembled generic query path; characterize before retirement. |
| `internal/modules/indicators/query_paging_test.go` | Proves target-local query SQL uses bounded keyset pagination without `OFFSET`. | `TestIndicatorPageSQLIsKeysetBounded`. | Broad Go suite; not mapped to focused Indicator rows. | Target-private query builder, query-page and view-schema types. | Self. | Query contract by behavior. | Indicators/Projections characterization. | medium | Equivalent coverage must move to the production generic path before `query.go` removal. |
| `internal/modules/indicators/recovery_state.go` | Declares three authoritative Indicator tables for recovery. | `RecoveryStateContribution`. | Recovery assembly state catalog. | Recovery state contract. | Recovery catalog/restore tests indirectly. | Recovery state catalog. | Indicators recovery contribution. | low | `indicator_grid_projection` correctly remains Projections-owned derived state. |
| `internal/modules/indicators/resolution_integration_test.go` | Real HTTP create/query, attribution, observations/lifecycle projection, and rebuild-equality evidence. | `TestIndicatorsRoute_Integration`. | Focused Indicator integration test row. | Server/app support, workbook scenario query helper, timeline fixtures, target Store. | Self. | Indicator behavior and incident-portability verification IDs. | Indicators integration test. | medium | Strong route/rebuild baseline; direct analyst-entry fixtures now use `manual_entry`. |
| `internal/modules/indicators/revision_append_port.go` | Consumer-owned narrow adapter to the Revisions append-only facade. | Private port and adapter methods. | Root Store only. | Revisions appender, PGX, UUID. | Indicator store tests exercise it; Revisions tests cover appended history/intents. | Revisions/collaboration behavior contracts. | Indicators application adapter. | medium | `AppendRecordRevisionTx` already appends the collaboration intent through Revisions. |
| `internal/modules/indicators/revision_provider_contribution.go` | Composes Indicator record delete/restore, row rollback, and observation/interval rollback contributions. | `RevisionProviderContribution`. | Revision assembly. | Indicator delete/restore and rollback subpackages; Revisions provider contract. | Revision assembly, delete/restore, and child rollback integration tests. | Revisions provider catalog. | Indicators revision contribution. | high | Legitimate source-owner contribution; exported subpackage surface can be narrowed later. |
| `internal/modules/indicators/rollbackprovider/children.go` | Describes and inverses observation/lifecycle history entries, locks affected records, tombstones or restores child state, and reports changed fields. | `ChildProvider`, `NewChildProvider`, `DescribeTx`, `ApplyInverseTx`. | Indicator revision contribution through Revisions. | PGX, rollback contract, direct Indicator child and Records SQL. | Local child tests and `revisions/indicator_children_test.go`. | Core 02 rollback/portability behavior. | Indicators rollback provider. | critical | High-risk cross-record effects; must preserve exact protected sets, tombstones, projection effects, and history. |
| `internal/modules/indicators/rollbackprovider/children_test.go` | Parser compatibility, identity-drift rejection, and affected-field evidence. | Three exported test functions. | Broad Go suite; not mapped to focused Indicator rows. | Target rollback provider internals. | Self. | Rollback behavior only. | Indicators tests. | medium | Expand before moving provider code. |
| `internal/modules/indicators/rollbackprovider/provider.go` | Validates historical Indicator values and restores the authoritative Indicator source row. | `Provider`, `NewProvider`, `ValidateRollbackValue`, `RestoreTx`. | Indicator revision contribution through Revisions. | PGX, rollback contract, duplicated indicator validation/dedupe rules. | Local provider tests and Revisions rollback tests. | Indicator identity/history contract. | Indicators rollback provider. | high | Duplicates live create identity semantics and can drift. |
| `internal/modules/indicators/rollbackprovider/provider_test.go` | Historical value filtering and dedupe-key stability evidence. | Two exported test functions. | Broad Go suite; not mapped to focused Indicator rows. | Target rollback provider internals. | Self. | Indicator rollback contract. | Indicators tests. | medium | Preserve while replacing duplicate identity logic. |
| `internal/modules/indicators/store.go` | Canonical create/dedupe, observation create/resolve, lifecycle append, transactions, incident state, route idempotency, Records ports, Revisions, projection refresh, normalization, validation, and direct SQL. | `IndicatorCreateValidationError`; observation/lifecycle parameter and record types; five public `Store` methods. | Workbook create adapter, Imports facade, Network Flow participant, target and cross-owner tests. | Auth, Records, Revisions, Postgres/PGX, projection table, JSON/net/url/time utilities. | Indicator store/route/unit/IP tests; workbook, revisions, Network Flow, projection, and import tests indirectly. | View schema, projection provider, import, portability, collaboration, recovery, and generated consumers. | Indicators application facade plus source repository; Projections for storage mutation boundary. | critical | Primary mixed-responsibility refactor target. |
| `internal/modules/indicators/testsupport/fixtures.go` | Shared example values and direct projection fixture lookup. | `Example`, `Examples`, time constants, `ProjectionRow`, `LookupProjection`. | Indicator target tests. | Postgres test DB, UUID, direct projection query. | Indicator store and route tests. | Test evidence only. | Indicators test support. | low | Must remain test-only. |
| `internal/modules/indicators/testsupport/routetest/inventory.go` | Incident authorization control inventory for Indicator query. | `ControlQuery`. | Incident HTTP conformance tests. | Generic route inventory and Indicator view ID. | Incidents conformance suites. | Route inventory/evidence accounting. | Indicators test support with Workbook route ownership. | medium | Local inventory covers query; generic workbook scenario inventory covers create/replay/auth. |
| `internal/modules/indicators/unit_test.go` | Focused repeated-observation separation and projection aggregation evidence. | `TestIndicatorObservationSeparation_Unit`. | Focused Indicator store row. | App/store fixtures, timeline fixtures, revisions support, target Store. | Self. | `module.indicators` test-family evidence. | Indicators tests. | medium | Repeated direct analyst-entry observations use `manual_entry`. |

## 3. Module Boundary Diagnosis

The current target is a legitimate permanent source-owner module, not an accidental boundary created solely by the directory name. Core 01 requires an Indicators concern, and Core 02 assigns canonical indicators, source-bound observations, lifecycle intervals, and the named Network Flow transaction participant to that owner. The package is nevertheless mixed-responsibility: the root combines application coordination, fixed persistence, transport-shaped admission, identity rules, projection mutation, and a redundant query implementation.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Canonical Indicator identity, normalization, validation, and dedupe | `api.go`, `store.go`, duplicated in `rollbackprovider/provider.go` | Indicators | split | Core 02 REQ-02-072 through REQ-02-074C; IP tests; live/rollback implementations | Create one owner-internal identity component while preserving exact current conformant outputs. |
| Canonical row create application coordination | `Store.CreateIndicatorRow` | Indicators | keep | Core 01 row-create contract; workbook contribution; route integration test | Keep as a facade, but separate transaction/idempotency coordination from persistence. |
| HTTP JSON/view-schema admission and HTTP error construction | `api.go`, called by Workbook contribution | Workbook transport adapter plus Indicators semantic validation | move | Generic workbook routes own the public envelope; `api.go` imports `httpapi` | Do not change request fields, normalization, errors, or hash/replay semantics. |
| Indicator source persistence | `store.go` | Indicators | split | Schema ownership maps `indicators`, `indicator_observations`, and `indicator_state_intervals` to Indicators | Fixed SQL may remain, but behind a source repository seam. |
| Observation capture/resolution | `store.go` | Indicators | keep | Core 02 REQ-02-075 through REQ-02-078 and REQ-02-260 | Raw text, source binding, repeated rows, attribution, and rollback are frozen. |
| Indicator lifecycle interval append | `store.go` | Indicators | keep | Core 02 REQ-02-079, REQ-02-080, and REQ-02-260 | Append-only semantics remain distinct from observation times. |
| Network Flow canonical find/create participant | `network_flow_participant.go`, root Store helpers | Indicators | keep | Core 02 REQ-02-074C and adopted Network Flow NLSpec | Remain binding-only and caller-transaction scoped. |
| Projection source derivation and query descriptor | `projectionprovider/**` | Indicators source provider | keep | Core 01 REQ-01-622 and live provider descriptor | Projection storage lifecycle remains Projections-owned. |
| Projection row refresh | `store.go` | Indicators source provider behind a Projections-owned coordination port | split | Root writes `indicator_grid_projection`; ownership manifest assigns table to Projections | Preserve same-transaction refresh and rebuild equality. |
| Projection query execution | `query.go` and generic Projections service | Projections query service with Indicators query descriptor | move | Live workbook assembly calls generic `projectionQuery.QueryRowsPage`; no production target-local caller found | Retire only after generic keyset behavior is characterized. |
| Imports apply | `import_create.go` | Indicators owner facade coordinated by Imports | keep | Import owner registry and target contracts | No parser or job coordination belongs in Indicators. |
| Incident portability | `incident_bundle_*` | Indicators owner port coordinated by Incident Bundles | keep | Core 01 REQ-01-639/640 and source catalog | Full declared invariant validation is currently incomplete. |
| Recovery state | `recovery_state.go` | Indicators contribution coordinated by Recovery | keep | Recovery state catalog | Only authoritative source tables are contributed. |
| Record and child history/rollback | `revision_*`, `deleterestore`, `rollbackprovider` | Indicators providers coordinated by Revisions | keep | Core 02 REQ-02-260 and Revisions provider catalog | Narrow exported providers; do not centralize source semantics in Revisions. |
| Collaboration publication | Revisions `Appender`, reached through `revision_append_port.go` | Collaboration/Revisions | keep | Core 01 REQ-01-271A; `AppendRecordRevisionTx` appends one intent | Indicators supplies row semantics but owns no WebSocket transport. |
| Saved-view behavior | Generic saved-view/query controllers and contracts | Saved Views/Workbook | defer | No saved-view code or import exists in the target | Freeze view/query identity; no Indicator-specific saved-view module. |
| Frontend timestamp filter inference | `apps/web/src/workbook/models/workbookQuery.ts` | Workbook/view-contract consumer | defer | Raw checks for Indicator observation timestamp field keys | Prefer future `read_kind` metadata consumption; not part of this backend module refactor. |
| Frontend shell/controller and grid-vendor integration | Generic Workbook shell, UI contracts, grid adapter | Workbook/UI/grid adapter | defer | No target production frontend or grid-vendor import exists | No reason to create an Indicator-specific frontend controller or vendor adapter. |
| Future normalization registry | Current Indicator owner; draft Reference Pack direction | Indicators until a later adopted owner changes it | defer | Reference Pack NLSpec is draft | Do not move behavior based on draft ownership. |

Repository/framework mismatch finding: the live production query is already generic and Projections-owned, while the target still retains a separate direct query implementation. In addition, the server exposes the generic projection catalog through the historically named `Runtime.Timeline.ProjectionCatalog` bundle. These names and duplicate paths are current implementation state, not evidence that Timeline owns Indicator behavior.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/cartulary.view.indicators.v1/rows` | Workbook route; Indicators create semantics | Core 01 REQ-01-032, REQ-01-057, REQ-01-331; workbook route/contribution | Indicator route integration; generic workbook scenario inventory; browser create cases | Closed body, minimum signal, CSRF, viewer denial, hidden incident, `201` first commit, `200` exact replay, divergent `client_txn_conflict`, response-row equality | critical | Transport admission movement must not change the wire. |
| `POST /api/v1/incidents/{incident_id}/views/cartulary.view.indicators.v1/query` | Workbook/Projections | Core 01 query and projection requirements; generic projection assembly | Indicator route integration, workbook query tests, browser shell tests | Default sort, filters, grouping, keyset cursor binding, terminal paging, auth loss, create/query row parity | critical | Production query uses generic Projections, not `Store.QueryRowsPage`. |
| `cartulary.view.indicators.v1` discovery and view-row shape | Core 01/view schema; source semantics Indicators | Authored view-schema contract and view-contract consumers | View-schema registry, view-contract, UI selector, Workbook surface tests | Exact 13 cells, technical fields, create-only writeability, no grid editing, inspector features, stable selector identity | high | No `view_schema_id` or field-key change is planned. |
| Saved views over the Indicator schema | Saved Views/Workbook | Core 01/03/04 saved-view contracts; browser saved-view scenario | Saved Views browser and Workbook controller tests | Saved query/layout continues to resolve the same base schema after query-path cleanup | high | Target owns neither persistence nor authorization. |
| Canonical create and dedupe | Indicators | Core 02 REQ-02-072 through REQ-02-074B | Canonical lifecycle/unit tests and IP tests | All indicator/value kinds, hash-pair rules, defanged/STIX non-identity, concurrent same-key create, idempotency/dedupe distinction | critical | Identity logic must have one implementation. |
| `indicator_find_or_create_participant_v1` | Indicators; Network Flow consumer | Core 02 REQ-02-074C; adopted Network Flow NLSpec | Participant rollback; Network Flow route/behavior tests | Created/reused result, canonical return, caller rollback, concurrent convergence, no observation/lifecycle/link/binding side effect | critical | Keep caller transaction and no independent audit/publication. |
| Indicator observation create/resolve | Indicators | Core 02 REQ-02-075 through REQ-02-078, REQ-02-260, origin registry | Store/unit tests and Revisions child rollback integration | `IND-ORIGIN-001` through `IND-ORIGIN-005` and `AC-ORIGIN-001` through `AC-ORIGIN-006` | critical | Current tests contain invalid `interactive_cell` fixtures; SL-06 direction is authorized but unimplemented. |
| Indicator lifecycle append | Indicators | Core 02 REQ-02-079/080/260 | Canonical lifecycle/unit and Revisions child rollback tests | Append-only ordering/coherence, lifecycle/observation time separation, rollback tombstone, projection selection | high | No public target-specific route was found. |
| Projection refresh/query/rebuild | Indicators source provider; Projections storage/query | Core 01 REQ-01-342 through REQ-01-363 and REQ-01-621/622 | Route rebuild equality, workbook query, projection catalog/rebuild tests | Same-transaction refresh, query/load row equality, generic keyset paging, deterministic incident/restore rebuild | critical | Provider capability and authored manifest must stay aligned. |
| Record revisions and change sets | Revisions with Indicator source contribution | Core 02 history; Revisions appender and provider catalog | Indicator attribution test; Revisions integration and child rollback tests | Create/import/observation/lifecycle revision count and exact before/after rows | critical | Preserve actor, timestamp, source, client transaction, and request identity. |
| `record_changed` WebSocket semantics | Collaboration/Revisions | Core 01 REQ-01-267 and REQ-01-271A; Revisions appender | Revisions rollback event tests; no Indicator create assertion in focused owner tests | Exactly one create intent per record revision, canonical field keys/view, rollback behavior, no intent after transaction rollback | critical | Target has no direct WebSocket transport code. |
| Delete, restore, row rollback, observation rollback, interval rollback | Revisions routes; Indicators providers | Core 01 record routes; Core 02 REQ-02-260 | Revisions delete/restore, rollback, and indicator-child tests | Provider-catalog wiring, locks, row versions, tombstones, projections, ordinary events, exact replay | critical | Provider extraction must not change protected sets. |
| Indicator Imports facade | Imports coordinator; Indicators source facade | Import target/index contracts and owner registry | Imports characterization/registry and cross-owner tests | Create/dedupe, transaction rollback, projection/revision result, normalized errors | high | No CSV/XLSX parsing belongs in Indicators. |
| Incident Bundle three-file family | Incident Bundles coordinator; Indicators source port | Core 01 REQ-01-639/640 and source catalog | Incident Bundle catalog and round-trip tests; focused owner row currently shares portability verification | `IND-PORT-001` through `IND-PORT-006` and `AC-PORT-001` through `AC-PORT-007` after owner adoption | critical | Current validation proves only representation legality; the tracker does not own the missing partition. |
| Recovery ownership | Recovery coordinator; Indicators contribution | Recovery state catalog | Recovery catalog and restore rebuild tests | All three source tables restored, projection excluded and rebuilt | high | Projection is derived and must not enter authoritative recovery state. |
| Authorization and incident lifecycle | Auth/Workbook/Incidents | Core 04 incident roles; Workbook routes; root incident-open guard | Generic route scenario tests and Indicator route test | Membership re-derivation, editor minimum for create, member query, closed-incident mutation rejection, replay ordering | critical | Do not move route authorization into Indicator domain code. |
| Generated protocol/view contracts | Contract owners and generators | Authored view schema, OpenAPI source, generated-artifact policy | Generation drift, JSON shape, protocol/view-contract tests | Zero generated drift after any authored input change | high | Generated roots are never hand-edited. |
| Generic frontend/grid/selectors | Workbook, View Contracts, UI Contracts, grid adapter | View-contract and selector source; browser scenarios | View-contract tests, UI selector tests, Workbook unit/browser tests | Same shell selection, read-only existing rows, create draft, saved-view continuity, stable selectors | medium | No Indicator-specific frontend or vendor layer is proposed. |
| Harness/test accounting | Testing Harness and `module.indicators` owner | Owner verification contract and test-family manifest | Three mapped service-backed rows | Map newly added/moved owner tests without treating rows as runtime architecture | high | IP, paging, rollback, boundary, participant, and cross-owner tests currently rely on broader suites. |

### 4.1 Observation-origin implementation contract

The following requirements translate the existing Core 02 closed vocabulary into SL-06. The direction is authorized as an intentional implementation-conformance correction. This authorization does not permit a repository change outside a later implementation task.

| Requirement ID | Requirement |
| --- | --- |
| `IND-ORIGIN-001` | `indicator_observation.origin_kind` MUST accept and persist exactly `manual_entry`, `clipboard_paste`, `csv_import`, `xlsx_import`, `api_import`, `extraction`, or `system`. It MUST reject every other value before mutation. |
| `IND-ORIGIN-002` | Every live producer MUST emit the token assigned by the producer mapping below. HTTP transport alone MUST NOT classify an operation as `api_import`. |
| `IND-ORIGIN-003` | Indicators MUST own one transport-neutral parser and semantic error. All ordinary create, import, extraction, rollback restoration, and portability preparation paths MUST use that parser. Transport owners MUST map the semantic error to their existing public error contracts. |
| `IND-ORIGIN-004` | Before rejection is enabled, the implementation task MUST complete and record the persisted-data preflight below. No runtime alias, repair-on-read, fallback, or permissive compatibility path is authorized. |
| `IND-ORIGIN-005` | Invalid origin input MUST fail before the first database write and MUST leave every owner and collaborator state named in the acceptance matrix unchanged. Exact valid tokens and provenance MUST survive history, rollback, and portability round trips. |

The exhaustive producer mapping is:

| Producer or action | Required token |
| --- | --- |
| Direct analyst entry through the workbook, inspector, or ordinary interactive API operation | `manual_entry` |
| Workbook clipboard paste | `clipboard_paste` |
| CSV file import | `csv_import` |
| XLSX file import | `xlsx_import` |
| Structured import submitted through an API import contract | `api_import` |
| Parser, extraction, or machine-assisted source-text extraction | `extraction` |
| Explicit trusted internal system-process operation | `system` |

An ordinary human-originated mutation transported over HTTP defaults to `manual_entry`. The `system` token has no caller-selectable default: it MUST require an explicit trusted system-operation context, and an ordinary caller MUST NOT self-classify as `system`.

The implementation boundary is conceptually equivalent to:

```go
type ObservationOrigin string

func ParseObservationOrigin(raw string) (ObservationOrigin, error)
```

The concrete name and file placement are intentionally unspecified. The behavior is not: comparison MUST be byte-exact and MUST NOT trim whitespace, fold case, expand aliases, accept extension prefixes, or substitute a default. A database constraint MAY mirror the seven-token set as defense in depth, but it MUST NOT be the sole validation boundary. The Indicator-owned error MUST contain semantic classification sufficient for existing adapters and MUST NOT import HTTP, job, or archive-transport types.

The persisted-data preflight MUST inspect current fixtures, migrations and seed data, retained supported Incident Bundles, upgrade fixtures, and authoritative `indicator_observation` rows in every supported deployment context. It MUST select exactly one outcome:

| Observed state | Required action | Authorization consequence |
| --- | --- | --- |
| No non-Core value exists. | Enable exact validation without data migration. | SL-06 may proceed. |
| Non-Core values exist only in disposable test fixtures. | Replace each fixture with the token for its actual producer. | SL-06 may proceed after fixture review. |
| Persisted `interactive_cell` rows have provable direct-user-entry provenance. | Use an explicit forward migration to `manual_entry` and record the migration evidence. | A separately scoped migration task is required before enablement. |
| Persisted provenance cannot be proven. | Do not guess or rewrite. Require operator remediation or quarantine/fail-closed handling. | A separately authorized remediation plan is required; SL-06 enablement remains blocked. |

Known `interactive_cell` fixtures in `indicators_test.go`, `unit_test.go`, and `resolution_integration_test.go` MUST become `manual_entry` unless direct fixture evidence establishes another producer class. The correction MUST preserve source binding, raw observed text or span, deterministic source location, and separate rows for repeated observations.

#### Observation-origin acceptance matrix

| Acceptance ID | Scenario | Required result |
| --- | --- | --- |
| `AC-ORIGIN-001` | Parse each of the seven exact tokens. | Each token is accepted unchanged; no eighth value is accepted. |
| `AC-ORIGIN-002` | Parse `interactive_cell`, an empty or missing required value, surrounding whitespace, `Manual_Entry`, an unknown value, and an extension-prefixed value. | Every case returns the typed Indicator semantic error before mutation. |
| `AC-ORIGIN-003` | Exercise each live producer class. | The stored token equals the producer mapping; ordinary HTTP analyst entry stores `manual_entry`. |
| `AC-ORIGIN-004` | Reject an origin at the latest reachable pre-write boundary. | Observation rows, source and Indicator versions, change sets, revisions, projections, collaboration intents, and idempotency-success state remain unchanged. |
| `AC-ORIGIN-005` | Preserve a valid token through history, rollback, and Incident Bundle export/import. | The exact token round trips without normalization or aliasing. |
| `AC-ORIGIN-006` | Submit repeated observations with identical origin, text, and normalized candidate but distinct stable identities. | The observations remain separate and retain independent provenance. |

SL-06 authorization recorded by the user for this tracker is:

> **SL-06 is authorized as an intentional implementation-conformance correction.** The Indicators owner shall accept and persist `indicator_observation.origin_kind` only as one of `manual_entry`, `clipboard_paste`, `csv_import`, `xlsx_import`, `api_import`, `extraction`, or `system`. `interactive_cell`, unknown tokens, aliases, case variants, and whitespace-normalized variants shall be rejected before mutation. Direct analyst-entry fixtures shall use `manual_entry`; fixtures representing another producer shall use that producer's exact Core token. No runtime compatibility alias, fallback token, or permissive read/write path is authorized. Any existing persisted non-Core value requires a separately reviewed data-remediation decision.

### 4.2 Indicator Incident Bundle owner-closure contract

| Requirement ID | Requirement |
| --- | --- |
| `IND-PORT-001` | Before SL-07, Core 01 MUST adopt exact portable row contracts and deterministic export ordering, and Core 04 MUST adopt complete binary acceptance criteria. |
| `IND-PORT-002` | Core 01 MUST adopt an exclusive rule for all ten declared invariants plus deterministic multi-defect precedence and tie-breaking before code emits a public exact `invariant_id`. |
| `IND-PORT-003` | After adoption, the Indicators port MUST keep Descriptor, Export, Prepare, Apply, and Validate responsibilities within the source-port boundaries defined below. |
| `IND-PORT-004` | After adoption, every semantic failure MUST select one deterministic exact invariant and map to the safe public error contract without leaking hostile or internal content. |
| `IND-PORT-005` | Apply, final-state validation, repair/rebuild coordination, publication, and terminal success MUST remain atomic in the Incident Bundles coordinator's final transaction. |
| `IND-PORT-006` | SL-07 completion MUST provide every acceptance case in `AC-PORT-001` through `AC-PORT-007` and current authored harness accounting. |

Current-state finding: the Indicators descriptor declares the following ten Core 01 invariant IDs, while live validation proves only `indicators.representation_legal`:

1. `indicators.representation_legal`
2. `indicators.normalization_exact`
3. `indicators.identity_unique`
4. `indicators.observation_same_incident`
5. `indicators.observation_ordered`
6. `indicators.observation_coherent`
7. `indicators.interval_same_incident`
8. `indicators.interval_ordered`
9. `indicators.interval_coherent`
10. `indicators.repeated_observations_preserved`

The tracker MUST NOT supply missing product authority. Before SL-07 implementation, adopted Core 01 MUST define the exact members, types, required presence, nullability, bounds, and canonical forms for `data/indicators.ndjson`, `data/indicator_observations.ndjson`, and `data/indicator_state_intervals.ndjson`; deterministic export order; the exclusive ten-invariant partition; multi-defect precedence and tie-breaking; and the Prepare-versus-Validate allocation. Adopted Core 04 MUST supply binary acceptance criteria covering every invariant. The corresponding authored traceability and verification inputs MUST then be updated and generated outputs MUST be produced through Make.

The current exporter bytes and member ordering MUST be characterized before the Core 01 amendment is written. `SELECT *`, physical relation order, and illustrative DDL MUST NOT be treated as portable-shape authority. The proposed default is to preserve a characterized deterministic exporter order. If no stable contract exists, the proposed fallback is ascending stable owner identity: Indicator `record_id`, observation ID, then interval ID for their respective files. No byte-order change may be combined with invariant enforcement without separate authorization.

#### Proposed Core 01 invariant partition

This table is proposed owner language. It is non-authoritative until adopted in Core 01 and MUST NOT be implemented directly from this tracker.

| Precedence | Invariant | Exclusive rule proposed for adoption | Minimum negative fixture |
| ---: | --- | --- | --- |
| 1 | `indicators.representation_legal` | Indicator rows have the exact admitted shape and types; incident identity matches the import context; type, value kind, display, normalized value, hash pair, defanged/STIX value, and dedupe-basis fields form a legal Core 02 representation. IP family/value-kind and hash-pair rules belong here. | `ipv4_addr` with non-`atomic`; a half-populated hash pair; unknown indicator type. |
| 2 | `indicators.normalization_exact` | Imported canonical display, normalized value, hash value, and stored or recomputed dedupe basis already equal the owner canonicalization result. Import rejects rather than repairs noncanonical values. | Noncanonical IPv6 text; normalized value differs from the owner result; dedupe key differs from recomputation. |
| 3 | `indicators.identity_unique` | Active imported Indicator identities are unique under the incident-scoped, type-specific dedupe key. Apply does not silently reuse, merge, or ignore a duplicate. | Two different Indicator record IDs recompute to the same active canonical identity. |
| 4 | `indicators.observation_same_incident` | Every observation, source record, optional resolved Indicator, and incident-scoped reference belongs to the import incident. A non-record extension resource is not a source record. | Foreign-incident source record or resolved Indicator. |
| 5 | `indicators.observation_ordered` | Creation, update, resolution, deletion/tombstone, and row-version values obey the adopted chronology and monotonicity rules; export order is deterministic. | `updated_at < created_at`, deletion before creation, or nonpositive row version. |
| 6 | `indicators.observation_coherent` | Observation rows have the exact admitted shape; source field, origin, locator, raw text/span, parse result, extraction metadata, resolution metadata, attribution, and tombstone members form one legal tuple. Origin uses only `IND-ORIGIN-001`. | `interactive_cell`; resolved observation without target or attribution; half-populated tombstone tuple. |
| 7 | `indicators.interval_same_incident` | Every interval, Indicator, supporting record, and predecessor belongs to the required incident; a predecessor belongs to the same Indicator. | Foreign-incident Indicator or support; predecessor for a different Indicator. |
| 8 | `indicators.interval_ordered` | Validity, creation, supersession, and deletion chronology obey the adopted order; supersession is acyclic and points backward in owner-defined history order. | `valid_to < valid_from`; forward or cyclic supersession. |
| 9 | `indicators.interval_coherent` | Interval rows have the exact admitted shape; state, confidence, rationale, support, attribution, supersession, and tombstone members form a legal append-only tuple. Observation time is not substituted for lifecycle-validity time. | Out-of-bounds confidence; invalid state tuple; half-populated tombstone or supersession tuple. |
| 10 | `indicators.repeated_observations_preserved` | Final transaction state contains exactly one row per admitted stable observation identity. Distinct IDs remain distinct regardless of identical content, source, origin, or resolved Indicator; required active and tombstoned rows remain present. | Two identical-content observations with distinct IDs; fail if either is coalesced, omitted, or overwritten. |

The proposed deterministic failure rule is: select the lowest precedence number among all defects. Within that invariant, order by logical file path, then valid stable owner identity ascending, then—only when stable identity is missing or invalid—an internal SHA-256 digest of the bounded raw logical row. The digest MUST NOT appear in public errors, logs, or telemetry.

#### Proposed source-port interface allocation

This allocation also requires Core 01 adoption before it becomes binding product behavior.

| Operation | Indicators responsibility |
| --- | --- |
| `Descriptor` | Declare the three paths, family ID, dependency, owner relations, contract major, stable identities, and all ten invariant IDs. |
| `Export` | Read only Indicator-owned source state and emit deterministic canonical files. |
| `PrepareImport` | Strictly decode the adopted row shapes, enforce bounds, verify canonical input and row-local invariants, and return an opaque prepared value without mutation. |
| `ApplyImportTx` | Use fixed owner-controlled SQL or SQLC in the supplied transaction and require exact affected-row counts. |
| `ValidateImportTx` | Compare admitted input with final transaction state and prove aggregate, cross-row, and cross-family invariants before publication. |

The proposed implementation uses an unexported prepared-Indicators type, one closed Indicator bundle-invariant type representing exactly the ten IDs, and a typed invariant-error constructor fixed to family `indicators`. It reuses the Indicator identity component from SL-02 and narrow Records/actor lookup capabilities. It MUST NOT contain authorization, HTTP mapping, projection publication, job coordination, generic metadata-derived SQL, independent commits, generic conflict-ignore behavior, or prepared-value reuse across ports. Neither Prepare nor Validate repairs, skips, merges, defaults, or canonicalizes input unless an adopted owner explicitly permits that operation.

An Indicator invariant failure must map through Incident Bundles to exactly:

```text
error.code = incident_bundle_import_rejected
reason_code = source_family_invalid
retryable = false
source_family_id = indicators
invariant_id = <one exact adopted Indicator invariant>
```

The public error, logs, telemetry, and job summary MUST NOT disclose imported values, raw NDJSON, SQL or database errors, relation or column names, object keys or staging paths, actor/provider identifiers, credentials, or cryptographic material. The port MUST classify semantic defects explicitly and MUST NOT parse database error text or map every database failure to `indicators.representation_legal`.

All source Apply operations, all Validate operations, revision repair, attribution flush, projection rebuild, imported-incident administration, audit, publication, and terminal success MUST share the coordinator's final database transaction. Only its commit may expose the incident or final object references. SL-07 does not authorize a new bundle version, source contract major, route, public error code, compatibility importer, Indicator-specific job, or Indicator-specific transaction coordinator.

#### Portability acceptance matrix

| Acceptance ID | Scenario | Required result |
| --- | --- | --- |
| `AC-PORT-001` | One valid-JSON fixture violates exactly one proposed invariant. | Each of the ten fixtures returns family `indicators`, the exact adopted invariant ID, `retryable=false`, and no visible state. |
| `AC-PORT-002` | At least three fixtures violate multiple invariants; input row order varies. | The adopted precedence and tie-break choose the same exact invariant ID. |
| `AC-PORT-003` | Export valid v2, import into an empty target, and re-export. | All three files are authoritatively equivalent under the adopted row and order contract. |
| `AC-PORT-004` | Import a retained valid v1 bundle and export v2. | Supported v1 semantics are preserved; no new version or source major is required. |
| `AC-PORT-005` | Import repeated identical-content observations with distinct IDs and active/tombstoned observation and interval history. | Distinct provenance and tombstones survive; active reads and projections exclude tombstones as owned; re-export preserves history. |
| `AC-PORT-006` | Inject failure after Indicator Apply, during Indicator Validate, and after validation but before commit. | No incident, membership, workbook preference, audit success, projection, final object reference, idempotency success, or terminal-success result becomes visible. |
| `AC-PORT-007` | Supply hostile values, SQL-like text, paths, and secrets. | Public errors, logs, telemetry, and job summaries contain none of the forbidden content. |

### 4.3 Requirement traceability

| Requirement | Authoritative or required owner | Slice | Acceptance evidence | Entry gate | Current status |
| --- | --- | --- | --- | --- | --- |
| `IND-ORIGIN-001` through `IND-ORIGIN-003` | Adopted Core 02; authorized SL-06 translation | SL-02, SL-06 | `AC-ORIGIN-001` through `AC-ORIGIN-003` | SL-00 characterization and SL-02 identity boundary | Authorized; not implemented |
| `IND-ORIGIN-004` | Deployment/data-remediation authority | SL-06 | Recorded preflight outcome and any separately authorized migration evidence | Preflight completed before enablement | READY for preflight |
| `IND-ORIGIN-005` | Adopted Core 01/Core 02 transaction and history owners | SL-06 | `AC-ORIGIN-004` through `AC-ORIGIN-006` | Exact parser implemented | Authorized; not implemented |
| `IND-PORT-001` | Required Core 01 and Core 04 amendments | Owner-document task before SL-07 | Adopted row/order contract and binary Core 04 criteria | Owner changes adopted | BLOCKED: owner definition incomplete |
| `IND-PORT-002` | Proposed Core 01 invariant partition | Owner-document task before SL-07 | `AC-PORT-001`, `AC-PORT-002` | Exclusive partition and precedence adopted | BLOCKED: owner definition incomplete |
| `IND-PORT-003` | Proposed Core 01 source-port phase allocation | SL-05, SL-07 after adoption | Port boundary tests and `AC-PORT-003` through `AC-PORT-005` | SL-05 plus owner adoption | BLOCKED |
| `IND-PORT-004` | Core 01 Incident Bundle error owner and Core 04 security owner | SL-07 after adoption | `AC-PORT-001`, `AC-PORT-002`, `AC-PORT-007` | Exact invariant selection adopted | BLOCKED |
| `IND-PORT-005` | Core 01 transaction/publication owner | SL-07 after adoption | `AC-PORT-006` | Final-transaction contract adopted | BLOCKED |
| `IND-PORT-006` | Testing Harness evidence routing only | SL-08 | Focused rows, retained run roots, generated drift gates | Product requirements adopted and implemented | BLOCKED |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Indicators is a valid permanent source-owner boundary. | Core 01 required modules; Core 02 Indicator and observation requirements; schema ownership manifest. | low | `intentional/no_action` | Indicators | Retain the module and refactor within its adopted responsibility. |
| The root `Store` mixes transaction/idempotency coordination, source persistence, revision append, identity rules, and projection mutation. | `api.go` Store fields and `store.go` methods/SQL. | critical | `should_fix` | Indicators facade plus source repository; Projections coordination port | Establish characterization, then split without changing the compatibility facade initially. |
| HTTP and view-schema decoding leaks into the Indicator root. | `api.go` imports `httpapi` and is invoked only by Workbook contribution. | high | `should_fix` | Workbook adapter with Indicator semantic command | Move JSON/error adaptation only after exact request/hash/error characterization. |
| Target-local query duplicates the assembled generic production query. | No production `Store.QueryRows*` caller; workbook assembly calls Projections query service using Indicator query surfaces. | high | `should_fix` | Projections query service with Indicator descriptor | Move keyset coverage, then retire the redundant surface. |
| Root source mutations directly write Projections-owned storage. | `store.go` refresh SQL; schema ownership and provider descriptor assign `indicator_grid_projection` to Projections. | critical | `should_fix` | Indicators projection source/provider behind Projections port | Preserve transactionality; align provider capability and authored manifest if row refresh becomes catalog-visible. |
| Live create and rollback duplicate Indicator validation/dedupe behavior. | `store.go` and `rollbackprovider/provider.go`. | critical | `must_fix` | Indicators internal identity component | Centralize before further provider movement and retain historical-value compatibility. |
| Portability declares ten invariants but validates only representation legality. | `incident_bundle_source_port.go`; Core 01 REQ-01-640; source catalog; `IND-PORT-001` through `IND-PORT-006`. | critical | `must_fix` | Core 01/Core 04 first; then Indicators source port | Do not let code invent the portable shape or public invariant partition. Adopt the owner closure, authorize SL-07, then prove `AC-PORT-001` through `AC-PORT-007`. |
| Indicator tests persist `interactive_cell`, outside the closed origin vocabulary. | Four target fixtures versus Core 02 §18 exact token registry; `IND-ORIGIN-001` through `IND-ORIGIN-005`. | high | `must_fix` | Indicators observation owner | SL-06 direction is authorized. Complete the data preflight, then implement and prove `AC-ORIGIN-001` through `AC-ORIGIN-006` in a later task. |
| Focused Indicator evidence covers only three service-backed tests. | `tools/test_families/module.indicators.json`; `make explain-test-owner OWNER=module.indicators`. | high | `should_fix` | Testing Harness authored owner inputs | Add rows for new characterization/moved tests; broad rows remain collaborator evidence. |
| The generic frontend infers Indicator timestamp filters from raw field keys. | `apps/web/src/workbook/models/workbookQuery.ts`. | medium | `defer` | Workbook/View Contracts | Future metadata-driven cleanup; do not create backend-to-frontend coupling in this refactor. |
| Reference Pack draft suggests a possible future normalization registry. | Draft NLSpec only; current Core assigns canonical identity to Indicators. | medium | `defer` | Indicators until an adopted owner changes it | Do not move normalization based on draft material. |
| No direct saved-view, object-store, grid-vendor, or WebSocket transport implementation exists in the target. | Target import/source scan and live caller inspection. | low | `intentional/no_action` | Existing Saved Views, platform, grid adapter, Collaboration/Revisions owners | Freeze the affected contracts; add no Indicator-specific substitute. |
| Generic projection composition is exposed through historically Timeline-named runtime state. | `server/runtime.go` and Indicator rebuild integration path. | medium | `defer` | Application/projection assembly | Treat as composition naming debt, not evidence that Timeline owns Indicators. |
| Existing boundary guard does not prove the proposed transport/projection seams. | `boundary_guard_test.go` and backend-boundary manifest. | medium | `should_fix` | Indicators and backend boundary policy | Strengthen guards in the same slice that introduces each seam. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01, WF-02, WF-04 | Fix the repository baseline, authority order, allowed write, and handoff state. | This tracker only. | `make lint-markdown` for the planning artifact. | Baseline and source posture recorded; planning task makes no implementation change. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Account for every target file, caller, dependency, and test surface. | `internal/modules/indicators/**` plus live callers. | Re-run file count/search before implementation. | All 26 current files have one Section 2 row. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Bind each affected behavior to its adopted owner and current contract projection; route `IND-PORT-001` and `IND-PORT-002` to Core 01/Core 04 rather than this tracker. | Core/adopted specs, view/import/projection/portability/recovery contracts. | `make generate-drift`; `make json-shape-check` when implementation changes authored inputs. | Every contract in Section 4 has an owner and test posture; missing portability authority remains an explicit owner gate. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Distinguish conformant baseline behavior from current owner mismatches and missing evidence. | Indicator tests, Workbook scenarios, Revisions and Network Flow tests, harness owner inputs. | Focused service-backed rows and later added characterization rows. | SL-00 covers the freeze map without blessing `interactive_cell` or the proposed portability partition as current behavior. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Locate transport, persistence, projection, frontend, generated, and harness coupling. | Indicator root, provider subpackages, assemblies, boundary manifests, generic frontend consumers. | `make backend-module-boundary-check`; conditional frontend boundary check. | Findings classified exactly as Section 5. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define a stable Indicator facade, internal identity/source seams, and owner contribution boundaries. | Indicators root/internal packages, Workbook contribution, projection assembly. | Focused owner rows plus boundary checks. | No public route/view/participant change and no undecided ownership remains. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order the smallest reversible slices, isolate behavior corrections, and enforce the preflight and owner-adoption gates for SL-06 and SL-07. | Packages named in Section 7. | Per-slice commands in Section 7. | Every slice has dependency, rollback, binary completion criteria, and an authorization posture. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-06 | WF-08 | Keep owner evidence aligned when tests or package seams move. | Authored verification/test-family inputs and affected tests. | `make agent-finalize`, generation/policy checks, focused owner slice. | All new/moved tests are accounted for without using the harness as architecture. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow-to-broad verification, record artifacts/failures, and hand off only a slice whose entry gates are satisfied. | Tracker plus implementation files authorized in the later task. | Section 8 sequence. | Clean scoped diff, completed checks recorded, `IND-ORIGIN-*` and `IND-PORT-*` evidence traceable, remaining blockers explicit. |

WF-00 through WF-08 planning is complete in this tracker. SL-06 product direction is authorized but remains future implementation work. SL-07 cannot enter implementation until the Core 01/Core 04 owner closure is adopted and a later task authorizes the slice. Executable characterization, owner-input changes, production changes, and product validation runs remain future work.

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | none | Add owner-mapped baseline characterization for create/replay/conflict/auth, collaboration intent, import, participant rollback, projection query/rebuild parity, portability, and rollback. Do not characterize `interactive_cell` or the proposed portability partition as current valid behavior. | Indicator tests; Workbook scenario tests; Revisions/Network Flow collaborator tests; authored test-family inputs. | Tests could accidentally bless current owner drift or assert implementation-private SQL. | Preserve all current focused tests; add public/facade assertions and one create `record_changed` assertion. | `make service-backed-test-slice OWNER=module.indicators`; `make test-fast` | Revert the characterization commit; no production state is involved. | Required conformant behavior fails before intentional regressions, owner gaps remain explicit, and all added rows are owner-accounted. |
| SL-01 | SL-00 | Use Workbook-owned JSON/view-schema admission and translate to a transport-neutral Indicator create command; keep Indicator semantic validation. | Indicator `api.go`; Workbook mutation contribution/admission. | Request shape, error details, normalized request hash, replay status, and row envelope. | Closed-body, minimum-signal, hash equivalence, first/replay/divergent route tests. | Focused Indicator integration row; `make test-fast`; `make browser-e2e-webserver-backed` | Revert the adapter commit as one unit, restoring the old decoder. | Indicator root no longer constructs HTTP API errors; all frozen route tests pass byte/shape-equivalent assertions. |
| SL-02 | SL-00 | Centralize canonicalization, type/value/hash validation, and dedupe in one Indicator-internal component used by live create, import, Network Flow, portability validation, and rollback. This slice MUST precede SL-06 and SL-07. | Indicator root/internal identity package; rollback provider; import/participant adapters. | Canonical values, historical rollback acceptance, duplicate identity, IP exactness, concurrent reuse. | Preserve all IP tests; add non-IP registry, live/rollback parity, import/participant parity, and concurrency cases. | `make service-backed-test-slice OWNER=module.indicators`; `make backend-module-boundary-check`; `make test-fast` | Revert the identity-component commit; no schema migration is allowed. | One semantic implementation remains and every caller produces the same conformant identity/dedupe result. |
| SL-03 | SL-01, SL-02 | Retain the public internal `Store` compatibility facade while separating transaction/idempotency/revision coordination from fixed Indicator, observation, and lifecycle repositories. | Indicator Store/facade and new owner-internal source repository files; app construction only if dependencies change. | Transaction boundaries, lock order, idempotency, revisions, incident-open checks, partial commits. | Fault-boundary rollback, exact replay, observation resolution, lifecycle append, and collaboration intent counts. | `make service-backed-test-slice OWNER=module.indicators`; `make backend-module-boundary-check`; `make test-fast` | Revert the facade/repository commit; do not combine with projection cleanup. | Public internal Store methods remain compatible, repository methods never own commits, and focused/broad tests pass. |
| SL-04 | SL-00, SL-03 | Characterize the generic query route, move row-refresh derivation behind the Indicator projection contribution/Projections port, align the authored provider capability if required, and retire `query.go` only after equivalent generic keyset coverage exists. | Indicator projection provider; Projections/provider assembly; Indicator facade; authored provider contract if capability changes. | Same-transaction projection refresh, query sorting/filtering/cursors, row shape, rebuild parity, restore ordering. | Move keyset test to production generic path; preserve route query, query/load parity, incident rebuild, and restore rebuild tests. | `make service-backed-test-slice OWNER=module.indicators`; `make generate-drift`; `make json-shape-check`; `make test-fast` | Revert the projection/query commit, including authored input and regenerated outputs, as one unit. | No production caller or test relies on target-local query code; provider/catalog drift is zero and query/rebuild results match baseline. |
| SL-05 | SL-02, SL-03 | Keep portability, recovery, delete/restore, rollback, and revision contributions under Indicators while narrowing exported surfaces and strengthening boundary guards. | Indicator contribution/provider packages; revision/recovery/portability assemblies; backend boundary inputs. | Provider catalog completeness, delete/restore, child rollback, recovery and bundle ordering. | Preserve Revisions delete/restore and child rollback integration; add catalog construction and forbidden-import cases. | `make backend-module-boundary-check`; `make service-backed-test-slice OWNER=module.indicators`; `make test-fast` | Revert the provider-encapsulation commit and matching boundary input together. | Assemblies import only approved contributions, all catalogs validate, and no source semantics move into a coordinator. |
| SL-06 | SL-00, SL-02, IND-015; authorization recorded in Section 4.1 | **Authorized direction; requires a later implementation task:** implement `IND-ORIGIN-001` through `IND-ORIGIN-005`, reject non-Core origins, and replace invalid fixtures according to their actual producers. | Indicator-owned origin boundary; ordinary create/import/extraction/rollback/portability adapters; affected tests/fixtures; migration inputs only if IND-015 requires a separately authorized migration. | Accepted behavior changes for invalid callers; unproven persisted provenance; partial mutation; error-envelope drift. | `AC-ORIGIN-001` through `AC-ORIGIN-006`; preserve source binding, raw text/location, history, and repeated provenance. | `make service-backed-test-slice OWNER=module.indicators`; `make backend-module-boundary-check`; `make test-fast`; drift checks if authored inputs change | Revert the complete behavior-correction commit as one unit. Revert any separately authorized migration in its own reviewed rollback path; do not introduce a compatibility alias. | IND-015 records an admissible preflight outcome; every origin acceptance case passes; exact invalid inputs fail before mutation; public adapters retain their envelopes; history and portability round trips preserve exact tokens; focused evidence and drift checks pass. |
| SL-07 | SL-00, SL-02, SL-05, IND-016; later implementation authorization | **BLOCKED:** after `IND-PORT-001` and `IND-PORT-002` are adopted and the slice is authorized, enforce `IND-PORT-003` through `IND-PORT-006` before Incident Bundle publication. | Adopted Core 01/Core 04 owner text first; then Indicator source port/identity component, Incident Bundle tests, and authored harness inputs. | Code could invent product behavior; previously accepted malformed bundles may fail; invariant selection, error safety, and final-transaction atomicity may drift. | `AC-PORT-001` through `AC-PORT-007`, including ten exact negative fixtures, precedence, v1/v2 round trips, repeated/tombstone preservation, failure injection, and safe errors. | `make service-backed-test-slice OWNER=module.indicators`; `make backend-module-boundary-check`; `make test-fast`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Revert the behavior-correction commit as one unit; never weaken the descriptor or add repair, skip, merge, compatibility, or warning-only behavior. | Owner amendments are adopted; authorization is recorded; every acceptance case passes; exact safe failures are deterministic; final-transaction failure publishes no state; focused evidence and drift checks pass. |
| SL-08 | SL-01 through SL-06 when included; SL-07 only after IND-016 and authorization | Update authored verification/test-family inputs for moved/new evidence, regenerate through Make, run narrow-to-broad checks, and append the final implementation handoff. | Authored verification/harness inputs, tracker, generated outputs produced by Make only. | Missing owner accounting, stale generated topology, or overclaiming validation. | All moved/new tests, existing three owner rows, and the applicable `AC-ORIGIN-*` or `AC-PORT-*` matrix. | `make agent-finalize`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make test-fast`; `make check` | Revert only authored accounting plus its generated projections if the implementation slices remain independently green. | Clean scoped diff, all applicable gates recorded with run roots, no hand-edited generated file, and current handoff log complete. |

Every slice MUST be one reviewable, revertible commit unless an authorized data migration requires its own commit and rollback path. Public contracts named in Section 4 MUST remain unchanged unless the active remediation plan explicitly corrects them. A slice that exposes an unlisted product behavior change MUST stop and be reclassified `requires later authorization` before implementation.

## 8. Validation Plan

The commands below were discovered from the live Make task surface and `make task-guide ROLE=module-author OWNER=module.indicators`. The implementation baseline in Section 1 records the product validation already executed for this authorized run.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make service-backed-test-slice OWNER=module.indicators ROWS=module.indicators.store.indicator_storage_keeps_source_bound_observation_07c064324e,module.indicators.store.indicators_persist_as_canonical_indicator_record_3998750e91` | Two mapped Indicator store rows using Postgres fixtures. | yes | Despite `_Unit` test names, the harness classifies them as service-backed integration evidence. |
| integration | `make service-backed-test-slice OWNER=module.indicators ROWS=module.indicators.integration.indicator_create_and_query_routes_persist_canoni_e3a29b7aaf` | Real Indicator HTTP create/query and projection rebuild row. | yes | Run the whole owner with `make service-backed-test-slice OWNER=module.indicators` after each cross-cutting slice. |
| e2e/browser | `make browser-e2e-webserver-backed` | Generic workbook system-view selection, Indicator create/query, saved-view continuity, and keyboard/grid anchors. | no | Required after Workbook/public wiring changes, not for an isolated backend-only move. |
| generated drift | `make generate-drift && make generated-artifact-policy-check && make json-shape-check` | Generated view/protocol artifacts, policy, and authored JSON contracts/manifests. | yes | Run before to establish baseline and after any authored contract/harness input change. Never hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module import/SQL boundary policy. | yes | Add `make frontend-import-boundary-check` only if a later authorized slice changes frontend ownership files. |
| full check | `make agent-finalize` followed by `make check` | Harness maintenance and full developer verification gate. | no | Run at final implementation handoff; use `make test-fast` as the intermediate broad loop. |
| documentation | `make lint-markdown` | Authored Markdown, including this tracker. | yes | The only executable validation required for this planning-only change. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| IND-001 | Establish source posture and repository baseline | WF-00 | DONE | none | Section 1; branch/commit/time and authority list | Baseline and one-file write scope are explicit. |
| IND-002 | Inventory all target files and live callers | WF-01 | DONE | IND-001 | Section 2 with 26 rows | Every current target file is accounted for. |
| IND-003 | Map public contracts and owner boundaries | WF-02 | DONE | IND-002 | Sections 3 and 4 | Every discovered contract has an owner and test posture. |
| IND-004 | Identify characterization gaps without blessing owner drift | WF-03 | DONE | IND-003 | SL-00 and Sections 4/5 | Required baseline tests and invalid fixtures are distinguished. |
| IND-005 | Classify coupling and target owners | WF-04 | DONE | IND-002 | Section 5 | Each finding has one allowed classification and planning action. |
| IND-006 | Define the facade/repository and transport seams | WF-05 | TODO | IND-004, IND-005 | SL-01 through SL-03 | Later implementation preserves the Store facade and frozen wire behavior. |
| IND-007 | Consolidate projection/query ownership | WF-05 | TODO | IND-004, IND-006 | SL-04 | Generic query and provider refresh are the only production paths with parity coverage. |
| IND-008 | Encapsulate source-owner provider contributions | WF-05 | TODO | IND-006 | SL-05 | Catalogs and boundary guards pass with narrower exported surfaces. |
| IND-009 | Correct observation origin vocabulary | WF-06 | TODO | IND-004, IND-015, SL-02; product direction authorized, later implementation task required | SL-06; `IND-ORIGIN-*`; `AC-ORIGIN-*`; RB-001 | Exact Core 02 tokens are enforced before mutation, public adapters remain compatible, all acceptance cases pass, and retained evidence is recorded. |
| IND-010 | Close all declared portability invariants | WF-06 | BLOCKED | IND-004, IND-008, IND-016; later implementation authorization | SL-07; `IND-PORT-*`; `AC-PORT-*`; RB-002 | Adopted owner text exists and authorized validation proves all ten invariants, deterministic safe failure, and final-transaction atomicity. |
| IND-011 | Replace frontend raw Indicator timestamp inference with metadata | WF-04 | DEFERRED | Later Workbook/View Contracts owner task | Section 5 finding | Adopted frontend owner chooses and verifies a metadata-driven design. |
| IND-012 | Move normalization to a shared Reference Data registry | WF-04 | DEFERRED | Later adopted owner specification | Draft Reference Pack NLSpec evidence | An adopted owner explicitly changes current Core ownership. |
| IND-013 | Update harness accounting for new/moved tests | WF-07 | TODO | IND-006 through IND-009 as implemented; IND-010 only after unblocked and included | SL-08 and authored owner manifests | Focused owner routing covers new evidence and generated topology is clean. |
| IND-014 | Run final validation and append implementation handoff | WF-08 | TODO | IND-007, IND-008, IND-013, plus authorized corrections if included | Section 8 commands and future run roots | Clean scoped diff and all required results/blockers recorded. |
| IND-015 | Complete SL-06 persisted-data preflight | WF-06 | TODO | Later implementation task with access to supported deployment and retained-fixture state | `IND-ORIGIN-004`; Section 4.1 decision table | Exactly one preflight outcome is recorded; any migration/remediation dependency has separate authority before SL-06 enablement. |
| IND-016 | Adopt the Indicator portability owner closure | WF-02 | BLOCKED | Later owner-document task for Core 01, Core 04, and traceability inputs | `IND-PORT-001`, `IND-PORT-002`; proposed Section 4.2 tables | Exact rows/order, exclusive invariants, deterministic precedence, phase allocation, and binary acceptance criteria are adopted without using this tracker as authority. |
| IND-017 | Bootstrap the authorized implementation ledger | WF-00 | DONE | User implementation authorization | Section 1 active execution protocol and baseline run roots | Authorization, starting state, sequence, checkpoint rule, and baseline evidence are explicit. |
| IND-018 | Complete SL-00 characterization and owner accounting | WF-03 | DONE | IND-017 | Two new focused rows, corrected test origins, generated topology, and recorded run roots | Identity and valid portability surfaces are executable regression locks; focused, broad, and drift gates pass. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | Authority and planning-only scope fixed at the requested repository baseline. | Inspected planning framework, domain, Core 00-04, adopted NLSpecs, and live repository; touched only this tracker. | Read-only `sed`, `rg`, `find`, `jq`, and Git inspection; `make task-guide`; `make help`; `make help-all`; `make lint-markdown`. | No owner contradiction; Core 05 not applicable; tracker Markdown validation passed. An initial `sed` used a misnamed Core 01 path, failed without mutation, and was corrected to `01_architecture_storage_and_view_contracts.md`. | None for planning. | Begin SL-00 only in a later authorized implementation task. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | The tracker distinguishes repository facts, owner requirements, authorized implementation requirements, and non-authoritative owner proposals. | Inspected `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, this tracker, and repository status; touched only this tracker. | Read-only `wc`, `sed`, `rg`, `find`, date, and Git inspection; `apply_patch`; `make lint-markdown`. | SL-06 direction recorded as authorized; SL-07 remains owner-blocked; Markdown validation passed at `.cartulary/test-results/20260803T133511Z-p3857129`; no owner or product file changed. | RB-002; RB-001 retains a data-preflight gate. | In a later implementation task, run SL-00/SL-02 and IND-015 before SL-06. |
| 2026-08-03T10:04:36-04:00 | Codex / implementation session | Complete remediation authorized; tracker converted to the active execution ledger. | Updated this tracker only and preserved the clean production baseline. | Git inspection, baseline evidence review, `apply_patch`, and `make lint-markdown`. | Starting commit, dirty state, execution order, per-workstream commit gate, and four passing baseline run roots recorded; Markdown passed at `.cartulary/test-results/20260803T140517Z-p3878050`. | None for SL-00. | Commit tracker bootstrap, then begin SL-00 characterization. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | Indicators confirmed as a legitimate source owner with a mixed root Store, duplicate query path, and provider seams. | Inspected all 26 target files, assemblies, callers, schema ownership, and backend boundary manifest; touched only this tracker. | `find internal/modules/indicators -type f`; targeted `rg`; direct `sed`; `make explain-target` for relevant targets. | Every target file inventoried; structural slices SL-01 through SL-05 are sequenced. | Behavior corrections remain separately blocked by RB-001/RB-002. | Add conformant characterization before moving code. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | The structural diagnosis is unchanged; SL-02 is now an explicit prerequisite for both conformance corrections. | Rechecked the 26-file target count and revised only this tracker. | `find internal/modules/indicators -type f`; `rg`; Git inspection; `apply_patch`; `make lint-markdown`. | Origin validation and portability interfaces are specified as transport-neutral source-owner boundaries without moving code. | SL-07 cannot proceed from tracker language; IND-016 is required. | Preserve SL-00 through SL-05 order; complete IND-015 before SL-06. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | No Indicator-specific frontend controller or grid-vendor layer exists; generic Workbook/View Contracts own the surface. | Inspected generic view-contract, selector, query-model, Workbook unit, and browser Indicator cases; touched only this tracker. | Targeted `rg` and direct `sed`. | Public selector/view identities are frozen; raw timestamp-field inference is deferred to its owner. | No backend blocker; frontend cleanup is intentionally deferred. | Run frontend boundary/browser checks only if a later slice touches frontend ownership or public wiring. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | Frontend ownership and frozen selector/view behavior remain unchanged. | Reused the prior inspected frontend evidence; touched only this tracker. | Tracker reads, `apply_patch`, and `make lint-markdown`; no frontend command ran. | No Indicator-specific controller, saved-view owner, or grid-vendor layer is proposed. | None for this tracker revision. | Run frontend boundary/browser checks only when a later slice changes frontend ownership or visible Workbook wiring. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | View, import, projection, portability, recovery, and generated-consumer risks mapped. | Inspected authored contracts/manifests listed in Section 1; touched only this tracker. | `jq`, `rg`, direct reads, and `make explain-target` for drift targets. | Generated roots identified as downstream-only; provider/source ownership is explicit. | Full portability validation requires RB-002 authorization. | Change authored inputs only when a later slice requires it, then regenerate through Make. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | The ten portability IDs are mapped, but the exact portable shape, partition, precedence, and Core 04 criteria remain missing owner authority. | Inspected analysis recommendations and current tracker evidence; touched only this tracker, not Core or authored contracts. | `sed`, `rg`, Git inspection, `apply_patch`, and `make lint-markdown`; no generation ran. | Proposed owner language is quarantined from implementation requirements; generated consumers remain downstream-only. | `BLOCKED: owner definition incomplete` under RB-002/IND-016. | Adopt Core 01/Core 04 and authored traceability before authorizing SL-07. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | Focused owner has three service-backed rows; broader tests cover additional behavior without focused Indicator mapping. | Inspected Indicator tests, cross-owner tests, owner verification contract, and test-family manifest; touched only this tracker. | `make task-guide ROLE=module-author OWNER=module.indicators`; `make explain-test-owner OWNER=module.indicators`; `jq`; targeted searches; `make lint-markdown`. | Exact focused commands and row IDs recorded in Section 8. Tracker Markdown validation passed; no product tests, generation, or product validation ran. | New executable characterization remains future work. | Author SL-00 tests and rows before structural implementation. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | Binary `AC-ORIGIN-*` and `AC-PORT-*` matrices now define future evidence without changing harness routing. | Inspected analysis test recommendations and existing tracker test posture; touched only this tracker. | Read-only document inspection, `apply_patch`, and `make lint-markdown`. | Markdown validation passed at `.cartulary/test-results/20260803T133511Z-p3857129`; no product test, harness generation, drift check, or retained product run executed. | Executable tests remain future work; portability tests await IND-016. | Add owner-mapped SL-00 evidence, then implement only slices whose gates are satisfied. |
| 2026-08-03T10:13:53-04:00 | Codex / SL-00 implementation | Identity and portable-row characterization are owner-accounted and invalid Indicator fixtures no longer bless `interactive_cell`. | Added identity and portability characterization tests; updated three existing Indicator fixture files, the Indicator test-family manifest, generated topology, and this tracker. | `make format`; `make test-catalog-check`; focused `make test-slice`; `make service-backed-test-slice OWNER=module.indicators`; `make test-fast`; `make generate`; drift/policy/shape checks; `make lint-markdown`. | Identity row passed at `.cartulary/test-results/20260803T141034Z-p3887120`; six service-backed units passed at `.cartulary/test-results/20260803T141040Z-p3887800`; 346 fast units passed at `.cartulary/test-results/20260803T141054Z-p3889600`; final drift/policy/shape checks passed at `.cartulary/test-results/20260803T141339Z-p3942258`, `p3942260`, and `p3942262`; Markdown passed at `.cartulary/test-results/20260803T141444Z-p3945638`. The first identity run failed because the test guessed the current missing-hash field attribution; the observed behavior was characterized and the rerun passed. The first drift run correctly reported the new harness-row digest and passed after `make generate`. | None. | Commit SL-00, then run IND-015 persisted-origin preflight. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | Generic Workbook route re-derives membership/role; Indicator Store enforces incident-open and route idempotency; Revisions appends collaboration intent. | Inspected Core 04 authorization, Workbook routes, Indicator Store, Revisions appender, and generic route scenario inventory; touched only this tracker. | Direct reads and targeted `rg`. | Create/query authorization, CSRF, replay ordering, attribution, and event semantics are in the freeze map. | None for structural planning. | Add focused create auth/replay/event characterization in SL-00. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | SL-06 now requires exact pre-write origin rejection and trusted context for `system`; proposed SL-07 requires safe deterministic errors and final-transaction atomicity. | Inspected analysis security/atomicity recommendations and existing owner posture; touched only this tracker. | `sed`, `rg`, `apply_patch`, and `make lint-markdown`; no security or product test ran. | Authorization ownership and public envelopes remain frozen; hostile data and internal details are explicitly excluded from proposed portability errors. | SL-06 needs IND-015; SL-07 needs IND-016 and later authorization. | Prove no partial mutation/publication in the applicable future acceptance matrices. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-03T08:32:42-04:00 | Codex / tracker-authoring session | Decision-complete slice order recorded; no production refactor performed. | Touched only this tracker. | Read-only discovery commands listed above; `make lint-markdown`. | Markdown validation passed; structural work can begin at SL-00 without rediscovery. | RB-001 and RB-002 require later behavior-change authorization. | In a later authorized task, re-check baseline, run Section 8 pre-implementation commands, and implement SL-00 only. |
| 2026-08-03T09:28:27-04:00 | Codex / NLSpec tracker-revision session | The tracker is decision-complete for SL-06 entry/acceptance and for the owner work that must precede SL-07; no production refactor occurred. | Touched only this tracker. | Read-only repository/document inspection, `apply_patch`, and `make lint-markdown`. | SL-06 direction is authorized but unimplemented; RB-001 is READY for preflight; RB-002 remains BLOCKED; Markdown validation passed at `.cartulary/test-results/20260803T133511Z-p3857129`. | IND-015 may expose a separate migration/remediation dependency; IND-016 is mandatory before SL-07. | Next authorized session starts with SL-00/SL-02 and IND-015, not SL-06 enablement or SL-07 code. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | SL-06 product direction is authorized; the persisted-data preflight and implementation evidence remain entry and closure gates. | Exact rejection cannot be enabled safely until supported fixtures, bundles, upgrade data, and deployments are classified without guessing provenance. | Complete IND-015 using the Section 4.1 decision table; obtain separate migration/remediation authority if required; then implement `IND-ORIGIN-*` and pass `AC-ORIGIN-*` in a later task. | READY — authorization recorded; preflight and implementation outstanding |
| RB-002 | `BLOCKED: owner definition incomplete` for the Indicator Incident Bundle portable row contract, exclusive invariant partition, precedence, phase allocation, and complete acceptance criteria. | The source port must emit an exact deterministic public `invariant_id`; code or this tracker cannot safely invent missing owner behavior. | Adopt IND-016 in Core 01/Core 04 and traceability inputs, then obtain later SL-07 implementation authorization and pass `AC-PORT-*`. | BLOCKED |

No `BLOCKED: owner contradiction` entry is present because no conflict between adopted owners was found.

## 12. Binary Completion Criteria

### 12.1 Tracker-revision completion

- [x] Every file in `internal/modules/indicators` is inventoried; all 28 current files have an individual Section 2 row and none is silently excluded.
- [x] Every discovered public contract risk has a current owner, evidence reference, existing-test posture, and required characterization posture in Section 4.
- [x] Every workflow has previous/subsequent dependencies, a goal, validation, and a handoff checkpoint.
- [x] Current-state findings, adopted-owner requirements, authorized implementation requirements, proposed owner language, defaults, and intentional implementation freedom are unambiguously distinguished.
- [x] `IND-ORIGIN-001` through `IND-ORIGIN-005` define the exact vocabulary, producer mapping, interface boundary, compatibility posture, preflight, atomicity, and closure evidence; SL-06 authorization is recorded without claiming implementation.
- [x] `IND-PORT-001` through `IND-PORT-006` define the required owner-closure path, proposed invariant partition, interface allocation, deterministic safe failure, atomicity, and evidence while explicitly remaining non-authoritative until adoption.
- [x] Every implementation slice is behavior-preserving unless an authorized correction is explicit; SL-06 is authorized but unimplemented, and SL-07 remains owner-blocked and unauthorized.
- [x] Canonical Make-owned validation commands are discovered and recorded; no command is claimed to have passed unless it was actually executed successfully.
- [x] No owner contradiction was found; future contradictions must be recorded as `BLOCKED: owner contradiction` without choosing a side.
- [x] Repository/framework mismatches are recorded: the live generic query path coexists with a redundant target-local query, and generic projection composition is exposed through a historically Timeline-named runtime bundle.
- [x] The seven handoff areas contain a current session row with files, commands, result, blockers, and next action.
- [x] This task changed only the tracker and performed no production refactor.

### 12.2 Future implementation closure

- [x] SL-00 owner-conformant characterization is complete and does not bless `interactive_cell` or proposed portability language as current behavior.
- [ ] SL-02 provides one Indicator identity implementation shared by live create, import, Network Flow, rollback, and portability consumers.
- [ ] IND-015 records exactly one admissible persisted-data outcome and every required migration or remediation has separate authority.
- [ ] SL-06 is implemented in a later authorized repository-change task and all `AC-ORIGIN-*` cases pass with current harness evidence.
- [ ] IND-016 is adopted by Core 01/Core 04 and its authored traceability inputs before any SL-07 implementation begins.
- [ ] SL-07 receives later implementation authorization and all `AC-PORT-*` cases pass with deterministic safe errors, retained v1/v2 compatibility, and final-transaction atomicity.
- [ ] Every affected authored contract or harness input is regenerated through Make; generated roots contain no hand edits and all required drift gates pass.
- [ ] The applicable focused, boundary, broad, finalize, and full checks pass and their run roots are recorded in a later implementation handoff.
