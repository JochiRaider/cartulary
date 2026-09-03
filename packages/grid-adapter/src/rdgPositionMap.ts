import type { GridDataRow, GridRowIdentity, GridSurfaceIdentity } from "./core";
import { gridRowIdentityKey } from "./core";
import {
  buildSemanticPresentationModel,
  emptySemanticPresentationModel,
  type GridSemanticGroupingModel,
  type GridSemanticPresentationModel,
  gridAnchorKey,
} from "./semanticPresentation";

export type GridRdgPosition = {
  readonly idx: number;
  readonly rowIdx: number;
};

export type GridRdgPositionMap = {
  readonly positions: ReadonlyMap<string, GridRdgPosition>;
  readonly surface: GridSurfaceIdentity;
};

export type GridRdgPresentationModel<Row> = GridSemanticPresentationModel<Row> &
  GridRdgPositionMap;

export function emptyRdgPresentationModel<Row>(
  surface: GridSurfaceIdentity,
): GridRdgPresentationModel<Row> {
  return {
    ...emptySemanticPresentationModel<Row>(surface),
    positions: new Map(),
  };
}

export function buildRdgPresentationModel<Row>({
  allowCreateRows,
  columns,
  columnKeys,
  dataRows,
  fieldKeys,
  grouping = null,
  rowIndexes,
  surface,
}: Parameters<typeof buildSemanticPresentationModel<Row>>[0] & {
  readonly columnKeys: readonly string[];
  readonly grouping?: GridSemanticGroupingModel<Row> | null | undefined;
  readonly rowIndexes?: ReadonlyMap<string, number> | undefined;
}): GridRdgPresentationModel<Row> {
  const model = buildSemanticPresentationModel({
    allowCreateRows,
    columns,
    dataRows,
    fieldKeys,
    grouping,
    surface,
  });
  return {
    ...model,
    positions: buildRdgPositions({
      columnKeys,
      dataRows,
      fieldKeys,
      rowIndexes,
      surface,
    }),
  };
}

function buildRdgPositions<Row>({
  columnKeys,
  dataRows,
  fieldKeys,
  rowIndexes,
  surface,
}: {
  readonly columnKeys: readonly string[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly fieldKeys: readonly string[];
  readonly rowIndexes?: ReadonlyMap<string, number> | undefined;
  readonly surface: GridSurfaceIdentity;
}): ReadonlyMap<string, GridRdgPosition> {
  const positions = new Map<string, GridRdgPosition>();
  const columnIndexes = new Map(
    columnKeys.map((columnKey, index) => [columnKey, index]),
  );
  for (const [fallbackRowIdx, row] of dataRows.entries()) {
    const rowIdx =
      rowIndexes?.get(gridRowIdentityKey(row.rowIdentity)) ?? fallbackRowIdx;
    if (rowIdx < 0) continue;
    registerRowPositions(
      positions,
      columnIndexes,
      row,
      fieldKeys,
      rowIdx,
      surface,
    );
  }
  return positions;
}

function registerRowPositions(
  positions: Map<string, GridRdgPosition>,
  columnIndexes: ReadonlyMap<string, number>,
  row: { readonly rowIdentity: GridRowIdentity },
  fieldKeys: readonly string[],
  rowIdx: number,
  surface: GridSurfaceIdentity,
): void {
  for (const fieldKey of fieldKeys) {
    const idx = columnIndexes.get(fieldKey);
    if (idx === undefined) continue;
    positions.set(
      gridAnchorKey({ fieldKey, rowIdentity: row.rowIdentity, surface }),
      { idx, rowIdx },
    );
  }
}
