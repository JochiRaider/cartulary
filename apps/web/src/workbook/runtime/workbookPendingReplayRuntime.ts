import {
  createWorkbookPendingQueueModel,
  type PendingReplayScope,
  type PendingReplayUnitState,
  type WorkbookPendingQueueModel,
} from "./pending/workbookPendingQueue";

export type WorkbookPendingQueueRuntime = {
  model: WorkbookPendingQueueModel;
  refreshBlockedRecordIds: Map<string, number>;
  refreshInFlightCount: number;
  refreshReplayBlockAllCount: number;
  resetRefreshInFlight: boolean;
};

export type WorkbookPendingRefreshBlockScope =
  | { kind: "all" }
  | { kind: "record"; recordId: string }
  | { kind: "none" };

export function createWorkbookPendingQueueRuntime(
  scope: PendingReplayScope,
): WorkbookPendingQueueRuntime {
  return {
    model: createWorkbookPendingQueueModel(scope),
    refreshBlockedRecordIds: new Map(),
    refreshInFlightCount: 0,
    refreshReplayBlockAllCount: 0,
    resetRefreshInFlight: false,
  };
}

export function refreshBlocksWorkbookPendingRecord(
  pending: WorkbookPendingQueueRuntime,
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

export function refreshBlocksWorkbookPendingUnit(
  pending: WorkbookPendingQueueRuntime,
  unit: PendingReplayUnitState,
): boolean {
  return refreshBlocksWorkbookPendingRecord(pending, unit.recordId);
}

export function beginWorkbookPendingRefreshBlock(
  pending: WorkbookPendingQueueRuntime,
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

export function finishWorkbookPendingRefreshBlock(
  pending: WorkbookPendingQueueRuntime,
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
