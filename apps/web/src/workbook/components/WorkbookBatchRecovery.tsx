import { requireViewContract } from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../shared/WorkbookRecoveryBoundary";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { workbookBatchRecoveryItems } from "../runtime/workbookBatchRecoveryItems";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";
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
  const owner = runtime.batches;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const mutation = useSyncExternalStore(runtime.subscribe, runtime.getSnapshot);
  const selected = useWorkbookRecoverySource(
    "batch",
    workbookBatchRecoveryItems(
      snapshot,
      new Set(
        mutation.conflicts.flatMap((entry) =>
          entry.batchOperationId ? [entry.batchOperationId] : [],
        ),
      ),
    ),
  );
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
            const label =
              entry.plan.operation === "pasteWorkbookClipboard"
                ? "Paste"
                : entry.plan.request.kind === "fill_down_v1"
                  ? "Fill"
                  : "Tag assignment";
            return (
              <section key={entry.id} aria-label={`${label} ${index + 1}`}>
                <strong>
                  {label} {index + 1} ·{" "}
                  {requireViewContract(entry.plan.request.view_schema_id).title}
                </strong>
                <p>
                  {entry.phase === "uncertain"
                    ? "The result is unknown. Retry checks the original action without duplicating accepted work."
                    : entry.phase === "rejected"
                      ? (entry.failure?.message ??
                        "The batch was rejected. Review the original range.")
                      : entry.phase === "acknowledged"
                        ? `${entry.receipt?.rows.length ? "Accepted work is saved." : "No row changes were needed or accepted."}${conflicts.length ? ` ${conflicts.length} conflicts need review.` : ""}${entry.reconciliation === "required" ? " The view could not refresh." : ""}`
                        : "Waiting for earlier work or applying this batch."}
                </p>
                {entry.phase === "rejected" || entry.phase === "uncertain" ? (
                  <label
                    style={{ display: "grid", gap: "var(--ct-spacing-xs)" }}
                  >
                    Original input
                    <textarea
                      readOnly
                      aria-label="Original batch input"
                      value={
                        entry.plan.operation === "pasteWorkbookClipboard"
                          ? entry.plan.request.clipboard_text
                          : entry.plan.request.kind === "fill_down_v1"
                            ? (entry.plan.request.value ?? "")
                            : (entry.plan.request.tag_name ?? "")
                      }
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
              </section>
            );
          })}
      </section>
    </WorkbookRecoveryDetail>
  );
}
