import { useLayoutEffect, useRef } from "react";
import { workbookFormFieldsStyle as observationStack } from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type {
  ObservationOperation,
  ObservationOwnerPort,
} from "./observationOperation";
export function ObservationOperationStatus({
  owner,
  entry,
}: {
  owner: ObservationOwnerPort;
  entry: ObservationOperation;
}) {
  const container = useRef<HTMLDivElement>(null),
    status = useRef<HTMLParagraphElement>(null);
  const retainFocus = container.current?.contains(document.activeElement);
  useLayoutEffect(() => {
    if (retainFocus && document.activeElement === document.body)
      status.current?.focus();
  });
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <div ref={container} style={observationStack}>
      <p ref={status} role="status" tabIndex={-1}>
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Observation change saved. Records and history refreshed."
            : entry.reconciliation === "refreshing"
              ? "Observation change saved. Refreshing records and history…"
              : "Observation change saved. Refresh is still required."
          : entry.phase === "uncertain"
            ? "The observation outcome is unknown. The server may have saved it. Your original request is retained."
            : entry.phase === "rejected"
              ? "The observation change was not accepted. Your draft is retained."
              : entry.phase === "preparing"
                ? "Checking the source, observation and target…"
                : "Saving observation change…"}
      </p>
      {entry.failure ? <p>{entry.failure.message}</p> : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canReplay()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay original observation request
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.receipt && entry.reconciliation !== "complete" ? (
        <WorkbookInspectorActionButton
          disabled={pending || entry.reconciliation === "refreshing"}
          onClick={() => void owner.refresh(entry.attempt.id)}
        >
          Retry observation refresh
        </WorkbookInspectorActionButton>
      ) : null}
      {!pending &&
      (entry.phase === "rejected" || entry.reconciliation === "complete") ? (
        <WorkbookInspectorActionButton
          onClick={() => {
            const heading = container.current
              ?.closest("section")
              ?.querySelector("h3");
            owner.dismiss(entry.attempt.id);
            heading?.focus();
          }}
        >
          Dismiss observation result
        </WorkbookInspectorActionButton>
      ) : null}
    </div>
  );
}
