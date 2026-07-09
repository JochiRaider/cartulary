# links Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `internal/modules/links`
- Target label: `links`
- Normalized target label: `links`
- Output path: `docs/handoffs/links-module-refactor-tracker.md`
- Status: planning and documentation only.
- Allowed changes for this task: create or update this tracker file only.
- Non-goals: no production refactor, no tests, no contracts, no generated artifacts, no package configuration, no migrations, and no harness files.
- Implementation status: implementation requires a later explicitly authorized task.
- Current target existence: `internal/modules/links` exists.

Source hierarchy used:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. Domain vocabulary and implementation-support guides for terminology, package boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans, handoffs, and framework files as evidence only.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/testing-harness-nlspec.md`
- `docs/graph_projection_nlspec.md` was discovered as adopted/current for graph projection only; it does not own workbook-grid projection behavior.

No links-specific adopted subsystem NLSpec was found. The current owner posture is therefore Core-owned links/tags behavior plus live repository evidence.

Repository files inspected:

- All files under `internal/modules/links`.
- Direct callers and adapters under `internal/modules/workbook`, `internal/modules/timeline`, `internal/modules/tasksdecisions`, `internal/modules/entities`, `internal/modules/artifacts/linkednotes`, `internal/modules/incidentbundles`, `internal/modules/incidentportability`, and `internal/modules/reporting`.
- Projection/read-side files under `internal/modules/*/projectionprovider`.
- Revision/history/rollback files under `internal/modules/revisions`.
- Route registration and OpenAPI contract tests for workbook, revisions, entities, and timeline routes.
- Schema and ownership inputs: `db/migrations/00011_links_and_tags.sql`, `db/migrations/00014_artifacts_and_optional_surfaces.sql`, and `tools/schema_object_ownership_manifest.json`.
- Harness and phase evidence inputs: `docs/testing/phase8_coverage_ledger.md`, `tools/phase8_test_map.json`, `tools/go_test_duration_baselines.json`, `Makefile`, and Make-owned task guidance output.

Planning finding: `internal/modules/links` was a mixed-responsibility package. It is a legitimate source-state facade for `record_links` and `record_tags`, but the baseline package also contained behavior overlapping artifacts and handoff risk refs, entity merge side effects, revisions/history/rollback, incident bundle portability, reporting export providers, and workbook collection orchestration. As of the 2026-07-09 remediation session, Handoff risk refs moved behind artifacts ownership, timeline field-key invalidation mapping moved behind a timeline adapter, revisions use a links-owned rollback target provider for link/tag table reconstruction, projection providers use links-owned active read views, and tag CRUD is exposed through a narrower links-owned tag facade.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/links/.gitkeep` | Placeholder file only. | None. | None found. | None. | None. | None. | None; out of runtime scope. | low | Inventory-only file. |
| `internal/modules/links/store.go` | Core store facade for active `record_links`; insert/update/tombstone link rows; insert `supersedes` rows. | `Store`, `NewStore`, errors, `RecordLink`, `SupersedesLink`, `GetActiveLinkTx`, `UpsertLinkTx`, `TombstoneLinkTx`, `InsertSupersedesTx`. | Workbook store, timeline ports/lifecycle, task/decision store, entity mentions, entity merge, artifacts linked notes. | `pgx`, `pgtype`, `uuid`, direct SQL on `record_links`. | `internal/modules/links/phase8_links_tags_test.go`, entity/timeline/revisions/task tests indirectly. | OpenAPI and generated contract behavior can be affected through route results but file does not edit generated artifacts. | `links` for record-link source state. | high | Legitimate facade, but direct SQL is also consumed elsewhere. |
| `internal/modules/links/field_refs.go` | Field-key driven reference links; task-decision link sync; linked-note references; tags through a narrow tag facade; item-ref parsing; field-to-link-type mapping. | `UpsertFieldReferenceTx`, `SyncTaskDecisionReferenceTx`, `InsertLinkedNoteReferenceTx`, `TombstoneFieldReferenceTx`, `TagStore`, `NewTagStore`, `Store.Tags`, `TagStore.UpsertTagTx`, `TagStore.TombstoneTagTx`, `ParseRecordTagItemRef`, `FieldLinkType`. | Workbook mutations, task/decision create/patch/import, linked notes, collection actions. | `pgx`, `uuid`, direct SQL on `record_links`, `record_tags`. | Phase 8 links/tags tests; Phase 9 task/decision and notes tests indirectly. | View-schema collection/write-action contracts and OpenAPI route behavior may drift if changed. | `links` for record links/tags; artifacts owns `handoff_risk_refs`; `tasksdecisions` consumes task-decision sync. | high | Risk-ref mutation was removed from links during the 2026-07-09 remediation session. |
| `internal/modules/links/collection_actions.go` | Validates and applies `collection_actions_v1` payloads for record refs, party refs, and tags. Handoff risk-ref collection payloads are routed by workbook to the artifacts facade. | `CollectionActionPayload`, `CollectionAction`, `CollectionValidationError`, `ValidateCollectionPayloadsTx`, `ValidateCollectionPayloadTx`, `ApplyCollectionPayloadsTx`, `ApplyCollectionPayloadTx`. | Workbook create/patch, linked-note create, validation helpers. | `pgx`, `uuid`, direct SQL validation against `records`; calls link/tag store methods. | Phase 8 links/tags tests; workbook task/decision, artifact, and linked-note tests indirectly. | View-schema collection action behavior and generated TypeScript/OpenAPI contracts can be affected. | Split: `links` collection facade plus artifacts-owned risk refs and workbook-owned request translation. | high | Contains orchestration keyed by public `field_key` values from workbook surfaces; Handoff risk refs no longer enter this facade. |
| `internal/modules/links/merge_effects.go` | Repoints/dedupes link and tag rows during entity merge; returns mutation payloads and relationship effects. | `MergeMutation`, `RepointMergedLinksCommand`, `RepointMergedLinksResult`, `RepointMergedTagsCommand`, `RepointMergedTagsResult`, `RepointMergedLinksTx`, `RepointMergedTagsTx`. | `internal/modules/entities/merge/ports.go` and `merge_store.go`. | `pgx`, `pgtype`, direct SQL on `record_links` and `record_tags`; calls store link methods. | Entity merge and revisions tests indirectly; Phase 7 rollback/merge tests inspect `record_links` and `record_tags`. | Record history and rollback mutation surfaces can drift. | `links` for link/tag row repoint mechanics; `timeline/linkeffects` maps relationship effects to timeline field keys; `entities/merge` owns merge workflow; `revisions` owns mutation/history semantics. | high | Timeline field-key literals were removed from links during the 2026-07-09 remediation session. |
| `internal/modules/links/incident_bundle_portability.go` | Exports/imports incident bundle files for `record_links`, distinct `tags`, and `record_tags`. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`. | `internal/modules/incidentbundles/source.go`; `internal/modules/incidentportability/import_targets.go` names the port. | `incidentportability`, `pgx`, `uuid`, direct SQL on `record_links` and `record_tags`. | Incident bundle route/API tests indirectly; target descriptors inspected. | Incident bundle `data/record_links.ndjson`, `data/tags.ndjson`, `data/record_tags.ndjson`. | `links` owner-provider adapter for portability. | medium | Legitimate owner provider if it remains thin and port-shaped. |
| `internal/modules/links/reportingprovider/provider.go` | Reporting extension provider for support refs and link/tag derived export facts. | `CollectSupportRefsTx`, `CollectFieldsTx`, `CollectFactsTx`. | `internal/modules/reporting/export_materializer.go`; reporting boundary guard tests. | `reporting/exportprovider`, `pgx`, direct SQL on `record_links` and `record_tags`. | `internal/modules/reporting/boundary_guard_test.go` and reporting tests indirectly. | Reporting export model fields under provider key `links`; no generated files directly. | `links/reportingprovider` as a thin reporting adapter. | medium | Existing reporting guard allows this provider path. Keep provider isolated from core reporting internals. |
| `internal/modules/links/phase8_links_tags_test.go` | Characterization tests for typed link vocabulary, tags, projection/history/query, rollback, and collaboration event behavior. | Test functions only. | Make targets `backend-store`, `backend-integration`, phase 8 slices. | `timeline`, `authn`, `postgres`, phase testutil packages, HTTP/WebSocket harness. | Direct test file. | Harness phase accounting and duration baselines reference these tests. | Tests/harness evidence, not runtime owner. | medium | Must be preserved or intentionally remapped if files move. |

## 3. Module Boundary Diagnosis

Current classification:

- Legitimate thin application/service facade: partially true for `record_links` and `record_tags` CRUD and field-key mapping.
- Accidental catch-all: true for risk refs, collection orchestration, merge side effects, portability, and reporting provider colocation.
- View/projection orchestration layer: partially true through timeline invalidation hints and projection-dependent tests.
- Transport-adjacent adapter: indirectly true through workbook route request translation, but no HTTP handlers live here.
- Persistence-adjacent adapter: true; most production methods directly issue SQL.
- Mutation coordinator: true for `ApplyCollectionPayloadTx` and merge repoint methods.
- Frontend shell/controller surface: no frontend code in target.
- Grid-vendor integration layer: no direct grid-vendor coupling found.
- Misplaced home for logic owned by other modules: fixed for `handoff_risk_refs`; timeline field-key mapping, revisions link/tag rollback, and projection read coupling now use owner contracts. Collection orchestration remains split between workbook request translation and links source-state application.
- Mixed-responsibility package: reduced; remaining intentional adapters are incident bundle and reporting provider surfaces.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Active record-link source-state CRUD | `store.go` | `links` | keep | Core 01 required modules; Core 02 record-link rules; migration `00011_links_and_tags.sql`; schema owner `links-and-tags`. | Valid module responsibility. |
| Record tag source-state CRUD | `field_refs.go`, `merge_effects.go` | `links` or narrower `tags` subfacade inside links | keep/split | Core 02 names Tag; schema owner maps `record_tags` to `links`. | Split only if it reduces public surface without behavior drift. |
| Field-key relationship mapping | `collection_actions.go`, `field_refs.go` | `links` facade plus Core 01 view-schema contracts | keep/split | Core 01 REQ-01-311 requires server-derived link type/direction from field/action route. | Must not derive behavior from labels. |
| Handoff risk refs | `artifacts/handoff_risk_refs.go`, `workbook/mutation_store.go` | `artifacts` Handoff child-row facade | moved | `handoff_risk_refs` schema owner is `artifacts-and-optional-surfaces`, not links. | Links production code no longer mutates `handoff_risk_refs`; public `handoff.open_risk_refs` behavior is preserved by workbook dispatch. |
| Entity merge link/tag repoint | `merge_effects.go` | `entities/merge` workflow plus `links` row-level port | split | Core 02 merge rules; `entities/merge` orchestrates merge and change set. | Link/tag row operations can remain in links; workflow invalidation belongs outside. |
| Timeline invalidation hints for merge | `timeline/linkeffects`, `entities/merge` | `timeline` invalidation adapter plus entity merge workflow | moved | Link merge effects now return relationship link types; `timeline/linkeffects` maps those effects to `timeline.host_refs` and `timeline.identity_refs`. | Future timeline field-key changes are localized outside links. |
| Incident bundle link/tag files | `incident_bundle_portability.go` | `links` owner-provider adapter | keep | Core 01 incident bundle registry includes `record_links`, `tags`, `record_tags`; import target owner is links. | Thin provider adapter is acceptable. |
| Reporting link/tag export facts | `reportingprovider/provider.go` | `links/reportingprovider` as reporting adapter | keep | Reporting materializer imports allowed provider; boundary guard requires provider shape. | Keep provider thin; do not move into reporting core. |
| Workbook collection request translation | `workbook/mutation_store.go` plus `collection_actions.go` | `workbook` for request DTO translation; `links` for source-state application | split | Workbook store converts workbook `CollectionActionPayload` to links payload. | Avoid making links own workbook route envelopes. |
| Revision/history/rollback behavior for links/tags | `revisions/*`, `links/revisionprovider` | `revisions` selectors and semantics plus links-owned target provider | split | Revisions rollback/history route semantics stay revisions-owned; link/tag value load/apply/canonical reconstruction moved to `links/revisionprovider`. | Future link/tag schema changes have one owner-side provider to update. |
| Projection read models over links/tags | `active_record_links_v1`, `active_record_tags_v1`, projection providers | links-owned SQL read contracts consumed by projection providers | split | Projection providers now join owner-defined active read views instead of raw link/tag tables. | Efficient SQL joins remain available while active-row and endpoint-not-deleted semantics are centralized. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` | Workbook/Core 01 | `internal/modules/workbook/routes.go`, `mutation_api.go`, Core 01 view contracts. | Workbook OpenAPI tests; Phase 8/9 backend tests. | Preserve collection action create cases for tags and record refs. | high | Create path applies links collection payloads and task decision reference sync. |
| `PATCH /api/v1/records/{record_id}` | Workbook/Core 01 | `workbook/routes.go`, `mutation_store.go`, `mutation_api.go`. | `phase8_links_tags_test.go::TestPhase8_LinkTagProjectionHistoryQuery_I_8_03`. | Add targeted patch cases before moving collection action application. | high | Patch path touches projections, revisions, WebSocket change keys. |
| `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query` | Workbook/Core 01 | `workbook/routes.go`; projection providers. | Phase 8 query tests. | Preserve tag/ref query cells and projection counts. | high | Projection providers read `record_links` and `record_tags` directly. |
| `POST /api/v1/records/{record_id}/linked-notes` | Workbook/artifacts/Core 01 | `workbook/routes.go`, `artifacts/linkednotes/facade.go`, OpenAPI test. | Workbook OpenAPI test; Phase 9 notes tests. | Characterize references_artifact link and note tag behavior if moving linked-note refs. | high | Uses `InsertLinkedNoteReferenceTx` and collection payloads. |
| Timeline and decision supersede routes | Workbook/timeline/tasksdecisions/Core 01/Core 02 | `workbook/routes.go`, `timeline/lifecycle_store.go`, `tasksdecisions/mutations.go`. | Phase 9 task/decision tests; incident bundle supersede tests. | Preserve active `supersedes` uniqueness and conflict behavior. | high | `InsertSupersedesTx` is shared by timeline and decisions. |
| Entity mention resolve/dismiss/revert | Entities/Core 02 | `entities/routes.go`, `entities/mentions/mention_lifecycle.go`. | Entity Phase 4 tests and OpenAPI tests. | Preserve link create/tombstone mutation values. | high | Entity mentions call links facade for `observed_on_host` and `observed_as_identity`. |
| Entity merge | Entities/Core 02/Core 01 destructive-operation rules | `entities/routes.go`, `entities/merge/merge_store.go`, `links/merge_effects.go`. | Phase 7 merge/rollback tests and entity tests. | Characterize link/tag repoint/dedupe counts and timeline invalidations before moving. | high | Merge side effects cross links, tags, mentions, assessments, projections, revisions. |
| `GET /api/v1/records/{record_id}/history` and rollback | Revisions/Core 01/Core 02 | `revisions/routes.go`, `rollback_api.go`, `rollback_store.go`, `store.go`. | Phase 7 rollback tests; Phase 8 tag rollback test. | Preserve `record_link` and `record_tag` history item refs and rollback payloads. | high | Revisions directly read/write link/tag tables. |
| WebSocket `record_changed` | Collaboration/Core 01/Core 03 | Phase 8 test uses `phase4test.ConnectViewSocket` and `RequireRecordChanged`. | `TestPhase8_LinkTagProjectionHistoryQuery_I_8_03`. | Preserve `changed_field_keys` for tag/evidence link mutations. | high | Links package does not publish events, but mutations affect event payload. |
| Incident bundle files | Incident portability/Core 01 | Core 01 bundle registry; `links/incident_bundle_portability.go`; import target descriptors. | Incident bundle route/API tests. | Preserve `record_links.ndjson`, `tags.ndjson`, and `record_tags.ndjson`. | medium | Generated artifacts not hand-edited; provider remains source owner. |
| Reporting export facts and support refs | Reporting extension/Core 01/Snapshot and Reporting profile | `links/reportingprovider/provider.go`, `reporting/export_materializer.go`, reporting boundary guard. | Reporting boundary guard and reporting tests. | Preserve provider key `links`, support ref paths, and duplicate path validation. | medium | Keep as owner-provider adapter. |
| View-schema collection actions | Core 01 view schema registry | Core 01 REQ-01-311/REQ-01-569; contracts for timeline/notes/task/decision views. | Workbook OpenAPI contract tests; Phase 8/9 tests. | Preserve fail-closed unknown ops, tag normalization, no client confidence. | high | Changes can drift generated protocol/view contracts. |
| Generated OpenAPI/contracts | Core 01/contracts layer | `internal/gen/contracts`, `contracts/view-schemas`, OpenAPI tests. | OpenAPI contract tests. | Run generator/drift only if later implementation changes derived contracts. | medium | Do not hand-edit generated files. |
| Harness/test accounting | Testing harness NLSpec | `docs/testing/phase8_coverage_ledger.md`, `tools/phase8_test_map.json`, Make target guidance. | Phase 8 rows U-8-01 and I-8-03 reference links tests. | If tests move, update owner inputs before generated ledgers. | medium | Phase maps are evidence accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `links` package directly mutates `handoff_risk_refs` although schema ownership maps it to artifacts. | Baseline: `field_refs.go`; `collection_actions.go`; `tools/schema_object_ownership_manifest.json`. Current: `artifacts/handoff_risk_refs.go`; `workbook/mutation_store.go`. | Handoff/artifact behavior may be hidden behind links facade. | resolved | `artifacts` Handoff child-row facade | Implemented by routing `handoff.open_risk_refs` workbook payloads to artifacts and removing links risk-ref mutation APIs. |
| `collection_actions.go` mixes public workbook field-key orchestration with source table mutation. | `ApplyCollectionPayloadTx`; Core 01 relationship routing rules. | Workbook route behavior can drift during source-state refactor. | should_fix | `workbook` for route DTOs; `links` for source-state application | Split DTO translation from row-level link/tag operations after characterization. |
| Merge link/tag repoint returns timeline field-key invalidation hints. | Baseline: `merge_effects.go::mergeLinkTypeFieldKey`. Current: `links` returns relationship effects; `timeline/linkeffects` maps effects to field keys. | Links source-state package leaks projection/workbook field identity. | resolved | `entities/merge` plus timeline adapter | Implemented by moving field-key mapping out of links. |
| Projections directly query links-owned tables from multiple owner packages. | Baseline projection providers for artifacts, assessments, entities, indicators, tasksdecisions. Current providers consume `active_record_links_v1` and `active_record_tags_v1`. | Source-schema changes can break projections without facade coverage. | resolved | links-owned SQL read contracts plus per-surface projection providers | Implemented by additive migration and schema-owner manifest update. |
| Revisions directly read/write `record_links` and `record_tags`. | Baseline: `revisions/rollback_store.go`, `delete_restore_store.go`, `store.go`. Current: `links/revisionprovider` owns link/tag rollback target SQL. | Rollback/history drift, stale target behavior, mutation payload shape drift. | resolved | `revisions` route semantics plus links target provider | Implemented provider contract; revisions retains selectors, eligibility, idempotency, and route semantics. |
| Reporting core imports `links/reportingprovider` under an explicit guard. | `reporting/export_materializer.go`, `reporting/boundary_guard_test.go`. | Low if provider remains thin; high if reporting logic moves into links core. | intentional/no_action | `links/reportingprovider` plus `reporting` | Preserve provider pattern; keep reporting core free of owner-table reads. |
| Incident bundle import/export names links as owner for `record_links` and `record_tags`. | `incident_bundle_portability.go`; `incidentportability/import_targets.go`; Core 01 bundle registry. | Bundle compatibility drift if file names or row identity change. | intentional/no_action | `links` owner-provider adapter | Keep current files stable unless Core 01 changes. |
| Package exposes many public methods on one `Store`. | `store.go`, `field_refs.go`, `collection_actions.go`, `merge_effects.go`, `links/revisionprovider`. | Shallow facade encourages peer modules to depend on too-wide surface. | partially_resolved | Narrow owner ports | Tag CRUD now has a narrow facade and revisions use a target provider; relationship write/store splitting can continue in later cleanup. |
| Test package name and phase file name are phase-shaped. | `phase8_links_tags_test.go`; phase ledger/map. | Runtime low; evidence accounting can become stale if test moves. | defer | Test/harness accounting | If renamed or moved, update phase-map owner inputs and drift targets. |
| No direct grid-vendor coupling found in target. | Search of links package and callers inspected. | None. | intentional/no_action | `/packages/grid-adapter` remains owner if future frontend work appears | No action. |
| No production transport/platform imports inside `internal/modules/links` except database transaction abstractions. | Links files import `pgx`, `uuid`, `incidentportability`, reporting export provider. | Persistence coupling remains; transport coupling not observed. | defer | `links` store/persistence adapter | Consider store interface split later; no immediate transport fix. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, target path, constraints, and tracker baseline. | This tracker, framework, Core docs, domain, harness spec. | No product validation; `git diff --check` for tracker only if desired. | Tracker Section 1 and handoff log seeded. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every file in `internal/modules/links`. | `internal/modules/links/**`. | None; source inspection only. | Section 2 complete. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map route, view-schema, portability, reporting, revision, and collaboration contracts. | Workbook/revision/entity/timeline routes and APIs, Core 01-04. | `make explain-phase PHASE=phase8` for guidance only. | Section 4 owner map complete. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-06 | Identify existing tests and missing pre-move tests. | Links tests, phase maps, OpenAPI tests, revisions/entity/task tests. | `make backend-store`, `make backend-integration` when implementation begins. | Required characterization rows known. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05 | Classify mixed responsibilities and known boundary leaks. | Direct callers, projection providers, revisions, schema owner manifest. | `make backend-module-boundary-check` later. | Section 5 findings complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-04 | WF-06, WF-07 | Define narrow behavior-preserving facade split. | `links`, workbook adapters, artifacts/handoff, entities/merge, reportingprovider. | No implementation validation until slices start. | Candidate owner decisions recorded. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07, WF-08 | Sequence smallest safe refactor slices. | Same as WF-05 plus revisions/projections as deferred risk areas. | Per-slice Make targets. | Section 7 complete. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-06 | WF-08 | Plan test moves or characterization additions without hand-editing generated outputs. | `tools/phase8_test_map.json`, ledgers only through owner inputs/generators if needed. | `make phase-ledger-drift`, `make phase-schedule-drift` only if evidence inputs change. | Harness action marked required or not applicable per slice. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Record validation commands, blockers, and next task. | Tracker, Make targets. | `make agent-finalize`, narrow targets, then broad gate as needed. | Handoff tables updated. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-01 | none | Characterize current links/tags public behavior before movement; no production behavior change. | `internal/modules/links/phase8_links_tags_test.go`, existing workbook/entity/revision tests. | None if test-only; generated/harness accounting if tests move. | Preserve U-8-01 and I-8-03; add focused cases only where existing coverage is missing. | `make backend-store`; `make backend-integration`; broaden to `make phase-slice PHASE=phase8` when needed. | Drop new characterization tests if they are wrong; do not alter production. | Current link/tag behavior is captured before refactor edits. |
| S-02 | S-01 | Introduce narrow owner ports/adapters around link CRUD, tag CRUD, field refs, and supersedes without changing SQL or exported behavior. | `internal/modules/links`, caller ports in workbook/timeline/tasksdecisions/entities/artifacts. | Route payloads, mutation idempotency, WebSocket changed keys, history mutation payloads. | Preserve Phase 8 tests, OpenAPI tests, entity mention/merge tests. | `make backend-store`; `make backend-integration`; `make backend-module-boundary-check`. | Revert adapter indirection while keeping tests. | Callers depend on narrow interfaces; behavior and SQL are unchanged. |
| S-03 | S-01, S-02 | Move or wrap `handoff_risk_refs` mutation behind artifacts/handoff ownership; requires later authorization if ownership or public behavior changes. | `links/field_refs.go`, `links/collection_actions.go`, `internal/modules/artifacts`, handoff surface code. | Handoff open risk refs, collection item refs, revision mutation payloads, projection/query cells. | Add/keep handoff risk-ref collection create/patch/query/history characterization. | `make backend-store`; `make backend-integration`; `make phase-slice PHASE=phase8` if phase-owned rows cover it. | Restore links-owned risk-ref methods if artifacts wrapper changes behavior. | `handoff_risk_refs` owner boundary is explicit and behavior-preserving, or remains deferred with documented owner decision. |
| S-04 | S-02 | Keep incident bundle and reporting behavior as thin owner-provider adapters; clarify boundaries without moving generated or public bundle names. | `links/incident_bundle_portability.go`, `links/reportingprovider/provider.go`, incident bundle source, reporting materializer. | Bundle file names, provider keys, support-ref paths, reporting duplicate path checks. | Preserve incident bundle and reporting boundary tests. | `make backend-integration`; `make backend-module-boundary-check`; `make generated-artifact-policy-check` if any generation-adjacent files change. | Revert provider wiring only; do not alter bundle contents. | Provider contracts remain stable and imports stay allowed only through guardrails. |
| S-05 | S-01, S-02 | Plan and later isolate revisions/projection direct table coupling behind owner APIs; no table/schema behavior change. | `internal/modules/revisions`, `internal/modules/*/projectionprovider`, `internal/modules/links`. | Rollback/history item refs, projection counts, search/filter/sort results. | Add targeted rollback and projection-count characterization before edits. | `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase8`; browser targets only for UI/WebSocket-impacting changes. | Revert facade changes before schema or payload changes. | Direct table coupling is either reduced through behavior-preserving owner APIs or deferred with tests and risks recorded. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-store` | Store-backed Go tests, including phase 8 links/tags store coverage. | yes | Discovered via `make explain-target TARGET=backend-store DETAIL=summary`. Requires Postgres/object store. |
| integration | `make backend-integration` | HTTP/workbook/integration tests, including Phase 8 link/tag projection/history/query behavior. | yes | Discovered via Make target guidance. Requires Postgres/object store. |
| e2e/browser | `make phase-slice PHASE=phase8` or targeted browser targets named by `make explain-phase PHASE=phase8` | Browser/WebSocket/workbook behavior when a slice affects UI, collaboration, or shell state. | no for pure backend adapter moves; yes when WebSocket/UI behavior can drift | Phase guidance lists webserver-backed and stateful browser evidence for phase 8. |
| generated drift | `make generated-artifact-policy-check`; `make generate-drift` only if derived contracts are regenerated | Generated-file policy and generated-output drift. | no unless generated/contract inputs change | Do not hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module import boundaries from `tools/backend_module_boundaries.json`. | yes for boundary-moving slices | Reporting also has owner-local boundary guard tests. |
| full check | `make test-fast`; `make check` | Broad local gate. | no before first implementation slice; yes before final handoff of broad refactor | `make check` includes browser stack. Choose narrowest first, then broaden by risk. |
| phase slice | `make service-backed-slice PHASE=phase8`; `make phase-slice PHASE=phase8` | Phase 8 service-backed and full phase evidence. | yes for behavior-affecting slices | Use phase maps as evidence accounting, not architecture authority. |
| finalization | `make agent-finalize` | Harness-maintenance/finalization check. | yes before broad end-of-run verification | Pass `RESULTS_DIR` only when retaining a successful full warm check run. |

Commands searched/discovered: `make help`, `make task-guide ROLE=feature-dev PHASE=phase8`, `make explain-target TARGET=backend-store DETAIL=summary`, `make explain-target TARGET=backend-integration DETAIL=summary`, `make explain-target TARGET=backend-module-boundary-check DETAIL=summary`, and `make explain-phase PHASE=phase8`.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| LT-001 | Establish links target scope and authority posture | WF-00 | DONE | none | Section 1 | Target and source hierarchy are explicit. |
| LT-002 | Inventory every file in `internal/modules/links` | WF-01 | DONE | LT-001 | Section 2 | Every target file is inventoried or out of scope. |
| LT-003 | Record mixed-responsibility boundary diagnosis | WF-04 | DONE | LT-002 | Section 3 and Section 5 | Legitimate and questionable ownership is separated. |
| LT-004 | Map public contract freeze surface | WF-02 | DONE | LT-002 | Section 4 | Route, view, revision, WebSocket, portability, reporting, generated, and harness risks are listed. |
| LT-005 | Plan characterization coverage | WF-03 | DONE | LT-004 | Existing Phase 7/8/9 backend-store coverage; remediation session validation | Missing behavior was covered by existing backend-store characterization for this slice; no phase-map changes were required. |
| LT-006 | Define narrow links/tag owner ports | WF-05 | DONE | LT-005 | `TagStore`, `links/revisionprovider`, timeline link effects adapter, projection views | Callers rely on narrower behavior ports without behavior drift in passing backend-store validation. |
| LT-007 | Resolve `handoff_risk_refs` owner mismatch | WF-05 | DONE | LT-005 | RB-001; `artifacts/handoff_risk_refs.go`; workbook dispatch | Risk-ref mutation is artifacts-owned; links production code no longer mutates `handoff_risk_refs`. |
| LT-008 | Preserve portability/reporting provider boundaries | WF-05 | DONE | LT-006 | Incident bundle and reporting provider files unchanged in role; backend boundary check passed | Provider contracts remain stable and guardrails remain in place. |
| LT-009 | Plan revisions/projection coupling isolation | WF-06 | DONE | LT-005, LT-006 | `links/revisionprovider`; `active_record_links_v1`; `active_record_tags_v1` | Direct coupling is reduced through owner contracts. |
| LT-010 | Select validation per implementation slice | WF-08 | DONE | LT-005 | Section 8; backend-store and backend-module-boundary-check runs | Narrow Make targets are attached to the implemented slice. |
| LT-011 | Update handoff after each later session | WF-08 | DONE | implementation session | Section 10 | Handoff log records the remediation session and validation evidence. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | Tracker created from live repository inspection; target exists. | Touched: `docs/handoffs/links-module-refactor-tracker.md`. Inspected: framework, Core 00-04, domain, harness spec, links files, callers, schema inputs. | `sed`, `find`, `rg`, `wc`, `git status`, `git rev-parse`, `date -Is`, `make help`, `make task-guide ROLE=feature-dev PHASE=phase8`, `make explain-target ...`, `make explain-phase PHASE=phase8`. | Planning tracker captures scope, inventory, contracts, findings, workstreams, validation, blockers. | Implementation not authorized in this task. | Start S-01 characterization in a later authorized task. |
| 2026-07-09T10:49:02-04:00 | Codex remediation session | Authorized RB remediation implemented: links remains source owner for `record_links`/`record_tags`, while risk refs, timeline field-key mapping, revision link/tag rollback, and projection reads now sit behind owner contracts. | Touched links, artifacts, workbook, timeline, entities merge, revisions, projection providers, migration, schema owner manifest, migration history manifest, and this tracker. | `rg`, `sed`, `gofmt`, `make backend-module-boundary-check`, `make backend-store`, `make json-shape-check`, `make migration-drift`, `make generated-artifact-policy-check`, `make agent-finalize`, `make backend-integration`, `make service-backed-slice PHASE=phase8`, `make phase-slice PHASE=phase8`, `make lint-markdown`. | All planned remediation validation passed. Latest key run roots: backend-store `.cartulary/test-results/20260709T144334Z-p67464`; backend-integration `.cartulary/test-results/20260709T144424Z-p85355`; phase-slice `.cartulary/test-results/20260709T144526Z-p1194`; service-backed-slice `.cartulary/test-results/20260709T144721Z-p28362`. | None known. | Keep future link/tag expansion on the owner contracts introduced here. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | `links` is a valid source-state home for `record_links` and `record_tags`, but current package is mixed. | Inspected `internal/modules/links/**`, workbook/timeline/tasksdecisions/entities/artifacts callers, projection providers, revisions. Touched tracker only. | `rg`, `sed`, `find`. | Boundary findings recorded. | `handoff_risk_refs` owner mismatch remains unresolved. | Narrow ports before moving behavior. |
| 2026-07-09T10:49:02-04:00 | Codex remediation session | Boundary leaks from links to artifacts, timeline field keys, revisions rollback SQL, and projection raw table reads were remediated. | Touched `internal/modules/artifacts/handoff_risk_refs.go`, `internal/modules/workbook/mutation_store.go`, `internal/modules/timeline/linkeffects/linkeffects.go`, `internal/modules/entities/merge/*`, `internal/modules/revisions/*`, `internal/modules/links/*`, and projection providers. | `make backend-module-boundary-check`; `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase8`; `make phase-slice PHASE=phase8`. | Boundary, store, integration, service-backed Phase 8, and full Phase 8 validation passed. Latest boundary run root: `.cartulary/test-results/20260709T144334Z-p67492`. | None known from boundary validation. | Keep future caller additions on the owner contracts rather than raw table/package coupling. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | No frontend code in target; UI/WebSocket behavior can still drift through workbook contracts. | Inspected route registrations and phase guidance. Touched tracker only. | `make explain-phase PHASE=phase8`, `rg`. | Browser validation needed only for UI/WebSocket-impacting implementation slices. | None for tracker. | Use phase 8 browser guidance when slice affects collaboration or workbook UI. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | Generated contracts are downstream; no generated file edits planned. | Inspected view-schema/OpenAPI contract tests and generated-artifact policy references. Touched tracker only. | `rg`, `sed`. | Contract freeze map includes OpenAPI, view-schema collection actions, incident bundle files, and reporting facts. | Generated drift only relevant if owner inputs change later. | Do not hand-edit generated roots; use Make generators when authorized. |
| 2026-07-09T10:49:02-04:00 | Codex remediation session | Public contracts remain unchanged; new database read views are additive owner contracts and generated roots were not edited. | Touched `db/migrations/00025_links_projection_read_contracts.sql`, `tools/schema_object_ownership_manifest.json`, and `tools/migration_history_manifest.json`. | `make json-shape-check`; `make migration-drift`; `make generated-artifact-policy-check`; `make agent-finalize`. | JSON shape, migration drift, generated-artifact policy, and finalization all passed. Run roots: json-shape `.cartulary/test-results/20260709T143614Z-p82162`; migration-drift `.cartulary/test-results/20260709T143614Z-p82186`; generated policy `.cartulary/test-results/20260709T144121Z-p56725`; agent-finalize `.cartulary/test-results/20260709T143625Z-p84776`. | None known. | No generated roots were edited; rerun drift checks if the migration changes again. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | Phase 8 links/tags tests are primary characterization evidence. | Inspected Phase 8 test file, ledger/map references, Make target guidance. Touched tracker only. | `make task-guide ROLE=feature-dev PHASE=phase8`, `make explain-target ...`, `make explain-phase PHASE=phase8`. | Validation plan populated with Make-owned targets. | No product validation run in this planning-only session. | Run narrow targets when implementation begins. |
| 2026-07-09T10:49:02-04:00 | Codex remediation session | Existing Phase 7/8/9 backend-store, backend integration, and Phase 8 slice tests cover rollback/history, links/tags, Handoff risk refs, projections, workbook dispatch, and collaboration-sensitive behavior touched by the slice. | Touched backend production files and tracker; no phase map or generated ledger inputs changed. | `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase8`; `make phase-slice PHASE=phase8`. | Passed on final code: backend-store 130 tests, backend-integration 163 tests, service-backed Phase 8 41 tests, full Phase 8 slice 49 tests. | None known. | Add provider-focused unit tests if future link/tag schema changes expand rollback behavior. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | Links package does not own auth; route authorization is re-derived by workbook/entities/revisions/timeline services. | Inspected Core 04 authz posture and route registrations. Touched tracker only. | `sed`, `rg`. | Refactor must preserve route-level incident role and CSRF/auth behavior by not moving authorization into links. | None for tracker. | Include route authorization checks in characterization if service boundaries move. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T10:08:47-04:00 | Codex planning/documentation session | Main open risk is source-owner ambiguity around risk refs and hidden projection/revision coupling. | Touched tracker only. | Discovery commands only. | Risks and blockers listed. | RB-001 and deferred projection/revision coupling. | Begin with S-01, then S-02. |
| 2026-07-09T10:49:02-04:00 | Codex remediation session | RB-001 through RB-005 owner decisions are implemented and validation breadth for the touched backend/workbook/Phase 8 surfaces passed. | Touched tracker and implementation files listed above. | `make backend-module-boundary-check`; `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase8`; `make phase-slice PHASE=phase8`; drift/finalization checks. | Remediation validation passed on final code. | None known. | Later phases should treat these owner contracts as the starting point, not raw table/package access. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Should `handoff_risk_refs` remain reachable through `links`, or move/wrap behind artifacts/handoff ownership? | Schema owner manifest maps `handoff_risk_refs` to artifacts, but links mutating it hid Handoff ownership. | Authorized remediation plan plus backend-store characterization. | RESOLVED: moved behind artifacts-owned Handoff risk-ref facade; links production code no longer mutates `handoff_risk_refs`. |
| RB-002 | Should timeline field-key invalidation mapping remain in `links/merge_effects.go`? | Link merge side effects returning `timeline.host_refs` and `timeline.identity_refs` leaked workbook identifiers. | Entity merge behavior and timeline projection invalidation characterization. | RESOLVED: links returns relationship effects; `timeline/linkeffects` maps relationship effects to timeline field keys. |
| RB-003 | What facade should revisions use for rollback/history link/tag operations, if any? | Revisions directly reading/writing `record_links` and `record_tags` made rollback brittle for future link/tag schema changes. | Phase 7 rollback/history coverage and revisions route semantics. | RESOLVED: revisions uses links-owned rollback target provider while keeping selectors, eligibility, idempotency, and routes revisions-owned. |
| RB-004 | What facade should projection providers use for link/tag-derived counts? | Multiple projections directly queried links-owned tables, so source-table changes could break projection behavior. | Projection behavior coverage and schema-owner manifest update. | RESOLVED: added links-owned `active_record_links_v1` and `active_record_tags_v1` read views and updated projection providers to consume them. |
| RB-005 | Should `record_tags` stay under `links` or split into a narrower tags submodule/facade? | Core 02 names tags as a domain concept, while Core 01 groups links/tags/coordination. | Authorized owner decision and Phase 8 link/tag behavior. | RESOLVED: `record_tags` stays under Links and Tags; tag CRUD is exposed through a narrow `TagStore` facade, not a new top-level module. |

## 12. Binary Completion Criteria

- Every file in `internal/modules/links` is inventoried or explicitly out of scope: complete.
- Every discovered public contract risk has an owner and test posture: complete for implemented remediation; backend-store, backend integration, Phase 8 slices, backend boundary, migration drift, JSON shape, generated policy, and finalization passed.
- Every proposed workflow has dependencies and exit criteria: complete.
- Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`: complete.
- Validation commands are discovered or marked `TODO` with a reason: complete.
- Contradictions are marked `BLOCKED: owner contradiction`: no owner contradiction found during this inspection.
- Repository/framework mismatches are recorded as planning findings: complete; mixed-responsibility and `handoff_risk_refs` mismatch recorded.
- Handoff sections are current enough for another agent to continue without rediscovery: complete for the 2026-07-09 remediation session.
