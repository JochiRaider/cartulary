import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import type { InspectorRecordHistoryAction } from "../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorConfirmation } from "../inspector/presentation/WorkbookInspectorFeedback";
import { useWorkbookRecordHistoryController } from "../inspector/useWorkbookRecordHistoryController";
import { WorkbookRecordHistoryPanel } from "../inspector/WorkbookInspectorRecordHistory";
import { updateWorkbookInspectorSubject } from "../inspector/workbookInspectorSubject";
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
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const summary = useRef<HTMLElement>(null);
  const focusRequested = useRef(false);
  useLayoutEffect(() => {
    if (open && focusRequested.current) {
      focusRequested.current = false;
      summary.current?.focus({ preventScroll: true });
    }
  }, [open]);
  if (!owner || entries.length === 0) return null;
  const close = () => {
    const restoreFocus = summary.current?.parentElement?.contains(
      document.activeElement,
    );
    setOpen(false);
    if (restoreFocus) trigger.current?.focus({ preventScroll: true });
  };
  return (
    <div style={{ position: "relative" }}>
      <WorkbookInspectorActionButton
        ref={trigger}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            focusRequested.current = true;
            setOpen(true);
          }
        }}
      >
        History actions ({entries.length})
      </WorkbookInspectorActionButton>
      {open ? (
        <section
          aria-label="History action recovery"
          style={{
            position: "absolute",
            zIndex: 30,
            insetInlineEnd: 0,
            inlineSize: "min(34rem, 90vw)",
            maxBlockSize: "75vh",
            overflow: "auto",
            padding: "var(--ct-spacing-md)",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            boxShadow: "var(--ct-elevation-popover)",
          }}
          onKeyDown={(event) => {
            if (
              event.key === "Escape" &&
              !(
                event.target instanceof Element &&
                event.target.closest('[role="alertdialog"]')
              )
            ) {
              event.stopPropagation();
              close();
            }
          }}
        >
          <section
            ref={summary}
            aria-label="History action recovery summary"
            tabIndex={-1}
          >
            <strong>History actions</strong>
            <p>
              Recover retained actions or review current history. Closing this
              panel keeps admitted actions.
            </p>
          </section>
          <WorkbookInspectorActionButton onClick={close}>
            Close history actions
          </WorkbookInspectorActionButton>
          {entries.map((entry) => (
            <HistoryRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              owner={owner}
              coordinate={runtime.coordinateHistory.bind(runtime)}
            />
          ))}
        </section>
      ) : null}
    </div>
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
    };
  }, []);
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
      {review && !entry.currentHistory ? (
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
    owner,
    subject,
    ...(history ? { initialHistory: history } : {}),
    canMutate: entry.phase === "acknowledged" || entry.phase === "rejected",
    ownerEffects: {
      deleteAccepted: () => {},
      restoreAccepted: () => {},
      rollbackAccepted: () => {},
      refresh: () => owner.refreshSurface(entry.attempt.subject.viewSchemaId),
    },
  });
  return (
    <WorkbookRecordHistoryPanel
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
