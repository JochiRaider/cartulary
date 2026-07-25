# app Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/app/{migrate,operator,revisionassembly,server,serverprocess}` and `internal/testutil/appsupport` |
| Target label | Explicit application composition facades |
| Output path | `docs/handoffs/app-module-refactor-tracker.md` |
| Repository state inspected | Branch `main`, commit `6f363b481f539a9d9f41a382bb5dc65db1c09e15` |
| Implementation status | **COMPLETE.** RB-001 is resolved by owner-declared revision contributions and an exact `internal/app/revisionassembly` aggregator; final validation and accounting refresh pass. |
| Allowed change in this session | Production composition, owner facades, revisions validation, tests, guides, repository procedure, boundary policy, harness owner inputs, generated ledgers/schedules, and this tracker. |
| Non-goals | No Core 00-04, domain vocabulary, database migration, public HTTP/WS/view contract, generated protocol, frontend behavior, or visual-golden change. |

Source hierarchy used for this tracker:

1. Adopted subsystem NLSpecs for their named subsystem.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication; no Core 05 claim is made here.
4. `docs/domain.md` and implementation-support guides for vocabulary, boundaries, and execution support.
5. Current repository code and tests for implementation state.
6. Prior plans, handoffs, and the planning framework as evidence only.

Owner and support documents inspected:

- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/domain.md`
- `docs/graph_projection_nlspec.md`
- `docs/network-flow-activity-nlspec.md`
- `docs/opentelemetry-instrumentation-nlspec.md`
- `docs/report-composition-nlspec.md`
- `docs/reporting-subsystem-nlspec.md`
- `docs/testing-harness-nlspec.md`
- `docs/guides/cartulary_repository_bootstrap_guide.md`
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`

Repository evidence inspected includes every file listed in Section 2; the three `cmd/*` callers; callers in
`internal/testutil`, module tests, and module test-support packages; `tools/phase10browserrestore`; the runtime route
registrars and recovery/revision provider surfaces reached by `internal/app`; `tools/backend_module_boundaries.json`;
the test-support inventory, phase/test maps, Make task surface, and boundary checker inputs.

The framework matched the repository's general modular-monolith posture but did not establish `app` as a permanent
root module. The authorized remediation replaces that mixed package with explicit composition facades. Observable
CLI and server behavior remains protected; the repository-internal root Go import is deliberately retired.

### 1.1 Accepted target architecture

```text
internal/app/
  migrate/              # exact cmd/migrate facade
  operator/             # exact cmd/operator facade
  revisionassembly/     # owner contribution aggregation and dependency injection
  server/               # exact cmd/server facade and runtime lifecycle
  serverprocess/        # retained test-only process evidence

internal/testutil/
  appsupport/           # reusable application test composition
```

Revision provider ownership is `source-owner facade -> internal/app/revisionassembly -> internal/modules/revisions
-> internal/app/server`. Source owners construct their provider contributions; Revisions owns the contribution
types, private current-profile requirements, validation, catalogs, and generic coordination. Application assembly
imports owner facades only. Package-init registration, mutable global registries, direct provider-subpackage imports,
and blanket `internal/app/**` authorization are forbidden. Non-nil PostgreSQL and object-store dependencies injected
through `server.Options` are borrowed; the runtime closes only owned resources, once, in reverse acquisition order.

## 2. Pre-Remediation Repository Inventory

This inventory records the inspected baseline before implementation. Superseded source paths are retained here as
historical evidence; the accepted current paths are listed in Section 1.1 and the session handoff.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/app/migrate.go` | Parses and runs the migration binary command, loads config, opens PostgreSQL, runs authored migrations, and maps remediation/errors to process exit behavior. | `RunMigrateCLIContext` | `cmd/migrate/main.go`; its unit tests | Platform config, PostgreSQL, and migrations | `migrate_test.go` | Migration inputs are authored SQL; no generated file is written here. | Narrow `internal/app/migrate` composition facade | High | Exact `migrate up` grammar, context propagation, remediation output, and exit codes are frozen. |
| `internal/app/migrate_test.go` | Characterizes migration CLI grammar, failures, context, and remediation output. | Test-only package surface | `make backend-unit` through test topology | Test doubles around migrate composition | Self-contained tests | None | Stay with migrate facade tests | Medium | Evidence, not runtime architecture. |
| `internal/app/operator.go` | Builds the operator command surface and dispatches object-store init, recovery, and migration-evidence commands. | `OperatorObjectStoreInitResult`, `RunOperatorCLIContext`; exported object-store result schema surface | `cmd/operator/main.go`; recovery process tests | Object store, recovery composition, migration evidence, config/JSON/error helpers | `operator_test.go`, registry/recovery/evidence tests | JSON command envelopes; no generated source | Narrow `internal/app/operator` composition facade | High | Command grammar, single-envelope output, redaction, and exit mapping are observable. |
| `internal/app/operator_migration_evidence.go` | Loads deployment state and builds/writes operator migration-evidence output. | Package-private operator subcommand surface | `operator.go` | Platform config, PostgreSQL, migrations, migration-evidence builder | Unit and integration evidence tests | Evidence JSON only; not Core 05 publication by itself | Operator facade delegating to platform migration-evidence capability | High | Preserve output envelope and DB lifecycle. |
| `internal/app/operator_migration_evidence_integration_test.go` | Verifies migration-evidence behavior against PostgreSQL. | Test-only package surface | `make backend-integration` | PostgreSQL test harness and migration-evidence command | Self-contained test | Evidence artifact content only | Operator integration tests | Medium | Service-backed evidence row. |
| `internal/app/operator_migration_evidence_test.go` | Verifies the Phase 0 migration-evidence command contract. | Test-only package surface | `make backend-unit` | Operator command helpers and fixtures | Self-contained test | None | Operator unit tests | Medium | Preserve its evidence-row mapping when moved. |
| `internal/app/operator_recovery.go` | Parses canonical recovery commands and injects operator operations, stores, and projection restore rebuilding. | Package-private operator subcommand surface | `operator.go` | Recovery `operatorcli`, `operatorops`, object store, PostgreSQL, projection adapters | `operator_recovery_test.go`; module recovery process test | Recovery JSON envelopes; no generated source | Operator facade with behavior owned by recovery | High | Legitimate composition; recovery semantics remain recovery-owned. |
| `internal/app/operator_recovery_test.go` | Freezes canonical/retired recovery grammar, unsafe paths, error envelopes, and closed exit codes. | Test-only package surface | `make backend-unit` | Recovery command parser and error mapping | Self-contained test | None | Operator/recovery composition tests | High | Required before moving the facade. |
| `internal/app/operator_registry.go` | Validates and dispatches exact non-overlapping operator command paths and usage. | Package-private registry surface | `operator.go` | Standard-library context, I/O, and strings | `operator_registry_test.go` | None | Operator facade | Medium | Process composition logic, not a domain module. |
| `internal/app/operator_registry_test.go` | Verifies duplicate, prefix-ambiguous, exact, and invalid-namespace routing. | Test-only package surface | `make backend-unit` | Operator registry | Self-contained test | None | Operator unit tests | Medium | Keep adjacent to registry implementation. |
| `internal/app/operator_test.go` | Verifies object-store init output, redaction, and retired operator aliases. | Test-only package surface | `make backend-unit` | Operator facade and object-store test doubles | Self-contained test | JSON result schema | Operator unit tests | High | Protects public command behavior. |
| `internal/app/revisions.go` | Constructs the revisions command service, record catalogs, imported attribution, delete/restore/rollback providers, and projection adapters across peer owners. | `NewRevisionsCommandService` | `runtime.go`; parties Phase 9 test; revisions Phase 7 test | Revisions plus peer owner-internal provider packages and projection adapters | `revisions_test.go`; parties/revisions tests | Indirect protocol/view effects; no generated source imported directly | Dedicated narrow revision-composition boundary; exact package deferred pending cycle proof | Critical | Main reason for the broad `internal/app/**` owner-port exception. Do not move peer behavior into revisions. |
| `internal/app/revisions_test.go` | Checks that every required revision provider is registered. | Test-only package surface | `make backend-unit` | Revision service composition and provider catalog | Self-contained test | Provider/kind contract coverage | Revision-composition tests | High | Catalog completeness needs stronger characterization before movement. |
| `internal/app/runtime.go` | Assembles config validation, profiles, secrets, auth, telemetry, PostgreSQL, object store, bootstrap, modules, jobs, WebSocket hub, revisions, routes, HTTP handler, and cleanup. | `Options`, `Runtime`, `NewRuntime`, `(*Runtime).Close` | `server.go`; `internal/testutil/httptestx`; Phase 10 browser-restore tool | Many domain modules and platform adapters | Runtime Phase 0 tests; all server process tests; module harness tests | Indirect generated protocol/view contracts through registered modules | Narrow server/application runtime composition facade | Critical | Legitimate assembly mixed with a large route/dependency graph; startup order and cleanup are contracts. |
| `internal/app/runtime_phase0_integration_test.go` | Verifies invalid config, bootstrap creation/rollback/failure/skip/recovery, and readiness against services. | Test-only package surface | `make backend-integration` | Runtime, PostgreSQL/object-store harnesses, bootstrap fixtures | Self-contained test | None | Server runtime integration tests | High | Baselines fail-closed startup. |
| `internal/app/runtime_phase0_test.go` | Verifies fail-closed Phase 0 startup without broad service setup. | Test-only package surface | `make backend-unit` | Runtime seam/test dependencies | Self-contained test | None | Server runtime unit tests | High | Preserve before decomposing runtime construction. |
| `internal/app/server.go` | Loads server config, selects build profile, creates/closes runtime, serves HTTP, writes diagnostics, and maps cancellation/errors to exit codes. | `RunServerContext` | `cmd/server/main.go`; server tests | Config, application runtime, HTTP runtime, diagnostics | `server_test.go`; process suite indirectly | HTTP/process envelope, no direct generated source | Narrow `internal/app/server` composition facade | High | Thin process facade is legitimate; current root placement is not required. |
| `internal/app/server_profile_harness.go` | Build-tagged harness profile that enables guarded test clocks, controls, module test routes, and inherited listeners. | Build-profile package surface only | Harness build of `server.go` | Auth/saved-view/timeline/network-flow test routes; harness and HTTP runtime | Harness profile and server-process tests | Harness contracts only | Server harness adapter, retained outside production module logic | Critical | Exact `CARTULARY_ENABLE_TEST_ROUTES=1` gate is security-sensitive. |
| `internal/app/server_profile_harness_test.go` | Verifies test routes enable only for the exact value `1`. | Test-only package surface | Harness unit target | Harness server profile | Self-contained test | Harness route contract | Server harness tests | High | Must remain in a harness-tagged validation path. |
| `internal/app/server_profile_production.go` | Build-tagged production profile that rejects harness-only environment and serves normally. | Build-profile package surface only | Production build of `server.go` | Config diagnostics and HTTP runtime | `server_test.go` | None | Server production adapter | High | Fail-closed rejection prevents test controls in production. |
| `internal/app/server_test.go` | Freezes diagnostics, cleanup, setup/serve failures, cancellation, profile rejection, and writer failure behavior. | Test-only package surface | `make backend-unit` | Server seams, runtime, config/profile logic | Self-contained test | None | Server facade tests | High | Characterization baseline for server split. |
| `internal/app/serverprocess/embedded_frontend_process_test.go` | Verifies packaged root and embedded frontend assets through a server process. | Process-test package surface | `make backend-process` | Shared process harness and HTTP client | Self-contained process test | Embedded web build artifact contract | Server packaging/process tests | High | This is not frontend shell/controller ownership. |
| `internal/app/serverprocess/networkflow_runtime_routes_process_test.go` | Verifies Network Flow harness route contribution in the real server process. | Process-test package surface | `make backend-process` | Shared harness and network-flow test controls | Self-contained process test | Harness route contract | Network Flow process evidence, hosted by server suite | High | Behavior belongs to Network Flow; app proves assembly. |
| `internal/app/serverprocess/phase0_e2e_test.go` | Verifies readiness, invalid-config diagnostics, bootstrap success/failure/skip/recovery in a process. | Process-test package surface | `make backend-process` and Phase 0 topology | Shared process harness, config/bootstrap fixtures | Self-contained process tests | Phase 0 evidence rows | Server/bootstrap process evidence | Critical | Do not infer runtime ownership from the phase label. |
| `internal/app/serverprocess/phase10_config_test.go` | Verifies effective backup-root configuration exposed to recovery behavior. | Process-test package surface | `make backend-process` and Phase 10 topology | Shared process harness and config route/client | Self-contained process test | Phase 10 configuration evidence | Recovery/config process evidence | High | Protects configuration wiring, not an app-owned domain rule. |
| `internal/app/serverprocess/phase10_recovery_sentinel_test.go` | Verifies fresh-environment restore consistency, projection rebuilding, and absence of forbidden public recovery routes. | Process-test package surface | `make backend-process` and Phase 10 topology | Recovery operator process, runtime, DB/object store, HTTP | Self-contained process tests | Phase 10 recovery evidence | Recovery process evidence | Critical | Cross-process sentinel; retain until replacement has equal coverage. |
| `internal/app/serverprocess/phase1_process_test.go` | Supplemental process smoke for login/logout, CSRF, socket cap, enrollment, password/admin/TOTP flows. | Process-test package surface | `make backend-process`; supplemental Phase 1 map | Shared harness, auth HTTP/WebSocket clients | Self-contained process tests | Supplemental smoke accounting | Auth process evidence | Critical | Explicitly supplemental, not authoritative row ownership. |
| `internal/app/serverprocess/phase2_smoke_test.go` | Process smoke for incidents, workbook preferences, membership administration, extensions, and deployment-admin boundaries. | Process-test package surface | `make backend-process`; Phase 2 map | Shared harness and HTTP clients | Self-contained process tests | Phase 2 smoke accounting | Cross-module process evidence | High | Mixed scenario belongs to process validation, not an app domain module. |
| `internal/app/serverprocess/phase5_smoke_test.go` | Process smoke for evidence upload/attachment and timeline projection. | Process-test package surface | `make backend-process`; Phase 5 map | Shared harness, evidence/timeline HTTP clients | Self-contained process test | Phase 5 smoke accounting | Evidence/timeline process evidence | High | Protects cross-module assembly and projection visibility. |
| `internal/app/serverprocess/runtime_routes_process_test.go` | Verifies test routes are disabled by default, fail closed, and preserve security/reset behavior when enabled. | Process-test package surface | `make backend-process` | Shared harness, test controls, auth and HTTP | Self-contained process tests | Harness route/security contract | Harness process evidence | Critical | Security characterization for any server facade change. |
| `internal/app/serverprocess/shared_process_harness_test.go` | Provides `TestMain` and shared PostgreSQL/S3/server-process lifecycle helpers. | Test-only shared harness surface | Every test in `serverprocess` | Process execution, PostgreSQL/S3 test harnesses, fixtures | All server-process files | Harness lifecycle/accounting | Server process test support; location may stay adjacent | High | Test-only composition; phase maps do not dictate runtime packages. |
| `internal/app/testsupport/incident_store.go` | Builds an incident store with workbook startup preference bootstrap for module tests. | `NewIncidentStore` | Incident and timeline store/scenario test support | Incidents store, workbook startup bootstrap, PostgreSQL | Downstream module tests | Store/view preference behavior indirectly | `internal/testutil/appsupport` or equivalent test-only owner | Medium | Reusable test composition is misplaced under production assembly. |
| `internal/app/testsupport/runtime.go` | Starts PostgreSQL/S3 test services, isolated databases, object-store buckets, and an application HTTP test server. | `Runtime`, `ServerHarness`, `ServerOptions`, `StartRuntime`, database/server methods | Auth, collaboration, incidents, and timeline test-support packages | `internal/testutil` harnesses, HTTP API, object store, fixtures | Downstream module tests | Test route and HTTP contracts indirectly | `internal/testutil/appsupport` or equivalent | High | Test-only service facade; inventory currently declares `app_runtime`/`platform_facade`. |

## 3. Module Boundary Diagnosis

The target is a **mixed-responsibility package**. It contains a legitimate thin application/service composition role,
transport- and persistence-adjacent assembly, a revision mutation coordinator, and test orchestration. It is not a
frontend shell/controller, grid-vendor integration layer, or owner of projection/view-schema semantics. The package is
also an accidental catch-all to the extent that three binary facades, revision registration, runtime assembly, and
reusable test support share one root.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Binary process/application composition | Root `internal/app` | Narrow server, migrate, and operator composition facades | split | Three `cmd/*` roots each call a different exported app entry point. | Preserve thin `cmd/*` composition roots. |
| Server lifecycle and module assembly | `runtime.go`, `server.go`, profile files | Narrow server/runtime application facade | keep/split | `NewRuntime`, route registrars, resource construction/close, and `RunServerContext` | Legitimate app responsibility; reduce its import and public surface. |
| Migration CLI composition | `migrate.go` | Narrow migrate application facade | move | Only `cmd/migrate` calls `RunMigrateCLIContext`. | Migration behavior stays platform/migrations-owned. |
| Operator CLI composition | Operator source files | Narrow operator application facade | move | Only `cmd/operator` uses the dispatcher; subcommands delegate to recovery/platform capabilities. | Keep one public operator process contract. |
| Revision provider/catalog composition | `revisions.go` | Exact revision-composition package or owner registration facade | defer | Direct imports of many peer owner-internal provider packages; broad boundary exception | Final package needs import-cycle and ownership proof before implementation. |
| Projection refresh/rebuild wiring | `runtime.go`, `revisions.go`, `operator_recovery.go` | Projection adapters plus calling composition facade | keep | Calls approved projection adapters and recovery rebuilder; no projection logic found in app. | Delegated orchestration, not projection ownership. |
| HTTP and WebSocket route graph assembly | `runtime.go` | Server/runtime facade; semantics stay with registering modules | keep | Module `RegisterRoutes` functions and collaboration hub are composed into HTTP API. | Characterize omission, order, dependency, and authorization failures. |
| Saved-view and view-schema wiring | `runtime.go` through module registrars | Saved Views and view-contract owners | keep | No app-local schema or saved-view algorithm found. | App only risks mis-wiring or omission. |
| Test-only runtime/store composition | `internal/app/testsupport` | `internal/testutil/appsupport` or equivalent | move | Imported by multiple module test-support packages; service-starting inventory marks a test facade. | Later task must update test inventory/accounting. |
| Server process scenarios | `internal/app/serverprocess` | Process-evidence suite adjacent to server composition, with subsystem ownership recorded in maps | defer | Scenarios span auth, incidents, evidence, recovery, and network flow. | Do not reorganize solely from historical phase names. |
| Embedded frontend delivery | Server process test; actual serving in platform/runtime assembly | Web build owner plus server packaging adapter | keep | Process test observes root/assets; no app frontend state exists. | Not a frontend module boundary. |
| Grid-adapter/vendor behavior | Not found | Grid adapter/UI package if later found | defer | No direct grid-vendor imports or implementation in `internal/app`. | No action without new evidence. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Go process entry points and runtime lifecycle | Application composition | `RunServerContext`, `RunMigrateCLIContext`, `RunOperatorCLIContext`, `Options`, `Runtime`, `NewRuntime`, `Close` | App unit/integration tests; `httptestx` consumers | Freeze construction failure ordering, cleanup once-only behavior, and all repository callers before moving packages. | Critical | Internal Go paths may change only in a caller-migration slice. |
| Migration command grammar and output | Migrations/platform behavior composed by app | `migrate.go` | `migrate_test.go` | Existing coverage is adequate; retain exact unsupported-grammar-before-config and remediation cases. | High | Preserve exit codes and stderr format. |
| Operator object-store init and migration evidence | Object store and migration-evidence owners, composed by operator facade | Operator source and JSON result type | Operator unit and integration tests | Add a facade-level registry snapshot if command separation changes construction. | High | Preserve redaction and one-envelope behavior. |
| Recovery CLI grammar, safety, restore, and projection rebuilding | Recovery subsystem | `operator_recovery.go`; Core 01 recovery sections | Recovery unit tests; Phase 10 sentinel; recovery module process tests | Add a composition test proving the exact rebuilder/store dependencies after movement. | Critical | No public recovery HTTP route may appear. |
| Startup, health, readiness, bootstrap, and shutdown | Platform runtime/auth bootstrap with app orchestration | `runtime.go`, `server.go` | Runtime Phase 0 tests; Phase 0 process tests; server unit tests | Add an assembly-order/cleanup characterization around telemetry, jobs, hub, object store, and DB failure seams. | Critical | Fail closed and close every initialized resource. |
| HTTP route graph and request/response envelopes | Each registering module; HTTP API owns transport aggregation | Registrars assembled in `runtime.go` | Module tests; server process Phase 0/1/2/5/10/Network Flow tests | Add one route-contribution inventory test that detects omitted or duplicated registrars without redefining route semantics. | Critical | Includes auth/account/users, incidents, extensions, jobs, imports, Network Flow, reporting, reference packs, bundles, Saved Views, view schemas, entities, evidence, assessments, workbook, timeline, and revisions. |
| Canonical incident WebSocket path, authorization, and event semantics | Collaboration | Collaboration route registrar and hub supplied by runtime | Phase 1 process socket test; collaboration tests | Retain an authorized `GET /ws/v1/incidents/{incident_id}` composition smoke and unauthorized failure. | Critical | Do not infer obsolete view-specific paths as active contracts. |
| Harness-only row/query/mutation/save-state/conflict/inspector controls | Harness runtime and contributing modules | Harness profile additional routes | Harness profile unit; runtime-route and Network Flow process tests | Add a contribution-set assertion if profile construction moves. | Critical | Exact enable gate is `CARTULARY_ENABLE_TEST_ROUTES=1`; production rejects harness env. |
| Entity row/query/mutation and revision/change-set behavior | Entities, workbook, revisions, and record owners | Route registrars plus revision catalogs | Revisions catalog test; Phase 7/9 tests; module tests | Enumerate expected record kinds/providers in a table-driven characterization independent of package path. | Critical | Preserve delete/restore/rollback, imported attribution, and projection side effects. |
| Saved-view and view-schema behavior | Saved Views and view-contract owners | Registrars/dependencies passed by runtime | Module tests and harness routes | No app-specific semantic test; route/dependency inventory is sufficient unless movement exposes a new seam. | High | Never duplicate schemas in app. |
| Projection refresh behavior | Projection modules and approved adapters | Revision, recovery, evidence, timeline wiring | Phase 5 smoke; Phase 10 sentinel; projection module tests | Preserve rebuild invocation and visible post-mutation state for each revision provider group. | Critical | App must not absorb projection algorithms. |
| Storage semantics and transaction ownership | PostgreSQL/object store and domain stores | Runtime/store construction | Runtime integration; migration evidence; Phase 10; module integration tests | Characterize partial-startup cleanup and transaction boundary only where current seams are uncovered. | High | Composition may inject stores but must not relocate their semantics. |
| Authorization outcomes and enterprise reconciliation | Auth/authorization owners | Auth manifest/reconcile and route dependencies in runtime | Phase 1/2 process tests; auth module tests | Add only missing composition cases for lost-admin/reconcile failure and route dependency omission. | Critical | Authorization stays in module/platform layers. |
| Embedded frontend root/assets | Web build and server packaging | Process observation | Embedded frontend process smoke | Preserve root, asset status/content headers, and missing-asset behavior if server facade changes. | High | No frontend shell state exists in app. |
| Generated protocol and view contracts | `contracts/*` and generators | Modules reached by runtime consume generated surfaces; app has no direct generated import | Generate-drift and module tests | No generated hand edit; run drift checks after any contract-adjacent change. | High | App refactor should not change generated output. |
| Harness/test evidence accounting | Testing harness owner | Phase maps, test rows, test-support inventory, task surface | Harness contract checks | Update paths only after moves, preserving row identity and supplemental/authoritative posture. | High | Evidence accounting is not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| The root combines three binary facades, server runtime, revisions, and test concerns. | All 33 files and three `cmd/*` callers | High | `should_fix` | Narrow application composition facades | Separate by process and composition responsibility through behavior-preserving slices. |
| Revision composition imports many peer owner-internal provider packages. | `revisions.go`; owner-port rules in `tools/backend_module_boundaries.json` | Critical | `must_fix` | Exact revision-composition boundary plus owner-exposed ports | Prove a non-cyclic registration design, then remove the blanket importer. |
| `internal/app/**` is broadly authorized as a revision-provider composition importer. | `owner_port_only_imports` allowlists | Critical | `must_fix` | Exact importer package(s) | Later boundary update must name the exact composition location; no wildcard app subtree. |
| Every production `cmd/*` internal import is restricted to the broad `internal/app` prefix. | `thin-command-production-imports` policy and three main files | High | `must_fix` | Exact server/migrate/operator facades | Replace with command-specific exact facade allowances after callers migrate. |
| App test support is imported inward by several module test-support packages. | Imports of `internal/app/testsupport` from auth, collaboration, incidents, and timeline tests | Medium | `should_fix` | `internal/testutil/appsupport` or equivalent | Move in an isolated test-only slice and update the service-starting inventory/maps. |
| Runtime directly constructs platform adapters and domain module dependencies. | `runtime.go` | High | `intentional/no_action` | Application composition | Keep dependency construction here unless a concrete domain behavior is found. |
| Recovery behavior is delegated while app injects stores and projection rebuilder. | `operator_recovery.go` and recovery owner APIs | High | `intentional/no_action` | Recovery owns behavior; operator facade owns composition | Freeze recovery CLI and rebuild outcomes; do not move recovery logic into app. |
| Revision mutation/projection effects are obscured by a single large provider catalog. | `revisions.go` record catalogs and adapters | Critical | `should_fix` | Revision composition plus record owners | Add catalog characterization and explicit registration groupings before relocation. |
| Harness-only routes are selected in the application server profile. | Build-tagged profile files and exact-one test | Critical | `intentional/no_action` | Server harness adapter | Preserve build tags, exact gate, production rejection, and route security. |
| No app-local saved-view/view-schema algorithm or duplicated row logic was found. | Runtime uses registrars/dependencies; no matching implementation in app | Medium | `intentional/no_action` | Existing subsystem owners | Recheck during implementation diff; do not create an app schema abstraction. |
| No direct grid-vendor coupling was found. | Imports and source inspection | Low | `intentional/no_action` | Existing frontend grid adapter | No app work unless later evidence changes. |
| Generated files are not directly edited or owned by app. | Import scan and generated-artifact policy | High | `intentional/no_action` | Contract generators | Run drift checks; never hand-edit generated roots. |
| Phase/test maps locate app tests but do not define runtime ownership. | Testing harness NLSpec and phase maps | Medium | `intentional/no_action` | Harness evidence accounting | Change maps only for path/accounting consequences of an authorized move. |

No owner-document contradiction was found. If a later session finds one, it must record
`BLOCKED: owner contradiction` and stop that decision.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 through WF-08 | Lock scope, authority, repository revision, and sole-write rule. | Tracker and owner/support documents | Repository status and source hierarchy review | Scope and authority row is current. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory all 33 target files and inbound/outbound edges. | `internal/app/**`, callers, task maps | File enumeration and import/symbol searches | Every target file has a Section 2 row. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map app-observable behavior to its normative/module owner. | Core docs, NLSpecs, route/provider code, tests | Evidence links and owner review | Every contract risk has an owner/test posture. |
| WF-03 | Characterization test gap analysis | parallel | WF-01, WF-02 | WF-05, WF-06 | Identify missing tests without implementing them. | App and module tests, harness rows | Test name/topology inspection | Gaps are named in Section 4 and Slice APP-S00. |
| WF-04 | Boundary/coupling scan | parallel | WF-01, WF-02 | WF-05 | Classify imports, test support, platform adjacency, and policy exceptions. | App imports, boundary manifest/checker, generated policy | `make backend-module-boundary-check` in later implementation | All findings are classified. |
| WF-05 | Facade and ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Define narrow facades and the unresolved revision-composition proof obligation. | App packages and affected callers | Static/import-cycle design review | Facade roles are fixed; revision location remains blocked until proven. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Order independently reversible behavior-preserving moves. | Production/test packages named in Section 7 | Slice-specific Make targets | Every slice has dependency, rollback, and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Preserve evidence rows and test-support declarations after path moves. | Phase maps, test-support inventory, task manifests | Harness and JSON-shape checks | Accounting changes are limited to actual moved paths. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | None | Run narrow-to-broad verification and record actual results. | Whole affected surface and tracker | Section 8 commands | Clean diff, results, blockers, and next action recorded. |

## 7. Proposed Refactor Slice Plan

These implementation slices are authorized by the remediation task. Observable product/operator behavior remains
preserved except for the explicitly corrected lifecycle ownership defects and retired repository-internal root import.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| APP-S00 | WF-03 | Add missing composition characterization before moving code: route contribution inventory, startup/cleanup seams, revision catalog groups, and operator recovery dependencies. | Existing app tests plus narrowly owned module tests | Tests must describe current behavior, not invent ownership or semantics. | Preserve all existing tests; add table-driven route/provider and failure-cleanup cases. | `make backend-unit`; `make backend-integration`; `make backend-process` | Revert characterization-only commits if they encode unsupported behavior. | Tests fail on an intentionally simulated omitted registrar/provider and pass on the unchanged baseline. |
| APP-S01 | APP-S00 | Move reusable application service/store test composition from `internal/app/testsupport` to `internal/testutil/appsupport` or the exact repository-approved testutil name. | App testsupport and auth/collaboration/incidents/timeline test-support callers | Test fixture lifecycle, DB isolation, bucket setup, test-route mode | Preserve downstream module test suites and service-starting posture. | `make backend-unit`; `make backend-integration`; `make json-shape-check` | Restore old package/imports and inventory rows as one slice. | No production import is introduced; all callers use the testutil location; accounting is current. |
| APP-S02 | APP-S00 | Isolate revision provider/catalog construction behind one exact composition boundary and migrate callers atomically without forwarding wrappers. | Revisions contribution types, source-owner facades, `internal/app/revisionassembly`, parties/revisions tests, server runtime | Provider completeness, authorization, imported attribution, delete/restore/rollback, projection refresh, cycles | Preserve Phase 7/9 and catalog tests; add missing, duplicate, unexpected, nil, and ownership validation. | `make backend-unit`; `make backend-integration`; `make backend-module-boundary-check` | Restore the old constructor and catalog as a unit; do not partially register providers. | Exact importer design is cycle-free, every provider remains registered, and no peer behavior moves into revisions/app. |
| APP-S03 | APP-S00 | Separate migrate and operator application facades and migrate their command roots while keeping exact command behavior. | Migrate/operator sources and tests; `cmd/migrate`; `cmd/operator`; recovery and platform dependencies | CLI grammar, stdout/stderr, JSON schemas, redaction, exit codes, DB/object-store closure | Preserve all migrate/operator tests and recovery process coverage; add facade registry snapshot if needed. | `make backend-unit`; `make backend-integration`; `make build-migrate`; `make build-operator` | Keep each command facade move in its own revertible commit if practical. | Commands import their exact facade and emit byte/shape-compatible outcomes for covered cases. |
| APP-S04 | APP-S00, APP-S02 | Separate server/runtime composition facade atomically, keeping module route semantics delegated and build profiles secure; add no forwarding facade. | Runtime/server/profile files, `cmd/server`, `httptestx`, Phase 10 tool, process harness | Startup order, readiness, routes, WebSocket, authorization wiring, profile gates, cleanup, embedded assets | Preserve runtime/server tests and all server-process scenarios; add exact assembly-order/route inventory and cleanup tests. | `make backend-unit`; `make backend-integration`; `make backend-process`; `make build-server` | Restore the pre-move package as a unit if callers cannot be migrated atomically. | All callers use the narrow server facade; production rejects harness inputs; route and lifecycle baselines pass. |
| APP-S05 | APP-S01, APP-S02, APP-S03, APP-S04 | Remove transitional root facades and replace broad boundary policy with exact command/facade and revision-composition allowances. | Remaining app root, all callers, `tools/backend_module_boundaries.json`, boundary checker fixtures | Accidental import expansion, missing callers, command composition escape, generated-policy drift | Preserve boundary checker tests and scan all production/test imports. | `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check` | Restore forwarding facades and previous policy together if exact rules cannot pass. | No blanket `internal/app/**` revision authorization remains; each `cmd/*` has only its matching exact facade. **This work is authorized** to edit production, tests, and tools policy. |
| APP-S06 | APP-S01, APP-S05 | Update test-support inventory, phase/test paths, and generated topology only where prior slices changed owned inputs; regenerate through Make. | Owner manifests/maps and downstream generated topology outputs | Evidence row loss, supplemental/authoritative posture drift, hand-edited generated outputs | Preserve stable evidence IDs and suite selection. | `make generate`; `make generate-drift`; `make json-shape-check` | Revert owner-input changes and regenerate; never hand-edit outputs. | Generated drift is clean and every moved test retains its prior evidence posture. **This work is authorized** for harness inputs. |
| APP-S07 | APP-S06 | Final narrow-to-broad validation, diff audit, and implementation handoff. | Entire affected surface | Latent process, security, contract, or topology regression | Run all preserved/new tests and inspect artifacts for failures. | `make agent-finalize`; then `make check` | If broad validation fails, revert only the owning slice after identifying the first bad boundary. | Required commands pass or failures are reported with run roots and relatedness; handoff has no undisclosed changes. |

## 8. Validation Plan

Canonical target discovery used `make help`, `make help-all`, and `make explain-target` for the relevant targets. Those
commands described the task surface; they did not execute the validation suites.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Backend unit tests, including app CLI/runtime/revision tests | yes | Run for baseline and after each code slice. |
| integration | `make backend-integration` | Service-backed backend integration, including bootstrap and migration evidence | yes | Requires repository-managed PostgreSQL/object-store readiness. |
| e2e/browser | `make backend-process` | Server-process contracts across Phases 0/1/2/5/10 and Network Flow | yes | Add `make browser-e2e-webserver-backed` only when embedded web or browser-observable routing is affected. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Generated outputs, protected roots, JSON manifests/maps | yes | Use `make generate` only after authorized owner-input changes; never hand-edit generated files. |
| import-boundary/static | `make backend-module-boundary-check` | Module imports, command-root restrictions, owner-port allowlists | yes | Must fail against an intentionally broad/old rule once exact rules are adopted, then pass. |
| full check | `make agent-finalize`; then `make check` | Final repository finalization and broad verification | no | Run after slice validations; retain/report the successful run root when available. |

Tracker-only verification completed as follows:

| Command | Result | Evidence |
| --- | --- | --- |
| `make lint-markdown` | PASS | Exit code 0; the target emitted no separate result artifact. |
| `make generated-artifact-policy-check` | PASS | Run root `.cartulary/test-results/20260713T223939Z-p2107210` |
| `make json-shape-check` | PASS | Run root `.cartulary/test-results/20260713T223944Z-p2107398` |

No backend unit, integration, process, browser, generation, boundary, or full-check suite was run because this session
changed only the planning tracker. `make agent-finalize` was also skipped because no broader end-of-run verification
or retained successful warm-check result was used.

Implementation-session verification supersedes that planning-only posture:

| Validation | Result | Evidence |
| --- | --- | --- |
| Pre-change backend baselines | PASS | Unit `20260713T231312Z-p2121953` (236 tests); integration `20260713T231325Z-p2123562` (172); process `20260713T231408Z-p2129337` (34); boundary `20260713T230029Z-p2118987`. |
| Exact facade builds | PASS | `build-server` `20260713T232157Z-p2160032`; `build-migrate` `20260713T232204Z-p2166787`; `build-operator` `20260713T232206Z-p2167993`. |
| Backend unit/integration/process | PASS | Current focused runs: unit `20260713T233109Z-p2199514` (236 tests), integration `20260713T233329Z-p2212959` (172), process `20260713T233414Z-p2221802` (34). |
| Phase slices | PASS | Phase 0 `20260714T000004Z-p2962784`; Phase 1 `20260713T233532Z-p2250514`; Phase 2 `20260713T233706Z-p2274369`; Phase 3 `20260713T233735Z-p2292490`; Phase 7 `20260713T233803Z-p2313847`; Phase 9 `20260713T233826Z-p2327682`; Phase 10 `20260713T233900Z-p2344115`; Phase 12 `20260713T233928Z-p2364373`. |
| Boundary, harness, JSON, generation, drift, lint, and security | PASS | Exact boundary `20260713T233140Z-p2206503`; harness contract `20260713T233259Z-p2211681` (53 tests); lint `20260713T234016Z-p2389913`; targeted gosec `20260713T234103Z-p2405687`; generation and drift targets passed. |
| Browser-backed validation | PASS | `browser-e2e-webserver-backed` `20260713T233952Z-p2378426`; full checks also executed browser functional shards. |
| Retained-run maintenance | PASS | Warm check root `20260713T235353Z-p2779004`; `agent-finalize RESULTS_DIR=...` passed at `20260713T235534Z-p2861180` and refreshed three timing/accounting files. |
| Final broad validation | PASS | `make check` root `20260714T000431Z-p3067373`: 136/136 work units, 1,120 tests, zero failed/missing/unmapped. |

Two earlier broad attempts failed only because `/tmp` was 99% full (`ENOSPC` during linking/cache writes), and one
diagnostic rerun correctly rejected an in-repository `TMPDIR`. The final canonical checks used external
`TMPDIR=/home/jochi/.cache/cartulary-tmp` and Go caches on the main filesystem; the failures were infrastructure-only
and unrelated to the refactor. A prune-only duration refresh also proved too narrow for required supplemental shards;
the final baseline was rebuilt through Make from successful `test-fast` plus retained warm-check evidence, and obsolete
root-package timing keys were removed before schedules and coverage were revalidated.

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| APP-001 | Lock target scope, authority, and sole-write rule | WF-00 | DONE | None | Section 1 | Authority and non-goals are explicit. |
| APP-002 | Inventory all 33 target files and callers | WF-01 | DONE | APP-001 | Section 2 | Every file has a concrete row. |
| APP-003 | Map public contracts to owners/tests | WF-02 | DONE | APP-002 | Section 4 | Every discovered risk has owner and test posture. |
| APP-004 | Identify characterization gaps | WF-03 | DONE | APP-003 | Section 4 and APP-S00 | Gaps are named without test edits. |
| APP-005 | Classify coupling and boundary exceptions | WF-04 | DONE | APP-002, APP-003 | Sections 3 and 5 | Broad app exceptions are `must_fix`; no unsupported ownership claim remains. |
| APP-006 | Implement characterization baseline | WF-03 | DONE | APP-004, RB-001 | Baseline run roots in Section 10 | Authorized baseline passed before implementation. |
| APP-007 | Relocate reusable app test support | WF-06, WF-07 | DONE | APP-006, RB-001 | `internal/testutil/appsupport` and six owner-local consumers | No production package imports testutil; obsolete dedicated inventory entry is removed. |
| APP-008 | Prove and isolate revision composition | WF-05, WF-06 | DONE | APP-006, RB-001 | Owner facade contributions and `internal/app/revisionassembly` | Exact cycle-free boundary retains the ten record and seven non-row capabilities. |
| APP-009 | Split migrate/operator facades | WF-06 | DONE | APP-006, RB-001 | `internal/app/migrate`, `internal/app/operator` | Exact command callers compile and CLI characterization remains green. |
| APP-010 | Split server/runtime facade | WF-06 | DONE | APP-006, APP-008, RB-001 | `internal/app/server` | All runtime callers migrated; exact route registry and owned-resource stack added. |
| APP-011 | Replace broad boundary-manifest authorization | WF-04, WF-06 | DONE | APP-007 through APP-010, RB-001 | `tools/backend_module_boundaries.json` and negative fixtures | Exact command, owner-provider, incident-bundle, harness-profile, and revision-assembly rules pass. |
| APP-012 | Update harness/test accounting from moved paths | WF-07 | DONE | APP-007, APP-011, RB-001 | Phase maps, support inventory, occurrence classifications, generated ledgers/schedules | Stable evidence identities and generated topology use current paths. |
| APP-013 | Run implementation finalization and full check | WF-08 | DONE | APP-012 | Finalization `20260713T235534Z-p2861180`; final check `20260714T000431Z-p3067373` | Required targets pass; infrastructure-only failed attempts are attributed in Section 8. |
| APP-014 | Complete implementation handoff | WF-08 | DONE | APP-013 | Sections 8 and 10 | Diff, results, blockers, and final state are current. |
| APP-015 | Create and validate this planning tracker | WF-00 through WF-08 | DONE | APP-001 through APP-005 | This file and Section 8 results | Narrow documentation/policy checks passed and final status is tracker-only. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Scope and precedence locked at commit `6f363b4`; tracker created as sole intended write. | Inspected framework, Core 00-04, adopted NLSpecs, domain guide, app tree, callers, and policies; touched this tracker only. | Framework/source reads; `git status --short --branch`; `git rev-parse HEAD`; `rg --files internal/app` | Target exists with 33 files; no owner contradiction found; initial tree was clean. | RB-001 | Run tracker-only checks and update this log with actual results. |
| 2026-07-13T18:40:02-04:00 | Codex / initial app planning | Tracker-only verification complete. | Touched this tracker only; `.cartulary/test-results` contains ignored command artifacts. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git status --short` | All three checks passed; repository status listed only the new tracker as an authored change. | RB-001 apply only to later implementation. | Hand off the tracker; begin APP-S00 only after later authorization. |
| 2026-07-13T19:26:48-04:00 | Codex / authorized remediation | Exact facades, owner revision contributions, server lifecycle ownership, test-support move, boundary policy, and phase accounting implemented. | Production Go, tests, guides, repository procedure, boundary/harness owner inputs, generated ledgers and schedules, and this tracker. | Baselines: `make backend-unit`, `make backend-integration`, `make backend-process`, `make backend-module-boundary-check`; implementation: format/build/unit/boundary/JSON/generators. | Baselines passed at `20260713T231312Z-p2121953`, `20260713T231325Z-p2123562`, `20260713T231408Z-p2129337`, and `20260713T230029Z-p2118987`; focused current unit and boundary runs pass. | None; RB-001 closed. | Complete narrow phase/process validation, finalization, broad check, and duration refresh. |
| 2026-07-13T20:02:09-04:00 | Codex / authorized remediation | Remediation and end-to-end handoff complete. | Final implementation, tests, owner docs, policy, harness accounting, generated schedules/ledgers, duration baselines, and tracker audited. | Eight phase slices; builds; unit/integration/process/browser; boundary/harness/static/security/drift; retained-run finalization; final `make check`; `git diff --check` and stale-path scans. | Final check passed 136/136 work units and 1,120 tests at `20260714T000431Z-p3067373`; no root Go package, compatibility wrapper, blanket provider exception, stale root timing, or unresolved blocker remains. | None. Infrastructure-only ENOSPC attempts are attributed in Section 8. | Ready for review and commit. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Mixed application-composition package diagnosed; narrow facades and exact importer policy are required. | Inspected all production app files, `cmd/*` callers, revision providers, runtime registrars, and `tools/backend_module_boundaries.json`; touched tracker only. | Symbol/import searches; exact source reads; Make target discovery | Legitimate assembly is mixed with revision catalog and test support; broad app policy is `must_fix`. | RB-001 | Implement APP-S00 through APP-S05 in a later authorized task. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | No frontend shell/controller state or direct grid-vendor integration found in app. | Inspected app imports, embedded frontend process test, runtime assembly, and frontend-serving edge; touched tracker only. | Import searches and exact test/source reads | App observes packaged root/assets but does not own frontend state or grid behavior. | None | Recheck only if later server-facade diff touches web delivery. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Direct generated-source ownership was not found; indirect protocol/view contracts are frozen. | Inspected runtime/revision imports, contract ownership docs, generated-artifact policy, and drift targets; touched tracker only. | Import searches; `make help`; `make help-all`; relevant `make explain-target` calls | Generated files stay downstream and must not be hand-edited. | RB-001 for future owner-input changes | Run drift/policy checks for every authorized contract-adjacent slice. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Unit, integration, process, phase, and reusable test-support posture mapped. | Inspected every app test, phase/test maps, task surface, test-support inventory, and downstream test-support callers; touched tracker only. | Test-name searches; exact test reads; Make target discovery | Existing coverage is broad; composition inventory/order and revision grouping need characterization. | RB-001 | APP-S00 first; update accounting only after paths actually move. |
| 2026-07-13T18:40:02-04:00 | Codex / initial app planning | Documentation/policy checks complete; implementation suites remain intentionally unrun. | Tracker only | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check` | PASS; policy run root `20260713T223939Z-p2107210`, JSON run root `20260713T223944Z-p2107398`. | RB-001 | Future implementation establishes its baseline with the Section 8 unit/integration/process targets. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Harness gating, production rejection, auth/reconcile wiring, recovery route absence, and WebSocket authorization are contract risks. | Inspected server profiles, runtime auth wiring, recovery sentinel, runtime-route tests, and collaboration route composition; touched tracker only. | Source/test reads and route searches | Security behavior remains module/platform-owned but is vulnerable to assembly regression. | RB-001 | Preserve exact harness gate and add only composition-level characterization. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-13T18:34:44-04:00 | Codex / initial app planning | Planning is decision-complete except for the evidence-gated revision package choice; no implementation has started. | Inspected all evidence named above; touched tracker only. | Repository/source/Make discovery commands | Required order is APP-S00 through APP-S07; broad boundary exceptions cannot survive APP-S05. | RB-001 | Obtain implementation authorization, then begin with characterization and cycle proof. |
| 2026-07-13T18:40:02-04:00 | Codex / initial app planning | Tracker complete and narrowly validated; no production refactor performed. | Tracker only | Three tracker checks and working-tree status audit | All requested tracker criteria pass; only later implementation blockers remain. | RB-001 | Preserve this handoff history and append the next implementation session. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Which exact package/registration design can compose all revision providers without cycles or a new blanket owner-internal exception? | Choosing by label could move behavior across owners, create cycles, or preserve the current catch-all under a new name. | Compile/static proof for owner-exposed contribution ports and an exact application aggregator. | **CLOSED—design proven.** Owner root facades construct revisions-owned contributions; `internal/app/revisionassembly` imports only those facades, Revisions, projection adapters, and PostgreSQL. Builds, catalog validation tests, and the exact boundary allowlist pass without a blanket exception. |

## 12. Binary Completion Criteria

- [x] Every file in `internal/app` is inventoried; no file is silently out of scope.
- [x] Every discovered public contract risk has an owner and test posture.
- [x] Every proposed workflow has dependencies and a handoff/exit checkpoint.
- [x] Every proposed implementation slice is behavior-preserving; behavior-changing alternatives require later authorization.
- [x] Validation commands are discovered from the public Make task surface.
- [x] No owner contradiction was found; the required blocker wording is reserved for any later contradiction.
- [x] The framework/repository mismatch is recorded: an existing app directory is not evidence of a permanent app module.
- [x] The broad revision-provider and command-import authorization is scheduled for mandatory later replacement, not treated as desired architecture.
- [x] Handoff sections identify evidence, current blockers, and the ordered next action without requiring repository rediscovery.
- [x] RB-001 is closed with a compiled, statically enforced owner-contribution graph.
- [x] The root `internal/app` Go package and all compatibility wrappers are removed.
- [x] Exact application facades, reusable test composition, and retained `serverprocess` evidence have explicit owners.
- [x] Final narrow-to-broad validation, duration refresh, and diff audit are recorded without undisclosed failures.

The historical planning rows above remain as baseline evidence. The completed remediation state is owned by Section
1.1, the APP-006 through APP-014 rows, the implementation-session verification/handoff entries, and closed RB-001.
