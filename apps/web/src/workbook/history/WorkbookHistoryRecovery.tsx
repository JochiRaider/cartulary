import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../shared/workbookRecoveryNavigation";
import type { InspectorRecordHistoryAction } from "../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorConfirmation } from "../inspector/presentation/WorkbookInspectorFeedback";
import { useWorkbookRecordHistoryController } from "../inspector/useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryState } from "../inspector/useWorkbookRecordHistoryState";
import { WorkbookRecordHistoryPanel } from "../inspector/WorkbookInspectorRecordHistory";
import { updateWorkbookInspectorSubject } from "../inspector/workbookInspectorSubject";
import { HistoryLookupFeedback } from "./HistoryLookupFeedback";
import { historyOperationStatus } from "./historyOperationPresentation";
import { useWorkbookHistoryRuntime } from "./WorkbookHistoryContext";
import type { WorkbookRecordHistoryOwner } from "./WorkbookRecordHistoryOwner";
import {
  type HistoryOperation,
  historyOperationLabel,
} from "./workbookHistoryOperation";

const empty = [] as const;
const emptySnapshot = () => empty;
const emptySubscribe = () => () => {};
const actions: ReadonlySet<InspectorRecordHistoryAction> = new Set([
  "delete",
  "restore",
  "rollback",
]);

export function WorkbookHistoryRecovery() {
  const runtime = useWorkbookHistoryRuntime();
  const owner = runtime?.history;
  const entries = useSyncExternalStore(
    owner?.subscribe ?? emptySubscribe,
    owner?.getSnapshot ?? emptySnapshot,
  );
  const items: readonly WorkbookRecoveryItem[] = entries.map(
    (entry, order) => ({
      id: entry.attempt.id,
      label: historyOperationLabel(entry.attempt),
      origin: entry.attempt.subject.label,
      sheetRef: { kind: "view_schema", id: entry.attempt.subject.viewSchemaId },
      refreshViews:
        entry.receipt && entry.reconciliation !== "complete"
          ? [entry.attempt.subject.viewSchemaId]
          : [],
      order,
      summary: entry.receipt
        ? entry.reconciliation === "complete"
          ? "Completed"
          : "Saved; refresh required"
        : entry.phase === "uncertain"
          ? "Outcome unconfirmed"
          : entry.phase === "rejected"
            ? "Review required"
            : "In progress",
      attention:
        entry.receipt && entry.reconciliation === "complete"
          ? "completed"
          : entry.receipt ||
              entry.phase === "uncertain" ||
              entry.phase === "rejected"
            ? "attention"
            : "progress",
    }),
  );
  const selected = useWorkbookRecoverySource("history", items);
  if (!owner || !runtime) return null;
  return (
    <WorkbookRecoveryDetail source="history" item={selected}>
      <section aria-label="History action recovery">
        {entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <HistoryRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              owner={owner}
              coordinate={runtime.coordinateHistory.bind(runtime)}
            />
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}

function HistoryRecoveryEntry({
  entry,
  owner,
  coordinate,
}: {
  entry: HistoryOperation;
  owner: WorkbookRecordHistoryOwner;
  coordinate: (recordId: string, signal: AbortSignal) => Promise<number | null>;
}) {
  const [review, setReview] = useState(false);
  const [replacement, setReplacement] = useState(false);
  const mounted = useRef(true);
  useLayoutEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
      owner.cancelReview(entry.attempt.id);
    };
  }, [owner, entry.attempt.id]);
  const label = historyOperationLabel(entry.attempt);
  const bind = () =>
    owner.bindRecovery(entry.attempt.id, {
      isCurrent: () => mounted.current,
      coordinate,
      acknowledged: () => {},
      reconcile: () => owner.refreshSurface(entry.attempt.subject.viewSchemaId),
    });
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  const allowed = owner.permitted(entry.attempt.operation);
  return (
    <article
      aria-label={`${label}: ${entry.attempt.subject.label}`}
      style={{
        borderBlockStart: "var(--ct-border-hairline)",
        paddingBlock: "var(--ct-spacing-md)",
        display: "grid",
        gap: "var(--ct-spacing-sm)",
      }}
    >
      <strong>{label}</strong>
      <span>
        {entry.attempt.subject.surfaceLabel} · {entry.attempt.subject.label}
      </span>
      <p role="status">{historyOperationStatus(entry)}</p>
      {entry.phase === "preparing" ? (
        <HistoryLookupFeedback
          state={entry.checking}
          onContinue={() => void owner.continueChecking(entry.attempt.id)}
          onRestart={() => void owner.continueChecking(entry.attempt.id, true)}
          onCancel={() => owner.cancelChecking(entry.attempt.id)}
        />
      ) : null}
      {review ? (
        <HistoryLookupFeedback
          state={entry.reviewState}
          onContinue={() => void owner.review(entry.attempt.id, true)}
          onRestart={() => void owner.review(entry.attempt.id, true, true)}
          onCancel={() => owner.cancelReview(entry.attempt.id)}
        />
      ) : null}
      {entry.receipt?.kind === "rollback" ? (
        <p>
          This action affected {entry.receipt.affectedRecordIds.length} records.
        </p>
      ) : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={!allowed || pending}
          onClick={() => {
            bind();
            void owner.replay(entry.attempt.id);
          }}
        >
          Replay exact action
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.phase === "acknowledged" && entry.reconciliation !== "complete" ? (
        <>
          <p>
            Refresh requires the {entry.attempt.subject.surfaceLabel} surface to
            be open.
          </p>
          <WorkbookInspectorActionButton
            disabled={pending || entry.reconciliation === "refreshing"}
            onClick={() => void owner.refresh(entry.attempt.id)}
          >
            Refresh completed action
          </WorkbookInspectorActionButton>
        </>
      ) : null}
      <WorkbookInspectorActionButton
        disabled={pending}
        onClick={() => {
          setReview(true);
          void owner.review(entry.attempt.id);
        }}
      >
        Review current history
      </WorkbookInspectorActionButton>
      {entry.failure?.kind === "client_txn_conflict" &&
      owner.canReplace(entry.attempt.id) ? (
        <WorkbookInspectorActionButton
          disabled={pending || !allowed}
          onClick={() => setReplacement(true)}
        >
          Prepare replacement action
        </WorkbookInspectorActionButton>
      ) : null}
      {replacement && owner.canReplace(entry.attempt.id) ? (
        <WorkbookInspectorConfirmation
          operation={label}
          subject={entry.attempt.subject.label}
          confirmLabel="Confirm replacement action"
          destructive={entry.attempt.operation === "delete"}
          onCancel={() => setReplacement(false)}
          onConfirm={() => {
            setReplacement(false);
            bind();
            void owner.retryWithNewId(entry.attempt.id);
          }}
        />
      ) : null}
      {review && !entry.currentHistory && !entry.reviewState ? (
        <p role="status">
          {entry.reviewFailure
            ? "Current history could not be read. Try reviewing again."
            : "Reading current history."}
        </p>
      ) : null}
      {review && entry.currentHistory ? (
        <HistoryCurrentReview entry={entry} owner={owner} />
      ) : null}
      {!pending &&
      (entry.phase === "rejected" ||
        (entry.phase === "acknowledged" &&
          entry.reconciliation === "complete")) ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          onClick={() => owner.dismiss(entry.attempt.id)}
        >
          Dismiss completed review
        </WorkbookInspectorActionButton>
      ) : null}
    </article>
  );
}

function HistoryCurrentReview({
  entry,
  owner,
}: {
  entry: HistoryOperation;
  owner: WorkbookRecordHistoryOwner;
}) {
  const history = entry.currentHistory;
  const subject = history
    ? updateWorkbookInspectorSubject(entry.attempt.subject, {
        kind: history.deleted ? "deleted" : "live",
        recordId: history.record_id,
        rowVersion: history.row_version,
      })
    : null;
  const controller = useWorkbookRecordHistoryController({
    presentation: useWorkbookRecordHistoryState(),
    subject,

    canMutate: entry.phase === "acknowledged" || entry.phase === "rejected",
    ownerEffects: {
      deleteAccepted: () => {},
      restoreAccepted: () => {},
      rollbackAccepted: () => {},
      refresh: () => owner.refreshSurface(entry.attempt.subject.viewSchemaId),
    },
  });
  const openHistory = controller.commands.open;
  useEffect(() => {
    openHistory();
  }, [openHistory]);
  return (
    <WorkbookRecordHistoryPanel
      browsingControls={controller.commands}
      actions={actions}
      canMutate={entry.phase === "acknowledged" || entry.phase === "rejected"}
      idleRecordId={subject?.recordId}
      state={controller.snapshot}
      onOpenHistory={controller.commands.open}
      onCancelPendingAction={controller.commands.cancel}
      onConfirmPendingAction={() => void controller.commands.confirm()}
      onPreviewDeleteRestore={controller.commands.previewDeleteRestore}
      onPreviewRollback={controller.commands.previewRollback}
    />
  );
}
