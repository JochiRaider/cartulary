import { useSyncExternalStore } from "react";
import { historyOperationStatus } from "./historyOperationPresentation";
import { useWorkbookHistoryRuntime } from "./WorkbookHistoryContext";

const empty = [] as const;
const emptySnapshot = () => empty;
const emptySubscribe = () => () => {};

export function WorkbookHistoryLocalStatus({ recordId }: { recordId: string }) {
  const owner = useWorkbookHistoryRuntime()?.history;
  const entries = useSyncExternalStore(
    owner?.subscribe ?? emptySubscribe,
    owner?.getSnapshot ?? emptySnapshot,
  );
  const entry = entries.find(
    (entry) => entry.attempt.subject.recordId === recordId,
  );
  if (
    !entry ||
    entry.phase === "rejected" ||
    (entry.phase === "acknowledged" && entry.reconciliation === "complete")
  )
    return null;
  return (
    <p role="status">
      {historyOperationStatus(entry)} Open History actions for recovery.
    </p>
  );
}
