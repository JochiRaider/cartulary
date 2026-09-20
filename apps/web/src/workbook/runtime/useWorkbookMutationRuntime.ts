import { useMemo, useSyncExternalStore } from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookConflictEntry } from "./workbookConflictModel";
import {
  projectWorkbookStatusForSurface,
  type WorkbookMutationSnapshot,
  type WorkbookStatusPresentation,
} from "./workbookMutationStatusProjector";

export type WorkbookMutationStatusSource = {
  readonly subscribe: (listener: () => void) => () => void;
  readonly getSnapshot: () => WorkbookMutationSnapshot;
};

export function useWorkbookMutationRuntime(
  runtime: WorkbookMutationStatusSource,
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

/** Grid coordination observes conflict changes for its schema, not save progress. */
export function useWorkbookMutationConflicts(
  source: WorkbookMutationStatusSource,
  viewSchemaId: string,
): readonly WorkbookConflictEntry[] {
  const read = useMemo(() => {
    let previous: readonly WorkbookConflictEntry[] | undefined;
    let selected: readonly WorkbookConflictEntry[] = [];
    return () => {
      const current = source.getSnapshot().conflicts;
      if (current === previous) return selected;
      previous = current;
      const next = current.filter(
        (entry) => entry.origin.viewSchemaId === viewSchemaId,
      );
      if (
        next.length !== selected.length ||
        next.some((entry, index) => entry !== selected[index])
      )
        selected = Object.freeze(next);
      return selected;
    };
  }, [source, viewSchemaId]);
  return useSyncExternalStore(source.subscribe, read, read);
}
