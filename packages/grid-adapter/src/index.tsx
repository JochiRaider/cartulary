import {
  type CSSProperties,
  type ForwardedRef,
  forwardRef,
  type Key,
  type MouseEventHandler,
  type PropsWithChildren,
  type ReactNode,
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

export const gridAdapterVendor = "react-data-grid";

export type GridSortDirection = "asc" | "desc";

export type GridSortEntry = {
  readonly fieldKey: string;
  readonly direction: GridSortDirection;
};

export type GridColumn<Row> = {
  readonly fieldKey: string;
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: Row) => ReactNode;
  readonly sortableFieldKey?: string | null;
  readonly sortDisabled?: boolean | undefined;
  readonly sortDisabledReason?: string | null | undefined;
  readonly align?: "left" | "center" | "right" | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridRow<Row> = {
  readonly key: string;
  readonly recordId: string | null;
  readonly data: Row;
  readonly onSelect?: MouseEventHandler<HTMLTableRowElement> | undefined;
  readonly selected?: boolean | undefined;
  readonly variant?: "default" | "draft" | undefined;
  readonly testId?: string | undefined;
};

export type GridActionsColumn<Row> = {
  readonly label: string;
  readonly renderCell: (row: GridRow<Row>) => ReactNode;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridViewportProps = PropsWithChildren<{
  readonly className?: string | undefined;
  readonly style?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}>;

export type GridTableProps<Row> = {
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly emptyMessage?: ReactNode | undefined;
  readonly getGroupLabel?: (
    row: Row,
    fieldKey: string,
  ) => string | null | undefined;
  readonly getGroupRowTestId?: (
    fieldKey: string,
    value: string,
  ) => string | undefined;
  readonly groupBy?: string | null | undefined;
  readonly onToggleSort?: ((fieldKey: string) => void) | undefined;
  readonly rows: readonly GridRow<Row>[];
  readonly sort?: readonly GridSortEntry[] | undefined;
};

type AdapterGridGroupRow = {
  readonly groupBy: string;
  readonly groupLabel: string | null;
  readonly key: string;
  readonly kind: "group";
  readonly testId?: string | undefined;
};

type AdapterGridDataRow<Row> = {
  readonly gridRow: GridRow<Row>;
  readonly key: string;
  readonly kind: "data";
};

type AdapterGridRow<Row> = AdapterGridGroupRow | AdapterGridDataRow<Row>;

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
      buildRenderedRows({
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
      columns={rdgColumns}
      enableVirtualization={false}
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

export function assertGridRows<Row extends RecordIdentity>(
  rows: readonly Row[],
) {
  const seen = new Set<string>();
  for (const row of rows) {
    if (row.recordId === null) {
      continue;
    }
    const normalized = row.recordId.trim();
    if (normalized === "") {
      throw new Error(
        "Grid adapter invariant failed: missing record_id on a saved row.",
      );
    }
    if (seen.has(normalized)) {
      throw new Error(
        `Grid adapter invariant failed: duplicate record_id "${normalized}".`,
      );
    }
    seen.add(normalized);
  }
}

export function reconcileRecordRows<Row extends RecordIdentity>(
  previousRows: readonly Row[],
  nextRows: readonly Row[],
): readonly Row[] {
  const previousByRecordId = new Map<string, Row>();
  for (const row of previousRows) {
    if (row.recordId !== null && row.recordId.trim() !== "") {
      previousByRecordId.set(row.recordId, row);
    }
  }
  return nextRows.map((row) => {
    if (row.recordId === null || row.recordId.trim() === "") {
      return row;
    }
    const previous = previousByRecordId.get(row.recordId);
    if (previous === undefined) {
      return row;
    }
    if (shallowEqualRecord(previous, row)) {
      return previous;
    }
    return row;
  });
}

type RecordIdentity = {
  readonly recordId: string | null;
};

type BuildRenderedRowsProps<Row> = {
  readonly getGroupLabel?:
    | ((row: Row, fieldKey: string) => string | null | undefined)
    | undefined;
  readonly getGroupRowTestId?:
    | ((fieldKey: string, value: string) => string | undefined)
    | undefined;
  readonly groupBy: string | null;
  readonly rows: readonly GridRow<Row>[];
};

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

function buildRenderedRows<Row>({
  getGroupLabel,
  getGroupRowTestId,
  groupBy,
  rows,
}: BuildRenderedRowsProps<Row>): readonly AdapterGridRow<Row>[] {
  if (groupBy === null || getGroupLabel === undefined) {
    return rows.map((row) => ({
      gridRow: row,
      key: row.key,
      kind: "data",
    }));
  }

  const renderedRows: AdapterGridRow<Row>[] = [];
  let activeGroupLabel: string | null = null;

  for (const row of rows) {
    const nextGroupLabel = normalizeGroupLabel(
      getGroupLabel(row.data, groupBy),
    );
    if (nextGroupLabel !== activeGroupLabel) {
      activeGroupLabel = nextGroupLabel;
      renderedRows.push({
        groupBy,
        groupLabel: nextGroupLabel,
        key: `group:${groupBy}:${nextGroupLabel ?? "empty"}`,
        kind: "group",
        testId:
          nextGroupLabel === null || getGroupRowTestId === undefined
            ? undefined
            : getGroupRowTestId(groupBy, nextGroupLabel),
      });
    }
    renderedRows.push({
      gridRow: row,
      key: row.key,
      kind: "data",
    });
  }

  return renderedRows;
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
            {row.groupLabel ?? "Unassigned"}
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
            {row.groupLabel ?? "Unassigned"}
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

function shallowEqualRecord<Row extends object>(left: Row, right: Row) {
  const leftRecord = left as Record<string, unknown>;
  const rightRecord = right as Record<string, unknown>;
  const leftKeys = Object.keys(leftRecord);
  const rightKeys = Object.keys(rightRecord);
  if (leftKeys.length !== rightKeys.length) {
    return false;
  }
  return leftKeys.every((key) => Object.is(leftRecord[key], rightRecord[key]));
}

function sortStateForField(
  sort: readonly GridSortEntry[],
  fieldKey: string,
): GridSortEntry | undefined {
  return sort.find((entry) => entry.fieldKey === fieldKey);
}

function normalizeGroupLabel(value: string | null | undefined): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
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
  overflow: "auto",
  overflowAnchor: "none" as const,
  borderRadius: "1rem",
  border: "1px solid rgb(199 214 207)",
  background: "rgb(255 255 255 / 0.82)",
  maxHeight: "70vh",
};

const gridStyle = {
  blockSize: "auto",
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
