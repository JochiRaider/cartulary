# postgres Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/platform/postgres`
- **Target label:** `postgres`, derived from the final path segment and already valid lowercase kebab case.
- **Output path:** `docs/handoffs/postgres-module-refactor-tracker.md`
- **Repository snapshot:** branch `main`, commit `d526d0ad`, inspected on 2026-08-07.
- **Status:** planning and documentation only.
- **Allowed change in this session:** this tracker file only.
- **Non-goals:** no production or test refactor, SQL or migration change, generated artifact edit, package/configuration change, harness edit, contract edit, or behavior change.
- **Implementation authorization:** every implementation slice in this tracker requires a later, separately authorized task.

Normative terms in this tracker use the meanings below:

| Term | Meaning in this tracker |
| --- | --- |
| MUST / MUST NOT | A binary condition for the later refactor. An implementation that violates it is incomplete. |
| SHOULD / SHOULD NOT | The required default. A deviation requires an owner-approved amendment to this tracker before implementation. |
| MAY | Intentional implementation freedom that does not alter an observable contract. |
| Current | A fact inspected at commit `d526d0ad7ed531b296b026e91a54cd8bb39302cf`. |
| Final | The required post-refactor state after all implementation gates and acceptance criteria pass. |

The source hierarchy used here is: adopted subsystem NLSpecs for their named scopes; Core 00 through Core 04 for implementation conformance; Core 05 only for claim-bearing timed or fixture-sensitive publication; domain vocabulary and implementation-support guides; live code and tests; and prior trackers or handoffs as evidence only. No owner contradiction was found. If later inspection finds one, the affected item must be recorded as `BLOCKED: owner contradiction` without choosing a side.

Owner documents inspected:

- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, read first as required, as planning doctrine rather than repository-state evidence.
- `docs/spec/00_document_set_status_and_precedence.md` for precedence.
- `docs/spec/01_architecture_storage_and_view_contracts.md` for the modular monolith, PostgreSQL authority, migration lineage, schema ownership, recovery, and runtime assembly.
- `docs/spec/04_security_deployment_and_conformance.md` for database-root binding, managed-service environment resolution, safe path handling, diagnostics, and startup readiness.
- `docs/opentelemetry-instrumentation-nlspec.md` for the PostgreSQL telemetry scope, operation span, duration metric, error classes, and forbidden attributes.
- `docs/testing-harness-nlspec.md` for Make ownership, verification rows, execution selection, and evidence accounting.
- Relevant vocabulary and owner navigation in `docs/domain.md`; it does not make physical package names or migrations into domain modules.
- `docs/research/nlspec-spec.md` for specification-writing discipline: behavioral completeness, unambiguous interfaces, explicit defaults, mapping tables, and binary acceptance criteria. It is writing doctrine, not a product-contract owner.
- `temp/analysis-notes.md` as approved planning input for RB-001 and RB-002. It is evidence of the selected approach, not proof that any proposed Core, package, registry, or harness change already exists.

Core 02, Core 03, and Core 05 do not directly govern the package surfaces discovered: the target contains no domain model implementation, interaction protocol, or claim-bearing benchmark publication. This is an applicability finding, not a reduction of their authority elsewhere.

Repository files inspected include every entry under `internal/platform/postgres`, the representative callers in `internal/app/server`, `internal/app/migrate`, `internal/app/operator`, `internal/app/recoveryassembly`, and `internal/modules/networkflow`; `db/migrations/source.go`; the live migration and schema-ownership manifests; recovery catalog fixtures; generated-artifact policy; backend-boundary inputs; test-support inventory; and the `platform.postgres` verification owner and family files. Searches found 154 Go files importing the target package; the inventory therefore names representative inbound callers rather than repeating every store and test utility.

The live repository contains 57 numbered SQL migrations. Archived migration trackers describing 23 or 40 migrations are stale historical evidence and are not used as current inventory. Where framework or archived state differs from the repository, this tracker follows the live repository and records the mismatch. If the source snapshot changes before implementation, the implementer MUST rediscover Core identifier availability, owner-ID availability, row-ID collisions, exact test symbols, selector coverage, migration count, and affected caller paths before changing any authored input.

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Final owner or disposition | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/platform/postgres/.gitkeep` | Empty-directory sentinel retained from an earlier package shape. | None. | None. | None. | None. | None. | None. | low | Explicitly out of behavioral scope; inventory only. |
| `internal/platform/postgres/db.go` | Narrow PostgreSQL operation port used by stores and wrappers. | `DB` with `Exec`, `Query`, `QueryRow`, and `BeginTx`. | Server assembly, module stores, auth storage, telemetry, migration evidence, and test utilities. | `pgx`, `pgconn`. | Exercised through store, transaction, telemetry, and integration tests across the repository. | No generated surface; indirectly underpins storage contracts. | `platform.postgres`. | high | Legitimate shared persistence-adapter surface with broad blast radius. |
| `internal/platform/postgres/entity_alias_migration_test.go` | Characterizes migration 31 alias normalization, tombstoning, and invalid legacy-row preflight. | Three integration tests. | Harness selectors and direct package testing. | Embedded migrations, PostgreSQL fixture helpers, entity tables. | This file is the test. | Migration 31 and schema-owner/test-family projections. | `module.entities` test ownership. | medium | Source-owner behavior is tested from the platform package; one current row is attributed to `platform.postgres`. |
| `internal/platform/postgres/evidence_blob_uniqueness_migration_test.go` | Characterizes migration 53 evidence-blob uniqueness preflight and enforcement. | One integration test. | `module.evidence` verification family. | Embedded migrations, PostgreSQL fixtures, evidence storage schema. | This file is the test. | Migration 53 and evidence owner mappings. | `module.evidence` test ownership. | medium | Test placement does not make evidence behavior a PostgreSQL module responsibility. |
| `internal/platform/postgres/extension_job_cutover_migration_test.go` | Characterizes migration 34 extension-job cutover and rejection of retired handlers. | Two integration tests. | `module.extensions` verification family; backend boundary exception input. | Embedded migrations, PostgreSQL fixtures, extension job schema. | This file is the test. | Migration 34, extension test-family row, and authored boundary rule. | `module.extensions` test ownership. | high | Relocation must update the authored retired-handler boundary input atomically. |
| `internal/platform/postgres/graph_projection_migration_test.go` | Characterizes migration 32 reset of unreferenced derived graph state and rejection of referenced state. | Two tests. | `module.graphprojection` verification family. | Embedded migrations, PostgreSQL fixtures, graph tables. | This file is the test. | Migration 32 and graph-projection test-family row. | `module.graphprojection` test ownership. | medium | Tests migration effects only; no projection refresh runtime is implemented here. |
| `internal/platform/postgres/incident_bundle_storage_reference_migration_test.go` | Characterizes migration 37 storage-reference cutover and populated-table preflight. | Two integration tests. | `module.incidentbundles` verification family. | Embedded migrations, PostgreSQL fixtures, incident-bundle schema. | This file is the test. | Migration 37 and incident-bundle test-family row. | `module.incidentbundles` test ownership. | medium | Storage-reference contract belongs to the source owner. |
| `internal/platform/postgres/migration_remediation.go` | Inspects migration lineage and emits typed remediation reports for incompatible histories. | `MigrationRemediationReport`, `MigrationRemediationFinding`, `MigrationRemediationError`, `Error`, `ReportJSON`. | `Migrate`, schema readiness, migrate CLI, and operator/server tests. | `database/sql`, JSON, filesystem inspection, migration metadata tables. | `migration_remediation_test.go`, `schema_readiness_test.go`, migrate and server tests. | `cartulary.migration_remediation_report.v1`; migration history manifest and lineage schema. | Final owner `module.database_migrations`; planned root `internal/modules/database_migrations`. | high | Observable JSON/error behavior must remain byte- and field-compatible. |
| `internal/platform/postgres/migration_remediation_test.go` | Characterizes accepted current lineage and typed rejection of historical or wrong lineage. | Three tests. | Direct package testing; not found in current `platform.postgres` family selectors. | Remediation internals and PostgreSQL migration fixtures. | This file is the test. | Remediation report schema behavior. | Final owner `module.database_migrations`. | medium | Verification accounting must be explicit before relocation. |
| `internal/platform/postgres/migrationevidence/evidence.go` | Audits the authored migration manifest, embedded source bytes, naming/markers/checksums, and live Goose ledger. | `Result`, `DatabaseBinding`, `ManifestSummary`, `SourceAudit`, `GooseLedger`, `GooseState`, `Finding`, `Build`. | Operator migration-evidence command and its tests. | `postgres.DB`, filesystem/JSON/checksum logic, `goose_db_version`. | Package unit test and operator unit/integration tests. | `cartulary.migration_history_evidence.v1`; `tools/migration_history_manifest.json`. | Final owner `module.database_migrations`; planned evidence subpackage. | high | Evidence is transport-neutral but currently nested under the connection adapter. |
| `internal/platform/postgres/migrationevidence/evidence_test.go` | Characterizes manifest/source-audit findings. | One test. | Direct package testing; not found in the current platform family selectors. | Migration-evidence builder and test filesystem. | This file is the test. | Migration-history evidence schema. | Final owner `module.database_migrations`. | medium | Add explicit verification ownership before moving. |
| `internal/platform/postgres/postgres.go` | Combines database binding/connection setup with migration source discovery and Goose execution. | DSN constants; `Settings`, `Binding`, `MigrationSource`, `MigrationStatus`; source constructors; `ResolveSettings`, `Setup`, `OpenSQL`, `EnvKeyForServiceRef`, `Migrate`. | Config/server/migrate/operator assembly, embedded migrations, many integration tests, `pgtest`, and recovery tooling. | `pgxpool`, `database/sql`, Goose, filesystem, environment, `rootedfs`, synchronization. | Settings, bootstrap, cancellation, context, support, readiness, remediation, operator, and service-backed tests. | Migration manifest/source, Core 04 root binding, and Goose lineage behavior. | Split: connection pieces remain `platform.postgres`; migration pieces move to `module.database_migrations`. | high | Primary mixed-responsibility file; global Goose BaseFS/logger guards are migration mechanics, not domain logic. |
| `internal/platform/postgres/postgres_context_test.go` | Characterizes canceled-context handling and guarded Goose context/BaseFS behavior. | Four tests. | Direct package testing; not found in current family selectors. | Goose seams and migration internals. | This file is the test. | Migration execution semantics. | Final owner `module.database_migrations`. | high | Concurrency/cancellation behavior is observable and needs retained coverage. |
| `internal/platform/postgres/postgres_settings_test.go` | Characterizes filesystem-root and managed-service DSN resolution and unsafe-path rejection. | `TestPostgresRootBindingResolution`. | `platform.postgres` root-DSN verification row. | `rootedfs`, environment bindings, temporary filesystem. | This file is the test. | Core 04 database-root binding rules. | `platform.postgres`. | high | Security-sensitive path and secret-resolution behavior remains with the adapter. |
| `internal/platform/postgres/postgres_support_test.go` | Guards idempotent, lineage-safe statements in the bootstrap migration. | `TestSchemaBootstrapMigrationGuard`. | `platform.postgres` schema-bootstrap-guard row. | Authored migration 00001 text. | This file is the test. | Bootstrap migration and lineage contract. | Final owner `module.database_migrations`. | medium | Text guard is migration mechanics, not proof of schema completeness. |
| `internal/platform/postgres/postgres_test.go` | Exercises fresh schema bootstrap and long-running migration cancellation. | Two integration tests. | `platform.postgres` schema-bootstrap row; direct package testing. | PostgreSQL fixtures, Goose, embedded migrations. | This file is the test. | Migration source/manifest and bootstrap contract. | Final owner `module.database_migrations`. | high | The cancellation test was not found in the current exact family selectors. |
| `internal/platform/postgres/recovery_state.go` | Contributes migration-lineage metadata to the recovery-state catalog. | `RecoveryStateContribution`. | `internal/app/recoveryassembly`. | `internal/platform/recoverystate`. | Recovery catalog assembly/fixture tests. | Recovery-state catalog fixture and `schema_migration_lineage` ownership. | Final owner `module.database_migrations`. | medium | Current contribution owner is already `module.database_migrations`. |
| `internal/platform/postgres/reference_pack_storage_reference_migration_test.go` | Characterizes migration 38 reference-pack storage-reference cutover and preflight. | Two integration tests. | `module.reference_data` verification family. | Embedded migrations, PostgreSQL fixtures, reference-pack schema. | This file is the test. | Migration 38 and reference-data test-family row. | `module.reference_data` test ownership. | medium | Source-owner contract test is physically under platform PostgreSQL. |
| `internal/platform/postgres/saved_views_storage_hardening_migration_test.go` | Characterizes migration 52 saved-view hardening and count-only preflight reporting. | Two integration tests. | `module.savedviews` verification family. | Embedded migrations, PostgreSQL fixtures, saved-view schema. | This file is the test. | Migration 52 and saved-view test-family row. | `module.savedviews` test ownership. | high | Storage behavior is affected; no saved-view runtime or view-schema implementation lives here. |
| `internal/platform/postgres/schema_readiness.go` | Blocks server startup when schema metadata, head version, or lineage is incompatible. | `EnsureSchemaReady`. | Server runtime dependency assembly and readiness tests. | `pgxpool`, embedded migration source, migration remediation, config diagnostics. | `schema_readiness_test.go` and server assembly tests. | Diagnostic reason codes, migration lineage, startup sequencing. | Final owner `module.database_migrations`. | high | Ordering and exact diagnostic behavior are observable. |
| `internal/platform/postgres/schema_readiness_test.go` | Characterizes current, absent, behind, ahead, historical, and wrong-lineage readiness cases. | Six tests. | Direct package testing; not found in current family selectors. | Readiness logic and PostgreSQL migration fixtures. | This file is the test. | Startup diagnostics and remediation report. | Final owner `module.database_migrations`. | high | Must be accounted for before package movement. |
| `internal/platform/postgres/telemetry.go` | Wraps `DB` operations with approved OpenTelemetry spans and duration metrics. | `InstrumentDB`; returned value remains the `DB` interface. | Server runtime assembly. | Platform telemetry, OpenTelemetry APIs, `pgx`, `pgconn`. | `telemetry_test.go`; OpenTelemetry conformance evidence. | Adopted telemetry corpus and golden/registry references. | `platform.postgres`. | high | Correctly transport-neutral and PostgreSQL-specific; SQL and binding values are not emitted. |
| `internal/platform/postgres/telemetry_test.go` | Characterizes no-SDK behavior preservation and stable PostgreSQL error classes. | Two tests plus fake `DB`/row methods. | OpenTelemetry conformance evidence; not found in current platform family selectors. | Telemetry wrapper, `pgx`, `pgconn`. | This file is the test. | PostgreSQL telemetry contract. | `platform.postgres`. | high | Static evidence pointers do not substitute for explicit test execution accounting. |
| `internal/platform/postgres/testsupport/migrationtest/apply.go` | Applies embedded migrations through a positive target version in tests. | `ApplyThrough`. | Migration contract tests inside and outside the target. | `postgres.Migrate`, `database/sql`. | `apply_test.go` and source-owner migration tests. | Test-support inventory entry `platform_postgres`. | Final owner `module.database_migrations`; planned owner-local test support. | medium | Test-only helper; no production mutation path. |
| `internal/platform/postgres/testsupport/migrationtest/apply_test.go` | Guards rejection of non-positive target versions before database access. | One test. | Direct package testing; not found in current family selectors. | Migration test helper. | This file is the test. | Test-support behavior. | Final owner `module.database_migrations`. | low | Verification ownership must follow any helper relocation. |
| `internal/platform/postgres/transaction_runner.go` | Centralizes transaction begin, callback, deferred rollback, and commit. | `TransactionRunner`, `NewTransactionRunner`, `WithinTx`. | Network Flow store/module assembly and server composition. | `DB`, `pgx.Tx`, `pgx.TxOptions`. | Network Flow transaction/store tests and broader server tests. | No generated contract; transaction semantics are observable. | Keep in `platform.postgres`; redesign is separately deferred. | high | The callback exposes `pgx.Tx`; this refactor does not change it. |

## 3. Module Boundary Diagnosis

The target is a legitimate persistence-adjacent platform adapter with mixed responsibilities. It is not a frontend shell/controller, grid-vendor integration layer, HTTP/WebSocket transport, or domain module. Its production code contains no timeline, collaboration, entity/indicator, evidence, link, saved-view, view-contract, or projection-refresh orchestration. References to several of those concerns occur in migration contract tests, not runtime ownership.

**PGT-REQ-001 — Final ownership.** The final implementation MUST retain `internal/platform/postgres` as the PostgreSQL adapter for its narrow `DB` port, root/service binding resolution, handle construction, telemetry wrapper, and unchanged generic transaction lifecycle. It MUST establish the logical owner `module.database_migrations` at `internal/modules/database_migrations` for migration source handling, Goose coordination, readiness, remediation, evidence, recovery contribution, and migration test support. The physical root is an implementation binding; it does not make migration SQL or schema evolution a source-state domain model.

This planning decision resolves RB-001. It does not supersede the current Core corpus. Production movement MUST NOT begin until the owner clarification proposed in Section 7 is adopted.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Narrow database operation port | `db.go` | `platform.postgres` | keep | 154 importing Go files include module stores and platform adapters; interface exposes only pgx database operations. | Broad use increases change risk but does not itself indicate accidental ownership. |
| Filesystem-root and managed-service DSN resolution | Connection portion of `postgres.go` | `platform.postgres` | keep | Core 04 binding rules; config/server/migrate/operator assembly callers; settings tests. | Preserve path safety, service-ref normalization, and error redaction. |
| Pool and `database/sql` construction | Connection portion of `postgres.go` | `platform.postgres` | keep | Server, migrate, and operator composition roots construct handles here. | Resource ownership remains with application runtimes. |
| PostgreSQL telemetry | `telemetry.go` | `platform.postgres` | keep | Adopted OpenTelemetry NLSpec names the PostgreSQL scope and exact signals. | SQL, bind values, database names, and server attributes remain forbidden. |
| Generic transaction lifecycle | `transaction_runner.go` | `platform.postgres` | keep | Only Network Flow production assembly consumes it; callback exposes `pgx.Tx`. | Keep unchanged in this refactor; redesign remains separately deferred. |
| Migration source, Goose execution, and process-global guard | Migration portion of `postgres.go` | `module.database_migrations` | split | Embedded source caller, migrate CLI, server readiness, test utilities, and migration tests. | Preserve cancellation, BaseFS serialization, logger restoration, and status values. |
| Schema readiness and remediation | `schema_readiness.go`, `migration_remediation.go` | `module.database_migrations` | move | Server startup, migrate CLI, typed diagnostics, and lineage tables. | Exact error reasons and JSON schema are frozen. |
| Migration evidence audit | `migrationevidence/**` | `module.database_migrations`; operator retains CLI transport | move | Operator command delegates to a transport-neutral builder. | Preserve one JSON document plus newline at the command boundary. |
| Recovery catalog contribution | `recovery_state.go` | `module.database_migrations` | move | Contribution is named `module.database_migrations` and owns `schema_migration_lineage`. | Update recovery assembly and fixtures only through their owners. |
| Migration test helper | `testsupport/migrationtest/**` | `module.database_migrations` owner-local test support | move | Helper only coordinates `Migrate` through a version. | Update the authored test-support inventory with the move. |
| Source-owner migration contract tests | Seven root `*_migration_test.go` files | Entities, Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, and Saved Views | move | Active test-family rows already map six files to source owners; schema/object ownership identifies the related modules. | Entity-alias routing is anomalously attributed to `platform.postgres` and needs correction. |

**PGT-REQ-002 — Import direction.** The final dependency graph MUST satisfy every row below. An unspecified import is permitted only when it complies with the repository-wide boundary policy and does not transfer a responsibility in the table above.

| Importing scope | Imported scope | Rule | Reason |
| --- | --- | --- | --- |
| `internal/platform/postgres` | `internal/modules/database_migrations` | MUST NOT | The connection adapter cannot depend on the migration lifecycle owner or re-export it. |
| `internal/modules/database_migrations` | `internal/app/**` | MUST NOT | Process, transport, and startup composition remain application-facade responsibilities. |
| `internal/modules/database_migrations` | Source-owner modules | MUST NOT | Migration lifecycle ownership does not transfer schema meaning or product behavior. |
| `internal/modules/database_migrations` | `db/migrations` | MUST NOT | Callers inject a source; the engine cannot import the repository source and create a cycle. |
| `db/migrations` | `internal/modules/database_migrations` | MAY, only from the Go source wrapper | `db/migrations.Source` constructs the embedded `MigrationSource`; SQL files remain data and import nothing. |
| Server, migrate, operator, recovery, and test composition | PostgreSQL and database migrations | MAY | Application composition opens handles, constructs sources, and joins the two concerns. |
| Source-owner production packages | Database-migration execution/readiness APIs | MUST NOT | Product modules cannot execute or inspect deployment migration lifecycle. |
| Source-owner tests | `database_migrations/testsupport/migrationtest` | MAY | Owner tests may apply authored migrations through the owner-local test helper. |

**PGT-REQ-003 — Secret boundary.** Database migrations MUST receive already opened database handles. It MUST NOT accept, resolve, retain, log, serialize, or expose a raw DSN, credential, database-root path, secret-bearing settings/binding object, or service secret. Migration evidence MAY retain its sanitized `DatabaseBinding` projection containing only `BindingKind` and optional `ServiceRef`; neither field carries connection material. Root confinement, `postgres.dsn` handling, service-reference normalization, and connection creation MUST remain in `platform.postgres`.

## 4. Public Contract and Behavior Freeze Map

**PGT-REQ-004 — API relocation.** The completed refactor MUST implement the following API map. Unless a row explicitly changes a parameter type, only the import path may change. `platform.postgres` MUST contain no compatibility alias, re-export, forwarding wrapper, or duplicate migration implementation at completion.

| Surface | Current path | Final path | Required final interface |
| --- | --- | --- | --- |
| PostgreSQL operation port | `internal/platform/postgres` | unchanged | `DB` retains `Exec`, `Query`, `QueryRow`, and `BeginTx` with their current pgx signatures. |
| Binding and connection | `internal/platform/postgres` | unchanged | DSN constants, `Settings`, `Binding`, `ResolveSettings`, `Setup`, `OpenSQL`, and `EnvKeyForServiceRef` retain current signatures. |
| PostgreSQL instrumentation | `internal/platform/postgres` | unchanged | `InstrumentDB(DB, string) DB` and the approved signals remain unchanged. |
| Transaction lifecycle | `internal/platform/postgres` | unchanged | `TransactionRunner`, `NewTransactionRunner`, and `WithinTx` retain current signatures and behavior. |
| Migration source and status | `internal/platform/postgres` | `internal/modules/database_migrations` | `MigrationSource`, `MigrationStatus`, `NewMigrationSource`, and `NewEmbeddedMigrationSource` retain names, fields, and signatures. |
| Migration execution | `internal/platform/postgres` | `internal/modules/database_migrations` | `Migrate(context.Context, *sql.DB, MigrationSource, string, ...string) (MigrationStatus, error)`; `GooseLogFileEnv` moves with it. |
| Read-only ledger port | Not present | `internal/modules/database_migrations` | `LedgerReader` exposes exactly `Query(context.Context, string, ...any) (pgx.Rows, error)` and `QueryRow(context.Context, string, ...any) pgx.Row`. It exposes no mutation, transaction, connection, or secret method. |
| Schema readiness | `internal/platform/postgres` | `internal/modules/database_migrations` | `EnsureSchemaReady(context.Context, *pgxpool.Pool, MigrationSource) error` retains the concrete entrypoint and nil-pool behavior; its query implementation delegates to `LedgerReader`. |
| Remediation | `internal/platform/postgres` | `internal/modules/database_migrations` | `MigrationRemediationReport`, `MigrationRemediationFinding`, `MigrationRemediationError`, `Error`, and `ReportJSON` retain names and JSON fields. |
| Recovery contribution | `internal/platform/postgres` | `internal/modules/database_migrations` | `RecoveryStateContribution() recoverystate.Contribution` retains its return contract. |
| Migration evidence | `internal/platform/postgres/migrationevidence` | `internal/modules/database_migrations/migrationevidence` | Existing exported constants/types remain; `Build` replaces `postgres.DB` with `database_migrations.LedgerReader` and otherwise retains its parameters and result. |
| Migration test support | `internal/platform/postgres/testsupport/migrationtest` | `internal/modules/database_migrations/testsupport/migrationtest` | `ApplyThrough(context.Context, *sql.DB, database_migrations.MigrationSource, int) (database_migrations.MigrationStatus, error)`. |

**PGT-REQ-005 — Defaults and boundary cases.** These behaviors are frozen. A value not listed here remains governed by the current implementation and its owner; this refactor MUST NOT invent a new default.

| Input or condition | Required behavior |
| --- | --- |
| `MigrationSource.Path` omitted | Normalize to `"."` before source inspection or execution. |
| `MigrationSource.Name` omitted | `MigrationStatus.Directory` uses the normalized path; a non-empty name takes precedence for display only. |
| `MigrationSource.BaseFS` nil | Inspect and run the filesystem path. A non-nil BaseFS uses the embedded source. |
| `ExpectedLineageID` empty | Skip lineage preflight. It does not disable source inspection or Goose execution. |
| `ExpectedLineageBoundary` empty when remediation is required | Use `migration_lineage`. |
| Migration directory contains no file other than optional `.gitkeep` | Return success with `MigrationStatus.Empty=true`; do not invoke Goose. |
| `Migrate` or readiness context is nil | Return the current nil-context error. |
| Context is already canceled | Return the context error before source inspection, BaseFS mutation, guard acquisition, or Goose execution, as applicable. |
| Cancellation occurs while waiting for the embedded-source guard | Return promptly with the context error and do not mutate Goose BaseFS. |
| `EnsureSchemaReady` receives a nil `*pgxpool.Pool` | Return nil without querying. The concrete entrypoint is retained to preserve this behavior. |
| Readiness source is empty | Return nil without querying migration metadata. |
| Database has no applied metadata or is behind repository head | Return config diagnostics with reason `schema_migration_required`. |
| Database is ahead of repository head | Return config diagnostics with reason `schema_version_ahead`. |
| Applied database lineage differs from the expected lineage | Return `MigrationRemediationError` using schema `cartulary.migration_remediation_report.v1`. |
| Embedded Goose execution | Serialize BaseFS access, set the selected BaseFS only for the run, and restore it to nil on every return path. |
| Goose logger environment is empty | Use the current stderr logger. A non-empty path retains current directory creation, append, permissions, serialization, close, and logger-restoration behavior. |
| Migration evidence manifest path is empty | Return the current required-path error; `Build` does not substitute `DefaultManifestPath`. |
| Migration evidence succeeds | Preserve caller `collected_at`; trim binding strings; set `evidence_only=true` and `rewrite_authorized=false`; emit deterministic audit and finding order. |
| Operator evidence succeeds | Write exactly one `cartulary.migration_history_evidence.v1` JSON document followed by one line feed and no additional stdout. |
| `ApplyThrough` version is zero or negative | Reject before database access. |

**PGT-REQ-006 — Frozen behavior.** Migration SQL bytes, filenames, numbering, ordering, lineage identifiers, Goose ledger interpretation, accepted command behavior, cancellation checkpoints, readiness ordering, diagnostics, remediation/evidence JSON, recovery catalog semantics, CLI grammar/streams/exit mapping, authorization outcomes, and PostgreSQL telemetry MUST remain unchanged. Incidental internal decomposition MAY change. Goose Provider adoption and `TransactionRunner` redesign are excluded and require separate authorization.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `DB` method set and pgx return semantics | `platform.postgres` | `db.go` and repository-wide store imports. | Store, telemetry, transaction, and integration tests. | Preserve compile-time consumers and add no methods during relocation. | high | Internal Go API, but repository-wide blast radius is large. |
| Database binding and DSN resolution | Core 04 and `platform.postgres` | `Binding`, `Settings`, root file `postgres.dsn`, normalized managed-service environment key. | `TestPostgresRootBindingResolution` and assembly tests. | Preserve missing/unsafe/empty root and invalid service-ref cases. | high | DSNs must not appear in diagnostics or telemetry. |
| Connection construction and borrowed-resource semantics | Server/migrate/operator application facades plus `platform.postgres` | `Setup`, `OpenSQL`, server runtime ownership rules. | Runtime dependency and application tests. | Preserve created-versus-borrowed close ownership and failure cleanup. | high | A package move must not alter lifecycle order. |
| Migration source and execution | Database migration infrastructure | `MigrationSource`, `MigrationStatus`, `Migrate`, embedded `db/migrations` source. | Bootstrap, cancellation, context, support, and service-backed tests. | Ensure every current context/cancellation/global-guard test is selected by the harness. | high | Commands, status fields, empty-source behavior, and Goose serialization are frozen. |
| Schema startup readiness | Database migration infrastructure and server facade | `EnsureSchemaReady` and server dependency ordering. | Six readiness tests and server assembly tests. | Retain absent, behind, ahead, current, historical, and wrong-lineage cases. | high | Must continue before schema-dependent runtime subsystems. |
| Migration remediation JSON | Database migration infrastructure | `cartulary.migration_remediation_report.v1` structs and `ReportJSON`. | Three remediation tests plus readiness/migrate callers. | Add a stable JSON/envelope assertion if existing caller tests do not cover all fields. | high | Field names, reason codes, hints, counts, and omission behavior are observable. |
| Operator migration-history evidence | Operator CLI transport plus database migration evidence builder | `cartulary.migration_history_evidence.v1`, operator command, manifest/source/ledger audit. | Evidence package unit test and operator unit/integration tests. | Preserve exactly one JSON document followed by one newline and no extra stdout. | high | Builder may move; CLI contract must not. |
| PostgreSQL telemetry | Adopted OpenTelemetry NLSpec and `platform.postgres` | Span `cartulary.postgres.operation`, metric `cartulary.postgres.operation.duration`, safe attributes, five error classes. | `telemetry_test.go` and OpenTelemetry conformance references. | Ensure the Go tests are explicitly selected in addition to static conformance evidence. | high | SQL, bind values, table/database/server names remain excluded. |
| Transaction lifecycle | `platform.postgres`, consumed by Network Flow | `TransactionRunner.WithinTx`. | Network Flow and server tests. | Preserve begin failure, callback failure, rollback attempt, commit failure, and success ordering. | high | Any abstraction change requires later owner authorization. |
| Recovery-state migration metadata | Database migration recovery owner | `RecoveryStateContribution`, recovery catalog fixture, schema ownership manifest. | Recovery catalog assembly/fixture tests. | Preserve contributor ID, category, table name, counts, and deterministic ordering. | medium | The recovery catalog is observable operator evidence. |
| Entity alias migration behavior | Entities owner | Migration 31 and entity-alias test file. | Three migration tests. | Preserve case-insensitive uniqueness, tombstoning, and count/sample preflight output. | high | Only test ownership/location is proposed to change. |
| Graph projection migration behavior | Graph Projection owner | Migration 32 and graph test file. | Two migration tests. | Preserve reset of unreferenced derived state and rejection of referenced state. | high | No runtime projection-refresh behavior is owned here. |
| Extension job cutover | Extensions owner | Migration 34, extension test file, retired-handler boundary input. | Two migration tests. | Preserve rejection of every retired handler before mutation. | high | Boundary-rule update is mandatory with relocation. |
| Incident bundle storage references | Incident Bundles owner | Migration 37 and test-family mapping. | Two migration tests. | Preserve fresh-schema and populated-table preflight cases. | high | No object-store payload behavior is implemented here. |
| Reference pack storage references | Reference Data owner | Migration 38 and test-family mapping. | Two migration tests. | Preserve fresh-schema and populated-table preflight cases. | high | Storage-reference semantics remain unchanged. |
| Saved-view storage hardening | Saved Views owner | Migration 52 and test-family mapping. | Two migration tests. | Preserve fresh schema and count-only preflight reporting. | high | No saved-view API or view-schema runtime belongs to this package. |
| Evidence blob uniqueness | Evidence owner | Migration 53 and test-family mapping. | One integration test. | Preserve preflight and database-level uniqueness enforcement. | high | No evidence application logic belongs to this package. |
| Generated protocol, view, UI, route, WebSocket, and grid contracts | Their respective owners; not this target | No generated roots, route handlers, WebSocket paths, frontend files, or grid imports occur in the target. | Existing owner suites outside this package. | None unless a later implementation changes a consumer-visible flow. | low | Explicit non-applicability finding; do not invent a contract. |
| Verification and harness accounting | Testing Harness owner plus source owners | `contracts/verification/owners/platform.postgres.json`, `tools/test_families/platform.postgres.json`, source-owner family files, execution topology. | Seven current platform rows and six source-owner migration rows. | Add explicit rows for uncovered readiness, remediation, context, telemetry, evidence, and helper tests before relocation. | high | Phase and family rows are execution/evidence accounting, not runtime architecture. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| `postgres.go` combines connection/binding concerns with migration source discovery and Goose orchestration. | Direct source inspection and distinct caller groups. | A change to migration mechanics can affect the broadly imported connection package. | `should_fix` | Split between `platform.postgres` and `module.database_migrations`. | Adopt the owner requirements, then perform the atomic API move in Section 7. |
| Domain migration contract tests are physically housed in the platform adapter. | Seven target files exercise entities, evidence, extensions, graph projection, incident bundles, reference data, and saved views. | Physical placement obscures source ownership and makes platform tests a catch-all. | `should_fix` | Existing source-owner modules. | Move tests without changing SQL/assertions and update authored test-family rows. |
| Verification-family selection does not explicitly cover several target tests. | Exact selectors omit context/cancellation, readiness, remediation, telemetry, migration-evidence, and migration-test-support tests. | Tests may exist without reliable Make-owned execution evidence. | `should_fix` | Testing Harness plus the relevant source owner. | Add owner rows before moving files; confirm `unmapped=0` through harness targets. |
| Harness and boundary inputs must change with test/package moves. | Test-family paths, verification owner, test-support inventory, and a retired-handler boundary exception name current paths. | Stale selectors can silently omit tests or fail boundary checks. | `must_fix` | Testing Harness and boundary-policy owners. | Treat authored mapping changes as part of the same slice; regenerate through Make only. |
| Generated outputs and immutable migration projections are downstream artifacts. | Generated-artifact policy, migration history manifest, and topology manifests. | Hand edits can create false evidence or drift. | `must_fix` | Respective contract/harness owners. | Edit only authored owner inputs and run the documented generators/drift checks. Never change SQL for a file move. |
| `TransactionRunner` exposes `pgx.Tx` to Network Flow. | `transaction_runner.go`, Network Flow store/module, and server assembly. | A generic abstraction could erase required transaction behavior or create cross-owner coupling. | `defer` | `platform.postgres` pending Network Flow/cross-owner review. | Characterize transaction order and decide in a later authorized design task. |
| Direct SQL exists in migration readiness/remediation/evidence code. | Queries inspect `goose_db_version` and `schema_migration_lineage`. | Moving behind a generic store could hide migration-specific semantics. | `intentional/no_action` | `module.database_migrations`. | Preserve read-only metadata inspection through `LedgerReader`; migration execution continues over `*sql.DB`. |
| The `DB` port is imported broadly by modules. | 154 Go importer files; representative module stores use only the narrow interface. | Broad blast radius for signature changes. | `intentional/no_action` | `platform.postgres`. | Freeze the method set during this refactor. |
| PostgreSQL telemetry depends on platform telemetry, not domain modules. | `telemetry.go` and adopted telemetry owner. | Moving it elsewhere could duplicate or weaken telemetry rules. | `intentional/no_action` | `platform.postgres`. | Keep the wrapper and exact signal contract in place. |
| No authorization decision is performed in target production code. | Direct source inspection; auth stores consume `DB` but policy checks are elsewhere. | Inventing an auth seam here would misplace policy. | `intentional/no_action` | Existing authentication/authorization owners. | Preserve current dependency direction; add no policy logic. |
| No duplicated row/view-schema logic or grid-vendor coupling was found. | No route, view-contract, frontend, or grid imports in the target. | Speculative work would expand scope without evidence. | `intentional/no_action` | Existing domain/frontend/grid owners. | Record non-applicability and do not create a frontend workstream. |
| Migration test support is isolated under a production platform path but contains no production mutation route. | `testsupport/migrationtest` exposes only `ApplyThrough`; test-support inventory registers it. | Relocation can break test imports or ownership accounting. | `should_fix` | `module.database_migrations` owner-local test support. | Move only after exact owner rows and a retained baseline exist. |

### Authority placement required before implementation

**PGT-REQ-007 — Authority admission.** The package move MUST NOT start until the following proposed owner changes are adopted. The identifiers are available at the recorded snapshot; a changed snapshot triggers re-enumeration rather than silent reuse.

| Authority | Proposed identifier and placement | Normative content | Current task effect |
| --- | --- | --- | --- |
| Core 01 §2.1A | `REQ-01-657` | `database_migrations` owns source identity/inspection, execution, global runner coordination, lineage preflight, head/lineage readiness, remediation, evidence, migration recovery metadata, and migration-only test support; it excludes connectivity, secrets, generic DB/transaction/telemetry, app transport, recovery orchestration, source-owner semantics, and authored SQL. | Tracker records the required amendment; Core 01 is not edited here. |
| Core 00 §5.1 | `REQ-00-069` | Owner-matrix row names Core 01 as primary, Core 04/source owners/Telemetry/Testing Harness as bounded secondary owners, and states that file placement or evidence routing does not transfer lifecycle ownership. | Tracker records the required amendment; Core 00 is not edited here. |
| Core 04 §9 | `AC-537` | Binary conformance proves the adapter/migration split and zero drift in SQL, identity, readiness, remediation, evidence, recovery, operator output, and telemetry privacy. | Tracker records the required criterion; Core 04 is not edited here. |
| Implementation guide and backend boundary input | No product requirement ID | Map the logical owner to `internal/modules/database_migrations` and enforce PGT-REQ-002. | Later authored implementation-support change. |

Core 02, Core 03, Core 05, `docs/domain.md`, the OpenTelemetry NLSpec, the Testing Harness NLSpec, and the Extensions NLSpec require no substantive amendment for this split. Their existing boundaries remain authoritative.

The repository pins Goose v3.27.0, whose local module source includes the Provider API. That is feasibility evidence only. This refactor MUST retain the current global BaseFS/logger guard and cancellation behavior inside `module.database_migrations`; Provider adoption MUST be tracked and validated separately.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Establish authority, snapshot, permitted write, and current tracker history. | This tracker; authority documents; Git status. | Confirm target/output state and clean baseline. | Authority and constraints recorded; no non-tracker mutation. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every target entry, symbols, callers, dependencies, tests, and contract connections. | All 26 target entries and representative callers/manifests. | File count, importer search, direct source reads. | Every target file has an inventory disposition. |
| WF-02 | Contract-owner mapping | parallel | WF-01 | WF-05 | Map connection, migration, telemetry, transaction, recovery, domain migration, and harness contracts to owners. | Owner docs, source-owner families, schema/recovery manifests. | Cross-check exact source and active owner rows. | Final owner map in PGT-REQ-001 is complete. |
| WF-03 | Characterization test gap analysis | parallel | WF-01 | WF-05, WF-07 | Identify tests not explicitly selected and specify exact pre-move rows. | Target tests, verification owners, test-family files. | Exact-symbol collision and selector searches. | Seven rows select all 18 previously unaccounted tests exactly once. |
| WF-04 | Boundary/coupling scan | parallel | WF-01 | WF-05 | Separate legitimate adapter seams from migration/test-placement coupling. | Target production files, importers, boundary and test-support inputs. | Import/path searches and boundary rule review. | Findings classified with required four-value taxonomy. |
| WF-05 | Facade or ownership redesign plan | chain | WF-02, WF-03, WF-04 | WF-06 | Record `module.database_migrations`, its physical roots, exclusions, API map, and deferred redesigns. | Sections 3 through 5; no production code. | Review against Core 01/Core 04 and adopted NLSpecs. | Planning decision is complete; authority-admission gate remains explicit. |
| WF-06 | Slice sequencing plan | chain | WF-05 | WF-07 | Apply the mandatory authority-first, evidence-before-move sequence. | Packages and callers named in Section 7. | Per-slice narrow Make targets and rollback boundaries. | Every ordered slice has a predecessor, exit criterion, and rollback. |
| WF-07 | Harness/test/accounting update plan | chain | WF-03, WF-06 | WF-08 | Register owners, add exact rows, rehome immutable row identities, and synchronize support/boundary/generated inputs. | Authored verification/family/boundary/support inputs and generated downstream outputs. | Owner explanations, slices, generation, drift, and boundary checks. | No zero/duplicate selector, stale row/path, unsupported helper, runtime alias, or hand-edited generated file. |
| WF-08 | Validation and final handoff | chain | WF-07 | None | Run narrow-to-broad validation and hand off exact evidence or failures. | Code/doc changes from later authorized slices and retained run artifacts. | Section 8 commands; `make agent-finalize` before broad end-of-run verification. | Binary criteria pass and remaining blockers are explicit. |

## 7. Proposed Refactor Slice Plan

All slices are proposals for a later authorized implementation task. Every slice MUST preserve behavior. Any semantic, SQL, diagnostic, telemetry, JSON, route, authorization, or wire-format change is `requires later authorization` and is outside this plan.

### 7.1 Owner and verification registration

**PGT-REQ-008 — Owner registration.** Before any row or source move, the authored verification registry and test-owner registry MUST register both owners exactly once.

| Owner | Verification contract | Required verifications | Test-family manifest |
| --- | --- | --- | --- |
| `module.database_migrations` | `contracts/verification/owners/module.database_migrations.json`, schema `cartulary.verification_contract.v3` | `module.database_migrations.verification.behavior_contract`: architecture/base/`go_test`; `module.database_migrations.verification.migration_mechanics`: architecture/support/`go_test` | `tools/test_families/module.database_migrations.json`, schema `cartulary.test_family_manifest.v3` |
| `app.migrate` | `contracts/verification/owners/app.migrate.json`, schema `cartulary.verification_contract.v3` | `app.migrate.verification.behavior_contract`: architecture/base/`go_test` | `tools/test_families/app.migrate.json`, schema `cartulary.test_family_manifest.v3` |

`platform.postgres.verification.behavior_contract` MUST remain active for connection and telemetry. `platform.postgres.verification.migration_mechanics` MUST remain active until its final referring row is retired.

### 7.2 Exact pre-move characterization rows

**PGT-REQ-009 — Exact test accounting.** The seven rows below MUST be created against the current package paths before source movement. Their 18 test symbols MUST resolve exactly once with no overlap. After movement, only the marked package paths may change; row IDs, test symbols, owners, collaborators, verification IDs, execution profiles, fixture capabilities, tiers, and postures MUST remain stable.

| Row ID | Owner / collaborators | Verification | Current package → final package | Exact tests | Evidence / runtime / resource / fixture / tier / posture |
| --- | --- | --- | --- | --- | --- |
| `module.database_migrations.unit.context_cancellation_and_goose_guard` | `module.database_migrations` / `app.migrate` | `migration_mechanics` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrateCanceledContextSkipsInspectionAndGoose`<br>`TestRunGooseEmbeddedCanceledContextReturnsWhileGuardHeld`<br>`TestRunGooseEmbeddedCanceledContextSkipsBaseFSAndGoose`<br>`TestRunGoosePassesContextToGoose` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.integration.long_running_migration_cancellation` | `module.database_migrations` / `app.migrate` | `migration_mechanics` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrateCancelsLongRunningMigration` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `module.database_migrations.integration.schema_readiness_matrix` | `module.database_migrations` / `app.server` | `behavior_contract` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestEnsureSchemaReadyAllowsCurrentHead`<br>`TestEnsureSchemaReadyRejectsAheadCurrentLine`<br>`TestEnsureSchemaReadyRejectsBehindCurrentLine`<br>`TestEnsureSchemaReadyRejectsEmptyDatabase`<br>`TestEnsureSchemaReadyRejectsHistoricalLineAboveHead`<br>`TestEnsureSchemaReadyReportsWrongLineage` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `module.database_migrations.integration.migration_lineage_remediation_matrix` | `module.database_migrations` / `app.migrate`, `app.server` | `behavior_contract` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrationLineagePreflightAllowsCurrentLine`<br>`TestMigrationLineagePreflightRejectsHistoricalLine`<br>`TestMigrationLineagePreflightReportsObservedWrongLineage` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `platform.postgres.unit.postgres_telemetry_behavior` | `platform.postgres` / `platform.telemetry` | `platform.postgres.verification.behavior_contract` | `./internal/platform/postgres` → unchanged | `TestPostgresErrorClass`<br>`TestTelemetryDBPreservesDBBehaviorNoSDK` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.unit.migration_evidence_source_audit` | `module.database_migrations` / `app.operator` | `migration_mechanics` | `./internal/platform/postgres/migrationevidence` → `./internal/modules/database_migrations/migrationevidence` | `TestMigrationEvidenceSourceAuditReportsManifestAndSourceFindings` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.unit.migration_test_support_validation` | `module.database_migrations` / none | `migration_mechanics` | `./internal/platform/postgres/testsupport/migrationtest` → `./internal/modules/database_migrations/testsupport/migrationtest` | `TestApplyThroughRejectsNonPositiveVersionBeforeDatabaseAccess` | unit / none / standard / none / fast / implementation |

The shorthand `behavior_contract` and `migration_mechanics` in this table refer to the verification IDs under the row owner. Every row uses runner `go`, status `active`, and family ID `<owner>.unit` or `<owner>.integration` as encoded in its row ID.

### 7.3 Immutable row-ID migration crosswalk

**PGT-REQ-010 — Owner migration.** An existing row that changes semantic owner MUST receive the exact new ID below. The old ID MUST be retired in the same authored change. Runtime aliases, dual-active selectors, recycled IDs, and old-ID compatibility readers are forbidden.

| Retired row ID | New row ID | Final owner / collaborator | Selector disposition |
| --- | --- | --- | --- |
| `platform.postgres.integration.migration_evidence_database_semantics_86249968df` | `module.database_migrations.integration.migration_evidence_database_semantics` | `module.database_migrations` / `app.operator` | Keep app/operator package and exact integration symbol; replace verification IDs with both new migration-owner verifications. |
| `platform.postgres.integration.real_postgre_sql_bootstrap_creates_the_required_b4fa366b90` | `module.database_migrations.integration.schema_bootstrap` | `module.database_migrations` / none | Move package path with the test; preserve symbol, integration profiles, fixture, tier, and both verification meanings. |
| `platform.postgres.support_integration.entity_alias_migration31_empty_upgrade_c5951bf6f3` | `module.entities.support_integration.entity_alias_migration_31` | `module.entities` / `module.database_migrations` | Move to the Entities test package; preserve three exact symbols and execution profiles; use `module.entities.verification.behavior_contract`. |
| `platform.postgres.support_unit.migrate_facade_uses_narrow_settings` | `app.migrate.unit.migrate_facade_contract` | `app.migrate` / `module.database_migrations` | Keep app/migrate package and six exact symbols; use `app.migrate.verification.behavior_contract`. |
| `platform.postgres.support_unit.migration_evidence_projection_semantics_130d030c9a` | `module.database_migrations.unit.migration_evidence_projection_semantics` | `module.database_migrations` / `app.operator` | Keep app/operator package and exact unit symbol; replace verification IDs with both new migration-owner verifications. |
| `platform.postgres.support_unit.schema_bootstrap_migration_guard_22ed444261` | `module.database_migrations.unit.schema_bootstrap_guard` | `module.database_migrations` / none | Move package path with the test; preserve symbol and profiles; use migration mechanics verification. |

The existing root-DSN row MUST remain under `platform.postgres`. Existing Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, and Saved Views migration rows MUST retain owner and row identity; only their selector package paths change when their tests move. Operator transport/output rows MUST remain under `app.operator` and add the migration owner only as collaborator where required.

This crosswalk is temporary implementation evidence. After final reconciliation, the implementer MUST retain the reconciliation run reference in the handoff and remove any executable or standalone temporary crosswalk artifact. This planning table is not a selector source or runtime alias.

### 7.4 Test-support inventory replacement

**PGT-REQ-011 — Test support.** The current inventory entry MUST be replaced atomically with the final entry below.

| Field | Current | Final |
| --- | --- | --- |
| `path` | `internal/platform/postgres/testsupport` | `internal/modules/database_migrations/testsupport` |
| `owner` | `platform_postgres` | `database_migrations` |
| `posture` | `platform_facade` | `owner_local` |
| `runtime_scan` | `included` | `excluded` |
| `support_scan` | `included` | `included` |
| `service_starting` | `false` | `false` |
| `rationale` | Current database-contract helper rationale | Migration-contract test support only; it starts no service and exposes no deployable surface. |

### 7.5 Mandatory implementation slices

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | None | Adopt proposed Core 00/Core 01 ownership and Core 04 conformance text; update the implementation guide. | Core owners and guide named in Section 5. | Starting code movement without authority creates an unsupported owner. | No product behavior change. | Owner-document review and `make lint-markdown`. | Revert the authority change before any downstream slice. | PGT-REQ-007 is adopted and internally consistent. |
| S-01 | S-00 | Register `module.database_migrations` and `app.migrate` with the exact verification contracts. | Verification/test-owner registries and two new owner manifests. | Invalid or duplicate IDs prevent harness resolution. | Registry/schema checks only. | `make json-shape-check`; owner explanation commands after generation. | Revert registrations and new manifests together. | Both owner IDs and every verification resolve exactly once. |
| S-02 | S-01 | Add the seven PGT-REQ-009 rows at current package paths. | New/current test-family manifests and authored topology inputs. | Zero or duplicate symbol resolution; wrong fixture classification. | All 18 exact symbols. | Owner task guides, owner explanations, generated catalog checks. | Revert rows and generated outputs together. | Eighteen symbols resolve exactly once with the specified profiles. |
| S-03 | S-02 | Retain a pre-move baseline for migration, PostgreSQL, migrate, operator, and server owners. | No source mutation; retained run artifacts. | A missing baseline makes behavior drift unattributable. | All new rows plus affected existing rows. | Owner unit/service-backed slices from Section 8. | No code rollback; discard incomplete evidence and rerun. | Successful run roots cover every affected active row and gate. |
| S-04 | S-03 | Atomically move migration execution, readiness, remediation, evidence, recovery contribution, and test support; update embedded source and app/test imports. | New database-migrations roots, PostgreSQL files, `db/migrations/source.go`, application/test composition. | API, nil/default, cancellation, JSON, startup, recovery, and secret-boundary drift. | Preserve every contract in Sections 4 and 7. | Narrow owner slices; three builds; migration drift. | Revert the atomic move; do not run migration Down or modify a database as rollback. | No migration API remains exported by PostgreSQL and PGT-REQ-001 through PGT-REQ-006 hold. |
| S-05 | S-04 | Change only current→final package paths in the seven new rows. | Database-migrations and PostgreSQL family manifests. | Immutable row identity or test inventory drift. | Same 18 symbols. | Owner explanations and slices. | Revert selector-path edits with the source move if necessary. | Row semantic digests differ only where path-derived inputs require it; identity and profiles remain stable. |
| S-06 | S-05 | Rehome the six old platform rows using PGT-REQ-010 and move seven source-owner migration tests. | Test-family manifests, module test packages, boundary exception for the extension cutover test. | Lost helpers, duplicate selectors, or owner drift. | All bootstrap, evidence, migrate-facade, entity, and other source-owner migration symbols. | Affected owner slices; boundary check; migration drift. | Revert each owner move and its row/path changes as one unit. | No row is owned from former physical placement; assertions and SQL are unchanged. |
| S-07 | S-06 | Update backend boundaries, recovery assembly/fixtures, and PGT-REQ-011 test-support inventory. | Authored boundary, recovery, and support inputs. | Stale old paths, contributor drift, or runtime test-support scanning. | Boundary, recovery catalog, and support-inventory checks. | `make backend-module-boundary-check`; `make json-shape-check`; recovery owner slice. | Revert authored inputs with the associated path move. | PGT-REQ-002, PGT-REQ-003, and PGT-REQ-011 are machine-enforced. |
| S-08 | S-07 | Generate downstream artifacts only through Make. | Generated topology/task surfaces downstream of authored inputs. | Hand-edited generated output or incomplete generation. | Generated-policy and drift checks. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`. | Revert authored inputs and their generated projections together. | Generation and protected-root drift are clean. |
| S-09 | S-08 | Retire `platform.postgres.verification.migration_mechanics` after its final reference disappears. | PostgreSQL verification contract and generated projections. | Premature retirement or orphaned verification. | All remaining PostgreSQL rows and migration-owner rows. | Owner explanations; JSON shape and generation drift. | Restore the verification if any live reference remains. | No active row references it and PostgreSQL behavior verification remains active. |
| S-10 | S-09 | Run final narrow-to-broad validation, reconcile the row crosswalk, and update handoff evidence. | Retained run roots and this tracker. | False completion from an aggregate pass that omitted an explicit owner gate. | All acceptance criteria in Section 12. | Section 8 in order; `make agent-finalize` before broad final verification. | Revert the failing implementation slice, not migrations or user data. | Every Section 12 criterion is true and retained evidence is named. |

## 8. Validation Plan

These are discovered Make-owned commands for later implementation. No product test, build, drift, conformance, or full-check command was run during this planning-only session.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| owner discovery | `make task-guide ROLE=module-author OWNER=module.database_migrations`; `make explain-test-owner OWNER=module.database_migrations` | Generated migration-owner task surface and exact row inventory. | yes | These commands are expected to reject the owner until S-01/S-02 and generation complete; pre-registration rejection is not a product failure. |
| unit | `make test-slice OWNER=module.database_migrations`; `make test-slice OWNER=platform.postgres`; `make test-slice OWNER=app.migrate`; `make test-slice OWNER=app.operator`; `make test-slice OWNER=app.server` | Migration mechanics/behavior, adapter behavior, and application facade contracts. | yes | Retain pre-move roots after S-03 and post-move roots after S-10. A broad pass does not replace an explicit required owner root. |
| integration | `make service-backed-test-slice OWNER=module.database_migrations` plus the affected source-owner service-backed slices | Readiness, remediation, cancellation, bootstrap, evidence semantics, and source-owner migration effects. | yes | Requires `postgres_migration`; all 18 new symbols and rehomed integration rows must resolve exactly once. |
| e2e/browser | `N/A: no applicable target surface discovered` | No direct route, WebSocket, frontend, selector, or grid contract. | no | Use the applicable browser target only if later work changes a consumer-visible application flow. |
| build | `make build-migrate`; `make build-server`; `make build-operator` | Application composition and moved internal API imports. | no | Required after S-04 and at final validation. |
| generated drift | `make generate-drift`; `make generated-artifact-policy-check`; `make migration-drift` | Generated topology/contracts, protected roots, and authored SQL/manifest lineage. | yes | Run before movement and after generation. Migration drift proves that package movement did not alter authored SQL or manifest lineage. |
| import-boundary/static | `make backend-module-boundary-check`; `make otel-conformance`; `make json-shape-check` | Package direction, retired-handler rules, telemetry contract evidence, JSON inputs. | yes | Boundary and telemetry checks complement, but do not replace, Go tests. |
| full check | `make test-fast`, then `make agent-finalize RESULTS_DIR=<successful-run-root>`, then `make check` | Repository-wide fast and full verification plus handoff evidence maintenance. | no | Use the exact retained successful run root when available; report an unset `RESULTS_DIR` rather than inventing evidence. |
| documentation | `make lint-markdown` | Tracker and repository Markdown. | no | Required for this documentation-only change; result is recorded in the session log after execution. |

## 9. Top-Level Work Tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| PGT-001 | Establish source posture and planning-only guardrails. | WF-00 | DONE | None | Section 1 and clean baseline inspection. | Authority, scope, allowed write, and non-goals are explicit. |
| PGT-002 | Inventory all target files and representative coupling. | WF-01 | DONE | PGT-001 | Section 2; 26 target entries and 154 Go importers. | Every target file is inventoried or explicitly out of scope. |
| PGT-003 | Map observable contracts to owners and tests. | WF-02 | DONE | PGT-002 | Sections 3 and 4. | Every discovered contract has a final planning owner and test posture. |
| PGT-004 | Specify characterization/accounting closure. | WF-03 | DONE | PGT-002 | PGT-REQ-008 through PGT-REQ-010. | Seven rows cover all 18 currently unaccounted exact symbols; row migrations have exact new IDs. |
| PGT-005 | Classify boundary and coupling findings. | WF-04 | DONE | PGT-002 | Section 5. | All findings use only the required classification values. |
| PGT-006 | Decide permanent database-migrations owner/package boundary. | WF-05 | DONE | PGT-003, PGT-004, PGT-005 | PGT-REQ-001 through PGT-REQ-003; RB-001 resolution. | Logical owner, physical roots, retained surfaces, import rules, and exclusions are explicit. |
| PGT-007 | Specify exact verification ownership and row migration. | WF-07 | DONE | PGT-004, PGT-006 | PGT-REQ-008 through PGT-REQ-010; RB-002 resolution. | Owners, verification contracts, seven new rows, 18 tests, and six row migrations are decision-complete. |
| PGT-008 | Adopt Core and guide ownership prerequisites. | WF-05 | TODO | PGT-006 | S-00; IG-001. | Proposed Core and guide changes are adopted before code or harness movement. |
| PGT-009 | Register owners, add exact rows, and retain baseline. | WF-07 | TODO | PGT-007, PGT-008 | S-01 through S-03; IG-002. | Owners/rows resolve exactly once and pre-move evidence covers all affected gates. |
| PGT-010 | Move migration code and update callers/selectors. | WF-06 | TODO | PGT-009 | S-04 and S-05 evidence. | Final API/dependency map holds with frozen behavior. |
| PGT-011 | Relocate source-owner migration tests and row identities. | WF-06 | TODO | PGT-010 | S-06 evidence. | Tests execute under semantic owners with unchanged assertions and migration bytes. |
| PGT-012 | Update authored boundary, recovery, support, and generation inputs. | WF-07 | TODO | PGT-011 | S-07 through S-09 evidence. | No stale path, alias, orphan verification, or generated drift remains. |
| PGT-013 | Run narrow-to-broad implementation validation and final handoff. | WF-08 | TODO | PGT-012 | S-10 retained roots and reconciliation. | Every Section 12 implementation criterion passes. |
| PGT-014 | Frontend, route, WebSocket, view-contract, and grid refactor. | WF-04 | DEFERRED | None | Non-applicability findings in Sections 3 through 5. | Reopen only if later repository evidence reveals a direct affected surface. |
| PGT-015 | Redesign the pgx transaction abstraction. | WF-05 | DEFERRED | PGT-006 | Transaction finding and Network Flow owner review. | Later authority either retains the adapter or approves a characterized replacement. |
| PGT-016 | Adopt Goose Provider API. | WF-05 | DEFERRED | PGT-013 | Pinned-module feasibility finding only. | Separate authorization and compatibility plan cover every migration-runner behavior. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Planning snapshot complete; implementation unauthorized. | Inspected framework, Core 00/01/04, adopted telemetry/harness NLSpecs, relevant `domain.md`; touched only this tracker. | `git status --short --branch`; direct `sed` reads; `rg`/`find` inventory searches. | Authority order and single-write boundary recorded; no owner contradiction found. | RB-001, RB-002 block implementation, not tracker completion. | Obtain owner approval before any package or harness mutation. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Planning decisions complete; authority adoption remains a future gate. | Inspected `temp/analysis-notes.md`, `docs/research/nlspec-spec.md`, Core identifier spaces, relevant owner/harness inputs, and this tracker; touched only this tracker. | `sed`, `rg`, `find`, `jq`, Git status, and local pinned Goose source inspection. | RB-001 and RB-002 resolved as planning decisions; PGT-REQ-007 prevents premature implementation. | IG-001, IG-002, IG-003. | Adopt S-00 before any production, test, contract, or harness move. |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Mixed adapter/migration boundary diagnosed. | All 26 target entries; representative server/migrate/operator/recovery/Network Flow callers; backend boundary inputs; touched only this tracker. | `rg` import/symbol/caller searches; direct source reads. | Keep connection/DB/telemetry; candidate migration split; transaction redesign deferred. | RB-001. | Ratify the database-migrations owner and permanent package boundary. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Final planning boundary specified. | Re-read migration, readiness, remediation, evidence, recovery, telemetry, transaction, source wrapper, and caller seams; touched only this tracker. | Exact symbol, call-shape, nil-use, and dependency searches. | `module.database_migrations` and its roots, API map, import rules, and secret boundary are unambiguous; transaction and Provider redesign remain deferred. | IG-001. | Adopt owner text, then execute S-04 only after the evidence baseline. |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Deferred as not applicable to current target. | Target imports and repository caller search; touched only this tracker. | `rg` for route, WebSocket, frontend, view-contract, selector, and grid coupling. | No frontend shell/controller, route, WebSocket, generated view, or grid-vendor surface found. | None. | Reopen only if a later code slice changes a consumer-visible frontend flow. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Non-applicability retained. | Reused current target/caller evidence; touched only this tracker. | No new frontend command; prior targeted searches remain the evidence. | The exact final boundary introduces no frontend, route, WebSocket, selector, view-contract, or grid surface. | None. | Keep PGT-014 deferred unless later repository evidence changes applicability. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Contract freeze map complete; no generated artifact changed. | Migration source/manifest, schema ownership, recovery fixture, generated-artifact policy, telemetry contracts; touched only this tracker. | Direct reads and `rg`; `make help`; `make help-all`; `make explain-target` discovery calls. | Live lineage is 57 migrations; archived 23/40 counts are stale; Make-owned drift commands identified. | RB-001 for final migration owner. | Change authored inputs only in later slices and regenerate through Make. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Authority/API/default map specified; no owner or generated file changed. | Core 00/01/04 identifiers, migration source, generated policy, recovery/schema manifests, and local Goose v3.27.0 source; touched only this tracker. | `rg` identifier maxima and Provider API inspection; direct reads. | Proposed IDs are REQ-00-069, REQ-01-657, and AC-537 at this snapshot; Provider adoption is separate. | IG-001, IG-003. | Re-enumerate on snapshot drift; otherwise adopt S-00 and use Make generation in S-08. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Existing owner rows mapped; exact-selection gaps recorded. | All target tests; `platform.postgres` verification/family files; source-owner families; test-support inventory; touched only this tracker. | `make task-guide ROLE=module-author OWNER=platform.postgres`; `make explain-test-owner OWNER=platform.postgres`; `make explain-target TARGET=backend-unit DETAIL=rows`; `make explain-target TARGET=backend-integration DETAIL=rows`; exact-selector `rg`. | Seven platform rows and six source-owner migration rows found; several target tests lack explicit family selectors. These were discovery commands, not test execution. | RB-002. | Add explicit authored rows before moving tests. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Exact accounting design complete; repository inputs unchanged. | Test-family schema, registries, PostgreSQL/app/source-owner manifests, all 18 test definitions, migrate facade tests, and test-support inventory; touched only this tracker. | `jq` family inspection; exact-symbol occurrence/collision searches; direct test reads. | Seven schema-valid row designs cover 18 currently unselected symbols; six owner migrations have new immutable IDs and a temporary crosswalk. | IG-002, IG-003. | Implement S-01/S-02, prove exact resolution, then retain S-03 baseline. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Security-sensitive configuration and telemetry surfaces frozen; no authorization policy found in target. | `postgres.go`, settings tests, telemetry code/tests, Core 04, telemetry NLSpec; touched only this tracker. | Direct reads and targeted `rg`. | Preserve root confinement, managed-service key normalization, DSN secrecy, telemetry minimization, and current auth dependency direction. | None. | Include settings and telemetry checks in every relevant later slice. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Secret and telemetry boundaries are explicit acceptance requirements. | Binding, readiness, evidence, telemetry, and caller interfaces; touched only this tracker. | Direct source and interface-use searches. | Migration owner receives handles only; DSNs/credentials remain in PostgreSQL; no authorization surface is added. | IG-001. | Enforce PGT-REQ-002/003 in S-07 and retain telemetry/settings evidence. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Tracker complete; implementation remains unauthorized. | Touched only `docs/handoffs/postgres-module-refactor-tracker.md`. | `git diff --check`; `make lint-markdown`. | Both passed. Markdown run root: `.cartulary/test-results/20260807T221341Z-p2074291`. No product validation or implementation was performed. | RB-001, RB-002. | Verify the final diff is tracker-only, then hand off for owner review. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Revised tracker complete; implementation remains unauthorized. | Touched only `docs/handoffs/postgres-module-refactor-tracker.md`; preserved the prior staged state. | `git diff --check`; `make lint-markdown`. | Both passed; Markdown run root `.cartulary/test-results/20260807T230719Z-p2092390`. No product test, build, generator, migration, Core, contract, harness, or production change was performed. | IG-001, IG-002, IG-003. | Confirm staged/unstaged path scope, then hand off S-00. |

## 11. Open Questions and Blockers

The original design questions are resolved. Their resolution is binding on this refactor plan but does not claim that the required repository changes already exist.

| ID | Resolution | Current status |
| --- | --- | --- |
| RB-001 | `module.database_migrations`, realized at `internal/modules/database_migrations`, owns migration source/execution, readiness, remediation, migration evidence, migration recovery contribution, and migration test support. `platform.postgres` retains connectivity, binding resolution, DB operations, telemetry, and the unchanged transaction runner. Application facades retain process/transport behavior; `db/migrations` remains the sole authored SQL source. | RESOLVED |
| RB-002 | Seven authored rows in PGT-REQ-009 select the 18 previously unaccounted tests by exact Go symbol. They are registered at current paths before movement and retain identity while only specified package paths change afterward. PGT-REQ-010 gives exact new identities to the six existing rows that change semantic owner. | RESOLVED |

The following are implementation-admission gates, not unresolved design choices:

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| IG-001 | Core and guide ownership amendments are not adopted. | This tracker cannot create normative product ownership. Moving code first would invert the required authority order. | Adopt S-00 with the exact PGT-REQ-007 allocation. | BLOCKED |
| IG-002 | New owners, verification contracts, exact rows, and pre-move evidence do not exist. | Source movement without exact current-path selection can silently lose or duplicate evidence. | Complete S-01 through S-03 and retain successful owner roots. | BLOCKED |
| IG-003 | Identifier and selector decisions are snapshot-bound. | A changed repository can consume proposed IDs, add tests, or alter mappings. | If HEAD differs from `d526d0ad7ed531b296b026e91a54cd8bb39302cf`, repeat the enumerations required by Section 1 before S-00. | BLOCKED on snapshot change; otherwise no action |

No `BLOCKED: owner contradiction` entry is present because no contradiction was discovered among adopted owners. If an adopted owner rejects PGT-REQ-007 or assigns the same responsibility elsewhere, implementation MUST stop and record `BLOCKED: owner contradiction`; the tracker MUST NOT choose a side.

## 12. Binary Completion Criteria

### Tracker revision completion

- [x] Every file in `internal/platform/postgres` is inventoried or explicitly out of scope.
- [x] Every discovered public contract risk has a final planning owner and test posture.
- [x] Every proposed workflow has dependencies and an exit criterion.
- [x] Every proposed implementation slice is behavior-preserving; behavior changes require later authorization.
- [x] Public/internal interfaces, import directions, defaults, edge cases, failure outcomes, and intentional deferrals are explicit.
- [x] Seven rows account for all 18 previously unselected exact test symbols, and six owner migrations have exact new IDs.
- [x] Validation commands are discovered; commands unavailable until registration say so; browser/e2e is explicitly inapplicable from repository evidence.
- [x] No contradiction was found; the required `BLOCKED: owner contradiction` treatment is recorded for future discoveries.
- [x] Repository/framework and live/archive mismatches are recorded as planning findings.
- [x] Handoff sections contain enough current source, command, risk, and next-action detail to continue without rediscovering the target.

### Refactor definition of done

Unchecked criteria describe future implementation state. No aggregate pass may close a criterion whose named owner slice or artifact is absent.

| Acceptance ID | Requirements | Binary postcondition | Required evidence |
| --- | --- | --- | --- |
| PGT-AC-001 | PGT-REQ-001, PGT-REQ-007 | Core 01 assigns migration lifecycle ownership, Core 00 maps it, Core 04 contains the zero-drift criterion, and the guide maps the physical root. | Adopted owner-document review plus Markdown validation. |
| PGT-AC-002 | PGT-REQ-001, PGT-REQ-004 | `platform.postgres` contains only DB port, binding/settings, handle construction, telemetry, and unchanged transaction lifecycle; no migration symbol or package remains below it. | Boundary check, package inventory, PostgreSQL owner slice, and builds. |
| PGT-AC-003 | PGT-REQ-002, PGT-REQ-003 | Every allowed/forbidden import direction passes, and no migration API accepts or exposes secret-bearing configuration. | Backend boundary check, targeted source scan, server/migrate/operator builds. |
| PGT-AC-004 | PGT-REQ-004 | Every moved API exists only at its final path; `LedgerReader` has exactly two read methods; no alias, re-export, or duplicate implementation exists. | Compile/build evidence, exported-symbol scan, and boundary check. |
| PGT-AC-005 | PGT-REQ-005, PGT-REQ-006 | Source defaults, nil/cancellation behavior, guard/logger restoration, readiness outcomes, remediation/evidence serialization, operator framing, recovery identity, and telemetry signals match the pre-move baseline. | Migration, PostgreSQL, app migrate/operator/server, recovery, and telemetry owner evidence. |
| PGT-AC-006 | PGT-REQ-006 | All 57 migration SQL filenames, numbers, order, bytes, manifest hashes, and lineage identifiers are unchanged. | `make migration-drift` and Git diff inspection. |
| PGT-AC-007 | PGT-REQ-008, PGT-REQ-009 | Both owners and their verification contracts resolve exactly once; all 18 new symbols resolve exactly once before and after movement with unchanged row identity/profiles. | Owner explanations, pre/post owner slices, generated catalog reconciliation. |
| PGT-AC-008 | PGT-REQ-010 | Every old owner row is retired, every mapped new row is active, existing semantic-owner rows retain identity, and no runtime alias or dual selector exists. | Final owner manifests, temporary crosswalk reconciliation report, owner explanations. |
| PGT-AC-009 | PGT-REQ-011 | Test-support inventory contains only the final root/owner/posture/scan values and no stale platform entry. | JSON shape, support-inventory check, and exact path scan. |
| PGT-AC-010 | PGT-REQ-002, PGT-REQ-006 | Recovery assembly/catalog, extension retired-handler boundary, source-owner migration tests, and application facades retain their required semantics at final paths. | Recovery and affected source-owner slices, boundary check, three builds. |
| PGT-AC-011 | PGT-REQ-008 through PGT-REQ-011 | Generated outputs derive from authored inputs, protected roots contain no manual edit, and generation/json/boundary drift is clean. | `make generate-drift`, generated-artifact policy, JSON shape, and boundary checks. |
| PGT-AC-012 | All | Every required narrow gate and final broad gate succeeds at the same final source snapshot, and the handoff names retained run roots and any skipped non-applicable gate. | Section 8 commands, `make agent-finalize`, final tracker handoff, and clean scoped diff. |

The refactor is complete only when PGT-AC-001 through PGT-AC-012 all pass. Goose Provider adoption, transaction redesign, and frontend/browser work are not part of this definition of done.

### Current handoff summary

- **Tracker:** revised in place at `docs/handoffs/postgres-module-refactor-tracker.md`; prior handoff history is preserved.
- **Files inspected:** all 26 target entries; the named authority documents; `temp/analysis-notes.md`; `docs/research/nlspec-spec.md`; representative callers; Core identifier spaces; migration/recovery/schema/generated/boundary/support inputs; verification registries/contracts/families; exact test definitions; and pinned Goose v3.27.0 source.
- **Commands run:** prior discovery commands plus read-only `jq`, exact symbol/ID/collision searches, local module-source inspection, `git diff --check`, and `make lint-markdown` (passed; run root `.cartulary/test-results/20260807T230719Z-p2092390`).
- **Remaining implementation gates:** IG-001 authority adoption and IG-002 owner/row registration plus retained baseline; IG-003 applies only if the source snapshot changes.
- **Implementation status:** no production refactor, test edit, contract edit, generated-file edit, migration change, package change, or harness change was performed.
