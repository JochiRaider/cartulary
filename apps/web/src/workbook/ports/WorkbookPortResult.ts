import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";

export type WorkbookPortResult<Accepted> =
  | WorkbookOperationOutcome<Accepted>
  | { readonly kind: "aborted" };

export function workbookOperationFailureIsAccessLoss(
  failure: WorkbookOperationFailure,
): boolean {
  return (
    failure.kind === "authentication_required" ||
    failure.kind === "authorization_lost" ||
    failure.kind === "stale_target"
  );
}
