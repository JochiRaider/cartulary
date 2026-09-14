import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";

export type WorkbookPortResult<Accepted> =
  | WorkbookOperationOutcome<Accepted>
  | { readonly kind: "aborted" };

/** Operation failures request scoped recovery; only the authority owner can retire an incident. */
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
      readonly kind: "operation_rejected";
      readonly failure: WorkbookOperationFailure;
    };

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
  return {
    kind:
      failure.kind === "stale_target"
        ? "record_unavailable"
        : "operation_rejected",
    failure,
  };
}
