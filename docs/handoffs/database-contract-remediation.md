# Database contract remediation handoff

## Scope and decisions

This work remediates the three database assertions retained in
`windows-v2-check-final`, plus the directly related ownership-schema and
migration-requirement drift. The starting commit is
`3bae6b5fc0720dd7d35d61747f3c366e9d85f650`, with existing Windows V2 work in the
checkout. No Reference Pack production, publication, activation, or database
reset is part of this change.

The user selected append-only committed-source associations. Runtime keeps
`SELECT, INSERT` on `import_committed_source_records`; it receives no direct
`UPDATE`, `DELETE`, or `TRUNCATE`. Parent physical deletion uses existing FK
cascades. Recovery retains `table_restore`. Reclaimable committed-source
application records, blobs and chunks keep their existing access policy.

## Requirement and implementation closure

| Gap | Owner and remediation | Executed focused evidence |
| --- | --- | --- |
| Rollback test assumed whole-chain atomicity and missed its intended barrier | Core 01 REQ-01-657/661 and Harness REQ-810/811 clarify per-step scratch transactions. Separate tests address 38→37, 47→46, committed 48→47 before rejection, and historical rollback followed by forward traversal. | Migration slice passed; ledger and semantic catalog comparisons execute independently. |
| Migration 47 lacked complete source-FK index coverage | Imports-owned migration 49 adds a B-tree index on `(incident_id, committed_source_id)` without changing keys, rows or earlier SQL. | Populated 48→49→48→49 scratch sequence preserved exact source/association data; independent catalog witness reports a valid, ready index and 345 FKs. |
| Generator asserted coverage without inspecting indexes | A pure, ordered key-bearing DDL model computes final-head coverage. Generation and source validation require exact per-FK claims; PostgreSQL independently inspects installed keys, method, validity and readiness. | Generated-drift check and Harness boundary row passed; source and database negative cases cover missing, reordered, partial, expression and INCLUDE-only indexes. |
| Association ACL classification contradicted SQL | Core 01 REQ-01-685 and Core 04 REQ-04-153 explicitly classify associations as `table_append_only`; authored generator policy now agrees with existing grants. | Purpose-role ACL row passed; populated migration test proves direct mutation denial, runtime append, Recovery restoration, soft-deletion preservation and parent cascade. Public Imports transfer also passed. |
| Active requirements and JSON Schema described obsolete catalog limits | Current migration requirements select v4 and authored head; ownership v2 schema uses filename syntax and positive versions, with exact allocation checked separately. Full JSON Schema validation now participates in source admission and pre-write generation. | Schema and migration-input checks passed. Negative schema cases reject unknown members, invalid versions and mismatched allocation. |

Existing historical SQL and history entries for versions 1–48 were compared
against Git HEAD and remain byte-identical. Historical coverage is not rewritten:
coverage admission evaluates the complete current Up catalog. This work adds no
unindexed exception or production rollback capability.

Final review also closed a dependency edge case in the new source model:
PostgreSQL drops an index when a removed column occurs only in its `INCLUDE`
list or expression. The model retains INCLUDE dependencies, follows their
renames, and rejects column drops with unresolved expression dependencies.
Explicitly drop affected indexes before dropping columns. Unsupported backing
constraint suffixes are rejected rather than silently ignored.

## Changed paths

Authored specification and documentation:

- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/04_security_deployment_and_conformance.md`
- `docs/testing-harness-nlspec.md`
- `docs/handoffs/database-contract-remediation.md`

Authored implementation, contracts and tests:

- `db/migrations/00049_import_committed_source_index.sql`
- `tools/database-migrations/owner-allocation.mjs`
- `tools/database-migrations/generate-catalog-projections.mjs`
- `tools/database-migrations/foreign-key-coverage.mjs`
- `tools/schemas/cartulary.schema_object_ownership_manifest.v2.schema.json`
- `tools/harness/generated-artifacts/database-contract-drift/schema-object-ownership.mjs`
- `tools/harness/generated-artifacts/tests/test-boundary-projections.mjs`
- `internal/modules/database_migrations/catalog_capture_test.go`
- `internal/modules/database_migrations/migrations_test.go`
- `internal/modules/database_migrations/rebaseline_manifest_integration_test.go`
- `internal/modules/database_migrations/import_source_index_integration_test.go`
- `internal/testutil/pgtest/pgtest.go` (scratch semantics comment)
- `tools/test_families/module.database_migrations.json`

Make-generated projections:

- `tools/migration_history_manifest.json`
- `tools/schema_object_ownership_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/harness/contract/generated/schema-validators/validateCartularySchemaObjectOwnershipManifestV2.cjs`
- Twenty-four downstream schema-validator files whose generated identifiers
  renumbered after the ownership-schema change. Comparing after normalization of
  generated `schema`, `validate`, and `pattern` identifiers found no other changes
  in those twenty-four files.

The pre-existing Windows V2 changes are separate from this inventory. Generated
files were refreshed through Make, never edited directly. SQLC output has no
additional delta from the index-only migration.

## Identities and installed facts

| Artifact | SHA-256 of exact bytes |
| --- | --- |
| Migration 49 | `352716f0fc8743aae4b945f216459ebf242665fde91d021a3df8074f3e461659` |
| Migration-history manifest | `010640aa8e9c8c7d825b046a345792b9bff2228094b718de289dd3cf9d4fea79` |
| Ownership manifest v2 | `7e91270a40fc7713879b4ca9c6a020c4eaec308627e8f9f0faf4e74028a6b8b4` |

The current catalog has 49 migrations, 2,096 ownership entries and 345 covered
foreign keys. The independent live witness in
`.cartulary/test-results/db-contract-migrations-01/postgres-catalog-state.json`
records the new index as non-unique, valid and ready, with the exact definition:

```sql
CREATE INDEX import_committed_source_records_source_fk_idx
ON public.import_committed_source_records USING btree (incident_id, committed_source_id)
```

## Executed commands and evidence

Each command below used `CARTULARY_TEST_RUN_ID` equal to its listed run ID. Roots
are `.cartulary/test-results/<run-id>`. Exit is the observed outer process exit;
work-unit counts are not test counts.

The final source identity is
`sha256:41715ab0dbee4a6b517e1d2b0c45e453d2486c5ca5acc80220ca70c75d3ff026`.
The migration and boundary `-02` slices have this exact source identity, as does
the final `db-contract-check-final-02` invocation. The earlier full check used
`sha256:242031a21fefb07a7370640cf1a35d4739cf87feb46b7fc1fe3b27ea75b01830`.

The migration selection was:
`module.database_migrations.integration.production_preflight_state_matrix`,
`module.database_migrations.integration.production_ddl_v3_recurrence`, and
`module.database_migrations.unit.production_ddl_v3_recurrence`.

| Run ID | Command/selection | Exit and result |
| --- | --- | --- |
| `db-contract-generate-01` through `db-contract-generate-04` | `make generate` | 2; development failures identified unsupported existing trigger, NULLS-index and CTAS forms. Corrected before successful generation. |
| `db-contract-generate-05` | `make generate` | 2; newly edited test selectors required ASCII sorting. Corrected authored routing. |
| `db-contract-generate-06` and `db-contract-generate-07` | `make generate` | 0; passed. |
| `db-contract-format-01` | `make format` | 0; passed. |
| `db-contract-format-02` | `make format` | 0; passed after the index-dependency review correction. |
| `db-contract-generate-08` | `make generate` | 0; passed after the index-dependency review correction. |
| `db-contract-migrations-01` | `make test-slice OWNER=module.database_migrations ROWS=<migration selection above>` | 0; 4 work units passed. |
| `db-contract-acl-01` | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.purpose_identity_roles_and_acl` | 0; 3 work units passed. |
| `db-contract-imports-01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.committed_source_transfer` | 0; 3 work units passed. |
| `db-contract-schema-01` | `make source-check CHECK=schema` | 0; 3 work units passed. |
| `db-contract-input-01` | `make source-check CHECK=migration-input` | 0; 3 work units passed. |
| `db-contract-generated-01` | `make source-check CHECK=generated` | 0; 5 work units passed, including regeneration drift comparison. |
| `db-contract-markdown-01` | `make lint-markdown` | 0; passed before this handoff was added. |
| `db-contract-harness-01` | `make harness-contract SUITE=contract` | 2; 12/13 work units passed; one separate pre-existing Reference Pack measurement-contract failure described below. Cleanup: 22 steps, zero failed/blocked. |
| `db-contract-finalize-01` | `make agent-finalize` | 0; passed. No failed check root was supplied as successful retained maintenance evidence. |
| `db-contract-boundaries-final` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | 0; passed after finalization. This row is absent from `make check` and provides the additional generator/schema contract qualification. |
| `db-contract-markdown-final` | `make lint-markdown` | 0; passed after this handoff was initially added. |
| `db-contract-check-final` | `make check` | 0; 338/338 work units passed in 981.75 seconds; 573 cleanup steps, zero failed/blocked. This source preceded the final index-dependency review correction. |
| `db-contract-migrations-02` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v3_recurrence` | 0; 3/3 work units passed, including the PostgreSQL index-dependency cases; 14 cleanup steps, zero failed/blocked. |
| `db-contract-boundaries-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries` | 0; 4/4 work units passed, including column-drop/rename dependency rejection and explicit index removal; 13 cleanup steps, zero failed/blocked. |
| `db-contract-finalize-02` | `make agent-finalize` | 0; passed after the index-dependency correction. |
| `db-contract-check-final-02` | `make check` | 130; cancelled at the user's request. 29 work units passed, zero failed, 309 cancelled. All 65 cleanup steps completed with zero failed/blocked. |

Discovery also executed `make task-guide`, `make target-plan TARGET=check
DETAIL=rows JSON=1`, and retained-run observation. These are selection/diagnosis,
not passing test evidence. The final check is incomplete, not a passing result.
The earlier full check passed before the final index-dependency correction;
focused migration and boundary checks passed after that correction on the final
source identity. The user stopped further verification. Final handoff-only
edits were therefore not followed by another Markdown lint invocation.

Cancellation used SIGINT on the exact run's verified process group after native
session interrupts did not stop it. The owning process exited 130, and the
Harness classified the outcome as `interrupted` / `cancelled_or_interrupted`.
Read-only `make explain-run RESULTS_DIR=.cartulary/test-results/db-contract-check-final-02
DETAIL=summary` exited 0 and reported successful cleanup. No further tests ran.

## Separate remaining blocker

`harness.browser.behavior.artifact_assurance` fails in
`test-measurement-proofs.mjs:271`: the performance contract expects
`base-2026-10-02.1`, while the browser fixture reads built-in release
`base-2026-11-06.1`. The payload digest agrees; the base-version identity does not.
The failure is `product/test_assertion_failure`, not fixture admission or cleanup.
All four involved source files are unchanged from the starting commit:

- `contracts/performance/reference-pack-progress.v1.json`
- `contracts/reference-packs/builtins/release.v2.json`
- `tools/harness/browser/reference-pack-fixture.mjs`
- `tools/harness/browser/tests/test-measurement-proofs.mjs`

Closure requires the performance owner to reconcile the authored workload with
the intended exact Base release, regenerate affected projections, and execute the
artifact-assurance row. Do not remove the exact-version comparison or report the
broader Harness contract suite as passing. The retained failure is under
`db-contract-harness-01/unit-logs/row-harness.browser.behavior.artifact_assurance/stderr.log`.

## Operational handoff

Migration 49 preserves existing rows and grants. Its ordinary transactional index
build needs additional disk space and can block concurrent writes. Apply it using
the normal explicitly selected deployment migration workflow, with an appropriate
write-maintenance window. A deployment left at version 48 remains behind the
repository head and must not be presented as ready for the new executable.

No deployment database was migrated during this work; service-backed executions
used Harness-owned disposable databases. The Down section is scratch cleanup,
not an operator rollback or recovery promise. There is no automatic reset,
historical conversion, new public API, or Reference Pack identity change.
