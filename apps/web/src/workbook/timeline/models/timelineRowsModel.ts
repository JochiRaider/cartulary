import type { GridDataRow, GridDraftRow } from "@cartulary/grid-adapter";
import {
  gridRowGutterTestId,
  gridRowTestId,
  workbookInlineDraftRowTestId,
} from "@cartulary/ui-contracts";
import type { ReactNode } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookVersionedRecord } from "./workbookRecordFreshness";
import {
  createDraftRow,
  inputFocusKey,
  type WorkbookRow,
} from "./workbookTimelineModel";

type TimelineGridRows = {
  readonly draftRow?: GridDraftRow<WorkbookRow> | undefined;
  readonly recordRows: readonly GridDataRow<WorkbookRow>[];
};

type EnsureTimelineDraftRowResult = {
  readonly rows: WorkbookRow[];
  readonly draftFocusKey: string | null;
};

export function ensureTimelineDraftRow({
  nextDraftIndex,
  rows,
}: {
  readonly nextDraftIndex: () => number;
  readonly rows: WorkbookRow[];
}): EnsureTimelineDraftRowResult {
  if (rows.some((row) => row.recordId === null)) {
    return { rows, draftFocusKey: null };
  }
  const draftIndex = nextDraftIndex();
  return {
    rows: [...rows, createDraftRow(draftIndex)],
    draftFocusKey: inputFocusKey(`draft-${draftIndex}`, "activitySynopsisText"),
  };
}

function requireCommittedRowVersion(row: WorkbookVersionedRecord): number {
  if (row.rowVersion === null) {
    throw new Error("Committed Timeline grid row is missing row_version.");
  }
  return row.rowVersion;
}

export function buildTimelineGridRows<TPresence>({
  presenceForRow,
  renderDraftGutterContent,
  renderSavedGutterContent,
  rows,
}: {
  readonly presenceForRow: (recordId: string | null) => readonly TPresence[];
  readonly renderDraftGutterContent: (row: WorkbookRow) => ReactNode;
  readonly renderSavedGutterContent: (input: {
    readonly ordinal: string;
    readonly presences: readonly TPresence[];
    readonly recordId: string;
    readonly row: WorkbookRow;
  }) => ReactNode;
  readonly rows: readonly WorkbookRow[];
}): TimelineGridRows {
  const recordRows: GridDataRow<WorkbookRow>[] = [];
  let draftRow: GridDraftRow<WorkbookRow> | undefined;
  rows.forEach((row, index) => {
    const rowPresence = presenceForRow(row.recordId);
    const ordinal = row.recordId === null ? "+" : String(index + 1);
    if (row.recordId === null) {
      draftRow = {
        kind: "draft",
        data: row,
        gutterContent: renderDraftGutterContent(row),
        gutterLabel: ordinal,
        testId: workbookInlineDraftRowTestId(timelineViewSchemaId),
      };
      return;
    }
    const rowVersion = requireCommittedRowVersion(row);
    recordRows.push({
      kind: "data",
      mutationIdentity: {
        kind: "core_row_version",
        baseRowVersion: rowVersion,
      },
      rowIdentity: { kind: "core_record", recordId: row.recordId },
      data: row,
      gutterContent: renderSavedGutterContent({
        ordinal,
        presences: rowPresence,
        recordId: row.recordId,
        row,
      }),
      gutterLabel: ordinal,
      gutterTestId: gridRowGutterTestId(timelineViewSchemaId, row.recordId),
      testId: gridRowTestId(timelineViewSchemaId, row.recordId),
    });
  });
  return { draftRow, recordRows };
}
