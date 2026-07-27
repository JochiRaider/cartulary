# workbook Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current posture |
| --- | --- |
| Target path | `internal/modules/workbook` |
| Target label | `workbook`, derived from the final path segment and normalized to lowercase kebab case |
| Output path | `docs/handoffs/workbook-module-refactor-tracker.md` |
| Repository baseline | Clean `main` at `8e4639d45c6b047540512eb38af1700fb20c5140` |
| Mode | Planning and documentation only |
| Allowed change | This tracker file only |
| Non-goals | No production code, test, contract, generated artifact, package configuration, migration, SQL, harness, or frontend change |
| Later authority | Any characterization-test, refactor, contract, generated, or harness change requires a later explicitly authorized implementation task |

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs for their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. The modular-refactor framework and prior handoffs as evidence only.

Core 05 is not applicable to this planning-only refactor because no timed,
fixture-sensitive, benchmark, or other claim-bearing publication is proposed.
No owner contradiction was found. If a contradiction appears during a later
implementation task, that task must record `BLOCKED: owner contradiction` and
must not choose a side.

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md`
- `docs/opentelemetry-instrumentation-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_implementation_testing_guide.md`

Repository evidence inspected includes every file listed in Section 2 and the
relevant application composition, projection catalog, boundary manifests,
schema ownership, authored OpenAPI, view-schema registry, WebSocket schema,
verification owner, test-family manifest, frontend workbook shell, frontend
import boundaries, and grid-adapter dependency declarations. Important
cross-target evidence includes:

- `internal/app/server/runtime.go`
- `internal/app/projectionassembly/catalog.go`
- `internal/modules/projections/provider_boundary.go`
- `internal/modules/codegenboundary/sqlc_boundary_test.go`
- `internal/modules/imports/boundary_test.go`
- `contracts/openapi-source/owners/module.workbook/openapi.json`
- `contracts/openapi-source/owners/module.timeline/openapi.json`
- `contracts/projection-providers/index.json`
- `contracts/view-schemas/index.json`
- `contracts/ws/index.schema.json`
- `contracts/verification/owners/module.workbook.json`
- `tools/backend_module_boundaries.json`
- `tools/frontend_import_boundaries.json`
- `tools/schema_object_ownership_manifest.json`
- `tools/test_catalog_owner.json`
- `tools/test_families/module.workbook.json`
- `tools/test_support_inventory.json`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/models/workbookStartup.ts`
- `apps/web/src/services/workbookApi.ts`
- `packages/grid-adapter/package.json`

The planning framework is doctrine and a reusable template, not evidence that
`workbook` is a durable backend module. The live tree contains 56 files and is
materially broader than a thin facade. This mismatch is a planning finding:
the permanent boundary must be determined from the current responsibilities
and owner contracts, not from the existing directory name.

## 2. Current-State Repository Inventory

All 56 files under `internal/modules/workbook` are in scope for inventory. None
is excluded.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/workbook/batch_admission.go` | Strict Timeline clipboard and bulk request admission, hashing, and Tabular Ingest planning | `TimelineClipboardPasteRequest`, `TimelineBatchTarget`, `TimelineBulkMutationRequest`, decode, plan, and hash functions | Workbook route handlers and clipboard/bulk tests | `tabularingest`, `timeline`, `fieldnorm`, `httpapi`, `viewschema` | Clipboard unit and integration tests | Consumes Timeline view-schema metadata; no generated write | Thin workbook admission facade, with Timeline and Tabular Ingest ports | high | Clipboard paste is distinct from file imports; keep that separation |
| `internal/modules/workbook/boundary_guard_test.go` | Verifies workbook import rules are represented in the backend boundary manifest | Test-local manifest types and `TestWorkbookBoundaryRulesAreManifestBacked` | Go test discovery | `tools/backend_module_boundaries.json` via file read | Self | Asserts authored boundary manifest; no generated write | Workbook boundary verification | medium | Must move or change only with authored boundary policy |
| `internal/modules/workbook/catalog_evidence_sentinel_test.go` | Sentinel coverage for clipboard and coordination catalog evidence | Four sentinel tests | Go test discovery and test catalog selectors | Testing package only | Self | Coupled to semantic test-family rows | Workbook verification accounting | low | Evidence sentinel, not runtime architecture |
| `internal/modules/workbook/clipboard_paste_integration_test.go` | Persists Timeline and entity-origin paste, bulk batches, conflicts, ordering, and cross-incident rejection | Four integration tests | Go test discovery | Workbook routes, Timeline, entities, projections, database harness | Self and shared scenario helpers | Supports workbook integration rows and view contracts | Workbook cross-owner admission verification | high | Preserve exact one-visible-batch and incident isolation behavior |
| `internal/modules/workbook/clipboard_paste_unit_test.go` | Characterizes quoted CSV/TSV parsing and visible action grouping | `TestSharedPasteAndBulkPlanningGroupsOneVisibleAction_Unit` | Go test discovery | Workbook batch admission and Tabular Ingest | Self | Supports a `module.workbook.unit` row | Workbook admission verification | medium | Direct evidence for the public paste planner |
| `internal/modules/workbook/conflict_route_test.go` | Characterizes optimistic concurrency, same-field conflicts, collection review, keep-saved, and conflict durability | Conflict tests plus reusable test-only conflict helpers | Go test discovery and sibling workbook tests | Workbook routes, revisions, projections, collaboration, scenario harness | Self and sibling conflict tests | Exercises OpenAPI errors, view schemas, and WebSocket effects | Workbook public conflict verification; source semantics belong revisions and source owners | high | Existing breadth makes conflict extraction risky |
| `internal/modules/workbook/coordination_surfaces_test.go` | Characterizes artifact-backed coordination defaults, validation, saved views, projections, collections, semantic filters, and duplicate coalescing | Coordination tests and test-only value/state helpers | Go test discovery | Artifacts, saved views, projections, workbook store, WebSocket test support | Self | Characterization reference in projection-provider catalog | Workbook surface integration verification with artifacts as source owner | high | UI-native surface does not make workbook the source owner |
| `internal/modules/workbook/entity_owner_adapter.go` | Adapts Host/Identity owner mutation results and errors to workbook DTOs | No exported symbol; package-private adapters | `mutation_store.go` | `entities/hostidentity` | Workbook entity create/patch and mutation tests | Maps view-contract DTO behavior; no generated write | Workbook anti-corruption adapter or entities-owned route adapter | medium | Split from generic mutation coordination when owner ports are introduced |
| `internal/modules/workbook/grouping_contract_test.go` | Verifies Timeline grouping is declared and presentation-only | `TestTimelineGroupingAndWorkbookPresentationOnly_Unit` | Go test discovery | `viewschema`, workbook query semantics | Self | Reads generated view-schema projection through `viewschema` | Workbook query-contract verification | medium | Does not grant runtime ownership to harness or grid code |
| `internal/modules/workbook/integration_test.go` | Exercises concurrent Timeline edit conflict resolution and `record_changed` delivery | Integration test plus test-only create, conflict, login, and assertion helpers | Go test discovery and sibling tests | Timeline facade, workbook routes, collaboration WebSocket, database harness | Self | Exercises WS and mutation contracts | Workbook route integration verification | high | Freezes cross-layer behavior before moves |
| `internal/modules/workbook/mutation_api.go` | Defines generic workbook mutation DTOs, strict decoders, field and collection policy, hashes, and stable error shapes | `MutationResult`, create/patch/conflict DTOs, validation errors, decode/hash/build functions | `routes.go`, `mutation_store.go`, external owner tests using workbook DTOs | Artifacts, entities, evidence, links, tasks/decisions, `fieldnorm`, `httpapi`, `viewschema` | Mutation decoder, timestamp, coordination, optional-surface, and integration tests | Consumes generated view-schema registry and OpenAPI shapes | Thin workbook admission contract plus owner-contributed policy adapters | high | Hard-coded owner and field policy is a drift hotspot |
| `internal/modules/workbook/mutation_api_test.go` | Strict-decoder and hash characterization for values, collections, IDs, relationship confidence, and nullability | Ten decoder/hash tests | Go test discovery | Workbook mutation API and owner contract helpers | Self | Verifies generated view-schema metadata is honored | Workbook admission verification with source-owner fixtures | high | Must remain green while policy is delegated |
| `internal/modules/workbook/mutation_store.go` | Dispatches create, patch, conflict, linked-note, and supersede operations to source owners; directly coordinates keep-saved conflict persistence | Exported methods on `Store`; package-private adapters | `routes.go`, workbook tests, and external owner tests constructing `Store` | Artifacts, linked notes, entities, evidence, links, parties, revisions, tasks/decisions, Timeline, auth, PostgreSQL | Mutation, conflict, coordination, optional-surface, parties, and owner tests | Produces public mutation rows and revision/WebSocket consequences | Split between source-owner facades, revisions conflict service, and thin workbook coordinator | critical | Direct transaction and idempotency work in `clearWorkbookConflict` is persistence-adjacent |
| `internal/modules/workbook/notes_indicators_test.go` | Characterizes Notes, Indicators, Assessments, Task/Decision, coordination projection queries, and hot projection rebuilds | Six integration tests and one test-only state type | Go test discovery | Multiple source owners, projections, workbook query store, database | Self | Exercises projection-provider descriptors and view schemas | Projection integration verification, partitioned by source owner | high | Several behaviors should move with their owner tests |
| `internal/modules/workbook/openapi_contract_test.go` | Verifies workbook record-mutation OpenAPI operations and schemas | `TestWorkbookOpenAPIRecordMutationContracts` | Go test discovery | Contract test facade and authored OpenAPI | Self | Directly verifies authored and generated OpenAPI surfaces | Workbook public contract verification | high | Keep with the public workbook facade |
| `internal/modules/workbook/optional_surfaces_store_test.go` | Characterizes optional artifact surface storage, projection query behavior, collection operations, and confidence bands | Three tests and test-only patch/state helpers | Go test discovery | Artifacts, projections, workbook store | Self | Reads optional view-schema contracts | Artifacts and workbook integration verification | high | Source semantics belong artifacts; route shape remains workbook-facing |
| `internal/modules/workbook/owner_setup_helpers_test.go` | Shared package-local setup and durable-state helpers for workbook tests | Test-only helper functions including `RecordVersion` | Sibling workbook tests | Owner stores, SQL fixtures, revisions, projections | Sibling workbook tests | No generated write | Workbook test support, split by semantic owner | medium | Keep only helpers specific to workbook route characterization |
| `internal/modules/workbook/paging.go` | Converts projection query results into keyset cursor positions and public pages | No exported symbol; package-private pagination helpers | `routes.go` | `pagination`, `querypage`, `viewschema` | Query pagination and row-wire tests | Implements Core query/cursor contract | Thin workbook query transport adapter | high | Legitimate facade concern if provider-neutral |
| `internal/modules/workbook/parties_integration_test.go` | Characterizes Party helper fields, raw text, stable links, and query persistence | Party integration test plus a test-only query helper | Go test discovery | Parties owner, workbook store, projections, database | Self | Projection characterization reference | Parties owner verification plus workbook route integration | high | Candidate to split between parties and workbook route tests |
| `internal/modules/workbook/query_store.go` | Routes Host, Identity, and Indicator queries to owner stores and other schemas to projection query service | `QueryStore`, `NewQueryStore`, `Query`; `Store.Query` delegation | `routes.go`, `store.go`, workbook and projection tests | Entities, indicators, projections, PostgreSQL, query/view contracts | Query, projection, pagination, and external projection tests | Named in projection import allowlist and provider manifest | Projections query catalog with a thin workbook query facade | high | Constructor creates owner stores from DB instead of receiving providers |
| `internal/modules/workbook/route_guard_characterization_test.go` | Verifies authentication, CSRF, incident state, role, record, and token checks happen before mutation | `TestWorkbookRouteGuardsFailBeforeMutation` plus test-only state helpers | Go test discovery | Workbook routes, auth, incidents, collaboration socket, database | Self | Freezes OpenAPI error and authorization outcomes | Workbook transport security verification | critical | Required before changing route decomposition |
| `internal/modules/workbook/routes.go` | Registers 13 workbook operations and implements authentication, authorization, decoding, owner dispatch, response envelopes, session sliding, and composition | `CreateRowHandlerFactory`, `RouteOption`, `WithTimelineOwner`, `WithProjectionQuery`, `WithCreateRowHandler`, `RegisterRoutes` | `internal/app/server/runtime.go`, application test support, route tests | Incidents, entities, indicators, projections, records, revisions, Timeline, startup, auth, HTTP/query/view platform packages | Nearly all route and integration tests | Binds authored `module.workbook` OpenAPI operations | Thin workbook HTTP/application facade after internal split | critical | Legitimate facade mixed with owner construction and detailed dispatch |
| `internal/modules/workbook/row_wire_test.go` | Characterizes full and sparse row wire families, WebSocket patches, cursor authorization, search, prefix, and null ordering | Seven tests | Go test discovery | Workbook query routes, projections, collaboration, pagination | Self | Exercises `view_row_v1`, WS, and view-query contracts | Workbook query and collaboration contract verification | critical | Public wire compatibility freeze |
| `internal/modules/workbook/service_ports.go` | Declares package-private query, mutation, record-target, Timeline, entity, indicator, and conflict-token ports | Package-private interfaces and type aliases | `routes.go`, `store.go` | Source owners, auth, query, view contracts | Compile-time use plus route/mutation tests | Mirrors public DTO and owner boundaries | Thin workbook facade ports, redesigned as provider contributions | high | Existing ports are useful seams but still owner-specific |
| `internal/modules/workbook/shared_harness_test.go` | Verifies workbook route inventory and conformance across envelopes, CSRF, replay, authorization, projection, and WebSocket effects | Two tests plus reusable test-only conformance helpers | Go test discovery and sibling tests | Workbook scenario/routetest support and multiple source-owner routes | Self and sibling conflict tests | Direct evidence for support integration rows | Workbook cross-owner route conformance verification | high | Inventory currently covers routes owned outside workbook |
| `internal/modules/workbook/startup/api.go` | Defines startup and preference DTOs, strict `sheet_ref` decoding, canonicalization, explicit selection parsing, and response resources | Startup constants, records, requests, decode/parse/build functions | Workbook routes, startup tests, incidents tests | `httpapi`, `viewschema` | Startup API, store, workbook startup, and incidents tests | Implements authored startup OpenAPI and view-schema identity | Workbook startup application/transport facade | high | Keep semantic selection contract independent of persistence |
| `internal/modules/workbook/startup/api_test.go` | Characterizes preference decoding, explicit selectors, extension workspace claims, visibility, and canonical form | Four tests | Go test discovery | Startup API and workspace registry | Self | Verifies startup OpenAPI payload semantics | Workbook startup verification | high | Retain with startup facade |
| `internal/modules/workbook/startup/bootstrap/bootstrap.go` | Inserts incident and user workbook preference bootstrap rows during incident creation using SQLc | `IncidentCreatePreferencesPort`, constructor, `PreferencesTx` | Server runtime, incidents store, records and merge test support | Generated SQL and PostgreSQL transaction types | Startup store, incidents, records, and merge tests | Imports `internal/gen/sql`; allowlisted by SQLc boundary | Incidents persistence/bootstrap port | high | Schema ownership manifest assigns preference tables to `incidents` |
| `internal/modules/workbook/startup/store.go` | Persists preferences, resolves startup candidates transactionally, clears invalid pointers, and delegates saved-view visibility | `ErrPreferencesNotFound`, `Store`, `NewStore`, validation/get/put/resolve methods | Workbook routes, startup and incidents tests | Generated SQL, `savedviews`, PostgreSQL, `viewschema`, `httpapi` | Startup store and workbook startup tests | Imports generated SQL and consumes saved-view/view-schema contracts | Split: incidents preference repository, savedviews resolver, workbook startup coordinator | critical | Direct SQL and multi-owner transaction are the strongest startup split seam |
| `internal/modules/workbook/startup/store_test.go` | Verifies bootstrap/upsert timestamps, structural no-ops, extension workspace round trips, claim loss, and pointer clearing | Two integration tests | Go test discovery | Incidents, auth fixtures, startup store, database | Self | Exercises generated SQL behavior and startup contracts | Split between incidents persistence and workbook startup verification | high | Preserve atomic clear and timestamp semantics |
| `internal/modules/workbook/startup/workspace_registry.go` | Resolves claimed extension workspaces and role-filtered availability | `WorkspaceResolver`, availability types, `WorkspaceRegistry`, constructor and methods | Startup store/API and server composition | Extension publication types from `httpapi` | Startup API/store and browser startup tests | Implements extension workspace availability contract | Workbook shell startup registry facade | medium | Does not own extension data stores |
| `internal/modules/workbook/store.go` | Aggregates all source-owner mutation/query facades, auth, record targets, conflict tokens, and projection rows | View-schema constants, `Store`, `NewStore` | Routes, app/test composition, workbook and external owner tests | Artifacts, entities, evidence, indicators, parties, projections, records, revisions, tasks/decisions, Timeline, auth, PostgreSQL | Broad workbook and owner test population | Named in projection import allowlist | Compatibility facade backed by application-injected providers | critical | Current constructor is an accidental composition root inside a module |
| `internal/modules/workbook/store_artifact_surface_test.go` | Verifies coordination artifact surfaces use contract-backed filters | `TestCoordinationArtifactSurfacesUseContractFilters` | Go test discovery | Artifacts, workbook query/store, view schemas | Self | Reads authored view-schema filters | Artifacts projection verification | medium | Candidate owner-local test after preserving route characterization |
| `internal/modules/workbook/telemetry.go` | Emits safe workbook query and mutation spans/metrics and maps errors to closed telemetry tokens | Package-private `Service` methods and helpers | Workbook route handlers | OTel, platform telemetry, owner errors, view registry | Telemetry tests and OTel conformance tool | Must match adopted OTel NLSpec and telemetry registry | Workbook facade instrumentation | high | Keep near the facade unless an application instrumentation seam is introduced |
| `internal/modules/workbook/telemetry_test.go` | Verifies safe telemetry mappings, API-error classification, and no-SDK behavior | Three tests | Go test discovery and OTel conformance source checks | Workbook telemetry and platform telemetry | Self | Named by OTel conformance script | Workbook telemetry verification | high | Move only with telemetry registry and NLSpec conformance |
| `internal/modules/workbook/testsupport/routetest/preferences.go` | Declares public/control preference route inventory entries | `PublicPreferences`, `ControlPreferences` | Incidents route tests | Timeline constants and shared route inventory | Incidents integration and HTTP conformance tests | Supports route accounting; no generated write | Split between workbook route support and shared route inventory | medium | Cross-owner use is intentional evidence, but ownership should be explicit |
| `internal/modules/workbook/testsupport/routetest/shared_harness.go` | Defines workbook shared-harness IDs, requirements, inventory, and validation | Harness ID and requirement types, inventory functions | Workbook and incidents shared harness tests | Standard library testing | Shared harness tests | Supports test-family accounting | Workbook test support | medium | Evidence map, not runtime architecture |
| `internal/modules/workbook/testsupport/scenariotest/conformance.go` | Executes route replay/history conformance and change-set attribution assertions | `RouteConformanceCase`, route/replay/assertion helpers | Workbook shared harness tests | Timeline test assertions, auth, audit and contract test utilities | Workbook shared harness | Supports integration evidence rows | Split between workbook route conformance and shared test infrastructure | high | Some helpers are owner-neutral |
| `internal/modules/workbook/testsupport/scenariotest/contracts.go` | Exposes allowed field keys and sorted row-key assertions | `AllowedFieldKeys`, `SortedRowFieldKeys` | Workbook and other owner scenario tests | View-schema test support | Cross-owner scenario tests | Reads generated view-schema contracts through an approved facade | Shared contract test support | medium | Candidate for `internal/testutil` or platform view-schema test support |
| `internal/modules/workbook/testsupport/scenariotest/harness.go` | Starts full app/server and runtime fixtures with PostgreSQL and object store services | `ServerHarness`, `RuntimeHarness`, start and assertion helpers | Many workbook, entities, evidence, revisions, recovery, saved-view, and platform tests | App config, platform services, `internal/testutil` service harnesses | Broad cross-owner integration population | No generated write; participates in test accounting | Shared application test composition under `internal/testutil/appsupport` | critical | Broad cross-owner callers show this is not workbook-local support |
| `internal/modules/workbook/testsupport/scenariotest/route_inventory.go` | Defines a large cross-owner route matrix, expected envelopes, replay, authorization, projection, and WebSocket effects | Route enums, context, entry types, validation and inventory functions | Workbook shared harness tests | Records fixtures and golden contracts | Shared harness and conformance tests | Supplies semantic evidence metadata | Split into owner-local route inventories plus workbook composition checks | critical | Includes entities, evidence, merge, and other non-workbook routes |
| `internal/modules/workbook/testsupport/scenariotest/runtime.go` | Provides login, incident creation, DB seeding, lookup, HTTP, and response helpers | Numerous exported scenario/runtime helpers | Many modules outside workbook | Incidents, records fixtures, startup bootstrap, auth, PostgreSQL, HTTP test utilities | Broad cross-owner integration population | No generated write | Shared test infrastructure with owner-local fixture extensions | critical | Test-only assumptions are not in production, but ownership is too broad |
| `internal/modules/workbook/testsupport/scenariotest/views.go` | Queries workbook views and asserts collection-item shapes | Exported query and collection helpers | Workbook, entities, evidence, links, revisions, and saved-view tests | HTTP test utility | Cross-owner scenario tests | Exercises view-row contracts | Workbook-specific view test support | medium | Can remain owner-local if callers deliberately depend on workbook surface behavior |
| `internal/modules/workbook/testsupport/scenariotest/ws.go` | Connects to incident WebSocket and asserts `record_changed` or silence | `RecordChangeSocketPayload`, connect and assertion helpers | Workbook and other owner integration tests | Platform WS and shared WebSocket test utility | Cross-owner scenario tests | Exercises `contracts/ws/index.schema.json` | Collaboration test support or workbook-specific effect facade | high | Generic socket lifecycle belongs collaboration/shared support |
| `internal/modules/workbook/timelineadmission/api_errors.go` | Maps Timeline domain, idempotency, query, and transition errors to stable HTTP envelopes | `MutationAPIErrorContext`, `ClassifyMutationAPIError` | Timeline admission routes and workbook Timeline handlers | Timeline, auth, `httpapi`, `viewquery` | Error classifier and route tests | Implements Timeline and common error contracts | Timeline transport/admission | high | Package path conflicts with operation owner `module.timeline` |
| `internal/modules/workbook/timelineadmission/assertions_test.go` | Provides package-local API-error assertions for Timeline admission tests | Test-only helper | Timeline admission tests | `httpapi`, testing | Decoding and classifier tests | No generated write | Timeline verification | low | Move with Timeline admission tests |
| `internal/modules/workbook/timelineadmission/decoding.go` | Strictly decodes Timeline query, create, patch, conflict, action, supersede, and time-conversion requests | Seven exported decoder functions | Workbook Timeline handlers, Timeline routes, Timeline and owner tests | Timeline, conflict tokens, field normalization, `httpapi`, `viewquery`, `viewschema` | Decoding, Timeline, links, and task/decision tests | Consumes generated Timeline view schema and public envelopes | Timeline transport/admission | critical | Owner-specific field policy should live behind Timeline |
| `internal/modules/workbook/timelineadmission/decoding_test.go` | Characterizes Timeline patch, action, profile, collection, and visible-text decoding | Two top-level test suites with many cases | Go test discovery | Timeline admission decoder and view schemas | Self | Supports workbook unit row for Timeline admission | Timeline verification | high | Reassign evidence when package moves |
| `internal/modules/workbook/timelineadmission/error_classifier_test.go` | Characterizes Timeline public error mapping | `TestTimelineMutationAPIErrorClassifier_Unit` | Go test discovery | Timeline admission classifier | Self | Supports Timeline admission error evidence | Timeline verification | medium | Preserve exact status/code/details |
| `internal/modules/workbook/timelineadmission/hashing.go` | Canonically hashes Timeline create, patch, conflict, and lifecycle actions | Four exported hash functions | Workbook Timeline handlers, Timeline routes, Timeline and task tests | Timeline and conflict tokens | Hashing, Timeline, links, and task tests | Implements idempotency request identity | Timeline transport/admission | high | Hash stability is observable through replay behavior |
| `internal/modules/workbook/timelineadmission/hashing_test.go` | Verifies Timeline create decoding and hash normalization | `TestTimelineAdmissionCreateAndHashContracts` | Go test discovery | Timeline admission decoder and hashing | Self | Supports a workbook unit evidence row | Timeline verification | medium | Move with admission implementation |
| `internal/modules/workbook/timelineadmission/routes.go` | Registers three `module.timeline` routes and performs auth, role checks, decoding, session sliding, and response writing | `RouteOptions`, `RegisterRoutes` | `internal/app/server/runtime.go` | Incidents, Timeline facade, auth, HTTP transport | Timeline route and server tests | Binds authored `module.timeline` OpenAPI operations | Timeline HTTP adapter | critical | Clearly misplaced by path while already correctly owned in route metadata |
| `internal/modules/workbook/timestamp_contract_test.go` | Verifies timestamp direct-scalar decoding for workbook patches | `TestTimestampInstantPatchDecoder_Unit` | Go test discovery | Workbook mutation API and view contracts | Self | Reads generated timestamp/write-kind metadata | Workbook admission verification | medium | Preserve when owner policy is delegated |
| `internal/modules/workbook/workbook_integration_test.go` | Characterizes all discovered base surfaces, common query behavior, pagination, cursor continuity, and coordination projections | Six integration tests | Go test discovery | Workbook query routes, projections, startup, database harness | Self | Exercises all 17 view-schema registrations and query envelopes | Workbook public query integration verification | critical | Primary cross-surface facade characterization |
| `internal/modules/workbook/workbook_mutation_integration_test.go` | Characterizes Parties, Evidence, coordination, Notes, Task/Decision, collections, and conflict side effects | Six integration tests and test-only side-effect state | Go test discovery | Workbook mutation routes and multiple source owners | Self | Exercises mutation OpenAPI, view schemas, revisions, and WS effects | Split owner semantics from workbook route integration verification | critical | Must be decomposed only after equivalent owner-local coverage exists |
| `internal/modules/workbook/workbook_startup_test.go` | Characterizes preference authorization, saved-view visibility, explicit/home/default fallback, invalid-pointer clearing, extension workspaces, and base-surface startup | Three tests plus test-only key-ring helper | Go test discovery | Workbook routes, startup, saved views, extensions, auth, database | Self | Exercises startup OpenAPI and frontend-consumed response shape | Workbook startup integration verification | critical | Primary behavior freeze for the startup split |

## 3. Module Boundary Diagnosis

The current target is a mixed-responsibility package. It contains a legitimate
thin application/service facade in concept, but the implementation also acts as
an accidental catch-all, view/projection orchestrator, transport-adjacent
adapter, persistence-adjacent adapter, and mutation coordinator. It is not
itself a frontend shell or grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Public workbook route registration and common envelopes | `routes.go` | Workbook application/HTTP facade | keep and split | Thirteen operations are authored under `module.workbook` | Keep the external facade; remove owner construction and detailed semantics |
| Public admission DTOs, strict JSON, and stable hashes | `mutation_api.go`, `batch_admission.go` | Workbook admission plus owner-contributed policies | split | Core 03 distinguishes public admission from source mutation | Common wire shape may remain while field legality comes from owners/contracts |
| Clipboard and bulk admission | `batch_admission.go`, `routes.go` | Workbook, Tabular Ingest, Timeline/entities | keep | Core 03 implementation note and current `tabularingest` calls | Do not move ordinary clipboard paste into file imports |
| Source-owner mutation dispatch and DTO adaptation | `mutation_store.go`, `entity_owner_adapter.go`, `store.go` | Source owners with a thin workbook coordinator | split | Calls artifacts, evidence, entities, parties, tasks/decisions, and Timeline facades | Workbook must not become source-state owner |
| Non-Timeline conflict-clear transaction | `mutation_store.go` | Authoritative source owner with Revisions support | move | Direct PostgreSQL transaction, idempotency, record target, and projection row load | RB-001 resolves the semantic command to the source owner; Revisions remains shared mechanism |
| Projection query dispatch | `query_store.go`, `paging.go` | Projections catalog plus workbook query facade | split | Projection manifest and import allowlist name workbook files | Preserve query/paging contract; inject providers |
| Startup selection and `sheet_ref` response semantics | `startup/api.go`, `startup/store.go`, `startup/workspace_registry.go` | Workbook startup application facade | keep and split | Core 02/03 define startup selection and shell behavior | Selection can remain workbook-facing |
| Workbook preference SQL and incident-create bootstrap | `startup/store.go`, `startup/bootstrap/bootstrap.go` | Incidents persistence | move | Schema ownership manifest assigns preference tables to `incidents` | Route ownership may remain workbook while storage ownership moves |
| Saved-view startup resolution | `startup/store.go` | Saved Views behind a workbook startup port | split | Calls `savedviews.ResolveStartupVisibleForUpdateTx` | Do not duplicate saved-view visibility rules |
| Timeline-specific transport and admission | `timelineadmission/*` | Timeline | move | Routes bind `module.timeline`; Timeline tests import the package | Preserve paths, envelopes, roles, hashes, and session behavior |
| Workbook telemetry | `telemetry.go` | Workbook facade instrumentation | keep | Adopted OTel NLSpec names `cartulary.workbook` query/mutation scopes | Move only if the facade itself moves |
| Owner-local workbook test helpers | `testsupport/*` | Workbook test support | keep and split | `tools/test_support_inventory.json` declares an owner-local root | Keep only genuinely workbook-specific helpers |
| Shared server, fixture, and WebSocket test helpers | `testsupport/scenariotest/*` | `internal/testutil/appsupport`, collaboration support, or source-owner support | move and split | Many non-workbook packages import them | Preserve semantic test ownership during moves |
| Frontend shell and Timeline controllers | `apps/web/src/workbook/**` | `/apps/web` and possible future feature seams | defer | Current frontend boundary rules and large controller surfaces | A separate target and authorization are required |
| Direct grid-vendor integration | `packages/grid-adapter` | Grid Adapter | keep | Workbook imports `@cartulary/grid-adapter`; only adapter imports `react-data-grid` | No target defect found |

### Approved closure posture

The `2026-07-26` closure instruction is the approval record for Workbook,
Revisions, application assembly, Incidents, Saved Views, and the active source
owners. It closes RB-001 through RB-003 as architecture decisions. Production
work remains unauthorized until S-00 characterizes the observable baseline and
S-00A promotes the confirmed allocation into the applicable owner and
machine-boundary inputs.

No public route, operation ID, status, error or success envelope, schema ID,
field-key grammar, cursor, WebSocket behavior, authorization outcome, session
behavior, or OpenTelemetry vocabulary may change in these structural slices.
Any such change is `requires later authorization`.

| Decision | Approved permanent allocation | Implementation gate |
| --- | --- | --- |
| RB-001 conflict resolution | Workbook owns route admission. Each authoritative source owner owns its conflict-resolution command and final source revalidation. Revisions supplies shared token, conflict-window, history, and replay mechanics. Collaboration publishes only after commit. | S-00 conflict/replay characterization and S-00A owner-authority promotion before S-05 |
| RB-002 contribution interface | Application assembly constructs one immutable Workbook Contribution Catalog with query/create indexes keyed by `view_schema_id` and patch/conflict indexes keyed by authoritative `record_type`. | S-00 17-surface and construction-failure characterization and S-00A boundary adoption before S-03/S-04 |
| RB-003 startup persistence | Incidents owns preference SQL, repository behavior, and incident-create/import bootstrap. Workbook retains preference/startup routes and ordered startup selection. Saved Views and workspace availability are injected. | S-00 startup/concurrency characterization and S-00A owner-authority promotion before S-02 |

### Approved internal interface model

The following names express the required seam. A later implementation may
adjust naming to established Go conventions, but it must not weaken the
ownership, inputs, lookup dimensions, immutability, or transaction semantics.

| Boundary | Required surface and semantics |
| --- | --- |
| `ConflictResolutionOwner` | `ResolveConflict(context.Context, ConflictResolutionCommand) (ConflictResolutionResult, error)`. The command contains the authenticated actor, visible target, verified token, normalized resolution and digest, and verified view/field context. |
| Revisions conflict support | Token preflight plus one unit-of-work coordinator for conflict-window, history, replay, and revision mechanics. Source owners provide the authoritative load/lock, legality, current-value, resolution, mutation, and canonical-row operations. Revisions must not become a generic source mutation owner. |
| `WorkbookContributionCatalog` | Immutable lookups `QueryFor(ViewSchemaID)`, `CreateFor(ViewSchemaID)`, `PatchFor(RecordType)`, and `ConflictFor(RecordType)`. Query and create remain surface-scoped; patch and conflict remain record-scoped. |
| Query provider | Receives authorized incident scope, the exact active view schema, normalized query, and validated keyset window; returns rows and continuation position. Workbook alone encodes the public opaque cursor and response page. |
| Create provider | Receives actor, incident, exact view schema, `client_txn_id`, and normalized field-key values; owns lifecycle, defaults, source validation, replay, history, projections, and canonical result. |
| Patch provider | Receives actor, visible record target, authoritative loaded record type, view context, base row version, `client_txn_id`, and normalized changes; owns field legality, concurrency, history, projections, and post-commit effects. |
| Startup unit of work | Runs DB-backed preference and Saved View resolution atomically without exposing `pgx.Tx` through the Workbook API. |
| Preference repository | Incidents-backed get/replace operations plus conditional clear-if-current operations. Public default replacement remains admin-only; automatic startup repair can only write `null` for the exact invalid value. |
| Startup Saved View resolver | Saved Views owns identity, incident scope, visibility, and deletion checks behind a transaction-aware port; Workbook owns candidate ordering. |
| Incident preference bootstrap | Participates in ordinary incident creation and incident-bundle final publication in the same transaction and with the same publication timestamp. |

Catalog construction occurs once under application assembly and is immutable
before route registration. The expected keys come from the active view-schema
registry and code-backed projection descriptors, never specification prose or a
second handwritten runtime manifest.

| Invalid catalog condition | Required result |
| --- | --- |
| Duplicate key in any dimension | Assembly failure |
| Unknown view schema or record type | Assembly failure |
| Missing query provider for an active queryable surface | Assembly failure |
| Query provider disagrees with the active projection descriptor | Assembly failure |
| Create provider exists for a non-create surface, or a create-capable surface lacks one | Assembly failure |
| Patch provider exists where no current field is patchable | Assembly failure |
| Conflict provider is absent for a record type with a conflict-capable writable field | Assembly failure |
| Contribution owner disagrees with the active source-owner/discriminator mapping | Assembly failure |
| Deprecated or experimental contribution reaches production without owner-defined compatibility | Assembly failure |
| Lookup misses after admission of a known active schema | Internal assembly defect; it must not create a public `provider_not_found` contract |

The current registry has the following 17 surfaces. This table is planning
evidence; application construction and tests must derive the live set from the
machine registry so future phases can add a contribution without editing a
central switch or this document.

| Surface | Source record type or discriminator | Contribution owner |
| --- | --- | --- |
| Timeline | `timeline_event` | Timeline |
| Hosts | `host` | Entities |
| Identities | `identity` | Entities |
| Evidence | `evidence` | Evidence |
| Notes | `artifact_type='note'` | Artifacts |
| Indicators | `indicator` | Indicators; current create-only policy does not imply a patch/conflict contribution |
| Compromise Assessments | `assessment` | Assessments create provider; append-only policy omits patch/conflict contributions |
| Task Requests | `task_request` | Tasks and Decisions |
| Decisions | `decision` | Tasks and Decisions |
| Parties | `party` | Parties |
| Communications Log | `artifact_type='comm_log'` | Artifacts |
| Handoff | `artifact_type='handoff'` | Artifacts |
| Status Review | `artifact_type='status_review'` | Artifacts |
| Lesson | `artifact_type='lesson'` | Artifacts |
| Findings | `artifact_type='finding'` | Artifacts when the optional surface is active |
| Investigative Queries | `artifact_type='investigative_query'` | Artifacts when the optional surface is active |
| Forensic Keywords | `artifact_type='forensic_keyword'` | Artifacts when the optional surface is active |

The catalog does not absorb clipboard paste, Timeline bulk operations,
Timeline mark-reviewed or time-conversion routes, linked-note creation,
supersede, evidence blob attachment and handle issuance, merge, rollback, or
extension workspaces. Those remain dedicated ports because a universal action
registry would recreate the catch-all being removed.

A temporary compatibility implementation is permitted only as an explicit
contribution for every exact key it serves. There is no missing-key fallback to
the old switch, global mutable or `init()` registration, reflection-based
discovery, route probing, or second JSON runtime registry. Every key resolves
exactly once throughout migration, and compatibility entries are removed as
soon as their owner implementations pass equivalence.

### Current behavior findings retained by the plan

| Behavior | Current evidence | Required posture |
| --- | --- | --- |
| `keep_saved` | `clearWorkbookConflict` writes route-idempotency response state and returns the current canonical row; durability tests show no source write, row-version advance, change set, source revision, projection mutation, or Collaboration event | Preserve; the source-owner command must reproduce the no-op domain outcome |
| Exact conflict replay | Current owner and Workbook tests return the original success and do not repeat mutation, revision, projection, or event effects | Preserve |
| Divergent or stale conflict replay | Current direct characterization is incomplete across token reuse, changed payload, changed `client_txn_id`, and later source mutation | S-00 must characterize the current status/envelope and side effects; structural slices preserve the result unless a separate behavior-change task is authorized |
| Startup pointer clear | The store locks preference rows with `FOR UPDATE`, clears inside the same transaction, and advances `updated_at`; the current SQL does not replace `incident_workbook_preferences.updated_by_user_id` | Preserve the current attribution in S-02; changing it requires owner authority and later behavior-change authorization |
| Concurrent pointer replacement | Current row locking prevents a replacement between the validated read and clear; the target port additionally requires clear-if-current semantics | S-00 adds an explicit no-clobber test; S-02 must keep one atomic DB-backed unit of work or reread and restart after a failed comparison |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/incidents/{incident_id}/workbook-preferences/default` | `module.workbook`; persistence objects owned by incidents | Workbook OpenAPI owner, Core 01/02/04 | `workbook_startup_test.go`, startup store tests | Preserve membership lookup, resource shape, null handling, and session sliding | high | Storage movement must not move the route |
| `PUT /api/v1/incidents/{incident_id}/workbook-preferences/default` | `module.workbook`; persistence objects owned by incidents | Workbook OpenAPI owner, Core 04 | Startup API/store and workbook startup tests | Preserve admin-only role, CSRF, canonical `sheet_ref`, structural no-op timestamps, and attribution | critical | No behavior change authorized |
| `GET /api/v1/incidents/{incident_id}/workbook-preferences/me` | `module.workbook`; persistence objects owned by incidents | Workbook OpenAPI owner, Core 01/04 | Startup and incident route tests | Preserve caller-scoped lookup and membership failure behavior | high | Viewer access remains valid |
| `PUT /api/v1/incidents/{incident_id}/workbook-preferences/me` | `module.workbook`; persistence objects owned by incidents | Workbook OpenAPI owner, Core 04 | Startup API/store and workbook startup tests | Preserve self-only update, CSRF, role-independent membership access, no-op timestamps | critical | Must not acquire admin semantics |
| `GET /api/v1/incidents/{incident_id}/workbook-startup` | `module.workbook` | Workbook OpenAPI owner, Core 01/02/03 | `workbook_startup_test.go`, startup API/store tests, browser row | Preserve explicit to home to default to Timeline order, saved-view visibility, extension availability, atomic no-clobber pointer clearing, current timestamps, and attribution | critical | Frontend consumes exact response members |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | `module.workbook`; providers owned by projections/source modules | Workbook OpenAPI, Core 01 query contract, 17 view schemas | `workbook_integration_test.go`, `row_wire_test.go`, projection tests | Add provider-catalog composition equivalence for every registered schema before switch removal | critical | Preserve full row envelope, filter/sort/group, cursor, and live auth |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | `module.workbook`; source mutation owned per record family | Workbook OpenAPI and view schemas | Mutation integration, coordination, entity, evidence, assessment, Timeline tests | Add a table-driven 17-surface dispatch ownership check where current coverage is indirect | critical | Assessment currently uses a custom route option |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste` | `module.workbook`; plan/apply delegated | Workbook OpenAPI, Core 03 | Clipboard unit/integration and browser tests | Preserve CSV/TSV parsing, exact headers, entity-origin path, conflicts, ordering, and cross-incident rejection | critical | File imports remain separate |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations` | `module.workbook`; Timeline apply owner | Workbook OpenAPI, Core 03 | Clipboard/bulk integration tests | Preserve bounds, fill/tag semantics, one visible batch, hashes, and target authorization | critical | Do not masquerade as clipboard paste |
| `PATCH /api/v1/records/{record_id}` | `module.workbook`; source owner per record | Workbook OpenAPI, Core 01/02/04 | Mutation, conflict, route-guard, and owner tests | Preserve guard precedence and owner-specific legality across every writable surface | critical | Route remains stable through internal extraction |
| `POST /api/v1/records/{record_id}/linked-notes` | `module.workbook`; artifacts/linked-notes source owner | Workbook OpenAPI, Core 02/03 | Mutation and linked-note coverage in workbook tests | Preserve source incident lookup, idempotency, typed links, and returned row | high | Adapter should remain narrow |
| `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` | `module.workbook` admission; source-owner command with Revisions support | Workbook OpenAPI, Core 01/02/03, approved RB-001 allocation | `conflict_route_test.go`, integration and route-guard tests | Add registry-derived owner matrix for all three resolution kinds, exact/divergent replay, stale tokens, guard precedence, rollback, canonical rows, and side effects | critical | Permanent internal ownership is resolved; S-00 evidence remains the implementation gate |
| `POST /api/v1/records/{record_id}/supersede` | `module.workbook`; Timeline or tasks/decisions source owner | Workbook OpenAPI, Core 02/03 | Mutation integration and Timeline/task tests | Preserve reviewer/admin role, replacement semantics, multi-record changes, and replay | critical | Public endpoint intentionally spans two record types |
| `GET /api/v1/incidents/{incident_id}/timeline-time-conversion-profile` | `module.timeline` | Timeline OpenAPI owner | Timeline admission and Timeline route tests | Preserve membership, response timestamps, nullable values, and session sliding | high | Implementation is currently under workbook path |
| `PUT /api/v1/incidents/{incident_id}/timeline-time-conversion-profile` | `module.timeline` | Timeline OpenAPI owner | Timeline decoder and route tests | Preserve admin-only update, version conflicts, strict decoding, and error mapping | critical | Move with Timeline admission |
| `POST /api/v1/records/{record_id}/mark-reviewed` | `module.timeline` | Timeline OpenAPI owner | Timeline admission and lifecycle tests | Preserve reviewer/admin authorization, lifecycle guards, idempotency, and response | critical | Move with Timeline admission |
| Seventeen registered workbook view schemas | Core 01 owners by surface; machine projection under `contracts/view-schemas` | `contracts/view-schemas/index.json` and generated registry | Workbook all-surface, grouping, optional-surface, projection, frontend contract tests | Characterize provider ownership and writable/read-only capabilities before central switches change | critical | Includes 14 required and 3 standardized optional surfaces |
| `view_row_v1`, query metadata, cursor continuation, full and sparse patch families | Core 01; projections and platform view-query | Core 01 and generated view contracts | `row_wire_test.go`, `workbook_integration_test.go` | Preserve cursor tamper rejection and authorization recheck after provider extraction | critical | No offset fallback |
| Saved-view and startup `sheet_ref` semantics | Saved Views and Workbook | Core 02/03, startup code, saved-view owner | Startup and saved-view tests plus browser rows | Preserve visibility, immutable base schema identity, extension workspace distinction, and fallback | critical | Saved-view persistence must not move into workbook |
| Projection refresh/rebuild and provider descriptors | Projections with source-owner providers | Projection manifest, application catalog, schema ownership | Projection and workbook integration tests | Preserve provider descriptors, refresh capabilities, rebuild order, and row version | critical | Update authored descriptors and allowlists, then generate |
| Revision, change-set, replay, and idempotency outcomes | Revisions shared mechanics, auth idempotency, and authoritative source owners | Core 01/02 and current facades | Conflict, mutation, shared harness, owner tests | Characterize exact/divergent/stale conflict replay, non-Timeline conflict outcomes, multi-record supersede boundaries, and rollback of every partial failure | critical | `keep_saved` has no domain mutation; mutating resolutions retain ordinary owner revision/projection effects |
| `GET /ws/v1/incidents/{incident_id}` and replayable `record_changed` | Collaboration; events contributed by source owners | WS schema and collaboration code | Workbook integration, row-wire, coordination, shared harness, collaboration tests | Preserve identity, ordering, changed keys, affected views, patch/invalidate fallback, replay, and auth | critical | Workbook does not own socket transport |
| Authorization and CSRF outcomes | Auth and Incidents; enforced by route adapters | Core 04 and route handlers | Route guard, shared harness, startup, cursor auth tests | Extend route-wide guard matrix when handlers split | critical | Read membership and write role sets must remain exact |
| OpenTelemetry query/mutation vocabulary | Adopted OpenTelemetry NLSpec | `telemetry.go`, platform registry, OTel conformance tool | `telemetry_test.go`, platform registry tests | Preserve scope, span names, metric names, attributes, and safe tokens | high | No raw IDs or content in telemetry |
| Frontend startup/query/mutation/controller behavior | `/apps/web` | Core 03, frontend workbook code, frontend boundary manifest | 24 Vitest rows and browser/stateful/a11y/visual rows under `module.workbook` | Run affected frontend rows only if a later backend slice changes an interface | high | Frontend source is not part of this target |
| Grid adapter and stable UI selector contracts | Grid Adapter and UI Contracts | Dev guide, package code, frontend import boundaries | Grid-adapter, UI-contract, workbook unit and browser tests | No new characterization needed for backend-only moves | medium | Direct vendor isolation is currently correct |
| Harness and evidence accounting | Testing Harness NLSpec and machine owner catalogs | Verification owner and 86-row test family manifest | Owner-slice and harness contract targets | Reassign rows only when semantic behavior ownership moves; never infer from filenames | high | Accounting does not define runtime ownership |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `timelineadmission` implements `module.timeline` routes below the workbook path | Timeline OpenAPI owner, route binding, external Timeline callers | Wrong package ownership can perpetuate reverse dependencies | `must_fix` | Timeline | Move admission, hashes, errors, routes, and tests together after characterization |
| Workbook startup imports SQLc for tables assigned to incidents | `startup/store.go`, bootstrap, SQLc boundary allowlist, schema ownership manifest | Storage rules and incident creation are split across owners | `must_fix` | Incidents persistence plus Workbook startup and Saved Views ports | Apply approved RB-003 after S-00/S-00A; retain route ownership and atomic selection |
| Generic keep-saved conflict handling opens a transaction inside workbook | `mutation_store.go` uses pool, auth idempotency, record targets, and projection row loads | Revision and persistence semantics are hidden in a UI-shaped module | `must_fix` | Authoritative source owner with Revisions support | Apply approved RB-001 after registry-derived conflict characterization |
| `Store` constructs many owner facades and platform stores | `store.go`, `query_store.go`, `routes.go` | Workbook behaves as an internal composition root | `should_fix` | Application assembly | Inject the immutable catalog; temporary compatibility must be registered per exact key with no missing-key fallback |
| Query routing uses hard-coded owner exceptions | Host, Identity, and Indicator switch followed by generic projection service | New schemas can drift between contracts and code | `should_fix` | Application-assembled query contributions backed by Projections/source owners | Apply approved RB-002 and characterize all 17 schemas plus fail-closed construction |
| Mutation decoding and dispatch duplicate owner-specific field and view rules | Hard-coded switches and writable-field lists in mutation API/store | Owner contract changes require workbook edits | `should_fix` | Source-owner create/patch/conflict contributions plus view contracts | Retain common wire admission while moving legality and commit semantics behind the approved catalog |
| Workbook route handlers construct auth, incident, startup, and store dependencies from PostgreSQL | `newService` composition in `routes.go` | Transport code controls application composition | `should_fix` | Application assembly with narrow injected ports | Preserve `RegisterRoutes` while changing internal construction |
| Startup directly calls saved-view transaction helper | `sqlSavedViewResolver` in `startup/store.go` | Visibility rules may be duplicated or tied to SQL transaction types | `should_fix` | Saved Views behind the Workbook startup unit of work | Inject the approved resolver port without exposing `pgx.Tx` through Workbook APIs |
| Workbook route inventory includes entity merge, evidence handles, and other owner routes | `testsupport/scenariotest/route_inventory.go` | Test ownership can be mistaken for runtime ownership | `should_fix` | Respective owners plus a workbook composition matrix | Split semantic inventories and preserve cross-owner conformance coverage |
| Shared scenario runtime is imported by many modules | External imports of `workbook/testsupport/scenariotest` | Reusable app composition is hidden in an owner-local tree | `should_fix` | `internal/testutil/appsupport` | Move owner-neutral start/seed/HTTP helpers; retain workbook view helpers |
| Workbook production code does not directly import Collaboration for publication | Source owners create event intents; workbook tests observe them | Moving publication into workbook would violate ownership | `intentional/no_action` | Collaboration and source owners | Freeze the WS consequence contract; do not introduce workbook publication |
| Clipboard paste uses Tabular Ingest and not file Imports | Core 03 implementation note and current imports | Conflating paths would change provenance and workflow | `intentional/no_action` | Workbook plus Tabular Ingest and source owners | Preserve the current distinction |
| Frontend workbook controllers are large and include Timeline-specific state | `apps/web/src/workbook` inventory and file sizes | A backend refactor could expand into an unrelated frontend rewrite | `defer` | `/apps/web` future target | Freeze interfaces only; require separate authorization for frontend moves |
| Workbook uses only the grid-adapter facade | Imports use `@cartulary/grid-adapter`; vendor dependency is in adapter package | Direct vendor leakage would destabilize UI contracts | `intentional/no_action` | Grid Adapter | Keep the boundary and its static check |
| Generated contracts are consumed behind approved facades | `viewschema` loads generated Go contracts; generated roots are policy-protected | Hand edits would drift from owners | `intentional/no_action` | Contract owners and generators | Change authored inputs only and run Make-owned generation/drift checks |
| Test-only database seeding and SQL assertions remain outside production | Scenario helpers and `_test.go` files contain direct fixtures | Moving fixture shortcuts into production would leak test assumptions | `intentional/no_action` | Test support | Preserve test-only placement while splitting shared support |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Pin authority, baseline, scope, and one-file restriction | This tracker and owner documents | `make lint-markdown` for this planning artifact | Baseline and source posture recorded |
| WF-01 | Complete target inventory | chain | WF-00 | WF-02, WF-04 | Keep one current-state row for each of the 56 files | Entire `internal/modules/workbook` tree | File count and inventory reconciliation | No unseen target file remains |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map HTTP, WS, query, startup, revision, telemetry, frontend, and harness contracts | OpenAPI owners, view schemas, WS schema, app and target facades | Contract-source review; no product test required for planning | Every contract has an owner and evidence posture |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-05, WF-06 | Execute S-00 evidence planning for conflict replay/staleness, four catalog dimensions, startup concurrency/attribution, route precedence, and exact responses | Workbook, Timeline, owner, frontend, and harness tests | `make test-slice OWNER=module.workbook`; service-backed slice where required | Every approved ownership move has observable pre-move evidence |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05, WF-06 | Classify persistence, transport, owner, generated, frontend, and test-support coupling | Root package, startup, Timeline admission, boundary manifests | `make backend-module-boundary-check` during implementation | Findings classified without treating tests as architecture |
| WF-05 | Approved facade design and owner-authority promotion | chain | WF-02, WF-03, WF-04 | WF-06 | Apply RB-001 through RB-003 to a thin Workbook facade and complete S-00A owner/machine-boundary promotion | Core owners, authored boundary/schema-owner/provider inputs, routes, stores, startup, app assembly, owner facades | Owner review; `make lint-markdown`; `make json-shape-check`; `make generate-drift`; `make backend-module-boundary-check` as affected | Confirmed behavior and approved allocation are authoritative before production movement |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Keep Timeline relocation independent, then execute S-02 to S-05 as one ordered ownership chain | Target packages, app assembly, owner packages | Narrow owner slice after each later implementation slice | Each slice has one dispatch authority, rollback, and completion criteria |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Move tests and authored owner rows with semantics, then regenerate | Test support, owner tests, verification and family inputs | Owner slices, JSON shape, generated drift | No row is inferred from path or title |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Run proportional validation and leave a resumable handoff | Changed implementation and authored inputs from later task | Narrow checks, then `make check` when risk warrants | Results, run roots, failures, skips, and next action recorded |

## 7. Proposed Refactor Slice Plan

All slices below are behavior-preserving and require a later authorized
implementation task. Any alternative that changes an observable contract must
be marked `requires later authorization` and planned separately.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Characterize the public facade before production movement: conflict replay/staleness and side effects; query/create/patch/conflict dispatch; startup bootstrap, fallback, concurrent repair, timestamps and attribution; route precedence; exact response equivalence | Workbook and source-owner tests; authored test-family inputs only when new executable evidence needs accounting | Tests could encode the proposed implementation instead of current observable behavior | Registry-derived conflict matrix for every active conflict-capable writable record type and all three outcomes; all 17 query surfaces and every supported mutation dimension; startup create/import/replay/rollback and no-clobber matrix; canonical rows and error envelopes | `make test-slice OWNER=module.workbook`; `make service-backed-test-slice OWNER=module.workbook` | Revert only new characterization and its authored accounting together | The exact baseline, including unknown divergent/stale replay outcomes, is executable evidence for every later slice |
| S-00A | S-00 and approved RB-001 through RB-003 | Promote confirmed public behavior and the approved ownership allocation into applicable Core owner sections and authored machine boundary/schema-owner/provider inputs; do not change production behavior | Core 00-04 as applicable, authored boundary and ownership inputs, provider contracts, verification owners | Tracker guidance could be mistaken for runtime authority or an uncertain behavior could be adopted prematurely | Owner-contract and machine-input validation; parity between active schemas, descriptors, expected catalog dimensions, and owner IDs | `make lint-markdown`; `make json-shape-check`; `make generate-drift`; `make backend-module-boundary-check` | Revert owner inputs and their generated refresh together; never hand-edit generated roots | Source-owner conflict semantics, Revisions support, the four catalog dimensions, and Incidents preference ownership are authoritative before code moves |
| S-01 | S-00, S-00A | Relocate Timeline admission, hashing, error mapping, and three Timeline routes behind the Timeline package boundary | `workbook/timelineadmission`, Timeline package, server composition, importing tests | HTTP paths, roles, hashes, errors, session sliding | Timeline decoder/hash/error/route tests and workbook route tests | `make test-slice OWNER=module.workbook`; `make backend-module-boundary-check` | Restore the old package path and imports as one slice | No production Timeline route code remains below Workbook; public behavior is unchanged |
| S-02 | S-00A | Move preference SQL/repository behavior and create/import bootstrap to Incidents; keep Workbook DTOs, routes, validation, and ordered selection; inject Saved Views and immutable workspace availability through one startup unit of work | Workbook startup, Incidents, Saved Views, incident creation/import publication, application assembly | Atomic repair, no-clobber concurrency, timestamps, preserved `updated_by_user_id`, fallback, visibility, no-store response | Startup API/store/integration; incident-create/import bootstrap atomicity and replay; saved-view deletion/visibility; extension loss; role and no-op tests | `make service-backed-test-slice OWNER=module.workbook`; relevant Incidents/Saved Views owner slices; `make backend-module-boundary-check` | Retain the public startup facade and revert the unit-of-work/repositories/bootstrap wiring together | Workbook production code imports no preference SQLc; Incidents owns persistence/bootstrap; Saved Views owns visibility; route and response behavior is identical |
| S-03 | S-02 | Construct the immutable Workbook Contribution Catalog in application assembly and migrate all query keys from DB-created stores and switches to exact query contributions | `query_store.go`, `paging.go`, `store.go`, projection/server assembly, owner query providers and descriptors | All 17 schemas, cursor identity, live auth, full/sparse rows, rebuild/refresh, invalid assembly | All-surface query, catalog duplicate/missing/unknown/owner-mismatch failures, row-wire, projection, grouping, pagination, cursor and authorization tests | `make service-backed-test-slice OWNER=module.workbook`; `make generate-drift`; `make backend-module-boundary-check` | Register the old implementation only as explicit per-key compatibility contributions; never fall through after a catalog miss | Every active query surface resolves exactly once before serving and public output is identical |
| S-04 | S-03 | Add exact create-by-view and patch/conflict-by-record-type contributions; split common Workbook admission from source-owner legality, replay, transaction, and result adaptation | Mutation API/store, service ports, source-owner facades, application assembly | Validation codes, hashes, idempotency, lifecycle, collections, changed keys, authoritative record-type dispatch | Catalog construction failures; all create-capable surfaces; owner-local patch/conflict eligibility; decoder, mutation, route-guard and replay tests | `make test-slice OWNER=module.workbook`; `make service-backed-test-slice OWNER=module.workbook`; `make backend-module-boundary-check` | Use explicit per-key compatibility contributions while migrating; no global registration, reflection, or legacy fallback | One catalog is the sole dispatch authority and no source-state rule exists only in generic Workbook switches |
| S-05 | S-04 | Move all non-Timeline conflict commands to authoritative source owners; inject Revisions token/window/history/replay support while Workbook retains admission and envelope mapping | `mutation_store.go`, source-owner conflict facades, Revisions support, projections, application assembly | `keep_saved` no-op semantics, exact/divergent/stale replay, current row, source revisions, projections, post-commit events, security precedence | Registry-derived conflict matrix, durability and rollback, route guard, canonical row, owner and WebSocket consequence tests | `make service-backed-test-slice OWNER=module.workbook`; relevant source-owner and Revisions slices | Keep the old implementation only as explicit record-type contributions until equivalent results pass; remove it without a fallback path | Workbook performs admission/delegation only; source owners commit semantics; Revisions supplies shared mechanics; each outcome matches S-00 |
| S-06 | S-01, S-05 | Split test support and move owner-specific tests and semantic catalog rows with their behavior owners | Workbook testsupport, `internal/testutil/appsupport`, source-owner tests, authored verification inputs | Lost coverage, duplicate selectors, changed service fixtures, false ownership | Shared harness inventory, route conformance, owner slices | `make test-slice OWNER=module.workbook`; `make service-backed-test-slice OWNER=module.workbook`; `make json-shape-check` | Move each helper and its callers in one reviewable commit | Workbook support contains only Workbook-specific helpers and every moved row resolves once |
| S-07 | S-03, S-04 | Defer frontend seam work unless a backend interface must change; if authorized, split shell-generic and Timeline feature controllers without vendor leakage | `apps/web/src/workbook`, frontend services, view/UI/grid contracts | Pending replay, startup, projection refresh, focus, selectors, browser behavior | Affected Vitest, browser, stateful, accessibility, and visual rows | `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check` | Keep backend wire compatibility so frontend changes can be reverted independently | `DEFERRED` until a separate authorization names the frontend target |
| S-08 | S-06 and any authorized S-07 | Update remaining authored boundaries/provider/verification inputs, regenerate through Make, remove all compatibility contributions and switches, and perform final validation | Authored contracts/manifests, generators, obsolete wrappers and switches | Generated drift, OpenAPI compatibility, telemetry, harness accounting, accidental dual authority | All retained owner and contract evidence plus absence checks for fallbacks/global registries | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make openapi-compatibility-check`; `make otel-conformance`; `make check` | Revert authored input and generated output as one slice; never hand-edit generated files | No compatibility path or second authority remains, all owners account for behavior, and selected full validation passes |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.workbook` | Current non-service workbook Go, Vitest, and other owner-selected rows | yes | Establish the focused baseline and use `ROWS=` for later exact slice selection |
| integration | `make service-backed-test-slice OWNER=module.workbook` | Current service-backed Go and browser rows selected for workbook | yes | Required before persistence, route, mutation, or projection moves |
| e2e/browser | `make browser-e2e-webserver-backed` | Workbook browser contract and dependent owner scenarios | no | Run after public route/interface changes; add stateful, a11y, or visual targets only when affected |
| generated drift | `make generate-drift` | Generated Go/TypeScript/contracts/topology drift | no | Required after authored contract, provider, or harness inputs change |
| generated policy | `make generated-artifact-policy-check` | Prohibits invalid generated-root edits | no | Required whenever generation or generated paths are implicated |
| contract shape | `make json-shape-check` | Authored contract and manifest JSON | no | Required after verification, provider, view-schema, or boundary input changes |
| OpenAPI compatibility | `make openapi-compatibility-check` | Workbook and Timeline API compatibility | no | Required if authored OpenAPI inputs change |
| import-boundary/static | `make backend-module-boundary-check` | Backend ownership and import rules | yes | Update authored rules only after the permanent seam is approved |
| frontend static | `make frontend-import-boundary-check` | Frontend owner and feature-facade rules | no | Only for a separately authorized frontend slice |
| telemetry | `make otel-conformance` | Workbook/Timeline instrumentation contract | no | Required if route instrumentation location or vocabulary changes |
| full check | `make check` | Repository developer verification gate | no | Run after narrow slices pass and risk warrants the full gate |
| tracker documentation | `make lint-markdown` | This planning-only closure update | yes | The only validation run for this closure pass |

`make agent-finalize` is intentionally skipped for this planning-only task:
`RESULTS_DIR` is unset, no broader retained run is being reused, and the target
may update tracked generated or baseline artifacts outside the one-file write
allowance. Product suites are also skipped because no product code, tests, or
contracts changed.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| WB-001 | Normalize `internal/modules/workbook` to target label `workbook` and lock planning-only scope | Scope | DONE | none | Section 1 | Target, output, authority, and exclusions are explicit |
| WB-002 | Inventory all 56 target files | Discovery | DONE | WB-001 | Section 2 | 27 non-test/support and 29 test files each have one row |
| WB-003 | Map public HTTP, WS, view, startup, revision, telemetry, frontend, and harness contracts | Contracts | DONE | WB-002 | Section 4 | Each discovered contract has an owner and test posture |
| WB-004 | Diagnose module boundaries and coupling | Architecture | DONE | WB-002, WB-003 | Sections 3 and 5 | Findings are classified without treating the directory as authority |
| WB-005 | Close characterization gaps for provider and conflict seams | Tests | TODO | WB-003, WB-004 | S-00 | Every risky move has direct behavior evidence |
| WB-005A | Promote approved ownership and confirmed behavior into owner/machine authority | Contracts/boundaries | TODO | WB-005 | S-00A and resolved RB-001 through RB-003 | Owner documents and machine inputs agree before production movement |
| WB-006 | Relocate Timeline admission behind Timeline | Backend | TODO | WB-005A | S-01 | No Timeline-owned production route code remains under Workbook |
| WB-007 | Split startup selection from Incidents preference persistence | Backend | TODO | WB-005A | S-02 and resolved RB-003 | Storage and application ownership are explicit with identical behavior |
| WB-008 | Introduce the immutable Workbook Contribution Catalog and query contributions | Backend | TODO | WB-007 | S-03 and resolved RB-002 | All 17 schemas resolve once without DB construction or fallback in Workbook |
| WB-009 | Split mutation admission and source-owner dispatch | Backend | TODO | WB-008 | S-04 and resolved RB-002 | Owner semantics no longer live only in Workbook switches |
| WB-010 | Extract non-Timeline conflict persistence | Backend | TODO | WB-009 | S-05 and resolved RB-001 | Workbook delegates conflict commands to source owners with Revisions support |
| WB-011 | Split shared and owner-local test support | Tests/harness | TODO | WB-006, WB-007, WB-008, WB-009, WB-010 | S-06 | Helpers and semantic rows live with their actual owners |
| WB-012 | Plan frontend controller split as a separate target | Frontend | DEFERRED | WB-003 | S-07 | Separate authorization names the frontend scope |
| WB-013 | Update authored boundary/contract/accounting inputs and regenerate | Contracts/harness | TODO | WB-011 and any authorized WB-012 | S-08 | No hand edits; generated drift and accounting pass |
| WB-014 | Run proportional final validation | Validation | TODO | WB-013 | Section 8 | Narrow checks pass before any selected broad gate |
| WB-015 | Create current tracker and planning handoff | Handoff | DONE | WB-001 through WB-004 | This file | Another agent can begin S-00 without rediscovery |
| WB-016 | Record approved RB-001 through RB-003 closure decisions | Handoff/architecture | DONE | WB-004 | Sections 3 and 11; `2026-07-26` closure instruction | Architecture questions are resolved and remaining evidence gates are explicit |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Planning scope, authority, baseline, and output are fixed | Inspected framework, domain, Core 00-05, OTel and Testing Harness NLSpecs, implementation guides; touched only this tracker | `git status`, `git rev-parse`, `sed`, `rg` | Clean `main` baseline recorded; no owner contradiction | None for planning | Begin S-00 only in a later authorized implementation task |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | RB-001 through RB-003 are approved architecture decisions; implementation remains unauthorized | Inspected `temp/analysis-notes.md`, this tracker, live app assembly, Workbook service/query/startup/conflict code, current SQL, view-schema registry, and prior closure patterns; touched only this tracker | `git status`, `git rev-parse`, `sed`, `rg`, `jq` | Approval and decision posture recorded without changing owner documents or executable authority | S-00 evidence and S-00A authority promotion gate later implementation, not this planning update | Execute S-00 in a separately authorized task |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | All 56 target files are inventoried; package diagnosed as mixed responsibility | Inspected all target paths plus server composition, projection assembly, provider/codegen/import boundaries; touched only this tracker | `rg --files`, `wc -l`, `rg`, `sed`, `jq` | Thin facade seam and misplaced persistence/Timeline/test-support responsibilities identified | RB-001, RB-002, RB-003 affect implementation design, not tracker completion | Add direct characterization, then execute S-01 through S-06 in order |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | Permanent backend seams are decided | Inspected Workbook store/ports/query/startup/conflict paths, application route/projection assembly, Incidents, Saved Views, Revisions, and active view schemas; touched only this tracker | `sed`, `rg`, `jq` | Source-owner conflict commands, four-index immutable catalog, and Incidents preference ownership replace the former open candidates | No architecture blocker; S-00/S-00A gate production movement | After the gates, keep S-01 independent and execute S-02 to S-05 in order |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Frontend is an affected consumer and a deferred refactor target | Inspected workbook file inventory, `WorkbookShell.tsx`, startup model, API facade, frontend import boundaries, and grid-adapter manifest; touched only this tracker | `rg --files`, `rg`, `sed`, `wc -l` | Shell/controller complexity recorded; no direct grid-vendor leak found | Separate authorization required for S-07 | Preserve backend wire contracts; do not modify frontend in backend slices |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | Frontend remains a contract consumer, not part of the approved backend work | Reused the inspected frontend boundary and grid-adapter evidence; touched only this tracker | Targeted `rg` and tracker review | Backend catalog and startup ports preserve public client contracts; no frontend move was added | S-07 remains separately authorized | Run frontend validation only if a later authorized interface change affects consumers |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Thirteen workbook and three Timeline HTTP operations, 17 view schemas, WS contract, projection descriptors, and generated boundaries are frozen | Inspected authored OpenAPI owners, view-schema index, WS schema, projection-provider index, generated policy evidence; touched only this tracker | `jq`, `rg`, `sed` | Contract map completed; no generated file was edited | None for planning | Update authored inputs only if a later slice changes ownership metadata |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | Public contracts remain frozen; internal ownership and contribution dimensions are approved | Inspected the active 17-surface registry, projection descriptors, authored owners, and current query/mutation dispatch; touched only this tracker | `jq`, `rg`, `sed` | Fail-closed catalog construction, one-key-one-provider migration, and no second runtime manifest are recorded | S-00 must prove parity; S-00A must promote confirmed ownership before code movement | Derive catalog expectations from machine inputs and regenerate only through Make in later slices |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Current owner has 86 rows across 13 families and Go, Vitest, and Playwright runners | Inspected verification owner, test-family manifest, test catalog, test-support inventory, route inventory, and target tests; touched only this tracker | `jq`, `rg`, `sed`, `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=module.workbook`, `make explain-test-owner OWNER=module.workbook`, `make explain-target` | Canonical narrow and broad validation routes recorded | No product baseline was run in this documentation task | Use exact semantic rows during S-00 and reassign authored rows only with behavior owners |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | S-00 now has exact conflict, catalog, startup, precedence, rollback, and response-equivalence evidence requirements | Inspected current conflict durability/guard tests, startup store/tests, registry metadata, and verification guidance; touched only this tracker | `rg`, `sed`, `jq` | Existing evidence is retained and missing divergent/stale replay and concurrency cases are explicit | No product test was run in this documentation pass | Add evidence before movement and derive owner/test matrices from active metadata rather than a manual list |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Read membership, write roles, admin preference control, reviewer supersede, CSRF, session sliding, and guard precedence are frozen | Inspected workbook and Timeline handlers, Core 04, route-guard and shared-harness tests; touched only this tracker | `rg`, `sed` | Authorization belongs at route/auth/incident seams; no domain authorization move is proposed | None for planning | Extend route-wide guard characterization before handlers are split |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | Workbook retains authentication, CSRF, hidden-record lookup, role gates, token preflight, and common envelope order; source owners revalidate semantic authority | Inspected conflict route ordering, startup roles, current pointer-clear SQL, and closure guidance; touched only this tracker | `rg`, `sed` | Catalog registration is explicitly not an authorization grant; every request remains deny-by-default and owner-revalidated | S-00 security-precedence evidence precedes S-04/S-05 | Preserve route ordering and test Timeline and non-Timeline owners uniformly |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-26T20:23:47-04:00 | Codex planning session | Tracker is ready for handoff; implementation remains unauthorized | Inspected current Git state and all evidence summarized above; touched only this tracker | `git status`; `make lint-markdown` | Passed; retained run root `.cartulary/test-results/20260727T002901Z-p53762` | RB-001, RB-002, RB-003 are implementation questions, not current planning blockers | A later session starts with S-00 and records exact selected row IDs and retained run roots |
| 2026-07-26T21:05:13-04:00 | Codex closure-guidance session | RB-001 through RB-003 are resolved decisions; S-00 and S-00A are the remaining pre-production gates | Inspected closure guidance and current implementation evidence; touched only this tracker | `git status`; targeted `sed`, `rg`, and `jq`; `make lint-markdown` | Passed; retained run root `.cartulary/test-results/20260727T010859Z-p68755` | No planning blocker or owner contradiction; production remains unauthorized | Start S-00, then S-00A; execute S-02 to S-05 sequentially and S-01 independently |

## 11. Open Questions and Blockers

There is no current planning blocker and no owner contradiction. The
`2026-07-26` closure instruction is the approval record for all three decisions.
The remaining gates are characterization and authority-promotion work, not open
architecture questions and not authorization to change production code.

| ID | Closure decision | Why it matters | Remaining evidence or authority gate | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Workbook owns conflict-route admission; each authoritative source owner exposes the conflict-resolution command; Revisions supplies shared token, conflict-window, history, revision, and replay mechanics; `keep_saved` has no source revision, projection mutation, or Collaboration event | Prevents Workbook or Revisions from becoming a generic source-domain owner while retaining one shared implementation of conflict mechanics | S-00 registry-derived outcome/replay/staleness/rollback/security matrix, then S-00A owner and boundary promotion before S-05 | `RESOLVED: owner decision approved; implementation pending S-00` |
| RB-002 | Application assembly builds one immutable Workbook Contribution Catalog with query/create keyed by `view_schema_id` and patch/conflict keyed by authoritative `record_type`; construction is fail-closed and no global registration, legacy fallthrough, reflection, or second runtime manifest is allowed | Makes future surface growth additive and owner-driven while ensuring one runtime dispatch authority | S-00 17-surface and invalid-construction characterization, then S-00A provider/boundary authority before S-03/S-04 | `RESOLVED: owner decision approved; implementation pending S-00` |
| RB-003 | Incidents owns preference SQL, repository behavior, and incident-create/import bootstrap; Workbook retains preference/startup routes and ordered selection; Saved Views and workspace availability are injected through one atomic startup unit of work | Aligns persistence with schema ownership without changing the public Workbook resource or startup experience | S-00 bootstrap/replay/rollback/fallback/concurrency/timestamp/attribution evidence, then S-00A owner authority before S-02 | `RESOLVED: owner decision approved; implementation pending S-00` |

## 12. Binary Completion Criteria

| Criterion | Result | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/workbook` is inventoried or explicitly out of scope | PASS | Section 2 contains one row for each of 56 files; none is excluded |
| Every discovered public contract risk has an owner and test posture | PASS | Section 4 covers HTTP, WS, 17 view schemas, startup, saved views, projections, revisions, auth, telemetry, frontend, grid, and harness |
| Every proposed workflow has dependencies and exit criteria | PASS | Section 6 defines WF-00 through WF-08 and the S-00/S-00A gate posture |
| Every proposed implementation slice is behavior-preserving unless explicitly marked otherwise | PASS | Section 7 excludes behavior changes, requires separate authorization, and sequences S-02 through S-05 |
| Validation commands are discovered or marked `TODO` with a reason | PASS | Section 8 contains Make-owned commands discovered from the live task surface |
| Contradictions are marked `BLOCKED: owner contradiction` | PASS | No contradiction was found; Section 1 defines the required fail-closed handling |
| Repository/framework mismatches are recorded as planning findings | PASS | Section 1 records that the live 56-file mixed package is broader than the template's thin-module doctrine |
| Handoff sections are current enough for another agent to continue without rediscovery | PASS | Sections 9-11 preserve history and append closure state to all seven workstream handoffs |
| RB-001 through RB-003 have one approved closure posture | PASS | Sections 3 and 11 record source-owner conflict commands, the immutable four-index catalog, and Incidents preference persistence |
| Future surface growth does not require a central switch or parallel authority | PASS | Section 3 requires machine-derived expected keys, fail-closed construction, exact per-key migration, and removal of all compatibility fallbacks |
| Production movement remains gated on evidence and executable authority | PASS | S-00 captures current behavior and S-00A promotes it before S-01 through S-05 |
| Only the permitted tracker file was touched | PASS | Session log and final Git verification record this tracker as the sole tracked change |
| No production refactor was performed | PASS | This artifact contains planning and documentation only |
