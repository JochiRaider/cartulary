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
| FE-P11 map | `16f200de20effdc0655f1f59acb78cead88548f7f3132e5cbd25bf005f29926d` |
| FE-P11 generated ledger | `1baa338199c4ed706fbb4dc6ca5b42652a657283ce06a6719d8e082cd1486313` |
| FE-P11 evidence freshness | `2da37c1ea11ed92c602d963dbd5cd168da97e9de7b03425b52d130aacef09f09` |

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
`5321447c81493dfee058c5a57e77f9aac3f8843f7b3fe5cec0dc56ec9ce388cf`, ledger
digest `83768f981e59a603ef3619c4ab69afc7179aaf3c8978c78d277ead365ee34af2`,
and freshness digest
`21e01658cea3f7f4ed0e231f051c17b174c69c356e56246f140c7eff7a162bfe`.
FE-P10 is dependency context only; FE-P11 needs its own row-owned evidence.

Primary source digests inspected:

| Source | SHA-256 |
| --- | --- |
| `docs/guides/cartulary_frontend_implementation_testing_guide.md` | `e90d682716cd4f65f18c0c8cea43456bff13da2a9726093273fb29ff0f87aea7` |
| `tools/frontend_phase_maps/fe_p11_test_map.json` | `16f200de20effdc0655f1f59acb78cead88548f7f3132e5cbd25bf005f29926d` |
| `tools/frontend_phase_registry.json` | `556abf7923c10fceae87e473ceb1d7e0d0f1f3f33948a3c31b5a1e4815640837` |
| `docs/testing/frontend_phase_coverage_ledgers/fe_p11_coverage_ledger.md` | `1baa338199c4ed706fbb4dc6ca5b42652a657283ce06a6719d8e082cd1486313` |
| `tools/frontend_visual_fixture_registry.json` | `20f466a5bdcf4a9f395c2da5d79c7b87118f7c978d4081b9e22441c9fe8e7e57` |
| `tools/schemas/cartulary.frontend_visual_fixture_registry.v2.schema.json` | `b59506688fe7c6ef20e70ddf88ca95de36035195a8358f1ee1c3f4f534bfdb63` |
| `docs/testing-harness-nlspec.md` | `4303343266385f42e86c077c2ac441d236edb8728d33327af08c30c4e0cc9d06` |
| `docs/opentelemetry-instrumentation-nlspec.md` | `e763ef88ef0420f6c4e1ee1c7bf69733451d4da8475d44347cb1a5c8e06e4451` |
| `docs/design.md` | `e28345fac8ba22fc58264454237af209360a84af0c714ff4e1c94c6028d8cd05` |
| `docs/guides/cartulary-dev-guide.md` | `a4b8fb4b9e3b03c905ed276d19a692559ddf1e70396f224f8f8b2a3f68e58776` |
| `docs/guides/cartulary-ui-ux-design-guide.md` | `3229622b552fed5c15b158d3bd5d7a7e91f99bf4581e40124657b88298a09b26` |
| `docs/guides/cartulary_visual_golden_maintenance.md` | `553122518fccb35ec25b5c1f8f197d8983f52678cdce198c02a745b2a185c836` |
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
| `make phase-ledger-drift` | pass | `.cartulary/test-results/20260613T185132Z-p80825`; generated ledgers matched owner inputs. |
| `make json-shape-check` | pass | `.cartulary/test-results/20260613T185124Z-p80465`; JSON shape and digest mirrors passed. |
| `make phase-schedule-drift` | pass | `.cartulary/test-results/20260613T185132Z-p80830`; generated schedules matched current renderer inputs. |
| `make generated-artifact-policy-check` | pass | `.cartulary/test-results/20260613T185132Z-p80862`; generated-artifact policy passed. |
| `make browser-e2e-visual` | pass | `.cartulary/test-results/20260613T185149Z-p82291`; `FE-V-P11-03` closed as `design_direction`; `FE-V-P11-01` and `FE-V-P11-02` remained blocked in the live map. |
| `make agent-finalize` | pass | `.cartulary/test-results/20260613T185336Z-p93968`; generated outputs unchanged; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
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

Represent live facts still missing after inspection only as explicit blockers
with reason codes and owners.

## Closure Vocabulary

The following terms are normative for this plan. If later text uses a term in
this section, the definition below controls.

| Term | Closed meaning |
| --- | --- |
| `current evidence` | A retained artifact set produced after the latest owner-input digest set named in this plan. Row-accounting artifacts must embed matching guide, registry, and FE-P11 map digests; companion source digests and validation outputs must match the generated ledger, visual fixture registry, and evidence-freshness digests named in this plan. Timestamp recency alone is not current evidence. |
| `fresh execution` | A Make-owned target run from the repository root whose retained run root was created after the latest owner-input change and whose closing artifacts and companion validation outputs match the digest set in this plan. A target rerun with stale embedded digests or stale companion validation is not fresh execution. |
| `row-owned evidence` | A target-owned `cartulary.frontend_row_accounting.v3` artifact for the mapped target whose `target_name`, `command_id`, `phase_namespace`, `accounting_scope`, `row_results[]`, `scenario_results[]`, guide digest, registry digest, and FE-P11 phase-map digest match the live FE-P11 map. Companion digest validation is required by `current evidence`; it is not a replacement for row-owned evidence. |
| `closure` | For a FE-P11 row, closure means the required target has `target_status="pass"`, the row has `row_results[].closure_status="closed"`, `row_results[].failure_reason=""`, the row has no unresolved blocker, the recorded `evidence_class` matches the live FE-P11 map, and the scenario-title rule in the row closure matrix is satisfied. |
| `diagnostic-only` | Evidence retained for investigation or handoff. Diagnostic-only evidence MUST NOT close a FE-P11 row and MUST NOT be cited as product conformance or Core 05 publication evidence. |
| `blocked` | The row remains incomplete for closure and must keep a blocker with a concrete reason code and owner. Blocked rows MAY be absent from target accounting without failing the target when the live map still has `claim_status="blocked"`. |
| `stale` | The artifact has a valid shape but does not match the current digest set or was produced before the latest owner-input change. Stale evidence MAY support diagnosis but MUST NOT close FE-P11 rows. |

The existing interfaces cited by this plan are
`cartulary.frontend_row_accounting.v3`,
`cartulary.frontend_accessibility_preflight_summary.v1`,
`cartulary.frontend_accessibility_summary.v2`,
`cartulary.frontend_phase_test_map.v3`,
`cartulary.frontend_phase_registry.v2`, and
`cartulary.frontend_visual_fixture_registry.v2`.

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

1. Require deterministic visual fixture mechanics and row accounting for the
   owned-stack Playwright visual suite.
2. Require the visual fixture matrix, including default Timeline shell coverage
   and the closed FE-VFIX fixture registry.
3. Preserve `FE-VFIX-14` as the exposed `dark_graphite` theme-state fixture
   owned by `FE-V-P11-03`.
4. Require accessibility coverage for keyboard access, visible focus, System
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

## Row Closure Matrix

The matrix below is the only FE-P11 row-closure interface for this plan. A row
with `claim_status="blocked"` in the live map remains blocked until the live map
is promoted and the matching closure predicates below are satisfied by current
evidence.

| Row | Required target and `command_id` | Required accounting artifact | Required `accounting_scope.mode` | Scenario title rule | Required closure fields | Closure evidence class | Omission behavior | Blocked behavior |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `FE-V-P11-01` | `browser-e2e-visual`; `cartulary.harness.command.browser_e2e_visual.v1` | `<run-root>/browser-e2e-visual/frontend-row-accounting.json` | `active_target` for direct target runs; `selected_rows` only for explicit FE-P11 selected-row execution. | Exact mapped scenario title is required because `scenario_title_required=true`. | Target `target_status="pass"`; row `closure_status="closed"`; row `failure_reason=""`; embedded guide, registry, and map digests match; companion validation satisfies the digest and freshness rules; no unresolved blocker. | `implementation_support` only. | Missing, stale, v1/v2-only, screenshot-only, or fixture-registry-only evidence is non-closure. | Until promoted from blocked, missing row accounting retains `visual_fixture_not_recaptured_for_frontend_row`. |
| `FE-V-P11-02` | `browser-e2e-visual`; `cartulary.harness.command.browser_e2e_visual.v1` | `<run-root>/browser-e2e-visual/frontend-row-accounting.json` | `active_target` for direct target runs; `selected_rows` only for explicit FE-P11 selected-row execution. | Exact mapped scenario title is required because `scenario_title_required=true`. | Same closure fields as `FE-V-P11-01`. | `implementation_support` only. | Fixture registry `current` status, golden filenames, and Playwright screenshots are non-closure without row-owned accounting. | Until promoted from blocked, missing matrix row accounting retains `visual_fixture_not_recaptured_for_frontend_row`. |
| `FE-V-P11-03` | `browser-e2e-visual`; `cartulary.harness.command.browser_e2e_visual.v1` | `<run-root>/browser-e2e-visual/frontend-row-accounting.json` | `active_target` for direct target runs; `selected_rows` only for explicit FE-P11 selected-row execution. | Exact mapped scenario title is required because `scenario_title_required=true`. | Target `target_status="pass"`; row `evidence_class="design_direction"`; row `closure_status="closed"`; row `failure_reason=""`; no unresolved blocker. | `design_direction` only. | Absence of default Timeline shell evidence is not a failure for this row. | This row MUST NOT close `FE-VFIX-01` and MUST NOT expand beyond exposed theme-state readiness. |
| `FE-A11Y-P11-01` | `browser-e2e-a11y-preflight`; `cartulary.harness.command.browser_e2e_a11y_preflight.v1` | `<run-root>/browser-e2e-a11y-preflight/frontend-row-accounting.json` | `active_target` for direct target runs; `selected_rows` only for explicit FE-P11 selected-row execution. | Exact mapped scenario title is required because `scenario_title_required=true`. | Target `target_status="pass"`; row `evidence_class="design_direction"`; row `closure_status="closed"`; row `failure_reason=""`; embedded guide, registry, and map digests match; companion validation satisfies the digest and freshness rules; no unresolved blocker. | `design_direction` only. | `cartulary.frontend_accessibility_preflight_summary.v1` without v3 row accounting is support-only. Raw axe or Playwright accessibility output is non-closure. | Until promoted from blocked, missing preflight row accounting retains `frontend_phase_row_not_implemented`. |
| `FE-S-P11-01` | `frontend-typecheck` / `cartulary.harness.command.frontend_typecheck.v1`; `frontend-unit` / `cartulary.harness.command.frontend_unit.v1`; `frontend-import-boundary-check` / `cartulary.harness.command.frontend_import_boundary_check.v1`; `lint-biome` / `cartulary.harness.command.lint_biome.v1`; `generated-artifact-policy-check` / `cartulary.harness.command.generated_artifact_policy_check.v1`; `generate-drift` / `cartulary.harness.command.generate_drift.v1`; `phase-ledger-drift` / `cartulary.harness.command.phase_ledger_drift.v1`; `phase-schedule-drift` / `cartulary.harness.command.phase_schedule_drift.v1`; `check` / `cartulary.harness.command.check.v1` | `<run-root>/frontend-unit/frontend-row-accounting.json` for the only mapped target with `frontend_row_accounting_required=true`; target summaries for all other listed targets. | `active_target` for direct `frontend-unit`; `selected_rows` only for explicit FE-P11 selected-row execution. | `scenario_titles[]` is intentionally empty; target-level closure permits `closing_scenario_titles=[]`. | Every listed target has `target_status="pass"` in current evidence; the `frontend-unit` row result has `closure_status="closed"` and `failure_reason=""`; embedded guide, registry, and map digests match; companion validation satisfies the digest and freshness rules; no unresolved blocker. | `implementation_support` only. | A pass from `make check` alone is non-closure. Omitted non-row-accounting target summaries keep the support predicate incomplete. | Until promoted from blocked, missing direct implementation evidence retains `frontend_phase_row_not_implemented`. |
| `FE-S-P11-02` | `check` / `cartulary.harness.command.check.v1`; `release-check` / `cartulary.harness.command.release_check.v1`; `browser-e2e-a11y` / `cartulary.harness.command.browser_e2e_a11y.v1` | `<run-root>/browser-e2e-a11y/frontend-row-accounting.json` for the only mapped target with `frontend_row_accounting_required=true`; target summaries for `check` and `release-check`. | `active_target` for direct `browser-e2e-a11y`; `selected_rows` only for explicit FE-P11 selected-row execution. | The mapped scenario title is diagnostic because `scenario_title_required=false`; `closing_scenario_titles=[]` is acceptable unless the live map changes to `scenario_title_required=true`. | `check`, `release-check`, and `browser-e2e-a11y` have `target_status="pass"` when release readiness is claimed; the a11y row result has `closure_status="closed"` and `failure_reason=""`; embedded guide, registry, and map digests match; companion validation satisfies the digest and freshness rules; no unresolved blocker. | `implementation_support` only. | Without fresh `release-check`, release readiness is unclaimed. A `check` pass alone is non-closure. | Until promoted from blocked, missing release-readiness composition retains `frontend_phase_row_not_implemented`. |
| `FE-S-P11-03` | `browser-e2e-measurement` / `cartulary.harness.command.browser_e2e_measurement.v1`; `browser-e2e-visual` / `cartulary.harness.command.browser_e2e_visual.v1`; `benchmark-claim-check` / `cartulary.harness.command.benchmark_claim_check.v1` | `<run-root>/browser-e2e-measurement/frontend-row-accounting.json` and `<run-root>/browser-e2e-visual/frontend-row-accounting.json`; target summary for `benchmark-claim-check`. | `active_target` for direct browser targets; `selected_rows` only for explicit FE-P11 selected-row execution. | The mapped scenario title is diagnostic because `scenario_title_required=false`; `closing_scenario_titles=[]` is acceptable unless the live map changes to `scenario_title_required=true`. | All three targets have `target_status="pass"`; browser row results have `closure_status="closed"` and `failure_reason=""`; embedded guide, registry, and map digests match; companion validation satisfies the digest and freshness rules; no unresolved blocker; `claim_publication_intent="none"` remains recorded. | `claim_publication_boundary` only. | No visual, responsiveness, measurement, retained-run, or benchmark-check artifact may be described as Core 05 publication evidence while intent is `none`. | Until promoted from blocked, missing boundary evidence retains `frontend_phase_row_not_implemented`. |

## Digest And Freshness Rules

Accepted FE-P11 artifact sets MUST match the current guide digest, frontend
registry digest, FE-P11 map digest, generated ledger digest, visual fixture
registry digest, and evidence-freshness digest recorded in this plan. Row-owned
accounting artifacts MUST embed the guide, registry, and FE-P11 map digests.
Digests not carried by the row-accounting schema MUST be satisfied by companion
source digests or validation outputs. If any owner input digest changes, every
prior retained FE-P11 closure artifact becomes stale for closure until rerun
under the new digest set.

`latest artifact` means the latest retained run for the target that also matches
the required digest set and schema version. A newer timestamp with stale digests,
missing v3 row accounting, missing target summary, or failed target status is not
a latest artifact for closure.

The following artifact classes MAY support diagnosis or handoff only unless the
row closure matrix also accepts them as part of row-owned evidence: Playwright
screenshots, Playwright traces, accessibility summaries, preflight summaries,
target explanations, generated ledgers, retained run roots, broad `make check`
summaries, base `phase11` artifacts, and plan text.

## Visual Fixture Matrix Closure

The FE-P11 visual fixture matrix MUST contain exactly one registry entry for
each identifier `FE-VFIX-01` through `FE-VFIX-15`. Duplicate fixture IDs,
unknown fixture IDs, missing fixture IDs, or fixture IDs outside that closed
range are blockers for visual-matrix closure.

Each fixture with `status="current"` MUST declare all fields required by
`cartulary.frontend_visual_fixture_registry.v2`: fixture ID, status, owner phase
IDs, owner row IDs, Playwright scenario title, primary golden filename, all
supporting golden artifacts, seed, viewport, device scale factor, browser zoom,
theme, density, scroll normalization, capture scope, focus state, editor state, inspector state,
dynamic-mask policy, blocked reason, and replacement fixture ID. `current` means
schema-valid fixture metadata plus an owned Playwright scenario and committed
golden. `current` does not close FE-P11 without mapped row-owned accounting.

| Fixture status | Required closure behavior |
| --- | --- |
| `current` | Metadata is schema-valid, the owned Playwright scenario and committed golden exist, and the owning FE-P11 row still closes only through the row closure matrix. |
| `missing` | The owning row MUST remain blocked with a concrete missing-fixture reason. Generic placeholder text is non-closure. |
| `retired` | The fixture MUST name `replacement_fixture_id` or a removal reason before any dependent row can complete. |

`FE-VFIX-14` metadata is reconciled between the visual golden guide and the
live registry. It is a selector-only exposed-theme design fixture with viewport
`1280x720`, `capture_scope.kind="selector"` using
`[data-design-fixture='exposed-theme']`, non-applicable workbook-grid scroll,
`dynamic_masks=[]`, and `no_dynamic_regions=true`. The registry remains the
live source for current FE-P11 fixture metadata, and `FE-VFIX-14` MUST NOT be
used for closure expansion beyond `FE-V-P11-03` design-direction readiness.

## Accessibility Matrix Closure

`FE-A11Y-P11-01` closure requires both the normalized preflight summary and v3
row-owned accounting from `browser-e2e-a11y-preflight`. Raw third-party
accessibility output, browser logs, screenshots, or Playwright traces MUST NOT
substitute for `cartulary.frontend_accessibility_preflight_summary.v1` plus
required `cartulary.frontend_row_accounting.v3`.

| Coverage row | Target | Required summary schema | Row accounting | Expected pass state | Omission behavior |
| --- | --- | --- | --- | --- | --- |
| Keyboard access | `browser-e2e-a11y-preflight` | `cartulary.frontend_accessibility_preflight_summary.v1` | Required v3 row accounting for `FE-A11Y-P11-01`. | Preflight summary has no failed keyboard-access check for the mapped row, and row accounting closes the row. | Missing summary or missing row accounting keeps `FE-A11Y-P11-01` blocked. |
| Visible focus | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | No failed visible-focus check for the mapped row, and row accounting closes the row. | Same as above. |
| System views | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | System-view access/name/focus checks pass for the mapped row. | Same as above. |
| Grid navigation | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Grid navigation checks pass without private selector-only evidence. | Same as above. |
| Edit entry | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Edit-entry checks pass for keyboard operation. | Same as above. |
| Edit exit | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Edit-exit checks pass and do not depend on color alone. | Same as above. |
| `Esc` | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | `Esc` checks pass for the expected interaction priority. | Same as above. |
| ARIA states | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Active tab, surface, menu, group, conflict, save, presence, and evidence states are named or communicated by non-color-only state. | Same as above. |
| Icon-only labels | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Icon-only controls have accessible names. | Same as above. |
| Contrast | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Contrast checks pass for the mapped row. | Same as above. |
| Non-color-only empty/loading/error/blocked states | `browser-e2e-a11y-preflight` | Same as above. | Same as above. | Empty, loading, error, and blocked states are distinguishable without color alone. | Same as above. |

## Support And Readiness Composition

Support rows close target-level predicates only through the target set declared
in the live FE-P11 map. Non-row-accounting target summaries are required support
inputs where listed, but they do not close rows that require row accounting
unless the required row-accounting target also satisfies the row closure matrix.

| Support row | Required targets | Diagnostic or conditional targets | Accounting boundary | Non-closure boundary |
| --- | --- | --- | --- | --- |
| `FE-S-P11-01` | `frontend-typecheck`, `frontend-unit`, `frontend-import-boundary-check`, `lint-biome`, `generated-artifact-policy-check`, `generate-drift`, `phase-ledger-drift`, `phase-schedule-drift`, and `check` MUST pass for closure. | None. | Only `frontend-unit` requires v3 row accounting; other targets require current passing target summaries. | `make check` alone MUST NOT close this row. |
| `FE-S-P11-02` | `check`, `release-check`, and `browser-e2e-a11y` MUST pass when release readiness is claimed. | `release-check` is conditional on a release-readiness claim; without it, release readiness is unclaimed and the row remains incomplete or blocked. | `browser-e2e-a11y` requires v3 row accounting; `check` and `release-check` require current passing target summaries. | A `check` pass without `release-check` and `browser-e2e-a11y` is non-closure. |
| `FE-S-P11-03` | `browser-e2e-measurement`, `browser-e2e-visual`, and `benchmark-claim-check` MUST pass before boundary closure is claimed. | None while `claim_publication_intent="none"`. | Visual and measurement targets require v3 row accounting; `benchmark-claim-check` requires a current passing target summary. | Measurement, visual, responsiveness, and benchmark-check evidence MUST NOT become Core 05 publication evidence while intent is `none`. |

## Core 05 Boundary Rules

The live FE-P11 map sets `claim_publication_intent="none"` for every FE-P11 row.
Therefore measurement, responsiveness, visual, accessibility, readiness,
retained-run, and benchmark-check artifacts are implementation-quality,
design-direction, claim-boundary, or diagnostic evidence only. They MUST NOT be
represented as product conformance, benchmark publication, fixture-sensitive
publication, visual publication, or Core 05 publication readiness.

`benchmark-claim-check` is required before `FE-S-P11-03` boundary closure is
claimed. Its passing target summary validates the no-publication boundary; it
does not publish a claim and does not activate Core 05 by itself.

If the live FE-P11 map, a retained artifact, or a future owner input changes any
FE-P11 row to `claim_publication_intent="claim_bearing_publication"`, FE-P11
completion MUST block until Core 05 publication requirements are satisfied by a
separate owner-approved path. The current plan does not define that publication
path.

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

Exit condition: all FE-P11 owner inputs satisfy the digest and freshness rules
or blockers are explicit.

### Sprint 2: Visual Fixture Discipline

Objective: close or retain visual rows only through the row closure matrix entry
for `browser-e2e-visual` and v3 frontend row accounting.

Actions:

1. Validate that the fixture registry contains exactly `FE-VFIX-01` through
   `FE-VFIX-15` and satisfies the visual fixture matrix closure rules.
2. Enforce `FE-VFIX-14` as the selector-only exposed theme-state support
   specimen for `FE-V-P11-03`.
3. Re-run `make browser-e2e-visual` and evaluate
   `browser-e2e-visual/frontend-row-accounting.json` against the row closure
   matrix.
4. Keep `FE-V-P11-01` and `FE-V-P11-02` blocked until their exact row-owned
   scenario titles close under the digest and freshness rules.
5. Keep `FE-V-P11-03` design-direction only even when visual target passes.

Exit condition: visual rows have current row-owned evidence under the row
closure matrix or blockers with concrete reason codes and owners.

### Sprint 3: Accessibility Readiness

Objective: make FE-P11 accessibility preflight closure explicit or keep the row
blocked.

Actions:

1. Require the preflight target to emit the FE-P11 row accounting required by
   the live map and row closure matrix, or record the missing accounting as a
   blocker.
2. Cover keyboard access, visible focus, System views, grid navigation,
   edit entry/exit, `Esc`, ARIA states, icon-only labels, contrast, and
   non-color-only states.
3. Run `make browser-e2e-a11y-preflight` directly.
4. Retain `cartulary.frontend_accessibility_preflight_summary.v1` as preflight
   support only unless `cartulary.frontend_row_accounting.v3` closes the row.

Exit condition: `FE-A11Y-P11-01` closes as design-direction readiness only or
stays blocked with a concrete missing-evidence reason code and owner.

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
   composition and evaluate row accounting against the row closure matrix.
4. Run `make check` after row-owned evidence and drift checks satisfy the digest
   and freshness rules.
5. Run `make release-check` only when the release-readiness boundary is being
   claimed.

Exit condition: support readiness rows satisfy the support target composition
matrix or retain explicit blockers, and no broad target is used as row closure
without mapped row-owned evidence.

### Sprint 5: Core 05 Boundary And Finalization

Objective: keep measurement evidence informative unless Core 05 publication
requirements are separately satisfied.

Actions:

1. Run `make browser-e2e-measurement` when claiming `FE-S-P11-03` boundary
   closure.
2. Run `make benchmark-claim-check` before claiming `FE-S-P11-03` boundary
   closure.
3. Require visual, responsiveness, retained-run, and measurement artifacts to
   remain outside Core 05 claim-bearing evidence while
   `claim_publication_intent="none"`.
4. Run `make check` as final broad health after row-owned evidence satisfies the
   digest and freshness rules.
5. Run `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` when
   retaining a successful full `make check` run. If no retained full-check root
   is used, report that retained-run maintenance was skipped because
   `RESULTS_DIR` was unset.

Exit condition: FE-P11 reaches `row_complete` or retains blockers with owner
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

Closing FE-P11 evidence MUST satisfy the row closure matrix. Where the mapped
target requires frontend row accounting, the closing artifact MUST be
`cartulary.frontend_row_accounting.v3` at the target-owned
`<target>/frontend-row-accounting.json` path. It MUST record the mapped target,
the mapped command ID, `target_status="pass"`,
`row_results[].closure_status="closed"`, `row_results[].failure_reason=""`, the
required scenario-title behavior, matching digests under the digest and
freshness rules, and no unresolved row blocker.

`FE-V-P11-03` closes only as design-direction readiness. It does not satisfy
`FE-VFIX-01`, does not expose a theme switcher, does not claim a light or
high-contrast theme, and does not activate Core 05 publication authority.

`FE-VFIX-14` must not satisfy `FE-VFIX-01`. The default Timeline shell fixture
requires the compact Timeline first viewport with row gutter, header
affordances, selected row, focused Summary cell, adjacent inspector, status
strip, no admin/demo stack, and only Core 01 default Timeline columns.
The visual golden guide and fixture registry agree that `FE-VFIX-14` uses
viewport `1280x720`, selector capture scope, non-applicable workbook-grid
scroll, and no dynamic regions.

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
scenario titles are absent, if the visual fixture matrix violates the visual
fixture matrix closure rules, or if any artifact attempts to promote visual,
accessibility, measurement, readiness, retained, or benchmark-check evidence
into product conformance or Core 05 publication evidence.

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
23. Closure expansion from `FE-VFIX-14` beyond exposed theme-state
    design-direction readiness.
24. Any Core 05 claim-bearing publication path for FE-P11 while every live row
    has `claim_publication_intent="none"`.

## Completion Criteria

This plan distinguishes document closure from row closure and phase activation.
Do not collapse these states into a single completion claim.

| State | Required predicates | Prohibited interpretation |
| --- | --- | --- |
| `blocked_or_incomplete` | At least one row closure predicate is false, at least one required artifact is stale or missing, or at least one blocker in this plan remains unresolved. | MUST NOT be represented as FE-P11 activation, `active_green`, or closure for blocked rows. |
| `handoff_complete` | This plan has current source digests, explicit blocker handling, closed closure vocabulary, row closure matrix, digest and freshness rules, visual fixture closure rules, accessibility matrix, support composition matrix, Core 05 boundary rules, and strict non-claims. | MUST NOT be represented as row closure by itself. |
| `row_complete` | The live FE-P11 map still contains exactly seven expected rows; every row is closed by current row-owned evidence under the row closure matrix or explicitly blocked with owner disposition; all required support targets pass; visual, accessibility, support, and Core 05 boundary rules are satisfied; and no generated protocol artifact, generated ledger, generated schedule, lockfile, or tool-managed dependency artifact is hand-edited. | MUST NOT be represented as `active_green` while `FE-P11-ACTIVATION-BLOCKER-01` remains in the registry. |
| `active_green` | `row_complete` predicates are true; `tools/frontend_phase_registry.json` records `status="active"` and `row_rollup_state="active_green"` for `FE-P11`; activation blockers are empty; and the frontend registry, FE-P11 map, generated ledger, row evidence, target schedule metadata, evidence-class owner metadata, and evidence freshness are promoted together. | MUST NOT be inferred from target passes, generated ledgers, retained artifacts, or this plan text. |

FE-P11 is `handoff_complete` only when this document satisfies the matrix and
boundary requirements above. FE-P11 is `row_complete` only when all seven live
rows satisfy the row closure matrix or are explicitly blocked with owner
disposition. FE-P11 is `active_green` only when the live registry transition is
performed with the required promotion set.

Generated-artifact policy, JSON shape, generation drift, phase-ledger drift, and
phase-schedule drift checks MUST pass after any owner-input change. Broad
`make check` MUST pass only after row-owned evidence satisfies the digest and
freshness rules. `make release-check` MUST run before release readiness is
claimed. `make agent-finalize RESULTS_DIR=<successful-full-check-run-root>` MUST
run when retaining a successful full check run; otherwise retained-run
maintenance MUST be reported as skipped because `RESULTS_DIR` was unset.

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
