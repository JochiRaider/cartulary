import { useCallback, useReducer, useRef } from "react";
import {
  initialWorkbookRecordHistoryState,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryOperationId,
  type WorkbookRecordHistoryRequestId,
  type WorkbookRecordHistorySubject,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "../../inspector/workbookRecordHistoryModel";
import {
  selectTimelineInspectorHistorySubject,
  type TimelineInspectorHistorySubject,
} from "../models/timelineHistoryModel";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type { TimelineInspectorHistorySubject };

export function useTimelineHistoryState({
  draftRow,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly selectedRow: WorkbookRow | null;
}) {
  const [rowHistory, dispatchRowHistory] = useReducer(
    workbookRecordHistoryReducer,
    null,
    initialWorkbookRecordHistoryState,
  );
  const currentHistoryRecordIdRef = useRef<string | null>(null);
  const rowHistoryRequestSeqRef = useRef(0);
  const rowHistoryOperationSeqRef = useRef(0);

  const inspectorHistorySubject = selectTimelineInspectorHistorySubject({
    draftRow,
    rowHistory,
    selectedRow,
  });
  const currentHistoryRecordId =
    inspectorHistorySubject.kind === "live" ||
    inspectorHistorySubject.kind === "deleted"
      ? inspectorHistorySubject.recordId
      : null;
  currentHistoryRecordIdRef.current = currentHistoryRecordId;
  const currentHistoryRowVersion =
    inspectorHistorySubject.kind === "live" ||
    inspectorHistorySubject.kind === "deleted"
      ? inspectorHistorySubject.rowVersion
      : null;
  const currentHistoryDeleted = inspectorHistorySubject.kind === "deleted";
  const activeHistoryLiveRecordId =
    inspectorHistorySubject.kind === "live"
      ? inspectorHistorySubject.recordId
      : null;

  const beginRowHistoryRequest =
    useCallback((): WorkbookRecordHistoryRequestId => {
      const requestId = workbookRecordHistoryRequestId(
        rowHistoryRequestSeqRef.current + 1,
      );
      rowHistoryRequestSeqRef.current = requestId.value;
      return requestId;
    }, []);
  const beginRowHistoryOperation =
    useCallback((): WorkbookRecordHistoryOperationId => {
      const operationId = workbookRecordHistoryOperationId(
        rowHistoryOperationSeqRef.current + 1,
      );
      rowHistoryOperationSeqRef.current = operationId.value;
      return operationId;
    }, []);
  const cancelRowHistoryRequests = useCallback(() => {
    rowHistoryRequestSeqRef.current += 1;
  }, []);
  const rowHistoryRequestIsCurrent = useCallback(
    (requestId: WorkbookRecordHistoryRequestId) =>
      rowHistoryRequestSeqRef.current === requestId.value,
    [],
  );
  const currentHistoryRecordIdMatches = useCallback(
    (recordId: string) => currentHistoryRecordIdRef.current === recordId,
    [],
  );
  const clearRowHistory = useCallback(() => {
    cancelRowHistoryRequests();
    dispatchRowHistory({ type: "clear" });
  }, [cancelRowHistoryRequests]);
  const retargetRowHistory = useCallback(
    (subject: WorkbookRecordHistorySubject | null) => {
      dispatchRowHistory({ subject, type: "retarget" });
    },
    [],
  );
  const sendRowHistoryEvent = useCallback(
    (event: WorkbookRecordHistoryEvent) => dispatchRowHistory(event),
    [],
  );

  return {
    commands: {
      beginRowHistoryOperation,
      beginRowHistoryRequest,
      cancelRowHistoryRequests,
      clearRowHistory,
      currentHistoryRecordIdMatches,
      dispatchRowHistory: sendRowHistoryEvent,
      retargetRowHistory,
      rowHistoryRequestIsCurrent,
    },
    snapshot: {
      activeHistoryLiveRecordId,
      currentHistoryDeleted,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      inspectorHistorySubject,
      rowHistory,
      rowHistoryPendingAction: rowHistory.pendingAction,
    },
  };
}
