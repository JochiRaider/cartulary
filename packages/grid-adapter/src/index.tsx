import {
  forwardRef,
  type CSSProperties,
  type ForwardedRef,
  type MouseEventHandler,
  type PropsWithChildren,
  type ReactNode,
} from "react";

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
  readonly minWidth?: string | undefined;
  readonly width?: string | undefined;
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
  readonly minWidth?: string | undefined;
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

export const GridViewport = forwardRef<HTMLDivElement, GridViewportProps>(
  function GridViewport(
    { children, className, style, testId }: GridViewportProps,
    ref: ForwardedRef<HTMLDivElement>,
  ) {
    return (
      <div
        className={className}
        data-grid-adapter={gridAdapterVendor}
        data-testid={testId}
        ref={ref}
        style={{
          ...viewportStyle,
          ...style,
        }}
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

  const renderedColumns = [...columns];
  const totalColumnCount =
    renderedColumns.length + (actionsColumn === undefined ? 0 : 1);

  return (
    <table style={tableStyle}>
      <thead>
        <tr>
          {renderedColumns.map((column) => {
            const sortState = sortStateForField(
              sort,
              column.sortableFieldKey ?? column.fieldKey,
            );
            const canToggleSort =
              onToggleSort !== undefined &&
              column.sortableFieldKey !== null &&
              column.sortableFieldKey !== undefined &&
              !column.sortDisabled;
            const headerContent = canToggleSort ? (
              <button
                aria-label={`Sort ${column.label}`}
                data-grid-field-key={column.fieldKey}
                data-testid={column.headerTestId}
                style={headerButtonStyle}
                title={column.sortDisabledReason ?? undefined}
                type="button"
                onClick={() => {
                  if (column.sortableFieldKey) {
                    onToggleSort(column.sortableFieldKey);
                  }
                }}
              >
                <span>{column.label}</span>
                <span style={headerMetaStyle}>
                  {sortState === undefined
                    ? "Sort"
                    : sortState.direction === "asc"
                      ? "Asc"
                      : "Desc"}
                </span>
              </button>
            ) : (
              <span
                data-grid-editable={column.sortDisabled ? "false" : undefined}
                data-grid-field-key={column.fieldKey}
                data-testid={column.headerTestId}
                title={column.sortDisabledReason ?? undefined}
              >
                {column.label}
              </span>
            );
            return (
              <th
                key={column.fieldKey}
                style={{
                  ...headCellStyle,
                  ...columnWidthStyle(column),
                  ...(column.align === "right"
                    ? rightAlignedStyle
                    : column.align === "center"
                      ? centeredStyle
                      : null),
                }}
              >
                {headerContent}
              </th>
            );
          })}
          {actionsColumn ? (
            <th
              style={{
                ...headCellStyle,
                minWidth: actionsColumn.minWidth ?? "11rem",
              }}
            >
              {actionsColumn.label}
            </th>
          ) : null}
        </tr>
      </thead>
      <tbody>
        {rows.length < 1 ? (
          <tr>
            <td colSpan={totalColumnCount} style={emptyCellStyle}>
              {emptyMessage}
            </td>
          </tr>
        ) : (
          renderGridBody({
            actionsColumn,
            columns: renderedColumns,
            getGroupLabel,
            getGroupRowTestId,
            groupBy,
            rows,
          })
        )}
      </tbody>
    </table>
  );
}

type RenderGridBodyProps<Row> = {
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly getGroupLabel?:
    | ((row: Row, fieldKey: string) => string | null | undefined)
    | undefined;
  readonly getGroupRowTestId?:
    | ((fieldKey: string, value: string) => string | undefined)
    | undefined;
  readonly groupBy: string | null;
  readonly rows: readonly GridRow<Row>[];
};

function renderGridBody<Row>({
  actionsColumn,
  columns,
  getGroupLabel,
  getGroupRowTestId,
  groupBy,
  rows,
}: RenderGridBodyProps<Row>) {
  const parts: ReactNode[] = [];
  let activeGroupLabel: string | null = null;
  for (const row of rows) {
    const nextGroupLabel =
      groupBy === null || getGroupLabel === undefined
        ? null
        : normalizeGroupLabel(getGroupLabel(row.data, groupBy));
    if (groupBy !== null && nextGroupLabel !== activeGroupLabel) {
      activeGroupLabel = nextGroupLabel;
      parts.push(
        <tr
          data-testid={
            nextGroupLabel === null || getGroupRowTestId === undefined
              ? undefined
              : getGroupRowTestId(groupBy, nextGroupLabel)
          }
          key={`group:${groupBy}:${nextGroupLabel ?? "empty"}`}
        >
          <td
            colSpan={columns.length + (actionsColumn === undefined ? 0 : 1)}
            style={groupCellStyle}
          >
            <strong>{nextGroupLabel ?? "Unassigned"}</strong>
          </td>
        </tr>,
      );
    }
    parts.push(
      <tr
        key={row.key}
        data-grid-record-id={row.recordId ?? ""}
        data-testid={row.testId}
        style={{
          ...(row.variant === "draft" ? draftRowStyle : null),
          ...(row.selected ? selectedRowStyle : null),
        }}
        onClick={row.onSelect}
      >
        {columns.map((column) => (
          <td
            key={column.fieldKey}
            data-grid-field-key={column.fieldKey}
            style={{
              ...bodyCellStyle,
              ...columnWidthStyle(column),
              ...(column.align === "right"
                ? rightAlignedStyle
                : column.align === "center"
                  ? centeredStyle
                  : null),
            }}
          >
            {column.renderCell(row.data)}
          </td>
        ))}
        {actionsColumn ? (
          <td
            style={{
              ...bodyCellStyle,
              minWidth: actionsColumn.minWidth ?? "11rem",
            }}
          >
            {actionsColumn.renderCell(row)}
          </td>
        ) : null}
      </tr>,
    );
  }
  return parts;
}

type RecordIdentity = {
  readonly recordId: string | null;
};

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

function normalizeGroupLabel(
  value: string | null | undefined,
): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}

function columnWidthStyle<Row>(column: GridColumn<Row>): CSSProperties {
  return {
    minWidth: column.minWidth,
    width: column.width,
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

const tableStyle = {
  width: "100%",
  borderCollapse: "collapse" as const,
  minWidth: "78rem",
};

const headCellStyle = {
  position: "sticky" as const,
  top: 0,
  zIndex: 1,
  padding: "0.9rem 0.85rem",
  textAlign: "left" as const,
  fontSize: "0.8rem",
  letterSpacing: "0.08em",
  textTransform: "uppercase" as const,
  background: "rgb(242 247 243)",
  borderBottom: "1px solid rgb(207 221 214)",
};

const headerButtonStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
  width: "100%",
  border: 0,
  padding: 0,
  background: "transparent",
  color: "inherit",
  font: "inherit",
  cursor: "pointer",
  textTransform: "inherit" as const,
  letterSpacing: "inherit",
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

const emptyCellStyle = {
  ...bodyCellStyle,
  textAlign: "center" as const,
  color: "rgb(70 92 85)",
};

const groupCellStyle = {
  ...bodyCellStyle,
  background: "rgb(247 249 247)",
  color: "rgb(53 79 72)",
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
