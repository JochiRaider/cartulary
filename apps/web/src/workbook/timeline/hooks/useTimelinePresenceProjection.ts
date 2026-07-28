import { useCallback, useMemo } from "react";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import type { WorkbookPresenceDraft } from "../../runtime/workbookCollaborationMessages";
import {
  type PresenceRecord,
  presenceMatchesSheet,
} from "../../utils/workbookPresence";

type TimelineMutableRef<T> = {
  current: T;
};

export function useTimelinePresenceProjection({
  activeSheetRef,
  connectionId,
  currentPresenceRef,
  presenceRecords,
  sendPresenceUpdate,
  setCurrentPresence,
}: {
  readonly activeSheetRef: WorkbookSheetRef;
  readonly connectionId: string | null;
  readonly currentPresenceRef: TimelineMutableRef<WorkbookPresenceDraft>;
  readonly presenceRecords: readonly PresenceRecord[];
  readonly sendPresenceUpdate: (presence: WorkbookPresenceDraft) => void;
  readonly setCurrentPresence: (presence: WorkbookPresenceDraft) => void;
}) {
  const activeSheetPresenceRecords = useMemo(
    () =>
      [...presenceRecords]
        .filter((presence) => presenceMatchesSheet(presence, activeSheetRef))
        .filter((presence) => presence.connection_id !== connectionId)
        .sort((left, right) => {
          const byName = left.display_name.localeCompare(right.display_name);
          return byName === 0
            ? left.connection_id.localeCompare(right.connection_id)
            : byName;
        }),
    [activeSheetRef, connectionId, presenceRecords],
  );

  const presenceForRow = useCallback(
    (recordId: string | null) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) => presence.record_id === recordId,
          ),
    [activeSheetPresenceRecords],
  );

  const editingPresenceForCell = useCallback(
    (recordId: string | null, fieldKey: string) =>
      recordId === null
        ? []
        : activeSheetPresenceRecords.filter(
            (presence) =>
              presence.record_id === recordId &&
              presence.field_key === fieldKey &&
              presence.mode === "editing",
          ),
    [activeSheetPresenceRecords],
  );

  const handleEditModePresence = useCallback(
    (recordId: string | null, fieldKey: string, editing: boolean) => {
      const next = editing
        ? { fieldKey, mode: "editing" as const, recordId }
        : {
            fieldKey: null,
            mode: "viewing" as const,
            recordId: recordId ?? currentPresenceRef.current.recordId,
          };
      currentPresenceRef.current = next;
      setCurrentPresence(next);
      sendPresenceUpdate(next);
    },
    [currentPresenceRef, sendPresenceUpdate, setCurrentPresence],
  );

  return {
    activeSheetPresenceRecords,
    editingPresenceForCell,
    handleEditModePresence,
    presenceForRow,
  };
}
