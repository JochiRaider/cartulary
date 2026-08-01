import { useCallback, useEffect, useRef } from "react";
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
  pendingSavesRefsRef,
  setPendingQueueSnapshot,
}: {
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly conflictQueueRef: TimelineMutableRef<
    Record<string, LocalConflictState>
  >;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefsRef: TimelineMutableRef<
    WorkbookPendingSavesRefs<TMeta>
  >;
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
        pendingMutationCount: pendingSavesRefsRef.current.pendingOpsRef.current,
      });
    },
    [conflictQueueRef, pendingSavesRefsRef],
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
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      setPendingQueueSnapshot(workbookPendingQueueSnapshot(pending));
      publishSaveStatePresentation(pending, conflicts);
    },
    [
      conflictQueueRef,
      pendingSavesRefsRef,
      publishSaveStatePresentation,
      setPendingQueueSnapshot,
    ],
  );
  const publishPendingQueueStateRef = useRef(publishPendingQueueState);
  publishPendingQueueStateRef.current = publishPendingQueueState;

  const beginRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      beginWorkbookPendingRefreshBlock(pending, scope);
      publishPendingQueueState();
    },
    [pendingSavesRefsRef, publishPendingQueueState],
  );

  const finishRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefsRef.current.pendingQueueRef.current;
      finishWorkbookPendingRefreshBlock(pending, scope);
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
        mutationRuntime.projectSurfaceSaveState("cartulary.view.timeline.v2", {
          primaryLabel: "Conflict",
          secondaryMessage: "Conflict requires review.",
        });
        return;
      }
      publishSaveStatePresentation(
        pendingSavesRefsRef.current.pendingQueueRef.current,
      );
    },
    [mutationRuntime, pendingSavesRefsRef, publishSaveStatePresentation],
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
    refs: {
      beginRefreshInFlightRef,
      finishRefreshInFlightRef,
      publishPendingQueueStateRef,
    },
  };
}
