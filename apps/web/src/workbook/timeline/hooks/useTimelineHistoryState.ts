import { useCallback, useReducer, useRef } from "react";
import {
  initialWorkbookRecordHistoryState,
  type WorkbookRecordHistoryEvent,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
} from "../../inspector/workbookRecordHistoryModel";
import { selectTimelineInspectorHistorySubject } from "../models/timelineHistoryModel";
import type { WorkbookRow } from "../models/timelineRowModel";

export function useTimelineHistoryState({
  draftRow,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly selectedRow: WorkbookRow | null;
}) {
  const [rowHistory, reactDispatchRowHistory] = useReducer(
    workbookRecordHistoryReducer,
    null,
    initialWorkbookRecordHistoryState,
  );
  const rowHistoryRef = useRef(rowHistory);
  rowHistoryRef.current = rowHistory;
  const sendRowHistoryEvent = useCallback(
    (event: WorkbookRecordHistoryEvent) => {
      const next = workbookRecordHistoryReducer(rowHistoryRef.current, event);
      rowHistoryRef.current = next;
      reactDispatchRowHistory(event);
      return next;
    },
    [],
  );

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
