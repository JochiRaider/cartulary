# revisions Module Refactoring Tracker and Handoff

Current implementation handoff: Section 13 is the authoritative continuation point for the next Revisions remediation iteration. Sections 1 through 12 are preserved as the historical planning and first-remediation record; where their future-work descriptions differ from the live tree, Section 13 supersedes them.

## 1. Scope and Source Posture

- Target path: `internal/modules/revisions`.
- Target label: `revisions`.
- Normalized target label: `revisions`.
- Output path: `docs/handoffs/revisions-module-refactor-tracker.md`.
- Status: implementation handoff. The original tracker was planning-only; a later authorized remediation pass has now completed the first structural slices.
- Allowed changes for the original planning session: create or update only this tracker file. The remediation session also touched owner specs and production code listed in Section 10.
- Non-goals retained for the current remediation pass: no generated artifact hand edits, migrations, bundle v1 file-name changes, public HTTP route shape changes, error-code changes, WebSocket event-shape changes, or conflict-token wire-compatibility changes.
- Implementation posture: shared primitive moves and the Timeline rollback provider seam are implemented; remaining source-owner rollback and delete/restore provider migrations are staged owner by owner.
- Prior tracker state: no existing `docs/handoffs/revisions-module-refactor-tracker.md` was found, so no prior handoff history needed preservation.

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

Repository files and surfaces inspected:

- `AGENTS.md`
- all files under `internal/modules/revisions`
- targeted callers under `internal/app`, `internal/modules/workbook`, `internal/modules/timeline`, `internal/modules/incidentbundles`, `internal/modules/imports`, `internal/modules/entities`, `internal/modules/evidence`, `internal/modules/indicators`, `internal/modules/parties`, `internal/modules/tasksdecisions`, `internal/modules/artifacts`, and `internal/modules/assessments`
- generated and derived contract surfaces under `contracts/openapi`, `contracts/errors`, `contracts/ws`, `packages/view-contracts`, `packages/ui-contracts`, and `internal/gen/contracts`
- phase and harness evidence under `docs/testing`, `tools/phase7_test_map.json`, and public Make target listings

Planning findings:

- The original baseline target path contained 20 Go files: 10 production files and 10 test files. After remediation, conflict-token files moved out of revisions and `command_service.go`, `projection_port.go`, and `rollback_providers.go` were added, so the target path currently contains 21 Go files.
- `revisions` is a legitimate logical concern for change sets, row revisions, history, rollback, restore, and destructive-operation contention.
- The current package is broader than that logical concern. It also contains route/auth/session/collaboration glue, workbook same-field conflict support, incident bundle portability adapters, projection rebuild orchestration, and direct owner-specific rollback/delete logic.
- No owner contradiction was found in inspected owner documents.
- Repository/framework mismatch: the framework catalog describes `revisions` narrowly as change sets, row revisions, rollback, and restore, while the live package includes additional workbook, transport, projection, portability, and source-owner coupling.

Commands run for discovery, not validation:

- `sed` reads of the planning framework, AGENTS instructions, owner docs, target files, and selected adjacent files.
- `rg --files internal/modules/revisions`
- `rg -n '^func |^type |^const |^var ' internal/modules/revisions/*.go`
- `rg -l 'github.com/JochiRaider/cartulary/internal/modules/revisions' internal cmd apps packages db contracts tools docs`
- targeted `rg` searches for routes, rollback, history, generated contracts, phase7 evidence, and Make targets.
- `make help`
- `git status --short --branch`, `git branch --show-current`, and `git rev-parse --short HEAD`

Tracker-only validation run after creation:

- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260709T212646Z-p33678`.
- `make json-shape-check` passed with run root `.cartulary/test-results/20260709T212650Z-p33852`.
- `make agent-finalize` was skipped because this task allowed writes only to this tracker file and `agent-finalize` may update harness-maintenance artifacts.

Implementation validation run after remediation:

- `make backend-unit` passed with run root `.cartulary/test-results/20260709T220055Z-p49805` after fixing one unused import found by the first run.
- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260709T220405Z-p57548`.
- `make json-shape-check` passed with run root `.cartulary/test-results/20260709T220406Z-p57583`.
- `make phase-slice PHASE=phase7` passed with run root `.cartulary/test-results/20260709T220421Z-p59480`.
- `make phase-slice PHASE=phase6` passed with run root `.cartulary/test-results/20260709T220501Z-p76072`.
- `make phase-slice PHASE=phase3` passed with run root `.cartulary/test-results/20260709T220650Z-p98041`.
- `make phase-slice PHASE=phase4` passed with run root `.cartulary/test-results/20260709T220734Z-p22063`.
- `make phase-slice PHASE=phase11` passed with run root `.cartulary/test-results/20260709T220820Z-p40374`.
- `make backend-module-boundary-check` passed with run root `.cartulary/test-results/20260709T220905Z-p56872`.
- `make test-fast` passed with run root `.cartulary/test-results/20260709T220912Z-p57066`.
- `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260709T220912Z-p57066` failed retained-run preflight because the supplied `test-fast` run is not valid retained full warm-check evidence; failure run root `.cartulary/test-results/20260709T221141Z-p7906`.
- `make agent-finalize` passed without retained evidence with run root `.cartulary/test-results/20260709T221305Z-p8414`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- After the final tracker update, `make lint-markdown` passed, `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260709T221438Z-p11607`, `make json-shape-check` passed with run root `.cartulary/test-results/20260709T221443Z-p11784`, and `make agent-finalize` passed with run root `.cartulary/test-results/20260709T221447Z-p12126`.
- After extracting delete/restore source providers, `make backend-unit` passed with run root `.cartulary/test-results/20260709T222133Z-p18807`, `make phase-slice PHASE=phase7` passed with run root `.cartulary/test-results/20260709T222232Z-p27142`, `make phase-slice PHASE=phase3` passed with run root `.cartulary/test-results/20260709T222310Z-p43027`, `make phase-slice PHASE=phase4` passed with run root `.cartulary/test-results/20260709T222402Z-p61477`, `make phase-slice PHASE=phase5` passed with run root `.cartulary/test-results/20260709T222455Z-p79553`, `make phase-slice PHASE=phase9` passed with run root `.cartulary/test-results/20260709T222532Z-p93541`, `make backend-module-boundary-check` passed with run root `.cartulary/test-results/20260709T222627Z-p8097`, `make test-fast` passed with run root `.cartulary/test-results/20260709T222632Z-p8283`, and `make phase-slice PHASE=phase11` passed with run root `.cartulary/test-results/20260709T222850Z-p56320`.

## 2. Baseline and Remediation Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/revisions/attribution_resolver_registry.go` | Registry for imported attribution resolvers required by claimed extension profiles. | `ErrDuplicateAttributionResolver`, `ErrMissingAttributionResolver`, `ExtensionClaim`, `AttributionResolverRegistry`, `NewAttributionResolverRegistry`, registry methods. | `internal/app/runtime.go`; local tests. | package-local `ImportedAttributionResolver`; `errors`, `fmt`. | `attribution_resolver_registry_test.go`. | none directly; protects incident portability attribution behavior. | revisions plus incident portability boundary | Medium | Validates the `incident_portability` resolver without joining sidecar attribution tables directly. |
| `internal/modules/revisions/attribution_resolver_registry_test.go` | Unit coverage for resolver requirement and duplicate registration. | test-only package `revisions`. | Go test only. | local registry and fake resolver. | self. | none. | revisions test evidence | Low | In scope as target test inventory. |
| `internal/modules/revisions/conflict_token.go` | Baseline-only row. HMAC-backed opaque conflict token v2 codec for workbook and timeline conflict resolution moved to `internal/modules/workbook/conflicts/conflict_token.go`. | No longer present in revisions. Replacement exports `ConflictTokenVersion`, `ConflictTokenClaims`, `ConflictTokenCodec`, `NewConflictTokenCodec`, `NewConflictTokenCodecForTesting`, `RequestHashTokenValue`, `Issue`, and `Parse`. | `internal/modules/workbook`, `internal/modules/timeline`, and revisions conflict helpers now import workbook conflict ownership. | `internal/platform/authn`, `uuid`, crypto/json/base64 via workbook conflict package. | `internal/modules/workbook/conflicts/conflict_token_test.go`, `workbook/phase6_conflict_support_test.go`, timeline/workbook conflict tests. | Public conflict token contract appears in OpenAPI conflict route surfaces; generated roots are downstream only. | workbook conflict ownership | High | v2 token compatibility is preserved while ownership moved. |
| `internal/modules/revisions/conflict_token_test.go` | Baseline-only row. Unit coverage for conflict token roundtrip, tamper rejection, and invalid record ID rejection moved to `internal/modules/workbook/conflicts/conflict_token_test.go`. | No longer present in revisions. | Go test only. | workbook conflict token codec. | self. | none. | workbook conflict evidence | Medium | Characterization moved with the codec. |
| `internal/modules/revisions/delete_restore_api.go` | Strict request decode, request hash, route keys, and API error mapping for delete/restore plus shared rollback errors. | `DeleteRestoreRequest`, `DecodeDeleteRestoreRequest`, `DeleteRestoreRequestHash`; package error helpers. | `routes.go`, local tests. | `fieldnorm`, `httpapi`, JSON/hash/http. | `phase7_delete_restore_test.go`, `phase7_integration_test.go`, `phase7_locks_test.go`. | `contracts/errors/index.json` and OpenAPI encode same error vocabulary downstream. | revisions route contract | High | Wire shape and reason normalization are frozen. |
| `internal/modules/revisions/delete_restore_store.go` | Transaction coordinator for soft-delete and restore, row-version/idempotency handling, destructive restore lock, source tombstone updates, projection rebuild, and delete preconditions. | errors, `RowVersionConflictError`, `RecordLockedError`, `RecordDeleteBlockedError`, `DeleteRestoreResult`, `DeleteRestoreAdapterTypes`, `SoftDeleteRecord`, `RestoreRecord`, `LockRecordEnvelopesNowaitTx`. | `routes.go`, `entities/merge/ports.go`, `parties/phase9_parties_test.go`, local tests. | `authn`, `viewschema`, `projectionadapters`, raw SQL over `records` and multiple source tables. | `phase7_delete_restore_test.go`, `phase7_integration_test.go`, `phase7_locks_test.go`. | OpenAPI delete/restore responses and error contracts; projection contracts downstream. | revisions coordinator plus source-owner ports | Critical | Direct source-owner adapter map covers timeline, entities, parties, indicators, artifacts, tasks, decisions, evidence, and assessments. |
| `internal/modules/revisions/imported_attribution_boundary_test.go` | Boundary test ensuring revisions code does not directly query incident-bundle attribution sidecars. | test-only package `revisions`. | Go test only. | file walk/read of revisions package. | self. | none. | revisions boundary evidence | Medium | Guard should be retained or replaced if attribution boundary moves. |
| `internal/modules/revisions/incident_bundle_portability.go` | Incident bundle export/import/repair adapter for revision history substrate. | `ExportIncidentBundleFiles`, `ImportIncidentBundleFilesTx`, `RepairIncidentBundleImportedSequencesTx`. | `internal/modules/incidentbundles/source.go`, incident bundle tests. | `incidentportability`, `pgx`, raw SQL over `change_sets`, `change_set_mutations`, `record_revisions`. | Incident bundle integration tests indirectly; phase11 ledger references imported history and rollback behavior. | Incident bundle logical files `data/change_sets.ndjson`, `data/change_set_mutations.ndjson`, `data/record_revisions.ndjson`. | revisions history substrate plus incident portability adapter | High | Legitimate provider adapter, but ownership boundary should remain explicit. |
| `internal/modules/revisions/phase7_delete_restore_test.go` | Phase7 unit/store coverage for delete/restore route preconditions and adapter matrix. | tests `TestPhase7_SoftDeleteRoutePreconditions_U_7_03`, `TestPhase7_RestoreTombstonePreconditions_U_7_04`, `TestPhase7_DeleteRestoreAdapterMatrix_U_7_03_U_7_04`. | Go test only. | phase4 test harness, HTTP helpers, DB fixtures. | self. | reads generated contract expectations indirectly through route behavior. | revisions test evidence | Medium | Supports delete/restore characterization before store refactor. |
| `internal/modules/revisions/phase7_history_helpers_test.go` | Test seed helpers for phase7 history rows, change sets, mutations, and revisions. | package-local test helpers. | revisions tests. | DB fixture SQL, uuid/json helpers. | phase7 history/rollback/integration tests. | none. | revisions test support | Low | In scope as target test inventory. |
| `internal/modules/revisions/phase7_history_test.go` | Phase7 history envelope, OpenAPI contract, history ref stability, and retained-history invariant coverage. | tests `TestPhase7_RecordHistoryEnvelope_U_7_01`, `TestPhase7_RecordHistoryOpenAPIContract_U_7_01`, `TestPhase7_HistoryEntryRefStability_U_7_02`, `TestPhase7_RetainedHistoryInvariants_U_7_07`. | Go test only. | generated OpenAPI contracts, phase4 harness, DB fixtures. | self. | `internal/gen/contracts` and `contracts/openapi/cartulary.openapi.yaml` indirectly. | revisions contract evidence | High | Primary characterization for history route and ref stability. |
| `internal/modules/revisions/phase7_integration_test.go` | Integration coverage for delete/restore/rollback atomicity, pagination binding, stale failure closure, retained history, and merge rollback. | tests `I-7-01` through `I-7-05`. | Go test only. | phase4 harness, WebSocket/test helpers, DB fixtures. | self. | phase7 ledgers and generated contracts indirectly. | revisions integration evidence | High | Key evidence for projection, collaboration, idempotency, and source-state behavior. |
| `internal/modules/revisions/phase7_locks_test.go` | Destructive-operation lock precedence coverage for rollback, restore, and merge. | `TestPhase7_DestructiveOperationLocks_U_7_06`. | Go test only. | DB locking fixtures, route helpers. | self. | error contract `record_locked` indirectly. | revisions destructive contention evidence | High | Lock ordering is observable behavior. |
| `internal/modules/revisions/phase7_rollback_test.go` | Rollback route selector, request validation, idempotency, history metadata, reason registry, rollback semantics, and helper fixture coverage. | `TestPhase7_RollbackSelectorUnion_U_7_05` plus extensive helpers. | Go test only. | phase4 harness, generated contracts, DB fixtures. | self. | OpenAPI and error registry contract expectations. | revisions rollback evidence | High | Primary characterization for rollback before any boundary split. |
| `internal/modules/revisions/rollback_api.go` | Strict rollback request decode, closed target union, normalization, and request hash. | `RollbackTarget`, `RollbackRequest`, `DecodeRollbackRequest`, `RollbackRequestHash`, `Normalized`. | `routes.go`, local tests. | `fieldnorm`, `httpapi`, JSON/hash/uuid. | `phase7_rollback_test.go`, `phase7_locks_test.go`, integration tests. | `RecordRollbackRequest`, target union, and errors in OpenAPI/errors contracts. | revisions route contract | High | Closed selector union is public API. |
| `internal/modules/revisions/rollback_mapping_guard_test.go` | Unit guards for timeline rollback changed-field keys, mention fallback key, and timeline source mapping. | package-local tests. | Go test only. | package-private rollback mapping helpers. | self. | none. | revisions rollback mapping evidence, likely timeline-owner seam | Medium | Indicates owner-specific mapping currently lives inside revisions. |
| `internal/modules/revisions/rollback_store.go` | Rollback transaction coordinator and implementation for history-entry, change-set, and row-restore targets across row-backed records, links, tags, mentions, aliases, preserved identifiers, projections, idempotency, locks, and history append. | rollback errors, `RollbackPreconditionError`, `RollbackResult`, `RollbackRecordChange`, `Store.RollbackRecord`; many private target/mapping helpers. | `routes.go`; tests; indirectly workbook/browser history workflows. | `authn`, `projectionadapters`, `workbook/collectionpolicy`, `links/revisionprovider`, raw SQL over many source-owner tables. | `phase7_rollback_test.go`, `phase7_integration_test.go`, `phase7_locks_test.go`, `rollback_mapping_guard_test.go`. | OpenAPI rollback schema, errors registry, phase ledgers, projection contracts. | revisions coordinator plus owner-specific rollback ports | Critical | Largest mixed-responsibility file; highest-risk boundary split. |
| `internal/modules/revisions/routes.go` | HTTP route registrar and service for record history, delete, restore, rollback, auth, membership role checks, pagination, session sliding, and collaboration publish. | `Service`, `RouteOptions`, `RouteOption`, `WithImportedAttributionResolver`, `RegisterRoutes`. | `internal/app/runtime.go`, route tests through server harness. | `collaboration`, `incidents`, `authn`, `httpapi`, `httpauth`, `pagination`. | phase7 route/integration tests. | OpenAPI route shapes, WS `record_changed` schema, auth/error contracts. | revisions route adapter with transport/platform boundary | High | Thin facade in shape, but it directly owns transport/session/collaboration concerns. |
| `internal/modules/revisions/store.go` | Core revisions store for change-set, mutation, revision insertions, history record lookup, history list materialization, history refs, rollback-action availability, and imported source actor attribution. | `Store`, `ImportedAttributionResolver`, `StoreOptions`, link/tag provider interfaces, `ChangeSetParams`, `MutationParams`, `RecordRevisionParams`, `RecordHistoryRecord`, `RecordHistoryItem`, `NewStore`, `NewStoreWithOptions`, insert/list methods. | Many modules: workbook, timeline, entities, imports, evidence, indicators, assessments, parties, tasks/decisions, artifacts, incident bundles. | `incidents`, `links/revisionprovider`, `postgres`, raw SQL over history tables and related mutation targets. | phase7 tests plus caller module tests. | OpenAPI history schema; incident bundle history substrate; schema ownership manifests. | revisions append/history facade | Critical | Legitimate core surface, but public store breadth should be narrowed behind semantic facades. |
| `internal/modules/revisions/workbook_conflicts.go` | Workbook patch conflict-window loading, same-field conflict payload building, collection conflict action application, deterministic local IDs, text merge suggestions, and conflict token issue helper. | `WorkbookPatchConflictWindow`, `WorkbookPatchChangedField`, `WorkbookPatchChange`, `WorkbookCollectionActionPayload`, `WorkbookCollectionAction`, `SameFieldConflictParams`, conflict helper functions. | `internal/modules/workbook`, `internal/modules/timeline`, workbook conflict tests. | `artifacts/riskrefs`, `links`, `workbook/collectionpolicy`, `viewschema`, raw SQL over `record_revisions`. | `workbook/phase6_conflict_support_test.go`, target conflict token tests, workbook/timeline route tests. | conflict route OpenAPI surfaces; generated UI selectors indirectly. | workbook/concurrency surface using revisions history substrate | High | Strong candidate to move or split behind workbook-owned conflict facade. |

## 3. Module Boundary Diagnosis

The current target is a mixed-responsibility package. It contains legitimate revision/history substrate, but should not be assumed to be a permanent module boundary in its current shape.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Change-set, mutation, and record-revision append substrate | `store.go` | `revisions` | keep | `InsertChangeSetTx`, `InsertMutationTx`, `InsertRecordRevisionTx`; many inbound module callers. | This is the most defensible core revisions responsibility. |
| Record history materialization and stable history refs | `store.go`, `routes.go` | `revisions` | keep | `GET /api/v1/records/{record_id}/history`, phase7 history tests, OpenAPI schema. | Preserve public item ordering, refs, and pagination binding. |
| Delete/restore orchestration | `delete_restore_store.go`, `delete_restore_api.go`, `routes.go` | split between `revisions` coordinator and source-owner ports | split | Adapter map and source tombstone updates cover many record types. | Coordinator belongs with revisions; source updates and owner preconditions should move behind owner ports. |
| Rollback orchestration | `rollback_store.go`, `rollback_api.go`, `routes.go` | split between `revisions` coordinator and owner rollback providers | split | Rollback handles history selectors, locks, idempotency, source updates, link/tag/entity mappings, projection rebuild. | Keep selector/idempotency/history semantics; move owner-specific source mutation and mapping behind ports. |
| Destructive-operation lock primitive | `delete_restore_store.go`, `rollback_store.go` | `revisions` or shared concurrency primitive | defer | `LockRecordEnvelopesNowaitTx` used by rollback/restore and entity merge adapter. | Needs owner decision whether destructive lock is revision-owned or shared source-state concurrency. |
| HTTP route/auth/session/pagination glue | `routes.go` | transport route adapter plus revisions facade | split | Direct use of `httpauth`, `authn`, `pagination`, `collaboration`. | Preserve route shapes; narrow service facade over revisions application logic. |
| Collaboration publish after mutations | `routes.go` | `collaboration` integration hook invoked by revisions route/app layer | split | `publishDeleteRestoreChange`, `publishRollbackChanges` emit `record_changed`. | Preserve event semantics and affected view behavior. |
| Projection rebuild | `delete_restore_store.go`, `rollback_store.go` | `projections` port | move | Direct `projectionadapters.NewRowProjector(nil).RebuildIncidentTx`. | Revisions should request refresh through a projection-owned port. |
| Workbook same-field conflict support | `workbook_conflicts.go`, `conflict_token.go` | workbook/concurrency, with revisions history query port | split | Workbook and timeline use conflict tokens and conflict helpers. | Keep token compatibility; move conflict payload rules out of revisions if feasible. |
| Incident bundle history provider | `incident_bundle_portability.go`, `attribution_resolver_registry.go` | incident portability adapter around revisions history substrate | keep or split | Incident bundle source includes revisions export/import and sequence repair. | Provider adapter is legitimate if explicitly retained as revisions-owned. |
| Link/tag rollback target operations | `store.go`, `rollback_store.go` with `links/revisionprovider` | `links` owner provider plus revisions coordinator | split | Link/tag provider interfaces already exist. | Pattern can guide other owner ports. |
| Entity mention, alias, preserved identifier rollback | `rollback_store.go` | `entities` owner provider | move | Direct SQL for entity-related mutation targets. | Candidate for owner port after characterization. |
| Timeline source mapping rollback | `rollback_store.go`, `rollback_mapping_guard_test.go` | `timeline` owner provider | move | Timeline-specific field-key/source-column mapping guarded in revisions tests. | High-value first extraction candidate. |
| Imports/tabular ingest revision appends | inbound callers in `imports`, `ownerfacade`, owner create paths | `imports` plus source owners using revisions append facade | keep revisions append facade | Imports call `revisions.NewStore().InsertChangeSetTx` through owner finalization paths. | Append facade should remain stable, but imports should not know rollback internals. |
| Frontend shell/controller state | none in target | `/apps/web` | defer | No frontend code under `internal/modules/revisions`. | Browser phase7 evidence exists, but no backend target file directly owns frontend shell. |
| Grid-vendor integration layer | none in target | `/packages/grid-adapter` | defer | No direct grid adapter/vendor imports found in target. | No action for this target. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/records/{record_id}/history` | Core 01/Core 02, implemented by `routes.go` and `store.go` | OpenAPI `getRecordHistory`; Core history route; phase7 ledger U-7-01/U-7-02/U-7-07. | `phase7_history_test.go`, `phase7_integration_test.go`, browser phase7 history rows. | Preserve or add tests before changing pagination/history materialization internals. | High | Includes newest-first ordering, tombstone row version, history refs, rollback action metadata. |
| `DELETE /api/v1/records/{record_id}` | Core 01/Core 04, implemented by `routes.go` and `delete_restore_store.go` | Route registration and Core delete/restore requirements. | `phase7_delete_restore_test.go`, `phase7_integration_test.go`. | Add owner-port characterization before extracting source adapter map. | High | Role gate is editor/reviewer/admin; preserve idempotency and conflict outcomes. |
| `POST /api/v1/records/{record_id}/restore` | Core 01/Core 04, implemented by `routes.go` and `delete_restore_store.go` | Route registration and destructive restore behavior. | `phase7_delete_restore_test.go`, `phase7_locks_test.go`, integration tests. | Add per-owner source-state characterization if moving tombstone updates. | High | Role gate is reviewer/admin; destructive lock precedence is observable. |
| `POST /api/v1/records/{record_id}/rollback` | Core 01/Core 02/Core 04, implemented by `routes.go`, `rollback_api.go`, `rollback_store.go` | OpenAPI `rollbackRecord`, errors registry, phase7 ledger U-7-05/I-7-01/I-7-03/I-7-05. | `phase7_rollback_test.go`, `phase7_locks_test.go`, `phase7_integration_test.go`, browser phase7 rows. | Required before moving owner-specific rollback source SQL. | Critical | Closed target union and append-only rollback history must not drift. |
| WebSocket `record_changed` after delete/restore/rollback | Core 01 collaboration contract plus route owner | `routes.go` publish helpers; `contracts/ws/index.schema.json`. | `phase7_integration_test.go`, browser phase7 rows. | Preserve event kind, affected views, row versions, and no-event failure cases. | High | Delete publishes remove; restore/rollback publish invalidate. |
| Workbook row patch conflict payloads and conflict tokens | Core 03 workbook conflict behavior, current helpers in revisions | `workbook_conflicts.go`, `conflict_token.go`, conflict route/OpenAPI surfaces. | `conflict_token_test.go`, `workbook/phase6_conflict_support_test.go`, workbook route tests. | Add or retain tests before moving helpers to workbook/concurrency. | High | Token byte shape is opaque but parse/validation compatibility is public. |
| Change-set/mutation/revision append API | Core 02 history substrate, current `Store` facade | `Store.InsertChangeSetTx`, `InsertMutationTx`, `InsertRecordRevisionTx`; many inbound modules. | Covered indirectly by owner module tests and phase7 tests. | Characterize caller expectations before narrowing `Store`. | Critical | Internal API, but broad caller surface raises migration risk. |
| History entry refs and rollback selectors | Core 01/Core 02 | `store.go`, `rollback_api.go`, OpenAPI schemas, domain `history_entry_ref`. | U-7-02, U-7-05, I-7-04. | Preserve exact stability and selector availability semantics. | Critical | Refs are opaque selectors, not storage IDs. |
| Projection refresh after delete/restore/rollback | Core 01 projections and source-state separation | Direct calls to projection adapter in stores. | I-7-01, I-7-03, I-7-05. | Needed before moving projection rebuild behind a port. | High | Failures must remain closed transactions. |
| Authorization and visibility outcomes | Core 04, route/service layer | `routes.go` role checks and membership rederivation. | Phase7 route tests; frontend phase9 evidence for destructive authorization. | Preserve 404/403/conflict distinctions before route facade refactor. | High | Feature visibility is not authorization. |
| Incident bundle revision history files | Core 01 incident portability | `incident_bundle_portability.go`, Core bundle registry. | Incident bundle tests and phase11 ledger. | Add direct characterization if adapter ownership changes. | High | File names and history completeness are portability contracts. |
| Generated OpenAPI and error registry surfaces | derived contracts | `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `internal/gen/contracts`. | Phase7 OpenAPI/error-registry tests. | No hand edits; regenerate only after owner-approved contract changes. | High | Generated roots are downstream. |
| View-schema history/rollback feature bindings | contracts/view-schemas and view-contracts package | `record_history_route`, `record_rollback_route`, `history.rollback`, `rollback_target_unavailable`. | `packages/view-contracts` tests; browser phase rows. | Needed only if feature binding ownership changes. | Medium | Treat as contract evidence, not runtime architecture. |
| Harness/test accounting | `docs/testing-harness-nlspec.md` and phase maps | `docs/testing/phase7_coverage_ledger.md`, `tools/phase7_test_map.json`. | Make phase targets. | Keep accounting updated if tests move or rows change. | Medium | Phase maps are evidence accounting only. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Direct source-owner SQL is embedded in rollback. | `rollback_store.go` updates and restores timeline, host, identity, indicator, evidence, assessment, party, task, decision, artifact, mention, alias, and preserved identifier data. | Source owner behavior can drift silently during revisions refactor. | must_fix | source owners via rollback ports | Define owner rollback provider ports and extract one owner at a time after characterization. |
| Delete/restore adapter map hardcodes source tables and view schemas. | `delete_restore_store.go` `deleteRestoreAdapters` and source tombstone methods. | New record types require revisions edits and may miss owner preconditions. | should_fix | source owners with revisions coordinator | Plan per-owner delete/restore source-state ports while preserving route contract. |
| Projection rebuild is called directly from revisions transactions. | `projectionadapters.NewRowProjector(nil).RebuildIncidentTx` in delete/restore and rollback paths. | Projection lifecycle details leak into mutation coordinator. | should_fix | projections | Introduce a projection refresh port with identical transactional behavior. |
| HTTP transport, auth, pagination, session sliding, and collaboration publish live in revisions route service. | `routes.go` imports `httpauth`, `authn`, `pagination`, `collaboration`. | Route facade can obscure application boundary and makes tests route-shaped. | should_fix | platform transport plus revisions app facade | Keep route adapter thin; move core commands behind application service if implementing. |
| Workbook conflict logic lives in revisions. | `workbook_conflicts.go` imports workbook collection policy, viewschema, links, artifacts risk refs. | Workbook behavior changes require revisions edits. | should_fix | workbook/concurrency with revisions history query port | Decide conflict helper ownership before moving file. |
| Conflict token codec is shared by workbook and timeline through revisions. | `conflict_token.go`; callers in workbook and timeline. | Token compatibility is public and cross-module. | defer | shared concurrency or workbook/revisions facade | Keep stable until a single owner and migration path are clear. |
| Link/tag rollback has partial owner-provider boundary. | `StoreOptions` link/tag rollback providers default to `links/revisionprovider`. | Pattern is good but incomplete compared with other owner data. | intentional/no_action | links plus revisions coordinator | Retain as pattern; use it for future owner ports. |
| Incident bundle portability adapter reads raw revisions history tables. | `incident_bundle_portability.go`. | Raw access is legitimate but can be mistaken for caller-owned access. | intentional/no_action | revisions provider adapter | Record retained provider exception; keep import/export file names stable. |
| Imported attribution sidecar access is intentionally indirect. | `attribution_resolver_registry.go`; `imported_attribution_boundary_test.go`. | Direct sidecar joins would violate portability boundary. | intentional/no_action | incident bundles attribution provider | Preserve guard when moving history materialization. |
| Store public surface is broad and used by many modules. | Inbound import search shows direct `NewStore`, insert methods, errors, lock helper, conflict helpers. | Narrowing surface can break many callers. | should_fix | revisions facade with caller-shaped ports | Sequence public-internal API cleanup after characterization. |
| Phase/test naming leaks into production test organization but not runtime. | phase7 test files and ledgers. | Refactor might confuse evidence accounting for architecture. | intentional/no_action | harness/accounting | Treat phase rows as evidence only; do not move runtime based on phase labels. |
| Generated contract surfaces are tested but not owner inputs. | Generated OpenAPI and error registry are referenced by tests. | Hand edits would create drift. | must_fix | contracts/generators | Never hand-edit generated roots; update owner inputs and regenerate only when authorized. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish authority, target, output path, repo state, and current tracker. | this tracker; framework; AGENTS; owner docs | `make lint-markdown` after tracker edit | Tracker has section 1 and current session log. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every target file and package surface. | all `internal/modules/revisions/*.go` | `rg --files internal/modules/revisions`; no validation claim | Section 2 complete for all 20 files. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map routes, generated contracts, WS events, history, rollback, delete/restore, conflicts, projections, auth. | revisions routes/API/store files; contracts; docs/testing | `make task-guide ROLE=feature-dev PHASE=phase7` before implementation | Section 4 names contracts and tests. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-06, WF-07 | Identify existing tests and gaps before movement. | phase7 tests; workbook/timeline conflict tests; browser phase ledgers | `make explain-phase PHASE=phase7` before implementation | Gaps are marked TODO or tied to slices. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05, WF-06 | Classify direct SQL, transport/platform, projection, generated, and owner coupling. | target files plus inbound callers | `make backend-module-boundary-check` when boundary manifests change | Section 5 findings classified. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-04 | WF-06 | Decide future facade/port shape without changing behavior. | `store.go`, `rollback_store.go`, `delete_restore_store.go`, `routes.go`, owner modules | targeted tests per owner before each extraction | Section 7 slices are behavior-preserving. |
| WF-06 | Slice sequencing plan | chain | WF-03, WF-05 | WF-07, WF-08 | Define smallest safe order for future refactor. | tracker; future code only under later authorization | no code validation for planning | Slice table has dependencies and rollback notes. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Keep phase maps and test rows as evidence accounting if tests move. | `docs/testing`, `tools/phase7_test_map.json`, Make targets if authorized later | `make json-shape-check`, phase ledger drift targets when touched | Handoff names accounting work separately from runtime architecture. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow docs validation for tracker and define later implementation gates. | this tracker | `make lint-markdown`; optional drift checks for docs-only | Session handoff and blockers are current. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/revisions-module-refactor-tracker.md` | None; documentation only. | None added. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` if feasible. | Revert only this tracker file. | Tracker exists with all required sections. |
| S-01 | S-00 | Add or refresh characterization for public history/delete/restore/rollback behavior before code movement. | `internal/modules/revisions/*_test.go`; phase7 ledgers only if test rows change. | Route envelopes, errors, idempotency, locks, refs, projection and WS semantics. | Preserve U-7-01 through U-7-07 and I-7-01 through I-7-05; add focused owner-port tests where gaps exist. | `make backend-store`; `make backend-integration`; `make phase-slice PHASE=phase7`. | Drop only added characterization if no production movement follows. | Existing behavior is pinned before extraction. |
| S-02 | S-01 | Define a narrow revisions append/history facade while preserving existing `Store` methods for callers during transition. | `store.go`; caller local ports in workbook, timeline, entities, imports, evidence, indicators, assessments, parties, tasks/decisions, artifacts. | Internal caller compatibility; history output must not drift. | Caller module tests plus phase7 history tests. | `make backend-store`; `make backend-module-boundary-check` if manifests change. | Retain old `Store` methods until all callers move. | New facade exists and old behavior passes. |
| S-03 | S-01 | Split transport route adapter from revisions command service without changing route registration or envelopes. | `routes.go`, API helper files, new internal service file if authorized. | HTTP route shapes, auth outcomes, session slide, pagination, collaboration publish. | Phase7 route/integration tests and auth/visibility cases. | `make backend-integration`; `make phase-slice PHASE=phase7`. | Revert route facade extraction only. | Route handlers delegate to a narrower command facade with identical responses. |
| S-04 | S-01, S-02 | Extract owner-specific rollback source mutation behind ports, one owner at a time. | `rollback_store.go`; first candidate `timeline`, then entities/links/evidence/indicators/assessments/tasks/parties/artifacts. | Rollback affected records, changed fields, projections, history append, stale target errors. | Existing phase7 rollback/integration tests plus owner-specific characterization for each moved owner. | `make phase-slice PHASE=phase7`; owner phase slice as needed. | Move one owner per commit/slice; retain old path until tests pass. | One owner no longer has direct source rollback SQL in revisions. |
| S-05 | S-01, S-02 | Extract delete/restore source-state adapter behavior behind owner ports. | `delete_restore_store.go`; source owner packages. | Delete/restore preconditions, tombstone behavior, projections, row version, idempotency. | Phase7 delete/restore tests; per-owner adapter matrix coverage. | `make backend-store`; `make backend-integration`; `make phase-slice PHASE=phase7`. | Keep adapter map as fallback during port migration. | At least one owner delete/restore path is provider-backed. |
| S-06 | S-01 | Move workbook same-field conflict logic behind workbook/concurrency ownership or a narrow revisions history query port. | `workbook_conflicts.go`, `conflict_token.go`, workbook/timeline callers. | Conflict token compatibility and conflict payload shape. | `conflict_token_test.go`, `workbook/phase6_conflict_support_test.go`, workbook/timeline conflict route tests. | `make backend-store`; `make backend-integration`; `make frontend-unit` if frontend conflict handling changes. | Keep token codec stable; do not change token version without owner approval. | Workbook conflict behavior no longer depends on broad revisions helpers. |
| S-07 | S-02 | Isolate projection rebuild calls behind a projections-owned port. | `delete_restore_store.go`, `rollback_store.go`, projection adapters. | Projection refresh timing and transaction closure. | I-7-01/I-7-03/I-7-05. | `make backend-integration`; `make phase-slice PHASE=phase7`. | Restore direct adapter call if port changes transactional semantics. | Revisions invokes a projection port with identical rebuild behavior. |
| S-08 | S-02 | Clarify incident portability adapter ownership while preserving file names and imported attribution behavior. | `incident_bundle_portability.go`, `attribution_resolver_registry.go`, `incidentbundles/source.go`. | Bundle completeness, imported source actor attribution, sequence repair. | Incident bundle integration tests; `imported_attribution_boundary_test.go`. | `make backend-integration`; phase11 bundle targets if discovered; `make json-shape-check` if manifests change. | Keep current provider adapter if ownership is unclear. | Portability responsibilities are documented or moved without file-contract drift. |
| S-09 | S-03 through S-08 | Remove obsolete wrappers and update boundary guardrails after callers move. | target files; `tools/backend_module_boundaries.json` only if authorized. | Generated and harness drift if manifests change. | Boundary checker coverage. | `make backend-module-boundary-check`; `make json-shape-check`; `make generated-artifact-policy-check`. | Revert guardrail updates if they block legitimate retained calls. | No stale internal surface remains without an explicit retained exception. |

Any slice that changes public behavior, route shapes, generated contract shapes, incident bundle file names, or retained history semantics requires later authorization before implementation.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-store`; `make test-fast` for broader local loop | backend store/unit rows including phase7 store tests | yes for code refactor; no for tracker-only docs | `backend-store` is present in public task surface; choose the narrower owner/phase target after `make task-guide`. |
| integration | `make backend-integration`; `make phase-slice PHASE=phase7`; `make service-backed-slice PHASE=phase7` | route, DB, projection, WS, idempotency, rollback/delete/restore behavior | yes for rollback/delete/restore route or store movement | Discover exact row selection with `make task-guide ROLE=feature-dev PHASE=phase7` and `make explain-phase PHASE=phase7`. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make frontend-unit` | workbook history, rollback preview/action, destructive confirmation, conflict handling | no for backend-only internals unless route/UI behavior or conflict UI changes | Phase/browser rows are evidence accounting, not runtime architecture. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift` when owner inputs or generated outputs are touched | generated roots, JSON manifests, generated contract drift | yes when contracts/manifests/generated inputs change; docs-only tracker should run policy/shape if feasible | Do not hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make frontend-import-boundary-check` if frontend packages change | backend module ownership and frontend import boundaries | yes when moving packages, ports, or boundary manifests | No frontend target code is in scope for tracker-only work. |
| full check | `make check` | developer verification gate | no before planning; yes before high-risk or final implementation handoff when broad risk warrants | Run after narrow commands pass or when cross-module risk requires it. |

Tracker-only validation recommendation:

- `make lint-markdown`
- `make generated-artifact-policy-check`
- `make json-shape-check`
- `make agent-finalize` before broader end-of-run verification; if no `RESULTS_DIR` is supplied, report retained-run maintenance as skipped.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| RT-001 | Create target-specific tracker for `internal/modules/revisions`. | WF-00 | DONE | none | this file | Tracker exists at required output path. |
| RT-002 | Confirm target path and enumerate every target file. | WF-01 | DONE | RT-001 | `rg --files internal/modules/revisions`; Section 2 | All 20 target files inventoried. |
| RT-003 | Record source hierarchy and owner docs inspected. | WF-00 | DONE | RT-001 | Section 1 | Authority order and owner docs are explicit. |
| RT-004 | Diagnose whether `revisions` is a clean boundary. | WF-04 | DONE | RT-002 | Sections 3 and 5 | Mixed-responsibility findings are classified. |
| RT-005 | Map public behavior freeze surfaces. | WF-02 | DONE | RT-002 | Section 4 | Every discovered public contract risk has owner/test posture. |
| RT-006 | Plan characterization before implementation. | WF-03 | DONE | RT-005 | Sections 6 and 7 | Risky movement is gated by existing or planned tests. |
| RT-007 | Plan behavior-preserving implementation slices. | WF-06 | DONE | RT-004, RT-006 | Section 7 | Slices have dependencies, validations, rollback notes, and completion criteria. |
| RT-008 | Record validation command posture. | WF-08 | DONE | RT-007 | Section 8 | Commands are Make-owned or marked as discovery/conditional. |
| RT-009 | Append session handoff logs. | WF-08 | DONE | RT-008 | Section 10 | Workstream-specific handoff rows are present. |
| RT-010 | Execute source-owner rollback port extraction. | WF-05 | PARTIAL | RT-006 | Timeline provider extracted under `internal/modules/timeline/rollbackprovider`; remaining owners still staged | Direct Timeline rollback SQL removed from revisions first; other source-owner SQL should move one owner at a time. |
| RT-011 | Execute workbook conflict ownership split. | WF-05 | DONE | RT-006 | `internal/modules/workbook/conflicts`; phase6 validation | Conflict token codec ownership is workbook-aligned without v2 token drift. |
| RT-012 | Execute projection port isolation. | WF-05 | DONE | RT-006 | `internal/modules/revisions/projection_port.go`; phase7 validation | Revisions no longer constructs projection adapters directly from rollback/delete/restore paths. |
| RT-013 | Run tracker-only validation. | WF-08 | DONE | RT-009 | `make lint-markdown` pass; `make generated-artifact-policy-check` pass at `.cartulary/test-results/20260709T212646Z-p33678`; `make json-shape-check` pass at `.cartulary/test-results/20260709T212650Z-p33852` | Historical planning validation; `make agent-finalize` skipped due tracker-only write constraint. |
| RT-014 | Run implementation validation. | WF-08 | DONE | RT-010, RT-011, RT-012 | Section 1 implementation validation list and Section 10 handoff row | Phase slices, boundary check, `test-fast`, and `agent-finalize` completed for this remediation pass. |
| RT-015 | Execute delete/restore source-owner provider extraction. | WF-05 | DONE | RT-006 | Provider packages under source owners plus `internal/modules/records/deleterestore`; phase7/phase3/phase4/phase5/phase9 validation | Revisions delete/restore no longer owns the source table adapter map or party delete precondition SQL. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Tracker created for `revisions`; planning-only scope recorded. | Inspected framework, AGENTS, Core 00-04, domain, harness spec; touched only this tracker. | `sed`, `rg`, `git status`, `git branch --show-current`, `git rev-parse --short HEAD`, `make help`, `make lint-markdown`, `make generated-artifact-policy-check`, `make json-shape-check`. | Target exists; output tracker was absent before this session; tracker-only validation passed. | No owner contradiction found; `make agent-finalize` skipped because only tracker writes are permitted. | Later authorized implementation can start with S-01 characterization. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Backend target is mixed-responsibility: legitimate revisions substrate plus transport, projection, workbook conflict, portability, and source-owner coupling. | Inspected all `internal/modules/revisions/*.go`; touched only this tracker. | `rg --files`; `rg -n '^func |^type |^const |^var '`; targeted `sed` of production files. | Section 2 inventories all files and Section 3 records ownership candidates. | Future source-owner port design still needs implementation authorization. | Start with characterization and one-owner extraction, likely timeline rollback mapping. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | No direct frontend shell or grid-vendor integration exists in the backend target. Frontend/browser evidence exists for history and rollback behavior. | Inspected contract and phase evidence references; touched only this tracker. | Targeted `rg` over `packages/view-contracts`, `packages/ui-contracts`, `docs/testing`, and `tools`. | Frontend changes are not part of tracker-only scope. | Future conflict or history UI changes would require frontend target validation. | Keep frontend validation conditional on route/UI behavior changes. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Generated OpenAPI, error registry, WS schema, view-contracts, and ui-contracts surfaces were identified as downstream evidence. | Inspected contract search output; touched only this tracker. | Targeted `rg` over `contracts`, `internal/gen`, `packages/view-contracts`, `packages/ui-contracts`. | No generated files edited or proposed for hand edits. | Any public contract change needs owner-doc authorization first. | Use `make generated-artifact-policy-check` and `make json-shape-check` for tracker/doc drift posture. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Phase7 store/integration/browser rows and local unit tests cover history, delete/restore, rollback, locks, OpenAPI/errors, conflict tokens, and imported attribution boundary. | Inspected `internal/modules/revisions/*_test.go`, `docs/testing/phase7_coverage_ledger.md`, `tools/phase7_test_map.json`; touched only this tracker. | `rg -n '^func Test'`; targeted `rg` over docs/testing and tools; `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`. | Existing evidence is strong for route/store behavior; tracker-only validation passed. | Exact future row selection should be discovered before implementation. | Run `make task-guide ROLE=feature-dev PHASE=phase7` and `make explain-phase PHASE=phase7` before code refactor. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Routes rederive membership and role gates in `routes.go`; Core 04 keeps feature visibility separate from authorization. | Inspected `routes.go`, Core 04, phase7 tests; touched only this tracker. | `sed` and targeted `rg`. | Auth outcomes are listed as frozen behavior. | Route facade refactor must preserve 404/403/conflict distinctions. | Add focused auth characterization before moving route command boundaries. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-09T21:21:31Z | Codex | Open risks are source-owner rollback SQL, workbook conflict ownership, projection refresh coupling, and broad `Store` surface. | Touched only this tracker. | Discovery commands and tracker-only validation listed in Section 1. | Tracker validation passed; no production refactor has been performed. | `make agent-finalize` skipped because it may mutate non-tracker harness artifacts. | Later authorized task can start S-01. |
| 2026-07-09T22:00:55Z | Codex | Initial remediation implementation started: conflict-token ownership moved to workbook conflict package, destructive lock SQL moved to records, Timeline rollback provider seam introduced, projection rebuild calls moved behind a revisions store port, and revisions routes now call a command facade. | Touched specs, this tracker, `internal/modules/workbook/conflicts`, `internal/modules/records`, `internal/modules/revisions`, `internal/modules/timeline`, and entity merge ports. | `make backend-unit` initially failed on an unused import, then passed at `.cartulary/test-results/20260709T220055Z-p49805`. | Backend unit coverage passed after the first shared-primitive changes. | Remaining owner-provider extraction is partial: non-Timeline rollback providers and delete/restore source providers still need owner-by-owner migration. | Continue with phase7/phase6/phase3 validation, then migrate additional source owners one at a time. |
| 2026-07-09T22:13:05Z | Codex | First remediation pass completed and validated: specs clarify ownership, conflict tokens are workbook-owned, destructive locks are records-owned, Timeline rollback provider is extracted, projection rebuild is behind a revisions port, and route handlers delegate to a command facade. | Same implementation files as previous row plus final tracker update. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; phase slices for phase7, phase6, phase3, phase4, and phase11; `make backend-module-boundary-check`; `make test-fast`; `make agent-finalize`. | Validation passed; run roots are listed in Section 1. Retained-run `agent-finalize` failed preflight only because the supplied `test-fast` run is not qualifying full warm-check evidence, then plain `agent-finalize` passed. | Non-Timeline rollback providers and delete/restore source providers remained staged at this checkpoint. | Continue owner-by-owner provider extraction, starting with row-backed owners only after adding provider contract tests. |
| 2026-07-09T22:28:50Z | Codex | Delete/restore source-provider extraction completed: source snapshots, source tombstone handling, view-schema lookup, and party delete blockers moved out of revisions into owner provider packages. Rollback now reuses the provider registry for source snapshots/view schemas where it still coordinates non-Timeline rollback behavior. | Added `internal/modules/records/deleterestore` plus source-owner provider packages under assessments, artifacts, entities, evidence, indicators, parties, tasksdecisions, and timeline; updated revisions delete/restore and rollback stores. | `make backend-unit`; `make phase-slice PHASE=phase7`; `make phase-slice PHASE=phase3`; `make phase-slice PHASE=phase4`; `make phase-slice PHASE=phase5`; `make phase-slice PHASE=phase9`; `make backend-module-boundary-check`; `make test-fast`; `make phase-slice PHASE=phase11`. | All listed commands passed; run roots are listed in Section 1. | Remaining owner-provider extraction is now focused on non-Timeline rollback source-state reconstruction/application and relationship/non-row mutation providers. | Continue by extracting row-backed rollback providers from `rollback_store.go` one owner at a time, then relationship/non-row mutation providers. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Which source owner should be extracted first from `rollback_store.go`? | Extraction order affects risk; timeline has explicit guard tests, while entities/links/evidence have cross-record effects. | Existing phase coverage plus owner-specific characterization. | DECIDED: Timeline first; initial Timeline rollback source provider extracted under `internal/modules/timeline/rollbackprovider`. |
| RB-002 | Should conflict token ownership remain in revisions or move with workbook/timeline conflict handling? | Token compatibility crosses workbook and timeline routes. | Core 03 conflict behavior plus current workbook/timeline tests. | DECIDED: moved to workbook conflict ownership in `internal/modules/workbook/conflicts`, preserving v2 token compatibility. |
| RB-003 | Should destructive-operation locking remain revisions-owned or move to a shared concurrency primitive? | Merge already uses the lock helper, so ownership is broader than rollback alone. | Core 01 destructive-operation family and entity merge tests. | DECIDED: moved SQL lock primitive to `internal/modules/records` as a record-envelope concurrency primitive. |
| RB-004 | Should incident bundle revision import/export remain in revisions as an owner-provider adapter? | Bundle file names and history completeness are portability contracts. | Core incident bundle section and incident bundle integration tests. | DECIDED: retained as revisions/history owner-provider behavior; Core 01 now names this ownership boundary. |
| RB-005 | What exact Make rows should be used for each future owner extraction? | Broad runs are expensive and phase maps are evidence accounting only. | `make task-guide ROLE=feature-dev PHASE=phase7`, `make explain-phase PHASE=phase7`, and target-plan inspection. | DECIDED for current pass: used `make phase-slice PHASE=phase7`, phase3, phase4, phase5, phase6, phase9, and phase11 coverage as listed in Section 1. Rediscover before the next owner extraction because phase maps may change. |

No `BLOCKED: owner contradiction` item is currently recorded.

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| every file in `internal/modules/revisions` is inventoried or explicitly out of scope | DONE | Section 2 preserves the original 20-file baseline, notes moved conflict-token files, and Section 10 records new remediation files. |
| every discovered public contract risk has an owner and test posture | DONE | Section 4 maps routes, WS events, workbook conflicts, projections, auth, generated contracts, incident bundles, and harness accounting. |
| every proposed workflow has dependencies and exit criteria | DONE | Section 6 lists WF-00 through WF-08 with dependencies and handoff checkpoints. |
| every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization` | DONE | Section 7 states behavior-preserving default and later authorization rule. |
| validation commands are discovered or marked `TODO` with a reason | DONE | Section 8 lists Make-owned commands and discovery requirements. |
| contradictions are marked `BLOCKED: owner contradiction` | DONE | No owner contradiction found; Section 11 states none recorded. |
| repository/framework mismatches are recorded as planning findings | DONE | Section 1 records the mismatch between framework catalog and live package responsibilities. |
| handoff sections are current enough for another agent to continue without rediscovery | DONE | Section 10 records planning state, implementation changes, commands, blockers, and next actions. |

This tracker is now an implementation handoff. The first structural remediation pass and delete/restore provider extraction are validated, and the remaining work is scoped to owner-by-owner rollback provider extraction beyond Timeline.

## 13. 2026-07-09 Next Iteration: Rollback Ownership and Facade Closure

### 13.1 Iteration authority, scope, and live posture

This section is the implementation-ready handoff for the next Revisions refactor iteration. It preserves the earlier tracker as history and trusts the live source tree at commit `249669af` where the earlier baseline or proposed slices are stale.

Tracker-only scope for this session:

- update only this handoff file;
- do not change runtime code, specs, migrations, authored manifests, generated files, or dependency artifacts;
- record required future implementation work without representing it as already complete;
- carry forward public behavior only where Core 00 through Core 04 or another durable owner contract requires it.

Sources rechecked for this iteration:

- `docs/domain.md`, especially Revisions and Audit vocabulary, record-envelope membership, relationship families, and the source-state-versus-workbook distinction;
- Core 00 through Core 04, especially Core 01 section 3.3.5.0, Core 02 sections 14 through 15, Core 03 history/conflict behavior, and Core 04 rollback, collaboration, and conflict conformance;
- `docs/testing-harness-nlspec.md`, `make task-guide ROLE=feature-dev PHASE=phase7`, `make explain-phase PHASE=phase7`, and task-guide discovery for phases 3, 4, 5, 6, 8, 9, and 11;
- all production and test files under `internal/modules/revisions`;
- current provider and owner surfaces under `internal/modules/records`, `timeline`, `entities`, `links`, `indicators`, `evidence`, `assessments`, `parties`, `tasksdecisions`, `artifacts`, `workbook`, `projections`, `incidentbundles`, and `incidentportability`;
- current OpenAPI, error-registry, WebSocket, schema-ownership, phase-map, and backend-boundary inputs.

Current live posture:

| Area | Live state | Iteration disposition |
| --- | --- | --- |
| conflict token | Codec and tests are owned by `internal/modules/workbook/conflicts`; token version remains `2`. | No ownership move. Preserve v2 compatibility as a regression gate. |
| destructive lock | Fail-fast canonical record-envelope locking is owned by `internal/modules/records`. Revisions retains only local error adaptation. | No ownership move. Preserve Core 01 lock ordering and evaluation precedence. |
| Timeline rollback source | `internal/modules/timeline/rollbackprovider` owns Timeline source decoding, update, and touch behavior. | No re-extraction. Use its seam as the first row-provider pattern, then remove duplicated Revisions mapping coverage. |
| delete/restore sources | Source snapshots, tombstone behavior, view-schema selection, and Party delete preconditions are in source-owner `deleterestore` packages. | No source move. Later replace the Revisions global catalog with explicit application composition. |
| route commands | Revisions routes delegate history, delete, restore, and rollback operations through `CommandService`. | Retain. Further transport/auth/session splitting is deferred. |
| projections | Delete, restore, and rollback invoke `ProjectionRebuilder` rather than constructing a projector at each call site. | Retain port semantics. Later remove the default adapter fallback when application assembly supplies it. |
| incident history bundle | Revisions owns the three history-substrate bundle files and imported sequence repair. | No action; this is explicit Core 01 owner-provider behavior. |
| remaining rollback source behavior | `rollback_store.go` still decodes and applies Host, Identity, Party, Indicator, Evidence, Assessment, Task Request, Decision, and Artifact source state. | Must move owner by owner. |
| remaining non-row behavior | Links and Tags are partially provider-backed; Entity mention, alias, and preserved-identifier behavior remains in Revisions. Indicator observation and lifecycle-interval mutations are emitted but not visible/addressable through current rollback dispatch. | Complete source-owner providers and close the unsupported Indicator target gap. |
| broad facade | `Store` still combines append-only history writes, history reads, rollback/delete/restore commands, workbook conflict queries, attribution, target providers, and projection dependencies. | Narrow after provider extraction. |

Live-code mismatches with earlier sections are intentional historical differences, not contradictions to be rewritten:

- Section 2 describes direct delete/restore source-table adapters in Revisions; those adapters have moved to source-owner packages.
- Sections 3, 5, and 7 leave conflict tokens, destructive locking, projection isolation, route command delegation, and delete/restore providers undecided or staged; all are implemented now.
- Section 7 treats Timeline as the first future rollback extraction; Timeline extraction is complete.
- Section 9 marks `RT-015` complete, which matches the live tree, while earlier inventory prose still describes the pre-extraction implementation.
- The live tree reveals an additional gap not closed by the earlier tracker: `internal/modules/indicators` emits `indicator_observation` and `indicator_state_interval` mutation targets, but Revisions history visibility, single-entry addressing, protected-set calculation, validation, and apply dispatch do not recognize them.

### 13.2 Durable ownership boundary

Core 02 requires the history substrate to own stable selectors and reversible mutation accounting while explicitly forbidding it from owning source-field vocabulary or source-table reconstruction rules for every record family. Apply that split as follows.

| Revisions retains | Source owners retain | Application assembly retains |
| --- | --- | --- |
| request normalization; history and selector lookup; `history_entry_ref` allocation and stability; target visibility; change-set ordering; later-mutation checks; idempotency; transaction boundaries; incident-open checks; shared destructive-lock invocation; plan-level preconditions; record-envelope row-version advancement; inverse change-set, mutation, and record-revision append; projection-port invocation; response and collaboration result data | source-value decoding; record-type and target-kind field vocabulary; source-table update/touch behavior; non-row identity validation; source mutation; owner stale/not-found/not-reversible detection; affected first-class record calculation; owner-specific changed-field keys; target-specific atomic companion requirements | construction of delete/restore, row rollback, non-row rollback, projection, and attribution providers; mapping exact record types or mutation target kinds to providers; failure on duplicate or missing required registrations |

Provider rules for every extraction:

- A provider receives an existing transaction and never begins, commits, or rolls it back.
- A provider does not append `change_sets`, `change_set_mutations`, or `record_revisions`.
- A provider does not advance `records.row_version`, rebuild projections, publish WebSocket events, or construct HTTP errors.
- Owner errors normalize to target-not-found, stale-target, or target-not-reversible semantics at the Revisions boundary.
- Provider registration is explicit. Do not use package `init`, mutable global self-registration, or a silent default/fallback path.
- Move one owner or target family at a time and delete its old Revisions switch branch and helper functions in the same slice.

Minimum future provider contracts:

| Contract | Required behavior | Result consumed by Revisions |
| --- | --- | --- |
| row source rollback provider | Decode a retained history value into owner source state; restore that state; touch owner state when a non-row mutation advances the containing first-class record. | Success or normalized owner error. Revisions supplies record ID, actor, time, and next row version. |
| non-row mutation provider | Validate an owner mutation target; calculate affected first-class records and target-specific whole-change-set requirements; apply the inverse source mutation; load canonical before/after values. | Canonical inverse before/after values plus affected record IDs and exact changed-field keys per affected record. |
| existing delete/restore source provider | Snapshot source plus envelope, apply delete state, touch source, resolve view schema, and validate owner delete preconditions. | Existing behavior; only catalog construction moves out of Revisions. |

Use a cycle-free contract package such as `internal/modules/revisions/rollbackcontract` for shared provider DTOs and normalized provider errors. Owner packages may import that contract package; the root Revisions package and `internal/app` may also import it. The contract must not contain owner field mappings or SQL.

### 13.3 Remaining `rollback_store.go` responsibility inventory

| Current responsibility | Current symbols or area | Correct owner | Action |
| --- | --- | --- | --- |
| rollback transaction, idempotency, incident-open check, lock ordering, addressed-row re-read, row-version check, commit | `Store.RollbackRecord`, `loadRollbackProtectedSetTx` | Revisions plus shared `records` lock primitive | Keep in Revisions. |
| history-entry, change-set, and row-restore plan loading | `loadHistoryEntryRollbackPlanTx`, `loadChangeSetRollbackPlanTx`, `loadRowRestorePlanTx` | Revisions | Keep history-table queries and plan ordering in Revisions; ask providers for owner target effects. |
| later-mutation and isolated-reversal eligibility | `ensureNoLaterRollbackPlanMutationTx`, `ensureNoLaterRollbackTargetMutationTx`, `historyEntryRequiresChangeSetTx` | Revisions policy with owner target metadata | Keep generic history checks; replace hard-coded attached-evidence/target rules with provider metadata. |
| inverse history append, first-class row versions, revisions, projection call | `applyRollbackPlanTx`, `applyChangeSetRollbackPlanTx`, `insertRollbackMutationTx`, record-revision helpers | Revisions | Keep. Remove target-family source switches after provider dispatch exists. |
| record-envelope load and update | `loadRollbackRecordEnvelopeTx`, `updateRollbackRecordEnvelopeTx` | `records` primitive consumed by Revisions | Keep orchestration local for now; do not move source rules back into `records`. |
| Host source decode/update/touch | `updateHostFromRollbackSourceTx`, Host cases in `rollbackSourceForRecordType`, `directRollbackSourceForRecordType`, and touch switch | `internal/modules/entities/rollbackprovider` | Must extract first. |
| Party source decode/update | Party mapping plus `updateGenericWorkbookSourceTx` table/column list | `internal/modules/parties/rollbackprovider` | Must extract second; eliminate blanket nil assignment for absent fields. |
| Identity source decode/update/touch | `updateIdentityFromRollbackSourceTx`, Identity mapping/direct/touch cases | `internal/modules/entities/rollbackprovider` | Must extract after Identity characterization. |
| Evidence source decode/update | `updateEvidenceFromRollbackSourceTx` and Evidence mapping/direct cases | `internal/modules/evidence/rollbackprovider` | Must extract; protect blob and association semantics. |
| Indicator row decode/update/touch | `updateIndicatorFromRollbackSourceTx` and Indicator mapping/direct/touch cases | `internal/modules/indicators/rollbackprovider` | Must extract before Indicator child targets. |
| Assessment source decode/update | `updateAssessmentFromRollbackSourceTx` and Assessment mapping/direct cases | `internal/modules/assessments/rollbackprovider` | Must extract; retain merge-subject behavior. |
| Task Request and Decision source decode/update | mappings and `updateGenericWorkbookSourceTx` table/column lists | `internal/modules/tasksdecisions/rollbackprovider` | Must extract together after lifecycle characterization. |
| Artifact source decode/update | Artifact mapping and broad `updateGenericWorkbookSourceTx` column list | `internal/modules/artifacts/rollbackprovider` | Must extract last among row providers because behavior is variant-sensitive. |
| generic dynamic-table rollback writer | `updateGenericWorkbookSourceTx`, `joinSQLAssignments` | none | Remove. Each owner must use explicit owner SQL and preserve fields not represented by the rollback value. |
| Link and Tag validation/source mutation | `LinkRollbackTargetProvider`, `TagRollbackTargetProvider`, wrapper helpers | `internal/modules/links/revisionprovider` | Provider-backed today; extend rather than duplicate. |
| Link and Tag affected records/changed keys | `affectedRecordsForRollbackTarget`, `rollbackRecordLinkChangedFieldKeysTx`, `rollbackRecordTagChangedFieldKeysTx`, `collectionpolicy` dependency | Links owner provider | Move. Revisions consumes provider effects. |
| Entity mention load/restore and field key | mention load/restore helpers and `rollbackMentionFieldKey` | `internal/modules/entities/rollbackprovider` | Move source behavior and changed-field calculation; Revisions retains companion change-set coordination. |
| Entity alias and preserved identifier identity/load/tombstone | associated identity structs and helpers | `internal/modules/entities/rollbackprovider` | Move. These remain whole-change-set targets for merge reversal. |
| Indicator observation and lifecycle interval | absent from history/rollback switches despite owner-emitted mutation entries | `internal/modules/indicators/rollbackprovider` | Add owner provider behavior and Revisions provider dispatch; do not add new Indicator SQL to Revisions. |
| target-kind validation and affected-record parsing | `validateRollbackTarget`, `rollbackRecordTypeForTarget`, `affectedRecordsForRollbackTarget(s)`, `firstClassRollbackTargetKind` | Revisions for common row targets; owner providers for non-row and source-specific semantics | Replace closed switches with registry dispatch while retaining the public request target union. |
| payload decode/equality/sorting helpers | history JSON decode, idempotency payload decode, canonical changed-key and affected-ID sorting | Revisions | Keep only generic helpers. Move source-field extraction helpers with providers. |

### 13.4 Public contract freeze

No implementation slice in this iteration is authorized to change these contracts. If an implementation appears to require a change, stop and obtain owner-spec authority rather than widening the refactor.

| Contract | Stable requirements | Required evidence |
| --- | --- | --- |
| history HTTP route | `GET /api/v1/records/{record_id}/history`, common success/paging envelopes, newest-first logical history, tombstone row version, structured rollback actions, authorization and visibility distinctions | Phase7 history unit/integration/browser rows and OpenAPI artifact checks. |
| delete HTTP route | `DELETE /api/v1/records/{record_id}`, existing request normalization, role gate, idempotency-before-fresh-state behavior, success envelope, delete blockers, and soft-delete history | Phase7 delete/restore and integration rows. |
| restore HTTP route | `POST /api/v1/records/{record_id}/restore`, reviewer/admin gate, protected set of the target record, lock-before-fresh-state precedence, success envelope, append-only restore history | Phase7 delete/restore, lock, and integration rows. |
| rollback HTTP route | `POST /api/v1/records/{record_id}/rollback`, closed request target union `history_entry`, `change_set`, and `row_restore`; normalized reason; idempotency; canonical affected IDs; append-only reversal | Phase7 rollback, lock, integration, and browser rows. |
| error vocabulary | Preserve `invalid_mutation_payload`, `invalid_rollback_request`, `client_txn_conflict`, `row_version_conflict`, `record_locked`, `record_deleted_use_restore`, `record_already_deleted`, `record_delete_blocked`, `record_not_deleted`, `rollback_target_not_found`, and `rollback_precondition_failed`. Preserve current `rollback_precondition_failed` reason codes, including `target_not_reversible`, `entry_requires_change_set`, `dependent_later_changes`, and `stale_target`. | Generated error registry plus route tests. New owner errors must map into this vocabulary. |
| WebSocket | Preserve `record_changed` v1, required payload members, row versions, change-set and actor attribution, canonical unique `changed_field_keys[]`, canonical unique non-empty `affected_views[]`, and `patch`/`invalidate`/`remove` meanings. Do not add a rollback-only event. | Phase7 integration/browser rows and `contracts/ws/index.schema.json`. |
| history selector | Once issued, `history_entry_ref` remains stable and attached to the same logical item for the retained-history lifetime in that deployment. Current ineligibility is expressed through action metadata and existing errors, not ref deletion/reassignment. Portability import may reissue refs, after which the same stability rule applies. | U-7-02, retained-history tests, phase11 portability coverage. |
| bundle v1 files | Preserve `data/change_sets.ndjson`, `data/change_set_mutations.ndjson`, and `data/record_revisions.ndjson`; keep Revisions as their owner provider. | Phase11 incident-portability integration and bundle registry checks. |
| conflict token | Preserve opaque conflict-token v2 issue/parse compatibility, HMAC purpose scope, claim binding, request-hash validation, current conflict-window validation, and stale-token behavior. Clients never parse or mint tokens. | Workbook conflict-token tests and phase6 conflict slices. |

No generated contract change is expected. Do not edit `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, generated phase ledgers, generated schedules, `go.sum`, or `pnpm-lock.yaml` by hand.

### 13.5 Gap classification and remediation ledger

| ID | Classification | Remediation and owner | Affected areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| NI-01 | must fix now | Add row source providers for Host, Party, Identity, Evidence, Indicator, Assessment, Task Request, Decision, and Artifact. Delete the corresponding Revisions source mapping/update/touch branch in each slice. | implementation, tests, documentation | Core 02 assigns source vocabulary and reconstruction to source owners. Owner providers make new record families extensible without growing a central switch. | Internal Go movement only; no public contract or database migration expected. Do not retain compatibility fallbacks. | Revisions remains coupled to every source schema; partial values can clear unrelated owner fields; future record growth requires central edits. | No listed owner table, column list, or field mapping remains in `rollback_store.go`; owner unit tests and phase7 plus owner phase slices pass. |
| NI-02 | must fix now | Add `indicator_observation` and `indicator_state_interval` provider support, history visibility, provider dispatch, protected-set effects, and append-only inverse semantics. Add authored tombstone columns through a future migration and make owner reads/projections exclude tombstoned children. | spec conformance, schema, implementation, tests, documentation | These mutation kinds are emitted today but are not visible/addressable in Revisions. Core 02 requires target-granular reversible accounting while preserving observations and append-only lifecycle history. | Requires authored migration and owner query updates; bundle file names and public APIs remain unchanged. Resolution rollback restores prior values; create rollback tombstones rather than hard-deletes. | Whole change-set rollback can fail for an otherwise owner-emitted change set, and Indicator history is incomplete. Hard deletion would violate preservation goals. | Observation create/resolve and lifecycle create appear in the correct record history; change-set rollback is atomic, preserves source text/history, updates affected projections/row versions, emits ordinary `record_changed`, and passes migration drift. |
| NI-03 | must fix now | Move Entity mention, alias, and preserved-identifier validation/load/apply behavior into `internal/modules/entities/rollbackprovider`. | implementation, tests, documentation | These are Entity-owned source objects. Revisions should coordinate inverse history and companion ordering, not know their SQL or identity fields. | Internal package migration only. Existing target kinds and merge change sets remain stable. | Entity schema changes can silently break rollback; merge reversal remains coupled to Revisions internals. | Mention resolve/dismiss restoration, companion links, alias/preserved tombstones, merge fan-out, stale errors, affected IDs, and projections remain identical. |
| NI-04 | should fix in this iteration | Extend `internal/modules/links/revisionprovider` to return affected records, exact changed-field effects, target validation, and whole-change-set requirements for Links and Tags. | implementation, tests, documentation | Source operations are already provider-backed, but target semantics still leak through Revisions and `collectionpolicy` imports. Completing the provider removes the partial boundary. | Internal provider interface change. Consolidate separate link/tag provider fields only after tests use the new registry. | New link/tag families continue to require Revisions changes; changed-key or protected-set drift can corrupt collaboration behavior. | Valid link create/delete and tag create/patch/delete rollback pass; owner errors map correctly; changed keys, affected IDs, lock sets, projections, and phase8/phase7 evidence remain stable. |
| NI-05 | should fix in this iteration | Assemble delete/restore, row rollback, non-row rollback, projection, and attribution provider catalogs in `internal/app`; inject them into the Revisions command layer. | architecture, implementation, tests, documentation | Application assembly belongs in `internal/app`. Explicit catalogs fail closed and keep future growth out of Revisions. | Internal constructor and route-option migration. No package-init registration and no default provider substitution. | Mutable globals and hidden defaults make missing/duplicate providers non-deterministic and keep Revisions importing source owners. | Startup/tests reject duplicate and missing required providers; Revisions production files do not import source-owner provider implementations; all registered record/target kinds have contract tests. |
| NI-06 | should fix in this iteration | Introduce a stateless append-only `Appender`, a history reader, and a private command store. Migrate callers to consumer-owned ports and make `CommandService` compose the private stores. Remove the broad exported `Store` surface after the last caller moves. | architecture, implementation, tests, documentation | The current `Store` mixes unrelated lifecycles and permits variadic/nil-database construction solely because append methods use caller transactions. Narrow types make dependencies and testing explicit. | Repository-internal Go API migration across owner modules. Use semantic append methods and migrate all callers in the same iteration; do not leave deprecated aliases. | Future changes keep broad blast radius, tests can construct invalid stores, and provider/command dependencies leak into append-only callers. | No direct `*revisions.Store` field remains outside Revisions; append callers depend on local ports; command construction requires explicit DB/dependencies; phases 3 through 9 and 11 compile and pass. |
| NI-07 | remove/simplify | Remove `DeleteRestoreAdapterTypes`, duplicate Timeline source-mapping guards, `rollbackSourceOwnerProviders`, separate link/tag option fields after registry migration, broad `StoreOptions`, and the default projection adapter fallback. | implementation, tests, documentation | These are test-only exports, transitional globals, duplicated coverage, or hidden compatibility paths. Removing them prevents accidental permanent APIs. | Tests move to package-internal registry assertions or owner provider tests. Application assembly must supply the projection provider. | Temporary surfaces become depended on and make later expansion more brittle. | Repository search finds no callers or obsolete symbols; equivalent owner characterization remains; backend unit and boundary checks pass. |
| NI-08 | should fix in this iteration | Move workbook payload, collection-action, deterministic-local-ID, and text-merge conflict behavior to `internal/modules/workbook/conflicts`. Leave Revisions with a narrow history-window query port only. | architecture, implementation, tests, documentation | Token ownership moved, but most workbook-specific conflict rules still live in Revisions and import workbook/source helpers. Completing the split produces a coherent owner boundary. | Internal type/function import migration; conflict-token v2 and public conflict payloads do not change. | Workbook evolution continues to require Revisions changes and prevents a clean final module boundary. | Phase6 store/integration/browser conflict evidence passes; v2 tokens round-trip; stale tokens and same-field payloads remain unchanged; Revisions no longer imports workbook conflict policy. |
| NI-09 | must fix now | Add a Revisions boundary test that rejects raw source-owner table access, source field mappings, and source-owner provider implementation imports. Update authored boundary inputs only when a repository-wide rule is safe. | tests, harness, documentation | Structural ownership needs an executable regression guard, not only a tracker statement. | May require authored `tools/backend_module_boundaries.json` changes; generated harness outputs must be regenerated through Make if owner inputs change. | Later tactical fixes can silently reintroduce direct SQL and undo the extraction. | `make backend-unit` and `make backend-module-boundary-check` fail on deliberate fixtures and pass on the final tree; only history tables and the shared record envelope remain directly accessed by Revisions. |
| NI-10 | defer with explicit reason | Do not further split authentication, membership, session sliding, pagination, or collaboration publication out of `routes.go` in this iteration. | architecture, documentation | Routes already delegate commands. Further transport decomposition does not unblock source ownership and would expand public-boundary risk. | None. Re-evaluate only with a dedicated route/platform refactor. | Route file remains broad, but current behavior is characterized and no source-owner SQL is introduced. | Phase7 route/auth/pagination/collaboration tests remain green; no new route logic is added during provider work. |
| NI-11 | no action | Keep Revisions ownership of incident bundle history files, attribution boundary, and imported sequence repair. | tests, documentation | Core 01 explicitly assigns these files to the Revisions/history provider. | Preserve exact bundle names; imported deployments may reissue history refs under the owner rule. | Moving them would blur portability coordination and history ownership. | Phase11 bundle tests and imported-attribution boundary test pass; file names remain exact. |
| NI-12 | no action | Keep conflict-token ownership in workbook, destructive locking in records, source-owner delete/restore providers, the projection port, command route delegation, and current route registration. | tests, documentation | These completed structural fixes match current owner specs. | No migration. Later cleanup may change dependency injection only, not behavior or ownership. | Reopening completed work adds churn and compatibility risk. | Their existing unit, phase, and boundary evidence remains part of every final gate. |

### 13.6 Characterization gates and extraction order

Do not move an owner until the named missing characterization is committed and passing. Keep end-to-end route behavior in Revisions tests even when provider unit tests move to owner packages.

| Order | Owner/target | Existing strength | Required characterization before extraction | Primary Make evidence |
| --- | --- | --- | --- | --- |
| R1 | Host in `entities/rollbackprovider` | Strong single-entry, whole-change-set, row-restore, projection, merge, idempotency, and retained-ref coverage. | Add provider tests for nested `source`, view `cells`, direct retained values, required/default fields, update, and touch. | `make backend-unit`; `make backend-store`; phase4 and phase7 slices. |
| R2 | Party in `parties/rollbackprovider` | Direct history-entry rollback covers `display_name`; phase9 owns broader Party behavior. | Cover every mapped nullable/reference field and prove absent rollback values preserve current unrelated fields rather than writing `NULL`. | backend store/integration; phase7 and phase9 slices. |
| R3 | Identity in `entities/rollbackprovider` | Owner create/patch coverage exists; direct rollback coverage is missing. | Add single-entry, whole-row restore, nullable identifier, canonical/merged state, row-version, projection, and changed-key cases. | backend unit/store; phase4 and phase7 slices. |
| R4 | Evidence in `evidence/rollbackprovider` | Direct lifecycle rollback and attached-evidence change-set coverage exist. | Cover blob identity, upload/lifecycle fields, party refs, source values, row restore, and the rule that non-row associations are not implicitly changed. | backend store/integration; phase5 and phase7 slices. |
| R5 | Indicator row in `indicators/rollbackprovider` | Delete/restore source tombstone coverage exists; direct row rollback is thin. | Cover identity/dedupe fields, nullable normalized/hash/STIX values, delete-state clearing, direct row rollback, row restore, projection, and unchanged-field preservation. | backend unit/store/integration; phase4, phase7, and phase9 slices. |
| R6 | Assessment in `assessments/rollbackprovider` | Merge-subject rollback is covered. | Add direct row rollback and row restore for subject, state, score, rationale, assessor, timestamp, delete state, and projection. | backend store/integration; phase4, phase7, and phase9 slices. |
| R7 | Task Request and Decision in `tasksdecisions/rollbackprovider` | Owner lifecycle and Decision supersession behavior are covered outside Revisions. | Add direct rollback/restore for legal lifecycle fields and cross-record refs; add whole-change-set Decision supersession reversal and affected-record checks. | backend unit/store/integration; phase7 and phase9 slices. |
| R8 | Artifact variants in `artifacts/rollbackprovider` | Artifact create/query and variant registry coverage exist; rollback coverage is missing. | Cover `note`, `comm_log`, `handoff`, `status_review`, `lesson`, `finding`, `investigative_query`, and `forensic_keyword`; prove subtype-unrelated columns are not cleared. | backend unit/store/integration; phase7 and phase9 slices. |
| N1 | Links and Tags | Link create/delete, attached evidence, supersede link, and merge tag behavior are covered; valid standalone tag rollback is incomplete. | Add valid tag create/patch/delete, provider error translation, exact affected records, changed keys, lock sets, and projection results. | backend unit/store/integration; phase7 and phase8 slices. |
| N2 | Entity mentions | Resolve, dismiss, restore, companion links, lock sets, and merge repointing are covered. | Add provider-level not-found, stale, malformed, and changed-field results; retain end-to-end companion ordering. | backend unit/store/integration; phase4 and phase7 slices. |
| N3 | Entity alias and preserved identifier | Merge change-set rollback covers successful tombstoning. | Add provider identity parsing, not-found, stale, duplicate, incident binding, and canonical before/after tests. | backend unit/store/integration; phase4 and phase7 slices. |
| N4 | Indicator observation and interval | Owner create/resolve/lifecycle tests exist; no Revisions history or rollback coverage exists. | Add history visibility, target addressing policy, create tombstone, resolution restore, interval tombstone, protected-set, projection, first-class row-version, idempotency, append-only history, and ordinary `record_changed` cases. | migration drift; backend unit/store/integration; phase4, phase7, and phase9 slices. |

Extraction rules:

- Row-backed providers precede non-row providers.
- Within a provider package, move only the listed record type or target family for the current slice; do not combine unrelated owners for convenience.
- Use the Timeline provider as a structural example, not as a universal field model.
- Owners with strong characterization move before owners with polymorphic or broad cross-record effects.
- Artifact variants move last among row-backed owners.
- Do not start the broad `Store` cleanup until all row and non-row provider paths dispatch through explicit contracts.

### 13.7 Phased workstreams

| Workstream | Depends on | Implementation sequence | Principal risks | Exit criteria |
| --- | --- | --- | --- | --- |
| WS-13-00 characterization | none | Add the provider and end-to-end tests from section 13.6; update authored phase-map inputs only for new authoritative evidence. | Tests may accidentally characterize an existing source-clearing bug as required behavior. Use owner specs and full source snapshots to distinguish compatibility from defects. | Every owner in the next extraction slice has passing decode/apply/touch or non-row target tests plus phase7 route evidence. |
| WS-13-01 row provider contract | WS-13-00 for Host | Introduce cycle-free row contract and normalized errors; adapt Timeline without behavior change; add explicit catalog validation. | Contract may absorb projection/history responsibilities and recreate a broad facade. | Contract contains no owner fields/SQL and cannot commit, append history, version records, rebuild, or publish. |
| WS-13-02 row owner extraction | WS-13-01 | Execute R1 through R8 in order. For each: add owner provider, register explicitly, switch one type, delete old Revisions branch, run narrow and phase validation. | Partial dual paths, absent-field clearing, owner default drift, polymorphic Artifact loss. | All supported first-class record types are provider-backed; generic dynamic-table writer and non-Timeline source mappings are gone. |
| WS-13-03 non-row provider contract | WS-13-02, N1 characterization | Add target-kind registry returning validation, atomicity, affected records, inverse values, and changed-field effects. Migrate existing Links/Tags provider adapters first. | Provider could take over history ordering or record versioning; affected-set errors can violate lock precedence. | Revisions owns orchestration only; provider results are deterministic, canonically sorted by Revisions, and acquired before the full protected-set lock. |
| WS-13-04 non-row extraction | WS-13-03 | Execute N1 through N4. Add Indicator tombstone migration before enabling create reversal. | Merge/mention companion ordering, cross-record projections, migration/backfill behavior, and previously unsupported targets. | Every emitted current mutation target is either provider-backed and reversible or explicitly marked unavailable by owner-spec authority; no live emitted kind falls through an accidental default. |
| WS-13-05 facade and composition | WS-13-02, WS-13-04 | Add `Appender`, history reader, and private command store; assemble catalogs and projection dependency in `internal/app`; migrate consumer-local ports; remove broad Store constructors/options and hidden fallbacks. | Large compile-time caller migration and accidental route/conflict behavior change. | Outside Revisions, callers use append/history-specific local ports; routes use `CommandService`; source providers are assembled only in application composition. |
| WS-13-06 workbook conflict ownership | WS-13-05 | Move workbook-specific types and pure conflict behavior to `internal/modules/workbook/conflicts`; retain a narrow Revisions history-window query adapter. | Conflict token or payload drift; dependency cycle between workbook and Revisions. | Phase6 passes, token v2 is unchanged, and Revisions imports no workbook conflict policy or artifact/link conflict helper. |
| WS-13-07 compatibility cleanup | WS-13-05, WS-13-06 | Remove symbols listed in NI-07, move/replace duplicated tests, and update consumer names without aliases. | Removing a test-only export before equivalent internal coverage exists. | Repository search shows zero obsolete callers; no deprecated compatibility layer remains. |
| WS-13-08 boundary closure | WS-13-07 | Add/strengthen Revisions source-table/import guards; run boundary, drift, phase, and broad gates; update this tracker and session log with results. | Over-broad boundary allowlists can hide violations; hand-editing generated harness outputs can create drift. | `rollback_store.go` is orchestration-only, all contracts in section 13.4 pass, boundary checks are executable, and no generated file was hand-edited. |

Per-slice rollback strategy is source-control reversion of that single owner slice. Do not preserve a runtime fallback to the former Revisions implementation as a rollback mechanism.

### 13.8 Make-owned validation matrix

Always rediscover current rows before implementation with `make task-guide ROLE=feature-dev PHASE=<phase>` and `make explain-phase PHASE=<phase>`. The following targets are current at this handoff.

| Change area | Narrow loop | Required phase evidence | Additional gates |
| --- | --- | --- | --- |
| provider contract or pure owner decoding | `make backend-unit` | owner phase plus phase7 when the test is authoritative rollback evidence | `make backend-module-boundary-check` when imports move. |
| row provider source application | `make backend-store`; `make backend-integration` | phase7 plus phase4 for entities/Indicators, phase5 for Evidence, or phase9 for Party/Assessment/Task/Decision/Artifact | `make test-fast` after a group of owner slices. |
| Links and Tags | backend unit/store/integration | phase7 and phase8 | boundary check. |
| Entity non-row targets | backend unit/store/integration | phase4 and phase7 | retain merge protected-set and lock evidence. |
| Indicator child schema/provider | backend unit/store/integration | phase4, phase7, and phase9 | `make migration-drift`; projection and portability checks if schema serialization changes. |
| Store/Appender caller migration | backend unit/store/integration | phases 3, 4, 5, 6, 7, 8, 9, and 11 selected according to moved callers | boundary check, `make test-fast`, then `make check` when final cross-module risk warrants. |
| workbook conflict move | backend unit/store/integration | phase6; phase7 only if history-window code changes | frontend/browser phase6 evidence is required by the phase slice. |
| incident bundle regression | backend integration | phase11 | `make json-shape-check`; exact file-name assertions. |
| final boundary closure | `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check` | phase7 and every owner phase changed since the last green checkpoint | `make agent-finalize` before broad end-of-run verification; `make test-fast`; optional `make check`. |

`make agent-finalize RESULTS_DIR=<root>` may use only a qualifying successful full warm `make check` root. A `test-fast`, phase-slice, service-backed-only, browser-only, or other partial run is not valid retained-run maintenance evidence. When no qualifying full warm root is supplied, run plain `make agent-finalize` and report that retained-run maintenance was skipped because `RESULTS_DIR` was unset.

If authored phase maps or other owner inputs change, regenerate their downstream ledgers, schedules, or topology outputs through the relevant Make generator and run the corresponding drift checks. Never hand-edit generated outputs.

### 13.9 Current planning evidence and tracker validation

Planning baseline gathered before this tracker update:

- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260709T225637Z-p87623`.
- `make json-shape-check` passed at `.cartulary/test-results/20260709T225637Z-p87766`.
- `make backend-module-boundary-check` passed at `.cartulary/test-results/20260709T225638Z-p88083`.
- `make task-guide ROLE=feature-dev PHASE=phase7` and `make explain-phase PHASE=phase7` confirmed phase7 currently selects `backend-store`, `backend-integration`, and `browser-e2e-webserver-backed` evidence.
- Task-guide discovery confirmed owner coverage in phase4 for Entities/Indicators, phase5 for Evidence, phase6 for conflicts, phase8 for Links/Tags, phase9 for coordination/Assessment/Artifact work, and phase11 for portability.

Post-edit tracker validation for this session:

- `make lint-markdown` passed.
- `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260709T230439Z-p91606`.
- `make json-shape-check` passed at `.cartulary/test-results/20260709T230440Z-p91748`.
- `make backend-module-boundary-check` passed at `.cartulary/test-results/20260709T230508Z-p93790`.
- `make agent-finalize` passed at `.cartulary/test-results/20260709T230444Z-p92102`; generated maintenance was unchanged, cached actions were reused, and retained-run maintenance was skipped because `RESULTS_DIR` was unset.

### 13.10 Binary iteration exit criteria

The next implementation iteration is complete only when all of the following are true:

- every current first-class record type uses a source-owner row rollback provider;
- every current emitted non-row mutation target is visible to the correct record history and handled by an owner provider, including Indicator observations and intervals;
- `rollback_store.go` contains no source-owner table SQL, field-key-to-column maps, dynamic owner table names, or owner-specific identity structs;
- Revisions still owns stable history selectors, inverse history accounting, transaction/idempotency behavior, record versioning, and response/event results;
- application assembly provides explicit complete provider catalogs and rejects missing or duplicate registrations;
- no external module stores or constructs the broad `*revisions.Store` facade;
- workbook conflict behavior is workbook-owned while Revisions exposes only the narrow history query needed to evaluate stale windows;
- all temporary wrappers and duplicated Timeline guards listed in NI-07 are removed;
- public contracts in section 13.4 and incident bundle file names are unchanged;
- owner phase slices, phase7, boundary checks, generated-policy/shape checks, finalizer, and broad verification required by the touched surface pass;
- any failure is recorded with its target, run root or summary artifact, relation to the change, and remaining blocker;
- no `BLOCKED: owner contradiction` exists. If one is discovered, stop that slice and record the conflicting owner requirements rather than inventing behavior.
