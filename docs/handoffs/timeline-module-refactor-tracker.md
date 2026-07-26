# timeline Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/modules/timeline`
- **Derived target label:** `timeline`
- **Output path:** `docs/handoffs/timeline-module-refactor-tracker.md`
- **Status:** Planning and documentation only.
- **Allowed change:** This tracker file only.
- **Non-goals:** No production refactor, test edit, contract edit, generated-file
  edit, package change, migration, harness change, route change, schema change, or
  observable behavior change is authorized by this task.
- **Later authorization:** Implementation, including characterization-test or
  verification-catalog changes, requires a later explicitly authorized task.

The label is the lowercase kebab-case basename of the target path. It contains no
spaces, separators, shell metacharacters, or unsafe filename characters.

### Source hierarchy

1. No adopted Timeline-specific subsystem NLSpec was found. An adopted NLSpec for
   another subsystem does not govern the Timeline workbook projection.
2. Core 00 through Core 04 and the active `module.timeline` machine requirement
   govern current implementation-conformance behavior.
3. Core 05 is relevant only to claim-bearing timed or fixture-sensitive
   publication. This tracker makes no such claim, so Core 05 is not invoked.
4. Domain vocabulary and implementation-support guides inform terminology,
   package-boundary direction, and verification mechanics.
5. Current repository code and tests establish current implementation state.
6. The planning framework and prior plans or handoffs are evidence only.

Active requirements may count as executable authority only through their owned
verification mapping. Planned requirements do not count as passing evidence. No
owner contradiction was found during this session. A later contradiction must be
recorded as `BLOCKED: owner contradiction` without choosing a side.

### Owner and guidance documents inspected

- `contracts/requirements/registry.json`
- `contracts/requirements/owners/module.timeline.json`
- `contracts/verification/owners/module.timeline.json`
- `tools/test_catalog_owner.json`
- `tools/test_families/module.timeline.json`
- `contracts/view-schemas/cartulary.view.timeline.v2.json`
- `tools/backend_module_boundaries.json`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/archive/Timeline-Module-Refactoring-Tracker-2.md`
- `temp/analysis-notes.md`

### Repository files inspected

Every file under `internal/modules/timeline` was inspected and is inventoried in
Section 2. Searches and exact-source reads also covered the inbound application
assembly, workbook, imports, incident-bundle, recovery, projections, entities,
mentions, merge, reporting, revisions, saved-view, collaboration, frontend, grid
contract, generated protocol, view-schema, and verification-catalog surfaces
identified by live imports or contract identifiers.

### Planning findings relative to prior guidance

- The framework describes a desired narrow Timeline owner for capture and mutation
  semantics, but the live 54-file target also contains transport, persistence,
  projection, workbook/import, cross-module effect, provider, and test-support
  responsibilities. The framework is therefore direction, not proof that the
  current package is already the correct permanent boundary.
- The archived June 2026 tracker records useful completed remediation, including a
  command façade and private store, but predates current subpackages such as
  `rowsnapshot`, `mentioneffects`, and `workbookprojection`. Its completion state
  is not current evidence and is not copied into this tracker.
- `temp/analysis-notes.md` supplies the closure guidance adopted in the
  `2026-07-25T23:18:54-04:00` session. It resolves planning intent but does not
  replace Core or machine-owner authority. Its external research is informative
  and was evaluated only through July 25, 2026; no source first published on July
  26, 2026 is claimed as verified.

### Closure posture

RB-001, RB-002, RB-004, and RB-005 are resolved as planning decisions. RB-003 is
`DECISION_RESOLVED; EVIDENCE_PENDING`: exact-selector accounting rules are
settled, but the required exhaustive disposition artifact does not yet exist.
These resolutions remove design ambiguity from the tracker. They do not authorize
production, test, catalog, contract, schema, migration, route, or generated-file
changes. The resolved decisions must be promoted into their owning Core,
appendix, and machine inputs in a later authorized prerequisite before production
implementation.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/timeline/.gitkeep` | Directory placeholder | None | None | None | None | None | Timeline repository layout | low | Explicitly out of runtime scope because it contains no behavior. |
| `internal/modules/timeline/api.go` | Timeline request decoding, hashes, collection-action validation, row payloads, and change-field calculation | `CreateRequest`, `PatchRequest`, action DTOs, decode functions, request hashes, payload builders | Workbook routes/stores, imports, tests | HTTP API errors, view-schema query types, conflict tokens, row presenter, field normalization | `decoder_test.go`, `request_test.go`, `support_test.go`, workbook tests | Timeline view schema; generated protocol and DB-facing types are consumed indirectly | Timeline contract plus workbook transport boundary | high | Large mixed-purpose file combining wire, validation, application, and presentation concerns. |
| `internal/modules/timeline/api_errors.go` | Converts stable mutation failures to HTTP API errors | `MutationAPIErrorContext`, `ClassifyMutationAPIError` | Workbook and Timeline route coordination | HTTP API error package; Timeline error types | `error_classifier_test.go`, route envelope tests | Public error-envelope contract | Timeline application adapter | medium | Error codes and authorization visibility are observable. |
| `internal/modules/timeline/auto_resolution.go` | Alias candidate lookup and exact/suppressor auto-resolution policy | Private helper surface | Timeline create, patch, and paste stores | Direct SQL, entities/aliases, links | `request_test.go`, `resolution_integration_test.go` | Mention and link contract | Timeline orchestration using Entities/Links ports | high | Interactive-only eligibility and provenance must remain exact. |
| `internal/modules/timeline/boundary_guard_test.go` | Enforces production import, façade, private-store, and error-classification boundaries | Test functions only | Test runner | Go AST/source inspection | This file | Backend module-boundary policy | Timeline test support | medium | All five executable test symbols are currently unselected across active family manifests; disposition is required under RB-003. |
| `internal/modules/timeline/bulk_mutation.go` | Decodes workbook bulk commands and currently compiles them into Timeline clipboard-paste requests | `BulkMutationRouteKey`, request/target DTOs, decoder, hash and `BulkMutationClipboardRequest` | Workbook routes/stores | HTTP API, Timeline clipboard contract, view schema | Workbook bulk and route tests | Workbook mutation route and Timeline view contract | Workbook admission plus Timeline command handlers | high | RB-001 resolves ownership; `fill_down_v1` and `multi_row_tag_assignment_v1` must stop using paste as a code-reuse disguise. |
| `internal/modules/timeline/clipboard_paste.go` | Parses TSV/CSV, validates Timeline headers, builds paste plans, and retains raw import columns | Paste request/plan/result DTOs, decoder, parser, plan builder, hash | Workbook clipboard routes/stores and tests | Shared `tabularingest.BatchPlan`, HTTP API, Timeline contract helpers | `clipboard_paste_test.go`, workbook clipboard tests | Clipboard route contract; Timeline source fields | Tabular Ingest planning plus Timeline target application | high | `BatchPlan` is the current unversioned migration input; `tabular_row_plan_v1` is planned, not current authority. |
| `internal/modules/timeline/clipboard_paste_store.go` | Transactional multi-row create/patch with mentions, tags, evidence, revisions, projection refresh, and idempotency | Private store methods through `Facade` | `facade.go`, workbook mutation flow | Direct SQL; entities, links, evidence, revisions, projections, idempotency | `clipboard_paste_test.go`, workbook clipboard integration tests, Timeline integration tests | Timeline mutation, projection, revision, and collection contracts | Timeline mutation coordinator | high | Atomic batch semantics and side-effect order are observable. |
| `internal/modules/timeline/clipboard_paste_test.go` | Characterizes parsing, mapping, raw capture, binding, and cross-incident rejection | Test functions only | Test runner | Clipboard API and plan helpers | This file | Clipboard and mention-binding contracts | Timeline/workbook test support | medium | The parsing test is selected; `TestClipboardPasteRejectsCrossIncidentRecordTarget` is currently unselected and requires RB-003 disposition. |
| `internal/modules/timeline/collection_descriptors.go` | Maps field keys to mention, record-reference, and tag collection policy | `CollectionFamily`, `CollectionPolicy`, `LookupCollectionPolicy` | Timeline mutation and collection stores | Timeline field vocabulary | Unit and integration collection tests | Timeline view-field and collection-action contract | Timeline | medium | Legitimate Timeline-owned semantic policy. |
| `internal/modules/timeline/decoder_test.go` | Characterizes patch validation and visible text fields | Test functions only | Test runner | `api.go` decoders and field normalization | This file | Timeline request and visible-source-text contract | Timeline test support | medium | Protects exact request validation. |
| `internal/modules/timeline/deleterestore/provider.go` | Supplies Timeline tables to generic delete/restore coordination | `NewProvider` | Timeline revision contribution | Records delete/restore provider API | Revision and recovery integration tests | Revision/delete-restore contract | Timeline source-owner adapter | medium | Keep narrow unless domain behavior appears. |
| `internal/modules/timeline/error_classifier_test.go` | Characterizes mutation error to HTTP envelope mapping | Test function only | Test runner | `api_errors.go` | This file | Public error-envelope contract | Timeline test support | medium | Required if transport/error code is split. |
| `internal/modules/timeline/facade.go` | Stable application command façade over the private store; bulk currently delegates to clipboard paste | `Facade`, command DTOs, `NewFacade` | Workbook, imports, recovery, routes, test assembly | Private Timeline store and query/profile services | Boundary guards, store and integration tests | Timeline command/query behavior | Timeline application layer | high | Legitimate façade; preserve caller compatibility while giving explicit bulk kinds distinct handlers. |
| `internal/modules/timeline/hooks.go` | Injects before-commit test hooks through HTTP dependency overrides | `TestFacadeOption`, testing constructors and dependency helpers | Test/server profile harnesses | HTTP dependency set, Timeline façade, Postgres | Conflict and integration tests | Test-only transport seam | Timeline test support | medium | Test-only assumptions leak through a production package. |
| `internal/modules/timeline/incident_bundle_portability.go` | Exports and imports raw Timeline profile/event source families | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Incident-bundle source/routes | Incident portability, direct SQL, Timeline source tables | Incident-bundle integration tests | Bundle portability contract | Timeline source-owner adapter | high | Projections and runtime state must remain excluded. |
| `internal/modules/timeline/lifecycle_store.go` | Implements review and supersede mutations, history, links, projection refresh, and replay | Private store methods through `Facade` | Timeline and workbook route coordination | Direct SQL; links, revisions, projections, idempotency | `store_test.go`, `timeline_event_integration_test.go` | Lifecycle, supersede, revision, row-version contract | Timeline mutation coordinator | high | Reviewer/admin authorization and terminal superseded state are frozen. |
| `internal/modules/timeline/linkeffects/linkeffects.go` | Converts Timeline link types into field invalidations | `LoadTimelineInvalidationsTx` | Entities merge ports | Direct SQL, Timeline mention effects | Entity merge and Timeline integration tests | Link-derived Timeline field contract | Timeline-owned effect adapter with Links consumer | high | Permanent adapter placement is deferred. |
| `internal/modules/timeline/mentioneffects/mentioneffects.go` | Loads and updates Timeline sources for mention changes, builds rows, and triggers projection rebuilds | Timeline view ID, source/invalidation types, load/update/row/rebuild helpers | Entities mention and merge ports; `linkeffects` | Direct SQL, projection adapters, `rowsnapshot` | Entity resolution/merge and Timeline tests | Mention provenance and projection contracts | Timeline-owned effect adapter | high | Duplicates row/source logic and exposes physical projection coordination. |
| `internal/modules/timeline/mentions_collections_store.go` | Hydrates and mutates mentions, entities, links, tags, and attached evidence | Private store methods | Main store, paste store, query store | Direct SQL; entities, mentions, links, evidence, projections | `unit_test.go`, `resolution_integration_test.go`, integration suites | Collection action, provenance, evidence-count, projection contracts | Timeline coordinator through peer-owner ports | high | Cross-owner effects and refresh side effects require explicit seams. |
| `internal/modules/timeline/ports.go` | Constructs concrete auth, record, revision, projection, link, entity, and mention dependencies | Private ports and adapters | Timeline private store | Platform auth/storage and several module adapters, including a direct Projections adapter import | Store, integration, and boundary tests | Authorization, revision, projection, link and mention contracts | Timeline composition boundary | high | RB-002 requires an injected Timeline-owned writer port implemented and wired by Projections, not a Timeline import of Projections runtime adapters. |
| `internal/modules/timeline/projection_contract_test.go` | Guards Timeline projection field and row-shape expectations | Test function only | Test runner | Projection contract helpers | This file | Timeline view/projection contract | Timeline test support | medium | Characterization must survive projection-seam changes. |
| `internal/modules/timeline/query_projection_paging_test.go` | Guards keyset-bounded Timeline page SQL | Test function only | Test runner | `query_projection_store.go` | This file | Query pagination contract | Timeline test support | medium | `TestTimelinePageSQLIsKeysetBounded` is currently unselected and requires RB-003 disposition before SQL movement. |
| `internal/modules/timeline/query_projection_store.go` | Directly queries `timeline_grid_projection`, applies filter/sort/page policy, hydrates collections, and builds rows | Private store methods through `Facade` | Workbook query store, recovery probe | Direct projection SQL/sqlc, view schema, collection hydration, row presenter | Paging, schema guard, store and integration tests | Timeline query, sort, filter, group, and paging contract | Timeline descriptor intent plus Projections query execution | high | RB-002 and RB-004 assign all projection-table SQL and physical paging to Projections. |
| `internal/modules/timeline/query_schema_guard_test.go` | Guards query field-to-schema mapping | Test function only | Test runner | Query store and authored view schema | This file | `cartulary.view.timeline.v2` | Timeline test support | medium | Detects contract drift. |
| `internal/modules/timeline/reportingprovider/provider.go` | Collects Timeline fields and facts for reporting export | `CollectFieldsTx`, `CollectFactsTx` | Reporting export materializer | Direct SQL, reporting provider types | Reporting boundary and export tests | Reporting source-provider contract | Timeline source-owner adapter | medium | Retain as a narrow adapter unless semantics leak inward. |
| `internal/modules/timeline/request_test.go` | Characterizes auto-resolution eligibility and manual confidence | Test functions only | Test runner | API and resolution helpers | This file | Mention-resolution contract | Timeline test support | medium | Protects manual versus auto-match provenance. |
| `internal/modules/timeline/resolution_integration_test.go` | Exercises auto-resolution eligibility, suppressors, and manual confidence against storage | Test functions only | Service-backed test runner | Timeline service, entities, aliases, links, Postgres | This file | Entity mention/link provenance contract | Timeline integration evidence | high | Service-backed characterization. |
| `internal/modules/timeline/revision_provider_contribution.go` | Contributes Timeline delete/restore and rollback providers to revision assembly | `RevisionProviderContribution` | `internal/app/revisionassembly` | Revisions, delete/restore and rollback providers | Revision integration tests | Revision provider catalog | Timeline source-owner adapter | medium | Legitimate source-owner assembly contribution. |
| `internal/modules/timeline/rollbackprovider/provider.go` | Restores and touches Timeline source state during rollback | `TimelineProvider`, `NewTimelineProvider` | Timeline revision contribution | Revisions and direct SQL | `rollbackprovider/provider_test.go`, revision integration tests | Rollback/source-history contract | Timeline source-owner adapter | high | Row version and projection follow-up must remain consistent. |
| `internal/modules/timeline/rollbackprovider/provider_test.go` | Characterizes Timeline rollback value mapping | Test function only | Test runner | Rollback provider | This file | Rollback source-value contract | Timeline test support | medium | `TestTimelineProviderSourceForRollbackValueMapsTimelineCells` is currently unselected and needs Timeline/Revisions ownership disposition. |
| `internal/modules/timeline/routes.go` | Registers Timeline-owned HTTP routes, auth/session handling, direct in-memory collaboration publication, and test routes | `Service`, `RegisterRoutes`, `RegisterTestRoutes` | Server runtime and profile harness | HTTP platform, auth, Collaboration publisher/hub, Timeline façade | Support and integration route/WS tests | Time-profile, mark-reviewed, HTTP envelope and WS contracts | Timeline transport adapter; Collaboration delivery owner | high | RB-005 requires routes to stop publishing and Timeline transactions to append a durable typed intent. |
| `internal/modules/timeline/rowpresenter/rowpresenter.go` | Builds the stable Timeline row wire shape from a normalized record | `Record`, `BuildRow` | Root API and row snapshot | Field normalization and standard library | Projection/schema/integration tests | Timeline view row contract | Timeline presentation | high | Pure leaf, but parallel derivation paths feed it. |
| `internal/modules/timeline/rowsnapshot/rowsnapshot.go` | Loads Timeline source, collections, derived link/alias/time data, and builds a record row snapshot | `ErrRecordNotFound`, `Snapshot`, `BuildRecordRowTx` | Mention effects | Direct SQL; links, row presenter, time contract, field normalization | Entity mention/merge and Timeline tests | Timeline source/row/projection contract | Timeline source snapshot service | high | Duplicates query/store hydration and derivation logic. |
| `internal/modules/timeline/state.go` | Defines capture-state vocabulary, transition policy, and supersede validation | State error and transition/validation helpers | Timeline stores and API tests | UUID and Timeline request types | `store_test.go`, `support_test.go`, integration tests | Rough/enriched/reviewed/superseded lifecycle contract | Timeline | high | Legitimate domain policy with externally visible results. |
| `internal/modules/timeline/store.go` | Main transactional create, import, patch, conflict, history, projection, and idempotency service | Conflict errors/claims, mutation result, substrate snapshot, time profile, conflict parser; private store | `facade.go`, test support | Direct SQL; auth, records, revisions, projections, entities, links, evidence, idempotency | `store_test.go`, `timeline_event_integration_test.go`, support and workbook tests | Mutation, history, conflict, lifecycle, projection and auth contracts | Timeline mutation application/persistence | high | Core mixed coordinator; behavior-preserving decomposition only. |
| `internal/modules/timeline/store_test.go` | Unit-characterizes identity, state, replay, concurrency, history, rollback, closed incidents, and evidence sort | Test functions and shared `BaseTime` | Test runner and package tests | Private store/fakes | This file | Timeline mutation/query/history contracts | Timeline test support | high | Essential characterization baseline. |
| `internal/modules/timeline/support_integration_test.go` | Exercises the incident-role authorization matrix | Test function only | Service-backed test runner | Timeline routes/runtime | This file | Core 04 incident authorization | Timeline security evidence | high | Authorization must be re-derived at route time. |
| `internal/modules/timeline/support_test.go` | Unit-characterizes create coverage, state helpers, hashes, payloads, and supersede guards | Test functions only | Test runner | API/state helpers | This file | Timeline request/payload contract | Timeline test support | medium | Protects stable payload shapes. |
| `internal/modules/timeline/testsupport/asserttest/assertions.go` | Reusable Timeline database, revision, projection, and WebSocket assertions | Assertion DTOs and functions | Timeline and related module tests | SQL/Postgres, platform WebSocket | Tests that import this helper | Test evidence shapes only | Timeline-specific test support | low | Keep local unless a genuinely cross-module abstraction emerges. |
| `internal/modules/timeline/testsupport/routetest/inventory.go` | Supplies route-inventory control entries for Timeline test scenarios | `ControlQuery`, `ControlCreateAndLive` | Timeline and incident route tests | Route inventory helpers | Route/inventory tests | Public route inventory | Timeline-specific test support | medium | Test accounting helper, not runtime architecture. |
| `internal/modules/timeline/testsupport/routetest/timeline.go` | Issues Timeline create requests in route tests | `CreateRow` | Timeline and cross-module route tests | HTTP test server and auth fixtures | Route tests | Timeline create route contract | Timeline-specific test support | low | Test client only. |
| `internal/modules/timeline/testsupport/scenariotest/harness.go` | Starts Timeline runtime/server integration harnesses | `RuntimeHarness`, server alias, `StartRuntime` | Timeline integration tests | Shared application test support | Integration tests | Test runtime composition | Timeline-specific test support | medium | Harness topology is verification support, not runtime ownership. |
| `internal/modules/timeline/testsupport/scenariotest/runtime.go` | Timeline scenario helpers for create/query, row lookup, change/revision assertions | Scenario helper functions | Timeline and related integration tests | HTTP test server, database, WebSocket assertions | Integration tests | Timeline route, history, and row contracts | Timeline-specific test support | medium | Preserve semantic assertions during refactor. |
| `internal/modules/timeline/testsupport/scenariotest/ws.go` | Connects to incident WebSocket and asserts Timeline change messages | Socket payload/client and assertion helpers | Timeline collaboration tests | WebSocket test client | WS integration tests | `/ws/v1/incidents/{incident_id}` and `record_changed` | Timeline-specific test support | medium | Event path and payload are observable. |
| `internal/modules/timeline/testsupport/storetest/harness.go` | Starts Postgres-backed Timeline store harnesses | `StoreHarness`, `StartStore` | Store and provider tests | Shared Postgres test support | Store tests | Test persistence setup | Timeline-specific test support | low | Keep local while Timeline-specific. |
| `internal/modules/timeline/time_conversion_store.go` | Reads/writes singleton incident time profile and derives UTC/local pairs | Private store methods through `Facade` | Timeline routes and row derivation | Direct SQL, `timecontract` | `timecontract_test.go`, Timeline profile integration test | Time-conversion profile and dual-field contract | Timeline | high | Profile auth and exact conversion behavior are public. |
| `internal/modules/timeline/timecontract/timecontract.go` | Parses and formats Timeline UTC/local time values with fixed offsets | Parse/format functions and `LocalParseResult` | Timeline store, row snapshot, workbook projection | Standard time/string packages | `timecontract/timecontract_test.go`, integration tests | Timeline time-pair contract | Timeline | medium | Current live callers are Timeline-local; no move is supported by evidence. |
| `internal/modules/timeline/timecontract/timecontract_test.go` | Characterizes Timeline time parsing and formatting | Test function only | Test runner | `timecontract` | This file | Timeline time-pair contract | Timeline test support | medium | `TestTimelineTimeContractParsingAndFormatting` is currently unselected and requires RB-003 disposition. |
| `internal/modules/timeline/timeline_event_integration_test.go` | End-to-end backend coverage for create, patch, replay, rollback, conflicts, envelopes, rough capture, projections, auth, lifecycle, WS, and time profile | Test functions only | Service-backed test runner | Full server/runtime/Postgres test support | This file | Most observable Timeline backend contracts | Timeline integration evidence | high | Primary broad characterization source. |
| `internal/modules/timeline/timing.go` | Carries create-timing instrumentation through context | Timing recorder interfaces and context helper | Workbook telemetry and Timeline create flow | Context/time only | Measurement and workbook tests | Measurement support; no product behavior | Workbook/Timeline instrumentation seam | low | Implementation-support evidence, not Core 05 publication. |
| `internal/modules/timeline/unit_assertions_test.go` | Shared package-local Timeline unit assertions/fakes | Test-only private helpers | Timeline unit tests | Testing package and Timeline types | Package tests | Test support only | Timeline test support | low | No production surface. |
| `internal/modules/timeline/unit_test.go` | Unit-characterizes binding modes, duplicate mention provenance, and attached evidence | Test functions only | Test runner | Timeline request/store logic | This file | Mention and evidence collection contracts | Timeline test support | high | Binding/provenance tests are selected; `TestAttachedEvidenceCreateAndPatch` is currently unselected. |
| `internal/modules/timeline/workbookprojection/store.go` | Reads all Timeline source inputs needed to rebuild workbook projections | Existing `ProjectionInput`, `ListProjectionInputsTx` | Projections store/provider registry | Direct Timeline-source SQL, `timecontract` | Projection rebuild/integration tests | Timeline projection derivation contract | Timeline canonical source provider consumed by Projections | high | Preserve the exact `ProjectionInput` member set; add single-record mutation and deterministic keyset-paged listing without a parallel DTO. |

## 3. Module Boundary Diagnosis

The current target is a **mixed-responsibility package**. It contains a legitimate
Timeline source-record/application core, but it is also a mutation coordinator,
view/projection orchestration layer, transport-adjacent adapter,
persistence-adjacent adapter, workbook/import target, and home for several
cross-module provider and effect adapters. It is not a frontend shell or
grid-vendor integration layer; frontend and grid packages consume its public view
contract indirectly.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Capture, lifecycle, source-record semantics | Root state, API, and store files | Timeline | keep | Core 01–03; active `module.timeline` requirement; store tests | Central legitimate module responsibility. |
| Stable application command façade | `facade.go` | Timeline | keep | Live workbook/import/recovery callers; boundary guards | Preserve command behavior and public shape while internals move. |
| Wire decoding and HTTP error mapping | `api.go`, `api_errors.go`, `routes.go` | Timeline transport adapter plus workbook route owner | split | Direct route callers and public envelopes | Split by layer without changing route or payload shapes. |
| Source snapshot, hydration, and row presentation | `query_projection_store.go`, `rowsnapshot`, `rowpresenter`, collection store | Timeline | split | Duplicate SQL/derivation paths and shared row contract | Create one Timeline-owned semantic derivation path. |
| Physical workbook projection persistence/rebuild/query | Query store, `workbookprojection`, mention effects, projection adapters | Projections with Timeline source/descriptor ports | move/split | Projections imports Timeline inputs while Timeline imports a Projections adapter and queries projection storage | RB-002 resolves the seam: Projections owns physical SQL/lifecycle; mutation initiator owns the transaction. |
| Clipboard paste and bulk command semantics | `clipboard_paste.go`, `bulk_mutation.go`, paste store | Workbook admission, Tabular Ingest planning, Timeline application | split | Workbook owns generic routes; current Timeline bulk compiles through paste; shared tabular parser exists | RB-001 resolves permanent semantic ownership and prohibits bulk-through-paste reuse. |
| Mention/entity/link/evidence effects | Collection store, `auto_resolution.go`, `mentioneffects`, `linkeffects` | Respective peer owner plus narrow Timeline effect ports | split/defer | Live peer-module imports and direct SQL | Preserve provenance, invalidations, and transactional behavior. |
| Revision, delete/restore, rollback | Revision contribution and provider subpackages | Timeline source-owner adapters | keep | Revision assembly imports the contribution | Keep narrow; generic coordination remains Revisions-owned. |
| Incident bundle and reporting inputs | Portability and reporting providers | Timeline source-owner adapters | keep | Incident bundles and reporting call them directly | Export raw source facts only; exclude projections/runtime state. |
| Collaboration publication | `routes.go`, mutation results, Collaboration publisher and in-memory hub | Timeline intent producer plus Collaboration sequencer/delivery | move/split | Routes publish directly after façade success; replay and sequence are currently in-memory | RB-005 requires durable transaction-coupled intent and removes publication from routes. |
| Time parsing/profile semantics | `timecontract`, time profile store, row derivation | Timeline | keep/split | Only Timeline production callers found | Consolidate use; do not move solely because it is a subpackage. |
| Module-specific test support | `testsupport`, package tests | Timeline test support | keep | Helpers encode Timeline routes, rows, revisions, and WS | Move only reusable composition to `internal/testutil`. |
| Frontend/grid consumption | Outside target | Web shell and grid adapter | defer | Searches find field/view contract consumers, no vendor import in target | Contract risk only; no backend-to-vendor dependency found. |

### Resolved owner allocation

| Responsibility | Permanent owner | Required boundary |
| --- | --- | --- |
| Browser paste capture | Web Workbook/grid adapter | Produce stable interaction intent; never send vendor coordinates or rendered row positions to the backend. |
| Workbook clipboard and bulk admission | Workbook | Own existing routes, authorization/admission, stable targets, command lookup, and public envelopes. |
| Tabular parsing and mapping | Tabular Ingest | Produce one ordered field-keyed normalized plan; ordinary clipboard paste creates no durable import session/unit. |
| Timeline paste application | Timeline | Apply source defaults, conflicts, raw capture, owner-port effects, history, idempotency, projection mutation, and event intent. |
| Timeline bulk commands | Timeline | Handle `fill_down_v1` and `multi_row_tag_assignment_v1` as distinct exact commands over stable targets. |
| Timeline projection source semantics | Timeline | Own the canonical source snapshot, derivation service, query descriptor intent, and existing `ProjectionInput` DTO. |
| Projection storage/query/rebuild | Projections | Own all `timeline_grid_projection` SQL, bounded query execution, upsert/delete, rebuild, and storage lifecycle. |
| Mutation transaction coordination | Initiating authoritative source owner | Pass one transaction to Timeline and peer-owner participants; no participant commits independently. |
| Timeline source SQL | Timeline-private persistence | Access `timeline_events` and Timeline profile/raw-capture invariants only. |
| Peer-owner source state | Records, Revisions, Entities/Mentions, Links/Tags, Evidence, idempotency owners | Expose exact transactional ports; Timeline does not write peer tables directly. |
| Collaboration intent production | Timeline application | Append one durable semantic intent per committed affected record in the source transaction. |
| Collaboration sequencing/delivery | Collaboration | Assign event identity/sequence after commit, retain replay, retry, and authorize delivery/resume. |

### Planned internal interface records

These are planned internal contracts, not current public or machine authority.
Their authoritative Core, appendix, manifest, and schema inputs must be adopted in
SL-00A before production implementation. Existing public route, request, response,
error, view-schema, and WebSocket shapes remain unchanged.

#### Batch planning and owner application

`tabular_row_plan_v1` is the planned versioned successor boundary to the current
unversioned `tabularingest.BatchPlan`. It carries the immutable target view
schema, normalized source format, source columns in ordinal order, rows and values
in source order, intentionally unmapped raw values, entity binding modes,
deterministic mapping fingerprint, and closed warnings. It contains no HTTP
objects, grid-vendor coordinates, database handles, SQL, rendered-label
identities, or parser-specific objects.

`owner_batch_apply_v1` receives server-derived incident/actor identity, exact view
schema, public `client_txn_id`, ordered create/record targets with base row
versions, one closed operation discriminator, the tabular plan only for
`clipboard_paste_v1`, a command payload only for its exact bulk kind, and a
caller-supplied transaction capability. It reuses the current public batch result
and does not commit independently.

The batch invariants remain: validate every target before row-version evaluation;
reject the whole batch for missing, foreign, deleted, wrong-type, wrong-surface,
or hidden targets; preserve current unknown-column raw capture; create one
attributable change set for the committed portion; leave same-field conflicts
unresolved; and make exact replay create no second source, history, projection,
intent, or public event effect.

#### Projection source and writer

The exact current member set of
`internal/modules/timeline/workbookprojection.ProjectionInput` is the only
compatibility row DTO during this refactor. No parallel DTO is permitted.

| Planned interface | Owner | Required behavior |
| --- | --- | --- |
| `BuildProjectionMutationTx(transaction, incident_id, record_id)` | Timeline | Return a closed `upsert` carrying current `ProjectionInput`, or `delete` for a source that no longer projects; malformed source is an error. |
| `ListProjectionInputsTx(transaction_or_snapshot, incident_id, querypage_window)` | Timeline | Return deterministic keyset-paged, snapshot-consistent, projection-eligible rows; no `OFFSET`. |
| `ApplyTimelineProjectionTx(transaction, projection_mutation)` | Required by Timeline; implemented by Projections | Apply projection upsert/delete only, in the caller transaction, without source reads, authorization, history, or independent commit. |

Create, patch, paste, bulk, lifecycle, peer-owner effects, and rollback are
coordinated by the owner that initiated the authoritative mutation. Projection
failure aborts that transaction. Projections coordinates incident/restore
rebuild, while Timeline supplies ordered source rows. Workbook coordinates the
public query while Projections executes physical bounded paging from
Timeline-owned descriptor intent.

#### SQL ownership

| SQL responsibility | Permanent owner and access rule |
| --- | --- |
| Timeline events, raw capture, and time profile | Timeline-private persistence behind Timeline application/source interfaces. |
| Generic record envelope | Records owner port; Timeline does not duplicate envelope SQL. |
| Change sets, revisions, mutations, rollback history | Revisions/History owner port with Timeline before/after values. |
| Entity/mention state | Entities/Mentions transactional ports; no direct Timeline writes. |
| Record links and tags | Links/Tags transactional ports. |
| Evidence and attachment state | Evidence port; badge/count derivation uses owner-approved read data. |
| Idempotency | Existing route-scoped owner/platform contract; no Timeline-specific duplicate. |
| `timeline_grid_projection` and projection query SQL | Projections-owned persistence/query engine only. |
| Portability/reporting source reads | Narrow Timeline source-owner adapters using Timeline-private persistence or canonical derivation. |
| Connection, transaction, execution, cancellation, generic DB error handling | Domain-neutral platform storage with no Timeline field or lifecycle policy. |

All repositories accept the caller transaction and never commit, roll back,
publish, or start an independent transaction. Any temporary cross-owner SQL
exception must have an exact query/table allowlist, owner approval, boundary test,
and removal slice; wildcard exceptions are forbidden.

#### Durable collaboration intent

Timeline will require an interface equivalent to
`AppendRecordChangeIntentTx(transaction, record_change_intent_v1)`. The intent
contains a deterministic unique key, incident/record/resulting row version,
change-set and actor identity, canonical changed field keys and affected views,
and batch mutation ordinal where applicable. It excludes public `event_id`,
`stream_seq`, `emitted_at`, connection/subscriber/session data, SQL details, and
members outside the current `record_changed` contract.

The intent is inserted with source state, history, idempotency success, and
projection refresh. Collaboration claims it after commit, assigns event identity
and per-incident sequence, appends replay state, marks publication, and delivers
only to currently authorized subscribers. Restart, replay, or dispatcher retry
must not duplicate an event. Broadcast failure does not roll back source state;
the durable log remains replayable.

| Condition | Source commit | New intent | New public event |
| --- | --- | --- | --- |
| Authentication, CSRF, authorization, malformed request, invalid target, idempotency conflict, or conflict-only batch failure | no | no | no |
| Exact successful idempotent replay | no new commit | no | no |
| Partial batch with committed rows and conflicts | committed portion | one per committed record | one per committed record |
| Projection, history, or intent insertion failure | no | no | no |
| Successful change with no connected subscriber | yes | yes | yes, retained for replay |
| Dispatcher unavailable or process restarts after commit | yes | pending durable intent | delayed, exactly once |
| Broadcast fails after log append | yes | published intent | existing replayable event; no new sequence |
| Membership is revoked before delivery | yes | yes | not delivered to that unauthorized connection |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `cartulary.view.timeline.v2` fields, visibility, sort, filter, group, and write policy | Timeline view-contract owner | Authored view schema; Core 01 §7.4.1; query/presenter code | Schema guard, projection contract, browser/frontend owner rows | Retain field-by-field row and query matrix | high | Authored contract input must not be hand-edited as a refactor shortcut. |
| Generic view query route | Workbook route owner with Timeline query façade | Core 01 route family; workbook query store; Timeline query store | Paging, schema guard, store, integration, browser tests | Filter/sort/group/cursor and hidden collection cases | high | Physical projection storage must remain invisible to callers. |
| Create and patch routes | Workbook route owner with Timeline mutation façade | Core 01 mutation routes; `api.go`, `store.go` | Decoder, store, integration, workbook tests | Blank/one-value capture, replay, closed incident, actor-scoped idempotency | high | Preserve envelopes and row versions. |
| Clipboard-paste route | Workbook admission, Tabular Ingest plan, Timeline application | Core 03 clipboard behavior; clipboard files | Timeline and workbook clipboard tests | Batch atomicity, raw unknown columns, mention origin, cross-incident rejection | high | RB-001 is resolved; existing public request/result envelopes remain unchanged. |
| Bulk-mutation route | Workbook admission with exact Timeline command handlers | Core 03 bulk behavior; `bulk_mutation.go` | Workbook bulk/route tests | `fill_down_v1`, tag assignment, stable record IDs, replay | high | Stop compiling bulk through paste; vendor coordinates never enter the backend contract. |
| Planned `tabular_row_plan_v1` and `owner_batch_apply_v1` | Tabular Ingest contract and source-owner applicator | Closure decision; current `tabularingest.BatchPlan` and Timeline façade | Clipboard/bulk unit and integration tests | Ordered mapping, raw preservation, exact operation dispatch, transaction ownership | high | Internal-only planned contract; requires later Core/machine adoption. |
| Mark-reviewed and supersede routes | Timeline | Core 01/03; routes and lifecycle store | Store and Timeline integration tests | Role gates, terminal supersede, replacement links, replay/rollback | high | Lifecycle is observable source state. |
| Conflict resolution route and token | Workbook/Timeline application boundary | Core 03 OCC; conflict decoder/store | Store and integration conflict tests | Same-field conflict, stale/invalid token, resolution replay | high | Field-level OCC and error envelopes are frozen. |
| Timeline time-conversion profile GET/PUT | Timeline | Core 01 REQ-01-611 through REQ-01-613; route/store | Time contract and profile integration tests | Defaults, membership read, admin write, round-trip pairs | high | Exact auth and conversion fields must remain. |
| Incident WebSocket and `record_changed` | Timeline intent producer with Collaboration sequencing/delivery | Core 01 WS contract; current direct route publisher; scenario WS helpers | Canonical WS integration and collaboration tests | Current membership, changed record/version, affected views, no-event/retry/restart failures | high | Public path remains `/ws/v1/incidents/{incident_id}`; planned durable intent is internal. |
| Row version, idempotency, revisions, change sets, and rollback | Timeline plus Records/Revisions owners | Core 02/03; store and providers | Store, Timeline integration, revision tests | Exact replay, concurrent mutation, coupled supersede, rollback projection follow-up | high | Projection rows are not history. |
| Mention provenance and manual/auto resolution | Entities/Mentions/Links with Timeline source owner | Core 02/03; auto-resolution and collection code | Request, unit, resolution integration, entity tests | Interactive-only auto-match, suppressors, null manual confidence, repeated mentions | high | No implicit stub creation from Timeline capture. |
| Tags, typed links, record references, and attached evidence | Tags/Links/Evidence with Timeline collection policy | Core 02; collection descriptor/store | Unit, links/tags, store and integration tests | Add/remove, invalid targets, evidence counts, projection invalidation | high | Collection side effects must stay transactional. |
| Saved-view and view-schema compatibility | Saved Views and view-contract owners | Core 01/04; saved-view tests; view schema | Saved-view and frontend tests | Existing saved views against unchanged field IDs and query semantics | high | Saved-view scope does not alter incident row authorization. |
| Projection refresh, query, and deterministic rebuild | Projections with Timeline source/descriptor input | Core 01 §8; current adapter/provider registry; schema ownership manifest | Projection contract and Timeline integration tests | Commit/rollback refresh, keyset query, rebuild equivalence, invalidation coverage | high | Preserve `ProjectionInput`; projection-table SQL and storage lifecycle are Projections-owned. |
| Planned projection mutation/source/writer ports | Timeline contract with Projections implementation | Closure decision; current `workbookprojection` and direct adapter import | Projection, paging, rebuild, rollback, entity-effect tests | Upsert/delete, caller transaction, page determinism, fail-closed errors | high | Internal contract requires owner-input adoption before code movement. |
| Planned `record_change_intent_v1` | Timeline semantic intent and Collaboration durable delivery | Closure decision; current route publisher and in-memory hub | Route, WS, replay, authorization and no-event tests | Exactly-once intent/event posture, post-commit sequence, restart/retry | high | Requires separately authorized contract and migration work; public event shape is frozen. |
| Incident role authorization | Auth platform and route/application adapters | Core 04 REQ-04-021 through REQ-04-029 and REQ-04-127 | Authorization matrix and route integration tests | Membership loss, viewer/editor/reviewer/admin gates, hidden incident | high | `deployment_admin` is not an incident-access bypass. |
| Incident bundle portability | Incident Bundles with Timeline source provider | Core portability contract; provider code | Incident-bundle integration tests | Raw profile/events round trip, attribution, exclusion of projections | high | Source-owner validation remains required. |
| Reporting source fields/facts | Reporting with Timeline provider | Reporting provider and materializer | Reporting boundary/export tests | Stable support refs and source field/fact selection | medium | Narrow provider boundary. |
| Generated HTTP/protocol surface | Contract/codegen owners | OpenAPI/protocol inputs and generated consumers | Codegen boundary and HTTP conformance tests | Generated-drift checks if owner inputs change | high | Never hand-edit generated roots. |
| Frontend selectors and grid-adapter field IDs | Web and grid-adapter owners consuming Timeline contract | Live TypeScript searches and view schema | Frontend unit, browser, visual, accessibility rows | Selector/field compatibility only if backend row contract changes | medium | No direct grid-vendor import exists in this target. |
| `module.timeline` harness accounting | Verification owner | Exact-selector Harness contract; owner verification file and 49-row test-family manifest | 32 Go, 11 Playwright, and 6 Vitest runner rows | Exhaustive `timeline_test_disposition_v1` and selector validation | high | Live all-owner comparison found ten unselected executable identities; filename presence itself remains irrelevant. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Timeline constructs projection adapters while Projections imports Timeline projection inputs | `ports.go`, `mentioneffects`, `workbookprojection`, projection store/provider registry | Cyclic ownership and hard-to-isolate rebuild semantics | `must_fix` | Projections owns storage/query/rebuild; Timeline owns source/descriptor input | Implement the RB-002 source/mutation/writer ports after owner-authority adoption. |
| Source loading, collection hydration, time derivation, and row construction are duplicated | Query store, `rowsnapshot`, collection store, row presenter, workbook projection | Drift between query, mutation effects, and rebuild rows | `must_fix` | Timeline | Define one source snapshot/row derivation service and characterize equivalence. |
| `api.go` combines wire decoding, hashes, policy, payloads, and presentation | Exact exported surface and callers | A small change can alter several public contracts | `should_fix` | Timeline application/transport adapters | Partition by concern while retaining compatibility exports until callers migrate. |
| Direct SQL is distributed across root and semantic provider subpackages | Store, auto-resolution, snapshot, time, portability, reporting, effect providers | Persistence details obscure semantic ownership and permits cross-owner writes | `must_fix` | Table/invariant owner under the RB-004 matrix | Inventory every statement, move projection/peer SQL to owner ports, and keep platform storage domain-neutral. |
| Test dependency overrides live in a production package | `hooks.go` and server profile harness callers | Test assumptions enlarge production surface | `should_fix` | Timeline test support/application assembly | Design a test composition seam without changing runtime ownership. |
| Large mutation coordinators hide revision, projection, mention, link, and evidence effects | Main, lifecycle, paste, and collection stores | Ordering or transaction regression during extraction | `should_fix` | Timeline coordinator using explicit owner ports | Characterize transaction and replay outcomes before extraction. |
| Workbook clipboard and bulk semantics are embedded in Timeline and bulk compiles through paste | Workbook owns routes; Timeline owns parser/plan/store; shared Tabular Ingest exists | Duplicate target behavior and loss of exact command semantics | `must_fix` | Workbook admission, Tabular Ingest plans, Timeline applicator/handlers | Implement RB-001 and remove `BulkMutationClipboardRequest` reuse only after parity tests. |
| Mention/link/evidence effects reach Timeline SQL and projection adapters | Entities imports Timeline subpackages; collection/effect code touches peer state and projections | Peer owners can depend on Timeline internals or bypass table ownership | `should_fix` | Peer-owner transaction ports plus Timeline derivation participant | Apply RB-002/RB-004 ports after the canonical derivation service exists. |
| Collaboration publication is route-adjacent and in-memory | Timeline routes publish after mutation; hub assigns sequence/replay in memory | Commit/event gaps, loss on restart, duplicates on retry, route coupling | `must_fix` | Timeline intent port and Collaboration sequencer/delivery | Implement RB-005 durable intent and no-event/retry matrix after contract/migration authorization. |
| Route-time membership and role checks occur in application/transport adapters | Core 04 and live route/workbook service calls | Moving checks inward could leak resource existence | `intentional/no_action` | Auth/application adapters | Preserve re-authorization and hidden-incident behavior. |
| Revision, reporting, portability, delete/restore, and rollback providers are source-owner adapters | Live assembly/import callers and narrow provider files | Wholesale moves would invert ownership | `intentional/no_action` | Timeline | Audit narrowness; retain unless behavior leakage is demonstrated. |
| `timecontract` is Timeline-local in current production imports | Live import search | Premature sharing would create an ownerless helper | `intentional/no_action` | Timeline | Consolidate its use without moving it. |
| No direct grid-vendor dependency exists in the target | Live Go/TypeScript import search | False-positive architecture work | `intentional/no_action` | Grid adapter remains outside Timeline | Track only view/selector compatibility. |
| Ten executable Timeline test identities are unselected across active family manifests | Exact symbol comparison against all active Go selectors | Retained tests lack accountable executable evidence or explicit retirement | `must_fix` | Verification owner, with postcondition owners as collaborators | Complete `timeline_test_disposition_v1`; add/move/delete only from approved exact dispositions. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Fix authority, scope, repository revision, and write boundary | This tracker; requirement/verification owners | `make task-guide ROLE=module-author OWNER=module.timeline` | Source posture and session log complete. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03 | Account for every target file, caller, dependency, and risk | All `internal/modules/timeline` files; inbound callers | Repository searches and exact-source review | All 54 files have current-state rows. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-04 | Map routes, view/protocol fields, auth, projection, revision, and provider contracts | Contract owners, view schema, routes, façades, providers | `make explain-test-owner OWNER=module.timeline` | Every discovered contract has an owner and evidence posture. |
| WF-03 | Characterization and exact-selector accounting | parallel | WF-01 | WF-04, WF-07 | Produce `timeline_test_disposition_v1` for every executable identity and establish the baseline | Target tests, test support, all owner family manifests | Owner explanation plus exact selector comparison and validation | Every retained test is selected exactly once or has an approved action. |
| WF-04 | Boundary and coupling scan | chain | WF-02, WF-03 | WF-05 | Classify coupling and distinguish legitimate Timeline ownership from catch-all behavior | Target production files and peer-owner adapters | `make backend-module-boundary-check` during implementation | Findings have classifications and proposed owners. |
| WF-05 | Owner-authority promotion and interface design | chain | WF-04 | WF-06, WF-07 | Promote resolved tracker decisions into owning Core/appendix/machine inputs and define exact acyclic ports | Core 01/03, Appendix I, provider/boundary/schema-owner inputs, internal contract schemas | Owner review, Markdown/JSON/drift and boundary checks | RB-001, RB-002, RB-004, and RB-005 decisions are authoritative before production edits. |
| WF-06 | Slice sequencing | chain | WF-05 | WF-08 | Convert approved boundaries into smallest behavior-preserving changes | Packages named in Section 7 | Narrow owner targets after each slice | Every slice has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-05 | WF-08 | Apply only approved dispositions and close RB-003 evidence | Verification owner/catalog inputs and tests if later authorized | Exact-selector, overlap, catalog drift, and owner explanation targets | Every identity has one disposition; retained evidence resolves exactly once. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Run narrow-to-broad verification and refresh handoff evidence | Tracker and later authorized changed files | `make agent-finalize`, risk-appropriate `make test-fast` and `make check` | Successful commands, artifacts, failures, and remaining risks are recorded. |

## 7. Proposed Refactor Slice Plan

All slices require a later authorized implementation task. They are intended to
preserve behavior. If an approved design introduces behavior change, that slice
must be relabeled `requires later authorization` before implementation.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00A | Closure decisions in this tracker; later owner authorization | Promote RB-001/RB-002/RB-004/RB-005 into Core 01/03, Appendix I, provider/boundary/schema-owner inputs, and planned internal contract schemas before code movement | Owner documents and authored machine inputs only; generated outputs through generators | Tracker evidence must not silently become runtime authority | Owner contract, JSON-shape, drift, and boundary-policy checks | `make lint-markdown`; `make json-shape-check`; `make generate-drift`; `make backend-module-boundary-check` | Revert authored owner inputs and their generated refresh together | Core and machine owners agree on interfaces, table ownership, transactions, and event intent. |
| SL-00 | WF-03; RB-003 evidence completion | Produce `timeline_test_disposition_v1`, repair only proven selector gaps, and run the frozen characterization baseline | Timeline tests/testsupport; verification inputs only if separately authorized | Incomplete or duplicate accounting could hide regressions | Preserve retained tests; disposition the ten unselected identities; add tests only for demonstrated contract gaps | `make test-slice OWNER=module.timeline`; `make service-backed-test-slice OWNER=module.timeline` | Revert approved test/catalog changes without production impact | Every executable identity has one disposition and retained evidence selects exactly once. |
| SL-01 | SL-00, SL-00A | Partition request decoding, error mapping, row presentation, and command declarations while retaining compatibility exports | `api.go`, `api_errors.go`, `rowpresenter`, `facade.go`, workbook callers | Request validation, hashes, errors, row shape | Decoder, error classifier, payload, boundary and frontend contract tests | `make test-slice OWNER=module.timeline` | Revert leaf moves and compatibility forwarding together | No route/payload/hash/export behavior changes; mixed API file is reduced. |
| SL-02 | SL-01, SL-00A | Consolidate Timeline-owned source SQL, hydration, collection inputs, time derivation, and row construction behind private persistence plus one derivation service | Query store, `rowsnapshot`, row presenter, collection/time stores, mention effects | Query rows, evidence count, time pairs, mention/link fields | Projection contract, schema guard, paging, time, entity effects, integration query tests | `make service-backed-test-slice OWNER=module.timeline`; `make backend-module-boundary-check` | Keep old derivation path until equivalence is proven, then remove it atomically | Query, effect, refresh, and rebuild paths derive byte-equivalent contract rows without cross-owner SQL. |
| SL-03 | SL-02, SL-00A | Implement Timeline source/mutation ports and the Projections-owned writer/query/rebuild seam while preserving `ProjectionInput` | Timeline ports/workbookprojection/mention effects; Projections query/store/providers | Transactional upsert/delete, bounded paging, deterministic rebuild, invalidations | Projection contract, paging, rebuild, rollback, restore and peer-effect tests | `make service-backed-test-slice OWNER=module.timeline`; `make backend-module-boundary-check`; provider manifest parity checks | Retain old adapter wiring until new ports pass parity; no source-data migration | Timeline imports no Projections runtime adapter, and projection storage/query behavior is Projections-owned and unchanged. |
| SL-04 | SL-02, SL-03, SL-00A | Narrow entity, mention, link, tag, evidence, revision, record, idempotency, and projection effects behind explicit caller-transaction ports | Collection store, auto-resolution, mention/link effects, ports, peer owner adapters | Provenance, link metadata, evidence count, history, transaction ordering | Resolution, attached evidence, links/tags, merge, rollback and replay tests | `make service-backed-test-slice OWNER=module.timeline`; peer-owner slices | Reconnect prior concrete adapters as one reversible change | Timeline writes only Timeline-owned source state and all cross-owner effects remain atomic. |
| SL-05 | SL-00, SL-00A | Implement Workbook admission, versioned Tabular Ingest plans, Timeline owner application, and distinct fill/tag bulk handlers | Clipboard/bulk files, Timeline façade/store, workbook routes/stores, shared Tabular Ingest | Batch atomicity, hashes, raw columns, stable IDs, mention origin, operation identity | Timeline/workbook clipboard and bulk unit/integration tests, including cross-incident targets and exact replay | `make service-backed-test-slice OWNER=module.timeline`; relevant Workbook/Tabular Ingest owner slices | Preserve current public façade/envelopes until all internal callers migrate; revert migration together | No bulk handler compiles through paste, one mapping engine exists, and public behavior is unchanged. |
| SL-06 | SL-01, SL-04, SL-00A; separate contract/migration authorization | Add durable `record_change_intent_v1`, Collaboration dispatcher/log sequencing, route decoupling, and test-only composition cleanup | Timeline routes/store/ports, Collaboration, application assembly, authored migration/contract inputs | Auth order, source atomicity, sequence/replay, retry/restart, session sliding, WS payload | Authorization, route envelope, canonical WS, no-event matrix, retry/restart, replay retention and membership-loss tests | `make service-backed-test-slice OWNER=module.timeline`; Collaboration owner slice; `make browser-e2e-webserver-backed` | Deploy durable intent path behind compatible assembly and revert application wiring/migration only under an approved rollback plan | Routes do not publish, committed changes yield exactly one durable/replayable event, and public WS behavior is unchanged. |
| SL-07 | SL-04 | Audit and retain narrow portability, reporting, delete/restore, and rollback providers | Provider subpackages, revision contribution, incident bundle/reporting callers | Bundle round trip, report facts, rollback row versions/projections | Incident-bundle, reporting, revision, rollback provider tests | Relevant owner slices plus `make backend-module-boundary-check` | Revert provider-by-provider; no data migration is expected | Providers contain source-owner adaptation only and generic coordination stays with peer owners. |
| SL-08 | SL-03, SL-04, SL-05, SL-06, SL-07 | Run narrow-to-broad verification, inspect drift, and update handoff | All later-authorized changed files and this tracker | Cross-owner and frontend contract regression | All preserved/added tests and browser rows selected by changed contracts | `make agent-finalize`; risk-appropriate `make test-fast` and `make check` | Roll back the first failing slice, not unrelated successful slices | Required targets pass or failures are documented with run roots and ownership. |

## 8. Validation Plan

Discovery commands passed during the planning sessions. No product test suite was
run. The initial tracker `make lint-markdown` run root was
`.cartulary/test-results/20260726T024332Z-p557264`; the closure-pass run also
passed with run root `.cartulary/test-results/20260726T032316Z-p570550`.
`git diff --check` and cumulative `git diff HEAD --check` passed.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Tracker Markdown | yes | Required for this planning-only change. |
| owner discovery | `make task-guide ROLE=module-author OWNER=module.timeline` | Current owner routing | yes | Passed during discovery. |
| owner accounting | `make explain-test-owner OWNER=module.timeline` | Active requirement and 49-row owner family | yes | Passed during discovery; disposition must also compare every exact executable identity across all owners. |
| harness/accounting contract | `make harness-contract` | Exact selectors, overlap, owner/verification/catalog integrity | yes | Required after later authorized RB-003 catalog changes; no filename-derived row creation. |
| unit | `make test-slice OWNER=module.timeline` | Owner-selected Go/Vitest/Playwright non-service rows | yes | Use narrower `ROWS=` only after the owner explanation identifies safe row IDs. |
| integration | `make service-backed-test-slice OWNER=module.timeline` | Owner-selected service-backed rows | yes | Required before mutation/projection/persistence slices. |
| e2e/browser | `make browser-e2e-support`; `make browser-e2e-webserver-backed` | Browser support and live-server workbook behavior | no | Required when route, WS, workbook, or frontend contracts are touched. |
| e2e/browser | `make browser-e2e-measurement`; `make browser-e2e-visual` | Measurement and visual support | no | Run only when affected; this tracker makes no Core 05 claim. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Contract/codegen outputs and JSON authority | no | Required if a later authorized task changes authored contract inputs; never hand-edit outputs. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module dependency direction | yes | Run after every boundary-changing slice. |
| import-boundary/static | `make frontend-import-boundary-check` | Frontend/grid module dependency direction | no | Required only if frontend consumers change. |
| focused broad | `make test-fast` | Fast repository verification | no | Run after owner slices when cross-owner risk justifies it. |
| finalization | `make agent-finalize` | Handoff and retained-evidence maintenance | no | Run before broader end-of-run verification as repository procedure requires. |
| full check | `make check` | Full repository check | no | Run at the end when the cumulative risk or owner guidance requires it. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| TL-001 | Establish Timeline source posture and write boundary | WF-00 | DONE | None | Section 1 and planning session log | Authority and allowed-file scope are explicit. |
| TL-002 | Inventory all 54 target files and live callers | WF-01 | DONE | TL-001 | Section 2 | Every target file has an individual row. |
| TL-003 | Map observable contracts and owners | WF-02 | DONE | TL-002 | Section 4 | Every discovered contract risk has an owner and evidence posture. |
| TL-004 | Record initial coupling and boundary diagnosis | WF-04 | DONE | TL-003 | Sections 3 and 5 | Findings use the required classifications and avoid permanent unsupported decisions. |
| TL-005 | Complete exact-selector disposition and accounting | WF-03, WF-07 | BLOCKED | TL-002 | RB-003; ten unselected identities; planned `timeline_test_disposition_v1` | Every executable identity has one disposition and retained evidence selects exactly once. |
| TL-006 | Approve the Timeline/Projections seam | WF-05 | DONE | TL-004 | RB-002 closure decision and Section 3 port/transaction record | Timeline owns source semantics; Projections owns physical query/storage/rebuild. |
| TL-007 | Approve clipboard/bulk ownership | WF-05 | DONE | TL-004 | RB-001 closure decision and Section 3 owner allocation | Workbook, Tabular Ingest, and Timeline responsibilities are explicit. |
| TL-008 | Approve persistence consolidation depth | WF-05 | DONE | TL-004 | RB-004 closure decision and SQL ownership matrix | SQL follows table/invariant ownership and platform storage stays neutral. |
| TL-009 | Characterize and consolidate row derivation | WF-03, WF-06 | TODO | TL-005, TL-008, TL-016 | SL-00 through SL-02 | All row-producing paths are equivalent and use one semantic service. |
| TL-010 | Invert projection ownership | WF-06 | TODO | TL-006, TL-009, TL-016 | SL-03 | Projection dependency direction is acyclic and tests pass. |
| TL-011 | Narrow cross-owner mutation effects | WF-06 | TODO | TL-008, TL-009, TL-010, TL-016 | SL-04 | Explicit ports preserve transactional side effects. |
| TL-012 | Refactor workbook paste/bulk boundary | WF-06 | TODO | TL-007, TL-016 | SL-05 | Distinct commands use the approved ownership without wire changes. |
| TL-013 | Approve durable collaboration publication ownership | WF-05 | DONE | TL-004 | RB-005 closure decision and Section 3 event-intent record | Timeline owns intent; Collaboration owns post-commit sequence/replay/delivery. |
| TL-014 | Audit source-owner providers | WF-06 | TODO | TL-011 | SL-07 | Providers remain narrow and peer coordination remains outside Timeline. |
| TL-015 | Complete implementation verification and handoff | WF-08 | TODO | TL-010, TL-011, TL-012, TL-014, TL-017 | SL-08 and updated session logs | Required commands and outcomes are recorded with no unresolved regression. |
| TL-016 | Promote closure decisions into owning authority | WF-05 | TODO | TL-006, TL-007, TL-008, TL-013 | SL-00A | Core, appendix, provider/boundary/schema-owner and internal contract inputs agree before code edits. |
| TL-017 | Implement durable collaboration intent and route decoupling | WF-06 | TODO | TL-011, TL-013, TL-016 | SL-06 | Routes do not publish and committed changes produce exactly one replayable authorized event. |

## 10. Session Handoff Log

Initial session time: `2026-07-25T22:39:55-04:00`; blocker-closure session time:
`2026-07-25T23:18:54-04:00`. The requested output did not previously exist, so
the initial session had no current-file history to preserve. The archived tracker
is cited as historical evidence only. Earlier handoff rows preserve the state
known at their timestamp and are superseded by later rows where closure posture
changed. For both planning sessions, the only file touched is this tracker.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Scope and authority mapped from machine owners and live repository | Inspected framework, Core 00–04, domain, requirement/verification owners, archived tracker; touched this tracker only | `sed`, `rg`, `jq`, `make help`, `make help-all`, owner discovery commands, `make lint-markdown`, `git diff --check` | No owner contradiction found; framework/live-state mismatch recorded; documentation validation passed | None for tracker completion | Preserve this source posture in any later implementation handoff. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Closure guidance adopted as subordinate planning evidence | Inspected `temp/analysis-notes.md`, Core/repository evidence, and this tracker; touched this tracker only | `sed`, `rg`, `jq`, `find`, exact-selector comparison, `make lint-markdown`, diff/status checks | RB-001, RB-002, RB-004, and RB-005 planning decisions resolved; documentation validation passed; no implementation authorized | RB-003 evidence remains; TL-016 authority promotion is TODO | Complete SL-00A and SL-00 in later authorized owner tasks. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Live target is a mixed-responsibility package with a legitimate Timeline core | Inspected all 54 target files and inbound backend callers; touched this tracker only | `find`, `rg`, exact source reads | Projection, row derivation, workbook/import, persistence, effects, and provider boundaries classified | RB-001, RB-002, RB-004, RB-005 block implementation choices | Obtain owner decisions before SL-02 through SL-07. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Owner allocation, SQL matrix, projection ports, batch boundary, and durable event-intent design are decision-complete | Inspected Timeline/Workbook/Tabular Ingest/Projections/Collaboration sources; touched this tracker only | `sed`, `rg`, `jq` | Four design blockers no longer block slice design; current bulk-through-paste, projection imports/SQL, and route publication are recorded as implementation gaps | Production edits still require SL-00A, SL-00, and separate authorization | Promote authority, establish characterization, then execute dependency-ordered slices. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Frontend and grid adapter are contract consumers, not code inside the target | Inspected live TypeScript/view-ID consumers by search; touched this tracker only | `rg` across `apps` and `packages` | No direct backend grid-vendor coupling found; selector/field compatibility remains a risk | None until a row/view contract changes | Run frontend boundary/unit/browser targets only for affected later slices. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Browser paste and vendor callbacks are fixed as thin interaction adapters | Inspected closure guidance and current Workbook/grid boundary evidence; touched this tracker only | `sed`, `rg` | Stable backend targets and existing public envelopes remain; no vendor coordinate may cross the boundary | None for this document pass | Preserve frontend selectors and run affected owner/browser rows during SL-05/SL-06. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Timeline view, route, WS, revision, portability, and generated surfaces are frozen | Inspected view schema, Core contracts, OpenAPI/protocol references, provider code; touched this tracker only | `jq`, `rg`, `sed` | Contract map completed; no generated file edit proposed | RB-002 affects the internal projection interface, not public wire behavior | Run drift checks only if later owner inputs change. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Planned internal batch, projection, SQL-owner, and collaboration intent contracts are specified without public wire changes | Inspected current `BatchPlan`, `ProjectionInput`, provider descriptor/registry, ownership manifest, route publisher and hub; touched this tracker only | `sed`, `rg`, `jq` | Existing public routes/view/event shapes are frozen; generated outputs remain generator-owned | TL-016 must promote planning decisions into owner authority | Author Core/appendix/machine inputs in SL-00A before production code. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Existing unit, integration, browser, support, measurement, and visual owner rows mapped | Inspected target tests/testsupport and owner test-family/verification files; touched this tracker only | `make task-guide ROLE=module-author OWNER=module.timeline`; `make explain-test-owner OWNER=module.timeline`; `rg` | Discovery commands passed; no product test suite ran during the read-only planning pass | RB-003 | Verification owner must classify unmatched-by-name tests before catalog edits. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Exact-selector rule is resolved and ten live executable identities are currently unselected | Inspected all Timeline test symbols and all active family manifests; touched this tracker only | `find`, `rg`, `jq`, exact package/symbol comparison | RB-003 is `DECISION_RESOLVED; EVIDENCE_PENDING`; filename visibility is not accounting | `timeline_test_disposition_v1` and selector validation do not yet exist | Verification owner dispositions every identity before catalog/test changes or SL-00 completion. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Incident membership and role checks are frozen at route/application boundaries | Inspected Core 04, Timeline/workbook routes, auth tests; touched this tracker only | `sed`, `rg` | Viewer/editor/reviewer/admin, hidden-incident, current-membership, and no deployment-admin bypass posture recorded | None for planning; later route moves carry high risk | Re-run authorization and WS characterization for SL-06. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Authorization remains Core 04-owned while Collaboration gains post-commit delivery responsibility | Inspected Core authorization posture, route checks, Collaboration replay and delivery code; touched this tracker only | `sed`, `rg` | No new authorization model; delivery/resume must re-evaluate current incident membership and role | Later SL-06 needs membership-loss and no-event characterization | Keep route admission checks and prove authorization-filtered delivery after event-port migration. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T22:39:55-04:00 | Codex / tracker creation | Tracker is decision-ready; implementation is not authorized | Inspected sources listed above; touched this tracker only | Repository discovery, Make-owned command discovery, Markdown lint, and diff checks | Five stable implementation blockers recorded; documentation validation passed | RB-001 through RB-005 | Begin later session with WF-03/WF-05 owner decisions, not production edits. |
| 2026-07-25T23:18:54-04:00 | Codex / blocker closure | Four planning blockers are resolved; one verification decision awaits evidence | Inspected closure notes, affected live interfaces, owner manifests, and tracker cross-references; touched this tracker only | Read-only repository/exact-selector analysis, Markdown lint, and cumulative diff checks | Slice plan is explicit and documentation validation passed | RB-003 evidence; TL-016 owner-authority promotion; all implementation needs later authorization | Start with SL-00A and SL-00, then follow SL-01 through SL-08 dependencies. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Workbook owns command admission/public envelopes; Tabular Ingest owns normalized plans; Timeline owns Timeline application and committed effects. | Prevents duplicate parsers/domain interpreters and preserves exact bulk semantics. | Planning decision adopted here; promote the owner split and internal contracts through SL-00A before code. | `RESOLVED`; no longer blocks SL-05 design. |
| RB-002 | Timeline owns canonical source/descriptor semantics and `ProjectionInput`; Projections owns projection SQL/query/write/rebuild/storage; the initiating source owner coordinates the transaction. | Removes the current two-way runtime dependency while preserving source/projection atomicity. | Planning decision adopted here; promote ports, descriptor/boundary inputs, and transaction rules through SL-00A. | `RESOLVED`; no longer blocks SL-03 design. |
| RB-003 | Every executable test identity requires one exact verification-owner disposition; filename or package visibility has no accounting meaning. | Active requirements need exact accountable executable evidence, and retained tests may not be silently unselected or duplicated. | Complete and approve `timeline_test_disposition_v1`; repair/delete/move only from its evidence; pass selector/catalog validation. | `DECISION_RESOLVED; EVIDENCE_PENDING`; still blocks SL-00 completion and catalog edits. |
| RB-004 | SQL resides with the owner of the tables and invariants: Timeline-private source SQL, Projections projection SQL, peer-owner ports for peer state, domain-neutral platform mechanics. | Prevents platform policy leakage, cross-owner writes, and one all-purpose Timeline repository. | Planning decision and SQL matrix adopted here; promote ownership/boundary inputs through SL-00A. | `RESOLVED`; no longer blocks SL-02/SL-04 design. |
| RB-005 | Timeline appends a durable typed event intent in the source transaction; Collaboration owns post-commit sequence, replay, retries, authorization-filtered delivery, and duplicate prevention; routes do not publish. | Closes the commit/publication gap and makes retry/restart semantics explicit without changing public WebSocket behavior. | Planning decision adopted here; later Core/internal contract and migration authorization through SL-00A/SL-06. | `RESOLVED`; no longer blocks SL-06 design. |

Resolved status means the planning choice is fixed. It does not mean the owner
authority, migration, production code, tests, catalogs, or generated artifacts
have been changed.

### RB-003 required disposition artifact

The verification owner must produce one exhaustive
`timeline_test_disposition_v1` row for every independently executable Timeline
test identity.

| Field | Required contract |
| --- | --- |
| `test_path` | Exact repository-relative test path. |
| `runner` | Exact registered runner. |
| `test_symbol_or_title` | Exact Go symbol, full Vitest title, or Playwright scenario identity/title. |
| `verified_postcondition` | One precise product, security, architecture, build, or support postcondition. |
| `classification` | One closed classification from the next table. |
| `owner_id` | Required for every retained executable test. |
| `collaborator_ids[]` | Present and possibly empty. |
| `verification_ids[]` | Nonempty for retained executable evidence. |
| `row_id` | Existing or proposed exact catalog row. |
| `selector_status` | `selected_exactly_once`, `unselected`, or `selected_more_than_once`. |
| `required_action` | `none`, `add_to_existing_row`, `create_verification_and_row`, `move_to_owner`, or `delete_obsolete_test`. |
| `rationale` | Concise evidence-based explanation. |

| Classification | Required posture |
| --- | --- |
| `direct_product_evidence` | Select exactly once through the product/security postcondition owner. |
| `architecture_support_evidence` | Select through a support-profile architecture/build/harness verification. |
| `cross_owner_product_evidence` | Assign the externally visible postcondition owner and list Timeline as collaborator where appropriate. |
| `helper_only` | No independent row; name the selected executable tests that consume it. |
| `duplicate_or_obsolete` | Retire only in a later authorized change after coverage comparison. |
| `unaccounted_gap` | Add or repair verification/catalog evidence before SL-00 completes. |

The live all-owner exact-selector comparison found these currently unselected
identities. Their classification is provisional until the verification owner
approves the disposition.

| Package | Executable identity | Provisional classification |
| --- | --- | --- |
| `./internal/modules/timeline` | `TestTimelineProductionImportBoundaries` | `architecture_support_evidence` |
| `./internal/modules/timeline` | `TestTimelineProductionFacadeCallersUseCommandBoundary` | `architecture_support_evidence` |
| `./internal/modules/timeline` | `TestWorkbookRoutesDoNotClassifyTimelineStoreErrors` | `architecture_support_evidence` |
| `./internal/modules/timeline` | `TestTimelineStoreInternalsStayPrivate` | `architecture_support_evidence` |
| `./internal/modules/timeline` | `TestTimelineFacadeDoesNotExposeLegacyMutationShims` | `architecture_support_evidence` |
| `./internal/modules/timeline` | `TestTimelinePageSQLIsKeysetBounded` | `direct_product_evidence` |
| `./internal/modules/timeline` | `TestClipboardPasteRejectsCrossIncidentRecordTarget` | `direct_product_evidence` |
| `./internal/modules/timeline` | `TestAttachedEvidenceCreateAndPatch` | `direct_product_evidence` |
| `./internal/modules/timeline/rollbackprovider` | `TestTimelineProviderSourceForRollbackValueMapsTimelineCells` | `direct_product_evidence` or `cross_owner_product_evidence` |
| `./internal/modules/timeline/timecontract` | `TestTimelineTimeContractParsingAndFormatting` | `direct_product_evidence` |

RB-003 is fully closed only when every executable identity has one disposition,
every retained test is selected exactly once, every active Timeline requirement
resolves to executable evidence, support tests do not claim product behavior, all
no-change conclusions are explained, and catalog/selector validation passes.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/timeline` is inventoried or explicitly out of scope. | PASS | Section 2 has 54 individual rows; `.gitkeep` is explicitly non-runtime. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps routes, view/schema, WS, revisions, projections, auth, providers, frontend, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 supplies predecessor/successor relationships and handoff checkpoints. |
| Every proposed slice is behavior-preserving unless explicitly marked otherwise. | PASS | Section 7 declares behavior preservation and later authorization for any behavior change. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 contains Make-owned commands and conditional scopes; no command is invented. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Section 1 requires the exact marker if one appears later. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Section 1 records the mixed live package and stale archived completion state. |
| Handoff sections are current enough to continue without rediscovery. | PASS | Section 10 records all required workstream handoffs, inspections, results, blockers, and next actions. |
| Former design blockers have one consistent closure posture. | PASS | RB-001, RB-002, RB-004, and RB-005 are planning-resolved; their implementation rows remain `TODO`. |
| Verification uncertainty is neither guessed nor hidden. | PASS | RB-003 is decision-resolved/evidence-pending, with an exact artifact schema and ten live unselected identities. |
| Tracker planning evidence is not treated as owner authority. | PASS | SL-00A and TL-016 require Core/appendix/machine promotion before production implementation. |

This tracker is complete as a planning artifact. It does not authorize or claim
completion of any production refactor.
