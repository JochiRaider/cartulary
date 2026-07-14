import { useCallback, useState } from "react";
import type { PresenceRecord } from "../../utils/workbookPresence";
import type { TimelineLiveUpdateRefs } from "../models/timelineControllerPorts";
import {
  reduceTimelineCollaboration,
  type TimelineCollaborationAction,
  type TimelineCollaborationEffect,
} from "../services/timelineCollaborationEffects";
import type { TimelinePresenceDraft } from "../services/workbookCollaborationMessages";

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
    collaborationStateRef,
    dispatchCollaborationRef,
    presenceUpdateTimerRef,
    socketReconnectAfterAuthRef,
  } = refs;

  const dispatchCollaboration = useCallback(
    (
      action: TimelineCollaborationAction,
    ): readonly TimelineCollaborationEffect[] => {
      const reduction = reduceTimelineCollaboration(
        collaborationStateRef.current,
        action,
      );
      collaborationStateRef.current = reduction.state;
      return reduction.effects;
    },
    [collaborationStateRef],
  );
  dispatchCollaborationRef.current = dispatchCollaboration;

  return {
    commands: {
      setCurrentPresence,
      setPresenceRecords,
    },
    refs: {
      currentPresenceRef,
      collaborationStateRef,
      dispatchCollaborationRef,
      presenceUpdateTimerRef,
      socketReconnectAfterAuthRef,
    },
    snapshot: {
      currentPresence,
      presenceRecords,
    },
  };
}
