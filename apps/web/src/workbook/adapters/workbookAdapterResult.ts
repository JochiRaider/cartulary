import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";

export function invalidWorkbookAdapterResult<Accepted>(
  message: string,
): WorkbookOperationOutcome<Accepted> {
  return {
    kind: "rejected",
    failure: { kind: "invalid_contract", message },
  };
}

export function normalizeWorkbookAdapterFailure<Accepted>(
  outcome: WorkbookOperationOutcome<unknown>,
  invalidMessage: string,
): WorkbookOperationOutcome<Accepted> {
  return outcome.kind === "rejected" &&
    outcome.failure.kind === "invalid_contract"
    ? invalidWorkbookAdapterResult(invalidMessage)
    : (outcome as WorkbookOperationOutcome<Accepted>);
}

export function workbookAdapterCaughtResult<Accepted>(
  error: unknown,
  signal: AbortSignal,
  message: string,
): WorkbookPortResult<Accepted> {
  if (
    signal.aborted ||
    (error instanceof Error && error.name === "AbortError")
  ) {
    return { kind: "aborted" };
  }
  const failure: WorkbookOperationFailure = { kind: "retryable", message };
  return { kind: "rejected", failure };
}
