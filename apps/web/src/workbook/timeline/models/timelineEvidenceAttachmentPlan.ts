import type { TimelineFileSource } from "../../features/evidence/timelineFileOperation";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineEvidenceActionContext = {
  readonly authorized: boolean;
  readonly capabilityAvailable: boolean;
  readonly selectedRowKey: string | null;
  readonly surfaceKey: string;
};

/** An interaction identity, never an Inspector subject or a visual row position. */
export type TimelineFileTarget =
  | { readonly kind: "record"; readonly recordId: string }
  | { readonly kind: "row"; readonly key: string }
  | { readonly kind: "unavailable" }
  | { readonly kind: "ambiguous" };

export type TimelineFileTargetResolution =
  | { readonly kind: "resolved"; readonly source: TimelineFileSource }
  | { readonly kind: "missing" | "unavailable" | "ambiguous" };

export const timelineFileTargetInstruction =
  "Select the Timeline row or draft for this file.";

export function captureTimelineFileSource(
  row: WorkbookRow,
): TimelineFileSource {
  return {
    key: row.key,
    recordId: row.recordId,
    rowVersion: row.rowVersion,
    label:
      row.values.activitySynopsisText ||
      row.values.rawActivityText ||
      "Timeline draft",
  };
}

/** Resolve once at the gesture boundary. Known-but-missing targets never fall back. */
export function resolveTimelineFileTarget(
  target: TimelineFileTarget | null,
  rows: readonly WorkbookRow[],
  resolveRowKey: (key: string) => string,
): TimelineFileTargetResolution {
  if (target === null) return { kind: "missing" };
  if (target.kind === "unavailable" || target.kind === "ambiguous")
    return target;
  const matches = rows.filter((row) =>
    target.kind === "record"
      ? row.recordId === target.recordId
      : row.key === resolveRowKey(target.key),
  );
  if (matches.length > 1) return { kind: "ambiguous" };
  const row = matches[0];
  if (
    !row ||
    row.viewSchemaId !== timelineViewSchemaId ||
    row.captureState === "superseded" ||
    (row.recordId === null) !== (row.rowVersion === null)
  )
    return { kind: "unavailable" };
  return { kind: "resolved", source: captureTimelineFileSource(row) };
}
