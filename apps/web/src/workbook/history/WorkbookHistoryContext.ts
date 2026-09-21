import {
  createContext,
  useCallback,
  useContext,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";

export const WorkbookHistoryContext = createContext<Pick<
  WorkbookMutationRuntime,
  "history" | "coordinateHistory"
> | null>(null);
export function useWorkbookHistoryRuntime() {
  return useContext(WorkbookHistoryContext);
}

export function useWorkbookHistorySurfaceRefresh(
  viewSchemaId: string,
  refresh: () => Promise<void> | void,
) {
  const runtime = useWorkbookHistoryRuntime();
  const refreshRef = useRef(refresh);
  useLayoutEffect(() => {
    refreshRef.current = refresh;
  });
  const refreshCurrent = useCallback(() => refreshRef.current(), []);
  useLayoutEffect(
    () => runtime?.history.registerSurface(viewSchemaId, refreshCurrent),
    [runtime, viewSchemaId, refreshCurrent],
  );
}

const emptySnapshot = () => null;
const emptySubscription = () => () => {};
export function useHistoryActionPermission(
  operation: "delete" | "restore" | "rollback",
) {
  const runtime = useWorkbookHistoryRuntime();
  useSyncExternalStore(
    runtime?.history.subscribe ?? emptySubscription,
    runtime?.history.getSnapshot ?? emptySnapshot,
  );
  return runtime?.history.permitted(operation) === true;
}

export function useHistoryRecordPending(recordId: string | null) {
  const runtime = useWorkbookHistoryRuntime();
  const entries = useSyncExternalStore(
    runtime?.history.subscribe ?? emptySubscription,
    runtime?.history.getSnapshot ?? emptySnapshot,
  );
  return (
    entries?.some(
      (entry) =>
        entry.attempt.subject.recordId === recordId &&
        (entry.transportPending ||
          entry.phase === "preparing" ||
          entry.phase === "submitting" ||
          entry.phase === "uncertain" ||
          (entry.phase === "acknowledged" &&
            (entry.reconciliation === "pending" ||
              entry.reconciliation === "refreshing"))),
    ) === true
  );
}
