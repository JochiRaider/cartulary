import { requireViewContract } from "@cartulary/view-contracts";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import {
  type WorkbookRecordHistoryState,
  workbookRecordHistoryLoadedData,
} from "../../inspector/workbookRecordHistoryModel";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookRow } from "./workbookTimelineModel";

const timelineInspectorConfig =
  requireViewContract(timelineViewSchemaId).inspectorConfig;

export function selectTimelineInspectorHistorySubject({
  draftRow,
  rowHistory,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly rowHistory: WorkbookRecordHistoryState;
  readonly selectedRow: WorkbookRow | null;
}): WorkbookInspectorSubject | null {
  const rowHistoryData = workbookRecordHistoryLoadedData(rowHistory);
  const matchedRowHistoryData =
    rowHistoryData !== null &&
    rowHistoryData.record_id === rowHistory.subject?.recordId
      ? rowHistoryData
      : null;
  const deletedRowHistoryData =
    matchedRowHistoryData?.deleted === true ? matchedRowHistoryData : null;
  const selectedLiveRecordId = selectedRow?.recordId ?? null;
  const deletedRowIsActiveSubject =
    deletedRowHistoryData !== null &&
    (selectedLiveRecordId === null ||
      selectedLiveRecordId === deletedRowHistoryData.record_id);
  if (deletedRowIsActiveSubject && deletedRowHistoryData !== null) {
    return buildWorkbookInspectorSubject({
      config: timelineInspectorConfig,
      kind: "deleted",
      label: "Deleted timeline row",
      recordId: deletedRowHistoryData.record_id,
      rowVersion: deletedRowHistoryData.row_version,
      surfaceLabel: "Timeline",
    });
  }
  if (selectedLiveRecordId !== null) {
    return buildWorkbookInspectorSubject({
      config: timelineInspectorConfig,
      kind: "live",
      label:
        selectedRow?.values.activitySynopsisText.trim() ||
        "Selected timeline row",
      recordId: selectedLiveRecordId,
      rowVersion: selectedRow?.rowVersion,
      surfaceLabel: "Timeline",
    });
  }
  if (draftRow !== null) return null;
  return null;
}
