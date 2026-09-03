import type {
  GridCellAnchor,
  GridCellRange,
  GridCellTarget,
  GridClipboardDimensions,
  GridClipboardInput,
  GridColumn,
  GridDataRow,
  GridGroupingDescriptor,
  GridGroupingScalar,
  GridNavigationIntent,
  GridPasteRowTarget,
  GridPasteTargetResolution,
  GridRowIdentity,
  GridSurfaceIdentity,
  SemanticDataGridProps,
} from "./core";
import {
  gridRowIdentitiesEqual,
  gridRowIdentityKey,
  gridSurfaceIdentitiesEqual,
  gridSurfaceIdentityKey,
} from "./core";

export type GridSemanticCoordinateModel = {
  readonly fieldKeys: readonly string[];
  readonly rowIdentities: readonly GridRowIdentity[];
  readonly surface: GridSurfaceIdentity;
};

export type GridSemanticPresentationModel<Row> = GridSemanticCoordinateModel & {
  readonly allowCreateRows: boolean;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly grouping: GridSemanticGroupingModel<Row> | null;
};

export type GridSemanticGroupBucket<Row> = {
  readonly id: string;
  readonly label: string | null;
  readonly rows: readonly GridDataRow<Row>[];
  readonly value: GridGroupingScalar;
};

export type GridSemanticGroupingModel<Row> = {
  readonly buckets: readonly GridSemanticGroupBucket<Row>[];
  readonly collapsedGroupIds: ReadonlySet<string>;
  readonly scope: string;
};

export function gridAnchorKey(anchor: GridCellAnchor): string {
  return `${gridSurfaceIdentityKey(anchor.surface)}\u0000${gridRowIdentityKey(anchor.rowIdentity)}\u0000${anchor.fieldKey}`;
}

export function gridClipboardInputDimensions(
  input: GridClipboardInput,
): GridClipboardDimensions | null {
  if (input.kind === "scalar") {
    return { columnCount: 1, rowCount: 1 };
  }
  const columnCount = input.values[0]?.length ?? 0;
  if (
    input.values.length < 1 ||
    columnCount < 1 ||
    input.values.some((row) => row.length !== columnCount)
  ) {
    return null;
  }
  return { columnCount, rowCount: input.values.length };
}

export function coreRecordId<Row>(row: GridDataRow<Row>): string | null {
  return row.rowIdentity.kind === "core_record"
    ? row.rowIdentity.recordId
    : null;
}

export function coreRowVersion<Row>(row: GridDataRow<Row>): number | null {
  return row.mutationIdentity?.kind === "core_row_version"
    ? row.mutationIdentity.baseRowVersion
    : null;
}

export function isCoreRecordRow<Row>(
  row: GridDataRow<Row>,
): row is GridDataRow<Row> & {
  readonly rowIdentity: Extract<
    GridRowIdentity,
    { readonly kind: "core_record" }
  >;
} {
  return row.rowIdentity.kind === "core_record";
}

export function emptySemanticCoordinateModel(
  surface: GridSurfaceIdentity,
): GridSemanticCoordinateModel {
  return {
    fieldKeys: [],
    rowIdentities: [],
    surface,
  };
}

export function emptySemanticPresentationModel<Row>(
  surface: GridSurfaceIdentity,
): GridSemanticPresentationModel<Row> {
  return {
    ...emptySemanticCoordinateModel(surface),
    allowCreateRows: false,
    columns: [],
    dataRows: [],
    grouping: null,
  };
}

export function sameGridCellAnchor(
  left: GridCellAnchor | null,
  right: GridCellAnchor | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.fieldKey === right.fieldKey &&
      gridRowIdentitiesEqual(left.rowIdentity, right.rowIdentity) &&
      gridSurfaceIdentitiesEqual(left.surface, right.surface))
  );
}

export function sameGridCellRange(
  left: GridCellRange | null,
  right: GridCellRange | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      sameGridCellAnchor(left.start, right.start) &&
      sameGridCellAnchor(left.end, right.end))
  );
}

export function semanticPresentationContainsAnchor(
  model: GridSemanticCoordinateModel,
  anchor: GridCellAnchor,
): boolean {
  return (
    gridSurfaceIdentitiesEqual(model.surface, anchor.surface) &&
    model.fieldKeys.includes(anchor.fieldKey) &&
    model.rowIdentities.some((identity) =>
      gridRowIdentitiesEqual(identity, anchor.rowIdentity),
    )
  );
}

export function buildSemanticCoordinateModel<Row>({
  dataRows,
  fieldKeys,
  surface,
}: {
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly fieldKeys: readonly string[];
  readonly surface: GridSurfaceIdentity;
}): GridSemanticCoordinateModel {
  return {
    fieldKeys,
    rowIdentities: dataRows.map((row) => row.rowIdentity),
    surface,
  };
}

export function buildSemanticPresentationModel<Row>({
  allowCreateRows,
  columns,
  dataRows,
  fieldKeys,
  grouping = null,
  surface,
}: {
  readonly allowCreateRows: boolean;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly fieldKeys: readonly string[];
  readonly grouping?: GridSemanticGroupingModel<Row> | null | undefined;
  readonly surface: GridSurfaceIdentity;
}): GridSemanticPresentationModel<Row> {
  return {
    ...buildSemanticCoordinateModel({
      dataRows,
      fieldKeys,
      surface,
    }),
    allowCreateRows,
    columns,
    dataRows,
    grouping,
  };
}

export function buildSemanticGroupBuckets<Row>(
  dataRows: readonly GridDataRow<Row>[],
  grouping: GridGroupingDescriptor<Row>,
): readonly GridSemanticGroupBucket<Row>[] {
  const buckets: Array<{
    readonly id: string;
    readonly label: string | null;
    readonly rows: Array<GridDataRow<Row>>;
    readonly value: GridGroupingScalar;
  }> = [];
  const bucketsById = new Map<string, (typeof buckets)[number]>();
  for (const row of dataRows) {
    const value = grouping.getValue(row.data);
    const id = encodeGroupValue(value);
    const existing = bucketsById.get(id);
    if (existing !== undefined) {
      existing.rows.push(row);
      continue;
    }
    const bucket = {
      id,
      label: grouping.formatLabel(value),
      rows: [row],
      value,
    };
    bucketsById.set(id, bucket);
    buckets.push(bucket);
  }
  return buckets;
}

export function moveSemanticAnchor(
  positionMap: GridSemanticCoordinateModel,
  anchor: GridCellAnchor,
  columnDelta: number,
  rowDelta: number,
): GridCellAnchor | null {
  const columnIndex = positionMap.fieldKeys.indexOf(anchor.fieldKey);
  const rowIndex = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  const fieldKey = positionMap.fieldKeys[columnIndex + columnDelta];
  const rowIdentity = positionMap.rowIdentities[rowIndex + rowDelta];
  if (
    columnIndex < 0 ||
    rowIndex < 0 ||
    fieldKey === undefined ||
    rowIdentity === undefined
  ) {
    return null;
  }
  return { fieldKey, rowIdentity, surface: positionMap.surface };
}

export function navigateSemanticPresentation<Row>(
  model: GridSemanticPresentationModel<Row>,
  current: GridCellAnchor,
  intent: GridNavigationIntent,
): GridCellAnchor | null {
  if (!gridSurfaceIdentitiesEqual(current.surface, model.surface)) return null;
  const columnIndex = model.fieldKeys.indexOf(current.fieldKey);
  const rowIndex = model.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, current.rowIdentity),
  );
  if (columnIndex < 0 || rowIndex < 0) return null;

  let nextColumnIndex = columnIndex;
  let nextRowIndex = rowIndex;
  const pageSize = Math.max(1, Math.floor(intent.pageSize ?? 1));
  switch (intent.key) {
    case "ArrowDown":
      nextRowIndex += 1;
      break;
    case "ArrowLeft":
      nextColumnIndex -= 1;
      break;
    case "ArrowRight":
      nextColumnIndex += 1;
      break;
    case "ArrowUp":
      nextRowIndex -= 1;
      break;
    case "PageDown":
      nextRowIndex = Math.min(
        model.rowIdentities.length - 1,
        rowIndex + pageSize,
      );
      break;
    case "PageUp":
      nextRowIndex = Math.max(0, rowIndex - pageSize);
      break;
    case "Home":
      nextColumnIndex = 0;
      if (intent.ctrlOrMetaKey === true) nextRowIndex = 0;
      break;
    case "End":
      nextColumnIndex = model.fieldKeys.length - 1;
      if (intent.ctrlOrMetaKey === true) {
        nextRowIndex = model.rowIdentities.length - 1;
      }
      break;
    case "Enter":
      nextRowIndex += intent.shiftKey === true ? -1 : 1;
      break;
    case "Tab":
      nextColumnIndex += intent.shiftKey === true ? -1 : 1;
      break;
  }
  const fieldKey = model.fieldKeys[nextColumnIndex];
  const rowIdentity = model.rowIdentities[nextRowIndex];
  if (
    nextColumnIndex < 0 ||
    nextRowIndex < 0 ||
    fieldKey === undefined ||
    rowIdentity === undefined
  ) {
    return null;
  }
  return { fieldKey, rowIdentity, surface: model.surface };
}

export function planSemanticPasteTargets<Row>(
  model: GridSemanticPresentationModel<Row>,
  current: GridCellAnchor,
  dimensions: GridClipboardDimensions,
): GridPasteTargetResolution | null {
  if (
    current.rowIdentity.kind !== "core_record" ||
    current.surface.kind !== "view_schema" ||
    !gridSurfaceIdentitiesEqual(current.surface, model.surface) ||
    current.rowIdentity.recordId.trim() === "" ||
    !Number.isInteger(dimensions.columnCount) ||
    dimensions.columnCount < 1 ||
    !Number.isInteger(dimensions.rowCount) ||
    dimensions.rowCount < 1
  ) {
    return null;
  }
  const startColumnIndex = model.fieldKeys.indexOf(current.fieldKey);
  const startRowIndex = model.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, current.rowIdentity),
  );
  if (startColumnIndex < 0 || startRowIndex < 0) return null;
  const targetColumns = model.fieldKeys.slice(
    startColumnIndex,
    startColumnIndex + dimensions.columnCount,
  );
  if (
    targetColumns.length !== dimensions.columnCount ||
    !targetColumns.every((fieldKey) =>
      model.columns.some(
        (column) =>
          column.fieldKey === fieldKey &&
          column.contractWritable === true &&
          column.editor !== undefined,
      ),
    )
  ) {
    return null;
  }

  const rowTargets: GridPasteRowTarget[] = [];
  let createIndex = 0;
  for (let offset = 0; offset < dimensions.rowCount; offset += 1) {
    const rowIdentity = model.rowIdentities[startRowIndex + offset];
    if (rowIdentity === undefined) {
      if (!model.allowCreateRows) return null;
      rowTargets.push({
        createIndex,
        kind: "create",
        surface: current.surface,
      });
      createIndex += 1;
      continue;
    }
    const row = model.dataRows.find((candidate) =>
      gridRowIdentitiesEqual(candidate.rowIdentity, rowIdentity),
    );
    if (
      row === undefined ||
      row.rowIdentity.kind !== "core_record" ||
      row.mutationIdentity === undefined
    ) {
      return null;
    }
    rowTargets.push({
      kind: "record",
      mutationIdentity: row.mutationIdentity,
      rowIdentity: row.rowIdentity,
      surface: current.surface,
    });
  }
  return { columns: targetColumns, rowTargets };
}

export function resolveVisibleGridCellRange<Row>({
  columns,
  dataRows,
  positionMap,
  range,
}: {
  readonly columns: readonly SemanticDataGridProps<Row>["columns"][number][];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly positionMap: GridSemanticCoordinateModel;
  readonly range: GridCellRange;
}) {
  if (
    !gridSurfaceIdentitiesEqual(range.start.surface, positionMap.surface) ||
    !gridSurfaceIdentitiesEqual(range.end.surface, positionMap.surface)
  ) {
    return null;
  }
  const startColumn = positionMap.fieldKeys.indexOf(range.start.fieldKey);
  const endColumn = positionMap.fieldKeys.indexOf(range.end.fieldKey);
  const startRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.start.rowIdentity),
  );
  const endRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.end.rowIdentity),
  );
  if (startColumn < 0 || endColumn < 0 || startRow < 0 || endRow < 0) {
    return null;
  }
  const fieldKeys = positionMap.fieldKeys.slice(
    Math.min(startColumn, endColumn),
    Math.max(startColumn, endColumn) + 1,
  );
  if (
    !fieldKeys.every((fieldKey) =>
      columns.some((column) => column.fieldKey === fieldKey),
    )
  ) {
    return null;
  }
  const rowIdentities = positionMap.rowIdentities.slice(
    Math.min(startRow, endRow),
    Math.max(startRow, endRow) + 1,
  );
  if (
    rowIdentities.some(
      (rowIdentity) =>
        !dataRows.some((row) =>
          gridRowIdentitiesEqual(row.rowIdentity, rowIdentity),
        ),
    )
  ) {
    return null;
  }
  return { fieldKeys, rowIdentities };
}

export function gridCellRangeContains(
  positionMap: GridSemanticCoordinateModel,
  range: GridCellRange | null,
  anchor: GridCellAnchor,
): boolean {
  if (range === null) return false;
  if (sameGridCellAnchor(range.start, range.end)) return false;
  const startColumn = positionMap.fieldKeys.indexOf(range.start.fieldKey);
  const endColumn = positionMap.fieldKeys.indexOf(range.end.fieldKey);
  const column = positionMap.fieldKeys.indexOf(anchor.fieldKey);
  const startRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.start.rowIdentity),
  );
  const endRow = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, range.end.rowIdentity),
  );
  const row = positionMap.rowIdentities.findIndex((identity) =>
    gridRowIdentitiesEqual(identity, anchor.rowIdentity),
  );
  if (
    startColumn < 0 ||
    endColumn < 0 ||
    column < 0 ||
    startRow < 0 ||
    endRow < 0 ||
    row < 0
  ) {
    return false;
  }
  return (
    column >= Math.min(startColumn, endColumn) &&
    column <= Math.max(startColumn, endColumn) &&
    row >= Math.min(startRow, endRow) &&
    row <= Math.max(startRow, endRow)
  );
}

export function semanticAnchor<Row>(
  row: GridDataRow<Row>,
  fieldKey: string,
  columns: readonly SemanticDataGridProps<Row>["columns"][number][],
  surface: GridSurfaceIdentity,
): GridCellAnchor | null {
  if (!isDataRow(row)) return null;
  if (!columns.some((column) => column.fieldKey === fieldKey)) return null;
  return { fieldKey, rowIdentity: row.rowIdentity, surface };
}

export function semanticTarget<Row>(
  row: GridDataRow<Row>,
  fieldKey: string,
  columns: readonly SemanticDataGridProps<Row>["columns"][number][],
  surface: GridSurfaceIdentity,
): GridCellTarget | null {
  if (
    !isDataRow(row) ||
    surface.kind !== "view_schema" ||
    row.rowIdentity.kind !== "core_record" ||
    row.mutationIdentity === undefined
  ) {
    return null;
  }
  const column = columns.find((candidate) => candidate.fieldKey === fieldKey);
  if (column?.contractWritable !== true || column.editor === undefined) {
    return null;
  }
  return {
    fieldKey,
    mutationIdentity: row.mutationIdentity,
    rowIdentity: row.rowIdentity,
    surface,
  };
}

export function dedupeGridTargets(
  targets: readonly GridCellTarget[],
): readonly GridCellTarget[] {
  const seen = new Set<string>();
  return targets.filter((target) => {
    const key = gridAnchorKey(target);
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export function isDataRow<Row>(
  candidate: unknown,
): candidate is GridDataRow<Row> {
  if (typeof candidate !== "object" || candidate === null) return false;
  const row = candidate as Partial<GridDataRow<Row>>;
  return row.kind === "data" && row.rowIdentity !== undefined;
}

export function encodeGroupValue(value: GridGroupingScalar): string {
  if (value === null) return "n:null";
  if (typeof value === "boolean") return value ? "b:true" : "b:false";
  if (typeof value === "number") {
    if (!Number.isFinite(value)) {
      throw new Error("Grid grouping values must be finite numbers.");
    }
    return `d:${String(value)}`;
  }
  return `s:${value}`;
}
