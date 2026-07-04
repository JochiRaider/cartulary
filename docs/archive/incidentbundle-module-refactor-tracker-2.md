# incidentbundles Module Refactoring Tracker and Handoff Document

## 1. Scope and Source Posture

Target directory: `internal/modules/incidentbundles`.

Output path: `docs/handoffs/incidentbundle-module-refactor-tracker-2.md`.

This document is a planning-only and handoff-only tracker. It authorizes no
production refactor, generated output update, migration rewrite, contract edit,
codegen run, route change, behavior change, or implementation patch unless a
later task explicitly authorizes that work.

The 2026-07-02 remediation implementation session is that later authorized
task. It is limited to the remaining durable cleanups around route assembly,
job orchestration, filesystem staging, store admission boundaries, optional
capability semantics, and the production-compiled worker test hook. Q-001
through Q-005 remain closed baseline remediation and must be preserved.

Implementation cadence requirement: update this tracker after each workstream
is completed and before the next workstream starts. Each update must record the
workstream status, substantive files changed, validation run or intentionally
skipped, blockers, and next action.

Allowed change in this session: create this tracker from live repository
inspection and the local modular-refactor planning framework. Later
implementation work must preserve observable behavior by default, including
route shape, request and response envelopes, job resources, authorization
outcomes, bundle layout, storage semantics, workbook interaction behavior,
projection refresh behavior, generated contract surfaces, and harness
accounting.

Non-goals for this tracker:

- Do not move, rename, or edit production code under `internal/modules`.
- Do not hand-edit generated roots such as `internal/gen/**`,
  `packages/protocol-ts/src/generated/**`, or
  `packages/ui-contracts/src/generated/**`.
- Do not hand-edit generated harness outputs or tool-managed dependency files.
- Do not introduce a new module boundary decision for `incidentbundles`.
- Do not treat phase maps or test rows as runtime architecture.

Authority order used:

1. Adopted subsystem NLSpecs for their named subsystem only.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication.
4. `docs/domain.md` and implementation-support guides for terminology, package
   boundaries, harness mechanics, and execution support.
5. Current repository code and tests for current implementation state.
6. Prior plans and framework files as evidence, not authority.

No `BLOCKED: owner contradiction` item was found during this inspection.

Repo/framework mismatch: the planning framework module catalog does not list
`incidentbundles`, but Core 01 explicitly owns the Incident Portability Extension
Profile and the `/api/v1/incident-bundles/*` route family. Treat that mismatch
as a planning finding. It is not proof that `incidentbundles` is either invalid
or clean as a permanent module boundary.

Architectural finding: `internal/modules/incidentbundles` currently owns
legitimate Incident Portability facade behavior, but the directory must not be
assumed to be a valid permanent module boundary merely because it exists. The
current package is a mixture: route and job facade, archive and verification
core, export/import coordinator, descriptor persistence, filesystem staging,
imported actor and attribution sidecars, and orchestration over owner ports for
records, timeline, parties, entities, indicators, artifacts, task/decision
state, evidence, assessments, links, revisions, saved views, projections, and
incidents.

## 2. Current-State Repository Inventory

| path | current responsibility | exported/public symbols or package surface | inbound callers | outbound dependencies | tests touching it | generated artifacts or contracts touched | suspected target owner module | risk level | notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/incidentbundles/api.go` | Export/import request decoding, media constants, upload allowlist, canonical JSON/hash helpers, route-owned API error helpers. | `ProfileID`, mode/result/media constants, `IncidentBundleFileContentTypes`, `ExportRequest`, `ImportMetadataRequest`, `DecodeExportRequest`, `DecodeImportMetadata`. | `routes.go`, `store.go`, `worker_service.go`, `api_test.go`, integration tests. | `internal/platform/httpapi`, `net/http`, JSON, UUID parsing, hashing. | `api_test.go` covers export canonicalization, import metadata normalization, OpenAPI/error registry expectations. | OpenAPI/error/extension registries are asserted by tests; no generated file is touched by this file. | Keep in incident portability facade. | medium | Cohesive route contract code. Preserve closed reason-code behavior. |
| `internal/modules/incidentbundles/bundle.go` | Deterministic bundle manifest/archive construction; required structured file registry; ZIP/TAR/TAR.GZ verification; checksum, member path, optional section, capability, byte-limit, member-count, and compression-ratio validation. | `BundleFormat`, `BundleVersion`, `ManifestInput`, `BundleArchive`, `BundleManifest`, `ManifestFile`, `VerificationInput`, `VerifiedBundle`, `VerificationError`, `BuildBundleArchive`, `VerifyBundle`. | `source.go`, `worker_service.go`, `api_test.go`, route integration tests. | `internal/platform/config`, archive libraries, JSON/hash/path helpers. | `api_test.go` covers deterministic manifest/checksum, verifier failures, compression-ratio boundary; integration tests exercise import/export. | Bundle logical layout from Core 01; OpenAPI/error registry expectations through tests. | Keep archive core in incident portability facade. | high | Archive layout and deterministic bytes are observable portability behavior. |
| `internal/modules/incidentbundles/source.go` | Cross-owner export/import orchestration; incident source bundle build/import; actor export/import; attribution buffering; evidence blob staging through evidence port; revision sequence repair; projection rebuild; import finalization. | `BundleBuilder`, `Importer`, `BuiltIncidentBundle`, `ImportParams`, `Build`, `Import`. | `worker_service.go`; integration tests through route jobs. | Owner portability ports in records, timeline, parties, entities, indicators, artifacts, tasks/decisions, evidence, assessments, links, revisions, savedviews, incidents; shared `incidentportability`; `projections/adapters`; `objectstore`; `pgx`. | `routes_integration_test.go` covers import/export round trip, blob preservation, history/rollback, projection rebuild, final-publication auth, failure cleanup. | Bundle structured file registry; projection and owner source state; phase11 evidence rows. | Split or defer seams behind clearer owners while keeping incident portability as coordinator. | critical | Most important boundary risk: broad orchestration is valid, but owner-specific state rules must remain behind owner ports. |
| `internal/modules/incidentbundles/store.go` | Incident-bundle job payload persistence, route idempotency replay/conflict, export descriptor persistence, manifest-file rows, job failure metadata. | `ErrNotFound`, `Store`, `NewStore`, `JobAcceptedResult`, `JobPayload`, `ExportAcceptedParams`, `ImportAcceptedParams`, `ExportCompleteParams`, `DescriptorRecord`, store methods. | `routes.go`, `worker_service.go`, integration tests. | `pgxpool`, `pgx`, `internal/platform/authn`, `internal/platform/jobs`, SQL tables `incident_bundle_*`, `route_idempotency`, `jobs`. | Integration tests cover idempotency, descriptors, job payload failure state, descriptor reads. | Schema ownership manifest assigns incident-bundle tables to `incidentbundles`; OpenAPI descriptor behavior tested. | Keep persistence facade, review job/auth substrate boundary. | high | Cohesive for portability-owned tables, but couples directly to auth/job platform substrate. |
| `internal/modules/incidentbundles/routes.go` | Extension-gated HTTP registration, service construction, deployment-admin auth, export membership admission, import upload admission, job dispatch/recovery, session sliding. | `Service`, `RegisterRoutes`; handlers are unexported. | `internal/app/runtime.go`; browser import panel indirectly through HTTP; integration tests. | `incidents.Access`, incidents import finalizer, workbook startup bootstrap, `authn`, `httpapi`, `httpauth`, bundle file store, worker, store. | `routes_integration_test.go` covers route auth, import/export idempotency, descriptors, upload envelope failures, final-publication auth. | OpenAPI route family; extension discovery contracts; phase11 rows. | Transport facade with service split candidates. | high | Route shape and auth outcomes are frozen. Avoid moving domain decisions into transport. |
| `internal/modules/incidentbundles/worker_service.go` | Async export/import job recovery and execution; job status transitions; bundle build/import; persisted file use; terminal success/failure summaries. | Unexported `incidentBundleWorker` and methods. | `routes.go` service construction and dispatch; integration tests through job polling. | `Store`, `BundleBuilder`, `Importer`, `bundleFileStore`, `jobs.Manager`, `jobs.Runner`, `httpapi.DependencySet`, projections adapter, incidents finalizer. | Integration tests cover terminal summaries, cancellation/auth, import failure reasons, staging cleanup. | Common job resource semantics; result summary codes and resource refs. | Service/job orchestration facade, with platform job boundary review. | high | Worker owns observable terminal codes; keep result summaries exact. |
| `internal/modules/incidentbundles/bundle_files.go` | Temporary import staging and export-bundle persistence under configured roots or OS temp fallback; cleanup helper. | Unexported `bundleFileStore` and methods. | `routes.go`, `worker_service.go`, integration staging assertions. | `os`, `filepath`, configured temp/export roots, UUID. | Integration tests assert staging cleanup and persisted export path behavior. | Runtime-root behavior; no generated artifacts. | Candidate platform file-root adapter or private incident portability adapter. | medium | Direct filesystem writes are observable through configured root semantics and cleanup. |
| `internal/modules/incidentbundles/attribution_resolver.go` | Registers imported attribution resolver with revisions and reads incident-bundle attribution sidecars by incident, source table, source column, and source row IDs. | `init()` side effect registers `importedAttributionResolver`; resolver type is unexported. | `internal/modules/revisions` resolver registry at package init; revision history reads indirectly. | `revisions`, `pgx`, SQL table `incident_bundle_imported_attributions`. | `revisions/imported_attribution_boundary_test.go`; integration history assertions in `routes_integration_test.go`. | Revision/history behavior; no generated artifacts. | Split boundary between incident portability sidecars and revisions read model. | medium | Init-time registration is convenient but hidden. Harden only with characterization. |
| `internal/modules/incidentbundles/worker_hooks.go` | Test-only worker-start hook for pausing async worker execution in integration tests. | `SetIncidentBundleWorkerStartHookForTesting`. | `worker_service.go`; `routes_integration_test.go`. | `sync`. | Integration tests use hook for auth/final-publication race characterization. | None. | Replace with test-runtime guarded dependency override during remediation. | low | Exported test hook is production-compiled; remove it as part of WS-04. |
| `internal/modules/incidentbundles/api_test.go` | Unit characterization for request normalization, bundle determinism, verifier failures, OpenAPI route/schema presence, closed error registries. | Test functions only. | Make/harness phase11 backend-unit rows. | Contract files, config limits, archive helpers. | Self. | `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `contracts/extensions/index.json`, protocol expectations. | Keep as contract characterization. | critical | Do not weaken before route/archive movement. |
| `internal/modules/incidentbundles/routes_integration_test.go` | Integration characterization for export/import routes, idempotency, jobs, auth, upload envelope, failure durability, source-state preservation, projection rebuild, history/rollback, blob storage, final-publication auth. | Test functions and helpers only. | Make/harness phase11 backend-integration rows. | `phase2test`, `httptestx`, object store, DB fixtures, route helpers, incidentbundle constants. | Self. | Phase11 map, coverage ledger, scheduler manifest, duration baselines. | Keep and extend for any behavior movement. | critical | Broad fixture is the primary behavior freeze for future slices. |

Additional inbound and adjacent surfaces discovered:

- Production registration: `internal/app/runtime.go` imports `incidentbundles` and
  appends `incidentbundles.RegisterRoutes()`.
- Frontend caller: `apps/web/src/app/IncidentImportPanel.tsx` posts to
  `/api/v1/incident-bundles/import`.
- Frontend visibility: `apps/web/src/app/App.tsx` exposes the import panel when
  `incident_portability` is claimed.
- Contract surfaces: `contracts/openapi/cartulary.openapi.yaml`,
  `contracts/errors/index.json`, `contracts/extensions/index.json`,
  `contracts/otel/error_class_registry.json`, and
  `packages/protocol-ts/src/index.test.ts`.
- Schema inputs: `db/migrations/00030_phase11_incident_bundles.sql`,
  `db/migrations/00031_phase11_incident_bundle_remediation.sql`,
  `db/migrations/00032_phase11_incident_bundle_job_scope.sql`, and
  `tools/schema_object_ownership_manifest.json`.
- Harness accounting: `tools/phase11_test_map.json`,
  `docs/testing/phase11_coverage_ledger.md`,
  `tools/scheduler_manifest.json`,
  `tools/execution_topology_manifest.json`, and
  `tools/go_test_duration_baselines.json`.

## 3. Module Boundary Diagnosis

Diagnosis: `incidentbundles` is a mixture.

It is a legitimate Incident Portability facade for:

- the extension route family under `/api/v1/incident-bundles/*`;
- route-level request decoding and family error mapping;
- background job admission and terminal result summaries;
- export descriptor and portability-owned sidecar persistence;
- deterministic bundle archive construction and verification.

It is not primarily workbook orchestration. It affects workbook behavior
indirectly by importing authoritative source state, rebuilding projections,
preserving saved views, preserving history and rollback substrate, moving object
blob bytes through the evidence owner, and finalizing imported incident startup
visibility.

It is also not a proof of a permanent clean module boundary. Current code still
coordinates many owner modules, holds sidecar attribution state that revisions
consume, directly carries object-store and database transaction dependencies,
and uses shared portability helpers that perform generic SQL import.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Incident-bundle route family and extension gating | `routes.go` | `incidentbundles` facade with `httpapi` extension registry | keep | `RegisterRoutes`, Core 01 section 17.5 | Preserve route paths and unclaimed-extension behavior. |
| Export/import request decoding and family errors | `api.go` | `incidentbundles` facade | keep | `DecodeExportRequest`, `DecodeImportMetadata`, error registry tests | Closed reason sets are public contract. |
| Bundle archive layout and verifier | `bundle.go` | `incidentbundles` archive core | keep | `BuildBundleArchive`, `VerifyBundle`, Core 01 section 12.3 | Deterministic bytes and required file registry are observable. |
| Cross-owner source export/import ordering | `source.go` | Incident portability coordinator plus owner ports | split/defer | `incidentBundleExportPorts`, `incidentBundleImportPorts` | Keep coordinator; avoid ownerless source-family logic. |
| Shared NDJSON SQL import helper | `internal/modules/incidentportability/portability.go` used by owner ports | Shared incident portability support or stronger owner-specific ports | defer | `ImportNDJSON` uses `jsonb_populate_record` | Review before adding new source families. |
| Evidence blob byte movement | `source.go` via `evidence.IncidentBundleBlobPortability` | Evidence owner plus incident portability coordinator | keep current split; review adapter dependency | Evidence blob port owns lifecycle; `source.go` supplies object store | Preserve target-owned storage keys and cleanup. |
| Import final publication | `source.go` through `incidents.IncidentBundleImportFinalizer` | Incidents plus workbook startup bootstrap | keep owner port | Core 01 REQ-01-609; incidents finalizer tests | Final visibility and membership/audit behavior are owner-owned. |
| Projection rebuild after import | `source.go`, `worker_service.go` | Projections | keep owner adapter | `projectionadapters.NewIncidentImportRebuilder` | Projection state is derived, not source truth. |
| Job payload and descriptor persistence | `store.go` | Incident portability facade plus platform jobs/auth substrate | keep with review | `incident_bundle_*`, `jobs`, `route_idempotency` | Store owns portability tables but not job policy semantics alone. |
| Filesystem staging and persisted bundles | `bundle_files.go`, `worker_service.go` | Private adapter or platform runtime-root adapter | split/defer | `stageBundle`, `persistBundle`, configured roots | Preserve temp/export root semantics and cleanup. |
| Imported attribution sidecars | `source.go`, `attribution_resolver.go` | Incident portability sidecar owner plus revisions resolver port | split/defer | `incident_bundle_imported_attributions`, revisions resolver | Hidden init registration is a boundary smell. |
| Saved-view export/import | Owner port in `savedviews`, orchestrated by `source.go` | `savedviews` owner | keep current owner port; characterize if changed | `data/saved_views.ndjson` | Saved view scope must not alter authorization. |
| Harness phase/test accounting | Tests, phase maps, ledgers, scheduler manifests | Harness support | defer | phase11 rows and Make targets | Evidence accounting only, not runtime architecture. |

## 4. Public Contract and Behavior Freeze Map

| Contract surface | Existing behavior to freeze | Existing tests or evidence | Required characterization tests |
| --- | --- | --- | --- |
| HTTP `POST /api/v1/incident-bundles/export` | JSON body with required `incident_id` and `client_txn_id`; optional `reference_pack_mode`, `optional_sections`, `required_capabilities`; rejects user-supplied `history_mode` and `blob_mode`; returns `202` common job resource. | `TestPhase11_U_11_INCIDENT_BUNDLES_01_DecodeExportRequestCanonicalizesAndRejectsModes`; `TestPhase11_I_11_INCIDENT_BUNDLES_01_ExportJobIdempotencyAndDescriptor`; Core 01 section 17.5. | Existing coverage is sufficient for docs-only planning; add targeted tests before changing admission or idempotency. |
| HTTP `GET /api/v1/incident-bundles/{bundle_id}` | Singleton descriptor read; rejects pagination; requires deployment admin plus incident membership; non-visible descriptors return `incident_bundle_not_found`; descriptor serializes canonical fixed modes and arrays. | `TestPhase11_I_11_INCIDENT_BUNDLES_06_DescriptorPaginationAndCanonicalManifest`; OpenAPI/error tests. | Add tests only if descriptor store/resource mapping moves. |
| HTTP `POST /api/v1/incident-bundles/import` | Shared upload envelope with required metadata `client_txn_id`; file content-type allowlist; idempotency includes exact uploaded file SHA-256; no durable import resource; terminal success emits imported incident ref. | `TestPhase11_I_11_INCIDENT_BUNDLES_02_ImportEnvelopeIdempotencyAndImportedIncidentOpen`; `TestPhase11_I_11_INCIDENT_BUNDLES_07_ImportEnvelopeFailuresCreateNoDurableState`; Core 01 REQ-01-485. | Existing coverage is sufficient before route-preserving refactor; add tests before changing upload parser use. |
| WebSocket events | No incidentbundle-specific WebSocket path or event was found in target package. Common job progress, if any, remains common job/collaboration-owned. | `rg` over target package and route registration found no WS surface. | No incidentbundle-specific characterization is required unless common job progress publication changes. |
| Common job result summaries | Export terminal success uses `incident_bundle_exported` and one `incident_bundle` ref; import uses `incident_bundle_imported` and one imported `incident` ref. | `routes_integration_test.go` assertions; Core 01 section 17.5; constants in `api.go`. | Preserve exact code/ref count/route in worker split. |
| Workbook row/query/mutation behavior | Import preserves authoritative source state and rebuilds projections so ordinary workbook startup/query/history/rollback work for imported incident. | Main import integration test; superseded timeline integration test; Core 01 REQ-01-448. | Add focused characterization before any owner-port reorder or projection rebuild change. |
| Saved-view or view-schema behavior | `data/saved_views.ndjson` is required bundle state via `savedviews` owner port; generated view contracts are not directly edited by target package. | Required structured file registry; savedviews owner port; broad import integration source-state checks; `internal/modules/savedviews/incident_bundle_portability_test.go`. | Saved-view characterization is now owner-owned and fail-closed; preserve it before any saved-view import/export movement. |
| Projection refresh behavior | Import rebuilds projections inside the import transaction before final publication. | `projectionadapters.NewIncidentImportRebuilder`; integration tests check projection count and imported workbook query behavior. | Preserve transaction timing; characterize failure path before adapter changes. |
| Authorization checks | Export route and export jobs require deployment admin plus current incident membership; import is deployment-scoped until target incident exists; final publication rechecks submitter active deployment admin. | Core 04 REQ-04-023; `TestPhase11_I_11_INCIDENT_BUNDLES_03_ExportJobAuthorizationReDerivesIncidentMembership`; `TestPhase11_I_11_INCIDENT_BUNDLES_08_ImportFinalPublicationRechecksSubmitterAvailability`. | Existing tests are high-value. Add tests before changing job policy construction. |
| Revision/change-set behavior | Bundle preserves change sets, mutations, record revisions, history substrate, imported attribution sidecars, and rollback behavior. | Main import integration test; `revisions/imported_attribution_boundary_test.go`; Core 01 REQ-01-564. | Preserve imported attribution resolver semantics before registration refactor. |
| Generated protocol/view contracts | OpenAPI exposes incident-bundle routes and schema; error/extension registries include route family and codes; protocol tests include `incident_portability`. | `api_test.go`; contract files; protocol test. | Run drift checks if any owner input changes. Do not hand-edit generated roots. |
| Harness/test accounting | Phase11 backend-unit/integration rows account for incident-bundle evidence; scheduler and duration artifacts include incidentbundle rows. | `tools/phase11_test_map.json`, `docs/testing/phase11_coverage_ledger.md`, scheduler/duration artifacts. | Treat as evidence accounting only. Update through owner inputs and Make generators if tests move or names change. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Broad source-state orchestration can become ownerless if new bundle files are added directly in `source.go`. | `source.go` coordinates owner ports and required structured files in `bundle.go`. | New source family could bypass owner invariants or tests. | `must_fix` | Incident portability coordinator plus owning source modules. | Any new source family must add owner port, required file registry update, and characterization together. |
| Shared `incidentportability.ImportNDJSON` uses `jsonb_populate_record`; raw table names have been removed from owner ports. | `internal/modules/incidentportability/import_targets.go`; `internal/modules/incidentportability/portability.go`; owner `incident_bundle_portability.go` files. | Generic SQL import could bypass owner validation if future source families construct unregistered descriptors. | `must_fix` | Shared incident portability support plus owning source modules. | DONE for remediation pass: import now requires registered `ImportTargetDescriptor`; future source families must update Core 01 and the implementation registry together. |
| `store.go` couples incident-bundle persistence to auth route-idempotency and common jobs. | Imports `authn` and `jobs`; SQL touches `route_idempotency`, `jobs`, and `incident_bundle_job_payloads`. | Persistence refactor could drift idempotency or job auth semantics. | `should_fix` | Incident portability facade plus platform auth/jobs. | Keep behavior tests green before splitting store adapters. |
| `routes.go` remains transport-adjacent but constructs incident finalizer, access checker, auth store, file store, and worker. | `newService` wires platform, incidents, workbook bootstrap, file roots, worker. | Route facade can accumulate application assembly responsibilities. | `should_fix` | `internal/app` assembly plus incident portability service. | Plan a small service-construction seam only after contracts are characterized. |
| `worker_service.go` owns job execution and terminal summaries directly. | Calls `jobs.Manager` transitions and builds result summaries. | Moving job logic can change observable terminal codes or refs. | `should_fix` | Common jobs plus incident portability worker. | If split, freeze result summaries and cancellation behavior first. |
| Direct filesystem staging is private but platform-root semantics matter. | `bundle_files.go` writes under configured temporary/export roots or OS temp fallback. | Incorrect move could leak files or break cleanup. | `should_fix` | Platform runtime-root adapter or private incident portability file adapter. | Characterize path and cleanup behavior before changing. |
| Init-time attribution resolver registration hid revision sidecar dependency. | `attribution_resolver.go`; `internal/modules/revisions/attribution_resolver_registry.go`; `internal/app/runtime.go`. | Import order or package registration bugs could affect history attribution. | `should_fix` | Revisions resolver port plus incident portability sidecar owner. | DONE for remediation pass: revisions owns explicit registry; app runtime registers incidentbundle resolver before route assembly. |
| No duplicate workbook row/view-schema logic found in target package. | Target package does not define view-schema fields or workbook row rendering; it imports projections adapter and savedviews port. | Misdiagnosing it as workbook owner would cause wrong module move. | `intentional/no_action` | Workbook/projections/savedviews owners. | Keep tracker language clear: workbook effects are indirect. |
| Optional embedded snapshots/reference packs are verifier-level only in current implementation evidence. | `bundleOptionalSectionsAllowed`; Core 01 allows optional sections. | Implementing optional embed behavior could expand scope beyond refactor. | `defer` | Reporting/snapshots and reference_data owners. | Defer unless owner docs and tests authorize implementation. |
| Archive verifier and route contract are legitimate incident portability facade behavior. | Core 01 section 12.3 and 17.5; `bundle.go`, `api.go`, `routes.go`. | Moving them for purity could increase public contract risk. | `intentional/no_action` | `incidentbundles` facade. | Preserve unless a later owner decision creates a new archive package boundary. |
| Test-only hook is exported from production package. | `SetIncidentBundleWorkerStartHookForTesting` in `worker_hooks.go`. | Test seam could leak into production assumptions. | `defer` | Test harness or explicit worker controller. | Leave until worker split plan defines a replacement. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Create planning-only tracker and record authority/source posture. | `docs/handoffs/incidentbundle-module-refactor-tracker-2.md` | `make lint-markdown`, `make generated-artifact-policy-check`, `make json-shape-check` | Tracker path exists and states docs-only scope. |
| WF-01 | Target package inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every file in `internal/modules/incidentbundles`. | All files in target package. | Review file inventory and tests. | Every target file has owner/risk/test posture. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map routes, jobs, bundle layout, auth, projection, history, saved-view, generated contract surfaces to owners. | Core 00/Core 01/Core 04/domain/harness docs; target code; contracts. | `api_test.go` contract tests; static inspection. | Contract freeze map is complete or TODO-marked. |
| WF-03 | Characterization test gap analysis | chain | WF-01, WF-02 | WF-05, WF-06 | Identify tests that must exist before any behavior-preserving movement. | `api_test.go`, `routes_integration_test.go`, owner tests, phase maps. | `make backend-unit`, `make backend-integration` when implementation starts. | Missing tests are explicit TODOs, not guesses. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05 | Classify platform imports, owner ports, shared SQL helpers, object-store/file-root dependencies, resolver registration. | `source.go`, `store.go`, `routes.go`, `worker_service.go`, `bundle_files.go`, `attribution_resolver.go`, owner portability files. | `make backend-module-boundary-check`; `make lint`. | Findings have classifications and proposed owners. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Plan behavior-preserving seams without changing route or bundle contracts. | `source.go`, `store.go`, `worker_service.go`, `routes.go`, owner ports. | Unit/integration characterization before patching. | Proposed seams list keep/move/split/defer. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Define smallest safe slices and rollback criteria. | Tracker and future implementation files. | Phase11 slice after implementation. | Slices have dependency, risk, validation, completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Keep phase/test accounting as evidence only and update through owners/generators if needed. | `tools/phase11_test_map.json`, ledgers, scheduler, duration baselines. | `make phase-ledger-drift`, `make phase-schedule-drift`, `make json-shape-check` if accounting changes. | Harness impact is no-op or generator-backed. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run appropriate Make-owned validation and record results. | Tracker, future changed files. | Docs-only now; broader commands after code changes. | Handoff log current enough for next agent. |

## 7. Proposed Refactor Slice Plan

| Slice | Dependency | Exact intended change | Files or packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | none | Create this tracker only. | `docs/handoffs/incidentbundle-module-refactor-tracker-2.md` | None if docs-only. | Existing tests unchanged. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | Delete the new tracker file. | Tracker has all required sections and output path. |
| S-01 | S-00 | Add or refresh characterization before any source-family or owner-port changes. | `api_test.go`, `routes_integration_test.go`, relevant owner tests. | Could overfit implementation instead of public behavior. | Preserve phase11 U/I rows; add saved-view/source-family/projection failure tests only as needed. | `make backend-unit`; `make backend-integration` | Revert added tests if they assert non-contract internals. | Missing behavior-freeze gaps are closed or explicitly TODO. |
| S-02 | S-01 | Review `incidentportability` helper ownership and owner-port contracts before expanding source families. | `internal/modules/incidentportability`, owner `incident_bundle_portability.go` files. | Raw SQL import behavior could drift row defaults or constraints. | Owner-port tests plus broad import round trip. | `make backend-unit`; `make backend-integration`; `make lint` | Keep existing helper if no safer behavior-preserving split is ready. | New source-family plan names owner ports and validation. |
| S-03 | S-01 | Split route/worker/file/job orchestration only after behavior tests are green. | `routes.go`, `worker_service.go`, `bundle_files.go`, `store.go`. | Route shape, job status, idempotency, staging cleanup, terminal summaries. | Preserve all route integration tests; add focused store/worker tests if seams are introduced. | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase11` | Restore prior file boundaries if any public contract shifts. | Transport facade is thinner and all observed behavior is unchanged. |
| S-04 | S-01 | Harden imported attribution resolver registration boundary. | `attribution_resolver.go`, `internal/modules/revisions`, app assembly if explicit wiring is chosen. | History attribution and rollback selectors can drift. | Preserve revisions boundary test and import history assertions. | `make backend-unit`; `make backend-integration` | Revert to init registration if explicit wiring causes startup risk. | Revisions receives resolver through clear boundary with same behavior. |
| S-05 | S-03 or S-04 if tests move | Update harness/accounting only through owner inputs and Make generators if test names or rows change. | Phase maps, generated ledgers/schedules, duration baselines as applicable. | Evidence accounting can drift from runtime behavior. | Existing phase11 row coverage remains implemented. | `make phase-ledger-drift`; `make phase-schedule-drift`; `make json-shape-check`; `make agent-finalize` | Revert generated outputs and owner inputs together. | Harness artifacts agree with live tests and generated policy. |

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Pure backend unit evidence, including incident-bundle API/archive/contract tests. | yes for backend implementation; no for docs-only S-00 | Public Make target discovered by `make help-all` and `make explain-target`. |
| integration | `make backend-integration` | Service-backed backend integration evidence, including incidentbundle route/import/export tests. | yes for source, job, storage, projection, auth, or import/export implementation | Requires Postgres and object store. |
| e2e/browser | `make browser-e2e-webserver-backed` | Browser functional evidence, including phase11 frontend-visible flows. | no for backend-only or docs-only; yes if frontend import panel or visible workflow changes | Phase11 task guide lists it in phase slice. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift` if owner inputs/contracts change | Generated artifact policy, JSON contract/manifest shape, generated drift. | yes for docs-only policy checks; `make generate-drift` only when generated inputs change | Do not hand-edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make lint`; `make frontend-import-boundary-check` only for frontend import changes | Backend module boundary rules, static lint, and frontend import boundary. | yes before handoff for implementation touching backend imports; no for tracker-only except Markdown lint | `backend-module-boundary-check` emits `cartulary.backend_module_boundary_summary.v1` from `tools/backend_module_boundaries.json`. |
| docs | `make lint-markdown` | Authored Markdown structure. | yes for S-00 | Narrow docs-only validation. |
| phase slice | `make phase-slice PHASE=phase11` | Phase11 manifest-row slice: backend unit, backend integration, frontend unit, browser functional. | no for docs-only; yes after code changes affecting incident portability contracts | Discovered by `make task-guide ROLE=feature-dev PHASE=phase11`. |
| service-backed slice | `make service-backed-slice PHASE=phase11` | Phase11 service-backed rows. | no for docs-only; yes after import/export service-backed behavior changes | Requires Postgres, object store, browser stack. |
| full check | `make agent-finalize`; then `make test-fast` or `make check` | End-of-run harness maintenance and broad verification. | no for docs-only; yes depending on implementation breadth | Latest full evidence: `make check` passed with `tests=977 failed=0` at `.cartulary/test-results/20260703T004155Z-p68590`. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| -- | --------- | ---------- | ------ | ---------- | -------------------- | -------------- |
| T-001 | Define target module and docs-only scope. | WF-00 | DONE | none | Section 1 of this tracker. | Target directory, output path, non-goals, and authority order are explicit. |
| T-002 | Inventory every target file. | WF-01 | DONE | T-001 | Section 2 table. | All 11 files in `internal/modules/incidentbundles` are inventoried. |
| T-003 | Record architectural finding about module boundary. | WF-02 | DONE | T-002 | Section 1 and Section 3. | Package is classified as a mixture, not assumed permanent. |
| T-004 | Map public contracts and tests. | WF-02/WF-03 | DONE | T-002 | Section 4. | Route, job, workbook, saved-view, projection, auth, revisions, generated, and harness surfaces have test posture. |
| T-005 | Classify coupling findings. | WF-04 | DONE | T-002 | Section 5. | Findings use `must_fix`, `should_fix`, `defer`, or `intentional/no_action`. |
| T-006 | Define workstreams and dependency chain. | WF-06 | DONE | T-005 | Section 6. | WF-00 through WF-08 have dependencies and checkpoints. |
| T-007 | Define behavior-preserving slices. | WF-06 | DONE | T-006 | Section 7. | Slices include dependencies, validation, rollback, and completion criteria. |
| T-008 | Discover validation commands. | WF-08 | DONE | T-006 | Section 8; `make help`, `make help-all`, `make task-guide`, `make explain-target`. | Commands are Make-owned or marked TODO with reason. |
| T-009 | Validate tracker-only change. | WF-08 | DONE | T-008 | Prior docs-only session command output. | Docs-only validation passed or failures were recorded. |
| T-010 | Future characterization before implementation. | WF-03 | DONE | T-004 | `incidentportability/import_targets_test.go`, `savedviews/incident_bundle_portability_test.go`, `revisions/attribution_resolver_registry_test.go`, phase11 integration assertions. | New invariants are covered by owner unit tests and existing phase11 route integration tests. |
| T-011 | Source-family owner-port review. | WF-05 | DONE | T-010 | `incidentportability.ImportTargetDescriptor`; owner import ports. | Shared helper accepts only registered descriptors; raw table-name import calls are removed. |
| T-012 | Future route/worker/file seam split. | WF-05/WF-06 | DEFERRED | T-010 | Future implementation task. | Public route/job behavior remains unchanged after split. |
| T-013 | Attribution resolver hardening. | WF-05/WF-06 | DONE | T-010 | `revisions.AttributionResolverRegistry`; `incidentbundles.ImportedAttributionResolver`; app runtime wiring. | Revisions attribution behavior remains explicitly wired and tested. |
| T-014 | Future harness accounting updates if tests move. | WF-07 | DEFERRED | T-012 or T-013 | Make-generated ledgers/schedules. | Accounting updated through owner inputs/generators only. |
| T-015 | Final broad validation evidence. | WF-08 | DONE | T-010, T-011, T-013 | `make check` run root `.cartulary/test-results/20260703T004155Z-p68590`. | Full repository check passed with `work_units=279/279`, `tests=977`, `failed=0`. |
| T-016 | Start authorized remaining remediation implementation run and record tracker baseline. | WS-00 | DONE | T-015 | This tracker; user-provided remediation plan. | Tracker states active implementation scope and update cadence. |
| T-017 | Clarify optional capability export semantics in Core/OpenAPI. | WS-01 | DONE | T-016 | `docs/spec/01_architecture_storage_and_view_contracts.md`; `contracts/openapi/cartulary.openapi.yaml`. | Core and OpenAPI state that current export admission rejects non-empty `required_capabilities[]` until capabilities are implemented. |
| T-018 | Add characterization tests for remaining gaps. | WS-02 | DONE | T-017 | `internal/modules/incidentbundles/api_test.go`; `internal/modules/incidentbundles/routes_integration_test.go`. | Tests cover export capability rejection, route finalizer setup, file-store behavior, job summaries, and dependency override hook. |
| T-019 | Add route assembly seam and app-wired import finalizer. | WS-03 | DONE | T-018 | `internal/modules/incidentbundles/routes.go`; `internal/modules/incidentbundles/api.go`; `internal/app/runtime.go`. | Claimed profile setup fails without import finalizer; app runtime wires the incidents-owned finalizer; export admission rejects unsupported non-empty required capabilities. |
| T-020 | Extract job result sink and replace global worker hook. | WS-04 | DONE | T-019 | `internal/modules/incidentbundles/worker_service.go`; `internal/modules/incidentbundles/worker_hooks.go`; tests. | Result summaries use package-private transition helpers and tests use test-runtime guarded dependency overrides instead of an exported global hook. |
| T-021 | Add file-store interface and localize idempotent admission helper. | WS-05 | DONE | T-018 | `internal/modules/incidentbundles/bundle_files.go`; `internal/modules/incidentbundles/routes.go`; `internal/modules/incidentbundles/worker_service.go`; `internal/modules/incidentbundles/store.go`. | File roots, permissions, cleanup, idempotency, and job auth scopes remain unchanged behind clearer private seams. |
| T-022 | Update harness accounting only if test names or evidence rows change. | WS-06 | DONE | T-019, T-020, T-021 | `tools/phase11_test_map.json`; generated phase ledger/schedule artifacts; generated contract artifacts. | Phase11 accounting includes new unit rows; generated ledgers/schedules were refreshed through Make; JSON, phase drift, artifact policy, and generated drift checks pass. |
| T-023 | Final validation and handoff closeout for remaining remediation. | WS-07 | DONE | T-022 | `make check` run root `.cartulary/test-results/20260703T012357Z-p93975`; WS-07 narrow validation run roots; this tracker. | Full repository check passed with `work_units=279/279`, `tests=981`, `failed=0`; no remediation blockers remain. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | Creating planning-only tracker after prior Plan Mode produced a complete plan. | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, `docs/domain.md`, Core 00/Core 01/Core 04, harness NLSpec, target package. | `git status --short`, `find`, `date -Is`; prior inspection commands listed below. | Scope and source posture recorded. | None. | Complete docs-only validation and final report. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | `incidentbundles` is a legitimate incident-portability facade and a mixed coordinator, not a proven permanent module boundary. | All files under `internal/modules/incidentbundles`; owner portability files; `internal/modules/incidentportability/portability.go`. | `rg`, `sed`, `find`, `wc -l`. | Boundary diagnosis and inventory populated. | Closed by descriptor-backed import registry in remediation pass. | Preserve owner-port descriptors for future source-family additions. |
| 2026-07-03T00:08:40-04:00 | Codex remediation session | Backend boundary drift now has a public Make-owned check. | `tools/backend_module_boundaries.json`, `tools/harness/static-analysis/backend-module-boundary-check-cli.mjs`, task-surface manifests. | `make backend-module-boundary-check`, `make help-all`, `make explain-target TARGET=backend-module-boundary-check DETAIL=summary`, `make lint`. | Boundary target passed and is discoverable from the public task surface. | None. | Keep any new exceptions in `tools/backend_module_boundaries.json` explicit and owner-reviewed. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | Incident-bundle contracts are Core 01 owned; generated files are downstream and not edited. | `contracts/openapi/cartulary.openapi.yaml`, `contracts/errors/index.json`, `contracts/extensions/index.json`, `packages/protocol-ts/src/index.test.ts`, Core 01 sections 12.3 and 17.5. | `rg`, `sed`. | Contract freeze map populated. | None for docs-only tracker. | Run drift checks if future owner inputs change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | Phase11 backend unit/integration rows cover current incident-bundle behavior; harness maps are evidence accounting only. | `api_test.go`, `routes_integration_test.go`, `tools/phase11_test_map.json`, `docs/testing/phase11_coverage_ledger.md`, scheduler/duration artifacts. | `make help`, `make help-all`, `make task-guide ROLE=feature-dev PHASE=phase11`, `make explain-target TARGET=backend-unit DETAIL=summary`, `make explain-target TARGET=backend-integration DETAIL=summary`, `make explain-target TARGET=frontend-unit DETAIL=summary`, `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`. | Validation commands discovered. | Closed by new `make backend-module-boundary-check`. | Use phase11 slices and full checks after code changes. |
| 2026-07-03T00:41:55-04:00 | User verification session | Full repository check passed after remediation. | Full changed tree. | `make check`. | Passed: `work_units=279/279`, `tests=977`, `failed=0`, run root `.cartulary/test-results/20260703T004155Z-p68590`. | None. | Use `make explain-run RESULTS_DIR=.cartulary/test-results/20260703T004155Z-p68590` for detailed retained evidence if needed. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | Export auth, import deployment-scope auth, and final-publication recheck are contract-owned and characterized. | `routes.go`, `store.go`, `worker_service.go`, `source.go`, Core 04 REQ-04-023, Core 01 REQ-01-609, integration tests. | `rg`, `sed`. | Auth risks mapped in Sections 4 and 5. | None for docs-only tracker. | Preserve tests before changing job policy or finalizer wiring. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-02T19:16:32-04:00 | Codex documentation session | Tracker created as a planning artifact. No production refactor run. | `docs/handoffs/incidentbundle-module-refactor-tracker-2.md`. | `apply_patch`; validation pending at creation time. | S-00 in progress. | Open questions in Section 11. | Run `make lint-markdown`, `make generated-artifact-policy-check`, and `make json-shape-check`. |
| 2026-07-02T19:54:33-04:00 | Codex remediation session | Q-001 through Q-005 were closed by descriptor-backed import targets, backend boundary target, explicit attribution resolver wiring, and saved-view import validation. | Core 01, harness NLSpec, tracker, incident portability owner ports, revisions, app runtime, task-surface owner input, generated task-surface artifacts. | `make phase-schedules`, `make backend-module-boundary-check`, `make backend-unit`, `make backend-integration`, `make lint`, `make generated-artifact-policy-check`, `make json-shape-check`, `make generate-drift`, `make agent-finalize`, `make phase-slice PHASE=phase11`, `make service-backed-slice PHASE=phase11`, `make test-fast`. | Targeted, phase, generated-drift, lint, and fast-suite validation passed; `make test-fast` passed on rerun after one infrastructure-only preflight cleanup timeout. | None. | Keep Section 11 closed; only future route/worker/file seam work remains deferred. |
| 2026-07-03T00:41:55-04:00 | User verification session | Final full-gate evidence is available. | Full changed tree. | `make check`. | Passed: `work_units=279/279`, `tests=977`, `failed=0`, run root `.cartulary/test-results/20260703T004155Z-p68590`. | None. | No remediation blocker remains. |
| 2026-07-02T21:00:32-04:00 | Codex remaining-remediation implementation session | WS-00 tracker baseline completed for the newly authorized implementation plan. | This tracker. | `git status --short`, `rg`, `sed`, `date -Is`. | Worktree was clean; active workstream rows T-016 through T-023 added. | None. | Start WS-01 spec and contract cleanup. |
| 2026-07-02T21:02:15-04:00 | Codex remaining-remediation implementation session | WS-01 spec and contract cleanup completed. | `docs/spec/01_architecture_storage_and_view_contracts.md`; `contracts/openapi/cartulary.openapi.yaml`; this tracker. | Not run; deferred until WS-02/WS-03 because implementation and tests still need to be aligned with the new contract text. | Core 01 and OpenAPI now describe the current empty implemented required-capability export set. | None. | Start WS-02 characterization tests. |
| 2026-07-02T21:07:17-04:00 | Codex remaining-remediation implementation session | WS-02 characterization tests added. | `internal/modules/incidentbundles/api_test.go`; `internal/modules/incidentbundles/routes_integration_test.go`; this tracker. | Not run; tests intentionally reference WS-03/WS-04/WS-05 seams that are not implemented yet. | Coverage added for export capability rejection, route dependency setup, worker result summary helpers, file-store path/permission cleanup, and dependency-override hook guarding. | None. | Start WS-03 route assembly seam. |
| 2026-07-02T21:08:27-04:00 | Codex remaining-remediation implementation session | WS-03 route assembly seam completed. | `internal/modules/incidentbundles/routes.go`; `internal/modules/incidentbundles/api.go`; `internal/app/runtime.go`; this tracker. | Not run; deferred until WS-04/WS-05 compile all new test seams. | `incidentbundles.RegisterRoutes` accepts `WithImportFinalizer`, app runtime wires the incidents finalizer, and non-empty export `required_capabilities[]` now fails at admission. | None. | Start WS-04 worker result sink and hook cleanup. |
| 2026-07-02T21:10:36-04:00 | Codex remaining-remediation implementation session | WS-04 worker result sink and hook cleanup completed. | `internal/modules/incidentbundles/worker_service.go`; `internal/modules/incidentbundles/worker_hooks.go`; `internal/modules/incidentbundles/api_test.go`; `internal/modules/incidentbundles/routes_integration_test.go`; this tracker. | Not run; deferred until WS-05 completes remaining compile-affecting file-store/admission changes and formatting. | Global mutable worker hook was replaced by a test-runtime guarded dependency override; export/import success summaries are constructed by package-private helpers used by the worker result sink. | None. | Start WS-05 file-store interface and admission cleanup. |
| 2026-07-02T21:12:12-04:00 | Codex remaining-remediation implementation session | WS-05 file-store interface and admission cleanup completed. | `internal/modules/incidentbundles/bundle_files.go`; `internal/modules/incidentbundles/routes.go`; `internal/modules/incidentbundles/worker_service.go`; `internal/modules/incidentbundles/store.go`; this tracker. | Not run; WS-06 begins formatting and drift/accounting checks. | Filesystem staging/persistence is behind a private adapter interface; export/import job admission share one private idempotent replay/conflict helper while preserving the single DB transaction. | None. | Start WS-06 formatting, drift, and accounting checks. |
| 2026-07-02T21:17:33-04:00 | Codex remaining-remediation implementation session | WS-06 harness and drift updates completed. | `tools/phase11_test_map.json`; `docs/testing/phase11_coverage_ledger.md`; `tools/execution_topology_render_index.json`; `internal/gen/contracts/contracts_gen.go`; `packages/protocol-ts/src/generated/contracts.ts`; this tracker. | `make format`; `make phase-ledgers`; `make phase-schedules`; `make generated-artifact-policy-check`; `make json-shape-check`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make phase-test-name-check`; `make generate`; `make generate-drift`. | Formatting, phase ledger/schedule generation, phase drift, JSON shape, generated artifact policy, phase test-name, and final generated drift checks passed. One pre-refresh `make phase-ledgers` run failed because the new incident-bundle unit rows were not yet in `expected_ids`; one pre-refresh `make generate-drift` run found OpenAPI generated contract drift. Both were resolved through Make-owned owner-input/generator updates. Final generated drift root: `.cartulary/test-results/20260703T011715Z-p70580`. | None. | Start WS-07 final validation and handoff. |
| 2026-07-02T21:28:44-04:00 | Codex remaining-remediation implementation session | WS-07 validation and handoff completed using user-supplied full-check evidence. | Full changed tree; this tracker. | `make lint-markdown`; `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check`; `make backend-unit`; user-supplied `make check`. | Narrow checks passed: backend module boundary root `.cartulary/test-results/20260703T011807Z-p72343`, generated artifact policy root `.cartulary/test-results/20260703T011807Z-p72363`, JSON shape root `.cartulary/test-results/20260703T011807Z-p72397`, backend unit root `.cartulary/test-results/20260703T011818Z-p73352`. Full `make check` passed with `work_units=279/279`, `tests=981`, `failed=0`, run root `.cartulary/test-results/20260703T012357Z-p93975`. The in-progress standalone `make backend-integration` run was superseded and is not used as evidence. | None. | Optional retained-run maintenance: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260703T012357Z-p93975`; otherwise review and commit the remediation set. |

Prior non-mutating inspection commands used as source evidence:

- `git status --short`
- `wc -l docs/handoffs/cartulary_modular_refactor_planning_framework.md docs/handoffs/incidentbundle-module-refactor-tracker.md docs/domain.md internal/modules/incidentbundles/*.go`
- `find internal/modules/incidentbundles -maxdepth 1 -type f -print | sort`
- `rg --files internal/modules/incidentbundles docs/handoffs docs/spec docs | sort`
- `rg -n "github.com/JochiRaider/cartulary/internal/modules/incidentbundles|incidentbundles\\." --glob '*.go' --glob '*.tsx' --glob '*.ts' .`
- `rg -n "incidentbundles|incident_portability|/api/v1/incident-bundles" internal/app apps/web packages contracts db/queries db/migrations tools`
- `rg -n "incident-bundles|incident_bundle|incident_portability|INCIDENT_BUNDLES" tools/phase11_test_map.json docs/testing/phase11_coverage_ledger.md tools/scheduler_manifest.json tools/execution_topology_manifest.json tools/go_test_duration_baselines.json tools/generated_artifact_policy.json tools/schema_object_ownership_manifest.json`
- `make help`
- `make help-all`
- `make task-guide ROLE=feature-dev PHASE=phase11`
- `make explain-target TARGET=backend-unit DETAIL=summary`
- `make explain-target TARGET=backend-integration DETAIL=summary`
- `make explain-target TARGET=frontend-unit DETAIL=summary`
- `make explain-target TARGET=browser-e2e-webserver-backed DETAIL=summary`

One exploratory `rg` command included invalid root `contracts/tools`, returned partial
results, and was rerun cleanly with valid roots.

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| Q-001 | Closed: `incidentportability.ImportNDJSON` may remain only as a registered descriptor-backed owner-port helper, not as a raw table-name importer. | Prevents source-family imports from bypassing owner invariants. | Core 01 required-file/source-family registry; `incidentportability.ImportTargetDescriptor`; backend boundary check. | Closed by remediation pass; future source families must update Core 01 and the implementation registry together. |
| Q-002 | Closed: create and use `make backend-module-boundary-check` as the narrower backend import-boundary/static check. | Boundary drift should fail before broad refactors proceed. | `tools/backend_module_boundaries.json`, `tools/harness/static-analysis/backend-module-boundary-check-cli.mjs`, task-surface owner input, generated Make target. | Closed by remediation pass; target emits `cartulary.backend_module_boundary_summary.v1`. |
| Q-003 | Closed: attribution resolver registration moves from hidden `init()` to explicit app/runtime wiring. | Hidden registration can make revisions behavior depend on package import side effects. | `revisions.AttributionResolverRegistry`, `incidentbundles.ImportedAttributionResolver`, `internal/app/runtime.go`. | Closed by remediation pass; duplicate and missing resolver cases are unit-tested. |
| Q-004 | Closed: saved-view-specific characterization and validation were added before broader saved-view portability movement. | Saved views combine owner state, visibility, view schemas, and import security. | `savedviews/incident_bundle_portability_test.go`; `incidentbundles` phase11 round-trip tests. | Closed by remediation pass; future movement must preserve these tests. |
| Q-005 | Closed: no owner contradiction found. The framework omission of `incidentbundles` is a planning mismatch; Core 01 owns the Incident Portability profile and routes. | Prevents non-authoritative planning artifacts from blocking implementation. | Core 00 authority order; Core 01 Incident Portability sections; framework catalog inspection. | Closed; future contradictions must cite exact owner excerpts and must not treat research, phase maps, generated outputs, or guides as runtime owners. |

## 12. Binary Completion Criteria

The tracker is complete only when all of the following are true:

- Every file in `internal/modules/incidentbundles` is inventoried or explicitly
  out of scope.
- Every discovered public contract risk has an owner and test posture.
- Every proposed workflow has dependencies and exit criteria.
- Every implementation slice is behavior-preserving unless explicitly marked as
  requiring later authorization.
- Validation commands are discovered or marked as `TODO:` with reason.
- Handoff sections are current enough for another agent to continue without
  rediscovery.
- No production refactor was run while creating this tracker.
- Files inspected are summarized in this tracker.
- Commands run are summarized in this tracker and final response.
- The tracker was created at
  `docs/handoffs/incidentbundle-module-refactor-tracker-2.md`.
- Remaining blockers are listed in Section 11.

Completion status for this tracker and remediation pass:

- File inventory: complete for all files found under
  `internal/modules/incidentbundles`.
- Public contract risk map: complete for discovered HTTP, job, workbook,
  saved-view, projection, authorization, revision, generated-contract, and
  harness surfaces.
- Proposed workflows and slices: complete for planning handoff.
- Validation commands: discovered; `make backend-module-boundary-check` closes
  the prior backend-boundary command gap.
- Production refactor: original tracker creation ran no production refactor;
  this remediation session includes implementation patches authorized by the
  user.
- Final full-gate evidence: user-provided `make check` passed with
  `work_units=279/279`, `tests=977`, `failed=0`, run root
  `.cartulary/test-results/20260703T004155Z-p68590`.
- Remaining blockers: none recorded for Q-001 through Q-005. Future
  route/worker/file seam work remains explicitly deferred, not blocked.
