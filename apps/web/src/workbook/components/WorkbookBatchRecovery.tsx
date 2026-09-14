import { requireViewContract } from "@cartulary/view-contracts";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import { RecoverySurface } from "./RecoverySurface";
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
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLElement>(null);
  useEffect(() => {
    if (!snapshot.authority) setOpen(false);
  }, [snapshot.authority]);
  const entries = snapshot.entries;
  if (!entries.length && !snapshot.admissionError) return null;
  const unsettled = entries.filter(
    (entry) =>
      entry.phase !== "acknowledged" ||
      entry.reconciliation !== "complete" ||
      runtime
        .getSnapshot()
        .conflicts.some((conflict) => conflict.batchOperationId === entry.id),
  );
  const status = snapshot.admissionError
    ? "Review"
    : entries.some((entry) => entry.phase === "uncertain")
      ? "Retry"
      : entries.some((entry) => entry.reconciliation === "required")
        ? "Refresh"
        : entries.some((entry) => entry.phase === "rejected") ||
            runtime
              .getSnapshot()
              .conflicts.some((entry) => entry.batchOperationId)
          ? "Review"
          : unsettled.length
            ? "Pending"
            : "Complete";
  const statusMessage =
    snapshot.admissionError ??
    (status === "Retry"
      ? "Batch outcome unknown. Retry is available."
      : status === "Refresh"
        ? "Batch accepted. Refresh is still needed."
        : status === "Complete"
          ? "Batch complete."
          : "Batch work pending or needs review.");
  const close = () => {
    const restore = panel.current?.contains(document.activeElement);
    setOpen(false);
    if (restore) trigger.current?.focus({ preventScroll: true });
  };
  return (
    <>
      <WorkbookInspectorActionButton
        ref={trigger}
        style={{ whiteSpace: "nowrap", flexShrink: 0 }}
        title={statusMessage}
        aria-expanded={open}
        onClick={() => {
          if (!open) {
            const active = runtime.getSnapshot().conflicts[0];
            if (active) runtime.dismissConflict(active.key, false);
          }
          setOpen(!open);
        }}
      >
        Batch actions{unsettled.length ? ` (${unsettled.length})` : ""}:{" "}
        {status}
      </WorkbookInspectorActionButton>
      <span
        role="status"
        aria-label="Batch action updates"
        style={visuallyHiddenStyle}
      >
        {statusMessage}
      </span>
      {open ? (
        <RecoverySurface
          ref={panel}
          aria-label="Batch action recovery"
          data-grid-editor-external-action="true"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              close();
            }
          }}
        >
          <strong>Paste, fill and tag actions</strong>
          <p>{statusMessage}</p>
          {snapshot.admissionError ? (
            <p role="alert">{snapshot.admissionError}</p>
          ) : null}
          {entries.map((entry, index) => {
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
                    onClick={() => {
                      if (trigger.current)
                        activateConflict?.(trigger.current, {
                          kind: "same_field_resolver",
                          conflictKey: firstConflict.key,
                        });
                      setOpen(false);
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
          <WorkbookInspectorActionButton onClick={close}>
            Close
          </WorkbookInspectorActionButton>
        </RecoverySurface>
      ) : null}
    </>
  );
}
