import { decisionSupersessionTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { DecisionSupersessionOperation } from "./decisionSupersessionOperation";

export function WorkbookDecisionSupersessionRecovery({
  runtime,
}: {
  readonly runtime: WorkbookMutationRuntime;
}) {
  const owner = runtime.decisionSupersession;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.id,
        label: "Decision supersession",
        origin: entry.attempt.review.target.label,
        sheetRef: { kind: "view_schema", id: "cartulary.view.decisions.v1" },
        refreshViews:
          entry.receipt && entry.reconciliation !== "complete"
            ? ["cartulary.view.decisions.v1"]
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
      }))
    : [];
  const selected = useWorkbookRecoverySource("decision-supersession", items);
  return (
    <WorkbookRecoveryDetail source="decision-supersession" item={selected}>
      <section
        data-testid={decisionSupersessionTestId("recovery")}
        aria-label="Decision action recovery"
      >
        {snapshot.entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <DecisionRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              runtime={runtime}
            />
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}

function DecisionRecoveryEntry({
  entry,
  runtime,
}: {
  readonly entry: DecisionSupersessionOperation;
  readonly runtime: WorkbookMutationRuntime;
}) {
  const owner = runtime.decisionSupersession;
  const { review } = entry.attempt;
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <article
      aria-label={`Supersession of ${review.target.label}`}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        paddingBlock: "var(--ct-spacing-md)",
        borderBlockStart: "var(--ct-border-hairline)",
      }}
    >
      <strong>Decision supersession</strong>
      <p>
        Target: {review.target.label} ({review.target.recordId})<br />
        Replacement: {review.replacement.label} ({review.replacement.recordId})
      </p>
      <p style={{ whiteSpace: "pre-wrap" }}>Reason: {review.reason}</p>
      <p role="status">
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Supersession accepted. Both Decisions and related projections refreshed."
            : entry.reconciliation === "refreshing"
              ? "Supersession accepted. Refresh in progress."
              : "Supersession accepted. Refresh is still required."
          : entry.phase === "preparing"
            ? "Waiting for earlier writes and checking the captured review."
            : entry.phase === "submitting"
              ? "Supersession submitted; waiting for acknowledgement."
              : entry.phase === "uncertain"
                ? "The outcome is unknown. The server may have committed. Exact replay uses the same transaction and request."
                : "Supersession rejected. Refresh and review again before a new attempt."}
      </p>
      {entry.failure ? <p>{entry.failure.message}</p> : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canSubmit()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay exact supersession request
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.transportPending && entry.phase === "uncertain" ? (
        <p>
          Transport has not settled. Closing this panel does not cancel a server
          commit.
        </p>
      ) : null}
      {entry.receipt ? (
        <>
          <p>
            Accepted change set: {entry.receipt.change_set_id}. Target status:{" "}
            {entry.receipt.target_status}.
          </p>
          <details>
            <summary>Decision receipt</summary>
            <p>
              View: {entry.receipt.view_schema_id}
              <br />
              Target: {entry.receipt.target_record_id}, version{" "}
              {entry.receipt.target_row_version}
              <br />
              Replacement: {entry.receipt.superseding_record_id}, version{" "}
              {entry.receipt.superseding_row_version}
            </p>
            <p style={{ whiteSpace: "pre-wrap" }}>
              Recorded reason: {entry.receipt.reason}
            </p>
          </details>
          {entry.reconciliation !== "complete" ? (
            <WorkbookInspectorActionButton
              disabled={pending || entry.reconciliation === "refreshing"}
              onClick={() => void owner.refresh(entry.attempt.id)}
            >
              Retry refresh
            </WorkbookInspectorActionButton>
          ) : null}
          <p>Review or reverse this change through existing record History.</p>
        </>
      ) : null}
      {!pending &&
      (entry.phase === "rejected" ||
        (entry.receipt && entry.reconciliation === "complete")) ? (
        <WorkbookInspectorActionButton
          onClick={() => owner.dismiss(entry.attempt.id)}
        >
          Dismiss Decision action
        </WorkbookInspectorActionButton>
      ) : null}
    </article>
  );
}
