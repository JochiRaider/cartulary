import { useMemo, useSyncExternalStore } from "react";
import { selectWorkbookStatusSecondary } from "../utils/workbookStatusSecondary";
import type {
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "./WorkbookMutationRuntime";

export function useWorkbookMutationRuntime(
  runtime: WorkbookMutationRuntime,
  activeSurfaceId?: string,
): WorkbookMutationSnapshot {
  const snapshot = useSyncExternalStore(
    runtime.subscribe,
    runtime.getSnapshot,
    runtime.getSnapshot,
  );
  return useMemo(() => {
    if (activeSurfaceId === undefined) return snapshot;
    return {
      ...snapshot,
      secondaryMessage:
        selectWorkbookStatusSecondary(
          snapshot.secondaryCandidates,
          activeSurfaceId,
        )?.message ?? null,
    };
  }, [activeSurfaceId, snapshot]);
}
