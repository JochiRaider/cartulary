import { useEffect, useState } from "react";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  type WorkbookPendingQueueSnapshot,
  workbookPendingQueueSnapshot,
} from "../../runtime/workbookPendingReplayRuntime";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";

export function useTimelinePendingSaves({
  mutationRuntime,
}: {
  readonly mutationRuntime: WorkbookMutationRuntime;
}) {
  const sharedPendingRuntime = mutationRuntime.pendingQueue();
  const refs = timelinePendingSavesRefsFor(
    mutationRuntime,
    sharedPendingRuntime,
  );
  const [pendingQueueSnapshot, setPendingQueueSnapshot] =
    useState<WorkbookPendingQueueSnapshot>(() =>
      workbookPendingQueueSnapshot(sharedPendingRuntime),
    );
  useEffect(() => {
    setPendingQueueSnapshot(workbookPendingQueueSnapshot(sharedPendingRuntime));
  }, [sharedPendingRuntime]);

  return {
    commands: {
      setPendingQueueSnapshot,
    },
    refs,
    snapshot: {
      pendingQueueSnapshot,
    },
  };
}
