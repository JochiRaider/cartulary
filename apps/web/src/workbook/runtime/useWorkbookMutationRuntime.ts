import { useSyncExternalStore } from "react";
import type {
  WorkbookMutationRuntime,
  WorkbookMutationSnapshot,
} from "./WorkbookMutationRuntime";

export function useWorkbookMutationRuntime(
  runtime: WorkbookMutationRuntime,
): WorkbookMutationSnapshot {
  return useSyncExternalStore(
    runtime.subscribe,
    runtime.getSnapshot,
    runtime.getSnapshot,
  );
}
