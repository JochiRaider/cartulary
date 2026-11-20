import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type { PendingReplayPublicError } from "./pending/workbookPendingQueue";

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

/**
 * Projects one rejected replay unit into queue settlement. A temporarily
 * unavailable server refused the request before applying it: the unit settles
 * as retryable, so it stays queued in FIFO order with its captured request and
 * transaction identity and is never halted (Core 03 REQ-03-086a, REQ-03-100).
 */
export function workbookPendingMutationFailureResult(
  failure: WorkbookOperationFailure,
): {
  readonly status: number;
  readonly error: PendingReplayPublicError;
} {
  const publicCode =
    failure.publicCode === undefined ? {} : { publicCode: failure.publicCode };
  if (workbookFailureLifecycle(failure).kind === "temporarily_unavailable") {
    return {
      status: pendingFailureStatus.retryable,
      error: {
        code: "retryable",
        ...publicCode,
        message: failure.message,
        retryable: true,
      },
    };
  }
  if (failure.kind === "same_field_conflict") {
    return {
      status: 409,
      error: {
        code: "same_field_conflict",
        ...publicCode,
        message: failure.message,
        conflict: failure.conflict,
      },
    };
  }
  return {
    status: pendingFailureStatus[failure.kind],
    error: {
      code: failure.kind,
      ...publicCode,
      message: failure.message,
      ...(failure.kind === "retryable" ? { retryable: true } : {}),
      ...(failure.kind === "validation" && failure.fields?.[0] !== undefined
        ? { details: { field_key: failure.fields[0].field } }
        : {}),
    },
  };
}
