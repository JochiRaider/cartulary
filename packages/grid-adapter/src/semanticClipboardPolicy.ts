import type {
  GridCellAnchor,
  GridCellCopyIntent,
  GridCellPasteIntent,
  GridCellRange,
  GridClipboardInput,
  GridColumn,
  GridDataRow,
  GridFillIntent,
  GridSurfaceIdentity,
} from "./core";
import { formatGridClipboardTSV, gridRowIdentitiesEqual } from "./core";
import {
  dedupeGridTargets,
  type GridSemanticPresentationModel,
  gridClipboardInputDimensions,
  planSemanticPasteTargets,
  resolveVisibleGridCellRange,
  semanticTarget,
} from "./semanticPresentation";

export function planSemanticCopy<Row>({
  anchor,
  columns,
  dataRows,
  model,
  range,
}: {
  readonly anchor: GridCellAnchor;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly model: GridSemanticPresentationModel<Row>;
  readonly range: GridCellRange;
}): { readonly intent: GridCellCopyIntent; readonly text: string } | null {
  const expandedRange = resolveVisibleGridCellRange({
    columns,
    dataRows,
    positionMap: model,
    range,
  });
  if (expandedRange === null) return null;
  const values = expandedRange.rowIdentities.map((rowIdentity) => {
    const row = dataRows.find((candidate) =>
      gridRowIdentitiesEqual(candidate.rowIdentity, rowIdentity),
    );
    return expandedRange.fieldKeys.map((fieldKey) => {
      const column = columns.find(
        (candidate) => candidate.fieldKey === fieldKey,
      );
      return row === undefined
        ? ""
        : (column?.getClipboardValue?.(row.data) ?? "");
    });
  });
  return {
    intent: { anchor, expandedRange, range },
    text: formatGridClipboardTSV(values),
  };
}

export function planSemanticPaste<Row>({
  input,
  model,
  target,
}: {
  readonly input: GridClipboardInput;
  readonly model: GridSemanticPresentationModel<Row>;
  readonly target: ReturnType<typeof semanticTarget<Row>>;
}): GridCellPasteIntent | null {
  if (target === null) return null;
  const dimensions = gridClipboardInputDimensions(input);
  if (dimensions === null) return null;
  const targetResolution = planSemanticPasteTargets(model, target, dimensions);
  if (targetResolution === null) return null;
  const lastFieldKey = targetResolution.columns.at(-1);
  const lastRecordTarget = targetResolution.rowTargets
    .filter((candidate) => candidate.kind === "record")
    .at(-1);
  const range: GridCellRange =
    lastFieldKey === undefined || lastRecordTarget === undefined
      ? { start: target, end: target }
      : {
          start: target,
          end: {
            fieldKey: lastFieldKey,
            rowIdentity: lastRecordTarget.rowIdentity,
            surface: target.surface,
          },
        };
  return { input, range, target, targetResolution };
}

export function planSemanticFill<Row>({
  columnKey,
  columns,
  dataRows,
  model,
  sourceRow,
  surface,
  targetRow,
}: {
  readonly columnKey: string;
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly model: GridSemanticPresentationModel<Row>;
  readonly sourceRow: GridDataRow<Row>;
  readonly surface: GridSurfaceIdentity;
  readonly targetRow: GridDataRow<Row>;
}): GridFillIntent | null {
  const source = semanticTarget(sourceRow, columnKey, columns, surface);
  const target = semanticTarget(targetRow, columnKey, columns, surface);
  const column = columns.find((candidate) => candidate.fieldKey === columnKey);
  if (
    source === null ||
    target === null ||
    column?.contractWritable !== true ||
    column.editor === undefined ||
    column.valueKind === "collection"
  ) {
    return null;
  }
  const range: GridCellRange = {
    start: {
      fieldKey: source.fieldKey,
      rowIdentity: source.rowIdentity,
      surface: source.surface,
    },
    end: {
      fieldKey: target.fieldKey,
      rowIdentity: target.rowIdentity,
      surface: target.surface,
    },
  };
  const expanded = resolveVisibleGridCellRange({
    columns,
    dataRows,
    positionMap: model,
    range,
  });
  if (expanded === null || expanded.fieldKeys.length !== 1) return null;
  const targets = expanded.rowIdentities.flatMap((rowIdentity) => {
    if (gridRowIdentitiesEqual(rowIdentity, source.rowIdentity)) return [];
    const row = dataRows.find((candidate) =>
      gridRowIdentitiesEqual(candidate.rowIdentity, rowIdentity),
    );
    return row?.mutationIdentity === undefined
      ? []
      : [
          {
            fieldKey: columnKey,
            mutationIdentity: row.mutationIdentity,
            rowIdentity,
            surface,
          },
        ];
  });
  if (
    targets.length === 0 ||
    targets.length !== expanded.rowIdentities.length - 1
  ) {
    return null;
  }
  return { range, source, target, targets: dedupeGridTargets(targets) };
}

export function planSemanticFillFromRange<Row>({
  columns,
  dataRows,
  model,
  range,
  surface,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly dataRows: readonly GridDataRow<Row>[];
  readonly model: GridSemanticPresentationModel<Row>;
  readonly range: GridCellRange | null;
  readonly surface: GridSurfaceIdentity;
}): GridFillIntent | null {
  if (range === null) return null;
  const expanded = resolveVisibleGridCellRange({
    columns,
    dataRows,
    positionMap: model,
    range,
  });
  const fieldKey = expanded?.fieldKeys[0];
  const first = expanded?.rowIdentities[0];
  const last = expanded?.rowIdentities.at(-1);
  if (
    expanded === null ||
    expanded.fieldKeys.length !== 1 ||
    expanded.rowIdentities.length < 2 ||
    fieldKey === undefined ||
    first === undefined ||
    last === undefined
  ) {
    return null;
  }
  const sourceRow = findRow(dataRows, first);
  const targetRow = findRow(dataRows, last);
  return sourceRow === undefined || targetRow === undefined
    ? null
    : planSemanticFill({
        columnKey: fieldKey,
        columns,
        dataRows,
        model,
        sourceRow,
        surface,
        targetRow,
      });
}

export function mergeSemanticFillIntents(
  pending: GridFillIntent | null,
  next: GridFillIntent,
): GridFillIntent {
  if (
    pending === null ||
    !gridRowIdentitiesEqual(
      pending.source.rowIdentity,
      next.source.rowIdentity,
    ) ||
    pending.source.fieldKey !== next.source.fieldKey
  ) {
    return next;
  }
  return {
    ...next,
    targets: dedupeGridTargets([...pending.targets, ...next.targets]),
  };
}

function findRow<Row>(
  rows: readonly GridDataRow<Row>[],
  identity: GridDataRow<Row>["rowIdentity"],
): GridDataRow<Row> | undefined {
  return rows.find((row) => gridRowIdentitiesEqual(row.rowIdentity, identity));
}
