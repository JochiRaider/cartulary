import { useCallback, useEffect, useRef } from "react";
import { type SheetRef, sheetRefsEqual } from "../../../shared/sheetRef";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  beginWorkbookPendingRefreshBlock,
  finishWorkbookPendingRefreshBlock,
  type WorkbookPendingQueueSnapshot,
  type WorkbookPendingRefreshBlockScope,
  workbookPendingQueueSnapshot,
} from "../../runtime/workbookPendingReplayRuntime";
import type { LocalConflictState } from "../models/timelineConflictState";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";

export function useTimelineSaveStatePresentation({
  conflictQueue,
  mutationRuntime,
  pendingQueueSnapshot,
  pendingSavesRefs,
  setPendingQueueSnapshot,
  sheetRef,
}: {
  readonly conflictQueue: Record<string, LocalConflictState>;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly pendingQueueSnapshot: WorkbookPendingQueueSnapshot;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
  readonly setPendingQueueSnapshot: (
    snapshot: WorkbookPendingQueueSnapshot,
  ) => void;
  readonly sheetRef: SheetRef;
}) {
  const originRef = useRef(sheetRef);
  if (!sheetRefsEqual(originRef.current, sheetRef))
    originRef.current = sheetRef;
  const originSheetRef = originRef.current;
  const publishSaveStatePresentation = useCallback(
    () => mutationRuntime.notifyPendingChanged(),
    [mutationRuntime],
  );
  const publishPendingQueueState = useCallback(() => {
    setPendingQueueSnapshot(
      workbookPendingQueueSnapshot(pendingSavesRefs.pendingQueueRef.current),
    );
    mutationRuntime.notifyPendingChanged();
  }, [mutationRuntime, pendingSavesRefs, setPendingQueueSnapshot]);
  const beginRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      beginWorkbookPendingRefreshBlock(pending, scope);
      const finishReport = mutationRuntime.beginRefreshStatus(originSheetRef);
      publishPendingQueueState();
      let finished = false;
      return () => {
        if (finished) return;
        finished = true;
        finishWorkbookPendingRefreshBlock(pending, scope);
        finishReport();
        publishPendingQueueState();
        mutationRuntime.requestDrain();
      };
    },
    [
      mutationRuntime,
      originSheetRef,
      pendingSavesRefs,
      publishPendingQueueState,
    ],
  );
  const beginSave = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );

  useEffect(() => {
    if (
      pendingQueueSnapshot.queuedCount + pendingQueueSnapshot.inFlightCount ===
        0 &&
      Object.keys(conflictQueue).length === 0
    )
      return;
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warnBeforeUnload);
    return () => window.removeEventListener("beforeunload", warnBeforeUnload);
  }, [
    conflictQueue,
    pendingQueueSnapshot.inFlightCount,
    pendingQueueSnapshot.queuedCount,
  ]);

  return {
    commands: {
      beginRefreshInFlight,
      beginSave,
      publishPendingQueueState,
      publishSaveStatePresentation,
    },
  };
}
