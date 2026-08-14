import { useEffect, useMemo, useRef, useState } from "react";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import {
  type WorkbookPendingQueueRuntime,
  type WorkbookPendingQueueSnapshot,
  workbookPendingQueueSnapshot,
} from "../../runtime/workbookPendingReplayRuntime";

export function useTimelinePendingSaves<TMeta>({
  mutationRuntime,
}: {
  readonly mutationRuntime: WorkbookMutationRuntime;
}) {
  const pendingOpsRef = useRef(0);
  const pendingSignaturesRef = useRef(new Map<string, string>());
  const collectionKeyboardCommitRef = useRef(new Map<string, string>());
  const pendingSocketTxnTimeoutsRef = useRef(new Map<string, number>());
  const saveQueueRef = useRef(Promise.resolve());
  const sharedPendingRuntime = mutationRuntime.pending<TMeta>();
  const pendingQueueRef =
    useRef<WorkbookPendingQueueRuntime<TMeta>>(sharedPendingRuntime);
  if (pendingQueueRef.current !== sharedPendingRuntime) {
    pendingQueueRef.current = sharedPendingRuntime;
  }
  const pendingReplayOrderRef = useRef(1);
  const pendingReplayTimerRef = useRef<number | null>(null);
  const refs = useMemo(
    () => ({
      collectionKeyboardCommitRef,
      pendingOpsRef,
      pendingQueueRef,
      pendingReplayOrderRef,
      pendingReplayTimerRef,
      pendingSignaturesRef,
      pendingSocketTxnTimeoutsRef,
      saveQueueRef,
    }),
    [],
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
