import type { CartularyErrorPresentation } from "@cartulary/ui-contracts";
import type { WorkbookSameFieldConflictPayload } from "../runtime/workbookConflictModel";

export type WorkbookOperationFieldFailure = {
  readonly field: string;
  readonly message: string;
};

export type WorkbookOperationFailure = (
  | {
      readonly kind: "validation";
      readonly message: string;
      readonly fields?: readonly WorkbookOperationFieldFailure[];
    }
  | {
      readonly kind: "same_field_conflict";
      readonly message: string;
      readonly conflict: WorkbookSameFieldConflictPayload;
    }
  | { readonly kind: "client_txn_conflict"; readonly message: string }
  | { readonly kind: "authentication_required"; readonly message: string }
  | { readonly kind: "authorization_lost"; readonly message: string }
  | { readonly kind: "stale_target"; readonly message: string }
  | { readonly kind: "retryable"; readonly message: string }
  | { readonly kind: "invalid_contract"; readonly message: string }
  | { readonly kind: "terminal"; readonly message: string }
) & { readonly presentation?: CartularyErrorPresentation | undefined };

export type WorkbookOperationOutcome<Accepted> =
  | { readonly kind: "accepted"; readonly value: Accepted }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure };
