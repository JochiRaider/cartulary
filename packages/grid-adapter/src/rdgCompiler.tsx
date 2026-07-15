import type { KeyboardEvent } from "react";
import type { Column } from "react-data-grid";
import type {
  GridActionsColumn,
  GridColumn,
  GridDraftRow,
  GridRecordRow,
  GridRowGutter,
} from "./core";

export const gridActionsColumnKey = "__cartulary_actions__";
export const gridRowGutterColumnKey = "__cartulary_row_gutter__";

type CompileGridColumnsInput<Row> = {
  readonly actionsColumn: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly rowGutter: GridRowGutter | undefined;
  readonly viewSchemaId: string;
};

export function compileGridColumns<Row>({
  actionsColumn,
  columns,
  rowGutter,
  viewSchemaId,
}: CompileGridColumnsInput<Row>): readonly Column<
  GridRecordRow<Row>,
  GridDraftRow<Row>
>[] {
  const firstColumnKey =
    rowGutter !== undefined
      ? gridRowGutterColumnKey
      : (columns[0]?.fieldKey ??
        (actionsColumn === undefined ? undefined : gridActionsColumnKey));
  const compiled: Array<Column<GridRecordRow<Row>, GridDraftRow<Row>>> = [];

  if (rowGutter !== undefined) {
    compiled.push({
      frozen: true,
      key: gridRowGutterColumnKey,
      minWidth: rowGutter.minWidth,
      name: (
        <span
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: "Row gutter",
              fieldKey: gridRowGutterColumnKey,
              testId: rowGutter.headerTestId,
            })
          }
        >
          {rowGutter.label}
        </span>
      ),
      renderCell: ({ row }) => (
        <span
          data-grid-field-key={gridRowGutterColumnKey}
          ref={(node) => markSemanticGutterCell(node, row.gutterTestId)}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </span>
      ),
      renderSummaryCell: ({ row }) => (
        <span
          {...draftMarker(row, gridRowGutterColumnKey, firstColumnKey)}
          data-grid-field-key={gridRowGutterColumnKey}
        >
          {row.gutterContent ?? row.gutterLabel ?? ""}
        </span>
      ),
      resizable: false,
      width: rowGutter.width,
    });
  }

  for (const column of columns) {
    compiled.push({
      cellClass: dataCellClass(column.align),
      editable:
        column.contractWritable === true && column.renderEditCell !== undefined,
      key: column.fieldKey,
      minWidth: column.minWidth,
      name: column.label,
      renderHeaderCell: ({ sortDirection }) => (
        <span
          data-grid-field-key={column.fieldKey}
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: column.label,
              fieldKey: column.fieldKey,
              testId: column.headerTestId,
              title: column.sortDisabledReason ?? undefined,
            })
          }
        >
          {column.label}
          {sortDirection === "ASC" ? " Asc" : null}
          {sortDirection === "DESC" ? " Desc" : null}
        </span>
      ),
      renderCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the gridcell; this wrapper only prevents duplicate vendor navigation after a nested control handles the key.
        <span
          className="cartulary-grid-cell-content"
          data-grid-field-key={column.fieldKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {column.renderCell({
            anchor: {
              fieldKey: column.fieldKey,
              recordId: row.recordId,
              viewSchemaId,
            },
            row: row.data,
          })}
        </span>
      ),
      renderEditCell:
        column.contractWritable === true && column.renderEditCell !== undefined
          ? ({ onClose, onRowChange, row }) =>
              column.renderEditCell?.({
                closeEditor: (commit) => {
                  onClose(commit, true);
                },
                row: row.data,
                target: {
                  baseRowVersion: row.rowVersion,
                  fieldKey: column.fieldKey,
                  recordId: row.recordId,
                  viewSchemaId,
                },
                updateRow: (data) => {
                  onRowChange({ ...row, data });
                },
              })
          : undefined,
      renderSummaryCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the summary gridcell; this wrapper only prevents duplicate vendor navigation after a nested control handles the key.
        <span
          {...draftMarker(row, column.fieldKey, firstColumnKey)}
          className="cartulary-grid-cell-content"
          data-grid-field-key={column.fieldKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {column.renderDraftCell?.({
            fieldKey: column.fieldKey,
            row: row.data,
            viewSchemaId,
          }) ?? null}
        </span>
      ),
      resizable: true,
      summaryCellClass: "cartulary-grid-data-cell",
      sortable:
        column.sortableFieldKey !== null &&
        column.sortableFieldKey !== undefined &&
        !column.sortDisabled,
      width: column.width,
    });
  }

  if (actionsColumn !== undefined) {
    compiled.push({
      key: gridActionsColumnKey,
      minWidth: actionsColumn.minWidth,
      name: (
        <span
          ref={(node) =>
            markSemanticHeaderCell(node, {
              accessibleLabel: actionsColumn.label,
              fieldKey: gridActionsColumnKey,
              testId: actionsColumn.headerTestId,
            })
          }
        >
          {actionsColumn.label}
        </span>
      ),
      renderCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the gridcell; this wrapper only prevents duplicate vendor navigation after a nested action handles the key.
        <span
          className="cartulary-grid-cell-content"
          data-grid-field-key={gridActionsColumnKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {actionsColumn.renderCell(row)}
        </span>
      ),
      renderSummaryCell: ({ row }) => (
        // biome-ignore lint/a11y/noStaticElementInteractions: RDG owns the summary gridcell; this wrapper only prevents duplicate vendor navigation after a nested action handles the key.
        <span
          {...draftMarker(row, gridActionsColumnKey, firstColumnKey)}
          className="cartulary-grid-cell-content"
          data-grid-field-key={gridActionsColumnKey}
          onKeyDown={stopVendorNavigationForInteractiveContent}
          role="presentation"
        >
          {actionsColumn.renderDraftCell?.(row)}
        </span>
      ),
      resizable: false,
      width: actionsColumn.width,
    });
  }

  return compiled;
}

function markSemanticHeaderCell(
  node: HTMLSpanElement | null,
  options: {
    readonly accessibleLabel: string;
    readonly fieldKey: string;
    readonly testId: string | undefined;
    readonly title?: string | undefined;
  },
) {
  const header = node?.closest<HTMLElement>('[role="columnheader"]');
  if (header === undefined || header === null) return;
  header.dataset.gridFieldKey = options.fieldKey;
  if (options.testId !== undefined) header.dataset.testid = options.testId;
  header.setAttribute("aria-label", options.accessibleLabel);
  if (options.title !== undefined) header.title = options.title;
}

function draftMarker<Row>(
  row: GridDraftRow<Row>,
  columnKey: string,
  firstColumnKey: string | undefined,
) {
  const isMarkerCell = columnKey === firstColumnKey;
  return {
    "data-cartulary-grid-draft-row": isMarkerCell ? "true" : undefined,
    ref: (node: HTMLSpanElement | null) => {
      const semanticRow = node?.closest<HTMLElement>('[role="row"]');
      if (semanticRow === undefined || semanticRow === null) return;
      semanticRow.dataset.cartularyGridDraftRow = "true";
      if (row.testId === undefined) {
        semanticRow.removeAttribute("data-testid");
      } else {
        semanticRow.dataset.testid = row.testId;
      }
    },
  };
}

function stopVendorNavigationForInteractiveContent(
  event: KeyboardEvent<HTMLSpanElement>,
) {
  if (
    event.target !== event.currentTarget &&
    (event.key.startsWith("Arrow") ||
      event.key === "Enter" ||
      event.key === "Tab")
  ) {
    event.stopPropagation();
  }
}

function alignmentClass(align: GridColumn<unknown>["align"]) {
  return align === undefined ? undefined : `cartulary-grid-cell-${align}`;
}

function dataCellClass(align: GridColumn<unknown>["align"]) {
  return ["cartulary-grid-data-cell", alignmentClass(align)]
    .filter((value) => value !== undefined)
    .join(" ");
}

function markSemanticGutterCell(
  node: HTMLSpanElement | null,
  testId: string | undefined,
) {
  const semanticCell = node?.closest<HTMLElement>('[role="gridcell"]');
  if (semanticCell === undefined || semanticCell === null) return;
  semanticCell.dataset.gridFieldKey = gridRowGutterColumnKey;
  if (testId === undefined) {
    semanticCell.removeAttribute("data-testid");
  } else {
    semanticCell.dataset.testid = testId;
  }
}
