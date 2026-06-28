import { useCallback, useRef, useState } from "react";
import type {
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../components/TimelineHistoryPanel";
import type { WorkbookRow } from "../models/workbookTimelineModel";

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

function emptyRowHistoryState(): RecordHistoryState {
  return {
    recordId: null,
    status: "idle",
    data: null,
    message: null,
  };
}

function selectInspectorHistorySubject({
  draftRow,
  rowHistory,
  selectedRow,
}: {
  readonly draftRow: WorkbookRow | null;
  readonly rowHistory: RecordHistoryState;
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

  const inspectorHistorySubject = selectInspectorHistorySubject({
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
