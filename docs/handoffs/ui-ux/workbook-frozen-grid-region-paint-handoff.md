# Frozen grid-region paint correction

## Scope and authority

The execution recheck found clean `main` at
`52da9ed3b93e87efb4a5bb545718546f5f32eaa3`, with the root `AGENTS.md`
applicable. The production owner is `package.grid_adapter`; Timeline and Workbook
own composition and routed browser verification. Core 03 REQ-03-295 and
REQ-03-297 govern freezing and selection, with its keyboard, editing, and
retained-authoring clauses. `docs/design.md` §§8, 12.5, and 14 guide shared
region placement, correction access, focus, and accessibility. `docs/domain.md`
supplies vocabulary and owner navigation. The refactor digest supplied review
routing and acceptance prompts, not a second behavior owner.

The correction changes adapter paint only. It does not change the freeze
boundary, whole-prefix admission, minimum scrollable-area budget, configured
field identity, saved-view layout, widths, geometry helpers, selection store,
record mutation, API, schema, dependency, or virtualization. No migration is
needed.

## Reproduction and confirmed cause

Before the production edit, a seeded `cartulary-ui-review` editor session on
`BROWSER-REVIEW` at 1024×720 compact density reproduced the report. At grid
`scrollLeft = 1100`, the first Activity Date (UTC) cell began at x=29 while the
row-number gutter occupied x=43–101. Timestamps visibly crossed the gutter on
the mounted rows and draft presentation. The baseline browser characterization
also failed a pixel comparison of the gutter with its crossing scrolling cell
visible versus hidden. An independently frozen data cell failed the same
comparison at 1440×900 comfortable density. Baseline failing run roots were
`.cartulary/test-results/20260928T004143Z-p45569` and
`.cartulary/test-results/20260928T004538Z-p15756`.

Computed body, selection, gutter, frozen-data, and draft surfaces inherited
transparent backgrounds in the production renderer. The vendor's frozen body
cell stack was below the adapter's active cell (`z-index: 2`) and editing cell
(`z-index: 4`); an active crossing cell won `elementFromPoint` at the gutter.
The header already painted an opaque surface. The scrolling editor had
`overflow: visible`, so clipping it would threaten correction controls. These
measured surfaces and hit results establish the paint and stack causes. The
vendor's inheritance and editor overflow were investigated as contributors;
the browser measurements, rather than CSS assumptions, determined the fix.

## Correction and changed paths

- `packages/grid-adapter/src/styles.css` gives structural cells, effective
  frozen data cells, and creation-row cells token-backed opaque surfaces. Group
  row frozen cells retain the group surface token; the vendor renders those
  cells without the adapter's ordinary gutter class. Row pending, stale, and
  draft images and cell range/read-only images compose above the opaque surface.
  The selected-row left bar is painted on its frozen selection cell so the
  opaque surface does not cover that cue. Cell borders, outlines, markers, and
  correction overflow remain in place.
- The adapter has one local paint order: scrolling active cells, scrolling
  editors, frozen cells, frozen editors, sticky header/creation cells, then
  their frozen intersections. The old separate active and editing `z-index`
  declarations were retired into that order. No compiler class or second
  geometry model was needed.
- `apps/web/e2e/workbook-column-sizing.spec.ts` adds two focused production
  renderer scenarios using `record_id`, `field_key`, and semantic test IDs.
  They compare pixels with the crossing source shown and hidden, inspect
  computed surfaces, and hit-test frozen targets. They cover Timeline and Hosts,
  1024×720 compact and 1440×900 comfortable, 2× CSS zoom, an active scrolling
  cell/editor, a selection checkbox, frozen data, draft striping, and group
  row surface. The selection action dispatched no record write; its checkbox
  and left bar remain visible. The scrolling editor also retains its draft,
  native focus, and caret while it crosses behind the gutter.
- `apps/web/e2e/keyboard.spec.ts` strengthens the existing virtualized-row
  scenario by scrolling horizontally before revealing the off-screen semantic
  row, then checking its frozen gutter after the vertical reveal.
- The existing frozen-layout browser scenario now moves the editor caret with
  native keys and records its position through Columns, freeze suspension,
  and resumption. It stays at position 4 with the same draft/editor identity.
- `tools/test_families/module.workbook.json` authors the two new browser rows:
  `module.workbook.browser.frozen_region_paint_01360d5734` and
  `module.workbook.browser.frozen_data_region_paint_ea8329bf5c`.
  `tools/browser_e2e_batch_manifest.json` and
  `tools/execution_topology_render_index.json` were regenerated, not hand-edited.

## Browser and product evidence

The two new rows pass at
`.cartulary/test-results/20260928T010501Z-p87257` (11/11 selected units).
The before/after pixel comparisons now match for the gutter, frozen data cell,
zoomed frozen data cell, and Hosts gutter. Browser hit testing lands on the
gutter, frozen data cell, and native selection checkbox. The checkbox toggles
without a record mutation. The grouped frozen cell computes the same surface
as its row and remains sticky.
The structural row passed again with the selected-row bar at
`.cartulary/test-results/20260928T011952Z-p25449` (11/11).
Its final active-editor boundary check passed at
`.cartulary/test-results/20260928T014930Z-p61113` (11/11): the gutter's
pixels and hit target stayed stable while the scrolling editor remained
focused with its draft and caret intact.

The post-correction seeded review session was
`.cartulary/test-results/20260928T010603Z-p20780/ui-review/session.json`,
bundle `bundle-11`. At the original 1024×720 compact/closed-inspector state,
the first date cell again began at x=29 and the gutter occupied x=43–101. The
gutter computed `rgb(17, 19, 24)` from `--ct-colors-surface-1` while the
ordinary scrolling cell remained transparent. The inspected screenshot kept
row numbers and selection controls readable as timestamps crossed behind
them. The session was stopped after inspection; its private links expire.
The review capture disabled axe because it was used for physical paint; the
selected keyboard and correction-access product scenarios provide separate
interaction evidence, not blanket accessibility certification.

The existing `module.workbook` frozen layout, threshold, Find virtualization,
and frozen/scrolling correction-control rows pass at
`.cartulary/test-results/20260928T005913Z-p70891` (14/14). They exercise
whole-prefix and suspension behavior, retained draft/editor identity,
saved-view bytes and clean state, no layout-triggered record writes, Find
reveal, Commit/Cancel/clear access, and a 2× editor case. The selected
`module.timeline` virtualized gutter and range accessibility rows pass at
`.cartulary/test-results/20260928T010025Z-p8797` (13/13). The selected
`package.grid_adapter` placement and mounted-editor reveal rows pass at
`.cartulary/test-results/20260928T005904Z-p68622` (3/3). These targeted rows
cover vertical virtualization after horizontal scrolling, semantic focus and
range continuity, and editor reveal without remounting or clipping controls.
The strengthened virtualization row passed again at
`.cartulary/test-results/20260928T011703Z-p79493` (11/11 selected units).
The native-caret frozen-layout row passed at
`.cartulary/test-results/20260928T013447Z-p77960` (11/11); its attached
samples remained at position 4 through every recorded layout stage.

## Verification and limitations

- `make format` passed after authored source edits; final run root
  `.cartulary/test-results/20260928T014921Z-p56680`.
- `make generate`, `make generate-drift`, `make generated-artifact-policy-check`,
  and `make test-catalog-check` passed. Final drift/policy roots:
  `.cartulary/test-results/20260928T011244Z-p32908` and
  `.cartulary/test-results/20260928T011244Z-p32979`.
- `make agent-finalize` passed before final broader checks at
  `.cartulary/test-results/20260928T015133Z-p94156`. `RESULTS_DIR` was unset
  because there was no qualifying successful full warm check run; retained-run
  maintenance was skipped.
- `make frontend-typecheck`, `make lint-biome`, and
  `make frontend-import-boundary-check` passed at run roots
  `.cartulary/test-results/20260928T015251Z-p98565`,
  `.cartulary/test-results/20260928T015251Z-p98638`, and
  `.cartulary/test-results/20260928T015251Z-p98667`.
- The canonical grid-adapter visual fixture row
  `package.grid_adapter.visual.capture_test_only_grid_adapter_support_specimens_9c222633ba`
  passed at `.cartulary/test-results/20260928T013925Z-p11836` (11/11).
  Its committed golden remains current, so the visual-golden maintenance
  procedure did not require a refresh.
- `make json-shape-check` passed at
  `.cartulary/test-results/20260928T011549Z-p72416`; `make lint-markdown`
  passed at `.cartulary/test-results/20260928T015251Z-p98754`.

Expected red characterization runs above failed on the original product paint.
An initial slice could not route the new rows until `make generate` projected
the authored catalog, and an initial `make agent-finalize` reported that stale
topology; both passed afterward. A grouped-row probe failed at
`.cartulary/test-results/20260928T010217Z-p49186` because its test locator
assumed group cells carried the ordinary gutter class. Vendor inspection showed
they do not; the test and focused group-cell rule were corrected, and the final
selected run passed. No unresolved failure is attributed to this correction.
Programmatic caret probes failed at
`.cartulary/test-results/20260928T012114Z-p62961`,
`.cartulary/test-results/20260928T012328Z-p818`, and
`.cartulary/test-results/20260928T012451Z-p38518`: directly setting the DOM
selection did not model native key movement through the controlled editor. The
scenario now uses native keys; retained caret, text, and editor identity pass.

The narrow rows were chosen for the changed paint, editor, selection, and
virtualization risks. Broad backend and browser suites were not run. The
seeded screenshots are review evidence, not a public accessibility or product
conformance claim.

## Digest acceptance assessment

| Rows | Assessment and evidence |
| --- | --- |
| A001, A003, A024 | PASS: owner clauses, clean baseline, source boundaries, and independent authored test routing were checked. No test or generator depends on Markdown. |
| A004–A006 | PASS: existing dark-graphite surface tokens and shared compact/comfortable density were used; no local palette or density registry was added. |
| A007, A009, A011, A014 | PASS for this slice: creation striping, overflow, saved-view/draft continuity, and correction controls were covered by the selected browser rows. |
| A015, A016, A019, A020 | PASS for affected presentation: existing conflict/pending/stale and range cues retain their CSS layers, while selected range, focus, editor, zoom, and hit-target checks pass. No full accessibility profile is claimed. |
| A021–A025 | PASS: virtualized semantic rows, seeded production captures, semantic selectors, authored routing, and generated policy/drift were checked. |
| A026, A027 | PASS: compatibility, no migration, rollback, and this handoff are recorded. |
| A002, A008, A010, A012, A013, A017, A018 | N/A: no structural boundary, responsive chrome, inspector dispatcher, transaction/replay, authorization, or Evidence lifecycle change. |

## Compatibility and rollback

This is a source-level rendering correction with no stored-data change. To
rollback, revert the focused CSS, the two browser scenarios and strengthened
existing browser assertions, the authored test-family rows, and their
generated topology derivatives. No data rollback or migration step is required.
