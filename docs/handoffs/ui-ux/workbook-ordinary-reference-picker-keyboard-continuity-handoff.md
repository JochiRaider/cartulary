# Ordinary reference-picker keyboard continuity

## Baseline and authority

Implementation started from clean `main` at `4daa23ed50ac96d77fd3ed3a8a5d90352020e56c` (ahead of `origin/main` by one commit). The Evidence Collector Party journey reproduced at 1440 × 900 and 1024 × 720: native Tab reached the sole staged Remove button, Enter or Space unmounted it, and the candidate selector was not focused. The failing browser row is in `.cartulary/test-results/20260928T163632Z-p92532`; the controlled-read red run is `.cartulary/test-results/20260928T163823Z-p25966`.

The governing owners are `docs/design.md` §§12.4–12.5 and 14; Core 03 §2.3A, applicable §13, and REQ-03-299/100; Core 01 reference and view contracts, projected by `contracts/view-references/index.json`; and Core 02 field semantics. `docs/domain.md` and the UI/UX digest supplied vocabulary and navigation. No owner contradiction or specification amendment was found. The NLSpec research essay, historical handoffs, and bundled prompts were source material only.

## Correction and ownership

- `WorkbookReferenceControl.tsx` now passes stable `workbookReferenceKey` identities to the existing `useSelectedReferenceRemovalFocus` helper. A focused removal moves to the next surviving Remove button, then the previous, or to the enabled candidate selector. An empty page uses a labelled local focusable group. Duplicate visible labels gain identity in their Remove names. The helper reveals focus only inside the local picker scrollport.
- The ordinary picker has a presentation-only read intention for First, Previous, Next, and Retry. Its initiating control remains mounted and focusable during an admitted read and, when exhausted, while it remains focused. Stable names, `aria-busy`, and `aria-disabled` distinguish pending or unavailable actions; synchronous guards prevent another request. Retry survives failure clearing at read start and can be activated after repeated failure. Focus departure, source/filter replacement, parent replacement, closure, and concealment retire the intention. Read completion never reclaims focus.
- `WorkbookReferenceSelection`, the read port, view contracts, and consumers retain page, cursor, failure, selection, cancellation, and authority decisions. The picker did not acquire a selection store or retry engine. The unhandled Remove path and native pending-disable path were retired for this picker; the already-corrected candidate browser remains separate. No Refresh, filter, page-limit, selection-limit, First, or retry-eligibility change was made.
- Native Tab/Shift+Tab, Escape, select interaction, and parent return remain under the popup; the local empty-candidate fallback does not trigger Tab wrap. Browsing, removal, and Cancel do not write. Inspector and Task acceptance still update drafts; the compact grid still uses its existing commit callback. Source, filter, record, field, selection order, explicit clearing, and protected-data behavior remain owner-controlled.

No route, schema, storage, database, or persistence migration is needed. Roll back by reverting the bounded picker, test, authored routing, and generated routing-index changes together; no data conversion is needed.

## Evidence and digest acceptance

| Rows | Assessment and evidence |
| --- | --- |
| A001–A003 | PASS. Current owners, manifests, source/import boundaries, clean baseline, and consumer callbacks were checked before placement. The change is one picker presentation seam. |
| A004–A005, A009, A011 | PASS. No new design/theme registry or scroll owner. Browser focus and page/grid scroll observations cover local continuity; source, parent, closure, and late-read checks cover lifetime boundaries. |
| A014, A016–A017, A019 | PASS. Controlled unit and production-browser rows cover raw draft/Cancel, held and exhausted controls, repeated failure, authority concealment, keyboard focus, stable names, and visible focus. |
| A023–A027 | PASS. Tests use semantic roles and stable identity; authored test-family selectors are routed independently of docs; `make generate` produced only the expected routing-index hash update; generation and policy drift checks pass. This handoff records rollback and exact evidence. |
| A006–A008, A010, A012–A013, A015, A018, A020–A022 | N/A. Density, creation, responsive chrome, inspector dispatch, mutation replay/acknowledgement, conflict, Evidence file lifecycle, component redesign, virtualization, and golden publication are outside this bounded picker change. Two viewport keyboard checks and a seeded rendering were still reviewed. |

The seeded `cartulary-ui-review` session `.cartulary/test-results/20260928T165006Z-p79484` showed the focused Remove control and, after Space, the focused candidate selector in a 1440 × 900 image. Its focus observation identified the selector as the active element. The exact session stopped with `state=closed`, `status=ok`; private screenshots expired and caller scratch was removed. This review is visual evidence, not a product-test or accessibility-conformance claim.

## Verification

All commands ran from the repository root through public Make targets. Current successful runs:

| Command | Result and run root |
| --- | --- |
| `make format` | PASS `.cartulary/test-results/20260928T172109Z-p63646` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.reference_controls,web.workbook.regression.reference_selection` | PASS `.cartulary/test-results/20260928T172119Z-p68097` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.reference_selection_recovery,module.workbook.browser.reference_grid_commit` | PASS `.cartulary/test-results/20260928T171400Z-p61720` |
| `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.task_lifecycle_patch_recovery` | PASS `.cartulary/test-results/20260928T171501Z-p96797` |
| `make generate` | PASS `.cartulary/test-results/20260928T171859Z-p48013` |
| `make agent-finalize` | PASS `.cartulary/test-results/20260928T172133Z-p68791`; `RESULTS_DIR` unset |
| `make frontend-typecheck` | PASS `.cartulary/test-results/20260928T172157Z-p73108` |
| `make frontend-import-boundary-check` | PASS `.cartulary/test-results/20260928T172157Z-p73142` |
| `make lint-biome` | PASS `.cartulary/test-results/20260928T172157Z-p73238` |
| `make json-shape-check` | PASS `.cartulary/test-results/20260928T172157Z-p72922` |
| `make generate-drift` | PASS `.cartulary/test-results/20260928T172157Z-p72972` |
| `make generated-artifact-policy-check` | PASS `.cartulary/test-results/20260928T172157Z-p73051` |
| `make lint-markdown` | PASS `.cartulary/test-results/20260928T172238Z-p79369`; rerun after this update is reported separately |
| `git diff --check` | PASS after source and generated-index changes; rerun after this update is reported separately |

The first expanded browser read assertion failed in `.cartulary/test-results/20260928T164400Z-p63629` because Playwright `selectOption` did not focus the source selector; the test now explicitly focuses it. The first final unit run failed in `.cartulary/test-results/20260928T170001Z-p95611` because a programmatic focus change was asserted before React flushed its blur update; the assertion now uses `act`. Both rows subsequently passed. No golden was refreshed.

The later `make check` run failed with 964/970 units passing at `.cartulary/test-results/20260928T180220Z-p91269` (`run-summary.json`). Its picker rows `web.workbook.regression.reference_controls` and `web.workbook.regression.reference_selection` passed. Six independent units failed:

| Unit | Finding |
| --- | --- |
| `package.grid_adapter.boundary_support.manual_package_surface_reachability` | `frontend-fallow-static` reported unused `classifyPreviewProbe` in `apps/web/src/workbook/features/evidence/resolveSolePreviewableEvidence.ts`; package-surface findings were zero. See `frontend-fallow-static/tool-run-summary.json`. |
| `web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039` | Two unowned test-id literals in `WorkbookAuthoringReferencePicker.test.tsx` and six raw dynamic selectors in `workbook-column-sizing.spec.ts`. See the row's `vitest-failure-details.json`. |
| `web.workbook.regression.timeline_autosave_subscription_isolation` | Timeline composition expected no unchanged publications but observed three. See the row's `vitest-failure-details.json`. |
| `module.records.architecture.boundary` | Forbidden `collaboration_event_intents` source token in `internal/modules/entities/candidate_discovery_integration_test.go`. See the row's `stderr.log`. |
| `web.workbook.regression.workbookshell_surfaces_suite_668e482b1e` | Evidence Verify attach preview state was `Available` when the test expected `Preview unavailable`. See the row's `vitest-failure-details.json`. |
| `target:go-vulncheck` | Blocking symbol-reachable `GO-2026-6508` finding against OpenTelemetry log gRPC exporter `v0.20.0`; scanner reports fixed `v0.21.0`. See `unit-artifacts/target-go-vulncheck/govulncheck-findings.json`. |

These failures point to files and dependencies outside the picker diff; the Timeline and WorkbookShell assertions could merit isolated reruns before classifying them as persistent. The smallest next check is an owner-routed rerun of those two rows, with the other four findings handled by their respective owners. `make check` did not yield a successful full warm run, so retained-run maintenance remains skipped with `RESULTS_DIR` unset. The visual-golden suite was not run because no intended golden change was made.
