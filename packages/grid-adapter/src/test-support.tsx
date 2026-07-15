// biome-ignore-all lint/a11y/noNoninteractiveElementToInteractiveRole: The test grid intentionally preserves RDG role attributes for selector compatibility.
// biome-ignore-all lint/a11y/noRedundantRoles: Explicit roles keep workbook tests independent of native accessibility-role inference.
// biome-ignore-all lint/a11y/useFocusableInteractive: This test renderer mirrors RDG's query surface, not a production interaction model.
import { gridScrollportClassName } from "@cartulary/ui-contracts";
import { Fragment, type KeyboardEvent } from "react";

import {
  assertGridRows,
  buildGridPresentationRows,
  type GridViewportProps,
  gridUnassignedGroupLabel,
  type WorkbookDataGridProps,
} from "./core";

export {
  assertGridRows,
  buildGridPresentationRows,
  type GridActionsColumn,
  type GridBlockSizing,
  type GridCellAnchor,
  type GridCellCopyIntent,
  type GridCellMutationIntent,
  type GridCellSelection,
  type GridCellTarget,
  type GridColumn,
  type GridDraftRow,
  type GridFillIntent,
  type GridGroupingDescriptor,
  type GridGroupingScalar,
  type GridHandle,
  type GridNavigationIntent,
  type GridNavigationKey,
  type GridRecordRow,
  type GridRowGutter,
  type GridSortDirection,
  type GridSortEntry,
  type GridViewportProps,
  navigateGridCellAnchor,
  resolveGridCellAnchor,
  type WorkbookDataGridProps,
} from "./core";

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

export function WorkbookDataGrid<Row>({
  actionsColumn,
  columns,
  draftRow,
  emptyMessage = "No rows",
  fillViewportInline = false,
  grouping = null,
  onSelectRecord,
  onToggleSort,
  recordRows,
  rowGutter,
  sort = [],
  viewSchemaId,
}: WorkbookDataGridProps<Row>) {
  assertGridRows(recordRows);

  const renderedRows = buildGridPresentationRows({
    grouping,
    rows: recordRows,
  });
  const totalColumnCount =
    columns.length +
    (rowGutter === undefined ? 0 : 1) +
    (actionsColumn === undefined ? 0 : 1);

  return (
    <table
      className={gridScrollportClassName()}
      role="grid"
      style={fillViewportInline ? { minWidth: 0, width: "100%" } : undefined}
    >
      <thead>
        <tr role="row">
          {rowGutter === undefined ? null : (
            <th role="columnheader" scope="col">
              <span data-testid={rowGutter.headerTestId}>
                {rowGutter.label ?? ""}
              </span>
            </th>
          )}
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
        {renderedRows.length === 0 && draftRow === undefined ? (
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
          <Fragment>
            {renderedRows.map((row) =>
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
                  aria-selected={
                    row.gridRow.selected === true ? "true" : undefined
                  }
                  data-grid-record-id={row.gridRow.recordId ?? ""}
                  data-testid={row.gridRow.testId}
                  key={row.key}
                  role="row"
                  tabIndex={onSelectRecord === undefined ? undefined : 0}
                  onClick={() => {
                    onSelectRecord?.(row.gridRow.recordId);
                  }}
                  onKeyDown={(event: KeyboardEvent<HTMLTableRowElement>) => {
                    if (event.key === "Enter" || event.key === " ") {
                      onSelectRecord?.(row.gridRow.recordId);
                    }
                  }}
                >
                  {rowGutter === undefined ? null : (
                    <th
                      data-grid-field-key="__cartulary_row_gutter__"
                      data-testid={row.gridRow.gutterTestId}
                      scope="row"
                    >
                      {row.gridRow.gutterContent ??
                        row.gridRow.gutterLabel ??
                        ""}
                    </th>
                  )}
                  {columns.map((column) => (
                    <td
                      data-grid-field-key={column.fieldKey}
                      key={column.fieldKey}
                      role="gridcell"
                    >
                      {column.renderCell({
                        anchor: {
                          fieldKey: column.fieldKey,
                          recordId: row.gridRow.recordId,
                          viewSchemaId,
                        },
                        row: row.gridRow.data,
                      })}
                    </td>
                  ))}
                  {actionsColumn === undefined ? null : (
                    <td role="gridcell">
                      {actionsColumn.renderCell(row.gridRow)}
                    </td>
                  )}
                </tr>
              ),
            )}
            {draftRow === undefined ? null : (
              <tr
                data-cartulary-grid-draft-row="true"
                data-grid-record-id=""
                data-testid={draftRow.testId}
                role="row"
              >
                {rowGutter === undefined ? null : (
                  <th
                    data-grid-field-key="__cartulary_row_gutter__"
                    scope="row"
                  >
                    {draftRow.gutterContent ?? draftRow.gutterLabel ?? ""}
                  </th>
                )}
                {columns.map((column) => (
                  <td
                    data-grid-field-key={column.fieldKey}
                    key={column.fieldKey}
                    role="gridcell"
                  >
                    {column.renderDraftCell?.({
                      fieldKey: column.fieldKey,
                      row: draftRow.data,
                      viewSchemaId,
                    }) ?? null}
                  </td>
                ))}
                {actionsColumn === undefined ? null : (
                  <td role="gridcell">
                    {actionsColumn.renderDraftCell?.(draftRow)}
                  </td>
                )}
              </tr>
            )}
          </Fragment>
        )}
      </tbody>
    </table>
  );
}
