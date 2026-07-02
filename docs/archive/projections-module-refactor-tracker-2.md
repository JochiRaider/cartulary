# Projections Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Target directory: `internal/modules/projections`.

Output artifact: `docs/handoffs/projections-module-refactor-tracker-2.md`.

Tracker initialized: 2026-07-01T11:14:22-04:00.

Session baseline:

| Item | Value |
| --- | --- |
| Branch | `main` |
| Commit | `0daeae6` |
| Dirty state for target scope before creation | No tracked changes reported for `docs/handoffs/projections-module-refactor-tracker-2.md`, `docs/handoffs/projections-module-refactor-tracker.md`, or `internal/modules/projections` |
| Requested mode | Planning and documentation only |
| Implementation authorization | Not authorized for production refactor, generated-file edit, migration, formatter rewrite, or test rewrite |

Allowed changes for this tracker:

- Create this Markdown handoff tracker.
- Record live repository findings for `internal/modules/projections`.
- Record architectural findings, risks, workflows, proposed behavior-preserving slices, validation commands, and handoff checkpoints.
- Mark unknowns as `TODO:` rather than inventing missing behavior.

Non-goals:

- Do not change production Go, TypeScript, SQL, migrations, generated artifacts, harness manifests, phase maps, tests, route behavior, WebSocket behavior, public envelopes, storage behavior, authorization behavior, or generated contracts.
- Do not treat `internal/modules/projections` as a permanent module boundary merely because the directory exists.
- Do not treat phase maps, phase ledgers, or test rows as runtime architecture.
- Do not promote prior plans or framework files into proof of current repository state.

Source hierarchy used by this tracker:

| Order | Source family | Use in this tracker |
| --- | --- | --- |
| 1 | Adopted subsystem NLSpecs for their named subsystem only | `docs/graph_projection_nlspec.md` is now adopted for graph-projection behavior only. It does not govern the workbook projection facade tracked here. |
| 2 | Core 00 through Core 04 | Runtime behavior, module boundaries, public contracts, view schemas, projections, restore, authorization, row wire, and collaboration semantics. |
| 3 | Core 05 | Only for claim-bearing timed, benchmark, fixture-sensitive, or publication evidence. Not used as Base Profile runtime authority here. |
| 4 | Domain vocabulary and implementation-support guides | Vocabulary, boundary interpretation, generated-file policy, harness mechanics, Make invocation, and evidence accounting. |
| 5 | Current repository code and tests | Current implementation state, package surfaces, caller graph, tests, contracts touched, and characterization posture. |
| 6 | Prior plans and framework files | Planning template and historical evidence only. They are not proof of current repository state. |

Architectural finding: `internal/modules/projections` is currently a hybrid workbook projection facade, not proof of a permanent module boundary. It legitimately owns projection-facing facade behavior: `Store`, provider registry, generic projection query and row serialization, restore rebuild orchestration, and projection telemetry. Source-domain projection SQL has already moved behind source-owner `projectionprovider` packages for timeline, entities, indicators, assessments, artifacts, evidence, parties, and tasks/decisions. The main remaining boundary risks are `query.go` row/query mapping, provider rebuild ordering, restore readiness, and contract drift.

No owner-document contradiction was found during this planning pass. Any later contradiction between owner documents must be marked `BLOCKED: owner contradiction` and not resolved inside this tracker.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/projections/.gitkeep` | Empty directory sentinel. | None. | Git-tracked path only. | None. | None. | None. | Out of behavioral scope. | low | Inventory only; no behavior. |
| `internal/modules/projections/assessments.go` | Compatibility facade for assessment projection row refresh and incident rebuild. Delegates SQL to `internal/modules/assessments/projectionprovider`. | `RefreshAssessmentTx`, `RebuildIncidentAssessmentsTx`. Private core methods bridge registry dispatch to provider calls. | `internal/modules/assessments/store.go`, `internal/modules/assessments/import_create.go`, `internal/modules/entities/merge_store.go`, `internal/modules/incidentbundles/source.go`, `internal/modules/revisions/delete_restore_store.go`, `internal/modules/revisions/rollback_store.go`, `rebuild.go` through registry dispatch, tests. | `internal/modules/assessments/projectionprovider`, `context`, `uuid`, `pgx`. | Assessment Phase 9 contract tests, workbook Phase 9 tests, revisions delete/restore and rollback tests, provider registry tests indirectly. | `contracts/view-schemas/cartulary.view.assessments.v1.json`; generated contract consumers by alignment only. | Keep facade in `projections`; source SQL belongs to `assessments/projectionprovider`. | medium | Behavior-preserving facade; do not move public method names without caller migration. |
| `internal/modules/projections/entity_grids.go` | Compatibility facade for host, identity, and indicator projection rebuilds. Delegates source-domain SQL to `entities/projectionprovider` and `indicators/projectionprovider`. | `RebuildIncidentHosts`, `RebuildIncidentHostsTx`, `RebuildIncidentIdentities`, `RebuildIncidentIdentitiesTx`, `RebuildIncidentIndicators`, `RebuildIncidentIndicatorsTx`. | `internal/modules/timeline/ports.go`, `internal/modules/entities/mention_lifecycle.go`, `internal/modules/evidence/store.go`, `internal/modules/incidentbundles/source.go`, revisions delete/restore and rollback stores, recovery restore via registry, Phase 4/5/7/9 tests. | `internal/modules/entities/projectionprovider`, `internal/modules/indicators/projectionprovider`, `pgx`, `uuid`. | Entity Phase 4 tests, timeline/entity support tests, evidence Phase 5 tests, revisions Phase 7 tests, recovery Phase 10 tests. | Host, identity, and indicator view-schema contracts; generated consumers by alignment only. | Keep facade in `projections`; source SQL belongs to entities and indicators providers. | high | Rebuild ordering matters because indicator/entity summary fields can depend on timeline/entity state. |
| `internal/modules/projections/provider_registry.go` | Provider descriptor model, built-in provider registration, required surface coverage, duplicate owner checks, projection table-family owner checks, rebuild graph validation, topological rebuild order, and refresh/rebuild dispatch. | `ProviderDescriptor` plus package-private registry/provider types and `Store` dispatch helpers. | `Store.NewStore`, `Store.providerRegistry`, refresh/rebuild methods in this package, `RebuildRestoreProjections`, provider registry tests. | `context`, `fmt`, `sort`, `uuid`, `pgx`; package-level view schema constants and facade methods. | `provider_registry_test.go`; recovery restore tests indirectly through rebuild order; projection package tests. | `tools/schema_object_ownership_manifest.json` by owner-label agreement; view-schema contracts by provider coverage. | `projections` facade. | critical | High-risk internal contract. Required surfaces and restore rebuild order must remain deterministic. |
| `internal/modules/projections/provider_registry_test.go` | Package-local characterization of built-in provider order, required surface coverage, optional artifact surface mapping, duplicate detection, unknown table-family rejection, owner mismatch rejection, missing dependency rejection, and cycle rejection. | Test-only. | Go test runner. | `reflect`, `strings`, `testing`; package-private registry helpers. | Direct tests in this file. | None directly. | Projection test evidence. | low | Test evidence only; not runtime architecture. |
| `internal/modules/projections/query.go` | Generic projection-backed workbook query engine: surface map, SQL builder, sort/filter/group handling, artifact surface contract filters, collection cell shaping, `view_row_v1` row serialization, and transaction row loading. | `SupportsQuerySurface`, `(*Store).QueryRows`, `(*Store).LoadRowTx`. Package-private field/surface/query helpers. | `internal/modules/workbook/store.go`, `internal/modules/workbook/mutation_store.go`, `internal/modules/artifacts/import_projection.go`, `internal/modules/artifacts/linkednotes/facade.go`, `internal/modules/assessments/import_create.go`, `internal/modules/evidence/import_projection.go`, `internal/modules/parties/import_projection.go`, `internal/modules/tasksdecisions/import_projection.go`, `internal/modules/tasksdecisions/supersede_facade.go`, workbook/query tests. | `internal/platform/viewschema`, `pgx`, `uuid`, projection tables, `records`, `record_links`, `record_tags`, `parties`, `handoff_risk_refs`, artifact/evidence/assessment/task/decision projection tables. | `query_test.go`, workbook Phase 8/9 query and grouping tests, assessment/evidence/task/decision/coordination tests, saved-view query tests indirectly. | `contracts/view-schemas/*`, `packages/view-contracts/src/index.ts`, `internal/gen/contracts/**`, `packages/protocol-ts/src/generated/**` by contract alignment only. | Keep in `projections` until stronger characterization proves a smaller query-provider split is safe. | critical | Highest row-wire risk. Field maps, null handling, collection shapes, sort/filter/group behavior, and `LoadRowTx` snapshots must not drift. |
| `internal/modules/projections/query_test.go` | Package-local characterization for generic projection surface matrix, contract field coverage, default sort/sort/filter/group mapping, row shape, null shape, collection shape, artifact contract filters, and grouped-row full-row serialization. | Test-only. | Go test runner. | `encoding/json`, `reflect`, `strings`, `testing`, `time`, `uuid`, `viewschema`; package-private query helpers. | Direct tests in this file. | View-schema contract alignment by test evidence only. | Projection test evidence. | low | Test evidence only. Extend before moving query behavior. |
| `internal/modules/projections/rebuild.go` | Restore-time projection rebuild orchestration: lists incidents and rebuilds every registered provider in registry topological order inside one transaction. | `RebuildRestoreProjections`; package-private `listRestoreProjectionIncidentIDs`. | `internal/modules/recovery/restore.go`, `internal/app/operator.go`, `tools/phase10browserrestore/main.go`, `cmd/server/main_phase10_recovery_sentinel_test.go`, recovery tests. | `pgx`, `pgtype`, `uuid`, provider registry dispatch, `incidents` table. | Recovery Phase 10 backup/restore tests, command sentinel tests, provider registry tests indirectly. | Recovery/operator artifacts and harness accounting by behavior only. | `projections` facade or a future recovery-owned adapter over the facade. | critical | Recovery readiness behavior. Preserve restore order, transaction behavior, and failure wrapping. |
| `internal/modules/projections/store.go` | Defines projection `Store`, Postgres handle, provider registry wiring, timeline projection input, timeline projection upsert into `timeline_grid_projection`, timeline rebuild facade, and UUID helper. | `Store`, `TimelineProjectionInput`, `NewStore`, `UpsertTimelineRowTx`, `RebuildIncidentTimeline`, `RebuildIncidentTimelineTx`. | Timeline store/lifecycle/clipboard ports, incident bundles, revisions, evidence support refresh, recovery restore, app/operator wiring, tests. | `internal/modules/timeline/projectionprovider`, `internal/platform/postgres`, `pgx`, `pgtype`, `uuid`. | Timeline Phase 3/4 tests, evidence Phase 5 tests, revisions Phase 7 tests, recovery Phase 10 tests, projections registry tests indirectly. | `timeline_grid_projection` schema and timeline view-schema contract by behavior only. | Keep facade/upsert in `projections`; timeline source derivation belongs to `timeline/projectionprovider`. | high | Compatibility facade. Direct table upsert remains in facade because timeline provider supplies derived input via callback. |
| `internal/modules/projections/tasks_decisions.go` | Compatibility facade for task request and decision projection row refresh and incident rebuild. Delegates SQL to `tasksdecisions/projectionprovider`. | `RefreshTaskRequestTx`, `RefreshDecisionTx`, `RebuildIncidentTaskRequestsTx`, `RebuildIncidentDecisionsTx`. | Workbook mutation store, tasksdecisions import refresh, decision supersede facade, incident bundles, revisions, recovery restore via registry, Phase 9 tests. | `internal/modules/tasksdecisions/projectionprovider`, `pgx`, `uuid`. | Task/decision Phase 9 store tests, coordination tests, revisions tests, supersede tests. | Task request and decision view-schema contracts; generated protocol contracts by alignment only. | Keep facade in `projections`; source SQL belongs to tasksdecisions provider. | high | Decision supersession uses projection refresh and `LoadRowTx` before/after mutation for revision payloads. |
| `internal/modules/projections/telemetry.go` | Facade-level projection rebuild telemetry span creation and safe `view_schema_id` vocabulary. | Package-private `startProjectionSpan`, `telemetryServiceVersion`, `safeProjectionViewSchemaID`; view schema constants. | Rebuild methods in this package. | `internal/platform/telemetry`, OpenTelemetry attributes/codes/trace, `strings`. | `telemetry_test.go`; OTel conformance script references. | OTel golden/conformance references by behavior only. | `projections` facade with platform telemetry primitive. | medium | Keep telemetry centralized; providers should return errors/descriptors rather than scatter telemetry logic. |
| `internal/modules/projections/telemetry_test.go` | Unit characterization for safe telemetry vocabulary and no-SDK behavior. | Test-only. | Go test runner; OTel conformance references. | `testing`; package-private telemetry helpers. | Direct tests in this file. | `scripts/check-otel-conformance.mjs` references this test. | Projection telemetry evidence. | low | Test evidence only. |
| `internal/modules/projections/workbook_hot.go` | Compatibility facade for artifact, evidence, and party projection refresh/rebuild plus shared hot-projection transaction wrapper. Delegates source SQL to artifact, evidence, and party projection providers. | `RefreshArtifactTx`, `RebuildIncidentArtifacts`, `RebuildIncidentArtifactsTx`, `RefreshEvidenceTx`, `RebuildIncidentEvidence`, `RebuildIncidentEvidenceTx`, `RefreshPartyTx`, `RebuildIncidentParties`, `RebuildIncidentPartiesTx`. | Workbook mutation store, artifact import refresh, linked notes facade, evidence store/import refresh, party store/import refresh, incident bundles, recovery restore via registry, Phase 9 tests. | `internal/modules/artifacts/projectionprovider`, `internal/modules/evidence/projectionprovider`, `internal/modules/parties/projectionprovider`, `pgx`, `uuid`. | Workbook Phase 9 notes/coordination/party tests, evidence tests, artifact optional surface tests, recovery tests. | Notes, evidence, parties, coordination, and optional artifact-backed view-schema contracts by behavior only. | Keep facade in `projections`; source SQL belongs to artifact/evidence/party providers. | high | Artifact provider backs notes, comm log, handoff, status review, lesson, findings, investigative queries, and forensic keywords through `artifact_grid_projection`. |

Generated roots must not be hand-edited: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`, and generated task-surface outputs. `packages/view-contracts/src/index.ts` is authored but contract-derived; update only through a separate contract-aware task if needed.

## 3. Workbook Boundary Diagnosis

`internal/modules/workbook` is a mixed transport-adjacent application facade, query dispatcher, and mutation coordinator. It owns route-facing workbook composition, query dispatch, row create/patch orchestration, authorization interaction with routes, mutation result shaping, conflict/revision calls, collaboration publication paths, and owner-facade calls. It is not proof that workbook presentation or transport should own source-record semantics, projection materialization, saved-view lifecycle, revision conflict semantics, or collaboration publication.

`internal/modules/projections` is not the workbook package. It is currently a projection/query facade and view-row materialization layer. Its durable posture should be: keep the facade, avoid catch-all growth, and defer any further split until characterization tests protect behavior.

Current diagnosis: mixture of legitimate thin projection facade, view/projection orchestration layer, restore rebuild coordinator, and row/query materialization layer. It is not currently an accidental general workbook catch-all, because source-domain projection SQL has already moved to owner `projectionprovider` packages. The remaining risk is that `query.go` still centralizes generic row/query SQL maps for many workbook surfaces.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Workbook query route authorization and pagination | `internal/modules/workbook/routes.go` and `store.go` | Workbook routes/application facade | keep | Query route authenticates, requires incident membership, decodes query, then calls `Store.QueryRows`. | `projections` should not gain public route/auth responsibility. |
| Projection-backed row query materialization | `internal/modules/projections/query.go` | Projections facade; possible future per-domain query providers | defer | `QueryRows` and `LoadRowTx` shape `view_row_v1` and row refresh snapshots. | Split only after characterization covers every surface. |
| Timeline query path | `internal/modules/workbook/store.go` dispatches to timeline facade | Timeline | keep | Workbook dispatches `cartulary.view.timeline.v2` to `timelineStore.QueryTimelineRows`. | `projections` rebuilds/upserts timeline projection but does not own public timeline query path. |
| Hosts/identities/indicators query path | `internal/modules/workbook/store.go` dispatches to entities store | Entities/indicators | keep | Workbook dispatches host/identity/indicator schemas to entity store methods. | Rebuild facade remains in projections for compatibility. |
| Source-domain projection SQL | `internal/modules/*/projectionprovider/provider.go` | Source owner providers | keep | Provider packages exist for artifacts, assessments, entities, evidence, indicators, parties, tasksdecisions, timeline. | Current state already implements hybrid boundary. |
| Projection table schema ownership | `tools/schema_object_ownership_manifest.json` row `workbook-projections` | Projections implementation-support owner | keep | Manifest maps `*_grid_projection` families to `projections`. | Manifest is implementation-support evidence, not product authority. |
| Restore projection rebuild | `internal/modules/projections/rebuild.go`; called by `internal/modules/recovery/restore.go` | Projections facade or future recovery adapter over facade | defer | Core 01 restore order requires projection rebuild after Postgres/object restore. | Preserve `RebuildRestoreProjections` until an adapter has equivalent characterization. |
| Revision/change-set row snapshots | `LoadRowTx` callers in workbook mutation, tasks decision supersede, owner import refresh | Revisions plus source owner facades using projections row refresh | keep | `LoadRowTx` provides current full row snapshots after projection refresh. | Row shape is public-contract sensitive. |
| Import apply row refresh | Owner modules call projections behind import facades | Source owner facades | keep | `internal/modules/imports/boundary_test.go` forbids direct imports of `workbook` and `projections`. | Import dispatcher must not call projections directly. |
| Collaboration row changes | Workbook/revisions/collaboration modules | Collaboration and route owners | keep | Core 01 owns `record_changed`; projections only contributes row shape/refresh inputs. | Do not move WS publication into projections. |
| Telemetry for projection rebuild | `telemetry.go` | Projections facade with platform telemetry primitive | keep | Tests bound safe view schema vocabulary. | Avoid scattering telemetry into provider packages. |

## 4. Public Contract and Behavior Freeze Map

| Contract surface | Observable behavior to freeze | Existing tests or evidence | Required characterization before refactor |
| --- | --- | --- | --- |
| HTTP view query route | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query`; common envelope, `rows[]`, `meta.query`, body-only pagination, stable error behavior. | Workbook integration tests, Phase 8 query tests, Phase 9 surface tests, platform viewquery tests. | Add/keep per-surface projection query characterization for any `query.go` move. |
| HTTP row create | `POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows` returns `data.view_schema_id`, `data.change_set_id`, and full `data.row`. | Workbook Phase 9 tests, assessment/evidence/party/task/decision tests. | Preserve full `view_row_v1` after refresh. |
| HTTP record patch | `PATCH /api/v1/records/{record_id}` uses `view_schema_id`, optimistic row versions, field-key changes, conflict behavior, refreshed full row. | Workbook mutation/integration tests, Phase 6 conflict tests, Phase 9 tests. | Characterize `LoadRowTx` output before changing row builders. |
| HTTP supersede | `POST /api/v1/records/{record_id}/supersede` for Timeline and Decision targets; decision response includes `cartulary.view.decisions.v1`; changed rows refresh projections. | Timeline tests, tasksdecisions supersede tests, revisions tests. | Ensure decision projection refresh and `LoadRowTx` snapshots are unchanged. |
| HTTP delete/restore/rollback | Record-scoped routes preserve row versions, change sets, remove/invalidate behavior, and revision history. | Revisions Phase 7 tests, recovery and workbook tests. | Characterize affected projection rebuild/refresh behavior if facade changes. |
| HTTP evidence attach/access | Evidence attach returns evidence `view_row_v1`; preview/download handles recheck authorization. | Evidence Phase 5 tests, workbook evidence tests. | Preserve evidence projection refresh and row shape. |
| View-schema discovery | `GET /api/v1/view-schemas` and singleton discovery expose exact standardized schemas and field registries. | Viewschema registry tests, generated contract checks. | Run `make json-shape-check` if contract inputs change. |
| Saved-view/startup routes | Saved views remain additive over immutable `view_schema_id`; startup returns selected sheet ref and base view schema. | Savedviews Phase 8 tests, incidents startup tests. | Query semantic changes must preserve saved-view query JSON behavior. |
| WebSocket connection | `/ws/v1/incidents/{incident_id}` remains collaboration stream root. | Collaboration Phase 6 tests, platform WS tests. | Projection refactors must not add WS route behavior. |
| WebSocket `record_changed` | Preserve sorted `changed_field_keys[]`, sorted `affected_views[]`, `patch_cells`, `invalidate`, and `remove`; affected views keyed by base `view_schema_id`. | Phase 6/7/9 tests, Core 01/03 contract coverage. | Add event-level characterization if changed-field extraction or row patch generation changes. |
| Workbook row wire | Full `view_row_v1` has top-level `record_id`, `row_version`, `cells`, `group_values`; cell objects are `{ "value": ... }`. | `query_test.go`, workbook tests, generated protocol contract. | Preserve null cells and technical-field placement. |
| Workbook query semantics | Sort/filter/group behavior uses stable field keys, declared schema capability, token/prefix/range/equality behavior, default sort tail, and grouping values. | `query_test.go`, viewquery tests, Phase 8/9 tests. | Per-surface matrix before any query split. |
| Collection cell shape | `collection_value_v1` shapes for tags, record refs, party refs, support refs, risk refs; ordering remains deterministic. | `query_test.go`, coordination collection tests. | Canonical JSON comparison before moving collection expression helpers. |
| Projection refresh | Source writes refresh affected projection rows transactionally through facade/provider calls. | Owner store tests, Phase 9 hot projection tests. | Characterize per affected surface before method signature changes. |
| Projection rebuild | Incident rebuild and restore rebuild regenerate disposable projection tables from authoritative source state. | Provider registry tests, Phase 10 recovery tests. | Preserve provider topological order and failure behavior. |
| Authorization checks | Query and mutation routes rederive auth; projection internals do not authorize by themselves. | Route tests and Core 04 acceptance coverage. | Do not introduce direct projection routes or auth bypasses. |
| Revision/change-set snapshots | Before/after rows and row versions for revisions/conflicts/supersede remain stable. | Revisions tests, tasksdecisions supersede tests. | `LoadRowTx` row shape characterization. |
| Generated protocol/view contracts | Generated roots remain downstream; no hand-editing. | `generated-artifact-policy-check`, `json-shape-check`, generated contract tests. | Run drift checks if owner inputs change. |
| Harness/test accounting | Phase maps and ledgers remain evidence accounting only. | `make task-guide`, scheduler manifests. | Update owner inputs first if rows change; do not infer runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `projections` facade is legitimate but not a permanent proof of all current responsibilities. | `Store`, provider registry, query engine, restore rebuild, and telemetry exist under `internal/modules/projections`; source SQL providers live under source owners. | Medium | intentional/no_action | Projections facade | Keep facade while preventing catch-all growth. |
| `query.go` centralizes row/query SQL maps for many domain surfaces. | Surface map covers assessments, evidence, artifact-backed surfaces, parties, task requests, decisions, and optional artifact surfaces. | Critical | must_fix | Projections query facade, possible later query providers | Add/keep characterization before any movement; defer split until tests protect every surface. |
| Provider registry is an internal cross-domain rebuild coordinator. | `provider_registry.go` validates required surfaces and topological rebuild order. | Critical | should_fix | Projections facade | Harden registry tests before changing provider descriptors or rebuild order. |
| Restore rebuild lives in projections and is called by recovery. | `rebuild.go` implements `RebuildRestoreProjections`; `recovery/restore.go` calls it after object-store restore. | Critical | intentional/no_action | Recovery over projections facade | Preserve public internal method or add adapter only after restore characterization. |
| Source-domain SQL already moved behind providers. | Provider packages exist under artifacts, assessments, entities, evidence, indicators, parties, tasksdecisions, timeline. | Medium | intentional/no_action | Source owner providers | Maintain caller isolation; do not move provider SQL back into facade. |
| Projection package imports platform telemetry. | `telemetry.go` imports `internal/platform/telemetry`. | Medium | intentional/no_action | Projections facade with platform telemetry primitive | Keep telemetry centralized and safe-vocabulary tested. |
| Projection facade exports test-sensitive API used by owner facades. | `LoadRowTx` is used by imports, linked notes, and supersede logic for full row snapshots. | High | should_fix | Projections facade plus owner facades | Treat `LoadRowTx` as compatibility surface; characterize before signature changes. |
| Authorization remains outside projection internals. | Workbook routes authenticate and authorize before calling store/facade methods. | High | intentional/no_action | Route owners/Core 04 | Do not add projection routes or route-level auth into projection package. |
| Import dispatcher must not call projections directly. | `internal/modules/imports/boundary_test.go` forbids imports of `internal/modules/projections`; owner import refresh files call projections behind owner packages. | Medium | intentional/no_action | Imports dispatcher and source owner facades | Keep boundary test; add rows only through owner facades. |
| Generated contracts and projection query maps can drift. | `query.go` field keys must match `contracts/view-schemas/*` and generated consumers. | Critical | must_fix | Core 01 contracts plus projections query facade | Run `json-shape-check` when contracts/manifests change; add field-map tests before refactor. |
| Phase maps can be mistaken for architecture. | Phase 8/9/10 rows cover projection behavior, but harness docs say phase maps are evidence accounting. | Medium | intentional/no_action | Harness/evidence accounting | Keep phase labels out of runtime structure. |
| Graph projection NLSpec is adopted for a separate subsystem. | `docs/graph_projection_nlspec.md` is adopted for graph-oriented projection behavior only. | Medium | intentional/no_action | Graph projection activation tracker | Do not apply graph-projection lifecycle or query rules to workbook projection read models. |
| Schema object ownership manifest maps projection tables to projections. | `tools/schema_object_ownership_manifest.json` `workbook-projections` owner row covers `*_grid_projection`. | Medium | intentional/no_action | Projections implementation-support owner | Do not alter schema ownership without separate owner decision. |
| Test-only assumptions should not leak into production API. | Current query characterization is package-local; previous exported test helpers are absent. | Low | intentional/no_action | Package-local tests | Keep test helpers unexported unless runtime API needs them. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Capture authority order, branch/commit, dirty tree, target directory, and planning-only scope. | This tracker, framework file, `docs/domain.md` | `git status --short`, `make help` | Tracker records source posture and no production refactor authorization. |
| WF-01 | Projections and workbook-adjacent inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every file in `internal/modules/projections`, exported surface, callers, dependencies, tests, generated risks, and owner candidates. | `internal/modules/projections/**`, caller modules | `rg`, `find`, `git ls-files` | Inventory table has no generic placeholders. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map Core 00-04, view schemas, routes, WebSocket events, generated contracts, restore/import owners. | Core specs, workbook routes/store, collaboration, imports, recovery, contracts | `make json-shape-check` when changed | Freeze map identifies existing tests or characterization gaps. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-06 | Identify missing tests before future movement of query, provider, restore, or telemetry logic. | Projection tests, workbook tests, imports/recovery tests | `make backend-unit`, `make backend-store` | Gap table names tests to preserve or add before movement. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05 | Classify direct SQL coupling, platform imports, source-owner leakage, auth placement, and test-only production API. | Projection package, source-owner modules, schema ownership manifest | `make lint-go`, `make go-gosec-targeted` | Findings table has classification and required action. |
| WF-05 | Facade/ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Preserve `projections` facade and do not decide permanent module boundary. | Projections facade, owner provider packages, workbook/import/recovery callers | Characterization suite | Decision checkpoint records facade posture and public contract freeze. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Define smallest safe behavior-preserving slices for any future code movement. | Selected future files only | `make phase-slice PHASE=phase9` for workbook projection changes | Each slice has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Plan test map, generated ledger, or manifest updates only if owner inputs change; keep phase maps evidence-only. | `tools/phase*_test_map.json`, generated ledgers through generators only | drift checks if harness inputs change | Handoff says whether harness accounting changed or was skipped. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow validation, broaden by risk, update tracker and handoff log. | Tracker plus touched implementation files in later task | `make agent-finalize`; `make test-fast`; `make check` as risk requires | Next agent can resume without rediscovery. |

## 7. Proposed Refactor Slice Plan

| Slice | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create tracker only. | `docs/handoffs/projections-module-refactor-tracker-2.md` | Docs stale risk only. | No code tests required; preserve command/source log. | `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-markdown` if docs scope includes handoffs. | Delete only this tracker file if rejected. | New tracker exists with all required sections. |
| S-01 | S-00 | Add or confirm characterization for `QueryRows`, `LoadRowTx`, row shape, nulls, collections, sort/filter/group. | `internal/modules/projections/query.go`, `query_test.go`, workbook query tests. | Critical row-wire drift. | Preserve `query_test.go`; add store-backed rows for representative projection-backed surfaces if gaps remain. | `make backend-unit`; `make backend-store`; `make phase-slice PHASE=phase9` | Revert tests only; no production behavior change. | Tests fail on row/query contract drift. |
| S-02 | S-01 | Harden provider registry and rebuild-order tests before descriptor changes. | `provider_registry.go`, `provider_registry_test.go`, provider facade files. | Restore/provider ordering drift. | Preserve required surface coverage, optional artifact provider mapping, cycle and missing dependency tests. | `make backend-unit`; `make backend-store` | Revert registry-only changes if order is not ready. | Required surfaces and rebuild order are covered. |
| S-03 | S-01 | Characterize restore rebuild compatibility before touching `rebuild.go`. | `rebuild.go`, recovery tests, app/operator wiring. | Recovery readiness. | Preserve Phase 10 restore and restore-verification checks. | `make backend-store`; `make backend-integration`; `make service-backed-slice PHASE=phase10` | Revert restore-adapter changes as one slice. | Restore rebuild still opens/query-probes incidents. |
| S-04 | S-01 | Add package-level projection import guard after owner-approved facade allowlist. | `internal/modules/projections/boundary_guard_test.go`; Appendix I; provider descriptor manifest. | Cross-module leaks or freezing incidental file structure. | Preserve existing import guards; enforce Appendix I allowlist; ignore `_test.go`; distinguish provider internals from approved facades. | `make backend-unit`; `make lint-go` | Remove or adjust the guard only with owner-approved facade update. | AC-473 guard passes and no unapproved production imports exist. |
| S-05 | S-02, S-03 | Defer any further provider/query split until S-01 through S-04 pass. | Future selected provider or query package only. | Public behavior drift. | Owner-specific characterization plus projection package tests. | Same as touched owner slice, plus Phase 9/10 as applicable. | Roll back one provider/query split at a time. | No behavior change without characterization. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Go unit and static package tests | yes for future code changes | Covers projections registry/query tests. |
| integration | `make backend-store` | Service-backed store tests | yes for query/refresh/rebuild work | Requires Postgres/object store. |
| integration | `make backend-integration` | Broader route/workflow integration | yes for route, revision, restore, or cross-module impact | Broaden by risk. |
| e2e/browser | `make browser-e2e-webserver-backed` | Cross-stack workbook browser behavior | no for tracker creation; yes if frontend/query UX behavior changes | Use only for UI-visible projection/query changes. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check` | Generated policy, JSON shape, contract/manifests | yes when docs/contracts/manifests touched | Do not hand-edit generated roots. |
| import-boundary/static | `make frontend-import-boundary-check`; `make lint-go`; `make go-gosec-targeted` | Static boundaries and security | no for tracker creation; yes if imports/boundaries/security-sensitive code changes | Use Make wrappers. |
| phase slice | `make phase-slice PHASE=phase9` | Workbook projections and surfaces | yes for workbook projection work | Use `make task-guide ROLE=feature-dev PHASE=phase9`. |
| restore slice | `make service-backed-slice PHASE=phase10` | Recovery and restore | yes for `rebuild.go` or restore rebuild changes | Restore rebuild contract. |
| full check | `make agent-finalize`; `make test-fast`; `make check` | End-of-run or broad handoff verification | no for tracker creation; yes before broad implementation handoff | Use retained `RESULTS_DIR` only if supplied. |

If commands are missing or unclear for a future narrow slice, write `TODO: command discovery required` in that slice before implementation.

## 8A. Remediation Execution Controller

This tracker is the controlling artifact for the projection gap remediation effort started on 2026-07-01T13:20:14-04:00. Each workstream below must update this tracker after completion and before the next workstream starts.

Historical authority gate result for this completed workbook-projection remediation: no adopted projections-specific subsystem NLSpec was found at run start. That result has been superseded for graph projection by `docs/handoffs/graph-projection-activation-tracker.md`; `docs/graph_projection_nlspec.md` is now adopted only for the separate graph-projection subsystem. Core 00 through Core 04 remain the controlling product authority for workbook projection behavior covered by this tracker.

| Workstream | Status | Depends on | Required update before next workstream | Exit criteria |
| --- | --- | --- | --- | --- |
| WS-00 Tracker and authority gate | DONE | none | Add this execution controller and a handoff log entry. | Tracker records workstreams, authority status, and handoff protocol. |
| WS-01 Specification cleanup | DONE | WS-00 | Record spec files touched, validation commands, and any owner contradiction. | Core 01 and Appendix I name the restore adapter request/result shape and descriptor vocabulary. |
| WS-02 Query and row-wire characterization | DONE | WS-01 | Record test files touched, failing drift fixed or deferred, and validation run roots. | Query/row-wire tests fail on schema-cell, null, collection, grouping, and `LoadRowTx` parity drift. |
| WS-03 Provider descriptor and boundary policy | DONE | WS-02 | Record descriptor, manifest, guard, and validation changes. | Code-backed descriptors own manifest and boundary policy metadata. |
| WS-04 Recovery-owned restore adapter | DONE | WS-03 | Record adapter contract, caller migrations, restore tests, and validation results. | Recovery uses a structured restore projection rebuild adapter and readiness fails closed on rebuild failure. |
| WS-05 Query engine structural cleanup | DONE | WS-04 | Record query split files, registry wiring, and parity validation. | `QueryRows` and `LoadRowTx` remain stable while query internals are cohesive and registry-backed. |
| WS-06 Validation and handoff completion | DONE | WS-05 | Record final run roots, skipped checks with reason, and remaining risks. | `make agent-finalize`, `make test-fast`, and broad/risk-appropriate verification are reported. |

Handoff protocol for this remediation:

- Update this section's workstream status before starting the next workstream.
- Add a row to the session handoff log after each completed workstream.
- If a contradiction between owner documents is discovered, mark the current workstream `BLOCKED: owner contradiction` and do not resolve it locally.
- Do not hand-edit generated roots. Update owner inputs and run Make-owned drift targets when generated artifacts are affected.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| T-001 | Define target directory and documentation-only scope. | WF-00 | DONE | none | This tracker section 1 | Target and non-goals are explicit. |
| T-002 | Inspect `internal/modules/projections` directly. | WF-01 | DONE | T-001 | File list in section 2 | Every tracked file is inventoried or marked out of behavioral scope. |
| T-003 | Map current facade/provider architecture. | WF-01 | DONE | T-002 | Sections 2, 3, 5 | Facade, provider packages, and ownership candidates are recorded. |
| T-004 | Map public contracts affected by projections. | WF-02 | DONE | T-002 | Section 4 | Contract freeze map names tests or characterization posture. |
| T-005 | Record workbook boundary diagnosis. | WF-02 | DONE | T-003 | Section 3 | Workbook/projections distinction is explicit. |
| T-006 | Classify coupling and boundary findings. | WF-04 | DONE | T-003 | Section 5 | Each finding has risk and classification. |
| T-007 | Define workflow dependency map. | WF-06 | DONE | T-004, T-006 | Section 6 | Workflows have dependency and checkpoint columns. |
| T-008 | Define behavior-preserving slice plan. | WF-06 | DONE | T-007 | Section 7 | Each slice has dependency, validation, rollback, and completion criterion. |
| T-009 | Discover validation commands. | WF-08 | DONE | T-007 | Section 8 | Make-owned commands are listed or scoped. |
| T-010 | Seed handoff log. | WF-08 | DONE | T-008 | Section 10 | Handoff rows identify files inspected/touched, commands, blockers, and next action. |
| T-011 | Track future query characterization expansion. | WF-03 | TODO | T-004 | S-01 | Query/row-wire gaps are closed before query movement. |
| T-012 | Track future restore rebuild characterization. | WF-03 | TODO | T-004 | S-03 | Restore rebuild changes have Phase 10 evidence. |
| T-013 | Track future provider registry hardening. | WF-03 | TODO | T-006 | S-02 | Required surface/rebuild-order tests protect descriptor changes. |
| T-014 | Add S-04 projection import guard. | WF-04 | DONE | T-006 | `internal/modules/projections/boundary_guard_test.go`; Appendix I; AC-473 | Guard reflects owner-approved production facades and distinguishes test imports. |
| T-015 | Run verification for tracker and closure artifacts. | WF-08 | DONE | T-010 | `make generated-artifact-policy-check`, `make json-shape-check`, `make lint-markdown`, `make backend-unit`, `make lint-go` | Docs, manifest, projection guard tests, and Go lint passed. |
| T-016 | Close Q-001 through Q-005 as authority, ownership, and boundary decisions. | WF-05 | DONE | T-014 | Core 00, Core 01, Core 04, Appendix I, section 11 | All five questions are closed without requiring projection behavior migration. |
| T-017 | Add validation-only provider descriptor manifest. | WF-02 | DONE | T-016 | `contracts/projection-providers/index.json`; `tools/schemas/cartulary.projection_provider_manifest.v1.schema.json`; `provider_manifest_test.go` | Manifest is shape-checked and mirrors the code-backed registry. |
| T-018 | Update traceability and profile DoD navigation. | WF-08 | DONE | T-016 | Appendix F; Core 04 base claim manifest | REQ-00-062..REQ-00-063, REQ-01-621..REQ-01-626, and AC-469..AC-473 are navigable. |
| T-019 | Promote tracker into remediation execution controller. | WS-00 | DONE | T-018 | Section 8A | Workstreams and handoff protocol are recorded before implementation. |
| T-020 | Tighten restore adapter and descriptor vocabulary in owner docs. | WS-01 | DONE | T-019 | Core 01 and Appendix I | Spec cleanup validates through docs and JSON-shape targets. |
| T-021 | Expand query and row-wire characterization. | WS-02 | DONE | T-020 | `internal/modules/projections/query_test.go`, `query_store_test.go`, workbook route test | Tests fail on schema-cell or `LoadRowTx` parity drift. |
| T-022 | Move provider descriptor and boundary metadata into code-backed descriptors. | WS-03 | DONE | T-021 | `provider_registry.go`, manifest, guard tests | Manifest and guard derive from one code-backed policy. |
| T-023 | Introduce recovery-owned restore projection adapter contract. | WS-04 | DONE | T-022 | `internal/modules/recovery/restorecontract`, recovery/projections callers | Restore readiness uses structured rebuild results. |
| T-024 | Split query internals while preserving facade behavior. | WS-05 | DONE | T-023 | Projection query implementation files | Characterization suite passes before and after split. |
| T-025 | Complete validation and final handoff. | WS-06 | DONE | T-024 | Final run roots and handoff log | Final tracker status reports completed work, skipped checks, and residual risks. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | Tracker creation authorized as documentation-only and validation completed. | Inspected framework, `docs/domain.md`, Core 00-04 excerpts, harness target guidance, projections files, provider packages, workbook/recovery/import/revision callers. Touched this tracker only. | `sed`, `find`, `rg`, `wc`, `git status`, `git rev-parse`, `make help`, `make task-guide ROLE=feature-dev PHASE=phase9`, `make explain-target`, `make -qp` target scan, `date`, `make generated-artifact-policy-check`, `make json-shape-check`, `make lint-markdown`. | Planning evidence gathered; no production refactor; docs/harness checks passed. | None for tracker creation. | Future implementation starts with characterization slices, not production movement. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | `projections` is a hybrid facade. Source SQL providers already exist in source-owner packages. | `internal/modules/projections/**`, `internal/modules/*/projectionprovider/provider.go`, workbook/recovery/import/revision callers. | `find`, `rg`, `sed`, `git ls-files`. | Current state recorded in sections 2, 3, and 5. | Future split requires characterization. | Start S-01 before moving query behavior. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | Contract risks are view-query, row wire, WebSocket row patches, view-schema registry, generated protocol/view consumers, and saved-view query behavior. | Core specs, `tools/generated_artifact_policy.json`, contract/generated roots. Touched this tracker only. | `rg`, `sed`, `find`, `make explain-target TARGET=json-shape-check DETAIL=summary`, `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary`, `make generated-artifact-policy-check`, `make json-shape-check`. | Generated roots identified as read-only; generated artifact policy and JSON shape checks passed. | None for tracker creation. | Re-run drift checks if a later task changes contracts, schemas, manifests, or generated policy inputs. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | Existing evidence spans projections unit tests, workbook Phase 8/9 tests, recovery Phase 10 tests, revisions tests, and provider owner tests. | Test files searched; touched tracker only. | `rg`, `make task-guide ROLE=feature-dev PHASE=phase9`, `make explain-target` for validation targets, `make -qp` target scan, `make lint-markdown`. | Make-owned validation plan recorded; named future targets exist; markdown lint passed. | Broad test execution not required for docs-only creation. | Future code movement begins with S-01/S-02/S-03 characterization. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | Authorization remains route-owned; projections methods accept IDs/transactions and must not become public routes. | Workbook routes, Core 04 excerpts, collaboration contracts. Touched tracker only. | `sed`, `rg`. | Security boundary recorded as intentional/no_action. | Future direct projection route would require owner spec. | Keep auth checks outside projections internals. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-01T11:14:22-04:00 | Codex tracker creation session | Remaining risks: query row-wire drift, provider rebuild order, restore readiness, contract drift, and generated-surface mismatch. | This tracker. | This session's inspection commands. | Risks seeded in sections 5, 7, 9, and 11. | No implementation blockers for tracker creation. | Future implementer starts with S-01 characterization. |
| 2026-07-01T12:59:04-04:00 | Codex closure session | Core/Appendix/manifest/guard closure implemented without public query, restore, rebuild, or row-wire behavior changes. | Core 00, Core 01, Core 04, Appendix F, Appendix I, README, `docs/domain.md`, tracker, provider manifest/schema/checker, projection manifest and boundary tests. | `rg`, `sed`, `date`, `mkdir -p`, `gofmt`, `make generated-artifact-policy-check`, `make json-shape-check`, `make lint-markdown`, `make backend-unit`, `make lint-go`, `make agent-finalize`. | Q-001 through Q-005 closed as authority, ownership, and boundary decisions; docs/harness checks passed; backend unit passed 90 tests; Go lint passed; agent finalize passed. | None for tracker closure; future implementation movement still requires S-01 through S-03 characterization. | Future behavior movement starts with S-01/S-02/S-03 characterization, not a broad projection rewrite. |
| 2026-07-01T13:20:14-04:00 | Codex remediation implementation session | WS-00 started. Worktree was clean and no adopted projections-specific subsystem NLSpec was found. | `docs/handoffs/projections-module-refactor-tracker-2.md`, `docs/graph_projection_nlspec.md`, Core projection authority references. | `git status --short`, `rg`, `sed`, `date`. | Tracker promoted to remediation execution controller. | None. | Start WS-01 specification cleanup. |
| 2026-07-01T13:22:20-04:00 | Codex remediation implementation session | WS-01 completed and WS-02 started. | `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/I_projection_authority_boundary_and_characterization.md`, this tracker. | `make lint-markdown`, `make json-shape-check`. | Restore adapter request/result vocabulary and provider descriptor vocabulary tightened; markdown lint passed; JSON shape passed with run root `.cartulary/test-results/20260701T172214Z-p1805470`. | None. | Add query and row-wire characterization tests. |
| 2026-07-01T13:28:40-04:00 | Codex remediation implementation session | WS-02 completed and WS-03 started. | `internal/modules/projections/query_test.go`, `internal/modules/projections/query_store_test.go`, `internal/modules/workbook/workbook_integration_test.go`, this tracker. | `gofmt`, `make backend-unit`, `make backend-store`, `make phase-slice PHASE=phase9`. | Schema-exact generic projection field coverage, store-backed `QueryRows`/`LoadRowTx` parity, and a route-boundary projection query test were added. Backend unit passed with run root `.cartulary/test-results/20260701T172700Z-p1807870`; backend store passed with run root `.cartulary/test-results/20260701T172718Z-p1810059`; Phase 9 slice passed with run root `.cartulary/test-results/20260701T172758Z-p1819365`. | None. | Move provider descriptor and boundary metadata into code-backed descriptors. |
| 2026-07-01T13:36:17-04:00 | Codex remediation implementation session | WS-03 completed and WS-04 started. | `internal/modules/projections/provider_registry.go`, `internal/modules/projections/provider_boundary.go`, `internal/modules/projections/provider_manifest_test.go`, `internal/modules/projections/boundary_guard_test.go`, `internal/modules/projections/provider_registry_test.go`, `internal/modules/projections/rebuild.go`, this tracker. | `gofmt`, `make backend-unit`, `make lint-go`, `make json-shape-check`, `make generated-artifact-policy-check`. | Production provider descriptors now own schema version, status, capabilities, restore participation, and facade packages; registry validation rejects unsupported schema/status, active unsupported restore rebuild, invalid facade policies, capability/implementation mismatches, and query capability drift. Manifest and import guard now derive approved data from code-backed sources. Backend unit passed with run root `.cartulary/test-results/20260701T173520Z-p1835220`; Go lint passed; JSON shape check passed with run root `.cartulary/test-results/20260701T173609Z-p1846786`; generated artifact policy check passed with run root `.cartulary/test-results/20260701T173609Z-p1846811`. | None. | Introduce recovery-owned restore projection adapter contract. |
| 2026-07-01T13:51:47-04:00 | Codex remediation implementation session | WS-04 completed and WS-05 started. | `internal/modules/recovery/restorecontract/projections.go`, `internal/modules/recovery/restore.go`, `internal/modules/recovery/verification.go`, `internal/modules/recovery/phase10_restore_projection_contract_test.go`, `internal/modules/projections/rebuild.go`, `internal/modules/projections/rebuild_test.go`, `cmd/server/main_phase10_recovery_sentinel_test.go`, this tracker. | `gofmt`, `make backend-unit`, `make backend-store`, `make backend-integration`, `make service-backed-slice PHASE=phase10`, `make backend-process`, retry `make service-backed-slice PHASE=phase10`. | Recovery now owns the structured projection rebuild request/result contract and fails closed before readiness unless projection rebuild reports ready or not-applicable. Projections validate restore requests before transaction state, rebuild providers in registry order, report provider-level status and row counts, and preserve stale-row replacement. Backend unit passed with run root `.cartulary/test-results/20260701T174348Z-p1850690`; backend store passed with run root `.cartulary/test-results/20260701T174427Z-p1857200`; backend integration passed with run root `.cartulary/test-results/20260701T174508Z-p1867373`. Initial Phase 10 service-backed slice failed at `.cartulary/test-results/20260701T174603Z-p1879736` because `cmd/server` Phase 10 sentinel test fakes still used the obsolete adapter; after migrating those fakes, backend-process passed with run root `.cartulary/test-results/20260701T174930Z-p1894875` and Phase 10 service-backed slice passed with run root `.cartulary/test-results/20260701T175033Z-p1906033`. | None. | Split query internals while preserving facade behavior. |
| 2026-07-01T14:02:05-04:00 | Codex remediation implementation session | WS-05 completed and WS-06 started. | `internal/modules/projections/query.go`, `query_types.go`, `query_sql.go`, `query_row.go`, `query_surface_expr.go`, `query_surface_artifacts.go`, `query_surface_assessments.go`, `query_surface_evidence.go`, `query_surface_parties.go`, `query_surface_tasks_decisions.go`, `provider_registry.go`, `provider_registry_test.go`, `query_test.go`, `telemetry_test.go`, `rebuild_test.go`, this tracker. | `gofmt`, non-canonical local probe `go test ./internal/modules/projections`, `make backend-unit`, `make backend-store`, `make phase-slice PHASE=phase9`. | `query.go` now contains only the public query/load facade; SQL/filter building, row serialization, shared query types, expression helpers, and provider-owned surface descriptors are split into cohesive files. Query-capable provider descriptors register `QuerySurfaces`, and the provider registry owns the query surface map used by `SupportsQuerySurface`, `QueryRows`, and `LoadRowTx`. A local package probe first failed on a stale test fixture column and was fixed before Make validation. Backend unit passed with run root `.cartulary/test-results/20260701T180001Z-p1923671`; backend store passed with run root `.cartulary/test-results/20260701T180040Z-p1930101`; Phase 9 slice passed with run root `.cartulary/test-results/20260701T180122Z-p1940282`. | None. | Run final validation and handoff completion. |
| 2026-07-01T14:07:50-04:00 | Codex remediation implementation session | WS-06 completed. Projection remediation is complete end-to-end. | This tracker plus all files listed in WS-00 through WS-05 handoff rows. | `make agent-finalize`, `make test-fast`, `make check`. | Final validation passed. Agent finalize passed with run root `.cartulary/test-results/20260701T180246Z-p1953956`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. `make test-fast` passed 970 tests with run root `.cartulary/test-results/20260701T180258Z-p1955347`. `make check` passed 276/276 work units and 973 tests with run root `.cartulary/test-results/20260701T180506Z-p2003368`. Compatibility note: no intended public HTTP, WebSocket, generated protocol, or DB schema change; internal restore adapter and provider/query internals changed as planned. | None. | Ready for owner review or commit. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| Q-001 | Is there an adopted projections-specific subsystem NLSpec? | Adopted NLSpecs outrank Core docs for their named subsystem. | Core 00 REQ-00-062; Appendix I status record; `docs/graph_projection_nlspec.md` is adopted for graph projection only. | Closed / resolved / non-blocking for workbook projections. Graph projection is tracked separately and must not be applied to workbook projection read models. |
| Q-002 | Should `query.go` remain centralized permanently or split into query providers? | It controls public row/query behavior for many surfaces. | Core 01 REQ-01-623; AC-471; Appendix I query characterization matrix. | Closed / implementation deferred. `query.go` remains current facade but is not a permanent normative requirement; future split requires parity characterization. |
| Q-003 | Should restore rebuild orchestration move behind a recovery-owned adapter? | Recovery readiness is critical and currently calls `projections.RebuildRestoreProjections`. | Core 01 REQ-01-624 and REQ-01-625; AC-472; Appendix I restore characterization. | Closed / adapter contract required / behavior preserved. Initial adapter may delegate to `projections.RebuildRestoreProjections`. |
| Q-004 | Should provider descriptors become a generated or manifest-backed contract? | Could reduce drift between view schemas, table owners, and providers. | Core 01 REQ-01-622; AC-470; `contracts/projection-providers/index.json`; provider manifest test. | Closed / validation manifest only. Code-backed registry remains authoritative; runtime manifest authority deferred. |
| Q-005 | Are additional boundary guard tests needed for projections imports? | Guard tests can prevent future cross-module leakage but can also freeze intentional facades too early. | Core 01 REQ-01-626; AC-473; Appendix I import graph; `boundary_guard_test.go`. | Closed / S-04 action complete. Guard enforces package-level production boundaries with a separate test-only posture. |

No current `BLOCKED: owner contradiction` item was found.

## 12. Binary Completion Criteria

This tracker is complete when all of the following are true:

- Every file in `internal/modules/projections` is inventoried or explicitly out of scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every implementation slice is behavior-preserving unless explicitly marked as requiring later authorization.
- Validation commands are discovered or marked as `TODO:` with reason.
- Handoff sections are current enough for another agent to continue without rediscovery.
- No production refactor was run as part of tracker creation.
- Files inspected and commands run are summarized.
- The tracker creation or update status is stated.
- The exact output path is stated.
- Remaining blockers are stated.

Tracker creation status: created as a documentation-only planning artifact and updated for projection authority/boundary closure.

Exact output path: `/home/jochi/code/cartulary/docs/handoffs/projections-module-refactor-tracker-2.md`.

Files inspected during planning and creation:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/domain.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`
- `docs/handoffs/projections-module-refactor-tracker.md`
- `tools/generated_artifact_policy.json`
- `tools/schema_object_ownership_manifest.json`
- `internal/modules/projections/**`
- `internal/modules/*/projectionprovider/provider.go`
- Projection-adjacent workbook, recovery, imports, revisions, timeline, entities, evidence, assessments, parties, tasksdecisions, and artifacts callers found by `rg`.

Commands run during planning and creation:

- `sed` to read framework, domain, spec, harness, projections, provider, and generated-policy files.
- `find` to list `internal/modules/projections`, provider packages, contract roots, and generated roots.
- `rg` to find callers, imports, tests, contracts, routes, WebSocket contract references, and validation target references.
- `wc -l internal/modules/projections/*.go`.
- `git status --short`.
- `git rev-parse --abbrev-ref HEAD` and `git rev-parse --short HEAD`.
- `git ls-files` for the target directory.
- `make help`.
- `make task-guide ROLE=feature-dev PHASE=phase9`.
- `make explain-target TARGET=<target> DETAIL=summary` for `backend-unit`, `backend-store`, `backend-integration`, `frontend-unit`, `generated-artifact-policy-check`, `json-shape-check`, `frontend-import-boundary-check`, `phase-slice`, and `agent-finalize`.
- `make -qp | awk ... | rg ...` to confirm named future validation targets exist in the Make database.
- `date --iso-8601=seconds`.
- `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260701T151853Z-p1766460`.
- `make json-shape-check` passed with run root `.cartulary/test-results/20260701T151853Z-p1766476`.
- `make lint-markdown` passed.
- Closure update: `mkdir -p contracts/projection-providers`.
- Closure update: `gofmt -w internal/modules/projections/boundary_guard_test.go internal/modules/projections/provider_manifest_test.go`.
- Closure update: `date --iso-8601=seconds`.
- Closure validation: `make generated-artifact-policy-check` passed with run root `.cartulary/test-results/20260701T165820Z-p1791880`.
- Closure validation: `make json-shape-check` passed with run root `.cartulary/test-results/20260701T165824Z-p1792065`.
- Closure validation: `make lint-markdown` passed.
- Closure validation: `make backend-unit` passed with run root `.cartulary/test-results/20260701T165840Z-p1792821`.
- Closure validation: `make lint-go` passed.
- Closure validation: `make agent-finalize` passed with run root `.cartulary/test-results/20260701T165936Z-p1797319`.

Remaining blockers:

- No blocker for this documentation artifact.
- Future implementation remains blocked on characterization for any `query.go`, `provider_registry.go`, or `rebuild.go` behavior movement.
- Future owner-document contradictions, if discovered, must be marked `BLOCKED: owner contradiction` instead of resolved here.
