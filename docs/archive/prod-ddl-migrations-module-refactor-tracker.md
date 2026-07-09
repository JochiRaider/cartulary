# Production DDL Rebaseline Tracker and Handoff

## 1. Authority and Decisions

- Owner authorization: the implementation prompt explicitly approved rewriting/rebasing historical migrations.
- Applied database evidence: no deployed/shared database applied-version evidence was supplied for this rebaseline. Compatibility policy is therefore reset/reimport, not in-place upgrade.
- Rebaseline shape: a single owner-grouped runnable line in `db/migrations`, followed by future forward migrations in the same directory.
- Historical archive policy: rely on git history and this handoff; do not maintain a second runtime migration package.
- Current lineage: `cartulary.prod_ddl_rebaseline.v1`, with remediation boundary `prod_ddl_rebaseline_v1`.
- Legacy policy: preserve only current owner-backed behavior. Retired historical upgrade mechanics are not current product behavior.

## 2. Current Migration Line

The current line has 23 authored migrations:

1. `00001_database_infrastructure.sql`
2. `00002_auth_accounts_and_enterprise.sql`
3. `00003_deployment_admin_and_recovery.sql`
4. `00004_incidents_and_preferences.sql`
5. `00005_platform_jobs.sql`
6. `00006_record_revision_substrate.sql`
7. `00007_timeline_source.sql`
8. `00008_entity_source.sql`
9. `00009_indicator_source.sql`
10. `00010_assessment_source.sql`
11. `00011_links_and_tags.sql`
12. `00012_parties.sql`
13. `00013_tasks_and_decisions.sql`
14. `00014_artifacts_and_optional_surfaces.sql`
15. `00015_evidence_and_object_blobs.sql`
16. `00016_workbook_projections.sql`
17. `00017_saved_views.sql`
18. `00018_imports.sql`
19. `00019_reference_data.sql`
20. `00020_graph_projection.sql`
21. `00021_reporting.sql`
22. `00022_report_composition.sql`
23. `00023_incident_bundles.sql`

The first migration creates `schema_migration_lineage` and inserts the current lineage marker. Goose `Down` sections remain goose reversibility only and are not production rollback evidence.

## 3. Runtime Behavior

`db/migrations/source.go` exposes the embedded SQL source plus expected lineage metadata. `internal/platform/postgres` checks that metadata before running migrations.

Fresh databases have no applied goose version and can apply the current line. Partial databases already on the current line must retain the expected lineage row and can continue forward. Databases with applied goose versions but no matching lineage fail before migration SQL runs with a structured remediation report:

- `schema_id`: `cartulary.migration_remediation_report.v1`
- `boundary`: `prod_ddl_rebaseline_v1`
- `reason_code`: `historical_migration_lineage`
- remediation: reset the database or use an explicit export/import path before adopting the rebaseline.

There is no goose-ledger bridge and no maintained legacy migration line.

## 4. Ownership Cleanup

`tools/schema_object_ownership_manifest.json` now has narrower owners for:

- `database_migrations`
- `parties`
- `tasksdecisions`
- `artifacts`
- `reportcomposition`

`links` now owns record links and tags only. Workbook projection tables remain under `projections`; source owners do not claim read-model tables.

## 5. Gap Status

| ID | Status | Notes |
| --- | --- | --- |
| RB-001 | RESOLVED | Owner authorization recorded. Historical chain replaced. Existing applied DBs must reset/reimport. |
| RB-002 | RESOLVED | Runnable artifact is the owner-grouped baseline in `db/migrations`; no docs-only baseline and no dual package. |
| RB-003 | IMPLEMENTED | Historical upgrade fixtures and runtime preflights were removed. Current schema/route behavior remains covered by current tests. |
| RB-004 | RESOLVED | Ownership manifest split broad coordination ownership into narrower owners. |
| RB-005 | RESOLVED | Historical version-anchor constants were removed from `source.go`; tests use lineage/current-schema behavior. |
| IG-001 | RESOLVED | SQLC and derived contract generation completed through Make; generated drift passed. |
| IG-002 | IMPLEMENTED | Source-owner DDL and projection/report-composition ownership are separated in the manifest and baseline ordering. |

## 6. Validation Log

Run roots from this implementation pass:

- `make phase-schedules`: passed at `.cartulary/test-results/20260709T020702Z-p40691`.
- `make json-shape-check`: initially failed on stale phase schedules, then on two uncovered ownership patterns; passed after fixes at `.cartulary/test-results/20260709T020744Z-p45096`.
- `make generate`: final rerun passed at `.cartulary/test-results/20260709T021938Z-p85366`.
- `make build-migrate`: final rerun passed at `.cartulary/test-results/20260709T022004Z-p88425`.
- `make migration-drift`: final rerun passed at `.cartulary/test-results/20260709T022004Z-p88447`.
- `make generated-artifact-policy-check`: final rerun passed at `.cartulary/test-results/20260709T022004Z-p88477`.
- `make generate-drift`: final rerun passed at `.cartulary/test-results/20260709T022004Z-p88509`.
- `make json-shape-check`: final rerun passed at `.cartulary/test-results/20260709T022004Z-p88574`.
- `make phase-schedule-drift`: passed at `.cartulary/test-results/20260709T021727Z-p70546`.
- `make lint-scripts`: passed at `.cartulary/test-results/20260709T021728Z-p70649`.
- `make backend-module-boundary-check`: final rerun passed at `.cartulary/test-results/20260709T022020Z-p94730`.
- `make test-fast`: first failed on a stale migration filename test and then on a migration-evidence expected-count assertion; passed after fixes and final SQLC regeneration at `.cartulary/test-results/20260709T022021Z-p94919`.
- `make agent-finalize`: final rerun passed at `.cartulary/test-results/20260709T022310Z-p62000`; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- `make lint-markdown`: passed during final documentation validation; final run root is reported in the session handoff.

Broader backend integration, browser, and full `make check` were not run in this pass. Current validation covered the required rebaseline, codegen, boundary, script, and fast test surfaces.

## 7. Handoff Notes

- Generated roots must remain downstream of Make generation.
- Historical databases are intentionally incompatible with the new line unless reset or exported/imported.
- If production data migration is later required, it must be designed as an explicit export/import workflow, not as an implicit migration bridge.
- Keep future migration files owner- or behavior-shaped and preserve the current lineage policy unless a later owner-authorized rebaseline supersedes it.
