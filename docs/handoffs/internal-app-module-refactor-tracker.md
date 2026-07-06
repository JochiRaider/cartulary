# internal-app Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Value |
| --- | --- |
| Target path | `internal/app` |
| Target label | `internal-app` |
| Output path | `docs/handoffs/internal-app-module-refactor-tracker.md` |
| Planning status | Planning and documentation only. This tracker records findings and future slices; it does not authorize code movement. |
| Allowed changes in this session | Create this tracker file only. |
| Non-goals | No production code, tests, contracts, generated artifacts, package configuration, migrations, harness files, or lockfiles are changed. |
| Source hierarchy | Adopted subsystem NLSpecs for named subsystems; Core 00 through Core 04; Core 05 only for claim-bearing timed or fixture-sensitive publication; `docs/domain.md` and implementation-support guides; current repository code and tests; prior plans and framework files as evidence only. |
| Implementation gate | Any implementation requires a later authorized task. Preserve observable behavior by default. |

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`
- `AGENTS.md`
- `docs/spec/00_document_set_status_and_precedence.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`
- `docs/domain.md`
- `docs/guides/cartulary-dev-guide.md`
- `docs/guides/cartulary_repository_bootstrap_guide.md`
- `docs/guides/cartulary_implementation_testing_guide.md`

Repository files inspected:

- `internal/app/runtime.go`
- `internal/app/migrate.go`
- `internal/app/operator.go`
- `internal/app/operator_migration_evidence.go`
- `internal/app/recovery_probe.go`
- `internal/app/migrate_test.go`
- `internal/app/runtime_phase0_test.go`
- `internal/app/runtime_phase0_integration_test.go`
- `internal/app/operator_test.go`
- `internal/app/operator_migration_evidence_test.go`
- `internal/app/operator_migration_evidence_integration_test.go`
- `cmd/server/main.go`
- `cmd/migrate/main.go`
- `cmd/operator/main.go`
- `cmd/server/main_phase10_recovery_sentinel_test.go`
- `cmd/operator/operator_phase10_test.go`
- `tools/phase10browserrestore/main.go`
- `internal/testutil/httptestx/httptestx.go`
- `internal/testutil/testruntime/reset_integration_test.go`
- `tools/backend_module_boundaries.json`
- `contracts/otel/import_boundary.json`
- `tools/generated_artifact_policy.json`
- `tools/phase0_test_map.json`

The target path exists. No prior tracker existed at the output path, so no prior handoff history required preservation.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/app/runtime.go` | Application runtime assembly for config validation, telemetry bootstrap, Postgres and object-store setup, bootstrap preflight, reference-data seed, job manager, WebSocket hub, auth master-key/cursor codec, module route registration, and HTTP handler construction. | `Options`, `Runtime`, `NewRuntime`, `(*Runtime).Close`; package seams `newJobsManager`, `setupPostgres`, `setupObjectStore`, `runBootstrap`, `newWSHub`, `newHTTPHandler`. | `cmd/server/main.go`; `tools/phase10browserrestore/main.go`; `internal/testutil/httptestx`; `internal/testutil/testruntime`; internal app tests. | `internal/modules/*` route registrars, workbook startup bootstrap, incident bundles, revisions attribution, `platform/config`, `httpapi`, `jobs`, `ws`, `postgres`, `objectstore`, `pagination`, `authn`, `telemetry`, `bootstrap`. | `runtime_phase0_test.go`; `runtime_phase0_integration_test.go`; callers' process and browser restore tests. | No generated root touched. Phase maps and build-input discovery mention `internal/app` as evidence/accounting only. | Keep as thin application assembly facade under `internal/app`. | High | Public behavior risk is route/dependency wiring, readiness, bootstrap failure ordering, and startup side effects. |
| `internal/app/migrate.go` | Migration CLI wrapper: parse command/args, load config, open SQL DB, call Postgres migration runner over embedded `db/migrations`, print remediation report JSON on migration remediation errors. | `RunMigrateCLI`, `RunMigrateCLIContext`; unexported `migrateRunner`, parse helpers, CLI result structs. | `cmd/migrate/main.go`; `migrate_test.go`. | `db/migrations`; `platform/config`; `platform/postgres`; standard `flag`, `database/sql`, `slog`. | `migrate_test.go`. | No generated root touched; uses authored SQL migration source. | Thin app CLI facade, with migration mechanics owned by `platform/postgres` and authored SQL inputs. | Medium | Behavior freeze includes positional command handling, `-command`, exit codes, stderr logging, and remediation JSON. |
| `internal/app/operator.go` | Deployment-local operator CLI runner for backup capture, backup metadata inspection, restore, restore verification, object-store init, object-store migration, config/object-store/Postgres runtime opening, preflight checks, artifact writing, JSON result emission, and deployment-admin authorization checks. | `RunOperatorCLIContext`, `RunOperatorCLI`, `RestoreVerificationTargetMarkerPath`; exported schema constants and JSON DTOs such as `OperatorBackupCaptureResult`, `BackupMetadataInspection`, `OperatorRestoreResult`, `OperatorRestoreVerificationResult`, `OperatorRestoreVerificationDueResult`, `OperatorObjectStoreInitResult`, and `OperatorObjectStoreMigrationResult`. | `cmd/operator/main.go`; `cmd/operator/operator_phase10_test.go`; `operator_test.go`; operator migration-evidence code via shared `operatorRunner`. | `internal/modules/recovery`; `internal/modules/projections/adapters`; `platform/authn`, `config`, `objectstore`, `postgres`; direct SQL over backup/object/restore metadata; local filesystem artifact reads/writes. | `operator_test.go`; `cmd/operator/operator_phase10_test.go`; recovery and phase10 harness rows. | No generated contract root touched. Current JSON schema IDs are handwritten constants. Phase10 evidence maps depend on behavior. | Split candidate: app keeps thin CLI composition, recovery owns backup/restore orchestration, platform owns storage/config adapters, object-store migration may remain recovery/platform support. | High | Mixed-responsibility package. Current repo behavior conflicts with Core 01/Core 04 recovery CLI authority in command grammar, result envelope, and `deployment_admin` authorization. |
| `internal/app/operator_migration_evidence.go` | Migration history evidence capture: parse manifest, audit embedded SQL migration source, inspect `goose_db_version`, emit evidence-only JSON findings. | `OperatorMigrationEvidenceResult`, `OperatorMigrationEvidenceDatabaseBinding`, `OperatorMigrationEvidenceManifestSummary`, `OperatorMigrationEvidenceSourceAudit`, `OperatorMigrationEvidenceGooseLedger`, `OperatorMigrationEvidenceGooseState`, `OperatorMigrationEvidenceFinding`. | `operator.go` command dispatch; `operator_migration_evidence_test.go`; `operator_migration_evidence_integration_test.go`. | `db/migrations`; `platform/config`; `platform/postgres`; filesystem, JSON, SHA-256, regexp helpers. | `operator_migration_evidence_test.go`; `operator_migration_evidence_integration_test.go`; phase0 coverage ledger/map rows. | No generated root touched. Reads `tools/migration_history_manifest.json` by default but does not generate it. | Candidate owner is `platform/postgres` migration evidence support or deployment operator support; defer exact owner. | Medium | Evidence-only surface is tested and should not be treated as runtime architecture just because phase maps cite it. |
| `internal/app/recovery_probe.go` | Restore-verification workbook probe: choose first incident, look up timeline view schema, execute built-in timeline workbook query, fail when incident data exists but no rows return. | `RestoreVerificationWorkbookProbe` with `ProbeRestoredBackup`. | `operator.go`; `cmd/server/main_phase10_recovery_sentinel_test.go`; `tools/phase10browserrestore/main.go`. | `internal/modules/recovery`; `internal/modules/timeline`; `platform/postgres`; `platform/viewschema`; direct SQL incident lookup. | Phase10 operator and browser restore tests through callers. | No generated root touched. Indirectly relies on view-schema registry. | Candidate owner is recovery-owned probe adapter using a timeline/workbook query port. | High | This is restore verification behavior, not application assembly. Move only with characterization around workbook query semantics. |
| `internal/app/migrate_test.go` | Unit characterization for migration CLI parsing, config/open failures, remediation report JSON, context propagation, and compatibility wrapper. | Test package surface only. | Make backend unit target; phase maps through package selection. | Test doubles for migrate runner; standard test packages. | Self. | No generated root touched. | Follow migration CLI facade or migration evidence owner after refactor. | Low | Preserve as characterization for `migrate.go` behavior. |
| `internal/app/runtime_phase0_test.go` | Unit characterization for fail-closed startup before dependency wiring on invalid config and bootstrap preflight failure. | Test package surface only. | Make backend unit target; phase0 coverage ledger/map. | Runtime seam overrides; platform config/bootstrap test fixtures. | Self. | No generated root touched. | Follow app runtime facade. | Medium | Guards startup ordering and no partial dependency startup. |
| `internal/app/runtime_phase0_integration_test.go` | Integration characterization for invalid config readiness, first-admin bootstrap, bootstrap failure cases, skip/recovery behavior, audit and DB side effects. | Test package surface only. | Make backend integration target; phase0 coverage ledger/map. | Testcontainers/Postgres/object-store fixtures, SQL assertions, runtime startup. | Self. | No generated root touched. | Follow app runtime facade and bootstrap/auth module seams. | High | These tests freeze startup and bootstrap observable behavior. |
| `internal/app/operator_test.go` | Unit characterization for object-store init JSON result and failure redaction. | Test package surface only. | Make backend unit target; operator package tests. | Fake object-store init runner and config error paths. | Self. | No generated root touched. | Follow operator CLI facade or platform object-store adapter. | Low | Narrowly covers safe output for `operator object-store init`. |
| `internal/app/operator_migration_evidence_test.go` | Unit characterization for migration-evidence input parsing, manifest/source audit, redacted JSON, and missing goose metadata evidence payload. | Test package surface only. | Make backend unit target; phase0 coverage ledger/map. | Fake postgres pool/rows, temp manifest files, embedded migration fixture use. | Self. | No generated root touched. | Follow migration evidence owner after refactor. | Medium | Preserves evidence-only and non-rewrite semantics. |
| `internal/app/operator_migration_evidence_integration_test.go` | Real Postgres integration characterization for migrated goose ledger, deployment-admin gate, protected applied history, and DB-only version findings. | Test package surface only. | Make backend integration target; phase0 coverage ledger/map. | Real Postgres migrations, auth/admin seed, migration evidence command. | Self. | No generated root touched. | Follow migration evidence owner after refactor. | Medium | Current test includes deployment-admin behavior that conflicts with later Core 04 recovery CLI boundary if reused for recovery commands. |

## 3. Module Boundary Diagnosis

`internal/app` is not a valid permanent module boundary merely because the directory exists. The current repository supports keeping `runtime.go` as a thin application/service facade, but the package also contains operator recovery behavior, migration evidence, direct storage probes, CLI parsing, direct SQL checks, and workbook restore-verification behavior that should be reviewed for clearer owners.

The current target is:

- a legitimate thin application/service facade for server runtime assembly;
- a transport-adjacent adapter for CLI entrypoints;
- a persistence-adjacent adapter for migration, backup, restore, and evidence inspection;
- a mutation coordinator for recovery and object-store migration operations;
- a view/projection orchestration layer through restore rebuild and workbook probe wiring;
- a mixed-responsibility package for operator/recovery support;
- not a frontend shell/controller surface;
- not a grid-vendor integration layer.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Server runtime assembly | `runtime.go` | `internal/app` application facade | keep | `NewRuntime` wires platform services and module route registrars. | Keep thin; avoid adding domain behavior. |
| HTTP route and dependency registration | `runtime.go` | `internal/app` plus owning modules and `platform/httpapi` | keep | `httpOptions.AdditionalRoutes` prepends module route registrars before caller routes. | Freeze route order and dependency set. |
| WebSocket/job hub wiring | `runtime.go` | `internal/app` with `platform/ws` and `platform/jobs` | keep | `newWSHub`, jobs progress hub, telemetry configuration. | Collaboration behavior remains module/platform owned. |
| Migration CLI composition | `migrate.go` | `internal/app` facade with `platform/postgres` migration implementation | defer | `RunMigrateCLIContext` wraps config load, SQL open, and `postgres.Migrate`. | Movement is optional unless package split pursues thinner app facade. |
| Backup and restore operator orchestration | `operator.go` | `internal/modules/recovery` with `internal/app` CLI facade | split | Calls recovery capture, restore, verify, catalog, backup storage, and object-store migration services. | Preserve current behavior until Core mismatch is authorized for change. |
| Operator config/storage adapters | `operator.go` | `internal/platform/config`, `platform/postgres`, `platform/objectstore`, recovery adapter ports | split | Opens source/target config, Postgres, object stores, backup storage. | Preflight policy should not be scattered in app CLI wrapper. |
| Deployment-admin authorization for operator recovery | `operator.go`, `operator_migration_evidence.go` | TODO: owner reconciliation required | defer | `authorizeDeploymentAdmin` and direct `authn.NewStore(...).GetUserByNormalizedEmail`; Core 04 says recovery CLI is local-operator authority, not `deployment_admin`. | Behavior change requires later authorization. |
| Restore verification workbook probe | `recovery_probe.go` | `internal/modules/recovery` facade with timeline/workbook query port | split | Direct incident SQL, timeline view-schema lookup, `timeline.NewFacade(...).QueryTimelineRows`. | Correct capability belongs to recovery verification, not app assembly. |
| Migration history evidence | `operator_migration_evidence.go` | `platform/postgres` migration evidence support or deployment operator support | defer | Reads manifest, embedded SQL migrations, and goose ledger. | Evidence-only; do not infer runtime module from phase maps. |
| Test characterization | `internal/app/*_test.go` | Follow the owning production seam after movement | defer | Unit/integration tests cover runtime, migrate, operator, migration evidence. | Test package location can change only after behavior owner is chosen. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| HTTP runtime route composition and readiness | `internal/app` facade plus `platform/httpapi` and route-owning modules | `runtime.go` registers auth, incidents, extensions, jobs, imports, reporting, reference data, incident bundles, saved views, view schemas, collaboration, entities, evidence, assessments, workbook, timeline, and revisions. | `runtime_phase0_test.go`; `runtime_phase0_integration_test.go`; caller process tests. | Add or retain tests that freeze route registrar order, dependency propagation, caller `AdditionalRoutes`, `/readyz` readiness behavior, and startup failure ordering before moving runtime wiring. | High | No route shape change is authorized by this tracker. |
| WebSocket hub and job progress wiring | `internal/app` facade with `platform/ws`, `platform/jobs`, collaboration module | `runtime.go` creates hub, configures jobs progress hub, and passes hub through dependency set. | Runtime phase0 tests verify no startup on failures; collaboration tests outside this inventory may cover event behavior. | Characterize hub/dependency preservation if moving runtime wiring. | Medium | WebSocket paths and event semantics are observable even though handlers live elsewhere. |
| Workbook row/query/mutation, saved-view, and view-schema route availability | Owning modules under `internal/modules/*`; app only wires registrars | `runtime.go` route registration includes `savedviews`, `viewschemas`, `workbook`, `timeline`, and `revisions`. | Module tests outside `internal/app`; runtime startup tests. | Before changing app wiring, run module route characterization or add an assembly test that proves these route families remain mounted. | High | App must not become owner of workbook semantics. |
| Projection refresh and restore rebuild | `internal/modules/recovery` and projection modules | `operator.go` uses `projectionadapters.NewRestoreRebuilder(targetPool)` during restore and verification. | `cmd/operator/operator_phase10_test.go`; recovery tests outside this package. | Characterize restore rebuild invocation through recovery-owned adapter before moving. | High | Core 01 says recovery owns the restore projection rebuild adapter contract. |
| Migration CLI argument and remediation behavior | `internal/app/migrate.go` facade and `platform/postgres` migration runner | `RunMigrateCLIContext`, `parseMigrateCLIArgs`, `postgres.MigrationRemediationError` handling. | `migrate_test.go`. | Existing tests are sufficient before pure package movement; add process-level coverage only if `cmd/migrate` wiring changes. | Medium | Preserve stdout/stderr and exit code behavior. |
| Operator CLI command grammar and JSON output | Current implementation in `operator.go` | Commands include `backup capture`, `backup-metadata latest`, `migration-evidence capture`, `restore latest`, `restore-verify latest`, `restore-verify due`, `object-store init`, and `object-store-migration run`; handwritten schema constants emit current JSON shapes. | `operator_test.go`; `operator_migration_evidence_test.go`; `operator_migration_evidence_integration_test.go`; `cmd/operator/operator_phase10_test.go`. | Add characterization for every command before movement if existing process tests do not cover a command. | High | Core 01 command/result requirements differ; changing them requires later authorization. |
| Recovery CLI authorization outcome | Core 04 owns target behavior; current repo implements `deployment_admin` gates | `authorizeDeploymentAdmin` and direct authn checks in `operator.go` and migration evidence command. | `cmd/operator/operator_phase10_test.go`; `operator_migration_evidence_integration_test.go`; phase ledgers. | Record current tests as characterization only; do not treat them as owner authority. | High | `RB-001` blocks behavior-correcting implementation. |
| Backup, restore, object-store migration storage semantics | Recovery module plus operator adapter code | `operator.go` captures artifacts, reads object blob index, opens backup storage, preflights source/target config, writes migration artifacts. | `cmd/operator/operator_phase10_test.go`; recovery module tests outside this file. | Characterize target preflight, redaction, artifact paths, and no-mutation-before-preflight before movement. | High | Direct SQL/storage coupling in app should move behind recovery/platform seams. |
| Restore verification workbook probe | Recovery owner candidate; currently `internal/app` | `recovery_probe.go` finds first incident and queries timeline rows with default query meta. | Phase10 restore/browser tests through callers. | Add a focused characterization test for zero-incident and incident-with-timeline cases before moving. | High | Built-in workbook query semantics are observable. |
| Generated protocol and view contracts | `contracts/*`, generated roots, and owner specs | `internal/app` imports no generated roots; generated policy forbids hand edits. | Drift checks; module tests outside target. | Run generated drift checks if any future move changes generated inputs or route contracts. | Medium | This tracker does not touch generated files. |
| Harness/test accounting | `docs/testing-harness-nlspec.md`, phase maps, ledgers, Make targets | `tools/phase0_test_map.json`, phase10 task guide, coverage ledgers cite `internal/app` tests as evidence. | Existing Make targets. | Update owner inputs only if future test locations or evidence rows change. | Medium | Phase maps are evidence accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `runtime.go` imports both platform services and module route registrars. | `NewRuntime` constructs dependency set and app route list. | App facade could accumulate domain logic. | intentional/no_action | `internal/app` facade | Keep as assembly only; reject new domain behavior in runtime wiring. |
| Operator recovery orchestration lives in app package. | `operator.go` performs backup capture, restore, restore verification, object-store migration, artifact writing, and preflight. | Mixed responsibility makes recovery behavior hard to characterize and move safely. | should_fix | `internal/modules/recovery` plus app CLI facade | Plan behavior-preserving extraction after characterization. |
| Recovery CLI currently uses `deployment_admin` authorization. | `authorizeDeploymentAdmin`; `-deployment-admin-email` flags; phase10/operator tests. | Current repo behavior conflicts with Core 04 local-operator authority. | must_fix | TODO: owner decision required | `RB-001`; do not change behavior without explicit authorization. |
| Operator command grammar and result schemas differ from Core 01. | Current commands and schema IDs in `operator.go`; Core 01 requires logical commands and `cartulary.operator_recovery_result.v1`. | Behavior change would affect operator automation and evidence. | must_fix | Core 01 recovery CLI owner with app compatibility facade | `RB-002`; separate compatibility/refactor from behavior-correcting task. |
| Direct SQL/storage checks are embedded in operator app code. | Blob index query, goose schema version query, target row count query, advisory locks, object-store list checks. | App layer becomes persistence policy owner. | should_fix | Recovery service and platform storage adapters | Extract behind interfaces only after freezing current semantics. |
| Restore verification workbook query is implemented in `internal/app`. | `RestoreVerificationWorkbookProbe` directly uses timeline facade and view-schema lookup. | Workbook/timeline semantics can drift during recovery refactor. | should_fix | Recovery-owned probe port; timeline/workbook module implementation | Add focused characterization before movement. |
| Migration evidence mixes deployment CLI, migration manifest audit, embedded source audit, and goose ledger SQL. | `operator_migration_evidence.go`. | Evidence-only behavior may be mistaken for runtime architecture. | defer | `platform/postgres` migration evidence or operator support | Decide owner after recovery CLI mismatch is resolved. |
| Generated roots are not directly imported or edited by target. | Search found no target writes to generated roots; generated policy lists protected roots. | Low current risk, high future hand-edit risk. | intentional/no_action | Generated artifact owners | Continue using Make generators and drift checks only. |
| No frontend shell/controller or grid-vendor coupling was found in `internal/app`. | Searches found no `apps/web`, `packages/grid-adapter`, or vendor imports in target. | Low | intentional/no_action | Not applicable | Do not create frontend workstreams for this target unless future evidence appears. |
| Phase maps and ledgers cite `internal/app` tests. | `tools/phase0_test_map.json`; phase task guides; coverage ledgers. | Evidence accounting could be mistaken for architecture. | intentional/no_action | Harness/accounting owners | Treat maps as verification routing only. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Establish target, source hierarchy, branch/commit posture, and write tracker only. | `docs/handoffs/internal-app-module-refactor-tracker.md`; framework and owner docs. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`. | Tracker has scope/source posture and session log. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every `internal/app` file, callers, tests, dependencies, and generated-contract posture. | `internal/app/*`; `cmd/*`; testutil callers; tools maps. | Search plus tracker review; no production edits. | Inventory table complete. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-05 | Map observable contracts and owner docs before any refactor slice. | Owner docs, runtime/operator/migrate/recovery probe files. | `make task-guide ROLE=feature-dev PHASE=phase0`; `make task-guide ROLE=feature-dev PHASE=phase10`. | Contract freeze map complete. |
| WF-03 | Characterization test gap analysis | chain | WF-02 | WF-05, WF-06 | Decide which current tests are sufficient and where pre-move tests are needed. | `internal/app/*_test.go`; `cmd/operator/*_test.go`; recovery tests. | `make backend-unit`; `make backend-integration`; phase slices as needed. | Missing characterization is listed before code movement. |
| WF-04 | Boundary/coupling scan | chain | WF-01 | WF-05, WF-06 | Identify app/domain/platform/recovery coupling and generated-file risks. | `internal/app/*.go`; boundary manifests. | `make backend-module-boundary-check`. | Findings classified as `must_fix`, `should_fix`, `defer`, or `intentional/no_action`. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06, WF-07 | Define behavior-preserving target seams for runtime, migrate, operator, migration evidence, and recovery probe. | Future task may touch `internal/app`, `internal/modules/recovery`, and platform packages. | No code validation until implementation task. | Design plan separates behavior-preserving moves from `requires later authorization` changes. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07, WF-08 | Convert ownership plan into smallest safe implementation slices. | Tracker and future implementation files. | Per-slice commands named in Section 7. | Each slice has rollback and completion criteria. |
| WF-07 | Harness/test/accounting update plan | parallel | WF-03, WF-06 | WF-08 | Determine whether moving tests or packages requires phase map owner-input updates. | Test files, phase maps, ledgers, Make targets. | `make generated-artifact-policy-check`; `make json-shape-check`; phase-slice commands. | Harness changes are owner-input first, not generated-file hand edits. |
| WF-08 | Validation and final handoff | chain | WF-06, WF-07 | none | Run narrow checks for tracker or later implementation and record handoff. | Tracker; future changed code/tests. | Tracker: docs/static checks. Implementation: phase and full gates as risk requires. | Session log current enough to resume without rediscovery. |

## 7. Proposed Refactor Slice Plan

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | none | Create this tracker only. | `docs/handoffs/internal-app-module-refactor-tracker.md`. | Documentation may misstate current behavior. | Not applicable. | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`. | Revert tracker file only. | Tracker has all required sections and no production refactor. |
| SL-01 | SL-00 | Preserve and characterize runtime app facade before any movement. | `internal/app/runtime.go`; runtime tests; possible assembly tests. | HTTP route registration, readiness, bootstrap ordering, WebSocket/job dependencies. | Preserve phase0 runtime tests; add route/dependency assembly characterization if moving wiring. | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase0`. | Revert runtime-only changes. | `NewRuntime` remains behaviorally identical and app stays assembly-only. |
| SL-02 | SL-01 | Optionally isolate migration CLI internals while preserving public wrappers. | `internal/app/migrate.go`; possible `platform/postgres` helper; `cmd/migrate`. | CLI args, stderr, remediation JSON, exit codes. | Preserve `migrate_test.go`; add process-level test only if `cmd/migrate` changes. | `make backend-unit`; `make phase-slice PHASE=phase0`. | Revert migration facade move. | `RunMigrateCLIContext` behavior unchanged. |
| SL-03 | SL-00, SL-03 characterization from WF-03 | Extract operator recovery implementation behind a facade while preserving current command grammar and JSON. | `internal/app/operator.go`; `internal/modules/recovery`; platform adapters; `cmd/operator` tests. | Operator automation, backup/restore semantics, JSON schemas, redaction, authorization outcomes. | Preserve operator unit/integration/process tests; add missing command characterization first. | `make backend-unit`; `make backend-store`; `make backend-process`; `make phase-slice PHASE=phase10`. | Revert facade extraction before changing behavior. | App exposes thin CLI wrapper; current observable operator behavior unchanged. |
| SL-04 | SL-03 | Move restore-verification workbook probe behind recovery-owned port while preserving timeline query behavior. | `internal/app/recovery_probe.go`; `internal/modules/recovery`; `internal/modules/timeline`; callers. | Restore verification pass/fail, zero-incident handling, view-schema lookup, built-in workbook query behavior. | Add focused probe tests for zero incident and incident timeline query before movement. | `make backend-unit`; `make backend-integration`; `make service-backed-slice PHASE=phase10`. | Revert probe port and caller wiring. | Recovery owns probe interface; current query semantics unchanged. |
| SL-05 | SL-03 | Move migration-evidence helper to a clearer migration/operator evidence owner while preserving JSON. | `internal/app/operator_migration_evidence.go`; possible `platform/postgres` evidence package; tests. | Evidence JSON shape, finding order, manifest/source/goose checks, deployment-admin current behavior. | Preserve unit and integration migration-evidence tests. | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase0`. | Revert package movement and imports. | Same evidence payload and command behavior under a clearer owner. |
| SL-06 | RB-001, RB-002 resolved | Reconcile recovery CLI with Core 01/Core 04 command, output, and authorization requirements. | `internal/app/operator.go`; `cmd/operator`; recovery modules; tests and owner docs if changed. | Behavior change: command grammar, result envelope, exit codes, local-operator authorization, journaling, progress. | New owner-aligned characterization and conformance tests required. | `make phase-slice PHASE=phase10`; `make service-backed-slice PHASE=phase10`; `make browser-e2e-webserver-backed`; `make test-fast`; possibly `make check`. | Requires later authorization; rollback all behavior changes together. | Core mismatch resolved with owner-approved behavior and passing evidence. |
| SL-07 | SL-03 to SL-05 | Add or tighten boundary guardrails after movement. | Boundary manifests and owner inputs, not generated outputs by hand. | Import-boundary drift or generated-file hand-edit risk. | Preserve moved tests and add boundary tests only if a new rule is introduced. | `make backend-module-boundary-check`; `make generated-artifact-policy-check`; `make json-shape-check`. | Revert boundary metadata if it over-constrains valid imports. | No new domain/platform leak; generated roots untouched. |

Any slice that changes command grammar, JSON envelopes, exit codes, authorization outcomes, route shape, WebSocket semantics, workbook query behavior, storage semantics, generated contracts, or harness accounting is `requires later authorization`.

## 8. Validation Plan

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit | `make backend-unit` | Runtime, migration CLI, operator unit tests, migration-evidence unit tests. | yes | Use before and after moving app package logic. |
| integration | `make backend-integration`; `make phase-slice PHASE=phase0`; `make phase-slice PHASE=phase10`; `make service-backed-slice PHASE=phase10` | Startup/bootstrap, migration evidence with Postgres, recovery/operator service-backed behavior. | yes | Phase slices were discovered with `make task-guide ROLE=feature-dev PHASE=phase0` and `PHASE=phase10`. |
| e2e/browser | `make browser-e2e-webserver-backed` | Phase10 restore/browser readiness and public route absence when recovery behavior touches browser-visible readiness. | no for tracker; yes for recovery/browser restore implementation | Do not run for documentation-only tracker unless broader gate is requested. |
| generated drift | `make generated-artifact-policy-check`; `make json-shape-check` | Generated roots and JSON manifests. | yes for tracker and implementation | Tracker does not edit generated roots. |
| import-boundary/static | `make backend-module-boundary-check`; `make lint-markdown` | Backend import boundaries and markdown style. | yes | `backend-module-boundary-check` is needed when package ownership or imports change; `lint-markdown` covers this tracker. |
| full check | `make test-fast`; `make check` | Aggregate local verification gate. | no for tracker; yes when high-risk implementation slices complete | Run `make agent-finalize` before broader end-of-run verification, per repository procedure. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| IA-001 | Define `internal-app` target, source posture, and write limits. | WF-00 | DONE | none | Section 1 | Target path, label, allowed write, and non-goals are explicit. |
| IA-002 | Inventory all files under `internal/app`. | WF-01 | DONE | IA-001 | Section 2 | Every target file is inventoried. |
| IA-003 | Map public contract and behavior freeze risks. | WF-02 | DONE | IA-002 | Section 4 | Contract table names owner, evidence, tests, and risk. |
| IA-004 | Classify module-boundary posture. | WF-04 | DONE | IA-002 | Sections 3 and 5 | Findings distinguish keep, split, defer, and no-action cases. |
| IA-005 | Record recovery CLI authority mismatches. | WF-02 | DONE | IA-003 | RB-001, RB-002 | Mismatches are blockers, not implementation decisions. |
| IA-006 | Add focused characterization plan for runtime and recovery probe. | WF-03 | TODO | IA-003 | Section 7 | Missing tests are identified before future movement. |
| IA-007 | Design behavior-preserving operator/recovery facade extraction. | WF-05 | BLOCKED | IA-005, IA-006 | RB-001, RB-002 | Owner-approved path separates compatibility from behavior correction. |
| IA-008 | Plan migration evidence owner move. | WF-05 | DEFERRED | IA-006 | SL-05 | Owner candidate is chosen after recovery CLI mismatch is resolved. |
| IA-009 | Confirm harness/accounting updates for any future moved tests. | WF-07 | TODO | IA-006 | Section 8 | Owner inputs are identified before generated artifacts are refreshed. |
| IA-010 | Execute implementation slices. | WF-06 | TODO | IA-006, IA-007 | Later authorized task | Behavior-preserving slices land with named validation. |
| IA-011 | Final implementation validation and handoff. | WF-08 | TODO | IA-010 | Later authorized task | Commands run, results, blockers, and remaining risks are recorded. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | Tracker created for `internal/app`; no production refactor authorized. | Inspected framework, AGENTS, domain, Core 00/Core 01/Core 04, harness docs, guides, target files, callers, and tool manifests. Touched only this tracker. | `sed`, `rg`, `git status`, `git rev-parse`, `date`, `make help`, `make task-guide ROLE=feature-dev PHASE=phase0`, `make task-guide ROLE=feature-dev PHASE=phase10`. | Source posture established. | None for tracker creation. | Run tracker validation commands and preserve file-only diff. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | `runtime.go` is legitimate app assembly; operator, migration evidence, and recovery probe are mixed-responsibility candidates. | Inspected `internal/app/*.go` and backend callers. Touched only this tracker. | `rg -n "^(func|type|const|var)\\b" internal/app/*.go`; targeted `sed` reads. | Boundary findings recorded. | RB-001 and RB-002 block behavior-correcting recovery work. | Add characterization before future code movement. |

### Frontend module boundary, if applicable

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | No frontend shell/controller or grid-vendor integration found in `internal/app`. | Inspected imports and caller search results. Touched only this tracker. | `rg` over target, callers, tools, contracts, and docs. | Frontend workstream not applicable for this target. | None. | Reopen only if future evidence shows frontend coupling. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | No generated roots are owned or touched by `internal/app`; operator JSON contracts are handwritten constants. | Inspected generated policy, import boundary, phase map, target code. Touched only this tracker. | `rg` contract/generated searches; `sed` over policy files. | Codegen risk recorded as indirect drift risk. | RB-002 for recovery CLI result schema mismatch. | Use Make drift checks after tracker and any future generated-input change. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | Existing tests cover runtime startup, migration CLI, operator object-store init, migration evidence, and phase10 operator behavior; some characterization gaps remain for package movement. | Inspected `internal/app/*_test.go`, `cmd/operator/operator_phase10_test.go`, phase maps, task-guide output. Touched only this tracker. | `rg --files internal/app`; `make task-guide ROLE=feature-dev PHASE=phase0`; `make task-guide ROLE=feature-dev PHASE=phase10`. | Validation plan recorded. | Missing focused probe and assembly characterization before future movement. | Add tests in later implementation task only. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | Current repo requires `deployment_admin` email for several operator commands; Core 04 says recovery CLI invocation must be local operator authority and not `deployment_admin`. | Inspected `operator.go`, `operator_migration_evidence.go`, Core 04, phase10 tests and guides. Touched only this tracker. | `rg -n "deployment_admin|operator backup|operator restore|restore-verify|operator_recovery|local operator" ...`; targeted `sed` reads. | Repo-vs-owner mismatch recorded. | RB-001. | Obtain explicit owner-aligned authorization before changing behavior. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-05T23:42:47Z | Codex planning/documentation session | Tracker is the only intended diff. Future code work must split behavior-preserving movement from behavior-correcting recovery CLI changes. | Touched only this tracker. | Pending validation after write. | Ready for narrow docs/static validation. | RB-001, RB-002, RB-003, RB-004. | Run validation, then hand off with results. |

## 11. Open Questions and Blockers

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Repository/tests require `deployment_admin` authorization for operator recovery flows, while Core 04 REQ-04-106 says recovery CLI invocation is authorized by local operator authority and must not be authorized by `deployment_admin`. | Changing this alters security behavior and operator automation. | Owner-approved recovery CLI authorization task; updated tests and conformance mapping. | BLOCKED for behavior change; current behavior preserved by default. |
| RB-002 | Repository operator command grammar and result schemas differ from Core 01 REQ-01-593 and REQ-01-594. | Changing command names, flags, stdout JSON, progress, and exit mappings is an observable contract change. | Owner-approved compatibility/transition plan or explicit behavior-correction task. | BLOCKED for behavior change; current behavior preserved by default. |
| RB-003 | Exact long-term owner for migration-evidence capture is not decided. | Moving evidence helpers too early could mix migration infrastructure, operator support, and phase accounting. | Owner mapping after recovery CLI scope is settled. | TODO: owner decision required before movement. |
| RB-004 | Exact recovery-owned port shape for restore verification workbook probe is not defined. | Probe movement could change timeline/view-schema query behavior or misplace workbook semantics. | Focused characterization tests and recovery facade design. | TODO: required before SL-04. |
| RB-005 | Harness/accounting updates for future moved tests are unknown until implementation slice paths are chosen. | Hand-editing generated phase ledgers or maps is prohibited; owner inputs must be updated first. | Future slice diff and Make-owned generation/drift guidance. | TODO: command/input discovery required during implementation. |

## 12. Binary Completion Criteria

| Criterion | Status | Evidence |
| --- | --- | --- |
| Every file in `internal/app` is inventoried or explicitly out of scope. | PASS | Section 2 lists all files from `rg --files internal/app`. |
| Every discovered public contract risk has an owner and test posture. | PASS | Section 4 maps contracts, owners, evidence, existing tests, and required characterization. |
| Every proposed workflow has dependencies and exit criteria. | PASS | Section 6 lists dependencies, validation, and handoff checkpoints. |
| Every proposed implementation slice is behavior-preserving unless explicitly marked `requires later authorization`. | PASS | Section 7 marks SL-06 and any behavior-changing contract edits as requiring later authorization. |
| Validation commands are discovered or marked `TODO` with a reason. | PASS | Section 8 lists Make-owned commands and notes when commands are implementation-dependent. |
| Contradictions are marked `BLOCKED: owner contradiction`. | PASS | No owner-document contradiction was found. Repo-vs-owner mismatches are recorded as RB-001 and RB-002 instead. |
| Repository/framework mismatches are recorded as planning findings. | PASS | Sections 3 and 5 record that `internal/app` is only partly a legitimate app facade and is not automatically a permanent module boundary. |
| Handoff sections are current enough for another agent to continue without rediscovery. | PASS | Section 10 records inspected/touched files, commands, results, blockers, and next actions by workstream. |

