---
doc_id: THR-S0-SOURCE-LIMITS
title: Testing Harness Recovery Source-Limit Log
status: active
role: source-limit-log
---

# Testing Harness Recovery Source-Limit Log

## Document role

This log records unavailable, incomplete, or intentionally uninspected sources
during testing harness reverse-specification recovery. Do not delete unresolved
rows. Mark resolved limits with a follow-up note instead.

## Source limits

| Source-limit ID | Surface | Limit type | What was inspected | What was not inspected | Impact | Follow-up needed | Evidence status | Notes |
|---|---|---|---|---|---|---|---|---|
| SL-0001 | `.github/**` | `inaccessible_file` | S0 checked repository root with `test -d .github && find .github -maxdepth 3 -type f \| sort \|\| printf '.github absent\n'`. S1 rechecked with `test -d .github && find .github -maxdepth 3 -type f \| sort \|\| true`, which printed no files. S1 also inspected `scripts/ci/**`. | CI workflow files under `.github/**`; directory was absent. | S1 cannot record repository CI workflow behavior from `.github` sources. S2 must classify CI as absent, external, or represented by provider-neutral scripts if no other evidence appears. | During S2, map `make ci` and `scripts/ci/**`, then ask for owner decision if provider CI mapping is required. | `source_limit` | Required S0/S1 input named CI configuration, but no `.github/` directory exists in the working tree. See `AMB-0001`. |
| SL-0002 | Broad verification commands | `unexecuted_command` | S1 ran non-mutating discovery commands: `make help`, `make help-all`, `make task-guide`, `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `make target-plan`, `make target-plan-json`, `make explain-phase PHASE=phase0` through `phase6`, and selected `make explain-target ... DETAIL=summary` commands. | Runtime execution of `make test-fast`, `make test`, `make check`, `make ci`, `make release-check`, browser E2E targets, backend test targets, frontend unit/typecheck targets, service-backed targets, lint targets, and build targets. | S1 inventory is static/discovery-based. It does not prove runtime success/failure behavior or failure bundle contents. | S2 should trace command declarations without executing broad gates. S3/S5 may inspect retained artifacts or run targeted commands if later authorized. | `source_limit` | S1 intentionally avoided broad verification and implementation-changing commands. |
| SL-0003 | Live service-backed runtime behavior | `unexecuted_command` | S1 inspected `docker-compose.dev.yml`, `configs/dev/**`, `Makefile`, `tools/testservices/**`, `internal/testutil/pgtest/**`, `internal/testutil/s3test/**`, `internal/testutil/testcontainersx/**`, `apps/web/playwright*.config.ts`, and `scripts/start-web-e2e.sh`. | Live Docker, Postgres, MinIO, browser stack, port allocation, readiness, reset, retry, and teardown behavior during an actual service-backed run. | Service lifecycle conclusions remain static until S4/S6 recover runtime evidence. | S4 should build `service_lifecycle_map`; S6 should record resource/timing hazards. | `source_limit` | Static inspection found service surfaces, but S1 did not start backing services. |
| SL-0004 | Retained `.cartulary/test-results/**` artifacts | `partial_search` | S1 listed existing retained run summaries under `.cartulary/test-results/**`, inspected ignore rules with `git check-ignore -v`, and observed Make default `CARTULARY_TEST_RESULTS_DIR`. | Full provenance, freshness, completeness, and schema coverage of retained artifacts; failure-only traces/screenshots/log bundles not exhaustively opened. | Existing retained artifacts cannot be treated as authoritative current run evidence. | S3/S5 should recover artifact ownership, schema, freshness, and consumer rules. | `source_limit` | `.cartulary/test-results/**` is ignored by the unanchored `test-results/` rule. See `AMB-0002` and `AMB-0010`. |
| SL-0005 | Planned phase7/phase8 files | `inaccessible_file` | S1 inspected `tools/phase_registry.json` and checked `tools/phase7_test_map.json`, `tools/phase8_test_map.json`, `docs/testing/phase7_coverage_ledger.md`, and `docs/testing/phase8_coverage_ledger.md` with `test -e`. | Planned phase7/phase8 manifest and ledger files; all four named files were absent. | Phase7/phase8 cannot be inventoried as active harness files during S1. | Later phase owner decision or future phase creation should resolve whether planned rows belong in the recovered harness spec. | `source_limit` | Active phase evidence is phase0 through phase6. See `AMB-0004`. |
