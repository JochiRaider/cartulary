# FE-P11: Visual, Accessibility, And Readiness Gates

## Summary

FE-P11 is a frontend verification-readiness phase for visual fixtures,
accessibility coverage, density/theme/typography/status patterns,
command-surface integration, measurement-support separation, Core 05
publication-boundary protection, and readiness-gate composition. It introduces
no new product behavior.

This file is an implementation plan and handoff artifact only. It does not
implement FE-P11 behavior, close blocked rows, activate FE-P11, update generated
ledgers, edit generated protocol artifacts, refresh visual goldens, or claim
product conformance from visual, accessibility, measurement, readiness, retained
artifact, or broad-check evidence.

The live FE-P11 contract is
`docs/guides/cartulary_frontend_implementation_testing_guide.md` Section 4.11
plus `tools/frontend_phase_maps/fe_p11_test_map.json`. Current live FE-P11
posture is `status=planned`, `row_rollup_state=partially_implemented`, with
activation blocker `FE-P11-ACTIVATION-BLOCKER-01`. Current digest mirrors are:

| Source | SHA-256 |
| --- | --- |
| FE-P11 map | `dc3b7305ab850254ad39d90327373656e3d3843684b41df6d480a59acec1d03b` |
| FE-P11 generated ledger | `1baa338199c4ed706fbb4dc6ca5b42652a657283ce06a6719d8e082cd1486313` |
| FE-P11 evidence freshness | `2445ab0b3deae4f82ba6cbee8755ddf98c7d7b8a5fd41baf8cf4ffcd7b7a44e1` |

`FE-V-P11-03` is the only live implemented FE-P11 row. It closes as
design-direction readiness only from current `browser-e2e-visual` frontend row
accounting. The remaining six FE-P11 rows are blocked in the live map and must
not be represented as complete.

## Authority Model

Core 00 through Core 04 own product behavior. FE-P11 must not add product
runtime semantics through guide text, plan text, visual fixtures, accessibility
checks, readiness gates, or retained artifacts.

Core 05 applies only to claim-bearing timed, benchmark, fixture-sensitive,
visual-publication, or publication evidence. FE-P11 includes one
claim-publication-boundary support row, but the live map sets
`claim_publication_intent="none"`, so Core 05 is a boundary-protection source,
not an activated publication claim.

`docs/testing-harness-nlspec.md` owns harness mechanics only: command
invocation, target selection, scheduling, fixture lifecycle, artifact emission,
cleanup, frontend row-accounting, retained-run validation, and verification
gates. It does not own FE-P11 product behavior.

Design, visual, accessibility, measurement, readiness, and retained artifacts
are not product conformance unless the live row map explicitly says so. The
live FE-P11 map has no `product_conformance` rows.

Base `phase11` artifacts under `tools/phase11_test_map.json`,
`docs/testing/phase11_coverage_ledger.md`, and base phase result roots are
separate from frontend `FE-P11`. They must not be used to close frontend rows.

`docs/opentelemetry-instrumentation-nlspec.md` is telemetry-subsystem authority
only. It does not redefine FE-P11 visual, accessibility, readiness, product, or
Core 05 claim-publication boundaries.

## Sprint 1 Repo Baseline

Before this file was authored, `FRONTEND_PHASE11_IMPLEMENTATION_PLAN.md` was
absent and `git status --short` produced no output. Prior FE-P6 through FE-P10
plans were inspected only as style and handoff examples; their row closure,
evidence roots, target status, and digests are not FE-P11 facts.

Live FE-P10 dependency context from the frontend registry: `status=active`,
`row_rollup_state=active_green`, activation blockers `[]`, map digest
`42cfca5547fda580aaf3388826fd0bf81d71bff33976480ebcf0a91d51067a5e`, ledger
digest `83768f981e59a603ef3619c4ab69afc7179aaf3c8978c78d277ead365ee34af2`,
and freshness digest
`23cd90b9d276b65223da638e515e4ac562973b2d85f12439aca4e387102c849f`.
FE-P10 is dependency context only; FE-P11 needs its own row-owned evidence.

Primary source digests inspected:

| Source | SHA-256 |
| --- | --- |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `459ffe40579a33702160d51d5033145e15e48573dfc4e5c97c481e4781062462` |
| `tools/frontend_phase_maps/fe_p11_test_map.json` | `dc3b7305ab850254ad39d90327373656e3d3843684b41df6d480a59acec1d03b` |
| `tools/frontend_phase_registry.json` | `ea0f5dc922a1b29a30f421c927bc29690067c6fe7201a13f521dd47dcbc87477` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p11_coverage_ledger.md` | `1baa338199c4ed706fbb4dc6ca5b42652a657283ce06a6719d8e082cd1486313` |
| `tools/frontend_visual_fixture_registry.json` | `4baf8e10bf2676ae26b68c543d8e4bcefc303f8dec2b3596a4ff6608d9eea26c` |
| `docs/testing-harness-nlspec.md` | `f8857f2d67316ba43ac9c7da71040b26fb0f66250c991508b6570d3cf367af83` |
| `docs/opentelemetry-instrumentation-nlspec.md` | `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451` |
| `docs/design.md` | `e28345fac8ba22fc58264454237af209360a84af0c714ff4e1c94c6028d8cd05` |
| `docs/guides/cartulary-dev-guide.md` | `a4b8fb4b9e3b03c905ed276d19a692559ddf1e70396f224f8f8b2a3f68e58776` |
| `docs/guides/cartulary-ui-ux-design-guide.md` | `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26` |
| `docs/guides/cartulary_visual_golden_maintenance.md` | `b9a372d62fb890e72de140e29da2319fd8e202083e96819675e1edf1f64e07c3` |
| Core 00 | `e3b2e5e9ed4f47d29694612d571f3255437a9a1acbceb31fc38d9229756a682f` |
| Core 01 | `1c55b261681c59e948356d8f80e2d3f5ab8936d33db5742d18a31f701a81bac9` |
| Core 02 | `bb92665e26804b8c465d961fdef39b78e3f07c389a26a6478e9a210ce393d3fa` |
| Core 03 | `fb561f66e61cf75e777a8c1c4d618d1064ca3b36e8d02435e85131ce631f5b10` |
| Core 04 | `ab4d03966850879625141165d7902f108cfe989914e2b01ed42e2ff7968f6da1` |
| Core 05 | `ee2f572430b75b41ccd20d4dede9c72251b3a4432db2ccf525bec9415da7ef89` |

Command ledger:

| Command | Result | FE-P11 planning use |
| --- | --- | --- |
| `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P11` | pass | Reports planned phase guidance and selected-row execution hints. |
| `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P11` | pass | Reports seven rows, one implemented row, six blocked rows, and planned-phase non-executable posture. |
| `make explain-target TARGET=<mapped-target> DETAIL=summary` | pass | All mapped FE-P11 targets are explainable; `release-check`, `browser-e2e-measurement`, and `benchmark-claim-check` report no latest artifact. |
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260613T180901Z-p60996`; generated ledgers matched owner inputs. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260613T180901Z-p60994`; JSON shape and digest mirrors passed. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260613T180901Z-p60995`; generated-artifact policy passed. |
| Static `rg`, `sed`, and `jq` inspections | pass with one discarded probe | One discarded `jq` probe against `check/target-summary.json` failed because `.children` was not an array; no plan fact relies on it. |

## Source Limits

Generated ledgers are downstream companions, not row-closure authority. Update
authored owner inputs first, then regenerate through Make-owned targets when a
generated ledger refresh is required.

The live `tools/frontend_phase_maps/fe_p11_test_map.json` has `guide_digest`
but no top-level `guide_path`. Do not invent a `guide_path` field in FE-P11
facts or downstream handoff text.

The latest inspected `browser-e2e-a11y-preflight` pass emits
`cartulary.frontend_accessibility_preflight_summary.v1`, but its target summary
has no `cartulary.frontend_row_accounting` extension. It does not close
`FE-A11Y-P11-01`.

`release-check`, `browser-e2e-measurement`, and `benchmark-claim-check` need
fresh execution before any FE-P11 closure or readiness claim can cite them.

Retained artifacts, target explanations, base `phase11` artifacts, screenshot
files, visual fixture `current` status, `make check`, and this plan do not
close FE-P11 rows without current mapped row-owned evidence.

Use `TODO:` only for live facts still missing after inspection.

## FE-P10 Handoff Inputs

FE-P10 is dependency context only. Its active-green registry state shows that
the previous frontend phase is no longer the FE-P11 activation blocker, but
FE-P10 evidence must not be imported as FE-P11 row closure.

The FE-P10 strict non-claims continue to apply to FE-P11 until FE-P11 produces
direct row-owned evidence: no visual/product promotion, no
accessibility/product promotion, no Core 05 publication, no benchmark claim, no
fixture-sensitive publication claim, no visual-publication claim, no generated
ledger closure, and no closure from broad checks or retained old artifacts.

FE-P11 must keep the FE-P10 visual and accessibility evidence classes separate.
FE-P10 product-conformance rows do not make FE-P11 readiness rows product
conformance.

## FE-P11 Scope And Non-Scope

FE-P11 scope:

1. Verify deterministic visual fixture mechanics and row accounting for the
   owned-stack Playwright visual suite.
2. Verify the visual fixture matrix, including default Timeline shell coverage
   and the closed FE-VFIX fixture registry.
3. Preserve `FE-VFIX-14` as the exposed `dark_graphite` theme-state fixture
   owned by `FE-V-P11-03`.
4. Verify accessibility coverage for keyboard access, visible focus, System
   views, grid navigation, edit entry/exit, `Esc`, ARIA states, icon-only
   labels, contrast, and non-color-only state communication.
5. Compose readiness gates for frontend typecheck, unit tests, import-boundary
   checks, Biome lint, generated drift, generated-artifact policy, phase
   ledger drift, phase schedule drift, broad check, release check, and a11y.
6. Keep measurement-support evidence separate from Core 05 claim-publication
   evidence unless Core 05 is explicitly activated by a future owner input.

FE-P11 non-scope:

1. New product behavior.
2. New route semantics.
3. New security policy.
4. New theme switcher, light theme, or dedicated high-contrast theme claim.
5. Product-conformance claims from visual evidence.
6. Product-conformance claims from accessibility evidence.
7. Claim-bearing benchmark publication.
8. Visual-publication evidence without Core 05 publication satisfaction.
9. Hand edits to generated ledgers, generated protocol artifacts, generated
   schedules, lockfiles, or tool-managed dependency artifacts.
10. Default Timeline shell closure from `FE-VFIX-14`.

## Live Row Inventory

The live FE-P11 map contains exactly the expected seven rows. If a later live
map differs, record an explicit blocker instead of reconciling silently.

| Row | Layer | Evidence class | Current status | Mapped targets | Current closure and blocker handling |
| --- | --- | --- | --- | --- | --- |
| `FE-V-P11-01` | visual | `implementation_support` | blocked | `browser-e2e-visual` | Blocked by `visual_fixture_not_recaptured_for_frontend_row`. Requires direct visual frontend row accounting for the owned-stack Playwright visual suite. |
| `FE-V-P11-02` | visual | `implementation_support` | blocked | `browser-e2e-visual` | Blocked by `visual_fixture_not_recaptured_for_frontend_row`. Requires direct visual frontend row accounting for the closed fixture matrix. |
| `FE-V-P11-03` | visual | `design_direction` | implemented | `browser-e2e-visual` | Closed as design-direction readiness only at `.cartulary/test-results/20260613T173530Z-p10945/browser-e2e-visual/frontend-row-accounting.json`. |
| `FE-A11Y-P11-01` | accessibility | `design_direction` | blocked | `browser-e2e-a11y-preflight` | Blocked by `frontend_phase_row_not_implemented`. The latest preflight pass has no frontend row-accounting extension and is not closure. |
| `FE-S-P11-01` | support | `implementation_support` | blocked | `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`, `lint-biome`, `generated-artifact-policy-check`, `generate-drift`, `phase-ledger-drift`, `phase-schedule-drift`, `check` | Blocked by `frontend_phase_row_not_implemented`. Requires readiness-composition evidence without promoting target passes to product conformance. |
| `FE-S-P11-02` | support | `implementation_support` | blocked | `check`, `release-check`, `browser-e2e-a11y` | Blocked by `frontend_phase_row_not_implemented`. Requires release-readiness composition without representing blocked fixtures or planned phases as complete. |
| `FE-S-P11-03` | support | `claim_publication_boundary` | blocked | `browser-e2e-measurement`, `browser-e2e-visual`, `benchmark-claim-check` | Blocked by `frontend_phase_row_not_implemented`. Requires measurement-support separation and Core 05 boundary protection. |

`FE-VFIX-14` is current in `tools/frontend_visual_fixture_registry.json`, owned
only by `FE-P11` and `FE-V-P11-03`, with primary golden
`apps/web/e2e/workbook.visual.spec.ts-snapshots/fe-v-p11-03-exposed-theme-states-linux.png`.
It is the exposed theme-state fixture and must not satisfy `FE-VFIX-01`.

## Sprint Plan

### Sprint 1: Baseline And Command Surface

Objective: lock live authority and target surface before implementation.

Actions:

1. Re-run `make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P11`.
2. Re-run `make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P11`.
3. Re-run `make explain-target TARGET=<mapped-target> DETAIL=summary` for each
   mapped FE-P11 target.
4. Re-check frontend guide Section 4.11, FE-P11 map, frontend registry, FE-P11
   generated ledger, visual fixture registry, Core 00 through Core 05, harness
   NLSpec, dev guide, UI/UX guide, design contract, visual golden guide, and
   OpenTelemetry NLSpec for authority boundaries.
5. Record source mismatches, stale digests, missing target artifacts, or row
   inventory drift as blockers.

Exit condition: all FE-P11 owner inputs are current or blockers are explicit.

### Sprint 2: Visual Fixture Discipline

Objective: close or retain visual rows only through `browser-e2e-visual` and
frontend row accounting.

Actions:

1. Verify `FE-VFIX-01` through `FE-VFIX-15` fixture registry shape and status.
2. Prove `FE-VFIX-14` remains only the exposed theme-state support specimen for
   `FE-V-P11-03`.
3. Re-run `make browser-e2e-visual` and inspect
   `browser-e2e-visual/frontend-row-accounting.json`.
4. Keep `FE-V-P11-01` and `FE-V-P11-02` blocked until their exact row-owned
   scenario titles close under current digests.
5. Keep `FE-V-P11-03` design-direction only even when visual target passes.

Exit condition: visual rows have current row-owned evidence or precise blockers.

### Sprint 3: Accessibility Readiness

Objective: make FE-P11 accessibility preflight closure explicit or keep the row
blocked.

Actions:

1. Verify the preflight target emits the FE-P11 row accounting required by the
   live map, or record the missing accounting as a blocker.
2. Cover keyboard access, visible focus, System views, grid navigation,
   edit entry/exit, `Esc`, ARIA states, icon-only labels, contrast, and
   non-color-only states.
3. Run `make browser-e2e-a11y-preflight` directly.
4. Retain `cartulary.frontend_accessibility_preflight_summary.v1` as preflight
   support only unless `cartulary.frontend_row_accounting.v3` closes the row.

Exit condition: `FE-A11Y-P11-01` closes as design-direction readiness only or
stays blocked with a precise missing-evidence reason.

### Sprint 4: Readiness Composition

Objective: prove readiness gates compose without treating blocked rows or
planned phases as complete.

Actions:

1. Run `make frontend-typecheck`, `make frontend-unit`,
   `make frontend-import-boundary-check`, and `make lint-biome`.
2. Run `make generated-artifact-policy-check`, `make generate-drift`,
   `make phase-ledger-drift`, `make phase-schedule-drift`, and
   `make json-shape-check`.
3. Run `make browser-e2e-a11y` for implemented accessibility readiness
   composition and inspect row accounting.
4. Run `make check` after row-owned evidence and drift checks are current.
5. Run `make release-check` only when the release-readiness boundary is being
   claimed.

Exit condition: support readiness rows have current target evidence or explicit
blockers, and no broad target is used as row closure without mapped row-owned
evidence.

### Sprint 5: Core 05 Boundary And Finalization

Objective: keep measurement evidence informative unless Core 05 publication
requirements are separately satisfied.

Actions:

1. Run `make browser-e2e-measurement` for measurement-support evidence when
   needed.
2. Run `make benchmark-claim-check` to verify claim-publication boundary
   handling.
3. Verify visual, responsiveness, retained-run, and measurement artifacts are
   not described as Core 05 claim-bearing evidence.
4. Run `make check` as final broad health after row-owned evidence is current.
5. Run `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when
   retaining a successful full `make check` run. If no retained full-check root
   is used, report that retained-run maintenance was skipped because
   `RESULTS_DIR` was unset.

Exit condition: FE-P11 has either full closure or retained blockers with owner
disposition, final health evidence is recorded, and no non-claim evidence is
promoted beyond its live evidence class.

## Validation Commands

These commands are the prescribed FE-P11 validation sequence. They are not
evidence merely by being listed here.

```sh
make task-guide ROLE=phase-author PHASE_NAMESPACE=frontend PHASE=FE-P11
make explain-phase PHASE_NAMESPACE=frontend PHASE=FE-P11
make browser-e2e-visual
make browser-e2e-a11y-preflight
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make lint-biome
make generated-artifact-policy-check
make json-shape-check
make generate-drift
make phase-ledger-drift
make phase-schedule-drift
make check
make browser-e2e-a11y
make release-check
make browser-e2e-measurement
make benchmark-claim-check
make agent-finalize RESULTS_DIR=<successful-full-check-run-root>
```

`make phase-ledgers` is allowed only after authored map or registry changes
require a Make-owned generated-ledger refresh. It is not a row-evidence
substitute.

## Artifact And Row-Accounting Rules

Closing FE-P11 evidence must use `cartulary.frontend_row_accounting.v3` where
the mapped target requires frontend row accounting. A closing artifact must
match the current guide, registry, FE-P11 map, and generated ledger digests; it
must record the mapped target and command ID, target pass status, row pass or
closed status, exact scenario title when `scenario_title_required=true`, and no
unresolved row blocker.

`FE-V-P11-03` closes only as design-direction readiness. It does not satisfy
`FE-VFIX-01`, does not expose a theme switcher, does not claim a light or
high-contrast theme, and does not activate Core 05 publication authority.

`FE-VFIX-14` must not satisfy `FE-VFIX-01`. The default Timeline shell fixture
requires the compact Timeline first viewport with row gutter, header
affordances, selected row, focused Summary cell, adjacent inspector, status
strip, no admin/demo stack, and only Core 01 default Timeline columns.

Visual screenshots, Playwright traces, accessibility summaries, preflight
summaries, target explanations, generated ledgers, retained artifacts, and broad
`make check` may support diagnosis or handoff. They do not close FE-P11 rows
unless the mapped current row-owned accounting rules above are also satisfied.

## Blocker Register

| Blocker | Owner | Current handling |
| --- | --- | --- |
| `FE-P11-ACTIVATION-BLOCKER-01` | `frontend_phase_owner` | FE-P11 remains planned until direct row evidence and freshness validation are promoted together. |
| `FE-V-P11-01-BLOCKER-01` | `frontend_phase_owner` | Retain until owned-stack visual suite row accounting is recaptured under the closed fixture registry. |
| `FE-V-P11-02-BLOCKER-01` | `frontend_phase_owner` | Retain until visual fixture matrix row accounting is recaptured under the closed fixture registry. |
| `FE-A11Y-P11-01-BLOCKER-01` | `frontend_phase_owner` | Retain until direct accessibility preflight row evidence exists for the FE-P11 row. |
| `FE-S-P11-01-BLOCKER-01` | `frontend_phase_owner` | Retain until readiness-composition evidence is added and promoted in the live map. |
| `FE-S-P11-02-BLOCKER-01` | `frontend_phase_owner` | Retain until release-readiness composition evidence is added and promoted in the live map. |
| `FE-S-P11-03-BLOCKER-01` | `frontend_phase_owner` | Retain until claim-publication-boundary evidence is added and promoted in the live map. |
| `FE-P11-SOURCE-LIMIT-01` | `frontend_phase_owner` | FE-P11 map has no top-level `guide_path`; avoid asserting one. |
| `FE-P11-SOURCE-LIMIT-02` | `frontend_phase_owner` | Latest inspected a11y preflight pass has no frontend row-accounting extension; it is not row closure. |
| `FE-P11-SOURCE-LIMIT-03` | `frontend_phase_owner` | `release-check`, `browser-e2e-measurement`, and `benchmark-claim-check` have no latest artifact and require fresh execution before closure claims. |

Record a new blocker if the live map stops matching the seven-row inventory in
this file, if a generated ledger diverges from the map, if a mapped target is no
longer explainable, if row accounting is stale or v1/v2-only, if exact required
scenario titles are absent, or if any artifact attempts to promote visual,
accessibility, measurement, readiness, or retained evidence into product
conformance.

## Strict Non-Claims

This plan does not claim:

1. FE-P11 activation.
2. Closure for the six blocked FE-P11 rows.
3. New product behavior.
4. Product conformance for any FE-P11 visual evidence.
5. Product conformance for any FE-P11 accessibility evidence.
6. Product conformance for readiness-gate evidence.
7. Core 05 publication readiness.
8. Benchmark readiness.
9. Fixture-sensitive publication readiness.
10. Visual-publication readiness.
11. Light theme support.
12. Dedicated high-contrast theme support.
13. Theme-switcher support.
14. Default Timeline shell closure from `FE-VFIX-14`.
15. Closure from generated ledgers.
16. Closure from fixture registry `current` status alone.
17. Closure from screenshots alone.
18. Closure from retained artifacts alone.
19. Closure from target explanations alone.
20. Closure from broad `make check` alone.
21. Closure from base `phase11` artifacts.
22. Closure from this plan text.

## Completion Criteria

FE-P11 is complete only when all of the following are true:

1. The live FE-P11 map still contains exactly the seven expected rows or any
   difference is recorded as an explicit blocker.
2. All seven live rows are implemented with current row-owned evidence or are
   explicitly blocked with owner disposition.
3. `FE-P11-ACTIVATION-BLOCKER-01` is removed only when the frontend registry,
   map, generated ledger, row evidence, target schedule metadata, evidence
   class owner metadata, and evidence freshness are promoted together.
4. Visual rows close only from `make browser-e2e-visual` row accounting and keep
   visual evidence out of product conformance.
5. Accessibility rows close only from mapped accessibility row accounting and
   keep accessibility evidence out of product conformance.
6. Readiness rows compose their mapped targets without representing blocked
   fixtures or planned phases as complete.
7. `FE-S-P11-03` preserves Core 05 publication-boundary separation.
8. Generated-artifact policy, JSON shape, generation drift, phase-ledger drift,
   and phase-schedule drift checks pass after any owner-input change.
9. Broad `make check` passes after row-owned evidence is current.
10. `make release-check` runs when release readiness is claimed.
11. `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` runs when
    retaining a successful full check run, or retained-run maintenance is
    explicitly reported as skipped because `RESULTS_DIR` was unset.
12. No generated protocol artifact, generated ledger, generated schedule,
    lockfile, or tool-managed dependency artifact is hand-edited.

If any condition is false, FE-P11 remains blocked or incomplete.

## Repository Handoff Notes

There is no declared `FE-P12` in the current frontend registry. Final FE-P11
handoff should go to repository and release readiness rather than a next
frontend phase.

The final handoff must include:

1. Final FE-P11 registry status, row rollup state, activation blockers, and
   digests.
2. Row-by-row FE-P11 status for all seven rows.
3. Direct evidence roots for every closed row.
4. Remaining blockers with reason codes and resolution owners.
5. Validation command results, including drift checks, broad health, release
   readiness when claimed, and Core 05 boundary checks.
6. `agent-finalize` result and retained check root, or an explicit statement
   that retained-run maintenance was skipped because `RESULTS_DIR` was unset.
7. Any Make-owned generated-ledger refresh performed after authored input
   changes.
8. Confirmation that no generated protocol artifacts or generated ledgers were
   hand-edited.
9. Strict non-claims that still apply after FE-P11.
