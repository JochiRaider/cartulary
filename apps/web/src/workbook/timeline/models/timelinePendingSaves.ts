import type { GridEditCommitOutcome } from "@cartulary/grid-adapter";
import type { WorkbookPendingQueueRuntime } from "../../runtime/workbookPendingReplayRuntime";
import type {
  TimelineMutableRef,
  TimelineReplayContext,
} from "./timelineControllerPorts";

export type TimelinePendingSavesRefs = {
  readonly collectionCommits: Map<
    string,
    {
      readonly revision: number;
      readonly value: string;
      outcome?: GridEditCommitOutcome;
      readonly listeners: Set<(outcome: GridEditCommitOutcome) => void>;
    }
  >;
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
    collectionCommits: new Map(),
    pendingQueueRef: { current: pendingQueue },
    pendingReplayOrderRef: { current: 1 },
    replayContextByUnitId: new Map(),
    pendingSignaturesRef: { current: new Map() },
    saveQueueRef: { current: Promise.resolve() },
  };
  refsByMutationRuntime.set(mutationRuntime, refs);
  return refs;
}
