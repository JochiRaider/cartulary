// biome-ignore-all lint/a11y/noNoninteractiveElementToInteractiveRole: The test grid intentionally preserves RDG role attributes for selector compatibility.
// biome-ignore-all lint/a11y/noRedundantRoles: Explicit roles keep workbook tests independent of native accessibility-role inference.
// biome-ignore-all lint/a11y/useFocusableInteractive: This test renderer mirrors RDG's query surface, not a production interaction model.
import type { KeyboardEvent, MouseEvent as ReactMouseEvent } from "react";

import {
  assertGridRows,
  buildGridPresentationRows,
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

export const gridAdapterVendor = "semantic-test-grid";

export function GridViewport({
  children,
  className,
  style,
  testId,
}: GridViewportProps) {
  return (
    <div className={className} data-testid={testId} style={style}>
      {children}
    </div>
  );
}

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

  const renderedRows = buildGridPresentationRows({
    getGroupLabel,
    getGroupRowTestId,
    groupBy,
    rows,
  });
  const totalColumnCount =
    columns.length + (actionsColumn === undefined ? 0 : 1);

  return (
    <table role="grid">
      <thead>
        <tr role="row">
          {columns.map((column) => {
            const canToggleSort =
              onToggleSort !== undefined &&
              column.sortableFieldKey !== null &&
              column.sortableFieldKey !== undefined &&
              !column.sortDisabled;
            const sortState = sort.find(
              (entry) =>
                entry.fieldKey === (column.sortableFieldKey ?? column.fieldKey),
            );
            return (
              <th key={column.fieldKey} role="columnheader" scope="col">
                <button
                  data-grid-field-key={column.fieldKey}
                  data-testid={column.headerTestId}
                  disabled={!canToggleSort}
                  title={column.sortDisabledReason ?? undefined}
                  type="button"
                  onClick={() => {
                    onToggleSort?.(column.sortableFieldKey ?? column.fieldKey);
                  }}
                >
                  <span>{column.label}</span>
                  {canToggleSort ? (
                    <span>
                      {sortState === undefined
                        ? "Sort"
                        : sortState.direction === "asc"
                          ? "Asc"
                          : "Desc"}
                    </span>
                  ) : null}
                </button>
              </th>
            );
          })}
          {actionsColumn === undefined ? null : (
            <th role="columnheader" scope="col">
              <span>{actionsColumn.label}</span>
            </th>
          )}
        </tr>
      </thead>
      <tbody>
        {renderedRows.length === 0 ? (
          <tr role="row">
            <td
              colSpan={totalColumnCount}
              data-grid-empty="true"
              role="gridcell"
            >
              {emptyMessage}
            </td>
          </tr>
        ) : (
          renderedRows.map((row) =>
            row.kind === "group" ? (
              <tr key={row.key} role="row">
                <td colSpan={totalColumnCount} role="gridcell">
                  <strong data-testid={row.testId}>
                    {row.groupLabel ?? gridUnassignedGroupLabel}
                  </strong>
                </td>
              </tr>
            ) : (
              <tr
                data-grid-record-id={row.gridRow.recordId ?? ""}
                data-testid={row.gridRow.testId}
                key={row.key}
                role="row"
                tabIndex={row.gridRow.onSelect === undefined ? undefined : 0}
                onClick={(event) => {
                  row.gridRow.onSelect?.(event);
                }}
                onKeyDown={(event: KeyboardEvent<HTMLTableRowElement>) => {
                  if (event.key === "Enter" || event.key === " ") {
                    row.gridRow.onSelect?.(
                      event as unknown as ReactMouseEvent<HTMLTableRowElement>,
                    );
                  }
                }}
              >
                {columns.map((column) => (
                  <td
                    data-grid-field-key={column.fieldKey}
                    key={column.fieldKey}
                    role="gridcell"
                  >
                    {column.renderCell(row.gridRow.data)}
                  </td>
                ))}
                {actionsColumn === undefined ? null : (
                  <td role="gridcell">
                    {actionsColumn.renderCell(row.gridRow)}
                  </td>
                )}
              </tr>
            ),
          )
        )}
      </tbody>
    </table>
  );
}
