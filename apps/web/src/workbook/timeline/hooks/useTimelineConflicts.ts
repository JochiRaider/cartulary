import { useCallback, useState } from "react";
import type {
  LocalConflictState,
  PasteConflictGroupState,
} from "../models/workbookTimelineModel";

type TimelineMutableRef<T> = {
  current: T;
};

export function useTimelineConflicts({
  conflictQueueRef,
}: {
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
}) {
  const [conflictQueue, setConflictQueue] = useState<
    Record<string, LocalConflictState>
  >({});
  const [activeConflictKey, setActiveConflictKey] = useState<string | null>(
    null,
  );
  const [pasteConflictGroup, setPasteConflictGroup] =
    useState<PasteConflictGroupState | null>(null);
  const setConflictQueueState = useCallback(
    (
      updater: (
        current: Record<string, LocalConflictState>,
      ) => Record<string, LocalConflictState>,
    ) => {
      setConflictQueue((current) => {
        const next = updater(current);
        conflictQueueRef.current = next;
        return next;
      });
    },
    [conflictQueueRef],
  );

  return {
    commands: {
      setActiveConflictKey,
      setConflictQueueState,
      setPasteConflictGroup,
    },
    refs: {
      conflictQueueRef,
    },
    snapshot: {
      activeConflictKey,
      conflictQueue,
      pasteConflictGroup,
    },
  };
}
