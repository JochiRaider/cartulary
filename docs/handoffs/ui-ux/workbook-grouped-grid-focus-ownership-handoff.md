# Grouped Grid Adapter focus ownership correction

## Scope and authority

Implemented on `main` from `ceca6d7c4bedd5ab7b97f910e6e931ca8e360302`.
The working tree was clean at the start of this slice. No unrelated work or
generated input was changed. Core 03 §13.2 owns keyboard/focus departure,
§§13.3–13.4 preserve mutation and range behavior, and §§14.4–14.6 (especially
REQ-03-229–233) own one derived group level, row state, bucket order, and
presentation-only activation. `docs/design.md` §§8.4–8.6 and §14 guide the
visible focus and accessible presentation; `docs/domain.md` supplies vocabulary.
The shared source owner is `packages/grid-adapter`; the independent verification
owners are `package.grid_adapter` and, for retained Timeline range/fill behavior,
`module.timeline`.

## Diagnosis and correction

RDG 7.0.0-beta.59 renders an empty `div[tabindex]` directly under its treegrid.
With no selected group, that focus sink has no row-position style; after group
selection, it receives RDG's group-row focus. The old adapter never consumed
`renderGroupCell.tabIndex`, and the sink was exposed as an invalid treegrid child.
A focused two-bucket browser regression first reproduced the violation. A later
initial viewer regression and seeded review exposed the distinct unselected
case, which the first focused regression did not cover. These are Axe and DOM
observations, not a screen-reader user-impact finding.

`packages/grid-adapter/src/SemanticDataGrid.tsx` now reconciles only its mounted
RDG instance. The empty direct-child sink is removed from the accessibility tree;
when RDG selects a group row, its row index and selected row resolve the mounted
group button, which receives visible focus and a roving Tab stop. The button
retains its label, the row retains `aria-level`/`aria-expanded`, and RDG retains
selection and arrow navigation. Native button Enter/Space activates exactly one
toggle because button clicks no longer propagate to the vendor group-cell click
path. Tab and Shift+Tab depart through the existing grid-region helper. The
observer and focus listeners clean up on unmount. Removed/replaced group focus
falls back to the grid root only while the grid still owns focus; external focus
is left alone. No alternate focus, expansion, query, or selection store was added.

The existing ungrouped fill-handle correction, bucket construction/order,
scoped expansion, draft exclusion, row/extension identity, and all public
semantic contracts remain in place. Superseded repeated group Tab stops and
duplicate click propagation are retired. No schema, API, dependency, or data
migration is required.

## Evidence and limits

Changed authored paths: `packages/grid-adapter/src/SemanticDataGrid.tsx`,
`packages/grid-adapter/src/index.test.tsx`, and
`apps/web/e2e/workbook.a11y.spec.ts`. Unit coverage includes bucket order,
scope, draft exclusion, extension identity, one toggle transition, initial and
selected sink ownership, row replacement, regrouping, outside focus, and unmount.
The production browser row records DOM/AX focus and group relationship after
pointer, Enter, Space, arrows, Tab, and Shift+Tab. It asserts zero relevant Axe
violations before and after selection for editor and viewer, no group writes,
single toggle, read-only viewer access, ungrouped fill, and editor cancellation.

| Public command or review | Current result and run root |
| --- | --- |
| `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789 CARTULARY_HARNESS_CACHE_MODE=off` | Pass 2/2 after formatting, `.cartulary/test-results/20260929T031134Z-p97552`. |
| `make service-backed-test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.accessibility.verify_grid_cells_editors_group_rows_active_cell_3156bd379d CARTULARY_HARNESS_CACHE_MODE=off` | Pass 11/11 after formatting, `.cartulary/test-results/20260929T031146Z-p98126`. |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.spreadsheet_bulk,module.timeline.browser.range_scrolling CARTULARY_HARNESS_CACHE_MODE=off` | Pass 12/12, `.cartulary/test-results/20260929T024656Z-p54490`. |
| `make format` | Pass 2/2, `.cartulary/test-results/20260929T030102Z-p74194`; diff contains only the three intended source/test paths before this handoff. |
| `make agent-finalize` | Pass 1/1, `.cartulary/test-results/20260929T030755Z-p87754`. `RESULTS_DIR` was unset because no eligible successful full warm check was supplied; retained-run maintenance was skipped. |
| `make frontend-typecheck` | Pass 2/2, `.cartulary/test-results/20260929T030821Z-p91995`. |
| `make frontend-import-boundary-check` | Pass 2/2, `.cartulary/test-results/20260929T030821Z-p92042`. |
| `make lint-biome` | Pass 2/2, `.cartulary/test-results/20260929T030821Z-p92119`. |
| `make lint-markdown` | Pass after the final evidence update, `.cartulary/test-results/20260929T031021Z-p95512`. |
| Seeded `make ui-review` / `make ui-review-stop` | Final review `.cartulary/test-results/20260929T030120Z-p78833`: editor Timeline 1440×900 and 1024×720, viewer Timeline, Network Flow graph contributors. Original images and accessibility observations inspected. No `aria-required-children` finding in any final capture. Exact stop receipt says `cleanup: complete`. |

The first seeded review at `.cartulary/test-results/20260929T024820Z-p94072`
found the initial unfocused violation in both Timeline roles and Network Flow.
Its click/Enter focused states were clean. An attempted `select None` UI-review
action timed out (operation 15): the action schema disallows the empty option
value. A fresh snapshot confirmed no selection change. That review session was
stopped with complete cleanup; no result from it was relabeled as final success.
The final full-page review still reports the existing `page-has-heading-one`
finding, outside this grouped-grid slice. No screen-reader executable was
available (`orca`, `nvda`, Narrator, speech-dispatcher, and `espeak` were checked),
so no announcement claim is made. Axe and AX snapshots are not screen-reader
sessions.

Failed focused runs retained during regression development: intentional red
ownership checks `.cartulary/test-results/20260929T022555Z-p50478` and
`.cartulary/test-results/20260929T025746Z-p6619`; characterization/assertion
sequence corrections `.cartulary/test-results/20260929T022413Z-p17118`,
`.cartulary/test-results/20260929T022937Z-p84582`,
`.cartulary/test-results/20260929T023601Z-p85853`, and
`.cartulary/test-results/20260929T023744Z-p19175`; an adapter effect-lifetime
test failure `.cartulary/test-results/20260929T023204Z-p18282`; and a viewer
fixture-label mismatch `.cartulary/test-results/20260929T024352Z-p87308`
(this run also reported fixture cleanup failure). Their later passing runs are
separate evidence, not replacements for those failures.

## Digest acceptance and rollback

For this slice, A001–A003 and A023–A027 pass through the owner/source/routing
review, scoped diff, stable semantic selectors, authored-only changes, and this
handoff. A004–A010, A012–A013, A015, A017–A018, and A020 are outside the changed
behavior. A011, A014, A016, A019, A021, and A022 pass for the applicable grouped
focus, editing cancellation, read-only, virtualization, and reviewed visual
states described above; they do not certify unrelated workbook workflows or
whole-page accessibility. No golden refresh was needed because the intended
visual presentation did not change.

Rollback is a bounded revert of the two Grid Adapter source/test paths, the
browser regression, and this handoff. It requires no data migration. Preserve
the preceding fill-handle correction when reverting.
