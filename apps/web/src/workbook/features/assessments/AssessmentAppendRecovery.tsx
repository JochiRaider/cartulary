import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { secondaryButtonStyle } from "../../components/workbookGridControlStyles";
import type { WorkbookAssessmentAuthoringOwner } from "./WorkbookAssessmentAuthoringOwner";
export function AssessmentAppendRecovery({
  owner,
}: {
  readonly owner: WorkbookAssessmentAuthoringOwner;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.clientTxnId,
        label: "Assessment append",
        origin:
          entry.attempt.review.draft.values.subjectDisplayText || "Assessments",
        sheetRef: entry.attempt.review.sheetRef,
        refreshViews:
          entry.receipt && entry.refresh !== "complete"
            ? ["cartulary.view.assessments.v1"]
            : [],
        order,
        summary: entry.receipt
          ? entry.refresh === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected"
              ? "Review required"
              : "Appending",
        attention:
          entry.receipt && entry.refresh === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected"
              ? "attention"
              : "progress",
      }))
    : [];
  const selected = useWorkbookRecoverySource("assessment", items);
  return (
    <WorkbookRecoveryDetail source="assessment" item={selected}>
      <section aria-label="Retained Assessment appends">
        <p>
          Closing the inspector retains dispatched appends and their results for
          this incident session.
        </p>
        <ul>
          {snapshot.entries
            .filter((entry) => entry.attempt.clientTxnId === selected)
            .map((entry) => (
              <li key={entry.attempt.clientTxnId}>
                <p>
                  {entry.receipt
                    ? "Assessment created."
                    : entry.phase === "uncertain"
                      ? "Append result unconfirmed."
                      : entry.phase === "rejected"
                        ? "Append rejected; the editable draft is retained."
                        : "Appending assessment…"}{" "}
                  {entry.message}
                </p>
                {entry.receipt ? (
                  <p>
                    Record: {entry.receipt.data.row.record_id}. Version:{" "}
                    {entry.receipt.data.row.row_version}. Change set:{" "}
                    {entry.receipt.data.change_set_id}.
                  </p>
                ) : null}
                {entry.phase === "uncertain" ? (
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    disabled={
                      entry.transportPending ||
                      snapshot.preparing ||
                      !owner.canReplay()
                    }
                    onClick={() => void owner.replay(entry.attempt.clientTxnId)}
                  >
                    Recover assessment append
                  </button>
                ) : null}
                {entry.receipt && entry.refresh !== "complete" ? (
                  <button
                    style={secondaryButtonStyle}
                    type="button"
                    disabled={entry.refresh === "refreshing"}
                    onClick={() =>
                      void owner.retryRefresh(entry.attempt.clientTxnId)
                    }
                  >
                    Retry assessment refresh
                  </button>
                ) : null}
              </li>
            ))}
        </ul>
      </section>
    </WorkbookRecoveryDetail>
  );
}
