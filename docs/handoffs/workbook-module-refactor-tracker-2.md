# Workbook Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Target directory: `internal/modules/workbook`.

Output path: `docs/handoffs/workbook-module-refactor-tracker-2.md`.

Allowed changes for this session:

- Create this handoff tracker.
- Record current repository evidence and planning findings.
- Identify behavior-preserving future refactor slices.
- Mark unknowns as `TODO:` rather than guessing.
- Mark owner-document contradictions as `BLOCKED: owner contradiction`.

Non-goals:

- Do not prove that `internal/modules/workbook` is a valid permanent module boundary merely because the directory exists.
- Do not move code, rename packages, edit generated roots, update lockfiles, or run production refactors.
- Do not treat phase maps, phase ledgers, or test rows as runtime architecture. They are verification and evidence accounting only.

Authority order for this tracker:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans and framework files as evidence, not authority.

Planning posture:

- The local framework `docs/handoffs/cartulary_modular_refactor_planning_framework.md` is used as planning doctrine and template shape, not as proof of current repository state.
- `docs/domain.md` is used for domain vocabulary and concept boundaries.
- Prior tracker content is evidence only. It may describe earlier or broader workbook ownership that no longer exists.
- Current branch and commit from inspection: branch `main`, commit `52d473c`.
- `git status --short` output had no dirty entries before this tracker was created.
- Repository/framework mismatch to preserve as a finding: current `internal/modules/workbook` has six production files, not the older or broader workbook-owned shape described by some prior planning material.
- Current WebSocket contract found in the repository is `GET /ws/v1/incidents/{incident_id}` with replayable `record_changed` messages. A per-view `/views/{view_schema_id}/changes` WebSocket path was not found in current contract evidence.
- No owner-document contradiction was discovered in this planning pass.

Execution update:

- A later task authorized implementation of the remediation plan. This tracker remains the handoff artifact, but now records implemented contract/spec/import-dispatcher remediation in addition to the original planning snapshot.
- `internal/modules/workbook` is still treated as a compatibility route/application facade, not as a permanent domain owner. The implementation work intentionally did not move linked-note, decision-supersede, or generic workbook mutation orchestration out of `mutation_store.go`; those remain future owner-facade work.

## 2. Current-State Repository Inventory

Generated and contract surfaces that workbook changes may affect:

- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/ws/index.schema.json`
- `contracts/view-schemas/*.json`
- `packages/protocol-ts/src/generated/contracts.ts`
- `packages/view-contracts/src/index.ts`
- `packages/ui-contracts/src/index.ts`

Contract gap disposition:

- DONE: `POST /api/v1/records/{record_id}/linked-notes` is now present in `contracts/openapi/cartulary.openapi.yaml` with operationId `createRecordLinkedNote`, request schema `LinkedNoteCreateRequest`, and `ViewMutationEnvelope` success responses. `make generate` refreshed `internal/gen/contracts/contracts_gen.go` and `packages/protocol-ts/src/generated/contracts.ts`, and `internal/modules/workbook/openapi_contract_test.go` pins the generated OpenAPI posture.
- DONE: `PATCH /api/v1/records/{record_id}` is now documented as generic record patch with operationId `patchRecord` and no timeline-only tag.
- TODO: linked-note implementation ownership is still workbook-coordinated in `mutation_store.go`; move it behind artifacts plus links owner facades in a later behavior-preserving slice.

| path | current responsibility | exported/public symbols or package surface | inbound callers | outbound dependencies | tests touching it | generated artifacts or contracts touched | suspected target owner module | risk level | notes |
| ---- | ---------------------- | ------------------------------------------ | --------------- | --------------------- | ---------------- | ---------------------------------------- | ----------------------------- | ---------- | ----- |
| `internal/modules/workbook/routes.go` | HTTP route facade, auth/CSRF/role checks, route dispatch, telemetry, response/error mapping, record-change publication | `RegisterRoutes`, `Service`, route handlers for query, rows create, clipboard paste, bulk mutations, record patch, linked notes, supersede, conflict resolve | `internal/app/runtime.go`; frontend/API clients through HTTP routes; tests in workbook and browser suites | `auth`, `authn`, `httpapi`, `pagination`, `viewquery`, `viewschema`, `timeline`, `entities`, `incidents`, `revisions`, `collaboration` | Workbook integration/mutation integration; phase6 conflict route/support/integration; phase9 route guard, clipboard, notes/indicators, parties, task decisions; frontend/browser references | OpenAPI route surface; generated TS protocol clients; WS `record_changed` publication semantics through collaboration | keep as transport/application facade; domain work should move to timeline/entities/artifacts/evidence/links/tasksdecisions/revisions/projections/collaboration | high | Correct long-term role appears to be route compatibility plus transport policy, not source-domain ownership. |
| `internal/modules/workbook/store.go` | Workbook compatibility store and view-schema constants; dispatches query rows to timeline, entities, or projections | `Store`, `NewStore`, `QueryRows`, view schema constants | `routes.go`; peer tests in owner modules | `projections`, `timeline`, `entities`, `artifacts`, `evidence`, `links`, `revisions`, `tasksdecisions` | `optional_surfaces_store_test.go`; `store_artifact_surface_test.go`; workbook integration tests; phase8/phase9 query tests; peer tests | View-schema contracts and query response contracts | temporary compatibility query facade; future owner query facades or projections | high | Imports no longer use this store after the import dispatcher remediation; query dispatch still delegates and remains a compatibility surface. |
| `internal/modules/workbook/mutation_api.go` | Route-facing DTOs, decoders, validation errors, request hashes, collection action parsing, mutation payloads | `CreateRequest`, `PatchRequest`, `LinkedNoteCreateRequest`, `ConflictResolveRequest`, `MutationResult`, decoders, hash builders, validation/conflict errors, mutation payload builders | `routes.go`; workbook tests; peer tests | standard JSON/time/UUID helpers; route-level validation; owner-specific field semantics encoded in DTO parsing | `mutation_api_test.go`; phase6 conflict tests; phase9 clipboard/notes/party/task tests; workbook mutation integration | OpenAPI schemas; generated protocol contracts; idempotency hash behavior | compatibility API adapter; future owner-specific adapters after characterization | high | Imports no longer consume workbook DTOs; wire hashes and request normalization remain observable behavior for workbook routes. |
| `internal/modules/workbook/mutation_store.go` | Cross-owner mutation transaction coordinator for generic rows, linked notes, patch/conflict resolution, decision supersede, idempotency, revisions, projection refresh | `CreateWorkbookRow`, `CreateLinkedNote`, `PatchWorkbookRow`, `ResolveWorkbookConflict`, `SupersedeDecision` | `routes.go`; workbook tests | `artifacts`, `evidence`, `entities`, `links`, `projections`, `revisions`, `tasksdecisions`, timeline facade, postgres transactions | workbook mutation integration; phase6 conflict; phase9 notes/indicators, parties, task decisions, timestamp tests | Revision/change-set contracts; projection refresh contracts; generated route responses | split toward artifacts/evidence/entities/links/tasksdecisions/revisions/projections; keep wrapper until owner facades exist | critical | This remains the largest boundary risk: it owns cross-owner side effects even when source behavior belongs elsewhere. |
| `internal/modules/workbook/paging.go` | Cursor/page slicing and sort comparison for query responses | package-private page and comparison helpers | `store.go` query paths and tests | `pagination`, `viewquery` row sorting semantics | phase8 row-wire/grouping; workbook integration; paging indirectly through query tests | Query response envelope and paging meta contracts | keep as compatibility adapter or move to shared query/projections helper | medium | Observable paging meta and sort behavior must freeze before moving. |
| `internal/modules/workbook/telemetry.go` | Workbook query/mutation spans, metrics, safe attribute mapping | package-private telemetry helpers and metric names | `routes.go` handlers | OpenTelemetry metrics/traces, safe attribute mapping | `telemetry_test.go`; integration tests indirectly | Harness/accounting and telemetry naming only; not generated API | keep in route/application facade while route names remain workbook-compatible | medium | Metric/span names are observable for operators and harness evidence. |
| `internal/modules/workbook/mutation_api_test.go` | Characterizes mutation DTO decoding, validation, hashes, collection actions, and payload construction | Test-only coverage for `mutation_api.go` package surface | Go test runner | workbook DTO helpers | Directly touches mutation API behavior | Confirms OpenAPI/protocol-compatible wire normalization | workbook compatibility tests, later owner adapter tests | high | Preserve when splitting DTO ownership. |
| `internal/modules/workbook/optional_surfaces_store_test.go` | Verifies optional workbook query surfaces and store behavior | Test-only | Go test runner | workbook store and optional owner stores | Directly touches `store.go` | View-schema/query surface behavior | projections/owner query tests | medium | Useful characterization for query facade cleanup. |
| `internal/modules/workbook/phase6_conflict_route_test.go` | Route-level same-field conflict and resolution behavior | Test-only | Go test runner | routes, revisions, auth/test harness | Direct route conflict tests | OpenAPI conflict route and response shape | revisions plus route facade | high | Required before conflict route ownership changes. |
| `internal/modules/workbook/phase6_conflict_support_test.go` | Shared support for phase6 conflict tests | Test-only helpers | Go test runner | workbook conflict test harness | Phase6 tests | Harness evidence only | revisions/workbook test support | medium | Test helper, not runtime owner evidence. |
| `internal/modules/workbook/phase6_integration_test.go` | Phase6 workbook integration behavior around conflicts and patching | Test-only | Go test runner | workbook routes/store, revisions | Phase6 integration | Conflict and patch contracts | revisions plus route facade | high | Preserve for conflict/payload changes. |
| `internal/modules/workbook/phase6_shared_harness_test.go` | Shared harness setup for phase6 workbook tests | Test-only helpers | Go test runner | backend testutil and workbook setup | Phase6 tests | Harness evidence only | test harness | medium | Keep as evidence accounting. |
| `internal/modules/workbook/phase8_grouping_contract_test.go` | Query grouping contract behavior | Test-only | Go test runner | workbook query/store | Phase8 tests | View/query contract | projections/view contracts | medium | Supports projection/query move planning. |
| `internal/modules/workbook/phase8_row_wire_test.go` | Row wire shape and field projection behavior | Test-only | Go test runner | workbook rows/query | Phase8 tests | `view_row_v1` and generated view contracts | view contracts/projections | high | Freeze row shape before refactor. |
| `internal/modules/workbook/phase9_clipboard_paste_integration_test.go` | Clipboard paste integration through workbook route | Test-only | Go test runner | routes, timeline/entities clipboard paste | Phase9 clipboard tests | Clipboard paste route contract | timeline/entities plus route facade | high | Route remains workbook-shaped even if owner logic moves. |
| `internal/modules/workbook/phase9_clipboard_paste_unit_test.go` | Clipboard paste decoding/unit behavior | Test-only | Go test runner | clipboard paste API helpers | Phase9 clipboard unit tests | Request normalization/hashes | timeline/entities adapters | medium | Preserve request shape and error behavior. |
| `internal/modules/workbook/phase9_coordination_surfaces_test.go` | Coordination workbook surfaces and route/store interactions | Test-only | Go test runner | workbook store and owner stores | Phase9 coordination tests | View/query contract evidence | projections plus owner facades | medium | Evidence for split sequencing. |
| `internal/modules/workbook/phase9_notes_indicators_test.go` | Notes and indicator workbook behavior | Test-only | Go test runner | artifacts/entities/workbook mutation/query paths | Phase9 notes/indicator tests | Row wire and mutation contracts | artifacts/entities/projections | high | Directly informs linked-note and artifact/indicator ownership. |
| `internal/modules/workbook/phase9_parties_integration_test.go` | Parties integration through workbook behavior | Test-only | Go test runner | entities party stores and workbook routes | Phase9 party integration | Party row contracts | entities/projections | high | Preserve party row/query/mutation behavior. |
| `internal/modules/workbook/phase9_party_refs_test.go` | Party reference behavior and field refs | Test-only | Go test runner | entities, links, workbook mutation paths | Phase9 party refs | Links/reference contracts | entities plus links | high | Boundary evidence for links ownership. |
| `internal/modules/workbook/phase9_route_guard_characterization_test.go` | Route guard characterization for auth/session/CSRF/roles | Test-only | Go test runner | routes and auth test harness | Phase9 route guard tests | Authorization and session behavior | route facade/auth | critical | Must remain green for any route facade change. |
| `internal/modules/workbook/phase9_sprint0_sentinel_test.go` | Sentinel coverage for phase9 workbook scope | Test-only | Go test runner | workbook package | Phase9 sentinel | Harness accounting | test harness | low | Evidence accounting only. |
| `internal/modules/workbook/phase9_task_decisions_store_test.go` | Task and decision mutation/store behavior | Test-only | Go test runner | tasksdecisions, links, revisions, workbook store | Phase9 task/decision tests | Task/decision row and mutation contracts | tasksdecisions plus links/revisions | high | Required before moving decision supersede orchestration. |
| `internal/modules/workbook/phase9_timestamp_contract_test.go` | Timestamp normalization and contract behavior | Test-only | Go test runner | mutation API/store | Phase9 timestamp tests | Wire/time contract behavior | route adapter and owner stores | medium | Preserve before DTO split. |
| `internal/modules/workbook/store_artifact_surface_test.go` | Artifact-backed surface behavior in workbook store | Test-only | Go test runner | artifacts/projections/workbook store | Artifact surface tests | View-schema/query contract evidence | artifacts/projections | high | Evidence that artifact logic should not remain workbook-owned. |
| `internal/modules/workbook/telemetry_test.go` | Telemetry labels, spans, metrics, and safe attributes | Test-only | Go test runner | telemetry helpers | Direct telemetry tests | Telemetry/harness accounting | route/application facade | medium | Freeze metric/span names. |
| `internal/modules/workbook/workbook_integration_test.go` | End-to-end workbook query/route integration behavior | Test-only | Go test runner | routes, store, auth, owner modules | Integration tests | OpenAPI route and row envelope behavior | compatibility facade plus owner facades | critical | Primary behavior-preservation evidence. |
| `internal/modules/workbook/workbook_mutation_integration_test.go` | End-to-end workbook mutation, revision, projection, and response behavior | Test-only | Go test runner | mutation store, routes, revisions, projections | Mutation integration tests | Mutation route/response/revision contracts | compatibility facade plus owner mutation facades | critical | Primary mutation refactor guard. |

## 3. Workbook Boundary Diagnosis

Current diagnosis: `internal/modules/workbook` is a mixture. It is currently a mixed compatibility/application facade, not a proven permanent domain boundary.

Evidence supports these simultaneous roles:

- Legitimate thin HTTP/application facade for stable route shapes, auth/CSRF/role checks, route envelopes, error mapping, session sliding, and telemetry labels.
- View/projection orchestration adapter for query dispatch and row paging over timeline, entities, and projections.
- Mutation coordinator for workbook-shaped create/patch/conflict/supersede/linked-note operations.
- Transport-adjacent adapter for HTTP method/path handling and route-facing DTOs.
- Misplaced or transitional home for source-domain orchestration that should move behind clearer owners when characterization exists.

Many source-domain behaviors are already behind owner modules: timeline, entities/parties, artifacts, evidence, links, tasks/decisions, projections, revisions, collaboration, and imports/tabular ingest. The remaining workbook role should be treated as compatibility until owner-specific facades and tests prove a narrower boundary.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| -------------------- | ---------------- | ----------------------- | --------------------------- | -------- | ----- |
| HTTP route registration and stable route envelopes | `routes.go` | workbook route/application facade | keep/defer | `RegisterRoutes` owns current HTTP paths and response writing | Keep while preserving route compatibility. |
| Session sliding, authn, CSRF, membership, role guards | `routes.go` | route facade plus auth/authn platform | keep/defer | Route guard tests and handler code | Route layer can retain transport/application policy. |
| Query dispatch for timeline/entity/projection rows | `store.go` | projections and owner query facades | split/defer | `QueryRows` delegates by `view_schema_id` | Keep compatibility wrapper until owner query facades cover all surfaces. |
| Query paging and sort slicing | `paging.go` | query/projections helper or workbook compatibility adapter | keep/defer | Paging helpers are package-private and route-observable | Move only after row/meta characterization. |
| Timeline create/patch/paste/bulk/supersede/conflict | `routes.go`, `mutation_store.go`, timeline facade calls | `timeline` facade | keep route dispatch only; move owner behavior | Timeline facade already handles several operations | Workbook should not own timeline semantics. |
| Hosts, identities, indicators, parties | `store.go`, `mutation_store.go` | `entities` | split/move | Entity query and mutation paths use entities stores | Preserve workbook routes while moving entity decisions behind entities adapters. |
| Notes, communication log, handoff, status review, lesson, findings, investigative queries, forensic keywords | `mutation_store.go`, `store.go` | `artifacts` plus `links` and `projections` | split/move | Artifact stores and projection refresh already exist | Linked-note creation is the highest-risk remaining artifact/link candidate. |
| Evidence rows and lifecycle | `mutation_store.go`, `store.go` | `evidence` | split/move | Evidence workbook mutation APIs exist in evidence module | Keep route compatibility while moving lifecycle validation behind evidence. |
| Links, tags, risk refs, linked-note references | `mutation_store.go` | `links` | split/move | Link module owns field refs and linked-note refs | Workbook should coordinate less link state directly. |
| Patch conflict window, payloads, tokens, revisions | `mutation_store.go`, `routes.go` | `revisions` | split/move | Revisions module owns conflict builders and codecs | Route may continue mapping conflict responses. |
| Query materialization and projection refresh | `mutation_store.go`, `store.go` | `projections` | split/move | Projection module owns support checks and refresh methods | Refresh side effects should be explicit owner contracts. |
| Record-change publication | `routes.go` | `collaboration` | keep dispatch; preserve event semantics | Current WS contract is incident-scoped `record_changed` | Workbook route can publish through collaboration, not own WS contract. |
| Import apply for non-timeline rows | `internal/modules/imports/routes.go` target registry and owner dispatch | `imports` plus owner create facades | keep import dispatcher; expand owner facades | Imports production code no longer imports workbook; timeline and entities targets dispatch through owner create paths; unsupported/no-facade targets fail closed | Evidence: `internal/modules/imports/targets.go`, `boundary_test.go`, and Phase 11 import integration tests. |

## 4. Public Contract and Behavior Freeze Map

Every listed surface is behavior-preserving by default. Route shape, envelopes, WebSocket paths, event semantics, row wire shape, storage side effects, authorization outcomes, generated contracts, and harness accounting must remain stable unless a later task explicitly authorizes a behavior change.

| Contract surface | Existing behavior to freeze | Existing tests or evidence | Required characterization tests or gaps |
| ---------------- | --------------------------- | -------------------------- | -------------------------------------- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | Route shape, request envelope, query filters/sort/grouping, hidden-field behavior, paging meta, status codes, auth/CSRF/session outcomes | OpenAPI route; `workbook_integration_test.go`; phase8 grouping/row-wire; phase9 coordination; route guard tests | Add owner-specific query facade characterization before moving query dispatch. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | Route shape, create DTO normalization, idempotency hash, mutation result envelope, projection refresh, revision writes, `record_changed` publication | OpenAPI route; mutation API tests; workbook mutation integration; phase9 notes/parties/task tests | Preserve hash fixtures before splitting DTO ownership. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste` | Clipboard paste request shape, target-specific dispatch, response envelope, auth and CSRF behavior | OpenAPI route; phase9 clipboard paste integration/unit tests | Confirm timeline/entity owner expectations before moving adapters. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations` | Bulk mutation route shape, route key, normalization, error mapping, timeline dispatch | OpenAPI route; timeline bulk mutation tests; workbook route tests | TODO: confirm non-timeline rejection/handling remains covered. |
| `PATCH /api/v1/records/{record_id}` | Patch DTO normalization, same-field conflict shape, revision/change-set writes, projection refresh, row response, status codes | OpenAPI operationId `patchRecord`; `openapi_contract_test.go`; phase6 conflict tests; workbook mutation integration; revisions conflict token/workbook tests | DONE for generated/public contract naming; TODO: owner-specific patch adapters remain future work. |
| `POST /api/v1/records/{record_id}/linked-notes` | Route shape, request/response envelope, linked artifact creation, link reference write, projection refresh, revision/change-set write, `record_changed` publication | OpenAPI operationId `createRecordLinkedNote`; `LinkedNoteCreateRequest`; `openapi_contract_test.go`; frontend/tests references; phase9 notes/indicator tests | DONE for generated/public contract mapping; TODO: move implementation orchestration behind artifacts plus links owner facade. |
| `POST /api/v1/records/{record_id}/supersede` | Supersede route shape, decision lifecycle, link/reference side effects, revision/change-set writes, status/error mapping | OpenAPI route; phase9 task decisions tests | Add owner-facade characterization before moving to tasksdecisions. |
| `POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve` | Conflict token handling, request hash, conflict payload shape, merge semantics, status/error mapping | OpenAPI route; phase6 conflict route/support/integration/shared harness; revisions conflict tests | Preserve same-field conflict fixtures before changing route/store split. |
| `GET /ws/v1/incidents/{incident_id}` | Incident-scoped WebSocket path, replayable `record_changed` event, payload fields, event ordering expectations | `contracts/ws/index.schema.json`; collaboration publisher tests; route publication in workbook | Do not replace with per-view path without new authority. |
| Workbook row/query/mutation behavior | `view_row_v1` wire shape, row IDs/types, field keys, hidden fields, timestamps, query sort/group behavior | Phase8 row-wire/grouping; phase9 timestamp; workbook integration | Characterize each owner surface before moving source logic. |
| Saved-view or view-schema behavior | Stable `view_schema_id` constants and schema-driven field behavior | View-schema JSON contracts; UI contract packages; store constants | TODO: confirm any saved-view-specific callers before changing constants. |
| Projection refresh behavior | Mutations refresh affected derived rows before responses/events where current behavior does so | Mutation integration; projection store tests; owner module tests | Add explicit refresh expectations for linked notes and supersede slices. |
| Authorization checks | Session, CSRF, membership, and role outcomes on every workbook route | Phase9 route guard characterization; auth test harness | Keep in route facade unless platform-auth owner change is authorized. |
| Revision/change-set behavior | Idempotency, row version, conflict token, change-set writes, actor/client transaction fields | Phase6 conflict tests; revisions tests; workbook mutation integration | Preserve before moving mutation coordination into owners. |
| Generated protocol/view contracts | OpenAPI, WS schema, view schema, generated TS packages | Contract files and generated packages listed above | Run drift checks after any implementation. Do not hand-edit generated roots. |
| Harness/test accounting | Phase6/phase8/phase9 rows remain evidence accounting, not architecture | `make task-guide`, phase tests, harness docs | Use phase slices for validation; do not infer runtime ownership from phase maps. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| ------- | -------- | ---- | -------------- | -------------- | ------------------------ |
| Linked-notes route previously had missing OpenAPI mapping | `contracts/openapi/cartulary.openapi.yaml` now declares `POST /api/v1/records/{record_id}/linked-notes`; generated OpenAPI is pinned by `openapi_contract_test.go` | Remaining risk is implementation ownership, not public contract absence | should_fix | artifacts plus links plus route facade | Move linked-note orchestration behind artifacts plus links facade in a later slice. |
| `mutation_store.go` cross-coordinates many owner side effects | It creates/patches artifact, evidence, entity, task/decision, link, revision, projection, and idempotency side effects | Boundary changes can silently alter storage, revision, or event behavior | must_fix | owner facades plus temporary workbook compatibility wrapper | Split only after mutation behavior fixtures cover each owner path. |
| Imports production code previously depended on workbook DTO/store for arbitrary non-timeline import apply | `internal/modules/imports/routes.go` now depends on timeline/entities owner create paths and `boundary_test.go` rejects production workbook imports | Remaining targets without owner facades fail closed until owners provide create contracts | defer | imports plus owner dispatchers | Add missing owner create facades and apply-journal support before expanding importability. |
| Route handlers know many owner error mappings | `routes.go` maps timeline, entity, revision, workbook, and auth errors | Error/status drift during owner moves | should_fix | route facade with owner-specific error adapters | Inventory status codes and add focused route characterization. |
| Workbook owns route-facing DTOs for multiple source domains | `mutation_api.go` parses fields for artifact/evidence/entity/task/decision surfaces | DTO split can break idempotency hashes and normalization | should_fix | owner-specific API adapters behind stable route wire | Freeze hash fixtures before adapter extraction. |
| Query compatibility store dispatches across timeline, entities, and projections | `store.go` chooses owner by `view_schema_id` | Accidental workbook ownership may persist | defer | projections and owner query facades | Keep until all owner query facades cover current routes. |
| Paging adapter is package-private but route-observable | `paging.go` controls cursor/page slicing and sort comparison | Sort/cursor drift affects clients | defer | query/projections helper or workbook compatibility | Move only after query contract tests are owner-independent. |
| Auth/CSRF/role checks live in route facade | `routes.go` applies transport/application policy before owner dispatch | Moving into owner modules would mix transport policy with domain logic | intentional/no_action | route facade plus auth platform | Keep in route layer unless platform-auth architecture changes. |
| Telemetry names are workbook-labeled | `telemetry.go` emits workbook query/mutation spans and metrics | Renaming can break dashboards and harness evidence | defer | route/application facade | Preserve names until observability migration is authorized. |
| Generated files are downstream only | Generated roots are declared by repository procedure | Hand-editing would violate repo policy | intentional/no_action | contract owners plus generators | Update owner inputs and run drift targets only in later implementation. |
| Phase maps and test rows are evidence accounting only | Harness docs and repo procedure define mechanics | Treating them as runtime architecture would mislead boundary design | intentional/no_action | test harness | Use them for validation selection, not module ownership. |
| Workbook route facade can remain if it hides only transport/application policy | Current routes preserve public HTTP behavior | Removing facade too early risks client breakage | intentional/no_action | route/application facade | Narrow facade after owner facades are complete. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| ----------- | ---- | -------------------------- | --------------------------- | ----------------------------- | ---- | -------------------- | ---------- | ------------------ |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01, WF-02 | Establish authority, snapshot, and planning-only output | `docs/handoffs/workbook-module-refactor-tracker-2.md`; framework/docs/spec/domain inputs | Docs/static validation | Tracker created with source posture and repo snapshot. |
| WF-01 | Workbook package inventory | chain | WF-00 | WF-02, WF-04 | Inventory every current file under `internal/modules/workbook` | All production and test files under `internal/modules/workbook` | `find`, `rg`, future `make backend-unit` as needed | Every file inventoried or explicitly out of scope. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Map routes, generated contracts, WS events, view schemas, and owner candidates | `routes.go`, contracts, frontend contract packages, owner modules | `make generate-drift`, `make json-shape-check` after implementation | Each public contract has owner/test posture. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-07 | Identify tests required before moving behavior | Workbook tests, owner tests, browser/frontend tests | Phase slices and targeted backend commands | Linked-notes/import/supersede/conflict/projection gaps recorded. |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05 | Classify coupling as must_fix, should_fix, defer, or intentional | Workbook production files, imports module, owner modules | Static/import-boundary checks and backend tests after code changes | Findings table complete with proposed owners. |
| WF-05 | Facade or ownership redesign plan | chain | WF-03, WF-04 | WF-06 | Define narrowed workbook facade and owner-facade responsibilities | Workbook package plus timeline/entities/artifacts/evidence/links/tasksdecisions/revisions/projections/collaboration/imports | Design review plus characterization tests | Owners and compatibility wrappers identified. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-08 | Sequence behavior-preserving implementation slices | Future implementation files from WF-05 | Slice-specific validation commands | Each slice has dependency, rollback, and completion criteria. |
| WF-07 | Harness/test/accounting update plan | chain | WF-03 | WF-08 | Select Make targets and phase slices for each refactor area | Makefile task surface, test harness docs, phase tests | `make task-guide`, `make phase-slice`, `make service-backed-slice` | Validation table and command posture complete. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Leave enough evidence for next agent without rediscovery | This tracker and validation outputs | Docs/static validation for tracker; broader checks for later code | Handoff log current and blockers explicit. |

## 7. Proposed Refactor Slice Plan

This tracker was created before implementation. A later task authorized remediation work; completed slices are recorded in Section 9 and Section 10. Remaining future slices must preserve observable behavior unless a later task explicitly authorizes a behavior change.

| Slice ID | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| -------- | ---------- | --------------------- | -------------------------------- | -------------- | ------------------------ | ------------------ | ------------- | -------------------- |
| S-00 | none | Create this tracker only | `docs/handoffs/workbook-module-refactor-tracker-2.md` | Documentation drift only | Markdown/static checks | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | Revert tracker document only | Tracker exists with inventory, findings, workflows, slices, validation, handoff, and blockers. |
| S-01 | S-00 | Add or confirm characterization for linked-notes OpenAPI/route behavior, import apply compatibility, decision supersede, same-field conflict, projection refresh, and `record_changed` payloads | Workbook tests; contracts; owner tests; frontend/browser references | Missing coverage can hide behavior changes in later slices | Add linked-notes contract test; preserve phase6, phase8, phase9, collaboration, revisions, timeline guard tests | `make backend-unit`; `make backend-integration`; phase6/phase8/phase9 slices as relevant | Keep current workbook implementation until tests are green | Gaps are covered or explicitly documented with TODO owner evidence. |
| S-02 | S-01 | Move linked-note orchestration behind artifacts plus links facade while preserving current route and wire response | `internal/modules/workbook/mutation_store.go`; `internal/modules/artifacts`; `internal/modules/links`; revisions/projections helpers | Linked artifact ID, reference writes, projection refresh, revisions, status codes, `record_changed` | Linked-notes route characterization; phase9 notes/indicator; workbook mutation integration | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase9`; `make service-backed-slice PHASE=phase9` | Retain workbook wrapper that calls old path until facade proves equivalent | Workbook no longer directly owns linked-note side effects except compatibility delegation. |
| S-03 | S-01 | Move decision supersede orchestration behind tasksdecisions plus links plus revisions facade while preserving route | `mutation_store.go`; `internal/modules/tasksdecisions`; `internal/modules/links`; `internal/modules/revisions` | Decision lifecycle, supersede status, link refs, revisions, projection refresh, status codes | Phase9 task decisions; workbook mutation integration; revisions tests | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase9` | Keep workbook wrapper delegating to prior implementation | Supersede route delegates to owner facade with unchanged behavior. |
| S-04 | S-01 | Reduce workbook mutation DTO ownership by adding owner-specific adapters without changing wire hashes | `mutation_api.go`; owner adapter packages; route tests | Hash drift, normalization drift, validation errors, generated schema mismatch | Mutation API hash fixtures; phase6 conflict; phase9 timestamp; OpenAPI drift checks | `make backend-unit`; `make generate-drift`; `make json-shape-check` | Keep exported workbook DTO wrappers that call new adapters | Route-facing hashes and validation remain byte-for-byte compatible where expected. |
| S-05 | S-01, S-04 | Replace imports-to-workbook production dependency with import-owned dispatcher for non-timeline targets | `internal/modules/imports`; owner create facades | Import request behavior, idempotency, target routing, validation errors | Import route/store tests plus owner create tests | `make backend-unit`; `make phase-slice PHASE=phase11`; targeted import tests if available | Retain unsupported targets as fail-closed until owner facade exists | DONE for removing imports production dependency on workbook and adding timeline/entities dispatch; TODO for remaining owner create facades and explicit apply journal. |
| S-06 | S-02, S-03, S-04, S-05 | Cleanup workbook to route/query/paging/telemetry compatibility facade only | `routes.go`; `store.go`; `mutation_store.go`; `mutation_api.go`; owner modules | Public routes, row wire, auth, revisions, projections, WS event semantics | Full workbook integration/mutation integration; phase6/phase8/phase9; frontend conditional checks | `make agent-finalize`; `make test-fast`; `make check` when risk warrants | Keep compatibility wrappers until all owner paths validate | Workbook owns no source-domain logic beyond transport/application compatibility. |

## 8. Validation Plan

Use Make-owned commands only. Commands below were discovered from the current repo task surface and task-guide output. Direct `go`, `pnpm`, Vitest, Playwright, Biome, and raw script commands remain developer conveniences unless invoked by Make.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| ---------------- | ------- | ----- | ------------------------------- | ----- |
| unit | `make backend-unit` | Backend unit tests, including workbook and owner packages | yes | Use before and after owner-facade changes. |
| integration | `make backend-store`; `make backend-integration`; `make phase-slice PHASE=phase9`; `make service-backed-slice PHASE=phase9` | Store/integration and phase9 workbook behavior | yes | Also run `make phase-slice PHASE=phase6` and `make phase-slice PHASE=phase8` when conflict/query behavior changes. |
| e2e/browser | `make browser-e2e` or target selected by `make task-guide ROLE=feature-dev PHASE=phase9` | Browser-visible workbook behavior | no for tracker-only; conditional for route/UI-impacting implementation | Use when frontend route behavior, clipboard, or row wire surfaces change. |
| generated drift | `make generated-artifact-policy-check`; `make generate-drift`; `make json-shape-check` | Generated roots, contract drift, JSON shape | yes for contract-impacting implementation | For tracker-only creation, run artifact policy and JSON shape checks; `generate-drift` is reserved for implementation/contract changes. |
| import-boundary/static | `make frontend-import-boundary-check`; `make frontend-typecheck`; `make frontend-unit` | Frontend import boundaries and contract consumers | no for tracker-only; conditional for generated/client contract changes | Use when OpenAPI/view/WS contracts or frontend-facing row behavior changes. |
| full check | `make agent-finalize`; then `make test-fast` or `make check` | End-of-run hygiene and broad repository verification | no for tracker-only; yes for broad implementation slices | `make check` is broader and should follow risk-based escalation. |
| docs/static for this tracker | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | Markdown plus generated-artifact/JSON policy guards | yes for tracker-only completion | These are the planned validation commands for S-00. |

## 9. Top-Level Work Tracker

Statuses are limited to `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED`.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| WB2-00 | Create planning-only tracker | WF-00 | DONE | none | This file | Tracker exists at `docs/handoffs/workbook-module-refactor-tracker-2.md`. |
| WB2-01 | Inventory all files under `internal/modules/workbook` | WF-01 | DONE | WB2-00 | Section 2 inventory table | Every current production and test file is listed. |
| WB2-02 | Record workbook boundary diagnosis | WF-04 | DONE | WB2-01 | Section 3 diagnosis and decision table | Boundary classified as mixed compatibility/application facade. |
| WB2-03 | Map public contracts and freeze posture | WF-02 | DONE | WB2-01 | Section 4 contract map | Routes, WS event, row/query/mutation, revisions, projections, auth, generated surfaces, and harness accounting are listed. |
| WB2-04 | Track linked-notes contract gap | WF-02 | DONE | WB2-03 | OpenAPI route, generated contracts, `openapi_contract_test.go` | OpenAPI/generated posture confirmed; implementation ownership remains in WB2-06. |
| WB2-05 | Track imports compatibility dependency | WF-04 | DONE | WB2-02 | Core import dispatcher specs; `imports/targets.go`; `boundary_test.go`; Phase 11 tests | Imports production package no longer depends on workbook DTO/store. |
| WB2-06 | Track mutation-store orchestration split | WF-05 | TODO | WB2-02, WB2-03 | Findings and slices S-02/S-03/S-04 | Owner facades designed with characterization coverage. |
| WB2-07 | Define workstreams and slice sequence | WF-06 | DONE | WB2-02, WB2-03 | Sections 6 and 7 | WF-00 through WF-08 and S-00 through S-06 have dependencies and exit criteria. |
| WB2-08 | Discover validation commands | WF-07 | DONE | WB2-00 | Section 8 validation plan; Make task-surface inspection | Commands are listed or marked conditional. |
| WB2-09 | Complete handoff tables | WF-08 | DONE | WB2-07, WB2-08 | Section 10 handoff log | All required handoff categories have rows. |
| WB2-10 | Run tracker-only validation | WF-08 | DONE | WB2-00 | `make lint-markdown` passed; `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260630T225618Z-p10118`; `make json-shape-check` passed with run root `.cartulary/test-results/20260630T225622Z-p10298` | Tracker-only validation results recorded. |
| WB2-11 | Add/confirm characterization tests before code movement | WF-03 | IN_PROGRESS | WB2-03 | `openapi_contract_test.go`; Phase 11 import tests; existing phase9/phase6 tests | Linked-note and generic patch public contract characterization added; owner-facade mutation characterization remains future work. |
| WB2-12 | Move linked-note orchestration behind owner facade | WF-05 | DEFERRED | WB2-11 | Slice S-02 | Later implementation preserves route behavior and passes validation. |
| WB2-13 | Replace imports-to-workbook production dependency | WF-05 | DONE | WB2-11 | `internal/modules/imports/{targets.go,routes.go,store.go,boundary_test.go,imports_integration_test.go}` | Imports production package has no workbook import and dispatches only through registered owner targets or fail-closed errors. |
| WB2-14 | Add remaining owner create facades for non-timeline/non-entity import targets | WF-05 | TODO | WB2-13 | `owner_create_contract_unavailable` fail-closed behavior | Evidence, artifacts, tasks/decisions, parties, assessments, and coordination targets have owner create facades or are explicitly not importable. |
| WB2-15 | Add explicit import apply journal | WF-05 | TODO | WB2-13 | Core 03 import dispatcher algorithm | Import apply records durable per-row dispatch/provenance/retry journal outside workbook. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Planning-only tracker created from current repo evidence and local framework doctrine | Inspected `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, `docs/domain.md`, core specs, prior tracker, workbook package; touched this tracker only | `find`, `rg`, `sed`, `wc`; `git branch --show-current`; `git rev-parse --short HEAD`; `git status --short` | Authority order and clean snapshot recorded | None for tracker creation | Run tracker-only validation and record results. |
| 2026-06-30 | Codex remediation session | Implementation authorized for workbook gap remediation; specs, contracts, imports, tests, generated artifacts, and this handoff updated | Core specs, domain/guides, contracts, `internal/modules/imports`, `internal/modules/workbook/openapi_contract_test.go`, generated contract outputs, frontend phase accounting | `rg`; `sed`; `make generate`; validation commands listed below | Workbook remains a compatibility facade; import dispatcher no longer depends on workbook; linked-note/generic patch public contracts are generated | Remaining owner facades and apply journal | Continue with owner-facade extraction for linked notes, decisions, and remaining import targets. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Workbook classified as mixed compatibility/application facade with source behavior candidates for owner modules | Inspected all production files under `internal/modules/workbook` and owner module seams in timeline, entities, artifacts, evidence, links, tasksdecisions, revisions, projections, collaboration, imports | `find internal/modules/workbook`; `rg` for callers/imports/routes/symbols; `sed` inspections | Boundary diagnosis and findings recorded | TODO: characterization required before implementation movement | Start S-01 before moving any logic. |
| 2026-06-30 | Codex remediation session | Imports production dependency on workbook removed; workbook mutation/query facade remains otherwise unchanged | Touched `internal/modules/imports/routes.go`, `store.go`, `targets.go`, `boundary_test.go`, `imports_integration_test.go`; added `internal/modules/workbook/openapi_contract_test.go` | `rg internal/modules/workbook internal/modules/imports`; `gofmt`; `make backend-unit`; `make phase-slice PHASE=phase11`; `make phase-slice PHASE=phase9` | Import apply now uses a registry and owner create dispatch for timeline/entities; unsupported/no-facade targets fail closed | Owner create facades for evidence/artifacts/tasks/decisions/parties/assessments/coordination are still missing | Add facades per owner, then remove remaining fail-closed temporary registry entries when supported. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Current HTTP routes, OpenAPI-backed routes, WS path, generated-contract surfaces, and linked-notes gap recorded | Inspected `routes.go`, `contracts/openapi/cartulary.openapi.yaml`, `contracts/ws/index.schema.json`, view-schema/contract package references | `rg` for route paths, `record_changed`, generated contract references | Found current WS route `GET /ws/v1/incidents/{incident_id}` and no OpenAPI match for linked-notes route | TODO: confirm linked-notes generated/public contract posture | Add/confirm linked-notes contract characterization. |
| 2026-06-30 | Codex remediation session | Linked-notes route added to OpenAPI/generated contracts; record patch renamed to generic `patchRecord` | Touched `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `internal/gen/contracts/contracts_gen.go`, `packages/protocol-ts/src/generated/contracts.ts`, frontend E2E helpers/specs | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check` | Generated/public contract gap resolved; frontend helper names align with `patchRecord` | Frontend still issues linked-note route through existing fetch helper; no generated runtime client replacement was found in this slice | Introduce generated-client usage later only if the repo has or adds a public client abstraction for these routes. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Workbook test files inventoried and phase6/phase8/phase9 validation posture mapped | Inspected workbook `*_test.go` names/responsibilities and Make task surface | `make help`; `make help-all \| rg ...`; `make task-guide ROLE=feature-dev PHASE=phase9`; `make task-guide ROLE=local-dev PHASE=phase9`; `make explain-target TARGET=test-fast DETAIL=summary`; `make explain-target TARGET=check DETAIL=summary` | Validation commands captured in Section 8 | Resolved by tracker-only validation row below | Use narrow Make targets first for future implementation. |
| 2026-06-30 | Codex planning session | Tracker-only validation completed | Touched this tracker only | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | All passed. Artifact policy run root: `.cartulary/test-results/20260630T225618Z-p10118`; JSON shape run root: `.cartulary/test-results/20260630T225622Z-p10298` | None for tracker-only validation | Use broader validation only for future implementation slices. |
| 2026-06-30 | Codex remediation session | Remediation validation completed for generated contracts, backend unit, Phase 9 workbook, Phase 11 imports, and frontend checks | Touched files listed in this handoff; frontend phase registry/maps refreshed for guide digest drift | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-unit`; `make frontend-typecheck`; `make frontend-unit`; `make frontend-import-boundary-check`; `make phase-slice PHASE=phase11`; `make phase-slice PHASE=phase9` | All listed checks passed after refreshing frontend phase guide digests | `make json-shape-check` initially failed on stale frontend guide digest and was remediated through registry/map digest refresh | Run `make agent-finalize` and a broad gate if time/risk requires. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Auth/session/CSRF/role checks treated as observable route-facade behavior | Inspected `routes.go` and phase9 route guard test coverage | `rg` for route guards/auth references; `sed` inspections | Security behavior listed in freeze map and intentional/no_action findings | None for planning | Preserve route guard tests for any facade change. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| ---- | ------------- | ------------- | -------------------------- | ------------ | ------ | -------- | ----------- |
| 2026-06-30 | Codex planning session | Main risks are linked-notes contract gap, imports-to-workbook dependency, route DTO ownership, and mutation-store orchestration | This tracker | Planning commands listed above | Risks are tied to workstreams, slices, and open questions | TODO: OQ-01, OQ-02, OQ-03, OQ-04 | Begin with S-01 characterization before implementation. |
| 2026-06-30 | Codex remediation session | Main remaining risks are owner-facade extraction, missing import owner create facades, explicit import apply journal, and route DTO ownership split | This tracker plus touched specs/contracts/imports tests | Validation commands listed above | Public contract gaps and import-to-workbook production coupling are resolved | OQ-03 remains; owner facades/apply journal are TODO | Continue with S-02/S-03/S-04 and import owner facade expansion. |

## 11. Open Questions and Blockers

No owner contradiction was discovered in this planning pass.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| -- | ------------------- | -------------- | ---------------------------- | -------------- |
| OQ-01 | Confirm OpenAPI/generated contract posture for `POST /api/v1/records/{record_id}/linked-notes` | Route existed before the public contract mapping | OpenAPI owner evidence, route tests, frontend references, generated contract policy | DONE: OpenAPI/generated contracts and `openapi_contract_test.go` now cover it. |
| OQ-02 | Decide whether imports should keep workbook compatibility for arbitrary non-timeline imported rows or use an import-owned dispatcher | Imports depended on workbook DTO/store, which blocked clean ownership boundaries | Imports owner guidance, characterization tests, owner create facades | DONE: import-owned dispatcher adopted; production imports package no longer imports workbook. |
| OQ-03 | `TODO: determine whether route-facing workbook DTOs remain a stable public adapter or split into owner adapters after characterization` | DTO moves risk hash, validation, and generated-contract drift | Core route contract evidence, mutation API tests, OpenAPI schemas | TODO |
| OQ-04 | Confirm generated/public contract posture for non-timeline `PATCH /api/v1/records/{record_id}` | Live workbook patch path handles multiple source domains; OpenAPI evidence was timeline-oriented | OpenAPI route owner evidence, workbook mutation tests, owner module guidance | DONE: OpenAPI/generated contracts now use generic `patchRecord`; owner patch adapters remain future work under OQ-03. |
| OQ-05 | `TODO: add owner create facades and apply journal for all importable non-timeline/non-entity targets` | Current dispatcher intentionally fails closed for registered targets whose owners do not yet expose import create contracts; Core 03 now requires durable apply-journal semantics | Owner facade designs, migrations if needed, import replay/idempotency tests | TODO |

## 12. Binary Completion Criteria

The tracker is complete only when:

- Every file in `internal/modules/workbook` is inventoried or explicitly out of scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every implementation slice is behavior-preserving unless explicitly marked as requiring later authorization.
- Validation commands are discovered or marked as TODO with reason.
- Handoff sections are current enough for another agent to continue without rediscovery.

Completion status for this tracker:

| Criterion | Status | Evidence |
| --------- | ------ | -------- |
| Every current workbook file inventoried | DONE | Section 2 lists six production files and all current test files. |
| Public contract risks have owner and test posture | DONE | Sections 4, 5, and 11 list contract owners, tests, and TODO gaps. |
| Workflows have dependencies and exit criteria | DONE | Section 6. |
| Implementation slices are behavior-preserving | DONE | Section 7 marks all future slices as behavior-preserving and rollback-capable. |
| Validation commands discovered or TODO-marked | DONE | Section 8. |
| Handoff sections are current | DONE | Section 10. |
| Tracker-only validation run | DONE | `make lint-markdown`, `make generated-artifact-policy-check`, and `make json-shape-check` passed. Artifact policy run root: `.cartulary/test-results/20260630T225618Z-p10118`; JSON shape run root: `.cartulary/test-results/20260630T225622Z-p10298`. |

Post-remediation posture:

- Public contract and import-dispatcher remediation has been implemented under later authorization.
- Start remaining work with owner-facade characterization before moving workbook mutation orchestration.
- Keep observable behavior frozen: route shape, request/response envelopes, WebSocket paths and event semantics, workbook interaction behavior, storage semantics, authorization outcomes, generated contract surfaces, and harness accounting.

Files inspected during planning included:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/handoffs/workbook-module-refactor-tracker.md`
- Core spec files under `docs/spec/`
- All files under `internal/modules/workbook`
- Owner module seams in `internal/modules/timeline`, `internal/modules/entities`, `internal/modules/artifacts`, `internal/modules/evidence`, `internal/modules/links`, `internal/modules/tasksdecisions`, `internal/modules/revisions`, `internal/modules/projections`, `internal/modules/collaboration`, and `internal/modules/imports`
- Contract files under `contracts/openapi`, `contracts/ws`, and `contracts/view-schemas`
- Frontend/generated contract references under `packages/*`

Commands run during planning included:

- `find`, `rg`, `sed`, and `wc` inspections over workbook, owner modules, contracts, specs, testutil, and docs
- `make help`
- `make help-all | rg ...`
- `make task-guide ROLE=feature-dev PHASE=phase9`
- `make task-guide ROLE=local-dev PHASE=phase9`
- `make explain-target TARGET=test-fast DETAIL=summary`
- `make explain-target TARGET=check DETAIL=summary`
- `git branch --show-current`
- `git rev-parse --short HEAD`
- `git status --short`

Commands run after tracker creation:

- `make lint-markdown`
- `make generate`
- `make generate-drift`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make backend-unit`
- direct focused import package test: `go test ./internal/modules/imports -run 'TestImportsProductionPackageDoesNotImportWorkbook|TestPhase11_I_11_IMPORT_02_TargetRegistryAndEntityOwnerFacade|TestPhase11_I_11_IMPORT_02_MissingOwnerFacadeFailsClosed' -count=1`
- `make frontend-typecheck`
- `make frontend-unit`
- `make frontend-import-boundary-check`
- `make phase-slice PHASE=phase11`
- `make phase-slice PHASE=phase9`

Tracker state:

- Production remediation was run only after later authorization.
- This tracker was created and updated at `docs/handoffs/workbook-module-refactor-tracker-2.md`.
- Remaining blockers: OQ-03 and OQ-05.
