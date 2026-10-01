# Workbook structural accessibility correction

## Scope and checkout

UI-REVIEW-A11Y-01 and UI-REVIEW-A11Y-02 are implemented at the shared Grid Adapter
and Workbook identity presentation boundaries. The full golden refresh and two
ordinary visual validations pass. Changes remain uncommitted; no push, PR, merge
or deployment occurred.

The initial checkout was clean `main` at
`84424f70c169eeb97722180d1dab14c646ea4da4`. The reviewed
`1f23eb98178ea755b2a75538bb6c7d74643dc321` is an ancestor, six commits behind.
Those commits retain readable original-source and selected-reference labels,
status-strip accessibility, readable Timeline selection and dismissed mentions,
and relationship focus reveal. Their source changes remain intact. The Columns
focus-continuity handoff was read and its current routed regressions selected.

The user subsequently authorized updating goldens across the full catalog. That
authorization includes inherited visual differences; it does not make an image
comparison an accessibility or conformance assertion.

## Owners and decisions

- Design §§7.1, 8.1 and 14.1 own this bounded presentation direction. They now
  specify content-backed gutter defaults, grouping precedence, one shared native
  incident heading, truthful unavailable identity and subordinate headings.
  Core identity, interaction, presence and layout obligations are unchanged.
- `package.grid_adapter` owns ordinary gutter normalization. The private helper
  preserves the public ReactNode interface: omitted/null/boolean/blank strings
  become `Row`, while explicit content such as `Source row` survives. Production
  keeps content in the compiled column name; no inherited header renderer can
  override the grouping label. The accessible name derives from rendered content.
- Production and test bindings share the rule. Grouping replaces the structural
  header with its label or field key and adds exactly one structural column even
  without a supplied gutter. Draft/data cell accounting follows the same rule.
- `WorkbookIncidentIdentityDisclosure` owns the native h1 around its existing
  button. Its disclosed region remains a sibling. Margin/font resets preserve
  compact geometry. Existing full button naming, expansion, Escape/focus return
  and outside dismissal remain. TopBar uses the identity error when unavailable;
  accepted identity remains visible during permitted refresh failures.
- Network Analysis and Inspector retain subordinate h2 headings. No producer or
  extension implements a second incident-heading path. No public API, dependency,
  storage, migration, saved-layout, authorization or operation-owner change.

The real shared decisions are the meaning of an otherwise empty structural
header and the heading level of incident identity. They belong at existing shared
presentation boundaries. Both production/test bindings remain necessary supported
consumers. No speculative generic abstraction or compatibility layer was added.
Rollback is a source/test/spec/routing and golden revert together, followed by
generation and ordinary visual validation; no data rollback is needed.

## Changed and inspected paths

Production: `packages/grid-adapter/src/core.ts`, `rdgCompiler.tsx`,
`test-support.tsx`; `apps/web/src/workbook/components/WorkbookIncidentIdentityDisclosure.tsx`
and `WorkbookShellTopBar.tsx`. `SemanticDataGrid.tsx` was inspected; its grouping
name override, freezing and production structural-column behavior remain intact.

Behavioral tests: adapter `index.test.tsx`, Network Flow `NetworkFlowSemanticGrid.test.tsx`,
Workbook `WorkbookShell.surfaces.test.tsx`, new identity disclosure component test,
and `apps/web/e2e/workbook.a11y.spec.ts`. Source guides describe the shared owners.
Tests independently assert content, names, identity changes, loading/error/refresh,
grouped/ungrouped transitions, omitted/empty/explicit labels, disclosure operation
and navigation. They read no Markdown.

Authored verification routing changed in `tools/test_families` for
`package.grid_adapter`, `web.workbook`, `module.workbook`, `web.networkflow` and
`module.networkflow`. `make generate` regenerated the execution topology render
index. Research/domain documents and handoffs were context, never executable inputs.

## Passed verification

Run IDs below are exact directories under `.cartulary/test-results/`. Unit counts
include setup/fixture/finalizer units and are not assertion counts. All test slices
used `CARTULARY_HARNESS_CACHE_MODE=off`.

| Command or selected route | Result and run |
| --- | --- |
| `make generate` | PASS, `20261001T003706Z-p74128` after the final selector update (initial generation: `20260930T235702Z-p41960`). |
| `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789,package.grid_adapter.regression.column_sizing_contract,package.grid_adapter.regression.frozen_column_placement` | PASS 4/4, `20261001T000057Z-p55721`. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.incident_heading_retains_full_identity_and_discl_7c5cefeb7e,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.column_sizing_contract,web.workbook.regression.frozen_column_layout` | PASS 6/6, `20261001T000138Z-p89426`. |
| Final expanded heading/refresh tests: same Workbook slice, first two rows above | PASS 3/3, `20261001T003707Z-p74814`, including the final authored refresh-test selector. Earlier pass `20261001T002307Z-p34412` preceded that selector addition. |
| Network semantic-grid unit rows: `operational_state_policy_94cf786881`, `restores_semantic_focus_5fbb80e46c`, `supports_keyboard_column_9fbd5af078`, owner `web.networkflow` | PASS 4/4, `20261001T000159Z-p14426`. |
| Adapter accessibility row `package.grid_adapter.accessibility.verify_grid_cells_editors_group_rows_active_cell_3156bd379d` via `make service-backed-test-slice` | PASS 11/11, `20261001T004148Z-p93690`. |
| Workbook shell accessibility row `module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159` via `make service-backed-test-slice` | PASS 11/11, `20261001T000401Z-p79644`. |
| Claimed row `module.networkflow.accessibility.verify_claimed_network_analysis_keyboard_focus_a_c81856fdc1` via `make service-backed-test-slice` | PASS 11/11, `20261001T004148Z-p93688`. |
| `make format` | PASS 2/2, `20261001T002249Z-p29670`. |
| `make agent-finalize` | PASS 1/1, `20261001T003821Z-p82739`; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| Final combined Workbook service slice, ten rows described below | PASS 16/16, `20261001T002657Z-p42462`. |
| `make frontend-typecheck` | PASS 2/2, `20261001T002406Z-p40044`. |
| `make lint-biome` | PASS 2/2, `20261001T003909Z-p87323`. |
| `make frontend-import-boundary-check` | PASS 2/2, `20261001T002406Z-p40068`. |
| `make generate-drift` | PASS 4/4, `20261001T003909Z-p87018`. |
| `make generated-artifact-policy-check` | PASS 3/3, `20261001T003909Z-p87051`. |
| `make json-shape-check` | PASS 3/3, `20261001T003909Z-p87043`. |
| `make browser-e2e-visual-update CARTULARY_HARNESS_CACHE_MODE=off` | PASS 12/12, `20261001T002025Z-p90787`; 112 PNGs refreshed, 143 retained. |
| First post-refresh `make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off` | PASS 12/12, `20261001T002656Z-p42334`. |
| Second successful post-refresh ordinary visual run, same command | PASS 12/12, `20261001T003638Z-p44113`; both successful runs reconcile all 255 captures with no mapping errors. |
| `make lint-markdown` | PASS, `20261001T004321Z-p59783` (ad hoc documentation summary). |

The selected Workbook service rows also cover Columns Width Escape/Cancel,
movement, Freeze/Unfreeze, unavailable-Fit fallback, sizing gestures and saved
layouts, constrained viewports, editor/viewer grid entry, supersession, desktop
surface selection and System-view roving entry. Those rows passed in
`20261001T000030Z-p64337`; that run's shell heading assertion failed and was
repaired/retested separately above. The exact selected rows are retained in its
run manifest; a mixed run is not reported as wholly passing.

The final combined command was:

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159,module.workbook.accessibility.column_sizing,module.workbook.browser.column_sizing_gestures_saved,module.workbook.browser.column_sizing_surfaces,module.workbook.browser.column_sizing_viewport,module.workbook.browser.desktop_built_in_surface_selector_keyboard_journ_3475c2c6d2,module.workbook.browser.built_in_grid_entry_editor_59f8176b4a,module.workbook.browser.built_in_grid_entry_viewer_c3bbf9714f,module.workbook.browser.built_in_grid_entry_supersession_5f9a05b19f,module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956 CARTULARY_HARNESS_CACHE_MODE=off
```

The Network semantic-grid unit command was:

```bash
make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.networkflowsemanticgrid_operational_state_policy_94cf786881,web.networkflow.regression.networkflowsemanticgrid_restores_semantic_focus_5fbb80e46c,web.networkflow.regression.networkflowsemanticgrid_supports_keyboard_column_9fbd5af078 CARTULARY_HARNESS_CACHE_MODE=off
```

The browser helper explicitly runs `empty-table-header` and
`page-has-heading-one` on the main document, requires no violations or incomplete
results, and checks the rules' pass list where applicable. It separately asserts
visible meaningful gutter text and one fully named incident h1. Existing ARIA,
keyboard and contrast checks remain. CSS 200% zoom and text spacing are exercised
through the routed fixture; the review harness was not expanded.

## Fresh rendered review

Supported seeded sessions used final production code:

| Profile / run | Inspected states |
| --- | --- |
| `default`, `20261001T000253Z-p23804` | Editor populated Timeline at 1440×900; grouped Capture State at 1024×720/125% with disclosure expanded; empty Timeline/draft at 768×640; System-view Indicators; viewer populated at 1440×900 and grouped at 1024×720. |
| `network_flow_claimed`, `20261001T001508Z-p59745` | Editor/viewer accepted rows at 1440×900; editor disclosure at 1024×720/125%; filtered-empty editor at 768×640; viewer keyboard/grid at 768×640. |

Screenshots and accessibility trees show one full-identity h1, content-backed Row
or Source row, and grouping precedence. Default draft affordances remain; viewer
creation is unavailable. Presence remains semantically named; deterministic row,
cell and overflow presence is additionally covered by adapter and visual fixtures.
Enter/Space, Escape, outside-focus dismissal and surface navigation retain their
existing control owners. Source row can truncate in its unchanged fixed width;
the accessible content remains complete.

The default grid/status strip owns its constrained layout without document
scrolling. Claimed Network Analysis preserves its existing inner horizontal grid
scrolling. At narrow/zoomed heights, its query chrome leaves little visible grid
space; this review does not certify that layout as fully usable at every size.

All consumed capture components had byte length and SHA-256 checked before
inspection. Private artifacts were consumed before exact-session stop. Both
foreground processes exited 0 and terminal receipts say `cleanup=complete`:
default digest `e96ef80ca7060c5e4229727bedeba1d5c489e7887e1a96c2a1e2ef2f901a4142`;
claimed digest `0785d3af3787667bba83a332151c834de21998abd60e00e3f9ff303ebd96d641`.
Default counted two console errors/two failed requests, including initial
unauthenticated preparation; claimed counted two console errors/no failed requests.
These counts are not silently presented as zero-error application runs.

## Failed and incomplete evidence

- Initial ordinary visual run `20261001T000108Z-p59151` failed 10/12 units:
  34 scenario results failed with 112 screenshot errors; no non-screenshot
  assertion error. Claimed visual group passed. Reconciliation accounted for
  255 active captures/goldens and 29 registered fixtures, with zero missing,
  orphaned, ambiguous or unresolved registered mappings.
- Most visual differences are the newly visible Row glyph (81 comparison pixels
  at compact 100%, 264 at 200%, 75 with text spacing). The six contextual
  comparisons documented by the preceding original-source handoff persist:
  desktop Task/Decision add 81 gutter pixels; four narrow comparison counts are
  unchanged. Other retained-authoring/Inspector, dismissed-mention and Evidence
  differences are part of the user's subsequent full-catalog refresh instruction.
- Initial unit/browser assertions confused DOM-unit and Chromium heading naming:
  Chromium includes the child button's full accessible label. The tests now
  assert each engine's actual full identity. Failed runs were
  `20260930T235727Z-p45091`, `20261001T000029Z-p63737`,
  `20261001T000030Z-p64337`, `20261001T000031Z-p64747`; subsequent passes are above.
  Earlier type/format errors in new test code were repaired before final checks.
- A first default session (`20261001T000110Z-p59824`) failed preparation with
  `resource_conflict/capacity_exceeded`; exact stop retained exit 4 and complete
  cleanup. The later fresh sessions supplied the review evidence.
- Canonical image import in default operation 28 returned `unsafe_artifact`.
  That import is incomplete, not a canonical-comparison pass. Exact report-selected
  PNG reference inspection is image-only evidence, separate from reconciliation
  and product browser results; no metadata, limits or provenance was rewritten.
- Broad live Axe scans completed with incomplete `color-contrast`; these are not
  contrast passes. Claimed operation 19 additionally reported
  `scrollable-region-focusable` at 1024×720/125%, while identity disclosure was
  open and the grid was below the visible work area. This separate advisory
  needs Network Analysis layout/focus follow-up. Neither requested rule appeared
  in violations or incomplete results. The targeted browser assertions are the
  positive execution evidence for those two rules.
- Redundant adapter rerun `20261001T003142Z-p44459` was cancelled (10 passed,
  one cancelled graph unit). Ordinary visual retry `20261001T003209Z-p75857`
  was interrupted (10 passed/two cancelled), with a cleanup/artifact-scan error
  for unfinished trace material. Neither run counts as a pass. No trace contents
  were copied or scanner checks bypassed.

## Golden review and cleanup

The [golden inventory](workbook-structural-accessibility-golden-inventory.md)
lists every changed filename, authored owner row and registered fixture. All
112 changed images were imported as exact PNG references and inspected through
supported artifact sessions, using native crops for the Row-only differences
and full images for the additional authoring/Inspector changes and representative
full-page/zoom layouts. Component/crop bytes and SHA-256 were verified before
viewing. Standalone references have no DOM/Axe channels and are not canonical
comparison passes. Product comparison evidence remains the ordinary visual runs.

Artifact sessions `20261001T002328Z-p35742`, `20261001T003200Z-p72323` and
`20261001T003458Z-p36428` each ended `closed`, exit 0, `cleanup=complete`, and
their foreground processes exited. Their terminal digests are respectively
`f4d818bbad7cc5ead023e22c31c1db57bfc3bb2d4a4971f635dec81cb15240ef`,
`ac4e8e1b97b6d1a9f794f6e510df962af60a12df0d86a701f1d28893a35eb0d3`, and
`9235f52d16bad1ef8e25f5862491b9c3cd5d65360f0e500a7ed829973ccff21a`.
An initial crop operation hit capacity while browser jobs were running; after
that contention ended the bounded retry succeeded. No source image was altered
by the review tools, and no screenshot settings changed.

## Unrun and limits

Native screen-reader testing, browser-native zoom, complete WCAG conformance,
release/conformance publication and full backend `make check` are unrun: the
selected change is frontend presentation, and Axe is not evidence for those
claims. Retained successful-full-run maintenance was skipped with no `RESULTS_DIR`.
No new browser driver, harness mechanism, masking or comparison tolerance was added.
All task-owned request/diagnostic scratch was removed after the exact-session
cleanup and image review. Retained structural receipts and ordinary test artifacts
remain under their recorded run roots; expired private review links are not offered
as durable evidence.

## Final qualification and acceptance

The full refresh and both successful ordinary visual runs use the unchanged
renderer/fixture contract. All 112 changed images were inspected, and both
successful reconciliations report all 255 active capture/golden mappings without
errors. No comparison tolerance or screenshot preparation changed.

The digest assessment is restricted to this presentation correction:

| Criteria | Assessment and scope |
| --- | --- |
| A001–A004 | PASS: owning design clauses, private shared decisions, clean baseline reconciliation, source/verification ownership and existing typography tokens. |
| A005–A007 | N/A: no theme/density/creation policy changes; existing variants and draft affordances are exercised as retained behavior. |
| A008–A009 | PASS for unchanged heading/gutter geometry and Workbook shell scroll ownership at the selected widths. Claimed Network Analysis constrained work-area usability remains a separately recorded limitation, not a new guarantee. |
| A010 | N/A: no Inspector routing or operation ownership changes. |
| A011, A014 | PASS: adapter semantic continuity, editor/viewer entry, disclosure focus return and Columns Escape/Cancel/movement/freeze/Fit rows. |
| A012–A013, A015 | N/A: transaction, acknowledgement/recovery and conflict contracts are unchanged. |
| A016 | PASS: empty/populated, draft/read-only, loading/failure and accepted identity refresh remain distinct. |
| A017–A018 | N/A: no authorization, refresh-admission or Evidence lifecycle changes. |
| A019–A020 | PASS for the selected heading/header, keyboard, name and typography/zoom regressions. Incomplete broad contrast and the separate scrollability advisory are excluded from this bounded pass; no full accessibility claim. |
| A021 | N/A: no virtualizer/row-identity algorithm change; existing adapter focus and column-accounting regressions remain selected. |
| A022 | PASS: all changed images reviewed; full update and two fresh ordinary visual validations passed with complete reconciliation. |
| A023–A026 | PASS: semantic selectors, Markdown-independent tests, authored routing/generated projections, unchanged public interfaces and no migration. |
| A027 | PASS: owner decisions, source/routing changes, refresh inventory, exact evidence, cleanup, failures and limits are recorded independently of executable checks. |

UI-REVIEW-A11Y-01 and UI-REVIEW-A11Y-02 are closed for their named defects:
meaningful visible structural headers and one incident heading now exist, the
explicit rules pass without incomplete results, and selected retained behavior
passes. This does not close the separate Network Analysis constrained-scroll
observation or incomplete contrast work. Native assistive-technology compatibility
and overall WCAG conformance remain unclaimed.
