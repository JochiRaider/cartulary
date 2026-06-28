import type { WorkbookRow } from "./workbookTimelineModel";

export type TimelineInspectorHistorySubject =
  | {
      readonly kind: "live";
      readonly recordId: string;
      readonly rowVersion: number | null;
    }
  | {
      readonly kind: "deleted";
      readonly recordId: string;
      readonly rowVersion: number;
    }
  | { readonly kind: "draft" }
  | { readonly kind: "none" };

export type TimelineHistoryStateLike = {
  readonly recordId: string | null;
  readonly data: {
    readonly deleted?: boolean | undefined;
    readonly record_id: string;
    readonly row_version: number;
  } | null;
};

export function selectTimelineInspectorHistorySubject({
  draftRow,
  rowHistory,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly rowHistory: TimelineHistoryStateLike;
  readonly selectedRow: WorkbookRow | null;
}): TimelineInspectorHistorySubject {
  const matchedRowHistoryData =
    rowHistory.data !== null &&
    rowHistory.data.record_id === rowHistory.recordId
      ? rowHistory.data
      : null;
  const deletedRowHistoryData =
    matchedRowHistoryData?.deleted === true ? matchedRowHistoryData : null;
  const selectedLiveRecordId = selectedRow?.recordId ?? null;
  const deletedRowIsActiveSubject =
    deletedRowHistoryData !== null &&
    (selectedLiveRecordId === null ||
      selectedLiveRecordId === deletedRowHistoryData.record_id);
  if (deletedRowIsActiveSubject && deletedRowHistoryData !== null) {
    return {
      kind: "deleted",
      recordId: deletedRowHistoryData.record_id,
      rowVersion: deletedRowHistoryData.row_version,
    };
  }
  if (selectedLiveRecordId !== null) {
    return {
      kind: "live",
      recordId: selectedLiveRecordId,
      rowVersion: selectedRow?.rowVersion ?? null,
    };
  }
  if (draftRow !== null) {
    return { kind: "draft" };
  }
  return { kind: "none" };
}
