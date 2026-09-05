import { useMemo, useSyncExternalStore } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
import {
  projectWorkbookStatusForSurface,
  type WorkbookStatusPresentation,
} from "./workbookMutationStatusProjector";

export function useWorkbookMutationRuntime(
  runtime: WorkbookMutationRuntime,
  activeSheetRef?: SheetRef,
): WorkbookStatusPresentation {
  const snapshot = useSyncExternalStore(
    runtime.subscribe,
    runtime.getSnapshot,
    runtime.getSnapshot,
  );
  return useMemo(
    () => projectWorkbookStatusForSurface(snapshot, activeSheetRef),
    [snapshot, activeSheetRef],
  );
}
