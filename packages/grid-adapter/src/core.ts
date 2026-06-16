import type {
  CSSProperties,
  MouseEventHandler,
  PropsWithChildren,
  ReactNode,
} from "react";

export type GridSortDirection = "asc" | "desc";

export type GridDensity = "compact" | "default" | "comfortable";

export type GridSortEntry = {
  readonly fieldKey: string;
  readonly direction: GridSortDirection;
};

export type GridColumn<Row> = {
  readonly contractWritable?: boolean | undefined;
  readonly editorAdapter?: GridEditorAdapter<Row> | undefined;
  readonly fieldKey: string;
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: Row) => ReactNode;
  readonly rendererAdapter?: GridRendererAdapter<Row> | undefined;
  readonly sortableFieldKey?: string | null;
  readonly sortDisabled?: boolean | undefined;
  readonly sortDisabledReason?: string | null | undefined;
  readonly valueKind?: string | undefined;
  readonly align?: "left" | "center" | "right" | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridAdapterCleanup = () => void;

export type GridEditorAdapter<Row> = {
  readonly cleanup?: GridAdapterCleanup | undefined;
  readonly renderEditor: (row: Row) => ReactNode;
};

export type GridRendererAdapter<Row> = {
  readonly cleanup?: GridAdapterCleanup | undefined;
  readonly renderCell: (row: Row) => ReactNode;
};

export type GridRow<Row> = {
  readonly key: string;
  readonly recordId: string | null;
  readonly data: Row;
  readonly gutterContent?: ReactNode | undefined;
  readonly gutterLabel?: string | undefined;
  readonly gutterTestId?: string | undefined;
  readonly onSelect?: MouseEventHandler<HTMLTableRowElement> | undefined;
  readonly selected?: boolean | undefined;
  readonly variant?: "default" | "draft" | undefined;
  readonly testId?: string | undefined;
};

export type GridRowGutter = {
  readonly headerTestId?: string | undefined;
  readonly label?: ReactNode | undefined;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridActionsColumn<Row> = {
  readonly headerTestId?: string | undefined;
  readonly label: string;
  readonly renderCell: (row: GridRow<Row>) => ReactNode;
  readonly minWidth?: number | undefined;
  readonly width?: number | undefined;
};

export type GridChrome = "sheet" | "framed";

export type GridViewportProps = PropsWithChildren<{
  readonly className?: string | undefined;
  readonly chrome?: GridChrome | undefined;
  readonly style?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}>;

export type GridTableProps<Row> = {
  readonly actionsColumn?: GridActionsColumn<Row> | undefined;
  readonly columns: readonly GridColumn<Row>[];
  readonly density?: GridDensity | undefined;
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
  readonly rowGutter?: GridRowGutter | undefined;
  readonly rows: readonly GridRow<Row>[];
  readonly sort?: readonly GridSortEntry[] | undefined;
};

export type GridPresentationGroupRow = {
  readonly groupBy: string;
  readonly groupLabel: string | null;
  readonly key: string;
  readonly kind: "group";
  readonly testId?: string | undefined;
};

export type GridPresentationDataRow<Row> = {
  readonly gridRow: GridRow<Row>;
  readonly key: string;
  readonly kind: "data";
};

export type GridPresentationRow<Row> =
  | GridPresentationGroupRow
  | GridPresentationDataRow<Row>;

export type GridCellAnchor = {
  readonly fieldKey: string;
  readonly recordId: string;
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
};

export type GridPasteRecordRowTarget = {
  readonly kind: "record";
  readonly recordId: string;
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

export type GridRendererRegistry<Row> = {
  readonly fallbackRenderer: GridRendererAdapter<Row>;
  readonly fieldRenderers?:
    | ReadonlyMap<string, GridRendererAdapter<Row>>
    | Readonly<Record<string, GridRendererAdapter<Row> | undefined>>
    | undefined;
  readonly valueKindRenderers?:
    | ReadonlyMap<string, GridRendererAdapter<Row>>
    | Readonly<Record<string, GridRendererAdapter<Row> | undefined>>
    | undefined;
};

export type ResolveGridRendererProps<Row> = {
  readonly column: GridColumn<Row>;
  readonly registry: GridRendererRegistry<Row>;
};

type BuildGridPresentationRowsProps<Row> = {
  readonly getGroupLabel?:
    | ((row: Row, fieldKey: string) => string | null | undefined)
    | undefined;
  readonly getGroupRowTestId?:
    | ((fieldKey: string, value: string) => string | undefined)
    | undefined;
  readonly groupBy?: string | null | undefined;
  readonly rows: readonly GridRow<Row>[];
};

export type RecordIdentity = {
  readonly recordId: string | null;
};

export const gridUnassignedGroupLabel = "Unassigned";

export function isGridColumnEditable<Row>(column: GridColumn<Row>): boolean {
  return column.contractWritable === true && column.editorAdapter !== undefined;
}

export function resolveGridRenderer<Row>({
  column,
  registry,
}: ResolveGridRendererProps<Row>): GridRendererAdapter<Row> {
  return (
    column.rendererAdapter ??
    lookupGridAdapter(registry.fieldRenderers, column.fieldKey) ??
    lookupGridAdapter(registry.valueKindRenderers, column.valueKind) ??
    registry.fallbackRenderer
  );
}

export function cleanupGridAdapters<Row>(
  adapters: readonly (
    | GridColumn<Row>
    | GridEditorAdapter<Row>
    | GridRendererAdapter<Row>
    | null
    | undefined
  )[],
) {
  const cleanups = new Set<GridAdapterCleanup>();
  for (const adapter of adapters) {
    if (adapter === null || adapter === undefined) {
      continue;
    }
    if ("cleanup" in adapter && adapter.cleanup !== undefined) {
      cleanups.add(adapter.cleanup);
      continue;
    }
    if ("editorAdapter" in adapter && adapter.editorAdapter?.cleanup) {
      cleanups.add(adapter.editorAdapter.cleanup);
    }
    if ("rendererAdapter" in adapter && adapter.rendererAdapter?.cleanup) {
      cleanups.add(adapter.rendererAdapter.cleanup);
    }
  }
  for (const cleanup of cleanups) {
    cleanup();
  }
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

export function buildGridPresentationRows<Row>({
  getGroupLabel,
  getGroupRowTestId,
  groupBy,
  rows,
}: BuildGridPresentationRowsProps<Row>): readonly GridPresentationRow<Row>[] {
  if (
    groupBy === null ||
    groupBy === undefined ||
    getGroupLabel === undefined
  ) {
    return rows.map((row) => ({
      gridRow: row,
      key: row.key,
      kind: "data",
    }));
  }

  const buckets: Array<{
    groupKeyValue: string;
    groupLabel: string | null;
    rows: Array<GridRow<Row>>;
  }> = [];
  const bucketsByKey = new Map<
    string,
    {
      groupKeyValue: string;
      groupLabel: string | null;
      rows: Array<GridRow<Row>>;
    }
  >();
  const recordlessRows: Array<GridRow<Row>> = [];

  for (const row of rows) {
    if (row.recordId === null) {
      recordlessRows.push(row);
      continue;
    }
    const nextGroupLabel = normalizeGroupLabel(
      getGroupLabel(row.data, groupBy),
    );
    const bucketMapKey =
      nextGroupLabel === null ? "group:null" : `group:value:${nextGroupLabel}`;
    let bucket = bucketsByKey.get(bucketMapKey);
    if (bucket === undefined) {
      bucket = {
        groupKeyValue: nextGroupLabel ?? "empty",
        groupLabel: nextGroupLabel,
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
        bucket.groupLabel === null || getGroupRowTestId === undefined
          ? undefined
          : getGroupRowTestId(groupBy, bucket.groupLabel),
    });
    for (const row of bucket.rows) {
      presentationRows.push({
        gridRow: row,
        key: row.key,
        kind: "data",
      });
    }
  }

  for (const row of recordlessRows) {
    presentationRows.push({
      gridRow: row,
      key: row.key,
      kind: "data",
    });
  }

  return presentationRows;
}

export function resolveGridCellAnchor<Row>({
  columns,
  presentationRows,
  selection,
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
  if (recordId === null || recordId.trim() === "") {
    return null;
  }
  return {
    fieldKey: selection.fieldKey,
    recordId,
  };
}

export function navigateGridCellAnchor<Row>({
  columns,
  current,
  intent,
  presentationRows,
}: NavigateGridCellAnchorProps<Row>): GridCellAnchor | null {
  const currentRowIndex = presentationRows.findIndex(
    (row) =>
      row.kind === "data" &&
      row.gridRow.recordId !== null &&
      row.gridRow.recordId === current.recordId,
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
    (row) =>
      row.kind === "data" &&
      row.gridRow.recordId !== null &&
      row.gridRow.recordId === current.recordId,
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
      rowTargets.push({ createIndex, kind: "create" });
      createIndex += 1;
      continue;
    }
    if (presentationRow.kind !== "data") {
      return null;
    }
    const recordId = presentationRow.gridRow.recordId;
    if (recordId === null || recordId.trim() === "") {
      if (!allowCreateRows) {
        return null;
      }
      rowTargets.push({ createIndex, kind: "create" });
      createIndex += 1;
      continue;
    }
    rowTargets.push({ kind: "record", recordId });
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

function lookupGridAdapter<Row>(
  registry:
    | ReadonlyMap<string, GridRendererAdapter<Row>>
    | Readonly<Record<string, GridRendererAdapter<Row> | undefined>>
    | undefined,
  key: string | undefined,
): GridRendererAdapter<Row> | undefined {
  if (registry === undefined || key === undefined || key.trim() === "") {
    return undefined;
  }
  if (
    typeof (registry as ReadonlyMap<string, GridRendererAdapter<Row>>).get ===
    "function"
  ) {
    return (registry as ReadonlyMap<string, GridRendererAdapter<Row>>).get(key);
  }
  return (
    registry as Readonly<Record<string, GridRendererAdapter<Row> | undefined>>
  )[key];
}

function normalizeGroupLabel(value: string | null | undefined): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}
