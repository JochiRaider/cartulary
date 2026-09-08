import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import type { ImportOperation } from "../../imports/workbookImportState";
import { importFailureMessage } from "../../services/importClient";

export function ImportOperationNotice({
  operation,
  controller,
  canRetry,
}: {
  readonly operation: ImportOperation;
  readonly controller: WorkbookImportController;
  readonly canRetry: boolean;
}) {
  return (
    <div aria-live="polite">
      {operation.phase === "pending" ? (
        <p>Awaiting {operation.attempt.kind} acknowledgement…</p>
      ) : (
        <>
          <p role="alert">
            {operation.attempt.kind === "select"
              ? "Mapping approval is retained. Selection was not confirmed. "
              : ""}
            {operation.phase === "uncertain"
              ? "Outcome uncertain. "
              : "Request rejected. "}
            {operation.failure
              ? importFailureMessage(operation.failure)
              : "Refresh before continuing."}
          </p>
          {operation.failure?.field ? (
            <p>Field: {operation.failure.field}</p>
          ) : null}
          <button
            type="button"
            disabled={!canRetry}
            onClick={() => void controller.retryWrite()}
          >
            {operation.phase === "uncertain"
              ? "Retry exact request"
              : operation.attempt.kind === "select"
                ? "Retry selection"
                : "Try again"}
          </button>
        </>
      )}
    </div>
  );
}
export const importActionsStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
};
export const importSectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
};
