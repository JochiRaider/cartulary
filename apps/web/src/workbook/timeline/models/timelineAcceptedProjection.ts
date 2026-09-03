import type { WorkbookRow } from "./timelineRowModel";
import { ensureTimelineDraftRow } from "./timelineRowsModel";

export type TimelineAcceptedProjection = {
  readonly createdFromDraft: boolean;
  readonly draftFocusKey: string | null;
  readonly previousRow: WorkbookRow | undefined;
  readonly rows: WorkbookRow[];
};

export function projectAcceptedTimelineRow({
  committed,
  currentRows,
  nextDraftIndex,
  rowKey,
}: {
  readonly committed: WorkbookRow;
  readonly currentRows: readonly WorkbookRow[];
  readonly nextDraftIndex: () => number;
  readonly rowKey: string;
}): TimelineAcceptedProjection {
  const previousRow =
    currentRows.find((candidate) => candidate.key === rowKey) ??
    currentRows.find((candidate) => candidate.recordId === committed.recordId);
  let replaced = false;
  let nextRows = currentRows.map((row) => {
    if (
      row.key !== rowKey &&
      (committed.recordId === null || row.recordId !== committed.recordId)
    ) {
      return row;
    }
    replaced = true;
    return committed;
  });
  if (!replaced) {
    const draftIndex = nextRows.findIndex((row) => row.recordId === null);
    nextRows =
      draftIndex === -1
        ? [...nextRows, committed]
        : [
            ...nextRows.slice(0, draftIndex),
            committed,
            ...nextRows.slice(draftIndex),
          ];
  }
  const hydrated = ensureTimelineDraftRow({ nextDraftIndex, rows: nextRows });
  return {
    createdFromDraft: previousRow?.recordId === null,
    draftFocusKey: hydrated.draftFocusKey,
    previousRow,
    rows: hydrated.rows,
  };
}
