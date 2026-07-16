# networkflow Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Item | Current value |
| --- | --- |
| Target path | `internal/modules/networkflow` |
| Target label | `networkflow` |
| Output path | `docs/handoffs/networkflow-module-refactor-tracker.md` |
| Repository snapshot | Clean `main` at `63a19bf6e0c3f17a9b53afb58afa8bd04fb4c550` when planning began |
| Status | Historical module remediation complete; Network Analysis grid-adapter adoption is executing under Section 14, with `NF-GA-00` through `NF-GA-08` `COMPLETE` and `NF-GA-09` awaiting activation |
| Authorized change | Specification, implementation, tests, contracts, generated outputs, configuration, migrations, harness inputs, documentation, and this tracker |
| Current execution authorization | Specification, contract, generated, backend, frontend, test, harness-input, visual, and tracker changes required by Section 14; no dependency, lockfile, database, saved-view, or React Data Grid version migration |
| Preserved non-goals | No richer graph canvas, observation creation, restore/purge behavior, existing-route response change, or HTTP schema-major expansion; one additive Core import mapping-preview route is explicitly authorized |
| Implementation authority | The 2026-07-13 remediation task explicitly authorized the behavior corrections and refactor described by the adopted plan |

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs for their named subsystems.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for later claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md`, `docs/design.md`, implementation guides, and the Testing Harness NLSpec for terminology, design direction, package boundaries, and execution mechanics.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and the modular-refactor framework as evidence and planning doctrine only.

Owner and support documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was read first and used as planning doctrine, not repository-state evidence.
- `docs/network-flow-activity-nlspec.md`.
- `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/04_security_deployment_and_conformance.md`.
- `docs/graph_projection_nlspec.md`.
- `docs/testing-harness-nlspec.md`.
- `docs/domain.md` and `docs/design.md`.

Repository surfaces inspected:

- every one of the 28 files under `internal/modules/networkflow`, inventoried in Section 2;
- `internal/app/runtime.go`, `internal/app/server_profile_harness.go`, and the Network Flow process-test callers;
- the Network Analysis frontend under `apps/web/src/networkFlow`, its Phase 12 browser test, `WorkbookShell.tsx`, and workbook startup/collaboration consumers;
- `contracts/network-flow/*`, `contracts/index.json`, `contracts/extensions/index.json`, and the derived OpenAPI, Go, and TypeScript contract consumers;
- Network Flow migrations `00028` through `00030`, the schema-object ownership manifest, backend boundary policy, Graph Projection boundary guard, Network Flow accounting manifest, and Phase 12 test map;
- the live Make task surface through `make help`, `make help-all`, `make task-guide`, `make explain-phase`, and `make explain-target`.

No normative owner contradiction was found. At the planning snapshot, `docs/domain.md` described Network Flow as draft while Core 00 and `docs/network-flow-activity-nlspec.md` marked it adopted/current. That subordinate vocabulary drift and the implementation/owner mismatches recorded below have now been remediated. The inventories and findings in Sections 2 through 8 are retained as the historical baseline that motivated the change; Sections 9 through 13 own current completion status.

The framework expects a narrow module facade and private implementation seams. The planning snapshot instead exposed a broad store and import facade and combined transport, persistence, transaction, projection-adapter, and mutation coordination. The implemented root `Module`, owner ports, transaction runner, Graph Projection adapter, and frontend seams resolve that baseline mismatch; some repository-private storage types remain exported only where external-package store tests require them.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/networkflow/api.go` | Profile identity, lifecycle vocabulary, limits, sentinel errors, and typed validation errors | `ProfileID`, `WorkspaceKey`, schema/profile/status constants, error values, validation error types, `Limits`, `DefaultLimits`, `LifecycleStates` | Other Network Flow files, application assembly through package constructors, package and external-package tests | Standard library only | Phase 12 unit/contract tests, store and route tests | `contracts/network-flow/index.json`, `schemas.v1.json`, `errors.v1.json`, derived Go/TS/OpenAPI surfaces | `networkflow` contract/application core | medium | Legitimate extension-owned vocabulary; changing values is observable behavior. |
| `internal/modules/networkflow/binding_store.go` | Indicator-binding persistence, binding dedupe, and target-indicator lookup | `NetworkFlowRowRef`, `IndicatorBindingRecord`, `CreateIndicatorBindingParams`, `Store.CreateOrReuseIndicatorBindingTx`, `Store.GetActiveIndicator` | `indicator_link.go`, route integration tests | `pgx`, `internal/modules/indicators`, shared `Store` database handle | Route integration and Phase 12 indicator-link selectors | Indicator-binding schemas/routes; migration `00030_network_flow_indicator_bindings.sql` | Split: Network Flow binding repository plus indicators-owned read participant | high | Reads the `indicators` table and reconstructs `indicators.IndicatorRecord` directly. |
| `internal/modules/networkflow/csv_parser.go` | RFC 4180 preview/apply parsing, row validation, normalization, diagnostics, and unmapped-raw handling | `ParsedCSV`, `CSVRecord`, `ParseCSVPreview`, `ParseCSVApply`, `ValidateRows` | `import_facade.go`, Phase 12 tests | Standard CSV/time/network helpers plus Network Flow mapping/digest logic | Phase 12 unit/contract tests | Import preview/apply and row/diagnostic schemas | `networkflow` source-profile and normalization core | high | Legitimate owner logic; timestamp behavior has an owner/implementation gap recorded in RB-003. |
| `internal/modules/networkflow/digest.go` | Canonical digests and IDs for rows, endpoints, edges, diagnostics, mappings, and safe values | `SourceRowDigest`, `NormalizedRowDigest`, `RowID`, `EndpointID`, `FlowEdgeID`, `DiagnosticID`, `MappingFingerprint`, `SafeDigest` | Parser, mapping, graph, import, and tests | Standard crypto/JSON plus UUID | Phase 12 digest, graph, safe-digest, and fixture tests | Network Flow identity and schema algorithms | `networkflow` identity core; key lifecycle remains security-owned | high | Identity changes would invalidate persisted and public references. |
| `internal/modules/networkflow/graph.go` | Flow-specific table-scope resolution, graph composition, ephemeral Graph Projection input, response annotations, contributor recomputation, and graph audit | Unexported request/composition helpers and `Service` methods | `routes.go`, route integration and Phase 12 tests | `internal/modules/graphprojection`, `platform/httpapi`, Network Flow store/query | Route integration, Phase 12 graph selectors, Graph Projection boundary guard | Graph request/result/contributor schemas and routes | Keep composition in `networkflow`; split an approved `graphprojection` adapter seam | high | The direct Graph Projection dependency is explicitly allowed only in `graph.go` and `routes.go`. |
| `internal/modules/networkflow/import_facade.go` | Extension import-owner preview, mapping materialization, source verification, validation, and atomic apply | `ImportFacade`, options, `NewImportFacade`, `PrepareImportUnitMapping`, `ApplyImportUnitTx` | `internal/app/runtime.go`, generic import runtime through module override, tests | Concrete `imports.Store`, `platform/jobs`, `pgx`, Network Flow store/parser | Phase 12 import selectors and store/integration coverage | Import integration block in `routes.v1.json`; mapping/import schemas | Keep Network Flow owner facade; depend on imports-owned source/session ports | high | Legitimate owner facade, but it exposes concrete storage and job dependencies. |
| `internal/modules/networkflow/indicator_link.go` | Indicator-link request admission, row/graph selector resolution, create-or-find orchestration, binding commit, replay, and response mapping | Unexported request types and `Service` handler methods | `routes.go` | `indicators`, `authn`, `httpapi`, `pgx`, direct shared store pool | Route integration and Phase 12 indicator selectors | Indicator link/binding request and result schemas/routes | Network Flow application service using indicators-owned transaction participant | high | Constructs an indicators store from `s.store.pool` inside route orchestration. |
| `internal/modules/networkflow/mapping.go` | Public mapping DTOs, closed mapping variants, validation/materialization, Cisco SNA aliases, and source suggestions | `TimestampProfile`, mapping/source DTOs, `MaterializeApprovedMapping`, `DecodeApprovedMapping`, `MarshalApprovedMapping`, `SuggestCiscoSNAMapping`, `SourceAliasMatchKey` | Import facade, parser, digests, tests | Standard JSON plus Unicode normalization | Phase 12 mapping, timestamp, and alias tests | `schemas.v1.json`, generated protocol contracts | `networkflow` mapping core | high | Public mapping wire shape; Unicode behavior participates in RB-003. |
| `internal/modules/networkflow/names.go` | Source filename sanitization, table-name normalization, deterministic collision suffixes | `SanitizeSourceFilenameDisplay`, `NormalizeTableDisplayNameInput`, `DeriveTableDisplayName` | Store/import code and tests | Unicode normalization and runtime Unicode whitespace classification | Store tests and Phase 12 filename/name selectors | Table and import result schemas | `networkflow` table lifecycle core | medium | Uses `unicode.IsSpace` rather than the NLSpec's closed Unicode 17.0 set. |
| `internal/modules/networkflow/query.go` | Strict query DTO admission, table scopes, filters, sorts, comparators, filtering, and in-memory ordering | `Filter`, `SortSpec`, `TableScope`, `RowQueryRequest`, `RejectedRowsQueryRequest`, `FieldEndpointIP` | `routes.go`, `graph.go`, tests | `platform/httpapi` and standard JSON/network/time helpers | Route integration and Phase 12 query selectors | Query/filter/sort/table-scope schemas and route contracts | `networkflow` query core, separated from transport | critical | Effective sort and sortable ID behavior differ from the adopted owner; see RB-001. |
| `internal/modules/networkflow/resources.go` | Map-shaped HTTP serialization for tables, rows, diagnostics, limits, profiles, and bindings | Unexported resource assemblers | Route, graph, indicator, and import handlers | Network Flow records and standard library | Route integration and Phase 12 resource-shape tests | Most Network Flow success schemas and generated outputs | Network Flow HTTP adapter | high | Transport serialization is mixed into the same package as domain/application logic. |
| `internal/modules/networkflow/routes.go` | Route registration, service assembly, extension claim checks, auth/roles, idempotency, table/query handlers, cursors, audit, and WebSocket invalidation | `Service`, `RegisterRoutes` | `internal/app/runtime.go`; route tests | Incidents, Graph Projection, auth/session/http helpers, Postgres, WebSocket platform, concrete store | Route integration, Phase 12 contract tests, process tests | All eleven route contracts, errors, envelopes, OpenAPI, WS invalidation semantics | Thin Network Flow HTTP adapter plus application facade | critical | Central mixed-responsibility file; owns page-offset continuation and master-derived keys today. |
| `internal/modules/networkflow/security.go` | AES-GCM cursor encoding, TTL, payload binding, validation, and error collapse | `CursorCodec`, `CursorBinding`, `CursorPayload`, `NewCursorCodec`, `Encode`, `Decode`, `Validate` | `routes.go` and route/Phase 12 tests | Standard cryptography and JSON | Route pagination/invalidation tests and Phase 12 cursor selectors | Cursor schemas/errors and Core 04 key lifecycle contract | Split: Network Flow cursor semantics plus security-owned key-ring adapter | critical | Payload stores an offset and codec supports one key, conflicting with keyset and rotation requirements. |
| `internal/modules/networkflow/store.go` | Concrete PostgreSQL persistence for tables, immutable rows, diagnostics, lifecycle, limits, and audit events | `Store`, store options, table/row/diagnostic DTOs, create/rename/delete params, `NewStore`, lifecycle/read methods | Application runtime, routes, import facade, indicator linker, tests | `platform/postgres`, `pgx`; direct SQL to Network Flow, `incidents`, and administrative-audit tables | Store, route integration, and Phase 12 tests | Migrations `00029`/`00030`, table/row/diagnostic schemas | Network Flow repositories plus incident/audit owner ports | critical | Module owns `network_flow_*` storage, but cross-owner locks and audit writes must be isolated. |
| `internal/modules/networkflow/phase12_network_flow_contract_test.go` | Phase 12 acceptance-selector wrappers, focused source/contract assertions, fixture routing, and egress scans | Named `TestPhase12NetworkFlow_*` selectors | `tools/phase12_test_map.json`, backend-unit target | Target package internals, owner documents, fixture manifests, selected platform source | This file is test evidence | Phase 12 map, ledger/schedule/accounting outputs | `networkflow` characterization/evidence tests | high | Many selectors share broad helpers or fall through to fixture evidence; accounting is not runtime characterization. |
| `internal/modules/networkflow/phase12_network_flow_unit_test.go` | Unit acceptance selectors, frozen-fixture verification, algorithm vectors, and owner-document checks | Named `TestPhase12NetworkFlow_*` selectors | Phase 12 map and backend-unit target | Target internals, Graph Projection, repository files/fixtures | This file is test evidence | Phase 12 map and Network Flow fixture manifests | `networkflow` characterization/evidence tests | high | Some acceptance labels rely primarily on fixture/structural evidence; audit before refactor. |
| `internal/modules/networkflow/routes_integration_test.go` | Claimed/unclaimed routing, table query/pagination, lifecycle, idempotency, audit, WS invalidation, graph, contributors, and indicator links | Three external-package integration tests plus local helpers | Backend-store/integration support and Phase 12 service slice | Shared auth/incident test support, Postgres, object store, HTTP and WS harnesses | This file is integration evidence | Route, envelope, audit, and WS contracts | `networkflow` integration tests | high | Strong broad-path coverage but currently characterizes offset continuation as successful. |
| `internal/modules/networkflow/store_test.go` | Store creation, persistence, lifecycle, limits, name rules, immutability, and Phase 12 store selectors | Store and Phase 12 tests | Backend-store target and Phase 12 map | Postgres test harness, target store API | This file is store evidence | Storage migrations and lifecycle contracts | `networkflow` store tests | medium | Preserve as repository characterization while splitting persistence. |
| `internal/modules/networkflow/harnesscontrol/controls.go` | Groups Network Flow fault, randomness, auth-transition, and audit-assertion controls into one harness contribution | `Controls`, `NewControls`, `Clear`, `Contribution` | `internal/app/server_profile_harness.go`, process tests | `platform/harnessruntime` and owner-local registries | Harness-control integration tests and process tests | Test-support inventory and harness schemas | `networkflow/harnesscontrol` | medium | Intentional owner-local runtime-scanned test support; never part of ordinary production assembly. |
| `internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion.go` | Guarded audit assertion route, pending assertion registry, exact consume, and reset | Registry/assertion types and constructors/route/consume/reset methods | `controls.go`, future product dependency adapters, tests | `platform/httpapi` and harness runtime guard | Matching integration test | Audit-assertion control schema and Phase 12 harness rows | `networkflow/harnesscontrol` | medium | Harness mechanics only; cannot define product audit semantics. |
| `internal/modules/networkflow/harnesscontrol/network_flow_auth_transition.go` | Guarded auth-transition route, tuple-scoped pending state, consume, and reset | Registry/transition types and constructors/route/consume/reset methods | `controls.go`, future product dependency adapters, tests | `platform/httpapi` and harness runtime guard | Matching integration test | Auth-transition control schema and Phase 12 harness rows | `networkflow/harnesscontrol` | medium | Harness mechanics only; preserves hidden-resource disclosure rules. |
| `internal/modules/networkflow/harnesscontrol/network_flow_fault.go` | Guarded one-shot fault registry and control route | Registry/fault types and constructors/route/consume/reset methods | `controls.go`, future product dependency adapters, tests | `platform/httpapi` and harness runtime guard | Matching integration test | Fault-control schema and Phase 12 harness rows | `networkflow/harnesscontrol` | medium | Harness mechanics only; no product fault API. |
| `internal/modules/networkflow/harnesscontrol/network_flow_randomness.go` | Guarded deterministic string/UUID/hex randomness streams with fail-closed exhaustion and reset | Registry/state types and constructors/route/consume/state/reset methods | `controls.go`, future product dependency adapters, tests | UUID, `platform/httpapi`, harness runtime guard | Matching integration test | Randomness-control schema and Phase 12 harness rows | `networkflow/harnesscontrol` | medium | Harness values must remain undisclosed and unavailable outside harness assembly. |
| `internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion_integration_test.go` | Audit-control availability, guard, validation, exact/no-audit assertions, duplicate protection, and reset coverage | Six integration tests | Harness integration-support target | Owner-local helpers and guarded test server | This file is harness evidence | Audit-control harness accounting | `networkflow/harnesscontrol` tests | low | Preserve route/schema/guard behavior if paths move. |
| `internal/modules/networkflow/harnesscontrol/network_flow_auth_transition_integration_test.go` | Auth-transition availability, guard, tuple matching, duplicates, validation, and reset coverage | Six integration tests | Harness integration-support target | Owner-local helpers and guarded test server | This file is harness evidence | Auth-transition harness accounting | `networkflow/harnesscontrol` tests | low | Preserve hidden-resource and independent-key behavior. |
| `internal/modules/networkflow/harnesscontrol/network_flow_fault_integration_test.go` | Fault-control availability, guard, arming, pending conflict, validation, and reset coverage | Six integration tests | Harness integration-support target | Owner-local helpers and guarded test server | This file is harness evidence | Fault-control harness accounting | `networkflow/harnesscontrol` tests | low | Preserve exact boundary/correlation consumption. |
| `internal/modules/networkflow/harnesscontrol/network_flow_randomness_integration_test.go` | Randomness availability, guard, stream behavior, duplicates, validation, binary-body rejection, and reset coverage | Eight integration tests | Harness integration-support target | Owner-local helpers and guarded test server | This file is harness evidence | Randomness-control harness accounting | `networkflow/harnesscontrol` tests | low | Preserve ordered duplicate values and fail-closed exhaustion. |
| `internal/modules/networkflow/harnesscontrol/test_helpers_test.go` | Shared harness-control server, auth, envelope, and object-store test helpers | Test-only helper functions | Four harness-control integration-test files | Generic harness/runtime and test dependencies | All harness-control integration tests | None directly | `networkflow/harnesscontrol` tests | low | Test-only owner-local support; not a production helper surface. |

No file under the target is out of scope. The schema ownership manifest assigns `network_flow_*` objects to `networkflow`; there are no authored `db/queries` inputs for this module, and current SQL is embedded in the store. The frontend does not import a grid vendor or `@cartulary/grid-adapter`; it renders HTML tables and dense graph/contributor regions directly.

## 3. Module Boundary Diagnosis

The target is a legitimate Network Flow extension boundary, but the current package is a mixed-responsibility package. It acts as a view/projection orchestration layer, transport-adjacent adapter, persistence-adjacent adapter, mutation coordinator, and broad application service. It is not an accidental home for unrelated Base Profile behavior, and it should not be dissolved merely because it is large.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Profile identity, limits, lifecycle vocabulary, errors | `api.go` | `networkflow` | keep | Adopted Network Flow NLSpec and generated contract family | Keep stable behind a narrower public facade. |
| CSV source profile, mapping, normalization, row validation, diagnostics | `csv_parser.go`, `mapping.go`, `names.go`, `digest.go` | `networkflow` | keep | NLSpec Sections 6, 8 through 12; import owner facade | These are extension-owned translation and identity rules, not generic imports. |
| Import session/source lookup and job mechanics | `import_facade.go` concrete dependencies | `imports` and platform jobs behind owner ports | split | Core import owner boundary; concrete `imports.Store`/jobs imports | Network Flow retains preview/apply owner semantics and mapping payloads. |
| Table, immutable-row, diagnostic, and binding persistence | `store.go`, `binding_store.go` | `networkflow` repositories | split | Schema ownership manifest and migrations `00029`/`00030` | Keep Network Flow SQL private; remove cross-owner SQL and broad store exposure. |
| Incident serialization lock | `store.go` direct `incidents` table lock | `incidents` transaction participant | move | Direct SQL to sibling-owned table | Preserve atomicity through an owner-provided lock/transaction capability. |
| Indicator lookup/create participation | `binding_store.go`, `indicator_link.go` | `indicators` participant plus Network Flow binding service | split | Direct indicator-table read and `indicators.NewStore(s.store.pool)` | Network Flow owns candidate resolution and binding provenance, not indicator persistence internals. |
| Audit event meaning and delivery | `store.go`, `graph.go`, routes | Network Flow event semantics plus Core-owned transactional audit port | split | Direct write to `deployment_admin_audit_events`; owner audit requirements | Preserve exact occurrence, no-audit replay, redaction, and transaction ordering. |
| HTTP request admission, envelopes, resources, route authorization | `routes.go`, `query.go`, `resources.go`, `indicator_link.go` | Thin Network Flow transport adapter plus application service | split | Eleven registered routes and broad `Service` dependency set | Route-time policy remains Network Flow/Core-owned; generic transport/session mechanics remain platform-owned. |
| Cursor semantic binding and cryptography | `security.go`, `routes.go` | Network Flow query semantics plus Core 04 key-ring provider | split | NLSpec Section 13.5 and Core 04 Network Flow key-ring requirements | Keyset state and key rotation require separately authorized behavior remediation. |
| Flow-specific graph composition | `graph.go` | `networkflow` | keep | Network Flow NLSpec Section 14 | It owns source mapping, aggregation, annotations, and contributor semantics. |
| Generic graph derivation | `graph.go` calls `graphprojection.Service.ProjectEphemeral` | `graphprojection` | keep boundary / split adapter | Adopted Graph Projection NLSpec and boundary guard allowlist | Do not move flow semantics into Graph Projection or expose Graph Projection publicly. |
| Harness fault/randomness/auth/audit controls | `harnesscontrol/*` | `networkflow/harnesscontrol` | keep | Testing Harness NLSpec TH-HARNESS-REQ-657 through 664 | Owner-local runtime-scanned mechanics are intentional and harness-only. |
| Network Analysis controller and rendering | `apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx` | Network Flow frontend feature | split | One component owns loading, tables, queries, graph, contributors, import, link, and rendering | Split internally without changing UI, selectors, or workspace identity. |
| Public browser wire types and routes | `networkFlowClient.ts` | Generated protocol/view-contract adapter plus feature client | split | Manual duplicates of Network Flow schema IDs and response shapes | Generated files remain downstream and must never be hand-edited. |
| Generic WebSocket lifecycle | `useNetworkFlowExtensionEvents.ts` | Workbook collaboration client with Network Flow event interpreter | split | Extension hook owns hello/resume/ping/reconnect/dedupe | Preserve `/ws/v1/incidents/{incident_id}` and `extension_resource_changed`. |
| Saved views, view schemas, Timeline, Evidence, Links, Core revisions | Not found in the target | Existing owner modules | defer / no move | No imports or runtime paths found; NLSpec explicitly excludes these areas | Flow rows are not Core records; graphs are ephemeral; no saved graph views or projection refresh exist. |
| Grid-vendor integration | Not found | Existing frontend grid adapter only if later required | defer | No vendor/grid-adapter import in Network Flow frontend | Introducing a grid adapter or changing presentation is outside this refactor plan. |

## 4. Public Contract and Behavior Freeze Map

Observable behavior is frozen unless a later task explicitly authorizes a conformance correction. A mismatch below is not silently promoted to desired behavior: characterize current behavior, obtain authority, then correct it in a separate atomic slice.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Claimed/unclaimed `network_flow_activity` profile and reserved route root | Core 00/Core 01 plus Network Flow NLSpec | Core 00 REQ-00-064, extension discovery contract, app config/assembly | Unclaimed route integration test, config and process tests | Claimed startup and unclaimed dispatch remain byte/status stable | high | No workspace or routes when unclaimed. |
| `GET .../source-profiles` | Network Flow NLSpec/contract | `routes.v1.json`, `routes.go`, source-profile resource | Route integration and Phase 12 resource/limit selectors | Exact envelope, limits, ordering, role, and strict query admission | medium | Viewer route. |
| `GET .../tables` | Network Flow NLSpec/contract | Route contract, store list, table resource | Route integration and store tests | Active-only order, hidden/soft-deleted behavior, exact envelope | high | Soft-deleted tables stay retained but absent. |
| `GET .../tables/{table_id}` | Network Flow NLSpec/contract | Route contract and active-table lookup | Phase 12 route status evidence; partial integration coverage | Active, hidden, missing, and disclosed soft-deleted outcomes | high | No deleted-table inspection route. |
| `PATCH .../tables/{table_id}` | Network Flow plus Core idempotency/audit | Route contract, rename handler/store/audit/WS | Route integration covers changed, no-op, replay, divergent replay | Role loss, stale version, hidden resource, exact audit/WS occurrence | critical | Preserve table ID, rows, mapping, and replay semantics. |
| `DELETE .../tables/{table_id}` | Network Flow plus Core idempotency/audit | Route contract, soft-delete handler/store/WS | Route integration covers delete, replay, cursor invalidation | Reviewer role, stale/hidden cases, retained-count effects, all scoped invalidations | critical | Terminal soft delete; no restore or purge. |
| `POST .../tables/{table_id}/query` | Network Flow query/cursor owner | Query schemas, `query.go`, `routes.go`, cursor codec | Route integration and Phase 12 selectors | Full effective sort, sortable ID fields, null ordering, keyset continuation, auth and delete recheck | critical | Current offset continuation and abbreviated effective sort differ from owner; RB-001. |
| `POST .../rows/query` | Network Flow query/cursor owner | Cross-table scope/filter/sort contracts | Phase 12 scope/filter/graph helpers | All scope variants, order, hidden IDs, empty scope, duplicate IDs, keyset continuation | critical | Existing helpers do not close every branch. |
| `POST .../tables/{table_id}/rejected-rows/query` | Network Flow diagnostic/cursor owner | Diagnostic schema and route contract | Route integration and fixture evidence | Exact diagnostic comparator tuple, filters, cursor continuation, redaction | critical | Current continuation uses an offset. |
| `POST .../graphs/query` | Network Flow graph owner plus Graph Projection | Graph contract, `graph.go`, approved facade | Route integration, Phase 12 graph tests, Graph Projection tests | Exact semantic digest, limits, overlap, aggregation, adapter failure classes, no retained state | critical | Graph Projection remains internal and ephemeral. |
| `POST .../graphs/contributors/query` | Network Flow graph/cursor owner | Contributor schema and recomputation code | Route integration and Phase 12 selectors | Exact contributor comparator/keyset cursor, stale digest, authorization recheck | critical | Must recompute from authoritative immutable rows. |
| `POST .../indicator-links` | Network Flow binding plus Core indicator owner | Link/binding schemas, indicator participant, binding store | Route integration and Phase 12 indicator selectors | Existing/create target variants, roles, exact replay visibility, dedupe, atomic audit/binding | critical | No automatic observation creation in v1. |
| Import-owner preview and apply | Network Flow NLSpec plus generic imports owner | `routes.v1.json` import integration, import facade, schemas | Phase 12 unit/contract and store coverage | Source change, preview/apply boundary, all-rejected, atomic publish, cancellation/retry | critical | Generic upload framing remains imports-owned. |
| Table/row/diagnostic storage | Network Flow NLSpec | Migrations `00029`/`00030`, store and schema manifest | Store tests and route integration | Constraint/immutability, rollback, retained counts, diagnostic ordering | critical | No migration is proposed by the refactor. |
| Row/query mutation and revisions | Network Flow NLSpec explicitly excludes row editing/Core revisions | Owner non-goals and immutable-row trigger/tests | Immutable-row store test | Preserve rejection/absence of row mutation and revision routes | high | Not Core entity row/query/mutation behavior. |
| Current authorization and no deployment-admin bypass | Core 04 plus Network Flow matrix | Route role helpers and owner requirements | Source scan selector and route tests | Viewer/editor/reviewer/admin matrix, membership loss, hidden resources, replay recheck | critical | Authorization must be rederived at route time. |
| Route idempotency and transaction ordering | Core idempotency owner plus Network Flow | Route/store/indicator code and route integration | Rename/delete/link replay coverage | Failure non-persistence, target visibility replay, crash/rollback boundaries | critical | Hidden transaction coordination must become explicit without reordering effects. |
| Audit occurrence and redaction | Network Flow event definitions plus Core audit delivery | Direct audit SQL, Phase 12 audit controls | Route integration and structural/fixture selectors | Exact changed/no-op/replay/failure counts and safe payloads | critical | Storage realization may move; event meaning cannot. |
| WebSocket path, handshake, and `extension_resource_changed` | Core 03 collaboration plus Network Flow consequences | `platform/ws`, route publications, frontend hook | Route integration and Phase 12 browser test | Rename/delete/auth loss, reconnect/resume/dedupe, stale state, no hidden disclosure | high | Preserve generic `/ws/v1/incidents/{incident_id}`. |
| Network Analysis workspace and UI selectors | Network Flow NLSpec plus `design.md` direction | Workspace component, shell registry, UI-contract selector builders | Frontend unit tests and Phase 12 browser test | Workspace identity, roles, stable selectors, superseded responses, graph stale/link intervals | high | UI presentation changes are out of scope. |
| Generated Network Flow contracts | Network Flow owner inputs and contract generator | `contracts/network-flow/*`, contracts index, OpenAPI, generated Go/TS | JSON-shape, generation-drift, contract tests | Drift remains zero after any owner-input change | critical | Never hand-edit generated roots. |
| Saved views and view-schema behavior | Core 01/Core 03 owners; not Network Flow v1 | NLSpec exclusions and no target imports | Workbook tests outside target | Preserve absence of saved-view/view-schema promotion | medium | Not applicable to this refactor. |
| Projection refresh behavior | Graph Projection retained lifecycle owner; not used here | Ephemeral-only Network Flow adapter | Graph Projection and Network Flow graph tests | Preserve absence of retained refresh/state | medium | Not applicable; graph responses are ephemeral. |
| Grid-adapter/vendor contract | No current Network Flow contract | No grid-adapter/vendor imports | Current frontend/browser tests | Preserve current selectors and visible behavior if internals split | low | Do not introduce a vendor dependency in this refactor. |
| Phase 12 harness and evidence accounting | Testing Harness NLSpec | `tools/phase12_test_map.json`, accounting, ledgers/schedules, fixture manifests | Backend unit/store/process/browser selectors | Prove moved selectors still execute exact product paths; do not rely on labels alone | critical | Evidence mapping is verification accounting, not runtime architecture. |

Current owner/implementation risks requiring explicit treatment:

- `security.go` carries `Offset`, and route handlers page by array offset, while NF-REQ-125 through NF-REQ-126c require opaque live-authorized keyset cursors.
- `effectiveSort` returns only `source_row_number` when no sort is supplied and does not append the owner-defined default tail; `isSortField` omits the public ID sort keys admitted by the contract.
- routes derive one cursor and one safe-digest key from the auth master key under fixed `master-derived-v1` IDs, while Core 04 requires distinct configured key rings, rotation, and bounded decrypt-only retention.
- `trimUnicodeWhitespace` uses the Go runtime's `unicode.IsSpace`, and timestamp parsing does not implement the full closed precision, IANA ruleset, and NetFlow uptime variants required by the adopted owner.
- Phase 12 accounting reports complete coverage, but several selector functions share broad structural helpers or fall through to fixture evidence. Those rows cannot be treated as sufficient behavior characterization without inspecting what product path executes.

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Binding store reads and scans the sibling-owned `indicators` table | `binding_store.go` SQL and `scanIndicatorRecord` | Indicator schema changes can silently break Network Flow; ownership is duplicated | `must_fix` | `indicators` read/transaction participant | Define a narrow indicator lookup/find-or-create port before moving code. |
| Network Flow store locks the sibling-owned `incidents` table | `store.go` `SELECT id FROM incidents ... FOR UPDATE` | Atomicity depends on another module's storage layout | `must_fix` | `incidents` transaction/lock participant | Preserve lock ordering through an owner-provided capability. |
| Network Flow writes `deployment_admin_audit_events` directly | `store.go` audit insert | Audit storage, retention, and transactional delivery are coupled to implementation tables | `must_fix` | Core audit delivery port | Separate event construction from owner-provided transactional persistence. |
| Indicator route reaches `s.store.pool` and constructs `indicators.NewStore` | `indicator_link.go` | Route layer owns persistence composition and sibling internals | `must_fix` | Application assembly and indicators participant | Inject the participant through the application facade. |
| Route and graph handlers begin/coordinate transactions and side effects | `routes.go`, `indicator_link.go`, `graph.go` | Idempotency, audit, mutations, and WS ordering are implicit | `must_fix` | Network Flow application service/UoW | Specify transaction steps and characterize rollback/replay before extraction. |
| Import facade depends on concrete `imports.Store` and platform jobs | `import_facade.go`, `internal/app/runtime.go` | Generic import persistence leaks into extension implementation | `should_fix` | Imports owner facade and job port | Retain owner facade operations while replacing concrete dependencies. |
| Broad exported `Store` is both repository and cross-owner coordinator | Exported store types/methods across `store.go`/`binding_store.go` | Callers can bypass application invariants | `should_fix` | Private Network Flow repositories behind application facade | Inventory callers, introduce narrow interfaces, then retire unused exports. |
| HTTP admission, authorization, resources, and application behavior share files/package state | `routes.go`, `query.go`, `resources.go`, `indicator_link.go` | Refactors can reorder observable failures and side effects | `should_fix` | Thin HTTP adapter plus Network Flow application service | Extract by characterized operation, not by mechanical file split alone. |
| Frontend manually duplicates Network Flow wire types and schema IDs | `networkFlowClient.ts` versus generated protocol contracts | Contract drift can compile unnoticed | `should_fix` | Non-generated protocol/view-contract adapter | Reuse generated types through an authored adapter; never edit generated files. |
| Network Flow event hook owns generic WebSocket lifecycle | `useNetworkFlowExtensionEvents.ts` | Reconnect/resume mechanics can diverge from workbook collaboration | `should_fix` | Workbook collaboration client plus Network Flow interpreter | Preserve message semantics while moving generic lifecycle behind a shared seam. |
| Network Flow client owns generic import-session choreography | `networkFlowClient.ts` | Upload/job/session behavior can diverge across import targets | `should_fix` | Imports frontend coordinator plus Network Flow mapping adapter | Split generic session flow from Network Flow mapping/apply payload construction. |
| Direct SQL for `network_flow_*` tables | Store files and schema ownership manifest | SQL is persistence-adjacent but currently owner-aligned | `intentional/no_action` | Private `networkflow` repository | Keep ownership; isolate it rather than moving it to platform storage. |
| Flow-specific composition calls approved ephemeral Graph Projection facade | `graph.go`; `boundary_guard_test.go` allowlist | Moving composition would blur authoritative source ownership | `intentional/no_action` | `networkflow` composer and `graphprojection` engine | Split for cohesion only; keep the existing semantic boundary. |
| Owner-local harness controls import harness runtime | `harnesscontrol/*`, harness profile assembly, TH-HARNESS-REQ-664 | Moving them into production or generic platform would erase semantic ownership | `intentional/no_action` | `networkflow/harnesscontrol` | Preserve guarded routes, reset hooks, scans, and harness-only registration. |
| No direct grid-vendor coupling exists | No matching imports in Network Flow frontend | Introducing a vendor during refactor would expand scope and UI risk | `intentional/no_action` | Current feature UI | Keep the current rendering contract. |
| Cursor, key lifecycle, Unicode, and timestamp behavior differs from owners | Section 4 evidence and RB-001 through RB-003 | A behavior-preserving refactor could entrench nonconformance; a correction changes behavior | `defer` | Named behavior/security owners | Require explicit later authorization and dedicated characterization/correction slices. |
| Phase labels and fixture rows can overstate runtime characterization | Phase 12 selector switches and harness owner rules | Refactor could pass accounting without preserving behavior | `defer` | Network Flow tests plus harness accounting | Close RB-004 before affected implementation slices. |

No test-only helper import was found in production Network Flow code. The fixed `master-derived-v1` key IDs are production assembly behavior, not harness fixtures, but they remain a security-owner mismatch. No generated file appears hand-edited; the risk is future refactors changing owner inputs without running Make-owned generation/drift gates.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, snapshot, write scope, and handoff ledger | This tracker and owner documents only | Tracker-only documentation checks | Snapshot, authority order, and restrictions recorded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-04 | Keep a one-row-per-file live inventory and caller/dependency map | `internal/modules/networkflow/**`, app/frontend callers | `rg` inventory and boundary inspection | All 28 files and external consumers accounted for. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-03, WF-05 | Freeze every observable route, storage, UI, generated, and harness contract | Owner specs, contracts, routes/resources/frontend | `make json-shape-check`; contract inspection | Every risk has an owner and current test posture. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05 | Distinguish runtime characterization from fixture/accounting evidence | Network Flow tests, Phase 12 map/fixtures | `make backend-unit`; `make backend-store`; selected browser evidence | RB-004 closed or affected slices remain blocked. |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-05 | Identify cross-owner SQL, transport, transaction, platform, frontend, and generated seams | Store/routes/import/indicator/graph/frontend and boundary manifests | Backend/frontend boundary targets | Findings classified with proposed owners. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define narrow application facade, repositories, and owner ports without wire changes | Network Flow and adjacent owner interfaces | Design review against contract freeze map | Interface changes and side-effect ordering are decision-complete. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Order the smallest reversible backend, frontend, and behavior-gated slices | Packages listed in Section 7 | Slice-specific Make targets | Every slice has dependency, rollback, and binary exit gate. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Update only authored evidence owners when tests/selectors move | Phase 12 map, fixture/accounting owners; generated ledgers downstream | Drift, harness-contract, phase slices | No hand-edited generated ledger or schedule. |
| WF-08 | Validation and final handoff | chain | WF-02, WF-03, WF-04, WF-05, WF-06, WF-07 | none | Execute narrow-to-broad validation and leave a resumable ledger | Tracker plus implementation diff in later task | Section 8 commands | Result roots, failures, skips, and next action recorded. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | WF-03 | Add direct characterization for cursor comparators, effective sort, ID sorts, security key selection, Unicode trimming, timestamp variants, and selector-to-runtime execution; do not change production behavior | Network Flow package tests, route/store integration, frontend tests, authored Phase 12 map only if selectors are added | Tests may expose current conformance failures | New focused unit/integration/browser cases plus all existing tests | `make backend-unit`<br>`make backend-store`<br>`make browser-e2e-webserver-backed` | Revert test additions and authored map changes together; never edit generated ledgers | RB-004 is closed and RB-001 through RB-003 have executable failing/passing characterization. |
| SL-01 | SL-00 and later behavior authorization | Correct keyset/effective-sort, key-ring lifecycle, closed Unicode, and timestamp/uptime behavior in separate atomic sub-slices | Query/security/config/app assembly/parser/mapping plus owner contract inputs only if required | Intentional observable behavior correction | Characterization from SL-00 becomes conformance coverage | `make phase-slice PHASE=phase12`<br>`make service-backed-slice PHASE=phase12` | Revert one behavior family at a time, including its tests/config | Each correction matches its adopted owner and is marked `requires later authorization`. |
| SL-02 | SL-00; unaffected areas may proceed while SL-01 remains blocked | Introduce a narrow internal Network Flow application facade and dependency interfaces while existing adapters delegate unchanged | New/private files under `networkflow`, app assembly, existing routes/import facade | Constructor and failure-order drift | Existing route/import/store characterization; facade dependency tests | `make backend-unit`<br>`make backend-module-boundary-check` | Revert facade and assembly wiring as one slice; retain existing concrete path until callers migrate | All operations enter through one application seam and public HTTP resources are byte-equivalent. |
| SL-03 | SL-02 | Split private table/row/diagnostic/binding repositories; replace direct incident lock and audit-table writes with owner ports without schema or SQL-semantic changes | `store.go`, `binding_store.go`, incident/audit owner interfaces, app assembly | Transaction order, locks, limits, immutability, audit occurrence | Store lifecycle/immutability, rollback, audit count, lock-order tests | `make backend-store`<br>`make backend-module-boundary-check` | Keep migrations unchanged; revert repository/port wiring together | Network Flow SQL touches only owner tables and all store/audit characterization passes. |
| SL-04 | SL-02, SL-03 | Replace concrete imports and indicators stores with owner-provided preview/apply, lookup, and find-or-create transaction participants | `import_facade.go`, `indicator_link.go`, `imports`, `indicators`, app assembly | Atomic import, indicator visibility, link replay, dedupe | Import source-change/all-rejected/atomicity and indicator replay/role/atomicity tests | `make phase-slice PHASE=phase12`<br>`make service-backed-slice PHASE=phase12` | Retain old adapters until all callers use ports, then remove them in the same verified slice | No Network Flow code reads sibling tables or constructs sibling stores. |
| SL-05 | SL-02 | Split pure flow graph composition/contributor recomputation from the approved `ProjectEphemeral` adapter | `graph.go`, private composer/adapter files, `routes.go` wiring | Graph IDs/digests, ordering, limits, errors, metadata, no-retention guarantee | Existing graph/Graph Projection tests plus focused composer golden cases | `make backend-unit`<br>`make phase-slice PHASE=phase12` | Keep adapter request/response fixtures stable; revert file split without data changes | Only the adapter imports Graph Projection and graph outputs/errors remain identical. |
| SL-06 | SL-02 through SL-05 | Make route registration and HTTP serialization thin; move operation orchestration behind the application facade | `routes.go`, `resources.go`, `query.go`, `indicator_link.go`, app wiring | Admission/auth/error order, idempotency, audit, WS publication | All eleven route families, strict JSON, roles, replay, audit, WS tests | `make service-backed-slice PHASE=phase12`<br>`make backend-module-boundary-check` | Migrate one route family per commit/slice and revert independently | Routes contain transport/admission only and every frozen route contract passes. |
| SL-07 | SL-00 and stable backend adapter | Split frontend generated-contract adapter, import coordinator, event interpreter, controller state, and presentation components without changing rendered behavior | `apps/web/src/networkFlow/*`, shared import/collaboration services, non-generated packages | Schema drift, stale/superseded state, roles, selectors, reconnect behavior | Existing frontend/unit/browser tests plus superseded response and reconnect cases | `make frontend-typecheck`<br>`make frontend-unit`<br>`make frontend-import-boundary-check`<br>`make browser-e2e-webserver-backed` | Preserve the old component entry point and selector builders until the split passes | No manual duplicate wire definitions remain where generated types exist; UI and selectors are unchanged. |
| SL-08 | Any slice that moves tests/selectors or owner inputs | Update authored Phase 12 maps/manifests and regenerate downstream ledgers/schedules through Make only | `tools/phase12_test_map.json` and other authored owners; generated outputs only through generators | Evidence loss, over-selection, hand-edited generated files | Selector resolution, fixture ownership, moved-test accounting | `make json-shape-check`<br>`make harness-contract`<br>`make generate-drift`<br>`make phase-ledger-drift`<br>`make phase-schedule-drift` | Revert owner inputs and generated results together; never patch generated files | Exact row/scenario/owner coverage is unchanged or explicitly increased. |
| SL-09 | SL-01 through SL-08 as applicable | Run final narrow-to-broad verification and update this handoff with results, deviations, and remaining blockers | Tracker and later authorized implementation diff | False completion claims or stale retained evidence | All affected tests and contract/boundary checks | `make agent-finalize` then `make phase-slice PHASE=phase12`, followed by `make check` when risk warrants | No rollback beyond reverting the failing implementation slice; never mark complete on a failed gate | All required gates pass, result roots are recorded, and the intentional diff is clean. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| tracker/documentation | `make lint-markdown` | Authored Markdown structure | yes | Required for this tracker-only task. |
| tracker policy | `make generated-artifact-policy-check` | Confirms generated roots/policy remain intact | yes | Required for this tracker-only task. |
| tracker/shape | `make json-shape-check` | Contracts, manifests, accounting, and JSON bootstrap shapes | yes | Required for this tracker-only task because the tracker cites active contract/harness state. |
| unit | `make backend-unit` | Pure backend Network Flow algorithms and Phase 12 selectors | yes | Establish a characterization baseline before backend slices. |
| integration | `make backend-store` | Network Flow store and service-backed persistence behavior | yes | Requires Postgres/object store according to live target guidance. |
| integration | `make service-backed-slice PHASE=phase12` | Phase 12 store/process/browser/support rows requiring services | yes | Use after checking `make task-guide ROLE=feature-dev PHASE=phase12`. |
| e2e/browser | `make browser-e2e-webserver-backed` | Network Analysis route-to-browser behavior | yes | Live Phase 12 browser evidence target. |
| generated drift | `make generate-drift` | Contract-derived Go/TS/OpenAPI and other generated inputs | no | Required whenever contract or generator owner inputs change; also run `make json-shape-check` and `make harness-contract`. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module ownership and harness assembly boundaries | yes | Add `make frontend-import-boundary-check`, `make frontend-typecheck`, and `make frontend-unit` for frontend slices. |
| phase slice | `make phase-slice PHASE=phase12` | All Phase 12 authoritative/support/supplemental dependencies | no | Primary later implementation gate after narrow checks. |
| full check | `make check` | Repository-wide developer verification gate | no | Run after `make agent-finalize` when later implementation risk warrants the broad gate. |

Command discovery was performed through the live Make surface. No raw Go, pnpm, Vitest, Playwright, or direct script command is proposed. For this tracker-only session, `make agent-finalize` is deliberately skipped: it may refresh harness-maintenance files, which would violate the one-file write authorization, and no successful full-run `RESULTS_DIR` was supplied.

Tracker-only validation results:

| Command | Result | Result root |
| --- | --- | --- |
| `make lint-markdown` | PASS | No retained result root emitted |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260713T193621Z-p1194898` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260713T193626Z-p1195089` |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| NF-PLAN-001 | Establish authority, snapshot, scope, and tracker | WF-00 | DONE | none | This tracker, source hierarchy, repository snapshot, tracker-only validation results | Tracker-only validations passed and the handoff log is current. |
| NF-PLAN-002 | Inventory all 28 target files and callers | WF-01 | DONE | NF-PLAN-001 | Section 2 | Every target file has a responsibility/owner/risk row. |
| NF-PLAN-003 | Freeze public/runtime contracts | WF-02 | DONE | NF-PLAN-002 | Section 4 and contract inputs | Every discovered contract has an owner and test posture. |
| NF-PLAN-004 | Identify characterization gaps | WF-03 | DONE | NF-PLAN-003 | RB-001 through RB-004 | Gaps are explicit and not disguised as refactor decisions. |
| NF-PLAN-005 | Classify module boundaries and coupling | WF-04 | DONE | NF-PLAN-002 | Sections 3 and 5 | Every finding is classified and has a proposed owner/action. |
| NF-PLAN-006 | Define facade/ownership redesign | WF-05 | DONE | NF-PLAN-003, NF-PLAN-004, NF-PLAN-005 | Sections 3, 5, and 7 | Future implementation decisions are sequenced and behavior-preserving by default. |
| NF-PLAN-007 | Define slice and harness/accounting sequence | WF-06, WF-07 | DONE | NF-PLAN-006 | Sections 6 through 8 | Every slice has dependencies, validation, rollback, and exit criteria. |
| NF-BEH-001 | Correct keyset cursor and effective-sort behavior | WF-03, WF-06 | DONE | NF-PLAN-004, SL-00 | `query.go`, `query_sql.go`, `security.go`, migration `00033`; Phase 12 and service-backed result roots | Live accepted-row, diagnostic, and contributor continuations use owner tuples and `nfc2`; offset helpers are absent. |
| NF-SEC-001 | Implement configured cursor/safe-digest key-ring lifecycle | WF-03, WF-06 | DONE | NF-PLAN-004, SL-00 | `keyring.go`, `secretpurpose/registry.go`, app/config assembly, key-ring schema and tests | Claimed startup is fail-closed; rings are independent, rotatable, purpose-separated, and purged on time boundaries. |
| NF-NORM-001 | Close Unicode and timestamp/uptime behavior gaps | WF-03, WF-06 | DONE | NF-PLAN-004, SL-00 | generated Unicode 17.0 and tzdb-2026c components, timestamp/text tests | Owner vectors and provenance checks pass without reparsing committed resources. |
| NF-TEST-001 | Add direct characterization missing from Phase 12 selector accounting | WF-03 | DONE | NF-PLAN-004 | accounting v2, explicit selector switches, runtime fixture executor, Phase 12 results | Every acceptance row maps exactly and structural-only evidence is rejected. |
| NF-REF-001 | Introduce application facade and owner ports | WF-05, WF-06 | DONE | NF-TEST-001 | `module.go`, `application.go`, `ports.go` | Runtime assembly and imports enter through one root module façade. |
| NF-REF-002 | Split persistence and cross-owner participants | WF-06 | DONE | NF-REF-001 | incident, indicator, audit, and transaction participants; module-boundary gate | Network Flow SQL is owner-local and cross-owner work uses injected participants. |
| NF-REF-003 | Split graph composer and adapter | WF-06 | DONE | NF-REF-001 | `graph_projection_adapter.go`, Graph Projection boundary guard | The approved adapter is the sole Graph Projection import. |
| NF-REF-004 | Thin HTTP routes and serialization | WF-06 | DONE | NF-REF-001, NF-REF-002, NF-REF-003 | application use cases and route integration/service-backed results | Routes no longer access pools or construct sibling stores; all eleven route families retain their public contracts. |
| NF-REF-005 | Split frontend protocol/controller/event/import surfaces | WF-06 | DONE | NF-TEST-001 | services contract adapter, shared import coordinator, controller, event interpreter, unit/browser results | Generated contracts are adapted through the permitted layer with selectors and visible behavior preserved. |
| NF-HARN-001 | Replace and activate Network Flow accounting v2 | WF-07 | DONE | Affected implementation slice | accounting v2 schema/input, harness checks, explicit fixture runtime evidence | JSON shape, harness contract, generation, ledger, schedule, and Phase 12 gates pass. |
| NF-HAND-001 | Validate implementation and finalize handoff | WF-08 | DONE | All selected refactor slices | Sections 10, 11, and 13 | Retained-run finalization, the 138-work-unit repository check, and the 14-work-unit release gate pass with the result roots recorded in Section 13. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Authority, write scope, tracker, and tracker-only validation complete | Tracker touched; framework, Network Flow NLSpec, Core 00 through Core 04, Graph Projection NLSpec, Testing Harness NLSpec, domain and design documents inspected | `git status --short`; `git branch --show-current`; `git rev-parse HEAD`; `sed`; `rg`; `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | Clean starting snapshot recorded; all three tracker-only validations passed; no owner contradiction | RB-001 through RB-004 block affected implementation | A later authorized task begins with SL-00 characterization. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | All 28 files inventoried; target diagnosed as a legitimate but mixed-responsibility extension module | Tracker touched; all `internal/modules/networkflow/**`, app assembly, schema/boundary manifests, Graph Projection guard inspected | `rg --files`; symbol/import/SQL searches; `sed`; `jq` | Cross-owner SQL, pool leakage, transaction coordination, and approved graph/harness seams identified | RB-001 through RB-004 | Begin SL-00 only under a later authorized implementation task. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Network Analysis frontend is one large feature controller/client/event surface; no grid vendor coupling found | Tracker touched; six Network Flow frontend/browser files, workbook shell/startup/collaboration consumers, UI selector builders inspected | `sed`; `rg` import/type/route/state/selector searches | Manual wire duplication and generic import/WebSocket mechanics identified; current HTML-table UI retained | RB-004 for behavior coverage; UI redesign is deferred | Implement SL-07 after characterization and backend adapter stability. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Active Network Flow contract family and derived outputs mapped; no generated edit proposed | Tracker touched; `contracts/network-flow/*`, contract indexes, OpenAPI, generated Go/TS, UI contracts inspected | `jq`; `rg`; `sed` | Eleven routes, import owner facade, schemas, errors, and generated consumers frozen | RB-001 through RB-003 are implementation/owner mismatches | If owner inputs later change, edit owners and run Make generation/drift; never patch generated files. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Phase 12 claims 107 authoritative rows plus support/supplemental evidence; harness controls are owner-local and intentional | Tracker touched; Network Flow tests/harness controls, Phase 12 map/accounting, Testing Harness owner sections inspected | `rg` test symbols; `sed` selector switches; `make task-guide ROLE=phase-author PHASE=phase12`; `make explain-phase PHASE=phase12` | Public validation targets discovered; accounting distinguished from runtime characterization | RB-004 | Add focused characterization before refactoring affected behavior; update authored maps only if selectors move. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Route-time role checks exist, but cursor/safe-digest configuration and continuation differ from adopted owners | Tracker touched; `security.go`, route assembly, config validation, Core 04 key-ring sections, route tests inspected | `rg` for cursor/offset/key derivation/config; `sed` exact implementations | No deployment-admin bypass found; one-key master-derived assembly and offset cursors confirmed | RB-001, RB-002, RB-003 | Obtain later authorization, add exact tests, then correct each behavior family atomically. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T15:22:38-04:00 | Codex planning session | Planning is decision-complete; implementation remains unauthorized | Tracker touched; all evidence surfaces listed in Section 1 inspected | `make help`; `make help-all`; task/phase/target explanation commands; repository searches | WF-00 through WF-08 and SL-00 through SL-09 are sequenced | RB-001 through RB-004 | A later task should start with SL-00, not a production file move. |

The rows above preserve the original planning-session history. The authorized remediation session that followed changed the owner documents, contracts, runtime, tests, frontend, migrations, generated outputs, harness inputs, and this tracker as summarized below.

### Remediation implementation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T17:00:00-04:00 | Codex remediation session | G-01 through G-08 implemented and focused validation complete | Network Flow owner documents/contracts; backend/app/platform owners; web feature/services/shared code; migration `00033`; accounting/schema/generation inputs; tracker | Focused specification, generation, backend, frontend, Phase 12, service-backed, and browser Make targets recorded in Sections 11 and 13 | All focused gates pass; RB-001 through RB-004 resolved | None | Run final repository-wide gates and record their roots. |
| 2026-07-13T21:56:34-04:00 | Codex remediation completion | Final acceptance selectors made criterion-explicit; claimed Phase 8 harness composition supplied guarded test-only key rings; release-evidence owner import corrected | Network Flow tests, Phase 8 workbook startup test, release-evidence harness, and this tracker | `make agent-finalize`, `make check`, and `make release-check` roots recorded in Section 13 | All 138 repository check work units and all 14 release work units pass; 1,177 release-gate tests pass | None | Handoff complete. |

## 11. Closed Questions and Blockers

| ID | Binding decision and implemented resolution | Direct completion evidence | Current status |
| --- | --- | --- | --- |
| RB-001 | Offset continuation was removed. Accepted rows use `row_keyset_v1`, diagnostics use the complete `diagnostic_keyset_v1` tuple, and graph contributors use `contributor_keyset_v1` with workspace table order. `effective_sort_v1` includes sortable IDs and the complete default tail; SQL uses whitelisted expressions, lexicographic after predicates, and `limit + 1`. Old tokens have no decoder. | `make backend-unit` at `.cartulary/test-results/20260713T213618Z-p1448123`; `make backend-store` at `.cartulary/test-results/20260713T211911Z-p1374172`; Phase 12 and service-backed roots below; migration drift at `.cartulary/test-results/20260713T212119Z-p1386318`. | `RESOLVED` |
| RB-002 | Claimed assembly now requires `cartulary.network_flow_key_rings.v1` selected by `network_flow_activity.key_ring_manifest_path`. Independent AES-GCM cursor and HMAC safe-digest capabilities are injected through the module, use `nfc2`, enforce startup rotation state, share a cross-purpose bootstrap registry, and purge retired material at the owner equality boundary. Authentication-master derivation and fallback Network Flow issuance were removed. | Backend unit root above; config, key-ring, tamper, rotation, expiry, purpose-reuse, and purge tests; JSON shape at `.cartulary/test-results/20260713T212003Z-p1382147`; generated-contract drift at `.cartulary/test-results/20260713T212023Z-p1383717`. | `RESOLVED` |
| RB-003 | Network Flow now uses an explicit Unicode White_Space set and generated Unicode 17.0 NFC tables, plus an embedded verified tzdb-2026c bundle and row-aware closed timestamp engine for RFC3339, epoch, and NetFlow uptime modes. Host/runtime timezone and Unicode tables are not authoritative. Committed resources are not backfilled. | Backend unit root above; provenance generation/drift roots above; Phase 12 root `.cartulary/test-results/20260713T213636Z-p1449892`; service-backed root `.cartulary/test-results/20260713T213701Z-p1465626`. | `RESOLVED` |
| RB-004 | Accounting moved atomically to `cartulary.network_flow_activity_accounting.v2`. All 107 rows carry owner text, behavior class, required evidence kinds, exact selectors, and supplemental fixtures. Selector fallthrough is fatal; each acceptance selector invokes an explicit product path, and fixture inputs execute through admission/canonicalization only as supplemental evidence. All 28 frozen fixtures retain byte/revision identity because no committed fixture byte changed. | Harness contract at `.cartulary/test-results/20260713T212038Z-p1385280`; Phase 12 and service-backed roots above; webserver-backed browser root `.cartulary/test-results/20260713T212254Z-p1420637` with 97 tests; phase ledger/schedule drift roots `.cartulary/test-results/20260713T212147Z-p1390282` and `.cartulary/test-results/20260713T212147Z-p1390284`. | `RESOLVED` |

There is no unresolved owner contradiction. The prior `docs/domain.md` adoption drift is corrected, and all four implementation blockers have passed their direct resolution gates.

## 12. Binary Completion Criteria

This planning tracker is complete only when all of the following are true:

- [x] Every file in `internal/modules/networkflow` is inventoried or explicitly out of scope.
- [x] Every discovered public contract risk has an owner and test posture.
- [x] Every proposed workflow has dependencies and exit criteria.
- [x] Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`.
- [x] Validation commands are discovered or marked `TODO` with a reason.
- [x] Contradictions are marked `BLOCKED: owner contradiction`; none were found.
- [x] Repository/framework mismatches are recorded as planning findings.
- [x] Handoff sections are current enough for another agent to continue without rediscovery.

Implementation completion requires the focused and broad gates in Section 13. RB-001 through RB-004 remain visible as resolved historical decisions so later work cannot silently reintroduce offsets, shared key purposes, runtime-dependent canonicalization, or structural-only evidence.

## 13. Remediation Closure Summary

### Implemented outcome

- Network Flow document version is `1.1.0` with contract major `1`; Core 01 retains ownership of the unchanged public extension-discovery response, while closed contract-discovery artifacts carry version and route metadata.
- The claimed deployment contract is complete: one atomic manifest owns independent cursor and safe-digest rings, startup is fail-closed, secret purposes are checked across authentication, telemetry, enterprise authentication, and Network Flow, and key material is never emitted.
- Query continuation is route-specific keyset pagination backed by migration `00033`; no legacy offset or cursor compatibility path remains.
- Unicode 17.0 NFC and tzdb-2026c are generated from pinned, hash-verified source inputs and embedded for deterministic runtime use.
- The backend is composed through one root module façade, application-owned transactions, consumer ports, provider-owned participants, and one Graph Projection adapter. Network Flow SQL touches only `network_flow_*` tables.
- Frontend contract adaptation, import coordination, controller state, collaboration interpretation, and presentation orchestration are separate seams; direct generated-protocol imports remain confined to permitted adapter layers.
- Harness accounting v2, explicit acceptance selectors, runtime fixture execution, and negative evidence checks replace structural or fixture-only fallthrough.

### Focused validation completed

| Gate | Result root |
| --- | --- |
| `make backend-unit` | `.cartulary/test-results/20260713T213618Z-p1448123` |
| `make backend-store` | `.cartulary/test-results/20260713T211911Z-p1374172` |
| `make backend-module-boundary-check` | `.cartulary/test-results/20260713T211905Z-p1373967` |
| `make frontend-unit` | `.cartulary/test-results/20260713T211719Z-p1370259` |
| `make frontend-typecheck` | `.cartulary/test-results/20260713T211831Z-p1373390` |
| `make frontend-import-boundary-check` | `.cartulary/test-results/20260713T211812Z-p1372616` |
| `make phase-slice PHASE=phase12` | `.cartulary/test-results/20260713T213636Z-p1449892` |
| `make service-backed-slice PHASE=phase12` | `.cartulary/test-results/20260713T213701Z-p1465626` |
| `make browser-e2e-webserver-backed` | `.cartulary/test-results/20260713T212254Z-p1420637` |

### Final repository validation

| Gate | Result root | Result |
| --- | --- | --- |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260713T214430Z-p1667120` | `.cartulary/test-results/20260713T215610Z-p1920799` | Pass; retained successful check evidence refreshed, generated maintenance unchanged |
| `make check` | `.cartulary/test-results/20260713T214430Z-p1667120` | Pass; 138/138 work units and 1,120 tests |
| `make release-check` | `.cartulary/test-results/20260713T215634Z-p1922938` | Pass; 14/14 work units and 1,177 tests, including a fresh 138-work-unit check |

The first broad check exposed and led to correction of one staticcheck violation and one claimed test-harness composition gap. The first release attempt completed all tests but exposed a harness import error in release-evidence emission; the owner import was corrected, `harness-contract` passed, and the complete release target then passed. No required gate remains failing or skipped.

## 14. Network Analysis Grid Adapter Adoption

### 14.1 Activation, authority, and baseline

This section is the controlling plan for adopting `@cartulary/grid-adapter`
in the Network Analysis extension workspace. Sections 1 through 13 remain the
completed historical module-refactor record. The initial planning update did
not claim implementation completion; the live workstream table and committed
completion checkpoints below now own the current execution state.

| Item | Approved planning value |
| --- | --- |
| Planning parent | `00cd433470376f4d6ffcf2c7b90b6c25a9f3cd8f` |
| Execution baseline commit | `6503537f55a6cde6fd5bfb2e85d975a015968f0a` |
| Baseline branch | `revision/grid-adapter` |
| Baseline worktree | Clean |
| Controlling artifact | This existing tracker; ownership is unambiguous and no second tracker is permitted |
| Normative authority | Core 00 through Core 04, then the adopted Network Flow Activity NLSpec inside its extension boundary |
| Vocabulary authority | `docs/domain.md`; Network Flow tables and rows are extension analytical resources, not Core record envelopes, workbook projections, view schemas, or saved views |
| Design direction | `docs/design.md`; preserve `extension_workspace` identity, fixed-height virtualization, keyboard operation, non-color states, and shared density/styling |
| Harness authority | `docs/testing-harness-nlspec.md`; it owns invocation, target selection, scheduling, artifacts, cleanup, and evidence accounting |
| Adapter reference | Completed F-RDG workstreams in `docs/handoffs/grid-adapter-module-refactor-tracker.md` |
| Current execution scope | Full Section 14 remediation, serially checkpointed |
| Overall adoption status | `ACTIVE` (`NF-GA-00` through `NF-GA-08` `COMPLETE`; `NF-GA-09 PENDING`) |

The inspected baseline has the following implementation boundaries:

| Concern | Current owner and behavior | Adoption consequence |
| --- | --- | --- |
| Accepted and rejected tables | `NetworkAnalysisWorkspace.tsx` renders manual HTML tables from fixed first-page controller results | Replace the accepted-row and diagnostic tables with semantic adapter grids; do not treat visual row numbers as identity |
| Grid behavior | Network Analysis has no grid-adapter import or grid state | Introduce adapter use only after the semantic extension boundary is available |
| Query behavior | Network Flow client/controller requests fixed 50-row pages without exposed filters, ordered sorts, or continuation navigation | Add owner-field queries and opaque-cursor navigation; the server remains authoritative |
| Selection | Graph state auto-selects the first edge; there is no semantic active cell, range, or inspector | Require explicit graph selection and add extension-owned active-cell and inspector state |
| Editing | Flow rows are immutable; corrections use reimport or table deletion | Do not expose editors, draft rows, paste, fill, or row creation |
| Visualization | The graph controller and workspace own a module-specific graph/edge presentation | Keep graph visualization outside the grid adapter; use the adapter only for contributor rows |
| Commands | Import and one graph-edge create-indicator path exist; table rename/delete, row linking, existing-indicator linking, and explicit table/graph selection are incomplete | Add only owner-adopted, role-gated module commands; do not enable generic bulk selection |
| Evidence | The Phase 12 browser file uses synthetic `page.setContent` surfaces, and component tests are not yet authoritative live-grid evidence | Add real production-grid evidence and reconcile Phase 12 accounting only in the final workstream |

The current Network Flow owner overrides generic grid design on route admission:
incident closure, authorization loss, and table soft delete make data
non-queryable and invalidate cursors. The UI must clear protected rows and
selection state on those transitions rather than retaining generic read/copy
access to a formerly authorized snapshot.

### 14.2 Approved adapter boundary

Network Analysis must not supply a fabricated `viewSchemaId`, `recordId`, or
row version. The adapter is generalized atomically around these public semantic
interfaces:

```ts
type GridSurfaceIdentity =
  | { kind: "view_schema"; viewSchemaId: string }
  | {
      kind: "extension_grid";
      extensionProfileId: string;
      workspaceKey: string;
      gridSchemaId: string;
    };

type GridRowIdentity =
  | { kind: "core_record"; recordId: string }
  | {
      kind: "extension_resource";
      extensionProfileId: string;
      resourceKind: string;
      resourceId: string;
    };

type GridMutationIdentity = {
  kind: "core_row_version";
  baseRowVersion: number;
};

type GridCellAnchor = {
  surface: GridSurfaceIdentity;
  rowIdentity: GridRowIdentity;
  fieldKey: string;
};
```

The approved adapter migration is:

- rename `WorkbookDataGrid` and `WorkbookDataGridProps` to
  `SemanticDataGrid` and `SemanticDataGridProps`;
- replace `GridRecordRow` with `GridDataRow`, carrying `rowIdentity` and an
  optional `mutationIdentity`;
- replace record-specific active-row callbacks with semantic row identities;
- keep bulk selection explicitly Core-record-only and product-command-gated;
- require `mutationIdentity` before an editor, paste, or fill target can be
  compiled;
- migrate existing workbook consumers atomically, without compatibility
  aliases, deprecated wrappers, or surface exception lists; and
- retain the React Data Grid import boundary and stylesheet singleton.

Application contracts, persisted state, focus anchors, and test selectors may
contain only the semantic identities above. React Data Grid indexes, native
handles, classes, DOM structure, and coordinates remain package-private.

### 14.3 Network Flow presentation and query model

Specification closure will amend the Network Flow NLSpec to document version
`1.2.0` while retaining contract major `1`. Existing route request and response
shapes remain unchanged. Core 01 adds one additive, side-effect-free
`POST /api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview`
route because the current imports service discards the owner facade's required
preview result. Closed repo-local derived artifacts at
`contracts/network-flow/presentation.v1.json` and
`contracts/network-flow/mapping-registry.v1.json` will be referenced by the
Network Flow contract index and frontend entrypoints, but not added to
`public_schema_ids`.

The preview route accepts only `target_kind`, `extension_profile_id`,
`owner_mapping_schema_id`, and `owner_mapping`; the server derives the source
capability, content hash, discovered columns, and unit coordinates. Its closed
generic wrapper returns the target/profile identifiers, owner result schema ID,
and validated `network_flow_import_preview_result_v1`. It persists no mapping,
selection, table, or audit occurrence. Existing `PUT .../mapping` remains the
explicit durable approval route, and apply is admitted only when its returned
mapping fingerprint equals the last displayed preview fingerprint.

The artifact owns three grid-schema identities:

- `network_flow.accepted_rows.v1`;
- `network_flow.rejected_rows.v1`; and
- `network_flow.graph_contributors.v1`.

Each schema declares stable field keys, label keys, value and renderer kinds,
filter and sort capabilities, copy and link eligibility, default visibility
and order, widths, minimum widths, and inspector-only fields. Generated
frontend metadata is downstream of that owner input and must contain no vendor
coordinate, class, handle, or DOM contract.

Accepted-row defaults are:

| Position | Field | Default width | Default posture |
| ---: | --- | ---: | --- |
| gutter | `source_row_number` | `64px` | Structural display only; not row identity |
| 1 | `network_flow.flow_start_utc` | `184px` | Visible |
| 2 | `network_flow.flow_end_utc` | `184px` | Visible |
| 3 | `network_flow.src_ip` | `168px` | Visible and linkable |
| 4 | `network_flow.src_port` | `88px` | Visible |
| 5 | `network_flow.dst_ip` | `168px` | Visible and linkable |
| 6 | `network_flow.dst_port` | `88px` | Visible |
| 7 | `network_flow.ip_protocol` | `104px` | Visible |
| 8 | `network_flow.bytes_count` | `120px` | Visible |
| 9 | `network_flow.packets_count` | `120px` | Visible |
| 10 | `network_flow.input_interface` | `144px` | Visible |
| 11 | `network_flow.output_interface` | `144px` | Visible |
| 12 | `network_flow.exporter_id` | `144px` | Hidden by default |
| 13 | `network_flow.tcp_flags` | `112px` | Hidden by default |
| 14 | `network_flow.application_label` | `160px` | Hidden by default |

Table and row identities, source and normalized digests, mapping fingerprint,
`network_flow.observation_source_ref`, and `unmapped_raw` are inspector-only.
Diagnostics default to source row, source column, field key, error code, reason
code, and message key. Contributor rows reuse accepted-row rendering and group
by table display name while using `network_flow_table_id` as group identity.

Query and layout ownership is fixed as follows:

- The server owns authorization, filter/sort validation, cursor admission,
  graph aggregation, link validation, rename, delete, and import.
- Grid sort callbacks emit ordered Network Flow field-key sorts. The adapter
  must not perform authoritative local filtering or sorting.
- Initial row and diagnostic queries omit `limit`, using the server-owned
  default. Continuations contain only their route-specific schema ID and opaque
  cursor token.
- Previous/Next navigation keeps a stack of page requests: page one reissues
  the initial query, and later previous pages reissue the cursor request that
  originally produced that page. An expired cursor resets to page one with an
  explanation.
- A row-query time window compiles to
  `network_flow.flow_end_utc >= start_utc` and
  `network_flow.flow_start_utc < end_utc`; graph queries use the owner-defined
  `time_range` object.
- Table, filter, sort, or time changes reset pagination, active range, and
  superseded request state. Late responses are discarded by request
  generation.
- Unchanged rows are reconciled by `network_flow_row_id`; visual positions and
  vendor indexes are never identities.
- Layout state is in memory, keyed by extension profile, workspace key, and
  grid-schema ID. It survives active-table changes for the same schema and
  resets on browser reload. It is not stored as a Core saved view, local-storage
  value, or server resource.

### 14.4 Adopted workflows and feature classification

Read-only range selection and copy are adopted for accepted rows, diagnostics,
and contributors. Clipboard values come from semantic field values rather than
rendered DOM. Active-cell state drives a Network Flow inspector that presents
complete values and provenance and exposes explicit IP-link commands only for
eligible source/destination fields.

Table rename, table soft delete, import, existing/create indicator linking,
graph table-scope selection, explicit graph vertex/edge selection, and the
contributor drawer remain module-owned, role-gated workflows. The graph does
not auto-select its first edge and is not compiled through the grid adapter.
Network Flow cell state remains extension-local and must not populate Core
presence `record_id` or `field_key` members.

| Classification | Adapter capabilities and decision |
| --- | --- |
| Already integrated | Package dependency, shared density/styling primitives, React Data Grid containment, semantic data-state primitives, and always-on virtualization exist repository-wide; Network Analysis has no runtime grid integration yet |
| Required for this effort | Extension-grid/resource identities; semantic component/row API; accepted, diagnostic, and contributor grids; column resize/reorder/visibility/reset; server multi-sort, filters, time window, and cursor paging; active cell and inspector; range/copy; contributor grouping; complete data/read-only/authorization states; keyboard/focus/accessibility; stable selectors; supported-load evidence |
| Useful future enhancement | Cross-session extension-layout persistence; extension-owned saved analysis configurations; richer node-edge visualization; additional flow-profile column groups; broader cross-table analysis workflows |
| Intentionally deferred | Bulk selection until an adopted bulk command exists; primary-flow-table grouping; summaries/footer aggregation; data-column freezing; dynamic row heights; RTL-specific behavior; additional adapter aggregation APIs |
| Inappropriate for the module | Existing-row editing; draft/create rows; paste; fill; row reorder; RDG-local filters/sorting; client aggregation; column spanning; nested trees; master/detail rows; vendor handles/indexes/DOM contracts; Network Flow row/cell presence |

Deferred and inappropriate capabilities require new owner authority before
enablement. Their availability in the package is not product authorization.

### 14.5 Gap register

| ID | Remediation and affected areas | Rationale and long-term benefit | Compatibility, risk, dependencies, and rollback | Exact validation and evidence |
| --- | --- | --- | --- | --- |
| `NF-GAP-01` | Close presentation semantics in NLSpec `1.2.0`; add derived grid metadata and generated frontend types. Areas: specification, contracts, generation, documentation, tests | One semantic source of truth supports future flow profiles without hard-coded surface policy | Contract major and HTTP shapes remain unchanged. First dependency. Revert specification, contract inputs, and generated outputs together. Leaving it open causes divergent columns/capabilities | All three closed grid schemas and displayed fields resolve. Run `make generate`, `make generate-drift`, `make json-shape-check`, and `make lint-markdown`. Semantic/generated evidence |
| `NF-GAP-02` | Generalize adapter surface, row, anchor, and component APIs. Areas: adapter, all consumers, tests, documentation | Prevents fake Core identities and supports later extension grids | Atomic internal TypeScript source break after `NF-GAP-01`; no shim. Revert adapter and migrated consumers together. Leaving it open violates domain boundaries | Core anchors remain stable; extension anchors contain only extension identities; mutationless rows cannot edit. Run `make frontend-unit`, `make frontend-typecheck`, `make frontend-import-boundary-check`, and `make lint-biome`. Semantic/live-grid evidence |
| `NF-GAP-03` | Replace fixed `limit=50` clients with semantic query state, filters, ordered sort, opaque-cursor stack, cancellation, and row reconciliation. Areas: implementation, tests | Enables trustworthy larger-result analysis without offsets or visual identity | Depends on metadata/adapter identities; no server contract change. Roll back to the prior first-page client. Leaving it open presents misleading partial analysis | Exact initial/continuation requests, default sort, overlap time window, Previous/Next replay, expiry recovery, and late-response discard. Webserver-backed/stateful evidence |
| `NF-GAP-04` | Compile metadata columns and add accessible session-only show/hide, move, resize, and reset. Areas: implementation, tests, design evidence | Stable fields survive labels, tables, and later profiles | Depends on metadata; no persisted-data migration. Rollback removes in-memory layout only. Leaving it open creates wide, inconsistent tables | Defaults/widths match metadata; state survives table changes, resets on reload/reset, and retains structural navigation. Semantic/live-grid/visual evidence |
| `NF-GAP-05` | Add network renderers for time, IP, port, protocol, counters, nulls, references, diagnostics, and clipboard. Areas: implementation, tests | Preserves network semantics while keeping renderers module-owned | Depends on columns; each renderer rolls back independently. Leaving it open produces ambiguous or lossy combined values | Complete accessible values; canonical IPs; decimal-string counter copy; provenance only in inspector. Unit/live-grid/accessibility/visual evidence |
| `NF-GAP-06` | Add active cell, inspector, explicit graph selection, existing/create linking, rename, delete, import, and contributor workflows. Areas: implementation, tests | Connects semantic selection only to adopted analysis commands | Depends on identities/queries; commands are independently reversible. Leaving it open risks stale or label-based targets | Requests use row/field/vertex/edge IDs and current graph digest; roles and confirmation match owners; presence stays workspace-only. Webserver/stateful evidence |
| `NF-GAP-07` | Configure every Network Flow grid read-only; expose range/copy but omit editor, paste, fill, draft, and bulk-selection props. Areas: specification, implementation, tests, tracker | Immutable flows are corrected through reimport/delete | Depends on adapter gating; range/copy can roll back independently. Leaving it open risks unsupported mutations | Keyboard/pointer input cannot edit, paste, or fill; copy works when authorized; no mutation is emitted. Semantic/live-grid/accessibility evidence |
| `NF-GAP-08` | Use adapter grouping only for graph contributors; retain server graph aggregation and module visualization. Areas: implementation, tests, documentation | Implements the one adopted grouping workflow without copying Timeline behavior | Depends on graph selection; contributor grouping can roll back to a flat semantic list. Leaving it open risks client aggregation or misleading primary grouping | Groups follow workspace table order and table identity; rows retain row IDs; sums remain server output. Unit/live-grid/webserver evidence |
| `NF-GAP-09` | Map loading, refreshing, empty, filtered-empty, stale, unavailable, permission-denied, and read-only states; clear protected state when route admission fails. Areas: implementation, tests | Prevents stale disclosure and distinguishes retryable conditions | Depends on query/auth events. Rollback must still clear protected data. Leaving it open is a disclosure risk | State precedence, retry, announcements, cursor invalidation, and clearing pass for role loss, closure, soft delete, and hidden resources. Stateful/accessibility evidence |
| `NF-GAP-10` | Add semantic focus restoration, keyboard control, announcements, token styling, non-color cues, and stable selectors. Areas: implementation, tests, design evidence | Makes the virtualized workspace operable without vendor DOM knowledge | Depends on anchors/states; styling rolls back separately and focus with the adapter slice. Leaving it open causes inaccessible navigation and brittle tests | Focus restores to row/field or deterministic fallback; inspector close returns focus; keyboard, ARIA, live-region, and non-color scenarios pass. Accessibility/visual evidence |
| `NF-GAP-11` | Validate fixed-height row/column virtualization at 1,000 rows and all declared columns, including stable row references. Areas: implementation, measurement tests, documentation | Establishes a supported page envelope while stored tables remain server-paged | Depends on completed grids; rollback removes reconciliation/measurement changes, not paging. Leaving it open risks DOM expansion and lost selection | Rendered rows/cells remain fewer than supplied rows/cells; endpoints are reachable; unchanged row objects retain identity; no timed claim. Measurement/live-grid evidence |
| `NF-GAP-12` | Enforce vendor imports, semantic styling, and selectors at package boundaries. Areas: implementation, tooling, tests | Preserves vendor replacement and semantic test ownership | Applies throughout; each violation rolls back with its consumer. Leaving it open leaks vendor contracts | No `react-data-grid` import outside the adapter and no vendor index/class/handle in app state or selectors. Run `make frontend-import-boundary-check` and `make lint-biome`. Boundary evidence |
| `NF-GAP-13` | Replace synthetic-only browser claims with real application/grid scenarios; update Phase 12 maps, accounting, visual registry, and retained evidence. Areas: tests, harness inputs, documentation, generated accounting | Aligns evidence with the production grid | Final dependency; authored maps and generated ledgers/schedules roll back together. Leaving it open permits false conformance claims | Every authoritative row names a real target/artifact and evidence classes remain distinct. Harness, phase, browser, and retained-run evidence |
| `NF-GAP-14` | Add the generic mapping-preview route, generated mapping registry, staged import coordinator, explicit ordinal-aware mapping modal, preview diagnostics, and approval-fingerprint binding. Areas: Core and extension specification, public contract, backend, frontend, tests | Makes the already-computed owner preview reachable and gives future analytical import targets a clean Core-owned preview boundary | One additive contract-major-1 route; existing approval/apply shapes remain unchanged; no database migration. Leaving it open forces hard-coded, unreviewed mappings | Preview is side-effect-free; duplicate/empty headers remain distinct; safe samples and blocking diagnostics render; suggestions never approve; apply requires the latest fingerprint. Backend/frontend/service-backed evidence |
| `NF-GAP-15` | Complete the required workspace header, inner tabs, accepted/diagnostic regions, diagnostic summary, filters, graph panel, contributor drawer, status strip, and role-gated import/rename/delete controls. Areas: implementation, tests, documentation/design evidence | Produces one coherent operational workspace instead of disconnected manual tables | Visual/interaction migration only. Leaving it open hides lifecycle and diagnostic context and prevents required workflows | Every normative region exists; active-table selection and role gates match the owner; lifecycle commands use stable table identity/version. Live-grid/webserver/accessibility/visual evidence |

### 14.6 Mandatory execution protocol and checkpoints

Workstream order is mandatory. `NF-GA-00` starts `ACTIVE`; every later
workstream starts `PENDING`. The allowed later states are `ACTIVE`, `BLOCKED`,
and `COMPLETE`. Before a slice begins, its tracker entry changes to `ACTIVE`.
A workstream becomes `COMPLETE` only after its implementation, exit gates, and
checkpoint are committed here; the next workstream uses that checkpoint commit
as its baseline and must not begin earlier. A failed exit gate is recorded as
`BLOCKED`, never skipped.

Every checkpoint must populate these fields:

- status and reviewer/date;
- baseline commit and end commit;
- files changed;
- commands and exact results;
- result roots and artifacts;
- evidence-map rows added or changed;
- residual risks and deferred features;
- compatibility/migration outcome;
- rollback boundary and whether it was exercised; and
- next workstream.

### 14.7 Serial implementation workstreams

#### `NF-GA-00` — Tracker activation and specification closure

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Dependencies | None |
| Remediation | Amend Network Flow NLSpec to `1.2.0` and Core import/security owners; define presentation and mapping metadata, additive generic mapping preview, session layout, data states, read-only behavior, supported page envelope, and every adopted/deferred/rejected capability; update derived contract inputs and generated frontend/backend metadata |
| Affected areas | Core and extension specification, public and repo-local contracts, generation inputs/outputs, documentation, semantic tests |
| Rationale | Close owner semantics before implementation and prevent Core-record/view-schema substitution |
| Long-term benefit | Later profiles and extension grids reuse one closed semantic presentation model |
| Compatibility/migration | Contract major remains `1`; one additive preview route; no existing route-shape, database, dependency, lockfile, or saved-view migration |
| Unresolved risk | Code would otherwise choose presentation defaults and capabilities ad hoc |
| Validation criteria | All decisions in Sections 14.2 through 14.5 are owned by the amended specification or derived metadata without Core/domain/design contradiction |
| Commands/evidence | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-markdown`; `git diff --check`; semantic/generated evidence |
| Rollback | Revert NLSpec, index/frontend-entrypoint inputs, presentation artifact, and generated output together |
| Exit criteria | Owner approval and clean generation/drift |
| Checkpoint | Completed in Section 14.7.1 |
| Next workstream | `NF-GA-01` |

##### 14.7.1 `NF-GA-00` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-15 America/New_York |
| Baseline commit | `6503537f55a6cde6fd5bfb2e85d975a015968f0a` |
| Implementation end commit | `cd429f1a701e617305f1a5c5a318dcdc7b81ea1b` |
| Files changed | Network Flow and Core owner specifications; Network Flow index/routes/frontend entrypoints; OpenAPI; presentation and mapping registries; contract-index schema and shape checker; protocol generator/facade and generated Go/TypeScript artifacts; generated execution-topology render index; this tracker |
| Substantive result | Adopted document version `1.2.0`; closed three-grid presentation registry and source-profile mapping registry; additive side-effect-free Core mapping-preview route and wrapper; explicit separation of preview from durable approval; generated frontend metadata; serial `NF-GA-00..09` tracker protocol; removed obsolete accessibility-preflight target |
| Required gates | `make generate` PASS (`.cartulary/test-results/20260716T032711Z-p64652`); `make generate-drift` PASS (`.cartulary/test-results/20260716T032718Z-p65990`); `make generated-artifact-policy-check` PASS (`.cartulary/test-results/20260716T032723Z-p67531`); `make json-shape-check` PASS (`.cartulary/test-results/20260716T032723Z-p67680`); `make lint-markdown` PASS; `git diff --check` PASS |
| Additional gate | `make frontend-typecheck` PASS |
| Corrective run | Initial `make json-shape-check` stopped because the shape-checker edit made generated phase schedules stale (`.cartulary/test-results/20260716T032649Z-p63709`); `make phase-schedules` regenerated owner outputs and passed (`.cartulary/test-results/20260716T032657Z-p64086`), after which shape validation passed |
| Evidence-map rows | No claim rows changed; specification/generated-contract evidence only. Phase schedule render metadata regenerated from its owner input |
| Compatibility / migration | Contract major remains `1`; one additive public route; no existing response change, database, dependency, lockfile, saved-view, or RDG migration |
| Residual risk / deferral | Runtime route, adapter, frontend workflow, and real evidence remain in later workstreams; richer graph and rejected capabilities remain deferred/rejected by Section 14.10 |
| Rollback | Revert `cd429f1a` atomically; not exercised because all gates passed |
| Next workstream | `NF-GA-01`; it may activate only after this checkpoint commit |

#### `NF-GA-01` — Adapter boundary and semantic model

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `dad8633301f91450fafb4544c6f94d4b6e06a35a` |
| Dependencies | `NF-GA-00 COMPLETE` |
| Remediation | Implement the generalized semantic interfaces and `SemanticDataGrid`; atomically migrate existing workbook consumers; retain Core-record-only bulk selection and mutation gating |
| Affected areas | `packages/grid-adapter`, workbook consumers, adapter semantic/live-grid tests, boundary documentation |
| Rationale | Remove Core-only names and identities from a reusable adapter boundary |
| Long-term benefit | Extension workspaces can adopt the grid without vendor leakage or fake Core resources |
| Compatibility/migration | Internal TypeScript source break; no shim, dependency, lockfile, or RDG upgrade |
| Unresolved risk | Fake identities would leak into contracts and compound future migrations |
| Validation criteria | Existing Core behavior remains identical; extension anchors are semantic; rows without mutation identity cannot edit |
| Commands/evidence | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome`; `git diff --check`; semantic/live-grid evidence |
| Rollback | Revert adapter and all consumer migrations atomically |
| Exit criteria | Existing workbook evidence and new extension-resource fixtures pass |
| Checkpoint | Completed in Section 14.7.2 |
| Next workstream | `NF-GA-02` |

##### 14.7.2 `NF-GA-01` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `dad8633301f91450fafb4544c6f94d4b6e06a35a`; activation `463162bab88192773cd3338fb9259ebd7c28a3ea` |
| Implementation end commit | `a352e93ddefe916fe5edaf0d5c22f2f78fa7f1b1` |
| Files changed | `packages/grid-adapter` semantic model, RDG compiler/component, facade, test support, and semantic/live-grid tests; all `apps/web` generic, Entity, Assessment, and Timeline adapter consumers and tests; adapter guidance; protocol facade version expectation; frontend import-boundary fixtures; test-accounting classification |
| Substantive result | Replaced the Core-specific adapter facade atomically with `SemanticDataGrid`, identity-neutral surfaces/rows/anchors/ranges, optional Core mutation identity, and Core-only bulk/mutation capabilities. Extension rows now render, group, select, focus, range-select, navigate, and copy with extension-resource identities; compile-time and runtime gates reject edit, paste, fill, draft, action, and bulk capabilities on extension surfaces. Vendor keys remain package-private and no compatibility alias or deprecated export remains |
| Required gates | `make frontend-unit` PASS (`.cartulary/test-results/20260716T040305Z-p10212`); `make frontend-typecheck` PASS; `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T040547Z-p14311`); `make lint-biome` PASS; `git diff --check` PASS |
| Additional gates | `make format` PASS (`.cartulary/test-results/20260716T040146Z-p5241`); `make json-shape-check` PASS (`.cartulary/test-results/20260716T040547Z-p14346`) after adding the two new semantic tests to accounting |
| Corrective runs | The first atomic-migration type check failed as expected while consumers still used the removed facade (`.cartulary/test-results/20260716T033808Z-p74938`); migration completed before the final pass. Early unit runs exposed stale row-shape assertions, the prior `1.1.0` document expectation, and then unmapped new tests; those fixtures and accounting were corrected, with the final passing root above. The first lint run found formatting/import-order drift (`.cartulary/test-results/20260716T040137Z-p4776`); `make format` resolved it. All failures were related to this change and none remain |
| Evidence-map rows | Added support-only Phase 12 classifications for `resolves extension anchors, ranges, grouping, and navigation without Core aliases` and `rejects duplicate extension identities and all mutation target compilation`; live-RDG tests remain support evidence under the existing adapter test-file classification. No extension-conformance or public claim row was added |
| Compatibility / migration | Intentional monorepo-internal TypeScript break completed atomically across every Core consumer; existing Core behavior remains covered. No deprecated export, wrapper, fake identity, database change, dependency/lockfile update, saved-view change, or RDG upgrade |
| Residual risk / deferral | Network Analysis has not adopted the adapter yet; that occurs after query ownership is established in later workstreams. Cross-session layout, Network Flow mutations, and vendor identities remain deferred or rejected under Section 14.10 |
| Rollback | Revert `a352e93d` atomically to restore the former adapter and all consumers; not exercised because all exit gates passed |
| Next workstream | `NF-GA-02`; it may activate only after this checkpoint commit |

#### `NF-GA-02` — Import preview and explicit mapping workflow

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `e9dc61566fba85d27c8a4efad2a4a29f9dda53d1` |
| Dependencies | `NF-GA-01 COMPLETE` |
| Remediation | Implement the additive Core mapping-preview route and validated owner wrapper; generate mapping registries; refactor import coordination into upload/discover, preview, and approve/select/apply; replace the fixed payload with an explicit ordinal-aware modal and fingerprint binding |
| Affected areas | Core and extension import services, generated mapping registry, shared frontend coordinator, Network Analysis UI, backend/frontend/service-backed tests |
| Rationale | Make the existing owner preview result reachable without turning preview into durable approval |
| Long-term benefit | Future analytical import targets reuse one safe, side-effect-free Core preview boundary |
| Compatibility/migration | Additive contract-major-1 route; existing approval/apply shapes unchanged; no database migration |
| Unresolved risk | Hard-coded mappings bypass explicit approval, duplicate-header resolution, and preview diagnostics |
| Validation criteria | Suggestions never approve; preview has no durable effects; safe samples and ordinals render; mapping changes invalidate preview; durable approval fingerprint must match before apply |
| Commands/evidence | Backend import tests; `make frontend-unit`; `make frontend-typecheck`; `make test-fast`; focused Phase 12 import evidence |
| Rollback | Revert route, registry consumers, coordinator, and modal together; existing durable import route remains intact |
| Exit criteria | Explicit mapping and preview work end to end for valid, conflicting, rejected, stale, and unauthorized inputs |
| Checkpoint | Completed in Section 14.7.3 |
| Next workstream | `NF-GA-03` |

##### 14.7.3 `NF-GA-02` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `e9dc61566fba85d27c8a4efad2a4a29f9dda53d1`; activation `aa3dadd450de5b5835bc14ee01b0faaeb6385fc4` |
| Implementation end commit | `de1494573ca437f75cc4b1891c981701b749ac71` |
| Files changed | Core Imports API/facade/routes/store and integration evidence; Network Flow mapping/import facade and contract evidence; mapping-registry generator and generated Go/TypeScript artifacts; protocol facade; three-stage shared import coordinator and tests; Network Flow mapping model/modal/controller/workspace tests; semantic selector facade; frontend inventory and test-accounting classification; NLSpec omission clarification |
| Substantive result | Added the closed Core-owned `POST .../mapping-preview` route with editor/reviewer/admin admission, server-derived source capability, strict owner-result validation, and no mapping, selection, table, idempotency, or apply side effect. Renamed the owner boundary to the semantically complete `ExtensionImportFacade`. Generated Go mapping-registry tables and refactored backend defaults, aliases, required fields, transforms, empty policies, derivations, and supported policies to consume them. Replaced the production fixed-header payload with explicit upload/discover, preview, and fingerprint-bound approve/select/apply stages. Added an accessible ordinal-aware modal with profile, timestamp, unknown-column, explicit-ignore, required-count, safe-sample, preview-count, diagnostic, and fingerprint presentation. Mapping changes invalidate the preview; mismatched durable approval stops before selection/apply; successful apply selects the exact returned table resource |
| Required gates | `make generate` PASS (`.cartulary/test-results/20260716T041211Z-p20328`); `make frontend-unit` PASS (`.cartulary/test-results/20260716T044310Z-p51853`); `make frontend-typecheck` PASS; final `make test-fast` PASS with 1,233 tests (`.cartulary/test-results/20260716T044935Z-p9866`); focused `make backend-integration-support` PASS after adding viewer denial and closed-request evidence |
| Additional gates | `make generate-drift` PASS (`.cartulary/test-results/20260716T044621Z-p96786`); `make json-shape-check` PASS (`.cartulary/test-results/20260716T044621Z-p96790`); `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T044621Z-p96771`); `make lint-biome` PASS; `make lint-markdown` PASS; `make format` PASS (`.cartulary/test-results/20260716T044849Z-p5876`); `git diff --check` PASS |
| Corrective runs | The first final typecheck exposed the expected remaining imports of the removed one-shot coordinator (`.cartulary/test-results/20260716T042857Z-p87825`) and then static selector ownership omissions (`.cartulary/test-results/20260716T043443Z-p89621`); the coordinator consumers and semantic selector facade were completed. The first unit run passed product tests but correctly failed on four unmapped mapping-model scenarios (`.cartulary/test-results/20260716T043533Z-p90694`); bounded regression accounting was added. Two subsequent workspace runs exposed unsupported Chai DOM matchers and one missing selector-helper import (`.cartulary/test-results/20260716T044156Z-p47771`, `.cartulary/test-results/20260716T044230Z-p49786`); assertions/imports were corrected. Early fast runs caught a missing Go import and an NLSpec MAY-omission sentence; both were related and corrected. No failure remains |
| Evidence-map rows | Added bounded `unowned_regression` classifications for generated-registry mapping-model tests and the real component-level explicit preview/returned-table-selection scenario. Extended Core service-backed integration evidence to cover exact closed request shape, viewer denial, replay-stable preview, absent durable unit/idempotency/table effects, and preview-to-approval fingerprint equality. No synthetic browser claim was promoted or relabeled; authoritative Phase 12 browser reconciliation remains `NF-GA-09` |
| Compatibility / migration | One additive contract-major-1 route. Existing durable `PUT .../mapping`, selection, apply, and response shapes are unchanged. No database, dependency, lockfile, saved-view, or RDG migration. The internal owner-facade rename is atomic with no compatibility alias |
| Residual risk / deferral | Production browser evidence for the modal and replacement of synthetic Phase 12 claims remain in later workstreams, especially `NF-GA-09`. Role-loss/incident-closure clearing while a modal is open belongs to `NF-GA-05`; keyboard focus trapping/restoration belongs to `NF-GA-07`. No persistent preview cache was added |
| Rollback | Revert `de149457` atomically to remove the additive route, registry consumer, coordinator, and modal while retaining the prior durable import route; not exercised because all exit gates passed |
| Next workstream | `NF-GA-03`; it may activate only after this checkpoint commit |

#### `NF-GA-03` — Query, pagination, reconciliation, and structured errors

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `c426c6c305c3dc04a3441179f89e9e788803f59f` |
| Dependencies | `NF-GA-02 COMPLETE` |
| Remediation | Add typed initial/continuation requests, structured API errors, owner filters, ordered sort, opaque-cursor replay, cancellation/generation guards, and immutable-row reconciliation; remove fixed `limit=50` and message substring parsing |
| Affected areas | Network Flow client/controllers, shared HTTP error boundary, module tests, live browser scenarios |
| Rationale | Provide scalable, stable-field analysis without client query authority |
| Long-term benefit | New flow types, columns, and larger result sets extend through metadata and server paging |
| Compatibility/migration | Reuse existing routes/cursors; no wire change; layout resets on reload |
| Unresolved risk | Partial or locally reordered results may mislead analysts |
| Validation criteria | Exact initial/continuation bodies, default sort, overlap time window, cursor history, expiry recovery, stale-response discard, structured errors, and stable row references |
| Commands/evidence | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; semantic/live-grid/webserver/stateful evidence |
| Rollback | Revert Network Analysis integration while retaining the generalized adapter |
| Exit criteria | Accepted/diagnostic grids operate against real routes with no fixed 50-row ceiling |
| Checkpoint | Completed in Section 14.7.4 |
| Next workstream | `NF-GA-04` |

##### 14.7.4 `NF-GA-03` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `c426c6c305c3dc04a3441179f89e9e788803f59f`; activation `5246596cd6c7ec9bacb35ccb11927efa68f1a089` |
| Implementation end commit | `ab8b41e65ab9fc638769bee3894b89fb3fdfadba` |
| Files changed | Network Flow query client and accepted/rejected controllers; generic paged-query hook; semantic query compiler and reconciliation model; structured error model and generated error-registry facade; graph, table, and indicator-link error consumers; Network Analysis pagination UI and semantic selectors; component, query, error, and hook tests; frontend test-accounting classifications |
| Substantive result | Replaced fixed 50-row first-page requests with closed typed initial and continuation unions. Initial accepted and rejected requests omit `limit` and therefore use the server-owned default 200; continuation bodies contain only their continuation schema ID and opaque cursor. Added registered accepted/diagnostic filters, ordered multi-sort pass-through, endpoint-IP support, overlap-window compilation, cursor history with exact previous-request replay, expiry/mismatch restart with explanation, cancellation and generation guards, and immutable owner-ID reconciliation that preserves object references only for contract-equal resources. API failures now retain status, code, field, reason, retry action, retryability, and safe public message. Authorization loss is classified from exact structured status/code/reason signals rather than message substrings |
| Required gates | Final `make frontend-unit` PASS (`.cartulary/test-results/20260716T052058Z-p15985`); `make frontend-typecheck` PASS; `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T051459Z-p73630`); `make browser-e2e-webserver-backed` PASS with 97 tests (`.cartulary/test-results/20260716T051512Z-p74797`); `make browser-e2e-stateful` PASS with 22 tests (`.cartulary/test-results/20260716T051733Z-p95284`) |
| Additional gates | `make lint-biome` PASS; `make format` PASS (`.cartulary/test-results/20260716T051417Z-p69248`); `git diff --check` PASS |
| Corrective runs | The first unit run was manually interrupted after a render loop prevented report emission (`.cartulary/test-results/20260716T050401Z-p59884`); a diagnostic single-worker rerun confirmed the same related loop (`.cartulary/test-results/20260716T050948Z-p63864`). The cause was an unstable inline error callback retriggering the disabled graph-clearing effect; restoring the stable state setter removed the loop. A temporary skipped-test diagnostic was removed before the passing full suite. No failed assertion or gate remains |
| Evidence-map rows | Added bounded `unowned_regression` classifications for query compilation/reconciliation, structured error preservation, and cursor/cancellation/generation semantics. Extended the real Network Analysis component test to prove the production initial accepted-row body is exactly the schema ID with no fixed limit. Existing webserver-backed and stateful suites remained green. No synthetic browser claim was promoted; final Phase 12 claim reconciliation remains `NF-GA-09` |
| Compatibility / migration | Existing query routes, cursor contracts, response shapes, and server authority are unchanged. The client now honors the advertised default page envelope instead of forcing 50. No database, dependency, lockfile, saved-view, contract-major, or RDG migration |
| Residual risk / deferral | Visible query/filter/sort controls, active range and inspector reset, and filtered-empty presentation depend on the metadata-driven grids in `NF-GA-04`. Protected-state clearing for authorization/lifecycle transitions belongs to `NF-GA-05`. Contributor pagination and graph scope remain `NF-GA-06`; no client sorting, filtering, or aggregation was introduced |
| Rollback | Revert `ab8b41e6` atomically to restore the former first-page controllers while retaining the generalized adapter and import-preview work; not exercised because all exit gates passed |
| Next workstream | `NF-GA-04`; it may activate only after this checkpoint commit |

#### `NF-GA-04` — Workspace grids, renderers, layout, and normative regions

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `3a5bbb8f0b297c0dc0f791a871800b49575d87ff` |
| Dependencies | `NF-GA-03 COMPLETE` |
| Remediation | Compile three grid schemas; replace accepted/diagnostic HTML tables; add metadata columns, network renderers, session layout, active-cell inspector, canonical range/copy, and every normative workspace region |
| Affected areas | Network Flow presentation/controllers, grid metadata compiler, styles, tests |
| Rationale | Establish one coherent metadata-driven operational workspace |
| Long-term benefit | Later columns and profiles extend through stable metadata and renderer kinds |
| Compatibility/migration | Session layout only; no row mutation or persisted client state |
| Unresolved risk | Manual tables and incomplete regions keep required context and workflows unavailable |
| Validation criteria | All regions and default columns resolve; layout/reset, clipboard values, inspector provenance, density, and shell geometry pass |
| Commands/evidence | `make frontend-unit`; `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make browser-e2e-visual`; semantic/live-grid/webserver/stateful/visual evidence |
| Rollback | Commands, inspector, contributor grid, and renderers are separately reversible |
| Exit criteria | Every normative workspace region exists; accepted and rejected resources use metadata-compiled production grids; default density, fixed-height rows, session layout, accessible full values, canonical copy, and active-cell inspection pass |
| Checkpoint | Completed in Section 14.7.5 |
| Next workstream | `NF-GA-05` |

##### 14.7.5 `NF-GA-04` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `3a5bbb8f0b297c0dc0f791a871800b49575d87ff`; activation `89ceb87f72c53f6648c22c5e782d7a55d3c31798` |
| Implementation end commit | `30fc01d1bf068ff9989375ae197195429e20b749` |
| Files changed | Network Analysis workspace and component tests; new accepted/rejected query controls; new semantic-grid frames and active-cell inspector; new generated-metadata presentation compiler/renderers and tests; new session-layout controller and tests; mapping-modal diagnostic rendering; Network Flow contract facade; semantic selector facade; frontend test-accounting classifications |
| Substantive result | Replaced the accepted-row and rejected-diagnostic manual HTML tables with `SemanticDataGrid` extension surfaces and immutable extension-resource row identities. Presentation metadata now owns column order, visibility, widths, minimums, filter/sort/copy capabilities, and inspector-only fields. Source row is a non-identity gutter for accepted rows. Network renderers preserve canonical timestamps, IPs, nullable ports/text, protocol numbers, decimal-string counters, TCP flags, and safe localized diagnostics while semantic clipboard values remain unformatted. Added metadata-driven show/hide/reorder/resize/reset state in browser-session memory keyed by profile, workspace, and grid schema, with no local storage or saved-view persistence. Added the complete header, table tabs, rows/rejected/graph modes, diagnostic summary, accepted/rejected filters, status strip, fixed-height grid region, pagination, and active-cell inspector. Exact component evidence proves the initial request omits `limit`, endpoint-CIDR and overlap-window filters compile to owner shapes, diagnostics use their owner query, hidden columns reset to defaults, and extension rows carry no mutation identity |
| Required gates | Final `make frontend-unit` PASS (`.cartulary/test-results/20260716T055348Z-p19594`); `make frontend-typecheck` PASS; `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T055348Z-p19635`); `make lint-biome` PASS; `make browser-e2e-webserver-backed` PASS with 97 tests (`.cartulary/test-results/20260716T055429Z-p22807`); `make browser-e2e-stateful` PASS with 22 tests (`.cartulary/test-results/20260716T055647Z-p42548`); `make browser-e2e-visual` PASS with 28 tests and no golden drift (`.cartulary/test-results/20260716T054909Z-p99714`) |
| Additional gates | `make format` PASS (`.cartulary/test-results/20260716T055336Z-p17219`); `git diff --check` PASS. The visual-golden maintenance guide was reviewed before visual validation; Phase 12 currently owns no visual row, so no unrelated golden was refreshed. Real Network Analysis visual-row ownership remains part of `NF-GA-09` evidence reconciliation |
| Corrective runs | Early frontend-unit runs exposed the expected manual-table count change and a missing active-cell fallback in the new grid (`.cartulary/test-results/20260716T053356Z-p30942`, `.cartulary/test-results/20260716T053553Z-p35962`, `.cartulary/test-results/20260716T053627Z-p38023`); the expectation and semantic row-selection fallback were corrected. A later unit run passed every product assertion but failed test accounting for the new diagnostics scenario (`.cartulary/test-results/20260716T053952Z-p47679`); the bounded row was added. The first formatting run exposed only related Biome findings in the new query/grid/presentation files (`.cartulary/test-results/20260716T053301Z-p25776`), which were corrected. Running webserver-backed and stateful browser targets concurrently caused both Vite processes to refresh the same `apps/web/dist` tree and one build failed copying a font (`.cartulary/test-results/20260716T054342Z-p56004`); the target passed when rerun alone and passed again after final edits. No product, security, or required gate failure remains |
| Evidence-map rows | Added bounded `unowned_regression` classifications for metadata compilation/rendering, extension-resource projection, session layout, accepted-grid anatomy/query behavior, and rejected-grid diagnostic query behavior. Stable selector helpers now own accepted rows, diagnostic resources, grids, filters, inspector, column menu, summary, reset, and workspace header without RDG DOM/classes. Existing webserver/stateful/visual suites remain green; authoritative Phase 12 synthetic-browser replacement remains `NF-GA-09` |
| Compatibility / migration | Visual and interaction migration only. Existing Network Flow query/import/graph/link routes and response shapes are unchanged. Layout is ephemeral browser-session state, not local storage, table metadata, Core saved views, or a server resource. No database, dependency, lockfile, saved-view, contract-major, or React Data Grid version migration |
| Residual risk / deferral | Exact authorization/lifecycle precedence, protected-state clearing, rename/delete commands, and strict role gates belong to `NF-GA-05`. Graph-scope semantics, explicit vertex selection, contributor production grid, and row/range/vertex linking belong to `NF-GA-06`. Focus restoration and complete keyboard/a11y evidence belong to `NF-GA-07`; the supported 1,000-row envelope belongs to `NF-GA-08`; real Phase 12 visual/claim accounting belongs to `NF-GA-09` |
| Rollback | Revert `30fc01d1` atomically to restore the former manual accepted/rejected presentation while retaining semantic query, adapter, and import-preview foundations; not exercised because all exit gates passed |
| Next workstream | `NF-GA-05`; it may activate only after this checkpoint commit |

#### `NF-GA-05` — Read-only safety, lifecycle controls, and authorization transitions

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `92a1de04f59bac3b7bd56db6ef4532e8f6df2baa` |
| Dependencies | `NF-GA-04 COMPLETE` |
| Remediation | Enforce mutation-free grids and exact roles; add rename/delete/import-result selection; implement data-state precedence, retry behavior, superseded-response handling, invalidation, and protected-state clearing |
| Affected areas | Network Flow controllers/workspace, authorization/collaboration interpretation, stateful tests |
| Rationale | Prevent stale disclosure while making service conditions understandable |
| Long-term benefit | Later workflows inherit one owner-correct route-admission transition model |
| Compatibility/migration | No authorization policy change; client mirrors current owner outcomes |
| Unresolved risk | Rows retained after closure or role loss are a security defect |
| Validation criteria | Every named state plus closure, soft delete, role loss, hidden resource, cursor expiry, refresh, and late-response transition; no protected residual state |
| Commands/evidence | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make service-backed-slice PHASE=phase12`; stateful/accessibility evidence |
| Rollback | State controller may roll back only if fail-closed clearing is preserved |
| Exit criteria | Authorization-loss scenarios leave no rows, cursors, range, inspector, or graph selection |
| Checkpoint | Completed in Section 14.7.6 |
| Next workstream | `NF-GA-06` |

##### 14.7.6 `NF-GA-05` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `92a1de04f59bac3b7bd56db6ef4532e8f6df2baa`; activation `a336d9d60f9b4886f156021728a3085f6d8cbbf1` |
| Implementation end commit | `cb900386b649565af71876189056fdbae9f3a093` |
| Files changed | Network Analysis workspace and component evidence; table/query/graph/collaboration controllers and lifecycle interpreter tests; structured Network Flow error model; table mutation client and contract decoder facade; semantic selector facade; protocol decoder facade; frontend test-accounting classifications; one type-safe accessibility-helper lint correction in the owner accessibility suite |
| Substantive result | Enforced the owner role matrix exactly: viewer read/copy; editor, reviewer, and admin import/rename/link; reviewer and admin soft-delete. Network Flow grids remain mutation-free and expose no editor, paste, fill, draft, create, reorder, or Core bulk-selection path. Added stable-ID, current-version rename and soft-delete commands, conflict-driven table-metadata refresh before retry, exact-name destructive confirmation, no-op feedback, and post-delete next-then-previous table selection. Added abort/generation guards for table lists and mutations. Structured authorization, incident-closure, inactive/deleted-table, and hidden-resource failures now drive explicit permission/lifecycle states without message parsing. Protected transitions immediately clear rows, diagnostics, cursors, page history, ranges/inspectors through grid teardown, graph/edge/contributor state, mapping workflow state, and table-scoped controls; collaboration authorization/session/closure events use the same fail-closed model. Non-protected refresh failures may retain authorized rows, while a replacement active table may recover only after owner metadata refresh. Role downgrade removes mutation surfaces without removing viewer-safe read/copy access. Import success continues to select the exact returned table reference |
| Required gates | Final `make frontend-unit` PASS (`.cartulary/test-results/20260716T062718Z-p72557`); `make frontend-typecheck` PASS; `make browser-e2e-webserver-backed` PASS with 97 tests (`.cartulary/test-results/20260716T062746Z-p74559`); `make browser-e2e-stateful` PASS with 22 tests (`.cartulary/test-results/20260716T063001Z-p94305`); `make browser-e2e-a11y` PASS with 20 tests (`.cartulary/test-results/20260716T063421Z-p29432`); `make service-backed-slice PHASE=phase12` PASS with 37 tests (`.cartulary/test-results/20260716T063547Z-p43948`) |
| Additional gates | `make format` PASS (`.cartulary/test-results/20260716T062707Z-p69926`); `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T063606Z-p58466`); `make lint-biome` PASS; `git diff --check` PASS |
| Corrective runs | The first formatting run found only the new dialog autofocus policy and was corrected (`.cartulary/test-results/20260716T061112Z-p67504`). An early unit run passed product assertions but exposed unmapped lifecycle evidence; accounting was added. Subsequent unit runs exposed that clearing rows also cleared the structured permission error (`.cartulary/test-results/20260716T061630Z-p78378`, `.cartulary/test-results/20260716T061756Z-p80863`); protected owner state is now retained independently from cleared data. One assertion then matched both the permission heading and message (`.cartulary/test-results/20260716T061906Z-p85636`) and was corrected to the semantic region. A late format run surfaced an existing equality-search lint diagnostic in the accessibility owner plus the new incident-reset dependency (`.cartulary/test-results/20260716T062539Z-p62205`); both were corrected, and the first type-safe equality rewrite exposed a nullable DOM type (`.cartulary/test-results/20260716T062626Z-p69346`) before the final guarded implementation. The first final accessibility run missed one pre-existing Timeline tab-order sentinel (`.cartulary/test-results/20260716T063244Z-p14749`); its isolated serial rerun passed all 20 rows at the final root above. The miss was outside Network Analysis and did not reproduce. No required gate remains failing |
| Evidence-map rows | Added bounded `unowned_regression` classifications for the exact role matrix, role downgrade, stable-ID rename/delete, version-conflict refresh, authorization-loss clearing, inactive-table clearing, controller selection/identity behavior, and incident-closure event mapping. Existing real webserver/stateful/accessibility/service-backed suites remained authoritative under their owner maps. No synthetic Network Analysis row was promoted; its replacement remains `NF-GA-09` |
| Compatibility / migration | Existing Network Flow rename/delete/query/import wire contracts and authorization policy are unchanged; the frontend now consumes the existing table-mutation result and mirrors owner outcomes. No database, dependency, lockfile, saved-view, contract-major, or RDG version migration. No legacy mutation behavior was retained because immutable analytical rows have no continuing edit value |
| Residual risk / deferral | Active/selected/all graph scope, explicit vertex selection, contributor production grid, and cell/range/vertex/edge linking remain `NF-GA-06`. Complete focus restoration, keyboard commands, announcements, non-color evidence, and stable production selectors remain `NF-GA-07`. Supported-load evidence remains `NF-GA-08`; synthetic Phase 12 claim replacement remains `NF-GA-09` |
| Rollback | Revert `cb900386` atomically to remove the new lifecycle command/transition layer; not exercised because all exit gates passed. Any partial rollback must preserve structured fail-closed protected-data clearing |
| Next workstream | `NF-GA-06`; it may activate only after this checkpoint commit |

#### `NF-GA-06` — Graph selection, contributor grid, and indicator linking

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `362f17311ecfc7103bb03f0ae3b6a1a3fca15bb2` |
| Dependencies | `NF-GA-05 COMPLETE` |
| Remediation | Add active/selected/all table scope, explicit vertex/edge selection, contributor grouping/paging, and row-cell, same-value row-range, vertex, and edge linking to existing/create indicator targets |
| Affected areas | Network Flow graph/link controllers, contributor presentation, Core Indicator integration, tests |
| Rationale | Bind adopted analysis commands to current semantic queries and immutable owner resources |
| Long-term benefit | Later graph and indicator workflows extend stable selectors rather than labels or vendor coordinates |
| Compatibility/migration | Existing server graph/link routes remain authoritative; graph visualization remains module-owned |
| Unresolved risk | Auto-selection, stale digests, or ambiguous candidate values can target the wrong resource |
| Validation criteria | Exact query/digest/IDs/field keys/canonical confirmation; no implicit selection; contributor order unchanged; range links only same canonical IP within limits |
| Commands/evidence | `make frontend-unit`; `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; focused Phase 12 graph/link evidence |
| Rollback | Graph/link controls and contributor grid are reversible without weakening read-only/auth clearing |
| Exit criteria | Every adopted graph/link command has stable semantic targeting and real-route evidence |
| Checkpoint | Completed in Section 14.7.7 |
| Next workstream | `NF-GA-07` |

##### 14.7.7 `NF-GA-06` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `362f17311ecfc7103bb03f0ae3b6a1a3fca15bb2`; activation `1b8fb66fd827833e22daf29eca1f3be8051f560c` |
| Implementation end commit | `2c2377bb35020bb41c583b468100134a8d24d71b` |
| Files changed | Network Analysis workspace and component evidence; graph, indicator-link, and grid-layout controllers; semantic row/range link compiler and tests; Network Flow client and presentation compiler/tests; contributor production grid; shared semantic-grid exact cell-anchor callback; Network Flow contract decoder facade; protocol decoder facade; semantic selector facade; frontend test-accounting classifications |
| Substantive result | Added explicit `active_table`, `selected_tables`, and `all_active_tables` graph scope with active-table default and workspace-ordered selected IDs. Graph queries carry the current owner filters and overlap time range, use abort/generation protection, and never auto-select a result. Projection IDs are translated back to stable owner endpoint/flow-edge IDs before contributor or link requests. Explicit vertex and edge selection drives exact initial/continuation contributor requests, replayable previous-page history, stale-digest invalidation, and a read-only `SemanticDataGrid` grouped by table display name plus stable table ID while retaining server contributor order and owner row identities. Added cell, same-value range, vertex, and source/destination edge linking to both existing and create-indicator targets. Row ranges compile only from exact Network Flow extension surfaces, use ordered owner row refs, enforce the advertised binding-row limit, reject mixed fields/values and foreign identities, and require byte-exact canonical confirmation. Link requests retain current semantic graph query, digest, owner IDs, linkable field key, and closed Core target variant. Graph rendering remains module-owned; no vendor graph abstraction or client sort/filter/aggregation was added |
| Required gates | Final `make frontend-unit` PASS (`.cartulary/test-results/20260716T074750Z-p20332`); `make frontend-typecheck` PASS (`.cartulary/test-results/20260716T074750Z-p20333`); final `make browser-e2e-webserver-backed` PASS with 97 tests (`.cartulary/test-results/20260716T074849Z-p23655`); `make browser-e2e-stateful` PASS with 22 tests (`.cartulary/test-results/20260716T073837Z-p59093`); focused Phase 12 rows `E-12-NFAC029-29`, `E-12-NFAC030-30`, `E-12-NFAC036-36`, and `E-12-NFAC037-37` PASS under base-phase accounting scope (`.cartulary/test-results/20260716T074317Z-p96807`) |
| Additional gates | `make format` PASS (`.cartulary/test-results/20260716T074738Z-p17969`); `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T074831Z-p22812`); `make lint-biome` PASS (`.cartulary/test-results/20260716T074831Z-p22810`); `git diff --check` PASS |
| Corrective runs | Early component runs exposed that a synthetic row callback targeted the first visible column instead of the clicked semantic cell and that the standard fetch fixture omitted the source-profile limit route (`.cartulary/test-results/20260716T072040Z-p76526`, `.cartulary/test-results/20260716T073240Z-p31024`); the adapter now compiles the clicked column anchor and the fixture serves the owner registry. A focused browser invocation passed its selected tests but failed whole-target frontend accounting because base-phase accounting was not disabled (`.cartulary/test-results/20260716T074146Z-p79569`); the corrected scoped invocation passed at the focused root above. Final range hardening first exposed an incomplete TypeScript narrowing and then a mixed-field range falling back to a single cell (`.cartulary/test-results/20260716T074608Z-p14352`, `.cartulary/test-results/20260716T074608Z-p14365`); the exact extension-identity type guard and fail-closed range logic corrected both. No product, security, or required gate remains failing |
| Evidence-map rows | Added bounded `unowned_regression` classifications for semantic cell selection, same-value range row-ref compilation, active-cell position independence, and fail-closed mixed/foreign/over-limit selection. Existing real webserver/stateful Phase 12 graph and link rows remained green, including the four focused rows above. Stable selectors now cover graph scope, owner vertex/edge identities, contributor grid/drawer, and indicator-link dialog actions without RDG DOM/classes or candidate values in selector tokens |
| Compatibility / migration | Existing graph, contributor, source-profile, and indicator-link public routes and response shapes are unchanged. The frontend now uses the advertised source-profile limit and existing closed selector/target variants. No database, dependency, lockfile, saved-view, contract-major, React Data Grid version, or durable-state migration. No compatibility wrapper, fabricated Core identity, implicit graph selection, or client-authoritative contributor ordering was retained |
| Residual risk / deferral | Deterministic focus restoration, inspector return focus, complete keyboard operation, live announcements, non-color cues, token review, and accessibility/visual evidence belong to `NF-GA-07`. The supported 1,000-row/all-column envelope belongs to `NF-GA-08`; real Phase 12 synthetic-evidence replacement and final accounting reconciliation remain `NF-GA-09` |
| Rollback | Revert `2c2377bb` atomically to remove graph-scope/link/contributor presentation changes while retaining the prior read-only lifecycle boundary; not exercised because all exit gates passed. Any partial rollback must preserve explicit selection, current graph digest validation, exact owner identities, canonical confirmation, and server-owned contributor ordering |
| Next workstream | `NF-GA-07`; it may activate only after this checkpoint commit |

#### `NF-GA-07` — Keyboard, focus, accessibility, and semantic styling

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `f80944ad8fb8c0ce809120a02c820b21798711b4` |
| Dependencies | `NF-GA-06 COMPLETE` |
| Remediation | Add keyboard query/layout controls, focus restoration, inspector return focus, announcements, non-color state cues, shared tokens, and stable semantic selectors |
| Affected areas | Grid/workspace accessibility, styles, selector builders, browser and visual tests |
| Rationale | Make virtualized analysis operable without vendor DOM knowledge |
| Long-term benefit | Semantic focus and state styling remain stable across vendor or layout changes |
| Compatibility/migration | Visual goldens change; no product-data migration |
| Unresolved risk | Inaccessible navigation and brittle vendor selectors |
| Validation criteria | Deterministic focus fallbacks, keyboard-only workflows, ARIA/live-region state, non-color cues, contrast, and stable selectors |
| Commands/evidence | `make browser-e2e-a11y`; `make browser-e2e-visual`; `make frontend-unit`; accessibility/visual evidence |
| Rollback | Visuals follow the golden-maintenance guide; focus rolls back with its adapter slice |
| Exit criteria | Automated accessibility passes and manual keyboard scenarios are recorded |
| Checkpoint | Completed in Section 14.7.8 |
| Next workstream | `NF-GA-08` |

##### 14.7.8 `NF-GA-07` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `f80944ad8fb8c0ce809120a02c820b21798711b4`; activation `c90913304d80cb290cdd09370b1c82b3b1936dc5` |
| Implementation end commit | `f3df12a4bc7c183a70262ae8d2616e3e4bbb44ea` |
| Files changed | Network Analysis workspace and component evidence; accepted/rejected semantic-grid frame and focused production-grid tests; mapping, rename, delete, indicator-link, inspector, graph-contributor, and query controls; Network Flow presentation compiler; shared semantic-grid focus handle and adapter tests; semantic selector facade; frontend test-accounting classifications |
| Substantive result | Added roving Network Flow table tabs with ArrowLeft/ArrowRight/Home/End activation, keyboard-operable query and column-layout actions, semantic per-cell selectors, and stable selector tokens for query, status, inspector, contributor, and graph announcements. Added an adapter-owned `focusRoot` escape hatch without exposing vendor coordinates. Immutable-row reconciliation now restores the same owner row/field when present, the same field on the nearest valid owner row when removed, and the grid root when a selected field or query context becomes invalid; it does not steal focus during table-tab activation. Inspector close/Escape returns to the semantic cell. Mapping, rename, delete, and indicator-link dialogs set deterministic initial focus, trap Tab/Shift+Tab, honor Escape when not busy, and return focus to the invoking control or workspace fallback. Contributor close/Escape returns to the explicitly selected graph object. Polite status/layout/graph announcements and assertive structured errors provide non-color state cues. New styling uses shared Cartulary spacing, border, surface, text, and danger tokens; Network Analysis retains its adopted fixed-height default density. No selector uses RDG DOM/classes, row indexes, candidate values, or serialized internal keys |
| Required gates | Final `make frontend-unit` PASS (`.cartulary/test-results/20260716T082105Z-p13762`); `make browser-e2e-a11y` PASS with 20 tests (`.cartulary/test-results/20260716T082343Z-p32847`); `make browser-e2e-visual` PASS with 28 tests and no golden drift (`.cartulary/test-results/20260716T082505Z-p47358`) |
| Additional gates | `make frontend-typecheck` PASS (`.cartulary/test-results/20260716T082134Z-p15823`); `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T082134Z-p15830`); `make lint-biome` PASS (`.cartulary/test-results/20260716T082134Z-p15825`); `make json-shape-check` PASS (`.cartulary/test-results/20260716T082003Z-p8933`); final `make format` PASS (`.cartulary/test-results/20260716T082100Z-p11484`); `git diff --check` PASS |
| Corrective runs | Early component validation rejected raw `data-testid` selector templates in the modal helper (`.cartulary/test-results/20260716T081027Z-p60058`) and formatting rejected a redundant landmark role (`.cartulary/test-results/20260716T080917Z-p57302`); the helper now uses the shared selector facade and named sections retain their native region semantics. A later unit run caught the initial query-reset fallback stealing focus from roving table tabs (`.cartulary/test-results/20260716T082003Z-p8934`); root focus is now conditional on a real semantic selection, and the full unit suite passed. One complete accessibility run then hit the pre-existing global Phase 11 `system-view-selector` focus race outside Network Analysis (`.cartulary/test-results/20260716T082159Z-p17088`); an unchanged rerun passed all 20 tests at the required root above. No product, security, or required gate remains failing |
| Evidence-map rows | Added bounded `unowned_regression` classifications for roving table tabs plus lifecycle-dialog focus, semantic nearest-row/root focus restoration, and keyboard column reorder announcements. Existing browser accessibility and visual matrices remained green. Component scenarios exercise mapping initial focus, trapped lifecycle dialogs, inspector return focus, graph-drawer return focus, stable production-grid cells, keyboard layout commands, and table navigation; real Phase 12 application/browser promotion remains deliberately owned by `NF-GA-09` rather than treating synthetic surfaces as authoritative |
| Compatibility / migration | The shared adapter adds one identity-neutral focus method and all monorepo consumers compile atomically. Public HTTP routes, owner resource shapes, grid identities, server authority, role policy, saved views, and persistent layout remain unchanged. No database, dependency, lockfile, React Data Grid version, contract-major, or durable-state migration. No compatibility alias, vendor selector, fabricated identity, or Network Flow mutation path was introduced |
| Residual risk / deferral | The supported 1,000-row/all-declared-column envelope, bounded DOM measurements, off-screen reachability, and continuity under load belong to `NF-GA-08`. Replacement of synthetic Phase 12 Network Analysis claims with real application/server evidence, final selector promotion, visual-registry reconciliation, and final accounting remain `NF-GA-09` |
| Rollback | Revert `f3df12a4` atomically to remove Network Analysis focus/keyboard/announcement behavior and the adapter root-focus method; not exercised because all exit gates passed. Any partial rollback must preserve modal focus containment, protected-state clearing, semantic selectors, and the absence of vendor coordinates |
| Next workstream | `NF-GA-08`; it may activate only after this checkpoint commit |

#### `NF-GA-08` — Virtualization and supported-load validation

| Field | Approved value |
| --- | --- |
| Status | `COMPLETE` |
| Activation baseline | `5d35d388a084cfdb954a9874a58d46e2e639f728` |
| Dependencies | `NF-GA-07 COMPLETE` |
| Remediation | Validate 1,000-row/all-column fixed-height virtualization, bounded DOM, scrolling, refresh reconciliation, range/inspector continuity, and server paging |
| Affected areas | Adapter diagnostics, Network Flow reconciliation, measurement and live-grid tests, handoff evidence |
| Rationale | Support growth without loading multi-million-row tables client-side |
| Long-term benefit | Establish one reusable supported page envelope and stable-reference policy |
| Compatibility/migration | No resource-limit or server-default change |
| Unresolved risk | Large pages may expand the DOM or lose semantic selection |
| Validation criteria | Rendered rows/cells remain fewer than supplied values; first/last rows are reachable; unchanged row objects retain identity; no vendor coordinates persist |
| Commands/evidence | `make browser-e2e-measurement`; `make frontend-unit`; `make browser-e2e-webserver-backed`; measurement/live-grid evidence without a timing claim |
| Rollback | Remove optimization-specific reconciliation/measurement fixtures; retain semantic paging |
| Exit criteria | Artifacts record the supported envelope without a Core 05 claim |
| Checkpoint | Completed in Section 14.7.9 |
| Next workstream | `NF-GA-09` |

##### 14.7.9 `NF-GA-08` completion checkpoint

| Checkpoint field | Recorded outcome |
| --- | --- |
| Status / reviewer / date | `COMPLETE`; Codex self-review; 2026-07-16 America/New_York |
| Baseline and activation commits | Baseline `5d35d388a084cfdb954a9874a58d46e2e639f728`; activation `409eac8d7e33d700be31217093fce88e61b0dddd` |
| Implementation end commit | `5dcfec46f3044eb951f38a7d160f75dff4c020c6` |
| Files changed | Network Flow query reconciliation, presentation projection, graph contributor controller, production semantic grids and tests; debug-harness load fixture through the workbook feature facade; semantic row/cell selector contracts; Phase 12 implementation guide, manifest, accounting, generated ledger/topology index, and real measurement scenario; obsolete `NF-AC-050` Go selector removed |
| Substantive result | Accepted rows, rejected diagnostics, and graph contributors now reconcile immutable owner resources by stable extension identity and reuse both unchanged owner objects and their `GridDataRow` wrappers. Changed resources replace only their matching wrapper, removed resources disappear, incoming server order remains authoritative, and contributor refreshes no longer replace every object. A debug-only fixture composes the actual three production grid components at 100 and 1,000 deterministic resources, exposes every declared non-inspector column through production layout controls, and performs equivalent refresh through the same application-owned reconciliation. No adapter-owned Network Flow policy, eager whole-table load, client sort/filter/aggregation, or test-only authorization route was added |
| Supported envelope | At the fixed 578-by-1,246 viewport, both 100 and 1,000 resources mounted exactly 22 accepted rows/242 cells, 22 rejected rows/132 cells, and 21 contributor rows/231 cells. Accepted scroll width was 1,992 pixels, contributor width 2,056 pixels, and 1,000-resource scroll heights were 32,032, 32,032, and 32,160 pixels respectively. The test reached first/last rows and the far-right application-label cell, traversed grouped contributors, and retained active field, two-row range, focus, and inspector state across an equivalent refresh. Counts are retained in the Phase 12 measurement stdout artifact and carry no timing threshold or Core 05 claim |
| Evidence ownership | `E-12-NFAC050-50` is now the authoritative explicit-only `measurement` row for the existing `NF-AC-050` evidence-classification criterion. The previous Go test only restated the selector and was removed. The Network Flow accounting owner and implementation guide now point at the real Playwright scenario; generated Phase 12 ledger/schedule inputs were regenerated. The row remains engineering evidence and is excluded from default warm checks |
| Required gates | `make frontend-unit` PASS (`.cartulary/test-results/20260716T113020Z-p5892`); `make browser-e2e-measurement` PASS with Phase 3 plus the new Phase 12 measurement (`.cartulary/test-results/20260716T112709Z-p72220`); `make browser-e2e-webserver-backed` PASS with 97 scenarios (`.cartulary/test-results/20260716T112755Z-p86724`) |
| Additional gates | `make service-backed-slice PHASE=phase12 ROWS=E-12-NFAC050-50` PASS, 1/1 work unit and one test (`.cartulary/test-results/20260716T113203Z-p11995`); `make frontend-typecheck` PASS; `make frontend-import-boundary-check` PASS (`.cartulary/test-results/20260716T113020Z-p5897`); `make lint-biome` PASS; `make json-shape-check` PASS (`.cartulary/test-results/20260716T113139Z-p11539`); `make generate-drift` PASS (`.cartulary/test-results/20260716T113231Z-p25662`); `make phase-ledger-drift` PASS (`.cartulary/test-results/20260716T113231Z-p25668`); `make phase-schedule-drift` PASS (`.cartulary/test-results/20260716T113232Z-p25681`); `make generated-artifact-policy-check` PASS (`.cartulary/test-results/20260716T113020Z-p5964`); final `make format`, `make lint-markdown`, and `git diff --check` PASS |
| Corrective runs | Initial JSON-shape validation correctly required phase schedule regeneration (`.cartulary/test-results/20260716T112008Z-p12726`), and the import boundary rejected direct debug-shell access to the Network Flow module (`.cartulary/test-results/20260716T112008Z-p12741`); the fixture now enters through `NetworkFlowFeature`. The first exact slice attempt (`.cartulary/test-results/20260716T112357Z-p36752`) exposed that a supplemental row could not be selected by the authoritative measurement runner. Rather than misclassify support evidence, the real measurement replaced the no-op unit selector as `NF-AC-050`'s criterion owner. A later shape run then found the stale separate Network Flow accounting pointer (`.cartulary/test-results/20260716T113020Z-p6017`); its authored selector was corrected and the final shape/slice runs passed. No product failure remains |
| Compatibility / migration | Public routes, request/response schemas, server query defaults and limits, extension resource identities, read-only policy, saved views, persistent layout, database schema, dependencies, lockfiles, and React Data Grid version are unchanged. The debug query fixture is non-product harness composition. Phase 12 evidence ownership changes from a local no-op Go selector to explicit browser measurement, so old retained evidence cannot close the current row |
| Residual risk / deferral | Real claimed/unclaimed application workflows, secure startup runtime profiles, protected-state browser transitions, FE-P12 ownership, actual Network Analysis accessibility/visual evidence, synthetic-browser removal, final vendor-boundary audit, and broad retained validation remain `NF-GA-09` |
| Rollback | Revert `5dcfec46` atomically to remove stable wrapper reconciliation and the supported-load fixture. Such a rollback reopens `NF-AC-050`; it must block the current row or supply equivalent real measurement and must not treat the retired no-op selector as sufficient evidence. Semantic paging, extension identities, and read-only boundaries remain mandatory |
| Next workstream | `NF-GA-09`; it may activate only after this checkpoint commit |

#### `NF-GA-09` — Evidence reconciliation, final validation, and handoff

| Field | Approved value |
| --- | --- |
| Status | `PENDING` |
| Dependencies | `NF-GA-08 COMPLETE` |
| Remediation | Replace synthetic authoritative browser rows with real app/grid evidence; classify semantic and live-grid tests; update Phase 12 maps, accounting, visual registry, ledgers/schedules, retained evidence, and this tracker |
| Affected areas | Tests, harness owner inputs, generated accounting, visual artifacts, final documentation |
| Rationale | Make evidence accurately describe production adapter behavior |
| Long-term benefit | Later claims can distinguish product, design, measurement, and publication evidence |
| Compatibility/migration | Test/accounting artifacts only; no runtime or data migration |
| Unresolved risk | Synthetic surfaces may be mistaken for live production-grid conformance |
| Validation criteria | Every gap is closed or explicitly deferred/rejected; no unowned authoritative test; maps and generated accounting agree |
| Commands/evidence | `make phase-ledger-drift`; `make phase-schedule-drift`; `make harness-contract`; `make phase-slice PHASE=phase12`; `make service-backed-slice PHASE=phase12`; `make generated-artifact-policy-check`; `make json-shape-check`; `make agent-finalize`; broaden to `make check` only when final risk/ownership warrants it |
| Rollback | Revert authored evidence maps and generated outputs together; runtime rollback remains bounded by earlier workstreams |
| Exit criteria | Final commits, successful run roots, evidence classes, deferrals, residual risk, and rollback posture are recorded; only then may adoption be marked complete |
| Checkpoint | Status; commits; files; commands/results; result roots; evidence rows; residual risk; migration; rollback; reviewer/date |
| Next workstream | None |

### 14.8 Evidence-class matrix

| Evidence class | Required coverage | Public Make target | Ownership and claim boundary |
| --- | --- | --- | --- |
| Semantic unit | Metadata compilation, identity serialization, query translation, cursor stack, layout reset, state precedence, capability rejection | `make frontend-unit` | Implementation support or mapped product evidence; no live-grid claim alone |
| Live production grid | Actual RDG callback wiring, layout events, range/copy, grouping, virtualization | `make frontend-unit` with production-grid suites | Must remain distinct from pure semantic tests |
| Webserver-backed | Real route query/filter/sort/paging, graph selectors, link/rename/delete workflows | `make browser-e2e-webserver-backed` | Phase 12 product-conformance evidence only when mapped |
| Stateful | Refresh, invalidation, late responses, role/closure/soft-delete transitions, focus continuity | `make browser-e2e-stateful` | Service-backed/stateful evidence |
| Accessibility | Keyboard, focus, ARIA, announcements, permission/read-only cues | `make browser-e2e-a11y` | Accessibility evidence; design direction alone is not Base conformance |
| Visual | Grid geometry, inspector, graph adjacency, and every named data state | `make browser-e2e-visual` | Design-direction evidence maintained through the visual-golden guide |
| Measurement | 1,000-row/all-column bounded DOM and identity stability | `make browser-e2e-measurement` | Informative unless Core 05 separately authorizes a claim |
| Claim-bearing | Adopted Phase 12 requirements and acceptance criteria with retained artifacts | `make phase-slice PHASE=phase12`; `make service-backed-slice PHASE=phase12` | Timed, benchmark, fixture-sensitive, or publication claims require Core 05 |
| Retained run | Final successful roots and summaries | `make agent-finalize RESULTS_DIR=<run-root>` | Record the exact root; absence must be reported, not inferred |

### 14.9 Migration and compatibility posture

| Surface | Expected migration |
| --- | --- |
| Database | None |
| Saved-view data | None; Network Flow remains outside Core saved views |
| Public HTTP protocol | Add one Core-owned mapping-preview route and generic wrapper; existing routes, request/response schemas, defaults, and contract major remain unchanged |
| Repo-local contract | Add closed presentation and mapping-registry metadata plus generated frontend/backend types under `NF-GA-00` |
| Dependency or lockfile | None |
| React Data Grid | No package version or vendor migration |
| Adapter TypeScript API | Required atomic source migration across repository consumers; no compatibility shim |
| Session layout | In-memory only; cross-table for the same grid schema; reset on reload; no data migration |
| Visual goldens | Expected only after implementation, maintained through the existing guide |

No unresolved behavior choice blocks this plan. Implementation still requires
the Network Flow owner to approve NLSpec `1.2.0`, the adapter owner to approve
the generalized semantic API, and the Phase 12/Testing Harness owners to
approve evidence-accounting changes. These are workstream exit approvals, not
permission to change the current tracker-only scope.

### 14.10 Explicit deferred and rejected register

| Capability | Posture | Reopening authority |
| --- | --- | --- |
| Cross-session layout or saved analysis configuration | Useful future enhancement | Network Flow extension-workspace and persistence owner amendment |
| Richer graph canvas | Useful future enhancement | Network Flow and Graph Projection presentation owners |
| Additional flow profiles/column groups | Useful future enhancement | Network Flow profile/specification revision |
| Bulk selection | Intentionally deferred | Adopted Network Flow bulk command and authorization contract |
| Primary flow-table grouping | Intentionally deferred | Adopted analysis workflow with owner grouping semantics |
| Summaries/footer aggregation | Intentionally deferred | Server-owned aggregation/result contract |
| Data-column freezing, dynamic row heights, RTL | Intentionally deferred | Design and adapter owner adoption |
| Existing-row editing, draft/create, paste, fill | Inappropriate | Requires a new owner lifecycle/mutation model; unavailable in the current profile |
| RDG-local filtering/sorting or client aggregation | Inappropriate | Requires a normative authority change; unavailable as an adapter convenience |
| Row reorder, spanning, nested tree, master/detail | Inappropriate | Requires a separately adopted module workflow |
| Vendor indexes/handles/classes/DOM or Network Flow cell presence | Inappropriate | Prohibited by the semantic boundary and Core presence model |

### 14.11 Historical planning-update validation

The initial planning-only update changed only this tracker. At that historical
checkpoint, phase ledgers, schedules, runtime suites, code generation, and broad
checks were intentionally not run because no corresponding owner input changed.
This table records that planning checkpoint only; the live implementation gates
and retained roots are owned by the per-workstream completion checkpoints.

| Command | Result | Result root or note |
| --- | --- | --- |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260716T024904Z-p49252` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260716T024904Z-p49253` |
| `make lint-markdown` | PASS | No retained result root emitted |
| `git diff --check` | PASS | No output |

Baseline inspection before this edit passed the same checks:

| Command | Baseline result | Baseline result root or note |
| --- | --- | --- |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260716T023142Z-p43466` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260716T023142Z-p43467` |
| `make lint-markdown` | PASS | No retained result root emitted |
| `git diff --check` | PASS | No output |
