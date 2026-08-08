# projections Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/projections`
- **Target label:** `projections`, derived from the final target-path segment and
  normalized to lowercase kebab case
- **Output path:** `docs/handoffs/projections-module-refactor-tracker.md`
- **Document status:** tracker-local normative plan; planning and documentation only
- **Allowed change in this session:** this tracker file only
- **Non-goals:** production implementation; behavior correction; test, owner,
  contract, generated artifact, migration, package, or harness-input changes;
  generated-file hand edits
- **Implementation gate:** a later authorized task MUST adopt the authority and
  authorization sequence in sections 6 and 7 before it changes implementation.

### Normative language and effect

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY** in
`PRF-*` requirements are normative for a later authorized refactor of this
target. They do not amend Core, adopt an architecture decision, authorize an
implementation, or define product behavior. When this tracker conflicts with an
adopted owner, the adopted owner prevails and the implementation MUST stop with
`BLOCKED: owner contradiction`.

| Requirement | Tracker-local rule |
| --- | --- |
| PRF-001 | The refactor MUST remain behavior-preserving unless a separately authorized behavior change names the affected owner requirement, contract, tests, migration posture, and rollback boundary. |
| PRF-002 | The implementation MUST NOT begin until the authority artifacts and separate Authorization A and Authorization B gates defined by PRF-060 exist. |
| PRF-003 | Public HTTP, WebSocket, cursor, authorization, saved-view, `view_row_v1`, telemetry, schema, migration, and source-semantic behavior MUST remain unchanged by default. |
| PRF-004 | Runtime code, tests, conformance, generators, and release evidence MUST NOT depend on this tracker, Appendix I, analysis notes, or other Markdown. |
| PRF-005 | Existing migrations MUST remain in place. Generated roots and generated harness/topology outputs MUST NOT be hand-edited. |

### Source hierarchy

1. Adopted subsystem NLSpecs apply only to their declared subsystem. The adopted
   Graph Projection NLSpec does not govern workbook-grid projections.
2. Core 00 through Core 04 own current implementation-conformance behavior.
3. Core 05 applies only to claim-bearing timed or fixture-sensitive publication;
   no such publication is planned here, so it is not applicable.
4. Domain vocabulary and implementation-support guides supply terminology,
   navigation, and execution support.
5. Current code and tests establish current implementation state.
6. This tracker, `temp/analysis-notes.md`, prior handoffs, Appendix I, the NLSpec
   writing guide, and the planning framework are evidence only and do not
   override an adopted owner.

No owner contradiction was found. Core 01 assigns source semantics and
query-descriptor intent to source owners and assigns physical derived-storage
lifecycle and generic projection coordination to Projections. The target is
therefore a legitimate subsystem, but its current root facade and owner
integration seams are broader and less consistent than the required boundary.
Repository nonconformance and stale explanatory evidence are findings, not
authority.

### Owner documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`, especially
  REQ-01-351 through REQ-01-368 and REQ-01-621 through REQ-01-626
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` for the
  explicit non-applicability determination
- `docs/graph_projection_nlspec.md` for its explicit graph-only scope
- `docs/opentelemetry-instrumentation-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `docs/research/nlspec-spec.md` for writing rigor only
- Non-normative evidence: `docs/spec/I_projection_authority_boundary_and_characterization.md`
  `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, and
  `temp/analysis-notes.md`

### Repository files inspected

All 29 files under `internal/modules/projections` were opened and are inventoried
in section 2. The inspection also covered the projection and workbook assembly
packages; Workbook query routes and tests; Collaboration event contracts and
tests; Timeline, Entities, Indicators, Assessments, Artifacts, Evidence, Parties,
and Tasks/Decisions provider contributions; Recovery and test-util adapters;
projection-provider, view-schema, OpenAPI, and WebSocket contracts; projection
migrations and authored/generated query inputs; generated Go and TypeScript
consumers; projection-backed Reporting providers; and verification/harness
mappings under `contracts/verification` and `tools`. Exact repository types were
confirmed in `internal/platform/postgres`, `internal/platform/querypage`,
`internal/platform/viewschema`, `internal/modules/workbook`,
`internal/modules/recovery/restorecontract`, and
`internal/modules/reporting/exportprovider`.

### Framework and explanatory-evidence deltas

- The planning framework intentionally presents possible module families; the
  live target contains no frontend controller or grid-vendor implementation.
  Those areas are contract consumers, not refactor destinations found in this
  package.
- Non-normative Appendix I describes an empty production root-import allowlist,
  an approved `internal/modules/projections/adapters` package, and a
  `provider_manifest_test.go` path. Live code instead permits numerous exact
  source-file root importers, approves no adapter package, leaves the adapters
  directory empty, and keeps manifest parity in
  `internal/app/projectionassembly/catalog_manifest_test.go`.
- Appendix I names `internal/modules/timeline/workbookprojection` as the Timeline
  facade. The live validation manifest names `internal/modules/timeline`.
- These deltas do not authorize choosing the Appendix over the repository. The
  Core package-level boundary requirement remains normative. This tracker now
  specifies a target design, but its adoption and implementation remain gated.

### Explicit defaults

| Subject | Required default |
| --- | --- |
| Observable behavior | Preserve exactly; ambiguity MUST resolve to preservation, not correction. |
| Descriptor and manifest versions | Descriptor v3 and validation manifest v4 remain current. If v3 cannot express a neutral query/read-shape contract, a separately authorized descriptor-version change is required. |
| Transactions | Caller-owned mutation transactions and restore-owned rebuild transactions remain atomic; no moved writer MAY begin, commit, or roll back a caller transaction. |
| Query order and page bounds | Existing explicit order, null placement, stable tie-breaker, bound `limit+1`, and no deep `OFFSET` remain mandatory. |
| Migrations | Existing migrations remain centrally located and are never moved or rewritten merely to express package ownership. |
| Test-only SQL | Test SQL requires an explicit fixture permission and never grants production table access. |
| Unknown behavior | Mark `TODO:` and block the affected slice; do not infer from a name, path, historical phase, or visible UI label. |

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/projections/assessments.go` | Applies and rebuilds assessment projection rows from assessment-owned source input. | `AssessmentSource`; `Store.ApplyAssessmentMutationTx`; `ApplyAssessmentMutationSQLTx`; `Store.RebuildIncidentAssessmentsTx` | `services.go`, provider factories, assessment assembly | pgx, UUID, assessment source DTOs | `rebuild_test.go`, `query_store_test.go`, assessment/workbook integration tests | Assessment view schema and `assessment_grid_projection` contract are affected indirectly | Projections for physical SQL; Assessments for source semantics | high | Existing example of the intended source-owner/projection-storage split. |
| `internal/modules/projections/boundary_guard_test.go` | Scans production Go imports and enforces the current projection import allowlist. | Test-only `TestProductionProjectionImportsUseApprovedFacades` | Harness row `module.projections.architecture...` | Go parser, filesystem, `provider_boundary.go` | Self | Projection-provider manifest import policy mirrors the code-backed list | Projections | high | Enforcement is based on source-file paths, contrary to REQ-01-626. |
| `internal/modules/projections/delete.go` | Deletes a derived row for the host, identity, or indicator view. | `Store.DeleteRowTx` | `Coordinator`, `EntityRows`, entity merge adapter | pgx, generated view-schema IDs | Entity merge and projection boundary tests; no complete provider deletion matrix found | Generated view-schema constants are consumed | Projections | high | Generic name masks intentionally partial provider coverage; characterize before redesign. |
| `internal/modules/projections/entity_grids.go` | Coordinates incident rebuilds for host, identity, and indicator projection tables. | Rebuild methods on `Store`; `IndicatorSource` | `RebuildService`, provider factories | entity and indicator provider contributions, pgx, telemetry | `rebuild_test.go`, entity/indicator integration tests | Provider descriptors and projection table contracts | Projections orchestration; source owners retain derivation intent | high | Host/identity call owner provider SQL directly; indicator uses an injected source. |
| `internal/modules/projections/provider_boundary.go` | Publishes the code-backed root/adapter/contract import policy. | `ImportPolicy`; `ProductionImportPolicy` | Boundary guard and projection manifest parity test | Static path strings | `boundary_guard_test.go`, projection assembly manifest test | `contracts/projection-providers/index.json` import policy | Projections | high | Exact file allowlist is a Core boundary mismatch; adapter list is empty. |
| `internal/modules/projections/provider_factories.go` | Converts provider descriptors and owner sources into executable provider callbacks. | `NewTimelineProvider`, `NewHostProvider`, `NewIdentityProvider`, `NewIndicatorProvider`, `NewAssessmentProvider`, `NewArtifactProvider`, `NewEvidenceProvider`, `NewPartyProvider`, `NewTaskRequestProvider`, `NewDecisionProvider` | `internal/app/projectionassembly` | Owner source interfaces, Store methods, provider registry types | Registry, assembly, rebuild, and parity tests | Code-backed provider registry and validation-only manifest | Projections | high | Central runtime binding point for all ten providers. |
| `internal/modules/projections/provider_registry.go` | Defines provider descriptors/capabilities, indexes views, and coordinates refresh and rebuild order. | `ProviderStatus`; `RestoreRebuildParticipation`; `ProviderCapabilities`; `ProviderDescriptor`; `Provider`; `Catalog`; `NewCatalog`; catalog accessors; Store refresh/rebuild methods | Projection assembly, Workbook catalog, Recovery, services, tests | pgx, provider validation, query surfaces | `provider_registry_test.go`, assembly catalog tests, restore/query tests | `contracts/projection-providers/index.json` and schema v4 | Projections | high | Code-backed runtime authority; manifest is validation-only. |
| `internal/modules/projections/provider_registry_test.go` | Characterizes provider validation, indexing, and deterministic rebuild order. | Test-only provider registry tests | Harness provider row | Provider registry | Self | Descriptor/manifest contract indirectly | Projections tests | medium | Must remain stable through any catalog type or facade move. |
| `internal/modules/projections/provider_registry_validation.go` | Validates descriptor schema, ownership, capabilities, query surfaces, and dependency ordering. | Internal validation surface | `NewCatalog` | schema-owner rules, `providercontract` | `provider_registry_test.go`, assembly catalog tests | Descriptor v3 and manifest v4 shape | Projections | high | Fail-closed invariants protect restore and query dispatch. |
| `internal/modules/projections/providercontract/query.go` | Stable internal persistence contract for owner-supplied compiled query descriptors. | `DescriptorSchemaVersion`; `FieldKind`; `QueryField`; `QuerySurface` | Timeline, Assessments, Artifacts, Evidence, Indicators, Parties, Tasks/Decisions, projection assembly | No runtime owner import | Query surface and owner integration tests | Descriptor schema v3 and view schemas | Projections contract; source owners supply intent | high | Not a public query language and must not redefine route semantics. |
| `internal/modules/projections/query.go` | Generic facade for provider-backed query paging and single-row loads. | `Store.SupportsQuerySurface`; `Store.QueryRows`; `Store.QueryRowsPage`; `Store.LoadRowTx` | `QueryService`, Workbook assembly, row facades, tests | pgx, query engine, querypage, viewschema, telemetry | `query_test.go`, `query_store_test.go`, Workbook integration tests | `view_row_v1`, view schemas, generated protocol consumers | Projections | high | Observable query behavior must remain route-compatible. |
| `internal/modules/projections/query_row.go` | Adapts scanned query-engine values into the common Workbook row shape. | Internal package surface | `query.go`, query tests | query engine and generic surface definitions | Query shape/parity tests | `view_row_v1` | Projections | high | Row shape, null, and collection behavior are contract-sensitive. |
| `internal/modules/projections/query_sql.go` | Adapts generic surfaces and normalized query metadata into bounded SQL. | Internal package surface | `query.go`, query tests | query engine, querypage, viewschema | Query SQL, timeline paging, and schema guard tests | Query and paging contracts | Projections | high | Must preserve bound values, nulls-last order, tie-breakers, and no deep `OFFSET`. |
| `internal/modules/projections/query_store_test.go` | Store-backed parity across source mutation/import/rebuild and query/load-row behavior. | Test-only `TestProjectionStoreQueryRowsAndLoadRowTxParity` | Harness storage row | Application composition, source-owner modules, Postgres test support | Self | Query row and provider contracts | Projections tests | medium | Integration evidence; host/identity specialized paths need separate owner evidence. |
| `internal/modules/projections/query_surfaces_test.go` | Builds test-only generic surfaces from owner query descriptors. | Test helper surface only | `query_test.go` | Owner projection-provider packages and `providercontract` | Query unit tests | View-schema/query descriptor mapping | Projections tests | low | Test-only provider imports do not grant production import permission. |
| `internal/modules/projections/query_test.go` | Characterizes safe SQL, view coverage, filters, rows, grouping, collections, and keyset paging. | Nine test functions | Harness query-contract row | Generic query helpers and contract surfaces | Self | View contracts and `view_row_v1` | Projections tests | medium | Core characterization barrier before query-engine movement. |
| `internal/modules/projections/query_types.go` | Holds internal generic query-surface and field representations/conversions. | Internal package surface | Query implementation and tests | `providercontract`, query engine | Query tests | Descriptor and view contracts indirectly | Projections | medium | Keep persistence types private to the subsystem. |
| `internal/modules/projections/queryengine/engine.go` | Compiles parameterized query/filter/keyset SQL, scans values, and builds common rows. | `FieldKind`; `Field`; `Surface`; `ScanRows`; `BuildQueryPageSQL`; `BuildRow` | Root query implementation and tests | pgx, querypage, viewschema, `links/readshape` | Root query tests | `view_row_v1` and normalized view-query semantics | Projections | high | Private engine has a peer-module helper dependency on Links. |
| `internal/modules/projections/rebuild.go` | Implements Recovery-facing deterministic restore rebuild and structured readiness/results. | `RestoreRebuilder`; constructors; `RebuildRestoreProjections` | Recovery assembly and `RebuildService` | Recovery restore contract, provider catalog, pgx | `rebuild_test.go`, Recovery integration tests | Restore request/result contracts and recovery state manifest | Projections implementation; Recovery orchestrates | high | Commit failure must not publish successful rebuilt-resource claims. |
| `internal/modules/projections/rebuild_test.go` | Characterizes fail-before-touch, deterministic source paging, stale-row replacement, ordering, and commit-failure claims. | Five test functions plus test doubles | Harness restore row | Postgres test support, Timeline/Assessment sources, Recovery contract | Self | Restore readiness and provider result contracts | Projections tests | medium | Retry, no-provider, and partial-provider failure coverage needs explicit review. |
| `internal/modules/projections/recovery_state.go` | Declares all ten projection tables as derived and rebuildable recovery state. | `RecoveryStateContribution` | Recovery assembly | recoverystate contract | Recovery/catalog tests | Recovery state catalog | Projections | high | Algorithm ID and all table declarations are compatibility-sensitive. |
| `internal/modules/projections/services.go` | Publishes root query/rebuild/coordinator services and owner-specific row facades. | `QueryService`; `RebuildService`; `Coordinator`; `TimelineRows`; `EntityRows`; `AssessmentRows`; `ArtifactRows`; `EvidenceRows`; `PartyRows`; `TaskDecisionRows`; constructors and methods | Application assembly, source owners, Workbook, Recovery, testutil | Store, owner DTOs/providers, pgx, querypage | Broad owner, Workbook, projection, import, and recovery tests | Provider, query, row, and restore contracts | Projections | high | Mixed and overly broad root facade; exact replacement API requires later authorization. |
| `internal/modules/projections/store.go` | Owns Store construction and physical Timeline projection upsert/delete/rebuild SQL. | `Store`; `TimelineSource`; `NewStore`; Timeline mutation and rebuild methods | Services, providers, assembly, tests | pgx, Postgres, Timeline projection DTOs, telemetry | Rebuild, query parity, Timeline integration tests | `timeline_grid_projection` migration and Timeline view schema | Projections | high | Clear example of physical storage ownership with source extraction injected. |
| `internal/modules/projections/tasks_decisions.go` | Coordinates task-request and decision incident rebuild through an injected owner source. | `TaskDecisionSource`; Store rebuild methods | Provider factories and services | pgx, Tasks/Decisions source contribution | Query parity, rebuild, import, Workbook tests | Task/decision projection tables and view schemas | Projections orchestration; Tasks/Decisions source semantics | high | Owner provider currently contains physical write SQL. |
| `internal/modules/projections/telemetry.go` | Emits bounded, safe projection-operation telemetry. | Internal telemetry helper surface | Store/query/rebuild operations | Platform telemetry | `telemetry_test.go` | OTel semantic contract | Projections | medium | Only view schema ID, operation, and result are allowed attributes. |
| `internal/modules/projections/telemetry_test.go` | Guards telemetry vocabulary and SDK independence. | Two test functions | Harness telemetry row | Source inspection and helper assertions | Self | OTel contract | Projections tests | low | Static evidence only. |
| `internal/modules/projections/timeline_query_paging_test.go` | Guards bounded Timeline keyset SQL. | `TestTimelinePageSQLIsKeysetBounded` | Harness full-tier Timeline paging row | Timeline query descriptor and SQL builder | Self | Workbook paging contract | Projections tests with Timeline collaboration | medium | Uses a dedicated Postgres fixture profile. |
| `internal/modules/projections/timeline_query_schema_guard_test.go` | Verifies Timeline query fields remain mapped to the view schema. | `TestUnit_TimelineQuerySchemaMappingGuard` | Harness query-contract row | Timeline contract/query surfaces | Self | Timeline view schema | Projections tests with Timeline collaboration | medium | Detects descriptor/schema drift. |
| `internal/modules/projections/workbook_hot.go` | Coordinates rebuilds for Artifact, Evidence, and Party hot-view projection tables. | Store rebuild methods for artifacts, evidence, and parties | Rebuild service and provider factories | Source-owner projection-provider packages, pgx, telemetry | Rebuild, query parity, Workbook/source-owner integration tests | Artifact/evidence/party view and table contracts | Projections orchestration; source owners retain derivation intent | high | Direct peer provider coupling must be reviewed with physical storage ownership. |

No target file is out of scope. The empty
`internal/modules/projections/adapters` directory has no file to inventory, but
its mismatch with the explanatory boundary design is a material finding.

## 3. Module Boundary Diagnosis

The target is a legitimate projection persistence and application-orchestration
subsystem. It is also a mixed-responsibility package because the root exports its
runtime catalog, generic services, restore adapter, and owner-specific row
facades to exact source files. It is not a frontend shell, grid-vendor layer,
public transport owner, authoritative source owner, or generic mutation/revision
owner.

### Target topology

| Requirement | Normative target |
| --- | --- |
| PRF-010 | `internal/modules/projections/adapters` MUST be the sole production construction boundary. Only `internal/app/projectionassembly` MAY import it. |
| PRF-011 | `internal/modules/projections/providercontract` MUST contain only immutable descriptor and query-intent types. It MUST contain the immutable descriptor facts currently exported by the root and MUST NOT expose executable providers, callbacks, catalogs, stores, mutable SQL executors, or table operations. |
| PRF-012 | `internal/modules/projections/testsupport` MUST contain test-only immutable capabilities. Its permissions MUST remain distinct from production permissions. |
| PRF-013 | `internal/modules/projections/internal/runtime`, `internal/storage`, and `internal/queryengine` MUST contain compiler-private implementation. Nested Go `internal` placement MUST supplement repository boundary checks with compiler-enforced containment. |
| PRF-014 | After transitional removal, the root `internal/modules/projections` package MUST expose no production API and MUST have an empty production importer set. |

```text
internal/modules/projections/
├── adapters/                 # sole production constructor
├── providercontract/         # immutable descriptor/query-intent contracts
├── testsupport/              # immutable test-only capabilities
└── internal/
    ├── runtime/              # catalog, validation, orchestration
    ├── storage/              # projection-table SQL and writers
    └── queryengine/          # SQL compilation, scanning, row materialization
```

The topology follows Go's compiler-enforced `internal` import rule as an
additional containment mechanism; the repository policy remains independently
required. See the [Go module layout guidance](https://go.dev/doc/modules/layout).

### Production import policy

| Importer package | Target import | Permission | Required condition |
| --- | --- | --- | --- |
| Any production package | Root `internal/modules/projections` | forbidden | The root importer set MUST equal the empty set after S-08. |
| `internal/app/projectionassembly` | `projections/adapters` | sole production constructor | No other package MAY construct or receive concrete runtime/storage/query-engine types. |
| `internal/app/projectionassembly` | `projections/providercontract` | immutable descriptor assembly | Only descriptor snapshots/query intent MAY cross this seam. |
| The eight source-owner facades below | `projections/providercontract` | immutable query intent | No facade MAY import `adapters`, root, or compiler-private packages. |
| Tests named by `testsupport` policy | `projections/testsupport` | test only | A test permission MUST NOT imply a production permission. |
| Any package outside `projections` | `projections/internal/**` | compiler-forbidden | No repository allowlist MAY attempt to override Go `internal` containment. |

### Source-owner facade map

PRF-015 requires each source owner to retain semantic mapping and descriptor
intent behind the following package-level facade. These eight rows are
exhaustive for the current ten providers.

| Source owner | Required facade package | Providers represented | Allowed exported content | Forbidden content |
| --- | --- | --- | --- | --- |
| Timeline | `internal/modules/timeline/workbookprojection` | timeline | Immutable descriptor/query intent; typed source and writer interfaces | Projection SQL execution, Store, adapter construction |
| Entities | `internal/modules/entities/workbookprojection` | host, identity | Host/identity descriptor intent, source hydration, typed writers | Projection SQL or generic query engine |
| Indicators | `internal/modules/indicators/workbookprojection` | indicator | Indicator descriptor/source intent and typed writer | Generic `viewSchemaID` mutation dispatch |
| Assessments | `internal/modules/assessments/workbookprojection` | assessment | Assessment descriptor/source intent and typed writer | Projection-table lifecycle |
| Artifacts | `internal/modules/artifacts/workbookprojection` | artifact | Artifact descriptor/source/report fact semantics and typed writer | Projection-table SQL execution |
| Evidence | `internal/modules/evidence/workbookprojection` | evidence | Evidence descriptor/source intent and typed writer | Projection-table lifecycle |
| Parties | `internal/modules/parties/workbookprojection` | party | Party descriptor/source intent and typed writer | Projection-table lifecycle |
| Tasks/Decisions | `internal/modules/tasksdecisions/workbookprojection` | task request, decision | Task/decision descriptor/source/report fact semantics and typed writers | Projection-table SQL execution |

### Current root-surface disposition

| Current root surface | Target disposition | Target consumer/interface | Exposure rule |
| --- | --- | --- | --- |
| `Store`, `NewStore` | Move to `internal/storage` | Constructed only inside `adapters.New` | MUST NOT escape the adapter. |
| `Provider`, executable callbacks, `Catalog`, `NewCatalog` | Move to `internal/runtime` | Adapter-private runtime | MUST NOT escape; immutable descriptors MAY be snapshotted separately. |
| `ProviderDescriptor`, status/capability/rebuild facts | Move to `providercontract.ProviderDescriptor` and immutable companion types | Source facades and projection assembly | Descriptor v3 semantics MUST remain unchanged. |
| `QueryService` | Remove after migration | Existing `workbook.QueryProvider` | Workbook receives only the consumer-owned interface. |
| `RebuildService`, `RestoreRebuilder` constructors | Remove/split after migration | Existing `restorecontract.ProjectionRebuilder` | Recovery receives only its consumer-owned interface. |
| `Coordinator` | Remove | Typed source-owner writer ports | Generic coordination MUST NOT remain public. |
| `TimelineRows` | Remove | `timeline/workbookprojection.Writer` | Typed caller-owned transaction methods only. |
| `EntityRows` | Split/remove | `entities/workbookprojection.Writer` | Typed host and identity methods only. |
| `AssessmentRows` | Remove | `assessments/workbookprojection.Writer` | Typed assessment methods only. |
| `ArtifactRows` | Remove | `artifacts/workbookprojection.Writer` | Typed artifact methods only. |
| `EvidenceRows` | Remove | `evidence/workbookprojection.Writer` | Typed evidence methods only. |
| `PartyRows` | Remove | `parties/workbookprojection.Writer` | Typed party methods only. |
| `TaskDecisionRows` | Split/remove | `tasksdecisions/workbookprojection.Writer` | Typed task-request and decision methods only. |
| `TimelineSource`, `IndicatorSource`, `AssessmentSource`, `TaskDecisionSource` | Replace with eight source-facade inputs in `Sources` | `adapters.New` and private storage/runtime | Source interfaces remain consumer-/source-owner-defined and MUST NOT escape. |
| `RestoreRebuilder`, `NewRestoreRebuilder`, `NewRestoreRebuilderFromStore` | Move implementation to private runtime/storage and remove constructors | `restorecontract.ProjectionRebuilder` | Only the Recovery-owned interface escapes. |
| `NewTimelineProvider` through `NewDecisionProvider` | Move all ten factories to private runtime adapter assembly | `adapters.New` | Executable providers/callbacks MUST NOT escape. |
| `RecoveryStateContribution` | Retain immutable contribution behind projection assembly | Recovery state assembly | Exact ten table IDs and algorithm identity remain contract-sensitive. |
| `ApplyAssessmentMutationSQLTx` | Move to private storage | Assessment writer implementation | Raw SQL helper MUST NOT escape. |
| Current `queryengine.FieldKind`, `Field`, `Surface`, and build/scan functions | Move under `projections/internal/queryengine` | Private Workbook query implementation | No source owner or consumer may import engine types. |
| Import policy export | Move to a private/code-backed policy owner | Boundary tests and authored validation mirror | Policy data MUST NOT be a production service API. |
| Test snapshots/helpers | Move to `testsupport` only where immutable | Approved tests | MUST NOT expose Store, SQL, callbacks, or construction. |

### Current-state responsibility diagnosis

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Physical derived projection storage lifecycle | Root Store files plus several source-owner provider packages | Projections, supplied with owner-defined source/descriptor intent | split | Core 01 REQ-01-622; provider descriptors name `projection_storage_owner_module=projections` | Exact SQL relocation requires later authorization. |
| Provider catalog, validation, and deterministic order | `provider_registry*.go`, projection assembly | Projections | keep | Code-backed catalog and manifest parity tests | Runtime authority remains code-backed. |
| Generic query compilation, paging, scanning, and row materialization | `query*.go`, `queryengine/engine.go` | Projections | keep | Core query contract and target query tests | Peer Links helper coupling should be removed or replaced by a neutral contract. |
| Source semantics and compiled query descriptor intent | Timeline, Entities, Indicators, Assessments, Artifacts, Evidence, Parties, Tasks/Decisions | Respective source owner | keep | Core 01 REQ-01-622 and owner provider packages | Do not move authoritative derivation meaning into Projections. |
| Root application/service facade | `services.go` and root exported catalog types | Package-level Projections adapters/contracts | split | REQ-01-626; live root importer list | Section 4 specifies the target interface; adoption and implementation remain pending. |
| Projection refresh during authoritative mutations | Source-owner transaction coordinators calling Projections | Source owner coordinates transaction; Projections applies derived row | keep | Store/assembly call graph and REQ-01-351 | Revisions and source mutation remain outside Projections. |
| Restore rebuild implementation | `rebuild.go` | Projections | keep | REQ-01-624/625 and recovery adapter | Recovery retains restore orchestration and readiness publication. |
| HTTP query route, envelope, validation, and authorization | `internal/modules/workbook/routes.go` and platform query utilities | Workbook/transport and platform contracts | move | Current route and authorization tests | No code is currently misplaced in the target; “move” means keep this responsibility outside it. |
| WebSocket `record_changed` publication | Collaboration plus source-owner transaction intents | Collaboration | move | Collaboration event and replay tests | Projections supplies rows/affected-view consequences but does not own the socket. |
| Saved-view persistence and startup selection | Saved Views and Workbook | Saved Views/Workbook | move | Saved-view and Workbook startup tests | Query normalization compatibility is the only projection coupling. |
| Host/identity specialized projection queries | Entities hostidentity store | Entities descriptor intent with Projections-owned physical query path | split | Live specialized SQL and PRF-032 | Projections selects bounded IDs; Entities hydrates exactly those IDs without reordering or requerying. |
| Test-only projection capabilities | `internal/testutil/httptestx` through `timelineassembly.Bundle` | Test-util capability backed by projection assembly | split | Live testutil composition | Test-only change; do not create production permissions. |
| Frontend shell/controller and grid-vendor integration | No implementation in target | `apps/web` and `packages/grid-adapter` | defer | No target imports or files found | Validate only if a later backend change affects observable selectors or rows. |

## 4. Public Contract and Behavior Freeze Map

### Adapter interface and construction contract

PRF-020 requires the adopted architecture decision to define the following
consumer-visible shape using the repository's existing types. Package aliases
below identify ownership; they are not a code patch.

```go
// package projections/adapters
type Dependencies struct {
    DB          postgres.DB
    Descriptors []providercontract.ProviderDescriptor
    Sources     Sources
}

type Ports struct {
    Workbook        workbook.QueryProvider
    Restore         restorecontract.ProjectionRebuilder
    Timeline        timelineprojection.Writer
    Entities        entityprojection.Writer
    Indicators      indicatorprojection.Writer
    Assessments     assessmentprojection.Writer
    Artifacts       artifactprojection.Writer
    Evidence        evidenceprojection.Writer
    Parties         partyprojection.Writer
    TasksDecisions  taskdecisionprojection.Writer
    ReportingProjectionProviders []exportprovider.FieldProvider
}

func New(Dependencies) (Ports, error)
```

The facade packages in section 3 own the final Go names represented by the
`*projection` aliases. `Sources` MUST be an explicit aggregate of the eight
source-owner facades; it MUST NOT be `any`, a callback map, or a service
locator. Immutable descriptors use
`providercontract.ProviderDescriptor`. Workbook querying MUST use the existing
`workbook.QueryProvider`, whose `QueryRowsPage(context.Context,
workbook.QueryCommand)` consumes `viewschema.QueryMeta` and `querypage.Window`
through `workbook.QueryCommand` and returns `querypage.Result`. Recovery MUST
continue to consume `restorecontract.ProjectionRebuilder`. `Dependencies.DB`
MUST use `postgres.DB`, not a concrete pool.

PRF-021 requires `adapters.New` to fail closed. It MUST return an error and a
zero, unusable `Ports` value for every condition below. On success, every
required port MUST be non-nil.

| Construction condition | Required result | Default |
| --- | --- | --- |
| Nil `postgres.DB` | fail | No deferred panic or partially usable port. |
| Nil required source dependency | fail and name the missing source role | All eight source facades are required unless the adopted descriptor set proves the source is inactive and the architecture decision explicitly permits omission. |
| Duplicate provider ID, table ID, or view ownership | fail | No first-wins or last-wins selection. |
| Missing active descriptor or missing dependency | fail | All ten current active providers remain required. |
| Unsupported descriptor version | fail | Descriptor v3 is the current accepted version. |
| Unresolved declared query surface | fail | No active query-capable descriptor may construct without a complete compiled surface. |
| Invalid source/storage/facade ownership | fail | Section 5 mappings are exact. |
| Incomplete or cyclic restore participation | fail | The ten-provider deterministic order must validate. |
| Successful validation | return complete `Ports` | No returned required interface may hold a nil concrete value. |

Successful construction MUST NOT expose a Store, runtime catalog, executable
provider, callback, SQL fragment, table name, query-engine type, or mutable
descriptor collection. `internal/app/projectionassembly` MAY retain an immutable
descriptor snapshot for Workbook assembly, but that snapshot MUST be separate
from `Ports` and MUST contain no executable capability.

### Typed mutation and query interfaces

PRF-022 requires source writers to accept `context.Context`, caller-owned
`pgx.Tx`, and typed source-owner inputs. Existing canonical row values remain
`map[string]any` where that is the current consumer contract. A writer MUST NOT
begin, commit, or roll back the transaction; perform authorization; mutate
source/history rows; publish WebSocket events; reinterpret source errors; or
expose SQL/table names.

| Facade writer | Required typed operations | Explicitly unsupported generic operation |
| --- | --- | --- |
| Timeline | `ApplyTimelineMutationTx(..., timelineprojection.ProjectionMutation)`; `RefreshHostTx(..., uuid.UUID)`; `RefreshIdentityTx(..., uuid.UUID)` | `Refresh(viewSchemaID, any)` and generic deletion |
| Entities | `RefreshHostTx`; `RefreshIdentityTx`; `DeleteHostTx`; `DeleteIdentityTx`; typed host/identity rebuild and load operations where current consumers require them | Generic entity-type or view-schema dispatch after migration |
| Indicators | `RefreshIndicatorTx`; `DeleteIndicatorTx`; typed rebuild/load operations | Generic `RefreshRowTx(viewSchemaID, ...)` and generic deletion |
| Assessments | Typed refresh/apply mutation/load/rebuild operations using the current assessment mutation type and `map[string]any` row | `Refresh(viewSchemaID, any)` |
| Artifacts | Typed artifact refresh/load/rebuild operations | Generic view-schema dispatch after artifact-view call sites are split |
| Evidence | Typed evidence refresh/load/rebuild operations | Generic view-schema dispatch |
| Parties | Typed party refresh/load/rebuild operations | Generic view-schema dispatch |
| Tasks/Decisions | Separate typed task-request and decision refresh/load/rebuild operations | Generic combined deletion or view-schema dispatch |

Only typed host, identity, and indicator deletion plus the existing Timeline
mutation deletion are in scope. The adapter MUST NOT invent deletion support for
assessment, artifact, evidence, party, task request, or decision providers. Any
new deletion behavior requires separate owner authorization.

PRF-023 requires projection-backed Reporting providers to be returned as
`exportprovider.FieldProvider` interfaces for `artifacts`,
`entities.hostidentity`, and `tasksdecisions`. Their source owners MUST retain
fact selection, semantic mapping, content-class, and source-family rules;
Projections MUST execute the projection-table SQL. The adapter MUST NOT expose a
Reporting-specific Store.

### Observable contract freeze

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | Workbook route; viewquery/viewschema normalize | Core 01 REQ-01-356; OpenAPI owner source; live route | Workbook projection route, paging, live-cursor, and tamper tests | Preserve route identity, request members, success/error envelopes, and cursor binding for every provider path | high | Target must not absorb transport behavior. |
| `view_row_v1` row shape | Workbook/view contracts; Projections materializes | Core 01 query/row requirements and generated contracts | Generic row shape, null/collection/group tests; store parity | Add parity at any new adapter seam, including host/identity specialized rows | high | Preserve top-level `record_id`, `row_version`, scalar cells, and collections. |
| Keyset paging | Workbook owns opaque cursor; Projections/provider queries own bounded page retrieval | Core 01 REQ-01-356; query SQL | Generic and Timeline keyset tests; Workbook pagination tests | Prove normalized tuple, default sort tail, `record_id` tie-breaker, nulls-last behavior, and bound `limit+1` | high | No deep pagination `OFFSET`. |
| Transactional projection refresh/delete | Source owner coordinates; Projections applies derived state | REQ-01-351 and owner mutation paths | Source-owner integration, query parity, merge tests | Characterize supported deletion matrix before changing `DeleteRowTx` | high | Source/revision mutation must remain atomic with refresh. |
| Deterministic incident rebuild | Projections | REQ-01-353/621 and provider catalog | Rebuild order, source paging, stale-row replacement | Repeat same source state and compare provider results/row counts | high | Rebuild must not mutate authoritative state. |
| Restore request/result and readiness | Recovery owns orchestration; Projections implements adapter | REQ-01-624/625; restore contracts | Invalid request, commit failure, result/order tests | Add explicit retry, no-active-provider, unsupported-provider, and partial-failure posture where absent | high | No pre-commit success claims. |
| `record_changed.affected_views` patch/invalidate/remove semantics | Collaboration and source owners | Core 01 WebSocket contract; WS schema | Collaboration intent/hub/replay tests; browser collaboration scenarios | Preserve row load and affected-view payload after facade migration | high | Socket path, auth, replay, and sequencing stay unchanged. |
| Incident query authorization | Workbook route and Incident membership | Core 04 authorization model | Live authorized cursor, continuation reauthorization, membership tests | Run unchanged after query facade migration | high | Deployment admin is not an implicit bypass. |
| Saved-view query compatibility | Saved Views/Workbook | Core 02/03 and saved-view contracts | Saved-view create/patch/startup and coordination tests | Preserve normalized sort/filter/group fields and schema evolution | medium | Projections does not own saved-view persistence. |
| Projection table/storage semantics | Projections lifecycle; source owners supply intent | Ten migrations, provider descriptors, recovery state | Source-owner and restore integration tests | Preserve per-table stale-row replacement, indexes, and non-authoritative posture | high | No migration is proposed by this plan. |
| Provider descriptor v3/catalog | Projections | Core 01 REQ-01-622; code-backed registry | Registry and manifest parity tests | Preserve all ten providers, capabilities, facades, table/view ownership, and rebuild order | high | Manifest v4 is authored validation evidence, not runtime authority. |
| Generated protocol/view contracts | Contract owners and generators | View schemas, OpenAPI, WS schema, generated Go/TS consumers | Protocol/view-contract package tests and drift checks | Regenerate only through owner inputs and Make targets if a later contract change is authorized | high | Never hand-edit generated roots. |
| Projection telemetry | Projections with platform telemetry policy | OTel NLSpec | Safe vocabulary and no-SDK tests | Preserve span name and bounded attributes after file/package moves | medium | Never emit SQL, table names, row IDs, or sensitive values. |
| Test-util projection capabilities | Test-util/application composition | `internal/testutil/httptestx` and appsupport | Store/server-backed owner tests | Characterize selected rebuild and owner-port capabilities before decoupling Timeline | medium | No production contract, but broad evidence depends on it. |
| Harness/test accounting | Verification owners and test catalog | Eight active `module.projections` rows | `make explain-test-owner OWNER=module.projections` | Keep row ownership, selectors, fixture profiles, and evidence classes current | medium | Phase maps and rows are evidence accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

### Provider, table, facade, writer, SQL, and rebuild map

PRF-030 requires every runtime access to a projection table to execute inside
Projections. Source owners MUST retain authoritative reads and pure semantic
mapping. “Target SQL destination” below is an ownership destination, not a
prescribed filename. The provider order is exhaustive and deterministic.

| Order | Provider | Projection table | Source owner | Required facade | Required writer operations | Current projection SQL location | Target SQL destination |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | timeline | `timeline_grid_projection` | Timeline | `timeline/workbookprojection` | Timeline mutation apply; host/identity refresh; incident rebuild | `projections/store.go`; Timeline query intent; authored/generated Timeline query inputs | `projections/internal/storage` |
| 2 | host | `host_grid_projection` | Entities | `entities/workbookprojection` | host refresh/delete/load/rebuild | `entities/hostidentity/projectionprovider/provider.go`; specialized query store; root orchestration | `projections/internal/storage` |
| 3 | identity | `identity_grid_projection` | Entities | `entities/workbookprojection` | identity refresh/delete/load/rebuild | `entities/hostidentity/projectionprovider/provider.go`; specialized query store; root orchestration | `projections/internal/storage` |
| 4 | indicator | `indicator_grid_projection` | Indicators | `indicators/workbookprojection` | indicator refresh/delete/load/rebuild | `indicators/internal/providers/projection/provider.go`; root orchestration | `projections/internal/storage` |
| 5 | assessment | `assessment_grid_projection` | Assessments | `assessments/workbookprojection` | typed mutation/refresh/load/rebuild | `projections/assessments.go`; assessment query intent | `projections/internal/storage` |
| 6 | artifact | `artifact_grid_projection` | Artifacts | `artifacts/workbookprojection` | artifact refresh/load/rebuild | `artifacts/projectionprovider/provider.go`; `projections/workbook_hot.go`; artifact Reporting provider | `projections/internal/storage` |
| 7 | evidence | `evidence_grid_projection` | Evidence | `evidence/workbookprojection` | evidence refresh/load/rebuild | `evidence/projectionprovider/provider.go`; `projections/workbook_hot.go` | `projections/internal/storage` |
| 8 | party | `party_grid_projection` | Parties | `parties/workbookprojection` | party refresh/load/rebuild | `parties/projectionprovider/provider.go`; `projections/workbook_hot.go` | `projections/internal/storage` |
| 9 | task_request | `task_request_grid_projection` | Tasks/Decisions | `tasksdecisions/workbookprojection` | task-request refresh/load/rebuild | `tasksdecisions/internal/providers/projection/provider.go`; root orchestration; Tasks/Decisions Reporting provider | `projections/internal/storage` |
| 10 | decision | `decision_grid_projection` | Tasks/Decisions | `tasksdecisions/workbookprojection` | decision refresh/load/rebuild | `tasksdecisions/internal/providers/projection/provider.go`; root orchestration; Tasks/Decisions Reporting provider | `projections/internal/storage` |

The rebuild dependency chain MUST remain exactly `timeline -> host -> identity ->
indicator -> assessment -> artifact -> evidence -> party -> task_request ->
decision`. Repeated construction and rebuild with the same descriptor and source
state MUST produce that order.

### Projection-table access rules

PRF-031 requires ten explicit production table-access rules. Each rule MUST
permit Projections runtime/storage code and MUST deny direct runtime access by
source modules, Reporting providers, Workbook, Recovery, and application
assembly. Test-only access requires a separately declared fixture exception.

| Rule ID | Table | Permitted production executor | Required denied examples |
| --- | --- | --- | --- |
| `projection.timeline` | `timeline_grid_projection` | Projections | Timeline, Workbook, Reporting |
| `projection.host` | `host_grid_projection` | Projections | Entities hostidentity, Reporting |
| `projection.identity` | `identity_grid_projection` | Projections | Entities hostidentity, Reporting |
| `projection.indicator` | `indicator_grid_projection` | Projections | Indicators, Workbook |
| `projection.assessment` | `assessment_grid_projection` | Projections | Assessments, Workbook |
| `projection.artifact` | `artifact_grid_projection` | Projections | Artifacts, Reporting |
| `projection.evidence` | `evidence_grid_projection` | Projections | Evidence, Workbook |
| `projection.party` | `party_grid_projection` | Projections | Parties, Workbook |
| `projection.task_request` | `task_request_grid_projection` | Projections | Tasks/Decisions, Reporting |
| `projection.decision` | `decision_grid_projection` | Projections | Tasks/Decisions, Reporting |

The implementation MUST prove exact set equality among:

1. active `providercontract.ProviderDescriptor` projection-table IDs;
2. production boundary-policy projection-table rule IDs;
3. Projections recovery-state table IDs; and
4. schema objects assigned to Projections by the schema ownership manifest.

Missing, duplicate, extra, inactive, or differently named members MUST fail the
boundary check. Assembly source-ownership scanning MAY add evidence, but MUST
NOT substitute for this four-way equality.

### SQL and semantic ownership closure

| Access category | Required owner | Required treatment |
| --- | --- | --- |
| Projection-table `SELECT`, `INSERT`, `UPDATE`, `DELETE`, and rebuild cleanup | Projections | MUST execute under `projections/internal/storage` or `internal/queryengine`. |
| Authoritative source reads | Named source owner | MUST remain behind its facade/source interface. |
| Query intent, field meaning, fact mapping, content class, source family | Named source owner | MUST remain immutable semantic input to Projections. |
| Artifact, host/identity, and task/decision projection-backed Reporting SQL | Projections execution; source owner semantics | MUST return `exportprovider.FieldProvider`; MUST preserve Reporting output semantics. |
| `db/queries/timeline.sql` and `internal/gen/sql/timeline.sql.go` projection access | Authored query input and generator-owned output | MUST enter the signed SQL inventory; generated Go MUST NOT be hand-edited. An unused query MUST be removed only through its authored owner and generator. |
| Migrations declaring ten projection tables | Migration owners | MUST remain in place and MUST NOT be rewritten for package ownership. |

PRF-032 requires the host/identity query path to select only a bounded, ordered
page of IDs under Projections and then ask Entities to hydrate exactly those IDs.
Entities MUST NOT reorder, filter, or independently requery the page. The result
MUST retain `limit+1` bounds and MUST NOT use deep `OFFSET`.

PRF-033 requires the generic query engine to stop importing
`links/readshape`. Links MAY supply immutable descriptor intent through a
neutral contract. If descriptor v3 cannot express the required read shape,
implementation MUST stop until a separately authorized descriptor-version
change exists.

### Classified findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Root import enforcement uses exact source-file paths. | `provider_boundary.go`; `boundary_guard_test.go`; Core 01 REQ-01-626 | Boundary permissions survive by incidental filename rather than approved package contract. | `must_fix` | Projections boundary policy | Design a package-path adapter/contract policy before moving any caller. |
| Live policy allows many root importers and no adapter package, contrary to Appendix I evidence. | Code-backed policy, manifest v4, empty adapters directory, Appendix I | Broad facade expansion and stale human guidance. | `must_fix` | Projections plus application assembly | Treat Core as authority; reconcile code, manifest, guard, and explanatory evidence only in a later authorized task. |
| Root package exports catalog runtime, query/rebuild services, and owner-specific row facades together. | `services.go`, root import graph | Large blast radius and weak semantic ownership. | `should_fix` | Projections adapters/contracts | Split stable facade roles without changing observable behavior. |
| Physical projection read/write SQL is distributed across Projections, source-owner providers, specialized query stores, and three Reporting providers. | Provider SQL sources; descriptors name Projections storage owner; reporting providers for artifact, host/identity, and task/decision facts | Physical-lifecycle ownership is incomplete and hard to enforce. | `should_fix` | Projections for projection-table execution; source owners for source semantics | Sign the exhaustive SQL inventory in section 5 before moving one provider at a time. |
| Generic query engine imports `links/readshape`. | `queryengine/engine.go` | Generic engine depends on a peer semantic helper. | `should_fix` | Links supplies descriptor intent; Projections consumes neutral compiled input | Characterize tag filter/row semantics, then choose a neutral contract or local adapter. |
| Host/identity queries duplicate generic paging/row responsibilities. | Entities hostidentity Store versus Projections query engine | Query drift, especially nulls-last ordering and row hydration. | `should_fix` | Entities intent plus Projections physical query path | Preserve bounded hydration and compare behavior before unification. |
| Generic delete facade supports only host, identity, and indicator. | `delete.go`; call search finds host/identity merge callers | A future caller may assume uniform provider coverage. | `should_fix` | Typed source-owner ports backed by Projections | Characterize, replace with typed host/identity/indicator deletion, and do not expand behavior. |
| Test-util projection capability is backed by `timelineassembly.Bundle`. | `internal/testutil/httptestx` | Test composition obscures projection ownership and complicates facade migration. | `should_fix` | Test-util plus projection assembly | Plan a test-only projection capability port after the production facade design. |
| SQL table-access policy covers four of ten projection tables. | `tools/backend_module_boundaries.json`; broader but different projection assembly SQL test | Storage boundary coverage is inconsistent across enforcement mechanisms. | `should_fix` | Harness boundary policy and Projections | Add all ten exact rules and four-way equality after authority adoption; update owner inputs, not generated outputs. |
| Workbook owns route admission, envelopes, and membership checks. | Workbook routes and integration tests | Moving these checks into Projections could weaken authorization. | `intentional/no_action` | Workbook/Incident authorization | Freeze and rerun existing tests; do not relocate. |
| Collaboration owns socket publication/replay. | Collaboration hub, intent, stream, and browser tests | Moving publication into Projections would mix derived storage with transport. | `intentional/no_action` | Collaboration | Preserve row/affected-view inputs only. |
| Saved Views owns persisted query/layout state. | Saved-view store and Workbook startup tests | Projection refactor could accidentally reinterpret stored queries. | `intentional/no_action` | Saved Views/Workbook | Run compatibility tests when query descriptors move. |
| No direct grid-vendor or frontend-shell import exists in target. | Full target import scan | Unnecessary frontend movement would expand scope. | `intentional/no_action` | Web shell and grid adapter | Keep out of backend module refactor unless an observable contract changes. |
| Platform Postgres, querypage, viewschema, recovery-state, and telemetry dependencies are used at adapter boundaries. | Target imports and package roles | Indiscriminate removal would obscure legitimate persistence/runtime responsibilities. | `intentional/no_action` | Projections and platform contracts | Retain unless a concrete inversion is required by facade design. |
| Target tests import owner provider internals only in test code. | `query_surfaces_test.go`; boundary guard excludes `_test.go` | Test imports could be mistaken for production permission. | `intentional/no_action` | Projections tests | Keep separate test-only posture and do not mirror it into production allowlists. |
| Generated contracts are downstream of owner inputs. | Generated artifact policy and generated consumers | Manual edits would drift or be overwritten. | `intentional/no_action` | Contract owners/generators | Change owner inputs and use Make generation only after authorization. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority posture, sole-write scope, and history. | Tracker only in this session | `make lint-markdown` | Scope and tracker-local normative effect are explicit. |
| WF-01 | Target inventory | chain | WF-00 | WF-02 | Inventory all 29 files, callers, SQL, tests, contracts, and generated inputs/outputs. | Entire target and discovered graph | Repository coverage audit | Every target file has one row; external seams are signed into mappings. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04, WF-AUTH | Freeze observable and internal consumer contracts. | Core owners; Workbook; Collaboration; Recovery; Reporting; source owners | Contract/test map review | Every contract has one current owner, evidence posture, and preservation rule. |
| WF-03 | Characterization gap analysis | parallel | WF-02 | WF-AUTH, WF-05 | Define deletion, restore, host/identity parity, and test-util matrices without asserting unobserved behavior. | Projection, owner, Workbook, Recovery, test-util tests | Matrix-to-test audit | Every matrix cell maps to an existing or explicitly required test. |
| WF-04 | Boundary and coupling scan | parallel | WF-02 | WF-AUTH, WF-05 | Close import, SQL, Reporting, query-helper, test-util, and policy inventories. | Target, assembly, source/Reporting providers, authored/generated SQL, policy inputs | Import/SQL inventories | All ten tables and every production access site are accounted for. |
| WF-AUTH | Adopt authority prerequisites | chain | WF-02, WF-03, WF-04 | WF-05 | Adopt Core 01 clarifications, Core 04 acceptance updates, and an internal architecture decision before implementation. | Core 01, Core 04, adopted internal architecture decision | Human/adoption evidence; no tracker claim substitutes | All three adopted artifacts exist and agree; otherwise implementation remains blocked. |
| WF-05 | Authorization A and characterization baseline | chain | WF-AUTH | WF-06 | Under separate Authorization A, add/run only S-00 characterization evidence. | Tests only; no structural production change | Narrow owner and service-backed slices | Every PRF-040 through PRF-043 matrix passes on pre-refactor behavior. |
| WF-06 | Facade, caller, and provider slice sequence | chain | WF-05 and separate Authorization B | WF-07 | Build adapter/contracts, migrate callers, then move SQL one provider at a time. | Projections, assembly, Workbook, Recovery, Reporting, eight source owners | Corresponding narrow owner slice immediately after each migration | Each slice has evidence and an independently reversible boundary. |
| WF-07 | Test-util, policy, manifest, and harness reconciliation | chain | WF-06 | WF-08 | Migrate test capability; update code-backed registry first, authored manifest/policies next, generated outputs only through tools. | Testutil, registry, authored validation and harness inputs | Boundary, shape, drift, owner slices | Four-way parity and separate prod/test permissions pass. |
| WF-08 | Root removal, validation, and final handoff | chain | WF-07 | none | Remove compatibility/root permissions and run narrow-to-broad final verification. | Target and verification evidence only | Section 8 sequence | Empty root imports, no root API, all DoD rows pass, evidence roots recorded. |

### Required authority artifacts

| Artifact | Minimum required normative content | Gate |
| --- | --- | --- |
| Core 01 clarification | Caller-owned transaction writer rule; all ten projection-table access enforcement; current-profile restore outcomes, including rollback and publication claims | MUST be adopted before Authorization A. |
| Core 04 acceptance update | Differential query parity; restore outcome matrix; package-import policy; typed deletion and rollback evidence | MUST be adopted before Authorization A. |
| Internal architecture decision | Exact package paths, interfaces in section 4, SQL placement, Reporting seam, transition order, compatibility removal, and rollback rules | MUST be adopted before Authorization A. |
| Authorization A | Tests-only S-00 scope; production implementation forbidden | MUST be separate and complete before Authorization B. |
| Authorization B | Named structural slices after the clean characterization baseline | MUST NOT be inferred from Authorization A or this tracker. |

## 7. Proposed Refactor Slice Plan

No slice below is authorized by this tracker. PRF-060 requires two distinct
authorizations: Authorization A permits S-00 tests only; Authorization B may
name structural slices only after S-00 is clean. Observable behavior MUST be
preserved. A behavior correction MUST be a separate slice marked `requires
later authorization` and MUST NOT be folded into structural movement.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| A-00 | none | Adopt the three authority artifacts in WF-AUTH and issue separate Authorization A. **Requires later authorization.** | Core 01, Core 04, internal architecture decision, authorization record | A tracker could be mistaken for adopted authority. | Human owner review plus contradiction audit | No product command proves adoption. | Withdraw Authorization A if artifacts conflict or omit a required interface/outcome. | Adopted artifacts exist, agree, and Authorization A names only S-00. |
| S-00 | A-00 | Add and run only missing characterization for PRF-040 through PRF-043. Structural production edits are forbidden under Authorization A. | Projection/owner/Workbook/Recovery/testutil tests | Tests could accidentally specify desired rather than current behavior. | All matrices below and the eight current projection rows | Projection, Workbook, Recovery, Entities, Reporting, and source-owner slices from section 8 | Revert new tests independently; do not “fix” a failed characterization. | Every matrix cell passes against unchanged implementation and evidence roots are retained. |
| S-01 | S-00 and Authorization B | Add `providercontract` descriptor facts, source facade contracts, `adapters.New`, private runtime/storage/queryengine packages, and compatibility delegation. **Requires later authorization.** | Projections target and `internal/app/projectionassembly` | Nil semantics, cycles, descriptor v3, hidden concrete leakage | Constructor failure matrix; registry; query; restore; boundary tests | `make test-slice OWNER=module.projections`; `make backend-module-boundary-check` | Remove the unconsumed adapter/contracts as one reversible change. | Exact section 3/4 APIs compile; fail-closed tests pass; no consumer migrated yet. |
| S-02 | S-01 | Migrate projection assembly, Workbook query, and Recovery ports, one consumer at a time. | Projection assembly, Workbook, Recovery | HTTP/cursor/auth, descriptor snapshot, restore outcomes | Workbook query/auth/cursor suite; Recovery restore matrix | Run the corresponding Workbook or Recovery owner slice immediately after its migration. | Revert one consumer; compatibility delegation remains. | Workbook and Recovery import no root/private implementation and all frozen contracts pass. |
| S-03 | S-01 | Migrate the eight source-owner facades and their assembly/mutation callers, one owner at a time. | Timeline, Entities, Indicators, Assessments, Artifacts, Evidence, Parties, Tasks/Decisions | Transaction atomicity, typed deletion, load rows, affected views | Source mutation/import/merge tests; deletion matrix; Collaboration tests | Run the migrated source module's narrow owner slice immediately. | Revert only that owner; do not combine owner migrations. | Every source owner imports only its facade/providercontract seam and retains semantics. |
| S-04 | S-02, S-03 | Move projection-table access provider by provider in rebuild order, including three Reporting paths; replace `links/readshape`; reconcile authored/generated query inputs. **Requires later authorization.** | Projections storage/queryengine; source/query/Reporting providers; authored SQL inputs | Row shape/order, reporting facts, stale rows, query bounds, SQL ownership | Per-provider refresh/query/rebuild parity; Reporting provider tests; SQL policy tests | Projections service-backed slice plus the corresponding Entities/Reporting/source-owner slice after each provider | Revert one provider's SQL and adapter delegation; never move multiple providers without an intervening green boundary. | Signed inventory has no runtime projection-table access outside Projections; all ten provider checks pass. |
| S-05 | S-02, S-03 | Replace Timeline-bundle-backed testutil projection capability with `projections/testsupport` and a projection-assembly-backed test port. | `internal/testutil/httptestx`, appsupport, test assembly | Fixture behavior could diverge from production ports. | Every row in the testutil caller map below | Relevant service-backed owner slices | Revert test-only composition independently. | All mapped callers use the replacement; no Timeline bundle exposure remains; no production permission was added. |
| S-06 | S-04, S-05 | Update the code-backed registry first, then authored validation manifest v4, ten table rules, production/test import policies, and harness mappings only where evidence requires. | Registry/policy; `contracts/projection-providers`; authored boundary/test-family inputs | Manifest drift, false accounting, generated hand edit | Registry/manifest parity, four-way equality, boundary and catalog tests | Drift/shape/boundary commands in section 8 | Revert authored inputs and tool-generated outputs together. | Descriptor, manifest, recovery state, schema ownership, import policy, and harness accounting agree. |
| S-07 | S-06 | Remove compatibility imports, legacy root services/types, and root Go package surface. | Projections root, boundary policy, final callers | Hidden importers, stale type assertions | Full production/test import scans and all owner slices | `make backend-module-boundary-check`; `make test-fast` | Restore compatibility layer if any named consumer/evidence remains. | Root importer set is empty and root exports no production API. |
| S-08 | S-07 | Execute final narrow-to-broad verification and record retained evidence. | Verification artifacts only | False success from partial or stale evidence | All frozen contracts and matrices | Section 8 exact order; `make agent-finalize`; `make check` when risk warrants | Revert to the last independently green slice; do not repair across slice boundaries. | Every implementation DoD row passes with command, result, and evidence root. |

### Deletion characterization matrix

PRF-040 requires the pre-refactor baseline and final implementation to pass
every cell for host, identity, and indicator deletion, plus Timeline deletion
through the existing mutation contract.

| Scenario | Required assertion |
| --- | --- |
| Correct incident and supported type | Exactly the intended derived row is absent after commit. |
| Record belongs to a different incident | No projection row is removed and the operation returns the current contract result. |
| Transaction rollback | The projection row remains byte-for-byte behaviorally equivalent; no success evidence is emitted. |
| Row already absent | Operation is idempotent under the current error/result contract. |
| Source and history | Authoritative source, revision, and history rows remain unchanged. |
| Publication | Deletion code publishes no WebSocket event; source coordinator retains publication ownership. |
| Telemetry | Span name/result vocabulary remains bounded and contains no SQL, table, row, actor, or source values. |
| Unsupported provider/type | No generic fallback runs; no typed operation exists. |

### Restore characterization matrix

PRF-041 sets transactional all-or-nothing restore rebuild as the default. A
failure before commit MUST leave no partial projection state and MUST NOT claim a
rolled-back provider as rebuilt. See the [PostgreSQL transaction
model](https://www.postgresql.org/docs/current/tutorial-transactions.html).

| Scenario | Required assertion |
| --- | --- |
| No active providers | Deterministic empty result/readiness under the adopted current-profile contract; no tables touched. |
| Nonparticipating provider | Excluded exactly according to its immutable descriptor. |
| Unsupported required provider/version | Fail before table mutation and name the unsupported requirement without leaking SQL. |
| First provider failure | Roll back all projection changes; no rebuilt-resource success. |
| Middle provider failure | Roll back earlier providers and do not invoke later providers. |
| Final provider failure | Roll back all earlier providers; no partial success claim. |
| Commit failure | Return failure and no successful rebuilt-resource claim. |
| Context cancellation | Stop deterministically, roll back, and preserve cancellation identity. |
| Retry after failure | Same source state can succeed without cleanup beyond the normal transaction/rebuild contract. |
| Stale projection rows | Successful rebuild removes stale rows and replaces the provider set exactly. |
| Invalid source reference | Apply the current provider error contract and roll back the entire rebuild. |
| Deterministic order | Invocation and result order equal the ten-provider chain for every run. |

### Host/identity differential parity matrix

PRF-042 requires the pre-move specialized path, migration adapter, and final
generic/private path to be tested against the same fixtures. PostgreSQL does not
guarantee unordered result order, so every path MUST issue explicit ordering;
see [`ORDER BY`](https://www.postgresql.org/docs/current/queries-order.html).

| Dimension | Required equality |
| --- | --- |
| Rows and cells | Same `view_row_v1` rows, scalar cells, row versions, nulls, and omitted/present fields. |
| Collections | Same member values, order, and empty/null representation. |
| Grouping | Same normalized grouping keys and row membership. |
| Paging | Same first/continuation pages, cursor-bound tuple semantics, `has_more`, and no duplicates/omissions. |
| Null order | Same required null placement for ascending and descending sorts. |
| Tie-breaking | Same final `record_id` tie-breaker and deterministic results. |
| Filters | Same operators, values, tag/read-shape behavior, and error posture. |
| Mutation | Same row immediately after a committed source mutation. |
| Merge/delete | Same removal/refresh results and affected-view consequences. |
| Load row | Same canonical `map[string]any` values for returned IDs. |
| Authorization continuation | Workbook reauthorizes every continuation; Projections receives no auth bypass. |
| Bounds | SQL fetches at most `limit+1`; hydration uses only returned IDs; no deep `OFFSET`. |

### Test-util capability migration map

PRF-043 requires one row per discovered caller. The replacement MUST implement
the production-equivalent consumer interface unless the row records a narrower
test-only exception. A `testsupport` permission MUST NOT permit production SQL.

| Discovered caller | Current Timeline-bundle capability | Behavioral owner | Replacement projection test port | Production-equivalent interface | Exception | Removal condition |
| --- | --- | --- | --- | --- | --- | --- |
| Indicators resolution integration | `RebuildIndicators` | Indicators/Projections | typed indicator rebuild | indicator writer/rebuild port | Test fixture may request direct rebuild | Old bundle field has zero callers. |
| Timeline auto-resolution integration | `RebuildTimeline` | Timeline/Projections | typed Timeline rebuild | Timeline writer/rebuild port | Test fixture direct rebuild | Same. |
| Entities resolution integration | `RebuildHosts`; host+identity rebuild | Entities/Projections | typed host and identity rebuild | Entities writer/rebuild port | Test fixture direct rebuild | Same. |
| Entities support suite | Timeline, host, identity, indicator rebuilds | Named source owners/Projections | explicit aggregate of four typed test capabilities | Four production-equivalent ports | Aggregate allowed only in test composition | No Timeline-owned aggregate remains. |
| Timeline projection-query deterministic rebuild integration | `RebuildTimeline` | Timeline/Projections | typed Timeline rebuild | Timeline writer/rebuild port | Test fixture direct rebuild | Old bundle field removed. |
| Evidence attached-projection integration | Timeline rebuild; host+identity rebuild; Evidence port | Evidence plus named source owners/Projections | explicit typed capabilities | Evidence, Timeline, and Entities ports | Test aggregate scoped to this fixture | Every capability names its owner. |
| Revisions indicator-child rollback integration | Indicator projection port plus source-text port | Revisions/Indicators | typed indicator writer plus unchanged source-text port | Indicator writer | Source-text port is not a Projections capability | No generic projection coordinator remains. |

No SQL schema migration, HTTP/WS contract change, frontend selector change, or
generated-contract shape change is proposed. If evidence requires one, it MUST
become a separate behavior-change plan with explicit authorization.

## 8. Validation Plan

PRF-070 requires the narrowest changed-owner command immediately after each
migration. A later implementation MUST stop on the first failure, retain its run
root, and roll back to the last independently green slice before broadening.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit/static | `make test-slice OWNER=module.projections ROWS=module.projections.architecture.boundaries_and_source_ownership_9b8de61f81,module.projections.provider.catalog_validation_684d752cb4,module.projections.query.contract_shape_and_keyset_59528aa56d,module.projections.telemetry.safe_boundary_8e0e774b19` | Boundary, provider validation, query shape, telemetry | yes | Establish baseline and repeat after relevant slices. |
| integration | `make service-backed-test-slice OWNER=module.projections ROWS=module.projections.assembly.catalog_and_source_ownership_927c3f08ce,module.projections.query.timeline_keyset_bounds_e67384caff,module.projections.rebuild.restore_contract_206379110d,module.projections.storage.query_row_parity_157462a15a` | Catalog/source SQL, Timeline paging, restore, row parity | yes | All current projection rows are service-backed according to owner explanation. |
| Workbook owner | `make test-slice OWNER=module.workbook`; `make service-backed-test-slice OWNER=module.workbook` | Route, envelopes, view row, cursor, saved-view compatibility, authorization | yes, at baseline and immediately after Workbook migration | Use explained row selectors when the owner guide provides a narrower complete set. |
| Recovery owner | `make test-slice OWNER=module.recovery`; `make service-backed-test-slice OWNER=module.recovery` | Restore request/result/readiness and retry/failure outcomes | yes, at baseline and immediately after Recovery migration | PRF-041 matrix must map to retained evidence. |
| Entities owner | `make test-slice OWNER=module.entities`; `make service-backed-test-slice OWNER=module.entities` | Host/identity mutation, merge/delete, hydration, paging parity | yes, at baseline and after each host/identity change | PRF-040 and PRF-042 apply. |
| Reporting owner | `make test-slice OWNER=module.reporting`; `make service-backed-test-slice OWNER=module.reporting` | Projection-backed artifact, host/identity, task/decision field providers | yes, before and after each Reporting SQL migration | Run the matching source-owner slice in the same checkpoint. |
| Source-owner modules | `make test-slice OWNER=module.timeline`; `make test-slice OWNER=module.indicators`; `make test-slice OWNER=module.assessments`; `make test-slice OWNER=module.artifacts`; `make test-slice OWNER=module.evidence`; `make test-slice OWNER=module.parties`; `make test-slice OWNER=module.tasksdecisions` | Source semantics, typed writers, mutation/import/rollback behavior | yes, for the owner being migrated | Use the corresponding `make service-backed-test-slice OWNER=...` whenever its explained rows require services. |
| e2e/browser | `make browser-e2e` | Workbook-visible rows, selectors, and Collaboration behavior | no | Required after a slice only if route, WebSocket, selector, or frontend-visible behavior is touched. |
| generated drift | `make generate-drift` | Generated Go/TypeScript projections from owner inputs | no | Required after any authorized owner-input/codegen change. |
| generated policy | `make generated-artifact-policy-check` | Generated-root write policy | no | Confirms no manual generated-root edits. |
| JSON/contract shape | `make json-shape-check` | Projection provider manifest and JSON owners | yes, before descriptor/manifest work | Manifest is authored validation evidence, not generated runtime authority. |
| import-boundary/static | `make backend-module-boundary-check` | Backend import and SQL table access policies | yes | Also run the projection architecture row because it owns additional import policy. |
| fast broad check | `make test-fast` | Broad non-service regression | no | Run after the final caller migration, not after documentation-only work. |
| evidence finalization | `make agent-finalize` | End-of-run evidence maintenance | no | Supply `RESULTS_DIR` only when retaining a successful run root; otherwise record it as unset. |
| full check | `make check` | Full repository verification | no | Run after implementation when the combined boundary/contract risk warrants it. |
| documentation | `make lint-markdown` | This tracker and repository Markdown | yes, for this session | Passed for the final revision; run root `.cartulary/test-results/20260808T030403Z-p2181649`. |

The discovery targets `make task-guide ROLE=module-author OWNER=module.projections`,
`make explain-test-owner OWNER=module.projections`, `make help`, and relevant
`make explain-target TARGET=... DETAIL=summary` calls completed successfully.
They discover routing and do not constitute product validation. A later
implementation MUST use this order: changed-owner unit/static, changed-owner
service-backed, Projections unit/static, Projections service-backed, boundary,
shape/drift/policy when owner inputs changed, `make test-fast`,
`make agent-finalize`, and `make check` when aggregate risk warrants it. Browser
E2E is conditional on route, WebSocket, selector, or frontend-visible behavior.
No product test, integration suite, browser suite, drift check, or full check was
run for this documentation revision. `make lint-markdown` passed at
`.cartulary/test-results/20260808T030403Z-p2181649`.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| PRT-001 | Inventory all 29 projection target files and current responsibilities. | WF-01 | DONE | WF-00 | Section 2 | Every target file has an evidence-backed row. |
| PRT-002 | Map provider, query, refresh, rebuild, restore, Reporting, and telemetry ownership. | WF-02 | DONE | PRT-001 | Sections 3 through 5 | Every discovered responsibility has a current owner, target owner, and preservation rule. |
| PRT-003 | Specify tracker-local `PRF-*` requirements and exact target interfaces. | WF-02 | DONE | PRT-001 | PRF-001 through PRF-070 in this tracker | Every requirement is testable and appears in section 12 traceability. |
| PRT-004 | Complete eight-facade, ten-provider/table, root-surface, SQL, Reporting, and testutil mappings. | WF-03, WF-04 | DONE | PRT-002 | Sections 3, 5, and 7 | Each declared set is exhaustive for current repository evidence. |
| PRT-005 | Adopt Core 01 clarification. | WF-AUTH | BLOCKED | PRT-003, PRT-004 | RB-001, RB-002 | Adopted owner text covers transaction writers, ten-table access, and restore outcomes. |
| PRT-006 | Adopt Core 04 acceptance updates. | WF-AUTH | BLOCKED | PRT-003, PRT-004 | RB-003, RB-004 | Adopted acceptance criteria cover parity, restore, imports, deletion, and rollback. |
| PRT-007 | Adopt internal architecture decision. | WF-AUTH | BLOCKED | PRT-003, PRT-004 | RB-001, RB-002 | Adopted decision names exact packages, interfaces, SQL placement, transition, and rollback. |
| PRT-008 | Issue Authorization A and complete S-00 characterization. | WF-05 | BLOCKED | PRT-005 through PRT-007 | RB-003, RB-005 | Separate authorization exists and every characterization matrix passes on unchanged production code. |
| PRT-009 | Issue Authorization B and implement adapters/contracts and caller migrations. | WF-06 | DEFERRED | PRT-008 | S-01 through S-03; RB-005 | Named slices pass immediate changed-owner evidence without behavior change. |
| PRT-010 | Move all ten providers' projection SQL, Reporting SQL, and neutralize peer query coupling. | WF-06 | DEFERRED | PRT-009 | S-04; RB-002 | Signed SQL inventory is closed and every provider checkpoint passes. |
| PRT-011 | Migrate testutil and reconcile registry, manifest, policies, and harness accounting. | WF-07 | DEFERRED | PRT-009, PRT-010 | S-05, S-06; RB-004 | Testutil map and four-way equality pass; only authorized authored inputs changed. |
| PRT-012 | Remove root compatibility and run final implementation verification. | WF-08 | DEFERRED | PRT-011 | S-07, S-08 | Empty root imports/API and every implementation DoD row pass with retained evidence. |
| PRT-013 | Validate this tracker as the sole documentation revision. | WF-00 | DONE | PRT-001 through PRT-004 | `make lint-markdown`; 29-path/12-heading coverage; worktree audit | Current revision passes Markdown lint and the sole-change audits. |

`DONE` in this table means that the named planning artifact or executed evidence
exists. It MUST NOT be interpreted as authority adoption, characterization
success, or implementation completion. No implementation row is `DONE`.

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Target exists; tracker was absent; planning-only scope enforced. | Inspected framework, Core 00-05, Graph Projection NLSpec, domain, Appendix I; touched only this tracker. | `sed`, `rg`, `git status`, `make lint-markdown` | Authority and scope recorded; Markdown lint passed; no owner contradiction found. | Production refactor is not authorized. | Hand off S-00 to a later authorized task. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Target architecture is specified in tracker-local normative language; adoption and implementation remain unauthorized. | Inspected `temp/analysis-notes.md`, NLSpec writing guide, adopted owners, exact live interfaces, and prior tracker; revised only this tracker. | `sed`, `rg`, `jq`, `wc`, `git status`, `date`, `make lint-markdown` | Added PRF requirements, explicit defaults, authority gates, traceability, and binary acceptance; final Markdown lint passed at `.cartulary/test-results/20260808T030403Z-p2181649`. | RB-001 through RB-005. | Adopt WF-AUTH artifacts before Authorization A. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Legitimate projection subsystem with mixed root facade and boundary nonconformance. | Inspected all 29 target files, assembly, callers, source-owner providers, testutil; touched only this tracker. | `rg`, `sed`, `find`, `wc`, `jq` | Root file allowlist, distributed SQL, peer helper, and testutil seams classified. | RB-001 and RB-002 require later owner-approved design. | Establish characterization, then approve package-level adapters and storage placement. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Sole adapter, immutable contracts, private implementation, empty root, and eight source facades are specified. | Re-inspected root exports, source ports, app assembly, provider registry, SQL sources, and Go package layout guidance; revised only this tracker. | `rg`, `sed`, `jq`, `make lint-markdown` | Exact target topology, root-surface disposition, constructor failures, and typed writer rules recorded. | RB-001 adoption and RB-002 signed SQL inventory remain pending. | Adopt the architecture decision, then issue Authorization A for S-00 only. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Not an implementation target; no frontend or grid-vendor code/import exists under Projections. | Inspected generated TypeScript/view consumers, Workbook browser references, and target imports; touched only this tracker. | `rg`, `jq` | Frontend remains an observable contract consumer only. | None unless a later slice changes rows, WS events, or selectors. | Run browser validation conditionally after any frontend-visible backend change. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Frontend remains outside the target; all frontend-visible behavior is frozen by default. | Re-used grounded view/protocol/selector evidence; revised only this tracker. | `rg`, `make lint-markdown` | No frontend shell or grid-vendor responsibility was assigned to Projections. | Browser evidence is conditional on a later frontend-visible change. | Keep frontend unchanged; run `make browser-e2e` only when section 8 conditions apply. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Descriptor v3 code registry is authoritative; manifest v4 is authored validation evidence; generated roots are downstream. | Inspected projection-provider manifest/schema, view schemas, OpenAPI, WS schema, migrations, recovery fixtures, generated Go/TS consumers; touched only this tracker. | `rg`, `sed`, `jq` | Contract freeze and no-hand-edit posture recorded. | Any descriptor/API shape change needs later authorization. | Update registry first and use Make-owned drift/generation paths only if implementation requires it. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Descriptor v3/manifest v4 remain the explicit default; target descriptor facts move to `providercontract` without a version change. | Re-inspected descriptor, Workbook query, Recovery, Reporting, view-schema, querypage, authored/generated query, migration, and generated-artifact surfaces; revised only this tracker. | `rg`, `sed`, `jq`, `make lint-markdown` | Exact interface types, three Reporting provider seams, generator ownership, and version-change gate recorded. | Neutral read-shape inability would require separate descriptor-version authorization. | Preserve v3/v4; update registry first and generated outputs only through Make. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Eight active projection rows discovered; no product validation run. | Inspected target/owner tests, module test family, verification owner, test catalog, and boundary owner inputs; touched only this tracker. | `make help`; `make task-guide ROLE=module-author OWNER=module.projections`; `make explain-test-owner OWNER=module.projections`; `make explain-target TARGET=... DETAIL=summary`; `make lint-markdown` | Canonical commands recorded; SQL table-policy gap identified; Markdown lint passed. | RB-003 and RB-004. | Run S-00 baselines in a later implementation task. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Four characterization matrices and changed-owner validation checkpoints are specified; product tests remain unrun. | Inspected discovered testutil callers, projection/owner test surfaces, and verification owner IDs; revised only this tracker. | `rg`, `jq`, `make lint-markdown` | Deletion, restore, host/identity parity, and testutil mappings now have testable acceptance; docs lint passed. | RB-003 needs Authorization A and executable evidence; RB-004 needs authorized policy work. | Under Authorization A, add/run S-00 only and retain run roots. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Workbook retains route admission and incident-membership authorization; Projections has no direct HTTP auth responsibility. | Inspected Core 04, Workbook route and cursor authorization tests, Collaboration socket/replay tests; touched only this tracker. | `rg`, `sed` | Authorization and transport ownership frozen as intentional/no-action. | None for planning. | Rerun existing authorization and Collaboration tests after any facade migration. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Authorization remains Workbook/Incident-owned; continuation reauthorization and no-bypass behavior are frozen. | Re-used grounded Core 04, Workbook, and Collaboration evidence; revised only this tracker. | `rg`, `make lint-markdown` | PRF-003 and parity acceptance explicitly prohibit moving authorization or publication into Projections. | Core 04 acceptance update is pending under WF-AUTH. | Add acceptance evidence before Authorization A; rerun owner tests after Workbook migration. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 22:06 EDT | Codex planning/documentation session | Tracker is complete; implementation remains deliberately blocked/deferred. | Inspected repository and evidence listed above; touched only this tracker. | Discovery commands above; `make lint-markdown`; coverage and diff audits | Workstreams, slices, rollback notes, and stable blockers recorded; documentation checks passed. | RB-001 through RB-005. | Begin only with separately authorized S-00 characterization. |
| 2026-08-07 22:55 EDT | Codex NLSpec tracker revision | Revised tracker is documentation-complete; all product DoD criteria remain pending. | Inspected notes, guide, owners, live interfaces, all mappings, and prior handoff history; revised only this tracker. | `sed`, `rg`, `jq`, `wc`, `git status`, `make lint-markdown` | PRF traceability, authorization gates, rollback boundaries, and DOD-01 through DOD-11 recorded; docs lint passed. | RB-001 through RB-005; no owner contradiction. | Adopt authority artifacts; do not begin product changes from this tracker alone. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Adopt the exact package topology, constructor, consumer-owned ports, import policy, and transition rules in sections 3 and 4. | The decision is specified here but tracker text cannot amend Core or create an adopted architecture boundary. | Core 01 clarification, adopted internal architecture decision, and implementation evidence. | BLOCKED: decision specified; adoption and implementation evidence pending. |
| RB-002 | Sign and implement the exhaustive projection SQL inventory, including source-owner projection/query providers, projection-backed Reporting providers, and authored/generated query inputs. | Every runtime projection-table access must execute under Projections without stealing source/reporting semantic ownership. | Provider-by-provider signed inventory, source/Reporting owner review, adopted SQL placement, and implementation evidence. | BLOCKED: decision specified; signed SQL inventory and implementation pending. |
| RB-003 | Pass every deletion, restore, host/identity differential-parity, and testutil capability characterization row. | Structural movement is unsafe until failure, rollback, ordering, hydration, publication, and fixture behavior are executable. | Separate Authorization A and retained S-00 test evidence on unchanged production code. | BLOCKED: open until all characterization matrices pass. |
| RB-004 | Implement ten exact table-access rules and prove four-way equality. | Four current policy rules and broader assembly scanning are incomplete and can disagree. | Adopted boundary decision; authored policy work; registry/recovery/schema equality tests. | BLOCKED: decision specified; ten-rule policy work pending. |
| RB-005 | Issue separate Authorization A and Authorization B. | This tracker is planning evidence only; tests-only characterization must not imply structural implementation permission. | Authorization A naming S-00 only; after it passes, Authorization B naming accepted structural slices. | BLOCKED: no production or test implementation until both gates exist in sequence. |

There is no `BLOCKED: owner contradiction` entry because no adopted-owner
contradiction was discovered. Repository/Core and repository/Appendix mismatches
are captured in sections 1 and 5.

## 12. Binary Completion Criteria

### Requirement-to-evidence traceability

The “binary criterion” column names the implementation Definition of Done row
that MUST pass. A planned command is not evidence until its result and run root
are recorded.

| Requirement | Adopted owner scope | Workflow | Slice | Required test scenario | Make command | Binary criterion |
| --- | --- | --- | --- | --- | --- | --- |
| PRF-001 | Core 00-04, affected subsystem owner | WF-AUTH | A-00, any behavior-change slice | Contract before/after differential; separate authorized correction | Affected owner slice | DOD-07, DOD-09 |
| PRF-002 | Core 00 authority posture | WF-AUTH | A-00, S-00 | Authorization records predate permitted changes | Human adoption audit; no product command | DOD-08 |
| PRF-003 | Core 01-04 and named subsystem owners | WF-02, WF-08 | S-00, S-08 | HTTP/WS/cursor/auth/saved-view/schema/telemetry/source behavior freeze | Workbook, Recovery, source-owner, Projections slices | DOD-07 |
| PRF-004 | Repository authority/harness owner | WF-07 | S-06 | Runtime and evidence dependency scan excludes Markdown | `make backend-module-boundary-check`; `make check` | DOD-10 |
| PRF-005 | Migration and generated-artifact owners | WF-07 | S-06, S-08 | Migration diff empty; generated outputs tool-derived only | `make generate-drift`; `make generated-artifact-policy-check` | DOD-10 |
| PRF-010 | Core 01 REQ-01-626 | WF-06 | S-01, S-07 | Sole adapter constructor and empty root import set | `make backend-module-boundary-check` | DOD-01 |
| PRF-011 | Core 01 descriptor ownership | WF-06 | S-01 | Immutable contract API; executable types cannot cross | `make test-slice OWNER=module.projections` | DOD-02 |
| PRF-012 | Testing Harness owner | WF-07 | S-05, S-06 | Test permission does not grant production import/SQL | `make backend-module-boundary-check` | DOD-01, DOD-06 |
| PRF-013 | Core 01 package boundary | WF-06 | S-01, S-07 | External compile/import attempts against nested private packages fail | `make backend-module-boundary-check` | DOD-01 |
| PRF-014 | Core 01 REQ-01-626 | WF-08 | S-07 | Root has no exported production API/importers | `make backend-module-boundary-check`; `make test-fast` | DOD-01 |
| PRF-015 | Core 01 source semantics | WF-06 | S-03 | Eight facades expose only approved immutable/typed seams | Source-owner slices; boundary check | DOD-01, DOD-03 |
| PRF-020 | Workbook, Recovery, Reporting consumer contracts | WF-06 | S-01, S-02 | Exact port type assertions and no concrete leakage | Projections, Workbook, Recovery, Reporting slices | DOD-02 |
| PRF-021 | Core 01 fail-closed composition | WF-06 | S-01 | Nil DB/source; duplicate/missing/version/query/owner/restore failures; all-success ports non-nil | `make test-slice OWNER=module.projections` | DOD-02 |
| PRF-022 | Source-owner transaction contracts | WF-05, WF-06 | S-00, S-03 | Typed writer, caller transaction, typed deletion, rollback | Source-owner service-backed slices | DOD-03, DOD-09 |
| PRF-023 | Reporting FieldProvider contract | WF-06 | S-01, S-04 | Three provider families preserve output facts while SQL moves | Reporting plus Artifacts/Entities/TasksDecisions slices | DOD-04, DOD-07 |
| PRF-030 | Core 01 storage-lifecycle ownership | WF-04, WF-06 | S-04 | Signed production SQL scan and ten provider parity checkpoints | Projections/Reporting/source slices; boundary check | DOD-04 |
| PRF-031 | Core 01 boundary enforcement | WF-07 | S-06 | Ten exact rules and four-way set equality, including extras/duplicates | `make backend-module-boundary-check`; `make json-shape-check` | DOD-05 |
| PRF-032 | Workbook/query paging contract | WF-05, WF-06 | S-00, S-04 | Host/identity differential parity and bounded exact-ID hydration | Projections and Entities service-backed slices | DOD-06, DOD-07 |
| PRF-033 | Core 01 modular boundary | WF-06 | S-04 | No `links/readshape` import; v3 neutral contract parity or authorized version gate | Projections query row; boundary and shape checks | DOD-01, DOD-07 |
| PRF-040 | Core 04 acceptance; source mutation owners | WF-05 | S-00, S-03 | Every deletion matrix row | Projections and affected source-owner service-backed slices | DOD-06, DOD-09 |
| PRF-041 | Core 01 restore; Core 04 acceptance | WF-05 | S-00, S-02, S-04 | Every restore matrix row, including commit/cancel/retry/order | Projections and Recovery service-backed slices | DOD-06, DOD-09 |
| PRF-042 | Core 01 query; Core 04 acceptance | WF-05 | S-00, S-04 | Every host/identity parity dimension | Projections, Workbook, Entities service-backed slices | DOD-06, DOD-07 |
| PRF-043 | Testing Harness and source owners | WF-05, WF-07 | S-00, S-05 | Every discovered testutil caller and removal condition | Relevant service-backed source-owner slices | DOD-06 |
| PRF-060 | Core 00 authority posture | WF-AUTH | A-00, S-00 | Separate Authorization A/B evidence and ordering | Human authorization audit | DOD-08 |
| PRF-070 | Testing Harness execution owner | WF-08 | S-01 through S-08 | Immediate changed-owner evidence and retained final sequence | Section 8 commands | DOD-11 |

### Tracker acceptance

| Criterion | Result | Evidence |
| --- | --- | --- |
| All 29 current target files are inventoried exactly once. | PASS | Section 2 and current path-coverage audit. |
| All 12 required top-level sections exist exactly once. | PASS | Current heading audit. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 contract-freeze map. |
| Eight source facades, ten providers/tables, current root surfaces, SQL categories, and testutil callers are exhaustively mapped. | PASS | Sections 3, 5, and 7. |
| Every workflow and slice has dependencies, validation, rollback, and an exit condition. | PASS | Sections 6 and 7. |
| Every `PRF-*` requirement maps to owner, workflow, slice, test, command, and binary criterion. | PASS | Requirement-to-evidence table above. |
| Unknown implementation evidence is blocked rather than guessed. | PASS | RB-001 through RB-005 and no implementation row marked `DONE`. |
| No owner contradiction was found; evidence/repository mismatches remain separate. | PASS | Sections 1, 5, and 11. |
| Existing handoff history is preserved and the revision session is appended. | PASS | Section 10. |
| Current documentation revision passes Markdown lint and is the sole worktree change. | PASS | `make lint-markdown` run root `.cartulary/test-results/20260808T030403Z-p2181649`; final coverage/status audits. |

### Implementation Definition of Done

These criteria define completion of a later authorized refactor. Their current
result is `PENDING`; planning completeness MUST NOT be confused with product
completion.

| ID | Binary criterion | Pass condition | Current result |
| --- | --- | --- | --- |
| DOD-01 | Package imports | Root production importer set is empty; only projection assembly imports `adapters`; only approved assembly/source facades import `providercontract`; production/test permissions are distinct; private packages are compiler-contained. | PENDING |
| DOD-02 | Exact construction interfaces | `adapters.New` uses the exact section 4 types, passes every fail-closed case, returns all non-nil required consumer ports, and exposes no concrete runtime/storage/query details. | PENDING |
| DOD-03 | Typed mutation interfaces | No generic mutation/deletion dispatcher remains; only typed host, identity, indicator, and Timeline mutation deletion exists; every writer preserves caller-owned transaction semantics. | PENDING |
| DOD-04 | Ten-table SQL ownership | Every runtime access to all ten projection tables, including Reporting and specialized query paths, executes under Projections; the signed SQL inventory is closed. | PENDING |
| DOD-05 | Descriptor and policy parity | Exact equality holds among active descriptor table IDs, ten boundary rules, recovery-state IDs, and Projections schema objects; descriptor v3/manifest v4 parity passes. | PENDING |
| DOD-06 | Characterization matrices | Every deletion, restore, host/identity differential, and testutil caller row passes before and after its affected migration. | PENDING |
| DOD-07 | Contract preservation | HTTP, WebSocket, cursor, auth, saved-view, `view_row_v1`, telemetry, schema, migration, Reporting fact, and source-semantic evidence is unchanged. | PENDING |
| DOD-08 | Authorization IDs | Adopted Core/architecture artifacts plus distinct Authorization A and Authorization B exist in the required order and name the executed slices. | PENDING |
| DOD-09 | Rollback boundaries | Each provider/consumer migration has an independently green checkpoint; transaction failures and rollback emit no success claim; behavior changes are separately authorized. | PENDING |
| DOD-10 | Generated and migration policy | Existing migrations remain in place; generated files are tool-owned; no runtime or evidence path depends on Markdown. | PENDING |
| DOD-11 | Retained verification | Narrow changed-owner, service-backed, Projections, boundary, drift/shape, broad, and finalization results are recorded with run roots and justified skips. | PENDING |

The tracker is complete only when its acceptance rows pass. The refactor is
complete only when all DOD-01 through DOD-11 rows pass after authorization. Open
`RB-*` items intentionally block implementation, not this planning artifact.
