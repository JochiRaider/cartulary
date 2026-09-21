import { workbookFormFieldsStyle as lifecycleStack } from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type {
  IndicatorLifecycleOwnerPort,
  LifecycleOperation,
} from "./indicatorLifecycleOperation";
export function IndicatorLifecycleOperationStatus({
  owner,
  entry,
}: {
  owner: IndicatorLifecycleOwnerPort;
  entry: LifecycleOperation;
}) {
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <div style={lifecycleStack}>
      <p role="status">
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Interval saved. Indicator and history refreshed."
            : entry.reconciliation === "refreshing"
              ? "Interval saved. Refreshing Indicator and history…"
              : "Interval saved. Refresh is still required."
          : entry.phase === "uncertain"
            ? "The interval outcome is unknown. The server may have saved it. Your original request is retained."
            : entry.phase === "rejected"
              ? "The interval was not accepted. Your draft is retained."
              : entry.phase === "preparing"
                ? "Checking the Indicator before appending…"
                : "Appending interval…"}
      </p>
      {entry.failure ? <p>{entry.failure.message}</p> : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canReplay()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay original interval request
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.receipt && entry.reconciliation !== "complete" ? (
        <WorkbookInspectorActionButton
          disabled={pending || entry.reconciliation === "refreshing"}
          onClick={() => void owner.refresh(entry.attempt.id)}
        >
          Retry interval refresh
        </WorkbookInspectorActionButton>
      ) : null}
      {!pending &&
      (entry.phase === "rejected" || entry.reconciliation === "complete") ? (
        <WorkbookInspectorActionButton
          onClick={() => owner.dismiss(entry.attempt.id)}
        >
          Dismiss interval result
        </WorkbookInspectorActionButton>
      ) : null}
    </div>
  );
}
