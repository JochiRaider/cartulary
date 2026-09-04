# Cartulary PostgreSQL 18.6 Pre-Production Migration Tracker

## 1. Scope and Source Posture

### Tracker control

| Field | Value |
| --- | --- |
| Target | Replace Cartulary's PostgreSQL 16 pre-production and fresh-provisioning baseline with exact PostgreSQL 18.6. |
| Output | `docs/handoffs/postgresql-18-6-migration-tracker.md` |
| Tracker status | Migration implementation and validation complete. |
| Allowed change in this session | The user authorized the complete PostgreSQL 18.6 remediation plan. |
| First executable slice | None; SL-00 through SL-09 are complete, with SL-05 retained as the planned `NO_ACTION`. |
| External-source freshness | Sources evaluated through 2026-09-04. |
| Status vocabulary | `TODO`, `READY`, `IN_PROGRESS`, `BLOCKED`, `DEFERRED`, `DONE`, `NO_ACTION`. |

This tracker uses the supplied planning brief as the requested deliverable shape and
`docs/research/nlspec-spec.md` as planning-method guidance. Neither source overrides
an adopted owner. The user's explicit pre-production, no-preservation, hard-cutover
direction closes the deployment posture: Cartulary will create fresh PostgreSQL 18.6
clusters and will not implement a data-preserving PostgreSQL 16 upgrade path.

### Non-goals

- Do not implement `pg_upgrade`, logical replication, dump/restore migration, dual
  clusters, dual readers or writers, compatibility views, lineage translators, or a
  browser-based database upgrade workflow.
- Do not preserve development, test, or pre-production PostgreSQL 16 data directories.
- Do not rewrite, renumber, squash, or replace authored SQL migrations.
- Do not create Production DDL Rebaseline v3 or migration 41 solely because the
  PostgreSQL server version changes.
- Do not adopt UUIDv7, virtual generated columns, temporal constraints, OAuth,
  asynchronous-I/O tuning, or planner tuning as part of the migration.
- Do not refresh unrelated Go, Node, container, or helper-tool dependencies.
- Do not reinterpret retained PostgreSQL 16 benchmark evidence as PostgreSQL 18.6
  evidence.

### Authority hierarchy

1. Adopted subsystem NLSpecs inside their declared scopes.
2. Core 00 through Core 04.
3. Core 05 only for claim-bearing benchmark publication.
4. `docs/domain.md` and implementation-support guides.
5. Current code, configuration, SQL, tests, and generated projections as evidence of
   current implementation state.
6. Previous trackers and handoffs as planning method and historical evidence only.

No contradiction between adopted owners was found. Core 01's statement that the
repository head is migration 37 disagrees with the live migration source and manifest,
which both end at 40. This is owner-document drift to repair in `SL-01`, not permission
to alter migration history.

### Repository baseline

| Fact | Observed value | Evidence |
| --- | --- | --- |
| Branch | `main` | `git branch --show-current` |
| Commit | `6074994b809a4741ffa6655b9267eb42d3566661` | `git rev-parse HEAD` |
| Working tree before tracker creation | Clean | `git status --short` |
| Observation time | 2026-09-04 12:45:08 EDT (`-0400`) | Local `date` |
| Current server baseline | PostgreSQL major 16 | Core 01, Core 04, Testing Harness, Compose, and pgtest |
| DDL lineage | `cartulary.prod_ddl_rebaseline.v2` | Migration 1 and migration evidence |
| Immutable migration boundary | Versions 1 through 29 | `tools/migration_history_manifest.json` |
| Repository migration head | Version 40 | Manifest and `db/migrations/00040_parties_mutation_request_hash_v1.sql` |
| Application schema | `public` | Migration source and owner contracts |
| Goose ledger | `public.goose_db_version` | Database Migrations implementation and tests |
| Lineage relation | `public.schema_migration_lineage` | Migration 1 and readiness checks |
| Extensions | Administrator-owned `pgcrypto` 1.3 and `citext` 1.6 in `public` | Migration 1 and provisioning inputs |
| Credential purposes | `runtime`, `migration`, `recovery` | `internal/platform/postgres` |
| Effective roles | `cartulary_runtime`, `cartulary_schema_owner`, `cartulary_recovery` | Platform PostgreSQL contract and provisioning |

### External primary sources

| Source | Version/date | Repository-relevant conclusion |
| --- | --- | --- |
| [PostgreSQL 17 release notes](https://www.postgresql.org/docs/release/17.0/) | PostgreSQL 17, 2024-09-26 | A major upgrade normally needs dump/restore, `pg_upgrade`, or logical replication; catalog/statistics changes and safer maintenance-function search paths require characterization, but no data-preserving upgrade is in scope. |
| [PostgreSQL 18 release notes](https://www.postgresql.org/docs/release/18.0/) | PostgreSQL 18, 2025-09-25 | `initdb` enables checksums by default; trigger-role, CSV `COPY`, generated-column, time-zone abbreviation, planner, AIO, authentication-warning, and catalog changes were screened against the repository. |
| [PostgreSQL 18.6 release notes](https://www.postgresql.org/docs/release/18.6/) | PostgreSQL 18.6, 2026-08-13 | Exact patched target; PostgreSQL 18.5 was never released. GIN statistics, `btree_gist`/`ltree`, output-plugin, and `pgcrypto` advisories were screened for applicability. |
| [Upgrading a PostgreSQL Cluster](https://www.postgresql.org/docs/18/upgrading.html) | PostgreSQL 18 docs, observed 2026-09-04 | Upgrade mechanisms are intentionally rejected because no PostgreSQL 16 cluster is supported for preservation. |
| [`pg_upgrade`](https://www.postgresql.org/docs/18/pgupgrade.html) | PostgreSQL 18 docs, observed 2026-09-04 | Old/new binaries, extension, platform, and checksum constraints do not justify retaining a migration path for disposable state. |
| [Logical replication restrictions](https://www.postgresql.org/docs/18/logical-replication-restrictions.html) | PostgreSQL 18 docs, observed 2026-09-04 | The repository has no logical publication, subscription, slot, or output-plugin surface to preserve. |
| [Official PostgreSQL container documentation](https://github.com/docker-library/docs/blob/master/postgres/README.md) | Observed 2026-09-04 | PostgreSQL 18 uses version-specific `PGDATA` below `/var/lib/postgresql` and the persistent volume boundary moves to `/var/lib/postgresql`. |
| [Official PostgreSQL container entrypoint](https://github.com/docker-library/postgres/blob/master/docker-entrypoint.sh) | Observed 2026-09-04 | Entrypoint old-directory detection is defense in depth; Cartulary additionally uses versioned volumes and runtime admission. |
| [PostgreSQL 18 `pgcrypto`](https://www.postgresql.org/docs/18/pgcrypto.html) and [`citext`](https://www.postgresql.org/docs/18/citext.html) | PostgreSQL 18 docs, observed 2026-09-04 | Both bundled extensions remain available; Cartulary must prove its exact 1.3/1.6 SQL versions can be installed rather than silently adopting extension defaults. |

The immutable multi-architecture image identity observed from the official registry on
2026-09-04 is:

`docker.io/library/postgres:18.6-alpine@sha256:d3e1620b530c944afa6e887d22eb899824da68e19c52024bf98f5220c88a65b2`

Implementation preflight must confirm that the tag still resolves to this index digest.
A mismatch is a supply-chain blocker; it must not silently refresh the planned digest.

## 2. Executive Migration Summary

Cartulary currently makes PostgreSQL major 16 normative in Core 01, Core 04, and the
Testing Harness, and implements that baseline through two Compose stacks and the
shared pgtest fixture. The selected target is exact PostgreSQL 18.6, identified by the
official multi-architecture image digest above.

The migration is a fresh-only baseline replacement. Existing development, test, and
pre-production database volumes are disposable, are never attached to PostgreSQL 18,
and are not rollback artifacts. The cutover creates new versioned volumes, provisions
the same roles and extension versions, and applies the unchanged migration chain from
1 through 40.

Production DDL Rebaseline v2, its immutable 1-29 boundary, Goose ledger, lineage
relation, application schema, object ownership, and all purpose-separated privileges
remain unchanged. The highest-risk changes are the PostgreSQL 18 container mount
boundary, exact-version admission for borrowed pools, trigger execution-role behavior,
and claim-bearing benchmark identity.

The implementation sequence is owners first, container/provisioning second, common
admission and compatibility third, fixtures and Recovery next, then security,
telemetry, performance, generation, documentation, and broad evidence. `SL-05` closes
the data-preserving path as `NO_ACTION`; it is not an omitted implementation detail.

## 3. Current-State Repository Inventory

### Runtime and supporting surfaces

- `docker-compose.dev.yml` owns the development PostgreSQL service and named volume.
- `deploy/mvp/docker-compose.yml`, `deploy/mvp/postgres-provision.sh`, and package
  scripts own the disconnected/on-prem service topology and provisioning.
- `configs/dev/postgres-provision.sql` provisions development roles, logins,
  extensions, grants, and default privileges.
- `internal/testutil/pgtest` owns the shared and isolated PostgreSQL Testcontainers
  baseline used across module integration tests and test-service sessions.
- `internal/platform/postgres` owns binding resolution, purpose-specific connection
  initialization, role admission, and database telemetry.
- `internal/modules/database_migrations` owns migration lifecycle, preflight,
  readiness, source inspection, and evidence over an already-open database.
- Recovery uses logical, encrypted, row-oriented artifact
  `cartulary.postgres_snapshot_artifact.v2`; it does not invoke PostgreSQL physical
  backup or major-upgrade tools.
- `sqlc.yaml` targets PostgreSQL and pgx/v5. Current dependencies are pgx 5.9.2,
  Goose 3.27.0, testcontainers-go 0.42.0, and SQLC 1.30.0.
- No PostgreSQL service was found in an additional Dockerfile, CI service, Nix file,
  or devcontainer. Those locations remain part of the final residual scan.

### PostgreSQL version-reference inventory

| ID | Path or artifact | Current value | Semantic owner | Runtime or support use | Required target value | Change class | Proposed slice | Validation evidence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| VR-001 | `docs/spec/01_architecture_storage_and_view_contracts.md` | PostgreSQL major 16; repository head 37 | Core 01 | Normative architecture and DDL admission | Exact 18.6; head 40 | `must_change` | SL-01 | Owner review, Markdown lint, migration manifest parity |
| VR-002 | `docs/spec/04_security_deployment_and_conformance.md` | PostgreSQL 16 in provisioning/conformance requirements | Core 04 | Normative security, deployment, Recovery | Exact 18.6 and fresh-only cutover | `must_change` | SL-01 | Owner review and conformance rows |
| VR-003 | `docs/testing-harness-nlspec.md` | PostgreSQL major/fixtures 16 | Testing Harness | Normative fixture topology and AC-095 | Exact 18.6 default fixtures; no ordinary 16 fixture | `must_change` | SL-01 | Harness contract and owner rows |
| VR-004 | `docker-compose.dev.yml` image | `postgres:16-alpine` | Harness local-development surface | Live development service | Exact digest-pinned 18.6 image | `must_change` | SL-02 | Compose rendering, readiness, package-independent persistence test |
| VR-005 | `docker-compose.dev.yml` volume | `postgres-data:/var/lib/postgresql/data` | Harness local-development surface | Persistent development database | `postgres-data-v18:/var/lib/postgresql`; explicit target `PGDATA` | `must_change` | SL-02 | Mount/ownership/restart test and stale-volume rejection |
| VR-006 | `deploy/mvp/docker-compose.yml` image | `postgres:16-alpine` | Core 04 deployment package | Live packaged service | Exact digest-pinned 18.6 image | `must_change` | SL-02 | `make standup-package-smoke` |
| VR-007 | `deploy/mvp/docker-compose.yml` volume | `cartulary-postgres-data:/var/lib/postgresql/data` | Core 04 deployment package | Persistent packaged database | `cartulary-postgres-data-v18:/var/lib/postgresql`; explicit target `PGDATA` | `must_change` | SL-02 | Package smoke and restart persistence |
| VR-008 | `internal/testutil/pgtest/pgtest.go` | `postgres:16-alpine` | pgtest/test services | Default service-backed fixture | Exact digest-pinned 18.6 image | `must_change` | SL-04 | All owner-routed service-backed slices |
| VR-009 | `internal/modules/database_migrations/rebaseline_manifest_integration_test.go` | Major equals 16 | Database Migrations | Manifest characterization | Exact `server_version_num=180006`; checksum on | `must_change` | SL-03 | Database Migrations integration row |
| VR-010 | `internal/testutil/testcontainersx/testcontainersx_test.go` | Six `postgres:16-alpine` fake request values | testcontainersx | Unit-only generic container request fixtures | Retain and classify as fixture-only, or use a non-versioned fake if residual policy requires; never pull | `intentional_no_action` | SL-04 | Unit row proves no runtime resolution |
| VR-011 | `tools/release-evidence/check-standup-package-smoke.sh` | Requires `/var/lib/postgresql/data` | Core 04 release evidence | Package topology assertion | Require parent mount and exact effective `PGDATA` | `must_change` | SL-02 | `make standup-package-smoke` |
| VR-012 | Core 05 benchmark profile | `cartulary.bench.ubuntu_24_04_postgres.2026q1` | Core 05 | Claim-bearing environment identity | Add `cartulary.bench.ubuntu_24_04_postgres_18_6.2026q3`; preserve old evidence | `must_change` | SL-01/SL-07 | Claim checker and qualified comparison |
| VR-013 | `tools/release-evidence/check-benchmark-claim.mjs` | 2026q1 PostgreSQL image ID | Core 05 projection | Claim admission | New 18.6 profile projection without rewriting retained evidence | `generated_or_projection` | SL-08 | `make benchmark-claim-check` |
| VR-014 | `tools/release-evidence/tests/test-benchmark-claim-check.sh` | 2026q1 profile fixture | Core 05 verification | Claim checker fixture | Add exact 18.6 positive/old-profile historical cases | `must_change` | SL-07 | Benchmark claim test target |
| VR-015 | `docs/spec/B_architecture_diagrams_and_explanatory_source_extract.md` | Old data mount and 2026q1 profile example | Non-normative archive | Historical explanatory material | Preserve; label historical only if current text could mislead | `intentional_no_action` | SL-08 | Residual scan classification |
| VR-016 | `docs/handoffs/links-module-refactor-tracker.md` | Link titled PostgreSQL 16 constraints | Historical handoff | Prior research citation | Preserve historical link | `intentional_no_action` | SL-09 | Residual scan classification |
| VR-017 | `tools/harness/services/tests/test-local-session.mjs` | `postgres:17.5-alpine` | Test-service harness | Unit fixture for image identity/digest behavior | Preserve as an explicitly fake/fixture-only value; it is not a target baseline | `intentional_no_action` | SL-04 | Harness unit tests |
| VR-018 | Migration 1, development/MVP provisioners, pgtest provisioner, migration admission | `pgcrypto` 1.3 and `citext` 1.6 in `public` | Core 01/Core 04 | Required extension baseline | Unchanged exact versions and schema | `must_test` | SL-02/SL-03 | Fresh install plus missing/wrong version/schema cases |
| VR-019 | `tools/migration_history_manifest.json` and authored migrations | Immutable through 29; head 40 | Core 01/Database Migrations | Production DDL source and evidence | Byte-for-byte unchanged | `intentional_no_action` | SL-03/SL-09 | `make migration-drift`; hash diff |
| VR-020 | Repository-wide physical-upgrade and replication search | No `pg_dump`, `pg_restore`, `pg_basebackup`, `pg_upgrade`, WAL archive, logical slot, or output-plugin implementation | Core 04/Recovery | Absence of physical upgrade path | Remain absent | `not_present` | SL-05/SL-09 | Content scan and Recovery evidence |

Extension type and function references in migrations and application queries are not
server-version assertions. They remain governed by VR-018 and must not be mechanically
rewritten.

## 4. Target-State Contract

### Exact service identity and storage

| Profile | Current image | Target image | Current volume mapping | Target volume mapping | Effective `PGDATA` |
| --- | --- | --- | --- | --- | --- |
| Development | `postgres:16-alpine` | Exact 18.6 index digest | `postgres-data:/var/lib/postgresql/data` | `postgres-data-v18:/var/lib/postgresql` | `/var/lib/postgresql/18/docker` |
| MVP/disconnected | `postgres:16-alpine` | Exact 18.6 index digest | `cartulary-postgres-data:/var/lib/postgresql/data` | `cartulary-postgres-data-v18:/var/lib/postgresql` | `/var/lib/postgresql/18/docker` |
| Testcontainers/test services | Shared 16 Alpine tag | Exact 18.6 index digest | Image-managed disposable data | Image-managed disposable parent volume | `/var/lib/postgresql/18/docker` |
| Recovery smoke | Inherits MVP/test fixture | Inherits exact 18.6 identity | Inherited | Inherited parent volume | `/var/lib/postgresql/18/docker` |

All repository-provisioned clusters set
`POSTGRES_INITDB_ARGS="--data-checksums --auth-host=scram-sha-256"`. Admission requires
`current_setting('server_version_num')::integer = 180006` and
`current_setting('data_checksums') = 'on'`. A managed service that cannot present both
facts is unsupported; no relaxed mode or warning-only path exists.

The data roots remain Docker-managed named volumes, so no SELinux relabel flag is
introduced. Host bind-mounted PostgreSQL data is outside the supported target shape.
Health checks continue to use `pg_isready`, but readiness is not granted until exact
server/checksum admission, purpose-role admission, migration readiness, and existing
application bootstrap gates all succeed.

### Platform interface

`internal/platform/postgres` will expose:

- `RequiredServerVersionNum = 180006`;
- `ReasonServerVersionMismatch = "postgres_server_version_mismatch"`; and
- one admission operation shared by owned pools, `database/sql` connections, and the
  borrowed server-pool path.

Physical connection initialization selects the required role, reads `session_user`,
`current_user`, `server_version_num`, and `data_checksums`, and accepts only the exact
purpose and engine baseline. `postgres.Setup` establishes one admitted connection
before returning. The server validates a borrowed `server.Options.Postgres` pool before
lease acquisition or other database operations. Failures emit closed reason codes and
never include a DSN, host, password, catalog payload, or raw vendor error.

### Explicitly unchanged behavior

- DDL lineage, migration bytes 1-40, manifest hashes, schema, ledger, and lineage table.
- Administrator-managed extension ownership and exact extension versions.
- Runtime, migration, and Recovery login-to-role mapping, `NOINHERIT` posture, object
  ownership, grants, default privileges, and negative privilege matrix.
- Runtime denial of DDL, `TRUNCATE`, sequence update, ledger mutation, replication, and
  `session_replication_role`; Recovery retains only its existing bounded capability.
- Backup artifact schema, encryption, inventory, audit, sequence, and projection rebuild
  semantics.
- Public HTTP, WebSocket, OpenAPI, and browser behavior.

## 5. PostgreSQL Compatibility Findings

| PostgreSQL change | Source version | Present in Cartulary? | Evidence path | Risk | Required action | Required test | Disposition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `initdb` checksums default on | 18 | Fresh clusters | Compose and pgtest initialization | An implicit default can drift | Set `--data-checksums`; admit only `on` | Query setting in every provisioned profile | `required_change` |
| Matching checksum posture for `pg_upgrade` | 18 | No | Fresh-only decision; no upgrade tooling | None for disposable clusters | Do not add `pg_upgrade` | Absence/residual scan | `no_action_with_evidence` |
| Container `PGDATA` and `VOLUME` move | 18 image | Yes | Both Compose files and package smoke | Wrong mount can appear to lose data | Mount `/var/lib/postgresql`; set versioned `PGDATA` and volume names | Rendered topology, ownership, restart, stale-volume failure | `required_change` |
| Generated columns default to virtual | 18 | No generated columns found | Authored migration scan | None | No DDL change | Static scan retained in handoff evidence | `not_present` |
| CSV `COPY FROM` handling of `\.` | 18 | No authored SQL `COPY` | Migration/query scan; imports are application-side | None | No change | Existing import integration rows | `not_present` |
| AFTER trigger execution-role semantics | 18 | Yes, many regular and constraint triggers | Migrations 12, 20, 28, 31, 32, 36, 37, 39 | Role-sensitive writes may change | Preserve SQL; characterize before editing | Trigger-backed owner flows under runtime, migration, and Recovery roles | `required_test` |
| Time-zone abbreviation precedence | 18 | No custom abbreviation config | Config and SQL scan | Timeline interpretation regression | No setting change | Timeline conversion and ordering tests | `required_test` |
| FTS/`pg_trgm` collation/reindex changes | 18 | No | Extension/index scan | None | No action | Residual scan | `not_present` |
| Explicit `C` collation | 18 | Yes | Migration 39 and Network Flow queries | Ordering drift | Retain explicit collation | Entity, party, and Network Flow ordering/parity | `required_test` |
| MD5 authentication deprecation | 18 | No database-auth MD5 config found | Provisioning/config scan | Image defaults may be implicit | Require SCRAM in init and provisioning | `password_encryption`; boolean SCRAM verifier check | `required_change` |
| AIO and changed I/O defaults | 18 | Engine behavior only | No repository tuning | Performance variance | Keep defaults during migration | Engineering performance comparison | `future_optimization` |
| Planner changes | 17/18 | All SQL can receive new plans | Queries and stores | Latency or plan regression | Do not pin plans or tune in migration | Result parity plus qualified workload comparison | `required_test` |
| Catalog/statistics-view changes | 17/18 | Stable catalogs queried; affected statistics views not found | Migration preflight, ACL tests, telemetry scan | Hidden operational query failure | Preserve queries unless characterization fails | Catalog, readiness, telemetry, role/ACL rows | `required_test` |
| Optimizer statistics preservation in `pg_upgrade` | 18 | No | No `pg_upgrade` | None | No action | Fresh `ANALYZE`/application exercise | `no_action_with_evidence` |
| External module binary compatibility | 17/18 | No third-party C extensions | Only bundled `pgcrypto` and `citext` | Exact SQL extension version availability | Keep exact 1.3/1.6; prove installation | Pristine and wrong-extension fixtures | `required_test` |
| CPU architecture and `char` signedness | 18 | No physical cross-platform transfer | Multi-arch image; fresh clusters | None | Use index digest; no physical transfer | Supported architecture image resolution | `no_action_with_evidence` |
| 18.6 security corrections | 18.6 | Exact server binary applies them | Target digest | Wrong patch would miss fixes | Exact patch admission and digest | Accept 18.6; reject 18.4, 18.7, 16, and 19 via admission fixtures | `required_change` |
| `output_plugin_libraries` restriction | 18.6 | No slots/plugins | Repository scan | None | No action | Residual scan | `not_present` |
| GIN `reltuples` repair guidance | 18.6 | One array GIN index created fresh | Migration 8 | Invalid statistics only if carried from affected server | No reindex path; analyze fresh index | Finite, nonnegative index statistics and query behavior | `required_test` |
| `btree_gist` or `ltree` reindex guidance | 18.6 | Neither extension present | Migration/extension scan | None | No action | Residual scan | `not_present` |
| `pgcrypto` faulty-ciphertext cleanup | 18.6 | No `pgp_*` use found | SQL/application scan | None | Use patched binary; no data cleanup | Extension/function smoke | `no_action_with_evidence` |
| Exact `pgcrypto`/`citext` packaged versions | 18.6 | Required | Migration 1 and provisioners | Defaults are newer than Cartulary's contract | Request explicit versions 1.3/1.6 | Catalog version/schema/owner checks | `required_test` |
| Non-contiguous 18.4→18.6 patch sequence | 18.6 | No version-order code found | Repository version scan | Future range logic could reject 18.6 | Use integer equality, never patch iteration | Exact admission unit matrix | `required_test` |

## 6. Authority and Owner Impact Map

| Surface | Owner | Change type | Dependency direction | Required evidence |
| --- | --- | --- | --- | --- |
| Core 01 PostgreSQL and DDL requirements | Core 01 | Normative amendment | Owner → migrations/platform/tests | Exact 18.6, head 40, unchanged lineage and SQL |
| Core 04 provisioning, security, and Recovery requirements | Core 04 | Normative amendment | Owner → deployment/platform/Recovery | Fresh-only posture, layout, checksum, SCRAM, ACL and Recovery evidence |
| Core 05 benchmark profile | Core 05 | Normative claim-profile addition | Owner → checker/fixtures/evidence | New identity and retained historical profile separation |
| Testing Harness Sections 2.1/9 and AC-095 | Testing Harness | Normative amendment | Owner → task surface/test services/catalog rows | Exact fixture identity, stale-state rejection, cleanup |
| OpenTelemetry NLSpec | OpenTelemetry | Verification only | Existing owner → PostgreSQL instrumentation tests | No secrets/vendor detail; spans and metrics unchanged |
| Extensions, Graph Projection, Network Flow, Reporting, Report Composition, Reference Pack NLSpecs | Respective subsystem owners | Characterization only | Existing owners → service-backed rows | Database behavior unchanged; amend only if a demonstrated incompatibility changes behavior |
| `contracts/**`, generated SQL/UI projections, topology outputs | Their source owners | Generated projection if owner input changes require it | Owners → generators → generated outputs | `make generate` and drift checks; never hand-edit |
| `internal/platform/postgres` | `platform.postgres` | Implementation and tests | Core 01/Core 04 → platform → application facades | Version/checksum/purpose admission, safe failures, borrowed pools |
| `internal/modules/database_migrations` | `module.database_migrations` | Test expectation/characterization | Core 01 → module → evidence | Full pristine chain, lineage, ledger, contamination, rollback-through-zero harness mechanics |
| `internal/modules/recovery` and recovery assembly | `module.recovery` | Characterization and target admission | Core 04 → Recovery → operator/package | Backup, restore, audit, sequence, rebuild, readiness |
| Server/migrate/operator facades | Matching `internal/app/*` owners | Composition changes only | Platform admission → application entrypoints | No database use before admission; borrowed resource lifetime unchanged |
| Development and test services | Harness command surface/pgtest | Configuration and harness changes | Testing Harness → Compose/broker/fixtures | Exact digest, session invalidation, cleanup, no ordinary PG16 fixture |
| MVP package | Core 04 deployment | Deployment and support changes | Core 04 → package scripts/evidence | Fresh start, exact layout, recovery smoke |
| Benchmark tooling | Core 05 | Claim projection and tests | Core 05 → checker → retained evidence | Qualified profile, no historical reinterpretation |

## 7. Migration Strategy Decision Record

### Decision

Only fresh provisioning is supported. Cartulary is pre-production, the user has
confirmed that no database needs preservation, and no adopted owner currently claims a
supported PostgreSQL 16 deployment upgrade. This decision does not weaken a supported
deployment guarantee.

| Candidate | Decision | Reason |
| --- | --- | --- |
| Fresh PostgreSQL 18.6 cluster | Selected | Smallest support surface; matches disposable state and existing append-only migration/bootstrap model. |
| Logical dump/restore | Rejected | Creates role, sequence, extension, downtime, and tool-version obligations for data that has no preservation requirement. |
| `pg_upgrade` | Rejected | Retains old binaries and physical cluster/checksum/platform coupling with no user value. |
| Logical replication cutover | Rejected | Adds slots, publications, dual running clusters, consistency coordination, and operator burden absent from current contracts. |

Old development/test/MVP PostgreSQL volumes are stopped, explicitly selected, and
deleted. New names prevent implicit attachment. The PostgreSQL 18 image's old-directory
detection is defense in depth; Cartulary also tests a synthetic `PG_VERSION=16` volume
and rejects a running PostgreSQL 16 endpoint through exact server admission. No
PostgreSQL 16 image remains in the default service surface after cutover.

Rollback before the new cluster accepts writes may revert source/configuration and
create another disposable PostgreSQL 16 cluster. After writes, there is no in-place
downgrade: discard the PostgreSQL 18 volume. After PostgreSQL 18.6 becomes the supported
baseline, recovery means restoring a Cartulary logical backup into a pristine 18.6
target, never restarting an old PostgreSQL 16 data directory.

Required owner amendments are Core 01, Core 04, Core 05, and Testing Harness. The
user's explicit hard-cutover request authorizes planning those amendments; no unresolved
owner decision remains.

## 8. Migration Workstreams

| Workflow ID | Name | Class | Required predecessors | Required successors | Goal | Likely surfaces | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-01 | Baseline and source freeze | Evidence | None | WF-02, WF-07 | Preserve reproducible before-state and complete inventory | Tracker, git/image/migration evidence | Repository scans and PG16 engineering run | SL-00 log |
| WF-02 | Owner adoption | Normative | WF-01 | WF-03 through WF-08 | Make exact 18.6 and fresh-only behavior authoritative | Core 01/04/05, Testing Harness | Owner review and Markdown lint | SL-01 log |
| WF-03 | Container and provisioning | Implementation | WF-02 | WF-04, WF-05, WF-06 | Establish exact image, layout, checksums, SCRAM, and safe reset | Compose, provisioners, package smoke, task surface | Topology and lifecycle evidence | SL-02 log |
| WF-04 | Platform and SQL compatibility | Implementation/test | WF-02, WF-03 | WF-05, WF-06 | Admit exact engine/purpose and prove unchanged DDL/app SQL | Platform, server/migrate/operator, Database Migrations | Owner unit/service-backed slices | SL-03 log |
| WF-05 | Test-service cutover | Harness | WF-03, WF-04 | WF-06 through WF-09 | Make every normal test fixture exact 18.6 and reject stale sessions | pgtest, test broker, catalog/topology | Harness and service-backed rows | SL-04 log |
| WF-06 | Recovery closure | Verification | WF-03 through WF-05 | WF-08, WF-09 | Prove backup/restore/rebuild on admitted 18.6 targets | Recovery/operator/MVP | Recovery rows and operational smoke | SL-06 log |
| WF-07 | Security, telemetry, performance | Verification/claim | WF-01, WF-04, WF-05 | WF-08, WF-09 | Prove invariants and establish distinct benchmark identity | PostgreSQL ACL tests, telemetry, benchmark tooling | Focused rows, comparison, claim check | SL-07 log |
| WF-08 | Projection and guidance refresh | Generation/docs | WF-02 through WF-07 | WF-09 | Refresh downstream artifacts through owners and publish safe procedures | Generators, contracts, guides, MVP README | Generation drift and Markdown lint | SL-08 log |
| WF-09 | Release closure | Evidence | WF-03 through WF-08 | None | Remove unclassified defaults and pass broad gates | Whole repository | Agent finalize, check, harness, release | SL-09 log |

## 9. Ordered Implementation Slice Plan

### SL-00 — Freeze the Baseline (`DONE`)

- **Objective:** Reconfirm repository/image facts and capture the qualified PostgreSQL
  16 engineering comparison window before removing the default 16 fixture.
- **Prerequisites:** Explicit authorization to implement the migration; clean or fully
  attributed worktree.
- **Surfaces:** This tracker, current test evidence roots, image-registry observation,
  migration manifest.
- **Changes:** Tracker evidence only; no product behavior.
- **Must remain unchanged:** Source, migration bytes, generated artifacts, and retained
  PostgreSQL 16 claim evidence.
- **Characterization:** Branch/commit/status, exact digest, migration head/hash, full
  version scan, warm representative workload evidence.
- **Validation:** `make help`, `make help-all`, owner `make task-guide` calls, and the
  existing qualified warm run used by `make harness-performance-check`.
- **Exit:** Every inventory row has an owner/disposition and the PG16 evidence root is
  retained with hardware, fixture, resource, and sampling identity.
- **Rollback:** Delete only newly produced planning/run artifacts; no runtime state was
  changed.
- **Handoff/commit:** Record run roots and digest observation; one planning checkpoint.

### SL-01 — Adopt the Normative Baseline (`DONE`)

- **Objective:** Make exact 18.6, fresh-only support, checksums, SCRAM, layout, and
  benchmark identity owner-controlled.
- **Prerequisites:** SL-00.
- **Surfaces:** Core 01, Core 04, Core 05, Testing Harness; owner-derived contracts only
  when their generators require them.
- **Changes:** Replace normative 16 requirements, correct head 37 to 40, add the new
  Core 05 profile without rewriting historical evidence, define exact admission and
  reset semantics.
- **Must remain unchanged:** Rebaseline v2, immutable 1-29 policy, roles/ACLs, logical
  Recovery contract, public APIs.
- **Characterization:** Confirm no concurrent owner edits alter these requirements.
- **Validation:** `make lint-markdown`; owner/contract structural checks discovered by
  `make task-guide` and `make explain-target`.
- **Exit:** Owner text is non-contradictory and all downstream work cites an adopted
  requirement.
- **Rollback:** Revert this slice as one owner-only commit before implementation begins.
- **Handoff/commit:** Owner amendment matrix and exact changed requirement IDs.

### SL-02 — Cut Over Containers, Storage, and Provisioning (`DONE`)

- **Objective:** Establish exact image and fresh-cluster topology.
- **Prerequisites:** SL-01.
- **Surfaces:** Development/MVP Compose, both provisioners, MVP README and smoke checker,
  authored task-surface/readiness inputs.
- **Changes:** Apply the exact digest, explicit `PGDATA`, parent mounts, versioned volume
  names, checksum/SCRAM init arguments, and SCRAM provisioning. Add public
  `postgres-baseline-reset` with dry-run and
  `CARTULARY_DESTRUCTIVE_CONFIRM=postgres-baseline-reset`; it stops PostgreSQL, resolves
  only Compose-labeled old/new PostgreSQL volumes, preserves SeaweedFS/object/backup
  volumes, deletes the selected volumes, starts 18.6, provisions, and migrates.
- **Must remain unchanged:** Object storage, backup/reference-pack volumes, credentials,
  health endpoint shape, and ordinary `db-reset` database-only semantics.
- **Characterization:** Render current Compose projects and capture exact volume labels;
  do not use globs or broad `down -v` cleanup.
- **Validation:** Lifecycle command tests, rendered Compose assertions,
  `make standup-package-smoke`.
- **Exit:** Fresh/restart persistence works at the target paths, dry-run is inert,
  missing/wrong confirmation fails before Docker mutation, and only PostgreSQL volumes
  can be deleted.
- **Rollback:** Before writes, restore old Compose/config and create a fresh disposable
  16 volume; never reinterpret the deleted volume as recoverable.
- **Handoff/commit:** Container/provisioning/reset changes and their tracker checkpoint.

### SL-03 — Close Platform, Driver, and SQL Compatibility (`DONE`)

- **Objective:** Fail closed on any engine other than admitted 18.6 and prove the
  unchanged application SQL chain.
- **Prerequisites:** SL-01 and SL-02.
- **Surfaces:** `internal/platform/postgres`, server/migrate/operator facades, Database
  Migrations tests, SQLC/generator verification.
- **Changes:** Add exact constants/reason, connection admission, eager owned-pool
  admission, and borrowed-pool validation. Update only server-version test expectations.
- **Must remain unchanged:** pgx/Goose/SQLC/testcontainers versions, authored SQL 1-40,
  query APIs, connection-purpose mapping, borrowed-resource lifetime.
- **Characterization:** Full migration chain, extension scripts, functions, triggers,
  GIN index, collations, catalog queries, locks, notifications, TLS/query modes,
  isolation, cancellation, prepared execution, and SQLSTATE mapping.
- **Validation:** Focused `platform.postgres` and `module.database_migrations` unit and
  service-backed slices plus `make migration-drift` and `make generate-drift`.
- **Exit:** 18.6/checksums-on succeeds; all other version/checksum cases fail before
  mutable work; migrations reach head 40 with exact lineage/ledger/hash evidence.
- **Rollback:** Revert admission implementation and its tests together before target
  writes; never add a compatibility toggle.
- **Handoff/commit:** Platform and Database Migrations checkpoint with reason-code and
  secret-redaction evidence.

### SL-04 — Cut Over Test Services and Fixtures (`DONE`)

- **Objective:** Make exact 18.6 the only ordinary PostgreSQL fixture.
- **Prerequisites:** SL-02 and SL-03.
- **Surfaces:** pgtest, testcontainersx fixtures, test-service broker/session metadata,
  harness owner/topology inputs, service-backed consumers.
- **Changes:** Pin shared fixtures, carry checksum/SCRAM init options, invalidate reusable
  sessions on image digest mismatch, and add synthetic stale-volume rejection.
- **Must remain unchanged:** Borrowed/owned cleanup rules, database isolation, lease
  summaries, extension/role provisioning, and fixture-only generic image tests.
- **Characterization:** Inventory every consumer of `pgtest.ContainerImage()` and every
  service-session identity input.
- **Validation:** `make test-slice OWNER=harness.test_catalog`, applicable harness
  contract targets, and all PostgreSQL-backed owner slices.
- **Exit:** No default fixture resolves PostgreSQL 16; a synthetic old `PG_VERSION`
  fails without running a 16 image; no database/container/volume/session leak remains.
- **Rollback:** Revert the fixture slice and destroy only its owned sessions/resources.
- **Handoff/commit:** Harness input, generated topology, tests, and tracker checkpoint.

### SL-05 — Data-Preserving Upgrade Path (`NO_ACTION`)

- **Objective:** Keep the closed fresh-only decision visible.
- **Prerequisites:** SL-01.
- **Surfaces:** Tracker and support guidance only.
- **Changes:** None; record rejection of dump/restore, `pg_upgrade`, and logical
  replication.
- **Must remain unchanged:** No PostgreSQL 16 runtime image, compatibility path, or
  operator action is added.
- **Characterization/validation:** Residual scan for physical upgrade and replication
  tools.
- **Exit:** All such occurrences are absent or historical source citations.
- **Rollback:** Not applicable.
- **Handoff/commit:** Included with the nearest documentation checkpoint; no standalone
  implementation commit.

### SL-06 — Close Backup, Restore, and Recovery (`DONE`)

- **Objective:** Prove current logical Recovery behavior on exact 18.6.
- **Prerequisites:** SL-03 and SL-04.
- **Surfaces:** Recovery application/store/process rows, operator Recovery composition,
  MVP recovery scripts and smoke evidence.
- **Changes:** Only target-admission plumbing or test expectations required by the new
  platform contract; no artifact schema or engine field.
- **Must remain unchanged:** Encryption, artifact inventory, catalog selection, audit,
  sequence restoration, projection rebuild, serving leases, object-store consistency.
- **Characterization:** Capture, latest inspection, selected restore, failures,
  verification, rebuild, and incident portability where currently claimed.
- **Validation:** Focused `module.recovery` slices and
  `make standup-operational-recovery-smoke`.
- **Exit:** An admitted pristine 18.6 target restores and becomes ready; incompatible
  targets fail before restore mutation; all retained artifact semantics match v2.
- **Rollback:** Revert Recovery plumbing; discard owned scratch targets and leave
  borrowed targets unchanged.
- **Handoff/commit:** Recovery evidence/run roots and tracker checkpoint.

### SL-07 — Validate Security, Telemetry, and Performance (`DONE`)

- **Objective:** Prove no privilege or observability widening and establish the new
  benchmark identity.
- **Prerequisites:** SL-00, SL-03, SL-04, and SL-06.
- **Surfaces:** PostgreSQL role tests, telemetry tests/NLSpec evidence, Core 05 checker
  and fixtures, qualified performance manifests.
- **Changes:** Add version/checksum/SCRAM and GIN-health assertions; add the 18.6 claim
  profile projection and tests.
- **Must remain unchanged:** Allow/deny/default matrices, redaction, semantic workload,
  hardware/resource claims, sampling rules, and old retained profile evidence.
- **Characterization:** New/recycled connections; future objects; trigger paths;
  workbook pagination, projections, mutations, revisions, imports, backup/restore,
  Recovery rebuild, Network Flow, Reporting, and harness database lifecycle.
- **Validation:** Focused platform/telemetry/affected module rows,
  `make otel-conformance`, `make harness-performance-check
  EVIDENCE_ROOTS_FILE=<qualified-pair-manifest>`, and `make benchmark-claim-check`.
- **Exit:** Functional parity and security invariants pass; performance result is
  recorded without automatic tuning; old/new profiles cannot be confused.
- **Rollback:** Revert only new claim projection if its evidence is unqualified; keep
  engineering results clearly non-claim-bearing.
- **Handoff/commit:** Security/telemetry commit, then benchmark-profile commit if claim
  evidence qualifies.

### SL-08 — Regenerate Projections and Update Guidance (`DONE`)

- **Objective:** Refresh all downstream outputs and publish exact hard-cutover guidance.
- **Prerequisites:** SL-01 through SL-07, excluding SL-05 which is already closed.
- **Surfaces:** Owner-driven generators/contracts/topology and current development,
  deployment, Recovery, and maintenance guides.
- **Changes:** Run repository generators; document exact image, mount, reset, stale-state
  rejection, and rollback cutoff. Preserve historical archives as classified.
- **Must remain unchanged:** Generated files are never hand-edited; no new product
  behavior is introduced in documentation.
- **Characterization:** Generator inputs/outputs and every current version/path mention.
- **Validation:** `make generate`, `make generate-drift`,
  `make generated-artifact-policy-check`, `make json-shape-check`, and
  `make lint-markdown`.
- **Exit:** Generated drift is zero and every remaining 16/old-path occurrence is
  explicitly historical or fixture-only.
- **Rollback:** Revert owner inputs and rerun generation; do not manually reverse output.
- **Handoff/commit:** Owner inputs, generated projections, guides, and tracker checkpoint.

### SL-09 — Final Verification and Handoff (`DONE`)

- **Objective:** Produce retained narrow-to-broad evidence and close the migration.
- **Prerequisites:** All change slices done and SL-05 closed `NO_ACTION`.
- **Surfaces:** Whole repository and retained test-result roots.
- **Changes:** Validation/accounting repairs only; no unplanned product behavior.
- **Must remain unchanged:** All boundaries and exclusions in this tracker.
- **Characterization:** Final version/path/tool scan, diff audit, cleanup inventory, and
  risk closure.
- **Validation:** The ordered matrix in Section 11.
- **Exit:** All binary implementation criteria in Section 16 pass, no unexpected 16
  baseline remains, and cleanup proves no leaked resources.
- **Rollback:** Any failure returns control to its owning slice; do not mark final
  completion or publish the new benchmark profile.
- **Handoff/commit:** Final tracker/evidence checkpoint after successful retained run.

## 10. Task and Status Table

| Task ID | Status | Depends on | Task | Required evidence | Binary completion |
| --- | --- | --- | --- | --- | --- |
| PG18-T001 | DONE | None | Complete planning tracker | This file, planning commands, final diff/status | All required sections populated; only tracker changed |
| PG18-T002 | DONE | PG18-T001 | Reconfirm baseline and capture PG16 comparison | Git/image/source inventory; retained passing PG16 measurement root `.cartulary/test-results/20260903T103552Z-p24774` | SL-00 exit passes |
| PG18-T003 | DONE | PG18-T002 | Amend Core 01/04 and Testing Harness | Requirement diff and owner checks | Exact 18.6/fresh-only contract adopted |
| PG18-T004 | DONE | PG18-T002 | Add Core 05 18.6 profile identity | Owner/profile diff | Historical v1 preserved; current `cartulary.perf.desktop_ref.v2` exact |
| PG18-T005 | DONE | PG18-T003 | Cut over Compose, provisioning, mount, and init | Topology and provisioning tests | Both stacks use exact digest/new paths/checksums/SCRAM |
| PG18-T006 | DONE | PG18-T005 | Add safe PostgreSQL baseline reset | Dry-run/confirmation/scope tests | Only named PostgreSQL volumes can be removed |
| PG18-T007 | DONE | PG18-T003, PG18-T005 | Add platform version/checksum admission | Unit and integration reason-code evidence | Owned/borrowed and recycled connections enforce exact baseline |
| PG18-T008 | DONE | PG18-T007 | Prove unchanged DDL/application compatibility | Migration/SQLC/service-backed results | Head 40 and all characterized SQL paths pass without SQL edits |
| PG18-T009 | DONE | PG18-T005, PG18-T007 | Cut over pgtest and test services | Harness rows and session evidence | Exact 18.6 is the only ordinary fixture; stale state rejected |
| PG18-T010 | NO_ACTION | PG18-T003 | Implement data-preserving upgrade | Fresh-only decision and residual scan | No upgrade implementation exists |
| PG18-T011 | DONE | PG18-T008, PG18-T009 | Validate Recovery | Recovery rows and operational smoke | Capture/restore/rebuild/readiness pass on 18.6 |
| PG18-T012 | DONE | PG18-T008, PG18-T009 | Validate security and telemetry | ACL, recycled connection, OTel evidence | No privilege, secret, telemetry, or source-owner boundary widening |
| PG18-T013 | DONE | PG18-T002, PG18-T011 | Run engineering comparison and profile validation | Qualified paired roots, claim check, and complete harness owner registration | Result recorded; no silent tuning/profile reinterpretation |
| PG18-T014 | DONE | PG18-T003 through PG18-T013 | Generate projections and update guidance | Final post-repair generation and drift/policy/shape roots | Zero drift and complete operator instructions |
| PG18-T015 | DONE | PG18-T014 | Run final gates and residual audit | Successful retained run root | Section 16 implementation criteria all true |

Final validation found three implemented performance-fixture lifecycle tests missing
from their authored harness owner row. The owning SL-07 repair registered them without
runtime changes; post-repair SL-08 reconciliation is complete and SL-09 is ready.
The first full `make check` at `.cartulary/test-results/20260904T193058Z-p61246`
then exposed test-only DDL using the admitted runtime pool, a Network Flow Recovery
fixture using the runtime purpose, and Party/Entity Incident Bundle validation calling
private helper functions unavailable to the intentionally narrow runtime role. These
failures returned control to SL-07. Test DDL now uses the fixture-admin handle, Recovery
uses an admitted Recovery-purpose handle, and owner Go normalization compares portable
claim sets without private-routine grants. The Entity validator also reuses the existing
owner-approved current-envelope loader instead of duplicating a Records query or widening
the boundary manifest. ACLs, migrations, and authored queries remain unchanged.
Sequential affected rows and final-source performance evidence pass; parallel
roots `.cartulary/test-results/20260904T195433Z-p7018` and
`.cartulary/test-results/20260904T195433Z-p7032` retain the shared warm-stamp contention
characterization and are superseded by sequential passing roots.

## 11. Validation Matrix

### Planning-time commands and results

| Command or inspection | Result |
| --- | --- |
| `git branch --show-current`, `git rev-parse HEAD`, `git status --short`, local `date` | Baseline recorded; tree was clean before tracker creation. |
| `make help` and filtered `make help-all` | Current public database, test, generation, smoke, performance, and release targets discovered. |
| `make task-guide ROLE=module-author OWNER=platform.postgres` | Focused unit and service-backed owner slices; broader `make test-fast`. |
| Equivalent task guides for `module.database_migrations` and `module.recovery` | Focused slices confirmed; Recovery also routes to browser-backed coverage. |
| `make explain-target TARGET=db-reset DETAIL=summary` | Existing target resets a database inside the cluster; it is not a volume-major reset. |
| `make explain-target TARGET=services-down DETAIL=summary` | Existing target preserves named volumes. |
| Content searches for versions, images, paths, extensions, tools, catalogs, SQL features, and dependencies | Populated Sections 3 and 5; no physical upgrade/replication implementation found. |
| Migration manifest and source inspection | Immutable boundary 29 and live head 40 confirmed. |
| Official registry manifest inspection | Exact 18.6 multi-architecture index digest recorded. |
| Primary PostgreSQL and official image source review | Compatibility and storage dispositions recorded. |

### Implementation evidence routing

| Concern | Owner / representative rows | Canonical command | Expected evidence / failure classification | Cleanup |
| --- | --- | --- | --- | --- |
| Version, purpose, roles, ACLs | `platform.postgres.integration.purpose_identity_roles_and_acl`; platform unit/support rows | `make test-slice OWNER=platform.postgres`; `make service-backed-test-slice OWNER=platform.postgres` | Exact admission and allow/deny/default matrices; any widening is change-related blocking failure | Owned scratch DB/container removed |
| DDL chain and readiness | `module.database_migrations.integration.{schema_bootstrap,schema_readiness_matrix,production_preflight_state_matrix,production_ddl_v2_recurrence,migration_evidence_database_semantics,concurrent_apply_locking}` | Focused Database Migrations slices; `make migration-drift` | Head 40, exact lineage/ledger/hash, contamination and malformed/ahead rejection | Owned scratch DB removed; borrowed DB unchanged |
| Test-service identity | `harness.test_catalog` plus pgtest/test-service harness rows | Focused harness slice and `make harness-contract` | Digest/session mismatch and stale PG16 state fail closed | Sessions, containers, databases, and volumes accounted for |
| Server admission/readiness | `app.server.integration.config_failures_on_real_postgres_and_object_stor_c832f91897`, bootstrap/readiness process rows | Focused `app.server` slices | No leases/listeners/jobs before engine/schema admission | Borrowed pool remains borrowed; owned pool closed |
| Backup and Recovery | Recovery database coverage, restore readiness, target serving admission, restore failure, process, and store rows | Focused `module.recovery` slices; `make standup-operational-recovery-smoke` | Exact 18.6 target, complete artifacts, sequences, audit, rebuild, and serving readiness | Scratch target and package resources removed |
| Telemetry and privacy | `platform.postgres.unit.postgres_telemetry_behavior`; `platform.telemetry.unit.privacy_and_redaction` and conformance rows | Focused telemetry slices; `make otel-conformance` | Stable semantic telemetry without secrets, hostnames, or vendor errors | Export fixtures removed |
| Package topology | Core 04 standup evidence | `make standup-package-smoke` | Exact digest, parent mount, effective `PGDATA`, checksums, health/readiness | Package containers and owned PG volume removed |
| Browser/database behavior | Affected app/module browser rows | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful` | Observable behavior unchanged | Harness-owned stack removed |
| Performance | Core 05 frontend measurement rows and benchmark profile | Focused `module.timeline` measurement rows | Qualified 100-sample predicates; variance classified, not silently tuned | Retain qualified v1/v2 evidence roots only |
| Claim identity | Core 05 checker and fixture | `make benchmark-claim-check` | New 18.6 identity accepted; old identity remains historical | No destructive cleanup |
| Generated/projection drift | Generated artifact policy | `make generate`; drift/policy/JSON checks | Generated output matches owner inputs | Reproducible generated outputs only |
| Broad closure | All routed owners | `make test-fast`, then agent-finalize/check/harness/release sequence | Retained successful run root; unrelated failures identified separately | Final leak audit |

Final validation order is:

1. `make migration-drift`
2. `make generate`
3. `make generate-drift`
4. `make generated-artifact-policy-check`
5. `make json-shape-check`
6. Focused owner unit and service-backed slices
7. `make otel-conformance`
8. `make standup-package-smoke`
9. `make standup-operational-recovery-smoke`
10. `make browser-e2e-webserver-backed`
11. `make browser-e2e-stateful`
12. `make test-fast`
13. Run the four qualified Core 05 frontend measurement rows.
14. `make benchmark-claim-check`
15. `make agent-finalize` without `RESULTS_DIR`.
16. `make check`
17. `make agent-finalize RESULTS_DIR=<successful-warm-check-root>`
18. `make harness-contract`
19. `make release-check`

The angle-bracket values above are evidence references produced by the preceding
qualified runs, not unresolved product decisions. Every command runs from repository
root through the public Make surface.

`harness-performance-check` is not a validator for the Core 05 frontend-measurement
aggregate. TH-HARNESS-REQ-079 requires that command to compare separate five-run windows
of clean-source public-command duration evidence. No such manifest was supplied by this
migration, and passing the dirty-source browser aggregate to it would violate the
adopted harness owner. The migration instead ran the exact Core 05 measurement rows and
the benchmark claim checker; the omission is intentional and does not weaken the
frontend threshold evidence.

## 12. Rollout, Backup, Restore, and Rollback Plan

### Preparation

1. Finish SL-00 and retain the PostgreSQL 16 engineering comparison; do not retain a
   PostgreSQL 16 runtime fixture.
2. Land owner amendments before implementation.
3. Stop the exact reusable test-service session and development services.
4. Render Compose metadata and identify old/new PostgreSQL volumes by project and
   Compose labels. Do not use globs, unresolved environment variables, or `down -v`.
5. Run `CARTULARY_CLEANUP_DRY_RUN=1 make postgres-baseline-reset`, inspect the exact
   PostgreSQL-only targets, then execute with
   `CARTULARY_DESTRUCTIVE_CONFIRM=postgres-baseline-reset`.

No pre-upgrade database backup is required because no PostgreSQL 16 data is supported
for preservation. Object-store, backup, reference-pack, and other non-PostgreSQL volumes
must be preserved by the reset.

### Cutover and readiness

1. Delete the old disposable PostgreSQL volume; it is not held as rollback state.
2. Create the versioned PostgreSQL 18 volume mounted at `/var/lib/postgresql`.
3. Start the exact image and verify effective `PGDATA`, ownership, checksums, SCRAM, and
   exact server admission.
4. Provision the administrator schema, extension versions, fixed roles, purpose logins,
   grants, and default privileges.
5. Apply the unchanged migration source from pristine state to head 40.
6. Verify lineage, Goose ledger, object ownership, security matrices, application
   bootstrap, backup/restore, projection rebuild, and operational readiness.
7. Repeat the PostgreSQL-only disposal/fresh-start procedure for the pre-production MVP
   package. Never delete the package's object, backup, reference-pack, temporary, or
   export volumes as a side effect.

### Backup and Recovery after cutover

- Create and inspect a post-cutover logical backup on PostgreSQL 18.6.
- Restore it into a separately provisioned pristine 18.6 target and verify row units,
  sequences, administrative audit, migration metadata, object-store references, and
  projection rebuild before serving admission.
- Keep `cartulary.postgres_snapshot_artifact.v2` unchanged. The target's retained
  admission evidence records exact engine identity; the portable artifact does not
  encode a source-engine compatibility promise.
- Reject a non-18.6 restore target before the first restore mutation.

### Rollback cutoff

- Before 18.6 accepts writes, source/configuration may be reverted and a new disposable
  PostgreSQL 16 cluster created.
- After the first 18.6 write, no in-place downgrade or PostgreSQL `Down` migration is a
  database-engine rollback. Pre-production rollback discards the PostgreSQL 18 volume
  and recreates disposable state.
- After 18.6 is adopted as the supported baseline, recover only into a pristine 18.6
  target from a validated Cartulary logical backup. Never restart or attach the deleted
  PostgreSQL 16 data directory.

## 13. Risk Register

| Risk | Probability | Impact | Trigger | Prevention | Detection | Owner | Rollback | Slice |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Wrong volume mount appears to lose data | Medium | High | Child path retained under PG18 image | Exact parent mount, explicit `PGDATA`, versioned volume | Compose/package mount and restart tests | Core 04/Harness | Revert before writes; recreate disposable cluster | SL-02 |
| Checksum posture differs | Low | High | Implicit/default or managed-service setting | Explicit init args and admission | `data_checksums` assertion | Core 01/platform.postgres | Reject target; recreate correctly | SL-02/03 |
| Extension version/schema mismatch | Medium | High | PG18 default extension version installed | Explicit 1.3/1.6 provisioning | Catalog positive/negative fixtures | Core 01/Database Migrations | Recreate pristine target | SL-02/03 |
| Collation or text ordering changes | Low | Medium | Result/order mismatch | Retain explicit `C` where authored | Entity/party/Network Flow parity | Source owners | Stop; owner-approved SQL follow-up only | SL-03/07 |
| Trigger executes with changed effective role | Medium | High | Trigger-backed mutation fails or gains access | Purpose admission and unchanged privilege baseline | Trigger flows under all applicable roles | Source owner/Core 04 | Revert; no privilege broadening workaround | SL-03/07 |
| Hidden PostgreSQL 16 fixture survives | Medium | Medium | Residual tag, service, or session | Central digest and final scans | Session mismatch and content audit | Testing Harness | Destroy owned stale session; repair source | SL-04/09 |
| Borrowed pool bypasses engine admission | Medium | High | Server receives `Options.Postgres` | Explicit borrowed-pool validation before leases | Server integration negative fixture | app.server/platform.postgres | Reject startup; borrowed resource untouched | SL-03 |
| Privilege widening | Low | Critical | Provisioning or PG behavior differs | Preserve exact grants/defaults; no compensating broad grants | Fresh/recycled allow/deny/ownership matrices | Core 04/platform.postgres | Revert provisioning; recreate cluster | SL-02/07 |
| Secret/vendor detail leaks in errors or telemetry | Low | High | New admission query fails | Closed reason codes and sanitized attributes | Privacy/redaction tests | platform.postgres/telemetry | Revert admission diagnostics | SL-03/07 |
| Backup/restore incompatibility | Low | Critical | Restore or projection rebuild fails on 18.6 | Early full Recovery slice | Operational Recovery smoke | Core 04/module.recovery | Block cutover; discard scratch target | SL-06 |
| Stale benchmark identity | Medium | High | Existing 2026q1 ID reused | New exact profile; preserve historical evidence | Claim checker | Core 05 | Withhold claim profile/publication | SL-01/07 |
| Generated artifact drift | Medium | Medium | Owners change without generation | Generator-only update policy | Generation and policy checks | Source owners/Harness | Revert input or rerun generator | SL-08 |
| Unsupported old-cluster rollback attempted | Medium | High | Operator treats old volume as rollback | Delete old volume; explicit cutoff guidance | Runbook review and stale-volume rejection | Core 04 | Recreate disposable state or restore to 18.6 | SL-02/08 |
| DDL lineage changes accidentally | Low | Critical | Migration/rebaseline edit appears | Hash guard and explicit no-DDL rule | Migration drift and git diff | Core 01/Database Migrations | Revert migration change | SL-03/09 |
| Performance regression | Medium | Medium | Qualified workload exceeds baseline | Same fixtures/resources/sampling; no tuning mixed in | Harness performance comparison | Core 05/source owner | Block claim/cutover; separate diagnosis | SL-07 |
| Cleanup leaks databases, volumes, roles, or containers | Medium | Medium | Failed fixture/reset path | Owned-resource accounting and exact targets | Harness summaries and Docker resource audit | Testing Harness | Run bounded owner cleanup | SL-04/09 |

### Final risk dispositions

| Risk group | Final disposition |
| --- | --- |
| Image, mount, checksum, SCRAM, and extension identity | Closed by exact registry-digest recheck, both provisioning profiles, platform admission, package smoke, and Database Migrations evidence. |
| SQL, collation, trigger-role, catalog, and GIN compatibility | Closed without migration or authored-query edits; source-owner and full-check rows pass on exact 18.6. |
| Borrowed/recycled connection admission, privilege widening, and diagnostic leakage | Closed by opaque admitted handles, eager/per-connection admission, allow/deny matrices, OTel conformance, and retained secret scans. |
| Recovery compatibility and target partial mutation | Closed by admitted target-before-mutation behavior, the full Recovery owner suites, and operational Recovery smoke; artifact v2 remains unchanged. |
| Stale fixture, generated drift, and unsupported rollback | Closed by digest-bound session identity, synthetic legacy-volume rejection, zero generated drift, fresh-only guidance, and classified residual references. |
| Benchmark identity and performance variance | Closed: v1 remains historical, v2 is current, all four 100-sample predicates qualify, and final PostgreSQL 18.6 p95 values remain below their thresholds. The observed v1-to-v2 deltas are +8.7 ms blank-row creation, +1.3 ms focus, effectively 0 ms selection, and -0.7 ms typing; no tuning is warranted in this migration. |
| Resource cleanup | Closed after the repository-owned session-down target removed an expired zero-borrower session and exact Docker cleanup removed one stopped historical Recovery-smoke project plus its six volumes/network and two detached PostgreSQL 16 comparison volumes. Active developer and unrelated Compose resources were preserved. |

## 14. Open Questions and Blockers

There are no open product decisions or owner contradictions. The following formerly
ambiguous topics are closed:

| Topic | Resolution | Closure evidence |
| --- | --- | --- |
| Preserve PostgreSQL 16 data? | No; all scoped data is disposable. | Direct user instruction and pre-production posture |
| Major upgrade mechanism? | None; fresh provision only. | Section 7 decision record |
| Change DDL lineage or add migration? | No. | No concrete PostgreSQL 18 schema incompatibility found; immutable manifest |
| Target patch? | Exact 18.6 only. | Digest and runtime integer admission |
| Checksum policy? | On in every supported target. | Explicit init and admission contract |
| Target data path? | Parent mount `/var/lib/postgresql`, `PGDATA=/var/lib/postgresql/18/docker`. | Official image behavior and Section 4 table |
| Extension versions? | Retain `pgcrypto` 1.3 and `citext` 1.6. | Existing owner/migration contract; fresh install test required |
| Dependency refresh? | No unless a characterized incompatibility blocks a slice. | Current supported protocol use; separate owner approval required if disproved |
| Recovery artifact engine field? | No; retain logical artifact v2 and record target admission in run evidence. | Existing portability boundary |
| Benchmark identity? | New `cartulary.bench.ubuntu_24_04_postgres_18_6.2026q3`; old evidence remains historical. | Core 05 amendment in SL-01 |

Registry digest mismatch, inability to install the exact extension versions, or a
concrete authored-SQL incompatibility discovered during implementation is a slice-local
blocking result, not an invitation for the implementer to choose a fallback. The
affected slice must stop, name the owning Core or subsystem owner, retain the failing
evidence, and obtain a revised owner decision. Unaffected earlier slices may remain done.

## 15. Session Handoff Log

| Field | Value |
| --- | --- |
| Timestamp | 2026-09-04 17:41 EDT |
| Agent/session | Codex `/root`; planning-document implementation |
| Branch/commit | `main` / `6074994b809a4741ffa6655b9267eb42d3566661` |
| Starting state | Clean worktree; tracker absent |
| Current state | Migration implementation, reconciliation, retained validation, audit, and cleanup are complete. |
| Files inspected | Repository instructions; planning framework and tracker convention; Core 00-05; Testing Harness and OpenTelemetry NLSpecs; domain/guides and database-relevant adopted NLSpecs; Compose/provisioning/package assets; PostgreSQL platform, Database Migrations, Recovery, pgtest/test services; migration and test manifests; dependency and SQLC inputs |
| Files changed through SL-09 | 63 paths: Core 01/04/05 and Testing Harness owners; development/MVP Compose, provisioners, and guidance; Recovery guidance and smoke scripts; authored/generated task surface and topology; bounded baseline-reset script; PostgreSQL admission and callers; pgtest/test-service fixtures; Database Migrations PostgreSQL 18 characterization; Recovery composition/integration fixtures; benchmark-profile registry/schema/checker/tests; performance-fixture connection accounting; owner-safe Party/Entity Incident Bundle validation; historical-example classifications; tracker. |
| Commands run | Baseline inventory and owner task guides; exact registry inspections; owner-driven generation/drift; focused platform, Database Migrations, Recovery, harness, Records, Entity, Incident Bundle, telemetry, package, browser, performance, and claim checks; `test-fast`; both finalizers; retained full `check`; harness/release gates; residual, protected-file, secret, Docker, session, and diff audits. |
| Planning result | Fresh-only hard cutover selected; exact digest/path/checksum/SCRAM/admission and unchanged DDL lineage closed |
| Blockers | None. Characterized failures were repaired in their owning slices without dependency refresh, ACL widening, migration/query edits, or compatibility fallback. One final-validation invocation used the nonexistent owner `harness.catalog`; it failed before work began and was immediately corrected to `harness.test_catalog`, which passed. |
| Validation | Retained `make check` `.cartulary/test-results/20260904T212033Z-p66841` passed 715/715; evidence-aware finalizer `.cartulary/test-results/20260904T212449Z-p89524`; harness contract `.cartulary/test-results/20260904T212511Z-p92667`; release check `.cartulary/test-results/20260904T212526Z-p93327` passed 881/881; retained secret scans and final audits pass. |
| Skipped checks | `harness-performance-check` was intentionally not run against frontend measurement evidence because TH-HARNESS-REQ-079 requires five clean-source public-command duration roots per comparison window; the applicable Core 05 100-sample measurement and claim gates passed. |
| Cleanup | `make test-services-session-down` root `.cartulary/test-results/20260904T213924Z-p29407`; no managed test-service session or test-smoke Docker resource remains. Exact stopped test artifacts removed are not recoverable; active developer/unrelated resources were preserved. |
| Next action | None required for this migration; review and commit the completed change set. |

### Migration execution checkpoints

| Slice | Status | Evidence | Handoff |
| --- | --- | --- | --- |
| SL-00 | `DONE` | Branch `main`, commit `6074994b809a4741ffa6655b9267eb42d3566661`; exact planned registry digest reconfirmed; immutable boundary 29 and head 40 reconfirmed; retained passing PostgreSQL 16 measurement root `.cartulary/test-results/20260903T103552Z-p24774` | Source and migration bytes unchanged; owner task routes reconfirmed; SL-01 is ready. |
| SL-01 | `DONE` | Core 01 REQ-01-661, Core 04 REQ-04-153/AC-542, Core 05 REQ-05-007/008/014/PC-002, and Testing Harness TH-HARNESS-REQ-811/AC-095 amended | Exact 18.6, head 40, checksums, SCRAM, fresh-only storage, and immutable benchmark-profile versioning are owner-controlled; `docs/domain.md` unchanged; SL-02 is ready. |
| SL-02 | `DONE` | Exact digest reconfirmed; reset dry-run `.cartulary/test-results/20260904T172328Z-p84409`; package smoke `.cartulary/test-results/20260904T173132Z-p57755`; generated task-surface root `.cartulary/test-results/20260904T172314Z-p81308` | Both Compose profiles use the versioned parent mount, explicit checksums/SCRAM, and exact 18.6 image. Reset accepts only `dev`/`mvp`, requires the exact token, and resolves Compose-labeled PostgreSQL volumes. PostgreSQL 18 array-type ACL incompatibility was repaired without widening privileges; earlier failing roots `20260904T172636Z-p71822`, `20260904T172910Z-p34055`, and `20260904T173018Z-p95859` retain the characterization trail. SL-03 is ready. |
| SL-03 | `DONE` | Admission matrix/redaction `.cartulary/test-results/20260904T174349Z-p17968`; DDL manifest unit `.cartulary/test-results/20260904T174408Z-p18661`; migration drift `.cartulary/test-results/20260904T174502Z-p29078`; generation drift `.cartulary/test-results/20260904T174544Z-p38356`; server startup-order unit `.cartulary/test-results/20260904T174718Z-p46716` | Platform setup and `OpenSQL` are eager; every physical connection validates purpose identity, exact `180006`, and checksums on through closed/redacted reasons. Server injection now accepts only the opaque admitted handle and preserves borrowed lifetime. Migrate/operator/Recovery/test composition follows the admitted boundary. Authored migrations and queries are unchanged. The expected PG16 shared-session rejection is retained at `.cartulary/test-results/20260904T173908Z-p43035`; SL-04 replaces that fixture before the service-backed rerun. |
| SL-04 | `DONE` | Platform PostgreSQL integration `.cartulary/test-results/20260904T175102Z-p74915`; full Database Migrations slice `.cartulary/test-results/20260904T175338Z-p14669`; harness catalog `.cartulary/test-results/20260904T175608Z-p33526`; package/synthetic-legacy-volume smoke `.cartulary/test-results/20260904T175819Z-p58321`; harness contract `.cartulary/test-results/20260904T175914Z-p21202` | pgtest and reusable services resolve the exact 18.6 digest with explicit `PGDATA`, checksums, and SCRAM. Attached fixtures are admitted before use, image reference/resolved ID remain session-compatibility inputs, and a synthetic `PG_VERSION=16` volume makes the 18.6 container exit without any PostgreSQL 16 runtime. Fixture-only fake image strings remain classified. SL-05 is ready. |
| SL-05 | `NO_ACTION` | Non-document source scan found no `pg_upgrade`, dump/restore transition, physical transfer, WAL migration, PostgreSQL publication/subscription/slot, dual-reader/writer, or PostgreSQL 16 runtime implementation. The sole textual match was an unrelated domain publication-intent assertion. | Fresh-only remains closed. No compatibility implementation, switch, image, or implied support promise was added; SL-06 is ready. |
| SL-06 | `DONE` | Full Recovery slice `.cartulary/test-results/20260904T180804Z-p15389`; operational Recovery smoke `.cartulary/test-results/20260904T182706Z-p99588`; migrate diagnostic unit `.cartulary/test-results/20260904T181717Z-p24484` | Recovery source and target connections cross the admitted platform boundary; target admission precedes restore mutation; backup capture/latest inspection/restore verification/readiness and cleanup pass on exact 18.6 while artifact v2 remains unchanged. The smoke now securely stages readable container inputs, carries source Recovery volumes into target verification, and emits clean JSON. Failed roots `.cartulary/test-results/20260904T180907Z-p73080`, `.cartulary/test-results/20260904T181053Z-p35580`, `.cartulary/test-results/20260904T181319Z-p97522`, `.cartulary/test-results/20260904T181426Z-p61330`, `.cartulary/test-results/20260904T181732Z-p24922`, `.cartulary/test-results/20260904T181847Z-p87094`, `.cartulary/test-results/20260904T182025Z-p49583`, `.cartulary/test-results/20260904T182155Z-p12238`, `.cartulary/test-results/20260904T182328Z-p74684`, and `.cartulary/test-results/20260904T182529Z-p37265` retain the ordered characterization trail. Backup/restore incompatibility risk is closed for the supported target; no engine portability claim was added. SL-07 is ready. |
| SL-07 | `DONE` | Second full-check characterization `.cartulary/test-results/20260904T204108Z-p97890`; repaired Records boundary `.cartulary/test-results/20260904T204823Z-p34380`; Entity `.cartulary/test-results/20260904T204836Z-p34741`; Incident Bundle `.cartulary/test-results/20260904T204920Z-p52167`; qualified final-source performance `.cartulary/test-results/20260904T205019Z-p69562`; claim check `.cartulary/test-results/20260904T205434Z-p22276` | Entity claim validation now reuses `loadEntityPortableEnvelopes`; no new Records SQL or boundary exception was added. Exact 18.6 security/telemetry behavior and current v2 claim identity pass; PostgreSQL 18.6 measurement remains below every Core 04 threshold and retained v1 evidence is unchanged. SL-08 is ready. |
| SL-08 | `DONE` | Generation `.cartulary/test-results/20260904T205556Z-p23994`; drift `.cartulary/test-results/20260904T205606Z-p27056`; generated policy `.cartulary/test-results/20260904T205614Z-p30184`; JSON shape `.cartulary/test-results/20260904T205616Z-p30613`; clean protected-input and diff audits | Remaining PostgreSQL 16/image/path references are limited to the six non-pulling `testcontainersx` unit strings, the synthetic `PG_VERSION=16` rejection fixture, immutable benchmark v1, and explicitly historical owner/archive/tracker text. Current guidance and generated surfaces are reconciled; SL-09 is ready. |
| SL-09 | `DONE` | Migration drift `.cartulary/test-results/20260904T205731Z-p32546`; generation/drift/policy/shape `.cartulary/test-results/20260904T205738Z-p35447`, `.cartulary/test-results/20260904T205748Z-p38488`, `.cartulary/test-results/20260904T205756Z-p41623`, `.cartulary/test-results/20260904T205757Z-p42046`; focused owner and conformance roots recorded above; package/recovery smokes `.cartulary/test-results/20260904T210541Z-p68976` and `.cartulary/test-results/20260904T210638Z-p31549`; browser roots `.cartulary/test-results/20260904T210739Z-p93964` and `.cartulary/test-results/20260904T211253Z-p53249`; `test-fast` `.cartulary/test-results/20260904T211508Z-p5049`; final performance/claim `.cartulary/test-results/20260904T211556Z-p11055` and `.cartulary/test-results/20260904T212009Z-p63368`; first finalizer `.cartulary/test-results/20260904T212014Z-p63635`; retained check/finalizer/harness/release roots `.cartulary/test-results/20260904T212033Z-p66841`, `.cartulary/test-results/20260904T212449Z-p89524`, `.cartulary/test-results/20260904T212511Z-p92667`, and `.cartulary/test-results/20260904T212526Z-p93327` | Exact digest still resolves; protected migrations/queries/domain/dependency locks are unchanged; residual references are classified; retained secret scans pass; bounded cleanup leaves no test-service/test-smoke resource. All binary criteria are closed. |

## 16. Binary Completion Criteria

### Planning tracker completion

- [x] All sixteen required sections exist.
- [x] Repository branch, commit, state, date/timezone, migration lineage/boundary/head,
  schema, ledger, lineage relation, extensions, roles, image, mounts, and dependencies
  are recorded.
- [x] Every material PostgreSQL version/path occurrence is scheduled or explicitly
  classified as historical, fixture-only, generated/projection, not present, or
  intentional no-action.
- [x] Fresh-only posture, rejected alternatives, exact target, digest, checksum, SCRAM,
  storage layout, extension baseline, admission, Recovery, benchmark, and rollback
  decisions are closed.
- [x] Owner impact, workstreams, slices, tasks, validation, risks, and handoff evidence
  are implementation-ready.
- [x] Tracker creation is `DONE`; SL-00 is `READY`; SL-05 is `NO_ACTION`; all other
  implementation slices are `TODO`; none is `IN_PROGRESS`.
- [x] `make lint-markdown` passed with retained run root
  `.cartulary/test-results/20260904T165113Z-p56043`.
- [x] Final diff/status proves this tracker is the only changed path.

### Eventual migration completion

- [x] Core 01, Core 04, Core 05, and Testing Harness adopt the target without
  contradiction; downstream projections have zero drift.
- [x] Every live/default PostgreSQL service uses the exact 18.6 index digest,
  `/var/lib/postgresql` parent mount, version-specific `PGDATA`, checksums, and SCRAM.
- [x] PostgreSQL-only reset is dry-runnable, confirmation-gated, narrowly targeted, and
  preserves all non-PostgreSQL state.
- [x] Owned, borrowed, new, and recycled connections accept only version number 180006
  with checksums on and retain exact purpose identities.
- [x] `pgcrypto` 1.3 and `citext` 1.6 are administrator-owned in `public`; wrong/missing
  version/schema cases fail before DDL.
- [x] Unchanged migrations 1-40 apply from pristine state; Rebaseline v2, immutable
  hashes 1-29, Goose ledger, lineage, contamination, ahead/malformed state, and
  harness-only rollback-through-zero checks pass.
- [x] Application SQL, functions, triggers, GIN index, collations, locks,
  LISTEN/NOTIFY, transactions, cancellation, and catalog/telemetry queries pass.
- [x] Runtime, migration, and Recovery ACL, ownership, default privilege, future-object,
  and denial matrices show no widening and no secret leakage.
- [x] Backup creation/inspection and restore/verification/rebuild succeed on pristine
  18.6; incompatible targets fail before mutation.
- [x] The 18.6 engineering comparison is qualified; any regression is dispositioned;
  the new benchmark identity never rewrites retained PostgreSQL 16 evidence.
- [x] All applicable focused, smoke, browser, performance, generation, agent-finalize,
  `check`, harness-contract, and release-check gates pass with retained run roots.
- [x] Residual scans find no unclassified PostgreSQL 16 default or obsolete live mount,
  and cleanup finds no leaked database, volume, role, container, or service session.
- [x] SL-00 through SL-09 are `DONE` or the planned `NO_ACTION`, the final handoff log is
  current, and no unplanned behavior entered the final repair slice.
