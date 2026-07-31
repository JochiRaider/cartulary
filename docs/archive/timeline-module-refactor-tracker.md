# timeline Module Refactoring Tracker and Handoff

> **Historical handoff:** This tracker records the remediation completed at
> commit `44c2ca0e`. It is not current architecture or executable authority.
> Current implementation work is tracked in
> `docs/handoffs/timeline-production-readiness-refactor-tracker.md`; executable
> requirements, schemas, manifests, catalogs, and harness inputs under
> `contracts/` and `tools/` remain authoritative.

## 1. Scope and Source Posture

- **Target path:** `internal/modules/timeline`
- **Derived target label:** `timeline`
- **Output path:** `docs/handoffs/timeline-module-refactor-tracker.md`
- **Status:** Authorized remediation complete; `SL-00A` through `SL-08` are
  complete and all binary exit criteria pass.
- **Authorized scope:** Execute `SL-00A` through `SL-08` sequentially across
  authored specification, contract, verification, implementation, migration,
  test, documentation, and application-assembly inputs. Generated artifacts may
  change only through their owning generators.
- **Public compatibility boundary:** Preserve the established HTTP, OpenAPI,
  `cartulary.view.timeline.v2`, saved-view, and WebSocket v1 contracts. Internal
  Go APIs and temporary compatibility seams have no independent compatibility
  guarantee and must be removed after repository callers migrate.
- **Non-goals:** Do not add implementation-support vocabulary to
  `docs/domain.md`, backfill collaboration history, preserve obsolete internal
  shims, or change public wire behavior.
- **Execution gate:** After each slice, update this tracker and append handoff
  evidence, then run `make lint-markdown` and `git diff --check`. A failed gate
  marks the slice `BLOCKED` and prevents dependent work from starting.

The label is the lowercase kebab-case basename of the target path. It contains no
spaces, separators, shell metacharacters, or unsafe filename characters.

### Source hierarchy

1. Executable authority lives in the requirements registry and owner catalogs
   under `contracts/requirements`, typed contracts under `contracts/**`, the
   verification catalogs under `contracts/verification`, and the machine-owned
   harness inputs under `tools/`.
2. No adopted Timeline-specific subsystem NLSpec was found. An adopted NLSpec for
   another subsystem does not govern the Timeline workbook projection.
3. Core 00 through Core 04, `docs/domain.md`, and implementation-support guidance
   explain vocabulary and design direction for people; they are not executable
   authority and no product check may depend on them.
4. Core 05 is relevant only to claim-bearing timed or fixture-sensitive
   publication. This remediation makes no such claim.
5. Current repository code and tests establish current implementation state.
6. This tracker controls sequencing and records evidence, but does not supersede
   a machine owner contract.

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
- an ephemeral `temp/analysis-notes.md` planning input that is no longer retained

### Repository files inspected

Every file under `internal/modules/timeline` was inspected and is inventoried in
Section 2. Searches and exact-source reads also covered the inbound application
assembly, workbook, imports, incident-bundle, recovery, projections, entities,
mentions, merge, reporting, revisions, saved-view, collaboration, frontend, grid
contract, generated protocol, view-schema, and verification-catalog surfaces
identified by live imports or contract identifiers.

### Planning findings relative to prior guidance

- The framework describes a desired narrow Timeline owner for capture and mutation
  semantics, but the baseline 54-file target also contained transport, persistence,
  projection, workbook/import, cross-module effect, provider, and test-support
  responsibilities. The framework is therefore direction, not proof that the
  current package is already the correct permanent boundary.
- The archived June 2026 tracker records useful completed remediation, including a
  command façade and private store, but predates current subpackages such as
  `rowsnapshot`, `mentioneffects`, and `workbookprojection`. Its completion state
  is not current evidence and is not copied into this tracker.
- An ephemeral analysis note supplied closure guidance for the
  `2026-07-25T23:18:54-04:00` session but was not retained as repository
  guidance. The dated decision log below preserves the relevant forensic fact;
  executable owner authority remains the only current behavioral source.

### Closure posture

RB-001 through RB-005 are closed. SL-00 closed RB-003 by making the active
owner catalogs the exhaustive exact-selector disposition, assigning all retained
Timeline and Projections identities exactly once, and replacing five redundant
Timeline source-scanning tests with equivalent machine boundary rules. Desired
state characterization that cannot pass against the legacy implementation is
owned by its implementation slice and must be added before that slice changes
production behavior. SL-06 completed the RB-005 implementation by replacing
route publication and in-memory sequencing/replay with transaction-coupled
durable intents and a Collaboration-owned sequencer.

Final discovery found 62 live files under `internal/modules/timeline`, including
the non-runtime `.gitkeep`, and nine retired baseline paths recorded below. Every
live file is inventoried. The final tree has one source snapshot/derivation path,
no Timeline-owned projection SQL, no route publisher or in-memory replay store,
and no temporary forwarding shim.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/timeline/.gitkeep` | Directory placeholder | None | None | None | None | None | Timeline repository layout | low | Explicitly out of runtime scope because it contains no behavior. |
| `internal/modules/timeline/api.go` | Retired mixed contract/decoder/hash/payload/presentation unit | None; file moved and partitioned in SL-01 | None | None | Decoder, hash, payload, error, projection-contract, and frontend payload rows | Timeline view schema and public payload contracts remain unchanged | Timeline application façade | low | Replaced by `contracts.go`, `commands.go`, `results.go`, `request_decoding.go`, `request_hashing.go`, `payloads.go`, and `presentation.go` without forwarding shims. |
| `internal/modules/timeline/api_errors.go` | Converts stable mutation failures to HTTP API errors | `MutationAPIErrorContext`, `ClassifyMutationAPIError` | Workbook and Timeline route coordination | HTTP API error package; Timeline error types | `error_classifier_test.go`, route envelope tests | Public error-envelope contract | Timeline application adapter | medium | Error codes and authorization visibility are observable. |
| `internal/modules/timeline/auto_resolution.go` | Alias eligibility and exact/suppressor auto-resolution policy over typed peer results | Private helper surface | Timeline create, patch, and paste stores | Entities/Mentions and Links ports | `request_test.go`, `resolution_integration_test.go` | Mention and link contract | Timeline orchestration using peer-owner ports | medium | Interactive-only eligibility and provenance remain exact without peer SQL. |
| `internal/modules/timeline/boundary_guard_test.go` | Retired in SL-00 after parity moved to machine boundary policy | None; file deleted | None | None | `make backend-module-boundary-check` | `tools/backend_module_boundaries.json` | Harness-owned boundary policy | low | The five redundant AST/source tests were deleted only after their import, façade, private-store, parser-dependency, and error-classification rules passed in the machine checker. |
| `internal/modules/timeline/bulk_mutation.go` | Retired bulk-through-clipboard compatibility path | None; file deleted in SL-05 | None | None | Workbook bulk command tests | Public Workbook bulk contract remains unchanged | Workbook admission plus Timeline command handlers | low | Workbook now dispatches exact fill-down and tag-assignment commands; no clipboard text or paste request is synthesized. |
| `internal/modules/timeline/clipboard_paste.go` | Converts validated `tabular_row_plan_v1` inputs into Timeline-owned source changes and raw-capture provenance | Private owner-row builders plus `ClipboardRawImportColumn` retained for import callers | Timeline façade and import source mapping | Versioned Tabular Ingest plan, Timeline field normalization and collection policy | `clipboard_paste_test.go`, workbook clipboard tests | Clipboard route contract; Timeline source fields | Timeline target application | medium | Parsing, public request admission, exact-header recognition, and plan construction are Workbook/Tabular Ingest-owned; only clipboard carries a tabular plan. |
| `internal/modules/timeline/clipboard_paste_store.go` | Transactional owner-batch application with source changes, peer effects, projection mutation, history, idempotency, and event intents | Private store methods through `Facade` | `facade.go`, Workbook mutation flow | Timeline repository and caller-transaction peer ports | `clipboard_paste_test.go`, Workbook and Timeline integration tests | Timeline mutation, projection, revision, collection, and collaboration contracts | Timeline mutation coordinator | high | Target preflight, atomic batch semantics, replay, and side-effect order are observable. |
| `internal/modules/timeline/clipboard_paste_test.go` | Characterizes plan mapping, raw capture, binding, and cross-incident rejection | Test functions only | Test runner | Clipboard plan and Timeline application helpers | This file | Clipboard and mention-binding contracts | Timeline/Workbook test support | medium | All executable identities, including cross-incident rejection, select exactly once. |
| `internal/modules/timeline/collection_descriptors.go` | Maps field keys to mention, record-reference, and tag collection policy | `CollectionFamily`, `CollectionPolicy`, `LookupCollectionPolicy` | Timeline mutation and collection stores | Timeline field vocabulary | Unit and integration collection tests | Timeline view-field and collection-action contract | Timeline | medium | Legitimate Timeline-owned semantic policy. |
| `internal/modules/timeline/commands.go` | Declares cohesive Timeline command and owner-batch application inputs | Command DTOs and operation discriminators | Workbook, imports, recovery, façade | Owner-neutral contract types | Decoder, hash, batch, and integration tests | HTTP behavior mapped without exporting transport concerns | Timeline application façade | medium | Distinct clipboard, fill-down, and tag-assignment commands have no forwarding aliases. |
| `internal/modules/timeline/contracts.go` | Declares stable Timeline request and source contract values | Request DTOs and closed values | Workbook and Timeline application callers | UUID and field contract types | Decoder and contract tests | HTTP/OpenAPI and Timeline view contracts | Timeline contract surface | high | Partitioned from the retired mixed API file without wire change. |
| `internal/modules/timeline/decoder_test.go` | Characterizes patch validation and visible text fields | Test functions only | Test runner | `api.go` decoders and field normalization | This file | Timeline request and visible-source-text contract | Timeline test support | medium | Protects exact request validation. |
| `internal/modules/timeline/deleterestore/provider.go` | Supplies Timeline tables to generic delete/restore coordination | `NewProvider` | Timeline revision contribution | Records delete/restore provider API | Revision and recovery integration tests | Revision/delete-restore contract | Timeline source-owner adapter | medium | Keep narrow unless domain behavior appears. |
| `internal/modules/timeline/derivation.go` | Composes the canonical Timeline source snapshot and semantic row derivation | Private derivation service | Store, effects, snapshots, providers | Source repository, peer batch reads, `workbookprojection.Derive` | Equivalence and service-backed tests | Projection input and public row contracts | Timeline source semantics | high | The only root orchestration path for source-to-row derivation. |
| `internal/modules/timeline/derivation_equivalence_test.go` | Proves canonical JSON equivalence across former row-producing paths | Test functions only | Test runner | Canonical repository and derivation | This file | Timeline row/projection contract | Timeline product evidence | high | Covers nullable values, time pairs, collections, provenance, evidence, replacements, and grouping. |
| `internal/modules/timeline/error_classifier_test.go` | Characterizes mutation error to HTTP envelope mapping | Test function only | Test runner | `api_errors.go` | This file | Public error-envelope contract | Timeline test support | medium | Required if transport/error code is split. |
| `internal/modules/timeline/facade.go` | Stable application command façade over the private store with exact clipboard, fill-down, and tag-assignment entry points | `Facade`, command DTOs, `NewFacade` | Workbook, imports, recovery, routes, test assembly | Private Timeline store and query/profile services | Boundary guards, store and integration tests | Timeline command/query behavior | Timeline application layer | medium | Bulk commands build owner-native rows directly and share only the transactional batch engine; obsolete bulk/paste forwarding surfaces were removed. |
| `internal/modules/timeline/hooks.go` | Retired production test-hook and module-override surface | None; file deleted in SL-04 | None | None | Port fakes and application-owned harness routes | Test-only composition | Application/test support | low | Caller-transaction projection failures now use `timeline/testsupport/fakeports`; the server profile registers application-owned test routes without a production Timeline test constructor. |
| `internal/modules/timeline/incident_bundle_portability.go` | Exports and imports raw Timeline profile/event source families | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx` | Incident-bundle source/routes | Incident portability, direct SQL, Timeline source tables | Incident-bundle integration tests | Bundle portability contract | Timeline source-owner adapter | high | Projections and runtime state must remain excluded. |
| `internal/modules/timeline/lifecycle_store.go` | Implements review and supersede mutations, history, links, projection refresh, and replay | Private store methods through `Facade` | Timeline and workbook route coordination | Direct SQL; links, revisions, projections, idempotency | `store_test.go`, `timeline_event_integration_test.go` | Lifecycle, supersede, revision, row-version contract | Timeline mutation coordinator | high | Reviewer/admin authorization and terminal superseded state are frozen. |
| `internal/modules/timeline/linkeffects/linkeffects.go` | Retired duplicate link-effect adapter | None; file deleted in SL-04 | None | None | Entity merge and Timeline integration tests | Link-derived Timeline field contract | Timeline mention-effect provider | low | Relationship invalidation mapping is now part of the injected Timeline source-owner effect provider. |
| `internal/modules/timeline/mentioneffects/mentioneffects.go` | Applies Timeline-owned mention/link consequences through injected peer read and projection ports | `Provider`, typed source/invalidation requests and results | Application-injected Entities mention and merge stores | Timeline source SQL plus Records, collection-read, and Projection ports | Entity resolution/merge and Timeline tests | Mention provenance and projection contracts | Timeline source-owner effect adapter | medium | It owns only Timeline source updates and deterministic invalidation mapping; no concrete peer store or Projections runtime construction remains. |
| `internal/modules/timeline/mentions_collections_store.go` | Coordinates Timeline collection actions through injected mention, entity, link, tag, evidence, and collection-read ports | Private store methods | Main store and paste store | Timeline-owned policy plus peer-owner transaction ports | `unit_test.go`, `resolution_integration_test.go`, integration suites | Collection action, provenance, evidence-count, projection contracts | Timeline mutation coordinator | medium | Evidence attachment validation and peer writes stay inside the initiating transaction; peer storage models do not cross the port boundary. |
| `internal/modules/timeline/payloads.go` | Builds stable mutation, history, and collaboration payloads | Payload helpers | Timeline façade/store and intent builder | Timeline result contracts | Payload, integration, and WS tests | HTTP and WebSocket v1 payload contracts | Timeline application contract | high | Serialization remains byte-compatible where observable. |
| `internal/modules/timeline/ports.go` | Declares Timeline-owned transaction-bound dependencies and typed peer results | `Dependencies` and narrow Records, Revisions, Entities/Mentions, Links/Tags, Evidence, Idempotency, Projections, Incident, and collection-read ports | Timeline façade/store and application assembly | Owner-neutral types, caller `pgx.Tx`, auth primitives, canonical derivation contract | Store, integration, boundary, and port-fake tests | Transaction, provenance, projection, and history contracts | Timeline application port surface | medium | Concrete construction moved to `internal/app/timelineassembly`; Timeline production code no longer imports Projections runtime packages or exposes test-only constructors. |
| `internal/modules/timeline/presentation.go` | Adapts canonical presenter records to stable Timeline rows | Presentation helpers | Façade and query results | Pure `rowpresenter` leaf | Projection, frontend, and integration tests | `cartulary.view.timeline.v2` | Timeline presentation adapter | high | Contains no persistence or policy. |
| `internal/modules/timeline/projection_contract_test.go` | Guards Timeline projection field and row-shape expectations | Test function only | Test runner | Projection contract helpers | This file | Timeline view/projection contract | Timeline test support | medium | Characterization must survive projection-seam changes. |
| `internal/modules/timeline/projection_mutation.go` | Builds closed upsert/delete mutations from canonical Timeline source state | `ProjectionMutation` and builder surface | Timeline stores, effects, Projections adapters | Canonical derivation and injected writer port | Projection, rollback, rebuild, and integration tests | Timeline projection source contract | Timeline projection semantics | high | Ineligible or malformed state fails closed; storage remains Projections-owned. |
| `internal/modules/timeline/projection_provider.go` | Publishes Timeline query descriptor and source-provider capabilities | Provider descriptor surface | Application and Projections assembly | Approved read-shape and provider contract types | Provider manifest and boundary tests | Projection-provider contract | Timeline source descriptor | medium | Contains no projection-table SQL or runtime adapter construction. |
| `internal/modules/timeline/query_projection_paging_test.go` | Retired Timeline-owned physical keyset SQL test | None; moved to Projections in SL-03 | None | None | Projections keyset query row | Query pagination contract | Projections product evidence | low | The moved test selects exactly once under `module.projections`. |
| `internal/modules/timeline/query_projection_store.go` | Retired Timeline-owned projection storage/query path | None; file deleted in SL-03 | None | None | Timeline and Projections query parity suites | Timeline query contract remains unchanged | Projections persistence/query engine | low | Physical SQL, filtering, paging, hydration, and storage now belong to Projections; no dual path remains. |
| `internal/modules/timeline/query_schema_guard_test.go` | Retired Timeline-local physical query schema guard | None; moved to Projections in SL-03 | None | None | Projections schema guard row | `cartulary.view.timeline.v2` | Projections support evidence | low | Descriptor/schema parity remains selected under the new owner. |
| `internal/modules/timeline/record_change_intents.go` | Converts committed Timeline semantic effects into deterministic Collaboration intents | Intent construction and append helpers | Create, patch, lifecycle, batch, rollback, restore, and peer-effect paths | Collaboration intent port and canonical payloads | Idempotency, WS, restart, and fault tests | `record_change_intent_v1` and WS v1 behavior | Timeline intent producer | high | Exact replay appends none; public identity, sequence, replay, and delivery stay Collaboration-owned. |
| `internal/modules/timeline/reportingprovider/provider.go` | Collects Timeline fields and facts for reporting export | `CollectFieldsTx`, `CollectFactsTx` | Reporting export materializer | Direct SQL, reporting provider types | Reporting boundary and export tests | Reporting source-provider contract | Timeline source-owner adapter | medium | Retain as a narrow adapter unless semantics leak inward. |
| `internal/modules/timeline/request_decoding.go` | Decodes and validates Timeline requests independently of policy and persistence | Decoder helpers | Workbook and Timeline routes | Contract declarations and field normalization | Decoder and route tests | HTTP/OpenAPI request contract | Timeline transport contract | high | No hashing, payload, presentation, or storage concern remains in this file. |
| `internal/modules/timeline/request_hashing.go` | Produces normalized stable idempotency hashes | Hash helpers | Workbook/Timeline admission and façade | Canonical encoding and contract declarations | Fixed-vector and replay tests | Public idempotency behavior | Timeline application contract | high | A fixed vector guards byte stability. |
| `internal/modules/timeline/request_test.go` | Characterizes auto-resolution eligibility and manual confidence | Test functions only | Test runner | API and resolution helpers | This file | Mention-resolution contract | Timeline test support | medium | Protects manual versus auto-match provenance. |
| `internal/modules/timeline/resolution_integration_test.go` | Exercises auto-resolution eligibility, suppressors, and manual confidence against storage | Test functions only | Service-backed test runner | Timeline service, entities, aliases, links, Postgres | This file | Entity mention/link provenance contract | Timeline integration evidence | high | Service-backed characterization. |
| `internal/modules/timeline/results.go` | Declares cohesive Timeline application result and conflict values | Result DTOs and errors | Workbook, routes, providers, tests | Owner-neutral values | Error, payload, replay, and integration tests | Public result/envelope mapping | Timeline application façade | high | Storage models do not cross this surface. |
| `internal/modules/timeline/revision_provider_contribution.go` | Contributes Timeline delete/restore and rollback providers to revision assembly | `RevisionProviderContribution` | `internal/app/revisionassembly` | Revisions, delete/restore and rollback providers | Revision integration tests | Revision provider catalog | Timeline source-owner adapter | medium | Legitimate source-owner assembly contribution. |
| `internal/modules/timeline/rollbackprovider/provider.go` | Restores and touches Timeline source state during rollback | `TimelineProvider`, `NewTimelineProvider` | Timeline revision contribution | Revisions and direct SQL | `rollbackprovider/provider_test.go`, revision integration tests | Rollback/source-history contract | Timeline source-owner adapter | high | Row version and projection follow-up must remain consistent. |
| `internal/modules/timeline/rollbackprovider/provider_test.go` | Characterizes Timeline rollback value mapping | Test function only | Test runner | Rollback provider | This file | Rollback source-value contract | Timeline evidence with Revisions collaborator | medium | The executable identity selects exactly once under Timeline. |
| `internal/modules/timeline/routes.go` | Registers Timeline-owned time-profile and lifecycle HTTP routes with admission checks | `Service`, `RegisterRoutes` | Server runtime | HTTP platform, auth, Timeline façade | Support and integration route tests | Time-profile, mark-reviewed, and HTTP envelope contracts | Timeline transport adapter | medium | Routes do not publish, sequence, retain events, or register production test hooks. |
| `internal/modules/timeline/rowpresenter/rowpresenter.go` | Builds the stable Timeline row wire shape from a normalized record | `Record`, `BuildRow` | Root API and row snapshot | Field normalization and standard library | Projection/schema/integration tests | Timeline view row contract | Timeline presentation | high | Pure leaf, but parallel derivation paths feed it. |
| `internal/modules/timeline/rowsnapshot/rowsnapshot.go` | Retired duplicate source snapshot and row hydration path | None; file deleted in SL-02 | None | None | Canonical equivalence and effect tests | Timeline source/row/projection contract | Canonical Timeline derivation | low | All callers migrated before deletion; no parallel loader remains. |
| `internal/modules/timeline/source_repository.go` | Adapts the private source repository into canonical snapshot inputs and batch peer hydration | Private repository façade | Canonical derivation and mutation paths | Timeline source persistence plus Records/collection read ports | Equivalence, store, and service-backed tests | Timeline source and projection input contracts | Timeline persistence/application seam | high | Reads Timeline tables directly and peer state only through approved ports. |
| `internal/modules/timeline/sourcerepository/repository.go` | Owns Timeline event, raw-capture, and time-profile source SQL | Repository types and transaction methods | Root source repository and narrow providers | SQLC/Postgres for Timeline-owned objects | Store, portability, time, and integration tests | Timeline source storage contract | Timeline private persistence | high | It does not access projection or peer-owner tables. |
| `internal/modules/timeline/state.go` | Defines capture-state vocabulary, transition policy, and supersede validation | State error and transition/validation helpers | Timeline stores and API tests | UUID and Timeline request types | `store_test.go`, `support_test.go`, integration tests | Rough/enriched/reviewed/superseded lifecycle contract | Timeline | high | Legitimate domain policy with externally visible results. |
| `internal/modules/timeline/store.go` | Timeline transaction coordinator for create, import, patch, conflict, lifecycle, history, and idempotency outcomes | Conflict errors/claims, mutation results, time profile, conflict parser; private store | `facade.go` and application composition | Timeline source repository plus injected peer-owner ports | `store_test.go`, `timeline_event_integration_test.go`, support and workbook tests | Mutation, history, conflict, lifecycle, projection and auth contracts | Timeline mutation application | medium | The initiating Timeline operation owns begin/commit; every peer receives its transaction and cannot commit, publish, or start nested work. |
| `internal/modules/timeline/store_test.go` | Unit-characterizes identity, state, replay, concurrency, history, rollback, closed incidents, and evidence sort | Test functions and shared `BaseTime` | Test runner and package tests | Private store/fakes | This file | Timeline mutation/query/history contracts | Timeline test support | high | Essential characterization baseline. |
| `internal/modules/timeline/support_integration_test.go` | Exercises the incident-role authorization matrix | Test function only | Service-backed test runner | Timeline routes/runtime | This file | Core 04 incident authorization | Timeline security evidence | high | Authorization must be re-derived at route time. |
| `internal/modules/timeline/support_test.go` | Unit-characterizes create coverage, state helpers, hashes, payloads, and supersede guards | Test functions only | Test runner | API/state helpers | This file | Timeline request/payload contract | Timeline test support | medium | Protects stable payload shapes. |
| `internal/modules/timeline/testsupport/asserttest/assertions.go` | Reusable Timeline database, revision, projection, and WebSocket assertions | Assertion DTOs and functions | Timeline and related module tests | SQL/Postgres, platform WebSocket | Tests that import this helper | Test evidence shapes only | Timeline-specific test support | low | Keep local unless a genuinely cross-module abstraction emerges. |
| `internal/modules/timeline/testsupport/fakeports/projection.go` | Supplies transaction-aware projection fakes without production hooks | Test-only fake writer/query behavior | Timeline package and application harness tests | Timeline port contracts | Failure, rollback, and replay tests | Test evidence only | Timeline test support | low | Replaces production dependency override keys and constructors. |
| `internal/modules/timeline/testsupport/routetest/inventory.go` | Supplies route-inventory control entries for Timeline test scenarios | `ControlQuery`, `ControlCreateAndLive` | Timeline and incident route tests | Route inventory helpers | Route/inventory tests | Public route inventory | Timeline-specific test support | medium | Test accounting helper, not runtime architecture. |
| `internal/modules/timeline/testsupport/routetest/timeline.go` | Issues Timeline create requests in route tests | `CreateRow` | Timeline and cross-module route tests | HTTP test server and auth fixtures | Route tests | Timeline create route contract | Timeline-specific test support | low | Test client only. |
| `internal/modules/timeline/testsupport/scenariotest/harness.go` | Starts Timeline runtime/server integration harnesses | `RuntimeHarness`, server alias, `StartRuntime` | Timeline integration tests | Shared application test support | Integration tests | Test runtime composition | Timeline-specific test support | medium | Harness topology is verification support, not runtime ownership. |
| `internal/modules/timeline/testsupport/scenariotest/runtime.go` | Timeline scenario helpers for create/query, row lookup, change/revision assertions | Scenario helper functions | Timeline and related integration tests | HTTP test server, database, WebSocket assertions | Integration tests | Timeline route, history, and row contracts | Timeline-specific test support | medium | Preserve semantic assertions during refactor. |
| `internal/modules/timeline/testsupport/scenariotest/ws.go` | Connects to incident WebSocket and asserts Timeline change messages | Socket payload/client and assertion helpers | Timeline collaboration tests | WebSocket test client | WS integration tests | `/ws/v1/incidents/{incident_id}` and `record_changed` | Timeline-specific test support | medium | Event path and payload are observable. |
| `internal/modules/timeline/testsupport/storetest/harness.go` | Starts Postgres-backed Timeline store harnesses | `StoreHarness`, `StartStore` | Store and provider tests | Shared Postgres test support | Store tests | Test persistence setup | Timeline-specific test support | low | Keep local while Timeline-specific. |
| `internal/modules/timeline/time_conversion_store.go` | Reads/writes singleton incident time profile and derives UTC/local pairs | Private store methods through `Facade` | Timeline routes and row derivation | Direct SQL, `timecontract` | `timecontract_test.go`, Timeline profile integration test | Time-conversion profile and dual-field contract | Timeline | high | Profile auth and exact conversion behavior are public. |
| `internal/modules/timeline/timecontract/timecontract.go` | Parses and formats Timeline UTC/local time values with fixed offsets | Parse/format functions and `LocalParseResult` | Timeline store, row snapshot, workbook projection | Standard time/string packages | `timecontract/timecontract_test.go`, integration tests | Timeline time-pair contract | Timeline | medium | Current live callers are Timeline-local; no move is supported by evidence. |
| `internal/modules/timeline/timecontract/timecontract_test.go` | Characterizes Timeline time parsing and formatting | Test function only | Test runner | `timecontract` | This file | Timeline time-pair contract | Timeline test support | medium | The executable identity selects exactly once under Timeline. |
| `internal/modules/timeline/timeline_event_integration_test.go` | End-to-end backend coverage for create, patch, replay, rollback, conflicts, envelopes, rough capture, projections, auth, lifecycle, WS, and time profile | Test functions only | Service-backed test runner | Full server/runtime/Postgres test support | This file | Most observable Timeline backend contracts | Timeline integration evidence | high | Primary broad characterization source. |
| `internal/modules/timeline/timing.go` | Carries create-timing instrumentation through context | Timing recorder interfaces and context helper | Workbook telemetry and Timeline create flow | Context/time only | Measurement and workbook tests | Measurement support; no product behavior | Workbook/Timeline instrumentation seam | low | Implementation-support evidence, not Core 05 publication. |
| `internal/modules/timeline/unit_assertions_test.go` | Shared package-local Timeline unit assertions/fakes | Test-only private helpers | Timeline unit tests | Testing package and Timeline types | Package tests | Test support only | Timeline test support | low | No production surface. |
| `internal/modules/timeline/unit_test.go` | Unit-characterizes binding modes, duplicate mention provenance, and attached evidence | Test functions only | Test runner | Timeline request/store logic | This file | Mention and evidence collection contracts | Timeline evidence with Evidence collaboration | high | Every identity, including attached evidence, selects exactly once. |
| `internal/modules/timeline/workbookprojection/derivation.go` | Purely derives the preserved `ProjectionInput` and presenter record from a canonical source snapshot | Derivation inputs/results | Root canonical derivation | `timecontract`, row presenter, contract values | Equivalence, projection, and rebuild tests | Timeline projection and row contracts | Timeline source semantics | high | One deterministic derivation serves live mutation, query, effect, and rebuild. |
| `internal/modules/timeline/workbookprojection/query_surfaces.go` | Declares Timeline-owned query descriptor/filter/sort/group semantics | Descriptor types and validation | Workbook query façade and Projections query engine | View contract types only | Query/schema/filter/sort/group tests | `cartulary.view.timeline.v2` | Timeline query semantics | high | Physical SQL compilation, cursor execution, and storage remain Projections-owned. |
| `internal/modules/timeline/workbookprojection/store.go` | Enumerates canonical Timeline projection inputs in deterministic keyset order | Existing `ProjectionInput`, `ListProjectionInputsTx` | Projections rebuild/provider registry | Timeline source repository and canonical derivation | Projection rebuild/integration tests | Timeline projection derivation contract | Timeline source provider consumed by Projections | high | Preserves the exact DTO and uses no `OFFSET` or projection-table SQL. |

## 3. Module Boundary Diagnosis

The baseline target was a **mixed-responsibility package**. It contained a
legitimate Timeline source-record/application core, but it was also a mutation
coordinator, view/projection orchestration layer, transport-adjacent adapter,
persistence-adjacent adapter, workbook/import target, and home for several
cross-module provider and effect adapters. The table records that baseline
diagnosis. SL-01 through SL-07 implemented the resolved owner allocation below;
the final boundary and broad validation evidence is recorded in Sections 10 and
12.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Capture, lifecycle, source-record semantics | Root state, API, and store files | Timeline | keep | Core 01–03; active `module.timeline` requirement; store tests | Central legitimate module responsibility. |
| Stable application command façade | `facade.go` | Timeline | keep | Live workbook/import/recovery callers; boundary guards | Preserve command behavior and public shape while internals move. |
| Wire decoding and HTTP error mapping | `api.go`, `api_errors.go`, `routes.go` | Timeline transport adapter plus workbook route owner | split | Direct route callers and public envelopes | Split by layer without changing route or payload shapes. |
| Source snapshot, hydration, and row presentation | `query_projection_store.go`, `rowsnapshot`, `rowpresenter`, collection store | Timeline | split | Duplicate SQL/derivation paths and shared row contract | Create one Timeline-owned semantic derivation path. |
| Physical workbook projection persistence/rebuild/query | Query store, `workbookprojection`, mention effects, projection adapters | Projections with Timeline source/descriptor ports | move/split | Projections imports Timeline inputs while Timeline imports a Projections adapter and queries projection storage | RB-002 resolves the seam: Projections owns physical SQL/lifecycle; mutation initiator owns the transaction. |
| Clipboard paste and bulk command semantics | Workbook `batch_admission.go`, Tabular Ingest planning, Timeline clipboard conversion and batch store | Workbook admission, Tabular Ingest planning, Timeline application | split complete | Workbook owns public decoding/admission and exact targets; Tabular Ingest owns the validated versioned plan; Timeline owns exact command application | SL-05 removed bulk-through-paste reuse while preserving public envelopes, hashes, ordered targets, conflicts, and replay. |
| Mention/entity/link/evidence effects | Collection store, `auto_resolution.go`, `mentioneffects`, `linkeffects` | Respective peer owner plus narrow Timeline effect ports | split/defer | Live peer-module imports and direct SQL | Preserve provenance, invalidations, and transactional behavior. |
| Revision, delete/restore, rollback | Revision contribution and provider subpackages | Timeline source-owner adapters | keep | Revision assembly imports the contribution | Keep narrow; generic coordination remains Revisions-owned. |
| Incident bundle and reporting inputs | Portability and reporting providers | Timeline source-owner adapters | keep | Incident bundles and reporting call them directly | Export raw source facts only; exclude projections/runtime state. |
| Collaboration publication | Baseline Timeline routes and in-memory hub; final Timeline intent producer and Collaboration durable stream | Timeline intent producer plus Collaboration sequencer/delivery | move/split complete | Source-owner intent, Collaboration store/dispatcher, and ephemeral hub | SL-06 made intent insertion transactional and removed route publication, hub sequencing, and in-memory replay. |
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

### Internal interface records

These are internal contracts adopted by machine authority in SL-00A and
implemented where their owning slice is complete. They are not public wire
contracts. Existing public route, request, response, error, view-schema, and
WebSocket shapes remain unchanged.

#### Batch planning and owner application

`tabular_row_plan_v1` is the implemented versioned successor to the retired
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

Timeline and other replayable source owners use Collaboration's
`AppendIntentTx(transaction, Intent)` port. A `record_changed` intent carries a
deterministic key, incident/record/resulting row-version identity, canonical
payload, source change-set identity, and mutation ordinal where applicable. It
excludes public `event_id`, `stream_seq`, connection/subscriber/session data, and
delivery state. Job progress and admitted extension-resource changes use the
same intent/sequencer boundary rather than independent event counters.

The intent is inserted in the authoritative source transaction. Collaboration's
dispatcher claims committed intents with bounded leases, atomically advances one
per-incident cursor, writes one replay event under a unique intent key, and then
delivers the stored event. Restart, duplicate claim, replay, or broadcast retry
does not allocate another event or sequence. Replay is durable and ordered;
retention preserves both the newest 10,000 incident events and events newer than
five minutes. Resume tokens are opaque, stored only by hash, survive restart,
and remain bound to session, incident, client instance, and expiry. Delivery and
resume recheck current incident membership.

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
| Generic view query route | Workbook route owner with Timeline descriptor and Projections query façade | Core 01 route family; Workbook query store; Timeline descriptor; Projections engine | Paging, schema guard, store, integration, browser tests | Filter/sort/group/cursor and hidden collection cases | high | Physical projection storage remains invisible to callers. |
| Create and patch routes | Workbook route owner with Timeline mutation façade | Core 01 mutation routes; `api.go`, `store.go` | Decoder, store, integration, workbook tests | Blank/one-value capture, replay, closed incident, actor-scoped idempotency | high | Preserve envelopes and row versions. |
| Clipboard-paste route | Workbook admission, Tabular Ingest plan, Timeline application | Core 03 clipboard behavior; clipboard files | Timeline and workbook clipboard tests | Batch atomicity, raw unknown columns, mention origin, cross-incident rejection | high | RB-001 is resolved; existing public request/result envelopes remain unchanged. |
| Bulk-mutation route | Workbook admission with exact Timeline command handlers | Core 03 bulk behavior; Workbook batch admission and Timeline commands | Workbook bulk/route tests | `fill_down_v1`, tag assignment, stable record IDs, replay | high | Bulk no longer compiles through paste; vendor coordinates never enter the backend contract. |
| `tabular_row_plan_v1` and `owner_batch_apply_v1` | Tabular Ingest contract and Timeline source-owner applicator | Machine owner requirements; Tabular Ingest plan and Timeline façade/store | Clipboard/bulk unit and integration tests | Ordered mapping, raw preservation, exact operation dispatch, transaction ownership | high | Implemented internal-only boundary with validation, deterministic mapping fingerprints, closed warnings, and no public wire change. |
| Mark-reviewed and supersede routes | Timeline | Core 01/03; routes and lifecycle store | Store and Timeline integration tests | Role gates, terminal supersede, replacement links, replay/rollback | high | Lifecycle is observable source state. |
| Conflict resolution route and token | Workbook/Timeline application boundary | Core 03 OCC; conflict decoder/store | Store and integration conflict tests | Same-field conflict, stale/invalid token, resolution replay | high | Field-level OCC and error envelopes are frozen. |
| Timeline time-conversion profile GET/PUT | Timeline | Core 01 REQ-01-611 through REQ-01-613; route/store | Time contract and profile integration tests | Defaults, membership read, admin write, round-trip pairs | high | Exact auth and conversion fields must remain. |
| Incident WebSocket and `record_changed` | Timeline intent producer with Collaboration sequencing/delivery | Core 01 WS contract; durable stream store/dispatcher; scenario WS helpers | Canonical WS integration and Collaboration fault tests | Current membership, changed record/version, affected views, no-event/retry/restart failures | high | Public path remains `/ws/v1/incidents/{incident_id}`; durable intent and replay are internal. |
| Row version, idempotency, revisions, change sets, and rollback | Timeline plus Records/Revisions owners | Core 02/03; store and providers | Store, Timeline integration, revision tests | Exact replay, concurrent mutation, coupled supersede, rollback projection follow-up | high | Projection rows are not history. |
| Mention provenance and manual/auto resolution | Entities/Mentions/Links with Timeline source owner | Core 02/03; auto-resolution and collection code | Request, unit, resolution integration, entity tests | Interactive-only auto-match, suppressors, null manual confidence, repeated mentions | high | No implicit stub creation from Timeline capture. |
| Tags, typed links, record references, and attached evidence | Tags/Links/Evidence with Timeline collection policy | Core 02; collection descriptor/store | Unit, links/tags, store and integration tests | Add/remove, invalid targets, evidence counts, projection invalidation | high | Collection side effects must stay transactional. |
| Saved-view and view-schema compatibility | Saved Views and view-contract owners | Core 01/04; saved-view tests; view schema | Saved-view and frontend tests | Existing saved views against unchanged field IDs and query semantics | high | Saved-view scope does not alter incident row authorization. |
| Projection refresh, query, and deterministic rebuild | Projections with Timeline source/descriptor input | Core 01 §8; current adapter/provider registry; schema ownership manifest | Projection contract and Timeline integration tests | Commit/rollback refresh, keyset query, rebuild equivalence, invalidation coverage | high | Preserve `ProjectionInput`; projection-table SQL and storage lifecycle are Projections-owned. |
| Projection mutation/source/writer ports | Timeline contract with Projections implementation | Machine owner requirements, `workbookprojection`, injected application adapters | Projection, paging, rebuild, rollback, entity-effect tests | Upsert/delete, caller transaction, page determinism, fail-closed errors | high | Implemented without a parallel projection DTO or runtime dependency cycle. |
| `record_change_intent_v1` | Source-owner semantic intents with Collaboration durable delivery | Machine owner requirements, additive migration, store/dispatcher/runtime | Route, WS, replay, authorization and fault tests | Exactly-once intent/event posture, post-commit sequence, restart/retry | high | Implemented for every replayable producer; public event shape is unchanged. |
| Incident role authorization | Auth platform and route/application adapters | Core 04 REQ-04-021 through REQ-04-029 and REQ-04-127 | Authorization matrix and route integration tests | Membership loss, viewer/editor/reviewer/admin gates, hidden incident | high | `deployment_admin` is not an incident-access bypass. |
| Incident bundle portability | Incident Bundles with Timeline source provider | Core portability contract; provider code | Incident-bundle integration tests | Raw profile/events round trip, attribution, exclusion of projections | high | Source-owner validation remains required. |
| Reporting source fields/facts | Reporting with Timeline provider | Reporting provider and materializer | Reporting boundary/export tests | Stable support refs and source field/fact selection | medium | Narrow provider boundary. |
| Generated HTTP/protocol surface | Contract/codegen owners | OpenAPI/protocol inputs and generated consumers | Codegen boundary and HTTP conformance tests | Generated-drift checks if owner inputs change | high | Never hand-edit generated roots. |
| Frontend selectors and grid-adapter field IDs | Web and grid-adapter owners consuming Timeline contract | Live TypeScript searches and view schema | Frontend unit, browser, visual, accessibility rows | Selector/field compatibility only if backend row contract changes | medium | No direct grid-vendor import exists in this target. |
| Timeline/Projections harness accounting | Verification owners | Exact-selector Harness contract and active owner family manifests | Timeline, Projections, and collaborator owner slices | Exhaustive exact-selector disposition and selector validation | high | Every retained identity selects exactly once; `make harness-contract` and the final broad run pass. |

## 5. Coupling and Boundary Findings

This table preserves the baseline diagnosis and recommended structural remedy.
All `must_fix` and `should_fix` findings were completed by SL-01 through SL-07;
final discovery found no surviving duplicate path, unauthorized SQL/import
exception, production test hook, or route/in-memory publisher.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Timeline constructs projection adapters while Projections imports Timeline projection inputs | `ports.go`, `mentioneffects`, `workbookprojection`, projection store/provider registry | Cyclic ownership and hard-to-isolate rebuild semantics | `must_fix` | Projections owns storage/query/rebuild; Timeline owns source/descriptor input | Implement the RB-002 source/mutation/writer ports after owner-authority adoption. |
| Source loading, collection hydration, time derivation, and row construction are duplicated | Query store, `rowsnapshot`, collection store, row presenter, workbook projection | Drift between query, mutation effects, and rebuild rows | `must_fix` | Timeline | Define one source snapshot/row derivation service and characterize equivalence. |
| `api.go` combines wire decoding, hashes, policy, payloads, and presentation | Exact exported surface and callers | A small change can alter several public contracts | `should_fix` | Timeline application/transport adapters | Partition by concern while retaining compatibility exports until callers migrate. |
| Direct SQL is distributed across root and semantic provider subpackages | Store, auto-resolution, snapshot, time, portability, reporting, effect providers | Persistence details obscure semantic ownership and permits cross-owner writes | `must_fix` | Table/invariant owner under the RB-004 matrix | Inventory every statement, move projection/peer SQL to owner ports, and keep platform storage domain-neutral. |
| Test dependency overrides live in a production package | `hooks.go` and server profile harness callers | Test assumptions enlarge production surface | `should_fix` | Timeline test support/application assembly | Design a test composition seam without changing runtime ownership. |
| Large mutation coordinators hide revision, projection, mention, link, and evidence effects | Main, lifecycle, paste, and collection stores | Ordering or transaction regression during extraction | `should_fix` | Timeline coordinator using explicit owner ports | Characterize transaction and replay outcomes before extraction. |
| Workbook clipboard and bulk semantics were embedded in Timeline and bulk compiled through paste | SL-05 Workbook admission, validated `tabular_row_plan_v1`, and exact Timeline handlers | Duplicate target behavior and loss of exact command semantics | `fixed` | Workbook admission, Tabular Ingest plans, Timeline applicator/handlers | Public parity and owner evidence pass; `BulkMutationClipboardRequest`, Timeline HTTP decoders, and the unversioned plan are removed. |
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
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03 | Account for every target file, caller, dependency, and risk | All `internal/modules/timeline` files; inbound callers | Repository searches and exact-source review | All 62 final files and nine retired baseline paths have current-state rows. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-04 | Map routes, view/protocol fields, auth, projection, revision, and provider contracts | Contract owners, view schema, routes, façades, providers | `make explain-test-owner OWNER=module.timeline` | Every discovered contract has an owner and evidence posture. |
| WF-03 | Characterization and exact-selector accounting | parallel | WF-01 | WF-04, WF-07 | Produce `timeline_test_disposition_v1` for every executable identity and establish the baseline | Target tests, test support, all owner family manifests | Owner explanation plus exact selector comparison and validation | Every retained test is selected exactly once or has an approved action. |
| WF-04 | Boundary and coupling scan | chain | WF-02, WF-03 | WF-05 | Classify coupling and distinguish legitimate Timeline ownership from catch-all behavior | Target production files and peer-owner adapters | `make backend-module-boundary-check` during implementation | Findings have classifications and proposed owners. |
| WF-05 | Owner-authority promotion and interface design | chain | WF-04 | WF-06, WF-07 | Promote resolved tracker decisions into owning Core/appendix/machine inputs and define exact acyclic ports | Core 01/03, Appendix I, provider/boundary/schema-owner inputs, internal contract schemas | Owner review, Markdown/JSON/drift and boundary checks | RB-001, RB-002, RB-004, and RB-005 decisions are authoritative before production edits. |
| WF-06 | Slice sequencing | chain | WF-05 | WF-08 | Convert approved boundaries into smallest behavior-preserving changes | Packages named in Section 7 | Narrow owner targets after each slice | Every slice has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-05 | WF-08 | Apply only approved dispositions and close RB-003 evidence | Verification owner/catalog inputs and tests if later authorized | Exact-selector, overlap, catalog drift, and owner explanation targets | Every identity has one disposition; retained evidence resolves exactly once. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Run narrow-to-broad verification and refresh handoff evidence | Tracker and later authorized changed files | `make agent-finalize`, risk-appropriate `make test-fast` and `make check` | Successful commands, artifacts, failures, and remaining risks are recorded. |

## 7. Proposed Refactor Slice Plan

The present task authorizes all slices. They execute sequentially, and each slice
must pass its tracker gate before a dependent slice begins. Public behavior is
frozen unless an executable owner requirement and migration plan explicitly
authorizes a change.

### Slice execution and gap-remediation ledger

| Slice | Status | Dependency state | Gap-remediation state | Binary exit state |
| --- | --- | --- | --- | --- |
| SL-00A | DONE | Root prerequisite complete | Focused Timeline, Workbook, Tabular Ingest, and Collaboration requirements are active; `module.projections` owns physical projection behavior; provider, schema-owner, boundary, and human guidance are aligned. | PASS |
| SL-00 | DONE | SL-00A complete | Exact Timeline/Projections test disposition is catalog-owned; redundant AST tests are retired; the baseline and owner slices pass. Slice-specific desired-state characterization is an entry gate for SL-02, SL-03, SL-05, and SL-06. | PASS |
| SL-01 | DONE | SL-00 complete | Request contracts, application commands/results, decoding, normalized hashing, payload construction, error classification, and presentation now have cohesive files; no compatibility shim was introduced. | PASS |
| SL-02 | DONE | SL-01 complete | One Timeline source repository reads source state and obtains record envelopes through a batch-capable Records port; one canonical derivation produces `ProjectionInput` and presenter records; query, mutation, snapshot, mention-effect, and rebuild paths share it; duplicate loaders, types, collection hydration, and time derivation are removed. | PASS |
| SL-03 | DONE | SL-02 complete | Timeline owns closed upsert/delete source semantics and deterministic keyset enumeration; Projections owns physical query, writer, storage, rebuild, restore, and adapter behavior; application callers inject the writer/query ports; the legacy Timeline query path is removed. | PASS |
| SL-04 | DONE | SL-02 and SL-03 complete | Timeline declares narrow caller-transaction ports; peer owners expose typed transaction operations; application assembly constructs adapters; source-less projection construction was removed from Timeline, Entities, and Evidence effects; production test hooks and duplicate effect adapters are retired; machine SQL/import boundaries pass. | PASS |
| SL-05 | DONE | SL-00 and SL-04 complete | Workbook owns Timeline clipboard/bulk decoding, target admission, hashes, and envelopes; Tabular Ingest emits validated immutable `tabular_row_plan_v1` values; Timeline applies clipboard plans and exact fill/tag commands through one transaction engine without paste synthesis. | PASS |
| SL-06 | DONE | SL-01, SL-04, and SL-05 complete | Collaboration owns durable intents, per-incident sequencing, replay retention, opaque hashed resume state, retry/lease dispatch, and current-membership delivery. Timeline and all replayable producers append transactionally; routes and the hub no longer sequence or retain events. | PASS |
| SL-07 | DONE | SL-04 and SL-06 complete | Portability retains only Timeline source/profile state; reporting emits canonical source facts and references; delete/restore and rollback retain only source reconstruction under Revisions coordination. Exact provider import and SQL/read-shape rules prevent projection, replay, runtime, or peer-orchestration growth. | PASS |
| SL-08 | DONE | SL-03 through SL-07 complete | Final discovery removed the last hub-owned replay observation seam; all focused, contract, drift, boundary, browser, fast, and broad checks pass; migration, compatibility, and skipped-suite posture are recorded. | PASS |

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00A | Closure decisions in this tracker; later owner authorization | Promote RB-001/RB-002/RB-004/RB-005 into Core 01/03, Appendix I, provider/boundary/schema-owner inputs, and planned internal contract schemas before code movement | Owner documents and authored machine inputs only; generated outputs through generators | Tracker evidence must not silently become runtime authority | Owner contract, JSON-shape, drift, and boundary-policy checks | `make lint-markdown`; `make json-shape-check`; `make generate-drift`; `make backend-module-boundary-check` | Revert authored owner inputs and their generated refresh together | Core and machine owners agree on interfaces, table ownership, transactions, and event intent. |
| SL-00 | WF-03; RB-003 evidence completion | Produce `timeline_test_disposition_v1`, repair only proven selector gaps, and run the frozen characterization baseline | Timeline tests/testsupport; verification inputs only if separately authorized | Incomplete or duplicate accounting could hide regressions | Preserve retained tests; disposition the ten unselected identities; add tests only for demonstrated contract gaps | `make test-slice OWNER=module.timeline`; `make service-backed-test-slice OWNER=module.timeline` | Revert approved test/catalog changes without production impact | Every executable identity has one disposition and retained evidence selects exactly once. |
| SL-01 | SL-00, SL-00A | Partition request decoding, error mapping, row presentation, and command declarations while retaining compatibility exports | `api.go`, `api_errors.go`, `rowpresenter`, `facade.go`, workbook callers | Request validation, hashes, errors, row shape | Decoder, error classifier, payload, boundary and frontend contract tests | `make test-slice OWNER=module.timeline` | Revert leaf moves and compatibility forwarding together | No route/payload/hash/export behavior changes; mixed API file is reduced. |
| SL-02 | SL-01, SL-00A | Consolidate Timeline-owned source SQL, hydration, collection inputs, time derivation, and row construction behind private persistence plus one derivation service | Query store, `rowsnapshot`, row presenter, collection/time stores, mention effects | Query rows, evidence count, time pairs, mention/link fields | Projection contract, schema guard, paging, time, entity effects, integration query tests | `make service-backed-test-slice OWNER=module.timeline`; `make backend-module-boundary-check` | Canonical JSON parity was proven before duplicate loaders and derivation paths were removed atomically; no compatibility path remains | Query, effect, refresh, and rebuild paths derive byte-equivalent contract rows through the source repository and canonical derivation service. |
| SL-03 | SL-02, SL-00A | Implement Timeline source/mutation ports and the Projections-owned writer/query/rebuild seam while preserving `ProjectionInput` | Timeline ports/workbookprojection/mention effects; Projections query/store/providers | Transactional upsert/delete, bounded paging, deterministic rebuild, invalidations | Projection contract, paging, rebuild, rollback, restore and peer-effect tests | `make service-backed-test-slice OWNER=module.timeline`; `make backend-module-boundary-check`; provider manifest parity checks | Retain old adapter wiring until new ports pass parity; no source-data migration | Timeline imports no Projections runtime adapter, and projection storage/query behavior is Projections-owned and unchanged. |
| SL-04 | SL-02, SL-03, SL-00A | Narrow entity, mention, link, tag, evidence, revision, record, idempotency, and projection effects behind explicit caller-transaction ports | Collection store, auto-resolution, mention/link effects, ports, peer owner adapters | Provenance, link metadata, evidence count, history, transaction ordering | Resolution, attached evidence, links/tags, merge, rollback and replay tests | `make service-backed-test-slice OWNER=module.timeline`; peer-owner slices | Reconnect prior concrete adapters as one reversible change | Timeline writes only Timeline-owned source state and all cross-owner effects remain atomic. |
| SL-05 | SL-00, SL-00A | Implement Workbook admission, versioned Tabular Ingest plans, Timeline owner application, and distinct fill/tag bulk handlers | Clipboard/bulk files, Timeline façade/store, workbook routes/stores, shared Tabular Ingest | Batch atomicity, hashes, raw columns, stable IDs, mention origin, operation identity | Timeline/workbook clipboard and bulk unit/integration tests, including cross-incident targets and exact replay | `make service-backed-test-slice OWNER=module.timeline`; relevant Workbook/Tabular Ingest owner slices | Preserve current public façade/envelopes until all internal callers migrate; revert migration together | No bulk handler compiles through paste, one mapping engine exists, and public behavior is unchanged. |
| SL-06 | SL-01, SL-04, SL-00A; separate contract/migration authorization | Add durable `record_change_intent_v1`, Collaboration dispatcher/log sequencing, route decoupling, and test-only composition cleanup | Timeline routes/store/ports, Collaboration, application assembly, authored migration/contract inputs | Auth order, source atomicity, sequence/replay, retry/restart, session sliding, WS payload | Authorization, route envelope, canonical WS, no-event matrix, retry/restart, replay retention and membership-loss tests | `make service-backed-test-slice OWNER=module.timeline`; Collaboration owner slice; `make browser-e2e-webserver-backed` | Deploy durable intent path behind compatible assembly and revert application wiring/migration only under an approved rollback plan | Routes do not publish, committed changes yield exactly one durable/replayable event, and public WS behavior is unchanged. |
| SL-07 | SL-04 | Audit and retain narrow portability, reporting, delete/restore, and rollback providers | Provider subpackages, revision contribution, incident bundle/reporting callers | Bundle round trip, report facts, rollback row versions/projections | Incident-bundle, reporting, revision, rollback provider tests | Relevant owner slices plus `make backend-module-boundary-check` | Revert provider-by-provider; no data migration is expected | Providers contain source-owner adaptation only and generic coordination stays with peer owners. |
| SL-08 | SL-03, SL-04, SL-05, SL-06, SL-07 | Run narrow-to-broad verification, inspect drift, and update handoff | All later-authorized changed files and this tracker | Cross-owner and frontend contract regression | All preserved/added tests and browser rows selected by changed contracts | `make agent-finalize`; risk-appropriate `make test-fast` and `make check` | Roll back the first failing slice, not unrelated successful slices | Required targets pass or failures are documented with run roots and ownership. |

## 8. Validation Plan

Discovery commands passed during the planning sessions. SL-00A added executable
owner authority and ran its focused Projections owner slice without changing a
public wire contract. The initial tracker `make lint-markdown` run root was
`.cartulary/test-results/20260726T024332Z-p557264`; the closure-pass run also
passed with run root `.cartulary/test-results/20260726T032316Z-p570550`.
`git diff --check` and cumulative `git diff HEAD --check` passed.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Tracker Markdown | yes | Required for this planning-only change. |
| owner discovery | `make task-guide ROLE=module-author OWNER=module.timeline` | Current owner routing | yes | Passed during discovery. |
| owner accounting | `make explain-test-owner OWNER=module.timeline` | Active Timeline and collaborator owner families | yes | Final exact-selector disposition and all-owner comparison pass. |
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
| TL-002 | Inventory the baseline and final target files and live callers | WF-01 | DONE | TL-001 | Section 2 | All 62 final files and nine retired baseline paths have individual rows. |
| TL-003 | Map observable contracts and owners | WF-02 | DONE | TL-002 | Section 4 | Every discovered contract risk has an owner and evidence posture. |
| TL-004 | Record initial coupling and boundary diagnosis | WF-04 | DONE | TL-003 | Sections 3 and 5 | Findings use the required classifications and avoid permanent unsupported decisions. |
| TL-005 | Complete exact-selector disposition and accounting | WF-03, WF-07 | DONE | TL-002, TL-016 | Exact rows in the Timeline, Projections, and Evidence family manifests; machine boundary rules; SL-00 owner and harness evidence | Every retained executable identity selects exactly once; obsolete duplicate architecture tests are removed. |
| TL-006 | Approve the Timeline/Projections seam | WF-05 | DONE | TL-004 | RB-002 closure decision and Section 3 port/transaction record | Timeline owns source semantics; Projections owns physical query/storage/rebuild. |
| TL-007 | Approve clipboard/bulk ownership | WF-05 | DONE | TL-004 | RB-001 closure decision and Section 3 owner allocation | Workbook, Tabular Ingest, and Timeline responsibilities are explicit. |
| TL-008 | Approve persistence consolidation depth | WF-05 | DONE | TL-004 | RB-004 closure decision and SQL ownership matrix | SQL follows table/invariant ownership and platform storage stays neutral. |
| TL-009 | Characterize and consolidate row derivation | WF-03, WF-06 | DONE | TL-005, TL-008, TL-016 | Canonical JSON equivalence tests, Timeline and Projections service-backed runs, SL-02 handoff | All row-producing paths are equivalent and use one semantic service. |
| TL-010 | Invert projection ownership | WF-06 | DONE | TL-006, TL-009, TL-016 | Closed mutations and keyset enumeration; Projections query/write/rebuild ownership; Timeline and Projections service evidence; boundary and harness gates | Projection dependency direction is acyclic and tests pass. |
| TL-011 | Narrow cross-owner mutation effects | WF-06 | DONE | TL-008, TL-009, TL-010, TL-016 | Timeline-owned ports, peer-owner transaction methods, `internal/app/timelineassembly`, owner slices, and machine boundary evidence from SL-04 | Explicit ports preserve transactional side effects. |
| TL-012 | Refactor workbook paste/bulk boundary | WF-06 | DONE | TL-007, TL-016 | Workbook batch admission, validated Tabular Ingest V1 plans, exact Timeline batch commands, and SL-05 owner/service/browser evidence | Distinct commands use the approved ownership without wire changes. |
| TL-013 | Approve durable collaboration publication ownership | WF-05 | DONE | TL-004 | RB-005 closure decision and Section 3 event-intent record | Timeline owns intent; Collaboration owns post-commit sequence/replay/delivery. |
| TL-014 | Audit source-owner providers | WF-06 | DONE | TL-011 | Provider source audit, exact provider import/SQL boundaries, and Timeline/Revisions/Reporting/Incident Bundles owner evidence from SL-07 | Providers remain narrow and peer coordination remains outside Timeline. |
| TL-015 | Complete implementation verification and handoff | WF-08 | DONE | TL-010, TL-011, TL-012, TL-014, TL-017 | SL-08 handoff, final owner/browser evidence, `.cartulary/test-results/20260726T104424Z-p2987438` | Required commands and outcomes are recorded with no unresolved regression. |
| TL-016 | Promote closure decisions into owning authority | WF-05 | DONE | TL-006, TL-007, TL-008, TL-013 | SL-00A machine owner catalogs, provider descriptor, boundary/schema-owner manifests, and explanatory Core/Appendix text | Machine owners agree on interfaces, table ownership, transactions, and event intent; SL-00A validation passed. |
| TL-017 | Implement durable collaboration intent and route decoupling | WF-06 | DONE | TL-011, TL-013, TL-016 | Collaboration migration/store/dispatcher/runtime, source-owner intent ports, membership-filtered sockets, SL-06 owner/browser/full-service evidence | Routes do not publish and committed changes produce exactly one replayable authorized event. |

## 10. Session Handoff Log

Initial session time: `2026-07-25T22:39:55-04:00`; blocker-closure session time:
`2026-07-25T23:18:54-04:00`. The requested output did not previously exist, so
the initial session had no current-file history to preserve. The archived tracker
is cited as historical evidence only. Earlier handoff rows preserve the state
known at their timestamp and are superseded by later rows where closure posture
changed. For both planning sessions, the only file touched is this tracker.

### Remediation execution

| Time | Slice | Files changed and substantive decisions | Commands and result roots | Failures and migration/compatibility impact | Blockers | Exact next workstream |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-07-25T23:59:28-04:00 | SL-00A | Activated focused requirements for Timeline source/derivation/batch/peer-effect/intent behavior and peer-owner admission/plan/stream behavior; registered `module.projections` requirements, verifications, and test-family ownership; reserved Collaboration durable-stream schema ownership; published Timeline query/refresh provider capabilities with a closed upsert/delete source-owner mutation; refreshed generated topology outputs; updated non-executable Core/Appendix guidance. | `make generate` PASS (`.cartulary/test-results/20260726T035323Z-p593831`); `make json-shape-check` PASS (`.../20260726T035648Z-p600127`); `make generate-drift` PASS (`.../20260726T035648Z-p600191`); `make generated-artifact-policy-check` PASS (`.../20260726T035648Z-p600284`); `make backend-module-boundary-check` PASS (`.../20260726T035648Z-p600542`); `make service-backed-test-slice OWNER=module.projections` PASS (`.../20260726T035759Z-p608164`); `make harness-contract` PASS (`.../20260726T035928Z-p611494`). | Initial generation exposed the canonical extension-registry digest and was repaired; initial harness validation exposed closed catalog totals and duration coverage for the new owner and was repaired from the successful owner run. No public contract or data migration; generated outputs changed through `make generate`. | None; RB-003 evidence is the active task, not an owner contradiction. | SL-00 exact test disposition and frozen baseline. |
| 2026-07-26T00:22:19-04:00 | SL-00 | Closed exact-selector accounting for all Timeline and Projections tests; assigned the cross-incident paste, attached-evidence, rollback-provider, time-contract, and legacy keyset tests to their accountable owners; replaced five bespoke Timeline AST/source tests with machine rules; activated every Projections query/store/rebuild/provider/boundary/adapter/telemetry test; repaired dormant provider vocabulary, SQL scanner, Notes filter authority, ordered collection, surface-matrix, and rebuild-fixture defects; refreshed generated view/topology artifacts and measured duration baselines. | `make backend-module-boundary-check` PASS (`.cartulary/test-results/20260726T040531Z-p616023`); `make generate` PASS (`.../20260726T041758Z-p673160`); `make go-test-duration-baseline-coverage` PASS (`.../20260726T041806Z-p675399`); `make test-slice OWNER=module.timeline` PASS (`.../20260726T041206Z-p637501`); `make service-backed-test-slice OWNER=module.projections` PASS (`.../20260726T041133Z-p634017`); focused Evidence row PASS (`.../20260726T041619Z-p670874`); `make harness-contract` PASS (`.../20260726T041815Z-p675705`); `make service-backed-test-slice OWNER=module.timeline` PASS (`.../20260726T041843Z-p676796`). | Initial expanded Projections run failed on four previously unselected product/support defects and passed after repair. No public wire or source-data migration. Notes gained its already-implemented `artifact_type=note` machine filter; generated artifacts changed through `make generate`. | None. Desired-state equivalence/replay/rollback/durable-delivery tests remain mandatory entry work in their corresponding implementation slices, not unresolved baseline identities. | SL-01 cohesive Timeline application and contract surface. |
| 2026-07-26T00:29:51-04:00 | SL-01 | Retired the 1,373-line mixed `api.go`; partitioned request contracts, façade commands, results, decoding, normalized hashing, payloads, presentation, and API errors into cohesive files; moved result declarations out of storage/parser units; added a fixed canonical action-hash vector; retained the root Timeline façade and introduced no forwarding alias or compatibility shim. | `make format` PASS (`.cartulary/test-results/20260726T042918Z-p712532`); `make generate` PASS (`.../20260726T042921Z-p715198`); six focused Timeline decoder/hash/payload/error/projection rows PASS (`.../20260726T042929Z-p717437`); two frontend payload-contract rows PASS (`.../20260726T042951Z-p718801`); `make backend-module-boundary-check` PASS (`.../20260726T042934Z-p718380`). | No failures after partition. No public contract, hash, data, or migration impact; the fixed hash vector confirms byte stability and callers retained the same root-package symbols. | None. | SL-02 canonical source snapshot and row derivation. |
| 2026-07-26T00:47:44-04:00 | SL-02 | Added a Timeline source repository that reads only `timeline_events` and composes batch-capable Records envelopes; added Records envelope reads; made `workbookprojection.Derive` the single time-contract, projection-input, and presenter-record derivation; routed normal mutations, snapshots, mention effects, and projection rebuild enumeration through it; consolidated collection hydration in `rowsnapshot`; deleted duplicate source structs/loaders, projection DTO mapping, presenter mapping, and time derivation; added canonical JSON equivalence for nulls, UTC/local pairs, tags, mentions/provenance, evidence, replacement links, and group values. | `make format` PASS (`.cartulary/test-results/20260726T044329Z-p731135`); `make generate` PASS (`.../20260726T044333Z-p733808`); focused canonical projection row PASS (`.../20260726T044347Z-p736149`); `make service-backed-test-slice OWNER=module.timeline` PASS (`.../20260726T044359Z-p737069`); `make service-backed-test-slice OWNER=module.projections` PASS (`.../20260726T044711Z-p765362`); `make backend-module-boundary-check` PASS (`.../20260726T044733Z-p768002`). | One initial focused command used stale row IDs and failed as a usage error (`.cartulary/test-results/20260726T044135Z-p727612`); it made no repository change and the corrected owner rows passed. No public wire, projection-schema, source-data, or migration impact; old derivation paths were removed after parity. | None. | SL-03 Timeline/Projections ownership inversion. |
| 2026-07-26T01:21:58-04:00 | SL-03 | Added closed Timeline projection mutations and keyset-paged source enumeration; moved Timeline projection query SQL, storage writes, rebuild/restore, paging tests, and schema guards to Projections; injected Projections writer/query adapters through application composition; removed Timeline query-store and runtime-adapter ownership; updated provider façade and catalog ownership; repaired generic tag filtering, decoded JSONB collection presentation, and normalized matched-alias metadata parity. | Focused Timeline and Projections rows PASS (`.cartulary/test-results/20260726T050109Z-p789975`, `.../20260726T050624Z-p808064`, `.../20260726T051353Z-p842737`, `.../20260726T051539Z-p846680`); `make service-backed-test-slice OWNER=module.projections` PASS (`.../20260726T051743Z-p871158`); `make service-backed-test-slice OWNER=module.timeline` PASS (`.../20260726T051603Z-p848293`); `make backend-module-boundary-check` PASS (`.../20260726T051603Z-p848417`); `make generate` PASS (`.../20260726T052018Z-p884931`); `make json-shape-check`, `make generate-drift`, and `make generated-artifact-policy-check` PASS (`.../20260726T052033Z-p887346`, `.../20260726T052033Z-p887316`, `.../20260726T052033Z-p887363`); `make harness-contract` PASS (`.../20260726T052127Z-p893291`). | Initial compilation exposed and removed a Timeline/Entities/Projections import cycle; initial Timeline service integration exposed decoded `jsonb` collection and matched-alias parity defects, both fixed with regression evidence; initial harness runs exposed intentional catalog-count and duration-baseline drift, then passed after machine-input refresh. No public wire, projection-schema, source-data, or migration impact; existing projection rows remain valid and no dual query/writer path survives. | None. | SL-04 explicit transactional peer-owner ports. |
| 2026-07-26T03:02:36-04:00 | SL-04 | Declared narrow Timeline-owned transaction ports and typed peer results; added Records envelope operations, Revisions reads, Entities/Mentions write/effect contracts, Evidence validation, and injected Projections contracts; moved concrete composition and approved cross-owner read-shape hydration into `internal/app/timelineassembly`; injected Timeline effects into Entities mention/merge stores and a Timeline-capable projector into Evidence; deleted production Timeline hooks, test-only constructors, module override keys, duplicate `linkeffects`, and snapshot test APIs; moved harness routes/failures to application/test support; tightened projection importer and backend SQL ownership manifests. | `make test-slice OWNER=module.timeline` PASS (`.cartulary/test-results/20260726T061450Z-p1092932`); `make service-backed-test-slice OWNER=module.timeline` PASS (`.../20260726T061808Z-p1122255`); Entities PASS (`.../20260726T063429Z-p1244037`); Revisions PASS (`.../20260726T063807Z-p1272351`); Links PASS (`.../20260726T064112Z-p1316003`); Evidence PASS (`.../20260726T065558Z-p1414759`); Projections PASS (`.../20260726T070047Z-p1450613`); Cross-owner Transaction PASS (`.../20260726T070112Z-p1452508`); `make build-server` PASS (`.../20260726T070129Z-p1454210`); `make backend-module-boundary-check` PASS (`.../20260726T070142Z-p1463456`). | Initial owner runs exposed stale internal query/rebuild callers and unused imports, all migrated atomically. Evidence then exposed a source-less internal projector on attach/quarantine; replacing it with an injected projection port repaired backend rebuild/publication and the live browser lifecycle test. `OWNER=module.records` was rejected because no such owner catalog exists (`.../20260726T063736Z-p1271869`); Records invariants are covered by Timeline and Cross-owner Transaction evidence. No public wire, hash, projection schema, source-data, or migration change; internal constructors changed without compatibility shims. | None. | SL-05 Workbook admission, versioned tabular plans, and distinct bulk commands. |
| 2026-07-26T03:42:58-04:00 | SL-05 | Moved Timeline clipboard and bulk public decoding, target admission, exact-header handling, and request hashing to Workbook; replaced unversioned `BatchPlan` with validated `tabular_row_plan_v1` carrying normalized format, ordered columns/rows, field mappings/binding modes, unmapped raw values, closed warnings, and a deterministic mapping fingerprint; added Timeline owner-batch target/operation commands and whole-batch target preflight; implemented fill-down and tag assignment as direct owner commands; deleted `bulk_mutation.go` and all bulk-through-paste construction; retained exact public routes, envelopes, hashes, rows, conflicts, replay, raw capture, and exact-header padding. | `make format` PASS (`.cartulary/test-results/20260726T071954Z-p1471488`); `make build-server` PASS (`.../20260726T072002Z-p1474383`); Tabular Ingest owner PASS (`.../20260726T072028Z-p1483820`); Workbook owner PASS, including webserver-backed clipboard/bulk coverage (`.../20260726T072036Z-p1484222`); Timeline owner PASS (`.../20260726T072646Z-p1528680`); Timeline service-backed PASS (`.../20260726T072951Z-p1557467`); Workbook service-backed PASS (`.../20260726T073259Z-p1585386`); `make backend-module-boundary-check` PASS (`.../20260726T073909Z-p1626849`); Entities owner regression PASS (`.../20260726T073928Z-p1627270`); `make lint-markdown` PASS (`.../20260726T074521Z-p1656113`); `git diff --check` PASS. | No product-test failure occurred. The long Workbook owner runs include browser/service work and completed normally. No public HTTP/OpenAPI/view/hash/result/data/migration change; internal Timeline decoder, bulk request, paste forwarding, and unversioned-plan surfaces were removed after atomic caller migration. | None. | SL-06 durable shared collaboration stream and route decoupling. |
| 2026-07-26T06:06:24-04:00 | SL-06 | Added Collaboration-owned durable intent, incident cursor, replay event, and hashed resume-token tables; implemented leased `SKIP LOCKED` dispatch, unique intent sequencing, durable replay/retention, restart-safe resume, high-water live cutover, membership-filtered delivery, runtime lifecycle, and telemetry; migrated Timeline, Revisions, Entities/Mentions, Evidence, jobs, extension finalization, and Network Flow to transaction-coupled intents; removed route publishers and hub sequencing/replay; added sparse row-patch and projection invalidation/removal semantics; made restore/recovery composition Timeline-source aware; refreshed migration history, boundaries, duration baselines, and generated topology. | Durable stream and socket rows PASS (`.cartulary/test-results/20260726T085245Z-p1889243`, `.../20260726T085306Z-p1890652`); affected Timeline, Evidence, Entities, Revisions, Projections, Recovery, App Server, workbook-wire, and Network Flow rows PASS (`.../20260726T092815Z-p2171087`, `.../20260726T093123Z-p2180316`, `.../20260726T093444Z-p2186218`, `.../20260726T093633Z-p2207624`, `.../20260726T093757Z-p2210539`, `.../20260726T094148Z-p2266668`); restore browser PASS (`.../20260726T094838Z-p2382612`); `make test-service-backed` PASS (`.../20260726T095828Z-p2482910`); duration regeneration PASS (`.../20260726T100354Z-p2696331`); `make harness-contract` PASS (`.../20260726T100359Z-p2696613`); `make migration-drift` PASS (`.../20260726T100431Z-p2697736`); boundary, JSON-shape, generation, and generated-policy gates PASS (`.../20260726T100504Z-p2700639`, `.../20260726T100528Z-p2704020`, `.../20260726T100516Z-p2701465`, `.../20260726T100531Z-p2704525`, `.../20260726T100543Z-p2708332`). | Initial broad service run exposed the remaining restore-helper composition and passed after using `timelineassembly.NewRestoreRebuilder`. A later full run failed because the 16 GB `/tmp` filesystem was exhausted by a 14 GB reproducible Go build cache (`.../20260726T095407Z-p2411660`); removing only `/tmp/cartulary-go-build` restored capacity and the clean rerun passed. Migration is additive with no history backfill; rollout requires a drained single-process cutover, old tokens may reset, and rollback must leave the new tables intact. HTTP, OpenAPI, Timeline view/saved-view, and WebSocket v1 shapes remain unchanged. | None. | SL-07 narrow source-owner provider audit. |
| 2026-07-26T06:21:48-04:00 | SL-07 | Audited Timeline incident-bundle portability, reporting, delete/restore, and rollback providers and retained each only as source-owner adaptation: bundles contain Timeline source/profile files only; reporting emits raw canonical source facts and support references; delete/restore and rollback mutate/reconstruct Timeline source while Revisions retains generic transaction, envelope, history, projection, and event coordination. Added exact per-provider import allowlists, restricted rollback SQL to Timeline source tables, and registered reporting's explicit `records` plus `timeline_events` read shape; no duplicate orchestration or secondary application façade remains. | Boundary and JSON-shape gates PASS (`.cartulary/test-results/20260726T101016Z-p2711761`, `.../20260726T101019Z-p2712039`); focused Timeline, Revisions, Reporting, and Incident Bundles slices PASS (`.../20260726T101034Z-p2713410`, `.../20260726T101335Z-p2742513`, `.../20260726T101520Z-p2763723`, `.../20260726T101533Z-p2765565`); their service-backed slices PASS (`.../20260726T101621Z-p2769662`, `.../20260726T101921Z-p2797102`, `.../20260726T102058Z-p2818191`, `.../20260726T102109Z-p2819456`). | No product-test failure and no public wire, source-data, projection-schema, portability, reporting, history, rollback, or migration change. Provider behavior was retained because each adapter materially preserves source-owner semantics; the enforcement change prevents later provider growth into projection/replay/runtime ownership. | None. | SL-08 final validation and handoff completion. |
| 2026-07-26T10:44:24-04:00 | SL-08 | Re-ran discovery against the final tree, inventoried all 62 live Timeline files and retired paths, and removed the last unbounded hub-owned `recordChanges` observation seam. Test routes now observe durable Collaboration replay rows, Hub tests consume incident delivery, and Collaboration intent tests remain exactly cataloged. Closed every RB/TL item and confirmed no temporary shim, dual publisher, duplicate derivation, unauthorized SQL/import exception, or `docs/domain.md` change. | Cleanup format PASS (`.cartulary/test-results/20260726T102621Z-p2824870`); focused Platform WS, Collaboration, Workbook, Timeline, and Revisions rows PASS (`.../20260726T102813Z-p2828428`, `.../20260726T102816Z-p2828789`, `.../20260726T102822Z-p2830677`, `.../20260726T102833Z-p2833083`, `.../20260726T102853Z-p2835129`); harness, JSON, generation, generated-policy, migration, backend-boundary, and frontend-boundary gates PASS (`.../20260726T102934Z-p2837519`, `.../20260726T103002Z-p2838590`, `.../20260726T103005Z-p2839091`, `.../20260726T103017Z-p2842892`, `.../20260726T103019Z-p2843242`, `.../20260726T103050Z-p2846899`, `.../20260726T103053Z-p2847234`); webserver-backed, support, and stateful browser suites PASS (`.../20260726T103110Z-p2848094`, `.../20260726T103705Z-p2891116`, `.../20260726T103809Z-p2895396`); `make agent-finalize` PASS with retained-run maintenance skipped because `RESULTS_DIR` was unset (`.../20260726T104127Z-p2919088`); `make test-fast` PASS with 859 tests (`.../20260726T104153Z-p2925544`); `make check` PASS all 168 units and 715 tests (`.../20260726T104424Z-p2987438`); final tracker `make lint-markdown` and `git diff --check` PASS (`.../20260726T105520Z-p3087663`). | The first final boundary run identified the application-owned test route's legitimate durable replay read; an exact harness read-shape allowance repaired it and the rerun passed. No unresolved failure. Migration `00041` is additive with no backfill; deploy via drained single-process cutover, allow legacy resume tokens to return `reset_required`, and never drop the new tables during operational rollback. HTTP, OpenAPI, `cartulary.view.timeline.v2`, saved-view, and WebSocket v1 contracts are unchanged. Measurement and visual browser suites were skipped because their implementation and fixtures did not change. | None. | None; remediation and handoff complete. |

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
| RB-001 | Workbook owns command admission/public envelopes; Tabular Ingest owns normalized plans; Timeline owns Timeline application and committed effects. | Prevents duplicate parsers/domain interpreters and preserves exact bulk semantics. | Machine owner contracts plus SL-05 owner/service/browser evidence. | `CLOSED` by SL-00A and SL-05. |
| RB-002 | Timeline owns canonical source/descriptor semantics and `ProjectionInput`; Projections owns projection SQL/query/write/rebuild/storage; the initiating source owner coordinates the transaction. | Removes the current two-way runtime dependency while preserving source/projection atomicity. | Machine owner contracts plus SL-03 owner, boundary, rebuild, and query evidence. | `CLOSED` by SL-00A and SL-03. |
| RB-003 | Every executable test identity requires one exact verification-owner disposition; filename or package visibility has no accounting meaning. | Active requirements need exact accountable executable evidence, and retained tests may not be silently unselected or duplicated. | Exact owner family manifests, machine boundary rules, owner slices, and harness catalog validation. | `CLOSED` by SL-00. |
| RB-004 | SQL resides with the owner of the tables and invariants: Timeline-private source SQL, Projections projection SQL, peer-owner ports for peer state, domain-neutral platform mechanics. | Prevents platform policy leakage, cross-owner writes, and one all-purpose Timeline repository. | Machine ownership/boundary inputs plus SL-02 through SL-04 and SL-07 boundary evidence. | `CLOSED` by SL-00A, SL-02, SL-03, SL-04, and SL-07. |
| RB-005 | Timeline appends a durable typed event intent in the source transaction; Collaboration owns post-commit sequence, replay, retries, authorization-filtered delivery, and duplicate prevention; routes do not publish. | Closes the commit/publication gap and makes retry/restart semantics explicit without changing public WebSocket behavior. | Machine owner contracts, migration `00041`, and SL-06/SL-08 stream, fault, browser, and broad evidence. | `CLOSED` by SL-00A, SL-06, and SL-08. |

All RB items are closed by executable authority, implementation, and passing
evidence. No owner contradiction or unresolved remediation blocker remains.

### RB-003 required disposition artifact

The active test-family catalogs are the exhaustive executable
`timeline_test_disposition_v1`; the table below records its closed field
contract and the ten dispositions that changed in SL-00.

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

| Package or rule | Executable identity | Final classification and disposition |
| --- | --- | --- |
| `tools/backend_module_boundaries.json` | Five former `boundary_guard_test.go` identities | `duplicate_or_obsolete`; equivalent machine import/token rules pass, so the bespoke tests are deleted. |
| `./internal/modules/timeline` | `TestTimelinePageSQLIsKeysetBounded` | `direct_product_evidence`; Projections owner row `module.projections.query.timeline_legacy_keyset_e67384caff`, Timeline collaborator, pending code movement in SL-03. |
| `./internal/modules/timeline` | `TestClipboardPasteRejectsCrossIncidentRecordTarget` | `direct_product_evidence`; added to the existing Timeline clipboard support row. |
| `./internal/modules/timeline` | `TestAttachedEvidenceCreateAndPatch` | `cross_owner_product_evidence`; Evidence owner row `module.evidence.store.attached_evidence_create_and_patch_aa33ea0168`, Timeline collaborator. |
| `./internal/modules/timeline/rollbackprovider` | `TestTimelineProviderSourceForRollbackValueMapsTimelineCells` | `cross_owner_product_evidence`; Timeline owner row `module.timeline.support_unit.rollback_provider_mapping_dafe7d4b5c`, Revisions collaborator. |
| `./internal/modules/timeline/timecontract` | `TestTimelineTimeContractParsingAndFormatting` | `direct_product_evidence`; Timeline owner row `module.timeline.support_unit.time_contract_parser_98d70c6b65`. |
| `./internal/modules/projections` and `/adapters` | All 22 executable identities | Query/store/rebuild tests are product evidence; provider, manifest, source-ownership, adapter, import-boundary, and telemetry tests are architecture/support evidence. Every identity selects exactly once through `module.projections`. |

RB-003 is closed: retained identities select exactly once, deleted identities
have machine-policy parity, owner slices pass, and `make harness-contract`
validates catalog closure.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/timeline` is inventoried or explicitly out of scope. | PASS | Section 2 has rows for all 62 final files and nine retired baseline paths; `.gitkeep` is explicitly non-runtime. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps routes, view/schema, WS, revisions, projections, auth, providers, frontend, and harness accounting. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 supplies predecessor/successor relationships and handoff checkpoints. |
| Every proposed slice is behavior-preserving unless explicitly marked otherwise. | PASS | Section 7 declares behavior preservation and later authorization for any behavior change. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 contains Make-owned commands and conditional scopes; no command is invented. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No contradiction was found; Section 1 requires the exact marker if one appears later. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Section 1 records the mixed live package and stale archived completion state. |
| Handoff sections are current enough to continue without rediscovery. | PASS | Section 10 records all required workstream handoffs, inspections, results, blockers, and next actions. |
| SL-00A executable authority and owner registration are complete. | PASS | Focused owner requirements, `module.projections`, provider capabilities, schema/boundary inputs, generated topology, and required validation are recorded in Section 10. |
| SL-00 exact-selector disposition and frozen baseline are complete. | PASS | All retained Timeline/Projections identities select exactly once; redundant architecture tests are replaced by machine rules; owner slices and harness validation pass. |
| SL-01 cohesive application and contract partition is complete. | PASS | Mixed `api.go` is retired; cohesive declarations/decoder/hash/payload/error/presentation files pass focused backend and frontend contract rows with a fixed hash vector. |
| SL-02 canonical source snapshot and row derivation are complete. | PASS | One source repository plus one derivation service serve mutation, snapshot, effect, and rebuild paths; canonical JSON parity and both Timeline and Projections service-backed slices pass. |
| SL-03 Timeline/Projections ownership inversion is complete. | PASS | Timeline exposes closed mutations and bounded source enumeration; Projections owns physical query/write/storage/rebuild/restore; application assembly injects ports; owner, boundary, generated-drift, and harness gates pass. |
| SL-04 transactional peer-owner ports are complete. | PASS | Timeline production code constructs no peer runtime adapter, writes only Timeline source state, and exposes no production test hook; Timeline, Entities, Revisions, Links, Evidence, Projections, and Cross-owner Transaction slices plus server build and machine boundaries pass. |
| SL-05 Workbook admission and exact owner-batch application are complete. | PASS | Workbook owns public admission and hashes; validated `tabular_row_plan_v1` is the sole clipboard plan; fill-down and tag assignment have distinct Timeline handlers; owner, service-backed, browser, entity-regression, build, and boundary evidence pass. |
| SL-06 durable shared collaboration stream and route decoupling are complete. | PASS | Source owners append transaction-coupled intents; Collaboration owns durable sequencing, replay, retention, opaque resume state, retry, and current-membership delivery; routes and the hub no longer publish or sequence; migration, focused owner/browser, full service-backed, harness, generated, and boundary evidence pass. |
| SL-07 narrow source-owner provider audit is complete. | PASS | Portability, reporting, delete/restore, and rollback retain only Timeline source adaptation; exact import and SQL/read-shape rules prevent secondary application, projection, replay, or runtime ownership; all four accountable owner and service-backed slices pass. |
| SL-08 validation and handoff completion is complete. | PASS | Final focused, contract, drift, migration, boundary, browser, fast, and 168-unit broad checks pass; cleanup, migration, compatibility, failure, and skipped-suite posture are recorded in Section 10. |
| Former design blockers have one consistent closure posture. | PASS | RB-001 through RB-005 are closed by machine authority, implementation, and passing evidence. |
| Verification uncertainty is neither guessed nor hidden. | PASS | RB-003 is closed by exact catalog disposition, machine boundary parity, and passing owner/harness evidence. |
| Tracker planning evidence is not treated as owner authority. | PASS | Executable authority resides in the machine catalogs; Core, Appendix, and tracker text are human guidance and evidence. |

This tracker records the completed authorized remediation. SL-00A through SL-08,
every TL item, every RB item, and every binary exit criterion are complete. No
unresolved blocker or regression remains.
