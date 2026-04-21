# React Data Grid Repository Research Report

Repository: `https://github.com/Comcast/react-data-grid`

Inspection baseline: repository default branch `main`, commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, inspected April 21, 2026.

# 1. Orientation and scope

Inspection baseline: I inspected `Comcast/react-data-grid` on the default branch `main` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, commit date `2026-04-17T15:48:10Z`, on April 21, 2026.[^1] The calibration supplied in the prompt is correct against this snapshot: the library source lives in `src`; `src/DataGrid.tsx` is the central orchestration component; the named adjacent components, hooks, utilities, renderer directories, tests, and `website` demo app are present.

[Grounded in configuration] The repository is a React/TypeScript data-grid library package named `react-data-grid`, version `7.0.0-beta.59`. It is distributed as an ESM package with root JS/types exports and a separate CSS export at `react-data-grid/lib/styles.css`.[^2] The package declares `react` and `react-dom` as peer dependencies at `^19.2` and has no declared runtime `"dependencies"` section, so its runtime dependency boundary is effectively React plus the published library bundle and stylesheet.[^2]

[Grounded in code] The public library boundary is `src/index.ts`. That file imports the style layer, exports `DataGrid` and `TreeDataGrid`, exports renderer helpers, exports `Row`, `Cell`, and selection utilities, and re-exports the major public types from `src/types.ts`.[^3] The main implementation boundary is therefore:

- `src`: package implementation and public API.
- `test`: behavioral verification across browser, node, and visual suites.
- `website`: Vite/TanStack Router demo and documentation examples.
- top-level configuration: build, package, TypeScript, Vite/Vitest/Playwright, lint, and formatting.

[Grounded in code] `DataGrid.tsx` owns the grid-level orchestration: prop normalization, column calculation, viewport calculation, active-position/edit state, scroll state, event handling, and final DOM assembly.[^8] Rendering is delegated to row, cell, header, summary, grouping, measuring, and scroll helper components. Data transformations and interaction mechanics are further delegated to hooks under `src/hooks` and utilities under `src/utils`.

# 2. Top-level repository layout

[Grounded in repository structure] The top-level repository contains `.github`, `.vscode`, `src`, `test`, `website`, package metadata, changelog, README, TypeScript configs, Vite/Vitest config, tsdown config, ESLint config, and formatting config.[^1]

## `src`

[Grounded in code] `src` is the library package. It contains the central components and the public export barrel:

- `DataGrid.tsx`: main grid component and imperative handle.
- `TreeDataGrid.tsx`: grouping/treegrid wrapper around `DataGrid`.
- `Row.tsx`, `Cell.tsx`, `EditCell.tsx`: body-row, display-cell, and edit-cell rendering.
- `HeaderRow.tsx`, `HeaderCell.tsx`: regular header rendering, sorting, resizing, and column drag/reorder event handling.
- `GroupedColumnHeaderRow.tsx`, `GroupedColumnHeaderCell.tsx`: nested/grouped column header rendering.
- `SummaryRow.tsx`, `SummaryCell.tsx`: top and bottom summary rows.
- `GroupRow.tsx`, `GroupCell.tsx`: grouped/tree row rendering.
- `Columns.tsx`: predefined selection column and selection cell/header renderers.
- `ScrollToCell.tsx`: scroll coordination sentinel.
- `DataGridDefaultRenderersContext.ts`: context for default renderer overrides.
- `types.ts`: public and internal type contracts.
- `hooks`: calculated columns, viewport rows/columns, active position, dimensions, column widths, row selection context, roving tab index, and stable callback helpers.
- `utils`: active-position navigation, col-span logic, DOM/event/keyboard/style helpers, and measuring-cell rendering.
- `cellRenderers`, `editors`, `style`: reusable renderers/editors and CSS-in-JS style modules.

## `test`

[Grounded in configuration] Vitest is configured with three test projects: browser tests under `test/browser`, visual tests under `test/visual`, and node tests under `test/node`.[^4]

[Grounded in tests] Browser tests cover the behavioral core: keyboard navigation, editing, row selection, copy/paste, drag fill, sorting, events, virtualization, row height, row class/header class, renderers, tree grid behavior, scroll-to-cell, direction/RTL behavior, label behavior, and column behavior such as resizing, colSpan, frozen columns, grouping, and order.[^16][^17][^18][^19][^20][^21][^22] Node tests include SSR rendering, and visual tests take screenshots of representative grid and tree-grid states.[^23]

## `website`

[Grounded in repository structure] `website` is the Vite demo/documentation app. The route listing includes examples for all-features grids, animation, cell navigation, column grouping, column spanning, column reordering, custom renderers, header filters, infinite scrolling, master/detail, million cells, no rows fallback, resizable grids, row grouping, row reordering, scroll-to-cell, tree view, and variable row height. I treated these examples as usage/documentation evidence, not as authoritative implementation behavior, because I did not inspect every route file deeply.

## `.github`

[Grounded in repository structure] `.github` is present in the top-level layout. In the downloaded snapshot I inspected, `.github/workflows` existed but no workflow files were visible from the extracted tree. I therefore do not rely on CI workflow behavior in this report.

## Package and build/test configuration files

[Grounded in configuration] `package.json` defines the package shape, peer dependencies, scripts, export map, published files, and CSS side-effect metadata.[^2] `tsdown.config.ts` defines the library build into `lib`, emits declarations, emits `styles.css`, skips bundling `node_modules`, and uses an `ecij` plugin with a version-derived class prefix to avoid style conflicts between multiple installed RDG versions.[^3] `vite.config.ts` defines the website app and the Vitest browser, visual, and node projects, including custom browser commands for column resizing and drag fill.[^4]

[Grounded in configuration] TypeScript config is split into a base config, source config, and test config. The base config uses modern/bundler-oriented settings such as `moduleResolution: "Bundler"`, `jsx: "react-jsx"`, `isolatedModules`, `exactOptionalPropertyTypes`, `noImplicitReturns`, `noUnusedLocals`, and `target: "ESNext"`.[^5]

# 3. Public API and external behavior surface

## Package exports

[Grounded in configuration] The package root resolves to `./lib/index.js` with `./lib/index.d.ts` types. The CSS entry point is exported as `./lib/styles.css`, and `package.json` marks `*.css` as side effects so bundlers do not accidentally drop the stylesheet.[^2]

[Grounded in code] `src/index.ts` exports:

- `DataGrid` as default and named export.
- `TreeDataGrid`.
- `DataGridDefaultRenderersContext`.
- `Row`, `Cell`, and `Columns` helpers.
- cell renderer helpers from `src/cellRenderers`.
- `renderTextEditor`.
- `renderHeaderCell` and sort-status helpers.
- row-selection hooks.
- public types including `Column`, `CalculatedColumn`, `DataGridProps`, `DataGridHandle`, renderer prop types, event types, `SortColumn`, `ColumnWidths`, `Renderers`, `TreeDataGridProps`, and grouping types.[^3][^7]

## Main `DataGrid` component contract

[Grounded in code] `DataGrid<R, SR, K>` accepts columns, rows, optional summary rows, optional `rowKeyGetter`, callbacks, dimensions, row/header/summary height configuration, selection state, sort state, default column options, event callbacks, clipboard/fill callbacks, virtualization controls, renderer overrides, row/header classes, direction, and data attributes.[^8]

[Grounded in docs] The README advertises the grid as a virtualized React data grid with TypeScript support, keyboard accessibility, frozen columns, column resizing, sorting, column spanning, row grouping, row selection, summary rows, dynamic row heights, no-rows fallback, cell editing, copy/paste, drag fill, custom renderers, and RTL support.[^6]

## Column definitions

[Grounded in code] `Column<R, SR>` is the public column definition type. Important public fields include `key`, `name`, `width`, `minWidth`, `maxWidth`, class names, renderers for cells/header/summary/group/edit cells, `editable`, `colSpan`, `frozen`, `resizable`, `sortable`, `draggable`, `sortDescendingFirst`, and `editorOptions`.[^7]

[Grounded in code] The implementation transforms public columns into `CalculatedColumn` objects through `useCalculatedColumns`. Calculated columns gain internal indexes, parent/group metadata, frozen offsets, resizable/sortable/draggable defaults, measured widths, and calculated style values.[^10]

## Row model

[Grounded in code] Rows are opaque consumer data objects. The grid treats each row as type parameter `R`; it does not require a built-in `id` field. Row identity comes from `rowKeyGetter` when provided, otherwise from the row index for rendering keys.[^8] For row selection, `rowKeyGetter` is mandatory because `selectedRows` is a `ReadonlySet<K>` of row keys, not row objects or indexes.[^8]

[Grounded in docs] The README recommends `rowKeyGetter` for optimal rendering and requires it for row selection.[^6]

## Controlled versus internal props

[Grounded in code] The dominant contract is controlled data with internal interaction state:

| State area | Control model | Implementation evidence |
|---|---:|---|
| `rows` | Consumer-controlled | Grid calls `onRowsChange` with a new rows array; it does not mutate consumer data in place.[^8] |
| `selectedRows` | Consumer-controlled | Selection only operates when both `selectedRows` and `onSelectedRowsChange` are present; callbacks receive a new set.[^8] |
| `sortColumns` | Consumer-controlled | Header sorting computes the next sort model and calls `onSortColumnsChange`; grid does not sort rows itself.[^12] |
| `columnWidths` | Internal or controlled | If both `columnWidths` and `onColumnWidthsChange` are supplied and no resize is active, the prop is treated as the width source; internal state is used during resize and synchronized afterward.[^8][^10] |
| active cell / edit mode | Internal, with notification | `useActivePosition` stores active/edit state; `onActivePositionChange` is notification-only.[^8][^10] |
| scroll position | Internal, with notification | `scrollTop`/`scrollLeft` live in `DataGrid`; `onScroll` is called after state update.[^8] |
| tree expansion | Consumer-controlled in `TreeDataGrid` | `expandedGroupIds` and `onExpandedGroupIdsChange` are passed into `TreeDataGrid`, which computes flattened rows from them.[^13] |

[Grounded in docs] Sorting is explicitly documented as controlled: the grid reports sort changes but does not reorder rows.[^6]

## Public callbacks

[Grounded in code] `DataGridProps` includes callback surfaces for row changes, active-position changes, scroll, cell mouse events, cell keydown, copy, paste, fill, column resize, and column reorder.[^8] Event callbacks receive structured args from `types.ts`; mouse and keyboard events are wrapped so consumers can call `preventGridDefault()` to prevent built-in grid behavior without necessarily preventing the browser’s native event behavior.[^7][^14]

## Renderer extension points

[Grounded in code] Renderer extension is split into column-level renderers and grid-level default renderers.

Column-level renderers:

- `renderCell`
- `renderHeaderCell`
- `renderSummaryCell`
- `renderGroupCell`
- `renderEditCell`[^7]

Grid-level `renderers` object:

- `renderCell`
- `renderCheckbox`
- `renderRow`
- `renderSortStatus`
- `noRowsFallback`[^7]

[Grounded in code] `DataGrid` resolves renderer precedence as local `renderers` prop, then `DataGridDefaultRenderersContext`, then built-in defaults.[^8] Tests verify this local-over-context precedence for no-rows fallback, checkbox renderer, sort-status renderer, cell renderer, and row renderer.[^22]

## CSS/style API

[Grounded in configuration and code] Consumers import `react-data-grid/lib/styles.css`. `src/index.ts` imports `./style/layers.css`, and tsdown emits CSS as `styles.css`.[^2][^3] Styling uses CSS variables, CSS layers, generated class names, and theme classes such as `rdg-light` and `rdg-dark`.[^15]

## `TreeDataGrid`

[Grounded in code] `TreeDataGrid` is exported and implemented as a wrapper around `DataGrid`. It accepts flat raw rows and grouping props, builds grouped rows recursively, flattens expanded groups, maps grouped row selection/copy/paste/row updates back to raw rows, and renders the underlying grid with `role="treegrid"`.[^13]

[Grounded in docs] The README documents `TreeDataGrid` as using the WAI-ARIA treegrid pattern and requiring consumer-provided grouping functions such as `groupBy`, `rowGrouper`, `expandedGroupIds`, and `onExpandedGroupIdsChange`.[^6]

## Public helper components and special columns

[Grounded in code] `Columns.tsx` exports `SelectColumn` and `SELECT_COLUMN_KEY`. `SelectColumn` is a frozen, fixed-width checkbox column with header, row, and group selection renderers.[^13] The package also exports default cell renderers, checkbox renderer helpers, sort-status renderers, `renderHeaderCell`, and `renderTextEditor`.[^3]

# 4. Core architectural units

## `DataGrid.tsx`

[Grounded in code] `DataGrid.tsx` is the central coordinator. It accepts the public props, normalizes defaults, calculates columns and viewport ranges, stores grid interaction state, wires event handlers, computes grid layout styles, and renders headers, rows, summary rows, measuring cells, drag-fill handles, frozen-column affordances, and scroll helpers.[^8]

It owns:

- grid root DOM attributes and refs;
- active position/edit-mode transitions;
- root keyboard handling;
- scroll position;
- column width state;
- drag-fill state;
- column resizing callback coordination;
- row update propagation through `onRowsChange`;
- selection callback coordination;
- renderer precedence.

It delegates:

- column normalization to `useCalculatedColumns`;
- dimensions to `useGridDimensions`;
- visible row/column windows to `useViewportRows` and `useViewportColumns`;
- width measuring/resizing calculations to `useColumnWidths`;
- active-position validation and reset logic to `useActivePosition`;
- row/cell/header/summary/group rendering to component modules;
- keyboard, colSpan, event, DOM, and style helpers to `src/utils`.

## Row rendering

[Grounded in code] `Row.tsx` renders a body row. It receives a column iterator from `DataGrid`, loops over viewport columns, replaces the active display cell with `EditCell` when an editor is active, and otherwise delegates each cell to the default or overridden `renderCell` implementation.[^11] It wraps the row in `RowSelectionContext` so selection cell renderers can use `useRowSelection`.[^11]

## Cell rendering

[Grounded in code] `Cell.tsx` renders a display cell. It computes editable/read-only state, class names, styles, roving tab index, and ARIA attributes, handles mouse events with grid-default suppression, and delegates displayed content to `column.renderCell`.[^11]

## Editing

[Grounded in code] `EditCell.tsx` renders a cell editor container and calls `column.renderEditCell`. It handles outside-click commit, Escape discard, Enter commit, Tab editor navigation, and edit-mode key callbacks.[^11] The editor’s row changes propagate back into `DataGrid` through `onRowChange`; `DataGrid` either updates the draft edit row or commits via `onRowsChange`.[^8]

## Headers

[Grounded in code] `HeaderRow.tsx` loops over viewport columns and renders `HeaderCell`. `HeaderCell` owns header interaction mechanics: sorting, resizing, keyboard resizing, drag/reorder events, ARIA sort attributes, and roving tab index.[^12] Nested header groups are rendered separately through `GroupedColumnHeaderRow` and `GroupedColumnHeaderCell`.[^12]

## Summary rows

[Grounded in code] `SummaryRow.tsx` and `SummaryCell.tsx` render top and bottom summary rows. They use the same viewport column iterator model as body rows, support active-position focus, and call `column.renderSummaryCell` for content.[^11]

## Group/tree units

[Grounded in code] `TreeDataGrid.tsx` constructs grouped rows and adapts them to `DataGrid`. `GroupRow.tsx` renders group rows with treegrid ARIA metadata, and `GroupCell.tsx` renders group cells and group toggles for the matching group-by level.[^13]

## Cell renderer utilities

[Grounded in code] `src/cellRenderers` contains reusable rendering pieces such as checkbox rendering, group toggle rendering, value rendering, and sortable header rendering.[^13] These are used by default column/header/selection behavior and are exported for consumers.

## Hook layer

[Grounded in code] Hooks isolate the high-complexity derived state:

- calculated columns and grouped header metadata;
- grid dimensions;
- viewport row and column windows;
- measured and resized column widths;
- active-position validation/reset;
- row/header selection contexts;
- roving tab index;
- stable callback wrappers.[^10]

## Utility layer

[Grounded in code] Utilities isolate deterministic mechanics: colSpan validation, active-position traversal, keyboard classification, event wrapping, DOM focus/scroll helpers, style generation, and hidden measuring-cell rendering.[^14]

## Style layer

[Grounded in code] `src/style` defines the grid’s generated CSS class names, CSS variables, layers, root/cell/row styling, frozen-column shadows, drag handle styling, selected/active states, and theme classes.[^15]

# 5. Internal data and state model

## Columns and calculated columns

[Grounded in code] Public columns are transformed into `CalculatedColumn` objects. Each calculated column has an internal `idx`, `parent`, `level`, frozen metadata, measured/assigned width, and merged default behavior.[^7][^10]

[Grounded in code] The calculated-column hook also computes:

- `colSpanColumns`: columns that can span multiple columns;
- `lastFrozenColumnIndex`;
- `headerRowsCount`;
- `templateColumns`;
- frozen-column CSS variables;
- total frozen width;
- viewport-column indices.[^10]

## Column groups

[Grounded in code] Column groups are represented by `ColumnGroup` and `ColumnOrColumnGroup`. During calculation, group parent metadata is attached to calculated columns, and grouped header rows are rendered by walking parent chains at each header level.[^7][^10][^12]

## Rows and row indexes

[Grounded in code] Body rows are indexed from `0` through `rows.length - 1` internally. Header rows, top summary rows, and bottom summary rows use internal row-index conventions outside the body range so active-position navigation can move through all row regions.[^8][^14]

## Row keys

[Grounded in code] `rowKeyGetter` provides stable row keys for rendering and selection. Without it, body row render keys fall back to row indexes. Selection operations assert that `rowKeyGetter` exists.[^8]

## Selected rows

[Grounded in code] `selectedRows` is a `ReadonlySet<K>`. The grid never owns an internal selection set. Header and row checkbox interactions create a new `Set` and call `onSelectedRowsChange`.[^8] Selection renderers consume row/header selection context instead of receiving selection state through global mutable state.[^13]

## Sort columns

[Grounded in code] `sortColumns` is a readonly list of `{ columnKey, direction }`. Header cells derive their current sort state from this list and call `onSortColumnsChange` with the next list when clicked or keyboard-activated.[^12] The grid does not reorder `rows`; row ordering remains the consumer’s responsibility.[^6]

## Column widths

[Grounded in code] Column widths are stored as a `Map` keyed by column key with metadata indicating measured or resized widths. `DataGrid` keeps an internal width map, optionally synchronized with the controlled `columnWidths` prop. `useColumnWidths` determines which columns must be measured and calculates grid-template columns.[^8][^10]

## Active position and edit mode

[Grounded in code] Active position is a state object with `idx`, `rowIdx`, and `mode`. `mode` is either active display mode or edit mode. In edit mode the state also carries the draft `row` and `originalRow`.[^10] `DataGrid` uses this state to determine the active cell, active row, active summary/header position, and whether to render `EditCell`.[^8]

## Viewport rows and columns

[Grounded in code] `useViewportRows` calculates which row indexes should be rendered based on scroll position, client height, total row height, row-height mode, and virtualization settings.[^10] `useViewportColumns` calculates which columns to render for each row, always retaining frozen columns and active columns needed for focus/edit continuity.[^10]

## Overscan

[Grounded in code] Row virtualization uses an overscan threshold of `4` when enabled. Column calculation overscans one column around the visible non-frozen column range.[^10]

## Summary rows

[Grounded in code] Top and bottom summary rows are separate arrays. They are rendered outside the body row list, but active-position and colSpan logic treat them as navigable row regions.[^8][^14]

## Frozen columns

[Grounded in code] Frozen columns are calculated with sticky offsets and CSS variables. Frozen columns are always included in the viewport column model, and the grid renders frozen-column shadow elements as scroll affordances.[^8][^10][^15]

## Col spans

[Grounded in code] Column-level `colSpan` can apply to header, row, top summary, and bottom summary cells. The utility accepts only integer spans greater than `1` and ignores spans that cross the frozen/non-frozen boundary.[^14]

## Grouped/tree rows

[Grounded in code] `TreeDataGrid` adds an internal group-row model containing `id`, `parentId`, `groupKey`, `isExpanded`, `childRows`, `level`, `posInSet`, `startRowIndex`, and `setSize`.[^13] It flattens raw rows and group rows into the `rows` array passed to `DataGrid`.

## Drag/fill state

[Grounded in code] `DataGrid` stores drag-fill state as a source column/row plus a dragged-over row. It renders a drag handle only when the active cell is editable, has an edit renderer, is in the viewport, and `onFill` is supplied.[^8]

## Scroll position and dimensions

[Grounded in code] Scroll state is stored as `scrollTop` and absolute `scrollLeft`; RTL left/right handling is normalized by storing `Math.abs(event.currentTarget.scrollLeft)`.[^8] Grid dimensions come from `ResizeObserver` in `useGridDimensions`.[^10]

# 6. Rendering flow

[Grounded in code] The primary rendering path is:

1. A consumer renders `<DataGrid columns={...} rows={...} />`.
2. `DataGrid` destructures props, applies defaults for role, dimensions, row/header/summary heights, direction, virtualized rendering, and default renderers.[^8]
3. `useGridDimensions` observes the root grid element and returns client width/height.[^10]
4. `useCalculatedColumns` merges default column options, flattens grouped columns, sorts select/frozen columns, computes grouped-header levels, frozen offsets, column spans, template columns, and viewport column indices.[^10]
5. `DataGrid` computes row-region counts, ARIA counts, boundary indexes, frozen-shadow styles, and renderer-context values.[^8]
6. `useActivePosition` validates and maintains the active position against current columns, rows, viewport columns, and row changes.[^10]
7. `useViewportRows` computes the rendered row index window and total row height.[^10]
8. `useViewportColumns` computes viewport columns and row-specific iterators that account for frozen columns, active columns, and colSpan coverage.[^10]
9. `useColumnWidths` computes `gridTemplateColumns`, determines which columns need hidden measurement, and provides `handleColumnResize`.[^10]
10. The root `<div>` is rendered with grid role, ARIA row/column counts, direction, class names, scroll handlers, key handlers, CSS variables, and grid-template styles.[^8]
11. Header groups render first through `GroupedColumnHeaderRow`, then the main `HeaderRow` renders regular headers.[^8][^12]
12. If there are no body rows and `noRowsFallback` exists, the fallback is rendered. Otherwise, top summary rows, body rows, and bottom summary rows render in order.[^8]
13. `Row` receives a column iterator from `DataGrid`, loops through visible columns, and renders either `EditCell` for the active edit cell or `Cell` for ordinary display cells.[^11]
14. `Cell` renders ARIA attributes, class names, focus behavior, mouse handlers, and content from `column.renderCell`.[^11]
15. `EditCell` renders the editor container and calls `column.renderEditCell` with draft row data and close/update functions.[^11]
16. Frozen-column shadow elements, drag-fill handle, hidden measuring cells, and `ScrollToCell` render after grid content.[^8]

[Reasoned inference] The rendering architecture is designed so that `DataGrid` performs cross-cutting orchestration once, then passes narrow render contracts to child components. This keeps row/cell/header components mostly presentational plus local interaction handling, while hooks/utilities handle derived layout and navigation state.

# 7. Column handling

## Raw columns to calculated columns

[Grounded in code] `useCalculatedColumns` recursively visits `ColumnOrColumnGroup` entries. Leaf columns become `CalculatedColumn` objects. Groups become parent metadata used later by grouped header rows.[^10]

[Grounded in code] Default column options include default width `"auto"`, minimum width `50`, `renderValue` as the default cell renderer, `renderHeaderCell` as the default header renderer, and false defaults for sorting, resizing, and dragging unless overridden.[^10]

## Select and frozen ordering

[Grounded in code] Calculated columns are sorted so the selection column appears first and frozen columns appear before non-frozen columns.[^10] There is an implementation TODO noting that this sort should keep grouped columns together; that is a fragile area for future grouped/frozen column changes.[^10]

## Widths and measurement

[Grounded in code] Widths can be numbers, strings, or auto-like values. `useColumnWidths` produces CSS grid-template column values and determines which columns must be measured by hidden measuring cells. Resized columns are stored as explicit widths, while measured columns can be recomputed when needed.[^10]

[Grounded in tests] Tests verify non-resizable columns do not resize, drag resizing updates the template and calls `onColumnResize`, min/max widths clamp correctly, keyboard resizing works, double-click auto-resize uses max-content behavior with bounds, flex columns remeasure when needed, and controlled `columnWidths` avoids unwanted callbacks after external prop changes.[^19]

## Resizable behavior

[Grounded in code] `HeaderCell` renders a resize handle when the column is resizable. Pointer movement computes the new width with direction-aware math, clamps to min/max, and calls `handleColumnResize`. Double-click auto-resizes by requesting measurement.[^12] Ctrl+Arrow keyboard resizing is also implemented in `HeaderCell`.[^12]

## Frozen columns

[Grounded in code] Frozen columns receive sticky positioning and frozen offsets from calculated-column style helpers. Frozen columns are always present in viewport column iterators.[^10][^14] `DataGrid` renders separate frozen-column shadow elements whose CSS depends on scroll state and top/bottom summary regions.[^8][^15]

## Column grouping

[Grounded in code] Grouped headers are rendered by `GroupedColumnHeaderRow`, which finds distinct parent groups for the visible columns at a given level, and `GroupedColumnHeaderCell`, which renders group names with colSpan/rowSpan and grid positioning.[^12]

## Column spans

[Grounded in code] `colSpan` can be configured per column and is evaluated through `getColSpan`. Spans apply across header, body, top summary, and bottom summary contexts. Invalid spans and spans crossing the frozen/non-frozen boundary are ignored.[^14] Viewport column iteration adjusts for colSpans so covered cells are skipped and offscreen spanning cells can still be rendered when needed.[^10]

## Sorting

[Grounded in code] Sorting is implemented as header interaction only. `HeaderCell` derives sort direction and priority from `sortColumns`, toggles ASC/DESC/none, supports `sortDescendingFirst`, supports Ctrl/Meta multi-sort, sets `aria-sort` for single-column sorting, and calls `onSortColumnsChange`.[^12]

[Grounded in tests] Sorting tests verify that non-sortable headers do nothing, single-sort cycles through states, multi-sort works with Ctrl/Meta, priority indicators are rendered, `sortDescendingFirst` changes the first direction, and keyboard sorting works.[^20]

## Column reordering

[Grounded in code] Header drag/drop uses `dataTransfer` and calls `onColumnsReorder(sourceKey, targetKey)`. The grid does not internally reorder the `columns` prop.[^12] Consumers must apply the new column order externally and re-render.

# 8. Row handling

## Row identity and keys

[Grounded in code] `DataGrid` keys body rows using `rowKeyGetter(row)` when supplied, otherwise the row index. Selection logic asserts that `rowKeyGetter` exists because selected rows are keyed by `K`.[^8]

[Reasoned inference] Stable row keys are important for preserving row component identity across reorder/filter/update operations. This is explicitly recommended in docs and reinforced by the selection contract.[^6][^8]

## Row indexes

[Grounded in code] Body rows use zero-based row indexes. Summary and header positions are represented using separate internal row index ranges so the active-position state can navigate across headers, body, and summaries.[^8][^14]

## Row heights

[Grounded in code] `rowHeight` can be a number or function. For fixed height, total height and row positions are simple arithmetic. For variable height, `useViewportRows` computes all row heights up front and uses binary search to find the visible row index.[^10]

[Grounded in code] The hook comment explicitly notes that calculating all variable row heights up front can have performance implications and suggests that a future approach similar to `react-window` may be preferable.[^10]

## Row class names

[Grounded in code] `Row` composes generated row classes, even/odd classes, active-row class, row selection class, and consumer-provided `rowClass(row, rowIdx)` output.[^11]

## Row selection state

[Grounded in code] Body rows receive `isRowSelected` and `selectRow` from `DataGrid`. Selection renderers consume these via `RowSelectionContext`.[^11][^13]

## Row renderer overrides

[Grounded in code] `renderers.renderRow` can replace the row renderer. `DataGrid` passes a `RenderRowProps` object that includes row data, row index, viewport-column iterators, active editor, selection state, row class callback, and event/update callbacks.[^7][^8]

[Grounded in tests] Tests verify both context-level and local `renderRow` overrides, with local renderer overriding context renderer.[^22]

## Group and tree rows

[Grounded in code] Group rows are synthetic rows created by `TreeDataGrid`. They render through `GroupRow`, not the ordinary `Row` component. Leaf rows still render through the configured row renderer after `TreeDataGrid` maps flattened indexes back to raw row indexes.[^13]

## Summary rows

[Grounded in code] Top and bottom summary rows render through `SummaryRow`; each summary cell calls `column.renderSummaryCell` if present.[^11] Summary rows are included in keyboard navigation and colSpan handling.[^14]

## No-rows fallback

[Grounded in code] If `rows.length === 0` and a `noRowsFallback` renderer exists, `DataGrid` renders the fallback instead of normal body rows.[^8] Renderer tests verify fallback precedence from prop over context.[^22]

# 9. Cell rendering and cell interaction

## Default display path

[Grounded in code] `Cell` is the default display-cell component. It receives a calculated column, row, row index, colSpan, active/drag state, event callbacks, `setActivePosition`, and `onRowChange`.[^11] It delegates content to `column.renderCell`, passing `column`, `row`, `rowIdx`, `isCellEditable`, `tabIndex`, and `onRowChange`.[^11]

## Custom `renderCell`

[Grounded in code] Consumers can provide `column.renderCell`. Grid-level default `renderers.renderCell` can also replace the cell component wrapper. These are separate concepts: `column.renderCell` renders cell content; `renderers.renderCell` replaces the cell component used by rows.[^7][^8][^11]

## Class names and styles

[Grounded in code] Cell class names combine generated cell class names, frozen styles, drag-over state, editable/read-only state, selection/active state, and consumer `cellClass` output.[^11][^14] Grid layout is expressed through inline grid-row/grid-column styles and generated CSS classes.[^14][^15]

## ARIA and focus

[Grounded in code] Display cells use `role="gridcell"`, `aria-colindex`, optional `aria-colspan`, `aria-selected`, and `aria-readonly` when not editable.[^11] Roving tab index logic gives the active cell or a focusable child `tabIndex=0`, while inactive cells are not tab stops.[^10][^11]

## Mouse events

[Grounded in code] `Cell` wraps mouse-down, click, double-click, and context-menu callbacks. Consumers receive row/column/select-position args and wrapped events. If the consumer calls `preventGridDefault()`, built-in behavior such as selection or edit entry is suppressed.[^11][^14]

[Grounded in tests] Event tests verify that `onCellMouseDown` can prevent selection, `onCellClick` can prevent default behavior while manually opening an editor, double-click prevention stops editor opening, and context-menu args are supplied.[^17]

## Keyboard events

[Grounded in code] Root key handling in `DataGrid` routes active-cell key presses into `onCellKeyDown`, honors `preventGridDefault()`, handles navigation keys, opens editors on default cell input, toggles row selection with Shift+Space, and prevents default browser tabbing when internal grid navigation should occur.[^8][^14]

## Editable-cell detection

[Grounded in code] The actual editable check requires `column.renderEditCell != null` and `column.editable` not false. If `editable` is a function, it must return true for the row.[^14]

[Important discrepancy] `types.ts` contains a comment indicating that setting `editable` without an editor should use a text input. The implementation does not do that: `renderEditCell` is required for editability, and `useCalculatedColumns` does not inject `renderTextEditor` by default.[^7][^10][^14] I treat the implementation and tests as authoritative.

## ColSpan interaction

[Grounded in code] Cell rendering receives `colSpan` from the viewport column iterator. The iterator skips covered columns, renders spanning source cells, and includes active/spanning cells even when virtualization would otherwise omit them.[^10][^14]

# 10. Editing workflow

## Entering edit mode

[Grounded in code] Edit mode can be entered by:

1. Keyboard input recognized by `isDefaultCellInput`, such as printable input, Enter, or F2, when the active cell is editable.[^8][^14]
2. Double-clicking an editable cell unless the double-click handler prevents grid default behavior.[^11]
3. A consumer calling `args.setActivePosition(true)` from a cell mouse event callback.[^11][^17]

[Grounded in tests] Editing tests verify double-click and Enter open editors, typing opens editors, Escape discards, outside clicks commit, and event callbacks can prevent opening or navigation.[^17]

## Selecting the editor

[Grounded in code] The editor renderer is `column.renderEditCell`. `DataGrid` creates an `EditCell` only when the active position is in edit mode, the active row is in or forced into the viewport, and the column has an editor renderer.[^8]

## Editor props

[Grounded in code] `EditCell` calls `column.renderEditCell` with `column`, draft `row`, `rowIdx`, `onRowChange`, `onClose`, and `editorPortalTarget`.[^11] The editor can call `onRowChange(updatedRow)` for draft changes or `onRowChange(updatedRow, true)` for commit.

## Draft update and commit path

[Grounded in code] In edit mode, `DataGrid` stores a draft row in active-position state. Non-committing editor changes replace the draft row. Committing changes call `updateRow`, which creates a new rows array using `rows.with(rowIdx, row)` and calls `onRowsChange(updatedRows, { indexes: [rowIdx], column })`.[^8]

[Grounded in code] If `onRowsChange` is absent or the row object is unchanged, `updateRow` does nothing.[^8] This means editing is operationally controlled by the consumer’s `rows` and `onRowsChange`.

## Commit and cancel actions

[Grounded in code] `EditCell` handles:

- Escape: close editor without commit.
- Enter: close editor with commit.
- outside mouse down: commit by default.
- Tab: if editor navigation says the grid should handle it, prevent default and call grid navigation.
- editor-provided `onClose(commitChanges, shouldFocusCell)`: explicit close path.[^11]

[Grounded in code] `editorOptions.commitOnOutsideClick` defaults to true. `editorOptions.displayCellContent` controls whether the display cell content remains visible while editing. `editorOptions.closeOnExternalRowChange` defaults to true and causes edit mode to close if the external row object changes.[^7][^10][^11]

## Default text editor

[Grounded in code] `renderTextEditor` is a helper editor. It renders an input, autofocuses/selects the value, writes changes by merging `[column.key]: event.target.value` into the row, and commits on blur through `onClose(true, false)`.[^13] It is exported, but not automatically applied to editable columns.

# 11. Selection workflow

## Row-key requirement

[Grounded in code] Selection depends on `rowKeyGetter`. Header select-all and row toggling assert that `rowKeyGetter` exists before operating.[^8] Without a key getter, the grid can render rows but cannot reliably maintain `selectedRows`.

## Controlled selection

[Grounded in code] Selection is controlled. The grid reads `selectedRows` and calls `onSelectedRowsChange(newSet)`. It does not store internal selected state.[^8]

## Header checkbox

[Grounded in code] The header checkbox value is derived by scanning all rows, skipping rows disabled by `isRowSelectionDisabled`, and checking whether all selectable row keys are present, none are present, or the result is indeterminate.[^8] Header toggle creates a new set that selects or deselects all selectable row keys.[^8]

[Grounded in tests] Tests verify select-all, indeterminate behavior, disabled rows being skipped, extra selected keys being preserved, and the header not being checked when no real rows are selected.[^16]

## Single-row and range selection

[Grounded in code] `selectRow` toggles an individual row key. When shift selection is requested and a previous row index exists, it toggles the range between the previous row and current row, skipping disabled rows.[^8]

[Grounded in tests] Tests verify checkbox toggling, keyboard toggling, shift-click range selection, and Shift+Space selection from the active cell.[^16]

## Selection renderers

[Grounded in code] `SelectColumn` uses `SelectCellFormatter` for body cells, a header renderer for select-all, and a group renderer for group selection. These formatters consume `useRowSelection` and `useHeaderRowSelection` contexts.[^13]

## Tree selection

[Grounded in code] `TreeDataGrid` maps raw selected row keys into a derived set that can include group-row IDs when all child rows are selected. When selection changes on a group row, it maps the group operation back to raw child row keys before calling the consumer’s `onSelectedRowsChange`.[^13]

# 12. Keyboard, focus, and accessibility behavior

## Active position

[Grounded in code] Active position is the single source of truth for keyboard focus inside the grid. It tracks column index, row-region index, and active/edit mode.[^10] `DataGrid` uses it to set active cell/row classes, determine tab stops, render editors, and scroll active positions into view.[^8]

## Arrow, Tab, Home/End, and Page navigation

[Grounded in code] `DataGrid` handles Arrow keys, Tab/Shift+Tab, Home/End, and PageUp/PageDown through `getNextPosition`, `getNextActivePosition`, and `navigate`.[^8][^14] Tab can exit the grid at the first or last navigable cell; otherwise it moves internally and prevents default browser tabbing.[^8][^14]

[Grounded in tests] Keyboard tests verify no initial active cell, Tab entry, arrows, PageUp/PageDown, Home/End, Ctrl+Home/Ctrl+End, row transitions, grid exit to focusable elements before/after the grid, active-cell restoration, focusable cell/header/summary renderers, active cell outside viewport handling, active reset when columns/rows are removed, and RTL arrow swapping.[^18]

## RTL-sensitive navigation

[Grounded in code] `keyboardUtils` maps ArrowLeft and ArrowRight according to direction. `DataGrid` also stores absolute scroll-left values to normalize RTL scrolling.[^8][^14]

## Editor versus grid navigation

[Grounded in code] `EditCell` first calls `onCellKeyDown` with edit-mode args and honors `preventGridDefault()`. Escape/Enter close the editor. Tab is delegated to `onEditorNavigation`, which decides whether the editor should keep focus or grid navigation should proceed.[^11][^14]

## Roving tab index

[Grounded in code] `useRovingTabIndex` supports the WAI-ARIA pattern where the active cell is the only tab stop unless it contains a focusable child. If the active cell contains a child with `tabIndex=0`, the hook focuses that child and gives the cell `tabIndex=-1`.[^10]

## ARIA

[Grounded in code] The root grid uses `role="grid"` by default, with ARIA row/column counts and data attributes. Header cells use `role="columnheader"`, body/summary cells use `role="gridcell"`, and rows use `role="row"`.[^8][^11][^12] Tree group rows add `aria-level`, `aria-setsize`, `aria-posinset`, and `aria-expanded`.[^13]

## Treegrid keyboard behavior

[Grounded in code] `TreeDataGrid` renders `DataGrid` with `role="treegrid"` and implements group-row keyboard behavior. On group rows, left/right collapse/expand groups according to direction, and collapsing can move focus to a parent group row.[^13]

[Grounded in tests] Tree grid tests verify row/cell focus movement, header/summary navigation, group expand/collapse through mouse and keyboard, and treegrid ARIA attributes.[^21]

# 13. Scrolling and virtualization

## Scroll state

[Grounded in code] `DataGrid` stores `scrollTop` and normalized absolute `scrollLeft`. The scroll handler uses `flushSync` to update state before invoking `onScroll`, reducing visible desynchronization between scroll state and rendered viewport.[^8]

## Grid dimensions

[Grounded in code] `useGridDimensions` uses `ResizeObserver` to track client width and height. It initializes to `1x1` and returns early when `ResizeObserver` is unavailable, supporting SSR/no-DOM execution.[^10]

## Viewport rows

[Grounded in code] `useViewportRows` calculates visible row indexes from scroll position and height. Fixed row height uses arithmetic; variable row height computes row-top positions and uses binary search.[^10] Virtualization can be disabled, in which case all rows are rendered.[^10]

## Viewport columns

[Grounded in code] Column virtualization is computed from column widths, scroll-left, viewport width, frozen width, and total columns. Frozen columns are always included, and one non-frozen column on each side is overscanned.[^10]

## Active cell preservation

[Grounded in code] `useViewportColumns` and `DataGrid` force active rows/columns into the render set when needed so active focus/edit state can survive virtualization boundaries.[^8][^10]

## Grid templates

[Grounded in code] `DataGrid` assembles `gridTemplateRows` from header rows, top summary rows, body row slots, bottom summary rows, and measured total body height. `useColumnWidths` supplies `gridTemplateColumns`.[^8][^10]

## Scroll-to-cell

[Grounded in code] The public imperative handle exposes `scrollToCell({ idx, rowIdx })`. Valid calls set a scroll target, and `ScrollToCell` renders a sentinel div at the target grid position and repeatedly calls `scrollIntoView` until scroll position stabilizes.[^8][^11]

[Grounded in tests] Tests cover virtualization behavior across rows and columns, frozen columns staying rendered, all frozen columns, summary rows, zero rows/columns, disabled virtualization, and colSpan scroll-to-cell cases.[^19]

[Grounded in tests] One virtualization test for disabled virtualization with some frozen columns is skipped as failing. That marks a fragile area if changing virtualization/frozen-column interactions.[^19]

# 14. Copy, paste, drag, and fill behavior

## Copy

[Grounded in code] Copy operates on the active cell if it is a body cell in the viewport. `DataGrid` calls `onCellCopy({ row, column }, event)` when available.[^8]

[Grounded in tests] Copy/paste tests verify copy callbacks receive row and column, and that copy/paste does not occur for headers or summary rows.[^20]

## Paste

[Grounded in code] Paste requires an active editable body cell, `onCellPaste`, and `onRowsChange`. The paste callback returns an updated row, and `DataGrid` commits that row through the same `updateRow` path used by editing.[^8]

[Grounded in tests] Tests verify paste updates rows and that paste is disabled on read-only cells while copy remains allowed.[^20]

## Fill handle visibility

[Grounded in code] The fill handle is rendered only when `onFill` exists, the active cell is in display mode, the active column is rendered, the column has `renderEditCell`, and the cell is editable.[^8]

## Drag-fill data path

[Grounded in code] Pointer down on the drag handle stores the source cell. Pointer movement tracks the dragged-over row. Lost pointer capture triggers `updateRows`, which iterates target rows, checks editability, calls `onFill({ columnKey, sourceRow, targetRow })`, collects changed rows/indexes, and calls `onRowsChange` once if any rows changed.[^8]

[Grounded in tests] Drag-fill tests verify the handle is hidden without `onFill`, double-click fills downward while skipping read-only cells, dragging down/up updates rows, and handle clicks refocus the active cell.[^20]

## Tree-grid interaction

[Grounded in code] `TreeDataGridProps` omits `onFill`, so fill is not exposed through the tree wrapper. Tree copy/paste explicitly prevents operations on group rows and delegates leaf operations to the raw callbacks.[^13]

# 15. Grouping and tree behavior

## `TreeDataGrid` as wrapper

[Grounded in code] `TreeDataGrid` is not a separate rendering engine. It computes grouped data and adapts props before rendering `DataGrid` with `role="treegrid"`.[^13]

## Group construction

[Grounded in code] `TreeDataGrid` sorts columns so grouped columns appear after the select column and are frozen. It recursively groups rows using `rowGrouper`, builds group-row objects with hierarchy metadata, and flattens groups according to `expandedGroupIds`.[^13]

## Expansion and collapse

[Grounded in code] Group rows are expanded when their group ID appears in `expandedGroupIds`. `toggleGroup` creates a new set and calls `onExpandedGroupIdsChange`.[^13] Group IDs default to a parent-based path plus group key, but `groupIdGetter` can override this.[^13]

## Rendering differences from plain grid

[Grounded in code] Flattened group rows render through `GroupRow`, while leaf rows render through the configured raw row renderer. Group rows use active position `idx = -1` to support row-level focus in the treegrid pattern.[^13]

## Accessibility differences

[Grounded in code] Tree group rows include `aria-level`, `aria-setsize`, `aria-posinset`, and `aria-expanded`. `TreeDataGrid` sets root role to `treegrid` and uses a tree-specific keydown adapter for left/right expand/collapse behavior.[^13]

## Behavioral evidence

[Grounded in tests] Tree tests verify no grouping when `groupBy` is empty or nonexistent, single and multiple grouping, custom group IDs, duplicate `groupBy` handling, click and keyboard toggles, ARIA attributes, group selection semantics, treegrid navigation, copy/paste restrictions on group rows, row updates through cell renderers, custom group cell rendering, and summary-row focus behavior.[^21]

# 16. Utility and hook layers

## Hooks

[Grounded in code] `useCalculatedColumns` is the column compiler. It applies defaults, flattens groups, computes parent/header metadata, orders select/frozen columns, computes frozen offsets and template columns, and identifies colSpan-capable columns.[^10]

[Grounded in code] `useViewportRows` is the row virtualization engine. It computes row positions, total body height, and the rendered row-index window.[^10]

[Grounded in code] `useViewportColumns` is the column virtualization engine. It combines visible columns, frozen columns, active columns, and colSpan correction into iterators consumed by rows and summaries.[^10]

[Grounded in code] `useColumnWidths` is the width and measurement engine. It determines which columns need measurement, reads measuring-cell dimensions, updates the column width map, and handles resize operations.[^10]

[Grounded in code] `useGridDimensions` abstracts root element measurement through `ResizeObserver`.[^10]

[Grounded in code] `useActivePosition` owns active/edit state validation, active row/column lookup, and reset/close behavior when rows or columns change.[^10]

[Grounded in code] `useRowSelection` and `useHeaderRowSelection` expose selection contexts to selection-cell renderers and throw if used outside the appropriate render context.[^10]

[Grounded in code] `useRovingTabIndex` implements active-cell and focusable-child tab-stop behavior.[^10]

[Grounded in code] `useLatestFunc` returns a stable callback that always calls the latest function reference, preventing memoized children from rerendering due solely to callback identity changes.[^10]

## Utilities

[Grounded in code] Active-position utilities implement editability checks, colSpan lookup per row region, next-position traversal, and grid-exit detection.[^14]

[Grounded in code] Col-span utilities validate spans and reject spans that cross frozen/non-frozen boundaries.[^14]

[Grounded in code] DOM utilities provide event propagation suppression and `scrollIntoView` with nearest alignment.[^14]

[Grounded in code] Event utilities create wrapper events with `preventGridDefault()` and `isGridDefaultPrevented()` while preserving the original event prototype.[^14]

[Grounded in code] Keyboard utilities classify non-input keys, detect Ctrl/Meta-modified events, map RTL arrow keys, and decide when editor navigation should be handled by the grid.[^14]

[Grounded in code] Style utilities produce grid-cell, frozen-cell, and header-cell inline styles and class-name composition.[^14]

[Grounded in code] Measuring helpers render hidden measuring cells with `data-measuring-cell-key` attributes so `useColumnWidths` can query actual max-content widths.[^14]

# 17. Performance strategy

## Explicit performance mechanisms

[Grounded in code] Rows, cells, header rows, summary rows, group rows, and related renderer components are memoized with React memoization where applicable.[^11][^12][^13]

[Grounded in code] `DataGrid` wraps many callback props with `useLatestFunc`; the source comment states this keeps callback identity stable and avoids breaking memoization.[^8][^10]

[Grounded in code] Row virtualization and column virtualization are first-class hooks. Frozen columns and active cells are retained while non-visible rows/columns are omitted.[^10]

[Grounded in docs] The README recommends `rowKeyGetter` for optimal rendering and advises consumers to preserve unchanged row object references so only changed rows rerender.[^6]

[Grounded in configuration] The build prefixes generated CSS classes with a version-derived `rdg-...` prefix to avoid style conflicts when multiple RDG versions are present on the same page.[^3]

## Measurement strategy

[Grounded in code] Auto and string-width columns are measured through hidden measuring cells. `useColumnWidths` skips remeasuring resized columns and remeasures flex-like columns when grid width changes.[^10]

## Controlled-state implications

[Reasoned inference] Because rows, sorting, and selection are controlled, consumers influence performance. Recreating all row objects or changing callback identities may defeat memoization, while stable keys and structural sharing let the grid reuse row/cell subtrees. This inference is consistent with the row-key and unchanged-row documentation and with the memoized component design.[^6][^8][^11]

## Known caveat

[Grounded in code] Variable row heights are computed for all rows up front, and the source comment calls out possible performance concerns for that approach.[^10]

# 18. Styling and DOM strategy

## Published CSS and CSS layers

[Grounded in configuration and code] The source imports `src/style/layers.css`, tsdown emits a bundled `styles.css`, and package exports expose it as `react-data-grid/lib/styles.css`.[^2][^3] The CSS layer file establishes ordered layers for root, focus sink, row, header row, summary row, group row, cell, header cell, summary cell, edit cell, and checkbox styles.[^15]

## CSS variables and themes

[Grounded in code] Root styles define CSS variables for background, header background, row hover/selected colors, text colors, border colors, scrollbar colors, font, and frozen-column shadow. Theme classes `rdg-dark` and `rdg-light` set `color-scheme`.[^15]

## Generated class names

[Grounded in code/configuration] Style modules define exported class-name constants through `ecij`, and the build prefixes generated class names with the package version.[^3][^15]

## Inline layout styles

[Grounded in code] Layout-critical values are passed as inline styles: grid templates, row/column positions, frozen offsets, scroll padding, CSS custom properties for header/summary/frozen dimensions, and measuring-cell positions.[^8][^14]

## Direct DOM interaction

[Grounded in code] The grid uses imperative DOM operations for focus, scrolling, measuring, resize observation, pointer capture, and drag/drop data transfer. Examples include querying active cells by `[data-active-cell="true"]`, focusing grid cells or focusable children, calling `scrollIntoView`, reading measuring-cell dimensions, and using pointer-capture events for resizing/fill operations.[^8][^10][^12][^14]

# 19. Tests, examples, and behavioral evidence

## Browser tests

[Grounded in tests] Browser tests are the richest behavior source. They validate event-default suppression, active-position callbacks, editing open/commit/cancel behavior, outside-click semantics, editor portals, `commitOnOutsideClick`, `closeOnExternalRowChange`, row selection, keyboard navigation, RTL navigation, sorting, copy/paste, drag fill, virtualization, colSpan navigation, column resizing, renderer precedence, and treegrid behavior.[^16][^17][^18][^19][^20][^21][^22]

## Node tests

[Grounded in tests] Node tests include SSR rendering with `window` undefined and assert that rendering produces nonempty markup. This supports the README/package claim that the package can run in SSR contexts, at least for server rendering a basic grid.[^23]

## Visual tests

[Grounded in tests] Visual tests render a basic grid with select column and summaries, and a tree grid with grouping and summaries, then compare screenshots. These protect layout and style regressions that behavioral assertions may miss.[^23]

## Custom browser commands

[Grounded in configuration] Vite/Vitest browser setup defines custom commands for column resizing and drag fill. Those commands abstract pointer interactions used repeatedly in tests.[^4]

## Website examples

[Grounded in repository structure] The `website/routes` directory provides example scenarios that correspond to major features: column grouping/spanning/reordering, custom renderers, infinite scrolling, no rows, row grouping, scroll-to-cell, tree view, variable row height, and others. I treat these as usage examples and scenario documentation. For implementation truth, code and tests remain stronger evidence.

# 20. Build, packaging, and distribution

## Package format

[Grounded in configuration] The package is ESM. `package.json` has `"type": "module"` and maps `main`, `module`, and `browser` to `./lib/index.js`; there is no separate CommonJS export in the inspected package metadata.[^2]

## Type declarations

[Grounded in configuration] `tsdown.config.ts` emits declarations using `tsconfig.src.json`, and `package.json` points `types` to `./lib/index.d.ts`.[^2][^3]

## CSS output

[Grounded in configuration] The tsdown CSS plugin writes `styles.css`; the package export map exposes `./lib/styles.css`; CSS files are marked as side effects.[^2][^3]

## Build scripts

[Grounded in configuration] `npm run build` runs `tsdown`. `build:website` runs TypeScript project build plus Vite build. `typecheck` runs `tsc -b` across the project. Test scripts run Vitest in normal, watch, CI, or screenshot-update mode.[^2]

## Test tooling

[Grounded in configuration] Vitest is configured with browser tests using Playwright providers for Chromium and Firefox, visual tests using Chromium, and node tests. Coverage is configured to include `src/**/*.{ts,tsx}` while excluding declarations, style modules, and `src/index.ts`.[^4]

## External dependencies

[Grounded in configuration] Runtime package dependencies are not declared beyond peer dependencies. Development dependencies include React, ReactDOM, TypeScript, Vite, Vitest, Playwright, tsdown, ecij, TanStack Router, and lint/format tooling.[^2]

# 21. Developer modification guide

## Add or modify a column feature

Start with `src/types.ts` if the feature changes public column shape. Then inspect `src/hooks/useCalculatedColumns.ts` for column normalization, ordering, grouped metadata, frozen metadata, and colSpan collection. Rendering behavior likely lives in `HeaderCell.tsx`, `Cell.tsx`, `SummaryCell.tsx`, and `GroupedColumnHeaderCell.tsx`. Width behavior belongs in `useColumnWidths.ts`; viewport behavior belongs in `useViewportColumns.ts`. Add tests under `test/browser/column` and consider website examples if the feature is externally visible.

## Add a new cell renderer behavior

Change `src/types.ts` if the renderer contract changes. For cell wrapper behavior, inspect `Cell.tsx` and `Row.tsx`. For reusable content renderers, use `src/cellRenderers`. If the behavior should be globally overridable, inspect `DataGridDefaultRenderersContext.ts` and the `Renderers` type. Add renderer precedence tests similar to `test/browser/renderers.test.tsx`.

## Change keyboard navigation

Primary files are `DataGrid.tsx`, `src/utils/activePositionUtils.ts`, `src/utils/keyboardUtils.ts`, and `src/hooks/useRovingTabIndex.ts`. For treegrid navigation, also inspect `TreeDataGrid.tsx`. Tests must cover `test/browser/keyboardNavigation.test.tsx`, colSpan navigation tests, and treegrid navigation tests because navigation crosses headers, body, summaries, colSpans, RTL, virtualization, and group rows.

## Change editing behavior

Primary files are `DataGrid.tsx`, `EditCell.tsx`, `src/utils/activePositionUtils.ts`, `src/utils/keyboardUtils.ts`, and `src/editors/renderTextEditor.tsx`. Watch the controlled `onRowsChange` contract: commit paths must produce a new rows array and correct `{ indexes, column }` metadata. Add or update tests in `test/browser/column/renderEditCell.test.tsx`, `events.test.tsx`, `copyPaste.test.tsx`, and drag-fill tests if editability changes.

## Change selection behavior

Primary files are `DataGrid.tsx`, `Columns.tsx`, `src/hooks/useRowSelection.ts`, `src/cellRenderers/renderCheckbox.tsx`, and `TreeDataGrid.tsx` for group selection. Tests should update `test/browser/rowSelection.test.tsx` and tree selection tests. Preserve the controlled `selectedRows` contract unless intentionally changing the public API.

## Change virtualization behavior

Primary files are `useViewportRows.ts`, `useViewportColumns.ts`, `useCalculatedColumns.ts`, `useColumnWidths.ts`, `ScrollToCell.tsx`, and `DataGrid.tsx`. Run virtualization, scroll-to-cell, colSpan, frozen-column, keyboard, and visual tests. The skipped disabled-virtualization/frozen-columns test is a known warning sign for this area.

## Change row rendering

Primary files are `Row.tsx`, `SummaryRow.tsx`, `GroupRow.tsx`, `DataGrid.tsx`, and the `RenderRowProps` type. Any changes to row class names or roles must be tested through row-class, keyboard, accessibility, and visual tests.

## Change styling or theme behavior

Primary files are under `src/style`, plus `src/utils/styleUtils.ts`, `tsdown.config.ts`, and `src/style/layers.css`. Verify CSS export behavior, generated class-name prefixing, visual tests, and any consumer-visible class names such as `rdg`, `rdg-light`, and `rdg-dark`.

## Add or update tests

Use `test/browser` for interaction contracts, `test/node` for SSR/package-neutral rendering behavior, and `test/visual` for layout/style regressions. Custom browser commands for resize/fill are configured in `vite.config.ts`.

## Add a website example

Use `website/routes` as the example surface. Before editing, inspect route conventions and generated route tree behavior; the formatting config ignores the generated route tree.[^5]

# 22. Evidence and inference ledger

## Strongly grounded conclusions

- The inspected package is `react-data-grid@7.0.0-beta.59`, ESM-only by package metadata, with React/ReactDOM peer dependencies and a separate CSS export.[^2]
- `src/index.ts` is the public export boundary, and `src/DataGrid.tsx` is the main coordinator.[^3][^8]
- Rows, sorting, and selection are controlled by consumers; the grid reports changes through callbacks rather than mutating or reordering data internally.[^6][^8][^12]
- Column widths can be internal or controlled, with special handling during active resize.[^8][^10]
- `TreeDataGrid` is a wrapper that computes grouped rows and renders `DataGrid` as a treegrid.[^13]
- Editing requires `renderEditCell` in the implementation, despite a type-comment implication that `editable` alone supplies a default text editor.[^7][^10][^14]
- Tests substantially cover the major interaction flows: editing, selection, keyboard/focus, sorting, virtualization, resizing, copy/paste, drag fill, renderer precedence, and tree behavior.[^16][^17][^18][^19][^20][^21][^22]

## Reasoned inferences

- The architecture intentionally centralizes cross-cutting orchestration in `DataGrid` while pushing layout calculations into hooks and local DOM behavior into leaf components. This is inferred from the module boundaries and call graph.[^8][^10][^11]
- Controlled rows plus memoized rows/cells mean consumer row identity strongly affects rendering performance. This is inferred from memoization and documentation recommending stable row keys and unchanged row object reuse.[^6][^11]
- The website routes serve as behavior examples rather than implementation authority. This is based on their scenario names and the stronger evidence available in source and tests.

## Areas needing further inspection before modification

- Exact website route implementation details. I verified the route surface but did not deeply inspect every example.
- CI workflow behavior under `.github`; no workflow files were visible in the extracted snapshot I inspected.
- Any undocumented compatibility expectations beyond package metadata, README, and tests.
- The active-position and colSpan utilities are intricate. Any navigation change should inspect all branches and rerun keyboard, colSpan, treegrid, and virtualization tests.
- Column grouping plus frozen/select-column reordering has an explicit TODO in `useCalculatedColumns`; changes there require focused testing.[^10]

## Potentially fragile assumptions

- Disabled virtualization with frozen columns has a skipped failing test, so behavior in that combination should not be assumed stable.[^19]
- Variable row-height virtualization calculates all row heights up front and is explicitly noted in code as a possible performance issue.[^10]
- The public type comment around default text editing conflicts with actual editability logic. Consumers may have expectations based on the comment or older behavior.[^7][^14]

## Highest-priority files for future investigation

1. `src/DataGrid.tsx`
2. `src/types.ts`
3. `src/hooks/useCalculatedColumns.ts`
4. `src/hooks/useViewportRows.ts`
5. `src/hooks/useViewportColumns.ts`
6. `src/hooks/useColumnWidths.ts`
7. `src/hooks/useActivePosition.ts`
8. `src/utils/activePositionUtils.ts`
9. `src/Cell.tsx`
10. `src/EditCell.tsx`
11. `src/HeaderCell.tsx`
12. `src/TreeDataGrid.tsx`
13. `src/Columns.tsx`
14. `test/browser/keyboardNavigation.test.tsx`
15. `test/browser/column/renderEditCell.test.tsx`
16. `test/browser/rowSelection.test.tsx`
17. `test/browser/virtualization.test.ts`
18. `test/browser/TreeDataGrid.test.tsx`

# Sources

[^1]: GitHub repository/API inspection for `Comcast/react-data-grid`, default branch `main`, retrieved April 21, 2026. Commit inspected: `5bdf1a1f714468eda05f162235a223075fa7ea21`; commit date `2026-04-17T15:48:10Z`; commit message `Bump oxfmt from 0.44.0 to 0.45.0 (#4031)`. Repository page top-level listing showed `.github`, `.vscode`, `src`, `test`, `website`, README, changelog, package/config files, and default branch `main`.
[^2]: `package.json` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 1-90. Covers package name/version, published files, ESM type, side effects, main/module/browser/types/export map, scripts, dev dependencies, and peer dependencies.
[^3]: `src/index.ts`, lines 1-57, and `tsdown.config.ts`, lines 6-27, at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`. Covers style import, public exports, build output, declaration generation, CSS output, and version-prefixed class-name plugin.
[^4]: `vite.config.ts` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 11-40, 51-75, 79-91, and 106-186. Covers custom browser commands, Vite app setup, coverage config, browser/visual/node Vitest projects, Playwright providers, viewport, and screenshot path.
[^5]: `tsconfig.base.json`, lines 2-17; `tsconfig.src.json`, lines 1-7; `tsconfig.test.json`, lines 1-9; `.oxfmtrc.json`, lines 1-67; `eslint.config.js`, lines 9-20, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^6]: `README.md` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 30-90, 252-330, 375-450, 510-699, 728-896, 1034-1113, and 1239-1290. Covers advertised features, installation/CSS import, columns/rows, row key and row-change behavior, column widths, selection, sorting, callbacks, ARIA/data attributes, TreeDataGrid, renderers, and column options.
[^7]: `src/types.ts` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 11-112, 123-242, 263-306, 316-343, and 355-383. Covers column types, calculated columns, group columns, renderer props, event args, row-change/fill/selection types, group rows, sort columns, colSpan/row-height args, renderers, active-position options, column widths, and direction.
[^8]: `src/DataGrid.tsx` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 80-222, 231-356, 361-478, 480-529, 534-706, 709-827, 829-925, 927-1095, and 1097-1324. Covers props, defaults, state, hooks, selection, scrolling, row updates, copy/paste, editing entry, resizing, drag fill, active-position navigation, editor construction, row rendering, root render, frozen shadows, measuring cells, scroll-to-cell, and DOM focus helpers.
[^9]: `src/hooks/useActivePosition.ts` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 15-141. Covers initial active position, validation, active row/column lookup, and reset/close behavior on row/column changes.
[^10]: `src/hooks/useCalculatedColumns.ts`, lines 28-290; `src/hooks/useViewportRows.ts`, lines 20-116; `src/hooks/useViewportColumns.ts`, lines 38-154; `src/hooks/useColumnWidths.ts`, lines 18-145; `src/hooks/useGridDimensions.ts`, lines 4-38; `src/hooks/useRovingTabIndex.ts`, lines 3-37; `src/hooks/useRowSelection.ts`, lines 10-62; `src/hooks/useLatestFunc.ts`, lines 5-19, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^11]: `src/Row.tsx`, lines 9-95; `src/Cell.tsx`, lines 37-133; `src/EditCell.tsx`, lines 14-205; `src/SummaryRow.tsx`, lines 41-90; `src/SummaryCell.tsx`, lines 16-49; `src/ScrollToCell.tsx`, lines 22-42, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^12]: `src/HeaderRow.tsx`, lines 55-106; `src/HeaderCell.tsx`, lines 88-317 and 325-382; `src/GroupedColumnHeaderRow.tsx`, lines 22-57; `src/GroupedColumnHeaderCell.tsx`, lines 23-50, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^13]: `src/TreeDataGrid.tsx`, lines 28-453; `src/GroupRow.tsx`, lines 36-105; `src/GroupCell.tsx`, lines 32-69; `src/Columns.tsx`, lines 1-73; `src/cellRenderers/renderCheckbox.tsx`, lines 25-43; `src/cellRenderers/renderToggleGroup.tsx`, lines 29-55; `src/cellRenderers/renderValue.tsx`, lines 3-4; `src/editors/renderTextEditor.tsx`, lines 36-56, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^14]: `src/utils/activePositionUtils.ts`, lines 11-237; `src/utils/colSpanUtils.ts`, lines 3-22; `src/utils/eventUtils.ts`, lines 3-20; `src/utils/keyboardUtils.ts`, lines 4-98; `src/utils/styleUtils.ts`, lines 4-60; `src/utils/domUtils.ts`, lines 3-9; `src/utils/renderMeasuringCells.tsx`, lines 5-21, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^15]: `src/style/core.ts`, lines 6-125; `src/style/cell.ts`, lines 3-68; `src/style/row.ts`, lines 5-56; `src/style/layers.css`, lines 1-15, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^16]: `test/browser/rowSelection.test.tsx` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 66-281. Covers checkbox selection, keyboard selection, header select-all/partial states, disabled rows, extra selected keys, shift range selection, and Shift+Space.
[^17]: `test/browser/column/renderEditCell.test.tsx`, lines 19-330, and `test/browser/events.test.tsx`, lines 56-192, at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`. Covers edit entry/commit/cancel/outside click/portal/options/focus behavior and mouse/key event default-prevention behavior.
[^18]: `test/browser/keyboardNavigation.test.tsx` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 34-359. Covers active-cell initialization, tab/arrows/page/home/end navigation, grid entry/exit, focusable renderers, virtualization-adjacent active cells, active reset, and RTL navigation.
[^19]: `test/browser/virtualization.test.ts`, lines 99-248; `test/browser/column/colSpan.test.ts`, lines 42-204; `test/browser/column/resizable.test.tsx`, lines 46-313, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^20]: `test/browser/copyPaste.test.tsx`, lines 71-137; `test/browser/dragFill.test.tsx`, lines 48-101; `test/browser/sorting.test.tsx`, lines 41-128, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
[^21]: `test/browser/TreeDataGrid.test.tsx` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 142-506. Covers grouping, group IDs, expansion/collapse, ARIA, selection, navigation, copy/paste, row updates, custom group rendering, and summary focus behavior.
[^22]: `test/browser/renderers.test.tsx` at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`, lines 100-277. Covers renderer context setup, no-rows fallback, checkbox renderer, sort-status renderer, cell renderer, and row renderer precedence.
[^23]: `test/node/ssr.test.tsx`, lines 25-31; `test/visual/basicGrid.test.tsx`, lines 43-55; `test/visual/treeGrid.test.tsx`, lines 64-80, all at commit `5bdf1a1f714468eda05f162235a223075fa7ea21`.
