# Entities Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Target directory: `internal/modules/entities`.

This tracker is a planning and handoff artifact only. It records the current
repository state, contract risks, refactor sequencing, and verification posture
for a later authorized implementation task. This document does not authorize
production refactors, generated-file edits, dependency edits, migrations, or
runtime behavior changes.

Allowed changes for this artifact:

- Create this Markdown tracker under `docs/handoffs/`.
- Record live repository findings from `internal/modules/entities` and narrow
  caller/contract searches.
- Preserve observable behavior as the default for every later slice.

Non-goals:

- Do not assume `entities` is a valid permanent module boundary merely because
  `internal/modules/entities` exists.
- Do not treat the framework, prior plans, phase maps, or test rows as runtime
  architecture.
- Do not hand-edit generated roots, generated harness/topology outputs,
  `go.sum`, `pnpm-lock.yaml`, or tool-managed artifacts.
- Do not perform production code movement in this tracker task.

Authority order used here:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md`, `docs/testing-harness-nlspec.md`, `docs/design.md`, and
   implementation-support guides for terminology, package boundaries, harness
   mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans and framework files as evidence, not authority.

Source posture:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md` is used as
  the planning template and doctrine, not proof of current repository state.
- `docs/domain.md` is used for vocabulary distinctions such as workbook surface
  versus source state, projection versus source state, entity mention versus
  stub entity, indicator observation versus indicator, and view schema ID versus
  label.
- `docs/testing-harness-nlspec.md` owns command invocation, target selection,
  artifact emission, generated-artifact rules, and harness mechanics.
- `docs/handoffs/entities-module-refactor-tracker.md` is prior evidence only.

Architectural finding:

`internal/modules/entities` is currently a mixture. It is not a proven permanent
module boundary and should not be treated as one by default. The package
legitimately owns host/identity entity behavior, exact-match and reusable
identifier semantics, entity mention lifecycle policy, and explicit host/identity
merge policy. It also coordinates workbook-facing create, query, patch, and
clipboard-paste behavior; revision/change-set writes; projection refreshes;
links, assessments, and timeline invalidations; import facades; and
collaboration record-change publication. Those coordinated behaviors should stay
behavior-preserving until a later slice moves them behind clearer owner ports or
module facades.

No `BLOCKED: owner contradiction` item was found during this planning pass.
Repo/framework mismatches are recorded as planning findings and adapted to the
live repository instead of inventing missing behavior.

## 2. Current-State Repository Inventory

| path | current responsibility | exported/public symbols or package surface | inbound callers | outbound dependencies | tests touching it | generated artifacts or contracts touched | suspected target owner module | risk level | notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/entities/.gitkeep` | Placeholder file. | None. | None. | None. | None. | None. | None. | low | Keep only while the directory exists. |
| `internal/modules/entities/api.go` | Host and identity view schema IDs, create request decoding, row builders, mutation payload helpers, collection action helpers. | `HostsViewSchemaID`, `IdentitiesViewSchemaID`, `CreateRequest`, `HostRecord`, `IdentityRecord`, `ReusableIdentifier`, `MutationResult`, `CollectionAction`, `DecodeCreateRequest`, `CreateRequestHash`, `BuildHostRow`, `BuildIdentityRow`, `BuildMutationPayload`. | `store.go`, `workbook`, `imports`, tests. | `fieldnorm`, `httpapi`, `viewschema`, `uuid`. | Phase 4 tests, OpenAPI contract test, workbook/import tests. | Host/identity view schema and OpenAPI/protocol shape indirectly. | `entities/host-identity`. | medium | Legitimate contract adapter for host/identity rows. |
| `internal/modules/entities/boundary_guard_test.go` | Production import boundary and route registration guard. | Test package surface only. | `make backend-unit`, phase slices. | Go parser and filesystem reads. | Itself. | Harness evidence only. | Test/harness. | medium | Guards that entities does not register workbook row-create routes, build clipboard plans, or publish directly through `platform/ws`. |
| `internal/modules/entities/clipboard_paste_store.go` | Executes prebuilt host/identity clipboard paste plans, idempotency, row creation/upsert, revisions, change sets, and projection refresh. | `ClipboardPasteResult`, `ClipboardPasteRowResult`, `ApplyClipboardPastePlan`, `EntityClipboardPasteRequestHash`. | `workbook/routes.go` through host/identity clipboard route dispatch; tests. | `authn`, `fieldnorm`, `incidents`, `imports/tabularingest`, owner ports, raw SQL. | Phase 9 clipboard tests and Phase 4 support tests. | Workbook clipboard-paste envelopes and row response contracts. | `entities` executor behind workbook/import facade. | high | Planning/building of paste plans belongs outside entities; execution now calls revisions/projections through local ports. |
| `internal/modules/entities/entitycontract/contract.go` | Shared host/identity entity constants for view schema IDs and entity type names. | `HostsViewSchemaID`, `IdentitiesViewSchemaID`, `EntityTypeHost`, `EntityTypeIdentity`. | `entities/api.go`, mention package and future callers needing constants without root `entities`. | None beyond Go standard language constants. | Boundary/unit tests through importing packages. | Host/identity view-schema constants only; no generated edits. | `entities/entitycontract`. | low | Narrow contract package added to avoid importing root `entities` for constants. |
| `internal/modules/entities/import_create.go` | Host/identity import-owner create facade consuming tabular ingest owner requests and finalizing rows through imports owner facade. | `ImportCreateCommand`, `CreateImportRowTx`. | `imports/owner_apply.go`; tests. | `authn`, `imports/ownerfacade`, `imports/tabularingest`, `uuid`. | Import apply tests and phase evidence. | Host/identity import target behavior; public import routes unchanged. | `entities` host/identity import facade. | high | Keeps import route/job ownership in imports while source create behavior lives with entities. |
| `internal/modules/entities/match.go` | Host/identity exact-match, alias, preserved identifier, conflict, and reusable identifier behavior. | `ExactMatchConflictError`; package-private matching helpers. | `store.go`, `clipboard_paste_store.go`, `import_create.go`, `merge_store.go`, tests. | `authn`, `fieldnorm`, owner record port, raw SQL. | Phase 4 exact-match, create, import, and merge tests. | Host/identity entity-origin behavior. | `entities/host-identity`. | medium | Legitimate entity-domain logic. |
| `internal/modules/entities/mentions/mention_api.go` | Entity mention action DTOs, request decoding, request hash, response payload shape, and API error helpers. | `MentionActionRequest`, `MentionActionResult`, `MentionEntityInvalidation`, `DecodeMentionActionRequest`, `MentionActionRequestHash`, mention API error helpers. | Root `entities/routes.go`, `workbook/routes.go`, tests. | `authn`, `httpapi`, `uuid`. | Phase 4 mention route/unit tests. | Entity mention resolve route and OpenAPI shape. | `entities/mentions` contract adapter. | medium | Public route name remains `/resolve` even though actions include resolve, dismiss, and restore. |
| `internal/modules/entities/mentions/mention_lifecycle.go` | Mention resolve/dismiss/restore lifecycle, transition policy, source/target validation, timeline source effects, link effects, revisions, projections, and invalidations through mention-local ports. | Mention errors, `MentionTransitionError`, `MentionRowVersionConflictError`, `MentionTargetValidationError`, `GetMentionActionAccess`, `ApplyMentionAction`, `ApplyMentionLifecycleTx`. | Root `entities/routes.go`, workbook route error mapping, tests. | `authn`, `incidents`, mention-local owner ports, raw `entity_mentions` SQL. | Phase 4 unit, integration, support, and browser evidence. | Mention route, timeline row effects, revisions, projections, WebSocket invalidations. | `entities/mentions` plus timeline/link/revision/projection ports. | high | Timeline no longer imports root `entities`; side effects remain behind narrowed adapters. |
| `internal/modules/entities/mentions/mention_resolution.go` | Tx-scoped mention resolution facade used by timeline flows. | `ErrInvalidMentionResolution`, `ResolveExistingFromMentionTx`. | `timeline/ports.go`, `timeline/mentions_collections_store.go`, tests. | Mention lifecycle helpers, `records`, raw SQL. | Phase 4 timeline/entity-origin tests. | Timeline mention resolution and source write-back behavior. | `entities/mentions` facade with timeline caller contract. | high | API name now reflects validation of an existing resolved record instead of implying implicit creation. |
| `internal/modules/entities/mentions/ports.go` | Mention-local owner-port interfaces and adapters for records, revisions, links, projections, and timeline mention effects. | Package-private ports and adapters plus `NewStore`. | Mention lifecycle/resolution files and timeline adapter construction. | `links`, `projections/adapters`, `records`, `revisions`, `timeline/mentioneffects`, `postgres`. | Boundary/unit tests through mention callers. | Harness evidence only. | `entities/mentions` adapter boundary. | high | Concentrates mention side-effect coupling away from root host/identity store. |
| `internal/modules/entities/merge_api.go` | Merge request/response DTOs, request hash, payload builder, API error mapping. | `MergeRequest`, `MergeExactMatchClassSummary`, `MergeSummary`, `MergeResult`, `MergeTimelineInvalidation`, `DecodeMergeRequest`, `MergeRequestHash`, `BuildMergePayload`. | `routes.go`, tests. | `authn`, `httpapi`, `uuid`. | Phase 4 merge tests, OpenAPI contract test. | Merge route request/response and generated OpenAPI/protocol shape. | `entities/merge` contract adapter. | medium | DTO ownership can remain while internals are split by owner ports. |
| `internal/modules/entities/merge_protected_set_test.go` | Characterizes merge protected-set coverage and fail-closed assessment repoint behavior. | Test package surface only. | `make backend-unit`, phase slices. | Phase test harness and DB helpers. | Itself. | Harness evidence only. | Test/harness for merge. | medium | Evidence that merge must protect records it directly mutates. |
| `internal/modules/entities/merge_store.go` | Explicit host/identity merge policy, survivor/loser validation, exact-match carry-forward, entity mention repointing, link/tag/assessment/revision/projection/timeline invalidation coordination. | `ErrMergeTargetNotFound`, merge error types, `GetMergeRouteIncident`, `MergeEntity`. | `routes.go`, tests. | `authn`, `fieldnorm`, `incidents`, owner ports, raw host/identity/mention SQL. | Phase 4 merge tests and merge protected-set tests. | Merge route, revisions, projections, WebSocket invalidations. | `entities/merge` plus owner ports to links, assessments, timeline, revisions, projections. | high | Legitimate merge policy is local; broad transaction coordination is the highest-risk boundary area. |
| `internal/modules/entities/openapi_contract_test.go` | Generated OpenAPI shape assertions for entity-owned merge and mention resolve contracts. | Test package surface only. | `make backend-unit`, phase slices. | `internal/gen/contracts`. | Itself. | Reads generated OpenAPI; no generated edit. | Entity route contract evidence. | medium | Workbook-owned assertions were moved to workbook OpenAPI tests. |
| `internal/modules/entities/patch_store.go` | Host/identity scalar patch executor with row-version checks, idempotency, revisions, projection refresh, and row payloads. | `ErrNoEffectivePatchChange`, `PatchRequest`, `PatchChange`, `PatchMutationResult`, `RowVersionConflictError`, `PatchEntityRow`. | `workbook/mutation_store.go`, tests. | `authn`, `incidents`, `records`, `revisions`, `viewschema`, raw SQL. | Phase 4 support tests and workbook mutation tests. | Workbook patch route behavior and revision/projection contracts. | `entities` facade plus `revisions` and `projections` ports. | high | Direct revision/projection coupling should be narrowed before code movement. |
| `internal/modules/entities/phase4_integration_test.go` | Phase 4 integration evidence for resolve route, entity-origin upsert, auth/CSRF ordering, merge route, idempotency, and indicator route compatibility. | Test package surface only. | Make phase/backend targets. | HTTP and DB harnesses, generated fixture data. | Itself. | Phase evidence accounting. | Test/harness evidence. | high | Indicator assertions here are compatibility evidence, not indicator ownership proof. |
| `internal/modules/entities/phase4_support_integration_test.go` | Support evidence for envelopes, CSRF, replay conflict, authorization re-derivation, default query meta, projection/WebSocket consequences, and record envelopes. | Test package surface only. | Make phase/backend targets. | HTTP and DB harnesses. | Itself. | Phase support evidence accounting. | Test/harness evidence. | medium | Important characterization source for behavior freeze. |
| `internal/modules/entities/phase4_unit_test.go` | Unit evidence for mention lifecycle, exact-match precedence, explicit merge, and indicator observation separation. | Test package surface only. | Make phase/backend targets. | Test harnesses and entities package. | Itself. | Phase evidence accounting. | Test/harness evidence. | medium | Indicator separation test confirms indicators are conceptually separate from entities. |
| `internal/modules/entities/ports.go` | Local owner-port interfaces and adapters for records, revisions, links, projections, assessments, and timeline mention effects used by host/identity and merge flows. | Package-private ports and adapters. | `store.go`, `patch_store.go`, `clipboard_paste_store.go`, `import_create.go`, `merge_store.go`. | `assessments`, `links`, `projections/adapters`, `records`, `revisions`, `timeline/mentioneffects`. | Boundary guard and backend unit/store tests. | Harness evidence only. | `entities` adapter boundary. | high | Concentrates sibling imports; this is acceptable only while contract boundaries stay explicit. |
| `internal/modules/entities/projectionprovider/provider.go` | Host/identity grid projection rebuild and row-refresh provider. | `RebuildIncidentHostsTx`, `RebuildIncidentIdentitiesTx`, `RefreshHostTx`, `RefreshIdentityTx`. | `internal/modules/projections/entity_grids.go`, projection provider registry. | Raw SQL over host/identity, records, links, evidence, object blobs, projection tables. | Phase 4 projection tests and registry evidence. | `contracts/projection-providers/index.json`, generated projection provider registry, host/identity view contracts. | Source-owned provider under `entities`, projection storage under `projections`. | medium | Keep unless projection provider architecture changes; row refresh is now advertised in the provider manifest. |
| `internal/modules/entities/routes.go` | HTTP route registration and handlers for explicit merge and entity mention resolve actions; delegates mention policy to `entities/mentions`; publishes record changes through collaboration publisher. | `Service`, `RegisterRoutes`. | `internal/app/runtime.go`, HTTP clients, tests. | `collaboration`, `authn`, `httpapi`, `httpauth`, `incidents`, `entities/mentions`, `uuid`. | Phase 4 route/support tests and browser evidence. | Route inventory, OpenAPI, WebSocket event behavior. | Transport adapter over entity merge/mention policy. | medium | Transport-adjacent, but acceptable as module route adapter. |
| `internal/modules/entities/store.go` | Store constructor; host/identity row query; host/identity create/upsert; idempotency; revisions/change sets; projection refresh through ports; row hydration. | `ErrInvalidCreateRequest`, `Store`, `NewStore`, `QueryHostRows`, `QueryIdentityRows`, `CreateHostRow`, `CreateIdentityRow`. | `workbook/store.go`, `workbook/routes.go`, import owner apply, tests. | `authn`, `incidents`, `postgres`, `viewschema`, owner ports, raw SQL. | Phase 4 create/query/idempotency tests, workbook/import tests. | Host/identity workbook rows, projections, revisions, OpenAPI behavior indirectly. | `entities/host-identity` facade plus revision/projection ports. | high | Main remaining catch-all pressure point, but direct revision/projection construction has been narrowed. |

## 3. Workbook Boundary Diagnosis

Current diagnosis: `internal/modules/entities` is a mixture. It is not a
workbook runtime owner, and the live repository already keeps generic workbook
routes in `internal/modules/workbook`. It is a legitimate host/identity service
facade for built-in Hosts and Identities sheets, but it also contains
cross-owner mutation coordination that should be split or wrapped by clearer
owner ports before later movement.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Host/identity entity-origin create and upsert | `api.go`, `store.go`, `match.go` | `entities/host-identity` | keep | Core 01/02 host and identity entity-origin behavior; workbook/import callers. | Legitimate entity behavior. |
| Host/identity row query and row shape | `store.go`, `api.go` | `entities` query facade with view schema contract | keep for now | Workbook store delegates host/identity queries to `entityStore`. | Do not move without query parity tests. |
| Host/identity scalar patch | `patch_store.go`, workbook mutation dispatcher | `entities` facade plus `revisions` and `projections` ports | split | Workbook `PATCH /api/v1/records/{record_id}` delegates host/identity patches. | Preserve row-version conflicts and idempotency. |
| Clipboard paste execution | `clipboard_paste_store.go`, workbook clipboard route | Workbook planner plus `entities` executor | split | Entities consumes `tabularingest.BatchPlan`; boundary guard prevents plan building in entities. | Keep parser/planner out of entities. |
| Generic workbook route registration | `workbook/routes.go` | `workbook` | keep moved | Routes for query, clipboard, bulk, create, patch are owned by workbook. | Entities must not register generic workbook row-create routes. |
| Entity mention action policy | `mentions/mention_api.go`, `mentions/mention_lifecycle.go`, `mentions/mention_resolution.go` | `entities/mentions` plus timeline/link/revision/projection ports | split done | Entity mention resolve route delegates to `entities/mentions`; timeline imports the narrowed mention package. | Policy moved behind a narrower boundary while the public route remains stable. |
| Timeline source row updates from mention actions | `mentions/ports.go` through `timeline/mentioneffects` adapter | `timeline` source owner | split done | Timeline imports `entities/mentions`; mention ports call timeline mention effects. | Root `entities` no longer couples timeline to host/identity, merge, or import internals. |
| Explicit host/identity merge policy | `merge_api.go`, `merge_store.go` | `entities/merge` | keep policy, split effects | Merge route is entity-specific and uses survivor/loser semantics. | Link/tag/assessment/timeline/revision effects should remain owner-owned. |
| Entity mention repoint during merge | `merge_store.go` | `entities/mentions` or `entities/merge` | defer | Merge directly updates `entity_mentions`. | Acceptable while mentions remain in entities. |
| Link/tag repoint during merge | `merge_store.go`, `ports.go` | `links` | split | Merge coordinates through link ports. | Owner-port contract should be tightened before moving. |
| Assessment repoint/protected set during merge | `merge_store.go`, `ports.go` | `assessments` | split | Merge protected-set tests cover assessment subject records. | Preserve fail-closed lock behavior. |
| Projection refresh/rebuild | `store.go`, `patch_store.go`, `clipboard_paste_store.go`, `import_create.go`, `merge_store.go`, `projectionprovider/provider.go` | Source provider in `entities`; storage orchestration in `projections` | split done for row refresh | Projection provider registry identifies source owner and projection storage owner separately. | Host/identity mutation flows call `RefreshEntityRowTx`; provider stays source-owned. |
| Collaboration WebSocket publication | `routes.go`, workbook routes, `collaboration.RecordChangePublisher` | `collaboration` transport/event publisher | keep as adapter | Boundary guard rejects direct `platform/ws` route import. | Preserve `record_changed` event semantics. |
| Import target create facades | `imports/targets.go`, `imports/owner_apply.go`, `entities/import_create.go` | `imports` route owner plus owner create facades | split done | Host and identity import targets use `entities.host.import_create` and `entities.identity.import_create`. | Route and job state stay in imports; source create semantics live in entities. |
| Party exact-match test evidence | `internal/modules/parties/phase9_parties_test.go` | `parties` | moved | Test now lives under parties and still exercises party behavior through workbook/import entrypoints. | Evidence location is no longer an entities ownership signal. |
| Broad OpenAPI mutation assertions | `entities/openapi_contract_test.go`, `workbook/openapi_contract_test.go` | Route owner evidence | split done | Generated OpenAPI assertions are split by route owner. | Entities keeps mention/merge assertions; workbook owns workbook mutation assertions. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Existing behavior to freeze | Existing tests or evidence | Required characterization tests |
| --- | --- | --- | --- |
| HTTP merge route | `POST /api/v1/records/{survivor_record_id}/merge`; reviewer/admin gate; survivor/loser row versions; idempotency; merge payload; loser/survivor projection changes. | `phase4_integration_test.go`, `merge_protected_set_test.go`, `openapi_contract_test.go`. | Add or confirm concurrency/protected-set characterization before moving merge internals. |
| HTTP entity mention route | `POST /api/v1/entity-mentions/{entity_mention_id}/resolve`; resolve/dismiss/restore action payload; row-version conflicts; target validation; source row response. | `phase4_unit_test.go`, `phase4_integration_test.go`, support tests. | Add byte-level response/invalidations characterization before splitting lifecycle ports. |
| Workbook query | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` for host/identity view schemas; query meta, paging, row shape, field keys. | Workbook tests, Phase 4 support tests, view schema contracts. | Add host/identity query parity tests before moving query facade. |
| Workbook clipboard paste | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste` for host/identity; batch order, conflicts, idempotency, row payloads. | Phase 9 clipboard tests and support tests. | Add focused invalid target and partial-failure characterization before splitting executor. |
| Workbook bulk mutation | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations`; timeline route currently maps entity mention errors/conflicts. | Workbook route tests and Phase 4 entity conflict tests. | Confirm no host/identity bulk mutation behavior is hidden in entities before moving mention errors. |
| Workbook row create | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` delegates host/identity creates to entities; preserves envelopes, auth order, idempotency, and WebSocket changes. | Phase 4 create/idempotency/auth tests, OpenAPI contract test. | Add owner-dispatch tests if create facade is narrowed. |
| Workbook patch | `PATCH /api/v1/records/{record_id}` delegates host/identity patches to entities; preserves row-version conflict and revision/projection behavior. | Phase 4 support tests, workbook mutation tests. | Add host/identity patch parity tests before extracting revision/projection ports. |
| Imports collection route | `/api/v1/import-sessions`; extension-profile gated import session creation. | Imports route tests and target registry. | No entity-specific test unless import facade changes. |
| Imports apply route | `/api/v1/import-sessions/{session_id}/apply`; host/identity import rows call entity create facade with `entities.host.import_create` and `entities.identity.import_create`. | Imports route/store tests and target registry. | Add import-facade regression before moving create API. |
| WebSocket path and event | `GET /ws/v1/incidents/{incident_id}` emits `record_changed` with `record_id`, `row_version`, `change_set_id`, `client_txn_id`, `actor_user_id`, `changed_field_keys`, and `affected_views`. | `phase4_support_integration_test.go`, collaboration publisher tests, platform WS tests. | Add route-specific event tests if route handlers are thinned. |
| Saved-view/view-schema behavior | Host and identity view schema IDs remain `cartulary.view.hosts.v1` and `cartulary.view.identities.v1`; saved views key by immutable view schema ID, not label. | View schema JSON contracts, workbook tests, UI contract tests. | Run drift checks after authored view-schema changes. |
| Projection refresh | Host/identity projection tables and provider rebuilds remain compatible with query rows, linked event counts, and attached evidence counts. | Phase 4 projection tests, projection provider registry. | Add projection parity tests before moving refresh calls. |
| Authorization checks | Merge requires reviewer/admin; mention action requires editor/reviewer/admin; workbook mutation routes require editor/reviewer/admin; query requires membership. | Route tests, support auth re-derivation tests, Core 04. | Confirm exact auth-ordering before moving route code. |
| Revision/change-set behavior | Entity create, paste, patch, mention, and merge flows preserve change sets, mutations, record revisions, row versions, route keys, and replay behavior. | Phase 4/9 tests and revision assertions. | Add mutation payload parity tests before port extraction. |
| Generated protocol/view contracts | OpenAPI, view schema JSON, generated Go/TS contracts, UI contracts remain downstream of owner inputs. | `openapi_contract_test.go`, generated artifact policy, `make generate-drift`. | Never hand-edit generated roots; run drift after authored inputs change. |
| Harness/test accounting | Phase maps and test rows are evidence accounting only, not runtime architecture. | `make task-guide`, phase slice outputs, testing harness NLSpec. | Regenerate ledgers only through Make-owned generators if test ownership paths move. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Characterization must precede code movement. | Routes, envelopes, revisions, projections, WebSocket events, and idempotency are affected by several files in this package. | high | must_fix | Current route/domain owners | Add or confirm characterization before any implementation slice moves behavior. |
| `store.go` remains a high-pressure facade. | It owns query/create/upsert, idempotency, row hydration, and command orchestration while calling revisions/projections through ports. | high | should_fix | `entities` plus `revisions` and `projections` ports | Keep query behavior stable; future work can narrow host/identity command/query packages if it reduces coupling. |
| `patch_store.go` coordinates patch behavior with revisions and projections through ports. | It still owns host/identity patch semantics and imports revision error values for deleted-record handling. | medium | should_fix | `entities` patch facade, `revisions`, `projections` | Preserve parity tests; consider a narrower revision error contract only if reuse grows. |
| Clipboard execution depends on `imports/tabularingest` plan shape. | `ApplyClipboardPastePlan` consumes `tabularingest.BatchPlan`. | medium | intentional/no_action | `imports/tabularingest` planning contract and `entities` executor | Keep planning out of entities; stable plan dependency is acceptable. |
| Merge coordinates several owners in one transaction. | `merge_store.go` coordinates links, tags, assessments, revisions, projections, timeline invalidations, and entity mentions. | high | should_fix | `entities/merge` plus owner ports | Tighten owner-port contracts and protected-set rules before movement. |
| Mention lifecycle coordinates timeline, links, revisions, projections, and invalidations. | `mentions/mention_lifecycle.go` calls mention-local owner ports and updates `entity_mentions`. | medium | should_fix | `entities/mentions` plus timeline/link/revision/projection owners | Initial boundary split is complete; future work should only split more if side-effect ports become too broad. |
| Timeline mention flows depend on the narrowed mention package. | `timeline/ports.go` and `timeline/mentions_collections_store.go` import `entities/mentions`, not root `entities`. | low | intentional/no_action | Timeline caller facade plus entity mention contract package | Current dependency is explicit and avoids pulling host/identity or merge internals into timeline. |
| Broad OpenAPI test ownership was corrected. | Workbook-owned assertions now live in `workbook/openapi_contract_test.go`; entity-owned mention/merge assertions remain under entities. | low | intentional/no_action | Route owner tests | Keep generated OpenAPI output untouched; update tests only when authored contracts change. |
| Party behavior test ownership was corrected. | Phase 9 party evidence now lives in `internal/modules/parties/phase9_parties_test.go`; phase map input points to parties. | low | intentional/no_action | `parties` | Phase ledger/schedule drift validation must confirm regenerated accounting. |
| Entity routes are transport-adjacent. | `routes.go` authenticates, authorizes, maps errors, slides sessions, and publishes record changes. | medium | intentional/no_action | Module route adapter plus `collaboration` publisher | Acceptable as route adapter; do not move into domain/store. |
| Direct SQL exists in domain-facing code. | Store, patch, merge, mention, and projection provider files contain raw SQL. | high | should_fix | Owning source modules and storage adapters | Move cross-owner SQL first; local host/identity SQL can remain until a repository-wide storage abstraction exists. |
| Generated contracts are downstream artifacts. | Generated roots are declared by policy and harness NLSpec. | high | intentional/no_action | Authored specs/contracts only | Use Make generation/drift targets; no hand edits. |
| Phase maps and target rows are evidence accounting only. | Framework and testing harness distinguish runtime architecture from evidence. | medium | intentional/no_action | Harness owners | Do not use phase maps to justify module ownership. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 through WF-08 | Create this tracker, record authority order, and confirm planning-only scope. | This file, framework, domain doc, harness NLSpec. | `make lint-markdown` if handoff docs are linted by config; otherwise record scope. | Tracker path exists and states no production refactor. |
| WF-01 | Workbook package inventory | chain | WF-00 | WF-02, WF-04 | Inventory every file in `internal/modules/entities` and separate workbook-owned routes from entity facades. | All target files, `internal/modules/workbook/routes.go`, `workbook/store.go`, `workbook/mutation_store.go`. | Static inspection plus `make backend-unit` before implementation. | Section 2 inventory complete. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map each observable route, WebSocket event, view schema, projection, import facade, and generated contract to the correct owner. | Entities, workbook, imports, collaboration, projections, contracts. | `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift` when authored inputs change. | Section 4 freeze map complete. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-06 | Identify tests required before any code movement. | Phase 4 tests, Phase 9 tests, workbook/import/collaboration tests. | `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9`. | Required characterization rows are explicit. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05, WF-06 | Classify imports, direct SQL, cross-owner effects, authorization placement, and misplaced tests. | `ports.go`, `store.go`, `patch_store.go`, `merge_store.go`, `mention_lifecycle.go`, boundary tests. | `make backend-unit`, `make lint`. | Findings table has classification and owner. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Decide behavior-preserving target facades and owner ports without implementation patches. | Same as WF-04 plus workbook/import/projection callers. | Design review plus focused unit characterization. | Facade contracts and TODO owner decisions recorded. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Define smallest safe behavior-preserving implementation slices. | Future touched files depend on selected slice. | Per-slice validation from Section 7. | Slice table complete with rollback notes. |
| WF-07 | Harness/test/accounting update plan | parallel/chain | WF-03, WF-06 | WF-08 | Plan test moves, phase accounting updates, and drift checks only when paths/contracts change. | Test files, phase maps, generated ledgers, Make targets. | `make task-guide`, phase slices, drift targets. | No runtime ownership inferred from accounting. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Close the tracker and leave another agent able to continue. | This file and final verification artifacts. | Narrow docs check now; broader targets during implementation. | Section 12 completion criteria evaluated. |

## 7. Proposed Refactor Slice Plan

| Slice | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 Tracker-only artifact | None | Create this planning document with live inventory, findings, workflows, slices, validation plan, and handoff log. | `docs/handoffs/entities-module-refactor-tracker-2.md`. | None to runtime; docs may drift if repo changes. | Markdown formatting only. | `make lint-markdown` where in scope; otherwise record handoff path scope. | Delete or revert this file only; no runtime rollback. | File exists and all 12 required sections are populated. |
| S-01 Characterization freeze | S-00 | Add or confirm tests for route envelopes, authorization ordering, idempotency, revisions, projections, and WebSocket events before moving behavior. | Entities tests, workbook tests, imports tests, collaboration tests. | Missing behavior parity can hide regressions during refactor. | Preserve Phase 4/9 tests; add focused route/payload tests where gaps exist. | `make backend-unit`, `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9`. | Revert only added tests if they misstate current behavior. | Every movement candidate has a test owner or explicit TODO. |
| S-02 Host/identity mutation facade narrowing | S-01 | Move revision/projection/idempotency side effects behind explicit ports while keeping host/identity create, patch, paste, query behavior unchanged. | `store.go`, `patch_store.go`, `clipboard_paste_store.go`, `ports.go`, workbook/import callers. | Route keys, replay semantics, row versions, projection refresh, and row payloads. | Host/identity create, patch, paste, import, and query parity tests. | `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9`. | Revert port extraction and keep existing direct calls. | Public route, row, revision, projection, and replay behavior is unchanged. |
| S-03 Merge owner-port split | S-01 | Tighten merge side effects through links, assessments, revisions, projections, and timeline invalidation owner ports while keeping entity merge policy local. | `merge_store.go`, `ports.go`, `links`, `assessments`, `timeline`, `revisions`, `projections`. | Survivor/loser semantics, protected-set locking, carried identifiers, invalidations, WebSocket changes. | Preserve merge protected-set and Phase 4 merge tests; add owner-port contract tests. | `make backend-unit`, `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase4`. | Revert to existing merge coordinator and ports. | All merge route and store characterization remains green. |
| S-04 Mention lifecycle boundary cleanup | S-01 | Define narrower mention lifecycle API/errors and reduce timeline-to-entities coupling without changing route or timeline behavior. | `mention_api.go`, `mention_lifecycle.go`, `mention_resolution.go`, `ports.go`, `timeline/ports.go`, `timeline/routes.go`, `timeline/mentions_collections_store.go`. | Mention transitions, source row writes, target validation, timeline projection invalidations, WebSocket events. | Preserve mention route/unit/integration tests; add API parity tests for timeline caller path. | `make backend-unit`, `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase4`. | Restore old imports/API if parity breaks. | Timeline and route behavior is unchanged with cleaner ownership. |
| S-05 Evidence re-home slice | S-01 | Move misplaced tests only if ownership and phase accounting are updated together. | `phase9_parties_test.go`, `openapi_contract_test.go`, possible phase maps and generated ledgers. | Phase accounting, import paths, broad contract coverage. | Preserve identical assertions after move. | `make backend-unit`, `make phase-slice PHASE=phase9`, `make phase-ledger-drift` if owner inputs change. | Revert test moves and phase-map edits. | Test ownership is clearer and accounting is regenerated through Make. |
| S-06 Contract/drift slice | Any implementation slice that changes authored contract inputs | Regenerate downstream contracts and ledgers from owner inputs only. | Authored contract files, manifests, generated roots via Make. | Generated OpenAPI/TS/UI contracts and projection provider registry drift. | Preserve contract tests. | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`. | Revert authored inputs and generated outputs together. | Drift targets are clean and no generated file was hand-edited. |
| S-07 Final broad verification | S-02 through S-06 as applicable | Run narrow slice validations first, then final gate by risk. | Whole repo through Make targets. | Broad regressions across workbook/import/collaboration/projections. | All relevant existing tests. | `make agent-finalize`, then `make test-fast` or `make check` by risk. | Revert the last behavior slice if failures are related. | Final handoff lists command results and artifacts. |

## 8. Validation Plan

Commands are Make-owned public targets discovered from `make help-all`,
`make task-guide ROLE=feature-dev PHASE=phase4`, and
`make task-guide ROLE=feature-dev PHASE=phase9`. Direct `go`, `pnpm`,
Vitest, Playwright, Biome, and raw scripts are developer conveniences unless a
Make-owned wrapper invokes them.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Pure backend unit evidence, including boundary guards and phase 4/9 unit rows. | yes | First executable gate for import-boundary and route-registration changes. |
| integration | `make backend-store`, `make backend-integration` | Service-backed store and HTTP integration behavior for entities, workbook, imports, projections, and collaboration effects. | yes for code movement | Required after any store, route, or owner-port implementation. |
| phase slice | `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9` | Authoritative phase rows that cover entity mentions, merge, entity-origin Host/Identity creation, clipboard/import/workbook behavior, and parties/indicators evidence. | yes for affected phases | Use `make task-guide ROLE=feature-dev PHASE=<phase>` before broad reruns. |
| service-backed | `make service-backed-slice PHASE=phase4`, `make service-backed-slice PHASE=phase9` | Service-backed phase rows, including DB/object-store/browser-backed coverage where selected. | no for docs, yes for risky runtime slices | Narrower than full `check` when services are needed. |
| e2e/browser | `make frontend-unit`, `make browser-e2e-webserver-backed` | Frontend unit and shared-stack browser behavior for workbook/inspector/user-facing flows. | no for docs, yes for route/user-facing changes | Phase guides show browser functional coverage for phase 4 and phase 9. |
| generated drift | `make generated-artifact-policy-check`, `make generate-drift`, `make json-shape-check` | Generated artifact policy, generated contract drift, JSON/manifest/bootstrap shapes. | yes when authored contract or manifest inputs change | Never hand-edit generated roots. |
| import-boundary/static | `make lint`, `make frontend-import-boundary-check` | Backend/frontend/static hygiene and frontend import boundaries. | no for docs, yes near final implementation | Backend import-boundary guards are Go tests under `make backend-unit`. |
| docs | `make lint-markdown` | Authored docs covered by markdownlint configuration. | yes for this doc if in lint scope | Current markdownlint globs do not include `docs/handoffs/**`; still useful to record if skipped or no-op. |
| full check | `make agent-finalize`, then `make test-fast` or `make check` by risk | End-of-run harness maintenance and aggregate verification. | no for docs, yes before handoff after runtime refactor | If retained run maintenance is used, pass `RESULTS_DIR=<successful full warm check run root>`. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| ENT2-001 | Create tracker file and record planning-only scope. | WF-00 | DONE | None | `docs/handoffs/entities-module-refactor-tracker-2.md`. | File exists with section 1 populated. |
| ENT2-002 | Inventory every live file under `internal/modules/entities`. | WF-01 | DONE | ENT2-001 | Section 2 inventory table. | `.gitkeep`, all Go files, and `projectionprovider/provider.go` are listed. |
| ENT2-003 | Diagnose workbook/entity boundary without assuming permanent module validity. | WF-01, WF-04 | DONE | ENT2-002 | Section 3 diagnosis table. | Architectural finding names mixture and owner candidates. |
| ENT2-004 | Freeze public contracts and behavior risks. | WF-02 | DONE | ENT2-002 | Section 4 freeze map. | HTTP, WebSocket, workbook, imports, projections, auth, revisions, generated contracts, and harness accounting are mapped. |
| ENT2-005 | Classify coupling findings. | WF-04 | DONE | ENT2-002 | Section 5 findings table. | Each finding has risk, classification, owner, and planning action. |
| ENT2-006 | Seed workstreams and dependency chain. | WF-06 | DONE | ENT2-003, ENT2-004, ENT2-005 | Section 6 workflow table. | WF-00 through WF-08 have dependencies and checkpoints. |
| ENT2-007 | Define behavior-preserving slices. | WF-06 | DONE | ENT2-006 | Section 7 slice table. | Each slice has dependency, change, risk, tests, validation, rollback, and completion criterion. |
| ENT2-008 | Discover Make-owned validation commands. | WF-07 | DONE | ENT2-004 | Section 8 validation table. | Commands are real public Make targets or marked as scoped/skipped. |
| ENT2-009 | Run production refactor. | WF-05, WF-06 | DONE | ENT2-001 through ENT2-008 | Host/identity projection/revision calls moved behind ports; entity mention boundary extracted. | Runtime code compiles and targeted Make validation passes. |
| ENT2-010 | Re-home party/OpenAPI evidence tests. | WF-07 | DONE | S-01 characterization freeze | `internal/modules/parties/phase9_parties_test.go`, `tools/phase9_test_map.json`, OpenAPI tests. | Ownership and phase accounting inputs are updated. |
| ENT2-011 | Execute runtime validation gates. | WF-08 | DONE | Later implementation slice | `make backend-unit`, `make backend-store`, `make backend-integration`, phase slices, service-backed slices, drift/static checks, `make agent-finalize`, and `make test-fast`. | Narrow and broad validation results are recorded after runtime changes. |
| ENT2-012 | Finalize this remediation handoff. | WF-08 | DONE | ENT2-001 through ENT2-011 | Section 10 handoff log and final response. | Tracker path, commands, skipped checks, and blockers are stated. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | Planning-only tracker created; no production refactor authorized or performed. | Inspected framework, domain doc, harness NLSpec, prior tracker, entities files; touched `docs/handoffs/entities-module-refactor-tracker-2.md`. | `git status`, `find`, `sed`, `rg`, `wc`, `jq`, `make help-all` filtered, `make task-guide ROLE=feature-dev PHASE=phase4`, `make task-guide ROLE=feature-dev PHASE=phase9`. | Authority posture and source hierarchy recorded. | Runtime refactor requires later explicit authorization. | Use this tracker as handoff input for the next implementation task. |
| 2026-07-02 EDT | Codex remediation session | Runtime remediation was authorized and completed after the tracker was created. | `docs/handoffs/entities-module-refactor-tracker-2.md`, `docs/guides/cartulary-dev-guide.md`, owner code and harness inputs listed below. | `rg`, `sed`, `jq`, `gofmt`, Make validation listed in test/harness row. | Scope changed from planning-only to implemented remediation; handoff updated with resolved blockers and validation results. | None for this remediation. | Future owner-spec decisions should be tracked separately. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | Entities diagnosed as a mixture: legitimate host/identity, mention, and merge policy plus cross-owner orchestration. | `internal/modules/entities/**`, workbook, imports, timeline, projections, collaboration callers. | `find internal/modules/entities`, targeted `rg` imports/callers, targeted `sed` file reads. | Inventory and boundary diagnosis completed. | Owner-port redesign choices remain open for implementation slices. | Start with characterization freeze before moving code. |
| 2026-07-02 EDT | Codex implementation session | Host/identity source behavior remains in root `entities`; entity mention lifecycle moved behind `internal/modules/entities/mentions`; timeline imports the narrowed mention package, not root entities. | `internal/modules/entities/**`, `internal/modules/timeline/{ports.go,mentions_collections_store.go,routes.go}`, `internal/modules/workbook/routes.go`. | `rg`, `sed`, `gofmt`, `make backend-unit`. | Boundary split compiled and backend unit passed during implementation. | Full store/integration/phase validation still pending. | Finish drift/phase validation and record results. |
| 2026-07-02 EDT | Codex remediation session | Host/identity mutation orchestration uses record/revision/projection ports; import apply calls the entities import-create facade; mention policy lives in `entities/mentions`. | `internal/modules/entities/{store.go,patch_store.go,clipboard_paste_store.go,import_create.go,ports.go,routes.go}`, `internal/modules/entities/mentions/**`, `internal/modules/imports/{owner_apply.go,routes.go}`, timeline/workbook callers. | `gofmt`, `make backend-unit`, `make backend-store`, `make backend-integration`, `make lint`, `make test-fast`. | Runtime boundaries compile and pass store, integration, lint, and broad fast tests. | None for completed remediation. | Defer only future optional package splits or deeper owner-port narrowing. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | HTTP routes, workbook/import delegation, WebSocket `record_changed`, host/identity view schemas, and generated contract surfaces mapped. | Workbook routes, imports routes/targets, collaboration publisher/platform WS snippets, contracts/view-schemas, projection provider registry references. | Targeted `rg`, `sed`, `jq`. | Freeze map created; generated roots untouched. | Later authored contract changes require drift/generation targets. | Run drift targets only if implementation changes owner inputs. |
| 2026-07-02 EDT | Codex implementation session | Host/identity projection providers now declare row refresh support; OpenAPI evidence split by owner while route shapes remain unchanged. | `internal/modules/projections/provider_registry.go`, `internal/modules/entities/projectionprovider/provider.go`, `contracts/projection-providers/index.json`, OpenAPI tests. | `sed`, `rg`, `gofmt`, `make backend-unit`. | Code-backed registry and manifest require drift validation. | Generated roots remain untouched. | Run `make generate-drift`, `make json-shape-check`, and generated artifact policy checks. |
| 2026-07-02 EDT | Codex remediation session | Projection provider manifest and registry agree on host/identity row refresh; generated roots remain untouched; phase schedules/ledgers were regenerated through Make. | `contracts/projection-providers/index.json`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, `docs/testing/phase9_coverage_ledger.md`, `tools/go_test_duration_baselines.json`. | `make phase-schedules`, `make phase-ledgers`, `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift`, `make phase-ledger-drift`, `make phase-schedule-drift`. | Drift/static checks pass after Make-owned regeneration. | None. | Re-run drift if owner inputs change again. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | Make-owned validation surface discovered; phase 4 and phase 9 guide entries recorded. | Harness NLSpec, Make help/task-guide output, entities tests. | `rg '^func Test' internal/modules/entities`, `make help-all` filtered, `make task-guide ROLE=feature-dev PHASE=phase4`, `make task-guide ROLE=feature-dev PHASE=phase9`. | Validation plan uses real targets. | Broad runtime validation deferred because this task changes docs only. | Run `make lint-markdown`; use phase slices after runtime changes. |
| 2026-07-02 EDT | Codex implementation session | Phase 9 party evidence moved to `parties`; OpenAPI assertions split between workbook and entities. | `internal/modules/parties/phase9_parties_test.go`, `tools/phase9_test_map.json`, `internal/modules/{entities,workbook}/openapi_contract_test.go`. | `mv`, `sed`, `rg`, `gofmt`, `make backend-unit`. | `make backend-unit` passed after build fixes. | Phase ledger/schedule drift checks still pending. | Run phase 9 slice and drift checks. |
| 2026-07-02 EDT | Codex remediation session | Full targeted validation set completed; missing duration baselines caused by the moved party test were refreshed from the successful phase 9 service-backed slice. | Backend tests, phase maps/ledgers/schedules, Go test duration baselines. | `make backend-unit`, `make backend-store`, `make backend-integration`, `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9`, `make service-backed-slice PHASE=phase4`, `make service-backed-slice PHASE=phase9`, `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260702T132119Z-p4119345`, `make go-test-duration-baseline-coverage`, `make agent-finalize`, `make test-fast`. | All listed commands pass after regenerated schedule/ledger and baseline updates. | None. | Browser/front-end-only gates were not run separately; `test-fast` was the final broad gate. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | Authorization surfaces mapped: merge reviewer/admin; mention and workbook mutations editor/reviewer/admin; query membership; WebSocket auth through collaboration route. | `entities/routes.go`, `workbook/routes.go`, collaboration routes/publisher, Core 04 references from prior inspection. | Targeted `sed` and `rg`. | Auth behavior is in freeze map. | Exact auth-ordering must be characterized before route code movement. | Add/confirm route ordering tests in S-01. |
| 2026-07-02 EDT | Codex remediation session | Public auth behavior was preserved while route adapters delegate to the new mention package and import owner facade. | `internal/modules/entities/routes.go`, `internal/modules/workbook/routes.go`, `internal/modules/imports/routes.go`, phase 4/9 tests. | `make backend-integration`, `make phase-slice PHASE=phase4`, `make service-backed-slice PHASE=phase4`, `make test-fast`. | Existing route authorization and envelope evidence passes. | None. | Add new auth tests only if future route ownership changes are authorized. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02 EDT | Codex current session | Open questions are limited to implementation blockers and evidence relocation blockers. | This tracker. | Documentation creation and narrow validation. | Remaining blockers listed in section 11. | No runtime implementation authorization in this tracker. | Pick S-01 before any code move. |
| 2026-07-02 EDT | Codex remediation session | OQ-001 through OQ-005 are resolved for this remediation. | This tracker, changed code/tests/docs, Make artifacts. | Validation commands listed above. | No remaining blocker for the implemented remediation. | Future owner-spec decisions are outside this task. | Use section 11 as closed-resolution context for future planning. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| OQ-001 | What exact owner-port contract should replace direct revision/projection calls in host/identity create, patch, and paste flows? | Moving code without this could change idempotency, row versions, change-set payloads, or projection refresh timing. | Characterization tests plus owner decision across `entities`, `revisions`, and `projections`. | Resolved: host/identity flows use local record/revision/projection ports, `RefreshEntityRowTx`, source-owned row refresh providers, and an import-create facade consuming `tabularingest.ImportOwnerCreateRequest`. |
| OQ-002 | Should timeline mention caller APIs depend on an entity mention contract package, a narrowed `entities` facade, or timeline-owned adapter types? | Timeline currently imports entities for mention flows and errors; changing this affects package boundaries and transition semantics. | Core/domain authority plus caller characterization from timeline and entities tests. | Resolved: timeline imports `internal/modules/entities/mentions`; root `entities` remains route adapter for the public mention route. |
| OQ-003 | Should `phase9_parties_test.go` be re-homed under `internal/modules/parties`? | It blocks clean evidence ownership but not runtime behavior. Moving it may require phase accounting updates. | Party owner confirmation and harness/phase-map update path. | Resolved: moved to `internal/modules/parties/phase9_parties_test.go`; `tools/phase9_test_map.json` updated. |
| OQ-004 | Should `openapi_contract_test.go` remain under entities or move to a contract/route-owner test location? | It is broad contract evidence and can mislead future agents about module ownership. | Contract owner decision and unchanged OpenAPI assertions after move. | Resolved: workbook-owned assertions moved to workbook OpenAPI test; entities OpenAPI test retains mention resolve and adds merge route coverage. |
| OQ-005 | Is a package split inside `entities` desired, or should the first implementation only narrow ports in place? | A package split raises import-cycle and generated-contract risk; port narrowing is smaller. | Implementation owner preference after S-01 characterization. | Resolved: staged split implemented for `entities/mentions`; host/identity and merge remain in root `entities` behind narrowed ports. |

No open `BLOCKED: owner contradiction` item is known from the inspected owner
documents.

## 12. Binary Completion Criteria

The tracker is complete only when all of the following are true:

- Every file in `internal/modules/entities` is inventoried or explicitly out of
  scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every implementation slice is behavior-preserving unless explicitly marked as
  requiring later authorization.
- Validation commands are discovered or marked as `TODO:` with reason.
- Handoff sections are current enough for another agent to continue without
  rediscovery.

Current completion status for this artifact:

- Inventory is complete for all live files found under `internal/modules/entities`.
- Public contract risks are mapped for HTTP, WebSocket, workbook, imports,
  saved/view schema, projections, authorization, revisions, generated contracts,
  and harness accounting.
- Workflows WF-00 through WF-08 and slices S-00 through S-07 have dependencies,
  validations, rollback notes, and completion criteria.
- Production refactor was run after implementation authorization; no public route,
  WebSocket path, generated root, migration, or dependency edit was intentionally
  changed.
- Files inspected include the framework, domain doc, harness NLSpec, prior
  entities tracker, target directory files, and relevant workbook/imports/
  timeline/projections/collaboration callers.
- Commands run include `git status`, `find`, `sed`, `rg`, `wc`, `jq`, filtered
  `make help-all`, `make task-guide ROLE=feature-dev PHASE=phase4`,
  `make task-guide ROLE=feature-dev PHASE=phase9`, `gofmt`, `make backend-unit`,
  `make backend-store`, `make backend-integration`,
  `make phase-slice PHASE=phase4`, `make phase-slice PHASE=phase9`,
  `make service-backed-slice PHASE=phase4`,
  `make service-backed-slice PHASE=phase9`, `make phase-schedules`,
  `make phase-ledgers`, `make generated-artifact-policy-check`,
  `make json-shape-check`, `make phase-ledger-drift`,
  `make phase-schedule-drift`, `make generate-drift`, `make lint-markdown`,
  `make lint`, `make go-test-duration-baselines`,
  `make go-test-duration-baseline-coverage`, `make agent-finalize`, and
  `make test-fast`.
- Tracker status: updated for the remediation implementation.
- Exact output path:
  `docs/handoffs/entities-module-refactor-tracker-2.md`.
- Remaining blockers: none for this remediation. Future owner-spec decisions
  outside this task remain intentionally deferred.
