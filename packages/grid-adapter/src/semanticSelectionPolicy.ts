import type { GridDataRow, GridSortEntry } from "./core";

export type SemanticBulkSelectionState<Row> = {
  readonly allSelected: boolean;
  readonly partiallySelected: boolean;
  readonly selectableIds: readonly string[];
  readonly selectableRows: readonly GridDataRow<Row>[];
};

export function resolveSemanticBulkSelection<Row>(
  rows: readonly GridDataRow<Row>[],
  selectedRecordIds: ReadonlySet<string>,
  isRecordSelectable?: (row: GridDataRow<Row>) => boolean,
): SemanticBulkSelectionState<Row> {
  const selectableRows = rows.filter(
    (row) =>
      row.rowIdentity.kind === "core_record" &&
      isRecordSelectable?.(row) !== false,
  );
  const selectableIds = selectableRows.flatMap((row) =>
    row.rowIdentity.kind === "core_record" ? [row.rowIdentity.recordId] : [],
  );
  const selectedCount = selectableIds.filter((id) =>
    selectedRecordIds.has(id),
  ).length;
  return {
    allSelected:
      selectableIds.length > 0 && selectedCount === selectableIds.length,
    partiallySelected:
      selectedCount > 0 && selectedCount < selectableIds.length,
    selectableIds,
    selectableRows,
  };
}

export function toggleAllSemanticRecords(
  state: SemanticBulkSelectionState<unknown>,
): ReadonlySet<string> {
  return state.allSelected ? new Set() : new Set(state.selectableIds);
}

export function toggleSemanticRecordRange<Row>({
  anchorRecordId,
  recordId,
  selectableRows,
  selectedRecordIds,
  shiftKey,
}: {
  readonly anchorRecordId: string | null;
  readonly recordId: string;
  readonly selectableRows: readonly GridDataRow<Row>[];
  readonly selectedRecordIds: ReadonlySet<string>;
  readonly shiftKey: boolean;
}): ReadonlySet<string> {
  const next = new Set(selectedRecordIds);
  const recordIds = selectableRows.flatMap((row) =>
    row.rowIdentity.kind === "core_record" ? [row.rowIdentity.recordId] : [],
  );
  const anchorIndex =
    anchorRecordId === null ? -1 : recordIds.indexOf(anchorRecordId);
  const recordIndex = recordIds.indexOf(recordId);
  if (recordIndex < 0) return selectedRecordIds;
  if (shiftKey && anchorIndex >= 0 && recordIndex >= 0) {
    const start = Math.min(anchorIndex, recordIndex);
    const end = Math.max(anchorIndex, recordIndex);
    for (const id of recordIds.slice(start, end + 1)) next.add(id);
    return next;
  }
  if (next.has(recordId)) next.delete(recordId);
  else next.add(recordId);
  return next;
}

export function nextSemanticSort(
  current: readonly GridSortEntry[],
  fieldKey: string,
  additive: boolean,
): readonly GridSortEntry[] {
  const currentIndex = current.findIndex(
    (entry) => entry.fieldKey === fieldKey,
  );
  const currentEntry = current[currentIndex];
  const nextEntry =
    currentEntry === undefined
      ? { fieldKey, direction: "asc" as const }
      : currentEntry.direction === "asc"
        ? { fieldKey, direction: "desc" as const }
        : null;
  const retained = additive
    ? current.filter((entry) => entry.fieldKey !== fieldKey)
    : [];
  return nextEntry === null ? retained : [...retained, nextEntry];
}
