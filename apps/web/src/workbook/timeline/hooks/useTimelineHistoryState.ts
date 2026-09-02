import { useCallback, useReducer, useRef } from "react";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import {
  initialWorkbookRecordHistoryState,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryOperationId,
  type WorkbookRecordHistoryRequestId,
  workbookRecordHistoryOperationId,
  workbookRecordHistoryPendingAction,
  workbookRecordHistoryReducer,
  workbookRecordHistoryRequestId,
} from "../../inspector/workbookRecordHistoryModel";
import { selectTimelineInspectorHistorySubject } from "../models/timelineHistoryModel";
import type { WorkbookRow } from "../models/workbookTimelineModel";

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
  const currentHistoryRecordIdRef = useRef<string | null>(null);
  const rowHistoryRequestSeqRef = useRef(0);
  const rowHistoryOperationSeqRef = useRef(0);
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
  currentHistoryRecordIdRef.current = currentHistoryRecordId;
  const currentHistoryRowVersion = inspectorHistorySubject?.rowVersion ?? null;
  const currentHistoryDeleted = inspectorHistorySubject?.kind === "deleted";
  const activeHistoryLiveRecordId =
    inspectorHistorySubject?.kind === "live"
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
    sendRowHistoryEvent({ type: "clear" });
  }, [cancelRowHistoryRequests, sendRowHistoryEvent]);
  const retargetRowHistory = useCallback(
    (subject: WorkbookInspectorSubject | null) => {
      sendRowHistoryEvent({ subject, type: "retarget" });
    },
    [sendRowHistoryEvent],
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
      rowHistoryPendingAction: workbookRecordHistoryPendingAction(rowHistory),
    },
  };
}
