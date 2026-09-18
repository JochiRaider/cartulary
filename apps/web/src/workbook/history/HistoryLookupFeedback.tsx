import { rowHistoryReadControlTestId } from "@cartulary/ui-contracts";
import { WorkbookInspectorActionButton } from "../inspector/presentation/WorkbookInspectorActions";
import type { HistoryLookupState } from "./HistoryPageLookup";

export function HistoryLookupFeedback({
  state,
  onContinue,
  onRestart,
  onCancel,
  purpose = "action",
}: {
  readonly purpose?: "action" | "change";
  readonly state: HistoryLookupState | undefined;
  readonly onContinue: () => void;
  readonly onRestart: () => void;
  readonly onCancel: () => void;
}) {
  if (!state || state.phase === "idle" || state.phase === "matched")
    return null;
  const recoverable = [
    "checking",
    "paused",
    "failed",
    "restart_required",
  ].includes(state.phase);
  return (
    <div>
      <p role={state.failure ? "alert" : "status"}>
        {state.failure?.message ??
          (state.phase === "checking"
            ? purpose === "change"
              ? "Looking for this change in current history…"
              : "Checking this action against current history…"
            : state.phase === "paused"
              ? "More history remains to be checked."
              : state.phase === "cancelled"
                ? "Checking cancelled."
                : "Review current history before continuing.")}
      </p>
      {state.phase === "paused" || state.phase === "failed" ? (
        <WorkbookInspectorActionButton
          data-testid={rowHistoryReadControlTestId("continue-checking")}
          onClick={onContinue}
        >
          {state.phase === "paused" ? "Continue checking" : "Retry checking"}
        </WorkbookInspectorActionButton>
      ) : null}
      {state.phase === "restart_required" ||
      (purpose === "change" &&
        ["changed", "cancelled", "unavailable"].includes(state.phase)) ? (
        <WorkbookInspectorActionButton
          data-testid={rowHistoryReadControlTestId("restart-checking")}
          onClick={onRestart}
        >
          Start checking again
        </WorkbookInspectorActionButton>
      ) : null}
      {recoverable ? (
        <WorkbookInspectorActionButton
          data-testid={rowHistoryReadControlTestId("cancel-checking")}
          tone="secondary"
          onClick={onCancel}
        >
          Cancel checking
        </WorkbookInspectorActionButton>
      ) : null}
    </div>
  );
}
