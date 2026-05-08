import { gridScrollportClassName } from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type ForwardedRef,
  forwardRef,
  type Key,
  useCallback,
  useMemo,
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
  type GridActionsColumn,
  type GridColumn,
  type GridRow,
  type GridSortDirection,
  type GridSortEntry,
  type GridTableProps,
  type GridViewportProps,
  reconcileRecordRows,
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

  const totalColumnCount =
    columns.length + (actionsColumn === undefined ? 0 : 1);
  const firstColumnKey =
    columns[0]?.fieldKey ??
    (actionsColumn === undefined ? "" : actionsColumnKey);

  const renderedRows = useMemo(
    () =>
      buildGridPresentationRows({
        getGroupLabel,
        getGroupRowTestId,
        groupBy,
        rows,
      }),
    [getGroupLabel, getGroupRowTestId, groupBy, rows],
  );

  const rdgColumns = useMemo(() => {
    const mappedColumns = columns.map((column) =>
      buildDataColumn({
        column,
        firstColumnKey,
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
    columns,
    firstColumnKey,
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
      className={gridScrollportClassName()}
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
        return (
          <strong data-testid={row.testId}>
            {row.groupLabel ?? gridUnassignedGroupLabel}
          </strong>
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
    renderHeaderCell: () => <span>{actionsColumn.label}</span>,
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
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.82)",
  blockSize: "min(70vh, 46rem)",
  minBlockSize: "18rem",
};

const gridStyle = {
  blockSize: "100%",
  minWidth: "78rem",
  width: "max-content",
  "--rdg-background-color": "rgb(255 255 255 / 0.82)",
  "--rdg-border-color": "rgb(199 214 207)",
  "--rdg-header-background-color": "rgb(242 247 243)",
  "--rdg-row-hover-background-color": "rgb(247 250 248)",
  "--rdg-row-selected-background-color": "rgb(232 244 239)",
  "--rdg-row-selected-hover-background-color": "rgb(232 244 239)",
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
  color: "rgb(87 112 104)",
};

const bodyCellStyle = {
  padding: "0.75rem",
  borderBottom: "1px solid rgb(232 238 234)",
  verticalAlign: "top" as const,
};

const groupCellStyle = {
  ...bodyCellStyle,
  background: "rgb(247 249 247)",
  color: "rgb(53 79 72)",
};

const emptyStateStyle = {
  gridColumn: "1 / -1",
  padding: "0.75rem",
  textAlign: "center" as const,
  color: "rgb(70 92 85)",
};

const draftRowStyle = {
  background: "rgb(252 249 241)",
};

const selectedRowStyle = {
  background: "rgb(232 244 239)",
};

const rightAlignedStyle = {
  textAlign: "right" as const,
};

const centeredStyle = {
  textAlign: "center" as const,
};
