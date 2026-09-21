import { useCallback } from "react";
import { useWorkbookRecordHistoryState } from "../../inspector/useWorkbookRecordHistoryState";
import { workbookRecordHistoryPendingAction } from "../../inspector/workbookRecordHistoryModel";
import { selectTimelineInspectorHistorySubject } from "../models/timelineHistoryModel";
import type { WorkbookRow } from "../models/timelineRowModel";

export function useTimelineHistoryState({
  draftRow,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly selectedRow: WorkbookRow | null;
}) {
  const { snapshot: rowHistory, dispatch: sendRowHistoryEvent } =
    useWorkbookRecordHistoryState();

  const inspectorHistorySubject = selectTimelineInspectorHistorySubject({
    draftRow,
    rowHistory,
    selectedRow,
  });
  const currentHistoryRecordId = inspectorHistorySubject?.recordId ?? null;
  const currentHistoryRowVersion = inspectorHistorySubject?.rowVersion ?? null;
  const currentHistoryDeleted = inspectorHistorySubject?.kind === "deleted";
  const activeHistoryLiveRecordId =
    inspectorHistorySubject?.kind === "live"
      ? inspectorHistorySubject.recordId
      : null;

  const clearRowHistory = useCallback(() => {
    sendRowHistoryEvent({ type: "clear" });
  }, [sendRowHistoryEvent]);

  return {
    commands: {
      clearRowHistory,
      dispatchRowHistory: sendRowHistoryEvent,
    },
    snapshot: {
      activeHistoryLiveRecordId,
      currentHistoryDeleted,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      inspectorHistorySubject,
      rowHistory,
      rowHistoryPendingAction: workbookRecordHistoryPendingAction(rowHistory),
    },
  };
}
