import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { PendingReplayPublicError } from "../utils/workbookPendingQueue";

const pendingFailureStatus = {
  authentication_required: 401,
  authorization_lost: 403,
  client_txn_conflict: 409,
  invalid_contract: 422,
  retryable: 503,
  same_field_conflict: 409,
  stale_target: 404,
  terminal: 422,
  validation: 400,
} as const satisfies Record<WorkbookOperationFailure["kind"], number>;

export function workbookPendingMutationFailureResult(
  failure: WorkbookOperationFailure,
): {
  readonly status: number;
  readonly error: PendingReplayPublicError;
} {
  if (failure.kind === "same_field_conflict") {
    return {
      status: 409,
      error: {
        code: "same_field_conflict",
        message: failure.message,
        conflict: failure.conflict,
      },
    };
  }
  return {
    status: pendingFailureStatus[failure.kind],
    error: {
      code: failure.kind,
      message: failure.message,
      ...(failure.kind === "retryable" ? { retryable: true } : {}),
      ...(failure.kind === "validation" && failure.fields?.[0] !== undefined
        ? { details: { field_key: failure.fields[0].field } }
        : {}),
    },
  };
}
