import { useCallback, useState } from "react";
import type {
  PresenceRecord,
  WorkbookPresenceMode,
} from "../../utils/workbookPresence";
import {
  type createWorkbookSocketLifecycleState,
  reduceWorkbookSocketLifecycle,
  type WorkbookSocketLifecycleAction,
  type WorkbookSocketLifecycleEffect,
} from "../services/workbookSocketLifecycle";

export type TimelinePresenceDraft = {
  fieldKey: string | null;
  mode: WorkbookPresenceMode;
  recordId: string | null;
};

type TimelineMutableRef<T> = {
  current: T;
};

type DispatchSocketLifecycle = (
  action: WorkbookSocketLifecycleAction,
) => WorkbookSocketLifecycleEffect[];

export type TimelineLiveUpdateRefs = {
  readonly activeSocketRef: TimelineMutableRef<WebSocket | null>;
  readonly currentPresenceRef: TimelineMutableRef<TimelinePresenceDraft>;
  readonly dispatchSocketLifecycleRef: TimelineMutableRef<DispatchSocketLifecycle>;
  readonly presenceUpdateTimerRef: TimelineMutableRef<number | null>;
  readonly socketConnectionIDRef: TimelineMutableRef<string | null>;
  readonly socketEstablishedRef: TimelineMutableRef<boolean>;
  readonly socketLastSeenStreamSeqRef: TimelineMutableRef<number>;
  readonly socketLifecycleRef: TimelineMutableRef<
    ReturnType<typeof createWorkbookSocketLifecycleState>
  >;
  readonly socketReconnectAfterAuthRef: TimelineMutableRef<(() => void) | null>;
  readonly socketResumeTokenRef: TimelineMutableRef<string | null>;
};

export function useTimelineLiveUpdates({
  refs,
}: {
  readonly refs: TimelineLiveUpdateRefs;
}) {
  const [currentPresence, setCurrentPresence] = useState<TimelinePresenceDraft>(
    {
      fieldKey: null,
      mode: "viewing",
      recordId: null,
    },
  );
  const [presenceRecords, setPresenceRecords] = useState<PresenceRecord[]>([]);
  const {
    activeSocketRef,
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketConnectionIDRef,
    socketEstablishedRef,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
    socketResumeTokenRef,
  } = refs;

  const syncSocketLifecycleRefs = useCallback(() => {
    const state = socketLifecycleRef.current;
    socketEstablishedRef.current = state.established;
    socketConnectionIDRef.current = state.connectionId;
    socketResumeTokenRef.current = state.resumeToken;
    socketLastSeenStreamSeqRef.current = state.lastSeenStreamSeq;
  }, [
    socketConnectionIDRef,
    socketEstablishedRef,
    socketLastSeenStreamSeqRef,
    socketLifecycleRef,
    socketResumeTokenRef,
  ]);

  const dispatchSocketLifecycle = useCallback(
    (
      action: WorkbookSocketLifecycleAction,
    ): WorkbookSocketLifecycleEffect[] => {
      const reduction = reduceWorkbookSocketLifecycle(
        socketLifecycleRef.current,
        action,
      );
      socketLifecycleRef.current = reduction.state;
      syncSocketLifecycleRefs();
      return reduction.effects;
    },
    [socketLifecycleRef, socketLifecycleRef.current, syncSocketLifecycleRefs],
  );
  dispatchSocketLifecycleRef.current = dispatchSocketLifecycle;

  return {
    commands: {
      dispatchSocketLifecycle,
      setCurrentPresence,
      setPresenceRecords,
      syncSocketLifecycleRefs,
    },
    refs: {
      activeSocketRef,
      currentPresenceRef,
      dispatchSocketLifecycleRef,
      presenceUpdateTimerRef,
      socketConnectionIDRef,
      socketEstablishedRef,
      socketLastSeenStreamSeqRef,
      socketLifecycleRef,
      socketReconnectAfterAuthRef,
      socketResumeTokenRef,
    },
    snapshot: {
      currentPresence,
      presenceRecords,
    },
  };
}
