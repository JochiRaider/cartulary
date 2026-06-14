// biome-ignore-all lint/a11y/noNoninteractiveElementToInteractiveRole: The adapter owns the workbook ARIA grid role contract for stable cross-renderer selectors.
// biome-ignore-all lint/a11y/noRedundantRoles: Explicit roles keep workbook selectors independent of native accessibility-role inference.
// biome-ignore-all lint/a11y/useSemanticElements: The adapter intentionally mirrors RDG's div-based ARIA grid selector surface.
// biome-ignore-all lint/a11y/useFocusableInteractive: Grid cells host focusable editors and actions while preserving the workbook selector contract.
// biome-ignore-all lint/a11y/useKeyWithClickEvents: Row click selection mirrors the existing grid behavior; cell keyboard handling remains owned by rendered cell controls.
// biome-ignore-all lint/a11y/noStaticElementInteractions: Header click sorting intentionally stays non-focusable to preserve workbook grid focus restoration.
import { gridScrollportClassName } from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type ForwardedRef,
  forwardRef,
  type MouseEvent,
  type ReactNode,
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import "react-data-grid/lib/styles.css";

import {
  assertGridRows,
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridColumn,
  type GridDensity,
  type GridPresentationRow,
  type GridRow,
  type GridRowGutter,
  type GridSortEntry,
  type GridTableProps,
  type GridViewportProps,
  gridUnassignedGroupLabel,
} from "./core";

export {
  assertGridRows,
  buildGridPresentationRows,
  cleanupGridAdapters,
  type GridActionsColumn,
  type GridAdapterCleanup,
  type GridCellAnchor,
  type GridCellSelection,
  type GridColumn,
  type GridDensity,
  type GridEditorAdapter,
  type GridNavigationIntent,
  type GridNavigationKey,
  type GridPasteRowTarget,
  type GridPasteTargetResolution,
  type GridRendererAdapter,
  type GridRendererRegistry,
  type GridRow,
  type GridRowGutter,
  type GridSortDirection,
  type GridSortEntry,
  type GridTableProps,
  type GridViewportProps,
  isGridColumnEditable,
  navigateGridCellAnchor,
  reconcileRecordRows,
  resolveGridCellAnchor,
  resolveGridPasteTargets,
  resolveGridRenderer,
} from "./core";

export const gridAdapterVendor = "react-data-grid";

type AdapterGridRow<Row> = GridPresentationRow<Row>;

const actionsColumnKey = "__cartulary_actions__";
const defaultGridMinWidth = 1248;
const defaultDataColumnMinWidth = 144;
const defaultDataColumnWidth = 224;
const defaultActionsColumnMinWidth = 144;
const defaultActionsColumnWidth = 176;
const defaultRowGutterMinWidth = 48;
const defaultRowGutterWidth = 56;
const defaultGridClientHeight = 720;
const gridHeaderHeight = 32;
const gridVirtualizationOverscanRows = 3;
const gridDensityMetrics = {
  compact: {
    cellPaddingVar: "--ct-density-compact-cellPadding",
    rowHeight: 28,
  },
  default: {
    cellPaddingVar: "--ct-density-default-cellPadding",
    rowHeight: 36,
  },
  comfortable: {
    cellPaddingVar: "--ct-density-comfortable-cellPadding",
    rowHeight: 44,
  },
} as const satisfies Record<
  GridDensity,
  {
    readonly cellPaddingVar: string;
    readonly rowHeight: number;
  }
>;

export const GridViewport = forwardRef<HTMLDivElement, GridViewportProps>(
  function GridViewport(
    { children, className, style, testId }: GridViewportProps,
    ref: ForwardedRef<HTMLDivElement>,
  ) {
    return (
      <div
        className={className}
        data-testid={testId}
        ref={ref}
        style={resolveViewportStyle(style)}
      >
        {children}
      </div>
    );
  },
);

export function GridTable<Row>({
  actionsColumn,
  columns,
  density = "default",
  emptyMessage = "No rows",
  getGroupLabel,
  getGroupRowTestId,
  groupBy = null,
  onToggleSort,
  rowGutter,
  rows,
  sort = [],
}: GridTableProps<Row>) {
  assertGridRows(rows);
  const [collapsedGroups, setCollapsedGroups] = useState<ReadonlySet<string>>(
    () => new Set(),
  );
  const gridRef = useRef<HTMLDivElement | null>(null);
  const [viewportState, setViewportState] = useState({
    clientHeight: 0,
    scrollTop: 0,
  });

  const totalColumnCount =
    columns.length +
    (rowGutter === undefined ? 0 : 1) +
    (actionsColumn === undefined ? 0 : 1);

  const presentationRows = useMemo(
    () =>
      buildGridPresentationRows({
        getGroupLabel,
        getGroupRowTestId,
        groupBy,
        rows,
      }),
    [getGroupLabel, getGroupRowTestId, groupBy, rows],
  );
  const renderedRows = useMemo(
    () => filterCollapsedGroups(presentationRows, collapsedGroups),
    [collapsedGroups, presentationRows],
  );
  const handleToggleGroup = useCallback((groupKey: string) => {
    setCollapsedGroups((current) => {
      const next = new Set(current);
      if (next.has(groupKey)) {
        next.delete(groupKey);
      } else {
        next.add(groupKey);
      }
      return next;
    });
  }, []);
  const refreshViewportState = useCallback(() => {
    const grid = gridRef.current;
    if (grid === null) {
      return;
    }
    setViewportState((current) => {
      const next = {
        clientHeight: grid.clientHeight,
        scrollTop: grid.scrollTop,
      };
      return current.clientHeight === next.clientHeight &&
        current.scrollTop === next.scrollTop
        ? current
        : next;
    });
  }, []);
  useLayoutEffect(() => {
    refreshViewportState();
  }, [refreshViewportState]);
  const gridTemplateColumns = useMemo(
    () => buildGridTemplateColumns(columns, actionsColumn, rowGutter),
    [actionsColumn, columns, rowGutter],
  );
  const gridInlineSize = useMemo(
    () => resolveGridInlineSize(columns, actionsColumn, rowGutter),
    [actionsColumn, columns, rowGutter],
  );
  const densityMetrics = gridDensityMetrics[density];
  const resolvedGridStyle = useMemo(
    () =>
      ({
        ...gridStyle,
        "--cartulary-grid-cell-padding": `var(${densityMetrics.cellPaddingVar})`,
        "--cartulary-grid-density": density,
        "--cartulary-grid-row-height": `${densityMetrics.rowHeight}px`,
        gridTemplateColumns,
        minWidth: gridInlineSize,
        width: gridInlineSize,
      }) satisfies CSSProperties & Record<string, string | number>,
    [
      density,
      densityMetrics.cellPaddingVar,
      densityMetrics.rowHeight,
      gridInlineSize,
      gridTemplateColumns,
    ],
  );
  const virtualRows = useMemo(
    () =>
      buildVirtualizedRows({
        clientHeight: viewportState.clientHeight,
        density,
        rows: renderedRows,
        scrollTop: viewportState.scrollTop,
      }),
    [
      density,
      renderedRows,
      viewportState.clientHeight,
      viewportState.scrollTop,
    ],
  );
  const handleScroll = useCallback(() => {
    refreshViewportState();
  }, [refreshViewportState]);

  return (
    <div
      className={`${gridScrollportClassName()} cartulary-grid`}
      ref={gridRef}
      role="grid"
      style={resolvedGridStyle}
      onScroll={handleScroll}
    >
      <div role="row" style={rowContentsStyle}>
        {rowGutter === undefined ? null : (
          <div
            data-grid-field-key="__cartulary_row_gutter__"
            role="columnheader"
            style={rowGutterHeaderStyle}
          >
            <span data-testid={rowGutter.headerTestId}>
              {rowGutter.label ?? ""}
            </span>
          </div>
        )}
        {columns.map((column) => (
          <div
            key={column.fieldKey}
            role="columnheader"
            style={headerCellStyle(column.align)}
          >
            {renderDataHeaderContent({
              column,
              onToggleSort,
              sort,
            })}
          </div>
        ))}
        {actionsColumn === undefined ? null : (
          <div role="columnheader" style={headerCellStyle(undefined)}>
            <span data-testid={actionsColumn.headerTestId}>
              {actionsColumn.label}
            </span>
          </div>
        )}
      </div>
      {renderedRows.length === 0 ? (
        <div role="row" style={rowContentsStyle}>
          <div data-grid-empty="true" role="gridcell" style={emptyStateStyle}>
            {emptyMessage}
          </div>
        </div>
      ) : (
        virtualRows.items.map((item) =>
          item.kind === "spacer" ? (
            <GridSpacer height={item.height} key={item.key} />
          ) : item.row.kind === "group" ? (
            <GroupRow
              collapsed={collapsedGroups.has(item.row.key)}
              density={density}
              key={item.key}
              row={item.row}
              totalColumnCount={totalColumnCount}
              onToggleGroup={handleToggleGroup}
            />
          ) : (
            <DataRow
              actionsColumn={actionsColumn}
              columns={columns}
              density={density}
              key={item.key}
              rowGutter={rowGutter}
              row={item.row.gridRow}
            />
          ),
        )
      )}
    </div>
  );
}

type RenderDataHeaderContentProps<Row> = {
  readonly column: GridColumn<Row>;
  readonly onToggleSort?: ((fieldKey: string) => void) | undefined;
  readonly sort: readonly GridSortEntry[];
};

type GroupRowProps<Row> = {
  readonly collapsed: boolean;
  readonly density: GridDensity;
  readonly row: Extract<AdapterGridRow<Row>, { readonly kind: "group" }>;
  readonly totalColumnCount: number;
  readonly onToggleGroup: (groupKey: string) => void;
};

type DataRowProps<Row> = {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly density: GridDensity;
  readonly rowGutter: GridRowGutter | undefined;
  readonly row: GridRow<Row>;
};

type HeaderSortButtonProps<Row> = {
  readonly children: ReactNode;
  readonly column: GridColumn<Row>;
  readonly onToggleSort: (fieldKey: string) => void;
  readonly sortFieldKey: string;
};

type VirtualizedRows<Row> = {
  readonly items: readonly VirtualizedRowItem<Row>[];
};

type VirtualizedRowItem<Row> =
  | {
      readonly height: number;
      readonly key: string;
      readonly kind: "spacer";
    }
  | {
      readonly key: string;
      readonly kind: "row";
      readonly row: AdapterGridRow<Row>;
    };

type BuildVirtualizedRowsProps<Row> = {
  readonly clientHeight: number;
  readonly density: GridDensity;
  readonly rows: readonly AdapterGridRow<Row>[];
  readonly scrollTop: number;
};

function renderDataHeaderContent<Row>({
  column,
  onToggleSort,
  sort,
}: RenderDataHeaderContentProps<Row>) {
  const canToggleSort =
    onToggleSort !== undefined &&
    column.sortableFieldKey !== null &&
    column.sortableFieldKey !== undefined &&
    !column.sortDisabled;
  const sortFieldKey = column.sortableFieldKey ?? column.fieldKey;
  const sortState = sortStateForField(sort, sortFieldKey);

  const content = (
    <>
      <span>{column.label}</span>
      {canToggleSort ? (
        <span style={headerMetaStyle}>
          {sortState === undefined
            ? "Sort"
            : sortState.direction === "asc"
              ? "Asc"
              : "Desc"}
        </span>
      ) : null}
    </>
  );

  if (!canToggleSort) {
    return (
      <span
        data-grid-field-key={column.fieldKey}
        data-testid={column.headerTestId}
        style={headerContentStyle(column.align)}
        title={column.sortDisabledReason ?? undefined}
      >
        {content}
      </span>
    );
  }

  return (
    <HeaderSortButton
      column={column}
      sortFieldKey={sortFieldKey}
      onToggleSort={onToggleSort}
    >
      {content}
    </HeaderSortButton>
  );
}

function HeaderSortButton<Row>({
  children,
  column,
  onToggleSort,
  sortFieldKey,
}: HeaderSortButtonProps<Row>) {
  const [isFocused, setIsFocused] = useState(false);

  return (
    <button
      data-grid-field-key={column.fieldKey}
      data-testid={column.headerTestId}
      style={{
        ...headerButtonStyle(column.align),
        ...(isFocused ? headerButtonFocusStyle : null),
      }}
      title={column.sortDisabledReason ?? undefined}
      type="button"
      onBlur={() => {
        setIsFocused(false);
      }}
      onClick={() => {
        onToggleSort(sortFieldKey);
      }}
      onFocus={() => {
        setIsFocused(true);
      }}
    >
      {children}
    </button>
  );
}

function GroupRow<Row>({
  collapsed,
  density,
  row,
  totalColumnCount,
  onToggleGroup,
}: GroupRowProps<Row>) {
  return (
    <div data-grid-row-kind="group" role="row" style={rowContentsStyle}>
      <div
        data-grid-group-key={row.key}
        data-grid-row-kind="group"
        role="gridcell"
        style={{
          ...groupCellStyle,
          gridColumn: `span ${Math.max(totalColumnCount, 1)}`,
          minBlockSize: gridPresentationRowHeight(row, density),
        }}
      >
        <button
          aria-expanded={!collapsed}
          data-testid={row.testId}
          style={groupToggleStyle}
          type="button"
          onClick={() => {
            onToggleGroup(row.key);
          }}
        >
          <span aria-hidden="true" style={groupToggleGlyphStyle(collapsed)} />
          <strong>{row.groupLabel ?? gridUnassignedGroupLabel}</strong>
        </button>
      </div>
    </div>
  );
}

function DataRow<Row>({
  actionsColumn,
  columns,
  density,
  row,
  rowGutter,
}: DataRowProps<Row>) {
  const rowStyle = rowStyleForVariant(row);
  const handleClick =
    row.onSelect === undefined
      ? undefined
      : (event: MouseEvent<HTMLDivElement>) => {
          row.onSelect?.(event as unknown as MouseEvent<HTMLTableRowElement>);
        };

  return (
    <div
      aria-selected={row.selected === true ? "true" : undefined}
      data-grid-record-id={row.recordId ?? ""}
      data-testid={row.testId}
      role="row"
      style={rowContentsStyle}
      onClick={handleClick}
    >
      {rowGutter === undefined ? null : (
        <div
          data-grid-field-key="__cartulary_row_gutter__"
          data-testid={row.gutterTestId}
          role="rowheader"
          style={rowGutterCellStyle(rowStyle, density)}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </div>
      )}
      {columns.map((column) => (
        <div
          data-grid-field-key={column.fieldKey}
          key={column.fieldKey}
          role="gridcell"
          style={bodyCellStyleForColumn(column, rowStyle, density)}
        >
          {column.renderCell(row.data)}
        </div>
      ))}
      {actionsColumn === undefined ? null : (
        <div
          data-grid-field-key={actionsColumnKey}
          role="gridcell"
          style={bodyCellStyleForRow(rowStyle, density)}
        >
          {actionsColumn.renderCell(row)}
        </div>
      )}
    </div>
  );
}

function GridSpacer({ height }: { readonly height: number }) {
  if (height <= 0) {
    return null;
  }
  return <div aria-hidden="true" style={gridSpacerStyle(height)} />;
}

function bodyCellStyleForColumn<Row>(
  column: GridColumn<Row>,
  rowStyle: CSSProperties | undefined,
  density: GridDensity,
): CSSProperties {
  return {
    ...bodyCellStyle,
    ...alignmentStyle(column.align),
    minBlockSize: rowHeightForDensity(density),
    ...(rowStyle ?? {}),
  };
}

function bodyCellStyleForRow(
  rowStyle: CSSProperties | undefined,
  density: GridDensity,
): CSSProperties {
  return {
    ...bodyCellStyle,
    minBlockSize: rowHeightForDensity(density),
    ...(rowStyle ?? {}),
  };
}

function buildGridTemplateColumns<Row>(
  columns: readonly GridColumn<Row>[],
  actionsColumn: GridActionsColumn<Row> | undefined,
  rowGutter: GridRowGutter | undefined,
): string {
  const widths =
    rowGutter === undefined ? [] : [`${resolveRowGutterWidth(rowGutter)}px`];
  widths.push(
    ...columns.map((column) => `${resolveDataColumnWidth(column)}px`),
  );
  if (actionsColumn !== undefined) {
    widths.push(`${resolveActionsColumnWidth(actionsColumn)}px`);
  }
  return widths.join(" ") || `${defaultDataColumnWidth}px`;
}

function buildVirtualizedRows<Row>({
  clientHeight,
  density,
  rows,
  scrollTop,
}: BuildVirtualizedRowsProps<Row>): VirtualizedRows<Row> {
  if (rows.length === 0) {
    return {
      items: [],
    };
  }

  const bodyViewportTop = Math.max(0, scrollTop - gridHeaderHeight);
  const effectiveClientHeight =
    clientHeight > 0 ? clientHeight : defaultGridClientHeight;
  const dataRowHeight = rowHeightForDensity(density);
  const overscanPx = dataRowHeight * gridVirtualizationOverscanRows;
  const windowTop = Math.max(0, bodyViewportTop - overscanPx);
  const windowBottom = bodyViewportTop + effectiveClientHeight + overscanPx;
  const items: VirtualizedRowItem<Row>[] = [];
  let pendingSpacerHeight = 0;
  let offsetTop = 0;
  let spacerIndex = 0;

  for (const row of rows) {
    const height = gridPresentationRowHeight(row, density);
    const rowTop = offsetTop;
    const rowBottom = rowTop + height;
    const shouldMount =
      (rowBottom >= windowTop && rowTop <= windowBottom) ||
      (row.kind === "data" && row.gridRow.selected === true);
    if (shouldMount) {
      if (pendingSpacerHeight > 0) {
        items.push({
          height: pendingSpacerHeight,
          key: `spacer-${spacerIndex}`,
          kind: "spacer",
        });
        spacerIndex += 1;
        pendingSpacerHeight = 0;
      }
      items.push({ key: row.key, kind: "row", row });
    } else {
      pendingSpacerHeight += height;
    }
    offsetTop = rowBottom;
  }

  if (pendingSpacerHeight > 0) {
    items.push({
      height: pendingSpacerHeight,
      key: `spacer-${spacerIndex}`,
      kind: "spacer",
    });
  }

  return {
    items,
  };
}

function gridPresentationRowHeight<Row>(
  row: AdapterGridRow<Row>,
  density: GridDensity,
): number {
  return row.kind === "group" ? gridHeaderHeight : rowHeightForDensity(density);
}

function rowHeightForDensity(density: GridDensity): number {
  return gridDensityMetrics[density].rowHeight;
}

function resolveGridInlineSize<Row>(
  columns: readonly GridColumn<Row>[],
  actionsColumn: GridActionsColumn<Row> | undefined,
  rowGutter: GridRowGutter | undefined,
): number {
  const dataWidth = columns.reduce(
    (sum, column) => sum + resolveDataColumnWidth(column),
    0,
  );
  const gutterWidth =
    rowGutter === undefined ? 0 : resolveRowGutterWidth(rowGutter);
  const actionsWidth =
    actionsColumn === undefined ? 0 : resolveActionsColumnWidth(actionsColumn);
  return Math.max(defaultGridMinWidth, gutterWidth + dataWidth + actionsWidth);
}

function resolveViewportStyle(style?: CSSProperties): CSSProperties {
  return {
    ...viewportStyle,
    ...style,
  };
}

function filterCollapsedGroups<Row>(
  rows: readonly AdapterGridRow<Row>[],
  collapsedGroups: ReadonlySet<string>,
): readonly AdapterGridRow<Row>[] {
  if (collapsedGroups.size === 0) {
    return rows;
  }
  const filtered: AdapterGridRow<Row>[] = [];
  let currentGroupCollapsed = false;
  for (const row of rows) {
    if (row.kind === "group") {
      currentGroupCollapsed = collapsedGroups.has(row.key);
      filtered.push(row);
      continue;
    }
    if (!currentGroupCollapsed) {
      filtered.push(row);
    }
  }
  return filtered;
}

function rowStyleForVariant<Row>(row: GridRow<Row>): CSSProperties | undefined {
  if (row.variant === "draft") {
    return draftRowStyle;
  }
  if (row.selected) {
    return selectedRowStyle;
  }
  return undefined;
}

function sortStateForField(
  sort: readonly GridSortEntry[],
  fieldKey: string,
): GridSortEntry | undefined {
  return sort.find((entry) => entry.fieldKey === fieldKey);
}

function alignmentStyle(
  align: "left" | "center" | "right" | undefined,
): CSSProperties {
  if (align === "right") {
    return rightAlignedStyle;
  }
  if (align === "center") {
    return centeredStyle;
  }
  return {};
}

function headerContentStyle(
  align: "left" | "center" | "right" | undefined,
): CSSProperties {
  return {
    ...headerContentBaseStyle,
    ...alignmentStyle(align),
  };
}

function headerButtonStyle(
  align: "left" | "center" | "right" | undefined,
): CSSProperties {
  return {
    ...headerButtonBaseStyle,
    ...headerContentStyle(align),
  };
}

function headerCellStyle(
  align: "left" | "center" | "right" | undefined,
): CSSProperties {
  return {
    ...headerCellBaseStyle,
    ...alignmentStyle(align),
  };
}

function gridSpacerStyle(height: number): CSSProperties {
  return {
    ...spacerStyle,
    blockSize: height,
  };
}

function resolveDataColumnWidth<Row>(column: GridColumn<Row>): number {
  return Math.max(
    column.minWidth ?? defaultDataColumnMinWidth,
    column.width ?? defaultDataColumnWidth,
  );
}

function resolveActionsColumnWidth<Row>(
  actionsColumn: GridActionsColumn<Row>,
): number {
  return Math.max(
    actionsColumn.minWidth ?? defaultActionsColumnMinWidth,
    actionsColumn.width ?? defaultActionsColumnWidth,
  );
}

function resolveRowGutterWidth(rowGutter: GridRowGutter): number {
  return Math.max(
    rowGutter.minWidth ?? defaultRowGutterMinWidth,
    rowGutter.width ?? defaultRowGutterWidth,
  );
}

const viewportStyle = {
  overflow: "hidden",
  overflowAnchor: "none" as const,
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  blockSize: "min(70vh, 46rem)",
  minBlockSize: "18rem",
};

const gridStyle = {
  display: "grid",
  alignContent: "start",
  blockSize: "100%",
  overflow: "auto",
  overflowAnchor: "none" as const,
  fontFamily: "var(--ct-typography-grid-cell-fontFamily)",
  fontSize: "var(--ct-typography-grid-cell-fontSize)",
  fontVariantNumeric: "tabular-nums",
  fontFeatureSettings: '"tnum" 1, "zero" 1',
  color: "var(--ct-component-grid-cell-textColor)",
  "--rdg-background-color": "var(--ct-colors-surface-1)",
  "--rdg-border-color": "var(--ct-colors-hairline)",
  "--rdg-color": "var(--ct-colors-ink)",
  "--rdg-header-background-color": "var(--ct-colors-surface-2)",
  "--rdg-row-hover-background-color": "var(--ct-colors-surface-2)",
  "--rdg-row-selected-background-color": "var(--ct-colors-surface-3)",
  "--rdg-row-selected-hover-background-color": "var(--ct-colors-surface-3)",
} satisfies CSSProperties & Record<string, string | number>;

const rowContentsStyle = {
  display: "contents",
};

const headerCellBaseStyle = {
  position: "sticky" as const,
  top: 0,
  zIndex: 1,
  boxSizing: "border-box" as const,
  display: "flex",
  alignItems: "center",
  minWidth: 0,
  minBlockSize: gridHeaderHeight,
  padding: "0.3rem 0.5rem",
  borderBottom: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  fontWeight: 650,
};

const rowGutterHeaderStyle = {
  ...headerCellBaseStyle,
  left: 0,
  zIndex: 3,
  justifyContent: "center",
  color: "var(--ct-colors-ink-subtle)",
};

const headerContentBaseStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
  width: "100%",
};

const headerButtonBaseStyle = {
  minWidth: 0,
  padding: 0,
  border: 0,
  background: "transparent",
  color: "inherit",
  cursor: "pointer",
  font: "inherit",
};

const headerButtonFocusStyle = {
  outline: "var(--ct-component-focus-ring-border)",
  outlineOffset: "var(--ct-component-focus-ring-offset)",
  boxShadow: "0 0 0 2px var(--ct-colors-accent)",
};

const headerMetaStyle = {
  fontSize: "0.68rem",
  color: "var(--ct-colors-ink-subtle)",
};

const bodyCellStyle = {
  position: "relative" as const,
  boxSizing: "border-box" as const,
  minWidth: 0,
  minBlockSize: "var(--cartulary-grid-row-height)",
  padding: "var(--cartulary-grid-cell-padding)",
  borderBottom: "var(--ct-border-hairline)",
  lineHeight: "var(--ct-typography-grid-cell-lineHeight)",
  overflowWrap: "anywhere" as const,
  verticalAlign: "top" as const,
};

function rowGutterCellStyle(
  rowStyle: CSSProperties | undefined,
  density: GridDensity,
): CSSProperties {
  return {
    ...bodyCellStyle,
    ...rowStyle,
    minBlockSize: rowHeightForDensity(density),
    position: "sticky",
    left: 0,
    zIndex: 2,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    gap: "0.2rem",
    background:
      typeof rowStyle?.background === "string"
        ? rowStyle.background
        : "var(--ct-colors-surface-1)",
    color: "var(--ct-colors-ink-subtle)",
    fontSize: "0.76rem",
    fontWeight: 600,
    overflow: "hidden",
    whiteSpace: "nowrap",
  };
}

const groupCellStyle = {
  ...bodyCellStyle,
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  minBlockSize: gridHeaderHeight,
};

const spacerStyle = {
  gridColumn: "1 / -1",
  minWidth: 0,
  pointerEvents: "none" as const,
};

const groupToggleStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.5rem",
  minWidth: 0,
  padding: 0,
  border: 0,
  background: "transparent",
  color: "inherit",
  font: "inherit",
  cursor: "pointer",
};

function groupToggleGlyphStyle(collapsed: boolean): CSSProperties {
  const blockBorder = "0.3rem solid transparent";
  const colorBorder = "0.35rem solid currentColor";
  return {
    display: "inline-block",
    width: 0,
    height: 0,
    borderTop: collapsed ? blockBorder : colorBorder,
    borderBottom: collapsed ? blockBorder : 0,
    borderLeft: collapsed ? colorBorder : blockBorder,
    borderRight: 0,
  };
}

const emptyStateStyle = {
  gridColumn: "1 / -1",
  padding: "0.75rem",
  textAlign: "center" as const,
  color: "var(--ct-colors-ink-muted)",
};

const draftRowStyle = {
  background: "var(--ct-colors-surface-2)",
};

const selectedRowStyle = {
  background: "var(--ct-colors-surface-3)",
};

const rightAlignedStyle = {
  textAlign: "right" as const,
};

const centeredStyle = {
  textAlign: "center" as const,
};
