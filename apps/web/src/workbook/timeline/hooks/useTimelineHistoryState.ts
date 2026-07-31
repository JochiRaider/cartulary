import { useCallback, useRef, useState } from "react";
import type {
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";
import {
  selectTimelineInspectorHistorySubject,
  type TimelineInspectorHistorySubject,
} from "../models/timelineHistoryModel";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type { TimelineInspectorHistorySubject };

function emptyRowHistoryState(): RecordHistoryState {
  return {
    recordId: null,
    status: "idle",
    data: null,
    message: null,
  };
}

export function useTimelineHistoryState({
  draftRow,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly selectedRow: WorkbookRow | null;
}) {
  const [rowHistory, setRowHistory] =
    useState<RecordHistoryState>(emptyRowHistoryState);
  const [rowHistoryPendingAction, setRowHistoryPendingAction] =
    useState<RowHistoryPendingAction | null>(null);
  const currentHistoryRecordIdRef = useRef<string | null>(null);
  const rowHistoryRequestSeqRef = useRef(0);

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

  const beginRowHistoryRequest = useCallback(() => {
    const requestSeq = rowHistoryRequestSeqRef.current + 1;
    rowHistoryRequestSeqRef.current = requestSeq;
    return requestSeq;
  }, []);
  const cancelRowHistoryRequests = useCallback(() => {
    rowHistoryRequestSeqRef.current += 1;
  }, []);
  const rowHistoryRequestIsCurrent = useCallback(
    (requestSeq: number) => rowHistoryRequestSeqRef.current === requestSeq,
    [],
  );
  const currentHistoryRecordIdMatches = useCallback(
    (recordId: string) => currentHistoryRecordIdRef.current === recordId,
    [],
  );
  const clearRowHistory = useCallback(() => {
    cancelRowHistoryRequests();
    setRowHistoryPendingAction(null);
    setRowHistory(emptyRowHistoryState());
  }, [cancelRowHistoryRequests]);

  return {
    commands: {
      beginRowHistoryRequest,
      cancelRowHistoryRequests,
      clearRowHistory,
      currentHistoryRecordIdMatches,
      rowHistoryRequestIsCurrent,
      setRowHistory,
      setRowHistoryPendingAction,
    },
    snapshot: {
      activeHistoryLiveRecordId,
      currentHistoryDeleted,
      currentHistoryRecordId,
      currentHistoryRowVersion,
      inspectorHistorySubject,
      rowHistory,
      rowHistoryPendingAction,
    },
  };
}
