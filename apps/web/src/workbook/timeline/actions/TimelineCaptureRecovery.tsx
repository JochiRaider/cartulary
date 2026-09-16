import { timelineCaptureActionTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { TimelineCaptureOperation } from "../ports/TimelineRecordActionPort";
import type { WorkbookTimelineCaptureActionOwner } from "./WorkbookTimelineCaptureActionOwner";

export function TimelineCaptureRecovery({
  owner,
}: {
  readonly owner: WorkbookTimelineCaptureActionOwner;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: String(entry.key),
        refreshViews:
          entry.receipt && entry.reconciliation !== "complete"
            ? ["cartulary.view.timeline.v2"]
            : [],
        label: "Timeline action",
        origin: entry.review.target.label,
        sheetRef: { kind: "view_schema", id: "cartulary.view.timeline.v2" },
        order,
        summary: entry.receipt
          ? entry.reconciliation === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected" || entry.phase === "preparation_failed"
              ? "Review required"
              : "In progress",
        attention:
          entry.receipt && entry.reconciliation === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected" ||
                entry.phase === "preparation_failed"
              ? "attention"
              : "progress",
      }))
    : [];
  const selected = useWorkbookRecoverySource("timeline-capture", items);
  return (
    <WorkbookRecoveryDetail source="timeline-capture" item={selected}>
      <section aria-label="Timeline action recovery">
        {snapshot.entries
          .filter((entry) => String(entry.key) === selected)
          .map((entry) => (
            <article
              key={entry.key}
              aria-label={`Timeline action for ${entry.review.target.label}`}
              style={{ marginBlock: "var(--ct-spacing-md)" }}
            >
              <strong>
                {entry.review.action === "mark-reviewed"
                  ? "Mark reviewed"
                  : "Supersede"}
                : {entry.review.target.label}
              </strong>
              <p>
                {entry.review.target.recordId} · reviewed version{" "}
                {entry.review.target.rowVersion}
              </p>
              {entry.review.reason ? (
                <p style={{ whiteSpace: "pre-wrap" }}>{entry.review.reason}</p>
              ) : null}
              {entry.review.action === "supersede" ? (
                <p>
                  Replacement:{" "}
                  {entry.review.replacement
                    ? `${entry.review.replacement.label} · ${entry.review.replacement.context} · ${entry.review.replacement.recordId}`
                    : "No replacement"}
                </p>
              ) : null}
              <p
                data-testid={timelineCaptureActionTestId(
                  "result",
                  entry.review.target.recordId,
                )}
              >
                {status(entry)}
              </p>
              {entry.failure ? (
                <p role="alert">{entry.failure.message}</p>
              ) : null}
              {entry.receipt ? (
                <p>
                  Saved version {entry.receipt.data.row_version} · change{" "}
                  {entry.receipt.data.change_set_id}. Your current selection and
                  filters are preserved; the completed row may leave the current
                  results.
                </p>
              ) : null}
              {entry.phase === "uncertain" ? (
                <>
                  <p>
                    The outcome remains unknown. Retry sends the exact original
                    request and transaction identity.
                  </p>
                  <WorkbookInspectorActionButton
                    disabled={!owner.canSubmit()}
                    title={owner.unavailableReason() ?? undefined}
                    data-testid={timelineCaptureActionTestId(
                      "retry",
                      entry.review.target.recordId,
                    )}
                    onClick={() => {
                      void owner.replay(entry.key);
                    }}
                  >
                    Retry exact action
                  </WorkbookInspectorActionButton>
                </>
              ) : null}
              {entry.receipt && entry.reconciliation !== "complete" ? (
                <WorkbookInspectorActionButton
                  disabled={entry.reconciliation === "refreshing"}
                  data-testid={timelineCaptureActionTestId(
                    "refresh",
                    entry.review.target.recordId,
                  )}
                  onClick={() => {
                    void owner.refresh(entry.key);
                  }}
                >
                  Refresh result
                </WorkbookInspectorActionButton>
              ) : null}
              {["preparation_failed", "rejected"].includes(entry.phase) ||
              entry.reconciliation === "complete" ? (
                <WorkbookInspectorActionButton
                  disabled={entry.transportPending}
                  onClick={() => owner.dismiss(entry.key)}
                >
                  Dismiss action
                </WorkbookInspectorActionButton>
              ) : null}
            </article>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
function status(entry: TimelineCaptureOperation) {
  if (entry.receipt)
    return `Timeline ${entry.review.action === "mark-reviewed" ? "review" : "supersession"} completed.${entry.reconciliation === "complete" ? "" : " Refresh still required."}`;
  if (entry.phase === "uncertain") return "Timeline action outcome unknown.";
  if (entry.phase === "preparation_failed")
    return "Timeline action needs another review. No action was sent.";
  if (entry.phase === "rejected") return "Timeline action rejected.";
  return "Timeline action in progress.";
}
