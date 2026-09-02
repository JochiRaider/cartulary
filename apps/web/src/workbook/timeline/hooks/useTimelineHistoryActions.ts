import { useCallback, useEffect } from "react";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  buildRecordRollbackTargetFromHistoryAction,
  type RecordHistoryData,
  type RecordHistoryItem,
  type RecordHistoryRollbackAction,
  type WorkbookRecordHistoryEvent,
  type WorkbookRecordHistoryOperationId,
  type WorkbookRecordHistoryPendingAction,
  type WorkbookRecordHistoryRequestId,
  type WorkbookRecordHistoryState,
  type WorkbookRecordHistorySubject,
} from "../../inspector/workbookRecordHistoryModel";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
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

export function useTimelineHistoryActions({
  acceptTimelineRecordVersion,
  activeHistoryLiveRecordId,
  beginRowHistoryOperation,
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
  dispatchRowHistory,
  retargetRowHistory,
  setSelectedRowId,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly acceptTimelineRecordVersion: (
    recordId: string,
    rowVersion: number,
  ) => void;
  readonly activeHistoryLiveRecordId: string | null;
  readonly beginRowHistoryOperation: () => WorkbookRecordHistoryOperationId;
  readonly beginRowHistoryRequest: () => WorkbookRecordHistoryRequestId;
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
  readonly dispatchRowHistory: (event: WorkbookRecordHistoryEvent) => void;
  readonly retargetRowHistory: (
    subject: WorkbookRecordHistorySubject | null,
  ) => void;
  readonly rowHistory: WorkbookRecordHistoryState;
  readonly rowHistoryPendingAction: WorkbookRecordHistoryPendingAction | null;
  readonly rowHistoryRequestIsCurrent: (
    requestId: WorkbookRecordHistoryRequestId,
  ) => boolean;
  readonly selectedRowRecordId: string | null;
  readonly setIsInspectorOpen: (isOpen: boolean) => void;
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
        readonly subject?: WorkbookRecordHistorySubject;
      } = {},
    ): Promise<RecordHistoryData | null> => {
      const requestId = beginRowHistoryRequest();
      const currentSubject = rowHistory.subject;
      const rowVersion =
        currentSubject?.recordId === recordId
          ? currentSubject.rowVersion
          : currentHistoryRowVersion;
      if (rowVersion === null || rowVersion === undefined) return null;
      const activeSubject: WorkbookRecordHistorySubject =
        options.subject ??
        (currentSubject?.recordId === recordId
          ? currentSubject
          : { kind: "live", recordId, rowVersion });
      retargetRowHistory(activeSubject);
      if (options.setLoading === true) {
        dispatchRowHistory({
          requestId,
          subject: activeSubject,
          type: "load_requested",
        });
      }
      const result = await historyPort.load({ recordId });
      if (!rowHistoryRequestIsCurrent(requestId)) {
        return null;
      }
      if (result.kind === "rejected") {
        dispatchRowHistory({
          error: workbookInspectorErrorPresentation(result.failure),
          requestId,
          subject: activeSubject,
          type: "load_rejected",
        });
        return null;
      }
      const historyData = result.value;
      if (!rowHistoryRequestIsCurrent(requestId)) {
        return null;
      }
      acceptTimelineRecordVersion(recordId, historyData.row_version);
      dispatchRowHistory({
        data: historyData,
        requestId,
        subject: activeSubject,
        type: "load_accepted",
      });
      return historyData;
    },
    [
      acceptTimelineRecordVersion,
      beginRowHistoryRequest,
      currentHistoryRowVersion,
      dispatchRowHistory,
      historyPort,
      retargetRowHistory,
      rowHistory.subject,
      rowHistoryRequestIsCurrent,
    ],
  );

  const openRowHistory = useCallback(
    (recordId: string) => {
      setSelectedRowId(recordId);
      setIsInspectorOpen(true);
      void fetchRecordHistory(recordId, {
        retainedData:
          rowHistory.subject?.recordId === recordId ? rowHistory.data : null,
        setLoading: true,
      });
    },
    [
      fetchRecordHistory,
      rowHistory.data,
      rowHistory.subject,
      setIsInspectorOpen,
      setSelectedRowId,
    ],
  );

  useEffect(() => {
    if (
      activeHistoryLiveRecordId === null ||
      rowHistory.phase === "idle" ||
      rowHistory.subject?.recordId === activeHistoryLiveRecordId
    ) {
      return;
    }
    void fetchRecordHistory(activeHistoryLiveRecordId, {
      retainedData:
        rowHistory.subject?.recordId === activeHistoryLiveRecordId
          ? rowHistory.data
          : null,
      setLoading: true,
    });
  }, [
    activeHistoryLiveRecordId,
    fetchRecordHistory,
    rowHistory.data,
    rowHistory.phase,
    rowHistory.subject,
  ]);

  const submitRowHistoryMutation = useCallback(
    ({
      idleOptions,
      missingVersionMessage,
      onSuccess,
      operationId,
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
      operationId: WorkbookRecordHistoryOperationId;
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
      enqueueSaveWork(async () => {
        const idleRecord = await waitForCommittedRecordIdle(
          recordId,
          idleOptions,
        );
        if (idleRecord === null) {
          clearViewportContinuity(viewportContinuityToken);
          dispatchRowHistory({
            error: workbookInspectorLocalErrorPresentation(
              missingVersionMessage,
            ),
            operationId,
            type: "operation_rejected",
          });
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await request(idleRecord.rowVersion, clientTxnId);
        if (result.kind === "rejected") {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          dispatchRowHistory({
            error: workbookInspectorErrorPresentation(result.failure),
            operationId,
            type: "operation_rejected",
          });
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
      dispatchRowHistory,
      enqueueSaveWork,
      finishSave,
      nextClientTxnId,
      resolvePendingSocketTxn,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const submitRowHistoryDeleteRestore = useCallback(
    (
      pending: Extract<
        WorkbookRecordHistoryPendingAction,
        { readonly kind: "destructive" }
      >,
      operationId: WorkbookRecordHistoryOperationId,
    ) => {
      const { operation, recordId } = pending;
      const viewportContinuityTarget: TimelineHistoryViewportContinuityTarget =
        selectedRowRecordId === recordId
          ? { kind: "row-inspect", recordId }
          : { kind: "scroll-only" };
      submitRowHistoryMutation({
        idleOptions: {
          fallbackRowVersion: pending.rowVersion,
          refreshIfMissing: operation !== "restore",
        },
        missingVersionMessage: "Missing row version for destructive action.",
        operationId,
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
          const nextSubject: WorkbookRecordHistorySubject = {
            kind: operation === "delete" ? "deleted" : "live",
            recordId,
            rowVersion: accepted.rowVersion,
          };
          dispatchRowHistory({
            feedback: workbookInspectorMessageFeedback(
              operation === "delete"
                ? `Deleted record ${recordId}.`
                : `Restored record ${recordId}.`,
              "polite",
            ),
            operationId,
            subject: nextSubject,
            type: "operation_accepted",
          });
          if (currentHistoryRecordIdMatches(recordId)) {
            await fetchRecordHistory(recordId, {
              setLoading: true,
              subject: nextSubject,
            });
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
      currentHistoryRecordIdMatches,
      dispatchRowHistory,
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
      pending: Extract<
        WorkbookRecordHistoryPendingAction,
        { readonly kind: "rollback" }
      >,
      operationId: WorkbookRecordHistoryOperationId,
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
        operationId,
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
          const nextSubject: WorkbookRecordHistorySubject = {
            kind:
              rowHistory.subject?.recordId === recordId
                ? rowHistory.subject.kind
                : "live",
            recordId,
            rowVersion: accepted.rowVersion,
          };
          dispatchRowHistory({
            feedback: workbookInspectorMessageFeedback(
              `Rolled back record ${recordId}.`,
              "polite",
            ),
            operationId,
            subject: nextSubject,
            type: "operation_accepted",
          });
          if (currentHistoryRecordIdMatches(recordId)) {
            await fetchRecordHistory(recordId, {
              setLoading: true,
              subject: nextSubject,
            });
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
      dispatchRowHistory,
      fetchRecordHistory,
      historyPort,
      loadRows,
      selectedRowRecordId,
      rowHistory.subject,
      submitRowHistoryMutation,
    ],
  );

  const previewRowHistoryDeleteRestore = useCallback(
    (operation: "delete" | "restore") => {
      const recordId = currentHistoryRecordId;
      if (recordId === null || recordId === undefined) {
        return;
      }
      if (currentHistoryRowVersion === null) return;
      dispatchRowHistory({
        pendingAction: {
          kind: "destructive",
          operation,
          recordId,
          rowVersion: currentHistoryRowVersion,
        },
        type: "preview",
      });
    },
    [currentHistoryRecordId, currentHistoryRowVersion, dispatchRowHistory],
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
      if (currentHistoryRowVersion === null) return;
      dispatchRowHistory({
        pendingAction: {
          action,
          historyItemRef: item.history_item_ref,
          kind: "rollback",
          recordId,
          rowVersion: currentHistoryRowVersion,
          target,
        },
        type: "preview",
      });
    },
    [currentHistoryRecordId, currentHistoryRowVersion, dispatchRowHistory],
  );

  const confirmRowHistoryPendingAction = useCallback(() => {
    const pending = rowHistoryPendingAction;
    if (pending === null) {
      return;
    }
    const operationId = beginRowHistoryOperation();
    dispatchRowHistory({ operationId, type: "submit" });
    if (pending.kind === "destructive") {
      submitRowHistoryDeleteRestore(pending, operationId);
      return;
    }
    submitRowHistoryRollbackTarget(pending, operationId);
  }, [
    beginRowHistoryOperation,
    dispatchRowHistory,
    rowHistoryPendingAction,
    submitRowHistoryDeleteRestore,
    submitRowHistoryRollbackTarget,
  ]);

  return {
    cancelRowHistoryPendingAction: () => dispatchRowHistory({ type: "cancel" }),
    confirmRowHistoryPendingAction,
    fetchRecordHistory,
    openRowHistory,
    previewRowHistoryDeleteRestore,
    previewRowHistoryRollback,
  };
}
