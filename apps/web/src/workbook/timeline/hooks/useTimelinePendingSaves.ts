import { useRef, useState } from "react";
import {
  type PendingReplayScope,
  type PendingReplayUnitInput,
  type PendingReplayUnitState,
  WorkbookPendingQueueModel,
} from "../../utils/workbookPendingQueue";

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
  authPaused: boolean;
  overflowMessage: string | null;
  resetRefreshInFlight: boolean;
};

export type TimelinePendingReplayAdmissionRequest<TMeta> =
  PendingReplayUnitInput & TMeta;

export type TimelineMutableRef<T> = {
  current: T;
};

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
  readonly socketClientInstanceIdRef: TimelineMutableRef<string | null>;
};

function timelineTabClientInstanceId(): string {
  const key = "cartulary.client_instance_id";
  try {
    const existing = window.sessionStorage.getItem(key);
    if (existing) {
      return existing;
    }
    const created =
      window.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
    window.sessionStorage.setItem(key, created);
    return created;
  } catch {
    return `${Date.now()}-${Math.random()}`;
  }
}

export function ensureTimelineTabClientInstanceId(
  ref: TimelineMutableRef<string | null>,
): string {
  const clientInstanceId = ref.current ?? timelineTabClientInstanceId();
  ref.current = clientInstanceId;
  return clientInstanceId;
}

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

export function useTimelinePendingSaves<TMeta>({
  incidentId,
}: {
  readonly incidentId: string;
}) {
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const collectionKeyboardCommitRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const socketClientInstanceIdRef = useRef<string | null>(null);
  const clientInstanceId = ensureTimelineTabClientInstanceId(
    socketClientInstanceIdRef,
  );
  const pendingQueueRef = useRef<TimelinePendingQueueRuntime<TMeta>>(
    createTimelinePendingQueueRuntime({ incidentId, clientInstanceId }),
  );
  const pendingReplayOrderRef = useRef(1);
  const pendingReplayTimerRef = useRef<number | null>(null);
  const pendingReplayAuthRetryRef = useRef<number | null>(null);
  const schedulePendingReplayRef = useRef<() => void>(() => undefined);
  const [pendingQueueSnapshot, setPendingQueueSnapshot] =
    useState<TimelinePendingQueueSnapshot>({
      queuedCount: 0,
      inFlightCount: 0,
      haltedMessage: null,
      authPaused: false,
      overflowMessage: null,
      resetRefreshInFlight: false,
    });

  return {
    commands: {
      setPendingQueueSnapshot,
    },
    refs: {
      collectionKeyboardCommitRef,
      pendingOpsRef,
      pendingQueueRef,
      pendingReplayAuthRetryRef,
      pendingReplayOrderRef,
      pendingReplayTimerRef,
      pendingSignaturesRef,
      pendingSocketTxnTimeoutsRef,
      saveQueueRef,
      schedulePendingReplayRef,
      socketClientInstanceIdRef,
    },
    snapshot: {
      pendingQueueSnapshot,
    },
  };
}
