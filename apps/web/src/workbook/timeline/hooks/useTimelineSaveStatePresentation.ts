import { useCallback, useRef } from "react";
import { type SheetRef, sheetRefsEqual } from "../../../shared/sheetRef";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  beginWorkbookPendingRefreshBlock,
  finishWorkbookPendingRefreshBlock,
  type WorkbookPendingRefreshBlockScope,
} from "../../runtime/workbookPendingReplayRuntime";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";

export function useTimelineSaveStatePresentation({
  mutationRuntime,
  pendingSavesRefs,
  sheetRef,
}: {
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly pendingSavesRefs: TimelinePendingSavesRefs;
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
  const beginRefreshInFlight = useCallback(
    (scope: WorkbookPendingRefreshBlockScope) => {
      const pending = pendingSavesRefs.pendingQueueRef.current;
      beginWorkbookPendingRefreshBlock(pending, scope);
      const finishReport = mutationRuntime.beginRefreshStatus(originSheetRef);
      let finished = false;
      return () => {
        if (finished) return;
        finished = true;
        finishWorkbookPendingRefreshBlock(pending, scope);
        finishReport();
        mutationRuntime.requestDrain();
      };
    },
    [mutationRuntime, originSheetRef, pendingSavesRefs],
  );

  return {
    commands: {
      beginRefreshInFlight,
      publishSaveStatePresentation,
    },
  };
}
