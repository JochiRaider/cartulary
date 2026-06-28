import { useCallback, useEffect, useRef } from "react";
import {
  deriveWorkbookSaveState,
  type WorkbookSaveStateConflictAnchor,
} from "../../utils/workbookPendingQueue";
import type { LocalConflictState } from "../models/workbookTimelineModel";
import {
  beginTimelinePendingRefreshBlock,
  finishTimelinePendingRefreshBlock,
  type TimelinePendingQueueRuntime,
  type TimelinePendingQueueSnapshot,
  type TimelinePendingRefreshBlockScope,
  type TimelinePendingSavesRefs,
  timelinePendingQueueSnapshot,
} from "./useTimelinePendingSaves";

export type TimelineSaveStateLabel = "Conflict" | "Saved" | "Syncing";

type TimelineMutableRef<T> = {
  current: T;
};

type TimelineSetState<T> = (value: T) => void;

function saveStateConflictAnchorsFromLocalConflicts(
  conflicts: Record<string, LocalConflictState>,
): WorkbookSaveStateConflictAnchor[] {
  return Object.values(conflicts).map((entry) => ({ ...entry.anchor }));
}

export function useTimelineSaveStatePresentation<TMeta>({
  conflictQueue,
  conflictQueueRef,
  pendingQueueSnapshot,
  pendingSavesRefsRef,
  setPendingQueueSnapshot,
  setSaveState,
  setSaveStateSecondaryMessage,
}: {
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly pendingQueueSnapshot: TimelinePendingQueueSnapshot;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    TimelinePendingSavesRefs<TMeta>
  >;
  readonly setPendingQueueSnapshot: TimelineSetState<TimelinePendingQueueSnapshot>;
  readonly setSaveState: TimelineSetState<TimelineSaveStateLabel>;
  readonly setSaveStateSecondaryMessage: TimelineSetState<string | null>;
}) {
  const computeSaveStatePresentation = useCallback(
    (
      pending: TimelinePendingQueueRuntime<TMeta>,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const snapshot = pending.model.snapshot();
      return deriveWorkbookSaveState({
        authPaused: snapshot.authPaused,
        halted: snapshot.halted,
        overflow: snapshot.overflow,
        sameFieldConflicts: snapshot.sameFieldConflicts,
        localDraftConflicts:
          saveStateConflictAnchorsFromLocalConflicts(conflicts),
        queuedCount: snapshot.queuedCount,
        inFlightCount: snapshot.inFlightCount,
        refreshPaused: pending.resetRefreshInFlight,
        pendingMutationCount: pendingSavesRefsRef.current.pendingOpsRef.current,
      });
    },
    [conflictQueueRef, pendingSavesRefsRef],
  );

  const publishSaveStatePresentation = useCallback(
    (
      pending: TimelinePendingQueueRuntime<TMeta>,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const presentation = computeSaveStatePresentation(pending, conflicts);
      setSaveState(presentation.primaryLabel);
      setSaveStateSecondaryMessage(presentation.secondaryMessage);
      return presentation;
    },
    [
      computeSaveStatePresentation,
      conflictQueueRef,
      setSaveState,
      setSaveStateSecondaryMessage,
    ],
  );

  const publishPendingQueueState = useCallback(() => {
    const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
    setPendingQueueSnapshot(timelinePendingQueueSnapshot(pending));
    publishSaveStatePresentation(pending);
  }, [
    pendingSavesRefsRef,
    publishSaveStatePresentation,
    setPendingQueueSnapshot,
  ]);
  const publishPendingQueueStateRef = useRef(publishPendingQueueState);
  publishPendingQueueStateRef.current = publishPendingQueueState;

  const beginRefreshInFlight = useCallback(
    (scope: TimelinePendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      beginTimelinePendingRefreshBlock(pending, scope);
      publishPendingQueueState();
    },
    [pendingSavesRefsRef, publishPendingQueueState],
  );

  const finishRefreshInFlight = useCallback(
    (scope: TimelinePendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      finishTimelinePendingRefreshBlock(pending, scope);
      publishPendingQueueState();
      pendingSavesRefsRef.current.schedulePendingReplayRef.current();
    },
    [pendingSavesRefsRef, publishPendingQueueState],
  );
  const beginRefreshInFlightRef = useRef(beginRefreshInFlight);
  beginRefreshInFlightRef.current = beginRefreshInFlight;
  const finishRefreshInFlightRef = useRef(finishRefreshInFlight);
  finishRefreshInFlightRef.current = finishRefreshInFlight;

  const beginSave = useCallback(() => {
    pendingSavesRefsRef.current.pendingOpsRef.current += 1;
    publishSaveStatePresentation(
      pendingSavesRefsRef.current.pendingQueueRef.current,
    );
  }, [pendingSavesRefsRef, publishSaveStatePresentation]);

  const finishSave = useCallback(
    (nextState: TimelineSaveStateLabel) => {
      pendingSavesRefsRef.current.pendingOpsRef.current = Math.max(
        0,
        pendingSavesRefsRef.current.pendingOpsRef.current - 1,
      );
      if (nextState === "Conflict") {
        setSaveState("Conflict");
        setSaveStateSecondaryMessage("Conflict requires review.");
        return;
      }
      publishSaveStatePresentation(
        pendingSavesRefsRef.current.pendingQueueRef.current,
      );
    },
    [
      pendingSavesRefsRef,
      publishSaveStatePresentation,
      setSaveState,
      setSaveStateSecondaryMessage,
    ],
  );

  useEffect(() => {
    const hasUnsavedRuntimeWork =
      pendingQueueSnapshot.queuedCount > 0 ||
      pendingQueueSnapshot.inFlightCount > 0 ||
      Object.keys(conflictQueue).length > 0;
    if (!hasUnsavedRuntimeWork) {
      return;
    }
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warnBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", warnBeforeUnload);
    };
  }, [
    conflictQueue,
    pendingQueueSnapshot.inFlightCount,
    pendingQueueSnapshot.queuedCount,
  ]);

  return {
    commands: {
      beginRefreshInFlight,
      beginSave,
      computeSaveStatePresentation,
      finishRefreshInFlight,
      finishSave,
      publishPendingQueueState,
      publishSaveStatePresentation,
    },
    refs: {
      beginRefreshInFlightRef,
      finishRefreshInFlightRef,
      publishPendingQueueStateRef,
    },
  };
}
