import {
  useCallback,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  useWorkbookRecoveryNavigation,
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import {
  type WorkbookRecoveryItem,
  workbookRecoveryKey,
} from "../../shared/workbookRecoveryNavigation";
import { WorkbookHistoryReview } from "../history/WorkbookHistoryReview";
import {
  type HistoryReviewLocator,
  historyReviewAuthorized,
} from "../history/workbookHistoryReview";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookBatchSnapshot } from "../runtime/workbookBatchOperation";
import { workbookBatchOutcome } from "../runtime/workbookBatchOutcome";
import { workbookBatchRecoveryItems } from "../runtime/workbookBatchRecoveryItems";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";
import { WorkbookBatchRecordChoices } from "./WorkbookBatchRecordChoices";
import { inputStyle } from "./workbookGridControlStyles";

type BatchRetryPresentation = {
  readonly batchId: string;
  readonly action: "mutation" | "refresh";
  readonly authority: NonNullable<WorkbookBatchSnapshot["authority"]>;
  readonly activation: number;
  readonly control: HTMLButtonElement;
  ownsFocus: boolean;
};

export function WorkbookBatchRecovery({
  runtime,
  activateConflict,
}: {
  readonly runtime: WorkbookMutationRuntime;
  readonly activateConflict?:
    | ((invoker: HTMLButtonElement, action: WorkbookStatusAction) => void)
    | undefined;
}) {
  useSyncExternalStore(runtime.history.subscribe, runtime.history.getSnapshot);
  const owner = runtime.batches;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const navigation = useWorkbookRecoveryNavigation();
  const mutation = useSyncExternalStore(runtime.subscribe, runtime.getSnapshot);
  const [retryPresentation, setRetryPresentation] =
    useState<BatchRetryPresentation | null>(null);
  const retryPresentationRef = useRef<BatchRetryPresentation | null>(null);
  const outcomeRef = useRef<HTMLParagraphElement | null>(null);
  const originalInputRef = useRef<HTMLTextAreaElement | null>(null);
  const outcomeId = useId();
  const clearRetryPresentation = useCallback(() => {
    retryPresentationRef.current = null;
    setRetryPresentation(null);
  }, []);
  const conflictCounts = new Map<string, number>();
  for (const conflict of mutation.conflicts) {
    const id = conflict.batchOperationId;
    if (id) conflictCounts.set(id, (conflictCounts.get(id) ?? 0) + 1);
  }
  const [review, setReview] = useState<{
    readonly batchId: string;
    readonly locator: HistoryReviewLocator;
  } | null>(null);
  const invoker = useRef<HTMLElement | null>(null);
  const restoreFocus = useRef(false);
  const currentReview =
    snapshot.authority &&
    review &&
    historyReviewAuthorized(review.locator, runtime.history.readScope)
      ? review
      : null;
  useLayoutEffect(() => {
    if (review && !currentReview) {
      setReview(null);
      invoker.current = null;
    }
  }, [review, currentReview]);
  const items: WorkbookRecoveryItem[] = [
    ...workbookBatchRecoveryItems(snapshot, conflictCounts),
  ];
  if (
    currentReview &&
    !items.some((item) => item.id === currentReview.batchId)
  ) {
    items.push({
      id: currentReview.batchId,
      label: "Change review",
      summary: "Reviewing the selected Timeline record",
      origin: "Timeline",
      sheetRef: { kind: "view_schema", id: currentReview.locator.viewSchemaId },
      attention: "completed",
      order: Number.MAX_SAFE_INTEGER,
    });
  }
  const selected = useWorkbookRecoverySource("batch", items, {
    detach: () => {
      clearRetryPresentation();
      setReview(null);
      invoker.current = null;
    },
  });
  useLayoutEffect(() => {
    const intent = retryPresentationRef.current;
    if (!intent) return;
    const attachment = navigation?.getSnapshot();
    const entry = snapshot.entries.find((item) => item.id === intent.batchId);
    if (
      snapshot.authority !== intent.authority ||
      selected !== intent.batchId ||
      !attachment?.open ||
      attachment.selected !== workbookRecoveryKey("batch", intent.batchId) ||
      attachment.activation !== intent.activation ||
      !entry
    ) {
      clearRetryPresentation();
      return;
    }
    const pending =
      intent.action === "mutation"
        ? entry.transportPending
        : entry.reconciliation === "refreshing";
    if (pending) return;
    const completed =
      intent.action === "mutation"
        ? entry.phase === "acknowledged"
        : entry.reconciliation === "complete";
    if (
      completed &&
      intent.ownsFocus &&
      document.activeElement === intent.control &&
      outcomeRef.current?.isConnected
    )
      outcomeRef.current.focus({ preventScroll: true });
    if (
      completed &&
      !intent.ownsFocus &&
      (document.activeElement === intent.control ||
        document.activeElement === originalInputRef.current ||
        document.activeElement === document.body)
    )
      return;
    clearRetryPresentation();
  });
  useLayoutEffect(() => {
    if (!retryPresentation) return;
    const revokeFocus = () => {
      const current = retryPresentationRef.current;
      if (current === retryPresentation) current.ownsFocus = false;
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Tab") revokeFocus();
    };
    const onFocusIn = (event: FocusEvent) => {
      if (event.target === retryPresentation.control) return;
      revokeFocus();
      const current = retryPresentationRef.current;
      const entry = owner
        .getSnapshot()
        .entries.find((item) => item.id === current?.batchId);
      if (
        current === retryPresentation &&
        entry &&
        (current.action === "mutation"
          ? !entry.transportPending
          : entry.reconciliation !== "refreshing")
      )
        clearRetryPresentation();
    };
    document.addEventListener("keydown", onKeyDown, true);
    document.addEventListener("focusin", onFocusIn, true);
    document.addEventListener("pointerdown", revokeFocus, true);
    document.addEventListener("scroll", revokeFocus, true);
    document.addEventListener("wheel", revokeFocus, true);
    return () => {
      document.removeEventListener("keydown", onKeyDown, true);
      document.removeEventListener("focusin", onFocusIn, true);
      document.removeEventListener("pointerdown", revokeFocus, true);
      document.removeEventListener("scroll", revokeFocus, true);
      document.removeEventListener("wheel", revokeFocus, true);
    };
  }, [retryPresentation, owner, clearRetryPresentation]);
  const retry = (
    batchId: string,
    action: BatchRetryPresentation["action"],
    control: HTMLButtonElement,
  ) => {
    const authority = owner.getSnapshot().authority;
    const entry = owner
      .getSnapshot()
      .entries.find((item) => item.id === batchId);
    const attachment = navigation?.getSnapshot();
    if (
      !authority ||
      !entry ||
      retryPresentationRef.current ||
      !attachment?.open ||
      attachment.selected !== workbookRecoveryKey("batch", batchId) ||
      (action === "mutation"
        ? entry.phase !== "uncertain" ||
          entry.transportPending ||
          !owner.canWrite()
        : entry.phase !== "acknowledged" || entry.reconciliation !== "required")
    )
      return;
    const intent: BatchRetryPresentation = {
      batchId,
      action,
      authority,
      activation: attachment.activation,
      control,
      ownsFocus: document.activeElement === control,
    };
    retryPresentationRef.current = intent;
    setRetryPresentation(intent);
    owner.retry(batchId);
  };
  useLayoutEffect(() => {
    // A detached DOM node can retain React props and the pruned receipt closure.
    if (invoker.current && !invoker.current.isConnected) invoker.current = null;
    if (!restoreFocus.current) return;
    restoreFocus.current = false;
    const target = invoker.current;
    if (
      target?.isConnected &&
      !target.closest("[hidden], [inert], [disabled]") &&
      (document.activeElement === document.body ||
        document.activeElement?.closest("#workbook-recovery-panel"))
    )
      target.focus({ preventScroll: true });
    invoker.current = null;
  });
  const entries = snapshot.entries;
  return (
    <WorkbookRecoveryDetail source="batch" item={selected}>
      <section aria-label="Batch action recovery">
        <style>{`.workbook-batch-outcome:focus { outline: var(--ct-component-focus-ring-border); outline-offset: var(--ct-component-focus-ring-offset); }`}</style>
        {snapshot.admissionError ? (
          <p role="alert">{snapshot.admissionError}</p>
        ) : null}
        {entries
          .filter((entry) => entry.id === selected)
          .map((entry) => {
            const index = entries.indexOf(entry);
            const conflicts = runtime
              .getSnapshot()
              .conflicts.filter(
                (conflict) => conflict.batchOperationId === entry.id,
              );
            const firstConflict = conflicts[0];
            const outcome = workbookBatchOutcome(entry, conflicts.length);
            const { label } = outcome;
            const attachedRetry =
              retryPresentation?.batchId === entry.id &&
              retryPresentation.authority === snapshot.authority;
            const replayRetry =
              attachedRetry && retryPresentation.action === "mutation";
            const refreshRetry =
              attachedRetry && retryPresentation.action === "refresh";
            const replayPending = replayRetry && entry.transportPending;
            const refreshPending =
              refreshRetry && entry.reconciliation === "refreshing";
            return (
              <section key={entry.id} aria-label={`${label} ${index + 1}`}>
                <p
                  className="workbook-batch-outcome"
                  role="status"
                  aria-label={`${label} outcome`}
                  id={outcomeId}
                  tabIndex={-1}
                  ref={outcomeRef}
                >
                  {outcome.detail}
                </p>
                {outcome.refresh !== "none" ? (
                  <p role="status">
                    {outcome.refresh === "refreshing"
                      ? "Refreshing the view…"
                      : outcome.refresh === "required"
                        ? "The view could not refresh. The acknowledged result is unchanged."
                        : "View refresh is pending."}
                  </p>
                ) : null}
                {entry.phase === "rejected" ||
                entry.phase === "uncertain" ||
                replayRetry ? (
                  <label
                    style={{ display: "grid", gap: "var(--ct-spacing-xs)" }}
                  >
                    Original input
                    <textarea
                      ref={originalInputRef}
                      readOnly
                      aria-label="Original batch input"
                      value={outcome.originalInput}
                      style={{
                        ...inputStyle,
                        maxInlineSize: "100%",
                        minInlineSize: 0,
                      }}
                    />
                  </label>
                ) : null}
                {entry.phase === "uncertain" || replayRetry ? (
                  <WorkbookInspectorActionButton
                    aria-busy={replayPending}
                    aria-disabled={
                      !owner.canWrite() ||
                      replayPending ||
                      entry.phase !== "uncertain"
                    }
                    aria-describedby={
                      entry.phase === "acknowledged" ? outcomeId : undefined
                    }
                    onClick={(event) =>
                      retry(entry.id, "mutation", event.currentTarget)
                    }
                  >
                    Retry {label.toLowerCase()}
                  </WorkbookInspectorActionButton>
                ) : null}
                {entry.phase === "acknowledged" &&
                (entry.reconciliation === "required" || refreshRetry) ? (
                  <WorkbookInspectorActionButton
                    aria-busy={refreshPending}
                    aria-disabled={
                      refreshPending || entry.reconciliation !== "required"
                    }
                    aria-describedby={
                      entry.reconciliation === "complete"
                        ? outcomeId
                        : undefined
                    }
                    onClick={(event) =>
                      retry(entry.id, "refresh", event.currentTarget)
                    }
                  >
                    Retry refresh
                  </WorkbookInspectorActionButton>
                ) : null}
                {firstConflict ? (
                  <WorkbookInspectorActionButton
                    onClick={(event) => {
                      activateConflict?.(event.currentTarget, {
                        kind: "same_field_resolver",
                        conflictKey: firstConflict.key,
                      });
                    }}
                  >
                    Review conflicts
                  </WorkbookInspectorActionButton>
                ) : null}
                {entry.phase === "rejected" || entry.phase === "waiting" ? (
                  <WorkbookInspectorActionButton
                    onClick={() => owner.discard(entry.id)}
                  >
                    Discard this action
                  </WorkbookInspectorActionButton>
                ) : null}
                {outcome.canReview && entry.receipt ? (
                  <div hidden={currentReview?.batchId === entry.id}>
                    <WorkbookBatchRecordChoices
                      receipt={entry.receipt}
                      onReview={(record) => {
                        const scope = runtime.history.readScope;
                        const latest = owner
                          .getSnapshot()
                          .entries.find((item) => item.id === entry.id);
                        if (
                          !scope ||
                          !latest?.receipt?.changeSetId ||
                          !latest.receipt.rows.some(
                            (row) => row.record_id === record.recordId,
                          )
                        )
                          return;
                        if (
                          currentReview?.batchId === entry.id &&
                          currentReview.locator.recordId === record.recordId
                        )
                          return;
                        invoker.current =
                          document.activeElement instanceof HTMLElement
                            ? document.activeElement
                            : null;
                        setReview({
                          batchId: entry.id,
                          locator: {
                            scope,
                            viewSchemaId: latest.receipt.viewSchemaId,
                            recordId: record.recordId,
                            changeSetId: latest.receipt.changeSetId,
                            label: record.label,
                          },
                        });
                      }}
                    />
                  </div>
                ) : null}
              </section>
            );
          })}
        {currentReview?.batchId === selected ? (
          <WorkbookHistoryReview
            key={`${currentReview.locator.recordId}:${currentReview.locator.changeSetId}`}
            locator={currentReview.locator}
            onClose={() => {
              restoreFocus.current = true;
              setReview(null);
            }}
          />
        ) : null}
      </section>
    </WorkbookRecoveryDetail>
  );
}
