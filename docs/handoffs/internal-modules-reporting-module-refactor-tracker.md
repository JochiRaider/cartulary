# internal-modules-reporting Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Target path: `internal/modules/reporting`.

Target label: `internal-modules-reporting`.

Output path: `docs/handoffs/internal-modules-reporting-module-refactor-tracker.md`.

Status: planning and documentation only. This tracker records current-state evidence, boundary findings, workstreams, and safe future slices. Implementation requires a later authorized task.

Allowed changes for this session: this tracker file only. No production code, tests, contracts, generated artifacts, package configuration, migrations, harness files, lockfiles, or generated outputs may be edited in this planning task.

Non-goals:

- Do not perform a production refactor.
- Do not change HTTP routes, generated contracts, SQL queries, migrations, or harness maps.
- Do not hand-edit generated roots declared by `tools/generated_artifact_policy.json`.
- Do not assume `internal-modules-reporting` is a permanent module boundary merely because `internal/modules/reporting` exists.
- Do not promote phase maps or test rows into runtime architecture.

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
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/report-composition-nlspec.md`
- `docs/domain.md`
- `docs/guides/cartulary_report_builder_implementation_stack_guide.md`
- `docs/testing-harness-nlspec.md`

Repository files inspected:

- `AGENTS.md`
- `internal/modules/reporting/.gitkeep`
- `internal/modules/reporting/api.go`
- `internal/modules/reporting/routes.go`
- `internal/modules/reporting/store.go`
- `internal/modules/reporting/redaction.go`
- `internal/modules/reporting/boundary_guard_test.go`
- `internal/modules/reporting/openapi_contract_test.go`
- `internal/modules/reporting/redaction_test.go`
- `internal/modules/reporting/reporting_integration_test.go`
- `internal/modules/reporting/traceability_test.go`
- `internal/app/runtime.go`
- `internal/platform/httpapi/extensions.go`
- `internal/modules/reportcomposition/routes.go`
- `internal/modules/reportcomposition/types.go`
- `internal/modules/reportcomposition/store.go`
- `db/queries/reporting_phase11.sql`
- `db/migrations/00025_phase11_snapshot_reporting.sql`
- `db/migrations/00026_phase11_reporting_remediation.sql`
- `db/migrations/00047_reporting_current_output_kinds.sql`
- `db/migrations/00048_report_composition_release_tuple.sql`
- `contracts/openapi/cartulary.openapi.yaml`
- `contracts/extensions/index.json`
- `contracts/reporting/fixtures/corpus.v1.json`
- `tools/generated_artifact_policy.json`
- `tools/schema_object_ownership_manifest.json`
- `tools/phase11_test_map.json`
- `docs/testing/phase11_coverage_ledger.md`
- `packages/protocol-ts/src/index.test.ts`
- `apps/web/e2e/phase2.support.spec.ts`
- `apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts`

Current session baseline:

- Branch: `main`
- Commit: `7900ff1edafed763d19236c2b488565712b7c0e3`
- Dirty tree before tracker write: clean
- Session time: `2026-07-06T20:10:22-04:00`

Planning findings:

- The planning framework references older Core footnote filenames. The live repository uses `01_architecture_storage_and_view_contracts.md`, `02_domain_model_schema_and_history.md`, `03_workbook_interaction_collaboration_and_workflows.md`, `04_security_deployment_and_conformance.md`, and `05_claim_publication_and_benchmark_reproducibility.md`.
- No owner contradiction was found during this pass.
- `internal/modules/reporting` exists and contains current implementation and tests. It is not blocker-only.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/reporting/.gitkeep` | Placeholder file only. | None. | None found. | None. | None. | None. | Out of scope: repository placeholder. | Low | Keep only if the directory would otherwise be empty. |
| `internal/modules/reporting/api.go` | Snapshot/release request DTOs, strict JSON decoders, closed vocabulary constants, output-option materialization, graph projection ref structural validation, composition tuple structural validation, recipient partition validation, API error builders. | `ProfileID`, derivation/schema/output/release constants, `CreateSnapshotRequest`, `CreateReleaseRequest`, `ReleaseActionRequest`, `DecodeCreateSnapshotRequest`, `DecodeCreateReleaseRequest`, `DecodeReleaseActionRequest`. | `routes.go`, `store.go`, `redaction.go`, unit tests, integration tests. | Standard JSON/hash/regexp packages, `uuid`, `internal/platform/httpapi`, reporting redaction/template helpers in same package. | `redaction_test.go`, `reporting_integration_test.go`, `openapi_contract_test.go` indirectly. | OpenAPI request/response schema and enum surfaces must stay aligned; no generated file is authored here. | Reporting facade plus route-contract helper. | High | Public behavior includes exact fields, defaulting, reason codes, idempotency normalization, graph ref shape, and composition tuple rules. |
| `internal/modules/reporting/routes.go` | HTTP route registration, service assembly, authentication, incident role checks, snapshot/release route handlers, release action handlers, render job dispatch/execution, job completion summaries, route parsing, route-scoped visibility errors. | `Service`, `RegisterRoutes`; unexported route handlers and job helpers. | `internal/app/runtime.go` appends `reporting.RegisterRoutes()` when building runtime. | `internal/modules/incidents`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`, `internal/platform/jobs`, `internal/platform/ws`, `uuid`, time/context/http. | `reporting_integration_test.go`; route contract also checked by `openapi_contract_test.go`. | Public OpenAPI paths and job resource refs; no generated files authored here. | Reporting facade with transport-adjacent adapter. | High | Contains legitimate route facade but also in-process job orchestration; incident module import is currently allowed only here by boundary guard. |
| `internal/modules/reporting/store.go` | SQL-backed snapshot/release persistence, route idempotency lookup/insert, job payload persistence, source-boundary resolution, export-field collection over live source/projection tables, composition tuple DB validation, graph projection run validation, release lifecycle state transitions, approval records, resource envelope construction, pgtype conversion helpers. | `ErrNotFound`, `ErrStateConflict`, `ErrApprovalRejected`, error structs, `Store`, record/params/result structs, `NewStore`, store methods for create/get/complete/action flows. | `routes.go`, `reporting_integration_test.go`, `redaction_test.go` indirectly through helpers. | `pgx`, `pgxpool`, generated SQLC package `internal/gen/sql`, `internal/platform/authn`, `internal/platform/jobs`, direct raw SQL over `incidents`, `change_sets`, workbook/source tables, projections, report composition tables, graph projection tables. | `reporting_integration_test.go`, `redaction_test.go` indirectly, generated SQLC compile coverage. | SQLC generated code from `db/queries/reporting_phase11.sql`; schema from reporting/composition/projection migrations. | Reporting persistence plus source/export projection orchestrator. | High | Main mixed-responsibility hotspot. Direct cross-table reads should be hidden behind stable source/export and owner validation facades before movement. |
| `internal/modules/reporting/redaction.go` | Reporting data contract structs, export-model construction, redaction profile registry, redaction application and manifest validation, template contract resolution, simplified rendering, self-contained output validation, canonical JSON/hash helpers. | Redaction/profile/content constants; error vars; `RedactionProfile`, `ExportModel`, `ExportField`, `RedactionManifest`, `TemplateContract`, `IncidentMetadataSnapshot`; `BuildExportModel`, `DefaultRedactionProfileRegistry`, `ResolveRedactionProfile`, `ValidateRedactionProfile`, `RedactExportModel`, `ValidateRedactionResult`, `ResolveTemplateContract`, `RenderOutput`, `ValidateSelfContainedOutput`. | `routes.go`, `store.go`, `api.go`, unit tests, integration tests. | Standard JSON/regexp/sort/string/time packages only; same-package constants. | `redaction_test.go`, `reporting_integration_test.go`. | Reporting fixture corpus and acceptance evidence rely on these schemas/behaviors; generated OpenAPI surfaces expose related output/release values. | Reporting render/redaction internals. | High | Current renderer is simplified relative to the full Reporting NLSpec; preserve behavior until deeper render-stack authorization exists. |
| `internal/modules/reporting/boundary_guard_test.go` | Import-boundary guard over production Go files in reporting. | Test-only helpers. | `go test ./internal/modules/reporting`, phase 11 backend-unit evidence. | `go/parser`, `go/token`, filesystem helpers. | Self-contained test. | Phase ledger/map evidence only; no generated artifacts. | Reporting tests and architecture guardrail. | Medium | Allows `routes.go` to import `internal/modules/incidents`; does not currently ban raw SQL coupling in `store.go`. |
| `internal/modules/reporting/openapi_contract_test.go` | Generated OpenAPI contract characterization for release enums, exact snapshot/release resources, and reporting paths. | Test-only helpers in external `reporting_test` package. | `go test ./internal/modules/reporting`, phase 11 backend-unit evidence. | `internal/gen/contracts`, JSON/reflection testing helpers. | Self-contained test. | Reads generated OpenAPI artifact from `internal/gen/contracts`; verifies `contracts/openapi/cartulary.openapi.yaml` content after generation. | Reporting contract evidence. | Medium | Generated artifact is read-only evidence; do not hand-edit generated output to satisfy this test. |
| `internal/modules/reporting/redaction_test.go` | Unit characterization for redaction precedence/actions/manifests, invalid profile cases, external validation, disclosure partitions, export model hash stability, decoders/defaults/registered reason codes, composition tuple and graph ref structural validation, self-contained output, render failure reasons. | Test-only package-level fixtures and helpers. | `go test ./internal/modules/reporting`, phase 11 backend-unit evidence. | `reporting` package internals, JSON/errors/testing/time/uuid. | Self-contained test. | Phase 11 ledger/map evidence. | Reporting characterization tests. | High | Important pre-move evidence for public reason codes and defaults. |
| `internal/modules/reporting/reporting_integration_test.go` | Service-backed characterization for snapshot replay/provenance, source boundary conflicts, export model coverage, recipient redaction isolation, release approval/publish/invalidate lifecycle, idempotency, render-failed releases, hidden resource behavior, exact resource shapes. | Test-only helpers and fixtures in external `reporting_test` package. | `go test ./internal/modules/reporting`, phase 11 backend-integration evidence. | `internal/testutil/phase2test`, `httptestx`, `authn`, live runtime routes, direct DB fixture setup. | Self-contained integration test. | Phase 11 ledger/map, schema/migration behavior, OpenAPI shape indirectly. | Reporting service characterization. | High | Broad behavioral freeze. Direct DB seeding mirrors current source/projection tables and should be preserved until owner fixtures exist. |
| `internal/modules/reporting/traceability_test.go` | Reporting NLSpec requirement, fixture, acceptance, coverage, and output-kind traceability checks. | Test-only helpers. | `go test ./internal/modules/reporting`, phase 11 backend-unit evidence. | Reads `docs/reporting-subsystem-nlspec.md` and `contracts/reporting/fixtures/corpus.v1.json`; regexp/JSON/filesystem helpers. | Self-contained test. | Reporting fixture corpus and phase evidence accounting. | Reporting evidence accounting. | Medium | Evidence/accounting only; does not define runtime architecture. |

## 3. Module Boundary Diagnosis

Diagnosis: `internal/modules/reporting` is a mixed-responsibility package. It contains a legitimate thin application facade for the Snapshot and Reporting route family, but it also owns transport-adjacent routing, persistence-adjacent SQL, source/export projection orchestration, release mutation coordination, composition and graph consumer validation, redaction/render logic, and contract evidence. It is not currently a frontend shell/controller surface or grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Snapshot/release HTTP facade | `routes.go`, `api.go` | Reporting facade under Core 01 route contract | Keep/split | `RegisterRoutes()` registers `/api/v1/snapshots` and `/api/v1/releases`; Core 01 §17.3 owns public route contract. | Keep public facade stable; split internal service/decode/store seams only. |
| Authentication/session sliding and incident role checks | `routes.go` | Transport/auth platform plus reporting route facade | Keep/split | `httpauth.AuthenticateRequest`, `SlideSessionIfNeeded`, `incidents.RequireIncidentRole`. | Route layer may call platform/auth; domain internals should not spread auth checks. |
| In-process async job orchestration | `routes.go`, `store.go` | Reporting facade plus `internal/platform/jobs` substrate | Split | `dispatchReportingJob`, `executeReportingJob`, `reporting_job_payloads`. | Preserve common job semantics; consider service boundary between route and worker. |
| Snapshot source-boundary resolution | `store.go` | Reporting snapshot facade with source-state ports | Split | `resolveSourceBoundaryTx` reads `incidents` and `change_sets`. | Boundary token is reporting/Core-owned, but source reads should be abstracted. |
| Export model construction from source records/projections | `store.go`, `redaction.go` | Reporting export collector with owner-provided source ports; projections/source modules own their tables | Split | `collectWorkbookExportFieldsTx` queries records, timeline, parties, evidence, projections, links, tags, mentions. | High coupling; behavior must be characterized before moving. |
| Redaction profile, manifest, output validation | `redaction.go` | Reporting render/redaction internals | Keep/split | `RedactExportModel`, `ValidateRedactionResult`, external validation tests. | Legitimate reporting responsibility; split into cohesive files/packages only if surface stays stable. |
| Template contract and simplified rendering | `redaction.go`, `routes.go` | Reporting render internals | Defer | `ResolveTemplateContract`, `RenderOutput`, `renderReleaseCandidate`. | Current implementation is simplified relative to NLSpec; deeper render-stack work needs later authorization. |
| Release persistence, approval lifecycle, publish/invalidate state | `store.go`, `routes.go` | Reporting release service/store | Split | `CreateRelease`, `ApproveRelease`, `PublishRelease`, `InvalidateRelease`. | Preserve exact resource shapes and reason codes while separating store from route orchestration. |
| Composition tuple validation and binding | `api.go`, `store.go`, `reportcomposition` facade | Reporting consumer boundary plus `reportcomposition` authoring owner | Split/remediated | `composition_id/version/sha256` decode stays in Reporting; DB tuple resolution and binding now flow through `reportcomposition.ResolveReleaseTupleTx` and `reportcomposition.BindReleaseTupleTx`. | Do not move authoring lifecycle into reporting; keep route tuple consumption behind the owner facade. |
| Graph projection ref validation | `api.go`, `store.go`, `graphprojection` facade | Reporting consumer boundary plus projections owner | Split/remediated | `validateSourceProjectionRefs` stays in Reporting; DB run validation now flows through `graphprojection.ValidateReportingProjectionRefsTx`. | Reporting consumes completed projection refs; graph lifecycle remains projection-owned. |
| OpenAPI/fixture/phase evidence | `openapi_contract_test.go`, `traceability_test.go`, phase maps | Contract/evidence accounting | Keep/defer | Tests read generated OpenAPI and fixture corpus. | Evidence maps are not runtime architecture. |
| Frontend shell/controller state | None found in target | `/apps/web` if future UI uses reporting routes | Defer | Search found no authored frontend caller for snapshot/release routes. | Current frontend hits are extension registry and unrelated decision enum. |
| Grid-vendor integration | None found in target | `/packages/grid-adapter` | Intentional/no_action | No `react-data-grid` or grid adapter imports found in reporting package. | No action for this target. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `POST /api/v1/snapshots` | Core 01 route contract; reporting implementation | `routes.go`, `api.go`, Core 01 §17.3 | `reporting_integration_test.go`, `openapi_contract_test.go` | Preserve request decoding, source-boundary omission/default/replay, accepted job shape. | High | Route is profile-gated by `snapshot_reporting`. |
| `GET /api/v1/snapshots/{snapshot_id}` | Core 01 route contract; Core 04 visibility | `routes.go`, `store.go`, Core 01 §17.3 | `reporting_integration_test.go`, `openapi_contract_test.go` | Preserve singleton pagination rejection, hidden snapshot mapping to `snapshot_not_found`, exact resource keys. | High | Resource must not expose rendered bytes or approval data. |
| `POST /api/v1/releases` | Core 01 route contract; Reporting render admission | `routes.go`, `api.go`, `store.go`, Core 01 §17.3 | `redaction_test.go`, `reporting_integration_test.go`, `openapi_contract_test.go` | Add focused characterization before splitting graph/composition validation if facade changes. | High | Selector validation order must preserve hidden snapshot behavior. |
| `GET /api/v1/releases/{release_id}` | Core 01 route contract; Core 04 visibility | `routes.go`, `store.go` | `reporting_integration_test.go`, `openapi_contract_test.go` | Preserve exact nullable/non-nullable resource shape and hidden release behavior. | High | Release resource excludes full manifest and rendered bytes. |
| Release actions `approve`, `publish`, `invalidate` | Core 01 action contract; Core 04 release gate | `routes.go`, `store.go`, Core 04 §2.1 | `reporting_integration_test.go` | Preserve idempotent replay, reason normalization, role rules, state conflicts, action response parity with GET. | High | `publish` and `invalidate` require visible release plus admin role. |
| Common success/error envelopes | Core 01 common API; reporting route family error registry | `api.go`, `routes.go`, Core 01 REQ-01-479 | `reporting_integration_test.go`, `redaction_test.go`, `openapi_contract_test.go` | Preserve family error code set and reason-code placement. | High | Hidden resource errors must not leak incident-level not-found. |
| Background job summaries | Core 01 jobs; reporting implementation | `routes.go`, `store.go`, `internal/platform/jobs` | `reporting_integration_test.go` | Preserve `snapshot_created`, `release_created`, `release_render_failed`, one resource ref. | High | No reporting-specific WebSocket route found; job progress uses shared infrastructure. |
| Snapshot export model | Reporting Subsystem NLSpec; Core 01 snapshot route | `store.go`, `redaction.go`, `reporting-subsystem-nlspec.md` | `reporting_integration_test.go`, `redaction_test.go` | Add pre-move tests if source collector is split by owner/table family. | High | Must exclude raw evidence/storage fields and preserve support refs/partitions. |
| Redaction profile and manifest behavior | Reporting Subsystem NLSpec; Core 04 security controls | `redaction.go`, Core 04 §2.1 | `redaction_test.go`, `reporting_integration_test.go` | Preserve action precedence, digest, manifest entries, external fail-closed checks. | High | External release cannot change live workbook visibility. |
| Render output and self-contained checks | Reporting Subsystem NLSpec | `redaction.go`, `routes.go` | `redaction_test.go`, `reporting_integration_test.go` | TODO: add broader render-stack characterization before replacing simplified renderer. | High | Current implementation renders Markdown/Mermaid placeholders, not full Slidev bundle stack. |
| Composition tuple consumption | Core 01 release tuple; Report Composition NLSpec authoring; Reporting consumption | `api.go`, `store.go`, `reportcomposition` package | `redaction_test.go`, `reportcomposition/release_tuple_facade_test.go`; integration coverage via release tuple paths is partial. | Add service-backed route coverage only if route error mapping or tuple admission order changes. | High | Reporting must not own composition authoring lifecycle. |
| Graph projection refs | Core 01 release tuple; Graph Projection owner; Reporting consumer | `api.go`, `store.go`, `graphprojection` package | `redaction_test.go`, `graphprojection/store_test.go::TestValidateReportingProjectionRefsTxReasonMatrix`. | Add service-backed route coverage only if route error mapping or graph-ref admission order changes. | High | Reporting consumes completed digest-bound projection refs only. |
| Authorization checks | Core 04; incidents module membership/role helpers | `routes.go`, `incidents` access helper | `reporting_integration_test.go` | Preserve route-scoped visibility and admin/reviewer/editor distinctions. | High | `deployment_admin` alone was not inspected as a reporting-specific route behavior in this pass. |
| Generated OpenAPI | Contract layer downstream of specs | `contracts/openapi/cartulary.openapi.yaml`, `internal/gen/contracts` | `openapi_contract_test.go` | Use `make generate`/`make generate-drift`; never hand-edit generated roots. | Medium | Authored OpenAPI source owner not changed in this tracker. |
| SQLC generated queries | Authored SQL input plus generated `internal/gen/sql` | `db/queries/reporting_phase11.sql`, `internal/gen/sql/reporting_phase11.sql.go` | Compile/tests. | Use `make generate` and drift checks if authored SQL changes. | Medium | Generated SQLC output must not be hand-edited. |
| Harness/test accounting | `docs/testing-harness-nlspec.md`, phase maps/ledgers | `tools/phase11_test_map.json`, `docs/testing/phase11_coverage_ledger.md` | `traceability_test.go`, phase targets | Preserve evidence rows; update owner inputs before generated ledgers/schedules. | Medium | Evidence accounting is not runtime architecture. |
| Frontend selector/grid contracts | `/apps/web`, `/packages/grid-adapter`, generated TS contracts | Search found no snapshot/release authored frontend caller. | Extension registry and workbook surface tests only. | TODO: characterize frontend reporting UI if a future route caller is added. | Low | No grid-vendor integration in reporting target. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Direct cross-table export SQL couples reporting to source/projection schemas. | `collectWorkbookExportFieldsTx` reads `records`, `timeline_events`, `host_grid_projection`, `identity_grid_projection`, `parties`, `evidence`, task/decision/artifact projections, links, tags, mentions. | High | `should_fix` | Reporting export collector plus owner-provided source ports. | Plan a source/export collection seam before moving SQL. |
| Graph projection validation reads projection tables directly. | Remediated: reporting now calls `graphprojection.ValidateReportingProjectionRefsTx`; `boundary_guard_test.go` forbids `graph_projection_runs` in reporting production files. | High | `done` | Projections owner facade consumed by reporting. | Keep facade tests current when graph projection lifecycle states or reason codes change. |
| Composition tuple validation and binding reach into composition tables. | Remediated: reporting now calls `reportcomposition.ResolveReleaseTupleTx` and `reportcomposition.BindReleaseTupleTx`; `boundary_guard_test.go` forbids composition owner table names in reporting production files. | High | `done` | Report composition authoring owner plus reporting release consumer boundary. | Keep facade tests current when tuple/version binding semantics change. |
| Route admission, idempotency, store mutations, and state transitions are split across `routes.go` and `store.go`. | `handleReleasesCollection` validates visibility then store create; store does route idempotency and state mutation. | High | `should_fix` | Reporting service facade with private store. | Plan service/store interfaces preserving exact route behavior. |
| Store depends on platform auth idempotency and jobs directly. | `store.go` imports `internal/platform/authn` and `internal/platform/jobs`. | Medium | `should_fix` | Reporting service/store with platform adapter ports. | Decide whether route idempotency remains store-local or moves behind a small port. |
| Render/redaction/template responsibilities are concentrated in one file. | `redaction.go` defines data contracts, profile registry, redaction, template resolution, rendering, asset checks. | Medium | `should_fix` | Reporting render/redaction internals. | Split only after characterization preserves exported symbols and reason codes. |
| Current renderer is simplified relative to adopted Reporting NLSpec. | `RenderOutput` emits Markdown or fixed Mermaid text; Reporting NLSpec describes deterministic Slidev/Mermaid bundle/render behavior. | High | `defer` | Reporting render stack. | Record as architectural finding; do not replace renderer in this refactor plan without later behavior authorization. |
| Import-boundary guard is narrow and file-specific. | Remediated for current production imports: `boundary_guard_test.go` requires every reporting production sibling-module import to be explicitly allowed and blocks direct graph/composition owner table references. | Medium | `done` | Reporting architecture tests. | Extend the raw table guard again after export providers replace source/projection SQL. |
| Generated artifacts are contract evidence, not edit targets. | `openapi_contract_test.go` reads `internal/gen/contracts`; policy declares `internal/gen/**` generated. | Medium | `intentional/no_action` | Generation pipeline. | Keep generated outputs read-only; use `make generate`/drift checks in later implementation. |
| Release redaction must not become live workbook authorization. | Core 04 §2.1, domain §13.13, integration test compares live workbook rows before/after partitioned release. | High | `intentional/no_action` | Core 04 and reporting release pipeline. | Freeze behavior in characterization tests before source/export movement. |
| No direct grid-vendor coupling found. | Reporting package imports no frontend/grid packages; searches found no `react-data-grid` in target. | Low | `intentional/no_action` | `/packages/grid-adapter` if ever applicable. | No reporting-specific action. |
| No reporting-specific WebSocket route found. | `routes.go` uses `WSHub` only through jobs/progress plumbing; collaboration owns `/ws` behavior. | Low | `defer` | Collaboration/jobs infrastructure. | Treat job progress as shared contract; do not invent reporting WebSocket semantics. |
| Test fixture setup mirrors source/projection table shapes. | `reporting_integration_test.go` seeds many source/projection tables directly. | Medium | `defer` | Test harness and owner fixtures. | Preserve for now; consider owner fixture helpers in later test cleanup. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `WF-00` | Session/source bootstrap and tracker initialization | root | none | `WF-01` | Record target, authority, dirty tree, constraints, and owner docs. | This tracker, owner docs. | `make lint-markdown` for tracker-only updates. | Target label/path and planning-only limits are explicit. |
| `WF-01` | Current-state target inventory | chain | `WF-00` | `WF-02`, `WF-03`, `WF-04` | Inventory every file in `internal/modules/reporting`. | All files under target path. | `make phase-slice PHASE=phase11` after implementation changes; not required for tracker-only. | Inventory table covers every file or marks scope. |
| `WF-02` | Contract-owner mapping | chain | `WF-01` | `WF-04`, `WF-05` | Map HTTP, job, export, redaction, release, composition, graph, and generated contracts to owners. | Reporting docs, Core 01, Core 04, domain, OpenAPI, migrations. | `make generated-artifact-policy-check`, `make json-shape-check` when contract inputs are touched. | Freeze map has owner and test posture. |
| `WF-03` | Characterization test gap analysis | parallel | `WF-01` | `WF-05`, `WF-06` | Identify existing evidence and missing pre-move tests. | Reporting tests, phase map, ledger. | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11`. | Missing graph/composition/render gaps are explicit. |
| `WF-04` | Boundary/coupling scan | chain | `WF-02` | `WF-05` | Classify direct SQL, platform, auth, jobs, projection, composition, generated, frontend, and grid coupling. | `api.go`, `routes.go`, `store.go`, `redaction.go`, boundary tests. | `make lint`; future targeted guard tests. | Findings table has classifications and owners. |
| `WF-05` | Facade or ownership redesign plan | chain | `WF-04`, `WF-03` | `WF-06` | Choose seams for route facade, source/export collector, store, render/redaction, graph/composition consumers. | Reporting package and adjacent owner packages. | TODO: command discovery required for a reporting-only package target. | Proposed owners and deferrals are explicit. |
| `WF-06` | Slice sequencing plan | chain | `WF-05` | `WF-07` | Sequence behavior-preserving slices with dependencies, rollback notes, and stop conditions. | Tracker and later implementation files. | Slice-specific plus `make test-fast` before broader handoff. | Each slice has completion criterion. |
| `WF-07` | Harness/test/accounting update plan | chain | `WF-06` | `WF-08` | Plan required Make-owned validations and evidence updates. | Phase maps/ledgers only if later changes need them. | `make agent-finalize`, drift targets, phase slices. | Validation table is current and commands are not invented. |
| `WF-08` | Validation and final handoff | chain | `WF-07` | none | Run selected validation, record results, blockers, and next action. | Tracker and changed implementation files in later tasks. | `make check` when risk warrants. | Session log is current enough to resume. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `S-00` | none | Create or update this planning tracker only. | `docs/handoffs/internal-modules-reporting-module-refactor-tracker.md` | Documentation drift only. | Preserve existing handoff history if any appears. | `make lint-markdown` | Revert only this tracker file. | Tracker has 12 required sections and no production files changed. |
| `S-01` | `S-00` | Confirm characterization gaps before movement. | `internal/modules/reporting/*_test.go`, phase map/ledger as evidence only. | None if tests are read-only; later test additions need authorization. | Preserve existing unit/integration/OpenAPI/traceability tests; TODO graph/composition DB cases. | `make phase-slice PHASE=phase11` | Drop only newly added characterization tests if introduced later. | Existing and missing behavior evidence is known. |
| `S-02` | `S-01` | Split request DTOs/decoders/error helpers behind stable reporting package surface. | `internal/modules/reporting/api.go` and new private reporting files if authorized. | Exact JSON, defaults, reason codes, OpenAPI alignment. | Preserve `TestPhase11_U_11_REPORTING_05_DecoderNormalizationAndRegisteredReasons`. | `make phase-slice PHASE=phase11` | Restore original file layout; no schema changes. | Public constants/types/functions remain source-compatible or callers are updated internally without behavior change. |
| `S-03` | `S-01` | Introduce reporting-owned source snapshot/export collector seam before moving direct cross-table SQL. | `store.go`, possible private collector file, owner source/projection packages later. | Export ordering, support refs, disclosure partitions, raw evidence omission, source-boundary replay. | Preserve `TestPhase11_I_11_REPORTING_01_SnapshotReplayAndReleaseProvenanceAreStable`; add table-family characterization if moving SQL. | `make service-backed-slice PHASE=phase11` | Keep old SQL path until new collector passes parity; remove new seam if parity fails. | Snapshot export JSON and hashes match current behavior for characterized fixtures. |
| `S-04` | `S-02` | Isolate release persistence and job payload behavior behind a store/service interface without schema changes. | `routes.go`, `store.go`, private store/service files. | Idempotency, action replay, release states, job summaries, resource shape. | Preserve reporting integration tests 02, 03, and 04. | `make service-backed-slice PHASE=phase11` | Restore direct route-to-store calls. | Route handlers depend on a narrower facade and all resource/action behavior is unchanged. |
| `S-05` | `S-03`, `S-04` | Delegate composition and graph validation through owner facades. | `store.go`, `internal/modules/reportcomposition`, `internal/modules/projections` if later authorized. | Composition tuple not-found/mismatch/digest reason codes; graph incomplete/not-bound/ambiguous/digest reasons. | TODO: add DB-backed characterization for graph/composition mismatch cases before moving. | `make phase-slice PHASE=phase11`; `make service-backed-slice PHASE=phase11` | Fall back to raw SQL validation until facade parity is proven. | Raw table coupling is removed or explicitly deferred with tests preserving current results. |
| `S-06` | `S-02`, `S-04` | Separate redaction, template, and render internals while preserving exported constants/types. | `redaction.go`, possible private redaction/render/template files. | Profile digest, manifest hash, render failure reasons, remote asset rejection. | Preserve all `redaction_test.go` cases. | `make phase-slice PHASE=phase11` | Restore single-file implementation if exported surface or hashes drift. | Cohesive internals with unchanged unit and integration results. |
| `S-07` | `S-03`, `S-05` | Update import-boundary guards and evidence maps only through authored inputs and generators if implementation movement requires it. | `boundary_guard_test.go`, phase maps/ledgers, generated outputs through Make only. | Generated drift, false boundary positives, phase evidence accounting. | Preserve boundary guard intent; update phase rows only with owner evidence. | `make generated-artifact-policy-check`; `make generate-drift`; `make phase-ledger-drift`; `make phase-schedule-drift` | Revert authored input changes; regenerate rather than hand-edit generated outputs. | Guardrails prevent new leaks without blocking intended facades. |
| `S-08` | `S-07` | Run validation and update final handoff. | Tracker, changed files from later implementation. | Broad regressions. | Preserve all existing reporting tests and selected broader gates. | `make agent-finalize`; `make test-fast`; `make check` when risk warrants. | Revert last implementation slice only if validation shows related regression. | Handoff lists exact commands, artifacts, failures, and remaining blockers. |

Slices requiring later authorization:

- `S-02` through `S-08` require a later authorized implementation task.
- Any behavior change, public route/schema change, generated contract update, migration, or harness map update requires separate authorization.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make phase-slice PHASE=phase11` | Phase 11 backend unit, frontend unit, backend integration, and browser functional evidence selected by manifest. | yes | Use after implementation slices that affect reporting behavior. |
| integration | `make service-backed-slice PHASE=phase11` | Phase 11 service-backed backend/browser evidence. | yes | Covers reporting route and DB behaviors more directly than tracker-only work. |
| e2e/browser | `make browser-e2e` or phase-selected browser evidence via `make phase-slice PHASE=phase11` | Browser-level extension/profile and related phase evidence. | no for backend-only slices; yes if frontend/report-builder UI changes. | No authored frontend reporting caller was found in this pass. |
| generated drift | `make generate-drift` | Generated sqlc and contract-derived outputs. | yes if authored SQL/contracts/spec-derived inputs change | Do not hand-edit `internal/gen/**`, generated TS roots, or `tools/task_surface.generated.mk`. |
| generated policy/static shape | `make generated-artifact-policy-check`; `make json-shape-check` | Generated-file policy and JSON/contract/manifest shape. | yes for contract/policy/docs that reference generated boundaries | Useful for tracker/docs and contract-adjacent changes. |
| migration drift | `make migration-drift` | Migration/schema drift against scratch database. | yes if migrations or schema-facing SQL changes | Uses Postgres service. |
| import-boundary/static | `make frontend-import-boundary-check`; `make lint`; TODO: command discovery required for reporting-only Go boundary target | Frontend boundary plus general lint. | yes if imports or frontend contracts change | Searched `make help`, `make help-all`, `make explain-target`, `make task-guide ROLE=feature-dev PHASE=phase11`, and `make explain-phase PHASE=phase11`; no reporting-only public Make target found. |
| full check | `make test-fast`; `make check` | Local aggregate and broader developer gate. | `make test-fast` yes before handoff; `make check` when risk or review requires | `make check` includes Postgres, object store, and browser stack. |
| tracker-only docs | `make lint-markdown` | Authored Markdown. | no for implementation; yes for this tracker write | Run for `S-00`; passed on `2026-07-06` with no output. |
| harness/accounting finalization | `make agent-finalize` | Harness-maintenance artifacts before broader verification. | yes before broader end-of-run verification | If retaining prior run evidence, pass `RESULTS_DIR`; otherwise report it was unset. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| `T-001` | Define target module and normalized label | `WF-00` | DONE | none | Target path `internal/modules/reporting`; label `internal-modules-reporting`. | Scope and label are explicit. |
| `T-002` | Read planning framework and owner posture | `WF-00` | DONE | `T-001` | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; owner docs listed in Section 1. | Authority and framework mismatch are recorded. |
| `T-003` | Inventory every target file | `WF-01` | DONE | `T-002` | Section 2 inventory table. | Every file in `internal/modules/reporting` is listed. |
| `T-004` | Map snapshot/release public contracts | `WF-02` | DONE | `T-003` | Section 4 freeze map; Core 01 §17.3. | Route/resource/error/job contracts have owners and evidence. |
| `T-005` | Map reporting/composition/projection boundaries | `WF-02` | DONE | `T-003` | Sections 3, 4, and 5. | Consumer versus owner boundaries are recorded. |
| `T-006` | Classify coupling risks | `WF-04` | DONE | `T-005` | Section 5 findings. | Each finding has classification and planning action. |
| `T-007` | Identify characterization posture | `WF-03` | DONE | `T-003` | Section 4 existing tests and required characterization rows. | Existing tests and TODO gaps are listed. |
| `T-008` | Sequence behavior-preserving future slices | `WF-06` | DONE | `T-006`, `T-007` | Section 7 slice plan. | Each slice has dependency, validation, rollback, and completion criterion. |
| `T-009` | Discover Make-owned validation commands | `WF-07` | DONE | `T-008` | Section 8 validation plan. | Commands are listed or TODO with search performed. |
| `T-010` | Create tracker file | `WF-08` | DONE | `T-001`..`T-009` | This file. | Tracker exists and only this file is changed. |
| `T-011` | Validate tracker-only Markdown | `WF-08` | DONE | `T-010` | `make lint-markdown` passed on `2026-07-06T20:15:21-04:00`. | Command passed with no output. |
| `T-012` | Later implementation execution | `WF-05`..`WF-08` | DEFERRED | Later authorization | This tracker. | A later authorized task begins `S-01` or later. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Planning-only tracker created from live repository inspection; target exists. | Inspected owner docs, reporting package, adjacent composition/runtime/contracts/migrations. Touched only `docs/handoffs/internal-modules-reporting-module-refactor-tracker.md`. | `sed`, `find`, `rg`, `wc`, `git status --short`, `git rev-parse`, `date`, `make help`, `make help-all`, `make explain-target`, `make task-guide ROLE=feature-dev PHASE=phase11`, `make explain-phase PHASE=phase11`, `make lint-markdown`. | No owner contradiction found; framework filename mismatch recorded; `make lint-markdown` passed. | None for tracker creation; implementation deferred. | Hand off; later implementation starts at `S-01` or a narrower authorized slice. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Reporting package diagnosed as mixed-responsibility: facade plus transport, persistence, export collection, graph/composition consumer validation, release lifecycle, redaction/render. | Inspected `api.go`, `routes.go`, `store.go`, `redaction.go`, tests. Touched tracker only. | `rg` symbol/import searches; `sed` exact source reads. | Boundary findings and slices recorded. | Raw SQL and owner-facade decisions require later implementation authorization. | Start with characterization gaps, then source/export seam. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | No authored frontend production caller found for snapshot/release routes; frontend evidence is extension registry visibility and unrelated decision enum. | Inspected `apps/web/e2e/phase2.support.spec.ts`, `apps/web/src/workbook/models/workbookSurfaceRegistry.test.ts`, `packages/protocol-ts/src/index.test.ts`. Touched tracker only. | `rg` over `apps`, `packages`, `contracts`; `sed` exact reads. | Frontend boundary marked defer/not directly applicable. | Future report builder UI may create new frontend boundaries. | Re-check if later task touches report builder or generated TS contracts. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Generated OpenAPI and SQLC are downstream evidence; authored inputs are SQL/spec/contracts only. | Inspected `contracts/openapi/cartulary.openapi.yaml`, `contracts/extensions/index.json`, `contracts/reporting/fixtures/corpus.v1.json`, `db/queries/reporting_phase11.sql`, migrations, generated policy. Touched tracker only. | `rg`, `sed`, `make explain-target` for drift targets. | Generated-file hand-edit risk recorded. | No generated files may be hand-edited. | Use `make generate`/drift checks in later implementation if inputs change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Phase 11 evidence covers reporting unit/integration/OpenAPI/traceability tests; reporting-only public Make target not found. | Inspected reporting tests, `tools/phase11_test_map.json`, `docs/testing/phase11_coverage_ledger.md`. Touched tracker only. | `make task-guide ROLE=feature-dev PHASE=phase11`, `make explain-phase PHASE=phase11`, `make explain-target` for key targets, `make lint-markdown`. | Validation plan recorded with TODO for reporting-only command discovery; tracker Markdown lint passed. | Missing DB-backed graph/composition characterization before facade move. | Later implementation should run phase slices and service-backed checks as risk requires. |
| `2026-07-06T20:38:27-04:00` | Codex remediation implementation | Graph/composition owner-facade slice implemented and validated. | Touched `internal/modules/graphprojection/reporting_refs.go`, `internal/modules/graphprojection/store_test.go`, `internal/modules/reportcomposition/release_tuple_facade.go`, `internal/modules/reportcomposition/release_tuple_facade_test.go`, `internal/modules/reporting/store.go`, `internal/modules/reporting/boundary_guard_test.go`, this tracker, and `docs/reporting-subsystem-nlspec.md`. | `make explain-target TARGET=backend-unit DETAIL=summary`; `make explain-target TARGET=phase-slice DETAIL=summary`; `make explain-target TARGET=test-fast DETAIL=summary`; `make backend-unit`; `make phase-slice PHASE=phase11 ROWS=U-11-REPORTING-01,U-11-REPORTING-02,U-11-REPORTING-03,U-11-REPORTING-04,U-11-REPORTING-05,U-11-REPORTING-06`; `make service-backed-slice PHASE=phase11 ROWS=I-11-REPORTING-01,I-11-REPORTING-02,I-11-REPORTING-03,I-11-REPORTING-04`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-markdown`; `make agent-finalize`; `make test-fast`; `make check`. | Passed: `backend-unit` at `.cartulary/test-results/20260707T003827Z-p3698572`; reporting unit slice at `.cartulary/test-results/20260707T004019Z-p3701802`; reporting service-backed slice at `.cartulary/test-results/20260707T004054Z-p3716085`; generated-artifact policy at `.cartulary/test-results/20260707T004131Z-p3728540`; JSON shape at `.cartulary/test-results/20260707T004131Z-p3728539`; `agent-finalize` at `.cartulary/test-results/20260707T004143Z-p3729461`; `test-fast` at `.cartulary/test-results/20260707T004148Z-p3729658`. `lint-markdown` exited 0 without a run root. `make check` failed at `.cartulary/test-results/20260707T004443Z-p3775659` on unrelated staticcheck `cmd/operator/operator_phase10_test.go:919:6: func operatorProcessEnv is unused (U1000)`. | Export-provider SQL, service/store split, full renderer, deployment-admin-specific reporting route coverage, and unrelated operator staticcheck cleanup remain open. | Continue with export-provider refactor characterization before moving source/projection SQL. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Core 04 release gate and live-workbook/export boundary recorded; reporting route authorization uses incident membership/roles. | Inspected Core 04 search hits and reporting route handlers/tests. Touched tracker only. | `rg`, `sed`. | Authorization and redaction freeze map recorded. | Deployment-admin-specific reporting route behavior was not separately characterized in inspected tests. | Preserve current route-scoped errors and role requirements before movement. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `2026-07-06T20:10:22-04:00` | Codex tracker creation | Open risks are direct source/projection SQL, graph/composition raw DB validation, simplified renderer, and generated/harness drift. | Touched tracker only. | One earlier `make help-all | rg ...` failed because of malformed regex and broken pipe; later grep-based discovery succeeded. | Risks and blocker IDs recorded. | Implementation deferred until later authorized task. | Begin later task at `S-01` characterization gap analysis. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| `RB-001` | TODO: identify or create a reporting-only public Make validation target, if one exists or is warranted. | Narrow validation reduces cost and risk for reporting-only refactors. | Make target inventory or harness owner decision. | Closed for this remediation slice; use documented row-filtered `make phase-slice PHASE=phase11 ROWS=<U-11-REPORTING rows>` and `make service-backed-slice PHASE=phase11 ROWS=<I-11-REPORTING rows>` rather than adding a bespoke target unless `ROWS` support is removed. |
| `RB-002` | TODO: characterize graph projection DB validation outcomes before replacing raw table reads. | Facade movement could alter `graph_projection_not_bound`, `graph_projection_not_completed`, `graph_projection_stale`, `graph_projection_digest_mismatch`, or ambiguity behavior. | Integration tests or projection owner facade evidence. | Closed for raw table replacement; `graphprojection.ValidateReportingProjectionRefsTx` owns DB validation and `TestValidateReportingProjectionRefsTxReasonMatrix` covers completed, replaced, missing run, not-completed state, stale snapshot, and digest mismatches. |
| `RB-003` | TODO: characterize composition tuple DB mismatch and binding cases before replacing raw table reads/writes. | Facade movement could alter not-found, version missing, template mismatch, digest mismatch, version parsing, or release-bound behavior. | Integration tests or reportcomposition owner facade evidence. | Closed for raw table replacement; `reportcomposition.ResolveReleaseTupleTx` and `BindReleaseTupleTx` own DB validation/binding and `release_tuple_facade_test.go` covers valid binding, missing/cross-incident resource hiding, missing version, template mismatch, digest mismatch, invalid version, and release-bound visibility. |
| `RB-004` | TODO: decide whether source/export collection should live as reporting-owned ports or owner-provided export providers. | This determines how direct SQL over source and projection tables is removed without weakening owner boundaries. | Architecture decision in later implementation planning. | Open; current tracker records finding only. |
| `RB-005` | TODO: resolve simplified renderer versus full Reporting NLSpec stack before deeper render refactor. | A refactor that changes output bytes, render manifests, or toolchain semantics may become behavior work rather than structural movement. | Reporting owner task and characterization/fixture decision. | Open; marked `defer`. |
| `RB-006` | TODO: deployment-admin reporting route behavior was not independently audited in this pass. | Core 04 route-inventory requirements may matter if route authorization is rearranged. | Existing tests or new route-inventory characterization. | Open; not blocking tracker creation, blocks high-confidence auth refactor. |

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/modules/reporting` is inventoried or explicitly out of scope. | DONE | Section 2 lists `.gitkeep`, four production Go files, and five test files. |
| Every discovered public contract risk has an owner and test posture. | DONE | Section 4 maps routes, jobs, export, release, auth, generated, harness, and frontend-adjacent contracts. |
| Every proposed workflow has dependencies and exit criteria. | DONE | Section 6 lists `WF-00` through `WF-08`. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | DONE | Section 7 marks later implementation slices as requiring authorization and preserves behavior by default. |
| Validation commands are discovered or marked `TODO` with a reason. | DONE | Section 8 lists Make-owned commands and records reporting-only target discovery TODO. |
| Contradictions are marked `BLOCKED: owner contradiction`. | DONE | No owner contradiction found in this pass. |
| Repository/framework mismatches are recorded as planning findings. | DONE | Section 1 records the live Core filename mismatch from framework footnotes. |
| Handoff sections are current enough for another agent to continue without rediscovery. | DONE | Section 10 includes workstream-specific handoff rows. |
| Implementation requires a later authorized task. | DONE | Section 1 and Section 7 state this explicitly. |
| No production refactor was performed in this tracker task. | DONE | Allowed changes are limited to this tracker file. |
