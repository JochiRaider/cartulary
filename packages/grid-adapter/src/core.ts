import type {
  CSSProperties,
  MouseEventHandler,
  PropsWithChildren,
  ReactNode,
} from "react";

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
  readonly headerTestId?: string | undefined;
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

function normalizeGroupLabel(value: string | null | undefined): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}
