# incident-modules Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target paths:** `internal/modules/incidentbundles`,
  `internal/modules/incidentportability`, and `internal/modules/incidents`.
- **Derived target label:** `incident-modules`.
- **Output path:** `docs/handoffs/incident-modules-refactor-tracker.md`.
- **Planning snapshot:** branch `main`, commit
  `9e92af306726245f442039d6b40bcbf57058a729`, with a clean tracked worktree
  before this tracker was created on `2026-07-27`.
- **Execution baseline:** clean commit
  `2bd4cc7011237ef8cd80acdcae378163ca0fadf1`; the repository had no adopting
  stable-release tag, so the retained-v1 deprecation clock remains pending.
- **Status:** AG-001, AG-002, and AG-003 are complete. SL-00, SL-01, SL-04,
  SL-06, and SL-07 are complete; SL-02 is the recommended next slice.
- **Authorized change:** The `2026-07-27` adoption task authorized Core owner,
  downstream implementation, authored contract, generated projection,
  verification, fixture, and final handoff changes required by AG-001 through
  AG-003.
- **Non-goals:** No unrelated persistence decomposition, general cleanup,
  frontend refactor, visual change, migration, dependency change, commit,
  push, or pull request was included.
- **Execution gate:** The authorized AG-001 through AG-003 package satisfied
  its owner-adoption and same-source-state verification gates. Remaining
  slices retain their own authorization and completion boundaries.

The label is already normalized lowercase kebab case. It contains no spaces,
path separators, shell metacharacters, or unsafe filename characters.
`incident-modules` is a planning seam spanning three packages; it is not a
proposed permanent module or verification owner.

### Source hierarchy

1. Adopted subsystem NLSpecs govern only their named subsystem.
2. Core 00 through Core 04 govern current implementation-conformance behavior.
3. Core 05 governs only claim-bearing timed or fixture-sensitive publication
   and is not applicable to this planning task.
4. `docs/domain.md` and implementation-support guides supply terminology,
   boundary direction, and harness mechanics.
5. Current repository code and tests establish current implementation state.
6. Prior trackers, handoffs, and this planning framework are evidence only.

No owner-owner contradiction was found. If one is found later, the exact marker
`BLOCKED: owner contradiction` MUST be added without choosing a side. The
bundle-version finding was an owner-to-downstream mismatch, not an owner-owner
contradiction. AG-001 adopted v2 current export with retained v1 import into
Core 01, AG-002 adopted its binary acceptance surface into Core 04, and AG-003
aligned the downstream implementation and projections.

### Normative language and activation

`MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, and `MAY` are normative within this
tracker. They prescribe future refactor execution and its acceptance evidence.
They do not make this handoff a product-behavior owner.

A tracker requirement that preserves adopted behavior is binding on a later
authorized refactor slice. A tracker requirement that changed, added, or
removed observable product behavior activated only after its named Core
adoption gate became `DONE`. Core 01 and Core 04 are now authoritative for the
adopted behavior; this tracker remains execution and continuation evidence.

The final planning dispositions are:

- bundle version `2` is the sole current export version;
- bundle version `1` is retained as import-only compatibility;
- a nil or unconfigured `JobRunner` is unsupported when Incident Portability is
  claimed;
- every imported source family executes shape and semantic validation through a
  typed source-owner port before publication.

### Owner documents inspected

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/extension-subsystem-nlspec.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/graph_projection_nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`

Core 05 was checked for applicability and excluded because this tracker makes no
claim-bearing timed or fixture-sensitive publication.

`docs/research/nlspec-spec.md` supplies the prescriptive authoring discipline
used by Section 6 and Section 13; it is not a Cartulary product-behavior owner.
Research reports R01 through R09 were reviewed as non-normative decision
support and are dispositioned in Section 6.5.

### Repository files inspected

Every one of the 56 files under the three target paths was inspected and is
inventoried in Section 2. Adjacent exact-source inspection covered:

- `internal/app/server/runtime.go`,
  `internal/app/workbookassembly/startup.go`, and
  `internal/modules/reporting/export_materializer.go`;
- `internal/modules/projections/services.go`,
  `internal/platform/jobs/runner.go`, and shared application test support;
- the `module.incidentbundles`, `module.incidents`, and `module.workbook`
  OpenAPI owner fragments, the assembled HTTP-operation projection, and
  `contracts/incident-bundles/compatibility.json`;
- `contracts/verification/owners/module.incidentbundles.json`,
  `contracts/verification/owners/module.incidents.json`,
  `tools/test_families/module.incidentbundles.json`, and
  `tools/test_families/module.incidents.json`;
- `tools/backend_module_boundaries.json`,
  `tools/test_support_inventory.json`, and the public Make task surface;
- `apps/web/src/app/IncidentAdminPanel.tsx`,
  `apps/web/src/app/IncidentImportPanel.tsx`,
  `apps/web/src/app/IncidentLanding.tsx`, and
  `apps/web/src/workbook/WorkbookShell.tsx`;
- `docs/archive/incidentbundle-module-refactor-tracker-2.md` and
  `docs/archive/incidents-module-refactor-tracker-2.md`.

The archived trackers are historical evidence. Their completed work and
validation claims are not copied forward as current evidence. The framework
correctly lists `incidents` and `incidentbundles` as candidate boundaries, but
does not establish `incidentportability` or the combined `incident-modules`
label as a permanent module.

## 2. Current-State Repository Inventory

All 56 live files are in scope. No file is excluded merely because it is a test,
test-support adapter, or subpackage.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/incidentbundles/api.go` | Closed export/import request decoding, normalization, media types, result and error tokens. | `ExportRequest`, `ImportMetadataRequest`, `DecodeExportRequest`, `DecodeImportMetadata`, public constants. | Bundle routes and worker; package tests. | `platform/httpapi`, JSON and UUID support. | `api_test.go`, route integration. | Incident-bundle OpenAPI schemas, error registry, protocol bindings. | Incident Bundles transport/application adapter. | high | Request normalization and exact closed vocabularies are observable. |
| `internal/modules/incidentbundles/api_test.go` | Characterizes bundle request, archive, storage, route registration, worker result, OpenAPI, and test-hook behavior. | Test functions and local stubs only. | Go test catalog: incident-bundle unit rows. | Contract test facade, archive/storage helpers, package internals. | This file. | OpenAPI and error registries; bundle compatibility behavior. | Incident Bundles product evidence. | high | Includes current v2 manifest assertions and exact public error sets. |
| `internal/modules/incidentbundles/attribution_resolver.go` | Resolves imported source-actor attribution for Revisions. | `IncidentPortabilityProfileID`, `ImportedAttributionResolver`. | Revision assembly and imported-history attribution. | Revisions port and `incident_bundle_imported_attributions` SQL. | Bundle/history integration coverage; revision collaborator tests. | Imported revision/history contract. | Incident Bundles source-owner adapter for Revisions. | medium | Narrow provider is preferable to Revisions reading bundle tables. |
| `internal/modules/incidentbundles/bundle.go` | Builds and verifies logical bundles, manifests, checksums, archive safety, required files, and version compatibility. | `ManifestInput`, `BundleArchive`, `BundleManifest`, `VerificationInput`, `VerifiedBundle`, `VerificationError`, `BuildBundleArchive`, `VerifyBundle`. | Worker build/import paths and unit tests. | Archive libraries, canonical JSON, hashing. | `api_test.go`, route integration. | Core 01 bundle layout and `contracts/incident-bundles/compatibility.json`. | Incident Bundles archive adapter. | critical | Current export is v2; retained v1 import uses legacy Timeline filenames. |
| `internal/modules/incidentbundles/cross_owner_import.go` | Supplies the Core incident-bundle participant for bounded multi-owner final commit. | Transaction descriptor, provider, capabilities, participant, result constructors. | Application assembly, bundle worker transaction execution. | Cross-owner transaction, Incidents finalizer, Jobs, object store, Postgres. | Extension portability and route integration. | Atomic import/publication and job finality contracts. | Incident Bundles mutation coordinator. | critical | Cross-owner writes are intentional only inside the declared transaction protocol. |
| `internal/modules/incidentbundles/extension_portability.go` | Enforces extension state/claim policy, participant admission, bounded preparation, staging, cleanup, and transaction participation. | `ExtensionPolicy`, invocation/result types, participant interface, staged scope, orchestrator, payload encoders, failure type. | Extension assembly, bundle builder/worker. | Cross-owner transaction and Staged Objects. | `extension_portability_test.go`, Network Flow bundle test. | Extension portability contracts and claim matrix. | Incident Bundles portability facade with Extension collaborators. | critical | Legitimate orchestration; must not absorb extension-owned state rules. |
| `internal/modules/incidentbundles/extension_portability_test.go` | Characterizes extension portability matrix, import cleanup, and closed descriptor behavior. | Test fakes and test functions only. | Incident-bundle unit and integration catalog rows. | Extension portability, Staged Objects, cross-owner transaction. | This file. | Extension portability verification contract. | Incident Bundles product evidence. | high | Protects participant and state/claim behavior. |
| `internal/modules/incidentbundles/job_finalization.go` | Defines terminal-success mutation and finality boundary for bundle jobs. | `ErrJobFinalizationIndeterminate`, `JobSuccessMutation`, `JobSuccessFinalization`, `JobSuccessFinalizer`. | Worker and application-provided extension job finalizer. | Jobs and cross-owner transaction. | Worker transition and route integration tests. | Common job resource and atomic publication contract. | Incident Bundles application port. | high | Finality ambiguity and cleanup behavior are observable. |
| `internal/modules/incidentbundles/limits.go` | Carries archive and incident-bundle resource limits. | `Limits`, `ArchiveLimits`, `IncidentBundleLimits`. | Routes, worker, archive verification, app configuration. | `platform/archivepolicy`. | Archive boundary and integration tests. | Core 04 resource-limit registry. | Incident Bundles configuration adapter. | high | Exact limit semantics must remain unchanged. |
| `internal/modules/incidentbundles/routes.go` | Registers three bundle operations and performs authentication, authorization, admission, staging, idempotency, session sliding, and recovery startup. | `Service`, `RouteOption`, `With*` options, `RegisterRoutes`. | Server runtime route assembly. | HTTP/auth platform, Incidents access, store, worker, storage, portability, transactions. | `api_test.go`, `routes_integration_test.go`. | `module.incidentbundles` OpenAPI owner fragment. | Incident Bundles transport facade. | critical | Export requires deployment admin plus membership; import is deployment-admin scoped. |
| `internal/modules/incidentbundles/routes_integration_test.go` | End-to-end export/import, authorization, idempotency, compatibility, atomic visibility, storage, projection, and optional-extension evidence. | Integration test functions and local harness helpers. | Incident-bundle integration catalog rows. | Full app harness, Timeline test support, object store, Postgres. | This file. | Bundle routes, jobs, archive, source-family, projection, and authorization contracts. | Incident Bundles product evidence. | critical | Primary broad characterization; uses incident scenario harness across owner boundaries. |
| `internal/modules/incidentbundles/source.go` | Builds and imports core incident source families, actors, blobs, attribution, revision repair, projection rebuild, and import finalization. | `BundleBuilder`, `Importer`, `BuiltIncidentBundle`, `ImportParams`, `PreparedImport`. | Worker and cross-owner import provider. | Incidents plus Records, Timeline, Parties, Entities, Indicators, Artifacts, Tasks/Decisions, Evidence, Assessments, Links, Revisions, Saved Views, Collaboration, Object Store, Postgres. | Route round-trip and failure integration tests. | Complete bundle source registry, revision, saved-view, and projection contracts. | Incident Bundles core portability coordinator. | critical | Legitimate coordinator with excessive concrete peer imports; candidate for an injected ordered owner-port catalog. |
| `internal/modules/incidentbundles/storage_port.go` | Defines opaque root-free staging/publication references and storage lifecycle port. | Reference types, parsers, `BundleStorage`. | Routes, worker, app storage adapter. | Standard library only. | Storage-reference and root-storage tests. | Core 04 runtime-root and path-safety behavior. | Incident Bundles storage port. | high | Intentional deep boundary; no raw filesystem mutation in module code. |
| `internal/modules/incidentbundles/store.go` | Persists bundle job payloads, export descriptors, manifests, and route idempotency; reads common jobs. | `Store`, acceptance/result/payload/descriptor types and methods. | Routes, worker, integration tests. | Authn and Jobs types; Postgres SQL. | Route integration and worker tests. | Job resource, descriptor, idempotency, and storage-reference contracts. | Incident Bundles private persistence adapter. | high | Direct SQL is mainly bundle-owned state; common job finality remains platform-owned. |
| `internal/modules/incidentbundles/worker_hooks.go` | Provides production-compiled dependency override parsing and exported worker-start test constructors. | `TestOption`, `WithWorkerStartHookForTesting`, `DependencySetForTesting`. | Incident-bundle tests and worker construction. | `platform/httpapi` test runtime guard. | `api_test.go`, authorization/final-publication integration tests. | Test-only behavior; no product contract. | Application/test support, not permanent production API. | medium | Test assumptions leak through exported production symbols. |
| `internal/modules/incidentbundles/worker_service.go` | Registers, recovers, dispatches, executes, and finalizes export/import jobs. | Package-private worker and result sink. | Routes and app-provided Jobs runner. | Jobs, cross-owner transaction, Incidents finalizer, storage, portability, importer. | Worker unit assertions and route integration. | Common job, bundle publication, import atomicity, and cleanup contracts. | Incident Bundles application/job adapter. | critical | IMR-JOBS resolves nil/unconfigured composition as unsupported; SL-01 and SL-06 MUST add named-runner evidence and remove every fallback after AG-001/AG-002. |
| `internal/modules/incidentportability/import_targets.go` | Central registry of owner/source family, bundle path, target relation, stable identity, required columns, and stated invariant policies. | `ImportTargetDescriptor`, registry access and validation helpers. | Owner incident-bundle portability adapters and shared importer. | Standard library only. | `import_targets_test.go`; bundle round trip indirectly. | Core 01 required source-family registry. | Source-owner portability metadata, currently centralized. | high | Has no independent verification owner; invariant descriptions are data, not enforcement. |
| `internal/modules/incidentportability/import_targets_test.go` | Characterizes registered descriptors, stable identity, rejection of unregistered targets, and export-only tags. | Test functions only. | Unmapped by a dedicated incidentportability owner; runs only through broader package execution if selected. | Package internals. | This file. | Bundle source-family registry. | Incident Portability support evidence. | medium | No `module.incidentportability` verification owner exists. |
| `internal/modules/incidentportability/portability.go` | Encodes/decodes NDJSON, imports registered rows with dynamic fixed-registry relations, remaps top-level user IDs, records attribution, and canonicalizes JSON. | `Queryer`, `File`, `AttributionRecorder`, error types, export/import/decode/remap/canonical helpers. | Incident-bundle adapters across source-owner modules. | pgx and standard JSON. | Descriptor tests and owner/bundle integration tests. | Structured bundle files and attribution semantics. | Owner-neutral serialization plus persistence-adjacent helper. | critical | Dynamic relation SQL is bounded by a private registry key but still centralizes owner persistence behavior. |
| `internal/modules/incidents/access.go` | Stable facade for visible incident, membership, and transaction-bound open-state decisions. | `Access`, `AccessService`, constructors and error classifiers. | Workbook, Timeline, Revisions, Evidence, Entities, Imports, Reporting, Saved Views, Collaboration, Assessments, Incident Bundles, Job API, Indicators, Artifacts, Tasks/Decisions. | Postgres interface, pgx transaction, Incidents store. | Incident unit and cross-module authorization tests. | Incident visibility, lifecycle, and hidden-not-found behavior. | Incidents application access port. | high | Legitimate small facade; peers should not depend on concrete Store. |
| `internal/modules/incidents/api.go` | Decodes and normalizes incident/lifecycle/membership requests and builds response resources. | Request and optional-value types, decoders, resource builders, last-admin guard. | Incident routes and package tests. | Authn and HTTP API types. | Request, lifecycle, support, and conformance tests. | `module.incidents` OpenAPI schemas and error registry. | Incidents transport/application contracts. | critical | Exact member sets, nullability, and normalization are observable. |
| `internal/modules/incidents/boundary_guard_test.go` | Scans production imports to forbid Workbook startup and platform WebSocket dependencies. | `TestIncidentsProductionImportBoundaries` and private parser helper. | Direct Go package tests only; no active incident catalog row found. | Go parser, filesystem. | This file. | Intended backend import-boundary policy. | Harness machine boundary policy. | medium | Its two unique rules are absent from `tools/backend_module_boundaries.json`. |
| `internal/modules/incidents/extra_integration_test.go` | Characterizes membership replay/conflict and incident/membership audit before/after values. | Integration test functions only. | Incident integration catalog rows. | Incident scenario/mutation helpers and Postgres. | This file. | Idempotency and administrative-audit contracts. | Incidents product evidence. | high | Protects deployment-local audit rather than revision history. |
| `internal/modules/incidents/hooks.go` | Reads a test-only precommit override from production dependency composition. | Package-private hook loader and hook structure. | Incident route/store construction and fault test helper. | `platform/httpapi` test runtime guard. | Incident rollback integration. | Test-only behavior. | Application/test support seam. | medium | Production-compiled test override; no exported API but still a test assumption in production. |
| `internal/modules/incidents/http_conformance_test.go` | Characterizes incident, membership, preference, query, Timeline, and live authorization route inventories. | Test functions and local helpers. | Incident support-integration catalog rows. | Route inventories from several owners, contract test facade, shared route harness. | This file. | HTTP envelopes, route inventory, authorization, workbook preferences. | Incidents product evidence with collaborators. | critical | Cross-owner route inventories are evidence accounting, not architecture. |
| `internal/modules/incidents/import_finalization.go` | Atomically creates the imported incident's initial admin membership, null workbook preferences, and administrative audit. | `ErrInitialAdminUnavailable`, Store finalizer method, `ImportBundleRequestID`. | Incident Bundles import transaction. | Generated SQL, preference bootstrap, administrative audit, users. | Store tests and bundle final-publication integration. | Core 01/Core 04 initial-admin publication contract. | Incidents import-finalization port. | critical | Must recheck active deployment-admin submitter before publication. |
| `internal/modules/incidents/incident_bundle_portability.go` | Exports/imports the incident source row and remaps source attribution. | `ExportIncidentBundleIncident`, `ImportIncidentBundleIncidentTx`. | Incident Bundles source coordinator. | Shared incidentportability and fixed Incidents SQL. | Bundle round-trip integration. | `data/incident.json` behavior. | Incidents source-owner portability adapter. | high | Legitimate owner adapter; should receive owner-neutral decoded data rather than generic relation control. |
| `internal/modules/incidents/integration_test.go` | Broad incident create, replay, authorization progression, patch, membership, preference, query, Timeline/live, and extension dispatch evidence. | Integration test functions plus contract helpers. | Incident integration catalog rows. | Full app harness and multiple owner route inventories. | This file. | Incident HTTP/auth, preferences, extensions, and idempotency contracts. | Incidents product evidence. | critical | Primary service-backed authorization matrix. |
| `internal/modules/incidents/inventory_helpers_test.go` | Builds cross-owner route fixtures, executes HTTP/WS controls, and seeds preference/Timeline state. | Package-local fixture and conversion helpers. | Incident conformance and integration tests. | Timeline test support, route inventory, HTTP/WS harness, Postgres. | Incident test files. | Test route/control accounting. | Incidents test support. | medium | Test composition does not define runtime ownership. |
| `internal/modules/incidents/lifecycle_access.go` | Owns create bootstrap defaults, request hashes, patch application, and membership/role error precedence. | Bootstrap/hash/patch/access helpers. | Routes and store. | HTTP API and UUID. | Lifecycle, request, unit, and integration tests. | Incident lifecycle and authorization contracts. | Incidents domain/application policy. | critical | Hidden 404 versus 403 precedence is observable. |
| `internal/modules/incidents/lifecycle_contract_test.go` | Characterizes lifecycle request errors and exact OpenAPI operations. | Test functions only. | Incident unit catalog. | Contract test facade and API decoders. | This file. | Lifecycle OpenAPI and error contracts. | Incidents product evidence. | high | Protects close/reopen request and envelope parity. |
| `internal/modules/incidents/lifecycle_integration_test.go` | Exercises close/reopen authorization, idempotency, transitions, audit, and collaboration effects. | Integration test function only. | Incident integration catalog. | Full app harness, mutation helper, route idempotency. | This file. | Lifecycle, collaboration, and audit contracts. | Incidents product evidence. | critical | Close/reopen effects cross Incidents and Collaboration. |
| `internal/modules/incidents/membership_audit_contract_test.go` | Characterizes exact membership-audit OpenAPI operation and envelope. | Test function only. | Incident unit catalog. | Contract test facade. | This file. | Membership administrative-audit OpenAPI. | Incidents product evidence. | high | Administrative audit must remain distinct from revisions. |
| `internal/modules/incidents/membership_audit_handlers.go` | Implements incident membership administrative-audit listing and query validation. | Package-private handler methods. | Incident routes. | Administrative audit, HTTP auth/API, list query. | Membership audit contract/integration and conformance tests. | Membership audit route and pagination contract. | Incidents transport adapter. | high | Incident admin authorization and safe filters are observable. |
| `internal/modules/incidents/membership_audit_integration_test.go` | Exercises audit authorization, scoping, filters, and keyset continuation. | Integration test function only. | Incident integration catalog. | Full app harness and membership mutations. | This file. | Membership audit and pagination behavior. | Incidents product evidence. | high | Protects scope-safe audit reads. |
| `internal/modules/incidents/open_guard.go` | Implements the private transaction-bound incident-open check. | Package-private helper. | Access facade and source-state mutation consumers. | Generated SQL and pgx transaction. | Incident lifecycle and peer closed-state tests. | Closed-incident mutation matrix. | Incidents lifecycle persistence adapter. | high | Same transaction is required; it does not own transaction lifecycle. |
| `internal/modules/incidents/openapi_contract_test.go` | Proves runtime route descriptors, responses, security, and schemas match OpenAPI. | Test function and local schema assertions. | Incident unit catalog. | Contract test facade. | This file. | `module.incidents` OpenAPI owner fragment and assembled document. | Incidents contract evidence. | critical | Generated/assembled artifacts must be changed through authored owners. |
| `internal/modules/incidents/pagination_integration_test.go` | Characterizes live incident/membership pagination, query scope, and membership revocation during continuation. | Integration test functions only. | Incident integration catalog. | Full app harness, cursor behavior, Postgres. | This file. | List/cursor authorization contracts. | Incidents product evidence. | high | Continuations are live-authorized, not immutable snapshots. |
| `internal/modules/incidents/ports.go` | Declares preference bootstrap, bundle finalization, collaboration session, Store, and route options. | Port interfaces and option types. | Incidents store/routes, app composition, Incident Bundles. | pgx, UUID, time. | Unit registration and integration tests. | Internal application composition contracts. | Incidents application ports. | high | Collaboration and preference seams are intentional dependency inversion. |
| `internal/modules/incidents/reportingprovider/provider.go` | Supplies incident snapshot and source-boundary state to Reporting inside a caller transaction. | `GetIncidentSnapshotTx`, `ResolveSourceBoundaryStateTx`. | Reporting export materializer. | Reporting consumer DTOs, Incidents and change-set SQL. | Reporting integration and export-model tests. | Reporting immutable source-boundary contract. | Incidents source-owner reporting adapter. | high | Intentional source-owner provider; Reporting owns materialization semantics. |
| `internal/modules/incidents/request_test.go` | Characterizes create/patch/membership decoding, list errors, workbook preference OpenAPI, and extension dispatch. | Test functions and local lookup stub. | Incident unit/support-unit catalog. | Contract test facade, Authn, extension profile contracts. | This file. | Incident and Workbook OpenAPI plus extension discovery. | Incidents product evidence. | high | Workbook preference assertions collaborate with `module.workbook`. |
| `internal/modules/incidents/routes.go` | Registers 11 incident operations and coordinates authentication, authorization, pagination, lifecycle, membership, audit, and collaboration notifications. | `Service`, `RegisterRoutes`. | Server runtime. | Authn, HTTP auth/API, list query, pagination, Access/Store, collaboration port. | All incident route contract/integration suites. | `module.incidents` OpenAPI routes. | Incidents transport facade. | critical | Route-level auth is correctly transport-adjacent and re-derived live. |
| `internal/modules/incidents/store.go` | Coordinates incident create/list/get/patch/lifecycle and membership CRUD with idempotency, preferences, and administrative audit. | `Store`, records/results/page/error types, public store methods. | Routes, Access, app/test support, import finalizer. | Generated SQL, workbook preferences, administrative audit, Authn, list query, Postgres. | Store, integration, pagination, conformance tests. | Incident storage, audit, preference, idempotency, and lifecycle contracts. | Incidents application plus private persistence. | critical | Mixed application/persistence responsibility is a split candidate within the Incidents owner. |
| `internal/modules/incidents/store_test.go` | Characterizes create/finalization atomicity, rollback, replay, stable location, version conflicts, and membership concurrency. | Test functions and failure fake. | Incident store/integration catalog rows. | Incident storetest, Postgres, preference port. | This file. | Incident persistence, import finalization, audit, and preference contracts. | Incidents product evidence. | critical | Essential before splitting Store internals. |
| `internal/modules/incidents/support_test.go` | Unit-characterizes membership patch/delete decoding and last-admin guard. | One test function. | Incident support-unit catalog row. | Package API helpers. | This file. | Membership mutation contract. | Incidents product evidence. | medium | Cataloged under semantic support-unit family. |
| `internal/modules/incidents/testsupport/faulttest/incidents.go` | Constructs guarded rollback-fault dependencies for incident create tests. | `IncidentCreateRollbackFaultDependencies`. | Incident integration test. | Production dependency override key through `httpapi.DependencySet`. | Incident rollback integration. | Test-only transaction rollback evidence. | Incidents test support. | medium | Coupled to production-compiled hook key. |
| `internal/modules/incidents/testsupport/mutationtest/mutations.go` | Queries and asserts incident resource, membership, and audit mutation artifacts. | Mutation owner types, database adapters, selectors, assertion helpers. | Incident integration/store tests. | Postgres interfaces and administrative-audit SQL. | Incident test suites. | Test mutation/audit accounting only. | Incidents test support. | low | Semantically incident-owned and not a runtime boundary. |
| `internal/modules/incidents/testsupport/routetest/routes.go` | Declares semantic incident and membership public/control route inventories. | Mutation-owner constants and inventory functions. | Incident conformance/integration tests. | Shared route inventory. | Incident tests. | HTTP/control test accounting. | Incidents test support. | medium | Owner-local semantic inventory is appropriate. |
| `internal/modules/incidents/testsupport/scenariotest/harness.go` | Wraps shared application runtime/server composition with convenience start modes. | `RuntimeHarness`, `ServerHarness`, `StartRuntime`, server start methods. | Incident tests plus Collaboration, Imports, Incident Bundles, Job API, Network Flow, Reference Data, Reporting, Saved Views, Timeline, View Schemas, Workbook, and platform tests. | `internal/testutil/appsupport`, HTTP API, HTTP test runtime. | Many cross-owner integration suites. | Test runtime composition and fixture policy. | Shared application test support. | high | Live cross-owner use conflicts with its owner-local inventory rationale. |
| `internal/modules/incidents/testsupport/scenariotest/incidents.go` | Provides incident and membership HTTP fixture actions. | Create/patch incident and membership helper functions. | Incident and cross-owner integration tests needing incident fixtures. | Auth test flow and HTTP test server. | Many integration suites. | Incident HTTP fixture behavior only. | Incidents source-owner test support. | medium | Keep incident-specific actions with Incidents even if callers are collaborators. |
| `internal/modules/incidents/testsupport/storetest/harness.go` | Starts a rollback-scoped incident Store harness. | `StoreHarness`, `StartStore`. | Incident, Auth, Network Flow, Saved Views, Timeline, and Workbook store tests. | Incidents Store, shared app support, pgtest, Postgres. | Cross-owner store tests needing incident state. | Test fixture policy. | Incidents source-owner store test support. | medium | Exposes Incidents Store, so it is not owner-neutral despite cross-owner callers. |
| `internal/modules/incidents/testsupport/storetest/store.go` | Creates incidents/memberships and snapshots replay side effects for store tests. | Replay DTOs, database adapters, fixture/query helpers. | Incident and cross-owner store tests. | Incidents, mutationtest, Authn, Postgres. | Store and startup tests. | Incident persistence test behavior. | Incidents test support. | medium | Domain-specific fixture adapter should remain owner-local. |
| `internal/modules/incidents/transaction_participant.go` | Supplies transaction-bound incident locking to cross-owner consumers. | `TransactionParticipant`, constructor, `LockIncidentTx`. | Cross-owner transaction/application assembly. | pgx and Incidents SQL. | Cross-owner transaction and lifecycle tests. | Transaction serialization behavior. | Incidents transaction port. | high | Legitimate narrow source-owner capability. |
| `internal/modules/incidents/unit_test.go` | Characterizes hidden-not-found/denied access and required collaboration route port. | Test functions and no-op collaboration fake. | Incident unit catalog. | Access and route composition. | This file. | Authorization precedence and composition contract. | Incidents product evidence. | high | Guards the explicit Collaboration dependency. |
| `internal/modules/incidents/workbookpreferences/bootstrap.go` | Atomically seeds incident-wide and creator/importer workbook preferences. | `Bootstrap`, constructor, port implementation. | Incidents create and import finalization. | Generated SQL. | Incident store/conformance and bundle import tests. | Core 02 preference schema and Core 01 create/import bootstrap. | Incident Workspace preference persistence. | high | Incident-owned persistence; Workbook owns the public preference routes. |
| `internal/modules/incidents/workbookpreferences/repository.go` | Reads, writes, locks, and conditionally repairs incident/user workbook preferences. | Records, `Repository`, `Session`, constructors and methods. | Workbook application assembly and Incidents bootstrap. | Generated SQL, Postgres, pgx. | Workbook startup/store and incident conformance tests. | Workbook preference route/startup contract. | Incident Workspace source-owner repository. | high | Conditional clear timestamps and attribution are observable. |

## 3. Module Boundary Diagnosis

The combined target is a **mixed planning seam**, not one module. The permanent
direction is two public owners—Incidents and Incident Bundles—plus owner-neutral
portability utilities and source-owner adapters.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Incident lifecycle, visible collection, metadata, and membership policy | `incidents` API, access, lifecycle, routes, and Store | Incidents | keep/split | Core 01–04; 11 OpenAPI operations; incident owner tests | Keep behavior and facade; split application coordination from persistence internally. |
| Incident membership administrative audit | Incidents Store and audit handler | Incidents with platform Administrative Audit substrate | keep | Core 02 deployment-local audit boundary; membership-audit tests | Must not become revision/change-set history. |
| Workbook preference persistence and create/import bootstrap | `incidents/workbookpreferences`, Store, import finalizer | Incident Workspace persistence; Workbook owns public routes | keep | Core 02 §11.2, Core 01 create/import, boundary guard, workbook assembly | Do not move solely because Workbook consumes the repository. |
| Collaboration close/revocation effects | Incidents routes through `CollaborationSessionPort` | Incidents intent plus Collaboration session/stream owner | keep | Core 03, Core 04, route/unit tests | Correct dependency inversion; no direct platform WebSocket dependency. |
| Reporting incident snapshot/source boundary | `incidents/reportingprovider` | Incidents source-owner adapter consumed by Reporting | keep | Adopted Reporting NLSpec and export materializer interface | Reporting owns export-model derivation, not live incident SQL. |
| Incident source portability | `incidents/incident_bundle_portability.go` | Incidents source-owner adapter | keep/split | Bundle round-trip and Core 01 | Keep fixed owner semantics; remove dependence on central relation policy when safe. |
| Bundle HTTP facade and job admission | `incidentbundles/api.go`, `routes.go` | Incident Bundles | keep | Core 01 routes and Core 04 authorization | Legitimate thin public facade over deeper orchestration. |
| Archive/version/integrity behavior | `incidentbundles/bundle.go` | Incident Bundles under Core 01 | keep, gated for format edits | Core 01 §12.3, compatibility projection, archive tests | RB-001 is resolved to v2 current export and retained v1 import; AG-001 and AG-002 MUST close before format movement. |
| Core source-family export/import order | `incidentbundles/source.go` | Incident Bundles coordinator consuming source-owner ports | split | Concrete imports of 13 owner modules; atomic route tests | Keep orchestration and order; inject owner ports from application composition. |
| Extension portability and bounded final commit | `incidentbundles/extension_portability.go`, `cross_owner_import.go` | Incident Bundles facade with Extensions, Staged Objects, and Cross-owner Transaction owners | keep/split | Adopted Extension NLSpec and transaction tests | Do not absorb participant-owned validation or state. |
| Bundle job/store/storage mechanics | Incident Bundles worker, Store, and storage port | Incident Bundles application and private adapters; Jobs/Object Store platform | split | App runtime composition, Jobs runner, storage policy | Keep root-free port; isolate private persistence and worker mechanics. |
| Shared NDJSON and attribution encoding | `incidentportability/portability.go` | Owner-neutral Incident Portability utility | keep/split | Cross-owner adapters use it; no own route or verification owner | Retain pure codec/remap helpers. |
| Source-family descriptor and target-relation policy | `incidentportability/import_targets.go` | Individual source owners plus an application-composed catalog | move/split | Central registry duplicates owner paths, relations, columns, and prose invariant names | Owner adapters should own fixed storage/invariant rules. |
| Workbook-grid projection rebuild after import | Injected `importProjectionRebuilder` | Projections owns physical rebuild; Incident Bundles invokes it | keep | Core 01/Core 03; Graph Projection NLSpec explicitly excludes workbook-grid rebuild | Existing injected port is the correct boundary. |
| Reusable application test runtime | `incidents/testsupport/scenariotest/harness.go` | `internal/testutil/appsupport` | move | Cross-owner callers and repository application-test boundary | Retain incident-specific HTTP helpers under Incidents. |
| Production-compiled test hooks | Incident Bundles worker hook and Incidents precommit hook | Jobs/application test composition | split/defer | Live test-only dependency overrides | Characterize runner/failure seams before deleting hooks. |
| Frontend incident shell/controller state | Outside target in `apps/web` | Web application | defer | Incident admin/landing/import panels and Workbook shell | Freeze contract only; no frontend movement in backend slices. |
| Grid-vendor integration | None in target | Grid Adapter package | intentional/no_action | No direct target import or vendor coordinate use found | UI selectors are downstream contract risks only. |

## 4. Public Contract and Behavior Freeze Map

This documentation task changes no public API, internal interface, schema,
generated type, or wire contract.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Incident list/create/get/patch | Core 01; `module.incidents` | Incidents OpenAPI owner, routes, API, Store | Unit, integration, pagination, conformance, Vitest, Playwright | Preserve exact route inventory, envelopes, Location, normalized replay, live pagination | critical | Six operations across collection/singleton paths. |
| Incident close/reopen | Core 01–04; `module.incidents` | Lifecycle owner clauses, routes, OpenAPI | Lifecycle unit/integration, IncidentAdminPanel Vitest, browser | Preserve role, reason, version, replay, closed matrix, collaboration effects | critical | Two operations; source-state writes fail while allowed reads/exports remain. |
| Membership list/create/patch/delete | Core 01–04; `module.incidents` | OpenAPI, routes, Store | Conformance, integration, store, Vitest, browser | Preserve selector, role, version, last-admin, replay, session-role refresh | critical | Four operations. |
| Membership administrative-audit list | Core 01/02/04; `module.incidents` | OpenAPI owner and administrative-audit handler | Contract, integration, frontend/browser | Preserve incident-admin scope, filters, paging, redaction | high | Deployment admin alone is insufficient. |
| Workbook preference get/put for default and current user | Core 01–03; `module.workbook` | Workbook OpenAPI owner, workbook routes, incident repository | Workbook startup/store, incident conformance, browser | Preserve nullable `sheet_ref`, auth, conditional repair, timestamps and attribution | critical | Four operations; storage remains Incident Workspace-owned. |
| Bundle export/import/descriptor | Core 01/04; `module.incidentbundles` | Bundle OpenAPI owner, routes, API | Unit and route integration | Preserve upload envelope, canonical selectors, descriptor, idempotency, errors | critical | Three operations. |
| Common job resource, polling/cancel, and terminal summaries | Core 01/04; Jobs/Job API with Incident Bundles | Worker, job finalizer, route result refs | Worker unit, bundle integration, Job API tests | Prove claimed-profile rejection of nil/unconfigured runners, named-handler recovery after activation, and durable recovery after dispatch failure | critical | Export job auth combines deployment admin and membership; import is deployment-scoped. |
| Bundle archive layout, checksum, member safety, versions | Core 01; Incident Bundles | Core 01 §12.3, `bundle.go`, compatibility projection | Archive unit and round-trip integration | Prove deterministic v2 export, retained-v1 translation, and rejection of unadmitted or mixed versions after AG-001 | critical | Version selection is manifest-only; no format edit may precede AG-001 and AG-002. |
| Core source-family round trip | Each source owner; Incident Bundles coordinates | Owner portability adapters and `source.go` | Broad bundle integration, selected owner tests | Prove exact path accounting, owner semantic validation, duplicate rejection, and failure rollback for every family | critical | Covers Records, Timeline, Parties, Entities, Indicators, Artifacts, Tasks/Decisions, Evidence, Assessments, Links, Revisions, Saved Views, and Incidents. |
| Historical actors and attribution | Core 01/02/04; Incident Bundles/Revisions | Actor files, attribution resolver, importer | Bundle history/rollback integration | Preserve inert actor behavior and imported attribution mapping | high | Import must not create login-capable users or memberships. |
| Revision/change-set/history preservation | Revisions with source-owner adapters | Bundle source coordinator and Revisions repair | Bundle round trip and Reporting/history tests | Preserve counts, selectors, sequence repair, rollback preconditions | critical | Administrative audit remains excluded. |
| Import atomic visibility and initial admin | Core 01/04; Incident Bundles plus Incidents finalizer | Cross-owner participant and import finalizer | Failure-family, final-publication, store tests | Add failure injection at remaining owner/invariant boundaries where absent | critical | No partial visible incident, preferences, membership, audit, or projections. |
| Workbook projection rebuild after import | Core 01/03; Projections | Injected rebuild service | Bundle open-after-import and superseded Timeline tests | Preserve rebuild-in-transaction and failure rollback | critical | Graph Projection NLSpec does not own this workbook-grid rebuild. |
| Saved-view/view-schema references | Saved Views, Workbook, view-contract owners | Bundle saved-view adapter, preference repository, frontend | Bundle integration, workbook startup/query tests | Preserve immutable schema IDs, saved view IDs, nullable repairs | high | Bundle/import must not infer behavior from display labels. |
| Incident WebSocket path and authorization effects | Core 03/04; Collaboration | Incidents collaboration port and collaboration stream | Lifecycle, collaboration, socket/browser tests | Preserve close error and membership-loss termination/recheck | critical | Target does not own live hub or public sequencing. |
| Authorization precedence and live re-derivation | Core 01/04; Incidents/Incident Bundles | Access helpers, routes, Job API | Control-boundary matrices, bundle auth tests | Preserve hidden 404, 403, deployment-admin combination, final recheck | critical | Authorization must occur before validation/detail disclosure where specified. |
| Generated OpenAPI/protocol contracts | OpenAPI owner fragments and generators | Module owner JSON, assembled OpenAPI, protocol-ts bindings | Contract unit tests, drift targets | No hand edits; regenerate only after authored owner changes | high | This task touches none. |
| Frontend lifecycle/membership/import and UI selectors | Web application and UI Contracts | IncidentAdminPanel, IncidentImportPanel, IncidentLanding, WorkbookShell | Five Vitest rows and seven Playwright rows under `module.incidents` | Run owner browser rows and broader browser gate if public behavior changes | high | No direct grid-vendor dependency in target. |
| Harness/test accounting | Testing Harness NLSpec and machine catalogs | Verification contracts and family manifests | 22 Incident Bundles rows; 40 Incidents rows | Reconcile boundary guard and test-support ownership before file moves | high | Incident Bundles: 11 unit/11 integration, Go only. Incidents: 28 Go/7 Playwright/5 Vitest; 25 service-backed. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Core 01 closes `bundle_version=1` with `timeline_time_conversion_profiles.ndjson` and `timeline_events.ndjson`, while code and compatibility projection export v2 with `timeline_time_profiles.ndjson`, `timeline_records.ndjson`, and `timeline_source_provenance.ndjson`. | Core 01 §12.3; `bundle.go`; compatibility JSON; current tests. | critical | `must_fix` | Core 01 owner, then Incident Bundles projection/implementation. | Adopt the IMR-BUNDLE v2-export/v1-import disposition through AG-001 and align downstream artifacts through AG-003. |
| `incidentbundles/source.go` directly imports many source-owner modules. | Exact import block and ordered build/import calls. | high | `should_fix` | Incident Bundles consumer port plus application composition. | Design an ordered typed source-port catalog; preserve file paths, order, transaction, attribution, and rollback. |
| Central import-target registry duplicates owner modules, paths, relations, stable IDs, required columns, and prose invariant names. | `incidentportability/import_targets.go`. | high | `should_fix` | Individual source owners; composed catalog for coordination. | Move fixed persistence and invariant rules behind source-owner adapters after characterization. |
| Shared importer constructs target relation/type dynamically from a private registered descriptor. | `incidentportability.ImportRows`; backend boundary allows raw helper only inside package. | high | `should_fix` | Source-owner persistence adapters. | Reduce shared package to codec/remap utilities; owner adapters use fixed SQL/SQLC. |
| Descriptor `PostImportInvariantCheck` values are descriptive, while shared validation checks registration, source identity, and required columns. | Descriptor fields and `ImportRows` implementation. | critical | `must_fix` | Each source owner. | Implement IMR-PORT through SL-04; every source owner MUST supply executable invariant validation and rollback evidence before central policy removal. |
| Incident scenario runtime wrapper is registered as owner-local but imported by numerous other owners. | Live inbound import scan; `tools/test_support_inventory.json` rationale. | medium | `must_fix` | Shared application test support for runtime composition; Incidents for incident actions. | Move only owner-neutral runtime convenience; keep incident HTTP/store helpers owner-local and update authored inventory. |
| Incidents production import guard has no active incident catalog selector and its two rules are absent from the machine boundary input. | Boundary test, incident family manifest, backend boundary JSON. | medium | `must_fix` | Harness backend boundary policy. | Project both rules into the authored machine policy, validate parity, then retire the redundant source scanner. |
| Production-compiled test hooks expose or consume module override keys. | Incident Bundles `worker_hooks.go`; Incidents `hooks.go`; test callers. | medium | `should_fix` | Jobs/application test support. | Characterize worker gating and transaction failure first; replace with owned seams and remove hooks only after callers migrate. |
| Bundle worker recovery has a nil-runner path that silently skips recoverable work because the nil receiver method returns success, while new dispatch can use an unmanaged goroutine; production runtime always injects a runner. | `worker_service.go`, `platform/jobs/runner.go`, routes startup, server runtime Jobs construction. | critical | `must_fix` | Incident Bundles plus platform Jobs. | Implement IMR-JOBS through SL-06: claimed composition requires a configured named runner and every nil, anonymous, or unmanaged fallback MUST be removed. |
| Incidents Store mixes application coordination and persistence. | Store methods span create/list/patch/lifecycle/membership, audit, preferences, and idempotency. | high | `should_fix` | Incidents. | Split private repositories/services while retaining Access and observable transactions. |
| Incident routes perform authentication/authorization through transport/platform helpers and Incidents access policy. | Incidents and bundle routes. | low | `intentional/no_action` | Transport adapter plus Incidents policy. | Preserve live re-derivation and error precedence. |
| Incidents reaches Collaboration through an explicit port. | `ports.go`, `routes.go`, boundary guard, unit test. | low | `intentional/no_action` | Incidents/Collaboration seam. | Keep; do not replace with direct WebSocket or hub access. |
| Bundle storage uses opaque root-free references and an injected port. | `storage_port.go`, app storage adapter, boundary policy. | low | `intentional/no_action` | Incident Bundles port; platform storage adapter. | Preserve strict parsing and no raw filesystem mutation. |
| Projection rebuild is an injected transaction-bound port. | `importProjectionRebuilder`, Projections service, app assembly. | low | `intentional/no_action` | Projections. | Keep physical projection ownership outside Incident Bundles. |
| Reporting provider lives under the source owner and implements a Reporting-owned consumer interface. | Incidents provider and Reporting export materializer. | low | `intentional/no_action` | Incidents adapter/Reporting consumer. | Keep narrow; do not move reporting derivation into Incidents. |
| Workbook preferences live under Incidents while Workbook owns public routes. | Core 02, boundary guard, workbook assembly, repository callers. | low | `intentional/no_action` | Incident Workspace persistence and Workbook application. | Preserve the current ownership split. |
| No target code imports grid-vendor APIs or frontend shell/controller state. | Production import scan and frontend consumer inspection. | low | `intentional/no_action` | Web/Grid Adapter downstream owners. | Freeze public contracts only; no backend-driven frontend movement. |

## 6. Normative Refactor Contract

This section resolves RB-001 through RB-003 for planning. IMR requirements that
alter observable behavior remain inactive until their named adoption gate is
`DONE`; they MUST NOT be implemented from this tracker alone.

| Artifact | Required content | Authority constraint |
| --- | --- | --- |
| Core 01 body | Bundle/version behavior, compatibility and retirement rules, claimed-profile Jobs lifecycle, source-owner responsibility, coordinator order, atomic publication, and public failure semantics. | Own each observable behavior exactly once. |
| Core 04 body | Binary acceptance and verification ownership for every adopted Core 01 behavior. | An AC cannot create a behavior absent from an adopted requirement. |
| Explicitly normative Core appendix | Exhaustive versioned path/row registry, source-family/identity/dependency/invariant registry, and v1 translation mapping. | The Core body MUST identify the appendix as normative; a supporting appendix cannot close AG-001. |
| `contracts/**` and generated artifacts | Typed machine projections, compatibility data, schemas, errors, catalogs, fixtures, and verification routing. | Downstream only; runtime and release evidence MUST NOT infer requirements from this tracker or Markdown. |
| Supporting guides and implementation design | Concrete Go interfaces, wiring diagrams, migration sequencing, telemetry queries, and operator playbooks. | Non-normative; MAY vary when observable behavior remains identical. |
| `docs/domain.md` | Existing vocabulary and owner navigation. | MUST NOT own bundle filenames, version defaults, runner nilability, SQL targets, or import algorithms. |

### 6.1 Bundle-version and Timeline compatibility contract

**IMR-BUNDLE-001** — After AG-001 and AG-002, every newly generated Incident
Bundle MUST use `bundle_format='cartulary.incident_bundle'` and numeric
`bundle_version=2`. Version `1` MUST remain importable and MUST NOT be exported.

**IMR-BUNDLE-002** — Import MUST select exactly one codec from the parsed
`manifest.bundle_version` before interpreting any source-family payload.
Filename presence, file order, archive order, prior import history, and caller
input MUST NOT select or override the codec.

**IMR-BUNDLE-003** — `bundle_version` is required. Omission, JSON `null`, a
non-integer value, an integer other than an admitted version, or an archive
whose Timeline path set does not match the selected version MUST fail before
source preparation or visible mutation. No fallback version exists.

| Version/input | Export | Import | Timeline path set | Required result |
| --- | --- | --- | --- | --- |
| `2` | required current output | required | `data/timeline_time_profiles.ndjson`, `data/timeline_records.ndjson`, `data/timeline_source_provenance.ndjson` | Use only the v2 codec. |
| `1` | forbidden | required while retention conditions hold | `data/timeline_time_conversion_profiles.ndjson`, `data/timeline_events.ndjson` | Use only the v1 translator and produce current source state. |
| omitted, `null`, non-integer | forbidden | forbidden | none admitted | Fail with `incident_bundle_import_rejected` and `malformed_manifest`. |
| integer outside `{1,2}` | forbidden | forbidden | none admitted | Fail with `incident_bundle_import_rejected` and `unsupported_bundle_version`. |
| v1 paths with version `2`, v2 paths with version `1`, or both path sets | forbidden | forbidden | mixed or mismatched | Fail with `incident_bundle_import_rejected` and `malformed_manifest`. |

**IMR-BUNDLE-004** — The v2 Timeline row contracts adopted by AG-001 MUST
match the following closed shapes. “Required” means the member MUST be present;
members not listed under “Allowed members” MUST fail closed. A required member
MAY be JSON `null` only when its source-owner semantic contract admits null.

| File | Stable row identity | Required members | Allowed members |
| --- | --- | --- | --- |
| `timeline_time_profiles.ndjson` | `incident_id` | `incident_id`, `enabled`, `profile_version`, `updated_at` | required members plus `local_offset_minutes`, `local_label`, `updated_by_user_id` |
| `timeline_records.ndjson` | `record_id` | `record_id`, `incident_id`, `capture_state`, `activity_utc_generated`, `activity_local_generated`, `activity_time_pair_state` | required members plus `reviewed_by_user_id`, `reviewed_at`, `superseded_by_user_id`, `superseded_at`, `date_entered_text`, `analyst_text`, `mitre_stage_text`, `device_object_text`, `ip_address_text`, `activity_utc_text`, `activity_local_text`, `raw_activity_text`, `activity_synopsis_text`, `data_source_text` |
| `timeline_source_provenance.ndjson` | `(record_id, source_row_ordinal, source_column_ordinal, source_identity_sha256)` | `record_id`, `source_identity_sha256`, `source_row_ordinal`, `source_column_ordinal`, `source_kind`, `source_metadata`, `source_header`, `raw_value`, `created_at` | required members plus `cell_kind` |

The v2 Timeline record envelope fields `row_version`, `created_at`,
`updated_at`, `created_by_user_id`, and `updated_by_user_id` MUST come from the
same-record row in `data/records.ndjson`; they MUST NOT be duplicated or
independently resolved from `timeline_records.ndjson`.

**IMR-BUNDLE-005** — The v1 translator MUST preserve `incident_id`,
`record_id`, `row_version`, timestamps, attribution, capture/review/supersession
state, and Timeline values. It MUST convert each admitted
`raw_capture.import_columns` item into v2 provenance without discarding its
source identity, row ordinal, column ordinal, header, raw value, or cell kind.
Malformed or non-representable legacy capture MUST fail; lossy success is
forbidden.

**IMR-BUNDLE-006** — Each admitted version MUST have one closed required core
path registry. An unknown or duplicate core member under `data/`, a missing
required member, or a required member without exactly one declared consumer or
validator MUST fail. Optional extensibility MUST use an owner-admitted
`ext/**` path and MUST NOT add an implicit core-file fallback.

**IMR-BUNDLE-007** — Structured export MUST retain the Core deterministic JSON,
row-ordering, UTF-8, checksum, path, member-type, extracted-byte,
compression-ratio, and member-count rules. Import MUST validate these boundaries
before source-owner execution.

**IMR-BUNDLE-008** — Version `1` import MUST remain available until all of the
following are true: two stable releases have exported v2; 180 days have elapsed
since the first such adopted stable release; successful v1-import telemetry is
zero for 30 consecutive days; operator inventory confirms no required v1
archive remains; and a later Core revision explicitly removes v1. The existing
projection date MUST NOT backdate the first adopted stable release.

**IMR-BUNDLE-009** — AG-001 MUST place current/legacy version behavior,
translation guarantees, version selection, failure behavior, and retirement
conditions in Core 01. The exhaustive file/row registry MAY live in an
explicitly normative Core 01 appendix. AG-002 MUST place binary conformance
criteria in Core 04. Machine schemas and compatibility data under `contracts/`
remain downstream projections and MUST be updated only through AG-003.

### 6.2 Job composition, dispatch, and recovery contract

**IMR-JOBS-001** — A claimed Incident Portability profile requires one non-nil
Jobs manager, one configured dequeue-gated runner, and one successfully
registered named Incident Bundle handler before its routes, listeners, or
readiness are published.

| Profile state | Runner state | Routes/listeners/readiness | Dispatch and recovery result |
| --- | --- | --- | --- |
| claimed | configured, gated, named handler registered | MAY publish after all other mandatory gates pass | Dispatch and recovery use the named handler. |
| claimed | nil, unconfigured, closed before publication, missing manager, or handler registration/recovery unavailable | MUST fail application assembly before publication | No work executes; durable queued work remains recoverable. |
| unclaimed | absent or present | Incident Bundle routes MUST be absent; no profile-specific runner requirement applies | No Incident Bundle handler is registered or invoked. |
| claimed after publication | named dispatch returns an error or the runner becomes unavailable | Readiness/admission MUST close according to the Jobs lifecycle owner | The job remains queued or recoverable; no fallback execution occurs. |

**IMR-JOBS-002** — Handler registration MUST precede recovery. Recovery MUST
use the same named handler used for new jobs and MUST remain blocked by the
dequeue gate until application publication activates that gate.

**IMR-JOBS-003** — New and recovered jobs MUST execute only through named
`DispatchJobID` behavior. Anonymous runner work, inline execution,
`context.Background()` goroutines, nil-receiver dispatch, and any equivalent
unmanaged fallback are forbidden.

**IMR-JOBS-004** — A dispatch or recovery error MUST NOT convert a durable
queued or recoverable job into success, failure, cancellation, or an
unrecoverable state. The owning Jobs lifecycle MUST preserve it for an eligible
process.

**IMR-JOBS-005** — Duplicate handler registration MUST fail assembly unless the
Jobs owner exposes an explicit idempotent registration identity and proves the
existing registration is the same handler contract.

**IMR-JOBS-006** — A restart after a committed import, export, or terminal
transition MUST recover to exactly one terminal result and MUST NOT duplicate
an incident, export descriptor, publication record, membership, audit event, or
terminal job result.

**IMR-JOBS-007** — The internal runner boundary MUST expose only named handler
registration, named recovery, and named job-ID dispatch to Incident Bundles.
The concrete Go shape MAY differ, but it MUST be behaviorally equivalent to:

| Operation | Input | Success | Failure |
| --- | --- | --- | --- |
| `RegisterHandler` | stable handler name and handler | handler is registered exactly once | assembly fails |
| `RecoverHandler` | context and stable handler name | eligible durable jobs are scheduled behind the gate | startup/recovery fails without executing fallback work |
| `DispatchJobID` | stable handler name and job UUID | durable job is admitted to managed execution | durable job remains queued or recoverable |

**IMR-JOBS-008** — Core 01 MUST own the claimed-profile composition, admission,
recovery, and component-loss behavior through AG-001. Core 04 MUST own the
startup, gate, restart, and no-fallback acceptance criteria through AG-002.

### 6.3 Typed source-owner portability contract

**IMR-PORT-001** — Application composition MUST assemble one closed ordered
catalog of source-owner ports. Incident Bundles owns catalog validation and
generic coordination. Each source owner constructs its port and retains its
row shape, fixed persistence, normalization, and semantic invariants.

**IMR-PORT-002** — Every port MUST satisfy the following language-neutral
interface. A prepared value is opaque to the coordinator, is bound to the port
and operation that created it, and MUST be passed only to that same port.

| Operation | Input | Output | Required behavior |
| --- | --- | --- | --- |
| `Descriptor` | none | immutable source descriptor | Declares family ID, contract major, exact paths and content roles, stable row identities, dependency IDs, owner ID, owner relation IDs, and closed invariant IDs. |
| `Export` | read-only query capability and `incident_id` | deterministic ordered files for the descriptor paths | Reads only owner state and emits the admitted current version. |
| `PrepareImport` | bounded read-only bundle capability and immutable import context | opaque prepared source or typed failure | Decodes exact shapes, normalizes, validates row-local semantics, and performs no database or visible-object mutation. |
| `ApplyImportTx` | transaction, matching prepared source, immutable import context | success or typed failure | Uses fixed owner SQL/SQLC and writes only owner relations in the supplied transaction. |
| `ValidateImportTx` | transaction and immutable import context | success or typed failure | Proves aggregate, cross-row, and declared cross-family invariants before publication. |

`owner relation IDs` are audit/ownership metadata. The coordinator MUST NOT
construct SQL, relation names, casts, or conflict policies from them.

**IMR-PORT-003** — The catalog MUST reject duplicate family IDs, duplicate
logical paths, uncovered required paths, a required path claimed more than
once, unsupported contract majors, unknown dependencies, dependency cycles,
missing ports, empty stable identities, empty invariant sets, and descriptors
whose owner relation IDs are absent from the schema-ownership projection.

**IMR-PORT-004** — The catalog MUST record one deterministic FK-safe current
order. Future families MUST declare dependency IDs; topological peers MUST sort
by stable `family_id` ascending. Catalog order MUST NOT come from Go import
order, map iteration, filesystem order, archive order, or physical table names.

**IMR-PORT-005** — A generic importer MUST NOT interpolate a descriptor-supplied
target relation and MUST NOT use `ON CONFLICT DO NOTHING`. Duplicate stable
identities and row-count mismatches are semantic failures, never compatibility
success.

**IMR-PORT-006** — The current owner ports MUST enforce at least the following
invariant classes. AG-001 MUST adopt the invariant ownership and publication
boundary; the exact closed invariant IDs MAY be projected through a normative
Core appendix and typed contract.

| Source owner | Required import invariants |
| --- | --- |
| Records | Every row belongs to the imported incident; record type, row version, timestamps, actor attribution, and deletion tuples are legal; every subtype-required envelope has exactly its admitted owner row. |
| Timeline | Version-selected exact shape; same-incident `timeline_event` envelope; coherent capture, review, supersession, generated-time, and paired-time state; provenance identities are unique and non-orphaned; v1 translation is lossless. |
| Parties | Same-incident party envelope; required identity fields and lifecycle state are valid; normalized string/reference pairs equal the owner normalization result. |
| Entities | Mentions remain observations; hosts and identities have the correct envelopes; resolution and merge-lineage tuples are coherent; aliases and preserved identifiers are normalized, classified, unique, and same-incident. |
| Indicators | Type/value/hash representation and normalization are legal; duplicate identities are rejected; observations and state intervals are same-incident, ordered, and coherent; repeated observations are not silently merged. |
| Artifacts | Every artifact has the correct envelope and exactly the admitted subtype; subtype lifecycle and required fields are legal; handoff-risk references target handoffs; all references are same-incident. |
| Tasks/Decisions | Envelope type and lifecycle are legal; owner/completion/decision dependent fields form admitted tuples; referenced records belong to the incident. |
| Evidence | Evidence envelope, object metadata, storage reference, byte size, digest, and lifecycle agree; staged bytes match the declared digest; custody events are ordered and reference same-incident evidence/records. |
| Assessments | The subject is a same-incident host or identity of the admitted type; state, confidence, rationale, timestamps, and lifecycle form legal tuples. |
| Links/Tags | Link endpoints are valid same-incident records; type, direction, field key, uniqueness, and deletion tuples are legal; tags are normalized; `tags.ndjson` exactly equals the distinct `(tag_name, normalized_tag_name)` catalog derived from imported record tags. |
| Revisions | Referenced change sets, mutations, revisions, records, and actors exist; mutation sequence is contiguous; `(record_id, row_version)` is unique; before/after history reconstructs imported current state; sequence repair runs only after validation. |
| Saved Views | UUIDs, incident/schema references, scope/owner tuple, display name, query, layout, version, and timestamps are valid; absent optional Reference Packs degrade only admitted overlays. |

**IMR-PORT-007** — Required core files outside the source-family table MUST have
the following exact dispositions.

| Logical input | Consumer | Required validation and effect |
| --- | --- | --- |
| `data/incident.json` | Incidents source-owner port | Exact shape, incident identity/key/lifecycle, attribution, and version tuples; creates only the unpublished incident source row. |
| `data/actors.ndjson` | Incident Bundles attribution adapter with Revisions semantics | Every referenced source actor has exactly one unique descriptor; descriptors are inert and MUST NOT create login, provider binding, deployment role, incident membership, or session state. Missing or malformed source actor IDs fail rather than being skipped. |
| `data/reference_pack_refs.json` | Incident Bundles coordinator with Reference Pack owner contract | Exact closed shape and reference identity; missing optional packs degrade only admitted overlays and MUST NOT change authoritative incident state. |
| admitted `ext/**` member | Matching claimed extension participant | Exact profile ID, contract major, payload schema, digest, resource bounds, and participant admission; an unknown, unclaimed, or mismatched participant is never invoked. |

**IMR-PORT-008** — All database application, aggregate validation, revision
repair, attribution flush, projection rebuild, initial-admin creation, audit
publication, and incident publication MUST share one final transaction. A
source port MUST NOT commit, start a nested independent transaction, or publish
visible state.

**IMR-PORT-009** — Evidence and participant bytes prepared before the final
transaction MUST remain non-visible. Failure MUST abandon them or retain them
only in an owner-admitted non-visible quarantine. Cleanup MUST be retry-safe and
MUST NOT delete already committed final objects.

**IMR-PORT-010** — Adding a required source family requires one adopted owner
contract, one registered port, exact path accounting, stable identities,
dependencies, invariant IDs, valid/invalid fixtures, rollback evidence, and
catalog/projection regeneration. Partial registration is forbidden.

**IMR-PORT-011** — Database constraints remain defense in depth. A successful
type cast, foreign-key check, uniqueness check, or insert is not evidence that
the owner semantic invariants passed.

**IMR-PORT-012** — `incidentportability` MAY retain bounded NDJSON
encode/decode, canonical JSON, safe value extraction, and actor-ID remap
utilities. It MUST NOT own source-family persistence, physical relation
selection, owner normalization, aggregate validation, or publication.

### 6.4 Import coordination and failure contract

**IMR-SEC-001** — The import coordinator MUST execute the following algorithm
in order:

1. Validate the complete application-composed catalog and exact path
   accounting.
2. Verify the archive, manifest, admitted version, checksums, resource limits,
   member types, and closed core paths.
3. Select the codec exclusively from `manifest.bundle_version`.
4. Invoke `PrepareImport` for every source family without visible mutation.
5. Stage evidence and participant bytes under non-visible logical references.
6. Begin the final database transaction.
7. Apply `incident.json`, the inert actor catalog, and source-owner ports in the
   catalog's recorded FK-safe order.
8. Invoke every applied port's `ValidateImportTx`.
9. Repair revision sequences, flush attribution, rebuild projections, and
   finalize initial administration, audit, and publication.
10. Commit exactly once. Only that commit makes the incident and final object
    references visible.
11. On any error or cancellation before commit, roll back, abandon or
    quarantine staged bytes, and retain no partial success.

**IMR-SEC-002** — A source-family verification failure MUST use
`incident_bundle_import_rejected` with `reason_code='source_family_invalid'`.
Details MUST contain exactly the closed `source_family_id` and `invariant_id`
when those identifiers can be reported safely. Unknown or unadmitted bundle
majors MUST use `reason_code='unsupported_bundle_version'`.

| Failure class | Public code | Public details | Retryable |
| --- | --- | --- | --- |
| unadmitted bundle major | `incident_bundle_import_rejected` | `reason_code='unsupported_bundle_version'` | `false` |
| owner row or aggregate invariant failure | `incident_bundle_import_rejected` | `reason_code='source_family_invalid'`, `source_family_id`, `invariant_id` | `false` |
| malformed admitted-version structure | `incident_bundle_import_rejected` | existing safe `malformed_manifest` reason | `false` |
| transient internal coordination/storage failure | existing owner-defined internal failure | no source contents or internal topology | owner-defined |

**IMR-SEC-003** — Public errors, logs, telemetry, readiness, and administrative
summaries MUST NOT contain raw imported row values, raw evidence or extension
bytes, SQL text, relation names, credentials, provider subjects, object keys,
staging IDs, host-absolute paths, or cryptographic key material.

**IMR-SEC-004** — Failure before commit MUST leave no visible incident,
membership, workbook preference, successful administrative audit, projection,
export/import success descriptor, terminal-success result, or final object
reference. A later retry MUST encounter only durable job state and admitted
non-visible staging/quarantine state.

**IMR-SEC-005** — Source-family failure behavior MUST be independent of archive
row order and map iteration. Given the same admitted bundle and deployment
state, two conforming implementations MUST select the same codec, port order,
invariant result, public reason, and visibility outcome.

**IMR-SEC-006** — The coordinator MUST NOT authorize from bundle contents.
Import submission, final deployment-admin recheck, initial incident
administration, current polling/cancellation authorization, and extension claim
admission remain server-derived under their existing owners.

### 6.5 Non-normative decision support

The following reports support the chosen boundaries but do not own them:

- [R01](../research/R01-aurora_incident_response_report.md) documents fragile
  partially implemented storage-format migration and weak import validation.
- [R03](../research/R03-Kanvas_technical_research_report.md) documents the
  portability and silent-drift risk of a workbook format without an explicit
  migration layer.
- [R04](../research/R04-responsive_browser_spreadsheet_ui_research_memo.md)
  supports separating typed canonical state from denormalized projections.
- [R06](../research/R06-spreadsheet_of_doom_dfir_research_report.md) and
  [R07](../research/R07-spreadsheet-of-doom-sod-report.cr.md) support stable
  identities, provenance, typed ingestion, transactions, rollback, explicit
  schema/versioning, and portable exports.
- [R08](../research/R08-handsontable-react-research-report.md) supplies only an
  analogy for registry-driven adapters and curated public/internal boundaries.
- [R02](../research/R02-cartulary_crm_tem_dfir_research_report.md),
  [R05](../research/R05-responsive-interface-design-report.cr.md), and
  [R09](../research/R09-react-data-grid-research-report.md) were reviewed and
  do not materially determine bundle wire compatibility, Jobs composition, or
  source-owner import authority.

The primary external references were checked for freshness as of `2026-07-26`:
[NIST SP 800-61 Rev. 3](https://csrc.nist.gov/pubs/sp/800/61/r3/final) is the
current final incident-response publication;
[RFC 8785](https://www.rfc-editor.org/info/rfc8785/) remains the canonical-JSON
reference; MITRE CWE 4.20 covers
[path traversal](https://cwe.mitre.org/data/definitions/22.html) and
[compressed-data amplification](https://cwe.mitre.org/data/definitions/409.html);
and PostgreSQL 18 documents current
[constraint](https://www.postgresql.org/docs/current/ddl-constraints.html) and
[rollback](https://www.postgresql.org/docs/18/sql-rollback.html) behavior. These
references support safety and interoperability choices only. Adopted repository
owners remain authoritative.

## 7. Refactor Workstreams

These workstreams describe planning completion separately from later
implementation. `DONE` below means the planning artifact is complete, not that
production work has occurred.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01, WF-02 | Fix repository snapshot, scope, authority, and allowed write. | Tracker only for this task. | `make lint-markdown` after authoring. | Scope and owner sources recorded; planning status `DONE`. |
| WF-01 | Target inventory | chain | WF-00 | WF-03, WF-04 | Inventory all 56 files and live callers/dependencies. | Three target directories and adjacent callers. | Read-only `rg`, `sed`, `jq`, `wc`. | Section 2 complete; planning status `DONE`. |
| WF-02 | Contract-owner mapping | parallel | WF-00 | WF-03, WF-05 | Map HTTP, WS, storage, auth, projections, generated, UI, and harness contracts. | Core/NLSpecs, OpenAPI owners, compatibility and verification contracts. | Read-only contract/catalog inspection. | Section 4 complete; planning status `DONE`. |
| WF-03 | Characterization test gap analysis | chain | WF-01, WF-02 | WF-05, WF-07 | Identify covered behavior and gaps before movement. | Target tests, owner family manifests, collaborator tests. | `make explain-test-owner` and `make target-plan-json` discovery. | IMR acceptance gaps and required tests recorded; planning status `DONE`. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Classify concrete imports, SQL, test support, auth, projection, and hook seams. | Target code, backend boundary and test-support inputs. | Read-only import/source scan. | Section 5 complete; planning status `DONE`. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Allocate permanent owners without inventing one combined module. | Incident Bundles, Incidents, owner adapters, app composition. | Owner review required before implementation. | Section 3 direction complete; planning status `DONE`. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order the smallest safe behavior-preserving slices. | Future files named in Section 8. | Per-slice commands in Section 8. | Dependencies/rollback/exit criteria complete; planning status `DONE`. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Reconcile boundary rules, support ownership, selectors, and generated projections only when needed. | Authored `tools/` inputs and affected tests; generated outputs via Make only. | Boundary, JSON-shape, drift, owner slices. | No phase map treated as runtime architecture; planning status `DONE`. |
| WF-08 | Validation and final handoff | chain | WF-01 through WF-07 | none | Record focused-to-broad validation and continuation state. | Tracker now; future implementation tree later. | Section 9 plus `make lint-markdown` now. | Seven handoff tables current; planning status `DONE`. |

## 8. Refactor Slice Plan and Status

Completed slices record the final AG-001 through AG-003 execution state.
Remaining slices retain their original boundaries and require a separately
authorized task.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | none | Adopt the IMR-BUNDLE, IMR-JOBS, IMR-PORT, and IMR-SEC observable-behavior clauses into Core 01 and their binary acceptance criteria into Core 04. | Core 01 and Core 04 owner sections; normative appendices. | Bundle major and paths, public failure reasons, claimed-profile composition, owner validation, atomic visibility, v1 retention. | Core requirement/acceptance traceability; owner review of every mapping row and default. | `make lint-markdown`; applicable owner/document-set checks. | Revert Core amendments together if owner review rejects the package. | AG-001 and AG-002 are `DONE`, every adopted clause has an AC, and no adopted owner contradicts the disposition. | `DONE` |
| SL-01 | SL-00 | Add exact characterization for deterministic v2 export, v1-to-v2 translation, unknown/mixed versions, Timeline provenance, nil/unconfigured runner rejection, named recovery, semantic-invalidity, duplicate identity, and every remaining atomic failure boundary. | Incident Bundle and source-owner tests; Jobs/app test support; verification owner inputs. | A test MUST derive from an adopted requirement and MUST NOT invent behavior. | Preserve all existing owner rows; add SQL-valid semantic-invalidity fixtures, special-file cases, dispatch/restart recovery, and rollback/non-visibility cases from Section 13. | Both `make test-slice` and `make service-backed-test-slice` owner commands after `make explain-test-owner`. | Revert an invalid test assumption rather than altering production away from the adopted owner; retain failing evidence for a genuine implementation gap. | Every IMR-AC row assigned to SL-01 has selected executable evidence and fails against any known nonconforming path. | `DONE` |
| SL-02 | none | Add the two Incidents production-import rules to authored backend boundary policy, prove parity, then delete the redundant unaccounted AST scanner. | `tools/backend_module_boundaries.json`, boundary checker tests, `incidents/boundary_guard_test.go`; generated inputs only if owner manifests require them. | Losing protection against Workbook startup or platform WS imports. | Boundary checker positive/negative fixtures; existing route-port unit test. | `make backend-module-boundary-check`, `make harness-contract`, `make json-shape-check`, `make generate-drift`. | Restore the test until machine-policy parity passes. | Machine policy rejects both imports and no unaccounted duplicate scanner remains. | `TODO` |
| SL-03 | none | Move only the owner-neutral runtime wrapper to shared application test support; retain incident actions, route inventories, mutation helpers, and Store fixtures with Incidents. | Incident `testsupport/scenariotest/harness.go`, `internal/testutil/appsupport`, cross-owner test imports, `tools/test_support_inventory.json`. | Test fixture mode, database/object-store lifecycle, route-mode guards, harness accounting. | Preserve every selected owner/collaborator row and exact fixture policy. | Affected owner slices, `make json-shape-check`, `make harness-contract`, `make test-fast`. | Retain a temporary forwarding helper for one slice only if all callers cannot move atomically; delete it before completion. | Cross-owner runtime composition imports shared appsupport directly; incident-specific helpers remain owner-local; inventory rationale matches live use. | `TODO` |
| SL-04 | SL-00, SL-01 | Implement the Section 6.3 typed catalog; migrate each owner to deterministic export, non-mutating preparation, fixed transaction writes, and aggregate validation; add explicit consumers for special files; then remove the generic relation importer and central invariant prose. | Incident Bundles coordinator, Incident Portability utilities, every source-owner portability adapter, application/revision assembly, authored catalog projections. | Exact path coverage/order, source identities, attribution, history, blobs, saved views, failure details, transaction rollback, visibility. | Per-owner valid/invalid/duplicate/orphan tests; complete v1/v2 round trips; exact path accounting; special-file validation; failure injection; history/projection reconstruction. | Incident Bundle and collaborator owner slices, `make backend-module-boundary-check`, drift checks, `make test-fast`, `make check`. | Migrate one source family behind the stable port interface at a time; never leave both paths active for one import. | Every required path has exactly one consumer/validator, every owner executes fixed persistence and invariant checks, no dynamic relation SQL or generic conflict-ignore remains, and atomic publication evidence passes. | `DONE` |
| SL-05 | none | Split Incidents application coordination from private persistence while preserving `Access`, routes, preferences, reporting provider, portability adapter, transaction participant, and import finalizer. | Incidents root package and private repository/service files; app/test composition. | Transactions, idempotency, audit, membership versions, last-admin, lifecycle, pagination, preferences, collaboration notifications. | All incident unit/store/integration/frontend/browser rows; no new public Go compatibility promise. | Both Incidents owner slices, `make backend-module-boundary-check`, `make test-fast`; browser gate if route behavior is touched. | Move one operation family at a time; keep a temporary internal facade only until repository callers migrate. | Routes depend on one cohesive application service, persistence is private, peers retain only Access/declared ports, and observable behavior is identical. | `TODO` |
| SL-06 | SL-00, SL-01 | Require claimed composition to inject a configured named runner; make registration/recovery failure fatal before publication; remove every fallback; replace production test hooks with Jobs/application-test seams. | Bundle worker/routes/tests, Incidents fault tests, Jobs runner and app test support. | Authorization-race timing, rollback injection, gate activation, durable recovery, restart exact-once behavior, production test-route rejection. | Preserve final-publication auth and create rollback; add claimed/unclaimed composition, nil/unconfigured runner, named dispatch/recovery, pre-activation gate, dispatch failure, component loss, and restart cases. | Focused owner slices, `make backend-module-boundary-check`, `make test-fast`. | Never retain two fault paths or any unmanaged execution fallback. | All Incident Bundle work uses the named runner, invalid claimed composition fails before publication, durable recovery evidence passes, production modules expose no test-only override API, and prior failure/race evidence remains selected. | `DONE` |
| SL-07 | AG-001, AG-002, and applicable implementation slices | Align authored compatibility, error, source-catalog, verification, test-support, boundary, and OpenAPI inputs with adopted owners and actual moved identities; regenerate every downstream artifact through Make. | `tools/` and `contracts/` owner inputs; generated roots only via generator. | Version/error drift, row loss/duplication, generated drift, public schema drift, support ownership. | Exact version/path/error/catalog selectors, requirement-to-row traceability, and boundary/harness contract tests. | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, owner slices. | Revert each authored input and its regenerated projection together if validation fails. | AG-003 is `DONE`, every active row is accounted exactly once, generated artifacts match adopted owners, and no generated file was hand-edited. | `DONE` |
| SL-08 | SL-02 through SL-07 | Run focused-to-broad validation, update tracker evidence, and close the overall refactor. | Whole affected tree and this tracker. | False completion from partial/stale evidence. | No new tests; execute the selected catalog and gates. | Owner slices, `make test-fast`, `make agent-finalize`, risk-appropriate `make check`. | Roll back only the failing slice, not unrelated completed slices; retain failure artifact/run root. | All remaining slices pass and all seven handoff tables are current. | `TODO` |

## 9. Validation Plan

Discovery commands establish command availability only; execution success is
recorded separately below. Focused owner and row slices ran before the broad
final gates.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.incidentbundles` | 22 Incident Bundle owner rows, including 11 unit rows where applicable to the slice planner. | yes | Use exact `ROWS=` only after `make explain-test-owner`; omission selects the owner plan. |
| unit | `make test-slice OWNER=module.incidents` | 40 Incidents owner rows, including Go, Vitest, and Playwright identities selected by the owner plan. | yes | Characterization baseline before Incidents movement. |
| integration | `make service-backed-test-slice OWNER=module.incidentbundles` | 11 service-backed Incident Bundle rows. | yes | Covers real export/import, authorization, storage, projection, and transaction behavior. |
| integration | `make service-backed-test-slice OWNER=module.incidents` | 25 service-backed Incidents rows. | yes | Includes integration and browser-backed rows according to catalog profiles. |
| e2e/browser | `make browser-e2e-webserver-backed` | Broader shared-stack browser behavior. | no | A dedicated run was not selected because no frontend behavior changed; browser-backed rows selected by `make check` passed. |
| generated drift | `make generate-drift` | Generated artifacts against authored owners. | no | Required for future owner/codegen changes. |
| generated drift | `make generated-artifact-policy-check` | Generated markers and lint-scope policy. | no | Never hand-edit generated roots. |
| generated drift | `make json-shape-check` | Authored JSON owner/schema shape. | no | Required for boundary, catalog, support-inventory, or contract input changes. |
| import-boundary/static | `make backend-module-boundary-check` | Backend module import/source policy. | yes | Required before and after SL-02/SL-04/SL-05/SL-06. |
| import-boundary/static | `make frontend-import-boundary-check` | Frontend package ownership. | no | Run only if a later authorized task touches frontend files; none are planned directly. |
| docs | `make lint-markdown` | Authored Markdown, including this tracker. | yes | The draft pass succeeded at `.cartulary/test-results/20260727T232509Z-p1796003`; a final rerun follows the last handoff edit. |
| diff hygiene | `git diff --check` | Whitespace and patch integrity for the tracker. | yes | The draft pass produced no output; a final rerun follows the last handoff edit. |
| decision closure | targeted `rg` scan | Obsolete unresolved-decision language in decision-bearing sections. | yes | `Open Questions`, `no behavior inferred`, and equivalent active wording MUST be absent; historical handoff rows MAY describe the prior session. |
| scope | `git status --short` | Repository changes after the tracker revision. | yes | Confirms the authorized owner, implementation, contract, test, generated, and handoff changes; no dependency lockfile or unrelated user change was overwritten. |
| focused broad | `make test-fast` | Backend/frontend focused local verification. | yes | Passed 865 tests in the final executable source state. |
| finalizer | `make agent-finalize` | Harness-maintenance artifacts before broad verification. | yes | Passed; retained-run maintenance was skipped because `RESULTS_DIR` was unset. |
| full check | `make check` | Full developer gate. | yes | Passed 177 of 177 work units and 721 tests in the final executable source state. |

Command discovery performed for this plan:

- `make task-guide ROLE=module-author OWNER=<owner>` and
  `make explain-test-owner OWNER=<owner>` passed for
  `module.incidentbundles`, `app.server`, `module.extensions`,
  `module.incidents`, `module.records`, `module.timeline`, `module.parties`,
  `module.entities`, `module.indicators`, `module.artifacts`,
  `module.tasksdecisions`, `module.evidence`, `module.assessments`,
  `module.links`, `module.revisions`, `module.savedviews`,
  `module.reference_data`, `module.networkflow`, and `platform.telemetry`.
- `make explain-target TARGET=<target> DETAIL=summary`
- `make target-plan-json TARGET=backend-unit`
- `make help-all`

### AG-001 through AG-003 execution evidence

All final broad results below came from the same executable source state. The
later handoff-only edits do not affect product or generated artifacts.

| Command | Result | Run root | Summary artifact |
| --- | --- | --- | --- |
| `make format` | PASS | `.cartulary/test-results/20260727T231028Z-p1621934` | `format/tool-run-summary.json` |
| `make generate` | PASS | `.cartulary/test-results/20260727T231036Z-p1624765` | `generate/tool-run-summary.json` |
| Final Incident Bundle safe-failure row slice | PASS | `.cartulary/test-results/20260727T231052Z-p1627272` | `test-slice/tool-run-summary.json` |
| `make json-shape-check` | PASS | `.cartulary/test-results/20260727T231052Z-p1627249` | `json-shape-check/tool-run-summary.json` |
| `make backend-module-boundary-check` | PASS | `.cartulary/test-results/20260727T231052Z-p1627527` | `backend-module-boundary-check/tool-run-summary.json` |
| `make harness-contract` | PASS | `.cartulary/test-results/20260727T231052Z-p1627557` | `harness-contract/tool-run-summary.json` |
| `make generate-drift` | PASS | `.cartulary/test-results/20260727T231125Z-p1631438` | `generate-drift/tool-run-summary.json` |
| `make generated-artifact-policy-check` | PASS | `.cartulary/test-results/20260727T231125Z-p1631463` | `generated-artifact-policy-check/tool-run-summary.json` |
| `make test-fast` | PASS, 865 tests | `.cartulary/test-results/20260727T231150Z-p1635697` | `tool-run-summary.json` |
| `make agent-finalize` | PASS; generated state unchanged; retained-run maintenance skipped because `RESULTS_DIR` was unset | `.cartulary/test-results/20260727T231427Z-p1689884` | `agent-finalize/tool-run-summary.json`, `agent-finalize/finalize-summary.json` |
| `make check` | PASS, 177 of 177 units and 721 tests | `.cartulary/test-results/20260727T231453Z-p1696375` | `tool-run-summary.json` |

Focused service-backed owner evidence:

| Owner or exact slice | Result | Run root | Summary artifact |
| --- | --- | --- | --- |
| Incident Bundles full service-backed owner, 12 tests | PASS | `.cartulary/test-results/20260727T223522Z-p1117717` | `service-backed-test-slice/tool-run-summary.json` |
| Incident Bundles v2 round trip and v1 translation | PASS | `.cartulary/test-results/20260727T223435Z-p1113447` | `service-backed-test-slice/tool-run-summary.json` |
| Incidents portability rows | PASS | `.cartulary/test-results/20260727T225905Z-p1445416` | `service-backed-test-slice/tool-run-summary.json` |
| Artifacts and Assessments | PASS | `.cartulary/test-results/20260727T225838Z-p1438873`, `.cartulary/test-results/20260727T225838Z-p1438889` | `service-backed-test-slice/tool-run-summary.json` |
| Entities and Evidence | PASS | `.cartulary/test-results/20260727T225839Z-p1438907`, `.cartulary/test-results/20260727T225839Z-p1438929` | `service-backed-test-slice/tool-run-summary.json` |
| Indicators, Links, and Parties | PASS | `.cartulary/test-results/20260727T225905Z-p1445432`, `.cartulary/test-results/20260727T225905Z-p1445448`, `.cartulary/test-results/20260727T225905Z-p1445472` | `service-backed-test-slice/tool-run-summary.json` |
| Records, Revisions, Saved Views, and Tasks/Decisions | PASS | `.cartulary/test-results/20260727T225933Z-p1452431`, `.cartulary/test-results/20260727T225933Z-p1452449`, `.cartulary/test-results/20260727T225933Z-p1452462`, `.cartulary/test-results/20260727T225933Z-p1452481` | `service-backed-test-slice/tool-run-summary.json` |
| Timeline | PASS | `.cartulary/test-results/20260727T225956Z-p1458710` | `service-backed-test-slice/tool-run-summary.json` |
| Extensions full owner after named-Jobs alignment | PASS | `.cartulary/test-results/20260727T223736Z-p1138731` | `service-backed-test-slice/tool-run-summary.json` |
| Clean full service-backed suite | PASS | `.cartulary/test-results/20260727T225110Z-p1349118` | `test-service-backed/tool-run-summary.json` |
| Duration baseline generation and coverage | PASS | `.cartulary/test-results/20260727T225520Z-p1424736`, `.cartulary/test-results/20260727T225714Z-p1430134` | `go-test-duration-baselines/tool-run-summary.json`, `go-test-duration-baseline-coverage/tool-run-summary.json` |

Focused unit and composition evidence:

| Owner or exact slice | Result | Run root | Summary artifact |
| --- | --- | --- | --- |
| Incident Bundles full unit owner | PASS | `.cartulary/test-results/20260727T230019Z-p1461430` | `test-slice/tool-run-summary.json` |
| Records and Timeline unit owners | PASS | `.cartulary/test-results/20260727T225956Z-p1458692`, `.cartulary/test-results/20260727T225956Z-p1458685` | `test-slice/tool-run-summary.json` |
| Reference Data special-file validation | PASS | `.cartulary/test-results/20260727T223150Z-p1078581` | `test-slice/tool-run-summary.json` |
| Network Flow extension participant owner, 55 tests | PASS | `.cartulary/test-results/20260727T223208Z-p1081318` | `test-slice/tool-run-summary.json` |
| Telemetry owner | PASS | `.cartulary/test-results/20260727T223208Z-p1081317` | `test-slice/tool-run-summary.json` |
| Server claimed/unclaimed composition row | PASS | `.cartulary/test-results/20260727T223119Z-p1075789` | `test-slice/tool-run-summary.json` |
| Extensions named Jobs row | PASS | `.cartulary/test-results/20260727T224602Z-p1265665` | `service-backed-test-slice/tool-run-summary.json` |

Related failures encountered during alignment were repaired before the final
same-state runs:

| Failure | Run root | Disposition |
| --- | --- | --- |
| OpenAPI review gate rejected an unregistered contract change | `.cartulary/test-results/20260727T215200Z-p973940` | Added the authored 2.0.0 release change set and regenerated. |
| Harness row/cardinality and baseline mismatches | `.cartulary/test-results/20260727T215024Z-p972217`, `.cartulary/test-results/20260727T215423Z-p978613`, `.cartulary/test-results/20260727T222924Z-p1073384` | Reconciled authored test families, owner catalog, topology, and duration baseline. |
| Backend boundary failures | `.cartulary/test-results/20260727T214915Z-p967131`, `.cartulary/test-results/20260727T225557Z-p1427582` | Repaired Revisions fixed-SQL ownership and added the two new assembly/provider import allowances. |
| Incident Bundle compile, expectation, and v1 fixture failures | `.cartulary/test-results/20260727T220056Z-p1000121`, `.cartulary/test-results/20260727T220356Z-p1014257`, `.cartulary/test-results/20260727T221316Z-p1038211`, `.cartulary/test-results/20260727T221340Z-p1039638` | Aligned the new catalog, safe failures, and retained-v1 test data. |
| Reference Data, Extensions, and Saved Views focused failures | `.cartulary/test-results/20260727T223119Z-p1075823`, `.cartulary/test-results/20260727T223600Z-p1120680`, `.cartulary/test-results/20260727T224429Z-p1258799` | Closed trailing-document validation, updated extension verification accounting, and updated the Saved Views validator signature. |
| Initial full service-backed failures | `.cartulary/test-results/20260727T223853Z-p1156835` | Fixed Jobs transaction fixture selection and the Saved Views test signature; the clean rerun passed. |
| Partial or contaminated duration-baseline evidence rejected | `.cartulary/test-results/20260727T223841Z-p1156466`, `.cartulary/test-results/20260727T225053Z-p1348737` | Generated the baseline only from the clean full service-backed root. |

Dedicated `make browser-e2e-visual` and
`make browser-e2e-measurement` runs were skipped because no frontend or visual
behavior changed. Browser-backed rows selected by `make check` still passed.
`make migration-drift` was skipped because no migration changed.

## 10. Top-Level Work Tracker

Only the status values `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`,
and `DROPPED` are valid.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Establish the three-package planning seam and safe label. | WF-00 | DONE | none | Section 1. | Paths, label, allowed write, non-goals, and later authorization are explicit. |
| T-002 | Inventory all 56 live target files. | WF-01 | DONE | T-001 | Section 2: 16 Bundle, 3 Portability, 37 Incidents rows. | Every file has responsibility, surface, callers, dependencies, tests, owner, and risk. |
| T-003 | Map owner documents and observable contracts. | WF-02 | DONE | T-001 | Sections 1 and 4. | Every discovered contract risk has an owner and test posture. |
| T-004 | Resolve the bundle v1/v2 planning disposition. | WF-02 | DONE | T-003 | RB-001 and IMR-BUNDLE; Core 01, bundle code, compatibility projection. | The tracker requires v2 current export, retained v1 import, manifest-only codec selection, exact Timeline paths, and bounded retirement. |
| T-005 | Map characterization coverage and gaps. | WF-03 | DONE | T-002, T-003 | Owner manifests, Section 4, RB-002/RB-003, Section 13. | Existing evidence and every required missing acceptance row are explicit. |
| T-006 | Implement the portability owner-port/catalog redesign. | WF-05 | DONE | T-005, T-016, T-017 | IMR-PORT, IMR-SEC, SL-04; typed catalog and owner-port evidence. | Ordered typed port catalog and per-owner persistence/invariant behavior pass all mapped evidence. |
| T-007 | Plan Incidents application/persistence layering. | WF-05 | TODO | T-005 | SL-05. | Access/route behavior is stable and private persistence is separated. |
| T-008 | Correct owner-neutral scenario runtime support placement. | WF-07 | TODO | T-005 | SL-03; support inventory and live callers. | Shared runtime composition is in appsupport and incident-specific helpers remain owner-local. |
| T-009 | Project Incidents import rules into machine boundary policy. | WF-07 | TODO | T-005 | SL-02; boundary guard and policy input. | Both rules pass in machine policy and redundant scanner is retired. |
| T-010 | Characterize named-runner recovery and replace production test hooks. | WF-03/WF-05 | DONE | T-005, T-016, T-017 | IMR-JOBS and SL-06; Jobs/application-test seam evidence. | Claimed/unclaimed composition and durable recovery pass; owned Jobs/app test seams close prior evidence; production test hooks are gone. |
| T-011 | Preserve workbook preference and reporting provider ownership. | WF-05 | DONE | T-003 | Sections 3 and 5. | Tracker records intentional boundaries and no unsupported move. |
| T-012 | Preserve projection and collaboration dependency inversion. | WF-05 | DONE | T-003 | Sections 3–5. | Rebuild and Collaboration remain explicit ports. |
| T-013 | Plan harness/catalog/generated updates without runtime inference. | WF-07 | DONE | T-005, T-006, T-007 | SL-07 and Section 9. | Authored inputs and Make regeneration/validation are named. |
| T-014 | Obtain implementation authorization. | WF-08 | DONE | T-004 through T-013 | User request dated `2026-07-27`. | The user explicitly authorized AG-001 through AG-003 and their intrinsic slices. |
| T-015 | Complete tracker validation and handoff. | WF-08 | DONE | T-001 through T-013 | `make lint-markdown`, `git diff --check`, obsolete-phrase scan, Git status; Sections 11 and 13. | Tracker checks pass and no out-of-scope file changed. |
| T-016 | Adopt the resolved behavior in Core 01. | WF-02 | DONE | T-004 | AG-001; REQ-01-635 through REQ-01-643. | Core 01 owns every observable requirement and its normative registries. |
| T-017 | Adopt acceptance and verification authority in Core 04. | WF-03 | DONE | T-016 | AG-002; AC-487 through AC-507 and traceability projection. | Every observable Core 01 requirement has binary Core 04 acceptance coverage and verification ownership. |
| T-018 | Align projections, implementation, fixtures, and generated artifacts. | WF-07 | DONE | T-016, T-017, applicable implementation slices | AG-003, SL-04, SL-06, SL-07, and Section 12 evidence. | All downstream artifacts agree with adopted owners and the same-source-state verification plan passes. |

## 11. Session Handoff Log

The first-session rows below are preserved as historical evidence of the
tracker's initial state. The decision-resolution row records the later
tracker-only NLSpec revision. Archived trackers remain cited in Section 1
rather than copied as current evidence.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Three existing packages form a planning seam only; planning is complete and implementation remains unauthorized. | Inspected framework, Core 00–04, adopted relevant NLSpecs, Domain, all target files; touched this tracker only. | `sed`, `rg`, `git status`, `git rev-parse`, `date`, `sha256sum`. | Authority order and clean baseline recorded; no owner-owner contradiction found. | RB-001. | Obtain owner disposition before any bundle-format task. |
| 2026-07-27T16:32:44-04:00 | Codex decision-resolution documentation session | RB-001 through RB-003 are resolved for planning; observable changes remain gated on Core adoption and later implementation authorization. | Re-inspected NLSpec authoring guidance, Core 01/04 clauses, bundle/runner/source-port implementations, projections, tests, and R01–R09; touched this tracker only. | `sed`, `rg`, Git inspection, `make lint-markdown`, `git diff --check`, inventory/traceability scans. | Added IMR requirements, matrices, deterministic coordination, resolved dispositions, AG gates, aligned slices, and binary acceptance criteria; document checks pass. | AG-001, AG-002, AG-003. | Authorize SL-00 to adopt the behavior in Core 01/Core 04 before any gated downstream implementation. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | Core 01 and Core 04 own the adopted behavior; downstream state is aligned and verified. | Owner specifications, authored contracts, implementation, fixtures, generated projections, and this final handoff. | Owner discovery, focused owner slices, generation/policy checks, `make test-fast`, `make agent-finalize`, `make check`. | AG-001, AG-002, and AG-003 satisfy their exit conditions; no owner contradiction was found. | None for adopted gates. | Execute SL-02 next. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Incidents and Incident Bundles are legitimate owners with mixed internals; Incident Portability is shared support, not a third public owner. | All 56 target files, app server/workbook/reporting/projection composition; touched this tracker only. | `rg --files`, `rg` import/symbol/caller scans, `sed`, `wc -l`. | Keep/move/split/defer allocation and future slices recorded. | RB-001, RB-002, RB-003. | Start only an authorized characterization or boundary-policy slice. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | Incident Bundles consumes a typed sourceport catalog assembled by `internal/app/incidentportabilityassembly`; each source owner retains fixed persistence and invariants. | Incident Bundles, sourceport, assembly, 13 source owners, Incidents, Jobs, Reference Data, telemetry, and boundary policy. | Focused owner slices, `make backend-module-boundary-check`, `make test-fast`, `make check`. | SL-04 and SL-06 are complete; dynamic relation SQL, generic conflict-ignore, and production test hooks are removed. | SL-02, SL-03, SL-05 remain. | Project the two remaining Incidents import rules and remove the duplicate AST scanner in SL-02. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Frontend is a downstream contract consumer; no target code imports grid-vendor or shell/controller state. | Inspected IncidentAdminPanel, IncidentImportPanel, IncidentLanding, WorkbookShell, frontend owner rows; touched this tracker only. | `rg`, `sed`, `jq`. | Frontend movement deferred; HTTP/generated/UI selector risks frozen. | None for backend planning. | Run owner browser rows and broader browser gate only if later public/frontend risk is introduced. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | No frontend source or visual behavior changed; protocol TypeScript changed only through generation. | Generated protocol error artifact and broad selected browser-backed evidence. | `make generate`, `make generated-artifact-policy-check`, `make check`. | Generated consumer projection and selected browser-backed rows pass. Dedicated visual and measurement suites were out of scope. | None. | Keep frontend refactoring out of the remaining backend slices. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Fourteen target-owner HTTP operations plus four Workbook preference operations are mapped; no generated edit is authorized. | OpenAPI owner fragments, compatibility projection, generated HTTP binding references, boundary policy; touched this tracker only. | `jq`, `rg`, `sed`, `git log`, `git show`. | Bundle v1/v2 mismatch isolated; authored-owner-before-generator rule recorded. | RB-001. | Resolve owner first; use `make generate` only in a later authorized owner-input task. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | Compatibility v2, source catalog, fixture manifest, traceability, errors, OpenAPI, verification owners, schemas, and generated projections agree. | Authored `contracts/incident-bundles/**`, error/OpenAPI owners, verification inputs, harness attachments/schemas; generated OpenAPI, Go, TypeScript, and topology outputs. | `make generate`, `make generate-drift`, `make generated-artifact-policy-check`, `make json-shape-check`, `make harness-contract`. | SL-07 and generated integrity AC-505 pass; generated files were changed only by the Make generator. | None. | Preserve authored-input-first regeneration in remaining slices. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | 22 Incident Bundle rows and 40 Incidents rows are mapped; boundary-test and support-inventory mismatches are explicit. | Verification contracts, family manifests, test-support inventory, task surface, target tests; touched this tracker only. | `make task-guide`, `make explain-test-owner`, `make explain-target`, `make target-plan-json`, `make help-all`, `jq`, `rg`, `make lint-markdown`. | Command discovery is not test success; the tracker-only Markdown lint passed. | RB-002 and RB-003 affect characterization. | In a later authorized implementation task, run owner slices before changing code. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | AC-487 through AC-507 have selected verification IDs and active owner rows; a new Records owner supplies unit and service-backed portability rows. | Verification registry/owners, all affected family manifests, tests, fixtures, harness attachments, topology, and duration baseline. | Owner discovery for 19 owners, focused unit/service rows, clean full service suite, `make harness-contract`, `make test-fast`, `make check`. | Focused and broad evidence passes; final roots and resolved failures are recorded in Section 9. | None for AG-001 through AG-003. | Reuse the selected rows while executing SL-02. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Live membership/deployment-admin checks, hidden-not-found precedence, path safety, actor isolation, and final-publication recheck are frozen. | Core 04, target routes/access/storage/worker/finalizer, authorization tests; touched this tracker only. | `rg`, `sed`, `jq`. | Existing transport/auth and storage port boundaries classified intentional. | RB-002 for worker recovery composition. | Preserve exact authorization order and recheck timing in every future slice. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | Authorization remains server-derived; actors are inert; source-family and unsupported-version failures are closed and redacted; v1 telemetry is attribute-free and post-commit. | Core security clauses, Incident Bundle errors/routes/worker, actor and extension consumers, Reference Data validator, telemetry registry/tests. | Safe-failure row, owner authorization/telemetry slices, service-backed rollback evidence, `make check`. | AC-499, AC-500, AC-502, and AC-507 pass without public internal-data leakage. | None. | Preserve these conditions in SL-02, SL-03, and SL-05. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-27T16:00:09-04:00 | Codex planning/documentation session | Planning is decision-complete except for three evidence/authority blockers that implementation must not guess through. | This tracker and evidence named by RB-001 through RB-003; touched this tracker only. | `rg`, `sed`, `jq`, Git inspection, Make discovery commands. | SL-00 through SL-08 have dependencies, rollback, validation, and completion criteria. | RB-001, RB-002, RB-003. | Seek an authorized blocker-resolution or characterization slice; do not begin broad refactor work. |
| 2026-07-27T19:19:47-04:00 | Codex AG-001–AG-003 execution session | Adoption and downstream alignment are complete; the remaining refactor work is unrelated boundary/test-support/persistence cleanup. | Final source state and this handoff. | Same-source-state final verification in Section 9. | AG-001 through AG-003 and SL-00, SL-01, SL-04, SL-06, SL-07 are `DONE`. | SL-02, SL-03, SL-05, then SL-08. | Execute SL-02 next, followed by SL-03 and SL-05 before final SL-08 closure. |

## 12. Resolved Decisions and Remaining Adoption Gates

RB identifiers are retained for traceability. Their planning questions are
closed. The AG rows record the completed adoption and alignment package.

| ID | Resolved disposition | Evidence establishing the disposition | Remaining execution dependency | Decision status |
| --- | --- | --- | --- | --- |
| RB-001 | Version `2` MUST be the sole current export format; version `1` MUST remain import-only until IMR-BUNDLE-008 permits an adopted removal. Version and Timeline codec selection MUST be manifest-only and exact. | Core 01 currently names v1; `bundle.go`, Timeline portability, compatibility projection, and current tests implement v2 plus v1 import. The current Timeline provenance model makes v2 the fidelity-preserving current contract. | AG-001, AG-002, AG-003; SL-00, SL-01, SL-07. | `DONE` |
| RB-002 | A nil or unconfigured runner is unsupported when Incident Portability is claimed. Claimed composition MUST fail before publication; all work MUST use the named gated runner. The unclaimed profile has no Incident Bundle route or runner requirement. | Core 01 requires background jobs and mandatory worker/dequeue readiness; production assembly always injects and gates a runner; the current nil path silently skips recovery and permits unmanaged dispatch. | AG-001 and AG-002; SL-01 and SL-06. | `DONE` |
| RB-003 | Only Saved Views currently performs substantial explicit owner validation among the centrally described families; Timeline performs partial bespoke validation outside that registry. The other registered adapters delegate to a generic importer that does not execute `PostImportInvariantCheck`. Every family MUST move to the IMR-PORT typed owner boundary. | Exact adapter reads, `import_targets.go`, the generic `jsonb_populate_record`/conflict-ignore path, special-file reads, and current tests. | AG-001 and AG-002; SL-01, SL-04, and AG-003. | `DONE` |

### Adoption and alignment gates

| Gate | Required action | Dependencies | Status | Exit condition |
| --- | --- | --- | --- | --- |
| AG-001 | Adopt the IMR-BUNDLE, IMR-JOBS, IMR-PORT, IMR-SEC, error, compatibility, and publication behavior in Core 01, including explicit normative registries. | RB-001 through RB-003 and authorized owner edit. | `DONE` | REQ-01-635 through REQ-01-643 define the behavior once; revised legacy clauses contain no v1/v2 contradiction or handoff delegation. |
| AG-002 | Add Core 04 binary acceptance criteria and verification ownership for every adopted AG-001 behavior. | AG-001. | `DONE` | AC-487 through AC-507 map bidirectionally to the adopted requirements and selected verification IDs. |
| AG-003 | Align authored contracts, compatibility state, errors, source catalogs, implementation, fixtures, verification routing, and generated projections through Make-owned workflows. | AG-001, AG-002, SL-04, SL-06, and SL-07. | `DONE` | Owners, projections, implementation, fixtures, and final same-source-state evidence agree; all generated changes came from `make generate`. |

### Adopted requirement mapping

Mappings are intentionally many-to-many where one Core requirement combines
several tracker clauses. The Core requirements, not this mapping, own behavior.

| Core 01 requirement | Adopted tracker requirements |
| --- | --- |
| REQ-01-635 | IMR-BUNDLE-001 through IMR-BUNDLE-003 and IMR-BUNDLE-006 through IMR-BUNDLE-008 |
| REQ-01-636 | IMR-BUNDLE-004, IMR-BUNDLE-005, IMR-BUNDLE-007, IMR-BUNDLE-008 |
| REQ-01-637 | IMR-JOBS-001, IMR-JOBS-005, IMR-JOBS-008 |
| REQ-01-638 | IMR-JOBS-002 through IMR-JOBS-004 and IMR-JOBS-006 through IMR-JOBS-008 |
| REQ-01-639 | IMR-PORT-001 through IMR-PORT-005, IMR-PORT-010, IMR-PORT-012 |
| REQ-01-640 | IMR-PORT-006, IMR-PORT-007, IMR-PORT-011 |
| REQ-01-641 | IMR-PORT-008, IMR-PORT-009, IMR-SEC-001, IMR-SEC-004 |
| REQ-01-642 | IMR-SEC-002, IMR-SEC-003, IMR-SEC-005, IMR-SEC-006 |
| REQ-01-643 | IMR-BUNDLE-009, IMR-JOBS-008, IMR-PORT-010, and AG-003 typed projection/traceability/generated-integrity requirements |

Core 04 adopted the tracker acceptance rows in exact sequence:

| Tracker acceptance IDs | Core 04 acceptance IDs |
| --- | --- |
| IMR-AC-001 through IMR-AC-021 | AC-487 through AC-507, respectively |

The authored `contracts/incident-bundles/traceability.json` projection records
each AC's exact Core requirement and verification-ID set. No adopted
requirement is orphaned and no acceptance criterion creates behavior absent
from Core 01.

### Authored and generated artifact alignment

| Owner or subsystem | Authored artifacts changed |
| --- | --- |
| Core owners | `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/04_security_deployment_and_conformance.md` |
| Incident Bundle contracts | `contracts/incident-bundles/compatibility.json`, `source_catalog.json`, `traceability.json`, `fixture_manifest.json` |
| Public errors and OpenAPI | `contracts/errors/index.json`, Incident Bundle OpenAPI owner, `contracts/openapi-releases/2.0.0.change-set.json` |
| Verification ownership | Affected files under `contracts/verification/owners/**`, the verification registry, `tools/test_catalog_owner.json`, and affected `tools/test_families/**`, including new `module.records` inputs |
| Harness and policy inputs | `tools/backend_module_boundaries.json`, `tools/harness_schema_attachments.json`, four Incident Bundle schemas, harness contract tests, and the Go test-duration baseline |
| Runtime and tests | Incident Bundle/sourceport, application assembly, Jobs, Incidents, all 13 source owners, Reference Data, Network Flow, telemetry, application test support, and their focused tests |

Generated only through `make generate`:

- `contracts/openapi/cartulary.openapi.yaml`;
- `internal/gen/contracterrors/artifacts_gen.go`,
  `internal/gen/contractopenapi/artifacts_gen.go`, and
  `internal/gen/openapioperations/catalog_gen.go`;
- `packages/protocol-ts/src/generated/errors-artifacts.ts`;
- `tools/execution_topology_render_index.json` and
  `tools/scheduler_manifest.json`.

No dependency lockfile or `go.sum` changed. No generated file was hand-edited.

### Verification ownership alignment

- `module.incidentbundles` now owns `compatibility`, `source_catalog`,
  `atomic_coordination`, `named_recovery`, `safe_failures`, and `traceability`.
- `app.server` owns `incident_portability_composition`; `module.extensions`
  owns `incident_portability_jobs`.
- Artifacts, Assessments, Entities, Evidence, Indicators, Links, Parties,
  Records, Revisions, Saved Views, Tasks/Decisions, and Timeline each own
  `verification.incident_portability`.
- Incidents owns `incident_portability` and `incident_portability_actors`;
  Reference Data owns `incident_portability_refs`; Network Flow owns
  `incident_portability_extension`; telemetry owns
  `incident_bundle_v1_import`.
- New active rows cover Records unit/service portability, retained-v1
  translation, named portability-job recovery, safe public failures, special
  consumers, and source-catalog validation. Existing owner rows gained the
  corresponding verification IDs without duplicate row selection.

## 13. Acceptance Criteria and Completion Contract

The IMR acceptance rows remain tracker trace labels. Core 04 owns their adopted
binary form as AC-487 through AC-507; the verification projection and selected
owner rows supply executable evidence.

| Acceptance ID | Requirement trace | Binary acceptance condition | Required evidence | Status |
| --- | --- | --- | --- | --- |
| IMR-AC-001 | IMR-BUNDLE-001, IMR-BUNDLE-007 | Two exports of the same incident and export inputs produce byte-identical v2 structured members, manifest checksums, and archive bytes. | AC-487; deterministic archive fixture and repeated export integration case. | `DONE` |
| IMR-AC-002 | IMR-BUNDLE-001, IMR-BUNDLE-004, IMR-PORT-006 | A valid v2 export imports into an empty deployment and re-exports with the same authoritative IDs, source state, history, attribution, Timeline state, and provenance. | AC-488; full v2 round-trip owner integration row. | `DONE` |
| IMR-AC-003 | IMR-BUNDLE-005, IMR-BUNDLE-008 | A valid v1 archive imports successfully and its subsequent v2 export is semantically equivalent, including translated provenance; malformed or lossy legacy capture fails. | AC-489; retained-v1 translator fixtures and v1-to-v2 integration row. | `DONE` |
| IMR-AC-004 | IMR-BUNDLE-002, IMR-BUNDLE-003, IMR-BUNDLE-006, IMR-BUNDLE-007 | Omitted, null, non-integer, unadmitted, mixed-path, mismatched-path, unknown-core-path, duplicate-core-path, missing-required-path, checksum, traversal, unsupported-member, extracted-byte, compression-ratio, and member-count failures are rejected before source preparation. | AC-490; table-driven archive verification cases. | `DONE` |
| IMR-AC-005 | IMR-BUNDLE-004, IMR-BUNDLE-005, IMR-PORT-006 | Timeline records bind to same-incident `timeline_event` envelopes; source digests and composite identities validate; no provenance row is lost, duplicated, or orphaned. | AC-491; Timeline owner fixtures and bundle integration. | `DONE` |
| IMR-AC-006 | IMR-JOBS-001, IMR-JOBS-005, IMR-JOBS-008 | Claimed invalid composition exits before routes, listeners, readiness, or work. | AC-492; application composition characterization. | `DONE` |
| IMR-AC-007 | IMR-JOBS-001 | Unclaimed composition exposes no Incident Bundle routes or profile-specific runner requirement. | AC-493; claimed/unclaimed composition matrix. | `DONE` |
| IMR-AC-008 | IMR-JOBS-002, IMR-JOBS-003, IMR-JOBS-007 | New and recoverable jobs execute only through the registered named handler behind the dequeue gate. | AC-494; runner and service-backed recovery evidence. | `DONE` |
| IMR-AC-009 | IMR-JOBS-004, IMR-JOBS-006 | Dispatch failure or component loss remains recoverable and restart produces exactly one terminal result. | AC-495; failure, restart, and duplicate-state assertions. | `DONE` |
| IMR-AC-010 | IMR-PORT-001 through IMR-PORT-004, IMR-PORT-007, IMR-PORT-010 | Every required versioned path has exactly one consumer or validator and invalid catalogs fail closed. | AC-496; catalog matrix and complete-path fixtures. | `DONE` |
| IMR-AC-011 | IMR-PORT-002, IMR-PORT-006, IMR-PORT-011 | Every source family rejects SQL-convertible rows that violate owner invariants. | AC-497; owner semantic-invalidity tests and service-backed owner rows. | `DONE` |
| IMR-AC-012 | IMR-PORT-003, IMR-PORT-005, IMR-PORT-006 | Duplicate identities, cross-family orphans, row-count mismatches, and tag inconsistency fail closed. | AC-498; owner and whole-bundle rejection cases. | `DONE` |
| IMR-AC-013 | IMR-PORT-007, IMR-SEC-006 | Source actors are complete, unique, and inert. | AC-499; actor catalog and authorization-state assertions. | `DONE` |
| IMR-AC-014 | IMR-PORT-008, IMR-PORT-009, IMR-SEC-001, IMR-SEC-004 | Failure or cancellation before commit leaves no visible partial success or object. | AC-500; coordinator failure and visibility evidence. | `DONE` |
| IMR-AC-015 | IMR-PORT-005, IMR-PORT-012 | Production import contains no relation-driven SQL interpolation or generic conflict-ignore path. | AC-501; boundary/static and owner-port evidence. | `DONE` |
| IMR-AC-016 | IMR-SEC-002, IMR-SEC-003 | Public failures use safe reasons and details without internal-data leakage. | AC-502; public contract and redaction tests. | `DONE` |
| IMR-AC-017 | IMR-SEC-005 | Archive and row permutation preserve deterministic outcomes. | AC-503; permutation and deterministic fixtures. | `DONE` |
| IMR-AC-018 | IMR-BUNDLE-009, IMR-JOBS-008, IMR-PORT-010 | Requirement, acceptance, verification, and active-row traceability has no orphan or duplicate. | AC-504; traceability and harness shape checks. | `DONE` |
| IMR-AC-019 | AG-003 | Generated artifacts equal their authored inputs and retain generator provenance. | AC-505; generation drift and artifact-policy checks. | `DONE` |
| IMR-AC-020 | IMR-BUNDLE-008 | Version 1 removal remains blocked until all five retirement conditions pass. | AC-506; compatibility policy with pending release clock. | `DONE` |
| IMR-AC-021 | IMR-PORT-007 | Each special input rejects invalid content and publishes no unauthorized state. | AC-507; special-input and rollback evidence. | `DONE` |

### Tracker-revision completion

The documentation revision is complete only when every row below is `PASS`.

| Criterion | Status | Evidence |
| --- | --- | --- |
| The pre-refactor inventory remains available and post-refactor additions/deletions are accounted. | PASS | Section 2 preserves the 56-file baseline; Section 12 records the new sourceport/assembly/owner files and removed generic importer and production hooks. |
| Every decision that previously required implementer judgment has one resolved disposition, owner, gate, slice, and binary acceptance path. | PASS | Sections 6, 8, 10, 12, and 13. |
| Every proposed workflow has dependencies and an exit checkpoint. | PASS | Section 7 contains WF-00 through WF-08. |
| Validation commands are Make-owned or documentation-only commands allowed by repository procedure. | PASS | Section 9. |
| Contradictions use `BLOCKED: owner contradiction`; no adopted owner-owner contradiction was found. | PASS | Sections 1 and 12 distinguish the resolved owner/projection mismatch from an owner-owner contradiction. |
| Handoff history is preserved and current rows exist in every materially affected category. | PASS | Section 11 retains the initial and decision-resolution rows and adds the `2026-07-27T19:19:47-04:00` execution row to all seven categories. |
| AG-001, AG-002, and AG-003 satisfy their exact exit conditions. | PASS | Section 12 records adopted requirements, acceptance mappings, aligned artifacts, and same-source-state evidence. |
| Authored and generated artifacts agree without manual generated edits. | PASS | `make generate`, drift, artifact-policy, JSON-shape, and harness checks pass in Section 9. |
| Markdown and diff hygiene pass after the final handoff edit. | PASS | Draft pass: `make lint-markdown` at `.cartulary/test-results/20260727T232509Z-p1796003` and clean `git diff --check`; the mandatory rerun after this last edit is reported with the final handoff. |
| Obsolete unresolved-decision language is absent from decision-bearing sections. | PASS | The final targeted scan found no active unresolved-decision language outside preserved historical/validation text. |

RB-001 through RB-003 and AG-001 through AG-003 are closed. SL-02 is the next
recommended slice, followed by SL-03 and SL-05; SL-08 remains the final overall
closure slice.
