# graphprojection Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/graphprojection`
- **Target label:** `graphprojection` (derived from the final path segment and already valid lowercase kebab case)
- **Output path:** `docs/handoffs/graphprojection-module-refactor-tracker.md`
- **Status:** Complete; GP-S00 through GP-S08 were executed in the required order with terminal evidence.
- **Execution baseline:** Clean commit `e9ad86714ead7eb42984ce1983283228f4bc8ecf`.
- **Authorized scope:** GP-S00 through GP-S08, including owner adoption, implementation, tests, contracts, generated projections, documentation, and tracker evidence.
- **Non-goals:** No public Graph route, ordinary retained-store activation, cursor secret, hidden feature flag, workbook-provider semantic change, or database migration.
- **Required execution order:** `GP-S00 → GP-S04 → GP-S01 → GP-S02 → GP-S03 → GP-S05 → GP-S06 → GP-S07 → GP-S08`.

The source hierarchy used for this tracker is:

1. Adopted subsystem NLSpecs within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; it is not applicable to this planning-only tracker.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, and the planning framework as evidence and doctrine only.

This tracker uses three statement classes. They MUST NOT be conflated:

| Statement class | Meaning | Normative effect |
| --- | --- | --- |
| **Current fact** | A claim verified against an adopted owner or exact live repository evidence | Describes the current baseline and MUST be preserved unless a later authorized change proves otherwise |
| **Adopted owner requirement** | Behavior adopted in Graph Projection NLSpec 1.2.0 and the amended Core owners | Current conformance authority for its named scope |
| **Future implementation requirement** | A binary condition for a later authorized slice | Governs acceptance of that slice only after its owner prerequisites are adopted |

`temp/analysis-notes.md` was treated as recommendation evidence, not as an instruction or owner document. `docs/research/nlspec-spec.md` supplied NLSpec writing qualities: behavioral completeness, explicit interfaces and defaults, mapping tables, binary acceptance criteria, conceptual fidelity, and non-redundant normative language. Neither document adopted Graph Projection behavior. GP-S04 moved the durable requirements into Graph Projection NLSpec 1.2.0 and the applicable Core owners; those owner documents, not this tracker, are current authority.

The complete `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read first. It supplies the planning structure and doctrine but is not evidence of current repository state.

Owner documents inspected were `docs/graph_projection_nlspec.md`, `docs/spec/00_document_set_status_and_precedence.md`, relevant projection, persistence, recovery, and consumer sections of `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/02_domain_model_schema_and_history.md`, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, and `docs/spec/04_security_deployment_and_conformance.md`, `docs/network-flow-activity-nlspec.md`, `docs/reporting-subsystem-nlspec.md`, `docs/report-composition-nlspec.md`, and `docs/domain.md`. `AGENTS.md` supplied repository procedure. No owner contradiction was found.

Repository evidence at the planning baseline included all 38 files under `internal/modules/graphprojection`; production importers in Network Flow, Reporting, and recovery assembly; Network Flow route, schema, generated TypeScript, frontend controller, workspace, and browser-row evidence; Reporting binding validation and tests; migration 00022; recovery catalog inputs and projections; Graph Projection fixtures; verification owner and test-family manifests; and the public Make task surface. GP-S01 added nine same-package store concern files, GP-S02 added five same-package facade concern files, GP-S06 added the Graph-owned restore contracts and their tests, and GP-S07 added the Graph restore service and narrow PostgreSQL writer. The target now contains 58 files and 12,233 Go lines. Section 2 inventories the current files individually.

The framework's generic list of possible phase-shaped or UI-shaped extraction targets does not match the live package. The adopted Graph Projection owner and current code establish a real graph-oriented derivation subsystem. The package is internally mixed across an application facade, deterministic engine, persistence adapters, recovery metadata, and owner-specific test infrastructure, but it is not an accidental cross-domain catch-all. It remains distinct from workbook Projections, saved views, frontend grid adaptation, and source-owner modules.

## 2. Current-State Repository Inventory

Every file in the target is in scope for inventory. No file is excluded.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/graphprojection/aggregation.go` | Derives aggregate vertices and edges, grouping keys, endpoint behavior, merged properties, and mapped metadata | Package-private aggregation engine | `project.go` | Standard library and owner-local types/helpers | Engine, remediation, unit fixture, and store fixture tests | GP-FIX output and digest fixtures | Graph Projection engine | High | Graph semantics; keep under Graph Projection |
| `internal/modules/graphprojection/binding.go` | Defines the completed retained-run binding consumed by Reporting | `ProjectionBinding`, `ProjectionBindingReader` | Reporting and `postgresbinding` | `context`, owner-local lifecycle states | Reporting binding tests and `postgresbinding/reader_test.go` | Reporting `source_projection_ref.v1` contract | Graph Projection consumer facade | High | Same-transaction lookup seam is observable |
| `internal/modules/graphprojection/boundary_guard_test.go` | Enforces imports, dormant retained composition, approved consumers, no public Graph routes, graph-table-only SQL, and safe errors | Owner-routed boundary and composition tests | Graph Projection verification row | Go parser, production/configuration source and contract files | Self | OpenAPI/WS absence, dormancy, SQL, and module-boundary evidence | Graph Projection tests | High | Scans every decomposed store file and ordinary production composition |
| `internal/modules/graphprojection/canonical.go` | Implements canonical JSON, tuple encodings, digests, and deterministic identifiers | Package-private canonicalization helpers | Admission, engine, resources, retained adapter helpers | Standard library cryptography/encoding | Engine, fixture, remediation, and store tests | GP-FIX canonical and digest transcripts | Graph Projection engine | Critical | Exact bytes and identifiers are frozen behavior |
| `internal/modules/graphprojection/conformance_fixture_candidate_test.go` | Produces explicit disposable candidate fixture output under test results | `TestGraphProjectionFixtureCandidate` | Dedicated `graph-projection-fixture-candidate` Make target only | `fixturetest`, test environment, owner engine | Self | Candidate artifacts only; never committed goldens | Graph Projection test tooling | Low | Intentionally excluded from normal backend rows |
| `internal/modules/graphprojection/conformance_fixtures_test.go` | Verifies canonical strings, scalar boundaries, identifiers, and digest transcripts | Seven `TestGPFIX*` tests | Unit fixture wrapper/full Go suite | `fixturetest` and retained adapter helpers | Self | GP-FIX-014..017, 022, 023, 036 | Graph Projection tests | Medium | Individual tests are represented through the unit wrapper where applicable |
| `internal/modules/graphprojection/conformance_remaining_fixtures_test.go` | Aggregates unit Graph Projection fixture execution | `TestGraphProjectionUnitBehaviorFixtures_Unit` and 18 `TestGPFIX*` functions | Engine test-family row | Unit fixture executor | Self | GP-FIX unit corpus | Graph Projection tests | Medium | Wrapper is the authored harness selector |
| `internal/modules/graphprojection/conformance_remaining_store_fixtures_test.go` | Executes retained lifecycle fixture scenarios against PostgreSQL | Store fixture wrapper and 11 `TestGPFIX*` functions | Storage lifecycle test-family row | `postgresstore`, `pgtest`, `fixturetest` | Self | Retained GP-FIX corpus | Graph Projection storage tests | High | Service-backed retained-state evidence |
| `internal/modules/graphprojection/conformance_remediation_test.go` | Characterizes admission, scalar, resource, idempotency, and safe-detail remediation | Eleven direct `Test*` functions | Engine canonical-behavior row | Root engine internals | Self | Adopted Graph Projection behavior | Graph Projection tests | Medium | All eleven direct tests were owner-routed in GP-S00 |
| `internal/modules/graphprojection/direct_mapping.go` | Derives direct vertices, direct edges, and reverse edges | Package-private mapping engine | `project.go` | Owner-local types and canonical helpers | Engine and fixture tests | GP-FIX projected graph output | Graph Projection engine | High | Deterministic graph identity and direction semantics |
| `internal/modules/graphprojection/engine_test.go` | Characterizes admission, canonicalization, mapping, aggregation, filters, scalars, and failed runs | Nine parallel `Test*` functions and test helpers | Engine test-family row | Root engine internals | Self | Core GP-FIX behavior indirectly | Graph Projection tests | Medium | Primary pure engine evidence |
| `internal/modules/graphprojection/ephemeral_service.go` | Orchestrates non-retained projection and secure invocation nonce generation | `ProjectEphemeral` plus private nonce helper | Network Flow through `Service` | Root deterministic engine and crypto randomness | Service, facade, Network Flow tests | Ephemeral Graph result | Graph Projection application facade | Critical | Moved intact in GP-S02; active consumer path isolated |
| `internal/modules/graphprojection/error_adaptation.go` | Adapts repository sentinels into closed query and lifecycle errors | Package-private error helpers | Query, invalidation, retained lifecycle files | Root error contracts | Facade and repository-sentinel tests | Safe Graph error envelopes | Graph Projection application facade | Critical | Moved intact in GP-S02 |
| `internal/modules/graphprojection/errors.go` | Defines closed lifecycle/query errors, retryability, and ordered JSON envelopes | `LifecycleError`, `QueryError`, constructors, match helpers | Service, adapters, Network Flow, Reporting tests | Standard library | Facade, boundary, Network Flow, and store tests | Error envelope behavior; Network Flow maps owner errors | Graph Projection facade | High | Source-authored values must not leak |
| `internal/modules/graphprojection/facade_contract_test.go` | Characterizes error envelopes, validation registry, repository error mapping, cancellation, and inspection | Five `Test*` functions | Engine test-family row/full suite | Root facade and fake repositories | Self | Internal facade contract | Graph Projection tests | Medium | One test is missing from the focused row |
| `internal/modules/graphprojection/fixturetest/fixture.go` | Loads, validates, and compares owner fixtures with path and digest safety | Fixture manifest types, `Executor`, `Load`, `Verify`, comparison and path helpers | Graph Projection unit/store/candidate tests | Standard library filesystem and cryptography | `fixturetest/fixture_test.go` and root fixture tests | `cartulary.graph_projection_fixture_manifest.v3` and GP-FIX artifacts | Graph Projection test infrastructure | Medium | Owner-specific utility; it does not derive expected semantics |
| `internal/modules/graphprojection/fixturetest/fixture_test.go` | Verifies fixture integrity, semantic mismatch detection, and path containment | Three `Test*` functions | Fixture verifier test-family row | `fixturetest` | Self | Fixture manifest/golden integrity | Graph Projection tests | Low | Keep beside owner fixtures |
| `internal/modules/graphprojection/graph_projection_migration_test.go` | Verifies head-schema columns, keys, constraints, and graph table relationships | `TestGraphProjectionHeadSchemaContract_Integration` | Migration-reset test-family row | `pgtest` and PostgreSQL catalog | Self | `db/migrations/00022_graph_projection.sql` | Graph Projection storage tests | High | Database contract evidence only |
| `internal/modules/graphprojection/input.go` | Rejects malformed/duplicate JSON, materializes defaults, normalizes input, and computes admission identities | Package-private input sources and admission functions | Service and retained adapter helpers | Standard library JSON, time, Unicode, and owner types | Engine, remediation, fixtures, and facade tests | Graph Projection input shape and digest fixtures | Graph Projection engine | Critical | Whole-input admission precedes semantic projection |
| `internal/modules/graphprojection/invalidation_service.go` | Orchestrates retained graph-view and run invalidation and validates requests | Public invalidation facade methods plus private validation | Facade consumers and tests | Retained repository port and root validation | Facade and store tests | Retained invalidation contract | Graph Projection application facade | High | Moved intact in GP-S02 |
| `internal/modules/graphprojection/limits.go` | Holds closed NLSpec resource limits and exposes them to owner adapters | `ResourceLimits` returning owner-local limit registry | Engine and `postgresstore` | None outside owner module | Engine, remediation, and store fixtures | NLSpec section 4.12 typed behavior | Graph Projection engine/adapter contract | High | Mutable only for sequential test fixture overrides; adapter surface narrowing is deferred |
| `internal/modules/graphprojection/objects.go` | Builds normalized canonical objects used in configuration/source transcripts | Package-private normalization helpers | Admission and canonical digest code | Owner-local types | Engine and fixture tests | GP-FIX digest transcripts | Graph Projection engine | High | Exact omission/null/default behavior matters |
| `internal/modules/graphprojection/postgresbinding/reader.go` | Reads one retained projection binding through a caller-owned transaction | `Reader`, `NewReader`, `LookupProjectionBinding` | Reporting store | `pgx`, root Graph Projection binding types | `reader_test.go`, Reporting graph-reference tests | Reporting release binding contract | Graph Projection persistence adapter | High | Independent pool reads are forbidden by Reporting owner |
| `internal/modules/graphprojection/postgresbinding/reader_test.go` | Proves binding visibility in the caller transaction and expiry handling | `TestReaderSeesProjectionBindingInsideCallerTransaction` | Transaction-binding test-family row | `pgtest`, root/store packages | Self | Reporting binding semantics | Graph Projection storage tests | Medium | Cross-owner transaction evidence |
| `internal/modules/graphprojection/postgresrestore/writer.go` | Atomically replaces all five derived Graph tables and proves committed postconditions | `Writer`, `New`, and the `RestorePublisher` implementation | Recovery assembly through the Graph restore service | Borrowed `postgres.DB`, `pgx`, root restore publication contract | `writer_test.go` and Graph boundary guards | Exact five-table restore publication | Graph Projection recovery adapter | Critical | Separate from dormant `postgresstore.Store`; no source-owner SQL or schema scan |
| `internal/modules/graphprojection/postgresrestore/writer_test.go` | Proves clear-only, fresh available publication, stale-history exclusion, and rollback | Three GP-RA service-backed tests | Graph restore-publication row | `postgresrestore`, `pgtest`, root restore plan | Self | GP-RA-01, 02, 12, and 16 evidence | Graph Projection tests | High | Executes against migrated PostgreSQL and verifies all five tables |
| `internal/modules/graphprojection/postgresstore/cursor.go` | Defines list-cursor state and errors and encodes/validates opaque authenticated cursors with AES-GCM | Package-private cursor codec and cursor value | `pagination.go` | Standard library cryptography and root cursor errors | Store pagination tests | Retained list cursor wire semantics | Graph Projection persistence adapter | High | Token opacity, integrity, and version AAD are unchanged |
| `internal/modules/graphprojection/postgresstore/idempotency.go` | Owns retained lifecycle and invalidation replay fingerprints, reads, and records | Package-private transaction helpers | Lifecycle and invalidation files | `pgx`, canonical Graph helpers | Store lifecycle/invalidation tests | Graph idempotency table | Graph Projection persistence adapter | High | Moved intact in GP-S01 |
| `internal/modules/graphprojection/postgresstore/invalidation.go` | Implements graph-view and projection-run invalidation transactions | Public repository methods and private orchestration | Root Service and store tests | `pgx`, retention/idempotency helpers | Store lifecycle and fixture tests | Graph view/run/invalidation state | Graph Projection persistence adapter | Critical | Moved intact in GP-S01 |
| `internal/modules/graphprojection/postgresstore/lifecycle.go` | Implements retained create/refresh admission, locking, computation, and terminal failure coordination | Public repository methods and private lifecycle helpers | Root Service and store tests | `pgx`, Graph engine, publication/retention/idempotency helpers | Store lifecycle and fixture tests | Graph view/run lifecycle | Graph Projection persistence adapter | Critical | SQL, locks, transaction boundaries, hooks, and errors are unchanged |
| `internal/modules/graphprojection/postgresstore/pagination.go` | Implements retained graph-view listing and cursor continuation | `ListGraphViews` | Root Service and store tests | Cursor codec, Graph limits, `pgx` query port | Pagination tests | Graph view list contract | Graph Projection persistence adapter | High | Moved intact in GP-S01 |
| `internal/modules/graphprojection/postgresstore/publication.go` | Publishes terminal run, view, vertex, and edge state | Package-private transaction helper | Lifecycle | `pgx`, Graph resource JSON | Store lifecycle and fixture tests | Four Graph publication tables | Graph Projection persistence adapter | Critical | Publication SQL and atomic transaction placement are unchanged |
| `internal/modules/graphprojection/postgresstore/queries.go` | Implements run, view, vertex, edge, and readable-run queries | Public repository query methods and private resolver | Root Service and store tests | `pgx`, Graph DTO/error contracts | Query and binding-adjacent tests | Retained query behavior | Graph Projection persistence adapter | High | Moved intact in GP-S01 |
| `internal/modules/graphprojection/postgresstore/retention.go` | Owns expiry and replaced/failed/invalidated retention pruning | Package-private transaction helpers | Lifecycle, queries, invalidation | `pgx`, Graph retention policy | Retention and fixture tests | Graph run retention | Graph Projection persistence adapter | Critical | Ranking, timestamps, and deletion SQL are unchanged |
| `internal/modules/graphprojection/postgresstore/sql_values.go` | Converts empty strings and JSON bytes to SQL nulls | Package-private value helpers | Lifecycle and publication | None | Store lifecycle tests | SQL parameter semantics | Graph Projection persistence adapter | Medium | Moved intact in GP-S01 |
| `internal/modules/graphprojection/postgresstore/store.go` | Owns Store state, options, hooks, and construction | `Store`, `Hooks`, `Options`, `New` | Root integration tests; no production constructor found | Platform `postgres.DB`, cursor construction, clock | Store, fixture, migration, and binding tests | Constructor contract | Graph Projection persistence adapter | Critical | Reduced from 1,317 to 43 lines; RB-002 dormancy remains |
| `internal/modules/graphprojection/postgresstore/traversal.go` | Implements bounded graph traversal and deterministic helper ordering | `Traverse` and package-private traversal helpers | Root Service and store tests | Root Graph DTOs and sorting helpers | Traversal tests | Retained traversal contract | Graph Projection persistence adapter | High | Moved intact in GP-S01 |
| `internal/modules/graphprojection/project.go` | Coordinates admitted projection, validation, direct/aggregate derivation, output validation, ordering, and capabilities | Package-private project engine | Service and retained adapter helpers | Owner-local engine files and standard library | Engine, facade, Network Flow, fixture, and store tests | Graph result and Network Flow embedded result contracts | Graph Projection engine | Critical | Central deterministic orchestration |
| `internal/modules/graphprojection/query_service.go` | Orchestrates retained reads, inspection, list, traversal, and request validation | Public query facade methods plus private result/validation helpers | Facade consumers and tests | Retained repository port and root query contracts | Facade and store tests | Retained query/traversal contract | Graph Projection application facade | High | Moved intact in GP-S02 |
| `internal/modules/graphprojection/query.go` | Defines query sentinels, request/result DTOs, list summaries, inspection, and traversal contracts | Query errors and all query request/result types | Service, repository, stores, Reporting binding mapping | Standard library errors | Facade and store tests | Retained query and traversal behavior | Graph Projection facade | High | Keep independent of transport envelopes |
| `internal/modules/graphprojection/recovery_state.go` | Publishes five derived graph tables and rebuild algorithm identity to Recovery | `RecoveryStateContribution` | `internal/app/recoveryassembly` | Platform `recoverystate` | Recovery catalog tests outside target | Recovery state catalog and generated recovery projection | Graph Projection recovery contribution | Critical | GP-S07 resolves the algorithm through a separate Graph participant |
| `internal/modules/graphprojection/restore_contract.go` | Defines the Graph-owned restore request, typed source registry, implementation binding, result, limits, safe errors, and participant port | Restore request/result/registry/binding types, closed constants, constructors, and `RestoreParticipant` | Recovery's narrow `restorecontract` bridge and Graph restore service | Generated Recovery bindings and owner-local canonicalization | Restore contract/service and Recovery codec/application tests | Three Graph restore schemas, current/legacy binding projection, and protected completion evidence | Graph Projection recovery contract | Critical | GP-S06 contract boundary implemented by GP-S07 |
| `internal/modules/graphprojection/restore_contract_test.go` | Verifies exact current registry/binding projections, strict result states, limits, and safe errors | Four owner-routed `Test*` functions | Graph engine test-family row | Graph restore contract and generated Recovery binding | Self | Registry, binding, and rebuild-result contract evidence | Graph Projection tests | Medium | Added in GP-S06 and routed through the Graph owner slice |
| `internal/modules/graphprojection/restore_service.go` | Performs closed request admission, typed source enumeration, operation-wide accounting, deterministic derivation, publication, and safe result adaptation | `RestoreService`, options, staged plan/proof/error types, constructor, and `Rebuild` | Recovery assembly | Graph engine, restore registry/binding contract, narrow publisher port | `restore_service_test.go` and assembly tests | GP-RA result and semantic-output digest | Graph Projection recovery application service | Critical | All preflight work precedes persistent mutation; table IDs are returned by defensive copy |
| `internal/modules/graphprojection/restore_service_test.go` | Proves clear-only, candidates/skips, fail-before-mutation, all aggregate limits, retry determinism, indeterminate outcomes, and safe errors | Seven owner-routed GP-RA tests | Graph canonical-behavior row | Root restore service and fake typed capabilities/publisher | Self | GP-RA-01..11, 13, 15, and 17 evidence | Graph Projection tests | High | Large limits are checked without production-scale allocations through the shared accounting object |
| `internal/modules/graphprojection/repository.go` | Defines validated persistence-facing options, invalidation DTO, and retained repository port | `RetainedProjectionOptions`, `RetainedInvalidation`, `RetainedRepository` | Service and `postgresstore` | `context`, `time`, owner types | Facade and store tests | Internal retained lifecycle surface | Graph Projection application boundary | High | Correct dependency inversion; preserve signatures |
| `internal/modules/graphprojection/resources.go` | Serializes exact ordered public resource maps for views, runs, graph output, schemas, and validation | Package-private resource constructors | Service and error/facade code | Owner-local types | Facade, service, engine, Network Flow, and fixtures | Network Flow embedded Graph Projection result | Graph Projection facade | Critical | Member presence/order and safe details are observable |
| `internal/modules/graphprojection/retained_engine.go` | Exposes admission/projection/failure/canonical helper seams to owner child adapters | `AdmitRetainedProjection`, `DeriveGraphViewID`, `ProjectAdmittedRetainedProjection`, failure, canonical, sort, and transcript helpers | `postgresstore`, owner external-package tests, Network Flow via Service | Root engine internals | Store, fixtures, and Network Flow tests | Retained lifecycle and GP-FIX transcripts | Graph Projection adapter contract | High | Broad owner-internal surface; narrowing deferred pending cycle-safe design |
| `internal/modules/graphprojection/retained_lifecycle_service.go` | Orchestrates retained create/refresh, accepted summaries, Graph view identity, and idempotency-key validation | Public create/refresh/identity facade methods plus private lifecycle helpers | Facade consumers and tests | Retained repository port and root admission | Facade and store tests | Retained lifecycle facade | Graph Projection application facade | Critical | Moved intact in GP-S02 |
| `internal/modules/graphprojection/schema.go` | Validates closed JSON object/array/scalar shapes and classifies admission errors | Package-private schema definitions and validators | `input.go` | Standard library JSON/regexp | Engine, remediation, and fixture tests | Closed Graph Projection input schema behavior | Graph Projection engine | High | Unknown member and scalar classification are frozen |
| `internal/modules/graphprojection/service.go` | Owns stable public facade DTOs, `ServiceOptions`, `Service`, and construction | Request/result types, `ServiceOptions`, `Service`, `NewService` | Network Flow; tests; no production retained repository assembly found | Root repository contract and time source | Service, facade, store, and Network Flow tests | Internal facade construction and DTO shape | Graph Projection application facade | Critical | Reduced from 562 to 85 lines; every public signature is unchanged |
| `internal/modules/graphprojection/service_test.go` | Proves ephemeral output is complete and creates no retained state | `TestProjectEphemeralReturnsFullNonRetainedResult` | Engine test-family row | Root Service | Self | Network Flow ephemeral consumer behavior | Graph Projection tests | Medium | Direct characterization of the active production path |
| `internal/modules/graphprojection/store_fixture_test.go` | Supplies external-package graph input and store fixture helpers | Test-only helpers | Store fixture and store tests | Public root Graph Projection surface | Root store tests | GP-FIX store inputs | Graph Projection tests | Low | Duplicate-looking helpers are intentionally external-package contract fixtures |
| `internal/modules/graphprojection/store_test.go` | Characterizes retained lifecycle, retention, invalidation, service facade, query states, pagination, and traversal | Seven service-backed `Test*` functions | Storage lifecycle test-family row | `postgresstore`, `pgtest`, root facade | Self | Retained lifecycle/query/storage behavior | Graph Projection storage tests | Critical | Primary persistence characterization |
| `internal/modules/graphprojection/types.go` | Defines closed projection input, configuration, lifecycle, graph, schema, validation, capability, and invalidation types | `ProjectionSchemaID`, `Optional`, state enums, and all public contract structs | All root engine/facade files, owner adapters, Network Flow/Reporting consumers | `time` | All engine, facade, fixture, consumer, and store tests | Network Flow Graph Projection result vocabulary | Graph Projection domain contract | Critical | Broad but semantically coherent owner model |
| `internal/modules/graphprojection/unit_fixture_executor_test.go` | Executes unit fixture operations and applies temporary sequential resource-limit overrides | Test-only executor and helpers | Unit fixture wrapper and candidate test | `fixturetest`, root engine internals | Root fixture tests | GP-FIX unit artifacts | Graph Projection tests | Medium | Test-only mutation is restored; no parallel fixture execution is present |
| `internal/modules/graphprojection/validation.go` | Validates admitted mappings, filters, aggregations, properties, metadata, resources, and projected output | Package-private validation functions | `project.go` | Standard library and owner types | Engine, remediation, fixtures, and facade tests | Validation issue behavior and resource fixtures | Graph Projection engine | Critical | Validation ordering and error codes are observable |
| `internal/modules/graphprojection/validation_registry.go` | Owns the closed issue-code/severity/detail registry, issue IDs, ordering, truncation, and summary cap | Package-private validation registry | Validation/project/resources | Standard library | Facade, engine, remediation, and fixture tests | Validation summary and issue contracts | Graph Projection engine | Critical | Exact issue identity and safe detail policy are frozen |

## 3. Module Boundary Diagnosis

The live target is a legitimate Graph Projection subsystem, a view/projection orchestration layer, an internal application/service facade, and a persistence-adjacent adapter family. It is also a mixed-responsibility package at the file level. It is not a transport adapter, frontend shell, grid-vendor layer, workbook projection owner, or mutation coordinator for authoritative source state.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Graph input admission, normalization, validation, identity, and deterministic derivation | Root engine files | `module.graphprojection` | keep | Adopted Graph Projection NLSpec and engine/fixture tests | Central owner behavior |
| Ephemeral and retained lifecycle/query application facade | `service.go`, `query.go`, `repository.go` | `module.graphprojection` | split | Service API, Network Flow consumer, store tests | Split files by concern; do not change package or API |
| Retained graph persistence | `postgresstore` | `module.graphprojection` persistence adapter | split | `RetainedRepository`, migration 00022, store tests | Keep direct SQL confined to graph tables |
| Reporting transaction binding | `postgresbinding` and Reporting store | Graph Projection supplies binding; Reporting owns admission/error mapping | keep | Reporting REQ-RPT-076 and same-transaction code/test | Intentional cross-owner seam |
| Recovery state declaration and rebuild | `recovery_state.go` plus missing runtime seam | Graph Projection owns semantics; Recovery coordinates | defer | Graph Projection NLSpec section 11.9 and recovery catalog | Catalog is present; behavior implementation is RB-001 |
| Owner fixture loading and verification | `fixturetest` | `module.graphprojection` test infrastructure | keep | Fixture schema, GP-FIX corpus, verifier tests | Do not move semantic expectations into generic test-util |
| Public graph query transport and authorization | Network Flow module | `module.networkflow` | keep | Network Flow routes, handlers, owner NLSpec, generated schema | Graph Projection exposes no public route |
| Frontend graph selection and presentation | `apps/web/src/networkFlow` | `web.networkflow` | keep | Controller/workspace and browser rows | Indirect consumer only; no Graph Projection frontend package |
| Reporting release tuple and graph reference failures | Reporting module | `module.reporting` | keep | Reporting owner and store/tests | Do not move release semantics into Graph Projection |
| Workbook projection/view-schema/saved-view behavior | Other named modules | Projections, view-schema owners, Saved Views | defer | Core/domain vocabulary and boundary guards | No live target evidence supports movement |
| Timeline, imports, entities, indicators, evidence, links, and collaboration behavior | Other named modules | Their adopted owners | defer | Root boundary guard shows no sibling-module imports | Explicit input bytes avoid source-owner coupling |
| Grid-vendor integration | None in target | Frontend grid adapter if later applicable | defer | No target or consumer import found | Filename/UI labels are not ownership evidence |

Planning finding: the framework warned that an existing directory may be an invalid permanent boundary. Live evidence instead supports retaining `graphprojection` as the permanent subsystem boundary while decomposing large concern-heavy files inside it. No implementation decision to move owner behavior is made by this tracker.

### Current production composition

The following matrix records current facts. It does not infer activation from implemented or tested capabilities.

| Surface | Current state | Exact evidence | Required preservation posture |
| --- | --- | --- | --- |
| Deterministic Graph Projection engine | Active | Network Flow constructs the root `Service` and calls `ProjectEphemeral` | Preserve inputs, deterministic identities, validation, output, and safe errors |
| Network Flow ephemeral graph path | Active only when the Network Flow profile is claimed | `internal/modules/networkflow/graph_projection_adapter.go` and Network Flow route ownership | Keep authorization and transport in Network Flow |
| Reporting `ProjectionBindingReader` | Active when Reporting validates graph references | Reporting constructs `postgresbinding.NewReader(tx)` inside its transaction | Preserve the borrowed transaction and Reporting-owned result mapping |
| Full retained `postgresstore.Store` | Not constructed by any discovered production caller | Production Go caller search found no `postgresstore.New` call | Do not delete or activate without an adopted owner change |
| Retained Graph Projection HTTP or WebSocket transport | Absent | Boundary test plus route and WebSocket contract search | Preserve absence |
| Retained lifecycle worker or scheduler | Absent | Application assembly and target-package inspection | Do not add a hidden worker or lifecycle hook |
| Retained cursor-key deployment configuration | Absent | Configuration and application-assembly search | Do not add a placeholder key or activation flag |
| Retained-view restore source registry | Present and intentionally empty | Generated registry artifact and `CurrentRestoreSourceRegistry` project exactly `entries=[]` | Add no source registration without a frozen owner contribution and new binding |
| Recovery contribution | Active catalog metadata | `RecoveryStateContribution` supplies five tables and `graphprojection.restore_rebuild.v1` | Preserve exact table and algorithm resolution |
| Recovery clear/rebuild port and implementation | Active only during admitted vNext restore/verification | Recovery imports the narrow `restorecontract` bridge; assembly constructs `RestoreService` with `postgresrestore.Writer`; workbook rebuild remains separate | Preserve exact participant order, terminal evidence, target lease, and ordinary retained-store dormancy |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Internal Go `Service` facade and request/result types | Graph Projection | `service.go`, adopted owner | Service, facade, store, Network Flow tests | Preserve every method signature and error mapping | Critical | Only ephemeral construction is found in production |
| Ephemeral projection result and no retained side effects | Graph Projection; Network Flow owns transport embedding | `ProjectEphemeral`, Network Flow adapter/schema | `TestProjectEphemeralReturnsFullNonRetainedResult`, Network Flow graph tests | Preserve zero-issue success, exact resource, cancellation/timeout mapping | Critical | Active production consumer path |
| `POST /api/v1/incidents/{incident_id}/network-flow/graphs/query` | Network Flow | `contracts/network-flow/routes.v1.json`, handler and schema | Network Flow unit/integration/browser rows | Preserve request/result envelope, viewer authorization, audit event, and error codes | Critical | Not a public Graph Projection route |
| `POST /api/v1/incidents/{incident_id}/network-flow/graphs/contributors/query` | Network Flow | Route catalog, handler, frontend drawer | Route integration and two selection browser rows | Preserve stable vertex/edge identity and recomputed selector semantics | High | Indirectly depends on projection identity |
| Network Flow generated `GraphProjectionEphemeralResult` and TypeScript types | Network Flow contract owner | `contracts/network-flow/schemas.v1.json`, generated TS | Contract tests and frontend unit/browser tests | Drift and Network Flow owner rows if shape changes | Critical | Generated files are read-only outputs |
| Retained create/refresh lifecycle and publication | Graph Projection | Service, repository, store, migration | Store lifecycle and GP-FIX integration tests | Preserve state transitions, one-active-run lock, previous-run selection, output publication | Critical | No production writer assembly found |
| Retention, idempotency, invalidation, and failed-run inspection | Graph Projection | Store and service implementations | Store, facade, remediation, retained fixtures | Preserve expiry/count ranking, replay fingerprints, selected run, safe failure details | Critical | Exact storage semantics |
| Direct graph queries, list cursor, and traversal | Graph Projection | Query DTOs, cursor, store | Query matrix, pagination, traversal tests | Preserve state eligibility, cursor integrity, ordering, bounds | High | Retained-only query surface |
| Canonical JSON, tuple transcripts, IDs, and digests | Graph Projection | Canonical/input/object code and fixtures | Engine and GP-FIX-014/022/023/036 | Byte-for-byte characterization remains required | Critical | No alternate canonicalization allowed |
| Validation registry, issue ordering, limits, and safe details | Graph Projection | Validation files and adopted owner | Engine, facade, remediation, fixture tests | Add omitted tests to focused owner accounting | Critical | RB-003 affects focused evidence completeness |
| Reporting `source_projection_ref.v1` binding | Reporting owns tuple; Graph Projection owns binding facts | Reporting REQ-RPT-076..080, root binding, transaction reader | Reporting graph ref test, reader integration test | Preserve available/replaced acceptance, snapshot/digest checks, same transaction | Critical | Reporting maps owner errors to Reporting reasons |
| Five Graph Projection tables and migration shape | Graph Projection | Migration 00022, store SQL, recovery catalog | Migration head-schema and store integration tests | Preserve table names, constraints, indexes, and graph-only SQL | Critical | No migration is proposed |
| Recovery contribution and `graphprojection.restore_rebuild.v1` | Graph Projection; Recovery coordinates | Owner section 11.9, `recovery_state.go`, recovery fixture | Recovery catalog tests outside target | TODO: characterize clear/rebuild once an implementation seam exists | Critical | Current implementation gap RB-001 |
| GP-FIX-001 through GP-FIX-036 and candidate mode | Graph Projection verification | Fixture manifests, `fixturetest`, candidate Make target | Unit/store wrappers and verifier tests | Preserve wrapper coverage; account for direct remediation tests | High | Candidate output remains disposable |
| Authorization outcomes | Network Flow and Reporting route owners | Network Flow handler membership checks; Reporting admission | Consumer route/integration tests | Preserve caller-owned denial precedence; Graph Projection adds no auth | High | Owner NLSpec explicitly does not add auth semantics |
| Public Graph Projection HTTP/WS routes | No public surface | Boundary test, OpenAPI and WS search | `TestNoPublicGraphProjectionRoutes` | Preserve absence | High | Adding one requires later authorization |
| Direct entity row/query/mutation behavior | Entity/source owners, not Graph Projection | Explicit-input-only boundary guard | Boundary test | None unless a later adapter is proposed | Low | Not applicable to current target |
| Saved views, workbook view schema, revision/change-set mutation, and grid adapter | Named external owners | Core/domain and negative import evidence | Boundary tests in relevant owners | None for the planned internal splits | Low | Not present in target |

### Adopted owner package

GP-S04 adopted the requirements summarized in this subsection through Graph Projection NLSpec 1.2.0 and narrowly applicable Core 00, Core 01, and Core 04 amendments. The owner text governs if this summary differs. The existing Core 01 workbook `RestoreProjectionRebuilder`, `restore_projection_rebuild_request_v1`, `restore_projection_rebuild_result_v1`, and provider registry remain unchanged. Graph Projection is not inserted into the workbook-provider registry; Recovery resolves `graphprojection.restore_rebuild.v1` to a distinct Graph-owned internal port.

#### Durable document placement

| Artifact | Adopted or projected content | Content excluded from that owner |
| --- | --- | --- |
| Core 00 | Updated Graph Projection adoption metadata, if the owner version changes, and the REQ-00-062 projection re-audit result | Graph SQL, registry fields, implementation mechanics, or test symbols |
| Core 01 | Recovery catalog resolution, adaptation into the Graph-owned port, readiness aggregation, durable terminal replay, and existing `projection_rebuild_failed` mapping | Graph mapping, validation, retention-validity semantics, or graph-table SQL |
| Core 02 | No current change; any future source owner must classify authoritative retained-view declaration state for backup | Graph views, runs, vertices, edges, or idempotency rows as authoritative inputs |
| Core 03 | No change | A workbook surface, Graph route, or browser restore control |
| Core 04 | Reuse of `target_generation_id`, target admission, serving lease, fail-before-touch posture, indeterminate-outcome reinitialization, and safe journal/audit rules | A current Graph cursor-key placeholder or second generation identity |
| Graph Projection NLSpec | Source registry, validity, exact five-table clear/rebuild, binding, readiness, safe result details, acceptance matrix, and current-profile retained-store dormancy | Recovery CLI grammar, operator authorization, or workbook-provider semantics |
| Testing Harness NLSpec | No RB-003 change; the existing exact-selector rules remain sufficient | Graph-specific test names or a special-case execution rule |
| Recovery machine contracts | Graph recovery contribution, registry and implementation-binding digests, exact algorithm resolution, and safe result adaptation | Documentation hashes, dynamic code loading, or unrestricted source discovery |
| `tools/test_families` | The eleven exact selectors added to the existing Graph engine row | Globs, title prefixes, raw commands, or duplicated fixture functions |
| Appendix I or implementation guides | Non-normative diagrams, rationale, package placement, lifecycle guidance, and diagnostic procedure | Sole ownership of defaults, required fields, readiness, or failure semantics |
| This tracker | Decision state, authorization boundary, likely files, validation, run roots, and remaining risk | Alternate product authority or a substitute for adopted owner text |

Core 05 remains inapplicable because this plan creates no timed or fixture-sensitive publication claim.

#### Internal request contract

Graph Projection NLSpec 1.2.0 defines `cartulary.graph_projection_restore_rebuild_request.v1` as an internal contract with the following semantic members. It is not an HTTP object, public protocol family, or extension capability.

| Member | Required | Nullable | Default | Proposed requirement |
| --- | ---: | ---: | --- | --- |
| `restore_operation_id` | Yes | No | None | Exact non-empty opaque Recovery operation identity |
| `restored_source_state_ref` | Yes | No | None | Immutable Recovery-owned typed capability over restored authoritative state; a raw DSN, schema scanner, or unrestricted query handle is invalid |
| `backup_set_id` | Yes | No | None | Exact selected backup identity |
| `consistency_point_at` | Yes | No | None | Exact backup consistency point; all registration, declaration, and retention validity is evaluated at this instant |
| `target_generation_id` | Yes | No | None | Exact Core 04 restore-target marker value; Graph MUST NOT mint another generation identity |
| `recovery_state_catalog_ref` | Yes | No | None | Exact frozen catalog selected from the backup manifest |
| `source_registry_ref` | Yes | No | None | Exact Graph restore registry snapshot selected from the backup manifest |
| `implementation_binding_ref` | Yes | No | None | Exact packaged Graph restore binding selected from the backup manifest |
| `context` | Yes | No | None | Recovery-supplied cancellation and deadline context; it is an invocation capability, not a serialized wire member |

Recovery MAY adapt one admitted restore operation into this request only after the frozen recovery catalog resolves the exact Graph algorithm. Request validation MUST complete before persistent Graph state changes.

#### Internal result contract

Graph Projection NLSpec 1.2.0 defines `cartulary.graph_projection_restore_rebuild_result.v1` with exactly the following semantic members.

| Member | Required | Nullable | Default | Proposed requirement |
| --- | ---: | ---: | --- | --- |
| `restore_operation_id` | Yes | No | None | Exact accepted request value |
| `target_generation_id` | Yes | No | None | Exact accepted request value |
| `status` | Yes | No | None | Exactly `succeeded` or `failed`; Graph has no `not_applicable` state |
| `readiness_outcome` | Yes | No | None | Exactly `ready` or `incomplete` |
| `algorithm_id` | Yes | No | None | Exactly `graphprojection.restore_rebuild.v1` |
| `implementation_binding_sha256` | Yes | No | None | Exact admitted binding digest |
| `source_registry_sha256` | Yes | No | None | Exact admitted registry digest |
| `cleared_table_ids[]` | Yes | No | `[]` on failure | On success, exactly `graph_projection_edges`, `graph_projection_idempotency`, `graph_projection_runs`, `graph_projection_vertices`, `graph_projection_views` in this ASCII order |
| `rebuilt_views[]` | Yes | No | `[]` | Successful rebuild results in ascending `source_registration_id`, then `candidate_id` ASCII order |
| `skipped_candidates[]` | Yes | No | `[]` | Omitted candidates carrying one closed skip reason in canonical order |
| `postcondition_sha256` | Yes | Yes | `null` on failure | On success, SHA-256 of the canonical committed completion summary |
| `warnings[]` | Yes | No | `[]` | Closed safe warnings only |
| `errors[]` | Yes | No | `[]` on success | Closed safe error summaries only; non-empty on failure |

Result state combinations MUST be closed as follows:

| `status` | `readiness_outcome` | Cleared/rebuilt/skipped claims | `postcondition_sha256` | `errors[]` |
| --- | --- | --- | --- | --- |
| `succeeded` | `ready` | Five cleared tables plus canonical rebuilt/skipped arrays, which may be empty | Non-null committed-summary digest | Empty |
| `failed` | `incomplete` | All three arrays empty; the result MUST NOT claim committed facts after rollback or an indeterminate outcome | `null` | Non-empty closed safe errors |

Every other status/readiness combination is invalid. Graph may return its participant result only after its database outcome is determinate and its local postconditions pass. Recovery MUST NOT publish overall restore readiness until it has durably written the exact accepted Graph result with the remaining terminal evidence. An indeterminate Graph commit never returns a successful participant result and forces target reinitialization.

Each `rebuilt_views[]` entry MUST contain exactly the safe completion facts `source_registration_id`, `candidate_id`, `graph_view_id`, `projection_run_id`, `source_snapshot_id`, `projection_version`, `normalized_configuration_sha256`, `normalized_source_sha256`, `vertex_count`, `edge_count`, and `canonical_output_sha256`. Results, warnings, errors, logs, telemetry, journals, and audit summaries MUST NOT contain raw source entities or relationships, configuration JSON, incident-authored values, SQL, database errors, credentials, object keys, source capability internals, or unrestricted source references.

#### Restore source registry

Graph Projection NLSpec 1.2.0 defines `cartulary.graph_projection_restore_source_registry.v1` as a code-backed registry assembled from typed source-owner contributions. A JSON projection supports closed-shape validation and drift review but does not load runtime callbacks or replace the executable registry.

| Entry member | Required | Nullable | Proposed requirement |
| --- | ---: | ---: | --- |
| `source_registration_id` | Yes | No | Stable globally unique identity |
| `source_owner_id` | Yes | No | Module or adopted subsystem owning the authoritative inputs |
| `enumerator_binding_id` | Yes | No | Exact packaged enumerator of retained-view declarations from restored authoritative state |
| `validity_binding_id` | Yes | No | Exact packaged validity evaluator at `consistency_point_at` |
| `projection_input_contract_id` | Yes | No | Versioned normalized Graph input contract emitted by the source owner |
| `implementation_binding_id` | Yes | No | Binding included in the Graph restore implementation digest |
| `status` | Yes | No | Exactly `active` or `retired` |
| `introduced_at` | Yes | No | Earliest consistency point at which the registration can be active |
| `retired_at` | Yes | Yes | Exact retirement boundary, or `null` while not retired |

Entries and enumerated candidates MUST be unique and ordered by ascending ASCII bytes of their stable identities. Duplicate registrations, candidate IDs, or graph-view identities; unknown bindings; or unstable enumeration MUST fail before mutation. The adopted exact current-profile default is:

```json
{"schema_id":"cartulary.graph_projection_restore_source_registry.v1","entries":[]}
```

An empty registry MUST still run the mandatory five-table clear and, after committed postcondition checks and durable Recovery completion, return `status='succeeded'`, `readiness_outcome='ready'`, all five `cleared_table_ids`, `rebuilt_views=[]`, and `skipped_candidates=[]`. It MUST NOT return `not_applicable` and MUST NOT preserve stale Graph rows.

Any future producer MUST first own authoritative declaration state outside the five derived Graph tables. That state MUST be backup-classified and sufficient to reconstruct stable view identity, normalized configuration or exact reconstruction inputs, source-snapshot selection, source entities and relationships, validity at the backup consistency point, and owner retention. Graph views, runs, and idempotency rows MUST NOT become reconstruction authority.

#### Implementation binding

The owner and Recovery machine-contract changes MUST define `cartulary.graph_projection_restore_implementation_binding.v1` as a closed canonical object with these required members:

| Member | Proposed requirement |
| --- | --- |
| `schema_id` | Exactly `cartulary.graph_projection_restore_implementation_binding.v1` |
| `algorithm_id` | Exactly `graphprojection.restore_rebuild.v1` |
| `binding_id` | Stable exact packaged binding identity |
| `graph_projection_contract_id` | Exact adopted Graph Projection contract/version identity |
| `recovery_state_catalog_sha256` | Exact frozen catalog digest |
| `source_registry_sha256` | Exact registry digest |
| `graph_table_ids[]` | The exact five table identities in canonical order |
| `graph_engine_algorithm_ids[]` | Closed ordered identities for all participating Graph algorithms |
| `graph_engine_algorithm_digests[]` | Digests aligned one-to-one with the algorithm identities |
| `database_schema_lineage` | Exact admitted database lineage identity |
| `database_schema_head` | Exact admitted schema head |
| `packaged_subject_sha256` | Digest of the packaged executable subject |
| `build_provenance_sha256` | Digest of the admitted build-provenance object |

The binding digest MUST be SHA-256 over the repository's canonical JSON serialization of that closed object. It MUST NOT include a documentation hash, repository or working-directory path, source line range, mutable branch, `current`/`latest` alias, or raw build log. The backup manifest MUST retain the exact source-registry and implementation-binding digests. A historical backup whose exact registry or implementation is unavailable MUST fail closed and MUST NOT be reinterpreted with current code.

#### Candidate validity and failure classification

A candidate MAY be rebuilt only when every condition below is true at `consistency_point_at`, not at restore wall-clock time:

1. Its source registration was active.
2. Its owner-controlled retained-view declaration existed and was valid.
3. Every required authoritative source input is present in restored state.
4. Its exact historical registry and implementation binding remain packaged.
5. Its normalized projection input passes admission under the exact historical contract selected by the backup.
6. Any owner expiration or retention boundary had not elapsed.

The only adopted skip reasons are `registration_not_yet_active`, `registration_retired_before_consistency_point`, and `declaration_expired_at_consistency_point`. A missing source input, unavailable historical implementation, schema mismatch, invalid normalized input, resource overflow, unresolved declaration, catalog mismatch, or digest mismatch MUST be a failure, not a skip. A successful rebuild MUST create exactly one fresh selected `available` run per valid view and MUST create no `replaced`, `failed`, `invalidated`, historical `available`, or idempotency row. Normal run retention applies only to runs created after restore.

The adopted invocation-wide limits are 128 registrations, 1,024 candidates, 268,435,456 normalized input bytes, 500,000 vertices, and 1,000,000 edges.

The exact completion proof tuple is:

```text
target_generation_id
restore_operation_id
backup_set_id
consistency_point_at
recovery_state_catalog_sha256
source_registry_sha256
implementation_binding_sha256
postcondition_sha256
```

Recovery MUST retain this tuple in its protected terminal evidence. Graph MUST NOT introduce a second generation identity or caller-selectable generation input.

#### Clear-and-rebuild algorithm

Graph Projection NLSpec 1.2.0 prescribes these phases in order.

**Phase A: fail before touch.** Before opening the publication transaction, the implementation MUST validate the request and cancellation state; verify the exact catalog, five tables, and algorithm; verify registry and binding digests; resolve every enumerator and validity evaluator; enumerate candidates canonically; reject duplicates and unstable ordering; validate and normalize every candidate; derive every candidate in bounded memory or transaction-local staging; and verify limits, identifiers, validation summaries, output order, and canonical digests. No persistent Graph row may change during Phase A.

**Phase B: atomic publication.** In one database transaction, the implementation MUST clear exactly the five Graph tables using an explicit `TRUNCATE TABLE graph_projection_edges, graph_projection_idempotency, graph_projection_runs, graph_projection_vertices, graph_projection_views RESTRICT`; insert fresh view, run, vertex, and edge rows for all valid candidates; insert no idempotency rows; and verify that only newly published state exists, no run is `accepted` or `computing`, every rebuilt view has exactly one selected `available` run, every digest and count matches staging, referential invariants hold, and idempotency is empty. It MUST commit before constructing success from committed facts. Recovery then owns terminal journal and administrative-audit publication.

**Phase C: terminal outcome.** Outcomes MUST map exactly as follows:

| Outcome | Required consequence |
| --- | --- |
| Preflight failure | No Graph rows touched; corrected retry permitted |
| Known transaction rollback | No Graph rows committed; retry permitted while Recovery admission remains valid |
| Successful commit and durable terminal evidence | Graph participant is ready when all Graph postconditions pass; overall restore readiness remains Recovery-owned |
| Unknown commit outcome, cancellation during mutation, timeout, exclusive-lease loss, or journal failure after an unprovable commit | Readiness remains false and the target MUST be reinitialized before reuse |
| Lost caller response after durable terminal evidence | Recovery replays the terminal result and MUST NOT invoke a second rebuild merely to reproduce the response |

A retry after a known rollback MAY use a different `projection_run_id` because the adopted Graph identity includes a run nonce. Deterministic vertex IDs, edge IDs, normalized configuration and source digests, and canonical output digests MUST remain equal. Prior Graph idempotency state MUST NOT prove restore replay.

The Graph participant result may be `ready` only after exact catalog, registry, and binding admission; determinate committed five-table clearing; empty idempotency; no incomplete run; one fresh selected available run for every valid candidate; no skipped active valid candidate; matching deterministic digests and counts; valid references; and matching target generation. Recovery's overall restore readiness additionally requires durable terminal evidence. Failure of any Graph condition yields `failed` and `incomplete`; terminal-evidence failure retains Recovery's existing closed failure mapping and MUST NOT be rewritten as Graph success.

#### Restore acceptance matrix

The owner and implementation slices MUST supply owner-routed evidence for all scenarios below. These are adopted acceptance criteria; they become implementation completion evidence only when their tests pass.

| ID | Scenario | Required result |
| --- | --- | --- |
| GP-RA-01 | Stale rows in all five tables; registry empty | All five tables empty; `succeeded`, `ready`, empty rebuilt/skipped arrays |
| GP-RA-02 | One active valid declaration | Exactly one fresh selected available run and expected deterministic graph output |
| GP-RA-03 | Declaration expired at consistency point | Skip with `declaration_expired_at_consistency_point`; no rows for the candidate |
| GP-RA-04 | Registration introduced after consistency point | Skip with `registration_not_yet_active`; no rows for the candidate |
| GP-RA-05 | Registration retired before consistency point | Skip with `registration_retired_before_consistency_point`; no rows for the candidate |
| GP-RA-06 | Active declaration missing a required source input | Fail before clear |
| GP-RA-07 | Source-registry digest mismatch | Fail before clear |
| GP-RA-08 | Implementation-binding digest mismatch or unavailable historical binding | Fail before clear |
| GP-RA-09 | Recovery catalog or five-table identity mismatch | Fail before clear |
| GP-RA-10 | Invalid normalized projection input or schema mismatch | Fail before clear |
| GP-RA-11 | Resource-limit overflow | Fail before clear |
| GP-RA-12 | Failure after clear and before commit | Full transaction rollback; no partial new state |
| GP-RA-13 | Commit outcome unknown, cancellation during mutation, timeout, or serving-lease loss | Target not ready; mandatory reinitialization |
| GP-RA-14 | Durable success followed by lost response | Replay terminal evidence without a second rebuild |
| GP-RA-15 | Retry after known rollback | Run ID may differ; deterministic identities and output digest equal |
| GP-RA-16 | Historical runs and idempotency existed before restore | Neither history nor idempotency is reconstructed |
| GP-RA-17 | Source or dependency errors contain sensitive values | Result, log, telemetry, journal, and audit expose only closed safe classifications |
| GP-RA-18 | Assembly and source enumeration boundaries | Recovery imports only the Graph restore port; Graph uses typed source capabilities and no unrestricted cross-owner SQL or schema scan |

#### Proposed retained-store dormancy

Graph Projection NLSpec 1.2.0 declares the full retained `postgresstore.Store` intentionally dormant in ordinary current-profile production composition. The package and conformance tests remain available, but ordinary startup MUST NOT construct the Store, resolve a retained cursor key, start a Graph retention worker, publish a Graph route, create retained graph views, or offer a configuration flag that activates them. Network Flow continues to use ephemeral projection, Reporting continues to use only its same-transaction binding reader, and Recovery uses only a narrow Graph restore port. Dormancy does not exempt the five tables from the mandatory clear/rebuild action; with the exact empty registry the action is clear-only.

Future activation MUST remain blocked until one adopted change supplies every item below:

| Gate | Required owner decision and evidence |
| --- | --- |
| Named producer and consumer | Identify who creates retained views and who reads them |
| Authoritative declaration state | Persist reconstructable source-owner state outside Graph-derived tables |
| Source registration | Supply exact enumerator, validity, input-contract, and implementation bindings |
| Application composition | Name constructor ownership, startup/shutdown order, and dependency cleanup |
| Authorization | Preserve Graph's authorization-agnostic boundary and require caller authorization |
| Transport | Define a route or job only if required; activation MUST NOT imply transport |
| Resource control | Define admission/query concurrency, transaction bounds, retention work, and backpressure |
| Cursor key | Adopt the exact secret-reference, wire-version, and rotation contract before query pagination activation |
| Recovery | Add registry, binding, backup, restore, and historical-implementation evidence |
| Operations | Add startup, shutdown, cancellation, rotation, Recovery, and failure-path tests |

No current Graph cursor key exists. If pagination is later activated, the adopted gate requires a dedicated 32-byte AES-256 key obtained through a secret reference; purpose isolation; one issuing key plus a bounded prior verification ring; an authenticated stable non-secret key ID and algorithm version; unique nonces per key; issuance only with the current key; prior-key acceptance only through the bounded cursor lifetime; no secret disclosure; fail-closed resolution and key-size validation; and no unbounded trial decryption. If the current envelope lacks an authenticated key ID and algorithm version, activation MUST use a new cursor wire version.

Any future retained worker MUST accept a borrowed database capability, MUST NOT close it, MUST make key resolution explicit and fallible, MUST start no hidden goroutine in `New`, and MUST expose explicit bounded `Start(ctx)`/`Stop(ctx)` behavior. Shutdown MUST stop admission, cancel workers, wait within the declared bound, release dependencies in reverse ownership order, and preserve the primary error when cleanup also fails.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Recovery catalog declared rebuildable graph state without executing owner behavior | GP-S07 Graph participant, narrow writer, Recovery coordination, terminal evidence, and GP-RA tests | Resolved; restore now clears the five tables and invokes typed current/future source registrations before readiness | `resolved` | Graph Projection semantics; Recovery coordination | RB-001 closed with full Graph/Recovery unit and service-backed evidence |
| Baseline `postgresstore/store.go` combined all retained persistence concerns in 1,317 lines | GP-S01 same-package decomposition and passing unit/service-backed evidence | Resolved; concern changes now have smaller review surfaces | `resolved` | Graph Projection persistence adapter | Constructor remains in 43-line `store.go`; behavior lives in nine cohesive concern files plus `cursor.go` |
| Baseline `service.go` combined retained lifecycle, queries, invalidation, error adaptation, and ephemeral orchestration | GP-S02 same-package decomposition and passing Graph/Network Flow evidence | Resolved; active ephemeral and dormant retained paths have separate review surfaces | `resolved` | Graph Projection application facade | DTOs and constructor remain in 85-line `service.go`; concern behavior lives in five cohesive files |
| Focused unit routing omits 11 non-wrapper contract tests | Source symbols compared with `tools/test_families/module.graphprojection.json` | `make test-slice` does not represent all direct contract evidence | `should_fix` | Graph Projection verification owner | GP-S00 adds exact selectors to the authored row and regenerates downstream topology |
| Adapter helpers are exported from `retained_engine.go` and `limits.go` | GP-S03 confirmed only `postgresstore` and owner tests consume them; boundary guard rejects other consumers | Internalization would create cycles, aliases, duplicated engine logic, or a disproportionate type extraction | `intentional/no_action` | Graph Projection | DONE: preserve the owner-child seam and rely on boundary guards |
| Retained store is integration-tested but intentionally absent from production composition | GP-S05 owner-routed production/configuration guard plus Network Flow and Reporting evidence | Resolved for current profile; future activation remains separately adopted | `resolved` | Graph Projection plus application assembly | Guard forbids construction, options/hooks, cursor key, hidden flag, route/WS, retained Network Flow use, Reporting full Store, and Recovery full Store |
| Restore and dormancy clauses require adopted owner behavior | Graph Projection NLSpec 1.2.0 and amended Core 00/Core 01/Core 04 | A tracker must not become alternate conformance authority | `resolved` | Graph Projection, Core 01/Core 04, Recovery machine-contract owners | GP-S04 adopted the clauses and completed the REQ-00-062 re-audit |
| Recovery currently constructs the workbook Projections rebuilder, not a Graph rebuild port | `internal/app/operator/operator_recovery.go`, `internal/app/projectionassembly`, and Recovery restore contracts | Coupling Graph to the workbook registry would erase ownership and change existing generic semantics | `intentional/no_action` | Recovery coordination and separate named projection owners | Preserve the generic adapter; GP-S06 adds a catalog-resolved Graph-specific port |
| Network Flow imports the root Service for ephemeral projection | Network Flow owner section 14 and exact adapter | Cross-owner errors or output shape could leak | `intentional/no_action` | Network Flow consumer; Graph Projection provider | Preserve the narrow service adapter and Network Flow-owned error mapping |
| Reporting imports root binding types and `postgresbinding` | Reporting owner requires same-transaction validation | Independent reads would violate snapshot admission | `intentional/no_action` | Shared seam with named owners | Preserve the transaction-borrowing reader and Reporting reason mapping |
| Root recovery contribution imports platform `recoverystate` | Repository procedure says source owners construct contributions | Moving it to Recovery would erase source ownership | `intentional/no_action` | Graph Projection contribution | Keep the contribution; add runtime behavior only with authorization |
| `postgresstore` imports platform PostgreSQL and uses direct SQL | Imports are confined to adapter child package; boundary test limits table names | SQL leakage into root would weaken the facade | `intentional/no_action` | Graph Projection persistence adapter | Keep root facade free of PostgreSQL and SQL restricted to graph tables |
| Frontend displays and selects embedded projection vertices/edges | Network Flow controller/workspace and generated types | Identity/order changes can break selection and contributor lookup | `intentional/no_action` | `web.networkflow` | Treat as an indirect contract; do not create a Graph Projection frontend shell |
| Test fixture executor temporarily mutates the limit registry | Sequential fixture code restores state; parallel engine tests resume after sequential tests | Future parallel fixture execution could race | `defer` | Graph Projection test infrastructure | Preserve current sequencing; revisit only if fixture parallelism is introduced |
| No grid-vendor, saved-view, workbook projection, source-table, or sibling-module imports exist in target production root | Boundary guard and exact import scan | Speculative moves would invent ownership | `intentional/no_action` | Existing named owners | Record negative evidence and take no action |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01, WF-02 | Fix target, authority, baseline commit, and execution boundary | Tracker and authorized workstream inputs | `make lint-markdown` after every slice | DONE: implementation rebaselined at clean `e9ad8671` |
| WF-01 | Target inventory | chain | WF-00 | WF-03, WF-04 | Account for all current target files and live package surfaces | Entire target, tracker | File count and exact source reads | DONE: 38 baseline files and nine GP-S01 additions inventoried |
| WF-02 | Contract-owner mapping | parallel | WF-00 | WF-03, WF-04 | Map Graph Projection, Network Flow, Reporting, Recovery, generated, and negative surfaces | Owner docs, consumers, contracts, tracker | Owner/source comparison | DONE: no owner contradiction found |
| WF-03 | Characterization test gap analysis | chain | WF-01, WF-02 | WF-05, WF-07 | Map existing tests and focused-row omissions | Target tests and test-family manifest | Make owner/task discovery | DONE as planning; GP-S00 implementation remains TODO |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05 | Classify imports, SQL, adapters, frontend, generated, and recovery seams | Target and production consumers | Boundary tests and static target discovery | DONE: findings classified in section 5 |
| WF-05 | Facade and ownership redesign plan | chain | WF-03, WF-04 | WF-06 | Retain the module, plan concern-level decomposition, and make RB-001/RB-002 owner-adoption decisions explicit | Service/store packages, owner documents, Recovery contracts | Freeze map, adopted-contract completeness, and document-placement review | DONE: planning and GP-S04 owner adoption complete; implementation continues |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order evidence closure, owner adoption, structural refactors, dormancy evidence, Recovery contracts, rebuild implementation, and final verification | Tracker and authorized owner/implementation packages | Per-slice commands and binary criteria in sections 7 and 8 | DONE: required execution order recorded |
| WF-07 | Harness/test/accounting update plan | chain | WF-03, WF-06 | WF-08 | Close focused selector gaps through authored inputs and Make generation | Test-family owner input and generated topology outputs | Owner unit slice and generation/drift gates | DONE: GP-S00 terminal evidence recorded |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 and every enacted authorized slice | None | Prove behavior and boundaries, report exact retained evidence, and close only implemented work | Changed owner, implementation, test, generated, and tracker inputs | Narrow owner slices, drift/boundary gates, `make agent-finalize`, then `make check` | TODO: follows GP-S07 |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GP-S00 | WF-03 | **DONE.** Added the eleven exact selectors listed below to the existing authored engine row; preserved fixture wrappers and the dedicated candidate target; regenerated topology only through Make | `tools/test_families/module.graphprojection.json`; `tools/execution_topology_render_index.json` | Evidence routing only; no runtime or test behavior | Every listed direct test, unit fixture wrapper, fixture verifier; no wrapped GP-FIX duplication | Generation, drift, policy, shape, focused Graph unit, boundary, and 766-unit `make check` all pass; exact run roots are in section 10 | Revert the authored family input and rerun generation; never hand-edit generated files | DONE: all eleven symbols are ASCII-sorted, unique, resolve exactly once without overlap, emit terminal evidence, and required commands pass |
| GP-S04 | GP-S00 | **DONE.** Adopted restore and dormancy requirements in Graph Projection NLSpec 1.2.0; amended Core 00/01/04; clarified domain and Appendix I; completed the REQ-00-062 re-audit | Graph Projection NLSpec, Core 00/Core 01/Core 04, `docs/domain.md`, Appendix I, tracker | Workbook contracts, routes, configuration, authorization, and provider registry remain unchanged | GP-RA-01..18 are adopted; owner boundaries and current-profile dormancy are explicit | `git diff --check`; targeted contradiction scan; `make lint-markdown` | Revert the isolated owner amendment and its adoption metadata; downstream implementation MUST NOT remain without adopted authority | DONE: exact owner text is adopted, no contradiction exists, and every affected projection is re-audited |
| GP-S01 | GP-S04 | **DONE.** Kept constructor/state in `store.go` and moved lifecycle/admission, publication, idempotency, retention, queries, pagination, traversal, and invalidation into same-package files; cursor owns cursor state and codec | `internal/modules/graphprojection/postgresstore` | Locks, publication atomicity, retention rank, cursor, errors, SQL table scope | All Graph owner and service-backed rows pass | `make format`; Graph unit and service-backed slices; backend boundary | Revert the isolated file-move slice; schema and data need no rollback | DONE: methods, APIs, SQL, locks, transactions, hooks, errors, and tests are behavior-identical |
| GP-S02 | GP-S01 | **DONE.** Kept stable DTOs, `Service`, `ServiceOptions`, and constructor in `service.go`; moved retained lifecycle, queries, invalidation, error adaptation, and ephemeral orchestration into same-package files | Root Graph Projection service/facade files | Network Flow active path, repository sentinel mapping, time/nonce behavior, envelopes | Graph owner, service-backed, and 74-unit Network Flow slices pass | `make format`; Graph unit/service-backed; Network Flow owner; backend boundary | Revert the isolated file-move slice; no persistent rollback | DONE: public API, byte/shape output, errors, cancellation, time, and nonce behavior are unchanged |
| GP-S03 | GP-S02 | **DONE — no code change.** Verified every exported helper consumer and preserved the owner-child adapter seam | `retained_engine.go`, `limits.go`, `postgresstore`, owner tests | Narrowing requires a cycle, aliases, duplicated deterministic algorithms, or a larger engine/type extraction | Existing Graph evidence from GP-S02 plus explicit boundary guard | Consumer `rg` audit; `make backend-module-boundary-check` | No rollback needed because no code changed | DONE: only Graph child adapters/tests consume helpers; preservation is the lowest-coupling design |
| GP-S05 | GP-S03 | **DONE.** Added owner-routed static composition guards over production Go/configuration; strengthened decomposed-store SQL scanning; proved Network Flow ephemeral-only and Reporting transaction-reader-only seams | Graph boundary test, Graph test family, generated topology | Guard precision, future narrow Recovery adapter, approved tests, Network Flow and Reporting seams | Constructor/options/hooks, cursor-key/flag, retained method, route/WS, full-Store, SQL, ephemeral, and caller-transaction checks | Format/generate; Graph, Network Flow, Reporting unit/service; generation drift/policy; backend boundary | Revert guard/routing/generated changes together; runtime remains dormant | DONE: all negative evidence and consumer slices pass; Recovery cannot import the full Store |
| GP-S06 | GP-S05 | **DONE.** Added closed Graph registry, implementation-binding, result, and Go request contracts; generated current and exact legacy bindings; added two v3-manifest artifacts; versioned protected journal payloads to v3 with a strict v2 decoder; propagated admitted operation/generation identities | Graph restore contract/bridge, Recovery vNext codec/application/evidence, Recovery schemas/fixtures/index, contract generator, generated Recovery projections, owner test manifests/tests | Unknown or missing bindings fail closed; only the prior codec digest with no supplied Graph artifacts selects the packaged legacy empty-registry binding; workbook contracts remain unchanged | Exact empty registry; canonical generated digests; one artifact of each kind; missing/forged binding; strict result states; safe errors; v2/v3 journal decode; generation return/propagation | Format/generate; generation drift/policy/shape; backend boundary; full Graph and Recovery unit slices | Revert authored contracts, generator/runtime/test changes together and regenerate; no database state changed | DONE: Graph 7/7 and Recovery 38/38 pass with clean generation, shape, and boundary evidence |
| GP-S07 | GP-S06 | **DONE.** Implemented Graph preflight and exact five-table atomic replacement through a narrow borrowed-Postgres writer; assembled the distinct catalog-resolved participant; preserved workbook rebuild; added protected completion, exact-identity replay, closed error mapping, and lease checks | Graph restore service/writer/tests; Recovery restore contract/application/assembly/evidence; operator CLI; owner docs and test manifests | Destructive clearing is rollback-protected; commit/cancellation/postcommit uncertainty returns incomplete; source values and dependencies remain outside safe results; ordinary retained Store remains dormant | GP-RA-01..18, existing Graph storage, Recovery readiness/journal/lease/process, Network Flow, Reporting, and operator parser evidence | Final S07 evidence: Graph 8/8 and 6/6; Recovery 38/38 and 26/26; Network Flow 74/74; Reporting 14/14 and 7/7; operator 1/1; generation/drift/policy/shape/boundary clean; exact roots in section 10 | Any indeterminate target must be discarded and reinitialized; production rollout remains outside this effort | DONE: all 18 scenarios and existing freeze-map contracts pass; RB-001 is closed |
| GP-S08 | GP-S00..GP-S07 for every enacted slice | **DONE.** Ran the exact narrow-to-broad matrix, finalized without incompatible retained evidence, corrected one staticcheck-only failure, reran the broad gate, and completed this handoff | Tracker and every file touched by authorized slices | Initial broad gate passed 766/767 and failed only lowercase error diagnostics; correction changed no product behavior; final broad gate passed 767/767 | Every section 4 contract and GP-RA row; Graph, Recovery, Network Flow, Reporting, operator, generation, boundary, Markdown, and full check evidence | Exact commands and roots are recorded in section 13 | Revert a future failing isolated change rather than weakening tests; no rollback is pending | DONE: all required commands pass; conditional browser and migration checks have exact justified skips |

The user authorized every slice in this plan on 2026-08-15. No slice activates a public Graph route or ordinary retained lifecycle; any future activation is outside this plan and remains gated by section 4. Execution MUST follow `GP-S00 → GP-S04 → GP-S01 → GP-S02 → GP-S03 → GP-S05 → GP-S06 → GP-S07 → GP-S08`, with tracker and Markdown evidence completed between slices.

GP-S00 MUST add exactly these ASCII-sorted symbols to `module.graphprojection.engine.canonical_behavior`:

```text
TestAdmissionRejectsNestedUnknownMemberWithClosedDetails
TestAdmissionRejectsVirtualOversizedInputBeforeRead
TestDuplicateIdentifiersUseOwnerCodeAndStableOrder
TestGraphViewIDMismatchDoesNotExposeDerivedIdentity
TestIdempotencyKeyUsesUnicodeScalarContract
TestIdempotencyPresenceDistinguishesOmittedNullAndEmpty
TestMalformedJSONIsNotMisclassifiedAsDuplicateMember
TestProjectedSystemMetadataFieldPathRejected
TestRepresentableScalarViolationIsAdmittedThenFails
TestResourceRegistryCollectionOverflow
TestServiceWrapsRepositorySentinelsInContractErrors
```

`TestGraphProjectionUnitBehaviorFixtures_Unit` MUST remain the single unit GP-FIX selector. Retained GP-FIX scenarios MUST remain under their service-backed wrapper. `TestGraphProjectionFixtureCandidate` MUST remain exclusive to the candidate-generation target. GP-S00 MUST NOT add individual wrapped `TestGPFIX*` functions or create a second remediation row.

## 8. Validation Plan

These are discovered public Make targets for the authorized remediation.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| Graph unit | `make test-slice OWNER=module.graphprojection ROWS=module.graphprojection.engine.canonical_behavior,module.graphprojection.fixtures.fixture_verifier` | Pure engine, facade, boundaries, and fixture verifier | yes | Focused row is incomplete until GP-S00 resolves RB-003 |
| Graph integration | `make service-backed-test-slice OWNER=module.graphprojection ROWS=module.graphprojection.storage.lifecycle,module.graphprojection.storage.migration_reset,module.graphprojection.storage.transaction_binding` | Retained PostgreSQL lifecycle, head schema, and Reporting binding | yes | Required before and after storage/facade structural slices and GP-S07 |
| Recovery unit | `make test-slice OWNER=module.recovery` | Catalog, contracts, request/result mapping, target generation, readiness, typed failures, and terminal replay | no | Required for GP-S06 and GP-S07 after new tests are routed through authored family inputs |
| Recovery integration | `make service-backed-test-slice OWNER=module.recovery` | Target admission, atomic publication, journal/audit, indeterminate outcomes, and assembly | no | Required for GP-S07; use `make explain-test-owner OWNER=module.recovery` before execution |
| Network Flow consumer | `make test-slice OWNER=module.networkflow` | Ephemeral adapter, generated envelope, authorization, and stable graph identity | no | Required if Graph facade files, error mapping, or semantic identity are touched |
| Reporting consumer | `make test-slice OWNER=module.reporting` and `make service-backed-test-slice OWNER=module.reporting` | Same-transaction binding and release tuple/error behavior | no | Required if binding types, reader, store decomposition, or Recovery publication semantics can affect retained binding facts |
| e2e/browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_default_graph_mode_fc65f0ecfd,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e` | Network Flow graph default, edge contributor drawer, and stable vertex identity | no | Required only if the Network Flow envelope or semantic graph identity changes |
| generation and drift | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, and `make json-shape-check` | Authored harness/contract inputs and all generated outputs | no | Required after GP-S00 or GP-S06; generated files MUST NOT be hand-edited |
| import-boundary/static | `make backend-module-boundary-check` | Backend module ownership policy | yes | Also preserve root `boundary_guard_test.go` through the unit row |
| full check | `make check` | Developer verification gate | no | Run after narrow evidence for an authorized implementation, not for this tracker-only task |
| documentation | `make lint-markdown` | Authored Markdown, including this tracker | yes | Required after every tracker checkpoint |

GP-S00 uses this order: `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make explain-test-owner OWNER=module.graphprojection`, the focused Graph unit slice, `make backend-module-boundary-check`, and `make check`. After a compatible successful retained run exists, it runs `make agent-finalize RESULTS_DIR=<compatible-successful-run-root>`. `make agent-finalize` MUST NOT be run with an invented, failed, or incompatible result root.

`make migration-drift` is required only if a slice unexpectedly changes authored SQL or migrations. The planned remediation requires no migration.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| GT-001 | Establish authority, scope, target, and clean baseline | WF-00 | DONE | None | Section 1; commit `e9ad8671` | Exact target/output/execution boundary recorded |
| GT-002 | Inventory all current target files | WF-01 | DONE | GT-001 | Section 2 | One live-repository row per baseline file and every decomposition addition |
| GT-003 | Map adopted owners and consumer contracts | WF-02 | DONE | GT-001 | Sections 3 and 4 | Every discovered contract has owner and evidence posture |
| GT-004 | Diagnose target boundary without assuming directory validity | WF-04 | DONE | GT-002, GT-003 | Sections 3 and 5 | Legitimate boundary and internal mixed concerns distinguished |
| GT-005 | Identify characterization and harness-accounting gaps | WF-03 | DONE | GT-002, GT-003 | RB-003 and GP-S00 | Existing rows and 11 omitted tests mapped |
| GT-006 | Make RB-001 through RB-003 closure decisions unambiguous | WF-05 | DONE | GT-003, GT-005 | Section 4 contracts/defaults/acceptance and section 11 | No planning choice remains; current and proposed states are distinct |
| GT-007 | Close focused test-family routing | WF-07 | DONE | GT-005 | Authored manifest, generated render index, and section 10 terminal evidence | Authored routing, generated topology, exact resolution, and focused owner verification pass |
| GT-008 | Decompose retained PostgreSQL store by concern | WF-05 | DONE | GT-007, GT-011 | GP-S01 files and terminal Graph evidence | Same APIs, SQL, locks, transactions, and errors with passing storage evidence |
| GT-009 | Decompose service facade by concern | WF-05 | DONE | GT-008 | GP-S02 files and Graph/Network Flow evidence | Same API, output, errors, and consumer behavior with passing evidence |
| GT-010 | Review adapter helper surface | WF-05 | DONE — NO CODE CHANGE | GT-009 | GP-S03 consumer audit and boundary evidence | Existing surface preserved because no cycle-safe reduction exists |
| GT-011 | Adopt restore and retained-dormancy owner requirements | WF-05 | DONE | GT-006, GT-007 | Graph Projection NLSpec 1.2.0, Core 00/01/04, domain, Appendix I | Owner text is adopted and the REQ-00-062 re-audit is complete |
| GT-012 | Prove dormant ordinary production composition | WF-05 | DONE | GT-010, GT-011 | Owner-routed guard, generated topology, Graph/consumer evidence | Negative construction/key/worker/route/flag evidence and consumer seams pass |
| GT-013 | Add Graph-specific Recovery contracts and projections | WF-05 | DONE | GT-012 | GP-S06 authored schemas/fixtures, generated binding projection, Graph/Recovery ports and evidence codecs | Exact closed contracts/defaults/digests, current/legacy selection, target identity, and journal compatibility pass without generic workbook drift |
| GT-014 | Implement Graph clear/rebuild and Recovery assembly | WF-05 | DONE | GT-009, GT-011..GT-013 | GP-S07 implementation and terminal GP-RA/owner evidence | All adopted owner requirements and all 18 scenarios pass |
| GT-015 | Run final proportional verification and close handoff | WF-08 | DONE | Every enacted GT-007..GT-014 item | GP-S08 roots and section 13 final handoff | Required narrow and 767/767 broad gates pass; skips are explicit |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Scope and authority mapping complete; tracker created | Inspected framework, AGENTS, adopted Graph Projection/Network Flow/Reporting owners, Core 00-04 sections, domain; touched only this tracker | `git status`, `git rev-parse`, targeted `rg`/`sed`, `wc` | Target exists with 38 files at clean `adfcfca6`; no owner contradiction; Core 05 not applicable | RB-001..RB-003 affect later implementation only | Obtain separate authorization before GP-S00 |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Current/proposed/future authority classes and decision-complete closure package added | Inspected this tracker, `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, Graph Projection NLSpec, Core 00/01/04 recovery clauses, and AGENTS; touched only this tracker | `git status`, `git rev-parse`, targeted `rg`/`sed`; `make lint-markdown` | RB decisions are explicit without claiming adoption; Markdown passed at `.cartulary/test-results/20260815T140032Z-p4110301` | RB-001 BLOCKED; RB-002 and RB-003 TODO in repository | Authorize GP-S00 independently or GP-S04 before behavior work |
| 2026-08-15 11:10 EDT | Codex remediation session, GP-S00 | Implementation authority and required sequence recorded at clean `e9ad8671`; evidence-routing closure complete | Touched tracker, authored Graph test-family manifest, and Make-generated execution-topology render index | `make generate`; generation drift/policy/shape gates; owner explanation; focused Graph unit slice; backend boundary; `make check`; `make lint-markdown` | RB-003 resolved; no runtime behavior changed; all required GP-S00 gates pass, including Markdown at `20260815T151112Z-p126627` | None | Begin GP-S04 |
| 2026-08-15 11:34 EDT | Codex remediation session, GP-S04 | Graph recovery and retained-dormancy authority adopted; REQ-00-062 re-audit complete | Touched Graph Projection NLSpec 1.2.0, Core 00/01/04, domain, Appendix I, and tracker | Exact owner reads; targeted contradiction `rg`; `git diff --check`; `make lint-markdown` (`20260815T151702Z-p130106`) | Workbook rebuilder remains separate; Graph request/result/registry/binding, limits, historical policy, safe failures, completion proof, and GP-RA-01..18 are adopted; Markdown passes | Implementation remains for GP-S05..07 | Begin GP-S01 |
| 2026-08-15 12:04 EDT | Codex remediation session, GP-S06 | Graph-specific Recovery machine contracts, generated bindings, manifest artifacts, identity propagation, and protected terminal-evidence versioning complete | Added Recovery schemas/fixtures/index and generated projections; touched contractgen, generation inputs, Graph restore contracts/tests, Recovery codec/application/evidence/restore contracts/tests, and owner test manifests | Final: `make format` (`20260815T160013Z-p401024`); `make generate` (`20260815T160016Z-p404484`); drift (`20260815T160025Z-p407353`); policy (`20260815T160033Z-p410244`); shape (`20260815T160034Z-p410650`); boundary (`20260815T160037Z-p411107`); Graph (`20260815T160043Z-p411498`); Recovery (`20260815T160106Z-p413141`); Markdown (`20260815T160511Z-p460406`). Corrected intermediate failures: empty-registry generator rejection (`20260815T154437Z-p274443`), stale generated legacy constants (`20260815T155051Z-p285534`), an over-broad Recovery import (`20260815T155136Z-p290457`), and obsolete contract-count assertions (`20260815T155333Z-p298508`) | All final gates pass: Graph 7/7, Recovery 38/38, generation 4/4, policy 3/3, shape 3/3, boundary 3/3, and Markdown. Intermediate failures were related and corrected structurally without weakening tests | Graph publication/Recovery invocation remains GP-S07 | Begin GP-S07 after Markdown checkpoint |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | Graph-owned clear/rebuild, Recovery coordination, protected terminal replay, and RB-001 closure complete | Added `restore_service.go`, `restore_service_test.go`, and `postgresrestore/writer.go`/`writer_test.go`; implemented Recovery bridge/assembly/application/evidence/operator coordination; added exact retry CLI identity; clarified Graph/Core 01 and operator guidance; extended Graph/Recovery manifests and generated topology | Final: format `20260815T170159Z-p1077063`; Graph unit/service `20260815T170446Z-p1164813`/`20260815T170511Z-p1166522`; Recovery unit/service `20260815T170233Z-p1082713`/`20260815T170348Z-p1128574`; Network Flow `20260815T170520Z-p1167668`; Reporting unit/service `20260815T170748Z-p1216182`/`20260815T170817Z-p1220908`; operator parser `20260815T170226Z-p1082180`; generate/drift/policy/shape/boundary `20260815T170823Z-p1222038`/`20260815T170832Z-p1224949`/`20260815T170840Z-p1227840`/`20260815T170841Z-p1228240`/`20260815T170844Z-p1228703`; Markdown `20260815T171344Z-p1231131` | Final listed gates pass. Corrected related intermediate failures exposed restore-normalized digest semantics (`20260815T164538Z-p764702`), a shadowed Graph result causing incomplete journal evidence (`20260815T164807Z-p772837`), and historical workbook-only caller identity compatibility (`20260815T165817Z-p994069`) without weakening tests | Production registry remains intentionally empty; older binaries fail closed on the new codec digest; indeterminate targets require reinitialization | Begin GP-S08 |
| 2026-08-15 13:38 EDT | Codex remediation session, GP-S08 | Final validation and handoff complete | Revalidated all authored/generated/runtime/test/documentation changes and updated only staticcheck-invalid diagnostic capitalization plus this tracker during finalization | Exact narrow, finalizer, lint, and broad roots are in section 13; final Markdown checkpoint `20260815T173839Z-p1755608` | Required narrow gates pass; final broad `make check` passes 767/767; final Markdown passes | No implementation blocker; rollout safeguards remain | Hand off completed workspace |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Legitimate subsystem with mixed internal concerns diagnosed | Inspected all 38 target files, production importers, migration, and recovery assembly; touched only this tracker | Exact source reads, import/caller/export/SQL searches | Keep Graph Projection boundary; split store and service files later; retained production assembly absent | RB-001, RB-002 | Start with GP-S00, not structural changes |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Boundary verdict preserved; current composition and dormant-store decision made explicit | Re-inspected Graph service/store/binding/recovery files, Network Flow adapter, Reporting store, Recovery catalog and operator composition; touched only this tracker | Targeted `rg`/`sed` production-constructor and adapter searches | Engine/ephemeral and Reporting seams are active; full Store, cursor configuration, workers, routes, and Graph restore port remain absent | Owner adoption and negative composition evidence remain | GP-S01/02 are behavior-preserving; GP-S05 MUST NOT activate the Store |
| 2026-08-15 11:52 EDT | Codex remediation session, GP-S01 | Retained PostgreSQL adapter decomposed without package or behavior change | Replaced the 1,317-line `postgresstore/store.go` with 43-line construction state; added nine concern files; moved cursor state into `cursor.go`; updated tracker inventory | `make format` (`20260815T151911Z-p132873`); `make test-slice OWNER=module.graphprojection` (`20260815T151929Z-p136495`); `make service-backed-test-slice OWNER=module.graphprojection` (`20260815T151956Z-p138270`); `make backend-module-boundary-check` (`20260815T152004Z-p139423`); `make lint-markdown` (`20260815T152117Z-p140208`) | Format 2/2, Graph 7/7, service-backed 5/5, boundary 3/3, and Markdown all pass; no SQL, migration, schema, API, or generated contract changed | None | Begin GP-S02 |
| 2026-08-15 12:27 EDT | Codex remediation session, GP-S02 | Root facade decomposed without package, API, or behavior change | Replaced 562-line `service.go` with 85-line DTO/construction file; added five concern files; updated tracker inventory | Initial `make format` failed at `20260815T152303Z-p143022` because extraction omitted one closing brace; source repaired; format passed at `20260815T152332Z-p147055`; Graph unit `20260815T152341Z-p150590`; Graph service `20260815T152408Z-p152461`; Network Flow `20260815T152417Z-p153592`; boundary `20260815T152650Z-p202924`; Markdown `20260815T152806Z-p203733` | Related formatting failure was fully corrected; Graph 7/7, service-backed 5/5, Network Flow 74/74, boundary 3/3, and Markdown pass; no envelope or identity changed | None | Record GP-S03 helper decision |
| 2026-08-15 12:30 EDT | Codex remediation session, GP-S03 | Exported owner-child helpers preserved by explicit no-code decision | Inspected `retained_engine.go`, `limits.go`, all Graph child packages, owner tests, and boundary guards; touched only tracker | Exact exported-symbol and consumer `rg`; `make backend-module-boundary-check` (`20260815T152907Z-p205833`); `make lint-markdown` (`20260815T152941Z-p206384`) | Only `postgresstore` and owner tests consume the helpers; boundary 3/3 and Markdown pass; narrowing would increase coupling or duplicate algorithms | None | Begin GP-S05 |
| 2026-08-15 12:36 EDT | Codex remediation session, GP-S05 | Ordinary retained Graph production composition is guarded as dormant | Extended `boundary_guard_test.go`; routed one new direct test; regenerated topology; updated tracker | `make format` (`20260815T153136Z-p208731`); `make generate` (`20260815T153144Z-p212236`); Graph `20260815T153158Z-p215187`; Network Flow `20260815T153226Z-p217032`; Reporting unit/service `20260815T153451Z-p260026`/`20260815T153523Z-p264674`; generation drift `20260815T153531Z-p265802`; policy `20260815T153539Z-p268661`; boundary `20260815T153540Z-p269133`; Markdown `20260815T153619Z-p269705` | All gates pass: Graph 7/7, Network Flow 74/74, Reporting 14/14 and 7/7, drift 4/4, policy 3/3, boundary 3/3, Markdown pass | Future Graph restore adapter must remain narrow and avoid `postgresstore` | Begin GP-S06 |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | Separate restore application service and narrow publication adapter preserve the module boundary | Added the four inventoried restore service/writer files; extended defensive table-ID and SQL/import guards; touched Recovery only through `restorecontract` plus application assembly | Graph unit/service and backend boundary roots listed in the scope row | Graph engine, retained Store, ephemeral facade, Reporting binding, and public transport remain separate; exact five-table writer passes committed reread and rollback tests | None | Final proportional verification in GP-S08 |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Frontend is an indirect Network Flow consumer | Inspected Network Flow generated types, graph controller, workspace, unit tests, and browser rows; touched only this tracker | Targeted `rg`/`sed` and test-family reads | No Graph Projection frontend shell or grid adapter; stable vertex/edge identity remains a consumer contract | None for planned same-package splits | Run graph browser rows only if envelope or identity changes |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Frontend boundary unchanged and explicitly excluded from restore/dormancy work | Reused exact Network Flow controller/workspace/generated-contract/browser evidence; touched only this tracker | Tracker and source cross-checks only | No Graph-owned frontend, grid adapter, public route, or browser recovery control is proposed | None for documentation; consumer risk remains conditional | Run the three browser rows only if a later slice changes envelope or graph identity |
| 2026-08-15 12:27 EDT | Codex remediation session, GP-S02 | Active Network Flow ephemeral consumer remains byte/shape compatible | Touched only root Graph facade organization; no Network Flow source, contract, envelope, graph identity, or frontend file changed | `make test-slice OWNER=module.networkflow` (`20260815T152417Z-p153592`) | 74/74 units pass; three conditional browser rows skipped because envelope and semantic graph identity are unchanged | None | Preserve the explicit browser skip through final handoff unless a later slice changes identity/envelope |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | Restore-normalized completion digest added without changing the Graph resource envelope or deterministic vertex/edge identities | No frontend, Network Flow contract, generated TypeScript, controller, or workspace file changed | Network Flow 74/74 at `20260815T170520Z-p1167668` | Active ephemeral consumer is unchanged; three conditional Graph-identity browser rows remain unnecessary | None | Record the explicit browser skip in GP-S08 |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Contract freeze map complete; no contract edited | Inspected Network Flow route/schema/generated outputs, Reporting OpenAPI/ref behavior, recovery fixture/generated projection, migration, and fixture schema; touched only this tracker | Targeted contract searches/reads; `make help-all` | Graph Projection has no direct generated protocol or public route; indirect Network Flow and Reporting surfaces are mapped | RB-001 for recovery behavior | Edit authored inputs only in later authorized work and regenerate via Make |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Four proposed Graph-internal Recovery contracts and placement boundaries are decision-complete; no contract changed | Inspected Core generic restore contract, Graph recovery contribution, Recovery catalog fixture, target marker, operator/projection assembly, and adoption rule; touched only this tracker | Targeted `rg`/`sed`/`jq` contract and assembly reads | Generic workbook rebuild remains unchanged; Graph-specific catalog-resolved contracts require GP-S04/GP-S06 | RB-001; proposed text is not adopted authority | Adopt in named owners, then edit authored machine inputs and regenerate only through Make |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | GP-S06 contracts are executable and the current binding matches the real frozen Recovery catalog | Corrected the authored current binding/catalog digest and dependent fixtures; regenerated Recovery projections; implemented runtime registry/binding selection and v3 completion persistence/replay | Generate/drift/policy/shape roots listed in scope row | Current and exact legacy empty-registry bindings pass; unknown/missing bindings fail closed; workbook contract remains distinct | Older binaries cannot read the new codec digest and must be retained with prior verified backups during rollout | Run all generation gates again in GP-S08 |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Five owner rows mapped; selector gap identified; no product tests run | Inspected all target tests, fixtures, verification owner, test-family and task surface; touched only this tracker | `jq`/`comm`; `make task-guide ROLE=module-author OWNER=module.graphprojection`; `make explain-test-owner OWNER=module.graphprojection`; four `make explain-target` calls; `make lint-markdown` | Two unit and three service-backed rows found; 11 non-wrapper tests are absent from focused routing; Markdown passed at `.cartulary/test-results/20260815T130934Z-p4096439` | RB-003 for later focused evidence | Later authorize GP-S00; do not infer product validation from the Markdown pass |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Exact eleven-symbol selector plan, fixture exclusions, and future validation order recorded; no harness edit or product test | Re-inspected remediation/facade symbols, Graph test-family manifest, Recovery family manifest, and candidate runner; touched only this tracker | `rg`, `jq`; `make task-guide ROLE=module-author OWNER=module.recovery`; `make explain-test-owner OWNER=module.recovery`; `make lint-markdown` | RB-003 design is DONE, repository routing remains TODO; Markdown passed at `.cartulary/test-results/20260815T140032Z-p4110301` | GP-S00 needs separate authorization and terminal evidence | Add only the eleven exact selectors, generate through Make, then run narrow-to-broad gates |
| 2026-08-15 11:10 EDT | Codex remediation session, GP-S00 | Eleven omitted direct contract tests are owner-routed without fixture duplication | Touched `tools/test_families/module.graphprojection.json`, generated `tools/execution_topology_render_index.json`, and this tracker | `make generate` (`20260815T150130Z-p4134257`); `make generate-drift` (`20260815T150148Z-p4137307`); `make generated-artifact-policy-check` (`20260815T150159Z-p4140232`); `make json-shape-check` (`20260815T150203Z-p4140661`); `make explain-test-owner OWNER=module.graphprojection`; focused `make test-slice` (`20260815T150218Z-p4141393`); `make backend-module-boundary-check` (`20260815T150225Z-p4141922`); `make check` (`20260815T150234Z-p4142415`); `make lint-markdown` (`20260815T151112Z-p126627`) | All commands pass; focused row resolves all requested symbols with terminal evidence; full gate passes 766/766 units; Markdown checkpoint passes | None | Begin GP-S04 |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | GP-RA-01..18 are executable through owner-routed unit, PostgreSQL, assembly, lease, journal, process, replay, and boundary evidence | Added seven Graph service tests, three Graph publication tests, Recovery assembly evidence, exact-retry process coverage, and expanded SQL/import guards; updated Graph and Recovery family inputs and generated topology | Graph 8/8 and 6/6; Recovery 38/38 and 26/26; operator parser 1/1; boundary 3/3; exact roots in scope row | Clear-only, valid rebuild, three skips, preflight failures, every aggregate limit, rollback, uncertainty, replay, retry determinism, stale-history exclusion, safe failures, and boundaries pass | None | Repeat required narrow-to-broad matrix in GP-S08 |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Caller-owned authorization boundary frozen | Inspected Network Flow graph handlers/route catalog, Reporting admission, Graph Projection errors and boundary guard; touched only this tracker | Targeted route/auth/error searches and exact reads | Network Flow and Reporting own authorization and public error mapping; Graph Projection adds no auth semantics and leaks no source values | None for planned same-package splits | Preserve denial precedence and no-public-route posture |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Safe restore results, typed source capabilities, lease outcome, and no-public-transport constraints specified for later adoption | Inspected Core target-marker/serving-lease clauses, Graph safe errors, Network Flow authorization, Reporting binding, and Recovery journal/error mapping; touched only this tracker | Targeted owner/source reads; no security suite run | Proposed result fields exclude source/config/SQL/secret material; caller-owned authorization remains frozen | RB-001 adoption and implementation evidence | GP-S04 MUST adopt safe boundaries before GP-S06/07; no route or auth moves into Graph |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | Closed Graph restore failures and target safety are enforced through commit, postcommit proof, terminal evidence, replay, and lease decisions | Touched Graph restore errors/service/writer, Recovery typed stage mapping/evidence/admission, operator exact retry identity, and safe boundary tests | Graph/Recovery/process/boundary evidence roots listed in scope row | No error-string classification, raw dependency error, SQL, source value, key, path, or capability reaches result/audit evidence; public routes and retained activation remain absent | Commit uncertainty and lost postcommit proof deliberately require target reinitialization | Preserve closed classifications in final broad gate |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-15 09:05 EDT | Codex planning/documentation session | Planning tracker complete; refactor unstarted | Inspected repository evidence listed in sections 1-5; touched only this tracker | Discovery commands only; no product validation | Safe sequence is evidence closure, store split, service split, deferred adapter review; behavior additions stay outside refactor | RB-001..RB-003 | Seek later authorization for GP-S00; do not activate retained or recovery behavior implicitly |
| 2026-08-15 09:59 EDT | Codex NLSpec-style tracker revision | Planning ambiguity closed; owner adoption, harness correction, and refactor remain unstarted | Inspected recommendations, style guidance, exact owners, current contracts/composition, test routing, and validation surface; touched only this tracker | Discovery commands plus `make lint-markdown`; no product validation, generation, drift, or finalizer | GP-S00..GP-S08 now have binary gates; Markdown passed at `.cartulary/test-results/20260815T140032Z-p4110301` | RB-001 BLOCKED; RB-002/RB-003 TODO | Choose an independently authorized next slice; never cite this tracker as adopted product authority |
| 2026-08-15 13:10 EDT | Codex remediation session, GP-S07 | All implementation workstreams are complete; only final validation/handoff remains | Reconciled source, owners, contracts, generated outputs, tests, and this tracker | All S07 final roots are in the scope row; Markdown passed at `20260815T171344Z-p1231131` | RB-001..RB-003 and GT-007..GT-014 are DONE; no migration or public Graph activation was introduced | Operational rollout must create and verify a fresh backup; keep the prior backup/binary pair until then | Begin GP-S08 exact validation matrix |
| 2026-08-15 13:38 EDT | Codex remediation session, GP-S08 | Remediation and handoff complete | Final source/status/diff reconciliation and section 13 exact inventory/evidence | Final `make check` 767/767 at `20260815T172858Z-p1576876`; final Markdown checkpoint follows | No repository blocker; all GT and RB items are DONE | Fresh-backup rollout verification, legacy binary/backup retention, and indeterminate-target reinitialization remain operational responsibilities | Hand off completed workspace |

## 11. Open Questions and Blockers

No owner contradiction or open remediation blocker remains. Planning and GP-S00 through GP-S08 are closed. `DONE` in the decision column means that another implementer need not choose a design. Repository status records whether the corresponding owner or implementation work exists.

| ID | Closed planning decision | Remaining repository blocker | Needed authority or evidence | Decision state | Repository status |
| --- | --- | --- | --- | --- | --- |
| RB-001 | Use a Graph-owned catalog-resolved clear-and-reconstruct action; always clear the exact five tables; rebuild only exact registry candidates valid at the backup consistency point; empty registry is successful clear-only | None for the current empty-registry profile; future source registrations remain separately frozen in the generated binding | GP-S07 Graph/Recovery implementation, GP-RA-01..18, consumer, boundary, and durable replay evidence | DONE | DONE |
| RB-002 | Keep the retained Store and its conformance evidence but declare it intentionally dormant in ordinary current-profile production; do not activate it or add a cursor key, route, worker, writer, or hidden flag | None for the current profile; future activation remains separately adopted | GP-S04 owner text and GP-S05 application/boundary/consumer evidence | DONE | DONE |
| RB-003 | Add the eleven exact ASCII-sorted source symbols in section 7 to the existing engine row; preserve all wrapper and candidate boundaries | None | GP-S00 authored manifest edit, Make-owned generation, exact resolution/overlap checks, focused owner slice, drift/policy/shape/boundary/full evidence | DONE | DONE |

## 12. Binary Completion Criteria

Planning, owner adoption, implementation, final validation, and handoff are complete.

- **PASS:** Every baseline file, all nine GP-S01 store files, and all five GP-S02 facade files in `internal/modules/graphprojection` are inventoried; none is silently excluded.
- **PASS:** Every discovered public or cross-owner contract risk has an owner, evidence source, test posture, and refactor risk.
- **PASS:** WF-00 through WF-08 have explicit dependencies, goals, validation, and handoff checkpoints.
- **PASS:** GP-S00 through GP-S08 have dependencies, exact intent, likely files, risk, tests, validation, rollback, authorization, and binary completion conditions.
- **PASS:** Current facts, adopted owner requirements, and future implementation requirements are explicitly distinct.
- **PASS:** The adopted request, result, source registry, and implementation binding have closed members, required/null/default rules, ordering, digest, and safe-data boundaries.
- **PASS:** Candidate validity, three closed skip reasons, fail-before-touch conditions, generation proof, atomic publication, outcome handling, readiness, and empty-registry defaults are unambiguous.
- **PASS:** GP-RA-01 through GP-RA-18 cover the complete adopted restore acceptance posture.
- **PASS:** Intentional retained-store dormancy, negative production constraints, ten future-activation gates, and conditional cursor/resource lifecycle are mapped without activating behavior.
- **PASS:** All eleven exact RB-003 selectors and all fixture-accounting exclusions are recorded.
- **PASS:** Canonical Make commands are discovered for Graph, Recovery, Network Flow, Reporting, generation/drift, boundaries, Markdown, and the full gate.
- **PASS:** No owner contradiction was found; none is silently resolved.
- **PASS:** Framework/live-repository mismatches are recorded as planning findings rather than implementation assumptions.
- **PASS:** Prior handoff history is preserved and all seven required handoff areas contain a new current-session row with enough evidence for continuation.
- **PASS:** Markdown lint passes for the revised tracker; its actual result is recorded in the current session log.
- **PASS:** GP-S00 touched only its authorized tracker, harness input, and Make-generated topology projection; no runtime behavior is represented as changed.
- **PASS:** GP-S08 final `make check` passes 767/767 after correcting the sole staticcheck-only diagnostic failure.

## 13. GP-S08 Final Handoff

### Planning and substantive completion

All workstreams completed in the required order. The result closes RB-001
through RB-003 without activating ordinary retained Graph behavior. The
PostgreSQL store and root facade are decomposed by concern; focused Graph test
routing is complete; retained production composition is explicitly dormant;
Graph-owned Recovery contracts and generated bindings are exact; and vNext
Recovery now executes a separate Graph participant before workbook rebuild
readiness. The Graph participant preflights typed source candidates, applies
one operation-wide budget, performs exact five-table replacement through a
narrow borrowed-Postgres writer, rereads committed postconditions, persists a
protected completion tuple, and replays matching terminal success by exact
operation identity without a second rebuild.

The current production registry remains exactly empty. Therefore current
restore behavior is successful clear-only after authoritative data restore.
The implementation is ready for future source-owner registrations only through
a separately frozen registry and implementation binding.

### Exact changed files

```text
contracts/index.json
contracts/recovery/fixtures/backup-integrity-manifest.v3.json
contracts/recovery/fixtures/graph-projection-restore-implementation-binding.v1.json
contracts/recovery/fixtures/graph-projection-restore-rebuild-result.v1.json
contracts/recovery/fixtures/graph-projection-restore-source-registry.v1.json
contracts/recovery/fixtures/operator-recovery-journal-payload.v3.json
contracts/recovery/graph-projection-restore-implementation-binding.v1.schema.json
contracts/recovery/graph-projection-restore-rebuild-result.v1.schema.json
contracts/recovery/graph-projection-restore-source-registry.v1.schema.json
contracts/recovery/index.json
contracts/recovery/operator-recovery-journal-payload.v3.schema.json
docs/domain.md
docs/graph_projection_nlspec.md
docs/guides/cartulary-dev-guide.md
docs/handoffs/graphprojection-module-refactor-tracker.md
docs/spec/00_document_set_status_and_precedence.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/04_security_deployment_and_conformance.md
docs/spec/H_operating_model_supporting_guidance.md
docs/spec/I_projection_authority_boundary_and_characterization.md
internal/app/operator/internal/recoverycli/cli.go
internal/app/operator/internal/recoverycli/cli_test.go
internal/app/operator/operator_recovery.go
internal/app/recoveryassembly/evidence_repository.go
internal/app/recoveryassembly/evidence_repository_test.go
internal/app/recoveryassembly/graphprojection_restore.go
internal/app/recoveryassembly/graphprojection_restore_test.go
internal/gen/contractextensions/artifacts_gen.go
internal/gen/contractrecovery/artifacts_gen.go
internal/gen/contractrecovery/graph_projection_restore_binding_gen.go
internal/modules/graphprojection/boundary_guard_test.go
internal/modules/graphprojection/ephemeral_service.go
internal/modules/graphprojection/error_adaptation.go
internal/modules/graphprojection/invalidation_service.go
internal/modules/graphprojection/postgresrestore/writer.go
internal/modules/graphprojection/postgresrestore/writer_test.go
internal/modules/graphprojection/postgresstore/cursor.go
internal/modules/graphprojection/postgresstore/idempotency.go
internal/modules/graphprojection/postgresstore/invalidation.go
internal/modules/graphprojection/postgresstore/lifecycle.go
internal/modules/graphprojection/postgresstore/pagination.go
internal/modules/graphprojection/postgresstore/publication.go
internal/modules/graphprojection/postgresstore/queries.go
internal/modules/graphprojection/postgresstore/retention.go
internal/modules/graphprojection/postgresstore/sql_values.go
internal/modules/graphprojection/postgresstore/store.go
internal/modules/graphprojection/postgresstore/traversal.go
internal/modules/graphprojection/query_service.go
internal/modules/graphprojection/restore_contract.go
internal/modules/graphprojection/restore_contract_test.go
internal/modules/graphprojection/restore_service.go
internal/modules/graphprojection/restore_service_test.go
internal/modules/graphprojection/retained_lifecycle_service.go
internal/modules/graphprojection/service.go
internal/modules/recovery/application/evidence.go
internal/modules/recovery/application/service.go
internal/modules/recovery/application/target_admission.go
internal/modules/recovery/application/target_admission_test.go
internal/modules/recovery/application/types.go
internal/modules/recovery/operatortest/operator_process_test.go
internal/modules/recovery/restore.go
internal/modules/recovery/restorecontract/graphprojection.go
internal/modules/recovery/vnext_codec.go
internal/modules/recovery/vnext_codec_test.go
internal/modules/recovery/vnext_graph_restore_artifacts_test.go
tools/contractgen/main.go
tools/contractgen/recovery_validation.go
tools/execution_topology_render_index.json
tools/generate_drift_scratch_inputs.json
tools/test_families/module.graphprojection.json
tools/test_families/module.recovery.json
```

### Final validation evidence

| Command | Result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `20260815T171448Z-p1233134` |
| `make generate-drift` | PASS, 4/4 | `20260815T171457Z-p1236024` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `20260815T171505Z-p1238924` |
| `make json-shape-check` | PASS, 3/3 | `20260815T171506Z-p1239324` |
| `make backend-module-boundary-check` | PASS, 3/3 | `20260815T171509Z-p1239788` |
| `make test-slice OWNER=module.graphprojection` | PASS, 8/8 | `20260815T171513Z-p1240166` |
| `make service-backed-test-slice OWNER=module.graphprojection` | PASS, 6/6 | `20260815T171516Z-p1241263` |
| `make test-slice OWNER=module.recovery` | PASS, 38/38 | `20260815T171533Z-p1242420` |
| `make service-backed-test-slice OWNER=module.recovery` | PASS, 26/26 | `20260815T171621Z-p1278395` |
| `make test-slice OWNER=module.networkflow` | PASS, 74/74 | `20260815T171712Z-p1314526` |
| `make test-slice OWNER=module.reporting` | PASS, 14/14 | `20260815T171930Z-p1353113` |
| `make service-backed-test-slice OWNER=module.reporting` | PASS, 7/7 | `20260815T171933Z-p1354211` |
| `make lint-markdown` | PASS | `20260815T171936Z-p1355328` |
| `make agent-finalize` | PASS, 1/1 | `20260815T171941Z-p1356120` |
| Initial `make check` | FAIL, 766/767; staticcheck-only lowercase error diagnostics | `20260815T171958Z-p1359005` |
| `make format` after diagnostic correction | PASS, 2/2 | `20260815T172827Z-p1568314` |
| `make lint-go` format subtarget | PASS | `20260815T172830Z-p1571984` |
| `make lint-go` vet subtarget | PASS | `20260815T172832Z-p1575277` |
| `make lint-go` staticcheck subtarget | PASS | `20260815T172834Z-p1575936` |
| Final `make check` | PASS, 767/767 | `20260815T172858Z-p1576876` |
| Final post-handoff `make lint-markdown` | PASS | `20260815T173839Z-p1755608` |

The three conditional Network Flow browser rows were not run because neither
the embedded Graph envelope nor Network Flow graph identity changed. Recovery
process/browser evidence included in the Recovery owner slice passed. The
planned implementation added no authored SQL, so `make migration-drift` was
not applicable. `make agent-finalize` ran without `RESULTS_DIR`; compatible
retained successful full-run evidence was unavailable, so retained-run
maintenance was skipped as required rather than attaching stale evidence.

### Compatibility and remaining operational risk

- Public Graph, Network Flow, Reporting, HTTP, WebSocket, and workbook rebuild
  contracts remain unchanged. `--operation-id` is an optional additive operator
  input that permits exact durable replay after response loss.
- No migration, cursor secret, retained worker, public route, ordinary retained
  writer, activation flag, or broad source-query capability was introduced.
- New backups carry the exact registry and binding artifacts. Historical input
  is accepted only through the packaged legacy empty-registry clear-only
  binding; every other unknown or mismatched binding fails closed.
- Older binaries fail closed on the new codec-registry digest. Retain the prior
  verified backup and matching binary until deployment creates and verifies a
  fresh backup with the new binary.
- An indeterminate publication target must be discarded and reinitialized.
  Safe telemetry remains limited to outcome, duration, and aggregate counts.
