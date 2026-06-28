import type { GridRow } from "@cartulary/grid-adapter";
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
  onSelectRow,
  presenceForRow,
  renderDraftGutterContent,
  renderSavedGutterContent,
  rows,
  selectedRowId,
}: {
  readonly onSelectRow: (recordId: string) => void;
  readonly presenceForRow: (recordId: string | null) => readonly TPresence[];
  readonly renderDraftGutterContent: (row: WorkbookRow) => ReactNode;
  readonly renderSavedGutterContent: (input: {
    readonly ordinal: string;
    readonly presences: readonly TPresence[];
    readonly recordId: string;
    readonly row: WorkbookRow;
  }) => ReactNode;
  readonly rows: readonly WorkbookRow[];
  readonly selectedRowId: string | null;
}): readonly GridRow<WorkbookRow>[] {
  return rows.map((row, index) => {
    const rowPresence = presenceForRow(row.recordId);
    const ordinal = row.recordId === null ? "+" : String(index + 1);
    return {
      key: row.key,
      recordId: row.recordId,
      data: row,
      gutterContent:
        row.recordId === null
          ? renderDraftGutterContent(row)
          : renderSavedGutterContent({
              ordinal,
              presences: rowPresence,
              recordId: row.recordId,
              row,
            }),
      gutterLabel: ordinal,
      gutterTestId:
        row.recordId === null
          ? undefined
          : gridRowGutterTestId(timelineViewSchemaId, row.recordId),
      onSelect: () => {
        if (row.recordId !== null) {
          onSelectRow(row.recordId);
        }
      },
      selected: row.recordId !== null && row.recordId === selectedRowId,
      testId:
        row.recordId === null
          ? workbookInlineDraftRowTestId(timelineViewSchemaId)
          : gridRowTestId(timelineViewSchemaId, row.recordId),
      variant: row.recordId === null ? "draft" : "default",
    };
  });
}
