import { useCallback, useEffect } from "react";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  beginWorkbookPendingRefreshBlock,
  finishWorkbookPendingRefreshBlock,
  type WorkbookPendingQueueRuntime,
  type WorkbookPendingQueueSnapshot,
  type WorkbookPendingRefreshBlockScope,
  type WorkbookPendingSavesRefs,
  workbookPendingQueueSnapshot,
} from "../../runtime/workbookPendingReplayRuntime";
import {
  deriveWorkbookSaveState,
  type WorkbookSaveStateConflictAnchor,
} from "../../utils/workbookPendingQueue";
import type { TimelineMutableRef } from "../models/timelineControllerPorts";
import type { LocalConflictState } from "../models/workbookTimelineModel";

type TimelineSaveStateLabel = "Conflict" | "Saved" | "Syncing";

type TimelineSetState<T> = (value: T) => void;

function saveStateConflictAnchorsFromLocalConflicts(
  conflicts: Record<string, LocalConflictState>,
): WorkbookSaveStateConflictAnchor[] {
  return Object.values(conflicts).map((entry) => ({ ...entry.anchor }));
}

export function useTimelineSaveStatePresentation<TMeta>({
  conflictQueue,
  conflictQueueRef,
  mutationRuntime,
  pendingQueueSnapshot,
  pendingSavesRefs,
  setPendingQueueSnapshot,
}: {
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefs: WorkbookPendingSavesRefs<TMeta>;
  readonly setPendingQueueSnapshot: TimelineSetState<WorkbookPendingQueueSnapshot>;
}) {
  const computeSaveStatePresentation = useCallback(
    (
      pending: WorkbookPendingQueueRuntime<TMeta>,
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
        pendingMutationCount: pendingSavesRefs.pendingOpsRef.current,
      });
    },
    [conflictQueueRef, pendingSavesRefs],
  );

  const publishSaveStatePresentation = useCallback(
    (
      pending: WorkbookPendingQueueRuntime<TMeta>,
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const presentation = computeSaveStatePresentation(pending, conflicts);
      mutationRuntime.projectSurfaceSaveState("cartulary.view.timeline.v2", {
        primaryLabel: presentation.primaryLabel,
        secondaryMessage: presentation.secondaryMessage,
      });
      return presentation;
    },
    [computeSaveStatePresentation, conflictQueueRef, mutationRuntime],
  );

  const publishPendingQueueState = useCallback(
    (
      conflicts: Record<string, LocalConflictState> = conflictQueueRef.current,
    ) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      setPendingQueueSnapshot(workbookPendingQueueSnapshot(pending));
      publishSaveStatePresentation(pending, conflicts);
    },
    [
      conflictQueueRef,
      pendingSavesRefs,
      publishSaveStatePresentation,
      setPendingQueueSnapshot,
    ],
  );
  const beginRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      beginWorkbookPendingRefreshBlock(pending, scope);
      publishPendingQueueState();
    },
    [pendingSavesRefs, publishPendingQueueState],
  );

  const finishRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      finishWorkbookPendingRefreshBlock(pending, scope);
      publishPendingQueueState();
      mutationRuntime.requestDrain();
    },
    [mutationRuntime, pendingSavesRefs, publishPendingQueueState],
  );
  const beginSave = useCallback(() => {
    pendingSavesRefs.pendingOpsRef.current += 1;
    publishSaveStatePresentation(pendingSavesRefs.pendingQueueRef.current);
  }, [pendingSavesRefs, publishSaveStatePresentation]);

  const finishSave = useCallback(
    (nextState: TimelineSaveStateLabel) => {
      pendingSavesRefs.pendingOpsRef.current = Math.max(
        0,
        pendingSavesRefs.pendingOpsRef.current - 1,
      );
      if (nextState === "Conflict") {
        mutationRuntime.projectSurfaceSaveState("cartulary.view.timeline.v2", {
          primaryLabel: "Conflict",
          secondaryMessage: "Conflict requires review.",
        });
        return;
      }
      publishSaveStatePresentation(pendingSavesRefs.pendingQueueRef.current);
    },
    [mutationRuntime, pendingSavesRefs, publishSaveStatePresentation],
  );

  useEffect(
    () => () => {
      mutationRuntime.clearSurfaceSaveState("cartulary.view.timeline.v2");
    },
    [mutationRuntime],
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
  };
}
