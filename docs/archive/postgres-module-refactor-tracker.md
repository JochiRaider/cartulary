# PostgreSQL Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

- **Target path:** `internal/platform/postgres`
- **Target label:** `postgres`, derived from the final path segment and already valid lowercase kebab case.
- **Output path:** `docs/handoffs/postgres-module-refactor-tracker.md`
- **Repository snapshot:** branch `main`, commit `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`, revalidated on 2026-08-08. N-00 remains historical evidence at `3be8ddda38deb7b28734e78e487dc229be300585`.
- **Status:** iteration one, S-00 through S-12, and iteration-two N-00 through N-05 are complete. N-06 readiness and remediation hardening is the next checkpoint; N-07 through N-09 remain ordered future work.
- **Allowed change in the N-01 checkpoint:** this tracker, Core 01, Core 04, and the adopted Testing Harness NLSpec only. Product code, tests, verification inputs, generated artifacts, SQL, guides, `AGENTS.md`, and `docs/domain.md` remain unchanged until their owning implementation checkpoint.
- **Non-goals:** no authored SQL or migration-history rewrite, data migration, frontend/browser change, PostgreSQL telemetry redesign, source-owner schema-semantic change, or compatibility forwarding package.
- **Implementation authorization:** the user authorized S-00 through S-12 on 2026-08-08 and authorized N-00 as a tracker-only planning slice on 2026-08-08. N-01 through N-09 define later implementation work. Each slice is a separate workstream, and this tracker MUST be updated after completing one slice and before beginning the next.

Normative terms in this tracker use the meanings below:

| Term | Meaning in this tracker |
| --- | --- |
| MUST / MUST NOT | A binary condition for the later refactor. An implementation that violates it is incomplete. |
| SHOULD / SHOULD NOT | The required default. A deviation requires an owner-approved amendment to this tracker before implementation. |
| MAY | Intentional implementation freedom that does not alter an observable contract. |
| Current | A fact inspected at commit `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`. |
| Final | The required post-refactor state after all implementation gates and acceptance criteria pass. |

Sections 2 through 12 preserve the completed first iteration as historical
evidence. Sections 13 through 20 are the controlling plan for iteration two and
supersede earlier future-state statements where the two iterations differ.

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

Core 02, Core 03, and Core 05 do not directly govern the package surfaces discovered: the target contains no domain model implementation, interaction protocol, or claim-bearing benchmark publication. This is an applicability finding, not a reduction of their authority elsewhere.

Repository files inspected include every entry under `internal/platform/postgres`, the representative callers in `internal/app/server`, `internal/app/migrate`, `internal/app/operator`, `internal/app/recoveryassembly`, and `internal/modules/networkflow`; `db/migrations/source.go`; the live migration and schema-ownership manifests; recovery catalog fixtures; generated-artifact policy; backend-boundary inputs; test-support inventory; and the `platform.postgres` and `platform.jobs` verification owner and family files. Searches found 158 Go files importing the target package; the inventory therefore names representative inbound callers rather than repeating every store and test utility.

The live repository contains 60 numbered SQL migrations. Archived migration trackers describing 23, 40, or 57 migrations are stale historical evidence and are not used as current inventory. Where framework or archived state differs from the repository, this tracker follows the live repository and records the mismatch. If the source snapshot changes before implementation, the implementer MUST rediscover Core identifier availability, owner-ID availability, row-ID collisions, exact test symbols, selector coverage, migration count, and affected caller paths before changing any authored input.

### 1.1 Completed iteration-one implementation decision record

This subsection supersedes earlier planning-only interface and slice statements
where they conflict. At the revalidated snapshot the target has 29 files, 158
Go importers, 40 top-level tests, 18 tests without exact family selection, and
13 active rows selecting `internal/platform/postgres`. Migrations 58 through 60
add three Platform Jobs migration-contract files; the migration-59 row is
incorrectly owned by `platform.postgres`. Ten source-owner migration test files
must therefore move, and seven existing row identities must change owner.

The final migration API is typed: `Apply`, `ApplyThrough`, and
`RollbackThrough`. `MigrationStatus` contains only `SourceName` and `Empty`.
The generic `Migrate(command, args...)` API, command-bearing status, pgtest
template fields, process-global Goose BaseFS/logger state, compatibility aliases,
and the migration testsupport forwarding package are removed. A per-invocation
Goose Provider uses a source-local filesystem, a per-call logger, and the
disabled global Go-migration registry. `Apply` is the sole production execution
operation; targeted operations are restricted to tests and repository harnesses.

The implementation sequence is S-00 authority, S-01 owner registration, S-02
characterization, S-03 retained baseline, S-04 behavior-preserving package move,
S-05 selector relocation, S-06 source-owner test relocation, S-07 boundary and
support policy, S-08 generation, S-09 obsolete verification retirement, S-10
Provider/typed-API modernization, S-11 final generation/reconciliation, and S-12
validation/handoff. No later slice may start before the prior slice checkpoint
is recorded here.

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
| Narrow database operation port | `db.go` | `platform.postgres` | keep | 158 importing Go files include module stores and platform adapters; interface exposes only pgx database operations. | Broad use increases change risk but does not itself indicate accidental ownership. |
| Filesystem-root and managed-service DSN resolution | Connection portion of `postgres.go` | `platform.postgres` | keep | Core 04 binding rules; config/server/migrate/operator assembly callers; settings tests. | Preserve path safety, service-ref normalization, and error redaction. |
| Pool and `database/sql` construction | Connection portion of `postgres.go` | `platform.postgres` | keep | Server, migrate, and operator composition roots construct handles here. | Resource ownership remains with application runtimes. |
| PostgreSQL telemetry | `telemetry.go` | `platform.postgres` | keep | Adopted OpenTelemetry NLSpec names the PostgreSQL scope and exact signals. | SQL, bind values, database names, and server attributes remain forbidden. |
| Generic transaction lifecycle | `transaction_runner.go` | `platform.postgres` | keep | Only Network Flow production assembly consumes it; callback exposes `pgx.Tx`. | Keep unchanged in this refactor; redesign remains separately deferred. |
| Migration source and Goose execution | Migration portion of `postgres.go` | `module.database_migrations` | split and modernize | Embedded source caller, migrate CLI, server readiness, test utilities, and migration tests. | Preserve observable behavior while replacing process-global Goose state and generic command grammar with typed per-invocation operations. |
| Schema readiness and remediation | `schema_readiness.go`, `migration_remediation.go` | `module.database_migrations` | move | Server startup, migrate CLI, typed diagnostics, and lineage tables. | Exact error reasons and JSON schema are frozen. |
| Migration evidence audit | `migrationevidence/**` | `module.database_migrations`; operator retains CLI transport | move | Operator command delegates to a transport-neutral builder. | Preserve one JSON document plus newline at the command boundary. |
| Recovery catalog contribution | `recovery_state.go` | `module.database_migrations` | move | Contribution is named `module.database_migrations` and owns `schema_migration_lineage`. | Update recovery assembly and fixtures only through their owners. |
| Migration test helper | `testsupport/migrationtest/**` | temporary `module.database_migrations` owner-local support, then remove | move then retire | Helper only translates a version into generic `Migrate` grammar. | Typed root operations make the wrapper redundant; remove its support-inventory entry without an alias. |
| Source-owner migration contract tests | Ten root `*_migration_test.go` files | Entities, Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, Saved Views, and Platform Jobs | move | Active test-family rows map most files to source owners; schema/object ownership identifies the related modules. | Entity-alias and Jobs execution routing are anomalously attributed to `platform.postgres` and need correction. |

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
| Source-owner tests and repository test harnesses | Typed `database_migrations` targeted operations | MAY | Tests may apply or roll back through explicit versions without gaining production command grammar. |

**PGT-REQ-003 — Secret boundary.** Database migrations MUST receive already opened database handles. It MUST NOT accept, resolve, retain, log, serialize, or expose a raw DSN, credential, database-root path, secret-bearing settings/binding object, or service secret. Migration evidence MAY retain its sanitized `DatabaseBinding` projection containing only `BindingKind` and optional `ServiceRef`; neither field carries connection material. Root confinement, `postgres.dsn` handling, service-reference normalization, and connection creation MUST remain in `platform.postgres`.

## 4. Public Contract and Behavior Freeze Map

**PGT-REQ-004 — API relocation.** The completed refactor MUST implement the following API map. Unless a row explicitly changes a parameter type, only the import path may change. `platform.postgres` MUST contain no compatibility alias, re-export, forwarding wrapper, or duplicate migration implementation at completion.

| Surface | Current path | Final path | Required final interface |
| --- | --- | --- | --- |
| PostgreSQL operation port | `internal/platform/postgres` | unchanged | `DB` retains `Exec`, `Query`, `QueryRow`, and `BeginTx` with their current pgx signatures. |
| Binding and connection | `internal/platform/postgres` | unchanged | DSN constants, `Settings`, `Binding`, `ResolveSettings`, `Setup`, `OpenSQL`, and `EnvKeyForServiceRef` retain current signatures. |
| PostgreSQL instrumentation | `internal/platform/postgres` | unchanged | `InstrumentDB(DB, string) DB` and the approved signals remain unchanged. |
| Transaction lifecycle | `internal/platform/postgres` | unchanged | `TransactionRunner`, `NewTransactionRunner`, and `WithinTx` retain current signatures and behavior. |
| Migration source and status | `internal/platform/postgres` | `internal/modules/database_migrations` | `MigrationSource` and its constructors retain their current contracts. `MigrationStatus` contains only `SourceName string` and `Empty bool`; command and pgtest template fields are removed. |
| Migration execution | `internal/platform/postgres` | `internal/modules/database_migrations` | `Apply(context.Context, *sql.DB, MigrationSource)`, `ApplyThrough(context.Context, *sql.DB, MigrationSource, int64)`, and `RollbackThrough(context.Context, *sql.DB, MigrationSource, int64)` return `(MigrationStatus, error)`; `GooseLogFileEnv` moves with them. No generic command API remains. |
| Read-only ledger port | Not present | `internal/modules/database_migrations` | `LedgerReader` exposes exactly `Query(context.Context, string, ...any) (pgx.Rows, error)` and `QueryRow(context.Context, string, ...any) pgx.Row`. It exposes no mutation, transaction, connection, or secret method. |
| Schema readiness | `internal/platform/postgres` | `internal/modules/database_migrations` | `EnsureSchemaReady(context.Context, *pgxpool.Pool, MigrationSource) error` retains the concrete entrypoint and nil-pool behavior; its query implementation delegates to `LedgerReader`. |
| Remediation | `internal/platform/postgres` | `internal/modules/database_migrations` | `MigrationRemediationReport`, `MigrationRemediationFinding`, `MigrationRemediationError`, `Error`, and `ReportJSON` retain names and JSON fields. |
| Recovery contribution | `internal/platform/postgres` | `internal/modules/database_migrations` | `RecoveryStateContribution() recoverystate.Contribution` retains its return contract. |
| Migration evidence | `internal/platform/postgres/migrationevidence` | `internal/modules/database_migrations/migrationevidence` | Existing exported constants/types remain; `Build` replaces `postgres.DB` with `database_migrations.LedgerReader` and otherwise retains its parameters and result. |
| Test preparation status | `MigrationStatus` template fields | `internal/testutil/pgtest` | `PreparationStatus` owns template-clone and prepared-database metadata independently of migration execution status. |

**PGT-REQ-005 — Defaults and boundary cases.** These behaviors are frozen. A value not listed here remains governed by the current implementation and its owner; this refactor MUST NOT invent a new default.

| Input or condition | Required behavior |
| --- | --- |
| `MigrationSource.Path` omitted | Normalize to `"."` before source inspection or execution. |
| `MigrationSource.Name` omitted | `MigrationStatus.SourceName` uses the normalized path; a non-empty name takes precedence for display only. |
| `MigrationSource.BaseFS` nil | Inspect and run the filesystem path. A non-nil BaseFS uses the embedded source. |
| `ExpectedLineageID` empty | Skip lineage preflight. It does not disable source inspection or Goose execution. |
| `ExpectedLineageBoundary` empty when remediation is required | Use `migration_lineage`. |
| Migration directory contains no file other than optional `.gitkeep` | Return success with `MigrationStatus.Empty=true`; do not invoke Goose. |
| A migration operation or readiness context is nil | Return the current nil-context error. |
| Context is already canceled | Return the context error before source inspection, BaseFS mutation, guard acquisition, or Goose execution, as applicable. |
| Cancellation occurs while waiting for the embedded-source guard | Return promptly with the context error and do not mutate Goose BaseFS. |
| `EnsureSchemaReady` receives a nil `*pgxpool.Pool` | Return nil without querying. The concrete entrypoint is retained to preserve this behavior. |
| Readiness source is empty | Return nil without querying migration metadata. |
| Database has no applied metadata or is behind repository head | Return config diagnostics with reason `schema_migration_required`. |
| Database is ahead of repository head | Return config diagnostics with reason `schema_version_ahead`. |
| Applied database lineage differs from the expected lineage | Return `MigrationRemediationError` using schema `cartulary.migration_remediation_report.v1`. |
| Embedded Goose execution | Construct an isolated Provider for the invocation using the selected filesystem and `WithDisableGlobalRegistry(true)`; no process-global BaseFS or registry mutation is permitted. |
| Goose logger environment is empty | Use the current stderr logger. A non-empty path retains current directory creation, append, permissions, and closure behavior through a per-invocation logger. |
| Migration evidence manifest path is empty | Return the current required-path error; `Build` does not substitute `DefaultManifestPath`. |
| Migration evidence succeeds | Preserve caller `collected_at`; trim binding strings; set `evidence_only=true` and `rewrite_authorized=false`; emit deterministic audit and finding order. |
| Operator evidence succeeds | Write exactly one `cartulary.migration_history_evidence.v1` JSON document followed by one line feed and no additional stdout. |
| `ApplyThrough` version is zero or negative | Reject before database access. |
| `RollbackThrough` version is negative | Reject before database access; version zero is valid. |
| Targeted operation fails after execution begins | Report the explicit target as remediation `to_version`; vendor error wording remains non-contractual. |

**PGT-REQ-006 — Frozen behavior.** All 60 migration SQL bytes, filenames, numbering, ordering, lineage identifiers, Goose ledger interpretation, cancellation checkpoints, readiness ordering, diagnostics, remediation/evidence JSON, recovery catalog semantics, deployable `migrate up` CLI grammar/streams/exit mapping, authorization outcomes, and PostgreSQL telemetry MUST remain unchanged. The internal generic command API and process-global Goose coordination are intentionally removed in favor of the typed Provider design. `TransactionRunner` redesign remains excluded.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `DB` method set and pgx return semantics | `platform.postgres` | `db.go` and repository-wide store imports. | Store, telemetry, transaction, and integration tests. | Preserve compile-time consumers and add no methods during relocation. | high | Internal Go API, but repository-wide blast radius is large. |
| Database binding and DSN resolution | Core 04 and `platform.postgres` | `Binding`, `Settings`, root file `postgres.dsn`, normalized managed-service environment key. | `TestPostgresRootBindingResolution` and assembly tests. | Preserve missing/unsafe/empty root and invalid service-ref cases. | high | DSNs must not appear in diagnostics or telemetry. |
| Connection construction and borrowed-resource semantics | Server/migrate/operator application facades plus `platform.postgres` | `Setup`, `OpenSQL`, server runtime ownership rules. | Runtime dependency and application tests. | Preserve created-versus-borrowed close ownership and failure cleanup. | high | A package move must not alter lifecycle order. |
| Migration source and execution | Database migration infrastructure | `MigrationSource`, typed operations, embedded `db/migrations` source. | Bootstrap, cancellation, context, support, and service-backed tests. | Select every current context/cancellation test and add Provider/source/logger isolation plus targeted validation. | high | Source defaults, empty behavior, explicit targets, borrowed handles, and isolation are frozen. |
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
| The `DB` port is imported broadly by modules. | 158 Go importer files; representative module stores use only the narrow interface. | Broad blast radius for signature changes. | `intentional/no_action` | `platform.postgres`. | Freeze the method set during this refactor. |
| PostgreSQL telemetry depends on platform telemetry, not domain modules. | `telemetry.go` and adopted telemetry owner. | Moving it elsewhere could duplicate or weaken telemetry rules. | `intentional/no_action` | `platform.postgres`. | Keep the wrapper and exact signal contract in place. |
| No authorization decision is performed in target production code. | Direct source inspection; auth stores consume `DB` but policy checks are elsewhere. | Inventing an auth seam here would misplace policy. | `intentional/no_action` | Existing authentication/authorization owners. | Preserve current dependency direction; add no policy logic. |
| No duplicated row/view-schema logic or grid-vendor coupling was found. | No route, view-contract, frontend, or grid imports in the target. | Speculative work would expand scope without evidence. | `intentional/no_action` | Existing domain/frontend/grid owners. | Record non-applicability and do not create a frontend workstream. |
| Migration test support is isolated under a production platform path but duplicates a typed operation. | `testsupport/migrationtest` translates `ApplyThrough` into generic command grammar; test-support inventory registers it. | Retaining the wrapper adds an unnecessary compatibility surface. | `should_fix` | `module.database_migrations` during the move, then remove after typed operations land. | Move only after exact owner rows and a retained baseline; delete it atomically with caller conversion. |

### Authority placement required before implementation

**PGT-REQ-007 — Authority admission.** The package move MUST NOT start until the following proposed owner changes are adopted. The identifiers are available at the recorded snapshot; a changed snapshot triggers re-enumeration rather than silent reuse.

| Authority | Proposed identifier and placement | Normative content | Current task effect |
| --- | --- | --- | --- |
| Core 01 §2.1A | `REQ-01-657` | `database_migrations` owns source identity/inspection, typed execution, per-invocation runner coordination, lineage preflight, head/lineage readiness, remediation, evidence, and migration recovery metadata; it excludes connectivity, secrets, generic DB/transaction/telemetry, app transport, recovery orchestration, source-owner semantics, and authored SQL. | ADOPTED in S-00. |
| Core 00 §5.1 | `REQ-00-069` | Owner-matrix row names Core 01 as primary, Core 04/source owners/Telemetry/Testing Harness as bounded secondary owners, and states that file placement or evidence routing does not transfer lifecycle ownership. | ADOPTED in S-00. |
| Core 04 §9 | `AC-537` | Binary conformance proves the adapter/migration split, typed/provider isolation, secret boundary, and zero drift in SQL, identity, readiness, remediation, evidence, recovery, operator output, and telemetry privacy. | ADOPTED in S-00. |
| Implementation guide and repository procedure | No product requirement ID | Map the logical owner to `internal/modules/database_migrations`, preserve `db/migrations` ownership, and enforce the typed/secret boundary. | ADOPTED in S-00; machine boundary inputs remain S-07. |

Core 02, Core 03, Core 05, `docs/domain.md`, the OpenTelemetry NLSpec, the Testing Harness NLSpec, and the Extensions NLSpec require no substantive amendment for this split. Their existing boundaries remain authoritative.

The repository pins Goose v3.27.0, whose local module source includes the Provider API. S-10 MUST use that per-invocation API with the global Go migration registry disabled, a source-local filesystem, and a per-call logger. Provider instances MUST NOT be closed because their supplied `*sql.DB` handles are borrowed.

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | None | WF-01 | Establish authority, snapshot, permitted write, and current tracker history. | This tracker; authority documents; Git status. | Confirm target/output state and clean baseline. | Authority and constraints recorded; no non-tracker mutation. |
| WF-01 | Target inventory | chain | WF-00 | WF-02, WF-03, WF-04 | Inventory every target entry, symbols, callers, dependencies, tests, and contract connections. | All 29 target entries and representative callers/manifests. | File count, importer search, direct source reads. | Every target file has an inventory disposition. |
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

**PGT-REQ-009 — Exact test accounting.** The seven rows for existing tests below MUST be created against the current package paths before source movement. Their 18 test symbols, plus the six symbols in the two characterization rows that follow, MUST resolve exactly once with no overlap. After movement, only the marked package paths may change; row IDs, test symbols, owners, collaborators, verification IDs, execution profiles, fixture capabilities, tiers, and postures MUST remain stable unless S-10 explicitly retires or replaces the guard/support row.

| Row ID | Owner / collaborators | Verification | Current package → final package | Exact tests | Evidence / runtime / resource / fixture / tier / posture |
| --- | --- | --- | --- | --- | --- |
| `module.database_migrations.unit.context_cancellation_and_goose_guard` | `module.database_migrations` / `app.migrate` | `migration_mechanics` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrateCanceledContextSkipsInspectionAndGoose`<br>`TestRunGooseEmbeddedCanceledContextReturnsWhileGuardHeld`<br>`TestRunGooseEmbeddedCanceledContextSkipsBaseFSAndGoose`<br>`TestRunGoosePassesContextToGoose` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.integration.long_running_migration_cancellation` | `module.database_migrations` / `app.migrate` | `migration_mechanics` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrateCancelsLongRunningMigration` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `module.database_migrations.integration.schema_readiness_matrix` | `module.database_migrations` / `app.server` | `behavior_contract` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestEnsureSchemaReadyAllowsCurrentHead`<br>`TestEnsureSchemaReadyRejectsAheadCurrentLine`<br>`TestEnsureSchemaReadyRejectsBehindCurrentLine`<br>`TestEnsureSchemaReadyRejectsEmptyDatabase`<br>`TestEnsureSchemaReadyRejectsHistoricalLineAboveHead`<br>`TestEnsureSchemaReadyReportsWrongLineage` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `module.database_migrations.integration.migration_lineage_remediation_matrix` | `module.database_migrations` / `app.migrate`, `app.server` | `behavior_contract` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrationLineagePreflightAllowsCurrentLine`<br>`TestMigrationLineagePreflightRejectsHistoricalLine`<br>`TestMigrationLineagePreflightReportsObservedWrongLineage` | integration / default / io_heavy / postgres_migration / standard / implementation |
| `platform.postgres.unit.postgres_telemetry_behavior` | `platform.postgres` / `platform.telemetry` | `platform.postgres.verification.behavior_contract` | `./internal/platform/postgres` → unchanged | `TestPostgresErrorClass`<br>`TestTelemetryDBPreservesDBBehaviorNoSDK` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.unit.migration_evidence_source_audit` | `module.database_migrations` / `app.operator` | `migration_mechanics` | `./internal/platform/postgres/migrationevidence` → `./internal/modules/database_migrations/migrationevidence` | `TestMigrationEvidenceSourceAuditReportsManifestAndSourceFindings` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.unit.migration_test_support_validation` | `module.database_migrations` / none | `migration_mechanics` | `./internal/platform/postgres/testsupport/migrationtest` → `./internal/modules/database_migrations/testsupport/migrationtest` | `TestApplyThroughRejectsNonPositiveVersionBeforeDatabaseAccess` | unit / none / standard / none / fast / implementation |
| `module.database_migrations.unit.defaults_logger_and_remediation_contract` | `module.database_migrations` / `app.migrate`, `app.server` | `behavior_contract`, `migration_mechanics` | `./internal/platform/postgres` → `./internal/modules/database_migrations` | `TestMigrationSourceDefaultsAndEmptyBehavior`<br>`TestMigrationOperationRejectsNilContextBeforeInspection`<br>`TestGooseLoggerConfigurationLifecycle`<br>`TestSchemaReadinessHandlesNilPoolAndEmptySource`<br>`TestMigrationRemediationReportJSONContract` | unit / none / standard / none / fast / implementation |
| `platform.postgres.unit.transaction_lifecycle` | `platform.postgres` / `module.networkflow` | `platform.postgres.verification.behavior_contract` | `./internal/platform/postgres` → unchanged | `TestTransactionRunnerLifecycle` | unit / none / standard / none / fast / implementation |

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
| `platform.postgres.integration.jobs_execution_migration59` | `platform.jobs.integration.execution_migration` | `platform.jobs` / `module.database_migrations` | Move package path with the test; preserve symbol, execution profiles, and Jobs behavior verification. |

The existing root-DSN row MUST remain under `platform.postgres`. Existing Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, and Saved Views migration rows MUST retain owner and row identity; only their selector package paths change when their tests move. Operator transport/output rows MUST remain under `app.operator` and add the migration owner only as collaborator where required.

This crosswalk is temporary implementation evidence. After final reconciliation, the implementer MUST retain the reconciliation run reference in the handoff and remove any executable or standalone temporary crosswalk artifact. This planning table is not a selector source or runtime alias.

### 7.4 Test-support inventory lifecycle

**PGT-REQ-011 — Test support.** S-07 MUST replace the current inventory entry atomically with the temporary moved entry below. S-10 MUST remove that entry and the package after all callers use typed root operations. No compatibility wrapper or inventory alias may remain.

| Field | Current | Temporary after S-07 |
| --- | --- | --- |
| `path` | `internal/platform/postgres/testsupport` | `internal/modules/database_migrations/testsupport` |
| `owner` | `platform_postgres` | `database_migrations` |
| `posture` | `platform_facade` | `owner_local` |
| `runtime_scan` | `included` | `excluded` |
| `support_scan` | `included` | `included` |
| `service_starting` | `false` | `false` |
| `rationale` | Current database-contract helper rationale | Migration-contract test support only; it starts no service and exposes no deployable surface. |

The final inventory contains neither path because the final implementation has no
migration test-support package.

### 7.5 Mandatory implementation slices

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | None | Adopt proposed Core 00/Core 01 ownership and Core 04 conformance text; update the implementation guide. | Core owners and guide named in Section 5. | Starting code movement without authority creates an unsupported owner. | No product behavior change. | Owner-document review and `make lint-markdown`. | Revert the authority change before any downstream slice. | PGT-REQ-007 is adopted and internally consistent. |
| S-01 | S-00 | Register `module.database_migrations` and `app.migrate` with the exact verification contracts. | Verification registry and two owner contracts. | Invalid or duplicate IDs prevent harness resolution. | Registry/schema checks only. | `make json-shape-check`; test-owner registration follows atomically with nonempty manifests in S-02 because the v3 schema forbids empty row arrays. | Revert registrations and new contracts together. | Both verification owners and every verification ID resolve exactly once. |
| S-02 | S-01 | Register both test owners with nonempty manifests; add the current-path rows, the defaults/logger/remediation row, and the PostgreSQL transaction-lifecycle row; add the six characterization symbols. | Test-owner registry, new/current test-family manifests, PostgreSQL/migration tests, and authored topology inputs. | Zero or duplicate symbol resolution; wrong fixture classification. | All 24 exact symbols. | Owner task guides, owner explanations, focused unit/service-backed rows. | Revert tests, manifests, and test-owner registrations as one unit. | Twenty-four symbols resolve exactly once with the specified profiles. |
| S-03 | S-02 | Retain a pre-move baseline for migrations, PostgreSQL, Platform Jobs, migrate, operator, server, recovery, and affected source owners. | No source mutation; retained run artifacts. | A missing baseline makes behavior drift unattributable. | All new rows plus affected existing rows. | Owner unit/service-backed slices from Section 8. | No code rollback; discard incomplete evidence and rerun. | Successful run roots cover every affected active row and gate. |
| S-04 | S-03 | Atomically move migration execution, readiness, remediation, evidence, recovery contribution, and temporary test support; update embedded source and app/test imports while retaining the legacy runner internally. | New database-migrations roots, PostgreSQL files, `db/migrations/source.go`, application/test composition. | API, nil/default, cancellation, JSON, startup, recovery, and secret-boundary drift. | Preserve every contract in Sections 4 and 7. | Narrow owner slices; three builds; migration drift. | Revert the atomic move; do not run migration Down or modify a database as rollback. | No migration API remains exported by PostgreSQL and the moved legacy runner passes its baseline. |
| S-05 | S-04 | Change only current→final package paths in the seven new rows. | Database-migrations and PostgreSQL family manifests. | Immutable row identity or test inventory drift. | Same 18 symbols. | Owner explanations and slices. | Revert selector-path edits with the source move if necessary. | Row semantic digests differ only where path-derived inputs require it; identity and profiles remain stable. |
| S-06 | S-05 | Rehome the seven old platform rows using PGT-REQ-010 and move ten source-owner migration tests. | Test-family manifests, module/platform test packages, boundary exception for the extension cutover test. | Lost helpers, duplicate selectors, or owner drift. | All bootstrap, evidence, migrate-facade, entity, Jobs, and other source-owner migration symbols. | Affected owner slices; boundary check; migration drift. | Revert each owner move and its row/path changes as one unit. | Ten files are owner-local; no row is owned from former physical placement; assertions and SQL are unchanged. |
| S-07 | S-06 | Update backend boundaries, recovery assembly/fixtures, and PGT-REQ-011 test-support inventory. | Authored boundary, recovery, and support inputs. | Stale old paths, contributor drift, or runtime test-support scanning. | Boundary, recovery catalog, and support-inventory checks. | `make backend-module-boundary-check`; `make json-shape-check`; recovery owner slice. | Revert authored inputs with the associated path move. | PGT-REQ-002, PGT-REQ-003, and PGT-REQ-011 are machine-enforced. |
| S-08 | S-07 | Generate downstream artifacts only through Make. | Generated topology/task surfaces downstream of authored inputs. | Hand-edited generated output or incomplete generation. | Generated-policy and drift checks. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`. | Revert authored inputs and their generated projections together. | Generation and protected-root drift are clean. |
| S-09 | S-08 | Retire `platform.postgres.verification.migration_mechanics` after its final reference disappears. | PostgreSQL verification contract and generated projections. | Premature retirement or orphaned verification. | All remaining PostgreSQL rows and migration-owner rows. | Owner explanations; JSON shape and generation drift. | Restore the verification if any live reference remains. | No active row references it and PostgreSQL behavior verification remains active. |
| S-10 | S-09 | Replace global Goose state and generic commands with per-invocation Provider-backed typed operations; simplify status, give pgtest its own preparation status, remove migration test support, and add capability/isolation/validation tests. | Database-migrations root, pgtest, callers, tests, and support inventory. | Provider semantic differences, accidental borrowed-DB closure, or incorrect targeted remediation metadata. | Provider source/logger concurrency, public surface, ledger capability, borrowed handle, and target validation. | Migration-owner slices, affected callers, builds, and stale-symbol scans. | Revert only this modernization slice to the validated moved legacy runner. | Typed operations are the only migration API and no global Goose or obsolete support surface remains. |
| S-11 | S-10 | Regenerate topology/task surfaces, reconcile all row migrations, and scan for old paths, IDs, commands, aliases, and support entries. | Authored owner/family inputs and generated projections. | Hidden stale topology or manually edited generated output. | All active row selections and artifact-policy evidence. | Generation, JSON, artifact-policy, reconciliation, and exact scans. | Revert authored modernization inputs with their generated projections. | Every final row resolves exactly once and all stale-reference scans are empty. |
| S-12 | S-11 | Run final narrow-to-broad validation at one source snapshot and complete the tracker handoff. | Retained run roots and this tracker. | False completion from an aggregate pass that omitted an explicit owner gate. | All acceptance criteria in Section 12. | Section 8 in order; `make agent-finalize` before broad final verification. | Reopen the failing slice; do not waive an applicable gate. | Every Section 12 criterion is true and retained evidence is named. |

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
| PGT-002 | Inventory all target files and representative coupling. | WF-01 | DONE | PGT-001 | Section 2; 29 target entries and 158 Go importers. | Every target file is inventoried or explicitly out of scope. |
| PGT-003 | Map observable contracts to owners and tests. | WF-02 | DONE | PGT-002 | Sections 3 and 4. | Every discovered contract has a final planning owner and test posture. |
| PGT-004 | Specify characterization/accounting closure. | WF-03 | DONE | PGT-002 | PGT-REQ-008 through PGT-REQ-010. | Seven rows cover all 18 currently unaccounted exact symbols; row migrations have exact new IDs. |
| PGT-005 | Classify boundary and coupling findings. | WF-04 | DONE | PGT-002 | Section 5. | All findings use only the required classification values. |
| PGT-006 | Decide permanent database-migrations owner/package boundary. | WF-05 | DONE | PGT-003, PGT-004, PGT-005 | PGT-REQ-001 through PGT-REQ-003; RB-001 resolution. | Logical owner, physical roots, retained surfaces, import rules, and exclusions are explicit. |
| PGT-007 | Specify exact verification ownership and row migration. | WF-07 | DONE | PGT-004, PGT-006 | PGT-REQ-008 through PGT-REQ-010; RB-002 resolution. | Owners, verification contracts, nine new rows, 24 tests, and seven row migrations are decision-complete. |
| PGT-008 | Adopt Core and guide ownership prerequisites. | WF-05 | DONE | PGT-006 | S-00; `REQ-00-069`, `REQ-01-657`, `AC-537`; development guide and `AGENTS.md`. | Core and guide changes are adopted before code or harness movement. |
| PGT-009 | Register owners, add exact rows, and retain baseline. | WF-07 | DONE | PGT-007, PGT-008 | S-01 through S-03 roots in the handoff log. | Owners/rows resolve exactly once and pre-move evidence covers all affected gates. |
| PGT-010 | Move migration code and update callers/selectors. | WF-06 | DONE | PGT-009 | S-04 and S-05 paths and retained roots in the handoff log. | The adapter/migration split holds with the validated moved legacy runner; typed modernization remains S-10. |
| PGT-011 | Relocate source-owner migration tests and row identities. | WF-06 | DONE | PGT-010 | S-06 source-owner paths, final row IDs, and retained roots. | Tests execute under semantic owners with unchanged assertions and migration bytes. |
| PGT-012 | Update authored boundary, recovery, support, and generation inputs. | WF-07 | DONE | PGT-011 | S-07 through S-09 evidence. | No stale path, alias, orphan verification, or generated drift remains. |
| PGT-013 | Modernize Goose execution and remove the generic/support APIs. | WF-06 | DONE | PGT-012 | S-10 implementation and focused evidence. | Provider isolation, typed operations, least-privilege ledger access, and borrowed handles are proven. |
| PGT-017 | Run final generation, narrow-to-broad validation, and handoff. | WF-08 | DONE | PGT-013 | S-11/S-12 retained roots plus the IG-004 closure evidence recorded below. | Every scoped and repository-wide criterion passes without waiver. |
| PGT-014 | Frontend, route, WebSocket, view-contract, and grid refactor. | WF-04 | DEFERRED | None | Non-applicability findings in Sections 3 through 5. | Reopen only if later repository evidence reveals a direct affected surface. |
| PGT-015 | Redesign the pgx transaction abstraction. | WF-05 | DEFERRED | PGT-006 | Transaction finding and Network Flow owner review. | Later authority either retains the adapter or approves a characterized replacement. |
| PGT-016 | Adopt Goose Provider API. | WF-05 | DONE | PGT-012 | S-10 and the adopted API decision in Section 1.1. | Per-invocation providers replace every process-global Goose mutation. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-07 18:07 EDT | Codex planning session | Planning snapshot complete; implementation unauthorized. | Inspected framework, Core 00/01/04, adopted telemetry/harness NLSpecs, relevant `domain.md`; touched only this tracker. | `git status --short --branch`; direct `sed` reads; `rg`/`find` inventory searches. | Authority order and single-write boundary recorded; no owner contradiction found. | RB-001, RB-002 block implementation, not tracker completion. | Obtain owner approval before any package or harness mutation. |
| 2026-08-07 19:05 EDT | Codex NLSpec revision session | Planning decisions complete; authority adoption remains a future gate. | Inspected `docs/research/nlspec-spec.md`, Core identifier spaces, relevant owner/harness inputs, and this tracker; touched only this tracker. | `sed`, `rg`, `find`, `jq`, Git status, and local pinned Goose source inspection. | RB-001 and RB-002 resolved as planning decisions; PGT-REQ-007 prevents premature implementation. | IG-001, IG-002, IG-003. | Adopt S-00 before any production, test, contract, or harness move. |
| 2026-08-08 EDT | Codex implementation S-00 | Rebaselined the controlling tracker and adopted database-migrations ownership plus typed Provider direction. | Core 00/Core 01/Core 04, development guide, `AGENTS.md`, tracker; `domain.md` reviewed and intentionally unchanged. | HEAD/count/ID collision scans; `git diff --check`; `make lint-markdown`. | Passed; Markdown run root `.cartulary/test-results/20260808T180253Z-p2950778`. Current snapshot facts and adopted IDs are recorded; no owner contradiction. | IG-002 remains for S-01 through S-03. | Begin S-01 owner registration. |
| 2026-08-08 EDT | Codex implementation S-01 | Verification owners and their three v3 verification IDs are registered exactly once; generated topology records the new contract inputs. | Added `contracts/verification/owners/app.migrate.json` and `module.database_migrations.json`; updated verification registry and generated render index. | `make json-shape-check` failed twice while discovering nonempty-row/evidence and generation invariants; `make generate`; `git diff --check`; `make json-shape-check`. | Final JSON gate passed at `.cartulary/test-results/20260808T180930Z-p2958374`; generation passed at `.cartulary/test-results/20260808T180919Z-p2956032`. Failed discovery roots: `.cartulary/test-results/20260808T180706Z-p2953765`, `20260808T180830Z-p2954883`, and `20260808T180859Z-p2955523`. | The test-owner schema forbids empty manifests and active verifications require row or public-target evidence, so test-owner registration moves atomically into S-02; temporary broad public targets are removed when exact rows land. | Begin S-02 characterization and exact row registration. |
| 2026-08-08 EDT | Codex implementation S-02 | Both test owners are registered with exact nonempty manifests; 24 formerly unselected characterization symbols and the six migrate-facade symbols resolve under their final semantic owners. | Added both family manifests, two PostgreSQL characterization files, and nine characterization rows; updated the PostgreSQL manifest, test-owner registry, schema-readiness versions, evidence phase fixture, and generated render index. The migrate-facade row was rehomed here, rather than S-06, because `app.migrate` cannot be actively registered with an empty manifest. | `make format` initially rejected collaborator typo `module.network_flow`; fixed to `module.networkflow`; `make format`; `make generate`; `make json-shape-check`; owner explanation/task-guide commands; owner unit and service-backed slices. | JSON passed at `.cartulary/test-results/20260808T181405Z-p2967101`; module migration tests passed at `20260808T181510Z-p2974922`; service-backed rows passed at `20260808T181552Z-p2980244`; app migrate passed at `20260808T181527Z-p2977406`; PostgreSQL passed at `20260808T181533Z-p2977791`. Initial migration-owner run `20260808T181418Z-p2968379` exposed stale version-40 readiness assertions and a non-phase evidence fixture; both were corrected to the 60-migration baseline. | No remaining S-02 blocker; six later crosswalk rows remain for S-06. | Retain the full pre-move affected-owner baseline in S-03. |
| 2026-08-08 EDT | Codex implementation S-03 | The pre-move baseline covers migration mechanics, PostgreSQL, Jobs, migrate/operator/server composition, recovery catalogs, and every owner-local migration contract. | Evidence-only slice; no source changes beyond this tracker checkpoint. | Retained S-02 migration/PostgreSQL/migrate roots; ran `test-slice` for Jobs, operator, and server; ran exact recovery catalog rows and six semantic-owner service-backed rows. | Jobs `20260808T181731Z-p2986022`, operator `20260808T181756Z-p2989141`, server `20260808T181822Z-p3009950`, recovery catalog `20260808T182012Z-p3072042`, Evidence `20260808T182021Z-p3072514`, Extensions `20260808T182031Z-p3074154`, Graph Projection `20260808T182049Z-p3075835`, Incident Bundles `20260808T182105Z-p3077499`, Reference Data `20260808T182118Z-p3079152`, and Saved Views `20260808T182131Z-p3080807` passed. | Full recovery convenience run `20260808T181902Z-p3036062` passed 35 rows and failed only its browser restore row because the existing fixture composed an empty Jobs catalog. Browser behavior is unrelated and remains non-applicable to this backend move; exact affected recovery rows pass. | Begin the atomic behavior-preserving package split in S-04. |
| 2026-08-08 EDT | Codex implementation S-04 | Migration execution, readiness, remediation, evidence, recovery contribution, and temporary test support now exist only under `internal/modules/database_migrations`; connection/settings/DB/telemetry/transactions remain under PostgreSQL. | Moved migration production/test packages; split `postgres.go`; updated embedded source, pgtest, tools, app facades, recovery assembly, and every direct migration caller. SQL and manifests are untouched. | `make format` exposed the catalog’s required source/selector atomicity; `make build-migrate`; `make build-server`; `make build-operator`; `make migration-drift`; post-move owner slices. | Migrate build `20260808T183006Z-p3091180`, server `20260808T183048Z-p3108509`, operator `20260808T183059Z-p3119354`, and migration drift `20260808T183123Z-p3132746` passed. Initial server build `20260808T183011Z-p3093888` found and led to removal of an unused PostgreSQL import. | The graph validator rejects a transient moved-test/old-selector state, so S-05 selector relocation was applied atomically. Bootstrap/guard tests remain at the old path until their S-06 semantic rehome. | Complete and validate S-05 selector relocation. |
| 2026-08-08 EDT | Codex implementation S-05 | The seven characterization rows select the moved module/evidence/test-support packages without changing row identities or execution profiles. | Updated only `tools/test_families/module.database_migrations.json` selectors, then regenerated the render index. | `make format`; `make generate`; module unit/service-backed slices; PostgreSQL and migrate slices; exact old-export/path scans; `git diff --check`. | Format `20260808T183045Z-p3105344`, generation `20260808T183116Z-p3130476`, migration owner `20260808T183134Z-p3135659`, migration service rows `20260808T183147Z-p3138400`, PostgreSQL `20260808T183158Z-p3140114`, and migrate `20260808T183216Z-p3149276` passed. | Rooted-filesystem and support-policy inputs intentionally retain old paths until S-07. | Rehome source-owner migration tests and remaining row identities in S-06. |
| 2026-08-08 EDT | Codex implementation S-06 | Ten migration contract files are owner-local; all seven semantic row migrations are complete, including the app row moved in S-02. | Moved tests to Entities, Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, Saved Views, Platform Jobs, and the migration owner; updated nine family manifests and removed six remaining PostgreSQL-owned migration rows. | `make format` exposed ASCII row ordering twice and then passed; `make generate`; migration, Jobs, and seven source-owner slices; old-ID/path scans; `make migration-drift`. | Format `20260808T183748Z-p3153174`, generation `20260808T183755Z-p3156277`, migration owner `20260808T183809Z-p3158622`, Jobs `20260808T183827Z-p3161255`, Entities `20260808T183857Z-p3164738`, Evidence `20260808T183912Z-p3167064`, Extensions `20260808T183924Z-p3169150`, Graph Projection `20260808T183949Z-p3170747`, Incident Bundles `20260808T184001Z-p3172286`, Reference Data `20260808T184015Z-p3174261`, Saved Views `20260808T184029Z-p3176180`, and migration drift `20260808T184052Z-p3178059` passed. | Removing the last migration row made the old PostgreSQL migration verification invalid immediately; it was retired atomically here and will be reconfirmed at the S-09 gate. Old IDs remain only in this human reconciliation table. | Update boundary, rooted-filesystem, recovery, and support policy in S-07. |
| 2026-08-08 EDT | Codex implementation S-07 | Boundary, rooted-filesystem, recovery, and temporary support projections now describe the moved package topology. | Updated `tools/backend_module_boundaries.json`, `internal/platform/rootedfs/production_boundary_test.go`, extension retired-handler allowance, and `tools/test_support_inventory.json`; removed empty former PostgreSQL support/evidence directories. Recovery assembly already pointed at the final owner from S-04 and its fixture bytes were unchanged. | `make format`; `make json-shape-check`; `make backend-module-boundary-check`; rooted-filesystem exact row; two exact recovery-catalog rows; old-path/ID scans; `git diff --check`. | JSON passed at `.cartulary/test-results/20260808T184402Z-p3186788`, boundary at `20260808T184604Z-p3188322`, rooted filesystem at `20260808T184618Z-p3188930`, and recovery catalogs at `20260808T184627Z-p3189332`. Discovery failures `20260808T184246Z-p3185486`, `20260808T184333Z-p3186166`, and `20260808T184405Z-p3187240` identified ordering, an empty stale directory, and the required readiness-test config allowance; all were corrected. | The support entry is intentionally temporary until S-10 deletes the package. No stale production path or retired row identity remains outside this human tracker. | Run structural generation and protected-artifact validation in S-08. |
| 2026-08-08 EDT | Codex implementation S-08 | Authored owner and selector inputs have been regenerated through the public Make surface; only the expected topology render index changed. | Regenerated `tools/execution_topology_render_index.json`; no generated Go/TypeScript root or authored SQL changed. | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; generated-path inspection; `git diff --check`. | Generation passed at `.cartulary/test-results/20260808T184740Z-p3191052`, drift at `20260808T184752Z-p3193383`, and protected-artifact policy at `20260808T184752Z-p3193391`. | None. | Reconfirm the already-atomic retirement of obsolete PostgreSQL migration verification in S-09. |
| 2026-08-08 EDT | Codex implementation S-09 | The obsolete PostgreSQL migration-mechanics verification has no active reference; PostgreSQL retains three focused adapter/telemetry/transaction rows and the migration owner retains eleven pre-modernization rows. | Reconfirmed `contracts/verification/owners/platform.postgres.json`, both owner manifests, registries, and generated topology; no additional source change was needed because retirement occurred atomically in S-06. | Exact verification-ID scan; `make explain-test-owner` for both owners; `make json-shape-check`; `make generate-drift`; `git diff --check`. | Owner explanations resolved 3 PostgreSQL and 11 migration rows; JSON passed at `.cartulary/test-results/20260808T184838Z-p3198010` and generation drift at `20260808T184838Z-p3198002`. | None. | Replace the moved legacy runner with typed per-invocation Goose Providers in S-10. |
| 2026-08-08 EDT | Codex implementation S-10 | Typed `Apply`, `ApplyThrough`, and `RollbackThrough` operations are the only migration execution API; each operation constructs an invocation-local Goose Provider/source/logger, and readiness/evidence accept the exact two-method `LedgerReader`. | Replaced the legacy runner and status, converted every production/test caller, added pgtest `PreparationStatus` plus explicit head/through helpers, removed migration test support/inventory, added Provider isolation/target validation/public capability/borrowed-handle evidence, and updated evidence characterization. | `make format`; `make generate`; migration owner unit/service slices; typed Jobs rows; app migrate/operator/server and PostgreSQL slices; three builds; `make backend-module-boundary-check`; `make migration-drift`; `make test-fast`; exact legacy/global/secret/support scans; `git diff --check`. | Migration owner `20260808T185845Z-p3210271`, migration service `20260808T191830Z-p3550588`, Jobs `20260808T191830Z-p3550604`, operator `20260808T191830Z-p3550601`, server `20260808T191904Z-p3574623`, PostgreSQL `20260808T191904Z-p3574630`, boundary `20260808T191331Z-p3447251`, migration drift `20260808T191904Z-p3574599`, and `test-fast` `20260808T191614Z-p3500000` passed. Builds passed at `20260808T185909Z-p3214374`, `p3214345`, and `p3214367`. | Initial `make test` root `20260808T185926Z-p3238087` found four related issues: stale pgtest/recovery/server imports, an owner-self boundary false positive, expected generation drift, and an Extensions assertion tied to legacy Goose detail formatting; all were fixed and their narrow gates now pass. Its Jobs supervision and Imports cancellation failures were unrelated timing/state assertions. | Regenerate and reconcile final authored/generated identities in S-11. |
| 2026-08-08 EDT | Codex implementation S-11 | Final authored owner/selector inputs and generated topology reconcile with 12 migration rows, 3 PostgreSQL rows, and 1 migrate-app row; all retired paths, IDs, commands, guards, support entries, and aliases are absent. | Regenerated only `tools/execution_topology_render_index.json`; reconciled final owner explanations and scanned contracts, tools, and Go sources. No generated Go/TypeScript root or migration SQL changed. | `make generate`; `make generate-drift`; `make json-shape-check`; `make generated-artifact-policy-check`; three owner explanations; exact stale-reference scans; SQL/generated-path diff inspection; `git diff --check`. | Generation passed at `.cartulary/test-results/20260808T192031Z-p3604600`, drift at `20260808T192046Z-p3607009`, JSON at `20260808T192046Z-p3607025`, and protected-artifact policy at `20260808T192046Z-p3607032`; every stale scan was empty. | None. | Run the single-snapshot final validation matrix and close the handoff in S-12. |
| 2026-08-08 EDT | Codex implementation S-12 | All scoped owner, contract, security, build, generation, migration, and handoff gates pass at the final source snapshot. The tracker is complete, but repository-wide clean aggregate acceptance remains blocked by IG-004. | Final validation touched only the intentional `SA1012` contract-test suppression and this tracker. Reviewed the complete scoped diff, all 60 SQL migrations, generated paths, retired identifiers, selectors, aliases, and untracked paths. | Owner discovery; all affected owner unit/service-backed slices; `make test`; static/contract/build/security targets; `make test-fast`; `make agent-finalize`; `make check`; exact failed-row reruns; `make release-check`; final Markdown and diff hygiene. | Focused roots: migrations unit `20260808T194004Z-p4005402` and service `20260808T192324Z-p3626351`; PostgreSQL `20260808T192324Z-p3626348`; Jobs migration rows `20260808T192345Z-p3628857`; migrate `20260808T192324Z-p3626359`; operator `20260808T192345Z-p3628859`; server `20260808T192345Z-p3628840`; recovery `20260808T192301Z-p3621970`; and source-owner roots `20260808T192209Z-p3613216` through `20260808T192301Z-p3621971`. Static/build roots are `20260808T193603Z-p3822626` through `20260808T193645Z-p3857692`; `test-fast` passed 349/349 at `20260808T191614Z-p3500000`; finalization passed at `20260808T194027Z-p4008540`. `make check` root `20260808T194044Z-p4011257` passed 737/739; both failures passed exact same-snapshot reruns at `20260808T194700Z-p4175093` and `20260808T194708Z-p4175644`. `release-check` root `20260808T194722Z-p4177195` passed 908/910 and failed only the pre-existing browser restore fixture. | IG-004: the aggregate-only Platform Jobs shutdown-timing and browser mutation-spy failures both passed in isolation. The browser restore fixture deterministically fails before ready because it composes an empty Jobs catalog; the same failure is retained in the S-03 pre-move baseline and browser behavior is outside this refactor. A first browser rerun used invalid owner `harness.browser.boundary_support` and failed with a usage error before the corrected `harness.browser` rerun passed. | A repository owner may repair the unrelated browser restore fixture and aggregate timing flakes, then rerun `make check` and `make release-check`; no PostgreSQL remediation slice needs reopening. |

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
| RB-001 | `module.database_migrations`, realized at `internal/modules/database_migrations`, owns migration source/execution, readiness, remediation, migration evidence, and migration recovery contribution. `platform.postgres` retains connectivity, binding resolution, DB operations, telemetry, and the unchanged transaction runner. Application facades retain process/transport behavior; `db/migrations` remains the sole authored SQL source. Temporary migration test support is removed after typed operations land. | RESOLVED |
| RB-002 | Nine authored rows in PGT-REQ-009 select 24 exact Go symbols before movement. Existing row migrations use the seven identities in PGT-REQ-010; S-10 explicitly replaces obsolete guard/support selections with Provider and typed-validation rows. | RESOLVED |

The following are implementation-admission gates, not unresolved design choices:

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| IG-001 | Core and guide ownership amendments are adopted. | The authority-first gate is satisfied. | `REQ-00-069`, `REQ-01-657`, `AC-537`, development guide, and `AGENTS.md`. | COMPLETE |
| IG-002 | New owners, verification contracts, exact rows, and pre-move evidence exist. | The source-movement admission gate is satisfied. | S-01 through S-03 evidence in the handoff log. | COMPLETE |
| IG-003 | Identifier and selector decisions are snapshot-bound. | A changed repository can consume proposed IDs, add tests, or alter mappings. | Iteration-one rebaseline from `6ff1f1e7d00608f85846aaa242aab89472a18ce3`; iteration-two rebaseline at `3be8ddda38deb7b28734e78e487dc229be300585`. | COMPLETE |
| IG-004 | The repository-wide Jobs, browser, and restore-fixture failures formerly blocking aggregate acceptance were repaired outside the PostgreSQL remediation. | A failed aggregate command could not be reported as a clean release pass even when every scoped owner gate succeeded. | Jobs 15/15 at `20260808T205120Z-p249492`; app server 34/34 at `20260808T205120Z-p249491`; restore browser fixture 12/12 at `20260808T204500Z-p214101`; `make test-fast` 349/349 at `20260808T205237Z-p284009`; browser aggregate 62/62 at `20260808T205237Z-p283970`; final `make check` 740/740 at `20260808T210323Z-p547937`; final generation drift 4/4 at `20260808T210950Z-p735318`. | COMPLETE; no waiver and no skipped check |

No `BLOCKED: owner contradiction` entry is present because no contradiction was discovered among adopted owners. If an adopted owner rejects PGT-REQ-007 or assigns the same responsibility elsewhere, implementation MUST stop and record `BLOCKED: owner contradiction`; the tracker MUST NOT choose a side.

## 12. Binary Completion Criteria

### Tracker revision completion

- [x] Every file in `internal/platform/postgres` is inventoried or explicitly out of scope.
- [x] Every discovered public contract risk has a final planning owner and test posture.
- [x] Every proposed workflow has dependencies and an exit criterion.
- [x] Every proposed implementation slice is behavior-preserving; behavior changes require later authorization.
- [x] Public/internal interfaces, import directions, defaults, edge cases, failure outcomes, and intentional deferrals are explicit.
- [x] Nine rows account for 24 exact test symbols, and seven owner migrations have exact new IDs.
- [x] Validation commands are discovered; commands unavailable until registration say so; browser/e2e is explicitly inapplicable from repository evidence.
- [x] No contradiction was found; the required `BLOCKED: owner contradiction` treatment is recorded for future discoveries.
- [x] Repository/framework and live/archive mismatches are recorded as planning findings.
- [x] Handoff sections contain enough current source, command, risk, and next-action detail to continue without rediscovering the target.

### Refactor definition of done

These criteria describe the completed first iteration. No aggregate pass closed a criterion whose named owner slice or artifact was absent.

| Acceptance ID | Requirements | Binary postcondition | Required evidence | Final status |
| --- | --- | --- | --- | --- |
| PGT-AC-001 | PGT-REQ-001, PGT-REQ-007 | Core 01 assigns migration lifecycle ownership, Core 00 maps it, Core 04 contains the zero-drift criterion, and the guide maps the physical root. | Adopted owner-document review plus Markdown validation. | PASS |
| PGT-AC-002 | PGT-REQ-001, PGT-REQ-004 | `platform.postgres` contains only DB port, binding/settings, handle construction, telemetry, and unchanged transaction lifecycle; no migration symbol or package remains below it. | Boundary check, package inventory, PostgreSQL owner slice, and builds. | PASS |
| PGT-AC-003 | PGT-REQ-002, PGT-REQ-003 | Every allowed/forbidden import direction passes, and no migration API accepts or exposes secret-bearing configuration. | Backend boundary check, targeted source scan, server/migrate/operator builds. | PASS |
| PGT-AC-004 | PGT-REQ-004 | Every moved API exists only at its final path; `LedgerReader` has exactly two read methods; no alias, re-export, or duplicate implementation exists. | Compile/build evidence, exported-symbol scan, and boundary check. | PASS |
| PGT-AC-005 | PGT-REQ-005, PGT-REQ-006 | Source defaults, nil/cancellation behavior, Provider/logger isolation, readiness outcomes, remediation/evidence serialization, operator framing, recovery identity, and telemetry signals match the retained contract. | Migration, PostgreSQL, app migrate/operator/server, recovery, and telemetry owner evidence. | PASS |
| PGT-AC-006 | PGT-REQ-006 | All 60 migration SQL filenames, numbers, order, bytes, manifest hashes, and lineage identifiers are unchanged. | `make migration-drift` and Git diff inspection. | PASS |
| PGT-AC-007 | PGT-REQ-008, PGT-REQ-009 | Both owners and their verification contracts resolve exactly once; all 24 initial characterization symbols resolve exactly once before and after movement except for the two explicitly replaced S-10 rows. | Owner explanations, pre/post owner slices, generated catalog reconciliation. | PASS |
| PGT-AC-008 | PGT-REQ-010 | Every old owner row is retired, every mapped new row is active, existing semantic-owner rows retain identity, and no runtime alias or dual selector exists. | Final owner manifests, temporary crosswalk reconciliation report, owner explanations. | PASS |
| PGT-AC-009 | PGT-REQ-011 | The migration test-support package and its inventory entry are absent; tests use typed owner operations directly. | JSON shape, support-inventory check, and exact path scan. | PASS |
| PGT-AC-010 | PGT-REQ-002, PGT-REQ-006 | Recovery assembly/catalog, extension retired-handler boundary, source-owner migration tests, and application facades retain their required semantics at final paths. | Recovery and affected source-owner slices, boundary check, three builds. | PASS |
| PGT-AC-011 | PGT-REQ-008 through PGT-REQ-011 | Generated outputs derive from authored inputs, protected roots contain no manual edit, and generation/json/boundary drift is clean. | `make generate-drift`, generated-artifact policy, JSON shape, and boundary checks. | PASS |
| PGT-AC-012 | All | Every required narrow gate and final broad gate succeeds at the same final source snapshot, and the handoff names retained run roots and any skipped non-applicable gate. | Section 8 commands, `make agent-finalize`, final tracker handoff, clean scoped diff, and the IG-004 closure roots. | PASS |

The first refactor iteration is complete and PGT-AC-001 through PGT-AC-012 pass. The first `make check` after the IG-004 repair found two obsolete private catalog helpers through staticcheck; those dead helpers were removed and the complete check reran successfully. Goose Provider adoption is included; transaction redesign becomes part of iteration two, while frontend/browser work remains outside this tracker.

### Current handoff summary

- **Implementation status:** S-00 through S-12 and PGT-AC-001 through PGT-AC-012 are complete. IG-004 is closed without waiver. N-00 is the completed document-only admission slice for iteration two; N-01 through N-09 are not implemented by this update.
- **Iteration-one planning baseline:** branch `main` at `6ff1f1e7d00608f85846aaa242aab89472a18ce3`; that rebaseline found 29 target files, 158 importers, 60 migrations, 40 top-level target tests, 18 initially unselected tests, 13 target rows, and ten source-owner migration files.
- **Iteration-two planning baseline:** N-00 remains historical at `3be8ddda38deb7b28734e78e487dc229be300585`; N-01 revalidated the unchanged scoped backend inventory at `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`. The current inventory and final-state decisions are in Sections 13 through 20. `docs/domain.md` remains intentionally unchanged.
- **S-00 changed paths:** `AGENTS.md`; Core 00, Core 01, and Core 04; `docs/guides/cartulary-dev-guide.md`; and this tracker.
- **S-01 through S-03 changed paths:** `contracts/verification/owners/{app.migrate,module.database_migrations,platform.postgres}.json`, `contracts/verification/registry.json`, `tools/test_catalog_owner.json`, `tools/test_families/{app.migrate,module.database_migrations,platform.postgres}.json`, the initial database-migrations characterization tests, and `internal/platform/postgres/transaction_runner_test.go`.
- **S-04 and S-05 changed paths:** migration production/tests formerly below `internal/platform/postgres/**`, their final `internal/modules/database_migrations/**` paths, `db/migrations/source.go`, `internal/app/{migrate,operator,recoveryassembly,server}/**`, `internal/testutil/pgtest/**`, `tools/recoverybrowserrestore/main.go`, and `tools/testservices/main.go`.
- **S-06 changed paths:** owner-local migration tests under Entities, Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, Saved Views, and Platform Jobs; the ten retired PostgreSQL test paths; and `tools/test_families/{module.entities,module.evidence,module.extensions,module.graphprojection,module.incidentbundles,module.reference_data,module.savedviews,platform.jobs,platform.postgres}.json`.
- **S-07 through S-09 changed paths:** `tools/backend_module_boundaries.json`, `internal/platform/rootedfs/production_boundary_test.go`, `tools/test_support_inventory.json`, recovery assembly, the PostgreSQL verification contract, owner catalogs, and generated topology projections.
- **S-10 changed paths:** `internal/modules/database_migrations/**`, `internal/testutil/pgtest/**`, all application callers above, `internal/modules/{evidence,indicators}/**` migration callers, `internal/platform/{administrativeaudit,jobs}/**` migration callers, and removal of `internal/platform/postgres/testsupport/migrationtest/**`.
- **S-11 changed path:** the Make-generated `tools/execution_topology_render_index.json`; generated Go and TypeScript roots did not change.
- **S-12 changed paths:** the intentional `SA1012` suppression in `internal/modules/database_migrations/migration_contract_test.go` and this final tracker checkpoint.
- **Substantive result:** PostgreSQL now retains connectivity, binding/settings, DB, telemetry, and transactions. Database Migrations exclusively owns typed migration execution, readiness, remediation, evidence, and recovery contribution. Every operation uses an invocation-local Goose Provider, source filesystem, and logger; the global registry is disabled; borrowed database handles remain open; and `LedgerReader` exposes only `Query` and `QueryRow`.
- **Compatibility and migration posture:** no forwarding package, alias, re-export, duplicate runner, generic command API, or migration test-support package remains. The `migrate up` CLI, remediation schema, evidence framing, readiness codes, telemetry, and redaction are retained. No data migration or SQL rewrite is required.
- **Migration integrity:** all 60 authored SQL filenames and bytes are unchanged; manifest hashes and lineage identifiers pass `make migration-drift`; `git diff --name-only -- 'db/migrations/*.sql'` is empty.
- **First-iteration final evidence:** every focused owner, boundary, telemetry, JSON, harness, generation, protected-artifact, build, security, Markdown, and diff gate named in S-12 passes. The subsequent IG-004 repair passed the Jobs owner 15/15, app-server owner 34/34, restore browser fixture 12/12, concurrent focused stress rows, `make test-fast` 349/349, browser aggregate 62/62, final `make check` 740/740, and generation drift 4/4 at the roots recorded in the IG-004 row. No check was skipped or waived.
- **Remaining action:** execute N-01 only after accepting the iteration-two authority and verification admission checkpoint. No first-iteration remediation slice needs reopening.

## 13. Iteration-Two Scope and Current Inventory

Iteration two removes compatibility and test mechanics that were useful during
the first structural move but do not belong in a production migration module.
It also closes fail-open source/readiness behavior, serializes production
migration execution across processes, constrains failure disclosure, and removes
the remaining transaction responsibility from the PostgreSQL adapter. This is a
deliberate internal API contraction, not a behavior-preservation exercise.

### 13.1 Rebaseline

N-00 remains immutable historical evidence at
`3be8ddda38deb7b28734e78e487dc229be300585`. Before N-01, the scoped inventory
was revalidated at `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`.
The intervening revision changes UI/design artifacts and this tracker but does
not change any scoped PostgreSQL, Database Migrations, migration SQL,
verification-owner, or targeted-call input. The current inventory is therefore:

| Surface | Current count | Rebaseline finding |
| --- | ---: | --- |
| Tracked `internal/platform/postgres` paths | 8 | `db.go`, `postgres.go`, `postgres_settings_test.go`, `telemetry.go`, `telemetry_test.go`, `transaction_runner.go`, `transaction_runner_test.go`, and redundant `.gitkeep`. |
| Tracked `internal/modules/database_migrations` paths | 14 | Five production files and nine test files, including the `migrationevidence` subpackage. |
| Imports of `database_migrations` | 35 | Thirty-five import declarations across 34 Go files; one operator integration test imports both the root and `migrationevidence` subpackage. Includes composition roots, migration-owning modules, test support, source-owner migration tests, and tools. |
| Authored numbered SQL migrations | 60 | Immutable for this iteration. |
| Direct targeted-operation calls | 32 | 26 `ApplyThrough` calls and 6 `RollbackThrough` calls. |
| `module.database_migrations` owner rows | 12 | Five service-backed and seven unit rows. |
| `platform.postgres` owner rows | 3 | Settings, telemetry, and transaction-lifecycle coverage. |

The current legacy surfaces are `MigrationStatus`, `PreparationStatus`,
path/default-oriented migration sources, production-exported test-only targeted
operations, process-global Goose log-path configuration, and
`platform.postgres.TransactionRunner`. The adapter's managed-service DSN
fragments and test-only `CARTULARY_POSTGRES_DSN` key are also unnecessarily
exported. These surfaces are not retained merely because iteration one retained
or introduced them.

`docs/domain.md` remains unchanged. Migration runners, PostgreSQL adapters,
transaction helpers, test fixtures, verification rows, test harnesses, and package topology remain
implementation-support concepts under its Section 6 classification. Core 00's
logical ownership assignment also remains unchanged.

### 13.2 Scope, compatibility, and non-goals

- N-01 through N-09 may change internal Go APIs, owner specifications,
  verification inputs, test-support APIs, package-private implementation, and
  generated topology projections derived from authored inputs.
- No compatibility alias, forwarding package, duplicate operation, deprecated
  grace period, command-string facade, or dual selector is permitted.
- No database migration or data rewrite is needed. All 60 SQL filenames,
  versions, order, bytes, manifest hashes, and lineage identifiers MUST remain
  unchanged.
- `migrate up` remains the only deployable migration grammar. Production does
  not gain down, redo, status, version-targeting, filesystem discovery, or
  operator-selected migration-source behavior.
- PostgreSQL settings resolution, connection construction, the narrow `DB`
  interface, and telemetry are otherwise unchanged.
- Frontend, route, WebSocket, workbook, browser, and claim-publication behavior
  is out of scope unless a later slice creates a direct dependency. Such a
  discovery MUST reopen applicability here before that slice continues.

## 14. Iteration-Two Gap Remediation Matrix

| Gap | Areas and remediation | Rationale | Long-term benefit | Compatibility or migration impact | Risk if unresolved | Complete when |
| --- | --- | --- | --- | --- | --- | --- |
| N-G01: Completion posture and inventory were stale | Documentation and planning. Revalidate against current HEAD, preserve N-00 history, and amend counts, decisions, row accounting, and snapshot assumptions before implementation. | Intervening UI/design work changed the repository snapshot and browser baseline without changing scoped backend state. | Later sessions receive accurate history without reopening iteration one or relying on stale repository-wide evidence. | Tracker-only; no runtime compatibility effect. | Failures can be misattributed, selectors can be obsolete, or final evidence can claim the wrong snapshot. | N-01 records the current commit and exact inventory; unique-ID, Markdown, and diff checks pass. |
| N-G02: Production source and result types expose mutable or meaningless state | Specification, implementation, tests, application facades, and guide. Replace `MigrationSource` with opaque immutable `Source`, remove `MigrationStatus`, make construction fallible, cache one canonical embedded snapshot, and make `Apply` error-only. Reject zero, nil-filesystem, invalid or escaping roots, absent lineage, empty or malformed catalogs, unexpected entries, invalid filenames, duplicate or non-contiguous versions, malformed markers, unsupported directives, and unbalanced statement blocks before database access. | Returning an apparently valid source without an error hides invalid state inside the opaque value; public fields and lazy filesystem reads also permit mutation and time-of-check/time-of-use drift. | Deterministic catalogs, smaller caller coupling, and a safe foundation for future catalog metadata. | Internal compile-time break across callers; no data, CLI grammar, or SQL migration. | Invalid or mutable catalogs may reach the database and future callers remain coupled to source layout. | Constructor/catalog, mutation-after-construction, zero-source, before-access, public-surface, and application-build checks pass. |
| N-G03: Test-only version targeting is a production-module capability | Testing Harness specification, pgtest implementation, source-owner tests, verification, and guide. Move apply-through and rollback-through behind an opaque pgtest-issued `MigrationDatabase` capability, bind it to the canonical embedded source and disposable scratch databases, convert all 32 calls, and remove production operations in the same checkpoint. | Free functions accepting arbitrary `*sql.DB` cannot enforce the scratch-database boundary. | Production expresses only deployable apply-to-head behavior while destructive history construction is explicitly harness-owned. | Test-only API break; no production rollback support or compatibility alias. Apply targets `<=0` and rollback targets `<0` fail before source or database access; rollback zero remains allowed. | Production or ordinary tests can silently depend on arbitrary rollback/version targeting. | The exact 26/6 inventory is converted; invalid-target and construction tests pass; no retired import or symbol remains. |
| N-G04: Migration execution is not process-safe and leaks vendor/logging concerns | Core specification, implementation, security tests, migrate/server facade tests, and documentation. Use Goose `SessionLocker` integration with advisory lock `4097083626`, locked initial/provider/final classification, one-second acquisition retries for at most five minutes, one bounded detached release path for at most 30 seconds, discarded logging, safe typed failures, exact facade output, and borrowed-handle protection. | Invocation-local providers do not serialize processes, and holding an external lock connection while Goose waits for another pool connection can deadlock a constrained pool. Raw provider/log errors are unstable and potentially sensitive. | Deterministic deployment behavior, bounded cleanup, prompt cancellation, and a reviewable diagnostic surface. | Goose log files and vendor wording disappear intentionally; successful grammar and exit mapping remain. | Interleaved DDL, leaked locks, blocked deployment, duplicate output, or secret and SQL disclosure. | Repeated concurrency, cancellation, preflight/provider/panic cleanup, unlock bounds, borrowed-handle, exact-stderr, and forbidden-data tests pass. |
| N-G05: Readiness and preflight accept ambiguous database histories | Core specification, implementation, readiness/remediation tests, and operator guidance. Introduce one pure catalog/history classifier used by Apply preflight, final verification, and readiness, with structural-invalid precedence over lineage mismatch and exact pristine/behind/current/ahead outcomes. | Maximum-version and lineage-membership checks accept gaps, duplicates, unknown versions, false rows, and mixed lineages. | One integrity model supports future migrations without normalizing or advancing corrupt databases. | Previously accepted corrupt databases fail closed; no automatic repair or historical-line bridge is added. | Corrupt state may be called ready or receive additional DDL, making recovery unsafe. | Pristine, zero-only, behind, current, ahead, duplicate, gap, unknown, false, out-of-order, missing/wrong/mixed lineage, combined-failure, nil-pool, and invalid-source matrices pass. |
| N-G06: Remediation and pgtest retain open-ended or duplicate evidence structures | Implementation, tests, and documentation. Replace exported remediation DTOs and `map[string]any` facts with private typed structures, remove unused `incident_id` and nil fallbacks, remove `PreparationStatus`, and make preparation events authoritative. | Open maps weaken schema review; unused/nil compatibility behavior increases states; status duplicates retained event evidence. | Closed schemas, simpler callers, and one test-preparation evidence path. | Callers use `RemediationReporter` rather than concrete DTOs; `PrepareDatabase` returns only database/error. JSON schema and meaningful fields remain stable. | Schema drift, unreviewed fields, and conflicting template/migration evidence can accumulate. | Public-surface tests, exact remediation JSON, preparation events, and stale-symbol scans pass. |
| N-G07: PostgreSQL retains dead exports and a low-cohesion transaction facade | Implementation, tests, boundaries, and verification. Delete `.gitkeep`; privatize DSN fragments; move the test DSN key to test support; delete the transaction runner; let Network Flow own a private helper over its existing `postgres.DB`; remove redundant server wiring. | These symbols either have no production consumer or abstract behavior used by a single module. | A smaller adapter centered on configuration, connections, DB operations, and telemetry; Network Flow transaction behavior evolves with its owner. | Network Flow dependency construction and a row owner change atomically. Transaction ordering is preserved; no other DB API changes. | Dead public API invites coupling, and a shared facade obscures the only behavior owner. | Old symbols/files/options are absent, Network Flow owns the exact lifecycle test, and adapter/build/boundary checks pass. |
| N-G08: Authority and verification still describe iteration-one compatibility | Core 01, Core 04, Testing Harness NLSpec, verification contracts, generated topology, guide, `AGENTS.md`, and tracker. Amend authority in N-01, then transition each executable row and implementation guide only when its corresponding code exists. | Active rows cannot select nonexistent or deliberately failing future tests, and guides must not describe topology that has not landed. | Specifications lead implementation while executable routing remains truthful at every checkpoint. | Rows transition atomically with their executable tests; unaffected IDs remain unchanged. | Contradictory owners, orphaned tests, false verification claims, or generated topology masking stale selectors. | IDs are unique; every active selector resolves exactly once; row edits regenerate clean topology; no executable crosswalk remains. |
| N-G09: Production completion lacks second-iteration evidence | Tests, security, builds, documentation, and handoff. Run narrow-to-broad validation at one source snapshot and close every criterion with retained roots and failure classification. | A clean design is not complete without evidence for concurrency, safety, migrations, owners, and aggregate repository behavior. | Auditable release readiness and a continuation point that does not depend on session memory. | No runtime impact beyond the implemented slices. | Partial success can be mistaken for production readiness or omit an owner gate. | N-09 records every applicable result, every failure/retry, migration immutability, no unexplained diff, and no skipped applicable check. |

## 15. Final Iteration-Two Interfaces and Behavioral Rules

### 15.1 Production API

The final exported production surface under
`internal/modules/database_migrations` is:

```go
type Source struct {
	// unexported immutable catalog snapshot and lineage metadata
}

func NewSource(
	fsys fs.FS,
	root string,
	lineageID string,
	lineageBoundary string,
) (Source, error)

func Apply(context.Context, *sql.DB, Source) error

type LedgerReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func EnsureSchemaReady(context.Context, *pgxpool.Pool, Source) error

type MigrationFailure interface {
	error
	ReasonCode() string
}

type RemediationReporter interface {
	MigrationFailure
	RemediationReportJSON() string
}

func RecoveryStateContribution() recoverystate.Contribution
```

`Source` MUST expose no fields or filesystem/path accessors. `NewSource` reads
and copies its catalog into private immutable storage and validates it without
database access. It rejects a zero source, nil filesystem, absolute,
root-escaping, or otherwise invalid root, missing lineage ID or boundary, empty
catalog, unexpected entry, filename outside the five-digit lower-snake form,
duplicate or non-contiguous versions, missing or repeated Up/Down markers,
unsupported directives, and unbalanced statement blocks. There is no `"."`
default, path-only constructor, operating-system filesystem discovery, or
empty-source success. `db/migrations.Source()` returns `(Source, error)` and
caches one validated immutable snapshot of the canonical embedded catalog; it
does not panic.

The following are removed without aliases or re-exports:

- `MigrationSource`, `MigrationStatus`, `NewMigrationSource`, and
  `NewEmbeddedMigrationSource`;
- production `ApplyThrough` and `RollbackThrough`;
- `GooseLogFileEnv` and migration-module directory/file creation;
- exported remediation report/finding data-transfer types, open-ended
  `map[string]any` facts, `incident_id`, and nil-receiver fallbacks; and
- obsolete `postgres` aliases used only to name the migrated module in tests.

### 15.2 Test-only targeted mechanics

`internal/testutil/pgtest` owns one opaque targeted capability:

```go
type MigrationDatabase struct {
	// unexported harness identity and *sql.DB
}

func (db *MigrationDatabase) SQL() *sql.DB
func (db *MigrationDatabase) ApplyThrough(context.Context, int64) error
func (db *MigrationDatabase) RollbackThrough(context.Context, int64) error
```

`Harness.MigrationDatabaseT` and `MigrationDatabaseThroughT` return this
capability. It cannot be constructed around an arbitrary database outside
pgtest, always uses `db/migrations.Source()`, is limited to disposable
migration-scratch databases created by the repository test harness, and does
not accept a source. Apply targets `<= 0` and rollback targets `< 0` fail before
source or database access; rollback target zero is allowed. The capability does
not create a production rollback contract. `PreparationStatus` is deleted and
`PrepareDatabase` returns only `(*TestDatabase, error)`; existing suite
preparation events remain the authoritative template-clone and migration
evidence. The process-wide Goose environment shim in `tools/testservices` is
deleted.

### 15.3 Apply locking and resource lifecycle

Every `Apply` performs the following ordered lifecycle:

1. Validate context, borrowed database handle, and complete source without
   database access. A canceled context returns promptly.
2. Acquire a session lock for an initial classification. Use advisory lock
   `4097083626`, cancellation-aware one-second retry intervals, and a five-minute
   acquisition ceiling. If this preflight fails after acquisition, explicitly
   release through the common cleanup path before returning.
3. If no migration is required, release the initial session lock and continue
   to the final locked classification. If work is required, release it and
   construct one Goose Provider with the immutable source, PostgreSQL dialect,
   disabled global Go-migration registry, and `io.Discard` logging.
4. Supply Goose with a validating `SessionLocker`. After it acquires the
   execution lock, it reclassifies through that same `*sql.Conn` before Goose
   executes SQL. This makes the migration SQL connection and execution lock
   session-identical and avoids holding an external pool connection while
   Goose waits for another.
5. Apply to head and never call `Provider.Close`, because the caller owns the
   `*sql.DB`.
6. Perform a final locked classification before returning success.
7. Cancellation, classification failure, provider failure, panic recovery, and
   cleanup failure all use one release path detached from the canceled caller
   context and bounded to 30 seconds. Cleanup never hides an earlier primary
   failure; a sole unlock failure is the stable cleanup failure.

Tests MUST prove all acquisition/release ceilings, session identity, explicit
unlock after locked preflight failure, primary-error precedence, and continued
caller use and closure of its own `*sql.DB` after every outcome.

### 15.4 Safe failures, remediation, and readiness

`MigrationFailure` exposes a closed safe reason set covering source invalidity,
unavailable database, lock acquisition, invalid history, required or ahead
schema, migration execution, cleanup, and failed postcondition. Context
cancellation remains `errors.Is`-compatible. `Error()` text, wrapping chain,
remediation JSON, application stderr, and retained logs MUST NOT contain SQL
text, bind values, DSNs, filesystem paths, server identity, environment values,
constraints, endpoints, or upstream/Goose error text.

The schema ID `cartulary.migration_remediation_report.v1`, existing historical
lineage reason, meaningful JSON member names, member order, and output bytes
remain. Implementation uses private typed facts with no `incident_id` or
nil-receiver fallback. Remediation-capable errors satisfy
`RemediationReporter`.
The migrate facade writes exactly `RemediationReportJSON() + "\n"` once to
stderr and suppresses a second generic error line for that failure; other
failures retain the existing exit-code framing without vendor detail.

A shared pure classifier consumes snapshots acquired through separate
`database/sql` and pgx adapters. Its precedence is:

1. invalid context, source, or database capability;
2. malformed ledger structure, including duplicate, false, non-positive,
   out-of-order, gapped, or unknown in-range versions, reported as
   `schema_migration_history_invalid`;
3. for structurally valid nonzero histories, a lineage set other than exactly
   the expected singleton, including missing, empty, wrong, or mixed lineage,
   reported through the existing historical-lineage remediation report;
4. a contiguous history beyond source head, reported as `schema_version_ahead`;
5. a valid prefix behind head, reported as `schema_migration_required` for
   readiness and eligible for `Apply`;
6. an exact prefix through head, classified current; and
7. no nonzero history and no lineage, classified pristine: readiness reports
   `schema_migration_required`, while `Apply` may initialize it.

Structural corruption wins when it coexists with lineage mismatch. Apply
preflight, provider-lock validation, final postcondition, and readiness use this
one classifier. A nil readiness pool and invalid source fail closed. No
automatic ledger repair or historical-lineage bridge is added. Migrate and
server facades detect `RemediationReporter`; server maps failures to its current
diagnostic envelope without making Database Migrations depend on
`platform/config`.

### 15.5 PostgreSQL and Network Flow boundary

`internal/platform/postgres` retains only its narrow `DB` port, settings and
binding resolution, connection construction, environment-key derivation, and
telemetry. `internal/platform/postgres/.gitkeep`, `TransactionRunner`,
`NewTransactionRunner`, and their test are deleted. The managed-service DSN
prefix/suffix become private. The `CARTULARY_POSTGRES_DSN` test key moves to
test support and is not a production package constant.

Network Flow replaces `ModuleDependencies.Transactions` and
`WithTransactionRunner` with a private helper over its existing `postgres.DB`.
Server composition no longer constructs or injects the redundant facade. Its
lifecycle remains begin, callback, rollback attempt after callback failure,
commit attempt after callback success, defensive deferred rollback after
successful commit, surfaced commit failure, and successful return only after
commit. The helper MUST preserve the original primary error when a rollback
attempt also fails.

## 16. Specification and Verification Admission

N-01 lands normative authority before code is contracted. Active replacement
rows and implementation guidance land only with the code and executable tests
they describe. Identifier availability was confirmed again at
`d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`.

### 16.1 Owner-document changes

- Amend `REQ-01-657` in Core 01 rather than adding a competing requirement. It
  assigns `database_migrations` only production apply-to-head execution and
  requires an embedded fail-closed source, advisory serialization, safe
  failures, exact lineage, and prefix-valid ledgers.
- Amend `AC-537` in Core 04. Remove empty-source and production targeted-operation
  expectations; add source rejection before database access, cross-process
  locking, cancellation and bounded unlock, safe failure disclosure, exact
  history/lineage, single remediation emission, and borrowed-handle criteria.
- Add snapshot-available `TH-HARNESS-REQ-810` and `TH-HARNESS-AC-094` to the
  Testing Harness NLSpec. They own the opaque canonical-source targeted
  migration capability only for disposable scratch databases, invalid-target
  before-access behavior, and preparation-event authority.
- Update `docs/guides/cartulary-dev-guide.md` and `AGENTS.md` only in the
  implementation checkpoint where their described production/test or
  PostgreSQL/Network Flow boundary lands. Leave Core 00 and `docs/domain.md`
  unchanged.

### 16.2 Verification-row transition

Each replacement row is admitted atomically with its executable replacement
test and generated topology. No active row may select a nonexistent or
deliberately failing future test, and old and new rows MUST NOT coexist at a
checkpoint.

| Retired row | Final row or rows | Final selector intent |
| --- | --- | --- |
| `module.database_migrations.integration.long_running_migration_cancellation` | `module.database_migrations.integration.concurrent_apply_locking` | Cross-process serialization, lock-wait cancellation, provider cancellation, and bounded release integration cases. |
| `module.database_migrations.integration.migration_lineage_remediation_matrix` | `module.database_migrations.integration.production_preflight_state_matrix` | Exact current, missing/wrong/mixed lineage, and invalid ledger-prefix integration cases. |
| `module.database_migrations.unit.defaults_logger_and_remediation_contract` | `module.database_migrations.unit.fail_closed_source_and_remediation_contract` | Zero/nil/empty/malformed source, nil readiness, and exact internal-facts remediation JSON. |
| `module.database_migrations.unit.provider_context_source_logger_isolation` | `module.database_migrations.unit.provider_source_and_safe_error_isolation` | Embedded root/provider isolation and safe error-redaction unit evidence. |
| `module.database_migrations.unit.targeted_operation_validation` | `module.database_migrations.unit.test_harness_targeted_operation_validation` | Exact pgtest-package target validation; the semantic owner remains Database Migrations while the executable helper is harness-only. |
| `platform.postgres.unit.transaction_lifecycle` | `module.networkflow.unit.transaction_lifecycle` | Network Flow's private helper and preserved begin/callback/rollback/commit ordering. |

Unchanged Database Migrations rows, all source-owner migration row identities,
and `app.migrate`, `app.server`, and `app.operator` facade rows retain their IDs.
The final manifests MUST have 12 Database Migrations rows, two PostgreSQL rows,
and one additional Network Flow row unless collision/selector discovery records
and approves an explicit tracker amendment before implementation. No executable
crosswalk remains after N-08.

## 17. Ordered Iteration-Two Workstreams

Each workstream is an indivisible checkpoint. After a slice is complete, update
Sections 18 through 20 with changed paths, commands, run roots, related and
unrelated failures, blockers, and exact next action before starting its
successor. Run `git diff --check` after every checkpoint and
`make lint-markdown` whenever the tracker or an owner document changes. A failed
slice reopens that slice; database rollback is never a code rollback strategy.

| Slice | Depends on | Work and checkpoint | Principal risk and rollback | Exit criteria |
| --- | --- | --- | --- | --- |
| N-00 — Rebaseline and prior closure | None | Recount the current tree, close IG-004/PGT-017/PGT-AC-012 from supplied evidence, and append this second-iteration decision record. Change only this tracker. | Misstated history or snapshot. Correct the tracker; do not touch product state. | Tracker-only diff; current inventory and decisions are explicit; Markdown and diff checks pass. |
| N-01 — Authority and tracker reconciliation | N-00 | Revalidate the tracker at current HEAD; adopt Core 01, Core 04, and Testing Harness authority; reserve and collision-check final IDs. Do not create active replacement rows or update implementation guides. | Owner contradiction or identifier collision. Revert the specification set and record the blocker. | One non-contradictory production/test contract; IDs are free; Markdown and diff checks pass; tracker checkpoint is appended. |
| N-02 — Retained baseline | N-01 | At one unchanged product snapshot retain Database Migrations, PostgreSQL, Network Flow, migrate, server, operator, pgtest topology, Platform Jobs, Platform Audit, Indicators, Entities, Evidence, Extensions, Graph Projection, Incident Bundles, Reference Data, and Saved Views results, plus the browser baseline. | Missing evidence prevents attribution. Rerun the missing exact owner; broad results do not substitute. | Every changing test and caller group has a successful named root. |
| N-03 — Test mechanics relocation | N-02 | Add the opaque pgtest migration capability, convert all 32 calls and affected raw-DB access, delete production targeted operations/tests, replace the targeted-validation row, regenerate topology, and update Harness/development guidance. | Wrong migration phase or scratch lifetime. Revert capability, callers, row, and generated output together. | Exact 26/6 conversion; no retired targeted symbol/import; affected owner and pgtest integration evidence passes. |
| N-04 — Migration API contraction | N-03 | Introduce opaque `Source` and error-only `Apply`; update `db/migrations.Source()` and all production callers; remove status, old constructors, path/default/empty-source behavior, production targeted operations, log environment/file creation, `PreparationStatus`, and the harness environment shim; update recovery restore to use the canonical source. | Startup, CLI, recovery, or provider resource behavior may drift. Revert this entire API slice to the N-03 checkpoint, without aliases. | Public-surface/stale-symbol scans are empty, source validation is fail-closed, preparation events remain, recovery behavior passes, and all three application builds pass. |
| N-05 — Locking and safe failure handling | N-04 | Implement the Section 15.3 lock lifecycle and safe stage/reason errors; use a discard Provider logger; simplify remediation transport and exact migrate stderr; add concurrency, cancellation, cleanup, borrowed-handle, and leak tests. | Session locks can leak, errors can hide primary failures, or diagnostics can disclose secrets/SQL. Revert only this slice to the validated N-04 apply implementation. | Repeated concurrent stress passes; all exits release the lock or report bounded cleanup failure; borrowed handles stay open; forbidden-data scans and exact stderr pass. |
| N-06 — Readiness and remediation hardening | N-05 | Share validated catalog/history classification across preflight and readiness; require exact prefix and exact lineage; replace exported/open remediation facts; remove incident/nil compatibility; add the full state matrix. | A classifier difference between apply and readiness can admit inconsistent state. Revert classifier and tests as a unit. | Invalid source/nil pool/missing-wrong-mixed lineage/duplicate-gap-unknown/behind-ahead-current cases return only their specified reason/report outcomes. |
| N-07 — PostgreSQL and Network Flow cleanup | N-06 | Delete `.gitkeep`, private the DSN fragments, move the test DSN key, delete the shared transaction runner/test, localize the helper to Network Flow, remove dependency options/server assembly, and complete the row transition. | Transaction error precedence or resource ordering may change. Revert the helper, wiring, and row together. | PostgreSQL has only its retained responsibilities; Network Flow exact lifecycle evidence, adapter tests, boundary checks, and server build pass. |
| N-08 — Boundary and generation reconciliation | N-07 | Update rooted-filesystem allowances, backend boundaries, family/owner inputs, and generated topology; run `make generate`; remove temporary crosswalks; scan for old symbols/IDs/paths/constants/env logging/aliases. | Overbroad allowlists or hand-edited projections can conceal stale topology. Revert authored inputs and their Make-generated projections together. | Zero stale references; JSON, harness, boundary, generation drift, artifact policy, and migration drift are clean; SQL diff is empty. |
| N-09 — Validation and handoff | N-08 | Execute Section 19 at one product-source snapshot; run finalization and broad/release gates; close every second-iteration criterion and record all evidence. The final evidence-only tracker append does not invalidate product results. | Aggregate success can mask an omitted owner, or a tracker edit can be mistaken for product drift. Reopen the narrow failing slice; never waive an applicable gate. | All PGT-AC-013 through PGT-AC-022 pass with named roots, no unexplained diff/artifact, and no compatibility or SQL change. |

## 18. Iteration-Two Work Tracker

| ID | Slice | Status | Depends on | Required artifact or evidence | Next action |
| --- | --- | --- | --- | --- | --- |
| PGT-018 | N-00 — Rebaseline and prior closure | DONE | None | Sections 13 through 20; tracker-only path inspection; Markdown root `20260808T213800Z-p749418`; clean diff check. | Stop this document-only session; N-01 is the next separately checkpointed slice. |
| PGT-019 | N-01 — Authority and tracker reconciliation | DONE | PGT-018 | Revalidated snapshot; adopted Core 01, Core 04, and Testing Harness owner text; unique reserved IDs; Markdown root `20260809T001329Z-p1999360`; clean diff check. | Begin N-02 retained baseline at the unchanged product snapshot. |
| PGT-020 | N-02 — Retained baseline | DONE | PGT-019 | Successful unit and service-backed owner roots, pgtest topology root `20260809T003124Z-p2441154`, and browser root `20260809T003211Z-p2445605`. | Begin N-03 test-mechanics relocation. |
| PGT-021 | N-03 — Test mechanics relocation | DONE | PGT-020 | Opaque pgtest capability and exact 26/6 disposition; targeted row and generated topology; owner roots; backend-integration root `20260809T005219Z-p2663445`; clean contract/drift scans. | Begin N-04 source and Apply contraction. |
| PGT-022 | N-04 — Migration API contraction | DONE | PGT-021 | Immutable fallible source, error-only apply, caller/recovery conversion, preparation-event simplification, final public-surface scans, owner roots, backend integration, builds, and contract/drift evidence. | Begin N-05 locking and safe failure handling. |
| PGT-023 | N-05 — Locking and safe failure handling | DONE | PGT-022 | Bounded session-lock lifecycle, validating execution locker, safe failure interfaces, exact facade output, repeated concurrency/cancellation roots, cleanup/postcondition tests, security/static/build evidence. | Begin N-06 readiness and remediation hardening. |
| PGT-024 | N-06 — Readiness and remediation hardening | DONE | PGT-023 | Shared pure classifier, SQL/pgx snapshots, production preflight/readiness matrices, exact private typed remediation JSON, and facade diagnostic evidence. | Begin N-07 PostgreSQL and Network Flow cleanup. |
| PGT-025 | N-07 — PostgreSQL and Network Flow cleanup | DONE | PGT-024 | Two-row PostgreSQL adapter inventory, owner-private Network Flow transaction helper and row, exact lifecycle tests, boundary and server build evidence. | Begin N-08 boundary and generation reconciliation. |
| PGT-026 | N-08 — Boundary and generation reconciliation | DONE | PGT-025 | Removed temporary boundary exception; exact six-row transition; stale scans; generated, JSON, harness, artifact, boundary, and migration evidence. | Begin N-09 final validation and handoff. |
| PGT-027 | N-09 — Validation and handoff | DONE | PGT-026 | Final product-source digest `e745f9a901556a638d251102a3410ff8a5d1f6534695df45b0125a5d19023ecf`; focused, static, security, build, browser, broad, finalization, and release evidence; closed PGT-AC-013 through PGT-AC-022. | Stop; iteration two is complete. |

### Iteration-two checkpoint log

| Time | Slice | Current state | Changed paths | Commands and retained roots | Failures or blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-08-08 EDT | N-00 | Prior iteration closed; current inventory and decision-complete second iteration admitted. | Only `docs/handoffs/postgres-module-refactor-tracker.md`; `docs/domain.md` and every product/specification/test/generated path are unchanged. | HEAD and tracked-path inventory; import/call/row/ID scans; owner discovery for Database Migrations and PostgreSQL; `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260808T213800Z-p749418`. | No failure or blocker. This evidence append receives one final Markdown/diff revalidation before handoff. | Stop this document-only session. Begin N-01 later as a separate authority-and-verification checkpoint. |
| 2026-08-08 EDT | N-01 | Authority and tracker reconciliation complete at `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`; domain vocabulary unchanged. | `docs/handoffs/postgres-module-refactor-tracker.md`, Core 01, Core 04, and `docs/testing-harness-nlspec.md`; product, test, verification-row, generated, SQL, guide, `AGENTS.md`, and `docs/domain.md` paths unchanged. | Revalidated 8 PostgreSQL paths, 14 Database Migrations paths, 35 imports across 34 Go files, 60 migrations, 32 substantive targeted calls, and 12/3 owner rows; unique definition scans for `REQ-01-657`, `AC-537`, `TH-HARNESS-REQ-810`, and `TH-HARNESS-AC-094`; `git diff --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260809T001329Z-p1999360`. | No failure or blocker. The revalidation corrected the supplied “35 files” shorthand to 35 declarations across 34 files; no active row or implementation guide was changed early. | Run final Markdown/diff hygiene for this append, then begin N-02 retained baseline without changing product files. |
| 2026-08-08 EDT | N-02 | Complete retained pre-change baseline at the unchanged product snapshot; domain vocabulary unchanged. | Evidence only plus this tracker checkpoint; product, specification, test, verification, generated, SQL, guide, `AGENTS.md`, and `docs/domain.md` inputs unchanged from N-01. | Unit roots: Database Migrations `20260809T001517Z-p2004614`, PostgreSQL `20260809T001532Z-p2007116`, Network Flow `20260809T001535Z-p2007628`, migrate `20260809T001711Z-p2040206`, server `20260809T001714Z-p2040554`, operator `20260809T001744Z-p2064190`, Jobs `20260809T001807Z-p2084778`, Audit `20260809T001829Z-p2086935`, Indicators `20260809T001840Z-p2088469`, Entities `20260809T001853Z-p2091704`, Evidence `20260809T002013Z-p2123199`, Extensions `20260809T002109Z-p2151697`, Graph Projection `20260809T002141Z-p2175793`, Incident Bundles `20260809T002156Z-p2177583`, Reference Data `20260809T002212Z-p2180250`, and Saved Views `20260809T002240Z-p2203238`. Service-backed roots in the same order excluding PostgreSQL and migrate: `20260809T002334Z-p2230944`, `20260809T002347Z-p2232748`, `20260809T002515Z-p2261315`, `20260809T002543Z-p2283464`, `20260809T002602Z-p2302932`, `20260809T002623Z-p2304691`, `20260809T002633Z-p2306133`, `20260809T002644Z-p2307563`, `20260809T002749Z-p2338034`, `20260809T002842Z-p2365334`, `20260809T002911Z-p2388212`, `20260809T002925Z-p2389652`, `20260809T002943Z-p2391219`, and `20260809T003012Z-p2413502`. `make backend-integration` passed 211/211 at `20260809T003124Z-p2441154`, including the pgtest topology work unit. `make browser-e2e` passed 53/53 at `20260809T003211Z-p2445605`. | No failure, retry, or blocker. Unit and service-backed results were retained separately; neither broad backend nor browser success substitutes for an owner root. | Run Markdown/diff hygiene for this checkpoint, then begin N-03 and update the tracker before N-04. |
| 2026-08-08 EDT | N-03 | Targeted history construction exists only behind the pgtest-issued `MigrationDatabase`; production exports are apply-to-head only; domain vocabulary unchanged. | `internal/testutil/pgtest`, Database Migrations runner/contract/readiness/remediation tests, ten source-owner migration test files, four Platform Jobs test files, Platform Audit migration evidence, the Database Migrations family input and generated render index, development guide, and this tracker. `docs/domain.md`, `AGENTS.md`, and all SQL/manifest bytes are unchanged. | Original 26 apply-through calls: 20 became direct capability calls, three scratch initializations became `MigrationDatabaseThroughT`, and three production lineage tests became apply-to-head checks. Original six rollback-through calls: five became direct capability calls and the removed production-rollback remediation test became an apply-to-head remediation check. The new invalid-target test also covers zero/unissued capability behavior before access. Format `20260809T005454Z-p2686184`; generation `20260809T004345Z-p2488506`; Database Migrations unit `20260809T004558Z-p2498696` and service `20260809T004620Z-p2501408`; Jobs through Saved Views service roots `20260809T004634Z-p2503202` through `20260809T005123Z-p2633914`; backend integration 211/211 at `20260809T005219Z-p2663445`; JSON `20260809T005457Z-p2689233`; harness `20260809T005500Z-p2689697`; generation drift `20260809T005503Z-p2690085`; migration drift `20260809T005511Z-p2692829`; boundary `20260809T005519Z-p2695613`; Markdown `20260809T005520Z-p2696005`; diff check passed. | Initial format failed before mutation because the amended selector list was not ASCII-sorted; it was sorted. Initial Database Migrations root `20260809T004356Z-p2490980` exposed a related scratch-pool binding defect: pgx `ConnString()` retained the admin database after field mutation. Pool creation now uses the parsed config with an explicit scratch database, and the exact failed owner slice passes. No unrelated failure or blocker. | Run final Markdown/diff hygiene for this append, then begin N-04 source and Apply contraction. |
| 2026-08-08 EDT | N-04 | Migration source construction is immutable and fail-closed, production apply is error-only, callers propagate source errors, and duplicate preparation status/log-file behavior is gone; domain vocabulary unchanged. | `db/migrations/source.go`; Database Migrations source, readiness/remediation, and contract tests; migrate/server assembly and tests; operator/evidence/recovery callers; pgtest/testservices preparation code and tests; affected source-owner tests; Database Migrations family input and generated render index; development guide; this tracker. `docs/domain.md`, `AGENTS.md`, and all 60 authored SQL files are unchanged. | Generation `20260809T010708Z-p2702686`; final format `20260809T010817Z-p2712260`; Database Migrations unit `20260809T010820Z-p2715326` and service `20260809T011041Z-p2803088`; migrate `20260809T010850Z-p2718227`; server unit/service `20260809T010853Z-p2718612` and `20260809T011054Z-p2804902`; operator unit/service `20260809T010932Z-p2745992` and `20260809T011121Z-p2827030`; recovery unit/service `20260809T010955Z-p2766829` and `20260809T011140Z-p2846464`; backend integration 211/211 at `20260809T011220Z-p2879432`; migrate/server/operator builds `20260809T011428Z-p2908435`, `20260809T011429Z-p2910265`, and `20260809T011440Z-p2921132`; migration drift `20260809T011525Z-p2932102`; generation drift `20260809T011534Z-p2934989`; artifact policy `20260809T011541Z-p2937706`; JSON `20260809T011542Z-p2938101`; harness `20260809T011545Z-p2938571`; boundary `20260809T011548Z-p2939018`; Markdown `20260809T011643Z-p2939635`; diff check passed. Exact scans find no retired source/status/path constructor, log environment/shim, preparation status, or production targeted operation; SQL diff is empty and preparation-event tests pass. | Initial Database Migrations root `20260809T010729Z-p2708390` found two related compile leftovers: one two-result `Apply` assignment and one unused import. Both were corrected, and the exact 14-row slice passes. No unrelated failure or blocker. | Run final Markdown/diff hygiene for this append, then begin N-05 locking and safe failure handling. |
| 2026-08-08 EDT | N-05 | Every production apply now uses the bounded advisory-lock lifecycle and one invocation-local, discard-logging Goose provider; safe migration interfaces drive exact migrate/server output; domain vocabulary unchanged. | Database Migrations failure, lock, apply, remediation, contract, concurrency, and public-surface code/tests; migrate/server facades and tests; Database Migrations, migrate, and server family inputs; generated render index; development guide; this tracker. `docs/domain.md`, `AGENTS.md`, and all authored SQL remain unchanged. | Final generation `20260809T013728Z-p2985281` and format `20260809T013736Z-p2987733`; Database Migrations unit `20260809T013842Z-p2996568` and service `20260809T013903Z-p2999238`; repeated exact concurrency/cancellation roots `20260809T013759Z-p2992240`, `20260809T013811Z-p2993670`, and `20260809T013823Z-p2995099`; migrate facade `20260809T013739Z-p2990813`; server facade `20260809T013742Z-p2991257`; provider/failure isolation `20260809T013746Z-p2991699`; boundary `20260809T013956Z-p3001331`; telemetry `20260809T013957Z-p3001619`; JSON `20260809T014003Z-p3006345`; harness `20260809T014006Z-p3006814`; migration drift `20260809T014009Z-p3007221`; generation drift `20260809T014016Z-p3010097`; artifact policy `20260809T014024Z-p3012845`; migrate/server builds `20260809T014025Z-p3013366` and `20260809T014027Z-p3015407`; targeted Gosec `20260809T014038Z-p3026270`; Markdown `20260809T014556Z-p3084274`; diff check passed. Tests cover shared execution-lock session identity, five-minute/one-second acquisition policy, 30-second detached release, preflight/provider/panic/postcondition cleanup, primary-error precedence, context identity, safe forbidden-data corpus, one-line facade output, and borrowed-handle usability. SQL diff and stale Provider-close/log scans are empty. | Early focused roots `20260809T012638Z-p2955043` through `20260809T013506Z-p2981555` exposed related synthetic-fixture and normalization issues: custom sources lacked lineage, PL/pgSQL blocks lacked Goose statement markers, and canceled execution could be hidden by cleanup. The fixtures and primary-error normalization were corrected. One manual server-row command used a nonexistent row ID; the exact current row then passed. Extra broad root `20260809T014104Z-p3050755` passed 210/211 but hit an unrelated Imports cancellation-convergence timeout; the exact failed row passed at `20260809T014439Z-p3082511`. No N-05 blocker remains; N-09 still requires a successful final broad run. | Run final Markdown/diff hygiene for this append, then begin N-06 readiness and remediation hardening. |
| 2026-08-08 EDT | N-06 | Apply preflight, execution-lock validation, final postcondition, and readiness now use one structural-first migration classifier; remediation transport is private and byte-stable; domain vocabulary unchanged. | Database Migrations state snapshot/classifier, readiness, apply preflight, remediation, contract and matrix tests; server readiness mapping and facade tests; Database Migrations and server family inputs; generated render index; development guide; this tracker. `docs/domain.md`, `AGENTS.md`, owner specifications, and all authored SQL remain unchanged. | Final generation `20260809T020145Z-p3171061` and format `20260809T020415Z-p3207507`; Database Migrations unit `20260809T020202Z-p3176727` and service `20260809T020232Z-p3179953`; server service `20260809T020259Z-p3181910`; migrate/server unit `20260809T020423Z-p3210812` and `20260809T020423Z-p3210853`; boundary `20260809T020423Z-p3211031`; JSON `20260809T020423Z-p3210738`; telemetry `20260809T020504Z-p3238211`; harness `20260809T020504Z-p3238914`; migration drift `20260809T020504Z-p3238418`; generation drift `20260809T020504Z-p3238362`; artifact policy `20260809T020504Z-p3238353`; migrate/server/operator builds `20260809T020505Z-p3239364`, `20260809T020505Z-p3239366`, and `20260809T020505Z-p3239395`; targeted Gosec `20260809T020635Z-p3273908`; Markdown `20260809T020635Z-p3273853`; diff check passed. Matrices cover pristine, zero-only, behind, current, ahead, duplicate, false, negative, repeated-zero, out-of-order, gap, missing/empty/wrong/mixed lineage, lineage without history, combined-failure precedence, nil pool, invalid source, and forged unknown in-range versions. Exact missing/wrong/mixed v1 JSON bytes pass; export/open-map/incident/platform-config/stale-row scans and SQL diff are empty. | Initial unit roots `20260809T015434Z-p3094318` and `20260809T015459Z-p3099840` exposed related compile and synthetic-fixture defects: the SQL reader interface had been removed during extraction, and the postcondition fixture represented an invalid zero-history-with-lineage state. Both were corrected. Initial server root `20260809T015803Z-p3116523` exposed one missing Database Migrations import and one stale test import; both were corrected. No unrelated failure or blocker. | Run final Markdown/diff hygiene for this append, then begin N-07 PostgreSQL and Network Flow cleanup. |
| 2026-08-08 EDT | N-07 | PostgreSQL now owns only its narrow DB/settings/binding/connection/telemetry adapter surface; Network Flow owns its private transaction lifecycle; domain vocabulary unchanged. | Deleted PostgreSQL `.gitkeep`, transaction facade, and facade test; privatized managed-service DSN fragments; moved the generic DSN test key to suiteservices; localized Network Flow transaction code/tests; removed module/server injection; moved the verification row; regenerated the render index; updated `AGENTS.md`, the development guide, and this tracker. `docs/domain.md` and all authored SQL remain unchanged. | Generation `20260809T021029Z-p3301661`; format `20260809T021037Z-p3304132`; PostgreSQL unit `20260809T021103Z-p3307729`; Network Flow unit `20260809T021103Z-p3307735`; server unit `20260809T021103Z-p3307740`; exact lifecycle row `20260809T021328Z-p3390172`; Network Flow/server service `20260809T021336Z-p3390640` and `20260809T021336Z-p3390649`; boundary `20260809T021518Z-p3442114`; JSON `20260809T021518Z-p3441705`; harness `20260809T021518Z-p3442204`; generation drift `20260809T021518Z-p3441716`; artifact policy `20260809T021518Z-p3441751`; migration drift `20260809T021518Z-p3441858`; server build `20260809T021518Z-p3442621`; Markdown `20260809T021613Z-p3459932`; diff check passed. Exact lifecycle evidence preserves begin/callback/rollback/commit order, defensive rollback after successful commit, commit failure, and primary callback error when rollback also fails. Counts are 12 Database Migrations rows, two PostgreSQL rows, and 60 Network Flow rows including the one moved row. Retired facade, injection, constant, row-ID, and test-key ownership scans are empty; the PostgreSQL directory has five retained files and SQL diff is empty. | No related or unrelated failure, retry, or blocker. | Run final Markdown/diff hygiene for this append, then begin N-08 reconciliation. |
| 2026-08-08 EDT | N-08 | Authored boundaries, owner families, test-support ownership, and generated topology now describe only the final iteration-two state; domain vocabulary unchanged. | Removed the obsolete Database Migrations deployment-config boundary allowlist; revalidated all family/owner inputs; regenerated the execution-topology render index; updated this tracker. `docs/domain.md`, product behavior, `AGENTS.md`, specifications, guides, and all authored SQL remain unchanged from N-07. | Final generation `20260809T021741Z-p3462258` and format `20260809T021749Z-p3464733`; boundary `20260809T021759Z-p3468410`; JSON `20260809T021758Z-p3468059`; harness `20260809T021759Z-p3468506`; generation drift `20260809T021759Z-p3468082`; artifact policy `20260809T021759Z-p3468120`; migration drift `20260809T021759Z-p3468182`; Markdown `20260809T021919Z-p3475659`; diff check passed. Global family-row IDs are unique. The exact manifest delta is five Database Migrations row replacements plus removal/addition of the PostgreSQL-to-Network-Flow lifecycle row; all unaffected row identities remain. Counts are 12 Database Migrations, two PostgreSQL, and 60 Network Flow rows. Retired IDs/symbols/paths/aliases/constants/log-environment and production targeted-operation scans are empty. All 60 tracked SQL files have an empty worktree diff and ordered filename/content aggregate SHA-256 `089cdbab837bee70ae09e710f9015dc66ad1198d6f55b499c3a47e84de910605`; lineage remains `cartulary.prod_ddl_rebaseline.v1` at boundary `prod_ddl_rebaseline_v1`. | No related or unrelated failure, retry, or blocker. | Run final Markdown/diff hygiene for this append, then begin N-09 discovery and single-snapshot final validation. |
| 2026-08-09 EDT | N-09 | Iteration-two validation and handoff are complete at one final product-source snapshot; domain vocabulary unchanged. | Evidence-only final tracker closure plus three validation-stability repairs: deterministic outside hover and a load-tolerant E2E observation ceiling, a load-tolerant Imports convergence ceiling, and a load-tolerant server-process HTTP ceiling. The new Network Flow fallback error was lowercased for staticcheck. `docs/domain.md`, all authored SQL, migration lineage, production behavior beyond the planned slices, and generated roots other than the Make-generated execution-topology render index are unchanged. | Section 19.4 records discovery, every focused root, repeated lock rows, static/security/build roots, finalization, browser/broad/release roots, failures, retries, skipped retained-run maintenance, source/SQL digests, row reconciliation, and final scans. Final broad roots are browser `20260809T034016Z-p804709`, fast `20260809T034327Z-p839765`, full `20260809T035917Z-p1071026`, check `20260809T040702Z-p1163900`, and release `20260809T040915Z-p1264344`. | Related N-09 findings were repaired and revalidated. Unrelated load-sensitive Imports, server-process, incident E2E, and collaboration E2E failures are retained and classified in Section 19.4; every required final gate subsequently passed. No blocker remains. | Stop; N-09 and iteration two are complete. |

## 19. Iteration-Two Validation and Handoff

### 19.1 Per-slice rule

Run the narrowest affected owner rows after each implementation slice. Before
choosing a command, inspect the current Make-owned task surface. Record the
result root, including related discovery failures and their repairs, in Section
18 before advancing. Documentation-only authority checkpoints use
`make lint-markdown` and `git diff --check`; product, generated, conformance,
and release checks begin only when their owning implementation or final slice
requires them.

### 19.2 Final validation order

1. Owner and mapping discovery:

   - `make task-guide ROLE=module-author OWNER=module.database_migrations`
   - `make explain-test-owner OWNER=module.database_migrations`
   - `make task-guide ROLE=module-author OWNER=platform.postgres`
   - `make explain-test-owner OWNER=platform.postgres`
   - `make task-guide ROLE=module-author OWNER=module.networkflow`
   - `make explain-test-owner OWNER=module.networkflow`
   - `make explain-target TARGET=<affected-target> DETAIL=rows`
   - exact symbol, row-ID, path, owner, and selector collision scans

2. Focused behavior:

   - unit and service-backed slices, as applicable, for
     `module.database_migrations`, `platform.postgres`, `module.networkflow`,
     `app.migrate`, `app.server`, `app.operator`, `platform.jobs`, and every
     source owner converted in N-03
   - repeated exact concurrent-locking and cancellation rows
   - `make test` for `internal/testutil/pgtest` and other topology-selected test
     utilities

3. Static, contracts, security, migration, generation, and builds:

   - `make backend-module-boundary-check`
   - `make otel-conformance`
   - `make json-shape-check`
   - `make harness-contract`
   - `make migration-drift`
   - `make generate-drift`
   - `make generated-artifact-policy-check`
   - `make build-migrate`
   - `make build-server`
   - `make build-operator`
   - `make go-gosec-targeted`
   - `make lint-markdown`
   - `git diff --check`

4. Finalize before broad verification:

   - `make agent-finalize`; pass `RESULTS_DIR` only for a successful compatible
     same-snapshot warm full-check root, otherwise leave it unset and record the
     skipped retained-run maintenance

5. Broad completion:

   - `make browser-e2e`
   - `make test-fast`
   - `make test`
   - `make check`
   - `make release-check`

Browser validation is included because the user requires it for final
repository readiness, even though no browser implementation is planned. A
browser failure is recorded and classified; it is not waived as irrelevant.

### 19.4 Final N-09 evidence

#### Snapshot and workstream reconciliation

- N-00 planning began from `3be8ddda38deb7b28734e78e487dc229be300585`;
  N-01 revalidated and implementation began from clean HEAD
  `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`. The final base commit remains
  `d2b70af6c1e1bed34cd3ab53744344ad8d2ad44c`; the complete final non-tracker
  diff plus intended new-source content has SHA-256
  `e745f9a901556a638d251102a3410ff8a5d1f6534695df45b0125a5d19023ecf`.
- N-01 changed Core 01, Core 04, and Testing Harness authority. N-02 retained
  the pre-change owner and browser baseline. N-03 moved all targeted history
  construction to the opaque pgtest capability. N-04 contracted source,
  apply, preparation, and caller APIs. N-05 added the bounded locked provider
  lifecycle and safe failure interfaces. N-06 added the shared structural-first
  classifier and private remediation projection. N-07 removed PostgreSQL dead
  exports and moved transaction ownership to Network Flow. N-08 reconciled
  boundaries, owner rows, and the generated render index. N-09 changed only
  validation stability and the staticcheck-safe Network Flow fallback text
  before retaining the final matrix. The individual Section 18 checkpoints
  contain the exact paths for every workstream.
- The final production API is the Section 15.1 API. Tests prove opaque,
  immutable, fail-closed sources; apply-to-head-only production behavior;
  pgtest-only targeted execution; bounded same-session locking; safe closed
  reason codes; byte-stable remediation JSON; structural-first readiness; and
  borrowed database ownership. PostgreSQL now owns connection/query/telemetry
  plumbing only, while Network Flow privately owns its transaction lifecycle.

#### Discovery and focused behavior

- `make task-guide ROLE=module-author` and `make explain-test-owner` succeeded
  for `module.database_migrations`, `platform.postgres`, and
  `module.networkflow`. `make explain-target ... DETAIL=rows` resolved every
  final affected target. The final topology is 12 Database Migrations rows,
  two PostgreSQL rows, and 60 Network Flow rows.
- Final-matrix unit roots: Database Migrations
  `20260809T022114Z-p3484005`; PostgreSQL `20260809T022114Z-p3484001`;
  Network Flow `20260809T022114Z-p3484010`; migrate
  `20260809T022114Z-p3484019`; server `20260809T022247Z-p3516109`;
  operator `20260809T022247Z-p3516104`; Jobs
  `20260809T022247Z-p3516110`; Audit `20260809T022247Z-p3516119`;
  Indicators `20260809T022330Z-p3563967`; Entities
  `20260809T022330Z-p3563966`; Evidence `20260809T022330Z-p3563975`;
  Extensions `20260809T022330Z-p3563974`; Graph Projection
  `20260809T022513Z-p3659681`; Incident Bundles
  `20260809T022513Z-p3659690`; Reference Data
  `20260809T022513Z-p3659680`; and Saved Views
  `20260809T022513Z-p3659700`.
- Final-matrix service-backed roots: Database Migrations
  `20260809T022625Z-p3720943`; Network Flow `20260809T022625Z-p3720950`;
  server `20260809T022625Z-p3720957`; operator
  `20260809T022757Z-p3773536`; Jobs `20260809T022757Z-p3773534`;
  Audit `20260809T022757Z-p3773545`; Indicators
  `20260809T022757Z-p3773552`; Entities `20260809T022831Z-p3797739`;
  Evidence `20260809T022831Z-p3797746`; Extensions
  `20260809T022831Z-p3797731`; Graph Projection
  `20260809T022831Z-p3797729`; Incident Bundles
  `20260809T022948Z-p3879492`; Reference Data
  `20260809T022948Z-p3879491`; and Saved Views
  `20260809T022948Z-p3879504`.
- The exact pgtest topology row passed at `20260809T023046Z-p3930127`.
  Exact concurrent-apply serialization and cancellation passed three
  consecutive times at `20260809T023053Z-p3930539`,
  `20260809T023109Z-p3932027`, and `20260809T023128Z-p3933550`.
  After the final staticcheck repair, the exact Network Flow lifecycle row
  passed at `20260809T033930Z-p799527`.

#### Static, security, build, finalization, and broad gates

- Final-snapshot roots: boundary `20260809T041848Z-p1436148`; telemetry
  `20260809T041849Z-p1436442`; JSON `20260809T041854Z-p1440277`; harness
  `20260809T041857Z-p1440746`; migration drift
  `20260809T041900Z-p1441160`; generation drift
  `20260809T041906Z-p1443859`; artifact policy
  `20260809T041913Z-p1446582`; migrate build
  `20260809T041915Z-p1447100`; server build
  `20260809T041916Z-p1448918`; operator build
  `20260809T041926Z-p1459766`; targeted Gosec
  `20260809T041935Z-p1470604`; and vulnerability analysis
  `20260809T041943Z-p1494488`. The completed tracker draft passed Markdown
  lint at `20260809T042222Z-p1495792`; `git diff --check` also passed.
- `make agent-finalize` passed after the final production-source edit at
  `20260809T033949Z-p801954`. `RESULTS_DIR` was unset, so retained successful
  warm-run maintenance was intentionally skipped; no compatible same-snapshot
  warm full-check root existed at finalization time.
- Final broad roots on the unchanged product-source digest are
  `make browser-e2e` 53/53 at `20260809T034016Z-p804709`,
  `make test-fast` 353/353 at `20260809T034327Z-p839765`,
  `make test` 876/876 at `20260809T035917Z-p1071026`,
  `make check` 744/744 at `20260809T040702Z-p1163900`, and
  `make release-check` 915/915 at `20260809T040915Z-p1264344`.

#### Failures, retries, and final integrity

- The first full run, `20260809T023911Z-p4101301`, exposed an incident E2E
  pointer-leave race. Its first exact reproduction failed at
  `20260809T025018Z-p53233`; using a stable outside locator and retaining the
  exact 5,000 ms fake-timer contract produced exact pass
  `20260809T025408Z-p80098`. The next full run,
  `20260809T030025Z-p191717`, exposed aggregate-load ceilings in Imports,
  server-process reset, and that same real-browser timer. Their isolated rows
  passed at `20260809T031603Z-p371261`, `20260809T031633Z-p372807`, and
  `20260809T031459Z-p347297`; load-tolerant evidence ceilings then passed at
  `20260809T031718Z-p395480`, `20260809T031733Z-p396995`, and
  `20260809T031804Z-p417303`. A parallel Imports diagnostic root
  `20260809T031459Z-p347281` hit an unrelated PostgreSQL deadlock and was not
  used as completion evidence.
- A post-stability full run passed 876/876 at
  `20260809T032418Z-p528414`. The following check root
  `20260809T033628Z-p680225` found the related lowercase-error staticcheck
  defect; the exact owner row and staticcheck then passed before finalization.
  The first final-snapshot full attempt, `20260809T034533Z-p893855`, hit an
  unrelated load-sensitive Collaboration offline-projection race; its exact
  work unit passed 12/12 at `20260809T035801Z-p1048366`, and the unchanged
  aggregate retry passed 876/876. No failed browser or other applicable gate is
  being waived.
- All six row transitions occurred exactly once; unaffected identities remain;
  global row IDs are unique. Retired row IDs, source/status/logger APIs,
  production targeted operations, compatibility aliases, forwarding paths,
  PostgreSQL transaction facade/injection, and stale selectors have zero live
  references. The only generated diff is the Make-generated
  `tools/execution_topology_render_index.json` projection.
- All 60 tracked SQL filenames and bytes have an empty worktree diff and the
  same ordered filename/content aggregate SHA-256
  `089cdbab837bee70ae09e710f9015dc66ad1198d6f55b499c3a47e84de910605`.
  Lineage remains `cartulary.prod_ddl_rebaseline.v1` at boundary
  `prod_ddl_rebaseline_v1`. `docs/domain.md` is unchanged. Every untracked
  source is an intentional new implementation/test path named by its
  workstream; no untracked generated, runtime, or product artifact and no
  unexplained diff remains.

### 19.3 Required final handoff

N-09 MUST record the planning and product snapshot commits; changed paths
grouped by slice; substantive specification, API, security, and ownership
changes; every command and result/run root; related versus unrelated failures
and retries; every skipped check with its reason; retained SQL filename/byte/hash
and lineage evidence; row transition reconciliation; generated projection
scope; and confirmation that no compatibility alias, forwarding package,
deprecated operation, old identifier, stale selector, SQL edit, or untracked
product artifact remains. An applicable failed gate keeps PGT-027 open.

## 20. Iteration-Two Binary Acceptance Criteria

| Acceptance ID | Required postcondition | Required evidence | Status |
| --- | --- | --- | --- |
| PGT-AC-013 | The tracker is rebaselined to `3be8ddda38deb7b28734e78e487dc229be300585`; IG-004, PGT-017, and PGT-AC-012 are closed from named successful evidence; the exact current inventory and decision-complete N-01 through N-09 plan are recorded; only this tracker changes. | N-00 inventory commands, path-scoped diff, Markdown lint root `20260808T213800Z-p749418`, and `git diff --check`. | PASS |
| PGT-AC-014 | Amended `REQ-01-657` and `AC-537` plus new `TH-HARNESS-REQ-810` and `TH-HARNESS-AC-094` define one non-contradictory production/test boundary; active rows and implementation guides remain truthful until their implementation checkpoints. | Owner/identifier review, scoped-diff review, Markdown root `20260809T001329Z-p1999360`, and clean diff check. | PASS |
| PGT-AC-015 | The production export set is exactly Section 15.1; `Source` is opaque and fail-closed; `Apply` is error-only; no retired source/status/logger/targeted symbol or compatibility surface exists. | Public-surface method-set tests, invalid-source before-access tests, exact stale-symbol scans, and three builds. | PASS |
| PGT-AC-016 | All 32 targeted call sites use the two canonical-source disposable-database pgtest helpers; invalid targets fail before access; `PreparationStatus` and the environment shim are absent. | Converted-call inventory, pgtest target-validation row, preparation event tests, and stale import/environment scans. | PASS |
| PGT-AC-017 | Every production apply uses lock `4097083626` with one-second waits and a five-minute acquisition ceiling; initial, Goose execution-session, and final classifications are locked; detached release is bounded to 30 seconds; and caller-owned DB lifetime is preserved. | Repeated concurrent, cancel-wait, preflight-failure, execution-failure, unlock, session-identity, final-postcondition, and borrowed-handle tests. | PASS |
| PGT-AC-018 | Provider logging is discarded; ordinary and remediation failures expose only safe stable information; migrate emits one remediation JSON object plus LF and no duplicate line. | Unit leak corpus covering SQL/binds/DSNs/paths/server/upstream text, exact stderr bytes, telemetry conformance, and targeted security check. | PASS |
| PGT-AC-019 | Source catalog, applied nonzero ledger rows, and lineage set satisfy the exact validation rules; nil readiness and invalid/empty sources fail closed; all specified reason/report codes are stable. | Production preflight and readiness state matrices plus exact remediation JSON tests. | PASS |
| PGT-AC-020 | PostgreSQL retains only DB/settings/binding/connection/telemetry responsibility; dead constants/files and shared transaction facade are absent; Network Flow privately preserves transaction ordering and error precedence. | Final adapter/export inventory, Network Flow lifecycle row, PostgreSQL/Network Flow/server slices, boundary and build checks. | PASS |
| PGT-AC-021 | Every row transition is reconciled exactly once; unaffected identities remain; authored inputs generate clean projections; all 60 migrations and lineage identifiers are byte-for-byte unchanged; no stale path/ID/alias exists. | Owner explanations, exact scans, JSON/harness/boundary checks, `make generate-drift`, artifact policy, migration drift, and SQL diff. | PASS |
| PGT-AC-022 | Every applicable focused, static, security, build, browser, broad, finalization, and release gate passes at one final product snapshot; the handoff contains complete evidence and the scoped tree has no unexplained change. | Section 19 matrix and final Section 18 checkpoint, including failure classification and retained roots. | PASS |

PGT-AC-013 through PGT-AC-022 are complete. Any future PostgreSQL migration
change begins as a new, separately planned iteration; it does not reopen or
rewrite the retained N-00 through N-09 evidence.
