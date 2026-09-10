import { useSyncExternalStore } from "react";
import { HistoryLookupFeedback } from "./HistoryLookupFeedback";
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
    <div>
      <p role="status">
        {historyOperationStatus(entry)} Open History actions for recovery.
      </p>
      {entry.phase === "preparing" && owner ? (
        <HistoryLookupFeedback
          state={entry.checking}
          onContinue={() => void owner.continueChecking(entry.attempt.id)}
          onRestart={() => void owner.continueChecking(entry.attempt.id, true)}
          onCancel={() => owner.cancelChecking(entry.attempt.id)}
        />
      ) : null}
    </div>
  );
}
