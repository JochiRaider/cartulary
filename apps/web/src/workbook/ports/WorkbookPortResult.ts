import type { HTTPOperationID } from "../../services/browserApi";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";

export type WorkbookPortResult<
  Accepted,
  OperationID extends HTTPOperationID = HTTPOperationID,
> =
  | WorkbookOperationOutcome<Accepted, OperationID>
  | { readonly kind: "aborted" };

/**
 * Operation failures request scoped recovery; only the authority owner can
 * retire an incident. Temporary unavailability is retryable, not a rejection.
 */
export type WorkbookFailureLifecycle =
  | {
      readonly kind: "authority_unavailable";
      readonly scope: "session" | "incident";
      readonly failure: WorkbookOperationFailure;
    }
  | {
      readonly kind: "record_unavailable";
      readonly failure: WorkbookOperationFailure;
    }
  | {
      readonly kind: "temporarily_unavailable";
      readonly failure: WorkbookOperationFailure;
    }
  | {
      readonly kind: "operation_rejected";
      readonly failure: WorkbookOperationFailure;
    };

/**
 * The declared `503` `service_unavailable` (Core 03 REQ-03-086a, Core 04
 * AC-618): the request was refused before admission, or its commit was proven
 * absent, so nothing was applied. The same captured request stays retryable;
 * the failure is never a contract failure or a terminal rejection.
 */
export function isWorkbookTemporaryUnavailability(
  failure: WorkbookOperationFailure,
): boolean {
  return failure.publicCode === "service_unavailable";
}

export function workbookFailureLifecycle(
  failure: WorkbookOperationFailure,
): WorkbookFailureLifecycle {
  if (failure.kind === "authentication_required")
    return { kind: "authority_unavailable", scope: "session", failure };
  if (
    failure.kind === "authorization_lost" ||
    failure.publicCode === "incident_not_found"
  )
    return { kind: "authority_unavailable", scope: "incident", failure };
  if (isWorkbookTemporaryUnavailability(failure))
    return { kind: "temporarily_unavailable", failure };
  return {
    kind:
      failure.kind === "stale_target"
        ? "record_unavailable"
        : "operation_rejected",
    failure,
  };
}
