import type {
  GridDataRow,
  GridDraftRow,
  GridGroupingScalar,
} from "@cartulary/grid-adapter";
import {
  gridRowGutterTestId,
  gridRowTestId,
  workbookInlineDraftRowTestId,
} from "@cartulary/ui-contracts";
import type { ReactNode } from "react";
import { compareWorkbookGroupValues } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { inputFocusKey } from "./timelineFieldRegistry";
import { createDraftRow, type WorkbookRow } from "./timelineRowModel";
import type { WorkbookVersionedRecord } from "./workbookRecordFreshness";

type TimelineGridRows = {
  readonly draftRow?: GridDraftRow<WorkbookRow> | undefined;
  readonly recordRows: readonly GridDataRow<WorkbookRow>[];
};

/** Selection context is committed presentation, independent of local authoring. */
export function timelineRecordSelectionPresentation(row: WorkbookRow): {
  readonly label: string;
  readonly description: string;
} {
  const normalize = (text: string) => text.replace(/\s+/gu, " ").trim();
  const synopsis = normalize(row.committedValues.activitySynopsisText);
  const utc = normalize(row.committedValues.activityUTCText);
  const local = normalize(row.committedValues.activityLocalText);
  const time = utc || (local ? `${local} (local time)` : "");
  const characters = Array.from(synopsis);
  const shortened = characters.length > 120;
  const conciseSynopsis = shortened
    ? `${characters.slice(0, 119).join("")}…`
    : synopsis;
  const context = (text: string) =>
    time
      ? `${text || "No synopsis"} — ${time}`
      : text || "No synopsis or activity time";
  return {
    label: `Select Timeline record: ${context(conciseSynopsis)}`,
    description: `${shortened ? `${context(synopsis)}. ` : ""}Record ID: ${row.recordId}`,
  };
}

export function compareTimelineGroupValues(
  field: string,
  left: GridGroupingScalar,
  right: GridGroupingScalar,
): number {
  if (left === null || right === null)
    return compareWorkbookGroupValues(left, right);
  const order =
    field === "timeline.capture_state"
      ? ["rough", "enriched", "reviewed", "superseded"]
      : field === "timeline.activity_time_pair_state"
        ? [
            "paired_generated",
            "paired_user_preserved",
            "paired_mismatch",
            "conversion_unavailable",
            "disabled",
            "empty",
          ]
        : null;
  return order
    ? order.indexOf(String(left)) - order.indexOf(String(right))
    : -compareWorkbookGroupValues(left, right);
}

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
  readonly presenceForRow: (recordId: string | null) => TPresence;
  readonly renderDraftGutterContent: (row: WorkbookRow) => ReactNode;
  readonly renderSavedGutterContent: (input: {
    readonly ordinal: string;
    readonly presences: TPresence;
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
