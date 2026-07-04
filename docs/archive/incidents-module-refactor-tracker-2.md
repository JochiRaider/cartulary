# incidents Module Refactoring Tracker and Handoff

Last updated: 2026-07-02T16:27:37-04:00.

Target directory: `internal/modules/incidents`.

This tracker began as a planning-only handoff artifact. The 2026-07-02
remediation session authorized internal implementation changes for Q-001
through Q-005. This tracker does not authorize further public product behavior,
route, WebSocket wire-shape, OpenAPI envelope, database migration, or generated
artifact changes.

## 1. Scope and Source Posture

Authorized remediation scope for this session:

- Add internal incident ports for workbook bootstrap and collaboration/session
  signaling.
- Wire concrete adapters from workbook/startup and collaboration at application
  composition time.
- Decouple reporting export-model construction from incident storage DTOs.
- Add characterization and boundary tests for the new seams.
- Update Core clarifications, test/harness accounting notes, and this tracker.

Non-goals:

- Do not change public HTTP routes, request/response envelopes, WebSocket wire
  event names, OpenAPI route shape, or database migrations.
- Do not hand-edit generated roots such as `internal/gen/**`,
  `packages/protocol-ts/src/generated/**`, or
  `packages/ui-contracts/src/generated/**`.
- Do not treat phase maps, phase rows, or harness grouping as runtime module
  ownership.
- Do not assume `internal/modules/incidents` is a valid permanent module
  boundary merely because the directory exists.

Authority hierarchy:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md` and implementation-support guides for terminology,
   package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans and framework files as evidence, not authority.

Repository posture:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` was used as
  the planning template and doctrine, not as proof of current repository state.
- Unknowns are recorded as `TODO:`.
- Owner-document contradictions must be recorded as
  `BLOCKED: owner contradiction`.
- Observable behavior remains frozen unless a later owner-authorized task
  explicitly changes it. Observable behavior includes route shape,
  request/response envelopes, WebSocket paths and event semantics, workbook
  interaction behavior, storage semantics, authorization outcomes, generated
  contract surfaces, and harness accounting.

Architectural finding:

`internal/modules/incidents` remains a mixture. It is mostly a legitimate
incident lifecycle, membership, visible-incident, and access facade. It is not
the current owner of workbook routes. The remediation keeps incidents as the
incident-create transaction coordinator while moving workbook preference
storage detail behind a `WorkbookBootstrapPort`, and keeps incident routes as
the semantic event source while moving WebSocket/session mechanics behind a
`CollaborationSessionPort`.

## 2. Current-State Repository Inventory

Live files discovered under `internal/modules/incidents`:

| path | current responsibility | exported/public symbols or package surface | inbound callers | outbound dependencies | tests touching it | generated artifacts or contracts touched | suspected target owner module | risk level | notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/incidents/.gitkeep` | Keeps the directory present when otherwise empty. | None. | Repository only. | None. | None. | None. | None. | low | Inventoried for completeness; no runtime behavior. |
| `internal/modules/incidents/access.go` | Incident access facade for visible-incident lookup, membership lookup, transactional open guard, and stable error classification. | `Access`, `AccessService`, `NewAccess`, `AccessFromStore`; methods include `GetVisibleIncident`, `GetIncidentMembershipForUser`, `EnsureOpenTx`, `IsIncidentClosed`, `IsIncidentNotFound`, `IsMembershipNotFound`. | Production imports from workbook, timeline, revisions, evidence, entities, imports, reporting routes, savedviews, collaboration, assessments, incidentbundles, jobapi, indicators, linked notes, host identity, merge, mentions, tasks/decisions; testutil phase stores. | `Store`, package-private `ensureIncidentOpenTx`, `postgres.DB`, `pgx.Tx`, `uuid`. | Incident package tests indirectly; peer phase tests for closed-incident and membership authorization behavior. | No generated file directly; behavior maps to Core 01 lifecycle and Core 04 authorization contracts. | `incidents` access port. | medium | Q-002 decision: keep `EnsureOpenTx(ctx, pgx.Tx, incidentID)` for now as an implementation-level same-transaction guard; it must not own transaction lifecycle. |
| `internal/modules/incidents/api.go` | Incident and membership request decoders, resource builders, API error helpers, forbidden-field validation, role/TLP token handling, and last-admin helper. | `CreateIncidentRequest`, `MembershipCreateRequest`, `MembershipPatchRequest`, `MembershipDeleteRequest`, `OptionalNullableString`, `IncidentPatchRequest`, `IncidentLifecycleRequest`, decode functions, `BuildIncidentResource`, `BuildMembershipResource`, `WouldLeaveNoIncidentAdmins`. | `routes.go`, `store.go`, package tests, phase2 store helpers. | `authn`, `httpapi`, `norm`, `uuid`, JSON decoding helpers. | `phase2_request_test.go`, `phase2_support_test.go`, `phase2_store_test.go`, `phase2_http_conformance_test.go`. | `contracts/openapi/cartulary.openapi.yaml`; generated contract reads through `internal/gen/contracts`; downstream generated TS if OpenAPI changes later. | `incidents` API/application facade. | medium-high | Incident patch/create reject workbook preference and saved-view fields, confirming those surfaces are not incident-owned mutation fields. |
| `internal/modules/incidents/boundary_guard_test.go` | Static import guard for the incidents package. | Test functions only. | Go test runner. | Go parser/token packages. | Self. | No generated artifacts. | Tests/harness. | medium | Fails if production incidents code imports `internal/modules/workbook/startup/bootstrap` or `internal/platform/ws`. |
| `internal/modules/incidents/hooks.go` | Package-private test-runtime guarded before-commit hook plumbing for store fault injection. | No exported symbols. | `routes.go` constructs `Store` with guarded hooks; `store.go` invokes package-private hook; phase2 testutil supplies the override key. | `httpapi`, `uuid`. | `phase2_integration_test.go` rollback/fault tests through phase2 testutil. | None. | Testability seam owned by incidents plus harness support. | low | Keep as test-only; do not promote this to production extension behavior. |
| `internal/modules/incidents/open_guard.go` | Package-private transactional lifecycle guard for rejecting source-state writes to closed incidents. | None; `ensureIncidentOpenTx` is unexported. | `access.go`; peers reach it only through `Access.EnsureOpenTx`. | SQLC queries, `pgx.Tx`, `uuid`, package errors. | Peer mutation tests indirectly, especially timeline/workbook/imports/evidence/revisions closed-incident behavior. | SQLC output from `db/queries/incidents_phase2.sql`; Core 01 closed-state operation matrix. | `incidents` lifecycle access port. | medium | Same-transaction shape is intentionally retained for this slice. |
| `internal/modules/incidents/phase2_logic.go` | Pure incident bootstrap defaults, idempotency request hashes, patch application, and membership/role access-error helpers. | `IncidentCreateBootstrap`, `DefaultIncidentCreateBootstrap`, `IncidentCreateIdempotencyScope`, `IncidentCreateRequestHash`, `IncidentLifecycleRequestHash`, `ApplyIncidentPatch`, `IncidentAccessError`, `RequireIncidentMembership`, `RequireIncidentRole`. | `routes.go`, `store.go`, peer route packages through `RequireIncidentMembership` and `RequireIncidentRole`, package tests. | `httpapi`, `uuid`; no storage dependency. | `phase2_request_test.go`, `phase2_unit_test.go`, route/integration tests indirectly. | OpenAPI behavior through public route responses. | `incidents` application/access helper facade. | medium | `IncidentAccessError` accepts `isDeploymentAdmin` but intentionally ignores it; preserve Core 04 deployment-admin non-substitution. |
| `internal/modules/incidents/ports.go` | Internal incident-owned seams for create-time workbook bootstrap and collaboration/session notification. | `WorkbookBootstrapPort`, `CollaborationSessionPort`, `StoreOptions`, `RouteOptions`. | `store.go`, `routes.go`, `internal/app/runtime.go`, tests and testutil helpers. | `context`, `time`, `uuid`, `pgx.Tx`. | Boundary guard tests, store rollback characterization, phase route tests. | No generated artifacts. | `incidents` application ports. | high | Q-001 and Q-003 closure point. Ports keep incident orchestration semantic while delegating workbook/session mechanics. |
| `internal/modules/incidents/routes.go` | HTTP registration and handlers for incident collection/item, lifecycle close/reopen, membership CRUD, authentication, session sliding, and semantic collaboration/session notifications. | `RegisterRoutes(options ...RouteOptions)`; unexported `Service` and handlers. | `internal/app/runtime.go` registers incidents routes with workbook and collaboration adapters; phase runtimes exercise routes. | `authn`, `httpapi`, `httpauth`, `listquery`, `pagination`, `postgres`, package store/API helpers, `CollaborationSessionPort`. | `phase2_http_conformance_test.go`, `phase2_integration_test.go`, `phase2_pagination_integration_test.go`, `phase2_extra_integration_test.go`; phase6 WS/control-boundary tests indirectly. | OpenAPI paths for incidents and memberships; WebSocket/session semantics through collaboration adapter. | `incidents` HTTP facade with collaboration/session port seam. | high | No direct `internal/platform/ws` import remains in production incidents code. |
| `internal/modules/incidents/store.go` | SQLC-backed incident/membership persistence, idempotency, audit events, lifecycle transitions, visible list/get, membership CRUD, and incident create transaction coordination. | Error vars, `Store`, `IncidentVersionConflictError`, `IncidentListPosition`, `IncidentListPageRequest`, `IncidentRecord`, `MembershipRecord`, result structs, `NewStore`, `NewStoreWithOptions`, list/get/create/update/lifecycle/membership methods. | `routes.go`, `access.go`, phase2/phase3/phase4 testutil, package tests. | `internal/gen/sql`, `authn`, `listquery`, `postgres`, `pgx`, `pgconn`, `pgtype`, package API helpers, `WorkbookBootstrapPort`. | `phase2_store_test.go`, `phase2_integration_test.go`, `phase2_extra_integration_test.go`, `phase2_pagination_integration_test.go`, `phase2_http_conformance_test.go`; peer tests through `Access`. | `db/queries/incidents_phase2.sql`; generated `internal/gen/sql/incidents_phase2.sql.go`; incident migrations; OpenAPI route semantics; workbook startup SQL input through adapter. | `incidents` persistence/application transaction owner, with workbook bootstrap port seam. | high | `CreateIncident` now requires a workbook bootstrap port and calls it inside the same transaction; missing port fails create. |
| `internal/modules/incidents/phase2_extra_integration_test.go` | Extra integration coverage for membership create replay/divergent conflict and incident/membership audit before/after payloads. | Test functions only. | Go test runner. | Phase2 HTTP harness, public incident routes. | Self. | Public HTTP and audit behavior evidence. | Tests/harness. | medium | Preserve before changing membership idempotency, audit, or route envelopes. |
| `internal/modules/incidents/phase2_http_conformance_test.go` | HTTP conformance for incident create bootstrap, route inventory envelopes, membership admin, membership CRUD, workbook preferences, extension discovery, and deployment-admin control boundaries. | Test functions only. | Go test runner. | `phase2test`, HTTP harness, route inventory helpers. | Self. | OpenAPI-like route envelope behavior; phase2 route inventory. | Tests/harness with split runtime ownership. | high | Includes workbook preference and extension discovery evidence even though runtime owners are `workbook` and `extensions`. Treat as accounting, not architecture. |
| `internal/modules/incidents/phase2_integration_test.go` | Integration tests for incident create rollback/idempotency/key conflict, authorization re-derivation, incident patch, membership no-op, extension discovery, and reserved route families. | Test functions only. | Go test runner. | `phase2test`, HTTP harness, fault dependency override. | Self. | Public route behavior and extension discovery evidence. | Tests/harness. | high | Rollback behavior is complemented by the new port-failure store characterization. |
| `internal/modules/incidents/phase2_inventory_helpers_test.go` | Shared route-inventory and fixture helpers for phase2 incident tests, including workbook preference, timeline, and WebSocket control-boundary helpers. | Test helpers only. | Phase2 incident tests. | `phase2test`, `timeline`, `platformws`, HTTP harness. | `phase2_http_conformance_test.go`, `phase2_integration_test.go`. | Harness accounting through route inventories. | Tests/harness. | medium | Test helper imports are not production ownership evidence. |
| `internal/modules/incidents/phase2_pagination_integration_test.go` | Incident and membership pagination/search/status/cursor characterization, including live membership changes during continuation. | Test functions only. | Go test runner. | Phase2 HTTP harness, public incident and membership routes. | Self. | List/query contract evidence. | Tests/harness. | high | Required before list query or pagination refactor. |
| `internal/modules/incidents/phase2_request_test.go` | Unit/contract tests for incident/membership decoders, patch logic, string/list-query contracts, OpenAPI workbook preference paths, extension discovery contract, and reserved-route dispatch. | Test functions and local OpenAPI helpers. | Go test runner. | `internal/gen/contracts`, `httpapi`, extension/profile helpers, package decode/build functions. | Self. | Reads generated OpenAPI contracts; generated files are downstream and not editable. | Tests/harness and contract checks. | high | Some assertions cover non-incident runtime owners as phase2 evidence accounting. |
| `internal/modules/incidents/phase2_store_test.go` | Store tests for incident create bootstrap, stable location, idempotency replay/divergence, side effects, patch conflicts, membership stale versions, and workbook bootstrap port failure rollback. | Test functions only. | Go test runner. | `phase2storetest`, package store/decoder APIs, workbook startup store checks for preference bootstrap rows. | Self. | Storage behavior evidence; SQLC/migration behavior through store. | Tests/harness. | high | New Q-001 characterization asserts a failing bootstrap port rolls back incident, membership, preference, audit, and idempotency writes. |
| `internal/modules/incidents/phase2_support_test.go` | Focused tests for membership patch/delete decoders and last-admin helper. | Test functions only. | Go test runner. | Package decode/helper functions. | Self. | Membership mutation contract evidence. | Tests/harness. | medium | Protects last-admin semantics. |
| `internal/modules/incidents/phase2_unit_test.go` | Unit test for incident access decision error classification. | Test functions only. | Go test runner. | Package access-error helper. | Self. | Authorization error envelope evidence. | Tests/harness. | medium | Preserves hidden incident `404 incident_not_found` versus visible insufficient-role `403 authorization_denied`. |

Generated and derived surfaces observed:

- `contracts/openapi/cartulary.openapi.yaml` contains incident, membership,
  workbook preference, workbook startup, view, saved-view, and extension paths.
- `internal/gen/contracts/**` and `internal/gen/sql/**` are generated and must
  not be hand-edited.
- `db/queries/incidents_phase2.sql` is the authored SQLC input for incident
  storage.
- `db/queries/workbook_startup_phase8.sql` remains relevant to the concrete
  workbook preference adapter.
- `tools/phase2_test_map.json` and `tools/phase11_test_map.json` record
  evidence/accounting posture. They are not runtime ownership proof.

## 3. Workbook Boundary Diagnosis

Diagnosis:

The current `incidents` package is a legitimate incident lifecycle,
membership, and access facade plus a transport-adjacent HTTP adapter and a
mutation coordinator for incident creation. Workbook route ownership remains in
`internal/modules/workbook`. The remaining workbook seam is intentionally an
incident-create orchestration port: incidents owns the create transaction
boundary, while workbook/startup owns preference storage details.

Evidence:

- `internal/modules/incidents/routes.go` registers only incident routes under
  `/api/v1/incidents`.
- `internal/modules/workbook/routes.go` registers workbook query, row,
  mutation, workbook preference, and workbook startup paths.
- `internal/modules/extensions/routes.go` registers `/api/v1/extensions`.
- `api.go` rejects saved-view and workbook-preference fields in incident
  create/patch payloads.
- `store.go` calls `WorkbookBootstrapPort` inside the incident-create
  transaction.
- `internal/modules/workbook/startup/bootstrap.NewIncidentCreatePreferencesPort`
  adapts existing `PreferencesTx` behavior without moving ownership into
  incidents.
- Phase maps and test rows remain evidence accounting only.

Decision table:

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Incident list/get visibility | `store.go`, `routes.go`, `access.go` | `incidents` | keep | `ListVisibleIncidents`, `GetVisibleIncident`, incident routes, broad peer use of `Access`. | Legitimate incident facade behavior. |
| Incident metadata create/patch | `api.go`, `store.go`, `routes.go`, `phase2_logic.go` | `incidents` | keep | Core 01 incident resource/create/mutability; OpenAPI incident paths. | Preserve field boundary and idempotency semantics. |
| Incident lifecycle close/reopen | `api.go`, `store.go`, `routes.go`, `open_guard.go`, `phase2_logic.go` | `incidents` lifecycle port | keep | Core 01 active/closed lifecycle; `ErrIncidentClosed`; `Access.EnsureOpenTx`; route handlers. | Q-002 keeps same-transaction `pgx.Tx` guard for now. |
| Incident membership CRUD and role authorization | `api.go`, `store.go`, `routes.go`, `phase2_logic.go`, `access.go` | `incidents` membership/access facade | keep | Membership routes, membership SQL, Core 04 membership-derived auth. | Legitimate incident boundary. |
| Same-transaction workbook preference bootstrap on incident create | `store.go` through `WorkbookBootstrapPort`; adapter in `workbook/startup/bootstrap` | `incidents` orchestration plus `workbook/startup` implementation | split | Core 01 create contract requires preference objects; workbook routes/storage behavior live under workbook/startup. | Q-001 implemented; port failure rolls back the transaction. |
| Workbook preference routes | `internal/modules/workbook/routes.go`; phase2 tests under incidents package still exercise them | `workbook` and `workbook/startup` | move already done; keep accounting note | Production route registration in workbook; OpenAPI workbook-preference tags. | Do not move back to incidents. |
| Workbook startup route and fallback behavior | `internal/modules/workbook/routes.go`, `internal/modules/workbook/startup/**` | `workbook/startup`, Core 03 startup order | move already done; keep accounting note | `GET /workbook-startup` registered in workbook; Core 03 startup order. | Incidents is only an access/bootstrap participant. |
| View-schema query/row/bulk/clipboard behavior | `internal/modules/workbook/routes.go` and peer stores | `workbook`, projections, source-owner modules | defer/no incident move | Workbook route registration outside incidents. | Incidents supplies membership/open guard only. |
| Saved views | `internal/modules/savedviews/routes.go` with `incidents.Access` | `savedviews` | defer/no incident move | Savedviews imports incident access facade. | Saved views are not incident-owned access control. |
| Projection refresh behavior | Peer modules and projections/workbook stores | `projections` plus source owner modules | defer/no incident move | No production projection refresh logic found in incidents. | Closed-incident guard can affect source mutation admission. |
| Revisions/change sets | `internal/modules/revisions/**` with `incidents.Access` | `revisions` | defer/no incident move | Revisions imports incident access and `ErrIncidentClosed`. | Incidents provides lifecycle authorization only. |
| Collaboration/session side effects | `routes.go` calls `CollaborationSessionPort`; adapter in `internal/modules/collaboration` owns hub/session mechanics | `collaboration`/platform WS with incidents event source | split | Close route still emits incident-close semantics; membership delete still revokes incident access for active sessions. | Q-003 implemented; incidents production code no longer imports `internal/platform/ws`. |
| Reporting incident metadata | `internal/modules/reporting` owns `IncidentMetadataSnapshot`; reporting routes still use `incidents.Access` for authorization | `reporting` export model plus incidents access facade | split | Reporting store/export model no longer imports `incidents.IncidentRecord`; routes still import incidents for access checks. | Q-005 implemented. |

Repo/prior-plan mismatch:

- Prior tracker material mentions historical/remediated surfaces such as
  `startup.go`; no such live file exists in `internal/modules/incidents`.
- The previous direct `store.go` dependency on
  `workbook/startup/bootstrap.PreferencesTx` has now been replaced by a port
  and adapter.

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner and affected code | Existing tests or required characterization |
| --- | --- | --- |
| Incident HTTP routes | `GET/POST /api/v1/incidents`, `GET/PATCH /api/v1/incidents/{incident_id}`, `POST /close`, `POST /reopen`, `GET/POST /memberships`, `PATCH/DELETE /memberships/{user_id}` in `routes.go`. | Existing phase2 HTTP/integration/pagination tests. No route shape change authorized. |
| Workbook preference/startup paths | `/workbook-preferences/default`, `/workbook-preferences/me`, `/workbook-startup` registered by workbook, affected by incident-create bootstrap. | Existing phase2 conformance/store tests and phase8 workbook startup tests. New bootstrap-port failure test protects rollback. |
| WebSocket/session semantics | Incident close notifies `incident_closed`; membership delete revokes affected incident access without broad session revocation. | Phase2 control-boundary helpers and phase6 collaboration/socket tests. New boundary guard prevents direct incidents-to-platform-WS imports. |
| Workbook row/query/mutation behavior | Owned by workbook and source modules; incidents supplies membership/open guard. | Phase8/phase9 workbook tests and peer mutation tests if lifecycle guard changes. |
| Saved-view/view-schema behavior | Owned by savedviews/workbook/contracts; incidents supplies access checks. | Savedviews and workbook phase tests if access facade changes. |
| Projection refresh behavior | No direct projection refresh logic in incidents. | Owner-module tests required if source-mutation lifecycle admission changes. |
| Authorization checks | Hidden non-membership remains `incident_not_found`/404; insufficient role remains `authorization_denied`/403; deployment admin is not sufficient for incident access. | `phase2_unit_test.go`, phase2 control-boundary tests, peer route tests. |
| Revision/change-set behavior | Revisions imports incident access and lifecycle errors; incidents does not own revision history. | Phase7 revisions tests if lifecycle guard signature or error mapping changes. |
| Storage behavior | Incident create creates incident, admin membership, default/user workbook preferences, audit, and route idempotency atomically; lifecycle states are `active` and `closed`; last-admin and version conflicts remain. | Existing phase2 store tests plus new bootstrap-port failure rollback test. |
| Generated protocol/view contracts | OpenAPI and generated SQL/contract outputs are downstream; no generated hand edits. | `generated-artifact-policy-check`, `json-shape-check`, drift checks if authored inputs change. |
| Harness/test accounting | Phase2 maps account workbook/extension rows as owner-specific evidence, not incidents architecture. Phase11 maps account reporting-owned snapshot model support evidence. | `json-shape-check`, phase ledger/schedule drift checks, phase slices. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Workbook bootstrap was a direct cross-module import from incidents. | Q-001 original evidence: `store.go` imported `workbook/startup/bootstrap`. Current evidence: `store.go` depends on `WorkbookBootstrapPort`, adapter lives in workbook/startup. | High if unresolved: incident create would keep coupling to workbook internals. | must_fix | `incidents` orchestration port plus `workbook/startup` adapter. | DONE: port, adapter, app wiring, test helper wiring, boundary guard, and rollback characterization added. |
| Collaboration/session side effects were hidden inside incident routes. | Q-003 original evidence: direct hub calls in routes. Current evidence: `routes.go` calls `CollaborationSessionPort`, adapter lives in collaboration. | High if unresolved: future multi-node/session security changes remain tied to route handlers. | must_fix | `collaboration` session notifier with incidents semantic event source. | DONE: port, adapter, app wiring, helper methods, and boundary guard added. |
| Peer modules broadly import `incidents.Access`. | Live imports from workbook, timeline, revisions, evidence, entities, imports, reporting routes, savedviews, collaboration, assessments, incidentbundles, jobapi, indicators, artifacts, tasks/decisions. | Medium: facade is useful, but it concentrates authorization and lifecycle dependencies. | intentional/no_action | `incidents` access port. | Keep peer usage on `Access`; avoid new concrete `Store` dependencies. |
| `Access.EnsureOpenTx` exposes `pgx.Tx`. | `access.go` method signature includes `pgx.Tx`; peer stores use same transaction shape. | Medium: storage abstraction leaks across domain modules. | defer | `incidents` lifecycle port or future platform transaction abstraction. | Q-002 accepted: document as implementation-level same-transaction guard; do not abstract prematurely. |
| Reporting consumed `incidents.IncidentRecord`. | Q-005 original evidence in `reporting/redaction.go` and `reporting/store.go`. Current evidence: reporting owns `IncidentMetadataSnapshot`; routes only import incidents for access checks. | Medium if unresolved: reporting export models would drift with incident storage DTO churn. | must_fix | `reporting` export model and incident metadata snapshot. | DONE: reporting store/export model now use a reporting-owned DTO; support test proves stable export hash for identical metadata. |
| Transport/platform imports exist in incident route/API code. | `routes.go` imports HTTP/auth/session/pagination packages; `api.go` imports `httpapi` and `authn`. | Medium: acceptable for HTTP facade, unsuitable for pure domain logic. | intentional/no_action | `incidents` route/API adapter layer. | Keep platform imports out of pure helpers; no further split authorized. |
| Direct SQL/storage coupling exists in `store.go`. | `store.go` uses SQLC, `postgres`, `pgx`, audit/idempotency store behavior. | Medium: expected in persistence layer, but must not leak into peer modules. | intentional/no_action | `incidents` persistence layer. | Keep SQLC behind store/access facades; do not hand-edit generated SQL output. |
| Phase2 tests include workbook and extension route evidence. | Phase2 route/request tests and public route inventory include workbook preference and extension discovery entries. | Medium: future agents may infer wrong runtime ownership. | should_fix | Harness/evidence accounting with runtime owners `workbook` and `extensions`. | DONE for this slice: phase2 map notes label workbook/extension rows as owner-specific evidence, not incidents architecture. |
| Old tracker mentions deleted/moved incident surfaces. | Prior tracker references historical `startup.go` and previous workbook/startup ownership in incidents. | Medium: stale prior evidence can mislead future implementation. | should_fix | Documentation/handoff. | DONE: this tracker supersedes those assumptions and records the live mismatch. |
| Authorization helpers ignore deployment-admin bypass. | `IncidentAccessError` takes `isDeploymentAdmin` but assigns `_ = isDeploymentAdmin`; Core 04 says deployment admin is not sufficient for incident access. | Low: correct behavior but easy to misread. | intentional/no_action | `incidents` access facade with Core 04 auth posture. | Preserve non-substitution and keep tests for 404/403 behavior. |
| Test-only route helpers span timeline, workbook, WebSocket, and incident fixtures. | `phase2_inventory_helpers_test.go` imports peer modules and builds cross-module route fixtures. | Low-medium: acceptable test harness support; dangerous if promoted to production architecture. | defer | Tests/harness. | Do not move production code based on helper shape; consider owner-aligned test relocation only with a separate harness plan. |

No `BLOCKED: owner contradiction` was found.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Planning/doc baseline | root | None | WF-01, WF-02 | Create and maintain this tracker as the current handoff artifact. | This tracker; framework; `docs/domain.md`; Core docs. | `make lint-markdown`. | DONE: tracker created and updated for remediation. |
| WF-01 | Package inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every live file in `internal/modules/incidents` and distinguish production from tests. | `internal/modules/incidents/**`. | `find`; `rg --files`; targeted `sed`. | DONE: Section 2 includes every live file. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05, WF-07 | Map incident, membership, workbook preference, startup, extension, WS, auth, storage, and generated surfaces to owners. | Incident routes/store; workbook/extensions routes; OpenAPI; Core specs. | Targeted `rg`; phase task guides. | DONE: Section 4 freeze map updated. |
| WF-03 | Characterization gap analysis | chain | WF-02 | WF-05, WF-06, WF-07 | Identify and close characterization gaps for moved seams. | Phase2 store tests; boundary guard tests; reporting tests. | Phase slices and targeted unit tests through Make. | DONE for Q-001, Q-003, Q-005 support coverage. |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05, WF-06 | Classify coupling risks and distinguish intentional facades from accidental ownership. | `access.go`, `routes.go`, `store.go`, reporting, peer imports. | Import guard tests; targeted searches. | DONE: Section 5 records closure state. |
| WF-05 | Facade or ownership redesign | chain | WF-03, WF-04 | WF-06 | Implement durable seams for workbook bootstrap, collaboration/session, and reporting snapshots. | Incidents ports/store/routes; workbook adapter; collaboration adapter; reporting store/redaction. | Phase2, phase6, phase8, phase11 slices. | DONE: validation passed. |
| WF-06 | Slice sequencing | chain | WF-05 | WF-07, WF-08 | Sequence changes so each slice is behavior-preserving and rollbackable. | This tracker and touched modules. | Section 7 commands. | DONE: slices closed or deferred with rollback notes. |
| WF-07 | Harness/accounting update | parallel | WF-02, WF-03 | WF-08 | Preserve evidence while clarifying owner labels. | `tools/phase2_test_map.json`, `tools/phase11_test_map.json`. | `json-shape-check`; phase drift checks. | DONE: schedules/ledgers regenerated through Make and drift checks passed. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Validate narrow owner slices, broaden appropriately, and record outcomes. | This tracker; test result artifacts. | See Section 8. | DONE: `make check` and `make agent-finalize` passed. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | None | Create/update this tracker and closure matrix. | `docs/handoffs/incidents-module-refactor-tracker-2.md`. | None to runtime. | Markdown lint. | `make lint-markdown`. | Revert tracker edits. | DONE after docs lint. |
| S-02 | S-01 | Add characterization and import-boundary guards before/with code movement. | `phase2_store_test.go`, `boundary_guard_test.go`, reporting boundary guard and support test. | Tests can overfit implementation if they assert private mechanics unrelated to boundary ownership. | Bootstrap-port rollback test; import guard tests; reporting stable-hash support test. | Included in phase/backend slices. | Revert tests with the related code slice if the seam is abandoned. | DONE: phase and broad validations passed. |
| S-03 | S-02 | Introduce `WorkbookBootstrapPort`, store options, workbook/startup adapter, runtime wiring, and test helper wiring. | `incidents/ports.go`, `incidents/store.go`, `workbook/startup/bootstrap`, `internal/app/runtime.go`, testutil helpers. | Incident create atomicity, idempotency replay, default/user preference rows, audit side effects. | Preserve phase2 store/HTTP tests; new failure rollback test. | `make phase-slice PHASE=phase2`; `make service-backed-slice PHASE=phase2`; `make phase-slice PHASE=phase8`. | Revert to direct adapter call only if port weakens atomicity. | DONE: phase2, service-backed phase2, and phase8 passed. |
| S-04 | S-02 | Introduce `CollaborationSessionPort`, collaboration adapter, runtime wiring, and replace direct route-to-hub calls. | `incidents/ports.go`, `incidents/routes.go`, `collaboration/incident_session_notifier.go`, `internal/app/runtime.go`. | WebSocket terminal event semantics, revocation scope, session preservation for other incidents. | Preserve phase2 control-boundary and phase6 socket tests; import guard prevents direct platform WS import. | `make phase-slice PHASE=phase6`; `make service-backed-slice PHASE=phase6`. | Revert adapter wiring if event ordering or scope drifts. | DONE: phase6 and service-backed phase6 passed. |
| S-05 | S-02 | Replace reporting use of `incidents.IncidentRecord` with `IncidentMetadataSnapshot`. | `reporting/store.go`, `reporting/redaction.go`, reporting tests. | Export model JSON/hash drift, reporting route access checks. | Preserve reporting integration tests; add stable export-model hash support test; import guard allows incidents only in reporting routes. | `make phase-slice PHASE=phase11`. | Revert DTO swap if export hashes drift for identical input. | DONE: phase11 and service-backed phase11 passed. |
| S-06 | S-03, S-04 | Relabel harness/accounting rows for workbook/extension and reporting support evidence without runtime ownership changes. | `tools/phase2_test_map.json`, `tools/phase11_test_map.json`. | Dropping evidence, creating unmapped tests, or generated ledger/schedule drift. | Preserve phase maps and generated drift checks. | `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`. | Revert map note changes if they create drift and regenerate only through Make if authorized. | DONE: schedules/ledgers regenerated through Make and drift checks passed. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Authored Markdown docs including this tracker and Core edits. | yes | Passed. |
| unit | `make backend-unit` | Boundary guards, pure helpers, generated-contract reads, reporting redaction tests. | yes for this change | Covered through phase slices and `make check`; no standalone backend-unit rerun after broad check. |
| integration | `make phase-slice PHASE=phase2` | Incident create/list/get/patch, membership admin, bootstrap preference evidence, phase2 route accounting. | yes | Passed: `.cartulary/test-results/20260702T194313Z-p1106763`, 69 tests. |
| integration | `make service-backed-slice PHASE=phase2` | Service-backed incident route/store behavior. | yes | Passed: `.cartulary/test-results/20260702T194403Z-p1132009`, 55 tests. |
| e2e/browser | `make browser-e2e-stateful` | Browser/session/WebSocket behavior. | conditional | Covered inside `make check`; standalone run skipped because phase6 service-backed and full check passed. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make generate-drift` if authored contract/SQL/generation inputs change | Generated artifact policy, JSON manifests, generated output drift. | yes for map changes | Passed after `make phase-schedules` and `make phase-ledgers` regenerated downstream artifacts. |
| import-boundary/static | `make lint` | Static/import checks and general lint. | conditional | Initial `make check` found `newStoreWithHooks` unused; removed it. `make lint` then passed at `.cartulary/test-results/20260702T195201Z-p1313780`. |
| phase-specific lifecycle | `make phase-slice PHASE=phase3` | Timeline and closed-incident source-mutation behavior. | conditional | Skipped as standalone because Q-002 did not change guard signature or behavior; covered by `make check`. |
| phase-specific collaboration | `make phase-slice PHASE=phase6`; `make service-backed-slice PHASE=phase6` | Collaboration, presence, WebSocket behavior. | yes | Passed: `.cartulary/test-results/20260702T194443Z-p1147886` and `.cartulary/test-results/20260702T194537Z-p1165246`. |
| phase-specific workbook startup | `make phase-slice PHASE=phase8`; `make service-backed-slice PHASE=phase8` if needed | Saved views and workbook startup/default surface behavior. | yes for phase slice | Phase slice passed: `.cartulary/test-results/20260702T194630Z-p1180376`; service-backed standalone skipped because full `make check` passed. |
| phase-specific reporting | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11` if needed | Snapshot/reporting export model and route behavior. | yes for phase slice | Passed: `.cartulary/test-results/20260702T194711Z-p1196627` and `.cartulary/test-results/20260702T194752Z-p1211403`. |
| broad check | `make test-fast`; `make check` | Broad local validation. | yes after narrow slices if feasible | `make test-fast` passed at `.cartulary/test-results/20260702T194840Z-p1224208`, 973 tests. Final `make check` passed at `.cartulary/test-results/20260702T195250Z-p1322957`, 976 tests. |
| finalizer | `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T195250Z-p1322957` | End-of-run validation/finalizer maintenance. | yes | Passed at `.cartulary/test-results/20260702T195542Z-p1407183`; retained successful check run evidence refreshed generated baseline files. Repeated after tracker closeout and passed unchanged at `.cartulary/test-results/20260702T195940Z-p1410879`. |

## 9. Top-Level Work Tracker

Status values: `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`,
`DROPPED`.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| IT2-001 | Create tracker and initial live inventory. | WF-00/WF-01 | DONE | none | This tracker; prior `make lint-markdown` result. | Every original incident file inventoried. |
| IT2-002 | Add workbook bootstrap port and adapter. | WF-05 | DONE | IT2-001 | `incidents/ports.go`, `store.go`, workbook bootstrap adapter, runtime wiring. | Incident create uses port and no production incidents code imports workbook bootstrap. |
| IT2-003 | Characterize workbook bootstrap failure rollback. | WF-03 | DONE | IT2-002 | `phase2_store_test.go`. | Failing port leaves no committed incident, membership, preference, audit, or idempotency writes. |
| IT2-004 | Keep lifecycle transaction guard shape and document Q-002. | WF-05 | DONE | IT2-001 | `access.go`, this tracker. | Guard remains same-transaction and is not documented as public contract. |
| IT2-005 | Add collaboration/session port and adapter. | WF-05 | DONE | IT2-001 | `incidents/ports.go`, `routes.go`, `collaboration/incident_session_notifier.go`, runtime wiring. | Incidents routes use semantic port; no production incidents code imports platform WS. |
| IT2-006 | Add incident import boundary guard. | WF-04 | DONE | IT2-002, IT2-005 | `incidents/boundary_guard_test.go`. | Guard fails on direct workbook bootstrap or platform WS imports. |
| IT2-007 | Relabel phase2 workbook/extension evidence accounting. | WF-07 | DONE | IT2-001 | `tools/phase2_test_map.json`. | Map notes say package location is not runtime ownership proof. |
| IT2-008 | Decouple reporting from incident storage DTOs. | WF-05 | DONE | IT2-001 | `reporting/redaction.go`, `reporting/store.go`. | Store/export model use `IncidentMetadataSnapshot`; routes keep access check import only. |
| IT2-009 | Add reporting boundary and hash-stability support coverage. | WF-03/WF-04 | DONE | IT2-008 | `reporting/boundary_guard_test.go`, `redaction_test.go`, `tools/phase11_test_map.json`. | Guard allows incidents import only in reporting routes; identical metadata snapshot yields stable export model/hash. |
| IT2-010 | Add Core behavior-level clarifications. | WF-00/WF-02 | DONE | IT2-001 | `docs/spec/01_architecture_storage_and_view_contracts.md`. | Core text clarifies atomic create observability and reporting export-model boundary without Go package paths. |
| IT2-011 | Run narrow and broad validations. | WF-08 | DONE | IT2-002 through IT2-010 | Section 8 commands and run roots. | Validation results and run roots recorded. |
| IT2-012 | Final handoff update. | WF-08 | DONE | IT2-011 | This tracker and final response. | Commands, skipped checks, blockers, and rollback notes are current. |
| IT2-013 | Fail fast when incident HTTP routes are registered without required internal ports. | WF-05 | DONE | IT2-002, IT2-005 | `routes.go`, `phase2_unit_test.go`; `make phase-slice PHASE=phase2` at `.cartulary/test-results/20260702T201521Z-p1420768`. | Route registration returns a setup error before incident create can run with missing workbook bootstrap or collaboration/session wiring. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | User authorized implementation for Q-001 through Q-005 with no public route/WS/OpenAPI/DB migration changes expected. | Planning framework, domain/spec docs, prior tracker, this tracker. | `rg`, `find`, `sed`, `git status --short --branch`, `date -Iseconds`. | Scope and authority updated; implementation changes are internal seams, docs, tests, and harness maps. | Superseded by validation closeout row. | Run Section 8 validation plan. |
| 2026-07-02T15:56:19-04:00 | Codex remediation session | Remediation implementation and validation complete. | This tracker; final validation artifacts. | `make check`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T195250Z-p1322957`. | Full check passed and finalizer refreshed retained-run evidence. | None known. | Handoff through final response. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | Incidents remains lifecycle/membership/access facade and create transaction coordinator; workbook and collaboration mechanics now sit behind ports. | `internal/modules/incidents/**`, workbook bootstrap adapter, collaboration adapter, runtime wiring. | Targeted `rg`, `sed`, `gofmt`. | Q-001 and Q-003 implementation seams added. | Superseded by validation closeout row. | Run phase2, phase6, and phase8 slices. |
| 2026-07-02T15:56:19-04:00 | Codex remediation session | Backend boundary changes validated. | `incidents/ports.go`, `store.go`, `routes.go`, workbook adapter, collaboration adapter, runtime wiring. | Phase2, phase6, phase8 slices; service-backed phase2/phase6. | Incident bootstrap and collaboration/session behavior passed narrow and broad checks. | None known. | Preserve ports as implementation seams; do not expose as public product contracts. |
| 2026-07-02T16:16:16-04:00 | Codex implementation session | Incident route composition now fails fast if either required internal port is missing. | `internal/modules/incidents/routes.go`, `internal/modules/incidents/phase2_unit_test.go`, this tracker. | `make format`; `make phase-slice PHASE=phase2`. | Phase2 slice passed at `.cartulary/test-results/20260702T201521Z-p1420768`. | None known. | Run remaining drift/static and broad validation before final handoff. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | No public route, OpenAPI, SQL, migration, or generated output hand edit is intended. Core behavior-level clarifications were added. | `docs/spec/01_architecture_storage_and_view_contracts.md`, `contracts/openapi/cartulary.openapi.yaml` inspected earlier. | Targeted `rg`; drift commands later completed. | Authored docs changed; generated roots not hand-edited. | Superseded by validation closeout row. | Run generated policy/shape/drift checks. |
| 2026-07-02T15:56:19-04:00 | Codex remediation session | Generated policy, JSON shape, generated drift, ledger drift, and schedule drift checks passed after Make-owned regeneration. | `docs/testing/phase2_coverage_ledger.md`, `docs/testing/phase11_coverage_ledger.md`, `tools/execution_topology_render_index.json`, finalizer baseline outputs. | `make generated-artifact-policy-check`; `make phase-schedules`; `make phase-ledgers`; drift checks; `make generate-drift`. | No generated file was hand-edited; downstream artifacts were regenerated through Make. | None known. | Keep generated files as produced by Make/finalizer. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | New incident bootstrap rollback test, import guards, reporting hash-stability support test, and phase map notes added. | Incident phase2 tests, reporting tests, `tools/phase2_test_map.json`, `tools/phase11_test_map.json`. | `rg`, `sed`, `gofmt`. | Coverage added for Q-001, Q-003, Q-004, Q-005. | Superseded by validation closeout row. | Run JSON shape, phase drift, and phase slices. |
| 2026-07-02T15:56:19-04:00 | Codex remediation session | Harness accounting and owner validations passed. | Phase maps, generated ledgers/schedules, incident/reporting tests. | `make json-shape-check`; phase drift checks; phase2/6/8/11 slices; service-backed phase2/6/11; `make test-fast`; `make check`. | Narrow and broad tests passed; first check staticcheck failure was fixed by removing unused helper. | None known. | Maintain phase map notes if moving tests later. |
| 2026-07-02T16:16:16-04:00 | Codex implementation session | Added support coverage that incident route registration rejects missing workbook bootstrap or collaboration/session ports before auth key or database setup. | `internal/modules/incidents/phase2_unit_test.go`. | `make phase-slice PHASE=phase2`. | Passed at `.cartulary/test-results/20260702T201521Z-p1420768`, 69 tests. | None known. | Continue with drift/static and broad validation. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | Membership-derived authorization, hidden incident `404`, insufficient role `403`, deployment-admin non-substitution, and same-transaction open guard remain unchanged. | `access.go`, `phase2_logic.go`, incident route/store tests. | Targeted `rg`, `sed`. | Q-002 closed by documentation/decision only; no guard signature change. | None known. | Preserve in phase validation. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T15:37:48-04:00 | Codex remediation session | Q-001 through Q-005 are implemented or intentionally documented; validation and generated drift response remained. | This tracker and touched implementation files. | Validation commands later completed; see Section 8. | No owner contradiction found. | Superseded by validation closeout row. | Record results, run roots, skipped checks, and any required generated updates. |
| 2026-07-02T15:56:19-04:00 | Codex remediation session | No open owner blockers remain for the requested remediation. | This tracker, touched implementation files, validation artifacts. | Full validation list in Section 8. | `VAL-001` closed; `make check` and finalizer passed. | None known. | Future work should treat transaction abstraction or test relocation as separate authorized tasks. |
| 2026-07-02T16:27:37-04:00 | Codex implementation session | Additional fail-fast route-port hardening is implemented and validated. | `routes.go`, `phase2_unit_test.go`, this tracker, finalizer-refreshed support files. | `make test-fast`; `make check`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T202405Z-p1590072`. | Full check passed at `.cartulary/test-results/20260702T202405Z-p1590072`; finalizer passed at `.cartulary/test-results/20260702T202658Z-p1673724`. | None known. | Handoff through final response. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| Q-001 | Should incident create use a workbook bootstrap port instead of importing workbook startup internals? | Keeps incident create responsible for orchestration while workbook/startup owns preference storage detail. | Owner direction accepted in remediation request; phase2/phase8 validation. | CLOSED: implemented `WorkbookBootstrapPort` and adapter; validation passed. |
| Q-002 | Should `Access.EnsureOpenTx` continue exposing `pgx.Tx`? | Same-transaction lifecycle check is required for source mutations; premature abstraction risks weakening closed-incident safety. | Owner direction accepted in remediation request; peer mutation tests if later changed. | CLOSED: retained and documented as implementation-level same-transaction guard, not product contract. |
| Q-003 | Should WebSocket close/revoke mechanics move behind a collaboration/session port? | Separates durable incident decisions from hub/session mechanics and prepares for future scaling. | Owner direction accepted in remediation request; phase6 validation. | CLOSED: implemented `CollaborationSessionPort` and collaboration adapter; validation passed. |
| Q-004 | Should phase2 workbook/extension tests be relabeled as owner-specific evidence? | Prevents package location from being misread as runtime architecture. | Harness map validation and phase drift checks. | CLOSED: phase2 map notes updated; validation passed. |
| Q-005 | Should reporting stop consuming `incidents.IncidentRecord` directly? | Keeps reporting export models stable against incident storage DTO churn. | Reporting tests and phase11 validation. | CLOSED: reporting-owned `IncidentMetadataSnapshot` added; validation passed. |
| VAL-001 | Validation and drift closeout. | The implementation was not ready for final handoff until Make-owned checks ran. | Section 8 commands and run-root artifacts. | CLOSED: all required validations passed; initial staticcheck failure was fixed. |

## 12. Binary Completion Criteria

Completion checklist for this remediation tracker:

- DONE: every live file in `internal/modules/incidents` is inventoried,
  including `.gitkeep`, `ports.go`, `boundary_guard_test.go`, and all
  `phase2_*_test.go` files.
- DONE: every discovered public contract risk has an owner candidate and test
  posture in Sections 3 and 4.
- DONE: every proposed workflow has dependencies, validation posture, and a
  handoff checkpoint.
- DONE: every implementation slice is behavior-preserving unless later
  authorization explicitly changes that.
- DONE: validation commands are discovered from Make-owned public targets or
  marked conditional.
- DONE: Q-001 through Q-005 have closure decisions and implementation status.
- DONE: validation results and run roots are recorded in Section 8 and the
  handoff log.
- DONE: no generated file was hand-edited.

Files inspected or touched during remediation:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/handoffs/incidents-module-refactor-tracker-2.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `internal/app/runtime.go`
- `internal/modules/incidents/**`
- `internal/modules/workbook/startup/bootstrap/bootstrap.go`
- `internal/modules/collaboration/incident_session_notifier.go`
- `internal/modules/reporting/redaction.go`
- `internal/modules/reporting/store.go`
- `internal/modules/reporting/redaction_test.go`
- `internal/modules/reporting/boundary_guard_test.go`
- `internal/testutil/phase2test/**`
- `internal/testutil/phase2storetest/**`
- `internal/testutil/phase3test/runtime.go`
- `internal/testutil/phase3storetest/runtime.go`
- `internal/testutil/phase4test/runtime.go`
- `internal/testutil/phase4storetest/runtime.go`
- `internal/modules/entities/merge/merge_protected_set_test.go`
- `internal/modules/workbook/routes.go`
- `internal/modules/extensions/routes.go`
- `contracts/openapi/cartulary.openapi.yaml`
- `db/queries/incidents_phase2.sql`
- `db/queries/workbook_startup_phase8.sql`
- `tools/phase2_test_map.json`
- `tools/phase11_test_map.json`

Commands run so far in this remediation session:

- `git status --short --branch`
- `date -Iseconds`
- `find internal/modules/incidents -maxdepth 1 -type f -printf '%f\n' | sort`
- `rg --files internal/modules/incidents`
- Targeted `rg` searches for store constructors, `CreateIncident` call sites,
  forbidden imports, reporting incident DTO imports, and updated symbols.
- Targeted `sed` reads of tracker, reporting tests, phase maps, and changed
  source files.
- `gofmt -w` over changed Go files.
- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260702T194224Z-p1103963`.
- `make json-shape-check` initially failed at
  `.cartulary/test-results/20260702T194224Z-p1103977` because phase schedule
  inputs were stale after map edits.
- `make phase-schedules` passed at
  `.cartulary/test-results/20260702T194239Z-p1104844`.
- `make json-shape-check` then passed at
  `.cartulary/test-results/20260702T194251Z-p1105065`.
- `make phase-ledger-drift` initially failed at
  `.cartulary/test-results/20260702T194251Z-p1105061` because phase2 coverage
  ledger output needed regeneration.
- `make phase-ledgers` passed at
  `.cartulary/test-results/20260702T194258Z-p1105894`.
- `make phase-ledger-drift` passed at
  `.cartulary/test-results/20260702T194305Z-p1106248`.
- `make phase-schedule-drift` passed at
  `.cartulary/test-results/20260702T194305Z-p1106250`.
- `make phase-slice PHASE=phase2` passed at
  `.cartulary/test-results/20260702T194313Z-p1106763`.
- `make service-backed-slice PHASE=phase2` passed at
  `.cartulary/test-results/20260702T194403Z-p1132009`.
- `make phase-slice PHASE=phase6` passed at
  `.cartulary/test-results/20260702T194443Z-p1147886`.
- `make service-backed-slice PHASE=phase6` passed at
  `.cartulary/test-results/20260702T194537Z-p1165246`.
- `make phase-slice PHASE=phase8` passed at
  `.cartulary/test-results/20260702T194630Z-p1180376`.
- `make phase-slice PHASE=phase11` passed at
  `.cartulary/test-results/20260702T194711Z-p1196627`.
- `make service-backed-slice PHASE=phase11` passed at
  `.cartulary/test-results/20260702T194752Z-p1211403`.
- `make generate-drift` passed at
  `.cartulary/test-results/20260702T194830Z-p1223219`.
- `make test-fast` passed at
  `.cartulary/test-results/20260702T194840Z-p1224208`.
- `make check` initially failed at
  `.cartulary/test-results/20260702T195057Z-p1275701` due
  `internal/modules/incidents/store.go:136:6: func newStoreWithHooks is unused
  (U1000)`; the obsolete helper was removed.
- `make lint` passed at
  `.cartulary/test-results/20260702T195201Z-p1313780`.
- `make check` passed at
  `.cartulary/test-results/20260702T195250Z-p1322957`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T195250Z-p1322957`
  passed at `.cartulary/test-results/20260702T195542Z-p1407183` and refreshed
  retained-run baseline/generated support files.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T195250Z-p1322957`
  passed again at `.cartulary/test-results/20260702T195940Z-p1410879` with
  generated files unchanged.
- `make format` passed at
  `.cartulary/test-results/20260702T201506Z-p1418990`.
- `make phase-slice PHASE=phase2` passed at
  `.cartulary/test-results/20260702T201521Z-p1420768`.
- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260702T201659Z-p1445994`.
- `make json-shape-check` passed at
  `.cartulary/test-results/20260702T201704Z-p1446179`.
- `make phase-ledger-drift` passed at
  `.cartulary/test-results/20260702T201708Z-p1446507`.
- `make phase-schedule-drift` passed at
  `.cartulary/test-results/20260702T201714Z-p1446844`.
- `make service-backed-slice PHASE=phase2` passed at
  `.cartulary/test-results/20260702T201719Z-p1447022`.
- `make phase-slice PHASE=phase6` passed at
  `.cartulary/test-results/20260702T201759Z-p1462871`.
- `make service-backed-slice PHASE=phase6` passed at
  `.cartulary/test-results/20260702T201852Z-p1479950`.
- `make phase-slice PHASE=phase8` passed at
  `.cartulary/test-results/20260702T201944Z-p1495061`.
- `make phase-slice PHASE=phase11` passed at
  `.cartulary/test-results/20260702T202028Z-p1511545`.
- `make service-backed-slice PHASE=phase11` passed at
  `.cartulary/test-results/20260702T202106Z-p1526514`.
- `make generate-drift` passed at
  `.cartulary/test-results/20260702T202139Z-p1538321`.
- `make test-fast` passed at
  `.cartulary/test-results/20260702T202145Z-p1539300`.
- `make check` passed at
  `.cartulary/test-results/20260702T202405Z-p1590072`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260702T202405Z-p1590072`
  passed at `.cartulary/test-results/20260702T202658Z-p1673724` and refreshed
  finalizer-managed scheduler and duration baseline support files.

Tracker status:

- Updated: `docs/handoffs/incidents-module-refactor-tracker-2.md`
- Production behavior change expected: none.
- Public route, WebSocket wire, OpenAPI envelope, and database migration
  changes expected: none.
- Generated files hand-edited: none.
- Downstream generated/support files updated through Make/finalizer:
  `docs/testing/phase2_coverage_ledger.md`,
  `docs/testing/phase11_coverage_ledger.md`,
  `tools/execution_topology_render_index.json`, `tools/scheduler_manifest.json`,
  and retained-run duration baseline JSON files. The latest finalizer run
  refreshed `tools/browser_e2e_duration_baselines.json`,
  `tools/execution_topology_render_index.json`,
  `tools/go_test_duration_baselines.json`,
  `tools/harness_smoke_duration_baselines.json`,
  `tools/scheduler_manifest.json`, and
  `tools/service_backed_make_target_duration_baselines.json`.
- Remaining blocker: none known.
