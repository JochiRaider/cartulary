import { useCallback, useEffect, useRef, useState } from "react";
import type { WorkbookPresenceDraft } from "../../collaboration/workbookCollaborationMessages";
import {
  presenceForCell,
  presenceForRow as scopedRowPresence,
  type WorkbookPresencePresentation,
} from "../../collaboration/workbookPresencePresentation";

const initialTimelinePresence: WorkbookPresenceDraft = {
  fieldKey: null,
  mode: "viewing",
  recordId: null,
};

export function useTimelinePresenceController({
  presence,
  publishPresence,
  resetKey,
}: {
  readonly presence: WorkbookPresencePresentation;
  readonly publishPresence: (presence: WorkbookPresenceDraft) => void;
  readonly resetKey: string | number;
}) {
  const [currentPresence, setCurrentPresence] = useState<WorkbookPresenceDraft>(
    initialTimelinePresence,
  );
  const currentPresenceRef = useRef<WorkbookPresenceDraft>(
    initialTimelinePresence,
  );
  currentPresenceRef.current = currentPresence;

  useEffect(() => {
    void resetKey;
    currentPresenceRef.current = initialTimelinePresence;
    setCurrentPresence(initialTimelinePresence);
  }, [resetKey]);

  const presenceForRow = useCallback(
    (recordId: string | null) => scopedRowPresence(presence, recordId),
    [presence],
  );
  const editingPresenceForCell = useCallback(
    (recordId: string | null, fieldKey: string) =>
      presenceForCell(presence, recordId, fieldKey),
    [presence],
  );
  const publishViewingPresence = useCallback(
    (recordId: string) => {
      if (currentPresenceRef.current.mode === "editing") return;
      publishNextPresence(
        { fieldKey: null, mode: "viewing", recordId },
        currentPresenceRef,
        setCurrentPresence,
        publishPresence,
      );
    },
    [publishPresence],
  );
  // Accepted grid focus is authoritative; an inspector selection alone is not.
  const publishFocusedPresence = useCallback(
    (recordId: string | null, fieldKey: string | null) => {
      const current = currentPresenceRef.current;
      if (
        current.mode === "editing" &&
        current.recordId === recordId &&
        current.fieldKey === fieldKey
      )
        return;
      publishNextPresence(
        {
          recordId,
          fieldKey: null,
          mode: recordId === null ? "idle" : "viewing",
        },
        currentPresenceRef,
        setCurrentPresence,
        publishPresence,
      );
    },
    [publishPresence],
  );
  const publishEditModePresence = useCallback(
    (recordId: string | null, fieldKey: string, editing: boolean) => {
      if (
        !editing &&
        (currentPresenceRef.current.fieldKey !== fieldKey ||
          currentPresenceRef.current.recordId !== recordId)
      )
        return;
      const next = editing
        ? { fieldKey, mode: "editing" as const, recordId }
        : {
            fieldKey: null,
            mode: "viewing" as const,
            recordId,
          };
      publishNextPresence(
        next,
        currentPresenceRef,
        setCurrentPresence,
        publishPresence,
      );
    },
    [publishPresence],
  );

  return {
    commands: {
      publishEditModePresence,
      publishFocusedPresence,
      publishViewingPresence,
    },
    snapshot: {
      currentPresence,
      currentPresenceRef,
      editingPresenceForCell,
      presenceForRow,
      presence,
    },
  };
}

function publishNextPresence(
  next: WorkbookPresenceDraft,
  currentPresenceRef: { current: WorkbookPresenceDraft },
  setCurrentPresence: (presence: WorkbookPresenceDraft) => void,
  publishPresence: (presence: WorkbookPresenceDraft) => void,
) {
  currentPresenceRef.current = next;
  setCurrentPresence(next);
  publishPresence(next);
}
