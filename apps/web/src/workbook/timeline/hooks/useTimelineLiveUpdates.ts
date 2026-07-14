import { useCallback, useState } from "react";
import type { PresenceRecord } from "../../utils/workbookPresence";
import type { TimelineLiveUpdateRefs } from "../models/timelineControllerPorts";
import type { TimelinePresenceDraft } from "../services/workbookCollaborationMessages";
import {
  reduceWorkbookSocketLifecycle,
  type WorkbookSocketLifecycleAction,
  type WorkbookSocketLifecycleEffect,
} from "../services/workbookSocketLifecycle";

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
    currentPresenceRef,
    dispatchSocketLifecycleRef,
    presenceUpdateTimerRef,
    socketConnectionIDRef,
    socketLifecycleRef,
    socketReconnectAfterAuthRef,
  } = refs;

  const syncSocketLifecycleRefs = useCallback(() => {
    const state = socketLifecycleRef.current;
    socketConnectionIDRef.current = state.connectionId;
  }, [socketConnectionIDRef, socketLifecycleRef]);

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
      currentPresenceRef,
      dispatchSocketLifecycleRef,
      presenceUpdateTimerRef,
      socketConnectionIDRef,
      socketLifecycleRef,
      socketReconnectAfterAuthRef,
    },
    snapshot: {
      currentPresence,
      presenceRecords,
    },
  };
}
