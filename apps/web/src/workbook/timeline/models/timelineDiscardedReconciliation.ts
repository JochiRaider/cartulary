import type { PendingReplayUnitState } from "../../utils/workbookPendingQueue";
import type { TimelineReplayContext } from "./timelineControllerPorts";
import {
  isTimelineCollectionDraftKey,
  timelineFieldBinding,
  timelineScalarBindings,
} from "./timelineFieldRegistry";
import { createDraftRow, type WorkbookRow } from "./timelineRowModel";

export type TimelineDiscardedReconciliation = {
  readonly cancelEdit: {
    readonly fieldKey: string;
    readonly recordId: string;
  } | null;
  readonly discardedFocusKey: string | null;
  readonly remainingFocusKeys: ReadonlySet<string>;
  readonly rows: WorkbookRow[] | null;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function pendingChanges(unit: PendingReplayUnitState) {
  return Array.isArray(unit.payloadIntent.changes)
    ? unit.payloadIntent.changes.filter(isRecord)
    : [];
}

function restoreCommittedScalarCells(row: WorkbookRow): WorkbookRow {
  if (row.rawRow === null) return row;
  const committedScalarCells = Object.fromEntries(
    timelineScalarBindings.map((binding) => [
      binding.fieldKey,
      { value: row.committedValues[binding.key] },
    ]),
  );
  return {
    ...row,
    rawRow: {
      ...row.rawRow,
      cells: { ...row.rawRow.cells, ...committedScalarCells },
    },
  };
}

function applyRemainingUnit(
  row: WorkbookRow,
  unit: PendingReplayUnitState,
  context: TimelineReplayContext | undefined,
): WorkbookRow {
  let next = row;
  for (const change of pendingChanges(unit)) {
    if (typeof change.field_key !== "string") continue;
    const binding = timelineFieldBinding(change.field_key);
    if (binding.kind !== "scalar" || !("value" in change)) continue;
    const value = typeof change.value === "string" ? change.value : "";
    next = {
      ...next,
      values: { ...next.values, [binding.key]: value },
      rawRow:
        next.rawRow === null
          ? null
          : {
              ...next.rawRow,
              cells: {
                ...next.rawRow.cells,
                [binding.fieldKey]: { value: change.value },
              },
            },
    };
  }
  if (
    context === undefined ||
    !isTimelineCollectionDraftKey(context.focusField)
  ) {
    return next;
  }
  return {
    ...next,
    collectionDrafts: {
      ...next.collectionDrafts,
      [context.focusField]:
        context.rowSnapshot.collectionDrafts[context.focusField],
    },
  };
}

export function reconcileDiscardedTimelineUnit({
  committedRow,
  contextByUnitId,
  currentRows,
  discardedUnit,
  nextDraftIndex,
  remainingUnits,
}: {
  readonly committedRow: WorkbookRow | null;
  readonly contextByUnitId: ReadonlyMap<string, TimelineReplayContext>;
  readonly currentRows: readonly WorkbookRow[];
  readonly discardedUnit: PendingReplayUnitState;
  readonly nextDraftIndex: () => number;
  readonly remainingUnits: readonly PendingReplayUnitState[];
}): TimelineDiscardedReconciliation {
  const discardedContext = contextByUnitId.get(discardedUnit.id);
  const remainingForRow = remainingUnits
    .filter(
      (unit) =>
        unit.rowKey === discardedUnit.rowKey ||
        (discardedUnit.recordId !== null &&
          unit.recordId === discardedUnit.recordId),
    )
    .sort((left, right) => left.enqueueOrder - right.enqueueOrder);
  const remainingFocusKeys = new Set(
    remainingForRow.flatMap((unit) => {
      const focusKey = contextByUnitId.get(unit.id)?.focusKey;
      return focusKey === undefined ? [] : [focusKey];
    }),
  );
  const discardedFocusField = discardedContext?.focusField;
  const scalarBinding =
    discardedFocusField === undefined ||
    isTimelineCollectionDraftKey(discardedFocusField)
      ? null
      : (timelineScalarBindings.find(
          (binding) => binding.key === discardedFocusField,
        ) ?? null);
  const cancelEdit =
    discardedUnit.recordId === null || scalarBinding === null
      ? null
      : {
          fieldKey: scalarBinding.fieldKey,
          recordId: discardedUnit.recordId,
        };

  if (discardedUnit.kind === "create") {
    const rows = currentRows.filter((row) => row.key !== discardedUnit.rowKey);
    return {
      cancelEdit,
      discardedFocusKey: discardedContext?.focusKey ?? null,
      remainingFocusKeys,
      rows: rows.some((row) => row.recordId === null)
        ? rows
        : [...rows, createDraftRow(nextDraftIndex())],
    };
  }

  const currentRow = currentRows.find(
    (row) =>
      row.key === discardedUnit.rowKey ||
      row.recordId === discardedUnit.recordId,
  );
  const baseRow = committedRow ?? currentRow;
  if (baseRow === undefined) {
    return {
      cancelEdit,
      discardedFocusKey: discardedContext?.focusKey ?? null,
      remainingFocusKeys,
      rows: null,
    };
  }
  let reconciled = restoreCommittedScalarCells({
    ...baseRow,
    key: discardedUnit.rowKey,
    values: { ...baseRow.committedValues },
    pendingSignature: remainingForRow.at(-1)?.mutationSignature ?? null,
  });
  for (const unit of remainingForRow) {
    reconciled = applyRemainingUnit(
      reconciled,
      unit,
      contextByUnitId.get(unit.id),
    );
  }
  return {
    cancelEdit,
    discardedFocusKey: discardedContext?.focusKey ?? null,
    remainingFocusKeys,
    rows: currentRows.map((row) =>
      row.key === discardedUnit.rowKey ||
      row.recordId === discardedUnit.recordId
        ? reconciled
        : row,
    ),
  };
}
