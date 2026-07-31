import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
} from "react";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type {
  RecordHistoryData,
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../models/timelineHistoryModel";
import type {
  TimelineHistoryMutationAccepted,
  TimelineHistoryPort,
} from "../ports/TimelineHistoryPort";

type TimelineHistoryViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type TimelineHistoryLoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
  viewportContinuityToken?: number;
};

type TimelineCommittedRecordIdleResult = {
  row: unknown;
  rowVersion: number;
};

export function buildRecordRollbackTargetFromHistoryAction(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): Record<string, unknown> | null {
  if (!item.available_rollback_actions.includes(action)) {
    return null;
  }
  if (action === "history_entry") {
    return typeof item.history_entry_ref === "string" &&
      item.history_entry_ref.trim() !== ""
      ? { kind: "history_entry", history_entry_ref: item.history_entry_ref }
      : null;
  }
  if (action === "change_set") {
    return typeof item.change_set_id === "string" &&
      item.change_set_id.trim() !== ""
      ? { kind: "change_set", change_set_id: item.change_set_id }
      : null;
  }
  return isPositiveInteger(item.revision_no)
    ? { kind: "row_restore", restore_to_revision_no: item.revision_no }
    : null;
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

export function useTimelineHistoryActions({
  acceptTimelineRecordVersion,
  activeHistoryLiveRecordId,
  beginRowHistoryRequest,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  currentHistoryRecordId,
  currentHistoryRecordIdMatches,
  currentHistoryRowVersion,
  enqueueSaveWork,
  finishSave,
  historyPort,
  loadRows,
  nextClientTxnId,
  resolvePendingSocketTxn,
  rowHistory,
  rowHistoryPendingAction,
  rowHistoryRequestIsCurrent,
  selectedRowRecordId,
  setIsInspectorOpen,
  setRowHistory,
  setRowHistoryPendingAction,
  setSelectedRowId,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly acceptTimelineRecordVersion: (
    recordId: string,
    rowVersion: number,
  ) => void;
  readonly activeHistoryLiveRecordId: string | null;
  readonly beginRowHistoryRequest: () => number;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target: TimelineHistoryViewportContinuityTarget,
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly currentHistoryRecordId: string | null;
  readonly currentHistoryRecordIdMatches: (recordId: string) => boolean;
  readonly currentHistoryRowVersion: number | null;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly historyPort: TimelineHistoryPort;
  readonly loadRows: (options: TimelineHistoryLoadRowsOptions) => Promise<void>;
  readonly nextClientTxnId: () => string;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowHistory: RecordHistoryState;
  readonly rowHistoryPendingAction: RowHistoryPendingAction | null;
  readonly rowHistoryRequestIsCurrent: (requestSeq: number) => boolean;
  readonly selectedRowRecordId: string | null;
  readonly setIsInspectorOpen: (isOpen: boolean) => void;
  readonly setRowHistory: Dispatch<SetStateAction<RecordHistoryState>>;
  readonly setRowHistoryPendingAction: Dispatch<
    SetStateAction<RowHistoryPendingAction | null>
  >;
  readonly setSelectedRowId: (recordId: string | null) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
    options?: {
      readonly fallbackRowVersion?: number | null | undefined;
      readonly refreshIfMissing?: boolean;
    },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
}) {
  const fetchRecordHistory = useCallback(
    async (
      recordId: string,
      options: {
        readonly retainedData?: RecordHistoryData | null;
        readonly setLoading?: boolean;
      } = {},
    ): Promise<RecordHistoryData | null> => {
      const requestSeq = beginRowHistoryRequest();
      if (options.setLoading === true) {
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "loading",
          data:
            options.retainedData?.record_id === recordId
              ? options.retainedData
              : null,
          message: null,
        });
      }
      const result = await historyPort.load({ recordId });
      if (!rowHistoryRequestIsCurrent(requestSeq)) {
        return null;
      }
      if (result.kind === "rejected") {
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: result.failure.message,
        });
        return null;
      }
      const historyData = result.value;
      if (!rowHistoryRequestIsCurrent(requestSeq)) {
        return null;
      }
      acceptTimelineRecordVersion(recordId, historyData.row_version);
      setRowHistoryPendingAction(null);
      setRowHistory({
        recordId,
        status: "ready",
        data: historyData,
        message: null,
      });
      return historyData;
    },
    [
      acceptTimelineRecordVersion,
      beginRowHistoryRequest,
      historyPort,
      rowHistoryRequestIsCurrent,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const openRowHistory = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      void fetchRecordHistory(recordId, {
        retainedData: rowHistory.recordId === recordId ? rowHistory.data : null,
        setLoading: true,
      });
    },
    [
      fetchRecordHistory,
      rowHistory.data,
      rowHistory.recordId,
      setIsInspectorOpen,
      setSelectedRowId,
    ],
  );

  useEffect(() => {
    if (
      activeHistoryLiveRecordId === null ||
      rowHistory.status === "idle" ||
      rowHistory.recordId === activeHistoryLiveRecordId
    ) {
      return;
    }
    void fetchRecordHistory(activeHistoryLiveRecordId, {
      retainedData:
        rowHistory.recordId === activeHistoryLiveRecordId
          ? rowHistory.data
          : null,
      setLoading: true,
    });
  }, [
    activeHistoryLiveRecordId,
    fetchRecordHistory,
    rowHistory.data,
    rowHistory.recordId,
    rowHistory.status,
  ]);

  const submitRowHistoryMutation = useCallback(
    ({
      idleOptions,
      missingVersionMessage,
      onSuccess,
      recordId,
      request,
      viewportContinuityTarget,
    }: {
      idleOptions?: {
        readonly fallbackRowVersion?: number | null | undefined;
        readonly refreshIfMissing?: boolean;
      };
      missingVersionMessage: string;
      onSuccess: (
        accepted: TimelineHistoryMutationAccepted,
        viewportContinuityToken: number,
      ) => Promise<void>;
      recordId: string;
      request: (
        baseRowVersion: number,
        clientTxnId: string,
      ) => Promise<WorkbookOperationOutcome<TimelineHistoryMutationAccepted>>;
      viewportContinuityTarget: TimelineHistoryViewportContinuityTarget;
    }) => {
      const clientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity(
        viewportContinuityTarget,
      );
      beginSave();
      setRowHistoryPendingAction(null);
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
      enqueueSaveWork(async () => {
        const idleRecord = await waitForCommittedRecordIdle(
          recordId,
          idleOptions,
        );
        if (idleRecord === null) {
          clearViewportContinuity(viewportContinuityToken);
          setRowHistory((current) =>
            current.recordId === recordId
              ? {
                  ...current,
                  message: missingVersionMessage,
                }
              : current,
          );
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await request(idleRecord.rowVersion, clientTxnId);
        if (result.kind === "rejected") {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setRowHistory((current) =>
            current.recordId === recordId
              ? {
                  ...current,
                  message: result.failure.message,
                }
              : current,
          );
          finishSave("Conflict");
          return;
        }
        await onSuccess(result.value, viewportContinuityToken);
        finishSave("Saved");
      });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      nextClientTxnId,
      resolvePendingSocketTxn,
      setRowHistory,
      setRowHistoryPendingAction,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const submitRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const viewportContinuityTarget: TimelineHistoryViewportContinuityTarget =
        selectedRowRecordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion: currentHistoryRowVersion,
          refreshIfMissing: operation !== "restore",
        },
        missingVersionMessage: "Missing row version for destructive action.",
        recordId,
        viewportContinuityTarget,
        request: (baseRowVersion, clientTxnId) =>
          historyPort.deleteOrRestore({
            baseRowVersion,
            clientTxnId,
            operation,
            recordId,
          }),
        onSuccess: async (accepted, viewportContinuityToken) => {
          acceptTimelineRecordVersion(recordId, accepted.rowVersion);
          if (currentHistoryRecordIdMatches(recordId)) {
            await fetchRecordHistory(recordId);
          }
          if (operation === "restore") {
            setSelectedRowId(recordId);
          }
          await loadRows({
            showLoading: false,
            viewportContinuityToken,
          });
        },
      });
    },
    [
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRecordIdMatches,
      currentHistoryRowVersion,
      fetchRecordHistory,
      historyPort,
      loadRows,
      selectedRowRecordId,
      setSelectedRowId,
      submitRowHistoryMutation,
    ],
  );

  const submitRowHistoryRollbackTarget = useCallback(
    (
      pending: Extract<RowHistoryPendingAction, { readonly kind: "rollback" }>,
    ) => {
      const { recordId, target } = pending;
      if (recordId.trim() === "") {
        return;
      }
      const viewportContinuityTarget: TimelineHistoryViewportContinuityTarget =
        selectedRowRecordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion:
            currentHistoryRecordId === recordId
              ? currentHistoryRowVersion
              : pending.rowVersion,
        },
        missingVersionMessage: "Missing row version for rollback.",
        recordId,
        viewportContinuityTarget,
        request: (baseRowVersion, clientTxnId) =>
          historyPort.rollback({
            baseRowVersion,
            clientTxnId,
            recordId,
            target,
          }),
        onSuccess: async (accepted, viewportContinuityToken) => {
          acceptTimelineRecordVersion(recordId, accepted.rowVersion);
          if (currentHistoryRecordIdMatches(recordId)) {
            await fetchRecordHistory(recordId);
          }
          await loadRows({
            showLoading: false,
            viewportContinuityToken,
          });
        },
      });
    },
    [
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRecordIdMatches,
      currentHistoryRowVersion,
      fetchRecordHistory,
      historyPort,
      loadRows,
      selectedRowRecordId,
      submitRowHistoryMutation,
    ],
  );

  const previewRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      setRowHistoryPendingAction({
        kind: "destructive",
        operation,
        recordId,
        rowVersion: currentHistoryRowVersion,
      });
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
    },
    [
      currentHistoryRecordId,
      currentHistoryRowVersion,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const previewRowHistoryRollback = useCallback(
    (item: RecordHistoryItem, action: RecordHistoryRollbackAction) => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      const target = buildRecordRollbackTargetFromHistoryAction(item, action);
      if (target === null) {
        return;
      }
      setRowHistoryPendingAction({
        kind: "rollback",
        action,
        historyItemRef: item.history_item_ref,
        recordId,
        rowVersion: currentHistoryRowVersion,
        target,
      });
      setRowHistory((current) =>
        current.recordId === recordId ? { ...current, message: null } : current,
      );
    },
    [
      currentHistoryRecordId,
      currentHistoryRowVersion,
      setRowHistory,
      setRowHistoryPendingAction,
    ],
  );

  const confirmRowHistoryPendingAction = useCallback(() => {
    const pending = rowHistoryPendingAction;
    if (pending === null) {
      return;
    }
    if (pending.kind === "destructive") {
      submitRowHistoryDeleteRestore(pending.operation);
      return;
    }
    submitRowHistoryRollbackTarget(pending);
  }, [
    rowHistoryPendingAction,
    submitRowHistoryDeleteRestore,
    submitRowHistoryRollbackTarget,
  ]);

  return {
    confirmRowHistoryPendingAction,
    fetchRecordHistory,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
  };
}
