import type { CSSProperties, PropsWithChildren, ReactNode } from "react";

export type GridSortDirection = "asc" | "desc";

export type GridDensity = "compact" | "default" | "comfortable";

export type GridSortEntry = {
  readonly fieldKey: string;
  readonly direction: GridSortDirection;
};

export type GridColumn<Row> = {
  readonly contractWritable?: boolean | undefined;
  readonly fieldKey: string;
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (context: GridCellRenderContext<Row>) => ReactNode;
  readonly renderDraftCell?:
    | ((context: GridDraftCellRenderContext<Row>) => ReactNode)
    | undefined;
  readonly renderEditCell?:
    | ((context: GridCellEditorContext<Row>) => ReactNode)
    | undefined;
  readonly sortableFieldKey?: string | null;
  readonly sortDisabled?: boolean | undefined;
  readonly sortDisabledReason?: string | null | undefined;
  readonly valueKind?: string | undefined;
  readonly align?: "left" | "center" | "right" | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridRecordRow<Row> = {
  readonly kind: "record";
  readonly recordId: string;
  readonly rowVersion: number;
  readonly data: Row;
  readonly gutterContent?: ReactNode | undefined;
  readonly gutterLabel?: string | undefined;
  readonly gutterTestId?: string | undefined;
  readonly selected?: boolean | undefined;
  readonly testId?: string | undefined;
};

export type GridDraftRow<Row> = {
  readonly kind: "draft";
  readonly data: Row;
  readonly gutterContent?: ReactNode | undefined;
  readonly gutterLabel?: string | undefined;
  readonly testId?: string | undefined;
};

export type GridSemanticRow<Row> = GridDraftRow<Row> | GridRecordRow<Row>;

export type GridRowGutter = {
  readonly headerTestId?: string | undefined;
  readonly label?: ReactNode | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridActionsColumn<Row> = {
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: GridRecordRow<Row>) => ReactNode;
  readonly renderDraftCell?:
    | ((row: GridDraftRow<Row>) => ReactNode)
    | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridChrome = "sheet" | "framed";
export type GridBlockSizing = "standalone" | "fill";

export type GridViewportProps = PropsWithChildren<{
  readonly blockSizing?: GridBlockSizing | undefined;
  readonly className?: string | undefined;
  readonly chrome?: GridChrome | undefined;
  readonly style?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}>;

export type WorkbookDataGridProps<Row> = {
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly columnWidths?: Readonly<Record<string, number>> | undefined;
  readonly density?: GridDensity | undefined;
  readonly draftRow?: GridDraftRow<Row> | undefined;
  readonly emptyMessage?: ReactNode | undefined;
  readonly fillViewportInline?: boolean | undefined;
  readonly grouping?: GridGroupingDescriptor<Row> | null | undefined;
  readonly onActiveCellChange?:
    | ((anchor: GridCellAnchor | null) => void)
    | undefined;
  readonly onCopyCell?: ((intent: GridCellCopyIntent) => void) | undefined;
  readonly onEditCell?: ((intent: GridCellMutationIntent) => void) | undefined;
  readonly onFillCells?: ((intent: GridFillIntent) => void) | undefined;
  readonly onPasteCell?: ((intent: GridCellMutationIntent) => void) | undefined;
  readonly onToggleSort?: ((fieldKey: string) => void) | undefined;
  readonly onSortChange?:
    | ((sort: readonly GridSortEntry[]) => void)
    | undefined;
  readonly onColumnWidthChange?:
    | ((fieldKey: string, width: number) => void)
    | undefined;
  readonly onSelectRecord?: ((recordId: string) => void) | undefined;
  readonly onSelectedRecordIdsChange?:
    | ((recordIds: ReadonlySet<string>) => void)
    | undefined;
  readonly recordRows: readonly GridRecordRow<Row>[];
  readonly rowGutter?: GridRowGutter | undefined;
  readonly sort?: readonly GridSortEntry[] | undefined;
  readonly selectedRecordIds?: ReadonlySet<string> | undefined;
  readonly viewSchemaId: string;
};

export type GridPresentationGroupRow = {
  readonly groupBy: string;
  readonly groupLabel: string | null;
  readonly key: string;
  readonly kind: "group";
  readonly testId?: string | undefined;
};

export type GridPresentationDataRow<Row> = {
  readonly gridRow: GridRecordRow<Row>;
  readonly key: string;
  readonly kind: "data";
};

export type GridPresentationRow<Row> =
  | GridPresentationGroupRow
  | GridPresentationDataRow<Row>;

export type GridCellAnchor = {
  readonly fieldKey: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export type GridCellTarget = GridCellAnchor & {
  readonly baseRowVersion: number;
};

export type GridCellRenderContext<Row> = {
  readonly anchor: GridCellAnchor;
  readonly row: Row;
};

export type GridDraftCellRenderContext<Row> = {
  readonly fieldKey: string;
  readonly row: Row;
  readonly viewSchemaId: string;
};

export type GridCellEditorContext<Row> = {
  readonly closeEditor: (commit: boolean) => void;
  readonly row: Row;
  readonly target: GridCellTarget;
  readonly updateRow: (row: Row) => void;
};

export type GridGroupingScalar = boolean | number | string | null;

export type GridGroupingDescriptor<Row> = {
  readonly fieldKey: string;
  readonly label?: string | undefined;
  readonly getValue: (row: Row) => GridGroupingScalar;
  readonly formatLabel: (value: GridGroupingScalar) => string | null;
  readonly getTestId?:
    | ((
        fieldKey: string,
        value: GridGroupingScalar,
        label: string | null,
      ) => string | undefined)
    | undefined;
};

export type GridHandle = {
  readonly focusAnchor: (anchor: GridCellAnchor) => boolean;
  readonly getScrollElement: () => HTMLDivElement | null;
  readonly scrollToAnchor: (anchor: GridCellAnchor) => boolean;
};

export type GridCellCopyIntent = {
  readonly anchor: GridCellAnchor;
};

export type GridCellMutationIntent = {
  readonly target: GridCellTarget;
};

export type GridFillIntent = {
  readonly source: GridCellTarget;
  readonly target: GridCellTarget;
};

export type GridCellSelection = {
  readonly fieldKey: string;
  readonly rowIndex: number;
};

export type GridNavigationKey =
  | "ArrowDown"
  | "ArrowLeft"
  | "ArrowRight"
  | "ArrowUp"
  | "Enter"
  | "Tab";

export type GridNavigationIntent = {
  readonly key: GridNavigationKey;
  readonly shiftKey?: boolean | undefined;
};

export type ResolveGridCellAnchorProps<Row> = {
  readonly columns: readonly GridColumn<Row>[];
  readonly presentationRows: readonly GridPresentationRow<Row>[];
  readonly selection: GridCellSelection;
  readonly viewSchemaId: string;
};

export type NavigateGridCellAnchorProps<Row> = {
  readonly columns: readonly GridColumn<Row>[];
  readonly current: GridCellAnchor;
  readonly intent: GridNavigationIntent;
  readonly presentationRows: readonly GridPresentationRow<Row>[];
};

export type GridPasteCreateRowTarget = {
  readonly createIndex: number;
  readonly kind: "create";
  readonly viewSchemaId: string;
};

export type GridPasteRecordRowTarget = {
  readonly baseRowVersion: number;
  readonly kind: "record";
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export type GridPasteRowTarget =
  | GridPasteCreateRowTarget
  | GridPasteRecordRowTarget;

export type ResolveGridPasteTargetsProps<Row> = {
  readonly allowCreateRows?: boolean | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly current: GridCellAnchor;
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly presentationRows: readonly GridPresentationRow<Row>[];
};

export type GridPasteTargetResolution = {
  readonly columns: readonly string[];
  readonly rowTargets: readonly GridPasteRowTarget[];
};

type BuildGridPresentationRowsProps<Row> = {
  readonly grouping?: GridGroupingDescriptor<Row> | null | undefined;
  readonly rows: readonly GridSemanticRow<Row>[];
};

export type RecordIdentity = {
  readonly recordId: string | null;
};

export const gridUnassignedGroupLabel = "Unassigned";

export function isGridColumnEditable<Row>(column: GridColumn<Row>): boolean {
  return (
    column.contractWritable === true && column.renderEditCell !== undefined
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

export function buildGridPresentationRows<Row>({
  grouping,
  rows,
}: BuildGridPresentationRowsProps<Row>): readonly GridPresentationRow<Row>[] {
  if (grouping === null || grouping === undefined) {
    return rows.flatMap((row) =>
      row.kind === "draft"
        ? []
        : [{ gridRow: row, key: row.recordId, kind: "data" as const }],
    );
  }

  const groupBy = grouping.fieldKey;
  const buckets: Array<{
    groupValue: GridGroupingScalar;
    groupKeyValue: string;
    groupLabel: string | null;
    rows: Array<GridRecordRow<Row>>;
  }> = [];
  const bucketsByKey = new Map<
    string,
    {
      groupKeyValue: string;
      groupLabel: string | null;
      groupValue: GridGroupingScalar;
      rows: Array<GridRecordRow<Row>>;
    }
  >();
  for (const row of rows) {
    if (row.kind === "draft") {
      continue;
    }
    const groupValue = grouping.getValue(row.data);
    const nextGroupLabel = normalizeGroupLabel(
      grouping.formatLabel(groupValue),
    );
    const bucketMapKey = encodeGridGroupingScalar(groupValue);
    let bucket = bucketsByKey.get(bucketMapKey);
    if (bucket === undefined) {
      bucket = {
        groupKeyValue: bucketMapKey,
        groupLabel: nextGroupLabel,
        groupValue,
        rows: [],
      };
      bucketsByKey.set(bucketMapKey, bucket);
      buckets.push(bucket);
    }
    bucket.rows.push(row);
  }

  const presentationRows: GridPresentationRow<Row>[] = [];
  for (const bucket of buckets) {
    presentationRows.push({
      groupBy,
      groupLabel: bucket.groupLabel,
      key: `group:${groupBy}:${bucket.groupKeyValue}:0`,
      kind: "group",
      testId:
        grouping.getTestId === undefined
          ? undefined
          : grouping.getTestId(groupBy, bucket.groupValue, bucket.groupLabel),
    });
    for (const row of bucket.rows) {
      presentationRows.push({
        gridRow: row,
        key: row.recordId,
        kind: "data",
      });
    }
  }

  return presentationRows;
}

export function resolveGridCellAnchor<Row>({
  columns,
  presentationRows,
  selection,
  viewSchemaId,
}: ResolveGridCellAnchorProps<Row>): GridCellAnchor | null {
  if (
    !Number.isInteger(selection.rowIndex) ||
    selection.rowIndex < 0 ||
    !columns.some((column) => column.fieldKey === selection.fieldKey)
  ) {
    return null;
  }
  const row = presentationRows[selection.rowIndex];
  if (row === undefined || row.kind !== "data") {
    return null;
  }
  const recordId = row.gridRow.recordId;
  if (recordId.trim() === "") {
    return null;
  }
  return {
    fieldKey: selection.fieldKey,
    recordId,
    viewSchemaId,
  };
}

export function navigateGridCellAnchor<Row>({
  columns,
  current,
  intent,
  presentationRows,
}: NavigateGridCellAnchorProps<Row>): GridCellAnchor | null {
  const currentRowIndex = presentationRows.findIndex(
    (row) => row.kind === "data" && row.gridRow.recordId === current.recordId,
  );
  const currentColumnIndex = columns.findIndex(
    (column) => column.fieldKey === current.fieldKey,
  );
  if (currentRowIndex < 0 || currentColumnIndex < 0) {
    return null;
  }

  const target = navigateGridCellCoordinates({
    columnIndex: currentColumnIndex,
    columnCount: columns.length,
    intent,
    rowIndex: currentRowIndex,
    rowCount: presentationRows.length,
  });
  if (target === null) {
    return null;
  }
  const targetColumn = columns[target.columnIndex];
  if (targetColumn === undefined) {
    return null;
  }
  return resolveGridCellAnchor({
    columns,
    presentationRows,
    selection: {
      fieldKey: targetColumn.fieldKey,
      rowIndex: target.rowIndex,
    },
    viewSchemaId: current.viewSchemaId,
  });
}

export function resolveGridPasteTargets<Row>({
  allowCreateRows = true,
  columns,
  current,
  pastedColumnCount,
  pastedRowCount,
  presentationRows,
}: ResolveGridPasteTargetsProps<Row>): GridPasteTargetResolution | null {
  if (
    current.recordId.trim() === "" ||
    current.fieldKey.trim() === "" ||
    !Number.isInteger(pastedRowCount) ||
    pastedRowCount < 1 ||
    !Number.isInteger(pastedColumnCount) ||
    pastedColumnCount < 1
  ) {
    return null;
  }
  const startColumnIndex = columns.findIndex(
    (column) => column.fieldKey === current.fieldKey,
  );
  if (startColumnIndex < 0) {
    return null;
  }
  const targetColumns = columns
    .slice(startColumnIndex, startColumnIndex + pastedColumnCount)
    .map((column) => column.fieldKey);
  if (targetColumns.length === 0) {
    return null;
  }

  const startRowIndex = presentationRows.findIndex(
    (row) => row.kind === "data" && row.gridRow.recordId === current.recordId,
  );
  if (startRowIndex < 0) {
    return null;
  }

  const rowTargets: GridPasteRowTarget[] = [];
  let createIndex = 0;
  for (let offset = 0; offset < pastedRowCount; offset += 1) {
    const presentationRow = presentationRows[startRowIndex + offset];
    if (presentationRow === undefined) {
      if (!allowCreateRows) {
        return null;
      }
      rowTargets.push({
        createIndex,
        kind: "create",
        viewSchemaId: current.viewSchemaId,
      });
      createIndex += 1;
      continue;
    }
    if (presentationRow.kind !== "data") {
      return null;
    }
    const recordId = presentationRow.gridRow.recordId;
    if (recordId.trim() === "") {
      return null;
    }
    rowTargets.push({
      baseRowVersion: presentationRow.gridRow.rowVersion,
      kind: "record",
      recordId,
      viewSchemaId: current.viewSchemaId,
    });
  }

  return {
    columns: targetColumns,
    rowTargets,
  };
}

function navigateGridCellCoordinates({
  columnCount,
  columnIndex,
  intent,
  rowCount,
  rowIndex,
}: {
  readonly columnCount: number;
  readonly columnIndex: number;
  readonly intent: GridNavigationIntent;
  readonly rowCount: number;
  readonly rowIndex: number;
}): { columnIndex: number; rowIndex: number } | null {
  let nextColumnIndex = columnIndex;
  let nextRowIndex = rowIndex;
  switch (intent.key) {
    case "ArrowDown":
      nextRowIndex += 1;
      break;
    case "ArrowUp":
      nextRowIndex -= 1;
      break;
    case "ArrowLeft":
      nextColumnIndex -= 1;
      break;
    case "ArrowRight":
      nextColumnIndex += 1;
      break;
    case "Enter":
      nextRowIndex += intent.shiftKey === true ? -1 : 1;
      break;
    case "Tab":
      nextColumnIndex += intent.shiftKey === true ? -1 : 1;
      break;
  }
  if (
    nextRowIndex < 0 ||
    nextRowIndex >= rowCount ||
    nextColumnIndex < 0 ||
    nextColumnIndex >= columnCount
  ) {
    return null;
  }
  return {
    columnIndex: nextColumnIndex,
    rowIndex: nextRowIndex,
  };
}

function normalizeGroupLabel(value: string | null | undefined): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}

function encodeGridGroupingScalar(value: GridGroupingScalar): string {
  if (value === null) {
    return "n:null";
  }
  if (typeof value === "boolean") {
    return value ? "b:true" : "b:false";
  }
  if (typeof value === "number") {
    if (!Number.isFinite(value)) {
      throw new Error("Grid grouping values must be finite numbers.");
    }
    return `d:${String(value)}`;
  }
  return `s:${value}`;
}
