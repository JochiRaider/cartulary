import { useRef, useState } from "react";
import {
  createTimelinePendingQueueRuntime,
  type TimelinePendingQueueRuntime,
  type TimelinePendingQueueSnapshot,
} from "../models/timelinePendingReplayModel";

export function useTimelinePendingSaves<TMeta>({
  clientInstanceId,
  incidentId,
}: {
  readonly clientInstanceId: string;
  readonly incidentId: string;
}) {
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const collectionKeyboardCommitRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
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
    },
    snapshot: {
      pendingQueueSnapshot,
    },
  };
}
