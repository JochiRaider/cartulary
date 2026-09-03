import type { WorkbookPendingQueueRuntime } from "../../runtime/workbookPendingReplayRuntime";
import type {
  TimelineMutableRef,
  TimelineReplayContext,
} from "./timelineControllerPorts";

export type TimelinePendingSavesRefs = {
  readonly collectionKeyboardCommitRef: TimelineMutableRef<Map<string, string>>;
  readonly pendingOpsRef: TimelineMutableRef<number>;
  readonly pendingQueueRef: TimelineMutableRef<WorkbookPendingQueueRuntime>;
  readonly pendingReplayOrderRef: TimelineMutableRef<number>;
  readonly replayContextByUnitId: Map<string, TimelineReplayContext>;
  readonly pendingSignaturesRef: TimelineMutableRef<Map<string, string>>;
  readonly saveQueueRef: TimelineMutableRef<Promise<void>>;
};

const refsByMutationRuntime = new WeakMap<object, TimelinePendingSavesRefs>();

/** Keeps Timeline-owned queue context for the lifetime of the shell runtime. */
export function timelinePendingSavesRefsFor(
  mutationRuntime: object,
  pendingQueue: WorkbookPendingQueueRuntime,
): TimelinePendingSavesRefs {
  const existing = refsByMutationRuntime.get(mutationRuntime);
  if (existing !== undefined) {
    existing.pendingQueueRef.current = pendingQueue;
    return existing;
  }
  const refs: TimelinePendingSavesRefs = {
    collectionKeyboardCommitRef: { current: new Map() },
    pendingOpsRef: { current: 0 },
    pendingQueueRef: { current: pendingQueue },
    pendingReplayOrderRef: { current: 1 },
    replayContextByUnitId: new Map(),
    pendingSignaturesRef: { current: new Map() },
    saveQueueRef: { current: Promise.resolve() },
  };
  refsByMutationRuntime.set(mutationRuntime, refs);
  return refs;
}
