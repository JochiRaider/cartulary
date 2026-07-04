---
doc_id: THR-S1-UNINVOKED
title: Testing Harness Recovery Uninvoked Surface List
status: active
role: uninvoked-surface-list
---

# Testing Harness Recovery Uninvoked Surface List

## Document role

This S1 artifact records harness-like surfaces that were not invoked by the
entrypoints discovered during inventory, or whose invocation is indirect enough
that S2 must trace it before treating the surface as active behavior.

Rows use cautious labels. This document does not declare dead code.

## Discovery method

- `make task-surface-report TASK_SURFACE_REPORT_ARGS=--all` was used as the
  primary task-surface evidence.
- `rg --fixed-strings` checks searched for direct references from Make,
  manifests, scripts, package files, app code, docs, and tools.
- Imports and test-runner pattern invocation were considered separately from
  direct file-path references.

## Uninvoked or orphan-like surfaces

| Surface ID | Surface | Discovery result | Provisional classification | Evidence | Evidence status | Follow-up |
|---|---|---|---|---|---|---|
| US-0001 | `.github/**` | No `.github/` directory was present in the working tree. | `source_limit` | `test -d .github && find .github -maxdepth 3 -type f \| sort \|\| true` returned no files. | `source_limit` | S2 must decide whether CI is external/provider-owned, absent, or represented by `scripts/ci/**`. |
| US-0002 | `scripts/test-run-go-target-fast.sh` | Direct path and basename searches found no caller in Make, task manifests, scripts, app code, package files, or docs. | `uninvoked_by_discovered_entrypoints` | `rg -n "test-run-go-target-fast\|scripts/test-run-go-target-fast" Makefile tools scripts docs apps package.json pnpm-workspace.yaml` returned no matches. | `observed` | S2 should determine whether this is retired harness smoke coverage, a missing task-surface row, or intentionally manual. |
| US-0003 | `scripts/.gitkeep` | Placeholder file; no execution role. | `placeholder` | `find scripts -maxdepth 3 -type f`; direct reference search. | `observed` | No S2 command tracing needed unless directory ownership policy is being documented. |
| US-0004 | `scripts/render-phase-ledger.mjs` as a standalone CLI | No public Make target invokes this file directly, but it is imported by `scripts/render-phase-ledgers.mjs` and `scripts/check-phase-ledger-drift.mjs`. | `library_only_not_standalone_entrypoint` | `rg -n "render-phase-ledger\|scripts/render-phase-ledger" ...` showed imports but public targets call plural `render-phase-ledgers.mjs`. | `observed` | S2 should trace the plural entrypoint and mark the singular file as an implementation module, not an independent command. |
| US-0005 | `tools/phase7_test_map.json`, `tools/phase8_test_map.json`, `docs/testing/phase7_coverage_ledger.md`, `docs/testing/phase8_coverage_ledger.md` | `tools/phase_registry.json` names planned phase7/phase8 paths, but the files are absent. | `planned_missing_surface` | `jq '.' tools/phase_registry.json`; explicit `test -e` checks reported all four paths absent. | `source_limit` | Record in S2/S7 as planned future phase coverage, not current active harness surface. |
| US-0006 | Retained `.cartulary/test-results/**` directories from earlier runs | Existing ignored artifacts are read by investigation commands when supplied, but are not source entrypoints themselves. | `retained_runtime_artifact_not_entrypoint` | `find .cartulary -maxdepth 3 -type f`; `git check-ignore -v .cartulary/test-results`. | `observed` | S3/S5 should recover artifact schema/authority and explain-run consumption rules. |

## Cleared candidates

| Surface | Why not listed as uninvoked | Evidence |
|---|---|---|
| `apps/web/e2e/*.test.ts` | These files are not Playwright specs, but `apps/web/vite.config.ts` includes `e2e/**/*.test.ts` in the Vitest `harness-node` project. | `sed -n '1,200p' apps/web/vite.config.ts` |
| `apps/web/e2e/*.spec.ts` | These are selected by Playwright configs and browser batch manifests rather than direct Make path references. | `apps/web/playwright.shared.config.ts`; `make explain-target TARGET=browser-e2e-webserver-backed` |
| `internal/testutil/**` helpers | Most helpers are invoked through Go package imports, not file paths. | `rg` import/use searches and Go test file inventory |
| `tools/*.json` manifests | Every top-level `tools/*.json` manifest checked had some discovered reference or task-surface role; planned phase7/phase8 paths are the exception because the files do not exist. | `task-surface-report --all`; `rg --fixed-strings` reference checks |
