# Workbook View-Bar Text-Layout Refactor Handoff

## Workstream tracker

| Workstream | Status | Exit evidence |
| --- | --- | --- |
| WS-01 — Baseline and characterization | DONE | Focused unit and browser fixtures reproduce the overflow-name and 1024px geometry defects. |
| WS-02 — Structural remediation | DONE | Shared allocation slots and the five-family query grid pass the characterized unit and browser cases. |
| WS-03 — Regression and accessibility coverage | DONE | Long-name/value, text-spacing, focus, overflow, dirty-state, and height-invariance coverage passes. |
| WS-04 — Visual reconciliation and final handoff | DONE | Three intentional goldens were manually reviewed; final browser, generation, documentation, and scope audits pass. |

## Baseline

- Date: 2026-08-30.
- Branch: `main`.
- Commit: `1a580e5afca1c6e7f85de88592992773ff7f6c98` (`UI/UX Refactor Digest update`).
- Initial Git status: clean.
- `make doctor`: PASS at `.cartulary/test-results/20260830T224642Z-p3684954`.
- Toolchain pins: Go `1.26.6`, Node `24.15.0`, pnpm `10.33.0`, and ShellCheck `0.11.0`.

### Authorized paths

- `apps/web/src/workbook/components/WorkbookViewBar.tsx`
- `apps/web/src/workbook/components/WorkbookGridControls.tsx`
- `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.tsx`
- `apps/web/src/workbook/components/AssessmentWorkbookSurface.tsx`
- `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookViewBarRegion.tsx`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/WorkbookShell.query.test.tsx`
- `apps/web/e2e/keyboard.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `tools/test_families/web.workbook.json`
- `tools/execution_topology_render_index.json` (generated from the authored
  catalog row)
- The three intentionally changed workbook goldens listed under Visual
  reconciliation
- This dedicated handoff

No product generated root, dependency lockfile, design token, route, schema,
storage, authorization, or persistence path is authorized. The generated
topology render index is the required projection of the authorized test-catalog
row; it does not alter a runtime or product interface.

## Owner clauses

- `docs/design.md` §7.4 owns width-only chrome selection: base at or above
  1280px, narrow desktop at or above 1024px, compact desktop at or above
  768px, and owner-defined degraded behavior below that minimum. Height does
  not select chrome mode.
- `docs/design.md` §7.5 owns view-bar placement and active-chip capacities:
  base 8, narrow desktop 6, compact desktop 0. Compact keeps active query state
  available through Filters. Vertical-only resizing preserves rendered
  location, truncation policy, popover route, and accessible label. The
  overflow accessible name is `Filters, <N> hidden`.
- `docs/design.md` §8.3 owns exact order: saved-view actions, Sort, Group,
  Filters, Columns, active chips or their overflow route, Inspector, and Add
  row when creation is allowed. It also owns the fixed 40px view-bar geometry.
- `docs/design.md` §14, D-AC-024, and D-AC-035 own keyboard, focus, text-layout,
  supported-width, and vertical-resize acceptance.
- Core 03 is consulted only for unchanged query, saved-view, focus-return, and
  dirty-state semantics.
- No adopted-owner contradiction was found.

## Before evidence

- The existing 1440×900 base-desktop golden keeps all required controls
  distinguishable.
- D-VFIX-010 (1024×720 narrow desktop) visibly compresses the Filters button
  to its icon plus the first letter of its label. This is evidence only until
  reproduced through deterministic geometry.
- D-VFIX-011 (768×640 compact desktop) routes all active chips through Filters.
- The existing saved-view/query-control fixture shows saved-view status,
  query controls, active chips, and the right rail competing for the fixed row.
- A cataloged focused unit fixture supplies one group, eight ordered sorts, two
  filters, and an unbroken tag token. It deterministically confirms 8/6/0
  capacity and 3/5/11 hidden counts. Before remediation it fails because the
  accessible overflow name is
  `Filters, 2 active filters, 3 active query chips hidden`, not the owned
  `Filters, 3 hidden`.
- The owner-routed browser fixture supplies a long saved-view name, one group,
  and eight ordered sorts. At 1024×720, computed rectangles reproduce the
  collision: Filters extends to inline coordinate `798.73px` while Columns
  begins at `727.53px`. The controls overlap despite remaining in DOM and
  keyboard order. The same fixture confirms the owned 8/6/0 capacities and
  captures height-paired states for 1440, 1024, and 768 widths.

## Advisory query dispositions

- ADOPT: make flex/grid children explicitly shrinkable; keep long-token
  wrapping inside operable disclosure surfaces; preserve focus return after
  dismissing a popover.
- ADAPT: apply generic truncation and chip-reflow advice to Cartulary's fixed
  one-row geometry, owned 8/6/0 capacities, and existing Filters disclosure
  route.
- REJECT: mobile typography or card conversion, second-row wrapping, toolbar
  scrolling, arbitrary breakpoints, and generic overflow hiding. These conflict
  with R035 or adopted design ownership.
- The narrow UX query produced relevant text-layout and overflow guidance. The
  initial React query produced no material results; a narrow `focus rerender`
  follow-up supplied focus-return guidance. The digest snapshot was not
  mutated.

## Implementation and after evidence

- `WorkbookViewBar` retains its two-rail grid and exact DOM order. Its right
  rail is unshrinkable max-content. Narrow and compact modes allocate a
  `clamp(10rem, 20vw, 18rem)` saved-view slot, a zero-minimum query slot that
  consumes the balance, and a zero-minimum supplemental slot. Base keeps the
  existing unsaved-view projection pixel-stable; a selected base saved view
  adopts the same bounded family allocation where long names and status text
  require it.
- `ActiveSurfaceSavedViewSelector` keeps `View:` visible at base, uses the
  existing compact selector treatment at narrow and compact widths, preserves
  a `6.5rem` operable minimum, and supplies full selected-name and status
  disclosure through the native accessible value, live region, and `title`.
- `WorkbookGridControls` owns one five-family grid: bounded Sort, bounded
  Group, max-content Filters, max-content Columns, and the remaining inline
  space for ordered chips. Immutable `Sort`, `Group:`, `Filters`, and `Columns`
  text cannot be sacrificed to dynamic values.
- Chip buttons share remaining space, preserve the existing block size and a
  `1.8rem` minimum inline extent, retain visible focus, ellipsize only their
  content, and expose their complete names through `aria-label` and `title`.
  Filters reports exactly `Filters, <N> hidden` whenever capacity hides query
  chips. Hidden chips remain operable in the existing Filters dialog.
- Long and unbroken values wrap with `overflow-wrap: anywhere` in Filters,
  Sort, and Columns disclosures; the fixed 40px bar does not wrap.
- The characterized 1024px overlap is removed. The after fixture asserts
  `Filters.scrollWidth <= Filters.clientWidth + 1`, each successive control's
  left edge is at least the preceding control's right edge minus one device
  pixel, every control remains within the view-bar rectangle, and document
  inline overflow remains at most one device pixel. Those relationships pass
  at 1440×900, 1024×720, and 768×640 and at the paired shorter heights.
- The deterministic query fixture is contract-valid: the grouping field is
  one of the eight ordered user sorts, avoiding the existing request rule that
  prepends an absent group field. It therefore exercises one group, eight
  sorts, and two filters without exceeding the public eight-sort limit.

No menu state, query mutation, saved-view dirty state, Escape behavior, focus
restoration, Inspector dispatch, create callback, authorization callback, or
responsive threshold changed.

## Visual reconciliation

- Ordinary validation first failed at
  `.cartulary/test-results/20260831T000707Z-p793614`, as expected for changed
  registered fixtures. Reconciliation reported 49 capture intents, 49 active
  mappings, no missing golden, and no ambiguous mapping. An initial fixture
  mistake that made grouping prepend a ninth sort was rejected during manual
  review rather than accepted as a golden.
- `make browser-e2e-visual-update` passed at
  `.cartulary/test-results/20260831T001242Z-p923759`. Two unrelated one-pixel
  inspector refreshes were restored byte-for-byte. The only retained golden
  changes are:
  - `incident-directory-narrow-desktop-workbook-shell-linux.png`
  - `incident-directory-compact-desktop-workbook-shell-linux.png`
  - `workbook-query-saved-view-query-controls-linux.png`
- Every retained old/current/diff set was reviewed at original resolution.
  Narrow shows six ordered chips and complete Filters text; compact shows zero
  chips and the Filters disclosure route; the saved-view/query fixture shows
  the full base capacity of eight. Inspector and Add row remain visible in all
  three. No clipping, overlap, document scrolling, or query-error banner is
  present.
- Post-review `make browser-e2e-visual` passed at
  `.cartulary/test-results/20260831T001553Z-p968896`. Final validation after
  the selector-policy and containment corrections also passed at
  `.cartulary/test-results/20260831T004105Z-p1379350`.

## Validation log

| Command | Result | Evidence |
| --- | --- | --- |
| `make doctor` | PASS | `.cartulary/test-results/20260830T224642Z-p3684954` |
| `make author-test-row-id FAMILY_ID=web.workbook.regression ...` | PASS | Allocated `web.workbook.regression.workbook_query_controls_preserve_ordered_chip_ca_0fc544e3c5`. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbook_query_controls_preserve_ordered_chip_ca_0fc544e3c5` | EXPECTED FAIL before remediation | `.cartulary/test-results/20260830T230558Z-p3691503`; exact Filters overflow name mismatch. |
| `make service-backed-test-slice OWNER=web.design ROWS=web.design.browser.workbook_responsive_frame_geometry_e78e1d1ac3` | EXPECTED FAIL before remediation | `.cartulary/test-results/20260830T230629Z-p3692405`; 1024px Filters/Columns rectangle overlap. |
| Focused workbook query-control row after remediation | PASS | `.cartulary/test-results/20260830T230938Z-p3735972`; ordered 8/6/0 capacity, 3/5/11 hidden counts, exact overflow names, long token, and Escape focus return. |
| Responsive frame row after remediation | PASS | `.cartulary/test-results/20260830T231422Z-p3864624`; base, narrow, compact, and height-paired computed geometry pass. |
| `make frontend-typecheck` after accessibility coverage | PASS | `.cartulary/test-results/20260830T232736Z-p4165259`. |
| Saved-view/query accessibility row | PASS | `.cartulary/test-results/20260830T233242Z-p142615`; long name, eight sorts, long filter/group values, valid unbroken token, exact accessible names, focus return, dirty state, and WCAG 2.2 text-spacing geometry at 1440/1024/768. |
| Responsive frame row after final allocation tuning | PASS | `.cartulary/test-results/20260830T233337Z-p185207`. |
| Focused workbook query-control row after final allocation tuning | PASS | `.cartulary/test-results/20260830T233431Z-p226606`. |
| Responsive frame row with the final contract-valid eight-sort fixture | PASS | `.cartulary/test-results/20260831T000611Z-p752265`; base/narrow/compact and all height pairs pass. |
| Saved-view/query accessibility row with final allocation | PASS | `.cartulary/test-results/20260831T001048Z-p839725`; WCAG text spacing, full names, focus, and no stale/unavailable query state pass. |
| `make frontend-typecheck` after final allocation | PASS | `.cartulary/test-results/20260831T001034Z-p839281`. |
| Ordinary visual validation | EXPECTED FAIL before refresh | `.cartulary/test-results/20260831T000707Z-p793614`; intentional registered-fixture differences plus one fixture field-list mistake subsequently corrected. |
| Targeted default workbook visual row | Expected narrow-only mismatch before refresh | `.cartulary/test-results/20260831T001142Z-p882351`; unchanged 1440 default capture passed before the intended narrow mismatch. |
| `make browser-e2e-visual-update` | PASS | `.cartulary/test-results/20260831T001242Z-p923759`; unrelated binary refreshes removed. |
| `make browser-e2e-visual` after manual review | PASS | `.cartulary/test-results/20260831T001553Z-p968896`. |
| `make agent-finalize` before broad verification | EXPECTED FAIL, then PASS after regeneration | Initial `.cartulary/test-results/20260831T001840Z-p1014352` identified the stale topology projection created by the new catalog row. `make generate` passed at `.cartulary/test-results/20260831T001915Z-p1015409`; the rerun passed at `.cartulary/test-results/20260831T001930Z-p1018353`. |
| `make task-guide ROLE=module-author OWNER=web.workbook` | PASS | Routed to the focused `web.workbook` owner slice, with `test-fast` as broader coverage. |
| `make task-guide ROLE=module-author OWNER=web.design` | PASS | Routed to the focused and service-backed `web.design` slices plus applicable browser targets. |
| `make test-slice OWNER=web.workbook` | PASS, 139/139 | `.cartulary/test-results/20260831T002159Z-p1052238`. |
| `make test-slice OWNER=web.design` | PASS, 15/15 | `.cartulary/test-results/20260831T002244Z-p1067281`. |
| `make service-backed-test-slice OWNER=web.design` | PASS, 15/15 | `.cartulary/test-results/20260831T002356Z-p1114471`. |
| `make frontend-typecheck` | PASS | `.cartulary/test-results/20260831T002502Z-p1160329`. |
| `make frontend-unit` | EXPECTED FAIL, then PASS, 391/391 | Initial `.cartulary/test-results/20260831T002515Z-p1160854` found a repository selector-policy violation in the new geometry helper. The helper now builds selectors outside `page.evaluate`; the rerun passed at `.cartulary/test-results/20260831T002713Z-p1183860`. |
| `make frontend-import-boundary-check` | PASS | `.cartulary/test-results/20260831T002851Z-p1220873`. |
| `make browser-e2e-webserver-backed` | PASS, 60/60 | `.cartulary/test-results/20260831T002900Z-p1221314`. |
| `make browser-e2e-measurement` | PASS, 22/22 | `.cartulary/test-results/20260831T003508Z-p1278727`. |
| `make browser-e2e-a11y` | PASS, 12/12 | `.cartulary/test-results/20260831T003938Z-p1335355`. |
| Final `make browser-e2e-visual` | PASS, 12/12 | `.cartulary/test-results/20260831T004105Z-p1379350`. |
| `make generate-drift` | PASS, 4/4 | `.cartulary/test-results/20260831T004250Z-p1423811`. |
| `make generated-artifact-policy-check` | PASS, 3/3 | `.cartulary/test-results/20260831T004304Z-p1427159`. |
| Post-format `make frontend-typecheck` | PASS | `.cartulary/test-results/20260831T004540Z-p1438800`. |
| Post-format `make frontend-unit` | PASS, 391/391 | `.cartulary/test-results/20260831T004540Z-p1438809`. |
| Post-format `make browser-e2e-a11y` | PASS, 12/12 | `.cartulary/test-results/20260831T004719Z-p1476240`. |
| Extra `make lint-biome` audit | Warning-only failure outside slice | `.cartulary/test-results/20260831T004525Z-p1438205`; zero errors remain in changed files. The target is nonzero only for eight pre-existing warnings in unchanged `packages/grid-adapter/src/styles.css` (`noImportantStyles` and `noDescendingSpecificity`). |
| Final `make agent-finalize` | PASS | `.cartulary/test-results/20260831T004903Z-p1521584`. |
| `make lint-markdown` | PASS | Final handoff rerun: `.cartulary/test-results/20260831T005047Z-p1525824`. |
| `git diff --check` | PASS | No whitespace errors. |

## Final scope audit

Final Git status contains only the authorized shared workbook implementation,
chrome-mode forwarding, focused unit/browser coverage, three reviewed workbook
goldens, the authored `web.workbook` catalog row and generated topology render
index, and this handoff. There are no changes to generated product roots,
lockfiles, design tokens, routes, APIs, schemas, authorization, storage, or
persistence. No pre-existing user change was present or overwritten.

## Compatibility and rollback

This is a presentation and accessibility resilience refactor with no runtime
API, route, schema, authorization, storage, stored-data migration,
responsive-threshold, density, or generated-interface change.

Rollback is limited to the shared allocation and truncation styles, chrome-mode
prop forwarding, focused coverage, the three intentional goldens, the authored
test-catalog row, and this handoff. No data rollback or migration is required.

## Deferred issues

- A future localization-owned fixture may add representative translated
  strings. This slice covers additional inline consumption through long saved
  names and values, unbroken tokens, and the WCAG 2.2 text-spacing override;
  it does not introduce or change localization infrastructure.
- Below 768px remains the owner-defined degraded desktop behavior. No mobile
  mode, toolbar scroll route, second row, or new breakpoint is introduced.
