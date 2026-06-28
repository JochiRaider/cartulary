import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
} from "react";
import { apiPath } from "../../../services/browserApi";
import {
  fetchJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import type {
  RecordHistoryData,
  RecordHistoryItem,
  RecordHistoryRollbackAction,
  RecordHistoryState,
  RowHistoryPendingAction,
} from "../components/TimelineHistoryPanel";
import {
  buildTimelineDeleteRestorePayload,
  buildTimelineRollbackPayload,
} from "../services/timelineMutationRequests";

type RecordHistoryEnvelope = {
  data: RecordHistoryData;
};

type RecordDeleteRestoreEnvelope = {
  data: {
    record_id: string;
    incident_id: string;
    row_version: number;
    deleted: boolean;
    deleted_at: string | null;
    deleted_by_user_id: string | null;
    change_set_id: string;
  };
};

type RecordRollbackEnvelope = {
  data: {
    incident_id: string;
    record_id: string;
    row_version: number;
    target: Record<string, unknown>;
    target_change_set_id: string;
    rollback_change_set_id: string;
    affected_record_ids: string[];
  };
};

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

const rowHistoryRollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

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

function normalizeRecordHistoryData(
  data: RecordHistoryData,
): RecordHistoryData {
  if (!Array.isArray(data.items)) {
    throw new Error("row history items must be an array");
  }
  const seenItemRefs = new Set<string>();
  for (const item of data.items) {
    if (!isNonEmptyString(item.history_item_ref)) {
      throw new Error("row history item is missing history_item_ref");
    }
    if (seenItemRefs.has(item.history_item_ref)) {
      throw new Error("row history item has duplicate history_item_ref");
    }
    seenItemRefs.add(item.history_item_ref);
    if (!isNonEmptyString(item.change_set_id)) {
      throw new Error("row history item is missing change_set_id");
    }
    if (
      item.history_entry_ref !== undefined &&
      !isNonEmptyString(item.history_entry_ref)
    ) {
      throw new Error("row history item has invalid history_entry_ref");
    }
    if (
      item.revision_no !== undefined &&
      !isPositiveInteger(item.revision_no)
    ) {
      throw new Error("row history item has invalid revision_no");
    }
    validateRowHistoryActions(item);
  }
  return data;
}

function validateRowHistoryActions(item: RecordHistoryItem) {
  if (!Array.isArray(item.available_rollback_actions)) {
    throw new Error("row history actions must be an array");
  }
  let previousIndex = -1;
  for (const action of item.available_rollback_actions as unknown[]) {
    if (!isRowHistoryRollbackAction(action)) {
      throw new Error("row history action token is invalid");
    }
    const actionIndex = rowHistoryRollbackActionOrder.indexOf(action);
    if (actionIndex <= previousIndex) {
      throw new Error("row history actions are not canonical");
    }
    previousIndex = actionIndex;
    if (buildRecordRollbackTargetFromHistoryAction(item, action) === null) {
      throw new Error("row history action is missing its selector");
    }
  }
}

function isRowHistoryRollbackAction(
  value: unknown,
): value is RecordHistoryRollbackAction {
  return (
    value === "history_entry" ||
    value === "change_set" ||
    value === "row_restore"
  );
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

export function useTimelineHistoryActions({
  acceptTimelineRecordVersion,
  activeHistoryLiveRecordId,
  apiBase,
  beginRowHistoryRequest,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  currentHistoryRecordId,
  currentHistoryRecordIdMatches,
  currentHistoryRowVersion,
  enqueueSaveWork,
  finishSave,
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
  readonly apiBase?: string | undefined;
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
      const result = await fetchJSON<RecordHistoryEnvelope>(
        apiPath(apiBase, `/api/v1/records/${recordId}/history`),
      );
      if (!rowHistoryRequestIsCurrent(requestSeq)) {
        return null;
      }
      if (!result.ok) {
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: parseErrorMessage(result.payload),
        });
        return null;
      }
      let historyData: RecordHistoryData;
      try {
        const envelope = readEnvelope<RecordHistoryEnvelope>(result.payload);
        historyData = normalizeRecordHistoryData(envelope.data);
        if (historyData.record_id !== recordId) {
          throw new Error("row history response record mismatch");
        }
      } catch {
        if (!rowHistoryRequestIsCurrent(requestSeq)) {
          return null;
        }
        setRowHistoryPendingAction(null);
        setRowHistory({
          recordId,
          status: "error",
          data: null,
          message: "Invalid row history response.",
        });
        return null;
      }
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
      apiBase,
      beginRowHistoryRequest,
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
        payload: unknown,
        viewportContinuityToken: number,
      ) => Promise<void>;
      recordId: string;
      request: (
        baseRowVersion: number,
        clientTxnId: string,
      ) => Promise<{ ok: boolean; payload: unknown }>;
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
        if (!result.ok) {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setRowHistory((current) =>
            current.recordId === recordId
              ? {
                  ...current,
                  message: parseErrorMessage(result.payload),
                }
              : current,
          );
          finishSave("Conflict");
          return;
        }
        await onSuccess(result.payload, viewportContinuityToken);
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
        request: (baseRowVersion, clientTxnId) => {
          const path =
            operation === "delete"
              ? `/api/v1/records/${recordId}`
              : `/api/v1/records/${recordId}/restore`;
          return fetchJSON<RecordDeleteRestoreEnvelope>(
            apiPath(apiBase, path),
            {
              method: operation === "delete" ? "DELETE" : "POST",
              body: JSON.stringify(
                buildTimelineDeleteRestorePayload({
                  baseRowVersion,
                  clientTxnId,
                  operation,
                }),
              ),
            },
          );
        },
        onSuccess: async (payload, viewportContinuityToken) => {
          const envelope = readEnvelope<RecordDeleteRestoreEnvelope>(payload);
          acceptTimelineRecordVersion(recordId, envelope.data.row_version);
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
      apiBase,
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRecordIdMatches,
      currentHistoryRowVersion,
      fetchRecordHistory,
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
          fetchJSON<RecordRollbackEnvelope>(
            apiPath(apiBase, `/api/v1/records/${recordId}/rollback`),
            {
              method: "POST",
              body: JSON.stringify(
                buildTimelineRollbackPayload({
                  baseRowVersion,
                  clientTxnId,
                  target,
                }),
              ),
            },
          ),
        onSuccess: async (payload, viewportContinuityToken) => {
          const envelope = readEnvelope<RecordRollbackEnvelope>(payload);
          acceptTimelineRecordVersion(recordId, envelope.data.row_version);
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
      apiBase,
      acceptTimelineRecordVersion,
      currentHistoryRecordId,
      currentHistoryRecordIdMatches,
      currentHistoryRowVersion,
      fetchRecordHistory,
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
