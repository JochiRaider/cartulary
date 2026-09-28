# Timeline fill accessibility ownership correction

## Decision and boundary

The starting checkout was clean `main` at
`f3e1b8fcbe1d84f34a90bdc196ef0a33d96231ad`. The baseline production
browser scan reproduced a critical `aria-required-children` violation: RDG
7.0.0-beta.59 rendered its fill handle directly under the `grid`, where the
adapter had assigned `role="img"` and `aria-label="Drag to fill this value"`.
The recorded Axe finding included the grid and the offending handle as a
related node. The finding disappeared in edit mode and returned with the
navigation handle. The separate grouped `treegrid` focus element has its own
`aria-required-children` finding; this change does not address it.

Core 03 §§13.2–13.4 govern keyboard ownership, exact fill label, admission,
semantic targets, range behavior, and vendor double-click suppression. Design
§§8.5 and 14 govern grid presentation and accessibility. Core 01 retains
bulk-mutation authority. `docs/domain.md` supplied vocabulary. The source
owner is `packages/grid-adapter`; Timeline remains the admission and mutation
consumer. Verification routing belongs to the current catalog, not to the
source or this handoff. The digest owner/repository maps, rules, acceptance
table, local prompt, and relevant query guidance were used for navigation and
assessment. Research and historical handoffs did not supply authorization.

The adapter now leaves RDG's React-owned pointer handle in place and marks it
`aria-hidden`. Its stable selector and tooltip remain. While the handle is
mounted, the current registered semantic `gridcell` receives
`aria-description="Drag to fill this value"`, appended to any existing
description. The mounted grid, existing active-cell registration, and handle
presence govern reconciliation. Selection, edit, row removal, virtualization,
and grid replacement clean up the old association. The unconditional image
role and label decoration is retired. No new selected-cell state, tab stop,
live announcement, DOM reparenting, API, schema, or persistence model was added.

Changed authored paths:

- `packages/grid-adapter/src/SemanticDataGrid.tsx`: accessible ownership and
  cleanup at the adapter boundary.
- `packages/grid-adapter/src/index.test.tsx`: handle, semantic cell, rerender,
  and removed-row assertions.
- `apps/web/e2e/workbook.a11y.spec.ts`: production Axe and CDP accessibility
  characterization across navigation, edit, Escape, moved selection, and
  grouped unavailability; complete state observations are test attachments.
- `apps/web/e2e/timeline-grid-entry.spec.ts`, `apps/web/e2e/keyboard.spec.ts`,
  `apps/web/e2e/timeline-range-selection.spec.ts`: retained pointer/keyboard
  fill and double-click tests use the cell description; virtualized selection
  checks one described current cell.

`tools/frontend_source_ownership.json`, `tools/frontend_import_boundaries.json`,
`tools/generated_artifact_policy.json`, and the authored test families were
inspected. No routing input or generated output changed. Direct RDG handling
remains in `packages/grid-adapter`; runtime and tests do not read Markdown.

## Evidence

All listed test slices used `CARTULARY_HARNESS_CACHE_MODE=off`; no cached pass
was used for this correction. Run roots are under `.cartulary/test-results/`.

| Command or review | Result and run root |
| --- | --- |
| `make help`; `make help-all`; `make task-guide ROLE=module-author OWNER=package.grid_adapter`; `make task-guide ROLE=module-author OWNER=module.timeline` | Passed; confirmed public routing. |
| `make service-backed-test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.accessibility.verify_grid_cells_editors_group_rows_active_cell_3156bd379d CARTULARY_HARNESS_CACHE_MODE=off` before production edit | Failed as intended on the new grid assertion; `20260928T214554Z-p30870`. Attachments recorded navigation, edit, restored, moved, and grouped ownership. |
| Same accessibility command after correction and final assertion | Passed 11/11; `20260928T220420Z-p77236`. An earlier corrected pass was `20260928T215043Z-p75537`. |
| `make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.index_suite_d804ef9789 CARTULARY_HARNESS_CACHE_MODE=off` | Passed 2/2; final `20260928T221016Z-p27510`. Earlier `20260928T215323Z-p15522` failed on `null.not.toContain`; `20260928T220918Z-p19840` failed when an added unmount check read a detached pre-edit cell. Both test assertions were corrected. |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.spreadsheet_bulk,module.timeline.browser.range_scrolling CARTULARY_HARNESS_CACHE_MODE=off` | Passed 12/12; `20260928T215323Z-p15531`. Existing fill assertions cover explicit pointer and keyboard targets, no double-click or selection mutation, focus/range, and pointer hit geometry; the scrolling test checks virtualized description ownership. |
| `make format` | Passed; final `20260928T221006Z-p23012`. An initial attempt found a root-pinned Axe import declaration issue; the authored test now documents the harness pin and Biome passes. |
| `make agent-finalize` | Passed; `20260928T215459Z-p56469`. `RESULTS_DIR` was unset because no eligible successful full warm check was supplied; retained-run maintenance was skipped. |
| `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome` | Passed; final typecheck `20260928T220918Z-p20006`, import boundary `20260928T220421Z-p77450`, Biome `20260928T220421Z-p77522`. |
| `make lint-markdown` | Passed; `20260928T220918Z-p20014`. |

The isolated seeded UI review used the default profile and run root
`20260928T215535Z-p60599`. At 1440×900, the editor showed a selected fillable
cell and the visual handle, removed it in edit mode, and restored it on Escape;
ArrowDown moved the visible selection to the next cell. At 1024×720, the
selected cell and handle remained visible after horizontal scrolling. Vertical
scrolling rendered later rows without a visual regression. Capture State
grouping removed the fill affordance. The viewer actor showed the
read-only presentation without the handle. Original images and accessibility
observations were inspected before stopping the exact review session. Its
terminal receipt reported `status=ok`, `state=closed`, `cleanup=complete`; only
caller-owned scratch was removed. Private images expired and are not handoff
links.

The editor navigation and edit captures reported no grid
`aria-required-children` finding. They still reported `empty-table-header`
(minor) and `page-has-heading-one` (moderate). The grouped treegrid capture
reported a critical `aria-required-children` finding on its separate focus
element, outside this fill correction. These are not suppressed. Axe also
reported incomplete contrast checks; this slice makes no contrast claim.
No screen-reader session was available (`orca` was not installed), so the
observed accessible ownership and keyboard tests do not establish a
screen-reader navigation result or whole-application conformance.

## Digest acceptance assessment

| Rows | Assessment |
| --- | --- |
| A001–A003 | Pass: Core/design owner mapping, bounded adapter decision, clean baseline, current manifests, and independent test routing recorded above. |
| A011, A014, A016, A019, A021 | Pass for this fill slice: mounted-cell cleanup, edit/Escape, grouped/read-only absence, accessible cell description, and virtualized focus/scroll are covered by fresh slices and rendered review. Other workflow and whole-app accessibility obligations remain outside this slice. |
| A022–A027 | Pass for this slice: inspected production images at both requested viewports; selector contracts retained; no Markdown runtime dependency or generated edit; no migration; this handoff and rollback are explicit. No visual row was needed because paint and geometry did not change. |
| A004–A010, A012–A013, A015, A017–A018, A020 | N/A: no token, theme, density, creation, responsive chrome, overflow ownership, inspector, transaction, recovery, conflict, evidence, or component-variant behavior changed. The narrow 1024×720 inspection is evidence for this handle, not a full responsive acceptance run. |

Compatibility is limited to the existing handle selector, tooltip, pointer
handlers and geometry, keyboard `Ctrl/Cmd+D` route, and source-owned mutation
rules. Rollback reverts the adapter and the five test files together. No data
or API migration is expected. The diff remains uncommitted; no push or deploy
was performed.
