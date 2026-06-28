import { useCallback, useMemo } from "react";
import type { WorkbookSheetRef } from "../../models/workbookStartup";
import {
  type PresenceRecord,
  presenceMatchesSheet,
} from "../../utils/workbookPresence";
import type { TimelinePresenceDraft } from "./useTimelineLiveUpdates";

type TimelineMutableRef<T> = {
  current: T;
};

export function useTimelinePresenceProjection({
  activeSheetRef,
  currentPresenceRef,
  presenceRecords,
  sendPresenceUpdate,
  setCurrentPresence,
  socketConnectionIDRef,
}: {
  readonly activeSheetRef: WorkbookSheetRef;
  readonly currentPresenceRef: TimelineMutableRef<TimelinePresenceDraft>;
  readonly presenceRecords: readonly PresenceRecord[];
  readonly sendPresenceUpdate: (presence: TimelinePresenceDraft) => void;
  readonly setCurrentPresence: (presence: TimelinePresenceDraft) => void;
  readonly socketConnectionIDRef: TimelineMutableRef<string | null>;
}) {
  const activeSheetPresenceRecords = useMemo(
    () =>
      [...presenceRecords]
        .filter((presence) => presenceMatchesSheet(presence, activeSheetRef))
        .filter(
          (presence) =>
            presence.connection_id !== socketConnectionIDRef.current,
        )
        .sort((left, right) => {
          const byName = left.display_name.localeCompare(right.display_name);
          return byName === 0
            ? left.connection_id.localeCompare(right.connection_id)
            : byName;
        }),
    [activeSheetRef, presenceRecords, socketConnectionIDRef],
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
