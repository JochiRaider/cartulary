import {
  type PendingReplayScope,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  WorkbookPendingQueueModel,
} from "../utils/workbookPendingQueue";

export type WorkbookMutableRef<T> = {
  current: T;
};

export type WorkbookPendingQueueRuntime<TMeta> = {
  model: WorkbookPendingQueueModel;
  metaByUnitId: Map<string, TMeta>;
  refreshBlockedRecordIds: Map<string, number>;
  refreshInFlightCount: number;
  refreshReplayBlockAllCount: number;
  resetRefreshInFlight: boolean;
  replayScheduled: boolean;
};

export type WorkbookPendingQueueSnapshot = {
  queuedCount: number;
  inFlightCount: number;
  haltedMessage: string | null;
  blockedEdit: WorkbookBlockedEditRecovery | null;
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
};

export type WorkbookBlockedEditRecovery = {
  unitId: string;
  errorCode: string;
  message: string;
  canRetryWithNewClientTxnId: boolean;
  canDiscard: true;
};

export type WorkbookPendingReplayAdmissionRequest<TMeta> =
  PendingReplayUnitInput & TMeta;

export type WorkbookPendingRefreshBlockScope =
  | { kind: "all" }
  | { kind: "record"; recordId: string }
  | { kind: "none" };

export type WorkbookPendingSavesRefs<TMeta> = {
  readonly collectionKeyboardCommitRef: WorkbookMutableRef<Map<string, string>>;
  readonly pendingOpsRef: WorkbookMutableRef<number>;
  readonly pendingQueueRef: WorkbookMutableRef<
    WorkbookPendingQueueRuntime<TMeta>
  >;
  readonly pendingReplayOrderRef: WorkbookMutableRef<number>;
  readonly pendingReplayTimerRef: WorkbookMutableRef<number | null>;
  readonly pendingSignaturesRef: WorkbookMutableRef<Map<string, string>>;
  readonly pendingSocketTxnTimeoutsRef: WorkbookMutableRef<Map<string, number>>;
  readonly saveQueueRef: WorkbookMutableRef<Promise<void>>;
  readonly schedulePendingReplayRef: WorkbookMutableRef<() => void>;
};

export function createWorkbookPendingQueueRuntime<TMeta>(
  scope: PendingReplayScope,
): WorkbookPendingQueueRuntime<TMeta> {
  return {
    model: new WorkbookPendingQueueModel(scope),
    metaByUnitId: new Map(),
    refreshBlockedRecordIds: new Map(),
    refreshInFlightCount: 0,
    refreshReplayBlockAllCount: 0,
    resetRefreshInFlight: false,
    replayScheduled: false,
  };
}

export function workbookPendingQueueSnapshot<TMeta>(
  pending: WorkbookPendingQueueRuntime<TMeta>,
): WorkbookPendingQueueSnapshot {
  const snapshot = pending.model.snapshot();
  return {
    queuedCount: snapshot.queuedCount,
    inFlightCount: snapshot.inFlightCount,
    haltedMessage: snapshot.halted?.message ?? null,
    blockedEdit:
      snapshot.halted === null
        ? null
        : {
            unitId: snapshot.halted.unit_id,
            errorCode: snapshot.halted.error_code,
            message: snapshot.halted.message,
            canRetryWithNewClientTxnId:
              snapshot.halted.error_code === "client_txn_conflict",
            canDiscard: true,
          },
    authPaused: snapshot.authPaused,
    overflowMessage: snapshot.overflow?.message ?? null,
    resetRefreshInFlight: pending.resetRefreshInFlight,
  };
}

export function refreshBlocksWorkbookPendingRecord<TMeta>(
  pending: WorkbookPendingQueueRuntime<TMeta>,
  recordId: string | null,
): boolean {
  if (pending.resetRefreshInFlight !== true) {
    return false;
  }
  if (pending.refreshReplayBlockAllCount > 0) {
    return true;
  }
  return recordId !== null && pending.refreshBlockedRecordIds.has(recordId);
}

export function refreshBlocksWorkbookPendingUnit<TMeta>(
  pending: WorkbookPendingQueueRuntime<TMeta>,
  unit: PendingReplayUnitState,
): boolean {
  return refreshBlocksWorkbookPendingRecord(pending, unit.recordId);
}

export function beginWorkbookPendingRefreshBlock<TMeta>(
  pending: WorkbookPendingQueueRuntime<TMeta>,
  scope: WorkbookPendingRefreshBlockScope,
): void {
  pending.refreshInFlightCount += 1;
  if (scope.kind === "all") {
    pending.refreshReplayBlockAllCount += 1;
  } else if (scope.kind === "record") {
    pending.refreshBlockedRecordIds.set(
      scope.recordId,
      (pending.refreshBlockedRecordIds.get(scope.recordId) ?? 0) + 1,
    );
  }
  pending.resetRefreshInFlight = pending.refreshInFlightCount > 0;
}

export function finishWorkbookPendingRefreshBlock<TMeta>(
  pending: WorkbookPendingQueueRuntime<TMeta>,
  scope: WorkbookPendingRefreshBlockScope,
): void {
  pending.refreshInFlightCount = Math.max(0, pending.refreshInFlightCount - 1);
  if (scope.kind === "all") {
    pending.refreshReplayBlockAllCount = Math.max(
      0,
      pending.refreshReplayBlockAllCount - 1,
    );
  } else if (scope.kind === "record") {
    const currentCount = pending.refreshBlockedRecordIds.get(scope.recordId);
    if (currentCount === undefined || currentCount <= 1) {
      pending.refreshBlockedRecordIds.delete(scope.recordId);
    } else {
      pending.refreshBlockedRecordIds.set(scope.recordId, currentCount - 1);
    }
  }
  pending.resetRefreshInFlight = pending.refreshInFlightCount > 0;
}
