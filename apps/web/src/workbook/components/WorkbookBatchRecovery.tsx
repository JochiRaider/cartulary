import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../shared/workbookRecoveryNavigation";
import { WorkbookHistoryReview } from "../history/WorkbookHistoryReview";
import {
  type HistoryReviewLocator,
  historyReviewAuthorized,
} from "../history/workbookHistoryReview";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { workbookBatchOutcome } from "../runtime/workbookBatchOutcome";
import { workbookBatchRecoveryItems } from "../runtime/workbookBatchRecoveryItems";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";
import { WorkbookBatchRecordChoices } from "./WorkbookBatchRecordChoices";
import { inputStyle } from "./workbookGridControlStyles";

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
  const mutation = useSyncExternalStore(runtime.subscribe, runtime.getSnapshot);
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
      setReview(null);
      invoker.current = null;
    },
  });
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
            return (
              <section key={entry.id} aria-label={`${label} ${index + 1}`}>
                <p role="status">{outcome.detail}</p>
                {outcome.refresh !== "none" ? (
                  <p role="status">
                    {outcome.refresh === "refreshing"
                      ? "Refreshing the view…"
                      : outcome.refresh === "required"
                        ? "The view could not refresh. The acknowledged result is unchanged."
                        : "View refresh is pending."}
                  </p>
                ) : null}
                {entry.phase === "rejected" || entry.phase === "uncertain" ? (
                  <label
                    style={{ display: "grid", gap: "var(--ct-spacing-xs)" }}
                  >
                    Original input
                    <textarea
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
                {entry.phase === "uncertain" ? (
                  <WorkbookInspectorActionButton
                    disabled={!owner.canWrite() || entry.transportPending}
                    onClick={() => owner.retry(entry.id)}
                  >
                    Retry {label.toLowerCase()}
                  </WorkbookInspectorActionButton>
                ) : null}
                {entry.phase === "acknowledged" &&
                entry.reconciliation === "required" ? (
                  <WorkbookInspectorActionButton
                    onClick={() => owner.retry(entry.id)}
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
