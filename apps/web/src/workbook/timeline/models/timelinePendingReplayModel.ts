import {
  type PendingReplayScope,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  WorkbookPendingQueueModel,
} from "../../utils/workbookPendingQueue";
import type { TimelineMutableRef } from "./timelineControllerPorts";

export type TimelinePendingQueueRuntime<TMeta> = {
  model: WorkbookPendingQueueModel;
  metaByUnitId: Map<string, TMeta>;
  refreshBlockedRecordIds: Map<string, number>;
  refreshInFlightCount: number;
  refreshReplayBlockAllCount: number;
  resetRefreshInFlight: boolean;
  replayScheduled: boolean;
};

export type TimelinePendingQueueSnapshot = {
  queuedCount: number;
  inFlightCount: number;
  haltedMessage: string | null;
  blockedEdit: TimelineBlockedEditRecovery | null;
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
};

export type TimelineBlockedEditRecovery = {
  unitId: string;
  errorCode: string;
  message: string;
  canRetryWithNewClientTxnId: boolean;
  canDiscard: true;
};

export type TimelinePendingReplayAdmissionRequest<TMeta> =
  PendingReplayUnitInput & TMeta;

export type TimelinePendingRefreshBlockScope =
  | { kind: "all" }
  | { kind: "record"; recordId: string }
  | { kind: "none" };

export type TimelinePendingSavesRefs<TMeta> = {
  readonly collectionKeyboardCommitRef: TimelineMutableRef<Map<string, string>>;
  readonly pendingOpsRef: TimelineMutableRef<number>;
  readonly pendingQueueRef: TimelineMutableRef<
    TimelinePendingQueueRuntime<TMeta>
  >;
  readonly pendingReplayAuthRetryRef: TimelineMutableRef<number | null>;
  readonly pendingReplayOrderRef: TimelineMutableRef<number>;
  readonly pendingReplayTimerRef: TimelineMutableRef<number | null>;
  readonly pendingSignaturesRef: TimelineMutableRef<Map<string, string>>;
  readonly pendingSocketTxnTimeoutsRef: TimelineMutableRef<Map<string, number>>;
  readonly saveQueueRef: TimelineMutableRef<Promise<void>>;
  readonly schedulePendingReplayRef: TimelineMutableRef<() => void>;
};

export function createTimelinePendingQueueRuntime<TMeta>(
  scope: PendingReplayScope,
): TimelinePendingQueueRuntime<TMeta> {
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

export function timelinePendingQueueSnapshot<TMeta>(
  pending: TimelinePendingQueueRuntime<TMeta>,
): TimelinePendingQueueSnapshot {
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

export function refreshBlocksTimelinePendingRecord<TMeta>(
  pending: TimelinePendingQueueRuntime<TMeta>,
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

export function refreshBlocksTimelinePendingUnit<TMeta>(
  pending: TimelinePendingQueueRuntime<TMeta>,
  unit: PendingReplayUnitState,
): boolean {
  return refreshBlocksTimelinePendingRecord(pending, unit.recordId);
}

export function beginTimelinePendingRefreshBlock<TMeta>(
  pending: TimelinePendingQueueRuntime<TMeta>,
  scope: TimelinePendingRefreshBlockScope,
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

export function finishTimelinePendingRefreshBlock<TMeta>(
  pending: TimelinePendingQueueRuntime<TMeta>,
  scope: TimelinePendingRefreshBlockScope,
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
