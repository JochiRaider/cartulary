import { useState } from "react";
import type {
  PendingReplayUnitInput,
  WorkbookPendingQueueModel,
} from "./workbookPendingQueue";

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

type TimelineMutableRef<T> = {
  current: T;
};

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

export function timelineTabClientInstanceId(): string {
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

export function useTimelinePendingSaves<TMeta>({
  refs,
}: {
  readonly refs: TimelinePendingSavesRefs<TMeta>;
}) {
  const {
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
  } = refs;
  if (socketClientInstanceIdRef.current === null) {
    socketClientInstanceIdRef.current = timelineTabClientInstanceId();
  }
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
