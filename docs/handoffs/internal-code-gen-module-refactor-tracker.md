# internal-code-gen Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- Target path: `internal/gen`.
- Target label: `internal-code-gen`.
- Normalized label posture: lowercase kebab case; no spaces, path separators, shell metacharacters, or unsafe filename characters.
- Output path: `docs/handoffs/internal-code-gen-module-refactor-tracker.md`.
- Planning-only status: this tracker is documentation and handoff planning only.
- Allowed changes for this session: create this tracker file only.
- Non-goals: no production refactor, no source movement, no behavior change, no generated-artifact hand edit, no contract edit, no SQL/migration edit, no harness edit, no package configuration edit, and no test edit.
- Implementation posture: any implementation, code movement, authored SQL change, contract change, generated refresh, or deletion requires a later authorized task.
- Source hierarchy: adopted subsystem NLSpecs for their named subsystem only; Core 00 through Core 04 for implementation-conformance behavior; Core 05 only for claim-bearing timed or fixture-sensitive publication; `docs/domain.md` and implementation-support guides for terminology, boundaries, harness mechanics, and execution support; current repository code and tests for current implementation state; prior handoffs and framework files as evidence only.
- Architectural finding: `internal/gen` is a generated artifact root and is not proof that `internal-code-gen` is a valid permanent module boundary. Current evidence shows generated adapter surfaces for contract embedding and SQLC outputs. Behavioral ownership belongs to owner specs, authored contracts, authored SQL, and surrounding modules.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/testing-harness-nlspec.md`
- `docs/graph_projection_nlspec.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/report-composition-nlspec.md`
- `docs/opentelemetry-instrumentation-nlspec.md`
- `docs/network-flow-activity-nlspec.md` was inspected only to confirm its draft status and non-authority for this target.

Repository files inspected:

- `internal/gen/contracts/.gitkeep`
- `internal/gen/contracts/contracts_gen.go`
- `internal/gen/sql/.gitkeep`
- `internal/gen/sql/auth_phase1.sql.go`
- `internal/gen/sql/db.go`
- `internal/gen/sql/incidents_phase2.sql.go`
- `internal/gen/sql/models.go`
- `internal/gen/sql/recovery_phase10.sql.go`
- `internal/gen/sql/reporting_phase11.sql.go`
- `internal/gen/sql/savedviews_phase8.sql.go`
- `internal/gen/sql/timeline_phase3.sql.go`
- `internal/gen/sql/workbook_startup_phase8.sql.go`
- `db/queries/auth_phase1.sql`
- `db/queries/incidents_phase2.sql`
- `db/queries/recovery_phase10.sql`
- `db/queries/reporting_phase11.sql`
- `db/queries/savedviews_phase8.sql`
- `db/queries/timeline_phase3.sql`
- `db/queries/workbook_startup_phase8.sql`
- `sqlc.yaml`
- `tools/generated_artifact_policy.json`
- `tools/contractgen/main.go`
- `tools/harness/generated-artifacts/generate-artifacts.sh`
- `tools/harness/generated-artifacts/check-generate-drift.sh`
- Representative callers and tests in `internal/modules/incidents`, `internal/modules/timeline`, `internal/modules/savedviews`, `internal/modules/workbook/startup`, `internal/modules/recovery`, `internal/modules/reporting`, `internal/platform/viewschema`, `internal/platform/ws`, `internal/platform/authn`, and generated-contract tests.

Repository/framework mismatches and planning findings:

- The framework offers module names as a planning catalog. Live repository evidence shows `internal/gen` is generated, mixed by generator family, and not a domain module.
- The generated SQL package includes auth/session/bootstrap queries whose equivalent behavior is currently implemented with raw SQL in `internal/platform/authn/store.go`; no direct production caller of those generated auth query functions was found.
- The generated SQL package includes some apparently uncalled generated queries, including `GetSessionMemberships` and `GetLatestSuccessfulRetainedBackupSet`; removal or consolidation is deferred pending owner review of authored SQL inputs and generation behavior.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/gen/contracts/.gitkeep` | Keeps generated contract directory present. | None. | Repository layout and generated-artifact policy only. | None. | Generated artifact policy tests reference sentinel behavior. | Generated sentinel under `internal/gen`. | Generated artifact policy. | Low | Out of behavior scope; do not hand-edit except repository layout maintenance. |
| `internal/gen/contracts/contracts_gen.go` | Generated Go embedding of contract artifacts and SHA-256 values. | Package `contracts`; `Artifact`; `OpenAPIArtifacts`, `WSArtifacts`, `ViewSchemaArtifacts`, `ErrorArtifacts`, `ExtensionArtifacts`, per-family indexes, and `ContractArtifactIndex`. | Contract tests in modules, `internal/platform/viewschema/registry.go`, `internal/platform/ws/ws_test.go`, `internal/testutil/phase2test/contracts.go`. | Standard library only; generated from `contracts/*` by `tools/contractgen`. | OpenAPI, WS, view-schema, error, extension, evidence, incident, revision, workbook, reporting, reference-data, and viewschema tests inspect embedded JSON. | OpenAPI, WS, view-schema, error, and extension contract registries. | Contracts and view-contract registry, not `internal-code-gen`. | High | Generated contract surface can affect many public routes and view-schema behaviors; owner inputs are `contracts/*` and specs. |
| `internal/gen/sql/.gitkeep` | Keeps generated SQL directory present. | None. | Repository layout and generated-artifact policy only. | None. | Generated artifact policy tests reference sentinel behavior. | Generated sentinel under `internal/gen`. | Generated artifact policy. | Low | Out of behavior scope; do not hand-edit except repository layout maintenance. |
| `internal/gen/sql/db.go` | SQLC database adapter shell. | Package `sqlc`; `DBTX`, `Queries`, `New`, `WithTx`. | All SQLC-consuming module stores call `sqlc.New`. | `context`, `pgx/v5`, `pgconn`. | Indirectly through all stores using generated SQLC. | SQLC generated code from `sqlc.yaml`. | Persistence adapter shared by module stores. | Medium | Thin adapter; not domain behavior. |
| `internal/gen/sql/auth_phase1.sql.go` | Generated auth, session, and bootstrap token SQL queries. | `GetLocalUserByEmail`, `GetSessionByFingerprint`, `ListActiveSessionsForUser`, `GetBootstrapTokenByFingerprint` and row/param types. | No direct production import use found; `internal/platform/authn/store.go` implements similar queries with raw SQL. | `context`, `pgtype`, `DBTX`. | Auth handlers use platform auth store, not this generated file directly. | SQLC output from `db/queries/auth_phase1.sql`. | `auth` or platform auth persistence, pending owner decision. | Medium | Planning finding: generated queries appear unused by production; do not remove without authored SQL and generator review. |
| `internal/gen/sql/incidents_phase2.sql.go` | Generated incident, membership, visibility, lifecycle, and audit queries. | Incident query methods and row/param types including visibility, membership, lifecycle, audit, admin count. | `internal/modules/incidents/store.go`, `open_guard.go`, `import_finalization.go`. | `context`, `pgtype`, `DBTX`. | Phase 2 incident request tests; downstream route behavior tests. | SQLC output from `db/queries/incidents_phase2.sql`. | `incidents`; session membership query is TODO owner review. | High | Owns no behavior itself; surrounding incidents module maps SQL rows to domain records and API results. |
| `internal/gen/sql/models.go` | Generated table model structs for migrated schema. | Many exported structs including records, projections, users, sessions, incidents, saved views, recovery, reporting, reference data, import, evidence, and workbook preferences. | SQLC query return types and module conversion helpers. | `pgtype`. | Indirectly through all SQLC-backed stores. | SQLC output from migrations and `sqlc.yaml`. | Persistence model adapter, split by consuming modules. | High | Mixed schema model file spans many owners; not a domain boundary. |
| `internal/gen/sql/recovery_phase10.sql.go` | Generated backup-set and restore-verification persistence queries. | Backup-set and restore-verification query methods and row/param types. | `internal/modules/recovery/store.go`; `GetLatestSuccessfulRetainedBackupSet` appears uncalled. | `context`, `pgtype`, `DBTX`. | Recovery store/operator tests indirectly. | SQLC output from `db/queries/recovery_phase10.sql`. | `recovery`. | Medium | Uncalled generated query needs owner review before any authored SQL cleanup. |
| `internal/gen/sql/reporting_phase11.sql.go` | Generated reporting snapshot, release, job payload, release approval, invalidation, and state queries. | Reporting query methods and row/param types. | `internal/modules/reporting/store.go`. | `context`, `pgtype`, `DBTX`. | Reporting OpenAPI and lifecycle tests indirectly; store tests through module. | SQLC output from `db/queries/reporting_phase11.sql`. | `reporting`; report composition and graph projection are adjacent owners. | High | Snapshot/release public behavior is owned by Core and reporting NLSpecs, not generated SQL. |
| `internal/gen/sql/savedviews_phase8.sql.go` | Generated saved-view create/list/visibility/update/delete queries. | Saved-view query methods and row/param types. | `internal/modules/savedviews/store.go`; workbook startup resolves saved-view visibility through savedviews module. | `context`, `pgtype`, `DBTX`. | Phase 8 saved-view tests and workbook startup tests indirectly. | SQLC output from `db/queries/savedviews_phase8.sql`. | `savedviews`. | High | Saved-view visibility and mutation semantics are behaviorally owned by savedviews and Core 03. |
| `internal/gen/sql/timeline_phase3.sql.go` | Generated timeline projection row and projection-source queries. | `GetTimelineProjectionRow`, `ListTimelineProjectionRows`, `ListTimelineProjectionSourceRows` and row types. | `internal/modules/timeline/store.go`, `query_projection_store.go`, `projectionprovider/provider.go`. | `context`, `pgtype`, `DBTX`. | Timeline, revisions, evidence, workbook query, and projection-related tests indirectly. | SQLC output from `db/queries/timeline_phase3.sql`. | `timeline` and workbook projections; graph projection NLSpec does not own workbook projections. | High | Projection/source behavior must not be inferred from generated filename or SQL names. |
| `internal/gen/sql/workbook_startup_phase8.sql.go` | Generated workbook startup preference and pointer cleanup queries. | Preference query methods and row/param types. | `internal/modules/workbook/startup/store.go`, `bootstrap/bootstrap.go`. | `context`, `pgtype`, `DBTX`. | Phase 2 workbook-preference OpenAPI test and phase 8 startup/saved-view behavior tests indirectly. | SQLC output from `db/queries/workbook_startup_phase8.sql`. | Workbook startup under `workbook`, with savedviews as adjacent owner. | High | Startup surface selection is Core-owned behavior; SQLC is only persistence realization. |

Every file under `internal/gen` has been inventoried above. No file is marked missing.

## 3. Module Boundary Diagnosis

Diagnosis summary: `internal/gen` is a mixed generated root containing downstream adapter code. It is a generated contract registry surface and SQL persistence adapter surface. It is not a legitimate application/service facade, not a workbook orchestration owner, not a frontend shell/controller surface, not a grid-vendor integration layer, and not a domain module.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Generated artifact root policy | `internal/gen/**` | Harness generated-artifact policy and `docs/testing-harness-nlspec.md` | Keep | `tools/generated_artifact_policy.json`; harness NLSpec says generated files must not be hand-edited. | Root is an output location, not behavior owner. |
| Go contract embedding | `internal/gen/contracts/contracts_gen.go` | `contracts/*`, `tools/contractgen`, Core 00-04, adopted NLSpecs for their owned scope | Keep | `tools/contractgen/main.go` writes this file from `contracts/*`. | Generated registry supports tests and runtime view-schema loading. |
| SQLC database adapter shell | `internal/gen/sql/db.go` | Persistence adapter shared by consuming module stores | Keep | `sqlc.yaml` outputs to `internal/gen/sql`; callers use `sqlc.New`. | No module behavior should live here. |
| Auth/session generated SQL | `internal/gen/sql/auth_phase1.sql.go` | `auth` and platform auth persistence | Defer | Search found no direct production caller; `internal/platform/authn/store.go` uses raw SQL. | Requires owner decision before query cleanup. |
| Incident and membership SQL | `internal/gen/sql/incidents_phase2.sql.go` | `incidents`; possibly platform auth for session-membership summary if revived | Keep / split later | `internal/modules/incidents` imports generated SQL; `GetSessionMemberships` appears uncalled. | Keep generated file as output; inspect authored SQL if splitting. |
| Timeline projection SQL | `internal/gen/sql/timeline_phase3.sql.go` | `timeline` plus workbook projections | Keep / defer | Timeline store and projection provider consume generated rows. | Graph Projection NLSpec explicitly does not own workbook-grid projections. |
| Saved-view SQL | `internal/gen/sql/savedviews_phase8.sql.go` | `savedviews` | Keep | Savedviews store consumes generated queries. | Core 03 owns saved-view behavior. |
| Workbook startup SQL | `internal/gen/sql/workbook_startup_phase8.sql.go` | `workbook` startup, with savedviews as adjacent owner | Keep | Workbook startup store and bootstrap port consume generated queries. | Startup behavior must preserve sheet-ref fallback and pointer cleanup semantics. |
| Recovery SQL | `internal/gen/sql/recovery_phase10.sql.go` | `recovery` | Keep / defer | Recovery store consumes most generated queries; one latest backup query appears uncalled. | No change without recovery owner review. |
| Reporting SQL | `internal/gen/sql/reporting_phase11.sql.go` | `reporting` and adopted reporting/report-composition NLSpecs for their scopes | Keep | Reporting store consumes generated queries. | Generated query shape can affect snapshot/release persistence. |
| Generated table models | `internal/gen/sql/models.go` | Split by consuming module stores and authored schema owners | Keep | SQLC generated models span many tables and modules. | Mixed schema output is normal generated state, not module ownership. |
| Workbook view-schema registry loading | `internal/platform/viewschema/registry.go` consuming `gencontracts.ViewSchemaArtifacts` | View contracts and platform view-schema registry | Defer | Registry loads embedded view-schema JSON from generated contracts. | Correct owner is not `internal-code-gen`; future work may reduce direct generated dependency through facade if useful. |
| Unused generated SQL queries | Several generated SQL files | Owning module plus authored SQL owner | Defer | Search found generated functions with no production caller. | Planning-only finding; no deletion in this task. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP route shapes and envelopes | Core 01/Core 03/Core 04 plus adopted subsystem NLSpecs where named | `contracts/openapi/cartulary.openapi.yaml` embedded by `contracts_gen.go`; tests inspect generated OpenAPI. | Incidents, entities, workbook, savedviews, reporting, reference-data, revisions tests inspect OpenAPI artifact. | Preserve current OpenAPI artifact or regenerate only from owner inputs; add route-specific tests before moving owner logic. | High | Generated contract is not authority but is an observable surface. |
| WebSocket route and events | Core 01/Core 03/Core 04; platform WebSocket implementation | `contracts/ws/index.schema.json` embedded by `contracts_gen.go`; route is `GET /ws/v1/incidents/{incident_id}`. | `internal/platform/ws/ws_test.go` checks `job_progress`; phase 6 WS tests check replay filtering. | Characterize message shape and replayability before any contract generation change. | High | Messages include `presence_snapshot`, `record_changed`, `job_progress`, `session_revoked`. |
| Workbook row/query/mutation behavior | Core 01/Core 02/Core 03; source modules such as timeline/entities/evidence/savedviews | View-schema contracts, workbook OpenAPI tests, timeline projection SQL and store code. | Workbook OpenAPI tests; phase tests across timeline, revisions, evidence, savedviews. | Add owner-specific characterization if SQLC query ownership changes. | High | Do not infer behavior from generated SQL filenames. |
| Saved-view behavior | Core 03; `internal/modules/savedviews`; workbook startup for sheet-ref selection | `savedviews_phase8.sql.go`; savedviews store; workbook startup store. | Phase 8 saved-view tests and OpenAPI contract tests. | Preserve visibility, scope, version conflict, startup resolution, and layout/query normalization. | High | Saved-view scope must not become authorization for underlying rows. |
| View-schema behavior | Core 01/Core 03; `contracts/view-schemas/*`; `internal/platform/viewschema` | `ViewSchemaArtifacts` embedded in `contracts_gen.go`; registry loads 17 view-schema artifacts. | `internal/platform/viewschema/registry_test.go`; evidence/workbook tests inspect field keys. | Add registry load and field-order characterization before changing contract generation. | High | View schemas are generated contract surfaces, not generated-code-owned behavior. |
| Projection refresh behavior | Core 01 for workbook projections; Graph Projection NLSpec only for graph projections | `timeline_phase3.sql.go`, timeline projection provider, graph-projection NLSpec non-scope text. | Timeline and projection-adjacent tests indirectly. | Characterize rebuild source row ordering and projection row fields before moving SQL. | High | `internal/gen` is not projection owner. |
| Authorization checks | Core 04; modules and platform auth layers | Saved-view visibility SQL joins memberships; incident visibility SQL; auth store raw SQL. | Auth, incidents, savedviews, WS, and route tests. | Required before moving visibility SQL behind new facades. | High | Generated SQL may express access checks but does not own them. |
| Revision/change-set behavior | Core 02/Core 03; `revisions` and source modules | Revisions tests inspect OpenAPI generated contracts; timeline store uses row version and conflict token logic. | Phase 7 history and rollback tests. | Characterize row-version and conflict behavior before changing generated query consumers. | High | Generated SQL model row versions must preserve public envelope behavior. |
| Generated protocol/view contracts | Owner specs, `contracts/*`, `tools/contractgen`, generated-artifact harness | `contracts_gen.go` embeds OpenAPI, WS, view schemas, errors, extensions. | Contract tests across modules. | `make generate-drift` and targeted contract tests after owner input changes. | High | Hand edits forbidden. |
| Grid-adapter or UI-selector contracts | Frontend packages, not current target | No direct grid-adapter import from `internal/gen`. | Frontend checks not inspected beyond command discovery. | Only needed if generated contracts consumed by frontend change. | Medium | Use `make frontend-import-boundary-check` for frontend contract boundary changes. |
| Harness/test accounting | `docs/testing-harness-nlspec.md`; Make task surface | `make help`, `make explain-target`, generated-artifact policy. | Generated-artifact policy tests. | Generated drift and policy checks before any generated-output refresh. | Medium | Phase maps are evidence accounting only, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `internal/gen` is a generated mixed root, not a module boundary. | Generated artifact policy lists `internal/gen`; files have generated markers. | Mistaken module ownership could centralize unrelated behavior. | must_fix | Generated artifact policy and owner modules. | Treat `internal-code-gen` as tracker label only; do not create a domain module. |
| Generated SQLC package spans multiple bounded responsibilities. | `models.go` includes auth, incidents, records, evidence, imports, reporting, recovery, reference data, workbook preferences. | Cross-module coupling through shared generated package can hide owner decisions. | should_fix | Consuming module stores. | Plan facades or module-local conversion boundaries where SQLC types leak beyond persistence code. |
| Auth generated queries appear unused while platform auth store uses raw SQL. | Search found generated functions and raw SQL in `internal/platform/authn/store.go`. | Drift between authored SQLC query inputs and live auth persistence behavior. | defer | `auth` or platform auth persistence. | Owner decision required before deleting, adopting, or ignoring generated auth queries. |
| Some generated queries appear uncalled. | `GetSessionMemberships` and `GetLatestSuccessfulRetainedBackupSet` found only in generated files and authored SQL. | Dead generated query inputs can confuse ownership and tests. | defer | Owning modules for authored SQL. | Characterize whether queries are retained intentionally before cleanup. |
| Contract generated Go is consumed by runtime view-schema registry. | `internal/platform/viewschema/registry.go` imports `internal/gen/contracts`. | Contract generation drift can break runtime registry loading. | intentional/no_action | View-schema registry and contract generation. | Preserve generated registry loading unless a later facade change is authorized. |
| Contract generated Go is used by many tests as evidence. | OpenAPI, WS, view-schema, error, extension tests import generated contracts. | Tests can accidentally treat generated artifacts as behavior authority. | should_fix | Owner specs and contract tests. | Keep tests as drift/evidence checks; ensure tracker says generated files are not owners. |
| Visibility and authorization are expressed in generated SQL queries. | Incident and saved-view SQL joins `incident_memberships`; modules wrap errors and payloads. | Moving SQL without authorization characterization could change public visibility. | should_fix | Incidents, savedviews, Core 04. | Add characterization before persistence facade changes. |
| Workbook projection SQL is adjacent to timeline logic. | Timeline projection provider reads source rows through generated SQL and derives projection inputs. | Projection/source confusion could cause behavior drift. | should_fix | Timeline plus workbook projections. | Keep projection source state and derived projection state distinct in future slices. |
| Reporting persistence couples SQLC rows to snapshot/release state machine. | Reporting store consumes generated reporting rows and maps release state. | Persistence movement can alter release idempotency, invalidation, approvals, or render-failure state. | should_fix | Reporting. | Add reporting lifecycle characterization before changing SQLC usage. |
| Generated files have hand-edit risk. | AGENTS, harness NLSpec, and generated artifact policy all prohibit hand-editing `internal/gen/**`. | Manual edits would be overwritten and violate policy. | must_fix | Harness generated-artifact policy. | Any future generated change must update owner inputs and run generators. |
| No direct grid-vendor coupling found in this target. | Search focused on `internal/gen` callers; no frontend grid adapter dependency found. | Low. | intentional/no_action | Frontend grid adapter if future contract changes affect UI. | No action for this target. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Record target, authority, generated-artifact posture, and allowed write boundary. | Tracker only. | None required for planning; optional markdown lint later. | Tracker section 1 complete. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03 | Inventory every file under `internal/gen` and inspect callers. | `internal/gen/**`, `db/queries/**`, representative module stores. | Read-only discovery; no validation claim. | Inventory table complete. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-04, WF-05 | Map generated contract and SQL outputs to owner specs, authored inputs, and consuming modules. | `contracts/*`, `tools/contractgen`, `sqlc.yaml`, `db/queries/*`. | `make generate-drift` only after owner input changes. | Contract map complete. |
| WF-03 | Characterization test gap analysis | chain | WF-01 | WF-04, WF-07 | Identify existing tests and missing characterization before future changes. | Module tests, contract tests, harness target guide. | `make backend-unit`, `make backend-store`, `make backend-integration` by slice. | Test posture recorded. |
| WF-04 | Boundary/coupling scan | chain | WF-02, WF-03 | WF-05, WF-06 | Classify generated root, SQLC coupling, contract embedding, and unused queries. | Callers in modules and platform registry. | Static/import-boundary checks if frontend contracts change. | Coupling findings complete. |
| WF-05 | Facade or ownership redesign plan | chain | WF-04 | WF-06 | Plan behavior-preserving ownership changes only where evidence supports them. | Module store facades; authored SQL inputs; contract owner inputs. | Slice-specific backend and drift checks. | Proposed owner per finding. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Sequence minimal behavior-preserving future work. | Tracker and later authorized implementation files. | Per-slice validation commands named. | Slice table complete. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Choose Make-owned validation and avoid promoting phase maps to runtime architecture. | Make target metadata and harness docs. | `make generated-artifact-policy-check`, `make generate-drift`, phase slices by risk. | Validation plan complete. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Leave another session able to continue without rediscovery. | Tracker only for this task. | No validation pass claimed unless run. | Handoff log updated. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker and record current generated-root posture. | `docs/handoffs/internal-code-gen-module-refactor-tracker.md`. | None to runtime; documentation drift only. | Preserve handoff structure. | Optional `make lint-markdown`; not required for production behavior. | Delete or revert tracker only. | Tracker contains all required sections and no production files changed. |
| S-01 | S-00 | Map generated SQL outputs to authored SQL and consuming module owners. | `db/queries/*`, `sqlc.yaml`, module stores; no edits in this slice unless later authorized. | SQL row shapes, visibility, persistence semantics. | Preserve module store tests; add characterization for any moved SQLC call. | `make backend-unit`; add `make backend-store` or `make backend-integration` when persistence behavior is touched. | Revert owner-map edits or facade-only changes before generator refresh. | Every generated SQL file has a module owner or TODO owner decision. |
| S-02 | S-00 | Map generated contract output to contract owners and generator inputs. | `contracts/*`, `tools/contractgen`, generated outputs through `make generate`. | OpenAPI, WS, view-schema, error, extension registries. | Preserve contract tests that load `internal/gen/contracts`. | `make generated-artifact-policy-check`; `make generate-drift`; targeted module contract tests. | Revert owner contract input change and regenerated outputs together. | Contract output remains generated and drift-free. |
| S-03 | S-01 | Characterize unused generated SQL queries before any deletion proposal. | `db/queries/auth_phase1.sql`, `db/queries/incidents_phase2.sql`, `db/queries/recovery_phase10.sql`, owners in auth/incidents/recovery. | Auth/session, incident membership summary, backup selection behavior. | Add or identify owner tests proving unused status or intended use. | `make backend-unit`; `make backend-store` if SQL/migration behavior is affected. | Keep authored SQL unchanged if evidence is incomplete. | Owner accepts keep/delete/adopt decision; otherwise TODO remains. |
| S-04 | S-01 | Plan module-owned persistence facades if SQLC types leak beyond store boundaries. | Module stores in incidents, timeline, savedviews, workbook startup, recovery, reporting. | Public route envelopes, authorization, projection rows, release state. | Add behavior-specific characterization before moving call sites. | `make backend-unit`; `make backend-integration`; phase-specific slice when applicable. | Revert facade changes before changing generated or authored SQL. | Callers depend on module DTOs/facades, not generated SQLC rows. |
| S-05 | S-02 | Plan generated drift validation for any future authored SQL or contract change. | `db/queries/*`, `contracts/*`, `internal/gen/**`, `packages/protocol-ts/src/generated/**`, `packages/ui-contracts/src/generated/**`. | Generated contract surfaces and downstream TypeScript consumers. | Preserve generated artifact policy tests and contract shape tests. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make frontend-import-boundary-check` when TS consumers change. | Revert owner inputs and generated outputs as one unit. | Generated outputs are produced by Make-owned generators only. |

Any slice that changes runtime behavior, public contracts, authored SQL, generated outputs, tests, or source code requires later authorization.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Fast backend unit and module tests. | yes | Use for module store/facade refactors that do not need services. |
| integration | `make backend-store`; `make backend-integration` | SQL, Postgres-backed store behavior, service-backed backend flows. | yes | Required for SQLC call movement or authored SQL changes. |
| e2e/browser | `make browser-e2e-webserver-backed`; `make phase-slice PHASE=phase8` when saved-view/workbook contracts are involved. | Browser-visible workbook and saved-view behavior. | no for tracker; yes for UI-visible contract changes | Use only when generated contract or workbook behavior affects browser flows. |
| generated drift | `make generated-artifact-policy-check`; `make generate-drift`; `make migration-drift` when migrations or authored SQL are touched. | Generated-file policy, generator reproducibility, migration/schema drift. | yes for generated or SQL changes | Do not hand-edit `internal/gen/**`; update owner inputs then generate. |
| import-boundary/static | `make frontend-import-boundary-check` | Frontend generated-contract consumer boundaries. | no for backend-only SQL slices; yes for frontend contract consumer changes | No direct grid-vendor coupling found in this target. |
| full check | `make test-fast`; `make check` | Local aggregate gates. | no for tracker; yes for high-risk cross-contract implementation | Broaden only after narrow checks cover the slice. |

Command discovery performed: `make help`, `make task-guide ROLE=feature-dev PHASE=phase8`, and `make explain-target` for `generate-drift`, `generated-artifact-policy-check`, `backend-unit`, `backend-store`, `backend-integration`, `migration-drift`, `frontend-import-boundary-check`, `test-fast`, and `check`.

No validation command has been run for this tracker unless recorded in the session log.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define `internal-code-gen` as tracker label and generated-root target, not a domain module. | WF-00 | DONE | none | Section 1 source posture. | Scope and non-goals are explicit. |
| T-002 | Inventory every file under `internal/gen`. | WF-01 | DONE | T-001 | Section 2 inventory table. | All 12 files are listed or marked out of behavior scope. |
| T-003 | Map generated SQL files to authored SQL and consuming owner modules. | WF-02 | DONE | T-002 | Sections 2 and 3. | Each generated SQL file has an owner candidate or TODO/defer finding. |
| T-004 | Map generated contract Go to contract owners and runtime/test consumers. | WF-02 | DONE | T-002 | Sections 2, 3, and 4. | Contract registry risks and consumers are recorded. |
| T-005 | Identify public contract freeze risks. | WF-03 | DONE | T-003,T-004 | Section 4. | HTTP, WS, view-schema, SQL-backed, auth, revision, projection, and harness risks are listed. |
| T-006 | Classify coupling and boundary findings. | WF-04 | DONE | T-005 | Section 5. | Findings have classification, owner, and action. |
| T-007 | Seed future behavior-preserving workstreams. | WF-06 | DONE | T-006 | Section 6. | Workflows include dependencies and handoff checkpoints. |
| T-008 | Seed future implementation slice plan. | WF-06 | DONE | T-007 | Section 7. | Slices are behavior-preserving or marked as requiring later authorization. |
| T-009 | Discover Make-owned validation commands. | WF-07 | DONE | T-008 | Section 8. | Commands are named or scoped; no pass is claimed. |
| T-010 | Record handoff log for this planning-only session. | WF-08 | DONE | T-009 | Section 10. | Handoff entries list files, commands, blockers, and next action. |
| T-011 | Resolve owner decision for unused generated auth/session/recovery queries. | WF-05 | DEFERRED | T-003 | RB-001, RB-002. | Later owner decision accepts keep, adopt, or remove through authored inputs. |
| T-012 | Plan any module facade refactor for direct SQLC type coupling. | WF-05 | TODO | T-011 | S-04. | Later plan names exact facades and characterization tests. |

## 10. Session Handoff Log

### Scope and Authority Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | Tracker created for generated-root target; no production refactor performed. | Touched `docs/handoffs/internal-code-gen-module-refactor-tracker.md`; inspected framework, Core docs, domain doc, harness NLSpec, adopted/draft NLSpecs. | `sed`, `find`, `rg`, `git status`, `date -Is`, `git rev-parse`. | Authority posture recorded; no owner contradiction found. | None for tracker creation. | Later authorized implementation task may start with S-01 or S-03. |

### Backend Module Boundary Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | `internal/gen/sql` is generated persistence adapter output consumed by incidents, timeline, savedviews, workbook startup, recovery, and reporting. | Touched tracker only; inspected generated SQL, authored SQL, `sqlc.yaml`, representative module stores and platform auth store. | `rg` for generated imports and query names; `sed` for exact source files; `wc`; `find`. | Generated SQL owners mapped; unused generated query finding recorded. | RB-001 and RB-002 owner decisions remain deferred. | Characterize unused queries before proposing cleanup. |

### Frontend Module Boundary Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | No direct frontend shell/controller or grid-vendor code lives under `internal/gen`. Generated contracts may affect frontend consumers through protocol/view-contract packages. | Touched tracker only; inspected contract registry paths and frontend import-boundary target metadata. | `rg`, `jq`, `make explain-target TARGET=frontend-import-boundary-check DETAIL=summary`. | Frontend impact is contract-consumer risk only. | No frontend implementation blocker for this tracker. | Use frontend import-boundary and browser checks only if generated contract consumers change later. |

### Contract and Codegen Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | `internal/gen/contracts/contracts_gen.go` is generated by `tools/contractgen` from `contracts/*`; `internal/gen/sql` is generated by SQLC from `db/queries` and migrations. | Touched tracker only; inspected `contracts/*`, `tools/contractgen/main.go`, generated-artifact scripts, `tools/generated_artifact_policy.json`, `sqlc.yaml`. | `sed`, `jq`, `make explain-target TARGET=generate-drift DETAIL=summary`, `make explain-target TARGET=generated-artifact-policy-check DETAIL=summary`, `make explain-target TARGET=migration-drift DETAIL=summary`. | Codegen owner inputs and drift commands recorded. | No generated output may be hand-edited. | For any later codegen change, update owner input and run Make-owned generators/checks. |

### Tests and Harness Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | Validation plan names Make-owned targets but no validation pass is claimed. | Touched tracker only; inspected harness NLSpec and target metadata. | `make help`; `make task-guide ROLE=feature-dev PHASE=phase8`; `make explain-target` for relevant targets. | Narrow and broad validation commands recorded. | No validation was required for tracker-only write. | Run validation only in later implementation task or markdown lint if tracker maintenance requires it. |

### Security and Authorization Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | Generated SQL contains visibility/access SQL for incidents and saved views, but Core 04 and modules own authorization behavior. | Touched tracker only; inspected incident/savedview SQL and stores, platform auth store, Core 04 excerpts. | `rg`, `sed`. | Authorization risk recorded as behavior-freeze item. | Any SQL movement needs authorization characterization. | Preserve visibility outcomes before future persistence facade changes. |

### Open Risks and Next Session Handoff

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-08T20:48:13-04:00 | Codex planning and tracker creation | Tracker complete for planning handoff; future implementation is deferred. | Touched tracker only. | `git status --short --branch`; `git rev-parse --short HEAD`; discovery commands listed above. | Branch `main` at `7c5868f7`; working tree was clean before tracker creation. | RB-001 and RB-002 remain open. | Start with S-03 if cleanup is desired, or S-04 if facade refactor is desired. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Should generated auth SQL in `auth_phase1.sql.go` be adopted by platform auth, retained unused, or removed by deleting authored SQL queries? | Current auth persistence uses raw SQL, creating drift risk between generated query inputs and live behavior. | Auth/platform owner decision; characterization of session/bootstrap behavior; authored SQL review. | DEFERRED |
| RB-002 | Are uncalled generated queries such as `GetSessionMemberships` and `GetLatestSuccessfulRetainedBackupSet` intentionally retained? | Removing them requires authored SQL changes and regenerated outputs; retaining them should be intentional. | Incidents/recovery owner decision; search confirmation in implementation session. | DEFERRED |
| RB-003 | Should modules hide SQLC row/param types behind narrower persistence facades? | SQLC types currently appear in module conversion helpers and can couple code to generated schema shape. | Module owner review and characterization tests for affected stores. | TODO |
| RB-004 | What is the exact owner boundary for workbook projection SQL versus timeline module logic? | Timeline generated SQL reads projection/source rows, but projection behavior is Core-owned and graph projection is explicitly out of scope. | Core 01/Core 03 owner reading and timeline/projections design review. | TODO |

## 12. Binary Completion Criteria

The tracker is complete only when all criteria pass:

- Every file in `internal/gen` is inventoried or explicitly out of scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`.
- Validation commands are discovered or marked `TODO` with a reason.
- Contradictions are marked `BLOCKED: owner contradiction`.
- Repository/framework mismatches are recorded as planning findings.
- Handoff sections are current enough for another agent to continue without rediscovery.

Completion status for this tracker:

- Files inventoried: yes.
- Contract risks mapped: yes.
- Workflows and slices seeded: yes.
- Behavior changes proposed: no.
- Validation commands discovered: yes.
- Owner contradictions found: none.
- Repository/framework mismatch recorded: yes, `internal/gen` is generated output rather than a domain module.
- Handoff current: yes, as of 2026-07-08T20:48:13-04:00.
