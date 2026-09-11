import { useSyncExternalStore } from "react";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookExplicitPatchOwner } from "../../runtime/WorkbookExplicitPatchOwner";
import { taskValue } from "./taskLifecycleModel";

export function TaskPatchRecovery({
  owner,
  rows,
}: {
  readonly owner: WorkbookExplicitPatchOwner;
  readonly rows: readonly WorkbookQueryRow[];
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  if (!snapshot.authority || !snapshot.entries.length) return null;
  return (
    <section aria-label="Task changes" style={sectionStyle}>
      {snapshot.entries
        .filter(
          (entry, index, entries) =>
            entry.phase !== "acknowledged" ||
            entry.reconciliation !== "complete" ||
            !entries
              .slice(index + 1)
              .some(
                (later) =>
                  later.intent.baseline.record_id ===
                  entry.intent.baseline.record_id,
              ),
        )
        .map((entry) => {
          const row = entry.receipt?.row;
          const outside =
            row &&
            entry.reconciliation === "complete" &&
            !rows.some((visible) => visible.record_id === row.record_id);
          return (
            <div key={entry.id} style={entryStyle}>
              <p role="status" style={{ margin: 0 }}>
                {row
                  ? `Saved ${taskValue(row, "task.title") || "Task"}: ${taskValue(row, "task.status")}.`
                  : entry.phase === "coordinating"
                    ? "Waiting for earlier Task writes and current saved values."
                    : entry.phase === "submitting"
                      ? "Saving Task changes."
                      : entry.phase === "uncertain"
                        ? "The Task change could not be confirmed. The original request is retained; closing this panel does not undo it."
                        : entry.phase === "conflict"
                          ? "Resolve the saved-field conflict. Your complete Task draft is retained."
                          : "The Task change was rejected. Your unsaved draft is retained."}
              </p>
              {row ? (
                <p style={{ margin: 0 }}>
                  Owner: {taskValue(row, "task.owner_user_id") || "none"}.{" "}
                  {taskValue(row, "task.blocked_reason")
                    ? `Blocked reason: ${taskValue(row, "task.blocked_reason")}. `
                    : ""}
                  {taskValue(row, "task.completed_at")
                    ? `Completed: ${taskValue(row, "task.completed_at")}. `
                    : ""}
                  Version {row.row_version}.
                </p>
              ) : null}
              {entry.failure?.kind === "validation"
                ? entry.failure.fields?.map((field) => (
                    <p key={field.field} style={{ margin: 0 }}>
                      {field.message}
                    </p>
                  ))
                : null}
              {row && entry.reconciliation !== "complete" ? (
                <p style={{ margin: 0 }}>
                  {entry.reconciliation === "refreshing"
                    ? "Refreshing the current view."
                    : "The change was saved, but the current view still needs a refresh."}
                </p>
              ) : null}
              {outside ? (
                <p style={{ margin: 0 }}>
                  This Task is no longer in the current view. Your filters are
                  unchanged.
                </p>
              ) : null}
              {entry.phase === "uncertain" ? (
                <button
                  type="button"
                  style={buttonStyle}
                  disabled={!owner.canSubmit()}
                  onClick={() => void owner.replay(entry.id)}
                >
                  Retry original Task change
                </button>
              ) : null}
              {row && entry.reconciliation === "required" ? (
                <button
                  type="button"
                  style={buttonStyle}
                  onClick={() => void owner.refresh(entry.id)}
                >
                  Refresh Task view
                </button>
              ) : null}
              {entry.phase === "rejected" ||
              (row && entry.reconciliation === "complete") ? (
                <button
                  type="button"
                  style={buttonStyle}
                  onClick={() => owner.dismiss(entry.id)}
                >
                  Dismiss Task notice
                </button>
              ) : null}
            </div>
          );
        })}
    </section>
  );
}
const sectionStyle = {
  maxBlockSize: "calc(var(--ct-layout-statusStripHeight) * 5)",
  overflowY: "auto" as const,
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  padding: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-ink)",
  background: "var(--ct-colors-surface-2)",
};
const entryStyle = { display: "grid", gap: "var(--ct-spacing-xs)" };
const buttonStyle = {
  justifySelf: "start",
  padding: "var(--ct-spacing-xs)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-component-button-secondary-textColor)",
  font: "inherit",
};
