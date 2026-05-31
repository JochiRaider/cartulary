import { gridScrollportClassName } from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type ForwardedRef,
  forwardRef,
  type Key,
  useCallback,
  useMemo,
  useState,
} from "react";
import {
  DataGrid,
  Cell as RDGCell,
  type CellRendererProps as RDGCellRendererProps,
  type Column as RDGColumn,
  type RenderRowProps as RDGRenderRowProps,
  Row as RDGRow,
  type SortColumn as RDGSortColumn,
} from "react-data-grid";

import "react-data-grid/lib/styles.css";

import {
  assertGridRows,
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridColumn,
  type GridPresentationRow,
  type GridRow,
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
  type GridEditorAdapter,
  type GridNavigationIntent,
  type GridNavigationKey,
  type GridPasteRowTarget,
  type GridPasteTargetResolution,
  type GridRendererAdapter,
  type GridRendererRegistry,
  type GridRow,
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
const defaultDataColumnMinWidth = 144;
const defaultDataColumnWidth = 224;
const defaultActionsColumnMinWidth = 144;
const defaultActionsColumnWidth = 176;

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
  emptyMessage = "No rows",
  getGroupLabel,
  getGroupRowTestId,
  groupBy = null,
  onToggleSort,
  rows,
  sort = [],
}: GridTableProps<Row>) {
  assertGridRows(rows);
  const [collapsedGroups, setCollapsedGroups] = useState<ReadonlySet<string>>(
    () => new Set(),
  );

  const totalColumnCount =
    columns.length + (actionsColumn === undefined ? 0 : 1);
  const firstColumnKey =
    columns[0]?.fieldKey ??
    (actionsColumn === undefined ? "" : actionsColumnKey);

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

  const rdgColumns = useMemo(() => {
    const mappedColumns = columns.map((column) =>
      buildDataColumn({
        column,
        firstColumnKey,
        isGroupCollapsed: (groupKey) => collapsedGroups.has(groupKey),
        onToggleGroup: handleToggleGroup,
        onToggleSort,
        sort,
        totalColumnCount,
      }),
    );

    if (actionsColumn === undefined) {
      return mappedColumns;
    }

    return [
      ...mappedColumns,
      buildActionsColumn({
        actionsColumn,
        firstColumnKey,
        totalColumnCount,
      }),
    ];
  }, [
    actionsColumn,
    collapsedGroups,
    columns,
    firstColumnKey,
    handleToggleGroup,
    onToggleSort,
    sort,
    totalColumnCount,
  ]);

  const cellStylesByKey = useMemo(() => {
    const styles = new Map<string, CSSProperties>();
    for (const column of columns) {
      styles.set(column.fieldKey, {
        ...bodyCellStyle,
        ...alignmentStyle(column.align),
      });
    }
    if (actionsColumn) {
      styles.set(actionsColumnKey, bodyCellStyle);
    }
    return styles;
  }, [actionsColumn, columns]);

  const sortColumns = useMemo(() => toRDGSortColumns(sort), [sort]);
  const noRowsFallback = useMemo(
    () => (
      <div style={emptyStateStyle} data-grid-empty="true">
        {emptyMessage}
      </div>
    ),
    [emptyMessage],
  );
  const renderCell = useCallback(
    (key: Key, props: RDGCellRendererProps<AdapterGridRow<Row>, unknown>) => {
      const cellStyle =
        props.row.kind === "group" && props.column.key === firstColumnKey
          ? groupCellStyle
          : (cellStylesByKey.get(props.column.key) ?? bodyCellStyle);
      return <AdapterCell key={key} {...props} cellStyle={cellStyle} />;
    },
    [cellStylesByKey, firstColumnKey],
  );
  const renderRow = useCallback(
    (key: Key, props: RDGRenderRowProps<AdapterGridRow<Row>, unknown>) => (
      <AdapterRow
        key={key}
        {...props}
        rowStyle={
          props.row.kind === "group"
            ? undefined
            : rowStyleForVariant(props.row.gridRow)
        }
      />
    ),
    [],
  );
  const renderers = useMemo(
    () => ({
      noRowsFallback,
      renderCell,
      renderRow,
    }),
    [noRowsFallback, renderCell, renderRow],
  );
  const handleSortColumnsChange = useCallback(
    (nextSortColumns: readonly RDGSortColumn[]) => {
      const nextFieldKey = nextSortColumns[0]?.columnKey ?? sort[0]?.fieldKey;
      if (nextFieldKey) {
        onToggleSort?.(nextFieldKey);
      }
    },
    [onToggleSort, sort],
  );

  return (
    <DataGrid<AdapterGridRow<Row>, unknown>
      className={`${gridScrollportClassName()} cartulary-grid`}
      columns={rdgColumns}
      enableVirtualization
      renderers={renderers}
      rowHeight={gridRowHeight}
      rowKeyGetter={gridRowKeyGetter}
      rows={renderedRows}
      sortColumns={sortColumns}
      style={gridStyle}
      headerRowHeight={56}
      onSortColumnsChange={
        onToggleSort === undefined ? undefined : handleSortColumnsChange
      }
    />
  );
}

type BuildDataColumnProps<Row> = {
  readonly column: GridColumn<Row>;
  readonly firstColumnKey: string;
  readonly isGroupCollapsed: (groupKey: string) => boolean;
  readonly onToggleGroup: (groupKey: string) => void;
  readonly onToggleSort?: ((fieldKey: string) => void) | undefined;
  readonly sort: readonly GridSortEntry[];
  readonly totalColumnCount: number;
};

type BuildActionsColumnProps<Row> = {
  readonly actionsColumn: GridActionsColumn<Row>;
  readonly firstColumnKey: string;
  readonly totalColumnCount: number;
};

function AdapterCell<Row>({
  cellStyle,
  ...props
}: RDGCellRendererProps<AdapterGridRow<Row>, unknown> & {
  readonly cellStyle: CSSProperties;
}) {
  return (
    <RDGCell
      {...props}
      data-grid-field-key={props.column.key}
      style={cellStyle}
    />
  );
}

function AdapterRow<Row>({
  row,
  rowStyle,
  style,
  ...props
}: RDGRenderRowProps<AdapterGridRow<Row>, unknown> & {
  readonly rowStyle?: CSSProperties | undefined;
}) {
  if (row.kind === "group") {
    return <RDGRow {...props} row={row} style={style} />;
  }

  return (
    <RDGRow
      {...props}
      data-grid-record-id={row.gridRow.recordId ?? ""}
      data-testid={row.gridRow.testId}
      row={row}
      style={rowStyle === undefined ? style : { ...style, ...rowStyle }}
      onClick={
        row.gridRow.onSelect === undefined
          ? undefined
          : (event) => {
              row.gridRow.onSelect?.(
                event as unknown as React.MouseEvent<HTMLTableRowElement>,
              );
            }
      }
    />
  );
}

function buildDataColumn<Row>({
  column,
  firstColumnKey,
  isGroupCollapsed,
  onToggleGroup,
  onToggleSort,
  sort,
  totalColumnCount,
}: BuildDataColumnProps<Row>): RDGColumn<AdapterGridRow<Row>, unknown> {
  const canToggleSort =
    onToggleSort !== undefined &&
    column.sortableFieldKey !== null &&
    column.sortableFieldKey !== undefined &&
    !column.sortDisabled;

  return {
    key: column.fieldKey,
    minWidth: column.minWidth ?? defaultDataColumnMinWidth,
    name: column.label,
    renderCell: ({ row }) => {
      if (row.kind === "group") {
        if (column.fieldKey !== firstColumnKey) {
          return null;
        }
        const collapsed = isGroupCollapsed(row.key);
        return (
          <button
            aria-expanded={!collapsed}
            data-testid={row.testId}
            onClick={() => {
              onToggleGroup(row.key);
            }}
            style={groupToggleStyle}
            type="button"
          >
            <span aria-hidden="true" style={groupToggleGlyphStyle(collapsed)} />
            <strong>{row.groupLabel ?? gridUnassignedGroupLabel}</strong>
          </button>
        );
      }
      return column.renderCell(row.gridRow.data);
    },
    renderHeaderCell: () => {
      const sortState = sortStateForField(
        sort,
        column.sortableFieldKey ?? column.fieldKey,
      );
      return (
        <span
          data-grid-field-key={column.fieldKey}
          data-testid={column.headerTestId}
          style={headerContentStyle(column.align)}
          title={column.sortDisabledReason ?? undefined}
        >
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
        </span>
      );
    },
    sortable: canToggleSort,
    width: column.width ?? defaultDataColumnWidth,
    colSpan: (args) =>
      args.type === "ROW" &&
      args.row.kind === "group" &&
      column.fieldKey === firstColumnKey
        ? totalColumnCount
        : undefined,
  };
}

function buildActionsColumn<Row>({
  actionsColumn,
  firstColumnKey,
  totalColumnCount,
}: BuildActionsColumnProps<Row>): RDGColumn<AdapterGridRow<Row>, unknown> {
  return {
    key: actionsColumnKey,
    minWidth: actionsColumn.minWidth ?? defaultActionsColumnMinWidth,
    name: actionsColumn.label,
    renderCell: ({ row }) => {
      if (row.kind === "group") {
        if (firstColumnKey !== actionsColumnKey) {
          return null;
        }
        return (
          <strong data-testid={row.testId}>
            {row.groupLabel ?? gridUnassignedGroupLabel}
          </strong>
        );
      }
      return actionsColumn.renderCell(row.gridRow);
    },
    renderHeaderCell: () => (
      <span data-testid={actionsColumn.headerTestId}>
        {actionsColumn.label}
      </span>
    ),
    sortable: false,
    width: actionsColumn.width ?? defaultActionsColumnWidth,
    colSpan: (args) =>
      args.type === "ROW" &&
      args.row.kind === "group" &&
      firstColumnKey === actionsColumnKey
        ? totalColumnCount
        : undefined,
  };
}

function resolveViewportStyle(style?: CSSProperties): CSSProperties {
  return {
    ...viewportStyle,
    ...style,
  };
}

function gridRowHeight<Row>(row: AdapterGridRow<Row>) {
  return row.kind === "group" ? 48 : 168;
}

function gridRowKeyGetter<Row>(row: AdapterGridRow<Row>) {
  return row.key;
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

function toRDGSortColumns(
  sort: readonly GridSortEntry[],
): readonly RDGSortColumn[] {
  return sort.map((entry) => ({
    columnKey: entry.fieldKey,
    direction: entry.direction === "asc" ? "ASC" : "DESC",
  }));
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
  blockSize: "100%",
  minWidth: "78rem",
  width: "max-content",
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

const headerContentBaseStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
  width: "100%",
};

const headerMetaStyle = {
  fontSize: "0.7rem",
  color: "var(--ct-colors-ink-subtle)",
};

const bodyCellStyle = {
  padding: "var(--ct-component-grid-cell-padding)",
  borderBottom: "var(--ct-border-hairline)",
  verticalAlign: "top" as const,
};

const groupCellStyle = {
  ...bodyCellStyle,
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
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
