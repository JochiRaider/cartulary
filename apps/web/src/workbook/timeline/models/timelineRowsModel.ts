import type { GridDataRow, GridDraftRow } from "@cartulary/grid-adapter";
import {
  gridRowGutterTestId,
  gridRowTestId,
  workbookInlineDraftRowTestId,
} from "@cartulary/ui-contracts";
import type { ReactNode } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./workbookTimelineModel";

export type WorkbookVersionedRecord = {
  readonly recordId: string | null;
  readonly rowVersion: number | null;
};

export type WorkbookRecordFreshnessDecision = {
  readonly comparable: boolean;
  readonly stale: boolean;
};

export type TimelineGridRows = {
  readonly draftRow?: GridDraftRow<WorkbookRow> | undefined;
  readonly recordRows: readonly GridDataRow<WorkbookRow>[];
};

export function decideWorkbookRecordFreshness(
  incoming: WorkbookVersionedRecord,
  knownRowVersion: number | null | undefined,
): WorkbookRecordFreshnessDecision {
  if (
    incoming.recordId === null ||
    incoming.rowVersion === null ||
    knownRowVersion === null ||
    knownRowVersion === undefined
  ) {
    return {
      comparable: false,
      stale: false,
    };
  }
  return {
    comparable: true,
    stale: incoming.rowVersion < knownRowVersion,
  };
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
    if (row.rowVersion === null) {
      throw new Error("Committed Timeline grid row is missing row_version.");
    }
    recordRows.push({
      kind: "data",
      mutationIdentity: {
        kind: "core_row_version",
        baseRowVersion: row.rowVersion,
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
